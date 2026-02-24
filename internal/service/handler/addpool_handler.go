package handler

import (
	"math/big"

	"github.com/dijiacoder/staking-indexer/internal/gen/model"
	"github.com/dijiacoder/staking-indexer/internal/logger"
	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
)

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

	logger.Logger.Info("AddPool event processed and saved",
		zap.Int64("poolID", poolID.Int64()),
		zap.String("stTokenAddress", stTokenAddress.Hex()),
		zap.Int64("poolWeight", poolWeight.Int64()),
		zap.Int64("lastRewardBlock", bigInt64(lastRewardBlock)),
		zap.Int64("minDepositAmount", bigInt64(minDepositAmount)),
		zap.Int64("unstakeLockedBlocks", bigInt64(unstakeLockedBlocks)),
	)

	pool := &model.StakingPool{
		ChainID:             ctx.ChainID,
		ContractAddress:     ctx.ContractAddress,
		PoolID:              poolID.Int64(),
		StTokenAddress:      stTokenAddress.Hex(),
		PoolWeight:          poolWeight.Int64(),
		LastRewardBlock:     bigInt64(lastRewardBlock),
		AccZeroTokenPerSt:   0,
		StTokenAmount:       0,
		MinDepositAmount:    bigInt64(minDepositAmount),
		UnstakeLockedBlocks: bigInt64(unstakeLockedBlocks),
	}

	if err := ctx.Repo.SavePool(ctx.Ctx, pool); err != nil {
		logger.Logger.Error("save AddPool to staking_pools failed",
			zap.Error(err),
			zap.Int64("poolID", poolID.Int64()),
			zap.String("stTokenAddress", stTokenAddress.Hex()),
		)
		return err
	}

	return nil
}

// bigInt64 安全转换 big.Int 为 int64，nil 时返回 0
func bigInt64(n *big.Int) int64 {
	if n == nil {
		return 0
	}
	return n.Int64()
}
