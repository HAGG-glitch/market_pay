package store

import (
	"context"
	"sync"

	"github.com/rs/zerolog/log"
)

// InMemoryStateStore is an in-memory implementation of the StateStore interface for development/testing
type InMemoryStateStore struct {
	mu    sync.RWMutex
	store map[string]map[string]string
}

// NewInMemoryStateStore creates a new in-memory state store
func NewInMemoryStateStore() *InMemoryStateStore {
	return &InMemoryStateStore{
		store: make(map[string]map[string]string),
	}
}

// Load retrieves session state from memory
func (s *InMemoryStateStore) Load(ctx context.Context, sessionID string) (map[string]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	data, exists := s.store[sessionID]
	if !exists {
		log.Debug().Str("session_id", sessionID).Msg("session not found in memory store")
		return nil, nil
	}

	// Return a copy to prevent external modifications
	cloned := make(map[string]string, len(data))
	for k, v := range data {
		cloned[k] = v
	}

	log.Debug().
		Str("session_id", sessionID).
		Int("keys", len(cloned)).
		Msg("loaded session state from memory store")

	return cloned, nil
}

// Save stores session state in memory
func (s *InMemoryStateStore) Save(ctx context.Context, sessionID string, data map[string]string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if ctx.Err() != nil {
		return ctx.Err()
	}

	// Store a copy to prevent external modifications
	cloned := make(map[string]string, len(data))
	for k, v := range data {
		cloned[k] = v
	}

	s.store[sessionID] = cloned

	log.Debug().
		Str("session_id", sessionID).
		Int("keys", len(cloned)).
		Msg("saved session state to memory store")

	return nil
}

// Clear removes a session from memory
func (s *InMemoryStateStore) Clear(sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.store, sessionID)
	log.Debug().Str("session_id", sessionID).Msg("cleared session from memory store")
}

// ClearAll removes all sessions from memory
func (s *InMemoryStateStore) ClearAll() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.store = make(map[string]map[string]string)
	log.Debug().Msg("cleared all sessions from memory store")
}

// SessionCount returns the number of active sessions
func (s *InMemoryStateStore) SessionCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.store)
}
