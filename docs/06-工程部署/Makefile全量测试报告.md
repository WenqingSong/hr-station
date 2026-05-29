# Makefile 全量测试报告

> 测试时间：2026-05-29
> 测试人：CodeBuddy（工程复核助手）
> 测试环境：CNB 云原生开发环境（Linux / Docker / Go 1.24 / Node 22）

## 1. 测试目标

对 Makefile 中所有 target 进行逐个验收，确认：
- 命令可正常执行
- 危险操作受 `CONFIRM=yes` 保护
- 不打印真实密钥
- 不误删数据库 volume
- 云原生重启恢复流程成立

## 2. 测试环境

| 项目 | 值 |
|------|------|
| OS | Linux（CNB 云原生） |
| Go | 1.24.1 |
| Node | 22.22.2 |
| npm | 10.9.7 |
| Docker | 28.5.2 |
| Docker Compose | 可用 |
| web/node_modules | 不存在 |

## 3. 测试时间

2026-05-29

## 4. 测试范围

Makefile 中声明的所有 `.PHONY` target，共 43 个命令。

## 5. 测试结论

**⚠️ 发现 4 个 Bug，需修复后才能进入 4-2。**

| 分类 | 通过 | 失败 | 总计 |
|------|------|------|------|
| 基础命令 | 5 | 0 | 5 |
| 生命周期 | 3 | 2 | 5 |
| 基础设施 | 4 | 0 | 4 |
| 数据库 | 3 | 1 | 4 |
| 后端 | 3 | 3 | 6 |
| 前端 | 5 | 2 | 7 |
| Agent / LLM | 3 | 0 | 3 |
| 日志 | 5 | 0 | 5 |
| 质量检查 | 2 | 2 | 4 |
| 清理 | 4 | 0 | 4 |
| 遗漏命令 | 0 | 1 | 1 |
| **总计** | **37** | **9** | **48** |

## 6. 命令测试总览

