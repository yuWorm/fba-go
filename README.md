# fba-go

[English](README.en.md)

`fba-go` 是 [FastAPI Best Architecture](https://github.com/fastapi-practices/fastapi_best_architecture) 的 Go 版本迁移与演进实现。项目以 Go module 形式提供可复用后台核心能力，并基于 Fiber v3、插件注册体系和项目模板承载后台管理、认证、配置、字典、通知、任务调度等业务实现。

## 特性

- **Go module 核心**：应用启动、配置加载、响应结构、分页、认证中间件、RBAC、插件注册、Swagger、实时通信和任务抽象。
- **Fiber v3**：HTTP 层使用 Fiber v3。
- **插件体系**：核心保留插件注册与扫描能力；可修改业务代码优先放在项目模板的 `internal/app` 与 `plugins` 中。
- **接口兼容**：主仓库保留 core contract；完整 admin API 行为由模板仓库承载和验证。
- **模板系统**：官方模板仓库通过 submodule 引入到 `templates/fba-go-template`，也支持远程 Git 模板。

## 快速开始

安装 CLI：

```bash
go install github.com/yuWorm/fba-go/cmd/fbago@latest
```

使用 CLI 创建项目：

```bash
fbago init github.com/your-org/my-backend --dir ./my-backend
cd my-backend

go get github.com/yuWorm/fba-go@latest
make tidy
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

创建最小后端项目：

```bash
fbago init github.com/your-org/my-backend --dir ./my-backend

cd my-backend
go get github.com/yuWorm/fba-go@latest
make tidy
make test
make run
```

创建完整 admin starter 项目：

```bash
fbago init github.com/your-org/my-admin \
  --template github.com/yuWorm/fba-go-template/admin@master \
  --dir ./my-admin

cd my-admin
go get github.com/yuWorm/fba-go@latest
make tidy
make test
make run
```

## 仓库结构

| 路径 | 说明 |
| --- | --- |
| `core/` | 可复用核心能力与稳定接口 |
| `cmd/fbago/` | CLI：脚手架、插件扫描、Swagger、contract 测试 |
| `contracts/` | core smoke API contract 定义 |
| `templates/fba-go-template/` | 官方模板仓库 submodule |
| `docs/` | 迁移设计与实现说明 |
| `sources/fastapi-best-architecture/` | 可选本地参考源码目录，通常不随公开仓库发布 |

## 常用命令

| 命令 | 说明 |
| --- | --- |
| `go install github.com/yuWorm/fba-go/cmd/fbago@latest` | 安装 CLI |
| `fbago init <module>` | 创建项目 |
| `go run ./cmd/fbago template list` | 查看内置模板，本地开发使用 |
| `make test` | 运行 core 测试 |
| `make verify-template` | 运行官方 admin 模板与生成项目验证 |

## 本地开发

```bash
git clone --recursive https://github.com/yuWorm/fba-go.git
cd fba-go

# 如果 clone 时没有带 --recursive
git submodule update --init --recursive

make test
```

本地开发版 `fbago` 会自动在生成项目的 `go.mod` 中写入 `replace github.com/yuWorm/fba-go => <local-path>`。安装版或其他布局可使用 `--core-replace /path/to/fba-go` 或 `FBAGO_CORE_REPLACE`。

## 更多文档

- [迁移与设计文档](docs/fba_go_module_migration_ha_design.md)
- [模板仓库](https://github.com/yuWorm/fba-go-template)
