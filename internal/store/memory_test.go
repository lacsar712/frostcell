package store_test

import (
	"testing"
	"time"

	"github.com/lacsar712/frostcell/internal/model"
	"github.com/lacsar712/frostcell/internal/store"
)

func TestMemoryStore(t *testing.T) {
	s := store.NewMemoryStore()
	snap := model.WindowSnapshot{CellID: "c1", Count: 5, MeanTempC: -17}
	s.SaveSnapshot(snap)
	got, ok := s.GetSnapshot("c1")
	if !ok || got.Count != 5 {
		t.Fatalf("snapshot %+v", got)
	}
	rec := model.AlarmRecord{CellID: "c1", State: model.StateActive, UpdatedAt: time.Now()}
	s.SaveAlarm(rec)
	a, ok := s.GetAlarm("c1")
	if !ok || a.State != model.StateActive {
		t.Fatalf("alarm %+v", a)
	}
}

func TestAlarmIndex(t *testing.T) {
	s := store.NewMemoryStore()
	s.SaveAlarm(model.AlarmRecord{CellID: "c1", State: model.StateActive})
	s.SaveAlarm(model.AlarmRecord{CellID: "c2", State: model.StateNormal})
	idx := store.NewAlarmIndex(s)
	active := idx.ActiveAlarms()
	if len(active) != 1 {
		t.Fatalf("active %d", len(active))
	}
}
