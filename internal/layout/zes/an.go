// Package zes holds the byte-exact layouts of the ZES (MAJIC2) input
// segments consumed by the taxe foncière 2018 calculators.
package zes

import (
	"fmt"
)

// ANLen is the total length in bytes of a ZES AN segment
// (AIDFIC 13 + ADTSEG 387 = 400).
const ANLen = 400

// Group lengths of the AN segment.
const (
	// ANAidficLen is the length of the AIDFIC file-identifier group.
	ANAidficLen = 13
	// ANAdtsegLen is the length of the ADTSEG segment-body group.
	ANAdtsegLen = 387
	// ANFillerLen is the length of the named 'X'-FILLER area inside ADTSEG.
	ANFillerLen = 377
)

// Zero-based byte offsets of the AN fields inside the 400-byte segment.
//
//	AIDFIC  X(13)   offsets  0..12  (COBOL positions   1..13)
//	  JAN     9(4)           0..3
//	  DEPDIR  X(3)           4..6   = CODEP X(2) + CODIR X(1)
//	  CCOCOM  X(3)           7..9
//	  CCOIFP  X(3)          10..12
//	ADTSEG  X(387)  offsets 13..399 (COBOL positions  14..400)
//	  CCOENR  X(2)          13..14
//	  XDDAY6  X(6)          15..20
//	  FILLER  X(377)        21..397
//	  DNUURC  X(2)         398..399
const (
	anOffAidfic = 0
	anOffJan    = 0
	anOffDepdir = 4
	anOffCodep  = 4
	anOffCodir  = 6
	anOffCcocom = 7
	anOffCcoifp = 10
	anOffAdtseg = 13
	anOffCcoenr = 13
	anOffXdday6 = 15
	anOffFiller = 21
	anOffDnuurc = 398
)

// AN is the ZES AN segment (copybook XBXTDAN): a fixed 400-byte MAJIC2
// record made of the AIDFIC file identifier followed by the ADTSEG body.
// Fields are exposed as accessors over the raw storage, so group views
// (AIDFIC, DEPDIR, ADTSEG) and their elementary fields share bytes exactly
// like COBOL storage, and bytes never touched by a setter round-trip
// unchanged.
//
// MIGRATION_NOTE: the copybook uses an 'X' placeholder prefix substituted
// by COPY ... REPLACING. In Go each inclusion is simply a distinct AN
// value; no name mangling is needed. 'X'-FILLER becomes a real data name
// after substitution, so it is exposed through Filler/SetFiller.
type AN [ANLen]byte

// NewAN returns a segment initialised to spaces, as a COBOL MOVE SPACES on
// the alphanumeric group would.
func NewAN() AN {
	var a AN
	for i := range a {
		a[i] = ' '
	}
	return a
}

// ParseAN copies a 400-byte buffer into a new AN segment. It fails if b is
// not exactly ANLen bytes long.
func ParseAN(b []byte) (AN, error) {
	var a AN
	if len(b) != ANLen {
		return a, fmt.Errorf("zes: AN segment must be %d bytes, got %d", ANLen, len(b))
	}
	copy(a[:], b)
	return a, nil
}

// Bytes returns a copy of the raw 400 bytes of the segment.
func (a *AN) Bytes() []byte {
	out := make([]byte, ANLen)
	copy(out, a[:])
	return out
}

func (a *AN) get(off, n int) string { return string(a[off : off+n]) }

// put performs a COBOL alphanumeric MOVE into a field: left-justified,
// truncated on the right, padded with spaces.
func (a *AN) put(off, n int, v string) {
	for i := 0; i < n; i++ {
		if i < len(v) {
			a[off+i] = v[i]
		} else {
			a[off+i] = ' '
		}
	}
}

// ---- AIDFIC: file identifier (positions 1-13) ----

// Aidfic returns the 13-byte AIDFIC group (MAJIC2 file identifier).
func (a *AN) Aidfic() string { return a.get(anOffAidfic, ANAidficLen) }

// SetAidfic moves v into the 13-byte AIDFIC group.
func (a *AN) SetAidfic(v string) { a.put(anOffAidfic, ANAidficLen, v) }

// JanRaw returns the raw 4 bytes of JAN PIC 9(4) (year).
func (a *AN) JanRaw() string { return a.get(anOffJan, 4) }

