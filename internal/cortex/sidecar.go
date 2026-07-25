package cortex

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// Sidecar manages a local SurrealDB server process (ADR-003: the daemon auto-starts a
// managed sidecar). It reuses an already-healthy server if one is running, otherwise it
// launches the binary and waits for health.
type Sidecar struct {
	BinPath  string // path to the surreal binary (auto-located if empty)
	Addr     string // bind address, e.g. 127.0.0.1:8000
	User     string // default root
	Pass     string // default root
	DataPath string // "memory" if empty, else a file/dir path for durable storage

	cmd  *exec.Cmd
	owns bool // true if we started the process (and so should stop it)
}

// LocateSurreal finds the surreal binary: MIMIR_SURREAL_BIN, then PATH, then bin/surreal
// (walking up from the working directory), then next to the running executable.
func LocateSurreal() string {
	if p := os.Getenv("MIMIR_SURREAL_BIN"); p != "" {
		return p
	}
	if p, err := exec.LookPath("surreal"); err == nil {
		return p
	}
	names := []string{filepath.Join("bin", "surreal.exe"), filepath.Join("bin", "surreal"), "surreal.exe"}
	if dir, err := os.Getwd(); err == nil {
		for {
			for _, n := range names {
				cand := filepath.Join(dir, n)
				if _, err := os.Stat(cand); err == nil {
					return cand
				}
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		for _, n := range names {
			cand := filepath.Join(exeDir, n)
			if _, err := os.Stat(cand); err == nil {
				return cand
			}
		}
	}
	return ""
}

// Healthy reports whether the SurrealDB HTTP /health endpoint responds OK.
func (s *Sidecar) Healthy(ctx context.Context) bool {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+s.Addr+"/health", nil)
	if err != nil {
		return false
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// Start ensures a SurrealDB server is listening on s.Addr, launching the binary if
// needed, and blocks until it is healthy (or the context is done / timeout).
func (s *Sidecar) Start(ctx context.Context) error {
	if s.Addr == "" {
		s.Addr = "127.0.0.1:8000"
	}
	if s.User == "" {
		s.User = "root"
	}
	if s.Pass == "" {
		s.Pass = "root"
	}
	if s.Healthy(ctx) {
		return nil // reuse an already-running server
	}
	if s.BinPath == "" {
		s.BinPath = LocateSurreal()
	}
	if s.BinPath == "" {
		return fmt.Errorf("surreal binary not found: set MIMIR_SURREAL_BIN or put surreal on PATH")
	}
	data := "memory"
	if s.DataPath != "" && s.DataPath != "memory" {
		data = "file:" + s.DataPath
	}
	s.cmd = exec.Command(s.BinPath, "start", "--user", s.User, "--pass", s.Pass, "--bind", s.Addr, data)
	if err := s.cmd.Start(); err != nil {
		return fmt.Errorf("surreal: start: %w", err)
	}
	s.owns = true
	deadline := time.Now().Add(20 * time.Second)
	for time.Now().Before(deadline) {
		if s.Healthy(ctx) {
			return nil
		}
		select {
		case <-ctx.Done():
			s.Stop()
			return ctx.Err()
		case <-time.After(250 * time.Millisecond):
		}
	}
	s.Stop()
	return fmt.Errorf("surreal: timed out waiting for health on %s", s.Addr)
}

// Stop terminates the sidecar if this Sidecar started it.
func (s *Sidecar) Stop() {
	if s.owns && s.cmd != nil && s.cmd.Process != nil {
		_ = s.cmd.Process.Kill()
		_, _ = s.cmd.Process.Wait()
		s.owns = false
	}
}

// HTTPAddr returns the base URL of the SurrealDB HTTP API.
func (s *Sidecar) HTTPAddr() string { return "http://" + s.Addr }
