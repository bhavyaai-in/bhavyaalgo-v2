package blueprints

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"net/http"
	"strconv"

	"bhavyaaialgo/backend/db/gen"
)

func generateSecretKey() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (a *App) RegisterStrategyRoutes(mux *http.ServeMux) {
	// Strategy Types
	mux.HandleFunc("GET /api/strategy-types", a.authMiddleware(a.handleListStrategyTypes))
	mux.HandleFunc("POST /api/strategy-types", a.authMiddleware(a.handleCreateStrategyType))
	mux.HandleFunc("PUT /api/strategy-types/{id}", a.authMiddleware(a.handleUpdateStrategyType))
	mux.HandleFunc("DELETE /api/strategy-types/{id}", a.authMiddleware(a.handleDeleteStrategyType))

	// Strategies
	mux.HandleFunc("GET /api/strategies", a.authMiddleware(a.handleListStrategies))
	mux.HandleFunc("POST /api/strategies", a.authMiddleware(a.handleCreateStrategy))
	mux.HandleFunc("GET /api/strategies/{id}", a.authMiddleware(a.handleGetStrategy))
	mux.HandleFunc("PUT /api/strategies/{id}", a.authMiddleware(a.handleUpdateStrategy))
	mux.HandleFunc("DELETE /api/strategies/{id}", a.authMiddleware(a.handleDeleteStrategy))

	// Strategy Info
	mux.HandleFunc("GET /api/strategies/{id}/info", a.authMiddleware(a.handleListStrategyInfo))
	mux.HandleFunc("POST /api/strategies/{id}/info", a.authMiddleware(a.handleCreateStrategyInfo))
	mux.HandleFunc("DELETE /api/strategies/{id}/info/{infoId}", a.authMiddleware(a.handleDeleteStrategyInfo))

	// Strategy Joiners
	mux.HandleFunc("GET /api/strategies/{id}/joiners", a.authMiddleware(a.handleListStrategyJoiners))
	mux.HandleFunc("POST /api/strategies/{id}/joiners", a.authMiddleware(a.handleCreateStrategyJoiner))
	mux.HandleFunc("PUT /api/strategies/{id}/joiners/{joinerId}", a.authMiddleware(a.handleUpdateStrategyJoiner))
	mux.HandleFunc("DELETE /api/strategies/{id}/joiners/{joinerId}", a.authMiddleware(a.handleDeleteStrategyJoiner))

	// Strategy Positions & Orders
	mux.HandleFunc("GET /api/strategies/{id}/positions", a.authMiddleware(a.handleListStrategyPositions))
	mux.HandleFunc("GET /api/strategies/{id}/orders", a.authMiddleware(a.handleListStrategyOrders))

	// Strategy Place Order (multi-broker)
	mux.HandleFunc("POST /api/strategies/{id}/place-order", a.authMiddleware(a.handleStrategyPlaceOrder))
}

// ─── Strategy Types ────────────────────────────────────────────────────────────

