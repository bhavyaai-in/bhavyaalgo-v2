package aliceblue

import (
	"encoding/json"
	"fmt"
)

type profileResponse struct {
	Stat    string          `json:"status"`
	Message string          `json:"message"`
	Result  []profileResult `json:"result"`
}

type profileResult struct {
	ClientID   string   `json:"clientId"`
	ClientName string   `json:"clientName"`
	Exchanges  []string `json:"exchanges"`
	Products   []string `json:"products"`
}

func (c *Client) GetProfile(sessionToken string) (string, error) {
	data, err := c.doAPIRequest("GET", "/open-api/od/v1/profile/", sessionToken, "")
	if err != nil {
		return "", err
	}
	var resp profileResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return "", fmt.Errorf("failed to parse profile response: %w", err)
	}
	if len(resp.Result) == 0 {
		return "", fmt.Errorf("profile not found")
	}
	name := resp.Result[0].ClientName
	if name == "" {
		name = resp.Result[0].ClientID
	}
	return name, nil
}

func (c *Client) GetProfileRaw(sessionToken string) (map[string]any, error) {
	return c.DoGet("/open-api/od/v1/profile/", sessionToken)
}

func FetchProfile(sessionToken, appCode, apiSecret string) (string, error) {
	return NewClient(appCode, apiSecret).GetProfile(sessionToken)
}

func FetchProfileRaw(sessionToken, appCode, apiSecret string) (map[string]any, error) {
	return NewClient(appCode, apiSecret).GetProfileRaw(sessionToken)
}
