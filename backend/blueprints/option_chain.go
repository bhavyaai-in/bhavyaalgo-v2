package blueprints

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"bhavyaaialgo/backend/brokers/aliceblue"
	"bhavyaaialgo/backend/brokers/angel"
)

func (a *App) RegisterOptionChainRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/option-chain", a.authMiddleware(a.handleOptionChain))
	mux.HandleFunc("GET /api/option-chain/underlyings", a.authMiddleware(a.handleOptionChainUnderlyings))
	mux.HandleFunc("GET /api/option-chain/expiries", a.authMiddleware(a.handleOptionChainExpiries))
}

func extractUnderlying(symbol string) string {
	// Known index underlyings that may contain digits
	knownPrefixes := []string{"NIFTYNXT50", "SENSEX50"}
	for _, p := range knownPrefixes {
		if strings.HasPrefix(symbol, p) {
			return p
		}
	}
	// Indian F&O symbol: UNDERLYING + DDMMMYY + STRIKE + CE/PE
	// Extract leading uppercase letters that form the underlying name
	var i int
	for i < len(symbol) && symbol[i] >= 'A' && symbol[i] <= 'Z' {
		i++
	}
	if i == 0 {
		return ""
	}
	return symbol[:i]
}

func (a *App) handleOptionChainUnderlyings(w http.ResponseWriter, r *http.Request) {
	exchange := r.URL.Query().Get("exchange")
	if exchange == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "exchange is required"})
		return
	}
	rows, err := a.MarketDB.QueryContext(r.Context(), `
		SELECT DISTINCT symbol FROM master_contracts
		WHERE exchange = ? AND instrumenttype LIKE 'OPT%' AND expiry != ''
		AND (symbol LIKE '%CE' OR symbol LIKE '%PE')
		ORDER BY symbol
	`, exchange)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	defer rows.Close()
	seen := make(map[string]bool)
	var underlyings []string
	for rows.Next() {
		var sym string
		if err := rows.Scan(&sym); err == nil {
			if u := extractUnderlying(sym); u != "" && !seen[u] {
				seen[u] = true
				underlyings = append(underlyings, u)
			}
		}
	}
	writeJSON(w, http.StatusOK, underlyings)
}

func (a *App) handleOptionChainExpiries(w http.ResponseWriter, r *http.Request) {
	exchange := r.URL.Query().Get("exchange")
	underlying := r.URL.Query().Get("underlying")
	if exchange == "" || underlying == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "exchange and underlying are required"})
		return
	}
	rows, err := a.MarketDB.QueryContext(r.Context(), `
		SELECT DISTINCT expiry FROM master_contracts
		WHERE exchange = ? AND symbol LIKE ? AND instrumenttype LIKE 'OPT%'
		AND expiry != '' AND (symbol LIKE '%CE' OR symbol LIKE '%PE')
		ORDER BY expiry
	`, exchange, underlying+"%")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	defer rows.Close()
	var expiries []string
	for rows.Next() {
		var e string
		if err := rows.Scan(&e); err == nil && e != "" {
			expiries = append(expiries, e)
		}
	}
	writeJSON(w, http.StatusOK, expiries)
}

type symAndTok struct {
	Symbol string
	Token  string
}

type optionChainReq struct {
	Underlying  string `json:"underlying"`
	Exchange    string `json:"exchange"`
	Expiry      string `json:"expiry"`
	StrikeCount int    `json:"strike_count"`
}

type strikeData struct {
	Strike float64                `json:"strike"`
	CE     *optionStrikeDetail    `json:"ce,omitempty"`
	PE     *optionStrikeDetail    `json:"pe,omitempty"`
}

type optionStrikeDetail struct {
	Symbol   string  `json:"symbol"`
	Token    string  `json:"token"`
	Label    string  `json:"label"`
	LTP      float64 `json:"ltp"`
	Bid      float64 `json:"bid"`
	Ask      float64 `json:"ask"`
	BidQty   int64   `json:"bid_qty"`
	AskQty   int64   `json:"ask_qty"`
	Open     float64 `json:"open"`
	High     float64 `json:"high"`
	Low      float64 `json:"low"`
	Close    float64 `json:"close"`
	Volume   int64   `json:"volume"`
	OI       int64   `json:"oi"`
	Lotsize  int64   `json:"lotsize"`
	TickSize float64 `json:"tick_size"`
}

