package event

import (
	"context"
	"fmt"

	"github.com/dijiacoder/staking-indexer/internal/logger"
	"github.com/dijiacoder/staking-indexer/internal/metrics"
	"github.com/dijiacoder/staking-indexer/internal/repository"
	"github.com/dijiacoder/staking-indexer/internal/service/contracts"
	"github.com/dijiacoder/staking-indexer/internal/service/handler"
	"github.com/ethereum/go-ethereum/core/types"
	"go.uber.org/zap"
)

// Processor 事件处理器
// 职责：编排 handler 的分发流程
type Processor struct {
	handlerMgr *handler.EventHandlerManager
}

// NewEventProcessor 创建新的事件处理器
func NewEventProcessor(repo repository.ScannerRepository) *Processor {
	return &Processor{
		handlerMgr: handler.NewEventHandlerManager(repo),
	}
}

// ProcessEvents 批量处理事件：分发到对应 handler
func (ep *Processor) ProcessEvents(ctx context.Context, chainID int64, contractAddress string, logs []types.Log) error {
	labels := map[string]string{
		"chain_id":        fmt.Sprintf("%d", chainID),
		"contract_address": contractAddress,
	}

	stakingContract := contracts.NewStakingContract()

	for _, log := range logs {
		if len(log.Topics) == 0 {
			continue
		}

		// 获取事件名称
		eventName, exists := stakingContract.GetEventName(log.Topics[0])
		if exists && eventName != "" {
			// 记录解析到的事件
			labels["event_type"] = eventName
			metrics.EventsTotal.With(labels).Inc()
		}

		// 分发到 handler，让 handler 自己解析和处理
		if err := ep.handlerMgr.HandleEvent(ctx, chainID, contractAddress, log); err != nil {
			logger.Logger.Error("Failed to handle event",
				zap.Error(err),
				zap.String("tx_hash", log.TxHash.Hex()),
				zap.Uint("log_index", log.Index),
			)
			// 记录失败事件
			if eventName != "" {
				metrics.EventsFailedTotal.With(labels).Inc()
			}
			continue
		}
		delete(labels, "event_type") // 清理 label 用于下一次循环
	}

	return nil
}
