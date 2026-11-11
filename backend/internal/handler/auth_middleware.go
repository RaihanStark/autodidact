package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"autodidact-backend/internal/model"
)

type AuthHandler struct {
	expectedToken string
}

func NewAuthHandler(token string) *AuthHandler {
	return &AuthHandler{expectedToken: token}
}

func (a *AuthHandler) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If no token is configured, allow open local access
		if a.expectedToken == "" {
			next.ServeHTTP(w, r)
			return
		}

		// Allow public health check and auth verify
		if r.URL.Path == "/api/health" || r.URL.Path == "/api/auth/verify" {
			next.ServeHTTP(w, r)
			return
		}

		token := extractToken(r)
		if token != a.expectedToken {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Unauthorized: invalid or missing security key",
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (a *AuthHandler) Verify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req model.VerifyAuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		token := extractToken(r)
		req.Token = token
	}

	valid := (a.expectedToken == "" || req.Token == a.expectedToken)

	w.Header().Set("Content-Type", "application/json")
	if valid {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{
			"valid":   true,
			"message": "Access granted",
		})
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]any{
			"valid":   false,
			"message": "Invalid security key",
		})
	}
}

func extractToken(r *http.Request) string {
	if auth := r.Header.Get("Authorization"); auth != "" {
		if strings.HasPrefix(strings.ToLower(auth), "bearer ") {
			return strings.TrimSpace(auth[7:])
		}
		return strings.TrimSpace(auth)
	}
	if tok := r.Header.Get("X-Auth-Token"); tok != "" {
		return strings.TrimSpace(tok)
	}
	if tok := r.URL.Query().Get("token"); tok != "" {
		return strings.TrimSpace(tok)
	}
	return ""
}
