package app_test

import (
	"testing"
	"time"

	"github.com/lacsar712/frostcell/internal/app"
	"github.com/lacsar712/frostcell/internal/config"
	"github.com/lacsar712/frostcell/internal/model"
)

func TestAppProcessSample(t *testing.T) {
	cfg := config.Default()
	cfg.WindowDuration = time.Minute
	cfg.Cells[0].WindowDur = time.Minute
	application, err := app.New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 8, 20, 10, 0, 0, 0, time.UTC)
	sample := model.ProbeSample{
		CellID:  "cell-01",
		ProbeID: "p1",
		TempC:   -17,
		TS:      now,
	}
	res, err := application.ProcessSample(sample, now)
	if err != nil {
		t.Fatal(err)
	}
	if res.CellID != "cell-01" {
		t.Fatalf("result %+v", res)
	}
	snap, ok := application.Store().GetSnapshot("cell-01")
	if !ok || snap.Count != 1 {
		t.Fatalf("store snap %+v ok=%v", snap, ok)
	}
}

func TestAppHandler(t *testing.T) {
	cfg := config.Default()
	application, err := app.New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	h := application.Handler()
	if h == nil {
		t.Fatal("nil handler")
	}
}
