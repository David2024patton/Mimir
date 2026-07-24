package cortex

import (
	"context"
	"fmt"
	"sync/atomic"
)

// NeuronKind classifies a knowledge neuron.
type NeuronKind string

const (
	KindSource  NeuronKind = "source"
	KindNote    NeuronKind = "note"
	KindConcept NeuronKind = "concept"
	KindMemory  NeuronKind = "memory"
	KindSkill   NeuronKind = "skill"
)

// Neuron is one knowledge node in the Cortex (the Well / Mimisbrunnr).
type Neuron struct {
	ID        string
	Kind      NeuronKind
	Layer     string // activity | context | experience | identity | preference
	Title     string
	Content   string
	Embedding []float64
	Decay     float64
}

// Store is the Cortex persistence interface (SurrealDB-backed; E6.2).
type Store interface {
	PutNeuron(ctx context.Context, n Neuron) (string, error)
	GetNeuron(ctx context.Context, id string) (Neuron, error)
	Search(ctx context.Context, query string, limit int) ([]Neuron, error)
	Relate(ctx context.Context, from, to, kind string) error
}

var idCounter atomic.Int64

func newID() string {
	return fmt.Sprintf("n-%d", idCounter.Add(1))
}

// MemoryStore is an in-memory placeholder Store (replaced by SurrealDB in E1.4/E6.2).
type MemoryStore struct {
	neurons map[string]Neuron
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{neurons: map[string]Neuron{}}
}

func (m *MemoryStore) PutNeuron(ctx context.Context, n Neuron) (string, error) {
	if n.ID == "" {
		n.ID = newID()
	}
	m.neurons[n.ID] = n
	return n.ID, nil
}

func (m *MemoryStore) GetNeuron(ctx context.Context, id string) (Neuron, error) {
	return m.neurons[id], nil
}

func (m *MemoryStore) Search(ctx context.Context, query string, limit int) ([]Neuron, error) {
	// TODO(E6.2): full-text + vector + graph search via SurrealDB.
	out := make([]Neuron, 0, limit)
	for _, n := range m.neurons {
		out = append(out, n)
		if len(out) >= limit {
			break
		}
	}
	return out, nil
}

func (m *MemoryStore) Relate(ctx context.Context, from, to, kind string) error {
	// TODO(E6.2): RELATE from->synapse->to in SurrealDB.
	return nil
}
