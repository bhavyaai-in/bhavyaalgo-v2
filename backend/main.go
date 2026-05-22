package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"bhavyaaialgo/backend/blueprints"
	"bhavyaaialgo/backend/brokers/aliceblue"
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
	fmt.Printf("Starting server with config: %+v\n", cfg)
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
	app.RegisterStrategyRoutes(mux)
	app.RegisterSettingsRoutes(mux)

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

	// Start broker stream with priority: aliceblue > angel
	go func() {
		tokenSymbol := loadTokenSymbolMap(db)
		symbols := loadWatchlistSymbols(db)

		var aliceClientID, aliceSession, angelClient, angelAuth, angelFeed, angelKey string

		db.QueryRow(`SELECT broker_userid, broker_token FROM brokers WHERE broker_name='aliceblue' AND token_status='connected' AND DATE(broker_token_date) = DATE('now','localtime') LIMIT 1`).Scan(&aliceClientID, &aliceSession)
		db.QueryRow(`SELECT broker_userid, broker_token, feed_token, broker_api FROM brokers WHERE broker_name='angel' AND token_status='connected' AND DATE(broker_token_date) = DATE('now','localtime') LIMIT 1`).Scan(&angelClient, &angelAuth, &angelFeed, &angelKey)

		if aliceSession != "" {
			aliceblue.CreateWsSession(aliceSession, aliceClientID, "", "")
			hub.StartAliceBroker(aliceSession, aliceClientID, tokenSymbol, symbols)
			log.Printf("auto-start: aliceblue broker stream")
		} else if angelAuth != "" {
			hub.StartBroker(angelClient, angelAuth, angelFeed, angelKey, tokenSymbol, symbols)
			log.Printf("auto-start: angel broker stream")
		}
	}()

	staticDir := findStaticDir()
	if staticDir != "" {
		fs := http.FileServer(http.Dir(staticDir))
		mux.Handle("GET /assets/*", fs)
		mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, filepath.Join(staticDir, "index.html"))
		})
		mux.HandleFunc("GET /strategies", func(w http.ResponseWriter, r *http.Request) {
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
	for _, sql := range []string{watchlistsTableSQL, watchlistItemsTableSQL, watchlistItemsIdxSQL, masterContractsTableSQL, systemSettingsTableSQL, brokerListTableSQL, brokerColumnsTableSQL, brokersTableSQL, strategyTypesTableSQL, strategiesTableSQL, strategyInfoTableSQL, strategyJoinersTableSQL, positionsTableSQL, strategyPositionsTableSQL, ordersTableSQL} {
		if _, err := db.Exec(sql); err != nil {
			log.Fatalf("create table: %v", err)
		}
	}
	// Migrations for existing databases
	db.Exec(`ALTER TABLE master_contracts ADD COLUMN broker_name TEXT NOT NULL DEFAULT ''`)
	db.Exec(mcBrokerIndexSQL)

	// Clean up old index entries (NSE_INDEX/BSE_INDEX/MCX_INDEX)
	db.Exec(`DELETE FROM master_contracts WHERE exchange IN ('NSE_INDEX','BSE_INDEX','MCX_INDEX')`)
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
	rows, err := db.Query(`SELECT token, exchange, symbol FROM master_contracts`)
	if err != nil {
		log.Printf("load token symbol map: %v", err)
		return m
	}
	defer rows.Close()
	for rows.Next() {
		var token, exchange, symbol string
		if err := rows.Scan(&token, &exchange, &symbol); err != nil {
			continue
		}
		m[token] = symbol
		m[exchange+"|"+token] = symbol
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
	broker_name TEXT NOT NULL DEFAULT '',
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

const mcBrokerIndexSQL = `CREATE INDEX IF NOT EXISTS idx_mc_broker ON master_contracts(broker_name)`

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

const strategyTypesTableSQL = `
CREATE TABLE IF NOT EXISTS strategy_types (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	rules_explanation TEXT NOT NULL,
	created_at TEXT NOT NULL DEFAULT (datetime('now','localtime')),
	updated_at TEXT NOT NULL DEFAULT (datetime('now','localtime'))
)`

const strategiesTableSQL = `
CREATE TABLE IF NOT EXISTS strategies (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	strategy_secret_key TEXT NOT NULL,
	strategy_type INTEGER NOT NULL REFERENCES strategy_types(id),
	position_status INTEGER NOT NULL DEFAULT 0,
	instrument_token INTEGER NOT NULL DEFAULT 0,
	exchange TEXT NOT NULL DEFAULT '',
	side TEXT NOT NULL DEFAULT 'SELL',
	atm_otm REAL NOT NULL DEFAULT 0,
	image_url TEXT NOT NULL DEFAULT '',
	color TEXT NOT NULL,
	is_active INTEGER NOT NULL DEFAULT 0,
	is_locked INTEGER NOT NULL DEFAULT 0,
	message TEXT NOT NULL DEFAULT '',
	expiry_date TEXT NOT NULL,
	created_at TEXT NOT NULL DEFAULT (datetime('now','localtime')),
	updated_at TEXT NOT NULL DEFAULT (datetime('now','localtime'))
);
CREATE INDEX IF NOT EXISTS idx_strategies_type ON strategies(strategy_type);
CREATE INDEX IF NOT EXISTS idx_strategies_active ON strategies(is_active);`

const strategyInfoTableSQL = `
CREATE TABLE IF NOT EXISTS strategy_info (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	strategy_id INTEGER NOT NULL REFERENCES strategies(id) ON DELETE CASCADE,
	info TEXT NOT NULL,
	created_at TEXT NOT NULL DEFAULT (datetime('now','localtime')),
	updated_at TEXT NOT NULL DEFAULT (datetime('now','localtime'))
);
CREATE INDEX IF NOT EXISTS idx_si_strategy ON strategy_info(strategy_id);`

const strategyJoinersTableSQL = `
CREATE TABLE IF NOT EXISTS strategy_joiners (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	broker_id INTEGER NOT NULL REFERENCES brokers(id) ON DELETE CASCADE,
	strategy_id INTEGER NOT NULL REFERENCES strategies(id) ON DELETE CASCADE,
	quantity REAL NOT NULL,
	re_entry INTEGER NOT NULL,
	re_entry_triggered INTEGER NOT NULL DEFAULT 0,
	multiplier REAL NOT NULL DEFAULT 1,
	is_active INTEGER NOT NULL DEFAULT 0,
	created_at TEXT NOT NULL DEFAULT (datetime('now','localtime')),
	updated_at TEXT NOT NULL DEFAULT (datetime('now','localtime'))
);
CREATE INDEX IF NOT EXISTS idx_sj_broker ON strategy_joiners(broker_id);
CREATE INDEX IF NOT EXISTS idx_sj_strategy ON strategy_joiners(strategy_id);`

const positionsTableSQL = `
CREATE TABLE IF NOT EXISTS positions (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	broker_id INTEGER NOT NULL REFERENCES brokers(id),
	strategy_id INTEGER REFERENCES strategies(id) ON DELETE CASCADE,
	entry_order_id INTEGER UNIQUE,
	exit_order_id INTEGER UNIQUE,
	tradingsymbol TEXT NOT NULL,
	strategy_type INTEGER NOT NULL REFERENCES strategy_types(id),
	exchange TEXT NOT NULL,
	instrument_token INTEGER NOT NULL,
	broker_instrument_token INTEGER NOT NULL,
	quantity REAL NOT NULL DEFAULT 0,
	last_price REAL NOT NULL DEFAULT 0,
	buy_quantity REAL NOT NULL DEFAULT 0,
	sell_quantity REAL NOT NULL DEFAULT 0,
	multiplier REAL NOT NULL DEFAULT 0,
	side TEXT NOT NULL DEFAULT 'SELL',
	buy_price REAL NOT NULL DEFAULT 0,
	sell_price REAL NOT NULL DEFAULT 0,
	product TEXT NOT NULL,
	status TEXT NOT NULL DEFAULT '0',
	message TEXT NOT NULL DEFAULT '',
	created_at TEXT NOT NULL DEFAULT (datetime('now','localtime')),
	updated_at TEXT NOT NULL DEFAULT (datetime('now','localtime'))
);
CREATE INDEX IF NOT EXISTS idx_pos_broker ON positions(broker_id);
CREATE INDEX IF NOT EXISTS idx_pos_strategy ON positions(strategy_id);
CREATE INDEX IF NOT EXISTS idx_pos_strategy_type ON positions(strategy_type);`

const strategyPositionsTableSQL = `
CREATE TABLE IF NOT EXISTS strategy_positions (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	strategy_id INTEGER NOT NULL REFERENCES strategies(id) ON DELETE CASCADE,
	tradingsymbol TEXT NOT NULL,
	strategy_type INTEGER NOT NULL REFERENCES strategy_types(id),
	exchange TEXT NOT NULL,
	instrument_token INTEGER NOT NULL,
	quantity REAL NOT NULL DEFAULT 0,
	last_price REAL NOT NULL DEFAULT 0,
	buy_quantity REAL NOT NULL DEFAULT 0,
	multiplier REAL NOT NULL DEFAULT 0,
	sell_quantity REAL NOT NULL DEFAULT 0,
	side TEXT NOT NULL DEFAULT 'SELL',
	buy_price REAL NOT NULL DEFAULT 0,
	sell_price REAL NOT NULL DEFAULT 0,
	product TEXT NOT NULL,
	status TEXT NOT NULL DEFAULT '0',
	message TEXT NOT NULL DEFAULT '',
	created_at TEXT NOT NULL DEFAULT (datetime('now','localtime')),
	updated_at TEXT NOT NULL DEFAULT (datetime('now','localtime'))
);
CREATE INDEX IF NOT EXISTS idx_sp_strategy ON strategy_positions(strategy_id);`

const ordersTableSQL = `
CREATE TABLE IF NOT EXISTS orders (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	broker_id INTEGER NOT NULL REFERENCES brokers(id),
	strategy_id INTEGER REFERENCES strategies(id),
	position_id INTEGER REFERENCES positions(id),
	order_id TEXT NOT NULL,
	status_message TEXT NOT NULL DEFAULT '',
	tag TEXT NOT NULL,
	variety TEXT NOT NULL,
	tradingsymbol TEXT NOT NULL,
	exchange TEXT NOT NULL,
	instrument_token INTEGER NOT NULL,
	broker_instrument_token INTEGER NOT NULL,
	transaction_type TEXT NOT NULL,
	product TEXT NOT NULL,
	order_type TEXT NOT NULL,
	validity TEXT NOT NULL DEFAULT 'DAY',
	status TEXT NOT NULL,
	quantity REAL NOT NULL,
	price REAL NOT NULL,
	trigger_price REAL NOT NULL,
	average_price REAL NOT NULL,
	filled_quantity REAL NOT NULL,
	pending_quantity REAL NOT NULL,
	cancelled_quantity REAL NOT NULL,
	created_at TEXT NOT NULL DEFAULT (datetime('now','localtime')),
	updated_at TEXT NOT NULL DEFAULT (datetime('now','localtime'))
);
CREATE INDEX IF NOT EXISTS idx_orders_broker ON orders(broker_id);
CREATE INDEX IF NOT EXISTS idx_orders_strategy ON orders(strategy_id);
CREATE INDEX IF NOT EXISTS idx_orders_order_id ON orders(order_id);
CREATE INDEX IF NOT EXISTS idx_orders_tag ON orders(tag);`
