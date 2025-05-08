package config

import (
	"sync"
	"testing"
)

// 测试配置变更事件监听器
type testListener struct {
	events []ConfigChangeEvent
	mu     sync.Mutex
}

func newTestListener() *testListener {
	return &testListener{
		events: make([]ConfigChangeEvent, 0),
	}
}

func (l *testListener) OnConfigChange(event ConfigChangeEvent) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.events = append(l.events, event)
}

func (l *testListener) GetEvents() []ConfigChangeEvent {
	l.mu.Lock()
	defer l.mu.Unlock()
	// 复制事件以避免并发问题
	result := make([]ConfigChangeEvent, len(l.events))
	copy(result, l.events)
	return result
}

func (l *testListener) Reset() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.events = make([]ConfigChangeEvent, 0)
}

// 测试包装函数适配器
func TestListenerFunc(t *testing.T) {
	var receivedEvent ConfigChangeEvent

	// 创建适配器函数
	listener := ListenerFunc(func(event ConfigChangeEvent) {
		receivedEvent = event
	})

	// 创建测试事件
	testEvent := ConfigChangeEvent{
		Key:      "test.key",
		OldValue: "old-value",
		NewValue: "new-value",
	}

	// 调用适配器
	listener.OnConfigChange(testEvent)

	// 验证事件是否正确传递
	if receivedEvent.Key != testEvent.Key {
		t.Errorf("Expected event key '%s', got '%s'", testEvent.Key, receivedEvent.Key)
	}

	if receivedEvent.OldValue != testEvent.OldValue {
		t.Errorf("Expected old value '%v', got '%v'", testEvent.OldValue, receivedEvent.OldValue)
	}

	if receivedEvent.NewValue != testEvent.NewValue {
		t.Errorf("Expected new value '%v', got '%v'", testEvent.NewValue, receivedEvent.NewValue)
	}
}

// 测试监听器创建
func TestNewWatcher(t *testing.T) {
	// 创建模拟提供者
	provider := newMockProvider()

	// 创建监听器
	watcher := NewWatcher(provider)

	if watcher == nil {
		t.Fatal("Expected watcher to be initialized, got nil")
	}

	if watcher.provider != provider {
		t.Error("Expected watcher to use provided provider")
	}

	if watcher.listeners == nil {
		t.Error("Expected listeners map to be initialized")
	}

	if watcher.keyCache == nil {
		t.Error("Expected key cache to be initialized")
	}
}

// 测试监听特定键的配置变更
func TestWatcher_Watch(t *testing.T) {
	// 创建模拟提供者
	provider := newMockProvider()
	provider.Set("test.key", "initial-value")

	// 创建监听器
	watcher := NewWatcher(provider)

	// 创建测试监听器
	listener := newTestListener()

	// 监听键
	watcher.Watch("test.key", listener)

	// 验证监听器是否正确注册
	if len(watcher.listeners["test.key"]) != 1 {
		t.Errorf("Expected 1 listener for 'test.key', got %d", len(watcher.listeners["test.key"]))
	}

	// 验证键值是否正确缓存
	if cachedValue, ok := watcher.keyCache["test.key"]; !ok || cachedValue != "initial-value" {
		t.Errorf("Expected 'test.key' to be cached with value 'initial-value', got '%v'", cachedValue)
	}

	// 模拟配置变更
	provider.Set("test.key", "new-value")
	watcher.handleConfigChange()

	// 验证监听器是否收到事件
	events := listener.GetEvents()
	if len(events) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(events))
	}

	event := events[0]
	if event.Key != "test.key" {
		t.Errorf("Expected event key 'test.key', got '%s'", event.Key)
	}

	if event.OldValue != "initial-value" {
		t.Errorf("Expected old value 'initial-value', got '%v'", event.OldValue)
	}

	if event.NewValue != "new-value" {
		t.Errorf("Expected new value 'new-value', got '%v'", event.NewValue)
	}
}

