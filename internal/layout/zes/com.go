package zes

import (
	"fmt"
)

// COMLen is the total length in bytes of a ZES COM (commune) segment
// (AIDFIC 13 + ADTSEG 387 = 400).
const COMLen = 400

// Scaling factors of the implied-decimal numeric fields of the COM segment.
const (
	// COMRateScale is the scaling factor of the PIC 9(4)V9(6) household-waste
	// rates PBBOMP..PBBOME and of the PIC 9(12)V9(6) pseudo-rate PBBOMI:
	// accessors return the value multiplied by 1 000 000 (six implied
	// decimals), e.g. 12.345678 -> 12345678.
	COMRateScale = 1_000_000
	// COMPplomScale is the scaling factor of PPLOM PIC 9V99: the accessor
	// returns the coefficient multiplied by 100 (1,50 -> 150).
	COMPplomScale = 100
)

// OCCURS counts of the COM segment.
const (
	// COMJaCount is the OCCURS count of JCOJA (JCOINS + JCORIM pairs) and of
	// JGCJA (JGCINS + JGCRIM pairs).
	COMJaCount = 5
)

// Group / area lengths of the COM segment.
const (
	// COMAidficLen is the length of the AIDFIC file-identifier group.
	COMAidficLen = 13
	// COMAdtsegLen is the length of the ADTSEG segment-body group.
	COMAdtsegLen = 387
	// COMAnuzipLen is the length of the ANUZIP group (five ZIP zones X(3)).
	COMAnuzipLen = 15
	// COMFiller1Len is the length of the first named 'X'-FILLER (after PBBOME).
	COMFiller1Len = 22
	// COMFiller2Len is the length of the second named 'X'-FILLER (before DNUURC).
	COMFiller2Len = 34
)

// Zero-based byte offsets of the COM fields inside the 400-byte segment.
//
//	AIDFIC        X(13)          0..12   (COBOL positions   1..13)
//	  JAN           9(4)           0..3
//	  DEPDIR        X(3)           4..6  = CODEP X(2) + CODIR X(1)
//	  CCOCOM        X(3)           7..9
//	  CCOIFP        X(3)          10..12
//	ADTSEG        X(387)        13..399  (COBOL positions  14..400)
//	  CCOENR        X(2)          13..14
//	  GTOVOT        X(1)          15
//	  XDDAY6        X(6)          16..21
//	  PMJPO         9(3)          22..24
//	  DNURO         9(3)          25..27
//	  JDHOMO        X(6)          28..33
//	  DNULTF        X(1)          34
//	  GCOANU        X(1)          35
//	  NBSROL        9(2)          36..37
//	  DLISRO        X(7)          38..44
//	  JDATAU        9(8)          45..52
//	  JDACMP        9(8)          53..60
//	  ACOGEM        X(4)          61..64
//	  PBBOMP        9(4)V9(6)     65..74
//	  PBBOMA        9(4)V9(6)     75..84
//	  PBBOMB        9(4)V9(6)     85..94
//	  PBBOMC        9(4)V9(6)     95..104
//	  PBBOMD        9(4)V9(6)    105..114
//	  NBTAUXOM      9(1)         115
//	  JCOJA OCCURS 5 (8 bytes)   116..155  = JCOINS 9(4) + JCORIM 9(4)
//	  JGCJA OCCURS 5 (8 bytes)   156..195  = JGCINS 9(4) + JGCRIM 9(4)
//	  DLICOM        X(30)        196..225
//	  GNIDEP        X(1)         226
//	  GTOECO        X(1)         227
//	  CTYGC         X(1)         228
//	  DNUCOL        X(3)         229..231
//	  ACOSYN        X(4)         232..235  = CTYSYN X(1) + CCOSYN X(3)
//	  ACOGOM        X(4)         236..239
//	  DVLPOM        9(6)         240..245
//	  PPLOM         9V99         246..248
//	  DCMZO1        X(1)         249
//	  ANUZIP        X(15)        250..264  = DNUZIP/DNUZIA/DNUZIB/DNUZIC/DNUZID X(3)
//	  NPOCOM        9(7)         265..271
//	  PBBOME        9(4)V9(6)    272..281
//	  FILLER        X(22)        282..303
//	  DSEDOM        9(9)         304..312
//	  DSEDNB        9(9)         313..321
//	  ACOTXA        X(4)         322..325  = CTYTXA X(1) + CCOTXA X(3)
//	  PBBOMI        9(12)V9(6)   326..343
//	  ACOBIM        X(4)         344..347
//	  GTOMGP        X(1)         348
//	  DSIREC        X(14)        349..362
//	  GTOCFU        X(1)         363
//	  FILLER        X(34)        364..397
//	  DNUURC        X(2)         398..399
const (
	comOffAidfic   = 0
	comOffJan      = 0
	comOffDepdir   = 4
	comOffCodep    = 4
	comOffCodir    = 6
	comOffCcocom   = 7
	comOffCcoifp   = 10
	comOffAdtseg   = 13
	comOffCcoenr   = 13
	comOffGtovot   = 15
	comOffXdday6   = 16
	comOffPmjpo    = 22
	comOffDnuro    = 25
	comOffJdhomo   = 28
	comOffDnultf   = 34
	comOffGcoanu   = 35
	comOffNbsrol   = 36
	comOffDlisro   = 38
	comOffJdatau   = 45
	comOffJdacmp   = 53
	comOffAcogem   = 61
	comOffPbbomp   = 65
	comOffPbboma   = 75
	comOffPbbomb   = 85
	comOffPbbomc   = 95
	comOffPbbomd   = 105
	comOffNbtauxom = 115
	comOffJcoja    = 116
	comOffJgcja    = 156
	comOffDlicom   = 196
	comOffGnidep   = 226
	comOffGtoeco   = 227
	comOffCtygc    = 228
	comOffDnucol   = 229
	comOffAcosyn   = 232
	comOffCtysyn   = 232
	comOffCcosyn   = 233
	comOffAcogom   = 236
	comOffDvlpom   = 240
	comOffPplom    = 246
	comOffDcmzo1   = 249
	comOffAnuzip   = 250
	comOffDnuzip   = 250
	comOffDnuzia   = 253
	comOffDnuzib   = 256
	comOffDnuzic   = 259
	comOffDnuzid   = 262
	comOffNpocom   = 265
	comOffPbbome   = 272
	comOffFiller1  = 282
	comOffDsedom   = 304
	comOffDsednb   = 313
	comOffAcotxa   = 322
	comOffCtytxa   = 322
	comOffCcotxa   = 323
	comOffPbbomi   = 326
	comOffAcobim   = 344
	comOffGtomgp   = 348
	comOffDsirec   = 349
	comOffGtocfu   = 363
	comOffFiller2  = 364
	comOffDnuurc   = 398

	comRateLen   = 10 // PIC 9(4)V9(6)
	comPbbomiLen = 18 // PIC 9(12)V9(6)
	comJaLen     = 8  // year 9(4) + year 9(4)
	comYearLen   = 4  // PIC 9(4)
	comDateLen   = 8  // PIC 9(8)
	comSeuilLen  = 9  // PIC 9(9)
	comZipLen    = 3  // PIC X(3)
)

