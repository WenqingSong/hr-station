#!/usr/bin/env bash
# ============================================================
# Offer捕手 - 生产部署前置检查脚本（P1 增强版）
#
# 职责：只做安全检查，不执行部署
# 用法：
#   bash scripts/deploy/prod-check.sh
#   make prod-check   （推荐通过 Makefile 调用）
#
# 本脚本绝对禁止：
#   - 执行 docker compose up / down / build / restart
#   - 修改任何数据库
#   - 删除任何卷或数据
#   - 连接生产服务器
# ============================================================

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

ENV_FILE=".env.prod"
COMPOSE_FILE="docker-compose.prod.yml"
BOOTSTRAP_FILE="docker-compose.prod.bootstrap.yml"
CONFIG_CHECK_OUT="/tmp/offer-catcher-prod-compose-config.check"
BACKUP_DIR="backups/mysql"
PROJECT_NAME_DEFAULT="wenqing-hr-demo"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

ok=true

echo "╔══════════════════════════════════════════════════════════╗"
echo "║  Offer捕手 - 生产部署前置检查 P1（只读，不部署）         ║"
echo "╚══════════════════════════════════════════════════════════╝"
echo ""

# ── 1. 文件存在性检查 ──
echo "── 1. 必需文件检查 ──"

if [[ ! -f "$ENV_FILE" ]]; then
    echo -e "${RED}[ERROR]${NC} .env.prod not found."
    echo ""
    echo "  Create it from the template and fill in real secrets:"
    echo "    cp .env.prod.example .env.prod"
    echo "  Then edit .env.prod to replace all CHANGE_ME placeholders."
    echo ""
    ok=false
else
    echo -e "${GREEN}[OK]${NC} .env.prod"
fi

if [[ ! -f "$COMPOSE_FILE" ]]; then
    echo -e "${RED}[ERROR]${NC} docker-compose.prod.yml not found."
    ok=false
else
    echo -e "${GREEN}[OK]${NC} docker-compose.prod.yml"
fi

if [[ ! -f ".env.prod.example" ]]; then
    echo -e "${YELLOW}[WARN]${NC} .env.prod.example not found (template for future use)"
fi

if [[ ! -f "$BOOTSTRAP_FILE" ]]; then
    echo -e "${YELLOW}[WARN]${NC} docker-compose.prod.bootstrap.yml not found (needed for brand-new empty DB init)"
fi

echo ""

# 如果关键文件缺失，不再继续
if ! $ok; then
    exit 1
fi

# ── 2. Compose project name 检查 ──
echo "── 2. Compose Project Name 检查 ──"

project_name=""
if grep -qE '^COMPOSE_PROJECT_NAME=.' "$ENV_FILE" 2>/dev/null; then
    project_name=$(grep '^COMPOSE_PROJECT_NAME=' "$ENV_FILE" | tail -1 | cut -d= -f2- | sed -E 's/^["\x27]//; s/["\x27]$//')
    if [[ -z "$project_name" ]]; then
        project_name="$PROJECT_NAME_DEFAULT"
        echo -e "${YELLOW}[WARN]${NC} COMPOSE_PROJECT_NAME is empty in .env.prod, defaults to ${PROJECT_NAME_DEFAULT}"
    elif echo "$project_name" | grep -q "CHANGE_ME"; then
        echo -e "${RED}[ERROR]${NC} COMPOSE_PROJECT_NAME still has CHANGE_ME placeholder: ${project_name}"
        ok=false
    else
        echo -e "${GREEN}[OK]${NC} COMPOSE_PROJECT_NAME=${project_name}"
    fi
else
    project_name="$PROJECT_NAME_DEFAULT"
    echo -e "${YELLOW}[WARN]${NC} COMPOSE_PROJECT_NAME not set in .env.prod."
    echo "  Default: ${PROJECT_NAME_DEFAULT}"
    echo "  It is recommended to add COMPOSE_PROJECT_NAME to .env.prod."
fi

# 验证 docker-compose.prod.yml 中包含 name: 声明
if grep -qE '^name:\s*(\$|wenqing)' "$COMPOSE_FILE" 2>/dev/null; then
    echo -e "${GREEN}[OK]${NC} compose file declares project name"
else
    echo -e "${YELLOW}[WARN]${NC} compose file does not declare a fixed project name"
fi

echo ""

# ── 3. init.sql 挂载风险检查（P1 升级为 ERROR）──
echo "── 3. init.sql 挂载风险检查 ──"

if grep -q "server/sql/init.sql" "$COMPOSE_FILE" 2>/dev/null; then
    echo -e "${RED}[ERROR]${NC} Default prod compose still mounts server/sql/init.sql."
    echo "  init.sql contains DROP TABLE statements — unsafe for production."
    echo "  Remove the init.sql mount from docker-compose.prod.yml."
    echo "  For brand-new empty DB init, use docker-compose.prod.bootstrap.yml separately."
    ok=false
else
    echo -e "${GREEN}[OK]${NC} Default prod compose does NOT mount init.sql"
fi

# 如果 bootstrap 存在且挂载了 init.sql，提示
if [[ -f "$BOOTSTRAP_FILE" ]]; then
    if grep -q "server/sql/init.sql" "$BOOTSTRAP_FILE" 2>/dev/null; then
        echo -e "${YELLOW}[INFO]${NC} init.sql is in bootstrap override (${BOOTSTRAP_FILE}) — only for brand-new empty DB"
    fi
