package window

import "time"

// PrunePolicy describes how aggressively expired samples are removed.
type PrunePolicy struct {
	Grace time.Duration
}

// DefaultPrunePolicy returns zero grace (strict window boundary).
func DefaultPrunePolicy() PrunePolicy {
	return PrunePolicy{}
}

// Cutoff computes the oldest timestamp kept for window ending at ts.
func (p PrunePolicy) Cutoff(ts time.Time, windowDur time.Duration) time.Time {
	return ts.Add(-windowDur).Add(-p.Grace)
}

// CoverageFraction estimates how much of the window duration is covered by samples.
func CoverageFraction(oldest, newest time.Time, windowDur time.Duration) float64 {
	if oldest.IsZero() || newest.IsZero() || windowDur <= 0 {
		return 0
	}
	span := newest.Sub(oldest)
	if span <= 0 {
		return 0
	}
	f := float64(span) / float64(windowDur)
	if f > 1 {
		return 1
	}
	return f
}

// WindowClosedAt reports whether oldest sample is at or before cutoff.
func WindowClosedAt(oldest time.Time, now time.Time, windowDur time.Duration) bool {
	if oldest.IsZero() {
		return false
	}
	cutoff := now.Add(-windowDur)
	return !oldest.After(cutoff)
}
