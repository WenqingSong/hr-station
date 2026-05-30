#!/usr/bin/env bash
# ============================================================
# Offer Catcher - 日志查看
# ============================================================

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "${SCRIPT_DIR}/common.sh"

ACTION="${1:-all}"

case "$ACTION" in
    all)
        echo ""
        echo "日志查看入口："
        echo ""
        echo "  make logs-backend  查看 Go 后端日志 (.logs/server.log)"
        echo "  make logs-web      查看前端日志 (.logs/web.log)"
        echo "  make logs-mysql    查看 MySQL Docker 日志"
        echo "  make nginx-logs    查看 Nginx Docker 日志"
        echo "  make logs-agent    查看 Agent 相关日志"
        echo ""
        echo "日志文件目录: .logs/"
        echo ""

        # 显示日志文件状态
        if [ -d "$LOGS_DIR" ]; then
            echo "当前日志文件："
            ls -la "$LOGS_DIR/" 2>/dev/null || echo "  （空）"
        else
            echo "日志目录不存在，执行 make init 后自动创建"
        fi
        ;;

    backend)
        if [ ! -f "${LOGS_DIR}/server.log" ]; then
            log_warn "后端日志文件不存在: .logs/server.log"
            log_info "提示：执行 make backend-up 后生成日志"
            exit 0
        fi
        tail -f "${LOGS_DIR}/server.log"
        ;;

    web)
        if [ ! -f "${LOGS_DIR}/web.log" ]; then
            log_warn "前端日志文件不存在: .logs/web.log"
            log_info "提示：执行 make web-up 后生成日志"
            exit 0
        fi
        tail -f "${LOGS_DIR}/web.log"
        ;;

    mysql)
        cd "$PROJECT_ROOT"
        docker compose logs -f mysql
        ;;

    agent)
        # 当前查看后端日志中的 Agent 相关内容
        if [ ! -f "${LOGS_DIR}/server.log" ]; then
            log_warn "后端日志文件不存在: .logs/server.log"
            log_info "提示：阶段 4-2 实现 LLM Client 后，此处将查看 Agent/LLM 相关日志"
            exit 0
        fi
        log_info "查看 Agent 相关日志（从 .logs/server.log 中筛选）"
        grep -i -E "agent|llm|hunyuan" "${LOGS_DIR}/server.log" | tail -50
        echo ""
        log_info "完整日志请执行: make logs-backend"
        log_info "阶段 4-2 实现后，Agent 日志将独立输出"
        ;;

    nginx)
        cd "$PROJECT_ROOT"
        docker compose logs -f nginx
        ;;

    *)
        echo "用法: $0 {all|backend|web|mysql|agent|nginx}"
        exit 1
        ;;
esac
