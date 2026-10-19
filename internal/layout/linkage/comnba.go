package linkage

import "fmt"

// ComnbaLen is the total length in bytes of a non-bâti tax-base record
// (copybook XCOMNBA).
const ComnbaLen = 600

// OCCURS counts of the XCOMNBA record.
const (
	// ComnbaTauCount is the OCCURS count of GTOTAU (rate-applied indicators).
	ComnbaTauCount = 11
	// ComnbaNt13Count is the OCCURS count of MBNT13.
	ComnbaNt13Count = 2
	// ComnbaJt13Count is the OCCURS count of MBJT13.
	ComnbaJt13Count = 2
)

// Group lengths of the XCOMNBA record.
const (
	// ComnbaAidficLen is the length of the AIDFIC group.
	ComnbaAidficLen = 11
	// ComnbaAnuproLen is the length of the ANUPRO group.
	ComnbaAnuproLen = 6
	// ComnbaAbanbLen is the length of the ABANB group (12 x S9(10)).
	ComnbaAbanbLen = 120
	// ComnbaAbajagLen is the length of the ABAJAG group (12 x S9(10)).
	ComnbaAbajagLen = 120
	// ComnbaTrailerLen is the length of the trailing X(315) filler.
	ComnbaTrailerLen = 315
)

// Zero-based offsets inside the 600-byte XCOMNBA record.
//
//	CCOBNB   X          0
//	AIDFIC   X(11)      1..11   DAN 9(4) 1..4, AC3DIR 5..7 (CC2DEP 5..6,
//	                            CCODIR 7), CCOCOM 8..10, DSRPAR 11
//	ANUPRO   X(6)      12..17   CGROUP 12, NNUPRO 9(5) 13..17
//	A0011    GTOTAU X OCCURS 11                           18..28
//	ABANB    X(120)    29..148  MBACOM 29, MBADEP 39, MBAREG 49, MBASYN 59,
//	                            MBACU 69, MBATSE 79, MBNT13 OCCURS 2 89..108,
//	                            MBACA 109, MBABAP 119, MBACAA 129, MBAPOS 139
//	ABAJAG   X(120)   149..268  MBJCOM 149, MBJSYN 159, MBJCU 169, MBJTSE 179,
//	                            MBJT13 OCCURS 2 189..208, MBJDEP 209,
//	                            MB1PRE 219, MBJPRE 229, MBJECO 239,
//	                            MBJESY 249, MBJEIC 259
//	MBAGE3   S9(10)   269..278
//	CCOIFP   X(3)     279..281
//	CCPPER   X(3)     282..284
//	FILLER   X(315)   285..599
const (
	nbOffCcobnb  = 0
	nbOffAidfic  = 1
	nbOffDan     = 1
	nbOffAc3dir  = 5
	nbOffCc2dep  = 5
	nbOffCcodir  = 7
	nbOffCcocom  = 8
	nbOffDsrpar  = 11
	nbOffAnupro  = 12
	nbOffCgroup  = 12
	nbOffNnupro  = 13
	nbOffGtotau  = 18
	nbOffAbanb   = 29
	nbOffMbacom  = 29
	nbOffMbadep  = 39
	nbOffMbareg  = 49
	nbOffMbasyn  = 59
	nbOffMbacu   = 69
	nbOffMbatse  = 79
	nbOffMbnt13  = 89
	nbOffMbaca   = 109
	nbOffMbabap  = 119
	nbOffMbacaa  = 129
	nbOffMbapos  = 139
	nbOffAbajag  = 149
	nbOffMbjcom  = 149
	nbOffMbjsyn  = 159
	nbOffMbjcu   = 169
	nbOffMbjtse  = 179
	nbOffMbjt13  = 189
	nbOffMbjdep  = 209
	nbOffMb1pre  = 219
	nbOffMbjpre  = 229
	nbOffMbjeco  = 239
	nbOffMbjesy  = 249
	nbOffMbjeic  = 259
	nbOffMbage3  = 269
	nbOffCcoifp  = 279
	nbOffCcpper  = 282
	nbOffTrailer = 285

	nbAmtLen = 10 // PIC S9(10)
)

