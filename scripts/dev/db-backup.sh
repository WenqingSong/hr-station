#!/usr/bin/env bash
# ============================================================
# Offer Catcher - 安全数据库备份脚本
# 使用 MYSQL_PWD 环境变量避免 mysqldump warning 污染 SQL 文件
# ============================================================
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "${SCRIPT_DIR}/common.sh"

load_env

MYSQL_ROOT_PASSWORD=$(get_env_value "MYSQL_ROOT_PASSWORD" "password")
DB_NAME=$(get_env_value "DB_NAME" "offer_catcher")
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

BACKUP_DIR="${PROJECT_ROOT}/backups/db"
mkdir -p "${BACKUP_DIR}"

TS=$(date +"%Y%m%d_%H%M%S")

FULL_BACKUP="${BACKUP_DIR}/offer_catcher_full_${TS}.sql"
SEP_BACKUP="${BACKUP_DIR}/offer_catcher_sep_recommend_${TS}.sql"
SCHEMA_BACKUP="${BACKUP_DIR}/offer_catcher_schema_${TS}.sql"
ARCHIVE="${BACKUP_DIR}/offer_catcher_db_backup_${TS}.tar.gz"

# 使用 MYSQL_PWD 环境变量替代 -p 参数，避免 mysqldump warning 写入 SQL 文件
# 注意：MYSQL_PWD 不在命令行中暴露密码，mysqldump stderr 不会输出 password warning
log_info "[db-backup] dumping full database: ${FULL_BACKUP}"
docker exec -i offer-catcher-mysql sh -c "MYSQL_PWD='${MYSQL_ROOT_PASSWORD}' mysqldump -uroot --databases ${DB_NAME} --single-transaction --routines --triggers --events 2>/dev/null" > "${FULL_BACKUP}"

log_info "[db-backup] dumping SEP + recommendation tables: ${SEP_BACKUP}"
docker exec -i offer-catcher-mysql sh -c "MYSQL_PWD='${MYSQL_ROOT_PASSWORD}' mysqldump -uroot ${DB_NAME} semantic_profiles semantic_profile_tags recommendation_results 2>/dev/null" > "${SEP_BACKUP}"

log_info "[db-backup] dumping schema only: ${SCHEMA_BACKUP}"
docker exec -i offer-catcher-mysql sh -c "MYSQL_PWD='${MYSQL_ROOT_PASSWORD}' mysqldump -uroot --databases ${DB_NAME} --no-data --routines --triggers --events 2>/dev/null" > "${SCHEMA_BACKUP}"

log_info "[db-backup] creating archive: ${ARCHIVE}"
tar -czf "${ARCHIVE}" -C "${BACKUP_DIR}" "$(basename "${FULL_BACKUP}")" "$(basename "${SEP_BACKUP}")" "$(basename "${SCHEMA_BACKUP}")"

log_info "[db-backup] done"
ls -lh "${FULL_BACKUP}" "${SEP_BACKUP}" "${SCHEMA_BACKUP}" "${ARCHIVE}"
