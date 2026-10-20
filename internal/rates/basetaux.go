package rates

import (
	"context"
	"fmt"
	"sync"
)

// This file is the Go port of FMSTAU2, the DL/I "primitive" that reads the
// local-tax rate hierarchical database (PCB GPHXTD00) for a tax year,
// direction, commune and IFP code. It returns up to 7 records in the ZES
// return area (EL-REC) and reports its status through CR / RC.
//
// MIGRATION_NOTE: the segment copybooks (XB30..XB51, XBXTDDIR), the linkage
// copybook XTAU2, the IMS DBD, FMSPCB and CBLTDLI were not available at
// migration time. The control flow is transcribed exactly. Segment layouts
// are injected through BaseTauxLayout (lengths, INITIALIZE images and field
// offsets) and IMS access through PCBProvider / SegmentReader. Until a real
// IMS adapter exists, UnavailableBaseTaux reports CR 24 / RC 01, which is
// what the callers (EFITA3B8 / EFITA3N8) set on CALL ... ON EXCEPTION.

// Return codes reported by FMSTAU2 through CR / RC.
const (
	// CRPCBUnavailable: FMSPCB could not resolve the PCB (RC = RETURN-CODE).
	CRPCBUnavailable = 18
	// CRDLIError: unexpected DL/I status (RC = the two status bytes).
	CRDLIError = 16
	// CRBaseTauxComNotVoted / RCBaseTauxComNotVoted: commune found but
	// GTOVOT <> 'V'.
	CRBaseTauxComNotVoted = 12
	RCBaseTauxComNotVoted = 10
	// CRYearNotFound .. CRDirOriNotFound accompany RC 01 on a 'GE' status.
	CRYearNotFound      = 20
	CRBaseDirNotFound   = 21
	CRBaseComNotFound   = 22
	CRBaseIFPNotFound   = 23
	CRIFPSuiteNotFound  = 24
	CRDirOriNotFound    = 25
	RCBaseTauxNotFound  = 1
	baseTauxCRThreshold = 2 // every step runs only while CR NOT > 2
)

// BaseTauxPCBName is the PCB name resolved through FMSPCB.
const BaseTauxPCBName = "GPHXTD00"

// BaseTauxKeyLen is the length of the concatenated-key SSA qualification
// (SSAQ-CLE PIC X(13)).
const BaseTauxKeyLen = 13

// baseTauxOriAdtsegLen is the reference-modified length copied from the
// origin direction segment (XB35-ORI-ADTSEG(1:385)).
const baseTauxOriAdtsegLen = 385

const (
	dliStatusOK = "  "
	dliStatusGE = "GE"
)

// SegmentReader is the DL/I port behind one PCB. GU performs a Get Unique
// with the qualified SSA 'XEXTDXTD*C(' + ssaKey + ')' and fills segment (the
// I/O area) on success. It returns the two-byte PCB status ("  " = found,
// "GE" = not found, anything else = DL/I error). As with IMS, the I/O area
// must be left untouched on a non-blank status. A Go error is reserved for
// infrastructure failures (COBOL abend conditions).
type SegmentReader interface {
	GU(ctx context.Context, ssaKey [BaseTauxKeyLen]byte, segment []byte) (status string, err error)
}

// PCBProvider replaces CALL 'FMSPCB'. A non-zero returnCode is the COBOL
// RETURN-CODE of FMSPCB (reported as CR 18); err is an abend condition.
type PCBProvider interface {
	PCB(ctx context.Context, name string) (reader SegmentReader, returnCode int, err error)
}

// SegmentID identifies a base-taux working segment.
type SegmentID int

// Segments handled by FMSTAU2.
const (
	SegXB30 SegmentID = iota + 1 // year data
	SegXB35                      // direction data
	SegXB35Ori                   // origin direction data (XBXTDDIR layout)
	SegXB40                      // commune data
	SegXB45                      // IFP data
	SegXB46                      // IFP data, continuation
	SegXB36                      // TSE organism label
	SegXB37                      // EPCI label
	SegXB38                      // TEOM label
	SegXB47                      // syndicate labels
	SegXB41                      // commune e-mail address
	SegXB50                      // TSE/EPCI labels restitution
	SegXB51                      // syndicate labels / commune e-mail restitution
)

