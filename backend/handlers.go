package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"bhavyaaialgo/backend/db/gen"
	"bhavyaaialgo/backend/internal/service"
)

// BrokerResponse is the public API response (token fields excluded)
type BrokerResponse struct {
	ID              int64  `json:"id"`
	FriendlyName    string `json:"friendly_name"`
	BrokerUserid    string `json:"broker_userid"`
	BrokerPassword  string `json:"broker_password"`
	BrokerPin       string `json:"broker_pin"`
	BrokerQrKey     string `json:"broker_qr_key"`
	BrokerAPI       string `json:"broker_api"`
	BrokerAPISecret string `json:"broker_api_secret"`
	BrokerName      string `json:"broker_name"`
	TokenStatus     string `json:"token_status"`
	BrokerTokenDate string `json:"broker_token_date"`
	IsActive        bool   `json:"is_active"`
	IsAutologin     bool   `json:"is_autologin"`
	IsDisabled      bool   `json:"is_disabled"`
	FeedToken       string `json:"-"`
	BrokerToken     string `json:"-"`
	Message         string `json:"message"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

func brokerToResponse(b gen.Broker) BrokerResponse {
	tokenStatus := b.TokenStatus
	if tokenStatus == "connected" && !isTokenDateToday(b.BrokerTokenDate) {
		tokenStatus = "expired"
	}
	return BrokerResponse{
		ID:              b.ID,
		FriendlyName:    b.FriendlyName,
		BrokerUserid:    b.BrokerUserid,
		BrokerPassword:  b.BrokerPassword,
		BrokerPin:       b.BrokerPin,
		BrokerQrKey:     b.BrokerQrKey,
		BrokerAPI:       b.BrokerApi,
		BrokerAPISecret: b.BrokerApiSecret,
		BrokerName:      b.BrokerName,
		TokenStatus:     tokenStatus,
		BrokerTokenDate: b.BrokerTokenDate,
		IsActive:        b.IsActive != 0,
		IsAutologin:     b.IsAutologin != 0,
		IsDisabled:      b.IsDisabled != 0,
		Message:         b.Message,
		FeedToken:       b.FeedToken,
		BrokerToken:     b.BrokerToken,
		CreatedAt:       b.CreatedAt,
		UpdatedAt:       b.UpdatedAt,
	}
}

func isTokenDateToday(dateStr string) bool {
	dateStr = strings.TrimSpace(dateStr)
	if dateStr == "" {
		return false
	}
	// Expected format: "2026-05-19 13:46:52"
	parsed, err := time.Parse("2006-01-02 15:04:05", dateStr)
	if err != nil {
		return false
	}
	now := time.Now()
	return parsed.Year() == now.Year() && parsed.YearDay() == now.YearDay()
}

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
		token := service.Sessions.Create(creds.Email)
		writeJSON(w, http.StatusOK, map[string]string{"token": token})
	}
}

func handleMe(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if token == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing token"})
		return
	}
	email, ok := service.Sessions.Get(token)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid token"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"email": email})
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if token != "" {
		service.Sessions.Delete(token)
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "logged out"})
}

func handleListBrokers(w http.ResponseWriter, r *http.Request) {
	brokers, err := Q.ListBrokers(context.Background())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	resp := make([]BrokerResponse, 0, len(brokers))
	for _, b := range brokers {
		resp = append(resp, brokerToResponse(b))
	}
	writeJSON(w, http.StatusOK, resp)
}

func handleCreateBroker(w http.ResponseWriter, r *http.Request) {
	var req struct {
		FriendlyName    string `json:"friendly_name"`
		BrokerUserid    string `json:"broker_userid"`
		BrokerPassword  string `json:"broker_password"`
		BrokerPin       string `json:"broker_pin"`
		BrokerQrKey     string `json:"broker_qr_key"`
		BrokerAPI       string `json:"broker_api"`
		BrokerAPISecret string `json:"broker_api_secret"`
		BrokerName      string `json:"broker_name"`
		IsActive        bool   `json:"is_active"`
		IsAutologin     bool   `json:"is_autologin"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	active := int64(0)
	if req.IsActive {
		active = 1
	}
	autologin := int64(0)
	if req.IsAutologin {
		autologin = 1
	}
	id, err := Q.CreateBroker(context.Background(), gen.CreateBrokerParams{
		FriendlyName:    req.FriendlyName,
		BrokerUserid:    req.BrokerUserid,
		BrokerPassword:  req.BrokerPassword,
		BrokerPin:       req.BrokerPin,
		BrokerQrKey:     req.BrokerQrKey,
		BrokerApi:       req.BrokerAPI,
		BrokerApiSecret: req.BrokerAPISecret,
		BrokerName:      req.BrokerName,
		IsActive:        active,
		IsAutologin:     autologin,
		TokenStatus:     "",
		BrokerToken:     "",
		BrokerTokenDate: "2000-01-01 00:00:00",
		FeedToken:       "",
		IsDisabled:      0,
		Message:         "",
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	b, err := Q.GetBroker(context.Background(), id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, brokerToResponse(b))
}

func handleGetBroker(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	b, err := Q.GetBroker(context.Background(), id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}
	writeJSON(w, http.StatusOK, brokerToResponse(b))
}

func handleUpdateBroker(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	var req struct {
		FriendlyName    string `json:"friendly_name"`
		BrokerUserid    string `json:"broker_userid"`
		BrokerPassword  string `json:"broker_password"`
		BrokerPin       string `json:"broker_pin"`
		BrokerQrKey     string `json:"broker_qr_key"`
		BrokerAPI       string `json:"broker_api"`
		BrokerAPISecret string `json:"broker_api_secret"`
		BrokerName      string `json:"broker_name"`
		IsActive        bool   `json:"is_active"`
		IsAutologin     bool   `json:"is_autologin"`
		IsDisabled      bool   `json:"is_disabled"`
		Message         string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	existing, err := Q.GetBroker(context.Background(), id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}
	err = Q.UpdateBroker(context.Background(), gen.UpdateBrokerParams{
		FriendlyName:    req.FriendlyName,
		BrokerUserid:    req.BrokerUserid,
		BrokerPassword:  req.BrokerPassword,
		BrokerPin:       req.BrokerPin,
		BrokerQrKey:     req.BrokerQrKey,
		BrokerApi:       req.BrokerAPI,
		BrokerApiSecret: req.BrokerAPISecret,
		BrokerName:      req.BrokerName,
		TokenStatus:     existing.TokenStatus,
		BrokerToken:     existing.BrokerToken,
		BrokerTokenDate: existing.BrokerTokenDate,
		FeedToken:       existing.FeedToken,
		IsActive:        boolToInt64(req.IsActive),
		IsAutologin:     boolToInt64(req.IsAutologin),
		IsDisabled:      boolToInt64(req.IsDisabled),
		Message:         req.Message,
		ID:              id,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	updated, _ := Q.GetBroker(context.Background(), id)
	writeJSON(w, http.StatusOK, brokerToResponse(updated))
}

func handleDeleteBroker(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	if err := Q.DeleteBroker(context.Background(), id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func handleListBrokerList(w http.ResponseWriter, r *http.Request) {
	entries, err := Q.ListBrokerList(context.Background())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	type listEntry struct {
		ID             int64  `json:"id"`
		Name           string `json:"broker_name"`
		BrokerImageURL string `json:"broker_image_url"`
		IsActive       bool   `json:"is_active"`
		Message        string `json:"message"`
		CreatedAt      string `json:"created_at"`
		UpdatedAt      string `json:"updated_at"`
	}
	resp := make([]listEntry, 0, len(entries))
	for _, e := range entries {
		resp = append(resp, listEntry{
			ID:             e.ID,
			Name:           e.Name,
			BrokerImageURL: e.BrokerImageUrl,
			IsActive:       e.IsActive != 0,
			Message:        e.Message,
			CreatedAt:      e.CreatedAt,
			UpdatedAt:      e.UpdatedAt,
		})
	}
	writeJSON(w, http.StatusOK, resp)
}

func handleBrokerColumns(w http.ResponseWriter, r *http.Request) {
	rows, err := Q.ListBrokerColumns(context.Background())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	columns := make(map[string][]string, len(rows))
	for _, row := range rows {
		var cols []string
		if err := json.Unmarshal([]byte(row.ColumnsJson), &cols); err == nil {
			columns[row.BrokerName] = cols
		}
	}
	if columns == nil {
		columns = map[string][]string{}
	}
	writeJSON(w, http.StatusOK, columns)
}

func boolToInt64(v bool) int64 {
	if v {
		return 1
	}
	return 0
}
