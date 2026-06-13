package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"runtime"
	"time"

	marketdb "bhavyaaialgo/backend/db/market/gen"
	tradingdb "bhavyaaialgo/backend/db/trading/gen"
	"bhavyaaialgo/backend/internal/config"

	_ "modernc.org/sqlite"
)

type Databases struct {
	Market   *sql.DB
	Trading  *sql.DB
	MarketQ  *marketdb.Queries
	TradingQ *tradingdb.Queries
}

func New(cfg *config.Config) (*Databases, error) {
	marketPath := cfg.MarketDBPath
	if marketPath == "" {
		marketPath = "db/data-market.db"
	}
	tradingPath := cfg.TradingDBPath
	if tradingPath == "" {
		tradingPath = "db/data-trading.db"
	}

	marketDB, err := sql.Open("sqlite", marketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open market db: %w", err)
	}
	tradingDB, err := sql.Open("sqlite", tradingPath)
	if err != nil {
		marketDB.Close()
		return nil, fmt.Errorf("failed to open trading db: %w", err)
	}

	maxConns := runtime.GOMAXPROCS(0) * 2
	for _, d := range []*sql.DB{marketDB, tradingDB} {
		d.SetMaxOpenConns(maxConns)
		d.SetMaxIdleConns(maxConns)
		d.SetConnMaxLifetime(30 * time.Minute)
	}

	if err := marketDB.Ping(); err != nil {
		marketDB.Close()
		tradingDB.Close()
		return nil, fmt.Errorf("failed to ping market db: %w", err)
	}
	if err := tradingDB.Ping(); err != nil {
		marketDB.Close()
		tradingDB.Close()
		return nil, fmt.Errorf("failed to ping trading db: %w", err)
	}

	for _, d := range []*sql.DB{marketDB, tradingDB} {
		if _, err := d.Exec("PRAGMA journal_mode=WAL"); err != nil {
			log.Printf("pragma journal_mode: %v", err)
		}
		if _, err := d.Exec("PRAGMA synchronous=NORMAL"); err != nil {
			log.Printf("pragma synchronous: %v", err)
		}
		if _, err := d.Exec("PRAGMA busy_timeout=5000"); err != nil {
			log.Printf("pragma busy_timeout: %v", err)
		}
		if _, err := d.Exec("PRAGMA cache_size=-8000"); err != nil {
			log.Printf("pragma cache_size: %v", err)
		}
	}

	mq := marketdb.New(marketDB)
	tq := tradingdb.New(tradingDB)

	if err := createMarketTables(marketDB); err != nil {
		marketDB.Close()
		tradingDB.Close()
		return nil, fmt.Errorf("create market tables: %w", err)
	}
	if err := createTradingTables(tradingDB); err != nil {
		marketDB.Close()
		tradingDB.Close()
		return nil, fmt.Errorf("create trading tables: %w", err)
	}
	if err := seedDefaultStrategyTypes(tradingDB); err != nil {
		marketDB.Close()
		tradingDB.Close()
		return nil, fmt.Errorf("seed strategy types: %w", err)
	}

	if _, err := marketDB.Exec(`DELETE FROM master_contracts WHERE id NOT IN (SELECT MIN(id) FROM master_contracts GROUP BY symbol, exchange, instrumenttype, token)`); err != nil {
		log.Printf("clean master_contracts duplicates: %v", err)
	}

	return &Databases{
		Market:   marketDB,
		Trading:  tradingDB,
		MarketQ:  mq,
		TradingQ: tq,
	}, nil
}

func createMarketTables(database *sql.DB) error {
	for _, ddl := range MarketDDLs {
		if _, err := database.Exec(ddl); err != nil {
			return fmt.Errorf("exec market ddl: %w", err)
		}
	}
	var hasBroker bool
	if err := database.QueryRow(`SELECT COUNT(*) > 0 FROM pragma_table_info('master_contracts') WHERE name='broker_name'`).Scan(&hasBroker); err == nil && !hasBroker {
		if _, err := database.Exec(`ALTER TABLE master_contracts ADD COLUMN broker_name TEXT NOT NULL DEFAULT ''`); err != nil {
			log.Printf("migrate master_contracts broker_name: %v", err)
		}
	}
	if _, err := database.Exec(McBrokerIndexSQL); err != nil {
		log.Printf("create broker index: %v", err)
	}
	if _, err := database.Exec(`DELETE FROM master_contracts WHERE exchange IN ('NSE_INDEX','BSE_INDEX','MCX_INDEX')`); err != nil {
		log.Printf("clean index entries: %v", err)
	}
	return nil
}

func createTradingTables(database *sql.DB) error {
	for _, ddl := range TradingDDLs {
		if _, err := database.Exec(ddl); err != nil {
			return fmt.Errorf("exec trading ddl: %w", err)
		}
	}
	return nil
}

func seedDefaultStrategyTypes(database *sql.DB) error {
	var count int
	if err := database.QueryRow(`SELECT COUNT(*) FROM strategy_types`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	defaultTypes := []string{"single_leg", "multi_leg"}
	for _, name := range defaultTypes {
		if _, err := database.Exec(`INSERT INTO strategy_types (name, rules_explanation) VALUES (?, ?)`, name, ""); err != nil {
			return err
		}
	}
	return nil
}

func LoadWatchlistSymbols(ctx context.Context, database *sql.DB) []string {
	rows, err := database.QueryContext(ctx, `
		SELECT wi.token, COALESCE(mc.exchange, wi.exchange)
		FROM watchlist_items wi
		LEFT JOIN master_contracts mc ON
			mc.exchange = wi.exchange
			AND (mc.token = wi.token OR (wi.token LIKE '999%' AND mc.token = SUBSTR(wi.token, 4)))
		ORDER BY wi.watchlist_id, wi.sort_order
	`)
	if err != nil {
		log.Printf("load watchlist symbols: %v", err)
		return nil
	}
	defer rows.Close()
	var symbols []string
	for rows.Next() {
		var token, exchange string
		if err := rows.Scan(&token, &exchange); err != nil {
			log.Printf("scan watchlist row: %v", err)
			continue
		}
		symbols = append(symbols, exchange+"|"+token)
	}
	log.Printf("loaded %d watchlist symbols", len(symbols))
	return symbols
}

func LoadTokenSymbolMap(ctx context.Context, database *sql.DB) map[string]string {
	m := make(map[string]string)
	rows, err := database.QueryContext(ctx, `SELECT token, exchange, symbol FROM master_contracts`)
	if err != nil {
		log.Printf("load token symbol map: %v", err)
		return m
	}
	defer rows.Close()
	for rows.Next() {
		var token, exchange, symbol string
		if err := rows.Scan(&token, &exchange, &symbol); err != nil {
			log.Printf("scan token symbol row: %v", err)
			continue
		}
		m[token] = symbol
		m[exchange+"|"+token] = symbol
	}
	return m
}
