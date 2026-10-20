// Package rates holds the rate sources that feed the taxe foncière 2018
// calculators with taxation parameters.
package rates

import (
	"context"
	"fmt"
	"sync"

	"migrated-app/internal/layout/taudis"
	"migrated-app/internal/layout/zes"
)

// ZESSlotLen is the length of one EL-REC slot of the ZES return area.
const ZESSlotLen = 400

// ZESSlotCount is the OCCURS count of EL-REC in the ZES return area.
const ZESSlotCount = 7

// ZES is the EFITAUX2 return area (01 ZES. 05 EL-REC PIC X(400) OCCURS 7).
type ZES [ZESSlotCount][ZESSlotLen]byte

// newZESSpaces returns a ZES area as left by INITIALIZE ZES (all spaces).
func newZESSpaces() ZES {
	var z ZES
	for i := range z {
		for j := range z[i] {
			z[i][j] = ' '
		}
	}
	return z
}

// Return codes reported by EFITAUX2 through CR / RC.
const (
	// CRNotValidated / RCNotValidated: the commune was not validated by the
	// rate-control chain (GCOVEF <> 'V').
	CRNotValidated = 12
	RCNotValidated = 10
	// CRDirNotFound: the Direction article could not be read.
	CRDirNotFound = 20
	// CRComNotFound: the Commune article could not be read.
	CRComNotFound = 21
	// CRIFPNotFound: the IFP/Trésorerie article could not be read.
	CRIFPNotFound = 23
	// RCReadFailed is the RC that accompanies the read failures above.
	RCReadFailed = 1
)

// Store is the keyed access to the TAUDIS file (former VSAM KSDS). Read
// returns the 3000-byte record whose 12-byte primary key equals key.
// found=false is the equivalent of a FILE STATUS other than '00'; a non-nil
// error is reserved for infrastructure failures (COBOL abend conditions).
type Store interface {
	Read(ctx context.Context, key [taudis.EnteteLen]byte) (rec []byte, found bool, err error)
}

// IFPRecord is the read view of a TAUDIS IFP/Trésorerie body (copybook
// T84R) needed by EFITAUX2. Numeric getters return the zoned-decimal digit
// value (implied decimals included) and the NUMERIC class test result.
//
// MIGRATION_NOTE: declared here as an interface so that this file does not
// depend on accessor names of internal/layout/taudis/t84r.go that were not
// verified at migration time; the T84R type (or a thin adapter) must
// satisfy it. Indexed getters are 0-based (COBOL subscript i+1).
type IFPRecord interface {
	Ptbcom() (uint64, bool)
	Pnbcom() (uint64, bool)
	Ptbsyn() (uint64, bool)
	Pnbsyn() (uint64, bool)
	Ptbcu() (uint64, bool)
	Pnbcu() (uint64, bool)
	Pnbtxa() (uint64, bool)
	Ptbtsn(i int) (uint64, bool, error)
	Pnbtsn(i int) (uint64, bool, error)
	Pnbcaa() (uint64, bool)
	Ptbgem() (uint64, bool)
	Pnbgem() (uint64, bool)
	// Jcoin returns JCOINn for n in 1..5 (1 = year N, 5 = year N-4).
	Jcoin(n int) (uint64, bool, error)
	// Jcori returns JCORIn for n in 1..5 (1 = year N, 5 = year N-4).
	Jcori(n int) (uint64, bool, error)
	Ntxom() (uint64, bool)
	Pbbomp() (uint64, bool)
	Pbboma() (uint64, bool)
	Pbbomb() (uint64, bool)
	Pbbomc() (uint64, bool)
	Pbbomd() (uint64, bool)
	Pbbome() (uint64, bool)
	Ccobom() string
}

