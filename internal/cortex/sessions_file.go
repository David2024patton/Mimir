package cortex

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// FileSessionStore stores conversation sessions in a JSON file, matching
// the MemoryStore pattern for a zero-dependency local backend.
type FileSessionStore struct {
	mu       sync.Mutex
	path     string
	sessions map[string]*Session
}

// NewFileSessionStore creates a session store backed by dir/sessions.json.
// It loads any existing file on start; a missing file is not an error.
func NewFileSessionStore(dir string) (*FileSessionStore, error) {
	s := &FileSessionStore{
		path:     filepath.Join(dir, "sessions.json"),
		sessions: map[string]*Session{},
	}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

func (f *FileSessionStore) load() error {
	data, err := os.ReadFile(f.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	var list []*Session
	if err := json.Unmarshal(data, &list); err != nil {
		return err
	}
	for _, s := range list {
		f.sessions[s.ID] = s
	}
	return nil
}

func (f *FileSessionStore) save() error {
	list := make([]*Session, 0, len(f.sessions))
	for _, s := range f.sessions {
		list = append(list, s)
	}
	data, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(f.path), 0755); err != nil {
		return err
	}
	return os.WriteFile(f.path, data, 0644)
}

func (f *FileSessionStore) CreateSession(_ context.Context, title, userID string) (*Session, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	now := time.Now().UTC()
	id := "s" + now.Format("20060102150405") + fmt.Sprintf("%09d", now.Nanosecond())
	s := &Session{
		ID:        id,
		Title:     title,
		UserID:    userID,
		Messages:  []Message{},
		CreatedAt: now,
		UpdatedAt: now,
	}
	f.sessions[id] = s
	if err := f.save(); err != nil {
		delete(f.sessions, id)
		return nil, fmt.Errorf("save sessions: %w", err)
	}
	return s, nil
}

func (f *FileSessionStore) GetSession(_ context.Context, id string) (*Session, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	s, ok := f.sessions[id]
	if !ok {
		return nil, fmt.Errorf("session not found: %s", id)
	}
	cp := *s
	return &cp, nil
}

func (f *FileSessionStore) ListSessions(_ context.Context, userID string) ([]Session, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	list := make([]Session, 0, len(f.sessions))
	for _, s := range f.sessions {
		if s.UserID == userID {
			list = append(list, *s)
		}
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].UpdatedAt.After(list[j].UpdatedAt)
	})
	return list, nil
}

func (f *FileSessionStore) AppendMessage(_ context.Context, sessionID string, msg Message) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	s, ok := f.sessions[sessionID]
	if !ok {
		return fmt.Errorf("session not found: %s", sessionID)
	}
	s.Messages = append(s.Messages, msg)
	s.UpdatedAt = time.Now().UTC()
	return f.save()
}

func (f *FileSessionStore) DeleteSession(_ context.Context, id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.sessions, id)
	return f.save()
}

func (f *FileSessionStore) UpdateTitle(_ context.Context, id, title string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	s, ok := f.sessions[id]
	if !ok {
		return fmt.Errorf("session not found: %s", id)
	}
	s.Title = title
	s.UpdatedAt = time.Now().UTC()
	return f.save()
}