// 测试使用函数监听特定键的配置变更
func TestWatcher_WatchFunc(t *testing.T) {
	// 创建模拟提供者
	provider := newMockProvider()
	provider.Set("test.key", "initial-value")

	// 创建监听器
	watcher := NewWatcher(provider)

	// 跟踪事件
	var receivedEvent ConfigChangeEvent
	var callCount int

	// 使用函数监听
	watcher.WatchFunc("test.key", func(event ConfigChangeEvent) {
		receivedEvent = event
		callCount++
	})

	// 验证监听器是否正确注册
	if len(watcher.listeners["test.key"]) != 1 {
		t.Errorf("Expected 1 listener for 'test.key', got %d", len(watcher.listeners["test.key"]))
	}

	// 模拟配置变更
	provider.Set("test.key", "new-value")
	watcher.handleConfigChange()

	// 验证回调是否被调用
	if callCount != 1 {
		t.Errorf("Expected callback to be called once, got %d", callCount)
	}

	// 验证事件是否正确传递
	if receivedEvent.Key != "test.key" {
		t.Errorf("Expected event key 'test.key', got '%s'", receivedEvent.Key)
	}

	if receivedEvent.OldValue != "initial-value" {
		t.Errorf("Expected old value 'initial-value', got '%v'", receivedEvent.OldValue)
	}

	if receivedEvent.NewValue != "new-value" {
		t.Errorf("Expected new value 'new-value', got '%v'", receivedEvent.NewValue)
	}
}

// 测试设置要监听的键
func TestWatcher_WatchKeys(t *testing.T) {
	// 创建模拟提供者
	provider := newMockProvider()
	provider.Set("key1", "value1")
	provider.Set("key2", "value2")

	// 创建监听器
	watcher := NewWatcher(provider)

	// 设置要监听的键
	watcher.WatchKeys("key1", "key2")

	// 验证键过滤器是否正确设置
	if len(watcher.keyFilters) != 2 {
		t.Errorf("Expected 2 key filters, got %d", len(watcher.keyFilters))
	}

	// 验证键值是否正确缓存
	if cachedValue, ok := watcher.keyCache["key1"]; !ok || cachedValue != "value1" {
		t.Errorf("Expected 'key1' to be cached with value 'value1', got '%v'", cachedValue)
	}

	if cachedValue, ok := watcher.keyCache["key2"]; !ok || cachedValue != "value2" {
		t.Errorf("Expected 'key2' to be cached with value 'value2', got '%v'", cachedValue)
	}

	// 创建测试监听器
	listener1 := newTestListener()
	listener2 := newTestListener()

	// 监听键
	watcher.Watch("key1", listener1)
	watcher.Watch("key2", listener2)

	// 模拟只有key1变更
	provider.Set("key1", "new-value1")
	watcher.handleConfigChange()

	// 验证只有listener1收到事件
	events1 := listener1.GetEvents()
	if len(events1) != 1 {
		t.Errorf("Expected listener1 to receive 1 event, got %d", len(events1))
	}

	events2 := listener2.GetEvents()
	if len(events2) != 0 {
		t.Errorf("Expected listener2 to receive 0 events, got %d", len(events2))
	}
}

// 测试取消监听
func TestWatcher_Unwatch(t *testing.T) {
	// 创建模拟提供者
	provider := newMockProvider()
	provider.Set("test.key", "initial-value")

	// 创建监听器和测试监听器
	watcher := NewWatcher(provider)
	listener := newTestListener()

	// 监听键
	watcher.Watch("test.key", listener)

	// 验证监听器是否正确注册
	if len(watcher.listeners["test.key"]) != 1 {
		t.Errorf("Expected 1 listener for 'test.key', got %d", len(watcher.listeners["test.key"]))
	}

	// 取消监听
	watcher.Unwatch("test.key", listener)

	// 验证监听器是否被移除
	if listeners, ok := watcher.listeners["test.key"]; ok && len(listeners) > 0 {
		t.Errorf("Expected 'test.key' to have no listeners, got %d", len(listeners))
	}

	// 验证键缓存是否被清理
	if _, ok := watcher.keyCache["test.key"]; ok {
		t.Error("Expected 'test.key' to be removed from key cache")
	}

	// 测试取消不存在的监听器
	anotherListener := newTestListener()
	watcher.Unwatch("nonexistent.key", anotherListener)
	// 不应该抛出异常
}