var allBaseTauxSegments = []SegmentID{
	SegXB30, SegXB35, SegXB35Ori, SegXB40, SegXB45, SegXB46, SegXB36,
	SegXB37, SegXB38, SegXB47, SegXB41, SegXB50, SegXB51,
}

// FieldKind is the storage class of a segment field; it drives MOVE ZERO
// and numeric-to-numeric MOVE alignment.
type FieldKind int

// Field kinds.
const (
	FieldAlnum       FieldKind = iota // PIC X
	FieldZoned                        // PIC 9 DISPLAY (unsigned)
	FieldZonedSigned                  // PIC S9 DISPLAY, trailing IBM overpunch
	FieldPacked                       // COMP-3
)

// FieldRef locates a field inside its segment.
type FieldRef struct {
	Off  int
	Len  int
	Kind FieldKind
}

// FieldID identifies a segment field accessed by FMSTAU2.
type FieldID int

// Fields accessed by FMSTAU2.
const (
	FXB40Ctygc FieldID = iota + 1
	FXB40Dnucol
	FXB40Ctysyn
	FXB40Ccosyn
	FXB40Gtovot
	FXB40Gnidep
	FXB40Pbbomp
	FXB40Pbboma
	FXB40Pbbomb
	FXB40Pbbomc
	FXB40Pbbomd
	FXB40Pbbome
	FXB40Ntxom
	FXB40Acosyn
	FXB40Acogom
	FXB40Acobim
	FXB40Acotxa
	FXB40Acogem
	FXB40Jcoins1
	FXB40Jcoins2
	FXB40Jcoins3
	FXB40Jcoins4
	FXB40Jcoins5
	FXB40Jcorim1
	FXB40Jcorim2
	FXB40Jcorim3
	FXB40Jcorim4
	FXB40Jcorim5
	FXB46Pbbomp
	FXB46Pbboma
	FXB46Pbbomb
	FXB46Pbbomc
	FXB46Pbbomd
	FXB46Pbbome
	FXB46Ntxom
	FXB46Acosyn
	FXB46Acogom
	FXB46Acobim
	FXB46Acotxa
	FXB46Acogem
	FXB46Jcoins1
	FXB46Jcoins2
	FXB46Jcoins3
	FXB46Jcoins4
	FXB46Jcoins5
	FXB46Jcorim1
	FXB46Jcorim2
	FXB46Jcorim3
	FXB46Jcorim4
	FXB46Jcorim5
	FXB35Ptbdep
	FXB35Adtseg
	FXB35OriAdtseg
	FXB45Ccddir
	FXB45Acoets1
	FXB45Acoets2
	FXB37Vmailg
	FXB37Vmaild
	FXB50Aenr1
	FXB50Aenr2
	FXB50Aenr3
	FXB50Vmailg
	FXB50Vmaild
	FXB41Vmailg
	FXB41Vmaild
	FXB51Vmailg
	FXB51Vmaild
	fieldIDEnd
)

// fieldSegment maps each field to the segment that holds it.
func fieldSegment(f FieldID) SegmentID {
	switch {
	case f >= FXB40Ctygc && f <= FXB40Jcorim5:
		return SegXB40
	case f >= FXB46Pbbomp && f <= FXB46Jcorim5:
		return SegXB46
	case f == FXB35Ptbdep || f == FXB35Adtseg:
		return SegXB35
	case f == FXB35OriAdtseg:
		return SegXB35Ori
	case f >= FXB45Ccddir && f <= FXB45Acoets2:
		return SegXB45
	case f == FXB37Vmailg || f == FXB37Vmaild:
		return SegXB37
	case f >= FXB50Aenr1 && f <= FXB50Vmaild:
		return SegXB50
	case f == FXB41Vmailg || f == FXB41Vmaild:
		return SegXB41
	default:
		return SegXB51
	}
}

