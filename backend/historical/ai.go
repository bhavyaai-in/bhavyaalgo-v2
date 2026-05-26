package historical

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// AIRequest generates symbol suggestions from natural language queries.
type AIRequest struct {
	Query    string   `json:"query"`
	Exchange string   `json:"exchange"`
}

type AISymbol struct {
	Symbol   string `json:"symbol"`
	Exchange string `json:"exchange"`
	Name     string `json:"name"`
	Reason   string `json:"reason"`
}

type AIResponse struct {
	Symbols []AISymbol `json:"symbols"`
	Error   string      `json:"error,omitempty"`
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
Given a natural language query, generate a list of relevant stock/ETF/index symbols.

Rules:
- Only return symbols that exist on %s exchange
- Return the SYMBOL as per NSE/BSE listing (e.g., RELIANCE, TCS, SBIN, NIFTY, BANKNIFTY)
- If the user asks about indices like "Nifty 50", return the 50 constituent stocks
- If the user asks about specific criteria (turnover, market cap, sector), use your knowledge
- Return max 50 symbols per query
- For each symbol provide: symbol, exchange, a short name/description, and reason for inclusion

Respond in JSON format ONLY:
{"symbols": [{"symbol": "RELIANCE", "exchange": "NSE", "name": "Reliance Industries Ltd", "reason": "Nifty 50 constituent"}]}`, exchange)

	payload := map[string]any{
		"model": model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": query},
		},
		"temperature": 0.3,
		"max_tokens":  2000,
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", apiURL, bytes.NewReader(body))
	if err != nil {
		return &AIResponse{Error: fmt.Sprintf("request error: %v", err)}
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return &AIResponse{Error: fmt.Sprintf("api error: %v", err)}
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	// Parse OpenAI response
	var openAIResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &openAIResp); err != nil {
		return &AIResponse{Error: fmt.Sprintf("parse error: %v", err)}
	}
	if len(openAIResp.Choices) == 0 {
		return &AIResponse{Error: "no response from AI"}
	}

	content := openAIResp.Choices[0].Message.Content
	content = strings.TrimSpace(content)
	// Remove code block markers if present
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var aiResp AIResponse
	if err := json.Unmarshal([]byte(content), &aiResp); err != nil {
		return &AIResponse{Error: fmt.Sprintf("AI response parse error: %v - response: %s", err, content[:min(len(content), 500)])}
	}

	// Verify symbols exist in master_contracts and fill in exchange from DB
	var verified []AISymbol
	for _, s := range aiResp.Symbols {
		if s.Symbol == "" {
			continue
		}
		// Verify in DB
		var dbSymbol, dbExchange string
		err := db.QueryRow(
			"SELECT symbol, exchange FROM master_contracts WHERE symbol=? AND (exchange=? OR exchange='NSE' OR exchange='BSE') LIMIT 1",
			strings.ToUpper(s.Symbol), exchange,
		).Scan(&dbSymbol, &dbExchange)
		if err != nil {
			continue
		}
		s.Symbol = dbSymbol
		s.Exchange = dbExchange
		verified = append(verified, s)
		if len(verified) >= 50 {
			break
		}
	}

	return &AIResponse{Symbols: verified}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
