#!/usr/bin/env bash
# ============================================================
# Offer Catcher - 开发环境检查
# ============================================================

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "${SCRIPT_DIR}/common.sh"

ERRORS=0
WARNINGS=0

check_pass() {
    echo -e "  ${GREEN}✅${NC} $*"
}

check_warn() {
    echo -e "  ${YELLOW}⚠️${NC}  $*"
    WARNINGS=$((WARNINGS + 1))
}

check_fail() {
    echo -e "  ${RED}❌${NC} $*"
    ERRORS=$((ERRORS + 1))
}

echo ""
echo "=========================================="
echo "  开发环境检查 (只读，不修改任何文件)"
echo "=========================================="
echo ""

# --- 工作目录 ---
log_step "工作目录"
echo "  当前目录: $(pwd)"
echo "  项目根目录: ${PROJECT_ROOT}"
echo ""

# --- 工具检查 ---
log_step "工具检查"

if command -v node >/dev/null 2>&1; then
    check_pass "node: $(node --version)"
else
    check_fail "node: 未安装"
fi

if command -v npm >/dev/null 2>&1; then
    check_pass "npm: $(npm --version)"
else
    check_fail "npm: 未安装"
fi

if command -v go >/dev/null 2>&1; then
    check_pass "go: $(go version | awk '{print $3}')"
else
    check_fail "go: 未安装"
fi

if command -v docker >/dev/null 2>&1; then
    check_pass "docker: $(docker --version | awk '{print $3}' | tr -d ',')"
else
    check_fail "docker: 未安装"
fi

if command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1; then
    check_pass "docker compose: 可用"
else
    check_fail "docker compose: 不可用"
fi

if command -v curl >/dev/null 2>&1; then
    check_pass "curl: 可用"
else
    check_warn "curl: 未安装（健康检查需要）"
fi

echo ""

# --- 项目文件检查 ---
log_step "项目文件检查"

if [ -f "${PROJECT_ROOT}/.env" ]; then
    check_pass ".env: 存在"
else
    check_warn ".env: 不存在（执行 make init 可创建）"
fi

if [ -f "${PROJECT_ROOT}/.env.example" ]; then
    check_pass ".env.example: 存在"
else
    check_fail ".env.example: 不存在"
fi

if [ -f "${PROJECT_ROOT}/docker-compose.yml" ]; then
    check_pass "docker-compose.yml: 存在"
else
    check_fail "docker-compose.yml: 不存在"
fi

if [ -f "${PROJECT_ROOT}/server/go.mod" ]; then
    check_pass "server/go.mod: 存在"
else
    check_fail "server/go.mod: 不存在"
fi

if [ -f "${PROJECT_ROOT}/web/package.json" ]; then
    check_pass "web/package.json: 存在"
else
    check_fail "web/package.json: 不存在"
fi

if [ -f "${PROJECT_ROOT}/server/sql/init.sql" ]; then
    check_pass "server/sql/init.sql: 存在"
else
    check_warn "server/sql/init.sql: 不存在"
fi

if [ -d "${PROJECT_ROOT}/web/node_modules" ]; then
    check_pass "web/node_modules: 存在"
else
    check_warn "web/node_modules: 不存在（执行 make web-install 安装）"
fi

if [ -d "${PROJECT_ROOT}/scripts/dev" ]; then
    check_pass "scripts/dev/: 存在"
else
    check_warn "scripts/dev/: 不存在"
fi

echo ""

# --- Docker 服务状态 ---
log_step "Docker 服务状态"

if command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1; then
    if [ -f "${PROJECT_ROOT}/docker-compose.yml" ]; then
        cd "$PROJECT_ROOT"
        docker compose ps 2>/dev/null || check_warn "docker compose ps 执行失败"
    else
        check_warn "docker-compose.yml 不存在，跳过"
    fi
else
    check_warn "Docker 或 Docker Compose 不可用，跳过"
fi

echo ""

# --- 端口占用检查 ---
log_step "端口占用检查"

SERVER_PORT=$(get_env_value "SERVER_PORT" "8080")
WEB_PORT=$(get_env_value "WEB_PORT" "5173")
DB_PORT=$(get_env_value "DB_PORT" "3306")
NGINX_PORT=$(get_env_value "NGINX_PORT" "80")

if check_port "$DB_PORT"; then
    check_warn "端口 ${DB_PORT} (MySQL): 已占用"
else
    check_pass "端口 ${DB_PORT} (MySQL): 可用"
fi

if check_port "$NGINX_PORT"; then
    check_warn "端口 ${NGINX_PORT} (Nginx): 已占用"
else
    check_pass "端口 ${NGINX_PORT} (Nginx): 可用"
fi

if check_port "$SERVER_PORT"; then
    check_warn "端口 ${SERVER_PORT} (后端): 已占用"
else
    check_pass "端口 ${SERVER_PORT} (后端): 可用"
fi

if check_port "$WEB_PORT"; then
    check_warn "端口 ${WEB_PORT} (前端): 已占用"
else
    check_pass "端口 ${WEB_PORT} (前端): 可用"
fi

echo ""

# --- LLM / Agent 配置检查 ---
log_step "LLM / Agent 配置检查"

ENABLE_LLM=$(get_env_value "ENABLE_LLM_AGENT" "")
ENABLE_FALLBACK=$(get_env_value "ENABLE_LLM_FALLBACK" "")
LLM_PROVIDER=$(get_env_value "LLM_PROVIDER" "")
LLM_MODEL=$(get_env_value "LLM_MODEL" "")
LLM_BASE_URL=$(get_env_value "LLM_BASE_URL" "")
LLM_API_KEY=$(get_env_value "LLM_API_KEY" "")

if [ -n "$ENABLE_LLM" ]; then
    echo "  ENABLE_LLM_AGENT: ${ENABLE_LLM}"
else
    check_warn "ENABLE_LLM_AGENT: 未配置"
fi

if [ -n "$ENABLE_FALLBACK" ]; then
    echo "  ENABLE_LLM_FALLBACK: ${ENABLE_FALLBACK}"
else
    check_warn "ENABLE_LLM_FALLBACK: 未配置"
fi

if [ -n "$LLM_PROVIDER" ]; then
    echo "  LLM_PROVIDER: ${LLM_PROVIDER}"
else
    check_warn "LLM_PROVIDER: 未配置"
fi

if [ -n "$LLM_MODEL" ]; then
    echo "  LLM_MODEL: ${LLM_MODEL}"
    # 检查疑似拼写错误
    if [ "$LLM_MODEL" = "hhy3-preview" ]; then
        check_warn "LLM_MODEL 疑似拼写错误：hhy3-preview → hy3-preview"
    fi
else
    check_warn "LLM_MODEL: 未配置"
fi

if [ -n "$LLM_BASE_URL" ]; then
    echo "  LLM_BASE_URL: ${LLM_BASE_URL}"
else
    check_warn "LLM_BASE_URL: 未配置"
fi

# 不打印真实 Key
if [ -n "$LLM_API_KEY" ]; then
    echo "  LLM_API_KEY: 已配置"
else
    check_warn "LLM_API_KEY: 未配置（阶段 4-2 后启用 LLM 需要）"
fi

echo ""

# --- 汇总 ---
echo "=========================================="
if [ $ERRORS -eq 0 ] && [ $WARNINGS -eq 0 ]; then
    log_info "检查完成：所有项通过"
elif [ $ERRORS -eq 0 ]; then
    log_warn "检查完成：${WARNINGS} 个警告"
else
    log_error "检查完成：${ERRORS} 个错误，${WARNINGS} 个警告"
fi
echo "=========================================="
