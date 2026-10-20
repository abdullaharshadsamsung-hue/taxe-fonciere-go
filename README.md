# DGFiP Taxe Foncière (Go)

A Go migration of the DGFiP *taxe foncière* (French property tax) calculation, originally written in COBOL. It calculates taxe foncière for **built** properties (*bâti*) and **unbuilt** properties (*non bâti*), using the 2018 calculation rules (`bati2018`, `nonbati2018`). The calculation is exposed in four ways:

- an HTTP server (`cmd/server`)
- a command-line calculator (`cmd/tfcalc`)
- a loader for TAUDIS rate data (`cmd/taudisload`)
- a Go package (`pkg/tfcalc`)

> **Status:** The project builds and its unit tests pass. Its behavior has **not** been compared with the original COBOL programs. Do not use it for production tax calculations until that comparison is done. See [Known limitations](#known-limitations) and [Manual review required](#manual-review-required).

## Tech stack

| Component | Technology |
|---|---|
| Language | Go (standard toolchain, Go modules: `go.mod` / `go.sum`) |
| Database | PostgreSQL (reached through `DATABASE_URL`) |
| Interfaces | HTTP server, CLI, Go library package |

## Prerequisites

- **Go 1.27.** The migration was built and tested with the `golang:1.27-alpine` toolchain.
- **A running PostgreSQL instance** that the server can reach.
- A POSIX shell to run the commands below.

## Getting started

Run these steps in order, from the root of a fresh clone.

### 1. Install dependencies and build the server

```sh
go mod download
mkdir -p bin
go build -o bin/server ./cmd/server
```

### 2. Configure the environment

```sh
export DATABASE_URL="postgres://USER:PASSWORD@HOST:5432/DBNAME?sslmode=disable"
export PORT=8080
export JWT_SECRET="change-me"
```

Replace the placeholder values with your own. See [Environment variables](#environment-variables) for what each variable does.

### 3. Set up the database

No database setup command was recorded for this migration.

- The TAUDIS storage schema is defined in `internal/taudis/schema.go`.
- A loader for TAUDIS data is provided in `cmd/taudisload/main.go`.

Before you run the server, confirm the following in that code:

- how the schema gets created
- how the TAUDIS data gets loaded

### 4. Run the server

```sh
./bin/server
```

The server listens on port **8080**. You can reach it at `http://localhost:8080`.

### A first request to try

No request has yet been checked against the original system's behavior, so this README does not list one. The HTTP routes are defined in `cmd/server/router.go`.

## Database

| | Database |
|---|---|
| Original (COBOL) | No database platform was recorded for the original codebase. |
| Migrated (Go) | PostgreSQL, configured through `DATABASE_URL`. |

The reason for introducing PostgreSQL was not recorded during the migration. Before deploying, check that the data held in PostgreSQL (TAUDIS rates, see `internal/taudis/`) matches the data the original programs used.

## Running tests

```sh
go test ./...
```

All target unit tests passed during the migration. These tests check the Go code on its own terms. They do **not** show that the Go code produces the same results as the original COBOL programs.

The following packages have no test files:

- `internal/calc/nonbati2018/calc.go` (only `cotis.go` is tested)
- `internal/config`
- `internal/taudis`
- `pkg/tfcalc`
- all `cmd/*` programs

## Environment variables

| Variable | Required | Description |
|---|---|---|
| `DATABASE_URL` | Yes | PostgreSQL connection string used by the server. |
| `PORT` | Yes | HTTP listen port. Port `8080` was used during the migration. |
| `JWT_SECRET` | Yes | Secret used by the server for JWT handling. Use a strong, private value outside local development. |

Configuration loading is implemented in `internal/config/config.go`.

## Architecture overview

```
cmd/
  server/         HTTP server: entry point (main.go), routes (router.go),
                  data access (store.go), request/response layouts (layouts.go)
  tfcalc/         Command-line calculator
  taudisload/     Loader for TAUDIS rate data
pkg/
  tfcalc/         Public Go API for the tax calculation
internal/
  calc/
    bati2018/     Built-property calculation (calc.go) and contributions (cotis.go)
    nonbati2018/  Unbuilt-property calculation (calc.go) and contributions (cotis.go)
  dispatch/       Routes bâti / non-bâti requests to the correct calculator
  layout/
    linkage/      Translated COBOL linkage records:
                  COMBAT, COMNBA (inputs); RETB, RETNB (returns)
    taudis/       TAUDIS record layouts: T800, T84C, T84D, T84R
    zes/          ZES record layouts: AN, COM, DIR, SR
  rates/          Rate handling: base rates (basetaux.go), TAUDIS rates (taudis.go)
  taudis/         TAUDIS persistence: schema (schema.go) and store (store.go)
  config/         Environment-based configuration
```

**Request flow.** A request enters through `cmd/server`, `cmd/tfcalc` or `pkg/tfcalc`. `internal/dispatch` then selects the bâti or non-bâti calculator in `internal/calc`. The calculator reads rates through `internal/rates`, which is backed by `internal/taudis`. Inputs and outputs keep the structure of the original COBOL records, as defined in `internal/layout`.

## Migration notes

- **Source → target:** COBOL → Go (standard toolchain).
- **Modules migrated:** 20 of 23. The migration did not record which 3 modules are missing. It also listed no untranslated or unmigrable files. Resolve this discrepancy before relying on the code (see [Known limitations](#known-limitations)).
- **Recorded changes from the original:** none. No behavioral, structural or data-format changes were recorded.
  - PostgreSQL configuration (`DATABASE_URL`) appears in the migrated project, but no change or reason was recorded for it.
  - JWT configuration (`JWT_SECRET`) appears in the migrated project, but no change or reason was recorded for it.
- **Confidence:** the migration's own estimate is 93%. This is a self-assessment, not a measured result.

## Known limitations

- **Behavior not verified against the original.** No outputs from the Go code have been compared with outputs from the COBOL programs. Tax amounts, rounding and edge cases may differ.
- **3 of 23 modules unaccounted for.** The migration reports 20 of 23 modules migrated, but does not identify which 3 are missing or why.
- **No verified HTTP requests.** No endpoint has been confirmed to respond the way the original system does.
- **No recorded database setup.** There is no documented step for creating the schema or loading TAUDIS data.
- **Partial test coverage.** Several packages have no tests (see [Running tests](#running-tests)), including the unbuilt-property calculation in `internal/calc/nonbati2018/calc.go`.

## Manual review required

The migration flagged the following original COBOL modules as **low confidence**. For each one, find its Go counterpart and check it line by line against the COBOL source.

| Original module | Notes |
|---|---|
| T84R | TAUDIS T84R record. Likely counterpart: `internal/layout/taudis/t84r.go`. |
| XCOTB | Copybook. Locate its Go counterpart and verify field layouts. |
| EFITAUX2 | Program. Locate its Go counterpart. |
| FMSTAU2 | Program. Locate its Go counterpart. |
| EFITA3B8 | Program (bâti). Check against `internal/calc/bati2018/`. |
| EFITA3N8 | Program (non bâti). Check against `internal/calc/nonbati2018/`. |
| CTXTA3B | Program (bâti). Locate its Go counterpart. |
| CTXTA3N | Program (non bâti). Locate its Go counterpart. |

The bâti / non-bâti suggestions above are inferred from the module names only. Confirm each mapping yourself.

Also review:

- `internal/calc/nonbati2018/calc.go`, because it has no unit tests.
- Server configuration and authentication (`internal/config/config.go`, `cmd/server/`), because the use of `JWT_SECRET` has no documented counterpart in the original system.
- The database schema and loader (`internal/taudis/schema.go`, `cmd/taudisload/main.go`), to make sure the rate data matches what the original programs consumed.
