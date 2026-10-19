package linkage

import (
	"strings"
	"testing"
)

func TestRetBNewInitialize(t *testing.T) {
	r := NewRetB()
	if len(r.Bytes()) != RetBLen {
		t.Fatalf("len %d", len(r.Bytes()))
	}
	if r.Aidfic() != "0000       " {
		t.Errorf("aidfic %q", r.Aidfic())
	}
	if r.Anupro() != " 00000" {
		t.Errorf("anupro %q", r.Anupro())
	}
	if r.MctcomRaw() != "000000000{" {
		t.Errorf("mctcom raw %q", r.MctcomRaw())
	}
	if r.McbtsaRaw() != "0000000000" || r.TctomRaw() != "0000000000" {
		t.Errorf("unsigned raw")
	}
	if r.Filler() != strings.Repeat(" ", 40) || r.Trailer() != strings.Repeat(" ", 238) {
		t.Errorf("fillers not spaces")
	}
	for i := 0; i < RetBOmCount; i++ {
		if v, ok, err := r.Mctom(i); err != nil || !ok || v != 0 {
			t.Errorf("mctom %d", i)
		}
		if g, _ := r.Gtauom(i); g != "  " {
			t.Errorf("gtauom %d", i)
		}
	}
	for i := 0; i < RetBBt13Count; i++ {
		if v, ok, err := r.Mcbt13(i); err != nil || !ok || v != 0 {
			t.Errorf("mcbt13 %d", i)
		}
	}
	if v, ok := r.Tctdu(); !ok || v != 0 {
		t.Errorf("tctdu")
	}
}

func TestParseRetBLength(t *testing.T) {
	for _, n := range []int{0, 599, 600, 601} {
		_, err := ParseRetB(make([]byte, n))
		if (err == nil) != (n == RetBLen) {
			t.Errorf("len %d err %v", n, err)
		}
	}
}

func TestRetBParseOffsets(t *testing.T) {
	b := []byte(strings.Repeat(" ", RetBLen))
	put := func(off int, s string) { copy(b[off:], s) }
	put(0, "2024")
	put(4, "75")
	put(6, "1")
	put(7, "056")
	put(10, "A")
	put(11, "B")
	put(12, "01234")
	put(17, "X")
	put(18, "000000012J")
	put(28, "000000010}")
	put(38, "0000000003")
	put(78, "000000001A")
	put(88, "000000002K")
	put(98, "0000000099")
	put(108, "FILL")
	put(148+12*5, "RA000000005E")
	put(220, "000000007G")
	put(230, "000000008H")
	put(240, "000000009I")
	put(250, "000000001{")
	put(260, "00000000123C")
	put(272, "00000000000}")
	put(284, "00000000001p")
	put(316, "0000000042")
	put(346, "000000003y")
	put(356, "ABC")
	put(359, "DEF")
	put(362, "TR")
	r, err := ParseRetB(b)
	if err != nil {
		t.Fatal(err)
	}
	if v, ok := r.Dan(); !ok || v != 2024 {
		t.Errorf("dan %d", v)
	}
	if r.Cc2dep() != "75" || r.Ccodir() != "1" || r.Ac3dir() != "751" || r.Ccocom() != "056" || r.Dsrpar() != "A" {
		t.Errorf("ids")
	}
	if r.Aidfic() != "2024751056A" || r.Anupro() != "B01234" || r.Cgroup() != "B" || r.Ccobnb() != "X" {
		t.Errorf("groups")
	}
	if v, ok := r.Nnupro(); !ok || v != 1234 {
		t.Errorf("nnupro")
	}
	signed := []struct {
		name string
		f    func() (int64, bool)
		want int64
	}{
		{"mctcom", r.Mctcom, -121}, {"mctdep", r.Mctdep, -100}, {"mctreg", r.Mctreg, 3},
		{"mfa300", r.Mfa300, 17}, {"mcfano", r.Mcfano, 17}, {"mfn300", r.Mfn300, 18}, {"mcfnvn", r.Mcfnvn, 18},
		{"mfa800", r.Mfa800, 19}, {"momano", r.Momano, 19}, {"mfn800", r.Mfn800, 10}, {"momnvn", r.Momnvn, 10},
		{"tcthfr", r.Tcthfr, 1233}, {"tctfra", r.Tctfra, 0}, {"tctdu", r.Tctdu, -10}, {"mcota3", r.Mcota3, -39},
	}
	for _, c := range signed {
		if v, ok := c.f(); !ok || v != c.want {
			t.Errorf("%s = %d,%v want %d", c.name, v, ok, c.want)
		}
	}
	if v, _, _ := r.Mcbt13(0); v != 11 {
		t.Errorf("bt13[0] %d", v)
	}
	if v, _, _ := r.Mcbt13(1); v != -22 {
		t.Errorf("bt13[1] %d", v)
	}
	if v, ok := r.Mcbtsa(); !ok || v != 99 {
		t.Errorf("mcbtsa")
	}
	if g, _ := r.Gtauom(5); g != "RA" {
		t.Errorf("gtauom")
	}
	if v, _, _ := r.Mctom(5); v != 55 {
		t.Errorf("mctom %d", v)
	}
	if v, _ := r.Tctom(); v != 42 {
		t.Errorf("tctom")
	}
	if !strings.HasPrefix(r.Filler(), "FILL") || !strings.HasPrefix(r.Trailer(), "TR") || r.Ccoifp() != "ABC" || r.Ccpper() != "DEF" {
		t.Errorf("tail")
	}
	// non-numeric data errors
	if _, ok := r.Mctsyn(); ok {
		t.Errorf("spaces numeric?")
	}
	if _, ok := r.Mfa900(); ok {
		t.Errorf("mfa900 numeric?")
	}
	r.SetAidfic("20X4")
	if _, ok := r.Dan(); ok {
		t.Errorf("dan should be non-numeric")
	}
}

