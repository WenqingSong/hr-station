# ============================================================
# Offer Catcher - Makefile
# 适配 CNB 云原生开发环境反复重启场景
# ============================================================

.PHONY: help doctor init \
        up down restart status logs logs-mysql \
        dev-web dev-server dev \
        build-web build-server build \
        test-server check \
        clean reset

# ============================================================
# 默认目标
# ============================================================

help: ## 查看所有可用命令
	@echo ""
	@echo "Offer Catcher - Makefile 命令"
	@echo "=============================="
	@echo ""
	@echo "环境管理:"
	@echo "  make doctor       只读检查开发环境"
	@echo "  make init         初始化开发环境（幂等，可重复执行）"
	@echo ""
	@echo "Docker 服务:"
	@echo "  make up           启动 Docker Compose 服务"
	@echo "  make down         停止 Docker Compose 服务（不删卷）"
	@echo "  make restart      重启 Docker Compose 服务"
	@echo "  make status       查看容器状态"
	@echo "  make logs         查看全部服务日志"
	@echo "  make logs-mysql   查看 MySQL 日志"
	@echo ""
	@echo "开发运行:"
	@echo "  make dev-web      启动前端开发服务"
	@echo "  make dev-server   启动后端开发服务"
	@echo "  make dev          输出开发启动提示"
	@echo ""
	@echo "构建:"
	@echo "  make build-web    构建前端"
	@echo "  make build-server 构建后端"
	@echo "  make build        构建前后端"
	@echo ""
	@echo "测试与检查:"
	@echo "  make test-server  运行后端测试"
	@echo "  make check        执行基础检查"
	@echo ""
	@echo "清理:"
	@echo "  make clean        清理构建产物"
	@echo "  make reset        ⚠️ 危险：停止并清理容器卷（删除数据）"
	@echo ""

# ============================================================
# 环境检查
# ============================================================

doctor: ## 只读检查开发环境
	@echo "=========================================="
	@echo "  开发环境检查 (只读，不修改任何文件)"
	@echo "=========================================="
	@echo ""
	@echo "--- 工作目录 ---"
	@echo "当前目录: $$(pwd)"
	@echo ""
	@echo "--- 工具检查 ---"
	@command -v node >/dev/null 2>&1 && echo "✅ node: $$(node --version)" || echo "❌ node: 未安装"
	@command -v npm >/dev/null 2>&1 && echo "✅ npm: $$(npm --version)" || echo "❌ npm: 未安装"
	@command -v go >/dev/null 2>&1 && echo "✅ go: $$(go version | awk '{print $$3}')" || echo "❌ go: 未安装"
	@command -v docker >/dev/null 2>&1 && echo "✅ docker: $$(docker --version | awk '{print $$3}' | tr -d ',')" || echo "❌ docker: 未安装"
	@command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1 && echo "✅ docker compose: 可用" || echo "❌ docker compose: 不可用"
	@echo ""
	@echo "--- 项目文件检查 ---"
	@test -f .env && echo "✅ .env: 存在" || echo "⚠️  .env: 不存在（执行 make init 可创建）"
	@test -f .env.example && echo "✅ .env.example: 存在" || echo "⚠️  .env.example: 不存在（待阶段 1-3 创建）"
	@test -f docker-compose.yml && echo "✅ docker-compose.yml: 存在" || echo "❌ docker-compose.yml: 不存在"
	@test -f web/package.json && echo "✅ web/package.json: 存在" || echo "❌ web/package.json: 不存在"
	@test -d web/node_modules && echo "✅ web/node_modules: 存在" || echo "⚠️  web/node_modules: 不存在（需安装前端依赖）"
	@test -f server/go.mod && echo "✅ server/go.mod: 存在" || echo "⚠️  server/go.mod: 不存在（待阶段 1-5 初始化）"
	@echo ""
	@echo "--- Docker 服务状态 ---"
	@docker compose ps 2>/dev/null || echo "⚠️  docker compose 不可用或 docker-compose.yml 为空"
	@echo ""
	@echo "--- 端口占用检查 ---"
	@lsof -i :3306 >/dev/null 2>&1 && echo "⚠️  端口 3306: 已占用" || echo "✅ 端口 3306: 可用"
	@lsof -i :5173 >/dev/null 2>&1 && echo "⚠️  端口 5173: 已占用" || echo "✅ 端口 5173: 可用"
	@lsof -i :8080 >/dev/null 2>&1 && echo "⚠️  端口 8080: 已占用" || echo "✅ 端口 8080: 可用"
	@echo ""
	@echo "=========================================="
	@echo "  检查完成"
	@echo "=========================================="

# ============================================================
# 初始化开发环境
# ============================================================

init: ## 初始化开发环境（幂等，可重复执行）
	@echo "=========================================="
	@echo "  初始化开发环境"
	@echo "=========================================="
	@echo ""
	@echo "--- 1. 检查 .env ---"
	@if [ -f .env ]; then \
		echo "✅ .env 已存在，跳过"; \
	elif [ -f .env.example ]; then \
		cp .env.example .env; \
		echo "✅ 已从 .env.example 复制创建 .env"; \
	else \
		echo "⚠️  .env.example 不存在（待阶段 1-3 创建），请稍后手动创建 .env"; \
	fi
	@echo ""
	@echo "--- 2. 检查前端依赖 ---"
	@if [ -f web/package.json ]; then \
		if [ -d web/node_modules ]; then \
			echo "✅ web/node_modules 已存在，跳过安装"; \
		else \
			echo "⚠️  web/node_modules 不存在，请执行: cd web && npm install"; \
		fi; \
	else \
		echo "⚠️  web/package.json 不存在，跳过前端依赖检查"; \
	fi
	@echo ""
	@echo "--- 3. 检查后端依赖 ---"
	@if [ -f server/go.mod ]; then \
		echo "✅ server/go.mod 已存在"; \
		echo "   如需下载依赖，请执行: cd server && go mod download"; \
	else \
		echo "⚠️  server/go.mod 不存在（待阶段 1-5 初始化 Go 工程）"; \
	fi
	@echo ""
	@echo "=========================================="
	@echo "  初始化完成"
	@echo "  下一步: make up    (启动 Docker 服务)"
	@echo "          make dev   (查看开发启动方式)"
	@echo "=========================================="

