package cortex

import (
	"context"
	"path/filepath"
	"testing"
)

func TestMemoryPersistsToDisk(t *testing.T) {
	p := filepath.Join(t.TempDir(), "cortex.json")
	s1, err := NewMemoryStoreAt(p)
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	if _, err := s1.PutNeuron(context.Background(), Neuron{
		Kind: KindMemory, Content: "the codeword is MANGO77", Decay: 1,
	}); err != nil {
		t.Fatalf("put: %v", err)
	}

	// A brand-new store opened from the same file (a restart) must see the neuron.
	s2, err := NewMemoryStoreAt(p)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	got, _ := s2.Search(context.Background(), "MANGO77", 5)
	if len(got) != 1 || got[0].Content != "the codeword is MANGO77" {
		t.Fatalf("persisted search = %+v", got)
	}
}

func TestMemoryTokenSearch(t *testing.T) {
	s := NewMemoryStore()
	_, _ = s.PutNeuron(context.Background(), Neuron{Content: "alpha beta gamma delta", Decay: 1})
	if got, _ := s.Search(context.Background(), "gamma zzzz", 5); len(got) != 1 {
		t.Errorf("expected match on shared token gamma, got %d", len(got))
	}
	if got, _ := s.Search(context.Background(), "zzzz qq", 5); len(got) != 0 {
		t.Errorf("expected no match, got %d", len(got))
	}
}
