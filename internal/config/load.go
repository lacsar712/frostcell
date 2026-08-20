package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Load reads configuration from environment variables layered on defaults.
func Load() (Config, error) {
	cfg := Default()

	if v := os.Getenv("FROSTCELL_LISTEN"); v != "" {
		cfg.ListenAddr = v
	}
	if v := os.Getenv("FROSTCELL_HMAC_SECRET"); v != "" {
		cfg.HMACSecret = v
	}
	if v := os.Getenv("FROSTCELL_NOTIFY_URL"); v != "" {
		cfg.NotifyURL = v
	}
	if v := os.Getenv("FROSTCELL_NOTIFY_TIMEOUT"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return Config{}, fmt.Errorf("FROSTCELL_NOTIFY_TIMEOUT: %w", err)
		}
		cfg.NotifyTimeout = d
	}
	if v := os.Getenv("FROSTCELL_WINDOW"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return Config{}, fmt.Errorf("FROSTCELL_WINDOW: %w", err)
		}
		cfg.WindowDuration = d
	}
	if v := os.Getenv("FROSTCELL_EXCURSION_RATIO"); v != "" {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return Config{}, fmt.Errorf("FROSTCELL_EXCURSION_RATIO: %w", err)
		}
		cfg.ExcursionRatio = f
	}
	if v := os.Getenv("FROSTCELL_CLEAR_RATIO"); v != "" {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return Config{}, fmt.Errorf("FROSTCELL_CLEAR_RATIO: %w", err)
		}
		cfg.ClearRatio = f
	}
	if v := os.Getenv("FROSTCELL_PENDING_WINDOWS"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return Config{}, fmt.Errorf("FROSTCELL_PENDING_WINDOWS: %w", err)
		}
		cfg.PendingWindows = n
	}
	if v := os.Getenv("FROSTCELL_CLEARING_WINDOWS"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return Config{}, fmt.Errorf("FROSTCELL_CLEARING_WINDOWS: %w", err)
		}
		cfg.ClearingWindows = n
	}

	for i := range cfg.Cells {
		cfg.Cells[i].WindowDur = cfg.WindowDuration
		if err := cfg.Cells[i].Validate(); err != nil {
			return Config{}, fmt.Errorf("cell %q: %w", cfg.Cells[i].ID, err)
		}
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}
