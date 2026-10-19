package cortex

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

// F19: no source file in this package may assign the literal "root" as a
// password ever again. Usernames may default to the root account name; a
// known default password may not exist.
func TestNoRootPasswordLiteral(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	dir := filepath.Dir(thisFile)
	passRoot := regexp.MustCompile(`(?i)\bpass\w*\s*[:=]\s*"root"`)
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read package dir: %v", err)
	}
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		b, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		if loc := passRoot.FindIndex(b); loc != nil {
			t.Errorf("%s assigns \"root\" as a password near byte %d", name, loc[0])
		}
	}
}

func clearSurrealEnv(t *testing.T) {
	t.Helper()
	t.Setenv(envSurrealUser, "")
	t.Setenv(envSurrealPass, "")
}

func TestResolveSurrealCredsExplicitWins(t *testing.T) {
	t.Setenv(envSurrealUser, "envuser")
	t.Setenv(envSurrealPass, "envpass")
	user, pass, err := resolveSurrealCreds("cfguser", "cfgpass", "http://127.0.0.1:9")
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if user != "cfguser" || pass != "cfgpass" {
		t.Fatalf("got %q/%q, want cfguser/cfgpass", user, pass)
	}
}

func TestResolveSurrealCredsFromEnv(t *testing.T) {
	t.Setenv(envSurrealUser, "envuser")
	t.Setenv(envSurrealPass, "envpass")
	user, pass, err := resolveSurrealCreds("", "", "http://127.0.0.1:9")
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if user != "envuser" || pass != "envpass" {
		t.Fatalf("got %q/%q, want envuser/envpass", user, pass)
	}
}

func TestResolveSurrealCredsFromSidecarRegistry(t *testing.T) {
	clearSurrealEnv(t)
	addr := fmt.Sprintf("http://127.0.0.1:%d", freePort(t))
	registerSidecarCreds(addr, "sidecaruser", "sidecarpass")
	user, pass, err := resolveSurrealCreds("", "", addr)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if user != "sidecaruser" || pass != "sidecarpass" {
		t.Fatalf("got %q/%q, want the registered sidecar credentials", user, pass)
	}
}

func TestResolveSurrealCredsMissing(t *testing.T) {
	clearSurrealEnv(t)
	addr := fmt.Sprintf("http://127.0.0.1:%d", freePort(t))
	_, _, err := resolveSurrealCreds("", "", addr)
	if err == nil {
		t.Fatal("expected an error when no credentials are available")
	}
	if !strings.Contains(err.Error(), envSurrealPass) {
		t.Fatalf("error should name %s, got: %v", envSurrealPass, err)
	}
}

// NewSurrealStore must fail cleanly with the credential error, before any
// network access, when nothing provides a password.
func TestNewSurrealStoreMissingCreds(t *testing.T) {
	clearSurrealEnv(t)
	addr := fmt.Sprintf("http://127.0.0.1:%d", freePort(t))
	_, err := NewSurrealStore(context.Background(), SurrealConfig{Addr: addr})
	if err == nil {
		t.Fatal("expected an error when no credentials are configured")
	}
	if !strings.Contains(err.Error(), "no credentials") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// A zero-config Sidecar generates a per-run password and registers it, so
// the credential the client resolves for the sidecar address matches the
// one the sidecar will launch with.
func TestSidecarGeneratesMatchingPassword(t *testing.T) {
	clearSurrealEnv(t)
	sc := &Sidecar{
		Addr:     fmt.Sprintf("127.0.0.1:%d", freePort(t)),
		BinPath:  filepath.Join(t.TempDir(), "no-such-surreal-binary"),
		DataPath: "memory",
	}
	// No binary exists, so Start must fail; the password is generated and
	// registered before the launch attempt, which is what this test checks.
	if err := sc.Start(context.Background()); err == nil {
		t.Fatal("expected Start to fail with a nonexistent binary")
	}
	if sc.Pass == "" {
		t.Fatal("sidecar did not generate a password")
	}
	if len(sc.Pass) != 64 {
		t.Fatalf("generated password length = %d, want 64 hex chars", len(sc.Pass))
	}
	if _, err := hex.DecodeString(sc.Pass); err != nil {
		t.Fatalf("generated password is not hex: %v", err)
	}
	user, pass, err := resolveSurrealCreds("", "", sc.HTTPAddr())
	if err != nil {
		t.Fatalf("client cannot resolve sidecar credentials: %v", err)
	}
	if pass != sc.Pass || user != sc.User {
		t.Fatal("client-resolved credentials do not match the sidecar credentials")
	}
}

func TestGenerateSurrealPasswordRandom(t *testing.T) {
	a, err := generateSurrealPassword()
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	b, err := generateSurrealPassword()
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if a == b {
		t.Fatal("two generated passwords are identical")
	}
	if len(a) != 64 {
		t.Fatalf("password length = %d, want 64", len(a))
	}
}
