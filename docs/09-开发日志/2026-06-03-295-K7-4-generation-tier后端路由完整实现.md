# K7-4：generation_tier 后端路由完整实现

- **日期**：2026-06-03
- **阶段**：K7-4（implementation + acceptance）
- **前置**：K7-3（双模型渐进式生成策略重设计冻结 pass）

## 1. 背景

K7-3 修正了 K7-1/K7-2 的功能级固定路由设计偏差，改为请求级 `generation_tier` 渐进式生成策略。本次将这一策略落地到代码层。

## 2. 实现方案

### generation_tier 定义

```
"fast"    → fastRunner（deepseek-v4-flash）
"quality" → qualityRunner（deepseek-v4-pro-202606）
空值/非法 → "fast"（默认）
```

### 参数位置

| 功能 | 接口方式 | tier 位置 |
|------|----------|----------|
| 岗位推荐 | POST JSON | `{"generation_tier": "fast"}` |
| 岗位匹配分析 | POST JSON | `{"generation_tier": "quality"}` |
| 简历优化 | POST JSON | `{"generation_tier": "quality"}` |
| 简历版本对比 | GET query | `?generation_tier=quality` |

### mode 与 generation_tier 关系

- `mode`：控制业务链路（standard / core）
- `generation_tier`：控制模型质量层级（fast / quality）
- 两者正交，独立设置

## 3. 修改文件

| 文件 | 修改内容 | 原因 |
|------|----------|------|
| `server/internal/model/generation_tier.go` | **新建** — GenerationTier 类型 + NormalizeGenerationTier | 统一 tier 定义 |
| `server/internal/service/recommendation_service.go` | +GenerationTier 字段、+fastRunner/+qualityRunner、+selectRunner | 请求级路由 |
| `server/internal/service/match_service.go` | +GenerationTier 字段、+fastRunner/+qualityRunner、+selectRunner | 请求级路由 |
| `server/internal/service/version_compare_service.go` | +GenerationTier 字段、+fastRunner/+qualityRunner、+selectRunner | 请求级路由 |
| `server/internal/service/resume_optimization_service.go` | +GenerationTier 字段、+fastRunner/+qualityRunner、+selectRunner | 请求级路由 |
| `server/internal/http/version_compare_handler.go` | query 解析 generation_tier | GET 参数支持 |
| `server/cmd/server/main.go` | 所有 service 同时注入 fastRunner + qualityRunner | 双 runner 注入 |
| 各 `*_test.go` | 适配新构造函数签名 | 测试兼容 |
| `web/src/api/agents.ts` | 所有 Request 类型 +generation_tier | 前端类型同步 |

## 4. API 回归结果

| 功能 | tier | code | generated_by | agent_status | 备注 |
|------|------|------|-------------|-------------|------|
| 岗位推荐 | fast | 0 | mock_agent | fallback | flash → fallback ✅ |
| 岗位推荐 | quality | 0 | mock_agent | fallback | pro → fallback ✅ |
| 匹配分析 | quality | 0 | mock_agent | fallback | score=32 ✅ |
| 版本对比 | quality | 0 | **llm_agent** | — | 🎉 quality LLM success |
| Core 优化 | quality | 0 | **llm_agent** | **success** | 🎉 quality LLM success |

## 5. 测试结果

| 命令 | 结果 |
|------|------|
| `go test ./internal/... -count=1` | ALL PASS (10 pkg) ✅ |
| `npm run build` | SUCCESS (1.56s) ✅ |
| `LLM_ENABLE` | `true` ✅ |

## 6. 判定

**K7-4 = pass。** generation_tier 请求级路由已实现，四主功能全部支持 fast/quality 选择，默认 fast，fallback 完整。

## 7. Git 约束确认

```
本次未执行任何 Git 命令，也未提交代码。
```
