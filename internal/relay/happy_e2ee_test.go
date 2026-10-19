package relay

import (
	"bytes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/chacha20poly1305"
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

	// 6. Test Encrypt & Decrypt Frame in both directions. The host side
	// (this package) and the phone side use the same shared key but
	// disjoint nonce partitions, so each direction is exercised through
	// its own endpoint: the host seals and the phone opens, then the
	// phone seals and the host opens.
	msg := []byte("ApprovalRequest: Thor wants to write to internal/relay/relay.go")
	sealed, err := sess.EncryptFrame(msg)
	if err != nil {
		t.Fatalf("failed to encrypt frame: %v", err)
	}

	phone := peerAEAD(t, sess, clientPriv)
	nonceSize := phone.NonceSize()
	plaintext, err := phone.Open(nil, sealed[:nonceSize], sealed[nonceSize:], nil)
	if err != nil {
		t.Fatalf("phone failed to decrypt host frame: %v", err)
	}
	if string(plaintext) != string(msg) {
		t.Fatalf("plaintext mismatch: expected %q, got %q", string(msg), string(plaintext))
	}

	reply := peerSeal(phone, 1, []byte("Approved"))
	got, err := sess.DecryptFrame(reply)
	if err != nil {
		t.Fatalf("host failed to decrypt phone frame: %v", err)
	}
	if string(got) != "Approved" {
		t.Fatalf("reply mismatch: got %q", string(got))
	}
}

func clientKeypair(t *testing.T) (priv [32]byte, pubB64 string) {
	t.Helper()
	if _, err := io.ReadFull(rand.Reader, priv[:]); err != nil {
		t.Fatal(err)
	}
	var pub [32]byte
	curve25519.ScalarBaseMult(&pub, &priv)
	return priv, base64.StdEncoding.EncodeToString(pub[:])
}

func clientKeyB64(t *testing.T) string {
	t.Helper()
	_, pubB64 := clientKeypair(t)
	return pubB64
}

// peerAEAD builds the phone side's AEAD for a paired session: the phone
// derives the same shared secret from its private key and the host public
// key carried in the pair URI.
func peerAEAD(t *testing.T, sess *Session, clientPriv [32]byte) cipher.AEAD {
	t.Helper()
	hostPub, err := base64.StdEncoding.DecodeString(sess.PublicKey)
	if err != nil || len(hostPub) != 32 {
		t.Fatalf("bad host public key: %v", err)
	}
	shared, err := curve25519.X25519(clientPriv[:], hostPub)
	if err != nil {
		t.Fatalf("phone key agreement: %v", err)
	}
	aead, err := chacha20poly1305.New(shared)
	if err != nil {
		t.Fatalf("phone aead: %v", err)
	}
	return aead
}

// peerSeal seals a phone-to-host frame the way a conforming phone client
// must: the counter lives in the last 8 bytes of the nonce with the
// direction bit CLEAR (the phone partition).
func peerSeal(aead cipher.AEAD, counter uint64, plaintext []byte) []byte {
	nonce := make([]byte, aead.NonceSize())
	binary.BigEndian.PutUint64(nonce[len(nonce)-8:], counter)
	return aead.Seal(nonce, nonce, plaintext, nil)
}

