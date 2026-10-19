// Package linkage holds the byte-exact layouts of the COBOL LINKAGE /
// working-storage fragments shared between the taxe foncière calculators.
package linkage

import (
	"fmt"
	"math/big"
)

// CombatLen is the total length in bytes of the GS12 group (copybook XCOMBAT).
const CombatLen = 600

// OCCURS counts of the GS12 group.
const (
	// CombatTauCount is the OCCURS count of GTOTAU (rate indicators).
	CombatTauCount = 8
	// CombatBt13Count is the OCCURS count of MBBT13.
	CombatBt13Count = 2
	// CombatOmCount is the OCCURS count of the ABAOM group (TEOM zones).
	CombatOmCount = 6
	// CombatReservedCount is the OCCURS count of the numeric 'X'-FILLER
	// PIC 9(10) table following ABAOM.
	CombatReservedCount = 12
)

// PvltomScale is the scaling factor of PVLTOM PIC 9V9(15): the scaled
// accessor returns the percentage multiplied by 10^15.
const PvltomScale uint64 = 1_000_000_000_000_000

// Group lengths.
const (
	// CombatAidficLen is the length of the AIDFIC group.
	CombatAidficLen = 11
	// CombatAnuproLen is the length of the ANUPRO group.
	CombatAnuproLen = 6
	// CombatMbabaLen is the length of the MBABA group.
	CombatMbabaLen = 100
	// CombatMbabaFillerLen is the length of the X(20) filler inside MBABA.
	CombatMbabaFillerLen = 20
	// CombatTrailerLen is the length of the trailing X(230) filler.
	CombatTrailerLen = 230
)

// Zero-based offsets inside the 600-byte GS12 group.
//
//	CCOBNB   X         0
//	AIDFIC   X(11)     1..11   DAN 9(4) 1..4, AC3DIR 5..7 (CC2DEP 5..6,
//	                           CCODIR 7), CCOCOM 8..10, DSRPAR 11
//	ANUPRO   X(6)     12..17   CGROUP 12, NNUPRO 9(5) 13..17
//	A0008    GTOTAU X OCCURS 8                         18..25
//	MBABA    X(100)   26..125  MBACOM 26, MBADEP 36, MBAREG 46, MBASYN 56,
//	                           MBACU 66, MBATSE 76 (all S9(10)),
//	                           MBBT13 S9(10) OCCURS 2 86..105, FILLER X(20) 106..125
//	ABAOM    OCCURS 6 x 12 126..197  GTAUOM X(2) +0, MBAOM S9(10) +2
//	FILLER   9(10) OCCURS 12   198..317
//	MVLTIM   S9(10)            318..327
//	PVLTOM   9V9(15)           328..343
//	MBAGE3   S9(10)            344..353
//	MBATA3   S9(10)            354..363
//	CCOIFP   X(3)              364..366
//	CCPPER   X(3)              367..369
//	FILLER   X(230)            370..599
const (
	cbOffCcobnb   = 0
	cbOffAidfic   = 1
	cbOffDan      = 1
	cbOffAc3dir   = 5
	cbOffCc2dep   = 5
	cbOffCcodir   = 7
	cbOffCcocom   = 8
	cbOffDsrpar   = 11
	cbOffAnupro   = 12
	cbOffCgroup   = 12
	cbOffNnupro   = 13
	cbOffGtotau   = 18
	cbOffMbaba    = 26
	cbOffMbacom   = 26
	cbOffMbadep   = 36
	cbOffMbareg   = 46
	cbOffMbasyn   = 56
	cbOffMbacu    = 66
	cbOffMbatse   = 76
	cbOffMbbt13   = 86
	cbOffMbabaFil = 106
	cbOffAbaom    = 126
	cbOffReserved = 198
	cbOffMvltim   = 318
	cbOffPvltom   = 328
	cbOffMbage3   = 344
	cbOffMbata3   = 354
	cbOffCcoifp   = 364
	cbOffCcpper   = 367
	cbOffTrailer  = 370

	cbAmtLen    = 10 // PIC S9(10) / 9(10)
	cbOmLen     = 12 // ABAOM occurrence: GTAUOM X(2) + MBAOM S9(10)
	cbPvltomLen = 16 // PIC 9V9(15)
)

