// Package taudis is the SQLite-backed replacement of the TAUDIS VSAM KSDS
// read by EFITAUX2. A TAUDIS record (3000 bytes) is stored split into its
// 12-byte primary key — kept as the four 3-byte BLOB key columns of the
// T800 header (DEPDIR, CCOCOM, CCOIFP, CCPPER) — and the 2988-byte suite
// (the T84C / T84D / T84R body).
package taudis

// SuiteLen is the length of the record suite (body) stored in the table.
const SuiteLen = 2988

// KeyPartLen is the length of one of the four key columns.
const KeyPartLen = 3

// Schema is the DDL of the taudis table. Keys are BLOBs so that LOW-VALUE
// (0x00) filled key parts compare byte-exactly, as in the KSDS.
const Schema = `
CREATE TABLE IF NOT EXISTS taudis (
    cle_dir BLOB NOT NULL CHECK (length(cle_dir) = 3),
    cle_com BLOB NOT NULL CHECK (length(cle_com) = 3),
    cle_ifp BLOB NOT NULL CHECK (length(cle_ifp) = 3),
    cle_per BLOB NOT NULL CHECK (length(cle_per) = 3),
    suite   BLOB NOT NULL CHECK (length(suite) = 2988),
    PRIMARY KEY (cle_dir, cle_com, cle_ifp, cle_per)
) WITHOUT ROWID;`
