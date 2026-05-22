package aliceblue

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base32"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"
	"unicode"
)

type authResponse struct {
	Stat   string        `json:"stat"`
	Result []authSession `json:"result"`
}

type authSession struct {
	ClientID    string `json:"clientId"`
	UserSession string `json:"userSession"`
}

func generateTOTP(secret string) string {
	secret = strings.ToUpper(strings.TrimSpace(secret))
	var cleaned strings.Builder
	for _, r := range secret {
		if r == ' ' {
			continue
		}
		cleaned.WriteRune(unicode.ToUpper(r))
	}
	secret = cleaned.String()

	padLen := len(secret) % 8
	if padLen != 0 {
		secret += strings.Repeat("=", 8-padLen)
	}

	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(secret)
	if err != nil {
		key, err = base32.StdEncoding.DecodeString(secret)
		if err != nil {
			return ""
		}
	}

	steps := time.Now().Unix() / 30
	msg := make([]byte, 8)
	for i := 7; i >= 0; i-- {
		msg[i] = byte(steps & 0xFF)
		steps >>= 8
	}

	mac := hmac.New(sha1.New, key)
	mac.Write(msg)
	hash := mac.Sum(nil)

	offset := hash[len(hash)-1] & 0x0F
	code := (int(hash[offset]&0x7F) << 24) |
		(int(hash[offset+1]&0xFF) << 16) |
		(int(hash[offset+2]&0xFF) << 8) |
		int(hash[offset+3]&0xFF)

	return fmt.Sprintf("%06d", code%1000000)
}

func computeChecksum(userID, authCode, apiSecret string) string {
	h := sha256.New()
	h.Write([]byte(userID + authCode + apiSecret))
	return hex.EncodeToString(h.Sum(nil))
}

