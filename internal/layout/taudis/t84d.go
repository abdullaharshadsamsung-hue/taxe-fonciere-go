package taudis

import (
	"fmt"
)

// DirectionLetter is the record-type discriminator (DLETAR) of a TAUDIS
// Direction article.
const DirectionLetter = "D"

// RateScale is the scaling factor of the PIC 9(4)V9(6) rate fields: a rate
// accessor returns the value multiplied by 10^6 (six implied decimals).
const RateScale = 1_000_000

// Number of OCCURS entries in the Direction layout.
const (
	// TSECount is the OCCURS count of the ADTTSE table.
	TSECount = 5
	// CounterCount is the OCCURS count of the NBHCNT and NABOND tables.
	CounterCount = 16
)

// Byte offsets of the DONNEES-DIRECTION fields, relative to the start of the
// 2988-byte body (i.e. the record offset minus EnteteLen). Every data field
// is preceded by a 1-byte 'X'-FILLER separator, except VMAILD, VMDTAS, the
// ANBOND group (which directly follows ANACNT) and the subordinate items of
// the JDAHO1 / DHMRx groups.
const (
	dOffDletar     = 1    // X(1)
	dOffDlidir     = 3    // X(30)
	dOffJdacre     = 34   // X(8)
	dOffDsired     = 43   // X(14)
	dOffNbcom1     = 58   // 9(3)
	dOffJanrol     = 62   // 9(4)
	dOffJanmir     = 67   // X(2)
	dOffJbarol     = 70   // X(10)
	dOffJdaho1     = 81   // group (8) = JSSHOM 9(2) JANHO1 9(2) JMOHOM 9(2) JJRHOM 9(2)
	dOffJsshom     = 81   // 9(2)
	dOffJanho1     = 83   // 9(2)
	dOffJmohom     = 85   // 9(2)
	dOffJjrhom     = 87   // 9(2)
	dOffJdami1     = 90   // 9(8)
	dOffJdamjo     = 99   // 9(8)
	dOffDdrltf     = 108  // 9(1)
	dOffDdrrol     = 110  // 9(3)
	dOffCtyrol     = 114  // X(2)
	dOffDhmrg      = 117  // group (4) = DDRHMG 9 + DDRROG 9(3)
	dOffDdrhmg     = 117  // 9(1)
	dOffDdrrog     = 118  // 9(3)
	dOffDhmrd      = 122  // group (4) = DDRHMD 9 + DDRROD 9(3)
	dOffDdrhmd     = 122  // 9(1)
	dOffDdrrod     = 123  // 9(3)
	dOffDhmrr      = 127  // group (4) = DDRHMR 9 + DDRROR 9(3)
	dOffDdrhmr     = 127  // 9(1)
	dOffDdrror     = 128  // 9(3)
	dOffGcovef     = 132  // X
	dOffGorvef     = 134  // X
	dOffGcofdl     = 136  // X
	dOffCcofdl     = 138  // X
	dOffGtacom     = 140  // X
	dOffGtaco2     = 142  // X
	dOffGtoctx     = 144  // X
	dOffGtorop     = 146  // X
	dOffGpctfx     = 148  // X
	dOffGtodgs     = 150  // X
	dOffGrodep     = 152  // X
	dOffGrotas     = 154  // X
	dOffPtbdep     = 156  // 9(4)V9(6)
	dOffPtbtas     = 167  // 9(4)V9(6)
	dOffPnbagr     = 178  // 9(4)V9(6)
	dOffPnbagn     = 189  // 9(4)V9(6)
	dOffSiretChagr = 200  // X(14)
	dOffTpnagr     = 215  // 9(10)
	dOffMpnagr     = 226  // 9(10)
	dOffAdttse     = 236  // OCCURS 5, 60 bytes each
	dOffNbtse      = 537  // 9(2)
	dOffDsitas     = 540  // X(14)
	dOffTpbtas     = 555  // 9(10)
	dOffMpbtas     = 566  // 9(10)
	dOffMpbtax     = 577  // 9(10)
	dOffVboinf     = 588  // 9(7)
	dOffVbosup     = 596  // 9(7)
	dOffDdravi     = 604  // 9(7)
	dOffNarsae     = 612  // 9(6)
	dOffNarsiu     = 619  // 9(6)
	dOffAnacnt     = 626  // NBHCNT 9(7) OCCURS 16
	dOffAnbond     = 738  // NABOND 9(7) OCCURS 16 (no separator)
	dOffVmailg     = 851  // X(50)
	dOffVmaild     = 901  // X(50) (no separator)
	dOffVmgtas     = 952  // X(50)
	dOffVmdtas     = 1002 // X(50) (no separator)
	dOffLibre      = 1052 // FILLER X(1936)
	dLibreLen      = 1936

	rateLen    = 10
	amountLen  = 10
	counterLen = 7
	emailLen   = 50
	siretLen   = 14
)

