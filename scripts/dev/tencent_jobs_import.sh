#!/usr/bin/env bash
# ============================================================
# Offer Catcher - 腾讯招聘公开岗位清洗与入库脚本
# ============================================================
# 阶段：4-1.3.2
# 说明：
# 1. 从本地 data/tencent-jobs/raw/ 读取原始 JSON
# 2. 清洗腾讯岗位字段映射到 jobs 表
# 3. 生成导入 SQL（INSERT ... ON DUPLICATE KEY UPDATE）
# 4. 支持导入 MySQL
# 5. 不请求腾讯招聘接口
# 6. 不修改 raw JSON
# 7. 不修改 .env
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

# 加载数据库配置
load_db_config() {
    DB_HOST=$(get_config "DB_HOST" "127.0.0.1")
    DB_PORT=$(get_config "DB_PORT" "3306")
    DB_NAME=$(get_config "DB_NAME" "offer_catcher")
    DB_USER=$(get_config "DB_USER" "root")
    DB_PASSWORD=$(get_config "DB_PASSWORD" "password")
    TENCENT_JOBS_OUTPUT_DIR=$(get_config "TENCENT_JOBS_OUTPUT_DIR" "data/tencent-jobs/raw")
}

# ISO 8601 时间戳
iso_now() {
    date -u +"%Y-%m-%dT%H:%M:%S+08:00" 2>/dev/null || date +"%Y-%m-%dT%H:%M:%S+08:00"
}

# 当前时间（MySQL DATETIME 格式）
mysql_now() {
    date +"%Y-%m-%d %H:%M:%S" 2>/dev/null || echo "2026-05-29 00:00:00"
}

# 路径常量
RAW_DIR="${PROJECT_ROOT}/data/tencent-jobs/raw"
GENERATED_DIR="${PROJECT_ROOT}/data/tencent-jobs/generated"
POSITIONS_FILE="${RAW_DIR}/positions.json"
DETAILS_FILE="${RAW_DIR}/details.json"
DICTIONARY_FILE="${RAW_DIR}/dictionary.json"
PREVIEW_FILE="${GENERATED_DIR}/tencent_jobs_preview.json"
SQL_FILE="${GENERATED_DIR}/tencent_jobs_import.sql"

# MySQL 命令封装
mysql_exec() {
    docker exec -i offer-catcher-mysql mysql \
        -h"${DB_HOST}" -P"${DB_PORT}" \
        -u"${DB_USER}" -p"${DB_PASSWORD}" \
        "${DB_NAME}" "$@" 2>/dev/null
}

mysql_exec_sql() {
    docker exec -i offer-catcher-mysql mysql \
        -h"${DB_HOST}" -P"${DB_PORT}" \
        -u"${DB_USER}" -p"${DB_PASSWORD}" \
        "${DB_NAME}" -e "$1" 2>/dev/null
}

# 检查 MySQL 是否可用
check_mysql() {
    if ! docker exec offer-catcher-mysql mysql \
        -h"${DB_HOST}" -P"${DB_PORT}" \
        -u"${DB_USER}" -p"${DB_PASSWORD}" \
        "${DB_NAME}" -e "SELECT 1" &>/dev/null; then
        return 1
    fi
    return 0
}

# ============================================================
# Python 清洗脚本
# ============================================================

