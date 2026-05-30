default: dev

dev:
    #!/usr/bin/env bash
    set -euo pipefail
    trap 'kill 0' EXIT INT TERM
    just frontend &
    just backend &
    wait

frontend:
    bun --cwd web run dev

backend:
    wgo -file '\.go$' go run ./cmd/vivere-api
