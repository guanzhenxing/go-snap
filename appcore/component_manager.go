package appcore

// ComponentManager 组件管理器，负责管理所有应用组件的生命周期
type ComponentManager struct {
	// 组件名称到组件的映射
	components map[string]Component

	// 按组件类型分组的组件
	componentsByType map[ComponentType][]Component
}

// NewComponentManager 创建组件管理器
func NewComponentManager() *ComponentManager {
	return &ComponentManager{
		components:       make(map[string]Component),
		componentsByType: make(map[ComponentType][]Component),
	}
}

// RegisterComponent 注册组件
func (m *ComponentManager) RegisterComponent(component Component) {
	name := component.Name()
	m.components[name] = component

	// 按类型分组
	compType := component.Type()
	m.componentsByType[compType] = append(m.componentsByType[compType], component)
}

// GetComponent 获取组件
func (m *ComponentManager) GetComponent(name string) (Component, bool) {
	component, exists := m.components[name]
	return component, exists
}

// GetComponentByType 获取特定类型的第一个组件
func (m *ComponentManager) GetComponentByType(compType ComponentType) (Component, bool) {
	components := m.componentsByType[compType]
	if len(components) == 0 {
		return nil, false
	}
	return components[0], true
}

// GetComponentsByType 获取特定类型的所有组件
func (m *ComponentManager) GetComponentsByType(compType ComponentType) []Component {
	return m.componentsByType[compType]
}

// GetAllComponents 获取所有注册的组件
func (m *ComponentManager) GetAllComponents() map[string]Component {
	result := make(map[string]Component)
	for name, component := range m.components {
		result[name] = component
	}
	return result
}

// GetAllComponentsSorted 按初始化顺序获取所有组件
func (m *ComponentManager) GetAllComponentsSorted() []Component {
	// 按类型组织组件
	var result []Component

	// 按组件类型顺序添加
	for _, t := range []ComponentType{
		ComponentTypeInfrastructure,
		ComponentTypeDataSource,
		ComponentTypeCore,
		ComponentTypeWeb,
	} {
		comps := m.componentsByType[t]
		result = append(result, comps...)
	}

	return result
}

// GetAllComponentsForShutdown 获取关闭顺序的组件列表（反向）
func (m *ComponentManager) GetAllComponentsForShutdown() []Component {
	components := m.GetAllComponentsSorted()

	// 反转顺序
	for i, j := 0, len(components)-1; i < j; i, j = i+1, j-1 {
		components[i], components[j] = components[j], components[i]
	}

	return components
}