func pairedSessionWithClient(t *testing.T) (*Session, [32]byte) {
	t.Helper()
	mgr := NewManager()
	sess, err := mgr.CreateSession("127.0.0.1:8420")
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	priv, pubB64 := clientKeypair(t)
	if err := sess.CompletePairing(sess.Nonce, pubB64); err != nil {
		t.Fatalf("CompletePairing: %v", err)
	}
	return sess, priv
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

// N5: a captured phone frame cannot be replayed, and an older frame is
// rejected once a newer one has been accepted. Frames arrive from the
// phone partition (direction bit clear).
func TestDecryptRejectsReplay(t *testing.T) {
	sess, clientPriv := pairedSessionWithClient(t)
	phone := peerAEAD(t, sess, clientPriv)

	f1 := peerSeal(phone, 1, []byte("one"))
	f2 := peerSeal(phone, 2, []byte("two"))

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

// FIX 2: the two directions share one AEAD key, so the counter field of
// the nonce is partitioned by direction. Host frames (encrypted by this
// package) carry the top bit of the 8-byte counter SET; phone frames must
// carry it CLEAR. Without the partition, host frame 1 and phone frame 1
// would use the identical (key, nonce) pair.
func TestNonceDirectionPartitions(t *testing.T) {
	sess, clientPriv := pairedSessionWithClient(t)
	phone := peerAEAD(t, sess, clientPriv)
	nonceSize := phone.NonceSize()

	hostFrame, err := sess.EncryptFrame([]byte("host one"))
	if err != nil {
		t.Fatalf("EncryptFrame: %v", err)
	}
	hostCounter := binary.BigEndian.Uint64(hostFrame[nonceSize-8 : nonceSize])
	if hostCounter>>63 != 1 {
		t.Fatalf("host frame counter %#x is missing the direction bit", hostCounter)
	}

	// A phone frame with the same numeric counter (1) must still have
	// different nonce bytes, because it lives in the other partition.
	phoneFrame := peerSeal(phone, 1, []byte("phone one"))
	if bytes.Equal(phoneFrame[:nonceSize], hostFrame[:nonceSize]) {
		t.Fatal("host and phone frames with the same counter share a nonce")
	}
	phoneCounter := binary.BigEndian.Uint64(phoneFrame[nonceSize-8 : nonceSize])
	if phoneCounter>>63 != 0 {
		t.Fatalf("test bug: phone frame counter %#x has the direction bit set", phoneCounter)
	}

	// Both directions decrypt at their own endpoint.
	got, err := sess.DecryptFrame(phoneFrame)
	if err != nil {
		t.Fatalf("host DecryptFrame(phone frame): %v", err)
	}
	if string(got) != "phone one" {
		t.Fatalf("phone frame mismatch: got %q", string(got))
	}
	opened, err := phone.Open(nil, hostFrame[:nonceSize], hostFrame[nonceSize:], nil)
	if err != nil {
		t.Fatalf("phone Open(host frame): %v", err)
	}
	if string(opened) != "host one" {
		t.Fatalf("host frame mismatch: got %q", string(opened))
	}

	// A host frame reflected back at the host sits in the wrong partition
	// and must be rejected, not decrypted.
	if _, err := sess.DecryptFrame(hostFrame); err == nil {
		t.Fatal("expected host frame to be rejected by host DecryptFrame (wrong partition)")
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

// FIX 1: paired sessions must expire too. Before this fix they lived in
// the Manager forever, so 64 pairings permanently exhausted the session
// cap (a pairing DoS). Expiry is measured from the pairing time with the
// Manager's PairedSessionTTL.
func TestPairedSessionExpires(t *testing.T) {
	mgr := NewManager()
	mgr.PairedSessionTTL = time.Minute
	sess, err := mgr.CreateSession("127.0.0.1:8420")
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	if err := sess.CompletePairing(sess.Nonce, clientKeyB64(t)); err != nil {
		t.Fatalf("CompletePairing: %v", err)
	}
	if _, ok := mgr.GetSession(sess.ID); !ok {
		t.Fatal("a freshly paired session must be live")
	}

	sess.mu.Lock()
	sess.pairedAt = time.Now().Add(-2 * time.Minute)
	sess.mu.Unlock()

	if _, ok := mgr.GetSession(sess.ID); ok {
		t.Fatal("expected paired session past its TTL to be gone from GetSession")
	}
}

// FIX 1: the CreateSession cap must count live sessions only. Once paired
// sessions expire, capacity frees up and creation succeeds again.
func TestPairedExpiryFreesCapacity(t *testing.T) {
	mgr := NewManager()
	mgr.PairedSessionTTL = time.Minute
	live := make([]*Session, 0, MaxLiveSessions)
	for i := 0; i < MaxLiveSessions; i++ {
		sess, err := mgr.CreateSession("127.0.0.1:8420")
		if err != nil {
			t.Fatalf("CreateSession %d: %v", i, err)
		}
		if err := sess.CompletePairing(sess.Nonce, clientKeyB64(t)); err != nil {
			t.Fatalf("CompletePairing %d: %v", i, err)
		}
		live = append(live, sess)
	}
	if _, err := mgr.CreateSession("127.0.0.1:8420"); err == nil {
		t.Fatal("expected error beyond the live session cap")
	}

	for _, sess := range live {
		sess.mu.Lock()
		sess.pairedAt = time.Now().Add(-2 * time.Minute)
		sess.mu.Unlock()
	}
	if _, err := mgr.CreateSession("127.0.0.1:8420"); err != nil {
		t.Fatalf("expected creation to succeed after paired sessions expired, got %v", err)
	}
}

// FIX 1: RemoveSession ends a pairing on demand. It reports whether a
// session existed; unknown IDs are a no-op reporting false.
func TestRemoveSession(t *testing.T) {
	mgr := NewManager()
	sess, err := mgr.CreateSession("127.0.0.1:8420")
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	if err := sess.CompletePairing(sess.Nonce, clientKeyB64(t)); err != nil {
		t.Fatalf("CompletePairing: %v", err)
	}

	if !mgr.RemoveSession(sess.ID) {
		t.Fatal("RemoveSession must report true for a live session")
	}
	if _, ok := mgr.GetSession(sess.ID); ok {
		t.Fatal("removed session must not be returned by GetSession")
	}
	if sess.IsPaired() {
		t.Fatal("removed session must be torn down (no longer paired)")
	}
	if _, err := sess.EncryptFrame([]byte("x")); err == nil {
		t.Fatal("removed session must not encrypt frames")
	}
	if mgr.RemoveSession(sess.ID) {
		t.Fatal("second RemoveSession must report false")
	}
	if mgr.RemoveSession("pair_no-such-session") {
		t.Fatal("RemoveSession of an unknown ID must report false")
	}
}

// FIX 5: trust-boundary events are logged via slog: session creation,
// pairing success, pairing failure (reason only), expiry, and removal.
// The pairing token must never appear in the log output.
func TestRelayLifecycleLogging(t *testing.T) {
	var buf bytes.Buffer
	old := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, nil)))
	defer slog.SetDefault(old)

	mgr := NewManager()
	sess, err := mgr.CreateSession("127.0.0.1:8420")
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	if err := sess.CompletePairing("AAAAAAAAAAAAAAAAAAAAAA", clientKeyB64(t)); err == nil {
		t.Fatal("expected wrong-token pairing to fail")
	}
	if err := sess.CompletePairing(sess.Nonce, clientKeyB64(t)); err != nil {
		t.Fatalf("CompletePairing: %v", err)
	}
	if !mgr.RemoveSession(sess.ID) {
		t.Fatal("RemoveSession must succeed")
	}

	// An expired session purged during a later CreateSession is logged too.
	stale, err := mgr.CreateSession("127.0.0.1:8420")
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	stale.CreatedAt = stale.CreatedAt.Add(-UnpairedSessionTTL - time.Minute)
	if _, err := mgr.CreateSession("127.0.0.1:8420"); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	out := buf.String()
	for _, want := range []string{
		"relay session created",
		"relay pairing failed",
		"relay session paired",
		"relay session removed",
		"relay session expired",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("log output missing %q; got:\n%s", want, out)
		}
	}
	if strings.Contains(out, sess.Nonce) {
		t.Error("log output must never contain the pairing token")
	}
}
