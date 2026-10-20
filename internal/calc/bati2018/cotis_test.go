package bati2018

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"
)

func blankRecord(t *testing.T) []byte {
	t.Helper()
	c := NewCotis()
	b := c.Encode()
	if len(b) != CotisLen {
		t.Fatalf("NewCotis encode length = %d, want %d", len(b), CotisLen)
	}
	return b
}

func put(b []byte, pos int, s string) { copy(b[pos-1:pos-1+len(s)], s) }

type numField struct {
	name  string
	pos   int
	width int
	get   func(c *Cotis) int64
}

func signedFields() []numField {
	f := []numField{
		{"TAUCOM", 34, 10, func(c *Cotis) int64 { return c.Taucom }},
		{"TAUDEP", 44, 10, func(c *Cotis) int64 { return c.Taudep }},
		{"TAUREG", 54, 10, func(c *Cotis) int64 { return c.Taureg }},
		{"TAUSYN", 64, 10, func(c *Cotis) int64 { return c.Tausyn }},
		{"TAUCU", 74, 10, func(c *Cotis) int64 { return c.Taucu }},
		{"TAUTAS", 84, 10, func(c *Cotis) int64 { return c.Tautas }},
		{"TAUGEM", 94, 10, func(c *Cotis) int64 { return c.Taugem }},
		{"TAUTSE", 104, 10, func(c *Cotis) int64 { return c.Tautse }},
		{"TAUTSEN(1)", 114, 10, func(c *Cotis) int64 { return c.Tautsen[0] }},
		{"TAUTSEN(2)", 124, 10, func(c *Cotis) int64 { return c.Tautsen[1] }},
		{"PTBTST", 134, 10, func(c *Cotis) int64 { return c.Ptbtst }},
		{"BBCOM", 224, 10, func(c *Cotis) int64 { return c.Bbcom }},
		{"BBDEP", 234, 10, func(c *Cotis) int64 { return c.Bbdep }},
		{"BBREG", 244, 10, func(c *Cotis) int64 { return c.Bbreg }},
		{"BBSYN", 254, 10, func(c *Cotis) int64 { return c.Bbsyn }},
		{"BBCU", 264, 10, func(c *Cotis) int64 { return c.Bbcu }},
		{"BBTAS", 274, 10, func(c *Cotis) int64 { return c.Bbtas }},
		{"BBGEM", 284, 10, func(c *Cotis) int64 { return c.Bbgem }},
		{"BBTSE", 294, 10, func(c *Cotis) int64 { return c.Bbtse }},
		{"BBTSEN(1)", 304, 10, func(c *Cotis) int64 { return c.Bbtsen[0] }},
		{"BBTSEN(2)", 314, 10, func(c *Cotis) int64 { return c.Bbtsen[1] }},
		{"COTICOM", 526, 10, func(c *Cotis) int64 { return c.Coticom }},
		{"COTIDEP", 536, 10, func(c *Cotis) int64 { return c.Cotidep }},
		{"COTIREG", 546, 10, func(c *Cotis) int64 { return c.Cotireg }},
		{"COTISYN", 556, 10, func(c *Cotis) int64 { return c.Cotisyn }},
		{"COTICU", 566, 10, func(c *Cotis) int64 { return c.Coticu }},
		{"MCOTA3", 576, 10, func(c *Cotis) int64 { return c.Mcota3 }},
		{"MCOGE3", 586, 10, func(c *Cotis) int64 { return c.Mcoge3 }},
		{"COTITSE", 596, 10, func(c *Cotis) int64 { return c.Cotitse }},
		{"COTITSEN(1)", 606, 10, func(c *Cotis) int64 { return c.Cotitsen[0] }},
		{"COTITSEN(2)", 616, 10, func(c *Cotis) int64 { return c.Cotitsen[1] }},
		{"MCBTSA", 626, 10, func(c *Cotis) int64 { return c.Mcbtsa }},
		{"COTIS-OMI", 760, 10, func(c *Cotis) int64 { return c.CotisOmi }},
		{"FA300", 830, 10, func(c *Cotis) int64 { return c.Fa300 }},
		{"FN300", 840, 10, func(c *Cotis) int64 { return c.Fn300 }},
		{"FA800", 850, 10, func(c *Cotis) int64 { return c.Fa800 }},
		{"FN800", 860, 10, func(c *Cotis) int64 { return c.Fn800 }},
		{"FA900", 870, 10, func(c *Cotis) int64 { return c.Fa900 }},
		{"FN900", 880, 10, func(c *Cotis) int64 { return c.Fn900 }},
	}
	for n := 1; n <= CotisOmCount; n++ {
		i := n - 1
		f = append(f,
			numField{fmt.Sprintf("TAUOM1(%d)", n), 144 + i*10, 10, func(c *Cotis) int64 { return c.Tauom1[i] }},
			numField{fmt.Sprintf("BBTEOM(%d)", n), 324 + i*12 + 2, 10, func(c *Cotis) int64 { return c.TabOm[i].Bbteom }},
			numField{fmt.Sprintf("COTIS-OM(%d)", n), 770 + i*10, 10, func(c *Cotis) int64 { return c.CotisOm[i] }},
		)
	}
	return f
}

