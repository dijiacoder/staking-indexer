package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// 区块高度指标
	CurrentScannedBlock = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "staking_indexer_current_scanned_block",
			Help: "当前扫描区块高度",
		},
		[]string{"chain_id", "contract_address"},
	)

	ChainLatestBlock = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "staking_indexer_chain_latest_block",
			Help: "最新链上区块高度",
		},
		[]string{"chain_id", "contract_address"},
	)

	SafeBlock = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "staking_indexer_safe_block",
			Help: "Safe区块高度（考虑确认数）",
		},
		[]string{"chain_id", "contract_address"},
	)

	SyncLag = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "staking_indexer_sync_lag",
			Help: "同步延迟（safe_block - current_scanned_block）",
		},
		[]string{"chain_id", "contract_address"},
	)

	// 性能指标
	BlocksPerSecond = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "staking_indexer_blocks_per_second",
			Help: "每秒处理区块数",
		},
		[]string{"chain_id", "contract_address"},
	)

	// RPC 指标
	RPCRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "staking_indexer_rpc_requests_total",
			Help: "RPC请求总数",
		},
		[]string{"chain_id", "method"},
	)

	RPCErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "staking_indexer_rpc_errors_total",
			Help: "RPC错误数",
		},
		[]string{"chain_id", "method"},
	)

	RPCDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "staking_indexer_rpc_duration_seconds",
			Help:    "RPC请求耗时",
			Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 2, 5},
		},
		[]string{"chain_id", "method"},
	)

	// 事件处理指标
	EventsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "staking_indexer_events_total",
			Help: "解析到的事件总数",
		},
		[]string{"chain_id", "contract_address", "event_type"},
	)

	EventsFailedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "staking_indexer_events_failed_total",
			Help: "处理失败事件数",
		},
		[]string{"chain_id", "contract_address", "event_type"},
	)

	// Reorg 指标
	ReorgTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "staking_indexer_reorg_total",
			Help: "Reorg次数",
		},
		[]string{"chain_id", "contract_address"},
	)

	ReorgRollbackBlocks = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "staking_indexer_reorg_rollback_blocks_total",
			Help: "Reorg回滚区块数累计",
		},
		[]string{"chain_id", "contract_address"},
	)

	LastReorgBlock = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "staking_indexer_last_reorg_block",
			Help: "最近一次Reorg区块号",
		},
		[]string{"chain_id", "contract_address"},
	)
)
