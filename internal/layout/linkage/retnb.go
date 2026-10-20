package linkage

import "fmt"

// RetNBLen is the total length in bytes of the non-built-property
// assessment return record (copybook XRETNB).
const RetNBLen = 600

// OCCURS counts of the XRETNB record.
const (
	// RetNBNt13Count is the OCCURS count of MCNT13.
	RetNBNt13Count = 2
	// RetNBJt13Count is the OCCURS count of MDJT13.
	RetNBJt13Count = 2
)

// Group / filler lengths of the XRETNB record.
const (
	// RetNBAidficLen is the length of the AIDFIC group.
	RetNBAidficLen = 11
	// RetNBAnuproLen is the length of the ANUPRO group.
	RetNBAnuproLen = 6
	// RetNBFillerLen is the length of the first 'X'-FILLER X(10).
	RetNBFillerLen = 10
	// RetNBTrailerLen is the length of the trailing 'X'-FILLER X(145).
	RetNBTrailerLen = 145
)

// Zero-based offsets inside the 600-byte XRETNB record.
//
//	AIDFIC   X(11)      0..10   DAN 9(4) 0..3, AC3DIR 4..6 (CC2DEP 4..5,
//	                            CCODIR 6), CCOCOM 7..9, DSRPAR 10
//	ANUPRO   X(6)      11..16   CGROUP 11, NNUPRO 9(5) 12..16
//	CCOBNB   X         17
//	MCTCOM 18, MCTDEP 28, MCTREG 38, MCTSYN 48, MCTCU 58, MCTTSE 68 (S9(10))
//	MCNT13   S9(10) OCCURS 2   78..97
//	MCTCA 98, MCTBAP 108, MCTCAP 118 (S9(10))
//	MCTCAD   S9(5)            128..132
//	MCNTSA   9(10)            133..142
//	MCFFIX 143, MFA300 153 (=MCFANO), MFN300 163 (=MCFNVN), MCFABA 173,
//	MCFNVB 183, MCFACA 193, MCFNVC 203 (S9(10))
//	FILLER   X(10)            213..222
//	MDGPLA 223, MDGFAP 233, MDGFNV 243, MDJCOM 253, MDJSYN 263, MDJCU 273,
//	MDJTSE 283, MDJDEP 293 (S9(10))
//	MDJT13   S9(10) OCCURS 2  303..322
//	TCTHFR 323, TCTFRA 335, TCTDU 347 (S9(12))
//	MCNJAE 359, MFAJ1E 369, MFNJ1E 379, MCTCAS 389, MFA800 399 (=MCAANO),
//	MFN800 409 (=MCANVN), MFA900 419, MFN900 429, MCOGE3 439 (S9(10))
//	CCOIFP   X(3)             449..451
//	CCPPER   X(3)             452..454
//	FILLER   X(145)           455..599
const (
	rnbOffAidfic  = 0
	rnbOffDan     = 0
	rnbOffAc3dir  = 4
	rnbOffCc2dep  = 4
	rnbOffCcodir  = 6
	rnbOffCcocom  = 7
	rnbOffDsrpar  = 10
	rnbOffAnupro  = 11
	rnbOffCgroup  = 11
	rnbOffNnupro  = 12
	rnbOffCcobnb  = 17
	rnbOffMctcom  = 18
	rnbOffMctdep  = 28
	rnbOffMctreg  = 38
	rnbOffMctsyn  = 48
	rnbOffMctcu   = 58
	rnbOffMcttse  = 68
	rnbOffMcnt13  = 78
	rnbOffMctca   = 98
	rnbOffMctbap  = 108
	rnbOffMctcap  = 118
	rnbOffMctcad  = 128
	rnbOffMcntsa  = 133
	rnbOffMcffix  = 143
	rnbOffMfa300  = 153
	rnbOffMfn300  = 163
	rnbOffMcfaba  = 173
	rnbOffMcfnvb  = 183
	rnbOffMcfaca  = 193
	rnbOffMcfnvc  = 203
	rnbOffFiller  = 213
	rnbOffMdgpla  = 223
	rnbOffMdgfap  = 233
	rnbOffMdgfnv  = 243
	rnbOffMdjcom  = 253
	rnbOffMdjsyn  = 263
	rnbOffMdjcu   = 273
	rnbOffMdjtse  = 283
	rnbOffMdjdep  = 293
	rnbOffMdjt13  = 303
	rnbOffTcthfr  = 323
	rnbOffTctfra  = 335
	rnbOffTctdu   = 347
	rnbOffMcnjae  = 359
	rnbOffMfaj1e  = 369
	rnbOffMfnj1e  = 379
	rnbOffMctcas  = 389
	rnbOffMfa800  = 399
	rnbOffMfn800  = 409
	rnbOffMfa900  = 419
	rnbOffMfn900  = 429
	rnbOffMcoge3  = 439
	rnbOffCcoifp  = 449
	rnbOffCcpper  = 452
	rnbOffTrailer = 455

	rnbAmtLen = 10 // PIC S9(10) / 9(10)
	rnbCadLen = 5  // PIC S9(5)
	rnbTotLen = 12 // PIC S9(12)
)

