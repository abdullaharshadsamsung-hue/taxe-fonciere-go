package taudis

import (
	"bytes"
	"strings"
	"testing"
)

// buf returns a body of spaces with s written at 1-based position pos.
func bodyWith(pos int, s string) []byte {
	b := bytes.Repeat([]byte{' '}, BodyLen)
	copy(b[pos-1:], s)
	return b
}

func mustParse(t *testing.T, b []byte) Direction {
	t.Helper()
	d, err := ParseDirection(b)
	if err != nil {
		t.Fatalf("ParseDirection: %v", err)
	}
	return d
}

func TestLengths(t *testing.T) {
	if EnteteLen+BodyLen != 3000 {
		t.Fatalf("header+body = %d", EnteteLen+BodyLen)
	}
	if BodyLen != 2988 {
		t.Fatalf("BodyLen=%d", BodyLen)
	}
	d := NewDirection()
	if len(d.Bytes()) != 2988 {
		t.Fatal("Bytes length")
	}
}

func TestParseDirectionLength(t *testing.T) {
	tests := []struct {
		n       int
		wantErr bool
	}{{0, true}, {2987, true}, {2988, false}, {2989, true}, {3000, true}}
	for _, tc := range tests {
		_, err := ParseDirection(make([]byte, tc.n))
		if (err != nil) != tc.wantErr {
			t.Errorf("len %d: err=%v", tc.n, err)
		}
	}
}

func TestParseAlphaOffsets(t *testing.T) {
	tests := []struct {
		name string
		pos  int
		val  string
		get  func(*Direction) string
	}{
		{"DLETAR", 2, "D", (*Direction).Dletar},
		{"DLIDIR", 4, "DIRECTION DE PARIS XXXXXXXXXXX", (*Direction).Dlidir},
		{"JDACRE", 35, "20180101", (*Direction).Jdacre},
		{"DSIRED", 44, "12345678901234", (*Direction).Dsired},
		{"JANMIR", 68, "18", (*Direction).Janmir},
		{"JBAROL", 71, "2018-01-01", (*Direction).Jbarol},
		{"JDAHO1", 82, "20180315", (*Direction).Jdaho1},
		{"CTYROL", 115, "RG", (*Direction).Ctyrol},
		{"DHMRG", 118, "1234", (*Direction).Dhmrg},
		{"DHMRD", 123, "2345", (*Direction).Dhmrd},
		{"DHMRR", 128, "3456", (*Direction).Dhmrr},
		{"GCOVEF", 133, "V", (*Direction).Gcovef},
		{"GORVEF", 135, "C", (*Direction).Gorvef},
		{"GCOFDL", 137, "1", (*Direction).Gcofdl},
		{"CCOFDL", 139, "2", (*Direction).Ccofdl},
		{"GTACOM", 141, "*", (*Direction).Gtacom},
		{"GTACO2", 143, "*", (*Direction).Gtaco2},
		{"GTOCTX", 145, "T", (*Direction).Gtoctx},
		{"GTOROP", 147, "*", (*Direction).Gtorop},
		{"GPCTFX", 149, "G", (*Direction).Gpctfx},
		{"GTODGS", 151, "X", (*Direction).Gtodgs},
		{"GRODEP", 153, "Y", (*Direction).Grodep},
		{"GROTAS", 155, "Z", (*Direction).Grotas},
		{"PTBDEPRaw", 157, "0012345000", (*Direction).PtbdepRaw},
		{"PTBTASRaw", 168, "0001500000", (*Direction).PtbtasRaw},
		{"PNBAGRRaw", 179, "0002000000", (*Direction).PnbagrRaw},
		{"PNBAGNRaw", 190, "9999000000", (*Direction).PnbagnRaw},
		{"SIRETCHAGR", 201, "98765432109876", (*Direction).SiretChagr},
		{"DSITAS", 541, "11112222333344", (*Direction).Dsitas},
		{"VMAILG", 852, strings.Repeat("g", 50), (*Direction).Vmailg},
		{"VMAILD", 902, strings.Repeat("d", 50), (*Direction).Vmaild},
		{"VMGTAS", 953, strings.Repeat("t", 50), (*Direction).Vmgtas},
		{"VMDTAS", 1003, strings.Repeat("u", 50), (*Direction).Vmdtas},
		{"LIBRE", 1053, strings.Repeat("L", 1936), (*Direction).Libre},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			d := mustParse(t, bodyWith(tc.pos, tc.val))
			if got := tc.get(&d); got != tc.val {
				t.Errorf("got %q want %q", got, tc.val)
			}
		})
	}
}

