# FastAPI Best Architecture → Go Module 高可用复制与迁移实现文档

> 技术选型：Fiber v3、Zap、Swagger/OpenAPI、Dig、Asynq、Pond、Redis、DB 可配置、插件 Go module 化。  
> 目标：在**不修改现有前端**的前提下，将原 FastAPI 后端能力复制、迁移、演进为一个可复用的 Go 架构模块。  
> 本文基于前两份文档：
>
> 1. `fastapi-best-architecture Go 实现设计梳理.md`
> 2. `FastAPI Best Architecture → Go 前端零改动兼容规格.md`
>
> 本文不再重复完整源码评估，而是给出**可落地实现规格**。

---

## 0. 术语说明

你提到的 `asyncq`，Go 生态中主流、成熟、Redis 后端的分布式任务队列是：

```text
github.com/hibiken/asynq
```

本文后续统一使用 **Asynq**。如果你确实有另一个名为 `asyncq` 的内部库，可以把本文的 `task` 抽象层替换底层实现，但框架层接口不建议变化。

---

## 1. 总目标

### 1.1 最终形态

新架构不是一个具体业务项目，而是一个可复用 Go module：

```text
github.com/your-org/fba-go
```

业务项目只需要：

1. 引入该 Go module；
2. 传入 Redis、DB、Fiber options、日志配置、任务队列配置、插件配置；
3. import 或本地放置业务插件；
4. 运行代码生成；
5. 启动 HTTP / Worker / Scheduler。

业务项目主入口示例：

```go
package main

import (
    "context"
    "log"

    "github.com/gofiber/fiber/v3"
    "github.com/your-org/fba-go"
    "github.com/your-org/fba-go/core/config"
    "github.com/your-org/fba-go/core/plugin"

    // import 模式插件：用于自动扫描和生成注册代码
    _ "github.com/your-org/fba-plugin-admin"
    _ "github.com/your-org/fba-plugin-dict"
    _ "github.com/your-org/my-business-plugin-order"
)

func main() {
    app, err := fba.NewApplication(fba.Options{
        App: config.AppOptions{
            Environment: "prod",
            APIBasePath: "/api/v1",
            Timezone: "Asia/Shanghai",
        },
        Fiber: fiber.Config{
            BodyLimit: 20 * 1024 * 1024,
            Immutable: false,
        },
        Redis: config.RedisOptions{
            Mode: "sentinel",
            Addrs: []string{"10.0.0.1:26379", "10.0.0.2:26379"},
            MasterName: "mymaster",
            Password: "***",
            DB: 0,
        },
        Database: config.DatabaseOptions{
            Driver: "postgres",
            WriteDSN: "postgres://...",
            ReadDSN:  "postgres://...",
            MaxOpenConns: 100,
            MaxIdleConns: 20,
        },
        Logger: config.LoggerOptions{
            Level: "info",
            Encoding: "json",
        },
        Task: config.TaskOptions{
            RedisDB: 2,
            Concurrency: 32,
            Queues: map[string]int{
                "critical": 6,
                "default":  3,
                "low":      1,
            },
        },
        Plugins: plugin.Options{
            AutoInject: true,
            ImportScan: true,
            LocalScan: true,
            LocalDirs: []string{"./plugins"},
            PureDependencyPlugins: []string{
                "github.com/your-org/fba-plugin-oauth2-sdk",
            },
        },
    })
    if err != nil {
        log.Fatal(err)
    }

    if err := app.Run(context.Background()); err != nil {
        log.Fatal(err)
    }
}
```

### 1.2 迁移原则

当前前端不动，因此所有 Go 实现必须遵循：

| 项 | 要求 |
|---|---|
| 路由路径 | 完全一致，例如 `/api/v1/auth/login`、`/api/v1/sys/users` |
| HTTP Method | 完全一致 |
| Query 参数名 | 完全一致，保持 snake_case |
| Path 参数 | 完全一致，例如 `/{pk}`、`/{plugin}` |
| Body 字段 | 完全一致，保持 snake_case |
| 响应结构 | 默认 `{ code, msg, data }` |
| 分页结构 | `data.items / total / page / size / total_pages / links` |
| Cookie | refresh token 使用 `fba_refresh_token` |
| Header | access token 使用 `Authorization: Bearer <token>` |
| 日期格式 | 默认 `%Y-%m-%d %H:%M:%S` 风格字符串 |
| 权限码 | 完全保留，例如 `sys:user:del`、`dict:data:add` |
| Redis Key | 尽量保留原前缀，方便灰度、共存和迁移 |
| 表名 | 尽量保留原表名，降低前后端和数据迁移成本 |

Go 内部实现可以完全重构，但对前端暴露的 API 契约不能变。

---

## 2. 目标架构总览

### 2.1 分层结构

Go 版推荐保留原项目的分层思想，但用 Go idiom 表达：

```text
api / handler      HTTP 接口层，对齐 FastAPI router
schema / dto       请求/响应结构体，对齐 Pydantic schema
service            业务逻辑层
repo               数据访问层
model              数据库模型
middleware         Fiber 中间件
plugin             插件注册、发现、生成、装配
module             可复用框架核心
```

### 2.2 Go module 目录建议

```text
fba-go/
├── go.mod
├── README.md
├── core/
│   ├── app/                    # Application 生命周期、Run/Shutdown
│   ├── config/                 # 配置结构体与默认值
│   ├── di/                     # Dig 容器封装
│   ├── fiberx/                 # Fiber v3 封装、路由注册工具
│   ├── response/               # 统一响应、分页响应
│   ├── errors/                 # 统一错误模型
│   ├── middleware/             # RequestID、CORS、JWT、RBAC、日志等
│   ├── logger/                 # Zap 初始化与字段规范
│   ├── redisx/                 # Redis 封装与 Key 工具
│   ├── db/                     # DB Provider、事务、迁移抽象
│   ├── auth/                   # JWT、refresh token、session、密码策略
│   ├── rbac/                   # RBAC、权限码、数据权限
│   ├── pagination/             # page/size 分页与响应模型
│   ├── observability/          # Prometheus、OTel、health/readiness
│   ├── task/                   # Asynq client/server/scheduler 抽象
│   ├── pool/                   # Pond 封装，可选
│   ├── swagger/                # OpenAPI/Swagger 聚合和 Fiber handler
│   ├── plugin/                 # 插件接口、注册表、扫描、依赖排序
│   └── migration/              # 框架与插件迁移执行器
├── modules/
│   ├── admin/                  # 可选：内置 admin 插件
│   ├── dict/                   # 可选：内置 dict 插件
│   └── task/                   # 可选：内置 task 插件
├── cmd/
│   └── fbago/                 # 插件扫描、DI 生成、Swagger 聚合、契约检查
├── contracts/
│   ├── api.contract.yaml       # 前端兼容 API 契约
│   ├── response.contract.yaml  # 响应规范契约
│   └── redis.contract.yaml     # Redis Key 兼容契约
└── examples/
    ├── basic-app/
    ├── plugin-import-app/
    └── plugin-local-app/
```

### 2.3 业务项目目录建议

```text
my-backend/
├── go.mod
├── main.go
├── config.yaml
├── plugins.yaml
├── plugins/                    # 本地插件目录，可选
│   ├── order/
│   │   ├── go.mod 或 plugin.yaml
│   │   └── ...
│   └── payment/
├── internal/
│   └── generated/
│       ├── fba_plugins.gen.go
│       ├── fba_di.gen.go
│       ├── fba_routes.gen.go
│       ├── fba_tasks.gen.go
│       └── swagger.gen.json
└── docs/
    └── openapi.json
```

---

## 3. 核心 Go module API 设计

### 3.1 Application 对外接口

```go
package fba

type Application interface {
    HTTP() *fiber.App
    Container() *dig.Container
    Run(ctx context.Context) error
    RunHTTP(ctx context.Context) error
    RunWorker(ctx context.Context) error
    RunScheduler(ctx context.Context) error
    Shutdown(ctx context.Context) error
}

func NewApplication(opts Options) (Application, error)
```

### 3.2 Options 总配置

```go
type Options struct {
    App      config.AppOptions
    Fiber    fiber.Config
    Logger   config.LoggerOptions
    Database config.DatabaseOptions
    Redis    config.RedisOptions
    Auth     config.AuthOptions
    RBAC     config.RBACOptions
    Task     config.TaskOptions
    Swagger  config.SwaggerOptions
    Plugins  plugin.Options

    // 用于业务项目手动追加 Provider / Middleware / Hook
    Providers   []di.Provider
    Middlewares []fiber.Handler
    Hooks       app.Hooks
}
```

