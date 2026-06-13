package blueprints

import (
	"context"
	"fmt"
	"log"
	"time"

	"bhavyaaialgo/backend/brokers/aliceblue"
	"bhavyaaialgo/backend/brokers/angel"
)

// fetchLiveLTP fetches the live Last Traded Price (LTP) from a connected broker for a given symbol, exchange, and token.
// It tries the REST API first, and if it fails, falls back to the active WebSocket stream tick cache.
func (a *App) fetchLiveLTP(ctx context.Context, exchange, symbol, token string) (float64, error) {
	ltp, restErr := a.fetchLiveLTPFromREST(ctx, exchange, symbol, token)
	if restErr == nil && ltp > 0 {
		return ltp, nil
	}

	// Fallback to WebSocket cache
	if a.Hub != nil {
		if wsLtp, ok := a.Hub.GetLTP(token); ok && wsLtp > 0 {
			log.Printf("fetchLiveLTP REST failed (%v); fell back to WebSocket tick cache for token %s: %f", restErr, token, wsLtp)
			return wsLtp, nil
		}
	}

	return 0, fmt.Errorf("failed to fetch ltp from REST (%w) and no websocket tick cached", restErr)
}

// fetchLiveLTPFromREST fetches the live Last Traded Price (LTP) from a connected broker for a given symbol, exchange, and token via the broker REST API.
func (a *App) fetchLiveLTPFromREST(ctx context.Context, exchange, symbol, token string) (float64, error) {
	todayPattern := time.Now().Format("2006-01-02") + "%"
	brokerRow := a.TradingDB.QueryRowContext(ctx,
		`SELECT id, broker_name, broker_token, broker_api, broker_api_secret 
		 FROM brokers 
		 WHERE token_status='connected' AND broker_token_date LIKE ? 
		 ORDER BY broker_token_date DESC 
		 LIMIT 1`, todayPattern)
	var brokerID int64
	var brokerName, brokerToken, brokerAPI, brokerAPISecret string
	if err := brokerRow.Scan(&brokerID, &brokerName, &brokerToken, &brokerAPI, &brokerAPISecret); err != nil {
		return 0, fmt.Errorf("no connected broker found: %w", err)
	}

	var resp map[string]any
	var err error

	switch brokerName {
	case "angel":
		ac := angel.NewClient(brokerAPI)
		resp, err = ac.GetQuote(brokerToken, exchange, symbol, token)
		if err != nil {
			return 0, err
		}
		if dm, ok := resp["data"].(map[string]any); ok {
			if fetched, ok := dm["fetched"].([]any); ok && len(fetched) > 0 {
				if m, ok := fetched[0].(map[string]any); ok {
					q := parseAngelQuote(m)
					if q.ltp > 0 {
						return q.ltp, nil
					}
				}
			}
		}
	case "aliceblue":
		ac := aliceblue.NewClient(brokerAPI, brokerAPISecret)
		resp, err = ac.GetQuote(brokerToken, exchange, symbol, symbol)
		if err != nil {
			resp, err = ac.GetQuote(brokerToken, exchange, symbol, token)
		}
		if err != nil {
			return 0, err
		}
		var q quoteResult
		if resultList, ok := resp["result"].([]any); ok && len(resultList) > 0 {
			if firstItem, ok := resultList[0].(map[string]any); ok {
				q = parseAliceQuote(firstItem)
			}
		} else {
			q = parseAliceQuote(resp)
		}
		if q.ltp > 0 {
			return q.ltp, nil
		}
	default:
		return 0, fmt.Errorf("unsupported broker: %s", brokerName)
	}

	return 0, fmt.Errorf("failed to parse ltp from quote response: %v", resp)
}
