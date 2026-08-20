package threshold_test

import (
	"testing"

	"github.com/lacsar712/frostcell/internal/model"
	"github.com/lacsar712/frostcell/internal/threshold"
	"github.com/lacsar712/frostcell/internal/window"
)

func TestIsOverModes(t *testing.T) {
	cell := model.Cell{SetpointC: -18, DeltaC: 2, HysteresisC: 0.5}
	if !threshold.IsOver(-15.9, cell, threshold.ModeActive) {
		t.Fatal("expected over active limit")
	}
	if threshold.IsOver(-16.1, cell, threshold.ModeActive) {
		t.Fatal("expected under active limit")
	}
	if !threshold.IsOver(-16.0, cell, threshold.ModeClearing) {
		t.Fatal("expected over clearing limit")
	}
	if threshold.IsOver(-16.6, cell, threshold.ModeClearing) {
		t.Fatal("expected under clearing limit")
	}
}

func TestEvaluatorZeroCount(t *testing.T) {
	e := threshold.NewEvaluator(0.4, 0.2)
	stats := window.Stats{Count: 0}
	if e.HasExcursion(stats) {
		t.Fatal("count=0 must not excursion")
	}
	if e.IsClearing(stats) {
		t.Fatal("count=0 must not clearing")
	}
}

func TestEvaluatorRatios(t *testing.T) {
	e := threshold.NewEvaluator(0.4, 0.2)
	stats := window.Stats{Count: 10, OverCount: 4, OverRatio: 0.4}
	if !e.HasExcursion(stats) {
		t.Fatal("expected excursion at 40%")
	}
	stats = window.Stats{Count: 10, OverCount: 1, OverRatio: 0.1}
	if !e.IsClearing(stats) {
		t.Fatal("expected clearing below 20%")
	}
}
