# ============================================================
# Offer Catcher - Agents 服务 Makefile（独立）
# 适配 CNB 云原生环境：make start 一键编排
# ============================================================

SHELL := /usr/bin/env bash
.DEFAULT_GOAL := help

APP_NAME := offer-catcher-agents
BIN      := bin/agents
PORT     ?= 18080
PID_FILE := .run/agents.pid
LOG_FILE := logs/agents.log

# 从 .env 读取（缺失时回退默认值），避免 -include .env 的空值坑。
DB_USER     := $(shell grep -E '^DB_USER=' .env 2>/dev/null | head -1 | cut -d'=' -f2- || echo root)
DB_PASSWORD := $(shell grep -E '^DB_PASSWORD=' .env 2>/dev/null | head -1 | cut -d'=' -f2- || echo password)

.PHONY: help build run start stop restart health status logs db-init clean

help: ## 查看所有可用命令
	@echo ""
	@echo "[$(APP_NAME)] Makefile 命令"
	@echo "=============================="
	@echo "  make start     一键启动（确保 env → 起中间件 → 建库 → 起后端 → 健康检查）"
	@echo "  make stop      停止服务"
	@echo "  make restart   重启服务"
	@echo "  make health    健康检查 curl /healthz"
	@echo "  make status    查看进程状态"
	@echo "  make logs      查看服务日志"
	@echo "  make build     编译二进制"
	@echo "  make run       前台运行（开发调试）"
	@echo "  make db-init   创建数据库 offer_catcher_re"
	@echo "  make clean     清理运行产物"
	@echo ""

build: ## 编译二进制
	@mkdir -p bin
	@go build -o $(BIN) .
	@echo "[ok] 编译完成: $(BIN)"

run: ## 前台运行（开发调试）
	@go run .

start: ## 一键启动
	@echo "[1/6] 确保 .env 存在..."
	@[ -f .env ] || cp .env.example .env
	@echo "[2/6] 启动中间件 (MySQL)..."
	@docker compose -f ../docker-compose.yml up -d mysql
	@echo "[3/6] 等待 MySQL 就绪..."
	@for i in $$(seq 1 60); do \
		if docker exec offer-catcher-mysql mysqladmin ping -h localhost -u$(DB_USER) -p$(DB_PASSWORD) --silent >/dev/null 2>&1; then \
			break; \
		fi; \
		sleep 1; \
	done
	@echo "[4/6] 创建数据库..."
	@docker exec -i offer-catcher-mysql mysql -u$(DB_USER) -p$(DB_PASSWORD) < sql/init.sql
	@echo "[5/6] 编译并启动 $(APP_NAME)..."
	@mkdir -p bin logs .run
	@go build -o $(BIN) .
	@nohup ./$(BIN) >> $(LOG_FILE) 2>&1 & echo $$! > $(PID_FILE)
	@echo "[6/6] 健康检查..."
	@for i in $$(seq 1 30); do \
		if curl -sf http://localhost:$(PORT)/healthz >/dev/null 2>&1; then \
			echo "[ok] 启动成功 http://localhost:$(PORT)/healthz"; \
			exit 0; \
		fi; \
		sleep 1; \
	done; \
	echo "[error] 启动失败，请查看日志: make logs"; \
	exit 1

stop: ## 停止服务
	@if [ -f $(PID_FILE) ] && kill -0 $$(cat $(PID_FILE)) 2>/dev/null; then \
		kill $$(cat $(PID_FILE)) 2>/dev/null || true; \
		rm -f $(PID_FILE); \
		echo "[ok] 已停止"; \
	else \
		echo "[info] $(APP_NAME) 未运行"; \
		rm -f $(PID_FILE); \
	fi

restart: ## 重启服务
	@$(MAKE) --no-print-directory stop
	@sleep 1
	@$(MAKE) --no-print-directory start

health: ## 健康检查
	@curl -sf http://localhost:$(PORT)/healthz && echo "" || { echo "[error] 健康检查失败"; exit 1; }

status: ## 查看进程状态
	@if [ -f $(PID_FILE) ] && kill -0 $$(cat $(PID_FILE)) 2>/dev/null; then \
		echo "[status] $(APP_NAME) 运行中 (PID: $$(cat $(PID_FILE)), 端口: $(PORT))"; \
	else \
		echo "[status] $(APP_NAME) 未运行"; \
	fi

logs: ## 查看服务日志
	@if [ -f $(LOG_FILE) ]; then tail -f $(LOG_FILE); else echo "[warn] 日志文件不存在: $(LOG_FILE)"; fi

db-init: ## 创建数据库 offer_catcher_re（幂等）
	@docker exec -i offer-catcher-mysql mysql -u$(DB_USER) -p$(DB_PASSWORD) < sql/init.sql
	@echo "[ok] 数据库就绪"

clean: ## 清理运行产物
	@rm -rf bin .run logs
	@echo "[ok] 已清理"
