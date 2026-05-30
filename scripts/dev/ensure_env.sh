#!/usr/bin/env bash
# ensure_env.sh — 确保 .env 存在
# 优先级：.env 已存在 → .env.local → .env.example → 报错
# 不打印文件内容，不打印变量值，不打印 API Key
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

if [ -f ".env" ]; then
  echo "[ensure-env] .env already exists, skip creating."
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
  exit 0
fi

echo "[ensure-env] error: .env does not exist and neither .env.local nor .env.example was found." >&2
exit 1
