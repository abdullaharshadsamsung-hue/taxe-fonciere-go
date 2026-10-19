package zes

import (
	"bytes"
	"strings"
	"testing"
)

func buildCOMRecord() []byte {
	b := make([]byte, COMLen)
	for i := range b {
		b[i] = byte('A' + i%26)
	}
	w := func(off int, s string) { copy(b[off:], s) }
	w(0, "2018")
	w(4, "75")
	w(6, "1")
	w(7, "056")
	w(10, "000")
	w(13, "CO")
	w(15, "V")
	w(16, "180101")
	w(22, "123")
	w(25, "045")
	w(28, "150618")
	w(34, "L")
	w(35, "G")
	w(36, "07")
	w(38, "SERIES1")
	w(45, "20180101")
	w(53, "20181231")
	w(61, "GEM1")
	w(65, "0012345678")
	w(75, "0000000001")
	w(85, "9999999999")
	w(95, "0001000000")
	w(105, "0000500000")
	w(115, "5")
	for i := 0; i < 5; i++ {
		w(116+i*8, "2010201"+string(rune('0'+i)))
		w(156+i*8, "2000200"+string(rune('0'+i)))
	}
	w(196, "PARIS                         ")
	w(226, "N")
	w(227, "E")
	w(228, "C")
	w(229, "COL")
	w(232, "S123")
	w(236, "GOM1")
	w(240, "001500")
	w(246, "150")
	w(249, "D")
	w(250, "Z01Z02Z03Z04Z05")
	w(265, "2200000")
	w(272, "0007654321")
	w(304, "000012345")
	w(313, "000067890")
	w(322, "T456")
	w(326, "000000000001234567")
	w(344, "BIM1")
	w(348, "O")
	w(349, "12345678901234")
	w(363, "F")
	w(398, "UR")
	return b
}

func TestCOMLayoutConstants(t *testing.T) {
	if COMLen != 400 || COMAidficLen+COMAdtsegLen != COMLen {
		t.Fatalf("bad lengths")
	}
	if COMRateScale != 1_000_000 || COMPplomScale != 100 || COMJaCount != 5 {
		t.Fatalf("bad constants")
	}
}

func TestParseCOMLength(t *testing.T) {
	tests := []struct {
		n       int
		wantErr bool
	}{{0, true}, {399, true}, {400, false}, {401, true}}
	for _, tc := range tests {
		_, err := ParseCOM(make([]byte, tc.n))
		if (err != nil) != tc.wantErr {
			t.Errorf("len %d: err=%v", tc.n, err)
		}
	}
}

func TestParseCOMDoesNotMutateAndCopies(t *testing.T) {
	rec := buildCOMRecord()
	orig := append([]byte(nil), rec...)
	c, err := ParseCOM(rec)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(rec, orig) {
		t.Fatal("input mutated")
	}
	rec[0] = 'X'
	if c[0] != '2' {
		t.Fatal("parse aliases input")
	}
	out := c.Bytes()
	if !bytes.Equal(out, orig) {
		t.Fatal("round trip mismatch")
	}
	out[0] = 'Q'
	if c[0] != '2' {
		t.Fatal("Bytes aliases storage")
	}
}