func TestParseNumericOffsets(t *testing.T) {
	tests := []struct {
		name string
		pos  int
		raw  string
		want uint64
		get  func(*Direction) (uint64, bool)
	}{
		{"NBCOM1", 59, "123", 123, (*Direction).Nbcom1},
		{"JANROL", 63, "2018", 2018, (*Direction).Janrol},
		{"JSSHOM", 82, "20", 20, (*Direction).Jsshom},
		{"JANHO1", 84, "18", 18, (*Direction).Janho1},
		{"JMOHOM", 86, "03", 3, (*Direction).Jmohom},
		{"JJRHOM", 88, "15", 15, (*Direction).Jjrhom},
		{"JDAMI1", 91, "20180401", 20180401, (*Direction).Jdami1},
		{"JDAMJO", 100, "20180501", 20180501, (*Direction).Jdamjo},
		{"DDRLTF", 109, "7", 7, (*Direction).Ddrltf},
		{"DDRROL", 111, "456", 456, (*Direction).Ddrrol},
		{"DDRHMG", 118, "1", 1, (*Direction).Ddrhmg},
		{"DDRROG", 119, "234", 234, (*Direction).Ddrrog},
		{"DDRHMD", 123, "2", 2, (*Direction).Ddrhmd},
		{"DDRROD", 124, "345", 345, (*Direction).Ddrrod},
		{"DDRHMR", 128, "3", 3, (*Direction).Ddrhmr},
		{"DDRROR", 129, "456", 456, (*Direction).Ddrror},
		{"PTBDEP", 157, "0012345000", 12345000, (*Direction).Ptbdep},
		{"PTBTAS", 168, "0001500000", 1500000, (*Direction).Ptbtas},
		{"PNBAGR", 179, "0000000001", 1, (*Direction).Pnbagr},
		{"PNBAGN", 190, "9999999999", 9999999999, (*Direction).Pnbagn},
		{"TPNAGR", 216, "1234567890", 1234567890, (*Direction).Tpnagr},
		{"MPNAGR", 227, "0987654321", 987654321, (*Direction).Mpnagr},
		{"NBTSE", 538, "05", 5, (*Direction).Nbtse},
		{"TPBTAS", 556, "0000000100", 100, (*Direction).Tpbtas},
		{"MPBTAS", 567, "0000000200", 200, (*Direction).Mpbtas},
		{"MPBTAX", 578, "0000000300", 300, (*Direction).Mpbtax},
		{"VBOINF", 589, "0000001", 1, (*Direction).Vboinf},
		{"VBOSUP", 597, "9999999", 9999999, (*Direction).Vbosup},
		{"DDRAVI", 605, "0012345", 12345, (*Direction).Ddravi},
		{"NARSAE", 613, "000042", 42, (*Direction).Narsae},
		{"NARSIU", 620, "000043", 43, (*Direction).Narsiu},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			d := mustParse(t, bodyWith(tc.pos, tc.raw))
			v, ok := tc.get(&d)
			if !ok || v != tc.want {
				t.Errorf("got %d,%v want %d", v, ok, tc.want)
			}
		})
	}
}

func TestRateDecoding(t *testing.T) {
	d := mustParse(t, bodyWith(157, "0012345000"))
	v, ok := d.Ptbdep()
	if !ok || v/RateScale != 12 || v%RateScale != 345000 {
		t.Errorf("got %d", v)
	}
}

func TestNonNumeric(t *testing.T) {
	tests := []struct {
		name string
		pos  int
		raw  string
		get  func(*Direction) (uint64, bool)
	}{
		{"NBCOM1 spaces", 59, "   ", (*Direction).Nbcom1},
		{"PTBDEP letter", 157, "00A2345000", (*Direction).Ptbdep},
		{"JANROL mixed", 63, "20-8", (*Direction).Janrol},
	}
	for _, tc := range tests {
		d := mustParse(t, bodyWith(tc.pos, tc.raw))
		if _, ok := tc.get(&d); ok {
			t.Errorf("%s: expected ok=false", tc.name)
		}
	}
}

func TestPnbagnSentinel(t *testing.T) {
	tests := []struct {
		raw  string
		want bool
	}{{"9999000000", true}, {"9999999999", true}, {"0999900000", false}, {"0000000000", false}}
	for _, tc := range tests {
		d := mustParse(t, bodyWith(190, tc.raw))
		if got := d.PnbagnIsSentinel(); got != tc.want {
			t.Errorf("%s: got %v", tc.raw, got)
		}
	}
}

