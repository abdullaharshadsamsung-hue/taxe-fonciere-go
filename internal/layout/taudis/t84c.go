package taudis

import (
	"fmt"
)

// CommuneLetter is the record-type discriminator (DLETAR) of a TAUDIS
// Commune article.
const CommuneLetter = "C"

// PplomScale is the scaling factor of the PIC 9V99 PPLOM coefficient: the
// accessor returns the value multiplied by 100 (two implied decimals). The
// PIC 9(3)V9(6) neutralisation coefficients use RateScale (six decimals).
const PplomScale = 100

// Number of OCCURS entries in the Commune layout.
const (
	// SyndicateCount is the OCCURS count of the LISTSY table.
	SyndicateCount = 10
	// AgrjaCount is the OCCURS count of the AGRJAO table (years N-4 .. N).
	AgrjaCount = 5
)

// Byte offsets of the DONNEES-COMMUNE fields, relative to the start of the
// 2988-byte body (i.e. the record offset minus EnteteLen). Every data field
// is preceded by a 1-byte 'X'-FILLER separator, except VMAILD and the
// subordinate items of the JDAHO1 / JDAMI1 / JDAMJO / ACOGC groups.
const (
	cOffDletar = 1   // X
	cOffDnultf = 3   // X
	cOffDnuro  = 5   // 9(3)
	cOffGcoanu = 9   // X
	cOffDacrol = 11  // X(3)
	cOffGcovef = 15  // X
	cOffGorvef = 17  // X
	cOffCcofdl = 19  // X
	cOffGokcu  = 21  // X
	cOffNbiftr = 23  // 9(3)
	cOffNbrifp = 27  // 9(3)
	cOffGtocfu = 31  // X
	cOffGcndif = 33  // X
	cOffGcbdif = 35  // X
	cOffJdaho1 = 37  // group (8) = JSSHOM 9(2) JANHO1 9(2) JMOHOM 9(2) JJRHOM 9(2)
	cOffJsshom = 37  // 9(2)
	cOffJanho1 = 39  // 9(2)
	cOffJmohom = 41  // 9(2)
	cOffJjrhom = 43  // 9(2)
	cOffJdami1 = 46  // group (8) = JSSMIR X(2) JANMIR X(2) JMOMIR X(2) JJRMIR X(2)
	cOffJssmir = 46  // X(2)
	cOffJanmir = 48  // X(2)
	cOffJmomir = 50  // X(2)
	cOffJjrmir = 52  // X(2)
	cOffJdamjo = 55  // group (8) = JSSMJO X(2) JANMJO X(2) JMOMJO X(2) JJRMJO X(2)
	cOffJssmjo = 55  // X(2)
	cOffJanmjo = 57  // X(2)
	cOffJmomjo = 59  // X(2)
	cOffJjrmjo = 61  // X(2)
	cOffDlicom = 64  // X(30)
	cOffDsirec = 95  // X(14)
	cOffGnidep = 110 // X
	cOffGtomgp = 112 // X
	cOffGcepci = 114 // X
	cOffGtidf  = 116 // X
	cOffCtlorr = 118 // X
	cOffGsvrec = 120 // X
	cOffCcobmj = 122 // X(4)
	cOffNpocom = 127 // 9(7)
	cOffGmjpos = 135 // X
	cOffPmjpo  = 137 // 9(3)
	cOffDcmzo1 = 141 // X
	cOffMfsid1 = 143 // 9(10) (sole item of group MFSIDF)
	cOffDvlpom = 154 // 9(6)
	cOffPplom  = 161 // 9V99
	cOffNtxom  = 165 // 9
	cOffNtxom1 = 167 // 9
	cOffDsedom = 169 // 9(9)
	cOffDsednb = 179 // 9(9)
	cOffGrocom = 189 // X
	cOffPcfidf = 191 // X(3)
	cOffGtocom = 195 // X
	cOffAcogc  = 197 // group (4) = CTYGC1 X + CCOIC 9(3)
	cOffCtygc1 = 197 // X
	cOffCcoic  = 198 // 9(3)
	cOffRfepci = 202 // X(3)
	cOffGtocob = 206 // X
	cOffGtocon = 208 // X
	cOffGcoetp = 210 // 9(3)
	cOffListsy = 213 // OCCURS 10, 31 bytes each
	cOffAgrja  = 523 // OCCURS 5, 9 bytes each (AGRJA / AGRJAR REDEFINES)
	cOffVmailg = 569 // X(50)
	cOffVmaild = 619 // X(50) (no separator)
	cOffCncom  = 670 // 9(3)V9(6)
	cOffCnepci = 680 // 9(3)V9(6)
	cOffCndep  = 690 // 9(3)V9(6)
	cOffCnteom = 700 // 9(3)V9(6)
	cOffCntse  = 710 // 9(3)V9(6)
	cOffLibre  = 719 // FILLER X(2269)
	cLibreLen  = 2269

	cCoefLen   = 9
	cEmailLen  = 50
	cSiretLen  = 14
	cDateLen   = 8
	cDlicomLen = 30
)

// Layout of one LISTSY entry (31 bytes), offsets relative to the entry.
const (
	syndEntryLen    = 31
	syndOffSirenSyn = 1  // X(9)
	syndOffDlisyn   = 11 // X(20)
	syndSirenLen    = 9
	syndDlisynLen   = 20
)

