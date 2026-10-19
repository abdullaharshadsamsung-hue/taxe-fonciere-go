package zes

import (
	"fmt"
)

// SRLen is the total length in bytes of a ZES SR segment
// (AIDFIC 13 + ADTSEG 387 = 400).
const SRLen = 400

// Scaling factors of the implied-decimal numeric fields of the SR segment.
const (
	// SRRateScale is the scaling factor of the PIC 9(4)V9(6) rate fields:
	// accessors return the rate multiplied by 1 000 000 (six implied
	// decimals), e.g. 12.345678 -> 12345678.
	SRRateScale = 1_000_000
	// SRShareScale is the scaling factor of the PIC 9(3)V9(2) syndicate
	// repartition elements PBASY1 / PNBPSY / POMSY1: accessors return the
	// value multiplied by 100 (12,50 -> 1250).
	SRShareScale = 100
)

// OCCURS counts of the SR segment.
const (
	// SRSyndCount is the OCCURS count of the syndicate repartition tables
	// TPBASY, TPNBSY and TPOMSY.
	SRSyndCount = 10
	// SRTseCount is the OCCURS count of the parallel ACOETS / DNUTSE tables.
	SRTseCount = 2
)

// Group / area lengths of the SR segment.
const (
	// SRAidficLen is the length of the AIDFIC file-identifier group.
	SRAidficLen = 13
	// SRAdtsegLen is the length of the ADTSEG segment-body group.
	SRAdtsegLen = 387
	// SRFiller1Len is the length of the first named 'X'-FILLER (after PNBGEM).
	SRFiller1Len = 50
	// SRFiller2Len is the length of the second named 'X'-FILLER (after DNUTSE).
	SRFiller2Len = 21
)

// Zero-based byte offsets of the SR fields inside the 400-byte segment.
//
//	AIDFIC          X(13)        0..12   (COBOL positions   1..13)
//	  JAN             9(4)         0..3
//	  DEPDIR          X(3)         4..6  = CODEP X(2) + CODIR X(1)
//	  CCOCOM          X(3)         7..9
//	  CCOIFP          X(3)        10..12
//	ADTSEG          X(387)      13..399  (COBOL positions  14..400)
//	  CCOENR          X(2)        13..14
//	  LIGBENEFOM      9(2)        15..16
//	  TPBASY  PBASY1 OCCURS 10 9(3)V9(2)   17..66
//	  TPNBSY  PNBPSY OCCURS 10 9(3)V9(2)   67..116
//	  TPOMSY  POMSY1 OCCURS 10 9(3)V9(2)  117..166
//	  TAUCOM-B        9(4)V9(6)  167..176
//	  TAUSYNDSFP-B    9(4)V9(6)  177..186
//	  TAUCUDFPVN-B    9(4)V9(6)  187..196
//	  TAUTSE-B        9(4)V9(6)  197..206
//	  PNBTXA          9(4)V9(6)  207..216
//	  TAUCOM-NB       9(4)V9(6)  217..226
//	  TAUSYNDSFP-NB   9(4)V9(6)  227..236
//	  TAUCUDFPVN-NB   9(4)V9(6)  237..246
//	  TAUTSE-NB       9(4)V9(6)  247..256
//	  PNBCAA          9(4)V9(6)  257..266
//	  CCOEXC          X(2)       267..268
//	  GIMOMI          X(1)       269
//	  CCOBIM          9(2)       270..271
//	  PTBGEM          9(4)V9(6)  272..281
//	  PNBGEM          9(4)V9(6)  282..291
//	  FILLER          X(50)      292..341
//	  ACOETS OCCURS 2 X(4)       342..349
//	  DNUTSE OCCURS 2 9(2)       350..353
//	  FILLER          X(21)      354..374
//	  CCDDIR          X(3)       375..377 = CO2DEP X(2) + CCODIR X(1)
//	  PTBTGP          9(4)V9(6)  378..387
//	  PNBTGP          9(4)V9(6)  388..397
//	  DNUURC          X(2)       398..399
const (
	srOffAidfic      = 0
	srOffJan         = 0
	srOffDepdir      = 4
	srOffCodep       = 4
	srOffCodir       = 6
	srOffCcocom      = 7
	srOffCcoifp      = 10
	srOffAdtseg      = 13
	srOffCcoenr      = 13
	srOffLigbenefom  = 15
	srOffTpbasy      = 17
	srOffTpnbsy      = 67
	srOffTpomsy      = 117
	srOffTaucomB     = 167
	srOffTausyndsfpB = 177
	srOffTaucudfpvnB = 187
	srOffTautseB     = 197
	srOffPnbtxa      = 207
	srOffTaucomNB    = 217
	srOffTausyndsfpN = 227
	srOffTaucudfpvnN = 237
	srOffTautseNB    = 247
	srOffPnbcaa      = 257
	srOffCcoexc      = 267
	srOffGimomi      = 269
	srOffCcobim      = 270
	srOffPtbgem      = 272
	srOffPnbgem      = 282
	srOffFiller1     = 292
	srOffAcoets      = 342
	srOffDnutse      = 350
	srOffFiller2     = 354
	srOffCcddir      = 375
	srOffCo2dep      = 375
	srOffCcodir      = 377
	srOffPtbtgp      = 378
	srOffPnbtgp      = 388
	srOffDnuurc      = 398

	srRateLen   = 10 // PIC 9(4)V9(6)
	srShareLen  = 5  // PIC 9(3)V9(2)
	srTableLen  = SRSyndCount * srShareLen
	srYearLen   = 4 // PIC 9(4)
	srCode2Len  = 2 // PIC 9(2)
	srAcoetsLen = 4 // PIC X(4)
	srDnutseLen = 2 // PIC 9(2)
)

