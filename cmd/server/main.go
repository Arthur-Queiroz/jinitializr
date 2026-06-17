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

	srv := &http.Server{
		Addr:    addr,
		Handler: mux,
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
