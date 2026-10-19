package linkage

import (
	"math"
	"math/big"
	"strings"
	"testing"
)

func TestCombatConstants(t *testing.T) {
	if CombatLen != 600 {
		t.Fatalf("CombatLen=%d", CombatLen)
	}
	if CombatTauCount != 8 || CombatBt13Count != 2 || CombatOmCount != 6 || CombatReservedCount != 12 {
		t.Fatal("occurs counts wrong")
	}
	c := NewCombat()
	if len(c.Bytes()) != 600 {
		t.Fatal("Bytes len")
	}
	if strings.Trim(string(c.Bytes()), " ") != "" {
		t.Fatal("NewCombat not spaces")
	}
}

func TestParseCombat(t *testing.T) {
	tests := []struct {
		name    string
		n       int
		wantErr bool
	}{
		{"exact", 600, false},
		{"short", 599, true},
		{"long", 601, true},
		{"empty", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := make([]byte, tt.n)
			for i := range b {
				b[i] = byte('A' + i%26)
			}
			c, err := ParseCombat(b)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err=%v", err)
			}
			if !tt.wantErr && string(c.Bytes()) != string(b) {
				t.Fatal("content mismatch")
			}
		})
	}
}

func TestDecodeCombatAndBytesCopy(t *testing.T) {
	var raw [CombatLen]byte
	raw[0] = 'B'
	c := DecodeCombat(raw)
	if c.Ccobnb() != "B" {
		t.Fatal("decode")
	}
	b := c.Bytes()
	b[0] = 'Z'
	if c.Ccobnb() != "B" {
		t.Fatal("Bytes not a copy")
	}
}

