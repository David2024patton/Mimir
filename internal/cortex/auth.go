package cortex

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net"
	"net/smtp"
	"os"
	"sync"
	"time"
)

type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type AuthToken struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Token     string    `json:"token"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

type OTP struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Code      string    `json:"code"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	Used      bool      `json:"used"`
}

type rateLimitEntry struct {
	sendCount   int
	lastSend    time.Time
	verifyFails int
	lockedUntil time.Time
}

type AuthStore struct {
	cli        *surrealClient
	ns         string
	db         string
	rateLimits map[string]*rateLimitEntry
	mu         sync.Mutex
}

func NewAuthStore(cli *surrealClient) *AuthStore {
	return &AuthStore{
		cli:        cli,
		ns:         "mimir",
		db:         "mimir",
		rateLimits: make(map[string]*rateLimitEntry),
	}
}

func (a *AuthStore) Exec(ctx context.Context, sql string) error {
	results, err := a.cli.exec(ctx, a.ns, a.db, sql)
	if err != nil {
		return err
	}
	for _, r := range results {
		if r.Status == "ERR" {
			return fmt.Errorf("surreal: %s", string(r.Result))
		}
	}
	return nil
}

func (a *AuthStore) Query(ctx context.Context, sql string) ([]map[string]any, error) {
	results, err := a.cli.exec(ctx, a.ns, a.db, sql)
	if err != nil {
		return nil, err
	}
	var rows []map[string]any
	for _, r := range results {
		if r.Status == "ERR" {
			return nil, fmt.Errorf("surreal: %s", string(r.Result))
		}
		var arr []map[string]any
		if err := json.Unmarshal(r.Result, &arr); err != nil {
			continue
		}
		rows = append(rows, arr...)
	}
	return rows, nil
}

func (a *AuthStore) EnsureAuthTables(ctx context.Context) error {
	// Remove legacy fields from earlier password-based schema
	_ = a.Exec(ctx, `REMOVE FIELD password_hash ON user`)
	_ = a.Exec(ctx, `REMOVE FIELD password ON user`)
	stmts := []string{
		`DEFINE TABLE IF NOT EXISTS user SCHEMAFULL`,
		`DEFINE FIELD IF NOT EXISTS email ON user TYPE string`,
		`DEFINE FIELD IF NOT EXISTS created_at ON user TYPE datetime`,
		`DEFINE INDEX IF NOT EXISTS user_email ON user FIELDS email UNIQUE`,
		`DEFINE TABLE IF NOT EXISTS auth_token SCHEMAFULL`,
		`DEFINE FIELD IF NOT EXISTS user_id ON auth_token TYPE record<user>`,
		`DEFINE FIELD IF NOT EXISTS token ON auth_token TYPE string`,
		`DEFINE FIELD IF NOT EXISTS created_at ON auth_token TYPE datetime`,
		`DEFINE FIELD IF NOT EXISTS expires_at ON auth_token TYPE datetime`,
		`DEFINE INDEX IF NOT EXISTS auth_token_value ON auth_token FIELDS token UNIQUE`,
		`DEFINE TABLE IF NOT EXISTS otp SCHEMAFULL`,
		`DEFINE FIELD IF NOT EXISTS email ON otp TYPE string`,
		`DEFINE FIELD IF NOT EXISTS code ON otp TYPE string`,
		`DEFINE FIELD IF NOT EXISTS created_at ON otp TYPE datetime`,
		`DEFINE FIELD IF NOT EXISTS expires_at ON otp TYPE datetime`,
		`DEFINE FIELD IF NOT EXISTS used ON otp TYPE bool`,
	}
	for _, s := range stmts {
		if err := a.Exec(ctx, s); err != nil {
			return fmt.Errorf("auth schema: %s: %w", s[:50], err)
		}
	}
	return nil
}

