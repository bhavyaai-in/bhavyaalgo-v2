package main

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
)

func handleLogin(email, password string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var creds struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
			return
		}
		if creds.Email != email || creds.Password != password {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
			return
		}
		token := sessions.Create(creds.Email)
		writeJSON(w, http.StatusOK, map[string]string{"token": token})
	}
}

func handleMe(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if token == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing token"})
		return
	}
	email, ok := sessions.Get(token)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid token"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"email": email})
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if token != "" {
		sessions.Delete(token)
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "logged out"})
}

func handleListBrokers(w http.ResponseWriter, r *http.Request) {
	brokers, err := listBrokers()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if brokers == nil {
		brokers = []Broker{}
	}
	writeJSON(w, http.StatusOK, brokers)
}

func handleCreateBroker(w http.ResponseWriter, r *http.Request) {
	var b Broker
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	id, err := createBroker(&b)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	created, err := getBroker(id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func handleGetBroker(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	b, err := getBroker(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}
	writeJSON(w, http.StatusOK, b)
}

func handleUpdateBroker(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	var b Broker
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	if err := updateBroker(id, &b); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	updated, err := getBroker(id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func handleBrokerColumns(w http.ResponseWriter, r *http.Request) {
	data, err := os.ReadFile("broker_columns.json")
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "columns not found"})
		return
	}
	var columns map[string][]string
	if err := json.Unmarshal(data, &columns); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "invalid columns file"})
		return
	}
	writeJSON(w, http.StatusOK, columns)
}

func handleListBrokerList(w http.ResponseWriter, r *http.Request) {
	entries, err := listBrokerList()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if entries == nil {
		entries = []BrokerListEntry{}
	}
	writeJSON(w, http.StatusOK, entries)
}

func handleDeleteBroker(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	if err := deleteBroker(id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
