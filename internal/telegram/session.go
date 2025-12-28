package telegram

import (
	"fmt"
	"sync"
	"time"
)

// UserSession represents a Telegram user's vault session
type UserSession struct {
	TelegramUserID       int64
	VaultName            string
	SessionToken         string
	LastActivity         time.Time
	LoginState           LoginState
	PendingVault         string
	PasswordPromptMsgID  int // Message ID of password prompt to delete
}

// LoginState tracks the user's login flow state
type LoginState int

const (
	StateIdle LoginState = iota
	StateAwaitingVaultName
	StateAwaitingMasterPassword
)

// SessionManager manages user sessions with auto-expiry
type SessionManager struct {
	sessions      map[int64]*UserSession
	mu            sync.RWMutex
	sessionTTL    time.Duration
	cleanupTicker *time.Ticker
	done          chan bool
}

// NewSessionManager creates a new session manager
func NewSessionManager(sessionTTL time.Duration) *SessionManager {
	sm := &SessionManager{
		sessions:      make(map[int64]*UserSession),
		sessionTTL:    sessionTTL,
		cleanupTicker: time.NewTicker(1 * time.Minute),
		done:          make(chan bool),
	}

	go sm.cleanupExpiredSessions()
	return sm
}

// CreateSession creates or updates a session for a user
func (sm *SessionManager) CreateSession(userID int64, vaultName string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.sessions[userID] = &UserSession{
		TelegramUserID: userID,
		VaultName:      vaultName,
		LastActivity:   time.Now(),
		LoginState:     StateIdle,
	}
}

// GetSession retrieves a user's session
func (sm *SessionManager) GetSession(userID int64) (*UserSession, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, exists := sm.sessions[userID]
	if !exists {
		return nil, fmt.Errorf("session not found")
	}

	return session, nil
}

// UpdateActivity updates the last activity time for a session
func (sm *SessionManager) UpdateActivity(userID int64) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if session, exists := sm.sessions[userID]; exists {
		session.LastActivity = time.Now()
	}
}

// SetLoginState sets the login state for a user
func (sm *SessionManager) SetLoginState(userID int64, state LoginState, pendingVault string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if session, exists := sm.sessions[userID]; exists {
		session.LoginState = state
		session.PendingVault = pendingVault
		session.LastActivity = time.Now()
	} else {
		sm.sessions[userID] = &UserSession{
			TelegramUserID: userID,
			LoginState:     state,
			PendingVault:   pendingVault,
			LastActivity:   time.Now(),
		}
	}
}

// GetLoginState retrieves the login state for a user
func (sm *SessionManager) GetLoginState(userID int64) (LoginState, string) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if session, exists := sm.sessions[userID]; exists {
		return session.LoginState, session.PendingVault
	}
	return StateIdle, ""
}

// SetPasswordPromptMsgID stores the message ID of the password prompt
func (sm *SessionManager) SetPasswordPromptMsgID(userID int64, messageID int) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if session, exists := sm.sessions[userID]; exists {
		session.PasswordPromptMsgID = messageID
	}
}

// GetAndClearPasswordPromptMsgID retrieves and clears the password prompt message ID
func (sm *SessionManager) GetAndClearPasswordPromptMsgID(userID int64) int {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if session, exists := sm.sessions[userID]; exists {
		msgID := session.PasswordPromptMsgID
		session.PasswordPromptMsgID = 0
		return msgID
	}
	return 0
}

// IsAuthenticated checks if a user has an active vault session
func (sm *SessionManager) IsAuthenticated(userID int64) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, exists := sm.sessions[userID]
	if !exists {
		return false
	}

	return session.VaultName != "" && session.LoginState == StateIdle
}

// DeleteSession removes a user's session
func (sm *SessionManager) DeleteSession(userID int64) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	delete(sm.sessions, userID)
}

// cleanupExpiredSessions periodically removes expired sessions
func (sm *SessionManager) cleanupExpiredSessions() {
	for {
		select {
		case <-sm.cleanupTicker.C:
			sm.mu.Lock()
			now := time.Now()
			for userID, session := range sm.sessions {
				if now.Sub(session.LastActivity) > sm.sessionTTL {
					delete(sm.sessions, userID)
				}
			}
			sm.mu.Unlock()
		case <-sm.done:
			return
		}
	}
}

// Stop stops the session cleanup goroutine
func (sm *SessionManager) Stop() {
	sm.cleanupTicker.Stop()
	sm.done <- true
}
