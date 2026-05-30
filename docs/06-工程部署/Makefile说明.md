# Makefile 说明

> 最后更新：2026-05-30（阶段 4-3.7.B.R.2 回退全容器改造并落地混合式启动）

## 1. 设计目标

1. Makefile 作为统一命令入口，不直接写大量复杂 shell 逻辑
2. 具体执行逻辑下沉到 `scripts/dev/*.sh` 脚本
3. 云环境重启后快速恢复开发环境
4. 前端、后端、Docker、数据库、Agent 分开控制
5. 后续阶段持续扩展
6. 新模型 / 新开发者快速接手
7. 危险操作必须 `CONFIRM=yes` 确认
8. 不在脚本中打印真实密钥
9. **restart 不清空 volume，reset/reset-up 清空 volume 需 CONFIRM=yes**
10. **Nginx 未接入时，nginx-logs 等命令给出明确 pending 提示**

## 2. 总体结构

```
Makefile                    # 统一入口 / 命令编排层
scripts/dev/common.sh       # 通用函数（log、load_env、mask_secret 等）
scripts/dev/ensure_env.sh   # .env 自动生成
scripts/dev/doctor.sh       # 开发环境检查
scripts/dev/env.sh          # 环境初始化与配置展示
scripts/dev/infra.sh        # Docker Compose 基础设施管理
scripts/dev/mysql.sh        # 数据库导入/重置/shell/检查
scripts/dev/backend.sh      # 后端启动/停止/重启/日志/状态/健康检查
scripts/dev/web.sh          # 前端安装/启动/停止/重启/日志/状态/健康检查
scripts/dev/agent.sh        # Agent/LLM 环境检查
scripts/dev/logs.sh         # 统一日志查看（含 nginx 预留）
scripts/dev/quality.sh      # 代码质量检查
scripts/dev/clean.sh        # 清理操作
```

## 3. 云原生环境重启恢复流程

CNB 环境重启后，按以下顺序恢复：

```bash
make doctor     # 1. 检查环境
make init       # 2. 初始化（补齐 .env、创建目录）
make up         # 3. 一键启动完整开发环境
make health     # 4. 健康检查
```

如果前端依赖缺失：

```bash
make web-install
make web-up
```

如果数据库为空：

```bash
make db-import
```

如果想回到干净 Demo 数据：

```bash
make db-reset CONFIRM=yes
```

如果想完全清空并重启：

```bash
make reset-up CONFIRM=yes
```

## 4. 常用命令

| 命令 | 说明 | 只读 | 有副作用 |
|------|------|------|---------|
| `make help` | 查看所有命令 | ✅ | ❌ |
| `make doctor` | 检查开发环境 | ✅ | ❌ |
| `make init` | 初始化开发环境 | ❌ | ⚠️ |
| `make ensure-env` | 确保 .env 存在（优先 .env.local → .env.example） | ❌ | ⚠️ |
| `make check` | doctor + 健康检查 | ✅ | ❌ |
| `make status` | 查看所有服务运行状态和访问地址 | ✅ | ❌ |
| `make health` | 查看所有服务健康状态 | ✅ | ❌ |
| `make ps` | 显示 Docker Compose 容器状态 | ✅ | ❌ |

## 5. 生命周期命令

| 命令 | 说明 | 危险级别 |
|------|------|---------|
| `make up` | 一键启动完整开发环境 | 🟢 安全 |
| `make stop` | 停止前后端进程，保留 Docker 数据 | 🟢 安全 |
| `make down` | 停止所有服务，保留数据库 volume | 🟢 安全 |
| `make restart` | 软重启（stop + up，不清空 volume） | 🟢 安全 |
| `make restart-soft` | 同 restart（软重启） | 🟢 安全 |
| `make reset` | 停止并清空 volume | 🔴 需 CONFIRM=yes |
| `make reset-up` | 清空 volume 后重启 | 🔴 需 CONFIRM=yes |
| `make rebuild` | 清理运行态并重建 | 🔴 需 CONFIRM=yes |

### `make up` 执行顺序

```
0. 确保 .env 存在（ensure-env）
1. 环境检查（doctor）
2. 环境初始化（init）
3. 启动基础设施（infra-up）
4. 检查数据库（db-check）
5. 启动后端（backend-up）
6. 启动前端（web-up）
7. 健康检查
```

### `make up` 输出