// Compile-time layout checks (a negative or out-of-range index fails to build).
var (
	_ = [1]struct{}{}[rnbOffAidfic+RetNBAidficLen-rnbOffAnupro]
	_ = [1]struct{}{}[rnbOffAnupro+RetNBAnuproLen-rnbOffCcobnb]
	_ = [1]struct{}{}[rnbOffMcnt13+RetNBNt13Count*rnbAmtLen-rnbOffMctca]
	_ = [1]struct{}{}[rnbOffMctcad+rnbCadLen-rnbOffMcntsa]
	_ = [1]struct{}{}[rnbOffMcntsa+rnbAmtLen-rnbOffMcffix]
	_ = [1]struct{}{}[rnbOffMcfnvc+rnbAmtLen-rnbOffFiller]
	_ = [1]struct{}{}[rnbOffFiller+RetNBFillerLen-rnbOffMdgpla]
	_ = [1]struct{}{}[rnbOffMdjt13+RetNBJt13Count*rnbAmtLen-rnbOffTcthfr]
	_ = [1]struct{}{}[rnbOffTctdu+rnbTotLen-rnbOffMcnjae]
	_ = [1]struct{}{}[rnbOffMcoge3+rnbAmtLen-rnbOffCcoifp]
	_ = [1]struct{}{}[rnbOffTrailer+RetNBTrailerLen-RetNBLen]
)

// rnbAmtOffsets lists every scalar PIC S9(10) item of the record (REDEFINES
// aliases excluded: they share their base item's bytes).
var rnbAmtOffsets = [...]int{
	rnbOffMctcom, rnbOffMctdep, rnbOffMctreg, rnbOffMctsyn, rnbOffMctcu,
	rnbOffMcttse, rnbOffMctca, rnbOffMctbap, rnbOffMctcap, rnbOffMcffix,
	rnbOffMfa300, rnbOffMfn300, rnbOffMcfaba, rnbOffMcfnvb, rnbOffMcfaca,
	rnbOffMcfnvc, rnbOffMdgpla, rnbOffMdgfap, rnbOffMdgfnv, rnbOffMdjcom,
	rnbOffMdjsyn, rnbOffMdjcu, rnbOffMdjtse, rnbOffMdjdep, rnbOffMcnjae,
	rnbOffMfaj1e, rnbOffMfnj1e, rnbOffMctcas, rnbOffMfa800, rnbOffMfn800,
	rnbOffMfa900, rnbOffMfn900, rnbOffMcoge3,
}

// RetNB is the 600-byte non-built-property (non bâti) assessment return
// record of copybook XRETNB: file identifier, owner account, per-authority
// contributions (commune, département, région, syndicates, groupings, TSE,
// TSE EPF, chamber of agriculture, BAPSA, CAAA, GEMAPI), assessment /
// non-collection fees, reliefs (meadows and heath, young farmers) and
// totals.
//
// MIGRATION_NOTE: the 'X' placeholder of COPY XRETNB REPLACING disappears:
// each inclusion is a distinct RetNB value. The two 'X'-FILLER items become
// real data names after substitution and are exposed (Filler, Trailer) so
// their content is preserved byte-for-byte.
//
// MIGRATION_NOTE: same encoding conventions as RetB / Comnba (ASCII, zoned
// decimal USAGE DISPLAY, trailing overpunch sign on S9(n) fields; decoding
// accepts IBM and ASCII overpunch, encoding writes IBM). Numeric getters
// return the COBOL NUMERIC class-test result as ok; Raw accessors expose the
// bytes. OCCURS tables use 0-based accessors returning an error when out of
// range. The record has no PIC V field, so no decimal accessor is required.
//
// MIGRATION_NOTE: REDEFINES MCFANO/MCFNVN/MCAANO/MCANVN are same-picture
// aliases of MFA300/MFN300/MFA800/MFN800; they are exposed as alias methods
// reading/writing the same bytes.
//
// MIGRATION_NOTE: a record built from MOVE SPACES (rather than NewRetNB /
// INITIALIZE) has non-NUMERIC numeric fields until they are set; getters
// then report ok=false exactly like the COBOL class test.
type RetNB [RetNBLen]byte

// NewRetNB returns a record in the state produced by INITIALIZE: alphanumeric
// items (including the named 'X'-FILLER areas) are spaces and numeric items
// are zero.
func NewRetNB() RetNB {
	var r RetNB
	for i := range r {
		r[i] = ' '
	}
	r.SetDan(0)
	r.SetNnupro(0)
	for _, off := range rnbAmtOffsets {
		r.setSigned(off, rnbAmtLen, 0)
	}
	for i := 0; i < RetNBNt13Count; i++ {
		r.setSigned(rnbOffMcnt13+i*rnbAmtLen, rnbAmtLen, 0)
	}
	for i := 0; i < RetNBJt13Count; i++ {
		r.setSigned(rnbOffMdjt13+i*rnbAmtLen, rnbAmtLen, 0)
	}
	r.setSigned(rnbOffMctcad, rnbCadLen, 0)
	r.setUnsigned(rnbOffMcntsa, rnbAmtLen, 0)
	for _, off := range []int{rnbOffTcthfr, rnbOffTctfra, rnbOffTctdu} {
		r.setSigned(off, rnbTotLen, 0)
	}
	return r
}