// COMSegment is the write view of the ZES COM segment (copybook XBXTDCOM)
// filled by EFITAUX2. Indexed setters are 0-based.
//
// MIGRATION_NOTE: see IFPRecord; zes.COM (or an adapter) must satisfy it.
type COMSegment interface {
	SetAidfic(v string)
	SetJan(v uint64)
	SetDepdir(v string)
	SetCcocom(v string)
	SetJgcins(i int, v uint64) error
	SetJgcrim(i int, v uint64) error
	SetDnuro(v uint64)
	SetDnultf(v string)
	SetJdhomo(v string)
	SetGcoanu(v string)
	SetJcoins(i int, v uint64) error
	SetJcorim(i int, v uint64) error
	SetNbtauxom(v uint64)
	SetPbbomp(v uint64)
	SetPbboma(v uint64)
	SetPbbomb(v uint64)
	SetPbbomc(v uint64)
	SetPbbomd(v uint64)
	SetPbbome(v uint64)
	Bytes() []byte
}

// SRSegment is the write view of the ZES SR segment (copybook XBXTDSR).
//
// MIGRATION_NOTE: see IFPRecord; zes.SR (or an adapter) must satisfy it.
type SRSegment interface {
	SetAidfic(v string)
	SetJan(v uint64)
	SetDepdir(v string)
	SetCcocom(v string)
	SetCcoifp(v string)
	SetTaucomB(v uint64)
	SetTaucomNB(v uint64)
	SetTausyndsfpB(v uint64)
	SetTausyndsfpNB(v uint64)
	SetTaucudfpvnB(v uint64)
	SetTaucudfpvnNB(v uint64)
	SetPnbtxa(v uint64)
	SetTautseB(v uint64)
	SetTautseNB(v uint64)
	SetPtbtgp(v uint64)
	SetPnbtgp(v uint64)
	SetPnbcaa(v uint64)
	SetPtbgem(v uint64)
	SetPnbgem(v uint64)
	SetLigbenefom(v string)
	Bytes() []byte
}

// Layouts bundles the factories for the layouts EFITAUX2 needs but that are
// injected rather than referenced directly.
type Layouts struct {
	// ParseIFP decodes a 2988-byte T84R body.
	ParseIFP func(body []byte) (IFPRecord, error)
	// NewCOM returns a fresh COM segment (initial WORKING-STORAGE content).
	NewCOM func() COMSegment
	// NewSR returns a fresh SR segment (initial WORKING-STORAGE content).
	NewSR func() SRSegment
	// NewIFP returns the initial (never read) IFP record.
	NewIFP func() IFPRecord
}

// Request carries the EFITAUX2 linkage parameters as raw COBOL bytes so
// that the presence tests behave exactly like the original EVALUATE.
type Request struct {
	Animp  string // ANIMP  PIC 9(4)
	Coddir string // CODDIR PIC X(3)
	Codcom string // CODCOM PIC 9(3)
	Ccoifp string // CCOIFP PIC X(3)
	Ccpper string // CCPPER PIC X(3)
}

// Result is the EFITAUX2 output: CR / RC and the ZES area.
type Result struct {
	CR  int
	RC  int
	ZES ZES
}

// TaudisSource is the Go port of EFITAUX2. It is stateful: like the
// non-INITIAL COBOL program, the last records read and the return segments
// persist between calls (a call that skips a read reuses stale data).
// It is safe for concurrent use; calls are serialised.
type TaudisSource struct {
	mu      sync.Mutex
	store   Store
	layouts Layouts

	dirHdr taudis.Entete
	dir    taudis.Direction
	comHdr taudis.Entete
	com    taudis.Commune
	rHdr   taudis.Entete
	ifp    IFPRecord

	segAN  zes.AN
	segDIR zes.DIR
	segCOM COMSegment
	segSR  SRSegment
}

// NewTaudisSource builds the EFITAUX2 rate source over store.
func NewTaudisSource(store Store, l Layouts) (*TaudisSource, error) {
	if store == nil {
		return nil, fmt.Errorf("rates: nil TAUDIS store")
	}
	if l.ParseIFP == nil || l.NewCOM == nil || l.NewSR == nil || l.NewIFP == nil {
		return nil, fmt.Errorf("rates: incomplete layout factories")
	}
	// MIGRATION_NOTE: the COBOL WORKING-STORAGE areas have no VALUE clause;
	// their content before the first READ is undefined on the mainframe.
	// We use the INITIALIZE-like constructors (spaces / zeros).
	return &TaudisSource{
		store:   store,
		layouts: l,
		dirHdr:  taudis.NewEntete(),
		dir:     taudis.NewDirection(),
		comHdr:  taudis.NewEntete(),
		com:     taudis.NewCommune(),
		rHdr:    taudis.NewEntete(),
		ifp:     l.NewIFP(),
		segAN:   zes.NewAN(),
		segDIR:  zes.NewDIR(),
		segCOM:  l.NewCOM(),
		segSR:   l.NewSR(),
	}, nil
}

