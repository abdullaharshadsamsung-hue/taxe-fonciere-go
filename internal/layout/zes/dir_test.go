package zes

import (
	"strings"
	"testing"
)

// buildDIRRecord builds a 400-byte record with known content at 1-based positions.
func buildDIRRecord(t *testing.T) []byte {
	t.Helper()
	b := []byte(strings.Repeat(" ", DIRLen))
	set := func(pos1 int, s string) { copy(b[pos1-1:], s) }
	set(1, "2018")
	set(5, "75")
	set(7, "1")
	set(8, "056")
	set(11, "123")
	set(14, "DI")
	set(16, "180315")
	set(22, "0012345678")
	set(32, "0001500000")
	set(42, "0000000001")
	set(52, "9999999999")
	set(62, "0000100000")
	set(72, "0002000000")
	set(82, "0405")
	set(86, "0003000000")
	for i := 0; i < 5; i++ {
		set(96+i*8, []string{"20102015", "20112016", "20122017", "20132018", "20142019"}[i])
	}
	set(136, "FILLER-01")
	for i := 0; i < 5; i++ {
		set(145+i*4, []string{"AAAA", "BBBB", "CCCC", "DDDD", "EEEE"}[i])
	}
	set(165, "03")
	set(167, "0004000000")
	set(177, "12345678901234")
	set(191, "O")
	set(192, "22222222222222")
	set(206, "33333333333333")
	set(220, strings.Repeat("Z", 179))
	set(399, "U7")
	return b
}

func TestDIRLayoutConstants(t *testing.T) {
	if DIRLen != 400 || DIRAidficLen != 13 || DIRAdtsegLen != 387 {
		t.Fatalf("bad lengths")
	}
	if DIRAidficLen+DIRAdtsegLen != DIRLen {
		t.Fatalf("AIDFIC+ADTSEG != DIRLen")
	}
	if DIRJaCount != 5 || DIRTseCount != 5 || DIRFiller1Len != 9 || DIRFiller2Len != 179 {
		t.Fatalf("bad counts")
	}
	if DIRRateScale != 1000000 || DIRBapsaScale != 100 {
		t.Fatalf("bad scales")
	}
}

func TestParseDIRLength(t *testing.T) {
	tests := []struct {
		name    string
		n       int
		wantErr bool
	}{
		{"empty", 0, true},
		{"short", 399, true},
		{"exact", 400, false},
		{"long", 401, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseDIR(make([]byte, tc.n))
			if (err != nil) != tc.wantErr {
				t.Fatalf("err=%v wantErr=%v", err, tc.wantErr)
			}
		})
	}
}

func TestNewDIRSpaces(t *testing.T) {
	d := NewDIR()
	b := d.Bytes()
	if len(b) != DIRLen || string(b) != strings.Repeat(" ", DIRLen) {
		t.Fatalf("NewDIR not all spaces")
	}
	if _, ok := d.Jan(); ok {
		t.Fatalf("spaces should not be numeric")
	}
}

func TestDIRParseStringFields(t *testing.T) {
	d, err := ParseDIR(buildDIRRecord(t))
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"Aidfic", d.Aidfic(), "2018751056123"},
		{"JanRaw", d.JanRaw(), "2018"},
		{"Depdir", d.Depdir(), "751"},
		{"Codep", d.Codep(), "75"},
		{"Codir", d.Codir(), "1"},
		{"Ccocom", d.Ccocom(), "056"},
		{"Ccoifp", d.Ccoifp(), "123"},
		{"Ccoenr", d.Ccoenr(), "DI"},
		{"Xdday6", d.Xdday6(), "180315"},
		{"Filler1", d.Filler1(), "FILLER-01"},
		{"Dsitas", d.Dsitas(), "12345678901234"},
		{"Grotas", d.Grotas(), "O"},
		{"Dsired", d.Dsired(), "22222222222222"},
		{"Dsica", d.Dsica(), "33333333333333"},
		{"Filler2", d.Filler2(), strings.Repeat("Z", 179)},
		{"Dnuurc", d.Dnuurc(), "U7"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.want {
				t.Fatalf("got %q want %q", tc.got, tc.want)
			}
		})
	}
	if len(d.Adtseg()) != 387 || !strings.HasPrefix(d.Adtseg(), "DI180315") || !strings.HasSuffix(d.Adtseg(), "U7") {
		t.Fatalf("bad Adtseg %q", d.Adtseg())
	}
}

