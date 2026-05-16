package main

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"
)

type SessionStore struct {
	mu     sync.RWMutex
	tokens map[string]string
}

var sessions = &SessionStore{tokens: make(map[string]string)}

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

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing token"})
			return
		}
		_, ok := sessions.Get(token)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid token"})
			return
		}
		next(w, r)
	}
}
