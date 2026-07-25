package cortex

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// Embedder turns text into a vector for semantic (vector) recall. The Cortex uses it
// to embed neuron content and queries so Search can rank by meaning, not just shared
// words (F41 / E58). Implementations wrap a local model (Ollama) or a cloud provider.
type Embedder interface {
	Embed(ctx context.Context, text string) ([]float64, error)
}

// OllamaEmbedder embeds text via a local Ollama server's /api/embed endpoint (default
// model nomic-embed-text, 768-dim). Dependency-free: we own the wire format.
type OllamaEmbedder struct {
	BaseURL string // e.g. http://localhost:11434 (default)
	Model   string // e.g. nomic-embed-text (default)
	Client  *http.Client
}

func (e *OllamaEmbedder) client() *http.Client {
	if e.Client != nil {
		return e.Client
	}
	return http.DefaultClient
}

func (e *OllamaEmbedder) model() string {
	if e.Model != "" {
		return e.Model
	}
	return "nomic-embed-text"
}

// Embed returns the embedding vector for text.
func (e *OllamaEmbedder) Embed(ctx context.Context, text string) ([]float64, error) {
	base := e.BaseURL
	if base == "" {
		base = "http://localhost:11434"
	}
	body, err := json.Marshal(map[string]any{"model": e.model(), "input": text})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(base, "/")+"/api/embed", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := e.client().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("embed: HTTP %d from %s", resp.StatusCode, base)
	}
	var out struct {
		Embeddings [][]float64 `json:"embeddings"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	if len(out.Embeddings) == 0 || len(out.Embeddings[0]) == 0 {
		return nil, fmt.Errorf("embed: no embedding returned by %s", e.model())
	}
	return out.Embeddings[0], nil
}