// 测试取消所有监听
func TestWatcher_UnwatchAll(t *testing.T) {
	// 创建模拟提供者
	provider := newMockProvider()

	// 设置一些测试值
	provider.Set("key1", "value1")
	provider.Set("key2", "value2")

	// 创建监听器
	watcher := NewWatcher(provider)

	// 注册多个监听器
	listener1 := newTestListener()
	listener2 := newTestListener()

	watcher.Watch("key1", listener1)
	watcher.Watch("key2", listener2)
	watcher.WatchKeys("key1", "key2", "key3")

	// 取消所有监听
	watcher.UnwatchAll()

	// 验证所有监听器是否被清理
	if len(watcher.listeners) > 0 {
		t.Errorf("Expected no listeners after UnwatchAll, got %d", len(watcher.listeners))
	}

	// 验证键缓存是否被清理
	if len(watcher.keyCache) > 0 {
		t.Errorf("Expected empty key cache after UnwatchAll, got %d items", len(watcher.keyCache))
	}

	// 验证键过滤器是否被清理
	if watcher.keyFilters != nil {
		t.Errorf("Expected keyFilters to be nil after UnwatchAll, got %v", watcher.keyFilters)
	}
}

// 测试配置变更处理
func TestWatcher_handleConfigChange(t *testing.T) {
	// 创建模拟提供者
	provider := newMockProvider()

	// 设置初始值
	provider.Set("key1", "value1")
	provider.Set("key2", "value2")
	provider.Set("key3", "value3")

	// 创建监听器
	watcher := NewWatcher(provider)

	// 创建测试监听器
	listener1 := newTestListener()
	listener2 := newTestListener()
	listener3 := newTestListener()

	// 监听键
	watcher.Watch("key1", listener1)
	watcher.Watch("key2", listener2)
	watcher.Watch("key3", listener3)

	// 模拟多个配置变更
	provider.Set("key1", "new-value1")
	provider.Set("key2", "new-value2")
	// key3保持不变

	// 处理变更
	watcher.handleConfigChange()

	// 验证只有变更的配置触发了事件
	events1 := listener1.GetEvents()
	if len(events1) != 1 {
		t.Fatalf("Expected listener1 to receive 1 event, got %d", len(events1))
	}

	events2 := listener2.GetEvents()
	if len(events2) != 1 {
		t.Fatalf("Expected listener2 to receive 1 event, got %d", len(events2))
	}

	events3 := listener3.GetEvents()
	if len(events3) != 0 {
		t.Fatalf("Expected listener3 to receive 0 events, got %d", len(events3))
	}

	// 验证事件内容
	if events1[0].NewValue != "new-value1" {
		t.Errorf("Expected new value 'new-value1', got '%v'", events1[0].NewValue)
	}

	if events2[0].NewValue != "new-value2" {
		t.Errorf("Expected new value 'new-value2', got '%v'", events2[0].NewValue)
	}

	// 验证键缓存是否已更新
	if watcher.keyCache["key1"] != "new-value1" {
		t.Errorf("Expected key cache for 'key1' to be updated to 'new-value1', got '%v'", watcher.keyCache["key1"])
	}

	if watcher.keyCache["key2"] != "new-value2" {
		t.Errorf("Expected key cache for 'key2' to be updated to 'new-value2', got '%v'", watcher.keyCache["key2"])
	}
}

