package taudis

import (
	"fmt"
)

// IFPTresoLetter is the record-type discriminator (DLETAR) of a TAUDIS
// IFP / trésorerie article.
const IFPTresoLetter = "I"

// PcfsynScale is the scaling factor of the PIC 9(3)V9(2) syndicate
// allocation keys PCFSYA..PCFSYJ / PCFSYN(i): the accessor returns the value
// multiplied by 100 (two implied decimals). The PIC 9(4)V9(6) rate fields
// use RateScale (six implied decimals).
const PcfsynScale = 100

// OCCURS counts of the IFP / trésorerie layout.
const (
	// IFPTsePairCount is the OCCURS count of PTBTSE, PNBTSE, TBBTS, TBNTS,
	// MCBTS, MCNTS, MPBTSE, MPNTSE and ACOET.
	IFPTsePairCount = 2
	// IFPSyndCount is the OCCURS count of APBPOS, APNPOS and ACFPOS
	// (syndicates A..J).
	IFPSyndCount = 10
	// IFPJaCount is the OCCURS count of ACJARO (young-farmer deliberations,
	// occurrence 1 = JCOIN5/JCORI5 = year N-4, occurrence 5 = year N).
	IFPJaCount = 5
)

// tresoKind is the storage class of a DISPLAY field.
type tresoKind int

const (
	tresoAlpha    tresoKind = iota // PIC X(n)
	tresoUnsigned                  // PIC 9(n) / 9(n)V9(m), zoned, unsigned
	tresoSigned                    // PIC S9(n), zoned, trailing overpunch sign
)

// tresoField locates one elementary item inside the 2988-byte body.
type tresoField struct {
	off  int
	n    int
	kind tresoKind
}

// tresoLayoutBuilder lays the copybook out sequentially, exactly like the
// COBOL compiler assigns storage: every item starts where the previous one
// ended.
type tresoLayoutBuilder struct {
	off    int
	fields map[string]tresoField
	order  []string
	err    error
}

func (b *tresoLayoutBuilder) filler(n int) { b.off += n }

func (b *tresoLayoutBuilder) add(name string, n int, k tresoKind) {
	if _, dup := b.fields[name]; dup && b.err == nil {
		b.err = fmt.Errorf("taudis: T84R duplicate field %s", name)
	}
	b.fields[name] = tresoField{off: b.off, n: n, kind: k}
	b.order = append(b.order, name)
	b.off += n
}

// sx adds a 1-byte 'X'-FILLER separator followed by the field.
func (b *tresoLayoutBuilder) sx(name string, n int, k tresoKind) {
	b.filler(1)
	b.add(name, n, k)
}

// alias registers a REDEFINES name sharing the bytes of target.
func (b *tresoLayoutBuilder) alias(name, target string) {
	f, ok := b.fields[target]
	if !ok {
		if b.err == nil {
			b.err = fmt.Errorf("taudis: T84R alias %s targets unknown field %s", name, target)
		}
		return
	}
	if _, dup := b.fields[name]; dup && b.err == nil {
		b.err = fmt.Errorf("taudis: T84R duplicate field %s", name)
	}
	b.fields[name] = f
	b.order = append(b.order, name)
}

func occ(name string, i int) string { return fmt.Sprintf("%s(%d)", name, i) }

var syndLetters = [IFPSyndCount]string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J"}

