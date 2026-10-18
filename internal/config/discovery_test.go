package config

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestProbeOllamaFiltersEmbeddings(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/tags" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"models":[{"name":"nomic-embed-text:latest"},{"name":"qwen3:8b"},{"name":"deepseek-coder:6.7b"}]}`))
	}))
	defer srv.Close()

	models := probeOllama(context.Background(), srv.URL, time.Second)
	if len(models) != 2 {
		t.Fatalf("got %d models, want 2: %+v", len(models), models)
	}
	if models[0].Model != "qwen3:8b" || !models[0].Local || models[0].ProviderID != "ollama" {
		t.Errorf("unexpected first model: %+v", models[0])
	}
}

func TestMetaModelAPIDiscovery(t *testing.T) {
	t.Setenv("MIMIR_BASE_URL", "")
	t.Setenv("META_API_KEY", "meta-test-key")
	t.Setenv("MIMIR_MODEL", "")
	t.Setenv("META_MODEL", "")

	ep, ok := AutoDiscoverProvider(context.Background())
	if !ok {
		t.Fatal("expected AutoDiscoverProvider to succeed with META_API_KEY")
	}
	if ep.ID != "meta" || ep.BaseURL != "https://api.meta.ai/v1" || ep.Model != "muse-wonder-1.0" {
		t.Errorf("unexpected Meta provider: %+v", ep)
	}

	cat := MetaMuseCatalog()
	if len(cat) < 7 {
		t.Errorf("expected at least 7 Meta Muse models in catalog, got %d", len(cat))
	}
}
