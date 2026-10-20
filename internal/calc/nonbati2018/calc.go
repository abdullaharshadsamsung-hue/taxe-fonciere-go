package nonbati2018

// This file is the Go port of EFITA3N8, the 2018 non-built-property (taxe
// foncière non bâtie) calculator. It validates the caller's 600-byte
// COMNONBA block, gets the rates through EFITAUX2 (PARM = 'M') or FMSTAU2
// (any other PARM) and keeps them between calls. It then computes the
// contributions, the management fees, the young-farmer (JA) relief and the
// aggregated totals, and returns the 600-byte RETOURNB block plus CR / RC.
//
// MIGRATION_NOTE: EFITA3N8 creates no schema and does no file or database
// I/O of its own. Rate access goes through EFITAUX2 / FMSTAU2, which are
// ported in internal/rates, so this file needs no DDL.

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"sync"

	"migrated-app/internal/layout/linkage"
	"migrated-app/internal/rates"
)

// Validation / call return codes set by EFITA3N8 itself.
const (
	// CRValidation is the CR set by every parameter check.
	CRValidation = 12
	// RCBadArticleCode is set when CCOBNB is not '1' (non-built article).
	RCBadArticleCode = 1
	// RCBadYear is set when DAN is not '2018'.
	RCBadYear = 2
	// RCBadBase is set when one of the 20 taxable bases is not NUMERIC.
	RCBadBase = 11
	// CRCallFailed is the CR set by CALL ... ON EXCEPTION.
	CRCallFailed = 24
	// RCCallFailed is the RC set by CALL ... ON EXCEPTION.
	RCCallFailed = 1
)

// TaxYear is the only tax year this calculator accepts (COMNONBA-DAN).
const TaxYear = "2018"

// ArticleCodeNonBuilt is the only COMNONBA-CCOBNB value accepted.
const ArticleCodeNonBuilt = "1"

// ParmTaudis is the PARM value that selects EFITAUX2 (CROISIC). Any other
// value selects FMSTAU2 (MAJIC).
const ParmTaudis = "M"

// Fee rates, PIC 9V9(4) held as value * 10^4.
const (
	feeScale  = 4
	f300Frs   = 300 // W-F300FRS 0.0300
	f300Arn   = 100 // W-F300ARN 0.0100
	f800Frs   = 800 // W-F800FRS 0.0800
	f800Arn   = 440 // W-F800ARN 0.0440
	f900Frs   = 900 // W-F900FRS 0.0900
	f900Arn   = 540 // W-F900ARN 0.0540
	fProCaaa  = 781 // W-FPROCAAA 0.0781
	fArcCaaa  = 458 // W-FARCAAA 0.0458
	totDigits = 11  // W-TOTCOT3/8/9: PIC S9(11)
	brtDigits = 12  // W-COTIS-BRUTE: PIC S9(12)
	frsDigits = 10  // W-FRAIS-A-REPARTIR, W-COTDEGJA: PIC S9(10)
)

// Fixed CAAA right (département 57 only).
const (
	caaaDept       = "57"
	dfixCaaaPos    = 2  // W-DFIXCAAA-P
	dfixCaaaNeg    = -2 // W-DFIXCAAA-N
	limBasCaaaPos  = 8  // W-LIMBASCAAA-P
	limBasCaaaNeg  = -8 // W-LIMBASCAAA-N
	stMartinDir    = "971"
	stMartinCommune = "127"
)

// RateSet is the subset of the rate segments (TAU-D-DEP, TAU-R-ROL) that
// EFITA3N8 uses. Rates are held as value * 10^TauxScale. The Sr* fields are
// the TAU-R-ROL key compared by the rate cache.
type RateSet struct {
	// SrJan is TAU-R-JAN (raw 4 bytes).
	SrJan string
	// SrDepdir is TAU-R-DEPDIR (raw 3 bytes).
	SrDepdir string
	// SrCcocom is TAU-R-CCOCOM (raw 3 bytes).
	SrCcocom string
	// SrCcoifp is TAU-R-CCOIFP (raw 3 bytes).
	SrCcoifp string

	// TaucomNB is TAU-R-TAUCOM-NB, the commune rate.
	TaucomNB int64
	// TausyndsfpNB is TAU-R-TAUSYNDSFP-NB, the syndicate rate.
	TausyndsfpNB int64
	// TaucudfpvnNB is TAU-R-TAUCUDFPVN-NB, the EPCI rate.
	TaucudfpvnNB int64
	// Pnbgem is TAU-R-PNBGEM, the GEMAPI rate.
	Pnbgem int64
	// TautseNB is TAU-R-TAUTSE-NB, the TSE 1 rate.
	TautseNB int64
	// Pnbtgp is TAU-R-PNBTGP, the TSE 2 rate.
	Pnbtgp int64
	// Pnbtxa is TAU-R-PNBTXA, the additional tax rate.
	Pnbtxa int64
	// Pnbcaa is TAU-R-PNBCAA, the CAAA proportional rate.
	Pnbcaa int64
	// TauchagrNB is TAU-D-TAUCHAGR-NB, the chamber of agriculture rate.
	TauchagrNB int64
	// TausmagrNB is TAU-D-TAUSMAGR-NB, the Saint-Martin chamber rate.
	TausmagrNB int64
}

