package window_test

import (
	"testing"
	"time"

	"github.com/lacsar712/frostcell/internal/model"
	"github.com/lacsar712/frostcell/internal/window"
)

func TestSlidingWindowExpiry(t *testing.T) {
	cell := model.Cell{ID: "c1", SetpointC: -18, DeltaC: 2, WindowDur: 5 * time.Minute}
	isOver := func(t float64) bool { return t > cell.UpperLimit() }
	w := window.NewCellWindow(cell.ID, cell.WindowDur, isOver)

	base := time.Date(2026, 8, 20, 10, 0, 0, 0, time.UTC)
	w.AddSample(-17, base)
	w.AddSample(-17, base.Add(2*time.Minute))
	stats := w.AddSample(-17, base.Add(6*time.Minute))
	if stats.Count != 2 {
		t.Fatalf("expected 2 samples after expiry, got %d", stats.Count)
	}
}

func TestWindowClosed(t *testing.T) {
	cell := model.Cell{ID: "c1", WindowDur: 5 * time.Minute}
	w := window.NewCellWindow(cell.ID, cell.WindowDur, func(float64) bool { return false })
	base := time.Now()
	w.AddSample(-18, base.Add(-6*time.Minute))
	if w.IsWindowClosed(base) {
		t.Fatal("should not be closed with only old sample")
	}
	w.AddSample(-18, base.Add(-5*time.Minute))
	w.AddSample(-18, base)
	if !w.IsWindowClosed(base) {
		t.Fatal("should be closed spanning window")
	}
}

func TestManagerUnknownCell(t *testing.T) {
	m := window.NewManager()
	_, err := m.AddSample(model.ProbeSample{CellID: "x", ProbeID: "p", TempC: -18, TS: time.Now()})
	if err != model.ErrCellNotFound {
		t.Fatalf("got %v", err)
	}
}

func TestCoverageFraction(t *testing.T) {
	d := 5 * time.Minute
	now := time.Now()
	f := window.CoverageFraction(now.Add(-3*time.Minute), now, d)
	if f <= 0 || f > 1 {
		t.Fatalf("fraction %f", f)
	}
}
