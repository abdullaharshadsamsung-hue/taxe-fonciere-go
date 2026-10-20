package nonbati2018

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func validRecord() []byte {
	c := NewCotis()
	return c.Encode()
}

func TestNewCotisEncodeLength(t *testing.T) {
	c := NewCotis()
	b := c.Encode()
	if len(b) != CotisLen {
		t.Fatalf("len=%d", len(b))
	}
	z := Cotis{}
	if len(z.Encode()) != CotisLen {
		t.Fatal("zero value length")
	}
	mb, err := c.MarshalBinary()
	if err != nil || !bytes.Equal(mb, b) {
		t.Fatalf("MarshalBinary mismatch: %v", err)
	}
}

func TestNewCotisTable(t *testing.T) {
	tab := NewCotisTable()
	if len(tab) != PosteCount {
		t.Fatal("count")
	}
	want := NewCotis()
	for i, c := range tab {
		if !bytes.Equal(c.Encode(), want.Encode()) {
			t.Errorf("poste %d differs", i+1)
		}
	}
}

func TestFieldOffsets(t *testing.T) {
	old := SignEncoding
	defer func() { SignEncoding = old }()
	SignEncoding = SignIBM

	type tc struct {
		name  string
		set   func(c *Cotis)
		start int // 1-based
		want  string
	}
	tests := []tc{
		{"JANIPT", func(c *Cotis) { c.SetJanipt(2018) }, 1, "2018"},
		{"CCODEP", func(c *Cotis) { c.SetCcodep("75") }, 5, "75"},
		{"CCODIR", func(c *Cotis) { c.SetCcodir("1") }, 7, "1"},
		{"CCOCOM", func(c *Cotis) { c.SetCcocom(56) }, 8, "056"},
		{"DLETSR", func(c *Cotis) { c.SetDletsr("A") }, 11, "A"},
		{"FILLER1", func(c *Cotis) { c.Filler1 = "Z" }, 12, "Z"},
		{"CGROUP", func(c *Cotis) { c.SetCgroup("P") }, 13, "P"},
		{"DNUMCP", func(c *Cotis) { c.SetDnumcp(123) }, 14, "00123"},
		{"FILLER2", func(c *Cotis) { c.Filler2 = "ab" }, 19, "ab"},
		{"ACODNB", func(c *Cotis) { c.Acodnb = 7 }, 21, "7"},
		{"FILLER3", func(c *Cotis) { c.Filler3 = "123456789" }, 22, "123456789"},
		{"IMPOT", func(c *Cotis) { c.Impot = "I" }, 31, "I"},
		{"GNEXPS", func(c *Cotis) { c.Gnexps = "BN" }, 32, "BN"},
		{"FILLER4", func(c *Cotis) { c.Filler4 = "F4" }, 34, "F4          "},
		{"TAUCOM", func(c *Cotis) { c.Taucom = 12345000 }, 46, "001234500{"},
		{"TAUDEP", func(c *Cotis) { c.Taudep = 1 }, 56, "000000000A"},
		{"TAUREG", func(c *Cotis) { c.Taureg = 2 }, 66, "000000000B"},
		{"TAUSYN", func(c *Cotis) { c.Tausyn = 3 }, 76, "000000000C"},
		{"TAUCU", func(c *Cotis) { c.Taucu = 4 }, 86, "000000000D"},
		{"TAUGEM", func(c *Cotis) { c.Taugem = 5 }, 96, "000000000E"},
		{"TAUTSE", func(c *Cotis) { c.Tautse = 6 }, 106, "000000000F"},
		{"TAUTSEN1", func(c *Cotis) { c.Tautsen[0] = 7 }, 116, "000000000G"},
		{"TAUTSEN2", func(c *Cotis) { c.Tautsen[1] = 8 }, 126, "000000000H"},
		{"PNBTST", func(c *Cotis) { c.Pnbtst = 9 }, 136, "000000000I"},
		{"TAUCHAGR", func(c *Cotis) { c.Tauchagr = -1 }, 146, "000000000J"},
		{"TAUSMAGR", func(c *Cotis) { c.Tausmagr = -10 }, 156, "000000001}"},
		{"TAUBAPSA", func(c *Cotis) { c.Taubapsa = 1 }, 166, "000000000A"},
		{"TAUCAAA", func(c *Cotis) { c.Taucaaa = 1 }, 176, "000000000A"},
		{"TAUTXADD", func(c *Cotis) { c.Tautxadd = 1 }, 186, "000000000A"},
		{"BNBCOM", func(c *Cotis) { c.Bnbcom = 1 }, 196, "000000000A"},
		{"BNBDEP", func(c *Cotis) { c.Bnbdep = 1 }, 206, "000000000A"},
		{"BNBREG", func(c *Cotis) { c.Bnbreg = 1 }, 216, "000000000A"},
		{"BNBSYN", func(c *Cotis) { c.Bnbsyn = 1 }, 226, "000000000A"},
		{"BNBCU", func(c *Cotis) { c.Bnbcu = 1 }, 236, "000000000A"},
		{"BNBGEM", func(c *Cotis) { c.Bnbgem = 1 }, 246, "000000000A"},
		{"BNBTSE", func(c *Cotis) { c.Bnbtse = 1 }, 256, "000000000A"},
		{"BNBTSEN1", func(c *Cotis) { c.Bnbtsen[0] = 1 }, 266, "000000000A"},
		{"BNBTSEN2", func(c *Cotis) { c.Bnbtsen[1] = 1 }, 276, "000000000A"},
		{"BNBAGR", func(c *Cotis) { c.Bnbagr = 1 }, 286, "000000000A"},
		{"BNBBAP", func(c *Cotis) { c.Bnbbap = 1 }, 296, "000000000A"},
		{"BNBCAA", func(c *Cotis) { c.Bnbcaa = 1 }, 306, "000000000A"},
		{"FILLER6", func(c *Cotis) { c.Filler6 = "F6" }, 316, "F6        "},
		{"BNBPLJA", func(c *Cotis) { c.Bnbplja = 1 }, 326, "000000000A"},
		{"COTICOM", func(c *Cotis) { c.Coticom = 1 }, 336, "000000000A"},
		{"COTIDEP", func(c *Cotis) { c.Cotidep = 1 }, 346, "000000000A"},
		{"COTIREG", func(c *Cotis) { c.Cotireg = 1 }, 356, "000000000A"},
		{"COTISYN", func(c *Cotis) { c.Cotisyn = 1 }, 366, "000000000A"},
		{"COTICU", func(c *Cotis) { c.Coticu = 1 }, 376, "000000000A"},
		{"MCOGE3", func(c *Cotis) { c.Mcoge3 = 1 }, 386, "000000000A"},
		{"COTITSE", func(c *Cotis) { c.Cotitse = 1 }, 396, "000000000A"},
		{"COTITSEN1", func(c *Cotis) { c.Cotitsen[0] = 1 }, 406, "000000000A"},
		{"COTITSEN2", func(c *Cotis) { c.Cotitsen[1] = 1 }, 416, "000000000A"},
		{"MCNTSA", func(c *Cotis) { c.Mcntsa = 1 }, 426, "000000000A"},
		{"COTIAGR", func(c *Cotis) { c.Cotiagr = 1 }, 436, "000000000A"},
		{"COTISMAG", func(c *Cotis) { c.Cotismag = 1 }, 446, "000000000A"},
		{"COTIBAP", func(c *Cotis) { c.Cotibap = 1 }, 456, "000000000A"},
		{"DPROCAA", func(c *Cotis) { c.Dprocaa = 1 }, 466, "000000000A"},
		{"DFIXCAA", func(c *Cotis) { c.Dfixcaa = 12 }, 476, "0001B"},
		{"FFIXCAA", func(c *Cotis) { c.Ffixcaa = 1 }, 481, "000000000A"},
		{"FA300", func(c *Cotis) { c.Fa300 = 1 }, 491, "000000000A"},
		{"FN300", func(c *Cotis) { c.Fn300 = 1 }, 501, "000000000A"},
		{"FARB", func(c *Cotis) { c.Farb = 1 }, 511, "000000000A"},
		{"FRNVB", func(c *Cotis) { c.Frnvb = 1 }, 521, "000000000A"},
		{"FARC", func(c *Cotis) { c.Farc = 1 }, 531, "000000000A"},
		{"FRNVC", func(c *Cotis) { c.Frnvc = 1 }, 541, "000000000A"},
		{"FA800", func(c *Cotis) { c.Fa800 = 1 }, 551, "000000000A"},
		{"FN800", func(c *Cotis) { c.Fn800 = 1 }, 561, "000000000A"},
		{"FA900", func(c *Cotis) { c.Fa900 = 1 }, 571, "000000000A"},
		{"FN900", func(c *Cotis) { c.Fn900 = 1 }, 581, "000000000A"},
		{"MBJECO", func(c *Cotis) { c.Mbjeco = 1 }, 591, "000000000A"},
		{"MBJESY", func(c *Cotis) { c.Mbjesy = 1 }, 601, "000000000A"},
		{"MBJEIC", func(c *Cotis) { c.Mbjeic = 1 }, 611, "000000000A"},
		{"MCNJCO", func(c *Cotis) { c.Mcnjco = 1 }, 621, "000000000A"},
		{"MCNJSY", func(c *Cotis) { c.Mcnjsy = 1 }, 631, "000000000A"},
		{"MCNJIC", func(c *Cotis) { c.Mcnjic = 1 }, 641, "000000000A"},
		{"MCNJAE", func(c *Cotis) { c.Mcnjae = 1 }, 651, "000000000A"},
		{"MFAJAE", func(c *Cotis) { c.Mfajae = 1 }, 661, "000000000A"},
		{"MFNJAE", func(c *Cotis) { c.Mfnjae = 1 }, 671, "000000000A"},
		{"FRAIS", func(c *Cotis) { c.FraisTotaux = 42 }, 681, "000000000042"},
		{"CCOIFP", func(c *Cotis) { c.Ccoifp = "123" }, 693, "123"},
		{"CCPPER", func(c *Cotis) { c.Ccpper = "456" }, 696, "456"},
		{"FILLER", func(c *Cotis) { c.Filler = "END" }, 699, "END" + strings.Repeat(" ", 303)},
	}
	baseRec := NewCotis()
	base := baseRec.Encode()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCotis()
			tt.set(&c)
			b := c.Encode()
			if len(b) != CotisLen {
				t.Fatalf("len=%d", len(b))
			}
			s, e := tt.start-1, tt.start-1+len(tt.want)
			if got := string(b[s:e]); got != tt.want {
				t.Errorf("bytes %d..%d = %q want %q", tt.start, e, got, tt.want)
			}
			if !bytes.Equal(b[:s], base[:s]) || !bytes.Equal(b[e:], base[e:]) {
				t.Errorf("other bytes changed")
			}
			p, err := ParseCotis(b)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(p.Encode(), b) {
				t.Error("round trip mismatch")
			}
		})
	}
}

