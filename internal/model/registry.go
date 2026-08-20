package model

import "sync"

// CellRegistry tracks known cells for validation at ingest time.
type CellRegistry struct {
	mu    sync.RWMutex
	cells map[string]Cell
}

// NewCellRegistry creates an empty registry.
func NewCellRegistry() *CellRegistry {
	return &CellRegistry{cells: make(map[string]Cell)}
}

// Register adds or replaces a cell definition.
func (r *CellRegistry) Register(cell Cell) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cells[cell.ID] = cell
}

// Get returns a cell by id.
func (r *CellRegistry) Get(id string) (Cell, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.cells[id]
	return c, ok
}

// IDs returns all registered cell ids.
func (r *CellRegistry) IDs() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]string, 0, len(r.cells))
	for id := range r.cells {
		out = append(out, id)
	}
	return out
}

// Contains reports whether cell id is known.
func (r *CellRegistry) Contains(id string) bool {
	_, ok := r.Get(id)
	return ok
}
