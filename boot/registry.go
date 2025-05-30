// Package boot 提供类似Spring Boot的自动配置框架
// 该包实现了一个组件注册、依赖注入和生命周期管理系统
// 主要功能包括：
// - 组件注册与工厂模式支持
// - 基于拓扑排序的依赖解析
// - 组件生命周期管理（初始化、启动、停止）
// - 健康检查和指标收集
//
// 架构概述:
// ComponentRegistry 是核心类，管理所有组件及其依赖关系
// 组件可以通过两种方式注册:
// 1. 直接注册已创建的组件实例
// 2. 注册组件工厂，在需要时延迟创建组件
//
// 使用示例:
//
//	// 创建属性源
//	props := boot.NewDefaultPropertySource()
//	props.SetProperty("db.url", "jdbc:mysql://localhost:3306/mydb")
//
//	// 创建注册表
//	ctx := context.Background()
//	registry := boot.NewComponentRegistry(ctx, props)
//
//	// 注册组件
//	registry.RegisterComponent(&MyComponent{})
//
//	// 注册工厂
//	registry.RegisterFactory("dataSource", &DataSourceFactory{})
//
//	// 解析依赖
//	if err := registry.ResolveDependencies(); err != nil {
//	    // 处理错误
//	}
//
//	// 获取组件
//	if component, exists := registry.GetComponent("dataSource"); exists {
//	    // 使用组件
//	}
package boot

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"
)

// ComponentRegistry 组件注册表，是依赖注入系统的核心
// 负责组件的注册、查找、依赖解析和生命周期管理
// 线程安全，可在并发环境中使用
type ComponentRegistry struct {
	// components 已注册的组件实例，键为组件名称
	components map[string]Component
	// factories 组件工厂映射，用于按需延迟创建组件
	factories map[string]ComponentFactory
	// dependencies 组件依赖关系，键为组件名称，值为依赖的组件名称列表
	dependencies map[string][]string
	// dependencyGraph 依赖图，用于依赖解析和循环依赖检测
	dependencyGraph map[string][]string
	// mutex 保护并发访问的互斥锁
	mutex sync.RWMutex
	// factoryContext 用于创建组件的上下文
	factoryContext context.Context
	// propertySource 配置属性源
	propertySource PropertySource
	// healthChecker 健康检查器
	healthChecker ComponentHealthChecker
	// metrics 注册表性能和状态指标
	metrics *RegistryMetrics
}

// RegistryMetrics 注册表指标，收集组件注册表的性能和状态数据
type RegistryMetrics struct {
	// ComponentCount 已注册组件数量
	ComponentCount int
	// FactoryCount 已注册工厂数量
	FactoryCount int
	// DependencyResolvTime 依赖解析耗时
	DependencyResolvTime time.Duration
	// HealthCheckCount 健康检查次数
	HealthCheckCount int
	// FailedComponents 初始化或运行失败的组件列表
	FailedComponents []string
	// mutex 保护并发访问的互斥锁
	mutex sync.RWMutex
}

// NewComponentRegistry 创建一个新的组件注册表实例
// 参数：
//
//	ctx: 上下文，用于创建组件和传递取消信号
//	props: 属性源，为组件提供配置
//
// 返回：
//
//	初始化好的ComponentRegistry实例
//
// 示例：
//
//	registry := boot.NewComponentRegistry(context.Background(), boot.NewDefaultPropertySource())
func NewComponentRegistry(ctx context.Context, props PropertySource) *ComponentRegistry {
	return &ComponentRegistry{
		components:      make(map[string]Component),
		factories:       make(map[string]ComponentFactory),
		dependencies:    make(map[string][]string),
		dependencyGraph: make(map[string][]string),
		factoryContext:  ctx,
		propertySource:  props,
		healthChecker:   &DefaultHealthChecker{},
		metrics: &RegistryMetrics{
			FailedComponents: make([]string, 0),
		},
	}
}

