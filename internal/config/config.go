package config

import (
	"fmt"

	"github.com/lacsar712/frostcell/internal/model"
)

// Validate checks global configuration constraints.
func (c Config) Validate() error {
	if c.ListenAddr == "" {
		return fmt.Errorf("listen address is required")
	}
	if c.HMACSecret == "" {
		return fmt.Errorf("HMAC secret is required")
	}
	if c.WindowDuration <= 0 {
		return fmt.Errorf("window duration must be positive")
	}
	if c.ExcursionRatio <= 0 || c.ExcursionRatio > 1 {
		return fmt.Errorf("excursion ratio must be in (0,1]")
	}
	if c.ClearRatio < 0 || c.ClearRatio >= c.ExcursionRatio {
		return fmt.Errorf("clear ratio must be in [0, excursion ratio)")
	}
	if c.PendingWindows < 1 {
		return fmt.Errorf("pending windows must be at least 1")
	}
	if c.ClearingWindows < 1 {
		return fmt.Errorf("clearing windows must be at least 1")
	}
	if len(c.Cells) == 0 {
		return fmt.Errorf("at least one cell is required")
	}
	if c.CircuitThreshold < 1 {
		return fmt.Errorf("circuit threshold must be at least 1")
	}
	return nil
}

// CellByID returns the cell definition for id or nil.
func (c Config) CellByID(id string) *model.Cell {
	for i := range c.Cells {
		if c.Cells[i].ID == id {
			return &c.Cells[i]
		}
	}
	return nil
}

// CellMap builds an id-indexed map of configured cells.
func (c Config) CellMap() map[string]model.Cell {
	m := make(map[string]model.Cell, len(c.Cells))
	for _, cell := range c.Cells {
		m[cell.ID] = cell
	}
	return m
}
