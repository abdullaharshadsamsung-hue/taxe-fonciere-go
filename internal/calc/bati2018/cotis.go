// Package bati2018 holds the built-property (taxe foncière bâti) 2018
// calculator and its working records.
package bati2018

import (
	"errors"
	"fmt"
	"strings"
)

// CotisLen is the length in bytes of the XCOTB record ("FICHIER BASES ET
// COTISATIONS BATI").
const CotisLen = 895

// OCCURS counts of the XCOTB record.
const (
	// CotisTsenCount is the OCCURS count of TAUTSEN, BBTSEN and COTITSEN.
	CotisTsenCount = 2
	// CotisOmCount is the OCCURS count of TAUOM1, TAB-OM and COTIS-OM.
	CotisOmCount = 6
)

// Group lengths of the XCOTB record.
const (
	// CotisCleLen is the length of the CLE group (JANIPT, ACODIR, CCOCOM, DLETSR).
	CotisCleLen = 11
	// CotisAnuproLen is the length of the ANUPRO group (CGROUP, DNUMCP).
	CotisAnuproLen = 6
)

// Implied-decimal scales (number of digits after the PIC V) of the
// fixed-point numeric fields. Values are held as int64 = real value * 10^scale.
const (
	// TauxScale is the scale of every rate field PIC S9(4)V9(6)
	// (TAUCOM ... TAUTSEN, PTBTST, TAUOM1).
	TauxScale = 6
	// TauomiScale is the scale of TAUOMI PIC 9(12)V9(6).
	TauomiScale = 6
	// BbteiomScale is the scale of BBTEIOM PIC 9V9(15).
	BbteiomScale = 15
)

// Digit widths (total digits, integer + decimal) of the numeric pictures.
const (
	tauDigits     = 10 // S9(4)V9(6)
	tauomiDigits  = 18 // 9(12)V9(6)
	amtDigits     = 10 // S9(10)
	bbteiomDigits = 16 // 9V9(15)
)

// Named FILLER lengths.
const (
	filler1Len = 1
	filler2Len = 2
	filler3Len = 9
	fillerLen  = 2
	filler5Len = 114
	filler7Len = 124
)

// Layout of the 895-byte XCOTB record (0-based offsets, all USAGE DISPLAY):
//
//	CLE        0..10   JANIPT 9(4) 0..3, ACODIR 4..6 (CCODEP X(2) 4..5,
//	                   CCODIR X 6), CCOCOM 9(3) 7..9, DLETSR X 10
//	FILLER1   11
//	ANUPRO    12..17   CGROUP X 12, DNUMCP 9(5) 13..17
//	FILLER2   18..19
//	ACODB     20       9
//	FILLER3   21..29
//	IMPOT     30       X
//	GNEXPL    31..32   X(2)
//	TOT-TAUX  33..132  TAUCOM 33, TAUDEP 43, TAUREG 53, TAUSYN 63, TAUCU 73,
//	                   TAUTAS 83, TAUGEM 93, TAUTSE 103, TAUTSEN(2) 113..132
//	PTBTST   133..142  S9(4)V9(6)
//	TAUOM1   143..202  S9(4)V9(6) OCCURS 6
//	TAUOMI   203..220  9(12)V9(6)
//	FILLER   221..222
//	BBCOM 223, BBDEP 233, BBREG 243, BBSYN 253, BBCU 263, BBTAS 273,
//	BBGEM 283, BBTSE 293, BBTSEN(2) 303..322              (S9(10))
//	TAB-OM   323..394  OCCURS 6 x (GTAUOM X(2) +0, BBTEOM S9(10) +2)
//	BBTEIOM  395..410  9V9(15)
//	FILLER5  411..524
//	COTICOM 525, COTIDEP 535, COTIREG 545, COTISYN 555, COTICU 565,
//	MCOTA3 575, MCOGE3 585, COTITSE 595, COTITSEN(2) 605..624,
//	MCBTSA 625                                            (S9(10))
//	FILLER7  635..758
//	COTIS-OMI 759..768 S9(10)
//	COTIS-OM 769..828  S9(10) OCCURS 6
//	FA300 829, FN300 839, FA800 849, FN800 859, FA900 869, FN900 879 (S9(10))
//	CCOIFP   889..891  X(3)
//	CCPPER   892..894  X(3)