// RateDecoder reads a RateSet out of the ZES return area (slot 2 =
// TAU-D-DEP / XBXTDDIR, slot 4 = TAU-R-ROL / XBXTDSR). It must behave like
// COBOL group MOVEs followed by numeric reads (never fails).
//
// MIGRATION_NOTE: injected (as in bati2018) because the XBXTDSR / XBXTDDIR
// non-bâti accessors were not verified at migration time; the wiring code
// must provide an adapter over internal/layout/zes.
type RateDecoder interface {
	Decode(z *rates.ZES) RateSet
}

// TaudisFetcher is the EFITAUX2 entry point (implemented by
// *rates.TaudisSource).
type TaudisFetcher interface {
	Fetch(ctx context.Context, req rates.Request) (rates.Result, error)
}

// Output is what EFITA3N8 hands back: RETOURNBM, CRM and RCM.
type Output struct {
	// Retour is the 600-byte RETOURNB block.
	Retour linkage.RetNB
	// CR is the return code family (0, 12, 24 or a rate source code).
	CR int
	// RC is the detail return code.
	RC int
}

// Calculator is the EFITA3N8 port. Like the non-INITIAL COBOL program it
// keeps the last rate segments (and ZES) between calls and reuses them when
// year, direction, commune and IFP are unchanged. Calls are serialised.
type Calculator struct {
	mu      sync.Mutex
	taudis  TaudisFetcher
	base    rates.BaseTauxFetcher
	decoder RateDecoder

	zes    rates.ZES
	rates  RateSet
	cached bool
}

// NewCalculator builds the calculator. taudis (EFITAUX2) and base (FMSTAU2)
// may be nil: a nil source behaves like a failed dynamic CALL (CR 24 /
// RC 01). decoder is mandatory.
func NewCalculator(taudis TaudisFetcher, base rates.BaseTauxFetcher, decoder RateDecoder) (*Calculator, error) {
	if decoder == nil {
		return nil, errors.New("nonbati2018: nil rate decoder")
	}
	c := &Calculator{taudis: taudis, base: base, decoder: decoder}
	for i := range c.zes {
		for j := range c.zes[i] {
			c.zes[i][j] = ' '
		}
	}
	return c, nil
}

// Calculate runs EFITA3N8 for one article. comnonbam is COMNONBAM (group
// MOVE: space-padded / truncated to 600 bytes); parm is PARM X. A Go error
// is returned only for conditions that would abend the COBOL program;
// functional problems are reported through Output.CR / Output.RC.
func (c *Calculator) Calculate(ctx context.Context, comnonbam []byte, parm string) (Output, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	out := Output{Retour: linkage.NewRetNB()}
	cb := linkage.NewComnba()
	copy(cb[:], comnonbam)

	cr, rc, err := validate(&cb)
	if err != nil {
		return out, err
	}
	if cr > 0 {
		out.CR, out.RC = cr, rc
		return out, nil
	}

	if !c.cacheHit(&cb) {
		cr, rc, err = c.fetch(ctx, &cb, parm)
		if err != nil {
			return out, err
		}
	}

	table, err := c.compute(&cb)
	if err != nil {
		return out, err
	}
	fillRetour(&out.Retour, &cb, &table)

	// MIGRATION_NOTE: as in the source, the calculation still runs after a
	// rate source error (stale or blank rates); the non-zero CR/RC is
	// returned together with the computed block.
	out.CR, out.RC = cr, rc
	return out, nil
}

