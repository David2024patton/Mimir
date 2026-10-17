package config

import "os"

// ProviderEntry describes one built-in LLM provider in Mímir's OpenCode/TrueForge-style
// Provider Registry (supporting both BYOK and Mímir Well Subscription Gateway).
type ProviderEntry struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	BaseURL      string   `json:"base_url"`
	EnvKey       string   `json:"env_key"`
	DefaultModel string   `json:"default_model"`
	Models       []string `json:"models"`
	BillingMode  string   `json:"billing_mode"` // "subscription", "byok", or "local_free"
	Configured   bool     `json:"configured"`
}

// SubscriptionTier describes a Mímir Well managed subscription plan (inspired by
// OpenCode Zen & TrueForge Gateway) for users who want 1-click access without managing 10 API keys.
type SubscriptionTier struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Price       string   `json:"price"`
	Included    []string `json:"included"`
	Description string   `json:"description"`
}

// BuiltinSubscriptionTiers returns Mímir's hybrid BYOK + Managed Subscription tiers.
func BuiltinSubscriptionTiers() []SubscriptionTier {
	return []SubscriptionTier{
		{
			ID:    "sovereign-free",
			Name:  "Mímir Sovereign (BYOK & Local)",
			Price: "$0 / mo",
			Included: []string{
				"Bundled Liquid AI d1-omni-600M System-One Decision Model",
				"Unlimited local Ollama / LM Studio / llama.cpp + Meta Muse Glimmer",
				"Unlimited BYOK across all 14+ cloud providers",
				"Full Yggdrasil Mind Map & GitIngest Cortex memory",
			},
			Description: "100% free local-first & Bring-Your-Own-Key mode with zero lock-in.",
		},
		{
			ID:    "well-pro",
			Name:  "Mímir Well Pro (Managed Gateway)",
			Price: "$20 / mo",
			Included: []string{
				"1-key access (MIMIR_SUBSCRIPTION_KEY) to Muse Wonder 1.0, Muse Spark 1.0, MiMo-v2-Pro, Claude Sonnet 4.6, GPT-5, DeepSeek V3.2",
				"Hosted E2EE Mobile Relay for the Mímir Flutter Phone App",
				"Cloud-synced encrypted Cortex backup across Desktop & Phone",
			},
			Description: "OpenCode Zen-style unified subscription gateway: no separate API keys needed.",
		},
		{
			ID:    "odin-max",
			Name:  "Mímir Odin Max (Swarm & Remote Sandbox)",
			Price: "$100 / mo",
			Included: []string{
				"Everything in Well Pro + 5x frontier token budget (Muse Wonder Full / Opus 4.6)",
				"Parallel Sub-Agent Swarms (Odin + Thor + Loki + Heimdall running concurrently)",
				"Crabbox-style Remote Test & Build Offloading so E2E tests never slow down your laptop",
			},
			Description: "High-concurrency multi-agent swarm & offloaded cloud verification workers.",
		},
	}
}