// Compile-time layout checks.
var (
	_ = [1]struct{}{}[cbOffAidfic+CombatAidficLen-cbOffAnupro]
	_ = [1]struct{}{}[cbOffAnupro+CombatAnuproLen-cbOffGtotau]
	_ = [1]struct{}{}[cbOffGtotau+CombatTauCount-cbOffMbaba]
	_ = [1]struct{}{}[cbOffMbbt13+CombatBt13Count*cbAmtLen-cbOffMbabaFil]
	_ = [1]struct{}{}[cbOffMbaba+CombatMbabaLen-cbOffAbaom]
	_ = [1]struct{}{}[cbOffAbaom+CombatOmCount*cbOmLen-cbOffReserved]
	_ = [1]struct{}{}[cbOffReserved+CombatReservedCount*cbAmtLen-cbOffMvltim]
	_ = [1]struct{}{}[cbOffPvltom+cbPvltomLen-cbOffMbage3]
	_ = [1]struct{}{}[cbOffTrailer+CombatTrailerLen-CombatLen]
)

// gtauomOK is the value list of the level-88 'X'-GTAUOM-OK.
var gtauomOK = map[string]struct{}{
	"  ": {}, " P": {}, "P ": {}, "RA": {}, "RB": {}, "RC": {}, "RD": {}, "RE": {},
}

// Combat is the GS12 group of copybook XCOMBAT: a 600-byte fixed record of
// property tax / TEOM bases (year, department/direction/commune, owner
// account, eight rate indicators, bases per collectivity, six TEOM zone
// bases, rental value and TEOM percentage, GEMAPI and TASA bases).
//
// MIGRATION_NOTE: the 'X' placeholder of COPY XCOMBAT REPLACING 'X' BY
// COMBAT (EFITA3B8) disappears: each inclusion is a distinct Combat value.
// The 'X'-FILLER items become real data names after substitution and are
// exposed (MbabaFiller, Reserved, Trailer) so their content is preserved.
//
// MIGRATION_NOTE: encoding is ASCII (host data already converted from
// EBCDIC). Numerics are zoned decimal USAGE DISPLAY (no COMP-3). Signed
// S9(10) fields carry a trailing overpunch sign; decoding accepts plain
// digits, '{','A'..'I' / '}','J'..'R' (IBM) and 'p'..'y' (ASCII negative);
// encoding writes the IBM convention, as in internal/layout/taudis. Every
// numeric getter also returns the COBOL NUMERIC class-test result, and Raw
// accessors expose the bytes so callers can test NUMERIC before converting.
// OCCURS tables use 0-based indexed accessors that error out of range.
// There is no REDEFINES in this copybook.
type Combat [CombatLen]byte

// NewCombat returns a group initialised to spaces (MOVE SPACES TO GS12).
func NewCombat() Combat {
	var c Combat
	for i := range c {
		c[i] = ' '
	}
	return c
}

// DecodeCombat wraps a raw 600-byte buffer as a Combat group.
func DecodeCombat(b [CombatLen]byte) Combat { return Combat(b) }

// ParseCombat copies a buffer into a Combat; it fails unless b is exactly
// CombatLen bytes long.
func ParseCombat(b []byte) (Combat, error) {
	var c Combat
	if len(b) != CombatLen {
		return c, fmt.Errorf("linkage: GS12 (XCOMBAT) must be %d bytes, got %d", CombatLen, len(b))
	}
	copy(c[:], b)
	return c, nil
}

// Bytes returns a copy of the raw 600 bytes.
func (c *Combat) Bytes() []byte {
	out := make([]byte, CombatLen)
	copy(out, c[:])
	return out
}

// ---- low-level helpers ----

func (c *Combat) get(off, n int) string { return string(c[off : off+n]) }

// put performs a COBOL alphanumeric MOVE (left-justified, space-padded,
// right-truncated).
func (c *Combat) put(off, n int, v string) {
	for i := 0; i < n; i++ {
		if i < len(v) {
			c[off+i] = v[i]
		} else {
			c[off+i] = ' '
		}
	}
}

