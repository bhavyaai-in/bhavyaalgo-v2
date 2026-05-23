package server

import (
	"context"
	"database/sql"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"bhavyaaialgo/backend/brokers/aliceblue"
	marketdb "bhavyaaialgo/backend/db/market/gen"
	tradingdb "bhavyaaialgo/backend/db/trading/gen"
	internaldb "bhavyaaialgo/backend/internal/db"
	"bhavyaaialgo/backend/ws"

	"github.com/gorilla/websocket"
	"golang.org/x/crypto/bcrypt"
)

type Server struct {
	MarketDB           *sql.DB
	TradingDB          *sql.DB
	MarketQ            *marketdb.Queries
	TradingQ           *tradingdb.Queries
	Sessions           SessionStore
	Hub                *ws.Hub
	Upgrader           websocket.Upgrader
	Config             *Config
	adminPasswordHash  []byte
	logger             *slog.Logger
	mux                *http.ServeMux
	httpServer         *http.Server
	rateLimiters       map[string]*rateLimiter
	rateLimitMu        sync.Mutex
	done               chan struct{}
}

type Config struct {
	AdminEmail    string
	AdminPassword string
	Port          string
	DBPath        string
	BrokerAPIKey  string
	StaticDir     string
}

func New(cfg *Config, marketDB *sql.DB, tradingDB *sql.DB, marketQ *marketdb.Queries, tradingQ *tradingdb.Queries, sessions SessionStore, hub *ws.Hub, logger *slog.Logger) (*Server, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(cfg.AdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	s := &Server{
		MarketDB:          marketDB,
		TradingDB:         tradingDB,
		MarketQ:           marketQ,
		TradingQ:          tradingQ,
		Sessions:          sessions,
		Hub:               hub,
		Upgrader:          upgrader,
		Config:            cfg,
		adminPasswordHash: passwordHash,
		logger:            logger,
		mux:               http.NewServeMux(),
		rateLimiters:      make(map[string]*rateLimiter),
		done:              make(chan struct{}),
	}

	s.registerRoutes()
	return s, nil
}

func (s *Server) Mux() *http.ServeMux {
	return s.mux
}

func (s *Server) registerRoutes() {
	mux := s.mux

	mux.HandleFunc("GET /healthz", s.handleHealthz)
	mux.HandleFunc("GET /readyz", s.handleReadyz)
	mux.HandleFunc("GET /api/version", s.handleVersion)

	mux.HandleFunc("POST /api/login", s.rateLimit(s.handleLogin))
	mux.HandleFunc("GET /api/me", s.handleMe)
	mux.HandleFunc("POST /api/logout", s.handleLogout)

	mux.HandleFunc("GET /api/brokers", s.authMiddleware(s.handleListBrokers))
	mux.HandleFunc("POST /api/brokers", s.authMiddleware(s.handleCreateBroker))
	mux.HandleFunc("GET /api/brokers/{id}", s.authMiddleware(s.handleGetBroker))
	mux.HandleFunc("PUT /api/brokers/{id}", s.authMiddleware(s.handleUpdateBroker))
	mux.HandleFunc("DELETE /api/brokers/{id}", s.authMiddleware(s.handleDeleteBroker))
	mux.HandleFunc("GET /api/broker-list", s.authMiddleware(s.handleListBrokerList))
	mux.HandleFunc("GET /api/broker-columns", s.authMiddleware(s.handleBrokerColumns))

	mux.HandleFunc("GET /ws", s.wsAuthMiddleware(s.handleWebSocket))

	s.registerStaticRoutes(mux)
}

func (s *Server) registerStaticRoutes(mux *http.ServeMux) {
	if s.Config.StaticDir == "" {
		return
	}
	fs := http.FileServer(http.Dir(s.Config.StaticDir))
	mux.Handle("GET /assets/*", fs)
	mux.HandleFunc("GET /{path...}", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(s.Config.StaticDir, "index.html"))
	})
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.Error("ws upgrade error", "error", err)
		return
	}
	s.Hub.HandleWebSocket(conn)
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleReadyz(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := s.MarketDB.PingContext(ctx); err != nil {
		writeError(w, http.StatusServiceUnavailable, "not ready")
		return
	}
	if err := s.TradingDB.PingContext(ctx); err != nil {
		writeError(w, http.StatusServiceUnavailable, "not ready")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

func (s *Server) Run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go s.autoStartBrokerStream(context.Background())

	s.httpServer = &http.Server{
		Addr:         ":" + s.Config.Port,
		Handler:      s.corsMiddleware(s.requestLogging(s.mux)),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		s.logger.Info("server starting", "addr", "http://localhost:"+s.Config.Port)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("server error", "error", err)
		}
	}()

	<-ctx.Done()
	stop()

	s.logger.Info("shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s.Hub.StopBroker()

	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		s.logger.Error("server shutdown error", "error", err)
		return err
	}

	if err := s.MarketDB.Close(); err != nil {
		s.logger.Error("market db close error", "error", err)
	}
	if err := s.TradingDB.Close(); err != nil {
		s.logger.Error("trading db close error", "error", err)
	}

	close(s.done)
	s.logger.Info("server stopped")
	return nil
}

