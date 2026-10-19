package engines

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
)

// publishTree builds a fake "published main" directory served over HTTP.
func publishTree(t *testing.T, payload string, owner string) (string, Manifest) {
	t.Helper()
	// The fake publisher is a plain-http httptest server; opt in explicitly.
	t.Setenv("MIMIR_UPDATE_ALLOW_INSECURE", "1")
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

// N7: a plain http update base is refused unless the operator explicitly
// opts into an insecure channel.
func TestCheckForUpdates_RejectsPlainHTTPBase(t *testing.T) {
	pub, _ := publishTree(t, "engine-v0.3.0", RequiredOwner)
	// publishTree set the opt-in for its own flow; remove it for this check.
	if err := os.Unsetenv("MIMIR_UPDATE_ALLOW_INSECURE"); err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(http.FileServer(http.Dir(pub)))
	defer srv.Close()
	if _, err := CheckForUpdates(context.Background(), t.TempDir(), srv.URL); !errors.Is(err, ErrInsecureUpdateBase) {
		t.Fatalf("expected ErrInsecureUpdateBase, got %v", err)
	}
}

// writeManifest rewrites the manifest.json inside a published tree.
func writeManifest(t *testing.T, pub string, m Manifest) {
	t.Helper()
	raw, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pub, "engines", "bin", "manifest.json"), raw, 0o644); err != nil {
		t.Fatal(err)
	}
}

// N7: with MIMIR_ENGINE_PUBKEY pinned, unsigned manifests are refused,
// correctly signed ones pass, and tampering after signing is caught.
func TestCheckForUpdates_ManifestSignature(t *testing.T) {
	pub, m := publishTree(t, "engine-v0.3.0", RequiredOwner)
	srv := httptest.NewServer(http.FileServer(http.Dir(pub)))
	defer srv.Close()

	pubKey, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("MIMIR_ENGINE_PUBKEY", base64.StdEncoding.EncodeToString(pubKey))

	// Unsigned manifest + pinned key: refuse.
	if _, err := CheckForUpdates(context.Background(), t.TempDir(), srv.URL); err == nil {
		t.Fatal("expected refusal for unsigned manifest when a key is pinned")
	}

	// Signed manifest: accepted.
	sig, err := SignManifest(m, priv)
	if err != nil {
		t.Fatal(err)
	}
	m.Signature = sig
	writeManifest(t, pub, m)
	if _, err := CheckForUpdates(context.Background(), t.TempDir(), srv.URL); err != nil {
		t.Fatalf("expected signed manifest to pass, got %v", err)
	}

	// Tamper after signing (version bump, same signature): refused.
	m.Engines[0].Version = "9.9.9"
	writeManifest(t, pub, m)
	if _, err := CheckForUpdates(context.Background(), t.TempDir(), srv.URL); err == nil {
		t.Fatal("expected refusal for manifest tampered after signing")
	}
}

