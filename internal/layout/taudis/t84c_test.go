package taudis

import (
	"strings"
	"testing"
)

func TestCommuneLayoutLengths(t *testing.T) {
	if BodyLen != 2988 || RecordLen != 3000 || EnteteLen+BodyLen != RecordLen {
		t.Fatalf("bad lengths")
	}
	c := NewCommune()
	if len(c.Bytes()) != 2988 {
		t.Fatalf("bytes len %d", len(c.Bytes()))
	}
	if SyndicateCount != 10 || AgrjaCount != 5 {
		t.Fatal("occurs counts")
	}
}

func TestNewCommuneDefaults(t *testing.T) {
	c := NewCommune()
	if !c.IsCommune() || c.Dletar() != "C" {
		t.Fatal("dletar")
	}
	b := c.Bytes()
	if b[1] != 'C' {
		t.Fatal("second byte not C")
	}
	if b[0] != ' ' || b[2] != ' ' {
		t.Fatal("separator not space")
	}
	if v, ok := c.Dnuro(); v != 0 || !ok {
		t.Fatal("dnuro default")
	}
	if c.DnuroRaw() != "000" || c.Jdaho1() != "00000000" {
		t.Fatal("zero fill")
	}
	if c.Jdami1() != strings.Repeat(" ", 8) {
		t.Fatal("jdami1 should be spaces")
	}
	if c.Libre() != strings.Repeat(" ", 2269) {
		t.Fatal("libre")
	}
	for n := 1; n <= 5; n++ {
		v, ok, err := c.Jgcin(n)
		if err != nil || !ok || v != 0 {
			t.Fatalf("jgcin %d", n)
		}
	}
	if v, ok := c.Cncom(); v != 0 || !ok {
		t.Fatal("cncom")
	}
}

func TestParseCommune(t *testing.T) {
	tests := []struct {
		name string
		n    int
		err  bool
	}{
		{"exact", BodyLen, false},
		{"short", BodyLen - 1, true},
		{"long", BodyLen + 1, true},
		{"empty", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseCommune(make([]byte, tt.n))
			if (err != nil) != tt.err {
				t.Fatalf("err=%v", err)
			}
		})
	}
}

func TestCommuneRoundTrip(t *testing.T) {
	b := make([]byte, BodyLen)
	for i := range b {
		b[i] = byte(i%94 + 33)
	}
	c, err := ParseCommune(b)
	if err != nil {
		t.Fatal(err)
	}
	if string(c.Bytes()) != string(b) {
		t.Fatal("round trip mismatch")
	}
	out := c.Bytes()
	out[0] = 'Z'
	if c[0] == 'Z' {
		t.Fatal("Bytes aliases")
	}
}

