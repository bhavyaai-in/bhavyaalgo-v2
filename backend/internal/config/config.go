package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	AdminEmail       string
	AdminPassword    string
	Port             string
	DBPath           string
	MarketDBPath     string
	TradingDBPath    string
	HistoricalDBPath string
	BrokerAPIKey     string
	StaticDir        string
	AIAPIURL         string
	AIAPIKey         string
	AIModel          string
}

func Load() (*Config, error) {
	loadEnvFile()

	cfg := &Config{
		AdminEmail:       os.Getenv("ADMIN_EMAIL"),
		AdminPassword:    os.Getenv("ADMIN_PASSWORD"),
		Port:             os.Getenv("PORT"),
		DBPath:           os.Getenv("DB_PATH"),
		MarketDBPath:     os.Getenv("MARKET_DB_PATH"),
		TradingDBPath:    os.Getenv("TRADING_DB_PATH"),
		HistoricalDBPath: os.Getenv("HISTORICAL_DB_PATH"),
		BrokerAPIKey:     os.Getenv("BROKER_API_KEY"),
		StaticDir:        os.Getenv("STATIC_DIR"), // <-- Read from env
		AIAPIURL:         os.Getenv("AI_API_URL"),
		AIAPIKey:         os.Getenv("AI_API_KEY"),
		AIModel:          os.Getenv("AI_MODEL"),
	}

	// Apply defaults if empty
	if cfg.Port == "" {
		cfg.Port = "8081"
	}
	if cfg.DBPath == "" {
		cfg.DBPath = "data.db"
	}
	if cfg.MarketDBPath == "" {
		cfg.MarketDBPath = "db/data-market.db"
	}
	if cfg.TradingDBPath == "" {
		cfg.TradingDBPath = "db/data-trading.db"
	}

	if cfg.HistoricalDBPath == "" {
		cfg.HistoricalDBPath = "db/historical.duckdb"
	}

	if cfg.AdminEmail == "" || cfg.AdminPassword == "" {
		return nil, fmt.Errorf("ADMIN_EMAIL and ADMIN_PASSWORD must be set in .env")
	}

	return cfg, nil
}

func loadEnvFile() {
	candidates := []string{".env", "../.env"}
	for _, path := range candidates {
		f, err := os.Open(path)
		if err != nil {
			continue
		}
		defer f.Close()
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			if os.Getenv(key) == "" {
				os.Setenv(key, val)
			}
		}
		break
	}
}
