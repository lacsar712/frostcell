package threshold_test

import (
	"testing"

	"github.com/lacsar712/frostcell/internal/threshold"
	"github.com/lacsar712/frostcell/internal/window"
)

func TestHasExcursionZeroCount(t *testing.T) {
	e := threshold.NewEvaluator(0.4, 0.2)
	if e.HasExcursion(window.Stats{Count: 0, OverRatio: 0}) {
		t.Fatal("count==0 must not be treated as excursion")
	}
	if e.HasExcursion(window.Stats{Count: 0, OverRatio: 1}) {
		t.Fatal("count==0 must not be treated as excursion even if OverRatio looks high")
	}
}
