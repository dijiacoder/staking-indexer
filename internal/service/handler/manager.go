package handler

import (
	"context"
	"fmt"

	"github.com/dijiacoder/staking-indexer/internal/logger"
	"github.com/dijiacoder/staking-indexer/internal/service/contracts"
	"github.com/ethereum/go-ethereum/core/types"
	"go.uber.org/zap"
)

// EventHandlerManager 事件处理器管理器
type EventHandlerManager struct {
	handlers        map[string]EventHandler
	stakingContract *contracts.StakingContract
}

// NewEventHandlerManager 创建新的事件处理器管理器
func NewEventHandlerManager() *EventHandlerManager {
	manager := &EventHandlerManager{
		handlers:        make(map[string]EventHandler),
		stakingContract: contracts.NewStakingContract(),
	}

	// 注册所有处理器
	manager.RegisterHandler(NewSetZeroTokenEventHandler())
	manager.RegisterHandler(NewAddPoolEventHandler())
	manager.RegisterHandler(NewDepositEventHandler())
	manager.RegisterHandler(NewRequestUnstakeEventHandler())
	manager.RegisterHandler(NewClaimEventHandler())
	manager.RegisterHandler(NewWithdrawEventHandler())

	return manager
}

// RegisterHandler 注册事件处理器
func (m *EventHandlerManager) RegisterHandler(handler EventHandler) {
	eventName := handler.GetEventName()
	m.handlers[eventName] = handler
}

// GetHandler 获取指定事件的处理器
func (m *EventHandlerManager) GetHandler(eventName string) (EventHandler, bool) {
	handler, exists := m.handlers[eventName]
	return handler, exists
}

// HandleEvent 根据原始日志分发到对应处理器
func (m *EventHandlerManager) HandleEvent(ctx context.Context, chainID int64, contractAddress string, log types.Log) error {
	if len(log.Topics) == 0 {
		logger.Logger.Error("log has no topics")
		return fmt.Errorf("log has no topics")
	}

	eventHash := log.Topics[0]
	eventName, exists := m.stakingContract.GetEventName(eventHash)
	if !exists {
		logger.Logger.Debug("unknown event hash",
			zap.String("hash", eventHash.Hex()),
		)
		return nil
	}

	handler, exists := m.GetHandler(eventName)
	if !exists {
		logger.Logger.Warn("event handler does not exist",
			zap.String("event", eventName),
		)
		return nil
	}

	eventCtx := &EventHandlerContext{
		Log:             log,
		ChainID:         chainID,
		ContractAddress: contractAddress,
	}
	return handler.HandleEvent(eventCtx)
}