// Compile-time layout checks: consecutive fields must be contiguous and the
// last field must end exactly at COMLen.
var (
	_ = [1]struct{}{}[comOffPbbomd+comRateLen-comOffNbtauxom]
	_ = [1]struct{}{}[comOffJcoja+COMJaCount*comJaLen-comOffJgcja]
	_ = [1]struct{}{}[comOffJgcja+COMJaCount*comJaLen-comOffDlicom]
	_ = [1]struct{}{}[comOffAnuzip+COMAnuzipLen-comOffNpocom]
	_ = [1]struct{}{}[comOffPbbome+comRateLen-comOffFiller1]
	_ = [1]struct{}{}[comOffFiller1+COMFiller1Len-comOffDsedom]
	_ = [1]struct{}{}[comOffPbbomi+comPbbomiLen-comOffAcobim]
	_ = [1]struct{}{}[comOffFiller2+COMFiller2Len-comOffDnuurc]
	_ = [1]struct{}{}[comOffDnuurc+2-COMLen]
)

// COM is the ZES COM segment (copybook XBXTDCOM): a fixed 400-byte MAJIC2
// "commune" record made of the AIDFIC file identifier followed by the
// ADTSEG body (role and homologation dates, household-waste TEOM rates,
// young-farmer years, collectivity / syndicate codes, ZIP zones, overseas
// exemption thresholds, SIRET and miscellaneous indicators).
//
// Fields are exposed as accessors over the raw storage, so group views
// (AIDFIC, DEPDIR, ADTSEG, ACOSYN, ANUZIP, ACOTXA, JCOJA, JGCJA) and their
// elementary fields share bytes exactly like COBOL storage; bytes never
// touched by a setter (including both FILLER areas) round-trip unchanged.
//
// MIGRATION_NOTE: the copybook uses an 'X' placeholder prefix substituted
// by COPY ... REPLACING. In Go each inclusion is simply a distinct COM
// value. 'X'-FILLER becomes a real data name after substitution, so both
// filler areas are exposed (Filler1/Filler2) to preserve their content.
//
// MIGRATION_NOTE: storage is the raw byte image (ASCII digits for the
// zoned-decimal fields; EBCDIC conversion, if any, belongs to the I/O
// layer). Unsigned zoned-decimal (USAGE DISPLAY) numerics with an implied
// decimal point are returned as scaled integers (see COMRateScale and
// COMPplomScale) together with the result of the COBOL NUMERIC class test;
// the arithmetic primitives decide what to do with invalid data. Setters
// take already-scaled values and truncate high-order digits, as a COBOL
// numeric MOVE into an unsigned receiver does.
type COM [COMLen]byte

