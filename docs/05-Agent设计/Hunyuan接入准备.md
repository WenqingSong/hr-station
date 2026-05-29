# Hunyuan 接入准备

> 最后更新：2026-05-29（阶段 4-1.4.0 Prompt / Agent 真实接入方案）

## 1. 用户需要准备的信息

| 配置项 | 是否必填 | 填写位置 | 说明 | 示例 |
|--------|---------|---------|------|------|
| `LLM_PROVIDER` | 是 | .env | 模型提供商 | `hunyuan` |
| `LLM_BASE_URL` | 是 | .env | Hunyuan API Base URL | `https://api.hunyuan.cloud.tencent.com/v1` |
| `LLM_API_KEY` | 是 | .env | 腾讯云混元 API Key，只填本地 .env | 不要写入文档 |
| `LLM_MODEL` | 是 | .env | 可用模型名称 | `hunyuan-turbos-latest` |
| `LLM_TIMEOUT_SECONDS` | 否 | .env | 请求超时秒数 | `20` |
| `LLM_MAX_RETRIES` | 否 | .env | 失败重试次数 | `1` |
| `LLM_MAX_OUTPUT_TOKENS` | 否 | .env | 最大输出 token | `2048` |
| `LLM_TEMPERATURE` | 否 | .env | 温度 | `0.2` |
| `ENABLE_LLM_AGENT` | 是 | .env | 是否启用真实 LLM | `false` / `true` |
| `ENABLE_LLM_FALLBACK` | 是 | .env | 是否失败回退 Mock | `true` |

## 2. 用户需要完成的腾讯云操作

1. 登录腾讯云控制台
2. 开通混元大模型服务
3. 确认是否有免费额度
4. 创建或获取 API Key
5. 确认可用模型名称（建议 `hunyuan-turbos-latest`）
6. 确认 API Base URL（建议 `https://api.hunyuan.cloud.tencent.com/v1`）
7. 将 Key 和模型配置填写到本地 `.env`
8. 不要把 `.env` 提交到 Git

## 3. 用户填写后的本地 .env 示例

```env
ENABLE_LLM_AGENT=true
ENABLE_LLM_FALLBACK=true
LLM_PROVIDER=hunyuan
LLM_BASE_URL=https://api.hunyuan.cloud.tencent.com/v1
LLM_API_KEY=请填写你的真实Key
LLM_MODEL=hunyuan-turbos-latest
LLM_TIMEOUT_SECONDS=20
LLM_MAX_RETRIES=1
LLM_MAX_OUTPUT_TOKENS=2048
LLM_TEMPERATURE=0.2
```

**重要提醒**：
- `LLM_API_KEY` 只能填写在本地 `.env` 中
- 不要将真实 Key 写入 `.env.example`
- 不要将 `.env` 提交到代码仓库
- 如果 `.env` 已被 git 追踪，需执行 `git rm --cached .env`

## 4. 用户填写完成后的验证方式

本次（4-1）不执行验证。下一步 4-2 实现 LLM Client 后可验证：

1. 检查 `.env` 是否存在
2. 检查 `LLM_API_KEY` 是否非空
3. 使用 LLM Client 测试 `/v1/chat/completions` 接口
4. 要求模型返回固定 JSON 格式
5. JSON 解析成功后才允许进入业务 Agent 改造

## 5. Hunyuan API 调用示例

### 5.1 curl 测试

```bash
curl -s -X POST https://api.hunyuan.cloud.tencent.com/v1/chat/completions \
  -H "Authorization: Bearer $LLM_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "hunyuan-turbos-latest",
    "messages": [
      {"role": "system", "content": "你是一个简历分析助手。只输出 JSON，不要输出其他内容。"},
      {"role": "user", "content": "分析以下简历与岗位的匹配度，输出 JSON：{\"match_score\": 85, \"summary\": \"...\"}"}
    ],
    "temperature": 0.2,
    "max_tokens": 2048
  }'
```

### 5.2 预期响应结构

```json
{
  "id": "chatcmpl-xxx",
  "object": "chat.completion",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "{\"match_score\": 85, \"summary\": \"...\"}"
      },
      "finish_reason": "stop"
    }
  ],
  "usage": {
    "prompt_tokens": 100,
    "completion_tokens": 50,
    "total_tokens": 150
  }
}
```

## 6. 安全注意事项

1. **不要在 `.env.example` 中写入真实 API Key**
2. **不要将 `.env` 提交到代码仓库**
3. **API Key 应妥善保管，不要分享给他人**
4. **注意 API 调用费用，建议关注免费额度使用情况**
5. **生产环境应使用更严格的密钥管理方案**

## 阶段 4-1.4 更新

4-1.4 只允许手动 curl 或离线方式测试 Prompt：

- 不在 Go 后端中接入 Hunyuan
- 不记录 API Key 到文档
- .env.example 中 LLM_API_KEY 应保持为空
- curl 测试时使用本地 .env 中的 LLM_API_KEY
- 方案文档见 `docs/05-Agent设计/阶段4-1.4Prompt与Agent真实接入方案.md`

