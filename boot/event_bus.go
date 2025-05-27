package boot

import (
	"sync"
)

// EventBus 事件总线
type EventBus struct {
	listeners map[string][]EventListener
	mutex     sync.RWMutex
}

// NewEventBus 创建事件总线
func NewEventBus() *EventBus {
	return &EventBus{
		listeners: make(map[string][]EventListener),
	}
}

// Subscribe 订阅事件
func (b *EventBus) Subscribe(eventName string, listener EventListener) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if _, exists := b.listeners[eventName]; !exists {
		b.listeners[eventName] = []EventListener{}
	}

	b.listeners[eventName] = append(b.listeners[eventName], listener)
}

// Unsubscribe 取消订阅事件
func (b *EventBus) Unsubscribe(eventName string, listener EventListener) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if listeners, exists := b.listeners[eventName]; exists {
		for i, l := range listeners {
			if &l == &listener {
				b.listeners[eventName] = append(listeners[:i], listeners[i+1:]...)
				break
			}
		}
	}
}

// Publish 发布事件
func (b *EventBus) Publish(eventName string, eventData interface{}) {
	b.mutex.RLock()
	listeners, exists := b.listeners[eventName]
	b.mutex.RUnlock()

	if exists {
		for _, listener := range listeners {
			go listener(eventName, eventData)
		}
	}
}

// PublishSync 同步发布事件
func (b *EventBus) PublishSync(eventName string, eventData interface{}) {
	b.mutex.RLock()
	listeners, exists := b.listeners[eventName]
	b.mutex.RUnlock()

	if exists {
		for _, listener := range listeners {
			listener(eventName, eventData)
		}
	}
}

// HasListeners 是否有事件监听器
func (b *EventBus) HasListeners(eventName string) bool {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	listeners, exists := b.listeners[eventName]
	return exists && len(listeners) > 0
}

// Clear 清除所有事件监听器
func (b *EventBus) Clear() {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.listeners = make(map[string][]EventListener)
}
