package bati2018

// This file is the Go port of EFITA3B8, the 2018 built-property (taxe
// foncière bâtie) calculator. It validates the caller's 600-byte COMBAT
// block and gets the rates through EFITAUX2 (PARM = 'M') or FMSTAU2 (any
// other PARM), caching them between calls. It then computes the
// contributions and the management fees and returns the 600-byte RETOURB
// block plus the CR / RC pair.
//
// MIGRATION_NOTE: EFITA3B8 performs no schema creation and no file or
// database I/O of its own. Rate access is delegated to EFITAUX2 / FMSTAU2,
// ported in internal/rates. No DDL is therefore required here.

import (
	"context"
	"fmt"
	"math/big"
	"sync"

	"migrated-app/internal/layout/linkage"
	"migrated-app/internal/rates"
)

// Validation / call return codes set by EFITA3B8 itself.
const (
	// CRValidation is the CR set by every parameter check.
	CRValidation = 12
	// RCBadArticleCode is set when CCOBNB is not 2 (built article).
	RCBadArticleCode = 1
	// RCBadYear is set when DAN is not '2018'.
	RCBadYear = 2
	// RCBadOmZone is set when a GTAUOM fails the level-88 GTAUOM-OK.
	RCBadOmZone = 5
	// RCBadOmBase is set when a non-blank GTAUOM has a non-numeric MBAOM.
	RCBadOmBase = 6
	// RCBadBase is set when one of the taxable bases is not NUMERIC.
	RCBadBase = 11
	// CRCallFailed is the CR set by CALL ... ON EXCEPTION.
	CRCallFailed = 24
	// RCCallFailed is the RC set by CALL ... ON EXCEPTION.
	RCCallFailed = 1
)

// TaxYear is the only tax year this calculator accepts (COMBAT-DAN).
const TaxYear = "2018"

// ArticleCodeBuilt is the only COMBAT-CCOBNB value accepted (built article).
const ArticleCodeBuilt = 2

// ParmTaudis is the PARM value that selects EFITAUX2 (CROISIC, general TF
// roll). Any other value selects FMSTAU2 (MAJIC).
const ParmTaudis = "M"

// Fee rates, PIC 9V9(4) held as value * 10^4 (W-F300FRS ... W-F900ARN).
const (
	feeScale  = 4
	f300Frs   = 300 // 0.0300
	f300Arn   = 100 // 0.0100
	f800Frs   = 800 // 0.0800
	f800Arn   = 440 // 0.0440
	f900Frs   = 900 // 0.0900
	f900Arn   = 540 // 0.0540
	totDigits = 11  // W-TOTCOT3/8/9, W-TOTCOTOM, W-FRAIS: PIC S9(11)
)

// RateSet is the subset of the rate segments (TAU-E-AN, TAU-D-DEP,
// TAU-C-COM, TAU-R-ROL) that EFITA3B8 uses. Rates are held as value *
// 10^TauxScale. The Sr* fields are the key of the TAU-R-ROL segment,
// which the rate cache compares with the request.
type RateSet struct {
	// SrJan is TAU-R-JAN (raw 4 bytes).
	SrJan string
	// SrDepdir is TAU-R-DEPDIR (raw 3 bytes).
	SrDepdir string
	// SrCcocom is TAU-R-CCOCOM (raw 3 bytes).
	SrCcocom string
	// SrCcoifp is TAU-R-CCOIFP (raw 3 bytes).
	SrCcoifp string

	// TaucomB is TAU-R-TAUCOM-B, the commune rate.
	TaucomB int64
	// TaudepB is TAU-D-TAUDEP-B, the département rate.
	TaudepB int64
	// Ptbtas is TAU-D-PTBTAS, the TASA rate.
	Ptbtas int64
	// TausyndsfpB is TAU-R-TAUSYNDSFP-B, the syndicate rate.
	TausyndsfpB int64
	// TaucudfpvnB is TAU-R-TAUCUDFPVN-B, the EPCI rate.
	TaucudfpvnB int64
	// TautseB is TAU-R-TAUTSE-B, the TSE 1 rate.
	TautseB int64
	// Ptbtgp is TAU-R-PTBTGP, the TSE 2 rate.
	Ptbtgp int64
	// Ptbgem is TAU-R-PTBGEM, the GEMAPI rate.
	Ptbgem int64
	// Pbbom are TAU-C-PBBOMP, PBBOMA .. PBBOME (full rate, reduced A..E).
	Pbbom [CotisOmCount]int64
}