func TestTSEOffsets(t *testing.T) {
	for i := 1; i <= 5; i++ {
		base := 237 + 60*(i-1) // 1-based start of occurrence
		b := bytes.Repeat([]byte{' '}, BodyLen)
		copy(b[base-1+1:], "AB12")
		copy(b[base-1+6:], "0000000001")
		copy(b[base-1+17:], "0000000002")
		copy(b[base-1+28:], "0000000003")
		copy(b[base-1+39:], "0000000004")
		copy(b[base-1+50:], "0000000005")
		d := mustParse(t, b)
		e, err := d.TSE(i - 1)
		if err != nil {
			t.Fatal(err)
		}
		if e.Acoets() != "AB12" {
			t.Errorf("occ %d acoets %q", i, e.Acoets())
		}
		getters := []func() (uint64, bool){e.Tpbtse, e.Mpbtse, e.Mpbtsx, e.Tpntse, e.Mpntse}
		for k, g := range getters {
			if v, ok := g(); !ok || v != uint64(k+1) {
				t.Errorf("occ %d field %d = %d,%v", i, k, v, ok)
			}
		}
	}
}

func TestTSEIndexErrors(t *testing.T) {
	d := NewDirection()
	for _, i := range []int{-1, 5, 100} {
		if _, err := d.TSE(i); err == nil {
			t.Errorf("TSE(%d) expected error", i)
		}
	}
	for _, i := range []int{0, 4} {
		if _, err := d.TSE(i); err != nil {
			t.Errorf("TSE(%d): %v", i, err)
		}
	}
}

func TestTSEWriteThrough(t *testing.T) {
	d := NewDirection()
	e, _ := d.TSE(2)
	e.SetAcoets("XYZ")
	e.SetTpbtse(1)
	e.SetMpbtse(2)
	e.SetMpbtsx(3)
	e.SetTpntse(4)
	e.SetMpntse(5)
	b := d.Bytes()
	base := 236 + 120
	want := map[int]string{1: "XYZ ", 6: "0000000001", 17: "0000000002", 28: "0000000003", 39: "0000000004", 50: "0000000005"}
	for o, s := range want {
		if got := string(b[base+o : base+o+len(s)]); got != s {
			t.Errorf("offset %d got %q want %q", o, got, s)
		}
	}
}

func TestCounters(t *testing.T) {
	b := bytes.Repeat([]byte{' '}, BodyLen)
	for k := 1; k <= 16; k++ {
		copy(b[627+7*(k-1)-1:], strings.Repeat("0", 5)+string(rune('0'+k/10))+string(rune('0'+k%10)))
		copy(b[739+7*(k-1)-1:], "00001"+string(rune('0'+k/10))+string(rune('0'+k%10)))
	}
	d := mustParse(t, b)
	for k := 1; k <= 16; k++ {
		v, ok, err := d.Nbhcnt(k - 1)
		if err != nil || !ok || v != uint64(k) {
			t.Errorf("Nbhcnt(%d)=%d,%v,%v", k, v, ok, err)
		}
		v, ok, err = d.Nabond(k - 1)
		if err != nil || !ok || v != uint64(100+k) {
			t.Errorf("Nabond(%d)=%d,%v,%v", k, v, ok, err)
		}
	}
	if len(d.Anacnt()) != 112 || d.Anacnt() != string(b[626:738]) {
		t.Error("Anacnt raw")
	}
	if d.Anbond() != string(b[738:850]) {
		t.Error("Anbond raw")
	}
}

func TestCounterIndexErrors(t *testing.T) {
	d := NewDirection()
	for _, i := range []int{-1, 16, 99} {
		if _, _, err := d.Nbhcnt(i); err == nil {
			t.Errorf("Nbhcnt(%d)", i)
		}
		if _, _, err := d.Nabond(i); err == nil {
			t.Errorf("Nabond(%d)", i)
		}
		if err := d.SetNbhcnt(i, 1); err == nil {
			t.Errorf("SetNbhcnt(%d)", i)
		}
		if err := d.SetNabond(i, 1); err == nil {
			t.Errorf("SetNabond(%d)", i)
		}
	}
	before := d.Bytes()
	_ = d.SetNbhcnt(16, 5)
	if !bytes.Equal(before, d.Bytes()) {
		t.Error("failed set mutated")
	}
	if err := d.SetNbhcnt(15, 7); err != nil {
		t.Fatal(err)
	}
	if err := d.SetNabond(0, 8); err != nil {
		t.Fatal(err)
	}
	b := d.Bytes()
	if string(b[626+105:738]) != "0000007" || string(b[738:745]) != "0000008" {
		t.Error("counter positions")
	}
}

