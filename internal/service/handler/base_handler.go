package handler

// BaseEventHandler 基础事件处理器，提供通用功能
type BaseEventHandler struct {
	eventName string
}

func (h *BaseEventHandler) GetEventName() string {
	return h.eventName
}

func (h *BaseEventHandler) CanHandle(eventName string) bool {
	return h.eventName == eventName
}