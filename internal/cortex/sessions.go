// Package cortex implements Mímir's memory layer (the Cortex). Sessions provide
// multi-turn conversation persistence so the agent remembers what was said across
// page reloads and browser sessions (F2.2).
package cortex

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Session represents a conversation thread with persistent message history.
type Session struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	UserID    string    `json:"user_id"`
	Messages  []Message `json:"messages"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Message is one turn in a conversation.
type Message struct {
	Role      string `json:"role"`      // "user" or "assistant"
	Content   string `json:"content"`
	Timestamp int64  `json:"timestamp"` // Unix nanos
}

// SessionStore is the persistence interface for conversation threads.
type SessionStore interface {
	CreateSession(ctx context.Context, title, userID string) (*Session, error)
	GetSession(ctx context.Context, id string) (*Session, error)
	ListSessions(ctx context.Context, userID string) ([]Session, error)
	AppendMessage(ctx context.Context, sessionID string, msg Message) error
	DeleteSession(ctx context.Context, id string) error
	UpdateTitle(ctx context.Context, id, title string) error
}

// surrealSessionStore manages conversation sessions in SurrealDB.
type surrealSessionStore struct {
	cli *surrealClient
	ns  string
	db  string
}

// NewSessionStore returns a session store wired to SurrealDB.
func NewSessionStore(cli *surrealClient) *surrealSessionStore {
	return &surrealSessionStore{cli: cli, ns: "mimir", db: "mimir"}
}

// Exec runs a SurrealQL statement against the session store's namespace/database.
func (s *surrealSessionStore) Exec(ctx context.Context, sql string) error {
	results, err := s.cli.exec(ctx, s.ns, s.db, sql)
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

// Query runs a SurrealQL query and returns the results.
func (s *surrealSessionStore) Query(ctx context.Context, sql string) ([]map[string]any, error) {
	results, err := s.cli.exec(ctx, s.ns, s.db, sql)
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

// EnsureSessionsTable creates the session table if it doesn't exist.
func (s *surrealSessionStore) EnsureSessionsTable(ctx context.Context) error {
	query := `DEFINE TABLE IF NOT EXISTS session SCHEMAFULL`
	if err := s.Exec(ctx, query); err != nil {
		return fmt.Errorf("define session table: %w", err)
	}
	query = `DEFINE FIELD IF NOT EXISTS title ON session TYPE string`
	if err := s.Exec(ctx, query); err != nil {
		return fmt.Errorf("define title field: %w", err)
	}
	query = `DEFINE FIELD IF NOT EXISTS user_id ON session TYPE record<user>`
	if err := s.Exec(ctx, query); err != nil {
		return fmt.Errorf("define user_id field: %w", err)
	}
	query = `DEFINE FIELD IF NOT EXISTS messages ON session TYPE array`
	if err := s.Exec(ctx, query); err != nil {
		return fmt.Errorf("define messages field: %w", err)
	}
	query = `DEFINE FIELD IF NOT EXISTS messages.* ON session TYPE object`
	if err := s.Exec(ctx, query); err != nil {
		return fmt.Errorf("define messages item: %w", err)
	}
	query = `DEFINE FIELD IF NOT EXISTS messages.*.role ON session TYPE string`
	if err := s.Exec(ctx, query); err != nil {
		return fmt.Errorf("define messages.role: %w", err)
	}
	query = `DEFINE FIELD IF NOT EXISTS messages.*.content ON session TYPE string`
	if err := s.Exec(ctx, query); err != nil {
		return fmt.Errorf("define messages.content: %w", err)
	}
	query = `DEFINE FIELD IF NOT EXISTS messages.*.timestamp ON session TYPE number`
	if err := s.Exec(ctx, query); err != nil {
		return fmt.Errorf("define messages.timestamp: %w", err)
	}
	query = `DEFINE FIELD IF NOT EXISTS created_at ON session TYPE datetime`
	if err := s.Exec(ctx, query); err != nil {
		return fmt.Errorf("define created_at field: %w", err)
	}
	query = `DEFINE FIELD IF NOT EXISTS updated_at ON session TYPE datetime`
	if err := s.Exec(ctx, query); err != nil {
		return fmt.Errorf("define updated_at field: %w", err)
	}
	query = `DEFINE INDEX IF NOT EXISTS session_updated ON session FIELDS updated_at`
	if err := s.Exec(ctx, query); err != nil {
		return fmt.Errorf("define session_updated index: %w", err)
	}
	return nil
}

// CreateSession creates a new session and returns it.
func (s *surrealSessionStore) CreateSession(ctx context.Context, title, userID string) (*Session, error) {
	now := time.Now().UTC()
	sessID := "s" + now.Format("20060102150405") + fmt.Sprintf("%09d", now.Nanosecond())
	sess := &Session{
		ID:        sessID,
		Title:     title,
		UserID:    userID,
		Messages:  []Message{},
		CreatedAt: now,
		UpdatedAt: now,
	}
	msgsJSON, _ := json.Marshal(sess.Messages)
	query := fmt.Sprintf(
		`CREATE session:%s SET title = %s, user_id = user:%s, messages = %s, created_at = time::now(), updated_at = time::now()`,
		sess.ID, escapeStringForSurreal(sess.Title), userID, string(msgsJSON),
	)
	if err := s.Exec(ctx, query); err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}
	return sess, nil
}

func escapeStringForSurreal(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "\\'") + "'"
}

// ListSessions returns all sessions sorted by updated_at descending.
func (s *surrealSessionStore) ListSessions(ctx context.Context, userID string) ([]Session, error) {
	query := fmt.Sprintf(`SELECT * FROM session WHERE user_id = user:%s ORDER BY updated_at DESC`, userID)
	rows, err := s.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}
	var sessions []Session
	for _, row := range rows {
		b, _ := json.Marshal(row)
		var raw struct {
			ID        string    `json:"id"`
			Title     string    `json:"title"`
			UserID    string    `json:"user_id"`
			Messages  []Message `json:"messages"`
			CreatedAt string    `json:"created_at"`
			UpdatedAt string    `json:"updated_at"`
		}
		if err := json.Unmarshal(b, &raw); err != nil {
			continue
		}
		sess := Session{
			ID:       StripRecordID(raw.ID),
			Title:    raw.Title,
			UserID:   StripRecordID(raw.UserID),
			Messages: raw.Messages,
		}
		if t, err := time.Parse(time.RFC3339Nano, raw.CreatedAt); err == nil {
			sess.CreatedAt = t
		}
		if t, err := time.Parse(time.RFC3339Nano, raw.UpdatedAt); err == nil {
			sess.UpdatedAt = t
		}
		sessions = append(sessions, sess)
	}
	return sessions, nil
}

// stripRecordID removes the "table:" prefix SurrealDB adds to record IDs.
func StripRecordID(id string) string {
	if i := strings.Index(id, ":"); i >= 0 {
		return id[i+1:]
	}
	return id
}

// GetSession returns a session with all its messages.
func (s *surrealSessionStore) GetSession(ctx context.Context, id string) (*Session, error) {
	// Strip table prefix if the caller passed a full record ID like "session:abc"
	cleanID := StripRecordID(id)
	query := fmt.Sprintf(`SELECT * FROM session:%s`, cleanID)
	rows, err := s.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get session: %w", err)
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("session not found: %s", id)
	}
	b, _ := json.Marshal(rows[0])
	var raw struct {
		ID        string    `json:"id"`
		Title     string    `json:"title"`
		UserID    string    `json:"user_id"`
		Messages  []Message `json:"messages"`
		CreatedAt string    `json:"created_at"`
		UpdatedAt string    `json:"updated_at"`
	}
	if err := json.Unmarshal(b, &raw); err != nil {
		return nil, fmt.Errorf("unmarshal session: %w", err)
	}
	sess := &Session{
		ID:       StripRecordID(raw.ID),
		Title:    raw.Title,
		UserID:   StripRecordID(raw.UserID),
		Messages: raw.Messages,
	}
	if t, err := time.Parse(time.RFC3339Nano, raw.CreatedAt); err == nil {
		sess.CreatedAt = t
	}
	if t, err := time.Parse(time.RFC3339Nano, raw.UpdatedAt); err == nil {
		sess.UpdatedAt = t
	}
	return sess, nil
}

// AppendMessage adds a message to a session and updates the timestamp.
func (s *surrealSessionStore) AppendMessage(ctx context.Context, sessionID string, msg Message) error {
	msgJSON, _ := json.Marshal(msg)
	query := fmt.Sprintf(
		`UPDATE session:%s SET messages = array::append(messages, %s), updated_at = time::now()`,
		sessionID, string(msgJSON),
	)
	if err := s.Exec(ctx, query); err != nil {
		return fmt.Errorf("append message: %w", err)
	}
	return nil
}

// DeleteSession removes a session.
func (s *surrealSessionStore) DeleteSession(ctx context.Context, id string) error {
	query := fmt.Sprintf(`DELETE session:%s`, id)
	if err := s.Exec(ctx, query); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

// UpdateTitle updates a session's title.
func (s *surrealSessionStore) UpdateTitle(ctx context.Context, id, title string) error {
	query := fmt.Sprintf(`UPDATE session:%s SET title = '%s'`, id, title)
	if err := s.Exec(ctx, query); err != nil {
		return fmt.Errorf("update title: %w", err)
	}
	return nil
}
