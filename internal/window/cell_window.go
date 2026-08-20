package window

import (
	"sync"
	"time"

	"github.com/lacsar712/frostcell/internal/model"
)

// CellWindow maintains a mutex-protected sliding window for one cell.
type CellWindow struct {
	cellID     string
	windowDur  time.Duration
	buf        *ringBuffer
	mu         sync.Mutex
	lastEnd    time.Time
	isOverFn   func(tempC float64) bool
}

// NewCellWindow creates an empty window for the given cell duration.
func NewCellWindow(cellID string, windowDur time.Duration, isOverFn func(tempC float64) bool) *CellWindow {
	return &CellWindow{
		cellID:    cellID,
		windowDur: windowDur,
		buf:       newRingBuffer(256),
		isOverFn:  isOverFn,
	}
}

// SetOverPredicate updates the function used to classify over-limit samples.
func (w *CellWindow) SetOverPredicate(fn func(tempC float64) bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.isOverFn = fn
}

// AddSample ingests one probe reading at ts and returns updated stats.
func (w *CellWindow) AddSample(tempC float64, ts time.Time) Stats {
	w.mu.Lock()
	defer w.mu.Unlock()

	cutoff := ts.Add(-w.windowDur)
	w.buf.append(samplePoint{tempC: tempC, ts: ts}, cutoff)
	w.lastEnd = ts

	return computeStats(w.buf, w.isOverFn)
}

// Snapshot returns current stats without mutating state.
func (w *CellWindow) Snapshot(now time.Time) Stats {
	w.mu.Lock()
	defer w.mu.Unlock()

	cutoff := now.Add(-w.windowDur)
	w.buf.expireBefore(cutoff)
	if !w.lastEnd.IsZero() {
		w.lastEnd = now
	}

	return computeStats(w.buf, w.isOverFn)
}

// ToModelSnapshot builds a model snapshot at the given time.
func (w *CellWindow) ToModelSnapshot(now time.Time) model.WindowSnapshot {
	stats := w.Snapshot(now)
	end := now
	if !w.lastEnd.IsZero() {
		end = w.lastEnd
	}
	return stats.ToSnapshot(w.cellID, end, w.windowDur)
}

// WindowDuration returns configured duration.
func (w *CellWindow) WindowDuration() time.Duration {
	return w.windowDur
}

// SampleCount returns buffered sample count (for diagnostics).
func (w *CellWindow) SampleCount() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.buf.len()
}

// OldestSample returns the oldest buffered timestamp.
func (w *CellWindow) OldestSample() time.Time {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.buf.oldest()
}

// NewestSample returns the newest buffered timestamp.
func (w *CellWindow) NewestSample() time.Time {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.buf.newest()
}

// IsWindowClosed reports whether the oldest sample spans the full window duration.
// The caller's processing clock (now) is the sole reference — not sample timestamps.
func (w *CellWindow) IsWindowClosed(now time.Time) bool {
	w.mu.Lock()
	defer w.mu.Unlock()

	cutoff := now.Add(-w.windowDur)
	w.buf.expireBefore(cutoff)
	if w.buf.len() == 0 {
		return false
	}
	oldest := w.buf.oldest()
	return !oldest.After(cutoff)
}

// SampleTemps returns a copy of buffered temperatures (never aliases internal storage).
func (w *CellWindow) SampleTemps() []float64 {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.buf.copyTemps()
}
