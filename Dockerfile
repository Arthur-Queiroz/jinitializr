# syntax=docker/dockerfile:1

# Stage 1 — build the Vue SPA that gets embedded into the Go binary.
FROM node:22-alpine AS web
WORKDIR /web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

# Stage 2 — build the single Go binary with the SPA embedded.
FROM golang:1.24-alpine AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
# Bring in the freshly built SPA (web/dist is .dockerignored on purpose).
COPY --from=web /web/dist ./web/dist
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /bin/j-initializr ./cmd/server

# Stage 3 — minimal, non-root runtime image. The binary is fully self-contained
# (frontend + templates embedded), so a static base is all it needs.
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /bin/j-initializr /j-initializr
EXPOSE 8080
ENTRYPOINT ["/j-initializr"]
