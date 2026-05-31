#!/usr/bin/env bash
# ============================================================
# Offer Catcher - MySQL 数据库管理
# ============================================================

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "${SCRIPT_DIR}/common.sh"

ACTION="${1:-check}"

# 加载环境变量
load_env

# 数据库连接参数
DB_HOST=$(get_env_value "DB_HOST" "127.0.0.1")
DB_PORT=$(get_env_value "DB_PORT" "3306")
DB_NAME=$(get_env_value "DB_NAME" "offer_catcher")
DB_USER=$(get_env_value "DB_USER" "root")
DB_PASSWORD=$(get_env_value "DB_PASSWORD" "password")
MYSQL_ROOT_PASSWORD=$(get_env_value "MYSQL_ROOT_PASSWORD" "password")

# SQL 文件路径
INIT_SQL="${PROJECT_ROOT}/server/sql/init.sql"

# 核心表列表
CORE_TABLES="users resume_versions jobs recommendation_results match_analyses version_compare_results"

# 检查 MySQL 容器是否运行
check_mysql_container() {
    if ! docker compose -f "${PROJECT_ROOT}/docker-compose.yml" ps mysql 2>/dev/null | grep -q "Up\|healthy"; then
        log_error "MySQL 容器未运行，请先执行 make infra-up"
        return 1
    fi
    return 0
}

# 检查核心表是否都存在（返回 0=存在, 1=缺失）
tables_exist() {
    for table in $CORE_TABLES; do
        local count
        count=$(docker exec offer-catcher-mysql mysql -uroot -p"${MYSQL_ROOT_PASSWORD}" "$DB_NAME" \
            -s -N -e "SELECT COUNT(*) FROM ${table};" 2>/dev/null || echo "ERROR")
        if [ "$count" = "ERROR" ] || [ -z "$count" ]; then
            return 1
        fi
    done
    return 0
}

do_check() {
    log_step "数据库检查"
    if ! check_mysql_container; then
        return 1
    fi

    # 检查核心表是否存在
    local all_ok=true

    for table in $CORE_TABLES; do
        local count
        count=$(docker exec offer-catcher-mysql mysql -uroot -p"${MYSQL_ROOT_PASSWORD}" "$DB_NAME" \
            -s -N -e "SELECT COUNT(*) FROM ${table};" 2>/dev/null || echo "ERROR")

        if [ "$count" = "ERROR" ] || [ -z "$count" ]; then
            echo -e "  ${RED}❌${NC} ${table}: 不存在或查询失败"
            all_ok=false
        else
            echo -e "  ${GREEN}✅${NC} ${table}: ${count} 条记录"
        fi
    done

    echo ""
    if [ "$all_ok" = true ]; then
        log_info "所有核心表检查通过"
        return 0
    else
        log_warn "部分表缺失，请执行 make db-import 导入初始化 SQL"
        return 1
    fi
}

case "$ACTION" in
    import)
        log_step "导入数据库初始化 SQL"
        if [ ! -f "$INIT_SQL" ]; then
            log_error "server/sql/init.sql 不存在"
            exit 1
        fi

        # 检查 MySQL 是否在运行
        if ! docker compose -f "${PROJECT_ROOT}/docker-compose.yml" ps mysql 2>/dev/null | grep -q "Up\|healthy"; then
            log_error "MySQL 容器未运行，请先执行 make infra-up"
            exit 1
        fi

        # 等待 MySQL 就绪
        log_info "等待 MySQL 就绪..."
        local_count=0
        while ! docker exec offer-catcher-mysql mysqladmin ping -h localhost -uroot -p"${MYSQL_ROOT_PASSWORD}" --silent 2>/dev/null; do
            if [ $local_count -ge 30 ]; then
                log_error "MySQL 启动超时"
                exit 1
            fi
            sleep 1
            local_count=$((local_count + 1))
        done

        log_info "开始导入 server/sql/init.sql..."
        docker exec -i offer-catcher-mysql mysql -uroot -p"${MYSQL_ROOT_PASSWORD}" "$DB_NAME" < "$INIT_SQL"
        log_info "SQL 导入完成"
        ;;

    reset)
        if ! confirm_dangerous "db-reset"; then
            exit 1
        fi

        log_step "重置数据库（重新导入初始化 SQL）"
        if [ ! -f "$INIT_SQL" ]; then
            log_error "server/sql/init.sql 不存在"
            exit 1
        fi

        # 检查 MySQL 是否在运行
        if ! docker compose -f "${PROJECT_ROOT}/docker-compose.yml" ps mysql 2>/dev/null | grep -q "Up\|healthy"; then
            log_error "MySQL 容器未运行，请先执行 make infra-up"
            exit 1
        fi

        log_warn "将重新导入 init.sql，已有数据将被 DROP 后重建"
        docker exec -i offer-catcher-mysql mysql -uroot -p"${MYSQL_ROOT_PASSWORD}" "$DB_NAME" < "$INIT_SQL"
        log_info "数据库重置完成"
        ;;

    shell)
        log_step "进入 MySQL Shell"
        if ! docker compose -f "${PROJECT_ROOT}/docker-compose.yml" ps mysql 2>/dev/null | grep -q "Up\|healthy"; then
            log_error "MySQL 容器未运行，请先执行 make infra-up"
            exit 1
        fi
        docker exec -it offer-catcher-mysql mysql -uroot -p"${MYSQL_ROOT_PASSWORD}" "$DB_NAME"
        ;;

    check)
        do_check
        ;;

    tables-exist)
        # 静默检查核心表是否存在，返回 0=存在, 1=缺失
        if ! check_mysql_container; then
            exit 1
        fi
        if tables_exist; then
            exit 0
        else
            exit 1
        fi
        ;;

    *)
        echo "用法: $0 {import|reset|shell|check|tables-exist}"
        exit 1
        ;;
esac
