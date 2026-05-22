package aliceblue

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

type orderResponse struct {
	Status  string        `json:"status"`
	Message string        `json:"message"`
	Result  []orderResult `json:"result"`
}

type orderResult struct {
	BrokerOrderID string `json:"brokerOrderId"`
	RequestTime   string `json:"requestTime"`
}

func normalizeOrderPayload(payload map[string]any) map[string]any {
	out := map[string]any{}

	if v, ok := getStr(payload, "exchange"); ok {
		out["exchange"] = v
	}

	if v, ok := getStr(payload, "symboltoken"); ok {
		out["instrumentId"] = v
	} else if v, ok := getStr(payload, "instrumentId"); ok {
		out["instrumentId"] = v
	} else if v, ok := getStr(payload, "instrument_token"); ok {
		out["instrumentId"] = v
	}

	if v, ok := getStr(payload, "transactiontype"); ok {
		out["transactionType"] = v
	} else if v, ok := getStr(payload, "transactionType"); ok {
		out["transactionType"] = v
	}

	if v, ok := getInt(payload, "quantity"); ok {
		out["quantity"] = v
	}

	if v, ok := getStr(payload, "producttype"); ok {
		out["product"] = mapProduct(v)
	} else if v, ok := getStr(payload, "product"); ok {
		out["product"] = mapProduct(v)
	}

	if v, ok := getStr(payload, "variety"); ok {
		out["orderComplexity"] = mapVariety(v)
	} else if v, ok := getStr(payload, "orderComplexity"); ok {
		out["orderComplexity"] = v
	}

	if v, ok := getStr(payload, "ordertype"); ok {
		out["orderType"] = mapOrderType(v)
	} else if v, ok := getStr(payload, "orderType"); ok {
		out["orderType"] = v
	}

	if v, ok := getStr(payload, "duration"); ok {
		out["validity"] = v
	} else if v, ok := getStr(payload, "validity"); ok {
		out["validity"] = v
	}

	if v, ok := getStr(payload, "price"); ok {
		out["price"] = v
	}

	if v, ok := getStr(payload, "triggerprice"); ok {
		out["slTriggerPrice"] = v
	} else if v, ok := getStr(payload, "slTriggerPrice"); ok {
		out["slTriggerPrice"] = v
	}

	if v, ok := getStr(payload, "squareoff"); ok {
		out["slLegPrice"] = v
	}

	if v, ok := getStr(payload, "stoploss"); ok {
		out["targetLegPrice"] = v
	}

	out["disclosedQuantity"] = 0
	out["marketProtectionPercent"] = ""
	out["deviceId"] = fmt.Sprintf("%x", time.Now().UnixNano())
	out["trailingSlAmount"] = ""
	out["apiOrderSource"] = ""
	out["algoId"] = ""
	out["orderTag"] = ""

	return out
}

func mapVariety(v string) string {
	switch v {
	case "NORMAL":
		return "REGULAR"
	case "AMO", "BO", "CO":
		return v
	default:
		return v
	}
}

func mapProduct(v string) string {
	switch v {
	case "DELIVERY":
		return "LONGTERM"
	case "CARRYFORWARD":
		return "LONGTERM"
	case "MARGIN":
		return "MTF"
	case "MIS":
		return "INTRADAY"
	default:
		return v
	}
}

func mapOrderType(v string) string {
	if v == "SL-M" || v == "SLM" {
		return "SLM"
	}
	return v
}

func getInt(m map[string]any, key string) (int, bool) {
	v, ok := m[key]
	if !ok {
		return 0, false
	}
	switch x := v.(type) {
	case int:
		return x, true
	case float64:
		return int(x), true
	case string:
		i, err := strconv.Atoi(x)
		if err == nil {
			return i, true
		}
		return 0, false
	default:
		return 0, false
	}
}

func getStr(m map[string]any, key string) (string, bool) {
	v, ok := m[key]
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	if ok {
		return s, true
	}
	return fmt.Sprintf("%v", v), true
}

func (c *Client) PlaceOrder(sessionToken string, payload map[string]any) (map[string]any, error) {
	normalized := normalizeOrderPayload(payload)
	body, _ := json.Marshal([]any{normalized})
	data, err := c.doAPIRequest("POST", "/open-api/od/v1/orders/placeorder", sessionToken, string(body))
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("empty response from broker API")
	}
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("broker API error: %s", string(data))
	}
	return result, nil
}

func (c *Client) ModifyOrder(sessionToken string, payload map[string]any) (map[string]any, error) {
	body, _ := json.Marshal(payload)
	data, err := c.doAPIRequest("POST", "/open-api/od/v1/orders/modify", sessionToken, string(body))
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("empty response from broker API")
	}
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("broker API error: %s", string(data))
	}
	return result, nil
}

func (c *Client) CancelOrder(sessionToken, orderID string) (map[string]any, error) {
	payload := map[string]string{"brokerOrderId": orderID}
	body, _ := json.Marshal(payload)
	data, err := c.doAPIRequest("POST", "/open-api/od/v1/orders/cancel", sessionToken, string(body))
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("empty response from broker API")
	}
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("broker API error: %s", string(data))
	}
	return result, nil
}

func (c *Client) GetOrderBook(sessionToken string) ([]map[string]any, error) {
	data, err := c.doAPIRequest("GET", "/open-api/od/v1/orders/book", sessionToken, "")
	if err != nil {
		return nil, err
	}
	return parseResultList(data)
}

func (c *Client) GetTradeBook(sessionToken string) ([]map[string]any, error) {
	data, err := c.doAPIRequest("GET", "/open-api/od/v1/orders/trades", sessionToken, "")
	if err != nil {
		return nil, err
	}
	return parseResultList(data)
}

func (c *Client) GetPositions(sessionToken string) ([]map[string]any, error) {
	data, err := c.doAPIRequest("GET", "/open-api/od/v1/positions", sessionToken, "")
	if err != nil {
		return nil, err
	}
	return parseResultList(data)
}

func (c *Client) GetHoldings(sessionToken, productType string) (map[string]any, error) {
	return c.DoGet("/open-api/od/v1/holdings/"+productType, sessionToken)
}

func parseResultList(data []byte) ([]map[string]any, error) {
	var resp struct {
		Status  string           `json:"status"`
		Message string           `json:"message"`
		Result  []map[string]any `json:"result"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse list response: %w", err)
	}
	if resp.Result == nil {
		return []map[string]any{}, nil
	}
	return resp.Result, nil
}
