// Command tfcalc is the batch parity runner of the taxe foncière 2018
// calculators. It reads fixed-width input records and writes, for each of
// them, the 600-byte output record followed by CR and RC.
//
// Input record layout (one record per line, or a continuous stream of
// 603-byte records):
//
//	KIND    X(1)   'B' = bâti (CTXTA3B), 'N' = non bâti (CTXTA3N)
//	CALLER  9(2)   incoming CR: 99 = CROISIC (PARM 'M', EFITAUX2 over
//	               TAUDIS), anything else = MAJIC (FMSTAU2, Unavailable
//	               adapter: CR 24 / RC 01)
//	ENTREE  X(600) input record
//
// Output record layout (one per line):
//
//	SORTIE  X(600) output record
//	CR      9(2)
//	RC      9(2)
//
// The SQLite TAUDIS database is given by -db or TFCALC_DB.
package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"

	"migrated-app/internal/config"
	"migrated-app/internal/taudis"
	"migrated-app/pkg/tfcalc"
)

const (
	kindLen   = 1
	callerLen = 2
	inLen     = kindLen + callerLen + tfcalc.RecordLen
)

func main() {
	cfg := config.Register(flag.CommandLine)
	inPath := flag.String("in", "-", "input file of fixed-width records ('-' = stdin)")
	outPath := flag.String("out", "-", "output file ('-' = stdout)")
	flag.Parse()

	if err := run(context.Background(), cfg.DBPath, *inPath, *outPath); err != nil {
		fmt.Fprintln(os.Stderr, "tfcalc:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, dbPath, inPath, outPath string) error {
	store, err := taudis.OpenPath(ctx, dbPath)
	if err != nil {
		return err
	}
	defer store.Close()

	disp, err := tfcalc.New(store)
	if err != nil {
		return err
	}

	in := io.Reader(os.Stdin)
	if inPath != "-" {
		f, err := os.Open(inPath)
		if err != nil {
			return err
		}
		defer f.Close()
		in = f
	}
	out := io.Writer(os.Stdout)
	if outPath != "-" {
		f, err := os.Create(outPath)
		if err != nil {
			return err
		}
		defer f.Close()
		out = f
	}
	w := bufio.NewWriter(out)
	defer w.Flush()

	r := bufio.NewReader(in)
	n := 0
	failed := false
	for {
		rec, err := readRecord(r)
		if err == io.EOF {
			break
		}
		n++
		if err != nil {
			return fmt.Errorf("record %d: %w", n, err)
		}
		kind := tfcalc.Kind(rec[0])
		cr, err := strconv.Atoi(string(rec[kindLen : kindLen+callerLen]))
		if err != nil {
			return fmt.Errorf("record %d: caller %q is not numeric", n, rec[kindLen:kindLen+callerLen])
		}
		rc := 0
		sortie := make([]byte, tfcalc.RecordLen)
		for i := range sortie {
			sortie[i] = ' '
		}
		if cerr := disp.Call(ctx, kind, rec[kindLen+callerLen:], sortie, &cr, &rc); cerr != nil {
			if !errors.Is(cerr, tfcalc.ErrCalculatorUnavailable) {
				failed = true
			}
			fmt.Fprintf(os.Stderr, "tfcalc: record %d: %v\n", n, cerr)
		}
		if _, err := w.Write(sortie); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "%02d%02d\n", cr%100, rc%100); err != nil {
			return err
		}
	}
	if failed {
		return errors.New("one or more records failed")
	}
	return nil
}

// readRecord reads one input record. Lines are accepted (trailing CR/LF
// stripped, short lines padded with spaces); a stream without separators
// is consumed inLen bytes at a time.
func readRecord(r *bufio.Reader) ([]byte, error) {
	buf := make([]byte, 0, inLen)
	for len(buf) < inLen {
		b, err := r.ReadByte()
		if err == io.EOF {
			if len(buf) == 0 {
				return nil, io.EOF
			}
			break
		}
		if err != nil {
			return nil, err
		}
		if b == '\n' {
			if len(buf) == 0 {
				continue // blank line
			}
			break
		}
		buf = append(buf, b)
	}
	// Swallow a line terminator right after a full record.
	if len(buf) == inLen {
		if b, err := r.Peek(1); err == nil && (b[0] == '\r' || b[0] == '\n') {
			_, _ = r.ReadByte()
			if b[0] == '\r' {
				if b2, err := r.Peek(1); err == nil && b2[0] == '\n' {
					_, _ = r.ReadByte()
				}
			}
		}
	}
	if n := len(buf); n > 0 && buf[n-1] == '\r' {
		buf = buf[:n-1]
	}
	if len(buf) < kindLen+callerLen {
		return nil, fmt.Errorf("record too short (%d bytes)", len(buf))
	}
	for len(buf) < inLen {
		buf = append(buf, ' ')
	}
	return buf, nil
}