// NewCOM returns a segment initialised to spaces, as a COBOL MOVE SPACES on
// the alphanumeric group would.
func NewCOM() COM {
	var c COM
	for i := range c {
		c[i] = ' '
	}
	return c
}

// ParseCOM copies a 400-byte buffer into a new COM segment. It fails if b
// is not exactly COMLen bytes long.
func ParseCOM(b []byte) (COM, error) {
	var c COM
	if len(b) != COMLen {
		return c, fmt.Errorf("zes: COM segment must be %d bytes, got %d", COMLen, len(b))
	}
	copy(c[:], b)
	return c, nil
}

// Bytes returns a copy of the raw 400 bytes of the segment.
func (c *COM) Bytes() []byte {
	out := make([]byte, COMLen)
	copy(out, c[:])
	return out
}

// ---- low-level helpers ----

func (c *COM) get(off, n int) string { return string(c[off : off+n]) }

// put performs a COBOL alphanumeric MOVE into a field: left-justified,
// truncated on the right, padded with spaces.
func (c *COM) put(off, n int, v string) {
	for i := 0; i < n; i++ {
		if i < len(v) {
			c[off+i] = v[i]
		} else {
			c[off+i] = ' '
		}
	}
}

// isNumeric applies the COBOL NUMERIC class test for an unsigned
// zoned-decimal field (every byte a digit).
func (c *COM) isNumeric(off, n int) bool {
	for _, b := range c[off : off+n] {
		if b < '0' || b > '9' {
			return false
		}
	}
	return true
}

// unsigned decodes an unsigned zoned-decimal field as its digit string
// value (implied decimals included, i.e. already scaled). ok is false when
// the field is not NUMERIC; the value is then built from the low nibble of
// each byte (clamped to 9), as IBM COBOL arithmetic on invalid zoned data
// would, and must be treated as unreliable.
func (c *COM) unsigned(off, n int) (value uint64, ok bool) {
	ok = c.isNumeric(off, n)
	for _, b := range c[off : off+n] {
		dg := uint64(b & 0x0F)
		if dg > 9 {
			dg = 9 // MIGRATION_NOTE: invalid nibble; clamp to keep result in range
		}
		value = value*10 + dg
	}
	return value, ok
}

// setUnsigned performs a COBOL numeric MOVE of an already-scaled value into
// an unsigned zoned-decimal field: high-order digits beyond the field
// width are truncated.
func (c *COM) setUnsigned(off, n int, v uint64) {
	for i := n - 1; i >= 0; i-- {
		c[off+i] = byte('0' + v%10)
		v /= 10
	}
}

func comCheckIndex(field string, i, count int) error {
	if i < 0 || i >= count {
		return fmt.Errorf("zes: COM %s subscript %d out of range [0,%d)", field, i, count)
	}
	return nil
}

// ---- AIDFIC: file identifier (positions 1-13) ----

// Aidfic returns the 13-byte AIDFIC group (MAJIC2 file identifier).
func (c *COM) Aidfic() string { return c.get(comOffAidfic, COMAidficLen) }

// SetAidfic moves v into the 13-byte AIDFIC group.
func (c *COM) SetAidfic(v string) { c.put(comOffAidfic, COMAidficLen, v) }

// JanRaw returns the raw 4 bytes of JAN PIC 9(4) (year).
func (c *COM) JanRaw() string { return c.get(comOffJan, comYearLen) }

// JanIsNumeric reports whether JAN passes the COBOL NUMERIC class test.
func (c *COM) JanIsNumeric() bool { return c.isNumeric(comOffJan, comYearLen) }

// Jan returns JAN PIC 9(4) (year); ok is false when the field is not NUMERIC.
func (c *COM) Jan() (uint64, bool) { return c.unsigned(comOffJan, comYearLen) }

// SetJan performs a COBOL numeric MOVE into JAN PIC 9(4).
func (c *COM) SetJan(v uint64) { c.setUnsigned(comOffJan, comYearLen, v) }

// Depdir returns DEPDIR X(3): département + direction group.
func (c *COM) Depdir() string { return c.get(comOffDepdir, 3) }

// SetDepdir moves v into DEPDIR X(3).
func (c *COM) SetDepdir(v string) { c.put(comOffDepdir, 3, v) }

// Codep returns CODEP X(2): département code.
func (c *COM) Codep() string { return c.get(comOffCodep, 2) }

// SetCodep moves v into CODEP X(2).
func (c *COM) SetCodep(v string) { c.put(comOffCodep, 2, v) }

// Codir returns CODIR X(1): direction code.
func (c *COM) Codir() string { return c.get(comOffCodir, 1) }

// SetCodir moves v into CODIR X(1).
func (c *COM) SetCodir(v string) { c.put(comOffCodir, 1, v) }

// Ccocom returns CCOCOM X(3): MAJIC2 commune code.
func (c *COM) Ccocom() string { return c.get(comOffCcocom, 3) }

