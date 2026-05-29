#!/usr/bin/env bash
# ============================================================
# Offer Catcher - 环境初始化与配置展示
# ============================================================

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "${SCRIPT_DIR}/common.sh"

ACTION="${1:-init}"

case "$ACTION" in
    init)
        echo ""
        echo "=========================================="
        echo "  初始化开发环境"
        echo "=========================================="
        echo ""

        # 1. 检查 .env
        log_step "1. 检查 .env"
        if [ -f "${PROJECT_ROOT}/.env" ]; then
            log_info ".env 已存在，跳过（不会覆盖）"
        elif [ -f "${PROJECT_ROOT}/.env.example" ]; then
            cp "${PROJECT_ROOT}/.env.example" "${PROJECT_ROOT}/.env"
            log_info "已从 .env.example 复制创建 .env"
        else
            log_warn ".env.example 不存在，无法创建 .env"
        fi
        echo ""

        # 2. 创建运行时目录
        log_step "2. 创建运行时目录"
        init_runtime_dirs
        log_info ".run/ 和 .logs/ 目录已就绪"
        echo ""

        # 3. 检查前端依赖
        log_step "3. 检查前端依赖"
        if [ -f "${PROJECT_ROOT}/web/package.json" ]; then
            if [ -d "${PROJECT_ROOT}/web/node_modules" ]; then
                log_info "web/node_modules 已存在，跳过安装"
            else
                log_warn "web/node_modules 不存在，请执行: make web-install"
            fi
        else
            log_warn "web/package.json 不存在，跳过前端依赖检查"
        fi
        echo ""

        # 4. 检查后端依赖
        log_step "4. 检查后端依赖"
        if [ -f "${PROJECT_ROOT}/server/go.mod" ]; then
            log_info "server/go.mod 已存在"
            echo "   如需下载依赖，请执行: cd server && go mod download"
        else
            log_warn "server/go.mod 不存在"
        fi
        echo ""

        echo "=========================================="
        echo "  初始化完成"
        echo "  下一步: make up    (启动 Docker 服务)"
        echo "          make dev   (查看开发启动方式)"
        echo "=========================================="
        echo ""
        ;;

    show)
        echo ""
        echo "=========================================="
        echo "  当前配置概览"
        echo "=========================================="
        echo ""

        # 应用配置
        log_step "应用配置"
        echo "  APP_NAME:          $(get_env_value "APP_NAME" "未配置")"
        echo "  APP_ENV:           $(get_env_value "APP_ENV" "未配置")"
        echo "  SERVER_PORT:       $(get_env_value "SERVER_PORT" "8080")"
        echo "  WEB_PORT:          $(get_env_value "WEB_PORT" "5173")"
        echo "  VITE_API_BASE_URL: $(get_env_value "VITE_API_BASE_URL" "未配置")"
        echo ""

        # 数据库配置
        log_step "数据库配置"
        echo "  DB_HOST:     $(get_env_value "DB_HOST" "未配置")"
        echo "  DB_PORT:     $(get_env_value "DB_PORT" "3306")"
        echo "  DB_NAME:     $(get_env_value "DB_NAME" "未配置")"
        echo "  DB_USER:     $(get_env_value "DB_USER" "未配置")"
        echo "  DB_PASSWORD: 已配置/未配置（不显示值）"
        echo ""

        # JWT 配置
        log_step "JWT 配置"
        echo "  JWT_SECRET:      已配置/未配置（不显示值）"
        echo "  JWT_EXPIRE_HOURS: $(get_env_value "JWT_EXPIRE_HOURS" "168")"
        echo ""

        # LLM / Agent 配置
        log_step "LLM / Agent 配置"
        echo "  ENABLE_LLM_AGENT:    $(get_env_value "ENABLE_LLM_AGENT" "未配置")"
        echo "  ENABLE_LLM_FALLBACK:  $(get_env_value "ENABLE_LLM_FALLBACK" "未配置")"
        echo "  LLM_PROVIDER:        $(get_env_value "LLM_PROVIDER" "未配置")"
        echo "  LLM_BASE_URL:        $(get_env_value "LLM_BASE_URL" "未配置")"
        echo "  LLM_MODEL:           $(get_env_value "LLM_MODEL" "未配置")"
        echo "  LLM_API_KEY:         $(mask_secret "$(get_env_value "LLM_API_KEY" "")")"
        echo "  ENABLE_MOCK_AGENT:   $(get_env_value "ENABLE_MOCK_AGENT" "未配置")"
        echo ""

        # 运行状态
        log_step "运行状态"
        if is_process_running "${RUN_DIR}/server.pid"; then
            echo "  后端: 运行中 (PID: $(cat "${RUN_DIR}/server.pid"))"
        else
            echo "  后端: 未运行"
        fi

        if is_process_running "${RUN_DIR}/web.pid"; then
            echo "  前端: 运行中 (PID: $(cat "${RUN_DIR}/web.pid"))"
        else
            echo "  前端: 未运行"
        fi
        echo ""

        echo "=========================================="
        echo ""
        ;;

    *)
        echo "用法: $0 {init|show}"
        exit 1
        ;;
esac
