package service

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
)

type SessionStore struct {
	mu     sync.RWMutex
	tokens map[string]string
}

var Sessions = &SessionStore{tokens: make(map[string]string)}

func (s *SessionStore) Create(email string) string {
	b := make([]byte, 32)
	rand.Read(b)
	token := hex.EncodeToString(b)
	s.mu.Lock()
	s.tokens[token] = email
	s.mu.Unlock()
	return token
}

func (s *SessionStore) Get(token string) (string, bool) {
	s.mu.RLock()
	email, ok := s.tokens[token]
	s.mu.RUnlock()
	return email, ok
}

func (s *SessionStore) Delete(token string) {
	s.mu.Lock()
	delete(s.tokens, token)
	s.mu.Unlock()
}
