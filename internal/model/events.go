package model

import "time"

// ProcessingResult summarizes what happened after ingesting one sample.
type ProcessingResult struct {
	CellID       string
	Snapshot     WindowSnapshot
	Alarm        AlarmRecord
	Events       []AlarmEvent
	ProcessedAt  time.Time
}

// NotifyPayload is the outbound webhook body for alarm notifications.
type NotifyPayload struct {
	Event     string     `json:"event"`
	CellID    string     `json:"cellID"`
	State     AlarmState `json:"state"`
	Timestamp time.Time  `json:"timestamp"`
	MeanTempC float64    `json:"meanTempC"`
	MaxTempC  float64    `json:"maxTempC"`
	OverRatio float64    `json:"overRatio"`
}
