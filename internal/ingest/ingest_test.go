package ingest

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/David2024patton/Mimir/internal/cortex"
)

func TestIngestLocalDirAndPersistMindMapToCortex(t *testing.T) {
	tmp := t.TempDir()
	srcDir := filepath.Join(tmp, "sample-repo")
	if err := os.MkdirAll(filepath.Join(srcDir, "pkg"), 0o755); err != nil {
		t.Fatal(err)
	}
	goFile := `// Package pkg provides sample math helpers.
package pkg

import "fmt"

func ComputeHash(input string) string {
	return fmt.Sprintf("%x", input)
}

type Engine struct {
	Name string
}
`
	testFile := `package pkg

import "testing"

func TestComputeHash(t *testing.T) {
	if ComputeHash("a") == "" {
		t.Fatal("empty")
	}
}
`
	_ = os.WriteFile(filepath.Join(srcDir, "pkg", "engine.go"), []byte(goFile), 0o644)
	_ = os.WriteFile(filepath.Join(srcDir, "pkg", "engine_test.go"), []byte(testFile), 0o644)

	store, err := cortex.NewMemoryStoreAt(filepath.Join(tmp, "cortex.json"))
	if err != nil {
		t.Fatal(err)
	}

	digestPath := filepath.Join(tmp, "digest.txt")
	res, err := Run(context.Background(), Options{
		Source:        srcDir,
		WriteDigestTo: digestPath,
	}, store)
	if err != nil {
		t.Fatalf("ingest.Run failed: %v", err)
	}
	if res.MindMap.TotalFiles != 2 {
		t.Fatalf("got %d files, want 2", res.MindMap.TotalFiles)
	}
	if res.NeuronsSaved < 2 {
		t.Errorf("expected at least 2 neurons saved to Cortex, got %d", res.NeuronsSaved)
	}
	if !strings.Contains(res.Digest, "FILE: pkg/engine.go") {
		t.Errorf("digest missing FILE header:\n%s", res.Digest)
	}

	mm, ok := LatestMindMap(store)
	if !ok {
		t.Fatal("expected LatestMindMap to recover MindMap from Cortex")
	}
	compact := FormatMindMapCompact(*mm, 20)
	if !strings.Contains(compact, "ComputeHash") || !strings.Contains(compact, "Engine") {
		t.Errorf("expected MindMap compact output to include extracted symbols, got:\n%s", compact)
	}
}
