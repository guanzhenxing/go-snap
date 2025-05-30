# 贡献指南

感谢您对 Go-Snap 框架的关注！我们欢迎任何形式的贡献，包括但不限于代码提交、问题报告、功能建议、文档改进等。

## 🤝 如何贡献

### 1. 代码贡献

#### 开发环境设置

```bash
# 1. Fork 项目到你的 GitHub 账户

# 2. 克隆你的 Fork
git clone https://github.com/your-username/go-snap.git
cd go-snap

# 3. 添加上游仓库
git remote add upstream https://github.com/guanzhenxing/go-snap.git

# 4. 安装依赖
go mod download

# 5. 运行测试确保环境正常
go test ./...
```

#### 开发流程

```bash
# 1. 从最新的 main 分支创建特性分支
git checkout main
git pull upstream main
git checkout -b feature/your-feature-name

# 2. 进行开发
# 编写代码、测试、文档

# 3. 运行测试和检查
go test ./...
go vet ./...
go fmt ./...

# 4. 提交更改
git add .
git commit -m "feat: 添加新功能 xyz"

# 5. 推送到你的 Fork
git push origin feature/your-feature-name

# 6. 创建 Pull Request
```

### 2. 问题报告

在报告问题时，请包含以下信息：

- **Go 版本**: `go version`
- **操作系统**: Windows/macOS/Linux
- **Go-Snap 版本**: 使用的框架版本
- **问题描述**: 详细描述遇到的问题
- **重现步骤**: 提供复现问题的具体步骤
- **期望行为**: 描述期望的正确行为
- **实际行为**: 描述实际发生的行为
- **错误信息**: 完整的错误消息和堆栈跟踪

#### 问题报告模板

```markdown
**Go 版本**: go1.21.0

**操作系统**: macOS 13.0

**Go-Snap 版本**: v1.0.0

**问题描述**:
应用启动时缓存组件初始化失败

**重现步骤**:
1. 配置 Redis 缓存
2. 启动应用
3. 观察错误日志

**期望行为**:
应用正常启动，缓存组件成功初始化

**实际行为**:
应用启动失败，提示缓存连接错误

**错误信息**:
```
ERROR: 缓存组件初始化失败: dial tcp 127.0.0.1:6379: connect: connection refused
```

**配置文件**:
```yaml
cache:
  type: redis
  redis:
    addr: "localhost:6379"
```
```

### 3. 功能建议

提交功能建议时，请说明：

- **功能描述**: 详细描述建议的功能
- **使用场景**: 什么情况下会用到这个功能
- **实现思路**: 如果有想法，请描述可能的实现方式
- **优先级**: 这个功能的重要程度
- **替代方案**: 现有的解决方案或变通方法

### 4. 文档贡献

文档同样重要！您可以：

- 修复文档中的错误
- 改进现有文档的表述
- 添加缺失的文档
- 翻译文档到其他语言
- 添加更多使用示例

## 📋 代码规范

### Go 代码规范

我们遵循标准的 Go 代码规范：

#### 1. 代码格式

```bash
# 使用 gofmt 格式化代码
go fmt ./...

# 使用 goimports 处理导入
goimports -w .
```

#### 2. 命名规范

```go
// ✅ 好的命名
type UserService struct{}
func (s *UserService) GetUserByID(id string) (*User, error)
const MaxRetryCount = 3
var DefaultTimeout = time.Second * 30

// ❌ 避免的命名
type userservice struct{}
func (s *userservice) getUserById(id string) (*User, error)
const max_retry_count = 3
var default_timeout = time.Second * 30
```

#### 3. 注释规范

```go
// Package cache provides caching functionality for Go-Snap framework.
//
// It supports multiple cache backends including memory cache, Redis cache,
// and multi-level cache. The cache interface is unified across all backends.
package cache

// UserService provides user-related business operations.
type UserService struct {
    logger logger.Logger
    repo   UserRepository
}

// GetUserByID retrieves a user by their unique identifier.
//
// It returns an error if the user is not found or if there's a database error.
func (s *UserService) GetUserByID(id string) (*User, error) {
    // Implementation here
}
```

#### 4. 错误处理

```go
// ✅ 好的错误处理
func (s *UserService) CreateUser(user *User) error {
    if err := s.validateUser(user); err != nil {
        return errors.WrapWithCode(err, errors.CodeValidation, "用户验证失败")
    }
    
    if err := s.repo.Create(user); err != nil {
        return errors.WrapWithCode(err, errors.CodeDatabaseError, "创建用户失败")
    }
    
    return nil
}

// ❌ 避免的错误处理
func (s *UserService) CreateUser(user *User) error {
    err := s.validateUser(user)
    if err != nil {
        return err // 没有添加上下文
    }
    
    err = s.repo.Create(user)
    if err != nil {
        return fmt.Errorf("error: %v", err) // 使用通用错误
    }
    
    return nil
}
```

#### 5. 测试规范

```go
// 测试文件命名: xxx_test.go
package user

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

func TestUserService_GetUserByID(t *testing.T) {
    // 使用表驱动测试
    tests := []struct {
        name    string
        userID  string
        setup   func(*MockUserRepository)
        want    *User
        wantErr bool
    }{
        {
            name:   "成功获取用户",
            userID: "123",
            setup: func(m *MockUserRepository) {
                m.On("FindByID", "123").Return(&User{ID: "123", Name: "John"}, nil)
            },
            want:    &User{ID: "123", Name: "John"},
            wantErr: false,
        },
        {
            name:   "用户不存在",
            userID: "999",
            setup: func(m *MockUserRepository) {
                m.On("FindByID", "999").Return(nil, ErrUserNotFound)
            },
            want:    nil,
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockRepo := &MockUserRepository{}
            tt.setup(mockRepo)
            
            service := NewUserService(mockRepo)
            got, err := service.GetUserByID(tt.userID)
            
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.want, got)
            }
            
            mockRepo.AssertExpectations(t)
        })
    }
}
```