// OmZone is one occurrence of the TAB-OM group: a household-waste tax
// (TEOM) zone code and its taxable base.
type OmZone struct {
	// Gtauom is GTAUOM X(2), the TEOM zone/rate code ("P ", " P", "RA".."RE"
	// or spaces).
	Gtauom string
	// Bbteom is BBTEOM S9(10), the TEOM base.
	Bbteom int64
}

// Cotis is the XCOTB working record (COTISB in EFITA3B8): rates, taxable
// bases and assessed amounts per local authority for one built-property
// account, plus TEOM zones and management / non-collection fees.
//
// MIGRATION_NOTE: the 'X' placeholder of COPY XCOTB REPLACING 'X' BY COTISB
// disappears; each inclusion is a distinct Cotis value. The 'X'-FILLERn
// items become real data names after substitution, so they are kept as
// fields and round-trip byte-for-byte.
//
// MIGRATION_NOTE: the CLE and ANUPRO groups are kept as raw fixed-width
// strings because EFITA3B8 fills them with group MOVEs (COMBAT-AIDFIC and
// COMBAT-ANUPRO) and copies them back as groups to RETOURB; the elementary
// items (JANIPT, CCODEP, CCODIR, CCOCOM, DLETSR, CGROUP, DNUMCP) are exposed
// through accessor methods. Numeric subfields report the COBOL NUMERIC test.
//
// MIGRATION_NOTE: numeric fields are int64. Implied-decimal pictures hold the
// value scaled by 10^scale (see TauxScale, TauomiScale, BbteiomScale).
// Values wider than their picture are truncated on the high-order side when
// the record is encoded (COBOL MOVE semantics); unsigned pictures (TAUOMI,
// BBTEIOM, ACODB) drop the sign. Use Normalize to apply the same truncation
// in memory.
type Cotis struct {
	// Cle is the raw 11-byte CLE group.
	Cle string
	// Filler1 is 'X'-FILLER1 X.
	Filler1 string
	// Anupro is the raw 6-byte ANUPRO group (owner account).
	Anupro string
	// Filler2 is 'X'-FILLER2 X(2).
	Filler2 string
	// Acodb is ACODB 9, the article type code.
	Acodb int64
	// Filler3 is 'X'-FILLER3 X(9).
	Filler3 string
	// Impot is IMPOT X, the assessment indicator.
	Impot string
	// Gnexpl is GNEXPL X(2).
	Gnexpl string

	// Taucom is TAUCOM S9(4)V9(6), the commune rate (scale TauxScale).
	Taucom int64
	// Taudep is TAUDEP S9(4)V9(6), the département rate.
	Taudep int64
	// Taureg is TAUREG S9(4)V9(6), the region rate.
	Taureg int64
	// Tausyn is TAUSYN S9(4)V9(6), the syndicate rate.
	Tausyn int64
	// Taucu is TAUCU S9(4)V9(6), the EPCI rate.
	Taucu int64
	// Tautas is TAUTAS S9(4)V9(6), the TASA rate.
	Tautas int64
	// Taugem is TAUGEM S9(4)V9(6), the GEMAPI rate.
	Taugem int64
	// Tautse is TAUTSE S9(4)V9(6), the special equipment tax rate.
	Tautse int64
	// Tautsen is TAUTSEN S9(4)V9(6) OCCURS 2, the TSE rates.
	Tautsen [CotisTsenCount]int64
	// Ptbtst is PTBTST S9(4)V9(6).
	Ptbtst int64
	// Tauom1 is TAUOM1 S9(4)V9(6) OCCURS 6, the TEOM rates (full rate, then
	// reduced rates A..E).
	Tauom1 [CotisOmCount]int64
	// Tauomi is TAUOMI 9(12)V9(6) (scale TauomiScale).
	Tauomi int64
	// Filler is 'X'-FILLER X(2).
	Filler string

	// Bbcom is BBCOM S9(10), the commune base.
	Bbcom int64
	// Bbdep is BBDEP S9(10), the département base.
	Bbdep int64
	// Bbreg is BBREG S9(10), the region base.
	Bbreg int64
	// Bbsyn is BBSYN S9(10), the syndicate base.
	Bbsyn int64
	// Bbcu is BBCU S9(10), the EPCI base.
	Bbcu int64
	// Bbtas is BBTAS S9(10), the TASA base.
	Bbtas int64
	// Bbgem is BBGEM S9(10), the GEMAPI base.
	Bbgem int64
	// Bbtse is BBTSE S9(10), the TSE base.
	Bbtse int64
	// Bbtsen is BBTSEN S9(10) OCCURS 2, the TSE bases.
	Bbtsen [CotisTsenCount]int64
	// TabOm is TAB-OM OCCURS 6, the TEOM zones and bases.
	TabOm [CotisOmCount]OmZone
	// Bbteiom is BBTEIOM 9V9(15) (scale BbteiomScale).
	Bbteiom int64
	// Filler5 is 'X'-FILLER5 X(114).
	Filler5 string

	// Coticom is COTICOM S9(10), the commune contribution.
	Coticom int64
	// Cotidep is COTIDEP S9(10), the département contribution.
	Cotidep int64
	// Cotireg is COTIREG S9(10), the region contribution.
	Cotireg int64
	// Cotisyn is COTISYN S9(10), the syndicate contribution.
	Cotisyn int64
	// Coticu is COTICU S9(10), the EPCI contribution.
	Coticu int64
	// Mcota3 is MCOTA3 S9(10), the TASA contribution.
	Mcota3 int64
	// Mcoge3 is MCOGE3 S9(10), the GEMAPI contribution.
	Mcoge3 int64
	// Cotitse is COTITSE S9(10), the TSE contribution.
	Cotitse int64
	// Cotitsen is COTITSEN S9(10) OCCURS 2, the TSE contributions.
	Cotitsen [CotisTsenCount]int64
	// Mcbtsa is MCBTSA S9(10), the amalgamated TSE 1 + TSE 2 + TASA amount.
	Mcbtsa int64
	// Filler7 is 'X'-FILLER7 X(124).
	Filler7 string
	// CotisOmi is COTIS-OMI S9(10), the TEOMI amount.
	CotisOmi int64
	// CotisOm is COTIS-OM S9(10) OCCURS 6, the TEOM contributions.
	CotisOm [CotisOmCount]int64

	// Fa300 is FA300 S9(10), assessment fees at 3%.
	Fa300 int64
	// Fn300 is FN300 S9(10), relief / non-collection fees at 3%.
	Fn300 int64
	// Fa800 is FA800 S9(10), assessment fees at 8%.
	Fa800 int64
	// Fn800 is FN800 S9(10), relief / non-collection fees at 8%.
	Fn800 int64
	// Fa900 is FA900 S9(10), assessment fees at 9%.
	Fa900 int64
	// Fn900 is FN900 S9(10), relief / non-collection fees at 9%.
	Fn900 int64

	// Ccoifp is CCOIFP X(3), the absorbed-commune (IFP) code.
	Ccoifp string
	// Ccpper is CCPPER X(3), the treasury code.
	Ccpper string
}

