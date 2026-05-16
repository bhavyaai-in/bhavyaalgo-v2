package main

import (
	"bufio"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"bhavyaaialgo/backend/blueprints"
)

func main() {
	loadEnv()

	initDB()

	adminEmail := os.Getenv("ADMIN_EMAIL")
	adminPassword := os.Getenv("ADMIN_PASSWORD")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

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

	app := &blueprints.App{DB: db, Sessions: sessions}
	app.RegisterConnectBrokerRoutes(mux)
	app.RegisterBrokerProfileRoutes(mux)
	app.RegisterBrokerDataRoutes(mux)

	staticDir := findStaticDir()
	if staticDir != "" {
		fs := http.FileServer(http.Dir(staticDir))
		mux.Handle("GET /assets/*", fs)
		mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, filepath.Join(staticDir, "index.html"))
		})
	}

	addr := ":" + port
	log.Printf("Server starting on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

func loadEnv() {
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
			os.Setenv(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
		}
		break
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("json encode error: %v", err)
	}
}

func findStaticDir() string {
	candidates := []string{
		"../frontend/dist",
		"./frontend/dist",
	}
	for _, d := range candidates {
		abs, _ := filepath.Abs(d)
		if info, err := os.Stat(abs); err == nil && info.IsDir() {
			return abs
		}
	}
	return ""
}
