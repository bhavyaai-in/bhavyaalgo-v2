package historical

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type AIRequest struct {
	Query    string `json:"query"`
	Exchange string `json:"exchange"`
}

type AISymbol struct {
	Symbol   string `json:"symbol"`
	Exchange string `json:"exchange"`
	Name     string `json:"name"`
	Reason   string `json:"reason"`
}

type AIResponse struct {
	Symbols []AISymbol `json:"symbols"`
	Error   string     `json:"error,omitempty"`
}

var aiHTTP = &http.Client{Timeout: 60 * time.Second}

func parseSymbols(content string) []AISymbol {
	var aiResp AIResponse
	if err := json.Unmarshal([]byte(content), &aiResp); err == nil && len(aiResp.Symbols) > 0 {
		return aiResp.Symbols
	}
	var arr []AISymbol
	if err := json.Unmarshal([]byte(content), &arr); err == nil {
		return arr
	}
	return nil
}

func GenerateSymbols(db *sql.DB, query, exchange string) *AIResponse {
	apiURL := os.Getenv("AI_API_URL")
	apiKey := os.Getenv("AI_API_KEY")
	model := os.Getenv("AI_MODEL")

	if apiURL == "" || apiKey == "" {
		return &AIResponse{Error: "AI_API_URL and AI_API_KEY must be set in .env"}
	}
	if model == "" {
		model = "gpt-4o-mini"
	}

	systemPrompt := fmt.Sprintf(`You are a stock market assistant for Indian markets (NSE, BSE, NFO).
Given a user query, generate a JSON list of matching stock/index/ETF symbols.

Rules:
- Return symbols as per NSE/BSE listing (e.g. RELIANCE, TCS, SBIN, NIFTY, BANKNIFTY)
- All symbols must exist on %s exchange
- If the user asks for an index like "Nifty 50", "Nifty Next 50", "Bank Nifty", etc. — return ALL constituent stocks (50 for Nifty 50, 50 for Nifty Next 50, 12 for Bank Nifty, etc.)
- Always return the complete list — do not skip any constituent
- Include symbol, exchange, a short name/description, and reason for inclusion
- Return ONLY valid JSON, no markdown

{"symbols": [{"symbol": "RELIANCE", "exchange": "NSE", "name": "Reliance Industries Ltd", "reason": "Nifty 50 constituent"}]}`, exchange)

	payload := map[string]any{
		"model": model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": query},
		},
		"response_format": map[string]string{"type": "json_object"},
		"temperature":     0,
		"max_tokens":     4000,
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", apiURL, bytes.NewReader(body))
	if err != nil {
		return &AIResponse{Error: fmt.Sprintf("request error: %v", err)}
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := aiHTTP.Do(req)
	if err != nil {
		return &AIResponse{Error: fmt.Sprintf("api error: %v", err)}
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		b := strings.TrimSpace(string(respBody))
		if len(b) > 200 {
			b = b[:200]
		}
		if b == "" {
			b = "(empty body)"
		}
		return &AIResponse{Error: fmt.Sprintf("AI API returned HTTP %d: %s", resp.StatusCode, b)}
	}

	var openAIResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &openAIResp); err != nil {
		b := strings.TrimSpace(string(respBody))
		if len(b) > 200 {
			b = b[:200]
		}
		return &AIResponse{Error: fmt.Sprintf("parse error: %v - body: %s", err, b)}
	}
	if len(openAIResp.Choices) == 0 {
		return &AIResponse{Error: "no response from AI"}
	}

	content := openAIResp.Choices[0].Message.Content
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	symbols := parseSymbols(content)
	if symbols == nil {
		return &AIResponse{Error: fmt.Sprintf("AI response parse error: unrecognized JSON format - response: %s", content[:min(len(content), 500)])}
	}

	var verified []AISymbol
	var skipped []string
	seen := map[string]bool{}
	for _, s := range symbols {
		if s.Symbol == "" || seen[strings.ToUpper(s.Symbol)] {
			continue
		}
		seen[strings.ToUpper(s.Symbol)] = true
		var dbSymbol, dbExchange string
		err := db.QueryRow(
			"SELECT symbol, exchange FROM master_contracts WHERE symbol=? AND (exchange=? OR exchange='NSE' OR exchange='BSE') LIMIT 1",
			strings.ToUpper(s.Symbol), exchange,
		).Scan(&dbSymbol, &dbExchange)
		if err != nil {
			skipped = append(skipped, s.Symbol)
			continue
		}
		s.Symbol = dbSymbol
		s.Exchange = dbExchange
		verified = append(verified, s)
	}

	if len(skipped) > 0 {
		log.Printf("AI suggest: skipped %d symbols not in DB: %s", len(skipped), strings.Join(skipped, ", "))
	}

	return &AIResponse{Symbols: verified}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
