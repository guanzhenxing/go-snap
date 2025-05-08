package config

import (
	"sync"
)

// ConfigChangeEvent 配置变更事件
type ConfigChangeEvent struct {
	// Key 变更的配置键
	Key string
	// OldValue 变更前的值
	OldValue interface{}
	// NewValue 变更后的值
	NewValue interface{}
}

// ConfigChangeListener 配置变更监听器
type ConfigChangeListener interface {
	// OnConfigChange 配置变更回调方法
	OnConfigChange(event ConfigChangeEvent)
}

// ConfigChangeCallback 配置变更回调函数类型
type ConfigChangeCallback func(event ConfigChangeEvent)

// ListenerFunc 将函数转换为ConfigChangeListener的适配器
type ListenerFunc func(event ConfigChangeEvent)

// OnConfigChange 实现ConfigChangeListener接口
func (f ListenerFunc) OnConfigChange(event ConfigChangeEvent) {
	f(event)
}

// Watcher 配置变更监听器
type Watcher struct {
	provider   Provider
	listeners  map[string][]ConfigChangeListener
	mu         sync.RWMutex
	keyCache   map[string]interface{}
	keyFilters []string
}

// NewWatcher 创建配置变更监听器
func NewWatcher(p Provider) *Watcher {
	watcher := &Watcher{
		provider:  p,
		listeners: make(map[string][]ConfigChangeListener),
		keyCache:  make(map[string]interface{}),
	}

	// 注册配置变更回调
	p.OnConfigChange(watcher.handleConfigChange)

	return watcher
}

// Watch 监听特定键的配置变更
func (w *Watcher) Watch(key string, listener ConfigChangeListener) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// 缓存当前值
	w.keyCache[key] = w.provider.Get(key)

	// 添加监听器
	if _, ok := w.listeners[key]; !ok {
		w.listeners[key] = make([]ConfigChangeListener, 0)
	}
	w.listeners[key] = append(w.listeners[key], listener)
}

// WatchFunc 使用函数监听特定键的配置变更
func (w *Watcher) WatchFunc(key string, callback ConfigChangeCallback) {
	w.Watch(key, ListenerFunc(callback))
}

// WatchKeys 设置要监听的键
func (w *Watcher) WatchKeys(keys ...string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.keyFilters = keys

	// 初始缓存值
	for _, key := range keys {
		w.keyCache[key] = w.provider.Get(key)
	}
}

// Unwatch 取消对特定键的监听
func (w *Watcher) Unwatch(key string, listener ConfigChangeListener) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if listeners, ok := w.listeners[key]; ok {
		// 查找并移除监听器
		for i, l := range listeners {
			if l == listener {
				w.listeners[key] = append(listeners[:i], listeners[i+1:]...)
				break
			}
		}

		// 如果键没有监听器了，移除缓存
		if len(w.listeners[key]) == 0 {
			delete(w.listeners, key)
			delete(w.keyCache, key)
		}
	}
}

// UnwatchAll 取消所有监听
func (w *Watcher) UnwatchAll() {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.listeners = make(map[string][]ConfigChangeListener)
	w.keyCache = make(map[string]interface{})
	w.keyFilters = nil
}

// handleConfigChange 处理配置变更
func (w *Watcher) handleConfigChange() {
	w.mu.Lock()
	// 获取所有监听的键
	keys := make([]string, 0, len(w.keyCache))
	for k := range w.keyCache {
		keys = append(keys, k)
	}

	// 如果有键过滤器，只检查这些键
	if len(w.keyFilters) > 0 {
		filterMap := make(map[string]struct{}, len(w.keyFilters))
		for _, k := range w.keyFilters {
			filterMap[k] = struct{}{}
		}

		filteredKeys := make([]string, 0, len(w.keyFilters))
		for _, k := range keys {
			if _, ok := filterMap[k]; ok {
				filteredKeys = append(filteredKeys, k)
			}
		}
		keys = filteredKeys
	}

	// 检查每个键的变化
	changedKeys := make(map[string]ConfigChangeEvent)
	for _, key := range keys {
		oldValue := w.keyCache[key]
		newValue := w.provider.Get(key)

		// 比较值是否变化
		if !equals(oldValue, newValue) {
			changedKeys[key] = ConfigChangeEvent{
				Key:      key,
				OldValue: oldValue,
				NewValue: newValue,
			}
			// 更新缓存
			w.keyCache[key] = newValue
		}
	}
	w.mu.Unlock()

	// 通知监听器
	w.notifyListeners(changedKeys)
}

// notifyListeners 通知配置变更
func (w *Watcher) notifyListeners(changedKeys map[string]ConfigChangeEvent) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	for key, event := range changedKeys {
		if listeners, ok := w.listeners[key]; ok {
			// 复制监听器列表，避免在回调中修改
			listenersCopy := make([]ConfigChangeListener, len(listeners))
			copy(listenersCopy, listeners)

			for _, listener := range listenersCopy {
				listener.OnConfigChange(event)
			}
		}
	}
}

// equals 比较两个值是否相等
func equals(a, b interface{}) bool {
	// 简单相等检查
	if a == b {
		return true
	}

	// 如果类型不同，不相等
	if a == nil || b == nil {
		return false
	}

	// TODO: 实现更复杂的比较逻辑，如切片、映射等
	// 这里简单地使用!=，可能对复杂类型不够准确
	return a == b
}
