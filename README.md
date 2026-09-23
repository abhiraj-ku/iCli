# iCli - PostgreSQL query and index advisor

<p align="center">
	<img src="assets/icli-cover (1).png" alt="iCli PostgreSQL advisor logo" width="320">
</p>

`iCli` is a Go CLI that reads PostgreSQL query statistics, explains the most expensive queries, identifies possible execution-plan bottlenecks, and reports unused indexes.

## Demo

![iCli example output](assets/exampl.png)

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

Clone the repository and run the CLI with a PostgreSQL connection string:

```bash
git clone <repository-url>
cd pg_advisor
go run ./cmd/icli -dsn="postgres://user:password@localhost:5432/database?sslmode=disable"
```

The short flag `-d` is also supported:

```bash
go run ./cmd/icli -d="postgres://postgres@localhost:5432/postgres?sslmode=disable"
```

Build a binary with:

```bash
go build -o icli ./cmd/icli
./icli -dsn="postgres://user:password@localhost:5432/database?sslmode=disable"
```

## How It Works

1. Reads the PostgreSQL server version to select the correct execution-time column.
2. Queries `pg_stat_statements` and excludes iCli's own telemetry query.
3. Sanitizes parameter placeholders before requesting an `EXPLAIN (FORMAT JSON)` plan.
4. Walks the plan and reports detected issues.
5. Queries `pg_stat_user_indexes` and `pg_index` to find unused non-primary, non-unique indexes.

Query statistics and index statistics are collected by PostgreSQL and may reset when the server restarts or statistics are manually reset.
