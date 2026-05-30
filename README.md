# Offer Catcher / AI 求职匹配智能体

## 1. 项目简介

Offer Catcher 是一个面向学生求职场景的 AI 求职匹配与投递前决策 Demo。

目标不是简单复刻岗位推荐，而是帮助学生理解：

- 适合投哪些岗位
- 为什么适合
- 简历与岗位之间有哪些证据链
- 哪些地方需要优化
- 什么时候可以投递
- 什么时候应该先补充信息或调整目标

## 2. 当前定位

当前项目定位为：**面向学生的投递前决策与简历初筛优化系统**。

官方招聘平台更偏岗位入口和投递流程，本项目更偏投递前分析、JD 证据链、简历优化建议与决策辅助。

## 3. 核心能力

| 能力 | 说明 | 当前状态 |
|------|------|---------|
| 岗位推荐 | 基于简历匹配岗位并排序 | ✅ LLM + Mock fallback |
| 岗位匹配分析 | JD 证据链 + 投递准备度判断 | ✅ LLM + Mock fallback |
| 简历优化 | 目标岗位导向的优化建议 | ✅ Mock Agent（Pipeline 待接入） |
| 简历版本对比 | 分数/字段/风险变化对比 | ✅ Mock Agent |
| LLM Pipeline + Mock fallback | 统一调用链 + 结构化 fallback | ✅ 已实现 |
| JD 证据链 | 逐条对照岗位要求与简历经历 | ✅ 设计完成，部分落地 |
| CORS / Nginx / Makefile 标准化开发环境 | 混合式一键启动 | ✅ R.4 已验收 |

## 4. 技术栈

| 层 | 技术 |
|----|------|
| Backend | Go（net/http + database/sql） |
| Frontend | React + TypeScript + Vite + TailwindCSS |
| Database | MySQL 8.0 |
| Proxy | Nginx（Docker 容器） |
| Dev Infra | Docker Compose + Makefile + scripts/dev |
| LLM | Hunyuan / TokenHub 接入能力 |
| Build / DevOps | Makefile + scripts/dev |

## 5. 当前开发环境架构

当前开发环境采用**混合式架构（方案 B）**：

```
┌──────────────────────────────────────────────────────────┐
│                   Docker Compose                          │
│  ┌─────────────┐    ┌─────────────┐                      │
│  │  MySQL :3306 │    │  Nginx :80   │                      │
│  │  (容器)      │    │  (容器)      │                      │
│  └─────────────┘    └──────┬──────┘                      │
│                            │                              │
└────────────────────────────┼──────────────────────────────┘
                             │ host.docker.internal
                   ┌─────────┴─────────┐
                   │                   │
              ┌────▼────┐        ┌─────▼───┐
              │Backend   │        │Frontend  │
              │:8080     │        │:5173     │
              │(本机进程) │        │(本机进程) │
              └─────────┘        └──────────┘
```

**关键说明：**

- 当前**不是**全容器开发模式
- Backend / Frontend **不由** Docker Compose 管理
- Backend / Frontend 由 Makefile 以本机后台进程方式启动
- Docker Compose **只管理** MySQL + Nginx
- Nginx 通过 `host.docker.internal` 代理到宿主机服务

## 6. 快速启动

```bash
# 1. 复制环境变量（如果 .env 不存在）
cp .env.example .env

# 2. 一键启动
make up
```

如果 `.env` 不存在，`make up` 会自动从 `.env.example` 创建。

启动后访问：

| 入口 | 地址 |
|------|------|
| App | http://localhost/ |
| API | http://localhost/api/ |
| Health | http://localhost/healthz |
| Backend 直连 | http://localhost:8080 |
| Frontend 直连 | http://localhost:5173 |

## 7. 常用命令

### 生命周期

```bash
make up            # 启动完整混合式开发环境
make down          # 关闭完整混合式开发环境，不清 volume
make restart       # 软重启完整混合式开发环境，不清 volume
make restart-soft  # 同 restart
make status        # 查看完整开发环境状态
make ps            # 查看 Docker Compose 服务状态
make logs          # 查看完整开发环境日志
```