// SetCcocom moves v into CCOCOM X(3).
func (c *COM) SetCcocom(v string) { c.put(comOffCcocom, 3, v) }

// Ccoifp returns CCOIFP X(3): absorbed-commune code.
func (c *COM) Ccoifp() string { return c.get(comOffCcoifp, 3) }

// SetCcoifp moves v into CCOIFP X(3).
func (c *COM) SetCcoifp(v string) { c.put(comOffCcoifp, 3, v) }

// ---- ADTSEG: segment body (positions 14-400) ----

// Adtseg returns the 387-byte ADTSEG group (segment body).
func (c *COM) Adtseg() string { return c.get(comOffAdtseg, COMAdtsegLen) }

// SetAdtseg moves v into the 387-byte ADTSEG group.
func (c *COM) SetAdtseg(v string) { c.put(comOffAdtseg, COMAdtsegLen, v) }

// Ccoenr returns CCOENR X(2): record code.
func (c *COM) Ccoenr() string { return c.get(comOffCcoenr, 2) }

// SetCcoenr moves v into CCOENR X(2).
func (c *COM) SetCcoenr(v string) { c.put(comOffCcoenr, 2, v) }

// Gtovot returns GTOVOT X(1): valid-rates indicator.
func (c *COM) Gtovot() string { return c.get(comOffGtovot, 1) }

// SetGtovot moves v into GTOVOT X(1).
func (c *COM) SetGtovot(v string) { c.put(comOffGtovot, 1, v) }

// Xdday6 returns XDDAY6 X(6): date of the day on 6 characters.
func (c *COM) Xdday6() string { return c.get(comOffXdday6, 6) }

// SetXdday6 moves v into XDDAY6 X(6).
func (c *COM) SetXdday6(v string) { c.put(comOffXdday6, 6, v) }

// Pmjpo returns PMJPO PIC 9(3): building-land rental-value surcharge
// amount; ok is false when the field is not NUMERIC.
func (c *COM) Pmjpo() (uint64, bool) { return c.unsigned(comOffPmjpo, 3) }

// SetPmjpo performs a COBOL numeric MOVE into PMJPO PIC 9(3).
func (c *COM) SetPmjpo(v uint64) { c.setUnsigned(comOffPmjpo, 3, v) }

// Dnuro returns DNURO PIC 9(3): role number; ok is false when the field is
// not NUMERIC.
func (c *COM) Dnuro() (uint64, bool) { return c.unsigned(comOffDnuro, 3) }

// SetDnuro performs a COBOL numeric MOVE into DNURO PIC 9(3).
func (c *COM) SetDnuro(v uint64) { c.setUnsigned(comOffDnuro, 3, v) }

// Jdhomo returns JDHOMO X(6): role homologation date (JJMMAA).
func (c *COM) Jdhomo() string { return c.get(comOffJdhomo, 6) }

// SetJdhomo moves v into JDHOMO X(6).
func (c *COM) SetJdhomo(v string) { c.put(comOffJdhomo, 6, v) }

// Dnultf returns DNULTF X(1): land-tax taxation batch number.
func (c *COM) Dnultf() string { return c.get(comOffDnultf, 1) }

// SetDnultf moves v into DNULTF X(1).
func (c *COM) SetDnultf(v string) { c.put(comOffDnultf, 1, v) }

// Gcoanu returns GCOANU X(1): cancellation / redo code.
func (c *COM) Gcoanu() string { return c.get(comOffGcoanu, 1) }

// SetGcoanu moves v into GCOANU X(1).
func (c *COM) SetGcoanu(v string) { c.put(comOffGcoanu, 1, v) }

// Nbsrol returns NBSROL PIC 9(2): number of series-roles; ok is false when
// the field is not NUMERIC.
func (c *COM) Nbsrol() (uint64, bool) { return c.unsigned(comOffNbsrol, 2) }

// SetNbsrol performs a COBOL numeric MOVE into NBSROL PIC 9(2).
func (c *COM) SetNbsrol(v uint64) { c.setUnsigned(comOffNbsrol, 2, v) }

// Dlisro returns DLISRO X(7): list of series-roles.
func (c *COM) Dlisro() string { return c.get(comOffDlisro, 7) }

// SetDlisro moves v into DLISRO X(7).
func (c *COM) SetDlisro(v string) { c.put(comOffDlisro, 7, v) }

// Jdatau returns JDATAU PIC 9(8): rates validity date; ok is false when the
// field is not NUMERIC.
func (c *COM) Jdatau() (uint64, bool) { return c.unsigned(comOffJdatau, comDateLen) }

// SetJdatau performs a COBOL numeric MOVE into JDATAU PIC 9(8).
func (c *COM) SetJdatau(v uint64) { c.setUnsigned(comOffJdatau, comDateLen, v) }

// Jdacmp returns JDACMP PIC 9(8): supplementary role date; ok is false when
// the field is not NUMERIC.
func (c *COM) Jdacmp() (uint64, bool) { return c.unsigned(comOffJdacmp, comDateLen) }

