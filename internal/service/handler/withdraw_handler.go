package handler

import (
	"fmt"
	"math/big"

	"github.com/dijiacoder/staking-indexer/internal/logger"
	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
)

// WithdrawEventHandler 提现事件处理器
type WithdrawEventHandler struct {
	BaseEventHandler
}

func NewWithdrawEventHandler() *WithdrawEventHandler {
	return &WithdrawEventHandler{
		BaseEventHandler: BaseEventHandler{eventName: "Withdraw"},
	}
}

func (h *WithdrawEventHandler) HandleEvent(ctx *EventHandlerContext) error {
	log := ctx.Log

	if len(log.Topics) < 4 {
		return fmt.Errorf("insufficient topics for Withdraw event")
	}

	// 从 Log 中解析数据
	userAddress := common.BytesToAddress(log.Topics[1].Bytes())
	poolID := new(big.Int).SetBytes(log.Topics[2].Bytes())
	// blockNumber := new(big.Int).SetBytes(log.Topics[3].Bytes()) // 可选使用

	var amount *big.Int
	if len(log.Data) >= 32 {
		amount = new(big.Int).SetBytes(log.Data[0:32])
	}

	// TODO: 业务处理逻辑
	// 例如：更新 staking_pools 总质押量（减少）
	// 例如：更新 staking_user_positions 状态为"已提现"

	var amountFloat float64
	if amount != nil {
		amountFloat, _ = new(big.Float).SetInt(amount).Float64()
	}

	logger.Logger.Info("Withdraw event processed",
		zap.String("user", userAddress.Hex()),
		zap.Int64("pool_id", poolID.Int64()),
		zap.Float64("amount", amountFloat),
	)

	return nil
}
