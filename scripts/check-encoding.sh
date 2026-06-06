#!/usr/bin/env bash
set -euo pipefail

echo "[encoding-check] Scanning for mojibake patterns..."

DIRS="web/src server/internal server/sql scripts docs/09-开发日志 docs/设计文档"
EXCLUDE_DIRS="node_modules dist .git"

EXCLUDE_ARG=""
for d in $EXCLUDE_DIRS; do
  EXCLUDE_ARG="$EXCLUDE_ARG --exclude-dir=$d"
done

# Scan for UTF-8 replacement character and common mojibake patterns
FOUND=$(grep -rIn $EXCLUDE_ARG \
  $'\xef\xbf\xbd' \
  $DIRS 2>/dev/null || true)

if [ -n "$FOUND" ]; then
  echo "[encoding-check] FAIL: Mojibake/encoding errors found:"
  echo "$FOUND"
  exit 1
fi

echo "[encoding-check] OK: No encoding errors detected."