func TestCOMStringFields(t *testing.T) {
	rec := buildCOMRecord()
	c, _ := ParseCOM(rec)
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"Aidfic", c.Aidfic(), "2018751056000"},
		{"JanRaw", c.JanRaw(), "2018"},
		{"Depdir", c.Depdir(), "751"},
		{"Codep", c.Codep(), "75"},
		{"Codir", c.Codir(), "1"},
		{"Ccocom", c.Ccocom(), "056"},
		{"Ccoifp", c.Ccoifp(), "000"},
		{"Adtseg", c.Adtseg(), string(rec[13:400])},
		{"Ccoenr", c.Ccoenr(), "CO"},
		{"Gtovot", c.Gtovot(), "V"},
		{"Xdday6", c.Xdday6(), "180101"},
		{"Jdhomo", c.Jdhomo(), "150618"},
		{"Dnultf", c.Dnultf(), "L"},
		{"Gcoanu", c.Gcoanu(), "G"},
		{"Dlisro", c.Dlisro(), "SERIES1"},
		{"Acogem", c.Acogem(), "GEM1"},
		{"Dlicom", c.Dlicom(), "PARIS                         "},
		{"Gnidep", c.Gnidep(), "N"},
		{"Gtoeco", c.Gtoeco(), "E"},
		{"Ctygc", c.Ctygc(), "C"},
		{"Dnucol", c.Dnucol(), "COL"},
		{"Acosyn", c.Acosyn(), "S123"},
		{"Ctysyn", c.Ctysyn(), "S"},
		{"Ccosyn", c.Ccosyn(), "123"},
		{"Acogom", c.Acogom(), "GOM1"},
		{"Dcmzo1", c.Dcmzo1(), "D"},
		{"Anuzip", c.Anuzip(), "Z01Z02Z03Z04Z05"},
		{"Dnuzip", c.Dnuzip(), "Z01"},
		{"Dnuzia", c.Dnuzia(), "Z02"},
		{"Dnuzib", c.Dnuzib(), "Z03"},
		{"Dnuzic", c.Dnuzic(), "Z04"},
		{"Dnuzid", c.Dnuzid(), "Z05"},
		{"Filler1", c.Filler1(), string(rec[282:304])},
		{"Acotxa", c.Acotxa(), "T456"},
		{"Ctytxa", c.Ctytxa(), "T"},
		{"Ccotxa", c.Ccotxa(), "456"},
		{"Acobim", c.Acobim(), "BIM1"},
		{"Gtomgp", c.Gtomgp(), "O"},
		{"Dsirec", c.Dsirec(), "12345678901234"},
		{"Gtocfu", c.Gtocfu(), "F"},
		{"Filler2", c.Filler2(), string(rec[364:398])},
		{"Dnuurc", c.Dnuurc(), "UR"},
	}
	for _, tc := range tests {
		if tc.got != tc.want {
			t.Errorf("%s = %q, want %q", tc.name, tc.got, tc.want)
		}
	}
}

func TestCOMNumericFields(t *testing.T) {
	rec := buildCOMRecord()
	c, _ := ParseCOM(rec)
	tests := []struct {
		name string
		fn   func() (uint64, bool)
		want uint64
	}{
		{"Jan", c.Jan, 2018},
		{"Pmjpo", c.Pmjpo, 123},
		{"Dnuro", c.Dnuro, 45},
		{"Nbsrol", c.Nbsrol, 7},
		{"Jdatau", c.Jdatau, 20180101},
		{"Jdacmp", c.Jdacmp, 20181231},
		{"Pbbomp", c.Pbbomp, 12345678},
		{"Pbboma", c.Pbboma, 1},
		{"Pbbomb", c.Pbbomb, 9999999999},
		{"Pbbomc", c.Pbbomc, 1 * COMRateScale},
		{"Pbbomd", c.Pbbomd, 500000},
		{"Nbtauxom", c.Nbtauxom, 5},
		{"Dvlpom", c.Dvlpom, 1500},
		{"Pplom", c.Pplom, 150},
		{"Npocom", c.Npocom, 2200000},
		{"Pbbome", c.Pbbome, 7654321},
		{"Dsedom", c.Dsedom, 12345},
		{"Dsednb", c.Dsednb, 67890},
		{"Pbbomi", c.Pbbomi, 1234567},
	}
	for _, tc := range tests {
		v, ok := tc.fn()
		if !ok || v != tc.want {
			t.Errorf("%s = %d,%v want %d", tc.name, v, ok, tc.want)
		}
	}
	if !c.JanIsNumeric() {
		t.Error("JanIsNumeric false")
	}
	// 12.345678 example
	v, _ := c.Pbbomp()
	if v/COMRateScale != 12 || v%COMRateScale != 345678 {
		t.Error("rate scaling")
	}
}

func TestCOMNonNumeric(t *testing.T) {
	tests := []struct {
		raw    string
		want   uint64
		wantOk bool
	}{
		{"2018", 2018, true},
		{"20 8", 2008, false},
		{"    ", 0, false},
		{"20A8", 2018, false},
		{"0000", 0, true},
	}
	for _, tc := range tests {
		c := NewCOM()
		copy(c[0:4], tc.raw)
		v, ok := c.Jan()
		if ok != tc.wantOk || v != tc.want || c.JanIsNumeric() != tc.wantOk {
			t.Errorf("%q: got %d,%v", tc.raw, v, ok)
		}
	}
	c := NewCOM()
	c[0] = 0xFF
	copy(c[1:4], "000")
	if v, ok := c.Jan(); ok || v != 9000 {
		t.Errorf("clamp: %d %v", v, ok)
	}
}

