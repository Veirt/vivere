dev-server:
    wgo -cd server -file=.go go run cmd/main.go
dev-web:
    bun run dev

build: build-web
build-web:
    bun run build



