# DBStore 模块

DBStore 模块是 Go-Snap 框架的数据库 ORM 组件，基于 [GORM](https://gorm.io/) 构建，提供强大而灵活的数据库操作能力，支持多种数据库后端和企业级功能。

## 概述

DBStore 模块提供了一个统一的数据库操作接口，封装了 GORM 的强大功能，并与框架的其他组件无缝集成。它支持多种数据库、连接池管理、事务处理、数据迁移等功能，帮助开发者快速构建数据驱动的应用。

### 核心特性

- ✅ **多数据库支持** - MySQL, PostgreSQL, SQLite, SQL Server
- ✅ **ORM 功能** - 完整的对象关系映射
- ✅ **连接池管理** - 高性能的数据库连接池
- ✅ **事务支持** - 声明式和编程式事务管理
- ✅ **数据迁移** - 自动数据库结构迁移
- ✅ **查询构建器** - 灵活的查询构建
- ✅ **关联关系** - 一对一、一对多、多对多关系
- ✅ **软删除** - 逻辑删除支持
- ✅ **钩子函数** - 生命周期回调
- ✅ **批量操作** - 高效的批量插入/更新
- ✅ **分页查询** - 内置分页支持
- ✅ **预加载** - 解决 N+1 查询问题

## 快速开始

### 基础用法

```go
package main

import (
    "fmt"
    "github.com/guanzhenxing/go-snap/boot"
)

type User struct {
    ID    uint   `gorm:"primaryKey"`
    Name  string `gorm:"size:100;not null"`
    Email string `gorm:"uniqueIndex;size:100"`
    Age   int
}

func main() {
    app := boot.NewBoot().SetConfigPath("configs")
    application, _ := app.Initialize()
    
    // 获取数据库组件
    if dbComp, found := application.GetComponent("dbstore"); found {
        if dc, ok := dbComp.(*boot.DBStoreComponent); ok {
            db := dc.GetDB()
            
            // 自动迁移
            db.AutoMigrate(&User{})
            
            // 创建用户
            user := User{Name: "John Doe", Email: "john@example.com", Age: 30}
            result := db.Create(&user)
            if result.Error != nil {
                panic(result.Error)
            }
            
            // 查询用户
            var foundUser User
            db.First(&foundUser, user.ID)
            fmt.Printf("找到用户: %+v\n", foundUser)
        }
    }
}
```

### 配置文件

```yaml
# configs/application.yaml
dbstore:
  enabled: true
  driver: "mysql"
  dsn: "user:password@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
  max_idle_conns: 10
  max_open_conns: 100
  conn_max_lifetime: "1h"
  log_level: "info"
  slow_threshold: "200ms"
```

## 配置选项

### 基础配置

```yaml
dbstore:
  enabled: true                    # 是否启用数据库
  driver: "mysql"                  # 数据库驱动: mysql, postgres, sqlite, sqlserver
  dsn: ""                         # 数据源名称
  
  # 连接池配置
  max_idle_conns: 10              # 最大空闲连接数
  max_open_conns: 100             # 最大打开连接数
  conn_max_lifetime: "1h"         # 连接最大生存时间
  conn_max_idle_time: "30m"       # 连接最大空闲时间
  
  # 日志配置
  log_level: "info"               # 日志级别: silent, error, warn, info
  slow_threshold: "200ms"         # 慢查询阈值
  colorful: true                  # 是否彩色日志
  
  # 性能配置
  prepare_stmt: true              # 是否预编译语句
  disable_foreign_key_constraint: false # 是否禁用外键约束
  
  # 迁移配置
  auto_migrate: true              # 是否自动迁移
  migrate_tables: []              # 需要迁移的表结构
```

### 不同数据库的 DSN 格式

#### MySQL

```yaml
dbstore:
  driver: "mysql"
  dsn: "user:password@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
```

#### PostgreSQL

```yaml
dbstore:
  driver: "postgres"
  dsn: "host=localhost user=gorm password=gorm dbname=gorm port=9920 sslmode=disable TimeZone=Asia/Shanghai"
```

#### SQLite

```yaml
dbstore:
  driver: "sqlite"
  dsn: "gorm.db"
```

#### SQL Server

```yaml
dbstore:
  driver: "sqlserver"
  dsn: "sqlserver://gorm:LoremIpsum86@localhost:9930?database=gorm"
```

## API 参考

### 数据库操作

#### 创建记录

```go
// 创建单条记录
user := User{Name: "John", Email: "john@example.com"}
result := db.Create(&user)
if result.Error != nil {
    // 处理错误
}
fmt.Printf("新用户ID: %d\n", user.ID)

// 批量创建
users := []User{
    {Name: "Alice", Email: "alice@example.com"},
    {Name: "Bob", Email: "bob@example.com"},
}
db.Create(&users)

// 使用指定字段创建
db.Select("name", "age").Create(&user)

// 忽略指定字段创建
db.Omit("created_at").Create(&user)
```

#### 查询记录

```go
// 获取第一条记录
var user User
db.First(&user) // 按主键排序获取第一条
db.Take(&user)  // 获取一条记录，不排序
db.Last(&user)  // 按主键排序获取最后一条

// 按主键查询
db.First(&user, 10)                 // 查询主键为10的记录
db.First(&user, "10")               // 同上
db.Find(&users, []int{1, 2, 3})     // 查询主键为1,2,3的记录

// 条件查询
db.Where("name = ?", "John").First(&user)
db.Where("name <> ?", "John").Find(&users)
db.Where("name IN ?", []string{"John", "Alice"}).Find(&users)
db.Where("name LIKE ?", "%Jo%").Find(&users)
db.Where("age > ? AND created_at > ?", 18, lastWeek).Find(&users)

// 结构体条件查询
db.Where(&User{Name: "John", Age: 20}).First(&user)

// Map 条件查询
db.Where(map[string]interface{}{"name": "John", "age": 20}).Find(&users)

// 原生 SQL 查询
db.Where("name = ? AND age >= ?", "John", "22").Find(&users)
```

#### 更新记录

```go
// 更新单个字段
db.Model(&user).Update("name", "John Doe")

// 更新多个字段
db.Model(&user).Updates(User{Name: "John Doe", Age: 30})
db.Model(&user).Updates(map[string]interface{}{"name": "John Doe", "age": 30})

// 更新指定字段
db.Model(&user).Select("name").Updates(map[string]interface{}{"name": "John Doe", "age": 30})

// 批量更新
db.Where("age < ?", 18).Updates(User{Status: "minor"})

// 使用表达式更新
db.Model(&user).Update("age", gorm.Expr("age + ?", 1))
```

#### 删除记录

```go
// 删除记录
db.Delete(&user, 1)               // 按主键删除
db.Delete(&user, "10")            // 同上
db.Delete(&users, []int{1,2,3})   // 批量删除

// 条件删除
db.Where("age < ?", 18).Delete(&User{})

// 软删除（需要在模型中包含 gorm.DeletedAt 字段）
type User struct {
    ID        uint
    Name      string
    DeletedAt gorm.DeletedAt `gorm:"index"`
}

// 永久删除
db.Unscoped().Delete(&user)

// 查询包含软删除的记录
db.Unscoped().Where("age = 20").Find(&users)
```

### 高级查询

#### 预加载

```go
type User struct {
    ID      uint
    Name    string
    Orders  []Order
    Profile Profile
}

type Order struct {
    ID     uint
    UserID uint
    Amount float64
}

type Profile struct {
    ID     uint
    UserID uint
    Bio    string
}

// 预加载关联
db.Preload("Orders").Find(&users)
db.Preload("Orders").Preload("Profile").Find(&users)

// 条件预加载
db.Preload("Orders", "amount > ?", 100).Find(&users)

// 嵌套预加载
db.Preload("Orders.OrderItems").Find(&users)

// 自定义预加载
db.Preload("Orders", func(db *gorm.DB) *gorm.DB {
    return db.Order("orders.amount DESC")
}).Find(&users)
```

#### 连接查询

```go
// Left Join
db.Select("users.name, emails.email").
   Joins("left join emails on emails.user_id = users.id").
   Scan(&results)

// Inner Join
db.Joins("Profile").Find(&users)

// 带条件的 Join
db.Joins("Profile", db.Where(&Profile{Role: "admin"})).Find(&users)
```

#### 分组和聚合

```go
type Result struct {
    Date  time.Time
    Total int64
}

// 分组查询
db.Model(&Order{}).
   Select("date(created_at) as date, count(*) as total").
   Group("date(created_at)").
   Scan(&results)

// Having 条件
db.Model(&Order{}).
   Select("user_id, count(*) as total").
   Group("user_id").
   Having("count(*) > ?", 5).
   Scan(&results)
```

#### 子查询

```go
// 子查询
subQuery := db.Select("AVG(age)").Where("name LIKE ?", "name%").Table("users")
db.Select("AVG(age) as avgage").Group("name").Having("AVG(age) > (?)", subQuery).Find(&results)

// EXISTS 子查询
db.Where("EXISTS (?)", db.Select("1").Table("orders").Where("orders.user_id = users.id")).Find(&users)
```

### 事务处理

#### 手动事务

```go
// 开始事务
tx := db.Begin()

// 在事务中执行操作
if err := tx.Create(&user1).Error; err != nil {
    tx.Rollback()
    return err
}

if err := tx.Create(&user2).Error; err != nil {
    tx.Rollback()
    return err
}

// 提交事务
if err := tx.Commit().Error; err != nil {
    return err
}
```

#### 事务回调

```go
// 使用事务回调
err := db.Transaction(func(tx *gorm.DB) error {
    // 在事务中执行数据库操作
    if err := tx.Create(&user1).Error; err != nil {
        return err // 返回错误会自动回滚
    }
    
    if err := tx.Create(&user2).Error; err != nil {
        return err
    }
    
    // 返回 nil 提交事务
    return nil
})

if err != nil {
    // 处理事务错误
}
```

#### 嵌套事务

```go
db.Transaction(func(tx *gorm.DB) error {
    tx.Create(&user1)
    
    // 嵌套事务
    tx.Transaction(func(tx2 *gorm.DB) error {
        tx2.Create(&user2)
        return nil
    })
    
    return nil
})
```

### 钩子函数

```go
type User struct {
    ID        uint
    Name      string
    CreatedAt time.Time
    UpdatedAt time.Time
}

// BeforeCreate 钩子
func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
    // 在创建之前执行
    if u.Name == "" {
        return errors.New("name cannot be empty")
    }
    return
}

// AfterCreate 钩子
func (u *User) AfterCreate(tx *gorm.DB) (err error) {
    // 在创建之后执行
    fmt.Printf("用户 %s 已创建\n", u.Name)
    return
}

// BeforeUpdate 钩子
func (u *User) BeforeUpdate(tx *gorm.DB) (err error) {
    // 在更新之前执行
    return
}

// AfterFind 钩子
func (u *User) AfterFind(tx *gorm.DB) (err error) {
    // 在查询之后执行
    return
}
```

### 数据迁移

#### 自动迁移

```go
// 自动迁移模式
db.AutoMigrate(&User{})
db.AutoMigrate(&User{}, &Product{}, &Order{})

// 检查表是否存在
if !db.Migrator().HasTable(&User{}) {
    db.Migrator().CreateTable(&User{})
}

// 检查列是否存在
if !db.Migrator().HasColumn(&User{}, "Name") {
    db.Migrator().AddColumn(&User{}, "Name")
}
```

#### 手动迁移

```go
// 创建表
db.Migrator().CreateTable(&User{})

// 删除表
db.Migrator().DropTable(&User{})

// 重命名表
db.Migrator().RenameTable(&User{}, &UserInfo{})

// 添加列
db.Migrator().AddColumn(&User{}, "Age")

// 删除列
db.Migrator().DropColumn(&User{}, "Age")

// 修改列类型
db.Migrator().AlterColumn(&User{}, "Age")

// 创建索引
db.Migrator().CreateIndex(&User{}, "Name")
db.Migrator().CreateIndex(&User{}, "idx_user_name")

// 删除索引
db.Migrator().DropIndex(&User{}, "Name")
```

## 模型定义

### 基础模型

```go
// 基础模型结构
type BaseModel struct {
    ID        uint           `gorm:"primaryKey"`
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`
}