func (c *Client) Authenticate(userID, password, totpSecret string) (string, string, string, error) {
	totp := generateTOTP(totpSecret)
	if totp == "" {
		return "", "", "", fmt.Errorf("failed to generate TOTP from secret")
	}

	step1Body, _ := json.Marshal(map[string]string{"vendor": c.appCode})
	_, err := c.doAuthRequest("POST", "/omk/auth/sso/vendor/deatils", string(step1Body))
	if err != nil {
		return "", "", "", fmt.Errorf("vendor details failed: %w", err)
	}

	step2Body, _ := json.Marshal(map[string]string{"userId": userID})
	_, err = c.doAuthRequest("POST", "/omk/auth/access/verify/user", string(step2Body))
	if err != nil {
		return "", "", "", fmt.Errorf("user verification failed: %w", err)
	}

	step3Body, _ := json.Marshal(map[string]string{"userId": userID})
	encData, err := c.doAuthRequest("POST", "/omk/auth/access/client/enckey", string(step3Body))
	if err != nil {
		return "", "", "", fmt.Errorf("encryption key fetch failed: %w", err)
	}

	var encResp struct {
		Result []struct {
			EncKey string `json:"encKey"`
		} `json:"result"`
	}
	if err := json.Unmarshal(encData, &encResp); err != nil {
		return "", "", "", fmt.Errorf("failed to parse encKey response: %w", err)
	}
	if len(encResp.Result) == 0 || encResp.Result[0].EncKey == "" {
		return "", "", "", fmt.Errorf("no encryption key in response")
	}
	encKey := encResp.Result[0].EncKey

	encryptedPwd, err := opensslEncrypt(password, encKey)
	if err != nil {
		return "", "", "", fmt.Errorf("password encryption failed: %w", err)
	}

	step4Body, _ := json.Marshal(map[string]string{
		"userId":   userID,
		"userData": encryptedPwd,
		"source":   "WEB",
	})
	pwdData, err := c.doAuthRequest("POST", "/omk/auth/access/v2/pwd/validate", string(step4Body))
	if err != nil {
		return "", "", "", fmt.Errorf("password validation failed: %w", err)
	}

	var pwdResp struct {
		Result []struct {
			Token  string `json:"token"`
			KcRole string `json:"kcRole"`
		} `json:"result"`
	}
	if err := json.Unmarshal(pwdData, &pwdResp); err != nil {
		return "", "", "", fmt.Errorf("failed to parse password validation response: %w", err)
	}
	if len(pwdResp.Result) == 0 || pwdResp.Result[0].Token == "" {
		var errResp struct {
			Emsg string `json:"emsg"`
		}
		json.Unmarshal(pwdData, &errResp)
		if errResp.Emsg != "" {
			return "", "", "", fmt.Errorf("password validation failed: %s", errResp.Emsg)
		}
		return "", "", "", fmt.Errorf("password validation failed: no token in response")
	}
	token := pwdResp.Result[0].Token

	deviceID := fmt.Sprintf("%x", time.Now().UnixNano())

	step5Body, _ := json.Marshal(map[string]string{
		"userId":       userID,
		"totp":         totp,
		"source":       "WEB",
		"deviceId":     deviceID,
		"deviceNumber": deviceID,
		"vendor":       c.appCode,
	})

	totpData, err := c.doAuthRequestWithToken("POST", "/omk/auth/access/topt/verify", token, string(step5Body))
	if err != nil {
		return "", "", "", fmt.Errorf("TOTP verify request failed: %w", err)
	}

	type totpResult struct {
		RedirectURL  string `json:"redirectUrl"`
		AccessToken  string `json:"accessToken"`
	}
	var totpResp struct {
		Result []totpResult `json:"result"`
	}
	if err := json.Unmarshal(totpData, &totpResp); err != nil {
		return "", "", "", fmt.Errorf("failed to parse TOTP response: %w", err)
	}
	if len(totpResp.Result) == 0 {
		return "", "", "", fmt.Errorf("TOTP verification failed: empty result")
	}

	tr := totpResp.Result[0]


	if tr.RedirectURL == "" {
		return "", "", "", fmt.Errorf("TOTP verification failed: no redirect URL or access token in response")
	}

	parsedURL, err := url.Parse(tr.RedirectURL)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to parse redirect URL: %w", err)
	}
	authCode := parsedURL.Query().Get("authCode")
	if authCode == "" {
		return "", "", "", fmt.Errorf("no authCode in redirect URL")
	}

	checksum := computeChecksum(userID, authCode, c.apiSecret)

	sessionBody, _ := json.Marshal(map[string]string{"checkSum": checksum})
	sessionData, err := c.doAPIRequest("POST", "/open-api/od/v1/vendor/getUserDetails", "", string(sessionBody))
	if err != nil {
		return "", "", "", fmt.Errorf("session fetch failed: %w", err)
	}

	var sessionResp struct {
		Stat        string `json:"stat"`
		Emsg        string `json:"emsg"`
		ClientID    string `json:"clientId"`
		UserSession string `json:"userSession"`
	}
	if err := json.Unmarshal(sessionData, &sessionResp); err != nil {
		return "", "", "", fmt.Errorf("failed to parse session response: %w", err)
	}
	if sessionResp.Stat != "Ok" || sessionResp.UserSession == "" {
		msg := sessionResp.Emsg
		if msg == "" {
			msg = "unknown error"
		}
		return "", "", "", fmt.Errorf("session fetch failed: %s", msg)
	}
	return sessionResp.UserSession, sessionResp.UserSession, sessionResp.ClientID, nil
}

func (c *Client) GetSession(userID, authCode string) (string, error) {
	checksum := computeChecksum(userID, authCode, c.apiSecret)
	sessionBody, _ := json.Marshal(map[string]string{"checkSum": checksum})
	data, err := c.doAPIRequest("POST", "/open-api/od/v1/vendor/getUserDetails", "", string(sessionBody))
	if err != nil {
		return "", err
	}

	var resp struct {
		Stat        string `json:"stat"`
		Emsg        string `json:"emsg"`
		UserSession string `json:"userSession"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return "", fmt.Errorf("failed to parse session response: %w", err)
	}
	if resp.Stat != "Ok" || resp.UserSession == "" {
		msg := resp.Emsg
		if msg == "" {
			msg = "unknown error"
		}
		return "", fmt.Errorf("session fetch failed: %s", msg)
	}
	return resp.UserSession, nil
}

func AuthenticateBroker(userID, password, totpSecret, appCode, apiSecret string) (string, string, string, error) {
	client := NewClient(appCode, apiSecret)
	return client.Authenticate(userID, password, totpSecret)
}

func (c *Client) getUserSessionWithToken(accessToken, userID string) (string, error) {
	data, err := c.doAPIRequest("POST", "/open-api/od/v1/vendor/getUserDetails", accessToken, "{}")
	if err != nil {
		return "", err
	}

	var resp struct {
		Stat        string `json:"stat"`
		Emsg        string `json:"emsg"`
		UserSession string `json:"userSession"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return "", err
	}
	if resp.Stat == "Ok" && resp.UserSession != "" {
		return resp.UserSession, nil
	}
	return "", fmt.Errorf("getUserDetails failed with Bearer: %s", resp.Emsg)
}
