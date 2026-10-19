package zes

import (
	"strings"
	"testing"
)

func buildSRRecord() []byte {
	b := []byte(strings.Repeat(" ", SRLen))
	put := func(pos int, s string) { copy(b[pos-1:], s) } // 1-based
	put(1, "2018")
	put(5, "75")
	put(7, "1")
	put(8, "056")
	put(11, "123")
	put(14, "SR")
	put(16, "07")
	for i := 0; i < 10; i++ {
		put(18+i*5, "01250")
		put(68+i*5, "00100")
		put(118+i*5, "99999")
	}
	rates := []string{"0023456700", "0000000001", "0000000002", "0000000003", "0000000004",
		"0000000005", "0000000006", "0000000007", "0000000008", "9999999999"}
	for i, r := range rates {
		put(168+i*10, r)
	}
	put(268, "EX")
	put(270, "O")
	put(271, "42")
	put(273, "0001500000")
	put(283, "0002500000")
	put(293, strings.Repeat("F", 50))
	put(343, "AB12CD34")
	put(351, "0509")
	put(355, strings.Repeat("G", 21))
	put(376, "93")
	put(378, "2")
	put(379, "0000700000")
	put(389, "0000800000")
	put(399, "UR")
	return b
}

func TestParseSRLength(t *testing.T) {
	tests := []struct {
		name    string
		n       int
		wantErr bool
	}{
		{"exact", 400, false},
		{"short", 399, true},
		{"long", 401, true},
		{"empty", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseSR(make([]byte, tt.n))
			if (err != nil) != tt.wantErr {
				t.Fatalf("err=%v wantErr=%v", err, tt.wantErr)
			}
		})
	}
}

func TestParseSRStringFields(t *testing.T) {
	s, err := ParseSR(buildSRRecord())
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name, got, want string
	}{
		{"JanRaw", s.JanRaw(), "2018"},
		{"Depdir", s.Depdir(), "751"},
		{"Codep", s.Codep(), "75"},
		{"Codir", s.Codir(), "1"},
		{"Ccocom", s.Ccocom(), "056"},
		{"Ccoifp", s.Ccoifp(), "123"},
		{"Aidfic", s.Aidfic(), "2018751056123"},
		{"Ccoenr", s.Ccoenr(), "SR"},
		{"Ccoexc", s.Ccoexc(), "EX"},
		{"Gimomi", s.Gimomi(), "O"},
		{"Filler1", s.Filler1(), strings.Repeat("F", 50)},
		{"Filler2", s.Filler2(), strings.Repeat("G", 21)},
		{"Ccddir", s.Ccddir(), "932"},
		{"Co2dep", s.Co2dep(), "93"},
		{"Ccodir", s.Ccodir(), "2"},
		{"Dnuurc", s.Dnuurc(), "UR"},
		{"Tpbasy", s.Tpbasy(), strings.Repeat("01250", 10)},
		{"Tpnbsy", s.Tpnbsy(), strings.Repeat("00100", 10)},
		{"Tpomsy", s.Tpomsy(), strings.Repeat("99999", 10)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Fatalf("got %q want %q", tt.got, tt.want)
			}
		})
	}
	if len(s.Adtseg()) != SRAdtsegLen || !strings.HasPrefix(s.Adtseg(), "SR07") {
		t.Fatalf("Adtseg %q", s.Adtseg())
	}
	a0, _ := s.Acoets(0)
	a1, _ := s.Acoets(1)
	if a0 != "AB12" || a1 != "CD34" {
		t.Fatalf("Acoets %q %q", a0, a1)
	}
}

