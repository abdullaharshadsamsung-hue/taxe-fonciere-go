package rates

import (
	"bytes"
	"context"
	"testing"

	"migrated-app/internal/layout/taudis"
)

// ---------- mocks ----------

const recLen = 3000

type mockStore struct {
	recs  map[[taudis.EnteteLen]byte][]byte
	keys  [][taudis.EnteteLen]byte
	err   error
	errOn int // 1-based read index that fails; 0 = never
}

func (m *mockStore) Read(_ context.Context, key [taudis.EnteteLen]byte) ([]byte, bool, error) {
	m.keys = append(m.keys, key)
	if m.err != nil && (m.errOn == 0 || m.errOn == len(m.keys)) {
		return nil, false, m.err
	}
	r, ok := m.recs[key]
	if !ok {
		return nil, false, nil
	}
	out := make([]byte, len(r))
	copy(out, r)
	return out, true, nil
}

type mockIFP struct {
	ccobom string
}

func (m *mockIFP) Ptbcom() (uint64, bool) { return 1, true }
func (m *mockIFP) Pnbcom() (uint64, bool) { return 2, true }
func (m *mockIFP) Ptbsyn() (uint64, bool) { return 3, true }
func (m *mockIFP) Pnbsyn() (uint64, bool) { return 4, true }
func (m *mockIFP) Ptbcu() (uint64, bool)  { return 5, true }
func (m *mockIFP) Pnbcu() (uint64, bool)  { return 6, true }
func (m *mockIFP) Pnbtxa() (uint64, bool) { return 7, true }
func (m *mockIFP) Ptbtsn(i int) (uint64, bool, error) {
	return uint64(100 + i), true, nil
}
func (m *mockIFP) Pnbtsn(i int) (uint64, bool, error) {
	return uint64(200 + i), true, nil
}
func (m *mockIFP) Pnbcaa() (uint64, bool) { return 8, true }
func (m *mockIFP) Ptbgem() (uint64, bool) { return 9, true }
func (m *mockIFP) Pnbgem() (uint64, bool) { return 10, true }
func (m *mockIFP) Jcoin(n int) (uint64, bool, error) {
	return uint64(n * 10), true, nil
}
func (m *mockIFP) Jcori(n int) (uint64, bool, error) {
	return uint64(n * 100), true, nil
}
func (m *mockIFP) Ntxom() (uint64, bool)  { return 11, true }
func (m *mockIFP) Pbbomp() (uint64, bool) { return 12, true }
func (m *mockIFP) Pbboma() (uint64, bool) { return 13, true }
func (m *mockIFP) Pbbomb() (uint64, bool) { return 14, true }
func (m *mockIFP) Pbbomc() (uint64, bool) { return 15, true }
func (m *mockIFP) Pbbomd() (uint64, bool) { return 16, true }
func (m *mockIFP) Pbbome() (uint64, bool) { return 17, true }
func (m *mockIFP) Ccobom() string         { return m.ccobom }

type mockCOM struct {
	aidfic, depdir, ccocom, dnultf, jdhomo, gcoanu string
	jan, dnuro, nbtauxom                           uint64
	pbbom                                          [6]uint64
	jgcins, jgcrim, jcoins, jcorim                 map[int]uint64
	filler                                         byte
}

func newMockCOM() *mockCOM {
	return &mockCOM{
		jgcins: map[int]uint64{}, jgcrim: map[int]uint64{},
		jcoins: map[int]uint64{}, jcorim: map[int]uint64{}, filler: 'C',
	}
}
func (m *mockCOM) SetAidfic(v string)              { m.aidfic = v }
func (m *mockCOM) SetJan(v uint64)                 { m.jan = v }
func (m *mockCOM) SetDepdir(v string)              { m.depdir = v }
func (m *mockCOM) SetCcocom(v string)              { m.ccocom = v }
func (m *mockCOM) SetJgcins(i int, v uint64) error { m.jgcins[i] = v; return nil }
func (m *mockCOM) SetJgcrim(i int, v uint64) error { m.jgcrim[i] = v; return nil }
func (m *mockCOM) SetDnuro(v uint64)               { m.dnuro = v }
func (m *mockCOM) SetDnultf(v string)              { m.dnultf = v }
func (m *mockCOM) SetJdhomo(v string)              { m.jdhomo = v }
func (m *mockCOM) SetGcoanu(v string)              { m.gcoanu = v }
func (m *mockCOM) SetJcoins(i int, v uint64) error { m.jcoins[i] = v; return nil }
func (m *mockCOM) SetJcorim(i int, v uint64) error { m.jcorim[i] = v; return nil }
func (m *mockCOM) SetNbtauxom(v uint64)            { m.nbtauxom = v }
func (m *mockCOM) SetPbbomp(v uint64)              { m.pbbom[0] = v }
func (m *mockCOM) SetPbboma(v uint64)              { m.pbbom[1] = v }
func (m *mockCOM) SetPbbomb(v uint64)              { m.pbbom[2] = v }
func (m *mockCOM) SetPbbomc(v uint64)              { m.pbbom[3] = v }
func (m *mockCOM) SetPbbomd(v uint64)              { m.pbbom[4] = v }
func (m *mockCOM) SetPbbome(v uint64)              { m.pbbom[5] = v }
func (m *mockCOM) Bytes() []byte                   { return bytes.Repeat([]byte{m.filler}, ZESSlotLen) }

