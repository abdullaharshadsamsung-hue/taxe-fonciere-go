package taudis

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	_ "modernc.org/sqlite" // pure-Go SQLite driver, registered as "sqlite"

	layout "migrated-app/internal/layout/taudis"
)

// DriverName is the database/sql driver name of modernc.org/sqlite.
const DriverName = "sqlite"

// Store is the keyed access to the TAUDIS table. It implements rates.Store.
type Store struct {
	db    *sql.DB
	owned bool
}

// Open creates the taudis table if needed on db and returns a store over
// it. The caller keeps ownership of db.
func Open(ctx context.Context, db *sql.DB) (*Store, error) {
	if db == nil {
		return nil, errors.New("taudis: nil database")
	}
	if _, err := db.ExecContext(ctx, Schema); err != nil {
		return nil, fmt.Errorf("taudis: create schema: %w", err)
	}
	return &Store{db: db}, nil
}

// OpenPath opens (creating if needed) the SQLite database file at path and
// returns a store that owns the connection (released by Close).
func OpenPath(ctx context.Context, path string) (*Store, error) {
	if path == "" {
		return nil, errors.New("taudis: empty database path")
	}
	db, err := sql.Open(DriverName, path)
	if err != nil {
		return nil, fmt.Errorf("taudis: open %s: %w", path, err)
	}
	db.SetMaxOpenConns(1)
	s, err := Open(ctx, db)
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	s.owned = true
	return s, nil
}

// Close releases the database when the store owns it.
func (s *Store) Close() error {
	if s == nil || !s.owned {
		return nil
	}
	return s.db.Close()
}

func splitKey(key [layout.EnteteLen]byte) (dir, com, ifp, per []byte) {
	return key[0:3], key[3:6], key[6:9], key[9:12]
}

// Read returns the 3000-byte record (key + suite) whose key equals key.
// found=false is the equivalent of a FILE STATUS other than '00'.
func (s *Store) Read(ctx context.Context, key [layout.EnteteLen]byte) ([]byte, bool, error) {
	dir, com, ifp, per := splitKey(key)
	var suite []byte
	err := s.db.QueryRowContext(ctx,
		`SELECT suite FROM taudis WHERE cle_dir = ? AND cle_com = ? AND cle_ifp = ? AND cle_per = ?`,
		dir, com, ifp, per).Scan(&suite)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("taudis: read: %w", err)
	}
	if len(suite) != SuiteLen {
		return nil, false, fmt.Errorf("taudis: stored suite is %d bytes, want %d", len(suite), SuiteLen)
	}
	rec := make([]byte, 0, layout.RecordLen)
	rec = append(rec, key[:]...)
	rec = append(rec, suite...)
	return rec, true, nil
}

// Put inserts or replaces one 3000-byte record, split into the 12-byte key
// and the 2988-byte suite.
func (s *Store) Put(ctx context.Context, rec []byte) error {
	return put(ctx, s.db, rec)
}

type execer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func put(ctx context.Context, x execer, rec []byte) error {
	if len(rec) != layout.RecordLen {
		return fmt.Errorf("taudis: record must be %d bytes, got %d", layout.RecordLen, len(rec))
	}
	var key [layout.EnteteLen]byte
	copy(key[:], rec[:layout.EnteteLen])
	dir, com, ifp, per := splitKey(key)
	suite := append([]byte(nil), rec[layout.EnteteLen:]...)
	_, err := x.ExecContext(ctx,
		`INSERT OR REPLACE INTO taudis (cle_dir, cle_com, cle_ifp, cle_per, suite) VALUES (?, ?, ?, ?, ?)`,
		append([]byte(nil), dir...), append([]byte(nil), com...),
		append([]byte(nil), ifp...), append([]byte(nil), per...), suite)
	if err != nil {
		return fmt.Errorf("taudis: write: %w", err)
	}
	return nil
}

// Load writes every record produced by next inside a single transaction.
// next returns io.EOF-style termination by returning (nil, nil). It reports
// the number of records written.
func (s *Store) Load(ctx context.Context, next func() ([]byte, error)) (int, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("taudis: begin: %w", err)
	}
	n := 0
	for {
		rec, err := next()
		if err != nil {
			_ = tx.Rollback()
			return n, err
		}
		if rec == nil {
			break
		}
		if err := put(ctx, tx, rec); err != nil {
			_ = tx.Rollback()
			return n, fmt.Errorf("record %d: %w", n+1, err)
		}
		n++
	}
	if err := tx.Commit(); err != nil {
		return n, fmt.Errorf("taudis: commit: %w", err)
	}
	return n, nil
}
