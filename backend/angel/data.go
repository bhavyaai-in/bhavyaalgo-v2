package angel

import (
	"encoding/json"
	"fmt"
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
	Data []historicalCandle `json:"data"`
}

type historicalCandle struct {
	Timestamp string `json:"timestamp"`
	Open      string `json:"open"`
	High      string `json:"high"`
	Low       string `json:"low"`
	Close     string `json:"close"`
	Volume    int    `json:"volume"`
}

var TimeframeMap = map[string]string{
	"1m":  "ONE_MINUTE",
	"3m":  "THREE_MINUTE",
	"5m":  "FIVE_MINUTE",
	"10m": "TEN_MINUTE",
	"15m": "FIFTEEN_MINUTE",
	"30m": "THIRTY_MINUTE",
	"1h":  "ONE_HOUR",
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
	for _, c := range resp.Data {
		candles = append(candles, Candle{
			Timestamp: c.Timestamp,
			Open:      parseFloat(c.Open),
			High:      parseFloat(c.High),
			Low:       parseFloat(c.Low),
			Close:     parseFloat(c.Close),
			Volume:    c.Volume,
		})
	}
	return candles, nil
}