// buildTresoLayout mirrors the DONNEES-IFP-TRESO copybook item by item.
func buildTresoLayout() (map[string]tresoField, []string, error) {
	const (
		X = tresoAlpha
		U = tresoUnsigned
		S = tresoSigned
	)
	b := &tresoLayoutBuilder{fields: make(map[string]tresoField)}

	// Identification.
	b.sx("DLETAR", 1, X)
	b.sx("DNULTF", 1, X)
	b.sx("DNURO", 3, U)
	b.sx("GCOANU", 1, X)
	b.sx("DACROL", 3, X)
	b.sx("GCOVEF", 1, X)
	b.sx("GORVEF", 1, X)
	b.sx("CCOFDL", 1, X)
	b.sx("CCPDIR", 3, X)
	b.sx("CSPER", 3, X)

	// TAUX: PIC 9(4)V9(6).
	for _, n := range []string{"PTBCOM", "PTBSYN", "PTBCU", "PTBDEP"} {
		b.sx(n, 10, U)
	}
	for i := 1; i <= IFPTsePairCount; i++ {
		b.sx(occ("PTBTSN", i), 10, U)
	}
	for _, n := range []string{"PTBTAS", "PTBGEM", "PTBTST", "PNBCOM", "PNBSYN", "PNBCU", "PNBTXA"} {
		b.sx(n, 10, U)
	}
	for i := 1; i <= IFPTsePairCount; i++ {
		b.sx(occ("PNBTSN", i), 10, U)
	}
	for _, n := range []string{"PNBGEM", "PNBAGR", "PNBCAA", "PNBTST", "PNBAGN",
		"PBBOMP", "PBBOMA", "PBBOMB", "PBBOMC", "PBBOMD", "PBBOME"} {
		b.sx(n, 10, U)
	}
	b.sx("NTXOM", 1, U)

	// BASES: PIC 9(15).
	for _, n := range []string{"TBBCO", "TBBSY", "TBBCU1", "TBBDE"} {
		b.sx(n, 15, U)
	}
	for i := 1; i <= IFPTsePairCount; i++ {
		b.sx(occ("TBBTSN", i), 15, U)
	}
	for _, n := range []string{"TBBTAS", "TBBGEM", "TBNCO", "TBNSY", "TBNCU1", "TBNDE"} {
		b.sx(n, 15, U)
	}
	for i := 1; i <= IFPTsePairCount; i++ {
		b.sx(occ("TBNTSN", i), 15, U)
	}
	for _, n := range []string{"TBNGEM", "TBNC1", "TBNCAA", "TBBOM1", "TBBOMA", "TBBOMB",
		"TBBOMC", "TBBOMD", "TBBOME", "TBBOM2"} {
		b.sx(n, 15, U)
	}

	// COTISATIONS: PIC 9(10).
	for _, n := range []string{"MCBCO3", "MCBSY3", "MCBCU3", "MCBDE3"} {
		b.sx(n, 10, U)
	}
	for i := 1; i <= IFPTsePairCount; i++ {
		b.sx(occ("MCBTSN", i), 10, U)
	}
	for _, n := range []string{"MCBTA3", "MCBTSA", "MCBGE3", "MCBETA", "MCNCO3", "MCNSY3", "MCNCU3", "MCNDE3"} {
		b.sx(n, 10, U)
	}
	for i := 1; i <= IFPTsePairCount; i++ {
		b.sx(occ("MCNTSN", i), 10, U)
	}
	for _, n := range []string{"MCNTSA", "MCNGE3", "MCNC1", "MCNAP1", "MCNAF2", "MCNETA",
		"MCTCO1", "MCTSY1", "MCTCU1", "MCTDE1",
		"MC1OMP", "MC1OMA", "MC1OMB", "MC1OMC", "MC1OMD", "MC1OME", "MC1OMI", "MCTOM1"} {
		b.sx(n, 10, U)
	}
	b.alias("TCTOM", "MCTOM1") // TCTOM REDEFINES MCTOM1

	// PRODUITS-ATTENDUS: PIC 9(10).
	b.sx("MPBCOM", 10, U)
	for i, l := range syndLetters {
		b.sx("MPBSY"+l, 10, U)
		_ = i
	}
	for _, n := range []string{"MPBCU", "MPBDEP"} {
		b.sx(n, 10, U)
	}
	for i := 1; i <= IFPTsePairCount; i++ {
		b.sx(occ("MPBTSN", i), 10, U)
	}
	for _, n := range []string{"MPBTAS", "MPBGEM", "MPNCOM"} {
		b.sx(n, 10, U)
	}
	for _, l := range syndLetters {
		b.sx("MPNSY"+l, 10, U)
	}
	for _, n := range []string{"MPNCU", "MPNDEP"} {
		b.sx(n, 10, U)
	}
	for i := 1; i <= IFPTsePairCount; i++ {
		b.sx(occ("MPNTSN", i), 10, U)
	}
	for _, n := range []string{"MPNGEM", "MPNCAR", "MPNAAP", "MPOCOM", "MPOSY", "MPOCU"} {
		b.sx(n, 10, U)
	}
	// APBTAB REDEFINES APBSYN, APNTAB REDEFINES APNSYN.
	for i, l := range syndLetters {
		b.alias(occ("MPBSYN", i+1), "MPBSY"+l)
		b.alias(occ("MPNSYN", i+1), "MPNSY"+l)
	}

	// Beneficiaries / deliberations.
	b.sx("CCOBOM", 2, U)
	b.filler(1)
	b.add("ACOSYN", 0, X) // group marker, fixed below
	b.add("CTYSYN", 1, X)
	b.add("CCOSYN", 3, X)
	b.sx("ACOGOM", 4, X)
	b.sx("CCOBIM", 2, U)
	b.sx("ACOBIM", 4, X)
	b.sx("GIMOMI", 1, X)
	for i := 1; i <= IFPTsePairCount; i++ {
		b.sx(occ("ACOETS", i), 4, X)
		b.sx(occ("DNUTSE", i), 2, U)
	}
	b.sx("ACOTXA", 4, X)
	b.sx("ACOGEM", 4, X)
	for _, n := range []string{"MCTAMD", "MMTROL", "MCT80", "MROANU"} {
		b.sx(n, 10, U)
	}
	b.sx("NDGRPV", 6, U)
	b.sx("CCOEXC", 2, X)
	b.sx("MEXECF", 12, U)
	b.sx("NDGPV", 5, U)
	b.sx("MCTPV", 10, U)
	b.sx("MDGANU", 10, U)

	// CLEF-REPARTITION-SYND: PIC 9(3)V9(2), ACFTAB REDEFINES ACFSYN.
	for _, l := range syndLetters {
		b.sx("PCFSY"+l, 5, U)
	}
	for i, l := range syndLetters {
		b.alias(occ("PCFSYN", i+1), "PCFSY"+l)
	}

	// COMPTAGES.
	for _, n := range []string{"NIMTFB", "NIMOM", "NIMTFN", "NIMCAA", "NEXZUS", "NIMTEB"} {
		b.sx(n, 7, U)
	}
	b.sx("NIMCA", 6, U)
	for _, n := range []string{"NIBTSN", "NINTSN", "NIBTAS", "NIBGEM", "NINGEM"} {
		b.sx(n, 7, U)
	}
	b.sx("NBAFFO", 6, U)
	for _, n := range []string{"NROPRO", "NIMPRO", "N80PRO", "NEXECF", "NEXEC1"} {
		b.sx(n, 7, U)
	}

	// DELIB-JA-COMMUNALES: AGPJAG (no separator between JCOINn and JCORIn).
	for n := IFPJaCount; n >= 1; n-- {
		b.sx(fmt.Sprintf("JCOIN%d", n), 4, U)
		b.add(fmt.Sprintf("JCORI%d", n), 4, U)
	}
	// ACJAR REDEFINES AGPJAG: occurrence i holds JCOIN(6-i) / JCORI(6-i).
	for i := 1; i <= IFPJaCount; i++ {
		b.alias(occ("CINSN", i), fmt.Sprintf("JCOIN%d", IFPJaCount+1-i))
		b.alias(occ("CRIMN", i), fmt.Sprintf("JCORI%d", IFPJaCount+1-i))
	}
	for _, n := range []string{"VDJCOM", "VDJSYN", "VDJCU", "VDJETA"} {
		b.sx(n, 10, U)
	}
	for _, n := range []string{"NDGJAE", "NDGJAC", "NDGJAM", "NDGJAS", "NDGJAG"} {
		b.sx(n, 5, U)
	}
	b.sx("MCTZOM", 10, U)
	for _, n := range []string{"NIMETA", "NIMCO0", "NDGRJE", "NDGRJC"} {
		b.sx(n, 6, U)
	}

	// FRAIS.
	for _, n := range []string{"MFSTFB", "MFSTFN", "MFSCAA", "MFSETB", "MFSETN",
		"FRBA10", "FRBA20", "FRBA30", "FREBA1", "FREBA2", "FREBA3",
		"FRNB10", "FRNB20", "FRNB30", "FENB10", "FENB20", "FENB30",
		"FRBA80", "FRNB80",
		"FRBA36", "FRBA54", "FRBA90", "FREB36", "FREB54", "FREB90",
		"FRNB36", "FRNB54", "FRNB90", "FENB36", "FENB54", "FENB90"} {
		b.sx(n, 10, U)
	}
	for _, sfx := range []string{"41", "31"} {
		b.sx("MFSB"+sfx, 10, S)
		b.sx("MFSN"+sfx, 10, S)
		b.sx("MFEB"+sfx, 10, U)
		b.sx("MFEN"+sfx, 10, U)
		b.sx("MFSC"+sfx, 10, S)
	}

	// LISSAGES and LISSAGES-SYNDICATS: PIC S9(10).
	for _, n := range []string{"MLBCOM", "MLBSYN", "MLBCU", "MLBDEP", "MLBTS1", "MLBTS2",
		"MLBTAS", "MLBGEM", "MLBTOM"} {
		b.sx(n, 10, S)
	}
	for _, l := range syndLetters {
		b.sx("MLBSY"+l, 10, S)
	}

	// AGEREF and trailer.
	b.filler(1)
	agerefOff := b.off
	b.add("DCPNU1", 7, U)
	b.add("KCPCLE", 2, X)
	b.fields["AGEREF"] = tresoField{off: agerefOff, n: 9, kind: X}
	b.order = append(b.order, "AGEREF")
	b.sx("DCPNU9", 7, U)
	b.add("KNOCP", 2, X)
	b.add("LIBRE", 312, X) // trailing FILLER X(312)

	// Fix the ACOSYN group to cover CTYSYN + CCOSYN.
	if f, ok := b.fields["CTYSYN"]; ok {
		b.fields["ACOSYN"] = tresoField{off: f.off, n: 4, kind: X}
	}

	if b.err == nil && b.off != BodyLen {
		b.err = fmt.Errorf("taudis: T84R layout is %d bytes, want %d", b.off, BodyLen)
	}
	return b.fields, b.order, b.err
}

