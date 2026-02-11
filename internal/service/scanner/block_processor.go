package scanner

import (
	"context"
	"math/big"

	"github.com/dijiacoder/staking-indexer/internal/gen/model"
	"github.com/dijiacoder/staking-indexer/internal/logger"
	"github.com/dijiacoder/staking-indexer/internal/repository"
	"github.com/dijiacoder/staking-indexer/internal/service/event"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"
)

type BlockProcessor struct {
	repo           repository.ScannerRepository
	client         *ethclient.Client
	eventProcessor *event.EventProcessor
}

func NewBlockProcessor(repo repository.ScannerRepository, client *ethclient.Client) (*BlockProcessor, error) {
	return &BlockProcessor{
		repo:           repo,
		client:         client,
		eventProcessor: event.NewEventProcessor(),
	}, nil
}

func (p *BlockProcessor) ProcessBlock(ctx context.Context, chainID int64, contractAddress string, blockNumber int64) error {
	// 1. Get block header to save to chain_blocks
	header, err := p.client.HeaderByNumber(ctx, big.NewInt(blockNumber))
	if err != nil {
		return err
	}

	// 2. Fetch logs for this contract in this block
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(blockNumber),
		ToBlock:   big.NewInt(blockNumber),
		Addresses: []common.Address{common.HexToAddress(contractAddress)},
	}
	logs, err := p.client.FilterLogs(ctx, query)
	if err != nil {
		return err
	}

	logger.Logger.Info("Found logs in block",
		zap.Int("count", len(logs)),
		zap.Int64("block_number", blockNumber),
	)

	// 3. 分发到事件处理器
	if err := p.eventProcessor.ProcessEvents(ctx, logs); err != nil {
		return fmt.Errorf("failed to process events: %w", err)
	}

	// 4. Save block header for reorg detection
	blockModel := &model.ChainBlock{
		ChainID:     chainID,
		BlockNumber: blockNumber,
		BlockHash:   header.Hash().Hex(),
		ParentHash:  header.ParentHash.Hex(),
		IsConfirmed: 1,
	}
	if err := p.repo.SaveBlock(ctx, blockModel); err != nil {
		return err
	}

	return nil
}

func (p *BlockProcessor) GetHeader(ctx context.Context, blockNumber int64) (*model.ChainBlock, error) {
	header, err := p.client.HeaderByNumber(ctx, big.NewInt(blockNumber))
	if err != nil {
		return nil, err
	}
	return &model.ChainBlock{
		BlockNumber: blockNumber,
		BlockHash:   header.Hash().Hex(),
		ParentHash:  header.ParentHash.Hex(),
	}, nil
}
