package http

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"sync"
	"time"
)

const (
	CSRFTokenLength = 32
	CSRFTokenTTL    = 1 * time.Hour
	CSRFCookieName  = "csrf_token"
	CSRFHeaderName  = "X-CSRF-Token"
)

// CSRFManager manages CSRF tokens with automatic cleanup
type CSRFManager struct {
	tokens sync.Map // map[string]time.Time
}

// NewCSRFManager creates a new CSRF manager with automatic token cleanup
func NewCSRFManager() *CSRFManager {
	manager := &CSRFManager{}

	// Cleanup expired tokens every 15 minutes
	go func() {
		ticker := time.NewTicker(15 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			manager.cleanupExpiredTokens()
		}
	}()

	return manager
}

// generateToken creates a new cryptographically secure CSRF token
func (m *CSRFManager) generateToken() (string, error) {
	tokenBytes := make([]byte, CSRFTokenLength)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}

	token := base64.URLEncoding.EncodeToString(tokenBytes)
	m.tokens.Store(token, time.Now().Add(CSRFTokenTTL))

	return token, nil
}

// validateToken checks if a CSRF token is valid and not expired
func (m *CSRFManager) validateToken(token string) bool {
	if token == "" {
		return false
	}

	expiry, exists := m.tokens.Load(token)
	if !exists {
		return false
	}

	if time.Now().After(expiry.(time.Time)) {
		m.tokens.Delete(token)
		return false
	}

	return true
}

// cleanupExpiredTokens removes expired tokens from memory
func (m *CSRFManager) cleanupExpiredTokens() {
	now := time.Now()
	m.tokens.Range(func(key, value interface{}) bool {
		if now.After(value.(time.Time)) {
			m.tokens.Delete(key)
		}
		return true
	})
}

// CSRFMiddleware creates a middleware that validates CSRF tokens
func CSRFMiddleware(manager *CSRFManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip CSRF check for safe methods
			if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}

			// Skip CSRF check for /api/csrf-token endpoint
			if r.URL.Path == "/api/csrf-token" {
				next.ServeHTTP(w, r)
				return
			}

			// Get token from header
			headerToken := r.Header.Get(CSRFHeaderName)

			// Get token from cookie
			cookie, err := r.Cookie(CSRFCookieName)
			if err != nil || cookie.Value == "" {
				http.Error(w, "CSRF token missing", http.StatusForbidden)
				return
			}

			// Double-submit cookie pattern: header and cookie must match
			if headerToken != cookie.Value {
				http.Error(w, "CSRF token mismatch", http.StatusForbidden)
				return
			}

			// Validate token
			if !manager.validateToken(headerToken) {
				http.Error(w, "CSRF token invalid or expired", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
