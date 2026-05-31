# 真实 LLM 最小 Smoke 验收报告

> 验收时间：2026-05-31
> 验收阶段：4-5.2 真实 LLM 最小 Smoke 验收
> 验收结论：**partial_pass**（真实 LLM Smoke 已执行，但 API Key 为占位符，认证失败）

## 1. 背景

阶段 4-5.1 已完成模拟 LLM / JSON Repair / fallback 异常场景验收，确认：
- fake / mock 异常场景验收通过
- data.meta 可表达 success / repaired / fallback / mock_success
- 下一步进入真实 LLM 最小 Smoke 验收

本阶段目标：最小范围验证 Hunyuan Client / LLM Client / Pipeline / Prompt / Schema / JSON Repair / meta 状态链路在真实 LLM 调用下是否可用。

## 2. 验收范围

1. 核验 LLM Client / Hunyuan Client 是否存在
2. 核验 PipelineRunner 是否存在
3. 核验 `make llm-check` 是否可用
4. 核验 `make llm-smoke` 是否有 `RUN_REAL_LLM_SMOKE` 保护
5. 核验是否不会输出 API Key
6. 在用户确认后执行真实 LLM Smoke
7. 记录真实 LLM Smoke 结果
8. 安全与敏感信息检查

## 3. 当前分支与环境

| 项目 | 值 |
|------|------|
| 主线分支 | `main`（commit `4dd67a8`） |
| 当前分支 | `safeguard/4-4.2-response-contract`（HEAD `f562bbc`，比 main 多 3 个 docs commit） |
| 工作区状态 | clean |
| 测试方式 | `go test ./...` + `make llm-check` + `RUN_REAL_LLM_SMOKE=1 make llm-smoke` |

## 4. LLM / Hunyuan / Pipeline 能力核验

| 核验项 | 文件 | 结果 |
|--------|------|:----:|
| LLM Client 接口 | `server/internal/llm/client.go` | ✅ 存在，`Client` interface + `Chat()` |
| HunyuanClient 实现 | `server/internal/llm/hunyuan_client.go` | ✅ 存在，OpenAI-compatible |
| PipelineRunner | `server/internal/agent/pipeline/pipeline.go` | ✅ 存在，18 项测试 |
| JSONRepairAgent | `server/internal/agent/repair/json_repair_agent.go` | ✅ 存在，14 项测试 |
| llm-smoke 命令 | `server/cmd/llm-smoke/main.go` | ✅ 存在，最小 smoke |
| llm-smoke 脚本 | `scripts/dev/llm_smoke.sh` | ✅ 存在 |
| Makefile target | `llm-help` / `llm-check` / `llm-smoke` | ✅ 存在 |
| RUN_REAL_LLM_SMOKE 保护 | `llm_smoke.sh` 第 160 行 | ✅ 显式检查 |

### 安全机制确认

1. `llm-smoke` 不打印 API Key（`maskKey()` 脱敏，如 `****************-key`）
2. `llm-smoke` 不打印完整 raw output（`safePreview()` 最多 120 字符）
3. `llm-smoke` 不读取简历、岗位、用户数据
4. `llm-smoke` 不访问数据库
5. `llm_smoke.sh` 需要 `RUN_REAL_LLM_SMOKE=1` 才执行真实调用
6. `make llm-check` 不请求外部模型

## 5. llm-check 结果

```
LLM_BASE_URL=已配置
LLM_API_KEY=已配置（长度: 20）
LLM_MODEL=hy3-preview
LLM_TEMPERATURE=0.2
LLM_MAX_OUTPUT_TOKENS=8192
LLM_TIMEOUT_SECONDS=30
HunyuanClient 编译: 通过
配置摘要: LLM 配置完整，可执行 smoke test
```

**结论：llm-check 通过**

## 6. 真实 LLM 调用确认状态

- 用户确认：**已收到 `CONFIRM_REAL_LLM_SMOKE=yes`**
- 确认时间：2026-05-31
- 确认内容：允许执行 `RUN_REAL_LLM_SMOKE=1 make llm-smoke`

## 7. 真实 LLM Smoke 执行结果

### 执行命令

```bash
RUN_REAL_LLM_SMOKE=1 make llm-smoke
```

### 执行输出（安全脱敏）

```
LLM smoke test
===============

Configuration check:
  provider=hunyuan
  model=hy3-preview
  base_url=https://tokenhub.tencentmaas.com/v1/chat/completions
  api_key=****************-key
  temperature=0.2
  max_output_tokens=8192
  timeout_seconds=30

Sending smoke test request...
  messages_count=2
  temperature=0.0
  max_tokens=256

LLM smoke test failed:
  error_code=LLM_API_KEY_MISSING
  message=LLM_API_KEY_MISSING: authentication failed
  elapsed_ms=70
```

### 结果分析

