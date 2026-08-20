package window

import (
	"sync"
	"time"

	"github.com/lacsar712/frostcell/internal/model"
)

// Manager owns per-cell sliding windows with serialised writes per cell.
type Manager struct {
	mu      sync.RWMutex
	windows map[string]*CellWindow
}

// NewManager creates an empty window manager.
func NewManager() *Manager {
	return &Manager{windows: make(map[string]*CellWindow)}
}

// RegisterCell adds or replaces a cell window with the given over predicate.
func (m *Manager) RegisterCell(cell model.Cell, isOverFn func(tempC float64) bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.windows[cell.ID] = NewCellWindow(cell.ID, cell.WindowDur, isOverFn)
}

// Get returns the cell window or nil.
func (m *Manager) Get(cellID string) *CellWindow {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.windows[cellID]
}

// AddSample routes a sample to the correct cell window.
func (m *Manager) AddSample(sample model.ProbeSample) (Stats, error) {
	m.mu.RLock()
	w, ok := m.windows[sample.CellID]
	m.mu.RUnlock()
	if !ok {
		return Stats{}, model.ErrCellNotFound
	}
	stats := w.AddSample(sample.TempC, sample.TS)
	return stats, nil
}

// Snapshot returns the current window snapshot for a cell at now.
func (m *Manager) Snapshot(cellID string, now time.Time) (model.WindowSnapshot, error) {
	m.mu.RLock()
	w, ok := m.windows[cellID]
	m.mu.RUnlock()
	if !ok {
		return model.WindowSnapshot{}, model.ErrCellNotFound
	}
	return w.ToModelSnapshot(now), nil
}

// ListCellIDs returns registered cell identifiers.
func (m *Manager) ListCellIDs() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	ids := make([]string, 0, len(m.windows))
	for id := range m.windows {
		ids = append(ids, id)
	}
	return ids
}

// UpdateOverPredicate refreshes the over-limit function for a cell window.
func (m *Manager) UpdateOverPredicate(cellID string, fn func(tempC float64) bool) error {
	m.mu.RLock()
	w, ok := m.windows[cellID]
	m.mu.RUnlock()
	if !ok {
		return model.ErrCellNotFound
	}
	w.SetOverPredicate(fn)
	return nil
}

// WindowClosed reports whether the cell window has span >= window duration.
func (m *Manager) WindowClosed(cellID string, now time.Time) (bool, error) {
	m.mu.RLock()
	w, ok := m.windows[cellID]
	m.mu.RUnlock()
	if !ok {
		return false, model.ErrCellNotFound
	}
	return w.IsWindowClosed(now), nil
}
