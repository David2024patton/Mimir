package cortex

import (
	"context"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"
)

// fakeEmbedder is a deterministic 3-dim embedder so the SurrealDB integration test can
// exercise vector recall without depending on a live embedding model.
type fakeEmbedder struct{ dim int }

func (f fakeEmbedder) Embed(_ context.Context, text string) ([]float64, error) {
	v := make([]float64, f.dim)
	if strings.Contains(text, "cat") {
		v[0] = 1.0
	}
	if strings.Contains(text, "dog") {
		v[1] = 1.0
	}
	if strings.Contains(text, "fish") {
		v[2] = 1.0
	}
	return v, nil
}

func freePort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("free port: %v", err)
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}

// TestSurrealStore is an integration test against a real (sidecar) SurrealDB. It skips
// if the surreal binary is not available. It exercises put/get/hybrid-search/relate/
// remember/all over the HTTP /sql wire.
func TestSurrealStore(t *testing.T) {
	if LocateSurreal() == "" {
		t.Skip("surreal binary not found; skipping SurrealDB integration test")
	}
	sc := &Sidecar{Addr: fmt.Sprintf("127.0.0.1:%d", freePort(t)), DataPath: "memory"}
	if err := sc.Start(context.Background()); err != nil {
		t.Fatalf("sidecar start: %v", err)
	}
	defer sc.Stop()

	store, err := NewSurrealStore(context.Background(), SurrealConfig{
		Addr: sc.HTTPAddr(), Dimension: 3, Embedder: fakeEmbedder{dim: 3},
	})
	if err != nil {
		t.Fatalf("NewSurrealStore: %v", err)
	}
	ctx := context.Background()

	idA, err := store.PutNeuron(ctx, Neuron{Kind: KindMemory, Content: "cats are great pets", Decay: 1})
	if err != nil {
		t.Fatalf("put cats: %v", err)
	}
	idB, err := store.PutNeuron(ctx, Neuron{Kind: KindMemory, Content: "dogs are loyal pets", Decay: 1})
	if err != nil {
		t.Fatalf("put dogs: %v", err)
	}
	if _, err := store.PutNeuron(ctx, Neuron{Kind: KindMemory, Content: "fish swim in water", Decay: 1}); err != nil {
		t.Fatalf("put fish: %v", err)
	}

	got, err := store.GetNeuron(ctx, idA)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Content != "cats are great pets" {
		t.Errorf("get content = %q", got.Content)
	}

	// Hybrid recall: "cats" should surface the cat neuron first (vector [1,0,0] + BM25).
	hits, err := store.Search(ctx, "cats", 3)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(hits) == 0 {
		t.Fatal("search returned no hits")
	}
	if !strings.Contains(hits[0].Content, "cat") {
		t.Errorf("top hit = %q, expected the cat neuron", hits[0].Content)
	}

	if all := store.All(); len(all) < 3 {
		t.Errorf("All = %d neurons, want >= 3", len(all))
	}

	if err := store.Relate(ctx, Synapse{From: idA, To: idB, Kind: "relates_to"}); err != nil {
		t.Errorf("relate: %v", err)
	}
	if err := store.Remember(ctx, Engram{NeuronID: idA, Strength: 1.0}); err != nil {
		t.Errorf("remember: %v", err)
	}
}

// TestOllamaEmbedderLive verifies the real Ollama embedder (nomic-embed-text). It skips
// if Ollama is not reachable.
func TestOllamaEmbedderLive(t *testing.T) {
	e := &OllamaEmbedder{}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	vec, err := e.Embed(ctx, "hello world")
	if err != nil {
		t.Skipf("Ollama embed not available: %v", err)
	}
	if len(vec) == 0 {
		t.Fatal("empty embedding")
	}
	t.Logf("nomic-embed-text embedding dim = %d", len(vec))
}