// Compile-time layout checks: consecutive areas must be contiguous and the
// last field must end exactly at SRLen.
var (
	_ = [1]struct{}{}[srOffTpbasy+srTableLen-srOffTpnbsy]
	_ = [1]struct{}{}[srOffTpnbsy+srTableLen-srOffTpomsy]
	_ = [1]struct{}{}[srOffTpomsy+srTableLen-srOffTaucomB]
	_ = [1]struct{}{}[srOffPnbgem+srRateLen-srOffFiller1]
	_ = [1]struct{}{}[srOffFiller1+SRFiller1Len-srOffAcoets]
	_ = [1]struct{}{}[srOffAcoets+SRTseCount*srAcoetsLen-srOffDnutse]
	_ = [1]struct{}{}[srOffDnutse+SRTseCount*srDnutseLen-srOffFiller2]
	_ = [1]struct{}{}[srOffFiller2+SRFiller2Len-srOffCcddir]
	_ = [1]struct{}{}[srOffDnuurc+2-SRLen]
)

// SR is the ZES SR segment (copybook XBXTDSR): a fixed 400-byte MAJIC2
// "taux" record holding one commune's local rates — commune, syndicate,
// grouping of communes, TSE, GEMAPI, TEOM/TEOMI and TSE Grand Paris rates —
// plus the repartition tables of syndicate products.
//
// JAN / DEPDIR / CCOCOM / CCOIFP (the AIDFIC group) form the rate cache key
// used by the rate programs.
//
// Fields are exposed as accessors over the raw storage, so group views
// (AIDFIC, DEPDIR, ADTSEG, CCDDIR, TPBASY...) and their elementary fields
// share bytes exactly like COBOL storage; bytes never touched by a setter
// (including both FILLER areas) round-trip unchanged.
//
// MIGRATION_NOTE: the copybook uses an 'X' placeholder prefix substituted
// by COPY ... REPLACING. In Go each inclusion is simply a distinct SR
// value. 'X'-FILLER becomes a real data name after substitution, so both
// filler areas are exposed (Filler1/Filler2) to preserve their content.
//
// MIGRATION_NOTE: encoding is ASCII (the byte-exact layouts of this project
// all assume host data already converted from EBCDIC). All numerics are
// unsigned zoned decimal USAGE DISPLAY (no COMP-3, no sign) and are returned
// as scaled integers (see SRRateScale and SRShareScale) together with the
// result of the COBOL NUMERIC class test; the arithmetic primitives decide
// what to do with invalid data. Setters take already-scaled integers and
// truncate high-order digits like a COBOL numeric MOVE. OCCURS tables are
// exposed through 0-based indexed accessors that return an error on an
// out-of-range subscript. There is no REDEFINES in this copybook.
type SR [SRLen]byte

