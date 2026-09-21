# iCli🐘


`iCli` is a Go CLI tool that helps you figure out why your Postgres queries are slow. It looks at your database's query history, runs an explain plan on the worst offenders, and tells you exactly which indexes you need to add to fix them.

![iClidemo](assets/demo.gif)
*(Example: Finding a missing index on a 5M row table)*

## Why I Built This
I got tired of manually digging through `pg_stat_statements` and trying to read massive JSON `EXPLAIN` plans every time the database CPU spiked.. I built this to understand explain and pg_stat_statements a little better.

## What It Does
* **Finds slow queries:** Grabs the worst queries based on actual total execution time.
* **Spots bad plans:** Parses the `EXPLAIN (FORMAT JSON)` output to catch issues like massive sequential scans or disk-based sorts.
* **Cleans up dead weight:** Checks `pg_stat_user_indexes` to find indexes that are never used for reads but are slowing down your `INSERT`s.
* **Safe recommendations:** Spits out the exact `CREATE INDEX CONCURRENTLY` command you need so you don't lock your tables in production.

## Installation

You can install `iCli` directly via Go:

```bash
go install [github.com/yourusername/iCli/cmd/pgadvisor@latest](https://github.com/yourusername/pg-advisor/cmd/iCli@latest)
```

## Quick Start

Run the analyzer against any PostgreSQL database:

```bash
pgadvisor analyze --url "postgres://user:pass@localhost:5432/production_db"
```

### Example Output
```text
🔍 Analyzing Top 5 Expensive Queries...

Query (Calls: 14,500, Total Time: 4,230 ms):
SELECT * FROM orders WHERE status = $1 AND created_at > $2;

  [WARNING] High-cost Sequential Scan detected!
    - Table: orders
    - Estimated Rows Scanned: 450,000
    - Filter applied: (status = 'pending') AND (created_at > '2023-01-01')
    
    💡 RECOMMENDATION: 
    CREATE INDEX CONCURRENTLY idx_orders_status_created ON orders(status, created_at);
------------------------------------------------------------
```

## Architecture & How It Works

1. **Telemetry:** The tool queries `pg_stat_statements` to find slow queries.
2. **Sanitization:** It uses regex to replace parameters (`$1`, `$2`) with `NULL` to bypass Postgres parameter errors without executing the query.
3. **Execution Tree Traversal:** It runs `EXPLAIN (FORMAT JSON)` and unmarshals the output into a recursive Go struct (`PlanNode`). 
4. **Analysis:** A tree-walking algorithm inspects each node's `NodeType`, `Total Cost`, and `Plan Rows` against predefined heuristic thresholds.

## Local Development

```bash
# Clone the repository
git clone [https://github.com/yourusername/iCli.git](https://github.com/yourusername/iCli.git)
cd pg-advisor

# Start a dummy Postgres database with pg_stat_statements enabled
make db-up

# Build and run the CLI
make run
