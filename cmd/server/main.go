// Command server is the J Initializr bootstrap: it builds the dependency graph
// by hand and serves the HTTP API on :8080 using only the standard library.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	assets "github.com/Arthur-Queiroz/j-initializr"
	"github.com/Arthur-Queiroz/j-initializr/internal/catalog"
	"github.com/Arthur-Queiroz/j-initializr/internal/generator"
	"github.com/Arthur-Queiroz/j-initializr/internal/handler"
	"github.com/Arthur-Queiroz/j-initializr/internal/template"
	"github.com/Arthur-Queiroz/j-initializr/internal/zipper"
)

const (
	addr            = ":8080"
	shutdownTimeout = 10 * time.Second

	// Per-IP rate limit: a burst for normal page loads (catalog + a few
	// generates), sustained at a modest steady rate.
	rateLimit = 10 // requests per second
	rateBurst = 30
)

func main() {
	// Minimal structured logger for boot — text handler, stdlib only.
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Manual dependency injection: build everything, then wire it together.
	cat := catalog.New()
	tmpl := template.New()
	zip := zipper.New()
	gen := generator.New(cat, tmpl, zip)

	webFS, err := assets.WebFS()
	if err != nil {
		logger.Error("failed to mount embedded frontend", "err", err)
		os.Exit(1)
	}
	h := handler.New(gen, cat, webFS)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	// Cross-cutting concerns via middleware composition (no framework):
	// RequestID is outermost so every log line and panic carries an id; the
	// rate limiter is innermost so rejected requests are still logged.
	//
	// TRUST_CLOUDFLARE=true (set in the Compose file on the VPS) makes the
	// limiter key on CF-Connecting-IP. Behind the tunnel every connection comes
	// from cloudflared, so without it the per-IP limit would collapse into a
	// single global bucket. Off by default for local/direct runs.
	trustProxy := os.Getenv("TRUST_CLOUDFLARE") == "true"
	limiter := handler.NewRateLimiter(rateLimit, rateBurst, trustProxy)
	root := handler.Chain(mux,
		handler.RequestID,
		handler.Logging(logger),
		handler.Recover(logger),
		limiter.Limit,
	)

	srv := &http.Server{
		Addr:    addr,
		Handler: root,
		// Timeouts guard against slow/idle clients holding connections open.
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// Run the server in the background so main can wait for a shutdown signal.
	go func() {
		logger.Info("starting j-initializr", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server stopped", "err", err)
			os.Exit(1)
		}
	}()

	// Block until an interrupt arrives, then drain in-flight requests.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	logger.Info("shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("graceful shutdown failed", "err", err)
		os.Exit(1)
	}
	logger.Info("stopped")
}