func TestRetBSerializeRoundTrip(t *testing.T) {
	type sf struct {
		name string
		set  func(*RetB, int64)
		get  func(*RetB) (int64, bool)
		off  int
		n    int
	}
	fields := []sf{
		{"mctcom", (*RetB).SetMctcom, (*RetB).Mctcom, 18, 10},
		{"mctdep", (*RetB).SetMctdep, (*RetB).Mctdep, 28, 10},
		{"mctreg", (*RetB).SetMctreg, (*RetB).Mctreg, 38, 10},
		{"mctsyn", (*RetB).SetMctsyn, (*RetB).Mctsyn, 48, 10},
		{"mctcu", (*RetB).SetMctcu, (*RetB).Mctcu, 58, 10},
		{"mcttse", (*RetB).SetMcttse, (*RetB).Mcttse, 68, 10},
		{"mfa300", (*RetB).SetMfa300, (*RetB).Mfa300, 220, 10},
		{"mfn300", (*RetB).SetMfn300, (*RetB).Mfn300, 230, 10},
		{"mfa800", (*RetB).SetMfa800, (*RetB).Mfa800, 240, 10},
		{"mfn800", (*RetB).SetMfn800, (*RetB).Mfn800, 250, 10},
		{"tcthfr", (*RetB).SetTcthfr, (*RetB).Tcthfr, 260, 12},
		{"tctfra", (*RetB).SetTctfra, (*RetB).Tctfra, 272, 12},
		{"tctdu", (*RetB).SetTctdu, (*RetB).Tctdu, 284, 12},
		{"mfa900", (*RetB).SetMfa900, (*RetB).Mfa900, 296, 10},
		{"mfn900", (*RetB).SetMfn900, (*RetB).Mfn900, 306, 10},
		{"mvltim", (*RetB).SetMvltim, (*RetB).Mvltim, 326, 10},
		{"mcoge3", (*RetB).SetMcoge3, (*RetB).Mcoge3, 336, 10},
		{"mcota3", (*RetB).SetMcota3, (*RetB).Mcota3, 346, 10},
	}
	for _, f := range fields {
		for _, v := range []int64{0, 1, -121, 9999999999, -9999999999} {
			r := NewRetB()
			f.set(&r, v)
			got, ok := f.get(&r)
			if !ok || got != v {
				t.Errorf("%s(%d) = %d,%v", f.name, v, got, ok)
			}
			r2, err := ParseRetB(r.Bytes())
			if err != nil {
				t.Fatal(err)
			}
			if g2, _ := f.get(&r2); g2 != v {
				t.Errorf("%s reparse", f.name)
			}
		}
		r := NewRetB()
		f.set(&r, -121)
		if s := string(r[f.off : f.off+f.n]); s != strings.Repeat("0", f.n-3)+"12J" {
			t.Errorf("%s encoding %q", f.name, s)
		}
	}
}

