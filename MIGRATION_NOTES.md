# Migration Notes

**Model self-assessed confidence:** 93% (not a measured result)  
**Build at time of writing:** passed  
**Target unit tests:** passed  
**Behavior compared with the original:** not run

The final recommendation is in the pull request description.

---

## What was migrated

- `src/T800.cpy` → `internal/layout/taudis/t800.go` (90% confidence)
- `src/T84D.cpy` → `internal/layout/taudis/t84d.go` (89% confidence)
- `src/T84C.cpy` → `internal/layout/taudis/t84c.go` (88% confidence)
- `src/T84R.cpy` → `internal/layout/taudis/t84r.go` (85% confidence) ⚠️ needs review
- `src/XBXTDAN.cpy` → `internal/layout/zes/an.go` (94% confidence)
- `src/XBXTDDIR.cpy` → `internal/layout/zes/dir.go` (91% confidence)
- `src/XBXTDCOM.cpy` → `internal/layout/zes/com.go` (92% confidence)
- `src/XBXTDSR.cpy` → `internal/layout/zes/sr.go` (89% confidence)
- `src/XCOMBAT.cpy` → `internal/layout/linkage/combat.go` (88% confidence)
- `src/XCOMNBA.cpy` → `internal/layout/linkage/comnba.go` (91% confidence)
- `src/XRETB.cpy` → `internal/layout/linkage/retb.go` (88% confidence)
- `src/XRETNB.cpy` → `internal/layout/linkage/retnb.go` (89% confidence)
- `src/XCOTB.cpy` → `internal/calc/bati2018/cotis.go` (79% confidence) ⚠️ needs review
- `src/XCOTNB.cpy` → `internal/calc/nonbati2018/cotis.go` (89% confidence)
- `src/EFITAUX2.cob` → `internal/rates/taudis.go` (74% confidence) ⚠️ needs review
- `src/FMSTAU2.cob` → `internal/rates/basetaux.go` (78% confidence) ⚠️ needs review
- `src/EFITA3B8.cob` → `internal/calc/bati2018/calc.go` (72% confidence) ⚠️ needs review
- `src/EFITA3N8.cob` → `internal/calc/nonbati2018/calc.go` (75% confidence) ⚠️ needs review
- `src/CTXTA3B.cob` → `internal/dispatch/bati.go` (77% confidence) ⚠️ needs review
- `src/CTXTA3N.cob` → `internal/dispatch/nonbati.go` (82% confidence) ⚠️ needs review

## Not translated, and why

**Not migrated, per the agreed plan: EPCI/OM/TSE-group article never referenced by any program.**
- `src/T84G.cpy`

**Not migrated, per the agreed plan: BASEB is dead working storage: filled but never read or returned. The Go code omits the writes.**
- `src/XBASEB.cpy`

**Not migrated, per the agreed plan: BASENB is dead working storage: filled but never read or returned. The Go code omits the writes.**
- `src/XBASNB.cpy`

## Files requiring manual review

These files were migrated but scored below the confidence threshold.
Review them carefully before merging.

### `src/T84R.cpy`
Confidence: 85%

### `src/XCOTB.cpy`
Confidence: 79%
Issues:
  - [warning] Positive signed fields stored with a plain last digit are re-encoded as IBM overpunch ({/A-I), and ASCII p-y negatives are re-encoded as }/J-R, so parse->encode isn't byte-identical for those inputs.
  - [warning] Values wider than their field are silently truncated on the high-order side (COBOL MOVE semantics) instead of being flagged.

### `src/EFITAUX2.cob`
Confidence: 74%

### `src/FMSTAU2.cob`
Confidence: 78%

### `src/EFITA3B8.cob`
Confidence: 72%
Issues:
  - [info] Treated as doing nothing; that only holds if the slot-4 rate layout is exactly 400 bytes. If it is shorter, the source blanks the rest of slot 4.
  - [info] The port assumes entry 7 is the 3% assessment fee field and is zero at that point. This depends on the contribution-record copybook layout.

### `src/EFITA3N8.cob`
Confidence: 75%
Issues:
  - [warning] When the provider returns a Go error, the Go code stops and returns the error. COBOL would instead continue with CR 24/RC 01 (ON EXCEPTION). A nil provider is still mapped to 24/01.

### `src/CTXTA3B.cob`
Confidence: 77%
Issues:
  - [warning] Only EFITA3B8 is registered. Years mapped to B0-B7 return ErrCalculatorUnavailable, much like a COBOL CALL abend when the module is missing. In production, those modules do exist.
  - [info] When a calculator returns an error, the dispatcher sends back CR=00 and the original RCM, and SORTIE is not written. In COBOL the callee would have written these values in place.

### `src/CTXTA3N.cob`
Confidence: 82%
Issues:
  - [warning] EFITA3N0..EFITA3N7 are not registered, so these years return ErrCalculatorUnavailable instead of running a calculation.
