.PHONY: dev web build run test tidy docker clean

# Build the Vue SPA into web/dist — required before `go build`/`run` because
# assets.go embeds web/dist into the binary.
web:
	cd web && npm install && npm run build

# Build the single self-contained binary (frontend embedded). Rebuilds the SPA.
build: web
	go build -o bin/j-initializr ./cmd/server

# Run the server. Expects web/dist to exist (run `make web` once first).
run:
	go run ./cmd/server

# Frontend dev server with API proxy to :8080 (run the backend separately).
dev:
	cd web && npm run dev

test:
	go test ./...

tidy:
	go mod tidy

docker:
	docker build -t j-initializr .

clean:
	rm -rf bin web/dist