func TestAlphaFieldOffsets(t *testing.T) {
	tests := []struct {
		name string
		off  int
		n    int
		set  func(*Commune, string)
		get  func(*Commune) string
	}{
		{"Dletar", 1, 1, (*Commune).SetDletar, (*Commune).Dletar},
		{"Dnultf", 3, 1, (*Commune).SetDnultf, (*Commune).Dnultf},
		{"Gcoanu", 9, 1, (*Commune).SetGcoanu, (*Commune).Gcoanu},
		{"Dacrol", 11, 3, (*Commune).SetDacrol, (*Commune).Dacrol},
		{"Gcovef", 15, 1, (*Commune).SetGcovef, (*Commune).Gcovef},
		{"Gorvef", 17, 1, (*Commune).SetGorvef, (*Commune).Gorvef},
		{"Ccofdl", 19, 1, (*Commune).SetCcofdl, (*Commune).Ccofdl},
		{"Gokcu", 21, 1, (*Commune).SetGokcu, (*Commune).Gokcu},
		{"Gtocfu", 31, 1, (*Commune).SetGtocfu, (*Commune).Gtocfu},
		{"Gcndif", 33, 1, (*Commune).SetGcndif, (*Commune).Gcndif},
		{"Gcbdif", 35, 1, (*Commune).SetGcbdif, (*Commune).Gcbdif},
		{"Jdaho1", 37, 8, (*Commune).SetJdaho1, (*Commune).Jdaho1},
		{"Jdami1", 46, 8, (*Commune).SetJdami1, (*Commune).Jdami1},
		{"Jssmir", 46, 2, (*Commune).SetJssmir, (*Commune).Jssmir},
		{"Janmir", 48, 2, (*Commune).SetJanmir, (*Commune).Janmir},
		{"Jmomir", 50, 2, (*Commune).SetJmomir, (*Commune).Jmomir},
		{"Jjrmir", 52, 2, (*Commune).SetJjrmir, (*Commune).Jjrmir},
		{"Jdamjo", 55, 8, (*Commune).SetJdamjo, (*Commune).Jdamjo},
		{"Jssmjo", 55, 2, (*Commune).SetJssmjo, (*Commune).Jssmjo},
		{"Janmjo", 57, 2, (*Commune).SetJanmjo, (*Commune).Janmjo},
		{"Jmomjo", 59, 2, (*Commune).SetJmomjo, (*Commune).Jmomjo},
		{"Jjrmjo", 61, 2, (*Commune).SetJjrmjo, (*Commune).Jjrmjo},
		{"Dlicom", 64, 30, (*Commune).SetDlicom, (*Commune).Dlicom},
		{"Dsirec", 95, 14, (*Commune).SetDsirec, (*Commune).Dsirec},
		{"Gnidep", 110, 1, (*Commune).SetGnidep, (*Commune).Gnidep},
		{"Gtomgp", 112, 1, (*Commune).SetGtomgp, (*Commune).Gtomgp},
		{"Gcepci", 114, 1, (*Commune).SetGcepci, (*Commune).Gcepci},
		{"Gtidf", 116, 1, (*Commune).SetGtidf, (*Commune).Gtidf},
		{"Ctlorr", 118, 1, (*Commune).SetCtlorr, (*Commune).Ctlorr},
		{"Gsvrec", 120, 1, (*Commune).SetGsvrec, (*Commune).Gsvrec},
		{"Ccobmj", 122, 4, (*Commune).SetCcobmj, (*Commune).Ccobmj},
		{"Gmjpos", 135, 1, (*Commune).SetGmjpos, (*Commune).Gmjpos},
		{"Dcmzo1", 141, 1, (*Commune).SetDcmzo1, (*Commune).Dcmzo1},
		{"Grocom", 189, 1, (*Commune).SetGrocom, (*Commune).Grocom},
		{"Pcfidf", 191, 3, (*Commune).SetPcfidf, (*Commune).Pcfidf},
		{"Gtocom", 195, 1, (*Commune).SetGtocom, (*Commune).Gtocom},
		{"Acogc", 197, 4, (*Commune).SetAcogc, (*Commune).Acogc},
		{"Ctygc1", 197, 1, (*Commune).SetCtygc1, (*Commune).Ctygc1},
		{"Rfepci", 202, 3, (*Commune).SetRfepci, (*Commune).Rfepci},
		{"Gtocob", 206, 1, (*Commune).SetGtocob, (*Commune).Gtocob},
		{"Gtocon", 208, 1, (*Commune).SetGtocon, (*Commune).Gtocon},
		{"Vmailg", 569, 50, (*Commune).SetVmailg, (*Commune).Vmailg},
		{"Vmaild", 619, 50, (*Commune).SetVmaild, (*Commune).Vmaild},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCommune()
			before := c
			v := strings.Repeat("Q", tt.n)
			tt.set(&c, v)
			if tt.get(&c) != v {
				t.Fatalf("got %q", tt.get(&c))
			}
			for i := range c {
				in := i >= tt.off && i < tt.off+tt.n
				if in && c[i] != 'Q' {
					t.Fatalf("byte %d not written", i)
				}
				if !in && c[i] != before[i] {
					t.Fatalf("byte %d clobbered", i)
				}
			}
			tt.set(&c, v+"ZZ")
			if tt.get(&c) != v {
				t.Fatal("not truncated")
			}
			tt.set(&c, "")
			if tt.get(&c) != strings.Repeat(" ", tt.n) {
				t.Fatal("not padded")
			}
		})
	}
}

