package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"bhavyaaialgo/backend/blueprints"
	"bhavyaaialgo/backend/db/gen"
	"bhavyaaialgo/backend/internal/config"
	"bhavyaaialgo/backend/internal/service"
	"bhavyaaialgo/backend/internal/setup"
	"bhavyaaialgo/backend/ws"
	_ "modernc.org/sqlite"
)

var hub = ws.NewHub()

var upgrader = ws.NewUpgrader()

var db *sql.DB
var Q *gen.Queries

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	initDB()
	ctx := context.Background()
	setup.SeedFromFile(ctx, Q)
	go setup.DownloadMasterContract(ctx, Q)
	adminEmail := cfg.AdminEmail
	adminPassword := cfg.AdminPassword

	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/login", handleLogin(adminEmail, adminPassword))
	mux.HandleFunc("GET /api/me", handleMe)
	mux.HandleFunc("POST /api/logout", handleLogout)

	mux.HandleFunc("GET /api/brokers", authMiddleware(handleListBrokers))
	mux.HandleFunc("POST /api/brokers", authMiddleware(handleCreateBroker))
	mux.HandleFunc("GET /api/brokers/{id}", authMiddleware(handleGetBroker))
	mux.HandleFunc("PUT /api/brokers/{id}", authMiddleware(handleUpdateBroker))
	mux.HandleFunc("DELETE /api/brokers/{id}", authMiddleware(handleDeleteBroker))
	mux.HandleFunc("GET /api/broker-list", authMiddleware(handleListBrokerList))
	mux.HandleFunc("GET /api/broker-columns", authMiddleware(handleBrokerColumns))

	app := &blueprints.App{DB: db, Q: Q, Sessions: service.Sessions, Hub: hub}
	app.RegisterConnectBrokerRoutes(mux)
	app.RegisterBrokerProfileRoutes(mux)
	app.RegisterBrokerDataRoutes(mux)
	app.RegisterWatchlistRoutes(mux)

	mux.HandleFunc("GET /ws", func(w http.ResponseWriter, r *http.Request) {
		// Upgrade requires auth via query param or header
		token := r.Header.Get("Authorization")
		if token == "" {
			token = r.URL.Query().Get("token")
		}
		if token == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		_, ok := service.Sessions.Get(token)
		if !ok {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("ws upgrade: %v", err)
			return
		}
		hub.HandleWebSocket(conn)
	})

	// Start Angel One broker stream if a connected broker exists
	go func() {
		var clientCode, authToken, feedToken, apiKey string
		err := db.QueryRow(`SELECT broker_userid, broker_token, feed_token, broker_api FROM brokers WHERE token_status='connected' AND DATE(broker_token_date) = DATE('now','localtime') LIMIT 1`).Scan(&clientCode, &authToken, &feedToken, &apiKey)
		if err == nil && clientCode != "" {
			tokenSymbol := loadTokenSymbolMap(db)
			symbols := loadWatchlistSymbols(db)
			hub.StartBroker(clientCode, authToken, feedToken, apiKey, tokenSymbol, symbols)
		}
	}()

	staticDir := findStaticDir()
	if staticDir != "" {
		fs := http.FileServer(http.Dir(staticDir))
		mux.Handle("GET /assets/*", fs)
		mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, filepath.Join(staticDir, "index.html"))
		})
	}

	addr := ":" + cfg.Port
	log.Printf("Server starting on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

func initDB() {
	var err error
	db, err = sql.Open("sqlite", "data.db")
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}
	if err = db.Ping(); err != nil {
		log.Fatalf("failed to ping db: %v", err)
	}
	Q = gen.New(db)
	db.Exec("PRAGMA journal_mode=WAL")
	createTables()
	db.Exec(`DELETE FROM master_contracts WHERE id NOT IN (SELECT MIN(id) FROM master_contracts GROUP BY symbol, exchange, instrumenttype, token)`)
}

func createTables() {
	for _, sql := range []string{watchlistsTableSQL, watchlistItemsTableSQL, watchlistItemsIdxSQL, masterContractsTableSQL, systemSettingsTableSQL, brokerListTableSQL, brokerColumnsTableSQL, brokersTableSQL} {
		if _, err := db.Exec(sql); err != nil {
			log.Fatalf("create table: %v", err)
		}
	}
}



func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("json encode error: %v", err)
	}
}

func loadWatchlistSymbols(db *sql.DB) []string {
	rows, err := db.Query(`SELECT wi.token, mc.exchange FROM watchlist_items wi JOIN master_contracts mc ON mc.token = wi.token AND mc.exchange = wi.exchange ORDER BY wi.watchlist_id, wi.sort_order`)
	if err != nil {
		log.Printf("load watchlist symbols: %v", err)
		return nil
	}
	defer rows.Close()
	var symbols []string
	for rows.Next() {
		var token, exchange string
		if err := rows.Scan(&token, &exchange); err != nil {
			continue
		}
		symbols = append(symbols, exchange+"|"+token)
	}
	log.Printf("loaded %d watchlist symbols", len(symbols))
	return symbols
}

func loadTokenSymbolMap(db *sql.DB) map[string]string {
	m := make(map[string]string)
	rows, err := db.Query(`SELECT token, symbol FROM master_contracts`)
	if err != nil {
		log.Printf("load token symbol map: %v", err)
		return m
	}
	defer rows.Close()
	for rows.Next() {
		var token, symbol string
		if err := rows.Scan(&token, &symbol); err != nil {
			continue
		}
		m[token] = symbol
	}
	return m
}