// SetJdacmp performs a COBOL numeric MOVE into JDACMP PIC 9(8).
func (c *COM) SetJdacmp(v uint64) { c.setUnsigned(comOffJdacmp, comDateLen, v) }

// Acogem returns ACOGEM X(4): GEMAPI beneficiary code.
func (c *COM) Acogem() string { return c.get(comOffAcogem, 4) }

// SetAcogem moves v into ACOGEM X(4).
func (c *COM) SetAcogem(v string) { c.put(comOffAcogem, 4, v) }

// ---- household-waste (OM) rates PIC 9(4)V9(6), scaled by COMRateScale ----

// Pbbomp returns PBBOMP, the full household-waste rate, scaled by
// COMRateScale; ok is false when the field is not NUMERIC.
func (c *COM) Pbbomp() (uint64, bool) { return c.unsigned(comOffPbbomp, comRateLen) }

// SetPbbomp moves a COMRateScale-scaled value into PBBOMP.
func (c *COM) SetPbbomp(v uint64) { c.setUnsigned(comOffPbbomp, comRateLen, v) }

// Pbboma returns PBBOMA, the reduced RA household-waste rate, scaled by
// COMRateScale.
func (c *COM) Pbboma() (uint64, bool) { return c.unsigned(comOffPbboma, comRateLen) }

// SetPbboma moves a COMRateScale-scaled value into PBBOMA.
func (c *COM) SetPbboma(v uint64) { c.setUnsigned(comOffPbboma, comRateLen, v) }

// Pbbomb returns PBBOMB, the reduced RB household-waste rate, scaled by
// COMRateScale.
func (c *COM) Pbbomb() (uint64, bool) { return c.unsigned(comOffPbbomb, comRateLen) }

// SetPbbomb moves a COMRateScale-scaled value into PBBOMB.
func (c *COM) SetPbbomb(v uint64) { c.setUnsigned(comOffPbbomb, comRateLen, v) }

// Pbbomc returns PBBOMC, the reduced RC household-waste rate, scaled by
// COMRateScale.
func (c *COM) Pbbomc() (uint64, bool) { return c.unsigned(comOffPbbomc, comRateLen) }

// SetPbbomc moves a COMRateScale-scaled value into PBBOMC.
func (c *COM) SetPbbomc(v uint64) { c.setUnsigned(comOffPbbomc, comRateLen, v) }

// Pbbomd returns PBBOMD, the reduced RD household-waste rate, scaled by
// COMRateScale.
func (c *COM) Pbbomd() (uint64, bool) { return c.unsigned(comOffPbbomd, comRateLen) }

// SetPbbomd moves a COMRateScale-scaled value into PBBOMD.
func (c *COM) SetPbbomd(v uint64) { c.setUnsigned(comOffPbbomd, comRateLen, v) }

// Pbbome returns PBBOME, the reduced RE household-waste rate, scaled by
// COMRateScale.
func (c *COM) Pbbome() (uint64, bool) { return c.unsigned(comOffPbbome, comRateLen) }

// SetPbbome moves a COMRateScale-scaled value into PBBOME.
func (c *COM) SetPbbome(v uint64) { c.setUnsigned(comOffPbbome, comRateLen, v) }

// Pbbomi returns PBBOMI PIC 9(12)V9(6), the TEOMI pseudo-rate for new
// constructions, scaled by COMRateScale (18 digits always fit a uint64).
func (c *COM) Pbbomi() (uint64, bool) { return c.unsigned(comOffPbbomi, comPbbomiLen) }

// SetPbbomi moves a COMRateScale-scaled value into PBBOMI.
func (c *COM) SetPbbomi(v uint64) { c.setUnsigned(comOffPbbomi, comPbbomiLen, v) }

// Nbtauxom returns NBTAUXOM PIC 9(1): number of household-waste rates; ok
// is false when the field is not NUMERIC.
func (c *COM) Nbtauxom() (uint64, bool) { return c.unsigned(comOffNbtauxom, 1) }

// SetNbtauxom performs a COBOL numeric MOVE into NBTAUXOM PIC 9(1).
func (c *COM) SetNbtauxom(v uint64) { c.setUnsigned(comOffNbtauxom, 1, v) }

// ---- JCOJA / JGCJA OCCURS 5: young-farmer (JA) years ----

// Jcoja returns the raw 8-byte JCOJA(i+1) group (JCOINS + JCORIM) for the
// 0-based index i.
func (c *COM) Jcoja(i int) (string, error) {
	if err := comCheckIndex("JCOJA", i, COMJaCount); err != nil {
		return "", err
	}
	return c.get(comOffJcoja+i*comJaLen, comJaLen), nil
}

// SetJcoja moves v into the 8-byte JCOJA(i+1) group (0-based i).
func (c *COM) SetJcoja(i int, v string) error {
	if err := comCheckIndex("JCOJA", i, COMJaCount); err != nil {
		return err
	}
	c.put(comOffJcoja+i*comJaLen, comJaLen, v)
	return nil
}

