package taudis

import (
	"bytes"
	"fmt"
	"testing"
)

func mustNew(t *testing.T) IFPTreso {
	t.Helper()
	r, err := NewIFPTreso()
	if err != nil {
		t.Fatalf("NewIFPTreso: %v", err)
	}
	return r
}

func fieldOf(t *testing.T, name string) tresoField {
	t.Helper()
	f, ok := tresoFields[name]
	if !ok {
		t.Fatalf("field %s missing", name)
	}
	return f
}

func TestLayoutBuildsToBodyLen(t *testing.T) {
	if tresoLayoutErr != nil {
		t.Fatalf("layout error: %v", tresoLayoutErr)
	}
	if EnteteLen+BodyLen != 3000 {
		t.Fatalf("entete+body = %d", EnteteLen+BodyLen)
	}
	names := IFPTresoFieldNames()
	if len(names) == 0 || names[0] != "DLETAR" {
		t.Fatalf("unexpected first name: %v", names)
	}
	names[0] = "MUTATED"
	if IFPTresoFieldNames()[0] != "DLETAR" {
		t.Fatal("IFPTresoFieldNames must return a copy")
	}
}

func TestFieldOffsets(t *testing.T) {
	tests := []struct {
		name string
		off  int
		n    int
		kind tresoKind
	}{
		{"DLETAR", 1, 1, tresoAlpha},
		{"DNULTF", 3, 1, tresoAlpha},
		{"DNURO", 5, 3, tresoUnsigned},
		{"GCOANU", 9, 1, tresoAlpha},
		{"DACROL", 11, 3, tresoAlpha},
		{"GCOVEF", 15, 1, tresoAlpha},
		{"GORVEF", 17, 1, tresoAlpha},
		{"CCOFDL", 19, 1, tresoAlpha},
		{"CCPDIR", 21, 3, tresoAlpha},
		{"CSPER", 25, 3, tresoAlpha},
		{"PTBCOM", 29, 10, tresoUnsigned},
		{"PTBSYN", 40, 10, tresoUnsigned},
		{"PTBTSN(1)", 29 + 4*11, 10, tresoUnsigned},
		{"PTBTSN(2)", 29 + 5*11, 10, tresoUnsigned},
		{"PBBOME", 29 + 25*11, 10, tresoUnsigned},
		{"NTXOM", 315, 1, tresoUnsigned},
		{"TBBCO", 317, 15, tresoUnsigned},
		{"TBBTSN(2)", 317 + 5*16, 15, tresoUnsigned},
		{"TBBOM2", 317 + 23*16, 15, tresoUnsigned},
		{"MCBCO3", 317 + 24*16, 10, tresoUnsigned},
		{"LIBRE", BodyLen - 312, 312, tresoAlpha},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := fieldOf(t, tc.name)
			if f.off != tc.off || f.n != tc.n || f.kind != tc.kind {
				t.Errorf("got %+v, want off=%d n=%d kind=%d", f, tc.off, tc.n, tc.kind)
			}
		})
	}
}

func TestNewIFPTreso(t *testing.T) {
	r := mustNew(t)
	if r.Dletar() != "I" || !r.IsIFPTreso() {
		t.Fatalf("Dletar=%q", r.Dletar())
	}
	if r[0] != ' ' {
		t.Errorf("filler at 0 = %q", r[0])
	}
	if v, ok := r.Ptbcom(); v != 0 || !ok {
		t.Errorf("Ptbcom=%d,%v", v, ok)
	}
	if v, ok, err := r.Signed("MLBCOM"); err != nil || v != 0 || !ok {
		t.Errorf("MLBCOM=%d,%v,%v", v, ok, err)
	}
	if raw, _ := r.Raw("MLBCOM"); raw != "000000000{" {
		t.Errorf("MLBCOM raw=%q", raw)
	}
	if r.Libre() != string(bytes.Repeat([]byte(" "), 312)) {
		t.Error("LIBRE not spaces")
	}
	r.SetDletar("C")
	if r.IsIFPTreso() {
		t.Error("expected not IFP")
	}
}

