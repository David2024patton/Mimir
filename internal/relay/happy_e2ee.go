package relay

import (
	"crypto/cipher"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/skip2/go-qrcode"
	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/curve25519"
)

const (
	// UnpairedSessionTTL bounds how long an unpaired pairing session stays
	// valid. The QR code is a bearer credential until pairing completes, so
	// the window to use it is short.
	UnpairedSessionTTL = 10 * time.Minute
	// MaxLiveSessions caps the number of pairing sessions a Manager holds at
	// once, so the unauthenticated session-creation endpoint cannot be used
	// to grow memory without bound.
	MaxLiveSessions = 64
)

// Session represents an end-to-end encrypted mobile companion pairing session.
// Powered by Alpha Flux LLC Proprietary Engines (https://alphaflux.net).
type Session struct {
	ID        string    `json:"id"`
	Host      string    `json:"host"`
	PublicKey string    `json:"public_key"`
	Nonce     string    `json:"nonce"`
	CreatedAt time.Time `json:"created_at"`

	privateKey [32]byte
	publicKey  [32]byte
	sharedKey  [32]byte
	paired     bool
	aead       cipher.AEAD
	sendSeq    uint64 // last counter used by EncryptFrame (N5 replay protection)
	recvSeq    uint64 // highest counter accepted by DecryptFrame (N5 replay protection)
	mu         sync.RWMutex
	listeners  map[chan []byte]struct{}
}

// PairURI generates the mimir:// URL format for QR-code pairing.
func (s *Session) PairURI() string {
	return fmt.Sprintf("mimir://pair?id=%s&host=%s&pubkey=%s&nonce=%s",
		s.ID, s.Host, s.PublicKey, s.Nonce)
}

// GenerateQRCodeASCII generates a human-readable ASCII QR code for terminal pairing.
func (s *Session) GenerateQRCodeASCII() (string, error) {
	qr, err := qrcode.New(s.PairURI(), qrcode.Medium)
	if err != nil {
		return "", fmt.Errorf("failed to generate qr code: %w", err)
	}
	return qr.ToSmallString(false), nil
}

// Manager coordinates mobile pairing sessions and encrypted relays.
type Manager struct {
	sessions map[string]*Session
	mu       sync.RWMutex
}

// NewManager creates a thread-safe relay session manager.
func NewManager() *Manager {
	return &Manager{
		sessions: make(map[string]*Session),
	}
}

// CreateSession generates a new X25519 ephemeral keypair and session tokens.
// Unpaired sessions older than UnpairedSessionTTL are purged first, and
// creation fails once MaxLiveSessions live sessions exist.
func (m *Manager) CreateSession(host string) (*Session, error) {
	var privKey [32]byte
	if _, err := io.ReadFull(rand.Reader, privKey[:]); err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	var pubKey [32]byte
	curve25519.ScalarBaseMult(&pubKey, &privKey)

	nonceBytes := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, nonceBytes); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// 128-bit session ID: the ID alone must not be guessable, because it is
	// the only handle needed to reach the pairing endpoints.
	sessIDBytes := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, sessIDBytes); err != nil {
		return nil, fmt.Errorf("failed to generate session id: %w", err)
	}

	sessID := fmt.Sprintf("pair_%s", base64.RawURLEncoding.EncodeToString(sessIDBytes))
	pubKeyB64 := base64.StdEncoding.EncodeToString(pubKey[:])
	nonceB64 := base64.RawURLEncoding.EncodeToString(nonceBytes)

	sess := &Session{
		ID:         sessID,
		Host:       host,
		PublicKey:  pubKeyB64,
		Nonce:      nonceB64,
		CreatedAt:  time.Now(),
		privateKey: privKey,
		publicKey:  pubKey,
		listeners:  make(map[chan []byte]struct{}),
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.purgeExpiredLocked(time.Now())
	if len(m.sessions) >= MaxLiveSessions {
		return nil, fmt.Errorf("relay: too many live pairing sessions (max %d), try again later", MaxLiveSessions)
	}
	m.sessions[sessID] = sess

	return sess, nil
}

// purgeExpiredLocked deletes unpaired sessions past UnpairedSessionTTL.
// Callers must hold m.mu.
func (m *Manager) purgeExpiredLocked(now time.Time) {
	for id, s := range m.sessions {
		if !s.IsPaired() && now.Sub(s.CreatedAt) > UnpairedSessionTTL {
			delete(m.sessions, id)
		}
	}
}

// sessionExpired reports whether an unpaired session has outlived its TTL.
func (s *Session) sessionExpired(now time.Time) bool {
	return !s.IsPaired() && now.Sub(s.CreatedAt) > UnpairedSessionTTL
}

// GetSession returns a session by ID if found. Unpaired sessions past
// UnpairedSessionTTL are treated as gone and removed lazily.
func (m *Manager) GetSession(id string) (*Session, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.sessions[id]
	if !ok {
		return nil, false
	}
	if s.sessionExpired(time.Now()) {
		delete(m.sessions, id)
		return nil, false
	}
	return s, true
}