// Layout of one AGRJAO entry (9 bytes), offsets relative to the entry.
const (
	agrjaEntryLen  = 9
	agrjaOffGrinsn = 1 // 9(4)
	agrjaOffGrrimn = 5 // 9(4)
	agrjaYearLen   = 4
)

// Compile-time guard: the layout must end exactly at BodyLen (2988 bytes).
var _ = [1]struct{}{}[cOffLibre+cLibreLen-BodyLen]

// Compile-time guards: the OCCURS tables must end where the next field's
// separator starts.
var (
	_ = [1]struct{}{}[cOffListsy+SyndicateCount*syndEntryLen-cOffAgrja]
	_ = [1]struct{}{}[cOffAgrja+AgrjaCount*agrjaEntryLen+1-cOffVmailg]
)

// Commune is the DONNEES-COMMUNE body (copybook T84C) of a TAUDIS 'C'
// record: the 2988 bytes that follow the 12-byte Entete (which carries
// DEPDIR and CCOCOM). Fields are exposed as accessor methods over the raw
// storage so that encoding and decoding are byte-exact and FILLER bytes
// (separators and the trailing free zone) survive a round trip untouched.
//
// MIGRATION_NOTE: the copybook's 'X'- prefix is a COPY REPLACING placeholder
// (EFITAUX2 includes it as T-C-); in Go each inclusion is simply a distinct
// Commune value. The AGRJA / AGRJAR REDEFINES pair is exposed both as named
// accessors (Jgcin1..5, Jgcri1..5) and as an indexed view (Agrja), all
// sharing the same bytes.
type Commune [BodyLen]byte

// NewCommune returns a Commune body whose alphanumeric fields and FILLERs
// are spaces, numeric fields are zeros, and whose article letter DLETAR is
// 'C'.
func NewCommune() Commune {
	var c Commune
	for i := range c {
		c[i] = ' '
	}
	for _, f := range communeNumericFields() {
		c.setNum(f[0], f[1], 0)
	}
	c.SetDletar(CommuneLetter)
	return c
}

// communeNumericFields lists the (offset, length) of every unsigned
// zoned-decimal field of the layout.
func communeNumericFields() [][2]int {
	fields := [][2]int{
		{cOffDnuro, 3}, {cOffNbiftr, 3}, {cOffNbrifp, 3}, {cOffJdaho1, cDateLen},
		{cOffNpocom, 7}, {cOffPmjpo, 3}, {cOffMfsid1, 10}, {cOffDvlpom, 6},
		{cOffPplom, 3}, {cOffNtxom, 1}, {cOffNtxom1, 1}, {cOffDsedom, 9},
		{cOffDsednb, 9}, {cOffCcoic, 3}, {cOffGcoetp, 3},
		{cOffCncom, cCoefLen}, {cOffCnepci, cCoefLen}, {cOffCndep, cCoefLen},
		{cOffCnteom, cCoefLen}, {cOffCntse, cCoefLen},
	}
	for i := 0; i < AgrjaCount; i++ {
		base := cOffAgrja + i*agrjaEntryLen
		fields = append(fields,
			[2]int{base + agrjaOffGrinsn, agrjaYearLen},
			[2]int{base + agrjaOffGrrimn, agrjaYearLen})
	}
	return fields
}

// ParseCommune copies a 2988-byte body into a Commune. It fails when b does
// not have exactly BodyLen bytes.
func ParseCommune(b []byte) (Commune, error) {
	var c Commune
	if len(b) != BodyLen {
		return c, fmt.Errorf("taudis: commune body must be %d bytes, got %d", BodyLen, len(b))
	}
	copy(c[:], b)
	return c, nil
}

// Bytes returns a copy of the raw 2988 bytes (encoded layout).
func (c *Commune) Bytes() []byte {
	out := make([]byte, BodyLen)
	copy(out, c[:])
	return out
}

// IsCommune reports whether DLETAR holds the 'C' article letter.
func (c *Commune) IsCommune() bool { return c.Dletar() == CommuneLetter }

// ---- low-level helpers ----

func (c *Commune) get(off, n int) string { return string(c[off : off+n]) }

// put performs a COBOL alphanumeric MOVE: left-justified, truncated on the
// right, padded with spaces.
func (c *Commune) put(off, n int, v string) {
	for i := 0; i < n; i++ {
		if i < len(v) {
			c[off+i] = v[i]
		} else {
			c[off+i] = ' '
		}
	}
}

// isNum is the COBOL NUMERIC class test for an unsigned DISPLAY field.
func (c *Commune) isNum(off, n int) bool {
	for _, b := range c[off : off+n] {
		if b < '0' || b > '9' {
			return false
		}
	}
	return true
}

// num decodes an unsigned zoned-decimal field. ok is false when the field
// fails the NUMERIC test; the value is then built from the low nibble of
// each byte (as IBM COBOL arithmetic on invalid zoned data would) and must
// be treated as unreliable.
func (c *Commune) num(off, n int) (uint64, bool) {
	ok := c.isNum(off, n)
	var v uint64
	for _, b := range c[off : off+n] {
		dig := uint64(b & 0x0F)
		if dig > 9 {
			dig = 9 // MIGRATION_NOTE: invalid nibble clamped to keep the value in range
		}
		v = v*10 + dig
	}
	return v, ok
}