// 测试多个监听器监听同一个键
func TestWatcher_MultipleListenersPerKey(t *testing.T) {
	// 创建模拟提供者
	provider := newMockProvider()
	provider.Set("shared.key", "initial-value")

	// 创建监听器
	watcher := NewWatcher(provider)

	// 创建多个测试监听器
	listener1 := newTestListener()
	listener2 := newTestListener()
	listener3 := newTestListener()

	// 监听同一个键
	watcher.Watch("shared.key", listener1)
	watcher.Watch("shared.key", listener2)
	watcher.Watch("shared.key", listener3)

	// 验证监听器是否都正确注册
	if len(watcher.listeners["shared.key"]) != 3 {
		t.Errorf("Expected 3 listeners for 'shared.key', got %d", len(watcher.listeners["shared.key"]))
	}

	// 模拟配置变更
	provider.Set("shared.key", "new-value")
	watcher.handleConfigChange()

	// 验证所有监听器都收到事件
	for i, listener := range []*testListener{listener1, listener2, listener3} {
		events := listener.GetEvents()
		if len(events) != 1 {
			t.Errorf("Expected listener%d to receive 1 event, got %d", i+1, len(events))
			continue
		}

		event := events[0]
		if event.Key != "shared.key" {
			t.Errorf("Expected event key 'shared.key', got '%s'", event.Key)
		}

		if event.OldValue != "initial-value" {
			t.Errorf("Expected old value 'initial-value', got '%v'", event.OldValue)
		}

		if event.NewValue != "new-value" {
			t.Errorf("Expected new value 'new-value', got '%v'", event.NewValue)
		}
	}

	// 测试取消其中一个监听器
	watcher.Unwatch("shared.key", listener2)

	// 验证其他监听器仍然存在
	if len(watcher.listeners["shared.key"]) != 2 {
		t.Errorf("Expected 2 listeners for 'shared.key' after unwatching one, got %d", len(watcher.listeners["shared.key"]))
	}

	// 重置事件
	listener1.Reset()
	listener2.Reset()
	listener3.Reset()

	// 再次模拟配置变更
	provider.Set("shared.key", "newer-value")
	watcher.handleConfigChange()

	// 验证只有剩余的监听器收到事件
	if len(listener1.GetEvents()) != 1 {
		t.Errorf("Expected listener1 to receive 1 event, got %d", len(listener1.GetEvents()))
	}

	if len(listener2.GetEvents()) != 0 {
		t.Errorf("Expected listener2 to receive 0 events after unwatching, got %d", len(listener2.GetEvents()))
	}

	if len(listener3.GetEvents()) != 1 {
		t.Errorf("Expected listener3 to receive 1 event, got %d", len(listener3.GetEvents()))
	}
}

// 测试值相等比较函数
func TestEquals(t *testing.T) {
	tests := []struct {
		name     string
		a        interface{}
		b        interface{}
		expected bool
	}{
		{
			name:     "Equal strings",
			a:        "test",
			b:        "test",
			expected: true,
		},
		{
			name:     "Different strings",
			a:        "test",
			b:        "different",
			expected: false,
		},
		{
			name:     "Equal integers",
			a:        42,
			b:        42,
			expected: true,
		},
		{
			name:     "Different integers",
			a:        42,
			b:        43,
			expected: false,
		},
		{
			name:     "Equal booleans",
			a:        true,
			b:        true,
			expected: true,
		},
		{
			name:     "Different booleans",
			a:        true,
			b:        false,
			expected: false,
		},
		{
			name:     "First nil",
			a:        nil,
			b:        "not nil",
			expected: false,
		},
		{
			name:     "Second nil",
			a:        "not nil",
			b:        nil,
			expected: false,
		},
		{
			name:     "Both nil",
			a:        nil,
			b:        nil,
			expected: true,
		},
		{
			name:     "Different types",
			a:        42,
			b:        "42",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := equals(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("equals(%v, %v) = %v; want %v", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}
