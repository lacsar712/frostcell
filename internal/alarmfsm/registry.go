package alarmfsm

import "github.com/lacsar712/frostcell/internal/model"

// Registry provides read access to FSM states.
type Registry interface {
	AlarmRecord(cellID string, now interface{}) (model.AlarmRecord, error)
	ListAlarms(now interface{}) []model.AlarmRecord
}

// SnapshotReader exposes window snapshots.
type SnapshotReader interface {
	WindowSnapshot(cellID string, now interface{}) (model.WindowSnapshot, error)
}