// setNum performs a COBOL numeric MOVE into an unsigned DISPLAY field:
// high-order digits beyond n are truncated, zero-filled on the left.
func (c *Commune) setNum(off, n int, v uint64) {
	for i := off + n - 1; i >= off; i-- {
		c[i] = byte('0' + v%10)
		v /= 10
	}
}

// ---- taxation data ----

// Dletar returns DLETAR X, the article letter ('C' for Commune).
func (c *Commune) Dletar() string { return c.get(cOffDletar, 1) }

// SetDletar moves v into DLETAR X.
func (c *Commune) SetDletar(v string) { c.put(cOffDletar, 1, v) }

// Dnultf returns DNULTF X, the lot number.
func (c *Commune) Dnultf() string { return c.get(cOffDnultf, 1) }

// SetDnultf moves v into DNULTF X.
func (c *Commune) SetDnultf(v string) { c.put(cOffDnultf, 1, v) }

// Dnuro returns DNURO 9(3), the role number.
func (c *Commune) Dnuro() (uint64, bool) { return c.num(cOffDnuro, 3) }

// DnuroRaw returns the raw 3 bytes of DNURO.
func (c *Commune) DnuroRaw() string { return c.get(cOffDnuro, 3) }

// SetDnuro moves v into DNURO 9(3).
func (c *Commune) SetDnuro(v uint64) { c.setNum(cOffDnuro, 3, v) }

// Gcoanu returns GCOANU X, the cancellation-reconfiguration / regularisation
// code ('A' = whole commune, 'I' = IFP).
func (c *Commune) Gcoanu() string { return c.get(cOffGcoanu, 1) }

// SetGcoanu moves v into GCOANU X.
func (c *Commune) SetGcoanu(v string) { c.put(cOffGcoanu, 1, v) }

// Dacrol returns DACROL X(3), the former role number for a
// cancellation-reconfiguration.
func (c *Commune) Dacrol() string { return c.get(cOffDacrol, 3) }

// SetDacrol moves v into DACROL X(3).
func (c *Commune) SetDacrol(v string) { c.put(cOffDacrol, 3, v) }

// Gcovef returns GCOVEF X, the EFICT validation code ('V' = valid).
func (c *Commune) Gcovef() string { return c.get(cOffGcovef, 1) }

// SetGcovef moves v into GCOVEF X.
func (c *Commune) SetGcovef(v string) { c.put(cOffGcovef, 1, v) }

// Gorvef returns GORVEF X, the EFICT validation origin.
func (c *Commune) Gorvef() string { return c.get(cOffGorvef, 1) }

// SetGorvef moves v into GORVEF X.
func (c *Commune) SetGorvef(v string) { c.put(cOffGorvef, 1, v) }

// Ccofdl returns CCOFDL X, the EFICT pass code.
func (c *Commune) Ccofdl() string { return c.get(cOffCcofdl, 1) }

// SetCcofdl moves v into CCOFDL X.
func (c *Commune) SetCcofdl(v string) { c.put(cOffCcofdl, 1, v) }

// Gokcu returns GOKCU X, the EPCI validation ('V' = valid, or blank).
func (c *Commune) Gokcu() string { return c.get(cOffGokcu, 1) }

// SetGokcu moves v into GOKCU X.
func (c *Commune) SetGokcu(v string) { c.put(cOffGokcu, 1, v) }

// ---- commune characteristics ----

// Nbiftr returns NBIFTR 9(3), the number of IFP-trésorerie of the commune.
func (c *Commune) Nbiftr() (uint64, bool) { return c.num(cOffNbiftr, 3) }

// SetNbiftr moves v into NBIFTR 9(3).
func (c *Commune) SetNbiftr(v uint64) { c.setNum(cOffNbiftr, 3, v) }

// Nbrifp returns NBRIFP 9(3), the number of IFP of the commune (1 = no IFP,
// >1 = number of IFP).
func (c *Commune) Nbrifp() (uint64, bool) { return c.num(cOffNbrifp, 3) }

// SetNbrifp moves v into NBRIFP 9(3).
func (c *Commune) SetNbrifp(v uint64) { c.setNum(cOffNbrifp, 3, v) }

// Gtocfu returns GTOCFU X, the first-year-of-merger flag.
func (c *Commune) Gtocfu() string { return c.get(cOffGtocfu, 1) }

// SetGtocfu moves v into GTOCFU X.
func (c *Commune) SetGtocfu(v string) { c.put(cOffGtocfu, 1, v) }

// Gcndif returns GCNDIF X, the non-built non-publishable code.
func (c *Commune) Gcndif() string { return c.get(cOffGcndif, 1) }

// SetGcndif moves v into GCNDIF X.
func (c *Commune) SetGcndif(v string) { c.put(cOffGcndif, 1, v) }

// Gcbdif returns GCBDIF X, the built non-publishable code.
func (c *Commune) Gcbdif() string { return c.get(cOffGcbdif, 1) }

// SetGcbdif moves v into GCBDIF X.
func (c *Commune) SetGcbdif(v string) { c.put(cOffGcbdif, 1, v) }