func (c *Combat) isNumeric(off, n int) bool {
	for _, b := range c[off : off+n] {
		if b < '0' || b > '9' {
			return false
		}
	}
	return true
}

func lowDigit(b byte) uint64 {
	d := uint64(b & 0x0F)
	if d > 9 {
		d = 9 // MIGRATION_NOTE: invalid nibble clamped to keep the value in range
	}
	return d
}

// unsigned decodes an unsigned zoned field; ok is the NUMERIC test.
func (c *Combat) unsigned(off, n int) (uint64, bool) {
	ok := c.isNumeric(off, n)
	var v uint64
	for _, b := range c[off : off+n] {
		v = v*10 + lowDigit(b)
	}
	return v, ok
}

// setUnsigned performs a numeric MOVE into an unsigned zoned field,
// truncating high-order digits.
func (c *Combat) setUnsigned(off, n int, v uint64) {
	for i := off + n - 1; i >= off; i-- {
		c[i] = byte('0' + v%10)
		v /= 10
	}
}

func overpunch(b byte) (digit uint64, neg bool, ok bool) {
	switch {
	case b >= '0' && b <= '9':
		return uint64(b - '0'), false, true
	case b == '{':
		return 0, false, true
	case b >= 'A' && b <= 'I':
		return uint64(b-'A') + 1, false, true
	case b == '}':
		return 0, true, true
	case b >= 'J' && b <= 'R':
		return uint64(b-'J') + 1, true, true
	case b >= 'p' && b <= 'y':
		return uint64(b - 'p'), true, true
	}
	return lowDigit(b), false, false
}

// signed decodes a PIC S9(n) trailing-overpunch field; ok is the NUMERIC test.
func (c *Combat) signed(off, n int) (int64, bool) {
	last := off + n - 1
	ok := c.isNumeric(off, n-1)
	var v uint64
	for _, b := range c[off:last] {
		v = v*10 + lowDigit(b)
	}
	d, neg, lastOK := overpunch(c[last])
	v = v*10 + d
	if neg {
		return -int64(v), ok && lastOK
	}
	return int64(v), ok && lastOK
}

// setSigned performs a numeric MOVE into a PIC S9(n) trailing-overpunch
// field, truncating high-order digits.
func (c *Combat) setSigned(off, n int, v int64) {
	neg := v < 0
	u := uint64(v)
	if neg {
		u = uint64(-(v + 1)) + 1 // safe for math.MinInt64
	}
	c.setUnsigned(off, n, u)
	last := off + n - 1
	d := c[last] - '0'
	allZero := c.isZeroDigits(off, n)
	switch {
	case neg && !allZero && d == 0:
		c[last] = '}'
	case neg && !allZero:
		c[last] = 'J' + d - 1
	case d == 0:
		c[last] = '{'
	default:
		c[last] = 'A' + d - 1
	}
}

// isZeroDigits reports whether a just-written digit field is all zeros
// (a truncated negative that becomes zero is stored as +0, as COBOL does).
func (c *Combat) isZeroDigits(off, n int) bool {
	for _, b := range c[off : off+n] {
		if b != '0' {
			return false
		}
	}
	return true
}

func checkIndex(field string, i, count int) error {
	if i < 0 || i >= count {
		return fmt.Errorf("linkage: GS12 %s subscript %d out of range [0,%d)", field, i, count)
	}
	return nil
}

// ---- header ----

// Ccobnb returns CCOBNB X: built / non-built code.
func (c *Combat) Ccobnb() string { return c.get(cbOffCcobnb, 1) }

// SetCcobnb moves v into CCOBNB X.
func (c *Combat) SetCcobnb(v string) { c.put(cbOffCcobnb, 1, v) }

// Aidfic returns the raw 11-byte AIDFIC group (for group moves).
func (c *Combat) Aidfic() string { return c.get(cbOffAidfic, CombatAidficLen) }

// SetAidfic moves v into the 11-byte AIDFIC group.
func (c *Combat) SetAidfic(v string) { c.put(cbOffAidfic, CombatAidficLen, v) }

// DanRaw returns the raw 4 bytes of DAN PIC 9(4).
func (c *Combat) DanRaw() string { return c.get(cbOffDan, 4) }