// CompletePairing performs the X25519 Diffie-Hellman exchange with the
// client public key. The caller must present the pairing token carried in
// the QR code (the session Nonce): without it, anyone who learns the
// session ID could pair their own key and hijack the session. The token is
// compared in constant time. Pairing is one-shot: an already paired
// session rejects further attempts instead of silently replacing its keys.
func (s *Session) CompletePairing(pairingToken, clientPubKeyB64 string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.paired {
		return errors.New("session is already paired")
	}
	if time.Since(s.CreatedAt) > UnpairedSessionTTL {
		return errors.New("pairing session has expired")
	}

	presented, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(pairingToken))
	if err != nil {
		return errors.New("invalid pairing token")
	}
	expected, err := base64.RawURLEncoding.DecodeString(s.Nonce)
	if err != nil {
		return errors.New("invalid pairing token")
	}
	if subtle.ConstantTimeCompare(presented, expected) != 1 {
		return errors.New("invalid pairing token")
	}

	clientPubBytes, err := base64.StdEncoding.DecodeString(clientPubKeyB64)
	if err != nil || len(clientPubBytes) != 32 {
		return errors.New("invalid client public key: must be 32 bytes base64")
	}

	var clientPubKey [32]byte
	copy(clientPubKey[:], clientPubBytes)

	var shared [32]byte
	curve25519.ScalarMult(&shared, &s.privateKey, &clientPubKey)

	aead, err := chacha20poly1305.New(shared[:])
	if err != nil {
		return fmt.Errorf("failed to initialize chacha20poly1305 aead: %w", err)
	}

	s.sharedKey = shared
	s.aead = aead
	s.paired = true
	return nil
}

// IsPaired checks if the session has established shared secret.
func (s *Session) IsPaired() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.paired
}

// EncryptFrame seals arbitrary data using ChaCha20-Poly1305 AEAD. The AEAD
// nonce is derived from a per-session monotonic counter (stored in the last
// 8 bytes of the nonce, big endian), never random, so a nonce can never
// repeat under the session key. The wire format stays nonce || ciphertext.
func (s *Session) EncryptFrame(plaintext []byte) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.encryptFrameLocked(plaintext)
}

// encryptFrameLocked seals plaintext. Callers must hold s.mu.
func (s *Session) encryptFrameLocked(plaintext []byte) ([]byte, error) {
	if !s.paired || s.aead == nil {
		return nil, errors.New("session is not paired yet")
	}

	s.sendSeq++
	nonce := make([]byte, s.aead.NonceSize())
	binary.BigEndian.PutUint64(nonce[len(nonce)-8:], s.sendSeq)

	ciphertext := s.aead.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// DecryptFrame decrypts and verifies a sealed ciphertext frame. Frames
// whose counter is at or below the highest counter already accepted are
// rejected, so a captured frame cannot be replayed.
func (s *Session) DecryptFrame(ciphertext []byte) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.paired || s.aead == nil {
		return nil, errors.New("session is not paired yet")
	}

	nonceSize := s.aead.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short: missing nonce")
	}

	nonce := ciphertext[:nonceSize]
	sealed := ciphertext[nonceSize:]

	counter := binary.BigEndian.Uint64(nonce[len(nonce)-8:])
	if counter <= s.recvSeq {
		return nil, errors.New("replayed or out-of-order frame rejected")
	}

	plaintext, err := s.aead.Open(nil, nonce, sealed, nil)
	if err != nil {
		return nil, fmt.Errorf("authentication error: %w", err)
	}
	s.recvSeq = counter
	return plaintext, nil
}

// Broadcast sends an encrypted JSON message to all connected SSE/WS mobile
// listeners. There is deliberately no plaintext path: broadcasting on an
// unpaired session is an error, never a cleartext send.
func (s *Session) Broadcast(msgType string, payload any) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.paired || s.aead == nil {
		return errors.New("session is not paired; refusing to broadcast without encryption")
	}

	raw, err := json.Marshal(map[string]any{
		"type":      msgType,
		"payload":   payload,
		"timestamp": time.Now().Unix(),
	})
	if err != nil {
		return err
	}

	data, err := s.encryptFrameLocked(raw)
	if err != nil {
		return err
	}

	for ch := range s.listeners {
		select {
		case ch <- data:
		default:
		}
	}
	return nil
}

// RegisterListener adds a listener channel.
func (s *Session) RegisterListener(ch chan []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.listeners[ch] = struct{}{}
}

// UnregisterListener removes a listener channel.
func (s *Session) UnregisterListener(ch chan []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.listeners, ch)
	close(ch)
}

// StreamHandler exposes Server-Sent Events (SSE) for the mobile app stream.
// The stream carries encrypted frames for the paired phone only, so it
// refuses unpaired sessions and sends no wildcard CORS header.
func (s *Session) StreamHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !s.IsPaired() {
			http.Error(w, "session is not paired", http.StatusForbidden)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
			return
		}

		ch := make(chan []byte, 32)
		s.RegisterListener(ch)
		defer s.UnregisterListener(ch)

		// Initial connection frame
		fmt.Fprintf(w, "event: connected\ndata: %s\n\n", s.ID)
		flusher.Flush()

		ctx := r.Context()
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
				b64 := base64.StdEncoding.EncodeToString(msg)
				fmt.Fprintf(w, "data: %s\n\n", b64)
				flusher.Flush()
			}
		}
	}
}
