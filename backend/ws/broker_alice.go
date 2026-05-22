package ws

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
	"fmt"

	"github.com/gorilla/websocket"
)

const aliceWSURL = "wss://ws1.aliceblueonline.com/NorenWS/"

type AliceBrokerClient struct {
	sessionToken        string
	clientID            string
	hub                 *Hub
	tokenSymbol         map[string]string
	conn                *websocket.Conn
	mu                  sync.Mutex
	writeMu             sync.Mutex
	symbols             []string
	initialSymbols      []string
	running             bool
	stopCh              chan struct{}
	reconnectAttempt    int
	reconnectMaxRetries int
	reconnectMaxDelay   time.Duration
	connectTimeout      time.Duration
	closeCache          map[string]float64
}

func NewAliceBrokerClient(sessionToken, clientID string, hub *Hub, tokenSymbol map[string]string) *AliceBrokerClient {
	return &AliceBrokerClient{
		sessionToken:        sessionToken,
		clientID:            clientID,
		hub:                 hub,
		tokenSymbol:         tokenSymbol,
		closeCache:          make(map[string]float64),
		stopCh:              make(chan struct{}),
		reconnectMaxRetries: 300,
		reconnectMaxDelay:   60 * time.Second,
		connectTimeout:      7 * time.Second,
	}
}

func (b *AliceBrokerClient) SetInitialSymbols(symbols []string) {
	b.mu.Lock()
	b.initialSymbols = symbols
	b.mu.Unlock()
}

func (b *AliceBrokerClient) Run() {
	b.running = true
	for b.running {
		if b.reconnectAttempt > b.reconnectMaxRetries {
			log.Printf("alice broker: max reconnect attempts (%d) reached, stopping", b.reconnectMaxRetries)
			return
		}
		if b.reconnectAttempt > 0 {
			nextDelay := time.Duration(math.Pow(2, float64(b.reconnectAttempt))) * time.Second
			if nextDelay > b.reconnectMaxDelay {
				nextDelay = b.reconnectMaxDelay
			}
			log.Printf("alice broker: reconnecting in %v (attempt %d/%d)", nextDelay, b.reconnectAttempt, b.reconnectMaxRetries)
			b.mu.Lock()
			if b.conn != nil {
				b.conn.Close()
				b.conn = nil
			}
			b.mu.Unlock()
			select {
			case <-b.stopCh:
				return
			case <-time.After(nextDelay):
			}
		}
		b.reconnectAttempt++
		b.connect()
	}
}

func (b *AliceBrokerClient) connect() {
	d := websocket.DefaultDialer
	d.HandshakeTimeout = b.connectTimeout

	h := http.Header{}
	h.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	h.Set("Origin", "https://ws1.aliceblueonline.com")

	conn, _, err := d.Dial(aliceWSURL, h)
	if err != nil {
		log.Printf("alice broker: dial failed: %v", err)
		return
	}

	b.mu.Lock()
	b.conn = conn
	b.mu.Unlock()
	b.reconnectAttempt = 0
	log.Printf("alice broker: connected")

	suberToken := computeAliceSuberToken(b.sessionToken)
	actid := b.clientID + "_API"
	initMsg := map[string]string{
		"susertoken": suberToken,
		"t":          "c",
		"actid":      actid,
		"uid":        actid,
		"source":     "API",
	}
	body, _ := json.Marshal(initMsg)
	b.writeMu.Lock()
	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	err = conn.WriteMessage(websocket.TextMessage, body)
	b.writeMu.Unlock()
	if err != nil {
		log.Printf("alice broker: init write error: %v", err)
		b.mu.Lock()
		conn.Close()
		b.conn = nil
		b.mu.Unlock()
		return
	}

	authOk := false
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		conn.SetReadDeadline(deadline)
		_, resp, err := conn.ReadMessage()
		if err != nil {
			break
		}
		var initResp struct {
			T string `json:"t"`
			K string `json:"k"`
			S string `json:"s"`
		}
		if err := json.Unmarshal(resp, &initResp); err != nil {
			continue
		}
		if (initResp.T == "cf" && initResp.K == "OK") || (initResp.T == "ck" && initResp.S == "OK") {
			authOk = true
			break
		}
	}
	if !authOk {
		log.Printf("alice broker: init failed (no auth response)")
		b.mu.Lock()
		conn.Close()
		b.conn = nil
		b.mu.Unlock()
		return
	}
	conn.SetReadDeadline(time.Time{})
	log.Printf("alice broker: init OK")

	go b.pingLoop()

	b.mu.Lock()
	syms := make([]string, len(b.symbols))
	copy(syms, b.symbols)
	initSyms := make([]string, len(b.initialSymbols))
	copy(initSyms, b.initialSymbols)
	b.mu.Unlock()
	log.Printf("alice broker: subscribing to %d symbols", len(syms))
	if len(syms) > 0 {
		b.sendSubscribe(syms)
	} else if len(initSyms) > 0 {
		b.sendSubscribe(initSyms)
	}

	b.readLoop()
}

