#!/usr/bin/env bash
# ============================================================
# Offer Catcher - 腾讯招聘公开岗位数据源拉取脚本
# ============================================================
# 阶段：4-1.2.3
# 说明：
# 1. 从 .env / .env.example 读取腾讯招聘数据源配置
# 2. 低频拉取岗位列表、岗位详情、字典原始 JSON
# 3. 保存原始数据到 data/tencent-jobs/raw/
# 4. 不入库，不修改数据库，不修改后端接口
# 5. 不使用 Cookie / Token / 登录态
# ============================================================

set -euo pipefail

# 加载通用函数
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/common.sh"

# ============================================================
# 配置读取
# ============================================================

# 读取配置：环境变量最高优先 > .env > .env.example > 默认值
get_config() {
    local key="$1"
    local default="${2:-}"
    local value

    # 1. 环境变量最高优先
    value="${!key:-}"
    if [ -n "$value" ]; then
        echo "$value"
        return
    fi

    # 2. .env 文件
    value=$(get_env_value "$key" "")
    if [ -n "$value" ]; then
        echo "$value"
        return
    fi

    # 3. .env.example 兜底
    local env_example="${PROJECT_ROOT}/.env.example"
    if [ -f "$env_example" ]; then
        value=$(grep -E "^${key}=" "$env_example" 2>/dev/null | head -1 | cut -d'=' -f2-)
        if [ -n "$value" ]; then
            echo "$value"
            return
        fi
    fi

    echo "$default"
}

# 加载所有配置
load_tencent_config() {
    TENCENT_JOBS_ENABLE=$(get_config "TENCENT_JOBS_ENABLE" "true")
    TENCENT_JOBS_LIST_URL=$(get_config "TENCENT_JOBS_LIST_URL" "https://join.qq.com/api/v1/position/searchPosition")
    TENCENT_JOBS_DETAIL_URL=$(get_config "TENCENT_JOBS_DETAIL_URL" "https://join.qq.com/api/v1/jobDetails/getJobDetailsByPostId")
    TENCENT_JOBS_DICTIONARY_URL=$(get_config "TENCENT_JOBS_DICTIONARY_URL" "https://join.qq.com/api/v1/dictionary/")
    TENCENT_JOBS_PROJECT_MAPPING_IDS=$(get_config "TENCENT_JOBS_PROJECT_MAPPING_IDS" "2")
    TENCENT_JOBS_PAGE_SIZE=$(get_config "TENCENT_JOBS_PAGE_SIZE" "50")
    TENCENT_JOBS_MAX_PAGES=$(get_config "TENCENT_JOBS_MAX_PAGES" "3")
    TENCENT_JOBS_FETCH_DETAIL=$(get_config "TENCENT_JOBS_FETCH_DETAIL" "true")
    TENCENT_JOBS_FETCH_DICTIONARY=$(get_config "TENCENT_JOBS_FETCH_DICTIONARY" "true")
    TENCENT_JOBS_REQUEST_INTERVAL_MS=$(get_config "TENCENT_JOBS_REQUEST_INTERVAL_MS" "500")
    TENCENT_JOBS_TIMEOUT_SECONDS=$(get_config "TENCENT_JOBS_TIMEOUT_SECONDS" "10")
    TENCENT_JOBS_USER_AGENT=$(get_config "TENCENT_JOBS_USER_AGENT" "Mozilla/5.0")
    TENCENT_JOBS_REFERER=$(get_config "TENCENT_JOBS_REFERER" "https://join.qq.com/post.html?query=p_2")
    TENCENT_JOBS_ORIGIN=$(get_config "TENCENT_JOBS_ORIGIN" "https://join.qq.com")
    TENCENT_JOBS_OUTPUT_DIR=$(get_config "TENCENT_JOBS_OUTPUT_DIR" "data/tencent-jobs/raw")
}

# 毫秒级 sleep
sleep_ms() {
    local ms="${1:-500}"
    if command -v python3 &>/dev/null; then
        python3 -c "import time; time.sleep(${ms}/1000.0)"
    else
        local secs=$(( (ms + 999) / 1000 ))
        sleep "$secs"
    fi
}

# ISO 8601 时间戳
iso_now() {
    date -u +"%Y-%m-%dT%H:%M:%S+08:00" 2>/dev/null || date +"%Y-%m-%dT%H:%M:%S+08:00"
}

