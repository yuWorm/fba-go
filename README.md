# fba-go

[English](README.en.md)

`fba-go` 是 [FastAPI Best Architecture](https://github.com/fastapi-practices/fastapi_best_architecture) 的 Go 版本迁移与演进实现。项目以 Go module 形式提供可复用后台核心能力，并基于 Fiber v3、插件注册体系和项目模板承载后台管理、认证、配置、字典、通知、任务调度等业务实现。

## 特性

- **Go module 核心**：应用启动、配置加载、响应结构、分页、认证中间件、RBAC、插件注册、Swagger、实时通信和任务抽象。
- **Fiber v3**：HTTP 层使用 Fiber v3。
- **日志分流**：终端使用带颜色的常规文本日志，访问日志和错误日志文件保持结构化 JSON。
- **插件体系**：官方、第三方和项目本地模块统一通过 `plugins.yaml` 注册，`plugins.lock` 记录实际 Go module 版本与本地替换。
- **内置 Admin 模板**：`fbago` 直接携带薄 Admin starter；完整功能由独立版本化的 [`fba-go-admin`](https://github.com/yuWorm/fba-go-admin) module 提供。
- **可接管源码**：默认通过 Go module 升级；初始化时可用 `--template-replace`，已有项目可用 `fbago module use` 指向本地 fork 或 Git submodule。

## 快速开始

安装 CLI：

```bash
go install github.com/yuWorm/fba-go/cmd/fbago@latest
```

使用 CLI 创建默认 Admin 项目：

```bash
fbago init github.com/your-org/my-admin --dir ./my-admin
cd my-admin

fbago plugin sync --check
make test
make run
```

也可以在已有项目中直接引入 `fba-go`：

```go
package main

import (
	"context"
	"log"

	fba "github.com/yuWorm/fba-go"
)

func main() {
	opts, err := fba.LoadOptionsFromEnv()
	if err != nil {
		log.Fatal(err)
	}
	app, err := fba.NewApplication(opts)
	if err != nil {
		log.Fatal(err)
	}
	if err := app.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
```

运行：

```bash
go run .
```

## 使用脚手架

创建默认 Admin starter 项目：

```bash
fbago init github.com/your-org/my-admin --dir ./my-admin

cd my-admin
fbago plugin sync --check
make test
make run
```

创建最小后端项目：

```bash
fbago init github.com/your-org/my-backend \
  --template basic \
  --dir ./my-backend

cd my-backend
make tidy
make test
make run
```

## 仓库结构

| 路径 | 说明 |
| --- | --- |
| `core/` | 可复用核心能力与稳定接口 |
| `cmd/fbago/` | CLI，以及内置的 `basic`、`admin` 脚手架 |
| `contracts/` | core smoke API contract 定义 |
| `skills/` | FBA Go 框架、插件和脚手架的 AI 工程技能 |
| `scripts/verify-template.sh` | 内置 Admin starter 端到端验证脚本 |
| `docs/` | 迁移设计与实现说明 |
| `sources/fastapi-best-architecture/` | 可选本地参考源码目录，通常不随公开仓库发布 |

## 常用命令

| 命令 | 说明 |
| --- | --- |
| `go install github.com/yuWorm/fba-go/cmd/fbago@latest` | 安装 CLI |
| `fbago init <module>` | 使用内置 Admin 模板创建项目 |
| `fbago init <module> --template-replace ../fba-go-admin` | 使用本地 Admin checkout 初始化项目 |
| `fbago init <module> --template basic` | 创建最小后端项目 |
| `fbago secret generate [--bytes N]` | 生成适合 JWT、TOTP 等用途的 base64url 加密随机密钥，默认 32 字节 |
| `go run ./cmd/api admin bootstrap` | 使用 `.env` 的 `ADMIN_BOOTSTRAP_PASSWORD` 激活一次性 Admin 引导账号；未设置时交互输入并确认密码 |
| `go run ./cmd/fbago template list` | 查看内置模板，本地开发使用 |
| `fbago template diff` | 基于 `.fbago.yaml` 查看模板 managed 文件变化 |
| `fbago template update --dry-run` | 预览模板 managed 文件更新 |
| `fbago plugin sync` | 根据 `plugins.yaml` 生成注册代码、整理依赖并写入版本锁 |
| `fbago plugin sync --check` | 校验当前依赖图的注册代码、`go.mod`、`go.sum` 和插件锁；不查询远端版本 |
| `fbago plugin outdated` | 查询 manifest 中插件模块的当前版本、可用版本和本地替换 |
| `fbago plugin update --dry-run` | 预览 manifest 中插件模块的版本更新 |
| `fbago plugin update [PLUGIN_OR_MODULE] [--to VERSION]` | 更新去重后的 Go module，并自动重新同步插件注册与锁文件 |
| `fbago module use --path ../fba-go-admin github.com/yuWorm/fba-go-admin` | 使用本地 checkout 接管 Admin module |
| `fbago module reset github.com/yuWorm/fba-go-admin` | 移除本地接管，恢复 `go.mod` 选择的版本 |
| `make test` | 运行 core 测试 |
| `make verify-template` | 验证内置 Admin starter 与独立 Admin module 的集成 |

## 本地开发

```bash
git clone https://github.com/yuWorm/fba-go.git
git clone https://github.com/yuWorm/fba-go-admin.git
cd fba-go

make test
make verify-template
```

发布版生成项目使用固定语义化版本。联调未发布的 Admin 源码时，在
`fbago init` 中传入 `--template-replace <fba-go-admin checkout>`；项目生成后
则用 `fbago module use` 显式切换本地 checkout。

## 更多文档

- [迁移与设计文档](docs/fba_go_module_migration_ha_design.md)
- [Admin module 仓库](https://github.com/yuWorm/fba-go-admin)
