package main

import (
	"crypto/tls"
	"database/sql"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	_ "modernc.org/sqlite"
)

const (
	ExchangeNSE = 1
	ExchangeNFO = 2
	ExchangeBSE = 3
	ExchangeBFO = 4
	ExchangeMCX = 5
)

var exchangeMap = map[string]int{
	"NSE": ExchangeNSE, "NSE_INDEX": ExchangeNSE,
	"NFO":             ExchangeNFO,
	"BSE": ExchangeBSE, "BSE_INDEX": ExchangeBSE,
	"BFO":             ExchangeBFO,
	"MCX": ExchangeMCX, "MCX_INDEX": ExchangeMCX,
}

type tokenGroup struct {
	exchangeType int
	tokens       []string
}

func main() {
	dbPath := "data.db"
	if len(os.Args) > 1 {
		dbPath = os.Args[1]
	}

	clientCode, authToken, feedToken, apiKey := getBrokerAuth(dbPath)
	groups := getWatchlistTokenGroups(dbPath)

	fmt.Printf("Broker: %s\n", clientCode)
	fmt.Printf("Groups:\n")
	for _, g := range groups {
		fmt.Printf("  exchangeType=%d tokens=%v\n", g.exchangeType, g.tokens)
	}

	u, _ := url.Parse("wss://smartapisocket.angelone.in/smart-stream")
	d := websocket.DefaultDialer
	d.HandshakeTimeout = 10 * time.Second
	d.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	header := http.Header{}
	header.Set("Authorization", "Bearer "+authToken)
	header.Set("x-api-key", apiKey)
	header.Set("x-client-code", clientCode)
	header.Set("x-feed-token", feedToken)

	fmt.Println("\nConnecting...")
	conn, _, err := d.Dial(u.String(), header)
	if err != nil {
		log.Fatalf("dial: %v", err)
	}
	defer conn.Close()
	fmt.Println("Connected!")

	var writeMu sync.Mutex

	for _, g := range groups {
		req := map[string]any{
			"action": 1,
			"params": map[string]any{
				"mode": 2,
				"tokenList": []map[string]any{
					{"exchangeType": g.exchangeType, "tokens": g.tokens},
				},
			},
		}
		b, _ := json.Marshal(req)
		writeMu.Lock()
		conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
		err := conn.WriteMessage(websocket.TextMessage, b)
		writeMu.Unlock()
		if err != nil {
			log.Fatalf("subscribe: %v", err)
		}
		fmt.Printf("Subscribed exchangeType=%d tokens=%v\n", g.exchangeType, g.tokens)
	}

	go func() {
		for {
			time.Sleep(10 * time.Second)
			writeMu.Lock()
			conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
			err := conn.WriteMessage(websocket.TextMessage, []byte("ping"))
			writeMu.Unlock()
			if err != nil {
				return
			}
		}
	}()

	fmt.Println("\nWaiting for ticks (Ctrl+C to stop)...")
	for {
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			log.Fatalf("read: %v", err)
		}
		if msgType == websocket.TextMessage {
			if string(msg) == "ping" {
				continue
			}
			fmt.Printf("Text: %s\n", string(msg))
			continue
		}
		printTick(msg)
	}
}

func printTick(buf []byte) {
	if len(buf) < 27 {
		fmt.Printf("Short binary (%d bytes): %x\n", len(buf), buf)
		return
	}

	mode := buf[0]
	exchType := buf[1]
	token := strings.TrimRight(string(buf[2:27]), "\x00")

	switch mode {
	case 1:
		if len(buf) < 47 {
			fmt.Printf("Short LTP (%d bytes)\n", len(buf))
			return
		}
		seq := binary.LittleEndian.Uint64(buf[27:35])
		ts := binary.LittleEndian.Uint64(buf[35:43])
		ltp := int64(binary.LittleEndian.Uint32(buf[43:47]))
		fmt.Printf("[LTP] exch=%d token=%s ltp=%d seq=%d ts=%d\n", exchType, token, ltp, seq, ts)

	case 2:
		if len(buf) < 123 {
			fmt.Printf("Short Quote (%d bytes)\n", len(buf))
			return
		}
		seq := binary.LittleEndian.Uint64(buf[27:35])
		ts := binary.LittleEndian.Uint64(buf[35:43])
		ltp := int64(binary.LittleEndian.Uint64(buf[43:51]))
		vol := int64(binary.LittleEndian.Uint64(buf[67:75]))
		open := int64(binary.LittleEndian.Uint64(buf[91:99]))
		high := int64(binary.LittleEndian.Uint64(buf[99:107]))
		low := int64(binary.LittleEndian.Uint64(buf[107:115]))
		closeP := int64(binary.LittleEndian.Uint64(buf[115:123]))
		fmt.Printf("[QUOTE] exch=%d token=%s ltp=%d vol=%d O=%d H=%d L=%d C=%d seq=%d ts=%d\n",
			exchType, token, ltp, vol, open, high, low, closeP, seq, ts)

	case 3:
		if len(buf) < 347 {
			fmt.Printf("Short SnapQuote (%d bytes)\n", len(buf))
			return
		}
		ltp := int64(binary.LittleEndian.Uint64(buf[43:51]))
		vol := int64(binary.LittleEndian.Uint64(buf[67:75]))
		oi := int64(binary.LittleEndian.Uint64(buf[139:147]))
		fmt.Printf("[SNAP] exch=%d token=%s ltp=%d vol=%d oi=%d\n", exchType, token, ltp, vol, oi)

	case 4:
		fmt.Printf("[DEPTH] exch=%d token=%s (%d bytes)\n", exchType, token, len(buf))

	default:
		fmt.Printf("Unknown mode=%d exch=%d token=%s (%d bytes): %x\n", mode, exchType, token, len(buf), buf)
	}
}

func getWatchlistTokenGroups(dbPath string) []tokenGroup {
	db := openDB(dbPath)
	defer db.Close()

	rows, err := db.Query(`
		SELECT wi.token, mc.exchange
		FROM watchlist_items wi
		JOIN master_contracts mc ON mc.token = wi.token AND mc.exchange = wi.exchange
		ORDER BY wi.watchlist_id, wi.sort_order
	`)
	if err != nil {
		log.Fatalf("query: %v", err)
	}
	defer rows.Close()

	groups := map[int][]string{}
	var order []int
	for rows.Next() {
		var token, exchange string
		if err := rows.Scan(&token, &exchange); err != nil {
			log.Fatalf("scan: %v", err)
		}
		exType, ok := exchangeMap[exchange]
		if !ok {
			log.Printf("WARN: unknown exchange %q for token %s, skipping", exchange, token)
			continue
		}
		if _, exists := groups[exType]; !exists {
			order = append(order, exType)
		}
		groups[exType] = append(groups[exType], token)
	}

	var result []tokenGroup
	for _, exType := range order {
		result = append(result, tokenGroup{exchangeType: exType, tokens: groups[exType]})
	}
	return result
}

func getBrokerAuth(dbPath string) (clientCode, authToken, feedToken, apiKey string) {
	db := openDB(dbPath)
	defer db.Close()
	err := db.QueryRow(
		`SELECT broker_userid, broker_token, feed_token, broker_api FROM brokers WHERE token_status='connected' AND DATE(broker_token_date) = DATE('now','localtime') LIMIT 1`,
	).Scan(&clientCode, &authToken, &feedToken, &apiKey)
	if err != nil {
		log.Fatalf("get broker auth: %v", err)
	}
	return
}

func openDB(path string) *sql.DB {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	return db
}
