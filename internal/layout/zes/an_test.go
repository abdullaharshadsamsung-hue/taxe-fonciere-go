package zes

import (
	"bytes"
	"strings"
	"testing"
)

// buildRecord returns a deterministic 400-byte record with distinct bytes per position.
func buildRecord() []byte {
	b := make([]byte, ANLen)
	copy(b[0:4], "2023")
	copy(b[4:6], "75")
	copy(b[6:7], "1")
	copy(b[7:10], "056")
	copy(b[10:13], "789")
	copy(b[13:15], "AN")
	copy(b[15:21], "230115")
	for i := 21; i < 398; i++ {
		b[i] = byte('a' + (i % 26))
	}
	copy(b[398:400], "UR")
	return b
}

func TestConstants(t *testing.T) {
	if ANLen != 400 || ANAidficLen+ANAdtsegLen != ANLen {
		t.Fatalf("bad lengths")
	}
	if ANAidficLen != 13 || ANAdtsegLen != 387 || ANFillerLen != 377 {
		t.Fatalf("bad group lengths")
	}
	if 2+6+ANFillerLen+2 != ANAdtsegLen {
		t.Fatalf("ADTSEG fields not contiguous")
	}
	if anOffDnuurc+2 != ANLen || anOffFiller+ANFillerLen != anOffDnuurc {
		t.Fatalf("offsets not contiguous")
	}
}

func TestNewAN(t *testing.T) {
	a := NewAN()
	if !bytes.Equal(a.Bytes(), bytes.Repeat([]byte{' '}, ANLen)) {
		t.Fatal("NewAN not all spaces")
	}
}

func TestParseAN(t *testing.T) {
	tests := []struct {
		name    string
		in      []byte
		wantErr bool
	}{
		{"exact", buildRecord(), false},
		{"short", make([]byte, 399), true},
		{"long", make([]byte, 401), true},
		{"empty", nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, err := ParseAN(tt.in)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err=%v wantErr=%v", err, tt.wantErr)
			}
			if !tt.wantErr && !bytes.Equal(a.Bytes(), tt.in) {
				t.Fatal("round trip mismatch")
			}
		})
	}
}

func TestBytesIsCopy(t *testing.T) {
	a := NewAN()
	b := a.Bytes()
	b[0] = 'Z'
	if a[0] != ' ' {
		t.Fatal("Bytes did not return copy")
	}
}

func TestReadFields(t *testing.T) {
	rec := buildRecord()
	a, err := ParseAN(rec)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"Aidfic", a.Aidfic(), string(rec[0:13])},
		{"JanRaw", a.JanRaw(), "2023"},
		{"Depdir", a.Depdir(), "751"},
		{"Codep", a.Codep(), "75"},
		{"Codir", a.Codir(), "1"},
		{"Ccocom", a.Ccocom(), "056"},
		{"Ccoifp", a.Ccoifp(), "789"},
		{"Adtseg", a.Adtseg(), string(rec[13:400])},
		{"Ccoenr", a.Ccoenr(), "AN"},
		{"Xdday6", a.Xdday6(), "230115"},
		{"Filler", a.Filler(), string(rec[21:398])},
		{"Dnuurc", a.Dnuurc(), "UR"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Fatalf("got %q want %q", tt.got, tt.want)
			}
		})
	}
	if a.Aidfic() != a.JanRaw()+a.Codep()+a.Codir()+a.Ccocom()+a.Ccoifp() {
		t.Fatal("AIDFIC composition mismatch")
	}
	if a.Depdir() != a.Codep()+a.Codir() {
		t.Fatal("DEPDIR composition mismatch")
	}
	if a.Adtseg() != a.Ccoenr()+a.Xdday6()+a.Filler()+a.Dnuurc() {
		t.Fatal("ADTSEG composition mismatch")
	}
	if len(a.Aidfic()) != 13 || len(a.Adtseg()) != 387 || len(a.Filler()) != 377 {
		t.Fatal("bad group lengths")
	}
}

func TestXdday6NotValidated(t *testing.T) {
	a := NewAN()
	a.SetXdday6("ab/c!d")
	if a.Xdday6() != "ab/c!d" {
		t.Fatalf("got %q", a.Xdday6())
	}
}

