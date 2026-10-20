package main

// PostgreSQL-backed replacements for the two mainframe data stores used by
// the rate programs:
//   - the TAUDIS KSDS read by EFITAUX2 (rates.Store), table taudis_records,
//     keyed by the 12-byte TAUDIS primary key, holding 3000-byte records;
//   - the GPHXTD00 IMS database read by FMSTAU2 (rates.PCBProvider /
//     rates.SegmentReader), table gphxtd00_segments, keyed by the 13-byte
//     concatenated key used in the qualified SSA, holding the segment image.

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/lib/pq"

	"migrated-app/internal/layout/taudis"
	"migrated-app/internal/rates"
)

const schemaDDL = `
CREATE TABLE IF NOT EXISTS taudis_records (
    rec_key BYTEA PRIMARY KEY,
    record  BYTEA NOT NULL
);
CREATE TABLE IF NOT EXISTS gphxtd00_segments (
    ssa_key BYTEA PRIMARY KEY,
    segment BYTEA NOT NULL
);`

// openDB connects to PostgreSQL (retrying while the database starts) and
// makes sure the rate tables exist.
func openDB(ctx context.Context, url string) (*sql.DB, error) {
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	var pingErr error
	for i := 0; i < 30; i++ {
		pctx, cancel := context.WithTimeout(ctx, 3*time.Second)
		pingErr = db.PingContext(pctx)
		cancel()
		if pingErr == nil {
			break
		}
		select {
		case <-ctx.Done():
			_ = db.Close()
			return nil, ctx.Err()
		case <-time.After(time.Second):
		}
	}
	if pingErr != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping database: %w", pingErr)
	}
	if _, err := db.ExecContext(ctx, schemaDDL); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("create schema: %w", err)
	}
	return db, nil
}

// pgTaudisStore is the TAUDIS keyed READ (EFITAUX2).
type pgTaudisStore struct{ db *sql.DB }

func (s pgTaudisStore) Read(ctx context.Context, key [taudis.EnteteLen]byte) ([]byte, bool, error) {
	var rec []byte
	err := s.db.QueryRowContext(ctx,
		`SELECT record FROM taudis_records WHERE rec_key = $1`, key[:]).Scan(&rec)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return rec, true, nil
}

// pgPCBProvider replaces CALL 'FMSPCB' for the GPHXTD00 database.
type pgPCBProvider struct{ db *sql.DB }

// pcbUnknownRC is the RETURN-CODE reported for an unknown PCB name.
const pcbUnknownRC = 8

func (p pgPCBProvider) PCB(_ context.Context, name string) (rates.SegmentReader, int, error) {
	if name != rates.BaseTauxPCBName {
		return nil, pcbUnknownRC, nil
	}
	return pgSegmentReader{db: p.db}, 0, nil
}

// pgSegmentReader is the DL/I GU on GPHXTD00.
type pgSegmentReader struct{ db *sql.DB }

func (r pgSegmentReader) GU(ctx context.Context, ssaKey [rates.BaseTauxKeyLen]byte, segment []byte) (string, error) {
	var data []byte
	err := r.db.QueryRowContext(ctx,
		`SELECT segment FROM gphxtd00_segments WHERE ssa_key = $1`, ssaKey[:]).Scan(&data)
	if errors.Is(err, sql.ErrNoRows) {
		return "GE", nil // I/O area left untouched, as IMS does
	}
	if err != nil {
		return "", err
	}
	copy(segment, data)
	return "  ", nil
}
