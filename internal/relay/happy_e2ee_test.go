package relay

import (
	"crypto/rand"
	"encoding/base64"
	"io"
	"strings"
	"testing"
	"time"

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

	// 5. Complete handshake pairing (requires the QR pairing token)
	if err := sess.CompletePairing(sess.Nonce, clientPubB64); err != nil {
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

func clientKeyB64(t *testing.T) string {
	t.Helper()
	var clientPriv [32]byte
	if _, err := io.ReadFull(rand.Reader, clientPriv[:]); err != nil {
		t.Fatal(err)
	}
	var clientPub [32]byte
	curve25519.ScalarBaseMult(&clientPub, &clientPriv)
	return base64.StdEncoding.EncodeToString(clientPub[:])
}

func pairedSession(t *testing.T) *Session {
	t.Helper()
	mgr := NewManager()
	sess, err := mgr.CreateSession("127.0.0.1:8420")
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	if err := sess.CompletePairing(sess.Nonce, clientKeyB64(t)); err != nil {
		t.Fatalf("CompletePairing: %v", err)
	}
	return sess
}

// N3: pairing without the QR token (or with a wrong one) must fail, so
// knowing the session ID alone is not enough to hijack a session.
func TestCompletePairingRequiresToken(t *testing.T) {
	mgr := NewManager()
	sess, err := mgr.CreateSession("127.0.0.1:8420")
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	pub := clientKeyB64(t)

	if err := sess.CompletePairing("", pub); err == nil {
		t.Fatal("expected error for empty pairing token")
	}
	if err := sess.CompletePairing("AAAAAAAAAAAAAAAAAAAAAA", pub); err == nil {
		t.Fatal("expected error for wrong pairing token")
	}
	if sess.IsPaired() {
		t.Fatal("session must not be paired after failed attempts")
	}
	if err := sess.CompletePairing(sess.Nonce, pub); err != nil {
		t.Fatalf("expected correct token to pair, got %v", err)
	}
}

// N3: a paired session must reject re-pairing instead of silently
// overwriting the established keys.
func TestCompletePairingRejectsRePairing(t *testing.T) {
	sess := pairedSession(t)
	if err := sess.CompletePairing(sess.Nonce, clientKeyB64(t)); err == nil {
		t.Fatal("expected error when pairing an already paired session")
	}
}

// N3: session IDs carry 128 bits of randomness.
func TestSessionIDIs128Bit(t *testing.T) {
	mgr := NewManager()
	sess, err := mgr.CreateSession("127.0.0.1:8420")
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	raw, err := base64.RawURLEncoding.DecodeString(strings.TrimPrefix(sess.ID, "pair_"))
	if err != nil {
		t.Fatalf("session id is not raw-url base64: %v", err)
	}
	if len(raw) != 16 {
		t.Fatalf("expected 16-byte (128-bit) session id, got %d bytes", len(raw))
	}
}

// N3: unpaired sessions expire after the TTL, for both lookup and pairing.
func TestUnpairedSessionExpires(t *testing.T) {
	mgr := NewManager()
	sess, err := mgr.CreateSession("127.0.0.1:8420")
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	sess.CreatedAt = sess.CreatedAt.Add(-UnpairedSessionTTL - time.Minute)

	if _, ok := mgr.GetSession(sess.ID); ok {
		t.Fatal("expected expired session to be gone from GetSession")
	}
	if err := sess.CompletePairing(sess.Nonce, clientKeyB64(t)); err == nil {
		t.Fatal("expected pairing an expired session to fail")
	}
}

// N3: live sessions are capped so session creation cannot grow unbounded.
func TestSessionCap(t *testing.T) {
	mgr := NewManager()
	for i := 0; i < MaxLiveSessions; i++ {
		if _, err := mgr.CreateSession("127.0.0.1:8420"); err != nil {
			t.Fatalf("CreateSession %d: %v", i, err)
		}
	}
	if _, err := mgr.CreateSession("127.0.0.1:8420"); err == nil {
		t.Fatal("expected error beyond the live session cap")
	}
}

// N5: Broadcast on an unpaired session must error, never send plaintext.
func TestBroadcastUnpairedFails(t *testing.T) {
	mgr := NewManager()
	sess, err := mgr.CreateSession("127.0.0.1:8420")
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	if err := sess.Broadcast("ping", map[string]any{"x": 1}); err == nil {
		t.Fatal("expected broadcast on unpaired session to fail")
	}
}

// N5: a captured frame cannot be replayed, and an older frame is rejected
// once a newer one has been accepted.
func TestDecryptRejectsReplay(t *testing.T) {
	sess := pairedSession(t)

	f1, err := sess.EncryptFrame([]byte("one"))
	if err != nil {
		t.Fatalf("EncryptFrame: %v", err)
	}
	f2, err := sess.EncryptFrame([]byte("two"))
	if err != nil {
		t.Fatalf("EncryptFrame: %v", err)
	}

	if _, err := sess.DecryptFrame(f2); err != nil {
		t.Fatalf("DecryptFrame(f2): %v", err)
	}
	if _, err := sess.DecryptFrame(f1); err == nil {
		t.Fatal("expected older frame to be rejected after newer frame accepted")
	}
	if _, err := sess.DecryptFrame(f2); err == nil {
		t.Fatal("expected replayed frame to be rejected")
	}
}

// FIX 3: a low-order client public key must be rejected. The all-zero point
// forces an all-zero shared secret, which would hand anyone a known session
// key. Pairing with it must fail.
func TestCompletePairingRejectsLowOrderKey(t *testing.T) {
	mgr := NewManager()
	sess, err := mgr.CreateSession("127.0.0.1:8420")
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	zeroPub := base64.StdEncoding.EncodeToString(make([]byte, 32))
	if err := sess.CompletePairing(sess.Nonce, zeroPub); err == nil {
		t.Fatal("expected pairing with the all-zero (low-order) public key to fail")
	}
	if sess.IsPaired() {
		t.Fatal("session must not be paired after a low-order key attempt")
	}
}
