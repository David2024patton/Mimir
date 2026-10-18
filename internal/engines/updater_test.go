package engines

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// publishTree builds a fake "published main" directory served over HTTP.
func publishTree(t *testing.T, payload string, owner string) (string, Manifest) {
	t.Helper()
	pub := t.TempDir()
	binDir := filepath.Join(pub, "engines", "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	bin := filepath.Join(binDir, "mimir_tyr_engine.exe")
	if err := os.WriteFile(bin, []byte(payload), 0o755); err != nil {
		t.Fatal(err)
	}
	sum, size, _ := FileSHA256(bin)
	m := Manifest{
		SchemaVersion: "1.0", Owner: owner, Website: "https://alphaflux.net",
		License: "LICENSE-ENGINES.md", Attribution: AttributionNotice,
		Engines: []EngineEntry{{
			ID: "tyr-core", Version: "0.3.0", Binary: "engines/bin/mimir_tyr_engine.exe",
			SHA256: sum, SizeBytes: size,
		}},
	}
	raw, _ := json.Marshal(m)
	if err := os.WriteFile(filepath.Join(binDir, "manifest.json"), raw, 0o644); err != nil {
		t.Fatal(err)
	}
	return pub, m
}

func TestUpdateFlow_DetectsChangeAndSwapsVerifiedBinary(t *testing.T) {
	pub, _ := publishTree(t, "engine-v0.3.0", RequiredOwner)
	srv := httptest.NewServer(http.FileServer(http.Dir(pub)))
	defer srv.Close()

	local := t.TempDir()
	localBin := filepath.Join(local, "engines", "bin")
	_ = os.MkdirAll(localBin, 0o755)
	_ = os.WriteFile(filepath.Join(localBin, "mimir_tyr_engine.exe"), []byte("engine-v0.2.0"), 0o755)

	plan, err := CheckForUpdates(context.Background(), local, srv.URL)
	if err != nil {
		t.Fatalf("CheckForUpdates: %v", err)
	}
	if len(plan.Changes) != 1 || plan.Changes[0].ID != "tyr-core" {
		t.Fatalf("expected 1 changed engine, got %+v", plan.Changes)
	}

	updated, err := ApplyUpdates(context.Background(), local, plan)
	if err != nil || len(updated) != 1 {
		t.Fatalf("ApplyUpdates: updated=%v err=%v", updated, err)
	}
	got, _ := os.ReadFile(filepath.Join(localBin, "mimir_tyr_engine.exe"))
	if string(got) != "engine-v0.3.0" {
		t.Fatalf("binary not swapped, got %q", got)
	}

	// Second check is a no-op.
	plan2, err := CheckForUpdates(context.Background(), local, srv.URL)
	if err != nil || len(plan2.Changes) != 0 {
		t.Fatalf("expected up to date, got %+v err=%v", plan2.Changes, err)
	}
}

func TestUpdateFlow_RejectsTamperedDownload(t *testing.T) {
	pub, _ := publishTree(t, "engine-v0.3.0", RequiredOwner)
	// Tamper after the manifest was written: hash no longer matches.
	_ = os.WriteFile(filepath.Join(pub, "engines", "bin", "mimir_tyr_engine.exe"), []byte("evil"), 0o755)
	srv := httptest.NewServer(http.FileServer(http.Dir(pub)))
	defer srv.Close()

	local := t.TempDir()
	localBin := filepath.Join(local, "engines", "bin")
	_ = os.MkdirAll(localBin, 0o755)
	orig := filepath.Join(localBin, "mimir_tyr_engine.exe")
	_ = os.WriteFile(orig, []byte("engine-v0.2.0"), 0o755)

	plan, err := CheckForUpdates(context.Background(), local, srv.URL)
	if err != nil {
		t.Fatalf("CheckForUpdates: %v", err)
	}
	if _, err := ApplyUpdates(context.Background(), local, plan); !errors.Is(err, ErrBinaryIntegrity) {
		t.Fatalf("expected ErrBinaryIntegrity, got %v", err)
	}
	got, _ := os.ReadFile(orig)
	if string(got) != "engine-v0.2.0" {
		t.Fatalf("original binary must be untouched after failed verify, got %q", got)
	}
	if _, err := os.Stat(orig + ".download"); !os.IsNotExist(err) {
		t.Fatalf("temp download should be cleaned up")
	}
}

func TestUpdateFlow_RejectsManifestWithoutAlphaFlux(t *testing.T) {
	pub, _ := publishTree(t, "engine-v0.3.0", "Someone Else Inc")
	srv := httptest.NewServer(http.FileServer(http.Dir(pub)))
	defer srv.Close()
	if _, err := CheckForUpdates(context.Background(), t.TempDir(), srv.URL); !errors.Is(err, ErrAttributionRequired) {
		t.Fatalf("expected ErrAttributionRequired, got %v", err)
	}
}
