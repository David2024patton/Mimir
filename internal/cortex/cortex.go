// Package cortex is Mímir's brain (F6): a neural knowledge graph of neurons
// (knowledge nodes), synapses (relationships), and engrams (durable memories).
// This is the namesake feature - the memory that makes Mímir Mímir.
package cortex

import "context"

// NeuronKind classifies a knowledge neuron.
type NeuronKind string

const (
	KindSource  NeuronKind = "source"  // ingested content (docs, web, video)
	KindNote    NeuronKind = "note"    // a note the agent or user wrote
	KindConcept NeuronKind = "concept" // an abstracted concept
	KindMemory  NeuronKind = "memory"  // a learned preference/lesson
	KindSkill   NeuronKind = "skill"   // a reusable skill
)

// Neuron is one node in the knowledge graph.
type Neuron struct {
	ID        string
	Kind      NeuronKind
	Layer     string    // activity | context | experience | identity | preference
	Title     string
	Content   string
	Embedding []float64 // vector for RAG
	Decay     float64   // forgetting-curve weight (F6)
}

// Synapse is a relationship between two neurons.
type Synapse struct {
	From string
	To   string
	Kind string // references | derives_from | relates_to | contradicts
}

// Engram is a durable memory - a neuron hardened so it won't be forgotten.
type Engram struct {
	NeuronID string
	Strength float64
}

// Store is the persistence interface for the Cortex. The real implementation
// uses SurrealDB (graph + vector + document); tests can use an in-memory store.
type Store interface {
	PutNeuron(ctx context.Context, n Neuron) (string, error)
	GetNeuron(ctx context.Context, id string) (Neuron, error)
	Search(ctx context.Context, query string, limit int) ([]Neuron, error)
	Relate(ctx context.Context, s Synapse) error
	Remember(ctx context.Context, e Engram) error
}