func TestParseAndRoundTrip(t *testing.T) {
	tests := []struct {
		name    string
		n       int
		wantErr bool
	}{
		{"short", BodyLen - 1, true},
		{"long", BodyLen + 1, true},
		{"empty", 0, true},
		{"exact", BodyLen, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			b := make([]byte, tc.n)
			for i := range b {
				b[i] = byte('A' + i%26)
			}
			r, err := ParseIFPTreso(b)
			if (err != nil) != tc.wantErr {
				t.Fatalf("err=%v", err)
			}
			if !tc.wantErr && !bytes.Equal(r.Bytes(), b) {
				t.Error("round trip mismatch")
			}
		})
	}
	// fillers preserved through modification
	b := bytes.Repeat([]byte("#"), BodyLen)
	r, _ := ParseIFPTreso(b)
	_ = r.SetNum("PTBCOM", 5)
	out := r.Bytes()
	if out[28] != '#' || out[39] != '#' || out[0] != '#' {
		t.Error("filler separators modified")
	}
	out[0] = 'Z'
	if r[0] != '#' {
		t.Error("Bytes must return a copy")
	}
}

func TestNumericDecoding(t *testing.T) {
	tests := []struct {
		field  string
		raw    string
		want   uint64
		wantOK bool
	}{
		{"PTBCOM", "0012345000", 12345000, true},
		{"PCFSYA", "10000", 10000, true},
		{"DNURO", "123", 123, true},
		{"DNURO", "A12", 112, false},
		{"DNURO", "   ", 0, false},
	}
	for _, tc := range tests {
		t.Run(tc.field+"_"+tc.raw, func(t *testing.T) {
			r := mustNew(t)
			if err := r.SetRaw(tc.field, tc.raw); err != nil {
				t.Fatal(err)
			}
			v, ok, err := r.Num(tc.field)
			if err != nil || v != tc.want || ok != tc.wantOK {
				t.Errorf("got %d,%v,%v want %d,%v", v, ok, err, tc.want, tc.wantOK)
			}
		})
	}
	r := mustNew(t)
	_ = r.SetRaw("PTBCOM", "0012345000")
	if v, ok := r.Ptbcom(); v != 12345000 || !ok {
		t.Errorf("Ptbcom=%d", v)
	}
	_ = r.SetNum("DNURO", 12345)
	if raw, _ := r.Raw("DNURO"); raw != "345" {
		t.Errorf("truncation raw=%q", raw)
	}
}

func TestTypedRateAccessors(t *testing.T) {
	tests := []struct {
		field string
		get   func(r *IFPTreso) (uint64, bool)
	}{
		{"NTXOM", (*IFPTreso).Ntxom}, // 9(1): must get a single-digit value
		{"PTBCOM", (*IFPTreso).Ptbcom}, {"PNBCOM", (*IFPTreso).Pnbcom},
		{"PTBSYN", (*IFPTreso).Ptbsyn}, {"PNBSYN", (*IFPTreso).Pnbsyn},
		{"PTBCU", (*IFPTreso).Ptbcu}, {"PNBCU", (*IFPTreso).Pnbcu},
		{"PNBTXA", (*IFPTreso).Pnbtxa}, {"PNBCAA", (*IFPTreso).Pnbcaa},
		{"PTBGEM", (*IFPTreso).Ptbgem}, {"PNBGEM", (*IFPTreso).Pnbgem},
		{"NTXOM", (*IFPTreso).Ntxom}, {"PBBOMP", (*IFPTreso).Pbbomp},
		{"PBBOMA", (*IFPTreso).Pbboma}, {"PBBOMB", (*IFPTreso).Pbbomb},
		{"PBBOMC", (*IFPTreso).Pbbomc}, {"PBBOMD", (*IFPTreso).Pbbomd},
		{"PBBOME", (*IFPTreso).Pbbome}, {"CCOBOM", (*IFPTreso).Ccobom},
	}
	for i, tc := range tests {
		t.Run(tc.field, func(t *testing.T) {
			r := mustNew(t)
			want := uint64(i%9 + 1) // single digit so NTXOM 9(1) can hold it
			if err := r.SetNum(tc.field, want); err != nil {
				t.Fatal(err)
			}
			if v, ok := tc.get(&r); v != want || !ok {
				t.Errorf("got %d,%v", v, ok)
			}
		})
	}
	r := mustNew(t)
	_ = r.SetNum("CCOBOM", 7)
	if r.CcobomRaw() != "07" {
		t.Errorf("CcobomRaw=%q", r.CcobomRaw())
	}
}