func TestDIRParseNumericFields(t *testing.T) {
	d, _ := ParseDIR(buildDIRRecord(t))
	tests := []struct {
		name string
		fn   func() (uint64, bool)
		want uint64
	}{
		{"Jan", d.Jan, 2018},
		{"TaudepB", d.TaudepB, 12345678},
		{"TauregB", d.TauregB, 1500000},
		{"Ptbtas", d.Ptbtas, 1},
		{"TaudepNB", d.TaudepNB, 9999999999},
		{"TauregNB", d.TauregNB, 100000},
		{"TauchagrNB", d.TauchagrNB, 2000000},
		{"Taubap405NB", d.Taubap405NB, 405},
		{"TausmagrNB", d.TausmagrNB, 3000000},
		{"Nbtse", d.Nbtse, 3},
		{"Ptbdrb", d.Ptbdrb, 4000000},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			v, ok := tc.fn()
			if !ok || v != tc.want {
				t.Fatalf("got %d,%v want %d", v, ok, tc.want)
			}
		})
	}
	if v, _ := d.TaudepB(); v/DIRRateScale != 12 || v%DIRRateScale != 345678 {
		t.Fatalf("scaling wrong: %d", v)
	}
	if !d.JanIsNumeric() {
		t.Fatal("JAN should be numeric")
	}
}

func TestDIRNonNumeric(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want uint64
	}{
		{"spaces", "    ", 0},
		{"letter", "20A8", 2018},
		{"high nibble clamp", "20?8", 2098},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			d := NewDIR()
			d.SetAidfic(tc.raw)
			if d.JanIsNumeric() {
				t.Fatal("expected non-numeric")
			}
			v, ok := d.Jan()
			if ok || v != tc.want {
				t.Fatalf("got %d,%v want %d", v, ok, tc.want)
			}
		})
	}
	d := NewDIR()
	d.SetAdtseg("        12X4567890")
	if _, ok := d.TaudepB(); ok {
		t.Fatal("rate should be non-numeric")
	}
}

func TestDIRTables(t *testing.T) {
	d, _ := ParseDIR(buildDIRRecord(t))
	for i := 0; i < 5; i++ {
		ja, err := d.Jdeja(i)
		if err != nil {
			t.Fatal(err)
		}
		ins, ok1, err1 := d.Jdeins(i)
		rim, ok2, err2 := d.Jderim(i)
		if err1 != nil || err2 != nil || !ok1 || !ok2 {
			t.Fatalf("index %d errors", i)
		}
		if ins != uint64(2010+i) || rim != uint64(2015+i) || len(ja) != 8 {
			t.Fatalf("index %d: %q %d %d", i, ja, ins, rim)
		}
		a, err := d.Acoets(i)
		if err != nil || a != strings.Repeat(string(rune('A'+i)), 4) {
			t.Fatalf("acoets %d = %q %v", i, a, err)
		}
	}
	for _, i := range []int{-1, 5, 100} {
		if _, err := d.Jdeja(i); err == nil {
			t.Errorf("Jdeja(%d) no error", i)
		}
		if err := d.SetJdeja(i, "x"); err == nil {
			t.Errorf("SetJdeja(%d) no error", i)
		}
		if _, _, err := d.Jdeins(i); err == nil {
			t.Errorf("Jdeins(%d)", i)
		}
		if err := d.SetJdeins(i, 1); err == nil {
			t.Errorf("SetJdeins(%d)", i)
		}
		if _, _, err := d.Jderim(i); err == nil {
			t.Errorf("Jderim(%d)", i)
		}
		if err := d.SetJderim(i, 1); err == nil {
			t.Errorf("SetJderim(%d)", i)
		}
		if _, err := d.Acoets(i); err == nil {
			t.Errorf("Acoets(%d)", i)
		}
		if err := d.SetAcoets(i, "x"); err == nil {
			t.Errorf("SetAcoets(%d)", i)
		}
	}
}

func TestDIRSetNumericEncoding(t *testing.T) {
	tests := []struct {
		name string
		set  func(d *DIR)
		pos  int
		want string
	}{
		{"rate 1.5", func(d *DIR) { d.SetTaudepB(1500000) }, 22, "0001500000"},
		{"rate truncate high", func(d *DIR) { d.SetTauregB(123456789012) }, 32, "3456789012"},
		{"ptbtas", func(d *DIR) { d.SetPtbtas(7) }, 42, "0000000007"},
		{"taudepNB", func(d *DIR) { d.SetTaudepNB(1) }, 52, "0000000001"},
		{"tauregNB", func(d *DIR) { d.SetTauregNB(2) }, 62, "0000000002"},
		{"tauchagr", func(d *DIR) { d.SetTauchagrNB(3) }, 72, "0000000003"},
		{"bapsa", func(d *DIR) { d.SetTaubap405NB(405) }, 82, "0405"},
		{"bapsa trunc", func(d *DIR) { d.SetTaubap405NB(12345) }, 82, "2345"},
		{"tausmagr", func(d *DIR) { d.SetTausmagrNB(4) }, 86, "0000000004"},
		{"nbtse", func(d *DIR) { d.SetNbtse(123) }, 165, "23"},
		{"ptbdrb", func(d *DIR) { d.SetPtbdrb(5) }, 167, "0000000005"},
		{"jan", func(d *DIR) { d.SetJan(12018) }, 1, "2018"},
		{"jdeins 2", func(d *DIR) { _ = d.SetJdeins(2, 2020) }, 112, "2020"},
		{"jderim 4", func(d *DIR) { _ = d.SetJderim(4, 2021) }, 132, "2021"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			d := NewDIR()
			tc.set(&d)
			b := d.Bytes()
			if got := string(b[tc.pos-1 : tc.pos-1+len(tc.want)]); got != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
			if strings.Count(string(b), " ") != DIRLen-len(tc.want) {
				t.Fatal("setter touched other bytes")
			}
		})
	}
}