func TestCOMJaTables(t *testing.T) {
	c, _ := ParseCOM(buildCOMRecord())
	for i := 0; i < COMJaCount; i++ {
		d := uint64(i)
		if s, err := c.Jcoja(i); err != nil || s != "2010201"+string(rune('0'+i)) {
			t.Errorf("Jcoja(%d)=%q %v", i, s, err)
		}
		if v, ok, err := c.Jcoins(i); err != nil || !ok || v != 2010 {
			t.Errorf("Jcoins(%d)=%d", i, v)
		}
		if v, ok, err := c.Jcorim(i); err != nil || !ok || v != 2010+d {
			t.Errorf("Jcorim(%d)=%d", i, v)
		}
		if s, err := c.Jgcja(i); err != nil || s != "2000200"+string(rune('0'+i)) {
			t.Errorf("Jgcja(%d)=%q", i, s)
		}
		if v, ok, err := c.Jgcins(i); err != nil || !ok || v != 2000 {
			t.Errorf("Jgcins(%d)=%d", i, v)
		}
		if v, ok, err := c.Jgcrim(i); err != nil || !ok || v != 2000+d {
			t.Errorf("Jgcrim(%d)=%d", i, v)
		}
	}
}

func TestCOMJaIndexErrors(t *testing.T) {
	c := NewCOM()
	before := c
	for _, i := range []int{-1, 5, 100} {
		checks := []error{}
		_, e := c.Jcoja(i)
		checks = append(checks, e, c.SetJcoja(i, "x"))
		_, _, e = c.Jcoins(i)
		checks = append(checks, e, c.SetJcoins(i, 1))
		_, _, e = c.Jcorim(i)
		checks = append(checks, e, c.SetJcorim(i, 1))
		_, e = c.Jgcja(i)
		checks = append(checks, e, c.SetJgcja(i, "x"))
		_, _, e = c.Jgcins(i)
		checks = append(checks, e, c.SetJgcins(i, 1))
		_, _, e = c.Jgcrim(i)
		checks = append(checks, e, c.SetJgcrim(i, 1))
		for k, err := range checks {
			if err == nil {
				t.Errorf("index %d check %d: expected error", i, k)
			}
		}
	}
	if c != before {
		t.Error("out-of-range setter mutated storage")
	}
}

func TestCOMJaSetters(t *testing.T) {
	c := NewCOM()
	if err := c.SetJcoins(2, 2015); err != nil {
		t.Fatal(err)
	}
	if err := c.SetJcorim(2, 2020); err != nil {
		t.Fatal(err)
	}
	if err := c.SetJgcins(4, 12345); err != nil {
		t.Fatal(err)
	}
	if err := c.SetJgcrim(0, 7); err != nil {
		t.Fatal(err)
	}
	if got := string(c[132:140]); got != "20152020" {
		t.Errorf("JCOJA(3)=%q", got)
	}
	if got := string(c[188:192]); got != "2345" {
		t.Errorf("JGCINS(5)=%q", got)
	}
	if got := string(c[160:164]); got != "0007" {
		t.Errorf("JGCRIM(1)=%q", got)
	}
	_ = c.SetJcoja(0, "ABC")
	if got := string(c[116:124]); got != "ABC     " {
		t.Errorf("JCOJA(1)=%q", got)
	}
	_ = c.SetJgcja(1, "123456789")
	if got := string(c[164:172]); got != "12345678" || c[172] != ' ' {
		t.Errorf("JGCJA(2)=%q", got)
	}
}

func TestCOMNewIsSpaces(t *testing.T) {
	c := NewCOM()
	if string(c[:]) != strings.Repeat(" ", COMLen) {
		t.Fatal("not spaces")
	}
	if c.Filler1() != strings.Repeat(" ", 22) || c.Filler2() != strings.Repeat(" ", 34) {
		t.Fatal("filler not spaces")
	}
	if _, ok := c.Pbbomp(); ok {
		t.Fatal("spaces numeric")
	}
}

