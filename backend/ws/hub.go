package ws

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

func NewUpgrader() websocket.Upgrader {
	return websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
}

// Hub manages frontend WebSocket clients and subscription state.
type Hub struct {
	mu         sync.RWMutex
	clients    map[*Client]bool
	subscribed map[string]int // symbol -> subscriber count
	broker     *BrokerClient  // Angel One stream
	aliceBroker *AliceBrokerClient
	brokerName string // current active broker ("angel", "aliceblue", "")
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		subscribed: make(map[string]int),
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
	// Decrement subscription counts for this client's symbols
	for _, sym := range c.symbols {
		h.subscribed[sym]--
		if h.subscribed[sym] <= 0 {
			delete(h.subscribed, sym)
		}
	}
	h.mu.Unlock()
	h.syncBroker()
}

// Subscribe adds symbols for a client and syncs broker subscriptions.
func (h *Hub) Subscribe(c *Client, symbols []string) {
	h.mu.Lock()
	for _, sym := range symbols {
		h.subscribed[sym]++
		c.symbols = append(c.symbols, sym)
	}
	h.mu.Unlock()
	h.syncBroker()
}

// Unsubscribe removes symbols for a client.
func (h *Hub) Unsubscribe(c *Client, symbols []string) {
	h.mu.Lock()
	for _, sym := range symbols {
		h.subscribed[sym]--
		if h.subscribed[sym] <= 0 {
			delete(h.subscribed, sym)
		}
		// Remove from client's symbol list
		for i := len(c.symbols) - 1; i >= 0; i-- {
			if c.symbols[i] == sym {
				c.symbols = append(c.symbols[:i], c.symbols[i+1:]...)
			}
		}
	}
	h.mu.Unlock()
	h.syncBroker()
}

// SyncBroker adjusts broker subscriptions based on current needs.
func (h *Hub) syncBroker() {
	h.mu.RLock()
	symbols := make([]string, 0, len(h.subscribed))
	for sym := range h.subscribed {
		symbols = append(symbols, sym)
	}
	count := len(h.subscribed)
	brokerName := h.brokerName
	h.mu.RUnlock()

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

// BroadcastTick sends a market tick to all connected clients.
func (h *Hub) BroadcastTick(tick map[string]any) {
	msg := map[string]any{
		"type": "tick",
		"data": tick,
	}
	h.mu.RLock()
	for c := range h.clients {
		select {
		case c.send <- msg:
		default:
			// Client send buffer full, drop tick for this client
		}
	}
	h.mu.RUnlock()
}

// BroadcastNotification sends a notification to all connected clients.
func (h *Hub) BroadcastNotification(notification map[string]any) {
	msg := map[string]any{
		"type": "notification",
		"data": notification,
	}
	h.mu.RLock()
	for c := range h.clients {
		select {
		case c.send <- msg:
		default:
		}
	}
	h.mu.RUnlock()
}

// SetBroker attaches the Angel One streaming client.
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