// DecodeRetNB wraps a raw 600-byte buffer as a RetNB record.
func DecodeRetNB(b [RetNBLen]byte) RetNB { return RetNB(b) }

// ParseRetNB copies a buffer into a RetNB; it fails unless b is exactly
// RetNBLen bytes long.
func ParseRetNB(b []byte) (RetNB, error) {
	var r RetNB
	if len(b) != RetNBLen {
		return r, fmt.Errorf("linkage: XRETNB record must be %d bytes, got %d", RetNBLen, len(b))
	}
	copy(r[:], b)
	return r, nil
}

// EncodeRetNB returns the raw 600 bytes of r (MOVE record TO ...).
func EncodeRetNB(r *RetNB) []byte { return r.Bytes() }

// Bytes returns a copy of the raw 600 bytes.
func (r *RetNB) Bytes() []byte {
	out := make([]byte, RetNBLen)
	copy(out, r[:])
	return out
}

// ---- low-level helpers ----

func (r *RetNB) get(off, n int) string { return string(r[off : off+n]) }

// put performs a COBOL alphanumeric MOVE (left-justified, space-padded,
// right-truncated).
func (r *RetNB) put(off, n int, v string) {
	for i := 0; i < n; i++ {
		if i < len(v) {
			r[off+i] = v[i]
		} else {
			r[off+i] = ' '
		}
	}
}

func (r *RetNB) isNumeric(off, n int) bool {
	for _, b := range r[off : off+n] {
		if b < '0' || b > '9' {
			return false
		}
	}
	return true
}

func (r *RetNB) unsigned(off, n int) (uint64, bool) {
	ok := r.isNumeric(off, n)
	var v uint64
	for _, b := range r[off : off+n] {
		v = v*10 + lowDigit(b)
	}
	return v, ok
}

func (r *RetNB) setUnsigned(off, n int, v uint64) {
	for i := off + n - 1; i >= off; i-- {
		r[i] = byte('0' + v%10)
		v /= 10
	}
}

// setUnsignedFromSigned performs a numeric MOVE of a signed value into an
// unsigned PIC 9(n): the sign is dropped and high-order digits truncated.
func (r *RetNB) setUnsignedFromSigned(off, n int, v int64) {
	u := uint64(v)
	if v < 0 {
		u = uint64(-(v + 1)) + 1
	}
	r.setUnsigned(off, n, u)
}

func (r *RetNB) signed(off, n int) (int64, bool) {
	last := off + n - 1
	ok := r.isNumeric(off, n-1)
	var v uint64
	for _, b := range r[off:last] {
		v = v*10 + lowDigit(b)
	}
	d, neg, lastOK := overpunch(r[last])
	v = v*10 + d
	if neg {
		return -int64(v), ok && lastOK
	}
	return int64(v), ok && lastOK
}

// setSigned performs a numeric MOVE into a PIC S9(n) trailing-overpunch
// field: high-order digits beyond n are truncated and a negative value
// truncated to zero is stored as +0.
func (r *RetNB) setSigned(off, n int, v int64) {
	neg := v < 0
	r.setUnsignedFromSigned(off, n, v)
	last := off + n - 1
	d := r[last] - '0'
	allZero := true
	for _, b := range r[off : off+n] {
		if b != '0' {
			allZero = false
			break
		}
	}
	switch {
	case neg && !allZero && d == 0:
		r[last] = '}'
	case neg && !allZero:
		r[last] = 'J' + d - 1
	case d == 0:
		r[last] = '{'
	default:
		r[last] = 'A' + d - 1
	}
}

func rnbCheckIndex(field string, i, count int) error {
	if i < 0 || i >= count {
		return fmt.Errorf("linkage: XRETNB %s subscript %d out of range [0,%d)", field, i, count)
	}
	return nil
}

func (r *RetNB) amt(off int) (int64, bool) { return r.signed(off, rnbAmtLen) }
func (r *RetNB) setAmt(off int, v int64)    { r.setSigned(off, rnbAmtLen, v) }
func (r *RetNB) amtRaw(off int) string      { return r.get(off, rnbAmtLen) }

// ---- identifier ----

// Aidfic returns the raw 11-byte AIDFIC group (file identifier).
func (r *RetNB) Aidfic() string { return r.get(rnbOffAidfic, RetNBAidficLen) }

// SetAidfic moves v into the 11-byte AIDFIC group (group MOVE).
func (r *RetNB) SetAidfic(v string) { r.put(rnbOffAidfic, RetNBAidficLen, v) }

// DanRaw returns the raw 4 bytes of DAN PIC 9(4).
func (r *RetNB) DanRaw() string { return r.get(rnbOffDan, 4) }

