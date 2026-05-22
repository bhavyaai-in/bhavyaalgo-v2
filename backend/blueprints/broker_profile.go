package blueprints

import (
	"net/http"
	"os"
	"strconv"

	"bhavyaaialgo/backend/brokers/aliceblue"
	"bhavyaaialgo/backend/brokers/angel"
)

func (a *App) RegisterBrokerProfileRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/broker-profile/{id}", a.authMiddleware(a.handleBrokerProfile))
}

func (a *App) handleBrokerProfile(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	var authToken, brokerName, brokerAPI, brokerAPISecret string
	err = a.DB.QueryRow(
		`SELECT broker_token, broker_name, broker_api, broker_api_secret FROM brokers WHERE id = ?`, id,
	).Scan(&authToken, &brokerName, &brokerAPI, &brokerAPISecret)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "broker not found"})
		return
	}
	if authToken == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "broker not connected"})
		return
	}

	apiKey := getAPIKey(brokerAPI)
	apiSecret := getAPISecret(brokerAPISecret)

	var data map[string]any
	switch brokerName {
	case "angel":
		data, err = angel.FetchProfileRaw(authToken, apiKey)
	case "aliceblue":
		data, err = aliceblue.FetchProfileRaw(authToken, brokerAPI, apiSecret)
	default:
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unsupported broker: " + brokerName})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func getAPIKey(brokerAPI string) string {
	apiKey := os.Getenv("BROKER_API_KEY")
	if apiKey == "" {
		apiKey = brokerAPI
	}
	return apiKey
}

func getAPISecret(brokerAPISecret string) string {
	apiSecret := os.Getenv("BROKER_API_SECRET")
	if apiSecret == "" {
		apiSecret = brokerAPISecret
	}
	return apiSecret
}
