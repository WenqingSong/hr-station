# Offer 捕手

## 1. 项目简介

Offer 捕手 — 校招简历与岗位智能匹配平台

面向校招求职者的 AI 辅助工具，帮助用户管理简历、发现匹配岗位、获取优化建议，提升求职效率。

核心理念：**不替用户编造经历，而是帮用户更好地呈现真实经历。**

## 2. 当前阶段

**阶段 1：工程初始化（进行中）**

|| 编号 | 任务 | 状态 |
||------|------|------|
|| 1-1 | 创建工程初始化方案与阶段 1 计划 | ✅ 已完成 |
|| 1-2 | 创建 Makefile | ✅ 已完成 |
|| 1-3 | 创建 .env.example | ✅ 已完成 |
|| 1-4 | 编写 docker-compose.yml | ✅ 已完成 |
|| 1-5 | 初始化 Go 后端工程目录 | ✅ 已完成 |
|| 1-6 | 编写前后端运行说明 | ✅ 已完成 |
|| 1-7 | 阶段 1 汇总 | ⬜ 待执行 |

## 3. 快速启动

```bash
# 1. 检查开发环境
make doctor

# 2. 初始化（复制 .env、提示安装依赖）
make init

# 3. 启动 MySQL
make up

# 4. 启动前端（新终端）
make dev-web

# 5. 启动后端（新终端）
make dev-server
```

详细说明请阅读 [docs/06-工程部署/部署说明.md](docs/06-工程部署/部署说明.md)。

## 4. CNB 环境重启恢复

CNB 云原生开发环境重启后，按以下步骤恢复：

```bash
make doctor      # 检查环境
make init        # 补齐 .env
make up          # 启动 MySQL
make status      # 确认 MySQL 就绪
make dev-web     # 启动前端（新终端）
make dev-server  # 启动后端（新终端）
```

## 5. 常用命令

|| 命令 | 说明 |
||------|------|
|| `make help` | 查看所有命令 |
|| `make doctor` | 只读检查开发环境 |
|| `make init` | 初始化开发环境 |
|| `make up` | 启动 MySQL |
|| `make down` | 停止 MySQL（不删数据） |
|| `make dev-web` | 启动前端 |
|| `make dev-server` | 启动后端 |
|| `make status` | 查看容器状态 |
|| `make logs-mysql` | 查看 MySQL 日志 |

## 6. 当前可验证能力

- MySQL 服务可通过 `make up` 启动
- 后端 `/healthz` 健康检查可访问：`curl http://localhost:8080/healthz`
- 前端页面可启动预览（仍使用 Mock 数据）
- 当前阶段不要求前端与后端联通

## 7. 文档入口

|| 文档 | 说明 |
||------|------|
|| [docs/说明书.md](docs/说明书.md) | 项目全貌与文档索引 |
|| [docs/项目状态.md](docs/项目状态.md) | 当前项目整体状态 |
|| [docs/下一步任务.md](docs/下一步任务.md) | 当前待执行任务 |
|| [docs/06-工程部署/部署说明.md](docs/06-工程部署/部署说明.md) | 完整前后端运行说明 |

## 8. 注意事项

- 当前后端仅实现 `/healthz`，不连接数据库
- 当前前端使用 Mock 数据，不接后端 API
- `make down` 不删除 MySQL 数据，`make reset` 才删除（危险）
- LLM API Key 需手动填写在 `.env` 中，不要写入 `.env.example`