// Dan returns DAN PIC 9(4), the tax year; ok is the NUMERIC test.
func (r *RetNB) Dan() (uint64, bool) { return r.unsigned(rnbOffDan, 4) }

// SetDan moves v into DAN PIC 9(4), truncating high-order digits.
func (r *RetNB) SetDan(v uint64) { r.setUnsigned(rnbOffDan, 4, v) }

// Ac3dir returns AC3DIR X(3) (CC2DEP + CCODIR).
func (r *RetNB) Ac3dir() string { return r.get(rnbOffAc3dir, 3) }

// SetAc3dir moves v into AC3DIR X(3).
func (r *RetNB) SetAc3dir(v string) { r.put(rnbOffAc3dir, 3, v) }

// Cc2dep returns CC2DEP X(2), the département code.
func (r *RetNB) Cc2dep() string { return r.get(rnbOffCc2dep, 2) }

// SetCc2dep moves v into CC2DEP X(2).
func (r *RetNB) SetCc2dep(v string) { r.put(rnbOffCc2dep, 2, v) }

// Ccodir returns CCODIR X, the direction code.
func (r *RetNB) Ccodir() string { return r.get(rnbOffCcodir, 1) }

// SetCcodir moves v into CCODIR X.
func (r *RetNB) SetCcodir(v string) { r.put(rnbOffCcodir, 1, v) }

// Ccocom returns CCOCOM X(3), the MAJIC2 commune code.
func (r *RetNB) Ccocom() string { return r.get(rnbOffCcocom, 3) }

// SetCcocom moves v into CCOCOM X(3).
func (r *RetNB) SetCcocom(v string) { r.put(rnbOffCcocom, 3, v) }

// Dsrpar returns DSRPAR X, the parcel roll series.
func (r *RetNB) Dsrpar() string { return r.get(rnbOffDsrpar, 1) }

// SetDsrpar moves v into DSRPAR X.
func (r *RetNB) SetDsrpar(v string) { r.put(rnbOffDsrpar, 1, v) }

// Anupro returns the raw 6-byte ANUPRO group (owner account).
func (r *RetNB) Anupro() string { return r.get(rnbOffAnupro, RetNBAnuproLen) }

// SetAnupro moves v into the 6-byte ANUPRO group (group MOVE).
func (r *RetNB) SetAnupro(v string) { r.put(rnbOffAnupro, RetNBAnuproLen, v) }

// Cgroup returns CGROUP X, the owner group code.
func (r *RetNB) Cgroup() string { return r.get(rnbOffCgroup, 1) }

// SetCgroup moves v into CGROUP X.
func (r *RetNB) SetCgroup(v string) { r.put(rnbOffCgroup, 1, v) }

// NnuproRaw returns the raw 5 bytes of NNUPRO PIC 9(5).
func (r *RetNB) NnuproRaw() string { return r.get(rnbOffNnupro, 5) }

// Nnupro returns NNUPRO PIC 9(5), the owner account number; ok is the
// NUMERIC test.
func (r *RetNB) Nnupro() (uint64, bool) { return r.unsigned(rnbOffNnupro, 5) }

// SetNnupro moves v into NNUPRO PIC 9(5), truncating high-order digits.
func (r *RetNB) SetNnupro(v uint64) { r.setUnsigned(rnbOffNnupro, 5, v) }

// Ccobnb returns CCOBNB X, the built / non-built code.
func (r *RetNB) Ccobnb() string { return r.get(rnbOffCcobnb, 1) }

// SetCcobnb moves v into CCOBNB X.
func (r *RetNB) SetCcobnb(v string) { r.put(rnbOffCcobnb, 1, v) }

// ---- contributions, PIC S9(10) ----
// Setters perform a numeric MOVE: high-order digits beyond 10 are truncated.

// Mctcom returns MCTCOM S9(10), the commune contribution; ok is NUMERIC.
func (r *RetNB) Mctcom() (int64, bool) { return r.amt(rnbOffMctcom) }

// SetMctcom moves v into MCTCOM S9(10).
func (r *RetNB) SetMctcom(v int64) { r.setAmt(rnbOffMctcom, v) }

// MctcomRaw returns the raw 10 bytes of MCTCOM.
func (r *RetNB) MctcomRaw() string { return r.amtRaw(rnbOffMctcom) }

// Mctdep returns MCTDEP S9(10), the département contribution.
func (r *RetNB) Mctdep() (int64, bool) { return r.amt(rnbOffMctdep) }

// SetMctdep moves v into MCTDEP S9(10).
func (r *RetNB) SetMctdep(v int64) { r.setAmt(rnbOffMctdep, v) }

// MctdepRaw returns the raw 10 bytes of MCTDEP.
func (r *RetNB) MctdepRaw() string { return r.amtRaw(rnbOffMctdep) }

// Mctreg returns MCTREG S9(10), the region contribution.
func (r *RetNB) Mctreg() (int64, bool) { return r.amt(rnbOffMctreg) }

// SetMctreg moves v into MCTREG S9(10).
func (r *RetNB) SetMctreg(v int64) { r.setAmt(rnbOffMctreg, v) }

