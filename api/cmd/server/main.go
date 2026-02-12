package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"

	"github.com/playperu/cityquiz/internal/config"
	"github.com/playperu/cityquiz/internal/database"
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

	db, err := database.Open(ctx, cfg.DBPath)
	if err != nil {
		return fmt.Errorf("connecting to sqlite: %w", err)
	}
	defer db.Close()

	if err := migrations.Run(db); err != nil {
		return fmt.Errorf("running migrations: %w", err)
	}
	logger.Info("connected to sqlite", "path", cfg.DBPath)

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

	srv := server.New(cfg.HTTPAddr, logger, db, rdb)

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