func (a *App) handleOptionChain(w http.ResponseWriter, r *http.Request) {
	underlying := r.URL.Query().Get("underlying")
	exchange := r.URL.Query().Get("exchange")
	expiry := r.URL.Query().Get("expiry")
	strikeCountStr := r.URL.Query().Get("strike_count")

	if underlying == "" || exchange == "" || expiry == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "underlying, exchange, and expiry are required"})
		return
	}

	strikeCount := 10
	if strikeCountStr != "" {
		if v, err := strconv.Atoi(strikeCountStr); err == nil && v > 0 {
			strikeCount = v
		}
	}

	// Determine the options exchange and quote exchange
	optionsExchange := exchange
	quoteExchange := exchange
	if exchange == "NSE" || exchange == "NSE_INDEX" || exchange == "NFO" {
		optionsExchange = "NFO"
		if underlying == "NIFTY" || underlying == "BANKNIFTY" || underlying == "FINNIFTY" || underlying == "MIDCPNIFTY" {
			quoteExchange = "NSE_INDEX"
		} else {
			quoteExchange = "NSE"
		}
	} else if exchange == "BSE" || exchange == "BSE_INDEX" || exchange == "BFO" {
		optionsExchange = "BFO"
		quoteExchange = "BSE_INDEX"
	} else if exchange == "MCX" {
		optionsExchange = "MCX"
		quoteExchange = "MCX"
	} else if exchange == "CDS" {
		optionsExchange = "CDS"
		quoteExchange = "CDS"
	}

	// Get underlying LTP (non-fatal — chain still returned without ATM info)
	underlyingLTP, underlyingClose, underlyingToken, _ := a.getUnderlyingPrice(r.Context(), underlying, quoteExchange)

	// Expiry format: master_contracts stores in DD-MMM-YY (e.g. "30-JUN-26")
	// User sends DDMMMYY (e.g. "30JUN26") or DD-MMM-YY
	expiryFormatted := expiry
	if !strings.Contains(expiry, "-") {
		if len(expiry) >= 7 {
			expiryFormatted = expiry[:2] + "-" + expiry[2:5] + "-" + expiry[5:]
		}
	}
	expiryFormatted = strings.ToUpper(expiryFormatted)

	// Build symbol prefix: UNDERLYING + DDMMMYY (e.g. NIFTY30JUN26)
	symbolPrefix := underlying + strings.ToUpper(expiry)
	symbolPrefix = strings.ReplaceAll(symbolPrefix, "-", "")
	rows, err := a.MarketDB.QueryContext(r.Context(), `
		SELECT symbol, token, strike, lotsize, tick_size, instrumenttype
		FROM master_contracts
		WHERE exchange = ? AND expiry = ? AND symbol LIKE ? AND (instrumenttype LIKE 'OPT%')
		ORDER BY strike
	`, optionsExchange, expiryFormatted, symbolPrefix+"%")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	defer rows.Close()

	type optionRecord struct {
		symbol         string
		token          string
		strike         float64
		optType        string
		lotsize        int64
		tickSize       float64
	}
	var allOptions []optionRecord
	allStrikes := make(map[float64]bool)

	for rows.Next() {
		var symbol, token, instType string
		var strike float64
		var lotsize int64
		var tickSize float64
		if err := rows.Scan(&symbol, &token, &strike, &lotsize, &tickSize, &instType); err != nil {
			continue
		}
		optType := ""
		if strings.HasSuffix(symbol, "CE") {
			optType = "CE"
		} else if strings.HasSuffix(symbol, "PE") {
			optType = "PE"
		} else {
			continue
		}
		allOptions = append(allOptions, optionRecord{symbol, token, strike, optType, lotsize, tickSize})
		allStrikes[strike] = true
	}

	if len(allOptions) == 0 {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "no options found for " + underlying + " at " + expiryFormatted})
		return
	}

	// Get sorted unique strikes
	var uniqueStrikes []float64
	for s := range allStrikes {
		uniqueStrikes = append(uniqueStrikes, s)
	}
	sort.Float64s(uniqueStrikes)

	// Find ATM strike (use median fallback if LTP unavailable)
	atmStrike := underlyingLTP
	if atmStrike <= 0 {
		atmStrike = uniqueStrikes[len(uniqueStrikes)/2] // median strike
	}
	atmStrike = findATMStrike(atmStrike, uniqueStrikes)

	// Select strikes around ATM
	atmIdx := sort.SearchFloat64s(uniqueStrikes, atmStrike)
	start := atmIdx - strikeCount
	if start < 0 {
		start = 0
	}
	end := atmIdx + strikeCount + 1
	if end > len(uniqueStrikes) {
		end = len(uniqueStrikes)
	}
	selectedStrikes := uniqueStrikes[start:end]

	// Build CE/PE maps
	type optionInfo struct {
		symbol   string
		token    string
		lotsize  int64
		tickSize float64
	}
	ceMap := make(map[float64]optionInfo)
	peMap := make(map[float64]optionInfo)
	for _, o := range allOptions {
		info := optionInfo{o.symbol, o.token, o.lotsize, o.tickSize}
		if o.optType == "CE" {
			ceMap[o.strike] = info
		} else {
			peMap[o.strike] = info
		}
	}

	// Fetch quotes via connected broker (use option SYMBOL for cross-broker compat)
	var allSymbols []symAndTok
	for _, s := range selectedStrikes {
		if ce, ok := ceMap[s]; ok {
			allSymbols = append(allSymbols, symAndTok{ce.symbol, ce.token})
		}
		if pe, ok := peMap[s]; ok {
			allSymbols = append(allSymbols, symAndTok{pe.symbol, pe.token})
		}
	}

	quoteData, quoteErr := a.fetchOptionQuotes(r.Context(), allSymbols, optionsExchange)

	// Build chain response
	var chain []strikeData
	for _, s := range selectedStrikes {
		item := strikeData{Strike: s}

		if ce, ok := ceMap[s]; ok {
			label := calcLabel(s, atmStrike, uniqueStrikes, "CE")
			item.CE = &optionStrikeDetail{
				Symbol:   ce.symbol,
				Token:    ce.token,
				Label:    label,
				Lotsize:  ce.lotsize,
				TickSize: ce.tickSize,
			}
			if quoteErr == nil {
				// Try token key first, then symbol key (cross-broker compat)
				q, found := quoteData[ce.token]
				if !found {
					q, found = quoteData[ce.symbol]
				}
				if found {
					item.CE.LTP = q.ltp
					item.CE.Bid = q.bid
					item.CE.Ask = q.ask
					item.CE.BidQty = q.bidQty
					item.CE.AskQty = q.askQty
					item.CE.Open = q.open
					item.CE.High = q.high
					item.CE.Low = q.low
					item.CE.Close = q.close
					item.CE.Volume = q.volume
					item.CE.OI = q.oi
				}
			}
		}

		if pe, ok := peMap[s]; ok {
			label := calcLabel(s, atmStrike, uniqueStrikes, "PE")
			item.PE = &optionStrikeDetail{
				Symbol:   pe.symbol,
				Token:    pe.token,
				Label:    label,
				Lotsize:  pe.lotsize,
				TickSize: pe.tickSize,
			}
			if quoteErr == nil {
				q, found := quoteData[pe.token]
				if !found {
					q, found = quoteData[pe.symbol]
				}
				if found {
					item.PE.LTP = q.ltp
					item.PE.Bid = q.bid
					item.PE.Ask = q.ask
					item.PE.BidQty = q.bidQty
					item.PE.AskQty = q.askQty
					item.PE.Open = q.open
					item.PE.High = q.high
					item.PE.Low = q.low
					item.PE.Close = q.close
					item.PE.Volume = q.volume
					item.PE.OI = q.oi
				}
			}
		}

		chain = append(chain, item)
	}

	// Calculate PCR
	var totalCeOI, totalPeOI int64
	for _, item := range chain {
		if item.CE != nil {
			totalCeOI += item.CE.OI
		}
		if item.PE != nil {
			totalPeOI += item.PE.OI
		}
	}
	pcr := 0.0
	if totalCeOI > 0 {
		pcr = float64(totalPeOI) / float64(totalCeOI)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"underlying":       underlying,
		"underlying_ltp":   underlyingLTP,
		"underlying_close": underlyingClose,
		"underlying_token":  underlyingToken,
		"expiry":           expiryFormatted,
		"atm_strike":       atmStrike,
		"pcr":              math.Round(pcr*100) / 100,
		"chain":            chain,
	})
}

