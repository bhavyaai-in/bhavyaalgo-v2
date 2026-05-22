package blueprints

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"runtime/debug"

	"bhavyaaialgo/backend/db/gen"
	"bhavyaaialgo/backend/internal/httperr"
	"bhavyaaialgo/backend/ws"
)

type SessionStore interface {
	Get(token string) (string, bool)
}

type App struct {
	DB       *sql.DB
	Q        *gen.Queries
	Sessions SessionStore
	Hub      *ws.Hub
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	httperr.WriteJSON(w, status, v)
}

func (a *App) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing token"})
			return
		}
		_, ok := a.Sessions.Get(token)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid token"})
			return
		}
		next(w, r)
	}
}

func decodeJSON(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}

func recoverHandler(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("PANIC in %s %s: %v\n%s", r.Method, r.URL.Path, rec, debug.Stack())
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
			}
		}()
		next(w, r)
	}
}
