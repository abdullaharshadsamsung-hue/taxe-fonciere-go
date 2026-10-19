package linkage

import "fmt"

// RetBLen is the total length in bytes of the built-property assessment
// return record (copybook XRETB).
const RetBLen = 600

// OCCURS counts of the XRETB record.
const (
	// RetBBt13Count is the OCCURS count of MCBT13.
	RetBBt13Count = 2
	// RetBOmCount is the OCCURS count of the ACTOM group (TEOM zones).
	RetBOmCount = 6
)

// Group / filler lengths of the XRETB record.
const (
	// RetBAidficLen is the length of the AIDFIC group.
	RetBAidficLen = 11
	// RetBAnuproLen is the length of the ANUPRO group.
	RetBAnuproLen = 6
	// RetBFillerLen is the length of the first 'X'-FILLER X(40).
	RetBFillerLen = 40
	// RetBTrailerLen is the length of the trailing 'X'-FILLER X(238).
	RetBTrailerLen = 238
)

// Zero-based offsets inside the 600-byte XRETB record.
//
//	AIDFIC   X(11)      0..10   DAN 9(4) 0..3, AC3DIR 4..6 (CC2DEP 4..5,
//	                            CCODIR 6), CCOCOM 7..9, DSRPAR 10
//	ANUPRO   X(6)      11..16   CGROUP 11, NNUPRO 9(5) 12..16
//	CCOBNB   X         17
//	MCTCOM 18, MCTDEP 28, MCTREG 38, MCTSYN 48, MCTCU 58, MCTTSE 68 (S9(10))
//	MCBT13   S9(10) OCCURS 2   78..97
//	MCBTSA   9(10)             98..107
//	FILLER   X(40)            108..147
//	ACTOM    OCCURS 6 x 12    148..219  GTAUOM X(2) +0, MCTOM S9(10) +2
//	MFA300 220 (=MCFANO), MFN300 230 (=MCFNVN), MFA800 240 (=MOMANO),
//	MFN800 250 (=MOMNVN)      (S9(10))
//	TCTHFR 260, TCTFRA 272, TCTDU 284 (S9(12))
//	MFA900 296, MFN900 306 (S9(10)), TCTOM 316 (9(10)),
//	MVLTIM 326, MCOGE3 336, MCOTA3 346 (S9(10))
//	CCOIFP   X(3)             356..358
//	CCPPER   X(3)             359..361
//	FILLER   X(238)           362..599
const (
	rbOffAidfic  = 0
	rbOffDan     = 0
	rbOffAc3dir  = 4
	rbOffCc2dep  = 4
	rbOffCcodir  = 6
	rbOffCcocom  = 7
	rbOffDsrpar  = 10
	rbOffAnupro  = 11
	rbOffCgroup  = 11
	rbOffNnupro  = 12
	rbOffCcobnb  = 17
	rbOffMctcom  = 18
	rbOffMctdep  = 28
	rbOffMctreg  = 38
	rbOffMctsyn  = 48
	rbOffMctcu   = 58
	rbOffMcttse  = 68
	rbOffMcbt13  = 78
	rbOffMcbtsa  = 98
	rbOffFiller  = 108
	rbOffActom   = 148
	rbOffMfa300  = 220
	rbOffMfn300  = 230
	rbOffMfa800  = 240
	rbOffMfn800  = 250
	rbOffTcthfr  = 260
	rbOffTctfra  = 272
	rbOffTctdu   = 284
	rbOffMfa900  = 296
	rbOffMfn900  = 306
	rbOffTctom   = 316
	rbOffMvltim  = 326
	rbOffMcoge3  = 336
	rbOffMcota3  = 346
	rbOffCcoifp  = 356
	rbOffCcpper  = 359
	rbOffTrailer = 362

	rbAmtLen = 10 // PIC S9(10) / 9(10)
	rbTotLen = 12 // PIC S9(12)
	rbOmLen  = 12 // ACTOM occurrence: GTAUOM X(2) + MCTOM S9(10)
)

