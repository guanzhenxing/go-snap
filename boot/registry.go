package boot

import (
	"context"
	"sort"
	"sync"
)

// ComponentRegistry 组件注册表
type ComponentRegistry struct {
	components     map[string]Component
	factories      map[string]ComponentFactory
	dependencies   map[string][]string
	mutex          sync.RWMutex
	factoryContext context.Context
	propertySource PropertySource
}

// NewComponentRegistry 创建组件注册表
func NewComponentRegistry(ctx context.Context, props PropertySource) *ComponentRegistry {
	return &ComponentRegistry{
		components:     make(map[string]Component),
		factories:      make(map[string]ComponentFactory),
		dependencies:   make(map[string][]string),
		factoryContext: ctx,
		propertySource: props,
	}
}

// RegisterComponent 注册组件
func (r *ComponentRegistry) RegisterComponent(component Component) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	name := component.Name()
	if _, exists := r.components[name]; exists {
		return ErrComponentExists
	}

	r.components[name] = component
	return nil
}

// RegisterFactory 注册组件工厂
func (r *ComponentRegistry) RegisterFactory(name string, factory ComponentFactory) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.factories[name] = factory
	r.dependencies[name] = factory.Dependencies()
}

// GetComponent 获取组件
func (r *ComponentRegistry) GetComponent(name string) (Component, bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	// 先从已注册组件中查找
	component, exists := r.components[name]
	if exists {
		return component, true
	}

	// 如果组件不存在但有对应的工厂，则创建组件
	factory, exists := r.factories[name]
	if !exists {
		return nil, false
	}

	// 先解锁，避免在创建组件过程中发生死锁
	r.mutex.RUnlock()

	// 创建组件实例
	component, err := factory.Create(r.factoryContext, r.propertySource)
	if err != nil {
		return nil, false
	}

	// 重新加锁，注册组件
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.components[name] = component
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

// ResolveDependencies 解析所有组件依赖
func (r *ComponentRegistry) ResolveDependencies() error {
	// 先检查是否有循环依赖
	if err := r.checkCircularDependencies(); err != nil {
		return err
	}

	// 为每个工厂创建组件实例
	for name, factory := range r.factories {
		// 如果组件已经存在，则跳过
		if _, exists := r.components[name]; exists {
			continue
		}

		// 确保所有依赖都已创建
		for _, depName := range factory.Dependencies() {
			if _, exists := r.components[depName]; !exists {
				// 依赖未创建，尝试从工厂创建
				depFactory, exists := r.factories[depName]
				if !exists {
					return &ConfigError{Message: "组件依赖未找到: " + depName}
				}

				// 创建依赖组件
				depComponent, err := depFactory.Create(r.factoryContext, r.propertySource)
				if err != nil {
					return &ConfigError{Message: "创建依赖组件失败: " + depName, Cause: err}
				}

				// 注册依赖组件
				r.components[depName] = depComponent
			}
		}

		// 创建当前组件
		component, err := factory.Create(r.factoryContext, r.propertySource)
		if err != nil {
			return &ConfigError{Message: "创建组件失败: " + name, Cause: err}
		}

		// 注册组件
		r.components[name] = component
	}

	return nil
}

// checkCircularDependencies 检查循环依赖
func (r *ComponentRegistry) checkCircularDependencies() error {
	// 为每个组件创建访问标记
	visited := make(map[string]bool)
	stack := make(map[string]bool)

	// 对每个组件进行深度优先搜索
	for name := range r.factories {
		if !visited[name] {
			if err := r.dfs(name, visited, stack); err != nil {
				return err
			}
		}
	}

	return nil
}

// dfs 深度优先搜索
func (r *ComponentRegistry) dfs(name string, visited, stack map[string]bool) error {
	visited[name] = true
	stack[name] = true

	// 检查所有依赖
	for _, dep := range r.dependencies[name] {
		if !visited[dep] {
			if err := r.dfs(dep, visited, stack); err != nil {
				return err
			}
		} else if stack[dep] {
			// 发现循环依赖
			return &ConfigError{Message: "发现循环依赖: " + name + " -> " + dep}
		}
	}

	stack[name] = false
	return nil
}

// GetAllComponentsSorted 按初始化顺序获取所有组件
func (r *ComponentRegistry) GetAllComponentsSorted() []Component {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	// 按组件类型排序
	var result []Component

	// 按组件类型顺序添加
	for _, t := range []ComponentType{
		ComponentTypeInfrastructure,
		ComponentTypeDataSource,
		ComponentTypeCore,
		ComponentTypeWeb,
	} {
		// 获取指定类型的所有组件
		comps := r.GetComponentsByType(t)

		// 按名称排序，确保顺序一致
		sort.Slice(comps, func(i, j int) bool {
			return comps[i].Name() < comps[j].Name()
		})

		result = append(result, comps...)
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
