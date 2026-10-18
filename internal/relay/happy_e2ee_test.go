package relay

import (
	"crypto/rand"
	"encoding/base64"
	"io"
	"strings"
	"testing"

	"golang.org/x/crypto/curve25519"
)

func TestHappyE2EESessionFlow(t *testing.T) {
	mgr := NewManager()

	// 1. Host creates pairing session
	sess, err := mgr.CreateSession("192.168.0.83:8080")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}
	if sess.ID == "" || sess.PublicKey == "" || sess.Nonce == "" {
		t.Fatalf("invalid session properties: %+v", sess)
	}

	// 2. Verify QR URI format
	uri := sess.PairURI()
	if !strings.HasPrefix(uri, "mimir://pair?") {
		t.Fatalf("unexpected pair URI: %s", uri)
	}

	// 3. Verify terminal ASCII QR code generation
	asciiQR, err := sess.GenerateQRCodeASCII()
	if err != nil {
		t.Fatalf("failed to generate ascii qr: %v", err)
	}
	if len(asciiQR) == 0 {
		t.Fatal("empty ascii qr output")
	}

	// 4. Simulate client (phone) generating its own ephemeral X25519 keypair
	var clientPriv [32]byte
	if _, err := io.ReadFull(rand.Reader, clientPriv[:]); err != nil {
		t.Fatal(err)
	}
	var clientPub [32]byte
	curve25519.ScalarBaseMult(&clientPub, &clientPriv)
	clientPubB64 := base64.StdEncoding.EncodeToString(clientPub[:])

	// 5. Complete handshake pairing
	if err := sess.CompletePairing(clientPubB64); err != nil {
		t.Fatalf("failed to complete pairing: %v", err)
	}
	if !sess.IsPaired() {
		t.Fatal("session expected to be paired")
	}

	// 6. Test Encrypt & Decrypt Frame
	msg := []byte("ApprovalRequest: Thor wants to write to internal/relay/relay.go")
	sealed, err := sess.EncryptFrame(msg)
	if err != nil {
		t.Fatalf("failed to encrypt frame: %v", err)
	}

	plaintext, err := sess.DecryptFrame(sealed)
	if err != nil {
		t.Fatalf("failed to decrypt frame: %v", err)
	}
	if string(plaintext) != string(msg) {
		t.Fatalf("plaintext mismatch: expected %q, got %q", string(msg), string(plaintext))
	}
}
