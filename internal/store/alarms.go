package store

import (
	"sort"
	"time"

	"github.com/lacsar712/frostcell/internal/model"
)

// AlarmIndex provides sorted alarm queries.
type AlarmIndex struct {
	store Store
}

// NewAlarmIndex wraps a store for alarm listing.
func NewAlarmIndex(s Store) *AlarmIndex {
	return &AlarmIndex{store: s}
}

// ActiveAlarms returns alarms not in Normal state.
func (a *AlarmIndex) ActiveAlarms() []model.AlarmRecord {
	all := a.store.ListAlarms()
	out := make([]model.AlarmRecord, 0)
	for _, rec := range all {
		if rec.State != model.StateNormal {
			out = append(out, rec)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].CellID < out[j].CellID
	})
	return out
}

// ByCell returns alarm for cell or zero value.
func (a *AlarmIndex) ByCell(cellID string) model.AlarmRecord {
	rec, ok := a.store.GetAlarm(cellID)
	if !ok {
		return model.AlarmRecord{CellID: cellID, State: model.StateNormal, UpdatedAt: time.Now()}
	}
	return rec
}

// SnapshotReader reads window snapshots from store with live fallback.
type SnapshotReader struct {
	store    Store
	liveFunc func(cellID string, now time.Time) (model.WindowSnapshot, error)
}

// NewSnapshotReader creates a snapshot reader.
func NewSnapshotReader(s Store, live func(string, time.Time) (model.WindowSnapshot, error)) *SnapshotReader {
	return &SnapshotReader{store: s, liveFunc: live}
}

// Get returns stored snapshot or live computation.
func (r *SnapshotReader) Get(cellID string, now time.Time) (model.WindowSnapshot, error) {
	if snap, ok := r.store.GetSnapshot(cellID); ok {
		return snap, nil
	}
	if r.liveFunc != nil {
		return r.liveFunc(cellID, now)
	}
	return model.WindowSnapshot{}, model.ErrCellNotFound
}
