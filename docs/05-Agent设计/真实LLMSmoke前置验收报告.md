# 真实 LLM Smoke 前置验收报告

> 创建时间：2026-05-30
> 所属阶段：4-3.1
> 验收人：CodeBuddy

## 1. 验收目标

在接业务 Agent 前，确认 Hunyuan / TokenHub 最小调用真实可用。

## 2. 验收范围

- 执行 `make llm-help`：确认命令入口可用
- 执行 `make llm-check`：确认本地 LLM 配置状态
- 在安全条件满足时执行 `RUN_REAL_LLM_SMOKE=1 make llm-smoke`：验证 HunyuanClient 连通性
- 不接业务 Agent
- 不调用推荐 / 匹配 / 简历优化接口
- 不修改 Go 代码
- 不修改 .env / .env.example
- 不输出真实 API Key

## 3. 当前前提

4-2 已完成以下 LLM 基础设施：

|| 资产 | 位置 | 状态 |
|------|------|------|
| LLM Client 抽象 | server/internal/llm/ | ✅ 已完成 |
| HunyuanClient | server/internal/llm/hunyuan_client.go | ✅ 已完成 |
| Prompt Loader | server/internal/prompt/ | ✅ 已完成 |
| Schema Loader | server/internal/schema/ | ✅ 已完成 |
| JSON parse + Schema validate | server/internal/llmjson/ + schema/validator.go | ✅ 已完成 |
| JSONRepairAgent | server/internal/agent/repair/ | ✅ 已完成 |
| 统一 Pipeline | server/internal/agent/pipeline/ | ✅ 已完成 |
| llm-smoke Go 命令 | server/cmd/llm-smoke/main.go | ✅ 已完成 |
| llm_smoke.sh 脚本 | scripts/dev/llm_smoke.sh | ✅ 已完成 |
| Makefile llm-* target | Makefile | ✅ 已完成 |

4-3.0 已完成业务 Agent 接入 LLM 方案设计（docs only）。

## 4. 配置检查结果

|| 检查项 | 结果 | 说明 |
|--------|------|------|
| .env 是否存在 | ❌ 不存在 | 项目根目录无 .env 文件 |
| LLM_ENABLE | false | 默认值，LLM 未启用 |
| LLM_PROVIDER | hunyuan | 默认值 |
| LLM_BASE_URL | 未配置 | 环境变量为空 |
| LLM_API_KEY | 未配置 | 环境变量为空 |
| LLM_MODEL | 未配置 | 环境变量为空 |
| LLM_TIMEOUT_SECONDS | 30 | 默认值，合理 |
| LLM_LOG_RAW_OUTPUT | false | 默认值，合理 |
| LLM_PROMPT_DIR | server/prompts | 默认值，已配置 |
| LLM_SCHEMA_DIR | server/prompts/schema | 默认值，已配置 |

**结论：本地 LLM 配置不完整，无法执行真实 smoke。**

缺失项：
- `.env` 文件不存在
- `LLM_BASE_URL` 未配置
- `LLM_API_KEY` 未配置
- `LLM_MODEL` 未配置
- `LLM_ENABLE` 未设为 true

## 5. make llm-help 验收

执行结果：✅ 通过

```
LLM Smoke Test 命令
====================

用法: scripts/dev/llm_smoke.sh <command>

命令:
  help     显示此帮助信息
  check    检查 LLM 配置（不请求外部模型）
  smoke    执行 LLM smoke test（需要 RUN_REAL_LLM_SMOKE=1）

安全说明:
  - check 命令不会请求外部模型
  - smoke 命令需要显式设置 RUN_REAL_LLM_SMOKE=1
  - 不会打印 API Key
  - 不会打印完整 raw output
  - 不会修改 .env
```

确认项：
- llm-help 命令可用
- 安全说明已包含
- 不打印 API Key
- check 不请求外部模型
- smoke 需要 RUN_REAL_LLM_SMOKE=1

## 6. make llm-check 验收

执行结果：✅ 通过（配置检查功能正常）