// RateDecoder reads a RateSet out of the ZES return area: slot 1 = TAU-E-AN
// (XBXTDAN), 2 = TAU-D-DEP (XBXTDDIR), 3 = TAU-C-COM (XBXTDCOM),
// 4 = TAU-R-ROL (XBXTDSR). It must behave like COBOL group MOVEs followed by
// numeric reads (no error: invalid digits read from their low nibbles).
//
// MIGRATION_NOTE: injected so this calculator does not depend on accessor
// names of internal/layout/zes that were not verified at migration time;
// the wiring code must supply an adapter over the XBXTD* layouts.
type RateDecoder interface {
	Decode(z *rates.ZES) RateSet
}

// TaudisFetcher is the EFITAUX2 entry point (implemented by
// *rates.TaudisSource).
type TaudisFetcher interface {
	Fetch(ctx context.Context, req rates.Request) (rates.Result, error)
}

// Output is what EFITA3B8 hands back to its caller: RETOURM, CRM and RCM.
type Output struct {
	// Retour is the 600-byte RETOURB block.
	Retour linkage.RetB
	// CR is the return code family (0, 12, 24 or a rate source code).
	CR int
	// RC is the detail return code.
	RC int
}

// totals are the W-TOTCOTOM / W-TOTCOT3 / W-TOTCOT8 / W-TOTCOT9 working
// fields (PIC S9(11) GLOBAL).
type totals struct {
	om int64
	t3 int64
	t8 int64
	t9 int64
}

// Calculator is the EFITA3B8 port. Like the non-INITIAL COBOL program, it
// keeps the last rate segments (and the ZES area) between calls and reuses
// them when year, direction, commune and IFP are unchanged. Safe for
// concurrent use; calls are serialised.
type Calculator struct {
	mu      sync.Mutex
	taudis  TaudisFetcher
	base    rates.BaseTauxFetcher
	decoder RateDecoder

	zes    rates.ZES
	rates  RateSet
	cached bool // false = TAU-R-ROL never filled: the cache test never matches

	// lastTotals mirrors the GLOBAL working-storage totals, which keep the
	// value of the last successful calculation between calls.
	lastTotals totals
}

// NewCalculator builds the calculator. taudis (EFITAUX2) and base (FMSTAU2)
// may be nil: a nil source behaves like a failed dynamic CALL (CR 24 /
// RC 01). decoder is mandatory.
func NewCalculator(taudis TaudisFetcher, base rates.BaseTauxFetcher, decoder RateDecoder) (*Calculator, error) {
	if decoder == nil {
		return nil, fmt.Errorf("bati2018: nil rate decoder")
	}
	c := &Calculator{taudis: taudis, base: base, decoder: decoder}
	// ZES has no VALUE clause; spaces is the conventional initial content
	// (same choice as internal/rates).
	for i := range c.zes {
		for j := range c.zes[i] {
			c.zes[i][j] = ' '
		}
	}
	return c, nil
}

