package main

import (
	"context"
	"errors"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/dijiacoder/staking-indexer/internal/config"
	"github.com/dijiacoder/staking-indexer/internal/logger"
	"github.com/dijiacoder/staking-indexer/internal/repository"
	"github.com/dijiacoder/staking-indexer/internal/service/scanner"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// 1. Initialize Configuration from file
	configPath := flag.String("config", "config/config.toml", "path to config file")
	flag.Parse()

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		logger.Logger.Fatal("Failed to load config", zap.Error(err))
	}

	// 2. Initialize Database Connection
	db, err := gorm.Open(mysql.Open(cfg.Database.DSN), &gorm.Config{})
	if err != nil {
		logger.Logger.Fatal("Failed to connect to database", zap.Error(err))
	}

	if cfg.Database.Debug {
		db = db.Debug()
	}

	// 3. Initialize Ethereum RPC Client
	client, err := ethclient.Dial(cfg.Ethereum.RPCURL)
	if err != nil {
		logger.Logger.Fatal("Failed to connect to Ethereum RPC", zap.Error(err))
	}

	// 4. Wire Dependencies (Repository & Service)
	repo := repository.NewScannerRepository(db)
	svc, err := scanner.NewScannerService(
		repo,
		client,
		cfg,
	)
	if err != nil {
		logger.Logger.Fatal("Failed to create scanner service", zap.Error(err))
	}

	// 5. Setup Context and Signal Handling for Graceful Shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		logger.Logger.Info("Received signal. Initiating graceful shutdown...",
			zap.String("signal", sig.String()))
		cancel()
	}()

	// 6. Execute Scanner Service
	logger.Logger.Info("ZeroToken Stake Scanner started",
		zap.Int64("chain_id", cfg.Ethereum.ChainID),
		zap.String("contract", cfg.Ethereum.ContractAddr))
	if err := svc.Start(ctx); err != nil && !errors.Is(err, context.Canceled) {
		logger.Logger.Fatal("Scanner execution error", zap.Error(err))
	}

	logger.Logger.Info("Scanner stopped")
}