# 预览清洗结果
python_preview() {
    local positions_file="$1"
    local details_file="$2"
    local dictionary_file="$3"
    local output_file="$4"
    local generated_at="$5"

    python3 - "$positions_file" "$details_file" "$dictionary_file" "$output_file" "$generated_at" << 'PYEOF'
import json, sys, os, re
from datetime import datetime

positions_file = sys.argv[1]
details_file = sys.argv[2]
dictionary_file = sys.argv[3]
output_file = sys.argv[4]
generated_at = sys.argv[5]

# 读取 positions
with open(positions_file, 'r', encoding='utf-8') as f:
    pos_data = json.load(f)
positions = pos_data.get('items', [])

# 读取 details
details_map = {}
if os.path.exists(details_file):
    try:
        with open(details_file, 'r', encoding='utf-8') as f:
            det_data = json.load(f)
        for item in det_data.get('items', []):
            post_id = item.get('postId', '')
            if post_id and item.get('ok'):
                details_map[post_id] = item.get('response', {})
    except:
        pass

# 读取 dictionary snapshot
dict_snapshot = {}
if os.path.exists(dictionary_file):
    try:
        with open(dictionary_file, 'r', encoding='utf-8') as f:
            dict_data = json.load(f)
        dict_snapshot['fetched_at'] = dict_data.get('fetched_at', '')
    except:
        pass

# 预置关键词列表
KEYWORDS = [
    'Go', 'Java', 'Python', 'C\\+\\+', '后端', '后台', '算法', '机器学习',
    '深度学习', '云计算', '数据库', 'MySQL', 'Redis', 'Docker', 'Kubernetes',
    '微服务', '前端', 'React', 'Vue', '产品', '运营', '测试', '数据分析',
    '自然语言处理', 'NLP', '多模态', '分布式', 'Linux', '网络', '安全',
    '游戏', '移动端', 'iOS', 'Android', '设计', '嵌入式', '编译器',
    '高性能', '存储', '大数据', 'Spark', 'Flink', 'Hadoop',
    '推荐系统', '搜索', '图像', '计算机视觉', '语音', '大模型', 'LLM',
    'RAG', '向量数据库', '运维', 'SRE', 'DevOps', 'CI/CD'
]

def extract_keywords(title, category, desc, request):
    text = ' '.join(filter(None, [title, category, desc, request]))
    found = []
    for kw_pattern in KEYWORDS:
        if re.search(kw_pattern, text, re.IGNORECASE):
            # 去掉转义字符用于输出
            clean_kw = kw_pattern.replace('\\+', '+').replace('\\', '')
            found.append(clean_kw)
    # 去重
    seen = set()
    result = []
    for kw in found:
        kw_lower = kw.lower()
        if kw_lower not in seen:
            seen.add(kw_lower)
            result.append(kw)
    if not result and category:
        result.append(category)
    if not result and title:
        words = title.split('-')
        if words:
            result.append(words[0].strip())
    return result

def map_city(work_city_list, work_cities_str):
    if work_city_list and isinstance(work_city_list, list):
        return ','.join(work_city_list)
    if work_cities_str:
        return work_cities_str.strip()
    return ''

def map_department(intention_bgd_list, bgs_str):
    if intention_bgd_list and isinstance(intention_bgd_list, list) and len(intention_bgd_list) > 0:
        return intention_bgd_list[0].get('showTitle', '')
    if bgs_str:
        parts = bgs_str.strip().split()
        if parts:
            return parts[0]
    return ''

items = []
full_jd_count = 0
summary_only_count = 0
missing_detail_count = 0

for pos_item in positions:
    post_id = pos_item.get('postId', '')
    if not post_id:
        continue

    detail = details_map.get(post_id, None)

    # 字段映射
    title = ''
    department = ''
    city = ''
    category = ''
    job_type = pos_item.get('projectName', '')
    description = ''
    requirements = ''
    data_quality = 'summary_only'

    if detail:
        # 详情优先
        title = detail.get('title', '') or pos_item.get('positionTitle', '')
        department = map_department(detail.get('intentionBGDList', []), pos_item.get('bgs', ''))
        city = map_city(detail.get('workCityList', []), pos_item.get('workCities', ''))
        category = detail.get('tidName', '') or ''
        description = detail.get('desc', '') or pos_item.get('positionTitle', '')
        requirements = detail.get('request', '') or ''
        data_quality = 'full_jd' if (detail.get('desc') or detail.get('request')) else 'summary_only'
    else:
        # 仅列表
        title = pos_item.get('positionTitle', '')
        department = map_department(None, pos_item.get('bgs', ''))
        city = map_city(None, pos_item.get('workCities', ''))
        category = ''  # 列表只有 code，无 name
        description = pos_item.get('positionTitle', '')
        requirements = ''
        data_quality = 'summary_only'

    if not description:
        description = title

    # 提取技能关键词
    skill_keywords = extract_keywords(title, category, description, requirements)

    source_url = f'https://join.qq.com/post_detail.html?postid={post_id}'
    external_inner_id = str(pos_item.get('id', ''))

    if detail:
        full_jd_count += 1
    elif post_id in [item.get('postId') for item in details_map.values() if isinstance(item, dict)]:
        summary_only_count += 1
    else:
        missing_detail_count += 1

    items.append({
        'external_id': post_id,
        'title': title,
        'department': department,
        'city': city,
        'category': category,
        'job_type': job_type,
        'data_quality': data_quality,
        'skill_keywords_json': skill_keywords
    })

# 重新统计
full_jd_count = sum(1 for i in items if i['data_quality'] == 'full_jd')
summary_only_count = sum(1 for i in items if i['data_quality'] == 'summary_only')

preview = {
    'source': 'tencent_join',
    'generated_at': generated_at,
    'count': len(items),
    'summary': {
        'full_jd': full_jd_count,
        'summary_only': summary_only_count,
        'missing_detail': summary_only_count
    },
    'items': items
}

ensure_dir = os.path.dirname(output_file)
if ensure_dir and not os.path.exists(ensure_dir):
    os.makedirs(ensure_dir, exist_ok=True)

with open(output_file, 'w', encoding='utf-8') as f:
    json.dump(preview, f, ensure_ascii=False, indent=2)

print(json.dumps({
    'count': len(items),
    'full_jd': full_jd_count,
    'summary_only': summary_only_count
}, ensure_ascii=False))
PYEOF
}

