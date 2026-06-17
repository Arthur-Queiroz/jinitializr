# AGENTS.md

Operational rules for code agents working on J Initializr.

## Build & test

```sh
make web     # build frontend (required before any Go build — go:embed depends on web/dist)
make build   # build Go binary with embedded frontend
make test    # go test ./...
make tidy    # go mod tidy
```

Always run `make web` before `go build` or `go vet`. A missing `web/dist` causes
`assets.go` to fail with "no matching files found".

## Lint

```sh
gofmt -w .
go vet ./...
```

No external linters configured yet. If `staticcheck` is available, run it too.

## Architecture constraints

- **No external dependencies in Go.** The app uses only the standard library.
  Do not add modules to `go.mod` without explicit approval.
- **Frontend is Vue 3 + Vite**, lives in `web/`. Its `dist/` is embedded into the
  Go binary via `//go:embed` in `assets.go` (module root).
- **Templates** live in `internal/template/templates/` and are embedded the same way.
- **Manual DI only.** No DI frameworks. `cmd/server/main.go` wires everything.
- **`model/` is a leaf package** — no imports from other internal packages.
- **No import cycles.** If you need one, the type probably belongs in `model/`.

## Conventions

- Error wrapping: `fmt.Errorf("context: %w", err)`.
- Handlers translate errors to HTTP status; internal packages return errors, never
  write HTTP responses.
- Package names: lowercase singular (`handler`, not `handlers`).
- No stutter: `catalog.Catalog`, not `catalog.CatalogStruct`.
- Routing: stdlib `http.ServeMux` with Go 1.22+ method+path patterns
  (`mux.HandleFunc("POST /api/generate", h.Generate)`).
- Middleware: `func(http.Handler) http.Handler` composition, no framework.

## Adding a dependency to the catalog

1. Add an entry in `internal/catalog/catalog.go`.
2. Add corresponding `.tmpl` files in `internal/template/templates/`.
3. Do not modify the frontend — the catalog endpoint serves the new entry
   automatically and the UI renders it.

## What not to do

- Do not add a Go web framework (Chi/Gin are output options, not app dependencies).
- Do not move `assets.go` — `//go:embed` cannot reference parent directories.
- Do not add comments that describe *what* the code does, only *why*.
- Do not create `util`, `common`, `helpers`, or `api` packages.