// Fetch runs EFITAUX2 for req. A Go error is only returned for conditions
// that would abend the COBOL program (store failure, corrupt record).
func (s *TaudisSource) Fetch(ctx context.Context, req Request) (Result, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	res := Result{ZES: newZESSpaces()}
	aiftres := fixed(req.Ccoifp, 3) + fixed(req.Ccpper, 3)

	animpSet := !isZoneZero(req.Animp, 4)
	dirSet := !isSpaces(fixed(req.Coddir, 3))
	comSet := !isZoneZero(req.Codcom, 3)
	aifSet := !isSpaces(aiftres)

	switch {
	case animpSet && dirSet && comSet && aifSet:
		if cr, err := s.readDir(ctx, req); err != nil || cr != 0 {
			return earlyExit(res, cr, err)
		}
		if cr, err := s.readCom(ctx, req); err != nil || cr != 0 {
			return earlyExit(res, cr, err)
		}
		if s.com.Gcovef() != "V" {
			res.CR, res.RC = CRNotValidated, RCNotValidated
			return res, nil
		}
		if cr, err := s.readIFP(ctx, req); err != nil || cr != 0 {
			return earlyExit(res, cr, err)
		}
	case animpSet && dirSet && comSet && !aifSet:
		if cr, err := s.readDir(ctx, req); err != nil || cr != 0 {
			return earlyExit(res, cr, err)
		}
		if cr, err := s.readCom(ctx, req); err != nil || cr != 0 {
			return earlyExit(res, cr, err)
		}
	case animpSet && dirSet && !comSet && !aifSet:
		if cr, err := s.readDir(ctx, req); err != nil || cr != 0 {
			return earlyExit(res, cr, err)
		}
	default:
		// No WHEN matched: no read, segments are built from stale areas.
	}

	if err := s.fillReturn(req); err != nil {
		return res, err
	}

	segs := [][]byte{s.segAN.Bytes(), s.segDIR.Bytes(), s.segCOM.Bytes(), s.segSR.Bytes()}
	for i, b := range segs {
		if len(b) != ZESSlotLen {
			return res, fmt.Errorf("rates: ZES segment %d is %d bytes, want %d", i+1, len(b), ZESSlotLen)
		}
		copy(res.ZES[i][:], b)
	}
	return res, nil
}

// earlyExit mirrors EXIT PROGRAM from a read paragraph: ZES stays spaces.
func earlyExit(res Result, cr int, err error) (Result, error) {
	if err != nil {
		return res, err
	}
	res.CR, res.RC = cr, RCReadFailed
	return res, nil
}

func (s *TaudisSource) read(ctx context.Context, key [taudis.EnteteLen]byte) (taudis.Entete, []byte, bool, error) {
	rec, found, err := s.store.Read(ctx, key)
	if err != nil {
		return taudis.Entete{}, nil, false, fmt.Errorf("rates: read TAUDIS %q: %w", key[:], err)
	}
	if !found {
		return taudis.Entete{}, nil, false, nil
	}
	hdr, body, err := taudis.SplitRecord(rec)
	if err != nil {
		return taudis.Entete{}, nil, false, fmt.Errorf("rates: TAUDIS %q: %w", key[:], err)
	}
	return hdr, body, true, nil
}

// readDir is APPEL-DIR: key = CODDIR + LOW-VALUE x 9.
func (s *TaudisSource) readDir(ctx context.Context, req Request) (int, error) {
	key := buildKey(req.Coddir, "", "", "", 1)
	hdr, body, ok, err := s.read(ctx, key)
	if err != nil || !ok {
		return CRDirNotFound, err
	}
	d, err := taudis.ParseDirection(body)
	if err != nil {
		return 0, err
	}
	s.dirHdr, s.dir = hdr, d
	return 0, nil
}

