#!/usr/bin/env bash
# ============================================================
# Offer 捕手 - CI 辅助库
#
# 被 .cnb.yml 各阶段的 start-dependencies 引用：
#   bash -c 'source scripts/lib.sh; wait_for_deps'
#
# 说明：
#   .cnb.yml 中原有该引用，但 scripts/lib.sh 一直缺失，
#   导致 start-dependencies 阶段因 "source: No such file" 而失败。
#   本文件补齐该函数，使流水线能够真正等待测试依赖就绪。
# ============================================================

# 等待 docker-compose 中的 MySQL / Redis 变为 healthy。
# 容器名来自 docker-compose.yml 的 container_name（固定，非 compose 项目名前缀）。
wait_for_deps() {
  local max_attempts="${1:-90}"
  local attempt=0
  local mysql_state redis_state

  echo "[wait_for_deps] Waiting for MySQL/Redis to become healthy..."

  while [ "${attempt}" -lt "${max_attempts}" ]; do
    mysql_state=$(docker inspect -f '{{.State.Health.Status}}' offer-catcher-mysql 2>/dev/null || echo "missing")
    redis_state=$(docker inspect -f '{{.State.Health.Status}}' offer-catcher-redis 2>/dev/null || echo "missing")

    if [ "${mysql_state}" = "healthy" ] && [ "${redis_state}" = "healthy" ]; then
      echo "[wait_for_deps] MySQL=healthy Redis=healthy"
      return 0
    fi

    attempt=$((attempt + 1))
    if [ $((attempt % 10)) -eq 0 ]; then
      echo "[wait_for_deps] still waiting (mysql=${mysql_state} redis=${redis_state}, ${attempt}/${max_attempts})"
    fi
    sleep 2
  done

  echo "[wait_for_deps] ERROR: timed out after ${max_attempts} attempts (mysql=${mysql_state} redis=${redis_state})" >&2
  return 1
}
