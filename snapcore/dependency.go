package snapcore

import (
	"fmt"
	"sort"

	"github.com/guanzhenxing/go-snap/errors"
)

// DependencyGraph 管理组件间的依赖关系，确保正确的初始化和关闭顺序
type DependencyGraph struct {
	// 组件名称到组件的映射
	components map[string]Component

	// 组件依赖关系 (组件名 -> 依赖的组件名列表)
	dependencies map[string][]string

	// 逆依赖关系 (组件名 -> 依赖于该组件的组件名列表)
	reverseDependencies map[string][]string

	// 组件状态跟踪
	initialized map[string]bool
	started     map[string]bool
}

// NewDependencyGraph 创建一个新的依赖图
func NewDependencyGraph() *DependencyGraph {
	return &DependencyGraph{
		components:          make(map[string]Component),
		dependencies:        make(map[string][]string),
		reverseDependencies: make(map[string][]string),
		initialized:         make(map[string]bool),
		started:             make(map[string]bool),
	}
}

// AddComponent 添加组件到依赖图
func (g *DependencyGraph) AddComponent(component Component, dependencies ...string) error {
	name := component.Name()

	// 检查组件是否已存在
	if _, exists := g.components[name]; exists {
		return errors.Errorf("component '%s' already exists", name)
	}

	// 添加组件
	g.components[name] = component
	g.dependencies[name] = dependencies

	// 更新逆依赖关系
	for _, dep := range dependencies {
		g.reverseDependencies[dep] = append(g.reverseDependencies[dep], name)
	}

	// 初始化状态
	g.initialized[name] = false
	g.started[name] = false

	return nil
}

// RemoveComponent 从依赖图中移除组件
func (g *DependencyGraph) RemoveComponent(name string) {
	// 移除逆依赖
	for dep := range g.dependencies {
		for i, revDep := range g.reverseDependencies[dep] {
			if revDep == name {
				g.reverseDependencies[dep] = append(g.reverseDependencies[dep][:i], g.reverseDependencies[dep][i+1:]...)
				break
			}
		}
	}

	// 移除依赖
	delete(g.dependencies, name)

	// 移除组件
	delete(g.components, name)
	delete(g.initialized, name)
	delete(g.started, name)
}

// GetComponent 获取指定名称的组件
func (g *DependencyGraph) GetComponent(name string) (Component, bool) {
	component, exists := g.components[name]
	return component, exists
}

// HasComponent 检查组件是否存在
func (g *DependencyGraph) HasComponent(name string) bool {
	_, exists := g.components[name]
	return exists
}

// GetAllComponents 获取所有组件
func (g *DependencyGraph) GetAllComponents() map[string]Component {
	result := make(map[string]Component)
	for name, component := range g.components {
		result[name] = component
	}
	return result
}

// IsInitialized 检查组件是否已初始化
func (g *DependencyGraph) IsInitialized(name string) bool {
	initialized, exists := g.initialized[name]
	if !exists {
		return false
	}
	return initialized
}

// IsStarted 检查组件是否已启动
func (g *DependencyGraph) IsStarted(name string) bool {
	started, exists := g.started[name]
	if !exists {
		return false
	}
	return started
}

// SetInitialized 设置组件初始化状态
func (g *DependencyGraph) SetInitialized(name string, initialized bool) {
	g.initialized[name] = initialized
}

// SetStarted 设置组件启动状态
func (g *DependencyGraph) SetStarted(name string, started bool) {
	g.started[name] = started
}

// GetDependencies 获取组件的依赖
func (g *DependencyGraph) GetDependencies(name string) []string {
	return g.dependencies[name]
}

// GetReverseDependencies 获取依赖于指定组件的组件列表
func (g *DependencyGraph) GetReverseDependencies(name string) []string {
	return g.reverseDependencies[name]
}

