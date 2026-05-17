package angel

import (
	"encoding/json"
	"fmt"
)

type profileResponse struct {
	Data *profileData `json:"data"`
}

type profileData struct {
	Name        string `json:"name"`
	ClientName  string `json:"clientname"`
	DisplayName string `json:"displayname"`
	ClientCode  string `json:"clientcode"`
}

func (c *Client) GetProfile(authToken string) (string, error) {
	data, err := c.doRequest("GET", ProfileURL, authToken, "")
	if err != nil {
		return "", err
	}
	var resp profileResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return "", fmt.Errorf("failed to parse profile response: %w", err)
	}
	name, _, _ := extractProfileName(resp.Data)
	return name, nil
}

func (c *Client) GetProfileRaw(authToken string) (map[string]any, error) {
	data, err := c.doRequest("GET", ProfileURL, authToken, "")
	if err != nil {
		return nil, err
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse profile response: %w", err)
	}

	// Flatten the "data" field into the top level for cleaner display
	if dataObj, ok := raw["data"].(map[string]any); ok {
		// Insert name from profileData struct fields
		_ = dataObj
	}

	// Also try to extract structured profile for the name
	var resp profileResponse
	if err := json.Unmarshal(data, &resp); err == nil && resp.Data != nil {
		name, _, _ := extractProfileName(resp.Data)
		if name != "" {
			if raw["data"] != nil {
				if d, ok := raw["data"].(map[string]any); ok {
					d["_display_name"] = name
				}
			}
		}
	}

	return raw, nil
}

func FetchProfile(authToken, apiKey string) (string, error) {
	return NewClient(apiKey).GetProfile(authToken)
}

func FetchProfileRaw(authToken, apiKey string) (map[string]any, error) {
	return NewClient(apiKey).GetProfileRaw(authToken)
}

func extractProfileName(d *profileData) (string, string, string) {
	if d == nil {
		return "", "", ""
	}
	if d.Name != "" {
		return d.Name, "name", d.Name
	}
	if d.ClientName != "" {
		return d.ClientName, "clientname", d.ClientName
	}
	if d.DisplayName != "" {
		return d.DisplayName, "displayname", d.DisplayName
	}
	if d.ClientCode != "" {
		return d.ClientCode, "clientcode", d.ClientCode
	}
	return "", "", ""
}