func unsignedFields() []numField {
	return []numField{
		{"ACODB", 21, 1, func(c *Cotis) int64 { return c.Acodb }},
		{"TAUOMI", 204, 18, func(c *Cotis) int64 { return c.Tauomi }},
		{"BBTEIOM", 396, 16, func(c *Cotis) int64 { return c.Bbteiom }},
	}
}

func TestParseCotisLength(t *testing.T) {
	tests := []struct {
		name string
		n    int
	}{{"empty", 0}, {"short", CotisLen - 1}, {"long", CotisLen + 1}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := ParseCotis(bytes.Repeat([]byte("0"), tt.n)); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestParseCotisBlank(t *testing.T) {
	c, err := ParseCotis(blankRecord(t))
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range append(signedFields(), unsignedFields()...) {
		if v := f.get(&c); v != 0 {
			t.Errorf("%s = %d, want 0", f.name, v)
		}
	}
}

func TestParseSignedFieldPositions(t *testing.T) {
	for i, f := range signedFields() {
		t.Run(f.name, func(t *testing.T) {
			b := blankRecord(t)
			want := int64(1000000 + i*7)
			put(b, f.pos, fmt.Sprintf("%0*d", f.width, want))
			c, err := ParseCotis(b)
			if err != nil {
				t.Fatal(err)
			}
			if got := f.get(&c); got != want {
				t.Fatalf("got %d, want %d", got, want)
			}
			// neighbouring fields unaffected
			for _, o := range signedFields() {
				if o.name != f.name && o.get(&c) != 0 {
					t.Errorf("%s leaked into %s", f.name, o.name)
				}
			}
		})
	}
}

func TestParseUnsignedFieldPositions(t *testing.T) {
	tests := []struct {
		f   numField
		raw string
		v   int64
	}{
		{unsignedFields()[0], "7", 7},
		{unsignedFields()[1], "000000000012500000", 12500000},
		{unsignedFields()[2], "1234567890123456", 1234567890123456},
	}
	for _, tt := range tests {
		t.Run(tt.f.name, func(t *testing.T) {
			b := blankRecord(t)
			put(b, tt.f.pos, tt.raw)
			c, err := ParseCotis(b)
			if err != nil {
				t.Fatal(err)
			}
			if got := tt.f.get(&c); got != tt.v {
				t.Fatalf("got %d, want %d", got, tt.v)
			}
		})
	}
}

func TestParseAlphaAndKeyFields(t *testing.T) {
	b := blankRecord(t)
	put(b, 1, "2018")
	put(b, 5, "75")
	put(b, 7, "1")
	put(b, 8, "056")
	put(b, 11, "S")
	put(b, 12, "a")
	put(b, 13, "G")
	put(b, 14, "01234")
	put(b, 19, "bc")
	put(b, 22, "fill3xxxx")
	put(b, 31, "I")
	put(b, 32, "GN")
	put(b, 222, "fF")
	put(b, 412, strings.Repeat("5", 114))
	put(b, 636, strings.Repeat("7", 124))
	put(b, 890, "IFP")
	put(b, 893, "PER")
	for n := 1; n <= CotisOmCount; n++ {
		put(b, 324+(n-1)*12, fmt.Sprintf("R%d", n))
	}
	c, err := ParseCotis(b)
	if err != nil {
		t.Fatal(err)
	}
	strTests := []struct{ name, got, want string }{
		{"CLE", c.Cle, "2018751056S"},
		{"ACODIR", c.Acodir(), "751"},
		{"CCODEP", c.Ccodep(), "75"},
		{"CCODIR", c.Ccodir(), "1"},
		{"CCOCOMRaw", c.CcocomRaw(), "056"},
		{"DLETSR", c.Dletsr(), "S"},
		{"FILLER1", c.Filler1, "a"},
		{"ANUPRO", c.Anupro, "G01234"},
		{"CGROUP", c.Cgroup(), "G"},
		{"DNUMCPRaw", c.DnumcpRaw(), "01234"},
		{"FILLER2", c.Filler2, "bc"},
		{"FILLER3", c.Filler3, "fill3xxxx"},
		{"IMPOT", c.Impot, "I"},
		{"GNEXPL", c.Gnexpl, "GN"},
		{"FILLER", c.Filler, "fF"},
		{"FILLER5", c.Filler5, strings.Repeat("5", 114)},
		{"FILLER7", c.Filler7, strings.Repeat("7", 124)},
		{"CCOIFP", c.Ccoifp, "IFP"},
		{"CCPPER", c.Ccpper, "PER"},
	}
	for n := 1; n <= CotisOmCount; n++ {
		strTests = append(strTests, struct{ name, got, want string }{
			fmt.Sprintf("GTAUOM(%d)", n), c.TabOm[n-1].Gtauom, fmt.Sprintf("R%d", n)})
	}
	for _, tt := range strTests {
		if tt.got != tt.want {
			t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.want)
		}
	}
	numTests := []struct {
		name string
		f    func() (int64, bool)
		want int64
	}{
		{"JANIPT", c.Janipt, 2018},
		{"CCOCOM", c.Ccocom, 56},
		{"DNUMCP", c.Dnumcp, 1234},
	}
	for _, tt := range numTests {
		v, ok := tt.f()
		if !ok || v != tt.want {
			t.Errorf("%s = %d,%v want %d,true", tt.name, v, ok, tt.want)
		}
	}
}

func TestKeyNumericNotNumeric(t *testing.T) {
	tests := []struct {
		name string
		pos  int
		raw  string
		f    func(c *Cotis) (int64, bool)
	}{
		{"JANIPT", 1, "20A8", (*Cotis).Janipt},
		{"CCOCOM", 8, "0 5", (*Cotis).Ccocom},
		{"DNUMCP", 14, "12X45", (*Cotis).Dnumcp},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := blankRecord(t)
			put(b, tt.pos, tt.raw)
			c, err := ParseCotis(b)
			if err != nil {
				t.Fatal(err)
			}
			if _, ok := tt.f(&c); ok {
				t.Fatal("expected NUMERIC test to fail")
			}
		})
	}
	var z Cotis
	if _, ok := z.Janipt(); ok {
		t.Error("zero-value JANIPT should not be numeric")
	}
}

