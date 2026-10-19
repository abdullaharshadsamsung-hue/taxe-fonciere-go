package linkage

import (
	"math"
	"strings"
	"testing"
)

func mkRecord(t *testing.T, edits map[int]string) []byte {
	t.Helper()
	b := []byte(strings.Repeat(" ", ComnbaLen))
	for pos, s := range edits { // pos is 1-based
		copy(b[pos-1:], s)
	}
	return b
}

func TestParseComnbaLength(t *testing.T) {
	tests := []struct {
		name    string
		n       int
		wantErr bool
	}{
		{"exact", ComnbaLen, false},
		{"short", ComnbaLen - 1, true},
		{"long", ComnbaLen + 1, true},
		{"empty", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseComnba(make([]byte, tt.n))
			if (err != nil) != tt.wantErr {
				t.Fatalf("err=%v wantErr=%v", err, tt.wantErr)
			}
		})
	}
}

func TestNewComnbaSpaces(t *testing.T) {
	c := NewComnba()
	if string(c.Bytes()) != strings.Repeat(" ", ComnbaLen) {
		t.Fatal("not all spaces")
	}
	if _, ok := c.Dan(); ok {
		t.Fatal("DAN should be non-numeric")
	}
	if _, ok := c.Mbacom(); ok {
		t.Fatal("MBACOM should be non-numeric")
	}
}

func TestDecodeAndBytesCopy(t *testing.T) {
	var raw [ComnbaLen]byte
	raw[0] = 'N'
	c := DecodeComnba(raw)
	b := c.Bytes()
	b[0] = 'Z'
	if c.Ccobnb() != "N" {
		t.Fatal("Bytes must return a copy")
	}
}

func TestParseFieldPositions(t *testing.T) {
	rec := mkRecord(t, map[int]string{
		1: "N", 2: "2024", 6: "57", 8: "1", 9: "123", 12: "S",
		13: "G", 14: "00042", 19: "ABCDEFGHIJK",
		30: "000000001{", 40: "000000002{", 50: "000000003{", 60: "000000004{",
		70: "000000005{", 80: "000000006{", 90: "000000007{", 100: "000000008{",
		110: "000000009{", 120: "000000010{", 130: "000000011{", 140: "000000012{",
		150: "000000013{", 160: "000000014{", 170: "000000015{", 180: "000000016{",
		190: "000000017{", 200: "000000018{", 210: "000000019{", 220: "000000020{",
		230: "000000021{", 240: "000000022{", 250: "000000023{", 260: "000000024{",
		270: "000000025{", 280: "ABC", 283: "XYZ", 286: "TRAIL",
	})
	c, err := ParseComnba(rec)
	if err != nil {
		t.Fatal(err)
	}
	strs := []struct{ name, got, want string }{
		{"CCOBNB", c.Ccobnb(), "N"}, {"DANRaw", c.DanRaw(), "2024"},
		{"CC2DEP", c.Cc2dep(), "57"}, {"CCODIR", c.Ccodir(), "1"},
		{"AC3DIR", c.Ac3dir(), "571"}, {"CCOCOM", c.Ccocom(), "123"},
		{"DSRPAR", c.Dsrpar(), "S"}, {"AIDFIC", c.Aidfic(), "2024571123S"},
		{"CGROUP", c.Cgroup(), "G"}, {"NNUPRORaw", c.NnuproRaw(), "00042"},
		{"ANUPRO", c.Anupro(), "G00042"}, {"A0011", c.A0011(), "ABCDEFGHIJK"},
		{"CCOIFP", c.Ccoifp(), "ABC"}, {"CCPPER", c.Ccpper(), "XYZ"},
		{"Trailer", c.Trailer(), "TRAIL" + strings.Repeat(" ", 310)},
		{"ABANB", c.Abanb(), string(rec[29:149])},
		{"ABAJAG", c.Abajag(), string(rec[149:269])},
	}
	for _, s := range strs {
		if s.got != s.want {
			t.Errorf("%s=%q want %q", s.name, s.got, s.want)
		}
	}
	if v, ok := c.Dan(); !ok || v != 2024 {
		t.Errorf("Dan=%d,%v", v, ok)
	}
	if v, ok := c.Nnupro(); !ok || v != 42 {
		t.Errorf("Nnupro=%d,%v", v, ok)
	}
	for i := 0; i < ComnbaTauCount; i++ {
		g, err := c.Gtotau(i)
		if err != nil || g != string(rune('A'+i)) {
			t.Errorf("Gtotau(%d)=%q,%v", i, g, err)
		}
	}
	nt0, _, _ := c.Mbnt13(0)
	nt1, _, _ := c.Mbnt13(1)
	jt0, _, _ := c.Mbjt13(0)
	jt1, _, _ := c.Mbjt13(1)
	amts := []struct {
		name string
		f    func() (int64, bool)
		want int64
	}{
		{"MBACOM", c.Mbacom, 10}, {"MBADEP", c.Mbadep, 20}, {"MBAREG", c.Mbareg, 30},
		{"MBASYN", c.Mbasyn, 40}, {"MBACU", c.Mbacu, 50}, {"MBATSE", c.Mbatse, 60},
		{"MBNT13_1", func() (int64, bool) { return nt0, true }, 70},
		{"MBNT13_2", func() (int64, bool) { return nt1, true }, 80},
		{"MBACA", c.Mbaca, 90}, {"MBABAP", c.Mbabap, 100}, {"MBACAA", c.Mbacaa, 110},
		{"MBAPOS", c.Mbapos, 120}, {"MBJCOM", c.Mbjcom, 130}, {"MBJSYN", c.Mbjsyn, 140},
		{"MBJCU", c.Mbjcu, 150}, {"MBJTSE", c.Mbjtse, 160},
		{"MBJT13_1", func() (int64, bool) { return jt0, true }, 170},
		{"MBJT13_2", func() (int64, bool) { return jt1, true }, 180},
		{"MBJDEP", c.Mbjdep, 190}, {"MB1PRE", c.Mb1pre, 200}, {"MBJPRE", c.Mbjpre, 210},
		{"MBJECO", c.Mbjeco, 220}, {"MBJESY", c.Mbjesy, 230}, {"MBJEIC", c.Mbjeic, 240},
		{"MBAGE3", c.Mbage3, 250},
	}
	for _, a := range amts {
		v, ok := a.f()
		if !ok || v != a.want {
			t.Errorf("%s=%d,%v want %d", a.name, v, ok, a.want)
		}
	}
}

