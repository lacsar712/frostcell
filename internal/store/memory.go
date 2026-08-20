package store

import (
	"sync"
	"time"

	"github.com/lacsar712/frostcell/internal/model"
)

// Store persists recent window snapshots and alarm records in memory.
type Store interface {
	SaveSnapshot(snap model.WindowSnapshot)
	GetSnapshot(cellID string) (model.WindowSnapshot, bool)
	SaveAlarm(rec model.AlarmRecord)
	GetAlarm(cellID string) (model.AlarmRecord, bool)
	ListAlarms() []model.AlarmRecord
	ListSnapshots() []model.WindowSnapshot
}

// MemoryStore is a thread-safe in-memory implementation.
type MemoryStore struct {
	mu        sync.RWMutex
	snapshots map[string]model.WindowSnapshot
	alarms    map[string]model.AlarmRecord
	updated   time.Time
}

// NewMemoryStore creates an empty memory store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		snapshots: make(map[string]model.WindowSnapshot),
		alarms:    make(map[string]model.AlarmRecord),
	}
}

// SaveSnapshot upserts the latest window snapshot for a cell.
func (m *MemoryStore) SaveSnapshot(snap model.WindowSnapshot) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.snapshots[snap.CellID] = cloneSnapshot(snap)
	m.updated = time.Now()
}

// GetSnapshot returns the stored snapshot if present.
func (m *MemoryStore) GetSnapshot(cellID string) (model.WindowSnapshot, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.snapshots[cellID]
	if !ok {
		return model.WindowSnapshot{}, false
	}
	return cloneSnapshot(s), true
}

// SaveAlarm upserts alarm metadata.
func (m *MemoryStore) SaveAlarm(rec model.AlarmRecord) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.alarms[rec.CellID] = rec
	m.updated = time.Now()
}

// GetAlarm returns stored alarm record.
func (m *MemoryStore) GetAlarm(cellID string) (model.AlarmRecord, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	a, ok := m.alarms[cellID]
	return a, ok
}

// ListAlarms returns all alarm records sorted by cell id.
func (m *MemoryStore) ListAlarms() []model.AlarmRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]model.AlarmRecord, 0, len(m.alarms))
	for _, a := range m.alarms {
		out = append(out, a)
	}
	return out
}

// ListSnapshots returns all stored window snapshots.
func (m *MemoryStore) ListSnapshots() []model.WindowSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]model.WindowSnapshot, 0, len(m.snapshots))
	for _, s := range m.snapshots {
		out = append(out, cloneSnapshot(s))
	}
	return out
}

// LastUpdated returns the last mutation time.
func (m *MemoryStore) LastUpdated() time.Time {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.updated
}

func cloneSnapshot(snap model.WindowSnapshot) model.WindowSnapshot {
	if len(snap.Temps) == 0 {
		return snap
	}
	temps := make([]float64, len(snap.Temps))
	copy(temps, snap.Temps)
	snap.Temps = temps
	return snap
}
