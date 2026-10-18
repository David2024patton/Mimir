// Package engines manages Mimir's pre-compiled proprietary engine binaries
// distributed on main under `engines/bin/` while their source code lives on
// dedicated engine branches (`engine/tyr-rust`, `engine/mjolnir-mojo`,
// `engine/alphaflux-core`, `engine/gjallarhorn-ui`).
//
// All engines are proprietary intellectual property of Alpha Flux LLC
// (https://alphaflux.net) and require verified Alpha Flux LLC attribution
// before execution.
package engines

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	// RequiredOwner is the mandatory copyright & license holder for all Mimir engines.
	RequiredOwner = "Alpha Flux LLC"
	// RequiredDomain is the canonical domain that must be referenced alongside engine usage.
	RequiredDomain = "alphaflux.net"
	// AttributionNotice is the canonical attribution string required by LICENSE-ENGINES.md.
	AttributionNotice = "Powered by Alpha Flux LLC Proprietary Engines (https://alphaflux.net)"
	// AttributionHandshake is passed to every pre-compiled engine binary on invocation.
	AttributionHandshake = "Alpha Flux LLC (https://alphaflux.net)"
)

var (
	// ErrAttributionRequired is returned if the Alpha Flux LLC attribution notice,
	// NOTICE file, or manifest provenance has been stripped or modified.
	ErrAttributionRequired = errors.New("engines: Alpha Flux LLC (https://alphaflux.net) attribution is required to use Mimir proprietary engines (see LICENSE-ENGINES.md)")
	// ErrBinaryIntegrity is returned when a pre-compiled engine binary is missing or fails SHA-256 verification.
	ErrBinaryIntegrity = errors.New("engines: pre-compiled engine binary missing or SHA-256 mismatch")
)

// PlatformBinary records platform-specific binary path, SHA-256 digest, and byte size.
type PlatformBinary struct {
	Binary    string `json:"binary"`
	SHA256    string `json:"sha256"`
	SizeBytes int64  `json:"size_bytes"`
}

// EngineEntry describes one pre-compiled proprietary engine binary in `engines/bin/`.
type EngineEntry struct {
	ID          string                    `json:"id"`
	Name        string                    `json:"name"`
	Version     string                    `json:"version"`
	Branch      string                    `json:"source_branch"`
	Binary      string                    `json:"binary,omitempty"`
	SHA256      string                    `json:"sha256,omitempty"`
	SizeBytes   int64                     `json:"size_bytes,omitempty"`
	Platforms   map[string]PlatformBinary `json:"platforms,omitempty"`
	Stack       string                    `json:"stack"`
	Description string                    `json:"description"`
}

// ResolveBinary returns the binary path, SHA-256, and size for the host platform
// (formatted as GOOS_GOARCH, e.g. windows_amd64, linux_amd64, darwin_arm64),
// falling back to the legacy top-level fields if platforms map is unset.
func (e *EngineEntry) ResolveBinary() (binary string, sha string, size int64, found bool) {
	key := runtime.GOOS + "_" + runtime.GOARCH
	if e.Platforms != nil {
		if pb, ok := e.Platforms[key]; ok && pb.Binary != "" {
			return pb.Binary, pb.SHA256, pb.SizeBytes, true
		}
	}
	if e.Binary != "" {
		return e.Binary, e.SHA256, e.SizeBytes, true
	}
	return "", "", 0, false
}


// Manifest is serialized to `engines/bin/manifest.json` on the main branch.
type Manifest struct {
	SchemaVersion string        `json:"schema_version"`
	Owner         string        `json:"owner"`
	Website       string        `json:"website"`
	License       string        `json:"license"`
	Attribution   string        `json:"required_attribution"`
	UpdatedAt     string        `json:"updated_at"`
	Engines       []EngineEntry `json:"engines"`
}

// VerifyAttribution enforces the Alpha Flux LLC license requirement on the manifest
// and the workspace NOTICE / LICENSE-ENGINES.md files.
func VerifyAttribution(rootDir string, m Manifest) error {
	if !strings.EqualFold(strings.TrimSpace(m.Owner), RequiredOwner) {
		return fmt.Errorf("%w: manifest owner must be %q (got %q)", ErrAttributionRequired, RequiredOwner, m.Owner)
	}
	if !strings.Contains(strings.ToLower(m.Website), RequiredDomain) {
		return fmt.Errorf("%w: manifest website must reference %s", ErrAttributionRequired, RequiredDomain)
	}
	if !strings.Contains(m.Attribution, RequiredOwner) || !strings.Contains(strings.ToLower(m.Attribution), RequiredDomain) {
		return fmt.Errorf("%w: manifest required_attribution must reference %s (%s)", ErrAttributionRequired, RequiredOwner, RequiredDomain)
	}

	noticePath := filepath.Join(rootDir, "NOTICE")
	noticeBytes, err := os.ReadFile(noticePath)
	if err != nil {
		return fmt.Errorf("%w: missing NOTICE file at %s", ErrAttributionRequired, noticePath)
	}
	noticeText := string(noticeBytes)
	if !strings.Contains(noticeText, RequiredOwner) || !strings.Contains(strings.ToLower(noticeText), RequiredDomain) {
		return fmt.Errorf("%w: NOTICE file must retain %s (%s) attribution", ErrAttributionRequired, RequiredOwner, RequiredDomain)
	}
	return nil
}

