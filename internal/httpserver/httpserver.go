package httpserver

import (
	"context"
	"fmt"
	"net/http"
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
	return srv.newHTTPServer().ListenAndServe()
}