func (b *AliceBrokerClient) pingLoop() {
	for {
		time.Sleep(40 * time.Second)
		b.writeMu.Lock()
		b.mu.Lock()
		conn := b.conn
		b.mu.Unlock()
		if conn == nil {
			b.writeMu.Unlock()
			return
		}
		conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
		conn.WriteMessage(websocket.TextMessage, []byte(`{"k":"","t":"h"}`))
		b.writeMu.Unlock()
	}
}

func (b *AliceBrokerClient) readLoop() {
	defer func() {
		b.mu.Lock()
		if b.conn != nil {
			b.conn.Close()
			b.conn = nil
		}
		b.mu.Unlock()
	}()

	for {
		b.mu.Lock()
		conn := b.conn
		b.mu.Unlock()
		if conn == nil {
			return
		}
		_, msg, err := conn.ReadMessage()
		if err != nil {
			if b.running {
				log.Printf("alice broker: read error: %v", err)
			}
			return
		}
		tick := b.parseAliceTick(msg)
		if tick != nil {
			b.hub.BroadcastTick(tick)
		}
	}
}

func (b *AliceBrokerClient) Subscribe(symbols []string) {
	b.mu.Lock()
	b.symbols = symbols
	b.mu.Unlock()
	b.sendSubscribe(symbols)
}

func (b *AliceBrokerClient) Unsubscribe() {
	b.mu.Lock()
	b.symbols = nil
	b.mu.Unlock()
}

func (b *AliceBrokerClient) sendSubscribe(symbols []string) {
	if len(symbols) == 0 {
		return
	}
	var parts []string
	for _, sym := range symbols {
		sym = mapAliceSymbol(sym)
		parts = append(parts, sym)
	}
	k := strings.Join(parts, "#")
	fmt.Printf("alice broker: subscribing to symbols: %v (mapped: %s)\n", symbols, k)
	msg := map[string]string{"k": k, "t": "t"}
	body, _ := json.Marshal(msg)
	b.writeMu.Lock()
	b.mu.Lock()
	conn := b.conn
	b.mu.Unlock()
	if conn == nil {
		b.writeMu.Unlock()
		return
	}
	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	err := conn.WriteMessage(websocket.TextMessage, body)
	b.writeMu.Unlock()
	if err != nil {
		log.Printf("alice broker: subscribe error: %v", err)
	}
}

func mapAliceSymbol(sym string) string {
	parts := strings.SplitN(sym, "|", 2)
	if len(parts) != 2 {
		return sym
	}
	exch, token := parts[0], parts[1]
	exchMap := map[string]string{
		"NSE_INDEX": "NSE",
		"BSE_INDEX": "BSE",
		"MCX_INDEX": "MCX",
	}
	if mapped, ok := exchMap[exch]; ok {
		exch = mapped
		if len(token) >= 5 && token[:5] == "99926" {
			token = token[3:]
		} else if len(token) >= 4 && token[:3] == "999" {
			token = token[3:]
		}
	}
	return exch + "|" + token
}

