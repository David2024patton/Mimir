package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// Config is the top-level mimir.json configuration.
type Config struct {
	Port      int                 `json:"port"`
	LogLevel  string              `json:"log_level"`
	Home      string              `json:"home"`
	Telemetry bool                `json:"telemetry"`
	Providers map[string]Provider `json:"providers"`
	MCP       map[string]MCPServer `json:"mcp"`
}

// Provider configures one model backend.
type Provider struct {
	Name    string `json:"name"`
	BaseURL string `json:"base_url"`
	APIKey  string `json:"api_key"` // supports {env:VAR} / {file:path}
	Dialect string `json:"dialect"` // openai_compat | anthropic
}

// MCPServer configures an external MCP tool server.
type MCPServer struct {
	Command []string          `json:"command"`
	URL     string            `json:"url"`
	Env     map[string]string `json:"env"`
	Enabled bool              `json:"enabled"`
}

// Default returns a config with sane defaults.
func Default() *Config {
	home, _ := os.UserHomeDir()
	return &Config{
		Port:      8420,
		LogLevel:  "info",
		Home:      filepath.Join(home, ".mimir"),
		Telemetry: true, // on by default; opt-out in privacy settings (F32)
		Providers: map[string]Provider{},
		MCP:       map[string]MCPServer{},
	}
}

// Load reads mimir.json (or the opencode-compatible fallback) and applies defaults.
func Load(path string) (*Config, error) {
	cfg := Default()
	if path == "" {
		path = "mimir.json"
		if _, err := os.Stat(path); err != nil {
			path = "opencode.json" // import opencode-style config
		}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil // no config file: run with defaults
		}
		return nil, err
	}
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	cfg.interpolate()
	return cfg, nil
}

func (c *Config) interpolate() {
	for id, p := range c.Providers {
		p.APIKey = Interpolate(p.APIKey)
		c.Providers[id] = p
	}
}

// Interpolate replaces {env:VAR} and {file:path} tokens in s.
func Interpolate(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); {
		if strings.HasPrefix(s[i:], "{env:") {
			end := strings.Index(s[i:], "}")
			if end < 0 {
				b.WriteString(s[i:])
				break
			}
			b.WriteString(os.Getenv(s[i+5 : i+end]))
			i += end + 1
			continue
		}
		if strings.HasPrefix(s[i:], "{file:") {
			end := strings.Index(s[i:], "}")
			if end < 0 {
				b.WriteString(s[i:])
				break
			}
			if data, err := os.ReadFile(s[i+6 : i+end]); err == nil {
				b.WriteString(strings.TrimSpace(string(data)))
			}
			i += end + 1
			continue
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}
