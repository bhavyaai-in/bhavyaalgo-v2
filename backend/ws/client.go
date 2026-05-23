package ws

import (
	"encoding/json"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 4096
	sendBufSize    = 256
)

func (h *Hub) HandleWebSocket(conn *websocket.Conn) {
	c := &Client{
		hub:  h,
		conn: conn,
		send: make(chan []byte, sendBufSize),
	}
	h.Register(c)
	go c.writePump()
	c.readPump()
}

func (c *Client) readPump() {
	defer func() {
		c.hub.Unregister(c)
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
		var req clientMessage
		if err := json.Unmarshal(msg, &req); err != nil {
			continue
		}
		switch req.Type {
		case "subscribe":
			c.hub.Subscribe(c, req.Symbols)
			c.send <- mustJSON(map[string]any{"type": "subscribed", "symbols": req.Symbols})
		case "unsubscribe":
			c.hub.Unsubscribe(c, req.Symbols)
			c.send <- mustJSON(map[string]any{"type": "unsubscribed", "symbols": req.Symbols})
		case "list":
			syms := c.hub.GetSubscribed()
			c.send <- mustJSON(map[string]any{"type": "list", "symbols": syms})
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.conn.WriteMessage(websocket.TextMessage, msg)
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func mustJSON(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		return []byte(`{"error":"encode failed"}`)
	}
	return b
}
