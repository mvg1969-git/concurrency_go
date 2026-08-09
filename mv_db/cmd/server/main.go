package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"mv_db/internal/database"
	"mv_db/internal/database/compute"
	"mv_db/internal/database/engine"
	"mv_db/internal/database/storage"
	"mv_db/network"

	"go.uber.org/zap"
)

func main() {
	// Потом переделдать на чтение из конфига
	addr := "127.0.0.1:3223"

	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

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

	server := network.NewTCPServer(addr, db, logger)
	if err := server.Start(ctx); err != nil {
		logger.Fatal("server error", zap.Error(err))
	}
}