### 3.3 AppOptions

```go
type AppOptions struct {
    Name        string
    Version     string
    Environment string // dev/prod/test
    APIBasePath string // 默认 /api/v1
    Timezone    string // 默认 Asia/Shanghai

    StaticEnabled bool
    UploadDir     string

    DemoMode bool
}
```

### 3.4 DatabaseOptions

```go
type DatabaseOptions struct {
    Driver string // mysql/postgres

    WriteDSN string
    ReadDSN  string

    MaxOpenConns    int
    MaxIdleConns    int
    ConnMaxLifetime time.Duration
    ConnMaxIdleTime time.Duration

    AutoMigrate bool // 仅 dev 推荐 true；prod 必须 false
    MigrationLockKey string
}
```

推荐默认值：

```go
MaxOpenConns    = 100
MaxIdleConns    = 20
ConnMaxLifetime = 1h
ConnMaxIdleTime = 10m
AutoMigrate     = false in prod
```

### 3.5 RedisOptions

```go
type RedisOptions struct {
    Mode string // single/sentinel/cluster

    Addr  string
    Addrs []string

    Username string
    Password string
    DB       int

    MasterName string // sentinel

    PoolSize     int
    MinIdleConns int
    DialTimeout  time.Duration
    ReadTimeout  time.Duration
    WriteTimeout time.Duration

    KeyPrefix string // 默认 fba，可选
}
```

### 3.6 TaskOptions

```go
type TaskOptions struct {
    Enabled bool

    RedisMode string
    RedisAddr string
    RedisAddrs []string
    RedisDB int
    RedisPassword string
    RedisMasterName string

    Concurrency int
    Queues map[string]int

    SchedulerEnabled bool
    SchedulerLockKey string
    SchedulerLockTTL time.Duration
}
```

---

## 4. Fiber v3 HTTP 层设计

### 4.1 路由兼容要求

所有 API 必须保留原前缀：

```text
/api/v1
```

核心路由分组：

```text
/api/v1/auth
/api/v1/sys
/api/v1/logs
/api/v1/monitors
/api/v1/tasks
/api/v1/task-results
/api/v1/schedulers
/api/v1/dict-types
/api/v1/dict-datas
```

其中 `dict-types` 和 `dict-datas` 原本来自必需插件 `dict`，但为了前端零改动，Go 版建议将其作为默认内置插件或默认启用插件。

### 4.2 Fiber v3 上下文使用约束

Fiber 基于 fasthttp，Context 中的部分值会复用。框架内部必须遵守：

1. 不把 `c.Params()`、`c.Query()`、`c.Get()` 返回值的底层引用带出 handler；
2. 写入异步日志、任务、operation log 前，必须复制字符串；
3. 如果业务需要把请求字段放入 goroutine，必须使用复制后的值；
4. 如需绝对安全，可在用户传入 Fiber config 时允许开启 `Immutable: true`，但默认不建议，因为会影响性能。

### 4.3 中间件顺序

推荐顺序：

```text
1. RecoverMiddleware
2. RequestID / TraceID Middleware
3. ContextMiddleware
4. CORS Middleware
5. AccessMiddleware
6. MetricsMiddleware
7. I18nMiddleware
8. StateMiddleware(IP/UserAgent)
9. JWTAuthMiddleware
10. Route Handler
11. OperationLogMiddleware(after response)
12. ErrorHandler
```

Fiber 实现中可以通过包装 `Next()` 的方式实现 after-response 逻辑。

### 4.4 中间件职责

| 中间件 | 职责 |
|---|---|
| Recover | 捕获 panic，返回统一错误响应 |
| RequestID | 生成或读取 `X-Request-ID` |
| Context | 将 trace_id、ip、ua、user_id、permission 放入 request-scoped context |
| CORS | 保持原允许来源、凭证、暴露 header 逻辑 |
| Access | 记录开始时间、请求计数、in-progress metrics |
| Metrics | Prometheus 请求耗时、异常、响应计数 |
| I18n | 解析 `Accept-Language`，默认 `zh-CN` |
| State | 解析 IP、国家、省市、UA、OS、浏览器、设备 |
| JWTAuth | 解析 `Authorization: Bearer`，校验 JWT + Redis session |
| RBAC | 按 route 权限码校验用户菜单权限 |
| OperationLog | 记录 API 操作日志，异步批量入库 |

### 4.5 ErrorHandler

Fiber app 必须配置全局 ErrorHandler：

```go
fiber.New(fiber.Config{
    ErrorHandler: fbaerrors.FiberErrorHandler,
})
```

错误响应规则：

```json
{
  "code": 400,
  "msg": "请求参数非法: xxx",
  "data": null,
  "trace_id": "..."
}
```

正常成功响应不包含 `trace_id`，除非为了调试显式开启。

---

## 5. 统一响应协议

### 5.1 标准响应结构

```go
type Response[T any] struct {
    Code int    `json:"code"`
    Msg  string `json:"msg"`
    Data T      `json:"data"`
}
```

成功：

```json
{
  "code": 200,
  "msg": "成功",
  "data": {}
}
```

无 data：

```json
{
  "code": 200,
  "msg": "成功",
  "data": null
}
```

失败：

```json
{
  "code": 400,
  "msg": "失败",
  "data": null,
  "trace_id": "..."
}
```

### 5.2 分页响应结构

必须保持：

```json
{
  "code": 200,
  "msg": "成功",
  "data": {
    "items": [],
    "total": 0,
    "page": 1,
    "size": 20,
    "total_pages": 0,
    "links": {
      "first": "?page=1&size=20",
      "last": "?page=1&size=20",
      "self": "?page=1&size=20",
      "next": null,
      "prev": null
    }
  }
}
```

### 5.3 日期格式

所有 DTO 输出统一：

```text
YYYY-MM-DD HH:mm:ss
```

Go 中建议封装：

```go
type DateTime time.Time

func (d DateTime) MarshalJSON() ([]byte, error) {
    t := time.Time(d).In(appTimezone)
    return []byte(`"` + t.Format("2006-01-02 15:04:05") + `"`), nil
}
```

### 5.4 枚举输出

保持原数字/字符串值，不输出枚举名称。

```go
type StatusType int

const (
    StatusDisable StatusType = 0
    StatusEnable  StatusType = 1
)
```

### 5.5 特殊接口

`POST /api/v1/auth/login/swagger` 原实现返回非统一包裹：

```json
{
  "access_token": "...",
  "token_type": "Bearer",
  "user": {}
}
```

Go 版必须保留该行为。

---

## 6. API 契约保持策略

### 6.1 契约文件

框架 module 中保留：

```text
contracts/api.contract.yaml
contracts/response.contract.yaml
contracts/redis.contract.yaml
```

业务项目生成：

```text
internal/generated/api.routes.gen.go
internal/generated/api.contract.snapshot.json
```

### 6.2 契约测试

必须提供命令：

```bash
fbago contract test \
  --base-url http://127.0.0.1:8001 \
  --contract contracts/api.contract.yaml
```

测试内容：

1. 路由是否存在；
2. method 是否一致；
3. query/path/body 参数是否一致；
4. response JSON 字段是否一致；
5. 分页字段是否一致；
6. 错误响应是否包含 `trace_id`；
7. 登录、刷新、登出 cookie 行为是否一致；
8. 权限失败状态码和响应体是否一致；
9. 字典、菜单、权限码接口是否一致。

### 6.3 迁移验收优先级

第一优先级接口：

```text
GET  /api/v1/auth/captcha
POST /api/v1/auth/login
POST /api/v1/auth/refresh
POST /api/v1/auth/logout
GET  /api/v1/auth/codes
GET  /api/v1/sys/users/me
GET  /api/v1/sys/menus/sidebar
GET  /api/v1/dict-datas/type-codes/{code}
```

这几个接口决定前端是否能登录、加载菜单、加载按钮权限、加载字典数据。

---

## 7. 认证与 Token 兼容实现

### 7.1 JWT Payload

保持：

```json
{
  "session_uuid": "uuid",
  "exp": 1710000000,
  "sub": "10001"
}
```

字段说明：

