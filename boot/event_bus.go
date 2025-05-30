package boot

import (
	"sync"
)

// EventBus 事件总线，实现发布-订阅模式，用于组件间的解耦通信
// 允许组件发布事件并让其他组件订阅这些事件，无需直接依赖
type EventBus struct {
	// listeners 存储事件名称到监听器列表的映射
	listeners map[string][]EventListener
	// mutex 用于保护并发访问listeners映射
	mutex sync.RWMutex
}

// NewEventBus 创建并初始化一个新的事件总线实例
// 返回：
//
//	初始化的EventBus实例
func NewEventBus() *EventBus {
	return &EventBus{
		listeners: make(map[string][]EventListener),
	}
}

// Subscribe 订阅指定名称的事件
// 参数：
//
//	eventName: 要订阅的事件名称
//	listener: 当事件发生时调用的监听器函数
func (b *EventBus) Subscribe(eventName string, listener EventListener) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if _, exists := b.listeners[eventName]; !exists {
		b.listeners[eventName] = []EventListener{}
	}

	b.listeners[eventName] = append(b.listeners[eventName], listener)
}

// Unsubscribe 取消订阅指定名称的事件
// 参数：
//
//	eventName: 要取消订阅的事件名称
//	listener: 要移除的监听器函数
//
// 注意：
//
//	由于Go语言的函数比较限制，此方法可能无法正确移除所有情况下的监听器
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

// Publish 异步发布事件，不阻塞调用者
// 对每个监听器在独立的goroutine中调用，适合长时间运行的处理
// 参数：
//
//	eventName: 要发布的事件名称
//	eventData: 与事件一起传递的数据
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

// PublishSync 同步发布事件，阻塞直到所有监听器处理完成
// 在当前goroutine中顺序调用所有监听器，适合简短的处理
// 参数：
//
//	eventName: 要发布的事件名称
//	eventData: 与事件一起传递的数据
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

// HasListeners 检查指定事件是否有监听器
// 参数：
//
//	eventName: 要检查的事件名称
//
// 返回：
//
//	如果事件有至少一个监听器则返回true，否则返回false
func (b *EventBus) HasListeners(eventName string) bool {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	listeners, exists := b.listeners[eventName]
	return exists && len(listeners) > 0
}

// Clear 清除所有事件监听器
// 通常在应用关闭或需要重置事件总线时调用
func (b *EventBus) Clear() {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.listeners = make(map[string][]EventListener)
}
