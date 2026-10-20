package rates

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// ---------- fake layout ----------

const testSegLen = 400

type testLayout struct{ fields map[FieldID]FieldRef }

func newTestLayout() *testLayout {
	f := map[FieldID]FieldRef{}
	// XB40
	f[FXB40Ctygc] = FieldRef{0, 1, FieldAlnum}
	f[FXB40Dnucol] = FieldRef{1, 3, FieldAlnum}
	f[FXB40Ctysyn] = FieldRef{4, 1, FieldAlnum}
	f[FXB40Ccosyn] = FieldRef{5, 3, FieldAlnum}
	f[FXB40Gtovot] = FieldRef{8, 1, FieldAlnum}
	f[FXB40Gnidep] = FieldRef{9, 1, FieldAlnum}
	off := 10
	for id := FXB40Pbbomp; id <= FXB40Jcorim5; id++ {
		f[id] = FieldRef{off, 4, FieldZoned}
		off += 4
	}
	// XB46: same order as XB40 counterparts
	off = 0
	for id := FXB46Pbbomp; id <= FXB46Jcorim5; id++ {
		f[id] = FieldRef{off, 4, FieldZoned}
		off += 4
	}
	f[FXB35Ptbdep] = FieldRef{0, 4, FieldZoned}
	f[FXB35Adtseg] = FieldRef{10, 385, FieldAlnum}
	f[FXB35OriAdtseg] = FieldRef{0, 390, FieldAlnum}
	f[FXB45Ccddir] = FieldRef{0, 3, FieldAlnum}
	f[FXB45Acoets1] = FieldRef{3, 4, FieldAlnum}
	f[FXB45Acoets2] = FieldRef{7, 4, FieldAlnum}
	f[FXB37Vmailg] = FieldRef{0, 10, FieldAlnum}
	f[FXB37Vmaild] = FieldRef{10, 10, FieldAlnum}
	f[FXB50Aenr1] = FieldRef{0, 100, FieldAlnum}
	f[FXB50Aenr2] = FieldRef{100, 100, FieldAlnum}
	f[FXB50Aenr3] = FieldRef{200, 100, FieldAlnum}
	f[FXB50Vmailg] = FieldRef{300, 10, FieldAlnum}
	f[FXB50Vmaild] = FieldRef{310, 10, FieldAlnum}
	f[FXB41Vmailg] = FieldRef{0, 10, FieldAlnum}
	f[FXB41Vmaild] = FieldRef{10, 10, FieldAlnum}
	f[FXB51Vmailg] = FieldRef{0, 10, FieldAlnum}
	f[FXB51Vmaild] = FieldRef{10, 10, FieldAlnum}
	return &testLayout{fields: f}
}

func (l *testLayout) NewSegment(SegmentID) []byte {
	return []byte(strings.Repeat(" ", testSegLen))
}
func (l *testLayout) Field(id FieldID) FieldRef { return l.fields[id] }

// ---------- fake DL/I ----------

type resp struct {
	status string
	data   string
}

type fakeReader struct {
	responses map[string]resp
	calls     []string
	err       error
}

func (r *fakeReader) GU(_ context.Context, key [BaseTauxKeyLen]byte, seg []byte) (string, error) {
	k := string(key[:])
	r.calls = append(r.calls, k)
	if r.err != nil {
		return "", r.err
	}
	rp, ok := r.responses[k]
	if !ok {
		return "GE", nil
	}
	if rp.status == "  " {
		copy(seg, rp.data)
	}
	return rp.status, nil
}

type fakePCB struct {
	reader *fakeReader
	code   int
	err    error
	names  []string
}

func (p *fakePCB) PCB(_ context.Context, name string) (SegmentReader, int, error) {
	p.names = append(p.names, name)
	if p.err != nil || p.code != 0 {
		return nil, p.code, p.err
	}
	return p.reader, 0, nil
}

func ok(data string) resp { return resp{"  ", data} }

func commune(ctygc, dnucol, ctysyn, ccosyn, vot, nidep string) string {
	return ctygc + dnucol + ctysyn + ccosyn + vot + nidep
}