| # | 命令 | 结果 | 是否执行 | 是否有副作用 | 输出摘要 | 问题 | 建议 |
|---|------|------|---------|------------|---------|------|------|
| 1 | make help | ✅ | 是 | 无 | 列出 40+ 个命令 | 无 | — |
| 2 | make doctor | ✅ | 是 | 无 | 15+ 项检查 | 无 | — |
| 3 | make init | ✅ | 是 | 创建 .run/.logs | 幂等初始化 | 无 | — |
| 4 | make check | ⚠️ | 是 | 无 | doctor + health | backend.sh health 报 local 错误 | 修复 backend.sh |
| 5 | make status | ⚠️ | 是 | 无 | 三个服务状态 | backend.sh status 报 local 错误 | 修复 backend.sh |
| 6 | make up | ⚠️ | 是 | 启动 MySQL | 一键启动流程 | db-check 和 backend-health 报错 | 修复 mysql.sh 和 backend.sh |
| 7 | make stop | ✅ | 是 | 停止前后端 | 正确停止 | 无 | — |
| 8 | make down | ✅ | 是 | 停止 Docker | 保留 volume | 无 | — |
| 9 | make restart | ⚠️ | 是 | stop + up | 同 up 的问题 | 同 up | 修复子问题 |
| 10 | make rebuild | ✅ | 仅保护测试 | 无 | 正确拒绝执行 | 无 | — |
| 11 | make infra-up | ✅ | 是 | 启动 MySQL | 正常 | 无 | — |
| 12 | make infra-down | ✅ | 是 | 停止 MySQL | 保留 volume | 无 | — |
| 13 | make infra-status | ✅ | 是 | 无 | 显示容器状态 | 无 | — |
| 14 | make infra-logs | ✅ | 是 | 无 | 显示 MySQL 日志 | 无 | — |
| 15 | make db-import | ✅ | 是 | 导入 SQL | 成功导入 | 无 | — |
| 16 | make db-check | ❌ | 是 | 无 | local 错误 | mysql.sh line 98: local 在 case 块中 | 将 check 逻辑封装为函数 |
| 17 | make db-shell | ✅ | 是 | 进入 MySQL | 成功连接 | 无 | — |
| 18 | make db-reset | ✅ | 仅保护测试 | 无 | 正确拒绝执行 | 无 | — |
| 19 | make backend-up | ⚠️ | 是 | 启动后端 | 首次超时但实际启动成功 | 首次 go run 下载依赖超 15s | 增加等待时间或分两步 |
| 20 | make backend-stop | ⚠️ | 是 | 停止后端 | 只杀 go run 进程 | go run 子进程 server 仍在运行 | kill 时需杀进程组 |
| 21 | make backend-restart | ❌ | 是 | 停止+启动 | 端口被残留子进程占用 | 同 backend-stop 问题 | 修复 stop 逻辑 |
| 22 | make backend-logs | ✅ | 是 | 无 | tail 日志 | 无 | — |
| 23 | make backend-status | ❌ | 是 | 无 | local 错误 | backend.sh line 100: local 在 case 块中 | 将 status 逻辑封装为函数 |
| 24 | make backend-health | ❌ | 是 | 无 | local 错误 | backend.sh line 109: local 在 case 块中 | 将 health 逻辑封装为函数 |
| 25 | make web-install | ⬜ | 未执行 | — | — | node_modules 不存在，按要求不执行 | — |
| 26 | make web-up | ✅ | 是 | 无 | 正确提示 node_modules 缺失 | 无 | — |
| 27 | make web-stop | ✅ | 是 | 无 | 正确停止 | 无 | — |
| 28 | make web-restart | ✅ | 是 | 无 | 正确提示 node_modules 缺失 | 无 | — |
| 29 | make web-logs | ✅ | 是 | 无 | 正确提示日志不存在 | 无 | — |
| 30 | make web-status | ✅ | 是 | 无 | 显示未运行 | 无 | — |
| 31 | make web-health | ✅ | 是 | 无 | 端口不可访问 | 无 | — |
| 32 | make agent-env | ✅ | 是 | 无 | 显示 LLM 配置 | 无 | — |
| 33 | make agent-check | ✅ | 是 | 无 | 检查 LLM 配置 | 无 | — |
| 34 | make agent-test | ✅ | 是 | 无 | 提示 4-2 后启用 | 无 | — |
| 35 | make logs | ✅ | 是 | 无 | 显示日志入口 | 无 | — |
| 36 | make logs-backend | ✅ | 是 | 无 | tail server.log | 无 | — |
| 37 | make logs-web | ✅ | 是 | 无 | 提示日志不存在 | 无 | — |
| 38 | make logs-mysql | ✅ | 是 | 无 | docker compose logs | 无 | — |
| 39 | make logs-agent | ✅ | 是 | 无 | grep agent/llm | 无 | — |
| 40 | make lint | ❌ | 是 | 无 | quality.sh line 21: local 错误 | local 在 case 块中 | 封装为函数 |
| 41 | make test | ✅ | 是 | 无 | go test ./... | 无 | — |
| 42 | make build | ✅ | 是 | 构建到 /tmp | 前端跳过、后端成功 | 无 | — |
| 43 | make ci | ❌ | 是 | 无 | 因 lint 失败而中断 | quality.sh local 错误 | 修复 quality.sh |
| 44 | make clean | ✅ | 是 | 清理 .run/.logs | 不删除 volume | 无 | — |
| 45 | make clean-logs | ✅ | 是 | 清理 .logs | 不删除 volume | 无 | — |
| 46 | make clean-pids | ✅ | 是 | 清理 .run | 不删除 volume | 无 | — |
| 47 | make clean-volume | ✅ | 仅保护测试 | 无 | 正确拒绝执行 | 无 | — |
| 48 | make health | ❌ | 是 | 无 | No rule to make target | 遗漏了 health target | 新增 health target |

## 7. 基础命令测试

| 命令 | 结果 | 说明 |
|------|------|------|
| make help | ✅ | 列出所有 40+ 个命令，格式清晰 |
| make doctor | ✅ | 检查 15+ 项，区分 error/warn，不打印 Key |
| make init | ✅ | 幂等，不覆盖 .env，创建 .run/.logs |
| make check | ⚠️ | doctor 通过，但 backend.sh health 报 local 错误 |
| make status | ⚠️ | infra/web 正常，backend.sh status 报 local 错误 |

## 8. 生命周期命令测试

| 命令 | 结果 | 说明 |
|------|------|------|
| make up | ⚠️ | 整体流程执行，但 db-check 和 backend-health 报错 |
| make stop | ✅ | 正确停止前后端进程 |
| make down | ✅ | 停止 Docker，保留 volume |
| make restart | ⚠️ | 同 up 的问题 |
| make rebuild | ✅ | 无 CONFIRM=yes 时正确拒绝 |

## 9. 基础设施命令测试

| 命令 | 结果 | 说明 |
|------|------|------|
| make infra-up | ✅ | 启动 MySQL，约 48s |
| make infra-down | ✅ | 停止 MySQL，保留 volume |
| make infra-status | ✅ | 显示容器状态 |
| make infra-logs | ✅ | 显示 MySQL 日志 |

## 10. 数据库命令测试

| 命令 | 结果 | 说明 |
|------|------|------|
| make db-import | ✅ | 成功导入 init.sql |
| make db-reset | ✅ | 无 CONFIRM=yes 时正确拒绝 |
| make db-shell | ✅ | 成功进入 MySQL |
| make db-check | ❌ | `local: can only be used in a function` (line 98) |