| 字段 | 类型 | 说明 |
|---|---|---|
| session_uuid | string | 会话 UUID |
| exp | number | 过期 Unix timestamp |
| sub | string | 用户 ID 字符串 |

### 7.2 Access Token Redis Key

```text
fba:token:{user_id}:{session_uuid} = access_token
TTL = TOKEN_EXPIRE_SECONDS
```

### 7.3 Refresh Token Redis Key

```text
fba:refresh_token:{user_id}:{session_uuid} = refresh_token
TTL = TOKEN_REFRESH_EXPIRE_SECONDS
```

### 7.4 Refresh Cookie

必须保持：

```text
Name: fba_refresh_token
HttpOnly: true
```

生产建议增加：

```text
Secure: true
SameSite: Lax 或 None
```

但要确认现有前端跨域场景。如果前端和后端不同域且需要 cookie，必须使用：

```text
SameSite=None; Secure
```

### 7.5 多端登录逻辑

保持原语义：

| 用户字段 | 语义 |
|---|---|
| is_multi_login = false | 新登录会删除该用户旧 token |
| is_multi_login = true | 允许多端同时在线 |

### 7.6 用户缓存

```text
fba:user:{user_id} = user detail json
TTL = TOKEN_EXPIRE_SECONDS
```

用户信息、角色、菜单、部门、状态、权限变更后必须删除该缓存。

### 7.7 Go 接口设计

```go
type TokenService interface {
    CreateAccessToken(ctx context.Context, userID int64, multiLogin bool, extra map[string]any) (*AccessToken, error)
    CreateRefreshToken(ctx context.Context, userID int64, sessionUUID string, multiLogin bool) (*RefreshToken, error)
    Refresh(ctx context.Context, refreshToken string) (*NewToken, error)
    Revoke(ctx context.Context, userID int64, sessionUUID string) error
    Authenticate(ctx context.Context, token string) (*CurrentUser, error)
}
```

---

## 8. RBAC 与数据权限实现

### 8.1 权限码模型

保留菜单权限码：

```text
sys:user:del
sys:role:add
sys:role:menu:edit
data:scope:rule:edit
dict:data:add
sys:task:exec
```

### 8.2 路由权限声明

Go 中不要复制 FastAPI 的“依赖顺序敏感”问题。推荐路由声明：

```go
router.Delete("/:pk", userHandler.Delete,
    rbac.Require("sys:user:del"),
)
```

或：

```go
plugin.Route{
    Method: "DELETE",
    Path: "/sys/users/:pk",
    Handler: userHandler.Delete,
    Permission: "sys:user:del",
}
```

框架内部统一完成：

1. 设置 permission 到 request context；
2. 校验 JWT；
3. 校验 RBAC；
4. 执行 handler。

### 8.3 RBAC 规则

保持原语义：

1. 白名单接口跳过 RBAC；
2. 未认证请求返回 401；
3. 超级管理员跳过菜单权限校验；
4. 用户无启用角色返回 403；
5. 用户角色无菜单返回 403；
6. 非 GET/OPTIONS 请求，用户必须 `is_staff=true`；
7. 如果 route 设置了 permission，则必须存在于用户角色菜单的 `perms` 中。

### 8.4 数据权限规则

数据权限表：

```text
sys_data_scope
sys_data_rule
sys_role_data_scope
sys_data_scope_rule
```

规则转换：

| expression | SQL |
|---|---|
| 0 | = |
| 1 | != |
| 2 | > |
| 3 | >= |
| 4 | < |
| 5 | <= |
| 6 | IN |
| 7 | NOT IN |

operator：

| operator | 组合 |
|---|---|
| 0 | AND |
| 1 | OR |

模板变量：

```text
${user_id}
${dept_id}
${now}
```

Go 中建议设计：

```go
type DataPermissionFilter interface {
    Build(ctx context.Context, user *CurrentUser, models ...ModelRef) (SQLExpr, error)
}
```

如果使用 GORM：

```go
func (f *DataPermission) Apply(db *gorm.DB, user *CurrentUser, model string) *gorm.DB
```

如果使用 SQL builder：

```go
func (f *DataPermission) Where(user *CurrentUser, model string) (clause string, args []any, err error)
```

---

## 9. Redis 封装与高可用

### 9.1 Redis Client 模式

支持三种模式：

```text
single
sentinel
cluster
```

建议：

| 环境 | Redis 模式 |
|---|---|
| dev | single |
| test | single 或 container |
| prod 小规模 | sentinel |
| prod 大规模 | cluster，但要验证 Asynq lua 脚本兼容性 |

### 9.2 Key 兼容表

| 用途 | Key |
|---|---|
| Access Token | `fba:token:{user_id}:{session_uuid}` |
| Token Extra | `fba:token_extra_info:{user_id}:{session_uuid}` |
| Online Set | `fba:token_online` |
| Refresh Token | `fba:refresh_token:{user_id}:{session_uuid}` |
| JWT User Cache | `fba:user:{user_id}` |
| Login Captcha | `fba:login:captcha:{uuid}` |
| Login Failure | `fba:login:failure:{user_id}` |
| User Lock | `fba:user:lock:{user_id}` |
| Request Limiter | `fba:limiter:*` |
| Plugin State | `fba:plugin:{plugin}` |
| Plugin Changed | `fba:plugin:changed` |
| Cache PubSub | `fba:cache:invalidate` |
| Snowflake Node | `fba:snowflake:nodes:{datacenter}:{worker}` |
| Asynq | 建议 `fba:asynq:*` 或保留 Asynq 默认并配置隔离 DB |

### 9.3 Redis 封装接口

```go
type RedisClient interface {
    Get(ctx context.Context, key string) (string, error)
    Set(ctx context.Context, key string, value any, ttl time.Duration) error
    SetNX(ctx context.Context, key string, value any, ttl time.Duration) (bool, error)
    Del(ctx context.Context, keys ...string) error
    ScanPrefix(ctx context.Context, prefix string, count int64) ([]string, error)
    DeletePrefix(ctx context.Context, prefix string, exclude ...string) error
    Publish(ctx context.Context, channel string, payload string) error
    Subscribe(ctx context.Context, channels ...string) PubSub
}
```

### 9.4 HA 注意点

1. Redis 是 token、验证码、限流、任务队列、插件状态的核心依赖；
2. 生产必须部署高可用 Redis；
3. token/session 逻辑依赖 Redis，Redis 故障时 HTTP 可以降级为只读健康接口，其余认证接口应失败；
4. 对 `DeletePrefix` 必须使用 `SCAN`，禁止使用 `KEYS`；
5. 批量删除要分批，避免阻塞 Redis。

---

## 10. DB、表结构与迁移

### 10.1 表名兼容策略

为了前端和数据迁移最小成本，推荐保留表名：

```text
sys_user
sys_role
sys_menu
sys_dept
sys_user_role
sys_role_menu
sys_data_scope
sys_data_rule
sys_role_data_scope
sys_data_scope_rule
sys_login_log
sys_opera_log
sys_user_password_history
task_scheduler
task_result
task_set_result
dict_type
dict_data
```

如果 Go 版插件引入新表，使用命名规范：

```text
{plugin_id}_{business_name}
```

例如：

```text
order_order
order_order_item
payment_channel
```

### 10.2 迁移版本表

新增：

```sql
CREATE TABLE fba_schema_migrations (
    id BIGSERIAL PRIMARY KEY,
    scope VARCHAR(128) NOT NULL,
    version VARCHAR(128) NOT NULL,
    name VARCHAR(256) NOT NULL,
    checksum VARCHAR(128) NOT NULL,
    applied_at TIMESTAMP NOT NULL,
    execution_ms BIGINT NOT NULL,
    success BOOLEAN NOT NULL,
    error TEXT NULL,
    UNIQUE(scope, version)
);
```

scope 示例：

```text
core
plugin:admin
plugin:dict
plugin:order
```

### 10.3 迁移锁

生产执行 migration 必须加分布式锁：

```text
fba:migration:lock
```

流程：

```text
1. SetNX lock
2. 校验当前 migration 版本
3. 执行 pending migrations
4. 写入 fba_schema_migrations
5. 释放 lock
```

### 10.4 零停机迁移规范

禁止在一个版本内直接执行破坏性变更。

采用 expand / migrate / contract：

