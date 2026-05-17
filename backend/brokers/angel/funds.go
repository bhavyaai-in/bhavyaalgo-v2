package angel

import (
	"encoding/json"
	"fmt"
)

type marginResponse struct {
	Data *marginData `json:"data"`
}

type marginData struct {
	AvailableCash  string `json:"availablecash"`
	UtilisedPayout string `json:"utilisedpayout"`
	M2MRealized    string `json:"m2mrealized"`
	M2MUnrealized  string `json:"m2munrealized"`
	UtilisedDebits string `json:"utiliseddebits"`
}

func (c *Client) GetMargin(authToken string) (map[string]string, error) {
	data, err := c.doRequest("GET", MarginURL, authToken, "")
	if err != nil {
		return nil, err
	}
	var resp marginResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse margin response: %w", err)
	}
	if resp.Data == nil {
		return map[string]string{}, nil
	}
	available := parseFloat(resp.Data.AvailableCash)
	utilisedPayout := parseFloat(resp.Data.UtilisedPayout)
	collateral := available - utilisedPayout
	return map[string]string{
		"available_cash":   fmt.Sprintf("%.2f", available),
		"collateral":       fmt.Sprintf("%.2f", collateral),
		"m2m_realized":     fmt.Sprintf("%.2f", parseFloat(resp.Data.M2MRealized)),
		"m2m_unrealized":   fmt.Sprintf("%.2f", parseFloat(resp.Data.M2MUnrealized)),
		"utilised_debits":  fmt.Sprintf("%.2f", parseFloat(resp.Data.UtilisedDebits)),
	}, nil
}

func parseFloat(s string) float64 {
	var v float64
	fmt.Sscanf(s, "%f", &v)
	return v
}