// 用户模型
type User struct {
    BaseModel
    Name     string `gorm:"size:100;not null;comment:用户姓名"`
    Email    string `gorm:"uniqueIndex;size:100;comment:邮箱地址"`
    Password string `gorm:"size:255;not null;comment:密码"`
    Age      int    `gorm:"comment:年龄"`
    Status   int    `gorm:"default:1;comment:状态 1:正常 0:禁用"`
}

// 自定义表名
func (User) TableName() string {
    return "users"
}
```

### 字段标签

```go
type User struct {
    ID        uint      `gorm:"primaryKey"`                                    // 主键
    Name      string    `gorm:"size:100;not null;index"`                     // 大小、非空、索引
    Email     string    `gorm:"uniqueIndex;size:100"`                        // 唯一索引
    Age       int       `gorm:"check:age > 0"`                               // 检查约束
    Birthday  time.Time `gorm:"default:CURRENT_TIMESTAMP"`                   // 默认值
    Status    int       `gorm:"default:1;comment:用户状态"`                    // 默认值和注释
    Profile   string    `gorm:"type:text"`                                   // 指定类型
    Amount    float64   `gorm:"precision:10;scale:2"`                        // 精度
    IsActive  bool      `gorm:"column:is_active;default:true"`               // 自定义列名
    Metadata  JSON      `gorm:"type:json"`                                   // JSON 类型
    CreatedAt time.Time `gorm:"<-:create"`                                   // 只在创建时写入
    UpdatedAt time.Time `gorm:"<-:update"`                                   // 只在更新时写入
    ReadOnly  string    `gorm:"->"`                                          // 只读
    Ignored   string    `gorm:"-"`                                           // 忽略字段
}
```

### 关联关系

#### 一对一关系

```go
// 一对一 - 用户和资料
type User struct {
    ID      uint
    Name    string
    Profile Profile
}