var tresoFields, tresoOrder, tresoLayoutErr = buildTresoLayout()

// IFPTresoFieldNames returns the names of every addressable field of the
// T84R layout (COBOL names without the 'X'- prefix; OCCURS items are named
// NAME(i) with the 1-based COBOL subscript), in copybook order.
func IFPTresoFieldNames() []string {
	out := make([]string, len(tresoOrder))
	copy(out, tresoOrder)
	return out
}

// IFPTreso is the DONNEES-IFP-TRESO body (copybook T84R) of a TAUDIS 'I'
// record: the 2988 bytes that follow the 12-byte Entete (which carries
// DEPDIR, CCOCOM and CCOIFP). Fields are accessed over the raw storage so
// that decoding and encoding are byte-exact; separator FILLERs and the
// trailing FILLER X(312) survive a round trip untouched.
//
// MIGRATION_NOTE: the copybook's 'X'- prefix is a COPY REPLACING placeholder
// (EFITAUX2 includes it as T-R-); in Go each inclusion is a distinct
// IFPTreso value. Because the layout has ~300 elementary items, it is
// described by a declarative table (buildTresoLayout) and accessed through
// generic name-based accessors, plus typed accessors for the fields EFITAUX2
// actually consumes. REDEFINES (TCTOM, MPBSYN(i), MPNSYN(i), PCFSYN(i),
// CINSN(i), CRIMN(i)) are aliases sharing the same bytes.
type IFPTreso [BodyLen]byte

