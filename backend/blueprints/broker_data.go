package blueprints

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"bhavyaaialgo/backend/angel"
)

var errNoToken = errors.New("broker not connected")

func (a *App) RegisterBrokerDataRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/broker-orders/{id}", a.authMiddleware(a.handleBrokerOrders))
	mux.HandleFunc("GET /api/broker-positions/{id}", a.authMiddleware(a.handleBrokerPositions))
	mux.HandleFunc("GET /api/broker-holdings/{id}", a.authMiddleware(a.handleBrokerHoldings))
	mux.HandleFunc("GET /api/broker-margin/{id}", a.authMiddleware(a.handleBrokerMargin))
	mux.HandleFunc("POST /api/broker-cancel-order", a.authMiddleware(a.handleBrokerCancelOrder))
	mux.HandleFunc("POST /api/broker-modify-order", a.authMiddleware(a.handleBrokerModifyOrder))
	mux.HandleFunc("POST /api/broker-place-order", a.authMiddleware(a.handleBrokerPlaceOrder))
}

func (a *App) brokerClient(id int64) (*angel.Client, string, string, error) {
	var authToken, brokerName, brokerAPI string
	err := a.DB.QueryRow(
		`SELECT broker_token, broker_name, broker_api FROM brokers WHERE id = ?`, id,
	).Scan(&authToken, &brokerName, &brokerAPI)
	if err != nil {
		return nil, "", "", err
	}
	if authToken == "" {
		return nil, "", "", errNoToken
	}
	apiKey := getAPIKey(brokerAPI)
	client := angel.NewClient(apiKey)
	return client, authToken, brokerName, nil
}

func (a *App) handleBrokerOrders(w http.ResponseWriter, r *http.Request) {
	id := parseID(r)
	if id == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	client, token, name, err := a.brokerClient(id)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	switch name {
	case "angel":
		data, err := client.GetOrderBook(token)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, data)
	default:
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unsupported"})
	}
}

func (a *App) handleBrokerPositions(w http.ResponseWriter, r *http.Request) {
	id := parseID(r)
	if id == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	client, token, name, err := a.brokerClient(id)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	switch name {
	case "angel":
		data, err := client.GetPositions(token)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, data)
	default:
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unsupported"})
	}
}

func (a *App) handleBrokerHoldings(w http.ResponseWriter, r *http.Request) {
	id := parseID(r)
	if id == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	client, token, name, err := a.brokerClient(id)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	switch name {
	case "angel":
		// Holdings may return data as an object, not an array
		data, err := client.DoGet(angel.HoldingURL, token)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, data)
	default:
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unsupported"})
	}
}

func (a *App) handleBrokerMargin(w http.ResponseWriter, r *http.Request) {
	id := parseID(r)
	if id == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	client, token, name, err := a.brokerClient(id)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	switch name {
	case "angel":
		data, err := client.GetMargin(token)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, data)
	default:
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unsupported"})
	}
}

func checkAngelError(result map[string]any) error {
	if result == nil {
		return nil
	}
	if code, _ := result["errorcode"].(string); code != "" {
		msg, _ := result["message"].(string)
		return fmt.Errorf("%s: %s", code, msg)
	}
	return nil
}

func (a *App) handleBrokerPlaceOrder(w http.ResponseWriter, r *http.Request) {
	var req struct {
		BrokerID int64          `json:"broker_id"`
		Data     map[string]any `json:"data"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	client, token, name, err := a.brokerClient(req.BrokerID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	switch name {
	case "angel":
		result, err := client.PlaceOrder(token, req.Data)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
			return
		}
		if err := checkAngelError(result); err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, result)
	default:
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unsupported broker"})
	}
}

func (a *App) handleBrokerModifyOrder(w http.ResponseWriter, r *http.Request) {
	var req struct {
		BrokerID int64          `json:"broker_id"`
		OrderID  string         `json:"order_id"`
		Data     map[string]any `json:"data"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	req.Data["orderid"] = req.OrderID
	client, token, name, err := a.brokerClient(req.BrokerID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	switch name {
	case "angel":
		result, err := client.ModifyOrder(token, req.Data)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
			return
		}
		if err := checkAngelError(result); err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, result)
	default:
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unsupported broker"})
	}
}

func (a *App) handleBrokerCancelOrder(w http.ResponseWriter, r *http.Request) {
	var req struct {
		BrokerID int64  `json:"broker_id"`
		OrderID  string `json:"order_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	client, token, name, err := a.brokerClient(req.BrokerID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	switch name {
	case "angel":
		result, err := client.CancelOrder(token, req.OrderID)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
			return
		}
		if err := checkAngelError(result); err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, result)
	default:
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unsupported broker"})
	}
}

func parseID(r *http.Request) int64 {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0
	}
	return id
}
