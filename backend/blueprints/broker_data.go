package blueprints

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"bhavyaaialgo/backend/brokers/aliceblue"
	"bhavyaaialgo/backend/brokers/angel"
)

var errNoToken = errors.New("broker token not generated")

func (a *App) RegisterBrokerDataRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/broker-orders/{id}", a.authMiddleware(a.handleBrokerOrders))
	mux.HandleFunc("GET /api/broker-positions/{id}", a.authMiddleware(a.handleBrokerPositions))
	mux.HandleFunc("GET /api/broker-holdings/{id}", a.authMiddleware(a.handleBrokerHoldings))
	mux.HandleFunc("GET /api/broker-margin/{id}", a.authMiddleware(a.handleBrokerMargin))
	mux.HandleFunc("POST /api/broker-cancel-order", a.authMiddleware(a.handleBrokerCancelOrder))
	mux.HandleFunc("POST /api/broker-modify-order", a.authMiddleware(a.handleBrokerModifyOrder))
	mux.HandleFunc("POST /api/broker-place-order", a.authMiddleware(a.handleBrokerPlaceOrder))
}

type brokerClientResult struct {
	angelClient *angel.Client
	aliceClient *aliceblue.Client
	authToken   string
	brokerName  string
	tokenStatus string
}

func (a *App) brokerClient(id int64) (*brokerClientResult, error) {
	var authToken, brokerName, brokerAPI, brokerAPISecret, tokenStatus string
	err := a.TradingDB.QueryRow(
		`SELECT broker_token, broker_name, broker_api, broker_api_secret, token_status FROM brokers WHERE id = ?`, id,
	).Scan(&authToken, &brokerName, &brokerAPI, &brokerAPISecret, &tokenStatus)
	if err != nil {
		return nil, err
	}
	if authToken == "" {
		return nil, errNoToken
	}
	apiKey := getAPIKey(brokerAPI)
	apiSecret := getAPISecret(brokerAPISecret)
	result := &brokerClientResult{
		authToken:   authToken,
		brokerName:  brokerName,
		tokenStatus: tokenStatus,
	}
	switch brokerName {
	case "angel":
		result.angelClient = angel.NewClient(apiKey)
	case "aliceblue":
		result.aliceClient = aliceblue.NewClient(apiKey, apiSecret)
	}
	return result, nil
}

func (bc *brokerClientResult) apiError() string {
	if bc.tokenStatus == "connected" {
		return "could not connect broker"
	}
	return "broker token not generated"
}

func (a *App) handleBrokerOrders(w http.ResponseWriter, r *http.Request) {
	id := parseID(r)
	if id == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	bc, err := a.brokerClient(id)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	switch bc.brokerName {
	case "angel":
		data, err := bc.angelClient.GetOrderBook(bc.authToken)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": bc.apiError()})
			return
		}
		writeJSON(w, http.StatusOK, data)
	case "aliceblue":
		data, err := bc.aliceClient.GetOrderBook(bc.authToken)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": bc.apiError()})
			return
		}
		writeJSON(w, http.StatusOK, normalizeOrders(data))
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
	bc, err := a.brokerClient(id)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	switch bc.brokerName {
	case "angel":
		data, err := bc.angelClient.GetPositions(bc.authToken)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": bc.apiError()})
			return
		}
		writeJSON(w, http.StatusOK, data)
	case "aliceblue":
		data, err := bc.aliceClient.GetPositions(bc.authToken)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": bc.apiError()})
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
	bc, err := a.brokerClient(id)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	switch bc.brokerName {
	case "angel":
		data, err := bc.angelClient.DoGet(angel.HoldingURL, bc.authToken)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": bc.apiError()})
			return
		}
		writeJSON(w, http.StatusOK, data)
	case "aliceblue":
		data, err := bc.aliceClient.GetHoldings(bc.authToken, "cnc")
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": bc.apiError()})
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
	bc, err := a.brokerClient(id)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	switch bc.brokerName {
	case "angel":
		data, err := bc.angelClient.GetMargin(bc.authToken)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": bc.apiError()})
			return
		}
		writeJSON(w, http.StatusOK, data)
	case "aliceblue":
		data, err := bc.aliceClient.GetMargin(bc.authToken)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": bc.apiError()})
			return
		}
		writeJSON(w, http.StatusOK, data)
	default:
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unsupported"})
	}
}