// JanIsNumeric reports whether JAN passes the COBOL NUMERIC class test for
// an unsigned zoned-decimal field (every byte a digit).
func (a *AN) JanIsNumeric() bool {
	for _, c := range a[anOffJan : anOffJan+4] {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// Jan returns the value of JAN PIC 9(4) (year). ok is false when the field
// is not NUMERIC; the value is then built from the low nibble of each byte
// (clamped to 9), as IBM COBOL arithmetic on invalid zoned data would, and
// callers should treat it as unreliable.
func (a *AN) Jan() (value int, ok bool) {
	ok = a.JanIsNumeric()
	for _, c := range a[anOffJan : anOffJan+4] {
		d := int(c & 0x0F)
		if d > 9 {
			d = 9 // MIGRATION_NOTE: invalid nibble; clamp to keep result within 0..9999
		}
		value = value*10 + d
	}
	return value, ok
}

// SetJan performs a COBOL numeric MOVE into JAN PIC 9(4): the sign is
// dropped (unsigned receiver) and high-order digits beyond 4 are truncated.
func (a *AN) SetJan(v int) {
	if v < 0 {
		v = -v
	}
	v %= 10000
	for i := 3; i >= 0; i-- {
		a[anOffJan+i] = byte('0' + v%10)
		v /= 10
	}
}

// Depdir returns DEPDIR X(3): département + direction group.
func (a *AN) Depdir() string { return a.get(anOffDepdir, 3) }

// SetDepdir moves v into DEPDIR X(3).
func (a *AN) SetDepdir(v string) { a.put(anOffDepdir, 3, v) }

// Codep returns CODEP X(2): département code.
func (a *AN) Codep() string { return a.get(anOffCodep, 2) }

// SetCodep moves v into CODEP X(2).
func (a *AN) SetCodep(v string) { a.put(anOffCodep, 2, v) }

// Codir returns CODIR X(1): direction code.
func (a *AN) Codir() string { return a.get(anOffCodir, 1) }

// SetCodir moves v into CODIR X(1).
func (a *AN) SetCodir(v string) { a.put(anOffCodir, 1, v) }

// Ccocom returns CCOCOM X(3): MAJIC2 commune code.
func (a *AN) Ccocom() string { return a.get(anOffCcocom, 3) }

// SetCcocom moves v into CCOCOM X(3).
func (a *AN) SetCcocom(v string) { a.put(anOffCcocom, 3, v) }

// Ccoifp returns CCOIFP X(3): absorbed-commune code.
func (a *AN) Ccoifp() string { return a.get(anOffCcoifp, 3) }

// SetCcoifp moves v into CCOIFP X(3).
func (a *AN) SetCcoifp(v string) { a.put(anOffCcoifp, 3, v) }

// ---- ADTSEG: segment body (positions 14-400) ----

// Adtseg returns the 387-byte ADTSEG group (segment body).
func (a *AN) Adtseg() string { return a.get(anOffAdtseg, ANAdtsegLen) }

// SetAdtseg moves v into the 387-byte ADTSEG group.
func (a *AN) SetAdtseg(v string) { a.put(anOffAdtseg, ANAdtsegLen, v) }

// Ccoenr returns CCOENR X(2): record code (record-type discriminator).
func (a *AN) Ccoenr() string { return a.get(anOffCcoenr, 2) }

// SetCcoenr moves v into CCOENR X(2).
func (a *AN) SetCcoenr(v string) { a.put(anOffCcoenr, 2, v) }

// Xdday6 returns XDDAY6 X(6): date of the day on 6 characters.
func (a *AN) Xdday6() string { return a.get(anOffXdday6, 6) }

// SetXdday6 moves v into XDDAY6 X(6).
func (a *AN) SetXdday6(v string) { a.put(anOffXdday6, 6, v) }

// Filler returns the 377-byte named 'X'-FILLER area.
func (a *AN) Filler() string { return a.get(anOffFiller, ANFillerLen) }

// SetFiller moves v into the 377-byte 'X'-FILLER area.
func (a *AN) SetFiller(v string) { a.put(anOffFiller, ANFillerLen, v) }

// Dnuurc returns DNUURC X(2): CDIF regrouping unit number (UR).
func (a *AN) Dnuurc() string { return a.get(anOffDnuurc, 2) }

// SetDnuurc moves v into DNUURC X(2).
func (a *AN) SetDnuurc(v string) { a.put(anOffDnuurc, 2, v) }
