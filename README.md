# axon-base

Initialize Go module, add pgx/v5 and golang-migrate dependencies, create directory structure (pool, repository, migration, helpers). Test: `go mod tidy` succeeds, `go build ./...` compiles.

## Prerequisites

- Go 1.24+
- [just](https://github.com/casey/just)

## Build & Run

```bash
just build
just install
axon-base --help
```

## Development

```bash
just test   # run tests
just vet    # run go vet
```
