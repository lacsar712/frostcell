package web

import (
	"encoding/json"
	"html/template"
	"net/http"
	"time"

	"github.com/lacsar712/frostcell/internal/model"
	"github.com/lacsar712/frostcell/internal/store"
)

// StatusProvider supplies live alarm and window data for the dashboard.
type StatusProvider interface {
	ListAlarms(now time.Time) []model.AlarmRecord
	WindowSnapshot(cellID string, now time.Time) (model.WindowSnapshot, error)
}

// Handler serves the embedded ops dashboard at GET /.
type Handler struct {
	store    *store.MemoryStore
	provider StatusProvider
	cells    []model.Cell
	tmpl     *template.Template
}

// NewHandler creates the web dashboard handler.
func NewHandler(st *store.MemoryStore, provider StatusProvider, cells []model.Cell) *Handler {
	tmpl := template.Must(template.New("index").Funcs(template.FuncMap{
		"fmtRatio": func(r float64) string {
			return formatPercent(r)
		},
		"fmtTemp": func(t float64) string {
			return formatTemp(t)
		},
	}).Parse(indexHTML))
	return &Handler{
		store:    st,
		provider: provider,
		cells:    cells,
		tmpl:     tmpl,
	}
}

// ServeHTTP renders the dashboard at GET /.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	now := time.Now()
	type row struct {
		Cell       model.Cell
		Snapshot   model.WindowSnapshot
		Alarm      model.AlarmRecord
		UpperLimit float64
		ClearLimit float64
	}

	rows := make([]row, 0, len(h.cells))
	for _, cell := range h.cells {
		snap, _ := h.provider.WindowSnapshot(cell.ID, now)
		alarm, _ := h.store.GetAlarm(cell.ID)
		if alarm.CellID == "" {
			alarm = model.AlarmRecord{CellID: cell.ID, State: model.StateNormal, UpdatedAt: now}
		}
		rows = append(rows, row{
			Cell:       cell,
			Snapshot:   snap,
			Alarm:      alarm,
			UpperLimit: cell.UpperLimit(),
			ClearLimit: cell.ClearLimit(),
		})
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = h.tmpl.Execute(w, map[string]any{
		"Title":     "frostcell 运维状态",
		"Rows":      rows,
		"UpdatedAt": now,
	})
}

// JSONStatus writes dashboard data as JSON for polling clients.
func (h *Handler) JSONStatus(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	payload := map[string]any{
		"alarms":    h.provider.ListAlarms(now),
		"snapshots": h.store.ListSnapshots(),
		"updatedAt": now,
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(payload)
}

func formatPercent(r float64) string {
	return itoa(int(r*100+0.5)) + "%"
}

func formatTemp(t float64) string {
	sign := ""
	if t < 0 {
		sign = "-"
		t = -t
	}
	whole := int(t)
	frac := int((t-float64(whole))*10 + 0.5)
	return sign + itoa(whole) + "." + itoa(frac) + "°C"
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	b := []byte{}
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}

const indexHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>{{.Title}}</title>
  <style>
    :root { --bg:#0f172a; --card:#1e293b; --text:#e2e8f0; --accent:#38bdf8; --warn:#fbbf24; --danger:#f87171; --ok:#4ade80; }
    * { box-sizing:border-box; }
    body { font-family: system-ui, sans-serif; background: var(--bg); color: var(--text); margin:0; padding:1.5rem; }
    h1 { margin:0 0 .5rem; font-size:1.5rem; }
    .meta { color:#94a3b8; font-size:.875rem; margin-bottom:1.5rem; }
    table { width:100%; border-collapse:collapse; background:var(--card); border-radius:8px; overflow:hidden; }
    th, td { padding:.75rem 1rem; text-align:left; border-bottom:1px solid #334155; }
    th { background:#0b1220; font-weight:600; font-size:.75rem; text-transform:uppercase; letter-spacing:.05em; color:#94a3b8; }
    .state { display:inline-block; padding:.2rem .6rem; border-radius:999px; font-size:.75rem; font-weight:600; }
    .state-Normal { background:#14532d; color:var(--ok); }
    .state-Pending { background:#713f12; color:var(--warn); }
    .state-Active { background:#7f1d1d; color:var(--danger); }
    .state-Clearing { background:#1e3a5f; color:var(--accent); }
    footer { margin-top:1.5rem; font-size:.75rem; color:#64748b; }
  </style>
</head>
<body>
  <h1>{{.Title}}</h1>
  <p class="meta">冷链温区滑动窗口监测 · 更新于 {{.UpdatedAt.Format "2006-01-02 15:04:05"}}</p>
  <table>
    <thead>
      <tr>
        <th>温区</th>
        <th>设定点</th>
        <th>窗口样本</th>
        <th>均值 / 峰值</th>
        <th>超限占比</th>
        <th>告警状态</th>
      </tr>
    </thead>
    <tbody>
      {{range .Rows}}
      <tr>
        <td><strong>{{.Cell.Name}}</strong><br><small>{{.Cell.ID}}</small></td>
        <td>{{fmtTemp .Cell.SetpointC}} ± {{fmtTemp .Cell.DeltaC}}</td>
        <td>{{.Snapshot.Count}}</td>
        <td>{{fmtTemp .Snapshot.MeanTempC}} / {{fmtTemp .Snapshot.MaxTempC}}</td>
        <td>{{fmtRatio .Snapshot.OverRatio}} <small>(限 {{fmtTemp .UpperLimit}})</small></td>
        <td><span class="state state-{{.Alarm.State}}">{{.Alarm.State}}</span></td>
      </tr>
      {{end}}
    </tbody>
  </table>
  <footer>frostcell · POST /v1/probes/sample · GET /v1/alarms</footer>
</body>
</html>`