# ============================================================
# Docker Compose 服务管理
# ============================================================

up: ## 启动 Docker Compose 服务
	@if [ ! -f docker-compose.yml ]; then \
		echo "❌ docker-compose.yml 不存在，无法启动"; \
		exit 1; \
	fi
	@if [ ! -s docker-compose.yml ]; then \
		echo "⚠️  docker-compose.yml 为空（待阶段 1-4 编写），无法启动服务"; \
		echo "   请先完成阶段 1-4：编写 docker-compose.yml"; \
		exit 0; \
	fi
	docker compose up -d
	@echo ""
	@echo "✅ Docker Compose 服务已启动"
	@echo "   查看状态: make status"

down: ## 停止 Docker Compose 服务（不删卷）
	docker compose down
	@echo ""
	@echo "✅ Docker Compose 服务已停止（数据卷保留）"

restart: ## 重启 Docker Compose 服务
	docker compose down
	docker compose up -d
	@echo ""
	@echo "✅ Docker Compose 服务已重启"

status: ## 查看容器状态
	docker compose ps

logs: ## 查看全部服务日志
	docker compose logs -f

logs-mysql: ## 查看 MySQL 日志
	docker compose logs -f mysql

# ============================================================
# 开发运行
# ============================================================

dev-web: ## 启动前端开发服务
	@if [ ! -f web/package.json ]; then \
		echo "❌ web/package.json 不存在"; \
		exit 1; \
	fi
	@if [ ! -d web/node_modules ]; then \
		echo "⚠️  web/node_modules 不存在，请先安装依赖: cd web && npm install"; \
		exit 1; \
	fi
	cd web && npm run dev

dev-server: ## 启动后端开发服务
	@if [ ! -f server/go.mod ]; then \
		echo "⚠️  server/go.mod 不存在（待阶段 1-5 初始化 Go 工程）"; \
		echo "   请先完成阶段 1-5：初始化 Go 后端工程目录"; \
		exit 1; \
	fi
	cd server && go run ./cmd/server

dev: ## 输出开发启动提示
	@echo ""
	@echo "=========================================="
	@echo "  Offer Catcher 开发环境启动"
	@echo "=========================================="
	@echo ""
	@echo "请分别在不同终端运行以下命令："
	@echo ""
	@echo "  终端 1 - 启动 Docker 服务:"
	@echo "    make up"
	@echo ""
	@echo "  终端 2 - 启动前端:"
	@echo "    make dev-web"
	@echo ""
	@echo "  终端 3 - 启动后端:"
	@echo "    make dev-server"
	@echo ""
	@echo "CNB 环境重启恢复流程:"
	@echo "    make doctor → make init → make up → make dev-web → make dev-server"
	@echo ""
	@echo "=========================================="

# ============================================================
# 构建
# ============================================================

build-web: ## 构建前端
	@if [ ! -f web/package.json ]; then \
		echo "❌ web/package.json 不存在"; \
		exit 1; \
	fi
	@if [ ! -d web/node_modules ]; then \
		echo "⚠️  web/node_modules 不存在，请先安装依赖: cd web && npm install"; \
		exit 1; \
	fi
	cd web && npm run build
	@echo "✅ 前端构建完成"

build-server: ## 构建后端
	@if [ ! -f server/go.mod ]; then \
		echo "⚠️  server/go.mod 不存在（待阶段 1-5 初始化 Go 工程）"; \
		exit 1; \
	fi
	cd server && go build -o bin/server ./cmd/server
	@echo "✅ 后端构建完成: server/bin/server"

build: build-web build-server ## 构建前后端

# ============================================================
# 测试与检查
# ============================================================

test-server: ## 运行后端测试
	@if [ ! -f server/go.mod ]; then \
		echo "⚠️  server/go.mod 不存在（待阶段 1-5 初始化 Go 工程）"; \
		exit 1; \
	fi
	cd server && go test ./...

check: doctor ## 执行基础检查（环境检查）
	@echo ""
	@echo "--- 后端测试 ---"
	@if [ -f server/go.mod ]; then \
		cd server && go test ./...; \
	else \
		echo "⚠️  server/go.mod 不存在，跳过后端测试"; \
	fi

# ============================================================
# 清理与重置
# ============================================================

clean: ## 清理构建产物
	@echo "清理构建产物..."
	rm -rf web/dist
	rm -rf server/bin
	@echo "✅ 构建产物已清理"

reset: ## ⚠️ 危险：停止并清理容器卷（删除数据）
	@echo "=========================================="
	@echo "  ⚠️  危险操作：即将停止所有容器并删除数据卷"
	@echo "  这将清除 MySQL 中的所有数据！"
	@echo "=========================================="
	@echo ""
	@echo "如果确定要执行，请运行:"
	@echo "  docker compose down -v"
	@echo ""
	@echo "该命令不会自动执行，请手动确认后运行。"