// Jdaho1 returns the raw 8-byte JDAHO1 group (homologation date SSAAMMJJ).
func (c *Commune) Jdaho1() string { return c.get(cOffJdaho1, cDateLen) }

// SetJdaho1 moves v into the JDAHO1 group (alphanumeric group MOVE).
func (c *Commune) SetJdaho1(v string) { c.put(cOffJdaho1, cDateLen, v) }

// Jsshom returns JSSHOM 9(2), the homologation century.
func (c *Commune) Jsshom() (uint64, bool) { return c.num(cOffJsshom, 2) }

// SetJsshom moves v into JSSHOM 9(2).
func (c *Commune) SetJsshom(v uint64) { c.setNum(cOffJsshom, 2, v) }

// Janho1 returns JANHO1 9(2), the homologation year within the century.
func (c *Commune) Janho1() (uint64, bool) { return c.num(cOffJanho1, 2) }

// SetJanho1 moves v into JANHO1 9(2).
func (c *Commune) SetJanho1(v uint64) { c.setNum(cOffJanho1, 2, v) }

// Jmohom returns JMOHOM 9(2), the homologation month.
func (c *Commune) Jmohom() (uint64, bool) { return c.num(cOffJmohom, 2) }

// SetJmohom moves v into JMOHOM 9(2).
func (c *Commune) SetJmohom(v uint64) { c.setNum(cOffJmohom, 2, v) }

// Jjrhom returns JJRHOM 9(2), the homologation day.
func (c *Commune) Jjrhom() (uint64, bool) { return c.num(cOffJjrhom, 2) }

// SetJjrhom moves v into JJRHOM 9(2).
func (c *Commune) SetJjrhom(v uint64) { c.setNum(cOffJjrhom, 2, v) }

// Jdami1 returns the raw 8-byte JDAMI1 group (issue date SSAAMMJJ).
func (c *Commune) Jdami1() string { return c.get(cOffJdami1, cDateLen) }

// SetJdami1 moves v into the JDAMI1 group.
func (c *Commune) SetJdami1(v string) { c.put(cOffJdami1, cDateLen, v) }

// Jssmir returns JSSMIR X(2), the issue-date century.
func (c *Commune) Jssmir() string { return c.get(cOffJssmir, 2) }

// SetJssmir moves v into JSSMIR X(2).
func (c *Commune) SetJssmir(v string) { c.put(cOffJssmir, 2, v) }

// Janmir returns JANMIR X(2), the issue-date year within the century.
func (c *Commune) Janmir() string { return c.get(cOffJanmir, 2) }

// SetJanmir moves v into JANMIR X(2).
func (c *Commune) SetJanmir(v string) { c.put(cOffJanmir, 2, v) }

// Jmomir returns JMOMIR X(2), the issue-date month.
func (c *Commune) Jmomir() string { return c.get(cOffJmomir, 2) }

// SetJmomir moves v into JMOMIR X(2).
func (c *Commune) SetJmomir(v string) { c.put(cOffJmomir, 2, v) }

// Jjrmir returns JJRMIR X(2), the issue-date day.
func (c *Commune) Jjrmir() string { return c.get(cOffJjrmir, 2) }

// SetJjrmir moves v into JJRMIR X(2).
func (c *Commune) SetJjrmir(v string) { c.put(cOffJjrmir, 2, v) }

// Jdamjo returns the raw 8-byte JDAMJO group (surcharge date SSAAMMJJ).
func (c *Commune) Jdamjo() string { return c.get(cOffJdamjo, cDateLen) }

// SetJdamjo moves v into the JDAMJO group.
func (c *Commune) SetJdamjo(v string) { c.put(cOffJdamjo, cDateLen, v) }

// Jssmjo returns JSSMJO X(2), the surcharge-date century.
func (c *Commune) Jssmjo() string { return c.get(cOffJssmjo, 2) }

// SetJssmjo moves v into JSSMJO X(2).
func (c *Commune) SetJssmjo(v string) { c.put(cOffJssmjo, 2, v) }

// Janmjo returns JANMJO X(2), the surcharge-date year within the century.
func (c *Commune) Janmjo() string { return c.get(cOffJanmjo, 2) }

// SetJanmjo moves v into JANMJO X(2).
func (c *Commune) SetJanmjo(v string) { c.put(cOffJanmjo, 2, v) }

// Jmomjo returns JMOMJO X(2), the surcharge-date month.
func (c *Commune) Jmomjo() string { return c.get(cOffJmomjo, 2) }

// SetJmomjo moves v into JMOMJO X(2).
func (c *Commune) SetJmomjo(v string) { c.put(cOffJmomjo, 2, v) }

// Jjrmjo returns JJRMJO X(2), the surcharge-date day.
func (c *Commune) Jjrmjo() string { return c.get(cOffJjrmjo, 2) }

// SetJjrmjo moves v into JJRMJO X(2).
func (c *Commune) SetJjrmjo(v string) { c.put(cOffJjrmjo, 2, v) }

// Dlicom returns DLICOM X(30), the commune label.
func (c *Commune) Dlicom() string { return c.get(cOffDlicom, cDlicomLen) }

// SetDlicom moves v into DLICOM X(30).
func (c *Commune) SetDlicom(v string) { c.put(cOffDlicom, cDlicomLen, v) }

// Dsirec returns DSIREC X(14), the commune SIRET.
func (c *Commune) Dsirec() string { return c.get(cOffDsirec, cSiretLen) }