// NewCotis returns the record in the state produced by INITIALIZE COTISB:
// every alphanumeric item (including the named FILLERs) is spaces and every
// numeric item is zero.
//
// The zero value Cotis{} encodes the same, except that the numeric items of
// the CLE and ANUPRO groups (JANIPT, CCOCOM, DNUMCP) are spaces, as after a
// MOVE SPACES to those groups.
func NewCotis() Cotis {
	return Cotis{
		Cle:     "0000   000 ",
		Filler1: spaces(filler1Len),
		Anupro:  " 00000",
		Filler2: spaces(filler2Len),
		Filler3: spaces(filler3Len),
		Impot:   " ",
		Gnexpl:  "  ",
		Filler:  spaces(fillerLen),
		TabOm: [CotisOmCount]OmZone{
			{Gtauom: "  "}, {Gtauom: "  "}, {Gtauom: "  "},
			{Gtauom: "  "}, {Gtauom: "  "}, {Gtauom: "  "},
		},
		Filler5: spaces(filler5Len),
		Filler7: spaces(filler7Len),
		Ccoifp:  "   ",
		Ccpper:  "   ",
	}
}

// ---- CLE / ANUPRO elementary items ----

// Janipt returns JANIPT 9(4), the tax year; ok is the NUMERIC test.
func (c *Cotis) Janipt() (int64, bool) { return digitsValue(sub(c.Cle, CotisCleLen, 0, 4)) }