func happyResponses() map[string]resp {
	suite := ""
	for i := 1; i <= 22; i++ {
		suite += strings.Repeat(string(rune('0'+i%10)), 4)
	}
	return map[string]resp{
		"2018   000   ": ok("YEAR"),
		"2018380000   ": ok("DIRDATA"),
		"2018380185   ": ok(commune("G", "001", "S", "002", "V", "O")),
		"2018380G001  ": ok("EPCILABEL"),
		"2018380S002  ": ok("TEOMLABEL"),
		"2018380185123": ok("   T001T002"),
		"2018   185123": ok(suite),
		"2018380T001  ": ok("TSE1"),
		"2018380185Z  ": ok("SYND"),
		"2018380185Y  ": ok("MAILADDR01MAILADDR02"),
	}
}

func newSrc(t *testing.T, r *fakeReader) (*BaseTauxSource, *fakePCB) {
	t.Helper()
	p := &fakePCB{reader: r}
	s, err := NewBaseTauxSource(p, newTestLayout())
	if err != nil {
		t.Fatalf("NewBaseTauxSource: %v", err)
	}
	return s, p
}

func hasCall(calls []string, k string) bool {
	for _, c := range calls {
		if c == k {
			return true
		}
	}
	return false
}

// ---------- tests ----------

func TestNewBaseTauxSourceValidation(t *testing.T) {
	if _, err := NewBaseTauxSource(nil, newTestLayout()); err == nil {
		t.Error("expected error for nil PCB provider")
	}
	if _, err := NewBaseTauxSource(&fakePCB{}, nil); err == nil {
		t.Error("expected error for nil layout")
	}
	bad := newTestLayout()
	bad.fields[FXB35OriAdtseg] = FieldRef{0, 100, FieldAlnum}
	if _, err := NewBaseTauxSource(&fakePCB{}, bad); err == nil {
		t.Error("expected error for short ORI-ADTSEG")
	}
	bad2 := newTestLayout()
	bad2.fields[FXB40Ctygc] = FieldRef{399, 5, FieldAlnum}
	if _, err := NewBaseTauxSource(&fakePCB{}, bad2); err == nil {
		t.Error("expected error for field outside segment")
	}
}

func TestUnavailableBaseTaux(t *testing.T) {
	res, err := UnavailableBaseTaux{}.Fetch(context.Background(), BaseTauxRequest{})
	if err != nil || res.CR != 24 || res.RC != 1 || string(res.RCRaw[:]) != "01" {
		t.Fatalf("got %+v err %v", res, err)
	}
}

func TestPCBResolution(t *testing.T) {
	r := &fakeReader{responses: happyResponses()}
	p := &fakePCB{reader: r, code: 8}
	s, _ := NewBaseTauxSource(p, newTestLayout())
	res, err := s.Fetch(context.Background(), BaseTauxRequest{"2018", "380", "185", "123"})
	if err != nil {
		t.Fatal(err)
	}
	if res.CR != CRPCBUnavailable || res.RC != 8 {
		t.Errorf("CR/RC = %d/%d", res.CR, res.RC)
	}
	if len(r.calls) != 0 {
		t.Errorf("unexpected GU calls %v", r.calls)
	}
	if len(p.names) != 1 || p.names[0] != BaseTauxPCBName {
		t.Errorf("PCB names %v", p.names)
	}

	p2 := &fakePCB{err: errors.New("boom")}
	s2, _ := NewBaseTauxSource(p2, newTestLayout())
	if _, err := s2.Fetch(context.Background(), BaseTauxRequest{"2018", "380", "", ""}); err == nil {
		t.Error("expected error from PCB failure")
	}
}

func TestDLIInfrastructureError(t *testing.T) {
	r := &fakeReader{err: errors.New("ims down")}
	s, _ := newSrc(t, r)
	if _, err := s.Fetch(context.Background(), BaseTauxRequest{"2018", "380", "185", "123"}); err == nil {
		t.Error("expected error")
	}
}

