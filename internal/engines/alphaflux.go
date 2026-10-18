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

// EngineEntry describes one pre-compiled proprietary engine binary in `engines/bin/`.
type EngineEntry struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Version     string `json:"version"`
	Branch      string `json:"source_branch"`
	Binary      string `json:"binary"`
	SHA256      string `json:"sha256"`
	SizeBytes   int64  `json:"size_bytes"`
	Stack       string `json:"stack"`
	Description string `json:"description"`
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
// exists in `engines/bin/` and matches its recorded SHA-256 digest.
func VerifyBinaries(rootDir string, m Manifest) error {
	if err := VerifyAttribution(rootDir, m); err != nil {
		return err
	}
	for _, e := range m.Engines {
		binPath := filepath.Join(rootDir, filepath.FromSlash(e.Binary))
		sum, _, err := FileSHA256(binPath)
		if err != nil {
			return fmt.Errorf("%w: %s (%v)", ErrBinaryIntegrity, e.Binary, err)
		}
		if !strings.EqualFold(sum, e.SHA256) {
			return fmt.Errorf("%w: %s sha256=%s want=%s", ErrBinaryIntegrity, e.Binary, sum, e.SHA256)
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
	binPath := filepath.Join(rootDir, filepath.FromSlash(target.Binary))
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
	b.WriteString(fmt.Sprintf("- **License File**: `%s`\n\n", m.License))
	for _, e := range m.Engines {
		binPath := filepath.Join(rootDir, filepath.FromSlash(e.Binary))
		sum, sz, fErr := FileSHA256(binPath)
		status := "VERIFIED (pre-compiled binary ready)"
		if fErr != nil {
			status = "MISSING"
		} else if !strings.EqualFold(sum, e.SHA256) {
			status = "SHA256_MISMATCH"
		}
		b.WriteString(fmt.Sprintf("* **%s (`%s`)** -- Stack: `%s` | Source Branch: `%s`\n", e.Name, e.ID, e.Stack, e.Branch))
		b.WriteString(fmt.Sprintf("  - Binary: `%s` (%d KB | sha256: `%s...`) -> **%s**\n",
			e.Binary, sz/1024, truncateHex(e.SHA256, 12), status))
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