func TestParseSRNumericFields(t *testing.T) {
	s, _ := ParseSR(buildSRRecord())
	tests := []struct {
		name string
		f    func() (uint64, bool)
		want uint64
	}{
		{"Jan", s.Jan, 2018},
		{"Ligbenefom", s.Ligbenefom, 7},
		{"TaucomB", s.TaucomB, 23456700},
		{"TausyndsfpB", s.TausyndsfpB, 1},
		{"TaucudfpvnB", s.TaucudfpvnB, 2},
		{"TautseB", s.TautseB, 3},
		{"Pnbtxa", s.Pnbtxa, 4},
		{"TaucomNB", s.TaucomNB, 5},
		{"TausyndsfpNB", s.TausyndsfpNB, 6},
		{"TaucudfpvnNB", s.TaucudfpvnNB, 7},
		{"TautseNB", s.TautseNB, 8},
		{"Pnbcaa", s.Pnbcaa, 9999999999},
		{"Ccobim", s.Ccobim, 42},
		{"Ptbgem", s.Ptbgem, 1500000},
		{"Pnbgem", s.Pnbgem, 2500000},
		{"Ptbtgp", s.Ptbtgp, 700000},
		{"Pnbtgp", s.Pnbtgp, 800000},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, ok := tt.f()
			if !ok || v != tt.want {
				t.Fatalf("got %d,%v want %d", v, ok, tt.want)
			}
		})
	}
	if v, _ := s.TaucomB(); v/SRRateScale != 23 || v%SRRateScale != 456700 {
		t.Fatalf("scale decode wrong: %d", v)
	}
}

func TestShareTables(t *testing.T) {
	s, _ := ParseSR(buildSRRecord())
	type acc func(int) (uint64, bool, error)
	tests := []struct {
		name string
		f    acc
		want uint64
	}{
		{"Pbasy1", s.Pbasy1, 1250},
		{"Pnbpsy", s.Pnbpsy, 100},
		{"Pomsy1", s.Pomsy1, 99999},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i := 0; i < SRSyndCount; i++ {
				v, ok, err := tt.f(i)
				if err != nil || !ok || v != tt.want {
					t.Fatalf("i=%d got %d,%v,%v", i, v, ok, err)
				}
			}
			for _, i := range []int{-1, SRSyndCount} {
				if _, _, err := tt.f(i); err == nil {
					t.Fatalf("expected error for %d", i)
				}
			}
		})
	}
	if v, _, _ := s.Pbasy1(0); v != 1250 || float64(v)/SRShareScale != 12.5 {
		t.Fatal("12.50 decode")
	}
}

func TestDnutse(t *testing.T) {
	s, _ := ParseSR(buildSRRecord())
	for i, want := range []uint64{5, 9} {
		v, ok, err := s.Dnutse(i)
		if err != nil || !ok || v != want {
			t.Fatalf("Dnutse(%d)=%d,%v,%v", i, v, ok, err)
		}
	}
	for _, i := range []int{-1, 2} {
		if _, _, err := s.Dnutse(i); err == nil {
			t.Fatal("expected error")
		}
		if _, err := s.Acoets(i); err == nil {
			t.Fatal("expected error")
		}
		if err := s.SetDnutse(i, 1); err == nil {
			t.Fatal("expected error")
		}
		if err := s.SetAcoets(i, "X"); err == nil {
			t.Fatal("expected error")
		}
	}
}

func TestInvalidNumeric(t *testing.T) {
	tests := []struct {
		name string
		pos  int
		val  string
		f    func(*SR) (uint64, bool)
	}{
		{"JanAlpha", 1, "20A8", func(s *SR) (uint64, bool) { return s.Jan() }},
		{"JanSpaces", 1, "    ", func(s *SR) (uint64, bool) { return s.Jan() }},
		{"RateAlpha", 168, "00234X6700", func(s *SR) (uint64, bool) { return s.TaucomB() }},
		{"Ccobim", 271, " 4", func(s *SR) (uint64, bool) { return s.Ccobim() }},
		{"Pnbtgp", 389, "000080000.", func(s *SR) (uint64, bool) { return s.Pnbtgp() }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := buildSRRecord()
			copy(b[tt.pos-1:], tt.val)
			s, _ := ParseSR(b)
			if _, ok := tt.f(&s); ok {
				t.Fatal("expected not numeric")
			}
		})
	}
	b := buildSRRecord()
	copy(b[0:], "20A8")
	s, _ := ParseSR(b)
	if s.JanIsNumeric() {
		t.Fatal("JanIsNumeric should be false")
	}
	copy(b[17:], "01A50")
	s, _ = ParseSR(b)
	if _, ok, err := s.Pbasy1(0); ok || err != nil {
		t.Fatal("expected invalid share")
	}
}