func TestYearZero(t *testing.T) {
	r := &fakeReader{responses: happyResponses()}
	s, _ := newSrc(t, r)
	res, err := s.Fetch(context.Background(), BaseTauxRequest{"0000", "380", "185", "123"})
	if err != nil {
		t.Fatal(err)
	}
	if res.CR != 0 || res.RC != 0 || len(r.calls) != 0 {
		t.Fatalf("CR %d RC %d calls %v", res.CR, res.RC, r.calls)
	}
	for i := range res.ZES {
		if !isSpaces(string(res.ZES[i][:])) {
			t.Errorf("slot %d not spaces", i)
		}
	}
}

func TestHappyPath(t *testing.T) {
	r := &fakeReader{responses: happyResponses()}
	s, _ := newSrc(t, r)
	req := BaseTauxRequest{"2018", "380", "185", "123"}
	res, err := s.Fetch(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if res.CR != 0 || res.RC != 0 {
		t.Fatalf("CR/RC %d/%d", res.CR, res.RC)
	}
	z := func(i, a, b int) string { return string(res.ZES[i][a:b]) }
	checks := []struct{ name, got, want string }{
		{"year slot", z(0, 0, 4), "YEAR"},
		{"dir slot", z(1, 0, 7), "DIRDATA"},
		{"commune vot", z(2, 8, 9), "V"},
		{"suite pbbomp copied", z(2, 10, 14), "1111"},
		{"suite jcoins1 copied", z(2, 58, 62), "3333"},
		{"ifp slot", z(3, 0, 11), "   T001T002"},
		{"xb50 aenr1", z(4, 0, 4), "TSE1"},
		{"xb50 aenr2 spaces", strings.TrimSpace(z(4, 100, 200)), ""},
		{"xb50 aenr3", z(4, 200, 209), "EPCILABEL"},
		{"xb50 vmailg", z(4, 300, 310), "EPCILABEL "},
		{"xb51 mail", z(5, 0, 20), "MAILADDR01MAILADDR02"},
		{"synd slot", z(6, 0, 4), "SYND"},
	}
	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("%s: got %q want %q", c.name, c.got, c.want)
		}
	}
	for _, k := range []string{"2018   000   ", "2018380000   ", "2018380185   ", "2018380185123", "2018   185123", "2018380T002  "} {
		if !hasCall(r.calls, k) {
			t.Errorf("missing GU %q", k)
		}
	}
	if req.Acodir != "380" {
		t.Error("request mutated")
	}
}

func TestTeomCopiedToXB51(t *testing.T) {
	m := happyResponses()
	delete(m, "2018380185Y  ") // mail GE: XB51 keeps TEOM label
	r := &fakeReader{responses: m}
	s, _ := newSrc(t, r)
	res, _ := s.Fetch(context.Background(), BaseTauxRequest{"2018", "380", "185", "   "})
	if got := string(res.ZES[5][0:9]); got != "TEOMLABEL" {
		t.Errorf("xb51 = %q", got)
	}
	if res.CR != 0 {
		t.Errorf("GE on mail must be ignored, CR=%d", res.CR)
	}
}

