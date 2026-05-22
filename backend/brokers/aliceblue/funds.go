package aliceblue

import (
	"encoding/json"
	"fmt"
)

type limitsResponse struct {
	Stat    string         `json:"status"`
	Message string         `json:"message"`
	Result  []limitsResult `json:"result"`
}

type limitsResult struct {
	TradingLimit       float64 `json:"tradingLimit"`
	OpeningCashLimit   float64 `json:"openingCashLimit"`
	IntradayPayin      float64 `json:"intradayPayin"`
	CollateralMargin   float64 `json:"collateralMargin"`
	CreditForSell      float64 `json:"creditForSell"`
	AdhocMargin        float64 `json:"adhocMargin"`
	UtilizedMargin     float64 `json:"utilizedMargin"`
	BlockedForPayout   float64 `json:"blockedForPayout"`
	UtilizedSpanMargin float64 `json:"utilizedSpanMargin"`
}

func (c *Client) GetMargin(sessionToken string) (map[string]string, error) {
	data, err := c.doAPIRequest("GET", "/open-api/od/v1/limits/", sessionToken, "")
	if err != nil {
		return nil, err
	}
	var resp limitsResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse limits response: %w", err)
	}
	if len(resp.Result) == 0 {
		return map[string]string{}, nil
	}
	r := resp.Result[0]
	return map[string]string{
		"trading_limit":        fmt.Sprintf("%.2f", r.TradingLimit),
		"opening_cash_limit":   fmt.Sprintf("%.2f", r.OpeningCashLimit),
		"collateral_margin":    fmt.Sprintf("%.2f", r.CollateralMargin),
		"credit_for_sell":      fmt.Sprintf("%.2f", r.CreditForSell),
		"utilized_margin":      fmt.Sprintf("%.2f", r.UtilizedMargin),
		"blocked_for_payout":   fmt.Sprintf("%.2f", r.BlockedForPayout),
		"utilized_span_margin": fmt.Sprintf("%.2f", r.UtilizedSpanMargin),
	}, nil
}
