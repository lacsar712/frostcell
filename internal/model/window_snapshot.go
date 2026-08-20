package model

import "time"

// WindowSnapshot is a point-in-time view of sliding window statistics.
type WindowSnapshot struct {
	CellID    string    `json:"cellID"`
	Count     int       `json:"count"`
	OverCount int       `json:"overCount"`
	MaxTempC  float64   `json:"maxTempC"`
	MeanTempC float64   `json:"meanTempC"`
	OverRatio float64   `json:"overRatio"`
	WindowEnd time.Time `json:"windowEnd"`
	WindowDur time.Duration `json:"windowDur"`
	// Temps is a defensive copy of sample temperatures in the window.
	Temps []float64 `json:"temps,omitempty"`
}

// HasData reports whether the window contains at least one sample.
func (w WindowSnapshot) HasData() bool {
	return w.Count > 0
}
