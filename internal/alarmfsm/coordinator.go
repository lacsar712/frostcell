package alarmfsm

import (
	"sync"
	"time"

	"github.com/lacsar712/frostcell/internal/model"
	"github.com/lacsar712/frostcell/internal/threshold"
	"github.com/lacsar712/frostcell/internal/window"
)

// Coordinator wires windows, thresholds, and FSM instances per cell.
type Coordinator struct {
	mu               sync.RWMutex
	fsms             map[string]*FSM
	policies         map[string]*threshold.Policy
	windows          *window.Manager
	pendingWindows   int
	clearingWindows  int
}

// NewCoordinator creates a coordinator bound to a window manager.
func NewCoordinator(wm *window.Manager, pendingWindows, clearingWindows int, eval *threshold.Evaluator, cells []model.Cell) *Coordinator {
	c := &Coordinator{
		fsms:            make(map[string]*FSM),
		policies:        make(map[string]*threshold.Policy),
		windows:         wm,
		pendingWindows:  pendingWindows,
		clearingWindows: clearingWindows,
	}
	for _, cell := range cells {
		c.RegisterCell(cell, eval)
	}
	return c
}

// RegisterCell initialises FSM, policy, and window for a cell.
func (c *Coordinator) RegisterCell(cell model.Cell, eval *threshold.Evaluator) {
	c.mu.Lock()
	defer c.mu.Unlock()
	policy := threshold.NewPolicy(cell, eval)
	c.policies[cell.ID] = policy
	c.fsms[cell.ID] = NewFSM(cell.ID, c.pendingWindows, c.clearingWindows)
	c.windows.RegisterCell(cell, policy.ActiveOverPredicate())
}

// ProcessSample ingests a sample and runs FSM evaluation when the window closes.
func (c *Coordinator) ProcessSample(sample model.ProbeSample, now time.Time) (model.ProcessingResult, error) {
	c.mu.RLock()
	fsm, ok := c.fsms[sample.CellID]
	policy, pok := c.policies[sample.CellID]
	c.mu.RUnlock()
	if !ok || !pok {
		return model.ProcessingResult{}, model.ErrCellNotFound
	}

	c.syncOverPredicate(sample.CellID, fsm.State(), policy)

	stats, err := c.windows.AddSample(sample)
	if err != nil {
		return model.ProcessingResult{}, err
	}

	snap := stats.ToSnapshot(sample.CellID, sample.TS, policy.Cell.WindowDur)
	closed, err := c.windows.WindowClosed(sample.CellID, now)
	if err != nil {
		return model.ProcessingResult{}, err
	}

	prev := fsm.State()
	in := c.buildInput(fsm, policy, snap, closed, now)
	result := fsm.Step(in)
	if err := ValidateTransition(prev, result.Record.State); err != nil {
		return model.ProcessingResult{}, err
	}

	return model.ProcessingResult{
		CellID:      sample.CellID,
		Snapshot:    snap,
		Alarm:       result.Record,
		Events:      result.Events,
		ProcessedAt: now,
	}, nil
}

func (c *Coordinator) syncOverPredicate(cellID string, state model.AlarmState, policy *threshold.Policy) {
	var fn func(float64) bool
	switch state {
	case model.StateClearing, model.StateActive:
		fn = policy.ClearingOverPredicate()
	default:
		fn = policy.ActiveOverPredicate()
	}
	_ = c.windows.UpdateOverPredicate(cellID, fn)
}

func (c *Coordinator) buildInput(fsm *FSM, policy *threshold.Policy, snap model.WindowSnapshot, closed bool, now time.Time) Input {
	stats := window.Stats{
		Count:     snap.Count,
		OverCount: snap.OverCount,
		MaxTempC:  snap.MaxTempC,
		MeanTempC: snap.MeanTempC,
		OverRatio: snap.OverRatio,
	}

	hasExcursion := policy.ShouldEnterPending(stats, closed)
	isClearing := policy.ShouldEnterClearing(stats, closed)

	if fsm.State() == model.StateClearing {
		c.syncOverPredicate(snap.CellID, model.StateClearing, policy)
	}

	return Input{
		Now:          now,
		WindowClosed: closed,
		HasExcursion: hasExcursion,
		IsClearing:   isClearing,
		Stats:        snap,
	}
}

// AlarmRecord returns current alarm state for a cell.
func (c *Coordinator) AlarmRecord(cellID string, now time.Time) (model.AlarmRecord, error) {
	c.mu.RLock()
	fsm, ok := c.fsms[cellID]
	c.mu.RUnlock()
	if !ok {
		return model.AlarmRecord{}, model.ErrCellNotFound
	}
	return fsm.Record(now), nil
}

// ListAlarms returns all alarm records.
func (c *Coordinator) ListAlarms(now time.Time) []model.AlarmRecord {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]model.AlarmRecord, 0, len(c.fsms))
	for _, fsm := range c.fsms {
		out = append(out, fsm.Record(now))
	}
	return out
}

// WindowSnapshot returns the live window for a cell.
func (c *Coordinator) WindowSnapshot(cellID string, now time.Time) (model.WindowSnapshot, error) {
	return c.windows.Snapshot(cellID, now)
}