// MctregRaw returns the raw 10 bytes of MCTREG.
func (r *RetNB) MctregRaw() string { return r.amtRaw(rnbOffMctreg) }

// Mctsyn returns MCTSYN S9(10), the syndicates / districts without own
// taxation contribution.
func (r *RetNB) Mctsyn() (int64, bool) { return r.amt(rnbOffMctsyn) }

// SetMctsyn moves v into MCTSYN S9(10).
func (r *RetNB) SetMctsyn(v int64) { r.setAmt(rnbOffMctsyn, v) }

// MctsynRaw returns the raw 10 bytes of MCTSYN.
func (r *RetNB) MctsynRaw() string { return r.amtRaw(rnbOffMctsyn) }

// Mctcu returns MCTCU S9(10), the grouping-with-own-taxation contribution.
func (r *RetNB) Mctcu() (int64, bool) { return r.amt(rnbOffMctcu) }

// SetMctcu moves v into MCTCU S9(10).
func (r *RetNB) SetMctcu(v int64) { r.setAmt(rnbOffMctcu, v) }

// MctcuRaw returns the raw 10 bytes of MCTCU.
func (r *RetNB) MctcuRaw() string { return r.amtRaw(rnbOffMctcu) }

// Mcttse returns MCTTSE S9(10), the special equipment tax contribution.
func (r *RetNB) Mcttse() (int64, bool) { return r.amt(rnbOffMcttse) }

// SetMcttse moves v into MCTTSE S9(10).
func (r *RetNB) SetMcttse(v int64) { r.setAmt(rnbOffMcttse, v) }

// McttseRaw returns the raw 10 bytes of MCTTSE.
func (r *RetNB) McttseRaw() string { return r.amtRaw(rnbOffMcttse) }

// Mcnt13 returns MCNT13(i+1) S9(10) (0-based i), the non-built TSE EPF
// contribution.
func (r *RetNB) Mcnt13(i int) (int64, bool, error) {
	if err := rnbCheckIndex("MCNT13", i, RetNBNt13Count); err != nil {
		return 0, false, err
	}
	v, ok := r.amt(rnbOffMcnt13 + i*rnbAmtLen)
	return v, ok, nil
}

// SetMcnt13 moves v into MCNT13(i+1) S9(10) (0-based i).
func (r *RetNB) SetMcnt13(i int, v int64) error {
	if err := rnbCheckIndex("MCNT13", i, RetNBNt13Count); err != nil {
		return err
	}
	r.setAmt(rnbOffMcnt13+i*rnbAmtLen, v)
	return nil
}

// Mcnt13Raw returns the raw 10 bytes of MCNT13(i+1) (0-based i).
func (r *RetNB) Mcnt13Raw(i int) (string, error) {
	if err := rnbCheckIndex("MCNT13", i, RetNBNt13Count); err != nil {
		return "", err
	}
	return r.amtRaw(rnbOffMcnt13 + i*rnbAmtLen), nil
}

// Mctca returns MCTCA S9(10), the chamber of agriculture contribution.
func (r *RetNB) Mctca() (int64, bool) { return r.amt(rnbOffMctca) }

// SetMctca moves v into MCTCA S9(10).
func (r *RetNB) SetMctca(v int64) { r.setAmt(rnbOffMctca, v) }

// MctcaRaw returns the raw 10 bytes of MCTCA.
func (r *RetNB) MctcaRaw() string { return r.amtRaw(rnbOffMctca) }

// Mctbap returns MCTBAP S9(10), the BAPSA contribution.
func (r *RetNB) Mctbap() (int64, bool) { return r.amt(rnbOffMctbap) }

// SetMctbap moves v into MCTBAP S9(10).
func (r *RetNB) SetMctbap(v int64) { r.setAmt(rnbOffMctbap, v) }

// MctbapRaw returns the raw 10 bytes of MCTBAP.
func (r *RetNB) MctbapRaw() string { return r.amtRaw(rnbOffMctbap) }

// Mctcap returns MCTCAP S9(10), the CAAA proportional duty contribution.
func (r *RetNB) Mctcap() (int64, bool) { return r.amt(rnbOffMctcap) }

// SetMctcap moves v into MCTCAP S9(10).
func (r *RetNB) SetMctcap(v int64) { r.setAmt(rnbOffMctcap, v) }

// MctcapRaw returns the raw 10 bytes of MCTCAP.
func (r *RetNB) MctcapRaw() string { return r.amtRaw(rnbOffMctcap) }

// Mctcad returns MCTCAD S9(5), the CAAA fixed duty contribution.
func (r *RetNB) Mctcad() (int64, bool) { return r.signed(rnbOffMctcad, rnbCadLen) }

// SetMctcad moves v into MCTCAD S9(5): high-order digits beyond 5 are
// truncated, as a COBOL numeric MOVE does.
func (r *RetNB) SetMctcad(v int64) { r.setSigned(rnbOffMctcad, rnbCadLen, v) }

