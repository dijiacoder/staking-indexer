package handler

import (
	"github.com/dijiacoder/staking-indexer/internal/logger"
	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
)

// SetZeroTokenEventHandler 事件处理器
type SetZeroTokenEventHandler struct {
	BaseEventHandler
}

func NewSetZeroTokenEventHandler() *SetZeroTokenEventHandler {
	return &SetZeroTokenEventHandler{
		BaseEventHandler: BaseEventHandler{eventName: "SetZeroToken"},
	}
}

func (h *SetZeroTokenEventHandler) HandleEvent(ctx *EventHandlerContext) error {
	log := ctx.Log

	// 从 Log 中解析数据
	zeroTokenAddress := common.BytesToAddress(log.Topics[1].Bytes())

	logger.Logger.Info("SetZeroToken",
		zap.String("zeroTokenAddress", zeroTokenAddress.Hex()),
	)

	return nil
}