func TestSignedDecode(t *testing.T) {
	tests := []struct {
		raw    string
		want   int64
		wantOK bool
	}{
		{"0000000123", 123, true},
		{"000000012C", 123, true},
		{"000000012{", 120, true},
		{"000000012L", -123, true},
		{"000000012}", -120, true},
		{"000000012s", -123, true},
		{"00000 012C", 0, false},
		{"000000012 ", 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.raw, func(t *testing.T) {
			c, _ := ParseComnba(mkRecord(t, map[int]string{30: tt.raw}))
			v, ok := c.Mbacom()
			if ok != tt.wantOK {
				t.Fatalf("ok=%v", ok)
			}
			if ok && v != tt.want {
				t.Fatalf("v=%d want %d", v, tt.want)
			}
			if c.MbacomRaw() != tt.raw {
				t.Fatalf("raw=%q", c.MbacomRaw())
			}
		})
	}
}

func TestSetSignedEncoding(t *testing.T) {
	tests := []struct {
		in      int64
		wantRaw string
		wantV   int64
	}{
		{0, "000000000{", 0},
		{123, "000000012C", 123},
		{-123, "000000012L", -123},
		{-120, "000000012}", -120},
		{9999999999, "999999999I", 9999999999},
		{12345678901, "234567890{", 2345678900},
		{-10000000000, "000000000{", 0},
		{math.MinInt64, "547758080{", 0},
	}
	for _, tt := range tests {
		c := NewComnba()
		c.SetMbage3(tt.in)
		if c.Mbage3Raw() != tt.wantRaw {
			t.Errorf("Set(%d) raw=%q want %q", tt.in, c.Mbage3Raw(), tt.wantRaw)
		}
		if tt.in == math.MinInt64 {
			continue
		}
		if v, ok := c.Mbage3(); !ok || v != tt.wantV {
			t.Errorf("Set(%d) got %d,%v", tt.in, v, ok)
		}
	}
}