// ifpSuiteMoves is the list of MOVE XB46-x TO XB40-x of RECH-IFP-SUITE, in
// source order.
var ifpSuiteMoves = [][2]FieldID{
	{FXB46Pbbomp, FXB40Pbbomp},
	{FXB46Pbboma, FXB40Pbboma},
	{FXB46Pbbomb, FXB40Pbbomb},
	{FXB46Pbbomc, FXB40Pbbomc},
	{FXB46Pbbomd, FXB40Pbbomd},
	{FXB46Pbbome, FXB40Pbbome},
	{FXB46Ntxom, FXB40Ntxom},
	{FXB46Acosyn, FXB40Acosyn},
	{FXB46Acogom, FXB40Acogom},
	{FXB46Acobim, FXB40Acobim},
	{FXB46Acotxa, FXB40Acotxa},
	{FXB46Acogem, FXB40Acogem},
	{FXB46Jcoins1, FXB40Jcoins1},
	{FXB46Jcorim1, FXB40Jcorim1},
	{FXB46Jcoins2, FXB40Jcoins2},
	{FXB46Jcorim2, FXB40Jcorim2},
	{FXB46Jcoins3, FXB40Jcoins3},
	{FXB46Jcorim3, FXB40Jcorim3},
	{FXB46Jcoins4, FXB40Jcoins4},
	{FXB46Jcorim4, FXB40Jcorim4},
	{FXB46Jcoins5, FXB40Jcoins5},
	{FXB46Jcorim5, FXB40Jcorim5},
}

// BaseTauxLayout supplies the copybook-derived layouts that were missing at
// migration time. NewSegment returns a fresh INITIALIZEd image (spaces for
// alphanumerics, zeros for numerics) of the segment; Field returns the
// position of a field within its segment.
type BaseTauxLayout interface {
	NewSegment(id SegmentID) []byte
	Field(id FieldID) FieldRef
}

// BaseTauxRequest carries the FMSTAU2 linkage parameters as raw COBOL bytes.
type BaseTauxRequest struct {
	Janipt string // FIE01-JANIPT PIC 9(4)
	Acodir string // FIE01-ACODIR PIC X(3); FIE01-CODEP = first 2 bytes
	Ccocom string // FIE01-CCOCOM PIC 9(3)
	Ccoifp string // FIE01-CCOIFP PIC X(3)
}

// BaseTauxResult is the FMSTAU2 output. RCRaw holds the exact two bytes of
// RC PIC 9(2): after CR 16 they are the DL/I status (e.g. "AI"), which is
// not numeric; RC is the numeric reading of those bytes (low nibbles).
type BaseTauxResult struct {
	CR    int
	RC    int
	RCRaw [2]byte
	ZES   ZES
}

// BaseTauxFetcher is the FMSTAU2 entry point used by the calculators.
type BaseTauxFetcher interface {
	Fetch(ctx context.Context, req BaseTauxRequest) (BaseTauxResult, error)
}

// UnavailableBaseTaux is the default FMSTAU2 adapter while no IMS access
// exists: it behaves like the caller's CALL ... ON EXCEPTION (CR 24, RC 01).
// MIGRATION_NOTE: on ON EXCEPTION the caller's ZES is left unchanged; this
// adapter returns a spaces-filled ZES, callers must not rely on it.
type UnavailableBaseTaux struct{}

// Fetch always reports CR 24 / RC 01.
func (UnavailableBaseTaux) Fetch(_ context.Context, _ BaseTauxRequest) (BaseTauxResult, error) {
	res := BaseTauxResult{ZES: newZESSpaces()}
	res.setRC(CRIFPSuiteNotFound, RCBaseTauxNotFound)
	return res, nil
}

// BaseTauxSource is the Go port of FMSTAU2. Like the non-INITIAL COBOL
// program, the key work areas (W-TAU-CLE, W-ORI-CLE, W-LIB-CLE) and the
// origin direction segment persist between calls. Calls are serialised.
type BaseTauxSource struct {
	mu     sync.Mutex
	pcbs   PCBProvider
	layout BaseTauxLayout
	fields map[FieldID]FieldRef

	tauCle  [BaseTauxKeyLen]byte // W-TAU-CLE: AN 9(4) DIR X(3) COM 9(3) IFP X(3)
	oriCle  [BaseTauxKeyLen]byte // W-ORI-CLE
	libCle  [BaseTauxKeyLen]byte // W-LIB-CLE: AN DIR CTYGC DNUCOL FILLER X(2)
	xb35Ori []byte               // XB35-ORI, never INITIALIZEd by the program
}

