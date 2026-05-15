package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

type Message struct {
	Text string `json:"text"`
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/hello", handleHello)

	staticDir := findStaticDir()
	if staticDir != "" {
		fs := http.FileServer(http.Dir(staticDir))
		mux.Handle("GET /assets/*", fs)
		mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/" {
				http.ServeFile(w, r, filepath.Join(staticDir, "index.html"))
				return
			}
			http.ServeFile(w, r, filepath.Join(staticDir, "index.html"))
		})
	}

	addr := ":" + port
	log.Printf("Server starting on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

func handleHello(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Message{Text: "Hello from Go backend!"})
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
