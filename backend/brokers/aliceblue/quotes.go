package aliceblue

import "encoding/json"

func (c *Client) GetQuote(sessionToken, exchange, symbol, token string) (map[string]any, error) {
	return c.DoPost("/api/v2/market/quotes", sessionToken, `{"tokens":["`+exchange+`:`+token+`"]}`)
}

func (c *Client) GetMultiQuote(sessionToken string, tokens []string) (map[string]any, error) {
	body, _ := json.Marshal(map[string][]string{"tokens": tokens})
	return c.DoPost("/api/v2/market/quotes", sessionToken, string(body))
}
