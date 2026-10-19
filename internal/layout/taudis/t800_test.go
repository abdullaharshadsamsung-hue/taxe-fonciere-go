package taudis

import (
	"bytes"
	"testing"
)

func mustEntete(t *testing.T, s string) Entete {
	t.Helper()
	e, err := ParseEntete([]byte(s))
	if err != nil {
		t.Fatalf("ParseEntete(%q): %v", s, err)
	}
	return e
}

func TestConstants(t *testing.T) {
	if EnteteLen != 12 || RecordLen != 3000 || BodyLen != 2988 {
		t.Fatalf("bad constants %d %d %d", EnteteLen, RecordLen, BodyLen)
	}
	var e Entete
	if len(e) != 12 {
		t.Fatalf("Entete len %d", len(e))
	}
}

func TestNewEntete(t *testing.T) {
	e := NewEntete()
	if e.Key() != "            " {
		t.Fatalf("got %q", e.Key())
	}
}

func TestParseEntete(t *testing.T) {
	tests := []struct {
		name    string
		in      []byte
		want    string
		wantErr bool
	}{
		{"exact", []byte("ABCDEFGHIJKL"), "ABCDEFGHIJKL", false},
		{"longer", []byte("ABCDEFGHIJKLMNO"), "ABCDEFGHIJKL", false},
		{"short", []byte("ABC"), "", true},
		{"empty", nil, "", true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e, err := ParseEntete(tc.in)
			if (err != nil) != tc.wantErr {
				t.Fatalf("err=%v", err)
			}
			if !tc.wantErr && e.Key() != tc.want {
				t.Fatalf("got %q", e.Key())
			}
		})
	}
}

func TestSplitRecord(t *testing.T) {
	rec := bytes.Repeat([]byte("Z"), RecordLen)
	copy(rec, "ABCDEFGHIJKL")
	e, body, err := SplitRecord(rec)
	if err != nil {
		t.Fatal(err)
	}
	if e.Key() != "ABCDEFGHIJKL" || len(body) != BodyLen {
		t.Fatalf("got %q len %d", e.Key(), len(body))
	}
	body[0] = 'Q'
	if rec[EnteteLen] != 'Q' {
		t.Fatal("body should alias rec")
	}
	for _, n := range []int{0, 12, RecordLen - 1, RecordLen + 1} {
		if _, _, err := SplitRecord(make([]byte, n)); err == nil {
			t.Errorf("len %d: expected error", n)
		}
	}
}

func TestBytesIsCopy(t *testing.T) {
	e := mustEntete(t, "ABCDEFGHIJKL")
	b := e.Bytes()
	if string(b) != "ABCDEFGHIJKL" {
		t.Fatalf("got %q", b)
	}
	b[0] = 'x'
	if e.Key() != "ABCDEFGHIJKL" {
		t.Fatal("Bytes must return a copy")
	}
}

func TestViews(t *testing.T) {
	e := mustEntete(t, "ABCDEFGHIJKL")
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"Depdir", e.Depdir(), "ABC"},
		{"Codep", e.Codep(), "AB"},
		{"Codir", e.Codir(), "C"},
		{"Cletx1", e.Cletx1(), "DEFGHIJKL"},
		{"Ccocom", e.Ccocom(), "DEF"},
		{"Ccoifp", e.Ccoifp(), "GHI"},
		{"Ccpper", e.Ccpper(), "JKL"},
		{"Ctygc", e.Ctygc(), "D"},
		{"Ccogc", e.Ccogc(), "EFG"},
		{"Filler2", e.Filler2(), "HIJKL"},
		{"Ctyets", e.Ctyets(), "D"},
		{"Gcoet1Raw", e.Gcoet1Raw(), "EFG"},
		{"Dircom", e.Dircom(), "ABCDEF"},
		{"Filler4", e.Filler4(), "GHIJKL"},
		{"Key", e.Key(), "ABCDEFGHIJKL"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.want {
				t.Fatalf("got %q want %q", tc.got, tc.want)
			}
		})
	}
	if e.Dircom() != e.Codep()+e.Codir()+e.Ccocom() {
		t.Fatal("DIRCOM != CODEP||CODIR||CCOCOM")
	}
}

func TestTSEView(t *testing.T) {
	e := mustEntete(t, "1234X567ABCD")
	// Spec buffer: byte4='4'? Spec claims CTYETS='X'; use a buffer matching offsets.
	e = mustEntete(t, "123X567ABCDE")
	if e.Ctyets() != "X" {
		t.Fatalf("Ctyets %q", e.Ctyets())
	}
	v, ok := e.Gcoet1()
	if v != 567 || !ok || !e.Gcoet1IsNumeric() {
		t.Fatalf("Gcoet1 = %d %v", v, ok)
	}
	if e.Filler2() != "ABCDE" {
		t.Fatalf("Filler2 %q", e.Filler2())
	}
}

