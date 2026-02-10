package scanner

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/dijiacoder/staking-indexer/internal/gen/model"
	"github.com/dijiacoder/staking-indexer/internal/repository"
	"github.com/dijiacoder/staking-indexer/internal/service/contracts"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// EventData 解析后的事件数据结构
type EventData struct {
	EventType   string
	UserAddress common.Address
	PoolID      *big.Int
	Amount      *big.Int
}

type BlockProcessor struct {
	repo            repository.ScannerRepository
	client          *ethclient.Client
	contractABI     abi.ABI
	stakingContract *contracts.StakingContract
}

func NewBlockProcessor(repo repository.ScannerRepository, client *ethclient.Client) (*BlockProcessor, error) {
	// 解析ABI - 只包含我们真正需要解析的数据结构
	contractABIJSON := `[
		{
			"anonymous": false,
			"inputs": [
				{
					"indexed": true,
					"internalType": "address",
					"name": "user",
					"type": "address"
				},
				{
					"indexed": true,
					"internalType": "uint256",
					"name": "poolId",
					"type": "uint256"
				},
				{
					"indexed": false,
					"internalType": "uint256",
					"name": "amount",
					"type": "uint256"
				}
			],
			"name": "Deposit",
			"type": "event"
		},
		{
			"anonymous": false,
			"inputs": [
				{
					"indexed": true,
					"internalType": "address",
					"name": "user",
					"type": "address"
				},
				{
					"indexed": true,
					"internalType": "uint256",
					"name": "poolId",
					"type": "uint256"
				},
				{
					"indexed": false,
					"internalType": "uint256",
					"name": "amount",
					"type": "uint256"
				}
			],
			"name": "RequestUnstake",
			"type": "event"
		},
		{
			"anonymous": false,
			"inputs": [
				{
					"indexed": true,
					"internalType": "address",
					"name": "user",
					"type": "address"
				},
				{
					"indexed": true,
					"internalType": "uint256",
					"name": "poolId",
					"type": "uint256"
				},
				{
					"indexed": false,
					"internalType": "uint256",
					"name": "ZeroTokenReward",
					"type": "uint256"
				}
			],
			"name": "Claim",
			"type": "event"
		},
		{
			"anonymous": false,
			"inputs": [
				{
					"indexed": true,
					"internalType": "address",
					"name": "user",
					"type": "address"
				},
				{
					"indexed": true,
					"internalType": "uint256",
					"name": "poolId",
					"type": "uint256"
				},
				{
					"indexed": false,
					"internalType": "uint256",
					"name": "amount",
					"type": "uint256"
				},
				{
					"indexed": true,
					"internalType": "uint256",
					"name": "blockNumber",
					"type": "uint256"
				}
			],
			"name": "Withdraw",
			"type": "event"
		}
	]`

	parsedABI, err := abi.JSON(strings.NewReader(contractABIJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ABI: %w", err)
	}

	return &BlockProcessor{
		repo:            repo,
		client:          client,
		contractABI:     parsedABI,
		stakingContract: contracts.NewStakingContract(),
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
	fmt.Printf("Found %d logs in block %d\n", len(logs), blockNumber)

	// 3. Directly parse logs and collect events
	var events []*model.StakingEvent
	for _, log := range logs {
		if len(log.Topics) == 0 {
			continue
		}

		// 获取事件名称
		eventHash := log.Topics[0]
		eventName, exists := p.stakingContract.GetEventName(eventHash)
		if !exists {
			fmt.Printf("Skipping unknown event: %s\n", eventHash.Hex())
			continue
		}

		// 如果不是我们需要存储的事件，跳过
		if !p.stakingContract.IsUserInteractionEvent(eventName) {
			fmt.Printf("Skipping administrative event: %s\n", eventName)
			continue
		}

		// TODO

		// 解析事件数据
		eventData := p.parseEventData(eventName, log)
		if eventData == nil {
			continue // 解析失败或数据无效
		}

		// 构建事件对象
		event := &model.StakingEvent{
			ChainID:         chainID,
			ContractAddress: contractAddress,
			UserAddress:     eventData.UserAddress.Hex(),
			PoolID:          eventData.PoolID.Int64(),
			BlockNumber:     blockNumber,
			TxHash:          log.TxHash.Hex(),
			LogIndex:        int32(log.Index),
			EventType:       eventData.EventType,
		}

		// 设置金额
		if eventData.Amount != nil {
			fAmount, _ := new(big.Float).SetInt(eventData.Amount).Float64()
			event.Amount = fAmount
		}

		events = append(events, event)
		fmt.Printf("Processed event: Type=%s, User=%s, PoolID=%d, Amount=%v\n",
			eventData.EventType, eventData.UserAddress.Hex(),
			eventData.PoolID.Int64(), eventData.Amount)
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

// parseEventData 解析事件数据
func (p *BlockProcessor) parseEventData(eventName string, log types.Log) *EventData {
	var userAddress common.Address
	var poolID *big.Int
	var amount *big.Int

	switch eventName {
	case "Deposit", "RequestUnstake", "Claim":
		if len(log.Topics) >= 3 {
			userAddress = common.BytesToAddress(log.Topics[1].Bytes())
			poolID = new(big.Int).SetBytes(log.Topics[2].Bytes())
		}
		if len(log.Data) >= 32 {
			amount = new(big.Int).SetBytes(log.Data[0:32])
		}

	case "Withdraw":
		if len(log.Topics) >= 4 {
			userAddress = common.BytesToAddress(log.Topics[1].Bytes())
			poolID = new(big.Int).SetBytes(log.Topics[2].Bytes())
			// blockNumber 从 Topics[3] 获取
		}
		if len(log.Data) >= 32 {
			amount = new(big.Int).SetBytes(log.Data[0:32])
		}
	}

	// 验证必要字段
	if poolID == nil {
		fmt.Printf("Skipping event %s with nil poolID\n", eventName)
		return nil
	}

	return &EventData{
		EventType:   mapEventNameToType(eventName),
		UserAddress: userAddress,
		PoolID:      poolID,
		Amount:      amount,
	}
}

// mapEventNameToType 将事件名称映射到数据库中的事件类型
func mapEventNameToType(eventName string) string {
	switch eventName {
	case "Deposit":
		return "Deposit"
	case "RequestUnstake":
		return "Withdraw"
	case "Claim":
		return "Claim"
	case "Withdraw":
		return "WithdrawExecuted"
	default:
		return eventName
	}
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