// Compile-time layout checks (a negative or out-of-range index fails to build).
var (
	_ = [1]struct{}{}[nbOffAidfic+ComnbaAidficLen-nbOffAnupro]
	_ = [1]struct{}{}[nbOffAnupro+ComnbaAnuproLen-nbOffGtotau]
	_ = [1]struct{}{}[nbOffGtotau+ComnbaTauCount-nbOffAbanb]
	_ = [1]struct{}{}[nbOffMbnt13+ComnbaNt13Count*nbAmtLen-nbOffMbaca]
	_ = [1]struct{}{}[nbOffMbapos+nbAmtLen-nbOffAbajag]
	_ = [1]struct{}{}[nbOffAbanb+ComnbaAbanbLen-nbOffAbajag]
	_ = [1]struct{}{}[nbOffMbjt13+ComnbaJt13Count*nbAmtLen-nbOffMbjdep]
	_ = [1]struct{}{}[nbOffMbjeic+nbAmtLen-nbOffMbage3]
	_ = [1]struct{}{}[nbOffAbajag+ComnbaAbajagLen-nbOffMbage3]
	_ = [1]struct{}{}[nbOffMbage3+nbAmtLen-nbOffCcoifp]
	_ = [1]struct{}{}[nbOffTrailer+ComnbaTrailerLen-ComnbaLen]
)

// Comnba is the 600-byte non-bâti (unbuilt land) tax-base record of
// copybook XCOMNBA, included by EFITA3N8 as COMNONBA: built/non-built code,
// file identifier (year, département/direction, commune, roll series),
// owner account, eleven rate-applied indicators, non-bâti bases per
// beneficiary authority, young-farmer relief bases, GEMAPI base,
// absorbed-commune and treasury codes.
//
// MIGRATION_NOTE: the 'X' placeholder of COPY XCOMNBA REPLACING 'X' BY
// COMNONBA disappears: each inclusion is a distinct Comnba value. The
// trailing 'X'-FILLER X(315) becomes a real data name after substitution
// and is exposed (Trailer) so its content is preserved byte-for-byte.
//
// MIGRATION_NOTE: same encoding conventions as Combat (ASCII, zoned decimal
// USAGE DISPLAY, trailing overpunch sign on S9(10) accepting IBM '{','A'..'I'
// / '}','J'..'R' and ASCII 'p'..'y'; encoding writes the IBM convention).
// Every numeric getter returns the COBOL NUMERIC class-test result and Raw
// accessors expose the bytes so callers can test NUMERIC before converting
// (EFITA3N8 tests AIDFIC/ANUPRO components). OCCURS tables use 0-based
// indexed accessors that return an error when out of range. The record has
// no PIC V field, so no decimal accessor is needed; there is no REDEFINES.
//
// MIGRATION_NOTE: NewComnba initialises with spaces (MOVE SPACES): numeric
// fields are then non-NUMERIC until they are set.
type Comnba [ComnbaLen]byte

// NewComnba returns a record initialised to spaces (MOVE SPACES TO COMNONBA).
func NewComnba() Comnba {
	var c Comnba
	for i := range c {
		c[i] = ' '
	}
	return c
}

// DecodeComnba wraps a raw 600-byte buffer as a Comnba record.
func DecodeComnba(b [ComnbaLen]byte) Comnba { return Comnba(b) }

// ParseComnba copies a buffer into a Comnba; it fails unless b is exactly
// ComnbaLen bytes long.
func ParseComnba(b []byte) (Comnba, error) {
	var c Comnba
	if len(b) != ComnbaLen {
		return c, fmt.Errorf("linkage: COMNONBA (XCOMNBA) must be %d bytes, got %d", ComnbaLen, len(b))
	}
	copy(c[:], b)
	return c, nil
}