func TestTCTOMAlias(t *testing.T) {
	if fieldOf(t, "TCTOM") != fieldOf(t, "MCTOM1") {
		t.Fatal("TCTOM must share MCTOM1 storage")
	}
	r := mustNew(t)
	_ = r.SetNum("TCTOM", 42)
	if v, _, _ := r.Num("MCTOM1"); v != 42 {
		t.Errorf("MCTOM1=%d", v)
	}
	_ = r.SetNum("MCTOM1", 99)
	if v, _, _ := r.Num("TCTOM"); v != 99 {
		t.Errorf("TCTOM=%d", v)
	}
}

func TestSyndicateAliases(t *testing.T) {
	letters := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J"}
	prev := -1
	for i, l := range letters {
		t.Run(l, func(t *testing.T) {
			b := fieldOf(t, "MPBSY"+l)
			if b != fieldOf(t, occ("MPBSYN", i+1)) || b.n != 10 {
				t.Error("MPBSYN alias mismatch")
			}
			if prev >= 0 && b.off != prev+11 {
				t.Errorf("slot not 11 bytes: %d vs %d", b.off, prev)
			}
			prev = b.off
			if fieldOf(t, "MPNSY"+l) != fieldOf(t, occ("MPNSYN", i+1)) {
				t.Error("MPNSYN alias mismatch")
			}
			if fieldOf(t, "PCFSY"+l) != fieldOf(t, occ("PCFSYN", i+1)) || fieldOf(t, "PCFSY"+l).n != 5 {
				t.Error("PCFSYN alias mismatch")
			}
			r := mustNew(t)
			_ = r.SetNum("MPBSY"+l, uint64(100+i))
			_ = r.SetNum("MPNSY"+l, uint64(200+i))
			_ = r.SetRaw("PCFSY"+l, "10000")
			if v, ok, err := r.Mpbsyn(i); err != nil || !ok || v != uint64(100+i) {
				t.Errorf("Mpbsyn=%d,%v", v, err)
			}
			if v, ok, err := r.Mpnsyn(i); err != nil || !ok || v != uint64(200+i) {
				t.Errorf("Mpnsyn=%d,%v", v, err)
			}
			if v, ok, err := r.Pcfsyn(i); err != nil || !ok || v != 10000 {
				t.Errorf("Pcfsyn=%d,%v", v, err)
			}
		})
	}
}

func TestIndexErrors(t *testing.T) {
	r := mustNew(t)
	tests := []struct {
		name string
		call func() error
	}{
		{"Mpbsyn-1", func() error { _, _, e := r.Mpbsyn(-1); return e }},
		{"Mpbsyn10", func() error { _, _, e := r.Mpbsyn(IFPSyndCount); return e }},
		{"Mpnsyn10", func() error { _, _, e := r.Mpnsyn(10); return e }},
		{"Pcfsyn-1", func() error { _, _, e := r.Pcfsyn(-1); return e }},
		{"Ptbtsn2", func() error { _, _, e := r.Ptbtsn(IFPTsePairCount); return e }},
		{"Pnbtsn-1", func() error { _, _, e := r.Pnbtsn(-1); return e }},
		{"Jcoin0", func() error { _, _, e := r.Jcoin(0); return e }},
		{"Jcoin6", func() error { _, _, e := r.Jcoin(6); return e }},
		{"Jcori0", func() error { _, _, e := r.Jcori(0); return e }},
		{"Jcori6", func() error { _, _, e := r.Jcori(6); return e }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.call() == nil {
				t.Error("expected error")
			}
		})
	}
}