func TestKeyAccessors(t *testing.T) {
	c := Cotis{}
	if _, ok := c.Janipt(); ok {
		t.Error("spaces should not be numeric")
	}
	c.SetJanipt(2018)
	c.SetAcodir("750")
	c.SetCcocom(56)
	c.SetDletsr("A")
	c.SetCgroup("P")
	c.SetDnumcp(42)
	if v, ok := c.Janipt(); !ok || v != 2018 {
		t.Errorf("Janipt %d %v", v, ok)
	}
	if c.Acodir() != "750" || c.Ccodep() != "75" || c.Ccodir() != "0" {
		t.Errorf("acodir %q", c.Acodir())
	}
	if v, ok := c.Ccocom(); !ok || v != 56 || c.CcocomRaw() != "056" {
		t.Errorf("ccocom %d", v)
	}
	if c.Dletsr() != "A" || c.Cgroup() != "P" {
		t.Error("dletsr/cgroup")
	}
	if v, ok := c.Dnumcp(); !ok || v != 42 || c.DnumcpRaw() != "00042" {
		t.Error("dnumcp")
	}
	if c.Cle != "2018750056A" || c.Anupro != "P00042" {
		t.Errorf("groups %q %q", c.Cle, c.Anupro)
	}
	c.SetJanipt(-12018)
	if v, _ := c.Janipt(); v != 2018 {
		t.Errorf("truncation %d", v)
	}
	c.Cle = "20X8"
	if _, ok := c.Janipt(); ok {
		t.Error("non digit numeric")
	}
}

