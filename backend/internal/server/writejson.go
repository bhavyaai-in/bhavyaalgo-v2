package server

import (
	"net/http"

	"bhavyaaialgo/backend/internal/httperr"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	httperr.WriteJSON(w, status, v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	httperr.WriteError(w, status, msg)
}