// SendOTP generates a 6-digit code, stores it in SurrealDB, and emails it.
func (a *AuthStore) SendOTP(ctx context.Context, email string) error {
	now := time.Now().UTC()

	// Rate limit check: maximum 3 requests within 2 minutes; lockout if triggered
	a.mu.Lock()
	entry, ok := a.rateLimits[email]
	if !ok {
		entry = &rateLimitEntry{}
		a.rateLimits[email] = entry
	}
	if entry.lockedUntil.After(now) {
		a.mu.Unlock()
		return fmt.Errorf("too many attempts: account locked until %s", entry.lockedUntil.Format("15:04:05"))
	}
	if now.Sub(entry.lastSend) < 2*time.Minute {
		if entry.sendCount >= 3 {
			entry.lockedUntil = now.Add(15 * time.Minute)
			a.mu.Unlock()
			return fmt.Errorf("rate limit exceeded: too many login codes requested, please wait 15 minutes")
		}
		entry.sendCount++
	} else {
		entry.sendCount = 1
		entry.verifyFails = 0
	}
	entry.lastSend = now
	a.mu.Unlock()

	// Invalidate any existing unused OTPs for this email
	_ = a.Exec(ctx, fmt.Sprintf(
		`UPDATE otp SET used = true WHERE email = %s AND used = false`,
		escapeStringForSurreal(email),
	))

	code := generateOTPCode()
	otpID := "o" + now.Format("20060102150405") + fmt.Sprintf("%09d", now.Nanosecond())
	q := fmt.Sprintf(
		`CREATE otp:%s SET email = %s, code = %s, created_at = time::now(), expires_at = time::now() + 5m, used = false`,
		otpID, escapeStringForSurreal(email), escapeStringForSurreal(code),
	)
	if err := a.Exec(ctx, q); err != nil {
		return fmt.Errorf("create otp: %w", err)
	}

	// Send email (best-effort, log failure)
	if err := sendOTPEmail(email, code); err != nil {
		fmt.Fprintf(os.Stderr, "auth: send OTP email to %s: %v\n", email, err)
	} else {
		fmt.Fprintf(os.Stderr, "auth: OTP sent to %s\n", email)
	}
	return nil
}

// VerifyOTP validates a code and returns an auth token (creates user on first login).
func (a *AuthStore) VerifyOTP(ctx context.Context, email, code string) (*User, *AuthToken, error) {
	now := time.Now().UTC()

	a.mu.Lock()
	entry, ok := a.rateLimits[email]
	if !ok {
		entry = &rateLimitEntry{}
		a.rateLimits[email] = entry
	}
	if entry.lockedUntil.After(now) {
		a.mu.Unlock()
		return nil, nil, fmt.Errorf("account temporarily locked due to too many failed attempts, please retry later")
	}
	a.mu.Unlock()

	rows, err := a.Query(ctx, fmt.Sprintf(
		`SELECT * FROM otp WHERE email = %s AND code = %s AND used = false AND expires_at > time::now() ORDER BY created_at DESC LIMIT 1`,
		escapeStringForSurreal(email), escapeStringForSurreal(code),
	))
	if err != nil {
		return nil, nil, fmt.Errorf("query otp: %w", err)
	}
	if len(rows) == 0 {
		a.mu.Lock()
		entry.verifyFails++
		if entry.verifyFails >= 5 {
			entry.lockedUntil = now.Add(15 * time.Minute)
		}
		a.mu.Unlock()
		return nil, nil, fmt.Errorf("invalid or expired code")
	}

	b, _ := json.Marshal(rows[0])
	var raw struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(b, &raw); err != nil {
		return nil, nil, fmt.Errorf("decode otp row: %w", err)
	}
	otpID := StripRecordID(raw.ID)

	// Mark OTP as used
	if err := a.Exec(ctx, fmt.Sprintf(`UPDATE otp:%s SET used = true`, otpID)); err != nil {
		return nil, nil, fmt.Errorf("mark otp used: %w", err)
	}

	// Reset failure count on success
	a.mu.Lock()
	entry.verifyFails = 0
	a.mu.Unlock()

	// Find or create user
	user, err := a.findOrCreateUser(ctx, email)
	if err != nil {
		return nil, nil, err
	}

	tok, err := a.createToken(ctx, user.ID)
	if err != nil {
		return nil, nil, err
	}
	return user, tok, nil
}

func (a *AuthStore) findOrCreateUser(ctx context.Context, email string) (*User, error) {
	rows, err := a.Query(ctx, fmt.Sprintf(
		`SELECT * FROM user WHERE email = %s`, escapeStringForSurreal(email),
	))
	if err != nil {
		return nil, err
	}
	if len(rows) > 0 {
		b, _ := json.Marshal(rows[0])
		var raw struct {
			ID        string `json:"id"`
			Email     string `json:"email"`
			CreatedAt string `json:"created_at"`
		}
		if err := json.Unmarshal(b, &raw); err != nil {
			return nil, fmt.Errorf("decode user row: %w", err)
		}
		return &User{
			ID:        StripRecordID(raw.ID),
			Email:     raw.Email,
			CreatedAt: parseTime(raw.CreatedAt),
		}, nil
	}

	now := time.Now().UTC()
	userID := "u" + now.Format("20060102150405") + fmt.Sprintf("%09d", now.Nanosecond())
	q := fmt.Sprintf(
		`CREATE user:%s SET email = %s, created_at = time::now()`,
		userID, escapeStringForSurreal(email),
	)
	if err := a.Exec(ctx, q); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return &User{ID: userID, Email: email, CreatedAt: now}, nil
}

