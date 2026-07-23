package httpserver

import (
	"context"
	"fmt"

	"github.com/0Hoag/cryptocheck-api/internal/seeder"
)

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
	return srv.gin.Run(fmt.Sprintf(":%d", srv.port))
}
