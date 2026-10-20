// Package nonbati2018 holds the non-built-property (taxe foncière non bâti)
// 2018 calculator and its working records.
package nonbati2018

import (
	"errors"
	"fmt"
	"strings"
)

// CotisLen is the length in bytes of the XCOTNB record ("FICHIER BASES ET
// COTISATIONS NON-BATI").
const CotisLen = 1004

// CotisTsenCount is the OCCURS count of TAUTSEN, BNBTSEN and COTITSEN.
const CotisTsenCount = 2

// Postes of the COTISNB table (01 COTISNB. 02 EL-COTISNB OCCURS 5) in
// EFITA3N8. Only postes 1 and 4 are still filled; 2, 3 and 5 are kept by the
// COBOL program for interface compatibility and always stay initialized.
const (
	// PosteCount is the OCCURS count of EL-COTISNB.
	PosteCount = 5
	// PosteImposition is the 1-based poste holding the "I BN" assessment.
	PosteImposition = 1
	// PosteJeunesAgriculteurs is the 1-based poste holding the "G JA"
	// young-farmer relief.
	PosteJeunesAgriculteurs = 4
)

// Group lengths of the XCOTNB record.
const (
	// CotisCleLen is the length of the CLE group (JANIPT, ACODIR, CCOCOM, DLETSR).
	CotisCleLen = 11
	// CotisAnuproLen is the length of the ANUPRO group (CGROUP, DNUMCP).
	CotisAnuproLen = 6
)

// TauxScale is the implied-decimal scale of every rate field PIC S9(4)V9(6)
// (TAUCOM ... TAUTXADD). Rates are held as int64 = real value * 10^6.
const TauxScale = 6

// Digit widths of the numeric pictures.
const (
	acodnbDigits = 1  // 9(1)
	tauDigits    = 10 // S9(4)V9(6)
	amtDigits    = 10 // S9(10)
	dfixDigits   = 5  // S9(5)
	fraisDigits  = 12 // 9(12)
)

// Named FILLER lengths.
const (
	filler1Len = 1
	filler2Len = 2
	filler3Len = 9
	filler4Len = 12
	filler6Len = 10
	fillerLen  = 306
)

// SignMode selects how the trailing overpunched sign of signed zoned-decimal
// items is written by Encode.
type SignMode int

const (
	// SignIBM writes EBCDIC-style overpunch characters ('{', 'A'-'I' for
	// positive, '}', 'J'-'R' for negative), as produced on z/OS.
	SignIBM SignMode = iota
	// SignASCII writes the ASCII (Micro Focus / GnuCOBOL) convention: plain
	// digit for positive, 'p'-'y' for negative.
	SignASCII
)

// SignEncoding is the package-wide sign convention used by Encode. Decoding
// always accepts both conventions. It defaults to SignIBM and should only be
// changed at program start-up.
var SignEncoding = SignIBM

// Layout of the 1004-byte XCOTNB record (0-based offsets, all USAGE DISPLAY):
//
//	CLE        0..10   JANIPT 9(4) 0..3, ACODIR 4..6 (CCODEP X(2) 4..5,
//	                   CCODIR X 6), CCOCOM 9(3) 7..9, DLETSR X 10
//	FILLER1   11
//	ANUPRO    12..17   CGROUP X 12, DNUMCP 9(5) 13..17
//	FILLER2   18..19
//	ACODNB    20       9
//	FILLER3   21..29
//	IMPOT     30       X
//	GNEXPS    31..32   X(2)
//	FILLER4   33..44
//	TAUCOM 45, TAUDEP 55, TAUREG 65, TAUSYN 75, TAUCU 85, TAUGEM 95,
//	TAUTSE 105, TAUTSEN(2) 115..134, PNBTST 135, TAUCHAGR 145,
//	TAUSMAGR 155, TAUBAPSA 165, TAUCAAA 175, TAUTXADD 185   (S9(4)V9(6))
//	BNBCOM 195, BNBDEP 205, BNBREG 215, BNBSYN 225, BNBCU 235, BNBGEM 245,
//	BNBTSE 255, BNBTSEN(2) 265..284, BNBAGR 285, BNBBAP 295,
//	BNBCAA 305                                               (S9(10))
//	FILLER6  315..324
//	BNBPLJA  325..334  S9(10)
//	COTICOM 335, COTIDEP 345, COTIREG 355, COTISYN 365, COTICU 375,
//	MCOGE3 385, COTITSE 395, COTITSEN(2) 405..424, MCNTSA 425,
//	COTIAGR 435, COTISMAG 445, COTIBAP 455, DPROCAA 465      (S9(10))
//	DFIXCAA  475..479  S9(5)
//	FFIXCAA 480, FA300 490, FN300 500, FARB 510, FRNVB 520, FARC 530,
//	FRNVC 540, FA800 550, FN800 560, FA900 570, FN900 580    (S9(10))
//	MBJECO 590, MBJESY 600, MBJEIC 610, MCNJCO 620, MCNJSY 630,
//	MCNJIC 640, MCNJAE 650, MFAJAE 660, MFNJAE 670           (S9(10))
//	FRAIS-TOTAUX 680..691 9(12)
//	CCOIFP   692..694  X(3)
//	CCPPER   695..697  X(3)
//	FILLER   698..1003 X(306)

