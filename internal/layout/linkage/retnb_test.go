package linkage

import (
	"strings"
	"testing"
)

func TestNewRetNBInitialState(t *testing.T) {
	r := NewRetNB()
	if len(r.Bytes()) != RetNBLen {
		t.Fatalf("len=%d", len(r.Bytes()))
	}
	if v, ok := r.Dan(); v != 0 || !ok {
		t.Errorf("Dan=%d,%v", v, ok)
	}
	if v, ok := r.Nnupro(); v != 0 || !ok {
		t.Errorf("Nnupro=%d,%v", v, ok)
	}
	if r.MctcomRaw() != "000000000{" {
		t.Errorf("MctcomRaw=%q", r.MctcomRaw())
	}
	if v, ok := r.Mcntsa(); v != 0 || !ok {
		t.Errorf("Mcntsa=%d,%v", v, ok)
	}
	if v, ok := r.Tcthfr(); v != 0 || !ok {
		t.Errorf("Tcthfr=%d,%v", v, ok)
	}
	if v, ok := r.Mctcad(); v != 0 || !ok {
		t.Errorf("Mctcad=%d,%v", v, ok)
	}
	for i := 0; i < 2; i++ {
		if v, ok, err := r.Mcnt13(i); v != 0 || !ok || err != nil {
			t.Errorf("Mcnt13(%d)", i)
		}
		if v, ok, err := r.Mdjt13(i); v != 0 || !ok || err != nil {
			t.Errorf("Mdjt13(%d)", i)
		}
	}
	if r.Filler() != strings.Repeat(" ", 10) || r.Trailer() != strings.Repeat(" ", 145) {
		t.Error("fillers not spaces")
	}
	if r.Cc2dep() != "  " || r.Ccobnb() != " " || r.Ccoifp() != "   " {
		t.Error("alnum not spaces")
	}
}

func TestParseRetNB(t *testing.T) {
	tests := []struct {
		n       int
		wantErr bool
	}{{600, false}, {599, true}, {601, true}, {0, true}}
	for _, tc := range tests {
		b := []byte(strings.Repeat("Z", tc.n))
		r, err := ParseRetNB(b)
		if (err != nil) != tc.wantErr {
			t.Errorf("n=%d err=%v", tc.n, err)
		}
		if err == nil && string(r.Bytes()) != string(b) {
			t.Error("content mismatch")
		}
	}
}

func TestRoundTrip(t *testing.T) {
	var b [RetNBLen]byte
	for i := range b {
		b[i] = byte('!' + i%90)
	}
	r := DecodeRetNB(b)
	if string(EncodeRetNB(&r)) != string(b[:]) {
		t.Error("decode/encode mismatch")
	}
	p, _ := ParseRetNB(b[:])
	if string(p.Bytes()) != string(b[:]) {
		t.Error("parse mismatch")
	}
	out := r.Bytes()
	out[0] = 0
	if r[0] == 0 {
		t.Error("Bytes not a copy")
	}
}

func posRec(t *testing.T) RetNB {
	// buffer where each 1-based position region is made of digits
	var b [RetNBLen]byte
	for i := range b {
		b[i] = byte('0' + i%10)
	}
	return DecodeRetNB(b)
}

