package engines

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// releasePubKeyB64 is the release signing key of record (N7): the public
// half of the ed25519 keypair whose private half is held offline by the
// maintainer and never enters this repo. It is the same value pinned as
// MIMIR_ENGINE_PUBKEY in .env.example, docker-compose.yml,
// docker-compose.prod.yml, and the Dockerfile.
const releasePubKeyB64 = "uJQqyXeyx49sJYWeVttFPX9JnHuyIWCh/fk3CLqvuf4="

// TestShippedManifestSignatureVerifies is the arming gate for N7: the
// manifest shipped in engines/bin must carry a valid signature from the
// release key of record. If this test fails, the manifest was edited (or
// the signing key rotated) without re-signing, and every install with
// the key pinned would refuse the update; re-sign the manifest with the
// maintainer procedure in updater.go before shipping.
func TestShippedManifestSignatureVerifies(t *testing.T) {
	rawPub, err := base64.StdEncoding.DecodeString(releasePubKeyB64)
	if err != nil {
		t.Fatalf("release key of record is not valid base64: %v", err)
	}
	if len(rawPub) != ed25519.PublicKeySize {
		t.Fatalf("release key of record decodes to %d bytes, want %d", len(rawPub), ed25519.PublicKeySize)
	}
	pub := ed25519.PublicKey(rawPub)

	raw, err := os.ReadFile(filepath.Join("..", "..", "engines", "bin", "manifest.json"))
	if err != nil {
		t.Fatalf("read shipped manifest: %v", err)
	}
	var m Manifest
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("parse shipped manifest: %v", err)
	}
	if len(m.Engines) == 0 {
		t.Fatal("shipped manifest lists no engines")
	}
	if m.Signature == "" {
		t.Fatal("shipped manifest carries no signature; N7 is disarmed")
	}
	sig, err := base64.StdEncoding.DecodeString(m.Signature)
	if err != nil {
		t.Fatalf("shipped manifest signature is not valid base64: %v", err)
	}
	canonical, err := ManifestSigningBytes(m)
	if err != nil {
		t.Fatalf("signing bytes: %v", err)
	}
	if !ed25519.Verify(pub, canonical, sig) {
		t.Fatal("shipped manifest signature does not verify against the release key of record; re-sign before shipping")
	}

	// Tamper control: changing any signed content (here, a version
	// string) must invalidate the very same signature, proving the
	// signature actually covers the manifest body.
	tampered := m
	tampered.Engines = append([]EngineEntry(nil), m.Engines...)
	tampered.Engines[0].Version = "9.9.9"
	tamperedBytes, err := ManifestSigningBytes(tampered)
	if err != nil {
		t.Fatalf("tampered signing bytes: %v", err)
	}
	if ed25519.Verify(pub, tamperedBytes, sig) {
		t.Fatal("tampered manifest still verifies; the signature is not covering the manifest content")
	}
}