func TestParseNotNumericErrors(t *testing.T) {
	tests := []struct {
		name, field string
		pos         int
		raw         string
		errPos      int
	}{
		{"ACODB letter", "ACODB", 21, "A", 21},
		{"ACODB space", "ACODB", 21, " ", 21},
		{"TAUOMI", "TAUOMI", 210, "x", 204},
		{"BBTEIOM", "BBTEIOM", 396, " ", 396},
		{"TAUCOM bad digit", "TAUCOM", 34, "X", 34},
		{"TAUCOM bad sign", "TAUCOM", 43, "Z", 34},
		{"BBTEOM(3) bad sign", "BBTEOM(3)", 324 + 24 + 2 + 9, " ", 324 + 24 + 2},
		{"FN900 bad sign", "FN900", 889, "z", 880},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := blankRecord(t)
			put(b, tt.pos, tt.raw)
			_, err := ParseCotis(b)
			if err == nil {
				t.Fatal("expected error")
			}
			if !errors.Is(err, errNotNumeric) {
				t.Errorf("error %v is not errNotNumeric", err)
			}
			want := fmt.Sprintf("COTISB-%s at position %d", tt.field, tt.errPos)
			if !strings.Contains(err.Error(), want) {
				t.Errorf("error %q does not contain %q", err, want)
			}
		})
	}
}

