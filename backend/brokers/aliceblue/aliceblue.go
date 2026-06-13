package aliceblue

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"strings"
)

const (
	AuthBase = "https://clipx.bhavyaai.com/rqfarward?url=https://antdrn.aliceblueonline.com"
	BaseURL  = "https://clipx.bhavyaai.com/rqfarward?url=https://a3.aliceblueonline.com"
)

type Client struct {
	appCode    string
	apiSecret  string
	httpClient *http.Client
}

func NewClient(appCode, apiSecret string) *Client {
	jar, _ := cookiejar.New(nil)
	return &Client{
		appCode:    appCode,
		apiSecret:  apiSecret,
		httpClient: &http.Client{Jar: jar},
	}
}

func (c *Client) headers(sessionToken string) http.Header {
	h := http.Header{}
	h.Set("Content-Type", "application/json")
	h.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	if sessionToken != "" {
		h.Set("Authorization", "Bearer "+sessionToken)
	}
	return h
}

func (c *Client) authHeaders(sessionToken string) http.Header {
	h := c.headers(sessionToken)
	h.Set("Origin", "https://ant.aliceblueonline.com")
	h.Set("Referer", "https://ant.aliceblueonline.com/")
	return h
}

func (c *Client) doRequest(method, base, url, sessionToken, body string) ([]byte, error) {
	fullURL := base + url
	req, err := http.NewRequest(method, fullURL, strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header = c.headers(sessionToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		return nil, fmt.Errorf("authentication failed")
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		msg := string(data)
		if msg == "" {
			msg = "empty response"
		}
		if len(msg) > 200 {
			msg = msg[:200]
		}
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, msg)
	}
	return data, nil
}

func (c *Client) doAuthRequest(method, url, body string) ([]byte, error) {
	fullURL := AuthBase + url
	req, err := http.NewRequest(method, fullURL, strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header = c.authHeaders("")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		return nil, fmt.Errorf("authentication failed")
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		msg := string(data)
		if msg == "" {
			msg = "empty response"
		}
		if len(msg) > 200 {
			msg = msg[:200]
		}
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, msg)
	}
	return data, nil
}

func (c *Client) doAuthRequestWithToken(method, url, sessionToken, body string) ([]byte, error) {
	fullURL := AuthBase + url
	req, err := http.NewRequest(method, fullURL, strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header = c.authHeaders(sessionToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		return nil, fmt.Errorf("authentication failed")
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		msg := string(data)
		if msg == "" {
			msg = "empty response"
		}
		if len(msg) > 200 {
			msg = msg[:200]
		}
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, msg)
	}
	return data, nil
}

func (c *Client) doAPIRequest(method, url, sessionToken, body string) ([]byte, error) {
	return c.doRequest(method, BaseURL, url, sessionToken, body)
}

func (c *Client) DoGet(url, sessionToken string) (map[string]any, error) {
	data, err := c.doAPIRequest("GET", url, sessionToken, "")
	if err != nil {
		return nil, err
	}
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return result, nil
}

func (c *Client) DoPost(url, sessionToken, body string) (map[string]any, error) {
	data, err := c.doAPIRequest("POST", url, sessionToken, body)
	if err != nil {
		return nil, err
	}
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return result, nil
}

func CreateWsSession(sessionToken, clientID, appCode, apiSecret string) error {
	client := NewClient(appCode, apiSecret)
	body, _ := json.Marshal(map[string]string{"client_id": clientID, "source": "API"})
	_, err := client.DoPost("/open-api/od/v1/profile/createWsSess", sessionToken, string(body))
	return err
}

func opensslEncrypt(plaintext, password string) (string, error) {
	salt := make([]byte, 8)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	key, iv := evpBytesToKey([]byte(password), salt, 32, 16)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	plainBytes := pkcs7Pad([]byte(plaintext), aes.BlockSize)
	ciphertext := make([]byte, len(plainBytes))

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, plainBytes)

	result := make([]byte, 0, 16+len(ciphertext))
	result = append(result, []byte("Salted__")...)
	result = append(result, salt...)
	result = append(result, ciphertext...)

	return base64.StdEncoding.EncodeToString(result), nil
}

func evpBytesToKey(password, salt []byte, keyLen, ivLen int) ([]byte, []byte) {
	var digest []byte
	var last []byte

	for len(digest) < keyLen+ivLen {
		h := md5.New()
		h.Write(last)
		h.Write(password)
		h.Write(salt)
		last = h.Sum(nil)
		digest = append(digest, last...)
	}

	return digest[:keyLen], digest[keyLen : keyLen+ivLen]
}

func pkcs7Pad(data []byte, blockSize int) []byte {
	padLen := blockSize - len(data)%blockSize
	padding := make([]byte, padLen)
	for i := range padding {
		padding[i] = byte(padLen)
	}
	return append(data, padding...)
}
