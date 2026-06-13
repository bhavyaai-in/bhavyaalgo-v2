package blueprints

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	tradingdb "bhavyaaialgo/backend/db/trading/gen"
)

func generateSecretKey() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (a *App) RegisterStrategyRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/strategy-types", a.authMiddleware(a.handleListStrategyTypes))
	mux.HandleFunc("POST /api/strategy-types", a.authMiddleware(a.handleCreateStrategyType))
	mux.HandleFunc("PUT /api/strategy-types/{id}", a.authMiddleware(a.handleUpdateStrategyType))
	mux.HandleFunc("DELETE /api/strategy-types/{id}", a.authMiddleware(a.handleDeleteStrategyType))

	mux.HandleFunc("GET /api/strategies", a.authMiddleware(a.handleListStrategies))
	mux.HandleFunc("POST /api/strategies", a.authMiddleware(a.handleCreateStrategy))
	mux.HandleFunc("GET /api/strategies/{id}", a.authMiddleware(a.handleGetStrategy))
	mux.HandleFunc("PUT /api/strategies/{id}", a.authMiddleware(a.handleUpdateStrategy))
	mux.HandleFunc("DELETE /api/strategies/{id}", a.authMiddleware(a.handleDeleteStrategy))

	mux.HandleFunc("GET /api/strategies/{id}/info", a.authMiddleware(a.handleListStrategyInfo))
	mux.HandleFunc("POST /api/strategies/{id}/info", a.authMiddleware(a.handleCreateStrategyInfo))
	mux.HandleFunc("DELETE /api/strategies/{id}/info/{infoId}", a.authMiddleware(a.handleDeleteStrategyInfo))

	mux.HandleFunc("GET /api/strategies/{id}/joiners", a.authMiddleware(a.handleListStrategyJoiners))
	mux.HandleFunc("POST /api/strategies/{id}/joiners", a.authMiddleware(a.handleCreateStrategyJoiner))
	mux.HandleFunc("PUT /api/strategies/{id}/joiners/{joinerId}", a.authMiddleware(a.handleUpdateStrategyJoiner))
	mux.HandleFunc("DELETE /api/strategies/{id}/joiners/{joinerId}", a.authMiddleware(a.handleDeleteStrategyJoiner))

	mux.HandleFunc("GET /api/strategies/{id}/paper-positions", a.authMiddleware(a.handleListStrategyPositions))
	mux.HandleFunc("GET /api/strategies/{id}/positions", a.authMiddleware(a.handleListPositionsByStrategy))
	mux.HandleFunc("GET /api/strategies/{id}/orders", a.authMiddleware(a.handleListStrategyOrders))

	mux.HandleFunc("POST /api/strategies/{id}/place-order", a.authMiddleware(a.handleStrategyPlaceOrder))
}

