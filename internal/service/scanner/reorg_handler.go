package scanner

import (
	"context"
	"fmt"
	"math/big"

	"github.com/dijiacoder/staking-indexer/internal/logger"
	"github.com/dijiacoder/staking-indexer/internal/metrics"
	"github.com/dijiacoder/staking-indexer/internal/repository"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"
)

type ReorgHandler struct {
	repo   repository.ScannerRepository
	client *ethclient.Client
}

func NewReorgHandler(repo repository.ScannerRepository, client *ethclient.Client) *ReorgHandler {
	return &ReorgHandler{repo: repo, client: client}
}

// CheckAndHandleReorg checks if a reorg occurred and handles it if necessary.
// returns true if a reorg was handled, false otherwise.
func (h *ReorgHandler) CheckAndHandleReorg(ctx context.Context, chainID int64, contractAddress string,
	currentBlockNumber int64, currentParentHash string) (bool, error) {
	// 1. Get previous block from DB
	prevBlock, err := h.repo.GetBlockByNumber(ctx, chainID, currentBlockNumber-1)
	if err != nil {
		// If not found, we can't verify reorg, assume okay or it's the first block
		return false, nil
	}

	// 2. Compare DB's recorded hash for block N-1 with current block N's parent hash
	if prevBlock.BlockHash != currentParentHash {
		logger.Logger.Warn("Reorg detected",
			zap.Int64("block", currentBlockNumber),
			zap.Int64("prev_block", currentBlockNumber-1),
			zap.String("db_hash", prevBlock.BlockHash),
			zap.String("parent_hash", currentParentHash),
		)

		// 3. Find common ancestor by walking back
		commonAncestor, err := h.findCommonAncestor(ctx, chainID, currentBlockNumber-1)
		if err != nil {
			return false, fmt.Errorf("failed to find common ancestor: %w", err)
		}

		logger.Logger.Info("Found common ancestor, rolling back",
			zap.Int64("ancestor", commonAncestor),
		)

		// 记录回滚区块数
		rollbackBlocks := currentBlockNumber - 1 - commonAncestor
		if rollbackBlocks > 0 {
			labels := map[string]string{
				"chain_id":        fmt.Sprintf("%d", chainID),
				"contract_address": contractAddress,
			}
			metrics.ReorgRollbackBlocks.With(labels).Add(float64(rollbackBlocks))
		}

		// 4. Execute rollback in repository (atomic transaction)
		if err := h.repo.HandleReorg(ctx, chainID, contractAddress, commonAncestor); err != nil {
			return false, fmt.Errorf("failed to handle reorg rollback: %w", err)
		}

		return true, nil
	}

	return false, nil
}

func (h *ReorgHandler) findCommonAncestor(ctx context.Context, chainID int64, startBlock int64) (int64, error) {
	current := startBlock
	for current > 0 {
		dbBlock, err := h.repo.GetBlockByNumber(ctx, chainID, current)
		if err != nil {
			// If we can't find it in DB, we've gone back far enough or something is wrong
			return current, nil
		}

		header, err := h.client.HeaderByNumber(ctx, big.NewInt(current))
		if err != nil {
			return 0, err
		}

		if dbBlock.BlockHash == header.Hash().Hex() {
			return current, nil
		}
		current--
	}
	return 0, nil
}