# 生成导入 SQL
python_generate_sql() {
    local positions_file="$1"
    local details_file="$2"
    local dictionary_file="$3"
    local output_file="$4"
    local generated_at="$5"
    local synced_at="$6"

    python3 - "$positions_file" "$details_file" "$dictionary_file" "$output_file" "$generated_at" "$synced_at" << 'PYEOF'
import json, sys, os, re

positions_file = sys.argv[1]
details_file = sys.argv[2]
dictionary_file = sys.argv[3]
output_file = sys.argv[4]
generated_at = sys.argv[5]
synced_at = sys.argv[6]

# 读取 positions
with open(positions_file, 'r', encoding='utf-8') as f:
    pos_data = json.load(f)
positions = pos_data.get('items', [])

# 读取 details
details_map = {}
if os.path.exists(details_file):
    try:
        with open(details_file, 'r', encoding='utf-8') as f:
            det_data = json.load(f)
        for item in det_data.get('items', []):
            post_id = item.get('postId', '')
            if post_id and item.get('ok'):
                details_map[post_id] = item.get('response', {})
    except:
        pass

# 读取 dictionary snapshot
dict_snapshot = {}
if os.path.exists(dictionary_file):
    try:
        with open(dictionary_file, 'r', encoding='utf-8') as f:
            dict_data = json.load(f)
        dict_snapshot['fetched_at'] = dict_data.get('fetched_at', '')
    except:
        pass

# 预置关键词列表
KEYWORDS = [
    'Go', 'Java', 'Python', 'C\\+\\+', '后端', '后台', '算法', '机器学习',
    '深度学习', '云计算', '数据库', 'MySQL', 'Redis', 'Docker', 'Kubernetes',
    '微服务', '前端', 'React', 'Vue', '产品', '运营', '测试', '数据分析',
    '自然语言处理', 'NLP', '多模态', '分布式', 'Linux', '网络', '安全',
    '游戏', '移动端', 'iOS', 'Android', '设计', '嵌入式', '编译器',
    '高性能', '存储', '大数据', 'Spark', 'Flink', 'Hadoop',
    '推荐系统', '搜索', '图像', '计算机视觉', '语音', '大模型', 'LLM',
    'RAG', '向量数据库', '运维', 'SRE', 'DevOps', 'CI/CD'
]

def extract_keywords(title, category, desc, request):
    text = ' '.join(filter(None, [title, category, desc, request]))
    found = []
    for kw_pattern in KEYWORDS:
        if re.search(kw_pattern, text, re.IGNORECASE):
            clean_kw = kw_pattern.replace('\\+', '+').replace('\\', '')
            found.append(clean_kw)
    seen = set()
    result = []
    for kw in found:
        kw_lower = kw.lower()
        if kw_lower not in seen:
            seen.add(kw_lower)
            result.append(kw)
    if not result and category:
        result.append(category)
    if not result and title:
        words = title.split('-')
        if words:
            result.append(words[0].strip())
    return result

def map_city(work_city_list, work_cities_str):
    if work_city_list and isinstance(work_city_list, list):
        return ','.join(work_city_list)
    if work_cities_str:
        return work_cities_str.strip()
    return ''

def map_department(intention_bgd_list, bgs_str):
    if intention_bgd_list and isinstance(intention_bgd_list, list) and len(intention_bgd_list) > 0:
        return intention_bgd_list[0].get('showTitle', '')
    if bgs_str:
        parts = bgs_str.strip().split()
        if parts:
            return parts[0]
    return ''

def escape_sql(s):
    if s is None:
        return ''
    s = str(s)
    s = s.replace('\\', '\\\\')
    s = s.replace("'", "\\'")
    s = s.replace('\n', '\\n')
    s = s.replace('\r', '\\r')
    s = s.replace('\t', '\\t')
    s = s.replace('\0', '')
    return s

# 生成 SQL
lines = []
lines.append('-- ============================================================')
lines.append('-- 腾讯招聘公开岗位导入 SQL')
lines.append(f'-- 生成时间：{generated_at}')
lines.append(f'-- 数据来源：positions.json ({len(positions)} 条) + details.json ({len(details_map)} 条详情)')
lines.append('-- 策略：INSERT ... ON DUPLICATE KEY UPDATE（幂等导入）')
lines.append('-- 唯一业务键：external_source + external_id')
lines.append('-- ============================================================')
lines.append('')

full_jd_count = 0
summary_only_count = 0

for pos_item in positions:
    post_id = pos_item.get('postId', '')
    if not post_id:
        continue

    detail = details_map.get(post_id, None)

    # 字段映射
    title = ''
    department = ''
    city = ''
    category = ''
    job_type = pos_item.get('projectName', '')
    description = ''
    requirements = ''
    data_quality = 'summary_only'
    external_inner_id = str(pos_item.get('id', ''))

    if detail:
        title = detail.get('title', '') or pos_item.get('positionTitle', '')
        department = map_department(detail.get('intentionBGDList', []), pos_item.get('bgs', ''))
        city = map_city(detail.get('workCityList', []), pos_item.get('workCities', ''))
        category = detail.get('tidName', '') or ''
        description = detail.get('desc', '') or pos_item.get('positionTitle', '')
        requirements = detail.get('request', '') or ''
        data_quality = 'full_jd' if (detail.get('desc') or detail.get('request')) else 'summary_only'
    else:
        title = pos_item.get('positionTitle', '')
        department = map_department(None, pos_item.get('bgs', ''))
        city = map_city(None, pos_item.get('workCities', ''))
        category = ''
        description = pos_item.get('positionTitle', '')
        requirements = ''
        data_quality = 'summary_only'

    if not description:
        description = title

    # 提取技能关键词
    skill_keywords = extract_keywords(title, category, description, requirements)
    skill_keywords_json = json.dumps(skill_keywords, ensure_ascii=False)

    source_url = f'https://join.qq.com/post_detail.html?postid={post_id}'

    # 构建 raw_json
    raw_json_obj = {
        'source': 'tencent_join',
        'list_item': pos_item,
        'detail_item': detail if detail else None,
        'dictionary_snapshot': dict_snapshot if dict_snapshot else None,
        'imported_at': generated_at
    }
    raw_json_str = json.dumps(raw_json_obj, ensure_ascii=False)

    if data_quality == 'full_jd':
        full_jd_count += 1
    else:
        summary_only_count += 1

    # 生成 INSERT 语句
    lines.append(f"INSERT INTO `jobs` (`title`, `department`, `city`, `category`, `job_type`, `description`, `requirements`, `skill_keywords_json`, `external_source`, `external_id`, `external_inner_id`, `source_url`, `raw_json`, `data_quality`, `is_featured`, `synced_at`) VALUES")
    lines.append(f"('{escape_sql(title)}', '{escape_sql(department)}', '{escape_sql(city)}', '{escape_sql(category)}', '{escape_sql(job_type)}', '{escape_sql(description)}', '{escape_sql(requirements)}', '{escape_sql(skill_keywords_json)}', 'tencent_join', '{escape_sql(post_id)}', '{escape_sql(external_inner_id)}', '{escape_sql(source_url)}', '{escape_sql(raw_json_str)}', '{data_quality}', 0, '{synced_at}')")
    lines.append("ON DUPLICATE KEY UPDATE")
    lines.append(f"  title = VALUES(title),")
    lines.append(f"  department = VALUES(department),")
    lines.append(f"  city = VALUES(city),")
    lines.append(f"  category = VALUES(category),")
    lines.append(f"  job_type = VALUES(job_type),")
    lines.append(f"  description = VALUES(description),")
    lines.append(f"  requirements = VALUES(requirements),")
    lines.append(f"  skill_keywords_json = VALUES(skill_keywords_json),")
    lines.append(f"  external_inner_id = VALUES(external_inner_id),")
    lines.append(f"  source_url = VALUES(source_url),")
    lines.append(f"  raw_json = VALUES(raw_json),")
    lines.append(f"  data_quality = VALUES(data_quality),")
    lines.append(f"  is_featured = VALUES(is_featured),")
    lines.append(f"  synced_at = VALUES(synced_at),")
    lines.append(f"  updated_at = CURRENT_TIMESTAMP;")
    lines.append('')

# 写入文件
ensure_dir = os.path.dirname(output_file)
if ensure_dir and not os.path.exists(ensure_dir):
    os.makedirs(ensure_dir, exist_ok=True)

with open(output_file, 'w', encoding='utf-8') as f:
    f.write('\n'.join(lines))

print(json.dumps({
    'count': len(positions),
    'full_jd': full_jd_count,
    'summary_only': summary_only_count
}, ensure_ascii=False))
PYEOF
}

