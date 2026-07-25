package cortex

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

// MemoryStore is an in-memory implementation of Store, used by the walking
// skeleton and tests. The production implementation uses SurrealDB.
type MemoryStore struct {
	mu      sync.RWMutex
	neurons map[string]Neuron
	nextID  int
}

// NewMemoryStore returns an empty in-memory Cortex store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{neurons: map[string]Neuron{}, nextID: 1}
}

// PutNeuron stores a neuron, assigning an id if it has none.
func (m *MemoryStore) PutNeuron(ctx context.Context, n Neuron) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if n.ID == "" {
		n.ID = fmt.Sprintf("n-%d", m.nextID)
		m.nextID++
	}
	m.neurons[n.ID] = n
	return n.ID, nil
}

// GetNeuron returns a neuron by id.
func (m *MemoryStore) GetNeuron(ctx context.Context, id string) (Neuron, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.neurons[id], nil
}

// Search returns neurons matching the query. Walking-skeleton search is a naive
// substring match; the real implementation uses vector similarity (nomic
// embeddings) plus graph traversal.
func (m *MemoryStore) Search(ctx context.Context, query string, limit int) ([]Neuron, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var out []Neuron
	q := strings.ToLower(query)
	for _, n := range m.neurons {
		if strings.Contains(strings.ToLower(n.Content), q) {
			out = append(out, n)
			if len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

// Relate records a synapse (no-op in the in-memory skeleton).
func (m *MemoryStore) Relate(ctx context.Context, s Synapse) error { return nil }

// Remember hardens an engram (no-op in the in-memory skeleton).
func (m *MemoryStore) Remember(ctx context.Context, e Engram) error { return nil }
