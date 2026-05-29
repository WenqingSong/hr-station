# ============================================================
# Offer Catcher - Makefile
# 适配 CNB 云原生开发环境反复重启场景
# Makefile = 统一入口 / 命令编排层
# scripts/dev/*.sh = 具体执行逻辑
# ============================================================

SHELL := /usr/bin/env bash
.DEFAULT_GOAL := help

DEV_SCRIPT := scripts/dev

# ============================================================
# 命令声明
# ============================================================

.PHONY: help doctor init check status health \
        up stop down restart rebuild \
        infra-up infra-down infra-status infra-logs \
        db-import db-reset db-shell db-check \
        backend-up backend-stop backend-restart backend-logs backend-status backend-health \
        web-install web-up web-stop web-restart web-logs web-status web-health \
        agent-env agent-check agent-test \
        logs logs-backend logs-web logs-mysql logs-agent \
        lint test build ci \
        clean clean-logs clean-pids clean-volume \
        jobs-help jobs-check jobs-source-check \
        jobs-fetch-list jobs-fetch-detail jobs-fetch-details \
        jobs-fetch-dictionary jobs-fetch jobs-sample jobs-status jobs-clean \
        jobs-import-help jobs-import-check jobs-import-preview jobs-import-sql \
        jobs-import jobs-import-status jobs-db-status jobs-import-clean

# ============================================================
# 帮助
# ============================================================

help: ## 查看所有可用命令
	@echo ""
	@echo "Offer Catcher - Makefile 命令"
	@echo "=============================="
	@echo ""
	@echo "基础命令:"
	@echo "  make help          查看所有可用命令"
	@echo "  make doctor        检查开发环境（Docker/Go/Node/npm/.env/端口/目录）"
	@echo "  make init          初始化开发环境（创建目录、复制 .env、提示依赖安装）"
	@echo "  make check         执行 doctor + 轻量检查"
	@echo "  make status        查看所有服务运行状态"
	@echo "  make health        查看所有服务健康状态"
	@echo ""
	@echo "完整生命周期:"
	@echo "  make up            一键启动完整开发环境"
	@echo "  make stop          停止本地前后端进程，保留 Docker 数据"
	@echo "  make down          停止服务，保留数据库 volume，不清空数据"
	@echo "  make restart       stop + up"
	@echo "  make rebuild       ⚠️ 危险：清理运行态并重建，需要 CONFIRM=yes"
	@echo ""
	@echo "基础设施:"
	@echo "  make infra-up      启动 Docker Compose 基础设施（MySQL）"
	@echo "  make infra-down    停止 Docker Compose 服务，保留 volume"
	@echo "  make infra-status  查看 Docker Compose 状态"
	@echo "  make infra-logs    查看 Docker Compose 日志"
	@echo ""
	@echo "数据库:"
	@echo "  make db-import     导入 server/sql/init.sql"
	@echo "  make db-reset      ⚠️ 危险：重新导入初始化 SQL，需要 CONFIRM=yes"
	@echo "  make db-shell      进入 MySQL shell"
	@echo "  make db-check      检查核心表是否存在、记录数是否正常"
	@echo ""
	@echo "后端:"
	@echo "  make backend-up       后台启动 Go 后端"
	@echo "  make backend-stop     停止 Go 后端"
	@echo "  make backend-restart  重启 Go 后端"
	@echo "  make backend-logs     查看后端日志"
	@echo "  make backend-status   查看后端进程状态"
	@echo "  make backend-health   curl /healthz"
	@echo ""
	@echo "前端:"
	@echo "  make web-install      在 web/ 下执行 npm install"
	@echo "  make web-up           后台启动前端开发服务"
	@echo "  make web-stop         停止前端服务"
	@echo "  make web-restart      重启前端服务"
	@echo "  make web-logs         查看前端日志"
	@echo "  make web-status       查看前端进程状态"
	@echo "  make web-health       检查前端端口"
	@echo ""
	@echo "Agent / LLM:"
	@echo "  make agent-env        检查 LLM 相关环境变量（不打印真实 Key）"
	@echo "  make agent-check      检查 ENABLE_LLM_AGENT / LLM_PROVIDER / LLM_MODEL"
	@echo "  make agent-test       当前只输出提示：LLM Client 尚未实现"
	@echo ""
	@echo "日志:"
	@echo "  make logs           查看日志入口提示"
	@echo "  make logs-backend   查看后端日志"
	@echo "  make logs-web       查看前端日志"
	@echo "  make logs-mysql     查看 MySQL 日志"
	@echo "  make logs-agent     查看 Agent 相关日志"
	@echo ""
	@echo "质量检查:"
	@echo "  make lint           代码风格检查"
	@echo "  make test           运行后端测试"
	@echo "  make build          构建检查（前端 + 后端）"
	@echo "  make ci             doctor + lint + test + build"
	@echo ""
	@echo "清理:"
	@echo "  make clean          清理本地运行文件和构建产物，不删除数据库 volume"
	@echo "  make clean-logs     删除 .logs"
	@echo "  make clean-pids     删除 .run"
	@echo "  make clean-volume   ⚠️ 危险：删除 Docker volume，需要 CONFIRM=yes"
	@echo ""
	@echo "腾讯岗位数据源:"
	@echo "  make jobs-help              查看腾讯岗位数据源命令帮助"
	@echo "  make jobs-check             检查腾讯岗位数据源配置"
	@echo "  make jobs-source-check      同 jobs-check"
	@echo "  make jobs-fetch-list        拉取腾讯岗位列表原始 JSON"
	@echo "  make jobs-fetch-detail      拉取腾讯岗位详情原始 JSON"
	@echo "  make jobs-fetch-details     同 jobs-fetch-detail"
	@echo "  make jobs-fetch-dictionary  拉取腾讯招聘字典原始 JSON"
	@echo "  make jobs-fetch             一键拉取列表 + 详情 + 字典"
	@echo "  make jobs-sample            查看本地岗位样本摘要"
	@echo "  make jobs-status            查看本地原始数据文件状态"
	@echo "  make jobs-clean             清理本地腾讯岗位原始 JSON"
	@echo ""
	@echo "腾讯岗位入库:"
	@echo "  make jobs-import-help       查看腾讯岗位入库命令帮助"
	@echo "  make jobs-import-check      检查 raw JSON、MySQL、输出目录"
	@echo "  make jobs-import-preview    预览将导入的腾讯岗位"
	@echo "  make jobs-import-sql        生成腾讯岗位导入 SQL"
	@echo "  make jobs-import            执行腾讯岗位导入 MySQL"
	@echo "  make jobs-import-status     查看腾讯岗位导入状态"
	@echo "  make jobs-db-status         同 jobs-import-status"
	@echo "  make jobs-import-clean      清理 generated 目录中的预览和 SQL"
	@echo ""

