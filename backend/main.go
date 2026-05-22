package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"bhavyaaialgo/backend/blueprints"
	"bhavyaaialgo/backend/internal/config"
	internaldb "bhavyaaialgo/backend/internal/db"
	"bhavyaaialgo/backend/internal/server"
	"bhavyaaialgo/backend/internal/service"
	"bhavyaaialgo/backend/internal/setup"
	"bhavyaaialgo/backend/ws"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("config", "error", err)
		os.Exit(1)
	}
	logger.Info("starting server", "port", cfg.Port, "db", cfg.DBPath)

	database, q, err := internaldb.New(cfg)
	if err != nil {
		logger.Error("db init", "error", err)
		os.Exit(1)
	}

	sessions := service.NewSessionStore(24 * time.Hour)
	hub := ws.NewHub()

	ctx := context.Background()
	setup.SeedFromFile(ctx, q)
	go setup.DownloadMasterContract(ctx, q)

	staticDir := server.FindStaticDir("", os.Getenv("STATIC_DIR"))

	srv, err := server.New(&server.Config{
		AdminEmail:    cfg.AdminEmail,
		AdminPassword: cfg.AdminPassword,
		Port:          cfg.Port,
		DBPath:        cfg.DBPath,
		BrokerAPIKey:  cfg.BrokerAPIKey,
		StaticDir:     staticDir,
	}, database, q, sessions, hub, logger)
	if err != nil {
		logger.Error("server init", "error", err)
		os.Exit(1)
	}

	app := &blueprints.App{DB: database, Q: q, Sessions: sessions, Hub: hub}
	app.RegisterConnectBrokerRoutes(srv.Mux())
	app.RegisterBrokerProfileRoutes(srv.Mux())
	app.RegisterBrokerDataRoutes(srv.Mux())
	app.RegisterWatchlistRoutes(srv.Mux())
	app.RegisterStrategyRoutes(srv.Mux())
	app.RegisterSettingsRoutes(srv.Mux())

	if err := srv.Run(); err != nil {
		logger.Error("server", "error", err)
		os.Exit(1)
	}
}