### Commit 消息规范

我们使用 [Conventional Commits](https://www.conventionalcommits.org/) 规范：

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

#### 类型 (type)

- `feat`: 新功能
- `fix`: 修复 bug
- `docs`: 文档变更
- `style`: 代码格式变更（不影响功能）
- `refactor`: 重构（既不是新功能也不是修复）
- `perf`: 性能优化
- `test`: 添加或修改测试
- `build`: 构建系统或外部依赖变更
- `ci`: CI/CD 配置变更
- `chore`: 其他变更

#### 示例

```bash
# 新功能
feat(cache): 添加 Redis 集群支持

# 修复 bug
fix(boot): 修复组件循环依赖检测错误

# 文档更新
docs: 更新 Cache 模块使用文档

# 重构
refactor(config): 重构配置加载逻辑提高性能

# 性能优化
perf(logger): 优化日志写入性能

# 测试
test(user): 添加用户服务单元测试

# Breaking change
feat(boot)!: 重构应用启动 API

BREAKING CHANGE: Boot.Run() 方法签名已更改
```

## 🧪 测试指南

### 测试覆盖率

我们要求新代码的测试覆盖率不低于 80%。

```bash
# 运行测试并查看覆盖率
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 测试类型

#### 1. 单元测试

```go
func TestCacheComponent_Initialize(t *testing.T) {
    component := &CacheComponent{
        BaseComponent: boot.NewBaseComponent("cache", boot.ComponentTypeDataSource),
    }
    
    err := component.Initialize(context.Background())
    assert.NoError(t, err)
    assert.Equal(t, boot.ComponentStatusInitialized, component.GetStatus())
}
```

#### 2. 集成测试

```go
func TestUserService_Integration(t *testing.T) {
    // 跳过短测试
    if testing.Short() {
        t.Skip("跳过集成测试")
    }
    
    // 设置真实的数据库连接
    db := setupTestDatabase(t)
    defer cleanupTestDatabase(t, db)
    
    service := NewUserService(db)
    
    user := &User{Name: "Test User", Email: "test@example.com"}
    err := service.CreateUser(user)
    assert.NoError(t, err)
    
    retrieved, err := service.GetUserByID(user.ID)
    assert.NoError(t, err)
    assert.Equal(t, user.Name, retrieved.Name)
}
```

#### 3. 基准测试

```go
func BenchmarkCacheSet(b *testing.B) {
    cache := NewMemoryCache()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        key := fmt.Sprintf("key-%d", i)
        cache.Set(context.Background(), key, "value", time.Hour)
    }
}
```

### Mock 和 Stub

使用 [testify/mock](https://github.com/stretchr/testify) 或 [GoMock](https://github.com/golang/mock) 创建 mock：

```go
//go:generate mockgen -source=user.go -destination=mocks/user_mock.go

type MockUserRepository struct {
    mock.Mock
}

func (m *MockUserRepository) FindByID(id string) (*User, error) {
    args := m.Called(id)
    return args.Get(0).(*User), args.Error(1)
}
```

## 📦 发布流程

### 版本号规范

我们遵循 [Semantic Versioning](https://semver.org/) 规范：

- `MAJOR.MINOR.PATCH` (例如: 1.2.3)
- `MAJOR`: 不兼容的 API 变更
- `MINOR`: 向后兼容的功能增加
- `PATCH`: 向后兼容的问题修复

### 发布步骤

1. **更新版本号**
```bash
# 更新 version.go 文件
echo 'package version; const Version = "1.2.3"' > version.go
```

2. **更新 CHANGELOG**
```bash
# 在 CHANGELOG.md 中添加新版本信息
```

3. **创建标签**
```bash
git tag -a v1.2.3 -m "Release version 1.2.3"
git push origin v1.2.3
```

4. **创建 GitHub Release**
- 在 GitHub 上创建新的 Release
- 包含版本说明和变更列表

## 🎯 开发最佳实践

### 1. 设计原则

- **单一职责**: 每个组件、函数只做一件事
- **开放封闭**: 对扩展开放，对修改封闭
- **依赖倒置**: 依赖抽象而非具体实现
- **接口隔离**: 使用小而专一的接口
- **组合优于继承**: 通过组合实现功能扩展

### 2. 性能考虑

- 避免不必要的内存分配
- 使用对象池减少 GC 压力
- 合理使用缓存
- 异步处理非关键路径操作
- 使用连接池管理资源

### 3. 安全考虑

- 输入验证和清理
- 错误信息不暴露敏感数据
- 使用安全的默认配置
- 定期更新依赖库
- 遵循最小权限原则

### 4. 文档要求

- 公开 API 必须有完整注释
- 复杂逻辑要有内联注释
- 提供使用示例
- 维护 README 和 CHANGELOG
- API 变更要有迁移指南

## 🆘 获得帮助

如果在贡献过程中遇到问题：

1. **查看文档**: [Go-Snap 文档](README.md)
2. **搜索 Issues**: 看看是否有人遇到过相同问题
3. **提问**: 在 GitHub Issues 中提出问题
4. **讨论**: 在 GitHub Discussions 中参与讨论

## 📄 许可证

通过贡献代码，您同意您的贡献将在与项目相同的 [MIT 许可证](../LICENSE) 下发布。

## 🙏 致谢

感谢所有为 Go-Snap 框架做出贡献的开发者！

- 查看 [贡献者列表](https://github.com/guanzhenxing/go-snap/graphs/contributors)
- 特别感谢核心维护者和长期贡献者

---

**再次感谢您的贡献！** 🎉 