# ============================================================
# 基础命令
# ============================================================

doctor: ## 检查开发环境
	@$(DEV_SCRIPT)/doctor.sh

init: ## 初始化开发环境
	@$(DEV_SCRIPT)/env.sh init

check: doctor ## 执行 doctor + 健康检查
	@echo ""
	@echo "--- 健康检查 ---"
	@$(MAKE) --no-print-directory health

status: ## 查看所有服务运行状态
	@echo ""
	@echo "=== 服务运行状态 ==="
	@echo ""
	@echo "--- 基础设施 ---"
	@$(DEV_SCRIPT)/infra.sh status || true
	@echo ""
	@echo "--- 后端 ---"
	@$(DEV_SCRIPT)/backend.sh status || true
	@echo ""
	@echo "--- 前端 ---"
	@$(DEV_SCRIPT)/web.sh status || true

health: ## 查看所有服务健康状态
	@echo ""
	@echo "=== 健康检查 ==="
	@echo ""
	@echo "--- 基础设施 ---"
	@$(DEV_SCRIPT)/infra.sh status || true
	@echo ""
	@echo "--- 后端 ---"
	@$(DEV_SCRIPT)/backend.sh health || true
	@echo ""
	@echo "--- 前端 ---"
	@$(DEV_SCRIPT)/web.sh health || true
	@echo ""
	@echo "--- 数据库 ---"
	@$(DEV_SCRIPT)/mysql.sh check || true

# ============================================================
# 完整生命周期
# ============================================================

up: ## 一键启动完整开发环境
	@echo ""
	@echo "=========================================="
	@echo "  Offer Catcher - 一键启动开发环境"
	@echo "=========================================="
	@echo ""
	@echo "--- 1. 环境检查 ---"
	@$(DEV_SCRIPT)/doctor.sh
	@echo ""
	@echo "--- 2. 环境初始化 ---"
	@$(DEV_SCRIPT)/env.sh init
	@echo ""
	@echo "--- 3. 启动基础设施 ---"
	@$(DEV_SCRIPT)/infra.sh up
	@echo ""
	@echo "--- 4. 等待 MySQL 就绪 ---"
	@sleep 5
	@$(DEV_SCRIPT)/mysql.sh check || echo "  提示：如果数据库为空，请执行 make db-import"
	@echo ""
	@echo "--- 5. 启动后端 ---"
	@$(DEV_SCRIPT)/backend.sh up || echo "  ⚠️ 后端启动失败，可能是 MySQL 尚未就绪，稍后重试"
	@echo ""
	@echo "--- 6. 启动前端 ---"
	@$(DEV_SCRIPT)/web.sh up || echo "  ⚠️ 前端启动失败，可能需要先执行 make web-install"
	@echo ""
	@echo "--- 7. 健康检查 ---"
	@sleep 3
	@$(DEV_SCRIPT)/backend.sh health || true
	@$(DEV_SCRIPT)/web.sh health || true
	@echo ""
	@echo "=========================================="
	@echo "  ✅ 开发环境启动完成"
	@echo "=========================================="
	@echo ""
	@echo "  后端:  http://localhost:8080/healthz"
	@echo "  前端:  http://localhost:5173"
	@echo "  MySQL: localhost:3306"
	@echo ""
	@echo "  查看日志: make logs"
	@echo "  查看状态: make backend-status / make web-status"
	@echo ""

