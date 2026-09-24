# Contributing

Thanks for helping improve iCli.

## Prerequisites

- Go 1.25 or later
- PostgreSQL instance with `pg_stat_statements` enabled
- a valid DSN for the target database

## Local setup

1. Clone the repository.
2. Install dependencies:

```bash
go mod download
```

3. Start PostgreSQL locally and enable the extension:

```sql
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;
```

4. Export a DSN for local development:

```bash
export PGDSN="postgres://postgres@localhost:5432/postgres?sslmode=disable"
```

## Common commands

```bash
go test ./...
go run ./cmd/icli report
```

You can also pass the DSN directly:

```bash
go run ./cmd/icli -d="postgres://postgres@localhost:5432/postgres?sslmode=disable" report
```

## Development notes

- Format code before submitting changes:

```bash
gofmt -w ./...
```

- Keep changes focused and avoid unrelated refactors.
- Prefer small, reviewable pull requests.

## Pull requests

Before opening a PR:

- run `go test ./...`
- ensure the CLI still runs with your changes
- update docs when behavior changes
- include a clear summary of the problem and fix

## Reporting issues

Open an issue with:

- a minimal reproduction
- the command used
- expected behavior
- actual behavior
- relevant PostgreSQL version and schema details