// Build a buffer by fixed 1-based positions and verify getters.
func TestCombatFieldOffsets(t *testing.T) {
	b := []byte(strings.Repeat(" ", 600))
	set := func(pos1 int, s string) { copy(b[pos1-1:], s) }
	set(1, "B")
	set(2, "2024")
	set(6, "75")
	set(8, "1")
	set(9, "056")
	set(12, "S")
	set(13, "P")
	set(14, "00042")
	set(19, "ABCDEFGH")
	set(27, "000000012C")  // MBACOM 123
	set(37, "000000012L")  // MBADEP -123
	set(47, "0000000001")  // MBAREG
	set(57, "0000000002")  // MBASYN
	set(67, "0000000003")  // MBACU
	set(77, "0000000004")  // MBATSE
	set(87, "000000000}")  // MBBT13(1) -0
	set(97, "000000009I")  // MBBT13(2) 99
	set(107, "FILLER-MBABA-20BYTES")
	set(127, "RA000000005E")
	set(139, "ZZ000000000s") // ascii negative 3
	set(199, "0000000077")
	set(309, "0000000099")
	set(319, "000000100{")
	set(329, "1000000000000000")
	set(345, "000000000A")
	set(355, "000000000R")
	set(365, "IFP")
	set(368, "PER")
	set(371, "TRAILER")
	c, err := ParseCombat(b)
	if err != nil {
		t.Fatal(err)
	}
	strs := []struct{ name, got, want string }{
		{"Ccobnb", c.Ccobnb(), "B"},
		{"DanRaw", c.DanRaw(), "2024"},
		{"Cc2dep", c.Cc2dep(), "75"},
		{"Ccodir", c.Ccodir(), "1"},
		{"Ac3dir", c.Ac3dir(), "751"},
		{"Ccocom", c.Ccocom(), "056"},
		{"Dsrpar", c.Dsrpar(), "S"},
		{"Aidfic", c.Aidfic(), "2024751056S"},
		{"Cgroup", c.Cgroup(), "P"},
		{"NnuproRaw", c.NnuproRaw(), "00042"},
		{"Anupro", c.Anupro(), "P00042"},
		{"A0008", c.A0008(), "ABCDEFGH"},
		{"MbabaFiller", c.MbabaFiller(), "FILLER-MBABA-20BYTES"},
		{"MbacomRaw", c.MbacomRaw(), "000000012C"},
		{"MbadepRaw", c.MbadepRaw(), "000000012L"},
		{"MbaregRaw", c.MbaregRaw(), "0000000001"},
		{"MbasynRaw", c.MbasynRaw(), "0000000002"},
		{"MbacuRaw", c.MbacuRaw(), "0000000003"},
		{"MbatseRaw", c.MbatseRaw(), "0000000004"},
		{"MvltimRaw", c.MvltimRaw(), "000000100{"},
		{"PvltomRaw", c.PvltomRaw(), "1000000000000000"},
		{"Mbage3Raw", c.Mbage3Raw(), "000000000A"},
		{"Mbata3Raw", c.Mbata3Raw(), "000000000R"},
		{"Ccoifp", c.Ccoifp(), "IFP"},
		{"Ccpper", c.Ccpper(), "PER"},
		{"Trailer", c.Trailer(), "TRAILER" + strings.Repeat(" ", 223)},
		{"Mbaba len", string(rune(len(c.Mbaba()))), string(rune(100))},
	}
	for _, s := range strs {
		if s.got != s.want {
			t.Errorf("%s=%q want %q", s.name, s.got, s.want)
		}
	}
	if v, ok := c.Dan(); v != 2024 || !ok {
		t.Errorf("Dan=%d %v", v, ok)
	}
	if v, ok := c.Nnupro(); v != 42 || !ok {
		t.Errorf("Nnupro=%d %v", v, ok)
	}
	for i, w := range "ABCDEFGH" {
		g, err := c.Gtotau(i)
		if err != nil || g != string(w) {
			t.Errorf("Gtotau(%d)=%q %v", i, g, err)
		}
	}
	signed := []struct {
		name string
		f    func() (int64, bool)
		want int64
	}{
		{"Mbacom", c.Mbacom, 123},
		{"Mbadep", c.Mbadep, -123},
		{"Mbareg", c.Mbareg, 1},
		{"Mbasyn", c.Mbasyn, 2},
		{"Mbacu", c.Mbacu, 3},
		{"Mbatse", c.Mbatse, 4},
		{"Mvltim", c.Mvltim, 1000},
		{"Mbage3", c.Mbage3, 1},
		{"Mbata3", c.Mbata3, -9},
	}
	for _, s := range signed {
		v, ok := s.f()
		if v != s.want || !ok {
			t.Errorf("%s=%d %v want %d", s.name, v, ok, s.want)
		}
	}
	if v, ok, err := c.Mbbt13(0); v != 0 || !ok || err != nil {
		t.Errorf("Mbbt13(0)=%d %v %v", v, ok, err)
	}
	if v, ok, err := c.Mbbt13(1); v != 99 || !ok || err != nil {
		t.Errorf("Mbbt13(1)=%d %v %v", v, ok, err)
	}
	if g, _ := c.Gtauom(0); g != "RA" {
		t.Errorf("Gtauom(0)=%q", g)
	}
	if v, ok, _ := c.Mbaom(0); v != 55 || !ok {
		t.Errorf("Mbaom(0)=%d %v", v, ok)
	}
	if v, ok, _ := c.Mbaom(1); v != -3 || !ok {
		t.Errorf("Mbaom(1)=%d %v", v, ok)
	}
	if r, _ := c.MbaomRaw(0); r != "000000005E" {
		t.Errorf("MbaomRaw=%q", r)
	}
	if v, ok, _ := c.Reserved(0); v != 77 || !ok {
		t.Errorf("Reserved(0)=%d", v)
	}
	if v, ok, _ := c.Reserved(11); v != 99 || !ok {
		t.Errorf("Reserved(11)=%d", v)
	}
	if v, ok := c.PvltomScaled(); v != PvltomScale || !ok {
		t.Errorf("PvltomScaled=%d", v)
	}
	if r, ok := c.Pvltom(); r.Cmp(big.NewRat(1, 1)) != 0 || !ok {
		t.Errorf("Pvltom=%v", r)
	}
}