// MctcadRaw returns the raw 5 bytes of MCTCAD.
func (r *RetNB) MctcadRaw() string { return r.get(rnbOffMctcad, rnbCadLen) }

// Mcntsa returns MCNTSA PIC 9(10) (unsigned), the amalgamated non-built
// new TSE contribution; ok is the NUMERIC test.
func (r *RetNB) Mcntsa() (uint64, bool) { return r.unsigned(rnbOffMcntsa, rnbAmtLen) }

// SetMcntsa moves v into MCNTSA PIC 9(10): the picture is unsigned, so the
// sign is dropped (a negative value is stored as its absolute value) and
// high-order digits are truncated, as a COBOL MOVE does.
func (r *RetNB) SetMcntsa(v int64) { r.setUnsignedFromSigned(rnbOffMcntsa, rnbAmtLen, v) }

// McntsaRaw returns the raw 10 bytes of MCNTSA.
func (r *RetNB) McntsaRaw() string { return r.amtRaw(rnbOffMcntsa) }

// ---- fees (with REDEFINES aliases) ----

// Mcffix returns MCFFIX S9(10), the CAAA fixed fees.
func (r *RetNB) Mcffix() (int64, bool) { return r.amt(rnbOffMcffix) }

// SetMcffix moves v into MCFFIX S9(10).
func (r *RetNB) SetMcffix(v int64) { r.setAmt(rnbOffMcffix, v) }

// Mfa300 returns MFA300 S9(10), assessment fees at 3%.
func (r *RetNB) Mfa300() (int64, bool) { return r.amt(rnbOffMfa300) }

// SetMfa300 moves v into MFA300 S9(10).
func (r *RetNB) SetMfa300(v int64) { r.setAmt(rnbOffMfa300, v) }

// Mcfano returns MCFANO (REDEFINES MFA300), normal assessment fees.
func (r *RetNB) Mcfano() (int64, bool) { return r.Mfa300() }

// SetMcfano moves v into MCFANO (same bytes as MFA300).
func (r *RetNB) SetMcfano(v int64) { r.SetMfa300(v) }

// Mfn300 returns MFN300 S9(10), non-collection fees at 3%.
func (r *RetNB) Mfn300() (int64, bool) { return r.amt(rnbOffMfn300) }

// SetMfn300 moves v into MFN300 S9(10).
func (r *RetNB) SetMfn300(v int64) { r.setAmt(rnbOffMfn300, v) }

// Mcfnvn returns MCFNVN (REDEFINES MFN300), normal non-collection fees.
func (r *RetNB) Mcfnvn() (int64, bool) { return r.Mfn300() }

// SetMcfnvn moves v into MCFNVN (same bytes as MFN300).
func (r *RetNB) SetMcfnvn(v int64) { r.SetMfn300(v) }

// Mcfaba returns MCFABA S9(10), BAPSA assessment fees.
func (r *RetNB) Mcfaba() (int64, bool) { return r.amt(rnbOffMcfaba) }

// SetMcfaba moves v into MCFABA S9(10).
func (r *RetNB) SetMcfaba(v int64) { r.setAmt(rnbOffMcfaba, v) }

// Mcfnvb returns MCFNVB S9(10), BAPSA non-collection fees.
func (r *RetNB) Mcfnvb() (int64, bool) { return r.amt(rnbOffMcfnvb) }

// SetMcfnvb moves v into MCFNVB S9(10).
func (r *RetNB) SetMcfnvb(v int64) { r.setAmt(rnbOffMcfnvb, v) }

// Mcfaca returns MCFACA S9(10), CAAA assessment fees.
func (r *RetNB) Mcfaca() (int64, bool) { return r.amt(rnbOffMcfaca) }

// SetMcfaca moves v into MCFACA S9(10).
func (r *RetNB) SetMcfaca(v int64) { r.setAmt(rnbOffMcfaca, v) }

// Mcfnvc returns MCFNVC S9(10), CAAA non-collection fees.
func (r *RetNB) Mcfnvc() (int64, bool) { return r.amt(rnbOffMcfnvc) }

// SetMcfnvc moves v into MCFNVC S9(10).
func (r *RetNB) SetMcfnvc(v int64) { r.setAmt(rnbOffMcfnvc, v) }

// Filler returns the X(10) 'X'-FILLER area following MCFNVC.
func (r *RetNB) Filler() string { return r.get(rnbOffFiller, RetNBFillerLen) }

// SetFiller moves v into the X(10) 'X'-FILLER area.
func (r *RetNB) SetFiller(v string) { r.put(rnbOffFiller, RetNBFillerLen, v) }

// ---- reliefs (dégrèvements) ----

// Mdgpla returns MDGPLA S9(10), the meadows and heath relief amount.
func (r *RetNB) Mdgpla() (int64, bool) { return r.amt(rnbOffMdgpla) }

// SetMdgpla moves v into MDGPLA S9(10).
func (r *RetNB) SetMdgpla(v int64) { r.setAmt(rnbOffMdgpla, v) }

