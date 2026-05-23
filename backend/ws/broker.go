package ws

import (
	"encoding/binary"
	"encoding/json"
	"log"
	"math"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type BrokerClient struct {
	authToken           string
	feedToken           string
	clientCode          string
	apiKey              string
	hub                 *Hub
	tokenSymbol         map[string]string
	conn                *websocket.Conn
	mu                  sync.Mutex
	writeMu             sync.Mutex
	symbols             []string
	initialSymbols      []string
	running             bool
	stopCh              chan struct{}
	autoReconnect       bool
	reconnectMaxRetries int
	reconnectMaxDelay   time.Duration
	connectTimeout      time.Duration
	reconnectAttempt    int
	parser              AngelBinaryParser
}

const (
	brokerURL                   = "wss://smartapisocket.angelone.in/smart-stream"
	defaultReconnectMaxAttempts = 300
	reconnectMinDelay           = 5 * time.Second
	defaultReconnectMaxDelay    = 60 * time.Second
	defaultConnectTimeout       = 7 * time.Second

	ModeLTP       = 1
	ModeQuote     = 2
	ModeSnapQuote = 3
	ModeDepth     = 4

	ActionSubscribe   = 1
	ActionUnsubscribe = 0
)

var exchangeTypes = map[string]int{
	"NSE": 1, "NSE_INDEX": 1,
	"NFO": 2,
	"BSE": 3, "BSE_INDEX": 3,
	"BFO": 4,
	"MCX": 5, "MCX_INDEX": 5,
}

type AngelBinaryParser struct {
	Hub *Hub
}

func (bp *AngelBinaryParser) HandleIncomingStream(msg []byte, tokenSymbol map[string]string) {
	tick, ok := bp.parseTick(msg, tokenSymbol)
	if !ok {
		return
	}
	tick.Symbol = tokenSymbol[tick.Token]
	payload, err := json.Marshal(map[string]any{
		"type": "tick",
		"data": tick,
	})
	if err != nil {
		return
	}
	bp.Hub.BroadcastPreEncodedTick(tick.Token, payload)
}

func (bp *AngelBinaryParser) parseTick(buf []byte, tokenSymbol map[string]string) (Tick, bool) {
	if len(buf) < 27 {
		return Tick{}, false
	}
	mode := buf[0]
	exchType := int(buf[1])
	token := strings.TrimRight(string(buf[2:27]), "\x00")

	var t Tick
	t.Token = token
	t.ExchangeType = exchType

	switch mode {
	case ModeLTP:
		if len(buf) < 47 {
			return Tick{}, false
		}
		ltp := float64(int64(binary.LittleEndian.Uint32(buf[43:47]))) / 100
		if ltp <= 0 {
			return Tick{}, false
		}
		t.LTP = ltp
	case ModeQuote:
		if len(buf) < 123 {
			return Tick{}, false
		}
		ltp := int64(binary.LittleEndian.Uint64(buf[43:51]))
		if ltp <= 0 {
			return Tick{}, false
		}
		closeP := int64(binary.LittleEndian.Uint64(buf[115:123]))
		t.LTP = float64(ltp) / 100
		t.Change = float64(ltp-closeP) / 100
		t.Volume = int64(binary.LittleEndian.Uint64(buf[67:75]))
		t.Open = float64(int64(binary.LittleEndian.Uint64(buf[91:99]))) / 100
		t.High = float64(int64(binary.LittleEndian.Uint64(buf[99:107]))) / 100
		t.Low = float64(int64(binary.LittleEndian.Uint64(buf[107:115]))) / 100
		t.Close = float64(closeP) / 100
	case ModeSnapQuote:
		if len(buf) < 379 {
			return Tick{}, false
		}
		ltp := int64(binary.LittleEndian.Uint64(buf[43:51]))
		if ltp <= 0 {
			return Tick{}, false
		}
		closeP := int64(binary.LittleEndian.Uint64(buf[115:123]))
		t.LTP = float64(ltp) / 100
		t.Change = float64(ltp-closeP) / 100
		t.Volume = int64(binary.LittleEndian.Uint64(buf[67:75]))
		t.Open = float64(int64(binary.LittleEndian.Uint64(buf[91:99]))) / 100
		t.High = float64(int64(binary.LittleEndian.Uint64(buf[99:107]))) / 100
		t.Low = float64(int64(binary.LittleEndian.Uint64(buf[107:115]))) / 100
		t.Close = float64(closeP) / 100
		t.OI = int64(binary.LittleEndian.Uint64(buf[131:139]))
	default:
		log.Printf("broker: unknown mode %d for token %s", mode, token)
		return Tick{}, false
	}
	return t, true
}

func NewBrokerClient(clientCode, authToken, feedToken, apiKey string, hub *Hub, tokenSymbol map[string]string) *BrokerClient {
	return &BrokerClient{
		clientCode:          clientCode,
		authToken:           authToken,
		feedToken:           feedToken,
		apiKey:              apiKey,
		hub:                 hub,
		tokenSymbol:         tokenSymbol,
		stopCh:              make(chan struct{}),
		autoReconnect:       true,
		reconnectMaxRetries: defaultReconnectMaxAttempts,
		reconnectMaxDelay:   defaultReconnectMaxDelay,
		connectTimeout:      defaultConnectTimeout,
		parser:              AngelBinaryParser{Hub: hub},
	}
}

func (b *BrokerClient) Run() {
	b.running = true
	for b.running {
		if b.reconnectAttempt > b.reconnectMaxRetries {
			log.Printf("broker: max reconnect attempts (%d) reached, stopping", b.reconnectMaxRetries)
			return
		}
		if b.reconnectAttempt > 0 {
			nextDelay := time.Duration(math.Pow(2, float64(b.reconnectAttempt))) * time.Second
			if nextDelay > b.reconnectMaxDelay {
				nextDelay = b.reconnectMaxDelay
			}
			log.Printf("broker: reconnecting in %v (attempt %d/%d)", nextDelay, b.reconnectAttempt, b.reconnectMaxRetries)
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

func (b *BrokerClient) connect() {
	d := websocket.DefaultDialer
	d.HandshakeTimeout = b.connectTimeout

	header := http.Header{}
	header.Set("Authorization", b.authToken)
	header.Set("x-api-key", b.apiKey)
	header.Set("x-client-code", b.clientCode)
	header.Set("x-feed-token", b.feedToken)

	u, _ := url.Parse(brokerURL)
	conn, _, err := d.Dial(u.String(), header)
	if err != nil {
		log.Printf("broker: dial failed: %v", err)
		return
	}

	b.mu.Lock()
	b.conn = conn
	b.mu.Unlock()
	b.reconnectAttempt = 0
	log.Printf("broker: connected")

	go b.pingLoop()

	b.mu.Lock()
	syms := make([]string, len(b.symbols))
	copy(syms, b.symbols)
	initSyms := make([]string, len(b.initialSymbols))
	copy(initSyms, b.initialSymbols)
	b.mu.Unlock()
	if len(syms) > 0 {
		b.sendSubscribe(syms)
	} else if len(initSyms) > 0 {
		b.sendSubscribe(initSyms)
	}

	b.readLoop()
}

func (b *BrokerClient) pingLoop() {
	for {
		time.Sleep(10 * time.Second)
		b.writeMu.Lock()
		b.mu.Lock()
		conn := b.conn
		b.mu.Unlock()
		if conn == nil {
			b.writeMu.Unlock()
			return
		}
		conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
		conn.WriteMessage(websocket.TextMessage, []byte("ping"))
		b.writeMu.Unlock()
	}
}

func (b *BrokerClient) readLoop() {
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
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			if b.running {
				log.Printf("broker: read error: %v", err)
			}
			return
		}
		if msgType == websocket.TextMessage {
			continue
		}
		b.parser.HandleIncomingStream(msg, b.tokenSymbol)
	}
}

func (b *BrokerClient) Subscribe(symbols []string) {
	b.mu.Lock()
	b.symbols = symbols
	b.mu.Unlock()
	b.sendSubscribe(symbols)
}

func (b *BrokerClient) SetInitialSymbols(symbols []string) {
	b.mu.Lock()
	b.initialSymbols = symbols
	b.mu.Unlock()
}

func (b *BrokerClient) Unsubscribe() {
	b.mu.Lock()
	b.symbols = nil
	b.mu.Unlock()
}

func (b *BrokerClient) sendSubscribe(symbols []string) {
	if len(symbols) == 0 {
		return
	}
	groups := groupTokens(symbols)
	for _, g := range groups {
		req := map[string]any{
			"action": ActionSubscribe,
			"params": map[string]any{
				"mode": ModeQuote,
				"tokenList": []map[string]any{
					{"exchangeType": g.exchangeType, "tokens": g.tokens},
				},
			},
		}
		body, _ := json.Marshal(req)
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
			log.Printf("broker: subscribe error: %v", err)
		}
	}
}

func (b *BrokerClient) Stop() {
	b.running = false
	close(b.stopCh)
	b.mu.Lock()
	if b.conn != nil {
		b.conn.Close()
		b.conn = nil
	}
	b.mu.Unlock()
}

type tokenGroup struct {
	exchangeType int
	tokens       []string
}

func groupTokens(symbols []string) []tokenGroup {
	groups := map[int][]string{}
	var order []int
	for _, sym := range symbols {
		parts := strings.SplitN(sym, "|", 2)
		if len(parts) != 2 {
			continue
		}
		exType, ok := exchangeTypes[parts[0]]
		if !ok {
			continue
		}
		if _, exists := groups[exType]; !exists {
			order = append(order, exType)
		}
		groups[exType] = append(groups[exType], parts[1])
	}
	var result []tokenGroup
	for _, exType := range order {
		result = append(result, tokenGroup{exchangeType: exType, tokens: groups[exType]})
	}
	return result
}