stop: ## 停止本地前后端进程，保留 Docker 数据
	@echo "停止本地进程..."
	@$(DEV_SCRIPT)/backend.sh stop
	@$(DEV_SCRIPT)/web.sh stop
	@echo ""
	@echo "✅ 前后端进程已停止（Docker 数据保留）"

down: ## 停止服务，保留数据库 volume
	@$(DEV_SCRIPT)/backend.sh stop || true
	@$(DEV_SCRIPT)/web.sh stop || true
	@$(DEV_SCRIPT)/infra.sh down

restart: ## stop + up
	@$(DEV_SCRIPT)/backend.sh stop || true
	@$(DEV_SCRIPT)/web.sh stop || true
	@sleep 1
	@$(MAKE) up

rebuild: ## ⚠️ 危险：清理运行态并重建，需要 CONFIRM=yes
	@if [ "$(CONFIRM)" != "yes" ]; then \
		echo ""; \
		echo "危险操作：该命令会重置本地开发环境。"; \
		echo "如确认执行，请使用：make rebuild CONFIRM=yes"; \
		echo ""; \
		exit 1; \
	fi
	@echo "清理运行态..."
	@$(DEV_SCRIPT)/backend.sh stop || true
	@$(DEV_SCRIPT)/web.sh stop || true
	@$(DEV_SCRIPT)/clean.sh clean
	@$(DEV_SCRIPT)/infra.sh down
	@echo ""
	@$(MAKE) up

# ============================================================
# 基础设施
# ============================================================

infra-up: ## 启动 Docker Compose 基础设施
	@$(DEV_SCRIPT)/infra.sh up

infra-down: ## 停止 Docker Compose 服务
	@$(DEV_SCRIPT)/infra.sh down

infra-status: ## 查看 Docker Compose 状态
	@$(DEV_SCRIPT)/infra.sh status

infra-logs: ## 查看 Docker Compose 日志
	@$(DEV_SCRIPT)/infra.sh logs

# ============================================================
# 数据库
# ============================================================

db-import: ## 导入 server/sql/init.sql
	@$(DEV_SCRIPT)/mysql.sh import

db-reset: ## ⚠️ 危险：重新导入初始化 SQL，需要 CONFIRM=yes
	@$(DEV_SCRIPT)/mysql.sh reset

db-shell: ## 进入 MySQL shell
	@$(DEV_SCRIPT)/mysql.sh shell

db-check: ## 检查核心表
	@$(DEV_SCRIPT)/mysql.sh check

# ============================================================
# 后端
# ============================================================

backend-up: ## 后台启动 Go 后端
	@$(DEV_SCRIPT)/backend.sh up

backend-stop: ## 停止 Go 后端
	@$(DEV_SCRIPT)/backend.sh stop

backend-restart: ## 重启 Go 后端
	@$(DEV_SCRIPT)/backend.sh restart

backend-logs: ## 查看后端日志
	@$(DEV_SCRIPT)/backend.sh logs

backend-status: ## 查看后端进程状态
	@$(DEV_SCRIPT)/backend.sh status

backend-health: ## curl /healthz
	@$(DEV_SCRIPT)/backend.sh health

# ============================================================
# 前端
# ============================================================

web-install: ## 在 web/ 下执行 npm install
	@$(DEV_SCRIPT)/web.sh install

web-up: ## 后台启动前端开发服务
	@$(DEV_SCRIPT)/web.sh up

web-stop: ## 停止前端服务
	@$(DEV_SCRIPT)/web.sh stop

web-restart: ## 重启前端服务
	@$(DEV_SCRIPT)/web.sh restart

web-logs: ## 查看前端日志
	@$(DEV_SCRIPT)/web.sh logs

web-status: ## 查看前端进程状态
	@$(DEV_SCRIPT)/web.sh status

web-health: ## 检查前端端口
	@$(DEV_SCRIPT)/web.sh health

# ============================================================
# Agent / LLM
# ============================================================

