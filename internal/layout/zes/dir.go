package zes

import (
	"fmt"
)

// DIRLen is the total length in bytes of a ZES DIR segment
// (AIDFIC 13 + ADTSEG 387 = 400).
const DIRLen = 400

// Scaling factors of the implied-decimal numeric fields of the DIR segment.
const (
	// DIRRateScale is the scaling factor of the PIC 9(4)V9(6) rate fields:
	// accessors return the rate multiplied by 1 000 000 (six implied
	// decimals), e.g. 12.345678 -> 12345678.
	DIRRateScale = 1_000_000
	// DIRBapsaScale is the scaling factor of TAUBAP405-NB PIC 9(2)V9(2):
	// the accessor returns the rate multiplied by 100 (4,05 -> 405).
	DIRBapsaScale = 100
)

// OCCURS counts of the DIR segment.
const (
	// DIRJaCount is the OCCURS count of JDEJA (JDEINS + JDERIM pairs).
	DIRJaCount = 5
	// DIRTseCount is the OCCURS count of ACOETS.
	DIRTseCount = 5
)

// Group / area lengths of the DIR segment.
const (
	// DIRAidficLen is the length of the AIDFIC file-identifier group.
	DIRAidficLen = 13
	// DIRAdtsegLen is the length of the ADTSEG segment-body group.
	DIRAdtsegLen = 387
	// DIRFiller1Len is the length of the first named 'X'-FILLER (after JDEJA).
	DIRFiller1Len = 9
	// DIRFiller2Len is the length of the second named 'X'-FILLER (before DNUURC).
	DIRFiller2Len = 179
)

// Zero-based byte offsets of the DIR fields inside the 400-byte segment.
//
//	AIDFIC        X(13)        0..12   (COBOL positions   1..13)
//	  JAN           9(4)         0..3
//	  DEPDIR        X(3)         4..6  = CODEP X(2) + CODIR X(1)
//	  CCOCOM        X(3)         7..9
//	  CCOIFP        X(3)        10..12
//	ADTSEG        X(387)      13..399  (COBOL positions  14..400)
//	  CCOENR        X(2)        13..14
//	  XDDAY6        X(6)        15..20
//	  TAUDEP-B      9(4)V9(6)   21..30
//	  TAUREG-B      9(4)V9(6)   31..40
//	  PTBTAS        9(4)V9(6)   41..50
//	  TAUDEP-NB     9(4)V9(6)   51..60
//	  TAUREG-NB     9(4)V9(6)   61..70
//	  TAUCHAGR-NB   9(4)V9(6)   71..80
//	  TAUBAP405-NB  9(2)V9(2)   81..84
//	  TAUSMAGR-NB   9(4)V9(6)   85..94
//	  JDEJA OCCURS 5 (8 bytes)  95..134  = JDEINS 9(4) + JDERIM 9(4)
//	  FILLER        X(9)       135..143
//	  ACOETS OCCURS 5 X(4)     144..163
//	  NBTSE         9(2)       164..165
//	  PTBDRB        9(4)V9(6)  166..175
//	  DSITAS        X(14)      176..189
//	  GROTAS        X(1)       190
//	  DSIRED        X(14)      191..204
//	  DSICA         X(14)      205..218
//	  FILLER        X(179)     219..397
//	  DNUURC        X(2)       398..399
const (
	dirOffAidfic      = 0
	dirOffJan         = 0
	dirOffDepdir      = 4
	dirOffCodep       = 4
	dirOffCodir       = 6
	dirOffCcocom      = 7
	dirOffCcoifp      = 10
	dirOffAdtseg      = 13
	dirOffCcoenr      = 13
	dirOffXdday6      = 15
	dirOffTaudepB     = 21
	dirOffTauregB     = 31
	dirOffPtbtas      = 41
	dirOffTaudepNB    = 51
	dirOffTauregNB    = 61
	dirOffTauchagrNB  = 71
	dirOffTaubap405NB = 81
	dirOffTausmagrNB  = 85
	dirOffJdeja       = 95
	dirOffFiller1     = 135
	dirOffAcoets      = 144
	dirOffNbtse       = 164
	dirOffPtbdrb      = 166
	dirOffDsitas      = 176
	dirOffGrotas      = 190
	dirOffDsired      = 191
	dirOffDsica       = 205
	dirOffFiller2     = 219
	dirOffDnuurc      = 398

	dirRateLen   = 10 // PIC 9(4)V9(6)
	dirBapsaLen  = 4  // PIC 9(2)V9(2)
	dirJdejaLen  = 8  // JDEINS 9(4) + JDERIM 9(4)
	dirYearLen   = 4  // PIC 9(4)
	dirAcoetsLen = 4  // PIC X(4)
	dirNbtseLen  = 2  // PIC 9(2)
	dirSiretLen  = 14 // PIC X(14)
)