```text
版本 A：增加新字段、新表、新索引，老字段保留
版本 B：代码双读/双写或读新写新
版本 C：确认无老数据依赖后删除老字段
```

### 10.5 从 FastAPI 迁移到 Go

#### 方案 A：复用原数据库

优点：迁移最快。  
缺点：Go 模型必须完全适配原表结构。

适合：前端不动、后台尽快替换。

流程：

```text
1. 停止写入或进入维护窗口
2. 备份数据库
3. 部署 Go 服务到新端口
4. 运行 contract test
5. 灰度代理部分流量
6. 完全切流
```

#### 方案 B：新数据库 + 数据迁移

优点：可以修正历史设计。  
缺点：迁移复杂，不适合“前端不动”的第一阶段。

第一阶段不推荐。

### 10.6 Token 迁移策略

如果不要求用户无感：

```text
切换 Go 后端后强制重新登录
```

这是最稳妥方案。

如果要求用户无感：

1. JWT secret、algorithm、exp 规则保持一致；
2. Redis key 保持一致；
3. user cache JSON 结构保持一致；
4. refresh cookie 名称保持一致；
5. Go 版 `jwt_decode` 兼容原 token。

建议第一阶段允许强制重新登录，降低风险。

---

## 11. 日志系统：Zap

### 11.1 LoggerOptions

```go
type LoggerOptions struct {
    Level string // debug/info/warn/error
    Encoding string // json/console
    OutputPaths []string
    ErrorOutputPaths []string
    AccessLogPath string
    ErrorLogPath string
    Rotation RotationOptions
}
```

### 11.2 日志字段规范

所有日志统一包含：

```text
trace_id
request_id
user_id
username
method
path
status_code
cost_ms
ip
user_agent
service
version
environment
```

### 11.3 Zap 初始化

框架提供：

```go
func NewLogger(opts LoggerOptions) (*zap.Logger, error)
```

业务插件通过 DI 获取：

```go
type Params struct {
    dig.In
    Logger *zap.Logger
}
```

### 11.4 访问日志

AccessMiddleware 输出：

```json
{
  "level": "info",
  "trace_id": "...",
  "ip": "127.0.0.1",
  "method": "GET",
  "path": "/api/v1/sys/users",
  "status_code": 200,
  "cost_ms": 12.3
}
```

### 11.5 操作日志

保持表：

```text
sys_opera_log
```

处理方式：

```text
1. OperationLogMiddleware 收集请求信息
2. 敏感字段脱敏
3. 大字段截断
4. 写入 bounded channel
5. 后台 batch consumer 批量入库
```

### 11.6 操作日志队列设计

```go
type OperationLogQueueOptions struct {
    MaxSize int
    BatchSize int
    FlushInterval time.Duration
    OnFull string // block/drop_oldest/drop_new
}
```

推荐生产：

```text
OnFull = drop_new
```

避免操作日志队列阻塞主请求。

---

## 12. 监控与可观测性

### 12.1 Prometheus 指标

保留/新增指标：

```text
fba_request_in_progress
fba_request_total
fba_request_cost_time
fba_exception_total
fba_response_total
fba_db_pool_connections
fba_queue_size
fba_queue_exception_total
fba_task_enqueued_total
fba_task_processed_total
fba_task_failed_total
fba_plugin_enabled
```

### 12.2 HTTP 端点

```text
GET /metrics
GET /healthz
GET /readyz
```

`/readyz` 检查：

1. DB 是否可连接；
2. Redis 是否可连接；
3. migration 是否已完成；
4. 必需插件是否已注册；
5. Asynq worker/scheduler 如果启用，是否初始化完成。

### 12.3 监控 API 兼容

必须保留：

```text
GET /api/v1/monitors/server
GET /api/v1/monitors/redis
GET /api/v1/monitors/sessions
DELETE /api/v1/monitors/sessions/{pk}?session_uuid=xxx
```

Go 版 server 监控可用 `gopsutil` 或自行实现。

Redis 监控读取：

```text
INFO
DBSIZE
INFO commandstats
```

在线用户读取：

```text
fba:token:*
fba:token_extra_info:*
fba:token_online
```

---

## 13. Swagger / OpenAPI 实现

### 13.1 目标

1. 每个插件可以声明自己的 Swagger 注解或 OpenAPI fragment；
2. 框架可以聚合所有插件文档；
3. 最终暴露：

```text
GET /docs
GET /openapi
GET /swagger/doc.json
```

### 13.2 推荐实现

底层使用 `swaggo/swag` 生成 OpenAPI/Swagger 2.0。Fiber v3 适配有两种方式：

#### 方案 A：使用 fiber-swagger 中间件

如果版本支持 Fiber v3，则直接使用。

#### 方案 B：框架自带 Swagger UI Handler

更稳妥。框架自己提供：

```go
app.Get("/openapi", swagger.OpenAPIJSONHandler())
app.Get("/docs", swagger.UIHandler())
app.Get("/swagger/doc.json", swagger.OpenAPIJSONHandler())
```

这样不依赖 fiber-swagger 是否支持 v3。

### 13.3 插件 Swagger 聚合

插件提供：

```go
type SwaggerProvider interface {
    SwaggerSpec() swagger.Fragment
}
```

或通过代码生成读取注解：

```bash
fbago swagger scan \
  --plugins ./plugins \
  --imports \
  --out internal/generated/swagger.gen.json
```

聚合规则：

1. 合并 paths；
2. 合并 definitions/schemas；
3. 插件 schema 名加插件 ID 前缀，避免冲突；
4. 检测同 Method + Path 冲突；
5. 冲突时生成失败。

---

## 14. Dig 依赖注入设计

### 14.1 基本原则

Dig 只用于启动阶段构建依赖图，不允许在业务代码中当 service locator 使用。

正确：

```go
func NewUserHandler(p UserHandlerParams) *UserHandler
```

错误：

```go
func (h *UserHandler) Handle(c fiber.Ctx) error {
    container.Invoke(...)
}
```

### 14.2 核心 Provider

框架默认提供：

```text
*fiber.App
*zap.Logger
redis.UniversalClient
*gorm.DB 或 DBProvider
TokenService
PasswordService
RBACService
DataPermissionService
ResponseWriter
AsynqClient
AsynqServer
PondPool
PluginRegistry
MigrationRunner
```

### 14.3 命名依赖

读写库建议使用 dig name：

```go
container.Provide(NewWriteDB, dig.Name("db:write"))
container.Provide(NewReadDB, dig.Name("db:read"))
```

使用：

```go
type UserRepoParams struct {
    dig.In
    WriteDB *gorm.DB `name:"db:write"`
    ReadDB  *gorm.DB `name:"db:read" optional:"true"`
}
```

### 14.4 插件 Provider 注册

插件返回：

```go
type Provider func(*dig.Container) error
```

或：

```go
type Module interface {
    Providers() []any
}
```

框架负责：

```go
for _, p := range plugin.Providers() {
    container.Provide(p)
}
```

### 14.5 循环依赖检测

生成阶段和启动阶段都要检测。

1. 生成阶段检查 plugin dependency graph；
2. Dig Provide/Invoke 阶段检查 object graph；
3. 如果失败，输出可读错误和 DOT 图。

命令：

```bash
fbago di graph --out di.dot
```

---

## 15. 任务与定时任务：Asynq

### 15.1 目标

替代原 Celery，但前端接口不变：

```text
GET    /api/v1/tasks/registered
DELETE /api/v1/tasks/{task_id}/cancel
GET    /api/v1/task-results
GET    /api/v1/task-results/{pk}
DELETE /api/v1/task-results
GET    /api/v1/schedulers
GET    /api/v1/schedulers/all
GET    /api/v1/schedulers/{pk}
POST   /api/v1/schedulers
PUT    /api/v1/schedulers/{pk}
PUT    /api/v1/schedulers/{pk}/status
DELETE /api/v1/schedulers/{pk}
POST   /api/v1/schedulers/{pk}/execute
```

### 15.2 Task Registry

每个插件可以注册任务：

```go
type TaskDefinition struct {
    Type string
    Name string
    Queue string
    Handler asynq.Handler
}
```

插件接口：

```go
type TaskProvider interface {
    Tasks() []task.Definition
}
```

`GET /api/v1/tasks/registered` 返回所有注册任务：

```json
{
  "name": "发送邮件",
  "task": "email:send"
}
```

### 15.3 Asynq Client