type mockSR struct {
	aidfic, depdir, ccocom, ccoifp, ligbenefom string
	jan                                        uint64
	vals                                       map[string]uint64
	length                                     int
}

func newMockSR() *mockSR { return &mockSR{vals: map[string]uint64{}, length: ZESSlotLen} }
func (m *mockSR) SetAidfic(v string)       { m.aidfic = v }
func (m *mockSR) SetJan(v uint64)          { m.jan = v }
func (m *mockSR) SetDepdir(v string)       { m.depdir = v }
func (m *mockSR) SetCcocom(v string)       { m.ccocom = v }
func (m *mockSR) SetCcoifp(v string)       { m.ccoifp = v }
func (m *mockSR) SetTaucomB(v uint64)      { m.vals["TaucomB"] = v }
func (m *mockSR) SetTaucomNB(v uint64)     { m.vals["TaucomNB"] = v }
func (m *mockSR) SetTausyndsfpB(v uint64)  { m.vals["TausyndsfpB"] = v }
func (m *mockSR) SetTausyndsfpNB(v uint64) { m.vals["TausyndsfpNB"] = v }
func (m *mockSR) SetTaucudfpvnB(v uint64)  { m.vals["TaucudfpvnB"] = v }
func (m *mockSR) SetTaucudfpvnNB(v uint64) { m.vals["TaucudfpvnNB"] = v }
func (m *mockSR) SetPnbtxa(v uint64)       { m.vals["Pnbtxa"] = v }
func (m *mockSR) SetTautseB(v uint64)      { m.vals["TautseB"] = v }
func (m *mockSR) SetTautseNB(v uint64)     { m.vals["TautseNB"] = v }
func (m *mockSR) SetPtbtgp(v uint64)       { m.vals["Ptbtgp"] = v }
func (m *mockSR) SetPnbtgp(v uint64)       { m.vals["Pnbtgp"] = v }
func (m *mockSR) SetPnbcaa(v uint64)       { m.vals["Pnbcaa"] = v }
func (m *mockSR) SetPtbgem(v uint64)       { m.vals["Ptbgem"] = v }
func (m *mockSR) SetPnbgem(v uint64)       { m.vals["Pnbgem"] = v }
func (m *mockSR) SetLigbenefom(v string)   { m.ligbenefom = v }
func (m *mockSR) Bytes() []byte            { return bytes.Repeat([]byte{'S'}, m.length) }

// ---------- helpers ----------

type fixture struct {
	src   *TaudisSource
	store *mockStore
	com   *mockCOM
	sr    *mockSR
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	f := &fixture{
		store: &mockStore{recs: map[[taudis.EnteteLen]byte][]byte{}},
		com:   newMockCOM(),
		sr:    newMockSR(),
	}
	l := Layouts{
		ParseIFP: func(body []byte) (IFPRecord, error) { return &mockIFP{ccobom: "22"}, nil },
		NewCOM:   func() COMSegment { return f.com },
		NewSR:    func() SRSegment { return f.sr },
		NewIFP:   func() IFPRecord { return &mockIFP{ccobom: "34"} },
	}
	src, err := NewTaudisSource(f.store, l)
	if err != nil {
		t.Fatalf("NewTaudisSource: %v", err)
	}
	f.src = src
	return f
}

func spaceRecord(key [taudis.EnteteLen]byte) []byte {
	r := bytes.Repeat([]byte{' '}, recLen)
	copy(r, key[:])
	return r
}

func bodyOf(rec []byte) []byte { return rec[taudis.EnteteLen:] }

var (
	keyDir = buildKey("751", "", "", "", 1)
	keyCom = buildKey("751", "056", "", "", 2)
	keyIFP = buildKey("751", "056", "001", "002", 4)
)

// requireParsable skips if a spaces record cannot be split/parsed by the layouts.
func requireParsable(t *testing.T) {
	t.Helper()
	r := spaceRecord(keyDir)
	if _, _, err := taudis.SplitRecord(r); err != nil {
		t.Skipf("spaces record not splittable: %v", err)
	}
	if _, err := taudis.ParseDirection(bodyOf(r)); err != nil {
		t.Skipf("spaces direction not parsable: %v", err)
	}
	if _, err := taudis.ParseCommune(bodyOf(r)); err != nil {
		t.Skipf("spaces commune not parsable: %v", err)
	}
}

func allSpaces(b []byte) bool {
	for _, c := range b {
		if c != ' ' {
			return false
		}
	}
	return true
}