func TestCombatNumericClassTest(t *testing.T) {
	c := NewCombat()
	if _, ok := c.Dan(); ok {
		t.Error("spaces Dan numeric")
	}
	if _, ok := c.Mbacom(); ok {
		t.Error("spaces Mbacom numeric")
	}
	if _, ok := c.PvltomScaled(); ok {
		t.Error("spaces Pvltom numeric")
	}
	if _, ok, _ := c.Reserved(3); ok {
		t.Error("spaces Reserved numeric")
	}
	c.SetMbacom(5)
	b := c.Bytes()
	b[26] = 'X'
	c2, _ := ParseCombat(b)
	if _, ok := c2.Mbacom(); ok {
		t.Error("bad digit numeric")
	}
	b[26] = '0'
	b[35] = '#'
	c3, _ := ParseCombat(b)
	if _, ok := c3.Mbacom(); ok {
		t.Error("bad sign numeric")
	}
}

func TestCombatSetSigned(t *testing.T) {
	tests := []struct {
		v    int64
		raw  string
		back int64
	}{
		{0, "000000000{", 0},
		{1, "000000000A", 1},
		{123, "000000012C", 123},
		{-123, "000000012L", -123},
		{-10, "000000001}", -10},
		{9999999999, "999999999I", 9999999999},
		{-9999999999, "999999999R", -9999999999},
		{12345678901, "234567890A", 2345678901},
		{-10000000000, "000000000{", 0},
		{math.MinInt64, "547758080}", -5477580800 / 10 * 10},
	}
	for _, tt := range tests {
		c := NewCombat()
		c.SetMbacom(tt.v)
		if c.MbacomRaw() != tt.raw {
			t.Errorf("Set(%d) raw=%q want %q", tt.v, c.MbacomRaw(), tt.raw)
		}
		if v, ok := c.Mbacom(); v != tt.back || !ok {
			t.Errorf("Set(%d) back=%d %v", tt.v, v, ok)
		}
	}
}

func TestCombatSignedSettersRoundTrip(t *testing.T) {
	c := NewCombat()
	type pair struct {
		name string
		set  func(int64)
		get  func() (int64, bool)
	}
	ps := []pair{
		{"Mbacom", c.SetMbacom, c.Mbacom},
		{"Mbadep", c.SetMbadep, c.Mbadep},
		{"Mbareg", c.SetMbareg, c.Mbareg},
		{"Mbasyn", c.SetMbasyn, c.Mbasyn},
		{"Mbacu", c.SetMbacu, c.Mbacu},
		{"Mbatse", c.SetMbatse, c.Mbatse},
		{"Mvltim", c.SetMvltim, c.Mvltim},
		{"Mbage3", c.SetMbage3, c.Mbage3},
		{"Mbata3", c.SetMbata3, c.Mbata3},
	}
	for i, p := range ps {
		p.set(int64(-(i + 1) * 1000))
	}
	for i, p := range ps {
		if v, ok := p.get(); v != int64(-(i+1)*1000) || !ok {
			t.Errorf("%s=%d", p.name, v)
		}
	}
	if strings.Trim(c.MbabaFiller(), " ") != "" || strings.Trim(c.Trailer(), " ") != "" {
		t.Error("filler clobbered")
	}
}

