package historical

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"bhavyaaialgo/backend/brokers/aliceblue"
	"bhavyaaialgo/backend/brokers/angel"
)

type Candle struct {
	Timestamp string  `json:"timestamp"`
	Open      float64 `json:"open"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
	Close     float64 `json:"close"`
	Volume    int     `json:"volume"`
}

type Handler struct {
	MarketDB  *sql.DB
	TradingDB *sql.DB
}

var timeframes = []string{"1m", "3m", "5m", "10m", "15m", "30m", "1h", "1d"}

var angelTimeframes = map[string]string{
	"1m": "ONE_MINUTE", "3m": "THREE_MINUTE", "5m": "FIVE_MINUTE",
	"10m": "TEN_MINUTE", "15m": "FIFTEEN_MINUTE", "30m": "THIRTY_MINUTE",
	"1h": "ONE_HOUR", "1d": "ONE_DAY",
}

var aliceTimeframes = map[string]string{
	"1m": "1", "3m": "3", "5m": "5", "10m": "10", "15m": "15",
	"30m": "30", "1h": "60", "1d": "1440",
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/historical/exchanges", h.listExchanges)
	mux.HandleFunc("GET /api/historical/underlyings", h.listUnderlyings)
	mux.HandleFunc("POST /api/historical/download", h.download)
}

func (h *Handler) listExchanges(w http.ResponseWriter, r *http.Request) {
	rows, err := h.MarketDB.Query("SELECT DISTINCT exchange FROM master_contracts ORDER BY exchange")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()
	var exchanges []string
	for rows.Next() {
		var ex string
		if rows.Scan(&ex) == nil && ex != "" {
			exchanges = append(exchanges, ex)
		}
	}
	writeJSON(w, 200, exchanges)
}

func (h *Handler) listUnderlyings(w http.ResponseWriter, r *http.Request) {
	exchange := r.URL.Query().Get("exchange")
	if exchange == "" {
		http.Error(w, "exchange required", 400)
		return
	}
	rows, err := h.MarketDB.Query(
		"SELECT DISTINCT symbol FROM master_contracts WHERE exchange = ? AND instrumenttype NOT LIKE 'OPT%' AND instrumenttype NOT LIKE 'FUT%' AND symbol NOT LIKE '%/%' AND LENGTH(symbol) < 20 AND symbol NOT LIKE '0%' AND symbol NOT LIKE '1%' AND symbol NOT LIKE '2%' AND symbol NOT LIKE '3%' AND symbol NOT LIKE '4%' AND symbol NOT LIKE '5%' AND symbol NOT LIKE '6%' AND symbol NOT LIKE '7%' AND symbol NOT LIKE '8%' AND symbol NOT LIKE '9%' ORDER BY symbol LIMIT 500",
		exchange,
	)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()
	var symbols []string
	for rows.Next() {
		var sym string
		if rows.Scan(&sym) == nil && sym != "" {
			symbols = append(symbols, sym)
		}
	}
	writeJSON(w, 200, symbols)
}

type downloadReq struct {
	Symbol   string `json:"symbol"`
	Exchange string `json:"exchange"`
	Interval string `json:"interval"`
	From     string `json:"from"`
	To       string `json:"to"`
}

func (h *Handler) download(w http.ResponseWriter, r *http.Request) {
	var req downloadReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", 400)
		return
	}
	if req.Symbol == "" || req.Exchange == "" || req.Interval == "" || req.From == "" || req.To == "" {
		http.Error(w, "symbol, exchange, interval, from, to required", 400)
		return
	}

	var brokerName, brokerToken, brokerAPI, brokerAPISecret string
	err := h.TradingDB.QueryRow(
		"SELECT broker_name, broker_token, broker_api, COALESCE(broker_api_secret,'') FROM brokers WHERE token_status='connected' LIMIT 1",
	).Scan(&brokerName, &brokerToken, &brokerAPI, &brokerAPISecret)
	if err != nil {
		http.Error(w, "no connected broker", 400)
		return
	}

	var token string
	err = h.MarketDB.QueryRow(
		"SELECT token FROM master_contracts WHERE symbol = ? AND exchange = ? LIMIT 1",
		req.Symbol, req.Exchange,
	).Scan(&token)
	if err != nil {
		http.Error(w, "symbol not found on exchange", 404)
		return
	}

	var candles []Candle
	switch brokerName {
	case "angel":
		ac := angel.NewClient(brokerAPI)
		resolved := req.Interval
		if r, ok := angelTimeframes[req.Interval]; ok {
			resolved = r
		}
		raw, e := ac.GetHistoricalData(brokerToken, token, req.Exchange, resolved, req.From, req.To)
		if e != nil {
			err = e
			break
		}
		for _, c := range raw {
			candles = append(candles, Candle{c.Timestamp, c.Open, c.High, c.Low, c.Close, c.Volume})
		}
	case "aliceblue":
		ac := aliceblue.NewClient(brokerAPI, brokerAPISecret)
		resolved := req.Interval
		if r, ok := aliceTimeframes[req.Interval]; ok {
			resolved = r
		}
		raw, e := ac.GetHistoricalData(brokerToken, token, resolved, req.From, req.To, req.Exchange)
		if e != nil {
			err = e
			break
		}
		for _, c := range raw {
			candles = append(candles, Candle{c.Timestamp, c.Open, c.High, c.Low, c.Close, c.Volume})
		}
	default:
		http.Error(w, "unsupported broker: "+brokerName, 400)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if candles == nil || len(candles) == 0 {
		// Check if this is an index (AMXIDX type) - historical data not supported for indices
		var instType string
		h.MarketDB.QueryRow("SELECT instrumenttype FROM master_contracts WHERE symbol = ? AND exchange = ? LIMIT 1", req.Symbol, req.Exchange).Scan(&instType)
		if instType == "AMXIDX" {
			writeJSON(w, 200, map[string]any{
				"symbol": req.Symbol, "exchange": req.Exchange, "interval": req.Interval,
				"from": req.From, "to": req.To, "count": 0, "candles": []Candle{},
				"warning": "Historical data not available for index symbols. Only stock/equity instruments are supported.",
			})
			return
		}
	}
	if candles == nil {
		candles = []Candle{}
	}

	writeJSON(w, 200, map[string]any{
		"symbol":   req.Symbol,
		"exchange": req.Exchange,
		"interval": req.Interval,
		"from":     req.From,
		"to":       req.To,
		"count":    len(candles),
		"candles":  candles,
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