type quoteResult struct {
	ltp    float64
	bid    float64
	ask    float64
	bidQty int64
	askQty int64
	open   float64
	high   float64
	low    float64
	close  float64
	volume int64
	oi     int64
}

func (a *App) fetchOptionQuotes(ctx context.Context, symbols []symAndTok, exchange string) (map[string]quoteResult, error) {
	brokerRow := a.TradingDB.QueryRowContext(ctx,
		`SELECT id, broker_name, broker_token, feed_token, broker_api, broker_api_secret FROM brokers WHERE token_status='connected' LIMIT 1`)
	var brokerID int64
	var brokerName, brokerToken, feedToken, brokerAPI, brokerAPISecret string
	if err := brokerRow.Scan(&brokerID, &brokerName, &brokerToken, &feedToken, &brokerAPI, &brokerAPISecret); err != nil {
		return nil, fmt.Errorf("no connected broker found")
	}

	var allSymbols = symbols

	result := make(map[string]quoteResult)

	switch brokerName {
		case "angel":
		// Angel uses tokens in exchangeTokens format
		allTokens := make(map[string][]string)
		for _, s := range allSymbols {
			allTokens[exchange] = append(allTokens[exchange], s.Token)
		}
		ac := angel.NewClient(brokerAPI)
		resp, err := ac.GetMultiQuote(brokerToken, allTokens)
		if err != nil {
			return nil, err
		}
		if dataMap, ok := resp["data"].(map[string]any); ok {
			if fetched, ok := dataMap["fetched"].([]any); ok {
				for _, item := range fetched {
					if m, ok := item.(map[string]any); ok {
						q := parseAngelQuote(m)
						if q.ltp > 0 {
							if sym, ok := m["tradingSymbol"].(string); ok {
								result[sym] = q
							}
							if tok, ok := m["symbolToken"]; ok {
								result[fmt.Sprintf("%v", tok)] = q
							}
						}
					}
				}
			}
		}
	case "aliceblue":
		// Alice Blue — try symbols first (exchange:symbol)
		ac := aliceblue.NewClient(brokerAPI, brokerAPISecret)
		var symList []string
		for _, s := range allSymbols {
			symList = append(symList, exchange+":"+s.Symbol)
		}
		resp, err := ac.GetMultiQuote(brokerToken, symList)
		if err != nil {
			return nil, err
		}
		if fetched, ok := resp["result"].([]any); ok {
			for _, item := range fetched {
				if m, ok := item.(map[string]any); ok {
					q := parseAliceQuote(m)
					if q.ltp > 0 {
						if sym, ok := m["symbol"].(string); ok {
							result[sym] = q
						}
						if tok, ok := m["token"].(string); ok {
							result[tok] = q
						}
					}
				}
			}
		}
	default:
		return nil, fmt.Errorf("unsupported broker: %s", brokerName)
	}

	return result, nil
}

