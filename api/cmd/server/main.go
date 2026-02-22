package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"golang.org/x/sync/errgroup"

	"github.com/playperu/cityquiz/internal/config"
	"github.com/playperu/cityquiz/internal/database"
	"github.com/playperu/cityquiz/internal/server"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := run(ctx, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, stdout io.Writer) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	logger := slog.New(slog.NewJSONHandler(stdout, &slog.HandlerOptions{
		Level: cfg.LogLevel,
	}))

	// Derive data directory from DB_PATH.
	dbDir := filepath.Dir(cfg.DBPath)
	if dbDir == "." {
		dbDir = "."
	}
	if err := os.MkdirAll(dbDir, 0o755); err != nil {
		return fmt.Errorf("creating db directory: %w", err)
	}

	// Open admin DB (sits alongside the client DBs).
	adminDBPath := filepath.Join(dbDir, "_admin.db")
	adminDB, err := database.Open(ctx, adminDBPath)
	if err != nil {
		return fmt.Errorf("opening admin db: %w", err)
	}
	defer adminDB.Close()

	admin, err := server.NewAdminStore(ctx, adminDB)
	if err != nil {
		return fmt.Errorf("initializing admin store: %w", err)
	}
	logger.Info("admin db ready", "path", adminDBPath)

	// Create registry for per-client stores.
	clients := server.NewRegistry(dbDir)
	defer clients.Close()

	// Pre-open existing clients.
	existing, err := admin.ListClients(ctx)
	if err != nil {
		return fmt.Errorf("listing clients: %w", err)
	}
	for _, c := range existing {
		if _, err := clients.Get(ctx, c.Slug); err != nil {
			return fmt.Errorf("opening client %q: %w", c.Slug, err)
		}
		logger.Info("client db ready", "slug", c.Slug)
	}

	if err := server.SeedDemo(ctx, logger, admin, clients); err != nil {
		return fmt.Errorf("seeding demo: %w", err)
	}

	srv := server.New(cfg.HTTPAddr, logger, admin, clients, adminDB, cfg.SPADir)

	g, gctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		logger.Info("starting http server", "addr", cfg.HTTPAddr)
		return srv.Run(gctx)
	})

	g.Go(func() error {
		<-gctx.Done()
		logger.Info("shutting down http server")
		return srv.Shutdown(context.Background())
	})

	return g.Wait()
}
