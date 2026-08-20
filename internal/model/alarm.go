package model

import "time"

// AlarmState is the lifecycle state of a cell alarm.
type AlarmState string

const (
	StateNormal   AlarmState = "Normal"
	StatePending  AlarmState = "Pending"
	StateActive   AlarmState = "Active"
	StateClearing AlarmState = "Clearing"
)

// AlarmRecord tracks alarm metadata for a cell.
type AlarmRecord struct {
	CellID     string     `json:"cellID"`
	State      AlarmState `json:"state"`
	LastRaised time.Time  `json:"lastRaised,omitempty"`
	LastCleared time.Time `json:"lastCleared,omitempty"`
	UpdatedAt  time.Time  `json:"updatedAt"`
	Message    string     `json:"message,omitempty"`
}

// AlarmEvent is emitted when an alarm is raised or cleared.
type AlarmEvent struct {
	Type      string     `json:"type"`
	CellID    string     `json:"cellID"`
	State     AlarmState `json:"state"`
	Timestamp time.Time  `json:"timestamp"`
	MeanTempC float64    `json:"meanTempC"`
	MaxTempC  float64    `json:"maxTempC"`
	OverRatio float64    `json:"overRatio"`
}

const (
	EventAlarmRaised  = "AlarmRaised"
	EventAlarmCleared = "AlarmCleared"
)