# ============================================================
# Python 辅助脚本（通过临时文件传递数据，避免命令行过长）
# ============================================================

# 生成 positions.json
python_merge_positions() {
    local items_file="$1"
    local output_file="$2"
    local fetched_at="$3"
    local project_ids="$4"
    local page_size="$5"
    local max_pages="$6"
    local total="$7"
    local count="$8"

    python3 - "$items_file" "$output_file" "$fetched_at" "$project_ids" "$page_size" "$max_pages" "$total" "$count" << 'PYEOF'
import json, sys
items_file = sys.argv[1]
output_file = sys.argv[2]
fetched_at = sys.argv[3]
project_ids = sys.argv[4]
page_size = int(sys.argv[5])
max_pages = int(sys.argv[6])
total = int(sys.argv[7])
count = int(sys.argv[8])
with open(items_file, 'r') as f:
    items = json.load(f)
pids = [int(x.strip()) for x in project_ids.split(',') if x.strip()]
result = {
    'source': 'tencent_join',
    'fetched_at': fetched_at,
    'project_mapping_ids': pids,
    'page_size': page_size,
    'max_pages': max_pages,
    'total': total,
    'count': count,
    'items': items
}
with open(output_file, 'w', encoding='utf-8') as f:
    json.dump(result, f, ensure_ascii=False, indent=2)
PYEOF
}

# 合并详情到 details.json
python_merge_details() {
    local items_file="$1"
    local output_file="$2"
    local fetched_at="$3"

    python3 - "$items_file" "$output_file" "$fetched_at" << 'PYEOF'
import json, sys
items_file = sys.argv[1]
output_file = sys.argv[2]
fetched_at = sys.argv[3]
with open(items_file, 'r') as f:
    items = json.load(f)
result = {
    'source': 'tencent_join',
    'fetched_at': fetched_at,
    'count': len([i for i in items if i.get('ok')]),
    'items': items
}
with open(output_file, 'w', encoding='utf-8') as f:
    json.dump(result, f, ensure_ascii=False, indent=2)
PYEOF
}

# 生成 metadata.json
python_gen_metadata() {
    local output_dir="$1"
    local fetched_at="$2"
    local list_url="$3"
    local detail_url="$4"
    local dict_url="$5"
    local project_ids="$6"
    local page_size="$7"
    local max_pages="$8"
    local fetch_detail="$9"
    local fetch_dict="${10}"
    local interval_ms="${11}"
    local timeout_s="${12}"

    python3 - "$output_dir" "$fetched_at" "$list_url" "$detail_url" "$dict_url" \
        "$project_ids" "$page_size" "$max_pages" "$fetch_detail" "$fetch_dict" \
        "$interval_ms" "$timeout_s" << 'PYEOF'
import json, sys, os
output_dir = sys.argv[1]
fetched_at = sys.argv[2]
list_url = sys.argv[3]
detail_url = sys.argv[4]
dict_url = sys.argv[5]
project_ids = sys.argv[6]
page_size = int(sys.argv[7])
max_pages = int(sys.argv[8])
fetch_detail = sys.argv[9] == 'true'
fetch_dict = sys.argv[10] == 'true'
interval_ms = int(sys.argv[11])
timeout_s = int(sys.argv[12])
pids = [int(x.strip()) for x in project_ids.split(',') if x.strip()]
positions_count = 0
details_count = 0
dictionary_fetched = False
pf = os.path.join(output_dir, 'positions.json')
df = os.path.join(output_dir, 'details.json')
dicf = os.path.join(output_dir, 'dictionary.json')
if os.path.exists(pf):
    with open(pf, 'r') as f:
        positions_count = json.load(f).get('count', 0)
if os.path.exists(df):
    with open(df, 'r') as f:
        details_count = json.load(f).get('count', 0)
if os.path.exists(dicf):
    dictionary_fetched = True
metadata = {
    'source': 'tencent_join', 'fetched_at': fetched_at,
    'config': {
        'list_url': list_url, 'detail_url': detail_url, 'dictionary_url': dict_url,
        'project_mapping_ids': pids, 'page_size': page_size, 'max_pages': max_pages,
        'fetch_detail': fetch_detail, 'fetch_dictionary': fetch_dict,
        'request_interval_ms': interval_ms, 'timeout_seconds': timeout_s
    },
    'files': {'positions': 'positions.json', 'details': 'details.json', 'dictionary': 'dictionary.json'},
    'summary': {'positions_count': positions_count, 'details_count': details_count, 'dictionary_fetched': dictionary_fetched}
}
mf = os.path.join(output_dir, 'metadata.json')
with open(mf, 'w', encoding='utf-8') as f:
    json.dump(metadata, f, ensure_ascii=False, indent=2)
PYEOF
}