func TestGcoet1Invalid(t *testing.T) {
	tests := []struct {
		raw  string
		want int
	}{
		{"   ", 0},
		{"1A3", 113},
		{"ABC", 123},
		{"?? ", 990},
	}
	for _, tc := range tests {
		t.Run(tc.raw, func(t *testing.T) {
			e := mustEntete(t, "123X"+tc.raw+"ABCDE")
			v, ok := e.Gcoet1()
			if ok || e.Gcoet1IsNumeric() {
				t.Fatal("expected not numeric")
			}
			if v != tc.want || v < 0 || v > 999 {
				t.Fatalf("got %d want %d", v, tc.want)
			}
		})
	}
}

func TestSetGcoet1(t *testing.T) {
	tests := []struct {
		in   int
		want string
	}{
		{7, "007"}, {0, "000"}, {42, "042"}, {999, "999"},
		{1234, "234"}, {-7, "007"}, {-1005, "005"},
	}
	for _, tc := range tests {
		e := mustEntete(t, "ABCDEFGHIJKL")
		e.SetGcoet1(tc.in)
		if e.Gcoet1Raw() != tc.want {
			t.Errorf("SetGcoet1(%d)=%q want %q", tc.in, e.Gcoet1Raw(), tc.want)
		}
		if e.Ccogc() != tc.want {
			t.Errorf("CCOGC not updated: %q", e.Ccogc())
		}
		want := "ABCD" + tc.want + "HIJKL"
		if e.Key() != want {
			t.Errorf("key %q want %q", e.Key(), want)
		}
	}
	e := NewEntete()
	e.SetGcoet1(7)
	if e.Ccocom() != " 00" || e.Ccoifp() != "7  " {
		t.Fatalf("overlap: %q %q", e.Ccocom(), e.Ccoifp())
	}
}

func TestSetters(t *testing.T) {
	tests := []struct {
		name string
		set  func(e *Entete)
		want string
	}{
		{"Depdir", func(e *Entete) { e.SetDepdir("12A") }, "12A         "},
		{"DepdirShort", func(e *Entete) { e.SetDepdir("9") }, "9           "},
		{"Codep", func(e *Entete) { e.SetCodep("75") }, "75          "},
		{"CodepLong", func(e *Entete) { e.SetCodep("753") }, "75          "},
		{"Codir", func(e *Entete) { e.SetCodir("1") }, "  1         "},
		{"Cletx1", func(e *Entete) { e.SetCletx1("ABCDEFGHIJ") }, "   ABCDEFGHI"},
		{"Ccocom", func(e *Entete) { e.SetCcocom("123") }, "   123      "},
		{"CcocomShort", func(e *Entete) { e.SetCcocom("A") }, "   A        "},
		{"CcocomLong", func(e *Entete) { e.SetCcocom("ABCDE") }, "   ABC      "},
		{"Ccoifp", func(e *Entete) { e.SetCcoifp("456") }, "      456   "},
		{"Ccpper", func(e *Entete) { e.SetCcpper("789") }, "         789"},
		{"Ctygc", func(e *Entete) { e.SetCtygc("G") }, "   G        "},
		{"Ccogc", func(e *Entete) { e.SetCcogc("ABC") }, "    ABC     "},
		{"Filler2", func(e *Entete) { e.SetFiller2("VWXYZ") }, "       VWXYZ"},
		{"Ctyets", func(e *Entete) { e.SetCtyets("T") }, "   T        "},
		{"Dircom", func(e *Entete) { e.SetDircom("75A123") }, "75A123      "},
		{"Filler4", func(e *Entete) { e.SetFiller4("ABCDEFG") }, "      ABCDEF"},
		{"Empty", func(e *Entete) { e.SetCcocom("") }, "            "},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := NewEntete()
			tc.set(&e)
			if e.Key() != tc.want {
				t.Fatalf("got %q want %q", e.Key(), tc.want)
			}
		})
	}
}

func TestSharedStorage(t *testing.T) {
	e := NewEntete()
	e.SetCcocom("123")
	if e.Ctygc() != "1" || e.Ctyets() != "1" {
		t.Fatalf("Ctygc=%q Ctyets=%q", e.Ctygc(), e.Ctyets())
	}
	if e.Dircom()[3:6] != "123" {
		t.Fatalf("Dircom=%q", e.Dircom())
	}
	e.SetCodep("75")
	e.SetCodir("A")
	if e.Depdir() != "75A" || e.Dircom() != "75A123" {
		t.Fatalf("Depdir=%q Dircom=%q", e.Depdir(), e.Dircom())
	}
	if e.Dircom() != e.Codep()+e.Codir()+e.Ccocom() {
		t.Fatal("DIRCOM invariant broken")
	}
	e.SetFiller4("XXXXXX")
	if e.Ccoifp() != "XXX" || e.Ccpper() != "XXX" {
		t.Fatalf("filler4 not visible: %q", e.Key())
	}
	e.SetCletx1("")
	if e.Depdir() != "75A" || e.Cletx1() != "         " {
		t.Fatalf("got %q", e.Key())
	}
}