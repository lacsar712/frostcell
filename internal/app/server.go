package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/lacsar712/frostcell/internal/config"
)

// RunFromEnv loads config and runs until SIGINT/SIGTERM.
func RunFromEnv() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	return RunWithConfig(cfg)
}

// RunWithConfig starts the application with explicit configuration.
func RunWithConfig(cfg config.Config) error {
	application, err := New(cfg)
	if err != nil {
		return fmt.Errorf("init app: %w", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	return application.Run(ctx)
}