func TestCOMAlphaSetters(t *testing.T) {
	tests := []struct {
		name string
		off  int
		n    int
		set  func(*COM, string)
	}{
		{"Aidfic", 0, 13, (*COM).SetAidfic},
		{"Depdir", 4, 3, (*COM).SetDepdir},
		{"Codep", 4, 2, (*COM).SetCodep},
		{"Codir", 6, 1, (*COM).SetCodir},
		{"Ccocom", 7, 3, (*COM).SetCcocom},
		{"Ccoifp", 10, 3, (*COM).SetCcoifp},
		{"Adtseg", 13, 387, (*COM).SetAdtseg},
		{"Ccoenr", 13, 2, (*COM).SetCcoenr},
		{"Gtovot", 15, 1, (*COM).SetGtovot},
		{"Xdday6", 16, 6, (*COM).SetXdday6},
		{"Jdhomo", 28, 6, (*COM).SetJdhomo},
		{"Dnultf", 34, 1, (*COM).SetDnultf},
		{"Gcoanu", 35, 1, (*COM).SetGcoanu},
		{"Dlisro", 38, 7, (*COM).SetDlisro},
		{"Acogem", 61, 4, (*COM).SetAcogem},
		{"Dlicom", 196, 30, (*COM).SetDlicom},
		{"Gnidep", 226, 1, (*COM).SetGnidep},
		{"Gtoeco", 227, 1, (*COM).SetGtoeco},
		{"Ctygc", 228, 1, (*COM).SetCtygc},
		{"Dnucol", 229, 3, (*COM).SetDnucol},
		{"Acosyn", 232, 4, (*COM).SetAcosyn},
		{"Ctysyn", 232, 1, (*COM).SetCtysyn},
		{"Ccosyn", 233, 3, (*COM).SetCcosyn},
		{"Acogom", 236, 4, (*COM).SetAcogom},
		{"Dcmzo1", 249, 1, (*COM).SetDcmzo1},
		{"Anuzip", 250, 15, (*COM).SetAnuzip},
		{"Dnuzip", 250, 3, (*COM).SetDnuzip},
		{"Dnuzia", 253, 3, (*COM).SetDnuzia},
		{"Dnuzib", 256, 3, (*COM).SetDnuzib},
		{"Dnuzic", 259, 3, (*COM).SetDnuzic},
		{"Dnuzid", 262, 3, (*COM).SetDnuzid},
		{"Filler1", 282, 22, (*COM).SetFiller1},
		{"Acotxa", 322, 4, (*COM).SetAcotxa},
		{"Ctytxa", 322, 1, (*COM).SetCtytxa},
		{"Ccotxa", 323, 3, (*COM).SetCcotxa},
		{"Acobim", 344, 4, (*COM).SetAcobim},
		{"Gtomgp", 348, 1, (*COM).SetGtomgp},
		{"Dsirec", 349, 14, (*COM).SetDsirec},
		{"Gtocfu", 363, 1, (*COM).SetGtocfu},
		{"Filler2", 364, 34, (*COM).SetFiller2},
		{"Dnuurc", 398, 2, (*COM).SetDnuurc},
	}
	for _, tc := range tests {
		// long value: truncated on the right
		c := NewCOM()
		for i := range c {
			c[i] = '#'
		}
		long := strings.Repeat("9", tc.n+5)
		tc.set(&c, long)
		for i := 0; i < COMLen; i++ {
			want := byte('#')
			if i >= tc.off && i < tc.off+tc.n {
				want = '9'
			}
			if c[i] != want {
				t.Errorf("%s long: byte %d = %q", tc.name, i, c[i])
				break
			}
		}
		// short value: padded with spaces
		tc.set(&c, "x")
		if c[tc.off] != 'x' {
			t.Errorf("%s short: first byte %q", tc.name, c[tc.off])
		}
		for i := tc.off + 1; i < tc.off+tc.n; i++ {
			if c[i] != ' ' {
				t.Errorf("%s short: byte %d not space", tc.name, i)
				break
			}
		}
	}
}