// NewBaseTauxSource builds the FMSTAU2 rate source. It validates that every
// field fits inside its segment.
func NewBaseTauxSource(pcbs PCBProvider, layout BaseTauxLayout) (*BaseTauxSource, error) {
	if pcbs == nil {
		return nil, fmt.Errorf("rates: nil PCB provider")
	}
	if layout == nil {
		return nil, fmt.Errorf("rates: nil base-taux layout")
	}
	lens := make(map[SegmentID]int, len(allBaseTauxSegments))
	for _, id := range allBaseTauxSegments {
		seg := layout.NewSegment(id)
		if len(seg) == 0 {
			return nil, fmt.Errorf("rates: empty layout for segment %d", id)
		}
		lens[id] = len(seg)
	}
	fields := make(map[FieldID]FieldRef, int(fieldIDEnd))
	for f := FieldID(1); f < fieldIDEnd; f++ {
		ref := layout.Field(f)
		segLen := lens[fieldSegment(f)]
		if ref.Off < 0 || ref.Len <= 0 || ref.Off+ref.Len > segLen {
			return nil, fmt.Errorf("rates: field %d (%+v) does not fit segment %d (len %d)", f, ref, fieldSegment(f), segLen)
		}
		fields[f] = ref
	}
	if fields[FXB35OriAdtseg].Len < baseTauxOriAdtsegLen {
		return nil, fmt.Errorf("rates: XB35-ORI-ADTSEG shorter than %d bytes", baseTauxOriAdtsegLen)
	}
	s := &BaseTauxSource{
		pcbs:    pcbs,
		layout:  layout,
		fields:  fields,
		xb35Ori: layout.NewSegment(SegXB35Ori),
	}
	// MIGRATION_NOTE: WORKING-STORAGE without VALUE is undefined on the
	// mainframe before first use; we start from INITIALIZE-like content.
	s.tauCle = initTauCle()
	s.oriCle = initTauCle()
	copy(s.libCle[:], "             ") // FILLER VALUE SPACE, other parts set before use
	return s, nil
}

// initTauCle is INITIALIZE W-TAU-CLE.
func initTauCle() [BaseTauxKeyLen]byte {
	var k [BaseTauxKeyLen]byte
	copy(k[:], "0000   000   ")
	return k
}

// baseTauxCall holds the per-call state of one FMSTAU2 invocation.
type baseTauxCall struct {
	s      *BaseTauxSource
	reader SegmentReader
	res    *BaseTauxResult

	janipt string
	acodir string // FIE01-ACODIR, may be changed by PELP-COM
	ccocom string
	ccoifp string

	xb30, xb35, xb36, xb37, xb38, xb40, xb41, xb45, xb46, xb47, xb50, xb51 []byte
}

func (c *baseTauxCall) codep() string { return c.acodir[:2] }

func (c *baseTauxCall) ok() bool { return c.res.CR <= baseTauxCRThreshold }

func (c *baseTauxCall) field(seg []byte, f FieldID) []byte {
	ref := c.s.fields[f]
	return seg[ref.Off : ref.Off+ref.Len]
}

func (c *baseTauxCall) fieldStr(seg []byte, f FieldID) string { return string(c.field(seg, f)) }

// moveField is a field-to-field MOVE. Zoned numerics are right-aligned
// (left zero-padded / left-truncated); everything else is an alphanumeric
// move (left-aligned, space-padded, right-truncated).
func (c *baseTauxCall) moveField(dstSeg []byte, dst FieldID, srcSeg []byte, src FieldID) {
	dr, sr := c.s.fields[dst], c.s.fields[src]
	d, s := c.field(dstSeg, dst), c.field(srcSeg, src)
	if dr.Kind == FieldZoned && sr.Kind == FieldZoned && dr.Len != sr.Len {
		moveZoned(d, s)
		return
	}
	moveAlnum(d, s)
}

// zeroField is MOVE ZERO TO field.
func (c *baseTauxCall) zeroField(seg []byte, f FieldID) {
	ref := c.s.fields[f]
	d := c.field(seg, f)
	switch ref.Kind {
	case FieldPacked:
		for i := range d {
			d[i] = 0
		}
		d[len(d)-1] = 0x0C
	case FieldZonedSigned:
		for i := range d {
			d[i] = '0'
		}
		d[len(d)-1] = '{' // IBM positive zero overpunch
	default:
		for i := range d {
			d[i] = '0'
		}
	}
}

