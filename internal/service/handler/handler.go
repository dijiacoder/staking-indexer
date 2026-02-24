package handler

import (
	"context"

	"github.com/dijiacoder/staking-indexer/internal/repository"
	"github.com/ethereum/go-ethereum/core/types"
)

// EventHandlerContext 处理器上下文
type EventHandlerContext struct {
	Log             types.Log
	ChainID         int64
	ContractAddress string
	Repo            repository.ScannerRepository
	Ctx             context.Context
}

// EventHandler 事件处理器接口
// 职责：接收原始 Log，自己解析并执行业务处理逻辑
type EventHandler interface {
	CanHandle(eventName string) bool
	HandleEvent(ctx *EventHandlerContext) error
	GetEventName() string
}
