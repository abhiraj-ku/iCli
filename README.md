# iCli - PostgreSQL query and index advisor

`iCli` is a Go CLI that reads PostgreSQL query statistics, explains the most expensive queries, identifies possible execution-plan bottlenecks, and reports unused indexes.

## Demo

![iCli example output](/assets/demo.gif)

## Features

- Finds expensive queries using `pg_stat_statements`.
- Runs `EXPLAIN (FORMAT JSON)` on the selected queries.
- Reports plan issues such as sequential scans and memory-heavy operations.
- Finds non-primary, non-unique indexes with zero recorded scans.
- Prints actionable index and session recommendations.

## Requirements

- Go 1.25 or later.
- PostgreSQL with `pg_stat_statements` enabled.

Enable the extension at the server level, then restart PostgreSQL:

```sql
ALTER SYSTEM SET shared_preload_libraries = 'pg_stat_statements';
```

After restarting PostgreSQL, enable it in the target database:

```sql
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;
```

On a Homebrew PostgreSQL 18 installation, restart with:

```bash
brew services restart postgresql@18
```

## Run Locally

Clone the repository and run the CLI with a PostgreSQL connection string.

DSN resolution order:

1. `-d` / `--dsn`
2. `PGDSN` environment variable

Examples:

```bash
git clone <repository-url>
cd pg_advisor
export PGDSN="postgres://postgres@localhost:5432/postgres?sslmode=disable"
go run ./cmd/icli report
```

Or pass the DSN explicitly:

```bash
go run ./cmd/icli -d="postgres://postgres@localhost:5432/postgres?sslmode=disable" report
```

The long-form flag is also supported:

```bash
go run ./cmd/icli --dsn="postgres://postgres@localhost:5432/postgres?sslmode=disable" report
```

Build a binary with:

```bash
go build -o icli ./cmd/icli
export PGDSN="postgres://postgres@localhost:5432/postgres?sslmode=disable"
./icli report
```

## Developer Setup

### Prerequisites

- Go 1.25 or later
- PostgreSQL instance running locally or remotely
- `pg_stat_statements` enabled in the target database
- a valid DSN available through either `-d/--dsn` or `PGDSN`

### Local database setup

Create a local PostgreSQL database and enable the extension:

```sql
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;
```

Then export the connection string for local development:

```bash
export PGDSN="postgres://postgres@localhost:5432/postgres?sslmode=disable"
```

### Repo layout

- `cmd/icli` — CLI entrypoint
- `internals/analyzer` — execution-plan issue detection
- `internals/db` — SQL queries and PostgreSQL connectivity
- `internals/health` — database health collection and rendering
- `internals/report` — terminal UI output for findings and reports

### Common dev commands

```bash
go test ./...
go run ./cmd/icli report
GOFLAGS=-mod=mod go run ./cmd/icli report
```

### Troubleshooting

- If you see `connection refused`, confirm PostgreSQL is running and the host/port are correct.
- If the report is empty, ensure `pg_stat_statements` is enabled and the user has access to query stats.
- If the DSN is missing, pass `-d` / `--dsn` or export `PGDSN`.
- If formatting looks odd in a narrow terminal, run it in a wider terminal or resize the window.

## Releases

Prebuilt releases are available for these platforms:

| Platform | Architecture | Archive name pattern |
| --- | --- | --- |
| Linux | x86_64 | `<project>_Linux_x86_64.tar.gz` |
| macOS (Apple Silicon, M-series) | ARM 64-bit | `<project>_Darwin_arm64.tar.gz` |

Download the archive for your platform from the project's Releases page, extract it, and run `icli`:

```bash
tar -xzf <release-archive>.tar.gz
./icli -dsn="postgres://user:password@localhost:5432/database?sslmode=disable"
```

Windows and Intel macOS binaries are not included in releases.

## How It Works

1. Reads the PostgreSQL server version to select the correct execution-time column.
2. Queries `pg_stat_statements` and excludes iCli's own telemetry query.
3. Sanitizes parameter placeholders before requesting an `EXPLAIN (FORMAT JSON)` plan.
4. Walks the plan and reports detected issues.
5. Queries `pg_stat_user_indexes` and `pg_index` to find unused non-primary, non-unique indexes.

Query statistics and index statistics are collected by PostgreSQL and may reset when the server restarts or statistics are manually reset.