// Calculate runs EFITA3B8 for one article. combatm is COMBATM (moved as a
// group: padded with spaces / truncated to 600 bytes); parm is PARM X.
// A Go error is returned only for conditions that would abend the COBOL
// program (rate source infrastructure failure, layout misuse); every
// functional problem is reported through Output.CR / Output.RC.
func (c *Calculator) Calculate(ctx context.Context, combatm []byte, parm string) (Output, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	out := Output{Retour: linkage.NewRetB()}
	combat := moveCombat(combatm)

	cr, rc, err := validate(&combat)
	if err != nil {
		return out, err
	}
	if cr > 0 {
		out.CR, out.RC = cr, rc
		return out, nil
	}

	if !c.cacheHit(&combat) {
		cr, rc, err = c.fetch(ctx, &combat, parm)
		if err != nil {
			return out, err
		}
	}

	cotis, tot, err := c.compute(&combat)
	if err != nil {
		return out, err
	}
	c.lastTotals = tot
	if err := fillRetour(&out.Retour, &cotis, tot); err != nil {
		return out, err
	}
	// MIGRATION_NOTE: like the source, the calculation still runs after a
	// rate source error (stale or blank rates) and the non-zero CR/RC is
	// returned together with the computed block.
	out.CR, out.RC = cr, rc
	return out, nil
}

// moveCombat is MOVE COMBATM TO COMBAT.
func moveCombat(b []byte) linkage.Combat {
	cb := linkage.NewCombat()
	copy(cb[:], b)
	return cb
}

// ccobnbDigit returns the numeric value of COMBAT-CCOBNB as a COBOL numeric
// comparison / numeric MOVE sees it: the digit nibble of the byte.
//
// MIGRATION_NOTE: XCOMBAT was not available; CCOBNB is assumed to be a
// one-byte PIC 9 DISPLAY field. The source compares it with the numeric
// literal 2, which IBM COBOL performs on the digit (low) nibble, ignoring
// the zone. A blank field reads as 0 and is rejected.
func ccobnbDigit(cb *linkage.Combat) int64 {
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

// validate runs every check; the last failing check sets CR / RC (the
// source does not stop at the first failure). Blank OM zones get their
// base forced to zero, as in the source.
func validate(cb *linkage.Combat) (cr, rc int, err error) {
	if ccobnbDigit(cb) != ArticleCodeBuilt {
		cr, rc = CRValidation, RCBadArticleCode
	}
	if cb.DanRaw() != TaxYear {
		cr, rc = CRValidation, RCBadYear
	}

	_, ok1 := cb.Mbacom()
	_, ok2 := cb.Mbadep()
	_, ok3 := cb.Mbareg()
	_, ok4 := cb.Mbasyn()
	_, ok5 := cb.Mbacu()
	_, ok6 := cb.Mbage3()
	_, ok7 := cb.Mbata3()
	_, ok8, err := cb.Mbbt13(0)
	if err != nil {
		return 0, 0, err
	}
	_, ok9, err := cb.Mbbt13(1)
	if err != nil {
		return 0, 0, err
	}
	if !(ok1 && ok2 && ok3 && ok4 && ok5 && ok6 && ok7 && ok8 && ok9) {
		cr, rc = CRValidation, RCBadBase
	}

	for i := 0; i < linkage.CombatOmCount; i++ {
		zoneOK, err := cb.GtauomOK(i)
		if err != nil {
			return 0, 0, err
		}
		if !zoneOK {
			cr, rc = CRValidation, RCBadOmZone
		}
		g, err := cb.Gtauom(i)
		if err != nil {
			return 0, 0, err
		}
		_, baseOK, err := cb.Mbaom(i)
		if err != nil {
			return 0, 0, err
		}
		if g != "  " && !baseOK {
			cr, rc = CRValidation, RCBadOmBase
		}
		if g == "  " {
			if err := cb.SetMbaom(i, 0); err != nil {
				return 0, 0, err
			}
		}
	}
	return cr, rc, nil
}

// cacheHit is the EVALUATE TRUE ALSO TRUE ... on TAU-R-ROL. CCPPER is not
// part of the key (source behaviour).
//
// MIGRATION_NOTE: the comparisons are done on the raw bytes. COMBAT-DAN vs
// TAU-R-JAN is a numeric comparison in COBOL; it only differs from a raw
// compare for non-digit content, which validation has already ruled out for
// DAN ('2018').
func (c *Calculator) cacheHit(cb *linkage.Combat) bool {
	if !c.cached {
		return false
	}
	return cb.DanRaw() == c.rates.SrJan &&
		cb.Ac3dir() == c.rates.SrDepdir &&
		cb.Ccocom() == c.rates.SrCcocom &&
		cb.Ccoifp() == c.rates.SrCcoifp
}

// fetch calls EFITAUX2 or FMSTAU2 and refreshes the cached segments.
func (c *Calculator) fetch(ctx context.Context, cb *linkage.Combat, parm string) (cr, rc int, err error) {
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
				return 0, 0, fmt.Errorf("bati2018: EFITAUX2: %w", ferr)
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
				return 0, 0, fmt.Errorf("bati2018: FMSTAU2: %w", ferr)
			}
			cr, rc, c.zes = res.CR, res.RC, res.ZES
		}
		// MIGRATION_NOTE: MOVE ZESTAUX(4) TO TAU-R-ROL1 / MOVE TAU-R-ROL1 TO
		// ZESTAUX(4) is a no-op when XBXTDSR is 400 bytes long; if the
		// copybook is shorter, the tail of slot 4 is blanked. XBXTDSR length
		// was not available; review if it differs from 400.
	}
	c.rates = c.decoder.Decode(&c.zes)
	c.cached = true
	return cr, rc, nil
}