func TestTsePairs(t *testing.T) {
	r := mustNew(t)
	for i := 0; i < IFPTsePairCount; i++ {
		_ = r.SetNum(occ("PTBTSN", i+1), uint64(10+i))
		_ = r.SetNum(occ("PNBTSN", i+1), uint64(20+i))
	}
	for i := 0; i < IFPTsePairCount; i++ {
		if v, _, err := r.Ptbtsn(i); err != nil || v != uint64(10+i) {
			t.Errorf("Ptbtsn(%d)=%d", i, v)
		}
		if v, _, err := r.Pnbtsn(i); err != nil || v != uint64(20+i) {
			t.Errorf("Pnbtsn(%d)=%d", i, v)
		}
	}
	for _, base := range []string{"PTBTSN", "PNBTSN", "TBBTSN", "TBNTSN", "MCBTSN", "MCNTSN", "MPBTSN", "MPNTSN", "ACOETS", "DNUTSE"} {
		if _, ok := tresoFields[occ(base, 3)]; ok {
			t.Errorf("%s has more than 2 entries", base)
		}
		for i := 1; i <= 2; i++ {
			fieldOf(t, occ(base, i))
		}
	}
}

func TestJAAliases(t *testing.T) {
	for i := 1; i <= IFPJaCount; i++ {
		k := IFPJaCount + 1 - i
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			in := fieldOf(t, fmt.Sprintf("JCOIN%d", k))
			ri := fieldOf(t, fmt.Sprintf("JCORI%d", k))
			if fieldOf(t, occ("CINSN", i)) != in || fieldOf(t, occ("CRIMN", i)) != ri {
				t.Error("alias mismatch")
			}
			if ri.off != in.off+4 {
				t.Errorf("JCORI not adjacent: %d %d", in.off, ri.off)
			}
			if i > 1 && in.off != fieldOf(t, occ("CINSN", i-1)).off+9 {
				t.Error("occurrence not 9 bytes")
			}
			r := mustNew(t)
			_ = r.SetNum(occ("CINSN", i), uint64(2000+i))
			_ = r.SetNum(occ("CRIMN", i), uint64(2010+i))
			if v, ok, err := r.Jcoin(k); err != nil || !ok || v != uint64(2000+i) {
				t.Errorf("Jcoin=%d", v)
			}
			if v, ok, err := r.Jcori(k); err != nil || !ok || v != uint64(2010+i) {
				t.Errorf("Jcori=%d", v)
			}
		})
	}
	if _, ok := tresoFields[occ("CINSN", 6)]; ok {
		t.Error("ACJARO has more than 5 entries")
	}
}

func TestSignedFields(t *testing.T) {
	signed := []string{"MFSB41", "MFSN41", "MFSC41", "MFSB31", "MFSN31", "MFSC31",
		"MLBCOM", "MLBSYN", "MLBCU", "MLBDEP", "MLBTS1", "MLBTS2", "MLBTAS", "MLBGEM", "MLBTOM"}
	for _, l := range syndLetters {
		signed = append(signed, "MLBSY"+l)
	}
	set := map[string]bool{}
	for _, n := range signed {
		set[n] = true
		f := fieldOf(t, n)
		if f.kind != tresoSigned || f.n != 10 {
			t.Errorf("%s: %+v", n, f)
		}
	}
	for n, f := range tresoFields {
		if f.kind == tresoSigned && !set[n] {
			t.Errorf("unexpected signed field %s", n)
		}
	}
	for _, n := range []string{"MFEB41", "MFEN41", "MFEB31", "MFEN31"} {
		if fieldOf(t, n).kind != tresoUnsigned {
			t.Errorf("%s should be unsigned", n)
		}
	}
}

func TestSignedValues(t *testing.T) {
	tests := []struct {
		name   string
		set    *int64
		raw    string
		want   int64
		wantOK bool
		rawOut string
	}{
		{name: "neg", set: ptr(-123), want: -123, wantOK: true, rawOut: "000000012L"},
		{name: "negzero", set: ptr(-120), want: -120, wantOK: true, rawOut: "000000012}"},
		{name: "pos", set: ptr(123), want: 123, wantOK: true, rawOut: "000000012C"},
		{name: "poszero", set: ptr(0), want: 0, wantOK: true, rawOut: "000000000{"},
		{name: "plain", raw: "0000000123", want: 123, wantOK: true},
		{name: "gnu-neg", raw: "000000012s", want: -123, wantOK: true},
		{name: "bad-last", raw: "000000012 ", want: 120, wantOK: false},
		{name: "bad-body", raw: "000000A12C", want: 1123, wantOK: false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := mustNew(t)
			if tc.set != nil {
				if err := r.SetSigned("MLBCOM", *tc.set); err != nil {
					t.Fatal(err)
				}
			} else {
				_ = r.SetRaw("MLBCOM", tc.raw)
			}
			v, ok, err := r.Signed("MLBCOM")
			if err != nil || v != tc.want || ok != tc.wantOK {
				t.Errorf("got %d,%v,%v", v, ok, err)
			}
			if tc.rawOut != "" {
				if raw, _ := r.Raw("MLBCOM"); raw != tc.rawOut {
					t.Errorf("raw=%q", raw)
				}
			}
		})
	}
}