// Jcoins returns JCOINS(i+1) PIC 9(4) (0-based i), the commune JA
// installation year; ok is false when the field is not NUMERIC.
func (c *COM) Jcoins(i int) (uint64, bool, error) {
	return c.jaYear("JCOINS", comOffJcoja, i, 0)
}

// SetJcoins moves v into JCOINS(i+1) PIC 9(4) (0-based i).
func (c *COM) SetJcoins(i int, v uint64) error {
	return c.setJaYear("JCOINS", comOffJcoja, i, 0, v)
}

// Jcorim returns JCORIM(i+1) PIC 9(4) (0-based i), the commune year of
// return to taxation; ok is false when the field is not NUMERIC.
func (c *COM) Jcorim(i int) (uint64, bool, error) {
	return c.jaYear("JCORIM", comOffJcoja, i, comYearLen)
}

// SetJcorim moves v into JCORIM(i+1) PIC 9(4) (0-based i).
func (c *COM) SetJcorim(i int, v uint64) error {
	return c.setJaYear("JCORIM", comOffJcoja, i, comYearLen, v)
}

// Jgcja returns the raw 8-byte JGCJA(i+1) group (JGCINS + JGCRIM) for the
// 0-based index i.
func (c *COM) Jgcja(i int) (string, error) {
	if err := comCheckIndex("JGCJA", i, COMJaCount); err != nil {
		return "", err
	}
	return c.get(comOffJgcja+i*comJaLen, comJaLen), nil
}

// SetJgcja moves v into the 8-byte JGCJA(i+1) group (0-based i).
func (c *COM) SetJgcja(i int, v string) error {
	if err := comCheckIndex("JGCJA", i, COMJaCount); err != nil {
		return err
	}
	c.put(comOffJgcja+i*comJaLen, comJaLen, v)
	return nil
}

// Jgcins returns JGCINS(i+1) PIC 9(4) (0-based i), the commune-grouping JA
// installation year; ok is false when the field is not NUMERIC.
func (c *COM) Jgcins(i int) (uint64, bool, error) {
	return c.jaYear("JGCINS", comOffJgcja, i, 0)
}

// SetJgcins moves v into JGCINS(i+1) PIC 9(4) (0-based i).
func (c *COM) SetJgcins(i int, v uint64) error {
	return c.setJaYear("JGCINS", comOffJgcja, i, 0, v)
}

// Jgcrim returns JGCRIM(i+1) PIC 9(4) (0-based i), the commune-grouping
// year of return to taxation; ok is false when the field is not NUMERIC.
func (c *COM) Jgcrim(i int) (uint64, bool, error) {
	return c.jaYear("JGCRIM", comOffJgcja, i, comYearLen)
}

// SetJgcrim moves v into JGCRIM(i+1) PIC 9(4) (0-based i).
func (c *COM) SetJgcrim(i int, v uint64) error {
	return c.setJaYear("JGCRIM", comOffJgcja, i, comYearLen, v)
}

func (c *COM) jaYear(field string, base, i, rel int) (uint64, bool, error) {
	if err := comCheckIndex(field, i, COMJaCount); err != nil {
		return 0, false, err
	}
	v, ok := c.unsigned(base+i*comJaLen+rel, comYearLen)
	return v, ok, nil
}

func (c *COM) setJaYear(field string, base, i, rel int, v uint64) error {
	if err := comCheckIndex(field, i, COMJaCount); err != nil {
		return err
	}
	c.setUnsigned(base+i*comJaLen+rel, comYearLen, v)
	return nil
}

// ---- commune identification and indicators ----

// Dlicom returns DLICOM X(30): commune label.
func (c *COM) Dlicom() string { return c.get(comOffDlicom, 30) }

// SetDlicom moves v into DLICOM X(30).
func (c *COM) SetDlicom(v string) { c.put(comOffDlicom, 30, v) }

// Gnidep returns GNIDEP X(1): commune not taxable by the département flag.
func (c *COM) Gnidep() string { return c.get(comOffGnidep, 1) }

// SetGnidep moves v into GNIDEP X(1).
func (c *COM) SetGnidep(v string) { c.put(comOffGnidep, 1, v) }

// Gtoeco returns GTOECO X(1): school rate base indicator.
func (c *COM) Gtoeco() string { return c.get(comOffGtoeco, 1) }

// SetGtoeco moves v into GTOECO X(1).
func (c *COM) SetGtoeco(v string) { c.put(comOffGtoeco, 1, v) }

// Ctygc returns CTYGC X(1): commune grouping type.
func (c *COM) Ctygc() string { return c.get(comOffCtygc, 1) }

// SetCtygc moves v into CTYGC X(1).
func (c *COM) SetCtygc(v string) { c.put(comOffCtygc, 1, v) }