```go
type TaskClient interface {
    Enqueue(ctx context.Context, taskType string, payload any, opts ...TaskOption) (*TaskInfo, error)
    Cancel(ctx context.Context, taskID string) error
}
```

### 15.4 Asynq Worker

启动：

```go
srv := asynq.NewServer(redisOpt, asynq.Config{
    Concurrency: opts.Task.Concurrency,
    Queues: opts.Task.Queues,
})

mux := asynq.NewServeMux()
for _, def := range registry.Tasks() {
    mux.Handle(def.Type, def.Handler)
}

srv.Run(mux)
```

### 15.5 定时任务

保持表：

```text
task_scheduler
```

字段兼容：

```text
name
task
args
kwargs
queue
exchange
routing_key
start_time
expire_time
expire_seconds
type
interval_every
interval_period
crontab
one_off
enabled
total_run_count
last_run_time
remark
```

Go scheduler 服务：

```go
type SchedulerService interface {
    Reload(ctx context.Context) error
    Execute(ctx context.Context, schedulerID int64) error
    Enable(ctx context.Context, schedulerID int64, enabled bool) error
}
```

### 15.6 Scheduler 高可用

多个 HTTP 实例和 worker 实例可以同时运行，但 scheduler 只能有一个 leader。

使用 Redis lease：

```text
fba:task:scheduler:leader
```

流程：

```text
1. SetNX leader key，TTL 30s
2. leader 每 10s 续租
3. leader 加载 task_scheduler
4. 任务变更时发布 reload event 或更新 version key
5. 非 leader 只监听，不执行调度
6. leader 崩溃后 TTL 到期，其他实例接管
```

### 15.7 Task Result 兼容

前端读取的是 `task_result`。Asynq 默认不直接写同样结构，所以需要框架适配：

1. enqueue 时写入一条 pending 记录；
2. handler 开始时更新 `STARTED`；
3. 成功后更新 `SUCCESS` 和 result；
4. 失败后更新 `FAILURE`、traceback；
5. 重试时更新 retries。

状态值建议映射：

| Asynq | 兼容输出 |
|---|---|
| pending | PENDING |
| active | STARTED |
| completed | SUCCESS |
| retry | RETRY |
| archived | FAILURE |
| scheduled | PENDING |

---

## 16. Pond 池使用规范

Pond 不替代 Asynq。二者定位不同：

| 组件 | 用途 |
|---|---|
| Asynq | 跨进程、跨机器、可重试、可持久化任务 |
| Pond | 单进程内限制并发、批量处理、异步日志、轻量后台工作 |

推荐使用 Pond 的地方：

1. 操作日志批量入库；
2. IP 地理位置解析并发限制；
3. 大批量文件处理；
4. 导出任务内部的局部并发；
5. 插件扫描/代码生成时的并行解析。

不推荐用 Pond 做：

1. 需要重试的业务任务；
2. 需要跨实例消费的任务；
3. 定时任务；
4. 长时间运行且需要恢复的任务。

框架提供：

```go
type PoolProvider interface {
    Pool(name string) pond.Pool
}
```

配置：

```yaml
pools:
  operation_log:
    max_workers: 4
    queue_size: 10000
  file_process:
    max_workers: 16
    queue_size: 1000
```

---

## 17. 插件系统总体设计

### 17.1 设计目标

1. 所有业务都通过插件实现；
2. 插件是 Go module 或本地 Go package；
3. 插件可以注册路由、服务、仓储、任务、定时任务、migration、swagger；
4. 插件支持依赖排序；
5. 插件支持启用/禁用；
6. 插件支持“纯依赖模式”，即只作为编译依赖，不自动注入；
7. 插件自动扫描支持 import 模式和本地目录模式；
8. 生产环境不使用 Go runtime plugin `.so`，而是 build-time 生成注册代码，保证稳定和可观测。

### 17.2 为什么不用 Go runtime plugin

Go 的 runtime plugin `.so` 在生产中有明显限制：

1. 平台限制；
2. Go 版本和依赖版本必须严格一致；
3. 容器构建复杂；
4. 类型边界脆弱；
5. 不利于高可用灰度。

因此推荐：

```text
扫描 → 生成 Go 注册代码 → 编译进二进制
```

### 17.3 插件接口

```go
package plugin

type Module interface {
    Meta() Meta
    Register(ctx Context) error
}

type Meta struct {
    ID          string
    Name        string
    Version     string
    Description string
    Author      string
    Tags        []string

    DependsOn []Dependency
    Provides  []string

    AutoInjectDefault bool
    PureDependencyDefault bool
}

type Dependency struct {
    ID      string
    Version string
    Optional bool
}
```

### 17.4 Context

```go
type Context interface {
    Container() *dig.Container
    Router() fiber.Router
    APIGroup() fiber.Router
    Logger() *zap.Logger
    Config() config.Config

    Provide(constructor any, opts ...dig.ProvideOption) error
    Route(route Route) error
    Task(task task.Definition) error
    Migration(m migration.Migration) error
    Swagger(fragment swagger.Fragment) error
}
```

### 17.5 路由定义

```go
type Route struct {
    Method string
    Path string
    Summary string
    Tags []string
    Permission string
    AuthRequired bool
    Handler fiber.Handler
}
```

示例：

```go
ctx.Route(plugin.Route{
    Method: "GET",
    Path: "/sys/users/me",
    Summary: "获取当前用户信息",
    Tags: []string{"系统用户"},
    AuthRequired: true,
    Handler: userHandler.Me,
})
```

### 17.6 Provider 定义

插件内部：

```go
func (m *Module) Register(ctx plugin.Context) error {
    ctx.Provide(NewUserRepo)
    ctx.Provide(NewUserService)
    ctx.Provide(NewUserHandler)

    ctx.Route(plugin.Route{...})
    ctx.Task(task.Definition{...})
    ctx.Migration(migration.NewSQL(...))

    return nil
}
```

---

## 18. 插件扫描与代码生成

### 18.1 扫描模式

支持三种来源：

| 来源 | 说明 |
|---|---|
| import scan | 扫描业务项目 import 的插件包 |
| local scan | 扫描配置的 `plugins` 目录 |
| manifest | 显式读取 `plugins.yaml` |

最终三者合并，并去重。

### 18.2 import scan

业务项目通过 blank import 标记插件：

```go
import (
    _ "github.com/your-org/fba-plugin-dict"
    _ "github.com/your-org/my-plugin-order"
)
```

`fbago` 运行：

```bash
fbago plugin scan \
  --mode imports \
  --module . \
  --out internal/generated/fba_plugins.gen.go
```

扫描策略：

1. 执行 `go list -deps -json ./...`；
2. 找出带 `fba.plugin.yaml` 或导出 `FBAPlugin()` 的包；
3. 读取 plugin meta；
4. 生成 import 和 registry 代码。

插件包要求导出：

```go
func FBAPlugin() plugin.Module
```

### 18.3 local scan

配置：

```yaml
plugins:
  local_dirs:
    - ./plugins
```

目录结构：

```text
plugins/order/plugin.yaml
plugins/order/go.mod
plugins/order/plugin.go
```

plugin.yaml：

```yaml
id: order
name: 订单插件
module: github.com/your-org/my-backend/plugins/order
entry: FBAPlugin
auto_inject: true
pure_dependency: false
depends_on:
  - id: dict
    version: ">=0.0.1"
```

扫描命令：

```bash
fbago plugin scan \
  --mode local \
  --plugins-dir ./plugins \
  --out internal/generated/fba_plugins.gen.go
```

### 18.4 manifest 模式

plugins.yaml：

```yaml
plugins:
  - id: admin
    module: github.com/your-org/fba-plugin-admin
    mode: auto
  - id: dict
    module: github.com/your-org/fba-plugin-dict
    mode: auto
  - id: oauth2-sdk
    module: github.com/your-org/fba-plugin-oauth2-sdk
    mode: pure_dependency
  - id: order
    module: ./plugins/order
    mode: auto
```

mode：

| mode | 含义 |
|---|---|
| auto | 自动注入 Provider、Route、Task、Migration |
| disabled | 扫描到但不注册 |
| pure_dependency | 只作为依赖，不自动注入 |

### 18.5 生成代码示例

`internal/generated/fba_plugins.gen.go`：

