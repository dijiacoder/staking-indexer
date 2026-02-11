package handler

import (
	"github.com/ethereum/go-ethereum/core/types"
)

// EventHandlerContext 处理器上下文
type EventHandlerContext struct {
	// 原始日志，handler 自己解析组装数据
	Log types.Log
}

// EventHandler 事件处理器接口
// 职责：接收原始 Log，自己解析并执行业务处理逻辑
type EventHandler interface {
	CanHandle(eventName string) bool
	HandleEvent(ctx *EventHandlerContext) error
	GetEventName() string
}