// Dnucol returns DNUCOL X(3): local collectivity code.
func (c *COM) Dnucol() string { return c.get(comOffDnucol, 3) }

// SetDnucol moves v into DNUCOL X(3).
func (c *COM) SetDnucol(v string) { c.put(comOffDnucol, 3, v) }

// Acosyn returns ACOSYN X(4): household-waste beneficiary syndicate group
// (CTYSYN + CCOSYN).
func (c *COM) Acosyn() string { return c.get(comOffAcosyn, 4) }

// SetAcosyn moves v into the ACOSYN X(4) group.
func (c *COM) SetAcosyn(v string) { c.put(comOffAcosyn, 4, v) }

// Ctysyn returns CTYSYN X(1): household-waste syndicate type.
func (c *COM) Ctysyn() string { return c.get(comOffCtysyn, 1) }

// SetCtysyn moves v into CTYSYN X(1).
func (c *COM) SetCtysyn(v string) { c.put(comOffCtysyn, 1, v) }

// Ccosyn returns CCOSYN X(3): household-waste syndicate code.
func (c *COM) Ccosyn() string { return c.get(comOffCcosyn, 3) }

// SetCcosyn moves v into CCOSYN X(3).
func (c *COM) SetCcosyn(v string) { c.put(comOffCcosyn, 3, v) }

// Acogom returns ACOGOM X(4): household-waste managing collectivity.
func (c *COM) Acogom() string { return c.get(comOffAcogom, 4) }

// SetAcogom moves v into ACOGOM X(4).
func (c *COM) SetAcogom(v string) { c.put(comOffAcogom, 4, v) }

// Dvlpom returns DVLPOM PIC 9(6): commune average rental value; ok is false
// when the field is not NUMERIC.
func (c *COM) Dvlpom() (uint64, bool) { return c.unsigned(comOffDvlpom, 6) }

// SetDvlpom performs a COBOL numeric MOVE into DVLPOM PIC 9(6).
func (c *COM) SetDvlpom(v uint64) { c.setUnsigned(comOffDvlpom, 6, v) }

// Pplom returns PPLOM PIC 9V99, the TEOM ceiling coefficient, scaled by
// COMPplomScale; ok is false when the field is not NUMERIC.
func (c *COM) Pplom() (uint64, bool) { return c.unsigned(comOffPplom, 3) }

// SetPplom moves a COMPplomScale-scaled value into PPLOM.
func (c *COM) SetPplom(v uint64) { c.setUnsigned(comOffPplom, 3, v) }

// Dcmzo1 returns DCMZO1 X(1): national TC surcharge indicator.
func (c *COM) Dcmzo1() string { return c.get(comOffDcmzo1, 1) }

// SetDcmzo1 moves v into DCMZO1 X(1).
func (c *COM) SetDcmzo1(v string) { c.put(comOffDcmzo1, 1, v) }

// ---- ANUZIP: household-waste ZIP zones ----

// Anuzip returns the 15-byte ANUZIP group (DNUZIP..DNUZID).
func (c *COM) Anuzip() string { return c.get(comOffAnuzip, COMAnuzipLen) }

// SetAnuzip moves v into the 15-byte ANUZIP group.
func (c *COM) SetAnuzip(v string) { c.put(comOffAnuzip, COMAnuzipLen, v) }

// Dnuzip returns DNUZIP X(3): ZIP number of household-waste zone P.
func (c *COM) Dnuzip() string { return c.get(comOffDnuzip, comZipLen) }

// SetDnuzip moves v into DNUZIP X(3).
func (c *COM) SetDnuzip(v string) { c.put(comOffDnuzip, comZipLen, v) }

// Dnuzia returns DNUZIA X(3): ZIP number of household-waste zone RA.
func (c *COM) Dnuzia() string { return c.get(comOffDnuzia, comZipLen) }

// SetDnuzia moves v into DNUZIA X(3).
func (c *COM) SetDnuzia(v string) { c.put(comOffDnuzia, comZipLen, v) }

// Dnuzib returns DNUZIB X(3): ZIP number of household-waste zone RB.
func (c *COM) Dnuzib() string { return c.get(comOffDnuzib, comZipLen) }

// SetDnuzib moves v into DNUZIB X(3).
func (c *COM) SetDnuzib(v string) { c.put(comOffDnuzib, comZipLen, v) }

// Dnuzic returns DNUZIC X(3): ZIP number of household-waste zone RC.
func (c *COM) Dnuzic() string { return c.get(comOffDnuzic, comZipLen) }

// SetDnuzic moves v into DNUZIC X(3).
func (c *COM) SetDnuzic(v string) { c.put(comOffDnuzic, comZipLen, v) }

// Dnuzid returns DNUZID X(3): ZIP number of household-waste zone RD.
func (c *COM) Dnuzid() string { return c.get(comOffDnuzid, comZipLen) }

// SetDnuzid moves v into DNUZID X(3).
func (c *COM) SetDnuzid(v string) { c.put(comOffDnuzid, comZipLen, v) }