func TestAlphaOffsets(t *testing.T) {
	tests := []struct {
		name  string
		get   func(*RetNB) string
		set   func(*RetNB, string)
		start int // 1-based
		ln    int
	}{
		{"Aidfic", (*RetNB).Aidfic, (*RetNB).SetAidfic, 1, 11},
		{"Ac3dir", (*RetNB).Ac3dir, (*RetNB).SetAc3dir, 5, 3},
		{"Cc2dep", (*RetNB).Cc2dep, (*RetNB).SetCc2dep, 5, 2},
		{"Ccodir", (*RetNB).Ccodir, (*RetNB).SetCcodir, 7, 1},
		{"Ccocom", (*RetNB).Ccocom, (*RetNB).SetCcocom, 8, 3},
		{"Dsrpar", (*RetNB).Dsrpar, (*RetNB).SetDsrpar, 11, 1},
		{"Anupro", (*RetNB).Anupro, (*RetNB).SetAnupro, 12, 6},
		{"Cgroup", (*RetNB).Cgroup, (*RetNB).SetCgroup, 12, 1},
		{"Ccobnb", (*RetNB).Ccobnb, (*RetNB).SetCcobnb, 18, 1},
		{"Filler", (*RetNB).Filler, (*RetNB).SetFiller, 214, 10},
		{"Ccoifp", (*RetNB).Ccoifp, (*RetNB).SetCcoifp, 450, 3},
		{"Ccpper", (*RetNB).Ccpper, (*RetNB).SetCcpper, 453, 3},
		{"Trailer", (*RetNB).Trailer, (*RetNB).SetTrailer, 456, 145},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := posRec(t)
			orig := r.Bytes()
			if got := tc.get(&r); got != string(orig[tc.start-1:tc.start-1+tc.ln]) {
				t.Errorf("get=%q", got)
			}
			r.SetCgroup("") // no-op guard avoided: restore
			r = DecodeRetNB([RetNBLen]byte(orig))
			tc.set(&r, "#")
			want := "#" + strings.Repeat(" ", tc.ln-1)
			if tc.get(&r) != want {
				t.Errorf("after set=%q", tc.get(&r))
			}
			after := r.Bytes()
			for i := range after {
				if (i < tc.start-1 || i >= tc.start-1+tc.ln) && after[i] != orig[i] {
					t.Fatalf("byte %d changed", i)
				}
			}
			tc.set(&r, strings.Repeat("Q", tc.ln+5))
			if tc.get(&r) != strings.Repeat("Q", tc.ln) {
				t.Error("no truncation")
			}
		})
	}
}

func TestUnsignedIdentifiers(t *testing.T) {
	r := NewRetNB()
	r.SetDan(2024)
	if r.DanRaw() != "2024" || r[0] != '2' {
		t.Errorf("DanRaw=%q", r.DanRaw())
	}
	r.SetDan(12345)
	if v, _ := r.Dan(); v != 2345 {
		t.Errorf("Dan trunc=%d", v)
	}
	r.SetNnupro(42)
	if r.NnuproRaw() != "00042" || string(r[12:17]) != "00042" {
		t.Errorf("Nnupro=%q", r.NnuproRaw())
	}
	r.SetDan(2024)
	r.SetCc2dep("75")
	if r.Aidfic()[:6] != "202475" {
		t.Errorf("Aidfic=%q", r.Aidfic())
	}
	r.SetAidfic("20a1")
	if _, ok := r.Dan(); ok {
		t.Error("non-digit DAN should be invalid")
	}
	r.SetAnupro("A1234X")
	if _, ok := r.Nnupro(); ok {
		t.Error("non-digit NNUPRO should be invalid")
	}
	if r.Cgroup() != "A" {
		t.Error("cgroup")
	}
}

type amtField struct {
	name  string
	get   func(*RetNB) (int64, bool)
	set   func(*RetNB, int64)
	start int
	ln    int
}

