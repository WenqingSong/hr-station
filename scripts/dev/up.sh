#!/usr/bin/env bash
# ============================================================
# Offer Catcher - 一键启动完整开发环境
# 严格错误处理：任何关键步骤失败都会导致整体失败
# ============================================================

set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "${SCRIPT_DIR}/common.sh"

# ---- 状态追踪 ----
FAILURES=()
CURRENT_STEP=0
TOTAL_STEPS=10

step_header() {
    CURRENT_STEP=$((CURRENT_STEP + 1))
    echo ""
    echo "=========================================="
    echo "  [${CURRENT_STEP}/${TOTAL_STEPS}] $*"
    echo "=========================================="
}

add_failure() {
    FAILURES+=("$*")
}

# ---- Step 1: 确保 .env 存在 ----
step_header "确保 .env 存在"
if ! bash "${SCRIPT_DIR}/ensure_env.sh"; then
    add_failure "环境变量文件(.env)创建失败"
fi

# ---- Step 2: 开发环境预检（非致命，仅警告） ----
step_header "开发环境预检"
bash "${SCRIPT_DIR}/doctor.sh" || true

# ---- Step 3: 初始化运行时目录 ----
step_header "初始化运行时目录"
init_runtime_dirs
log_info ".run/ 和 .logs/ 目录已就绪"

# ---- Step 4: 启动 Docker 基础设施 ----
step_header "启动基础设施 (MySQL + Nginx)"
if ! bash "${SCRIPT_DIR}/infra.sh" up; then
    add_failure "Docker 基础设施启动失败"
fi

# ---- Step 5: 等待 MySQL 就绪 ----
step_header "等待 MySQL 就绪"
load_env
MYSQL_ROOT_PASSWORD=$(get_env_value "MYSQL_ROOT_PASSWORD" "password")
DB_NAME=$(get_env_value "DB_NAME" "offer_catcher")

mysql_ready=false
max_wait=60
wait_count=0
while [ $wait_count -lt $max_wait ]; do
    if docker exec offer-catcher-mysql mysqladmin ping -h localhost -uroot -p"${MYSQL_ROOT_PASSWORD}" --silent 2>/dev/null; then
        mysql_ready=true
        break
    fi
    sleep 1
    wait_count=$((wait_count + 1))
    if [ $((wait_count % 10)) -eq 0 ]; then
        log_info "等待 MySQL... (${wait_count}s/${max_wait}s)"
    fi
done

if [ "$mysql_ready" = true ]; then
    log_info "MySQL 已就绪 (等待 ${wait_count}s)"
else
    add_failure "MySQL 启动超时 (${max_wait}s)"
fi

