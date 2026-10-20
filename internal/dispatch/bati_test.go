package dispatch

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"migrated-app/internal/calc/bati2018"
)

type mockCalc struct {
	calls   int
	parm    string
	entree  []byte
	out     bati2018.Output
	err     error
	mutated bool
}

func (m *mockCalc) Calculate(ctx context.Context, combatm []byte, parm string) (bati2018.Output, error) {
	m.calls++
	m.parm = parm
	m.entree = append([]byte(nil), combatm...)
	return m.out, m.err
}

func makeEntree(year string) []byte {
	e := bytes.Repeat([]byte{'x'}, RecordLen)
	copy(e[1:5], year)
	return e
}

func makeOutput(fill byte, cr, rc int) bati2018.Output {
	var out bati2018.Output
	for i := range out.Retour {
		out.Retour[i] = fill
	}
	out.CR = cr
	out.RC = rc
	return out
}

func TestNewBatiDispatcherNil(t *testing.T) {
	d, err := NewBatiDispatcher(nil)
	if err == nil || d != nil {
		t.Fatalf("expected error and nil dispatcher, got %v, %v", d, err)
	}
	d, err = NewBatiDispatcher(&mockCalc{})
	if err != nil || d == nil {
		t.Fatalf("unexpected: %v, %v", d, err)
	}
}

func TestBatiYearProgram(t *testing.T) {
	tests := []struct {
		year string
		want string
		ok   bool
	}{
		{"2006", "EFITA3B6", true},
		{"2007", "EFITA3B7", true},
		{"2008", "EFITA3B8", true},
		{"2009", "EFITA3B8", true},
		{"2010", "EFITA3B0", true},
		{"2011", "EFITA3B1", true},
		{"2012", "EFITA3B2", true},
		{"2013", "EFITA3B3", true},
		{"2014", "EFITA3B4", true},
		{"2015", "EFITA3B5", true},
		{"2016", "EFITA3B6", true},
		{"2017", "EFITA3B7", true},
		{"2018", "EFITA3B8", true},
		{"2005", "", false},
		{"2019", "", false},
		{" 201", "", false},
		{"    ", "", false},
		{"", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.year, func(t *testing.T) {
			got, ok := BatiYearProgram(tt.year)
			if got != tt.want || ok != tt.ok {
				t.Errorf("BatiYearProgram(%q) = %q,%v want %q,%v", tt.year, got, ok, tt.want, tt.ok)
			}
		})
	}
}

func TestDispatchToEFITA3B8(t *testing.T) {
	tests := []struct {
		name     string
		year     string
		crm, rcm int
		wantParm string
	}{
		{"2018 croisic", "2018", CRCroisic, 3, "M"},
		{"2018 normal 00", "2018", 0, 0, " "},
		{"2018 normal 05", "2018", 5, 7, " "},
		{"2008", "2008", 0, 0, " "},
		{"2009 croisic", "2009", 99, 1, "M"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mockCalc{out: makeOutput('R', 12, 2)}
			d, _ := NewBatiDispatcher(m)
			entree := makeEntree(tt.year)
			orig := append([]byte(nil), entree...)
			sortie := make([]byte, RecordLen)
			cr, rc, err := d.Dispatch(context.Background(), entree, sortie, tt.crm, tt.rcm)
			if err != nil {
				t.Fatalf("err: %v", err)
			}
			if m.calls != 1 {
				t.Fatalf("calls=%d", m.calls)
			}
			if m.parm != tt.wantParm {
				t.Errorf("parm=%q want %q", m.parm, tt.wantParm)
			}
			if tt.wantParm == "M" && m.parm != bati2018.ParmTaudis {
				t.Errorf("parm=%q want ParmTaudis %q", m.parm, bati2018.ParmTaudis)
			}
			if !bytes.Equal(m.entree, orig) || !bytes.Equal(entree, orig) {
				t.Error("entree altered or not passed whole")
			}
			if cr != 12 || rc != 2 {
				t.Errorf("cr,rc=%d,%d", cr, rc)
			}
			if !bytes.Equal(sortie, bytes.Repeat([]byte{'R'}, RecordLen)) {
				t.Error("sortie not copied from callee")
			}
		})
	}
}

