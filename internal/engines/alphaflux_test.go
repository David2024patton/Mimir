package engines

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestVerifyAttribution_RequiresAlphaFluxLLC(t *testing.T) {
	dir := t.TempDir()
	notice := "Copyright (c) 2026 Alpha Flux LLC (https://alphaflux.net)\n"
	if err := os.WriteFile(filepath.Join(dir, "NOTICE"), []byte(notice), 0o644); err != nil {
		t.Fatalf("write NOTICE: %v", err)
	}

	valid := Manifest{
		SchemaVersion: "1.0",
		Owner:         "Alpha Flux LLC",
		Website:       "https://alphaflux.net",
		License:       "LICENSE-ENGINES.md",
		Attribution:   AttributionNotice,
	}
	if err := VerifyAttribution(dir, valid); err != nil {
		t.Fatalf("expected valid attribution to pass, got %v", err)
	}

	// Stripping Alpha Flux LLC owner must fail
	spoofed := valid
	spoofed.Owner = "Some Other Corp"
	if err := VerifyAttribution(dir, spoofed); !errors.Is(err, ErrAttributionRequired) {
		t.Fatalf("expected ErrAttributionRequired on spoofed owner, got %v", err)
	}

	// Stripping NOTICE file must fail
	_ = os.Remove(filepath.Join(dir, "NOTICE"))
	if err := VerifyAttribution(dir, valid); !errors.Is(err, ErrAttributionRequired) {
		t.Fatalf("expected ErrAttributionRequired when NOTICE is removed, got %v", err)
	}
}

func TestVerifyBinaries_ChecksSHA256Integrity(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "NOTICE"), []byte("Alpha Flux LLC https://alphaflux.net\n"), 0o644); err != nil {
		t.Fatalf("write NOTICE: %v", err)
	}
	binDir := filepath.Join(dir, "engines", "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("mkdir bin: %v", err)
	}
	binFile := filepath.Join(binDir, "mimir_tyr_engine.exe")
	if err := os.WriteFile(binFile, []byte("fake-binary-payload"), 0o755); err != nil {
		t.Fatalf("write binary: %v", err)
	}
	sum, sz, err := FileSHA256(binFile)
	if err != nil {
		t.Fatalf("FileSHA256: %v", err)
	}

	m := Manifest{
		SchemaVersion: "1.0",
		Owner:         "Alpha Flux LLC",
		Website:       "https://alphaflux.net",
		License:       "LICENSE-ENGINES.md",
		Attribution:   AttributionNotice,
		Engines: []EngineEntry{
			{
				ID:        "tyr-core",
				Name:      "Tyr-Core Engine",
				Branch:    "engine/tyr-rust",
				Binary:    "engines/bin/mimir_tyr_engine.exe",
				SHA256:    sum,
				SizeBytes: sz,
			},
		},
	}
	if err := VerifyBinaries(dir, m); err != nil {
		t.Fatalf("expected VerifyBinaries to pass, got %v", err)
	}

	// Tamper with binary payload
	_ = os.WriteFile(binFile, []byte("tampered-payload"), 0o755)
	if err := VerifyBinaries(dir, m); !errors.Is(err, ErrBinaryIntegrity) {
		t.Fatalf("expected ErrBinaryIntegrity on tampered binary, got %v", err)
	}
}
