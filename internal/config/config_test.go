package config

import (
	"os"
	"testing"
)

func TestInterpolateEnv(t *testing.T) {
	os.Setenv("MIMIR_TEST_KEY", "secret123")
	defer os.Unsetenv("MIMIR_TEST_KEY")
	if got, want := Interpolate("key={env:MIMIR_TEST_KEY}"), "key=secret123"; got != want {
		t.Errorf("Interpolate env = %q, want %q", got, want)
	}
}

func TestInterpolateMissingEnvIsEmpty(t *testing.T) {
	if got, want := Interpolate("key={env:MIMIR_DOES_NOT_EXIST}"), "key="; got != want {
		t.Errorf("Interpolate missing env = %q, want %q", got, want)
	}
}

func TestInterpolateFile(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "secret")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString("file-secret\n"); err != nil {
		t.Fatal(err)
	}
	f.Close()
	if got, want := Interpolate("{file:"+f.Name()+"}"), "file-secret"; got != want {
		t.Errorf("Interpolate file = %q, want %q", got, want)
	}
}

func TestDefault(t *testing.T) {
	cfg := Default()
	if cfg.Port != 8420 {
		t.Errorf("default port = %d, want 8420", cfg.Port)
	}
	if !cfg.Telemetry {
		t.Error("telemetry should default to true (opt-out)")
	}
}