func parseAngelQuote(m map[string]any) quoteResult {
	f := func(key string) float64 {
		if v, ok := m[key]; ok {
			switch val := v.(type) {
			case string:
				f, _ := strconv.ParseFloat(val, 64)
				return f
			case float64:
				return val
			}
		}
		return 0
	}
	i := func(key string) int64 {
		if v, ok := m[key]; ok {
			switch val := v.(type) {
			case string:
				n, _ := strconv.ParseInt(val, 10, 64)
				return n
			case float64:
				return int64(val)
			}
		}
		return 0
	}
	// Extract bid/ask from depth if available
	var bid, ask float64
	if depth, ok := m["depth"].(map[string]any); ok {
		if buy, ok := depth["buy"].([]any); ok && len(buy) > 0 {
			if first, ok := buy[0].(map[string]any); ok {
				bid = fFrom(first, "price")
			}
		}
		if sell, ok := depth["sell"].([]any); ok && len(sell) > 0 {
			if first, ok := sell[0].(map[string]any); ok {
				ask = fFrom(first, "price")
			}
		}
	}
	return quoteResult{
		ltp:    f("ltp"),
		bid:    bid,
		ask:    ask,
		bidQty: i("totBuyQuan"),
		askQty: i("totSellQuan"),
		open:   f("open"),
		high:   f("high"),
		low:    f("low"),
		close:  f("close"),
		volume: i("tradeVolume"),
		oi:     i("opnInterest"),
	}
}