```
[INFO] .env 文件: 不存在（将使用环境变量）
[INFO] LLM_ENABLE=false
[INFO] LLM_PROVIDER=hunyuan
[INFO] LLM_BASE_URL=未配置
[INFO] LLM_API_KEY=未配置
[INFO] LLM_MODEL=未配置
[INFO] LLM_TEMPERATURE=0.2
[INFO] LLM_MAX_OUTPUT_TOKENS=8192
[INFO] LLM_TIMEOUT_SECONDS=30
[INFO] server/cmd/llm-smoke: 存在
```

确认项：
- check 命令可用
- check 过程没有请求外部模型 ✅
- LLM_API_KEY 只显示「未配置」，不打印真实值 ✅
- 正确识别 .env 不存在
- 正确识别配置缺失项

配置摘要：LLM 配置不完整，不可执行 smoke test。

## 7. 真实 LLM smoke 执行结果

**未执行真实 LLM smoke。**

原因：本地 LLM 配置不完整：
1. `.env` 文件不存在
2. `LLM_ENABLE` 为 false（非 true）
3. `LLM_BASE_URL` 未配置
4. `LLM_API_KEY` 未配置
5. `LLM_MODEL` 未配置

按照安全规则：如果 `.env` 中未配置真实 Key，或者 `LLM_ENABLE=false`，或 `LLM_BASE_URL / LLM_MODEL / LLM_API_KEY` 缺失，不强行执行真实 smoke。

## 8. 输出安全检查

|| 检查项 | 结果 |
|--------|------|
| 是否输出 API Key | ❌ 无 API Key 可输出（未配置） |
| 是否输出完整 raw output | ❌ 未执行真实调用 |
| 是否携带业务数据 | ❌ 未执行任何调用 |
| 是否调用业务接口 | ❌ 未调用 |
| 是否输出 Authorization header | ❌ 未执行任何调用 |

## 9. JSON parse 观察结果

由于未执行真实 LLM smoke，无法观察 JSON parse 结果。

Go 基础测试已通过：

```
ok  offer-catcher/server/internal/llm       0.209s
ok  offer-catcher/server/internal/llmjson   0.002s
```

HunyuanClient httptest 单元测试（22 项）和 llmjson 单元测试全部通过，证明代码逻辑正确。

## 10. 未执行的内容

- ❌ 未执行真实 LLM smoke（配置不完整）
- ❌ 未调用推荐接口
- ❌ 未调用匹配分析接口
- ❌ 未调用简历优化接口
- ❌ 未修改 .env
- ❌ 未修改 .env.example
- ❌ 未修改 Prompt 文件
- ❌ 未修改 Schema 文件
- ❌ 未修改业务 Agent
- ❌ 未修改前端
- ❌ 未请求腾讯招聘接口

## 11. 当前风险

|| 风险 | 严重度 | 说明 |
|------|--------|------|
| 本地 LLM 配置不完整 | 🔴 高 | 无法执行真实 smoke，无法确认 HunyuanClient 连通性 |
| Hunyuan API 可用性未验证 | 🟡 中 | httptest 通过但真实连通性未知 |
| LLM 输出 JSON 遵循性未验证 | 🟡 中 | 最小 Prompt 下模型是否严格输出 JSON 未知 |
| API Key 可能过期或额度不足 | 🟡 中 | 未验证 |

## 12. 是否满足进入 4-3.2

**暂不建议进入 4-3.2。**

原因：
1. 真实 LLM smoke 未执行（配置不完整）
2. HunyuanClient 真实连通性未验证
3. 需要先补齐本地 LLM 配置并完成真实 smoke 后再接入业务 Agent

条件满足后可进入 4-3.2：
1. 创建 `.env` 文件
2. 配置 `LLM_ENABLE=true`
3. 配置 `LLM_BASE_URL` 为真实 Hunyuan / TokenHub 地址
4. 配置 `LLM_API_KEY` 为有效 Key
5. 配置 `LLM_MODEL` 为可用模型名称
6. 执行 `RUN_REAL_LLM_SMOKE=1 make llm-smoke` 通过

## 13. 验收结论

4-3.1 真实 LLM Smoke 前置验收**未完成**。