// NewSR returns a segment initialised to spaces, as a COBOL MOVE SPACES on
// the alphanumeric group would.
func NewSR() SR {
	var s SR
	for i := range s {
		s[i] = ' '
	}
	return s
}

// ParseSR copies a 400-byte buffer into a new SR segment. It fails if b is
// not exactly SRLen bytes long.
func ParseSR(b []byte) (SR, error) {
	var s SR
	if len(b) != SRLen {
		return s, fmt.Errorf("zes: SR segment must be %d bytes, got %d", SRLen, len(b))
	}
	copy(s[:], b)
	return s, nil
}

// Bytes returns a copy of the raw 400 bytes of the segment.
func (s *SR) Bytes() []byte {
	out := make([]byte, SRLen)
	copy(out, s[:])
	return out
}

// ---- low-level helpers ----

func (s *SR) get(off, n int) string { return string(s[off : off+n]) }

// put performs a COBOL alphanumeric MOVE into a field: left-justified,
// truncated on the right, padded with spaces.
func (s *SR) put(off, n int, v string) {
	for i := 0; i < n; i++ {
		if i < len(v) {
			s[off+i] = v[i]
		} else {
			s[off+i] = ' '
		}
	}
}

// isNumeric applies the COBOL NUMERIC class test for an unsigned
// zoned-decimal field (every byte a digit).
func (s *SR) isNumeric(off, n int) bool {
	for _, c := range s[off : off+n] {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// unsigned decodes an unsigned zoned-decimal field as its digit-string
// value (implied decimals included, i.e. already scaled). ok is false when
// the field is not NUMERIC; the value is then built from the low nibble of
// each byte (clamped to 9), as IBM COBOL arithmetic on invalid zoned data
// would, and must be treated as unreliable.
func (s *SR) unsigned(off, n int) (value uint64, ok bool) {
	ok = s.isNumeric(off, n)
	for _, c := range s[off : off+n] {
		dg := uint64(c & 0x0F)
		if dg > 9 {
			dg = 9 // MIGRATION_NOTE: invalid nibble; clamp to keep result in range
		}
		value = value*10 + dg
	}
	return value, ok
}

// setUnsigned performs a COBOL numeric MOVE of an already-scaled value into
// an unsigned zoned-decimal field: high-order digits beyond the field width
// are truncated.
func (s *SR) setUnsigned(off, n int, v uint64) {
	for i := n - 1; i >= 0; i-- {
		s[off+i] = byte('0' + v%10)
		v /= 10
	}
}

func srCheckIndex(field string, i, count int) error {
	if i < 0 || i >= count {
		return fmt.Errorf("zes: SR %s subscript %d out of range [0,%d)", field, i, count)
	}
	return nil
}

// ---- AIDFIC: file identifier (positions 1-13) ----

// Aidfic returns the 13-byte AIDFIC group (MAJIC2 file identifier).
func (s *SR) Aidfic() string { return s.get(srOffAidfic, SRAidficLen) }

// SetAidfic moves v into the 13-byte AIDFIC group.
func (s *SR) SetAidfic(v string) { s.put(srOffAidfic, SRAidficLen, v) }

// JanRaw returns the raw 4 bytes of JAN PIC 9(4) (year).
func (s *SR) JanRaw() string { return s.get(srOffJan, srYearLen) }

// JanIsNumeric reports whether JAN passes the COBOL NUMERIC class test.
func (s *SR) JanIsNumeric() bool { return s.isNumeric(srOffJan, srYearLen) }

// Jan returns JAN PIC 9(4) (year); ok is false when the field is not NUMERIC.
func (s *SR) Jan() (uint64, bool) { return s.unsigned(srOffJan, srYearLen) }

// SetJan performs a COBOL numeric MOVE into JAN PIC 9(4).
func (s *SR) SetJan(v uint64) { s.setUnsigned(srOffJan, srYearLen, v) }

// Depdir returns DEPDIR X(3): département + direction group.
func (s *SR) Depdir() string { return s.get(srOffDepdir, 3) }

// SetDepdir moves v into DEPDIR X(3).
func (s *SR) SetDepdir(v string) { s.put(srOffDepdir, 3, v) }

// Codep returns CODEP X(2): département code.
func (s *SR) Codep() string { return s.get(srOffCodep, 2) }

// SetCodep moves v into CODEP X(2).
func (s *SR) SetCodep(v string) { s.put(srOffCodep, 2, v) }

// Codir returns CODIR X(1): direction code.
func (s *SR) Codir() string { return s.get(srOffCodir, 1) }

// SetCodir moves v into CODIR X(1).
func (s *SR) SetCodir(v string) { s.put(srOffCodir, 1, v) }

// Ccocom returns CCOCOM X(3): MAJIC2 commune code.
func (s *SR) Ccocom() string { return s.get(srOffCcocom, 3) }

// SetCcocom moves v into CCOCOM X(3).
func (s *SR) SetCcocom(v string) { s.put(srOffCcocom, 3, v) }

// Ccoifp returns CCOIFP X(3): absorbed-commune code.
func (s *SR) Ccoifp() string { return s.get(srOffCcoifp, 3) }

// SetCcoifp moves v into CCOIFP X(3).
func (s *SR) SetCcoifp(v string) { s.put(srOffCcoifp, 3, v) }

// ---- ADTSEG: segment body (positions 14-400) ----

// Adtseg returns the 387-byte ADTSEG group (segment body).
func (s *SR) Adtseg() string { return s.get(srOffAdtseg, SRAdtsegLen) }

// SetAdtseg moves v into the 387-byte ADTSEG group.
func (s *SR) SetAdtseg(v string) { s.put(srOffAdtseg, SRAdtsegLen, v) }

// Ccoenr returns CCOENR X(2): record code.
func (s *SR) Ccoenr() string { return s.get(srOffCcoenr, 2) }

// SetCcoenr moves v into CCOENR X(2).
func (s *SR) SetCcoenr(v string) { s.put(srOffCcoenr, 2, v) }

// Ligbenefom returns LIGBENEFOM PIC 9(2), the household-waste (OM)
// beneficiary collectivity code; ok is false when the field is not NUMERIC.
func (s *SR) Ligbenefom() (uint64, bool) { return s.unsigned(srOffLigbenefom, srCode2Len) }

// SetLigbenefom moves v into LIGBENEFOM PIC 9(2).
func (s *SR) SetLigbenefom(v uint64) { s.setUnsigned(srOffLigbenefom, srCode2Len, v) }

// ---- syndicate product repartition tables, PIC 9(3)V9(2) OCCURS 10 ----

func (s *SR) share(field string, base, i int) (uint64, bool, error) {
	if err := srCheckIndex(field, i, SRSyndCount); err != nil {
		return 0, false, err
	}
	v, ok := s.unsigned(base+i*srShareLen, srShareLen)
	return v, ok, nil
}

func (s *SR) setShare(field string, base, i int, v uint64) error {
	if err := srCheckIndex(field, i, SRSyndCount); err != nil {
		return err
	}
	s.setUnsigned(base+i*srShareLen, srShareLen, v)
	return nil
}

// Tpbasy returns the raw 50-byte TPBASY group (built syndicate repartition table).
func (s *SR) Tpbasy() string { return s.get(srOffTpbasy, srTableLen) }

// SetTpbasy moves v into the 50-byte TPBASY group.
func (s *SR) SetTpbasy(v string) { s.put(srOffTpbasy, srTableLen, v) }

// Pbasy1 returns PBASY1(i+1) (0-based i), a built syndicate product
// repartition element scaled by SRShareScale; ok is false when the element
// is not NUMERIC.
func (s *SR) Pbasy1(i int) (uint64, bool, error) { return s.share("PBASY1", srOffTpbasy, i) }

// SetPbasy1 moves an SRShareScale-scaled value into PBASY1(i+1) (0-based i).
func (s *SR) SetPbasy1(i int, v uint64) error { return s.setShare("PBASY1", srOffTpbasy, i, v) }

// Tpnbsy returns the raw 50-byte TPNBSY group (non-built syndicate
// repartition table).
func (s *SR) Tpnbsy() string { return s.get(srOffTpnbsy, srTableLen) }

// SetTpnbsy moves v into the 50-byte TPNBSY group.
func (s *SR) SetTpnbsy(v string) { s.put(srOffTpnbsy, srTableLen, v) }

// Pnbpsy returns PNBPSY(i+1) (0-based i), a non-built syndicate product
// repartition element scaled by SRShareScale.
func (s *SR) Pnbpsy(i int) (uint64, bool, error) { return s.share("PNBPSY", srOffTpnbsy, i) }

// SetPnbpsy moves an SRShareScale-scaled value into PNBPSY(i+1) (0-based i).
func (s *SR) SetPnbpsy(i int, v uint64) error { return s.setShare("PNBPSY", srOffTpnbsy, i, v) }

// Tpomsy returns the raw 50-byte TPOMSY group (household-waste syndicate
// repartition table).
func (s *SR) Tpomsy() string { return s.get(srOffTpomsy, srTableLen) }

// SetTpomsy moves v into the 50-byte TPOMSY group.
func (s *SR) SetTpomsy(v string) { s.put(srOffTpomsy, srTableLen, v) }

// Pomsy1 returns POMSY1(i+1) (0-based i), a household-waste syndicate
// product repartition element scaled by SRShareScale.
func (s *SR) Pomsy1(i int) (uint64, bool, error) { return s.share("POMSY1", srOffTpomsy, i) }

// SetPomsy1 moves an SRShareScale-scaled value into POMSY1(i+1) (0-based i).
func (s *SR) SetPomsy1(i int, v uint64) error { return s.setShare("POMSY1", srOffTpomsy, i, v) }

// ---- rates PIC 9(4)V9(6), scaled by SRRateScale ----

// TaucomB returns TAUCOM-B, the commune built rate, scaled by SRRateScale;
// ok is false when the field is not NUMERIC.
func (s *SR) TaucomB() (uint64, bool) { return s.unsigned(srOffTaucomB, srRateLen) }

// SetTaucomB moves an SRRateScale-scaled value into TAUCOM-B.
func (s *SR) SetTaucomB(v uint64) { s.setUnsigned(srOffTaucomB, srRateLen, v) }

// TausyndsfpB returns TAUSYNDSFP-B, the syndicate built rate, scaled by
// SRRateScale.
func (s *SR) TausyndsfpB() (uint64, bool) { return s.unsigned(srOffTausyndsfpB, srRateLen) }

// SetTausyndsfpB moves an SRRateScale-scaled value into TAUSYNDSFP-B.
func (s *SR) SetTausyndsfpB(v uint64) { s.setUnsigned(srOffTausyndsfpB, srRateLen, v) }

// TaucudfpvnB returns TAUCUDFPVN-B, the grouping-of-communes built rate,
// scaled by SRRateScale.
func (s *SR) TaucudfpvnB() (uint64, bool) { return s.unsigned(srOffTaucudfpvnB, srRateLen) }

// SetTaucudfpvnB moves an SRRateScale-scaled value into TAUCUDFPVN-B.
func (s *SR) SetTaucudfpvnB(v uint64) { s.setUnsigned(srOffTaucudfpvnB, srRateLen, v) }

// TautseB returns TAUTSE-B, the built special equipment tax (TSE) rate,
// scaled by SRRateScale.
func (s *SR) TautseB() (uint64, bool) { return s.unsigned(srOffTautseB, srRateLen) }

// SetTautseB moves an SRRateScale-scaled value into TAUTSE-B.
func (s *SR) SetTautseB(v uint64) { s.setUnsigned(srOffTautseB, srRateLen, v) }

// Pnbtxa returns PNBTXA, the additional tax rate, scaled by SRRateScale.
func (s *SR) Pnbtxa() (uint64, bool) { return s.unsigned(srOffPnbtxa, srRateLen) }

// SetPnbtxa moves an SRRateScale-scaled value into PNBTXA.
func (s *SR) SetPnbtxa(v uint64) { s.setUnsigned(srOffPnbtxa, srRateLen, v) }

// TaucomNB returns TAUCOM-NB, the commune non-built rate, scaled by
// SRRateScale.
func (s *SR) TaucomNB() (uint64, bool) { return s.unsigned(srOffTaucomNB, srRateLen) }

// SetTaucomNB moves an SRRateScale-scaled value into TAUCOM-NB.
func (s *SR) SetTaucomNB(v uint64) { s.setUnsigned(srOffTaucomNB, srRateLen, v) }

// TausyndsfpNB returns TAUSYNDSFP-NB, the syndicate non-built rate, scaled
// by SRRateScale.
func (s *SR) TausyndsfpNB() (uint64, bool) { return s.unsigned(srOffTausyndsfpN, srRateLen) }

// SetTausyndsfpNB moves an SRRateScale-scaled value into TAUSYNDSFP-NB.
func (s *SR) SetTausyndsfpNB(v uint64) { s.setUnsigned(srOffTausyndsfpN, srRateLen, v) }

// TaucudfpvnNB returns TAUCUDFPVN-NB, the urban-community non-built rate,
// scaled by SRRateScale.
func (s *SR) TaucudfpvnNB() (uint64, bool) { return s.unsigned(srOffTaucudfpvnN, srRateLen) }

// SetTaucudfpvnNB moves an SRRateScale-scaled value into TAUCUDFPVN-NB.
func (s *SR) SetTaucudfpvnNB(v uint64) { s.setUnsigned(srOffTaucudfpvnN, srRateLen, v) }

// TautseNB returns TAUTSE-NB, the non-built TSE rate (outside
// Île-de-France), scaled by SRRateScale.
func (s *SR) TautseNB() (uint64, bool) { return s.unsigned(srOffTautseNB, srRateLen) }

// SetTautseNB moves an SRRateScale-scaled value into TAUTSE-NB.
func (s *SR) SetTautseNB(v uint64) { s.setUnsigned(srOffTautseNB, srRateLen, v) }

// Pnbcaa returns PNBCAA, the non-built C.A.A.A. rate, scaled by SRRateScale.
func (s *SR) Pnbcaa() (uint64, bool) { return s.unsigned(srOffPnbcaa, srRateLen) }

// SetPnbcaa moves an SRRateScale-scaled value into PNBCAA.
func (s *SR) SetPnbcaa(v uint64) { s.setUnsigned(srOffPnbcaa, srRateLen, v) }

// ---- exclusion / TEOMI ----

// Ccoexc returns CCOEXC X(2): exclusion code.
func (s *SR) Ccoexc() string { return s.get(srOffCcoexc, 2) }

// SetCcoexc moves v into CCOEXC X(2).
func (s *SR) SetCcoexc(v string) { s.put(srOffCcoexc, 2, v) }

// Gimomi returns GIMOMI X(1): TEOMI taxation indicator.
func (s *SR) Gimomi() string { return s.get(srOffGimomi, 1) }

// SetGimomi moves v into GIMOMI X(1).
func (s *SR) SetGimomi(v string) { s.put(srOffGimomi, 1, v) }

// Ccobim returns CCOBIM PIC 9(2), the TEOMI beneficiary collectivity code;
// ok is false when the field is not NUMERIC.
func (s *SR) Ccobim() (uint64, bool) { return s.unsigned(srOffCcobim, srCode2Len) }

// SetCcobim moves v into CCOBIM PIC 9(2).
func (s *SR) SetCcobim(v uint64) { s.setUnsigned(srOffCcobim, srCode2Len, v) }

// ---- GEMAPI ----

// Ptbgem returns PTBGEM, the built GEMAPI rate, scaled by SRRateScale.
func (s *SR) Ptbgem() (uint64, bool) { return s.unsigned(srOffPtbgem, srRateLen) }

// SetPtbgem moves an SRRateScale-scaled value into PTBGEM.
func (s *SR) SetPtbgem(v uint64) { s.setUnsigned(srOffPtbgem, srRateLen, v) }

// Pnbgem returns PNBGEM, the non-built GEMAPI rate, scaled by SRRateScale.
func (s *SR) Pnbgem() (uint64, bool) { return s.unsigned(srOffPnbgem, srRateLen) }

// SetPnbgem moves an SRRateScale-scaled value into PNBGEM.
func (s *SR) SetPnbgem(v uint64) { s.setUnsigned(srOffPnbgem, srRateLen, v) }

// ---- fillers ----

// Filler1 returns the 50-byte named 'X'-FILLER area following PNBGEM.
func (s *SR) Filler1() string { return s.get(srOffFiller1, SRFiller1Len) }

// SetFiller1 moves v into the 50-byte 'X'-FILLER area following PNBGEM.
func (s *SR) SetFiller1(v string) { s.put(srOffFiller1, SRFiller1Len, v) }

// Filler2 returns the 21-byte named 'X'-FILLER area following DNUTSE.
func (s *SR) Filler2() string { return s.get(srOffFiller2, SRFiller2Len) }

// SetFiller2 moves v into the 21-byte 'X'-FILLER area following DNUTSE.
func (s *SR) SetFiller2(v string) { s.put(srOffFiller2, SRFiller2Len, v) }

// ---- TSE beneficiaries (parallel tables ACOETS / DNUTSE) ----

// Acoets returns ACOETS(i+1) X(4) (0-based i), the TSE beneficiary body
// code. It is parallel to Dnutse(i).
func (s *SR) Acoets(i int) (string, error) {
	if err := srCheckIndex("ACOETS", i, SRTseCount); err != nil {
		return "", err
	}
	return s.get(srOffAcoets+i*srAcoetsLen, srAcoetsLen), nil
}

// SetAcoets moves v into ACOETS(i+1) X(4) (0-based i).
func (s *SR) SetAcoets(i int, v string) error {
	if err := srCheckIndex("ACOETS", i, SRTseCount); err != nil {
		return err
	}
	s.put(srOffAcoets+i*srAcoetsLen, srAcoetsLen, v)
	return nil
}

// Dnutse returns DNUTSE(i+1) PIC 9(2) (0-based i), the new-TSE item
// number paired with Acoets(i); ok is false when the element is not
// NUMERIC.
func (s *SR) Dnutse(i int) (uint64, bool, error) {
	if err := srCheckIndex("DNUTSE", i, SRTseCount); err != nil {
		return 0, false, err
	}
	v, ok := s.unsigned(srOffDnutse+i*srDnutseLen, srDnutseLen)
	return v, ok, nil
}

// SetDnutse moves v into DNUTSE(i+1) PIC 9(2) (0-based i).
func (s *SR) SetDnutse(i int, v uint64) error {
	if err := srCheckIndex("DNUTSE", i, SRTseCount); err != nil {
		return err
	}
	s.setUnsigned(srOffDnutse+i*srDnutseLen, srDnutseLen, v)
	return nil
}

// ---- CCDDIR: direction code group ----

// Ccddir returns CCDDIR X(3): direction code group (CO2DEP + CCODIR).
func (s *SR) Ccddir() string { return s.get(srOffCcddir, 3) }

// SetCcddir moves v into CCDDIR X(3).
func (s *SR) SetCcddir(v string) { s.put(srOffCcddir, 3, v) }

// Co2dep returns CO2DEP X(2): département code.
func (s *SR) Co2dep() string { return s.get(srOffCo2dep, 2) }

// SetCo2dep moves v into CO2DEP X(2).
func (s *SR) SetCo2dep(v string) { s.put(srOffCo2dep, 2, v) }

// Ccodir returns CCODIR X(1): direction code.
func (s *SR) Ccodir() string { return s.get(srOffCcodir, 1) }

// SetCcodir moves v into CCODIR X(1).
func (s *SR) SetCcodir(v string) { s.put(srOffCcodir, 1, v) }

// ---- TSE Grand Paris ----

// Ptbtgp returns PTBTGP, the built TSE Grand Paris rate, scaled by
// SRRateScale.
func (s *SR) Ptbtgp() (uint64, bool) { return s.unsigned(srOffPtbtgp, srRateLen) }

// SetPtbtgp moves an SRRateScale-scaled value into PTBTGP.
func (s *SR) SetPtbtgp(v uint64) { s.setUnsigned(srOffPtbtgp, srRateLen, v) }

// Pnbtgp returns PNBTGP, the non-built TSE Grand Paris rate, scaled by
// SRRateScale.
func (s *SR) Pnbtgp() (uint64, bool) { return s.unsigned(srOffPnbtgp, srRateLen) }

// SetPnbtgp moves an SRRateScale-scaled value into PNBTGP.
func (s *SR) SetPnbtgp(v uint64) { s.setUnsigned(srOffPnbtgp, srRateLen, v) }

// Dnuurc returns DNUURC X(2): CDIF regrouping unit number (UR).
func (s *SR) Dnuurc() string { return s.get(srOffDnuurc, 2) }

// SetDnuurc moves v into DNUURC X(2).
func (s *SR) SetDnuurc(v string) { s.put(srOffDnuurc, 2, v) }
