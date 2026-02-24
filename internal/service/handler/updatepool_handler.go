package handler

import (
	"math/big"

	"github.com/dijiacoder/staking-indexer/internal/logger"
	"go.uber.org/zap"
)

type UpdatePoolEventHandler struct {
	BaseEventHandler
}

func NewUpdatePoolEventHandler() *UpdatePoolEventHandler {
	return &UpdatePoolEventHandler{
		BaseEventHandler: BaseEventHandler{eventName: "UpdatePool"},
	}
}

func (h *UpdatePoolEventHandler) HandleEvent(ctx *EventHandlerContext) error {
	log := ctx.Log

	poolID := new(big.Int).SetBytes(log.Topics[1].Bytes())
	lastRewardBlock := new(big.Int).SetBytes(log.Topics[2].Bytes())

	var totalZeroToken *big.Int

	if len(log.Data) >= 32 {
		totalZeroToken = new(big.Int).SetBytes(log.Data[0:32])
	}

	logger.Logger.Info("UpdatePool event processed and saved",
		zap.Int64("poolID", poolID.Int64()),
		zap.Int64("lastRewardBlock", lastRewardBlock.Int64()),
		zap.Int64("totalZeroToken", bigInt64(totalZeroToken)),
	)

	return nil
}

//func bigInt64(n *big.Int) int64 {
//	if n == nil {
//		return 0
//	}
//	return n.Int64()
//}
