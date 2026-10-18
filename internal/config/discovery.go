package config

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"
)

// DiscoveredModel represents a model found via local port probing or cloud env vars.
type DiscoveredModel struct {
	ProviderID string `json:"provider_id"`
	BaseURL    string `json:"base_url"`
	APIKey     string `json:"api_key,omitempty"`
	Model      string `json:"model"`
	Local      bool   `json:"local"`
}

// AutoDiscoverProvider resolves a provider in priority order (inspired by smolcoder):
// 1. Explicit MIMIR_BASE_URL environment variable.
// 2. Cloud API keys (MIMO_API_KEY, OPENROUTER_API_KEY, OPENAI_API_KEY).
// 3. Local Ollama (:11434) and LM Studio (:1234) port probing.
func AutoDiscoverProvider(ctx context.Context) (EnvProvider, bool) {
	if ep, ok := ProviderFromEnv(); ok {
		return ep, true
	}

	// Check Mímir Well Subscription Gateway (OpenCode Zen-style unified subscription)
	if key := os.Getenv("MIMIR_SUBSCRIPTION_KEY"); key != "" {
		return EnvProvider{
			ID:      "mimir-well",
			BaseURL: envOr("MIMIR_WELL_URL", "https://api.mimirmind.com/v1"),
			APIKey:  key,
			Model:   envOr("MIMIR_MODEL", "muse-wonder-1.0"),
		}, true
	}

	// Check Meta Model API (Muse Wonder / Muse Spark / Muse Glimmer)
	if key := firstEnv("META_API_KEY", "MUSE_API_KEY"); key != "" {
		return EnvProvider{
			ID:      "meta",
			BaseURL: envOr("META_BASE_URL", "https://api.meta.ai/v1"),
			APIKey:  key,
			Model:   firstEnvDefault("muse-wonder-1.0", "MIMIR_MODEL", "META_MODEL"),
		}, true
	}

	// Check Xiaomi MiMo / OpenRouter / OpenAI / DeepSeek / Gemini / xAI / Groq keys
	if key := os.Getenv("MIMO_API_KEY"); key != "" {
		return EnvProvider{
			ID:      "mimo",
			BaseURL: envOr("MIMO_BASE_URL", "https://api.xiaomimimo.com/v1"),
			APIKey:  key,
			Model:   envOr("MIMIR_MODEL", "mimo-v2-pro"),
		}, true
	}
	if key := os.Getenv("OPENROUTER_API_KEY"); key != "" {
		return EnvProvider{
			ID:      "openrouter",
			BaseURL: "https://openrouter.ai/api/v1",
			APIKey:  key,
			Model:   envOr("MIMIR_MODEL", "xiaomi/mimo-v2-pro"),
		}, true
	}
	if key := os.Getenv("OPENAI_API_KEY"); key != "" {
		return EnvProvider{
			ID:      "openai",
			BaseURL: "https://api.openai.com/v1",
			APIKey:  key,
			Model:   envOr("MIMIR_MODEL", "gpt-4o-mini"),
		}, true
	}
	if key := os.Getenv("DEEPSEEK_API_KEY"); key != "" {
		return EnvProvider{
			ID:      "deepseek",
			BaseURL: "https://api.deepseek.com/v1",
			APIKey:  key,
			Model:   envOr("MIMIR_MODEL", "deepseek-chat"),
		}, true
	}
	if key := os.Getenv("GEMINI_API_KEY"); key != "" {
		return EnvProvider{
			ID:      "gemini",
			BaseURL: "https://generativelanguage.googleapis.com/v1beta/openai",
			APIKey:  key,
			Model:   envOr("MIMIR_MODEL", "gemini-2.5-flash"),
		}, true
	}
	if key := os.Getenv("XAI_API_KEY"); key != "" {
		return EnvProvider{
			ID:      "xai",
			BaseURL: "https://api.x.ai/v1",
			APIKey:  key,
			Model:   envOr("MIMIR_MODEL", "grok-4-fast"),
		}, true
	}
	if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" {
		return EnvProvider{
			ID:      "anthropic",
			BaseURL: envOr("ANTHROPIC_BASE_URL", "https://api.anthropic.com/v1"),
			APIKey:  key,
			Model:   envOr("MIMIR_MODEL", "claude-sonnet-4-6"),
		}, true
	}
	if key := os.Getenv("GROQ_API_KEY"); key != "" {
		return EnvProvider{
			ID:      "groq",
			BaseURL: "https://api.groq.com/openai/v1",
			APIKey:  key,
			Model:   envOr("MIMIR_MODEL", "llama-3.3-70b-versatile"),
		}, true
	}

	// Probe local Ollama and LM Studio servers
	models := DiscoverLocalModels(ctx, 350*time.Millisecond)
	if len(models) > 0 {
		m := models[0]
		if pref := os.Getenv("MIMIR_MODEL"); pref != "" {
			for _, cand := range models {
				if strings.EqualFold(cand.Model, pref) {
					m = cand
					break
				}
			}
		}
		return EnvProvider{
			ID:      m.ProviderID,
			BaseURL: m.BaseURL,
			APIKey:  m.APIKey,
			Model:   m.Model,
		}, true
	}

	return EnvProvider{}, false
}

