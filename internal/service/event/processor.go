package event

import (
	"context"

	"github.com/dijiacoder/staking-indexer/internal/logger"
	"github.com/dijiacoder/staking-indexer/internal/service/handler"
	"github.com/ethereum/go-ethereum/core/types"
	"go.uber.org/zap"
)

// EventProcessor 事件处理器
// 职责：编排 handler 的分发流程
type EventProcessor struct {
	handlerMgr *handler.EventHandlerManager
}

// NewEventProcessor 创建新的事件处理器
func NewEventProcessor() *EventProcessor {
	return &EventProcessor{
		handlerMgr: handler.NewEventHandlerManager(),
	}
}

// ProcessEvents 批量处理事件：分发到对应 handler
func (ep *EventProcessor) ProcessEvents(ctx context.Context, logs []types.Log) error {
	for _, log := range logs {
		if len(log.Topics) == 0 {
			continue
		}

		// 分发到 handler，让 handler 自己解析和处理
		if err := ep.handlerMgr.HandleEvent(log); err != nil {
			logger.Logger.Error("Failed to handle event",
				zap.Error(err),
				zap.String("tx_hash", log.TxHash.Hex()),
				zap.Uint("log_index", log.Index),
			)
			continue
		}
	}

	return nil
}
