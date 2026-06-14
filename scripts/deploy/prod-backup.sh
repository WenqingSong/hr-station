#!/usr/bin/env bash
# ============================================================
# Offer捕手 - 生产数据库备份脚本
#
# 职责：生产部署前对 MySQL 做逻辑备份（mysqldump → gzip）
# 用法：
#   bash scripts/deploy/prod-backup.sh
#   make prod-backup   （推荐通过 Makefile 调用）
#
# 本脚本规范：
#   - 只读数据库，不执行任何写操作
#   - 不 source .env.prod
#   - 不打印数据库密码
#   - 备份文件输出到 backups/mysql/
#   - 文件名格式: offer_catcher_YYYYmmdd_HHMMSS.sql.gz
# ============================================================

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

ENV_FILE=".env.prod"
COMPOSE_FILE="docker-compose.prod.yml"
BACKUP_DIR="backups/mysql"
PROJECT_NAME_DEFAULT="wenqing-hr-demo"

echo "[prod-backup] starting production database backup..."
echo ""

# ── 前置检查 ──
if [[ ! -f "$ENV_FILE" ]]; then
    echo "[prod-backup] ERROR: .env.prod not found."
    echo "[prod-backup] Create it from .env.prod.example and fill real secrets on the production server."
    exit 1
fi

if [[ ! -f "$COMPOSE_FILE" ]]; then
    echo "[prod-backup] ERROR: docker-compose.prod.yml not found."
    exit 1
fi

# 提取 project name（不 source .env.prod）
project_name=""
if grep -qE '^COMPOSE_PROJECT_NAME=.' "$ENV_FILE" 2>/dev/null; then
    project_name=$(grep '^COMPOSE_PROJECT_NAME=' "$ENV_FILE" | tail -1 | cut -d= -f2- | sed -E 's/^["\x27]//; s/["\x27]$//')
fi
if [[ -z "$project_name" ]]; then
    project_name="$PROJECT_NAME_DEFAULT"
fi

# 提取 DB_NAME（不打印值）
db_name=$(grep -E '^DB_NAME=' "$ENV_FILE" 2>/dev/null | head -1 | cut -d= -f2- | sed -E 's/^["\x27]//; s/["\x27]$//' || echo "offer_catcher")
if [[ -z "$db_name" ]]; then
    db_name="offer_catcher"
fi

# 创建备份目录
mkdir -p "$BACKUP_DIR"

timestamp="$(date +%Y%m%d_%H%M%S)"
backup_file="${BACKUP_DIR}/offer_catcher_${timestamp}.sql.gz"

echo "[prod-backup] compose project : ${project_name}"
echo "[prod-backup] database         : ${db_name}"
echo "[prod-backup] backup file      : ${backup_file}"
echo ""

# ── 检查 MySQL 容器状态 ──
container_name="${project_name}-mysql-1"
if ! docker ps --format '{{.Names}}' 2>/dev/null | grep -q "^${container_name}$"; then
    # 尝试常见命名变体
    alt_name="offer-catcher-mysql-prod"
    if docker ps --format '{{.Names}}' 2>/dev/null | grep -q "^${alt_name}$"; then
        container_name="$alt_name"
    else
        echo "[prod-backup] ERROR: MySQL container not running."
        echo "[prod-backup] Check: docker compose --project-name ${project_name} -f ${COMPOSE_FILE} ps"
        echo "[prod-backup] Note: Container name may be ${container_name} or offer-catcher-mysql-prod"
        exit 1
    fi
fi

echo "[prod-backup] MySQL container : ${container_name} (running)"
echo ""

# ── 执行备份（通过 docker exec 不暴露密码）──
# Docker compose exec 从容器环境中获取 MYSQL_ROOT_PASSWORD
echo "[prod-backup] running mysqldump via docker exec ..."

if ! docker exec "$container_name" \
    sh -c 'mysqldump -uroot -p"$MYSQL_ROOT_PASSWORD" --single-transaction --routines --triggers "$MYSQL_DATABASE"' \
    | gzip > "$backup_file"; then
    echo ""
    echo "[prod-backup] ERROR: mysqldump failed."
    echo "[prod-backup] Check that the MySQL container is healthy:"
    echo "  docker compose --project-name ${project_name} -f ${COMPOSE_FILE} ps"
    exit 1
fi

# ── 验证备份 ──
if [[ ! -s "$backup_file" ]]; then
    echo "[prod-backup] ERROR: backup file is empty: ${backup_file}"
    exit 1
fi

echo ""
echo "[prod-backup] ✅ backup complete"
echo "[prod-backup] file : ${backup_file}"
echo "[prod-backup] size : $(ls -lh "$backup_file" | awk '{print $5}')"
echo "[prod-backup] OK"
