package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	AdminEmail     string
	AdminPassword  string
	Port           string
	DBPath         string
	MarketDBPath   string
	TradingDBPath  string
	BrokerAPIKey   string
}

func Load() (*Config, error) {
	loadEnvFile()

	cfg := &Config{
		AdminEmail:    os.Getenv("ADMIN_EMAIL"),
		AdminPassword: os.Getenv("ADMIN_PASSWORD"),
		Port:          os.Getenv("PORT"),
		DBPath:        os.Getenv("DB_PATH"),
		BrokerAPIKey:  os.Getenv("BROKER_API_KEY"),
	}

	if cfg.Port == "" {
		cfg.Port = "8080"
	}
	if cfg.DBPath == "" {
		cfg.DBPath = "data.db"
	}
	if cfg.MarketDBPath == "" {
		cfg.MarketDBPath = os.Getenv("MARKET_DB_PATH")
	}
	if cfg.MarketDBPath == "" {
		cfg.MarketDBPath = "data-market.db"
	}
	if cfg.TradingDBPath == "" {
		cfg.TradingDBPath = os.Getenv("TRADING_DB_PATH")
	}
	if cfg.TradingDBPath == "" {
		cfg.TradingDBPath = "data-trading.db"
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