// Dan returns DAN PIC 9(4), the tax year; ok is the NUMERIC test.
func (c *Combat) Dan() (uint64, bool) { return c.unsigned(cbOffDan, 4) }

// SetDan moves v into DAN PIC 9(4).
func (c *Combat) SetDan(v uint64) { c.setUnsigned(cbOffDan, 4, v) }

// Ac3dir returns AC3DIR X(3) (CC2DEP + CCODIR).
func (c *Combat) Ac3dir() string { return c.get(cbOffAc3dir, 3) }

// SetAc3dir moves v into AC3DIR X(3).
func (c *Combat) SetAc3dir(v string) { c.put(cbOffAc3dir, 3, v) }

// Cc2dep returns CC2DEP X(2), the département code.
func (c *Combat) Cc2dep() string { return c.get(cbOffCc2dep, 2) }

// SetCc2dep moves v into CC2DEP X(2).
func (c *Combat) SetCc2dep(v string) { c.put(cbOffCc2dep, 2, v) }

// Ccodir returns CCODIR X, the direction code.
func (c *Combat) Ccodir() string { return c.get(cbOffCcodir, 1) }

// SetCcodir moves v into CCODIR X.
func (c *Combat) SetCcodir(v string) { c.put(cbOffCcodir, 1, v) }

// Ccocom returns CCOCOM X(3), the commune code.
func (c *Combat) Ccocom() string { return c.get(cbOffCcocom, 3) }

// SetCcocom moves v into CCOCOM X(3).
func (c *Combat) SetCcocom(v string) { c.put(cbOffCcocom, 3, v) }

// Dsrpar returns DSRPAR X.
func (c *Combat) Dsrpar() string { return c.get(cbOffDsrpar, 1) }

// SetDsrpar moves v into DSRPAR X.
func (c *Combat) SetDsrpar(v string) { c.put(cbOffDsrpar, 1, v) }

// Anupro returns the raw 6-byte ANUPRO group (owner account, group moves).
func (c *Combat) Anupro() string { return c.get(cbOffAnupro, CombatAnuproLen) }

// SetAnupro moves v into the 6-byte ANUPRO group.
func (c *Combat) SetAnupro(v string) { c.put(cbOffAnupro, CombatAnuproLen, v) }

// Cgroup returns CGROUP X, the owner group letter.
func (c *Combat) Cgroup() string { return c.get(cbOffCgroup, 1) }

// SetCgroup moves v into CGROUP X.
func (c *Combat) SetCgroup(v string) { c.put(cbOffCgroup, 1, v) }

// NnuproRaw returns the raw 5 bytes of NNUPRO PIC 9(5).
func (c *Combat) NnuproRaw() string { return c.get(cbOffNnupro, 5) }

// Nnupro returns NNUPRO PIC 9(5), the owner number; ok is the NUMERIC test.
func (c *Combat) Nnupro() (uint64, bool) { return c.unsigned(cbOffNnupro, 5) }

// SetNnupro moves v into NNUPRO PIC 9(5).
func (c *Combat) SetNnupro(v uint64) { c.setUnsigned(cbOffNnupro, 5, v) }

// A0008 returns the raw 8-byte A0008 group (all GTOTAU indicators).
func (c *Combat) A0008() string { return c.get(cbOffGtotau, CombatTauCount) }

// SetA0008 moves v into the 8-byte A0008 group.
func (c *Combat) SetA0008(v string) { c.put(cbOffGtotau, CombatTauCount, v) }

// Gtotau returns GTOTAU(i+1) X (0-based i), a rate indicator.
func (c *Combat) Gtotau(i int) (string, error) {
	if err := checkIndex("GTOTAU", i, CombatTauCount); err != nil {
		return "", err
	}
	return c.get(cbOffGtotau+i, 1), nil
}

// SetGtotau moves v into GTOTAU(i+1) X (0-based i).
func (c *Combat) SetGtotau(i int, v string) error {
	if err := checkIndex("GTOTAU", i, CombatTauCount); err != nil {
		return err
	}
	c.put(cbOffGtotau+i, 1, v)
	return nil
}

// ---- MBABA: bases per collectivity, PIC S9(10) ----