func TestDispatchUnportedYears(t *testing.T) {
	for _, y := range []string{"2006", "2007", "2010", "2011", "2012", "2013", "2014", "2015", "2016", "2017"} {
		t.Run(y, func(t *testing.T) {
			m := &mockCalc{}
			d, _ := NewBatiDispatcher(m)
			sortie := bytes.Repeat([]byte{'S'}, RecordLen)
			cr, rc, err := d.Dispatch(context.Background(), makeEntree(y), sortie, 99, 4)
			if !errors.Is(err, ErrCalculatorUnavailable) {
				t.Fatalf("err=%v", err)
			}
			if cr != 0 || rc != 4 || m.calls != 0 {
				t.Errorf("cr=%d rc=%d calls=%d", cr, rc, m.calls)
			}
			if !bytes.Equal(sortie, bytes.Repeat([]byte{'S'}, RecordLen)) {
				t.Error("sortie modified")
			}
		})
	}
}

func TestDispatchUnknownYears(t *testing.T) {
	tests := []struct {
		name   string
		entree []byte
	}{
		{"2005", makeEntree("2005")},
		{"2019", makeEntree("2019")},
		{"leading space", makeEntree(" 201")},
		{"spaces", makeEntree("    ")},
		{"shifted", append([]byte("x 2018"), make([]byte, 594)...)},
		{"nil entree", nil},
		{"short entree", []byte("x20")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mockCalc{}
			d, _ := NewBatiDispatcher(m)
			sortie := bytes.Repeat([]byte{'S'}, RecordLen)
			cr, rc, err := d.Dispatch(context.Background(), tt.entree, sortie, 99, 6)
			if err != nil {
				t.Fatalf("err=%v", err)
			}
			if cr != 0 || rc != 6 || m.calls != 0 {
				t.Errorf("cr=%d rc=%d calls=%d", cr, rc, m.calls)
			}
			if !bytes.Equal(sortie, bytes.Repeat([]byte{'S'}, RecordLen)) {
				t.Error("sortie modified")
			}
		})
	}
}

func TestDispatchShortSortie(t *testing.T) {
	m := &mockCalc{}
	d, _ := NewBatiDispatcher(m)
	cr, rc, err := d.Dispatch(context.Background(), makeEntree("2018"), make([]byte, RecordLen-1), 99, 3)
	if err == nil {
		t.Fatal("expected error")
	}
	if cr != 99 || rc != 3 || m.calls != 0 {
		t.Errorf("cr=%d rc=%d calls=%d", cr, rc, m.calls)
	}
}

func TestDispatchCalculatorError(t *testing.T) {
	boom := errors.New("boom")
	m := &mockCalc{err: boom}
	d, _ := NewBatiDispatcher(m)
	sortie := bytes.Repeat([]byte{'S'}, RecordLen)
	cr, rc, err := d.Dispatch(context.Background(), makeEntree("2018"), sortie, 99, 8)
	if !errors.Is(err, boom) {
		t.Fatalf("err=%v", err)
	}
	if cr != 0 || rc != 8 {
		t.Errorf("cr=%d rc=%d", cr, rc)
	}
	if !bytes.Equal(sortie, bytes.Repeat([]byte{'S'}, RecordLen)) {
		t.Error("sortie modified")
	}
}

func TestDispatchNoStateAcrossCalls(t *testing.T) {
	m := &mockCalc{out: makeOutput('R', 0, 0)}
	d, _ := NewBatiDispatcher(m)
	sortie := make([]byte, RecordLen)
	if _, _, err := d.Dispatch(context.Background(), makeEntree("2018"), sortie, 99, 0); err != nil {
		t.Fatal(err)
	}
	if m.parm != "M" {
		t.Fatalf("parm=%q", m.parm)
	}
	if _, _, err := d.Dispatch(context.Background(), makeEntree("2018"), sortie, 0, 0); err != nil {
		t.Fatal(err)
	}
	if m.parm != " " {
		t.Errorf("parm=%q, PARM must be overwritten each call", m.parm)
	}
}

func TestEAnnee(t *testing.T) {
	tests := []struct {
		in   []byte
		want string
	}{
		{nil, "    "},
		{[]byte("x"), "    "},
		{[]byte("x20"), "20  "},
		{[]byte("x2018rest"), "2018"},
	}
	for _, tt := range tests {
		if got := eAnnee(tt.in); got != tt.want {
			t.Errorf("eAnnee(%q)=%q want %q", tt.in, got, tt.want)
		}
	}
}