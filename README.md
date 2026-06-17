# J Initializr

A [start.spring.io](https://start.spring.io)-style project generator for Go. You
configure a Go project through a web form (module path, version, router,
dependencies) and download a `.zip` with a ready-to-run scaffold.

The generator produces **only the Go backend** — it does not generate a
frontend. (J Initializr itself has a Vue frontend, but that is separate from
what it generates, just as Spring Initializr is written in Java yet only
generates the Spring project.)

## Philosophy

Go does almost everything with the standard library. The app uses `net/http`,
`text/template`, `archive/zip` and `embed` — **zero required dependencies**. It
ships as a single binary with the frontend and templates embedded. Dogfooding is
the point: the tool demonstrates that Go needs no framework for the basics.

## Running locally

Build the frontend once, then run the server (it serves the SPA and the API on
`:8080`):

```sh
make web   # build web/dist (npm install + vite build)
make run   # go run ./cmd/server
```

Open http://localhost:8080.

For frontend development with hot reload, run the backend and the Vite dev
server (it proxies `/api` to `:8080`) in two terminals:

```sh
make run   # terminal 1
make dev   # terminal 2 → http://localhost:5173
```

## Build a single binary

```sh
make build   # → bin/j-initializr (frontend embedded)
```

`web/dist` must exist at `go build` time (the `make` targets handle this); a bare
`go build ./...` without it will fail with "no matching files found".

## Docker

```sh
make docker   # multi-stage build: Vue SPA + Go binary → distroless image
docker run -p 8080:8080 j-initializr
```

## API

- `GET /api/catalog` — routers and dependencies. Returned once on load; search is
  client-side.
- `POST /api/generate` — accepts a `ProjectConfig` JSON body, returns
  `application/zip`.

## Layout

```
cmd/server/        bootstrap: manual DI, HTTP server
assets.go          go:embed of web/dist (must sit at module root)
internal/
  handler/         HTTP: parse, call generator, return zip/json
  generator/       orchestration: config → file map → zip
  template/        text/template renderer + embedded .tmpl files
  zipper/          map[path][]byte → in-memory zip
  model/           central types (no logic)
  catalog/         registry of routers and dependencies
web/               Vue 3 + Vite (dist/ embedded into the binary)
```

See `CLAUDE.md` for conventions and the design rationale.
