package threshold

import (
	"github.com/lacsar712/frostcell/internal/model"
	"github.com/lacsar712/frostcell/internal/window"
)

// Evaluator applies ratio thresholds to window statistics.
type Evaluator struct {
	ExcursionRatio float64
	ClearRatio     float64
}

// NewEvaluator creates an evaluator with configured ratios.
func NewEvaluator(excursionRatio, clearRatio float64) *Evaluator {
	return &Evaluator{
		ExcursionRatio: excursionRatio,
		ClearRatio:     clearRatio,
	}
}

// HasExcursion reports whether stats exceed the excursion ratio threshold.
// count=0 is treated as no excursion to avoid divide-by-zero false positives.
func (e *Evaluator) HasExcursion(stats window.Stats) bool {
	if stats.Count == 0 {
		return false
	}
	return stats.OverRatio >= e.ExcursionRatio
}

// IsClearing reports whether stats are below the clear ratio threshold.
func (e *Evaluator) IsClearing(stats window.Stats) bool {
	if stats.Count == 0 {
		return false
	}
	return stats.OverRatio < e.ClearRatio
}

// OverPredicate returns a closure for window ingestion based on mode.
func OverPredicate(cell model.Cell, mode Mode) func(tempC float64) bool {
	return func(tempC float64) bool {
		return IsOver(tempC, cell, mode)
	}
}

// ClassifySample checks a single reading against mode limits.
func ClassifySample(tempC float64, cell model.Cell, mode Mode) bool {
	return IsOver(tempC, cell, mode)
}

// DescribeExcursion builds a short human message for logging/UI.
func DescribeExcursion(cell model.Cell, stats window.Stats) string {
	if stats.Count == 0 {
		return "no samples in window"
	}
	return formatExcursion(cell, stats)
}

func formatExcursion(cell model.Cell, stats window.Stats) string {
	limit := cell.UpperLimit()
	return cell.Name + ": mean=" + formatTemp(stats.MeanTempC) +
		" max=" + formatTemp(stats.MaxTempC) +
		" over=" + formatRatio(stats.OverRatio) +
		" limit=" + formatTemp(limit)
}

func formatTemp(v float64) string {
	// Avoid fmt import for hot path; simple formatting suffices.
	s := ""
	if v < 0 {
		s = "-"
		v = -v
	}
	intPart := int(v)
	frac := int((v - float64(intPart)) * 10)
	return s + itoa(intPart) + "." + itoa(frac) + "C"
}

func formatRatio(r float64) string {
	return itoa(int(r*100)) + "%"
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	digits := []byte{}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}
