package model

import "time"

// Cell describes a monitored cold-storage zone with temperature limits.
type Cell struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	SetpointC   float64 `json:"setpointC"`
	DeltaC      float64 `json:"deltaC"`
	HysteresisC float64 `json:"hysteresisC"`
	WindowDur   time.Duration
}

// UpperLimit returns the active excursion threshold (setpoint + delta).
func (c Cell) UpperLimit() float64 {
	return c.SetpointC + c.DeltaC
}

// ClearLimit returns the clearing threshold with hysteresis applied.
// It sits HysteresisC below UpperLimit so the temperature must cool past the
// excursion line by the hysteresis margin before an alarm can clear. Without
// this margin the reading dithers at the edge and the alarm churns between
// Active and Clearing.
func (c Cell) ClearLimit() float64 {
	return c.UpperLimit() - c.HysteresisC
}

// Validate checks that cell parameters are sane for monitoring.
func (c Cell) Validate() error {
	if c.ID == "" {
		return ErrInvalidCellID
	}
	if c.WindowDur <= 0 {
		return ErrInvalidWindow
	}
	if c.DeltaC <= 0 {
		return ErrInvalidDelta
	}
	if c.HysteresisC < 0 || c.HysteresisC >= c.DeltaC {
		return ErrInvalidHysteresis
	}
	return nil
}