// Layout of one ADTTSE entry (60 bytes), offsets relative to the entry.
const (
	tseEntryLen  = 60
	tseOffAcoets = 1  // X(4)
	tseOffTpbtse = 6  // 9(10)
	tseOffMpbtse = 17 // 9(10)
	tseOffMpbtsx = 28 // 9(10)
	tseOffTpntse = 39 // 9(10)
	tseOffMpntse = 50 // 9(10)
)

// Compile-time guard: the layout must end exactly at BodyLen (2988 bytes).
var _ = [1]struct{}{}[dOffLibre+dLibreLen-BodyLen]

// Direction is the DONNEES-DIRECTION body (copybook T84D) of a TAUDIS 'D'
// record: the 2988 bytes that follow the 12-byte Entete. Fields are exposed
// as accessor methods over the raw storage so that encoding and decoding are
// byte-exact and FILLER bytes (separators and the trailing free zone) survive
// a round trip untouched.
//
// MIGRATION_NOTE: the copybook's 'X'- prefix is a COPY REPLACING placeholder
// (EFITAUX2 includes it as T-D-); in Go each inclusion is simply a distinct
// Direction value.
type Direction [BodyLen]byte

// NewDirection returns a Direction body whose alphanumeric fields and
// FILLERs are spaces, numeric fields are zeros, and whose article letter
// DLETAR is 'D'.
func NewDirection() Direction {
	var d Direction
	for i := range d {
		d[i] = ' '
	}
	for _, f := range directionNumericFields() {
		d.setNum(f[0], f[1], 0)
	}
	d.SetDletar(DirectionLetter)
	return d
}

// directionNumericFields lists the (offset, length) of every unsigned
// zoned-decimal field of the layout.
func directionNumericFields() [][2]int {
	fields := [][2]int{
		{dOffNbcom1, 3}, {dOffJanrol, 4}, {dOffJdaho1, 8}, {dOffJdami1, 8},
		{dOffJdamjo, 8}, {dOffDdrltf, 1}, {dOffDdrrol, 3}, {dOffDhmrg, 4},
		{dOffDhmrd, 4}, {dOffDhmrr, 4}, {dOffPtbdep, rateLen},
		{dOffPtbtas, rateLen}, {dOffPnbagr, rateLen}, {dOffPnbagn, rateLen},
		{dOffTpnagr, amountLen}, {dOffMpnagr, amountLen}, {dOffNbtse, 2},
		{dOffTpbtas, amountLen}, {dOffMpbtas, amountLen}, {dOffMpbtax, amountLen},
		{dOffVboinf, 7}, {dOffVbosup, 7}, {dOffDdravi, 7}, {dOffNarsae, 6},
		{dOffNarsiu, 6},
		{dOffAnacnt, CounterCount * counterLen}, {dOffAnbond, CounterCount * counterLen},
	}
	for i := 0; i < TSECount; i++ {
		base := dOffAdttse + i*tseEntryLen
		for _, o := range []int{tseOffTpbtse, tseOffMpbtse, tseOffMpbtsx, tseOffTpntse, tseOffMpntse} {
			fields = append(fields, [2]int{base + o, amountLen})
		}
	}
	return fields
}

// ParseDirection copies a 2988-byte body into a Direction. It fails when b
// does not have exactly BodyLen bytes.
func ParseDirection(b []byte) (Direction, error) {
	var d Direction
	if len(b) != BodyLen {
		return d, fmt.Errorf("taudis: direction body must be %d bytes, got %d", BodyLen, len(b))
	}
	copy(d[:], b)
	return d, nil
}

// Bytes returns a copy of the raw 2988 bytes (encoded layout).
func (d *Direction) Bytes() []byte {
	out := make([]byte, BodyLen)
	copy(out, d[:])
	return out
}

// IsDirection reports whether DLETAR holds the 'D' article letter.
func (d *Direction) IsDirection() bool { return d.Dletar() == DirectionLetter }

// ---- low-level helpers ----

func (d *Direction) get(off, n int) string { return string(d[off : off+n]) }

// put performs a COBOL alphanumeric MOVE: left-justified, truncated on the
// right, padded with spaces.
func (d *Direction) put(off, n int, v string) {
	for i := 0; i < n; i++ {
		if i < len(v) {
			d[off+i] = v[i]
		} else {
			d[off+i] = ' '
		}
	}
}