func TestErrorCases(t *testing.T) {
	tests := []struct {
		name   string
		req    BaseTauxRequest
		modify func(m map[string]resp)
		cr     int
		rcRaw  string
	}{
		{"year GE", BaseTauxRequest{"2018", "", "", ""}, func(m map[string]resp) { delete(m, "2018   000   ") }, 20, "01"},
		{"year DLI", BaseTauxRequest{"2018", "", "", ""}, func(m map[string]resp) { m["2018   000   "] = resp{"AI", ""} }, 16, "AI"},
		{"dir GE", BaseTauxRequest{"2018", "380", "", ""}, func(m map[string]resp) { delete(m, "2018380000   ") }, 21, "01"},
		{"dir DLI", BaseTauxRequest{"2018", "380", "", ""}, func(m map[string]resp) { m["2018380000   "] = resp{"GB", ""} }, 16, "GB"},
		{"com GE", BaseTauxRequest{"2018", "380", "185", ""}, func(m map[string]resp) { delete(m, "2018380185   ") }, 22, "01"},
		{"com DLI", BaseTauxRequest{"2018", "380", "185", ""}, func(m map[string]resp) { m["2018380185   "] = resp{"AI", ""} }, 16, "AI"},
		{"com not voted", BaseTauxRequest{"2018", "380", "185", "123"}, func(m map[string]resp) {
			m["2018380185   "] = ok(commune("G", "001", "S", "002", "X", "O"))
		}, 12, "10"},
		{"epci DLI", BaseTauxRequest{"2018", "380", "185", ""}, func(m map[string]resp) { m["2018380G001  "] = resp{"AI", ""} }, 16, "AI"},
		{"epci GE tolerated", BaseTauxRequest{"2018", "380", "185", ""}, func(m map[string]resp) { delete(m, "2018380G001  ") }, 0, "00"},
		{"teom DLI", BaseTauxRequest{"2018", "380", "185", ""}, func(m map[string]resp) { m["2018380S002  "] = resp{"AI", ""} }, 16, "AI"},
		{"ifp GE", BaseTauxRequest{"2018", "380", "185", "123"}, func(m map[string]resp) { delete(m, "2018380185123") }, 23, "01"},
		{"ifp DLI", BaseTauxRequest{"2018", "380", "185", "123"}, func(m map[string]resp) { m["2018380185123"] = resp{"AI", ""} }, 16, "AI"},
		{"suite GE", BaseTauxRequest{"2018", "380", "185", "123"}, func(m map[string]resp) { delete(m, "2018   185123") }, 24, "01"},
		{"ori GE", BaseTauxRequest{"2018", "380", "185", "123"}, func(m map[string]resp) { m["2018380185123"] = ok("690T001T002") }, 25, "01"},
		{"tse DLI", BaseTauxRequest{"2018", "380", "185", "123"}, func(m map[string]resp) { m["2018380T001  "] = resp{"AI", ""} }, 16, "AI"},
		{"synd DLI", BaseTauxRequest{"2018", "380", "185", ""}, func(m map[string]resp) { m["2018380185Z  "] = resp{"AI", ""} }, 16, "AI"},
		{"mail DLI", BaseTauxRequest{"2018", "380", "185", ""}, func(m map[string]resp) { m["2018380185Y  "] = resp{"AI", ""} }, 16, "AI"},
		{"synd GE ignored", BaseTauxRequest{"2018", "380", "185", ""}, func(m map[string]resp) { delete(m, "2018380185Z  ") }, 0, "00"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := happyResponses()
			tc.modify(m)
			r := &fakeReader{responses: m}
			s, _ := newSrc(t, r)
			res, err := s.Fetch(context.Background(), tc.req)
			if err != nil {
				t.Fatal(err)
			}
			if res.CR != tc.cr || string(res.RCRaw[:]) != tc.rcRaw {
				t.Errorf("CR %d RC %q, want %d %q", res.CR, res.RCRaw[:], tc.cr, tc.rcRaw)
			}
		})
	}
}

func TestNotVotedSkipsLaterSteps(t *testing.T) {
	m := happyResponses()
	m["2018380185   "] = ok(commune("G", "001", "S", "002", "X", "O"))
	r := &fakeReader{responses: m}
	s, _ := newSrc(t, r)
	s.Fetch(context.Background(), BaseTauxRequest{"2018", "380", "185", "123"})
	for _, k := range []string{"2018380G001  ", "2018380185123", "2018380185Z  ", "2018380185Y  "} {
		if hasCall(r.calls, k) {
			t.Errorf("unexpected GU %q", k)
		}
	}
}

func TestGnidepZeroesPtbdep(t *testing.T) {
	m := happyResponses()
	m["2018380185   "] = ok(commune("G", "001", "S", "002", "V", "N"))
	r := &fakeReader{responses: m}
	s, _ := newSrc(t, r)
	res, _ := s.Fetch(context.Background(), BaseTauxRequest{"2018", "380", "185", ""})
	if got := string(res.ZES[1][0:4]); got != "0000" {
		t.Errorf("PTBDEP = %q", got)
	}
}

