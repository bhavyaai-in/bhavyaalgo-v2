package service

import (
	"crypto/rand"
	"database/sql"
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
	db     *sql.DB
}

func NewSessionStore(db *sql.DB, ttl time.Duration) *SessionStore {
	s := &SessionStore{
		tokens: make(map[string]sessionEntry),
		ttl:    ttl,
		db:     db,
	}

	if db != nil {
		rows, err := db.Query("SELECT token, email, created_at FROM sessions")
		if err != nil {
			log.Printf("SessionStore: failed to load sessions from DB: %v", err)
		} else {
			defer rows.Close()
			for rows.Next() {
				var token, email, createdAtStr string
				if err := rows.Scan(&token, &email, &createdAtStr); err != nil {
					log.Printf("SessionStore: failed to scan session: %v", err)
					continue
				}
				createdAt, err := time.Parse("2006-01-02 15:04:05", createdAtStr)
				if err != nil {
					log.Printf("SessionStore: failed to parse created_at for token %s: %v", token, err)
					continue
				}
				if time.Since(createdAt) > ttl {
					_, _ = db.Exec("DELETE FROM sessions WHERE token = ?", token)
				} else {
					s.tokens[token] = sessionEntry{email: email, createdAt: createdAt}
				}
			}
		}
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
	createdAt := time.Now()

	s.mu.Lock()
	s.tokens[token] = sessionEntry{email: email, createdAt: createdAt}
	s.mu.Unlock()

	if s.db != nil {
		createdAtStr := createdAt.Format("2006-01-02 15:04:05")
		_, err := s.db.Exec("INSERT OR REPLACE INTO sessions (token, email, created_at) VALUES (?, ?, ?)", token, email, createdAtStr)
		if err != nil {
			log.Printf("SessionStore: failed to insert session in DB: %v", err)
		}
	}
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

	if s.db != nil {
		_, err := s.db.Exec("DELETE FROM sessions WHERE token = ?", token)
		if err != nil {
			log.Printf("SessionStore: failed to delete session from DB: %v", err)
		}
	}
}

func (s *SessionStore) cleanup() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		s.mu.Lock()
		for token, entry := range s.tokens {
			if time.Since(entry.createdAt) > s.ttl {
				delete(s.tokens, token)
				if s.db != nil {
					_, _ = s.db.Exec("DELETE FROM sessions WHERE token = ?", token)
				}
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

