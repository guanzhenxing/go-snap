package boot

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"
)

// ComponentRegistry 组件注册表
type ComponentRegistry struct {
	components      map[string]Component
	factories       map[string]ComponentFactory
	dependencies    map[string][]string
	dependencyGraph map[string][]string // 依赖图
	mutex           sync.RWMutex
	factoryContext  context.Context
	propertySource  PropertySource
	healthChecker   ComponentHealthChecker
	metrics         *RegistryMetrics
}

// RegistryMetrics 注册表指标
type RegistryMetrics struct {
	ComponentCount       int
	FactoryCount         int
	DependencyResolvTime time.Duration
	HealthCheckCount     int
	FailedComponents     []string
	mutex                sync.RWMutex
}

// NewComponentRegistry 创建组件注册表
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

// RegisterComponent 注册组件
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

// RegisterFactory 注册组件工厂
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

// GetComponent 获取组件（使用双重检查锁定模式）
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

// GetAllComponents 获取所有组件
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

		r.components[name] = component
	}

	r.updateMetrics()
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
