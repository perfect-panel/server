package plugin

import (
	"sync"

	"google.golang.org/protobuf/types/known/structpb"
)

// EventHandler 是事件处理函数的签名
type EventHandler func(sub EventSubscription, event string, payload *structpb.Struct)

// EventBus 提供插件间的事件发布/订阅
type EventBus struct {
	mu          sync.RWMutex
	subscribers map[string][]EventSubscription // event -> subscriptions
	handler     EventHandler                   // 宿主级回调，用于调用 WASM
}

// NewEventBus 创建事件总线
func NewEventBus(handler EventHandler) *EventBus {
	return &EventBus{
		subscribers: make(map[string][]EventSubscription),
		handler:     handler,
	}
}

// Subscribe 订阅事件
func (eb *EventBus) Subscribe(sub EventSubscription) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	for _, existing := range eb.subscribers[sub.Event] {
		if existing.PluginName == sub.PluginName && existing.Handler == sub.Handler {
			return
		}
	}
	eb.subscribers[sub.Event] = append(eb.subscribers[sub.Event], sub)
}

// Unsubscribe 取消订阅
func (eb *EventBus) Unsubscribe(pluginName string) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	for event, subs := range eb.subscribers {
		remaining := subs[:0]
		for _, sub := range subs {
			if sub.PluginName != pluginName {
				remaining = append(remaining, sub)
			}
		}
		if len(remaining) == 0 {
			delete(eb.subscribers, event)
			continue
		}
		eb.subscribers[event] = remaining
	}
}

// Subscriptions 返回指定插件的事件订阅快照。pluginName 为空时返回全部订阅。
func (eb *EventBus) Subscriptions(pluginName string) []EventSubscription {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	result := make([]EventSubscription, 0)
	for _, subs := range eb.subscribers {
		for _, sub := range subs {
			if pluginName == "" || sub.PluginName == pluginName {
				result = append(result, sub)
			}
		}
	}
	return result
}

// Publish 发布事件给所有订阅者（异步分发）
func (eb *EventBus) Publish(event string, payload *structpb.Struct) {
	eb.mu.RLock()
	subs := append([]EventSubscription(nil), eb.subscribers[event]...)
	if event != "*" {
		subs = append(subs, eb.subscribers["*"]...)
	}
	handler := eb.handler
	eb.mu.RUnlock()

	for _, sub := range subs {
		if handler != nil {
			go handler(sub, event, payload)
		}
	}
}
