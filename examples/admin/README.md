# 管理面板示例

这个示例演示了如何使用 Go-Snap 框架的 Boot 模块创建一个完整的管理后台应用。本示例遵循标准的 Go 项目布局结构。

## 示例概述

本示例实现了一个基于 Web 的管理面板，包含以下功能：

- 用户认证（登录/注销）
- 仪表盘（系统概览）
- 用户管理（CRUD 操作）
- 系统管理
  - 系统指标监控
  - 组件健康状态
  - 组件管理（启动/停止/重启）

## 目录结构

```
examples/admin/
├── cmd/            - 应用入口点
│   └── main.go     - 主程序
├── configs/        - 配置文件目录
│   └── application.yaml - 应用配置
├── internal/       - 内部代码
│   ├── components/ - 组件实现
│   │   └── admin_component.go - 管理面板组件
│   └── handlers/   - 请求处理器
│       └── admin_handler.go   - 管理面板处理器
└── README.md       - 示例说明文档
```

## 使用方法

1. 运行示例：

```bash
go run examples/admin/cmd/main.go
```

2. 访问管理面板：

```
http://127.0.0.1:8081/admin
```

3. 登录凭据：

```
用户名: admin
密码: admin123
```

## API 端点说明

| 路径 | 方法 | 说明 | 认证 |
|------|------|------|------|
| `/admin/login` | POST | 用户登录 | 否 |
| `/admin/dashboard` | GET | 获取仪表盘数据 | 是 |
| `/admin/users` | GET | 获取用户列表 | 是 |
| `/admin/users` | POST | 创建用户 | 是 |
| `/admin/users/:id` | GET | 获取用户详情 | 是 |
| `/admin/users/:id` | PUT | 更新用户 | 是 |
| `/admin/users/:id` | DELETE | 删除用户 | 是 |
| `/admin/system/metrics` | GET | 获取系统指标 | 是 |
| `/admin/system/health` | GET | 获取系统健康状态 | 是 |
| `/admin/system/components` | GET | 获取组件列表 | 是 |
| `/admin/system/components/:name` | GET | 获取组件详情 | 是 |
| `/admin/system/components/:name/toggle` | POST | 操作组件(启动/停止/重启) | 是 |
| `/admin/health` | GET | 健康检查 | 否 |

## 技术要点

1. **标准 Go 项目布局**：
   - 遵循 Go 社区推荐的项目结构
   - 清晰的职责分离
   - 内部代码封装

2. **Boot 模块集成**：
   - 利用 Boot 模块的组件管理机制
   - 应用生命周期管理
   - 事件订阅机制

3. **组件化设计**：
   - 实现自定义 AdminComponent
   - 遵循组件生命周期
   - 支持健康检查

4. **Web 服务集成**：
   - 与 Go-Snap Web 模块集成
   - JWT 认证
   - RESTful API 设计

5. **配置管理**：
   - 基于 YAML 的配置
   - 支持默认值和覆盖

## 扩展建议

- 添加前端界面（使用 Vue.js 或 React）
- 实现真实的数据库存储
- 增强安全性（RBAC 权限控制）
- 添加审计日志
- 实现多语言支持 