// SeedAdmin ensures the super admin account exists.
func (a *AuthStore) SeedAdmin(ctx context.Context, email string) error {
	rows, err := a.Query(ctx, fmt.Sprintf(
		`SELECT * FROM user WHERE email = %s`, escapeStringForSurreal(email),
	))
	if err != nil {
		return err
	}
	if len(rows) > 0 {
		return nil // already exists
	}
	now := time.Now().UTC()
	q := fmt.Sprintf(
		`CREATE user:%s SET email = %s, created_at = time::now()`,
		"u"+now.Format("20060102150405")+fmt.Sprintf("%09d", now.Nanosecond()),
		escapeStringForSurreal(email),
	)
	return a.Exec(ctx, q)
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

func (a *AuthStore) ValidateToken(ctx context.Context, token string) (string, error) {
	// Look up by token hash only (F17: prevent plaintext bearer tokens at rest).
	// Legacy tokens stored in plaintext before hashing was introduced are
	// intentionally no longer accepted: their owners must sign in again to
	// receive a hashed token.
	tokenHash := hashToken(token)
	rows, err := a.Query(ctx, fmt.Sprintf(
		`SELECT * FROM auth_token WHERE token = %s AND expires_at > time::now()`,
		escapeStringForSurreal(tokenHash),
	))
	if err != nil {
		return "", fmt.Errorf("query token: %w", err)
	}
	if len(rows) == 0 {
		return "", fmt.Errorf("invalid or expired token")
	}
	b, _ := json.Marshal(rows[0])
	var raw struct {
		ID     string `json:"id"`
		UserID string `json:"user_id"`
	}
	if err := json.Unmarshal(b, &raw); err != nil {
		return "", fmt.Errorf("decode token row: %w", err)
	}
	return StripRecordID(raw.UserID), nil
}

func (a *AuthStore) createToken(ctx context.Context, userID string) (*AuthToken, error) {
	now := time.Now().UTC()
	tokID := "t" + now.Format("20060102150405") + fmt.Sprintf("%09d", now.Nanosecond())
	token := generateToken()
	tokenHash := hashToken(token)
	q := fmt.Sprintf(
		`CREATE auth_token:%s SET token = %s, user_id = user:%s, created_at = time::now(), expires_at = time::now() + 30d`,
		tokID, escapeStringForSurreal(tokenHash), userID,
	)
	if err := a.Exec(ctx, q); err != nil {
		return nil, fmt.Errorf("create token: %w", err)
	}
	// Return the plaintext token to the user for HTTP Authorization Bearer headers
	return &AuthToken{ID: tokID, UserID: userID, Token: token, CreatedAt: now, ExpiresAt: now.Add(30 * 24 * time.Hour)}, nil
}

func generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

func generateOTPCode() string {
	n, _ := rand.Int(rand.Reader, big.NewInt(1000000))
	return fmt.Sprintf("%06d", n.Int64())
}

func sendOTPEmail(to, code string) error {
	host := os.Getenv("MIMIR_SMTP_HOST")
	if host == "" {
		host = "smtp.titan.email"
	}
	port := os.Getenv("MIMIR_SMTP_PORT")
	if port == "" {
		port = "465"
	}
	user := os.Getenv("MIMIR_SMTP_USER")
	if user == "" {
		user = "mimir@itak.live"
	}
	pass := os.Getenv("MIMIR_SMTP_PASS")
	if pass == "" {
		return fmt.Errorf("MIMIR_SMTP_PASS environment variable is unset: refusing to send OTP email without configured credentials")
	}
	fromName := os.Getenv("MIMIR_SMTP_FROM_NAME")
	if fromName == "" {
		fromName = "Mímir"
	}
	fromAddr := os.Getenv("MIMIR_SMTP_FROM")
	if fromAddr == "" {
		fromAddr = user
	}

	msg := fmt.Sprintf("From: %s <%s>\r\nTo: %s\r\nSubject: Your Mímir login code\r\n\r\nYour Mímir login code is: %s\r\n\r\nThis code expires in 5 minutes.\r\n", fromName, fromAddr, to, code)

	addr := net.JoinHostPort(host, port)
	tc, err := tls.Dial("tcp", addr, &tls.Config{ServerName: host})
	if err != nil {
		return fmt.Errorf("tls dial: %w", err)
	}
	defer tc.Close()

	c, err := smtp.NewClient(tc, host)
	if err != nil {
		return fmt.Errorf("smtp client: %w", err)
	}
	defer c.Quit()

	auth := smtp.PlainAuth("", user, pass, host)
	if err := c.Auth(auth); err != nil {
		return fmt.Errorf("auth: %w", err)
	}
	if err := c.Mail(fromAddr); err != nil {
		return fmt.Errorf("mail from: %w", err)
	}
	if err := c.Rcpt(to); err != nil {
		return fmt.Errorf("rcpt: %w", err)
	}
	w, err := c.Data()
	if err != nil {
		return fmt.Errorf("data: %w", err)
	}
	_, err = w.Write([]byte(msg))
	if err != nil {
		return fmt.Errorf("write: %w", err)
	}
	return w.Close()
}

func parseTime(s string) time.Time {
	t, _ := time.Parse(time.RFC3339Nano, s)
	return t
}
