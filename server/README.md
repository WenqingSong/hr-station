# Offer 捕手后端服务

## 1. 当前阶段

阶段 1：工程初始化（骨架已搭建，业务功能待后续阶段实现）

## 2. 当前能力

- HTTP Server 监听端口（默认 8080，可通过 `SERVER_PORT` 环境变量配置）
- `/healthz` 健康检查接口
- 基础配置读取（`APP_ENV`、`LOG_LEVEL`、`SERVER_PORT`）

## 3. 目录结构

```
server/
├── go.mod                     # Go module 定义
├── cmd/
│   └── server/
│       └── main.go            # 程序入口
├── internal/
│   ├── config/
│   │   └── config.go          # 配置读取
│   ├── http/
│   │   └── router.go          # HTTP 路由与 Handler
│   ├── middleware/             # 中间件（阶段 2 实现）
│   ├── model/                 # 数据模型（阶段 3 实现）
│   ├── repository/            # 数据访问层（阶段 3 实现）
│   ├── service/               # 业务逻辑层（阶段 3 实现）
│   ├── agent/                 # AI Agent（阶段 3 实现）
│   └── llm/                   # LLM 调用（阶段 4 实现）
└── sql/
    └── init.sql               # 数据库初始化脚本（阶段 3 填写）
```

## 4. 启动方式

```bash
cd server
go run ./cmd/server
```

或通过 Makefile：

```bash
make dev-server
```

## 5. 健康检查

服务启动后，访问：

```bash
curl http://localhost:8080/healthz
```

响应：

```json
{
  "status": "ok",
  "service": "offer-catcher-server"
}
```

## 6. 当前未实现

1. 未连接数据库
2. 未实现用户认证
3. 未实现业务接口
4. 未实现 Agent
5. 未接入模型 API

## 7. 后续阶段

| 阶段 | 新增内容 |
|------|---------|
| 阶段 2 | 配置加载完善、CORS 中间件、日志、MySQL 连接、健康检查增强 |
| 阶段 3 | 认证接口、简历接口、岗位接口、匹配接口、Mock Agent |
| 阶段 4 | 真实 LLM 接入，替换 Mock Agent |
| 阶段 7 | Redis 缓存、完整容器化部署 |
