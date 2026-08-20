package ingest

import (
	"encoding/json"
	"time"
)

// rawSample is the JSON wire format with flexible timestamp parsing.
type rawSample struct {
	CellID  string          `json:"cellID"`
	ProbeID string          `json:"probeID"`
	TempC   float64         `json:"tempC"`
	TS      json.RawMessage `json:"ts"`
}

// ParseTimestamp supports RFC3339 strings and unix seconds/millis numbers.
func ParseTimestamp(raw json.RawMessage) (time.Time, error) {
	if len(raw) == 0 {
		return time.Time{}, errBadTimestamp
	}

	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
			return t, nil
		}
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			return t, nil
		}
		return time.Time{}, errBadTimestamp
	}

	var f float64
	if err := json.Unmarshal(raw, &f); err == nil {
		if f > 1e12 {
			return time.UnixMilli(int64(f)), nil
		}
		return time.Unix(int64(f), 0), nil
	}

	var i int64
	if err := json.Unmarshal(raw, &i); err == nil {
		if i > 1e12 {
			return time.UnixMilli(i), nil
		}
		return time.Unix(i, 0), nil
	}

	return time.Time{}, errBadTimestamp
}
