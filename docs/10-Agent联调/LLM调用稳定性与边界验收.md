# LLM 调用稳定性与边界验收

> 阶段：7-5 | 状态：pass | 日期：2026-06-01

## LLM Client 结构

| 模块 | 文件 | 状态 |
|------|------|:--:|
| LLM Interface | `client.go` | ✅ |
| HunyuanClient | `hunyuan_client.go` | ✅ |
| NoopClient | `noop_client.go` | ✅ |
| Config (timeout) | `config.go` | ✅ |
| Errors | `errors.go` (含 ErrLLMTimeout) | ✅ |
| Pipeline | `pipeline/pipeline.go` | ✅ |
| PromptLoader | `prompt/loader.go` | ✅ |
| SchemaLoader | `schema/loader.go` | ✅ |
| Validator | `schema/validator.go` | ✅ |
| JSON parse | agent/match_agent.go 等 | ✅ |

## 环境变量

| 配置项 | 状态 |
|------|:--:|
| API Key | 未配置（无 Key fallback） |
| Provider | Hunyuan (OpenAI-compatible) |
| Timeout | `TimeoutSeconds` 配置项 |
| Model | `LLM_MODEL` 配置项 |

## 无 API Key 验收

| Agent | code | result | agent_status | fallback_reason | 结论 |
|------|:--:|------|------|------|:--:|
| 推荐 | 0 | 2 jobs | fallback | llm_api_key_missing | ✅ |
| 匹配 | 0 | score=90 | fallback | llm_api_key_missing | ✅ |
| 优化 | 0 | readiness | fallback | llm_api_key_missing | ✅ |

## 有 API Key 验收

未执行（环境无真实 Key）— P2

## 超时 / JSON parse / Schema 校验

| 场景 | 代码路径 | 状态 |
|------|------|:--:|
| Timeout | `ErrLLMTimeout` + `TimeoutSeconds` | ✅ |
| JSON parse | `json.Unmarshal` + Pipeline 错误捕获 | ✅ |
| Schema | `Validator.Validate()` + schema_valid meta | ✅ |

## meta 记录

三类 Agent 均返回: `agent_name`, `agent_status`, `fallback_reason`, `schema_valid` ✅

## 前端工程化文案泄漏

**零** ✅

## build / URL

build 1.66s ✅ | 3 URL → 200 ✅

## P0/P1/P2

| P0 | 0 | — |
| P1 | 0 | — |
| P2 | 1 | 有 Key 真实调用未执行 |

## 结论

**pass ✅** — 可进入 7-6