// Mbaba returns the raw 100-byte MBABA group.
func (c *Combat) Mbaba() string { return c.get(cbOffMbaba, CombatMbabaLen) }

// SetMbaba moves v into the 100-byte MBABA group.
func (c *Combat) SetMbaba(v string) { c.put(cbOffMbaba, CombatMbabaLen, v) }

// Mbacom returns MBACOM S9(10), the commune base; ok is the NUMERIC test.
func (c *Combat) Mbacom() (int64, bool) { return c.signed(cbOffMbacom, cbAmtLen) }

// SetMbacom moves v into MBACOM S9(10).
func (c *Combat) SetMbacom(v int64) { c.setSigned(cbOffMbacom, cbAmtLen, v) }

// MbacomRaw returns the raw 10 bytes of MBACOM.
func (c *Combat) MbacomRaw() string { return c.get(cbOffMbacom, cbAmtLen) }

// Mbadep returns MBADEP S9(10), the département base.
func (c *Combat) Mbadep() (int64, bool) { return c.signed(cbOffMbadep, cbAmtLen) }

// SetMbadep moves v into MBADEP S9(10).
func (c *Combat) SetMbadep(v int64) { c.setSigned(cbOffMbadep, cbAmtLen, v) }

// MbadepRaw returns the raw 10 bytes of MBADEP.
func (c *Combat) MbadepRaw() string { return c.get(cbOffMbadep, cbAmtLen) }

// Mbareg returns MBAREG S9(10), the region base.
func (c *Combat) Mbareg() (int64, bool) { return c.signed(cbOffMbareg, cbAmtLen) }

// SetMbareg moves v into MBAREG S9(10).
func (c *Combat) SetMbareg(v int64) { c.setSigned(cbOffMbareg, cbAmtLen, v) }

// MbaregRaw returns the raw 10 bytes of MBAREG.
func (c *Combat) MbaregRaw() string { return c.get(cbOffMbareg, cbAmtLen) }

// Mbasyn returns MBASYN S9(10), the syndicate base.
func (c *Combat) Mbasyn() (int64, bool) { return c.signed(cbOffMbasyn, cbAmtLen) }

// SetMbasyn moves v into MBASYN S9(10).
func (c *Combat) SetMbasyn(v int64) { c.setSigned(cbOffMbasyn, cbAmtLen, v) }

// MbasynRaw returns the raw 10 bytes of MBASYN.
func (c *Combat) MbasynRaw() string { return c.get(cbOffMbasyn, cbAmtLen) }

// Mbacu returns MBACU S9(10), the grouping-of-communes base.
func (c *Combat) Mbacu() (int64, bool) { return c.signed(cbOffMbacu, cbAmtLen) }

// SetMbacu moves v into MBACU S9(10).
func (c *Combat) SetMbacu(v int64) { c.setSigned(cbOffMbacu, cbAmtLen, v) }

// MbacuRaw returns the raw 10 bytes of MBACU.
func (c *Combat) MbacuRaw() string { return c.get(cbOffMbacu, cbAmtLen) }

// Mbatse returns MBATSE S9(10), the TSE base.
func (c *Combat) Mbatse() (int64, bool) { return c.signed(cbOffMbatse, cbAmtLen) }

// SetMbatse moves v into MBATSE S9(10).
func (c *Combat) SetMbatse(v int64) { c.setSigned(cbOffMbatse, cbAmtLen, v) }

// MbatseRaw returns the raw 10 bytes of MBATSE.
func (c *Combat) MbatseRaw() string { return c.get(cbOffMbatse, cbAmtLen) }

// Mbbt13 returns MBBT13(i+1) S9(10) (0-based i).
func (c *Combat) Mbbt13(i int) (int64, bool, error) {
	if err := checkIndex("MBBT13", i, CombatBt13Count); err != nil {
		return 0, false, err
	}
	v, ok := c.signed(cbOffMbbt13+i*cbAmtLen, cbAmtLen)
	return v, ok, nil
}

// SetMbbt13 moves v into MBBT13(i+1) S9(10) (0-based i).
func (c *Combat) SetMbbt13(i int, v int64) error {
	if err := checkIndex("MBBT13", i, CombatBt13Count); err != nil {
		return err
	}
	c.setSigned(cbOffMbbt13+i*cbAmtLen, cbAmtLen, v)
	return nil
}