// NewIFPTreso returns a body whose alphanumeric fields and FILLERs are
// spaces, numeric fields are zeros, and whose article letter DLETAR is 'I'.
func NewIFPTreso() (IFPTreso, error) {
	var r IFPTreso
	if tresoLayoutErr != nil {
		return r, tresoLayoutErr
	}
	for i := range r {
		r[i] = ' '
	}
	for _, f := range tresoFields {
		switch f.kind {
		case tresoUnsigned:
			r.setUnsigned(f, 0)
		case tresoSigned:
			r.setSigned(f, 0)
		}
	}
	r.put(tresoFields["DLETAR"], IFPTresoLetter)
	return r, nil
}

// ParseIFPTreso copies a 2988-byte body into an IFPTreso. It fails when b
// does not have exactly BodyLen bytes.
func ParseIFPTreso(b []byte) (IFPTreso, error) {
	var r IFPTreso
	if tresoLayoutErr != nil {
		return r, tresoLayoutErr
	}
	if len(b) != BodyLen {
		return r, fmt.Errorf("taudis: ifp-treso body must be %d bytes, got %d", BodyLen, len(b))
	}
	copy(r[:], b)
	return r, nil
}

// Bytes returns a copy of the raw 2988 bytes (encoded layout).
func (r *IFPTreso) Bytes() []byte {
	out := make([]byte, BodyLen)
	copy(out, r[:])
	return out
}

