package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"os"

	_ "modernc.org/sqlite"
)

var db *sql.DB

type Broker struct {
	ID              int64  `json:"id"`
	FriendlyName    string `json:"friendly_name"`
	BrokerUserid    string `json:"broker_userid"`
	BrokerPassword  string `json:"broker_password"`
	BrokerPin       string `json:"broker_pin"`
	BrokerQrKey     string `json:"broker_qr_key"`
	BrokerAPI       string `json:"broker_api"`
	BrokerAPISecret string `json:"broker_api_secret"`
	BrokerName      string `json:"broker_name"`
	TokenStatus     string `json:"token_status"`
	BrokerToken     string `json:"-"`
	BrokerTokenDate string `json:"broker_token_date"`
	FeedToken       string `json:"-"`
	IsActive        bool   `json:"is_active"`
	IsAutologin     bool   `json:"is_autologin"`
	IsDisabled      bool   `json:"is_disabled"`
	Message         string `json:"message"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

type BrokerListEntry struct {
	ID             int64  `json:"id"`
	Name           string `json:"broker_name"`
	BrokerImageURL string `json:"broker_image_url"`
	IsActive       bool   `json:"is_active"`
	Message        string `json:"message"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

type brokerListJSON struct {
	Name     string `json:"name"`
	ImageURL string `json:"broker_image_url"`
	IsActive bool   `json:"is_active"`
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
	createTables()
	loadBrokerList()
}

func createTables() {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS broker_list (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			broker_image_url TEXT NOT NULL DEFAULT '',
			is_active INTEGER NOT NULL DEFAULT 1,
			message TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL DEFAULT (datetime('now','localtime')),
			updated_at TEXT NOT NULL DEFAULT (datetime('now','localtime'))
		)
	`)
	if err != nil {
		log.Fatalf("failed to create broker_list table: %v", err)
	}

	_, err = db.Exec(`
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
			is_active INTEGER NOT NULL DEFAULT 0,
			is_autologin INTEGER NOT NULL DEFAULT 0,
			is_disabled INTEGER NOT NULL DEFAULT 0,
			message TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL DEFAULT (datetime('now','localtime')),
			updated_at TEXT NOT NULL DEFAULT (datetime('now','localtime'))
		)
	`)
	if err != nil {
		log.Fatalf("failed to create brokers table: %v", err)
	}

	// Add feed_token column if missing (existing databases)
	db.Exec(`ALTER TABLE brokers ADD COLUMN feed_token TEXT NOT NULL DEFAULT ''`)
}

func loadBrokerList() {
	var count int
	db.QueryRow("SELECT COUNT(*) FROM broker_list").Scan(&count)
	if count > 0 {
		return
	}

	data, err := os.ReadFile("brokers_list.json")
	if err != nil {
		log.Printf("brokers_list.json not found, skipping: %v", err)
		return
	}

	var entries []brokerListJSON
	if err := json.Unmarshal(data, &entries); err != nil {
		log.Printf("failed to parse brokers_list.json: %v", err)
		return
	}

	for _, e := range entries {
		_, err := db.Exec(
			"INSERT INTO broker_list (name, broker_image_url, is_active) VALUES (?, ?, ?)",
			e.Name, e.ImageURL, boolToInt(e.IsActive),
		)
		if err != nil {
			log.Printf("failed to insert broker list entry %q: %v", e.Name, err)
		}
	}
	log.Printf("loaded %d brokers from brokers_list.json", len(entries))
}

func listBrokerList() ([]BrokerListEntry, error) {
	rows, err := db.Query(`SELECT * FROM broker_list ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []BrokerListEntry
	for rows.Next() {
		var e BrokerListEntry
		var isActive int
		err := rows.Scan(&e.ID, &e.Name, &e.BrokerImageURL, &isActive, &e.Message, &e.CreatedAt, &e.UpdatedAt)
		if err != nil {
			return nil, err
		}
		e.IsActive = isActive != 0
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

func listBrokers() ([]Broker, error) {
	rows, err := db.Query(`SELECT * FROM brokers ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var brokers []Broker
	for rows.Next() {
		var b Broker
		var isActive, isAutologin, isDisabled int
		err := rows.Scan(
			&b.ID, &b.FriendlyName, &b.BrokerUserid, &b.BrokerPassword, &b.BrokerPin,
			&b.BrokerQrKey, &b.BrokerAPI, &b.BrokerAPISecret, &b.BrokerName,
			&b.TokenStatus, &b.BrokerToken, &b.BrokerTokenDate,
			&isActive, &isAutologin, &isDisabled,
			&b.Message, &b.CreatedAt, &b.UpdatedAt,
			&b.FeedToken,
		)
		if err != nil {
			return nil, err
		}
		b.IsActive = isActive != 0
		b.IsAutologin = isAutologin != 0
		b.IsDisabled = isDisabled != 0
		brokers = append(brokers, b)
	}
	return brokers, rows.Err()
}

func getBroker(id int64) (*Broker, error) {
	var b Broker
	var isActive, isAutologin, isDisabled int
	err := db.QueryRow(`SELECT * FROM brokers WHERE id = ?`, id).Scan(
		&b.ID, &b.FriendlyName, &b.BrokerUserid, &b.BrokerPassword, &b.BrokerPin,
		&b.BrokerQrKey, &b.BrokerAPI, &b.BrokerAPISecret, &b.BrokerName,
		&b.TokenStatus, &b.BrokerToken, &b.BrokerTokenDate,
		&isActive, &isAutologin, &isDisabled,
		&b.Message, &b.CreatedAt, &b.UpdatedAt,
		&b.FeedToken,
	)
	if err != nil {
		return nil, err
	}
	b.IsActive = isActive != 0
	b.IsAutologin = isAutologin != 0
	b.IsDisabled = isDisabled != 0
	return &b, nil
}

func createBroker(b *Broker) (int64, error) {
	result, err := db.Exec(`
		INSERT INTO brokers (
			friendly_name, broker_userid, broker_password, broker_pin, broker_qr_key,
			broker_api, broker_api_secret, broker_name,
			token_status, broker_token, broker_token_date,
			feed_token,
			is_active, is_autologin, is_disabled,
			message
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		b.FriendlyName, b.BrokerUserid, b.BrokerPassword, b.BrokerPin, b.BrokerQrKey,
		b.BrokerAPI, b.BrokerAPISecret, b.BrokerName,
		b.TokenStatus, b.BrokerToken, b.BrokerTokenDate,
		b.FeedToken,
		boolToInt(b.IsActive), boolToInt(b.IsAutologin), boolToInt(b.IsDisabled),
		b.Message,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func updateBroker(id int64, b *Broker) error {
	_, err := db.Exec(`
		UPDATE brokers SET
			friendly_name=?, broker_userid=?, broker_password=?, broker_pin=?, broker_qr_key=?,
			broker_api=?, broker_api_secret=?, broker_name=?,
			token_status=?, broker_token=?, broker_token_date=?,
			feed_token=?,
			is_active=?, is_autologin=?, is_disabled=?,
			message=?,
			updated_at=datetime('now','localtime')
		WHERE id=?
	`,
		b.FriendlyName, b.BrokerUserid, b.BrokerPassword, b.BrokerPin, b.BrokerQrKey,
		b.BrokerAPI, b.BrokerAPISecret, b.BrokerName,
		b.TokenStatus, b.BrokerToken, b.BrokerTokenDate,
		b.FeedToken,
		boolToInt(b.IsActive), boolToInt(b.IsAutologin), boolToInt(b.IsDisabled),
		b.Message, id,
	)
	return err
}

func deleteBroker(id int64) error {
	_, err := db.Exec(`DELETE FROM brokers WHERE id = ?`, id)
	return err
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
