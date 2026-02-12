package main

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/redis/go-redis/v9"
	_ "github.com/tursodatabase/go-libsql"
	"golang.org/x/sync/errgroup"

	"github.com/playperu/cityquiz/internal/config"
	"github.com/playperu/cityquiz/internal/handler"
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

	// --- SQLite (via libSQL) ---
	db, err := sql.Open("libsql", "file:"+cfg.DBPath)
	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}
	defer db.Close()

	if _, err := db.ExecContext(ctx, "PRAGMA journal_mode=WAL"); err != nil {
		return fmt.Errorf("setting WAL mode: %w", err)
	}
	if _, err := db.ExecContext(ctx, "PRAGMA busy_timeout=5000"); err != nil {
		return fmt.Errorf("setting busy timeout: %w", err)
	}
	if _, err := db.ExecContext(ctx, "PRAGMA foreign_keys=ON"); err != nil {
		return fmt.Errorf("enabling foreign keys: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("pinging database: %w", err)
	}
	logger.Info("connected to sqlite", "path", cfg.DBPath)

	// --- Redis ---
	ropt, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		return fmt.Errorf("parsing redis url: %w", err)
	}
	rdb := redis.NewClient(ropt)
	defer rdb.Close()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("pinging redis: %w", err)
	}
	logger.Info("connected to redis")

	// --- Handlers ---
	healthHandler := handler.NewHealth(dbPinger{db}, redisPinger{rdb}, logger)
	wsEchoHandler := handler.NewWSEcho(logger)

	// --- HTTP Server ---
	srv := server.New(cfg.HTTPAddr, logger, healthHandler, wsEchoHandler)

	// --- Run ---
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

// dbPinger adapts *sql.DB to handler.Pinger.
type dbPinger struct {
	db *sql.DB
}

func (d dbPinger) Ping(ctx context.Context) error {
	return d.db.PingContext(ctx)
}

// redisPinger adapts *redis.Client to handler.Pinger.
type redisPinger struct {
	client *redis.Client
}

func (r redisPinger) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}
