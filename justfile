build:
    go build -o bin/axon-base ./cmd/axon-base

install: build
    cp bin/axon-base ~/.local/bin/axon-base

test:
    go test ./...

vet:
    go vet ./...