func TestNewDirection(t *testing.T) {
	d := NewDirection()
	if !d.IsDirection() || d.Dletar() != "D" {
		t.Error("letter")
	}
	b := d.Bytes()
	if b[0] != ' ' || b[2] != ' ' {
		t.Error("separators not spaces")
	}
	if d.Dlidir() != strings.Repeat(" ", 30) {
		t.Error("dlidir")
	}
	if d.Libre() != strings.Repeat(" ", 1936) {
		t.Error("libre")
	}
	if v, ok := d.Nbcom1(); !ok || v != 0 {
		t.Error("nbcom1")
	}
	if d.PtbdepRaw() != "0000000000" {
		t.Error("ptbdep")
	}
	e, _ := d.TSE(4)
	if v, ok := e.Mpntse(); !ok || v != 0 {
		t.Error("tse")
	}
	if e.Acoets() != "    " {
		t.Error("acoets")
	}
	d.SetDletar("C")
	if d.IsDirection() {
		t.Error("IsDirection with C")
	}
}

func TestSerializeAlpha(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"short", "PARIS", "PARIS" + strings.Repeat(" ", 25)},
		{"exact", strings.Repeat("A", 30), strings.Repeat("A", 30)},
		{"long", strings.Repeat("B", 30) + "XYZ", strings.Repeat("B", 30)},
		{"empty", "", strings.Repeat(" ", 30)},
	}
	for _, tc := range tests {
		d := NewDirection()
		d.SetDlidir(tc.in)
		b := d.Bytes()
		if string(b[3:33]) != tc.want || b[33] != ' ' {
			t.Errorf("%s: got %q", tc.name, b[3:34])
		}
	}
}

func TestSerializeNumeric(t *testing.T) {
	tests := []struct {
		name string
		set  func(*Direction, uint64)
		v    uint64
		off  int
		want string
	}{
		{"NBCOM1 7", (*Direction).SetNbcom1, 7, 58, "007"},
		{"NBCOM1 trunc", (*Direction).SetNbcom1, 12345, 58, "345"},
		{"PTBDEP 1.5", (*Direction).SetPtbdep, 1_500_000, 156, "0001500000"},
		{"JANROL", (*Direction).SetJanrol, 2018, 62, "2018"},
		{"DDRLTF trunc", (*Direction).SetDdrltf, 23, 108, "3"},
		{"TPNAGR", (*Direction).SetTpnagr, 42, 215, "0000000042"},
		{"VMAILD region untouched", (*Direction).SetNarsiu, 9, 619, "000009"},
	}
	for _, tc := range tests {
		d := NewDirection()
		tc.set(&d, tc.v)
		b := d.Bytes()
		if got := string(b[tc.off : tc.off+len(tc.want)]); got != tc.want {
			t.Errorf("%s: got %q", tc.name, got)
		}
	}
}

func TestRoundTrip(t *testing.T) {
	d := NewDirection()
	d.SetDlidir("PARIS")
	d.SetJdacre("20180101")
	d.SetDsired("12345678901234")
	d.SetNbcom1(7)
	d.SetJanrol(2018)
	d.SetJsshom(20)
	d.SetJanho1(18)
	d.SetJmohom(3)
	d.SetJjrhom(15)
	d.SetCtyrol("RD")
	d.SetDdrhmg(1)
	d.SetDdrrog(234)
	d.SetGcovef("V")
	d.SetGtoctx("C")
	d.SetGpctfx("P")
	d.SetPtbdep(12_345_000)
	d.SetPnbagn(9_999_000_000)
	d.SetSiretChagr("SIRETCHAGR0001")
	d.SetNbtse(3)
	d.SetMpbtax(777)
	d.SetVmailg("user")
	d.SetVmaild("@example.fr")
	d.SetVmdtas("@tas.fr")
	_ = d.SetNabond(5, 55)

	d2 := mustParse(t, d.Bytes())
	if d2 != d {
		t.Fatal("round trip mismatch")
	}
	if d2.Jdaho1() != "20180315" {
		t.Errorf("jdaho1 %q", d2.Jdaho1())
	}
	if d2.Dhmrg() != "1234" {
		t.Errorf("dhmrg %q", d2.Dhmrg())
	}
	if v, _ := d2.Ptbdep(); v != 12_345_000 {
		t.Error("ptbdep")
	}
	if !d2.PnbagnIsSentinel() {
		t.Error("sentinel")
	}
	if strings.TrimRight(d2.Vmaild(), " ") != "@example.fr" {
		t.Error("vmaild")
	}
	if v, _, _ := d2.Nabond(5); v != 55 {
		t.Error("nabond")
	}
}

func TestFillerPreserved(t *testing.T) {
	b := bytes.Repeat([]byte{'#'}, BodyLen)
	d := mustParse(t, b)
	d.SetDlidir("X")
	d.SetNbcom1(1)
	out := d.Bytes()
	if out[2] != '#' || out[33] != '#' || out[57] != '#' || out[61] != '#' {
		t.Error("separator overwritten")
	}
	if out[BodyLen-1] != '#' {
		t.Error("libre overwritten")
	}
	out[0] = 'Z'
	if d[0] == 'Z' {
		t.Error("Bytes not a copy")
	}
}