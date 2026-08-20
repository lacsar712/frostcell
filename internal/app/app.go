package app

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/lacsar712/frostcell/internal/alarmfsm"
	"github.com/lacsar712/frostcell/internal/config"
	"github.com/lacsar712/frostcell/internal/ingest"
	"github.com/lacsar712/frostcell/internal/model"
	"github.com/lacsar712/frostcell/internal/notify"
	"github.com/lacsar712/frostcell/internal/store"
	"github.com/lacsar712/frostcell/internal/threshold"
	"github.com/lacsar712/frostcell/internal/web"
	"github.com/lacsar712/frostcell/internal/window"
)

// App wires all frostcell subsystems.
type App struct {
	cfg          config.Config
	store        *store.MemoryStore
	windows      *window.Manager
	coordinator  *alarmfsm.Coordinator
	notifier     *notify.Service
	alarmIndex   *store.AlarmIndex
	snapReader   *store.SnapshotReader
	httpServer   *http.Server
}

// New constructs an App from configuration.
func New(cfg config.Config) (*App, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	mem := store.NewMemoryStore()
	wm := window.NewManager()
	eval := threshold.NewEvaluator(cfg.ExcursionRatio, cfg.ClearRatio)
	coord := alarmfsm.NewCoordinator(wm, cfg.PendingWindows, cfg.ClearingWindows, eval, cfg.Cells)

	for _, cell := range cfg.Cells {
		mem.SaveAlarm(model.AlarmRecord{
			CellID:    cell.ID,
			State:     model.StateNormal,
			UpdatedAt: time.Now(),
		})
	}

	var notifier *notify.Service
	if cfg.NotifyURL != "" {
		sender := notify.NewHTTPSender(cfg.NotifyURL, cfg.NotifyTimeout)
		notifier = notify.NewService(cfg.NotifyURL, sender, cfg.CircuitThreshold, cfg.CircuitCooldown)
	}

	a := &App{
		cfg:         cfg,
		store:       mem,
		windows:     wm,
		coordinator: coord,
		notifier:    notifier,
		alarmIndex:  store.NewAlarmIndex(mem),
		snapReader:  store.NewSnapshotReader(mem, coord.WindowSnapshot),
	}
	return a, nil
}

// ProcessSample implements ingest.Processor.
func (a *App) ProcessSample(sample model.ProbeSample, now time.Time) (model.ProcessingResult, error) {
	result, err := a.coordinator.ProcessSample(sample, now)
	if err != nil {
		return result, err
	}
	a.store.SaveSnapshot(result.Snapshot)
	a.store.SaveAlarm(result.Alarm)

	if len(result.Events) > 0 && a.notifier != nil {
		ctx, cancel := context.WithTimeout(context.Background(), a.cfg.NotifyTimeout)
		defer cancel()
		_ = a.notifier.Notify(ctx, result.Events)
	}
	return result, nil
}

// Handler returns the root HTTP handler.
func (a *App) Handler() http.Handler {
	mux := http.NewServeMux()
	routes := NewRoutes(a)
	routes.Register(mux)

	ingestHandler := ingest.NewHandler(a.cfg.HMACSecret, a)
	mux.Handle("/v1/probes/sample", ingestHandler)

	webHandler := web.NewHandler(a.store, a.coordinator, a.cfg.Cells)
	mux.Handle("/", webHandler)

	return mux
}

// Run starts the HTTP server until context cancellation.
func (a *App) Run(ctx context.Context) error {
	handler := a.Handler()
	a.httpServer = &http.Server{
		Addr:              a.cfg.ListenAddr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("frostcell listening on %s", a.cfg.ListenAddr)
		errCh <- a.httpServer.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return a.httpServer.Shutdown(shutdownCtx)
	case err := <-errCh:
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	}
}

// Store returns the memory store for tests and diagnostics.
func (a *App) Store() *store.MemoryStore {
	return a.store
}

// Coordinator exposes the alarm coordinator.
func (a *App) Coordinator() *alarmfsm.Coordinator {
	return a.coordinator
}

// Config returns loaded configuration.
func (a *App) Config() config.Config {
	return a.cfg
}

// Notifier returns the notification service if configured.
func (a *App) Notifier() *notify.Service {
	return a.notifier
}