func amtFields() []amtField {
	return []amtField{
		{"Mctcom", (*RetNB).Mctcom, (*RetNB).SetMctcom, 19, 10},
		{"Mctdep", (*RetNB).Mctdep, (*RetNB).SetMctdep, 29, 10},
		{"Mctreg", (*RetNB).Mctreg, (*RetNB).SetMctreg, 39, 10},
		{"Mctsyn", (*RetNB).Mctsyn, (*RetNB).SetMctsyn, 49, 10},
		{"Mctcu", (*RetNB).Mctcu, (*RetNB).SetMctcu, 59, 10},
		{"Mcttse", (*RetNB).Mcttse, (*RetNB).SetMcttse, 69, 10},
		{"Mctca", (*RetNB).Mctca, (*RetNB).SetMctca, 99, 10},
		{"Mctbap", (*RetNB).Mctbap, (*RetNB).SetMctbap, 109, 10},
		{"Mctcap", (*RetNB).Mctcap, (*RetNB).SetMctcap, 119, 10},
		{"Mctcad", (*RetNB).Mctcad, (*RetNB).SetMctcad, 129, 5},
		{"Mcffix", (*RetNB).Mcffix, (*RetNB).SetMcffix, 144, 10},
		{"Mfa300", (*RetNB).Mfa300, (*RetNB).SetMfa300, 154, 10},
		{"Mcfano", (*RetNB).Mcfano, (*RetNB).SetMcfano, 154, 10},
		{"Mfn300", (*RetNB).Mfn300, (*RetNB).SetMfn300, 164, 10},
		{"Mcfnvn", (*RetNB).Mcfnvn, (*RetNB).SetMcfnvn, 164, 10},
		{"Mcfaba", (*RetNB).Mcfaba, (*RetNB).SetMcfaba, 174, 10},
		{"Mcfnvb", (*RetNB).Mcfnvb, (*RetNB).SetMcfnvb, 184, 10},
		{"Mcfaca", (*RetNB).Mcfaca, (*RetNB).SetMcfaca, 194, 10},
		{"Mcfnvc", (*RetNB).Mcfnvc, (*RetNB).SetMcfnvc, 204, 10},
		{"Mdgpla", (*RetNB).Mdgpla, (*RetNB).SetMdgpla, 224, 10},
		{"Mdgfap", (*RetNB).Mdgfap, (*RetNB).SetMdgfap, 234, 10},
		{"Mdgfnv", (*RetNB).Mdgfnv, (*RetNB).SetMdgfnv, 244, 10},
		{"Mdjcom", (*RetNB).Mdjcom, (*RetNB).SetMdjcom, 254, 10},
		{"Mdjsyn", (*RetNB).Mdjsyn, (*RetNB).SetMdjsyn, 264, 10},
		{"Mdjcu", (*RetNB).Mdjcu, (*RetNB).SetMdjcu, 274, 10},
		{"Mdjtse", (*RetNB).Mdjtse, (*RetNB).SetMdjtse, 284, 10},
		{"Mdjdep", (*RetNB).Mdjdep, (*RetNB).SetMdjdep, 294, 10},
		{"Tcthfr", (*RetNB).Tcthfr, (*RetNB).SetTcthfr, 324, 12},
		{"Tctfra", (*RetNB).Tctfra, (*RetNB).SetTctfra, 336, 12},
		{"Tctdu", (*RetNB).Tctdu, (*RetNB).SetTctdu, 348, 12},
		{"Mcnjae", (*RetNB).Mcnjae, (*RetNB).SetMcnjae, 360, 10},
		{"Mfaj1e", (*RetNB).Mfaj1e, (*RetNB).SetMfaj1e, 370, 10},
		{"Mfnj1e", (*RetNB).Mfnj1e, (*RetNB).SetMfnj1e, 380, 10},
		{"Mctcas", (*RetNB).Mctcas, (*RetNB).SetMctcas, 390, 10},
		{"Mfa800", (*RetNB).Mfa800, (*RetNB).SetMfa800, 400, 10},
		{"Mcaano", (*RetNB).Mcaano, (*RetNB).SetMcaano, 400, 10},
		{"Mfn800", (*RetNB).Mfn800, (*RetNB).SetMfn800, 410, 10},
		{"Mcanvn", (*RetNB).Mcanvn, (*RetNB).SetMcanvn, 410, 10},
		{"Mfa900", (*RetNB).Mfa900, (*RetNB).SetMfa900, 420, 10},
		{"Mfn900", (*RetNB).Mfn900, (*RetNB).SetMfn900, 430, 10},
		{"Mcoge3", (*RetNB).Mcoge3, (*RetNB).SetMcoge3, 440, 10},
	}
}

