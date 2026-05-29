#!/usr/bin/env bash
# ============================================================
# Offer Catcher - 基础设施管理（Docker Compose）
# ============================================================

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "${SCRIPT_DIR}/common.sh"

ACTION="${1:-status}"
cd "$PROJECT_ROOT"

case "$ACTION" in
    up)
        log_step "启动基础设施"
        if [ ! -f "docker-compose.yml" ]; then
            log_error "docker-compose.yml 不存在，无法启动"
            exit 1
        fi
        docker compose up -d
        echo ""
        log_info "Docker Compose 服务已启动"
        echo "   查看状态: make infra-status"
        echo "   查看日志: make infra-logs"
        ;;

    down)
        log_step "停止基础设施"
        docker compose down
        echo ""
        log_info "Docker Compose 服务已停止（数据卷保留）"
        ;;

    status)
        log_step "基础设施状态"
        docker compose ps 2>/dev/null || log_warn "docker compose ps 执行失败"
        ;;

    logs)
        log_step "基础设施日志"
        docker compose logs -f
        ;;

    *)
        echo "用法: $0 {up|down|status|logs}"
        exit 1
        ;;
esac
