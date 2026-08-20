package model

import "time"

// ProbeSample is a single temperature reading from a probe gateway.
type ProbeSample struct {
	CellID  string    `json:"cellID"`
	ProbeID string    `json:"probeID"`
	TempC   float64   `json:"tempC"`
	TS      time.Time `json:"ts"`
}

// Validate ensures required fields are present and timestamps are usable.
func (s ProbeSample) Validate(now time.Time) error {
	if s.CellID == "" || s.ProbeID == "" {
		return ErrInvalidSample
	}
	if s.TS.IsZero() {
		return ErrInvalidSample
	}
	// Reject samples more than one hour in the future or older than 24h.
	if s.TS.After(now.Add(time.Hour)) {
		return ErrInvalidSample
	}
	if s.TS.Before(now.Add(-24 * time.Hour)) {
		return ErrInvalidSample
	}
	return nil
}