func TestRetBTruncationAndUnsigned(t *testing.T) {
	r := NewRetB()
	r.SetMctcom(12345678901)
	if v, _ := r.Mctcom(); v != 2345678901 {
		t.Errorf("trunc %d", v)
	}
	r.SetMctcom(-10000000000)
	if r.MctcomRaw() != "000000000{" {
		t.Errorf("neg zero %q", r.MctcomRaw())
	}
	r.SetTctdu(-1234567890123)
	if v, _ := r.Tctdu(); v != -234567890123 {
		t.Errorf("tctdu trunc %d", v)
	}
	r.SetMcbtsa(-55)
	if v, ok := r.Mcbtsa(); !ok || v != 55 {
		t.Errorf("mcbtsa sign")
	}
	r.SetTctom(-12345678901)
	if v, _ := r.Tctom(); v != 2345678901 {
		t.Errorf("tctom %d", v)
	}
	r.SetDan(12025)
	if r.DanRaw() != "2025" {
		t.Errorf("dan %q", r.DanRaw())
	}
	r.SetNnupro(1234567)
	if r.NnuproRaw() != "34567" {
		t.Errorf("nnupro %q", r.NnuproRaw())
	}
	r.SetCc2dep("2AB")
	r.SetCcodir("")
	if r.Ac3dir() != "2A " {
		t.Errorf("ac3dir %q", r.Ac3dir())
	}
	r.SetCcocom("1")
	r.SetDsrpar("Z")
	r.SetCgroup("P")
	r.SetCcobnb("B")
	r.SetCcoifp("12345")
	r.SetCcpper("X")
	r.SetFiller("f")
	r.SetTrailer("t")
	r.SetAnupro("Q")
	if r.Ccocom() != "1  " || r.Dsrpar() != "Z" || r.Anupro() != "Q     " || r.Ccobnb() != "B" ||
		r.Ccoifp() != "123" || r.Ccpper() != "X  " || r.Filler() != "f"+strings.Repeat(" ", 39) ||
		r.Trailer() != "t"+strings.Repeat(" ", 237) {
		t.Errorf("alpha moves")
	}
	if len(r.Bytes()) != RetBLen {
		t.Errorf("len")
	}
}

func TestRetBRedefines(t *testing.T) {
	cases := []struct {
		name  string
		setA  func(*RetB, int64)
		getB  func(*RetB) (int64, bool)
		setB  func(*RetB, int64)
		getA  func(*RetB) (int64, bool)
	}{
		{"mfa300/mcfano", (*RetB).SetMfa300, (*RetB).Mcfano, (*RetB).SetMcfano, (*RetB).Mfa300},
		{"mfn300/mcfnvn", (*RetB).SetMfn300, (*RetB).Mcfnvn, (*RetB).SetMcfnvn, (*RetB).Mfn300},
		{"mfa800/momano", (*RetB).SetMfa800, (*RetB).Momano, (*RetB).SetMomano, (*RetB).Mfa800},
		{"mfn800/momnvn", (*RetB).SetMfn800, (*RetB).Momnvn, (*RetB).SetMomnvn, (*RetB).Mfn800},
	}
	for _, c := range cases {
		r := NewRetB()
		c.setA(&r, 42)
		if v, _ := c.getB(&r); v != 42 {
			t.Errorf("%s a->b", c.name)
		}
		c.setB(&r, -7)
		if v, _ := c.getA(&r); v != -7 {
			t.Errorf("%s b->a", c.name)
		}
	}
}

func TestRetBOccurs(t *testing.T) {
	r := NewRetB()
	for i := 0; i < RetBOmCount; i++ {
		if err := r.SetGtauom(i, "R"); err != nil {
			t.Fatal(err)
		}
		if err := r.SetMctom(i, int64(-i-1)); err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < RetBOmCount; i++ {
		g, _ := r.Gtauom(i)
		v, ok, _ := r.Mctom(i)
		raw, _ := r.MctomRaw(i)
		if g != "R " || !ok || v != int64(-i-1) || raw != string(r[148+12*i+2:148+12*i+12]) {
			t.Errorf("om %d", i)
		}
	}
	for i := 0; i < RetBBt13Count; i++ {
		if err := r.SetMcbt13(i, int64(100+i)); err != nil {
			t.Fatal(err)
		}
		raw, _ := r.Mcbt13Raw(i)
		if raw != string(r[78+10*i:88+10*i]) {
			t.Errorf("bt13 raw %d", i)
		}
	}
	bad := []int{-1, 6}
	for _, i := range bad {
		if _, err := r.Gtauom(i); err == nil {
			t.Error("gtauom")
		}
		if err := r.SetGtauom(i, "x"); err == nil {
			t.Error("setgtauom")
		}
		if _, _, err := r.Mctom(i); err == nil {
			t.Error("mctom")
		}
		if err := r.SetMctom(i, 1); err == nil {
			t.Error("setmctom")
		}
		if _, err := r.MctomRaw(i); err == nil {
			t.Error("mctomraw")
		}
	}
	for _, i := range []int{-1, 2} {
		if _, _, err := r.Mcbt13(i); err == nil {
			t.Error("mcbt13")
		}
		if err := r.SetMcbt13(i, 1); err == nil {
			t.Error("setmcbt13")
		}
		if _, err := r.Mcbt13Raw(i); err == nil {
			t.Error("mcbt13raw")
		}
	}
}

func TestDecodeRetBAndBytesCopy(t *testing.T) {
	var b [RetBLen]byte
	b[0] = '1'
	r := DecodeRetB(b)
	out := r.Bytes()
	out[0] = '9'
	if r[0] != '1' {
		t.Error("Bytes not a copy")
	}
}