func TestSettersSerialize(t *testing.T) {
	s := NewSR()
	if string(s.Bytes()) != strings.Repeat(" ", SRLen) {
		t.Fatal("NewSR not spaces")
	}
	s.SetJan(2018)
	s.SetCodep("75")
	s.SetCodir("1")
	s.SetCcocom("056")
	s.SetCcoifp("123")
	s.SetCcoenr("SR")
	s.SetLigbenefom(7)
	_ = s.SetPbasy1(0, 1250)
	_ = s.SetPnbpsy(9, 100)
	_ = s.SetPomsy1(4, 99999)
	s.SetTaucomB(3500000)
	s.SetTausyndsfpB(1)
	s.SetTaucudfpvnB(2)
	s.SetTautseB(3)
	s.SetPnbtxa(4)
	s.SetTaucomNB(5)
	s.SetTausyndsfpNB(6)
	s.SetTaucudfpvnNB(7)
	s.SetTautseNB(8)
	s.SetPnbcaa(9)
	s.SetCcoexc("EX")
	s.SetGimomi("O")
	s.SetCcobim(42)
	s.SetPtbgem(10)
	s.SetPnbgem(11)
	_ = s.SetAcoets(0, "AB")
	_ = s.SetAcoets(1, "CD34XX")
	_ = s.SetDnutse(0, 5)
	_ = s.SetDnutse(1, 123)
	s.SetCo2dep("93")
	s.SetCcodir("2")
	s.SetPtbtgp(12)
	s.SetPnbtgp(13)
	s.SetDnuurc("UR")

	b := s.Bytes()
	if len(b) != SRLen {
		t.Fatal("len")
	}
	checks := []struct{ pos int; want string }{
		{1, "2018751056123"}, {14, "SR07"}, {18, "01250"}, {113, "00100"},
		{138, "99999"}, {168, "0003500000"}, {258, "0000000009"},
		{268, "EXO42"}, {273, "0000000010"}, {283, "0000000011"},
		{293, strings.Repeat(" ", 50)}, {343, "AB  CD34"}, {351, "0523"},
		{355, strings.Repeat(" ", 21)}, {376, "932"}, {379, "0000000012"},
		{389, "0000000013"}, {399, "UR"},
	}
	for _, c := range checks {
		if got := string(b[c.pos-1 : c.pos-1+len(c.want)]); got != c.want {
			t.Errorf("pos %d got %q want %q", c.pos, got, c.want)
		}
	}
	s2, err := ParseSR(b)
	if err != nil || s2 != s {
		t.Fatal("round trip")
	}
	if v, ok, _ := s2.Pbasy1(0); !ok || v != 1250 {
		t.Fatal("round trip share")
	}
}

func TestTruncationAndGroups(t *testing.T) {
	s := NewSR()
	s.SetJan(123456)
	if s.JanRaw() != "3456" {
		t.Fatalf("Jan trunc %q", s.JanRaw())
	}
	s.SetTaucomB(12345678901)
	if v, _ := s.TaucomB(); v != 2345678901 {
		t.Fatalf("rate trunc %d", v)
	}
	_ = s.SetPbasy1(1, 123456)
	if v, _, _ := s.Pbasy1(1); v != 23456 {
		t.Fatalf("share trunc %d", v)
	}
	s.SetDepdir("7512")
	if s.Codep() != "75" || s.Codir() != "1" {
		t.Fatal("depdir")
	}
	s.SetCcddir("93")
	if s.Co2dep() != "93" || s.Ccodir() != " " {
		t.Fatal("ccddir pad")
	}
	s.SetAidfic("2019")
	if s.JanRaw() != "2019" || s.Ccocom() != "   " {
		t.Fatal("aidfic")
	}
	s.SetFiller1("abc")
	s.SetFiller2("xyz")
	if s.Filler1() != "abc"+strings.Repeat(" ", 47) || s.Filler2() != "xyz"+strings.Repeat(" ", 18) {
		t.Fatal("fillers")
	}
	s.SetTpbasy("0010000200")
	if v, _, _ := s.Pbasy1(1); v != 200 {
		t.Fatal("tpbasy group")
	}
	s.SetTpnbsy("")
	s.SetTpomsy("")
	if s.Tpnbsy() != strings.Repeat(" ", 50) || s.Tpomsy() != strings.Repeat(" ", 50) {
		t.Fatal("tables")
	}
	s.SetAdtseg("ZZ")
	if s.Ccoenr() != "ZZ" || s.Dnuurc() != "  " || s.JanRaw() != "2019" {
		t.Fatal("adtseg")
	}
	// Bytes returns a copy
	b := s.Bytes()
	b[0] = 'X'
	if s.JanRaw() != "2019" {
		t.Fatal("Bytes aliasing")
	}
}