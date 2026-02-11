package handler

import (
	"fmt"
	"math/big"

	"github.com/dijiacoder/staking-indexer/internal/logger"
	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
)

// ClaimEventHandler 领取奖励事件处理器
type ClaimEventHandler struct {
	BaseEventHandler
}

func NewClaimEventHandler() *ClaimEventHandler {
	return &ClaimEventHandler{
		BaseEventHandler: BaseEventHandler{eventName: "Claim"},
	}
}

func (h *ClaimEventHandler) HandleEvent(ctx *EventHandlerContext) error {
	log := ctx.Log

	if len(log.Topics) < 3 {
		return fmt.Errorf("insufficient topics for Claim event")
	}

	// 从 Log 中解析数据
	userAddress := common.BytesToAddress(log.Topics[1].Bytes())
	poolID := new(big.Int).SetBytes(log.Topics[2].Bytes())

	var reward *big.Int
	if len(log.Data) >= 32 {
		reward = new(big.Int).SetBytes(log.Data[0:32])
	}

	// TODO: 业务处理逻辑
	// 例如：更新 staking_user_positions 奖励记录
	// 例如：累加用户历史奖励总量

	var rewardFloat float64
	if reward != nil {
		rewardFloat, _ = new(big.Float).SetInt(reward).Float64()
	}

	logger.Logger.Info("Claim event processed",
		zap.String("user", userAddress.Hex()),
		zap.Int64("pool_id", poolID.Int64()),
		zap.Float64("reward", rewardFloat),
	)

	return nil
}