func TestSignModes(t *testing.T) {
	old := SignEncoding
	defer func() { SignEncoding = old }()
	tests := []struct {
		mode SignMode
		v    int64
		want string
	}{
		{SignIBM, 0, "000000000{"},
		{SignIBM, -5, "000000000N"},
		{SignIBM, -20, "000000002}"},
		{SignASCII, 7, "0000000007"},
		{SignASCII, -7, "000000000w"},
		{SignASCII, -10, "000000001p"},
	}
	for _, tt := range tests {
		SignEncoding = tt.mode
		c := NewCotis()
		c.Bnbcom = tt.v
		b := c.Encode()
		if got := string(b[195:205]); got != tt.want {
			t.Errorf("mode %d v %d: %q want %q", tt.mode, tt.v, got, tt.want)
		}
		p, err := ParseCotis(b)
		if err != nil || p.Bnbcom != tt.v {
			t.Errorf("parse %d err %v", p.Bnbcom, err)
		}
	}
}

func TestTruncationAndNormalize(t *testing.T) {
	c := NewCotis()
	c.Dfixcaa = 1234567
	c.Acodnb = -23
	c.FraisTotaux = -5
	c.Bnbcom = 12345678901
	b := c.Encode()
	p, err := ParseCotis(b)
	if err != nil {
		t.Fatal(err)
	}
	if p.Dfixcaa != 34567 || p.Acodnb != 3 || p.FraisTotaux != 5 || p.Bnbcom != 2345678901 {
		t.Errorf("parsed %+v", p)
	}
	c.Normalize()
	if c.Dfixcaa != 34567 || c.Acodnb != 3 || c.FraisTotaux != 5 || c.Bnbcom != 2345678901 {
		t.Error("normalize mismatch")
	}
	c.Cotiagr = -10000000000
	c.Normalize()
	if c.Cotiagr != 0 {
		t.Error("negative zero")
	}
	c.Impot = "TOOLONG"
	if string(c.Encode()[30]) != "T" {
		t.Error("alpha truncation")
	}
}

