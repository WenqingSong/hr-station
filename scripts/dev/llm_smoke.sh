#!/usr/bin/env bash
# ============================================================
# LLM Smoke Test Script
# 用于验证 HunyuanClient 与 Hunyuan / TokenHub 的连通性
#
# 安全规则：
#   1. 不打印真实 API Key
#   2. 不打印完整 raw output
#   3. 不修改 .env
#   4. 真实调用需要 RUN_REAL_LLM_SMOKE=1
# ============================================================

set -euo pipefail

# --------------------------------------------------
# 加载通用函数
# --------------------------------------------------
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Source common.sh for log functions and PROJECT_ROOT
if [[ -f "${SCRIPT_DIR}/common.sh" ]]; then
  # shellcheck source=common.sh
  source "${SCRIPT_DIR}/common.sh"
else
  echo "Error: common.sh not found at ${SCRIPT_DIR}/common.sh" >&2
  exit 1
fi

# --------------------------------------------------
# .env 加载（安全方式）
# --------------------------------------------------
load_env_silent() {
  local env_file="${PROJECT_ROOT}/.env"
  if [[ -f "$env_file" ]]; then
    # Source .env silently, do not print contents
    set -a
    # shellcheck disable=SC1090
    source "$env_file" 2>/dev/null || true
    set +a
  fi
}

# --------------------------------------------------
# 命令实现
# --------------------------------------------------

do_help() {
  echo ""
  echo "LLM Smoke Test 命令"
  echo "===================="
  echo ""
  echo "用法: scripts/dev/llm_smoke.sh <command>"
  echo ""
  echo "命令:"
  echo "  help     显示此帮助信息"
  echo "  check    检查 LLM 配置（不请求外部模型）"
  echo "  smoke    执行 LLM smoke test（需要 RUN_REAL_LLM_SMOKE=1）"
  echo ""
  echo "安全说明:"
  echo "  - check 命令不会请求外部模型"
  echo "  - smoke 命令需要显式设置 RUN_REAL_LLM_SMOKE=1"
  echo "  - 不会打印 API Key"
  echo "  - 不会打印完整 raw output"
  echo "  - 不会修改 .env"
  echo ""
  echo "示例:"
  echo "  make llm-check                    # 检查配置"
  echo "  make llm-smoke                    # 提示需要 RUN_REAL_LLM_SMOKE=1"
  echo "  RUN_REAL_LLM_SMOKE=1 make llm-smoke  # 执行真实 smoke test"
  echo ""
}

