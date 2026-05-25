package historical

import (
	"database/sql"
	"fmt"
	"log"
	"sync"

	_ "github.com/marcboeker/go-duckdb"
)

type Store struct {
	db   *sql.DB
	once sync.Once
	path string
}

var globalStore *Store
var storeMu sync.Mutex

func GetStore() *Store {
	storeMu.Lock()
	defer storeMu.Unlock()
	if globalStore == nil {
		globalStore = &Store{path: "historical.duckdb"}
	}
	return globalStore
}

func (s *Store) Init() error {
	var err error
	s.once.Do(func() {
		s.db, err = sql.Open("duckdb", s.path)
		if err != nil {
			err = fmt.Errorf("duckdb open: %w", err)
			return
		}
		s.db.SetMaxOpenConns(1)
		_, err = s.db.Exec(`
			CREATE TABLE IF NOT EXISTS ohlcv (
				symbol VARCHAR NOT NULL,
				exchange VARCHAR NOT NULL,
				token VARCHAR NOT NULL,
				interval VARCHAR NOT NULL,
				timestamp VARCHAR NOT NULL,
				open DOUBLE NOT NULL,
				high DOUBLE NOT NULL,
				low DOUBLE NOT NULL,
				close DOUBLE NOT NULL,
				volume BIGINT NOT NULL,
				fetched_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				UNIQUE(symbol, exchange, token, interval, timestamp)
			)
		`)
		if err != nil {
			err = fmt.Errorf("duckdb create table: %w", err)
		}
	})
	return err
}

func (s *Store) Close() {
	if s.db != nil {
		s.db.Close()
	}
}

// HasData checks if data exists for the given symbol/exchange/token/interval within the date range
func (s *Store) HasData(symbol, exchange, token, interval, fromDate, toDate string) (bool, error) {
	if s.db == nil {
		return false, nil
	}
	var count int
	err := s.db.QueryRow(
		"SELECT COUNT(*) FROM ohlcv WHERE symbol = ? AND exchange = ? AND token = ? AND interval = ? AND timestamp >= ? AND timestamp <= ?",
		symbol, exchange, token, interval, fromDate, toDate,
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// SaveCandles bulk-inserts candle data, ignoring conflicts
func (s *Store) SaveCandles(symbol, exchange, token, interval string, candles []Candle) error {
	if s.db == nil || len(candles) == 0 {
		return nil
	}

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("duckdb begin: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(
		"INSERT OR IGNORE INTO ohlcv (symbol, exchange, token, interval, timestamp, open, high, low, close, volume) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
	)
	if err != nil {
		return fmt.Errorf("duckdb prepare: %w", err)
	}
	defer stmt.Close()


	for _, c := range candles {
		_, err = stmt.Exec(symbol, exchange, token, interval, c.Timestamp, c.Open, c.High, c.Low, c.Close, c.Volume)
		if err != nil {
			log.Printf("duckdb insert error: %v", err)
		}
	}

	return tx.Commit()
}

// GetCandles retrieves stored candles for the given parameters
func (s *Store) GetCandles(symbol, exchange, token, interval, fromDate, toDate string) ([]Candle, error) {
	if s.db == nil {
		return nil, nil
	}
	rows, err := s.db.Query(
		"SELECT timestamp, open, high, low, close, volume FROM ohlcv WHERE symbol = ? AND exchange = ? AND token = ? AND interval = ? AND timestamp >= ? AND timestamp <= ? ORDER BY timestamp",
		symbol, exchange, token, interval, fromDate, toDate,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var candles []Candle
	for rows.Next() {
		var c Candle
		if err := rows.Scan(&c.Timestamp, &c.Open, &c.High, &c.Low, &c.Close, &c.Volume); err != nil {
			continue
		}
		candles = append(candles, c)
	}
	return candles, nil
}
