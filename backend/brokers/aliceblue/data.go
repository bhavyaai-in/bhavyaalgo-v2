package aliceblue

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type Candle struct {
	Timestamp string  `json:"timestamp"`
	Open      float64 `json:"open"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
	Close     float64 `json:"close"`
	Volume    int     `json:"volume"`
}

type historicalResponse struct {
	Stat   string              `json:"stat"`
	Result [][]json.RawMessage `json:"result"`
}

func (c *Client) GetHistoricalData(sessionToken, token, resolution, from, to, exchange string) ([]Candle, error) {
	payload := map[string]string{
		"token":      token,
		"resolution": resolution,
		"from":       from,
		"to":         to,
		"exchange":   exchange,
	}
	body, _ := json.Marshal(payload)
	data, err := c.doAPIRequest("POST", "/open-api/od/ChartAPIService/api/chart/history", sessionToken, string(body))
	if err != nil {
		return nil, err
	}

	var resp historicalResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse historical data: %w", err)
	}
	if resp.Stat != "Ok" {
		var errResp struct {
			Emsg string `json:"emsg"`
		}
		json.Unmarshal(data, &errResp)
		if errResp.Emsg != "" {
			return nil, fmt.Errorf("historical data error: %s", errResp.Emsg)
		}
		return nil, fmt.Errorf("historical data request failed")
	}

	candles := make([]Candle, 0, len(resp.Result))
	for _, r := range resp.Result {
		if len(r) < 6 {
			continue
		}
		var ts string
		var o, h, l, c float64
		var v int64
		if err := json.Unmarshal(r[0], &ts); err != nil {
			continue
		}
		if err := json.Unmarshal(r[1], &o); err != nil {
			continue
		}
		if err := json.Unmarshal(r[2], &h); err != nil {
			continue
		}
		if err := json.Unmarshal(r[3], &l); err != nil {
			continue
		}
		if err := json.Unmarshal(r[4], &c); err != nil {
			continue
		}
		if err := json.Unmarshal(r[5], &v); err != nil {
			continue
		}
		candles = append(candles, Candle{
			Timestamp: ts,
			Open:      o,
			High:      h,
			Low:       l,
			Close:     c,
			Volume:    int(v),
		})
	}
	return candles, nil
}

func parseFloat(s string) float64 {
	v, _ := strconv.ParseFloat(s, 64)
	return v
}
