package angel

import "encoding/json"

const QuoteURL = BaseURL + "/rest/secure/angelbroking/market/v1/quote"

type quoteReq struct {
	Mode           string              `json:"mode"`
	ExchangeTokens map[string][]string `json:"exchangeTokens"`
}

func (c *Client) GetQuote(authToken string, exchange, symbol, token string) (map[string]any, error) {
	req := quoteReq{
		Mode: "FULL",
		ExchangeTokens: map[string][]string{
			exchange: {token},
		},
	}
	body, _ := json.Marshal(req)
	data, err := c.doRequest("POST", QuoteURL, authToken, string(body))
	if err != nil {
		return nil, err
	}
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) GetMultiQuote(authToken string, tokens map[string][]string) (map[string]any, error) {
	req := quoteReq{
		Mode:           "FULL",
		ExchangeTokens: tokens,
	}
	body, _ := json.Marshal(req)
	data, err := c.doRequest("POST", QuoteURL, authToken, string(body))
	if err != nil {
		return nil, err
	}
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}