// Mdgfap returns MDGFAP S9(10), the meadows relief on assessment fees.
func (r *RetNB) Mdgfap() (int64, bool) { return r.amt(rnbOffMdgfap) }

// SetMdgfap moves v into MDGFAP S9(10).
func (r *RetNB) SetMdgfap(v int64) { r.setAmt(rnbOffMdgfap, v) }

// Mdgfnv returns MDGFNV S9(10), the meadows relief on non-collection fees.
func (r *RetNB) Mdgfnv() (int64, bool) { return r.amt(rnbOffMdgfnv) }

// SetMdgfnv moves v into MDGFNV S9(10).
func (r *RetNB) SetMdgfnv(v int64) { r.setAmt(rnbOffMdgfnv, v) }

// Mdjcom returns MDJCOM S9(10), the young-farmer relief, commune share.
func (r *RetNB) Mdjcom() (int64, bool) { return r.amt(rnbOffMdjcom) }

// SetMdjcom moves v into MDJCOM S9(10).
func (r *RetNB) SetMdjcom(v int64) { r.setAmt(rnbOffMdjcom, v) }

// Mdjsyn returns MDJSYN S9(10), the young-farmer relief, syndicate share.
func (r *RetNB) Mdjsyn() (int64, bool) { return r.amt(rnbOffMdjsyn) }

// SetMdjsyn moves v into MDJSYN S9(10).
func (r *RetNB) SetMdjsyn(v int64) { r.setAmt(rnbOffMdjsyn, v) }

// Mdjcu returns MDJCU S9(10), the young-farmer relief, grouping share.
func (r *RetNB) Mdjcu() (int64, bool) { return r.amt(rnbOffMdjcu) }

// SetMdjcu moves v into MDJCU S9(10).
func (r *RetNB) SetMdjcu(v int64) { r.setAmt(rnbOffMdjcu, v) }

// Mdjtse returns MDJTSE S9(10), the young-farmer relief, TSE share.
func (r *RetNB) Mdjtse() (int64, bool) { return r.amt(rnbOffMdjtse) }

// SetMdjtse moves v into MDJTSE S9(10).
func (r *RetNB) SetMdjtse(v int64) { r.setAmt(rnbOffMdjtse, v) }

// Mdjdep returns MDJDEP S9(10), the young-farmer relief, département share.
func (r *RetNB) Mdjdep() (int64, bool) { return r.amt(rnbOffMdjdep) }

// SetMdjdep moves v into MDJDEP S9(10).
func (r *RetNB) SetMdjdep(v int64) { r.setAmt(rnbOffMdjdep, v) }

// Mdjt13 returns MDJT13(i+1) S9(10) (0-based i), the young-farmer relief,
// TSE EPF share.
func (r *RetNB) Mdjt13(i int) (int64, bool, error) {
	if err := rnbCheckIndex("MDJT13", i, RetNBJt13Count); err != nil {
		return 0, false, err
	}
	v, ok := r.amt(rnbOffMdjt13 + i*rnbAmtLen)
	return v, ok, nil
}

// SetMdjt13 moves v into MDJT13(i+1) S9(10) (0-based i).
func (r *RetNB) SetMdjt13(i int, v int64) error {
	if err := rnbCheckIndex("MDJT13", i, RetNBJt13Count); err != nil {
		return err
	}
	r.setAmt(rnbOffMdjt13+i*rnbAmtLen, v)
	return nil
}

// Mdjt13Raw returns the raw 10 bytes of MDJT13(i+1) (0-based i).
func (r *RetNB) Mdjt13Raw(i int) (string, error) {
	if err := rnbCheckIndex("MDJT13", i, RetNBJt13Count); err != nil {
		return "", err
	}
	return r.amtRaw(rnbOffMdjt13 + i*rnbAmtLen), nil
}

// ---- totals, PIC S9(12) ----

// Tcthfr returns TCTHFR S9(12), total contributions excluding fees.
func (r *RetNB) Tcthfr() (int64, bool) { return r.signed(rnbOffTcthfr, rnbTotLen) }

// SetTcthfr moves v into TCTHFR S9(12), truncating high-order digits.
func (r *RetNB) SetTcthfr(v int64) { r.setSigned(rnbOffTcthfr, rnbTotLen, v) }

// Tctfra returns TCTFRA S9(12), total fees.
func (r *RetNB) Tctfra() (int64, bool) { return r.signed(rnbOffTctfra, rnbTotLen) }

// SetTctfra moves v into TCTFRA S9(12), truncating high-order digits.
func (r *RetNB) SetTctfra(v int64) { r.setSigned(rnbOffTctfra, rnbTotLen, v) }

// Tctdu returns TCTDU S9(12), total due including fees.
func (r *RetNB) Tctdu() (int64, bool) { return r.signed(rnbOffTctdu, rnbTotLen) }

// SetTctdu moves v into TCTDU S9(12), truncating high-order digits.
func (r *RetNB) SetTctdu(v int64) { r.setSigned(rnbOffTctdu, rnbTotLen, v) }

// ---- young farmers (State share), Saint-Martin, 8% / 9% fees ----

