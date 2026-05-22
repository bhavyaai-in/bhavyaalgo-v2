package server

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"bhavyaaialgo/backend/db/gen"
	"golang.org/x/crypto/bcrypt"
)

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

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var creds struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	if creds.Email != s.Config.AdminEmail {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	if err := bcrypt.CompareHashAndPassword(s.adminPasswordHash, []byte(creds.Password)); err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	token, err := s.Sessions.Create(creds.Email)
	if err != nil {
		log.Printf("session create error: %v", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if token == "" {
		writeError(w, http.StatusUnauthorized, "missing token")
		return
	}
	email, ok := s.Sessions.Get(token)
	if !ok {
		writeError(w, http.StatusUnauthorized, "invalid token")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"email": email})
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if token != "" {
		s.Sessions.Delete(token)
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "logged out"})
}

func (s *Server) handleListBrokers(w http.ResponseWriter, r *http.Request) {
	brokers, err := s.Q.ListBrokers(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	resp := make([]BrokerResponse, 0, len(brokers))
	for _, b := range brokers {
		resp = append(resp, brokerToResponse(b))
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleCreateBroker(w http.ResponseWriter, r *http.Request) {
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
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	active := boolToInt64(req.IsActive)
	autologin := boolToInt64(req.IsAutologin)
	id, err := s.Q.CreateBroker(r.Context(), gen.CreateBrokerParams{
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
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	b, err := s.Q.GetBroker(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, brokerToResponse(b))
}

func (s *Server) handleGetBroker(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	b, err := s.Q.GetBroker(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	writeJSON(w, http.StatusOK, brokerToResponse(b))
}

func (s *Server) handleUpdateBroker(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
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
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	existing, err := s.Q.GetBroker(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	err = s.Q.UpdateBroker(r.Context(), gen.UpdateBrokerParams{
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
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	updated, _ := s.Q.GetBroker(r.Context(), id)
	writeJSON(w, http.StatusOK, brokerToResponse(updated))
}

func (s *Server) handleDeleteBroker(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := s.Q.DeleteBroker(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (s *Server) handleListBrokerList(w http.ResponseWriter, r *http.Request) {
	entries, err := s.Q.ListBrokerList(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
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

func (s *Server) handleBrokerColumns(w http.ResponseWriter, r *http.Request) {
	rows, err := s.Q.ListBrokerColumns(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
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

func (s *Server) handleVersion(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"version":   "0.1.0",
		"build":     "unknown",
		"goVersion": "1.26",
	})
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
	parsed, err := time.Parse("2006-01-02 15:04:05", dateStr)
	if err != nil {
		return false
	}
	now := time.Now()
	return parsed.Year() == now.Year() && parsed.YearDay() == now.YearDay()
}

func boolToInt64(v bool) int64 {
	if v {
		return 1
	}
	return 0
}