func TestCombatAlphaSetters(t *testing.T) {
	tests := []struct {
		name string
		set  func(string)
		get  func() string
		in   string
		want string
	}{
		{"Ccobnb", nil, nil, "BX", "B"},
		{"Cc2dep", nil, nil, "7", "7 "},
		{"Ccocom", nil, nil, "12345", "123"},
		{"Ccpper", nil, nil, "", "   "},
	}
	for _, tt := range tests {
		c := NewCombat()
		switch tt.name {
		case "Ccobnb":
			c.SetCcobnb(tt.in)
			tt.get = c.Ccobnb
		case "Cc2dep":
			c.SetCc2dep(tt.in)
			tt.get = c.Cc2dep
		case "Ccocom":
			c.SetCcocom(tt.in)
			tt.get = c.Ccocom
		case "Ccpper":
			c.SetCcpper(tt.in)
			tt.get = c.Ccpper
		}
		if g := tt.get(); g != tt.want {
			t.Errorf("%s=%q want %q", tt.name, g, tt.want)
		}
	}
	c := NewCombat()
	c.SetAidfic("2024751056S")
	c.SetAnupro("P00042")
	c.SetA0008("12345678")
	c.SetCcoifp("ABC")
	c.SetMbabaFiller("keep")
	c.SetTrailer("tail")
	c.SetDsrpar("D")
	c.SetCgroup("Q")
	c.SetCcodir("9")
	c.SetAc3dir("13")
	b := string(c.Bytes())
	if b[1:12] != "202413 056D" || b[12:18] != "Q00042" || b[18:26] != "12345678" ||
		b[364:367] != "ABC" || b[106:126] != "keep"+strings.Repeat(" ", 16) || b[370:374] != "tail" {
		t.Errorf("layout mismatch: %q", b[:130])
	}
	c.SetMbaba(strings.Repeat("9", 120))
	if c.Mbaba() != strings.Repeat("9", 100) || c.Get0Check() {
		t.Error("SetMbaba")
	}
}

// helper to ensure MBABA move didn't overflow into ABAOM (byte 126 untouched)
func (c *Combat) Get0Check() bool { return c[126] != ' ' }

func TestCombatUnsignedSetters(t *testing.T) {
	c := NewCombat()
	c.SetDan(12024)
	if c.DanRaw() != "2024" {
		t.Errorf("Dan truncation %q", c.DanRaw())
	}
	c.SetNnupro(7)
	if c.NnuproRaw() != "00007" {
		t.Errorf("Nnupro %q", c.NnuproRaw())
	}
	for i := 0; i < CombatReservedCount; i++ {
		if err := c.SetReserved(i, uint64(i)*11); err != nil {
			t.Fatal(err)
		}
	}
	b := string(c.Bytes())
	if b[198:208] != "0000000000" || b[308:318] != "0000000121" {
		t.Errorf("reserved order %q", b[198:318])
	}
	c.SetReserved(0, 12345678901)
	if v, _, _ := c.Reserved(0); v != 2345678901 {
		t.Errorf("reserved trunc %d", v)
	}
}

func TestCombatPvltom(t *testing.T) {
	tests := []struct {
		name string
		r    *big.Rat
		raw  string
	}{
		{"one", big.NewRat(1, 1), "1000000000000000"},
		{"half", big.NewRat(1, 2), "0500000000000000"},
		{"neg dropped", big.NewRat(-1, 4), "0250000000000000"},
		{"int trunc", big.NewRat(123, 10), "2300000000000000"},
		{"frac trunc", big.NewRat(1, 3), "0333333333333333"},
		{"zero", new(big.Rat), "0000000000000000"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCombat()
			if err := c.SetPvltom(tt.r); err != nil {
				t.Fatal(err)
			}
			if c.PvltomRaw() != tt.raw {
				t.Errorf("raw=%q want %q", c.PvltomRaw(), tt.raw)
			}
		})
	}
	c := NewCombat()
	if err := c.SetPvltom(nil); err == nil {
		t.Error("nil should fail")
	}
	c.SetPvltomScaled(12345)
	if c.PvltomRaw() != "0000000000012345" {
		t.Error("scaled")
	}
	r, _ := c.Pvltom()
	if r.Cmp(big.NewRat(12345, 1_000_000_000_000_000)) != 0 {
		t.Error("Pvltom rat")
	}
	if c[344] != ' ' || c[327] != ' ' {
		t.Error("pvltom overflow")
	}
}