### 4-1.4.2 手动调用验收记录

- 已执行一次匹配分析 Prompt 手动调用验收
- 使用 TokenHub（腾讯混元代理）+ hy3-preview 模型
- 首次调用 max_tokens=4096 导致截断，二次调用 max_tokens=8192 成功
- prompt_tokens=6604, completion_tokens=3618, total_tokens=10222
- LLM_MAX_OUTPUT_TOKENS 生产环境建议设置为 ≥8192
- 不记录 API Key
- 不记录完整模型原始输出
- 验收报告见 `docs/05-Agent设计/匹配分析Prompt首轮输出验收报告.md`

### 4-1.4.5 推荐 Prompt 首轮输出验收记录

- 已执行一次推荐 Prompt 手动调用验收
- 使用 TokenHub（腾讯混元代理）+ hy3-preview 模型
- max_tokens=8192，调用成功（finish_reason=stop）
- prompt_tokens=8692, completion_tokens=3832, total_tokens=12524
- 模型输出 JSON 合法，字段完整，证据链真实
- 不记录 API Key
- 不记录完整模型原始输出
- 验收报告见 `docs/05-Agent设计/推荐Prompt首轮输出验收报告.md`
- 后续推荐 Prompt 的模型调用也建议 max_tokens >= 8192

### 4-1.4.8 简历优化 Prompt 首轮输出验收记录

- 已执行一次简历优化 Prompt 手动调用验收
- 使用 TokenHub（腾讯混元代理）+ hy3-preview 模型
- max_tokens=8192，调用成功（finish_reason=stop）
- prompt_tokens=6723, completion_tokens=4032, total_tokens=10755
- 模型输出 JSON 合法，15 个顶层字段完整
- 不记录 API Key
- 不记录完整模型原始输出
- 验收报告见 `docs/05-Agent设计/简历优化Prompt首轮输出验收报告.md`
- 后续简历优化 Prompt 的模型调用也建议 max_tokens >= 8192

### 4-1.4.9B 简历优化 JSON Schema 登记记录

- 简历优化 JSON Schema 已登记：`server/prompts/schema/resume_optimization_schema.json`
- Schema 验证结果：JSON 合法，15 个顶层 required 字段完整，10 个 $defs 子结构完整
- Schema 已吸收 4-1.4.8 首轮验收问题（允许 rewrite_examples 全部 can_use_directly=false、涉及数字/指标等时必须 requires_user_verification=true、不得编造具体数字/指标/系统规模/职责或成果、p2 允许 0-4 项）
- 当前 4-1.4.9B 只登记 Schema，不再次调用 Hunyuan
- 后续简历优化 Prompt 的模型调用建议 LLM_MAX_OUTPUT_TOKENS ≥8192

### 4-1.4.7B 简历优化 Agent 指令模板登记记录

- 简历优化 Prompt 已登记：`server/prompts/resume_optimization_v1.md`
- 模板面向"目标岗位导向的简历优化"场景，支持简历 + 目标岗位 JD + 可选匹配分析结果输入
- 模板输出 P0/P1/P2 优化建议、关键词对齐、项目优化、技能栏优化、摘要优化、缺失证据补足计划、改写示例和真实性提醒
- 当前 4-1.4.7B 只登记 Prompt，不调用 Hunyuan
- 后续简历优化 Prompt 首轮输出验收建议继续使用 max_tokens >= 8192

### 4-1.4.6B 推荐 JSON Schema 登记记录

- 推荐 JSON Schema 已登记：`server/prompts/schema/recommendation_schema.json`
- Schema 验证结果：JSON 合法，7 个顶层 required 字段完整，7 个 $defs 子结构完整
- Schema 已吸收 4-1.4.5 首轮验收问题（四档推荐固定但不强制覆盖、recommend_level 与 recommend_label 一致、match_score 区间一致）
- 当前 4-1.4.6B 只登记 Schema，不再次调用 Hunyuan
- 后续推荐 Prompt 的模型调用建议 LLM_MAX_OUTPUT_TOKENS ≥8192

### 4-1.4.3B Schema 登记记录

- 匹配分析 JSON Schema 已登记：`server/prompts/schema/match_analysis_schema.json`
- Schema 验证结果：JSON 合法，17 个顶层 required 字段完整，9 个 $defs 子结构完整
- Schema 已吸收 4-1.4.2 首轮验收问题（rewrite_example 禁止编造数字、evidence_chain 支持 missing）
- 当前 4-1.4.3B 只登记 Schema，不再次调用 Hunyuan
- 后续 Hunyuan 调用建议 LLM_MAX_OUTPUT_TOKENS ≥8192

## 关联文档

- [阶段4Agent与模型接入方案.md](./阶段4Agent与模型接入方案.md)
- [LLMClient设计.md](./LLMClient设计.md)
- [JSON输出规范.md](./JSON输出规范.md)
- [../06-工程部署/环境变量.md](../06-工程部署/环境变量.md)