func normalizeOrders(orders []map[string]any) []map[string]any {
	fieldMap := map[string]string{
		"brokerOrderId":           "orderid",
		"tradingSymbol":           "tradingsymbol",
		"formattedInstrumentName": "tradingsymbol",
		"transactionType":         "transactiontype",
		"filledQuantity":          "filledshares",
		"orderType":               "ordertype",
		"product":                 "producttype",
		"orderStatus":             "orderstatus",
		"rejectionReason":         "text",
		"instrumentId":            "instrumenttoken",
		"slTriggerPrice":          "triggerprice",
		"requestTime":             "ordertime",
		"orderDate":               "ordertime",
		"orderDateTime":           "ordertime",
		"timestamp":               "ordertime",
		"createdAt":               "ordertime",
		"created_at":              "ordertime",
		"dateTime":                "ordertime",
	}
	for _, order := range orders {
		for oldKey, newKey := range fieldMap {
			if v, ok := order[oldKey]; ok {
				if _, exists := order[newKey]; !exists {
					order[newKey] = v
				}
			}
		}
	}
	return orders
}

func checkAliceError(result map[string]any) error {
	if result == nil {
		return nil
	}
	if status, _ := result["status"].(string); strings.EqualFold(status, "Not_Ok") || strings.EqualFold(status, "Error") || strings.EqualFold(status, "Rejected") {
		msg, _ := result["emsg"].(string)
		if msg == "" {
			msg, _ = result["message"].(string)
		}
		if msg != "" {
			return fmt.Errorf("%s", msg)
		}
		return fmt.Errorf("request failed")
	}
	if results, ok := result["result"].([]any); ok && len(results) > 0 {
		if first, ok := results[0].(map[string]any); ok {
			if msg, ok := first["message"].(string); ok && msg != "" {
				if brokerOrderID, _ := first["brokerOrderId"].(string); brokerOrderID == "" {
					return fmt.Errorf("%s", msg)
				}
			}
		}
	}
	return nil
}

func checkAngelError(result map[string]any) error {
	if result == nil {
		return nil
	}
	if code, _ := result["errorcode"].(string); code != "" {
		msg, _ := result["message"].(string)
		return fmt.Errorf("%s: %s", code, msg)
	}
	switch status := result["status"].(type) {
	case bool:
		if !status {
			msg, _ := result["message"].(string)
			if msg == "" {
				msg = "request failed"
			}
			return fmt.Errorf("%s", msg)
		}
	case string:
		if strings.EqualFold(status, "false") || strings.EqualFold(status, "error") || strings.EqualFold(status, "not_ok") {
			msg, _ := result["message"].(string)
			if msg == "" {
				msg = "request failed"
			}
			return fmt.Errorf("%s", msg)
		}
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
	bc, err := a.brokerClient(req.BrokerID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	switch bc.brokerName {
	case "angel":
		result, err := bc.angelClient.PlaceOrder(bc.authToken, req.Data)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
			return
		}
		if err := checkAngelError(result); err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, result)
	case "aliceblue":
		result, err := bc.aliceClient.PlaceOrder(bc.authToken, req.Data)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
			return
		}
		if err := checkAliceError(result); err != nil {
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
	req.Data["brokerOrderId"] = req.OrderID
	bc, err := a.brokerClient(req.BrokerID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	switch bc.brokerName {
	case "angel":
		req.Data["orderid"] = req.OrderID
		result, err := bc.angelClient.ModifyOrder(bc.authToken, req.Data)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
			return
		}
		if err := checkAngelError(result); err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, result)
	case "aliceblue":
		result, err := bc.aliceClient.ModifyOrder(bc.authToken, req.Data)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
			return
		}
		if err := checkAliceError(result); err != nil {
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
	bc, err := a.brokerClient(req.BrokerID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	switch bc.brokerName {
	case "angel":
		result, err := bc.angelClient.CancelOrder(bc.authToken, req.OrderID)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
			return
		}
		if err := checkAngelError(result); err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, result)
	case "aliceblue":
		result, err := bc.aliceClient.CancelOrder(bc.authToken, req.OrderID)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
			return
		}
		if err := checkAliceError(result); err != nil {
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