// ---- low-level helpers ----

func lookupTreso(name string) (tresoField, error) {
	f, ok := tresoFields[name]
	if !ok {
		return tresoField{}, fmt.Errorf("taudis: T84R has no field %q", name)
	}
	return f, nil
}

func (r *IFPTreso) get(f tresoField) string { return string(r[f.off : f.off+f.n]) }

// put performs a COBOL alphanumeric MOVE: left-justified, truncated on the
// right, padded with spaces.
func (r *IFPTreso) put(f tresoField, v string) {
	for i := 0; i < f.n; i++ {
		if i < len(v) {
			r[f.off+i] = v[i]
		} else {
			r[f.off+i] = ' '
		}
	}
}

// unsigned decodes an unsigned zoned-decimal field. ok is false when the
// field fails the NUMERIC test; the value is then built from the low nibble
// of each byte (clamped to 9) and must be treated as unreliable.
func (r *IFPTreso) unsigned(f tresoField) (uint64, bool) {
	ok := true
	var v uint64
	for _, c := range r[f.off : f.off+f.n] {
		if c < '0' || c > '9' {
			ok = false
		}
		d := uint64(c & 0x0F)
		if d > 9 {
			d = 9 // MIGRATION_NOTE: invalid nibble clamped to keep the value in range
		}
		v = v*10 + d
	}
	return v, ok
}

// setUnsigned performs a COBOL numeric MOVE into an unsigned DISPLAY field:
// high-order digits beyond the picture are truncated, zero-filled left.
func (r *IFPTreso) setUnsigned(f tresoField, v uint64) {
	for i := f.off + f.n - 1; i >= f.off; i-- {
		r[i] = byte('0' + v%10)
		v /= 10
	}
}

// overpunch decodes the last byte of a trailing-sign zoned field.
// Accepted: plain digits (positive), '{','A'..'I' (positive), '}','J'..'R'
// (negative, IBM EBCDIC→ASCII translation) and 'p'..'y' (negative,
// GnuCOBOL/Micro Focus ASCII convention).
func overpunch(c byte) (digit uint64, neg bool, ok bool) {
	switch {
	case c >= '0' && c <= '9':
		return uint64(c - '0'), false, true
	case c == '{':
		return 0, false, true
	case c >= 'A' && c <= 'I':
		return uint64(c-'A') + 1, false, true
	case c == '}':
		return 0, true, true
	case c >= 'J' && c <= 'R':
		return uint64(c-'J') + 1, true, true
	case c >= 'p' && c <= 'y':
		return uint64(c - 'p'), true, true
	}
	d := uint64(c & 0x0F)
	if d > 9 {
		d = 9
	}
	return d, false, false
}

// signed decodes a PIC S9(n) DISPLAY field with a trailing overpunch sign.
// ok is false when the field fails the NUMERIC test.
func (r *IFPTreso) signed(f tresoField) (int64, bool) {
	ok := true
	var v uint64
	last := f.off + f.n - 1
	for _, c := range r[f.off:last] {
		if c < '0' || c > '9' {
			ok = false
		}
		d := uint64(c & 0x0F)
		if d > 9 {
			d = 9
		}
		v = v*10 + d
	}
	d, neg, lastOK := overpunch(r[last])
	v = v*10 + d
	if neg {
		return -int64(v), ok && lastOK
	}
	return int64(v), ok && lastOK
}