func (a *App) handleListStrategyTypes(w http.ResponseWriter, r *http.Request) {
	types, err := a.Q.ListStrategyTypes(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if types == nil {
		types = []gen.StrategyType{}
	}
	writeJSON(w, http.StatusOK, types)
}

type strategyTypeInput struct {
	Name              string `json:"name"`
	RulesExplanation  string `json:"rules_explanation"`
}

func (a *App) handleCreateStrategyType(w http.ResponseWriter, r *http.Request) {
	var in strategyTypeInput
	if err := decodeJSON(r, &in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	id, err := a.Q.CreateStrategyType(r.Context(), gen.CreateStrategyTypeParams{
		Name:             in.Name,
		RulesExplanation: in.RulesExplanation,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]int64{"id": id})
}

func (a *App) handleUpdateStrategyType(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	var in strategyTypeInput
	if err := decodeJSON(r, &in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	if err := a.Q.UpdateStrategyType(r.Context(), gen.UpdateStrategyTypeParams{
		Name:             in.Name,
		RulesExplanation: in.RulesExplanation,
		ID:               id,
	}); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (a *App) handleDeleteStrategyType(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	if err := a.Q.DeleteStrategyType(r.Context(), id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// ─── Strategies ────────────────────────────────────────────────────────────────

type strategyInput struct {
	Name              string  `json:"name"`
	StrategySecretKey  string  `json:"strategy_secret_key"`
	StrategyType      int64   `json:"strategy_type"`
	PositionStatus    int64   `json:"position_status"`
	InstrumentToken   int64   `json:"instrument_token"`
	Exchange          string  `json:"exchange"`
	Side              string  `json:"side"`
	AtmOtm            float64 `json:"atm_otm"`
	ImageURL          string  `json:"image_url"`
	Color             string  `json:"color"`
	IsActive          int64   `json:"is_active"`
	IsLocked          int64   `json:"is_locked"`
	Message           string  `json:"message"`
	ExpiryDate        string  `json:"expiry_date"`
}

func (a *App) handleListStrategies(w http.ResponseWriter, r *http.Request) {
	list, err := a.Q.ListStrategies(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if list == nil {
		list = []gen.Strategy{}
	}
	writeJSON(w, http.StatusOK, list)
}

func (a *App) handleCreateStrategy(w http.ResponseWriter, r *http.Request) {
	var in strategyInput
	if err := decodeJSON(r, &in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	if in.StrategySecretKey == "" {
		in.StrategySecretKey = generateSecretKey()
	}
	id, err := a.Q.CreateStrategy(r.Context(), gen.CreateStrategyParams{
		Name:             in.Name,
		StrategySecretKey: in.StrategySecretKey,
		StrategyType:     in.StrategyType,
		PositionStatus:   in.PositionStatus,
		InstrumentToken:  in.InstrumentToken,
		Exchange:         in.Exchange,
		Side:             in.Side,
		AtmOtm:           in.AtmOtm,
		ImageUrl:         in.ImageURL,
		Color:            in.Color,
		IsActive:         in.IsActive,
		IsLocked:         in.IsLocked,
		Message:          in.Message,
		ExpiryDate:       in.ExpiryDate,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]int64{"id": id})
}

func (a *App) handleGetStrategy(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	s, err := a.Q.GetStrategy(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}
	writeJSON(w, http.StatusOK, s)
}

func (a *App) handleUpdateStrategy(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	var in strategyInput
	if err := decodeJSON(r, &in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	if in.StrategySecretKey == "" {
		existing, err := a.Q.GetStrategy(r.Context(), id)
		if err == nil {
			in.StrategySecretKey = existing.StrategySecretKey
		}
	}
	if err := a.Q.UpdateStrategy(r.Context(), gen.UpdateStrategyParams{
		Name:              in.Name,
		StrategySecretKey: in.StrategySecretKey,
		StrategyType:      in.StrategyType,
		PositionStatus:    in.PositionStatus,
		InstrumentToken:   in.InstrumentToken,
		Exchange:          in.Exchange,
		Side:              in.Side,
		AtmOtm:            in.AtmOtm,
		ImageUrl:      in.ImageURL,
		Color:         in.Color,
		IsActive:          in.IsActive,
		IsLocked:          in.IsLocked,
		Message:           in.Message,
		ExpiryDate:        in.ExpiryDate,
		ID:                id,
	}); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (a *App) handleDeleteStrategy(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	if err := a.Q.DeleteStrategy(r.Context(), id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// ─── Strategy Info ─────────────────────────────────────────────────────────────

type strategyInfoInput struct {
	Info string `json:"info"`
}

func (a *App) handleListStrategyInfo(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	list, err := a.Q.ListStrategyInfo(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if list == nil {
		list = []gen.StrategyInfo{}
	}
	writeJSON(w, http.StatusOK, list)
}

func (a *App) handleCreateStrategyInfo(w http.ResponseWriter, r *http.Request) {
	strategyID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	var in strategyInfoInput
	if err := decodeJSON(r, &in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	id, err := a.Q.CreateStrategyInfo(r.Context(), gen.CreateStrategyInfoParams{
		StrategyID: strategyID,
		Info:       in.Info,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]int64{"id": id})
}

func (a *App) handleDeleteStrategyInfo(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("infoId"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	if err := a.Q.DeleteStrategyInfo(r.Context(), id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// ─── Strategy Joiners ─────────────────────────────────────────────────────────

type strategyJoinerInput struct {
	BrokerID         int64   `json:"broker_id"`
	Quantity         float64 `json:"quantity"`
	ReEntry          int64   `json:"re_entry"`
	ReEntryTriggered int64   `json:"re_entry_triggered"`
	Multiplier       float64 `json:"multiplier"`
	IsActive         int64   `json:"is_active"`
}

func (a *App) handleListStrategyJoiners(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	list, err := a.Q.ListStrategyJoiners(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if list == nil {
		list = []gen.StrategyJoiner{}
	}
	writeJSON(w, http.StatusOK, list)
}

func (a *App) handleCreateStrategyJoiner(w http.ResponseWriter, r *http.Request) {
	strategyID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	var in strategyJoinerInput
	if err := decodeJSON(r, &in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	id, err := a.Q.CreateStrategyJoiner(r.Context(), gen.CreateStrategyJoinerParams{
		BrokerID:        in.BrokerID,
		StrategyID:      strategyID,
		Quantity:        in.Quantity,
		ReEntry:         in.ReEntry,
		ReEntryTriggered: in.ReEntryTriggered,
		Multiplier:      in.Multiplier,
		IsActive:        in.IsActive,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]int64{"id": id})
}

func (a *App) handleUpdateStrategyJoiner(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("joinerId"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	var in strategyJoinerInput
	if err := decodeJSON(r, &in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	if err := a.Q.UpdateStrategyJoiner(r.Context(), gen.UpdateStrategyJoinerParams{
		Quantity:        in.Quantity,
		ReEntry:         in.ReEntry,
		ReEntryTriggered: in.ReEntryTriggered,
		Multiplier:      in.Multiplier,
		IsActive:        in.IsActive,
		ID:              id,
	}); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (a *App) handleDeleteStrategyJoiner(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("joinerId"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	if err := a.Q.DeleteStrategyJoiner(r.Context(), id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// ─── Strategy Positions & Orders ──────────────────────────────────────────────

func (a *App) handleListStrategyPositions(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	list, err := a.Q.ListStrategyPositions(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if list == nil {
		list = []gen.StrategyPosition{}
	}
	writeJSON(w, http.StatusOK, list)
}

func (a *App) handleListStrategyOrders(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	list, err := a.Q.ListOrdersByStrategy(r.Context(), sql.NullInt64{Int64: id, Valid: true})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if list == nil {
		list = []gen.Order{}
	}
	writeJSON(w, http.StatusOK, list)
}

type strategyOrderResult struct {
	BrokerID   int64  `json:"broker_id"`
	BrokerName string `json:"broker_name"`
	Success    bool   `json:"success"`
	OrderID    string `json:"order_id,omitempty"`
	Error      string `json:"error,omitempty"`
}

func (a *App) handleStrategyPlaceOrder(w http.ResponseWriter, r *http.Request) {
	strategyID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	var req struct {
		Data map[string]any `json:"data"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	joiners, err := a.Q.ListStrategyJoiners(r.Context(), strategyID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if len(joiners) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "no active joiners for this strategy"})
		return
	}

	baseQty, _ := req.Data["quantity"].(string)
	baseQtyFloat := 0.0
	if baseQty != "" {
		baseQtyFloat, _ = strconv.ParseFloat(baseQty, 64)
	}

	results := make([]strategyOrderResult, 0, len(joiners))
	for _, j := range joiners {
		if j.IsActive == 0 {
			continue
		}
		qty := baseQtyFloat
		if qty == 0 {
			qty = 1
		}
		if j.Quantity > 0 {
			qty = qty * j.Quantity
		}
		if j.Multiplier > 0 {
			qty = qty * j.Multiplier
		}

		bc, err := a.brokerClient(j.BrokerID)
		if err != nil {
			results = append(results, strategyOrderResult{
				BrokerID: j.BrokerID, BrokerName: "", Success: false, Error: err.Error(),
			})
			continue
		}

		orderData := req.Data
		orderData["quantity"] = strconv.Itoa(int(qty))

		var result map[string]any
		switch bc.brokerName {
		case "angel":
			var placeErr error
			result, placeErr = bc.angelClient.PlaceOrder(bc.authToken, orderData)
			if placeErr != nil {
				results = append(results, strategyOrderResult{
					BrokerID: j.BrokerID, BrokerName: bc.brokerName, Success: false, Error: placeErr.Error(),
				})
				continue
			}
			if err := checkAngelError(result); err != nil {
				results = append(results, strategyOrderResult{
					BrokerID: j.BrokerID, BrokerName: bc.brokerName, Success: false, Error: err.Error(),
				})
				continue
			}
			orderID, _ := result["order_id"].(string)
			results = append(results, strategyOrderResult{
				BrokerID: j.BrokerID, BrokerName: bc.brokerName, Success: true, OrderID: orderID,
			})
		case "aliceblue":
			var placeErr error
			result, placeErr = bc.aliceClient.PlaceOrder(bc.authToken, orderData)
			if placeErr != nil {
				results = append(results, strategyOrderResult{
					BrokerID: j.BrokerID, BrokerName: bc.brokerName, Success: false, Error: placeErr.Error(),
				})
				continue
			}
			if err := checkAliceError(result); err != nil {
				results = append(results, strategyOrderResult{
					BrokerID: j.BrokerID, BrokerName: bc.brokerName, Success: false, Error: err.Error(),
				})
				continue
			}
			var orderID string
			if r, ok := result["result"].([]any); ok && len(r) > 0 {
				if first, ok := r[0].(map[string]any); ok {
					orderID, _ = first["brokerOrderId"].(string)
				}
			}
			if orderID == "" {
				orderID, _ = result["brokerOrderId"].(string)
			}
			results = append(results, strategyOrderResult{
				BrokerID: j.BrokerID, BrokerName: bc.brokerName, Success: true, OrderID: orderID,
			})
		default:
			results = append(results, strategyOrderResult{
				BrokerID: j.BrokerID, BrokerName: bc.brokerName, Success: false, Error: "unsupported broker",
			})
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"results": results})
}