```go
// Code generated by fbago. DO NOT EDIT.
package generated

import (
    "github.com/your-org/fba-go/core/plugin"

    admin "github.com/your-org/fba-plugin-admin"
    dict "github.com/your-org/fba-plugin-dict"
    order "github.com/your-org/my-backend/plugins/order"
)

func RegisterPlugins(reg *plugin.Registry) error {
    reg.Add(admin.FBAPlugin(), plugin.ModeAuto)
    reg.Add(dict.FBAPlugin(), plugin.ModeAuto)
    reg.Add(order.FBAPlugin(), plugin.ModeAuto)
    return nil
}
```

`internal/generated/fba_di.gen.go`：

```go
// Code generated by fbago. DO NOT EDIT.
package generated

import (
    "github.com/your-org/fba-go/core/plugin"
)

func RegisterPluginDI(ctx plugin.Context) error {
    return ctx.Registry().RegisterAll(ctx)
}
```

### 18.6 生成流程

推荐业务项目 Makefile：

```makefile
generate:
    fbago plugin scan --mode imports,local,manifest --plugins-dir ./plugins --manifest ./plugins.yaml --out internal/generated/fba_plugins.gen.go
    fbago plugin graph --out internal/generated/plugin_graph.dot
    fbago swagger scan --out docs/openapi.json
    fbago contract snapshot --out internal/generated/api.contract.snapshot.json
```

### 18.7 插件依赖排序

规则：

1. 所有插件构造成 DAG；
2. 检查循环依赖；
3. 检查缺失依赖；
4. pure dependency 可作为依赖满足项，但不自动注册路由；
5. disabled 插件不能满足强依赖，除非依赖声明 optional；
6. 生成排序后的注册顺序。

错误示例：

```text
plugin dependency cycle: order -> payment -> order
```

### 18.8 插件运行时启停

分两层：

| 层级 | 说明 |
|---|---|
| 编译层 | 插件是否被编译进二进制，由生成代码决定 |
| 运行层 | 插件是否启用，由 DB/Redis 状态决定 |

运行层状态：

```text
fba:plugin:{plugin_id}
```

路由如果插件被禁用，应返回：

```json
{
  "code": 500,
  "msg": "插件 xxx 未启用，请联系系统管理员",
  "data": null,
  "trace_id": "..."
}
```

### 18.9 纯依赖插件

纯依赖插件适用于：

1. 只提供 SDK；
2. 只提供公共 DTO；
3. 只提供工具函数；
4. 被其他插件 import 使用；
5. 不希望自动注册路由、任务、migration。

配置：

```yaml
plugins:
  - id: oauth2-sdk
    module: github.com/your-org/fba-plugin-oauth2-sdk
    mode: pure_dependency
```

框架行为：

```text
1. 可被扫描识别
2. 可参与依赖满足
3. 不调用 Register(ctx)
4. 不注册路由
5. 不注册 migration
6. 不注册 task
```

---

## 19. 插件作为业务唯一载体

### 19.1 原则

后续所有业务模块都应该通过插件提供，不直接写在主项目里。

主项目只负责：

1. 选择插件；
2. 配置插件；
3. 生成注册代码；
4. 启动框架。

业务逻辑都在插件中：

```text
user plugin
role plugin
menu plugin
dict plugin
order plugin
payment plugin
report plugin
```

### 19.2 插件内部分层

```text
plugin-order/
├── go.mod
├── plugin.yaml
├── plugin.go
├── api/
│   └── order_handler.go
├── dto/
│   └── order_dto.go
├── service/
│   └── order_service.go
├── repo/
│   └── order_repo.go
├── model/
│   └── order.go
├── task/
│   └── order_tasks.go
├── migration/
│   ├── 0001_init.up.sql
│   └── 0001_init.down.sql
└── docs/
    └── swagger.yaml
```

### 19.3 插件主入口

```go
package order

func FBAPlugin() plugin.Module {
    return &OrderPlugin{}
}

type OrderPlugin struct{}

func (p *OrderPlugin) Meta() plugin.Meta {
    return plugin.Meta{
        ID: "order",
        Name: "订单插件",
        Version: "0.1.0",
        DependsOn: []plugin.Dependency{
            {ID: "admin", Version: ">=0.1.0"},
            {ID: "dict", Version: ">=0.1.0", Optional: true},
        },
        AutoInjectDefault: true,
    }
}

func (p *OrderPlugin) Register(ctx plugin.Context) error {
    ctx.Provide(NewOrderRepo)
    ctx.Provide(NewOrderService)
    ctx.Provide(NewOrderHandler)

    ctx.Route(plugin.Route{
        Method: "GET",
        Path: "/orders",
        Summary: "分页获取订单",
        Tags: []string{"订单"},
        Permission: "order:list",
        AuthRequired: true,
        Handler: func(c fiber.Ctx) error {
            return ctx.InvokeHandler[OrderHandler](func(h *OrderHandler) fiber.Handler {
                return h.List
            })(c)
        },
    })

    ctx.Task(task.Definition{
        Type: "order:timeout-close",
        Name: "订单超时关闭",
        Handler: NewTimeoutCloseTaskHandler(ctx),
    })

    ctx.Migration(migration.SQLFile("order", "migration/0001_init.up.sql"))
    return nil
}
```

---

## 20. 高可用部署设计

### 20.1 进程拆分

推荐拆成三个二进制或同一二进制不同 mode：

```text
fba-server      HTTP API
fba-worker      Asynq Worker
fba-scheduler   定时任务调度器
```

启动命令：

```bash
myapp serve
myapp worker
myapp scheduler
```

或：

```bash
myapp all
```

生产建议分开部署。

### 20.2 HTTP 高可用

```text
Nginx / Ingress / LB
        ↓
HTTP Pod 1
HTTP Pod 2
HTTP Pod N
        ↓
Redis HA + DB HA
```

HTTP 服务必须无状态：

1. session/token 存 Redis；
2. 用户缓存存 Redis；
3. 上传文件建议使用对象存储，不建议本地磁盘；
4. 操作日志写 DB；
5. 插件状态存 Redis/DB。

### 20.3 Worker 高可用

Asynq worker 可多实例水平扩展：

```text
Worker 1
Worker 2
Worker N
```

要求：

1. task handler 必须幂等；
2. 任务 payload 必须可序列化；
3. 对外部副作用操作必须有业务幂等键；
4. 失败任务依赖 Asynq retry；
5. 长任务设置 timeout/deadline。

### 20.4 Scheduler 高可用

scheduler 多实例部署，但只有 leader 执行。

```text
Scheduler 1  ─┐
Scheduler 2   ├─ Redis lease: fba:task:scheduler:leader
Scheduler N  ─┘
```

### 20.5 DB 高可用

推荐：

| 场景 | 设计 |
|---|---|
| 小规模 | 单主 DB + 自动备份 |
| 中规模 | 主从读写分离 |
| 大规模 | 主从 + 连接池 + 慢查询监控 |

框架支持：

```go
WriteDB `name:"db:write"`
ReadDB  `name:"db:read" optional:"true"`
```

默认所有写走 write DB，读接口可配置走 read DB。

### 20.6 Redis 高可用

优先推荐 Sentinel。

如果使用 Redis Cluster，必须验证：

1. Asynq lua 脚本兼容；
2. key hash tag 策略；
3. scan prefix 行为；
4. pipeline/mget 跨 slot 行为。

### 20.7 灰度发布

推荐流程：

```text
1. 构建包含固定插件集的新版本
2. 执行 migration plan
3. 部署 Go 服务 shadow 环境
4. contract test
5. 小流量 canary
6. 对比关键接口响应
7. 扩大流量
8. 完全切流
9. 保留旧 FastAPI 服务一段时间用于回滚
```

---

## 21. 与原 FastAPI 的功能映射

| FastAPI 功能 | Go 实现 |
|---|---|
| FastAPI app | Fiber v3 app |
| Pydantic schema | Go DTO struct + validator |
| SQLAlchemy async | DB abstraction，推荐 GORM/sqlx 二选一 |
| RedisCli | redisx UniversalClient 封装 |
| JwtAuthMiddleware | Fiber JWTAuthMiddleware |
| RequestPermission + DependsRBAC | route Permission + RBAC middleware |
| OperationLogMiddleware | Fiber OperationLog middleware + batch writer |
| Loguru | Zap |
| Celery | Asynq |
| Celery Beat | Asynq Scheduler + DB scheduler + Redis leader lock |
| Plugin TOML | plugin.yaml + Go module metadata |
| Dynamic import | fbago build-time code generation |
| Prometheus | prometheus/client_golang |
| OTel | otel-go |
| Swagger | swaggo/swag + framework Swagger handler |

