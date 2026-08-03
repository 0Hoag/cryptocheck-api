package httpserver

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/0Hoag/cryptocheck-api/internal/seeder"
)

const (
	serverReadHeaderTimeout = 5 * time.Second
	serverReadTimeout       = 15 * time.Second
	// Scanner requests intentionally allow up to 45 seconds for upstream
	// explorers, so the server write timeout must remain above that budget.
	serverWriteTimeout = 60 * time.Second
	serverIdleTimeout  = 60 * time.Second
	serverShutdownWait = 10 * time.Second
)

func (srv HTTPServer) newHTTPServer() *http.Server {
	return &http.Server{
		Addr:              fmt.Sprintf(":%d", srv.port),
		Handler:           srv.gin,
		ReadHeaderTimeout: serverReadHeaderTimeout,
		ReadTimeout:       serverReadTimeout,
		WriteTimeout:      serverWriteTimeout,
		IdleTimeout:       serverIdleTimeout,
	}
}

func normalizeServerRunError(err error) error {
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func (srv HTTPServer) Run() error {
	ctx := context.Background()

	// Seed before initializing route dependencies. A clean local database must be
	// usable on the very first startup, before any repository or queue setup.
	if err := seeder.Run(ctx, srv.db); err != nil {
		return fmt.Errorf("seed database: %w", err)
	}

	if err := srv.mapHandlers(); err != nil {
		return err
	}

	srv.l.Infof(ctx, "Started server on :%d", srv.port)
	server := srv.newHTTPServer()
	errCh := make(chan error, 1)
	go func() {
		errCh <- server.ListenAndServe()
	}()

	shutdownCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	select {
	case err := <-errCh:
		return normalizeServerRunError(err)
	case <-shutdownCtx.Done():
		srv.l.Infof(ctx, "Shutting down HTTP server gracefully")
		ctx, cancel := context.WithTimeout(context.Background(), serverShutdownWait)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			return fmt.Errorf("graceful HTTP shutdown: %w", err)
		}
		return nil
	}
}