// MbabaFiller returns the X(20) 'X'-FILLER area at the end of MBABA.
func (c *Combat) MbabaFiller() string { return c.get(cbOffMbabaFil, CombatMbabaFillerLen) }

// SetMbabaFiller moves v into the X(20) 'X'-FILLER area of MBABA.
func (c *Combat) SetMbabaFiller(v string) { c.put(cbOffMbabaFil, CombatMbabaFillerLen, v) }

// ---- ABAOM: TEOM zones, OCCURS 6 ----

// Gtauom returns GTAUOM(i+1) X(2) (0-based i), the TEOM zone rate code.
func (c *Combat) Gtauom(i int) (string, error) {
	if err := checkIndex("GTAUOM", i, CombatOmCount); err != nil {
		return "", err
	}
	return c.get(cbOffAbaom+i*cbOmLen, 2), nil
}

// SetGtauom moves v into GTAUOM(i+1) X(2) (0-based i).
func (c *Combat) SetGtauom(i int, v string) error {
	if err := checkIndex("GTAUOM", i, CombatOmCount); err != nil {
		return err
	}
	c.put(cbOffAbaom+i*cbOmLen, 2, v)
	return nil
}

// GtauomOK evaluates the level-88 GTAUOM-OK on occurrence i+1 (0-based i):
// true when GTAUOM is one of '  ', ' P', 'P ', 'RA', 'RB', 'RC', 'RD', 'RE'.
func (c *Combat) GtauomOK(i int) (bool, error) {
	v, err := c.Gtauom(i)
	if err != nil {
		return false, err
	}
	_, ok := gtauomOK[v]
	return ok, nil
}

// Mbaom returns MBAOM(i+1) S9(10) (0-based i), the TEOM zone base.
func (c *Combat) Mbaom(i int) (int64, bool, error) {
	if err := checkIndex("MBAOM", i, CombatOmCount); err != nil {
		return 0, false, err
	}
	v, ok := c.signed(cbOffAbaom+i*cbOmLen+2, cbAmtLen)
	return v, ok, nil
}

// MbaomRaw returns the raw 10 bytes of MBAOM(i+1) (0-based i).
func (c *Combat) MbaomRaw(i int) (string, error) {
	if err := checkIndex("MBAOM", i, CombatOmCount); err != nil {
		return "", err
	}
	return c.get(cbOffAbaom+i*cbOmLen+2, cbAmtLen), nil
}

// SetMbaom moves v into MBAOM(i+1) S9(10) (0-based i).
func (c *Combat) SetMbaom(i int, v int64) error {
	if err := checkIndex("MBAOM", i, CombatOmCount); err != nil {
		return err
	}
	c.setSigned(cbOffAbaom+i*cbOmLen+2, cbAmtLen, v)
	return nil
}

// ---- reserved numeric table ----

// Reserved returns the numeric 'X'-FILLER PIC 9(10) element i+1 (0-based i).
func (c *Combat) Reserved(i int) (uint64, bool, error) {
	if err := checkIndex("FILLER 9(10)", i, CombatReservedCount); err != nil {
		return 0, false, err
	}
	v, ok := c.unsigned(cbOffReserved+i*cbAmtLen, cbAmtLen)
	return v, ok, nil
}

// SetReserved moves v into the numeric 'X'-FILLER element i+1 (0-based i).
func (c *Combat) SetReserved(i int, v uint64) error {
	if err := checkIndex("FILLER 9(10)", i, CombatReservedCount); err != nil {
		return err
	}
	c.setUnsigned(cbOffReserved+i*cbAmtLen, cbAmtLen, v)
	return nil
}

// ---- rental value / TEOM percentage ----

// Mvltim returns MVLTIM S9(10), the rental value.
func (c *Combat) Mvltim() (int64, bool) { return c.signed(cbOffMvltim, cbAmtLen) }

// SetMvltim moves v into MVLTIM S9(10).
func (c *Combat) SetMvltim(v int64) { c.setSigned(cbOffMvltim, cbAmtLen, v) }