// SetJanipt moves v into JANIPT 9(4) (sign dropped, high-order truncation).
func (c *Cotis) SetJanipt(v int64) { c.Cle = replace(c.Cle, CotisCleLen, 0, formatUnsigned(v, 4)) }

// Acodir returns ACODIR X(3) (CCODEP + CCODIR).
func (c *Cotis) Acodir() string { return sub(c.Cle, CotisCleLen, 4, 3) }

// SetAcodir moves v into ACODIR X(3).
func (c *Cotis) SetAcodir(v string) { c.Cle = replace(c.Cle, CotisCleLen, 4, fit(v, 3)) }

// Ccodep returns CCODEP X(2), the département code.
func (c *Cotis) Ccodep() string { return sub(c.Cle, CotisCleLen, 4, 2) }

// SetCcodep moves v into CCODEP X(2).
func (c *Cotis) SetCcodep(v string) { c.Cle = replace(c.Cle, CotisCleLen, 4, fit(v, 2)) }

// Ccodir returns CCODIR X, the direction code.
func (c *Cotis) Ccodir() string { return sub(c.Cle, CotisCleLen, 6, 1) }

// SetCcodir moves v into CCODIR X.
func (c *Cotis) SetCcodir(v string) { c.Cle = replace(c.Cle, CotisCleLen, 6, fit(v, 1)) }

// CcocomRaw returns the raw 3 bytes of CCOCOM 9(3).
func (c *Cotis) CcocomRaw() string { return sub(c.Cle, CotisCleLen, 7, 3) }

// Ccocom returns CCOCOM 9(3), the commune code; ok is the NUMERIC test.
func (c *Cotis) Ccocom() (int64, bool) { return digitsValue(c.CcocomRaw()) }

// SetCcocom moves v into CCOCOM 9(3) (sign dropped, high-order truncation).
func (c *Cotis) SetCcocom(v int64) { c.Cle = replace(c.Cle, CotisCleLen, 7, formatUnsigned(v, 3)) }

// Dletsr returns DLETSR X, the roll series letter.
func (c *Cotis) Dletsr() string { return sub(c.Cle, CotisCleLen, 10, 1) }

// SetDletsr moves v into DLETSR X.
func (c *Cotis) SetDletsr(v string) { c.Cle = replace(c.Cle, CotisCleLen, 10, fit(v, 1)) }

// Cgroup returns CGROUP X, the owner group code.
func (c *Cotis) Cgroup() string { return sub(c.Anupro, CotisAnuproLen, 0, 1) }

// SetCgroup moves v into CGROUP X.
func (c *Cotis) SetCgroup(v string) { c.Anupro = replace(c.Anupro, CotisAnuproLen, 0, fit(v, 1)) }

// DnumcpRaw returns the raw 5 bytes of DNUMCP 9(5).
func (c *Cotis) DnumcpRaw() string { return sub(c.Anupro, CotisAnuproLen, 1, 5) }

// Dnumcp returns DNUMCP 9(5), the owner account number; ok is NUMERIC.
func (c *Cotis) Dnumcp() (int64, bool) { return digitsValue(c.DnumcpRaw()) }

// SetDnumcp moves v into DNUMCP 9(5) (sign dropped, high-order truncation).
func (c *Cotis) SetDnumcp(v int64) {
	c.Anupro = replace(c.Anupro, CotisAnuproLen, 1, formatUnsigned(v, 5))
}