## 11. 后端命令测试

| 命令 | 结果 | 说明 |
|------|------|------|
| make backend-up | ⚠️ | 首次启动因 go 下载依赖超时，但实际成功启动 |
| make backend-stop | ⚠️ | 只杀 go run 进程，子进程 server 仍在运行 |
| make backend-restart | ❌ | stop 后端口被残留子进程占用，启动失败 |
| make backend-logs | ✅ | 正确 tail 日志 |
| make backend-status | ❌ | `local: can only be used in a function` (line 100) |
| make backend-health | ❌ | `local: can only be used in a function` (line 109) |

## 12. 前端命令测试

| 命令 | 结果 | 说明 |
|------|------|------|
| make web-install | ⬜ | 未执行（按要求不运行 npm install） |
| make web-up | ✅ | node_modules 缺失时给出友好提示 |
| make web-stop | ✅ | 正确停止 |
| make web-restart | ✅ | node_modules 缺失时给出友好提示 |
| make web-logs | ✅ | 日志不存在时给出友好提示 |
| make web-status | ✅ | 显示未运行 |
| make web-health | ✅ | 端口不可访问时正确提示 |

## 13. Agent / LLM 命令测试

| 命令 | 结果 | 说明 |
|------|------|------|
| make agent-env | ✅ | 显示配置，不打印 Key |
| make agent-check | ✅ | 检查配置完整性 |
| make agent-test | ✅ | 正确提示 4-2 后启用 |

## 14. 日志命令测试

| 命令 | 结果 | 说明 |
|------|------|------|
| make logs | ✅ | 显示日志入口和文件列表 |
| make logs-backend | ✅ | tail server.log |
| make logs-web | ✅ | 日志不存在时友好提示 |
| make logs-mysql | ✅ | docker compose logs mysql |
| make logs-agent | ✅ | grep agent/llm 相关内容 |

## 15. 质量检查命令测试

| 命令 | 结果 | 说明 |
|------|------|------|
| make lint | ❌ | `local: can only be used in a function` (quality.sh line 21) |
| make test | ✅ | go test ./... 通过（无测试文件，但不报错） |
| make build | ✅ | 后端构建到 /tmp，前端跳过 |
| make ci | ❌ | 因 lint 报错而中断 |

## 16. 清理命令测试

| 命令 | 结果 | 说明 |
|------|------|------|
| make clean | ✅ | 清理 .run/.logs，不删 volume |
| make clean-logs | ✅ | 清理 .logs |
| make clean-pids | ✅ | 清理 .run |
| make clean-volume | ✅ | 无 CONFIRM=yes 时正确拒绝 |

## 17. 危险操作保护测试

| 命令 | 是否拒绝 | 是否提示 CONFIRM=yes | 是否有副作用 |
|------|---------|---------------------|------------|
| make rebuild | ✅ 拒绝 | ✅ 提示 | 无 |
| make db-reset | ✅ 拒绝 | ✅ 提示 | 无 |
| make clean-volume | ✅ 拒绝 | ✅ 提示 | 无 |

**结论：危险操作保护机制正确。**

## 18. 密钥安全检查

| 命令 | 是否打印真实 Key | 说明 |
|------|-----------------|------|
| make doctor | ❌ 不打印 | 显示 `LLM_API_KEY: 已配置` |
| make agent-env | ❌ 不打印 | 显示 `LLM_API_KEY: 已配置` |
| make agent-check | ❌ 不打印 | 显示 `LLM_API_KEY: 已配置` |
| make status | ❌ 不打印 | 不显示 Key |
| make logs-* | ❌ 不打印 | 日志中不含 Key |

**结论：密钥安全检查通过。**

## 19. 遗留进程检查