// Npocom returns NPOCOM PIC 9(7): commune total population; ok is false
// when the field is not NUMERIC.
func (c *COM) Npocom() (uint64, bool) { return c.unsigned(comOffNpocom, 7) }

// SetNpocom performs a COBOL numeric MOVE into NPOCOM PIC 9(7).
func (c *COM) SetNpocom(v uint64) { c.setUnsigned(comOffNpocom, 7, v) }

// ---- fillers ----

// Filler1 returns the 22-byte named 'X'-FILLER area following PBBOME.
func (c *COM) Filler1() string { return c.get(comOffFiller1, COMFiller1Len) }

// SetFiller1 moves v into the 22-byte 'X'-FILLER area following PBBOME.
func (c *COM) SetFiller1(v string) { c.put(comOffFiller1, COMFiller1Len, v) }

// Filler2 returns the 34-byte named 'X'-FILLER area preceding DNUURC.
func (c *COM) Filler2() string { return c.get(comOffFiller2, COMFiller2Len) }

// SetFiller2 moves v into the 34-byte 'X'-FILLER area preceding DNUURC.
func (c *COM) SetFiller2(v string) { c.put(comOffFiller2, COMFiller2Len, v) }

// ---- overseas (DOM) exemption thresholds ----

// Dsedom returns DSEDOM PIC 9(9): DOM exemption threshold; ok is false when
// the field is not NUMERIC.
func (c *COM) Dsedom() (uint64, bool) { return c.unsigned(comOffDsedom, comSeuilLen) }

// SetDsedom performs a COBOL numeric MOVE into DSEDOM PIC 9(9).
func (c *COM) SetDsedom(v uint64) { c.setUnsigned(comOffDsedom, comSeuilLen, v) }

// Dsednb returns DSEDNB PIC 9(9): DOM non-built exemption threshold; ok is
// false when the field is not NUMERIC.
func (c *COM) Dsednb() (uint64, bool) { return c.unsigned(comOffDsednb, comSeuilLen) }

// SetDsednb performs a COBOL numeric MOVE into DSEDNB PIC 9(9).
func (c *COM) SetDsednb(v uint64) { c.setUnsigned(comOffDsednb, comSeuilLen, v) }

// ---- ACOTXA: additional-tax beneficiary ----

// Acotxa returns ACOTXA X(4): additional-tax beneficiary group
// (CTYTXA + CCOTXA).
func (c *COM) Acotxa() string { return c.get(comOffAcotxa, 4) }

// SetAcotxa moves v into the ACOTXA X(4) group.
func (c *COM) SetAcotxa(v string) { c.put(comOffAcotxa, 4, v) }

// Ctytxa returns CTYTXA X(1): additional-tax beneficiary type.
func (c *COM) Ctytxa() string { return c.get(comOffCtytxa, 1) }

// SetCtytxa moves v into CTYTXA X(1).
func (c *COM) SetCtytxa(v string) { c.put(comOffCtytxa, 1, v) }

// Ccotxa returns CCOTXA X(3): additional-tax beneficiary code.
func (c *COM) Ccotxa() string { return c.get(comOffCcotxa, 3) }

// SetCcotxa moves v into CCOTXA X(3).
func (c *COM) SetCcotxa(v string) { c.put(comOffCcotxa, 3, v) }

// ---- miscellaneous ----

// Acobim returns ACOBIM X(4): TEOMI beneficiary code.
func (c *COM) Acobim() string { return c.get(comOffAcobim, 4) }

// SetAcobim moves v into ACOBIM X(4).
func (c *COM) SetAcobim(v string) { c.put(comOffAcobim, 4, v) }

// Gtomgp returns GTOMGP X(1): commune belongs to the Grand Paris flag.
func (c *COM) Gtomgp() string { return c.get(comOffGtomgp, 1) }

// SetGtomgp moves v into GTOMGP X(1).
func (c *COM) SetGtomgp(v string) { c.put(comOffGtomgp, 1, v) }

// Dsirec returns DSIREC X(14): commune SIRET number.
func (c *COM) Dsirec() string { return c.get(comOffDsirec, 14) }

// SetDsirec moves v into DSIREC X(14).
func (c *COM) SetDsirec(v string) { c.put(comOffDsirec, 14, v) }

// Gtocfu returns GTOCFU X(1): first year of commune merger flag.
func (c *COM) Gtocfu() string { return c.get(comOffGtocfu, 1) }

// SetGtocfu moves v into GTOCFU X(1).
func (c *COM) SetGtocfu(v string) { c.put(comOffGtocfu, 1, v) }

// Dnuurc returns DNUURC X(2): CDIF regrouping unit number (UR).
func (c *COM) Dnuurc() string { return c.get(comOffDnuurc, 2) }

// SetDnuurc moves v into DNUURC X(2).
func (c *COM) SetDnuurc(v string) { c.put(comOffDnuurc, 2, v) }
