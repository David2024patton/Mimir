package cortex

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"unicode"
)

// MemoryStore is a Cortex store. With an empty path it is purely in-memory (tests);
// with a path it persists neurons to a JSON file so memory survives across runs - the
// dependency-free base that SurrealDB (E6.2) later upgrades to vector + graph recall.
type MemoryStore struct {
	mu      sync.Mutex
	path    string
	neurons map[string]Neuron
	nextID  int
}

// NewMemoryStore returns an in-memory store (no persistence).
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{neurons: map[string]Neuron{}, nextID: 1}
}

// NewMemoryStoreAt returns a store backed by path, loading any existing neurons.
// A missing file is not an error (the store starts empty and is created on first write).
func NewMemoryStoreAt(path string) (*MemoryStore, error) {
	s := &MemoryStore{path: path, neurons: map[string]Neuron{}, nextID: 1}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

func (m *MemoryStore) load() error {
	if m.path == "" {
		return nil
	}
	data, err := os.ReadFile(m.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		return nil
	}
	var ns []Neuron
	if err := json.Unmarshal(data, &ns); err != nil {
		return fmt.Errorf("cortex: corrupt memory file %s: %w", m.path, err)
	}
	for _, n := range ns {
		m.neurons[n.ID] = n
		if id := parseID(n.ID); id >= m.nextID {
			m.nextID = id + 1
		}
	}
	return nil
}

func (m *MemoryStore) save() error {
	if m.path == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(m.path), 0o755); err != nil {
		return err
	}
	ns := make([]Neuron, 0, len(m.neurons))
	for _, n := range m.neurons {
		ns = append(ns, n)
	}
	sort.Slice(ns, func(i, j int) bool { return ns[i].ID < ns[j].ID })
	data, err := json.MarshalIndent(ns, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(m.path, data, 0o644)
}

func parseID(id string) int {
	i := strings.LastIndex(id, "-")
	if i < 0 || i == len(id)-1 {
		return 0
	}
	var n int
	fmt.Sscanf(id[i+1:], "%d", &n)
	return n
}

// PutNeuron stores a neuron, assigning an id if it has none, and persists.
func (m *MemoryStore) PutNeuron(ctx context.Context, n Neuron) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if n.ID == "" {
		n.ID = fmt.Sprintf("n-%d", m.nextID)
		m.nextID++
	}
	m.neurons[n.ID] = n
	return n.ID, m.save()
}

// GetNeuron returns a neuron by id.
func (m *MemoryStore) GetNeuron(ctx context.Context, id string) (Neuron, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.neurons[id], nil
}

// Search returns neurons whose content shares a word (>=4 chars) with the query,
// ranked by the number of matching words. This is a basic lexical recall; vector +
// graph recall replace it once SurrealDB lands.
func (m *MemoryStore) Search(ctx context.Context, query string, limit int) ([]Neuron, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	qt := tokens(query)
	type scored struct {
		n     Neuron
		score int
	}
	var hits []scored
	for _, n := range m.neurons {
		hay := strings.ToLower(n.Content + " " + n.Title)
		sc := 0
		for _, t := range qt {
			if strings.Contains(hay, t) {
				sc++
			}
		}
		if sc > 0 {
			hits = append(hits, scored{n, sc})
		}
	}
	sort.Slice(hits, func(i, j int) bool {
		if hits[i].score != hits[j].score {
			return hits[i].score > hits[j].score
		}
		return hits[i].n.ID < hits[j].n.ID
	})
	out := make([]Neuron, 0, limit)
	for i, h := range hits {
		if i >= limit {
			break
		}
		out = append(out, h.n)
	}
	return out, nil
}

// Relate records a synapse (no-op until the graph store lands).
func (m *MemoryStore) Relate(ctx context.Context, s Synapse) error { return nil }

// Remember hardens an engram (no-op until the graph store lands).
func (m *MemoryStore) Remember(ctx context.Context, e Engram) error { return nil }

// All returns every neuron (for the /memory endpoint + tests).
func (m *MemoryStore) All() []Neuron {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]Neuron, 0, len(m.neurons))
	for _, n := range m.neurons {
		out = append(out, n)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// tokens extracts lowercase alphanumeric words of length >= 4 for lexical recall.
func tokens(s string) []string {
	var out []string
	var cur strings.Builder
	flush := func() {
		if cur.Len() >= 4 {
			out = append(out, cur.String())
		}
		cur.Reset()
	}
	for _, r := range strings.ToLower(s) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			cur.WriteRune(r)
		} else {
			flush()
		}
	}
	flush()
	return out
}
