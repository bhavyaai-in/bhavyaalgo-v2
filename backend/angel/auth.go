package angel

import (
	"encoding/json"
	"fmt"
)

type loginResponse struct {
	Data *loginData `json:"data"`
}

type loginData struct {
	JWTToken    string `json:"jwtToken"`
	FeedToken   string `json:"feedToken"`
	ClientName  string `json:"clientname"`
	DisplayName string `json:"displayname"`
	ClientCode  string `json:"clientcode"`
}

func (c *Client) Authenticate(clientCode, brokerPin, totp string) (string, string, string, error) {
	payload := map[string]string{
		"clientcode": clientCode,
		"password":   brokerPin,
		"totp":       totp,
	}
	body, _ := json.Marshal(payload)
	data, err := c.doRequest("POST", LoginURL, "", string(body))
	if err != nil {
		return "", "", "", err
	}
	var resp struct {
		Data *loginData `json:"data"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return "", "", "", fmt.Errorf("failed to parse login response: %w", err)
	}
	if resp.Data == nil || resp.Data.JWTToken == "" {
		var errResp struct {
			Message string `json:"message"`
		}
		json.Unmarshal(data, &errResp)
		if errResp.Message != "" {
			return "", "", "", fmt.Errorf("login failed: %s", errResp.Message)
		}
		return "", "", "", fmt.Errorf("login failed: no token in response")
	}

	profileName := resp.Data.ClientName
	if profileName == "" {
		profileName = resp.Data.DisplayName
	}
	if profileName == "" {
		profileName = resp.Data.ClientCode
	}

	return resp.Data.JWTToken, resp.Data.FeedToken, profileName, nil
}
