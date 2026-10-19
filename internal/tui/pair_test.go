package tui

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/David2024patton/Mimir/internal/agent"
	"github.com/David2024patton/Mimir/internal/cortex"
	"github.com/David2024patton/Mimir/internal/llm"
	"github.com/David2024patton/Mimir/internal/tools"
)

// runTUIPair drives the TUI through /pair against the given WebURL and
// returns everything it printed.
func runTUIPair(t *testing.T, webURL string) string {
	t.Helper()
	store, err := cortex.NewMemoryStoreAt(filepath.Join(t.TempDir(), "cortex.json"))
	if err != nil {
		t.Fatalf("store: %v", err)
	}
	reg := tools.Default(t.TempDir())
	ag := agent.New(agent.Config{
		Provider: &fakeTUIProvider{},
		Tools:    reg,
		Cortex:   store,
		Model:    "qwen3:8b",
		Usage:    llm.NewTracker(),
		Local:    true,
	})
	var out bytes.Buffer
	err = Run(context.Background(), Options{
		Version:   "0.2.0-test",
		Agent:     ag,
		Tools:     reg,
		Cortex:    store,
		StoreDesc: "file-backed test",
		WebURL:    webURL,
		In:        strings.NewReader("/pair\n/quit\n"),
		Out:       &out,
	})
	if err != nil {
		t.Fatalf("TUI Run failed: %v", err)
	}
	return out.String()
}

// TestTUIPairUsesServerSession proves FIX 5: /pair asks the running
// server for a pairing session (whose relay manager lives as long as the
// server) instead of minting one in a throwaway in-process manager whose
// QR could never complete pairing.
func TestTUIPairUsesServerSession(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/pair" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"session":{"id":"srv-session-123"},"pair_uri":"mimir://pair?session=srv-session-123","ascii_qr":"QR-ART-FROM-SERVER"}`))
	}))
	defer srv.Close()

	got := runTUIPair(t, srv.URL)
	if !strings.Contains(got, "srv-session-123") {
		t.Errorf("/pair did not present the server-issued session.\nFull output:\n%s", got)
	}
	if !strings.Contains(got, "QR-ART-FROM-SERVER") {
		t.Errorf("/pair did not present the server-issued QR payload.\nFull output:\n%s", got)
	}
}

// TestTUIPairWithoutServerExplains proves the failure path: with no
// reachable server, /pair says so and explains how to start one, instead
// of printing a QR for a session that can never pair.
func TestTUIPairWithoutServerExplains(t *testing.T) {
	got := runTUIPair(t, "http://127.0.0.1:1")
	if !strings.Contains(got, "could not reach") {
		t.Errorf("/pair against a dead server should explain the problem.\nFull output:\n%s", got)
	}
	if strings.Contains(got, "Scan this QR code") {
		t.Errorf("/pair printed a scannable QR for a session that can never pair.\nFull output:\n%s", got)
	}
}