# ============================================================
# help
# ============================================================

cmd_help() {
    cat <<EOF
腾讯招聘公开岗位清洗与入库脚本

用法：
  $(basename "$0") <命令>

命令：
  help              查看帮助
  check             检查 raw JSON、MySQL、输出目录、依赖工具
  preview           预览将导入的岗位数量和字段映射，生成 preview JSON
  generate-sql      根据 raw JSON 生成导入 SQL
  generate          generate-sql 的别名
  import            执行 generate-sql，然后导入 MySQL
  status            查看数据库中腾讯岗位数量和数据质量分布
  db-status         status 的别名
  clean-generated   清理生成的 SQL 和 preview JSON
  clean             clean-generated 的别名

Makefile 入口（推荐）：
  make jobs-import-help       查看帮助
  make jobs-import-check      检查环境
  make jobs-import-preview    预览
  make jobs-import-sql        生成 SQL
  make jobs-import            执行导入
  make jobs-import-status     查看状态
  make jobs-db-status         同 jobs-import-status
  make jobs-import-clean      清理生成文件

说明：
  - 从本地 data/tencent-jobs/raw/ 读取原始 JSON
  - 不请求腾讯招聘接口
  - 使用 ON DUPLICATE KEY UPDATE 实现幂等导入
  - 不删除本地 Mock 岗位
  - 不清空 jobs 表
  - 当前只导入 jobs 表，不修改后端接口

字段映射核心：
  title          <- detail.title 优先，list.positionTitle 兜底
  department     <- detail.intentionBGDList[0].showTitle 优先，list.bgs 兜底
  city           <- detail.workCityList 优先，list.workCities 兜底
  category       <- detail.tidName 优先，list.positionFamily 兜底
  job_type       <- list.projectName
  description    <- detail.desc 优先
  requirements   <- detail.request 优先
  external_id    <- postId
  source_url     <- https://join.qq.com/post_detail.html?postid={postId}
  data_quality   <- full_jd（有详情）或 summary_only（无详情）
EOF
}