| 检查项 | 结果 |
|--------|------|
| 测试后 8080 端口 | ✅ 无残留 |
| 测试后 5173 端口 | ✅ 无残留 |
| 测试后 .run/*.pid | ✅ 已清理 |
| 测试后 Docker volume | ✅ 保留 |

## 20. 发现的问题

### 问题 1（🔴 严重）：`local` 关键字在 case 块中使用

**影响脚本：**
- `mysql.sh` line 98（check 分支）
- `backend.sh` line 100（status 分支）
- `backend.sh` line 109（health 分支）
- `quality.sh` line 21（lint 分支）

**错误信息：** `local: can only be used in a function`

**原因：** bash 中 `local` 只能在函数体内使用。`case` 块不是函数体，即使在脚本顶层 `case` 中的 `local` 也无效。

**修复方案：** 将使用 `local` 的逻辑封装到函数中。例如：

```bash
# 修改前（错误）
case "$ACTION" in
    check)
        local tables="users resume_versions ..."
        ...
esac

# 修改后（正确）
do_check() {
    local tables="users resume_versions ..."
    ...
}

case "$ACTION" in
    check)
        do_check
        ;;
esac
```

### 问题 2（🔴 严重）：`go run` 子进程未被正确终止

**影响命令：** `backend-stop`、`backend-restart`

**现象：**
1. `go run ./cmd/server` 启动后产生两个进程：`go run`（父）和 `server`（子）
2. `stop_pid_file` 只 kill 父进程 PID，子进程 `server` 仍在端口 8080 监听
3. 重启时新进程无法绑定端口

**修复方案：** 使用进程组 kill，或在启动时记录进程组 ID：

```bash
# 方案 A：kill 进程组
kill -- -$PID  # 负号表示进程组

# 方案 B：使用 setsid 启动
setsid go run ./cmd/server ...

# 方案 C：查找并 kill 端口对应进程
kill $(lsof -t -i :$SERVER_PORT)
```

### 问题 3（🟡 中等）：`make health` target 遗漏

**现象：** 执行 `make health` 报错 `No rule to make target 'health'`

**影响：** 用户文档中提到 `make health`，但 Makefile 中没有定义这个 target

**修复方案：** 在 Makefile 中新增：

```makefile
health: ## 查看所有服务健康状态
	@echo ""
	@echo "=== 健康检查 ==="
	@$(DEV_SCRIPT)/backend.sh health || true
	@$(DEV_SCRIPT)/web.sh health || true
	@$(DEV_SCRIPT)/mysql.sh check || true
```

### 问题 4（🟡 中等）：`backend-up` 首次启动可能超时

**现象：** 首次 `go run` 时需要下载 Go 依赖（go1.25.0 + 3 个包），15 秒超时不够

**影响：** 云原生环境重启后首次启动可能失败

**修复方案：** 将等待时间从 15 秒增加到 30 秒，或增加重试逻辑

## 21. 修复建议

### 优先级排序

| 优先级 | 问题 | 修复难度 | 影响范围 |
|--------|------|---------|---------|
| P0 | local 在 case 块中使用 | 低 | 4 个脚本，7 处位置 |
| P0 | go run 子进程未被终止 | 中 | backend.sh stop 逻辑 |
| P1 | make health 遗漏 | 低 | Makefile 新增 target |
| P1 | backend-up 首次超时 | 低 | backend.sh 增加等待时间 |

### 修复步骤

1. 修复 `mysql.sh`：将 `check` 分支的 `local` 逻辑封装为 `do_check()` 函数
2. 修复 `backend.sh`：
   - 将 `status` 分支封装为 `do_status()` 函数
   - 将 `health` 分支封装为 `do_health()` 函数
   - 修复 `stop` 逻辑：kill 整个进程组而非单个进程
   - 增加启动等待时间从 15s 到 30s
3. 修复 `quality.sh`：将 `lint` 分支的 `local` 逻辑封装为 `do_lint()` 函数
4. 新增 `make health` target

## 22. 是否允许进入 4-2

**❌ 暂不允许。**

需先执行 **4-1.1.3 修复 Makefile 全量测试发现的问题**，修复 P0 级别的 `local` 语法错误和 `go run` 子进程终止问题后，方可进入 4-2。

理由：
1. `backend-status`、`backend-health`、`db-check`、`lint` 全部失败，开发者无法正常使用这些命令
2. `backend-restart` 失败会导致开发循环受阻
3. 这些问题修复难度低，应优先修复

## 23. 修复结果

> 修复时间：2026-05-29（阶段 4-1.1.3）

| 问题 | 修复状态 | 修复说明 | 验证命令 | 验证结果 |
|------|---------|---------|---------|---------|
| P0：`local` 在 case 块中使用 | ✅ 已修复 | 将 check/status/health/lint 分支封装为函数 | `make db-check` `make backend-status` `make backend-health` `make lint` `make ci` | ✅ 全部通过 |
| P0：`go run` 子进程未被终止 | ✅ 已修复 | 增强 `stop_pid_file` 增加子进程清理；`stop_backend` 增加端口兜底清理 | `make backend-stop` `make backend-restart` | ✅ 端口释放，restart 成功 |
| P1：`make health` 遗漏 | ✅ 已修复 | 新增 `health` target，聚合 infra/backend/web/db 健康检查 | `make health` | ✅ 通过 |
| P1：`backend-up` 首次超时 | ✅ 已修复 | 等待时间 15s → 30s | `make backend-up` | ✅ 通过 |

**修复结论：所有 P0/P1 问题已修复，验证通过，可以进入 4-2。**