// RegisterComponent 注册一个组件实例
// 参数：
//
//	component: 要注册的组件实例
//
// 返回：
//
//	error: 如果组件已存在或注册失败，返回错误；否则返回nil
//
// 线程安全：可从多个goroutine并发调用
// 示例：
//
//	err := registry.RegisterComponent(&LoggerComponent{})
func (r *ComponentRegistry) RegisterComponent(component Component) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	name := component.Name()
	if _, exists := r.components[name]; exists {
		return NewComponentError(name, "register", "组件已存在", ErrComponentExists)
	}

	r.components[name] = component
	r.updateMetrics()
	return nil
}

// RegisterFactory 注册一个组件工厂，用于延迟创建组件
// 参数：
//
//	name: 组件名称
//	factory: 组件工厂实例
//
// 返回：
//
//	error: 如果配置验证失败或注册失败，返回错误；否则返回nil
//
// 注意：
//   - 工厂配置会在注册时验证，但组件实例直到请求时才会创建
//   - 注册工厂会触发依赖图重建，在大量注册时可能影响性能
//
// 示例：
//
//	err := registry.RegisterFactory("database", &DatabaseFactory{})
func (r *ComponentRegistry) RegisterFactory(name string, factory ComponentFactory) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// 验证配置
	if err := factory.ValidateConfig(r.propertySource); err != nil {
		return NewComponentError(name, "register_factory", "配置验证失败", err)
	}

	r.factories[name] = factory
	r.dependencies[name] = factory.Dependencies()
	r.buildDependencyGraph()
	r.updateMetrics()
	return nil
}

// GetComponent 获取组件，如果组件不存在但有对应的工厂，则会创建
// 使用双重检查锁定模式以提高并发性能
// 参数：
//
//	name: 组件名称
//
// 返回：
//
//	Component: 组件实例，如果找不到则为nil
//	bool: 是否找到组件
//
// 性能：
//   - 对于已存在的组件，仅需要读锁，性能较高
//   - 对于需要创建的组件，会获取写锁，可能影响并发性能
//
// 示例：
//
//	if db, exists := registry.GetComponent("database"); exists {
//	    // 使用数据库组件
//	}
func (r *ComponentRegistry) GetComponent(name string) (Component, bool) {
	// 第一次检查（读锁）
	r.mutex.RLock()
	component, exists := r.components[name]
	if exists {
		r.mutex.RUnlock()
		return component, true
	}

	// 检查是否有对应的工厂
	factory, hasFactory := r.factories[name]
	r.mutex.RUnlock()

	if !hasFactory {
		return nil, false
	}

	// 第二次检查（写锁）
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// 再次检查组件是否已被其他goroutine创建
	if component, exists := r.components[name]; exists {
		return component, true
	}

	// 创建组件实例
	component, err := factory.Create(r.factoryContext, r.propertySource)
	if err != nil {
		r.recordFailedComponent(name, err)
		return nil, false
	}

	r.components[name] = component
	r.updateMetrics()
	return component, true
}

// GetComponentByType 获取指定类型的第一个组件
// 参数：
//
//	componentType: 组件类型枚举值
//
// 返回：
//
//	Component: 找到的第一个匹配类型的组件，如果没有则为nil
//	bool: 是否找到组件
//
// 注意：
//   - 如果有多个相同类型的组件，只返回找到的第一个
//   - 如果需要所有同类型组件，请使用GetComponentsByType
//
// 示例：
//
//	if webServer, exists := registry.GetComponentByType(ComponentTypeWeb); exists {
//	    // 使用Web服务器组件
//	}
func (r *ComponentRegistry) GetComponentByType(componentType ComponentType) (Component, bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	for _, component := range r.components {
		if component.Type() == componentType {
			return component, true
		}
	}

	return nil, false
}

// GetComponentsByType 获取指定类型的所有组件
// 参数：
//
//	componentType: 组件类型枚举值
//
// 返回：
//
//	[]Component: 所有匹配类型的组件切片，如果没有则为空切片
//
// 示例：
//
//	dataSources := registry.GetComponentsByType(ComponentTypeDataSource)
//	for _, ds := range dataSources {
//	    // 处理每个数据源
//	}
func (r *ComponentRegistry) GetComponentsByType(componentType ComponentType) []Component {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var result []Component
	for _, component := range r.components {
		if component.Type() == componentType {
			result = append(result, component)
		}
	}

	return result
}

