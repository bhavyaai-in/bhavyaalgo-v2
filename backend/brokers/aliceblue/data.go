package aliceblue

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
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
	// Convert time strings to millisecond timestamps
	layout := "2006-01-02 15:04"
	fromTime, err1 := time.Parse(layout, from)
	toTime, err2 := time.Parse(layout, to)
	fromMs := from
	toMs := to
	if err1 == nil && err2 == nil {
		fromMs = strconv.FormatInt(fromTime.UnixMilli(), 10)
		toMs = strconv.FormatInt(toTime.UnixMilli(), 10)
	}
	// Convert resolution: "1d" -> "D", "1m" -> "1", "5m" -> "5", etc.
	aliceRes := resolution
	if resolution == "1d" {
		aliceRes = "D"
	} else if len(resolution) > 1 && resolution[len(resolution)-1] == 'm' {
		aliceRes = resolution[:len(resolution)-1]
	}
	payload := map[string]string{
		"token":      token,
		"resolution": aliceRes,
		"from":       fromMs,
		"to":         toMs,
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