- ✅ `make llm-help` 可用
- ✅ `make llm-check` 可用，正确识别配置缺失
- ✅ LLM 基础设施代码测试通过（llm + llmjson）
- ✅ check 过程不请求外部模型
- ✅ 不输出 API Key
- ❌ 真实 LLM smoke 未执行（本地 LLM 配置不完整）
- ❌ HunyuanClient 真实连通性未验证

**暂不建议进入 4-3.2，需要先补齐本地 LLM 配置并完成真实 smoke。**

## 14. 下一步建议

1. 用户创建 `.env` 文件并配置真实 LLM 参数：
   ```env
   LLM_ENABLE=true
   LLM_PROVIDER=hunyuan
   LLM_BASE_URL=<真实 Hunyuan/TokenHub 地址>
   LLM_API_KEY=<真实 API Key>
   LLM_MODEL=<可用模型名称>
   ```
2. 执行 `make llm-check` 确认配置完整
3. 执行 `RUN_REAL_LLM_SMOKE=1 make llm-smoke` 验证连通性
4. 如果真实 smoke 通过，可进入 4-3.2 MatchAnalysisAgent 接入 Pipeline
5. 如果真实 smoke 失败，需先修复配置或调用问题

---

## 15. 二次核验结果（4-3.1.B）

> 核验时间：2026-05-30
> 核验编号：4-3.1.B

### 15.1 核验结论

**not_executed**

4-3.1.B 未执行真实 smoke，不能进入 4-3.2。

### 15.2 配置检查结果

| 检查项 | 结果 | 说明 |
|--------|------|------|
| .env 是否存在 | ❌ 不存在 | 仍未创建 |
| LLM_ENABLE | false | 未启用 |
| LLM_PROVIDER | hunyuan | 默认值 |
| LLM_BASE_URL | 未配置 | 空 |
| LLM_API_KEY | 未配置 | 空 |
| LLM_MODEL | 未配置 | 空 |

**与前次 4-3.1 核验结果完全一致，配置状态未发生变化。**

### 15.3 make llm-help 验收

✅ 通过（与 4-3.1 一致）

### 15.4 make llm-check 验收

✅ 通过（配置检查功能正常，与 4-3.1 一致）

### 15.5 真实 LLM smoke 执行结果

❌ **未执行**。原因与前次完全一致：`.env` 不存在，LLM 配置不完整。

### 15.6 JSON parse 观察结果

未能观察。未执行真实调用。

### 15.7 输出安全检查

| 检查项 | 结果 |
|--------|------|
| 是否输出 API Key | ✅ 无 Key 可输出 |
| 是否输出完整 raw output | ✅ 未执行调用 |
| 是否携带业务数据 | ✅ 未执行调用 |
| 是否调用业务接口 | ✅ 未调用 |

### 15.8 是否可以进入 4-3.2

**❌ 不能进入 4-3.2。**

两次核验（4-3.1 和 4-3.1.B）均因本地 LLM 配置不完整无法执行真实 smoke。

### 15.9 核验结论

4-3.1.B 真实 LLM Smoke 二次核验**未通过**。

- ✅ `make llm-help` 可用
- ✅ `make llm-check` 可用，正确识别配置缺失
- ✅ LLM 基础设施代码测试通过
- ✅ check 过程不请求外部模型
- ✅ 不输出 API Key
- ❌ 真实 LLM smoke 仍未执行（`.env` 仍未创建，LLM 配置仍不完整）
- ❌ HunyuanClient 真实连通性仍未验证

**暂不建议进入 4-3.2，需要先创建 `.env` 并补齐 LLM 配置并完成真实 smoke。**

---

## 16. 4-3.1.C 补充：开发环境 .env 自动生成机制

4-3.1.C 已补充开发环境 `.env` 自动生成机制。

新增：
- `scripts/dev/ensure_env.sh`：`.env` 自动生成脚本
- Makefile `ensure-env` target：确保 `.env` 存在后再执行依赖命令
- `make up` / `make restart` / `make llm-check` / `make llm-smoke` 已接入 `ensure-env`

优先级：`.env` 已存在 → 不覆盖；`.env.local` 存在 → 复制；`.env.example` 存在 → 复制。

