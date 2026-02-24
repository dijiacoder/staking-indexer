package handler

import (
	"fmt"
	"math/big"

	"github.com/dijiacoder/staking-indexer/internal/logger"
	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
)

// RequestUnstakeEventHandler 赎回请求事件处理器
type RequestUnstakeEventHandler struct {
	BaseEventHandler
}

func NewRequestUnstakeEventHandler() *RequestUnstakeEventHandler {
	return &RequestUnstakeEventHandler{
		BaseEventHandler: BaseEventHandler{eventName: "RequestUnstake"},
	}
}

func (h *RequestUnstakeEventHandler) HandleEvent(ctx *EventHandlerContext) error {
	log := ctx.Log

	if len(log.Topics) < 3 {
		return fmt.Errorf("insufficient topics for RequestUnstake event")
	}

	// 从 Log 中解析数据
	userAddress := common.BytesToAddress(log.Topics[1].Bytes())
	poolID := new(big.Int).SetBytes(log.Topics[2].Bytes())

	var amount *big.Int
	if len(log.Data) >= 32 {
		amount = new(big.Int).SetBytes(log.Data[0:32])
	}

	// TODO: 业务处理逻辑
	// 例如：更新 staking_user_positions 状态为"赎回中"
	// 例如：记录赎回请求时间

	var amountFloat float64
	if amount != nil {
		amountFloat, _ = new(big.Float).SetInt(amount).Float64()
	}

	logger.Logger.Info("RequestUnstake event processed",
		zap.String("user", userAddress.Hex()),
		zap.Int64("pool_id", poolID.Int64()),
		zap.Float64("amount", amountFloat),
	)

	return nil
}
