// Package tfcalc is the public API of the taxe foncière 2018 calculators.
// The contract is the one of the COBOL entry points CTXTA3B / CTXTA3N: a
// 600-byte input record, a 600-byte output record and the CR / RC pair.
//
// Rates come from EFITAUX2 over the TAUDIS store (PARM 'M', CROISIC
// callers, incoming CR 99) and from FMSTAU2 otherwise. No IMS access exists,
// so FMSTAU2 is the default rates.UnavailableBaseTaux adapter, which reports
// CR 24 / RC 01 exactly like the callers' CALL ... ON EXCEPTION.
package tfcalc

import (
	"context"
	"errors"
	"fmt"

	"migrated-app/internal/calc/bati2018"
	"migrated-app/internal/calc/nonbati2018"
	"migrated-app/internal/dispatch"
	"migrated-app/internal/layout/taudis"
	"migrated-app/internal/layout/zes"
	"migrated-app/internal/rates"
)

// RecordLen is the length of the input and output records.
const RecordLen = dispatch.RecordLen

// CRCroisic is the incoming CR that selects CROISIC mode (PARM 'M').
const CRCroisic = dispatch.CRCroisic

// ErrCalculatorUnavailable is returned for years whose year-specific
// program has not been ported.
var ErrCalculatorUnavailable = dispatch.ErrCalculatorUnavailable

// Kind selects the dispatcher.
type Kind byte

const (
	// KindBati is the built-property dispatcher (CTXTA3B).
	KindBati Kind = 'B'
	// KindNonBati is the non-built-property dispatcher (CTXTA3N).
	KindNonBati Kind = 'N'
)

// Result is the outcome of one call: the 600-byte output record and CR/RC.
type Result struct {
	Out [RecordLen]byte
	CR  int
	RC  int
}

// Dispatcher routes calls to the bâti and non-bâti dispatchers. Like the
// shared COBOL load modules, a single EFITAUX2 / FMSTAU2 instance serves
// both calculators. Safe for concurrent use.
type Dispatcher struct {
	bati    *dispatch.BatiDispatcher
	nonBati *dispatch.NonBatiDispatcher
}

// New builds a dispatcher reading TAUDIS through store. FMSTAU2 is the
// Unavailable adapter (CR 24 / RC 01).
func New(store rates.Store) (*Dispatcher, error) {
	return NewWithBase(store, rates.UnavailableBaseTaux{})
}

// NewWithBase builds a dispatcher with an explicit FMSTAU2 implementation.
func NewWithBase(store rates.Store, base rates.BaseTauxFetcher) (*Dispatcher, error) {
	layouts, err := taudisLayouts()
	if err != nil {
		return nil, fmt.Errorf("tfcalc: TAUDIS layouts: %w", err)
	}
	taudisSrc, err := rates.NewTaudisSource(store, layouts)
	if err != nil {
		return nil, fmt.Errorf("tfcalc: EFITAUX2: %w", err)
	}
	if base == nil {
		base = rates.UnavailableBaseTaux{}
	}
	batiCalc, err := bati2018.NewCalculator(taudisSrc, base, batiDecoder{})
	if err != nil {
		return nil, err
	}
	nonBatiCalc, err := nonbati2018.NewCalculator(taudisSrc, base, nonBatiDecoder{})
	if err != nil {
		return nil, err
	}
	bd, err := dispatch.NewBatiDispatcher(batiCalc)
	if err != nil {
		return nil, err
	}
	nd, err := dispatch.NewNonBatiDispatcher(nonBatiCalc)
	if err != nil {
		return nil, err
	}
	return &Dispatcher{bati: bd, nonBati: nd}, nil
}

func newOut() [RecordLen]byte {
	var o [RecordLen]byte
	for i := range o {
		o[i] = ' '
	}
	return o
}

// Bati runs CTXTA3B. cr / rc are the incoming CRM / RCM (CRM 99 = CROISIC).
// The output record starts as spaces and is left unchanged when nothing is
// called.
func (d *Dispatcher) Bati(ctx context.Context, in []byte, cr, rc int) (Result, error) {
	res := Result{Out: newOut()}
	var err error
	res.CR, res.RC, err = d.bati.Dispatch(ctx, in, res.Out[:], cr, rc)
	return res, err
}

// NonBati runs CTXTA3N. cr / rc are the incoming CR / RC (CR 99 = CROISIC).
func (d *Dispatcher) NonBati(ctx context.Context, in []byte, cr, rc int) (Result, error) {
	res := Result{Out: newOut()}
	var err error
	res.CR, res.RC, err = d.nonBati.Dispatch(ctx, in, res.Out[:], cr, rc)
	return res, err
}