// Compile-time layout check: the last field must end exactly at DIRLen.
var _ = [1]struct{}{}[dirOffDnuurc+2-DIRLen]

// DIR is the ZES DIR segment (copybook XBXTDDIR): a fixed 400-byte MAJIC2
// "direction / département" rates record made of the AIDFIC file identifier
// followed by the ADTSEG body (département / région / TASA / chamber of
// agriculture rates, young-farmer years, TSE beneficiaries and SIRET
// numbers).
//
// Fields are exposed as accessors over the raw storage, so group views
// (AIDFIC, DEPDIR, ADTSEG) and their elementary fields share bytes exactly
// like COBOL storage; bytes never touched by a setter (including both
// FILLER areas) round-trip unchanged.
//
// MIGRATION_NOTE: the copybook uses an 'X' placeholder prefix substituted
// by COPY ... REPLACING. In Go each inclusion is simply a distinct DIR
// value. 'X'-FILLER becomes a real data name after substitution, so both
// filler areas are exposed (Filler1/Filler2) to preserve their content.
//
// MIGRATION_NOTE: unsigned zoned-decimal (USAGE DISPLAY) numerics with an
// implied decimal point are returned as scaled integers (see DIRRateScale
// and DIRBapsaScale) together with the result of the COBOL NUMERIC class
// test; the arithmetic primitives decide what to do with invalid data.
type DIR [DIRLen]byte

// NewDIR returns a segment initialised to spaces, as a COBOL MOVE SPACES on
// the alphanumeric group would.
func NewDIR() DIR {
	var d DIR
	for i := range d {
		d[i] = ' '
	}
	return d
}

// ParseDIR copies a 400-byte buffer into a new DIR segment. It fails if b
// is not exactly DIRLen bytes long.
func ParseDIR(b []byte) (DIR, error) {
	var d DIR
	if len(b) != DIRLen {
		return d, fmt.Errorf("zes: DIR segment must be %d bytes, got %d", DIRLen, len(b))
	}
	copy(d[:], b)
	return d, nil
}

// Bytes returns a copy of the raw 400 bytes of the segment.
func (d *DIR) Bytes() []byte {
	out := make([]byte, DIRLen)
	copy(out, d[:])
	return out
}

// ---- low-level helpers ----

func (d *DIR) get(off, n int) string { return string(d[off : off+n]) }

// put performs a COBOL alphanumeric MOVE into a field: left-justified,
// truncated on the right, padded with spaces.
func (d *DIR) put(off, n int, v string) {
	for i := 0; i < n; i++ {
		if i < len(v) {
			d[off+i] = v[i]
		} else {
			d[off+i] = ' '
		}
	}
}

