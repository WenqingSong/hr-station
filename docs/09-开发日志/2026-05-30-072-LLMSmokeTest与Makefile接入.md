# 2026-05-30-072-LLMSmokeTest与Makefile接入

## 1. 本次任务

阶段 4-2.6：LLM Smoke Test / Makefile 接入

目标：为 HunyuanClient 提供最小连通性验证入口，不接业务 Agent。

## 2. 修改范围

### 新增
- `server/cmd/llm-smoke/main.go`：LLM smoke test Go 命令
- `scripts/dev/llm_smoke.sh`：LLM smoke test 脚本（help / check / smoke）
- Makefile 新增 `llm-help` / `llm-check` / `llm-smoke` target

### 未修改
- `.env` / `.env.example`
- `server/prompts/` / `server/prompts/schema/`
- `server/internal/agent/` / `server/internal/service/`
- `server/internal/controller/` / `server/internal/repository/`
- `server/sql/`
- `web/` / `docker-compose.yml` / `README.md`

## 3. Smoke Test 设计

### 目标

Smoke test 只回答一个问题：Hunyuan / TokenHub 当前配置是否能通过 HunyuanClient 正常返回内容？

### 最小 Prompt

```json
[
  {
    "role": "system",
    "content": "你是一个连通性测试助手。你只能输出严格 JSON，不要输出 Markdown。"
  },
  {
    "role": "user",
    "content": "请只输出一个 JSON 对象：{\"ok\":true,\"message\":\"pong\"}"
  }
]
```

### 通过标准

1. 成功读取 LLM_* 配置
2. LLM_ENABLE=true
3. LLM_PROVIDER=hunyuan
4. LLM_BASE_URL / LLM_API_KEY / LLM_MODEL 非空
5. HunyuanClient.Chat 返回无错误
6. ChatResponse.RawContent 非空
7. ChatResponse.RawOutputLength > 0
8. 不输出 API Key
9. 不输出完整 raw output（最多 120 字符预览）
10. 尝试 llmjson.ExtractAndParse，记录 json_parse_success

## 4. Go smoke 命令

`server/cmd/llm-smoke/main.go`

职责：
1. 从环境变量读取 LLM_* 配置
2. 构造 llm.Config
3. 校验 Config
4. 创建 HunyuanClient
5. 构造最小 ChatRequest（temperature=0, max_tokens=256）
6. 调用 Chat
7. 输出安全 smoke 结果
8. 尝试调用 llmjson.ExtractAndParse
9. 不输出 API Key
10. 不输出完整 raw output
11. 不读取简历 / 岗位 / 数据库

输出格式：
```
LLM smoke test result:
  provider=hunyuan
  model=...
  finish_reason=...
  raw_output_length=...
  duration_ms=...
  json_parse_success=true/false
  raw_output_preview=...
```

安全：
- maskKey() 对 API Key 做掩码处理
- safePreview() 对 raw output 做 120 字符截断 + 换行替换

## 5. 脚本设计

`scripts/dev/llm_smoke.sh`

### help
输出可用命令和安全说明。

### check
只做本地检查，不请求外部模型：
1. 检查 .env 是否存在
2. 检查 LLM_ENABLE
3. 检查 LLM_PROVIDER
4. 检查 LLM_BASE_URL 是否配置
5. 检查 LLM_API_KEY 是否配置（只输出已配置/未配置）
6. 检查 LLM_MODEL 是否配置
7. 检查 server/cmd/llm-smoke 是否存在
8. 检查 HunyuanClient 编译是否通过
9. 不打印 API Key
10. 不调用 Hunyuan

### smoke
执行真实 smoke test，但必须要求：
```bash
RUN_REAL_LLM_SMOKE=1 scripts/dev/llm_smoke.sh smoke
```
或：
```bash
RUN_REAL_LLM_SMOKE=1 make llm-smoke
```

如果未设置 RUN_REAL_LLM_SMOKE=1，跳过真实调用并提示。

如果已设置：
1. 从 .env 加载环境变量
2. 执行 go run ./cmd/llm-smoke
3. 请求次数控制为 1 次
4. 不输出 API Key
5. 不输出完整 raw output

### .env 加载
- 使用 source .env 方式加载
- 不打印 .env 内容
- 不修改 .env

## 6. Makefile 接入

新增 3 个 target：

| target | 说明 |
|--------|------|
| `make llm-help` | 查看 LLM smoke test 命令帮助 |
| `make llm-check` | 检查本地 LLM 配置，不请求外部模型 |
| `make llm-smoke` | 执行 LLM smoke test，默认需要 RUN_REAL_LLM_SMOKE=1 |

