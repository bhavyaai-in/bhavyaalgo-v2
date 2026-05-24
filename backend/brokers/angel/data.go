package angel

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
	Data [][]any `json:"data"`
}

type historicalCandle struct {
	Timestamp string
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    int
}

var TimeframeMap = map[string]string{
	"1m":  "ONE_MINUTE",
	"3m":  "THREE_MINUTE",
	"5m":  "FIVE_MINUTE",
	"10m": "TEN_MINUTE",
	"15m": "FIFTEEN_MINUTE",
	"30m": "THIRTY_MINUTE",
	"1h":  "ONE_HOUR",
	"1d":  "ONE_DAY",
}

func (c *Client) GetHistoricalData(authToken, symbol, exchange, interval, from, to string) ([]Candle, error) {
	resolution, ok := TimeframeMap[interval]
	if !ok {
		resolution = interval
	}
	payload := map[string]string{
		"exchange":       exchange,
		"symboltoken":    symbol,
		"interval":       resolution,
		"fromdate":       from,
		"todate":         to,
	}
	body, _ := json.Marshal(payload)
	data, err := c.doRequest("POST", BaseURL+"/rest/secure/angelbroking/historical/v1/getCandleData", authToken, string(body))
	if err != nil {
		return nil, err
	}
	var resp historicalResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse historical data: %w", err)
	}
	candles := make([]Candle, 0, len(resp.Data))
	for _, row := range resp.Data {
		if len(row) < 6 {
			continue
		}
		var ts string
		var o, h, l, c, v float64
		if t, ok := row[0].(string); ok {
			ts = t
		} else if t, ok := row[0].(float64); ok {
			ts = fmt.Sprintf("%v", t)
		}
		if val, ok := toFloat(row[1]); ok { o = val }
		if val, ok := toFloat(row[2]); ok { h = val }
		if val, ok := toFloat(row[3]); ok { l = val }
		if val, ok := toFloat(row[4]); ok { c = val }
		if val, ok := toFloat(row[5]); ok { v = val }
		candles = append(candles, Candle{Timestamp: ts, Open: o, High: h, Low: l, Close: c, Volume: int(v)})
	}
	return candles, nil
}

func toFloat(val any) (float64, bool) {
	switch v := val.(type) {
	case float64:
		return v, true
	case string:
		f, err := strconv.ParseFloat(v, 64)
		return f, err == nil
	case json.Number:
		f, err := v.Float64()
		return f, err == nil
	default:
		return 0, false
	}
}