func TestSignedFieldsOffsetsAndEncoding(t *testing.T) {
	for _, f := range amtFields() {
		t.Run(f.name, func(t *testing.T) {
			r := NewRetNB()
			orig := r.Bytes()
			f.set(&r, -125)
			lo, hi := f.start-1, f.start-1+f.ln
			want := strings.Repeat("0", f.ln-3) + "12N"
			if got := string(r[lo:hi]); got != want {
				t.Errorf("bytes=%q want %q", got, want)
			}
			if v, ok := f.get(&r); v != -125 || !ok {
				t.Errorf("get=%d,%v", v, ok)
			}
			after := r.Bytes()
			for i := range after {
				if (i < lo || i >= hi) && after[i] != orig[i] {
					t.Fatalf("byte %d changed", i)
				}
			}
			f.set(&r, 120)
			if r[hi-1] != '{' {
				t.Errorf("last=%c", r[hi-1])
			}
			f.set(&r, -10)
			if r[hi-1] != '}' {
				t.Errorf("neg zero digit last=%c", r[hi-1])
			}
			// truncation: 1 followed by ln zeros -> 0, stored as +0
			big := int64(1)
			for i := 0; i < f.ln; i++ {
				big *= 10
			}
			f.set(&r, -big)
			if v, ok := f.get(&r); v != 0 || !ok || r[hi-1] != '{' {
				t.Errorf("trunc=%d,%v,%c", v, ok, r[hi-1])
			}
			f.set(&r, big+7)
			if v, _ := f.get(&r); v != 7 {
				t.Errorf("trunc pos=%d", v)
			}
		})
	}
}

func TestSignedDecodeVariants(t *testing.T) {
	tests := []struct {
		raw    string
		want   int64
		wantOK bool
	}{
		{"0000000125", 125, true},
		{"000000012E", 125, true},
		{"000000012N", -125, true},
		{"000000012u", -125, true},
		{"000000012}", -1200 / 10, true},
		{"00000 012E", 125, false},
		{"000000012 ", 120, false},
	}
	for _, tc := range tests {
		r := NewRetNB()
		copy(r[18:28], tc.raw)
		v, ok := r.Mctcom()
		if ok != tc.wantOK || (tc.wantOK && v != tc.want) {
			t.Errorf("%q: %d,%v", tc.raw, v, ok)
		}
		if r.MctcomRaw() != tc.raw {
			t.Errorf("raw %q", r.MctcomRaw())
		}
	}
}

func TestRawAccessors(t *testing.T) {
	r := posRec(t)
	tests := []struct {
		name string
		got  string
		lo   int
		n    int
	}{
		{"Mctdep", r.MctdepRaw(), 28, 10}, {"Mctreg", r.MctregRaw(), 38, 10},
		{"Mctsyn", r.MctsynRaw(), 48, 10}, {"Mctcu", r.MctcuRaw(), 58, 10},
		{"Mcttse", r.McttseRaw(), 68, 10}, {"Mctca", r.MctcaRaw(), 98, 10},
		{"Mctbap", r.MctbapRaw(), 108, 10}, {"Mctcap", r.MctcapRaw(), 118, 10},
		{"Mctcad", r.MctcadRaw(), 128, 5}, {"Mcntsa", r.McntsaRaw(), 133, 10},
	}
	for _, tc := range tests {
		if tc.got != string(r[tc.lo:tc.lo+tc.n]) {
			t.Errorf("%s=%q", tc.name, tc.got)
		}
	}
}

func TestMcntsaUnsigned(t *testing.T) {
	tests := []struct {
		in   int64
		want uint64
		raw  string
	}{
		{125, 125, "0000000125"},
		{-125, 125, "0000000125"},
		{12345678901, 2345678901, "2345678901"},
	}
	for _, tc := range tests {
		r := NewRetNB()
		r.SetMcntsa(tc.in)
		if v, ok := r.Mcntsa(); v != tc.want || !ok {
			t.Errorf("%d: %d,%v", tc.in, v, ok)
		}
		if string(r[133:143]) != tc.raw {
			t.Errorf("raw=%q", r[133:143])
		}
	}
	r := NewRetNB()
	copy(r[133:143], "000000012E")
	if _, ok := r.Mcntsa(); ok {
		t.Error("overpunch in unsigned should be non-numeric")
	}
}