// Compile-time layout checks (a negative or out-of-range index fails to build).
var (
	_ = [1]struct{}{}[rbOffAidfic+RetBAidficLen-rbOffAnupro]
	_ = [1]struct{}{}[rbOffAnupro+RetBAnuproLen-rbOffCcobnb]
	_ = [1]struct{}{}[rbOffMcbt13+RetBBt13Count*rbAmtLen-rbOffMcbtsa]
	_ = [1]struct{}{}[rbOffFiller+RetBFillerLen-rbOffActom]
	_ = [1]struct{}{}[rbOffActom+RetBOmCount*rbOmLen-rbOffMfa300]
	_ = [1]struct{}{}[rbOffTcthfr+rbTotLen-rbOffTctfra]
	_ = [1]struct{}{}[rbOffTctdu+rbTotLen-rbOffMfa900]
	_ = [1]struct{}{}[rbOffMcota3+rbAmtLen-rbOffCcoifp]
	_ = [1]struct{}{}[rbOffTrailer+RetBTrailerLen-RetBLen]
)

// RetB is the 600-byte built-property (bâti) assessment return record of
// copybook XRETB, included by EFITA3B8 as RETOURB: file identifier, owner
// account, per-authority contributions (commune, département, région,
// syndicates, EPCI, TSE, TSE EPF, TEOM zones, TEOMI, GEMAPI, TASA),
// assessment / non-collection fees and totals.
//
// MIGRATION_NOTE: the 'X' placeholder of COPY XRETB REPLACING 'X' BY
// RETOURB disappears: each inclusion is a distinct RetB value. The two
// 'X'-FILLER items become real data names after substitution and are
// exposed (Filler, Trailer) so their content is preserved byte-for-byte.
//
// MIGRATION_NOTE: same encoding conventions as Combat / Comnba (ASCII,
// zoned decimal USAGE DISPLAY, trailing overpunch sign on S9(n) fields;
// decoding accepts IBM and ASCII overpunch, encoding writes IBM). Numeric
// getters return the COBOL NUMERIC class-test result; Raw accessors expose
// the bytes. OCCURS tables use 0-based accessors returning an error when
// out of range. The record has no PIC V field, so no decimal accessor is
// required.
//
// MIGRATION_NOTE: REDEFINES MCFANO/MCFNVN/MOMANO/MOMNVN are same-picture
// aliases of MFA300/MFN300/MFA800/MFN800; they are exposed as alias methods
// reading/writing the same bytes.
type RetB [RetBLen]byte

// NewRetB returns a record in the state produced by INITIALIZE RETOURB:
// alphanumeric items (including the named 'X'-FILLER areas) are spaces and
// numeric items are zero.
func NewRetB() RetB {
	var r RetB
	for i := range r {
		r[i] = ' '
	}
	r.SetDan(0)
	r.SetNnupro(0)
	for _, off := range []int{rbOffMctcom, rbOffMctdep, rbOffMctreg, rbOffMctsyn,
		rbOffMctcu, rbOffMcttse, rbOffMfa300, rbOffMfn300, rbOffMfa800,
		rbOffMfn800, rbOffMfa900, rbOffMfn900, rbOffMvltim, rbOffMcoge3, rbOffMcota3} {
		r.setSigned(off, rbAmtLen, 0)
	}
	for i := 0; i < RetBBt13Count; i++ {
		r.setSigned(rbOffMcbt13+i*rbAmtLen, rbAmtLen, 0)
	}
	for i := 0; i < RetBOmCount; i++ {
		r.setSigned(rbOffActom+i*rbOmLen+2, rbAmtLen, 0)
	}
	r.setUnsigned(rbOffMcbtsa, rbAmtLen, 0)
	r.setUnsigned(rbOffTctom, rbAmtLen, 0)
	for _, off := range []int{rbOffTcthfr, rbOffTctfra, rbOffTctdu} {
		r.setSigned(off, rbTotLen, 0)
	}
	return r
}

// DecodeRetB wraps a raw 600-byte buffer as a RetB record.
func DecodeRetB(b [RetBLen]byte) RetB { return RetB(b) }

// ParseRetB copies a buffer into a RetB; it fails unless b is exactly
// RetBLen bytes long.
func ParseRetB(b []byte) (RetB, error) {
	var r RetB
	if len(b) != RetBLen {
		return r, fmt.Errorf("linkage: RETOURB (XRETB) must be %d bytes, got %d", RetBLen, len(b))
	}
	copy(r[:], b)
	return r, nil
}