---

## 22. 兼容接口实现路线图

### Phase 0：契约冻结

产物：

```text
contracts/api.contract.yaml
contracts/response.contract.yaml
contracts/redis.contract.yaml
```

冻结内容：

1. API 路由；
2. DTO 字段；
3. 响应格式；
4. Cookie/header；
5. Redis key；
6. 表结构。

### Phase 1：核心框架 module

实现：

1. Fiber v3 app；
2. Dig container；
3. Zap logger；
4. Redis provider；
5. DB provider；
6. response/error；
7. middleware skeleton；
8. swagger handler；
9. plugin registry；
10. fbago 初版。

### Phase 2：admin/dict 兼容插件

实现：

1. auth；
2. captcha；
3. users；
4. roles；
5. menus；
6. depts；
7. data-rules；
8. data-scopes；
9. files；
10. dict-types；
11. dict-datas。

### Phase 3：日志、监控、任务

实现：

1. login logs；
2. operation logs；
3. server monitor；
4. redis monitor；
5. sessions monitor；
6. task registered；
7. task result；
8. task scheduler。

### Phase 4：插件系统增强

实现：

1. import scan；
2. local scan；
3. manifest；
4. pure dependency；
5. plugin dependency graph；
6. plugin swagger aggregation；
7. plugin migration；
8. plugin runtime status。

### Phase 5：高可用和灰度

实现：

1. readiness；
2. migration lock；
3. scheduler leader lock；
4. Asynq multi-worker；
5. Redis sentinel；
6. DB read/write；
7. contract test；
8. canary deploy。

---

## 23. 代码生成工具 fbago

### 23.1 命令设计

```bash
fbago init <module>
fbago template list
fbago plugin scan
fbago plugin graph
fbago di generate
fbago swagger scan
fbago contract snapshot
fbago contract test
fbago migration plan
fbago migration apply
```

### 23.2 init

```bash
fbago init github.com/your-org/my-backend
fbago init github.com/your-org/my-backend --template basic
fbago init github.com/your-org/my-backend --template ../fba-go-template/admin
fbago init github.com/your-org/my-backend --template github.com/your-org/fba-go-template/admin@v0.1.0
```

`init` 用于创建项目脚手架，语义对齐 `go mod init`：调用方传入 Go module name，工具按模板生成 `go.mod`、`Makefile`、`cmd/api`、`.env` 和项目内业务模块目录 `internal/app`。默认模板是内置 `basic`；完整 admin starter template 应维护在独立模板仓库中，并通过本地路径或 remote Git template spec 传给 `--template`。

`internal/app` 是用户项目自己的业务代码位置，admin、dict、config、notice、订单、支付等可修改业务模块都应优先放在这里；远程 plugin 更适合承载邮件、OAuth2、对象存储、任务队列等通用能力。

`basic` 模板的 `Makefile` 至少提供 `tidy`、`test`、`run`、`dev`、`build`、`clean`，并将 Go build cache 固定在项目目录内，避免本机或沙箱缓存权限影响初始化后的第一轮验证。

本地路径模板必须是一个完整模板目录，可以使用 `[[ .Module ]]` 渲染 module name，文件名以 `.tmpl` 结尾时会渲染并去掉后缀；`env.tmpl` 和 `gitignore.tmpl` 分别输出为 `.env` 和 `.gitignore`。本地模板路径面向“可直接运行、可独立测试”的模板仓库，因此模板仓库可以保留自己的真实 `go.mod`，同时提供 `go.mod.tmpl` 作为生成项目的 `go.mod`。

可运行模板仓库建议在根目录放置 `.fbago-template.yaml`：

```yaml
module: github.com/your-org/fba-go-template/admin
```

`fbago init` 会把该模板 module path 替换为用户传入的目标 module path，并且不会把 `.fbago-template.yaml` 复制到新项目。这样模板仓库源码可以直接使用 `github.com/your-org/fba-go-template/admin/internal/app/...` import 并通过 `go test ./...`，生成项目后这些 import 会变成 `github.com/your-org/my-backend/internal/app/...`。

初始化时会跳过 `.git`、`.hg`、`.svn`、`.codegraph`、`.cache`、`bin`、`tmp`、`node_modules` 目录，以及 `.DS_Store`、`Thumbs.db` 文件，避免把仓库元数据和本地构建产物复制进新项目。

remote Git template spec 支持两种形式：

```bash
# 简写：前三段是 Git 仓库，后续路径是模板子目录
fbago init github.com/your-org/my-backend --template github.com/your-org/fba-go-template/admin@v0.1.0

# 显式 Git URL：用 // 分隔仓库 URL 和模板子目录
fbago init github.com/your-org/my-backend --template https://github.com/your-org/fba-go-template.git//admin@v0.1.0
```

`@ref` 可指定 tag、branch 或 Git 可识别的 ref；不指定时使用仓库默认分支。`FBAGO_TEMPLATE_CACHE_DIR` 可指定 clone 临时 checkout 的父目录。

### 23.3 template list

```bash
fbago template list
```

输出当前内置模板名，例如 `basic`。外部模板仓库或本地路径模板不在此列表中。

### 23.4 plugin scan

```bash
fbago plugin scan \
  --mode imports,local,manifest \
  --plugins-dir ./plugins \
  --manifest ./plugins.yaml \
  --out internal/generated/fba_plugins.gen.go
```

输出：

```text
internal/generated/fba_plugins.gen.go
internal/generated/plugin_graph.dot
internal/generated/plugin_manifest.lock
```

### 23.5 di generate

```bash
fbago di generate \
  --plugins internal/generated/plugin_manifest.lock \
  --out internal/generated/fba_di.gen.go
```

### 23.6 swagger scan

```bash
fbago swagger scan \
  --plugins internal/generated/plugin_manifest.lock \
  --out docs/openapi.json
```

### 23.7 contract test

```bash
fbago contract test \
  --base-url http://127.0.0.1:8001 \
  --contract contracts/api.contract.yaml \
  --login-user admin \
  --login-password '***'
```

---

## 24. 安全设计

### 24.1 JWT

1. secret 必须通过环境变量或 secret manager 注入；
2. 支持密钥轮换：`kid` 可选；
3. Redis session 校验必须开启；
4. logout 必须删除 access token 和 refresh token；
5. password reset 必须删除用户全部 token。

### 24.2 Cookie

生产建议：

```text
HttpOnly = true
Secure = true
SameSite = Lax 或 None
Path = /
```

### 24.3 上传文件

为了前端兼容保留：

```text
POST /api/v1/sys/files/upload
返回 /static/upload/{filename}
```

但生产建议：

1. 上传到对象存储；
2. 本地 URL 做反向代理；
3. 校验扩展名和 MIME；
4. 限制文件大小；
5. 文件名去路径穿越；
6. 可选病毒扫描。

### 24.4 操作日志脱敏

递归脱敏字段：

```text
password
old_password
new_password
confirm_password
token
access_token
refresh_token
secret
client_secret
```

---

## 25. Go DTO 命名与 JSON 字段规范

Go 结构体字段用 PascalCase，JSON 必须 snake_case：

```go
type AuthLoginParam struct {
    Username string  `json:"username" validate:"required"`
    Password string  `json:"password" validate:"required"`
    UUID     *string `json:"uuid"`
    Captcha  *string `json:"captcha"`
}
```

禁止输出 camelCase。

### 25.1 常用 DTO

#### Login

```go
type GetLoginToken struct {
    AccessToken                string         `json:"access_token"`
    AccessTokenExpireTime      DateTime       `json:"access_token_expire_time"`
    SessionUUID                string         `json:"session_uuid"`
    PasswordExpireDaysRemaining *int          `json:"password_expire_days_remaining"`
    User                       GetUserInfo    `json:"user"`
}
```

#### PageData

```go
type PageData[T any] struct {
    Items      []T       `json:"items"`
    Total      int64     `json:"total"`
    Page       int       `json:"page"`
    Size       int       `json:"size"`
    TotalPages int       `json:"total_pages"`
    Links      PageLinks `json:"links"`
}
```

---

## 26. 关键接口清单保留