func TestJan(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    int
		wantNum bool
	}{
		{"2023", "2023", 2023, true},
		{"zeros", "0000", 0, true},
		{"9999", "9999", 9999, true},
		{"spaces", "    ", 0, false},
		{"letters", "20A3", 2013, false},
		{"clamp nibble", "20?3", 2093, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := NewAN()
			copy(a[0:4], tt.raw)
			if a.JanIsNumeric() != tt.wantNum {
				t.Fatalf("JanIsNumeric=%v", a.JanIsNumeric())
			}
			v, ok := a.Jan()
			if ok != tt.wantNum {
				t.Fatalf("ok=%v", ok)
			}
			if v < 0 || v > 9999 {
				t.Fatalf("out of range %d", v)
			}
			if tt.wantNum && v != tt.want {
				t.Fatalf("got %d want %d", v, tt.want)
			}
		})
	}
}

func TestSetJan(t *testing.T) {
	tests := []struct {
		in   int
		want string
	}{
		{2023, "2023"},
		{7, "0007"},
		{0, "0000"},
		{12345, "2345"},
		{100000, "0000"},
		{-2018, "2018"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			rec := buildRecord()
			a, _ := ParseAN(rec)
			a.SetJan(tt.in)
			if a.JanRaw() != tt.want {
				t.Fatalf("got %q want %q", a.JanRaw(), tt.want)
			}
			if !bytes.Equal(a[4:], rec[4:]) {
				t.Fatal("bytes outside JAN modified")
			}
			if !a.JanIsNumeric() {
				t.Fatal("not numeric after set")
			}
		})
	}
}

func TestSetters(t *testing.T) {
	type setter struct {
		name   string
		off, n int
		set    func(a *AN, v string)
		get    func(a *AN) string
	}
	setters := []setter{
		{"Aidfic", 0, 13, (*AN).SetAidfic, (*AN).Aidfic},
		{"Depdir", 4, 3, (*AN).SetDepdir, (*AN).Depdir},
		{"Codep", 4, 2, (*AN).SetCodep, (*AN).Codep},
		{"Codir", 6, 1, (*AN).SetCodir, (*AN).Codir},
		{"Ccocom", 7, 3, (*AN).SetCcocom, (*AN).Ccocom},
		{"Ccoifp", 10, 3, (*AN).SetCcoifp, (*AN).Ccoifp},
		{"Adtseg", 13, 387, (*AN).SetAdtseg, (*AN).Adtseg},
		{"Ccoenr", 13, 2, (*AN).SetCcoenr, (*AN).Ccoenr},
		{"Xdday6", 15, 6, (*AN).SetXdday6, (*AN).Xdday6},
		{"Filler", 21, 377, (*AN).SetFiller, (*AN).Filler},
		{"Dnuurc", 398, 2, (*AN).SetDnuurc, (*AN).Dnuurc},
	}
	for _, s := range setters {
		values := []string{"", "Q", strings.Repeat("Z", s.n), strings.Repeat("Y", s.n+3)}
		for _, v := range values {
			t.Run(s.name+"/"+v, func(t *testing.T) {
				rec := buildRecord()
				a, _ := ParseAN(rec)
				s.set(&a, v)
				want := v
				if len(want) > s.n {
					want = want[:s.n]
				} else {
					want += strings.Repeat(" ", s.n-len(want))
				}
				if got := s.get(&a); got != want {
					t.Fatalf("got %q want %q", got, want)
				}
				if !bytes.Equal(a[:s.off], rec[:s.off]) || !bytes.Equal(a[s.off+s.n:], rec[s.off+s.n:]) {
					t.Fatal("bytes outside field modified")
				}
			})
		}
	}
}

func TestSpecExamples(t *testing.T) {
	a := NewAN()
	a.SetCcoenr("A")
	if a.Ccoenr() != "A " {
		t.Fatalf("CCOENR %q", a.Ccoenr())
	}
	a.SetDnuurc("123")
	if a.Dnuurc() != "12" {
		t.Fatalf("DNUURC %q", a.Dnuurc())
	}
	a.SetCodep("13")
	a.SetCodir("2")
	if a.Depdir() != "132" {
		t.Fatalf("DEPDIR %q", a.Depdir())
	}
	a.SetJan(2023)
	if v, ok := a.Jan(); !ok || v != 2023 {
		t.Fatalf("Jan %d %v", v, ok)
	}
	if !strings.HasPrefix(a.Aidfic(), "2023132") {
		t.Fatalf("AIDFIC %q", a.Aidfic())
	}
}