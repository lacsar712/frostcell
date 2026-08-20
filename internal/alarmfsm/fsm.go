package alarmfsm

import (
	"fmt"
	"time"

	"github.com/lacsar712/frostcell/internal/model"
)

// Context tracks per-cell counters used by the state machine.
type Context struct {
	ConsecutiveExcursions int
	ClearingWindows       int
	LastWindowEnd         time.Time
}

// ResetCounters clears excursion and clearing counters.
func (c *Context) ResetCounters() {
	c.ConsecutiveExcursions = 0
	c.ClearingWindows = 0
}

// Input carries evaluation results into the FSM step.
type Input struct {
	Now          time.Time
	WindowClosed bool
	HasExcursion bool
	IsClearing   bool
	Stats        model.WindowSnapshot
}

// Result describes state transition outcome.
type Result struct {
	Record     model.AlarmRecord
	Events     []model.AlarmEvent
	Changed    bool
	NewContext Context
}

// FSM enforces valid alarm transitions for one cell.
type FSM struct {
	cellID           string
	pendingWindows   int
	clearingWindows  int
	state            model.AlarmState
	ctx              Context
	lastRaised       time.Time
	lastCleared      time.Time
}

// NewFSM creates a FSM in Normal state.
func NewFSM(cellID string, pendingWindows, clearingWindows int) *FSM {
	return &FSM{
		cellID:          cellID,
		pendingWindows:  pendingWindows,
		clearingWindows: clearingWindows,
		state:           model.StateNormal,
	}
}

// State returns current alarm state.
func (f *FSM) State() model.AlarmState {
	return f.state
}

// Context returns a copy of internal counters.
func (f *FSM) Context() Context {
	return f.ctx
}

// Record builds the current alarm record view.
func (f *FSM) Record(now time.Time) model.AlarmRecord {
	return model.AlarmRecord{
		CellID:      f.cellID,
		State:       f.state,
		LastRaised:  f.lastRaised,
		LastCleared: f.lastCleared,
		UpdatedAt:   now,
	}
}

// Step advances the FSM based on window evaluation input.
func (f *FSM) Step(in Input) Result {
	prev := f.state
	events := make([]model.AlarmEvent, 0, 2)
	ctx := f.ctx

	switch f.state {
	case model.StateNormal:
		ctx = f.stepNormal(in, &events)
	case model.StatePending:
		ctx = f.stepPending(in, &events)
	case model.StateActive:
		ctx = f.stepActive(in, &events)
	case model.StateClearing:
		ctx = f.stepClearing(in, &events)
	default:
		f.state = model.StateNormal
		ctx.ResetCounters()
	}

	f.ctx = ctx
	changed := prev != f.state
	rec := f.Record(in.Now)
	if changed && len(events) == 0 {
		events = append(events, f.buildEvent(in))
	}
	return Result{Record: rec, Events: events, Changed: changed, NewContext: ctx}
}

func (f *FSM) stepNormal(in Input, events *[]model.AlarmEvent) Context {
	ctx := f.ctx
	if in.WindowClosed && in.HasExcursion {
		f.state = model.StatePending
		ctx.ConsecutiveExcursions = 1
		ctx.ClearingWindows = 0
	}
	return ctx
}

func (f *FSM) stepPending(in Input, events *[]model.AlarmEvent) Context {
	ctx := f.ctx
	// Consecutive excursion windows only count once the window has closed;
	// mid-window samples must not pre-climb the counter toward Active.
	if in.WindowClosed && in.HasExcursion {
		ctx.ConsecutiveExcursions++
		if ctx.ConsecutiveExcursions >= f.pendingWindows {
			f.state = model.StateActive
			f.lastRaised = in.Now
			*events = append(*events, f.event(model.EventAlarmRaised, in))
			ctx.ConsecutiveExcursions = 0
		}
	} else if in.WindowClosed {
		f.state = model.StateNormal
		ctx.ResetCounters()
	}
	return ctx
}

func (f *FSM) stepActive(in Input, events *[]model.AlarmEvent) Context {
	ctx := f.ctx
	f.lastRaised = in.Now
	if in.WindowClosed && in.IsClearing {
		f.state = model.StateClearing
		ctx.ClearingWindows = 1
		ctx.ConsecutiveExcursions = 0
	}
	return ctx
}

func (f *FSM) stepClearing(in Input, events *[]model.AlarmEvent) Context {
	ctx := f.ctx
	if !in.WindowClosed {
		return ctx
	}
	if in.IsClearing {
		ctx.ClearingWindows++
		if ctx.ClearingWindows >= f.clearingWindows {
			f.state = model.StateNormal
			f.lastCleared = in.Now
			*events = append(*events, f.event(model.EventAlarmCleared, in))
			ctx.ResetCounters()
		}
	} else {
		f.state = model.StateActive
		ctx.ClearingWindows = 0
	}
	return ctx
}

func (f *FSM) event(typ string, in Input) model.AlarmEvent {
	return model.AlarmEvent{
		Type:      typ,
		CellID:    f.cellID,
		State:     f.state,
		Timestamp: in.Now,
		MeanTempC: in.Stats.MeanTempC,
		MaxTempC:  in.Stats.MaxTempC,
		OverRatio: in.Stats.OverRatio,
	}
}

func (f *FSM) buildEvent(in Input) model.AlarmEvent {
	typ := model.EventAlarmRaised
	if f.state == model.StateNormal {
		typ = model.EventAlarmCleared
	}
	return f.event(typ, in)
}

// CanTransition reports whether a direct transition is legal.
func CanTransition(from, to model.AlarmState) bool {
	switch from {
	case model.StateNormal:
		return to == model.StatePending
	case model.StatePending:
		return to == model.StateActive || to == model.StateNormal
	case model.StateActive:
		return to == model.StateActive || to == model.StateClearing
	case model.StateClearing:
		return to == model.StateNormal || to == model.StateActive
	default:
		return false
	}
}

// ValidateTransition returns an error for illegal jumps such as Normal→Active.
func ValidateTransition(from, to model.AlarmState) error {
	if from == model.StateNormal && to == model.StateActive {
		return fmt.Errorf("illegal transition: %s -> %s (must pass Pending)", from, to)
	}
	if !CanTransition(from, to) && from != to {
		return fmt.Errorf("illegal transition: %s -> %s", from, to)
	}
	return nil
}