// Mcnjae returns MCNJAE S9(10), the total young-farmer contribution borne
// by the State.
func (r *RetNB) Mcnjae() (int64, bool) { return r.amt(rnbOffMcnjae) }

// SetMcnjae moves v into MCNJAE S9(10).
func (r *RetNB) SetMcnjae(v int64) { r.setAmt(rnbOffMcnjae, v) }

// Mfaj1e returns MFAJ1E S9(10), young-farmer State assessment fees.
func (r *RetNB) Mfaj1e() (int64, bool) { return r.amt(rnbOffMfaj1e) }

// SetMfaj1e moves v into MFAJ1E S9(10).
func (r *RetNB) SetMfaj1e(v int64) { r.setAmt(rnbOffMfaj1e, v) }

// Mfnj1e returns MFNJ1E S9(10), young-farmer State non-collection fees.
func (r *RetNB) Mfnj1e() (int64, bool) { return r.amt(rnbOffMfnj1e) }

// SetMfnj1e moves v into MFNJ1E S9(10).
func (r *RetNB) SetMfnj1e(v int64) { r.setAmt(rnbOffMfnj1e, v) }

// Mctcas returns MCTCAS S9(10), the Saint-Martin chamber of agriculture
// contribution.
func (r *RetNB) Mctcas() (int64, bool) { return r.amt(rnbOffMctcas) }

// SetMctcas moves v into MCTCAS S9(10).
func (r *RetNB) SetMctcas(v int64) { r.setAmt(rnbOffMctcas, v) }

// Mfa800 returns MFA800 S9(10), assessment fees at 8%.
func (r *RetNB) Mfa800() (int64, bool) { return r.amt(rnbOffMfa800) }

// SetMfa800 moves v into MFA800 S9(10).
func (r *RetNB) SetMfa800(v int64) { r.setAmt(rnbOffMfa800, v) }

// Mcaano returns MCAANO (REDEFINES MFA800), chamber of agriculture normal
// assessment fees.
func (r *RetNB) Mcaano() (int64, bool) { return r.Mfa800() }

// SetMcaano moves v into MCAANO (same bytes as MFA800).
func (r *RetNB) SetMcaano(v int64) { r.SetMfa800(v) }

// Mfn800 returns MFN800 S9(10), non-collection fees at 8%.
func (r *RetNB) Mfn800() (int64, bool) { return r.amt(rnbOffMfn800) }

// SetMfn800 moves v into MFN800 S9(10).
func (r *RetNB) SetMfn800(v int64) { r.setAmt(rnbOffMfn800, v) }

// Mcanvn returns MCANVN (REDEFINES MFN800), chamber of agriculture
// non-collection fees.
func (r *RetNB) Mcanvn() (int64, bool) { return r.Mfn800() }

// SetMcanvn moves v into MCANVN (same bytes as MFN800).
func (r *RetNB) SetMcanvn(v int64) { r.SetMfn800(v) }

// Mfa900 returns MFA900 S9(10), assessment fees at 9%.
func (r *RetNB) Mfa900() (int64, bool) { return r.amt(rnbOffMfa900) }

// SetMfa900 moves v into MFA900 S9(10).
func (r *RetNB) SetMfa900(v int64) { r.setAmt(rnbOffMfa900, v) }

// Mfn900 returns MFN900 S9(10), non-collection fees at 9%.
func (r *RetNB) Mfn900() (int64, bool) { return r.amt(rnbOffMfn900) }

// SetMfn900 moves v into MFN900 S9(10).
func (r *RetNB) SetMfn900(v int64) { r.setAmt(rnbOffMfn900, v) }

// Mcoge3 returns MCOGE3 S9(10), the GEMAPI contribution.
func (r *RetNB) Mcoge3() (int64, bool) { return r.amt(rnbOffMcoge3) }

// SetMcoge3 moves v into MCOGE3 S9(10).
func (r *RetNB) SetMcoge3(v int64) { r.setAmt(rnbOffMcoge3, v) }

// ---- codes, trailer ----

// Ccoifp returns CCOIFP X(3), the absorbed-commune code.
func (r *RetNB) Ccoifp() string { return r.get(rnbOffCcoifp, 3) }

// SetCcoifp moves v into CCOIFP X(3).
func (r *RetNB) SetCcoifp(v string) { r.put(rnbOffCcoifp, 3, v) }

// Ccpper returns CCPPER X(3), the treasury code.
func (r *RetNB) Ccpper() string { return r.get(rnbOffCcpper, 3) }

// SetCcpper moves v into CCPPER X(3).
func (r *RetNB) SetCcpper(v string) { r.put(rnbOffCcpper, 3, v) }

// Trailer returns the trailing X(145) 'X'-FILLER area.
func (r *RetNB) Trailer() string { return r.get(rnbOffTrailer, RetNBTrailerLen) }

// SetTrailer moves v into the trailing X(145) 'X'-FILLER area.
func (r *RetNB) SetTrailer(v string) { r.put(rnbOffTrailer, RetNBTrailerLen, v) }
