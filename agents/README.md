# Offer Catcher - Agents 服务

基于 [GoFrame v2](https://goframe.org) 的独立后端服务，与主服务 `server/` 隔离，使用独立的数据库 `offer_catcher_re`。

## 目录结构

```
agents/
├── api/                          # 接口定义（请求/响应结构，GoFrame 规范）
│   └── health/v1/health.go       # /healthz 接口定义
├── internal/
│   ├── cmd/
│   │   ├── cmd.go                # 命令入口：加载配置、限时 DB 预检、注册路由、启动服务
│   │   ├── cmd_test.go           # 自动化接口测试（真实 HTTP GET /healthz）
│   │   └── config_test.go        # 配置加载测试
│   ├── controller/health/        # 健康检查控制器
│   └── middleware/               # 数据库健康检查中间件
├── manifest/config/config.yaml   # GoFrame 配置（server/logger/app，不含数据库凭据）
├── sql/init.sql                  # 数据库初始化脚本（创建 offer_catcher_re）
├── main.go                       # 入口：校验必填环境变量并注入 DB 配置
├── Makefile                      # 独立 Make 命令
├── go.mod
└── .env.example                  # 环境变量示例（不含默认密码）
```

## 快速开始

### 前置条件

- Go 1.24+
- MySQL 8.0（与主服务共用实例即可）

### 1. 初始化数据库

```bash
make db-init
```

该命令执行 `sql/init.sql`，幂等创建数据库 `offer_catcher_re`。应用启动过程**不会**自动创建数据库。

### 2. 配置环境变量（必填）

```bash
cp .env.example .env
```

数据库连接信息（`DB_HOST` / `DB_PORT` / `DB_USER` / `DB_PASSWORD`）为必填项，从环境变量读取；缺少任一必填项时，启动会输出明确错误并终止。`AGENTS_PORT` 与 `AGENTS_DB_NAME` 有内置默认值（`18080` / `offer_catcher_re`）。`main.go` 会自动向上级目录加载根 `.env`。

### 3. 一键启动

```bash
make start
```

启动时后端会**先对数据库执行限时 Ping**；若连接失败，将输出明确错误信息并**立即终止启动**，不会监听端口、不会对外提供服务。

## Make 命令

| 命令 | 说明 |
| --- | --- |
| `make help` | 查看所有可用命令 |
| `make doctor` | 检查 Go 环境 |
| `make build` | 编译二进制到 `bin/agents` |
| `make run` | 前台运行（开发调试用） |
| `make start` | 后台一键启动（含编译 + 健康检查等待） |
| `make stop` | 停止服务 |
| `make restart` | 重启服务 |
| `make health` | 健康检查 `curl /healthz` |
| `make status` | 查看进程状态 |
| `make logs` | 查看服务日志 |
| `make test` | 运行自动化接口测试 |
| `make db-init` | 创建数据库 `offer_catcher_re` |
| `make clean` | 清理二进制/日志/pid |

## 健康检查

服务启动后访问：

```bash
curl http://localhost:18080/healthz
```

返回示例：

```json
{
  "code": 0,
  "message": "",
  "data": {
    "service": "offer-catcher-agents",
    "env": "local",
    "app": "ok",
    "database": "ok"
  }
}
```

其中 `database` 字段由数据库健康检查中间件实时探测，取值 `ok` 或 `unavailable`。

## 启动流程与关键行为

1. **配置加载**：GoFrame 启动时自动加载 `manifest/config/config.yaml`（仅非敏感静态配置）；数据库连接信息由 `main.go` 从环境变量读取并程序化注入（`gdb.SetConfig`），配置文件不包含任何密码。
2. **数据库预检**：HTTP 服务监听前对默认数据库执行限时 Ping（5s 超时），失败则 `Fatal` 退出，绝不监听端口。
3. **数据库健康检查中间件**：`middleware.DBHealth` 在运行期间探测数据库连通性，结果写入请求上下文供 `/healthz` 读取；它只负责报告状态，不替代启动检查。
4. **优雅退出**：`config.yaml` 开启 `graceful: true` + `gracefulTimeout: 10`，收到 SIGINT/SIGTERM 后等待进行中请求完成再退出。

## 自动化测试

```bash
make test
```

包含两类测试：

- `TestConfigLoads`：验证 `config.yaml` 的 `app / server` 静态配置可正确解析。
- `TestHealthzEndpoint`：进程内启动 GoFrame 服务，对 `/healthz` 发起真实 HTTP 请求并断言响应。

## 日志

日志输出到 `logs/agents-{Ymd}.log`，同时输出到标准输出，按天滚动、保留 7 天、单文件 100M 上限。
