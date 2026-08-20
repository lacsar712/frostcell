package window_test

import (
	"testing"
	"time"

	"github.com/lacsar712/frostcell/internal/model"
	"github.com/lacsar712/frostcell/internal/store"
	"github.com/lacsar712/frostcell/internal/window"
)

func TestSnapshotTempsNotAliased(t *testing.T) {
	cell := model.Cell{ID: "c-alias", WindowDur: 5 * time.Minute, SetpointC: -18, DeltaC: 2}
	w := window.NewCellWindow(cell.ID, cell.WindowDur, func(temp float64) bool { return temp > cell.UpperLimit() })
	base := time.Date(2026, 8, 20, 14, 0, 0, 0, time.UTC)
	w.AddSample(-17, base)
	w.AddSample(-16.5, base.Add(time.Minute))

	snap := w.ToModelSnapshot(base.Add(2 * time.Minute))
	if len(snap.Temps) < 2 {
		t.Fatalf("expected temps in snapshot, got %+v", snap)
	}
	snap2 := w.ToModelSnapshot(base.Add(2 * time.Minute))
	if len(snap2.Temps) < 2 {
		t.Fatal("expected second snapshot temps")
	}
	if &snap.Temps[0] == &snap2.Temps[0] {
		t.Fatal("WindowSnapshot.Temps aliases internal window storage across snapshots")
	}
	snap.Temps[0] = 999
	if snap2.Temps[0] == 999 {
		t.Fatal("mutating one snapshot Temps affected another (shared backing array)")
	}

	direct := w.SampleTemps()
	direct2 := w.SampleTemps()
	if len(direct) < 2 || len(direct2) < 2 {
		t.Fatal("SampleTemps returned empty")
	}
	if &direct[0] == &direct2[0] {
		t.Fatal("SampleTemps aliases internal buffer")
	}

	mem := store.NewMemoryStore()
	mem.SaveSnapshot(snap)
	got, ok := mem.GetSnapshot(cell.ID)
	if !ok {
		t.Fatal("missing snapshot")
	}
	got.Temps[0] = 888
	got2, _ := mem.GetSnapshot(cell.ID)
	if got2.Temps[0] == 888 {
		t.Fatal("store snapshot Temps aliases stored slice")
	}
}
