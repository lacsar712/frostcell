package threshold

import (
	"github.com/lacsar712/frostcell/internal/model"
	"github.com/lacsar712/frostcell/internal/window"
)

// Policy combines cell limits and ratio thresholds for alarm decisions.
type Policy struct {
	Cell       model.Cell
	Evaluator  *Evaluator
	ActiveMode Mode
	ClearMode  Mode
}

// NewPolicy builds a policy for one cell.
func NewPolicy(cell model.Cell, eval *Evaluator) *Policy {
	return &Policy{
		Cell:       cell,
		Evaluator:  eval,
		ActiveMode: ModeActive,
		ClearMode:  ModeClearing,
	}
}

// ActiveOverPredicate returns the predicate used while monitoring excursions.
func (p *Policy) ActiveOverPredicate() func(float64) bool {
	return OverPredicate(p.Cell, p.ActiveMode)
}

// ClearingOverPredicate returns the predicate used while clearing alarms.
func (p *Policy) ClearingOverPredicate() func(float64) bool {
	return OverPredicate(p.Cell, p.ClearMode)
}

// ShouldEnterPending evaluates whether a closed window warrants Pending.
func (p *Policy) ShouldEnterPending(stats window.Stats, windowClosed bool) bool {
	if !windowClosed {
		return false
	}
	return p.Evaluator.HasExcursion(stats)
}

// ShouldActivate evaluates consecutive excursion windows for Active.
func (p *Policy) ShouldActivate(consecutiveExcursions int, pendingWindows int) bool {
	return consecutiveExcursions >= pendingWindows
}

// ShouldEnterClearing evaluates whether stats support Clearing from Active.
func (p *Policy) ShouldEnterClearing(stats window.Stats, windowClosed bool) bool {
	if !windowClosed || stats.Count == 0 {
		return false
	}
	return p.Evaluator.IsClearing(stats)
}

// ShouldReturnNormal evaluates clearing confirmation windows.
func (p *Policy) ShouldReturnNormal(clearingWindows int, required int) bool {
	return clearingWindows >= required
}

// RecomputeStatsWithMode recalculates over counts using a different mode.
func (p *Policy) RecomputeStatsWithMode(stats window.Stats, samples []float64, mode Mode) window.Stats {
	if len(samples) == 0 {
		return stats
	}
	var over int
	var max float64
	var sum float64
	for i, t := range samples {
		sum += t
		if i == 0 || t > max {
			max = t
		}
		if IsOver(t, p.Cell, mode) {
			over++
		}
	}
	count := len(samples)
	return window.Stats{
		Count:     count,
		OverCount: over,
		MaxTempC:  max,
		MeanTempC: sum / float64(count),
		OverRatio: float64(over) / float64(count),
	}
}