// Bytes returns a copy of the raw 600 bytes.
func (c *Comnba) Bytes() []byte {
	out := make([]byte, ComnbaLen)
	copy(out, c[:])
	return out
}

// ---- low-level helpers ----

func (c *Comnba) get(off, n int) string { return string(c[off : off+n]) }

// put performs a COBOL alphanumeric MOVE (left-justified, space-padded,
// right-truncated).
func (c *Comnba) put(off, n int, v string) {
	for i := 0; i < n; i++ {
		if i < len(v) {
			c[off+i] = v[i]
		} else {
			c[off+i] = ' '
		}
	}
}

func (c *Comnba) isNumeric(off, n int) bool {
	for _, b := range c[off : off+n] {
		if b < '0' || b > '9' {
			return false
		}
	}
	return true
}

// unsigned decodes an unsigned zoned field; ok is the NUMERIC test.
func (c *Comnba) unsigned(off, n int) (uint64, bool) {
	ok := c.isNumeric(off, n)
	var v uint64
	for _, b := range c[off : off+n] {
		v = v*10 + lowDigit(b)
	}
	return v, ok
}

// setUnsigned performs a numeric MOVE into an unsigned zoned field,
// truncating high-order digits.
func (c *Comnba) setUnsigned(off, n int, v uint64) {
	for i := off + n - 1; i >= off; i-- {
		c[i] = byte('0' + v%10)
		v /= 10
	}
}

