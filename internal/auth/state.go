package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

type StateManager struct {
	states map[string]StateEntry
	mutex  sync.RWMutex
}

type StateEntry struct {
	CreatedAt time.Time
	Provider  string
	UserAgent string
}

var globalStateManager *StateManager

func init() {
	globalStateManager = NewStateManager()
	// Start cleanup goroutine
	go globalStateManager.startCleanup()
}

func NewStateManager() *StateManager {
	return &StateManager{
		states: make(map[string]StateEntry),
	}
}

// GenerateState creates a new state token and stores it for validation
func (sm *StateManager) GenerateState(provider, userAgent string) (string, error) {
	logger := slog.With("component", "state_manager", "operation", "generate", "provider", provider)

	// Generate cryptographically secure random state
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		logger.Error("Failed to generate random bytes for state token", "error", err)
		return "", fmt.Errorf("failed to generate state token: %w", err)
	}

	state := base64.URLEncoding.EncodeToString(b)

	// Store state with metadata
	sm.mutex.Lock()
	sm.states[state] = StateEntry{
		CreatedAt: time.Now(),
		Provider:  provider,
		UserAgent: userAgent,
	}
	sm.mutex.Unlock()

	logger.Debug("OAuth state token generated and stored",
		"state_length", len(state),
		"user_agent_length", len(userAgent))

	return state, nil
}

// ValidateState checks if the state token is valid and removes it (one-time use)
func (sm *StateManager) ValidateState(state, provider, userAgent string) error {
	logger := slog.With("component", "state_manager", "operation", "validate", "provider", provider)

	if state == "" {
		logger.Warn("Empty state token provided")
		return fmt.Errorf("state token is required")
	}

	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	entry, exists := sm.states[state]
	if !exists {
		logger.Warn("Invalid or expired state token", "state_exists", false)
		return fmt.Errorf("invalid or expired state token")
	}

	// Remove state immediately (one-time use)
	delete(sm.states, state)

	// Check if state has expired (10 minutes max)
	if time.Since(entry.CreatedAt) > 10*time.Minute {
		logger.Warn("Expired state token",
			"created_at", entry.CreatedAt,
			"age_minutes", time.Since(entry.CreatedAt).Minutes())
		return fmt.Errorf("state token has expired")
	}

	// Validate provider matches
	if entry.Provider != provider {
		logger.Warn("State token provider mismatch",
			"expected_provider", entry.Provider,
			"received_provider", provider)
		return fmt.Errorf("state token provider mismatch")
	}

	// Validate user agent for additional security
	if entry.UserAgent != userAgent {
		logger.Warn("State token user agent mismatch - possible session hijacking attempt",
			"stored_user_agent", entry.UserAgent,
			"received_user_agent", userAgent)
		// Uncomment next line for strict user agent validation (optional)
		// return fmt.Errorf("state token user agent mismatch")
	}

	logger.Debug("State token validated successfully",
		"token_age_seconds", time.Since(entry.CreatedAt).Seconds())

	return nil
}

// startCleanup runs a background goroutine to clean up expired state tokens
func (sm *StateManager) startCleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	logger := slog.With("component", "state_manager", "operation", "cleanup")
	logger.Debug("Starting state cleanup goroutine")

	for range ticker.C {
		sm.cleanupExpiredStates()
	}
}

func (sm *StateManager) cleanupExpiredStates() {
	logger := slog.With("component", "state_manager", "operation", "cleanup_expired")

	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	now := time.Now()
	expiredCount := 0

	for state, entry := range sm.states {
		if now.Sub(entry.CreatedAt) > 10*time.Minute {
			delete(sm.states, state)
			expiredCount++
		}
	}

	if expiredCount > 0 {
		logger.Debug("Cleaned up expired state tokens",
			"expired_count", expiredCount,
			"remaining_count", len(sm.states))
	}
}

// GetStats returns statistics about the state manager (useful for monitoring)
func (sm *StateManager) GetStats() map[string]interface{} {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	return map[string]interface{}{
		"active_states": len(sm.states),
		"timestamp":     time.Now(),
	}
}

// Helper functions to use the global state manager
func GenerateOAuthState(provider, userAgent string) (string, error) {
	return globalStateManager.GenerateState(provider, userAgent)
}

func ValidateOAuthState(state, provider, userAgent string) error {
	return globalStateManager.ValidateState(state, provider, userAgent)
}
