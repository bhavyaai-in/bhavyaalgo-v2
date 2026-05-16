package blueprints

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"bhavyaaialgo/backend/angel"

	"github.com/xlzd/gotp"
)

type connectRequest struct {
	BrokerID int64 `json:"broker_id"`
}

type brokerAuthRow struct {
	ID              int64
	BrokerUserid    string
	BrokerPassword  string
	BrokerPin       string
	BrokerQrKey     string
	BrokerName      string
	BrokerAPI       string
	BrokerAPISecret string
}

func (a *App) RegisterConnectBrokerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/connect-broker", recoverHandler(a.authMiddleware(a.handleConnectBroker)))
}

func (a *App) handleConnectBroker(w http.ResponseWriter, r *http.Request) {
	var req connectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	var b brokerAuthRow
	err := a.DB.QueryRow(`
		SELECT id, broker_userid, broker_password, broker_pin, broker_qr_key,
		       broker_name, broker_api, broker_api_secret
		FROM brokers WHERE id = ?
	`, req.BrokerID).Scan(&b.ID, &b.BrokerUserid, &b.BrokerPassword, &b.BrokerPin,
		&b.BrokerQrKey, &b.BrokerName, &b.BrokerAPI, &b.BrokerAPISecret)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "broker not found"})
		return
	}

	totp := generateTOTP(b.BrokerQrKey)
	if totp == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "TOTP secret not configured in QR Key"})
		return
	}

	var authToken, feedToken, profileName string
	switch b.BrokerName {
	case "angel":
		authToken, feedToken, _, err = authenticateAngel(b, totp)
		if err == nil {
			profileName, err = fetchProfile(b.BrokerName, authToken, b.BrokerAPI)
		}
	default:
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unsupported broker: " + b.BrokerName})
		return
	}
	if err != nil {
		a.DB.Exec(`UPDATE brokers SET token_status=?, message=? WHERE id=?`,
			"error", err.Error(), b.ID)
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}

	a.DB.Exec(`
		UPDATE brokers SET
			broker_token=?, feed_token=?, token_status=?, broker_token_date=datetime('now','localtime'), message=?
		WHERE id=?
	`, authToken, feedToken, "connected", profileName, b.ID)

	writeJSON(w, http.StatusOK, map[string]any{
		"status":       "connected",
		"auth_token":   authToken,
		"feed_token":   feedToken,
		"profile_name": profileName,
	})
}

func generateTOTP(secret string) (code string) {
	if secret == "" {
		return ""
	}
	// Handle otpauth:// URL format
	if strings.HasPrefix(secret, "otpauth://") {
		for _, part := range strings.Split(secret, "&") {
			if strings.HasPrefix(part, "secret=") {
				secret = strings.TrimPrefix(part, "secret=")
				break
			}
			if strings.Contains(part, "?secret=") {
				secret = strings.Split(part, "?secret=")[1]
				break
			}
		}
	}
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return ""
	}
	defer func() {
		if rec := recover(); rec != nil {
			log.Printf("TOTP generation panic for secret ending in %q: %v", safeSuffix(secret, 4), rec)
			code = ""
		}
	}()
	return gotp.NewDefaultTOTP(secret).Now()
}

func safeSuffix(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return "..." + s[len(s)-n:]
}

func authenticateAngel(b brokerAuthRow, totp string) (string, string, string, error) {
	apiKey := os.Getenv("BROKER_API_KEY")
	if apiKey == "" {
		apiKey = b.BrokerAPI
	}
	client := angel.NewClient(apiKey)
	return client.Authenticate(b.BrokerUserid, b.BrokerPin, totp)
}

func fetchProfile(brokerName, authToken, brokerAPIKey string) (string, error) {
	apiKey := os.Getenv("BROKER_API_KEY")
	if apiKey == "" {
		apiKey = brokerAPIKey
	}
	switch brokerName {
	case "angel":
		client := angel.NewClient(apiKey)
		return client.GetProfile(authToken)
	default:
		return "", nil
	}
}
