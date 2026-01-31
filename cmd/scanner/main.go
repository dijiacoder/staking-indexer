package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/dijiacoder/staking-indexer/internal/config"
	"github.com/dijiacoder/staking-indexer/internal/repository"
	"github.com/dijiacoder/staking-indexer/internal/service/scanner"
	"github.com/ethereum/go-ethereum/ethclient"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// 1. Initialize Configuration from file
	configPath := flag.String("config", "config/config.toml", "path to config file")
	flag.Parse()

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. Initialize Database Connection
	db, err := gorm.Open(mysql.Open(cfg.Database.DSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 3. Initialize Ethereum RPC Client
	client, err := ethclient.Dial(cfg.Ethereum.RPCURL)
	if err != nil {
		log.Fatalf("Failed to connect to Ethereum RPC: %v", err)
	}

	// 4. Wire Dependencies (Repository & Service)
	repo := repository.NewScannerRepository(db)
	svc, err := scanner.NewScannerService(
		repo,
		client,
		cfg.Ethereum.ChainID,
		cfg.Ethereum.ContractAddr,
		cfg.Ethereum.Confirmations,
	)
	if err != nil {
		log.Fatalf("Failed to create scanner service: %v", err)
	}

	// 5. Setup Context and Signal Handling for Graceful Shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Printf("Received signal: %v. Initiating graceful shutdown...", sig)
		cancel()
	}()

	// 6. Execute Scanner Service
	log.Printf("MetaNode Stake Scanner started. ChainID: %d, Contract: %s", cfg.Ethereum.ChainID, cfg.Ethereum.ContractAddr)
	if err := svc.Start(ctx); err != nil && !errors.Is(err, context.Canceled) {
		log.Fatalf("Scanner execution error: %v", err)
	}

	log.Println("Scanner stopped.")
}