// Bytes returns a copy of the raw 600 bytes (MOVE RETOURB TO ...).
func (r *RetB) Bytes() []byte {
	out := make([]byte, RetBLen)
	copy(out, r[:])
	return out
}

// ---- low-level helpers ----

func (r *RetB) get(off, n int) string { return string(r[off : off+n]) }

// put performs a COBOL alphanumeric MOVE (left-justified, space-padded,
// right-truncated).
func (r *RetB) put(off, n int, v string) {
	for i := 0; i < n; i++ {
		if i < len(v) {
			r[off+i] = v[i]
		} else {
			r[off+i] = ' '
		}
	}
}

func (r *RetB) isNumeric(off, n int) bool {
	for _, b := range r[off : off+n] {
		if b < '0' || b > '9' {
			return false
		}
	}
	return true
}

func (r *RetB) unsigned(off, n int) (uint64, bool) {
	ok := r.isNumeric(off, n)
	var v uint64
	for _, b := range r[off : off+n] {
		v = v*10 + lowDigit(b)
	}
	return v, ok
}

func (r *RetB) setUnsigned(off, n int, v uint64) {
	for i := off + n - 1; i >= off; i-- {
		r[i] = byte('0' + v%10)
		v /= 10
	}
}

// setUnsignedFromSigned performs a numeric MOVE of a signed value into an
// unsigned PIC 9(n): the sign is dropped and high-order digits truncated.
func (r *RetB) setUnsignedFromSigned(off, n int, v int64) {
	u := uint64(v)
	if v < 0 {
		u = uint64(-(v + 1)) + 1
	}
	r.setUnsigned(off, n, u)
}

