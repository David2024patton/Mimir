package tools

import (
	"context"
	"fmt"
	"sort"
	"sync"
)

// TodoStatus is the state of a to-do item.
type TodoStatus string

const (
	TodoPending    TodoStatus = "pending"
	TodoInProgress TodoStatus = "in_progress"
	TodoCompleted  TodoStatus = "completed"
	TodoBlocked    TodoStatus = "blocked"
)

// Todo is one to-do item (the small model's working memory; F31/E3.5).
type Todo struct {
	ID      int        `json:"id"`
	Ord     int        `json:"ord"`
	Content string     `json:"content"`
	Status  TodoStatus `json:"status"`
}

// TodoStore is an in-memory to-do list (backed by SurrealDB in E1.4/E3.5).
type TodoStore struct {
	mu     sync.Mutex
	nextID int
	items  map[int]*Todo
}

func NewTodoStore() *TodoStore {
	return &TodoStore{nextID: 1, items: map[int]*Todo{}}
}

func (s *TodoStore) Add(content string) *Todo {
	s.mu.Lock()
	defer s.mu.Unlock()
	t := &Todo{ID: s.nextID, Ord: s.nextID, Content: content, Status: TodoPending}
	s.items[t.ID] = t
	s.nextID++
	return t
}

func (s *TodoStore) SetStatus(id int, status TodoStatus) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if t, ok := s.items[id]; ok {
		t.Status = status
		return true
	}
	return false
}

func (s *TodoStore) List() []*Todo {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]*Todo, 0, len(s.items))
	for _, t := range s.items {
		out = append(out, t)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Ord < out[j].Ord })
	return out
}

// TodoTool exposes the to-do store as an agent Tool.
type TodoTool struct{ Store *TodoStore }

func (t *TodoTool) Name() string        { return "todo" }
func (t *TodoTool) Description() string { return "Manage the task to-do list (add, update status, list)." }
func (t *TodoTool) Parameters() any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"action":  map[string]any{"type": "string", "enum": []string{"add", "update", "list"}},
			"content": map[string]any{"type": "string"},
			"id":      map[string]any{"type": "integer"},
			"status":  map[string]any{"type": "string"},
		},
		"required": []string{"action"},
	}
}

func (t *TodoTool) Run(ctx context.Context, args map[string]any) (string, error) {
	action, _ := args["action"].(string)
	switch action {
	case "add":
		content, _ := args["content"].(string)
		todo := t.Store.Add(content)
		return fmt.Sprintf("added todo #%d: %s", todo.ID, todo.Content), nil
	case "update":
		id, _ := args["id"].(float64)
		status, _ := args["status"].(string)
		if t.Store.SetStatus(int(id), TodoStatus(status)) {
			return fmt.Sprintf("todo #%d -> %s", int(id), status), nil
		}
		return "", fmt.Errorf("todo #%d not found", int(id))
	case "list":
		var s string
		for _, todo := range t.Store.List() {
			s += fmt.Sprintf("#%d [%s] %s\n", todo.ID, todo.Status, todo.Content)
		}
		return s, nil
	default:
		return "", fmt.Errorf("unknown action %q", action)
	}
}