// readCom is APPEL-COM: key = CODDIR + CODCOM + LOW-VALUE x 6.
func (s *TaudisSource) readCom(ctx context.Context, req Request) (int, error) {
	key := buildKey(req.Coddir, req.Codcom, "", "", 2)
	hdr, body, ok, err := s.read(ctx, key)
	if err != nil || !ok {
		return CRComNotFound, err
	}
	c, err := taudis.ParseCommune(body)
	if err != nil {
		return 0, err
	}
	s.comHdr, s.com = hdr, c
	return 0, nil
}

// readIFP is APPEL-AIFTRE: key = CODDIR + CODCOM + CCOIFP + CCPPER.
func (s *TaudisSource) readIFP(ctx context.Context, req Request) (int, error) {
	key := buildKey(req.Coddir, req.Codcom, req.Ccoifp, req.Ccpper, 4)
	hdr, body, ok, err := s.read(ctx, key)
	if err != nil || !ok {
		return CRIFPNotFound, err
	}
	r, err := s.layouts.ParseIFP(body)
	if err != nil {
		return 0, fmt.Errorf("rates: parse IFP body: %w", err)
	}
	s.rHdr, s.ifp = hdr, r
	return 0, nil
}

// buildKey builds the 12-byte TAUDIS key; parts beyond 'used' are
// LOW-VALUE filled.
func buildKey(dir, com, ifp, per string, used int) [taudis.EnteteLen]byte {
	var k [taudis.EnteteLen]byte // zero = LOW-VALUE
	parts := []string{dir, com, ifp, per}
	for i := 0; i < used; i++ {
		copy(k[i*3:i*3+3], fixed(parts[i], 3))
	}
	return k
}