func TestParseOverpunch(t *testing.T) {
	tests := []struct {
		last byte
		want int64
	}{
		{'5', 15}, {'{', 10}, {'A', 11}, {'I', 19},
		{'}', -10}, {'J', -11}, {'R', -19},
		{'p', -10}, {'q', -11}, {'y', -19},
	}
	for _, tt := range tests {
		t.Run(string(tt.last), func(t *testing.T) {
			b := blankRecord(t)
			put(b, 224, "000000001"+string(tt.last))
			c, err := ParseCotis(b)
			if err != nil {
				t.Fatal(err)
			}
			if c.Bbcom != tt.want {
				t.Fatalf("got %d want %d", c.Bbcom, tt.want)
			}
		})
	}
}

func TestEncodeNumeric(t *testing.T) {
	tests := []struct {
		name string
		set  func(c *Cotis)
		pos  int
		want string
	}{
		{"rate 1.5 scaled", func(c *Cotis) { c.Taucom = 1500000 }, 34, "000150000{"},
		{"positive", func(c *Cotis) { c.Taucom = 123 }, 34, "000000012C"},
		{"negative", func(c *Cotis) { c.Taucom = -123 }, 34, "000000012L"},
		{"negative ending 0", func(c *Cotis) { c.Bbcom = -10 }, 224, "000000001}"},
		{"zero", func(c *Cotis) { c.Bbcom = 0 }, 224, "000000000{"},
		{"truncated high-order", func(c *Cotis) { c.Taucom = 12345678901 }, 34, "234567890A"},
		{"negative zero after truncation", func(c *Cotis) { c.Fn900 = -10000000000 }, 880, "000000000{"},
		{"unsigned drops sign TAUOMI", func(c *Cotis) { c.Tauomi = -5 }, 204, "000000000000000005"},
		{"unsigned drops sign ACODB", func(c *Cotis) { c.Acodb = -3 }, 21, "3"},
		{"BBTEIOM", func(c *Cotis) { c.Bbteiom = 500000000000000 }, 396, "0500000000000000"},
		{"COTIS-OM(6)", func(c *Cotis) { c.CotisOm[5] = 9 }, 820, "000000000I"},
		{"TAUOM1(6)", func(c *Cotis) { c.Tauom1[5] = -1 }, 194, "000000000J"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCotis()
			tt.set(&c)
			b := c.Encode()
			if len(b) != CotisLen {
				t.Fatalf("len %d", len(b))
			}
			if got := string(b[tt.pos-1 : tt.pos-1+len(tt.want)]); got != tt.want {
				t.Fatalf("got %q want %q", got, tt.want)
			}
		})
	}
}

func TestEncodeUnsignedNeverNegative(t *testing.T) {
	c := NewCotis()
	c.Acodb, c.Tauomi, c.Bbteiom = -1, -2, -3
	c.SetJanipt(-2018)
	c.SetCcocom(-5)
	c.SetDnumcp(-42)
	p, err := ParseCotis(c.Encode())
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range unsignedFields() {
		if f.get(&p) < 0 {
			t.Errorf("%s negative", f.name)
		}
	}
	if v, ok := p.Janipt(); !ok || v != 2018 {
		t.Errorf("JANIPT %d %v", v, ok)
	}
	if v, ok := p.Ccocom(); !ok || v != 5 {
		t.Errorf("CCOCOM %d %v", v, ok)
	}
	if v, ok := p.Dnumcp(); !ok || v != 42 {
		t.Errorf("DNUMCP %d %v", v, ok)
	}
}