// setSigned performs a COBOL numeric MOVE into a PIC S9(n) DISPLAY field
// (trailing overpunch, no SIGN SEPARATE), truncating high-order digits.
//
// MIGRATION_NOTE: the sign is written with the IBM EBCDIC→ASCII overpunch
// convention ('{','A'..'I' positive, '}','J'..'R' negative), matching what a
// signed MOVE on z/OS produces (C/D zone). Review if the TAUDIS store uses
// another ASCII convention.
func (r *IFPTreso) setSigned(f tresoField, v int64) {
	neg := v < 0
	u := uint64(v)
	if neg {
		u = uint64(-v)
	}
	for i := f.off + f.n - 1; i >= f.off; i-- {
		r[i] = byte('0' + u%10)
		u /= 10
	}
	last := f.off + f.n - 1
	d := r[last] - '0'
	switch {
	case neg && d == 0:
		r[last] = '}'
	case neg:
		r[last] = 'J' + d - 1
	case d == 0:
		r[last] = '{'
	default:
		r[last] = 'A' + d - 1
	}
}

// ---- generic name-based accessors ----

// Raw returns the raw bytes of the named field (any kind), e.g. "PTBCOM",
// "PTBTSN(2)" or "ACOSYN".
func (r *IFPTreso) Raw(name string) (string, error) {
	f, err := lookupTreso(name)
	if err != nil {
		return "", err
	}
	return r.get(f), nil
}

// SetRaw performs an alphanumeric MOVE of v into the named field.
func (r *IFPTreso) SetRaw(name, v string) error {
	f, err := lookupTreso(name)
	if err != nil {
		return err
	}
	r.put(f, v)
	return nil
}

// Num returns the value of an unsigned numeric field (implied decimals are
// returned scaled: RateScale for 9(4)V9(6), PcfsynScale for 9(3)V9(2)).
// ok reports the COBOL NUMERIC test. It fails for unknown or non-unsigned
// fields.
func (r *IFPTreso) Num(name string) (uint64, bool, error) {
	f, err := lookupTreso(name)
	if err != nil {
		return 0, false, err
	}
	if f.kind != tresoUnsigned {
		return 0, false, fmt.Errorf("taudis: T84R field %s is not unsigned numeric", name)
	}
	v, ok := r.unsigned(f)
	return v, ok, nil
}

// SetNum moves v into an unsigned numeric field (scaled value for fields
// with implied decimals).
func (r *IFPTreso) SetNum(name string, v uint64) error {
	f, err := lookupTreso(name)
	if err != nil {
		return err
	}
	if f.kind != tresoUnsigned {
		return fmt.Errorf("taudis: T84R field %s is not unsigned numeric", name)
	}
	r.setUnsigned(f, v)
	return nil
}

// Signed returns the value of a PIC S9(10) field (MFSB41, MLBCOM, ...).
// ok reports the COBOL NUMERIC test.
func (r *IFPTreso) Signed(name string) (int64, bool, error) {
	f, err := lookupTreso(name)
	if err != nil {
		return 0, false, err
	}
	if f.kind != tresoSigned {
		return 0, false, fmt.Errorf("taudis: T84R field %s is not signed numeric", name)
	}
	v, ok := r.signed(f)
	return v, ok, nil
}

// SetSigned moves v into a PIC S9(10) field.
func (r *IFPTreso) SetSigned(name string, v int64) error {
	f, err := lookupTreso(name)
	if err != nil {
		return err
	}
	if f.kind != tresoSigned {
		return fmt.Errorf("taudis: T84R field %s is not signed numeric", name)
	}
	r.setSigned(f, v)
	return nil
}

// numNamed decodes a field known to exist in the layout.
func (r *IFPTreso) numNamed(name string) (uint64, bool) {
	f, ok := tresoFields[name]
	if !ok {
		return 0, false
	}
	return r.unsigned(f)
}

// ---- typed accessors used by EFITAUX2 ----

