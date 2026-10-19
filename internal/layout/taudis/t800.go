// Package taudis holds the byte-exact layouts of the TAUDIS file (taux
// d'imposition des collectivités locales, 2018 vintage) that replaces the
// former VSAM KSDS.
package taudis

import (
	"fmt"
)

// EnteteLen is the length in bytes of the TAUDIS-ENTETE header (copybook T800).
const EnteteLen = 12

// RecordLen is the total length of a TAUDIS record (T800 + T84x = 3000).
const RecordLen = 3000

// BodyLen is the length of the record body following the header (T84C,
// T84D, T84G, T84R).
const BodyLen = RecordLen - EnteteLen

// Field offsets inside the 12-byte header. The layout is:
//
//	CLETX (12)
//	  DEPDIR (3) = CODEP X(2) + CODIR X(1)
//	  CLETX1 (9) = CCOCOM X(3) + CCOIFP X(3) + CCPPER X(3)   (commune/IFP/trésorerie)
//	  CLETX2 (9) REDEFINES CLETX1 = CTYGC X(1) + CCOGC X(3) + FILLER X(5)  (EPCI)
//	  CLETX3 (9) REDEFINES CLETX1 = CTYETS X(1) + GCOET1 9(3) + FILLER X(5) (TSE)
//	CLETX4 (12) REDEFINES CLETX = DIRCOM X(6) + FILLER X(6)
const (
	offCodep  = 0
	offCodir  = 2
	offDepdir = 0
	offCletx1 = 3
	offCcocom = 3
	offCcoifp = 6
	offCcpper = 9
	offCtygc  = 3
	offCcogc  = 4
	offCtyets = 3
	offGcoet1 = 4
	offFiller = 7 // FILLER X(5) of CLETX2/CLETX3
	offDircom = 0
	offFill4  = 6 // FILLER X(6) of CLETX4
)

// Entete is the TAUDIS-ENTETE record (copybook T800): a fixed 12-byte key
// header. All the COBOL REDEFINES views (CLETX, CLETX1..CLETX4) are
// exposed as accessor methods over the same underlying storage, so writing
// through one view is immediately visible through the others, exactly like
// the shared COBOL storage.
//
// MIGRATION_NOTE: the copybook uses a 'X' placeholder prefix substituted by
// COPY ... REPLACING (EFITAUX2 includes it three times as T-D, T-C and
// T-R). In Go each inclusion is simply a distinct Entete value; no name
// mangling is needed.
type Entete [EnteteLen]byte

// NewEntete returns a header initialised to spaces, as a COBOL
// INITIALIZE / MOVE SPACES on an alphanumeric group would.
func NewEntete() Entete {
	var e Entete
	for i := range e {
		e[i] = ' '
	}
	return e
}

// ParseEntete copies the first 12 bytes of b into a new header. It fails if
// b is shorter than EnteteLen.
func ParseEntete(b []byte) (Entete, error) {
	var e Entete
	if len(b) < EnteteLen {
		return e, fmt.Errorf("taudis: entete needs %d bytes, got %d", EnteteLen, len(b))
	}
	copy(e[:], b[:EnteteLen])
	return e, nil
}

// SplitRecord decodes a 3000-byte TAUDIS record into its 12-byte header and
// the 2988-byte body (to be interpreted by the T84C/T84D/T84G/T84R layouts).
// The returned body aliases rec.
func SplitRecord(rec []byte) (Entete, []byte, error) {
	if len(rec) != RecordLen {
		return Entete{}, nil, fmt.Errorf("taudis: record must be %d bytes, got %d", RecordLen, len(rec))
	}
	e, err := ParseEntete(rec)
	if err != nil {
		return Entete{}, nil, err
	}
	return e, rec[EnteteLen:], nil
}

// Bytes returns a copy of the raw 12 bytes (the CLETX / CLETX4 group).
func (e Entete) Bytes() []byte {
	out := make([]byte, EnteteLen)
	copy(out, e[:])
	return out
}

// Key returns the full 12-byte key as a string (group CLETX).
func (e Entete) Key() string { return string(e[:]) }

func (e *Entete) get(off, n int) string { return string(e[off : off+n]) }

// put performs a COBOL alphanumeric MOVE into a field: left-justified,
// truncated on the right, padded with spaces.
func (e *Entete) put(off, n int, v string) {
	for i := 0; i < n; i++ {
		if i < len(v) {
			e[off+i] = v[i]
		} else {
			e[off+i] = ' '
		}
	}
}

// ---- Direction (DEPDIR) ----

// Depdir returns DEPDIR X(3): département + direction.
func (e *Entete) Depdir() string { return e.get(offDepdir, 3) }

// SetDepdir moves v into DEPDIR X(3).
func (e *Entete) SetDepdir(v string) { e.put(offDepdir, 3, v) }

// Codep returns CODEP X(2): département code.
func (e *Entete) Codep() string { return e.get(offCodep, 2) }

// SetCodep moves v into CODEP X(2).
func (e *Entete) SetCodep(v string) { e.put(offCodep, 2, v) }

// Codir returns CODIR X(1): direction code.
func (e *Entete) Codir() string { return e.get(offCodir, 1) }

// SetCodir moves v into CODIR X(1).
func (e *Entete) SetCodir(v string) { e.put(offCodir, 1, v) }