# ============================================================
# check
# ============================================================

cmd_check() {
    load_db_config

    log_step "腾讯岗位清洗与入库 - 环境检查"

    # 检查 python3
    if command -v python3 &>/dev/null; then
        log_info "python3: 已安装"
    else
        log_error "python3: 未安装（必须，用于 JSON 处理和 SQL 生成）"
        return 1
    fi

    echo ""
    log_step "原始数据文件"

    if [ -f "$POSITIONS_FILE" ]; then
        local pos_count
        pos_count=$(python3 -c "import json; print(json.load(open('${POSITIONS_FILE}')).get('count', 0))" 2>/dev/null || echo "0")
        log_info "positions.json: 存在（${pos_count} 条岗位）"
    else
        log_error "positions.json: 不存在（请先执行 make jobs-fetch）"
    fi

    if [ -f "$DETAILS_FILE" ]; then
        local det_count
        det_count=$(python3 -c "import json; print(json.load(open('${DETAILS_FILE}')).get('count', 0))" 2>/dev/null || echo "0")
        log_info "details.json: 存在（${det_count} 条详情）"
    else
        log_warn "details.json: 不存在（岗位将降级为 summary_only）"
    fi

    if [ -f "$DICTIONARY_FILE" ]; then
        log_info "dictionary.json: 存在"
    else
        log_warn "dictionary.json: 不存在（可选）"
    fi

    echo ""
    log_step "生成目录"
    ensure_dir "$GENERATED_DIR"
    log_info "输出目录: data/tencent-jobs/generated/"

    echo ""
    log_step "MySQL 连接"

    if check_mysql; then
        log_info "MySQL: 可用"
        log_info "数据库: ${DB_NAME}@${DB_HOST}:${DB_PORT}"
        log_info "用户: ${DB_USER}"
        log_info "密码: $(mask_secret "$DB_PASSWORD")"
    else
        log_error "MySQL: 不可用（请先启动 Docker: make infra-up）"
        log_warn "preview 和 generate-sql 不依赖 MySQL，仍可执行"
    fi

    echo ""
    log_step "安全边界"
    log_info "不请求腾讯招聘接口"
    log_info "不修改 raw JSON 文件"
    log_info "不修改 .env"
    log_info "不删除本地 Mock 岗位"
    log_info "使用 ON DUPLICATE KEY UPDATE 幂等导入"
}