func (r *RetB) signed(off, n int) (int64, bool) {
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
func (r *RetB) setSigned(off, n int, v int64) {
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

func rbCheckIndex(field string, i, count int) error {
	if i < 0 || i >= count {
		return fmt.Errorf("linkage: RETOURB %s subscript %d out of range [0,%d)", field, i, count)
	}
	return nil
}

func (r *RetB) amt(off int) (int64, bool) { return r.signed(off, rbAmtLen) }
func (r *RetB) setAmt(off int, v int64)    { r.setSigned(off, rbAmtLen, v) }
func (r *RetB) amtRaw(off int) string      { return r.get(off, rbAmtLen) }

// ---- identifier ----

// Aidfic returns the raw 11-byte AIDFIC group (file identifier).
func (r *RetB) Aidfic() string { return r.get(rbOffAidfic, RetBAidficLen) }

// SetAidfic moves v into the 11-byte AIDFIC group (group MOVE).
func (r *RetB) SetAidfic(v string) { r.put(rbOffAidfic, RetBAidficLen, v) }

// DanRaw returns the raw 4 bytes of DAN PIC 9(4).
func (r *RetB) DanRaw() string { return r.get(rbOffDan, 4) }

// Dan returns DAN PIC 9(4), the tax year; ok is the NUMERIC test.
func (r *RetB) Dan() (uint64, bool) { return r.unsigned(rbOffDan, 4) }

// SetDan moves v into DAN PIC 9(4), truncating high-order digits.
func (r *RetB) SetDan(v uint64) { r.setUnsigned(rbOffDan, 4, v) }

// Ac3dir returns AC3DIR X(3) (CC2DEP + CCODIR).
func (r *RetB) Ac3dir() string { return r.get(rbOffAc3dir, 3) }

// SetAc3dir moves v into AC3DIR X(3).
func (r *RetB) SetAc3dir(v string) { r.put(rbOffAc3dir, 3, v) }

// Cc2dep returns CC2DEP X(2), the département code.
func (r *RetB) Cc2dep() string { return r.get(rbOffCc2dep, 2) }

// SetCc2dep moves v into CC2DEP X(2).
func (r *RetB) SetCc2dep(v string) { r.put(rbOffCc2dep, 2, v) }

// Ccodir returns CCODIR X, the direction code.
func (r *RetB) Ccodir() string { return r.get(rbOffCcodir, 1) }

// SetCcodir moves v into CCODIR X.
func (r *RetB) SetCcodir(v string) { r.put(rbOffCcodir, 1, v) }

// Ccocom returns CCOCOM X(3), the MAJIC2 commune code.
func (r *RetB) Ccocom() string { return r.get(rbOffCcocom, 3) }

// SetCcocom moves v into CCOCOM X(3).
func (r *RetB) SetCcocom(v string) { r.put(rbOffCcocom, 3, v) }

// Dsrpar returns DSRPAR X, the parcel roll series.
func (r *RetB) Dsrpar() string { return r.get(rbOffDsrpar, 1) }

// SetDsrpar moves v into DSRPAR X.
func (r *RetB) SetDsrpar(v string) { r.put(rbOffDsrpar, 1, v) }

// Anupro returns the raw 6-byte ANUPRO group (owner account).
func (r *RetB) Anupro() string { return r.get(rbOffAnupro, RetBAnuproLen) }

// SetAnupro moves v into the 6-byte ANUPRO group (group MOVE).
func (r *RetB) SetAnupro(v string) { r.put(rbOffAnupro, RetBAnuproLen, v) }

// Cgroup returns CGROUP X, the owner group code.
func (r *RetB) Cgroup() string { return r.get(rbOffCgroup, 1) }

// SetCgroup moves v into CGROUP X.
func (r *RetB) SetCgroup(v string) { r.put(rbOffCgroup, 1, v) }

// NnuproRaw returns the raw 5 bytes of NNUPRO PIC 9(5).
func (r *RetB) NnuproRaw() string { return r.get(rbOffNnupro, 5) }

// Nnupro returns NNUPRO PIC 9(5), the owner account number; ok is the
// NUMERIC test.
func (r *RetB) Nnupro() (uint64, bool) { return r.unsigned(rbOffNnupro, 5) }

// SetNnupro moves v into NNUPRO PIC 9(5), truncating high-order digits.
func (r *RetB) SetNnupro(v uint64) { r.setUnsigned(rbOffNnupro, 5, v) }

// Ccobnb returns CCOBNB X, the built / non-built code.
func (r *RetB) Ccobnb() string { return r.get(rbOffCcobnb, 1) }

// SetCcobnb moves v into CCOBNB X.
func (r *RetB) SetCcobnb(v string) { r.put(rbOffCcobnb, 1, v) }

// ---- contributions, PIC S9(10) ----
// Setters perform a numeric MOVE: high-order digits beyond 10 are truncated.

// Mctcom returns MCTCOM S9(10), the commune contribution; ok is NUMERIC.
func (r *RetB) Mctcom() (int64, bool) { return r.amt(rbOffMctcom) }

// SetMctcom moves v into MCTCOM S9(10).
func (r *RetB) SetMctcom(v int64) { r.setAmt(rbOffMctcom, v) }

// MctcomRaw returns the raw 10 bytes of MCTCOM.
func (r *RetB) MctcomRaw() string { return r.amtRaw(rbOffMctcom) }

// Mctdep returns MCTDEP S9(10), the département contribution.
func (r *RetB) Mctdep() (int64, bool) { return r.amt(rbOffMctdep) }

// SetMctdep moves v into MCTDEP S9(10).
func (r *RetB) SetMctdep(v int64) { r.setAmt(rbOffMctdep, v) }

// MctdepRaw returns the raw 10 bytes of MCTDEP.
func (r *RetB) MctdepRaw() string { return r.amtRaw(rbOffMctdep) }

// Mctreg returns MCTREG S9(10), the region contribution.
func (r *RetB) Mctreg() (int64, bool) { return r.amt(rbOffMctreg) }

// SetMctreg moves v into MCTREG S9(10).
func (r *RetB) SetMctreg(v int64) { r.setAmt(rbOffMctreg, v) }

// MctregRaw returns the raw 10 bytes of MCTREG.
func (r *RetB) MctregRaw() string { return r.amtRaw(rbOffMctreg) }

// Mctsyn returns MCTSYN S9(10), the syndicates / districts without own
// taxation contribution.
func (r *RetB) Mctsyn() (int64, bool) { return r.amt(rbOffMctsyn) }

// SetMctsyn moves v into MCTSYN S9(10).
func (r *RetB) SetMctsyn(v int64) { r.setAmt(rbOffMctsyn, v) }

// MctsynRaw returns the raw 10 bytes of MCTSYN.
func (r *RetB) MctsynRaw() string { return r.amtRaw(rbOffMctsyn) }

// Mctcu returns MCTCU S9(10), the grouping-with-own-taxation contribution.
func (r *RetB) Mctcu() (int64, bool) { return r.amt(rbOffMctcu) }

// SetMctcu moves v into MCTCU S9(10).
func (r *RetB) SetMctcu(v int64) { r.setAmt(rbOffMctcu, v) }

// MctcuRaw returns the raw 10 bytes of MCTCU.
func (r *RetB) MctcuRaw() string { return r.amtRaw(rbOffMctcu) }

// Mcttse returns MCTTSE S9(10), the special equipment tax contribution.
func (r *RetB) Mcttse() (int64, bool) { return r.amt(rbOffMcttse) }

// SetMcttse moves v into MCTTSE S9(10).
func (r *RetB) SetMcttse(v int64) { r.setAmt(rbOffMcttse, v) }

// McttseRaw returns the raw 10 bytes of MCTTSE.
func (r *RetB) McttseRaw() string { return r.amtRaw(rbOffMcttse) }

// Mcbt13 returns MCBT13(i+1) S9(10) (0-based i), the TSE EPF contribution.
func (r *RetB) Mcbt13(i int) (int64, bool, error) {
	if err := rbCheckIndex("MCBT13", i, RetBBt13Count); err != nil {
		return 0, false, err
	}
	v, ok := r.amt(rbOffMcbt13 + i*rbAmtLen)
	return v, ok, nil
}

// SetMcbt13 moves v into MCBT13(i+1) S9(10) (0-based i).
func (r *RetB) SetMcbt13(i int, v int64) error {
	if err := rbCheckIndex("MCBT13", i, RetBBt13Count); err != nil {
		return err
	}
	r.setAmt(rbOffMcbt13+i*rbAmtLen, v)
	return nil
}

// Mcbt13Raw returns the raw 10 bytes of MCBT13(i+1) (0-based i).
func (r *RetB) Mcbt13Raw(i int) (string, error) {
	if err := rbCheckIndex("MCBT13", i, RetBBt13Count); err != nil {
		return "", err
	}
	return r.amtRaw(rbOffMcbt13 + i*rbAmtLen), nil
}

// Mcbtsa returns MCBTSA PIC 9(10) (unsigned), the amalgamated new TSE
// contribution; ok is the NUMERIC test.
func (r *RetB) Mcbtsa() (uint64, bool) { return r.unsigned(rbOffMcbtsa, rbAmtLen) }

// SetMcbtsa moves v into MCBTSA PIC 9(10): the picture is unsigned, so the
// sign is dropped (a negative value is stored as its absolute value) and
// high-order digits are truncated, as a COBOL MOVE does.
func (r *RetB) SetMcbtsa(v int64) { r.setUnsignedFromSigned(rbOffMcbtsa, rbAmtLen, v) }

// McbtsaRaw returns the raw 10 bytes of MCBTSA.
func (r *RetB) McbtsaRaw() string { return r.amtRaw(rbOffMcbtsa) }

// Filler returns the X(40) 'X'-FILLER area following MCBTSA.
func (r *RetB) Filler() string { return r.get(rbOffFiller, RetBFillerLen) }

// SetFiller moves v into the X(40) 'X'-FILLER area.
func (r *RetB) SetFiller(v string) { r.put(rbOffFiller, RetBFillerLen, v) }

// ---- ACTOM: TEOM contributions, OCCURS 6 ----

// Gtauom returns GTAUOM(i+1) X(2) (0-based i), the TEOM regime.
func (r *RetB) Gtauom(i int) (string, error) {
	if err := rbCheckIndex("GTAUOM", i, RetBOmCount); err != nil {
		return "", err
	}
	return r.get(rbOffActom+i*rbOmLen, 2), nil
}

// SetGtauom moves v into GTAUOM(i+1) X(2) (0-based i).
func (r *RetB) SetGtauom(i int, v string) error {
	if err := rbCheckIndex("GTAUOM", i, RetBOmCount); err != nil {
		return err
	}
	r.put(rbOffActom+i*rbOmLen, 2, v)
	return nil
}

// Mctom returns MCTOM(i+1) S9(10) (0-based i), the TEOM contribution.
func (r *RetB) Mctom(i int) (int64, bool, error) {
	if err := rbCheckIndex("MCTOM", i, RetBOmCount); err != nil {
		return 0, false, err
	}
	v, ok := r.amt(rbOffActom + i*rbOmLen + 2)
	return v, ok, nil
}

// SetMctom moves v into MCTOM(i+1) S9(10) (0-based i).
func (r *RetB) SetMctom(i int, v int64) error {
	if err := rbCheckIndex("MCTOM", i, RetBOmCount); err != nil {
		return err
	}
	r.setAmt(rbOffActom+i*rbOmLen+2, v)
	return nil
}

// MctomRaw returns the raw 10 bytes of MCTOM(i+1) (0-based i).
func (r *RetB) MctomRaw(i int) (string, error) {
	if err := rbCheckIndex("MCTOM", i, RetBOmCount); err != nil {
		return "", err
	}
	return r.amtRaw(rbOffActom + i*rbOmLen + 2), nil
}

// ---- fees (with REDEFINES aliases) ----

// Mfa300 returns MFA300 S9(10), assessment fees at 3%.
func (r *RetB) Mfa300() (int64, bool) { return r.amt(rbOffMfa300) }

// SetMfa300 moves v into MFA300 S9(10).
func (r *RetB) SetMfa300(v int64) { r.setAmt(rbOffMfa300, v) }

// Mcfano returns MCFANO (REDEFINES MFA300), normal assessment fees.
func (r *RetB) Mcfano() (int64, bool) { return r.Mfa300() }

// SetMcfano moves v into MCFANO (same bytes as MFA300).
func (r *RetB) SetMcfano(v int64) { r.SetMfa300(v) }

// Mfn300 returns MFN300 S9(10), non-collection fees at 3%.
func (r *RetB) Mfn300() (int64, bool) { return r.amt(rbOffMfn300) }

// SetMfn300 moves v into MFN300 S9(10).
func (r *RetB) SetMfn300(v int64) { r.setAmt(rbOffMfn300, v) }

// Mcfnvn returns MCFNVN (REDEFINES MFN300), normal non-collection fees.
func (r *RetB) Mcfnvn() (int64, bool) { return r.Mfn300() }

// SetMcfnvn moves v into MCFNVN (same bytes as MFN300).
func (r *RetB) SetMcfnvn(v int64) { r.SetMfn300(v) }

// Mfa800 returns MFA800 S9(10), assessment fees at 8%.
func (r *RetB) Mfa800() (int64, bool) { return r.amt(rbOffMfa800) }

// SetMfa800 moves v into MFA800 S9(10).
func (r *RetB) SetMfa800(v int64) { r.setAmt(rbOffMfa800, v) }

// Momano returns MOMANO (REDEFINES MFA800), TEOM normal assessment fees.
func (r *RetB) Momano() (int64, bool) { return r.Mfa800() }

// SetMomano moves v into MOMANO (same bytes as MFA800).
func (r *RetB) SetMomano(v int64) { r.SetMfa800(v) }

// Mfn800 returns MFN800 S9(10), non-collection fees at 8%.
func (r *RetB) Mfn800() (int64, bool) { return r.amt(rbOffMfn800) }

// SetMfn800 moves v into MFN800 S9(10).
func (r *RetB) SetMfn800(v int64) { r.setAmt(rbOffMfn800, v) }

// Momnvn returns MOMNVN (REDEFINES MFN800), TEOM normal non-collection fees.
func (r *RetB) Momnvn() (int64, bool) { return r.Mfn800() }

// SetMomnvn moves v into MOMNVN (same bytes as MFN800).
func (r *RetB) SetMomnvn(v int64) { r.SetMfn800(v) }

// ---- totals, PIC S9(12) ----

// Tcthfr returns TCTHFR S9(12), total contributions excluding fees.
func (r *RetB) Tcthfr() (int64, bool) { return r.signed(rbOffTcthfr, rbTotLen) }

// SetTcthfr moves v into TCTHFR S9(12), truncating high-order digits.
func (r *RetB) SetTcthfr(v int64) { r.setSigned(rbOffTcthfr, rbTotLen, v) }

// Tctfra returns TCTFRA S9(12), total fees.
func (r *RetB) Tctfra() (int64, bool) { return r.signed(rbOffTctfra, rbTotLen) }

// SetTctfra moves v into TCTFRA S9(12), truncating high-order digits.
func (r *RetB) SetTctfra(v int64) { r.setSigned(rbOffTctfra, rbTotLen, v) }

// Tctdu returns TCTDU S9(12), total due including fees.
func (r *RetB) Tctdu() (int64, bool) { return r.signed(rbOffTctdu, rbTotLen) }

// SetTctdu moves v into TCTDU S9(12), truncating high-order digits.
func (r *RetB) SetTctdu(v int64) { r.setSigned(rbOffTctdu, rbTotLen, v) }

// Mfa900 returns MFA900 S9(10), assessment fees at 9%.
func (r *RetB) Mfa900() (int64, bool) { return r.amt(rbOffMfa900) }

// SetMfa900 moves v into MFA900 S9(10).
func (r *RetB) SetMfa900(v int64) { r.setAmt(rbOffMfa900, v) }

// Mfn900 returns MFN900 S9(10), non-collection fees at 9%.
func (r *RetB) Mfn900() (int64, bool) { return r.amt(rbOffMfn900) }

// SetMfn900 moves v into MFN900 S9(10).
func (r *RetB) SetMfn900(v int64) { r.setAmt(rbOffMfn900, v) }

// Tctom returns TCTOM PIC 9(10) (unsigned), total TEOM contributions.
func (r *RetB) Tctom() (uint64, bool) { return r.unsigned(rbOffTctom, rbAmtLen) }

// SetTctom moves v into TCTOM PIC 9(10): unsigned picture, so the sign is
// dropped and high-order digits are truncated, as a COBOL MOVE does.
func (r *RetB) SetTctom(v int64) { r.setUnsignedFromSigned(rbOffTctom, rbAmtLen, v) }

// TctomRaw returns the raw 10 bytes of TCTOM.
func (r *RetB) TctomRaw() string { return r.amtRaw(rbOffTctom) }

// Mvltim returns MVLTIM S9(10), the TEOMI amount.
func (r *RetB) Mvltim() (int64, bool) { return r.amt(rbOffMvltim) }

// SetMvltim moves v into MVLTIM S9(10).
func (r *RetB) SetMvltim(v int64) { r.setAmt(rbOffMvltim, v) }

// Mcoge3 returns MCOGE3 S9(10), the GEMAPI contribution.
func (r *RetB) Mcoge3() (int64, bool) { return r.amt(rbOffMcoge3) }

// SetMcoge3 moves v into MCOGE3 S9(10).
func (r *RetB) SetMcoge3(v int64) { r.setAmt(rbOffMcoge3, v) }

// Mcota3 returns MCOTA3 S9(10), the TASA contribution.
func (r *RetB) Mcota3() (int64, bool) { return r.amt(rbOffMcota3) }

// SetMcota3 moves v into MCOTA3 S9(10).
func (r *RetB) SetMcota3(v int64) { r.setAmt(rbOffMcota3, v) }

// ---- codes, trailer ----

// Ccoifp returns CCOIFP X(3), the absorbed-commune code.
func (r *RetB) Ccoifp() string { return r.get(rbOffCcoifp, 3) }

// SetCcoifp moves v into CCOIFP X(3).
func (r *RetB) SetCcoifp(v string) { r.put(rbOffCcoifp, 3, v) }

// Ccpper returns CCPPER X(3), the treasury code.
func (r *RetB) Ccpper() string { return r.get(rbOffCcpper, 3) }

// SetCcpper moves v into CCPPER X(3).
func (r *RetB) SetCcpper(v string) { r.put(rbOffCcpper, 3, v) }

// Trailer returns the trailing X(238) 'X'-FILLER area.
func (r *RetB) Trailer() string { return r.get(rbOffTrailer, RetBTrailerLen) }

// SetTrailer moves v into the trailing X(238) 'X'-FILLER area.
func (r *RetB) SetTrailer(v string) { r.put(rbOffTrailer, RetBTrailerLen, v) }