// SetDsirec moves v into DSIREC X(14).
func (c *Commune) SetDsirec(v string) { c.put(cOffDsirec, cSiretLen, v) }

// Gnidep returns GNIDEP X, the commune-not-taxable-by-département flag.
func (c *Commune) Gnidep() string { return c.get(cOffGnidep, 1) }

// SetGnidep moves v into GNIDEP X.
func (c *Commune) SetGnidep(v string) { c.put(cOffGnidep, 1, v) }

// Gtomgp returns GTOMGP X, the Métropole du Grand Paris membership flag.
func (c *Commune) Gtomgp() string { return c.get(cOffGtomgp, 1) }

// SetGtomgp moves v into GTOMGP X.
func (c *Commune) SetGtomgp(v string) { c.put(cOffGtomgp, 1, v) }

// Gcepci returns GCEPCI X, the first-year-of-TPU-membership flag.
func (c *Commune) Gcepci() string { return c.get(cOffGcepci, 1) }

// SetGcepci moves v into GCEPCI X.
func (c *Commune) SetGcepci(v string) { c.put(cOffGcepci, 1, v) }

// Gtidf returns GTIDF X, the IDF fund payment flag.
func (c *Commune) Gtidf() string { return c.get(cOffGtidf, 1) }

// SetGtidf moves v into GTIDF X.
func (c *Commune) SetGtidf(v string) { c.put(cOffGtidf, 1, v) }

// Ctlorr returns CTLORR X, the census code ('R' = surveyed, 'N' = not).
func (c *Commune) Ctlorr() string { return c.get(cOffCtlorr, 1) }

// SetCtlorr moves v into CTLORR X.
func (c *Commune) SetCtlorr(v string) { c.put(cOffCtlorr, 1, v) }

// Gsvrec returns GSVREC X, the TF management service ('I' = SIP, 'F' = CDIF).
func (c *Commune) Gsvrec() string { return c.get(cOffGsvrec, 1) }

// SetGsvrec moves v into GSVREC X.
func (c *Commune) SetGsvrec(v string) { c.put(cOffGsvrec, 1, v) }

// Ccobmj returns CCOBMJ X(4), the CDIF or SIP code (MAJIC base code).
func (c *Commune) Ccobmj() string { return c.get(cOffCcobmj, 4) }

// SetCcobmj moves v into CCOBMJ X(4).
func (c *Commune) SetCcobmj(v string) { c.put(cOffCcobmj, 4, v) }

// Npocom returns NPOCOM 9(7), the commune population.
func (c *Commune) Npocom() (uint64, bool) { return c.num(cOffNpocom, 7) }

// SetNpocom moves v into NPOCOM 9(7).
func (c *Commune) SetNpocom(v uint64) { c.setNum(cOffNpocom, 7, v) }

// Gmjpos returns GMJPOS X, the POS applicability flag ('P' = yes, or blank).
func (c *Commune) Gmjpos() string { return c.get(cOffGmjpos, 1) }

// SetGmjpos moves v into GMJPOS X.
func (c *Commune) SetGmjpos(v string) { c.put(cOffGmjpos, 1, v) }

// Pmjpo returns PMJPO 9(3), the POS rate.
func (c *Commune) Pmjpo() (uint64, bool) { return c.num(cOffPmjpo, 3) }

// SetPmjpo moves v into PMJPO 9(3).
func (c *Commune) SetPmjpo(v uint64) { c.setNum(cOffPmjpo, 3, v) }

// Dcmzo1 returns DCMZO1 X, the TC surcharge indicator ('N' = national,
// 'D' = by municipal deliberation, blank = none).
func (c *Commune) Dcmzo1() string { return c.get(cOffDcmzo1, 1) }

// SetDcmzo1 moves v into DCMZO1 X.
func (c *Commune) SetDcmzo1(v string) { c.put(cOffDcmzo1, 1, v) }

// Mfsid1 returns MFSID1 9(10), the IDF fund payment amount (sole item of
// the MFSIDF group).
func (c *Commune) Mfsid1() (uint64, bool) { return c.num(cOffMfsid1, 10) }

// SetMfsid1 moves v into MFSID1 9(10).
func (c *Commune) SetMfsid1(v uint64) { c.setNum(cOffMfsid1, 10, v) }

// Mfsidf returns the raw 10-byte MFSIDF group.
func (c *Commune) Mfsidf() string { return c.get(cOffMfsid1, 10) }

// Dvlpom returns DVLPOM 9(6), the average rental value for TOM capping.
func (c *Commune) Dvlpom() (uint64, bool) { return c.num(cOffDvlpom, 6) }

// SetDvlpom moves v into DVLPOM 9(6).
func (c *Commune) SetDvlpom(v uint64) { c.setNum(cOffDvlpom, 6, v) }

// Pplom returns PPLOM 9V99, the TOM capping coefficient, scaled by
// PplomScale (e.g. 1.25 -> 125).
func (c *Commune) Pplom() (uint64, bool) { return c.num(cOffPplom, 3) }

// SetPplom moves a PplomScale-scaled value into PPLOM 9V99.
func (c *Commune) SetPplom(v uint64) { c.setNum(cOffPplom, 3, v) }