```
[Offer Catcher] Starting development environment...

[1/7] 环境检查
[2/7] 环境初始化
[3/7] 启动基础设施（MySQL + Nginx）
[4/7] 等待 MySQL 就绪
[5/7] 启动后端
[6/7] 启动前端
[7/7] 健康检查

==========================================
  ✅ Offer Catcher 开发环境已启动
==========================================

Access:
  App:       http://localhost/
  API:       http://localhost/api/
  Health:    http://localhost/healthz
  Backend:   http://localhost:8080
  Frontend:  http://localhost:5173
  MySQL:     localhost:3306

Useful commands:
  make status       查看服务状态
  make logs         查看日志
  make restart-soft 软重启（不清空数据）
  make reset-up     ⚠️ 清空数据并重启（CONFIRM=yes）
  make down         停止所有服务
```

### `make stop` vs `make down`

- `make stop`：只停止本地前后端进程，Docker 容器继续运行
- `make down`：停止前后端 + 停止 Docker 容器，但**不删除 volume**，数据保留

### `make restart` vs `make reset` vs `make reset-up`

- `make restart`：软重启，= stop + up，**不清空 volume**，数据保留
- `make restart-soft`：同 restart，语义更明确
- `make reset`：停止服务并**清空所有 volume**（含 MySQL 数据），必须 `CONFIRM=yes`
- `make reset-up`：reset 后重新 up，必须 `CONFIRM=yes`

设计理由：
1. restart 在行业习惯中通常不代表删除数据（`docker compose restart` 不删 volume）
2. 清空 volume 是破坏性操作，容易误删数据库，必须明确确认
3. 开发阶段可以提供 reset-up 满足快速重置需求，但有 CONFIRM 保护
4. 生产环境必须禁止清空 volume

## 6. 基础设施命令

| 命令 | 说明 |
|------|------|
| `make infra-up` | 启动 Docker Compose 基础设施（MySQL + Nginx） |
| `make infra-down` | 停止 Docker Compose 服务，保留 volume |
| `make infra-status` | 查看 Docker Compose 状态 |
| `make infra-logs` | 查看 Docker Compose 日志 |

## 7. 数据库命令

| 命令 | 说明 | 危险级别 |
|------|------|---------|
| `make db-import` | 导入 server/sql/init.sql | 🟢 可重复执行 |
| `make db-reset` | 重新导入初始化 SQL | 🔴 需 CONFIRM=yes |
| `make db-shell` | 进入 MySQL shell | 🟢 安全 |
| `make db-check` | 检查 6 张核心表记录数 | 🟢 只读 |

`db-check` 检查的 6 张核心表：users、resume_versions、jobs、recommendation_results、match_analyses、version_compare_results。

## 8. 后端命令

| 命令 | 说明 |
|------|------|
| `make backend-up` | 后台启动 Go 后端 |
| `make backend-stop` | 停止 Go 后端 |
| `make backend-restart` | 重启 Go 后端 |
| `make backend-logs` | 查看后端日志 |
| `make backend-status` | 查看后端进程状态 |
| `make backend-health` | curl /healthz |

后端以后台进程方式启动：
- PID 文件：`.run/server.pid`
- 日志文件：`.logs/server.log`
- 启动命令：`cd server && SERVER_PORT=$SERVER_PORT go run ./cmd/server`
- 不产生 `server/server` 编译产物
- 健康检查：`http://localhost:${SERVER_PORT}/healthz`

后端停止逻辑（4-1.1.3 修复）：
- 先终止子进程（`pgrep -P`），再终止主进程
- 兜底：如果端口仍有 go/server 相关残留进程，自动清理
- 停止后端口应完全释放，`backend-restart` 不再因端口残留失败

## 9. 前端命令

| 命令 | 说明 |
|------|------|
| `make web-install` | 在 web/ 下执行 npm install |
| `make web-up` | 后台启动前端开发服务 |
| `make web-stop` | 停止前端服务 |
| `make web-restart` | 重启前端服务 |
| `make web-logs` | 查看前端日志 |
| `make web-status` | 查看前端进程状态 |
| `make web-health` | 检查前端端口 |

前端以后台进程方式启动：
- PID 文件：`.run/web.pid`
- 日志文件：`.logs/web.log`
- `node_modules` 不存在时提示执行 `make web-install`
- 不自动安装依赖

## 10. Agent / LLM 命令

| 命令 | 说明 |
|------|------|
| `make agent-env` | 检查 LLM 相关环境变量（不打印真实 Key） |
| `make agent-check` | 检查 ENABLE_LLM_AGENT / LLM_PROVIDER / LLM_MODEL |
| `make agent-test` | 当前只提示 LLM Client 尚未实现 |

