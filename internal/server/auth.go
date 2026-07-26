package server

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/David2024patton/Mimir/internal/cortex"
)

type contextKey string

const ctxUserID contextKey = "user_id"

func authMiddleware(auth *cortex.AuthStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ah := r.Header.Get("Authorization")
			if ah == "" || !strings.HasPrefix(ah, "Bearer ") {
				http.Error(w, `{"error":"authorization required"}`, http.StatusUnauthorized)
				return
			}
			token := strings.TrimPrefix(ah, "Bearer ")
			userID, err := auth.ValidateToken(r.Context(), token)
			if err != nil {
				http.Error(w, `{"error":"invalid or expired token"}`, http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), ctxUserID, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func userIDFromContext(ctx context.Context) string {
	uid, _ := ctx.Value(ctxUserID).(string)
	return uid
}

func sendCodeHandler(auth *cortex.AuthStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Email string `json:"email"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
			return
		}
		if req.Email == "" {
			http.Error(w, `{"error":"email is required"}`, http.StatusBadRequest)
			return
		}
		if err := auth.SendOTP(r.Context(), req.Email); err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
			return
		}
		writeJSON(w, map[string]string{"status": "sent"})
	}
}

func verifyCodeHandler(auth *cortex.AuthStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Email string `json:"email"`
			Code  string `json:"code"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
			return
		}
		if req.Email == "" || req.Code == "" {
			http.Error(w, `{"error":"email and code are required"}`, http.StatusBadRequest)
			return
		}
		user, tok, err := auth.VerifyOTP(r.Context(), req.Email, req.Code)
		if err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusUnauthorized)
			return
		}
		writeJSON(w, map[string]any{
			"user":  map[string]any{"id": user.ID, "email": user.Email, "created_at": user.CreatedAt},
			"token": tok.Token,
		})
	}
}

func meHandler(auth *cortex.AuthStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := userIDFromContext(r.Context())
		rows, err := auth.Query(r.Context(), "SELECT * FROM user WHERE id = user:"+userID)
		if err != nil || len(rows) == 0 {
			http.Error(w, `{"error":"user not found"}`, http.StatusNotFound)
			return
		}
		b, _ := json.Marshal(rows[0])
		var raw struct {
			ID        string `json:"id"`
			Email     string `json:"email"`
			CreatedAt string `json:"created_at"`
		}
		json.Unmarshal(b, &raw)
		writeJSON(w, map[string]any{
			"id":         cortex.StripRecordID(raw.ID),
			"email":      raw.Email,
			"created_at": raw.CreatedAt,
		})
	}
}