// Ntxom returns NTXOM 9, the number of FDL OM rates (not set for an IFP).
func (c *Commune) Ntxom() (uint64, bool) { return c.num(cOffNtxom, 1) }

// SetNtxom moves v into NTXOM 9.
func (c *Commune) SetNtxom(v uint64) { c.setNum(cOffNtxom, 1, v) }

// Ntxom1 returns NTXOM1 9, the number of OM rates counted by EFIBA.
func (c *Commune) Ntxom1() (uint64, bool) { return c.num(cOffNtxom1, 1) }

// SetNtxom1 moves v into NTXOM1 9.
func (c *Commune) SetNtxom1(v uint64) { c.setNum(cOffNtxom1, 1, v) }

// Dsedom returns DSEDOM 9(9), the DOM built-property exemption threshold.
func (c *Commune) Dsedom() (uint64, bool) { return c.num(cOffDsedom, 9) }

// SetDsedom moves v into DSEDOM 9(9).
func (c *Commune) SetDsedom(v uint64) { c.setNum(cOffDsedom, 9, v) }

// Dsednb returns DSEDNB 9(9), the DOM non-built exemption threshold.
func (c *Commune) Dsednb() (uint64, bool) { return c.num(cOffDsednb, 9) }

// SetDsednb moves v into DSEDNB 9(9).
func (c *Commune) SetDsednb(v uint64) { c.setNum(cOffDsednb, 9, v) }

// Grocom returns GROCOM X, the commune role-copy option.
func (c *Commune) Grocom() string { return c.get(cOffGrocom, 1) }

// SetGrocom moves v into GROCOM X.
func (c *Commune) SetGrocom(v string) { c.put(cOffGrocom, 1, v) }

// Pcfidf returns PCFIDF X(3), the IDF modulation coefficient
// (alphanumeric in the copybook).
func (c *Commune) Pcfidf() string { return c.get(cOffPcfidf, 3) }

// SetPcfidf moves v into PCFIDF X(3).
func (c *Commune) SetPcfidf(v string) { c.put(cOffPcfidf, 3, v) }

// Gtocom returns GTOCOM X, the EFIRF progress flag ('*' = EFIRF10,
// blank = EFIRF25).
func (c *Commune) Gtocom() string { return c.get(cOffGtocom, 1) }

// SetGtocom moves v into GTOCOM X.
func (c *Commune) SetGtocom(v string) { c.put(cOffGtocom, 1, v) }

// ---- EPCI / TSE links ----

// Acogc returns the raw 4-byte ACOGC group (EPCI code of the commune).
func (c *Commune) Acogc() string { return c.get(cOffAcogc, 4) }

// SetAcogc moves v into the ACOGC group.
func (c *Commune) SetAcogc(v string) { c.put(cOffAcogc, 4, v) }

// Ctygc1 returns CTYGC1 X, the EPCI grouping type.
func (c *Commune) Ctygc1() string { return c.get(cOffCtygc1, 1) }

// SetCtygc1 moves v into CTYGC1 X.
func (c *Commune) SetCtygc1(v string) { c.put(cOffCtygc1, 1, v) }

// Ccoic returns CCOIC 9(3), the EPCI grouping code.
func (c *Commune) Ccoic() (uint64, bool) { return c.num(cOffCcoic, 3) }

// CcoicRaw returns the raw 3 bytes of CCOIC.
func (c *Commune) CcoicRaw() string { return c.get(cOffCcoic, 3) }

// SetCcoic moves v into CCOIC 9(3).
func (c *Commune) SetCcoic(v uint64) { c.setNum(cOffCcoic, 3, v) }

// Rfepci returns RFEPCI X(3), the EPCI tax regime ('FPA' or 'FPU').
func (c *Commune) Rfepci() string { return c.get(cOffRfepci, 3) }

// SetRfepci moves v into RFEPCI X(3).
func (c *Commune) SetRfepci(v string) { c.put(cOffRfepci, 3, v) }

// Gtocob returns GTOCOB X, the built EPCI merger/attachment flag
// ('F' = merger, 'R' = attachment, blank = n/a).
func (c *Commune) Gtocob() string { return c.get(cOffGtocob, 1) }

// SetGtocob moves v into GTOCOB X.
func (c *Commune) SetGtocob(v string) { c.put(cOffGtocob, 1, v) }

// Gtocon returns GTOCON X, the non-built EPCI merger/attachment flag
// ('F' = merger, 'R' = attachment, blank = n/a).
func (c *Commune) Gtocon() string { return c.get(cOffGtocon, 1) }

// SetGtocon moves v into GTOCON X.
func (c *Commune) SetGtocon(v string) { c.put(cOffGtocon, 1, v) }

// Gcoetp returns GCOETP 9(3), the TSE public establishment code.
func (c *Commune) Gcoetp() (uint64, bool) { return c.num(cOffGcoetp, 3) }

// GcoetpRaw returns the raw 3 bytes of GCOETP.
func (c *Commune) GcoetpRaw() string { return c.get(cOffGcoetp, 3) }

// SetGcoetp moves v into GCOETP 9(3).
func (c *Commune) SetGcoetp(v uint64) { c.setNum(cOffGcoetp, 3, v) }

// ---- LISTSY OCCURS 10 ----