安全规则：
- 不打印真实 `LLM_API_KEY` 值，只显示「已配置/未配置」
- 如果 `LLM_MODEL=hhy3-preview`，提示疑似拼写错误
- `agent-test` 当前不调用 Hunyuan API，阶段 4-2 后启用

### LLM Smoke Test 命令

| 命令 | 说明 | 是否请求外部接口 |
|------|------|-----------------|
| `make llm-help` | 查看 LLM smoke test 命令帮助 | 否 |
| `make llm-check` | 检查本地 LLM 配置 | 否 |
| `make llm-smoke` | 执行 LLM smoke test，需要 `RUN_REAL_LLM_SMOKE=1` | 需要 `RUN_REAL_LLM_SMOKE=1` |

说明：
1. `make llm-check` 不请求外部模型，只检查本地配置
2. `make llm-smoke` 默认不直接调用真实 Hunyuan，需要显式设置 `RUN_REAL_LLM_SMOKE=1`
3. `make llm-smoke` 不在 `make up` / `make ci` 中自动执行
4. smoke test 使用最小 Prompt，不携带简历、JD 或业务数据
5. 不输出 API Key，不输出完整 raw output

4-3.1 已通过 `RUN_REAL_LLM_SMOKE=1 make llm-smoke` 验证真实 LLM 连通性。`make llm-check` 可验证 LLM 配置完整性。`make llm-smoke` 仍需要 `RUN_REAL_LLM_SMOKE=1` 才触发真实模型请求。

4-3.1.D 已修复 llm-smoke 环境加载问题。`llm_smoke.sh` 不再使用 `source .env` 方式加载配置，而是复用 `common.sh` 的 `load_env()`，避免 `DB_DSN` 等复杂值（含括号 `tcp(...)`）导致配置读取失败。

## 11. 日志命令

| 命令 | 说明 |
|------|------|
| `make logs` | 查看日志入口提示 |
| `make logs-backend` | tail .logs/server.log |
| `make logs-web` | tail .logs/web.log |
| `make logs-mysql` | docker compose logs -f mysql |
| `make logs-agent` | 当前查看后端日志中的 Agent 相关内容 |
| `make backend-logs` | 同 logs-backend |
| `make frontend-logs` | 同 logs-web |
| `make db-logs` | 同 logs-mysql |
| `make nginx-logs` | Nginx 容器日志（`docker compose logs -f nginx`） |

日志目录 `.logs/` 由 `make init` 自动创建。

## 12. 质量检查命令

| 命令 | 说明 |
|------|------|
| `make lint` | Go gofmt 检查 + Shell 语法检查 |
| `make test` | cd server && go test ./... |
| `make build` | 前端 build + 后端 go build（输出到 /tmp） |
| `make ci` | doctor + lint + test + build |

`make build` 注意事项：
- 后端 go build 输出到 `/tmp/offer-catcher-server`，不残留 `server/server`
- 前端如果 `node_modules` 不存在，提示先执行 `make web-install`

## 13. 清理命令

| 命令 | 说明 | 危险级别 |
|------|------|---------|
| `make clean` | 清理运行文件和构建产物，不删除数据库 volume | 🟢 安全 |
| `make clean-logs` | 删除 .logs/ | 🟢 安全 |
| `make clean-pids` | 删除 .run/ | 🟢 安全 |
| `make clean-volume` | 删除 Docker volume | 🔴 需 CONFIRM=yes |

`make clean` 清理内容：
- `.run/*.pid`
- `.logs/*.log`
- `web/dist/`
- `server/bin/`
- `server/server`（残留二进制）

## 14. 危险操作确认

以下命令必须 `CONFIRM=yes`：

```bash
make reset CONFIRM=yes        # 停止并清空 volume
make reset-up CONFIRM=yes     # 清空 volume 后重启
make rebuild CONFIRM=yes      # 清理运行态并重建
make db-reset CONFIRM=yes     # 重新导入初始化 SQL
make clean-volume CONFIRM=yes # 删除 Docker volume
```

没有确认时输出提示并拒绝执行：

```text
[ERROR] This command will remove development volumes (including MySQL data).
        Re-run with: make reset CONFIRM=yes
```

## 15. 当前阶段限制

以下功能当前为 pending 状态，将在后续阶段接入：

