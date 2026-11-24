default: dev

dev:
    #!/usr/bin/env bash
    set -euo pipefail
    trap 'kill 0' EXIT INT TERM
    just frontend &
    just backend &
    wait

frontend:
    bun run --cwd web dev

backend:
    wgo -file '\.go$' go run ./cmd/vivere-api

check:
    test -z "$(gofmt -l cmd internal)"
    go vet ./...
    go test -race -count=1 ./...
    bun run --cwd web check

test:
    go test -race -count=1 ./...
    bun run --cwd web test

lint:
    go vet ./...
    bun run --cwd web lint

fmt:
    gofmt -w cmd internal
    bun run --cwd web fmt

build:
    go build ./...
    bun run --cwd web build
