#!/usr/bin/env bash
# ============================================================
# Offer Catcher - Agent / LLM 环境检查
# ============================================================

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "${SCRIPT_DIR}/common.sh"

ACTION="${1:-check}"

do_env() {
    log_step "LLM 环境变量检查"
    echo ""

    echo "  ENABLE_LLM_AGENT:    $(get_env_value "ENABLE_LLM_AGENT" "未配置")"
    echo "  ENABLE_LLM_FALLBACK: $(get_env_value "ENABLE_LLM_FALLBACK" "未配置")"
    echo "  LLM_PROVIDER:        $(get_env_value "LLM_PROVIDER" "未配置")"
    echo "  LLM_MODEL:           $(get_env_value "LLM_MODEL" "未配置")"
    echo "  LLM_BASE_URL:        $(get_env_value "LLM_BASE_URL" "未配置")"

    # 不打印真实 Key
    local api_key
    api_key=$(get_env_value "LLM_API_KEY" "")
    if [ -n "$api_key" ]; then
        echo "  LLM_API_KEY:         已配置"
    else
        echo "  LLM_API_KEY:         未配置"
    fi

    echo "  LLM_TIMEOUT_SECONDS: $(get_env_value "LLM_TIMEOUT_SECONDS" "未配置")"
    echo "  LLM_MAX_RETRIES:     $(get_env_value "LLM_MAX_RETRIES" "未配置")"
    echo "  LLM_MAX_OUTPUT_TOKENS:$(get_env_value "LLM_MAX_OUTPUT_TOKENS" "未配置")"
    echo "  LLM_TEMPERATURE:     $(get_env_value "LLM_TEMPERATURE" "未配置")"
    echo "  ENABLE_MOCK_AGENT:   $(get_env_value "ENABLE_MOCK_AGENT" "未配置")"
    echo ""
}

do_check() {
    log_step "Agent / LLM 配置检查"
    echo ""

    local has_issue=false

    # 检查 ENABLE_LLM_AGENT
    local enable_llm
    enable_llm=$(get_env_value "ENABLE_LLM_AGENT" "")
    if [ "$enable_llm" = "true" ]; then
        echo -e "  ${GREEN}✅${NC} ENABLE_LLM_AGENT=true（已启用真实 LLM）"
    elif [ "$enable_llm" = "false" ]; then
        echo -e "  ${YELLOW}⚠️${NC}  ENABLE_LLM_AGENT=false（当前使用 Mock Agent）"
    else
        echo -e "  ${YELLOW}⚠️${NC}  ENABLE_LLM_AGENT: 未配置"
        has_issue=true
    fi

    # 检查 LLM_PROVIDER
    local provider
    provider=$(get_env_value "LLM_PROVIDER" "")
    if [ "$provider" = "hunyuan" ]; then
        echo -e "  ${GREEN}✅${NC} LLM_PROVIDER=hunyuan"
    elif [ -n "$provider" ]; then
        echo -e "  ${YELLOW}⚠️${NC}  LLM_PROVIDER=${provider}（非 hunyuan，当前仅支持 hunyuan）"
        has_issue=true
    else
        echo -e "  ${YELLOW}⚠️${NC}  LLM_PROVIDER: 未配置"
        has_issue=true
    fi

    # 检查 LLM_MODEL
    local model
    model=$(get_env_value "LLM_MODEL" "")
    if [ -n "$model" ]; then
        echo -e "  ${GREEN}✅${NC} LLM_MODEL=${model}"
        # 检查疑似拼写错误
        if [ "$model" = "hhy3-preview" ]; then
            echo -e "  ${RED}❌${NC} LLM_MODEL 疑似拼写错误：hhy3-preview → hy3-preview"
            has_issue=true
        fi
    else
        echo -e "  ${YELLOW}⚠️${NC}  LLM_MODEL: 未配置"
        has_issue=true
    fi

    # 检查 LLM_BASE_URL
    local base_url
    base_url=$(get_env_value "LLM_BASE_URL" "")
    if [ -n "$base_url" ]; then
        echo -e "  ${GREEN}✅${NC} LLM_BASE_URL: 已配置"
    else
        echo -e "  ${YELLOW}⚠️${NC}  LLM_BASE_URL: 未配置"
        has_issue=true
    fi

    # 检查 LLM_API_KEY（不打印值）
    local api_key
    api_key=$(get_env_value "LLM_API_KEY" "")
    if [ -n "$api_key" ]; then
        echo -e "  ${GREEN}✅${NC} LLM_API_KEY: 已配置"
    else
        echo -e "  ${YELLOW}⚠️${NC}  LLM_API_KEY: 未配置（阶段 4-2 后启用 LLM 需要）"
    fi

    echo ""
    if [ "$has_issue" = true ]; then
        log_warn "部分配置缺失或异常，请检查 .env"
    else
        log_info "Agent / LLM 配置检查通过"
    fi
}

do_test() {
    log_step "Agent / LLM 连通性测试"
    echo ""
    log_info "LLM Client 尚未实现，请在阶段 4-2 完成后再使用此命令"
    echo ""
    echo "  当前状态："
    echo "    - 阶段 4-1 已完成：Hunyuan API 接入准备"
    echo "    - 阶段 4-2 待执行：实现 LLM Client 抽象与 Hunyuan Client"
    echo ""
    echo "  完成阶段 4-2 后，此命令将测试 LLM API 连通性"
}

case "$ACTION" in
    env)
        do_env
        ;;
    check)
        do_check
        ;;
    test)
        do_test
        ;;
    *)
        echo "用法: $0 {env|check|test}"
        exit 1
        ;;
esac