func (b *AliceBrokerClient) Stop() {
	b.running = false
	close(b.stopCh)
	b.mu.Lock()
	if b.conn != nil {
		b.conn.Close()
		b.conn = nil
	}
	b.mu.Unlock()
}

func computeAliceSuberToken(sessionToken string) string {
	h1 := sha256.Sum256([]byte(sessionToken))
	h1hex := hex.EncodeToString(h1[:])
	h2 := sha256.Sum256([]byte(h1hex))
	return hex.EncodeToString(h2[:])
}

func (b *AliceBrokerClient) parseAliceTick(msg []byte) map[string]any {
	var raw map[string]any
	if err := json.Unmarshal(msg, &raw); err != nil {
		return nil
	}

	t, _ := raw["t"].(string)
	if t != "tk" && t != "tf" {
		return nil
	}

	token, _ := raw["tk"].(string)
	if token == "" {
		return nil
	}

	tick := map[string]any{"token": token}

	if e, ok := raw["e"].(string); ok {
		tick["exchange"] = e
		tick["exchangeType"] = mapExchange(e)
	}

	if sym, ok := b.tokenSymbol[token]; ok {
		tick["symbol"] = sym
	} else if _, ok := raw["e"].(string); ok {
		prefixToken := "999" + token
		if sym, ok := b.tokenSymbol[prefixToken]; ok {
			tick["token"] = prefixToken
			tick["symbol"] = sym
		}
	}

	lp := parseAnyFloat(raw["lp"])
	if lp <= 0 {
		return nil
	}

	if t == "tk" {
		closeVal := parseAnyFloat(raw["c"])
		if closeVal > 0 {
			b.closeCache[token] = closeVal
		}
		if lp > 0 && closeVal > 0 {
			tick["change"] = lp - closeVal
			tick["close"] = closeVal
		}
		tick["ltp"] = lp

		if o, ok := raw["o"]; ok {
			tick["open"] = parseAnyFloat(o)
		}
		if h, ok := raw["h"]; ok {
			tick["high"] = parseAnyFloat(h)
		}
		if l, ok := raw["l"]; ok {
			tick["low"] = parseAnyFloat(l)
		}
		if v, ok := raw["v"].(string); ok && v != "" {
			tick["volume"] = parseInt64Str(v)
		}
		if oi, ok := raw["oi"].(string); ok && oi != "" {
			tick["oi"] = parseInt64Str(oi)
		}
	} else {
		tick["ltp"] = lp
		if closeVal, ok := b.closeCache[token]; ok && closeVal > 0 && lp > 0 {
			tick["change"] = lp - closeVal
			tick["close"] = closeVal
		}
		if pc, ok := raw["pc"].(string); ok && pc != "" && lp > 0 {
			pcVal := parseFloatStr(pc)
			if _, exists := tick["change"]; !exists && pcVal != 0 {
				tick["change"] = lp * pcVal / (100 + pcVal)
			}
		}
	}

if ap, ok := raw["ap"]; ok {
		tick["average"] = parseAnyFloat(ap)
	}
	if ft, ok := raw["ft"].(string); ok && ft != "" {
		tick["feedTime"] = ft
	}
	fmt.Printf("alice broker: parsed tick: %v\n", tick)
	return tick
}

func mapExchange(e string) int {
	switch e {
	case "NSE":
		return 1
	case "NFO":
		return 2
	case "BSE":
		return 3
	case "BFO":
		return 4
	case "MCX":
		return 5
	case "CDS":
		return 6
	default:
		return 0
	}
}

func parseAnyFloat(v any) float64 {
	switch value := v.(type) {
	case string:
		return parseFloatStr(value)
	case float64:
		return value
	case json.Number:
		f, _ := value.Float64()
		return f
	default:
		return 0
	}
}

func parseFloatStr(s string) float64 {
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

func parseInt64Str(s string) int64 {
	v, _ := strconv.ParseInt(s, 10, 64)
	return v
}