func (a *App) handleListStrategyTypes(w http.ResponseWriter, r *http.Request) {
	types, err := a.TradingQ.ListStrategyTypes(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if types == nil {
		types = []tradingdb.StrategyType{}
	}
	writeJSON(w, http.StatusOK, types)
}

type strategyTypeInput struct {
	Name             string `json:"name"`
	RulesExplanation string `json:"rules_explanation"`
}

func (a *App) handleCreateStrategyType(w http.ResponseWriter, r *http.Request) {
	var in strategyTypeInput
	if err := decodeJSON(r, &in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	id, err := a.TradingQ.CreateStrategyType(r.Context(), tradingdb.CreateStrategyTypeParams{
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
	if err := a.TradingQ.UpdateStrategyType(r.Context(), tradingdb.UpdateStrategyTypeParams{
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
	if err := a.TradingQ.DeleteStrategyType(r.Context(), id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

type strategyInput struct {
	Name              string  `json:"name"`
	StrategySecretKey string  `json:"strategy_secret_key"`
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
	list, err := a.TradingQ.ListStrategies(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if list == nil {
		list = []tradingdb.Strategy{}
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
	id, err := a.TradingQ.CreateStrategy(r.Context(), tradingdb.CreateStrategyParams{
		Name:              in.Name,
		StrategySecretKey: in.StrategySecretKey,
		StrategyType:      in.StrategyType,
		PositionStatus:    in.PositionStatus,
		InstrumentToken:   in.InstrumentToken,
		Exchange:          in.Exchange,
		Side:              in.Side,
		AtmOtm:            in.AtmOtm,
		ImageUrl:          in.ImageURL,
		Color:             in.Color,
		IsActive:          in.IsActive,
		IsLocked:          in.IsLocked,
		Message:           in.Message,
		ExpiryDate:        in.ExpiryDate,
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
	s, err := a.TradingQ.GetStrategy(r.Context(), id)
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
		existing, err := a.TradingQ.GetStrategy(r.Context(), id)
		if err == nil {
			in.StrategySecretKey = existing.StrategySecretKey
		}
	}
	if err := a.TradingQ.UpdateStrategy(r.Context(), tradingdb.UpdateStrategyParams{
		Name:              in.Name,
		StrategySecretKey: in.StrategySecretKey,
		StrategyType:      in.StrategyType,
		PositionStatus:    in.PositionStatus,
		InstrumentToken:   in.InstrumentToken,
		Exchange:          in.Exchange,
		Side:              in.Side,
		AtmOtm:            in.AtmOtm,
		ImageUrl:          in.ImageURL,
		Color:             in.Color,
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
	if err := a.TradingQ.DeleteStrategy(r.Context(), id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (a *App) handleListStrategyInfo(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	list, err := a.TradingQ.ListStrategyInfo(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if list == nil {
		list = []tradingdb.StrategyInfo{}
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
	id, err := a.TradingQ.CreateStrategyInfo(r.Context(), tradingdb.CreateStrategyInfoParams{
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
	if err := a.TradingQ.DeleteStrategyInfo(r.Context(), id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

type strategyInfoInput struct {
	Info string `json:"info"`
}

func (a *App) handleListStrategyJoiners(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	list, err := a.TradingQ.ListStrategyJoiners(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if list == nil {
		list = []tradingdb.StrategyJoiner{}
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
	id, err := a.TradingQ.CreateStrategyJoiner(r.Context(), tradingdb.CreateStrategyJoinerParams{
		BrokerID:         in.BrokerID,
		StrategyID:       strategyID,
		Quantity:         in.Quantity,
		ReEntry:          in.ReEntry,
		ReEntryTriggered: in.ReEntryTriggered,
		Multiplier:       in.Multiplier,
		IsActive:         in.IsActive,
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
	if err := a.TradingQ.UpdateStrategyJoiner(r.Context(), tradingdb.UpdateStrategyJoinerParams{
		Quantity:         in.Quantity,
		ReEntry:          in.ReEntry,
		ReEntryTriggered: in.ReEntryTriggered,
		Multiplier:       in.Multiplier,
		IsActive:         in.IsActive,
		ID:               id,
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
	if err := a.TradingQ.DeleteStrategyJoiner(r.Context(), id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

type strategyJoinerInput struct {
	BrokerID         int64   `json:"broker_id"`
	Quantity         float64 `json:"quantity"`
	ReEntry          int64   `json:"re_entry"`
	ReEntryTriggered int64   `json:"re_entry_triggered"`
	Multiplier       float64 `json:"multiplier"`
	IsActive         int64   `json:"is_active"`
}

func (a *App) handleListStrategyPositions(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	list, err := a.TradingQ.ListStrategyPositions(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if list == nil {
		list = []tradingdb.StrategyPosition{}
	}
	writeJSON(w, http.StatusOK, list)
}

func (a *App) handleListPositionsByStrategy(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	list, err := a.TradingQ.ListPositionsByStrategy(r.Context(), sql.NullInt64{Int64: id, Valid: true})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if list == nil {
		list = []tradingdb.Position{}
	}
	writeJSON(w, http.StatusOK, list)
}

func (a *App) handleListStrategyOrders(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	list, err := a.TradingQ.ListOrdersByStrategy(r.Context(), sql.NullInt64{Int64: id, Valid: true})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if list == nil {
		list = []tradingdb.Order{}
	}
	writeJSON(w, http.StatusOK, list)
}

type strategyOrderResult struct {
	BrokerID     int64  `json:"broker_id"`
	BrokerName   string `json:"broker_name"`
	LocalOrderID int64  `json:"local_order_id,omitempty"`
	Success      bool   `json:"success"`
	OrderID      string `json:"order_id,omitempty"`
	Message      string `json:"message,omitempty"`
}

type strategyPlaceOrderRequest struct {
	Data strategyOrderPayload `json:"data"`
}

type strategyOrderPayload struct {
	Tag             string `json:"tag"`
	Variety         string `json:"variety"`
	Tradingsymbol   string `json:"tradingsymbol"`
	Exchange        string `json:"exchange"`
	SymbolToken     any    `json:"symboltoken"`
	TransactionType string `json:"transactiontype"`
	EntryExit       string `json:"entry_exit"`
	SqOff           bool   `json:"sqoff"`
	Product         string `json:"product"`
	OrderType       string `json:"ordertype"`
	Quantity        any    `json:"quantity"`
	Price           any    `json:"price"`
	TriggerPrice    any    `json:"triggerprice"`
	Action          string `json:"action"`
}

func (p strategyOrderPayload) isSquareOff() bool {
	return p.SqOff || p.Action == "sqoff" || p.Action == "squareoff" || p.Action == "close"
}

func (p strategyOrderPayload) isNew() bool {
	return !p.isSquareOff()
}

func (p strategyOrderPayload) normalize() strategyOrderPayload {
	if p.Variety == "" {
		p.Variety = "regular"
	}
	if p.Product == "" {
		p.Product = "MIS"
	}
	if p.OrderType == "" {
		p.OrderType = "MARKET"
	}
	if p.TransactionType == "" {
		p.TransactionType = "BUY"
	}
	if p.EntryExit == "" {
		p.EntryExit = "ENTRY"
	}
	p.EntryExit = strings.ToUpper(strings.TrimSpace(p.EntryExit))
	p.Action = strings.ToLower(strings.TrimSpace(p.Action))
	return p
}


func (p strategyOrderPayload) toMap() map[string]any {
	return map[string]any{
		"tag":             p.Tag,
		"variety":         p.Variety,
		"tradingsymbol":   p.Tradingsymbol,
		"exchange":        p.Exchange,
		"symboltoken":     p.SymbolToken,
		"transactiontype": p.TransactionType,
		"product":         p.Product,
		"ordertype":       p.OrderType,
		"quantity":        p.Quantity,
		"price":           p.Price,
		"triggerprice":    p.TriggerPrice,
	}
}

func isBrokerSupported(name string) bool {
	name = strings.ToLower(strings.TrimSpace(name))
	return name == "angel" || name == "aliceblue"
}

type validJoiner struct {
	Joiner tradingdb.StrategyJoiner
	Broker tradingdb.Broker
}

func (a *App) getValidStrategyJoiners(ctx context.Context, strategyID int64) ([]validJoiner, error) {
	joiners, err := a.TradingQ.ListStrategyJoiners(ctx, strategyID)
	if err != nil {
		return nil, err
	}

	today := time.Now().Format("2006-01-02")
	var result []validJoiner

	for _, j := range joiners {
		if j.IsActive == 0 {
			continue
		}

		broker, err := a.TradingQ.GetBroker(ctx, j.BrokerID)
		if err != nil {
			log.Printf("get broker failed for brokerID=%d: %v", j.BrokerID, err)
			continue
		}

		if broker.IsActive == 0 || broker.IsDisabled == 1 {
			continue
		}

		if !isBrokerSupported(broker.BrokerName) {
			continue
		}

		if broker.TokenStatus != "connected" {
			continue
		}

		if !strings.HasPrefix(broker.BrokerTokenDate, today) {
			continue
		}

		result = append(result, validJoiner{
			Joiner: j,
			Broker: broker,
		})
	}

	return result, nil
}


func (a *App) handleStrategyPlaceOrder(w http.ResponseWriter, r *http.Request) {
	strategyID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	var req strategyPlaceOrderRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	payload := req.Data.normalize()
	if payload.Tradingsymbol == "" || payload.Exchange == "" || payload.SymbolToken == nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid order payload"})
		return
	}

	strategy, err := a.TradingQ.GetStrategy(r.Context(), strategyID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "strategy not found"})
		return
	}

	// 1. If it's a square-off order, square off all open positions and exit
	if payload.isSquareOff() {
		if err := a.squareOffAllOpenPositions(r.Context(), strategyID, payload.Tradingsymbol, payload.Exchange); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"status":        "squared_off",
			"tradingsymbol": payload.Tradingsymbol,
			"exchange":      payload.Exchange,
		})
		return
	}

	// 2. If it's a new order, validate quantity first
	if payload.isNew() && orderValueFloat(payload.toMap(), "quantity", 0) <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid order quantity"})
		return
	}

	// 3. Square off existing open positions of same symbol, exchange, and product (both strategy and joiners/brokers)
	if err := a.squareOffSameSymbolExchangeProductPositions(r.Context(), strategyID, payload.Tradingsymbol, payload.Exchange, payload.Product); err != nil {
		log.Printf("squareOffSameSymbolExchangeProductPositions error: strategyID=%d symbol=%s product=%s err=%v", strategyID, payload.Tradingsymbol, payload.Product, err)
	}

	// 4. Create strategy position directly as "open"
	var buyQty, sellQty, buyPrice, sellPrice float64
	side := strings.ToUpper(payload.TransactionType)
	priceVal := orderValueFloat(payload.toMap(), "price", 0)
	qtyVal := orderValueFloat(payload.toMap(), "quantity", 0)

	// Fetch live LTP if a connected broker token is available
	tokenStr := fmt.Sprintf("%d", int64(orderValueFloat(payload.toMap(), "symboltoken", 0)))
	if ltp, err := a.fetchLiveLTP(r.Context(), payload.Exchange, payload.Tradingsymbol, tokenStr); err == nil && ltp > 0 {
		priceVal = ltp
	} else if err != nil {
		log.Printf("fetchLiveLTP failed: %v", err)
	}

	if side == "BUY" {
		buyQty = qtyVal
		buyPrice = priceVal
	} else {
		sellQty = qtyVal
		sellPrice = priceVal
	}

	strategyPositionParams := tradingdb.CreateStrategyPositionParams{
		StrategyID:      strategyID,
		Tradingsymbol:   payload.Tradingsymbol,
		StrategyType:    strategy.StrategyType,
		Exchange:        payload.Exchange,
		InstrumentToken: int64(orderValueFloat(payload.toMap(), "symboltoken", 0)),
		Quantity:        qtyVal,
		LastPrice:       priceVal,
		BuyQuantity:     buyQty,
		Multiplier:      1,
		SellQuantity:    sellQty,
		Side:            side,
		BuyPrice:        buyPrice,
		SellPrice:       sellPrice,
		Product:         payload.Product,
		Status:          "open",
		Message:         "strategy order placed",
	}

	_, err = a.TradingQ.CreateStrategyPosition(r.Context(), strategyPositionParams)
	if err != nil {
		log.Printf("create strategy position failed: strategy_id=%d error=%v", strategyID, err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	// 5. Get joiners whose broker token is generated today
	validJoiners, err := a.getValidStrategyJoiners(r.Context(), strategyID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if len(validJoiners) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "no active joiners with valid today's token for this strategy"})
		return
	}

	baseQty := orderValueFloat(payload.toMap(), "quantity", 0)
	orderData := payload.toMap()
	log.Printf("placing strategy order: strategy_id=%d base_qty=%f order_data=%v", strategyID, baseQty, orderData)

	var wg sync.WaitGroup
	var mu sync.Mutex

	allSuccess := true
	processedJoiners := len(validJoiners)
	successCount := 0
	results := make([]strategyOrderResult, 0, processedJoiners)

	for _, vj := range validJoiners {
		wg.Add(1)
		go func(vj validJoiner) {
			defer wg.Done()
			ctx := r.Context()
			j := vj.Joiner
			b := vj.Broker

			qty := strategyJoinerQuantity(baseQty, j)
			od := cloneOrderData(orderData)
			log.Printf("placing strategy order concurrently: strategy_id=%d broker_id=%d quantity=%f order_data=%v", strategyID, j.BrokerID, qty, od)
			od["quantity"] = strconv.Itoa(int(qty))

			// Create database order record
			orderID, cerr := a.TradingQ.CreateOrder(ctx, strategyOrderCreateParams(strategyID, j, od, qty))
			if cerr != nil || orderID == 0 {
				mu.Lock()
				allSuccess = false
				errMsg := "failed to create order in database"
				if cerr != nil {
					errMsg = cerr.Error()
					log.Printf("create strategy order failed: %v", cerr)
				}
				results = append(results, strategyOrderResult{
					BrokerID:   j.BrokerID,
					BrokerName: b.BrokerName,
					Success:    false,
					Message:    errMsg,
				})
				mu.Unlock()
				return
			}

			markOrderError := func(message string) {
				a.markStrategyOrderError(ctx, orderID, message)
				mu.Lock()
				allSuccess = false
				results = append(results, strategyOrderResult{
					BrokerID:     j.BrokerID,
					BrokerName:   b.BrokerName,
					LocalOrderID: orderID,
					Success:      false,
					Message:      message,
				})
				mu.Unlock()
			}

			bc, clientErr := a.brokerClient(j.BrokerID)
			if clientErr != nil {
				markOrderError("broker client error: " + clientErr.Error())
				return
			}

			brokerOrderID, statusMessage, placeErr := placeStrategyOrder(bc, od)
			if placeErr != nil {
				errMsg := placeErr.Error()
				if statusMessage != "" {
					errMsg = fmt.Sprintf("%s (%s)", statusMessage, placeErr.Error())
				}
				markOrderError(errMsg)
				return
			}

			if statusMessage == "" {
				statusMessage = "Order placed successfully"
			}
			if _, err := a.TradingDB.ExecContext(
				ctx,
				"UPDATE orders SET status='placed', order_id=?, status_message=?, updated_at=datetime('now','localtime') WHERE id=?",
				brokerOrderID,
				statusMessage,
				orderID,
			); err != nil {
				markOrderError("failed to update placed order status: " + err.Error())
				return
			}

			fmt.Println("pos creation run ")
			if err := a.createPositionFromOrder(ctx, strategyID, j, od, qty, orderID); err != nil {
				log.Printf("create position: strategy_id=%d order_id=%s error=%v", strategyID, brokerOrderID, err)
				fmt.Println("position creation err", err)
			}

			mu.Lock()
			results = append(results, strategyOrderResult{
				BrokerID:     j.BrokerID,
				BrokerName:   b.BrokerName,
				LocalOrderID: orderID,
				Success:      true,
				OrderID:      brokerOrderID,
			})
			successCount++
			mu.Unlock()
		}(vj)
	}

	wg.Wait()

	writeJSON(w, http.StatusOK, map[string]any{
		"status": func() string {
			if allSuccess {
				return "success"
			}
			if successCount == 0 {
				return "error"
			}
			return "partial"
		}(),
		"success_count": successCount,
		"failed_count":  processedJoiners - successCount,
		"results":       results,
	})
}

func (a *App) squareOffPosition(ctx context.Context, pos tradingdb.Position) error {
	bc, err := a.brokerClient(pos.BrokerID)
	if err != nil {
		return fmt.Errorf("failed to get broker client: %w", err)
	}

	oppositeSide := "BUY"
	if strings.EqualFold(pos.Side, "BUY") {
		oppositeSide = "SELL"
	}

	variety := "NORMAL"
	tag := "sqoff"
	if pos.EntryOrderID.Valid {
		entryOrder, err := a.TradingQ.GetOrder(ctx, pos.EntryOrderID.Int64)
		if err == nil {
			if entryOrder.Variety != "" {
				variety = entryOrder.Variety
			}
			if entryOrder.Tag != "" {
				tag = entryOrder.Tag + "-sqoff"
			}
		}
	}

	payload := map[string]any{
		"variety":         variety,
		"tradingsymbol":   pos.Tradingsymbol,
		"exchange":        pos.Exchange,
		"symboltoken":     pos.InstrumentToken,
		"transactiontype": oppositeSide,
		"product":         pos.Product,
		"ordertype":       "MARKET",
		"quantity":        strconv.Itoa(int(pos.Quantity)),
	}

	createParams := tradingdb.CreateOrderParams{
		BrokerID:              pos.BrokerID,
		StrategyID:            pos.StrategyID,
		PositionID:            sql.NullInt64{Int64: pos.ID, Valid: true},
		OrderID:               "",
		StatusMessage:         "placing square off order",
		Tag:                   tag,
		Variety:               variety,
		Tradingsymbol:         pos.Tradingsymbol,
		Exchange:              pos.Exchange,
		InstrumentToken:       pos.InstrumentToken,
		BrokerInstrumentToken: pos.BrokerInstrumentToken,
		TransactionType:       oppositeSide,
		Product:               pos.Product,
		OrderType:             "MARKET",
		Validity:              "DAY",
		Status:                "placing",
		Quantity:              pos.Quantity,
		Price:                 0,
		TriggerPrice:          0,
		AveragePrice:          0,
		FilledQuantity:        0,
		PendingQuantity:       0,
		CancelledQuantity:     0,
	}

	orderID, err := a.TradingQ.CreateOrder(ctx, createParams)
	if err != nil {
		return fmt.Errorf("failed to create square off order in DB: %w", err)
	}

	brokerOrderID, statusMsg, placeErr := placeStrategyOrder(bc, payload)
	if placeErr != nil {
		errMsg := placeErr.Error()
		if statusMsg != "" {
			errMsg = fmt.Sprintf("%s (%s)", statusMsg, placeErr.Error())
		}
		a.markStrategyOrderError(ctx, orderID, errMsg)
		return fmt.Errorf("broker square off failed: %s", errMsg)
	}

	if statusMsg == "" {
		statusMsg = "Square off order placed successfully"
	}

	if _, err := a.TradingDB.ExecContext(
		ctx,
		"UPDATE orders SET status='placed', order_id=?, status_message=?, updated_at=datetime('now','localtime') WHERE id=?",
		brokerOrderID,
		statusMsg,
		orderID,
	); err != nil {
		log.Printf("failed to update square off order: %v", err)
	}

	err = a.TradingQ.UpdatePosition(ctx, tradingdb.UpdatePositionParams{
		EntryOrderID: pos.EntryOrderID,
		ExitOrderID:  sql.NullInt64{Int64: orderID, Valid: true},
		Quantity:     pos.Quantity,
		LastPrice:    pos.LastPrice,
		BuyQuantity:  pos.BuyQuantity,
		SellQuantity: pos.SellQuantity,
		Multiplier:   pos.Multiplier,
		Side:         pos.Side,
		BuyPrice:     pos.BuyPrice,
		SellPrice:    pos.SellPrice,
		Product:      pos.Product,
		Status:       "closed",
		Message:      "squared off successfully",
		ID:           pos.ID,
	})
	if err != nil {
		return fmt.Errorf("failed to update position status to closed: %w", err)
	}

	return nil
}

func (a *App) squareOffStrategyPosition(ctx context.Context, sp tradingdb.StrategyPosition) error {
	exitPrice := sp.LastPrice

	tokenStr := fmt.Sprintf("%d", sp.InstrumentToken)
	if ltp, err := a.fetchLiveLTP(ctx, sp.Exchange, sp.Tradingsymbol, tokenStr); err == nil && ltp > 0 {
		exitPrice = ltp
	} else if err != nil {
		log.Printf("fetchLiveLTP failed for squareoff: %v", err)
	}

	buyQty := sp.BuyQuantity
	buyPrice := sp.BuyPrice
	sellQty := sp.SellQuantity
	sellPrice := sp.SellPrice

	if strings.ToUpper(sp.Side) == "BUY" {
		sellQty = sp.Quantity
		sellPrice = exitPrice
	} else {
		buyQty = sp.Quantity
		buyPrice = exitPrice
	}

	return a.TradingQ.UpdateStrategyPosition(ctx, tradingdb.UpdateStrategyPositionParams{
		Quantity:     sp.Quantity,
		LastPrice:    exitPrice,
		BuyQuantity:  buyQty,
		Multiplier:   sp.Multiplier,
		SellQuantity: sellQty,
		Side:         sp.Side,
		BuyPrice:     buyPrice,
		SellPrice:    sellPrice,
		Product:      sp.Product,
		Status:       "closed",
		Message:      "squared off successfully",
		ID:           sp.ID,
	})
}

func (a *App) squareOffAllOpenPositions(ctx context.Context, strategyID int64, symbol, exchange string) error {
	todayPattern := time.Now().Format("2006-01-02") + "%"

	positions, err := a.TradingQ.ListActivePositionsByStrategy(ctx, tradingdb.ListActivePositionsByStrategyParams{
		StrategyID: sql.NullInt64{Int64: strategyID, Valid: true},
		CreatedAt:  todayPattern,
	})
	if err != nil {
		return fmt.Errorf("list active positions: %w", err)
	}

	strategyPositions, err := a.TradingQ.ListActiveStrategyPositions(ctx, tradingdb.ListActiveStrategyPositionsParams{
		StrategyID: strategyID,
		CreatedAt:  todayPattern,
	})
	if err != nil {
		return fmt.Errorf("list active strategy positions: %w", err)
	}

	for _, pos := range positions {
		if symbol != "" && !strings.EqualFold(pos.Tradingsymbol, symbol) {
			continue
		}
		if exchange != "" && !strings.EqualFold(pos.Exchange, exchange) {
			continue
		}

		if err := a.squareOffPosition(ctx, pos); err != nil {
			log.Printf("squareOffPosition error for pos ID %d: %v", pos.ID, err)
		}
	}

	for _, sp := range strategyPositions {
		if symbol != "" && !strings.EqualFold(sp.Tradingsymbol, symbol) {
			continue
		}
		if exchange != "" && !strings.EqualFold(sp.Exchange, exchange) {
			continue
		}

		if err := a.squareOffStrategyPosition(ctx, sp); err != nil {
			log.Printf("squareOffStrategyPosition error for sp ID %d: %v", sp.ID, err)
		}
	}

	return nil
}

func (a *App) squareOffSameSymbolSameSidePositions(ctx context.Context, strategyID int64, symbol, exchange, side string) error {
	todayPattern := time.Now().Format("2006-01-02") + "%"

	positions, err := a.TradingQ.ListActivePositionsByStrategy(ctx, tradingdb.ListActivePositionsByStrategyParams{
		StrategyID: sql.NullInt64{Int64: strategyID, Valid: true},
		CreatedAt:  todayPattern,
	})
	if err != nil {
		return fmt.Errorf("list active positions: %w", err)
	}

	strategyPositions, err := a.TradingQ.ListActiveStrategyPositions(ctx, tradingdb.ListActiveStrategyPositionsParams{
		StrategyID: strategyID,
		CreatedAt:  todayPattern,
	})
	if err != nil {
		return fmt.Errorf("list active strategy positions: %w", err)
	}

	for _, pos := range positions {
		if !strings.EqualFold(pos.Tradingsymbol, symbol) || !strings.EqualFold(pos.Exchange, exchange) || !strings.EqualFold(pos.Side, side) {
			continue
		}

		if err := a.squareOffPosition(ctx, pos); err != nil {
			log.Printf("squareOffPosition error for pos ID %d: %v", pos.ID, err)
		}
	}

	for _, sp := range strategyPositions {
		if !strings.EqualFold(sp.Tradingsymbol, symbol) || !strings.EqualFold(sp.Exchange, exchange) || !strings.EqualFold(sp.Side, side) {
			continue
		}

		if err := a.squareOffStrategyPosition(ctx, sp); err != nil {
			log.Printf("squareOffStrategyPosition error for sp ID %d: %v", sp.ID, err)
		}
	}

	return nil
}

func (a *App) squareOffSameSymbolExchangeProductPositions(ctx context.Context, strategyID int64, symbol, exchange, product string) error {
	todayPattern := time.Now().Format("2006-01-02") + "%"

	positions, err := a.TradingQ.ListActivePositionsByStrategy(ctx, tradingdb.ListActivePositionsByStrategyParams{
		StrategyID: sql.NullInt64{Int64: strategyID, Valid: true},
		CreatedAt:  todayPattern,
	})
	if err != nil {
		return fmt.Errorf("list active positions: %w", err)
	}

	strategyPositions, err := a.TradingQ.ListActiveStrategyPositions(ctx, tradingdb.ListActiveStrategyPositionsParams{
		StrategyID: strategyID,
		CreatedAt:  todayPattern,
	})
	if err != nil {
		return fmt.Errorf("list active strategy positions: %w", err)
	}

	for _, pos := range positions {
		if !strings.EqualFold(pos.Tradingsymbol, symbol) || !strings.EqualFold(pos.Exchange, exchange) || !strings.EqualFold(pos.Product, product) {
			continue
		}

		if err := a.squareOffPosition(ctx, pos); err != nil {
			log.Printf("squareOffPosition error for pos ID %d: %v", pos.ID, err)
		}
	}

	for _, sp := range strategyPositions {
		if !strings.EqualFold(sp.Tradingsymbol, symbol) || !strings.EqualFold(sp.Exchange, exchange) || !strings.EqualFold(sp.Product, product) {
			continue
		}

		if err := a.squareOffStrategyPosition(ctx, sp); err != nil {
			log.Printf("squareOffStrategyPosition error for sp ID %d: %v", sp.ID, err)
		}
	}

	return nil
}

func (a *App) markStrategyOrderError(ctx context.Context, orderID int64, message string) {
	if message == "" {
		message = "order failed"
	}
	if err := a.TradingQ.UpdateOrderStatus(ctx, tradingdb.UpdateOrderStatusParams{
		ID:            orderID,
		Status:        "error",
		StatusMessage: message,
	}); err != nil {
		log.Printf("update strategy order error status: order_id=%d error=%v", orderID, err)
	}
}

func (a *App) createPositionFromOrder(ctx context.Context, strategyID int64, joiner tradingdb.StrategyJoiner, orderData map[string]any, qty float64, orderID int64) error {
	strategy, err := a.TradingQ.GetStrategy(ctx, strategyID)
	if err != nil {
		log.Printf("get strategy: strategy_id=%d error=%v", strategyID, err)
		return fmt.Errorf("get strategy: %w", err)
	}

	symbol := orderValueString(orderData, "tradingsymbol", "")
	if symbol == "" {
		log.Printf("missing symbol in order data: %v", orderData)
		return nil
	}

	exchange := orderValueString(orderData, "exchange", "")
	token := orderValueFloat(orderData, "symboltoken", 0)
	if token == 0 {
		log.Printf("missing instrument token in order data: %v", orderData)
		return nil
	}

	action := orderValueString(orderData, "transactiontype", "BUY")
	product := orderValueString(orderData, "product", "INTRADAY")
	price := orderValueFloat(orderData, "price", 0)

	params := tradingdb.CreatePositionParams{
		BrokerID:              joiner.BrokerID,
		StrategyID:            sql.NullInt64{Int64: strategyID, Valid: true},
		EntryOrderID:          sql.NullInt64{Int64: orderID, Valid: true},
		Tradingsymbol:         symbol,
		StrategyType:          strategy.StrategyType,
		Exchange:              exchange,
		InstrumentToken:       int64(token),
		BrokerInstrumentToken: int64(token),
		Quantity:              qty,
		LastPrice:             price,
		Multiplier:            joiner.Multiplier,
		Product:               product,
		Status:                "open",
		Message:               "",
	}

	if action == "BUY" {
		params.BuyQuantity = qty
		params.BuyPrice = price
		params.Side = "BUY"
	} else {
		params.SellQuantity = qty
		params.SellPrice = price
		params.Side = "SELL"
	}

	if _, err := a.TradingQ.CreatePosition(ctx, params); err != nil {
		log.Printf("create position: strategy_id=%d error=%v", strategyID, err)
		return fmt.Errorf("create position: %w", err)
	}

	return nil
}

func strategyJoinerQuantity(baseQty float64, joiner tradingdb.StrategyJoiner) float64 {
	qty := baseQty
	if qty == 0 {
		qty = 1
	}
	if joiner.Quantity > 0 {
		qty *= joiner.Quantity
	}
	if joiner.Multiplier > 0 {
		qty *= joiner.Multiplier
	}
	return qty
}

func cloneOrderData(data map[string]any) map[string]any {
	out := make(map[string]any, len(data))
	for k, v := range data {
		out[k] = v
	}
	return out
}

func strategyOrderCreateParams(strategyID int64, joiner tradingdb.StrategyJoiner, orderData map[string]any, qty float64) tradingdb.CreateOrderParams {
	return tradingdb.CreateOrderParams{
		BrokerID:              joiner.BrokerID,
		StrategyID:            sql.NullInt64{Int64: strategyID, Valid: true},
		OrderID:               "",
		StatusMessage:         "",
		Status:                "placing",
		Quantity:              qty,
		Tag:                   orderValueString(orderData, "tag", ""),
		Variety:               orderValueString(orderData, "variety", "NORMAL"),
		Tradingsymbol:         orderValueString(orderData, "tradingsymbol", ""),
		Exchange:              orderValueString(orderData, "exchange", ""),
		TransactionType:       orderValueString(orderData, "transactiontype", ""),
		Product:               orderValueString(orderData, "product", "INTRADAY"),
		OrderType:             orderValueString(orderData, "ordertype", "MARKET"),
		Price:                 orderValueFloat(orderData, "price", 0),
		TriggerPrice:          orderValueFloat(orderData, "triggerprice", 0),
		Validity:              "DAY",
		AveragePrice:          0,
		FilledQuantity:        0,
		PendingQuantity:       0,
		CancelledQuantity:     0,
		InstrumentToken:       int64(orderValueFloat(orderData, "symboltoken", 0)),
		BrokerInstrumentToken: 0,
		PositionID:            sql.NullInt64{},
	}
}

func orderValueString(data map[string]any, key, fallback string) string {
	if v, ok := data[key]; ok && v != nil {
		if s := valueAsString(v); s != "" {
			return s
		}
	}
	return fallback
}

func orderValueFloat(data map[string]any, key string, fallback float64) float64 {
	if v, ok := data[key]; ok && v != nil {
		switch x := v.(type) {
		case float64:
			return x
		case float32:
			return float64(x)
		case int:
			return float64(x)
		case int64:
			return float64(x)
		case string:
			if x == "" {
				return fallback
			}
			if parsed, err := strconv.ParseFloat(x, 64); err == nil {
				return parsed
			}
		default:
			if parsed, err := strconv.ParseFloat(fmt.Sprintf("%v", x), 64); err == nil {
				return parsed
			}
		}
	}
	return fallback
}

func mapPayloadForBroker(brokerName string, payload map[string]any) map[string]any {
	brokerName = strings.ToLower(strings.TrimSpace(brokerName))
	out := cloneOrderData(payload)

	product := ""
	if p, ok := out["product"].(string); ok {
		product = strings.ToUpper(strings.TrimSpace(p))
	} else if p, ok := out["producttype"].(string); ok {
		product = strings.ToUpper(strings.TrimSpace(p))
	}

	variety := "regular"
	if v, ok := out["variety"].(string); ok {
		variety = strings.ToLower(strings.TrimSpace(v))
	}

	orderType := "MARKET"
	if o, ok := out["ordertype"].(string); ok {
		orderType = strings.ToUpper(strings.TrimSpace(o))
	}

	switch brokerName {
	case "angel":
		// Zerodha "regular" -> Angel "NORMAL"
		angelVariety := "NORMAL"
		switch variety {
		case "regular":
			angelVariety = "NORMAL"
		case "amo":
			angelVariety = "AMO"
		case "co":
			angelVariety = "CO"
		case "bo":
			angelVariety = "BO"
		default:
			angelVariety = strings.ToUpper(variety)
		}
		out["variety"] = angelVariety

		// Zerodha "CNC" -> Angel "DELIVERY", Zerodha "MIS" -> Angel "INTRADAY", Zerodha "NRML" -> Angel "CARRYFORWARD"
		productType := "INTRADAY"
		if product != "" {
			switch product {
			case "CNC":
				productType = "DELIVERY"
			case "MIS":
				productType = "INTRADAY"
			case "NRML":
				productType = "CARRYFORWARD"
			default:
				productType = product
			}
		}
		out["producttype"] = productType
		delete(out, "product")

		// Zerodha "SL-M" -> Angel "STOPLOSS"
		angelOrderType := "MARKET"
		switch orderType {
		case "MARKET":
			angelOrderType = "MARKET"
		case "LIMIT":
			angelOrderType = "LIMIT"
		case "SL", "SL-M", "STOPLOSS":
			angelOrderType = "STOPLOSS"
		default:
			angelOrderType = orderType
		}
		out["ordertype"] = angelOrderType

	case "aliceblue":
		// Alice Blue client normalizes it internally via normalizeOrderPayload.
		// It expects: variety="NORMAL", product="DELIVERY"|"MIS"|"CARRYFORWARD"
		aliceVariety := "NORMAL"
		switch variety {
		case "regular":
			aliceVariety = "NORMAL"
		case "amo":
			aliceVariety = "AMO"
		case "co":
			aliceVariety = "CO"
		case "bo":
			aliceVariety = "BO"
		default:
			aliceVariety = strings.ToUpper(variety)
		}
		out["variety"] = aliceVariety

		aliceProduct := "MIS"
		if product != "" {
			switch product {
			case "CNC":
				aliceProduct = "DELIVERY"
			case "MIS":
				aliceProduct = "MIS"
			case "NRML":
				aliceProduct = "CARRYFORWARD"
			default:
				aliceProduct = product
			}
		}
		out["product"] = aliceProduct

		aliceOrderType := "MARKET"
		if orderType != "" {
			switch orderType {
			case "SL-M":
				aliceOrderType = "SLM"
			default:
				aliceOrderType = orderType
			}
		}
		out["ordertype"] = aliceOrderType
	}

	return out
}

func placeStrategyOrder(bc *brokerClientResult, payload map[string]any) (string, string, error) {
	var result map[string]any
	var err error

	mappedPayload := mapPayloadForBroker(bc.brokerName, payload)

	switch bc.brokerName {
	case "angel":
		result, err = bc.angelClient.PlaceOrder(bc.authToken, mappedPayload)
		if err != nil {
			return "", "", err
		}
		if err := checkAngelError(result); err != nil {
			return "", brokerResultMessage(result), err
		}
		orderID := extractBrokerOrderID(result, "order_id", "orderid")
		if orderID == "" {
			orderID = extractBrokerOrderID(nestedMap(result, "data"), "orderid", "order_id", "uniqueorderid")
		}
		if orderID == "" {
			return "", brokerResultMessage(result), fmt.Errorf("angel order placed response missing order id")
		}
		return orderID, firstNonEmpty(brokerResultMessage(result), "Order placed successfully"), nil

	case "aliceblue":
		result, err = bc.aliceClient.PlaceOrder(bc.authToken, mappedPayload)
		if err != nil {
			return "", "", err
		}
		if err := checkAliceError(result); err != nil {
			return "", brokerResultMessage(result), err
		}
		orderID := extractBrokerOrderID(result, "brokerOrderId", "order_id", "orderid")
		if orderID == "" {
			orderID = extractAliceResultOrderID(result)
		}
		if orderID == "" {
			log.Printf("aliceblue order placement raw result: %v", result)
			return "", brokerResultMessage(result), fmt.Errorf("aliceblue order placed response missing order id")
		}
		return orderID, firstNonEmpty(brokerResultMessage(result), "Order placed successfully"), nil

	default:
		return "", "", fmt.Errorf("unsupported broker: %s", bc.brokerName)
	}
}

func extractAliceResultOrderID(result map[string]any) string {
	results, ok := result["result"].([]any)
	if !ok || len(results) == 0 {
		return ""
	}
	first, ok := results[0].(map[string]any)
	if !ok {
		return ""
	}
	return extractBrokerOrderID(first, "brokerOrderId", "order_id", "orderid")
}

func extractBrokerOrderID(result map[string]any, keys ...string) string {
	if result == nil {
		return ""
	}
	for _, key := range keys {
		if v := valueAsString(result[key]); v != "" {
			return v
		}
	}
	return ""
}

func brokerResultMessage(result map[string]any) string {
	if result == nil {
		return ""
	}
	if msg := firstNonEmpty(
		valueAsString(result["message"]),
		valueAsString(result["emsg"]),
		valueAsString(result["error"]),
	); msg != "" {
		return msg
	}
	if data := nestedMap(result, "data"); data != nil {
		if msg := firstNonEmpty(valueAsString(data["message"]), valueAsString(data["emsg"])); msg != "" {
			return msg
		}
	}
	if results, ok := result["result"].([]any); ok && len(results) > 0 {
		if first, ok := results[0].(map[string]any); ok {
			return firstNonEmpty(valueAsString(first["message"]), valueAsString(first["emsg"]))
		}
	}
	return ""
}

func nestedMap(result map[string]any, key string) map[string]any {
	if result == nil {
		return nil
	}
	if m, ok := result[key].(map[string]any); ok {
		return m
	}
	return nil
}

func valueAsString(v any) string {
	switch x := v.(type) {
	case string:
		return strings.TrimSpace(x)
	case nil:
		return ""
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", x))
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
