dev-server:
    wgo -cd server -file=.go go run ./cmd

dev-web:
    vp dev

build: build-web

build-web:
    vp build

check:
    bun run check