func fFrom(m map[string]any, key string) float64 {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case string:
			f, _ := strconv.ParseFloat(val, 64)
			return f
		case float64:
			return val
		}
	}
	return 0
}

func parseAliceQuote(m map[string]any) quoteResult {
	// Alice response format varies; try common paths
	var data map[string]any
	if d, ok := m["data"].(map[string]any); ok {
		data = d
	} else {
		data = m
	}
	f := func(key string) float64 {
		if v, ok := data[key]; ok {
			switch val := v.(type) {
			case string:
				f, _ := strconv.ParseFloat(val, 64)
				return f
			case float64:
				return val
			}
		}
		return 0
	}
	i := func(key string) int64 {
		if v, ok := data[key]; ok {
			switch val := v.(type) {
			case string:
				n, _ := strconv.ParseInt(val, 10, 64)
				return n
			case float64:
				return int64(val)
			}
		}
		return 0
	}
	return quoteResult{
		ltp:    f("ltp"),
		bid:    f("bid"),
		ask:    f("ask"),
		bidQty: i("bid_qty"),
		askQty: i("ask_qty"),
		open:   f("open"),
		high:   f("high"),
		low:    f("low"),
		close:  f("close"),
		volume: i("volume"),
		oi:     i("oi"),
	}
}

func (a *App) getUnderlyingPrice(ctx context.Context, symbol, exchange string) (float64, float64, string, error) {
	// Try exact exchange first, then try NSE for indices, then try any exchange
	// Exclude options (OPT*) and futures (FUT*) — only get the underlying contract
	var token, brExchange string
	err := a.MarketDB.QueryRowContext(ctx,
		`SELECT token, brexchange FROM master_contracts
		 WHERE symbol = ? AND (exchange = ? OR exchange = ?)
		 AND instrumenttype NOT LIKE 'OPT%' AND instrumenttype NOT LIKE 'FUT%'
		 LIMIT 1`,
		symbol, exchange, strings.TrimSuffix(exchange, "_INDEX"),
	).Scan(&token, &brExchange)
	if err != nil {
		// Broader fallback — any non-option/future row
		err = a.MarketDB.QueryRowContext(ctx,
			`SELECT token, brexchange FROM master_contracts
			 WHERE symbol = ? AND instrumenttype NOT LIKE 'OPT%' AND instrumenttype NOT LIKE 'FUT%'
			 LIMIT 1`,
			symbol,
		).Scan(&token, &brExchange)
		if err != nil {
			// Last resort: check if it's a known NSE index with a well-known token
			return a.getIndexPrice(ctx, symbol)
		}
	}
	// For index tokens, use the brexchange from DB (e.g. "NSE" for NIFTY)
	quoteExchange := brExchange
	if quoteExchange == "" {
		quoteExchange = exchange
	}

	brokerRow := a.TradingDB.QueryRowContext(ctx,
		`SELECT id, broker_name, broker_token FROM brokers WHERE token_status='connected' LIMIT 1`)
	var brokerID int64
	var brokerName, brokerToken string
	if err := brokerRow.Scan(&brokerID, &brokerName, &brokerToken); err != nil {
		return 0, 0, "", fmt.Errorf("no connected broker")
	}

	var ltp, closeP float64

	switch brokerName {
	case "angel":
		var apiKey string
		a.TradingDB.QueryRowContext(ctx, `SELECT broker_api FROM brokers WHERE id = ?`, brokerID).Scan(&apiKey)
		ac := angel.NewClient(apiKey)
		resp, err := ac.GetQuote(brokerToken, quoteExchange, symbol, token)
		if err != nil {
			return 0, 0, "", err
		}
		if dm, ok := resp["data"].(map[string]any); ok {
			if fetched, ok := dm["fetched"].([]any); ok && len(fetched) > 0 {
				if m, ok := fetched[0].(map[string]any); ok {
					q := parseAngelQuote(m)
					ltp = q.ltp
					closeP = q.close
				}
			}
		}
	case "aliceblue":
		var apiKey, apiSecret string
		a.TradingDB.QueryRowContext(ctx, `SELECT broker_api, broker_api_secret FROM brokers WHERE id = ?`, brokerID).Scan(&apiKey, &apiSecret)
		ac := aliceblue.NewClient(apiKey, apiSecret)
		// Alice Blue may accept symbol (exchange:symbol) instead of token
		resp, err := ac.GetQuote(brokerToken, quoteExchange, symbol, symbol)
		if err != nil {
			resp2, err2 := ac.GetQuote(brokerToken, quoteExchange, symbol, token)
			if err2 != nil {
				return 0, 0, "", err2
			}
			resp = resp2
		}
		q := parseAliceQuote(resp)
		ltp = q.ltp
		closeP = q.close
	default:
		return 0, 0, "", fmt.Errorf("unsupported broker: %s", brokerName)
	}

	if ltp <= 0 {
		return 0, 0, "", fmt.Errorf("failed to fetch LTP for %s", symbol)
	}
	return ltp, closeP, token, nil
}