func (s *Server) Done() <-chan struct{} {
	return s.done
}

func (s *Server) autoStartBrokerStream(ctx context.Context) {
	defer func() {
		if rec := recover(); rec != nil {
			log.Printf("PANIC in broker autostart: %v", rec)
		}
	}()

	tokenSymbol := internaldb.LoadTokenSymbolMap(ctx, s.MarketDB)
	symbols := internaldb.LoadWatchlistSymbols(ctx, s.MarketDB)

	var aliceClientID, aliceSession, angelClient, angelAuth, angelFeed, angelKey string

	s.TradingDB.QueryRowContext(ctx,
		`SELECT broker_userid, broker_token FROM brokers WHERE broker_name='aliceblue' AND token_status='connected' AND DATE(broker_token_date) = DATE('now') LIMIT 1`,
	).Scan(&aliceClientID, &aliceSession)
	s.TradingDB.QueryRowContext(ctx,
		`SELECT broker_userid, broker_token, feed_token, broker_api FROM brokers WHERE broker_name='angel' AND token_status='connected' AND DATE(broker_token_date) = DATE('now') LIMIT 1`,
	).Scan(&angelClient, &angelAuth, &angelFeed, &angelKey)

	if aliceSession != "" {
		if err := aliceblue.CreateWsSession(aliceSession, aliceClientID, "", ""); err != nil {
			log.Printf("aliceblue createWsSession: %v", err)
		}
		s.Hub.StartAliceBroker(aliceSession, aliceClientID, tokenSymbol, symbols)
		log.Printf("auto-start: aliceblue broker stream")
	} else if angelAuth != "" {
		s.Hub.StartBroker(angelClient, angelAuth, angelFeed, angelKey, tokenSymbol, symbols)
		log.Printf("auto-start: angel broker stream")
	}
}

func (s *Server) requestLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(ww, r)
		s.logger.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", ww.statusCode,
			"duration", time.Since(start).String(),
		)
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		allowedOrigins := []string{"http://localhost:5173", "http://localhost:3000", "http://localhost:8080"}
		allowOrigin := ""
		for _, allowed := range allowedOrigins {
			if origin == allowed {
				allowOrigin = origin
				break
			}
		}
		if allowOrigin == "" && origin != "" {
			allowOrigin = origin
		}
		if allowOrigin != "" {
			w.Header().Set("Access-Control-Allow-Origin", allowOrigin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

type rateLimiter struct {
	count    int
	window   time.Time
	max      int
	duration time.Duration
}

func (s *Server) rateLimit(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		s.rateLimitMu.Lock()
		rl, ok := s.rateLimiters[ip]
		if !ok {
			rl = &rateLimiter{window: time.Now(), max: 10, duration: time.Minute}
			s.rateLimiters[ip] = rl
		}
		if time.Since(rl.window) > rl.duration {
			rl.count = 0
			rl.window = time.Now()
		}
		rl.count++
		if rl.count > rl.max {
			s.rateLimitMu.Unlock()
			writeError(w, http.StatusTooManyRequests, "rate limit exceeded")
			return
		}
		s.rateLimitMu.Unlock()
		next(w, r)
	}
}

func FindStaticDir(existing, envDir string) string {
	if envDir != "" {
		if info, err := os.Stat(envDir); err == nil && info.IsDir() {
			return envDir
		}
	}
	if existing != "" {
		if info, err := os.Stat(existing); err == nil && info.IsDir() {
			return existing
		}
	}
	for _, d := range []string{"../frontend/dist", "./frontend/dist"} {
		abs, _ := filepath.Abs(d)
		if info, err := os.Stat(abs); err == nil && info.IsDir() {
			return abs
		}
	}
	return ""
}