// isNum is the COBOL NUMERIC class test for an unsigned DISPLAY field.
func (d *Direction) isNum(off, n int) bool {
	for _, c := range d[off : off+n] {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// num decodes an unsigned zoned-decimal field. ok is false when the field
// fails the NUMERIC test; the value is then built from the low nibble of
// each byte (as IBM COBOL arithmetic on invalid zoned data would) and must
// be treated as unreliable.
func (d *Direction) num(off, n int) (uint64, bool) {
	ok := d.isNum(off, n)
	var v uint64
	for _, c := range d[off : off+n] {
		dig := uint64(c & 0x0F)
		if dig > 9 {
			dig = 9 // MIGRATION_NOTE: invalid nibble clamped to keep the value in range
		}
		v = v*10 + dig
	}
	return v, ok
}

// setNum performs a COBOL numeric MOVE into an unsigned DISPLAY field:
// high-order digits beyond n are truncated, zero-filled on the left.
func (d *Direction) setNum(off, n int, v uint64) {
	for i := off + n - 1; i >= off; i-- {
		d[i] = byte('0' + v%10)
		v /= 10
	}
}

func checkIndex(name string, i, max int) error {
	if i < 0 || i >= max {
		return fmt.Errorf("taudis: %s index %d out of range [0,%d)", name, i, max)
	}
	return nil
}

// ---- identification ----

// Dletar returns DLETAR X(1), the article letter ('D' for Direction).
func (d *Direction) Dletar() string { return d.get(dOffDletar, 1) }

// SetDletar moves v into DLETAR X(1).
func (d *Direction) SetDletar(v string) { d.put(dOffDletar, 1, v) }

// Dlidir returns DLIDIR X(30), the direction label.
func (d *Direction) Dlidir() string { return d.get(dOffDlidir, 30) }

// SetDlidir moves v into DLIDIR X(30).
func (d *Direction) SetDlidir(v string) { d.put(dOffDlidir, 30, v) }

// Jdacre returns JDACRE X(8), the TAUDIS creation date.
func (d *Direction) Jdacre() string { return d.get(dOffJdacre, 8) }

// SetJdacre moves v into JDACRE X(8).
func (d *Direction) SetJdacre(v string) { d.put(dOffJdacre, 8, v) }

// Dsired returns DSIRED X(14), the département SIRET.
func (d *Direction) Dsired() string { return d.get(dOffDsired, siretLen) }

// SetDsired moves v into DSIRED X(14).
func (d *Direction) SetDsired(v string) { d.put(dOffDsired, siretLen, v) }

// Nbcom1 returns NBCOM1 9(3), the number of communes of the direction.
func (d *Direction) Nbcom1() (uint64, bool) { return d.num(dOffNbcom1, 3) }

// SetNbcom1 moves v into NBCOM1 9(3).
func (d *Direction) SetNbcom1(v uint64) { d.setNum(dOffNbcom1, 3, v) }

// Janrol returns JANROL 9(4), the tax year (AAAA).
func (d *Direction) Janrol() (uint64, bool) { return d.num(dOffJanrol, 4) }

// SetJanrol moves v into JANROL 9(4).
func (d *Direction) SetJanrol(v uint64) { d.setNum(dOffJanrol, 4, v) }

// Janmir returns JANMIR X(2), the year of issue (AA).
func (d *Direction) Janmir() string { return d.get(dOffJanmir, 2) }

// SetJanmir moves v into JANMIR X(2).
func (d *Direction) SetJanmir(v string) { d.put(dOffJanmir, 2, v) }

// Jbarol returns JBAROL X(10), the BASEROLE creation/modification date.
func (d *Direction) Jbarol() string { return d.get(dOffJbarol, 10) }

// SetJbarol moves v into JBAROL X(10).
func (d *Direction) SetJbarol(v string) { d.put(dOffJbarol, 10, v) }

// ---- homologation date JDAHO1 ----

// Jdaho1 returns the raw 8-byte JDAHO1 group (SSAAMMJJ).
func (d *Direction) Jdaho1() string { return d.get(dOffJdaho1, 8) }

// SetJdaho1 moves v into the JDAHO1 group (alphanumeric group MOVE).
func (d *Direction) SetJdaho1(v string) { d.put(dOffJdaho1, 8, v) }

// Jsshom returns JSSHOM 9(2), the homologation century.
func (d *Direction) Jsshom() (uint64, bool) { return d.num(dOffJsshom, 2) }

// SetJsshom moves v into JSSHOM 9(2).
func (d *Direction) SetJsshom(v uint64) { d.setNum(dOffJsshom, 2, v) }

// Janho1 returns JANHO1 9(2), the homologation year within the century.
func (d *Direction) Janho1() (uint64, bool) { return d.num(dOffJanho1, 2) }

// SetJanho1 moves v into JANHO1 9(2).
func (d *Direction) SetJanho1(v uint64) { d.setNum(dOffJanho1, 2, v) }

// Jmohom returns JMOHOM 9(2), the homologation month.
func (d *Direction) Jmohom() (uint64, bool) { return d.num(dOffJmohom, 2) }

// SetJmohom moves v into JMOHOM 9(2).
func (d *Direction) SetJmohom(v uint64) { d.setNum(dOffJmohom, 2, v) }

// Jjrhom returns JJRHOM 9(2), the homologation day.
func (d *Direction) Jjrhom() (uint64, bool) { return d.num(dOffJjrhom, 2) }

// SetJjrhom moves v into JJRHOM 9(2).
func (d *Direction) SetJjrhom(v uint64) { d.setNum(dOffJjrhom, 2, v) }

// ---- last roles ----

// Jdami1 returns JDAMI1 9(8), the issue date of the last role.
func (d *Direction) Jdami1() (uint64, bool) { return d.num(dOffJdami1, 8) }

// SetJdami1 moves v into JDAMI1 9(8).
func (d *Direction) SetJdami1(v uint64) { d.setNum(dOffJdami1, 8, v) }

// Jdamjo returns JDAMJO 9(8), the surcharge date of the last role.
func (d *Direction) Jdamjo() (uint64, bool) { return d.num(dOffJdamjo, 8) }

// SetJdamjo moves v into JDAMJO 9(8).
func (d *Direction) SetJdamjo(v uint64) { d.setNum(dOffJdamjo, 8, v) }

// Ddrltf returns DDRLTF 9(1), the last processed lot number.
func (d *Direction) Ddrltf() (uint64, bool) { return d.num(dOffDdrltf, 1) }

// SetDdrltf moves v into DDRLTF 9(1).
func (d *Direction) SetDdrltf(v uint64) { d.setNum(dOffDdrltf, 1, v) }

// Ddrrol returns DDRROL 9(3), the last role number.
func (d *Direction) Ddrrol() (uint64, bool) { return d.num(dOffDdrrol, 3) }

// SetDdrrol moves v into DDRROL 9(3).
func (d *Direction) SetDdrrol(v uint64) { d.setNum(dOffDdrrol, 3, v) }

// Ctyrol returns CTYROL X(2), the role type (RG, RD, RR).
func (d *Direction) Ctyrol() string { return d.get(dOffCtyrol, 2) }

// SetCtyrol moves v into CTYROL X(2).
func (d *Direction) SetCtyrol(v string) { d.put(dOffCtyrol, 2, v) }

// Dhmrg returns the raw 4-byte DHMRG group (last lot + initial general role).
func (d *Direction) Dhmrg() string { return d.get(dOffDhmrg, 4) }

// SetDhmrg moves v into the DHMRG group.
func (d *Direction) SetDhmrg(v string) { d.put(dOffDhmrg, 4, v) }

// Ddrhmg returns DDRHMG 9(1).
func (d *Direction) Ddrhmg() (uint64, bool) { return d.num(dOffDdrhmg, 1) }

// SetDdrhmg moves v into DDRHMG 9(1).
func (d *Direction) SetDdrhmg(v uint64) { d.setNum(dOffDdrhmg, 1, v) }

// Ddrrog returns DDRROG 9(3).
func (d *Direction) Ddrrog() (uint64, bool) { return d.num(dOffDdrrog, 3) }

// SetDdrrog moves v into DDRROG 9(3).
func (d *Direction) SetDdrrog(v uint64) { d.setNum(dOffDdrrog, 3, v) }

// Dhmrd returns the raw 4-byte DHMRD group (last lot + deferred general role).
func (d *Direction) Dhmrd() string { return d.get(dOffDhmrd, 4) }

// SetDhmrd moves v into the DHMRD group.
func (d *Direction) SetDhmrd(v string) { d.put(dOffDhmrd, 4, v) }

// Ddrhmd returns DDRHMD 9(1).
func (d *Direction) Ddrhmd() (uint64, bool) { return d.num(dOffDdrhmd, 1) }

// SetDdrhmd moves v into DDRHMD 9(1).
func (d *Direction) SetDdrhmd(v uint64) { d.setNum(dOffDdrhmd, 1, v) }

// Ddrrod returns DDRROD 9(3).
func (d *Direction) Ddrrod() (uint64, bool) { return d.num(dOffDdrrod, 3) }

// SetDdrrod moves v into DDRROD 9(3).
func (d *Direction) SetDdrrod(v uint64) { d.setNum(dOffDdrrod, 3, v) }

// Dhmrr returns the raw 4-byte DHMRR group (last lot + rebuilt general role).
func (d *Direction) Dhmrr() string { return d.get(dOffDhmrr, 4) }

// SetDhmrr moves v into the DHMRR group.
func (d *Direction) SetDhmrr(v string) { d.put(dOffDhmrr, 4, v) }

// Ddrhmr returns DDRHMR 9(1).
func (d *Direction) Ddrhmr() (uint64, bool) { return d.num(dOffDdrhmr, 1) }

// SetDdrhmr moves v into DDRHMR 9(1).
func (d *Direction) SetDdrhmr(v uint64) { d.setNum(dOffDdrhmr, 1, v) }

// Ddrror returns DDRROR 9(3).
func (d *Direction) Ddrror() (uint64, bool) { return d.num(dOffDdrror, 3) }

// SetDdrror moves v into DDRROR 9(3).
func (d *Direction) SetDdrror(v uint64) { d.setNum(dOffDdrror, 3, v) }

// ---- one-byte flags ----

// Gcovef returns GCOVEF X: EFICT validation code ('V' = valid).
func (d *Direction) Gcovef() string { return d.get(dOffGcovef, 1) }

// SetGcovef moves v into GCOVEF X.
func (d *Direction) SetGcovef(v string) { d.put(dOffGcovef, 1, v) }

// Gorvef returns GORVEF X: EFICT validation origin ('C' = CLIST, blank = batch).
func (d *Direction) Gorvef() string { return d.get(dOffGorvef, 1) }

// SetGorvef moves v into GORVEF X.
func (d *Direction) SetGorvef(v string) { d.put(dOffGorvef, 1, v) }

// Gcofdl returns GCOFDL X: FDL validation code ('1' = valid).
func (d *Direction) Gcofdl() string { return d.get(dOffGcofdl, 1) }

// SetGcofdl moves v into GCOFDL X.
func (d *Direction) SetGcofdl(v string) { d.put(dOffGcofdl, 1, v) }

// Ccofdl returns CCOFDL X: FDL pass number.
func (d *Direction) Ccofdl() string { return d.get(dOffCcofdl, 1) }

// SetCcofdl moves v into CCOFDL X.
func (d *Direction) SetCcofdl(v string) { d.put(dOffCcofdl, 1, v) }

// Gtacom returns GTACOM X: all communes taxed ('*' or blank).
func (d *Direction) Gtacom() string { return d.get(dOffGtacom, 1) }

// SetGtacom moves v into GTACOM X.
func (d *Direction) SetGtacom(v string) { d.put(dOffGtacom, 1, v) }

// Gtaco2 returns GTACO2 X: all Paris communes taxed ('*' or blank).
func (d *Direction) Gtaco2() string { return d.get(dOffGtaco2, 1) }

// SetGtaco2 moves v into GTACO2 X.
func (d *Direction) SetGtaco2(v string) { d.put(dOffGtaco2, 1, v) }

// Gtoctx returns GTOCTX X: 'C' = contentieux, 'T' = taxation.
func (d *Direction) Gtoctx() string { return d.get(dOffGtoctx, 1) }

// SetGtoctx moves v into GTOCTX X.
func (d *Direction) SetGtoctx(v string) { d.put(dOffGtoctx, 1, v) }

// Gtorop returns GTOROP X: primitive role flag ('*' = RMP).
func (d *Direction) Gtorop() string { return d.get(dOffGtorop, 1) }

// SetGtorop moves v into GTOROP X.
func (d *Direction) SetGtorop(v string) { d.put(dOffGtorop, 1, v) }

// Gpctfx returns GPCTFX X: 'G' = DGE, 'P' = CP.
func (d *Direction) Gpctfx() string { return d.get(dOffGpctfx, 1) }

// SetGpctfx moves v into GPCTFX X.
func (d *Direction) SetGpctfx(v string) { d.put(dOffGpctfx, 1, v) }

// Gtodgs returns GTODGS X: DGE-only taxation flag.
func (d *Direction) Gtodgs() string { return d.get(dOffGtodgs, 1) }

// SetGtodgs moves v into GTODGS X.
func (d *Direction) SetGtodgs(v string) { d.put(dOffGtodgs, 1, v) }

// Grodep returns GRODEP X: département role-copy option.
func (d *Direction) Grodep() string { return d.get(dOffGrodep, 1) }

// SetGrodep moves v into GRODEP X.
func (d *Direction) SetGrodep(v string) { d.put(dOffGrodep, 1, v) }

// Grotas returns GROTAS X: TASA-beneficiary region role-copy option.
func (d *Direction) Grotas() string { return d.get(dOffGrotas, 1) }

// SetGrotas moves v into GROTAS X.
func (d *Direction) SetGrotas(v string) { d.put(dOffGrotas, 1, v) }

// ---- rates PIC 9(4)V9(6), returned scaled by RateScale ----

// Ptbdep returns PTBDEP 9(4)V9(6), the département built-property rate,
// scaled by RateScale (e.g. 12.345678 -> 12345678).
func (d *Direction) Ptbdep() (uint64, bool) { return d.num(dOffPtbdep, rateLen) }

// PtbdepRaw returns the raw 10 bytes of PTBDEP.
func (d *Direction) PtbdepRaw() string { return d.get(dOffPtbdep, rateLen) }

// SetPtbdep moves a RateScale-scaled value into PTBDEP.
func (d *Direction) SetPtbdep(v uint64) { d.setNum(dOffPtbdep, rateLen, v) }

// Ptbtas returns PTBTAS 9(4)V9(6), the TASA built-property rate, scaled by
// RateScale.
func (d *Direction) Ptbtas() (uint64, bool) { return d.num(dOffPtbtas, rateLen) }

// PtbtasRaw returns the raw 10 bytes of PTBTAS.
func (d *Direction) PtbtasRaw() string { return d.get(dOffPtbtas, rateLen) }

// SetPtbtas moves a RateScale-scaled value into PTBTAS.
func (d *Direction) SetPtbtas(v uint64) { d.setNum(dOffPtbtas, rateLen, v) }

// Pnbagr returns PNBAGR 9(4)V9(6), the chamber of agriculture non-built
// rate, scaled by RateScale.
func (d *Direction) Pnbagr() (uint64, bool) { return d.num(dOffPnbagr, rateLen) }

// PnbagrRaw returns the raw 10 bytes of PNBAGR.
func (d *Direction) PnbagrRaw() string { return d.get(dOffPnbagr, rateLen) }

// SetPnbagr moves a RateScale-scaled value into PNBAGR.
func (d *Direction) SetPnbagr(v uint64) { d.setNum(dOffPnbagr, rateLen, v) }

// Pnbagn returns PNBAGN 9(4)V9(6), the Saint-Martin (971) consular chamber
// non-built rate, scaled by RateScale.
func (d *Direction) Pnbagn() (uint64, bool) { return d.num(dOffPnbagn, rateLen) }

// PnbagnRaw returns the raw 10 bytes of PNBAGN. EFITAUX2 tests the
// reference modification PNBAGN(1:4) against '9999' on these raw bytes.
func (d *Direction) PnbagnRaw() string { return d.get(dOffPnbagn, rateLen) }

// PnbagnIsSentinel reports whether PNBAGN(1:4) = '9999', the "no rate"
// marker tested by EFITAUX2 before moving PNBAGN.
func (d *Direction) PnbagnIsSentinel() bool { return d.get(dOffPnbagn, 4) == "9999" }

// SetPnbagn moves a RateScale-scaled value into PNBAGN.
func (d *Direction) SetPnbagn(v uint64) { d.setNum(dOffPnbagn, rateLen, v) }

// ---- chamber of agriculture ----

// SiretChagr returns SIRET-CHAGR X(14), the chamber of agriculture SIRET.
func (d *Direction) SiretChagr() string { return d.get(dOffSiretChagr, siretLen) }

// SetSiretChagr moves v into SIRET-CHAGR X(14).
func (d *Direction) SetSiretChagr(v string) { d.put(dOffSiretChagr, siretLen, v) }

// Tpnagr returns TPNAGR 9(10), the total expected chamber of agriculture revenue.
func (d *Direction) Tpnagr() (uint64, bool) { return d.num(dOffTpnagr, amountLen) }

// SetTpnagr moves v into TPNAGR 9(10).
func (d *Direction) SetTpnagr(v uint64) { d.setNum(dOffTpnagr, amountLen, v) }

// Mpnagr returns MPNAGR 9(10), the expected chamber of agriculture revenue
// (Paris directions prorata).
func (d *Direction) Mpnagr() (uint64, bool) { return d.num(dOffMpnagr, amountLen) }

// SetMpnagr moves v into MPNAGR 9(10).
func (d *Direction) SetMpnagr(v uint64) { d.setNum(dOffMpnagr, amountLen, v) }

// ---- ADTTSE OCCURS 5 ----

// TSEEntry is a view over one 60-byte ADTTSE occurrence. It aliases the
// Direction storage it was obtained from, so setters write through.
type TSEEntry struct {
	d    *Direction
	base int
}

// TSE returns a view over the ADTTSE occurrence i (0-based, COBOL subscript
// i+1). It fails when i is outside [0, TSECount).
func (d *Direction) TSE(i int) (TSEEntry, error) {
	if err := checkIndex("ADTTSE", i, TSECount); err != nil {
		return TSEEntry{}, err
	}
	return TSEEntry{d: d, base: dOffAdttse + i*tseEntryLen}, nil
}

// Acoets returns ACOETS X(4), the TSE establishment code.
func (t TSEEntry) Acoets() string { return t.d.get(t.base+tseOffAcoets, 4) }

// SetAcoets moves v into ACOETS X(4).
func (t TSEEntry) SetAcoets(v string) { t.d.put(t.base+tseOffAcoets, 4, v) }

// Tpbtse returns TPBTSE 9(10), the total expected built TSE revenue.
func (t TSEEntry) Tpbtse() (uint64, bool) { return t.d.num(t.base+tseOffTpbtse, amountLen) }

// SetTpbtse moves v into TPBTSE 9(10).
func (t TSEEntry) SetTpbtse(v uint64) { t.d.setNum(t.base+tseOffTpbtse, amountLen, v) }

// Mpbtse returns MPBTSE 9(10), the expected built TSE revenue (Paris prorata).
func (t TSEEntry) Mpbtse() (uint64, bool) { return t.d.num(t.base+tseOffMpbtse, amountLen) }

// SetMpbtse moves v into MPBTSE 9(10).
func (t TSEEntry) SetMpbtse(v uint64) { t.d.setNum(t.base+tseOffMpbtse, amountLen, v) }

// Mpbtsx returns MPBTSX 9(10), the expected built TSE revenue given by FDL.
func (t TSEEntry) Mpbtsx() (uint64, bool) { return t.d.num(t.base+tseOffMpbtsx, amountLen) }

// SetMpbtsx moves v into MPBTSX 9(10).
func (t TSEEntry) SetMpbtsx(v uint64) { t.d.setNum(t.base+tseOffMpbtsx, amountLen, v) }

// Tpntse returns TPNTSE 9(10), the total expected non-built TSE revenue.
func (t TSEEntry) Tpntse() (uint64, bool) { return t.d.num(t.base+tseOffTpntse, amountLen) }

// SetTpntse moves v into TPNTSE 9(10).
func (t TSEEntry) SetTpntse(v uint64) { t.d.setNum(t.base+tseOffTpntse, amountLen, v) }

// Mpntse returns MPNTSE 9(10), the expected non-built TSE revenue (Paris prorata).
func (t TSEEntry) Mpntse() (uint64, bool) { return t.d.num(t.base+tseOffMpntse, amountLen) }

// SetMpntse moves v into MPNTSE 9(10).
func (t TSEEntry) SetMpntse(v uint64) { t.d.setNum(t.base+tseOffMpntse, amountLen, v) }

// ---- TSE count and TASA ----

// Nbtse returns NBTSE 9(2), the number of TSE on the direction.
func (d *Direction) Nbtse() (uint64, bool) { return d.num(dOffNbtse, 2) }

// SetNbtse moves v into NBTSE 9(2).
func (d *Direction) SetNbtse(v uint64) { d.setNum(dOffNbtse, 2, v) }

// Dsitas returns DSITAS X(14), the SIRET of the TASA-beneficiary region.
func (d *Direction) Dsitas() string { return d.get(dOffDsitas, siretLen) }

// SetDsitas moves v into DSITAS X(14).
func (d *Direction) SetDsitas(v string) { d.put(dOffDsitas, siretLen, v) }

// Tpbtas returns TPBTAS 9(10), the total expected TASA revenue.
func (d *Direction) Tpbtas() (uint64, bool) { return d.num(dOffTpbtas, amountLen) }

// SetTpbtas moves v into TPBTAS 9(10).
func (d *Direction) SetTpbtas(v uint64) { d.setNum(dOffTpbtas, amountLen, v) }

// Mpbtas returns MPBTAS 9(10), the expected TASA revenue (Paris prorata).
func (d *Direction) Mpbtas() (uint64, bool) { return d.num(dOffMpbtas, amountLen) }

// SetMpbtas moves v into MPBTAS 9(10).
func (d *Direction) SetMpbtas(v uint64) { d.setNum(dOffMpbtas, amountLen, v) }

// Mpbtax returns MPBTAX 9(10), the expected TASA revenue given by FDL.
func (d *Direction) Mpbtax() (uint64, bool) { return d.num(dOffMpbtax, amountLen) }

// SetMpbtax moves v into MPBTAX 9(10).
func (d *Direction) SetMpbtax(v uint64) { d.setNum(dOffMpbtax, amountLen, v) }

// ---- invoice numbers and article counts ----

// Vboinf returns VBOINF 9(7), the invoice number lower bound.
func (d *Direction) Vboinf() (uint64, bool) { return d.num(dOffVboinf, 7) }

// SetVboinf moves v into VBOINF 9(7).
func (d *Direction) SetVboinf(v uint64) { d.setNum(dOffVboinf, 7, v) }

// Vbosup returns VBOSUP 9(7), the invoice number upper bound.
func (d *Direction) Vbosup() (uint64, bool) { return d.num(dOffVbosup, 7) }

// SetVbosup moves v into VBOSUP 9(7).
func (d *Direction) SetVbosup(v uint64) { d.setNum(dOffVbosup, 7, v) }

// Ddravi returns DDRAVI 9(7), the last invoice number assigned.
func (d *Direction) Ddravi() (uint64, bool) { return d.num(dOffDdravi, 7) }

// SetDdravi moves v into DDRAVI 9(7).
func (d *Direction) SetDdravi(v uint64) { d.setNum(dOffDdravi, 7, v) }

// Narsae returns NARSAE 9(6), the number of role articles without a
// mailing address.
func (d *Direction) Narsae() (uint64, bool) { return d.num(dOffNarsae, 6) }

// SetNarsae moves v into NARSAE 9(6).
func (d *Direction) SetNarsae(v uint64) { d.setNum(dOffNarsae, 6, v) }

// Narsiu returns NARSIU 9(6), the number of role articles without IUI.
func (d *Direction) Narsiu() (uint64, bool) { return d.num(dOffNarsiu, 6) }

// SetNarsiu moves v into NARSIU 9(6).
func (d *Direction) SetNarsiu(v uint64) { d.setNum(dOffNarsiu, 6, v) }

// ---- HCNT / BOND counters ----

// Anacnt returns the raw 112-byte ANACNT group.
func (d *Direction) Anacnt() string { return d.get(dOffAnacnt, CounterCount*counterLen) }

// Nbhcnt returns NBHCNT(i+1) 9(7), the HCNT counter i (0-based). It fails
// when i is outside [0, CounterCount).
func (d *Direction) Nbhcnt(i int) (uint64, bool, error) {
	if err := checkIndex("NBHCNT", i, CounterCount); err != nil {
		return 0, false, err
	}
	v, ok := d.num(dOffAnacnt+i*counterLen, counterLen)
	return v, ok, nil
}

// SetNbhcnt moves v into NBHCNT(i+1) 9(7).
func (d *Direction) SetNbhcnt(i int, v uint64) error {
	if err := checkIndex("NBHCNT", i, CounterCount); err != nil {
		return err
	}
	d.setNum(dOffAnacnt+i*counterLen, counterLen, v)
	return nil
}

// Anbond returns the raw 112-byte ANBOND group.
func (d *Direction) Anbond() string { return d.get(dOffAnbond, CounterCount*counterLen) }

// Nabond returns NABOND(i+1) 9(7), the BOND counter i (0-based). It fails
// when i is outside [0, CounterCount).
func (d *Direction) Nabond(i int) (uint64, bool, error) {
	if err := checkIndex("NABOND", i, CounterCount); err != nil {
		return 0, false, err
	}
	v, ok := d.num(dOffAnbond+i*counterLen, counterLen)
	return v, ok, nil
}

// SetNabond moves v into NABOND(i+1) 9(7).
func (d *Direction) SetNabond(i int, v uint64) error {
	if err := checkIndex("NABOND", i, CounterCount); err != nil {
		return err
	}
	d.setNum(dOffAnbond+i*counterLen, counterLen, v)
	return nil
}

// ---- e-mail addresses ----

// Vmailg returns VMAILG X(50), the left part of the département e-mail.
func (d *Direction) Vmailg() string { return d.get(dOffVmailg, emailLen) }

// SetVmailg moves v into VMAILG X(50).
func (d *Direction) SetVmailg(v string) { d.put(dOffVmailg, emailLen, v) }

// Vmaild returns VMAILD X(50), the right part of the département e-mail.
func (d *Direction) Vmaild() string { return d.get(dOffVmaild, emailLen) }

// SetVmaild moves v into VMAILD X(50).
func (d *Direction) SetVmaild(v string) { d.put(dOffVmaild, emailLen, v) }

// Vmgtas returns VMGTAS X(50), the left part of the TASA region e-mail.
func (d *Direction) Vmgtas() string { return d.get(dOffVmgtas, emailLen) }

// SetVmgtas moves v into VMGTAS X(50).
func (d *Direction) SetVmgtas(v string) { d.put(dOffVmgtas, emailLen, v) }

// Vmdtas returns VMDTAS X(50), the right part of the TASA region e-mail.
func (d *Direction) Vmdtas() string { return d.get(dOffVmdtas, emailLen) }

// SetVmdtas moves v into VMDTAS X(50).
func (d *Direction) SetVmdtas(v string) { d.put(dOffVmdtas, emailLen, v) }

// Libre returns the trailing free FILLER X(1936).
func (d *Direction) Libre() string { return d.get(dOffLibre, dLibreLen) }