fi

echo ""

# ── 4. 关键变量存在性检查（不打印值）──
echo "── 4. 关键生产变量检查 ──"

required_vars=(
    "MYSQL_ROOT_PASSWORD"
    "DB_NAME"
    "DB_DSN"
    "JWT_SECRET"
    "LLM_ENABLE"
    "LLM_PROVIDER"
    "LLM_API_KEY"
    "LLM_MODEL"
)

for key in "${required_vars[@]}"; do
    if ! grep -qE "^${key}=.+" "$ENV_FILE" 2>/dev/null; then
        echo -e "${RED}[ERROR]${NC} missing required variable: ${key}"
        ok=false
    else
        val=$(grep "^${key}=" "$ENV_FILE" | head -1 | cut -d= -f2- | tr -d '"' | tr -d "'")
        if echo "$val" | grep -q "CHANGE_ME"; then
            echo -e "${YELLOW}[WARN]${NC} ${key}=<placeholder still set, needs change>"
        else
            echo -e "${GREEN}[OK]${NC} ${key}=<set>"
        fi
    fi
done

echo ""

# ── 5. Feature flag 验证 ──
echo "── 5. 生产 Feature 标志检查 ──"

declare -A feature_flags=(
    ["ENABLE_MOCK_AGENT"]="false"
    ["LLM_FALLBACK_ENABLE"]="false"
)

for key in "${!feature_flags[@]}"; do
    want="${feature_flags[$key]}"
    actual=$(grep "^${key}=" "$ENV_FILE" 2>/dev/null | head -1 | cut -d= -f2- | tr -d '"' | tr -d "'" || echo "")
    if [[ "$actual" = "$want" ]]; then
        echo -e "${GREEN}[OK]${NC} ${key}=${want}"
    else
        echo -e "${YELLOW}[WARN]${NC} ${key}=${actual:-<not set>} (expected ${want})"
    fi
done

echo ""

# ── 6. .env.prod source 安全性警告 ──
echo "── 6. .env.prod 加载方式检查 ──"
echo -e "${YELLOW}[INFO]${NC} Do NOT source .env.prod in shell scripts."
echo "  Reason: DB_DSN contains tcp(), ?, &, = — shell-special characters."
echo "  Correct: docker compose --env-file .env.prod -f docker-compose.prod.yml config"
echo ""

# ── 7. Compose config 校验 ──
echo "── 7. Compose 配置校验 ──"

if docker compose --project-name "$project_name" --env-file "$ENV_FILE" -f "$COMPOSE_FILE" config >"$CONFIG_CHECK_OUT" 2>&1; then
    echo -e "${GREEN}[OK]${NC} compose config: passes"
    echo "  project: ${project_name}"
    echo "  Full output: ${CONFIG_CHECK_OUT}"
else
    echo -e "${RED}[ERROR]${NC} compose config: failed"
    echo "  --- error output ---"
    cat "$CONFIG_CHECK_OUT"
    ok=false
fi

echo ""

# ── 8. 备份目录检查 ──
echo "── 8. 备份目录检查 ──"

if [[ -d "$BACKUP_DIR" ]]; then
    backup_count=$(find "$BACKUP_DIR" -maxdepth 1 -type f | wc -l)
    echo -e "${GREEN}[OK]${NC} backup directory exists (${backup_count} file(s))"
    echo "  Location: ${BACKUP_DIR}/"
else
    echo -e "${YELLOW}[WARN]${NC} backup directory not found: ${BACKUP_DIR}"
    echo "  Create it with: mkdir -p ${BACKUP_DIR}"
fi

echo ""

# ── 9. 部署安全禁令 ──
echo "── 9. 生产安全禁令（每次部署前必读）──"
echo ""
echo "  ✅ Always use --project-name and --env-file .env.prod"
echo "  ❌ Never source .env.prod"
echo "  ❌ Never omit --env-file .env.prod"
echo "  ❌ Never omit --project-name"
echo "  ❌ Never run docker compose down -v in production"
echo "  ❌ Never delete mysql_data / redis_data volumes"
echo "  ❌ Never run server/sql/init.sql manually against production"
echo "  ❌ Never run make reset / make clean-volume / make db-reset against production"
echo "  ❌ Never change COMPOSE_PROJECT_NAME after first deployment"
echo ""

# ── 部署步骤提示 ──
echo "── 部署前正确流程 ──"
echo ""
echo "  1. scripts/deploy/prod-check.sh     (安全检查，只读)"
echo "  2. scripts/deploy/prod-backup.sh    (数据库备份)"
echo "  3. docker compose --project-name ${project_name} --env-file .env.prod -f docker-compose.prod.yml up -d --build"
echo "  4. docker compose --project-name ${project_name} -f docker-compose.prod.yml ps"
echo "  5. curl -f http://localhost/healthz"
echo ""

# ── 最终判定 ──
echo "╔══════════════════════════════════════════════════════════╗"
if $ok; then
    echo -e "║  ${GREEN}prod-check PASSED${NC}                                          ║"
    echo "╚══════════════════════════════════════════════════════════╝"
    exit 0
else
    echo -e "║  ${RED}prod-check FAILED${NC}                                          ║"
    echo "╚══════════════════════════════════════════════════════════╝"
    echo ""
    echo "Fix the errors above before attempting production deployment."
    exit 1
fi