# ============================================================
# preview
# ============================================================

cmd_preview() {
    load_db_config

    if [ ! -f "$POSITIONS_FILE" ]; then
        log_error "positions.json 不存在，请先执行 make jobs-fetch"
        return 1
    fi

    ensure_dir "$GENERATED_DIR"

    log_step "预览腾讯岗位清洗结果"

    local generated_at
    generated_at=$(iso_now)

    local result
    result=$(python_preview "$POSITIONS_FILE" "$DETAILS_FILE" "$DICTIONARY_FILE" "$PREVIEW_FILE" "$generated_at")

    if [ $? -ne 0 ]; then
        log_error "预览生成失败"
        return 1
    fi

    local count full_jd summary_only
    count=$(echo "$result" | python3 -c "import json,sys; print(json.load(sys.stdin)['count'])")
    full_jd=$(echo "$result" | python3 -c "import json,sys; print(json.load(sys.stdin)['full_jd'])")
    summary_only=$(echo "$result" | python3 -c "import json,sys; print(json.load(sys.stdin)['summary_only'])")

    log_info "将导入 ${count} 条腾讯岗位"
    log_info "  full_jd（完整JD）: ${full_jd}"
    log_info "  summary_only（仅摘要）: ${summary_only}"
    log_info "预览文件已保存: ${PREVIEW_FILE}"
}

# ============================================================
# generate-sql
# ============================================================