### 日志

```bash
make backend-logs   # 后端日志
make frontend-logs  # 前端日志
make nginx-logs     # Nginx 容器日志
make db-logs        # MySQL 容器日志
```

### 破坏性命令

```bash
make reset CONFIRM=yes      # 清开发 volume，必须 CONFIRM=yes
make reset-up CONFIRM=yes   # 清开发 volume 后重启，必须 CONFIRM=yes
```

⚠️ **`reset` / `reset-up` 会清空开发 volume（包括 MySQL 数据），必须谨慎使用。不带 `CONFIRM=yes` 会拒绝执行。**

## 8. 访问地址

| 入口 | Nginx 代理目标 | 说明 |
|------|---------------|------|
| `http://localhost/` | Frontend :5173 | 前端页面 |
| `http://localhost/api/` | Backend :8080 | 后端 API（不 strip `/api`） |
| `http://localhost/healthz` | Backend :8080/healthz | 健康检查 |

## 9. 环境变量

| 规则 | 说明 |
|------|------|
| `.env.example` | 模板文件，可提交 Git，只放占位符 |
| `.env` | 本地运行时配置，由 `ensure-env` 自动生成，不提交 |
| `.env.local` | 本地真实配置（含 API Key），不提交 |

**敏感变量必须本地配置：**

- `LLM_API_KEY` — LLM API 密钥
- `JWT_SECRET` — JWT 签名密钥
- `DB_PASSWORD` — 数据库密码

**生产环境应通过环境变量或 Secret 注入，不要使用 `.env` 文件。**

## 10. LLM 配置说明

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `LLM_ENABLE` | `false` | `true` 时启用真实 LLM Pipeline |
| `LLM_PROVIDER` | `hunyuan` | LLM 提供商 |
| `LLM_BASE_URL` | （空） | API Base URL，不要提交真实私有地址 |
| `LLM_API_KEY` | （空） | API Key，**只填写在本地 `.env`** |
| `LLM_MODEL` | （空） | 模型名称 |

### Smoke Test

```bash
make llm-check                          # 检查本地 LLM 配置（不请求外部模型）
RUN_REAL_LLM_SMOKE=1 make llm-smoke    # 真实连通性测试（需显式开关）
```

**安全规则：**

- `llm-check` 不请求外部模型
- `llm-smoke` 需要显式 `RUN_REAL_LLM_SMOKE=1` 才触发真实请求
- 不要提交真实 API Key
- 不要在文档 / 日志 / Git 中记录真实 Key

## 11. Nginx 与 API 代理

推荐前端通过 `/api` 访问后端（Nginx 同源代理，无需 CORS）：

```
浏览器 → http://localhost/ → Nginx → Frontend :5173
浏览器 → /api/... → Nginx → Backend :8080
```

关键规则：

- `/api/` 不 strip 路径前缀，后端收到完整 `/api/...` 路径
- Vite HMR WebSocket 代理已支持
- SSE / 长连接已预留（`proxy_buffering off`）

## 12. CORS 策略

| 场景 | 是否需要 CORS |
|------|-------------|
| 通过 `http://localhost/` 访问（Nginx 同源代理） | ❌ 不需要 |
| 前端直连 `http://localhost:8080` 调试 | ✅ 需要 |

**安全规则：**

- 日常推荐走 Nginx 同源代理
- 后端直连调试时使用 `CORS_ALLOWED_ORIGINS`
- 生产环境**不要**使用 `*`
- 生产环境只允许正式域名

## 13. 开发环境与部署边界

**当前已完成的是开发环境标准化，不是最终生产部署方案。**

| 维度 | 开发环境 | 生产部署（待实现） |
|------|---------|-----------------|
| Docker Compose | MySQL + Nginx | 需单独设计 |
| Backend/Frontend | 本机后台进程 | 需构建镜像或托管 |
| HTTPS | ❌ 无 | ✅ 必须配置 |
| 域名 | localhost | 正式域名 |
| CORS | 允许 localhost | 只允许正式域名 |
| 密钥 | `.env` 占位值 | Secret 注入 |