// GetAllComponents 获取所有已注册的组件
// 返回：
//
//	map[string]Component: 组件名称到组件实例的映射
//
// 注意：
//
//	返回的是组件的浅拷贝映射，修改返回的映射不会影响注册表
//
// 示例：
//
//	allComponents := registry.GetAllComponents()
//	for name, component := range allComponents {
//	    fmt.Printf("组件: %s, 类型: %v\n", name, component.Type())
//	}
func (r *ComponentRegistry) GetAllComponents() map[string]Component {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	result := make(map[string]Component)
	for name, component := range r.components {
		result[name] = component
	}

	return result
}

// ResolveDependencies 解析所有组件依赖（使用拓扑排序）
// 根据组件工厂声明的依赖关系创建所有组件实例
// 返回：
//
//	error: 如果存在循环依赖或创建组件失败，返回错误；否则返回nil
//
// 性能：
//   - 此方法可能耗时较长，建议在应用启动阶段调用
//   - 方法会记录执行时间到指标中
//
// 示例：
//
//	if err := registry.ResolveDependencies(); err != nil {
//	    log.Fatalf("依赖解析失败: %v", err)
//	}
func (r *ComponentRegistry) ResolveDependencies() error {
	start := time.Now()
	defer func() {
		r.metrics.mutex.Lock()
		r.metrics.DependencyResolvTime = time.Since(start)
		r.metrics.mutex.Unlock()
	}()

	// 先检查是否有循环依赖
	if err := r.checkCircularDependencies(); err != nil {
		return err
	}

	// 使用拓扑排序获取依赖顺序
	sortedNames, err := r.topologicalSort()
	if err != nil {
		return err
	}

	// 按依赖顺序创建组件
	for _, name := range sortedNames {
		if _, exists := r.components[name]; exists {
			continue // 组件已存在，跳过
		}

		factory, exists := r.factories[name]
		if !exists {
			continue // 没有工厂，跳过
		}

		// 确保所有依赖都已创建
		for _, depName := range factory.Dependencies() {
			if _, exists := r.components[depName]; !exists {
				return NewDependencyError(
					fmt.Sprintf("组件 %s 的依赖 %s 未找到", name, depName),
					[]string{name, depName},
					nil,
				)
			}
		}

		// 创建组件
		component, err := factory.Create(r.factoryContext, r.propertySource)
		if err != nil {
			r.recordFailedComponent(name, err)
			return NewComponentError(name, "create", "创建组件失败", err)
		}

		r.mutex.Lock()
		r.components[name] = component
		r.updateMetrics()
		r.mutex.Unlock()
	}

	return nil
}

// topologicalSort 拓扑排序
func (r *ComponentRegistry) topologicalSort() ([]string, error) {
	// 计算入度
	inDegree := make(map[string]int)
	for name := range r.factories {
		inDegree[name] = 0
	}

	for _, deps := range r.dependencies {
		for _, dep := range deps {
			inDegree[dep]++
		}
	}

	// 找到所有入度为0的节点
	queue := make([]string, 0)
	for name, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, name)
		}
	}

	result := make([]string, 0)
	for len(queue) > 0 {
		// 取出队首元素
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)

		// 减少相邻节点的入度
		for _, dep := range r.dependencies[current] {
			inDegree[dep]--
			if inDegree[dep] == 0 {
				queue = append(queue, dep)
			}
		}
	}

	// 检查是否有循环依赖
	if len(result) != len(r.factories) {
		return nil, NewDependencyError("存在循环依赖", result, nil)
	}

	return result, nil
}

// checkCircularDependencies 检查循环依赖
func (r *ComponentRegistry) checkCircularDependencies() error {
	visited := make(map[string]bool)
	stack := make(map[string]bool)

	for name := range r.factories {
		if !visited[name] {
			if err := r.dfs(name, visited, stack, []string{name}); err != nil {
				return err
			}
		}
	}

	return nil
}