func findStaticDir() string {
	for _, d := range []string{"../frontend/dist", "./frontend/dist"} {
		abs, _ := filepath.Abs(d)
		if info, err := os.Stat(abs); err == nil && info.IsDir() {
			return abs
		}
	}
	return ""
}

const watchlistsTableSQL = `CREATE TABLE IF NOT EXISTS watchlists (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	sort_order INTEGER NOT NULL DEFAULT 0,
	created_at TEXT NOT NULL DEFAULT (datetime('now','localtime')),
	updated_at TEXT NOT NULL DEFAULT (datetime('now','localtime'))
)`

const watchlistItemsTableSQL = `CREATE TABLE IF NOT EXISTS watchlist_items (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	watchlist_id INTEGER NOT NULL REFERENCES watchlists(id) ON DELETE CASCADE,
	symbol TEXT NOT NULL,
	brsymbol TEXT NOT NULL DEFAULT '',
	name TEXT NOT NULL DEFAULT '',
	exchange TEXT NOT NULL DEFAULT '',
	token TEXT NOT NULL DEFAULT '',
	expiry TEXT NOT NULL DEFAULT '',
	strike REAL NOT NULL DEFAULT 0,
	lotsize INTEGER NOT NULL DEFAULT 0,
	instrumenttype TEXT NOT NULL DEFAULT '',
	tick_size REAL NOT NULL DEFAULT 0,
	sort_order INTEGER NOT NULL DEFAULT 0,
	created_at TEXT NOT NULL DEFAULT (datetime('now','localtime'))
)`

const watchlistItemsIdxSQL = `CREATE INDEX IF NOT EXISTS idx_wi_watchlist ON watchlist_items(watchlist_id)`

const masterContractsTableSQL = `
CREATE TABLE IF NOT EXISTS master_contracts (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	symbol TEXT NOT NULL,
	brsymbol TEXT NOT NULL DEFAULT '',
	name TEXT NOT NULL DEFAULT '',
	exchange TEXT NOT NULL DEFAULT '',
	brexchange TEXT NOT NULL DEFAULT '',
	token TEXT NOT NULL DEFAULT '',
	expiry TEXT NOT NULL DEFAULT '',
	strike REAL NOT NULL DEFAULT 0,
	lotsize INTEGER NOT NULL DEFAULT 0,
	instrumenttype TEXT NOT NULL DEFAULT '',
	tick_size REAL NOT NULL DEFAULT 0,
	created_at TEXT NOT NULL DEFAULT (datetime('now','localtime'))
);
CREATE INDEX IF NOT EXISTS idx_mc_symbol ON master_contracts(symbol);
CREATE INDEX IF NOT EXISTS idx_mc_exchange ON master_contracts(exchange);`

const systemSettingsTableSQL = `
CREATE TABLE IF NOT EXISTS system_settings (
	key TEXT PRIMARY KEY,
	value TEXT NOT NULL DEFAULT '',
	updated_at TEXT NOT NULL DEFAULT (datetime('now','localtime'))
)`

const brokerColumnsTableSQL = `
CREATE TABLE IF NOT EXISTS broker_columns (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	broker_name TEXT NOT NULL UNIQUE,
	columns_json TEXT NOT NULL DEFAULT '[]',
	created_at TEXT NOT NULL DEFAULT (datetime('now','localtime')),
	updated_at TEXT NOT NULL DEFAULT (datetime('now','localtime'))
)`

const brokerListTableSQL = `
CREATE TABLE IF NOT EXISTS broker_list (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL UNIQUE,
	broker_image_url TEXT NOT NULL DEFAULT '',
	is_active INTEGER NOT NULL DEFAULT 1,
	message TEXT NOT NULL DEFAULT '',
	created_at TEXT NOT NULL DEFAULT (datetime('now','localtime')),
	updated_at TEXT NOT NULL DEFAULT (datetime('now','localtime'))
)`

const brokersTableSQL = `
CREATE TABLE IF NOT EXISTS brokers (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	friendly_name TEXT NOT NULL DEFAULT '',
	broker_userid TEXT NOT NULL,
	broker_password TEXT NOT NULL,
	broker_pin TEXT NOT NULL,
	broker_qr_key TEXT NOT NULL DEFAULT '',
	broker_api TEXT NOT NULL DEFAULT '',
	broker_api_secret TEXT NOT NULL DEFAULT '',
	broker_name TEXT NOT NULL,
	token_status TEXT NOT NULL DEFAULT '',
	broker_token TEXT NOT NULL DEFAULT '',
	broker_token_date TEXT NOT NULL DEFAULT '2000-01-01 00:00:00',
	feed_token TEXT NOT NULL DEFAULT '',
	is_active INTEGER NOT NULL DEFAULT 0,
	is_autologin INTEGER NOT NULL DEFAULT 0,
	is_disabled INTEGER NOT NULL DEFAULT 0,
	message TEXT NOT NULL DEFAULT '',
	created_at TEXT NOT NULL DEFAULT (datetime('now','localtime')),
	updated_at TEXT NOT NULL DEFAULT (datetime('now','localtime'))
)`
