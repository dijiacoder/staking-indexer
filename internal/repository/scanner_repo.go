package repository

import (
	"context"

	"github.com/dijiacoder/staking-indexer/internal/gen/model"
	"github.com/dijiacoder/staking-indexer/internal/gen/query"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ScannerRepository interface {
	UpdateCursor(ctx context.Context, chainID int64, contractAddress string, lastScanned int64, lastConfirmed int64) error

	GetBlockByNumber(ctx context.Context, chainID int64, blockNumber int64) (*model.ChainBlock, error)
	SaveBlock(ctx context.Context, block *model.ChainBlock) error

	SaveEventsAndProcessPositions(ctx context.Context, events []*model.StakingEvent) error

	HandleReorg(ctx context.Context, chainID int64, contractAddress string, rollbackToBlock int64) error
	GetCursor(ctx context.Context, chainID int64, contractAddress string) (*model.ChainScanCursor, error)

	SavePool(ctx context.Context, pool *model.StakingPool) error
}

type scannerRepository struct {
	db *gorm.DB
	q  *query.Query
}

func NewScannerRepository(db *gorm.DB) ScannerRepository {
	return &scannerRepository{
		db: db,
		q:  query.Use(db),
	}
}

func (r *scannerRepository) SavePool(ctx context.Context, pool *model.StakingPool) error {
	return r.q.StakingPool.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "chain_id"}, {Name: "contract_address"}, {Name: "pool_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"stake_token", "reward_token", "start_block", "end_block", "reward_per_block"}),
	}).Create(pool)
}

func (r *scannerRepository) GetCursor(ctx context.Context, chainID int64, contractAddress string) (*model.ChainScanCursor, error) {
	return r.q.ChainScanCursor.WithContext(ctx).Where(
		r.q.ChainScanCursor.ChainID.Eq(chainID),
		r.q.ChainScanCursor.ContractAddress.Eq(contractAddress),
	).First()
}

func (r *scannerRepository) UpdateCursor(ctx context.Context, chainID int64, contractAddress string, lastScanned int64, lastConfirmed int64) error {
	_, err := r.q.ChainScanCursor.WithContext(ctx).Where(
		r.q.ChainScanCursor.ChainID.Eq(chainID),
		r.q.ChainScanCursor.ContractAddress.Eq(contractAddress),
	).Updates(&model.ChainScanCursor{
		LastScannedBlock:   lastScanned,
		LastConfirmedBlock: lastConfirmed,
	})
	return err
}

func (r *scannerRepository) GetBlockByNumber(ctx context.Context, chainID int64, blockNumber int64) (*model.ChainBlock, error) {
	return r.q.ChainBlock.WithContext(ctx).Where(
		r.q.ChainBlock.ChainID.Eq(chainID),
		r.q.ChainBlock.BlockNumber.Eq(blockNumber),
	).First()
}

func (r *scannerRepository) SaveBlock(ctx context.Context, block *model.ChainBlock) error {
	return r.q.ChainBlock.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "chain_id"}, {Name: "block_number"}},
		DoUpdates: clause.AssignmentColumns([]string{"block_hash", "parent_hash", "is_confirmed"}),
	}).Create(block)
}

