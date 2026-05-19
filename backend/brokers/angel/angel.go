package angel

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	BaseURL     = "https://clipx.bhavyaai.com/rqfarward?url=https://apiconnect.angelone.in"
	LoginURL    = BaseURL + "/rest/auth/angelbroking/user/v1/loginByPassword"
	MarginURL   = BaseURL + "/rest/secure/angelbroking/user/v1/getRMS"
	OrderBookURL = BaseURL + "/rest/secure/angelbroking/order/v1/getOrderBook"
	TradeBookURL = BaseURL + "/rest/secure/angelbroking/order/v1/getTradeBook"
	PositionURL = BaseURL + "/rest/secure/angelbroking/order/v1/getPosition"
	HoldingURL  = BaseURL + "/rest/secure/angelbroking/portfolio/v1/getAllHolding"
	PlaceOrderURL = BaseURL + "/rest/secure/angelbroking/order/v1/placeOrder"
	ModifyOrderURL = BaseURL + "/rest/secure/angelbroking/order/v1/modifyOrder"
	CancelOrderURL = BaseURL + "/rest/secure/angelbroking/order/v1/cancelOrder"
	MarginCalcURL = BaseURL + "/rest/secure/angelbroking/margin/v1/batch"
	ProfileURL    = BaseURL + "/rest/secure/angelbroking/user/v1/getProfile"
)

type Client struct {
	apiKey     string
	httpClient *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:     apiKey,
		httpClient: &http.Client{},
	}
}

func (c *Client) headers(authToken string) http.Header {
	h := http.Header{}
	h.Set("Content-Type", "application/json")
	h.Set("Accept", "application/json")
	h.Set("X-UserType", "USER")
	h.Set("X-SourceID", "WEB")
	h.Set("X-ClientLocalIP", "CLIENT_LOCAL_IP")
	h.Set("X-ClientPublicIP", "CLIENT_PUBLIC_IP")
	h.Set("X-MACAddress", "MAC_ADDRESS")
	//just for testing to be removed in production as static ip will not make issue durint developement but in production it will be issue so we need to remove it
	h.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")


	h.Set("X-PrivateKey", c.apiKey)
	if authToken != "" {
		h.Set("Authorization", "Bearer "+authToken)
	}
	return h
}

func (c *Client) doRequest(method, url, authToken, body string) ([]byte, error) {
	req, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header = c.headers(authToken)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == 403 {
		return nil, fmt.Errorf("authentication failed")
	}
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		msg := string(data)
		if msg == "" {
			msg = "empty response"
		}
		if len(msg) > 200 {
			msg = msg[:200]
		}
		return nil, fmt.Errorf("API error: %s", msg)
	}
	return data, nil
}

func (c *Client) DoGet(url, authToken string) (map[string]any, error) {
	data, err := c.doRequest("GET", url, authToken, "")
	if err != nil {
		return nil, err
	}
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return result, nil
}
