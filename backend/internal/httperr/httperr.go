package httperr

import (
	"encoding/json"
	"log"
	"net/http"
)

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"-"`
}

func (e *Error) Error() string {
	return e.Code + ": " + e.Message
}

func Write(w http.ResponseWriter, err *Error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.Status)
	json.NewEncoder(w).Encode(err)
}

func WriteMsg(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("json encode error: %v", err)
	}
}

func WriteError(w http.ResponseWriter, status int, msg string) {
	WriteJSON(w, status, map[string]string{"error": msg})
}

var (
	InvalidRequest     = &Error{Code: "INVALID_REQUEST", Message: "invalid request", Status: 400}
	NotFound           = &Error{Code: "NOT_FOUND", Message: "resource not found", Status: 404}
	Unauthorized       = &Error{Code: "UNAUTHORIZED", Message: "unauthorized", Status: 401}
	InvalidCreds       = &Error{Code: "INVALID_CREDS", Message: "invalid credentials", Status: 401}
	InternalError      = &Error{Code: "INTERNAL_ERROR", Message: "internal server error", Status: 500}
	SessionExpired     = &Error{Code: "SESSION_EXPIRED", Message: "session expired", Status: 401}
	UnsupportedBroker  = &Error{Code: "UNSUPPORTED_BROKER", Message: "unsupported broker", Status: 400}
	BrokerNotConnected = &Error{Code: "BROKER_NOT_CONNECTED", Message: "broker not connected", Status: 400}
)