// Dletar returns DLETAR X, the article letter ('I' for IFP / trésorerie).
func (r *IFPTreso) Dletar() string { return r.get(tresoFields["DLETAR"]) }

// SetDletar moves v into DLETAR X.
func (r *IFPTreso) SetDletar(v string) { r.put(tresoFields["DLETAR"], v) }

// IsIFPTreso reports whether DLETAR holds the 'I' article letter.
func (r *IFPTreso) IsIFPTreso() bool { return r.Dletar() == IFPTresoLetter }

// Ptbcom returns PTBCOM 9(4)V9(6), the commune built rate, scaled by RateScale.
func (r *IFPTreso) Ptbcom() (uint64, bool) { return r.numNamed("PTBCOM") }

// Pnbcom returns PNBCOM 9(4)V9(6), the commune non-built rate, scaled by RateScale.
func (r *IFPTreso) Pnbcom() (uint64, bool) { return r.numNamed("PNBCOM") }

// Ptbsyn returns PTBSYN 9(4)V9(6), the syndicate built rate, scaled by RateScale.
func (r *IFPTreso) Ptbsyn() (uint64, bool) { return r.numNamed("PTBSYN") }

// Pnbsyn returns PNBSYN 9(4)V9(6), the syndicate non-built rate, scaled by RateScale.
func (r *IFPTreso) Pnbsyn() (uint64, bool) { return r.numNamed("PNBSYN") }

// Ptbcu returns PTBCU 9(4)V9(6), the EPCI built rate, scaled by RateScale.
func (r *IFPTreso) Ptbcu() (uint64, bool) { return r.numNamed("PTBCU") }

// Pnbcu returns PNBCU 9(4)V9(6), the EPCI non-built rate, scaled by RateScale.
func (r *IFPTreso) Pnbcu() (uint64, bool) { return r.numNamed("PNBCU") }

// Pnbtxa returns PNBTXA 9(4)V9(6), the additional non-built tax rate,
// scaled by RateScale.
func (r *IFPTreso) Pnbtxa() (uint64, bool) { return r.numNamed("PNBTXA") }

// Pnbcaa returns PNBCAA 9(4)V9(6), the chamber of agriculture (Alsace-Moselle)
// rate, scaled by RateScale.
func (r *IFPTreso) Pnbcaa() (uint64, bool) { return r.numNamed("PNBCAA") }

// Ptbgem returns PTBGEM 9(4)V9(6), the built GEMAPI rate, scaled by RateScale.
func (r *IFPTreso) Ptbgem() (uint64, bool) { return r.numNamed("PTBGEM") }

// Pnbgem returns PNBGEM 9(4)V9(6), the non-built GEMAPI rate, scaled by RateScale.
func (r *IFPTreso) Pnbgem() (uint64, bool) { return r.numNamed("PNBGEM") }

// Ptbtsn returns PTBTSN(i+1) 9(4)V9(6) (0-based i; occurrence 0 is the TSE
// rate, occurrence 1 the Grand Paris TSE rate), scaled by RateScale.
func (r *IFPTreso) Ptbtsn(i int) (uint64, bool, error) {
	if err := checkIndex("PTBTSE", i, IFPTsePairCount); err != nil {
		return 0, false, err
	}
	v, ok := r.numNamed(occ("PTBTSN", i+1))
	return v, ok, nil
}

// Pnbtsn returns PNBTSN(i+1) 9(4)V9(6) (0-based i), scaled by RateScale.
func (r *IFPTreso) Pnbtsn(i int) (uint64, bool, error) {
	if err := checkIndex("PNBTSE", i, IFPTsePairCount); err != nil {
		return 0, false, err
	}
	v, ok := r.numNamed(occ("PNBTSN", i+1))
	return v, ok, nil
}

// Jcoin returns JCOINn 9(4) for n in 1..5 (1 = year N, 5 = year N-4), the
// young-farmer installation year of the communal deliberation.
func (r *IFPTreso) Jcoin(n int) (uint64, bool, error) {
	if err := checkIndex("JCOIN suffix-1", n-1, IFPJaCount); err != nil {
		return 0, false, err
	}
	v, ok := r.numNamed(fmt.Sprintf("JCOIN%d", n))
	return v, ok, nil
}