// CotisOmSubscript returns COTIS-OM(ind) for a 1-based COBOL subscript,
// resolved the way an unchecked subscript is resolved in storage.
//
// MIGRATION_NOTE: EFITA3B8 sums COTIS-OM with PERFORM VARYING IND FROM 1 BY
// 1 UNTIL IND > 7, i.e. it reads one occurrence past the table. Without
// SSRANGE, COTIS-OM(7) addresses the next 10 bytes of the record, which are
// FA300 S9(10). This helper reproduces that quirk (ind 7 → Fa300); any other
// subscript outside 1..6 is reported as an error.
func (c *Cotis) CotisOmSubscript(ind int) (int64, error) {
	switch {
	case ind >= 1 && ind <= CotisOmCount:
		return c.CotisOm[ind-1], nil
	case ind == CotisOmCount+1:
		return c.Fa300, nil
	default:
		return 0, fmt.Errorf("bati2018: COTISB COTIS-OM subscript %d out of storage range [1,%d]", ind, CotisOmCount+1)
	}
}

// Normalize applies to every numeric field the truncation a COBOL MOVE into
// its picture would perform (high-order digits dropped, sign dropped for
// unsigned pictures, negative zero becomes zero), so in-memory values equal
// what Encode stores.
func (c *Cotis) Normalize() {
	c.Acodb = truncUnsigned(c.Acodb, 1)
	for _, p := range c.tauFields() {
		*p = truncSigned(*p, tauDigits)
	}
	c.Tauomi = truncUnsigned(c.Tauomi, tauomiDigits)
	for _, p := range c.amtFields() {
		*p = truncSigned(*p, amtDigits)
	}
	c.Bbteiom = truncUnsigned(c.Bbteiom, bbteiomDigits)
}

func (c *Cotis) tauFields() []*int64 {
	ps := []*int64{&c.Taucom, &c.Taudep, &c.Taureg, &c.Tausyn, &c.Taucu,
		&c.Tautas, &c.Taugem, &c.Tautse, &c.Tautsen[0], &c.Tautsen[1], &c.Ptbtst}
	for i := range c.Tauom1 {
		ps = append(ps, &c.Tauom1[i])
	}
	return ps
}

func (c *Cotis) amtFields() []*int64 {
	ps := []*int64{&c.Bbcom, &c.Bbdep, &c.Bbreg, &c.Bbsyn, &c.Bbcu, &c.Bbtas,
		&c.Bbgem, &c.Bbtse, &c.Bbtsen[0], &c.Bbtsen[1]}
	for i := range c.TabOm {
		ps = append(ps, &c.TabOm[i].Bbteom)
	}
	ps = append(ps, &c.Coticom, &c.Cotidep, &c.Cotireg, &c.Cotisyn, &c.Coticu,
		&c.Mcota3, &c.Mcoge3, &c.Cotitse, &c.Cotitsen[0], &c.Cotitsen[1],
		&c.Mcbtsa, &c.CotisOmi)
	for i := range c.CotisOm {
		ps = append(ps, &c.CotisOm[i])
	}
	return append(ps, &c.Fa300, &c.Fn300, &c.Fa800, &c.Fn800, &c.Fa900, &c.Fn900)
}

// ---- codec ----