func TestParseErrors(t *testing.T) {
	mut := func(pos int, ch byte) []byte {
		b := validRecord()
		b[pos-1] = ch
		return b
	}
	tests := []struct {
		name    string
		in      []byte
		numeric bool
		pos     string
	}{
		{"short", make([]byte, 10), false, ""},
		{"long", make([]byte, CotisLen+1), false, ""},
		{"ACODNB", mut(21, 'x'), true, "ACODNB at position 21"},
		{"TAUCOM digit", mut(46, ' '), true, "TAUCOM at position 46"},
		{"TAUCOM sign", mut(55, ' '), true, "TAUCOM at position 46"},
		{"DFIXCAA", mut(480, '?'), true, "DFIXCAA at position 476"},
		{"FRAIS", mut(681, 'A'), true, "FRAIS-TOTAUX at position 681"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseCotis(tt.in)
			if err == nil {
				t.Fatal("expected error")
			}
			if errors.Is(err, ErrNotNumeric) != tt.numeric {
				t.Errorf("ErrNotNumeric mismatch: %v", err)
			}
			if tt.pos != "" && !strings.Contains(err.Error(), tt.pos) {
				t.Errorf("err %v lacks %q", err, tt.pos)
			}
			var c Cotis
			if c.UnmarshalBinary(tt.in) == nil {
				t.Error("UnmarshalBinary should fail")
			}
		})
	}
}

func TestParseAlphaKeyNotValidated(t *testing.T) {
	b := validRecord()
	copy(b[0:4], "ABCD")
	c, err := ParseCotis(b)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := c.Janipt(); ok {
		t.Error("expected non-numeric JANIPT")
	}
}

func TestRateExampleAndUnmarshal(t *testing.T) {
	b := validRecord()
	copy(b[45:55], "001234500{")
	var c Cotis
	if err := c.UnmarshalBinary(b); err != nil {
		t.Fatal(err)
	}
	if c.Taucom != 12345000 {
		t.Errorf("taucom %d", c.Taucom)
	}
	copy(b[45:55], "0012345000")
	c2, err := ParseCotis(b)
	if err != nil || c2.Taucom != 12345000 {
		t.Error("plain digit sign")
	}
}

func TestFullRoundTrip(t *testing.T) {
	old := SignEncoding
	defer func() { SignEncoding = old }()
	for _, m := range []SignMode{SignIBM, SignASCII} {
		SignEncoding = m
		c := NewCotis()
		c.SetJanipt(2018)
		c.Impot, c.Gnexps = "G", "JA"
		c.Tautsen = [CotisTsenCount]int64{1, -2}
		c.Mcnjae = -999
		c.Dfixcaa = -99999
		c.FraisTotaux = 999999999999
		b := c.Encode()
		p, err := ParseCotis(b)
		if err != nil {
			t.Fatal(err)
		}
		if p != c {
			t.Errorf("mode %d: struct mismatch", m)
		}
		if !bytes.Equal(p.Encode(), b) {
			t.Error("bytes mismatch")
		}
	}
}