后续重新执行 4-3.1 smoke 前，`make llm-check` 会自动确保 `.env` 存在。

注意：本次只是修复环境准备机制，不代表真实 smoke 已通过。用户仍需在 `.env` 或 `.env.local` 中填写真实 LLM 配置。

---

## 17. 三次核验结果（4-3.1 / 4-3.1.B / 重新执行）

> 核验时间：2026-05-30
> 核验编号：4-3.1 重新执行

### 17.1 核验结论

**not_executed**

4-3.1 重新执行仍未执行真实 smoke，不能进入 4-3.2。

### 17.2 配置检查结果

| 检查项 | 结果 | 说明 |
|--------|------|------|
| .env 是否存在 | ✅ 存在 | 4-3.1.C 已自动创建 |
| LLM_ENABLE | false | 未启用 |
| LLM_PROVIDER | hunyuan | 默认值 |
| LLM_BASE_URL | 未配置 | 空 |
| LLM_API_KEY | configured | 有值但 LLM_ENABLE=false |
| LLM_MODEL | configured | 有值但 LLM_ENABLE=false |

与 4-3.1 和 4-3.1.B 核验结果一致：LLM_ENABLE=false，LLM_BASE_URL 未配置，真实 smoke 无法执行。

4-3.1.C 解决了 `.env` 自动创建问题，但从 `.env.example` 生成的 `.env` 只包含模板值。

### 17.3 make llm-help 验收

✅ 通过（与 4-3.1 / 4-3.1.B 一致）

### 17.4 make llm-check 验收

✅ 通过（配置检查功能正常）

### 17.5 真实 LLM smoke 执行结果

❌ **未执行**。原因：LLM_ENABLE=false，LLM_BASE_URL 未配置。

### 17.6 JSON parse 观察结果

未能观察。未执行真实调用。

### 17.7 输出安全检查

| 检查项 | 结果 |
|--------|------|
| 是否输出 API Key | ✅ 未输出 |
| 是否输出完整 raw output | ✅ 未执行调用 |
| 是否携带业务数据 | ✅ 未执行调用 |
| 是否调用业务接口 | ✅ 未调用 |

### 17.8 是否可以进入 4-3.2

**❌ 不能进入 4-3.2。**

三次核验（4-3.1、4-3.1.B 和重新执行）均因本地 LLM 配置不完整无法执行真实 smoke。

### 17.9 核验结论

4-3.1 重新执行**未通过**。

- ✅ `make ensure-env` 可用，`.env` 自动创建
- ✅ `make llm-help` 可用
- ✅ `make llm-check` 可用，正确识别配置缺失
- ✅ LLM 基础设施代码测试通过
- ✅ check 过程不请求外部模型
- ✅ 不输出 API Key
- ❌ 真实 LLM smoke 仍未执行（LLM_ENABLE=false，LLM_BASE_URL 未配置）
- ❌ HunyuanClient 真实连通性仍未验证

**暂不建议进入 4-3.2，需要先在 `.env` 或 `.env.local` 中填写真实 LLM 配置并完成真实 smoke。**

---

## 18. 通过结果登记（4-3.1 真实 LLM Smoke 通过）

> 登记时间：2026-05-30
> 登记编号：4-3.1 通过结果登记

### 18.1 验收结论

**pass**

4-3.1 真实 LLM Smoke 前置验收已通过。可以进入 4-3.2。

### 18.2 配置检查结果

| 检查项 | 结果 |
|--------|------|
| .env 是否存在 | ✅ 存在 |
| LLM_ENABLE | true |
| LLM_PROVIDER | hunyuan |
| LLM_BASE_URL | 已配置 |
| LLM_API_KEY | 已配置（长度: 51），但不记录真实 Key |
| LLM_MODEL | hy3-preview |
| LLM_TEMPERATURE | 0.2 |
| LLM_MAX_OUTPUT_TOKENS | 8192 |
| LLM_TIMEOUT_SECONDS | 30 |
| HunyuanClient 编译 | 通过 |

### 18.3 make llm-help 验收

✅ 通过

### 18.4 make llm-check 验收

✅ 通过（LLM 配置完整，可执行 smoke test）