// signed decodes a PIC S9(n) trailing-overpunch field; ok is the NUMERIC test.
func (c *Comnba) signed(off, n int) (int64, bool) {
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
// field: high-order digits beyond n are truncated and a negative value
// truncated to zero is stored as +0, as COBOL does.
func (c *Comnba) setSigned(off, n int, v int64) {
	neg := v < 0
	u := uint64(v)
	if neg {
		u = uint64(-(v + 1)) + 1 // safe for math.MinInt64
	}
	c.setUnsigned(off, n, u)
	last := off + n - 1
	d := c[last] - '0'
	allZero := true
	for _, b := range c[off : off+n] {
		if b != '0' {
			allZero = false
			break
		}
	}
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

func nbCheckIndex(field string, i, count int) error {
	if i < 0 || i >= count {
		return fmt.Errorf("linkage: COMNONBA %s subscript %d out of range [0,%d)", field, i, count)
	}
	return nil
}

func (c *Comnba) amt(off int) (int64, bool) { return c.signed(off, nbAmtLen) }

func (c *Comnba) setAmt(off int, v int64) { c.setSigned(off, nbAmtLen, v) }

func (c *Comnba) amtRaw(off int) string { return c.get(off, nbAmtLen) }

// ---- header ----

// Ccobnb returns CCOBNB X: built / non-built code (record discriminator).
func (c *Comnba) Ccobnb() string { return c.get(nbOffCcobnb, 1) }

// SetCcobnb moves v into CCOBNB X.
func (c *Comnba) SetCcobnb(v string) { c.put(nbOffCcobnb, 1, v) }

// Aidfic returns the raw 11-byte AIDFIC group (file identifier, group moves).
func (c *Comnba) Aidfic() string { return c.get(nbOffAidfic, ComnbaAidficLen) }

// SetAidfic moves v into the 11-byte AIDFIC group.
func (c *Comnba) SetAidfic(v string) { c.put(nbOffAidfic, ComnbaAidficLen, v) }

// DanRaw returns the raw 4 bytes of DAN PIC 9(4).
func (c *Comnba) DanRaw() string { return c.get(nbOffDan, 4) }

// Dan returns DAN PIC 9(4), the tax year; ok is the NUMERIC test.
func (c *Comnba) Dan() (uint64, bool) { return c.unsigned(nbOffDan, 4) }

// SetDan moves v into DAN PIC 9(4), truncating high-order digits.
func (c *Comnba) SetDan(v uint64) { c.setUnsigned(nbOffDan, 4, v) }

// Ac3dir returns AC3DIR X(3) (CC2DEP + CCODIR).
func (c *Comnba) Ac3dir() string { return c.get(nbOffAc3dir, 3) }

// SetAc3dir moves v into AC3DIR X(3).
func (c *Comnba) SetAc3dir(v string) { c.put(nbOffAc3dir, 3, v) }

// Cc2dep returns CC2DEP X(2), the département code (bytes 5–6 of AIDFIC,
// which drive the département-57 rule downstream).
func (c *Comnba) Cc2dep() string { return c.get(nbOffCc2dep, 2) }

// SetCc2dep moves v into CC2DEP X(2).
func (c *Comnba) SetCc2dep(v string) { c.put(nbOffCc2dep, 2, v) }

// Ccodir returns CCODIR X, the direction code.
func (c *Comnba) Ccodir() string { return c.get(nbOffCcodir, 1) }

// SetCcodir moves v into CCODIR X.
func (c *Comnba) SetCcodir(v string) { c.put(nbOffCcodir, 1, v) }

// Ccocom returns CCOCOM X(3), the MAJIC2 commune code.
func (c *Comnba) Ccocom() string { return c.get(nbOffCcocom, 3) }

// SetCcocom moves v into CCOCOM X(3).
func (c *Comnba) SetCcocom(v string) { c.put(nbOffCcocom, 3, v) }

// Dsrpar returns DSRPAR X, the parcel roll series.
func (c *Comnba) Dsrpar() string { return c.get(nbOffDsrpar, 1) }

// SetDsrpar moves v into DSRPAR X.
func (c *Comnba) SetDsrpar(v string) { c.put(nbOffDsrpar, 1, v) }

// Anupro returns the raw 6-byte ANUPRO group (owner account, group moves).
func (c *Comnba) Anupro() string { return c.get(nbOffAnupro, ComnbaAnuproLen) }

// SetAnupro moves v into the 6-byte ANUPRO group.
func (c *Comnba) SetAnupro(v string) { c.put(nbOffAnupro, ComnbaAnuproLen, v) }

// Cgroup returns CGROUP X, the owner group code.
func (c *Comnba) Cgroup() string { return c.get(nbOffCgroup, 1) }

// SetCgroup moves v into CGROUP X.
func (c *Comnba) SetCgroup(v string) { c.put(nbOffCgroup, 1, v) }

// NnuproRaw returns the raw 5 bytes of NNUPRO PIC 9(5).
func (c *Comnba) NnuproRaw() string { return c.get(nbOffNnupro, 5) }

// Nnupro returns NNUPRO PIC 9(5), the owner account number; ok is the
// NUMERIC test.
func (c *Comnba) Nnupro() (uint64, bool) { return c.unsigned(nbOffNnupro, 5) }

// SetNnupro moves v into NNUPRO PIC 9(5), truncating high-order digits.
func (c *Comnba) SetNnupro(v uint64) { c.setUnsigned(nbOffNnupro, 5, v) }

// A0011 returns the raw 11-byte A0011 group (all GTOTAU indicators).
func (c *Comnba) A0011() string { return c.get(nbOffGtotau, ComnbaTauCount) }

// SetA0011 moves v into the 11-byte A0011 group.
func (c *Comnba) SetA0011(v string) { c.put(nbOffGtotau, ComnbaTauCount, v) }

// Gtotau returns GTOTAU(i+1) X (0-based i), a rate-applied indicator.
func (c *Comnba) Gtotau(i int) (string, error) {
	if err := nbCheckIndex("GTOTAU", i, ComnbaTauCount); err != nil {
		return "", err
	}
	return c.get(nbOffGtotau+i, 1), nil
}

// SetGtotau moves v into GTOTAU(i+1) X (0-based i).
func (c *Comnba) SetGtotau(i int, v string) error {
	if err := nbCheckIndex("GTOTAU", i, ComnbaTauCount); err != nil {
		return err
	}
	c.put(nbOffGtotau+i, 1, v)
	return nil
}

// ---- ABANB: non-bâti bases, PIC S9(10) ----
// Setters perform a numeric MOVE: high-order digits beyond 10 are truncated.

// Abanb returns the raw 120-byte ABANB group.
func (c *Comnba) Abanb() string { return c.get(nbOffAbanb, ComnbaAbanbLen) }

// SetAbanb moves v into the 120-byte ABANB group.
func (c *Comnba) SetAbanb(v string) { c.put(nbOffAbanb, ComnbaAbanbLen, v) }

// Mbacom returns MBACOM S9(10), the commune base; ok is the NUMERIC test.
func (c *Comnba) Mbacom() (int64, bool) { return c.amt(nbOffMbacom) }

// SetMbacom moves v into MBACOM S9(10).
func (c *Comnba) SetMbacom(v int64) { c.setAmt(nbOffMbacom, v) }

// MbacomRaw returns the raw 10 bytes of MBACOM.
func (c *Comnba) MbacomRaw() string { return c.amtRaw(nbOffMbacom) }

// Mbadep returns MBADEP S9(10), the département base.
func (c *Comnba) Mbadep() (int64, bool) { return c.amt(nbOffMbadep) }

// SetMbadep moves v into MBADEP S9(10).
func (c *Comnba) SetMbadep(v int64) { c.setAmt(nbOffMbadep, v) }

// MbadepRaw returns the raw 10 bytes of MBADEP.
func (c *Comnba) MbadepRaw() string { return c.amtRaw(nbOffMbadep) }

// Mbareg returns MBAREG S9(10), the region base.
func (c *Comnba) Mbareg() (int64, bool) { return c.amt(nbOffMbareg) }

// SetMbareg moves v into MBAREG S9(10).
func (c *Comnba) SetMbareg(v int64) { c.setAmt(nbOffMbareg, v) }

// MbaregRaw returns the raw 10 bytes of MBAREG.
func (c *Comnba) MbaregRaw() string { return c.amtRaw(nbOffMbareg) }

// Mbasyn returns MBASYN S9(10), the syndicate / district without FP base.
func (c *Comnba) Mbasyn() (int64, bool) { return c.amt(nbOffMbasyn) }

// SetMbasyn moves v into MBASYN S9(10).
func (c *Comnba) SetMbasyn(v int64) { c.setAmt(nbOffMbasyn, v) }

// MbasynRaw returns the raw 10 bytes of MBASYN.
func (c *Comnba) MbasynRaw() string { return c.amtRaw(nbOffMbasyn) }

// Mbacu returns MBACU S9(10), the CU / DFP / new agglomeration base.
func (c *Comnba) Mbacu() (int64, bool) { return c.amt(nbOffMbacu) }

// SetMbacu moves v into MBACU S9(10).
func (c *Comnba) SetMbacu(v int64) { c.setAmt(nbOffMbacu, v) }

// MbacuRaw returns the raw 10 bytes of MBACU.
func (c *Comnba) MbacuRaw() string { return c.amtRaw(nbOffMbacu) }

// Mbatse returns MBATSE S9(10), the TSE base.
func (c *Comnba) Mbatse() (int64, bool) { return c.amt(nbOffMbatse) }

// SetMbatse moves v into MBATSE S9(10).
func (c *Comnba) SetMbatse(v int64) { c.setAmt(nbOffMbatse, v) }

// MbatseRaw returns the raw 10 bytes of MBATSE.
func (c *Comnba) MbatseRaw() string { return c.amtRaw(nbOffMbatse) }

// Mbnt13 returns MBNT13(i+1) S9(10) (0-based i), the non-bâti TSE EPF base.
func (c *Comnba) Mbnt13(i int) (int64, bool, error) {
	if err := nbCheckIndex("MBNT13", i, ComnbaNt13Count); err != nil {
		return 0, false, err
	}
	v, ok := c.amt(nbOffMbnt13 + i*nbAmtLen)
	return v, ok, nil
}

// SetMbnt13 moves v into MBNT13(i+1) S9(10) (0-based i).
func (c *Comnba) SetMbnt13(i int, v int64) error {
	if err := nbCheckIndex("MBNT13", i, ComnbaNt13Count); err != nil {
		return err
	}
	c.setAmt(nbOffMbnt13+i*nbAmtLen, v)
	return nil
}

// Mbnt13Raw returns the raw 10 bytes of MBNT13(i+1) (0-based i).
func (c *Comnba) Mbnt13Raw(i int) (string, error) {
	if err := nbCheckIndex("MBNT13", i, ComnbaNt13Count); err != nil {
		return "", err
	}
	return c.amtRaw(nbOffMbnt13 + i*nbAmtLen), nil
}

// Mbaca returns MBACA S9(10), the chamber of agriculture base.
func (c *Comnba) Mbaca() (int64, bool) { return c.amt(nbOffMbaca) }

// SetMbaca moves v into MBACA S9(10).
func (c *Comnba) SetMbaca(v int64) { c.setAmt(nbOffMbaca, v) }

// MbacaRaw returns the raw 10 bytes of MBACA.
func (c *Comnba) MbacaRaw() string { return c.amtRaw(nbOffMbaca) }

// Mbabap returns MBABAP S9(10), the BAPSA base.
func (c *Comnba) Mbabap() (int64, bool) { return c.amt(nbOffMbabap) }

// SetMbabap moves v into MBABAP S9(10).
func (c *Comnba) SetMbabap(v int64) { c.setAmt(nbOffMbabap, v) }

// MbabapRaw returns the raw 10 bytes of MBABAP.
func (c *Comnba) MbabapRaw() string { return c.amtRaw(nbOffMbabap) }

// Mbacaa returns MBACAA S9(10), the CAAA base.
func (c *Comnba) Mbacaa() (int64, bool) { return c.amt(nbOffMbacaa) }

// SetMbacaa moves v into MBACAA S9(10).
func (c *Comnba) SetMbacaa(v int64) { c.setAmt(nbOffMbacaa, v) }

// MbacaaRaw returns the raw 10 bytes of MBACAA.
func (c *Comnba) MbacaaRaw() string { return c.amtRaw(nbOffMbacaa) }

// Mbapos returns MBAPOS S9(10), the building-land surcharge base.
func (c *Comnba) Mbapos() (int64, bool) { return c.amt(nbOffMbapos) }

// SetMbapos moves v into MBAPOS S9(10).
func (c *Comnba) SetMbapos(v int64) { c.setAmt(nbOffMbapos, v) }

// MbaposRaw returns the raw 10 bytes of MBAPOS.
func (c *Comnba) MbaposRaw() string { return c.amtRaw(nbOffMbapos) }

// ---- ABAJAG: young-farmer relief bases, PIC S9(10) ----

// Abajag returns the raw 120-byte ABAJAG group.
func (c *Comnba) Abajag() string { return c.get(nbOffAbajag, ComnbaAbajagLen) }

// SetAbajag moves v into the 120-byte ABAJAG group.
func (c *Comnba) SetAbajag(v string) { c.put(nbOffAbajag, ComnbaAbajagLen, v) }

// Mbjcom returns MBJCOM S9(10), the young-farmer commune base.
func (c *Comnba) Mbjcom() (int64, bool) { return c.amt(nbOffMbjcom) }

// SetMbjcom moves v into MBJCOM S9(10).
func (c *Comnba) SetMbjcom(v int64) { c.setAmt(nbOffMbjcom, v) }

// MbjcomRaw returns the raw 10 bytes of MBJCOM.
func (c *Comnba) MbjcomRaw() string { return c.amtRaw(nbOffMbjcom) }

// Mbjsyn returns MBJSYN S9(10), the young-farmer syndicate base.
func (c *Comnba) Mbjsyn() (int64, bool) { return c.amt(nbOffMbjsyn) }

// SetMbjsyn moves v into MBJSYN S9(10).
func (c *Comnba) SetMbjsyn(v int64) { c.setAmt(nbOffMbjsyn, v) }

// MbjsynRaw returns the raw 10 bytes of MBJSYN.
func (c *Comnba) MbjsynRaw() string { return c.amtRaw(nbOffMbjsyn) }

// Mbjcu returns MBJCU S9(10), the young-farmer CU base.
func (c *Comnba) Mbjcu() (int64, bool) { return c.amt(nbOffMbjcu) }

// SetMbjcu moves v into MBJCU S9(10).
func (c *Comnba) SetMbjcu(v int64) { c.setAmt(nbOffMbjcu, v) }

// MbjcuRaw returns the raw 10 bytes of MBJCU.
func (c *Comnba) MbjcuRaw() string { return c.amtRaw(nbOffMbjcu) }

// Mbjtse returns MBJTSE S9(10), the young-farmer TSE base.
func (c *Comnba) Mbjtse() (int64, bool) { return c.amt(nbOffMbjtse) }

// SetMbjtse moves v into MBJTSE S9(10).
func (c *Comnba) SetMbjtse(v int64) { c.setAmt(nbOffMbjtse, v) }

// MbjtseRaw returns the raw 10 bytes of MBJTSE.
func (c *Comnba) MbjtseRaw() string { return c.amtRaw(nbOffMbjtse) }

// Mbjt13 returns MBJT13(i+1) S9(10) (0-based i), the young-farmer TSE EPF base.
func (c *Comnba) Mbjt13(i int) (int64, bool, error) {
	if err := nbCheckIndex("MBJT13", i, ComnbaJt13Count); err != nil {
		return 0, false, err
	}
	v, ok := c.amt(nbOffMbjt13 + i*nbAmtLen)
	return v, ok, nil
}

// SetMbjt13 moves v into MBJT13(i+1) S9(10) (0-based i).
func (c *Comnba) SetMbjt13(i int, v int64) error {
	if err := nbCheckIndex("MBJT13", i, ComnbaJt13Count); err != nil {
		return err
	}
	c.setAmt(nbOffMbjt13+i*nbAmtLen, v)
	return nil
}

// Mbjt13Raw returns the raw 10 bytes of MBJT13(i+1) (0-based i).
func (c *Comnba) Mbjt13Raw(i int) (string, error) {
	if err := nbCheckIndex("MBJT13", i, ComnbaJt13Count); err != nil {
		return "", err
	}
	return c.amtRaw(nbOffMbjt13 + i*nbAmtLen), nil
}

// Mbjdep returns MBJDEP S9(10), the young-farmer département base.
func (c *Comnba) Mbjdep() (int64, bool) { return c.amt(nbOffMbjdep) }

// SetMbjdep moves v into MBJDEP S9(10).
func (c *Comnba) SetMbjdep(v int64) { c.setAmt(nbOffMbjdep, v) }

// MbjdepRaw returns the raw 10 bytes of MBJDEP.
func (c *Comnba) MbjdepRaw() string { return c.amtRaw(nbOffMbjdep) }

// Mb1pre returns MB1PRE S9(10), the département PRES base.
func (c *Comnba) Mb1pre() (int64, bool) { return c.amt(nbOffMb1pre) }

// SetMb1pre moves v into MB1PRE S9(10).
func (c *Comnba) SetMb1pre(v int64) { c.setAmt(nbOffMb1pre, v) }

// Mb1preRaw returns the raw 10 bytes of MB1PRE.
func (c *Comnba) Mb1preRaw() string { return c.amtRaw(nbOffMb1pre) }

// Mbjpre returns MBJPRE S9(10), the young-farmer / PRES département base.
func (c *Comnba) Mbjpre() (int64, bool) { return c.amt(nbOffMbjpre) }

// SetMbjpre moves v into MBJPRE S9(10).
func (c *Comnba) SetMbjpre(v int64) { c.setAmt(nbOffMbjpre, v) }

// MbjpreRaw returns the raw 10 bytes of MBJPRE.
func (c *Comnba) MbjpreRaw() string { return c.amtRaw(nbOffMbjpre) }

// Mbjeco returns MBJECO S9(10), the young-farmer State-borne commune base.
func (c *Comnba) Mbjeco() (int64, bool) { return c.amt(nbOffMbjeco) }

// SetMbjeco moves v into MBJECO S9(10).
func (c *Comnba) SetMbjeco(v int64) { c.setAmt(nbOffMbjeco, v) }

// MbjecoRaw returns the raw 10 bytes of MBJECO.
func (c *Comnba) MbjecoRaw() string { return c.amtRaw(nbOffMbjeco) }

// Mbjesy returns MBJESY S9(10), the young-farmer State-borne syndicate base.
func (c *Comnba) Mbjesy() (int64, bool) { return c.amt(nbOffMbjesy) }

// SetMbjesy moves v into MBJESY S9(10).
func (c *Comnba) SetMbjesy(v int64) { c.setAmt(nbOffMbjesy, v) }

// MbjesyRaw returns the raw 10 bytes of MBJESY.
func (c *Comnba) MbjesyRaw() string { return c.amtRaw(nbOffMbjesy) }

// Mbjeic returns MBJEIC S9(10), the young-farmer State-borne CU base.
func (c *Comnba) Mbjeic() (int64, bool) { return c.amt(nbOffMbjeic) }

// SetMbjeic moves v into MBJEIC S9(10).
func (c *Comnba) SetMbjeic(v int64) { c.setAmt(nbOffMbjeic, v) }

// MbjeicRaw returns the raw 10 bytes of MBJEIC.
func (c *Comnba) MbjeicRaw() string { return c.amtRaw(nbOffMbjeic) }

// ---- GEMAPI base, codes, trailer ----

// Mbage3 returns MBAGE3 S9(10), the GEMAPI base.
func (c *Comnba) Mbage3() (int64, bool) { return c.amt(nbOffMbage3) }

// SetMbage3 moves v into MBAGE3 S9(10).
func (c *Comnba) SetMbage3(v int64) { c.setAmt(nbOffMbage3, v) }

// Mbage3Raw returns the raw 10 bytes of MBAGE3.
func (c *Comnba) Mbage3Raw() string { return c.amtRaw(nbOffMbage3) }

// Ccoifp returns CCOIFP X(3), the absorbed-commune code.
func (c *Comnba) Ccoifp() string { return c.get(nbOffCcoifp, 3) }

// SetCcoifp moves v into CCOIFP X(3).
func (c *Comnba) SetCcoifp(v string) { c.put(nbOffCcoifp, 3, v) }

// Ccpper returns CCPPER X(3), the treasury code.
func (c *Comnba) Ccpper() string { return c.get(nbOffCcpper, 3) }

// SetCcpper moves v into CCPPER X(3).
func (c *Comnba) SetCcpper(v string) { c.put(nbOffCcpper, 3, v) }

// Trailer returns the trailing X(315) 'X'-FILLER reserved area.
func (c *Comnba) Trailer() string { return c.get(nbOffTrailer, ComnbaTrailerLen) }

// SetTrailer moves v into the trailing X(315) 'X'-FILLER reserved area.
func (c *Comnba) SetTrailer(v string) { c.put(nbOffTrailer, ComnbaTrailerLen, v) }
