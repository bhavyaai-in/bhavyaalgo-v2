package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

func NewUpgrader() websocket.Upgrader {
	return websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
}

type Hub struct {
	mu          sync.RWMutex
	clients     map[*Client]bool
	subscribers map[string]map[*Client]bool
	subscribed  map[string]int
	broker      *BrokerClient
	aliceBroker *AliceBrokerClient
	brokerName  string
	ltpCache    map[string]float64
}

func NewHub() *Hub {
	return &Hub{
		clients:     make(map[*Client]bool),
		subscribers: make(map[string]map[*Client]bool),
		subscribed:  make(map[string]int),
		ltpCache:    make(map[string]float64),
	}
}

func (h *Hub) Register(c *Client) {
	h.mu.Lock()
	h.clients[c] = true
	h.mu.Unlock()
}

func (h *Hub) Unregister(c *Client) {
	h.mu.Lock()
	delete(h.clients, c)
	for _, sym := range c.symbols {
		h.removeClientSubscriber(c, sym)
	}
	h.mu.Unlock()
	h.syncBroker()
}

func (h *Hub) removeClientSubscriber(c *Client, sym string) {
	// Remove from original key
	if subs, ok := h.subscribers[sym]; ok {
		delete(subs, c)
		if len(subs) == 0 {
			delete(h.subscribers, sym)
		}
	}
	h.subscribed[sym]--
	if h.subscribed[sym] <= 0 {
		delete(h.subscribed, sym)
	}
	// Remove from broker-mapped key (mapped entries are only in subscribers, not subscribed)
	mapped := mapToBrokerKey(sym, h.brokerName)
	if mapped != sym {
		if subs, ok := h.subscribers[mapped]; ok {
			delete(subs, c)
			if len(subs) == 0 {
				delete(h.subscribers, mapped)
			}
		}
	}
}

func (h *Hub) Subscribe(c *Client, symbols []string) {
	h.mu.Lock()
	for _, sym := range symbols {
		if h.subscribers[sym] == nil {
			h.subscribers[sym] = make(map[*Client]bool)
		}
		h.subscribers[sym][c] = true
		h.subscribed[sym]++
		c.symbols = append(c.symbols, sym)
	}
	h.mu.Unlock()
	h.syncBroker()
}

func (h *Hub) Unsubscribe(c *Client, symbols []string) {
	h.mu.Lock()
	for _, sym := range symbols {
		h.removeClientSubscriber(c, sym)
		for i := len(c.symbols) - 1; i >= 0; i-- {
			if c.symbols[i] == sym {
				c.symbols = append(c.symbols[:i], c.symbols[i+1:]...)
			}
		}
	}
	h.mu.Unlock()
	h.syncBroker()
}

func (h *Hub) syncBroker() {
	h.mu.Lock()
	symbols := make([]string, 0, len(h.subscribed))
	for sym := range h.subscribed {
		symbols = append(symbols, sym)
	}
	count := len(h.subscribed)
	brokerName := h.brokerName

	if count > 0 {
		// Ensure subscribers are also stored under broker-mapped keys so that
		// incoming ticks (with broker-format tokens) find the right subscribers.
		for _, sym := range symbols {
			mapped := mapToBrokerKey(sym, brokerName)
			if mapped != sym {
				if h.subscribers[mapped] == nil {
					h.subscribers[mapped] = make(map[*Client]bool)
				}
				if orig, ok := h.subscribers[sym]; ok {
					for c := range orig {
						h.subscribers[mapped][c] = true
					}
				}
			}
		}
	}
	h.mu.Unlock()

	if count > 0 {
		switch brokerName {
		case "aliceblue":
			if h.aliceBroker != nil {
				h.aliceBroker.Subscribe(symbols)
			}
		default:
			if h.broker != nil {
				h.broker.Subscribe(symbols)
			}
		}
	} else {
		switch brokerName {
		case "aliceblue":
			if h.aliceBroker != nil {
				h.aliceBroker.Unsubscribe()
			}
		default:
			if h.broker != nil {
				h.broker.Unsubscribe()
			}
		}
	}
}

func mapToBrokerKey(sym, brokerName string) string {
	// For both brokers, ticks come back with just the raw token number,
	// so subscribers must be stored under the raw token key.
	parts := strings.SplitN(sym, "|", 2)
	if len(parts) == 2 {
		return parts[1]
	}
	return sym
}

func (h *Hub) BroadcastPreEncodedTick(token string, payload []byte) {
	h.mu.RLock()
	clients, ok := h.subscribers[token]
	if !ok || len(clients) == 0 {
		h.mu.RUnlock()
		return
	}
	for c := range clients {
		select {
		case c.send <- payload:
		default:
		}
	}
	h.mu.RUnlock()
}

func (h *Hub) BroadcastTick(tick map[string]any) {
	msg := map[string]any{
		"type": "tick",
		"data": tick,
	}
	h.mu.RLock()
	for c := range h.clients {
		buf, _ := json.Marshal(msg)
		select {
		case c.send <- buf:
		default:
		}
	}
	h.mu.RUnlock()
}

func (h *Hub) BroadcastNotification(notification map[string]any) {
	msg := map[string]any{
		"type": "notification",
		"data": notification,
	}
	h.mu.RLock()
	for c := range h.clients {
		buf, _ := json.Marshal(msg)
		select {
		case c.send <- buf:
		default:
		}
	}
	h.mu.RUnlock()
}

func (h *Hub) SetBroker(b *BrokerClient) {
	h.broker = b
}

func (h *Hub) GetSubscribed() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	syms := make([]string, 0, len(h.subscribed))
	for s := range h.subscribed {
		syms = append(syms, s)
	}
	return syms
}

func (h *Hub) StartBroker(clientCode, authToken, feedToken, apiKey string, tokenSymbol map[string]string, initialSymbols []string) {
	priority := BrokerPriority("angel")
	h.mu.Lock()
	currentP := BrokerPriority(h.brokerName)
	if priority < currentP {
		h.mu.Unlock()
		log.Printf("broker: skipping angel (aliceblue has higher priority)")
		return
	}
	h.mu.Unlock()

	h.StopBroker()
	b := NewBrokerClient(clientCode, authToken, feedToken, apiKey, h, tokenSymbol)
	b.SetInitialSymbols(initialSymbols)
	h.mu.Lock()
	h.broker = b
	h.brokerName = "angel"
	h.mu.Unlock()
	go b.Run()
	log.Printf("broker: angel stream started")
}

func (h *Hub) StartAliceBroker(sessionToken, clientID string, tokenSymbol map[string]string, initialSymbols []string) {
	h.StopBroker()
	b := NewAliceBrokerClient(sessionToken, clientID, h, tokenSymbol)
	b.SetInitialSymbols(initialSymbols)
	h.mu.Lock()
	h.aliceBroker = b
	h.brokerName = "aliceblue"
	h.mu.Unlock()
	go b.Run()
	log.Printf("broker: aliceblue stream started")
}

func (h *Hub) StopBroker() {
	h.mu.Lock()
	if h.aliceBroker != nil {
		h.aliceBroker.Stop()
		h.aliceBroker = nil
	}
	if h.broker != nil {
		h.broker.Stop()
		h.broker = nil
	}
	h.brokerName = ""
	h.mu.Unlock()
}

func (h *Hub) UpdateLTP(token string, ltp float64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.ltpCache == nil {
		h.ltpCache = make(map[string]float64)
	}
	h.ltpCache[token] = ltp
}

func (h *Hub) GetLTP(token string) (float64, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.ltpCache == nil {
		return 0, false
	}
	ltp, ok := h.ltpCache[token]
	return ltp, ok
}
