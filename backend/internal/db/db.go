package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"bhavyaaialgo/backend/db/gen"
	"bhavyaaialgo/backend/internal/config"

	_ "modernc.org/sqlite"
)

func New(cfg *config.Config) (*sql.DB, *gen.Queries, error) {
	database, err := sql.Open("sqlite", cfg.DBPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open db: %w", err)
	}

	database.SetMaxOpenConns(1)
	database.SetMaxIdleConns(1)

	if err := database.Ping(); err != nil {
		database.Close()
		return nil, nil, fmt.Errorf("failed to ping db: %w", err)
	}

	q := gen.New(database)

	if _, err := database.Exec("PRAGMA journal_mode=WAL"); err != nil {
		log.Printf("pragma journal_mode: %v", err)
	}
	if _, err := database.Exec("PRAGMA synchronous=NORMAL"); err != nil {
		log.Printf("pragma synchronous: %v", err)
	}

	if err := createTables(database); err != nil {
		database.Close()
		return nil, nil, fmt.Errorf("create tables: %w", err)
	}

	if _, err := database.Exec(`DELETE FROM master_contracts WHERE id NOT IN (SELECT MIN(id) FROM master_contracts GROUP BY symbol, exchange, instrumenttype, token)`); err != nil {
		log.Printf("clean master_contracts duplicates: %v", err)
	}

	return database, q, nil
}

func createTables(database *sql.DB) error {
	for _, ddl := range AllTables {
		if _, err := database.Exec(ddl); err != nil {
			return fmt.Errorf("exec ddl: %w", err)
		}
	}

	if _, err := database.Exec(`ALTER TABLE master_contracts ADD COLUMN broker_name TEXT NOT NULL DEFAULT ''`); err != nil {
		log.Printf("migrate master_contracts broker_name: %v", err)
	}
	if _, err := database.Exec(McBrokerIndexSQL); err != nil {
		log.Printf("create broker index: %v", err)
	}
	if _, err := database.Exec(`DELETE FROM master_contracts WHERE exchange IN ('NSE_INDEX','BSE_INDEX','MCX_INDEX')`); err != nil {
		log.Printf("clean index entries: %v", err)
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

func WithTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, timeout)
}
