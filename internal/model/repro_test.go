package model_test

import (
	"testing"

	"github.com/lacsar712/frostcell/internal/model"
	"github.com/lacsar712/frostcell/internal/threshold"
)

func TestClearLimitAppliesHysteresis(t *testing.T) {
	c := model.Cell{SetpointC: -18, DeltaC: 2, HysteresisC: 0.5}
	if c.UpperLimit() != -16 {
		t.Fatalf("UpperLimit=%v", c.UpperLimit())
	}
	if c.ClearLimit() != -16.5 {
		t.Fatalf("ClearLimit must apply hysteresis, got %v want -16.5", c.ClearLimit())
	}
	// Midpoint between clear and upper must be over only in clearing mode.
	mid := -16.2
	if threshold.IsOver(mid, c, threshold.ModeActive) {
		t.Fatalf("%v must not be over UpperLimit", mid)
	}
	if !threshold.IsOver(mid, c, threshold.ModeClearing) {
		t.Fatalf("%v must be over ClearLimit in clearing mode", mid)
	}
}
