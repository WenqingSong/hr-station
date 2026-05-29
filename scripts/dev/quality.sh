#!/usr/bin/env bash
# ============================================================
# Offer Catcher - 代码质量检查
# ============================================================

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "${SCRIPT_DIR}/common.sh"

ACTION="${1:-lint}"

do_lint() {
    log_step "代码风格检查"

    # Go 格式检查
    if [ -f "${PROJECT_ROOT}/server/go.mod" ]; then
        log_info "检查 Go 代码格式..."
        local unformatted
        unformatted=$(cd "${PROJECT_ROOT}/server" && gofmt -l . 2>/dev/null || true)
        if [ -n "$unformatted" ]; then
            log_warn "以下 Go 文件需要格式化:"
            echo "$unformatted"
            echo "  修复: cd server && gofmt -w ."
        else
            log_info "Go 代码格式检查通过"
        fi
    else
        log_warn "server/go.mod 不存在，跳过 Go 检查"
    fi

    # Shell 脚本检查
    if [ -d "${PROJECT_ROOT}/scripts/dev" ]; then
        log_info "检查 Shell 脚本语法..."
        local shell_errors=0
        for f in "${PROJECT_ROOT}/scripts/dev"/*.sh; do
            if [ -f "$f" ]; then
                if ! bash -n "$f" 2>/dev/null; then
                    log_warn "Shell 语法错误: $f"
                    shell_errors=$((shell_errors + 1))
                fi
            fi
        done
        if [ $shell_errors -eq 0 ]; then
            log_info "Shell 脚本语法检查通过"
        fi
    fi
}

case "$ACTION" in
    lint)
        do_lint
        ;;

    test)
        log_step "运行后端测试"
        if [ -f "${PROJECT_ROOT}/server/go.mod" ]; then
            cd "${PROJECT_ROOT}/server"
            go test ./...
        else
            log_error "server/go.mod 不存在"
            exit 1
        fi
        ;;

    build)
        log_step "构建检查"

        # 前端构建
        if [ -f "${PROJECT_ROOT}/web/package.json" ]; then
            if [ -d "${PROJECT_ROOT}/web/node_modules" ]; then
                log_info "构建前端..."
                cd "${PROJECT_ROOT}/web"
                npm run build
                log_info "前端构建完成"
            else
                log_warn "web/node_modules 不存在，跳过前端构建（请先执行 make web-install）"
            fi
        else
            log_warn "web/package.json 不存在，跳过前端构建"
        fi

        # 后端构建（输出到临时目录，避免残留 server/server）
        if [ -f "${PROJECT_ROOT}/server/go.mod" ]; then
            log_info "构建后端..."
            cd "${PROJECT_ROOT}/server"
            go build -o /tmp/offer-catcher-server ./cmd/server
            log_info "后端构建完成（二进制输出到 /tmp/offer-catcher-server，不残留 server/server）"
        else
            log_warn "server/go.mod 不存在，跳过后端构建"
        fi
        ;;

    ci)
        log_step "CI 检查（doctor + lint + test + build）"
        echo ""

        # 1. doctor
        "${SCRIPT_DIR}/doctor.sh"
        echo ""

        # 2. lint
        "${SCRIPT_DIR}/quality.sh" lint
        echo ""

        # 3. test
        "${SCRIPT_DIR}/quality.sh" test
        echo ""

        # 4. build
        "${SCRIPT_DIR}/quality.sh" build
        echo ""

        log_info "CI 检查完成"
        ;;

    *)
        echo "用法: $0 {lint|test|build|ci}"
        exit 1
        ;;
esac