// Cotis is one poste of the XCOTNB working table (COTISNB in EFITA3N8):
// rates, taxable bases and assessed contributions per local authority for
// one non-built-property article, plus CAAA rights, management fees and the
// young-farmer (JA) relief amounts (filled only on the "G JA" poste).
//
// MIGRATION_NOTE: the 'X' placeholder of COPY XCOTNB REPLACING 'X' BY
// COTISNB disappears; each poste is a distinct Cotis value. The 'X'-FILLERn
// items are real data names after substitution, so they are kept as fields
// and round-trip byte-for-byte.
//
// MIGRATION_NOTE: CLE and ANUPRO are kept as raw fixed-width strings because
// EFITA3N8 fills them with group MOVEs (COMNONBA-AIDFIC, COMNONBA-ANUPRO) and
// copies them back as groups to RETOURNB; the elementary items are exposed
// through accessor methods (CCODEP is bytes 5-6 of AIDFIC). Numeric subfields
// report the COBOL NUMERIC test.
//
// MIGRATION_NOTE: numeric fields are int64. Rates hold the value scaled by
// 10^TauxScale. Values wider than their picture are truncated on the
// high-order side when encoded (COBOL MOVE semantics) — e.g. DFIXCAA is only
// S9(5); unsigned pictures (ACODNB, FRAIS-TOTAUX) drop the sign. Use
// Normalize to apply the same truncation in memory.
type Cotis struct {
	// Cle is the raw 11-byte CLE group.
	Cle string
	// Filler1 is 'X'-FILLER1 X.
	Filler1 string
	// Anupro is the raw 6-byte ANUPRO group (owner account).
	Anupro string
	// Filler2 is 'X'-FILLER2 X(2).
	Filler2 string
	// Acodnb is ACODNB 9, the article type code.
	Acodnb int64
	// Filler3 is 'X'-FILLER3 X(9).
	Filler3 string
	// Impot is IMPOT X, the assessment indicator ('I' or 'G').
	Impot string
	// Gnexps is GNEXPS X(2) ("BN" or "JA").
	Gnexps string
	// Filler4 is 'X'-FILLER4 X(12).
	Filler4 string

	// Taucom is TAUCOM S9(4)V9(6), the commune rate (scale TauxScale).
	Taucom int64
	// Taudep is TAUDEP S9(4)V9(6), the département rate.
	Taudep int64
	// Taureg is TAUREG S9(4)V9(6), the region rate.
	Taureg int64
	// Tausyn is TAUSYN S9(4)V9(6), the communal syndicate rate.
	Tausyn int64
	// Taucu is TAUCU S9(4)V9(6), the EPCI rate.
	Taucu int64
	// Taugem is TAUGEM S9(4)V9(6), the GEMAPI rate.
	Taugem int64
	// Tautse is TAUTSE S9(4)V9(6), the special equipment tax rate.
	Tautse int64
	// Tautsen is TAUTSEN S9(4)V9(6) OCCURS 2, the TSE 1 / TSE 2 rates.
	Tautsen [CotisTsenCount]int64
	// Pnbtst is PNBTST S9(4)V9(6).
	Pnbtst int64
	// Tauchagr is TAUCHAGR S9(4)V9(6), the chamber of agriculture rate.
	Tauchagr int64
	// Tausmagr is TAUSMAGR S9(4)V9(6), the Saint-Martin consular chamber rate.
	Tausmagr int64
	// Taubapsa is TAUBAPSA S9(4)V9(6), the BAPSA rate.
	Taubapsa int64
	// Taucaaa is TAUCAAA S9(4)V9(6), the CAAA proportional-right rate.
	Taucaaa int64
	// Tautxadd is TAUTXADD S9(4)V9(6), the additional tax rate.
	Tautxadd int64

	// Bnbcom is BNBCOM S9(10), the commune base.
	Bnbcom int64
	// Bnbdep is BNBDEP S9(10), the département / additional tax base.
	Bnbdep int64
	// Bnbreg is BNBREG S9(10), the region base.
	Bnbreg int64
	// Bnbsyn is BNBSYN S9(10), the communal syndicate base.
	Bnbsyn int64
	// Bnbcu is BNBCU S9(10), the EPCI base.
	Bnbcu int64
	// Bnbgem is BNBGEM S9(10), the GEMAPI base.
	Bnbgem int64
	// Bnbtse is BNBTSE S9(10), the TSE base.
	Bnbtse int64
	// Bnbtsen is BNBTSEN S9(10) OCCURS 2, the TSE 1 / TSE 2 bases.
	Bnbtsen [CotisTsenCount]int64
	// Bnbagr is BNBAGR S9(10), the chamber of agriculture base.
	Bnbagr int64
	// Bnbbap is BNBBAP S9(10), the BAPSA base.
	Bnbbap int64
	// Bnbcaa is BNBCAA S9(10), the CAAA base.
	Bnbcaa int64
	// Filler6 is 'X'-FILLER6 X(10).
	Filler6 string
	// Bnbplja is BNBPLJA S9(10), the pre + JA relief base ("G JA" only).
	Bnbplja int64

	// Coticom is COTICOM S9(10), the commune contribution.
	Coticom int64
	// Cotidep is COTIDEP S9(10), the département / additional tax contribution.
	Cotidep int64
	// Cotireg is COTIREG S9(10), the region contribution.
	Cotireg int64
	// Cotisyn is COTISYN S9(10), the communal syndicate contribution.
	Cotisyn int64
	// Coticu is COTICU S9(10), the EPCI contribution.
	Coticu int64
	// Mcoge3 is MCOGE3 S9(10), the GEMAPI contribution.
	Mcoge3 int64
	// Cotitse is COTITSE S9(10), the TSE contribution.
	Cotitse int64
	// Cotitsen is COTITSEN S9(10) OCCURS 2, the TSE 1 / TSE 2 contributions.
	Cotitsen [CotisTsenCount]int64
	// Mcntsa is MCNTSA S9(10), the amalgamated TSE 1 + TSE 2 contribution.
	Mcntsa int64
	// Cotiagr is COTIAGR S9(10), the chamber of agriculture contribution.
	Cotiagr int64
	// Cotismag is COTISMAG S9(10), the Saint-Martin consular chamber contribution.
	Cotismag int64
	// Cotibap is COTIBAP S9(10), the BAPSA contribution.
	Cotibap int64
	// Dprocaa is DPROCAA S9(10), the CAAA proportional right.
	Dprocaa int64
	// Dfixcaa is DFIXCAA S9(5), the CAAA fixed right.
	Dfixcaa int64

	// Ffixcaa is FFIXCAA S9(10), fees on the CAAA fixed right.
	Ffixcaa int64
	// Fa300 is FA300 S9(10), assessment fees at 3%.
	Fa300 int64
	// Fn300 is FN300 S9(10), relief / non-collection fees at 3%.
	Fn300 int64
	// Farb is FARB S9(10), BAPSA assessment fees.
	Farb int64
	// Frnvb is FRNVB S9(10), BAPSA relief / non-collection fees.
	Frnvb int64
	// Farc is FARC S9(10), CAAA assessment fees.
	Farc int64
	// Frnvc is FRNVC S9(10), CAAA relief / non-collection fees.
	Frnvc int64
	// Fa800 is FA800 S9(10), assessment fees at 8%.
	Fa800 int64
	// Fn800 is FN800 S9(10), relief / non-collection fees at 8%.
	Fn800 int64
	// Fa900 is FA900 S9(10), assessment fees at 9%.
	Fa900 int64
	// Fn900 is FN900 S9(10), relief / non-collection fees at 9%.
	Fn900 int64

	// Mbjeco is MBJECO S9(10), the JA relief commune base, State share.
	Mbjeco int64
	// Mbjesy is MBJESY S9(10), the JA relief syndicate base, State share.
	Mbjesy int64
	// Mbjeic is MBJEIC S9(10), the JA relief EPCI base, State share.
	Mbjeic int64
	// Mcnjco is MCNJCO S9(10), the JA relief commune amount, State share.
	Mcnjco int64
	// Mcnjsy is MCNJSY S9(10), the JA relief syndicate amount, State share.
	Mcnjsy int64
	// Mcnjic is MCNJIC S9(10), the JA relief EPCI amount, State share.
	Mcnjic int64
	// Mcnjae is MCNJAE S9(10), the amalgamated JA relief, State share.
	Mcnjae int64
	// Mfajae is MFAJAE S9(10), the amalgamated JA assessment fees.
	Mfajae int64
	// Mfnjae is MFNJAE S9(10), the amalgamated JA relief / non-collection fees.
	Mfnjae int64

	// FraisTotaux is FRAIS-TOTAUX 9(12), the total fees.
	FraisTotaux int64

	// Ccoifp is CCOIFP X(3), the absorbed-commune (IFP) code.
	Ccoifp string
	// Ccpper is CCPPER X(3), the treasury code.
	Ccpper string
	// Filler is the trailing 'X'-FILLER X(306).
	Filler string
}

