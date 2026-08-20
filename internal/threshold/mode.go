package threshold

import "github.com/lacsar712/frostcell/internal/model"

// Mode selects which limit line applies when classifying samples.
type Mode int

const (
	ModeActive Mode = iota
	ModeClearing
)

// String returns a human-readable mode name.
func (m Mode) String() string {
	switch m {
	case ModeActive:
		return "active"
	case ModeClearing:
		return "clearing"
	default:
		return "unknown"
	}
}

// IsOver returns true when temp exceeds the limit for the given mode.
func IsOver(tempC float64, cell model.Cell, mode Mode) bool {
	switch mode {
	case ModeActive:
		return tempC > cell.UpperLimit()
	case ModeClearing:
		return tempC > cell.ClearLimit()
	default:
		return tempC > cell.UpperLimit()
	}
}

// LimitForMode returns the numeric threshold for the mode.
func LimitForMode(cell model.Cell, mode Mode) float64 {
	switch mode {
	case ModeActive:
		return cell.UpperLimit()
	case ModeClearing:
		return cell.ClearLimit()
	default:
		return cell.UpperLimit()
	}
}