func TestCOMNumericSetters(t *testing.T) {
	tests := []struct {
		name string
		off  int
		n    int
		set  func(*COM, uint64)
		get  func(*COM) (uint64, bool)
	}{
		{"Jan", 0, 4, (*COM).SetJan, (*COM).Jan},
		{"Pmjpo", 22, 3, (*COM).SetPmjpo, (*COM).Pmjpo},
		{"Dnuro", 25, 3, (*COM).SetDnuro, (*COM).Dnuro},
		{"Nbsrol", 36, 2, (*COM).SetNbsrol, (*COM).Nbsrol},
		{"Jdatau", 45, 8, (*COM).SetJdatau, (*COM).Jdatau},
		{"Jdacmp", 53, 8, (*COM).SetJdacmp, (*COM).Jdacmp},
		{"Pbbomp", 65, 10, (*COM).SetPbbomp, (*COM).Pbbomp},
		{"Pbboma", 75, 10, (*COM).SetPbboma, (*COM).Pbboma},
		{"Pbbomb", 85, 10, (*COM).SetPbbomb, (*COM).Pbbomb},
		{"Pbbomc", 95, 10, (*COM).SetPbbomc, (*COM).Pbbomc},
		{"Pbbomd", 105, 10, (*COM).SetPbbomd, (*COM).Pbbomd},
		{"Nbtauxom", 115, 1, (*COM).SetNbtauxom, (*COM).Nbtauxom},
		{"Dvlpom", 240, 6, (*COM).SetDvlpom, (*COM).Dvlpom},
		{"Pplom", 246, 3, (*COM).SetPplom, (*COM).Pplom},
		{"Npocom", 265, 7, (*COM).SetNpocom, (*COM).Npocom},
		{"Pbbome", 272, 10, (*COM).SetPbbome, (*COM).Pbbome},
		{"Dsedom", 304, 9, (*COM).SetDsedom, (*COM).Dsedom},
		{"Dsednb", 313, 9, (*COM).SetDsednb, (*COM).Dsednb},
		{"Pbbomi", 326, 18, (*COM).SetPbbomi, (*COM).Pbbomi},
	}
	for _, tc := range tests {
		c := NewCOM()
		tc.set(&c, 7)
		want := strings.Repeat("0", tc.n-1) + "7"
		if got := string(c[tc.off : tc.off+tc.n]); got != want {
			t.Errorf("%s zero-pad: %q", tc.name, got)
		}
		if tc.off > 0 && c[tc.off-1] != ' ' || tc.off+tc.n < COMLen && c[tc.off+tc.n] != ' ' {
			t.Errorf("%s touched neighbours", tc.name)
		}
		// truncation of high-order digit
		var mod uint64 = 1
		for i := 0; i < tc.n; i++ {
			mod *= 10
		}
		tc.set(&c, mod+3)
		if v, ok := tc.get(&c); !ok || v != 3 {
			t.Errorf("%s truncate: %d %v", tc.name, v, ok)
		}
		tc.set(&c, mod-1)
		if v, ok := tc.get(&c); !ok || v != mod-1 {
			t.Errorf("%s max: %d", tc.name, v)
		}
	}
	c := NewCOM()
	c.SetPbbomp(12345678)
	if string(c[65:75]) != "0012345678" {
		t.Error("rate encoding")
	}
	c.SetPplom(150)
	if string(c[246:249]) != "150" {
		t.Error("pplom encoding")
	}
}

func TestCOMGroupChildSharing(t *testing.T) {
	c := NewCOM()
	c.SetCodep("13")
	c.SetCodir("2")
	c.SetCtysyn("S")
	c.SetCcosyn("999")
	c.SetCtytxa("T")
	c.SetCcotxa("888")
	c.SetDnuzip("A")
	c.SetDnuzid("EEE")
	if c.Depdir() != "132" || c.Acosyn() != "S999" || c.Acotxa() != "T888" {
		t.Error("group views")
	}
	if c.Anuzip() != "A           EEE" {
		t.Errorf("Anuzip=%q", c.Anuzip())
	}
	c.SetAidfic("2019AB1XYZQRS")
	if v, _ := c.Jan(); v != 2019 || c.Codep() != "AB" || c.Codir() != "1" || c.Ccocom() != "XYZ" || c.Ccoifp() != "QRS" {
		t.Error("aidfic children")
	}
	if c.Aidfic()+c.Adtseg() != string(c[:]) {
		t.Error("aidfic+adtseg != record")
	}
}

func TestCOMSerializeRoundTrip(t *testing.T) {
	rec := buildCOMRecord()
	c, _ := ParseCOM(rec)
	c.SetDlicom(c.Dlicom())
	c.SetPbbomi(1234567)
	c.SetDsirec(c.Dsirec())
	if !bytes.Equal(c.Bytes(), rec) {
		t.Fatal("serialize(parse(r)) != r")
	}
	if len(c.Bytes()) != COMLen {
		t.Fatal("length")
	}
}