func TestSetUnsignedTruncation(t *testing.T) {
	c := NewComnba()
	c.SetDan(12025)
	if c.DanRaw() != "2025" {
		t.Errorf("DAN=%q", c.DanRaw())
	}
	c.SetNnupro(1234567)
	if c.NnuproRaw() != "34567" {
		t.Errorf("NNUPRO=%q", c.NnuproRaw())
	}
	c.SetDan(7)
	if c.DanRaw() != "0007" {
		t.Errorf("DAN=%q", c.DanRaw())
	}
}

func TestAlphaMoves(t *testing.T) {
	c := NewComnba()
	tests := []struct {
		name string
		set  func(string)
		get  func() string
		in   string
		want string
	}{
		{"CCOBNB trunc", c.SetCcobnb, c.Ccobnb, "NB", "N"},
		{"CCOCOM pad", c.SetCcocom, c.Ccocom, "1", "1  "},
		{"CCOIFP trunc", c.SetCcoifp, c.Ccoifp, "ABCD", "ABC"},
		{"CCPPER", c.SetCcpper, c.Ccpper, "XY", "XY "},
		{"AC3DIR", c.SetAc3dir, c.Ac3dir, "571", "571"},
		{"CC2DEP", c.SetCc2dep, c.Cc2dep, "2A", "2A"},
		{"CCODIR", c.SetCcodir, c.Ccodir, "9", "9"},
		{"DSRPAR", c.SetDsrpar, c.Dsrpar, "", " "},
		{"CGROUP", c.SetCgroup, c.Cgroup, "P", "P"},
		{"AIDFIC", c.SetAidfic, c.Aidfic, "2024", "2024       "},
		{"ANUPRO", c.SetAnupro, c.Anupro, "A12345X", "A12345"},
		{"A0011", c.SetA0011, c.A0011, "XX", "XX         "},
		{"Trailer", c.SetTrailer, c.Trailer, "Z", "Z" + strings.Repeat(" ", 314)},
	}
	for _, tt := range tests {
		tt.set(tt.in)
		if g := tt.get(); g != tt.want {
			t.Errorf("%s=%q want %q", tt.name, g, tt.want)
		}
	}
	c.SetAbanb("")
	if c.Abanb() != strings.Repeat(" ", 120) {
		t.Error("ABANB")
	}
	c.SetAbajag("1")
	if len(c.Abajag()) != 120 || c.Abajag()[0] != '1' {
		t.Error("ABAJAG")
	}
}

func TestIndexedAccessorsRange(t *testing.T) {
	c := NewComnba()
	for _, i := range []int{-1, ComnbaTauCount} {
		if _, err := c.Gtotau(i); err == nil {
			t.Errorf("Gtotau(%d) no err", i)
		}
		if err := c.SetGtotau(i, "X"); err == nil {
			t.Errorf("SetGtotau(%d) no err", i)
		}
	}
	for _, i := range []int{-1, 2} {
		if _, _, err := c.Mbnt13(i); err == nil {
			t.Error("Mbnt13")
		}
		if err := c.SetMbnt13(i, 1); err == nil {
			t.Error("SetMbnt13")
		}
		if _, err := c.Mbnt13Raw(i); err == nil {
			t.Error("Mbnt13Raw")
		}
		if _, _, err := c.Mbjt13(i); err == nil {
			t.Error("Mbjt13")
		}
		if err := c.SetMbjt13(i, 1); err == nil {
			t.Error("SetMbjt13")
		}
		if _, err := c.Mbjt13Raw(i); err == nil {
			t.Error("Mbjt13Raw")
		}
	}
	if string(c.Bytes()) != strings.Repeat(" ", ComnbaLen) {
		t.Error("out-of-range set modified record")
	}
}