func TestOccurs(t *testing.T) {
	type occ struct {
		name string
		get  func(*RetNB, int) (int64, bool, error)
		set  func(*RetNB, int, int64) error
		raw  func(*RetNB, int) (string, error)
		lo   int
	}
	for _, o := range []occ{
		{"MCNT13", (*RetNB).Mcnt13, (*RetNB).SetMcnt13, (*RetNB).Mcnt13Raw, 78},
		{"MDJT13", (*RetNB).Mdjt13, (*RetNB).SetMdjt13, (*RetNB).Mdjt13Raw, 303},
	} {
		t.Run(o.name, func(t *testing.T) {
			r := NewRetNB()
			for i := 0; i < 2; i++ {
				if err := o.set(&r, i, int64(-(i + 1))); err != nil {
					t.Fatal(err)
				}
			}
			for i := 0; i < 2; i++ {
				v, ok, err := o.get(&r, i)
				if err != nil || !ok || v != int64(-(i+1)) {
					t.Errorf("%d: %d,%v,%v", i, v, ok, err)
				}
				raw, err := o.raw(&r, i)
				if err != nil || raw != string(r[o.lo+i*10:o.lo+i*10+10]) {
					t.Errorf("raw %d=%q", i, raw)
				}
			}
			if string(r[o.lo:o.lo+20]) != "000000000J000000000K" {
				t.Errorf("bytes=%q", r[o.lo:o.lo+20])
			}
			for _, bad := range []int{-1, 2, 100} {
				if _, _, err := o.get(&r, bad); err == nil {
					t.Errorf("get %d no err", bad)
				}
				if err := o.set(&r, bad, 1); err == nil {
					t.Errorf("set %d no err", bad)
				}
				if _, err := o.raw(&r, bad); err == nil {
					t.Errorf("raw %d no err", bad)
				}
			}
		})
	}
}

func TestRedefinesShareStorage(t *testing.T) {
	tests := []struct {
		name    string
		setA    func(*RetNB, int64)
		getB    func(*RetNB) (int64, bool)
		setB    func(*RetNB, int64)
		getA    func(*RetNB) (int64, bool)
	}{
		{"MCFANO/MFA300", (*RetNB).SetMcfano, (*RetNB).Mfa300, (*RetNB).SetMfa300, (*RetNB).Mcfano},
		{"MCFNVN/MFN300", (*RetNB).SetMcfnvn, (*RetNB).Mfn300, (*RetNB).SetMfn300, (*RetNB).Mcfnvn},
		{"MCAANO/MFA800", (*RetNB).SetMcaano, (*RetNB).Mfa800, (*RetNB).SetMfa800, (*RetNB).Mcaano},
		{"MCANVN/MFN800", (*RetNB).SetMcanvn, (*RetNB).Mfn800, (*RetNB).SetMfn800, (*RetNB).Mcanvn},
	}
	for _, tc := range tests {
		r := NewRetNB()
		tc.setA(&r, 777)
		if v, _ := tc.getB(&r); v != 777 {
			t.Errorf("%s alias->base %d", tc.name, v)
		}
		tc.setB(&r, -9)
		if v, _ := tc.getA(&r); v != -9 {
			t.Errorf("%s base->alias %d", tc.name, v)
		}
	}
}

func TestSpacesRecordNotNumeric(t *testing.T) {
	var b [RetNBLen]byte
	for i := range b {
		b[i] = ' '
	}
	r := DecodeRetNB(b)
	if _, ok := r.Mctcom(); ok {
		t.Error("Mctcom")
	}
	if _, ok := r.Dan(); ok {
		t.Error("Dan")
	}
	if _, ok := r.Tctdu(); ok {
		t.Error("Tctdu")
	}
	if _, ok := r.Mcntsa(); ok {
		t.Error("Mcntsa")
	}
	if _, ok, _ := r.Mcnt13(0); ok {
		t.Error("Mcnt13")
	}
}

func TestFillerPreservedOnSets(t *testing.T) {
	r := NewRetNB()
	r.SetFiller("FILLERDATA")
	r.SetTrailer(strings.Repeat("T", 145))
	for _, f := range amtFields() {
		f.set(&r, -999999)
	}
	r.SetMcntsa(5)
	_ = r.SetMcnt13(1, 3)
	_ = r.SetMdjt13(1, 3)
	if r.Filler() != "FILLERDATA" || r.Trailer() != strings.Repeat("T", 145) {
		t.Error("filler clobbered")
	}
	p, err := ParseRetNB(r.Bytes())
	if err != nil || p != r {
		t.Error("round trip failed")
	}
}