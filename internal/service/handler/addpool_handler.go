package handler

import (
	"math/big"

	"github.com/dijiacoder/staking-indexer/internal/gen/model"
	"github.com/dijiacoder/staking-indexer/internal/logger"
	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
)

// AddPoolEventHandler addPool事件处理器
type AddPoolEventHandler struct {
	BaseEventHandler
}

func NewAddPoolEventHandler() *AddPoolEventHandler {
	return &AddPoolEventHandler{
		BaseEventHandler: BaseEventHandler{eventName: "AddPool"},
	}
}

func (h *AddPoolEventHandler) HandleEvent(ctx *EventHandlerContext) error {
	log := ctx.Log

	// 从 Log 中解析数据
	poolID := new(big.Int).SetBytes(log.Topics[1].Bytes())
	stTokenAddress := common.BytesToAddress(log.Topics[2].Bytes())
	poolWeight := new(big.Int).SetBytes(log.Topics[3].Bytes())

	var lastRewardBlock, minDepositAmount, unstakeLockedBlocks *big.Int

	if len(log.Data) >= 32 {
		lastRewardBlock = new(big.Int).SetBytes(log.Data[0:32])
	}
	if len(log.Data) >= 64 {
		minDepositAmount = new(big.Int).SetBytes(log.Data[32:64])
	}
	if len(log.Data) >= 96 {
		unstakeLockedBlocks = new(big.Int).SetBytes(log.Data[64:96])
	}

	logger.Logger.Info("AddPool event processed",
		zap.Int64("poolID", poolID.Int64()),
		zap.String("stTokenAddress", stTokenAddress.Hex()),
		zap.Int64("poolWeight", poolWeight.Int64()),
		zap.String("lastRewardBlock", bigIntToString(lastRewardBlock)),
		zap.String("minDepositAmount", bigIntToString(minDepositAmount)),
		zap.Int64("unstakeLockedBlocks", bigInt64(unstakeLockedBlocks)),
	)

	pool := &model.StakingPool{
		ChainID:         ctx.ChainID,
		ContractAddress: ctx.ContractAddress,
		PoolID:          poolID.Int64(),
		StakeToken:      stTokenAddress.Hex(),
		RewardToken:     "0x0", // AddPool 事件不提供，需要从配置或其他方式获取
		StartBlock:      lastRewardBlock.Int64(),
		EndBlock:        0, // AddPool 事件不提供
		RewardPerBlock:  0, // AddPool 事件不提供
	}

	return nil
}

// bigIntToString 安全转换 big.Int 为字符串，nil 时返回 "0"
func bigIntToString(n *big.Int) string {
	if n == nil {
		return "0"
	}
	return n.String()
}

// bigInt64 安全转换 big.Int 为 int64，nil 时返回 0
func bigInt64(n *big.Int) int64 {
	if n == nil {
		return 0
	}
	return n.Int64()
}
