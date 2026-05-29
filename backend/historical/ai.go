package historical

import (
	"bytes"
	"database/sql"
	"encoding/csv"
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
var nseCSVHTTP = &http.Client{Timeout: 15 * time.Second}

var knownIndices = []struct {
	key     string
	url     string
	display string
}{
	{"nifty 50", "https://archives.nseindia.com/content/indices/ind_nifty50list.csv", "Nifty 50"},
	{"nifty50", "https://archives.nseindia.com/content/indices/ind_nifty50list.csv", "Nifty 50"},
	{"nifty next 50", "https://archives.nseindia.com/content/indices/ind_niftynext50list.csv", "Nifty Next 50"},
	{"niftynext50", "https://archives.nseindia.com/content/indices/ind_niftynext50list.csv", "Nifty Next 50"},
	{"next 50", "https://archives.nseindia.com/content/indices/ind_niftynext50list.csv", "Nifty Next 50"},
	{"bank nifty", "https://archives.nseindia.com/content/indices/ind_niftybanklist.csv", "Bank Nifty"},
	{"banknifty", "https://archives.nseindia.com/content/indices/ind_niftybanklist.csv", "Bank Nifty"},
	{"nifty midcap", "https://archives.nseindia.com/content/indices/ind_niftymidcap100list.csv", "Nifty Midcap 100"},
	{"midcap", "https://archives.nseindia.com/content/indices/ind_niftymidcap100list.csv", "Nifty Midcap 100"},
	{"nifty smallcap", "https://archives.nseindia.com/content/indices/ind_niftysmallcap100list.csv", "Nifty Smallcap 100"},
	{"smallcap", "https://archives.nseindia.com/content/indices/ind_niftysmallcap100list.csv", "Nifty Smallcap 100"},
}

func matchIndex(q string) (string, string, bool) {
	lower := strings.ToLower(strings.TrimSpace(q))
	for _, idx := range knownIndices {
		if strings.Contains(lower, idx.key) {
			return idx.url, idx.display, true
		}
	}
	return "", "", false
}

func fetchNSEIndex(csvURL, indexName string) ([]AISymbol, error) {
	req, err := http.NewRequest("GET", csvURL, nil)
	if err != nil {
		return nil, fmt.Errorf("request error: %v", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")
	req.Header.Set("Accept", "text/csv,text/plain,*/*")

	resp, err := nseCSVHTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("NSE returned HTTP %d", resp.StatusCode)
	}

	reader := csv.NewReader(resp.Body)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("csv parse error: %v", err)
	}

	var symbols []AISymbol
	for i, row := range records {
		if i == 0 {
			continue
		}
		if len(row) < 3 {
			continue
		}
		sym := strings.TrimSpace(row[2])
		name := strings.TrimSpace(row[0])
		if sym == "" {
			continue
		}
		symbols = append(symbols, AISymbol{
			Symbol:   sym,
			Exchange: "NSE",
			Name:     name,
			Reason:   indexName + " constituent",
		})
	}
	return symbols, nil
}

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

func searchContracts(db *sql.DB, exchange string, keywords []string) ([]AISymbol, error) {
	if len(keywords) == 0 {
		return nil, nil
	}
	conditions := make([]string, 0, len(keywords))
	args := make([]any, 0, len(keywords)*2)
	for _, kw := range keywords {
		kw = strings.ToUpper(strings.TrimSpace(kw))
		if kw == "" {
			continue
		}
		conditions = append(conditions, "(symbol LIKE ? OR name LIKE ?)")
		args = append(args, "%"+kw+"%", "%"+kw+"%")
	}
	if len(conditions) == 0 {
		return nil, nil
	}
	typeFilter := `(instrumenttype='' OR instrumenttype='E' OR instrumenttype='D' OR instrumenttype='ETF' OR instrumenttype='INDEX' OR instrumenttype='AMXIDX')`
	symbolFilter := `symbol NOT GLOB '[0-9]*' AND symbol NOT GLOB '*[0-9][0-9][0-9]*'`
	query := fmt.Sprintf(
		`SELECT DISTINCT symbol, exchange, name FROM master_contracts
		 WHERE exchange=? AND %s AND %s AND (%s) ORDER BY symbol LIMIT 100`,
		typeFilter, symbolFilter, strings.Join(conditions, " AND "),
	)
	finalArgs := append([]any{exchange}, args...)
	rows, err := db.Query(query, finalArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	seen := map[string]bool{}
	var results []AISymbol
	for rows.Next() {
		var sym, ex, name string
		if rows.Scan(&sym, &ex, &name) == nil && !seen[strings.ToUpper(sym)] {
			seen[strings.ToUpper(sym)] = true
			results = append(results, AISymbol{
				Symbol:   sym,
				Exchange: ex,
				Name:     name,
				Reason:   "DB match: " + strings.Join(keywords, " "),
			})
		}
	}
	return results, nil
}

func extractKeywords(q string) []string {
	stopwords := map[string]bool{
		"stocks": true, "stock": true, "list": true, "companies": true,
		"company": true, "shares": true, "share": true, "index": true,
		"nifty": true, "sensex": true, "nse": true, "bse": true,
		"all": true, "top": true, "best": true, "in": true, "the": true,
		"a": true, "an": true, "of": true, "for": true, "and": true,
		"or": true, "to": true, "is": true, "that": true, "this": true,
		"with": true, "has": true, "have": true, "are": true, "was": true,
		"downloads": true, "download": true, "data": true, "every": true,
		"day": true, "watchlist": true, "should": true, "latest": true,
	}
	terms := strings.FieldsFunc(strings.ToLower(q), func(r rune) bool {
		return r == ' ' || r == ',' || r == '.' || r == '!' || r == '?' || r == '\'' || r == '"'
	})
	var keywords []string
	for _, t := range terms {
		t = strings.TrimSpace(t)
		if t == "" || stopwords[t] {
			continue
		}
		keywords = append(keywords, t)
	}
	return keywords
}

func verifyInDB(db *sql.DB, exchange string, symbols []AISymbol) (verified []AISymbol, skipped []string) {
	seen := map[string]bool{}
	for _, s := range symbols {
		key := strings.ToUpper(s.Symbol)
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		var dbSymbol, dbExchange string
		err := db.QueryRow(
			"SELECT symbol, exchange FROM master_contracts WHERE symbol=? AND (exchange=? OR exchange='NSE' OR exchange='BSE') LIMIT 1",
			key, exchange,
		).Scan(&dbSymbol, &dbExchange)
		if err != nil {
			skipped = append(skipped, key)
			continue
		}
		s.Symbol = dbSymbol
		s.Exchange = dbExchange
		verified = append(verified, s)
	}
	return
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

	// Step 1: Try NSE official index constituent data
	if csvURL, indexName, ok := matchIndex(query); ok {
		log.Printf("AI suggest: matched index %s, fetching from NSE...", indexName)
		nseSymbols, err := fetchNSEIndex(csvURL, indexName)
		if err != nil {
			log.Printf("AI suggest: NSE fetch failed: %v, falling back to AI", err)
		} else {
			verified, skipped := verifyInDB(db, exchange, nseSymbols)
			if len(skipped) > 0 {
				log.Printf("AI suggest: NSE %s: skipped %d not in DB: %s", indexName, len(skipped), strings.Join(skipped, ", "))
			}
			log.Printf("AI suggest: NSE %s → %d verified symbols", indexName, len(verified))
			return &AIResponse{Symbols: verified}
		}
	}

	// Step 2: Fall back to AI + DB search
	keywords := extractKeywords(query)
	log.Printf("AI suggest: extracted keywords: %v", keywords)

	dbResults, dbErr := searchContracts(db, exchange, keywords)
	if dbErr != nil {
		log.Printf("AI suggest: DB search error: %v", dbErr)
	}

	systemPrompt := `You are a stock market assistant for Indian markets.
Given a user query, generate a JSON list of matching stock/index/ETF symbols.

Rules:
- Use NSE/BSE listing format (e.g. RELIANCE, TCS, SBIN)
- For index queries, return ALL known constituent stocks
- Include symbol, exchange, name, and reason for each
- Return ONLY valid JSON, no markdown

{"symbols": [{"symbol": "RELIANCE", "exchange": "NSE", "name": "Reliance Industries Ltd", "reason": "index constituent"}]}`

	userMsg := query
	if len(dbResults) > 0 {
		var sample []string
		for i, s := range dbResults {
			if i >= 20 {
				break
			}
			sample = append(sample, fmt.Sprintf("%s (%s) - %s", s.Symbol, s.Exchange, s.Name))
		}
		userMsg = fmt.Sprintf("%s\n\nDatabase search found these matching contracts (use them to supplement your answer):\n%s\n(return all relevant symbols, including ones not in this list)", query, strings.Join(sample, "\n"))
	}

	payload := map[string]any{
		"model": model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userMsg},
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

	aiSymbols := parseSymbols(content)

	// Merge AI + DB results, verify all against DB
	var allSources []AISymbol
	allSources = append(allSources, aiSymbols...)
	allSources = append(allSources, dbResults...)
	verified, skipped := verifyInDB(db, exchange, allSources)

	if len(skipped) > 0 {
		log.Printf("AI suggest: skipped %d symbols: %s", len(skipped), strings.Join(skipped, ", "))
	}
	log.Printf("AI suggest: %d AI + %d DB → %d verified", len(aiSymbols), len(dbResults), len(verified))
	return &AIResponse{Symbols: verified}
}