生产部署注意事项：

1. 不建议直接把开发 `.env` 用于生产
2. 不建议在生产环境执行 `reset` / `reset-up`
3. 生产环境必须替换 `JWT_SECRET`、`DB_PASSWORD`
4. 生产环境必须配置 HTTPS 和正式域名
5. README 当前提供的是开发启动和部署准备说明

## 14. 常见问题

### Q1：make up 后访问哪里？

```
App:    http://localhost/
API:    http://localhost/api/
Health: http://localhost/healthz
```

### Q2：make down 会不会删数据库？

不会。`make down` 只停止服务和容器，不删除 volume。`make reset CONFIRM=yes` 才会清空 volume。

### Q3：为什么不是全容器开发？

全容器开发虽然更云原生，但对当前 Demo 阶段过重：后端容器热重启、前端容器 HMR、node_modules volume 等配置复杂度高。Go 后端本机运行更利于 IDE 调试，Vite 本机运行 HMR 更快。

### Q4：为什么使用 Nginx？

Nginx 作为统一反向代理入口，解决 CORS 问题，支持 `/api` 同源代理，并为后续生产部署提供基础。

### Q5：为什么后端和前端是本机后台进程？

更利于 IDE 断点调试（Go dlv debug）、代码修改即时生效（air 热重载）、原生 HMR（Vite 文件系统监听）。

### Q6：端口 5173 / 8080 被占用怎么办？

```bash
make backend-stop    # 清理后端进程
make web-stop        # 清理前端进程
make doctor          # 检查端口状态
```

### Q7：CORS 报错怎么办？

推荐通过 `http://localhost/`（Nginx 代理）访问，此时不需要 CORS。如果需要直连调试，确保 `.env` 中 `CORS_ALLOWED_ORIGINS` 包含前端 Origin。

### Q8：LLM_API_KEY 怎么配置？

在本地 `.env` 文件中设置，不要写入 `.env.example`：

```env
LLM_ENABLE=true
LLM_API_KEY=sk-your-real-api-key
LLM_BASE_URL=https://your-api-base-url
LLM_MODEL=your-model-name
```

### Q9：reset-up 什么时候用？

当数据库数据混乱或需要回到干净 Demo 数据时使用。**必须带 `CONFIRM=yes`，否则拒绝执行**。

### Q10：如何查看日志？

```bash
make logs            # 查看日志入口提示
make backend-logs   # 后端日志
make frontend-logs  # 前端日志
make nginx-logs     # Nginx 日志
make db-logs        # MySQL 日志
```

## 15. 项目目录说明

```
├── Makefile              # 统一开发命令入口
├── docker-compose.yml    # MySQL + Nginx 容器编排
├── nginx/                # Nginx 配置
├── scripts/dev/          # 开发脚本（15 个）
├── server/               # Go 后端
│   ├── cmd/server/       # 入口
│   ├── internal/         # 业务逻辑
│   ├── prompts/          # Agent Prompt 模板
│   └── sql/              # 数据库初始化 SQL
├── web/                  # React 前端
├── docs/                 # 项目文档
├── .env.example         # 环境变量模板
└── data/                 # 数据目录（不入 Git）
```

## 16. 当前限制

1. 前端部分能力仍是 Demo 展示（无真实路由、硬编码 Mock 数据）
2. 部分 Agent 仍有 Mock fallback
3. 生产部署还未最终收口
4. LLM 调用需要用户自己配置 Key
5. 当前不会提交真实密钥

## 17. 后续计划

- 前端接后端 API（阶段 5）
- ResumeOptimizationAgent 接入 Pipeline（4-3.8）
- 生产部署方案设计
- 前端体验优化

---

详细文档请阅读 [docs/说明书.md](docs/说明书.md)