// N7: ApplyUpdates refuses to move an engine backwards unless forced.
func TestApplyUpdates_RefusesDowngrade(t *testing.T) {
	pub, _ := publishTree(t, "engine-v0.2.0", RequiredOwner)
	// Remote manifest from publishTree says 0.3.0; rewrite it as 0.2.0.
	raw, err := os.ReadFile(filepath.Join(pub, "engines", "bin", "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var remote Manifest
	if err := json.Unmarshal(raw, &remote); err != nil {
		t.Fatal(err)
	}
	remote.Engines[0].Version = "0.2.0"
	writeManifest(t, pub, remote)

	srv := httptest.NewServer(http.FileServer(http.Dir(pub)))
	defer srv.Close()

	// Local install is at 0.3.0 with different binary bytes.
	local := t.TempDir()
	localBin := filepath.Join(local, "engines", "bin")
	_ = os.MkdirAll(localBin, 0o755)
	_ = os.WriteFile(filepath.Join(localBin, "mimir_tyr_engine.exe"), []byte("engine-v0.3.0-local"), 0o755)
	localManifest := Manifest{
		SchemaVersion: "1.0", Owner: RequiredOwner, Website: "https://alphaflux.net",
		License: "LICENSE-ENGINES.md", Attribution: AttributionNotice,
		Engines: []EngineEntry{{ID: "tyr-core", Version: "0.3.0", Binary: "engines/bin/mimir_tyr_engine.exe"}},
	}
	lraw, _ := json.Marshal(localManifest)
	_ = os.WriteFile(filepath.Join(localBin, "manifest.json"), lraw, 0o644)

	plan, err := CheckForUpdates(context.Background(), local, srv.URL)
	if err != nil {
		t.Fatalf("CheckForUpdates: %v", err)
	}
	if len(plan.Changes) != 1 {
		t.Fatalf("expected 1 change, got %+v", plan.Changes)
	}
	if _, err := ApplyUpdates(context.Background(), local, plan); !errors.Is(err, ErrDowngradeRefused) {
		t.Fatalf("expected ErrDowngradeRefused, got %v", err)
	}
	got, _ := os.ReadFile(filepath.Join(localBin, "mimir_tyr_engine.exe"))
	if string(got) != "engine-v0.3.0-local" {
		t.Fatalf("binary must be untouched after refused downgrade, got %q", got)
	}

	// With the force env, the same downgrade applies.
	t.Setenv("MIMIR_ENGINE_ALLOW_DOWNGRADE", "1")
	if _, err := ApplyUpdates(context.Background(), local, plan); err != nil {
		t.Fatalf("forced downgrade should apply, got %v", err)
	}
	got, _ = os.ReadFile(filepath.Join(localBin, "mimir_tyr_engine.exe"))
	if string(got) != "engine-v0.2.0" {
		t.Fatalf("expected downgraded binary content, got %q", got)
	}
}

// N8: the updater downloads, verifies, and replaces the platform-resolved
// binary that Invoke executes, not the legacy top-level path.
func TestUpdateFlow_UsesPlatformBinary(t *testing.T) {
	t.Setenv("MIMIR_UPDATE_ALLOW_INSECURE", "1")
	hostKey := runtime.GOOS + "_" + runtime.GOARCH
	platName := fmt.Sprintf("mimir_tyr_engine_%s", hostKey)
	if runtime.GOOS == "windows" {
		platName += ".exe"
	}
	platRel := "engines/bin/" + platName

	pub := t.TempDir()
	pubBin := filepath.Join(pub, "engines", "bin")
	_ = os.MkdirAll(pubBin, 0o755)
	_ = os.WriteFile(filepath.Join(pubBin, "mimir_tyr_engine.exe"), []byte("legacy-remote"), 0o755)
	_ = os.WriteFile(filepath.Join(pubBin, platName), []byte("platform-remote-v0.3.0"), 0o755)
	legSum, legSize, _ := FileSHA256(filepath.Join(pubBin, "mimir_tyr_engine.exe"))
	platSum, platSize, _ := FileSHA256(filepath.Join(pubBin, platName))
	m := Manifest{
		SchemaVersion: "1.0", Owner: RequiredOwner, Website: "https://alphaflux.net",
		License: "LICENSE-ENGINES.md", Attribution: AttributionNotice,
		Engines: []EngineEntry{{
			ID: "tyr-core", Version: "0.3.0",
			Binary: "engines/bin/mimir_tyr_engine.exe", SHA256: legSum, SizeBytes: legSize,
			Platforms: map[string]PlatformBinary{
				hostKey: {Binary: platRel, SHA256: platSum, SizeBytes: platSize},
			},
		}},
	}
	writeManifest(t, pub, m)
	srv := httptest.NewServer(http.FileServer(http.Dir(pub)))
	defer srv.Close()

	local := t.TempDir()
	localBin := filepath.Join(local, "engines", "bin")
	_ = os.MkdirAll(localBin, 0o755)
	_ = os.WriteFile(filepath.Join(localBin, "mimir_tyr_engine.exe"), []byte("legacy-local-stale"), 0o755)
	_ = os.WriteFile(filepath.Join(localBin, platName), []byte("platform-old"), 0o755)

	plan, err := CheckForUpdates(context.Background(), local, srv.URL)
	if err != nil {
		t.Fatalf("CheckForUpdates: %v", err)
	}
	if len(plan.Changes) != 1 {
		t.Fatalf("expected 1 change (platform binary), got %+v", plan.Changes)
	}
	if _, err := ApplyUpdates(context.Background(), local, plan); err != nil {
		t.Fatalf("ApplyUpdates: %v", err)
	}
	plat, _ := os.ReadFile(filepath.Join(localBin, platName))
	if string(plat) != "platform-remote-v0.3.0" {
		t.Fatalf("platform binary not updated, got %q", plat)
	}
	leg, _ := os.ReadFile(filepath.Join(localBin, "mimir_tyr_engine.exe"))
	if string(leg) != "legacy-local-stale" {
		t.Fatalf("legacy binary must be untouched on a platform host, got %q", leg)
	}
}

// FIX 1 (audit): the https-only rule was validated on the initial URL only,
// so a redirect could silently downgrade an update download to plain http
// on any host. The download client must refuse non-https redirect targets
// (unless the operator opted into an insecure channel) and cap redirects.
func TestHTTPGet_RefusesRedirectToPlainHTTP(t *testing.T) {
	t.Setenv("MIMIR_UPDATE_ALLOW_INSECURE", "")
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("downgraded-payload"))
	}))
	defer target.Close()
	entry := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL+"/file", http.StatusFound)
	}))
	defer entry.Close()

	body, err := httpGet(context.Background(), entry.URL+"/start", 1<<20)
	if !errors.Is(err, ErrInsecureRedirect) {
		t.Fatalf("expected ErrInsecureRedirect, got err=%v body=%q", err, body)
	}
}