// moveAlnum is an alphanumeric (group) MOVE.
func moveAlnum(dst, src []byte) {
	n := copy(dst, src)
	for i := n; i < len(dst); i++ {
		dst[i] = ' '
	}
}

// moveZoned is an unsigned DISPLAY numeric MOVE between different lengths.
func moveZoned(dst, src []byte) {
	for i := range dst {
		dst[i] = '0'
	}
	for i, j := len(dst)-1, len(src)-1; i >= 0 && j >= 0; i, j = i-1, j-1 {
		dst[i] = src[j]
	}
}

func (r *BaseTauxResult) setRC(cr, rc int) {
	r.CR = cr
	if rc < 0 {
		rc = -rc
	}
	rc %= 100
	r.RC = rc
	r.RCRaw = [2]byte{byte('0' + rc/10), byte('0' + rc%10)}
}

// setDLIError is MOVE 16 TO CR, MOVE DLI-RETOUR TO RC.
func (r *BaseTauxResult) setDLIError(status string) {
	r.CR = CRDLIError
	r.RCRaw = [2]byte{status[0], status[1]}
	r.RC = int(zonedValue(status))
}

func (c *baseTauxCall) gu(ctx context.Context, key [BaseTauxKeyLen]byte, seg []byte) (string, error) {
	st, err := c.reader.GU(ctx, key, seg)
	if err != nil {
		return "", fmt.Errorf("rates: DL/I GU %q: %w", key[:], err)
	}
	return fixed(st, 2), nil
}

// Fetch runs FMSTAU2 for req. A Go error is only returned for conditions
// that would abend the COBOL program (PCB provider / DL/I infrastructure
// failure).
func (s *BaseTauxSource) Fetch(ctx context.Context, req BaseTauxRequest) (BaseTauxResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	res := BaseTauxResult{ZES: newZESSpaces()}
	res.setRC(0, 0)
	l := s.layout
	// INIT: INITIALIZE of every working segment and CR / RC.
	c := &baseTauxCall{
		s: s, res: &res,
		janipt: fixed(req.Janipt, 4),
		acodir: fixed(req.Acodir, 3),
		ccocom: fixed(req.Ccocom, 3),
		ccoifp: fixed(req.Ccoifp, 3),
		xb30:   l.NewSegment(SegXB30),
		xb35:   l.NewSegment(SegXB35),
		xb36:   l.NewSegment(SegXB36),
		xb37:   l.NewSegment(SegXB37),
		xb38:   l.NewSegment(SegXB38),
		xb40:   l.NewSegment(SegXB40),
		xb41:   l.NewSegment(SegXB41),
		xb45:   l.NewSegment(SegXB45),
		xb46:   l.NewSegment(SegXB46),
		xb47:   l.NewSegment(SegXB47),
		xb50:   l.NewSegment(SegXB50),
		xb51:   l.NewSegment(SegXB51),
	}

	// ACCES-TAUX.
	reader, code, err := s.pcbs.PCB(ctx, BaseTauxPCBName)
	if err != nil {
		return res, fmt.Errorf("rates: FMSPCB %s: %w", BaseTauxPCBName, err)
	}
	if code != 0 {
		res.setRC(CRPCBUnavailable, code)
	} else if reader == nil {
		return res, fmt.Errorf("rates: FMSPCB %s returned no PCB", BaseTauxPCBName)
	} else {
		c.reader = reader
	}

	// PELP-DEPT-MULTI-DSF. W-ACODIR-ENTREE = req.Acodir; ACODIR is restored
	// on exit, so the caller never sees the direction found by PELP-COM.
	janiptSet := func() bool { return !isZoneZero(c.janipt, 4) }
	comSet := func() bool { return !isZoneZero(c.ccocom, 3) }
	dirSet := func() bool { return !isSpaces(c.acodir) }
	switch cd := c.codep(); {
	case c.ok() && janiptSet() && (cd == "75" || cd == "13" || cd == "59" || cd == "92") && comSet():
		if err := c.pelpCom(ctx); err != nil {
			return res, err
		}
	}

	// RECUP-SEG.
	if c.ok() && janiptSet() {
		if err := c.rechAnnee(ctx); err != nil {
			return res, err
		}
	}
	if c.ok() && janiptSet() && dirSet() {
		if err := c.rechDir(ctx); err != nil {
			return res, err
		}
	}
	if c.ok() && janiptSet() && dirSet() && comSet() {
		if err := c.rechCom(ctx); err != nil {
			return res, err
		}
		if c.ok() && !isSpaces(c.fieldStr(c.xb40, FXB40Ctygc)) && !isSpaces(c.fieldStr(c.xb40, FXB40Dnucol)) {
			if err := c.rechLepci(ctx); err != nil {
				return res, err
			}
		}
		if c.ok() && c.teomCodeSet() {
			if err := c.rechLteom(ctx); err != nil {
				return res, err
			}
		}
	}
	if c.ok() && janiptSet() && dirSet() && comSet() && !isSpaces(c.ccoifp) {
		if err := c.rechIFP(ctx); err != nil {
			return res, err
		}
		if zonedValue(c.janipt) > 2015 {
			// Not guarded by CR in the source.
			if err := c.rechIFPSuite(ctx); err != nil {
				return res, err
			}
			// From 2016 the TEOM grouping code is at SR level.
			if c.ok() && c.teomCodeSet() {
				if err := c.rechLteom(ctx); err != nil {
					return res, err
				}
			}
		}
		// Not guarded by CR in the source.
		if err := c.rechLib(ctx); err != nil {
			return res, err
		}
		if c.ok() && !isSpaces(c.fieldStr(c.xb45, FXB45Ccddir)) {
			if err := c.rechDirOri(ctx); err != nil {
				return res, err
			}
		}
	}
	if c.ok() && janiptSet() && dirSet() && comSet() {
		if err := c.rechLabelIFP(ctx, "Z  ", c.xb47, nil); err != nil { // RECH-LSYN
			return res, err
		}
	}
	if c.ok() && janiptSet() && dirSet() && comSet() {
		if err := c.rechLabelIFP(ctx, "Y  ", c.xb41, c.applyLmel); err != nil { // RECH-LMEL
			return res, err
		}
	}

	// Restitution.
	out := [ZESSlotCount][]byte{c.xb30, c.xb35, c.xb40, c.xb45, c.xb50, c.xb51, c.xb47}
	for i, seg := range out {
		moveAlnum(res.ZES[i][:], seg)
	}
	return res, nil
}