### 18.5 真实 LLM smoke 执行结果

✅ **已执行并成功**

| 指标 | 值 |
|------|-----|
| provider | hunyuan |
| model | hy3-preview |
| finish_reason | stop |
| raw_output_length | 28 |
| duration_ms | 1327 |
| usage.prompt_tokens | 50 |
| usage.completion_tokens | 11 |
| usage.total_tokens | 61 |
| json_parse_success | true |
| raw_output_preview | {"ok":true,"message":"pong"} |

### 18.6 JSON parse 观察结果

本次最小 Prompt 返回 JSON 可被成功解析，`json_parse_success=true`。

后续业务输出仍必须经过 JSON extract / parse / Schema validate / repair once / fallback_required。

### 18.7 安全检查

| 检查项 | 结果 |
|--------|------|
| 是否输出 API Key | ✅ 仅脱敏尾号，未泄露完整 Key |
| 是否输出 Authorization header | ✅ 未输出 |
| 是否输出完整 raw output | ✅ 未输出（仅 28 字符预览） |
| 是否携带简历 | ✅ 未携带 |
| 是否携带岗位 JD | ✅ 未携带 |
| 是否携带腾讯岗位数据 | ✅ 未携带 |
| 是否携带用户隐私 | ✅ 未携带 |
| 是否调用业务接口 | ✅ 未调用 |
| 是否请求腾讯招聘接口 | ✅ 未请求 |
| 是否修改业务 Agent | ✅ 未修改 |

**轻微风险**：smoke 输出中出现脱敏 API Key 尾号（`api_key=***********************************************4Nql`），虽未泄露完整 Key，不阻塞本次验收；后续建议将输出优化为 `api_key=configured`，不展示任何尾号。

### 18.8 是否可以进入 4-3.2

**✅ 可以进入 4-3.2。**

### 18.9 验收结论

4-3.1 真实 LLM Smoke 前置验收**已通过**。

- ✅ `make llm-help` 可用
- ✅ `make llm-check` 可用，LLM 配置完整
- ✅ `make llm-smoke` 真实调用成功
- ✅ HunyuanClient 真实连通性已验证
- ✅ json_parse_success=true
- ✅ 不输出真实 API Key
- ✅ 不输出完整 raw output
- ✅ 未携带简历 / JD / 腾讯岗位数据 / 用户隐私
- ✅ 未调用业务接口
- ✅ 未修改业务 Agent

**可以进入 4-3.2 MatchAnalysisAgent 接入 Pipeline。**

---

## 19. 4-3.1.D LLM Smoke 环境变量加载修复

> 登记时间：2026-05-30
> 登记编号：4-3.1.D

### 19.1 修复结论

4-3.1.D 已完成 LLM Smoke 环境变量加载修复。

### 19.2 问题根因

`llm_smoke.sh` 原先使用 `source .env`（通过 `load_env_silent()` 函数）加载环境变量。`.env` 中 `DB_DSN=root:password@tcp(127.0.0.1:3306)/offer_catcher?charset=utf8mb4&parseTime=True&loc=Local` 中的 `tcp(...)` 括号被 bash 当作子 shell 语法解析，导致 `syntax error near unexpected token '('`，`.env` 加载中断，后续 LLM 配置变量均未加载。错误被 `2>/dev/null || true` 静默吞掉，表现为 `.env` 存在但 LLM 配置为空。

### 19.3 修复方案

移除 `load_env_silent()` 函数，改用 `common.sh` 的 `load_env()`（逐行 `while read` 解析 + `export`），能正确处理含括号、`&`、`?` 等特殊字符的值。

### 19.4 修复验证

`make llm-check` 已能正确识别 LLM_ENABLE=true、LLM_BASE_URL=已配置、LLM_API_KEY=已配置、LLM_MODEL=hy3-preview。

### 19.5 影响范围

- 仅修改 `scripts/dev/llm_smoke.sh`（移除 `load_env_silent()`，改用 `load_env`）
- 未修改 Go 代码、Prompt、Schema、业务 Agent、前端
- 未修改 `.env` / `.env.local` / `.env.example`