// validate runs every check; the last failing check wins (source behaviour).
func validate(cb *linkage.Comnba) (cr, rc int, err error) {
	if cb.Ccobnb() != ArticleCodeNonBuilt {
		cr, rc = CRValidation, RCBadArticleCode
	}
	if cb.DanRaw() != TaxYear {
		cr, rc = CRValidation, RCBadYear
	}

	oks := make([]bool, 0, 20)
	add := func(_ int64, ok bool) { oks = append(oks, ok) }
	add(cb.Mbacom())
	add(cb.Mbadep())
	add(cb.Mbasyn())
	add(cb.Mbacu())
	add(cb.Mbage3())
	for i := 0; i < linkage.ComnbaNt13Count; i++ {
		_, ok, e := cb.Mbnt13(i)
		if e != nil {
			return 0, 0, e
		}
		oks = append(oks, ok)
	}
	add(cb.Mbaca())
	add(cb.Mbacaa())
	add(cb.Mbjcom())
	add(cb.Mbjsyn())
	add(cb.Mbjcu())
	for i := 0; i < linkage.ComnbaJt13Count; i++ {
		_, ok, e := cb.Mbjt13(i)
		if e != nil {
			return 0, 0, e
		}
		oks = append(oks, ok)
	}
	add(cb.Mbjdep())
	add(cb.Mb1pre())
	add(cb.Mbjpre())
	add(cb.Mbjeco())
	add(cb.Mbjesy())
	add(cb.Mbjeic())
	for _, ok := range oks {
		if !ok {
			cr, rc = CRValidation, RCBadBase
			break
		}
	}
	return cr, rc, nil
}

// cacheHit is the EVALUATE TRUE ALSO TRUE ... on TAU-R-ROL (CCPPER is not
// part of the key, source behaviour).
func (c *Calculator) cacheHit(cb *linkage.Comnba) bool {
	if !c.cached {
		return false
	}
	return cb.DanRaw() == c.rates.SrJan &&
		cb.Ac3dir() == c.rates.SrDepdir &&
		cb.Ccocom() == c.rates.SrCcocom &&
		cb.Ccoifp() == c.rates.SrCcoifp
}

// fetch calls EFITAUX2 or FMSTAU2 and refreshes the cached segments.
func (c *Calculator) fetch(ctx context.Context, cb *linkage.Comnba, parm string) (cr, rc int, err error) {
	if len(parm) > 0 && parm[:1] == ParmTaudis {
		if c.taudis == nil {
			cr, rc = CRCallFailed, RCCallFailed // ON EXCEPTION: ZES unchanged
		} else {
			res, ferr := c.taudis.Fetch(ctx, rates.Request{
				Animp:  cb.DanRaw(),
				Coddir: cb.Ac3dir(),
				Codcom: cb.Ccocom(),
				Ccoifp: cb.Ccoifp(),
				Ccpper: cb.Ccpper(),
			})
			if ferr != nil {
				return 0, 0, fmt.Errorf("nonbati2018: EFITAUX2: %w", ferr)
			}
			cr, rc, c.zes = res.CR, res.RC, res.ZES
		}
	} else {
		if c.base == nil {
			cr, rc = CRCallFailed, RCCallFailed
		} else {
			res, ferr := c.base.Fetch(ctx, rates.BaseTauxRequest{
				Janipt: cb.DanRaw(),
				Acodir: cb.Ac3dir(),
				Ccocom: cb.Ccocom(),
				Ccoifp: cb.Ccoifp(),
			})
			if ferr != nil {
				return 0, 0, fmt.Errorf("nonbati2018: FMSTAU2: %w", ferr)
			}
			cr, rc, c.zes = res.CR, res.RC, res.ZES
		}
		// MIGRATION_NOTE: the MOVE ZESTAUX(4) TO TAU-R-ROL1 / self-MOVEs /
		// MOVE TAU-R-ROL1 TO ZESTAUX(4) sequence does nothing when XBXTDSR
		// is 400 bytes long. If the copybook is shorter, the tail of slot 4
		// gets blanked. The XBXTDSR length was not available, so review this
		// if it is not 400.
	}
	c.rates = c.decoder.Decode(&c.zes)
	c.cached = true
	return cr, rc, nil
}

// table is the computed COTISNB (postes 1..5, index 0..4) plus the working
// totals of poste 1 that feed the aggregates.
type table struct {
	cotis [PosteCount]Cotis
	t3    int64
	t8    int64
	t9    int64
}