// NewCotis returns the record in the state produced by INITIALIZE COTISNB:
// every alphanumeric item (including the named FILLERs) is spaces and every
// numeric item is zero (stored as positive zero).
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
		Gnexps:  "  ",
		Filler4: spaces(filler4Len),
		Filler6: spaces(filler6Len),
		Ccoifp:  "   ",
		Ccpper:  "   ",
		Filler:  spaces(fillerLen),
	}
}

// NewCotisTable returns the five postes of EL-COTISNB as left by
// INITIALIZE COTISNB. Index i holds COBOL poste i+1.
func NewCotisTable() [PosteCount]Cotis {
	var t [PosteCount]Cotis
	for i := range t {
		t[i] = NewCotis()
	}
	return t
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

// Ccodep returns CCODEP X(2), the département code (bytes 5-6 of AIDFIC).
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

// ---- field layout ----

type fieldKind int

const (
	kindAlpha fieldKind = iota
	kindUnsigned
	kindSigned
)

// field describes one elementary (or raw group) item of the record, in
// storage order. Exactly one of str / num is set.
type field struct {
	name string
	kind fieldKind
	n    int
	str  *string
	num  *int64
}

func alphaF(name string, n int, p *string) field {
	return field{name: name, kind: kindAlpha, n: n, str: p}
}
func unsignedF(name string, n int, p *int64) field {
	return field{name: name, kind: kindUnsigned, n: n, num: p}
}
func signedF(name string, n int, p *int64) field {
	return field{name: name, kind: kindSigned, n: n, num: p}
}

// layout returns the record's items in storage order, bound to c.
func (c *Cotis) layout() []field {
	return []field{
		alphaF("CLE", CotisCleLen, &c.Cle),
		alphaF("FILLER1", filler1Len, &c.Filler1),
		alphaF("ANUPRO", CotisAnuproLen, &c.Anupro),
		alphaF("FILLER2", filler2Len, &c.Filler2),
		unsignedF("ACODNB", acodnbDigits, &c.Acodnb),
		alphaF("FILLER3", filler3Len, &c.Filler3),
		alphaF("IMPOT", 1, &c.Impot),
		alphaF("GNEXPS", 2, &c.Gnexps),
		alphaF("FILLER4", filler4Len, &c.Filler4),

		signedF("TAUCOM", tauDigits, &c.Taucom),
		signedF("TAUDEP", tauDigits, &c.Taudep),
		signedF("TAUREG", tauDigits, &c.Taureg),
		signedF("TAUSYN", tauDigits, &c.Tausyn),
		signedF("TAUCU", tauDigits, &c.Taucu),
		signedF("TAUGEM", tauDigits, &c.Taugem),
		signedF("TAUTSE", tauDigits, &c.Tautse),
		signedF("TAUTSEN(1)", tauDigits, &c.Tautsen[0]),
		signedF("TAUTSEN(2)", tauDigits, &c.Tautsen[1]),
		signedF("PNBTST", tauDigits, &c.Pnbtst),
		signedF("TAUCHAGR", tauDigits, &c.Tauchagr),
		signedF("TAUSMAGR", tauDigits, &c.Tausmagr),
		signedF("TAUBAPSA", tauDigits, &c.Taubapsa),
		signedF("TAUCAAA", tauDigits, &c.Taucaaa),
		signedF("TAUTXADD", tauDigits, &c.Tautxadd),

		signedF("BNBCOM", amtDigits, &c.Bnbcom),
		signedF("BNBDEP", amtDigits, &c.Bnbdep),
		signedF("BNBREG", amtDigits, &c.Bnbreg),
		signedF("BNBSYN", amtDigits, &c.Bnbsyn),
		signedF("BNBCU", amtDigits, &c.Bnbcu),
		signedF("BNBGEM", amtDigits, &c.Bnbgem),
		signedF("BNBTSE", amtDigits, &c.Bnbtse),
		signedF("BNBTSEN(1)", amtDigits, &c.Bnbtsen[0]),
		signedF("BNBTSEN(2)", amtDigits, &c.Bnbtsen[1]),
		signedF("BNBAGR", amtDigits, &c.Bnbagr),
		signedF("BNBBAP", amtDigits, &c.Bnbbap),
		signedF("BNBCAA", amtDigits, &c.Bnbcaa),
		alphaF("FILLER6", filler6Len, &c.Filler6),
		signedF("BNBPLJA", amtDigits, &c.Bnbplja),

		signedF("COTICOM", amtDigits, &c.Coticom),
		signedF("COTIDEP", amtDigits, &c.Cotidep),
		signedF("COTIREG", amtDigits, &c.Cotireg),
		signedF("COTISYN", amtDigits, &c.Cotisyn),
		signedF("COTICU", amtDigits, &c.Coticu),
		signedF("MCOGE3", amtDigits, &c.Mcoge3),
		signedF("COTITSE", amtDigits, &c.Cotitse),
		signedF("COTITSEN(1)", amtDigits, &c.Cotitsen[0]),
		signedF("COTITSEN(2)", amtDigits, &c.Cotitsen[1]),
		signedF("MCNTSA", amtDigits, &c.Mcntsa),
		signedF("COTIAGR", amtDigits, &c.Cotiagr),
		signedF("COTISMAG", amtDigits, &c.Cotismag),
		signedF("COTIBAP", amtDigits, &c.Cotibap),
		signedF("DPROCAA", amtDigits, &c.Dprocaa),
		signedF("DFIXCAA", dfixDigits, &c.Dfixcaa),

		signedF("FFIXCAA", amtDigits, &c.Ffixcaa),
		signedF("FA300", amtDigits, &c.Fa300),
		signedF("FN300", amtDigits, &c.Fn300),
		signedF("FARB", amtDigits, &c.Farb),
		signedF("FRNVB", amtDigits, &c.Frnvb),
		signedF("FARC", amtDigits, &c.Farc),
		signedF("FRNVC", amtDigits, &c.Frnvc),
		signedF("FA800", amtDigits, &c.Fa800),
		signedF("FN800", amtDigits, &c.Fn800),
		signedF("FA900", amtDigits, &c.Fa900),
		signedF("FN900", amtDigits, &c.Fn900),

		signedF("MBJECO", amtDigits, &c.Mbjeco),
		signedF("MBJESY", amtDigits, &c.Mbjesy),
		signedF("MBJEIC", amtDigits, &c.Mbjeic),
		signedF("MCNJCO", amtDigits, &c.Mcnjco),
		signedF("MCNJSY", amtDigits, &c.Mcnjsy),
		signedF("MCNJIC", amtDigits, &c.Mcnjic),
		signedF("MCNJAE", amtDigits, &c.Mcnjae),
		signedF("MFAJAE", amtDigits, &c.Mfajae),
		signedF("MFNJAE", amtDigits, &c.Mfnjae),

		unsignedF("FRAIS-TOTAUX", fraisDigits, &c.FraisTotaux),

		alphaF("CCOIFP", 3, &c.Ccoifp),
		alphaF("CCPPER", 3, &c.Ccpper),
		alphaF("FILLER", fillerLen, &c.Filler),
	}
}

// Normalize applies to every numeric field the truncation a COBOL MOVE into
// its picture would perform (high-order digits dropped, sign dropped for
// unsigned pictures, negative zero becomes zero), so in-memory values equal
// what Encode stores.
func (c *Cotis) Normalize() {
	for _, f := range c.layout() {
		switch f.kind {
		case kindUnsigned:
			*f.num = truncUnsigned(*f.num, f.n)
		case kindSigned:
			*f.num = truncSigned(*f.num, f.n)
		}
	}
}

// ---- codec ----

// ParseCotis decodes a 1004-byte XCOTNB record. It fails if b is not exactly
// CotisLen bytes or if a numeric item fails the COBOL NUMERIC class test
// (the error names the item and its 1-based position).
//
// Signed items accept both IBM ('{', 'A'-'I', '}', 'J'-'R') and ASCII
// ('p'-'y') trailing overpunch signs as well as a plain digit (positive).
func ParseCotis(b []byte) (Cotis, error) {
	var c Cotis
	if len(b) != CotisLen {
		return Cotis{}, fmt.Errorf("nonbati2018: COTISNB (XCOTNB) must be %d bytes, got %d", CotisLen, len(b))
	}
	pos := 0
	for _, f := range c.layout() {
		raw := b[pos : pos+f.n]
		switch f.kind {
		case kindAlpha:
			*f.str = string(raw)
		case kindUnsigned:
			v, ok := digitsValue(string(raw))
			if !ok {
				return Cotis{}, notNumeric(f.name, pos)
			}
			*f.num = v
		case kindSigned:
			v, ok := signedValue(raw)
			if !ok {
				return Cotis{}, notNumeric(f.name, pos)
			}
			*f.num = v
		}
		pos += f.n
	}
	if pos != CotisLen {
		return Cotis{}, fmt.Errorf("nonbati2018: COTISNB layout consumed %d bytes, want %d", pos, CotisLen)
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

// Encode returns the 1004-byte XCOTNB image of the record. Alphanumeric
// items are left-justified, space-padded and right-truncated; numeric items
// are zoned decimal with high-order truncation, and signed items carry a
// trailing overpunch sign in the SignEncoding convention (negative zero is
// stored as +0).
func (c *Cotis) Encode() []byte {
	buf := make([]byte, 0, CotisLen)
	for _, f := range c.layout() {
		switch f.kind {
		case kindAlpha:
			buf = append(buf, fit(*f.str, f.n)...)
		case kindUnsigned:
			buf = append(buf, formatUnsigned(*f.num, f.n)...)
		case kindSigned:
			buf = append(buf, formatSigned(*f.num, f.n, SignEncoding)...)
		}
	}
	return buf
}

// MarshalBinary implements encoding.BinaryMarshaler via Encode.
func (c *Cotis) MarshalBinary() ([]byte, error) {
	b := c.Encode()
	if len(b) != CotisLen {
		return nil, fmt.Errorf("nonbati2018: COTISNB encoded to %d bytes, want %d", len(b), CotisLen)
	}
	return b, nil
}

// ---- zoned-decimal helpers ----

var pow10 = [...]uint64{
	1, 10, 100, 1e3, 1e4, 1e5, 1e6, 1e7, 1e8, 1e9, 1e10,
	1e11, 1e12, 1e13, 1e14, 1e15, 1e16, 1e17, 1e18,
}

// ErrNotNumeric marks a numeric item that failed the COBOL NUMERIC class
// test while decoding a record.
var ErrNotNumeric = errors.New("not NUMERIC")

func notNumeric(name string, start int) error {
	return fmt.Errorf("nonbati2018: COTISNB-%s at position %d: %w", name, start+1, ErrNotNumeric)
}

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

func formatSigned(v int64, n int, mode SignMode) []byte {
	digits := []byte(formatUnsigned(v, n))
	last := digits[n-1] - '0'
	neg := v < 0 && absU(v)%pow10[n] != 0
	switch {
	case mode == SignASCII && neg:
		digits[n-1] = 'p' + last
	case mode == SignASCII:
		// positive: plain digit
	case neg && last == 0:
		digits[n-1] = '}'
	case neg:
		digits[n-1] = 'J' + last - 1
	case last == 0:
		digits[n-1] = '{'
	default:
		digits[n-1] = 'A' + last - 1
	}
	return digits
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

// signedValue decodes a signed zoned-decimal item with trailing overpunch.
func signedValue(raw []byte) (int64, bool) {
	n := len(raw)
	v, ok := digitsValue(string(raw[:n-1]))
	if n == 1 {
		v, ok = 0, true
	}
	digit, neg, signOK := overpunch(raw[n-1])
	if !ok || !signOK {
		return 0, false
	}
	v = v*10 + digit
	if neg {
		return -v, true
	}
	return v, true
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