// compute builds COTISB (bases, rates, contributions and fees) and returns
// it together with the working totals used to fill RETOURB.
func (c *Calculator) compute(cb *linkage.Combat) (Cotis, totals, error) {
	ct := NewCotis()
	r := &c.rates

	ct.Cle = cb.Aidfic()
	ct.Anupro = cb.Anupro()
	ct.Acodb = truncUnsigned(ccobnbDigit(cb), 1)
	ct.Impot = "I"
	ct.Gnexpl = "BN"

	ct.Taucom = truncSigned(r.TaucomB, tauDigits)
	ct.Taudep = truncSigned(r.TaudepB, tauDigits)
	ct.Tautas = truncSigned(r.Ptbtas, tauDigits)
	ct.Tausyn = truncSigned(r.TausyndsfpB, tauDigits)
	ct.Taucu = truncSigned(r.TaucudfpvnB, tauDigits)
	ct.Tautsen[0] = truncSigned(r.TautseB, tauDigits)
	ct.Tautsen[1] = truncSigned(r.Ptbtgp, tauDigits)
	ct.Taugem = truncSigned(r.Ptbgem, tauDigits)
	for i := range ct.Tauom1 {
		ct.Tauom1[i] = truncSigned(r.Pbbom[i], tauDigits)
	}

	ct.Bbcom, _ = cb.Mbacom()
	ct.Bbdep, _ = cb.Mbadep()
	ct.Bbsyn, _ = cb.Mbasyn()
	ct.Bbcu, _ = cb.Mbacu()
	ct.Bbgem, _ = cb.Mbage3()
	ct.Bbtas, _ = cb.Mbata3()
	for i := range ct.Bbtsen {
		v, _, err := cb.Mbbt13(i)
		if err != nil {
			return ct, totals{}, err
		}
		ct.Bbtsen[i] = v
	}
	for i := range ct.TabOm {
		g, err := cb.Gtauom(i)
		if err != nil {
			return ct, totals{}, err
		}
		v, _, err := cb.Mbaom(i)
		if err != nil {
			return ct, totals{}, err
		}
		ct.TabOm[i] = OmZone{Gtauom: g, Bbteom: v}
	}

	// Contributions.
	ct.Coticom = contribution(ct.Bbcom, ct.Taucom)
	ct.Cotidep = contribution(ct.Bbdep, ct.Taudep)
	ct.Cotisyn = contribution(ct.Bbsyn, ct.Tausyn)
	ct.Coticu = contribution(ct.Bbcu, ct.Taucu)
	ct.Mcoge3 = contribution(ct.Bbgem, ct.Taugem)
	ct.Mcota3 = contribution(ct.Bbtas, ct.Tautas)
	ct.Cotitsen[0] = contribution(ct.Bbtsen[0], ct.Tautsen[0])
	ct.Cotitsen[1] = contribution(ct.Bbtsen[1], ct.Tautsen[1])
	ct.Mcbtsa = truncSigned(ct.Cotitsen[0]+ct.Cotitsen[1]+ct.Mcota3, amtDigits)

	// Household waste (TEOM) by zone. BASEB-GTAUOM holds the same value as
	// COTISB-GTAUOM, so BASEB (never returned) is not materialised.
	// WHEN OTHER CONTINUE leaves the INITIALIZEd zero in place.
	for i, z := range ct.TabOm {
		switch z.Gtauom {
		case "  ":
			ct.CotisOm[i] = 0
		case "P ", " P":
			ct.CotisOm[i] = contribution(z.Bbteom, ct.Tauom1[0])
		case "RA":
			ct.CotisOm[i] = contribution(z.Bbteom, ct.Tauom1[1])
		case "RB":
			ct.CotisOm[i] = contribution(z.Bbteom, ct.Tauom1[2])
		case "RC":
			ct.CotisOm[i] = contribution(z.Bbteom, ct.Tauom1[3])
		case "RD":
			ct.CotisOm[i] = contribution(z.Bbteom, ct.Tauom1[4])
		case "RE":
			ct.CotisOm[i] = contribution(z.Bbteom, ct.Tauom1[5])
		}
	}
	// TEOMI: MVLTIM is not checked for NUMERIC by the source.
	ct.CotisOmi, _ = cb.Mvltim()

	// Fees. The source sums COTIS-OM(1..7): occurrence 7 lies outside the
	// 6-entry table (FA300 in storage, still zero here); reproduced through
	// CotisOmSubscript.
	var t totals
	for ind := 1; ind <= CotisOmCount+1; ind++ {
		v, err := ct.CotisOmSubscript(ind)
		if err != nil {
			return ct, totals{}, err
		}
		t.om = truncSigned(t.om+v, totDigits)
	}
	t.om = truncSigned(t.om+ct.CotisOmi, totDigits)
	t.t3 = truncSigned(ct.Coticom+ct.Cotidep+ct.Coticu+ct.Mcoge3, totDigits)
	t.t8 = truncSigned(t.om+ct.Cotisyn+ct.Mcota3, totDigits)
	t.t9 = truncSigned(ct.Cotitsen[0]+ct.Cotitsen[1], totDigits)

	ct.Fa300, ct.Fn300 = fees(t.t3, f300Arn, f300Frs)
	ct.Fa800, ct.Fn800 = fees(t.t8, f800Arn, f800Frs)
	ct.Fa900, ct.Fn900 = fees(t.t9, f900Arn, f900Frs)
	// Rebalance so that assessment fees >= relief fees (8% and 9% only).
	if ct.Fa800 < ct.Fn800 {
		ct.Fa800 = truncSigned(ct.Fa800+1, amtDigits)
		ct.Fn800 = truncSigned(ct.Fn800-1, amtDigits)
	}
	if ct.Fa900 < ct.Fn900 {
		ct.Fa900 = truncSigned(ct.Fa900+1, amtDigits)
		ct.Fn900 = truncSigned(ct.Fn900-1, amtDigits)
	}

	return ct, t, nil
}