// Known NSE index tokens (Angel tokens) for underlying price lookup
var knownIndexTokens = map[string]string{
	"NIFTY":     "99926000",
	"BANKNIFTY": "99926009",
	"FINNIFTY":  "99926037",
	"MIDCPNIFTY": "99926060",
	"NIFTYNXT50": "99926014",
	"SENSEX":     "99919001",
	"SENSEX50":   "99919007",
	"BANKEX":     "99919012",
	"NIFTYIT":    "99926012",
	"NIFTYPHARMA": "99926025",
	"NIFTYBANK":  "99926015",
}

func (a *App) getIndexPrice(ctx context.Context, symbol string) (float64, float64, string, error) {
	token, ok := knownIndexTokens[symbol]
	if !ok {
		return 0, 0, "", fmt.Errorf("no known token for %s", symbol)
	}
	brokerRow := a.TradingDB.QueryRowContext(ctx,
		`SELECT broker_name, broker_token, broker_api, broker_api_secret FROM brokers WHERE token_status='connected' LIMIT 1`)
	var brokerName, brokerToken, brokerAPI, brokerAPISecret string
	if err := brokerRow.Scan(&brokerName, &brokerToken, &brokerAPI, &brokerAPISecret); err != nil {
		return 0, 0, "", fmt.Errorf("no connected broker")
	}

	switch brokerName {
	case "angel":
		ac := angel.NewClient(brokerAPI)
		resp, err := ac.GetQuote(brokerToken, "NSE", symbol, token)
		if err != nil {
			return 0, 0, "", err
		}
		if dm, ok := resp["data"].(map[string]any); ok {
			if fetched, ok := dm["fetched"].([]any); ok && len(fetched) > 0 {
				if m, ok := fetched[0].(map[string]any); ok {
					q := parseAngelQuote(m)
					return q.ltp, q.close, token, nil
				}
			}
		}
	case "aliceblue":
		ac := aliceblue.NewClient(brokerAPI, brokerAPISecret)
		// Use symbol instead of token for Alice Blue (cross-broker compat)
		resp, err := ac.GetQuote(brokerToken, "NSE", symbol, symbol)
		if err != nil {
			resp, err = ac.GetQuote(brokerToken, "NSE", symbol, token)
			if err != nil {
				return 0, 0, "", err
			}
		}
		q := parseAliceQuote(resp)
		return q.ltp, q.close, token, nil
	}
	return 0, 0, "", fmt.Errorf("unable to fetch index price for %s", symbol)
}

func findATMStrike(ltp float64, strikes []float64) float64 {
	if len(strikes) == 0 {
		return 0
	}
	// Find the closest strike to LTP
	idx := sort.SearchFloat64s(strikes, ltp)
	if idx >= len(strikes) {
		return strikes[len(strikes)-1]
	}
	if idx == 0 {
		return strikes[0]
	}
	// Check which is closer
	if ltp-strikes[idx-1] < strikes[idx]-ltp {
		return strikes[idx-1]
	}
	return strikes[idx]
}

func calcLabel(strike, atmStrike float64, allStrikes []float64, optType string) string {
	if strike == atmStrike {
		return "ATM"
	}
	atmIdx := sort.SearchFloat64s(allStrikes, atmStrike)
	strikeIdx := sort.SearchFloat64s(allStrikes, strike)
	diff := strikeIdx - atmIdx

	if optType == "CE" {
		if diff < 0 {
			return fmt.Sprintf("ITM%d", -diff)
		}
		return fmt.Sprintf("OTM%d", diff)
	}
	// PE
	if diff < 0 {
		return fmt.Sprintf("OTM%d", -diff)
	}
	return fmt.Sprintf("ITM%d", diff)
}


