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

	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"

	"github.com/playperu/cityquiz/internal/config"
	"github.com/playperu/cityquiz/internal/database"
	"github.com/playperu/cityquiz/internal/handler/health"
	"github.com/playperu/cityquiz/internal/handler/wsecho"
	"github.com/playperu/cityquiz/internal/migrations"
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

	// --- SQLite ---
	db, err := database.Open(ctx, cfg.DBPath)
	if err != nil {
		return fmt.Errorf("connecting to sqlite: %w", err)
	}
	defer db.Close()

	if err := migrations.Run(db); err != nil {
		return fmt.Errorf("running migrations: %w", err)
	}
	logger.Info("connected to sqlite", "path", cfg.DBPath)

	// --- Redis ---
	rdb, err := openRedis(ctx, cfg.RedisURL)
	if err != nil {
		return fmt.Errorf("connecting to redis: %w", err)
	}
	defer rdb.Close()
	logger.Info("connected to redis")

	// --- HTTP Server ---
	srv := server.New(cfg.HTTPAddr, logger, func(r chi.Router) {
		r.Mount("/healthz", health.NewHandler(logger, map[string]health.Checker{
			"sqlite": dbChecker{db},
			"redis":  redisChecker{rdb},
		}).Routes())
		r.Mount("/ws", wsecho.NewHandler(logger).Routes())
	})

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

func openRedis(ctx context.Context, rawURL string) (*redis.Client, error) {
	opt, err := redis.ParseURL(rawURL)
	if err != nil {
		return nil, fmt.Errorf("parsing redis url: %w", err)
	}
	rdb := redis.NewClient(opt)
	if err := rdb.Ping(ctx).Err(); err != nil {
		rdb.Close()
		return nil, fmt.Errorf("pinging redis: %w", err)
	}
	return rdb, nil
}

// dbChecker adapts *sql.DB to health.Checker.
type dbChecker struct{ db *sql.DB }

func (d dbChecker) Check(ctx context.Context) error { return d.db.PingContext(ctx) }

// redisChecker adapts *redis.Client to health.Checker.
type redisChecker struct{ client *redis.Client }

func (r redisChecker) Check(ctx context.Context) error { return r.client.Ping(ctx).Err() }
