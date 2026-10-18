package relay

import (
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/skip2/go-qrcode"
	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/curve25519"
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

	sessIDBytes := make([]byte, 8)
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
	m.sessions[sessID] = sess
	m.mu.Unlock()

	return sess, nil
}

// GetSession returns a session by ID if found.
func (m *Manager) GetSession(id string) (*Session, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.sessions[id]
	return s, ok
}

// CompletePairing performs the X25519 Diffie-Hellman exchange with client public key.
func (s *Session) CompletePairing(clientPubKeyB64 string) error {
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

	s.mu.Lock()
	defer s.mu.Unlock()
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

// EncryptFrame seals arbitrary data using ChaCha20-Poly1305 AEAD.
func (s *Session) EncryptFrame(plaintext []byte) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !s.paired || s.aead == nil {
		return nil, errors.New("session is not paired yet")
	}

	nonce := make([]byte, s.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := s.aead.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// DecryptFrame decrypts and verifies a sealed ciphertext frame.
func (s *Session) DecryptFrame(ciphertext []byte) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !s.paired || s.aead == nil {
		return nil, errors.New("session is not paired yet")
	}

	nonceSize := s.aead.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short: missing nonce")
	}

	nonce := ciphertext[:nonceSize]
	sealed := ciphertext[nonceSize:]

	plaintext, err := s.aead.Open(nil, nonce, sealed, nil)
	if err != nil {
		return nil, fmt.Errorf("authentication error: %w", err)
	}
	return plaintext, nil
}

// Broadcast sends an encrypted JSON message to all connected SSE/WS mobile listeners.
func (s *Session) Broadcast(msgType string, payload any) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	raw, err := json.Marshal(map[string]any{
		"type":      msgType,
		"payload":   payload,
		"timestamp": time.Now().Unix(),
	})
	if err != nil {
		return err
	}

	var data []byte
	if s.paired && s.aead != nil {
		enc, err := s.EncryptFrame(raw)
		if err != nil {
			return err
		}
		data = enc
	} else {
		data = raw
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
func (s *Session) StreamHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

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