# 解析列表响应
python_parse_list_response() {
    local response_file="$1"
    local items_output="$2"

    python3 - "$response_file" "$items_output" << 'PYEOF'
import json, sys
response_file = sys.argv[1]
items_output = sys.argv[2]
with open(response_file, 'r') as f:
    data = json.load(f)
if data.get('status') != 0:
    print(f'ERROR: API status={data.get("status")}, message={data.get("message", "")}', file=sys.stderr)
    sys.exit(1)
position_list = data.get('data', {}).get('positionList', [])
total_count = data.get('data', {}).get('count', 0)
result = {'items': position_list, 'total': total_count, 'count': len(position_list)}
with open(items_output, 'w', encoding='utf-8') as f:
    json.dump(result, f, ensure_ascii=False)
PYEOF
}

# 解析详情响应并追加到 items 文件
python_append_detail() {
    local items_file="$1"
    local response_file="$2"
    local post_id="$3"

    python3 - "$items_file" "$response_file" "$post_id" << 'PYEOF'
import json, sys
items_file = sys.argv[1]
response_file = sys.argv[2]
post_id = sys.argv[3]
with open(items_file, 'r') as f:
    items = json.load(f)
try:
    with open(response_file, 'r') as f:
        resp = json.load(f)
    ok = resp.get('status') == 0 if isinstance(resp, dict) else False
    detail_data = resp.get('data', {}) if ok else resp
except:
    ok = False
    detail_data = {}
items.append({'postId': post_id, 'ok': ok, 'response': detail_data})
with open(items_file, 'w', encoding='utf-8') as f:
    json.dump(items, f, ensure_ascii=False)
PYEOF
}

# 合并 all_items 文件
python_merge_items() {
    local existing_file="$1"
    local new_file="$2"

    python3 - "$existing_file" "$new_file" << 'PYEOF'
import json, sys
existing_file = sys.argv[1]
new_file = sys.argv[2]
with open(existing_file, 'r') as f:
    existing = json.load(f)
with open(new_file, 'r') as f:
    new_data = json.load(f)
new_items = new_data.get('items', [])
existing.extend(new_items)
with open(existing_file, 'w', encoding='utf-8') as f:
    json.dump(existing, f, ensure_ascii=False)
PYEOF
}

# 提取 postId 列表
python_extract_post_ids() {
    local positions_file="$1"

    python3 - "$positions_file" << 'PYEOF'
import json, sys
with open(sys.argv[1], 'r') as f:
    data = json.load(f)
for item in data.get('items', []):
    post_id = item.get('postId', '')
    if post_id:
        print(post_id)
PYEOF
}

# 读取 JSON 字段
python_read_json_field() {
    local json_file="$1"
    local field="$2"
    local default="${3:-0}"

    python3 - "$json_file" "$field" "$default" << 'PYEOF'
import json, sys
json_file = sys.argv[1]
field = sys.argv[2]
default = sys.argv[3]
try:
    with open(json_file, 'r') as f:
        data = json.load(f)
    # Support nested fields like "data.count"
    value = data
    for key in field.split('.'):
        if isinstance(value, dict):
            value = value.get(key, default)
        else:
            value = default
            break
    print(value)
except:
    print(default)
PYEOF
}

# sample 命令
python_sample() {
    local positions_file="$1"
    local details_file="$2"

    python3 - "$positions_file" "$details_file" << 'PYEOF'
import json, sys, os
positions_file = sys.argv[1]
details_file = sys.argv[2]
with open(positions_file, 'r') as f:
    pos_data = json.load(f)
pos_count = pos_data.get('count', 0)
print(f'岗位数量：{pos_count}')
det_count = 0
if os.path.exists(details_file):
    try:
        with open(details_file, 'r') as f:
            det_data = json.load(f)
        det_count = det_data.get('count', 0)
    except:
        pass
print(f'详情数量：{det_count}')
print()
print('样本岗位：')
items = pos_data.get('items', [])
for item in items[:10]:
    title = item.get('positionTitle', '未知')
    cities = item.get('workCities', '').strip()
    category_code = item.get('positionFamily', '')
    print(f'  - {title} | {cities} | 分类code={category_code}')
if len(items) > 10:
    print(f'  ... 共 {len(items)} 条')
PYEOF
}

