#!/usr/bin/env bash
# ensure_env.sh — 确保 .env 存在 + 补齐双模型配置变量
# 优先级：.env 已存在 → .env.local → .env.example → 报错
# 不打印文件内容，不打印变量值，不打印 API Key
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

# ── 自愈补齐：确保指定环境变量存在于 .env（幂等，不覆盖已有值）──
ensure_env_var() {
  local key="$1"
  local value="$2"

  if [ ! -f ".env" ]; then
    return 0
  fi

  if ! grep -q "^${key}=" .env; then
    {
      echo ""
      echo "# Added by ensure-env for dual-model routing (P0-FIX)"
      echo "${key}=${value}"
    } >> .env
    echo "[ensure-env] added missing ${key}=${value} to .env"
  fi
}

if [ -f ".env" ]; then
  echo "[ensure-env] .env already exists, skip creating."

  # P0-FIX：已存在的 .env 可能缺失双模型变量，自愈补齐
  ensure_env_var "LLM_FAST_MODEL" "deepseek-v4-flash"
  ensure_env_var "LLM_QUALITY_MODEL" "deepseek-v4-pro-202606"
  ensure_env_var "LLM_DEFAULT_TIER" "fast"

  exit 0
fi

if [ -f ".env.local" ]; then
  cp ".env.local" ".env"
  echo "[ensure-env] created .env from .env.local."
  exit 0
fi

if [ -f ".env.example" ]; then
  cp ".env.example" ".env"
  echo "[ensure-env] created .env from .env.example."
  echo "[ensure-env] warning: .env was created from template; fill local secrets if needed."

  # P0-FIX：从 .env.example 复制后也确保双模型变量存在
  ensure_env_var "LLM_FAST_MODEL" "deepseek-v4-flash"
  ensure_env_var "LLM_QUALITY_MODEL" "deepseek-v4-pro-202606"
  ensure_env_var "LLM_DEFAULT_TIER" "fast"

  exit 0
fi

echo "[ensure-env] error: .env does not exist and neither .env.local nor .env.example was found." >&2
exit 1