完整字段以前端兼容契约文档为准。本节列出必须保留的核心路径。

### 26.1 Auth

```text
POST /api/v1/auth/login/swagger
POST /api/v1/auth/login
GET  /api/v1/auth/codes
POST /api/v1/auth/refresh
POST /api/v1/auth/logout
GET  /api/v1/auth/captcha
```

### 26.2 Sys

```text
GET    /api/v1/sys/users/me
GET    /api/v1/sys/users/{pk}
GET    /api/v1/sys/users/{pk}/roles
GET    /api/v1/sys/users
POST   /api/v1/sys/users
PUT    /api/v1/sys/users/{pk}
PUT    /api/v1/sys/users/{pk}/permissions
PUT    /api/v1/sys/users/me/password
PUT    /api/v1/sys/users/{pk}/password
PUT    /api/v1/sys/users/me/nickname
PUT    /api/v1/sys/users/me/avatar
PUT    /api/v1/sys/users/me/email
DELETE /api/v1/sys/users/{pk}

GET    /api/v1/sys/roles/all
GET    /api/v1/sys/roles/{pk}/menus
GET    /api/v1/sys/roles/{pk}/scopes
GET    /api/v1/sys/roles/{pk}
GET    /api/v1/sys/roles
POST   /api/v1/sys/roles
PUT    /api/v1/sys/roles/{pk}
PUT    /api/v1/sys/roles/{pk}/menus
PUT    /api/v1/sys/roles/{pk}/scopes
DELETE /api/v1/sys/roles

GET    /api/v1/sys/menus/sidebar
GET    /api/v1/sys/menus/{pk}
GET    /api/v1/sys/menus
POST   /api/v1/sys/menus
PUT    /api/v1/sys/menus/{pk}
DELETE /api/v1/sys/menus/{pk}

GET    /api/v1/sys/depts/{pk}
GET    /api/v1/sys/depts
POST   /api/v1/sys/depts
PUT    /api/v1/sys/depts/{pk}
DELETE /api/v1/sys/depts/{pk}

GET    /api/v1/sys/data-rules/models
GET    /api/v1/sys/data-rules/models/{model}/columns
GET    /api/v1/sys/data-rules/value-template-variables
GET    /api/v1/sys/data-rules/all
GET    /api/v1/sys/data-rules/{pk}
GET    /api/v1/sys/data-rules
POST   /api/v1/sys/data-rules
PUT    /api/v1/sys/data-rules/{pk}
DELETE /api/v1/sys/data-rules

GET    /api/v1/sys/data-scopes/all
GET    /api/v1/sys/data-scopes/{pk}
GET    /api/v1/sys/data-scopes/{pk}/rules
GET    /api/v1/sys/data-scopes
POST   /api/v1/sys/data-scopes
PUT    /api/v1/sys/data-scopes/{pk}
PUT    /api/v1/sys/data-scopes/{pk}/rules
DELETE /api/v1/sys/data-scopes

POST   /api/v1/sys/files/upload

GET    /api/v1/sys/plugins
GET    /api/v1/sys/plugins/changed
POST   /api/v1/sys/plugins
DELETE /api/v1/sys/plugins/{plugin}
PUT    /api/v1/sys/plugins/{plugin}/status
GET    /api/v1/sys/plugins/{plugin}
```

### 26.3 Logs

```text
GET    /api/v1/logs/login
DELETE /api/v1/logs/login
DELETE /api/v1/logs/login/all

GET    /api/v1/logs/opera
DELETE /api/v1/logs/opera
DELETE /api/v1/logs/opera/all
```

### 26.4 Monitors

```text
GET    /api/v1/monitors/server
GET    /api/v1/monitors/redis
GET    /api/v1/monitors/sessions
DELETE /api/v1/monitors/sessions/{pk}
```

### 26.5 Tasks

```text
GET    /api/v1/tasks/registered
DELETE /api/v1/tasks/{task_id}/cancel
GET    /api/v1/task-results/{pk}
GET    /api/v1/task-results
DELETE /api/v1/task-results
GET    /api/v1/schedulers/all
GET    /api/v1/schedulers/{pk}
GET    /api/v1/schedulers
POST   /api/v1/schedulers
PUT    /api/v1/schedulers/{pk}
PUT    /api/v1/schedulers/{pk}/status
DELETE /api/v1/schedulers/{pk}
POST   /api/v1/schedulers/{pk}/execute
```

### 26.6 Dict

```text
GET    /api/v1/dict-types/all
GET    /api/v1/dict-types/{pk}
GET    /api/v1/dict-types
POST   /api/v1/dict-types
PUT    /api/v1/dict-types/{pk}
DELETE /api/v1/dict-types

GET    /api/v1/dict-datas/all
GET    /api/v1/dict-datas/{pk}
GET    /api/v1/dict-datas/type-codes/{code}
GET    /api/v1/dict-datas
POST   /api/v1/dict-datas
PUT    /api/v1/dict-datas/{pk}
DELETE /api/v1/dict-datas
```

---

## 27. 迁移实施清单

### 27.1 开发期

```text
[ ] 建立 fba-go module
[ ] 实现 response/error/pagination
[ ] 实现 Fiber app 与 middleware
[ ] 实现 Zap logger
[ ] 实现 Redis provider
[ ] 实现 DB provider
[ ] 实现 JWT + Redis token
[ ] 实现 RBAC
[ ] 实现 plugin registry
[ ] 实现 fbago plugin scan
[ ] 实现 swagger handler
[ ] 实现 contract test
```

### 27.2 兼容期

```text
[ ] 完成 auth 接口
[ ] 完成 user/role/menu/dept 接口
[ ] 完成 dict 接口
[ ] 完成 logs 接口
[ ] 完成 monitors 接口
[ ] 完成 tasks 接口
[ ] 前端本地联调
[ ] contract test 全通过
```

### 27.3 上线期

```text
[ ] DB 备份
[ ] Redis 备份或确认允许强制重新登录
[ ] migration dry-run
[ ] 部署 Go shadow 服务
[ ] 执行 smoke test
[ ] 执行 contract test
[ ] 小流量 canary
[ ] 对比错误率、延迟、登录成功率
[ ] 完全切流
[ ] 保留旧服务回滚窗口
```

---

## 28. 风险与建议

### 28.1 最大风险：接口细节不一致

解决：契约测试优先于业务开发。

### 28.2 最大风险：插件自动扫描不可控

解决：生产只使用生成后的固定注册代码，不运行时动态扫描。

### 28.3 最大风险：任务状态与原 Celery 不一致

解决：实现 task_result 兼容适配层，不直接暴露 Asynq 原生状态。

### 28.4 最大风险：Redis Cluster 与 Asynq

解决：第一阶段优先 Redis Sentinel，不直接上 Cluster。

### 28.5 最大风险：Go 版时间/枚举/空值输出不同

解决：所有 DTO 使用统一 DateTime、枚举和响应 serializer。

---

## 29. 推荐第一版 MVP 范围

第一版不要一次实现所有插件机制的高级能力。建议 MVP：

```text
1. fba-go core module
2. admin/dict/task 三个内置插件
3. import scan + manifest
4. local scan 只支持同一 Go module 下目录
5. plugin dependency graph
6. swagger 聚合
7. contract test
8. Asynq worker + scheduler leader lock
9. Redis sentinel 支持
10. DB migration lock
```

暂缓：

```text
1. 跨 go.mod 本地插件自动构建
2. 插件热插拔
3. Go runtime plugin .so
4. Redis Cluster
5. 非兼容表结构重构
```

---

## 30. 最终落地结论

这套 Go 迁移不应该只是“把 FastAPI 代码翻译成 Go”，而应该抽象成：

```text
一个前端兼容的、插件驱动的、可复用的 Go 后端基础框架 module
```

核心策略：

1. **API 契约冻结**：前端不动，后端必须适配；
2. **Go module 化**：主框架作为可复用 module；
3. **插件业务化**：所有业务通过插件实现；
4. **代码生成优先**：自动扫描插件，但生成静态注册代码；
5. **Dig 只做启动期 DI**：不做运行时 service locator；
6. **Asynq 替代 Celery**：但通过兼容层保持前端任务接口不变；
7. **Zap + Prometheus + OTel**：生产可观测；
8. **迁移锁 + scheduler leader lock**：保证高可用部署安全；
9. **契约测试兜底**：防止前端联调时才发现不兼容。
