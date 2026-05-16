package blueprints

import (
	"net/http"
	"os"
	"strconv"

	"bhavyaaialgo/backend/angel"
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

	var authToken, brokerName, brokerAPI string
	err = a.DB.QueryRow(
		`SELECT broker_token, broker_name, broker_api FROM brokers WHERE id = ?`, id,
	).Scan(&authToken, &brokerName, &brokerAPI)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "broker not found"})
		return
	}
	if authToken == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "broker not connected"})
		return
	}

	apiKey := getAPIKey(brokerAPI)

	switch brokerName {
	case "angel":
		client := angel.NewClient(apiKey)
		data, err := client.GetProfileRaw(authToken)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, data)
	default:
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unsupported broker: " + brokerName})
	}
}

func getAPIKey(brokerAPI string) string {
	apiKey := os.Getenv("BROKER_API_KEY")
	if apiKey == "" {
		apiKey = brokerAPI
	}
	return apiKey
}
