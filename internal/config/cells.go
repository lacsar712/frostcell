package config

import (
	"encoding/json"
	"os"

	"github.com/lacsar712/frostcell/internal/model"
)

// LoadCellsFile merges cells from a JSON file if FROSTCELL_CELLS_FILE is set.
func LoadCellsFile(cfg *Config) error {
	path := os.Getenv("FROSTCELL_CELLS_FILE")
	if path == "" {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var cells []model.Cell
	if err := json.Unmarshal(data, &cells); err != nil {
		return err
	}
	for i := range cells {
		if cells[i].WindowDur == 0 {
			cells[i].WindowDur = cfg.WindowDuration
		}
	}
	cfg.Cells = cells
	return nil
}

// ApplyWindowDuration ensures every cell uses the global window duration.
func (c *Config) ApplyWindowDuration() {
	for i := range c.Cells {
		c.Cells[i].WindowDur = c.WindowDuration
	}
}