func TestEncodeAlpha(t *testing.T) {
	tests := []struct {
		name string
		set  func(c *Cotis)
		pos  int
		want string
	}{
		{"truncate", func(c *Cotis) { c.Impot = "XYZ" }, 31, "X"},
		{"pad", func(c *Cotis) { c.Gnexpl = "A" }, 32, "A "},
		{"pad ccpper", func(c *Cotis) { c.Ccpper = "ab" }, 893, "ab "},
		{"gtauom", func(c *Cotis) { c.TabOm[2].Gtauom = "P" }, 348, "P "},
		{"zero value cle", func(c *Cotis) { c.Cle = "" }, 1, strings.Repeat(" ", 11)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCotis()
			tt.set(&c)
			b := c.Encode()
			if got := string(b[tt.pos-1 : tt.pos-1+len(tt.want)]); got != tt.want {
				t.Fatalf("got %q want %q", got, tt.want)
			}
		})
	}
	var z Cotis
	if n := len(z.Encode()); n != CotisLen {
		t.Errorf("zero value encode length %d", n)
	}
}

func TestNormalizeMatchesEncode(t *testing.T) {
	c := NewCotis()
	c.Taucom = 12345678901
	c.Bbcom = -99999999999
	c.Tauomi = -1234567890123456789
	c.Bbteiom = 12345678901234567
	c.Acodb = 15
	c.Normalize()
	want := []struct {
		name      string
		got, want int64
	}{
		{"Taucom", c.Taucom, 2345678901},
		{"Bbcom", c.Bbcom, -9999999999},
		{"Tauomi", c.Tauomi, 234567890123456789},
		{"Bbteiom", c.Bbteiom, 2345678901234567},
		{"Acodb", c.Acodb, 5},
	}
	for _, w := range want {
		if w.got != w.want {
			t.Errorf("%s = %d want %d", w.name, w.got, w.want)
		}
	}
	p, err := ParseCotis(c.Encode())
	if err != nil {
		t.Fatal(err)
	}
	if p != c {
		t.Errorf("normalized value differs after round trip")
	}
}

func populated() Cotis {
	c := NewCotis()
	c.SetJanipt(2018)
	c.SetCcodep("2A")
	c.SetCcodir("0")
	c.SetCcocom(4)
	c.SetDletsr("B")
	c.SetCgroup("+")
	c.SetDnumcp(98765)
	c.Filler1, c.Filler2, c.Filler3 = "x", "yz", "123456789"
	c.Acodb, c.Impot, c.Gnexpl = 2, "I", "GN"
	c.Taucom, c.Taudep, c.Taureg, c.Tausyn = 21500000, -1, 0, 9999999999
	c.Taucu, c.Tautas, c.Taugem, c.Tautse = 5, 6, 7, 8
	c.Tautsen = [2]int64{-20, 30}
	c.Ptbtst = 12
	c.Tauom1 = [6]int64{1, -2, 3, -4, 5, -6}
	c.Tauomi = 999999999999999999
	c.Filler = "ff"
	c.Bbcom, c.Bbdep, c.Bbreg, c.Bbsyn, c.Bbcu = 100, -200, 300, 400, 500
	c.Bbtas, c.Bbgem, c.Bbtse = 600, 700, 800
	c.Bbtsen = [2]int64{900, -1000}
	for i := range c.TabOm {
		c.TabOm[i] = OmZone{Gtauom: fmt.Sprintf("R%c", 'A'+i), Bbteom: int64(i*111 - 200)}
	}
	c.Bbteiom = 123456789012345
	c.Filler5 = strings.Repeat("5", 114)
	c.Coticom, c.Cotidep, c.Cotireg, c.Cotisyn, c.Coticu = 1, 2, 3, 4, -5
	c.Mcota3, c.Mcoge3, c.Cotitse = 6, 7, 8
	c.Cotitsen = [2]int64{9, 10}
	c.Mcbtsa = 11
	c.Filler7 = strings.Repeat("7", 124)
	c.CotisOmi = 12
	c.CotisOm = [6]int64{13, 14, 15, 16, 17, 18}
	c.Fa300, c.Fn300, c.Fa800, c.Fn800, c.Fa900, c.Fn900 = 19, 20, 21, -22, 23, 24
	c.Ccoifp, c.Ccpper = "IFP", "PER"
	return c
}

