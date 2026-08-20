package config

import (
	"time"

	"github.com/lacsar712/frostcell/internal/model"
)

// Config holds runtime configuration for the frostcell service.
type Config struct {
	ListenAddr       string
	HMACSecret       string
	NotifyURL        string
	NotifyTimeout    time.Duration
	WindowDuration   time.Duration
	ExcursionRatio   float64
	ClearRatio       float64
	PendingWindows   int
	ClearingWindows  int
	Cells            []model.Cell
	CircuitThreshold int
	CircuitCooldown  time.Duration
}

// Default returns production-like defaults for a single demo cell.
func Default() Config {
	return Config{
		ListenAddr:       ":8080",
		HMACSecret:       "change-me-in-production",
		NotifyURL:        "",
		NotifyTimeout:    5 * time.Second,
		WindowDuration:   5 * time.Minute,
		ExcursionRatio:   0.40,
		ClearRatio:       0.20,
		PendingWindows:   2,
		ClearingWindows:  1,
		CircuitThreshold: 5,
		CircuitCooldown:  30 * time.Second,
		Cells: []model.Cell{
			{
				ID:          "cell-01",
				Name:        "冷藏厢 A",
				SetpointC:   -18.0,
				DeltaC:      2.0,
				HysteresisC: 0.5,
				WindowDur:   5 * time.Minute,
			},
		},
	}
}
