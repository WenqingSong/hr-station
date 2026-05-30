#!/usr/bin/env bash
# ============================================================
# Offer Catcher - 开发脚本通用函数
# ============================================================

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 项目根目录（脚本所在目录的上上级）
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"

# 运行时目录
RUN_DIR="${PROJECT_ROOT}/.run"
LOGS_DIR="${PROJECT_ROOT}/.logs"

# 日志函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $*"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $*"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $*"
}

log_step() {
    echo -e "${BLUE}--- $* ---${NC}"
}

# 确保目录存在
ensure_dir() {
    local dir="$1"
    if [ ! -d "$dir" ]; then
        mkdir -p "$dir"
    fi
}

# 加载 .env 文件（不打印内容）
# 将 .env 中的变量导出为当前 shell 环境变量
# 注意：使用逐行解析避免 DSN 等含特殊字符的值导致 source 语法错误
load_env() {
    local env_file="${PROJECT_ROOT}/.env"
    if [ ! -f "$env_file" ]; then
        return
    fi
    while IFS='=' read -r key value; do
        # 跳过注释和空行
        [[ -z "$key" || "$key" =~ ^[[:space:]]*# ]] && continue
        # 去除 key 前后空格
        key=$(echo "$key" | xargs)
        # 跳过空 key
        [ -z "$key" ] && continue
        # export 变量
        export "$key=$value"
    done < "$env_file"
}

# 获取 .env 中的变量值（不加载整个文件）
get_env_value() {
    local key="$1"
    local default="${2:-}"
    local env_file="${PROJECT_ROOT}/.env"

    if [ -f "$env_file" ]; then
        # 读取 .env 中的值，忽略注释和空行
        local value
        value=$(grep -E "^${key}=" "$env_file" 2>/dev/null | head -1 | cut -d'=' -f2-)
        if [ -n "$value" ]; then
            echo "$value"
            return
        fi
    fi

    # 尝试从环境变量获取
    if [ -n "${!key:-}" ]; then
        echo "${!key}"
        return
    fi

    echo "$default"
}

# 掩码密钥，只显示前4位和后4位
mask_secret() {
    local value="$1"
    if [ -z "$value" ]; then
        echo "未配置"
        return
    fi
    local len=${#value}
    if [ "$len" -le 8 ]; then
        echo "已配置"
        return
    fi
    echo "${value:0:4}...${value: -4}"
}

# 检查进程是否在运行（通过 PID 文件）
is_process_running() {
    local pid_file="$1"
    if [ ! -f "$pid_file" ]; then
        return 1
    fi
    local pid
    pid=$(cat "$pid_file" 2>/dev/null)
    if [ -z "$pid" ]; then
        return 1
    fi
    if kill -0 "$pid" 2>/dev/null; then
        return 0
    else
        # 进程已死，清理 PID 文件
        rm -f "$pid_file"
        return 1
    fi
}

# 递归获取 PID 的所有后代进程（深度优先）
# 用于处理 npm → sh → node 等多层子进程树
get_all_descendants() {
    local pid="$1"
    local children
    children=$(pgrep -P "$pid" 2>/dev/null || true)
    for child in $children; do
        echo "$child"
        get_all_descendants "$child"
    done
}

# 停止通过 PID 文件管理的进程（含完整进程树清理）
stop_pid_file() {
    local pid_file="$1"
    local name="$2"

    if [ ! -f "$pid_file" ]; then
        log_info "${name}: 未在运行（PID 文件不存在）"
        return 0
    fi

    local pid
    pid=$(cat "$pid_file" 2>/dev/null)

    if [ -z "$pid" ]; then
        rm -f "$pid_file"
        log_info "${name}: 已清理无效 PID 文件"
        return 0
    fi

    if kill -0 "$pid" 2>/dev/null; then
        log_info "${name}: 正在停止进程 (PID: ${pid})..."

        # 收集所有后代 PID（递归查找，处理 npm→sh→node 等多层进程树）
        local all_descendants
        all_descendants=$(get_all_descendants "$pid")

        # 先终止所有后代进程（从叶子到根）
        if [ -n "$all_descendants" ]; then
            for desc_pid in $all_descendants; do
                kill "$desc_pid" 2>/dev/null || true
            done
        fi

        # 终止主进程
        kill "$pid" 2>/dev/null

        # 等待所有进程退出
        local wait_count=0
        local still_running=1
        while [ $still_running -eq 1 ] && [ $wait_count -lt 10 ]; do
            still_running=0
            if kill -0 "$pid" 2>/dev/null; then
                still_running=1
            fi
            for desc_pid in $all_descendants; do
                if kill -0 "$desc_pid" 2>/dev/null; then
                    still_running=1
                    break
                fi
            done
            if [ $still_running -eq 1 ]; then
                sleep 1
                wait_count=$((wait_count + 1))
            fi
        done

        # 强制终止仍在运行的进程
        local force_needed=0
        for desc_pid in $all_descendants; do
            if kill -0 "$desc_pid" 2>/dev/null; then
                kill -9 "$desc_pid" 2>/dev/null || true
                force_needed=1
            fi
        done
        if kill -0 "$pid" 2>/dev/null; then
            kill -9 "$pid" 2>/dev/null || true
            force_needed=1
        fi
        if [ $force_needed -eq 1 ]; then
            log_warn "${name}: 强制停止进程"
        fi
        log_info "${name}: 已停止"
    else
        log_info "${name}: 进程已不在运行，清理 PID 文件"
    fi

    rm -f "$pid_file"
}

# 检查端口是否被占用
check_port() {
    local port="$1"
    if lsof -i ":${port}" >/dev/null 2>&1; then
        return 0  # 端口被占用
    else
        return 1  # 端口可用
    fi
}

# 等待端口可用
wait_for_port() {
    local port="$1"
    local max_wait="${2:-30}"
    local count=0
    while [ $count -lt $max_wait ]; do
        if check_port "$port"; then
            return 0
        fi
        sleep 1
        count=$((count + 1))
    done
    return 1
}

# 确认危险操作
confirm_dangerous() {
    local action="$1"
    if [ "${CONFIRM:-}" != "yes" ]; then
        log_error "危险操作：${action}"
        echo ""
        echo "如确认执行，请使用：make ${action} CONFIRM=yes"
        return 1
    fi
    return 0
}

# 初始化运行时目录
init_runtime_dirs() {
    ensure_dir "$RUN_DIR"
    ensure_dir "$LOGS_DIR"
}
