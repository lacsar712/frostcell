package app

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/lacsar712/frostcell/internal/model"
)

// Routes registers JSON API endpoints.
type Routes struct {
	app *App
}

// NewRoutes creates route handlers bound to the app.
func NewRoutes(app *App) *Routes {
	return &Routes{app: app}
}

// Register mounts API routes on mux.
func (r *Routes) Register(mux *http.ServeMux) {
	mux.HandleFunc("/v1/cells/", r.handleCellWindow)
	mux.HandleFunc("/v1/alarms", r.handleAlarms)
	mux.HandleFunc("/v1/health", r.handleHealth)
}

func (r *Routes) handleCellWindow(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(req.URL.Path, "/v1/cells/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 2 || parts[1] != "window" {
		http.NotFound(w, req)
		return
	}
	cellID := parts[0]
	now := time.Now()

	snap, err := r.app.coordinator.WindowSnapshot(cellID, now)
	if err != nil {
		if err == model.ErrCellNotFound {
			http.NotFound(w, req)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	alarm, err := r.app.coordinator.AlarmRecord(cellID, now)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]any{
		"cellID":   cellID,
		"window":   snap,
		"alarm":    alarm,
		"queriedAt": now,
	})
}

func (r *Routes) handleAlarms(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	now := time.Now()
	alarms := r.app.coordinator.ListAlarms(now)
	writeJSON(w, map[string]any{
		"alarms":    alarms,
		"queriedAt": now,
	})
}

func (r *Routes) handleHealth(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	payload := map[string]any{
		"status":  "ok",
		"cells":   len(r.app.cfg.Cells),
		"updated": r.app.store.LastUpdated(),
	}
	if r.app.notifier != nil {
		payload["notifyCircuit"] = r.app.notifier.CircuitState()
	}
	writeJSON(w, payload)
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}