// ---- CLETX1: commune / IFP / trésorerie ----

// Cletx1 returns the 9-byte CLETX1 group (shared storage with CLETX2/CLETX3).
func (e *Entete) Cletx1() string { return e.get(offCletx1, 9) }

// SetCletx1 moves v into the 9-byte CLETX1 group.
func (e *Entete) SetCletx1(v string) { e.put(offCletx1, 9, v) }

// Ccocom returns CCOCOM X(3): commune code.
func (e *Entete) Ccocom() string { return e.get(offCcocom, 3) }

// SetCcocom moves v into CCOCOM X(3).
func (e *Entete) SetCcocom(v string) { e.put(offCcocom, 3, v) }

// Ccoifp returns CCOIFP X(3): IFP code.
func (e *Entete) Ccoifp() string { return e.get(offCcoifp, 3) }

// SetCcoifp moves v into CCOIFP X(3).
func (e *Entete) SetCcoifp(v string) { e.put(offCcoifp, 3, v) }

// Ccpper returns CCPPER X(3): trésorerie (perception) code.
func (e *Entete) Ccpper() string { return e.get(offCcpper, 3) }

// SetCcpper moves v into CCPPER X(3).
func (e *Entete) SetCcpper(v string) { e.put(offCcpper, 3, v) }

// ---- CLETX2 (REDEFINES CLETX1): EPCI ----

// Ctygc returns CTYGC X(1): EPCI grouping type.
func (e *Entete) Ctygc() string { return e.get(offCtygc, 1) }

// SetCtygc moves v into CTYGC X(1).
func (e *Entete) SetCtygc(v string) { e.put(offCtygc, 1, v) }

// Ccogc returns CCOGC X(3): EPCI grouping code.
func (e *Entete) Ccogc() string { return e.get(offCcogc, 3) }

// SetCcogc moves v into CCOGC X(3).
func (e *Entete) SetCcogc(v string) { e.put(offCcogc, 3, v) }

// Filler2 returns the FILLER X(5) shared by CLETX2 and CLETX3 (bytes 8-12).
//
// MIGRATION_NOTE: after COPY REPLACING, the copybook's 'X'-FILLER names
// become real, duplicated data names; CLETX2 and CLETX3 FILLERs occupy the
// same bytes, so a single accessor covers both.
func (e *Entete) Filler2() string { return e.get(offFiller, 5) }

// SetFiller2 moves v into the FILLER X(5) of CLETX2/CLETX3.
func (e *Entete) SetFiller2(v string) { e.put(offFiller, 5, v) }

// ---- CLETX3 (REDEFINES CLETX1): établissement TSE ----

// Ctyets returns CTYETS X(1): TSE establishment type.
func (e *Entete) Ctyets() string { return e.get(offCtyets, 1) }

// SetCtyets moves v into CTYETS X(1).
func (e *Entete) SetCtyets(v string) { e.put(offCtyets, 1, v) }

// Gcoet1Raw returns the raw 3 bytes of GCOET1 PIC 9(3).
func (e *Entete) Gcoet1Raw() string { return e.get(offGcoet1, 3) }

// Gcoet1IsNumeric reports whether GCOET1 passes the COBOL NUMERIC class
// test for an unsigned zoned-decimal field (every byte a digit).
func (e *Entete) Gcoet1IsNumeric() bool {
	for _, c := range e[offGcoet1 : offGcoet1+3] {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// Gcoet1 returns the value of GCOET1 PIC 9(3), an unsigned zoned-decimal
// TSE establishment code. ok is false when the field is not NUMERIC; the
// value is then built from the low nibble of each byte, as an IBM COBOL
// arithmetic on invalid zoned data would, and callers should treat it as
// unreliable.
func (e *Entete) Gcoet1() (value int, ok bool) {
	ok = e.Gcoet1IsNumeric()
	for _, c := range e[offGcoet1 : offGcoet1+3] {
		d := int(c & 0x0F)
		if d > 9 {
			d = 9 // MIGRATION_NOTE: invalid nibble; clamp to keep result within 0..999
		}
		value = value*10 + d
	}
	return value, ok
}

// SetGcoet1 performs a COBOL numeric MOVE into GCOET1 PIC 9(3): the sign is
// dropped (unsigned receiver) and high-order digits beyond 3 are truncated.
func (e *Entete) SetGcoet1(v int) {
	if v < 0 {
		v = -v
	}
	v %= 1000
	e[offGcoet1] = byte('0' + v/100)
	e[offGcoet1+1] = byte('0' + (v/10)%10)
	e[offGcoet1+2] = byte('0' + v%10)
}

// ---- CLETX4 (REDEFINES CLETX): direction + commune ----

// Dircom returns DIRCOM X(6): direction (DEPDIR) followed by the commune
// code (CCOCOM).
func (e *Entete) Dircom() string { return e.get(offDircom, 6) }

// SetDircom moves v into DIRCOM X(6).
func (e *Entete) SetDircom(v string) { e.put(offDircom, 6, v) }

// Filler4 returns the FILLER X(6) of CLETX4 (bytes 7-12).
func (e *Entete) Filler4() string { return e.get(offFill4, 6) }

// SetFiller4 moves v into the FILLER X(6) of CLETX4.
func (e *Entete) SetFiller4(v string) { e.put(offFill4, 6, v) }