// ParseCotis decodes an 895-byte XCOTB record. It fails if b is not exactly
// CotisLen bytes or if a numeric item fails the COBOL NUMERIC class test
// (the error names the item and its 1-based position).
//
// Signed items accept both IBM ('{', 'A'-'I', '}', 'J'-'R') and ASCII
// ('p'-'y') trailing overpunch signs as well as a plain digit (positive).
func ParseCotis(b []byte) (Cotis, error) {
	var c Cotis
	if len(b) != CotisLen {
		return c, fmt.Errorf("bati2018: COTISB (XCOTB) must be %d bytes, got %d", CotisLen, len(b))
	}
	d := &decoder{b: b}

	c.Cle = d.alpha(CotisCleLen)
	c.Filler1 = d.alpha(filler1Len)
	c.Anupro = d.alpha(CotisAnuproLen)
	c.Filler2 = d.alpha(filler2Len)
	c.Acodb = d.unsigned("ACODB", 1)
	c.Filler3 = d.alpha(filler3Len)
	c.Impot = d.alpha(1)
	c.Gnexpl = d.alpha(2)

	c.Taucom = d.signed("TAUCOM", tauDigits)
	c.Taudep = d.signed("TAUDEP", tauDigits)
	c.Taureg = d.signed("TAUREG", tauDigits)
	c.Tausyn = d.signed("TAUSYN", tauDigits)
	c.Taucu = d.signed("TAUCU", tauDigits)
	c.Tautas = d.signed("TAUTAS", tauDigits)
	c.Taugem = d.signed("TAUGEM", tauDigits)
	c.Tautse = d.signed("TAUTSE", tauDigits)
	for i := range c.Tautsen {
		c.Tautsen[i] = d.signed(fmt.Sprintf("TAUTSEN(%d)", i+1), tauDigits)
	}
	c.Ptbtst = d.signed("PTBTST", tauDigits)
	for i := range c.Tauom1 {
		c.Tauom1[i] = d.signed(fmt.Sprintf("TAUOM1(%d)", i+1), tauDigits)
	}
	c.Tauomi = d.unsigned("TAUOMI", tauomiDigits)
	c.Filler = d.alpha(fillerLen)

	c.Bbcom = d.signed("BBCOM", amtDigits)
	c.Bbdep = d.signed("BBDEP", amtDigits)
	c.Bbreg = d.signed("BBREG", amtDigits)
	c.Bbsyn = d.signed("BBSYN", amtDigits)
	c.Bbcu = d.signed("BBCU", amtDigits)
	c.Bbtas = d.signed("BBTAS", amtDigits)
	c.Bbgem = d.signed("BBGEM", amtDigits)
	c.Bbtse = d.signed("BBTSE", amtDigits)
	for i := range c.Bbtsen {
		c.Bbtsen[i] = d.signed(fmt.Sprintf("BBTSEN(%d)", i+1), amtDigits)
	}
	for i := range c.TabOm {
		c.TabOm[i].Gtauom = d.alpha(2)
		c.TabOm[i].Bbteom = d.signed(fmt.Sprintf("BBTEOM(%d)", i+1), amtDigits)
	}
	c.Bbteiom = d.unsigned("BBTEIOM", bbteiomDigits)
	c.Filler5 = d.alpha(filler5Len)

	c.Coticom = d.signed("COTICOM", amtDigits)
	c.Cotidep = d.signed("COTIDEP", amtDigits)
	c.Cotireg = d.signed("COTIREG", amtDigits)
	c.Cotisyn = d.signed("COTISYN", amtDigits)
	c.Coticu = d.signed("COTICU", amtDigits)
	c.Mcota3 = d.signed("MCOTA3", amtDigits)
	c.Mcoge3 = d.signed("MCOGE3", amtDigits)
	c.Cotitse = d.signed("COTITSE", amtDigits)
	for i := range c.Cotitsen {
		c.Cotitsen[i] = d.signed(fmt.Sprintf("COTITSEN(%d)", i+1), amtDigits)
	}
	c.Mcbtsa = d.signed("MCBTSA", amtDigits)
	c.Filler7 = d.alpha(filler7Len)
	c.CotisOmi = d.signed("COTIS-OMI", amtDigits)
	for i := range c.CotisOm {
		c.CotisOm[i] = d.signed(fmt.Sprintf("COTIS-OM(%d)", i+1), amtDigits)
	}

	c.Fa300 = d.signed("FA300", amtDigits)
	c.Fn300 = d.signed("FN300", amtDigits)
	c.Fa800 = d.signed("FA800", amtDigits)
	c.Fn800 = d.signed("FN800", amtDigits)
	c.Fa900 = d.signed("FA900", amtDigits)
	c.Fn900 = d.signed("FN900", amtDigits)

	c.Ccoifp = d.alpha(3)
	c.Ccpper = d.alpha(3)

	if d.err != nil {
		return Cotis{}, d.err
	}
	if d.pos != CotisLen {
		return Cotis{}, fmt.Errorf("bati2018: COTISB layout consumed %d bytes, want %d", d.pos, CotisLen)
	}
	return c, nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler via ParseCotis.
func (c *Cotis) UnmarshalBinary(b []byte) error {
	v, err := ParseCotis(b)
	if err != nil {
		return err
	}
	*c = v
	return nil
}

// Encode returns the 895-byte XCOTB image of the record. Alphanumeric items
// are left-justified, space-padded and right-truncated; numeric items are
// zoned decimal with high-order truncation, and signed items carry an IBM
// trailing overpunch sign (negative zero is stored as +0).
func (c *Cotis) Encode() []byte {
	e := &encoder{buf: make([]byte, 0, CotisLen)}

	e.alpha(c.Cle, CotisCleLen)
	e.alpha(c.Filler1, filler1Len)
	e.alpha(c.Anupro, CotisAnuproLen)
	e.alpha(c.Filler2, filler2Len)
	e.unsigned(c.Acodb, 1)
	e.alpha(c.Filler3, filler3Len)
	e.alpha(c.Impot, 1)
	e.alpha(c.Gnexpl, 2)

	for _, v := range []int64{c.Taucom, c.Taudep, c.Taureg, c.Tausyn, c.Taucu,
		c.Tautas, c.Taugem, c.Tautse, c.Tautsen[0], c.Tautsen[1], c.Ptbtst} {
		e.signed(v, tauDigits)
	}
	for _, v := range c.Tauom1 {
		e.signed(v, tauDigits)
	}
	e.unsigned(c.Tauomi, tauomiDigits)
	e.alpha(c.Filler, fillerLen)

	for _, v := range []int64{c.Bbcom, c.Bbdep, c.Bbreg, c.Bbsyn, c.Bbcu,
		c.Bbtas, c.Bbgem, c.Bbtse, c.Bbtsen[0], c.Bbtsen[1]} {
		e.signed(v, amtDigits)
	}
	for _, z := range c.TabOm {
		e.alpha(z.Gtauom, 2)
		e.signed(z.Bbteom, amtDigits)
	}
	e.unsigned(c.Bbteiom, bbteiomDigits)
	e.alpha(c.Filler5, filler5Len)

	for _, v := range []int64{c.Coticom, c.Cotidep, c.Cotireg, c.Cotisyn,
		c.Coticu, c.Mcota3, c.Mcoge3, c.Cotitse, c.Cotitsen[0], c.Cotitsen[1],
		c.Mcbtsa} {
		e.signed(v, amtDigits)
	}
	e.alpha(c.Filler7, filler7Len)
	e.signed(c.CotisOmi, amtDigits)
	for _, v := range c.CotisOm {
		e.signed(v, amtDigits)
	}
	for _, v := range []int64{c.Fa300, c.Fn300, c.Fa800, c.Fn800, c.Fa900, c.Fn900} {
		e.signed(v, amtDigits)
	}

	e.alpha(c.Ccoifp, 3)
	e.alpha(c.Ccpper, 3)
	return e.buf
}

// MarshalBinary implements encoding.BinaryMarshaler via Encode.
func (c *Cotis) MarshalBinary() ([]byte, error) {
	b := c.Encode()
	if len(b) != CotisLen {
		return nil, fmt.Errorf("bati2018: COTISB encoded to %d bytes, want %d", len(b), CotisLen)
	}
	return b, nil
}

// ---- zoned-decimal helpers ----

var pow10 = [...]uint64{
	1, 10, 100, 1e3, 1e4, 1e5, 1e6, 1e7, 1e8, 1e9, 1e10,
	1e11, 1e12, 1e13, 1e14, 1e15, 1e16, 1e17, 1e18,
}

// errNotNumeric marks a failed NUMERIC class test.
var errNotNumeric = errors.New("not NUMERIC")

func absU(v int64) uint64 {
	if v < 0 {
		return uint64(-(v + 1)) + 1
	}
	return uint64(v)
}

// truncUnsigned is a MOVE into an unsigned PIC 9(n): sign dropped, high-order
// digits truncated.
func truncUnsigned(v int64, n int) int64 { return int64(absU(v) % pow10[n]) }

// truncSigned is a MOVE into PIC S9(n): high-order digits truncated, sign kept.
func truncSigned(v int64, n int) int64 {
	m := int64(absU(v) % pow10[n])
	if v < 0 {
		return -m
	}
	return m
}

func formatUnsigned(v int64, n int) string {
	u := absU(v) % pow10[n]
	b := make([]byte, n)
	for i := n - 1; i >= 0; i-- {
		b[i] = byte('0' + u%10)
		u /= 10
	}
	return string(b)
}

func digitsValue(s string) (int64, bool) {
	var v int64
	ok := len(s) > 0
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if ch < '0' || ch > '9' {
			ok = false
		}
		v = v*10 + int64(ch&0x0F)
	}
	return v, ok
}