// SyndicateEntry is a view over one 31-byte LISTSY occurrence. It aliases
// the Commune storage it was obtained from, so setters write through.
type SyndicateEntry struct {
	c    *Commune
	base int
}

// Syndicate returns a view over the LISTSY occurrence i (0-based, COBOL
// subscript i+1). It fails when i is outside [0, SyndicateCount).
func (c *Commune) Syndicate(i int) (SyndicateEntry, error) {
	if err := checkIndex("LISTSY", i, SyndicateCount); err != nil {
		return SyndicateEntry{}, err
	}
	return SyndicateEntry{c: c, base: cOffListsy + i*syndEntryLen}, nil
}

// SirenSyn returns SIREN-SYN X(9), the syndicate SIREN.
func (s SyndicateEntry) SirenSyn() string { return s.c.get(s.base+syndOffSirenSyn, syndSirenLen) }

// SetSirenSyn moves v into SIREN-SYN X(9).
func (s SyndicateEntry) SetSirenSyn(v string) { s.c.put(s.base+syndOffSirenSyn, syndSirenLen, v) }

// Dlisyn returns DLISYN X(20), the syndicate label.
func (s SyndicateEntry) Dlisyn() string { return s.c.get(s.base+syndOffDlisyn, syndDlisynLen) }

// SetDlisyn moves v into DLISYN X(20).
func (s SyndicateEntry) SetDlisyn(v string) { s.c.put(s.base+syndOffDlisyn, syndDlisynLen, v) }

// ---- AGRJA (young farmers deliberations) / AGRJAR REDEFINES ----

// AgrjaEntry is a view over one 9-byte AGRJAO occurrence (AGRJAR
// REDEFINES AGRJA). Occurrence 0 is year N-4 (JGCIN5/JGCRI5) and
// occurrence 4 is year N (JGCIN1/JGCRI1). It aliases the Commune storage.
type AgrjaEntry struct {
	c    *Commune
	base int
}

// Agrja returns a view over the AGRJAO occurrence i (0-based, COBOL
// subscript i+1). It fails when i is outside [0, AgrjaCount).
func (c *Commune) Agrja(i int) (AgrjaEntry, error) {
	if err := checkIndex("AGRJAO", i, AgrjaCount); err != nil {
		return AgrjaEntry{}, err
	}
	return AgrjaEntry{c: c, base: cOffAgrja + i*agrjaEntryLen}, nil
}

// agrjaYear returns the AGRJAO view for the named field suffix n (1 = year
// N, 5 = year N-4): JGCINn / JGCRIn live in occurrence 5-n (0-based).
func (c *Commune) agrjaYear(n int) AgrjaEntry {
	return AgrjaEntry{c: c, base: cOffAgrja + (AgrjaCount-n)*agrjaEntryLen}
}

// Grinsn returns GRINSN 9(4), the installation year.
func (a AgrjaEntry) Grinsn() (uint64, bool) { return a.c.num(a.base+agrjaOffGrinsn, agrjaYearLen) }

// GrinsnRaw returns the raw 4 bytes of GRINSN.
func (a AgrjaEntry) GrinsnRaw() string { return a.c.get(a.base+agrjaOffGrinsn, agrjaYearLen) }

// SetGrinsn moves v into GRINSN 9(4).
func (a AgrjaEntry) SetGrinsn(v uint64) { a.c.setNum(a.base+agrjaOffGrinsn, agrjaYearLen, v) }

// Grrimn returns GRRIMN 9(4), the year of return to taxation.
func (a AgrjaEntry) Grrimn() (uint64, bool) { return a.c.num(a.base+agrjaOffGrrimn, agrjaYearLen) }

// GrrimnRaw returns the raw 4 bytes of GRRIMN.
func (a AgrjaEntry) GrrimnRaw() string { return a.c.get(a.base+agrjaOffGrrimn, agrjaYearLen) }

// SetGrrimn moves v into GRRIMN 9(4).
func (a AgrjaEntry) SetGrrimn(v uint64) { a.c.setNum(a.base+agrjaOffGrrimn, agrjaYearLen, v) }

// Jgcin1 returns JGCIN1 9(4), the installation year for year N.
func (c *Commune) Jgcin1() (uint64, bool) { return c.agrjaYear(1).Grinsn() }

// Jgcri1 returns JGCRI1 9(4), the return-to-taxation year for year N.
func (c *Commune) Jgcri1() (uint64, bool) { return c.agrjaYear(1).Grrimn() }

// Jgcin2 returns JGCIN2 9(4), the installation year for year N-1.
func (c *Commune) Jgcin2() (uint64, bool) { return c.agrjaYear(2).Grinsn() }

// Jgcri2 returns JGCRI2 9(4), the return-to-taxation year for year N-1.
func (c *Commune) Jgcri2() (uint64, bool) { return c.agrjaYear(2).Grrimn() }

// Jgcin3 returns JGCIN3 9(4), the installation year for year N-2.
func (c *Commune) Jgcin3() (uint64, bool) { return c.agrjaYear(3).Grinsn() }

// Jgcri3 returns JGCRI3 9(4), the return-to-taxation year for year N-2.
func (c *Commune) Jgcri3() (uint64, bool) { return c.agrjaYear(3).Grrimn() }