func (r *scannerRepository) SaveEventsAndProcessPositions(ctx context.Context, events []*model.StakingEvent) error {
	if len(events) == 0 {
		return nil
	}

	return r.q.Transaction(func(tx *query.Query) error {
		// 1. Save Events
		if err := tx.StakingEvent.WithContext(ctx).Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "tx_hash"}, {Name: "log_index"}},
			DoUpdates: clause.AssignmentColumns([]string{"amount", "event_type", "block_number"}),
		}).Create(events...); err != nil {
			return err
		}

		// 2. Update Positions
		for _, ev := range events {
			pos, err := tx.StakingUserPosition.WithContext(ctx).Where(
				tx.StakingUserPosition.ChainID.Eq(ev.ChainID),
				tx.StakingUserPosition.ContractAddress.Eq(ev.ContractAddress),
				tx.StakingUserPosition.PoolID.Eq(ev.PoolID),
				tx.StakingUserPosition.UserAddress.Eq(ev.UserAddress),
			).FirstOrInit()
			if err != nil {
				return err
			}

			// Initialize fields if new
			if pos.StakedAmount == nil {
				zero := 0.0
				pos.StakedAmount = &zero
			}
			if pos.RewardDebt == nil {
				zero := 0.0
				pos.RewardDebt = &zero
			}
			pos.ChainID = ev.ChainID
			pos.ContractAddress = ev.ContractAddress
			pos.PoolID = ev.PoolID
			pos.UserAddress = ev.UserAddress

			switch ev.EventType {
			case "Deposit":
				*pos.StakedAmount += ev.Amount
			case "Withdraw":
				*pos.StakedAmount -= ev.Amount
			case "Claim":
				// Handle claim logic if needed, for now just update reward debt if it was relevant
				// Usually Claim doesn't change StakedAmount
			}

			if err := tx.StakingUserPosition.WithContext(ctx).Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "chain_id"}, {Name: "contract_address"}, {Name: "pool_id"}, {Name: "user_address"}},
				DoUpdates: clause.AssignmentColumns([]string{"staked_amount", "reward_debt", "updated_at"}),
			}).Create(pos); err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *scannerRepository) HandleReorg(ctx context.Context, chainID int64, contractAddress string, rollbackToBlock int64) error {
	return r.q.Transaction(func(tx *query.Query) error {
		// 1. Find events to rollback
		events, err := tx.StakingEvent.WithContext(ctx).Where(
			tx.StakingEvent.ChainID.Eq(chainID),
			tx.StakingEvent.ContractAddress.Eq(contractAddress),
			tx.StakingEvent.BlockNumber.Gt(rollbackToBlock),
		).Find()
		if err != nil {
			return err
		}

		// 2. Reverse positions
		for _, ev := range events {
			pos, err := tx.StakingUserPosition.WithContext(ctx).Where(
				tx.StakingUserPosition.ChainID.Eq(ev.ChainID),
				tx.StakingUserPosition.ContractAddress.Eq(ev.ContractAddress),
				tx.StakingUserPosition.PoolID.Eq(ev.PoolID),
				tx.StakingUserPosition.UserAddress.Eq(ev.UserAddress),
			).First()
			if err != nil {
				continue // If position not found, maybe it was already corrected?
			}

			switch ev.EventType {
			case "Deposit":
				*pos.StakedAmount -= ev.Amount
			case "Withdraw":
				*pos.StakedAmount += ev.Amount
			}

			if err := tx.StakingUserPosition.WithContext(ctx).Save(pos); err != nil {
				return err
			}
		}

		// 3. Delete events
		if _, err := tx.StakingEvent.WithContext(ctx).Where(
			tx.StakingEvent.ChainID.Eq(chainID),
			tx.StakingEvent.ContractAddress.Eq(contractAddress),
			tx.StakingEvent.BlockNumber.Gt(rollbackToBlock),
		).Delete(); err != nil {
			return err
		}

		// 4. Mark blocks as non-canonical (IsConfirmed = 0)
		if _, err := tx.ChainBlock.WithContext(ctx).Where(
			tx.ChainBlock.ChainID.Eq(chainID),
			tx.ChainBlock.BlockNumber.Gt(rollbackToBlock),
		).Update(tx.ChainBlock.IsConfirmed, 0); err != nil {
			return err
		}

		// 5. Update cursor
		if _, err := tx.ChainScanCursor.WithContext(ctx).Where(
			tx.ChainScanCursor.ChainID.Eq(chainID),
			tx.ChainScanCursor.ContractAddress.Eq(contractAddress),
		).Update(tx.ChainScanCursor.LastScannedBlock, rollbackToBlock); err != nil {
			return err
		}

		return nil
	})
}