// compute builds postes 1 (I BN) and 4 (G JA) of COTISNB.
func (c *Calculator) compute(cb *linkage.Comnba) (table, error) {
	var tb table
	tb.cotis = NewCotisTable()
	p1 := &tb.cotis[PosteImposition-1]
	r := &c.rates

	acodnb := ccobnbDigit(cb)
	p1.Cle = cb.Aidfic()
	p1.Anupro = cb.Anupro()
	p1.Acodnb = acodnb
	p1.Impot = "I"
	p1.Gnexps = "BN"

	p1.Taucom = truncSigned(r.TaucomNB, tauDigits)
	p1.Tausyn = truncSigned(r.TausyndsfpNB, tauDigits)
	p1.Taucu = truncSigned(r.TaucudfpvnNB, tauDigits)
	p1.Taugem = truncSigned(r.Pnbgem, tauDigits)
	p1.Tautsen[0] = truncSigned(r.TautseNB, tauDigits)
	p1.Tautsen[1] = truncSigned(r.Pnbtgp, tauDigits)
	p1.Tautxadd = truncSigned(r.Pnbtxa, tauDigits)
	p1.Tauchagr = truncSigned(r.TauchagrNB, tauDigits)
	p1.Taucaaa = truncSigned(r.Pnbcaa, tauDigits)
	p1.Tausmagr = truncSigned(r.TausmagrNB, tauDigits)

	p1.Bnbcom, _ = cb.Mbacom()
	p1.Bnbdep, _ = cb.Mbadep()
	p1.Bnbsyn, _ = cb.Mbasyn()
	p1.Bnbcu, _ = cb.Mbacu()
	p1.Bnbgem, _ = cb.Mbage3()
	for i := range p1.Bnbtsen {
		v, _, err := cb.Mbnt13(i)
		if err != nil {
			return tb, err
		}
		p1.Bnbtsen[i] = v
	}
	p1.Bnbagr, _ = cb.Mbaca()
	p1.Bnbcaa, _ = cb.Mbacaa()

	// Contributions.
	p1.Coticom = contribution(p1.Bnbcom, p1.Taucom)
	p1.Cotidep = contribution(p1.Bnbdep, p1.Tautxadd)
	p1.Cotisyn = contribution(p1.Bnbsyn, p1.Tausyn)
	p1.Coticu = contribution(p1.Bnbcu, p1.Taucu)
	p1.Mcoge3 = contribution(p1.Bnbgem, p1.Taugem)
	p1.Cotitsen[0] = contribution(p1.Bnbtsen[0], p1.Tautsen[0])
	p1.Cotitsen[1] = contribution(p1.Bnbtsen[1], p1.Tautsen[1])
	p1.Mcntsa = truncSigned(p1.Cotitsen[0]+p1.Cotitsen[1], amtDigits)
	p1.Cotismag = contribution(p1.Bnbagr, p1.Tausmagr)
	p1.Cotiagr = contribution(p1.Bnbagr, p1.Tauchagr)

	// CAAA fixed right: département 57 only, thresholds on |base| >= 8.
	if p1.Ccodep() == caaaDept {
		if p1.Bnbcaa > 0 && p1.Bnbcaa >= limBasCaaaPos {
			p1.Dfixcaa = dfixCaaaPos
		}
		if p1.Bnbcaa < 0 && p1.Bnbcaa <= limBasCaaaNeg {
			p1.Dfixcaa = dfixCaaaNeg
		}
	}
	// CAAA proportional right, only when the rate is non-zero.
	if p1.Taucaaa != 0 {
		p1.Dprocaa = contribution(p1.Bnbcaa, p1.Taucaaa)
	}

	// Fees.
	tb.t3 = truncSigned(p1.Coticom+p1.Coticu+p1.Cotidep+p1.Mcoge3, totDigits)
	tb.t8 = truncSigned(p1.Cotisyn+p1.Cotiagr, totDigits)
	tb.t9 = truncSigned(p1.Cotitsen[0]+p1.Cotitsen[1], totDigits)

	p1.Fa300, p1.Fn300 = fees(tb.t3, f300Arn, f300Frs)
	p1.Fa800, p1.Fn800 = fees(tb.t8, f800Arn, f800Frs)
	p1.Fa900, p1.Fn900 = fees(tb.t9, f900Arn, f900Frs)
	if p1.Fa800 < p1.Fn800 {
		p1.Fa800 = truncSigned(p1.Fa800+1, amtDigits)
		p1.Fn800 = truncSigned(p1.Fn800-1, amtDigits)
	}
	if p1.Fa900 < p1.Fn900 {
		p1.Fa900 = truncSigned(p1.Fa900+1, amtDigits)
		p1.Fn900 = truncSigned(p1.Fn900-1, amtDigits)
	}
	p1.Farc, p1.Frnvc = fees(p1.Dprocaa, fArcCaaa, fProCaaa)

	// Poste 4: young-farmer relief, only if one JA base is non-zero.
	mbjcom, _ := cb.Mbjcom()
	mbjsyn, _ := cb.Mbjsyn()
	mbjcu, _ := cb.Mbjcu()
	mbjeco, _ := cb.Mbjeco()
	mbjesy, _ := cb.Mbjesy()
	mbjeic, _ := cb.Mbjeic()
	if mbjcom != 0 || mbjsyn != 0 || mbjcu != 0 || mbjeco != 0 || mbjesy != 0 || mbjeic != 0 {
		p4 := &tb.cotis[PosteJeunesAgriculteurs-1]
		p4.Cle = cb.Aidfic()
		p4.Anupro = cb.Anupro()
		p4.Acodnb = acodnb
		p4.Impot = "G"
		p4.Gnexps = "JA"
		p4.Bnbcom = mbjcom
		p4.Bnbsyn = mbjsyn
		p4.Bnbcu = mbjcu
		p4.Mbjeco = mbjeco
		p4.Mbjesy = mbjesy
		p4.Mbjeic = mbjeic

		// Rates come from poste 1.
		p4.Coticom = contribution(p4.Bnbcom, p1.Taucom)
		p4.Cotisyn = contribution(p4.Bnbsyn, p1.Tausyn)
		p4.Coticu = contribution(p4.Bnbcu, p1.Taucu)
		p4.Mcnjco = contribution(p4.Mbjeco, p1.Taucom)
		p4.Mcnjsy = contribution(p4.Mbjesy, p1.Tausyn)
		p4.Mcnjic = contribution(p4.Mbjeic, p1.Taucu)
		p4.Mcnjae = truncSigned(p4.Mcnjco+p4.Mcnjsy+p4.Mcnjic, amtDigits)

		t3 := truncSigned(p4.Coticom+p4.Coticu+p4.Mcnjco+p4.Mcnjic, totDigits)
		t8 := truncSigned(p4.Cotisyn+p4.Mcnjsy, totDigits)
		// Amalgamated fees, rounded once on the summed products.
		arn := new(big.Int).Add(
			new(big.Int).Mul(big.NewInt(t3), big.NewInt(f300Arn)),
			new(big.Int).Mul(big.NewInt(t8), big.NewInt(f800Arn)))
		frs := new(big.Int).Add(
			new(big.Int).Mul(big.NewInt(t3), big.NewInt(f300Frs)),
			new(big.Int).Mul(big.NewInt(t8), big.NewInt(f800Frs)))
		p4.Mfajae = truncSigned(roundQuo(arn, feeScale), amtDigits)
		wFrais := truncSigned(roundQuo(frs, feeScale), frsDigits)
		p4.Mfnjae = truncSigned(wFrais-p4.Mfajae, amtDigits)
	}

	// Saint-Martin: the consular chamber contribution joins the totals.
	if cb.Ac3dir() == stMartinDir && cb.Ccocom() == stMartinCommune {
		tb.t3 = truncSigned(tb.t3+p1.Cotismag, totDigits)
	}
	return tb, nil
}