// DiscoverLocalModels probes local Ollama (127.0.0.1:11434) and LM Studio (127.0.0.1:1234)
// and returns all available chat/coding models (filtering out pure embedding models).
func DiscoverLocalModels(ctx context.Context, timeout time.Duration) []DiscoveredModel {
	if timeout <= 0 {
		timeout = 350 * time.Millisecond
	}
	ollamaURL := envOr("MIMIR_OLLAMA_URL", "http://127.0.0.1:11434")
	lmStudioURL := envOr("MIMIR_LMSTUDIO_URL", "http://127.0.0.1:1234")

	var out []DiscoveredModel
	out = append(out, probeOllama(ctx, ollamaURL, timeout)...)
	out = append(out, probeOpenAIModels(ctx, "lmstudio", lmStudioURL, timeout)...)
	return out
}

func probeOllama(ctx context.Context, base string, timeout time.Duration) []DiscoveredModel {
	base = strings.TrimRight(base, "/")
	cctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(cctx, http.MethodGet, base+"/api/tags", nil)
	if err != nil {
		return nil
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil
	}

	var payload struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil
	}

	var out []DiscoveredModel
	for _, m := range payload.Models {
		if isEmbeddingModel(m.Name) {
			continue
		}
		out = append(out, DiscoveredModel{
			ProviderID: "ollama",
			BaseURL:    base + "/v1",
			APIKey:     "ollama",
			Model:      m.Name,
			Local:      true,
		})
	}
	return out
}

func probeOpenAIModels(ctx context.Context, providerID, base string, timeout time.Duration) []DiscoveredModel {
	base = strings.TrimRight(base, "/")
	cctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(cctx, http.MethodGet, base+"/v1/models", nil)
	if err != nil {
		return nil
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil
	}

	var payload struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil
	}

	var out []DiscoveredModel
	for _, m := range payload.Data {
		if isEmbeddingModel(m.ID) {
			continue
		}
		out = append(out, DiscoveredModel{
			ProviderID: providerID,
			BaseURL:    base + "/v1",
			APIKey:     providerID,
			Model:      m.ID,
			Local:      true,
		})
	}
	return out
}

func isEmbeddingModel(name string) bool {
	lower := strings.ToLower(name)
	return strings.Contains(lower, "embed") || strings.Contains(lower, "bge-")
}

// MetaModelSpec describes a model from the Meta Model API (https://dev.meta.ai/docs/overview).
type MetaModelSpec struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Tier          string `json:"tier"` // T1, T3, T4, edge, media, vision
	ContextTokens int    `json:"context_tokens"`
	OpenWeights   bool   `json:"open_weights"`
	Description   string `json:"description"`
}

// MetaMuseCatalog returns the supported Meta Muse & SAM 3.1 models for Mímir.
func MetaMuseCatalog() []MetaModelSpec {
	return []MetaModelSpec{
		{
			ID:            "muse-wonder-1.0",
			Name:          "Muse Wonder 1.0",
			Tier:          "T1",
			ContextTokens: 1000000,
			OpenWeights:   false,
			Description:   "Meta flagship frontier reasoning & long-horizon coding agent model (1M context, parallel tools, Active Memory/Focus)",
		},
		{
			ID:            "muse-spark-1.0",
			Name:          "Muse Spark 1.0",
			Tier:          "T3",
			ContextTokens: 1000000,
			OpenWeights:   false,
			Description:   "High-speed coding & agentic execution workhorse (1M context, streamed tool args)",
		},
		{
			ID:            "muse-spark-open",
			Name:          "Muse Spark Open",
			Tier:          "T3",
			ContextTokens: 131072,
			OpenWeights:   true,
			Description:   "Open-weight release of Muse Spark (BF16 / NVFP4 checkpoints for self-hosted GPU clusters)",
		},
		{
			ID:            "muse-glimmer",
			Name:          "Muse Glimmer",
			Tier:          "edge",
			ContextTokens: 65536,
			OpenWeights:   true,
			Description:   "Compact Apache 2.0 open-weight edge model for local quantization, fast routing, and speculative decoding",
		},
		{
			ID:            "muse-voice-transcribe",
			Name:          "Muse Voice Transcribe",
			Tier:          "media",
			ContextTokens: 0,
			OpenWeights:   false,
			Description:   "Streaming & batch speech-to-text with speaker attribution and turn detection",
		},
		{
			ID:            "muse-image",
			Name:          "Muse Image",
			Tier:          "media",
			ContextTokens: 0,
			OpenWeights:   false,
			Description:   "Multi-turn text-to-image generation and iterative visual editing",
		},
		{
			ID:            "sam-3.1",
			Name:          "Segment Anything 3.1",
			Tier:          "vision",
			ContextTokens: 0,
			OpenWeights:   true,
			Description:   "Open-weight image & video object segmentation returning bounding boxes and masks from text prompts",
		},
	}
}

func firstEnv(keys ...string) string {
	for _, k := range keys {
		if v := os.Getenv(k); v != "" {
			return v
		}
	}
	return ""
}

func firstEnvDefault(def string, keys ...string) string {
	if v := firstEnv(keys...); v != "" {
		return v
	}
	return def
}
