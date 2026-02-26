package server

import (
	"context"
	"log/slog"
)

// SeedDemo creates the demo client, scenario, and game if no clients exist.
// Idempotent: does nothing if clients already exist.
func SeedDemo(ctx context.Context, logger *slog.Logger, admin *AdminDocStore, clients *Registry) error {
	existing, err := admin.ListClients(ctx)
	if err != nil {
		return err
	}
	if len(existing) > 0 {
		return nil
	}

	sc, err := admin.SeedDemoScenario(ctx)
	if err != nil {
		return err
	}

	if err := admin.CreateClient(ctx, "demo", "Demo"); err != nil {
		return err
	}
	store, err := clients.Create(ctx, "demo")
	if err != nil {
		return err
	}
	if sc != nil {
		if err := store.SeedDemoGame(ctx, sc); err != nil {
			return err
		}
	}

	logger.Info("demo client created and seeded")
	return nil
}
