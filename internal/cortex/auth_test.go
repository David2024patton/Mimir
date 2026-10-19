package cortex

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
)

// fakeTokenDB emulates the SurrealDB /sql endpoint for auth_token statements.
// It models a table holding exactly one stored token value: a SELECT matches
// only when one of the queried token values equals the stored value, like a
// real equality lookup against the stored column.
type fakeTokenDB struct {
	stored     string   // the token column value present in the table
	createSQL  []string // captured CREATE statements
	tokenQuery []string // token values seen in SELECT token = '...' clauses
}

var tokenEqRe = regexp.MustCompile(`token = '([^']*)'`)

func (f *fakeTokenDB) handler(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	sql := string(body)
	w.Header().Set("Content-Type", "application/json")
	switch {
	case strings.Contains(sql, "CREATE auth_token"):
		f.createSQL = append(f.createSQL, sql)
		fmt.Fprint(w, `[{"result":[],"status":"OK"}]`)
	case strings.Contains(sql, "SELECT * FROM auth_token"):
		var rows []map[string]any
		for _, m := range tokenEqRe.FindAllStringSubmatch(sql, -1) {
			f.tokenQuery = append(f.tokenQuery, m[1])
			if m[1] == f.stored {
				rows = []map[string]any{{
					"id":      "auth_token:t1",
					"user_id": "user:u42",
					"token":   f.stored,
				}}
			}
		}
		if rows == nil {
			rows = []map[string]any{}
		}
		out, _ := json.Marshal([]map[string]any{{"result": rows, "status": "OK"}})
		w.Write(out)
	default:
		fmt.Fprint(w, `[{"result":[],"status":"OK"}]`)
	}
}

func newFakeAuthStore(t *testing.T, db *fakeTokenDB) *AuthStore {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(db.handler))
	t.Cleanup(srv.Close)
	return NewAuthStore(&surrealClient{addr: srv.URL, auth: "Basic dGVzdDp0ZXN0"})
}

// A legacy row storing the bearer token in plaintext must no longer validate
// (F17 sunset): ValidateToken looks up the token hash only.
func TestValidateTokenRejectsPlaintextStoredToken(t *testing.T) {
	plaintext := generateToken()
	db := &fakeTokenDB{stored: plaintext}
	a := newFakeAuthStore(t, db)

	if _, err := a.ValidateToken(context.Background(), plaintext); err == nil {
		t.Fatal("plaintext-stored token validated; the plaintext fallback is still active")
	}
	for _, q := range db.tokenQuery {
		if q == plaintext {
			t.Fatal("ValidateToken queried the database with the plaintext token")
		}
	}
}

// A token stored as its SHA-256 hash (what createToken writes) must validate.
func TestValidateTokenAcceptsHashedStoredToken(t *testing.T) {
	plaintext := generateToken()
	db := &fakeTokenDB{stored: hashToken(plaintext)}
	a := newFakeAuthStore(t, db)

	userID, err := a.ValidateToken(context.Background(), plaintext)
	if err != nil {
		t.Fatalf("hashed token rejected: %v", err)
	}
	if userID != "u42" {
		t.Fatalf("userID = %q, want u42", userID)
	}
}

// createToken must persist the hash, never the plaintext bearer token.
func TestCreateTokenStoresHashOnly(t *testing.T) {
	db := &fakeTokenDB{}
	a := newFakeAuthStore(t, db)

	tok, err := a.createToken(context.Background(), "u42")
	if err != nil {
		t.Fatalf("createToken: %v", err)
	}
	if len(db.createSQL) != 1 {
		t.Fatalf("captured %d CREATE statements, want 1", len(db.createSQL))
	}
	sql := db.createSQL[0]
	if !strings.Contains(sql, "'"+hashToken(tok.Token)+"'") {
		t.Fatal("CREATE does not store the token hash")
	}
	if strings.Contains(sql, "'"+tok.Token+"'") {
		t.Fatal("CREATE stores the plaintext token")
	}
}
