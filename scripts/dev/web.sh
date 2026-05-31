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

    # 自动安装依赖（如果 node_modules 不存在）
    if [ ! -d "${WEB_DIR}/node_modules" ]; then
        log_info "web/node_modules 不存在，自动安装前端依赖..."
        if [ -f "${WEB_DIR}/package-lock.json" ]; then
            (cd "$WEB_DIR" && npm ci 2>&1) || (cd "$WEB_DIR" && npm install 2>&1) || {
                log_error "前端依赖安装失败"
                return 1
            }
        else
            (cd "$WEB_DIR" && npm install 2>&1) || {
                log_error "前端依赖安装失败"
                return 1
            }
        fi
        log_info "前端依赖安装完成"
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
    while ! check_port "$WEB_PORT" && [ $wait_count -lt 20 ]; do
        sleep 1
        wait_count=$((wait_count + 1))
    done

    if check_port "$WEB_PORT"; then
        log_info "前端已启动 (PID: ${pid}, 端口: ${WEB_PORT})"
    else
        log_error "前端启动超时，请查看日志: make logs-web"
        return 1
    fi
}

stop_web() {
    stop_pid_file "$PID_FILE" "前端"

    # 端口兜底：如果 5173 仍被占用，尝试清理本项目相关进程
    if check_port "$WEB_PORT"; then
        local port_pids=""
        if command -v lsof >/dev/null 2>&1; then
            port_pids=$(lsof -t -i ":${WEB_PORT}" -sTCP:LISTEN 2>/dev/null || true)
        fi

        if [ -n "$port_pids" ]; then
            for port_pid in $port_pids; do
                local cmd
                cmd=$(ps -p "$port_pid" -o args= 2>/dev/null || true)
                local cwd=""
                cwd=$(readlink /proc/"$port_pid"/cwd 2>/dev/null || true)

                # 判断是否属于当前项目前端进程
                local is_project_process=0
                if echo "$cmd" | grep -qi "vite"; then
                    is_project_process=1
                elif echo "$cmd" | grep -qi "npm.*dev"; then
                    is_project_process=1
                elif [ -n "$cwd" ] && echo "$cwd" | grep -q "$(basename "$WEB_DIR")"; then
                    is_project_process=1
                fi

                if [ "$is_project_process" -eq 1 ]; then
                    log_warn "端口 ${WEB_PORT} 仍有本项目前端残留进程 (PID: ${port_pid})，正在清理..."
                    # 同时清理该进程的后代
                    local descendants
                    descendants=$(get_all_descendants "$port_pid")
                    for desc_pid in $descendants; do
                        kill "$desc_pid" 2>/dev/null || true
                    done
                    kill "$port_pid" 2>/dev/null || true
                    sleep 1
                    if kill -0 "$port_pid" 2>/dev/null; then
                        kill -9 "$port_pid" 2>/dev/null || true
                    fi
                    for desc_pid in $descendants; do
                        if kill -0 "$desc_pid" 2>/dev/null; then
                            kill -9 "$desc_pid" 2>/dev/null || true
                        fi
                    done
                else
                    log_warn "端口 ${WEB_PORT} 被非本项目进程占用 (PID: ${port_pid})，请手动处理"
                    log_warn "  进程命令: ${cmd}"
                fi
            done
        fi
    fi
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
            exit 1
        fi
        ;;

    *)
        echo "用法: $0 {install|up|stop|restart|logs|status|health}"
        exit 1
        ;;
esac