func TestNumericFieldOffsets(t *testing.T) {
	tests := []struct {
		name string
		off  int
		n    int
		set  func(*Commune, uint64)
		get  func(*Commune) (uint64, bool)
	}{
		{"Dnuro", 5, 3, (*Commune).SetDnuro, (*Commune).Dnuro},
		{"Nbiftr", 23, 3, (*Commune).SetNbiftr, (*Commune).Nbiftr},
		{"Nbrifp", 27, 3, (*Commune).SetNbrifp, (*Commune).Nbrifp},
		{"Jsshom", 37, 2, (*Commune).SetJsshom, (*Commune).Jsshom},
		{"Janho1", 39, 2, (*Commune).SetJanho1, (*Commune).Janho1},
		{"Jmohom", 41, 2, (*Commune).SetJmohom, (*Commune).Jmohom},
		{"Jjrhom", 43, 2, (*Commune).SetJjrhom, (*Commune).Jjrhom},
		{"Npocom", 127, 7, (*Commune).SetNpocom, (*Commune).Npocom},
		{"Pmjpo", 137, 3, (*Commune).SetPmjpo, (*Commune).Pmjpo},
		{"Mfsid1", 143, 10, (*Commune).SetMfsid1, (*Commune).Mfsid1},
		{"Dvlpom", 154, 6, (*Commune).SetDvlpom, (*Commune).Dvlpom},
		{"Pplom", 161, 3, (*Commune).SetPplom, (*Commune).Pplom},
		{"Ntxom", 165, 1, (*Commune).SetNtxom, (*Commune).Ntxom},
		{"Ntxom1", 167, 1, (*Commune).SetNtxom1, (*Commune).Ntxom1},
		{"Dsedom", 169, 9, (*Commune).SetDsedom, (*Commune).Dsedom},
		{"Dsednb", 179, 9, (*Commune).SetDsednb, (*Commune).Dsednb},
		{"Ccoic", 198, 3, (*Commune).SetCcoic, (*Commune).Ccoic},
		{"Gcoetp", 210, 3, (*Commune).SetGcoetp, (*Commune).Gcoetp},
		{"Cncom", 670, 9, (*Commune).SetCncom, (*Commune).Cncom},
		{"Cnepci", 680, 9, (*Commune).SetCnepci, (*Commune).Cnepci},
		{"Cndep", 690, 9, (*Commune).SetCndep, (*Commune).Cndep},
		{"Cnteom", 700, 9, (*Commune).SetCnteom, (*Commune).Cnteom},
		{"Cntse", 710, 9, (*Commune).SetCntse, (*Commune).Cntse},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCommune()
			before := c
			tt.set(&c, 7)
			want := strings.Repeat("0", tt.n-1) + "7"
			if string(c[tt.off:tt.off+tt.n]) != want {
				t.Fatalf("raw %q", string(c[tt.off:tt.off+tt.n]))
			}
			for i := range c {
				if (i < tt.off || i >= tt.off+tt.n) && c[i] != before[i] {
					t.Fatalf("byte %d clobbered", i)
				}
			}
			if v, ok := tt.get(&c); v != 7 || !ok {
				t.Fatalf("got %d %v", v, ok)
			}
			tt.set(&c, 12345678901)
			p := uint64(1)
			for i := 0; i < tt.n; i++ {
				p *= 10
			}
			if v, _ := tt.get(&c); v != 12345678901%p {
				t.Fatalf("truncation got %d", v)
			}
			c[tt.off] = ' '
			if _, ok := tt.get(&c); ok {
				t.Fatal("expected not numeric")
			}
		})
	}
}

func TestScaledValues(t *testing.T) {
	c := NewCommune()
	copy(c[161:164], "125")
	if v, ok := c.Pplom(); !ok || v != 125 || float64(v)/PplomScale != 1.25 {
		t.Fatal("pplom")
	}
	copy(c[670:679], "001000000")
	if v, ok := c.Cncom(); !ok || v != RateScale {
		t.Fatal("cncom")
	}
}

func TestInvalidNumericDecode(t *testing.T) {
	c := NewCommune()
	copy(c[5:8], "1A2")
	v, ok := c.Dnuro()
	if ok {
		t.Fatal("ok should be false")
	}
	if v != 112 {
		t.Fatalf("got %d", v)
	}
	copy(c[5:8], "1?2")
	if v, _ := c.Dnuro(); v != 192 {
		t.Fatalf("clamp got %d", v)
	}
	if c.DnuroRaw() != "1?2" {
		t.Fatal("raw")
	}
}

func TestGroupsShareStorage(t *testing.T) {
	c := NewCommune()
	c.SetJdaho1("20180415")
	for _, tc := range []struct {
		f    func() (uint64, bool)
		want uint64
	}{{c.Jsshom, 20}, {c.Janho1, 18}, {c.Jmohom, 4}, {c.Jjrhom, 15}} {
		if v, ok := tc.f(); !ok || v != tc.want {
			t.Fatalf("got %d want %d", v, tc.want)
		}
	}
	c.SetJdami1("20190102")
	if c.Jssmir()+c.Janmir()+c.Jmomir()+c.Jjrmir() != "20190102" {
		t.Fatal("jdami1")
	}
	c.SetJdamjo("20200304")
	if c.Jssmjo()+c.Janmjo()+c.Jmomjo()+c.Jjrmjo() != "20200304" {
		t.Fatal("jdamjo")
	}
	c.SetAcogc("G123")
	if c.Ctygc1() != "G" || c.CcoicRaw() != "123" {
		t.Fatal("acogc")
	}
	c.SetMfsid1(42)
	if c.Mfsidf() != "0000000042" {
		t.Fatal("mfsidf")
	}
	c.SetGcoetp(9)
	if c.GcoetpRaw() != "009" {
		t.Fatal("gcoetp raw")
	}
}

