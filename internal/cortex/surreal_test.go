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

// newTestStore starts an in-memory SurrealDB sidecar and returns a SurrealStore wired
// to a deterministic 3-dim fake embedder. Skips if the surreal binary is unavailable.
func newTestStore(t *testing.T) *SurrealStore {
	t.Helper()
	if LocateSurreal() == "" {
		t.Skip("surreal binary not found; skipping SurrealDB integration test")
	}
	sc := &Sidecar{Addr: fmt.Sprintf("127.0.0.1:%d", freePort(t)), DataPath: "memory"}
	if err := sc.Start(context.Background()); err != nil {
		t.Fatalf("sidecar start: %v", err)
	}
	t.Cleanup(sc.Stop)
	store, err := NewSurrealStore(context.Background(), SurrealConfig{
		Addr: sc.HTTPAddr(), Dimension: 3, Embedder: fakeEmbedder{dim: 3},
	})
	if err != nil {
		t.Fatalf("NewSurrealStore: %v", err)
	}
	return store
}

// TestSurrealStore is an integration test against a real (sidecar) SurrealDB. It skips
// if the surreal binary is not available. It exercises put/get/hybrid-search/relate/
// remember/all over the HTTP /sql wire.
func TestSurrealStore(t *testing.T) {
	store := newTestStore(t)
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
		t.Fatalf("relate: %v", err)
	}
	// the synapse must actually exist in the graph: A's neighbours include B.
	nbs, err := store.neighbors(ctx, []string{idA})
	if err != nil {
		t.Fatalf("neighbors: %v", err)
	}
	foundB := false
	for _, nb := range nbs {
		if nb.ID == idB {
			foundB = true
		}
	}
	if !foundB {
		t.Errorf("graph edge A->B not found via neighbours: %+v", nbs)
	}
	if err := store.Remember(ctx, Engram{NeuronID: idA, Strength: 1.0}); err != nil {
		t.Errorf("remember: %v", err)
	}
}

// TestSurrealReinforcement proves recall strengthens a memory (F6): searching bumps the
// neuron's access_count and pushes its decay back toward 1.0.
func TestSurrealReinforcement(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	id, err := store.PutNeuron(ctx, Neuron{Kind: KindMemory, Content: "cats are great pets", Decay: 0.5})
	if err != nil {
		t.Fatalf("put: %v", err)
	}
	before, _ := store.GetNeuron(ctx, id)
	if before.AccessCount != 0 {
		t.Fatalf("fresh neuron access_count = %d, want 0", before.AccessCount)
	}
	if _, err := store.Search(ctx, "cats", 3); err != nil {
		t.Fatalf("search: %v", err)
	}
	after, _ := store.GetNeuron(ctx, id)
	if after.AccessCount < 1 {
		t.Errorf("access_count after recall = %d, want >= 1 (reinforcement)", after.AccessCount)
	}
	if after.Decay <= before.Decay {
		t.Errorf("decay after recall = %f, want > %f (reinforced)", after.Decay, before.Decay)
	}
}

// TestSurrealGraphExpansion proves recall traverses synapses (F6): a neuron related to a
// recalled neuron surfaces as a neighbour even when it does not match the query.
func TestSurrealGraphExpansion(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	idA, _ := store.PutNeuron(ctx, Neuron{Kind: KindMemory, Content: "cats purr softly", Decay: 1})
	idB, _ := store.PutNeuron(ctx, Neuron{Kind: KindMemory, Content: "fish swim deep", Decay: 1})
	if err := store.Relate(ctx, Synapse{From: idA, To: idB, Kind: "relates_to"}); err != nil {
		t.Fatalf("relate: %v", err)
	}
	// B ("fish") does not match the query "cats" - it must arrive via the graph edge.
	nbs, err := store.neighbors(ctx, []string{idA})
	if err != nil {
		t.Fatalf("neighbors: %v", err)
	}
	found := false
	for _, nb := range nbs {
		if nb.ID == idB {
			found = true
		}
	}
	if !found {
		t.Fatalf("neighbour B not reached from A: %+v", nbs)
	}
	hits, err := store.Search(ctx, "cats", 5)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	inHits := false
	for _, h := range hits {
		if h.ID == idB {
			inHits = true
		}
	}
	if !inHits {
		t.Errorf("graph neighbour B not surfaced by Search for 'cats': %+v", hits)
	}
}

// TestSurrealForgetting proves the forgetting curve (F6): un-reinforced neurons decay
// and are pruned, while engram-hardened neurons survive.
func TestSurrealForgetting(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	idEphemeral, _ := store.PutNeuron(ctx, Neuron{Kind: KindMemory, Content: "a passing thought", Decay: 1})
	idDurable, _ := store.PutNeuron(ctx, Neuron{Kind: KindMemory, Content: "a hardened lesson", Decay: 1})
	if err := store.Remember(ctx, Engram{NeuronID: idDurable, Strength: 1.0}); err != nil {
		t.Fatalf("remember: %v", err)
	}
	// decay halves each Forget; after 5 passes 1.0 -> ~0.03 (< 0.1 threshold).
	var pruned int
	for i := 0; i < 5; i++ {
		n, err := store.Forget(ctx)
		if err != nil {
			t.Fatalf("forget: %v", err)
		}
		pruned += n
	}
	if pruned < 1 {
		t.Errorf("expected the ephemeral neuron to be pruned, pruned=%d", pruned)
	}
	if got, _ := store.GetNeuron(ctx, idEphemeral); got.ID != "" {
		t.Errorf("ephemeral neuron should have been forgotten, got %+v", got)
	}
	if got, _ := store.GetNeuron(ctx, idDurable); got.ID == "" {
		t.Error("engram-hardened neuron should survive forgetting")
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