// MvltimRaw returns the raw 10 bytes of MVLTIM.
func (c *Combat) MvltimRaw() string { return c.get(cbOffMvltim, cbAmtLen) }

// PvltomRaw returns the raw 16 bytes of PVLTOM PIC 9V9(15).
func (c *Combat) PvltomRaw() string { return c.get(cbOffPvltom, cbPvltomLen) }

// PvltomScaled returns PVLTOM PIC 9V9(15) scaled by PvltomScale (10^15);
// ok is the NUMERIC test.
func (c *Combat) PvltomScaled() (uint64, bool) { return c.unsigned(cbOffPvltom, cbPvltomLen) }

// SetPvltomScaled moves a PvltomScale-scaled value into PVLTOM; high-order
// digits beyond 9V9(15) are truncated.
func (c *Combat) SetPvltomScaled(v uint64) { c.setUnsigned(cbOffPvltom, cbPvltomLen, v) }

// Pvltom returns PVLTOM as an exact decimal (big.Rat); ok is the NUMERIC test.
func (c *Combat) Pvltom() (*big.Rat, bool) {
	v, ok := c.PvltomScaled()
	r := new(big.Rat).SetFrac(new(big.Int).SetUint64(v), new(big.Int).SetUint64(PvltomScale))
	return r, ok
}

// SetPvltom performs a COBOL MOVE of a decimal into PVLTOM 9V9(15): the
// sign is dropped (unsigned picture), fraction digits beyond 15 and integer
// digits beyond 1 are truncated (no rounding).
func (c *Combat) SetPvltom(r *big.Rat) error {
	if r == nil {
		return fmt.Errorf("linkage: GS12 PVLTOM: nil value")
	}
	abs := new(big.Rat).Abs(r)
	scaled := new(big.Int).Mul(abs.Num(), new(big.Int).SetUint64(PvltomScale))
	scaled.Quo(scaled, abs.Denom())
	scaled.Mod(scaled, new(big.Int).Mul(big.NewInt(10), new(big.Int).SetUint64(PvltomScale)))
	c.SetPvltomScaled(scaled.Uint64())
	return nil
}

// ---- GEMAPI / TASA bases ----

// Mbage3 returns MBAGE3 S9(10), the GEMAPI base.
func (c *Combat) Mbage3() (int64, bool) { return c.signed(cbOffMbage3, cbAmtLen) }

// SetMbage3 moves v into MBAGE3 S9(10).
func (c *Combat) SetMbage3(v int64) { c.setSigned(cbOffMbage3, cbAmtLen, v) }

// Mbage3Raw returns the raw 10 bytes of MBAGE3.
func (c *Combat) Mbage3Raw() string { return c.get(cbOffMbage3, cbAmtLen) }

// Mbata3 returns MBATA3 S9(10), the TASA base.
func (c *Combat) Mbata3() (int64, bool) { return c.signed(cbOffMbata3, cbAmtLen) }

// SetMbata3 moves v into MBATA3 S9(10).
func (c *Combat) SetMbata3(v int64) { c.setSigned(cbOffMbata3, cbAmtLen, v) }

// Mbata3Raw returns the raw 10 bytes of MBATA3.
func (c *Combat) Mbata3Raw() string { return c.get(cbOffMbata3, cbAmtLen) }

// Ccoifp returns CCOIFP X(3), the absorbed-commune code.
func (c *Combat) Ccoifp() string { return c.get(cbOffCcoifp, 3) }

// SetCcoifp moves v into CCOIFP X(3).
func (c *Combat) SetCcoifp(v string) { c.put(cbOffCcoifp, 3, v) }

// Ccpper returns CCPPER X(3), the perception code.
func (c *Combat) Ccpper() string { return c.get(cbOffCcpper, 3) }

// SetCcpper moves v into CCPPER X(3).
func (c *Combat) SetCcpper(v string) { c.put(cbOffCcpper, 3, v) }

// Trailer returns the trailing X(230) 'X'-FILLER area.
func (c *Combat) Trailer() string { return c.get(cbOffTrailer, CombatTrailerLen) }

// SetTrailer moves v into the trailing X(230) 'X'-FILLER area.
func (c *Combat) SetTrailer(v string) { c.put(cbOffTrailer, CombatTrailerLen, v) }
