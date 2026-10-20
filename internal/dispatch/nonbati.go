package dispatch

// This file is the Go port of CTXTA3N, the entry point of the non-built
// property (taxe foncière non bâtie) calculator. CTXTA3N is only a switch
// on the tax year (ANNEE, bytes 2-5 of ENTREE). An incoming CR of 99 means
// the caller is CROISIC and becomes PARM 'M'; any other value becomes PARM
// space. CR is then reset to 00 and the year-specific EFITA3Nx program is
// CALLed with (ENTREE, SORTIE, CR, RC, PARM), all passed by reference.
//
// MIGRATION_NOTE: CTXTA3N creates no schema and does no file or database
// I/O, so this file needs no DDL.

import (
	"context"
	"errors"
	"fmt"

	"migrated-app/internal/calc/nonbati2018"
)

// NonBatiCalculator is a year-specific non-built-property calculator
// (EFITA3Nx) as seen by the dispatcher. *nonbati2018.Calculator
// implements it.
type NonBatiCalculator interface {
	Calculate(ctx context.Context, comnonbam []byte, parm string) (nonbati2018.Output, error)
}

// nonBatiPrograms is the EVALUATE ANNEE table: year -> load module name.
// ANNEE is PIC X(4) and the WHEN operands are numeric literals, so COBOL
// compares them as the alphanumeric strings "2010".."2018".
var nonBatiPrograms = map[string]string{
	"2010": "EFITA3N0",
	"2011": "EFITA3N1",
	"2012": "EFITA3N2",
	"2013": "EFITA3N3",
	"2014": "EFITA3N4",
	"2015": "EFITA3N5",
	"2016": "EFITA3N6",
	"2017": "EFITA3N7",
	"2018": "EFITA3N8",
}

// NonBatiDispatcher is the CTXTA3N port. It routes each call to a
// calculator registered under an EFITA3Nx program name.
type NonBatiDispatcher struct {
	programs map[string]NonBatiCalculator
}

// NewNonBatiDispatcher builds the dispatcher. efita3n8 is the 2018
// calculator (EFITA3N8) and must not be nil. The 2010-2017 programs
// (EFITA3N0..EFITA3N7) were not ported; calls for those years return
// ErrCalculatorUnavailable.
func NewNonBatiDispatcher(efita3n8 NonBatiCalculator) (*NonBatiDispatcher, error) {
	if efita3n8 == nil {
		return nil, errors.New("dispatch: nil EFITA3N8 calculator")
	}
	return &NonBatiDispatcher{
		programs: map[string]NonBatiCalculator{"EFITA3N8": efita3n8},
	}, nil
}

// NonBatiYearProgram returns the EFITA3Nx program name CTXTA3N would CALL
// for the given 4-character year, and false when the year has no WHEN
// branch.
func NonBatiYearProgram(year string) (string, bool) {
	p, ok := nonBatiPrograms[year]
	return p, ok
}

// Dispatch runs CTXTA3N. entree is ENTREE (only bytes 2-5, the year, are
// read here; the whole buffer goes to the calculator). sortie is SORTIE:
// it is updated in place, as BY REFERENCE storage would be, and must be at
// least RecordLen bytes long. cr and rc are the incoming CR / RC; the
// values after the call are returned.
//
// Outcomes:
//   - year not in the table (no WHEN OTHER): nothing is called, sortie is
//     not touched, CR 00 and RC unchanged.
//   - year mapped to a program that is not ported (2010-2017): CR 00, RC
//     unchanged, sortie untouched, and an error wrapping
//     ErrCalculatorUnavailable (a failed CALL abends in COBOL).
//   - otherwise the calculator's RETOURNB / CR / RC are copied back.
func (d *NonBatiDispatcher) Dispatch(ctx context.Context, entree, sortie []byte, cr, rc int) (int, int, error) {
	if len(sortie) < RecordLen {
		return cr, rc, fmt.Errorf("dispatch: SORTIE is %d bytes, want %d", len(sortie), RecordLen)
	}

	// IF CR = 99 MOVE 'M' TO PARM ELSE MOVE ' ' TO PARM. CR is PIC 99, so
	// the comparison is numeric.
	parm := " "
	if cr == CRCroisic {
		parm = nonbati2018.ParmTaudis
	}
	cr = 0 // MOVE 00 TO CR

	program, ok := NonBatiYearProgram(eAnnee(entree))
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
