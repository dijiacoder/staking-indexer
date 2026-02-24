package handler

import (
	"fmt"
	"math/big"

	"github.com/dijiacoder/staking-indexer/internal/logger"
	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
)

// DepositEventHandler 存款事件处理器
type DepositEventHandler struct {
	BaseEventHandler
}

func NewDepositEventHandler() *DepositEventHandler {
	return &DepositEventHandler{
		BaseEventHandler: BaseEventHandler{eventName: "Deposit"},
	}
}

func (h *DepositEventHandler) HandleEvent(ctx *EventHandlerContext) error {
	log := ctx.Log

	if len(log.Topics) < 3 {
		return fmt.Errorf("insufficient topics for Deposit event")
	}

	// 从 Log 中解析数据
	userAddress := common.BytesToAddress(log.Topics[1].Bytes())
	poolID := new(big.Int).SetBytes(log.Topics[2].Bytes())

	var amount *big.Int
	if len(log.Data) >= 32 {
		amount = new(big.Int).SetBytes(log.Data[0:32])
	}

	// TODO: 业务处理逻辑
	// 例如：更新 staking_pools 总质押量
	// 例如：更新 staking_user_positions 用户持仓

	var amountFloat float64
	if amount != nil {
		amountFloat, _ = new(big.Float).SetInt(amount).Float64()
	}

	logger.Logger.Info("Deposit event processed",
		zap.String("user", userAddress.Hex()),
		zap.Int64("pool_id", poolID.Int64()),
		zap.Float64("amount", amountFloat),
	)

	return nil
}
