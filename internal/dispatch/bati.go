// Package dispatch holds the Go ports of the MAJIC2 year dispatchers
// (CTXTA3B for built property, ...). A dispatcher reads the tax year from
// the caller's 600-byte input record and hands the call to the matching
// year-specific calculator.
package dispatch

// This file is the Go port of CTXTA3B. CTXTA3B is only a switch on the
// tax year (E-ANNEE, bytes 2-5 of ENTREE). An incoming CRM of 99 means the
// caller is CROISIC and becomes PARM 'M'; any other value becomes PARM
// space. CRM is then reset to 00 and the year-specific EFITA3Bx program is
// CALLed with (ENTREE, SORTIE, CRM, RCM, PARM), all passed by reference.
//
// MIGRATION_NOTE: CTXTA3B creates no schema and does no file or database
// I/O, so this file needs no DDL.

import (
	"context"
	"errors"
	"fmt"

	"migrated-app/internal/calc/bati2018"
)

// RecordLen is the length of the ENTREE and SORTIE linkage records.
const RecordLen = 600

// CRCroisic is the CRM value a caller sends to ask for CROISIC mode (PARM 'M').
const CRCroisic = 99

// ErrCalculatorUnavailable is returned when the tax year maps to an
// EFITA3Bx program that has not been ported to Go. In COBOL this would be
// a failed static/dynamic CALL, which abends.
var ErrCalculatorUnavailable = errors.New("dispatch: year-specific calculator not available")

// BatiCalculator is a year-specific built-property calculator (EFITA3Bx)
// as seen by the dispatcher. *bati2018.Calculator implements it.
type BatiCalculator interface {
	Calculate(ctx context.Context, combatm []byte, parm string) (bati2018.Output, error)
}

// batiPrograms is the EVALUATE E-ANNEE table: year -> load module name.
// The module names repeat ten years apart (EFITA3B6 is used for 2006 and
// 2016, and so on); in the original it is the same load module, which only
// accepts the year it was last built for.
var batiPrograms = map[string]string{
	"2006": "EFITA3B6",
	"2007": "EFITA3B7",
	"2008": "EFITA3B8",
	"2009": "EFITA3B8",
	"2010": "EFITA3B0",
	"2011": "EFITA3B1",
	"2012": "EFITA3B2",
	"2013": "EFITA3B3",
	"2014": "EFITA3B4",
	"2015": "EFITA3B5",
	"2016": "EFITA3B6",
	"2017": "EFITA3B7",
	"2018": "EFITA3B8",
}

// BatiDispatcher is the CTXTA3B port. It routes each call to a calculator
// registered under an EFITA3Bx program name.
type BatiDispatcher struct {
	programs map[string]BatiCalculator
}

// NewBatiDispatcher builds the dispatcher. efita3b8 is the 2018 calculator
// (EFITA3B8); it receives the 2008, 2009 and 2018 calls, as in the source.
// It must not be nil.
func NewBatiDispatcher(efita3b8 BatiCalculator) (*BatiDispatcher, error) {
	if efita3b8 == nil {
		return nil, errors.New("dispatch: nil EFITA3B8 calculator")
	}
	return &BatiDispatcher{
		programs: map[string]BatiCalculator{"EFITA3B8": efita3b8},
	}, nil
}

// BatiYearProgram returns the EFITA3Bx program name CTXTA3B would CALL for
// the given 4-character year, and false when the year has no WHEN branch.
func BatiYearProgram(year string) (string, bool) {
	p, ok := batiPrograms[year]
	return p, ok
}

// Dispatch runs CTXTA3B. entree is ENTREE (only bytes 2-5, the year, are
// read here; the whole buffer is handed to the calculator). sortie is
// SORTIE: it is updated in place, just as BY REFERENCE storage would be,
// and must be at least RecordLen bytes long. crm and rcm are the incoming
// CRM / RCM; the CRM / RCM values after the call are returned.
//
// Outcomes:
//   - year not in the table (no WHEN OTHER): nothing is called, sortie is
//     not touched, CR 00 and RC unchanged.
//   - year mapped to a program that is not ported: CR 00, RC unchanged,
//     sortie untouched, and an error wrapping ErrCalculatorUnavailable.
//   - otherwise the calculator's RETOURM / CRM / RCM are copied back.
//     EFITA3B8 always moves RETOURB to RETOURM, even after a validation
//     error (for example 12/02 for 2008 and 2009).
func (d *BatiDispatcher) Dispatch(ctx context.Context, entree, sortie []byte, crm, rcm int) (cr, rc int, err error) {
	if len(sortie) < RecordLen {
		return crm, rcm, fmt.Errorf("dispatch: SORTIE is %d bytes, want %d", len(sortie), RecordLen)
	}

	// IF CRM = 99 MOVE 'M' TO PARM ELSE MOVE ' ' TO PARM. CRM is PIC 9(2),
	// so the comparison is numeric.
	parm := " "
	if crm == CRCroisic {
		parm = bati2018.ParmTaudis
	}
	cr = 0 // MOVE 00 TO CRM
	rc = rcm

	program, ok := BatiYearProgram(eAnnee(entree))
	if !ok {
		// END-EVALUATE with no WHEN OTHER: nothing is called.
		return cr, rc, nil
	}
	calc, ok := d.programs[program]
	if !ok {
		return cr, rc, fmt.Errorf("dispatch: CALL %s: %w", program, ErrCalculatorUnavailable)
	}

	out, err := calc.Calculate(ctx, entree, parm)
	if err != nil {
		return cr, rc, fmt.Errorf("dispatch: CALL %s: %w", program, err)
	}
	copy(sortie[:RecordLen], out.Retour[:])
	return out.CR, out.RC, nil
}

// eAnnee returns E-ANNEE (bytes 2-5 of ENTREE). If the buffer is short,
// the missing bytes are treated as spaces, which matches no WHEN branch.
func eAnnee(entree []byte) string {
	year := []byte("    ")
	if len(entree) > 1 {
		copy(year, entree[1:])
	}
	return string(year)
}