// dfs 深度优先搜索
func (r *ComponentRegistry) dfs(name string, visited, stack map[string]bool, path []string) error {
	visited[name] = true
	stack[name] = true

	for _, dep := range r.dependencies[name] {
		if !visited[dep] {
			newPath := append(path, dep)
			if err := r.dfs(dep, visited, stack, newPath); err != nil {
				return err
			}
		} else if stack[dep] {
			// 发现循环依赖
			cycle := append(path, dep)
			return NewDependencyError(
				fmt.Sprintf("发现循环依赖: %v", cycle),
				cycle,
				nil,
			)
		}
	}

	stack[name] = false
	return nil
}

// buildDependencyGraph 构建依赖图
func (r *ComponentRegistry) buildDependencyGraph() {
	r.dependencyGraph = make(map[string][]string)
	for name, deps := range r.dependencies {
		r.dependencyGraph[name] = make([]string, len(deps))
		copy(r.dependencyGraph[name], deps)
	}
}

// GetAllComponentsSorted 按初始化顺序获取所有组件
func (r *ComponentRegistry) GetAllComponentsSorted() []Component {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var result []Component

	// 按组件类型顺序添加
	for _, t := range []ComponentType{
		ComponentTypeInfrastructure,
		ComponentTypeDataSource,
		ComponentTypeCore,
		ComponentTypeWeb,
	} {
		comps := r.getComponentsByTypeUnsafe(t)

		// 按名称排序，确保顺序一致
		sort.Slice(comps, func(i, j int) bool {
			return comps[i].Name() < comps[j].Name()
		})

		result = append(result, comps...)
	}

	return result
}

// getComponentsByTypeUnsafe 获取指定类型的所有组件（不加锁）
func (r *ComponentRegistry) getComponentsByTypeUnsafe(componentType ComponentType) []Component {
	var result []Component
	for _, component := range r.components {
		if component.Type() == componentType {
			result = append(result, component)
		}
	}
	return result
}

// GetAllComponentsForShutdown 获取关闭顺序的组件列表（反向）
func (r *ComponentRegistry) GetAllComponentsForShutdown() []Component {
	components := r.GetAllComponentsSorted()

	// 反转顺序
	for i, j := 0, len(components)-1; i < j; i, j = i+1, j-1 {
		components[i], components[j] = components[j], components[i]
	}

	return components
}

// HealthCheck 执行所有组件的健康检查
func (r *ComponentRegistry) HealthCheck() map[string]error {
	r.mutex.RLock()
	components := make(map[string]Component)
	for name, comp := range r.components {
		components[name] = comp
	}
	r.mutex.RUnlock()

	results := make(map[string]error)
	for name, component := range components {
		if err := r.healthChecker.CheckHealth(component); err != nil {
			results[name] = err
			r.recordFailedComponent(name, err)
		}
	}

	r.metrics.mutex.Lock()
	r.metrics.HealthCheckCount++
	r.metrics.mutex.Unlock()

	return results
}

// GetMetrics 获取注册表指标
func (r *ComponentRegistry) GetMetrics() *RegistryMetrics {
	r.metrics.mutex.RLock()
	defer r.metrics.mutex.RUnlock()

	// 返回副本
	return &RegistryMetrics{
		ComponentCount:       r.metrics.ComponentCount,
		FactoryCount:         r.metrics.FactoryCount,
		DependencyResolvTime: r.metrics.DependencyResolvTime,
		HealthCheckCount:     r.metrics.HealthCheckCount,
		FailedComponents:     append([]string{}, r.metrics.FailedComponents...),
	}
}

// updateMetrics 更新指标
func (r *ComponentRegistry) updateMetrics() {
	r.metrics.mutex.Lock()
	defer r.metrics.mutex.Unlock()

	r.metrics.ComponentCount = len(r.components)
	r.metrics.FactoryCount = len(r.factories)
}

// recordFailedComponent 记录失败的组件
func (r *ComponentRegistry) recordFailedComponent(name string, err error) {
	r.metrics.mutex.Lock()
	defer r.metrics.mutex.Unlock()

	// 避免重复记录
	for _, failed := range r.metrics.FailedComponents {
		if failed == name {
			return
		}
	}

	r.metrics.FailedComponents = append(r.metrics.FailedComponents, name)
}
