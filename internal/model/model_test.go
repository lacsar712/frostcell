package model_test

import (
	"testing"
	"time"

	"github.com/lacsar712/frostcell/internal/model"
)

func TestCellValidate(t *testing.T) {
	c := model.Cell{ID: "c1", DeltaC: 2, HysteresisC: 0.5, WindowDur: time.Minute}
	if err := c.Validate(); err != nil {
		t.Fatal(err)
	}
	c.HysteresisC = 3
	if err := c.Validate(); err == nil {
		t.Fatal("expected hysteresis error")
	}
}

func TestSampleValidate(t *testing.T) {
	now := time.Now()
	s := model.ProbeSample{CellID: "c1", ProbeID: "p", TempC: -18, TS: now}
	if err := s.Validate(now); err != nil {
		t.Fatal(err)
	}
	s.TS = now.Add(2 * time.Hour)
	if err := s.Validate(now); err == nil {
		t.Fatal("future sample should fail")
	}
}

func TestCellLimits(t *testing.T) {
	c := model.Cell{SetpointC: -18, DeltaC: 2, HysteresisC: 0.5}
	if c.UpperLimit() != -16 {
		t.Fatalf("upper %f", c.UpperLimit())
	}
	if c.ClearLimit() != -16.5 {
		t.Fatalf("clear %f", c.ClearLimit())
	}
}