# ---- Step 6: 检查/导入数据库表 ----
step_header "检查数据库表"
tables_ok=false
if [ "$mysql_ready" = true ]; then
    core_tables="users resume_versions jobs recommendation_results match_analyses version_compare_results"
    missing_tables=()

    for table in $core_tables; do
        count=$(docker exec offer-catcher-mysql mysql -uroot -p"${MYSQL_ROOT_PASSWORD}" "$DB_NAME" \
            -s -N -e "SELECT COUNT(*) FROM ${table};" 2>/dev/null || echo "ERROR")
        if [ "$count" = "ERROR" ] || [ -z "$count" ]; then
            missing_tables+=("$table")
        fi
    done

    if [ ${#missing_tables[@]} -eq 0 ]; then
        log_info "所有核心表已存在，跳过导入"
        tables_ok=true
    else
        log_warn "缺失表: ${missing_tables[*]}"
        log_info "自动导入 server/sql/init.sql..."
        if bash "${SCRIPT_DIR}/mysql.sh" import; then
            # 导入后再次验证
            still_missing=()
            for table in $core_tables; do
                count=$(docker exec offer-catcher-mysql mysql -uroot -p"${MYSQL_ROOT_PASSWORD}" "$DB_NAME" \
                    -s -N -e "SELECT COUNT(*) FROM ${table};" 2>/dev/null || echo "ERROR")
                if [ "$count" = "ERROR" ] || [ -z "$count" ]; then
                    still_missing+=("$table")
                fi
            done
            if [ ${#still_missing[@]} -eq 0 ]; then
                log_info "数据库表导入成功"
                tables_ok=true
            else
                add_failure "数据库表导入后仍然缺失: ${still_missing[*]}"
            fi
        else
            add_failure "数据库初始化 SQL 导入失败"
        fi
    fi
else
    add_failure "MySQL 不可用，跳过数据库检查"
fi

# ---- Step 7: 启动后端 ----
step_header "启动后端"
if ! bash "${SCRIPT_DIR}/backend.sh" up; then
    add_failure "后端启动失败"
fi

# ---- Step 8: 后端健康检查 ----
step_header "后端健康检查"
sleep 3
SERVER_PORT=$(get_env_value "SERVER_PORT" "8080")
backend_ok=false
if command -v curl >/dev/null 2>&1; then
    # 最多重试 5 次，间隔 3 秒
    retry=0
    while [ $retry -lt 5 ]; do
        http_code=$(curl -s -o /dev/null -w "%{http_code}" "http://localhost:${SERVER_PORT}/healthz" 2>/dev/null || echo "000")
        if [ "$http_code" = "200" ]; then
            break
        fi
        retry=$((retry + 1))
        if [ $retry -lt 5 ]; then
            log_info "等待后端就绪... (尝试 $((retry + 1))/5, HTTP ${http_code})"
            sleep 3
        fi
    done

    if [ "$http_code" = "200" ]; then
        log_info "后端健康检查通过 (HTTP 200)"
        backend_ok=true
    else
        add_failure "后端健康检查失败 (HTTP ${http_code})"
    fi
else
    if check_port "$SERVER_PORT"; then
        log_info "后端端口 ${SERVER_PORT} 可访问 (curl 不可用，仅检查端口)"
        backend_ok=true
    else
        add_failure "后端端口 ${SERVER_PORT} 不可访问"
    fi
fi

# ---- Step 9: 前端依赖安装 & 启动 ----
step_header "前端依赖检查与启动"
WEB_DIR="${PROJECT_ROOT}/web"

if [ -f "${WEB_DIR}/package.json" ] && [ ! -d "${WEB_DIR}/node_modules" ]; then
    log_info "web/node_modules 不存在，自动安装前端依赖..."
    if [ -f "${WEB_DIR}/package-lock.json" ]; then
        (cd "$WEB_DIR" && npm ci 2>&1) || (cd "$WEB_DIR" && npm install 2>&1) || {
            add_failure "前端依赖安装失败"
        }
    else
        (cd "$WEB_DIR" && npm install 2>&1) || {
            add_failure "前端依赖安装失败"
        }
    fi
    log_info "前端依赖安装完成"
fi

if ! bash "${SCRIPT_DIR}/web.sh" up; then
    add_failure "前端启动失败"
fi

# ---- Step 10: 最终健康检查 ----
step_header "最终健康检查"
WEB_PORT=$(get_env_value "WEB_PORT" "5173")
NGINX_PORT=$(get_env_value "NGINX_PORT" "80")
frontend_ok=false
nginx_ok=false

# 等待前端启动
sleep 3

# 检查前端端口（最多重试 3 次）
frontend_retry=0
while [ $frontend_retry -lt 3 ]; do
    if check_port "$WEB_PORT"; then
        break
    fi
    frontend_retry=$((frontend_retry + 1))
    if [ $frontend_retry -lt 3 ]; then
        log_info "等待前端就绪... (尝试 $((frontend_retry + 1))/3)"
        sleep 3
    fi
done

if check_port "$WEB_PORT"; then
    log_info "前端端口 ${WEB_PORT} 可访问"
    frontend_ok=true
else
    add_failure "前端端口 ${WEB_PORT} 不可访问"
fi

# 检查 Nginx 反向代理
if command -v curl >/dev/null 2>&1; then
    # 检查 /healthz
    nginx_retry=0
    nginx_health_code="000"
    while [ $nginx_retry -lt 3 ]; do
        nginx_health_code=$(curl -s -o /dev/null -w "%{http_code}" "http://localhost:${NGINX_PORT}/healthz" 2>/dev/null || echo "000")
        if [ "$nginx_health_code" = "200" ]; then
            break
        fi
        nginx_retry=$((nginx_retry + 1))
        if [ $nginx_retry -lt 3 ]; then
            sleep 2
        fi
    done

    if [ "$nginx_health_code" = "200" ]; then
        log_info "Nginx 反向代理 /healthz 可访问 (HTTP 200)"
        nginx_ok=true
    else
        add_failure "Nginx 反向代理 /healthz 不可访问 (HTTP ${nginx_health_code})"
    fi

    # 检查 / (非致命)
    nginx_root_code=$(curl -s -o /dev/null -w "%{http_code}" "http://localhost:${NGINX_PORT}/" 2>/dev/null || echo "000")
    if [ "$nginx_root_code" = "200" ] || [ "$nginx_root_code" = "304" ]; then
        log_info "Nginx 反向代理 / 可访问 (HTTP ${nginx_root_code})"
    else
        log_warn "Nginx 反向代理 / 返回 HTTP ${nginx_root_code}（前端可能仍在启动）"
    fi
else
    if check_port "$NGINX_PORT"; then
        log_info "Nginx 端口 ${NGINX_PORT} 可访问 (curl 不可用，仅检查端口)"
        nginx_ok=true
    else
        add_failure "Nginx 端口 ${NGINX_PORT} 不可访问"
    fi
fi

# ---- 最终汇总 ----
echo ""
echo "=========================================="
echo "  启动结果汇总"
echo "=========================================="

if [ ${#FAILURES[@]} -eq 0 ]; then
    echo -e "  ${GREEN}✅ 所有检查通过${NC}"
    echo ""
    echo "=========================================="
    echo -e "  ${GREEN}✅ Offer Catcher 开发环境已启动${NC}"
    echo "=========================================="
    echo ""
    echo "Access:"
    echo "  App:       http://localhost/"
    echo "  API:       http://localhost/api/"
    echo "  Health:    http://localhost/healthz"
    echo "  Backend:   http://localhost:${SERVER_PORT}"
    echo "  Frontend:  http://localhost:${WEB_PORT}"
    echo "  MySQL:     localhost:3306"
    echo ""
    echo "Useful commands:"
    echo "  make status       查看服务状态"
    echo "  make logs         查看日志"
    echo "  make restart-soft 软重启（不清空数据）"
    echo "  make down         停止所有服务"
    echo ""
    exit 0
else
    echo -e "  ${RED}❌ 以下步骤失败:${NC}"
    for failure in "${FAILURES[@]}"; do
        echo -e "    ${RED}•${NC} ${failure}"
    done
    echo ""
    echo "=========================================="
    echo -e "  ${RED}❌ Offer Catcher 开发环境启动失败${NC}"
    echo "=========================================="
    echo ""
    echo "Troubleshooting:"
    echo "  make doctor        检查开发环境"
    echo "  make logs          查看日志"
    echo "  make status        查看服务状态"
    echo "  make db-import     手动导入数据库"
    echo "  make web-install   手动安装前端依赖"
    echo ""
    exit 1
fi