# ============================================================
# help
# ============================================================

cmd_help() {
    cat <<EOF
腾讯招聘公开岗位数据源拉取脚本

用法：
  $(basename "$0") <命令>

命令：
  help              查看帮助
  check             检查配置项、URL、输出目录、依赖工具
  source-check      check 的别名
  fetch-list        拉取岗位列表原始 JSON
  fetch-details     基于 positions.json 中的 postId 拉取岗位详情原始 JSON
  fetch-detail      fetch-details 的别名
  fetch-dictionary  拉取字典原始 JSON
  fetch-all         依次执行 fetch-list、fetch-details、fetch-dictionary
  sample            打印本地岗位样本摘要
  status            查看本地原始数据文件状态
  clean             清理 data/tencent-jobs/raw 下生成的 JSON 文件

Makefile 入口：
  make jobs-help              查看帮助
  make jobs-check             检查配置
  make jobs-fetch-list        拉取列表
  make jobs-fetch-detail      拉取详情
  make jobs-fetch-dictionary  拉取字典
  make jobs-fetch             一键拉取
  make jobs-sample            查看样本
  make jobs-status            查看状态
  make jobs-clean             清理

环境变量覆盖：
  TENCENT_JOBS_PAGE_SIZE=5 TENCENT_JOBS_MAX_PAGES=1 $(basename "$0") fetch-all

安全说明：
  - 不使用 Cookie / Token / 登录态
  - 低频请求（默认间隔 500ms）
  - 出现 403/429/验证码时立即停止
  - 完整岗位数据默认不提交仓库
  - 当前只保存原始 JSON，不入库
EOF
}

# ============================================================
# check
# ============================================================

cmd_check() {
    load_tencent_config

    log_step "腾讯招聘公开岗位数据源配置检查"

    # 检查 curl
    if command -v curl &>/dev/null; then
        log_info "curl: 已安装"
    else
        log_error "curl: 未安装（必须）"
        return 1
    fi

    # 检查 python3
    if command -v python3 &>/dev/null; then
        log_info "python3: 已安装"
    else
        log_warn "python3: 未安装（JSON 解析和毫秒 sleep 需要）"
    fi

    echo ""
    log_step "配置项"

    # ENABLE
    if [ "$TENCENT_JOBS_ENABLE" = "true" ]; then
        log_info "TENCENT_JOBS_ENABLE = true（已启用）"
    else
        log_warn "TENCENT_JOBS_ENABLE = false（已禁用，拉取将被跳过）"
    fi

    # URL
    log_info "LIST_URL = ${TENCENT_JOBS_LIST_URL}"
    log_info "DETAIL_URL = ${TENCENT_JOBS_DETAIL_URL}"
    log_info "DICTIONARY_URL = ${TENCENT_JOBS_DICTIONARY_URL}"

    # 分页
    log_info "PROJECT_MAPPING_IDS = ${TENCENT_JOBS_PROJECT_MAPPING_IDS}"
    log_info "PAGE_SIZE = ${TENCENT_JOBS_PAGE_SIZE}"
    log_info "MAX_PAGES = ${TENCENT_JOBS_MAX_PAGES}"

    # 开关
    log_info "FETCH_DETAIL = ${TENCENT_JOBS_FETCH_DETAIL}"
    log_info "FETCH_DICTIONARY = ${TENCENT_JOBS_FETCH_DICTIONARY}"

    # 请求安全
    log_info "REQUEST_INTERVAL_MS = ${TENCENT_JOBS_REQUEST_INTERVAL_MS}"
    log_info "TIMEOUT_SECONDS = ${TENCENT_JOBS_TIMEOUT_SECONDS}"
    log_info "USER_AGENT = ${TENCENT_JOBS_USER_AGENT}"
    log_info "REFERER = ${TENCENT_JOBS_REFERER}"
    log_info "ORIGIN = ${TENCENT_JOBS_ORIGIN}"

    # 输出
    log_info "OUTPUT_DIR = ${TENCENT_JOBS_OUTPUT_DIR}"

    echo ""
    log_step "安全边界"
    log_info "不使用 Cookie / Token / 登录态"
    log_info "低频请求（间隔 ${TENCENT_JOBS_REQUEST_INTERVAL_MS}ms）"
    log_info "Demo 阶段最多 ${TENCENT_JOBS_MAX_PAGES} 页 × ${TENCENT_JOBS_PAGE_SIZE} 条/页"
    log_info "出现 403/429/验证码时立即停止"
}

