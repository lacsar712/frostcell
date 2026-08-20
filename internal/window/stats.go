package window

import (
	"time"

	"github.com/lacsar712/frostcell/internal/model"
)

// Stats aggregates sliding-window metrics.
type Stats struct {
	Count     int
	OverCount int
	MaxTempC  float64
	MeanTempC float64
	OverRatio float64
}

// computeStats calculates window statistics using the provided over predicate.
func computeStats(buf *ringBuffer, isOver func(tempC float64) bool) Stats {
	count := buf.len()
	if count == 0 {
		return Stats{}
	}

	var sum float64
	var overCount int
	maxTemp := buf.points[0].tempC

	for _, p := range buf.points {
		sum += p.tempC
		if p.tempC > maxTemp {
			maxTemp = p.tempC
		}
		if isOver(p.tempC) {
			overCount++
		}
	}

	ratio := float64(overCount) / float64(count)
	return Stats{
		Count:     count,
		OverCount: overCount,
		MaxTempC:  maxTemp,
		MeanTempC: sum / float64(count),
		OverRatio: ratio,
	}
}

// ToSnapshot converts stats into an API snapshot at windowEnd.
func (s Stats) ToSnapshot(cellID string, windowEnd time.Time, windowDur time.Duration) model.WindowSnapshot {
	return model.WindowSnapshot{
		CellID:    cellID,
		Count:     s.Count,
		OverCount: s.OverCount,
		MaxTempC:  s.MaxTempC,
		MeanTempC: s.MeanTempC,
		OverRatio: s.OverRatio,
		WindowEnd: windowEnd,
		WindowDur: windowDur,
	}
}