type Profile struct {
    ID     uint
    UserID uint // 外键
    Bio    string
    User   User // 反向引用
}

// 查询时预加载
db.Preload("Profile").Find(&users)
```

#### 一对多关系

```go
// 一对多 - 用户和订单
type User struct {
    ID     uint
    Name   string
    Orders []Order
}

type Order struct {
    ID     uint
    UserID uint // 外键
    Amount float64
    User   User
}

// 查询用户的所有订单
db.Preload("Orders").Find(&users)

// 创建用户和订单
user := User{
    Name: "John",
    Orders: []Order{
        {Amount: 100.0},
        {Amount: 200.0},
    },
}
db.Create(&user)
```

#### 多对多关系

```go
// 多对多 - 用户和角色
type User struct {
    ID    uint
    Name  string
    Roles []Role `gorm:"many2many:user_roles;"`
}

type Role struct {
    ID    uint
    Name  string
    Users []User `gorm:"many2many:user_roles;"`
}

// 关联数据
user := User{Name: "John"}
db.Create(&user)

role := Role{Name: "Admin"}
db.Create(&role)

// 建立关联
db.Model(&user).Association("Roles").Append(&role)

// 查询时预加载
db.Preload("Roles").Find(&users)
```

## 最佳实践

### 1. 模型设计

```go
// ✅ 好的做法：使用基础模型
type BaseModel struct {
    ID        uint           `gorm:"primaryKey"`
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`
}

type User struct {
    BaseModel
    Name  string `gorm:"size:100;not null;index;comment:用户姓名"`
    Email string `gorm:"uniqueIndex;size:100;comment:邮箱"`
}

// ✅ 实现 Tabler 接口自定义表名
func (User) TableName() string {
    return "users"
}

// ❌ 避免的做法：重复定义公共字段
type User struct {
    ID        uint `gorm:"primaryKey"`
    CreatedAt time.Time
    UpdatedAt time.Time
    Name      string
}
```

### 2. 查询优化

```go
// ✅ 使用索引优化查询
db.Where("email = ?", email).First(&user)

// ✅ 使用预加载避免 N+1 问题
db.Preload("Orders").Find(&users)

// ✅ 选择特定字段减少数据传输
db.Select("id", "name", "email").Find(&users)

// ✅ 使用限制减少返回数据
db.Limit(10).Offset(20).Find(&users)

// ❌ 避免的做法：没有索引的查询
db.Where("name = ?", name).Find(&users) // name 字段没有索引

// ❌ 避免的做法：在循环中查询关联数据
for _, user := range users {
    db.Where("user_id = ?", user.ID).Find(&orders) // N+1 问题
}
```

### 3. 事务使用

```go
// ✅ 使用事务保证数据一致性
err := db.Transaction(func(tx *gorm.DB) error {
    // 创建用户
    if err := tx.Create(&user).Error; err != nil {
        return err
    }
    
    // 创建订单
    order.UserID = user.ID
    if err := tx.Create(&order).Error; err != nil {
        return err
    }
    
    return nil
})

// ❌ 避免的做法：没有使用事务
db.Create(&user)
order.UserID = user.ID
db.Create(&order) // 如果失败，用户已创建但订单没有创建
```

### 4. 错误处理

```go
// ✅ 检查错误并处理
result := db.Create(&user)
if result.Error != nil {
    if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
        return errors.New("邮箱已存在")
    }
    return result.Error
}

// ✅ 检查影响的行数
result := db.Delete(&user)
if result.RowsAffected == 0 {
    return errors.New("用户不存在")
}

// ❌ 避免的做法：忽略错误
db.Create(&user) // 没有检查错误
```

### 5. 性能优化

```go
// ✅ 批量操作
users := []User{
    {Name: "Alice"},
    {Name: "Bob"},
}
db.CreateInBatches(users, 100)

// ✅ 使用原生 SQL 进行复杂查询
db.Raw("SELECT name FROM users WHERE age > ? AND status = ?", 18, 1).Scan(&names)

// ✅ 使用连接池配置
// 在配置文件中设置合适的连接池参数
```

## 集成示例

### 与缓存集成

```go
type UserService struct {
    db    *gorm.DB
    cache cache.Cache
}

func (s *UserService) GetUser(id uint) (*User, error) {
    cacheKey := fmt.Sprintf("user:%d", id)
    
    // 先从缓存获取
    if cached, found := s.cache.Get(context.Background(), cacheKey); found {
        if user, ok := cached.(*User); ok {
            return user, nil
        }
    }
    
    // 从数据库获取
    var user User
    if err := s.db.First(&user, id).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, errors.New("用户不存在")
        }
        return nil, err
    }
    
    // 缓存结果
    s.cache.Set(context.Background(), cacheKey, &user, time.Hour)
    
    return &user, nil
}

func (s *UserService) UpdateUser(user *User) error {
    err := s.db.Transaction(func(tx *gorm.DB) error {
        if err := tx.Save(user).Error; err != nil {
            return err
        }
        
        // 删除缓存
        cacheKey := fmt.Sprintf("user:%d", user.ID)
        s.cache.Delete(context.Background(), cacheKey)
        
        return nil
    })
    
    return err
}
```

### 与 Web 服务集成

```go
// 用户控制器
type UserController struct {
    userService *UserService
}

func (c *UserController) GetUser(ctx *gin.Context) {
    idStr := ctx.Param("id")
    id, err := strconv.ParseUint(idStr, 10, 32)
    if err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
        return
    }
    
    user, err := c.userService.GetUser(uint(id))
    if err != nil {
        if err.Error() == "用户不存在" {
            ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
            return
        }
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "内部服务器错误"})
        return
    }
    
    ctx.JSON(http.StatusOK, user)
}

func (c *UserController) CreateUser(ctx *gin.Context) {
    var user User
    if err := ctx.ShouldBindJSON(&user); err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    if err := c.userService.CreateUser(&user); err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    ctx.JSON(http.StatusCreated, user)
}
```

## 性能优化

### 1. 连接池配置

```yaml
dbstore:
  max_idle_conns: 10        # 空闲连接数
  max_open_conns: 100       # 最大连接数
  conn_max_lifetime: "1h"   # 连接最大生存时间
  conn_max_idle_time: "30m" # 连接最大空闲时间
```

### 2. 查询优化

```go
// 使用索引
type User struct {
    ID    uint   `gorm:"primaryKey"`
    Name  string `gorm:"index"`        // 单列索引
    Email string `gorm:"uniqueIndex"`  // 唯一索引
}

// 复合索引
type Order struct {
    ID       uint      `gorm:"primaryKey"`
    UserID   uint      `gorm:"index:idx_user_status"`
    Status   int       `gorm:"index:idx_user_status"`
    CreateAt time.Time `gorm:"index"`
}

// 使用合适的数据类型
type Product struct {
    ID          uint           `gorm:"primaryKey"`
    Name        string         `gorm:"size:100"`    // 指定合适的大小
    Description string         `gorm:"type:text"`   // 长文本使用 text
    Price       decimal.Decimal `gorm:"type:decimal(10,2)"` // 价格使用 decimal
}
```

### 3. 批量操作

```go
// 批量插入
users := make([]User, 1000)
for i := range users {
    users[i] = User{Name: fmt.Sprintf("User%d", i)}
}

// 分批插入
db.CreateInBatches(users, 100)

// 批量更新
db.Model(&User{}).Where("status = ?", 0).Updates(User{Status: 1})
```

## 故障排除

### 常见问题

#### 1. 连接数据库失败

**错误**: `dial tcp: connect: connection refused`

**解决方案**:
- 检查数据库服务是否运行
- 验证连接字符串格式
- 检查网络连通性
- 确认用户名密码正确

#### 2. 表不存在

**错误**: `Table 'database.table' doesn't exist`

**解决方案**:
- 运行自动迁移：`db.AutoMigrate(&Model{})`
- 手动创建表结构
- 检查表名是否正确

#### 3. 字段映射错误

**错误**: `Error scanning into destination`

**解决方案**:
- 检查结构体字段类型
- 确认数据库字段类型匹配
- 使用正确的 GORM 标签

### 调试技巧

```go
// 启用 SQL 日志
db.Debug().Find(&users)

// 获取生成的 SQL
sql := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
    return tx.Model(&User{}).Where("age > ?", 18).Find(&users)
})
fmt.Println(sql)

// 检查错误详情
result := db.Create(&user)
if result.Error != nil {
    fmt.Printf("错误: %v, 影响行数: %d\n", result.Error, result.RowsAffected)
}
```

## 参考资料

- [GORM 官方文档](https://gorm.io/)
- [Go-Snap Boot 模块](boot.md)
- [Go-Snap Config 模块](config.md)
- [Go-Snap 架构设计](../architecture.md) 