// BuiltinProviders returns Mímir's built-in OpenCode-style provider catalog, marking
// which providers are currently active via environment variables or local servers.
func BuiltinProviders() []ProviderEntry {
	entries := []ProviderEntry{
		{
			ID:           "mimir-well",
			Name:         "Mímir Well Gateway (Subscription)",
			BaseURL:      "https://api.mimirmind.com/v1",
			EnvKey:       "MIMIR_SUBSCRIPTION_KEY",
			DefaultModel: "muse-wonder-1.0",
			Models:       []string{"muse-wonder-1.0", "muse-spark-1.0", "mimo-v2-pro", "claude-sonnet-4-6", "gpt-5", "deepseek-v3.2"},
			BillingMode:  "subscription",
		},
		{
			ID:           "meta",
			Name:         "Meta Model API (Muse)",
			BaseURL:      "https://api.meta.ai/v1",
			EnvKey:       "META_API_KEY",
			DefaultModel: "muse-wonder-1.0",
			Models:       []string{"muse-wonder-1.0", "muse-spark-1.0", "muse-spark-open", "muse-glimmer", "muse-voice-transcribe", "muse-image", "sam-3.1"},
			BillingMode:  "byok",
		},
		{
			ID:           "mimo",
			Name:         "Xiaomi MiMo",
			BaseURL:      "https://api.xiaomimimo.com/v1",
			EnvKey:       "MIMO_API_KEY",
			DefaultModel: "mimo-v2-pro",
			Models:       []string{"mimo-v2-pro", "mimo-v2-flash", "mimo-v2-omni"},
			BillingMode:  "byok",
		},
		{
			ID:           "anthropic",
			Name:         "Anthropic Claude",
			BaseURL:      "https://api.anthropic.com/v1",
			EnvKey:       "ANTHROPIC_API_KEY",
			DefaultModel: "claude-sonnet-4-6",
			Models:       []string{"claude-opus-4-6", "claude-sonnet-4-6", "claude-haiku-4-5"},
			BillingMode:  "byok",
		},
		{
			ID:           "openai",
			Name:         "OpenAI",
			BaseURL:      "https://api.openai.com/v1",
			EnvKey:       "OPENAI_API_KEY",
			DefaultModel: "gpt-5",
			Models:       []string{"gpt-5", "gpt-5-mini", "o4-mini", "gpt-4o-mini"},
			BillingMode:  "byok",
		},
		{
			ID:           "gemini",
			Name:         "Google Gemini",
			BaseURL:      "https://generativelanguage.googleapis.com/v1beta/openai",
			EnvKey:       "GEMINI_API_KEY",
			DefaultModel: "gemini-2.5-pro",
			Models:       []string{"gemini-2.5-pro", "gemini-2.5-flash", "gemini-2.5-flash-lite"},
			BillingMode:  "byok",
		},
		{
			ID:           "deepseek",
			Name:         "DeepSeek",
			BaseURL:      "https://api.deepseek.com/v1",
			EnvKey:       "DEEPSEEK_API_KEY",
			DefaultModel: "deepseek-chat",
			Models:       []string{"deepseek-chat", "deepseek-reasoner"},
			BillingMode:  "byok",
		},
		{
			ID:           "xai",
			Name:         "xAI Grok",
			BaseURL:      "https://api.x.ai/v1",
			EnvKey:       "XAI_API_KEY",
			DefaultModel: "grok-4",
			Models:       []string{"grok-4", "grok-4-fast", "grok-code-fast-1"},
			BillingMode:  "byok",
		},
		{
			ID:           "groq",
			Name:         "Groq LPU",
			BaseURL:      "https://api.groq.com/openai/v1",
			EnvKey:       "GROQ_API_KEY",
			DefaultModel: "llama-3.3-70b-versatile",
			Models:       []string{"llama-3.3-70b-versatile", "qwen-qwq-32b"},
			BillingMode:  "byok",
		},
		{
			ID:           "openrouter",
			Name:         "OpenRouter (300+ Models)",
			BaseURL:      "https://openrouter.ai/api/v1",
			EnvKey:       "OPENROUTER_API_KEY",
			DefaultModel: "meta-llama/muse-wonder-1.0",
			Models:       []string{"meta-llama/muse-wonder-1.0", "xiaomi/mimo-v2-pro", "liquid/d1-omni-600m", "anthropic/claude-sonnet-4-6"},
			BillingMode:  "byok",
		},
		{
			ID:           "ollama",
			Name:         "Ollama (Local Sovereign)",
			BaseURL:      "http://127.0.0.1:11434/v1",
			EnvKey:       "MIMIR_OLLAMA_URL",
			DefaultModel: "liquid/d1-omni-600m",
			Models:       []string{"liquid/d1-omni-600m", "liquid/d1-3b", "muse-glimmer", "qwen3:8b", "qwen3-coder:30b"},
			BillingMode:  "local_free",
		},
		{
			ID:           "lmstudio",
			Name:         "LM Studio (Local Sovereign)",
			BaseURL:      "http://127.0.0.1:1234/v1",
			EnvKey:       "MIMIR_LMSTUDIO_URL",
			DefaultModel: "qwen3-8b-instruct",
			Models:       []string{"qwen3-8b-instruct", "muse-glimmer-gguf", "liquid-d1-omni-600m-gguf"},
			BillingMode:  "local_free",
		},
	}

	for i := range entries {
		if entries[i].BillingMode == "local_free" {
			entries[i].Configured = true
		} else if entries[i].EnvKey != "" && os.Getenv(entries[i].EnvKey) != "" {
			entries[i].Configured = true
		}
	}
	return entries
}