// totalsOf computes the working totals (W-TOTCOTOM, W-TOTCOT3/8/9) from the
// contributions of ct. TCTOM is the sum of the six TEOM zones plus TEOMI;
// the fees (FA300 ...) never enter it.
func totalsOf(ct *Cotis) totals {
	var t totals
	for _, v := range ct.CotisOm {
		t.om = truncSigned(t.om+v, totDigits)
	}
	t.om = truncSigned(t.om+ct.CotisOmi, totDigits)
	t.t3 = truncSigned(ct.Coticom+ct.Cotidep+ct.Coticu+ct.Mcoge3, totDigits)
	t.t8 = truncSigned(t.om+ct.Cotisyn+ct.Mcota3, totDigits)
	t.t9 = truncSigned(ct.Cotitsen[0]+ct.Cotitsen[1], totDigits)
	return t
}

// fillRetour fills RETOURB from COTISB and the working totals.
func fillRetour(rb *linkage.RetB, ct *Cotis, t totals) error {
	rb.SetAidfic(ct.Cle)
	rb.SetAnupro(ct.Anupro)
	rb.SetCcobnb(formatUnsigned(ct.Acodb, 1))

	rb.SetMctcom(ct.Coticom)
	rb.SetMctdep(ct.Cotidep)
	rb.SetMctsyn(ct.Cotisyn)
	rb.SetMctcu(ct.Coticu)
	rb.SetMcoge3(ct.Mcoge3)
	rb.SetMcota3(ct.Mcota3)
	for i := range ct.Cotitsen {
		if err := rb.SetMcbt13(i, ct.Cotitsen[i]); err != nil {
			return err
		}
	}
	rb.SetMcbtsa(ct.Mcbtsa)
	rb.SetMvltim(ct.CotisOmi)
	for i := range ct.TabOm {
		if err := rb.SetGtauom(i, ct.TabOm[i].Gtauom); err != nil {
			return err
		}
		if err := rb.SetMctom(i, ct.CotisOm[i]); err != nil {
			return err
		}
	}
	rb.SetTctom(t.om)
	rb.SetTcthfr(t.t3 + t.t8 + t.t9)

	rb.SetMfa300(ct.Fa300)
	rb.SetMfn300(ct.Fn300)
	rb.SetMfa800(ct.Fa800)
	rb.SetMfn800(ct.Fn800)
	rb.SetMfa900(ct.Fa900)
	rb.SetMfn900(ct.Fn900)

	feesTotal := ct.Fa300 + ct.Fn300 + ct.Fa800 + ct.Fn800 + ct.Fa900 + ct.Fn900
	rb.SetTctfra(feesTotal)
	// TCTDU is computed from the stored (truncated) RETOURB-TCTHFR, as the
	// source reads it back from RETOURB.
	tcthfr, _ := rb.Tcthfr()
	rb.SetTctdu(tcthfr + feesTotal)
	return nil
}

// contribution is COMPUTE x ROUNDED = base * rate / 100 into S9(10):
// rate has TauxScale decimals, rounding is half away from zero, and the
// result is truncated on the high-order side (no ON SIZE ERROR).
func contribution(base, rate int64) int64 {
	num := new(big.Int).Mul(big.NewInt(base), big.NewInt(rate))
	return truncSigned(roundQuo(num, TauxScale+2), amtDigits)
}

// fees computes FA = ROUNDED(total * arn) into S9(10), W-FRAIS =
// ROUNDED(total * frs) into S9(11) and FN = W-FRAIS - FA into S9(10).
func fees(total, arn, frs int64) (fa, fn int64) {
	fa = truncSigned(roundQuo(new(big.Int).Mul(big.NewInt(total), big.NewInt(arn)), feeScale), amtDigits)
	wFrais := truncSigned(roundQuo(new(big.Int).Mul(big.NewInt(total), big.NewInt(frs)), feeScale), totDigits)
	fn = truncSigned(wFrais-fa, amtDigits)
	return fa, fn
}

// roundQuo returns num / 10^scale rounded half away from zero. The result
// is reduced modulo 10^18 before conversion so it always fits an int64; the
// callers truncate further to the target picture.
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