// fillReturn is SERV-RET-LINK.
func (s *TaudisSource) fillReturn(req Request) error {
	animp := zonedValue(fixed(req.Animp, 4))
	spaces := "             "

	s.segAN.SetAidfic(spaces)
	s.segDIR.SetAidfic(spaces)
	s.segCOM.SetAidfic(spaces)
	s.segSR.SetAidfic(spaces)

	// General zones.
	s.segAN.SetJan(int(animp))
	s.segAN.SetCcocom("000") // MOVE ZERO to an X(3) field.

	// Direction zones.
	s.segDIR.SetJan(animp)
	s.segDIR.SetDepdir(s.comHdr.Depdir()) // sic: from the Commune header.
	v, _ := s.dir.Ptbdep()
	s.segDIR.SetTaudepB(v)
	v, _ = s.dir.Ptbtas()
	s.segDIR.SetPtbtas(v)
	v, _ = s.dir.Pnbagr()
	s.segDIR.SetTauchagrNB(v)
	if !s.dir.PnbagnIsSentinel() {
		v, _ = s.dir.Pnbagn()
		s.segDIR.SetTausmagrNB(v)
	} else {
		s.segDIR.SetTausmagrNB(0)
	}

	// Commune zones.
	c := s.segCOM
	c.SetJan(animp)
	c.SetDepdir(s.comHdr.Depdir())
	c.SetCcocom(s.comHdr.Ccocom())
	// JGCIN5/JGCRI5 (N-4) -> index 0 ... JGCIN1/JGCRI1 (N) -> index 4.
	for i := 0; i < taudis.AgrjaCount; i++ {
		n := taudis.AgrjaCount - i
		in, _, err := s.com.Jgcin(n)
		if err != nil {
			return err
		}
		ri, _, err := s.com.Jgcri(n)
		if err != nil {
			return err
		}
		if err := c.SetJgcins(i, in); err != nil {
			return err
		}
		if err := c.SetJgcrim(i, ri); err != nil {
			return err
		}
	}
	v, _ = s.com.Dnuro()
	c.SetDnuro(v)
	c.SetDnultf(s.com.Dnultf())
	c.SetJdhomo(s.com.Jdaho1())
	c.SetGcoanu(s.com.Gcoanu())

	// IFP / Trésorerie zones.
	r := s.ifp
	sr := s.segSR
	sr.SetJan(animp)
	sr.SetDepdir(s.rHdr.Depdir())
	sr.SetCcocom(s.rHdr.Ccocom())
	sr.SetCcoifp(s.rHdr.Ccoifp())
	v, _ = r.Ptbcom()
	sr.SetTaucomB(v)
	v, _ = r.Pnbcom()
	sr.SetTaucomNB(v)
	v, _ = r.Ptbsyn()
	sr.SetTausyndsfpB(v)
	v, _ = r.Pnbsyn()
	sr.SetTausyndsfpNB(v)
	v, _ = r.Ptbcu()
	sr.SetTaucudfpvnB(v)
	v, _ = r.Pnbcu()
	sr.SetTaucudfpvnNB(v)
	v, _ = r.Pnbtxa()
	sr.SetPnbtxa(v)
	tb1, _, err := r.Ptbtsn(0)
	if err != nil {
		return err
	}
	tn1, _, err := r.Pnbtsn(0)
	if err != nil {
		return err
	}
	tb2, _, err := r.Ptbtsn(1)
	if err != nil {
		return err
	}
	tn2, _, err := r.Pnbtsn(1)
	if err != nil {
		return err
	}
	sr.SetTautseB(tb1)
	sr.SetTautseNB(tn1)
	sr.SetPtbtgp(tb2)
	sr.SetPnbtgp(tn2)
	v, _ = r.Pnbcaa()
	sr.SetPnbcaa(v)
	v, _ = r.Ptbgem()
	sr.SetPtbgem(v)
	v, _ = r.Pnbgem()
	sr.SetPnbgem(v)

	// Communal JA deliberations go into the C segment.
	for i := 0; i < taudis.AgrjaCount; i++ {
		n := taudis.AgrjaCount - i
		in, _, err := r.Jcoin(n)
		if err != nil {
			return err
		}
		ri, _, err := r.Jcori(n)
		if err != nil {
			return err
		}
		if err := c.SetJcoins(i, in); err != nil {
			return err
		}
		if err := c.SetJcorim(i, ri); err != nil {
			return err
		}
	}
	v, _ = r.Ntxom()
	c.SetNbtauxom(v)
	v, _ = r.Pbbomp()
	c.SetPbbomp(v)
	v, _ = r.Pbboma()
	c.SetPbboma(v)
	v, _ = r.Pbbomb()
	c.SetPbbomb(v)
	v, _ = r.Pbbomc()
	c.SetPbbomc(v)
	v, _ = r.Pbbomd()
	c.SetPbbomd(v)
	v, _ = r.Pbbome()
	c.SetPbbome(v)
	// 22 = commune, 23 = EPCI, 34 = OM syndicate.
	sr.SetLigbenefom(r.Ccobom())
	return nil
}

// fixed pads/truncates s to n bytes (COBOL alphanumeric field semantics).
func fixed(s string, n int) string {
	if len(s) >= n {
		return s[:n]
	}
	b := []byte(s)
	for len(b) < n {
		b = append(b, ' ')
	}
	return string(b)
}

func isSpaces(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] != ' ' {
			return false
		}
	}
	return true
}

// isZoneZero is the numeric "= 0" test of an unsigned DISPLAY field; any
// non-digit byte makes the field count as non-zero (documented quirk).
func isZoneZero(s string, n int) bool {
	f := fixed(s, n)
	for i := 0; i < len(f); i++ {
		if f[i] != '0' {
			return false
		}
	}
	return true
}

// zonedValue decodes an unsigned DISPLAY field from the low nibbles, as a
// numeric MOVE on IBM COBOL would (invalid nibbles clamped to 9).
func zonedValue(s string) uint64 {
	var v uint64
	for i := 0; i < len(s); i++ {
		d := uint64(s[i] & 0x0F)
		if d > 9 {
			d = 9
		}
		v = v*10 + d
	}
	return v
}