func spaces(n int) string { return strings.Repeat(" ", n) }

// fit performs an alphanumeric MOVE into X(n).
func fit(s string, n int) string {
	if len(s) >= n {
		return s[:n]
	}
	return s + spaces(n-len(s))
}

func sub(group string, groupLen, off, n int) string {
	return fit(group, groupLen)[off : off+n]
}

func replace(group string, groupLen, off int, v string) string {
	g := fit(group, groupLen)
	return g[:off] + v + g[off+len(v):]
}

type decoder struct {
	b   []byte
	pos int
	err error
}

func (d *decoder) take(n int) []byte {
	s := d.b[d.pos : d.pos+n]
	d.pos += n
	return s
}

func (d *decoder) fail(name string, start int, err error) {
	if d.err == nil {
		d.err = fmt.Errorf("bati2018: COTISB-%s at position %d: %w", name, start+1, err)
	}
}

func (d *decoder) alpha(n int) string { return string(d.take(n)) }

func (d *decoder) unsigned(name string, n int) int64 {
	start := d.pos
	v, ok := digitsValue(string(d.take(n)))
	if !ok {
		d.fail(name, start, errNotNumeric)
		return 0
	}
	return v
}

func (d *decoder) signed(name string, n int) int64 {
	start := d.pos
	raw := d.take(n)
	v, ok := digitsValue(string(raw[:n-1]))
	digit, neg, signOK := overpunch(raw[n-1])
	if !ok || !signOK {
		d.fail(name, start, errNotNumeric)
		return 0
	}
	v = v*10 + digit
	if neg {
		return -v
	}
	return v
}

