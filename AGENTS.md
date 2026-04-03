# axon-base

Initialize Go module, add pgx/v5 and golang-migrate dependencies, create directory structure (pool, repository, migration, helpers). Test: `go mod tidy` succeeds, `go build ./...` compiles.

## Build & Test

```bash
go test ./...
go vet ./...
just build     # builds to bin/axon-base
just install   # copies to ~/.local/bin/axon-base
```

## Module Selections


## Deterministic / Non-deterministic Boundary

| From | To | Type |
|------|----|------|