// isNumeric applies the COBOL NUMERIC class test for an unsigned
// zoned-decimal field (every byte a digit).
func (d *DIR) isNumeric(off, n int) bool {
	for _, c := range d[off : off+n] {
		if c < '0' || c > '9' {
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
func (d *DIR) unsigned(off, n int) (value uint64, ok bool) {
	ok = d.isNumeric(off, n)
	for _, c := range d[off : off+n] {
		dg := uint64(c & 0x0F)
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
func (d *DIR) setUnsigned(off, n int, v uint64) {
	for i := n - 1; i >= 0; i-- {
		d[off+i] = byte('0' + v%10)
		v /= 10
	}
}

func dirCheckIndex(field string, i, count int) error {
	if i < 0 || i >= count {
		return fmt.Errorf("zes: DIR %s subscript %d out of range [0,%d)", field, i, count)
	}
	return nil
}

// ---- AIDFIC: file identifier (positions 1-13) ----

// Aidfic returns the 13-byte AIDFIC group (MAJIC2 file identifier).
func (d *DIR) Aidfic() string { return d.get(dirOffAidfic, DIRAidficLen) }

// SetAidfic moves v into the 13-byte AIDFIC group.
func (d *DIR) SetAidfic(v string) { d.put(dirOffAidfic, DIRAidficLen, v) }

// JanRaw returns the raw 4 bytes of JAN PIC 9(4) (year).
func (d *DIR) JanRaw() string { return d.get(dirOffJan, dirYearLen) }

// JanIsNumeric reports whether JAN passes the COBOL NUMERIC class test.
func (d *DIR) JanIsNumeric() bool { return d.isNumeric(dirOffJan, dirYearLen) }

// Jan returns JAN PIC 9(4) (year); ok is false when the field is not NUMERIC.
func (d *DIR) Jan() (uint64, bool) { return d.unsigned(dirOffJan, dirYearLen) }

// SetJan performs a COBOL numeric MOVE into JAN PIC 9(4).
func (d *DIR) SetJan(v uint64) { d.setUnsigned(dirOffJan, dirYearLen, v) }

// Depdir returns DEPDIR X(3): département + direction group.
func (d *DIR) Depdir() string { return d.get(dirOffDepdir, 3) }

// SetDepdir moves v into DEPDIR X(3).
func (d *DIR) SetDepdir(v string) { d.put(dirOffDepdir, 3, v) }

// Codep returns CODEP X(2): département code.
func (d *DIR) Codep() string { return d.get(dirOffCodep, 2) }

// SetCodep moves v into CODEP X(2).
func (d *DIR) SetCodep(v string) { d.put(dirOffCodep, 2, v) }

// Codir returns CODIR X(1): direction code.
func (d *DIR) Codir() string { return d.get(dirOffCodir, 1) }

// SetCodir moves v into CODIR X(1).
func (d *DIR) SetCodir(v string) { d.put(dirOffCodir, 1, v) }

// Ccocom returns CCOCOM X(3): MAJIC2 commune code.
func (d *DIR) Ccocom() string { return d.get(dirOffCcocom, 3) }

// SetCcocom moves v into CCOCOM X(3).
func (d *DIR) SetCcocom(v string) { d.put(dirOffCcocom, 3, v) }

// Ccoifp returns CCOIFP X(3): absorbed-commune code.
func (d *DIR) Ccoifp() string { return d.get(dirOffCcoifp, 3) }

// SetCcoifp moves v into CCOIFP X(3).
func (d *DIR) SetCcoifp(v string) { d.put(dirOffCcoifp, 3, v) }

// ---- ADTSEG: segment body (positions 14-400) ----

// Adtseg returns the 387-byte ADTSEG group (segment body).
func (d *DIR) Adtseg() string { return d.get(dirOffAdtseg, DIRAdtsegLen) }

// SetAdtseg moves v into the 387-byte ADTSEG group.
func (d *DIR) SetAdtseg(v string) { d.put(dirOffAdtseg, DIRAdtsegLen, v) }

// Ccoenr returns CCOENR X(2): record code.
func (d *DIR) Ccoenr() string { return d.get(dirOffCcoenr, 2) }

// SetCcoenr moves v into CCOENR X(2).
func (d *DIR) SetCcoenr(v string) { d.put(dirOffCcoenr, 2, v) }

// Xdday6 returns XDDAY6 X(6): date of the day on 6 characters.
func (d *DIR) Xdday6() string { return d.get(dirOffXdday6, 6) }

// SetXdday6 moves v into XDDAY6 X(6).
func (d *DIR) SetXdday6(v string) { d.put(dirOffXdday6, 6, v) }

// ---- rates PIC 9(4)V9(6), scaled by DIRRateScale ----

// TaudepB returns TAUDEP-B, the département built rate, scaled by
// DIRRateScale; ok is false when the field is not NUMERIC.
func (d *DIR) TaudepB() (uint64, bool) { return d.unsigned(dirOffTaudepB, dirRateLen) }

// SetTaudepB moves a DIRRateScale-scaled value into TAUDEP-B.
func (d *DIR) SetTaudepB(v uint64) { d.setUnsigned(dirOffTaudepB, dirRateLen, v) }

// TauregB returns TAUREG-B, the région built rate, scaled by DIRRateScale.
func (d *DIR) TauregB() (uint64, bool) { return d.unsigned(dirOffTauregB, dirRateLen) }

// SetTauregB moves a DIRRateScale-scaled value into TAUREG-B.
func (d *DIR) SetTauregB(v uint64) { d.setUnsigned(dirOffTauregB, dirRateLen, v) }

// Ptbtas returns PTBTAS, the built TASA rate, scaled by DIRRateScale.
func (d *DIR) Ptbtas() (uint64, bool) { return d.unsigned(dirOffPtbtas, dirRateLen) }

// SetPtbtas moves a DIRRateScale-scaled value into PTBTAS.
func (d *DIR) SetPtbtas(v uint64) { d.setUnsigned(dirOffPtbtas, dirRateLen, v) }

// TaudepNB returns TAUDEP-NB, the département non-built rate, scaled by
// DIRRateScale.
func (d *DIR) TaudepNB() (uint64, bool) { return d.unsigned(dirOffTaudepNB, dirRateLen) }

// SetTaudepNB moves a DIRRateScale-scaled value into TAUDEP-NB.
func (d *DIR) SetTaudepNB(v uint64) { d.setUnsigned(dirOffTaudepNB, dirRateLen, v) }

// TauregNB returns TAUREG-NB, the région non-built rate (also carries the
// Île-de-France TSE rate), scaled by DIRRateScale.
func (d *DIR) TauregNB() (uint64, bool) { return d.unsigned(dirOffTauregNB, dirRateLen) }

// SetTauregNB moves a DIRRateScale-scaled value into TAUREG-NB.
func (d *DIR) SetTauregNB(v uint64) { d.setUnsigned(dirOffTauregNB, dirRateLen, v) }

// TauchagrNB returns TAUCHAGR-NB, the chamber of agriculture non-built
// rate, scaled by DIRRateScale.
func (d *DIR) TauchagrNB() (uint64, bool) { return d.unsigned(dirOffTauchagrNB, dirRateLen) }

// SetTauchagrNB moves a DIRRateScale-scaled value into TAUCHAGR-NB.
func (d *DIR) SetTauchagrNB(v uint64) { d.setUnsigned(dirOffTauchagrNB, dirRateLen, v) }

// Taubap405NB returns TAUBAP405-NB PIC 9(2)V9(2), the non-agricultural
// BAPSA rate (4,05), scaled by DIRBapsaScale.
func (d *DIR) Taubap405NB() (uint64, bool) {
	return d.unsigned(dirOffTaubap405NB, dirBapsaLen)
}

// SetTaubap405NB moves a DIRBapsaScale-scaled value into TAUBAP405-NB.
func (d *DIR) SetTaubap405NB(v uint64) { d.setUnsigned(dirOffTaubap405NB, dirBapsaLen, v) }

// TausmagrNB returns TAUSMAGR-NB, the communal chamber of agriculture
// rate, scaled by DIRRateScale.
func (d *DIR) TausmagrNB() (uint64, bool) { return d.unsigned(dirOffTausmagrNB, dirRateLen) }

// SetTausmagrNB moves a DIRRateScale-scaled value into TAUSMAGR-NB.
func (d *DIR) SetTausmagrNB(v uint64) { d.setUnsigned(dirOffTausmagrNB, dirRateLen, v) }

// Ptbdrb returns PTBDRB, the rebased département built rate, scaled by
// DIRRateScale.
func (d *DIR) Ptbdrb() (uint64, bool) { return d.unsigned(dirOffPtbdrb, dirRateLen) }

// SetPtbdrb moves a DIRRateScale-scaled value into PTBDRB.
func (d *DIR) SetPtbdrb(v uint64) { d.setUnsigned(dirOffPtbdrb, dirRateLen, v) }

// ---- JDEJA OCCURS 5: young-farmer (JA) years ----

// Jdeja returns the raw 8-byte JDEJA(i+1) group (JDEINS + JDERIM) for the
// 0-based index i.
func (d *DIR) Jdeja(i int) (string, error) {
	if err := dirCheckIndex("JDEJA", i, DIRJaCount); err != nil {
		return "", err
	}
	return d.get(dirOffJdeja+i*dirJdejaLen, dirJdejaLen), nil
}

// SetJdeja moves v into the 8-byte JDEJA(i+1) group (0-based i).
func (d *DIR) SetJdeja(i int, v string) error {
	if err := dirCheckIndex("JDEJA", i, DIRJaCount); err != nil {
		return err
	}
	d.put(dirOffJdeja+i*dirJdejaLen, dirJdejaLen, v)
	return nil
}

// Jdeins returns JDEINS(i+1) PIC 9(4) (0-based i), the département JA
// installation year; ok is false when the field is not NUMERIC.
func (d *DIR) Jdeins(i int) (uint64, bool, error) {
	if err := dirCheckIndex("JDEINS", i, DIRJaCount); err != nil {
		return 0, false, err
	}
	v, ok := d.unsigned(dirOffJdeja+i*dirJdejaLen, dirYearLen)
	return v, ok, nil
}

// SetJdeins moves v into JDEINS(i+1) PIC 9(4) (0-based i).
func (d *DIR) SetJdeins(i int, v uint64) error {
	if err := dirCheckIndex("JDEINS", i, DIRJaCount); err != nil {
		return err
	}
	d.setUnsigned(dirOffJdeja+i*dirJdejaLen, dirYearLen, v)
	return nil
}

// Jderim returns JDERIM(i+1) PIC 9(4) (0-based i), the département year of
// return to taxation; ok is false when the field is not NUMERIC.
func (d *DIR) Jderim(i int) (uint64, bool, error) {
	if err := dirCheckIndex("JDERIM", i, DIRJaCount); err != nil {
		return 0, false, err
	}
	v, ok := d.unsigned(dirOffJdeja+i*dirJdejaLen+dirYearLen, dirYearLen)
	return v, ok, nil
}

// SetJderim moves v into JDERIM(i+1) PIC 9(4) (0-based i).
func (d *DIR) SetJderim(i int, v uint64) error {
	if err := dirCheckIndex("JDERIM", i, DIRJaCount); err != nil {
		return err
	}
	d.setUnsigned(dirOffJdeja+i*dirJdejaLen+dirYearLen, dirYearLen, v)
	return nil
}

// ---- fillers ----

// Filler1 returns the 9-byte named 'X'-FILLER area following JDEJA.
func (d *DIR) Filler1() string { return d.get(dirOffFiller1, DIRFiller1Len) }

// SetFiller1 moves v into the 9-byte 'X'-FILLER area following JDEJA.
func (d *DIR) SetFiller1(v string) { d.put(dirOffFiller1, DIRFiller1Len, v) }

// Filler2 returns the 179-byte named 'X'-FILLER area preceding DNUURC.
func (d *DIR) Filler2() string { return d.get(dirOffFiller2, DIRFiller2Len) }

// SetFiller2 moves v into the 179-byte 'X'-FILLER area preceding DNUURC.
func (d *DIR) SetFiller2(v string) { d.put(dirOffFiller2, DIRFiller2Len, v) }

// ---- TSE beneficiaries ----

// Acoets returns ACOETS(i+1) X(4) (0-based i), the TSE beneficiary body
// code.
func (d *DIR) Acoets(i int) (string, error) {
	if err := dirCheckIndex("ACOETS", i, DIRTseCount); err != nil {
		return "", err
	}
	return d.get(dirOffAcoets+i*dirAcoetsLen, dirAcoetsLen), nil
}

// SetAcoets moves v into ACOETS(i+1) X(4) (0-based i).
func (d *DIR) SetAcoets(i int, v string) error {
	if err := dirCheckIndex("ACOETS", i, DIRTseCount); err != nil {
		return err
	}
	d.put(dirOffAcoets+i*dirAcoetsLen, dirAcoetsLen, v)
	return nil
}

// Nbtse returns NBTSE PIC 9(2), the number of TSE on the DSF; ok is false
// when the field is not NUMERIC.
//
// MIGRATION_NOTE: NBTSE is informational only — ACOETS is a fixed OCCURS 5
// table, not OCCURS DEPENDING ON, so NBTSE never changes the layout.
func (d *DIR) Nbtse() (uint64, bool) { return d.unsigned(dirOffNbtse, dirNbtseLen) }

// SetNbtse moves v into NBTSE PIC 9(2).
func (d *DIR) SetNbtse(v uint64) { d.setUnsigned(dirOffNbtse, dirNbtseLen, v) }

// ---- SIRET numbers and options ----

// Dsitas returns DSITAS X(14), the SIRET of the TASA beneficiary region.
func (d *DIR) Dsitas() string { return d.get(dirOffDsitas, dirSiretLen) }

// SetDsitas moves v into DSITAS X(14).
func (d *DIR) SetDsitas(v string) { d.put(dirOffDsitas, dirSiretLen, v) }

// Grotas returns GROTAS X(1), the TASA role-copy option.
func (d *DIR) Grotas() string { return d.get(dirOffGrotas, 1) }

// SetGrotas moves v into GROTAS X(1).
func (d *DIR) SetGrotas(v string) { d.put(dirOffGrotas, 1, v) }

// Dsired returns DSIRED X(14), the SIRET of the direction.
func (d *DIR) Dsired() string { return d.get(dirOffDsired, dirSiretLen) }

// SetDsired moves v into DSIRED X(14).
func (d *DIR) SetDsired(v string) { d.put(dirOffDsired, dirSiretLen, v) }

// Dsica returns DSICA X(14), the SIRET of the chamber of agriculture.
func (d *DIR) Dsica() string { return d.get(dirOffDsica, dirSiretLen) }

// SetDsica moves v into DSICA X(14).
func (d *DIR) SetDsica(v string) { d.put(dirOffDsica, dirSiretLen, v) }

// Dnuurc returns DNUURC X(2): CDIF regrouping unit number (UR).
func (d *DIR) Dnuurc() string { return d.get(dirOffDnuurc, 2) }

// SetDnuurc moves v into DNUURC X(2).
func (d *DIR) SetDnuurc(v string) { d.put(dirOffDnuurc, 2, v) }