func (c *baseTauxCall) teomCodeSet() bool {
	return !isSpaces(c.fieldStr(c.xb40, FXB40Ctysyn)) && !isSpaces(c.fieldStr(c.xb40, FXB40Ccosyn))
}

// key setters of W-TAU-CLE.
func (c *baseTauxCall) setTauAn(v string)  { copy(c.s.tauCle[0:4], fixed(v, 4)) }
func (c *baseTauxCall) setTauDir(v string) { copy(c.s.tauCle[4:7], fixed(v, 3)) }
func (c *baseTauxCall) setTauCom(v string) { copy(c.s.tauCle[7:10], fixed(v, 3)) }
func (c *baseTauxCall) setTauIFP(v string) { copy(c.s.tauCle[10:13], fixed(v, 3)) }

// pelpCom is PELP-COM: search the direction the commune belongs to among
// all directions of the multi-direction department.
func (c *baseTauxCall) pelpCom(ctx context.Context) error {
	minDir, maxDir := 1, 2
	if c.codep() == "75" {
		minDir, maxDir = 4, 8
	}
	found := false
	ind := minDir
	c.setTauAn(c.janipt)
	c.setTauDir(c.acodir)
	c.s.tauCle[6] = byte('0' + ind)
	c.setTauCom(c.ccocom)
	c.setTauIFP("")
	for ind <= maxDir && !found {
		st, err := c.gu(ctx, c.s.tauCle, c.xb40)
		if err != nil {
			return err
		}
		switch st {
		case dliStatusGE:
			ind = (ind + 1) % 10 // IND-DIR PIC 9
			c.s.tauCle[6] = byte('0' + ind)
		case dliStatusOK:
			found = true
			c.acodir = string(c.s.tauCle[4:7])
			c.xb40 = c.s.layout.NewSegment(SegXB40)
			c.s.tauCle = initTauCle()
		default:
			c.res.setDLIError(st)
			found = true
		}
	}
	return nil
}

