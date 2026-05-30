#!/usr/bin/env bash
# ============================================================
# Offer Catcher - 前端开发服务管理
# ============================================================

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "${SCRIPT_DIR}/common.sh"

ACTION="${1:-status}"
PID_FILE="${RUN_DIR}/web.pid"
LOG_FILE="${LOGS_DIR}/web.log"

# 加载环境变量
load_env

WEB_PORT=$(get_env_value "WEB_PORT" "5173")
WEB_DIR="${PROJECT_ROOT}/web"

start_web() {
    # 检查是否已在运行
    if is_process_running "$PID_FILE"; then
        local pid
        pid=$(cat "$PID_FILE")
        log_warn "前端已在运行 (PID: ${pid})"
        return 0
    fi

    # 检查端口是否被占用
    if check_port "$WEB_PORT"; then
        log_error "端口 ${WEB_PORT} 已被占用，请检查或修改 WEB_PORT"
        return 1
    fi

    # 检查前端工程
    if [ ! -f "${WEB_DIR}/package.json" ]; then
        log_error "web/package.json 不存在"
        return 1
    fi

    if [ ! -d "${WEB_DIR}/node_modules" ]; then
        log_error "web/node_modules 不存在，请先执行: make web-install"
        return 1
    fi

    init_runtime_dirs

    log_info "启动前端开发服务 (端口: ${WEB_PORT})..."

    # 使用 nohup 后台启动，日志重定向
    # --host 0.0.0.0 使 Vite 监听所有网卡，Nginx 容器通过 host.docker.internal 可访问
    cd "$WEB_DIR"
    nohup npm run dev -- --host 0.0.0.0 >> "$LOG_FILE" 2>&1 &
    local pid=$!
    echo "$pid" > "$PID_FILE"

    # 等待启动
    local wait_count=0
    while ! check_port "$WEB_PORT" && [ $wait_count -lt 15 ]; do
        sleep 1
        wait_count=$((wait_count + 1))
    done

    if check_port "$WEB_PORT"; then
        log_info "前端已启动 (PID: ${pid}, 端口: ${WEB_PORT})"
    else
        log_warn "前端启动可能仍在进行中，请查看日志: make logs-web"
    fi
}

stop_web() {
    stop_pid_file "$PID_FILE" "前端"
}

case "$ACTION" in
    install)
        log_step "安装前端依赖"
        if [ ! -f "${WEB_DIR}/package.json" ]; then
            log_error "web/package.json 不存在"
            exit 1
        fi
        cd "$WEB_DIR"
        npm install
        log_info "前端依赖安装完成"
        ;;

    up)
        start_web
        ;;

    stop)
        stop_web
        ;;

    restart)
        stop_web
        sleep 1
        start_web
        ;;

    logs)
        if [ ! -f "$LOG_FILE" ]; then
            log_warn "日志文件不存在: ${LOG_FILE}"
            exit 0
        fi
        tail -f "$LOG_FILE"
        ;;

    status)
        if is_process_running "$PID_FILE"; then
            pid=$(cat "$PID_FILE")
            log_info "前端运行中 (PID: ${pid}, 端口: ${WEB_PORT})"
        else
            log_info "前端未运行"
        fi
        ;;

    health)
        log_step "前端健康检查"
        if check_port "$WEB_PORT"; then
            log_info "前端端口 ${WEB_PORT} 可访问"
        else
            log_error "前端端口 ${WEB_PORT} 不可访问"
        fi
        ;;

    *)
        echo "用法: $0 {install|up|stop|restart|logs|status|health}"
        exit 1
        ;;
esac