func TestRoundTrip(t *testing.T) {
	src := populated()
	rec := src.Encode()
	tests := []struct {
		name string
		rec  []byte
	}{{"populated", rec}, {"initialized", blankRecord(t)}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := ParseCotis(tt.rec)
			if err != nil {
				t.Fatal(err)
			}
			if out := c.Encode(); !bytes.Equal(out, tt.rec) {
				t.Fatalf("round trip mismatch\n got %q\nwant %q", out, tt.rec)
			}
		})
	}
	c, _ := ParseCotis(rec)
	if c != src {
		t.Errorf("parsed struct differs from source")
	}
}

func TestMarshalUnmarshalBinary(t *testing.T) {
	src := populated()
	b, err := src.MarshalBinary()
	if err != nil || len(b) != CotisLen {
		t.Fatalf("MarshalBinary: %v len %d", err, len(b))
	}
	var c Cotis
	if err := c.UnmarshalBinary(b); err != nil {
		t.Fatal(err)
	}
	if c != src {
		t.Error("unmarshal mismatch")
	}
	before := c
	if err := c.UnmarshalBinary(b[:10]); err == nil {
		t.Error("expected error on short input")
	}
	if c != before {
		t.Error("failed unmarshal modified receiver")
	}
}

func TestNewCotisInitialize(t *testing.T) {
	c := NewCotis()
	if v, ok := c.Janipt(); !ok || v != 0 {
		t.Errorf("JANIPT %d %v", v, ok)
	}
	if v, ok := c.Dnumcp(); !ok || v != 0 {
		t.Errorf("DNUMCP %d %v", v, ok)
	}
	if c.Ccodep() != "  " || c.Dletsr() != " " || c.Cgroup() != " " {
		t.Error("alpha key items not spaces")
	}
}

func TestSetters(t *testing.T) {
	c := NewCotis()
	c.SetCcocom(12345)
	if c.CcocomRaw() != "345" {
		t.Errorf("CcocomRaw %q", c.CcocomRaw())
	}
	c.SetAcodir("ABCD")
	if c.Acodir() != "ABC" || c.Ccodep() != "AB" || c.Ccodir() != "C" {
		t.Errorf("Acodir %q", c.Acodir())
	}
	c.SetCcodep("9")
	if c.Ccodep() != "9 " {
		t.Errorf("Ccodep %q", c.Ccodep())
	}
	if len(c.Cle) != CotisCleLen || len(c.Anupro) != CotisAnuproLen {
		t.Error("group length changed")
	}
}

func TestCotisOmSubscript(t *testing.T) {
	c := populated()
	tests := []struct {
		ind     int
		want    int64
		wantErr bool
	}{
		{1, 13, false}, {6, 18, false}, {7, 19, false},
		{0, 0, true}, {8, 0, true}, {-1, 0, true},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprint(tt.ind), func(t *testing.T) {
			v, err := c.CotisOmSubscript(tt.ind)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err %v", err)
			}
			if v != tt.want {
				t.Fatalf("got %d want %d", v, tt.want)
			}
		})
	}
}

func TestOccursCounts(t *testing.T) {
	var c Cotis
	if len(c.Tautsen) != 2 || len(c.Bbtsen) != 2 || len(c.Cotitsen) != 2 {
		t.Error("TSEN counts")
	}
	if len(c.Tauom1) != 6 || len(c.TabOm) != 6 || len(c.CotisOm) != 6 {
		t.Error("OM counts")
	}
}