func TestSyndicates(t *testing.T) {
	c := NewCommune()
	for i := 0; i < SyndicateCount; i++ {
		s, err := c.Syndicate(i)
		if err != nil {
			t.Fatal(err)
		}
		s.SetSirenSyn("123456789X")
		s.SetDlisyn("SYND")
		base := 213 + i*31
		if string(c[base+1:base+10]) != "123456789" {
			t.Fatalf("siren entry %d", i)
		}
		if s.Dlisyn() != "SYND"+strings.Repeat(" ", 16) || string(c[base+11:base+15]) != "SYND" {
			t.Fatalf("dlisyn %d", i)
		}
		if c[base] != ' ' || c[base+10] != ' ' {
			t.Fatal("separator")
		}
	}
	for _, i := range []int{-1, SyndicateCount} {
		if _, err := c.Syndicate(i); err == nil {
			t.Fatalf("expected error for %d", i)
		}
	}
}

func TestAgrjaRedefines(t *testing.T) {
	c := NewCommune()
	for i := 0; i < AgrjaCount; i++ {
		a, err := c.Agrja(i)
		if err != nil {
			t.Fatal(err)
		}
		a.SetGrinsn(uint64(2010 + i))
		a.SetGrrimn(uint64(2020 + i))
		base := 523 + i*9
		if string(c[base+1:base+5]) != a.GrinsnRaw() || a.GrrimnRaw() != string(c[base+5:base+9]) {
			t.Fatal("raw")
		}
	}
	named := []struct {
		in, ri func() (uint64, bool)
		idx    int
	}{
		{c.Jgcin1, c.Jgcri1, 4}, {c.Jgcin2, c.Jgcri2, 3}, {c.Jgcin3, c.Jgcri3, 2},
		{c.Jgcin4, c.Jgcri4, 1}, {c.Jgcin5, c.Jgcri5, 0},
	}
	for n, tc := range named {
		v, ok := tc.in()
		if !ok || v != uint64(2010+tc.idx) {
			t.Fatalf("jgcin%d=%d", n+1, v)
		}
		r, _ := tc.ri()
		if r != uint64(2020+tc.idx) {
			t.Fatalf("jgcri%d=%d", n+1, r)
		}
		v2, _, err := c.Jgcin(n + 1)
		r2, _, err2 := c.Jgcri(n + 1)
		if err != nil || err2 != nil || v2 != v || r2 != r {
			t.Fatal("indexed mismatch")
		}
	}
	if err := c.SetJgcin(1, 1999); err != nil {
		t.Fatal(err)
	}
	if err := c.SetJgcri(5, 1888); err != nil {
		t.Fatal(err)
	}
	a4, _ := c.Agrja(4)
	a0, _ := c.Agrja(0)
	if v, _ := a4.Grinsn(); v != 1999 {
		t.Fatal("setjgcin")
	}
	if v, _ := a0.Grrimn(); v != 1888 {
		t.Fatal("setjgcri")
	}
	for _, n := range []int{0, 6, -1} {
		if _, _, err := c.Jgcin(n); err == nil {
			t.Fatal("jgcin err")
		}
		if _, _, err := c.Jgcri(n); err == nil {
			t.Fatal("jgcri err")
		}
		if c.SetJgcin(n, 1) == nil || c.SetJgcri(n, 1) == nil {
			t.Fatal("set err")
		}
	}
	for _, i := range []int{-1, AgrjaCount} {
		if _, err := c.Agrja(i); err == nil {
			t.Fatal("agrja err")
		}
	}
}

func TestFillersPreserved(t *testing.T) {
	b := make([]byte, BodyLen)
	for i := range b {
		b[i] = '#'
	}
	c, _ := ParseCommune(b)
	c.SetDlicom("PARIS")
	c.SetNpocom(100)
	c.SetVmaild("x")
	if c[63] != '#' || c[126] != '#' || c[618] != '#' || c[669] != '#' {
		t.Fatal("separator modified")
	}
	if c.Libre() != strings.Repeat("#", 2269) {
		t.Fatal("libre modified")
	}
	if c.IsCommune() {
		t.Fatal("should not be commune")
	}
}