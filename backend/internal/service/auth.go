package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"sync"
	"time"
)

type sessionEntry struct {
	email     string
	createdAt time.Time
}

type SessionStore struct {
	mu     sync.RWMutex
	tokens map[string]sessionEntry
	ttl    time.Duration
}

func NewSessionStore(ttl time.Duration) *SessionStore {
	s := &SessionStore{
		tokens: make(map[string]sessionEntry),
		ttl:    ttl,
	}
	go s.cleanup()
	return s
}

func (s *SessionStore) Create(email string) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("rand.Read: %w", err)
	}
	token := hex.EncodeToString(b)
	s.mu.Lock()
	s.tokens[token] = sessionEntry{email: email, createdAt: time.Now()}
	s.mu.Unlock()
	return token, nil
}

func (s *SessionStore) Get(token string) (string, bool) {
	s.mu.RLock()
	entry, ok := s.tokens[token]
	s.mu.RUnlock()
	if !ok {
		return "", false
	}
	if time.Since(entry.createdAt) > s.ttl {
		s.Delete(token)
		return "", false
	}
	return entry.email, true
}

func (s *SessionStore) Delete(token string) {
	s.mu.Lock()
	delete(s.tokens, token)
	s.mu.Unlock()
}

func (s *SessionStore) cleanup() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		s.mu.Lock()
		for token, entry := range s.tokens {
			if time.Since(entry.createdAt) > s.ttl {
				delete(s.tokens, token)
			}
		}
		s.mu.Unlock()
	}
}

func (s *SessionStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.tokens)
}

func init() {
	log.Print("service/auth.go: use NewSessionStore(ttl) instead of global Sessions")
}