// checkGE applies the common EVALUATE: GE -> notFoundCR/01, other non-blank
// -> 16/status. It reports whether the status was blank.
func (c *baseTauxCall) checkGE(st string, notFoundCR int) bool {
	switch st {
	case dliStatusOK:
		return true
	case dliStatusGE:
		c.res.setRC(notFoundCR, RCBaseTauxNotFound)
	default:
		c.res.setDLIError(st)
	}
	return false
}

// checkLabel applies IF DLI-RETOUR NOT = SPACE AND NOT = 'GE' -> 16. It
// reports whether the ELSE branch (blank or GE) is taken.
func (c *baseTauxCall) checkLabel(st string) bool {
	if st != dliStatusOK && st != dliStatusGE {
		c.res.setDLIError(st)
		return false
	}
	return true
}

// rechAnnee is RECH-ANNEE.
func (c *baseTauxCall) rechAnnee(ctx context.Context) error {
	c.setTauAn(c.janipt)
	c.setTauDir("")
	c.setTauCom("000")
	c.setTauIFP("")
	st, err := c.gu(ctx, c.s.tauCle, c.xb30)
	if err != nil {
		return err
	}
	c.checkGE(st, CRYearNotFound)
	return nil
}

// rechDir is RECH-DIR.
func (c *baseTauxCall) rechDir(ctx context.Context) error {
	c.setTauDir(c.acodir)
	st, err := c.gu(ctx, c.s.tauCle, c.xb35)
	if err != nil {
		return err
	}
	c.checkGE(st, CRBaseDirNotFound)
	return nil
}

// rechCom is RECH-COM.
func (c *baseTauxCall) rechCom(ctx context.Context) error {
	c.setTauCom(c.ccocom)
	st, err := c.gu(ctx, c.s.tauCle, c.xb40)
	if err != nil {
		return err
	}
	if c.checkGE(st, CRBaseComNotFound) {
		if c.fieldStr(c.xb40, FXB40Gtovot) != "V" {
			c.res.setRC(CRBaseTauxComNotVoted, RCBaseTauxComNotVoted)
		}
		if c.fieldStr(c.xb40, FXB40Gnidep) == "N" {
			c.zeroField(c.xb35, FXB35Ptbdep)
		}
	}
	return nil
}

// rechIFP is RECH-IFP.
func (c *baseTauxCall) rechIFP(ctx context.Context) error {
	c.setTauIFP(c.ccoifp)
	st, err := c.gu(ctx, c.s.tauCle, c.xb45)
	if err != nil {
		return err
	}
	c.checkGE(st, CRBaseIFPNotFound)
	return nil
}

// rechIFPSuite is RECH-IFP-SUITE: overlays OM rates and JA deliberations
// from XB46 into XB40.
func (c *baseTauxCall) rechIFPSuite(ctx context.Context) error {
	c.setTauDir("")
	c.setTauIFP(c.ccoifp)
	st, err := c.gu(ctx, c.s.tauCle, c.xb46)
	if err != nil {
		return err
	}
	if c.checkGE(st, CRIFPSuiteNotFound) {
		for _, m := range ifpSuiteMoves {
			c.moveField(c.xb40, m[1], c.xb46, m[0])
		}
	}
	return nil
}

// rechDirOri is RECH-DIR-ORI: the commune changed department, read the
// origin direction and copy its data into XB35.
func (c *baseTauxCall) rechDirOri(ctx context.Context) error {
	copy(c.s.oriCle[0:4], c.janipt)
	copy(c.s.oriCle[4:7], fixed(c.fieldStr(c.xb45, FXB45Ccddir), 3))
	copy(c.s.oriCle[7:10], "000")
	copy(c.s.oriCle[10:13], "   ")
	st, err := c.gu(ctx, c.s.oriCle, c.s.xb35Ori)
	if err != nil {
		return err
	}
	if c.checkGE(st, CRDirOriNotFound) {
		src := c.field(c.s.xb35Ori, FXB35OriAdtseg)[:baseTauxOriAdtsegLen]
		moveAlnum(c.field(c.xb35, FXB35Adtseg), src)
		if c.fieldStr(c.xb40, FXB40Gnidep) == "N" {
			c.zeroField(c.xb35, FXB35Ptbdep)
		}
	}
	return nil
}