func TestCombatGtauomOK(t *testing.T) {
	tests := []struct {
		v    string
		want bool
	}{
		{"  ", true}, {" P", true}, {"P ", true}, {"RA", true}, {"RB", true},
		{"RC", true}, {"RD", true}, {"RE", true},
		{"RF", false}, {"PP", false}, {"ra", false}, {"00", false}, {"R", false},
	}
	for _, tt := range tests {
		for i := 0; i < CombatOmCount; i++ {
			c := NewCombat()
			if err := c.SetGtauom(i, tt.v); err != nil {
				t.Fatal(err)
			}
			got, err := c.GtauomOK(i)
			if err != nil || got != tt.want {
				t.Errorf("GtauomOK(%d,%q)=%v %v", i, tt.v, got, err)
			}
		}
	}
}

func TestCombatAbaomLayout(t *testing.T) {
	c := NewCombat()
	for i := 0; i < CombatOmCount; i++ {
		c.SetGtauom(i, "R"+string(rune('A'+i)))
		c.SetMbaom(i, int64(i+1))
	}
	b := string(c.Bytes())
	for i := 0; i < CombatOmCount; i++ {
		off := 126 + i*12
		want := "R" + string(rune('A'+i)) + "000000000" + string(rune('A'+i))
		if b[off:off+12] != want {
			t.Errorf("ABAOM(%d)=%q want %q", i, b[off:off+12], want)
		}
	}
	c.SetMbbt13(0, 5)
	c.SetMbbt13(1, -5)
	if b := string(c.Bytes()); b[86:96] != "000000000E" || b[96:106] != "000000000N" {
		t.Errorf("mbbt13 %q", b[86:106])
	}
}

func TestCombatIndexErrors(t *testing.T) {
	c := NewCombat()
	for _, i := range []int{-1, 100} {
		checks := []struct {
			name string
			err  error
		}{
			{"Gtotau", func() error { _, e := c.Gtotau(i); return e }()},
			{"SetGtotau", c.SetGtotau(i, "x")},
			{"Mbbt13", func() error { _, _, e := c.Mbbt13(i); return e }()},
			{"SetMbbt13", c.SetMbbt13(i, 1)},
			{"Gtauom", func() error { _, e := c.Gtauom(i); return e }()},
			{"SetGtauom", c.SetGtauom(i, "x")},
			{"GtauomOK", func() error { _, e := c.GtauomOK(i); return e }()},
			{"Mbaom", func() error { _, _, e := c.Mbaom(i); return e }()},
			{"MbaomRaw", func() error { _, e := c.MbaomRaw(i); return e }()},
			{"SetMbaom", c.SetMbaom(i, 1)},
			{"Reserved", func() error { _, _, e := c.Reserved(i); return e }()},
			{"SetReserved", c.SetReserved(i, 1)},
		}
		for _, ch := range checks {
			if ch.err == nil {
				t.Errorf("%s(%d) expected error", ch.name, i)
			}
		}
	}
	bounds := []struct {
		name  string
		count int
		f     func(int) error
	}{
		{"Gtotau", CombatTauCount, func(i int) error { _, e := c.Gtotau(i); return e }},
		{"Mbbt13", CombatBt13Count, func(i int) error { _, _, e := c.Mbbt13(i); return e }},
		{"Gtauom", CombatOmCount, func(i int) error { _, e := c.Gtauom(i); return e }},
		{"Reserved", CombatReservedCount, func(i int) error { _, _, e := c.Reserved(i); return e }},
	}
	for _, b := range bounds {
		if b.f(b.count-1) != nil {
			t.Errorf("%s last index errored", b.name)
		}
		if b.f(b.count) == nil {
			t.Errorf("%s count index no error", b.name)
		}
	}
	if strings.Trim(string(c.Bytes()), " ") != "" {
		t.Error("failed setters mutated record")
	}
}