#!/usr/bin/env bash
# ============================================================
# Offer Catcher - Go 后端管理
# ============================================================

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "${SCRIPT_DIR}/common.sh"

ACTION="${1:-status}"
PID_FILE="${RUN_DIR}/server.pid"
LOG_FILE="${LOGS_DIR}/server.log"

# 加载环境变量
load_env

SERVER_PORT=$(get_env_value "SERVER_PORT" "8080")

start_backend() {
    # 检查是否已在运行
    if is_process_running "$PID_FILE"; then
        local pid
        pid=$(cat "$PID_FILE")
        log_warn "后端已在运行 (PID: ${pid})"
        return 0
    fi

    # 检查端口是否被占用
    if check_port "$SERVER_PORT"; then
        log_error "端口 ${SERVER_PORT} 已被占用，请检查或修改 SERVER_PORT"
        return 1
    fi

    # 检查 Go 工程
    if [ ! -f "${PROJECT_ROOT}/server/go.mod" ]; then
        log_error "server/go.mod 不存在"
        return 1
    fi

    init_runtime_dirs

    log_info "启动 Go 后端 (端口: ${SERVER_PORT})..."

    # 使用 nohup 后台启动，日志重定向
    cd "${PROJECT_ROOT}/server"
    SERVER_PORT="$SERVER_PORT" nohup go run ./cmd/server >> "$LOG_FILE" 2>&1 &
    local pid=$!
    echo "$pid" > "$PID_FILE"

    # 等待启动（首次 go run 可能需要下载依赖，给予 30 秒）
    local wait_count=0
    while ! check_port "$SERVER_PORT" && [ $wait_count -lt 30 ]; do
        sleep 1
        wait_count=$((wait_count + 1))
    done

    if check_port "$SERVER_PORT"; then
        log_info "后端已启动 (PID: ${pid}, 端口: ${SERVER_PORT})"
    else
        log_error "后端启动超时，请查看日志: make logs-backend"
        # 检查进程是否还在
        if ! kill -0 "$pid" 2>/dev/null; then
            log_error "进程已退出，可能是数据库连接失败"
        fi
        return 1
    fi
}

stop_backend() {
    stop_pid_file "$PID_FILE" "后端"

    # 兜底：检查端口是否仍有 go/server 相关残留进程
    if check_port "$SERVER_PORT"; then
        local port_pid=""
        if command -v lsof >/dev/null 2>&1; then
            port_pid=$(lsof -t -i ":${SERVER_PORT}" -sTCP:LISTEN 2>/dev/null || true)
        fi
        if [ -n "$port_pid" ]; then
            local cmd
            cmd=$(ps -p "$port_pid" -o args= 2>/dev/null || true)
            if echo "$cmd" | grep -qi "go\|server"; then
                log_warn "端口 ${SERVER_PORT} 仍有残留进程 (PID: ${port_pid})，正在清理..."
                kill "$port_pid" 2>/dev/null || true
                sleep 1
                if kill -0 "$port_pid" 2>/dev/null; then
                    kill -9 "$port_pid" 2>/dev/null || true
                fi
            fi
        fi
    fi
}

do_status() {
    if is_process_running "$PID_FILE"; then
        local pid
        pid=$(cat "$PID_FILE")
        log_info "后端运行中 (PID: ${pid}, 端口: ${SERVER_PORT})"
    else
        log_info "后端未运行"
    fi
}

do_health() {
    local health_url="http://localhost:${SERVER_PORT}/healthz"
    log_step "后端健康检查: ${health_url}"
    if command -v curl >/dev/null 2>&1; then
        local response
        response=$(curl -s -w "\n%{http_code}" "$health_url" 2>/dev/null || echo "000")
        local http_code
        http_code=$(echo "$response" | tail -1)
        local body
        body=$(echo "$response" | head -n -1)

        if [ "$http_code" = "200" ]; then
            log_info "后端健康检查通过 (HTTP 200)"
            echo "  响应: $body"
        else
            log_error "后端健康检查失败 (HTTP ${http_code})"
            if [ -n "$body" ]; then
                echo "  响应: $body"
            fi
        fi
    else
        log_warn "curl 不可用，无法检查"
    fi
}

case "$ACTION" in
    up)
        start_backend
        ;;

    stop)
        stop_backend
        ;;

    restart)
        stop_backend
        sleep 1
        start_backend
        ;;

    logs)
        if [ ! -f "$LOG_FILE" ]; then
            log_warn "日志文件不存在: ${LOG_FILE}"
            exit 0
        fi
        tail -f "$LOG_FILE"
        ;;

    status)
        do_status
        ;;

    health)
        do_health
        ;;

    *)
        echo "用法: $0 {up|stop|restart|logs|status|health}"
        exit 1
        ;;
esac
