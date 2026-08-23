package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"

	"mv_db/internal/common"
	"mv_db/internal/config"
	"mv_db/internal/database"
	"mv_db/internal/database/compute"
	"mv_db/internal/database/engine"
	"mv_db/internal/database/filesystem"
	"mv_db/internal/database/storage"
	"mv_db/internal/database/storage/wal"
	"mv_db/network"

	"go.uber.org/zap"
)

const (
	// defaultFlushingBatchSize    = 100
	// defaultFlushingBatchTimeout = time.Millisecond * 10
	defaultMaxSegmentSize = 10 << 20
	// defaultWALDataDirectory     = "./data/spider/wal"
)

func main() {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Printf("Can't load config.yaml (%v). Use default values.", err)
		cfg = config.NewDefaultConfig()
	}
	logger.Info("Config is loaded:", zap.Object("config", cfg))

	wal, err := CreateWAL(cfg.WAL, logger)
	if err != nil {
		log.Fatalf("failed to initialize wal: %v", err)
	}

	if wal != nil {
		wal.Start(ctx)
	}

	var storageOptions []storage.StorageOption
	if wal != nil {
		storageOptions = append(storageOptions, storage.WithWAL(wal))
	}

	eng := engine.NewEngine()
	store, err := storage.NewStorage(eng, logger, storageOptions...)
	if err != nil {
		log.Fatalf("failed to init storage: %v", err)
	}

	comp := compute.NewCompute()
	db, err := database.NewDatabase(comp, store, logger)
	if err != nil {
		log.Fatalf("failed to init database: %v", err)
	}

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

func CreateWAL(cfg *config.WALConfig, logger *zap.Logger) (*wal.WAL, error) {
	if logger == nil {
		return nil, errors.New("logger is invalid")
	} else if cfg == nil {
		return nil, nil
	}

	segmentsDirectory := filesystem.NewSegmentsDirectory(cfg.DataDirectory)
	reader, err := wal.NewLogsReader(segmentsDirectory)
	if err != nil {
		return nil, err
	}

	maxSegmentSize := defaultMaxSegmentSize
	if cfg.MaxSegmentSize != "" {
		size, err := common.ParseSize(cfg.MaxSegmentSize)
		if err != nil {
			return nil, errors.New("max segment size is incorrect")
		}

		maxSegmentSize = size
	}

	segment := filesystem.NewSegment(cfg.DataDirectory, maxSegmentSize)
	writer, err := wal.NewLogsWriter(segment, logger)
	if err != nil {
		return nil, err
	}

	return wal.NewWAL(writer, reader, cfg.FlushingBatchTimeout, cfg.FlushingBatchLength)
}
