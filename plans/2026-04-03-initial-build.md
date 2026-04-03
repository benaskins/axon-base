# axon-base — Initial Build Plan
# 2026-04-03

Each step is commit-sized. Execute via `/iterate`.

## Step 1 — Set up project structure and dependencies

Initialize Go module, add pgx/v5 and golang-migrate dependencies, create directory structure (pool, repository, migration, helpers). Test: `go mod tidy` succeeds, `go build ./...` compiles.

Commit: `feat: initialize project structure with pgx and migrate dependencies`

## Step 2 — Implement connection pool wrapper

Create Pool type wrapping pgxpool.Pool with NewPool constructor accepting DSN or config params. Add health check method, graceful shutdown (Close), and pool metrics getter. Test: create pool from DSN, verify health check returns true, call Close and verify pool is closed.

Commit: `feat: implement Pool type with health check and graceful shutdown`

## Step 3 — Implement transaction helper

Create TxFunc callback type and WithTransaction function that begins a transaction, executes callback, commits on success, rolls back on error. Wrap errors with operation context. Test: successful transaction commits, failed transaction rolls back, error is wrapped with context.

Commit: `feat: implement transaction helper with rollback on error`

## Step 4 — Implement migration runner

Create migration runner using golang-migrate with embedded filesystem for SQL migrations. Add Migrate function for startup execution, supporting up/down migrations. Test: run migrations on test DB, verify tables created, run down migrations, verify tables dropped.

Commit: `feat: implement migration runner with embedded SQL files`

## Step 5 — Implement row scanning helpers

Create helper functions for mapping query results to Go structs (ScanRow, ScanRows). Use reflection carefully for field mapping. Test: query single row maps to struct, query multiple rows returns slice of structs.

Commit: `feat: implement row scanning helpers for struct mapping`

## Step 6 — Implement repository interface pattern

Define generic Repository interface with Create, Get, Update, Delete, List methods. Create a base implementation showing how to satisfy the interface. Test: create a concrete repository, verify all interface methods work with context propagation.

Commit: `feat: define repository interface pattern with CRUD methods`

## Step 7 — Add error wrapping utilities

Create error wrapping helpers that add context to pgx errors (WrapError, IsNotFoundError, IsUniqueViolation). Test: database errors are wrapped correctly, error predicates work as expected.

Commit: `feat: add error wrapping utilities with context`

## Step 8 — Add connection pool metrics and health endpoint helpers

Expose pool metrics struct (Active, Idle, Max, WaitTime) with getter method on Pool. Add helper to format metrics for health endpoints. Test: metrics reflect actual pool state, health endpoint format is valid JSON.

Commit: `feat: expose connection pool metrics for health checks`

## Step 9 — Write integration tests with real Postgres

Create integration test suite using a real Postgres instance (via testcontainer or local DB). Test all pool operations, transactions, migrations, and repository patterns. Test: all tests pass against real Postgres, no mocks used.

Commit: `test: add integration tests with real Postgres instance`

## Step 10 — Add documentation and examples

Document all public types and functions with godoc comments. Add example usage in README showing Pool creation, migrations, and repository pattern. Test: `go doc` shows all public API, examples compile and run.

Commit: `docs: add godoc comments and usage examples`