# ============================================================
# fetch-list
# ============================================================

cmd_fetch_list() {
    load_tencent_config

    if [ "$TENCENT_JOBS_ENABLE" != "true" ]; then
        log_error "TENCENT_JOBS_ENABLE=false，跳过拉取"
        return 1
    fi

    local output_dir="${PROJECT_ROOT}/${TENCENT_JOBS_OUTPUT_DIR}"
    ensure_dir "$output_dir"

    log_step "拉取岗位列表"

    local all_items_file="${output_dir}/.tmp_all_items.json"
    echo '[]' > "$all_items_file"

    local total=0
    local count=0
    local page=1

    while [ "$page" -le "$TENCENT_JOBS_MAX_PAGES" ]; do
        log_info "正在拉取第 ${page} 页（pageSize=${TENCENT_JOBS_PAGE_SIZE}）..."

        # 构建 POST 请求体
        local request_body
        request_body=$(python3 -c "
import json
print(json.dumps({
    'bgList': [],
    'keyword': '',
    'pageIndex': ${page},
    'pageSize': ${TENCENT_JOBS_PAGE_SIZE},
    'positionFidList': [],
    'projectIdList': [],
    'projectMappingIdList': [${TENCENT_JOBS_PROJECT_MAPPING_IDS}],
    'recruitCityList': [],
    'workCityList': [],
    'workCountryType': 0
}))
")

        # 发送请求
        local http_code
        local response_file="${output_dir}/.tmp_list_page_${page}.json"
        http_code=$(curl -s -w "%{http_code}" -o "$response_file" \
            -X POST "$TENCENT_JOBS_LIST_URL" \
            -H "Content-Type: application/json" \
            -H "User-Agent: ${TENCENT_JOBS_USER_AGENT}" \
            -H "Referer: ${TENCENT_JOBS_REFERER}" \
            -H "Origin: ${TENCENT_JOBS_ORIGIN}" \
            --max-time "$TENCENT_JOBS_TIMEOUT_SECONDS" \
            -d "$request_body" 2>/dev/null || echo "000")

        # 检查 HTTP 状态
        if [ "$http_code" = "403" ] || [ "$http_code" = "429" ]; then
            log_error "HTTP ${http_code}：触发限流或风控，立即停止"
            rm -f "$response_file"
            break
        fi

        if [ "$http_code" != "200" ]; then
            log_error "HTTP ${http_code}：请求失败，跳过此页"
            rm -f "$response_file"
            page=$((page + 1))
            continue
        fi

        # 解析响应到临时文件
        local page_items_file="${output_dir}/.tmp_page_items_${page}.json"
        if ! python_parse_list_response "$response_file" "$page_items_file"; then
            log_error "解析第 ${page} 页失败"
            rm -f "$response_file" "$page_items_file"
            page=$((page + 1))
            continue
        fi

        local page_count
        page_count=$(python_read_json_field "$page_items_file" "count" "0")
        total=$(python_read_json_field "$page_items_file" "total" "0")

        # 合并到 all_items
        python_merge_items "$all_items_file" "$page_items_file"

        count=$((count + page_count))
        log_info "第 ${page} 页获取 ${page_count} 条岗位（累计 ${count} 条，总数 ${total}）"

        # 如果本页岗位数 < pageSize，说明已到末尾
        if [ "$page_count" -lt "$TENCENT_JOBS_PAGE_SIZE" ]; then
            log_info "已到末尾页"
            rm -f "$response_file" "$page_items_file"
            break
        fi

        rm -f "$response_file" "$page_items_file"

        # 请求间隔
        if [ "$page" -lt "$TENCENT_JOBS_MAX_PAGES" ]; then
            sleep_ms "$TENCENT_JOBS_REQUEST_INTERVAL_MS"
        fi

        page=$((page + 1))
    done

    # 生成 positions.json
    local fetched_at
    fetched_at=$(iso_now)

    python_merge_positions "$all_items_file" "${output_dir}/positions.json" \
        "$fetched_at" "$TENCENT_JOBS_PROJECT_MAPPING_IDS" \
        "$TENCENT_JOBS_PAGE_SIZE" "$TENCENT_JOBS_MAX_PAGES" "$total" "$count"

    if [ $? -eq 0 ]; then
        log_info "positions.json 已保存（${count} 条岗位）"
    else
        log_error "positions.json 保存失败"
        rm -f "$all_items_file"
        return 1
    fi

    rm -f "$all_items_file"
}

# ============================================================
# fetch-details
# ============================================================

cmd_fetch_details() {
    load_tencent_config

    if [ "$TENCENT_JOBS_ENABLE" != "true" ]; then
        log_error "TENCENT_JOBS_ENABLE=false，跳过拉取"
        return 1
    fi

    if [ "$TENCENT_JOBS_FETCH_DETAIL" != "true" ]; then
        log_warn "TENCENT_JOBS_FETCH_DETAIL=false，跳过详情拉取"
        return 0
    fi

    local output_dir="${PROJECT_ROOT}/${TENCENT_JOBS_OUTPUT_DIR}"
    local positions_file="${output_dir}/positions.json"

    if [ ! -f "$positions_file" ]; then
        log_error "positions.json 不存在，请先执行 fetch-list"
        return 1
    fi

    log_step "拉取岗位详情"

    # 提取所有 postId
    local post_ids
    post_ids=$(python_extract_post_ids "$positions_file")

    if [ -z "$post_ids" ]; then
        log_warn "positions.json 中没有 postId，跳过详情拉取"
        return 0
    fi

    local total_ids
    total_ids=$(echo "$post_ids" | wc -l | tr -d ' ')

    log_info "共 ${total_ids} 个岗位需要拉取详情"

    # 使用临时文件存储详情 items
    local details_items_file="${output_dir}/.tmp_details_items.json"
    echo '[]' > "$details_items_file"

    local success_count=0
    local fail_count=0
    local idx=0

    while IFS= read -r post_id; do
        [ -z "$post_id" ] && continue
        idx=$((idx + 1))

        log_info "[${idx}/${total_ids}] 拉取详情 postId=${post_id}..."

        local http_code
        local response_file="${output_dir}/.tmp_detail_${post_id}.json"
        http_code=$(curl -s -w "%{http_code}" -o "$response_file" \
            -X GET "${TENCENT_JOBS_DETAIL_URL}?postId=${post_id}" \
            -H "User-Agent: ${TENCENT_JOBS_USER_AGENT}" \
            -H "Referer: ${TENCENT_JOBS_REFERER}" \
            -H "Origin: ${TENCENT_JOBS_ORIGIN}" \
            --max-time "$TENCENT_JOBS_TIMEOUT_SECONDS" 2>/dev/null || echo "000")

        if [ "$http_code" = "403" ] || [ "$http_code" = "429" ]; then
            log_error "HTTP ${http_code}：触发限流或风控，立即停止"
            rm -f "$response_file"
            break
        fi

        if [ "$http_code" != "200" ]; then
            log_warn "HTTP ${http_code}：postId=${post_id} 详情请求失败，跳过"
            # 追加失败记录
            python_append_detail "$details_items_file" "/dev/null" "$post_id" 2>/dev/null || true
            # 手动追加失败项
            python3 - "$details_items_file" "$post_id" << 'PYEOF'
import json, sys
items_file = sys.argv[1]
post_id = sys.argv[2]
with open(items_file, 'r') as f:
    items = json.load(f)
items.append({'postId': post_id, 'ok': False, 'response': None})
with open(items_file, 'w', encoding='utf-8') as f:
    json.dump(items, f, ensure_ascii=False)
PYEOF
            fail_count=$((fail_count + 1))
            rm -f "$response_file"
            sleep_ms "$TENCENT_JOBS_REQUEST_INTERVAL_MS"
            continue
        fi

        # 解析并追加到 items 文件
        python_append_detail "$details_items_file" "$response_file" "$post_id"
        success_count=$((success_count + 1))

        rm -f "$response_file"

        # 请求间隔
        sleep_ms "$TENCENT_JOBS_REQUEST_INTERVAL_MS"
    done <<< "$post_ids"

    # 生成 details.json
    local fetched_at
    fetched_at=$(iso_now)

    python_merge_details "$details_items_file" "${output_dir}/details.json" "$fetched_at"

    if [ $? -eq 0 ]; then
        log_info "details.json 已保存（成功 ${success_count}，失败 ${fail_count}）"
    else
        log_error "details.json 保存失败"
        rm -f "$details_items_file"
        return 1
    fi

    rm -f "$details_items_file"
}

# ============================================================
# fetch-dictionary
# ============================================================

cmd_fetch_dictionary() {
    load_tencent_config

    if [ "$TENCENT_JOBS_ENABLE" != "true" ]; then
        log_error "TENCENT_JOBS_ENABLE=false，跳过拉取"
        return 1
    fi

    if [ "$TENCENT_JOBS_FETCH_DICTIONARY" != "true" ]; then
        log_warn "TENCENT_JOBS_FETCH_DICTIONARY=false，跳过字典拉取"
        return 0
    fi

    local output_dir="${PROJECT_ROOT}/${TENCENT_JOBS_OUTPUT_DIR}"
    ensure_dir "$output_dir"

    log_step "拉取字典数据"

    local http_code
    local response_file="${output_dir}/.tmp_dictionary.json"
    http_code=$(curl -s -w "%{http_code}" -o "$response_file" \
        -X GET "${TENCENT_JOBS_DICTIONARY_URL}?types=RecruitType,BusinessGroup,RecruitProjectPostList,RecruitCity" \
        -H "User-Agent: ${TENCENT_JOBS_USER_AGENT}" \
        -H "Referer: ${TENCENT_JOBS_REFERER}" \
        -H "Origin: ${TENCENT_JOBS_ORIGIN}" \
        --max-time "$TENCENT_JOBS_TIMEOUT_SECONDS" 2>/dev/null || echo "000")

    if [ "$http_code" = "403" ] || [ "$http_code" = "429" ]; then
        log_error "HTTP ${http_code}：触发限流或风控，跳过字典拉取"
        rm -f "$response_file"
        return 1
    fi

    if [ "$http_code" != "200" ]; then
        log_warn "HTTP ${http_code}：字典请求失败（P1 可选，不阻塞主流程）"
        rm -f "$response_file"
        return 0
    fi

    # 生成 dictionary.json
    local fetched_at
    fetched_at=$(iso_now)

    python3 - "$response_file" "${output_dir}/dictionary.json" "$fetched_at" << 'PYEOF'
import json, sys
response_file = sys.argv[1]
output_file = sys.argv[2]
fetched_at = sys.argv[3]
with open(response_file, 'r') as f:
    data = json.load(f)
ok = data.get('status') == 0
result = {
    'source': 'tencent_join',
    'fetched_at': fetched_at,
    'ok': ok,
    'response': data.get('data', {})
}
with open(output_file, 'w', encoding='utf-8') as f:
    json.dump(result, f, ensure_ascii=False, indent=2)
PYEOF

    if [ $? -eq 0 ]; then
        log_info "dictionary.json 已保存"
    else
        log_warn "dictionary.json 保存失败（P1 可选，不阻塞主流程）"
    fi

    rm -f "$response_file"
}

# ============================================================
# fetch-all
# ============================================================

cmd_fetch_all() {
    log_step "依次拉取：列表 → 详情 → 字典"

    cmd_fetch_list

    cmd_fetch_details

    cmd_fetch_dictionary

    # 更新 metadata.json
    update_metadata

    log_info "全部拉取完成"
}

# ============================================================
# update_metadata
# ============================================================

update_metadata() {
    load_tencent_config

    local output_dir="${PROJECT_ROOT}/${TENCENT_JOBS_OUTPUT_DIR}"
    local fetched_at
    fetched_at=$(iso_now)

    python_gen_metadata "$output_dir" "$fetched_at" \
        "$TENCENT_JOBS_LIST_URL" "$TENCENT_JOBS_DETAIL_URL" "$TENCENT_JOBS_DICTIONARY_URL" \
        "$TENCENT_JOBS_PROJECT_MAPPING_IDS" "$TENCENT_JOBS_PAGE_SIZE" "$TENCENT_JOBS_MAX_PAGES" \
        "$TENCENT_JOBS_FETCH_DETAIL" "$TENCENT_JOBS_FETCH_DICTIONARY" \
        "$TENCENT_JOBS_REQUEST_INTERVAL_MS" "$TENCENT_JOBS_TIMEOUT_SECONDS"

    if [ $? -eq 0 ]; then
        log_info "metadata.json 已更新"
    else
        log_warn "metadata.json 更新失败（非关键）"
    fi
}

# ============================================================
# sample
# ============================================================

cmd_sample() {
    load_tencent_config

    local output_dir="${PROJECT_ROOT}/${TENCENT_JOBS_OUTPUT_DIR}"
    local positions_file="${output_dir}/positions.json"
    local details_file="${output_dir}/details.json"

    if [ ! -f "$positions_file" ]; then
        log_warn "positions.json 不存在，请先执行 fetch-list"
        return 0
    fi

    log_step "本地岗位样本"

    python_sample "$positions_file" "$details_file"
}

# ============================================================
# status
# ============================================================

cmd_status() {
    load_tencent_config

    local output_dir="${PROJECT_ROOT}/${TENCENT_JOBS_OUTPUT_DIR}"

    log_step "本地原始数据状态"

    echo "输出目录：${TENCENT_JOBS_OUTPUT_DIR}"
    echo ""

    local positions_file="${output_dir}/positions.json"
    local details_file="${output_dir}/details.json"
    local dictionary_file="${output_dir}/dictionary.json"
    local metadata_file="${output_dir}/metadata.json"

    # 文件存在检查
    if [ -f "$positions_file" ]; then
        log_info "positions.json：存在"
    else
        log_warn "positions.json：不存在"
    fi

    if [ -f "$details_file" ]; then
        log_info "details.json：存在"
    else
        log_warn "details.json：不存在"
    fi

    if [ -f "$dictionary_file" ]; then
        log_info "dictionary.json：存在"
    else
        log_warn "dictionary.json：不存在"
    fi

    if [ -f "$metadata_file" ]; then
        log_info "metadata.json：存在"
    else
        log_warn "metadata.json：不存在"
    fi

    # 数量统计
    echo ""
    if [ -f "$positions_file" ]; then
        local positions_count
        positions_count=$(python_read_json_field "$positions_file" "count" "0")
        echo "岗位数量：${positions_count}"
    fi

    if [ -f "$details_file" ]; then
        local details_count
        details_count=$(python_read_json_field "$details_file" "count" "0")
        echo "详情数量：${details_count}"
    fi

    if [ -f "$metadata_file" ]; then
        local fetched_at
        fetched_at=$(python_read_json_field "$metadata_file" "fetched_at" "未知")
        echo "最后拉取时间：${fetched_at}"
    fi
}

# ============================================================
# clean
# ============================================================

cmd_clean() {
    load_tencent_config

    local output_dir="${PROJECT_ROOT}/${TENCENT_JOBS_OUTPUT_DIR}"

    log_step "清理本地原始数据"

    local cleaned=0

    for f in positions.json details.json dictionary.json metadata.json; do
        local file_path="${output_dir}/${f}"
        if [ -f "$file_path" ]; then
            rm -f "$file_path"
            log_info "已删除：${f}"
            cleaned=$((cleaned + 1))
        fi
    done

    if [ "$cleaned" -eq 0 ]; then
        log_info "没有需要清理的文件"
    else
        log_info "已清理 ${cleaned} 个文件"
    fi
}

# ============================================================
# 主入口
# ============================================================

main() {
    local cmd="${1:-help}"

    case "$cmd" in
        help)
            cmd_help
            ;;
        check|source-check)
            cmd_check
            ;;
        fetch-list)
            cmd_fetch_list
            ;;
        fetch-details|fetch-detail)
            cmd_fetch_details
            ;;
        fetch-dictionary)
            cmd_fetch_dictionary
            ;;
        fetch-all)
            cmd_fetch_all
            ;;
        sample)
            cmd_sample
            ;;
        status)
            cmd_status
            ;;
        clean)
            cmd_clean
            ;;
        *)
            log_error "未知命令：${cmd}"
            cmd_help
            exit 1
            ;;
    esac
}

main "$@"