// fillRetour fills RETOURNB from COTISNB and computes the aggregates.
func fillRetour(rb *linkage.RetNB, cb *linkage.Comnba, tb *table) {
	p1 := &tb.cotis[PosteImposition-1]
	p4 := &tb.cotis[PosteJeunesAgriculteurs-1]

	rb.SetAidfic(p1.Cle)
	rb.SetAnupro(p1.Anupro)
	rb.SetCcobnb(formatUnsigned(p1.Acodnb, 1))

	rb.SetMctcom(p1.Coticom)
	rb.SetMctsyn(p1.Cotisyn)
	rb.SetMctdep(p1.Cotidep)
	rb.SetMctcu(p1.Coticu)
	rb.SetMcoge3(p1.Mcoge3)
	for i := range p1.Cotitsen {
		// Index is always within RetNBNt13Count (2); error impossible.
		_ = rb.SetMcnt13(i, p1.Cotitsen[i])
	}
	rb.SetMcntsa(p1.Mcntsa)
	rb.SetMctca(p1.Cotiagr)
	rb.SetMctcas(p1.Cotismag)
	rb.SetMctcap(p1.Dprocaa)
	rb.SetMctcad(p1.Dfixcaa)

	rb.SetMfa300(p1.Fa300)
	rb.SetMfn300(p1.Fn300)
	rb.SetMcfaca(p1.Farc)
	rb.SetMcfnvc(p1.Frnvc)
	rb.SetMfa800(p1.Fa800)
	rb.SetMfn800(p1.Fn800)
	rb.SetMfa900(p1.Fa900)
	rb.SetMfn900(p1.Fn900)

	rb.SetMdjcom(p4.Coticom)
	rb.SetMdjsyn(p4.Cotisyn)
	rb.SetMdjcu(p4.Coticu)
	rb.SetMcnjae(p4.Mcnjae)
	rb.SetMfaj1e(p4.Mfajae)
	rb.SetMfnj1e(p4.Mfnjae)

	brute := truncSigned(tb.t3+tb.t8+tb.t9+p1.Dprocaa+p1.Dfixcaa, brtDigits)
	degJA := truncSigned(p4.Coticom+p4.Cotisyn+p4.Coticu+p4.Mcnjae+p4.Mfajae+p4.Mfnjae, frsDigits)
	rb.SetTcthfr(brute - degJA)

	// TCTFRA is summed from the RETOURNB fields as stored (source reads
	// them back from RETOURNB).
	var feesTotal int64
	for _, get := range []func() (int64, bool){
		rb.Mfa300, rb.Mfn300, rb.Mcfaca, rb.Mcfnvc,
		rb.Mfa800, rb.Mfn800, rb.Mfa900, rb.Mfn900,
	} {
		v, _ := get()
		feesTotal += v
	}
	rb.SetTctfra(feesTotal)
	tctfra, _ := rb.Tctfra()
	rb.SetTctdu(brute + tctfra - degJA)
	_ = cb
}