func TestHTTPGet_FollowsRedirectChainWithinCap(t *testing.T) {
	t.Setenv("MIMIR_UPDATE_ALLOW_INSECURE", "1")
	srv := httptest.NewServer(redirectChainHandler(3))
	defer srv.Close()

	body, err := httpGet(context.Background(), srv.URL+"/hop?n=0", 1<<20)
	if err != nil {
		t.Fatalf("redirect chain within the cap should succeed: %v", err)
	}
	if string(body) != "landed" {
		t.Fatalf("unexpected body %q", body)
	}
}

func TestHTTPGet_TooManyRedirects(t *testing.T) {
	t.Setenv("MIMIR_UPDATE_ALLOW_INSECURE", "1")
	srv := httptest.NewServer(redirectChainHandler(7))
	defer srv.Close()

	_, err := httpGet(context.Background(), srv.URL+"/hop?n=0", 1<<20)
	if !errors.Is(err, ErrTooManyRedirects) {
		t.Fatalf("expected ErrTooManyRedirects, got %v", err)
	}
}

// redirectChainHandler redirects /hop?n=k to k+1 while k < hops, then
// serves the final payload.
func redirectChainHandler(hops int) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/hop", func(w http.ResponseWriter, r *http.Request) {
		n, _ := strconv.Atoi(r.URL.Query().Get("n"))
		if n < hops {
			http.Redirect(w, r, fmt.Sprintf("/hop?n=%d", n+1), http.StatusFound)
			return
		}
		_, _ = w.Write([]byte("landed"))
	})
	return mux
}

// The policy unit test covers the https side of the rule, which the live
// httptest servers above cannot exercise (their certificates are not
// trusted by the default transport, so the first hop would fail before
// any redirect is offered).
func TestUpdateRedirectPolicy(t *testing.T) {
	mkReq := func(rawurl string) *http.Request {
		t.Helper()
		req, err := http.NewRequest(http.MethodGet, rawurl, nil)
		if err != nil {
			t.Fatal(err)
		}
		return req
	}
	via := func(n int) []*http.Request {
		out := make([]*http.Request, n)
		for i := range out {
			out[i] = mkReq("https://updates.example.com/start")
		}
		return out
	}

	t.Run("https target allowed without override", func(t *testing.T) {
		t.Setenv("MIMIR_UPDATE_ALLOW_INSECURE", "")
		if err := updateRedirectPolicy(mkReq("https://cdn.example.com/bin"), via(1)); err != nil {
			t.Fatalf("https redirect target should be allowed, got %v", err)
		}
	})
	t.Run("http target refused without override", func(t *testing.T) {
		t.Setenv("MIMIR_UPDATE_ALLOW_INSECURE", "")
		if err := updateRedirectPolicy(mkReq("http://cdn.example.com/bin"), via(1)); !errors.Is(err, ErrInsecureRedirect) {
			t.Fatalf("expected ErrInsecureRedirect, got %v", err)
		}
	})
	t.Run("http target allowed with override", func(t *testing.T) {
		t.Setenv("MIMIR_UPDATE_ALLOW_INSECURE", "1")
		if err := updateRedirectPolicy(mkReq("http://cdn.example.com/bin"), via(1)); err != nil {
			t.Fatalf("override should allow http redirect target, got %v", err)
		}
	})
	t.Run("fifth redirect allowed, sixth refused", func(t *testing.T) {
		t.Setenv("MIMIR_UPDATE_ALLOW_INSECURE", "")
		if err := updateRedirectPolicy(mkReq("https://cdn.example.com/bin"), via(5)); err != nil {
			t.Fatalf("fifth redirect should be allowed, got %v", err)
		}
		if err := updateRedirectPolicy(mkReq("https://cdn.example.com/bin"), via(6)); !errors.Is(err, ErrTooManyRedirects) {
			t.Fatalf("expected ErrTooManyRedirects, got %v", err)
		}
	})
}
