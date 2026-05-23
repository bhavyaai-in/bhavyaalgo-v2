package ws

import (
	"github.com/gorilla/websocket"
)

type Client struct {
	hub     *Hub
	conn    *websocket.Conn
	send    chan []byte
	symbols []string
}

type clientMessage struct {
	Type    string   `json:"type"`
	Symbols []string `json:"symbols,omitempty"`
}

type Tick struct {
	Token        string  `json:"token"`
	ExchangeType int     `json:"e"`
	Symbol       string  `json:"s,omitempty"`
	LTP          float64 `json:"ltp"`
	Change       float64 `json:"change"`
	Volume       int64   `json:"v,omitempty"`
	Open         float64 `json:"o,omitempty"`
	High         float64 `json:"h,omitempty"`
	Low          float64 `json:"l,omitempty"`
	Close        float64 `json:"close,omitempty"`
	OI           int64   `json:"oi,omitempty"`
	Token999     string  `json:"token_999,omitempty"`
}
