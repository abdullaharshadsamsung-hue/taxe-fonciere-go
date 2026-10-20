// Command taudisload loads a VSAM unload of the TAUDIS KSDS (fixed
// 3000-byte records, no separators) into the SQLite taudis table. Each
// record is stored split into its 12-byte key and its 2988-byte suite.
//
// The SQLite database is given by -db or TFCALC_DB.
package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"migrated-app/internal/config"
	layout "migrated-app/internal/layout/taudis"
	"migrated-app/internal/taudis"
)

func main() {
	cfg := config.Register(flag.CommandLine)
	inPath := flag.String("in", "-", "VSAM unload file of 3000-byte records ('-' = stdin)")
	flag.Parse()

	n, err := run(context.Background(), cfg.DBPath, *inPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "taudisload:", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "taudisload: %d records loaded into %s\n", n, cfg.DBPath)
}

func run(ctx context.Context, dbPath, inPath string) (int, error) {
	in := io.Reader(os.Stdin)
	if inPath != "-" {
		f, err := os.Open(inPath)
		if err != nil {
			return 0, err
		}
		defer f.Close()
		in = f
	}
	r := bufio.NewReader(in)

	store, err := taudis.OpenPath(ctx, dbPath)
	if err != nil {
		return 0, err
	}
	defer store.Close()

	next := func() ([]byte, error) {
		rec := make([]byte, layout.RecordLen)
		_, err := io.ReadFull(r, rec)
		if err == io.EOF {
			return nil, nil
		}
		if errors.Is(err, io.ErrUnexpectedEOF) {
			return nil, fmt.Errorf("truncated trailing record (want %d bytes)", layout.RecordLen)
		}
		if err != nil {
			return nil, err
		}
		return rec, nil
	}
	return store.Load(ctx, next)
}