// overpunch decodes a trailing sign byte (plain digit, IBM or ASCII
// overpunch).
func overpunch(b byte) (digit int64, neg, ok bool) {
	switch {
	case b >= '0' && b <= '9':
		return int64(b - '0'), false, true
	case b == '{':
		return 0, false, true
	case b >= 'A' && b <= 'I':
		return int64(b-'A') + 1, false, true
	case b == '}':
		return 0, true, true
	case b >= 'J' && b <= 'R':
		return int64(b-'J') + 1, true, true
	case b >= 'p' && b <= 'y':
		return int64(b - 'p'), true, true
	}
	return 0, false, false
}

type encoder struct{ buf []byte }

func (e *encoder) alpha(s string, n int) { e.buf = append(e.buf, fit(s, n)...) }

func (e *encoder) unsigned(v int64, n int) { e.buf = append(e.buf, formatUnsigned(v, n)...) }

func (e *encoder) signed(v int64, n int) {
	digits := []byte(formatUnsigned(v, n))
	last := digits[n-1] - '0'
	neg := v < 0 && absU(v)%pow10[n] != 0
	switch {
	case neg && last == 0:
		digits[n-1] = '}'
	case neg:
		digits[n-1] = 'J' + last - 1
	case last == 0:
		digits[n-1] = '{'
	default:
		digits[n-1] = 'A' + last - 1
	}
	e.buf = append(e.buf, digits...)
}