// ccobnbDigit is MOVE COMNONBA-CCOBNB (X) TO ACODNB (9): the digit nibble.
func ccobnbDigit(cb *linkage.Comnba) int64 {
	s := cb.Ccobnb()
	if len(s) == 0 {
		return 0
	}
	d := int64(s[0] & 0x0F)
	if d > 9 {
		return 0
	}
	return d
}

// contribution is COMPUTE x ROUNDED = base * rate / 100 into S9(10), with
// rate at TauxScale decimals, half-away-from-zero rounding and high-order
// truncation (no ON SIZE ERROR).
func contribution(base, rate int64) int64 {
	num := new(big.Int).Mul(big.NewInt(base), big.NewInt(rate))
	return truncSigned(roundQuo(num, TauxScale+2), amtDigits)
}

// fees computes FA = ROUNDED(total*arn) into S9(10), W-FRAIS-A-REPARTIR =
// ROUNDED(total*frs) into S9(10) and FN = W-FRAIS - FA into S9(10).
func fees(total, arn, frs int64) (fa, fn int64) {
	fa = truncSigned(roundQuo(new(big.Int).Mul(big.NewInt(total), big.NewInt(arn)), feeScale), amtDigits)
	wFrais := truncSigned(roundQuo(new(big.Int).Mul(big.NewInt(total), big.NewInt(frs)), feeScale), frsDigits)
	fn = truncSigned(wFrais-fa, amtDigits)
	return fa, fn
}

// roundQuo returns num / 10^scale rounded half away from zero, reduced
// modulo 10^18 so it fits an int64 (callers truncate to the picture).
func roundQuo(num *big.Int, scale int) int64 {
	den := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(scale)), nil)
	q, r := new(big.Int).QuoRem(num, den, new(big.Int))
	twice := new(big.Int).Mul(new(big.Int).Abs(r), big.NewInt(2))
	if twice.Cmp(den) >= 0 {
		if num.Sign() < 0 {
			q.Sub(q, big.NewInt(1))
		} else {
			q.Add(q, big.NewInt(1))
		}
	}
	mod := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	neg := q.Sign() < 0
	q.Abs(q).Mod(q, mod)
	v := q.Int64()
	if neg {
		return -v
	}
	return v
}
