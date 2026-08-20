package model

import "errors"

var (
	ErrInvalidCellID     = errors.New("cell id is required")
	ErrInvalidWindow     = errors.New("window duration must be positive")
	ErrInvalidDelta      = errors.New("delta must be positive")
	ErrInvalidHysteresis = errors.New("hysteresis must be in [0, delta)")
	ErrInvalidSample     = errors.New("invalid probe sample")
	ErrCellNotFound      = errors.New("cell not found")
)