func TestRoundTripAndPositions(t *testing.T) {
	c := NewComnba()
	c.SetCcobnb("N")
	c.SetDan(2024)
	c.SetCc2dep("57")
	c.SetCcodir("0")
	c.SetCcocom("463")
	c.SetDsrpar("A")
	c.SetCgroup("+")
	c.SetNnupro(12345)
	for i := 0; i < ComnbaTauCount; i++ {
		_ = c.SetGtotau(i, "T")
	}
	setters := []struct {
		set  func(int64)
		get  func() (int64, bool)
		raw  func() string
		pos  int
		val  int64
	}{
		{c.SetMbacom, c.Mbacom, c.MbacomRaw, 30, 1},
		{c.SetMbadep, c.Mbadep, c.MbadepRaw, 40, -2},
		{c.SetMbareg, c.Mbareg, c.MbaregRaw, 50, 3},
		{c.SetMbasyn, c.Mbasyn, c.MbasynRaw, 60, -4},
		{c.SetMbacu, c.Mbacu, c.MbacuRaw, 70, 5},
		{c.SetMbatse, c.Mbatse, c.MbatseRaw, 80, 6},
		{c.SetMbaca, c.Mbaca, c.MbacaRaw, 110, 7},
		{c.SetMbabap, c.Mbabap, c.MbabapRaw, 120, 8},
		{c.SetMbacaa, c.Mbacaa, c.MbacaaRaw, 130, 9},
		{c.SetMbapos, c.Mbapos, c.MbaposRaw, 140, -10},
		{c.SetMbjcom, c.Mbjcom, c.MbjcomRaw, 150, 11},
		{c.SetMbjsyn, c.Mbjsyn, c.MbjsynRaw, 160, 12},
		{c.SetMbjcu, c.Mbjcu, c.MbjcuRaw, 170, 13},
		{c.SetMbjtse, c.Mbjtse, c.MbjtseRaw, 180, 14},
		{c.SetMbjdep, c.Mbjdep, c.MbjdepRaw, 210, 15},
		{c.SetMb1pre, c.Mb1pre, c.Mb1preRaw, 220, 16},
		{c.SetMbjpre, c.Mbjpre, c.MbjpreRaw, 230, 17},
		{c.SetMbjeco, c.Mbjeco, c.MbjecoRaw, 240, 18},
		{c.SetMbjesy, c.Mbjesy, c.MbjesyRaw, 250, 19},
		{c.SetMbjeic, c.Mbjeic, c.MbjeicRaw, 260, 20},
		{c.SetMbage3, c.Mbage3, c.Mbage3Raw, 270, 9876543210},
	}
	for _, s := range setters {
		s.set(s.val)
	}
	_ = c.SetMbnt13(0, 21)
	_ = c.SetMbnt13(1, -22)
	_ = c.SetMbjt13(0, 23)
	_ = c.SetMbjt13(1, 24)
	c.SetCcoifp("001")
	c.SetCcpper("002")

	b := c.Bytes()
	if len(b) != ComnbaLen {
		t.Fatal("len")
	}
	if string(b[285:]) != strings.Repeat(" ", 315) {
		t.Error("filler not spaces")
	}
	if string(b[0:18]) != "N2024570463A+12345" {
		t.Errorf("header=%q", b[0:18])
	}
	if string(b[279:285]) != "001002" {
		t.Errorf("codes=%q", b[279:285])
	}
	d, err := ParseComnba(b)
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range setters {
		if s.raw() != string(b[s.pos-1:s.pos+9]) {
			t.Errorf("pos %d raw mismatch", s.pos)
		}
		_ = d
		if v, ok := s.get(); !ok || v != s.val {
			t.Errorf("pos %d got %d,%v want %d", s.pos, v, ok, s.val)
		}
	}
	idx := []struct {
		raw  func(int) (string, error)
		get  func(int) (int64, bool, error)
		i    int
		pos  int
		want int64
	}{
		{d.Mbnt13Raw, d.Mbnt13, 0, 90, 21},
		{d.Mbnt13Raw, d.Mbnt13, 1, 100, -22},
		{d.Mbjt13Raw, d.Mbjt13, 0, 190, 23},
		{d.Mbjt13Raw, d.Mbjt13, 1, 200, 24},
	}
	for _, x := range idx {
		r, err := x.raw(x.i)
		if err != nil || r != string(b[x.pos-1:x.pos+9]) {
			t.Errorf("raw pos %d=%q", x.pos, r)
		}
		v, ok, err := x.get(x.i)
		if err != nil || !ok || v != x.want {
			t.Errorf("pos %d got %d", x.pos, v)
		}
	}
	if v, _ := d.Dan(); v != 2024 {
		t.Error("dan")
	}
	if v, _ := d.Nnupro(); v != 12345 {
		t.Error("nnupro")
	}
	if d.A0011() != strings.Repeat("T", 11) {
		t.Error("gtotau")
	}
}