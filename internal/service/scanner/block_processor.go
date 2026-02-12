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
	eventProcessor *event.Processor
}

func NewBlockProcessor(repo repository.ScannerRepository, client *ethclient.Client) (*BlockProcessor, error) {
	return &BlockProcessor{
		repo:           repo,
		client:         client,
		eventProcessor: event.NewEventProcessor(repo),
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

	if len(logs) != 0 {

		logger.Logger.Info("Found logs in block",
			zap.Int("count", len(logs)),
			zap.Int64("block_number", blockNumber),
		)

		// 3. 分发事件到处理器
		if err := p.eventProcessor.ProcessEvents(ctx, chainID, contractAddress, logs); err != nil {
			logger.Logger.Error("process events error", zap.Error(err))
			return err
		}
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
		logger.Logger.Error("save block error", zap.Error(err))
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