// CheckCircularDependencies 检查是否存在循环依赖
func (g *DependencyGraph) CheckCircularDependencies() error {
	// 构建完整的依赖图
	dependencies := make(map[string][]string)
	for name, deps := range g.dependencies {
		dependencies[name] = deps
	}

	// 访问标记
	visited := make(map[string]bool)
	stack := make(map[string]bool)

	// 对每个组件进行DFS检查
	for name := range g.components {
		if !visited[name] {
			if err := g.dfs(name, dependencies, visited, stack); err != nil {
				return err
			}
		}
	}

	return nil
}

// dfs 深度优先搜索检查循环依赖
func (g *DependencyGraph) dfs(name string, dependencies map[string][]string, visited, stack map[string]bool) error {
	visited[name] = true
	stack[name] = true

	for _, dep := range dependencies[name] {
		if !g.HasComponent(dep) {
			return errors.Errorf("component '%s' depends on non-existent component '%s'", name, dep)
		}

		if !visited[dep] {
			if err := g.dfs(dep, dependencies, visited, stack); err != nil {
				return err
			}
		} else if stack[dep] {
			return errors.Errorf("circular dependency detected: '%s' -> '%s'", name, dep)
		}
	}

	stack[name] = false
	return nil
}

// SortComponentsByDependency 根据依赖关系对组件进行拓扑排序
func (g *DependencyGraph) SortComponentsByDependency() ([]string, error) {
	// 检查是否有循环依赖
	if err := g.CheckCircularDependencies(); err != nil {
		return nil, err
	}

	// 构建入度表
	inDegree := make(map[string]int)
	for name := range g.components {
		inDegree[name] = 0
	}

	for _, deps := range g.dependencies {
		for _, dep := range deps {
			inDegree[dep]++
		}
	}

	// 找到所有入度为0的节点
	var queue []string
	for name := range g.components {
		if inDegree[name] == 0 {
			queue = append(queue, name)
		}
	}

	// 拓扑排序
	var result []string
	for len(queue) > 0 {
		// 出队
		node := queue[0]
		queue = queue[1:]
		result = append(result, node)

		// 减少相邻节点的入度
		for _, revDep := range g.reverseDependencies[node] {
			inDegree[revDep]--
			if inDegree[revDep] == 0 {
				queue = append(queue, revDep)
			}
		}
	}

	// 检查是否所有节点都已处理
	if len(result) != len(g.components) {
		return nil, fmt.Errorf("dependency graph has cycles")
	}

	return result, nil
}

// SortComponentsForInitialization 排序组件以便初始化
// 首先按组件类型排序，然后在每种类型内部按依赖关系排序
func (g *DependencyGraph) SortComponentsForInitialization() ([]string, error) {
	// 获取拓扑排序结果
	order, err := g.SortComponentsByDependency()
	if err != nil {
		return nil, err
	}

	// 按组件类型分组
	typeGroups := make(map[ComponentType][]string)
	for _, name := range order {
		component := g.components[name]
		compType := component.Type()
		typeGroups[compType] = append(typeGroups[compType], name)
	}

	// 按类型优先级合并
	var result []string
	for _, t := range []ComponentType{
		ComponentTypeInfrastructure,
		ComponentTypeDataSource,
		ComponentTypeCore,
		ComponentTypeWeb,
	} {
		result = append(result, typeGroups[t]...)
	}

	return result, nil
}

// SortComponentsForShutdown 排序组件以便关闭
// 关闭顺序与初始化相反
func (g *DependencyGraph) SortComponentsForShutdown() ([]string, error) {
	initOrder, err := g.SortComponentsForInitialization()
	if err != nil {
		return nil, err
	}

	// 反转顺序
	shutdownOrder := make([]string, len(initOrder))
	for i, name := range initOrder {
		shutdownOrder[len(initOrder)-1-i] = name
	}

	return shutdownOrder, nil
}

// GetComponentsByType 获取指定类型的所有组件
func (g *DependencyGraph) GetComponentsByType(t ComponentType) []Component {
	var result []Component
	for _, component := range g.components {
		if component.Type() == t {
			result = append(result, component)
		}
	}

	// 排序以保证稳定输出
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name() < result[j].Name()
	})

	return result
}
