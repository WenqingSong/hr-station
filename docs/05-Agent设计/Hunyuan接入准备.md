# Hunyuan 接入准备

> 最后更新：2026-05-29（阶段 4-1 Hunyuan API 接入准备）

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

## 关联文档

- [阶段4Agent与模型接入方案.md](./阶段4Agent与模型接入方案.md)
- [LLMClient设计.md](./LLMClient设计.md)
- [JSON输出规范.md](./JSON输出规范.md)
- [../06-工程部署/环境变量.md](../06-工程部署/环境变量.md)
