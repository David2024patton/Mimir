package cortex

import (
	"strings"
	"testing"
)

// FIX 4: the sidecar password must never appear in the process argument
// list, where any local user can read it via /proc/<pid>/cmdline for the
// life of the process. Credentials travel in the SURREAL_USER /
// SURREAL_PASS environment variables instead. These are code-level tests:
// no surreal binary exists in this environment, so the constructed
// command is inspected, not executed against a live server.
func TestSidecarStartCommandHidesPasswordFromArgv(t *testing.T) {
	const password = "per-run-secret-9f27c1"
	sc := &Sidecar{
		BinPath: "/nonexistent/surreal",
		Addr:    "127.0.0.1:8000",
		User:    "root",
		Pass:    password,
	}
	cmd := sc.startCommand("memory")

	for _, arg := range cmd.Args {
		if arg == "--pass" || arg == "--user" {
			t.Fatalf("credential flag %q present in argv: %v", arg, cmd.Args)
		}
		if strings.Contains(arg, password) {
			t.Fatalf("password present in argv: %v", cmd.Args)
		}
	}

	var gotUser, gotPass string
	for _, kv := range cmd.Env {
		if v, ok := strings.CutPrefix(kv, "SURREAL_USER="); ok {
			gotUser = v
		}
		if v, ok := strings.CutPrefix(kv, "SURREAL_PASS="); ok {
			gotPass = v
		}
	}
	if gotUser != "root" {
		t.Fatalf("SURREAL_USER in env = %q, want %q", gotUser, "root")
	}
	if gotPass != password {
		t.Fatal("SURREAL_PASS in env does not carry the sidecar password")
	}
}

// FIX 4: inherited SURREAL_USER / SURREAL_PASS from the parent environment
// must not shadow the sidecar's own credentials; the constructed Env
// carries exactly one entry for each.
func TestSidecarStartCommandOverridesInheritedCreds(t *testing.T) {
	t.Setenv("SURREAL_USER", "inherited-user")
	t.Setenv("SURREAL_PASS", "inherited-pass")
	sc := &Sidecar{
		BinPath: "/nonexistent/surreal",
		Addr:    "127.0.0.1:8000",
		User:    "sidecar-user",
		Pass:    "sidecar-pass",
	}
	cmd := sc.startCommand("memory")

	var users, passes []string
	for _, kv := range cmd.Env {
		if v, ok := strings.CutPrefix(kv, "SURREAL_USER="); ok {
			users = append(users, v)
		}
		if v, ok := strings.CutPrefix(kv, "SURREAL_PASS="); ok {
			passes = append(passes, v)
		}
	}
	if len(users) != 1 || users[0] != "sidecar-user" {
		t.Fatalf("SURREAL_USER entries = %v, want exactly [sidecar-user]", users)
	}
	if len(passes) != 1 || passes[0] != "sidecar-pass" {
		t.Fatalf("SURREAL_PASS entries = %v, want exactly [sidecar-pass]", passes)
	}
}
