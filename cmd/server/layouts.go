package main

// Layout adapters injected into the rate programs (internal/rates).

import (
	"migrated-app/internal/layout/taudis"
	"migrated-app/internal/layout/zes"
	"migrated-app/internal/rates"
)

// ifpRecord adapts *taudis.IFPTreso to rates.IFPRecord (CCOBOM is read raw,
// as the MOVE to LIGBENEFOM is done on the two bytes).
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

// taudisLayouts returns the EFITAUX2 layout factories.
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

// baseTauxSegLen is the length of every GPHXTD00 segment image (the ZES
// EL-REC slot length; the segments are restituted as-is into ZES).
const baseTauxSegLen = rates.ZESSlotLen

type baseField struct {
	seg rates.SegmentID
	ref rates.FieldRef
}

// baseTauxFields describes the FMSTAU2 field positions.
//
// MIGRATION_NOTE: XB35 / XB40 / XB45 are restituted into ZES slots 2/3/4
// which the calculators read as XBXTDDIR / XBXTDCOM / XBXTDSR, so their
// fields use those layouts (internal/layout/zes). XB46 (IFP continuation)
// is assumed to share the XBXTDCOM layout of the fields it overlays. The
// label segments (XB37, XB41, XB50, XB51) copybooks were not available;
// their positions below are provisional and must be reviewed.
var baseTauxFields = buildBaseTauxFields()

func buildBaseTauxFields() map[rates.FieldID]baseField {
	const (
		X = rates.FieldAlnum
		Z = rates.FieldZoned
	)
	m := map[rates.FieldID]baseField{}
	add := func(seg rates.SegmentID, f rates.FieldID, off, n int, k rates.FieldKind) {
		m[f] = baseField{seg: seg, ref: rates.FieldRef{Off: off, Len: n, Kind: k}}
	}
	// OM rates, codes and JA deliberations: same order for XB40 and XB46.
	comOM := func(seg rates.SegmentID, first rates.FieldID) {
		for i, off := range []int{65, 75, 85, 95, 105, 272} { // PBBOMP..PBBOME
			add(seg, first+rates.FieldID(i), off, 10, Z)
		}
		add(seg, first+6, 115, 1, Z) // NTXOM (NBTAUXOM)
		for i, off := range []int{232, 236, 344, 322, 61} { // ACOSYN ACOGOM ACOBIM ACOTXA ACOGEM
			add(seg, first+7+rates.FieldID(i), off, 4, X)
		}
		for i := 0; i < 5; i++ {
			add(seg, first+12+rates.FieldID(i), 116+8*i, 4, Z) // JCOINS(i+1)
			add(seg, first+17+rates.FieldID(i), 120+8*i, 4, Z) // JCORIM(i+1)
		}
	}
	add(rates.SegXB40, rates.FXB40Ctygc, 228, 1, X)
	add(rates.SegXB40, rates.FXB40Dnucol, 229, 3, X)
	add(rates.SegXB40, rates.FXB40Ctysyn, 232, 1, X)
	add(rates.SegXB40, rates.FXB40Ccosyn, 233, 3, X)
	add(rates.SegXB40, rates.FXB40Gtovot, 15, 1, X)
	add(rates.SegXB40, rates.FXB40Gnidep, 226, 1, X)
	comOM(rates.SegXB40, rates.FXB40Pbbomp)
	comOM(rates.SegXB46, rates.FXB46Pbbomp)

	add(rates.SegXB35, rates.FXB35Ptbdep, 21, 10, Z) // TAUDEP-B
	add(rates.SegXB35, rates.FXB35Adtseg, 13, 387, X)
	add(rates.SegXB35Ori, rates.FXB35OriAdtseg, 13, 387, X)

	add(rates.SegXB45, rates.FXB45Ccddir, 375, 3, X)
	add(rates.SegXB45, rates.FXB45Acoets1, 342, 4, X)
	add(rates.SegXB45, rates.FXB45Acoets2, 346, 4, X)

	add(rates.SegXB37, rates.FXB37Vmailg, 200, 50, X)
	add(rates.SegXB37, rates.FXB37Vmaild, 250, 50, X)
	add(rates.SegXB50, rates.FXB50Aenr1, 13, 100, X)
	add(rates.SegXB50, rates.FXB50Aenr2, 113, 100, X)
	add(rates.SegXB50, rates.FXB50Aenr3, 213, 85, X)
	add(rates.SegXB50, rates.FXB50Vmailg, 298, 50, X)
	add(rates.SegXB50, rates.FXB50Vmaild, 348, 50, X)
	add(rates.SegXB41, rates.FXB41Vmailg, 200, 50, X)
	add(rates.SegXB41, rates.FXB41Vmaild, 250, 50, X)
	add(rates.SegXB51, rates.FXB51Vmailg, 200, 50, X)
	add(rates.SegXB51, rates.FXB51Vmaild, 250, 50, X)
	return m
}

// baseTauxLayout implements rates.BaseTauxLayout.
type baseTauxLayout struct{}

// NewSegment returns an INITIALIZEd image: spaces, known numerics zero.
func (baseTauxLayout) NewSegment(id rates.SegmentID) []byte {
	seg := make([]byte, baseTauxSegLen)
	for i := range seg {
		seg[i] = ' '
	}
	for _, f := range baseTauxFields {
		if f.seg == id && f.ref.Kind == rates.FieldZoned {
			for i := f.ref.Off; i < f.ref.Off+f.ref.Len; i++ {
				seg[i] = '0'
			}
		}
	}
	return seg
}

func (baseTauxLayout) Field(id rates.FieldID) rates.FieldRef { return baseTauxFields[id].ref }
