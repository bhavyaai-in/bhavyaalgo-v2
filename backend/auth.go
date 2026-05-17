package main

import (
	"net/http"

	"bhavyaaialgo/backend/internal/httperr"
	"bhavyaaialgo/backend/internal/service"
)

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			httperr.WriteMsg(w, 401, "missing token")
			return
		}
		_, ok := service.Sessions.Get(token)
		if !ok {
			httperr.WriteMsg(w, 401, "invalid token")
			return
		}
		next(w, r)
	}
}