cmd_generate_sql() {
    load_db_config

    if [ ! -f "$POSITIONS_FILE" ]; then
        log_error "positions.json 不存在，请先执行 make jobs-fetch"
        return 1
    fi

    ensure_dir "$GENERATED_DIR"

    log_step "生成腾讯岗位导入 SQL"

    local generated_at synced_at
    generated_at=$(iso_now)
    synced_at=$(mysql_now)

    local result
    result=$(python_generate_sql "$POSITIONS_FILE" "$DETAILS_FILE" "$DICTIONARY_FILE" "$SQL_FILE" "$generated_at" "$synced_at")

    if [ $? -ne 0 ]; then
        log_error "SQL 生成失败"
        return 1
    fi

    local count full_jd summary_only
    count=$(echo "$result" | python3 -c "import json,sys; print(json.load(sys.stdin)['count'])")
    full_jd=$(echo "$result" | python3 -c "import json,sys; print(json.load(sys.stdin)['full_jd'])")
    summary_only=$(echo "$result" | python3 -c "import json,sys; print(json.load(sys.stdin)['summary_only'])")

    local sql_size
    sql_size=$(wc -c < "$SQL_FILE" | tr -d ' ')
    local sql_lines
    sql_lines=$(wc -l < "$SQL_FILE" | tr -d ' ')

    log_info "SQL 文件已生成: ${SQL_FILE}"
    log_info "  岗位数量: ${count}（full_jd: ${full_jd}, summary_only: ${summary_only}）"
    log_info "  文件大小: ${sql_size} bytes, ${sql_lines} lines"
}

# ============================================================
# import
# ============================================================

cmd_import() {
    load_db_config

    # 检查 raw JSON
    if [ ! -f "$POSITIONS_FILE" ]; then
        log_error "positions.json 不存在，请先执行 make jobs-fetch"
        return 1
    fi

    # 检查 MySQL
    if ! check_mysql; then
        log_error "MySQL 不可用，无法导入"
        return 1
    fi

    # 生成 SQL
    cmd_generate_sql

    if [ ! -f "$SQL_FILE" ]; then
        log_error "SQL 文件生成失败"
        return 1
    fi

    log_step "导入腾讯岗位到 MySQL"

    # 导入 SQL
    docker exec -i offer-catcher-mysql mysql \
        -h"${DB_HOST}" -P"${DB_PORT}" \
        -u"${DB_USER}" -p"${DB_PASSWORD}" \
        "${DB_NAME}" < "$SQL_FILE" 2>/dev/null

    if [ $? -ne 0 ]; then
        log_error "SQL 导入失败"
        return 1
    fi

    log_info "SQL 导入完成"

    # 验证导入结果
    log_step "导入结果验证"

    local tencent_count
    tencent_count=$(docker exec offer-catcher-mysql mysql \
        -h"${DB_HOST}" -P"${DB_PORT}" \
        -u"${DB_USER}" -p"${DB_PASSWORD}" \
        "${DB_NAME}" -N -e "SELECT COUNT(*) FROM jobs WHERE external_source = 'tencent_join'" 2>/dev/null || echo "0")
    tencent_count=$(echo "$tencent_count" | tr -d ' \n')

    local local_count
    local_count=$(docker exec offer-catcher-mysql mysql \
        -h"${DB_HOST}" -P"${DB_PORT}" \
        -u"${DB_USER}" -p"${DB_PASSWORD}" \
        "${DB_NAME}" -N -e "SELECT COUNT(*) FROM jobs WHERE external_source = 'local_mock'" 2>/dev/null || echo "0")
    local_count=$(echo "$local_count" | tr -d ' \n')

    local invalid_json
    invalid_json=$(docker exec offer-catcher-mysql mysql \
        -h"${DB_HOST}" -P"${DB_PORT}" \
        -u"${DB_USER}" -p"${DB_PASSWORD}" \
        "${DB_NAME}" -N -e "SELECT COUNT(*) FROM jobs WHERE external_source = 'tencent_join' AND JSON_VALID(raw_json) = 0" 2>/dev/null || echo "0")
    invalid_json=$(echo "$invalid_json" | tr -d ' \n')

    log_info "tencent_join 岗位: ${tencent_count} 条"
    log_info "local_mock 岗位: ${local_count} 条"
    log_info "无效 raw_json: ${invalid_json} 条"

    if [ "$tencent_count" -eq 0 ]; then
        log_warn "未导入任何腾讯岗位，请检查 raw JSON 数据"
    fi

    if [ "$local_count" -eq 0 ]; then
        log_error "本地 Mock 岗位丢失！请检查数据库"
    fi

    if [ "$invalid_json" -gt 0 ]; then
        log_error "存在 ${invalid_json} 条无效 raw_json，请检查"
    fi
}

