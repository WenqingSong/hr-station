# Dev-Reboot-Fix-B：修复开发环境重启后回到 mock 的启动问题

日期：2026-06-11

## 背景

`Dev-Reboot-Audit-A` 只读审计确认：
- 后端/前端/MySQL/Redis 均正常启动
- Agent 结果全是 mock（`generated_by=mock_agent` / `mock_quality`）
- 根因三层叠加：`ENABLE_MOCK_AGENT=true` + `LLM_FALLBACK_ENABLE=true` + `job_profiles` 表缺失

## 修复内容

### 1. job_profiles 表补齐

- `server/sql/init.sql`：新增 `job_profiles` 表 DDL（`DROP TABLE IF EXISTS` + `CREATE TABLE`）
  - 字段：id, job_id, direction, category, primary_tags_json, skill_tags_json, requirement_tags_json, responsibility_tags_json, risk_tags_json, evidence_json, source, version, created_at, updated_at
  - 主键：id (AUTO_INCREMENT)，唯一索引：uk_job_id (job_id)
- 与 Go model/repository 完全一致
- 无需手写 seed：启动时 `TagEvidenceService.EnsureAllOpenJobTagProfiles()` 自动 backfill 46 条

### 2. .env.example 默认值修复

| 配置项 | 修复前 | 修复后 |
|---|---|---|
| ENABLE_MOCK_AGENT | true | false |
| LLM_FALLBACK_ENABLE | true | false |
| LLM_PROVIDER | deepseek | hy3 |
| LLM_MODEL | deepseek-v4-flash-202605 | hy3-preview |
| LLM_BASE_URL | tokenhub | lkeap |
| LLM_TIMEOUT_SECONDS | 120 | 180 |

### 3. .env 非密钥配置修复

同 .env.example，将 `ENABLE_MOCK_AGENT=false`、`LLM_FALLBACK_ENABLE=false`、`LLM_PROVIDER=hy3`。

### 4. make check 增强

- `scripts/dev/mysql.sh`：`CORE_TABLES` 增加 `job_profiles` 和 `resume_profiles`
- 现在 `make check` / `make db-check` 会检查 8 张表，缺失时显式报错

## 验证结果

### Fast Match ✅
```
generated_by: fast_tag_engine  ← 不再是 mock_agent
```

### Deep Match ✅
```
DeepMatch agent1_context matched=3 missing=10 baseline=51
DeepMatch agent2_decision reasoning_effort=high
LLM-Hy3 model=hy3-preview reason_effort=high body_bytes=6005
DeepMatch agent3_verifier reasoning_effort=low
LLM-Hy3 model=hy3-preview reason_effort=low body_bytes=11384
DeepMatch stage=agent3_verifier verified=true
generated_by: llm_agent  ← 不再是 mock_quality
```

### 数据库
```
job_profiles: 46 row(s)
resume_profiles: 16 row(s)
jobs_count: 46
resumes_count: 8
所有 8 张核心表通过 make check
```

### 构建
- go build ✅
- go vet ✅
- go test pipeline ✅

## 未修改
- 前端 UI
- 生产配置
- mock 代码文件（保留但不再执行）

## 结论

`Dev-Reboot-Fix-B` 完成。Fast 使用 `fast_tag_engine`，Deep 使用真实 Hy3 3-agent 调用（Agent2 high + Agent3 low），模型不可用时返回明确错误而非 mock。