// setLibKey fills W-LIB-CLE. paris754 applies the '754' direction override
// for department 75 (RECH-LEPCI / RECH-LIB, not RECH-LTEOM).
func (c *baseTauxCall) setLibKey(paris754 bool, ctygc, dnucol string) {
	copy(c.s.libCle[0:4], c.janipt)
	dir := c.acodir
	if paris754 && c.codep() == "75" {
		dir = "754"
	}
	copy(c.s.libCle[4:7], dir)
	copy(c.s.libCle[7:8], fixed(ctygc, 1))
	copy(c.s.libCle[8:11], fixed(dnucol, 3))
	copy(c.s.libCle[11:13], "  ")
}

// rechLepci is RECH-LEPCI.
func (c *baseTauxCall) rechLepci(ctx context.Context) error {
	c.setLibKey(true, c.fieldStr(c.xb40, FXB40Ctygc), c.fieldStr(c.xb40, FXB40Dnucol))
	st, err := c.gu(ctx, c.s.libCle, c.xb37)
	if err != nil {
		return err
	}
	c.checkLabel(st)
	return nil
}

// rechLteom is RECH-LTEOM.
func (c *baseTauxCall) rechLteom(ctx context.Context) error {
	c.setLibKey(false, c.fieldStr(c.xb40, FXB40Ctysyn), c.fieldStr(c.xb40, FXB40Ccosyn))
	st, err := c.gu(ctx, c.s.libCle, c.xb38)
	if err != nil {
		return err
	}
	if c.checkLabel(st) {
		moveAlnum(c.xb51, c.xb38)
	}
	return nil
}

// rechLib is RECH-LIB: TSE labels and preparation of XB50.
func (c *baseTauxCall) rechLib(ctx context.Context) error {
	moveAlnum(c.field(c.xb50, FXB50Aenr3), c.xb37)
	c.moveField(c.xb50, FXB50Vmailg, c.xb37, FXB37Vmailg)
	c.moveField(c.xb50, FXB50Vmaild, c.xb37, FXB37Vmaild)

	steps := []struct {
		acoets FieldID
		aenr   FieldID
	}{
		{FXB45Acoets1, FXB50Aenr1}, // TSE
		{FXB45Acoets2, FXB50Aenr2}, // TSE GP, MA, GA
	}
	for _, step := range steps {
		c.xb36 = c.s.layout.NewSegment(SegXB36)
		c.setLibKey(true, "", "")
		// MOVE XB45-ACOETS(n) TO W-LIB-CLE1 (group X(4)).
		moveAlnum(c.s.libCle[7:11], c.field(c.xb45, step.acoets))
		st, err := c.gu(ctx, c.s.libCle, c.xb36)
		if err != nil {
			return err
		}
		c.checkLabel(st)
		moveAlnum(c.field(c.xb50, step.aenr), c.xb36)
	}
	return nil
}

// rechLabelIFP is RECH-LSYN (ifp 'Z  ') / RECH-LMEL (ifp 'Y  '); onOK runs
// in the ELSE branch (blank or GE status).
func (c *baseTauxCall) rechLabelIFP(ctx context.Context, ifp string, seg []byte, onOK func()) error {
	c.setTauDir(c.acodir)
	c.setTauIFP(ifp)
	st, err := c.gu(ctx, c.s.tauCle, seg)
	if err != nil {
		return err
	}
	// The copy into XB51 only happens when the segment was actually found:
	// on 'GE' the I/O area is untouched and XB51 keeps the TEOM label.
	if c.checkLabel(st) && st == dliStatusOK && onOK != nil {
		onOK()
	}
	return nil
}

// applyLmel copies the commune e-mail address into XB51.
func (c *baseTauxCall) applyLmel() {
	c.moveField(c.xb51, FXB51Vmailg, c.xb41, FXB41Vmailg)
	c.moveField(c.xb51, FXB51Vmaild, c.xb41, FXB41Vmaild)
}
