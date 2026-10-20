# DGFiP Taxe Foncière: Go migration

This is a Go port of the DGFiP *taxe foncière* 2018 calculation subprograms, which were originally written in COBOL. It covers:

- the **bâti** calculator (`CTXTA3B` → `EFITA3B8`)
- the **non bâti** calculator (`CTXTA3N` → `EFITA3N8`)
- the two rate-lookup programs they call:
  - `EFITAUX2`, which reads the TAUDIS VSAM KSDS
  - `FMSTAU2`, which reads the GPHXTD00 IMS database

The original code had only called subprograms and no main program. This repository therefore adds three entry points:

| Command | Purpose | Storage |
|---|---|---|
| `cmd/server` | HTTP API that exposes the bâti and non-bâti dispatchers | PostgreSQL (`DATABASE_URL`) |
| `cmd/tfcalc` | Batch parity runner. Reads fixed-width records and writes the 600-byte output record plus CR/RC. | SQLite (`-db` / `TFCALC_DB`) |
| `cmd/taudisload` | Loads a VSAM unload of the TAUDIS KSDS into the SQLite store | SQLite (`-db` / `TFCALC_DB`) |

> **Status**
> - Migration progress: 20 of 23 source modules migrated. The other 3 were intentionally not migrated (see [Known limitations](#known-limitations)).
> - Build: **passed**.
> - Go unit tests (`go test ./...`): **passed**.
> - Comparison with the original COBOL behavior: **not run**. No output has been checked against the original yet.
> - The migration tool reported a confidence of 93%. This is the model's own estimate, not a measured result.

---

## Tech stack

- **Go**. `go.mod` declares `go 1.23`. The migration was built and tested with the toolchain from the `golang:1.27-alpine` image. The module name is `migrated-app`.
- **HTTP router:** `github.com/go-chi/chi/v5` v5.1.0
- **Logging:** `github.com/rs/zerolog` v1.33.0
- **PostgreSQL driver:** `github.com/lib/pq` v1.10.9, used by `cmd/server`
- **SQLite driver:** `modernc.org/sqlite` v1.29.6. This is a pure-Go driver (no cgo), used by `cmd/tfcalc` and `cmd/taudisload`.

## Prerequisites

- The Go toolchain. The migration used Go 1.27 (`golang:1.27-alpine`). `go.mod` requires at least 1.23.
- A reachable **PostgreSQL** instance, needed for `cmd/server`.
- No external database is needed for `cmd/tfcalc` and `cmd/taudisload`. They create or open a local SQLite file.

---

## Getting started

Run all commands from the root of a fresh clone.

### 1. Install dependencies and build the server

```sh
go mod download
mkdir -p bin
go build -o bin/server ./cmd/server
```

### 2. Configure the environment

```sh
export DATABASE_URL='postgres://<user>:<password>@<host>:5432/<dbname>?sslmode=disable'
export PORT=8080            # optional, defaults to 8080
```

`JWT_SECRET` was listed by the migration run, but no code in this repository reads it and the router has no authentication middleware. You do not need to set it.

### 3. Database setup

You do not need to run a separate migration step. At startup, `cmd/server`:

- connects to `DATABASE_URL`, retrying the ping up to 30 times, once per second
- runs this DDL:

```sql
CREATE TABLE IF NOT EXISTS taudis_records (
    rec_key BYTEA PRIMARY KEY,   -- 12-byte TAUDIS primary key
    record  BYTEA NOT NULL       -- 3000-byte TAUDIS record
);
CREATE TABLE IF NOT EXISTS gphxtd00_segments (
    ssa_key BYTEA PRIMARY KEY,   -- 13-byte concatenated SSA key
    segment BYTEA NOT NULL       -- segment image
);
```

The repository has **no loader for these PostgreSQL tables**. Until you populate them, rate lookups behave as follows:

- TAUDIS lookups behave as "record not found".
- GPHXTD00 lookups return status `GE`.

### 4. Run the server

```sh
./bin/server
```

The server listens on `:8080`, or on `:$PORT` if `PORT` is set. It shuts down gracefully on SIGINT or SIGTERM, with a 10-second timeout.

### A first request to try

No request has been verified against the original system yet. Use the health endpoint as a smoke check. It is a new endpoint with no COBOL counterpart.

```sh
curl http://localhost:8080/healthz
# ok
```

### HTTP API

| Method | Path | Body | Response |
|---|---|---|---|
| `GET` | `/healthz` | — | `ok` |
| `POST` | `/api/bati/dispatch` | `{"entree": "<600-char record>", "cr": <int>, "rc": <int>}` | `{"sortie": "<600 chars>", "cr": <int>, "rc": <int>, "error": "<optional>"}` |
| `POST` | `/api/nonbati/dispatch` | same as above | same as above |

- **Input:** the `cr` field carries the incoming caller code, as in the COBOL linkage.
- **Output:** `sortie` is initialised to 600 spaces before the call.
- **Status codes:**

| Status | Meaning |
|---|---|
| `200` | Success |
| `400` | Invalid JSON |
| `422` | The calculator for the requested year is not available (`ErrCalculatorUnavailable`) |
| `500` | Any other error |

### Batch tools (SQLite)

`cmd/tfcalc` and `cmd/taudisload` share one option for the database location:

- Flag: `-db`
- Environment variable: `TFCALC_DB`
- Default: `taudis.db` (in the current directory)
- Precedence: flag, then environment variable, then default.

The SQLite table is created automatically when it does not exist.

`taudisload` reads a stream of fixed 3000-byte records with no separators. It reads from the file given by `-in`, or from stdin (`-in -`, the default). It writes all records in a single transaction. A truncated trailing record is an error.

`tfcalc` reads input records from `-in` and writes output to `-out`. Both default to `-` (stdin and stdout).

Input record format, one record per line or a continuous stream of 603-byte records:

| Field | PIC | Meaning |
|---|---|---|
| KIND | `X(1)` | `B` = bâti (CTXTA3B), `N` = non bâti (CTXTA3N) |
| CALLER | `9(2)` | `99` = CROISIC (PARM `M`, EFITAUX2 over TAUDIS). Any other value = MAJIC, which returns CR 24 / RC 01 (see the migration notes). |
| ENTREE | `X(600)` | input record |

Output: one line per record, containing `SORTIE X(600)`, then `CR 9(2)`, then `RC 9(2)`.

The process exits with status 1 if any record fails, except for records that fail only with `ErrCalculatorUnavailable`. Per-record errors are written to stderr.

Example invocation, built from the flags in the source code. This invocation was not run during the migration.

```sh
go run ./cmd/taudisload -db <sqlite-file> -in <taudis-unload-file>
go run ./cmd/tfcalc -db <sqlite-file> -in <input-records-file> -out <output-file>
```

---

## Database

| | Original (COBOL) | Migrated (Go) |
|---|---|---|
| TAUDIS rates (read by EFITAUX2) | VSAM KSDS, opened by the host program as an EXTERNAL file | SQLite table `taudis` (`cmd/tfcalc`, `cmd/taudisload`) **or** PostgreSQL table `taudis_records` (`cmd/server`) |
| Base taux (read by FMSTAU2) | IMS database GPHXTD00, via `FMSPCB` / `CBLTDLI` | PostgreSQL table `gphxtd00_segments` (`cmd/server` only). The batch runner uses an "Unavailable" adapter instead. |

Why the storage changed:

- **Injected store instead of host file:** the migration plan replaces the VSAM file opened by the host program with a store that is passed in. The host program no longer exists.
- **BLOB/BYTEA columns:** keys may contain LOW-VALUE (`0x00`) bytes and data may be EBCDIC. Binary columns keep both byte-for-byte.
- **SQLite key layout:** the SQLite `taudis` table splits the 12-byte key into four 3-byte BLOB columns: `cle_dir`, `cle_com`, `cle_ifp`, `cle_per`. These match the T800 header fields DEPDIR, CCOCOM, CCOIFP and CCPPER. The 2988-byte record body is stored separately in `suite`. The table is declared `WITHOUT ROWID`.

---

## Running tests

```sh
go test ./...
```

This command passed during the migration. Test files exist for these packages:

- `internal/calc/bati2018` (`calc_test.go`, `cotis_test.go`)
- `internal/calc/nonbati2018` (`cotis_test.go` only; `calc.go` has no test file)
- `internal/dispatch`
- `internal/layout/linkage`
- `internal/layout/taudis`
- `internal/layout/zes`
- `internal/rates`

No test file exists for these packages:

- `cmd/*`
- `internal/config`
- `internal/taudis`
- `pkg/tfcalc`

---

## Environment variables

| Variable | Used by | Required | Default | Description |
|---|---|---|---|---|
| `DATABASE_URL` | `cmd/server` | Yes. The server exits if it is unset. | — | PostgreSQL connection URL (`lib/pq` format) |
| `PORT` | `cmd/server` | No | `8080` | HTTP listen port |
| `JWT_SECRET` | — (not read by any code) | No | — | Listed by the migration run, but nothing in the repository reads it and the router has no authentication middleware. |
| `TFCALC_DB` | `cmd/tfcalc`, `cmd/taudisload` | No | `taudis.db` | SQLite database file. Overridden by `-db`. |

---

## Architecture overview

```
cmd/
  server/         HTTP API (chi). main.go: config + graceful shutdown;
                  router.go: wiring + /api/{bati,nonbati}/dispatch;
                  store.go: PostgreSQL stores for TAUDIS and GPHXTD00;
                  layouts.go: layout adapters + FMSTAU2 field positions
  tfcalc/         batch parity runner (fixed-width in → 600-byte out + CR/RC)
  taudisload/     VSAM unload → SQLite loader
pkg/
  tfcalc/         public dispatcher used by cmd/tfcalc (Kind, RecordLen,
                  ErrCalculatorUnavailable)
internal/
  config/         shared -db flag / TFCALC_DB
  dispatch/       bati.go / nonbati.go: entry dispatchers (CTXTA3B / CTXTA3N)
  calc/
    bati2018/     bâti 2018 calculator and contributions (EFITA3B8)
    nonbati2018/  non bâti 2018 calculator and contributions (EFITA3N8)
  rates/
    taudis.go     EFITAUX2: rate lookup over the TAUDIS store
    basetaux.go   FMSTAU2: rate lookup over the GPHXTD00 segments
  taudis/         SQLite-backed TAUDIS store (schema.go, store.go)
  layout/
    linkage/      linkage records: combat, comnba, retb, retnb
    taudis/       TAUDIS record layouts: t800 (key header), t84c, t84d, t84r
    zes/          ZES return-area segments: an, com, dir, sr
```

### Request flow

1. A dispatcher (`internal/dispatch`) receives the 600-byte input record together with CR/RC.
2. The dispatcher selects the calculator for the year. Only the 2018 calculators exist.
3. The calculator gets its rates from:
   - `rates.NewTaudisSource` (EFITAUX2) for CROISIC callers
   - `rates.NewBaseTauxSource` (FMSTAU2) for MAJIC callers
4. The ZES return area is decoded into a `RateSet`. This is done by `batiDecoder` and `nonBatiDecoder` in `cmd/server/router.go`.
5. The calculator writes the 600-byte output record.

As with the shared COBOL load modules, one instance of each rate program serves both calculators.

---

## Migration notes

Each change from the original COBOL code, with the reason for it:

1. **TAUDIS VSAM KSDS replaced by a SQLite table.**
   - The table is accessed through the pure-Go `modernc.org/sqlite` driver, and its DDL runs at startup.
   - Keys are stored as BLOBs so that LOW-VALUE (`0x00`) key bytes and EBCDIC data stay intact.
   - *Why:* the plan replaces the EXTERNAL VSAM file, which the host program opened, with an injected store. Keys contain LOW-VALUE and data may be EBCDIC, so binary columns are used.
   - *Note:* `cmd/server` uses PostgreSQL (`taudis_records`) for the same purpose instead of SQLite.

2. **New database configuration.**
   - The SQLite file is set with `-db` or `TFCALC_DB`, and defaults to `taudis.db`.
   - *Why:* the plan uses flags plus environment variables for configuration. The VSAM file was opened by the host program, which no longer exists (inferred).

3. **MAJIC callers (PARM not `M`) get no rates from the IMS base taux.**
   - FMSTAU2 is replaced by the "Unavailable" adapter, so every such call returns **CR 24 / RC 01**.
   - *Why:* the XB* copybooks, the IMS DBD and `FMSPCB` / `CBLTDLI` are missing, so the segment layouts cannot be derived. CR 24 / RC 01 is what COBOL's `ON EXCEPTION` produces when FMSTAU2 cannot be loaded.
   - *Note:* this describes `cmd/tfcalc`. `cmd/server` instead wires a PostgreSQL-backed FMSTAU2 (`gphxtd00_segments`) whose field positions are partly provisional (see [Manual review required](#manual-review-required)).

4. **Calculators for unmigrated years return `ErrCalculatorUnavailable`.**
   - This applies to bâti 2006, 2007 and 2010–2017, and to non-bâti 2010–2017.
   - The output record is left untouched; the run unit does not abend.
   - *Why:* `EFITA3B0`–`B7` and `EFITA3N0`–`N7` are not in the codebase. Go errors are reserved for conditions that abended in COBOL.

5. **New batch command `cmd/tfcalc`.**
   - It reads records made of kind, caller and 600 input bytes. It writes the 600-byte output plus CR and RC.
   - *Why:* it is a batch parity runner for diffing against golden vectors captured from production. The original code had only called subprograms and no main program.

6. **New command `cmd/taudisload`.**
   - It loads a VSAM unload of 3000-byte TAUDIS records into the SQLite table.
   - *Why:* the SQLite store that replaces the VSAM KSDS needs a way to be populated.

7. **Infrastructure errors that are collapsed into CR/RC codes are now logged with their cause.**
   - *Why:* the COBOL code masks them silently.

8. **Dead code omitted.**
   - The writes to the `BASEB` and `BASENB` working storage are not reproduced.
   - The `T84G` article is not translated.
   - *Why:* see [Known limitations](#known-limitations).

---

## Known limitations

### Behavioral parity has not been tested

- No output of the Go code has been compared with the COBOL programs.
- No golden vectors have been run through `cmd/tfcalc`.
- No HTTP request has been verified against the original.

### Source files not migrated (per the agreed plan)

- `T84G`: the EPCI/OM/TSE-group article is never referenced by any program.
- `XBASEB`: `BASEB` is dead working storage. It is filled but never read or returned.
- `XBASNB`: `BASENB` is dead working storage. It is filled but never read or returned.

### Only the 2018 calculators exist

All other years return `ErrCalculatorUnavailable`:

- over HTTP, as status 422
- in `tfcalc`, as a message on stderr, with the output record left as spaces

### MAJIC path (FMSTAU2)

- In `cmd/tfcalc`, every MAJIC call returns CR 24 / RC 01.
- In `cmd/server`, field positions for the label segments XB37, XB41, XB50 and XB51 are provisional, because their copybooks were not available.
- XB46 is assumed to share the XBXTDCOM layout.

### Two storage backends that do not share data

- `cmd/taudisload` loads data **only into SQLite**. It cannot populate the PostgreSQL tables used by `cmd/server`.
- The repository contains no loader for `taudis_records` or `gphxtd00_segments`.

### Test coverage gaps

- `internal/calc/nonbati2018/calc.go` has no unit test.
- `pkg/tfcalc`, `internal/taudis`, `internal/config` and the `cmd/*` commands have no unit tests.

---

## Manual review required

### Low-confidence modules

These modules were flagged as low-confidence by the migration. Verify their Go counterparts against the COBOL sources:

| Original | Go counterpart |
|---|---|
| `T84R` | `internal/layout/taudis/t84r.go` |
| `XCOTB` | bâti contribution logic, `internal/calc/bati2018/cotis.go` |
| `EFITAUX2` | `internal/rates/taudis.go`, `cmd/server/layouts.go` (`taudisLayouts`) |
| `FMSTAU2` | `internal/rates/basetaux.go`, `cmd/server/layouts.go` (`baseTauxFields`) |
| `EFITA3B8` | `internal/calc/bati2018/` |
| `EFITA3N8` | `internal/calc/nonbati2018/` |
| `CTXTA3B` | `internal/dispatch/bati.go` |
| `CTXTA3N` | `internal/dispatch/nonbati.go` |

### Other items to check

- **FMSTAU2 field map** (`cmd/server/layouts.go`, `buildBaseTauxFields`): the code carries a `MIGRATION_NOTE`. Positions for XB37, XB41, XB50 and XB51 are provisional, and the XB46 layout is assumed.
- **CCOBOM → LIGBENEFOM move** (`cmd/server/layouts.go`, `srSegment.SetLigbenefom`): this is a hand-written numeric move. Non-digit bytes are clamped to 9.
- **ZES decoders** (`cmd/server/router.go`, `batiDecoder` / `nonBatiDecoder`): the ZES slots are mapped as follows:
  - slot 2 → `XBXTDDIR`
  - slot 3 → `XBXTDCOM`
  - slot 4 → `XBXTDSR`

  Unsigned values are converted to `int64`, and the sign flag is discarded.
- **Server and batch runner disagree on MAJIC handling:** `cmd/server` queries PostgreSQL for FMSTAU2, while `cmd/tfcalc` always returns CR 24 / RC 01. Decide which behavior is intended.
- **`JWT_SECRET`:** it was listed by the migration run, but no code reads it and the server has no authentication. Decide whether authentication should be added.
- **Parity check:** run golden vectors captured from production through `cmd/tfcalc` and diff the output before relying on any result.