agent-env: ## 检查 LLM 相关环境变量
	@$(DEV_SCRIPT)/agent.sh env

agent-check: ## 检查 Agent/LLM 配置
	@$(DEV_SCRIPT)/agent.sh check

agent-test: ## 当前只输出提示
	@$(DEV_SCRIPT)/agent.sh test

# ============================================================
# 日志
# ============================================================

logs: ## 查看日志入口提示
	@$(DEV_SCRIPT)/logs.sh all

logs-backend: ## 查看后端日志
	@$(DEV_SCRIPT)/logs.sh backend

logs-web: ## 查看前端日志
	@$(DEV_SCRIPT)/logs.sh web

logs-mysql: ## 查看 MySQL 日志
	@$(DEV_SCRIPT)/logs.sh mysql

logs-agent: ## 查看 Agent 相关日志
	@$(DEV_SCRIPT)/logs.sh agent

# ============================================================
# 质量检查
# ============================================================

lint: ## 代码风格检查
	@$(DEV_SCRIPT)/quality.sh lint

test: ## 运行后端测试
	@$(DEV_SCRIPT)/quality.sh test

build: ## 构建检查
	@$(DEV_SCRIPT)/quality.sh build

ci: ## doctor + lint + test + build
	@$(DEV_SCRIPT)/quality.sh ci

# ============================================================
# 清理
# ============================================================

clean: ## 清理本地运行文件和构建产物
	@$(DEV_SCRIPT)/clean.sh clean

clean-logs: ## 删除 .logs
	@$(DEV_SCRIPT)/clean.sh logs

clean-pids: ## 删除 .run
	@$(DEV_SCRIPT)/clean.sh pids

clean-volume: ## ⚠️ 危险：删除 Docker volume，需要 CONFIRM=yes
	@$(DEV_SCRIPT)/clean.sh volume

# ============================================================
# 腾讯岗位数据源
# ============================================================

jobs-help: ## 查看腾讯岗位数据源命令帮助
	@$(DEV_SCRIPT)/tencent_jobs.sh help

jobs-check: ## 检查腾讯岗位数据源配置
	@$(DEV_SCRIPT)/tencent_jobs.sh check

jobs-source-check: ## 同 jobs-check
	@$(DEV_SCRIPT)/tencent_jobs.sh check

jobs-fetch-list: ## 拉取腾讯岗位列表原始 JSON
	@$(DEV_SCRIPT)/tencent_jobs.sh fetch-list

jobs-fetch-detail: ## 拉取腾讯岗位详情原始 JSON
	@$(DEV_SCRIPT)/tencent_jobs.sh fetch-details

jobs-fetch-details: ## 同 jobs-fetch-detail
	@$(DEV_SCRIPT)/tencent_jobs.sh fetch-details

jobs-fetch-dictionary: ## 拉取腾讯招聘字典原始 JSON
	@$(DEV_SCRIPT)/tencent_jobs.sh fetch-dictionary

jobs-fetch: ## 一键拉取列表 + 详情 + 字典
	@$(DEV_SCRIPT)/tencent_jobs.sh fetch-all

jobs-sample: ## 查看本地岗位样本摘要
	@$(DEV_SCRIPT)/tencent_jobs.sh sample

jobs-status: ## 查看本地原始数据文件状态
	@$(DEV_SCRIPT)/tencent_jobs.sh status

jobs-clean: ## 清理本地腾讯岗位原始 JSON
	@$(DEV_SCRIPT)/tencent_jobs.sh clean

# ============================================================
# 腾讯岗位入库
# ============================================================

jobs-import-help: ## 查看腾讯岗位入库命令帮助
	@$(DEV_SCRIPT)/tencent_jobs_import.sh help

jobs-import-check: ## 检查 raw JSON、MySQL、输出目录
	@$(DEV_SCRIPT)/tencent_jobs_import.sh check

jobs-import-preview: ## 预览将导入的腾讯岗位
	@$(DEV_SCRIPT)/tencent_jobs_import.sh preview

jobs-import-sql: ## 生成腾讯岗位导入 SQL
	@$(DEV_SCRIPT)/tencent_jobs_import.sh generate-sql

jobs-import: ## 执行腾讯岗位导入 MySQL
	@$(DEV_SCRIPT)/tencent_jobs_import.sh import

jobs-import-status: ## 查看腾讯岗位导入状态
	@$(DEV_SCRIPT)/tencent_jobs_import.sh status

jobs-db-status: ## 同 jobs-import-status
	@$(DEV_SCRIPT)/tencent_jobs_import.sh status

jobs-import-clean: ## 清理 generated 目录中的预览和 SQL
	@$(DEV_SCRIPT)/tencent_jobs_import.sh clean-generated
