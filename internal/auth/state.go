package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"planets-server/internal/shared/redis"
)

type StateManager struct {
	redis       *redis.Client
	memoryStore map[string]StateEntry
	mutex       sync.RWMutex
	useRedis    bool
}

type StateEntry struct {
	CreatedAt time.Time `json:"created_at"`
	Provider  string    `json:"provider"`
	UserAgent string    `json:"user_agent"`
}

var globalStateManager *StateManager

func InitStateManager(redisClient *redis.Client) {
	useRedis := redisClient != nil

	globalStateManager = &StateManager{
		redis:       redisClient,
		memoryStore: make(map[string]StateEntry),
		useRedis:    useRedis,
	}

	logger := slog.With("component", "state_manager", "operation", "init")
	if useRedis {
		logger.Info("OAuth state manager initialized with Redis")
	} else {
		logger.Warn("OAuth state manager using in-memory fallback (not production-safe)")
		go globalStateManager.startMemoryCleanup()
	}
}

func (sm *StateManager) GenerateState(provider, userAgent string) (string, error) {
	logger := slog.With("component", "state_manager", "operation", "generate", "provider", provider)

	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		logger.Error("Failed to generate random bytes for state token", "error", err)
		return "", fmt.Errorf("failed to generate state token: %w", err)
	}

	state := base64.URLEncoding.EncodeToString(b)
	entry := StateEntry{
		CreatedAt: time.Now(),
		Provider:  provider,
		UserAgent: userAgent,
	}

	if sm.useRedis {
		return state, sm.storeInRedis(state, entry, logger)
	}

	return state, sm.storeInMemory(state, entry, logger)
}

func (sm *StateManager) storeInRedis(state string, entry StateEntry, logger *slog.Logger) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	data, err := json.Marshal(entry)
	if err != nil {
		logger.Error("Failed to marshal state entry", "error", err)
		return fmt.Errorf("failed to marshal state entry: %w", err)
	}

	key := fmt.Sprintf("oauth:state:%s", state)
	err = sm.redis.Set(ctx, key, data, 10*time.Minute).Err()
	if err != nil {
		logger.Error("Failed to store state in Redis", "error", err)
		return fmt.Errorf("failed to store state in Redis: %w", err)
	}

	logger.Debug("OAuth state token stored in Redis",
		"state_length", len(state),
		"ttl", "10m")

	return nil
}

func (sm *StateManager) storeInMemory(state string, entry StateEntry, logger *slog.Logger) error {
	sm.mutex.Lock()
	sm.memoryStore[state] = entry
	sm.mutex.Unlock()

	logger.Debug("OAuth state token stored in memory",
		"state_length", len(state),
		"user_agent_length", len(entry.UserAgent))

	return nil
}

func (sm *StateManager) ValidateState(state, provider, userAgent string) error {
	logger := slog.With("component", "state_manager", "operation", "validate", "provider", provider)

	if state == "" {
		logger.Warn("Empty state token provided")
		return fmt.Errorf("state token is required")
	}

	if sm.useRedis {
		return sm.validateFromRedis(state, provider, userAgent, logger)
	}

	return sm.validateFromMemory(state, provider, userAgent, logger)
}

func (sm *StateManager) validateFromRedis(state, provider, userAgent string, logger *slog.Logger) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	key := fmt.Sprintf("oauth:state:%s", state)
	data, err := sm.redis.Get(ctx, key).Bytes()
	if err != nil {
		logger.Warn("Invalid or expired state token", "error", err)
		return fmt.Errorf("invalid or expired state token")
	}

	if err := sm.redis.Del(ctx, key).Err(); err != nil {
		logger.Error("Failed to delete state from Redis", "error", err)
	}

	var entry StateEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		logger.Error("Failed to unmarshal state entry", "error", err)
		return fmt.Errorf("invalid state data")
	}

	return sm.validateEntry(entry, provider, userAgent, logger)
}

func (sm *StateManager) validateFromMemory(state, provider, userAgent string, logger *slog.Logger) error {
	sm.mutex.Lock()
	entry, exists := sm.memoryStore[state]
	if exists {
		delete(sm.memoryStore, state)
	}
	sm.mutex.Unlock()

	if !exists {
		logger.Warn("Invalid or expired state token", "state_exists", false)
		return fmt.Errorf("invalid or expired state token")
	}

	return sm.validateEntry(entry, provider, userAgent, logger)
}

func (sm *StateManager) validateEntry(entry StateEntry, provider, userAgent string, logger *slog.Logger) error {
	if time.Since(entry.CreatedAt) > 10*time.Minute {
		logger.Warn("Expired state token",
			"created_at", entry.CreatedAt,
			"age_minutes", time.Since(entry.CreatedAt).Minutes())
		return fmt.Errorf("state token has expired")
	}

	if entry.Provider != provider {
		logger.Warn("State token provider mismatch",
			"expected_provider", entry.Provider,
			"received_provider", provider)
		return fmt.Errorf("state token provider mismatch")
	}

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

func (sm *StateManager) startMemoryCleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	logger := slog.With("component", "state_manager", "operation", "cleanup")
	logger.Debug("Starting memory cleanup goroutine")

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

	for state, entry := range sm.memoryStore {
		if now.Sub(entry.CreatedAt) > 10*time.Minute {
			delete(sm.memoryStore, state)
			expiredCount++
		}
	}

	if expiredCount > 0 {
		logger.Debug("Cleaned up expired state tokens",
			"expired_count", expiredCount,
			"remaining_count", len(sm.memoryStore))
	}
}

func GenerateOAuthState(provider, userAgent string) (string, error) {
	if globalStateManager == nil {
		return "", fmt.Errorf("state manager not initialized")
	}
	return globalStateManager.GenerateState(provider, userAgent)
}

func ValidateOAuthState(state, provider, userAgent string) error {
	if globalStateManager == nil {
		return fmt.Errorf("state manager not initialized")
	}
	return globalStateManager.ValidateState(state, provider, userAgent)
}