func TestDIRSetAlphaEncoding(t *testing.T) {
	tests := []struct {
		name string
		set  func(d *DIR)
		pos  int
		want string
	}{
		{"codep pad", func(d *DIR) { d.SetCodep("7") }, 5, "7 "},
		{"codep trunc", func(d *DIR) { d.SetCodep("759") }, 5, "75"},
		{"codir", func(d *DIR) { d.SetCodir("XY") }, 7, "X"},
		{"depdir", func(d *DIR) { d.SetDepdir("13") }, 5, "13 "},
		{"ccocom", func(d *DIR) { d.SetCcocom("1") }, 8, "1  "},
		{"ccoifp", func(d *DIR) { d.SetCcoifp("abcd") }, 11, "abc"},
		{"ccoenr", func(d *DIR) { d.SetCcoenr("D") }, 14, "D "},
		{"xdday6", func(d *DIR) { d.SetXdday6("1803151") }, 16, "180315"},
		{"jdeja", func(d *DIR) { _ = d.SetJdeja(1, "2010") }, 104, "2010    "},
		{"filler1", func(d *DIR) { d.SetFiller1("AB") }, 136, "AB       "},
		{"acoets", func(d *DIR) { _ = d.SetAcoets(3, "XY") }, 157, "XY  "},
		{"dsitas", func(d *DIR) { d.SetDsitas("123") }, 177, "123           "},
		{"grotas", func(d *DIR) { d.SetGrotas("O") }, 191, "O"},
		{"dsired", func(d *DIR) { d.SetDsired("123456789012345") }, 192, "12345678901234"},
		{"dsica", func(d *DIR) { d.SetDsica("9") }, 206, "9             "},
		{"filler2", func(d *DIR) { d.SetFiller2("Q") }, 220, "Q" + strings.Repeat(" ", 178)},
		{"dnuurc", func(d *DIR) { d.SetDnuurc("U12") }, 399, "U1"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			d, _ := ParseDIR([]byte(strings.Repeat("#", DIRLen)))
			tc.set(&d)
			b := d.Bytes()
			if got := string(b[tc.pos-1 : tc.pos-1+len(tc.want)]); got != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
			if strings.Count(string(b), "#") != DIRLen-len(tc.want) {
				t.Fatal("setter touched other bytes")
			}
		})
	}
}

func TestDIRGroupSharing(t *testing.T) {
	d := NewDIR()
	d.SetAidfic("2019132045678")
	if j, ok := d.Jan(); !ok || j != 2019 || d.Codep() != "13" || d.Codir() != "2" || d.Ccocom() != "045" || d.Ccoifp() != "678" {
		t.Fatalf("aidfic group not shared")
	}
	d.SetCodir("9")
	if d.Depdir() != "139" {
		t.Fatalf("depdir=%q", d.Depdir())
	}
}

func TestDIRRoundTrip(t *testing.T) {
	src := buildDIRRecord(t)
	d, _ := ParseDIR(src)
	if string(d.Bytes()) != string(src) {
		t.Fatal("round trip mismatch")
	}
	b := d.Bytes()
	b[0] = 'X'
	if d.JanRaw() != "2018" {
		t.Fatal("Bytes not a copy")
	}
	src[0] = 'Y'
	if d.JanRaw() != "2018" {
		t.Fatal("ParseDIR not a copy")
	}
	e := NewDIR()
	e.SetTaudepB(12345678)
	_ = e.SetJdeins(0, 2010)
	_ = e.SetAcoets(0, "AAAA")
	f, err := ParseDIR(e.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	v, _ := f.TaudepB()
	y, _, _ := f.Jdeins(0)
	a, _ := f.Acoets(0)
	if v != 12345678 || y != 2010 || a != "AAAA" {
		t.Fatal("serialize/parse mismatch")
	}
}