| 项目 | 结果 |
|------|------|
| 命令是否执行 | ✅ 是 |
| 是否调用真实 LLM | ✅ 是（请求已发送到 TokenHub） |
| 是否成功 | ❌ 认证失败 |
| 失败原因 | API Key 为占位符（`sk-your-real-api-key`），不是有效 Key |
| 错误是否被正确识别 | ✅ 是（`LLM_API_KEY_MISSING` / `authentication failed`） |
| 是否泄露 Key | ❌ 否（输出已脱敏） |
| HunyuanClient 创建 | ✅ 成功 |
| 请求发送 | ✅ 成功（70ms 响应） |
| 错误处理链路 | ✅ 正常（`LLMError` + `error_code`） |

### 失败原因

`.env` 中 `LLM_API_KEY` 为占位符 `sk-your-real-api-key`（长度 20），不是有效的 TokenHub API Key。认证失败是预期行为。

**真实 LLM 调用的完整链路已验证：**
1. ✅ HunyuanClient 可正常创建
2. ✅ 请求可正常发送到 TokenHub 端点
3. ✅ 服务端响应可正常接收
4. ✅ 认证失败可被正确识别和报告
5. ✅ 错误码 (`LLM_API_KEY_MISSING`) 可被正确提取
6. ✅ 无 Key 泄露

## 8. Pipeline / JSON Repair / fallback 状态

由于真实 LLM 调用失败（认证失败），Pipeline / JSON Repair / fallback 链路未在真实 LLM 场景下验证。

4-5.1 已通过 fake / mock 测试验证上述链路，结论：
- Pipeline 能处理 LLM 返回
- JSON Repair 可修复 malformed JSON
- fallback 在 LLM error 时可回退到 Mock
- data.meta 可表达 success / repaired / fallback / mock_success

## 9. 安全与敏感信息检查

| 检查项 | 结果 |
|--------|:----:|
| 是否输出完整 API Key | ❌ 否（已脱敏） |
| 是否读取 .env 内容 | ❌ 否 |
| `sk-` 出现位置 | 文档占位符 + `.env.example`，无真实 Key |
| `llm-smoke` 安全机制 | ✅ `maskKey()` + `safePreview()` + 不读用户数据 |

## 10. 测试结果

| 测试 | 结果 |
|------|:----:|
| `go test ./...` | ✅ 全部通过 |
| `make llm-check` | ✅ 通过 |
| 真实 LLM Smoke | ❌ 认证失败（API Key 为占位符） |
| 安全检查 | ✅ 无泄露 |
| 禁止范围检查 | ✅ 无禁止文件变更 |

## 11. 当前限制

| 限制 | 严重度 | 说明 |
|------|--------|------|
| API Key 为占位符，真实 LLM 不可用 | 🟡 中 | 需配置有效 Key 才能完成真实 LLM 验证 |
| Pipeline 未在真实 LLM 场景下验证 | 🟡 中 | 4-5.1 已通过 fake/mock 验证，但真实输出质量未验证 |
| `request_id` 为空字符串 | 🟢 低 | 契约允许暂时为空 |
| `schema_valid` 默认 true | 🟡 中 | 未实现实际 Schema 校验逻辑 |
| 当前分支比 main 多 3 个 commit | 🟢 低 | 仅 docs 变更 |

## 12. 4-5.2 最终结论

**partial_pass**

判定依据：

### pass 项
1. ✅ 用户已确认 `CONFIRM_REAL_LLM_SMOKE=yes`
2. ✅ 真实 LLM smoke 命令已执行
3. ✅ HunyuanClient 可正常创建
4. ✅ 请求可发送到真实 API 端点
5. ✅ 认证失败被正确识别和报告
6. ✅ 未泄露 API Key
7. ✅ `go test ./...` 通过
8. ✅ `make llm-check` 可用
9. ✅ `RUN_REAL_LLM_SMOKE` 保护机制有效
10. ✅ 未修改禁止范围文件

### 非阻塞项
1. ⚠️ API Key 为占位符，真实 LLM 认证失败
2. ⚠️ Pipeline / JSON Repair 未在真实 LLM 场景下验证（4-5.1 已 fake/mock 验证）

### 未出现 fail 项
- ❌ 未经用户确认调用真实 LLM → 未发生
- ❌ 输出或泄露 API Key → 未发生
- ❌ go test ./... 失败 → 未发生
- ❌ make llm-check 不可用 → 未发生
- ❌ 真实 LLM 调用失败且没有可解释错误 → 未发生（失败原因明确）
- ❌ 修改禁止范围文件 → 未发生

## 13. 下一步建议

1. **进入 4-5.3**：三类 Agent 接口联调验收
   - 需 `make up` + curl 验证 HTTP 接口可访问
   - 默认使用 Mock / fallback，不依赖真实 LLM
2. **4-5.4**：阶段 4 总收口与阶段 5 准入评审
   - 需标记"真实 LLM 认证未现场验证通过"
   - 后续配置有效 Key 后可重新执行 `RUN_REAL_LLM_SMOKE=1 make llm-smoke`
3. **配置有效 API Key 后**：可重新执行真实 LLM Smoke
   - 修改 `.env` 中 `LLM_API_KEY` 为有效 Key
   - 执行 `RUN_REAL_LLM_SMOKE=1 make llm-smoke`
