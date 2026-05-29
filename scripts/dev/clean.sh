#!/usr/bin/env bash
# ============================================================
# Offer Catcher - 清理操作
# ============================================================

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "${SCRIPT_DIR}/common.sh"

ACTION="${1:-clean}"

case "$ACTION" in
    clean)
        log_step "清理本地运行文件和构建产物"

        # 清理 PID 文件
        if [ -d "$RUN_DIR" ]; then
            rm -f "${RUN_DIR}"/*.pid
            log_info "已清理 .run/*.pid"
        fi

        # 清理日志
        if [ -d "$LOGS_DIR" ]; then
            rm -f "${LOGS_DIR}"/*.log
            log_info "已清理 .logs/*.log"
        fi

        # 清理前端构建产物
        if [ -d "${PROJECT_ROOT}/web/dist" ]; then
            rm -rf "${PROJECT_ROOT}/web/dist"
            log_info "已清理 web/dist"
        fi

        # 清理后端构建产物
        if [ -d "${PROJECT_ROOT}/server/bin" ]; then
            rm -rf "${PROJECT_ROOT}/server/bin"
            log_info "已清理 server/bin"
        fi

        # 清理残留的 server/server 二进制
        if [ -f "${PROJECT_ROOT}/server/server" ]; then
            rm -f "${PROJECT_ROOT}/server/server"
            log_info "已清理 server/server"
        fi

        log_info "清理完成（不删除数据库 volume）"
        ;;

    logs)
        log_step "清理日志文件"
        if [ -d "$LOGS_DIR" ]; then
            rm -f "${LOGS_DIR}"/*.log
            log_info "已清理 .logs/*.log"
        else
            log_info "日志目录不存在，无需清理"
        fi
        ;;

    pids)
        log_step "清理 PID 文件"
        if [ -d "$RUN_DIR" ]; then
            rm -f "${RUN_DIR}"/*.pid
            log_info "已清理 .run/*.pid"
        else
            log_info "PID 目录不存在，无需清理"
        fi
        ;;

    volume)
        if ! confirm_dangerous "clean-volume"; then
            exit 1
        fi

        log_step "删除 Docker 数据卷"
        log_error "⚠️ 此操作将删除 MySQL 数据卷，所有数据将丢失！"
        echo ""
        read -r -p "确认删除所有 Docker volume？输入 YES 继续: " confirm_input
        if [ "$confirm_input" != "YES" ]; then
            log_info "已取消"
            exit 0
        fi

        cd "$PROJECT_ROOT"
        docker compose down -v
        log_info "Docker 数据卷已删除"
        ;;

    *)
        echo "用法: $0 {clean|logs|pids|volume}"
        exit 1
        ;;
esac
