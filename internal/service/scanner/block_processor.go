package scanner

import (
	"context"
	"math/big"

	"github.com/dijiacoder/staking-indexer/internal/gen/model"
	"github.com/dijiacoder/staking-indexer/internal/repository"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type BlockProcessor struct {
	repo    repository.ScannerRepository
	client  *ethclient.Client
	decoder *EventDecoder
}

func NewBlockProcessor(repo repository.ScannerRepository, client *ethclient.Client, decoder *EventDecoder) *BlockProcessor {
	return &BlockProcessor{repo: repo, client: client, decoder: decoder}
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

	// 3. Decode and collect events
	var events []*model.StakingEvent
	for _, log := range logs {
		decoded, err := p.decoder.DecodeLog(log)
		if err != nil {
			continue // Skip logs we don't recognize or fail to decode
		}

		event := &model.StakingEvent{
			ChainID:         chainID,
			ContractAddress: contractAddress,
			PoolID:          decoded.PoolID.Int64(),
			UserAddress:     decoded.User.Hex(),
			BlockNumber:     blockNumber,
			TxHash:          log.TxHash.Hex(),
			LogIndex:        int32(log.Index),
		}

		switch decoded.Name {
		case "Deposit":
			event.EventType = "Deposit"
			fAmount, _ := new(big.Float).SetInt(decoded.Amount).Float64()
			event.Amount = fAmount
		case "RequestUnstake":
			event.EventType = "Withdraw"
			fAmount, _ := new(big.Float).SetInt(decoded.Amount).Float64()
			event.Amount = fAmount
		case "Claim":
			event.EventType = "Claim"
			fAmount, _ := new(big.Float).SetInt(decoded.Reward).Float64()
			event.Amount = fAmount
		default:
			continue
		}
		events = append(events, event)
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

	// 5. Save events and update user positions in a single transaction
	if err := p.repo.SaveEventsAndProcessPositions(ctx, events); err != nil {
		return err
	}

	return nil
}

func (p *BlockProcessor) GetHeader(ctx context.Context, blockNumber int64) (*model.ChainBlock, error) {
	// Note: We can't populate ChainID here as it's not passed to this method.
	// The caller should set the ChainID separately if needed.
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
