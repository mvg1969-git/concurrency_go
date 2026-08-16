package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"mv_db/internal/config"
	"mv_db/internal/database"
	"mv_db/internal/database/compute"
	"mv_db/internal/database/engine"
	"mv_db/internal/database/storage"
	"mv_db/network"

	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Printf("Can't load config.yaml (%v). Use default values.", err)
		cfg = config.NewDefaultConfig()
	}
	logger.Info("Config is loaded:", zap.Object("config", cfg))

	eng := engine.NewEngine()
	store, err := storage.NewStorage(eng, logger)
	if err != nil {
		log.Fatalf("failed to init storage: %v", err)
	}

	comp := compute.NewCompute()
	db, err := database.NewDatabase(comp, store, logger)
	if err != nil {
		log.Fatalf("failed to init database: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	var options []network.TCPServerOption
	if cfg.Network.MaxConnections != 0 {
		logger.Info("add max connactions: %v", zap.Int("max_connections", cfg.Network.MaxConnections))
		options = append(options, network.WithServerMaxConnectionsNumber(uint(cfg.Network.MaxConnections)))
	}
	server := network.NewTCPServer(logger, cfg.Network.Address, db, options...)
	if err := server.Start(ctx); err != nil {
		logger.Fatal("server error", zap.Error(err))
	}
}