// Jgcin4 returns JGCIN4 9(4), the installation year for year N-3.
func (c *Commune) Jgcin4() (uint64, bool) { return c.agrjaYear(4).Grinsn() }

// Jgcri4 returns JGCRI4 9(4), the return-to-taxation year for year N-3.
func (c *Commune) Jgcri4() (uint64, bool) { return c.agrjaYear(4).Grrimn() }

// Jgcin5 returns JGCIN5 9(4), the installation year for year N-4.
func (c *Commune) Jgcin5() (uint64, bool) { return c.agrjaYear(5).Grinsn() }

// Jgcri5 returns JGCRI5 9(4), the return-to-taxation year for year N-4.
func (c *Commune) Jgcri5() (uint64, bool) { return c.agrjaYear(5).Grrimn() }

// Jgcin returns JGCINn 9(4) for n in 1..5 (1 = year N, 5 = year N-4). It
// fails when n is outside that range.
func (c *Commune) Jgcin(n int) (uint64, bool, error) {
	if err := checkIndex("JGCIN suffix-1", n-1, AgrjaCount); err != nil {
		return 0, false, err
	}
	v, ok := c.agrjaYear(n).Grinsn()
	return v, ok, nil
}

// SetJgcin moves v into JGCINn 9(4) for n in 1..5.
func (c *Commune) SetJgcin(n int, v uint64) error {
	if err := checkIndex("JGCIN suffix-1", n-1, AgrjaCount); err != nil {
		return err
	}
	c.agrjaYear(n).SetGrinsn(v)
	return nil
}

// Jgcri returns JGCRIn 9(4) for n in 1..5 (1 = year N, 5 = year N-4). It
// fails when n is outside that range.
func (c *Commune) Jgcri(n int) (uint64, bool, error) {
	if err := checkIndex("JGCRI suffix-1", n-1, AgrjaCount); err != nil {
		return 0, false, err
	}
	v, ok := c.agrjaYear(n).Grrimn()
	return v, ok, nil
}

// SetJgcri moves v into JGCRIn 9(4) for n in 1..5.
func (c *Commune) SetJgcri(n int, v uint64) error {
	if err := checkIndex("JGCRI suffix-1", n-1, AgrjaCount); err != nil {
		return err
	}
	c.agrjaYear(n).SetGrrimn(v)
	return nil
}

// ---- e-mail ----

// Vmailg returns VMAILG X(50), the left part of the commune e-mail.
func (c *Commune) Vmailg() string { return c.get(cOffVmailg, cEmailLen) }

// SetVmailg moves v into VMAILG X(50).
func (c *Commune) SetVmailg(v string) { c.put(cOffVmailg, cEmailLen, v) }

// Vmaild returns VMAILD X(50), the right part of the commune e-mail.
func (c *Commune) Vmaild() string { return c.get(cOffVmaild, cEmailLen) }

// SetVmaild moves v into VMAILD X(50).
func (c *Commune) SetVmaild(v string) { c.put(cOffVmaild, cEmailLen, v) }

// ---- neutralisation coefficients PIC 9(3)V9(6), scaled by RateScale ----

// Cncom returns CNCOM 9(3)V9(6), the commune (TASA) neutralisation
// coefficient, scaled by RateScale (e.g. 1.000000 -> 1000000).
func (c *Commune) Cncom() (uint64, bool) { return c.num(cOffCncom, cCoefLen) }

// SetCncom moves a RateScale-scaled value into CNCOM.
func (c *Commune) SetCncom(v uint64) { c.setNum(cOffCncom, cCoefLen, v) }

// Cnepci returns CNEPCI 9(3)V9(6), the EPCI neutralisation coefficient,
// scaled by RateScale.
func (c *Commune) Cnepci() (uint64, bool) { return c.num(cOffCnepci, cCoefLen) }

// SetCnepci moves a RateScale-scaled value into CNEPCI.
func (c *Commune) SetCnepci(v uint64) { c.setNum(cOffCnepci, cCoefLen, v) }

// Cndep returns CNDEP 9(3)V9(6), the département neutralisation
// coefficient, scaled by RateScale.
func (c *Commune) Cndep() (uint64, bool) { return c.num(cOffCndep, cCoefLen) }

// SetCndep moves a RateScale-scaled value into CNDEP.
func (c *Commune) SetCndep(v uint64) { c.setNum(cOffCndep, cCoefLen, v) }

// Cnteom returns CNTEOM 9(3)V9(6), the TEOM neutralisation coefficient,
// scaled by RateScale.
func (c *Commune) Cnteom() (uint64, bool) { return c.num(cOffCnteom, cCoefLen) }

// SetCnteom moves a RateScale-scaled value into CNTEOM.
func (c *Commune) SetCnteom(v uint64) { c.setNum(cOffCnteom, cCoefLen, v) }

// Cntse returns CNTSE 9(3)V9(6), the TSE / GEMAPI neutralisation
// coefficient, scaled by RateScale.
func (c *Commune) Cntse() (uint64, bool) { return c.num(cOffCntse, cCoefLen) }

// SetCntse moves a RateScale-scaled value into CNTSE.
func (c *Commune) SetCntse(v uint64) { c.setNum(cOffCntse, cCoefLen, v) }

// Libre returns the trailing free FILLER X(2269).
func (c *Commune) Libre() string { return c.get(cOffLibre, cLibreLen) }
