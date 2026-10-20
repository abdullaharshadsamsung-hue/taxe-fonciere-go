package bati2018

import (
	"context"
	"math/big"
	"reflect"
	"testing"

	"migrated-app/internal/layout/linkage"
	"migrated-app/internal/rates"
)

// ---- mocks ----

type mockDecoder struct {
	calls int
	set   RateSet
}

func (m *mockDecoder) Decode(z *rates.ZES) RateSet {
	m.calls++
	return m.set
}

type mockTaudis struct {
	calls int
	res   rates.Result
	err   error
}

func (m *mockTaudis) Fetch(ctx context.Context, req rates.Request) (rates.Result, error) {
	m.calls++
	return m.res, m.err
}

// ---- constructor ----

func TestNewCalculator(t *testing.T) {
	tests := []struct {
		name    string
		decoder RateDecoder
		wantErr bool
	}{
		{"nil decoder", nil, true},
		{"valid decoder", &mockDecoder{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewCalculator(nil, nil, tt.decoder)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err=%v wantErr=%v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if c == nil {
					t.Fatal("nil calculator")
				}
				for i := range c.zes {
					for j := range c.zes[i] {
						if c.zes[i][j] != ' ' {
							t.Fatalf("zes[%d][%d] not blank", i, j)
						}
					}
				}
			}
		})
	}
}

// ---- validation failures ----

func TestCalculateValidationFailures(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		parm  string
	}{
		{"empty block (all spaces)", nil, "M"},
		{"garbage block", []byte("XXXXXXXXXXXXXXXXXXXXXXXXXXXX"), "B"},
	}
	valid := map[int]bool{RCBadArticleCode: true, RCBadYear: true, RCBadBase: true, RCBadOmZone: true, RCBadOmBase: true}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dec := &mockDecoder{}
			tf := &mockTaudis{}
			c, err := NewCalculator(tf, nil, dec)
			if err != nil {
				t.Fatal(err)
			}
			out, err := c.Calculate(context.Background(), tt.input, tt.parm)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if out.CR != CRValidation {
				t.Errorf("CR=%d want %d", out.CR, CRValidation)
			}
			if !valid[out.RC] {
				t.Errorf("RC=%d not a validation RC", out.RC)
			}
			if tf.calls != 0 || dec.calls != 0 {
				t.Errorf("rate source called on validation failure")
			}
			if !reflect.DeepEqual(out.Retour, linkage.NewRetB()) {
				t.Errorf("output block not the initialized block")
			}
		})
	}
}

func TestValidateBlankBlockFailsOnArticleOrLater(t *testing.T) {
	cb := moveCombat(nil)
	cr, rc, err := validate(&cb)
	if err != nil {
		t.Fatal(err)
	}
	if cr != CRValidation {
		t.Fatalf("cr=%d", cr)
	}
	// Blank bases are non-numeric, so a check at least as late as RC 11 wins
	// over RC 1/2.
	if rc == RCBadArticleCode || rc == RCBadYear {
		t.Errorf("rc=%d: later checks must override earlier ones", rc)
	}
}

func TestMoveCombat(t *testing.T) {
	tests := []struct {
		name string
		in   []byte
	}{
		{"nil", nil},
		{"short", []byte("2018")},
		{"long", make([]byte, 700)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cb := moveCombat(tt.in)
			n := len(tt.in)
			if n > len(cb) {
				n = len(cb)
			}
			for i := 0; i < n; i++ {
				if cb[i] != tt.in[i] {
					t.Fatalf("byte %d differs", i)
				}
			}
			ref := linkage.NewCombat()
			for i := n; i < len(cb); i++ {
				if cb[i] != ref[i] {
					t.Fatalf("byte %d not initialized", i)
				}
			}
		})
	}
}

// ---- arithmetic ----

func TestRoundQuo(t *testing.T) {
	tests := []struct {
		num   int64
		scale int
		want  int64
	}{
		{15, 1, 2},
		{14, 1, 1},
		{-15, 1, -2},
		{-14, 1, -1},
		{0, 4, 0},
		{12345, 2, 123},
		{12350, 2, 124},
	}
	for _, tt := range tests {
		if got := roundQuo(big.NewInt(tt.num), tt.scale); got != tt.want {
			t.Errorf("roundQuo(%d,%d)=%d want %d", tt.num, tt.scale, got, tt.want)
		}
	}
}

func TestContribution(t *testing.T) {
	tests := []struct {
		name       string
		base, rate int64
		want       int64
	}{
		{"round down", 1000, 12_345_000, 123},
		{"half up", 1000, 12_350_000, 124},
		{"negative half away", -1000, 12_350_000, -124},
		{"zero base", 0, 12_345_000, 0},
		{"zero rate", 5000, 0, 0},
		{"100 percent", 777, 100_000_000, 777},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := contribution(tt.base, tt.rate); got != tt.want {
				t.Errorf("got %d want %d", got, tt.want)
			}
		})
	}
}

func TestFees(t *testing.T) {
	tests := []struct {
		name           string
		total, arn, fr int64
		wantFa, wantFn int64
	}{
		{"3%", 1000, f300Arn, f300Frs, 10, 20},
		{"8%", 1000, f800Arn, f800Frs, 44, 36},
		{"9%", 1000, f900Arn, f900Frs, 54, 36},
		{"zero", 0, f800Arn, f800Frs, 0, 0},
		{"rounding", 25, f800Arn, f800Frs, 1, 1}, // 1.1→1, 2.0→2
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fa, fn := fees(tt.total, tt.arn, tt.fr)
			if fa != tt.wantFa || fn != tt.wantFn {
				t.Errorf("fees=(%d,%d) want (%d,%d)", fa, fn, tt.wantFa, tt.wantFn)
			}
			total := roundQuo(new(big.Int).Mul(big.NewInt(tt.total), big.NewInt(tt.fr)), feeScale)
			if fa+fn != total {
				t.Errorf("FN != total - FA")
			}
		})
	}
}

func TestTotalsOf(t *testing.T) {
	ct := NewCotis()
	ct.CotisOm = [CotisOmCount]int64{1, 2, 3, 4, 5, 6}
	ct.CotisOmi = 10
	ct.Coticom, ct.Cotidep, ct.Coticu, ct.Mcoge3 = 100, 200, 300, 400
	ct.Cotisyn, ct.Mcota3 = 7, 8
	ct.Cotitsen = [CotisTsenCount]int64{50, 60}
	ct.Fa300 = 999 // must not enter TCTOM
	got := totalsOf(&ct)
	want := totals{om: 31, t3: 1000, t8: 46, t9: 110}
	if got != want {
		t.Errorf("got %+v want %+v", got, want)
	}
}

func TestCotisOmSubscriptQuirk(t *testing.T) {
	ct := NewCotis()
	ct.Fa300 = 42
	v, err := ct.CotisOmSubscript(CotisOmCount + 1)
	if err != nil || v != 42 {
		t.Errorf("got %d,%v", v, err)
	}
	if _, err := ct.CotisOmSubscript(0); err == nil {
		t.Error("expected error")
	}
}