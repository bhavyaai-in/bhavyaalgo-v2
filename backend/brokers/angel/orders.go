package angel

import (
	"encoding/json"
	"fmt"
)

func (c *Client) GetOrderBook(authToken string) ([]map[string]any, error) {
	data, err := c.doRequest("GET", OrderBookURL, authToken, "")
	if err != nil {
		return nil, err
	}
	return parseListResponse(data)
}

func (c *Client) GetTradeBook(authToken string) ([]map[string]any, error) {
	data, err := c.doRequest("GET", TradeBookURL, authToken, "")
	if err != nil {
		return nil, err
	}
	return parseListResponse(data)
}

func (c *Client) GetPositions(authToken string) ([]map[string]any, error) {
	data, err := c.doRequest("GET", PositionURL, authToken, "")
	if err != nil {
		return nil, err
	}
	return parseListResponse(data)
}

func (c *Client) GetHoldings(authToken string) ([]map[string]any, error) {
	data, err := c.doRequest("GET", HoldingURL, authToken, "")
	if err != nil {
		return nil, err
	}
	return parseListResponse(data)
}

func (c *Client) PlaceOrder(authToken string, payload map[string]any) (map[string]any, error) {
	body, _ := json.Marshal(payload)
	data, err := c.doRequest("POST", PlaceOrderURL, authToken, string(body))
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("empty response from broker API")
	}
	// Try to extract error message even from non-JSON responses
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("broker API error: %s", string(data))
	}
	return result, nil
}

func (c *Client) ModifyOrder(authToken string, payload map[string]any) (map[string]any, error) {
	body, _ := json.Marshal(payload)
	data, err := c.doRequest("POST", ModifyOrderURL, authToken, string(body))
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

func (c *Client) CancelOrder(authToken, orderID string) (map[string]any, error) {
	payload := map[string]string{"orderid": orderID}

	// Try variety-based cancel first
	varietyPayload := map[string]string{"variety": "NORMAL", "orderid": orderID}
	body, _ := json.Marshal(varietyPayload)
	data, err := c.doRequest("POST", CancelOrderURL, authToken, string(body))
	if err != nil {
		body, _ = json.Marshal(payload)
		data, err = c.doRequest("POST", CancelOrderURL, authToken, string(body))
		if err != nil {
			return nil, err
		}
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

func parseListResponse(data []byte) ([]map[string]any, error) {
	var resp struct {
		Data []map[string]any `json:"data"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse list response: %w", err)
	}
	if resp.Data == nil {
		return []map[string]any{}, nil
	}
	return resp.Data, nil
}