| 功能 | 状态 | 预计阶段 |
|------|------|---------|
| Nginx 统一入口 | ✅ 已接入 | 4-3.7.B.R.2 |
| nginx-logs 命令 | ✅ 已接入 | 4-3.7.B.R.2 |
| CORS 配置化 | ❌ 未改 | 4-3.7.B.R.3 |
| 生产环境启动命令 | ❌ 未实现 | 后续独立阶段 |

## 16. 常见问题

### Q: CNB 环境重启后如何恢复？

```bash
make doctor → make init → make up → make health
```

### Q: `make init` 会覆盖 .env 吗？

不会。`.env` 已存在时跳过，不会覆盖。

### Q: `make up` 启动失败怎么办？

1. 先执行 `make doctor` 检查环境
2. 如果 MySQL 未就绪，等待后重试：`make backend-up`
3. 如果前端依赖缺失：`make web-install && make web-up`
4. 如果数据库为空：`make db-import`

### Q: `make down` 会删除数据库数据吗？

不会。`make down` 只停止服务和容器，不删除 volume。

### Q: `make restart` 会清空数据吗？

不会。`make restart` 是软重启（= stop + up），不清空 volume，数据保留。如需清空数据，使用 `make reset-up CONFIRM=yes`。

### Q: 如何完全清理数据库？

```bash
make reset-up CONFIRM=yes
```

这将删除 `mysql_data` 数据卷，MySQL 中所有数据将丢失，然后重新启动。

### Q: 后端日志在哪里？

```bash
make logs-backend    # 实时查看
make backend-logs    # 同上
cat .logs/server.log # 查看全部
```

### Q: 前端依赖缺失怎么办？

```bash
make web-install
```

### Q: `make agent-check` 显示 LLM_API_KEY 值吗？

不会。只显示「已配置/未配置」，不打印真实 Key。

### Q: `make nginx-logs` 看什么？

R.2 已接入 Nginx，`make nginx-logs` 输出 Nginx 容器实时日志，包含代理请求记录和错误信息。

## 18. 方案 B 冻结确认（4-3.7.B.R.1）

4-3.7.B.R.1 已冻结 make 命令用户语义。后续 R.2 会将内部实现调整为混合式启动（Docker Compose 管理 MySQL + Nginx，本机进程管理 Backend + Frontend），但 make up / down / restart / reset-up 的用户语义保持不变。

关键变更：
- `make up` 内部：infra-up 将包含 Nginx 容器（而非只有 MySQL）
- `make up` 输出：Access 段将显示 http://localhost/ 作为主入口
- `make nginx-logs`：将从 pending 提示改为真实 Nginx 日志
- `make status`：将显示 Nginx 容器状态

详细方案见 `docs/06-工程部署/混合式开发环境最终方案.md`。

## 19. 方案 B 落地确认（4-3.7.B.R.2）

4-3.7.B.R.2 已完成混合式启动落地。make up 启动 MySQL + Nginx 容器，并启动 Backend / Frontend 本机后台进程。

已实现的关键变更：
- `make up` 内部：infra-up 包含 Nginx 容器
- `make up` 输出：Access 段显示 `http://localhost/` 作为主入口
- `make nginx-logs`：输出真实 Nginx 日志（`docker compose logs -f nginx`）
- `make status`：显示 MySQL + Nginx 容器状态和访问地址
- `make infra-up` 帮助：描述为"启动 Docker Compose 基础设施（MySQL + Nginx）"
- Vite 启动命令：`npm run dev -- --host 0.0.0.0`（使 Nginx 容器可通过 `host.docker.internal` 代理）

详细方案见 `docs/06-工程部署/混合式开发环境最终方案.md`。

## 20. 混合式启动关闭契约修复（4-3.7.B.R.2.1）

4-3.7.B.R.2.1 已修复混合式启动关闭契约。

- `make down` 已能可靠关闭 frontend Vite 子进程，避免 5173 残留
- `make restart-soft` 已能基于干净关闭后重新启动
- `stop_pid_file()` 改为递归查找所有后代进程并终止（处理 npm→sh→node 多层进程树）
- `web.sh stop_web()` 新增端口 5173 兜底检查，确认属于本项目进程后清理
- 不属于本项目的端口占用进程不会强杀，仅输出 WARN 提示用户手动处理

## 关联文档

- [工程执行规范.md](./工程执行规范.md)
- [环境变量.md](./环境变量.md)
- [部署说明.md](./部署说明.md)
- [Docker说明.md](./Docker说明.md)
- [启动与部署标准化方案.md](./启动与部署标准化方案.md)
- [腾讯招聘信息接入方案.md](./腾讯招聘信息接入方案.md)

