# ============================================================
# Offer Catcher - Agents 服务 Makefile（独立）
# ============================================================

SHELL := /usr/bin/env bash
.DEFAULT_GOAL := help

# ---- 环境变量加载 ----
# 优先加载本地 .env（若存在）；不存在时使用下方默认值。
# 优先级：shell 环境 > .env > 默认值。
-include .env

# ---- 变量 ----
APP_NAME := offer-catcher-agents
BIN      := bin/agents
PORT     := $(or $(AGENTS_PORT),18080)
PID_FILE := .run/agents.pid
LOG_FILE := logs/agents.log

DB_HOST     ?= 127.0.0.1
DB_PORT     ?= 3306
DB_USER     ?= root
DB_PASSWORD ?= password
DB_NAME     := $(or $(AGENTS_DB_NAME),offer_catcher_re)

# 导出给子进程（nohup bin / curl / mysql）
export AGENTS_PORT AGENTS_DB_NAME DB_HOST DB_PORT DB_USER DB_PASSWORD

.PHONY: help doctor build run start stop restart health status logs test db-init clean

help: ## 查看所有可用命令
	@echo ""
	@echo "[$(APP_NAME)] Makefile 命令"
	@echo "=============================="
	@echo ""
	@echo "  make doctor    检查 Go 环境"
	@echo "  make build     编译二进制到 $(BIN)"
	@echo "  make run       前台运行（开发调试用）"
	@echo "  make start     后台一键启动"
	@echo "  make stop      停止服务"
	@echo "  make restart   重启服务"
	@echo "  make health    健康检查 curl /healthz"
	@echo "  make status    查看进程状态"
	@echo "  make logs      查看服务日志"
	@echo "  make test      运行自动化接口测试"
	@echo "  make db-init   创建数据库 offer_catcher_re"
	@echo "  make clean     清理二进制/日志/pid"
	@echo ""

doctor: ## 检查 Go 环境
	@echo "[doctor] 检查 Go 环境..."
	@command -v go >/dev/null 2>&1 || { echo "[error] 未找到 go，请先安装 Go 1.24+"; exit 1; }
	@echo "  go: $$(go version)"
	@echo "[ok] Go 环境就绪"

build: ## 编译二进制
	@mkdir -p bin
	@go build -o $(BIN) .
	@echo "[ok] 编译完成: $(BIN)"

run: ## 前台运行（开发调试用）
	@go run .

start: ## 后台一键启动
	@mkdir -p .run logs bin
	@if [ -f $(PID_FILE) ] && kill -0 $$(cat $(PID_FILE)) 2>/dev/null; then \
		echo "[warn] $(APP_NAME) 已在运行 (PID: $$(cat $(PID_FILE)))"; \
		exit 0; \
	fi
	@echo "[info] 编译并启动 $(APP_NAME) ..."
	@go build -o $(BIN) .
	@nohup $(BIN) >> $(LOG_FILE) 2>&1 & echo $$! > $(PID_FILE)
	@echo "[info] 已启动 (PID: $$(cat $(PID_FILE)))，等待端口 $(PORT) 就绪..."
	@for i in $$(seq 1 30); do \
		if curl -sf http://localhost:$(PORT)/healthz >/dev/null 2>&1; then \
			echo "[ok] 启动成功，健康检查通过 (http://localhost:$(PORT)/healthz)"; \
			exit 0; \
		fi; \
		sleep 1; \
	done; \
	echo "[error] 启动失败（可能数据库连接失败），请查看日志: make logs"; \
	exit 1

stop: ## 停止服务
	@if [ -f $(PID_FILE) ] && kill -0 $$(cat $(PID_FILE)) 2>/dev/null; then \
		pid=$$(cat $(PID_FILE)); \
		echo "[info] 停止 $(APP_NAME) (PID: $$pid) ..."; \
		kill $$pid; \
		for i in $$(seq 1 10); do kill -0 $$pid 2>/dev/null || break; sleep 1; done; \
		if kill -0 $$pid 2>/dev/null; then kill -9 $$pid; fi; \
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

health: ## 健康检查（失败返回非零退出码）
	@echo "[health] GET http://localhost:$(PORT)/healthz"
	@resp=$$(curl -s -w "\n%{http_code}" http://localhost:$(PORT)/healthz 2>/dev/null); \
	code=$$(echo "$$resp" | tail -1); \
	body=$$(echo "$$resp" | head -n -1); \
	if [ "$$code" = "200" ]; then \
		echo "[ok] 健康检查通过 (HTTP 200)"; \
		echo "  $$body"; \
	else \
		echo "[error] 健康检查失败 (HTTP $$code)"; \
		echo "  $$body"; \
		exit 1; \
	fi

status: ## 查看进程状态
	@if [ -f $(PID_FILE) ] && kill -0 $$(cat $(PID_FILE)) 2>/dev/null; then \
		echo "[status] $(APP_NAME) 运行中 (PID: $$(cat $(PID_FILE)), 端口: $(PORT))"; \
	else \
		echo "[status] $(APP_NAME) 未运行"; \
	fi

logs: ## 查看服务日志
	@if [ -f $(LOG_FILE) ]; then tail -f $(LOG_FILE); else echo "[warn] 日志文件不存在: $(LOG_FILE)"; fi

test: ## 运行自动化接口测试
	@go test ./...

db-init: ## 创建数据库 offer_catcher_re（幂等）
	@echo "[db-init] 创建数据库 $(DB_NAME) (host=$(DB_HOST):$(DB_PORT) user=$(DB_USER))"
	@mysql -h"$(DB_HOST)" -P"$(DB_PORT)" -u"$(DB_USER)" -p"$(DB_PASSWORD)" < sql/init.sql
	@echo "[ok] 数据库就绪"

clean: ## 清理运行产物
	@rm -rf bin .run logs
	@echo "[ok] 已清理 bin/ .run/ logs/"
