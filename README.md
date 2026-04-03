# axon-base

PostgreSQL foundation library for axon services. Provides connection pooling,
transaction helpers, SQL migration running, row scanning utilities, and a
generic repository interface — all without ORMs or query builders.

## Prerequisites

- Go 1.24+
- [just](https://github.com/casey/just)
- PostgreSQL (for integration tests)

## Packages

| Package | Purpose |
|---------|---------|
| `pool` | pgxpool wrapper with health checks and metrics |
| `migration` | Run embedded SQL migrations via golang-migrate |
| `scan` | Map query results to structs without reflection |
| `repository` | Generic CRUD interface contract |
| `errors` | Error wrapping and pgx error predicates |

## Usage

### Connection pool

```go
import "github.com/benaskins/axon-base/pool"

p, err := pool.NewPool(ctx, "postgres://postgres@localhost:5432/mydb")
if err != nil {
    return err
}
defer p.Close()

if !p.Healthy(ctx) {
    return errors.New("database unreachable")
}
```

### Pool metrics (health endpoints)

```go
m := p.Metrics()
data, err := m.HealthJSON()
// {"active":1,"idle":3,"total":4,"max":4,"wait_time_ms":0}
```

### Transactions

```go
import (
    "github.com/benaskins/axon-base/pool"
    "github.com/jackc/pgx/v5"
)

err = p.WithTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
    _, err := tx.Exec(ctx,
        "INSERT INTO orders (id, user_id) VALUES ($1, $2)",
        orderID, userID,
    )
    return err
})
```

### Migrations

Embed your SQL files and run them at startup:

```go
import (
    "database/sql"
    "embed"

    "github.com/benaskins/axon-base/migration"
    _ "github.com/jackc/pgx/v5/stdlib"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func runMigrations(dsn string) error {
    db, err := sql.Open("pgx", dsn)
    if err != nil {
        return err
    }
    defer db.Close()
    return migration.Migrate(db, migrationsFS, "migrations")
}
```

SQL files follow golang-migrate naming: `000001_create_users.up.sql`,
`000001_create_users.down.sql`.

### Row scanning

Map query columns to struct fields explicitly — column order must match field
order in the mapper:

```go
import "github.com/benaskins/axon-base/scan"

type User struct {
    ID    string
    Email string
    Name  string
}

// Single row
row := db.QueryRow(ctx, "SELECT id, email, name FROM users WHERE id = $1", id)
user, err := scan.Row(row, func(u *User) []any {
    return []any{&u.ID, &u.Email, &u.Name}
})

// Multiple rows
rows, err := db.Query(ctx, "SELECT id, email, name FROM users ORDER BY name")
if err != nil {
    return err
}
users, err := scan.Rows(rows, func(u *User) []any {
    return []any{&u.ID, &u.Email, &u.Name}
})
```

### Repository pattern

Implement the generic `Repository[T]` interface for any entity type:

```go
import (
    "github.com/benaskins/axon-base/repository"
    "github.com/benaskins/axon-base/scan"
    axerrors "github.com/benaskins/axon-base/errors"
)

type User struct {
    ID   string
    Name string
}

type UserRepository struct {
    pool *pool.Pool
}

// Compile-time check.
var _ repository.Repository[User] = (*UserRepository)(nil)

func (r *UserRepository) Get(ctx context.Context, id string) (User, error) {
    row := r.pool.DB().QueryRow(ctx,
        "SELECT id, name FROM users WHERE id = $1", id,
    )
    user, err := scan.Row(row, func(u *User) []any {
        return []any{&u.ID, &u.Name}
    })
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return User{}, axerrors.ErrNotFound
        }
        return User{}, fmt.Errorf("users.Get: %w", err)
    }
    return user, nil
}
```

### Error utilities

```go
import axerrors "github.com/benaskins/axon-base/errors"

// Wrap errors with operation context.
return axerrors.WrapError("users.Create", err)

// Check for not-found (works through wrapping chains).
if axerrors.IsNotFoundError(err) { ... }

// Check for unique constraint violations.
if axerrors.IsUniqueViolation(err) { ... }
```

## Development

```bash
just test   # run tests (requires Postgres at localhost:5432/workbench)
just vet    # run go vet
just build  # compile
```