func ptr(v int64) *int64 { return &v }

func TestAccessorErrors(t *testing.T) {
	r := mustNew(t)
	tests := []struct {
		name string
		call func() error
	}{
		{"Raw unknown", func() error { _, e := r.Raw("NOPE"); return e }},
		{"SetRaw unknown", func() error { return r.SetRaw("NOPE", "x") }},
		{"Num unknown", func() error { _, _, e := r.Num("NOPE"); return e }},
		{"Num alpha", func() error { _, _, e := r.Num("DLETAR"); return e }},
		{"Num signed", func() error { _, _, e := r.Num("MLBCOM"); return e }},
		{"SetNum alpha", func() error { return r.SetNum("CSPER", 1) }},
		{"SetNum unknown", func() error { return r.SetNum("NOPE", 1) }},
		{"Signed unsigned", func() error { _, _, e := r.Signed("PTBCOM"); return e }},
		{"Signed unknown", func() error { _, _, e := r.Signed("NOPE"); return e }},
		{"SetSigned unsigned", func() error { return r.SetSigned("PTBCOM", 1) }},
		{"SetSigned unknown", func() error { return r.SetSigned("NOPE", 1) }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.call() == nil {
				t.Error("expected error")
			}
		})
	}
}

func TestCompositeFields(t *testing.T) {
	acosyn := fieldOf(t, "ACOSYN")
	cty := fieldOf(t, "CTYSYN")
	cco := fieldOf(t, "CCOSYN")
	if acosyn.off != cty.off || acosyn.n != 4 || cty.n != 1 || cco.off != cty.off+1 || cco.n != 3 {
		t.Errorf("ACOSYN layout wrong: %+v %+v %+v", acosyn, cty, cco)
	}
	ageref := fieldOf(t, "AGEREF")
	d1 := fieldOf(t, "DCPNU1")
	k := fieldOf(t, "KCPCLE")
	d9 := fieldOf(t, "DCPNU9")
	kn := fieldOf(t, "KNOCP")
	lib := fieldOf(t, "LIBRE")
	if ageref.off != d1.off || ageref.n != 9 || d1.n != 7 || k.off != d1.off+7 || k.n != 2 {
		t.Error("AGEREF layout wrong")
	}
	if d9.off != ageref.off+10 || d9.n != 7 || kn.off != d9.off+7 || kn.n != 2 || lib.off != kn.off+2 {
		t.Error("trailer layout wrong")
	}
	if lib.off+lib.n != BodyLen {
		t.Error("LIBRE does not end at BodyLen")
	}

	r := mustNew(t)
	_ = r.SetRaw("CTYSYN", "S")
	_ = r.SetRaw("CCOSYN", "123")
	if v, _ := r.Raw("ACOSYN"); v != "S123" {
		t.Errorf("ACOSYN=%q", v)
	}
	_ = r.SetNum("DCPNU1", 1234567)
	_ = r.SetRaw("KCPCLE", "AB")
	if v, _ := r.Raw("AGEREF"); v != "1234567AB" {
		t.Errorf("AGEREF=%q", v)
	}
	_ = r.SetRaw("LIBRE", "xyz")
	if r.Libre()[:4] != "xyz " {
		t.Errorf("Libre=%q", r.Libre()[:4])
	}
	_ = r.SetRaw("CCOSYN", "12345")
	if v, _ := r.Raw("CCOSYN"); v != "123" {
		t.Errorf("truncation=%q", v)
	}
}