func TestOriginDirection(t *testing.T) {
	m := happyResponses()
	m["2018380185   "] = ok(commune("G", "001", "S", "002", "V", "N"))
	m["2018380185123"] = ok("690T001T002")
	m["2018690000   "] = ok("ORIGINDATA")
	r := &fakeReader{responses: m}
	s, _ := newSrc(t, r)
	res, err := s.Fetch(context.Background(), BaseTauxRequest{"2018", "380", "185", "123"})
	if err != nil {
		t.Fatal(err)
	}
	if res.CR != 0 {
		t.Fatalf("CR %d", res.CR)
	}
	if got := string(res.ZES[1][10:20]); got != "ORIGINDATA" {
		t.Errorf("adtseg = %q", got)
	}
	if got := string(res.ZES[1][0:4]); got != "0000" {
		t.Errorf("PTBDEP = %q", got)
	}
}

func TestYear2015SkipsSuite(t *testing.T) {
	m := map[string]resp{
		"2015   000   ": ok("Y"),
		"2015380000   ": ok("D"),
		"2015380185   ": ok(commune(" ", "   ", " ", "   ", "V", "O")),
		"2015380185123": ok("   T001T002"),
	}
	r := &fakeReader{responses: m}
	s, _ := newSrc(t, r)
	res, _ := s.Fetch(context.Background(), BaseTauxRequest{"2015", "380", "185", "123"})
	if hasCall(r.calls, "2015   185123") {
		t.Error("suite read for year <= 2015")
	}
	if res.CR != 0 {
		t.Errorf("CR %d", res.CR)
	}
}

func TestPELP(t *testing.T) {
	tests := []struct {
		name     string
		dir      string
		probes   map[string]resp
		wantDir  string
		wantCR   int
		wantCall []string
		noCall   []string
	}{
		{"dept13 found at 2", "130", map[string]resp{"2018132185   ": ok("x")}, "132", 0,
			[]string{"2018131185   ", "2018132185   ", "2018132000   "}, nil},
		{"dept13 all GE keeps dir", "130", nil, "130", 0,
			[]string{"2018131185   ", "2018132185   ", "2018130000   "}, nil},
		{"dept75 found at 5 epci uses 754", "750", map[string]resp{"2018755185   ": ok("x")}, "755", 0,
			[]string{"2018754185   ", "2018755185   ", "2018754G001  "}, []string{"2018756185   "}},
		{"dept59 DLI error stops", "590", map[string]resp{"2018591185   ": {"AI", ""}}, "", 16,
			[]string{"2018591185   "}, []string{"2018592185   ", "2018   000   "}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := map[string]resp{}
			for k, v := range tc.probes {
				m[k] = v
			}
			if tc.wantDir != "" {
				d := tc.wantDir
				m["2018   000   "] = ok("Y")
				m["2018"+d+"000   "] = ok("D")
				m["2018"+d+"185   "] = ok(commune("G", "001", " ", "   ", "V", "O"))
			}
			r := &fakeReader{responses: m}
			s, _ := newSrc(t, r)
			req := BaseTauxRequest{"2018", tc.dir, "185", ""}
			res, err := s.Fetch(context.Background(), req)
			if err != nil {
				t.Fatal(err)
			}
			if res.CR != tc.wantCR {
				t.Errorf("CR %d want %d (calls %v)", res.CR, tc.wantCR, r.calls)
			}
			for _, k := range tc.wantCall {
				if !hasCall(r.calls, k) {
					t.Errorf("missing GU %q in %v", k, r.calls)
				}
			}
			for _, k := range tc.noCall {
				if hasCall(r.calls, k) {
					t.Errorf("unexpected GU %q", k)
				}
			}
			if req.Acodir != tc.dir {
				t.Error("ACODIR not preserved")
			}
		})
	}
}