do_check() {
  log_info "LLM 配置检查"
  log_info "============="
  echo ""

  # Load .env if available
  load_env_silent

  # Check .env existence
  local env_file="${PROJECT_ROOT}/.env"
  if [[ -f "$env_file" ]]; then
    log_info ".env 文件: 存在"
  else
    log_info ".env 文件: 不存在（将使用环境变量）"
  fi
  echo ""

  # Check LLM_ENABLE
  local llm_enable="${LLM_ENABLE:-false}"
  log_info "LLM_ENABLE=$llm_enable"

  # Check LLM_PROVIDER
  local llm_provider="${LLM_PROVIDER:-hunyuan}"
  log_info "LLM_PROVIDER=$llm_provider"

  # Check LLM_BASE_URL
  local llm_base_url="${LLM_BASE_URL:-}"
  if [[ -n "$llm_base_url" ]]; then
    log_info "LLM_BASE_URL=已配置"
  else
    log_info "LLM_BASE_URL=未配置"
  fi

  # Check LLM_API_KEY (do not print real value)
  local llm_api_key="${LLM_API_KEY:-}"
  if [[ -n "$llm_api_key" ]]; then
    log_info "LLM_API_KEY=已配置（长度: ${#llm_api_key}）"
  else
    log_info "LLM_API_KEY=未配置"
  fi

  # Check LLM_MODEL
  local llm_model="${LLM_MODEL:-}"
  if [[ -n "$llm_model" ]]; then
    log_info "LLM_MODEL=$llm_model"
  else
    log_info "LLM_MODEL=未配置"
  fi

  # Check other config
  log_info "LLM_TEMPERATURE=${LLM_TEMPERATURE:-0.2}"
  log_info "LLM_MAX_OUTPUT_TOKENS=${LLM_MAX_OUTPUT_TOKENS:-8192}"
  log_info "LLM_TIMEOUT_SECONDS=${LLM_TIMEOUT_SECONDS:-30}"
  echo ""

  # Check smoke command exists
  local smoke_cmd="${PROJECT_ROOT}/server/cmd/llm-smoke"
  if [[ -d "$smoke_cmd" ]]; then
    log_info "server/cmd/llm-smoke: 存在"
  else
    log_info "server/cmd/llm-smoke: 不存在"
  fi

  # Check HunyuanClient compiles
  echo ""
  log_info "编译检查..."
  if (cd "${PROJECT_ROOT}/server" && go build ./internal/llm/... 2>/dev/null); then
    log_info "HunyuanClient 编译: 通过"
  else
    log_error "HunyuanClient 编译: 失败"
  fi
  echo ""

  # Summary
  log_info "配置摘要:"
  if [[ "$llm_enable" != "true" ]]; then
    log_info "  LLM 未启用（LLM_ENABLE=$llm_enable）"
  elif [[ -z "$llm_base_url" || -z "$llm_api_key" || -z "$llm_model" ]]; then
    log_warn "  LLM 已启用但配置不完整"
    [[ -z "$llm_base_url" ]] && log_warn "    缺少: LLM_BASE_URL"
    [[ -z "$llm_api_key" ]] && log_warn "    缺少: LLM_API_KEY"
    [[ -z "$llm_model" ]] && log_warn "    缺少: LLM_MODEL"
  else
    log_info "  LLM 配置完整，可执行 smoke test"
    log_info "  执行方式: RUN_REAL_LLM_SMOKE=1 make llm-smoke"
  fi
  echo ""
  log_info "安全提醒:"
  log_info "  - 本脚本不打印真实 API Key"
  log_info "  - 本脚本不请求外部模型（仅 check 命令）"
  log_info "  - smoke 命令需要 RUN_REAL_LLM_SMOKE=1 才会请求真实模型"
}

do_smoke() {
  # Check RUN_REAL_LLM_SMOKE gate
  if [[ "${RUN_REAL_LLM_SMOKE:-}" != "1" ]]; then
    echo ""
    log_info "真实 LLM smoke test 未执行：需要显式设置 RUN_REAL_LLM_SMOKE=1"
    echo ""
    log_info "执行方式:"
    log_info "  RUN_REAL_LLM_SMOKE=1 make llm-smoke"
    log_info "  或"
    log_info "  RUN_REAL_LLM_SMOKE=1 scripts/dev/llm_smoke.sh smoke"
    echo ""
    return 0
  fi

  # Load .env if available
  load_env_silent

  # Verify basic config before calling
  local llm_enable="${LLM_ENABLE:-false}"
  local llm_base_url="${LLM_BASE_URL:-}"
  local llm_api_key="${LLM_API_KEY:-}"
  local llm_model="${LLM_MODEL:-}"

  if [[ "$llm_enable" != "true" ]]; then
    echo ""
    log_warn "LLM smoke test 跳过: LLM_ENABLE 不是 true"
    return 1
  fi

  if [[ -z "$llm_base_url" ]]; then
    echo ""
    log_warn "LLM smoke test 跳过: LLM_BASE_URL 未配置"
    return 1
  fi

  if [[ -z "$llm_api_key" ]]; then
    echo ""
    log_warn "LLM smoke test 跳过: LLM_API_KEY 未配置"
    return 1
  fi

  if [[ -z "$llm_model" ]]; then
    echo ""
    log_warn "LLM smoke test 跳过: LLM_MODEL 未配置"
    return 1
  fi

  echo ""
  log_info "执行 LLM smoke test..."
  log_info "  请求次数: 1"
  log_info "  不输出 API Key"
  log_info "  不输出完整 raw output"
  echo ""

  # Run the Go smoke command
  cd "${PROJECT_ROOT}/server"
  go run ./cmd/llm-smoke
}

# --------------------------------------------------
# 命令分发
# --------------------------------------------------

case "${1:-help}" in
  help)
    do_help
    ;;
  check)
    do_check
    ;;
  smoke)
    do_smoke
    ;;
  *)
    echo "未知命令: $1"
    echo "使用: scripts/dev/llm_smoke.sh help"
    exit 1
    ;;
esac