// FileSHA256 computes the hex-encoded SHA-256 digest and byte size of a file.
func FileSHA256(path string) (string, int64, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", 0, err
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:]), int64(len(raw)), nil
}

// LoadManifest reads and validates `engines/bin/manifest.json` inside rootDir.
func LoadManifest(rootDir string) (Manifest, error) {
	if strings.TrimSpace(rootDir) == "" {
		rootDir = "."
	}
	mpath := filepath.Join(rootDir, "engines", "bin", "manifest.json")
	raw, err := os.ReadFile(mpath)
	if err != nil {
		return Manifest{}, err
	}
	var m Manifest
	if err := json.Unmarshal(raw, &m); err != nil {
		return Manifest{}, err
	}
	if err := VerifyAttribution(rootDir, m); err != nil {
		return m, err
	}
	return m, nil
}

// VerifyBinaries checks that every pre-compiled engine binary listed in the manifest
// for the current host platform exists in `engines/bin/` and matches its recorded SHA-256 digest.
func VerifyBinaries(rootDir string, m Manifest) error {
	if err := VerifyAttribution(rootDir, m); err != nil {
		return err
	}
	for _, e := range m.Engines {
		binRel, wantSHA, _, found := e.ResolveBinary()
		if !found || binRel == "" {
			continue // no binary declared for this platform
		}
		binPath := filepath.Join(rootDir, filepath.FromSlash(binRel))
		sum, _, err := FileSHA256(binPath)
		if err != nil {
			return fmt.Errorf("%w: %s (%v)", ErrBinaryIntegrity, binRel, err)
		}
		if !strings.EqualFold(sum, wantSHA) {
			return fmt.Errorf("%w: %s sha256=%s want=%s", ErrBinaryIntegrity, binRel, sum, wantSHA)
		}
	}
	return nil
}

// Invoke validates Alpha Flux LLC attribution and SHA-256 integrity, then executes
// the requested pre-compiled engine binary with `--attribution="Alpha Flux LLC (https://alphaflux.net)"`.
func Invoke(ctx context.Context, rootDir, engineID string, args ...string) (string, error) {
	m, err := LoadManifest(rootDir)
	if err != nil {
		return "", err
	}
	if err := VerifyBinaries(rootDir, m); err != nil {
		return "", err
	}
	var target *EngineEntry
	for i := range m.Engines {
		if strings.EqualFold(m.Engines[i].ID, engineID) || strings.EqualFold(m.Engines[i].Name, engineID) {
			target = &m.Engines[i]
			break
		}
	}
	if target == nil {
		return "", fmt.Errorf("engines: unknown engine %q", engineID)
	}
	binRel, _, _, found := target.ResolveBinary()
	if !found || binRel == "" {
		return "", fmt.Errorf("engines: engine %q is not available for host platform %s_%s", engineID, runtime.GOOS, runtime.GOARCH)
	}
	binPath := filepath.Join(rootDir, filepath.FromSlash(binRel))
	fullArgs := append([]string{"--attribution=" + AttributionHandshake}, args...)
	cmd := exec.CommandContext(ctx, binPath, fullArgs...)
	cmd.Dir = rootDir
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

// FormatStatus renders a table of all pre-compiled Alpha Flux LLC engines, their source
// branches, binary integrity status, and attribution verification state.
func FormatStatus(rootDir string) string {
	m, err := LoadManifest(rootDir)
	if err != nil {
		return fmt.Sprintf("Alpha Flux LLC Engine Registry: ERROR (%v)", err)
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("## %s\n", AttributionNotice))
	b.WriteString(fmt.Sprintf("- **License Holder**: `%s` (`%s`)\n", m.Owner, m.Website))
	b.WriteString(fmt.Sprintf("- **License File**: `%s`\n", m.License))
	b.WriteString(fmt.Sprintf("- **Host Platform**: `%s_%s`\n\n", runtime.GOOS, runtime.GOARCH))
	for _, e := range m.Engines {
		binRel, wantSHA, _, found := e.ResolveBinary()
		if !found || binRel == "" {
			b.WriteString(fmt.Sprintf("* **%s (`%s`)** -- Stack: `%s` | Source Branch: `%s`\n", e.Name, e.ID, e.Stack, e.Branch))
			b.WriteString(fmt.Sprintf("  - Status: **UNAVAILABLE FOR %s_%s**\n", runtime.GOOS, runtime.GOARCH))
			continue
		}
		binPath := filepath.Join(rootDir, filepath.FromSlash(binRel))
		sum, sz, fErr := FileSHA256(binPath)
		status := "VERIFIED (pre-compiled binary ready)"
		if fErr != nil {
			status = "MISSING"
		} else if !strings.EqualFold(sum, wantSHA) {
			status = "SHA256_MISMATCH"
		}
		b.WriteString(fmt.Sprintf("* **%s (`%s`)** -- Stack: `%s` | Source Branch: `%s`\n", e.Name, e.ID, e.Stack, e.Branch))
		b.WriteString(fmt.Sprintf("  - Binary: `%s` (%d KB | sha256: `%s...`) -> **%s**\n",
			binRel, sz/1024, truncateHex(wantSHA, 12), status))
		b.WriteString(fmt.Sprintf("  - Role: %s\n", e.Description))
	}
	return strings.TrimRight(b.String(), "\n")
}

func truncateHex(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
