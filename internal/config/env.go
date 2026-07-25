package config

import "os"

// EnvProvider holds provider settings read from the environment, so Mímir can run
// a turn with no config file - just MIMIR_BASE_URL (+ key + model).
type EnvProvider struct {
	ID      string
	BaseURL string
	APIKey  string
	Model   string
}

// ProviderFromEnv returns provider settings from the environment. ok is false when
// MIMIR_BASE_URL is unset (no provider configured).
func ProviderFromEnv() (EnvProvider, bool) {
	base := os.Getenv("MIMIR_BASE_URL")
	if base == "" {
		return EnvProvider{}, false
	}
	return EnvProvider{
		ID:      envOr("MIMIR_PROVIDER_ID", "openai-compatible"),
		BaseURL: base,
		APIKey:  os.Getenv("MIMIR_API_KEY"),
		Model:   os.Getenv("MIMIR_MODEL"),
	}, true
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
