# axon-base

Postgres database primitives: connection pool, repository interfaces, migrations, row scanning, and transactions.

## Module

- Module path: `github.com/benaskins/axon-base`
- Project type: library (no main package)

## Build & Test

```bash
just test    # go test -race ./...
just vet     # go vet ./...
```

Integration tests require Postgres at `localhost:5433` (database `workbench`, user `postgres`). Tests skip gracefully when unavailable.

## Architecture

| Package | Purpose |
|---------|---------|
| `pool` | pgx connection pool wrapper with health metrics and transactions |
| `repository` | Generic repository interfaces and base implementations |
| `migration` | Goose-based schema migrations with schema isolation |
| `scan` | Row scanning utilities for explicit struct mapping |

Read [AGENTS.md](./AGENTS.md) for architecture details.

## Constraints

- No ORM or query builder — all SQL must be explicit, no `SELECT *`
- Only Postgres (pgx/v5) — no other databases
- Tests must use real Postgres, not mocks or testcontainers
- No third-party assertion libraries — standard `testing` package only
