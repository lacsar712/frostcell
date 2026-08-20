package alarmfsm_test

import (
	"strings"
	"testing"
	"time"

	"github.com/lacsar712/frostcell/internal/alarmfsm"
	"github.com/lacsar712/frostcell/internal/model"
	"github.com/lacsar712/frostcell/internal/threshold"
	"github.com/lacsar712/frostcell/internal/window"
)

func testCell(id string, dur time.Duration) model.Cell {
	return model.Cell{
		ID:          id,
		Name:        id,
		SetpointC:   -18,
		DeltaC:      2,
		HysteresisC: 0.5,
		WindowDur:   dur,
	}
}

// driveClosedWindow seeds an oldest sample then evaluates at start+dur so the window is closed.
func driveClosedWindow(t *testing.T, coord *alarmfsm.Coordinator, cellID string, temp float64, start time.Time, dur time.Duration) model.ProcessingResult {
	t.Helper()
	_, err := coord.ProcessSample(model.ProbeSample{
		CellID: cellID, ProbeID: "p1", TempC: temp, TS: start,
	}, start)
	if err != nil {
		t.Fatal(err)
	}
	res, err := coord.ProcessSample(model.ProbeSample{
		CellID: cellID, ProbeID: "p1", TempC: temp, TS: start.Add(dur),
	}, start.Add(dur))
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func TestWindowClosedUsesProcessingClock(t *testing.T) {
	dur := 5 * time.Minute
	cell := testCell("c-clock", dur)
	wm := window.NewManager()
	eval := threshold.NewEvaluator(0.4, 0.2)
	coord := alarmfsm.NewCoordinator(wm, 2, 1, eval, []model.Cell{cell})

	base := time.Date(2026, 8, 20, 10, 0, 0, 0, time.UTC)
	_, err := coord.ProcessSample(model.ProbeSample{
		CellID: cell.ID, ProbeID: "p1", TempC: -15, TS: base,
	}, base)
	if err != nil {
		t.Fatal(err)
	}

	// Probe timestamp is 6 minutes ahead of the processing clock.
	lateTS := base.Add(6 * time.Minute)
	now := base.Add(2 * time.Minute)
	res, err := coord.ProcessSample(model.ProbeSample{
		CellID: cell.ID, ProbeID: "p1", TempC: -15, TS: lateTS,
	}, now)
	if err != nil {
		t.Fatal(err)
	}
	if res.Alarm.State != model.StateNormal {
		t.Fatalf("processing clock has not spanned the window; want Normal, got %s (window must not close on sample.TS)", res.Alarm.State)
	}
	closed, err := wm.WindowClosed(cell.ID, now)
	if err != nil {
		t.Fatal(err)
	}
	if closed {
		t.Fatal("IsWindowClosed must use processing now, not last sample timestamp")
	}
}

func TestActiveStateUsesActiveOverPredicate(t *testing.T) {
	dur := time.Minute
	cell := testCell("c-active", dur)
	wm := window.NewManager()
	eval := threshold.NewEvaluator(0.4, 0.2)
	coord := alarmfsm.NewCoordinator(wm, 1, 1, eval, []model.Cell{cell})

	policy := threshold.NewPolicy(cell, eval)
	mid := -16.2 // ClearLimit=-16.5 < mid < UpperLimit=-16
	if policy.ActiveOverPredicate()(mid) {
		t.Fatalf("ActiveOverPredicate must use UpperLimit; %v should not be over", mid)
	}
	if !policy.ClearingOverPredicate()(mid) {
		t.Fatalf("precondition: ClearingOverPredicate should count %v as over", mid)
	}

	base := time.Date(2026, 8, 20, 11, 0, 0, 0, time.UTC)
	res := driveClosedWindow(t, coord, cell.ID, -15, base, dur)
	if res.Alarm.State != model.StatePending {
		t.Fatalf("first closed excursion want Pending, got %s", res.Alarm.State)
	}
	res = driveClosedWindow(t, coord, cell.ID, -15, base.Add(2*dur), dur)
	if res.Alarm.State != model.StateActive {
		t.Fatalf("setup want Active, got %s", res.Alarm.State)
	}

	// Expire prior hot samples; only mid remains under Active predicate.
	ts := base.Add(20 * time.Minute)
	now := ts.Add(dur)
	res, err := coord.ProcessSample(model.ProbeSample{
		CellID: cell.ID, ProbeID: "p1", TempC: mid, TS: ts,
	}, now)
	if err != nil {
		t.Fatal(err)
	}
	if res.Snapshot.Count != 1 {
		t.Fatalf("expected single sample after expiry, got count=%d", res.Snapshot.Count)
	}
	if res.Snapshot.OverCount != 0 {
		t.Fatalf("Active state must keep ActiveOverPredicate; mid temp counted over=%d ratio=%v", res.Snapshot.OverCount, res.Snapshot.OverRatio)
	}
}

func TestFSMRejectsNormalToActiveSkip(t *testing.T) {
	err := alarmfsm.ValidateTransition(model.StateNormal, model.StateActive)
	if err == nil {
		t.Fatal("expected illegal transition error for Normal->Active")
	}
	if !strings.Contains(err.Error(), "must pass Pending") {
		t.Fatalf("clear FAIL message missing; got %v", err)
	}

	dur := time.Minute
	cell := testCell("c-fsm", dur)
	wm := window.NewManager()
	eval := threshold.NewEvaluator(0.4, 0.2)
	coord := alarmfsm.NewCoordinator(wm, 2, 1, eval, []model.Cell{cell})

	base := time.Date(2026, 8, 20, 12, 0, 0, 0, time.UTC)
	res := driveClosedWindow(t, coord, cell.ID, -15, base, dur)
	if res.Alarm.State != model.StatePending {
		t.Fatalf("first closed excursion with pendingWindows=2 must be Pending, got %s (must not skip to Active)", res.Alarm.State)
	}
}

func TestPendingRequiresWindowClosed(t *testing.T) {
	fsm := alarmfsm.NewFSM("c-pend", 2, 1)
	now := time.Now()
	fsm.Step(alarmfsm.Input{Now: now, WindowClosed: true, HasExcursion: true})
	if fsm.State() != model.StatePending {
		t.Fatalf("setup Pending, got %s", fsm.State())
	}
	before := fsm.Context().ConsecutiveExcursions
	fsm.Step(alarmfsm.Input{Now: now.Add(time.Second), WindowClosed: false, HasExcursion: true})
	fsm.Step(alarmfsm.Input{Now: now.Add(2 * time.Second), WindowClosed: false, HasExcursion: true})
	if fsm.State() == model.StateActive {
		t.Fatal("Pending must not increment toward Active without WindowClosed")
	}
	if fsm.Context().ConsecutiveExcursions != before {
		t.Fatalf("ConsecutiveExcursions changed without WindowClosed: before=%d after=%d", before, fsm.Context().ConsecutiveExcursions)
	}
}

func TestCoordinatorEmptyWindowNoExcursion(t *testing.T) {
	dur := time.Minute
	cell := testCell("c-empty", dur)
	eval := threshold.NewEvaluator(0.4, 0.2)
	policy := threshold.NewPolicy(cell, eval)
	stats := window.Stats{Count: 0, OverRatio: 0}
	if policy.ShouldEnterPending(stats, true) {
		t.Fatal("count=0 must not enter pending")
	}
	if eval.HasExcursion(stats) {
		t.Fatal("count=0 must not be treated as excursion")
	}
	if eval.HasExcursion(window.Stats{Count: 0, OverRatio: 1}) {
		t.Fatal("count=0 must not be treated as excursion even if OverRatio looks high")
	}
}