已加入 `.PHONY` 声明。
已加入 `make help` 输出。
不在 `make up` / `make ci` 中自动执行。

## 7. 安全边界

1. ✅ 使用 fake key 测试（smoke 命令使用 .env 中的真实 key，但不输出）
2. ✅ 测试日志中没有真实 API Key
3. ✅ 错误字符串不包含 fake key / real key
4. ✅ 未修改 .env
5. ✅ 未修改 .env.example
6. ✅ 未在 docs / 开发日志写入真实 API Key
7. ✅ 未调用真实 Hunyuan（因为没有设置 RUN_REAL_LLM_SMOKE=1 且本地 .env 中 LLM_ENABLE=false）
8. ✅ 未调用真实 TokenHub
9. ✅ 未请求外部网络
10. ✅ make llm-check 不请求外部模型
11. ✅ make llm-smoke 需要 RUN_REAL_LLM_SMOKE=1

## 8. 测试结果

### 执行命令

```bash
cd server && go build ./cmd/llm-smoke           # 编译通过
cd server && go test ./internal/llm -v           # 49 项全部 PASS
cd server && go test ./internal/llmjson           # PASS
cd server && go test ./...                        # 全量 PASS
cd server && go vet ./...                         # 无警告
make llm-help                                     # 正常输出
make llm-check                                    # 正常输出配置检查
make llm-smoke                                    # 提示需要 RUN_REAL_LLM_SMOKE=1
```

### 全量测试结果

```
ok  offer-catcher/server/internal/llm
ok  offer-catcher/server/internal/llmjson
ok  offer-catcher/server/internal/prompt
ok  offer-catcher/server/internal/schema
```

无回归。

## 9. 真实 Hunyuan smoke test 结果

**未执行真实 Hunyuan smoke test。**

原因：
1. 本地 .env 不存在或 LLM_ENABLE=false
2. 未设置 RUN_REAL_LLM_SMOKE=1

请求次数：0

不阻塞 4-2.6 完成。本阶段核心是提供 smoke test 入口和 Makefile 接入，真实连通性验证可在后续配置真实 API Key 后执行。

## 10. 文档更新

- 更新 `docs/05-Agent设计/阶段4-2LLMClient与Hunyuan接入方案.md`
- 更新 `docs/05-Agent设计/LLMClient设计.md`
- 更新 `docs/05-Agent设计/Hunyuan接入准备.md`
- 更新 `docs/05-Agent设计/JSON输出规范.md`
- 更新 `docs/05-Agent设计/失败兜底与Mock回退策略.md`
- 更新 `docs/06-工程部署/Makefile说明.md`
- 更新 `docs/06-工程部署/环境变量.md`
- 更新 `docs/06-工程部署/部署说明.md`
- 更新 `docs/项目状态.md`
- 更新 `docs/下一步任务.md`
- 更新 `docs/变更记录.md`
- 更新 `docs/说明书.md`

## 11. 未修改内容

- 未修改 `.env` / `.env.example`
- 未写入真实 API Key
- 未输出真实 API Key
- 未修改 Prompt 文件
- 未修改 Schema 文件
- 未修改业务 Agent
- 未实现 JSONRepairAgent 调用
- 未实现 Mock fallback
- 未接推荐 / 匹配 / 简历优化接口
- 未修改 SQL
- 未修改前端
- 未修改 docker-compose.yml

## 12. 当前风险

- 低风险：真实 Hunyuan 连通性尚未验证，需在配置真实 API Key 后执行 `RUN_REAL_LLM_SMOKE=1 make llm-smoke`
- 低风险：如果真实 Hunyuan 响应结构与 OpenAI-compatible 格式有差异，可能需要微调
- 低风险：smoke test 使用 temperature=0，部分模型可能不支持精确 0 值

## 13. 是否可以进入 4-2.7

可以进入 4-2.7。

理由：
1. HunyuanClient 连通性验证入口已实现
2. Makefile llm-help / llm-check / llm-smoke 已接入
3. 安全边界已验证
4. 不输出 API Key
5. 不输出完整 raw output
6. 全量测试通过，无回归
7. 真实 smoke test 可在后续配置后执行，不阻塞 Pipeline 层实现

## 14. 下一步建议

1. **4-2.7**：JSONRepairAgent 调用链路 — 实现 repair once + Mock fallback decision
2. **4-2.8**：统一 LLM Agent Pipeline — LLMRunner：Chat → parse → validate → repair → fallback
3. 配置真实 API Key 后执行 `RUN_REAL_LLM_SMOKE=1 make llm-smoke` 验证真实连通性