// Call is the byte-level adapter mirroring CALL 'CTXTA3B' / 'CTXTA3N'
// USING ENTREE SORTIE CR RC: out (at least RecordLen bytes) is updated in
// place and *cr / *rc are read then overwritten.
func (d *Dispatcher) Call(ctx context.Context, kind Kind, in, out []byte, cr, rc *int) error {
	if cr == nil || rc == nil {
		return errors.New("tfcalc: nil CR/RC")
	}
	var (
		c, r int
		err  error
	)
	switch kind {
	case KindBati:
		c, r, err = d.bati.Dispatch(ctx, in, out, *cr, *rc)
	case KindNonBati:
		c, r, err = d.nonBati.Dispatch(ctx, in, out, *cr, *rc)
	default:
		return fmt.Errorf("tfcalc: unknown kind %q", byte(kind))
	}
	*cr, *rc = c, r
	return err
}

// ---- rate decoders (ZES slot 2 = XBXTDDIR, 3 = XBXTDCOM, 4 = XBXTDSR) ----

type batiDecoder struct{}

func (batiDecoder) Decode(z *rates.ZES) bati2018.RateSet {
	dir := zes.DIR(z[1])
	com := zes.COM(z[2])
	sr := zes.SR(z[3])
	u := func(v uint64, _ bool) int64 { return int64(v) }
	return bati2018.RateSet{
		Pbbom: [bati2018.CotisOmCount]int64{
			u(com.Pbbomp()), u(com.Pbboma()), u(com.Pbbomb()),
			u(com.Pbbomc()), u(com.Pbbomd()), u(com.Pbbome()),
		},
		SrJan:       sr.JanRaw(),
		SrDepdir:    sr.Depdir(),
		SrCcocom:    sr.Ccocom(),
		SrCcoifp:    sr.Ccoifp(),
		TaucomB:     u(sr.TaucomB()),
		TaudepB:     u(dir.TaudepB()),
		Ptbtas:      u(dir.Ptbtas()),
		TausyndsfpB: u(sr.TausyndsfpB()),
		TaucudfpvnB: u(sr.TaucudfpvnB()),
		TautseB:     u(sr.TautseB()),
		Ptbtgp:      u(sr.Ptbtgp()),
		Ptbgem:      u(sr.Ptbgem()),
	}
}

type nonBatiDecoder struct{}

func (nonBatiDecoder) Decode(z *rates.ZES) nonbati2018.RateSet {
	dir := zes.DIR(z[1])
	sr := zes.SR(z[3])
	u := func(v uint64, _ bool) int64 { return int64(v) }
	return nonbati2018.RateSet{
		SrJan:        sr.JanRaw(),
		SrDepdir:     sr.Depdir(),
		SrCcocom:     sr.Ccocom(),
		SrCcoifp:     sr.Ccoifp(),
		TaucomNB:     u(sr.TaucomNB()),
		TausyndsfpNB: u(sr.TausyndsfpNB()),
		TaucudfpvnNB: u(sr.TaucudfpvnNB()),
		Pnbgem:       u(sr.Pnbgem()),
		TautseNB:     u(sr.TautseNB()),
		Pnbtgp:       u(sr.Pnbtgp()),
		Pnbtxa:       u(sr.Pnbtxa()),
		Pnbcaa:       u(sr.Pnbcaa()),
		TauchagrNB:   u(dir.TauchagrNB()),
		TausmagrNB:   u(dir.TausmagrNB()),
	}
}

// ---- EFITAUX2 layout adapters ----

// ifpRecord adapts *taudis.IFPTreso to rates.IFPRecord (CCOBOM read raw).
type ifpRecord struct{ *taudis.IFPTreso }

func (r ifpRecord) Ccobom() string { return r.CcobomRaw() }

// srSegment adapts *zes.SR to rates.SRSegment: MOVE CCOBOM (9(2)) TO
// LIGBENEFOM (9(2)) is a numeric move of the raw digits.
type srSegment struct{ *zes.SR }

func (s srSegment) SetLigbenefom(v string) {
	var n uint64
	for i := 0; i < len(v); i++ {
		d := uint64(v[i] & 0x0F)
		if d > 9 {
			d = 9
		}
		n = n*10 + d
	}
	s.SR.SetLigbenefom(n)
}

func taudisLayouts() (rates.Layouts, error) {
	if _, err := taudis.NewIFPTreso(); err != nil {
		return rates.Layouts{}, err
	}
	return rates.Layouts{
		ParseIFP: func(body []byte) (rates.IFPRecord, error) {
			r, err := taudis.ParseIFPTreso(body)
			if err != nil {
				return nil, err
			}
			return ifpRecord{&r}, nil
		},
		NewCOM: func() rates.COMSegment { c := zes.NewCOM(); return &c },
		NewSR:  func() rates.SRSegment { s := zes.NewSR(); return srSegment{&s} },
		NewIFP: func() rates.IFPRecord { r, _ := taudis.NewIFPTreso(); return ifpRecord{&r} },
	}, nil
}