## 腾讯岗位数据源命令

以下命令用于管理腾讯招聘公开岗位原始数据的拉取、查看和清理。

| 命令 | 说明 | 是否请求外部接口 | 是否入库 |
|------|------|-----------------|---------|
| make jobs-help | 查看帮助 | 否 | 否 |
| make jobs-check | 检查配置 | 否 | 否 |
| make jobs-source-check | 同 jobs-check | 否 | 否 |
| make jobs-fetch-list | 拉取岗位列表原始 JSON | 是 | 否 |
| make jobs-fetch-detail | 拉取岗位详情原始 JSON | 是 | 否 |
| make jobs-fetch-details | 同 jobs-fetch-detail | 是 | 否 |
| make jobs-fetch-dictionary | 拉取字典原始 JSON | 是 | 否 |
| make jobs-fetch | 一键拉取列表 + 详情 + 字典 | 是 | 否 |
| make jobs-sample | 查看本地样本 | 否 | 否 |
| make jobs-status | 查看本地文件状态 | 否 | 否 |
| make jobs-clean | 清理本地 raw JSON | 否 | 否 |

重要说明：

1. jobs-* 命令不会导入数据库
2. jobs-* 命令不会修改后端接口
3. jobs-* 命令不会修改前端页面
4. jobs-fetch 不会自动出现在 make up 中
5. 原始数据用于 4-1.3 适配前的本地缓存
6. 测试时可临时覆盖：`TENCENT_JOBS_PAGE_SIZE=5 TENCENT_JOBS_MAX_PAGES=1 make jobs-fetch`

### 腾讯岗位入库命令

以下命令用于将本地 raw JSON 清洗后导入 MySQL。

| 命令 | 说明 | 是否请求外部接口 | 是否入库 |
|------|------|-----------------|---------|
| make jobs-import-help | 查看帮助 | 否 | 否 |
| make jobs-import-check | 检查 raw JSON 和 MySQL | 否 | 否 |
| make jobs-import-preview | 预览将导入的岗位 | 否 | 否 |
| make jobs-import-sql | 生成导入 SQL | 否 | 否 |
| make jobs-import | 执行导入 | 否 | 是 |
| make jobs-import-status | 查看岗位统计 | 否 | 否 |
| make jobs-db-status | 同 jobs-import-status | 否 | 否 |
| make jobs-import-clean | 清理 generated 文件 | 否 | 否 |

重要说明：

1. jobs-fetch 负责拉取腾讯公开岗位原始 JSON
2. jobs-import 负责把本地 raw JSON 导入 MySQL
3. jobs-import 不请求腾讯接口
4. jobs-import 不修改后端接口
5. jobs-import 不修改前端页面
6. jobs-import 不会自动出现在 make up / make ci 中

推荐流程：

```bash
make jobs-import-check     # 1. 检查环境
make jobs-import-preview   # 2. 预览数据
make jobs-import-sql       # 3. 生成 SQL（可选，import 会自动执行）
make jobs-import           # 4. 执行导入
make jobs-import-status    # 5. 查看结果
```

## 17. ensure-env 机制

4-3.1.C 新增 `ensure-env` 机制。`make up` / `make restart` / `make restart-soft` / `make reset-up` / `make llm-check` / `make llm-smoke` 执行前会自动确保项目根目录存在 `.env`。

优先级：

1. `.env` 已存在 → 不覆盖
2. `.env` 不存在 + `.env.local` 存在 → 从 `.env.local` 复制
3. `.env` 不存在 + `.env.example` 存在 → 从 `.env.example` 复制
4. 三者都不存在 → 报错

说明：

- `.env.local` 用于保存本地真实开发配置（包含 API Key 等），不应提交 Git
- `.env` 由 ensure-env 自动生成，是 Docker / Makefile 实际使用的文件
- `.env.example` 是模板文件，可以提交 Git
- `ensure-env` 不打印文件内容，不打印 API Key

## R.4 更新

4-3.7.B.R.4 已完成混合式一键启动集成验收。`make up` / `make down` / `make restart` / `make status` / `make logs` 链路已通过集成验收。

## R.5 更新

4-3.7.B.R.5 已完成 README 与部署说明更新。Makefile 命令契约与 README 保持一致。4-3.7.B.R 返工阶段已完成。
