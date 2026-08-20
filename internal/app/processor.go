package app

import (
	"context"
	"time"

	"github.com/lacsar712/frostcell/internal/model"
)

// ProcessorBatch ingests multiple samples in order for simulation/testing.
type ProcessorBatch struct {
	app *App
}

// NewProcessorBatch creates a batch processor.
func NewProcessorBatch(app *App) *ProcessorBatch {
	return &ProcessorBatch{app: app}
}

// IngestAll processes samples sequentially sharing one clock progression.
func (p *ProcessorBatch) IngestAll(ctx context.Context, samples []model.ProbeSample) ([]model.ProcessingResult, error) {
	results := make([]model.ProcessingResult, 0, len(samples))
	for _, s := range samples {
		if ctx.Err() != nil {
			return results, ctx.Err()
		}
		res, err := p.app.ProcessSample(s, s.TS)
		if err != nil {
			return results, err
		}
		results = append(results, res)
	}
	return results, nil
}

// TickWindow advances evaluation at now without new samples (for periodic scans).
func (p *ProcessorBatch) TickWindow(ctx context.Context, cellID string, now time.Time) (model.ProcessingResult, error) {
	if ctx.Err() != nil {
		return model.ProcessingResult{}, ctx.Err()
	}
	snap, err := p.app.coordinator.WindowSnapshot(cellID, now)
	if err != nil {
		return model.ProcessingResult{}, err
	}
	alarm, err := p.app.coordinator.AlarmRecord(cellID, now)
	if err != nil {
		return model.ProcessingResult{}, err
	}
	p.app.store.SaveSnapshot(snap)
	p.app.store.SaveAlarm(alarm)
	return model.ProcessingResult{
		CellID:      cellID,
		Snapshot:    snap,
		Alarm:       alarm,
		ProcessedAt: now,
	}, nil
}
