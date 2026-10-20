# DGFiP Taxe Foncière: Go Port

This project is a Go port of a COBOL codebase that calculates French property tax (*taxe foncière*). The original was a plain COBOL codebase with no framework. The port covers two calculations:

- **Built property** (*propriétés bâties*), for the 2018 rules.
- **Unbuilt property** (*propriétés non bâties*), for the 2018 rules.

The calculation logic, the COBOL record and linkage layouts, and the rate tables are now Go packages. An HTTP server in `cmd/server` exposes them.

> **Status: the migration is not complete.**
>
> - 20 of 23 modules were migrated.
> - The project builds, and the Go unit tests pass.
> - **No one has compared its behavior with the original COBOL programs.** No request has been confirmed to produce the same result as the original.
> - The 93% figure is the migration model's own estimate. It is not a measured result.
> - Do not use this for real tax calculations until the outputs have been checked against the COBOL system.

## Tech stack

- **Language:** Go, standard library layout (`cmd/`, `internal/`)
- **Modules:** managed with `go.mod` and `go.sum`
- **Runtime:** a single HTTP server binary built from `./cmd/server`
- **Database:** reached through `DATABASE_URL`. A PostgreSQL instance was used during migration (see [Database](#database)).

## Prerequisites

- **Go 1.27.** The migration was built and tested with the `golang:1.27-alpine` toolchain.
- **A PostgreSQL instance** that you can reach with a connection URL.
- **A POSIX shell** with `mkdir`, for the commands below.

## Getting started

Run these steps in order from the root of a fresh clone.

### 1. Install dependencies and build

```sh
go mod download
mkdir -p bin
go build -o bin/server ./cmd/server
```

This creates the server binary at `bin/server`.

### 2. Configure the environment

```sh
export DATABASE_URL='postgres://USER:PASSWORD@HOST:5432/DBNAME?sslmode=disable'
export PORT=8080
export JWT_SECRET='replace-with-a-long-random-secret'
```

See [Environment variables](#environment-variables) for details on each variable.

### 3. Database setup

The migration did not provide any schema, migration or seed commands. There is no database setup step to run.

Make sure the database named in `DATABASE_URL` exists and accepts connections before you start the server.

### 4. Run the server

```sh
./bin/server
```

With `PORT=8080`, the server listens on `http://localhost:8080`.

### A first request to try

No request has been confirmed to return the same answer as the original system, so this README does not list an example request.

The HTTP routes are defined in `cmd/server/router.go`. Treat any response from them as unverified until it has been compared with the COBOL output.

## Database

| | |
|---|---|
| **Now** | The server reads a connection URL from `DATABASE_URL`. During migration it ran against a PostgreSQL instance, which is the only database it has been run with. |
| **Original** | This README does not document how the original COBOL codebase accessed data. |
| **Why it changed** | No reason was recorded. |

## Running tests

```sh
go test ./...
```

The Go unit tests passed during migration.

These packages have no test file:

- `internal/calc/nonbati2018/calc.go`
- `internal/config`
- `cmd/server`

Passing unit tests do not show that the code behaves like the original. That comparison has not been run.

## Environment variables

| Variable | Required | Purpose |
|---|---|---|
| `DATABASE_URL` | Yes | Connection URL for the database used by the server (PostgreSQL during migration). |
| `PORT` | Yes | TCP port the HTTP server listens on. The migration ran with `8080`. |
| `JWT_SECRET` | Yes | Secret the server uses for JWT handling. Use a long random value and never commit it. |

## Architecture overview

```
cmd/server/            HTTP server entry point
  main.go              startup and wiring
  router.go            HTTP route definitions
  store.go             database access (DATABASE_URL)
  layouts.go           server-side use of the record layouts

internal/config/       configuration loading (config.go)

internal/dispatch/     entry points that route a request to the right calculation
  bati.go              built-property dispatch
  nonbati.go           unbuilt-property dispatch

internal/calc/
  bati2018/            built-property calculation, 2018 rules (calc.go, cotis.go)
  nonbati2018/         unbuilt-property calculation, 2018 rules (calc.go, cotis.go)

internal/layout/       Go equivalents of COBOL record layouts
  linkage/             call interfaces: combat, comnba, retb, retnb
  taudis/              rate-distribution table records: t800, t84c, t84d, t84r
  zes/                 data area layouts: an, com, dir, sr

internal/rates/        rate data access
  basetaux.go          base rates
  taudis.go            rate-distribution tables
```

### How a request flows

1. `cmd/server` receives an HTTP request and decodes it using the `internal/layout` types.
2. It hands the request to `internal/dispatch`, which picks the built or unbuilt calculation.
3. The calculation in `internal/calc/...` computes the assessments (`cotis.go`), using rates from `internal/rates`.
4. The result goes back in the return layouts: `retb` for built property, `retnb` for unbuilt property.

### Where the COBOL code went

The package names keep the original COBOL member names. For example:

| Original member | Go file |
|---|---|
| `T84R` | `internal/layout/taudis/t84r.go` |
| `COMBAT` | `internal/layout/linkage/combat.go` |
| `RETB` | `internal/layout/linkage/retb.go` |

## Migration notes

| | |
|---|---|
| **Direction** | COBOL with no framework → Go, standard project layout |
| **Recorded changes from the original** | None. No intentional differences in behavior, interfaces or data handling were documented. |
| **Structural change** | COBOL programs and copybooks became Go packages. COBOL linkage sections and record layouts became Go types under `internal/layout`. |

The original was a batch- and linkage-call codebase. The Go port adds an HTTP server (`cmd/server`), configured through `PORT`, `DATABASE_URL` and `JWT_SECRET`. The reason for this change was not recorded.

## Known limitations

- **Behavior has not been compared with the original.** Nothing confirms that the Go code produces the same tax amounts, rounding or field formatting as the COBOL programs.
- **Three modules are unaccounted for.** Only 20 of 23 modules were migrated, but no untranslated files were listed. The 3 missing modules need to be identified against the original source tree.
- **No verified HTTP requests.** The server's endpoints have not been checked against reference outputs.
- **No database setup.** No schema or setup commands came with the migration. The tables `cmd/server/store.go` expects must be created by hand.
- **Unit test coverage has gaps.** See [Running tests](#running-tests) for the packages without tests.

## Manual review required

The migration model rated these original COBOL modules as low confidence. Review their Go counterparts line by line against the original source:

| Go file | Notes |
|---|---|
| `internal/layout/taudis/t84r.go` | Counterpart of the low-confidence `T84R` copybook. |

When reviewing, focus on:

- **Decimal arithmetic.** Check how COBOL `COMP-3` and implied-decimal fields map to Go types, and whether rounding and truncation match.
- **Field widths, padding and offsets** in `internal/layout/...`.
- **Rate table lookups and edge cases** in `internal/rates/basetaux.go` and `internal/rates/taudis.go`.

Before anyone relies on the output, set up a side-by-side comparison against the COBOL programs using real input records.