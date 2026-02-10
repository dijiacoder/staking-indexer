package scanner

import (
	"context"
	"fmt"
	"time"

	"github.com/dijiacoder/staking-indexer/internal/repository"
	"github.com/ethereum/go-ethereum/ethclient"
)

type ScannerService struct {
	repo          repository.ScannerRepository
	client        *ethclient.Client
	processor     *BlockProcessor
	reorgHandler  *ReorgHandler
	chainID       int64
	contractAddr  string
	confirmations int64
	scanInterval  time.Duration
}

func NewScannerService(
	repo repository.ScannerRepository,
	client *ethclient.Client,
	chainID int64,
	contractAddr string,
	confirmations int64,
) (*ScannerService, error) {
	processor, err := NewBlockProcessor(repo, client)
	if err != nil {
		return nil, err
	}

	return &ScannerService{
		repo:          repo,
		client:        client,
		processor:     processor,
		reorgHandler:  NewReorgHandler(repo, client),
		chainID:       chainID,
		contractAddr:  contractAddr,
		confirmations: confirmations,
		scanInterval:  5 * time.Second,
	}, nil
}

// Start runs the scanner loop.
func (s *ScannerService) Start(ctx context.Context) error {
	fmt.Printf("Starting scanner for chain %d, contract %s\n", s.chainID, s.contractAddr)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := s.scan(ctx); err != nil {
				fmt.Printf("Scan error: %v\n", err)
			}
			time.Sleep(s.scanInterval)
		}
	}
}

func (s *ScannerService) scan(ctx context.Context) error {
	// 1. Get current cursor from DB
	cursor, err := s.repo.GetCursor(ctx, s.chainID, s.contractAddr)
	if err != nil {
		return fmt.Errorf("failed to get cursor: %w", err)
	}

	// 2. Get latest block number from the chain
	latestBlock, err := s.client.BlockNumber(ctx)
	if err != nil {
		return fmt.Errorf("failed to get latest block: %w", err)
	}

	// 3. Calculate the highest safe block we can process
	safeBlock := int64(latestBlock) - s.confirmations
	if safeBlock <= cursor.LastScannedBlock {
		return nil // Up to date
	}

	// 4. Batch process blocks up to safeBlock (limit to 100 blocks per scan iteration)
	endBlock := safeBlock
	if endBlock > cursor.LastScannedBlock+100 {
		endBlock = cursor.LastScannedBlock + 100
	}

	fmt.Printf("Scanning blocks %d to %d (latest: %d, safe: %d)\n",
		cursor.LastScannedBlock+1, endBlock, latestBlock, safeBlock)

	for nextBlock := cursor.LastScannedBlock + 1; nextBlock <= endBlock; nextBlock++ {
		fmt.Printf("Scanning block %d\n", nextBlock)
		// A. Fetch current block header for reorg verification
		header, err := s.processor.GetHeader(ctx, nextBlock)
		if err != nil {
			return fmt.Errorf("failed to get header for block %d: %w", nextBlock, err)
		}

		// B. Verify chain continuity (Reorg Detection)
		reorged, err := s.reorgHandler.CheckAndHandleReorg(ctx, s.chainID, s.contractAddr, nextBlock, header.ParentHash)
		if err != nil {
			return fmt.Errorf("reorg check failed at block %d: %w", nextBlock, err)
		}

		if reorged {
			fmt.Printf("Reorg handled at block %d. Restarting scan loop...\n", nextBlock)
			return nil // Exit scan to let next iteration start from new cursor
		}

		// C. Process events in the block
		if err := s.processor.ProcessBlock(ctx, s.chainID, s.contractAddr, nextBlock); err != nil {
			return fmt.Errorf("failed to process block %d: %w", nextBlock, err)
		}

		// D. Update cursor in DB
		if err := s.repo.UpdateCursor(ctx, s.chainID, s.contractAddr, nextBlock, nextBlock); err != nil {
			return fmt.Errorf("failed to update cursor at block %d: %w", nextBlock, err)
		}
	}

	return nil
}