func zesUnfilled(z ZES) bool {
	for i := range z {
		if !allSpaces(z[i][:]) {
			return false
		}
	}
	return true
}

var fullReq = Request{Animp: "2018", Coddir: "751", Codcom: "056", Ccoifp: "001", Ccpper: "002"}

// ---------- tests ----------

func TestNewTaudisSource_Validation(t *testing.T) {
	ok := Layouts{
		ParseIFP: func([]byte) (IFPRecord, error) { return &mockIFP{}, nil },
		NewCOM:   func() COMSegment { return newMockCOM() },
		NewSR:    func() SRSegment { return newMockSR() },
		NewIFP:   func() IFPRecord { return &mockIFP{} },
	}
	tests := []struct {
		name    string
		store   Store
		mod     func(*Layouts)
		wantErr bool
	}{
		{"ok", &mockStore{}, func(*Layouts) {}, false},
		{"nil store", nil, func(*Layouts) {}, true},
		{"nil ParseIFP", &mockStore{}, func(l *Layouts) { l.ParseIFP = nil }, true},
		{"nil NewCOM", &mockStore{}, func(l *Layouts) { l.NewCOM = nil }, true},
		{"nil NewSR", &mockStore{}, func(l *Layouts) { l.NewSR = nil }, true},
		{"nil NewIFP", &mockStore{}, func(l *Layouts) { l.NewIFP = nil }, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			l := ok
			tc.mod(&l)
			s, err := NewTaudisSource(tc.store, l)
			if (err != nil) != tc.wantErr {
				t.Fatalf("err=%v wantErr=%v", err, tc.wantErr)
			}
			if !tc.wantErr && s == nil {
				t.Fatal("nil source")
			}
		})
	}
}

func TestFetch_Routing(t *testing.T) {
	tests := []struct {
		name     string
		req      Request
		wantKeys [][taudis.EnteteLen]byte
		wantCR   int
		wantRC   int
		filled   bool
	}{
		{"full, commune not validated", fullReq, [][taudis.EnteteLen]byte{keyDir, keyCom}, CRNotValidated, RCNotValidated, false},
		{"dir+com, no aiftres", Request{Animp: "2018", Coddir: "751", Codcom: "056", Ccoifp: "   ", Ccpper: "   "},
			[][taudis.EnteteLen]byte{keyDir, keyCom}, 0, 0, true},
		{"dir only", Request{Animp: "2018", Coddir: "751", Codcom: "000"},
			[][taudis.EnteteLen]byte{keyDir}, 0, 0, true},
		{"animp zero", Request{Animp: "0000", Coddir: "751", Codcom: "056", Ccoifp: "001", Ccpper: "002"}, nil, 0, 0, true},
		{"coddir spaces", Request{Animp: "2018", Coddir: "   ", Codcom: "056", Ccoifp: "001"}, nil, 0, 0, true},
		{"codcom zero with aiftres", Request{Animp: "2018", Coddir: "751", Codcom: "000", Ccoifp: "001", Ccpper: "002"}, nil, 0, 0, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			requireParsable(t)
			f := newFixture(t)
			f.store.recs[keyDir] = spaceRecord(keyDir)
			f.store.recs[keyCom] = spaceRecord(keyCom)
			f.store.recs[keyIFP] = spaceRecord(keyIFP)
			res, err := f.src.Fetch(context.Background(), tc.req)
			if err != nil {
				t.Fatalf("Fetch: %v", err)
			}
			if res.CR != tc.wantCR || res.RC != tc.wantRC {
				t.Fatalf("CR/RC=%d/%d want %d/%d", res.CR, res.RC, tc.wantCR, tc.wantRC)
			}
			if len(f.store.keys) != len(tc.wantKeys) {
				t.Fatalf("reads=%d want %d", len(f.store.keys), len(tc.wantKeys))
			}
			for i := range tc.wantKeys {
				if f.store.keys[i] != tc.wantKeys[i] {
					t.Errorf("read %d key=%q want %q", i, f.store.keys[i][:], tc.wantKeys[i][:])
				}
			}
			if tc.filled == zesUnfilled(res.ZES) {
				t.Errorf("filled=%v but ZES unfilled=%v", tc.filled, zesUnfilled(res.ZES))
			}
			for i := 4; i < ZESSlotCount; i++ {
				if !allSpaces(res.ZES[i][:]) {
					t.Errorf("EL-REC(%d) not spaces", i+1)
				}
			}
		})
	}
}

func TestBuildKey(t *testing.T) {
	tests := []struct {
		name string
		used int
		want string
	}{
		{"dir", 1, "751\x00\x00\x00\x00\x00\x00\x00\x00\x00"},
		{"com", 2, "751056\x00\x00\x00\x00\x00\x00"},
		{"ifp", 4, "751056001002"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			k := buildKey("751", "056", "001", "002", tc.used)
			if string(k[:]) != tc.want {
				t.Errorf("got %q want %q", k[:], tc.want)
			}
		})
	}
}