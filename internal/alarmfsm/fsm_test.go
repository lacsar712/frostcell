package alarmfsm_test

import (
	"testing"
	"time"

	"github.com/lacsar712/frostcell/internal/alarmfsm"
	"github.com/lacsar712/frostcell/internal/model"
)

func TestFSMNormalToPending(t *testing.T) {
	fsm := alarmfsm.NewFSM("c1", 2, 1)
	now := time.Now()
	res := fsm.Step(alarmfsm.Input{
		Now:          now,
		WindowClosed: true,
		HasExcursion: true,
		Stats:        model.WindowSnapshot{OverRatio: 0.5},
	})
	if res.Record.State != model.StatePending {
		t.Fatalf("got %s", res.Record.State)
	}
}

func TestFSMIllegalNormalToActive(t *testing.T) {
	err := alarmfsm.ValidateTransition(model.StateNormal, model.StateActive)
	if err == nil {
		t.Fatal("expected error for Normal->Active")
	}
}

func TestFSMPendingToActive(t *testing.T) {
	fsm := alarmfsm.NewFSM("c1", 2, 1)
	now := time.Now()
	fsm.Step(alarmfsm.Input{Now: now, WindowClosed: true, HasExcursion: true})
	res := fsm.Step(alarmfsm.Input{Now: now.Add(time.Minute), WindowClosed: true, HasExcursion: true})
	if res.Record.State != model.StateActive {
		t.Fatalf("got %s", res.Record.State)
	}
	if len(res.Events) != 1 || res.Events[0].Type != model.EventAlarmRaised {
		t.Fatalf("events: %+v", res.Events)
	}
}

func TestFSMActiveToClearingToNormal(t *testing.T) {
	fsm := alarmfsm.NewFSM("c1", 1, 1)
	now := time.Now()
	fsm.Step(alarmfsm.Input{Now: now, WindowClosed: true, HasExcursion: true})
	fsm.Step(alarmfsm.Input{Now: now, WindowClosed: true, HasExcursion: true})
	res := fsm.Step(alarmfsm.Input{Now: now, WindowClosed: true, IsClearing: true, Stats: model.WindowSnapshot{OverRatio: 0.1}})
	if res.Record.State != model.StateClearing {
		t.Fatalf("clearing got %s", res.Record.State)
	}
	res = fsm.Step(alarmfsm.Input{Now: now, WindowClosed: true, IsClearing: true, Stats: model.WindowSnapshot{OverRatio: 0.1}})
	if res.Record.State != model.StateNormal {
		t.Fatalf("normal got %s", res.Record.State)
	}
	if len(res.Events) == 0 || res.Events[len(res.Events)-1].Type != model.EventAlarmCleared {
		t.Fatalf("events: %+v", res.Events)
	}
}