// Jcori returns JCORIn 9(4) for n in 1..5 (1 = year N, 5 = year N-4), the
// return-to-taxation year of the communal deliberation.
func (r *IFPTreso) Jcori(n int) (uint64, bool, error) {
	if err := checkIndex("JCORI suffix-1", n-1, IFPJaCount); err != nil {
		return 0, false, err
	}
	v, ok := r.numNamed(fmt.Sprintf("JCORI%d", n))
	return v, ok, nil
}

// Ntxom returns NTXOM 9, the number of TEOM rates.
func (r *IFPTreso) Ntxom() (uint64, bool) { return r.numNamed("NTXOM") }

// Pbbomp returns PBBOMP 9(4)V9(6), the TEOM full-rate (plein), scaled by RateScale.
func (r *IFPTreso) Pbbomp() (uint64, bool) { return r.numNamed("PBBOMP") }

// Pbboma returns PBBOMA 9(4)V9(6), TEOM zone A rate, scaled by RateScale.
func (r *IFPTreso) Pbboma() (uint64, bool) { return r.numNamed("PBBOMA") }

// Pbbomb returns PBBOMB 9(4)V9(6), TEOM zone B rate, scaled by RateScale.
func (r *IFPTreso) Pbbomb() (uint64, bool) { return r.numNamed("PBBOMB") }

// Pbbomc returns PBBOMC 9(4)V9(6), TEOM zone C rate, scaled by RateScale.
func (r *IFPTreso) Pbbomc() (uint64, bool) { return r.numNamed("PBBOMC") }

// Pbbomd returns PBBOMD 9(4)V9(6), TEOM zone D rate, scaled by RateScale.
func (r *IFPTreso) Pbbomd() (uint64, bool) { return r.numNamed("PBBOMD") }

// Pbbome returns PBBOME 9(4)V9(6), TEOM zone E rate, scaled by RateScale.
func (r *IFPTreso) Pbbome() (uint64, bool) { return r.numNamed("PBBOME") }

// Ccobom returns CCOBOM 9(2), the TEOM beneficiary code.
func (r *IFPTreso) Ccobom() (uint64, bool) { return r.numNamed("CCOBOM") }

// CcobomRaw returns the raw 2 bytes of CCOBOM.
func (r *IFPTreso) CcobomRaw() string { return r.get(tresoFields["CCOBOM"]) }

// Mpbsyn returns MPBSYN(i+1) 9(10) (0-based i, syndicate A..J), the
// expected built syndicate proceeds (APBTAB REDEFINES APBSYN).
func (r *IFPTreso) Mpbsyn(i int) (uint64, bool, error) {
	if err := checkIndex("APBPOS", i, IFPSyndCount); err != nil {
		return 0, false, err
	}
	v, ok := r.numNamed(occ("MPBSYN", i+1))
	return v, ok, nil
}

// Mpnsyn returns MPNSYN(i+1) 9(10) (0-based i, syndicate A..J), the
// expected non-built syndicate proceeds (APNTAB REDEFINES APNSYN).
func (r *IFPTreso) Mpnsyn(i int) (uint64, bool, error) {
	if err := checkIndex("APNPOS", i, IFPSyndCount); err != nil {
		return 0, false, err
	}
	v, ok := r.numNamed(occ("MPNSYN", i+1))
	return v, ok, nil
}

// Pcfsyn returns PCFSYN(i+1) 9(3)V9(2) (0-based i, syndicate A..J), the
// syndicate allocation key scaled by PcfsynScale (ACFTAB REDEFINES ACFSYN).
func (r *IFPTreso) Pcfsyn(i int) (uint64, bool, error) {
	if err := checkIndex("ACFPOS", i, IFPSyndCount); err != nil {
		return 0, false, err
	}
	v, ok := r.numNamed(occ("PCFSYN", i+1))
	return v, ok, nil
}

// Libre returns the trailing reserved FILLER X(312).
func (r *IFPTreso) Libre() string { return r.get(tresoFields["LIBRE"]) }
