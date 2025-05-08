# Go-Snap 配置模块示例

本目录包含了Go-Snap配置模块的使用示例，展示了如何在项目中使用配置系统。

## 目录结构

```
config/
├── configs/                  # 配置文件目录
│   ├── app.yaml             # 默认配置文件
│   ├── app.development.yaml # 开发环境配置
│   └── app.production.yaml  # 生产环境配置
├── basic_example.go         # 基本配置使用示例
├── env_example.go           # 环境变量配置示例
├── watcher_example.go       # 配置监听示例
├── validator_example.go     # 配置验证示例
├── main.go                  # 主程序文件
└── README.md                # 本文件
```

## 运行示例

可以通过以下命令运行示例程序：

```bash
# 进入示例目录
cd examples/config

# 运行示例程序
go run .

# 直接指定要运行的示例
go run . 1  # 运行基本配置示例
go run . 2  # 运行环境变量示例
go run . 3  # 运行配置监听示例
go run . 4  # 运行配置验证示例
go run . 0  # 运行所有示例
```

## 示例说明

### 1. 基本配置示例 (basic_example.go)

展示了配置模块的基本用法，包括：
- 初始化配置系统
- 读取各种类型的配置值
- 使用结构体解析配置
- 检查配置项是否存在
- 设置新的配置值

### 2. 环境变量示例 (env_example.go)

展示了如何使用环境变量来覆盖配置文件中的值，包括：
- 设置环境变量前缀
- 启用自动环境变量支持
- 根据不同环境加载不同的配置

### 3. 配置监听示例 (watcher_example.go)

展示了如何监听配置变更，包括：
- 启用配置文件监听
- 创建监听器
- 监听特定配置项的变更
- 注册全局配置变更回调
- 取消监听

### 4. 配置验证示例 (validator_example.go)

展示了如何验证配置的有效性，包括：
- 添加必需项验证器
- 添加范围验证器
- 添加环境验证器
- 使用结构体和标签进行验证

## 配置文件说明

- `app.yaml`: 默认的应用配置
- `app.development.yaml`: 开发环境特定配置，当环境为development时会自动加载
- `app.production.yaml`: 生产环境特定配置，当环境为production时会自动加载 