# ============================================================
# status
# ============================================================

cmd_status() {
    load_db_config

    log_step "数据库岗位状态"

    if ! check_mysql; then
        log_error "MySQL 不可用"
        return 1
    fi

    echo ""

    # 按 external_source 和 data_quality 分组统计
    docker exec offer-catcher-mysql mysql \
        -h"${DB_HOST}" -P"${DB_PORT}" \
        -u"${DB_USER}" -p"${DB_PASSWORD}" \
        "${DB_NAME}" -e "
SELECT external_source, data_quality, COUNT(*) AS count
FROM jobs
GROUP BY external_source, data_quality
ORDER BY external_source, data_quality;
" 2>/dev/null

    echo ""

    # tencent_join 总数
    local tencent_count
    tencent_count=$(docker exec offer-catcher-mysql mysql \
        -h"${DB_HOST}" -P"${DB_PORT}" \
        -u"${DB_USER}" -p"${DB_PASSWORD}" \
        "${DB_NAME}" -N -e "SELECT COUNT(*) FROM jobs WHERE external_source = 'tencent_join'" 2>/dev/null || echo "0")
    tencent_count=$(echo "$tencent_count" | tr -d ' \n')

    # local_mock 总数
    local local_count
    local_count=$(docker exec offer-catcher-mysql mysql \
        -h"${DB_HOST}" -P"${DB_PORT}" \
        -u"${DB_USER}" -p"${DB_PASSWORD}" \
        "${DB_NAME}" -N -e "SELECT COUNT(*) FROM jobs WHERE external_source = 'local_mock'" 2>/dev/null || echo "0")
    local_count=$(echo "$local_count" | tr -d ' \n')

    # 无效 raw_json
    local invalid_json
    invalid_json=$(docker exec offer-catcher-mysql mysql \
        -h"${DB_HOST}" -P"${DB_PORT}" \
        -u"${DB_USER}" -p"${DB_PASSWORD}" \
        "${DB_NAME}" -N -e "SELECT COUNT(*) FROM jobs WHERE external_source = 'tencent_join' AND JSON_VALID(raw_json) = 0" 2>/dev/null || echo "0")
    invalid_json=$(echo "$invalid_json" | tr -d ' \n')

    log_info "tencent_join 岗位: ${tencent_count} 条"
    log_info "local_mock 岗位: ${local_count} 条"
    log_info "invalid_raw_json: ${invalid_json} 条"
}

# ============================================================
# clean-generated
# ============================================================

cmd_clean_generated() {
    log_step "清理生成的临时文件"

    local cleaned=0

    if [ -f "$SQL_FILE" ]; then
        rm -f "$SQL_FILE"
        log_info "已删除: tencent_jobs_import.sql"
        cleaned=$((cleaned + 1))
    fi

    if [ -f "$PREVIEW_FILE" ]; then
        rm -f "$PREVIEW_FILE"
        log_info "已删除: tencent_jobs_preview.json"
        cleaned=$((cleaned + 1))
    fi

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
        check)
            cmd_check
            ;;
        preview)
            cmd_preview
            ;;
        generate-sql|generate)
            cmd_generate_sql
            ;;
        import)
            cmd_import
            ;;
        status|db-status)
            cmd_status
            ;;
        clean-generated|clean)
            cmd_clean_generated
            ;;
        *)
            log_error "未知命令：${cmd}"
            cmd_help
            exit 1
            ;;
    esac
}

main "$@"
