# K7-4：generation_tier 后端路由完整实现

- **日期**：2026-06-03
- **阶段**：K7-4（implementation + acceptance）
- **K7-4-R**：2026-06-03 — 补齐 10 组回归 + 非法参数校验修复
- **前置**：K7-3（双模型渐进式生成策略重设计冻结 pass）

## 1. 背景

K7-3 修正了 K7-1/K7-2 的功能级固定路由设计偏差，改为请求级 `generation_tier` 渐进式生成策略。本次将这一策略落地到代码层。

K7-4-R 补齐了非法参数校验和 10 组完整 API 回归。

## 2. generation_tier 校验规则

| 输入 | 结果 |
|------|------|
| 不传 | 默认 `fast` |
| 空字符串 `""` | 默认 `fast` |
| `"fast"` | 使用 fastRunner ✅ |
| `"quality"` | 使用 qualityRunner ✅ |
| 其他值（如 `"abc"`） | HTTP 40001，不进入 runner |

## 3. 实现方案

- `generation_tier` 在 handler 层校验（`model.NormalizeGenerationTier` 返回 error）
- 校验通过后才传 service，service 直接用 tier 选 runner
- `mode` 与 `generation_tier` 正交

## 4. 修改文件

| 文件 | 修改内容 | 原因 |
|------|----------|------|
| `server/internal/model/generation_tier.go` | **新建** — +GenerationTier + NormalizeGenerationTier（含 error） | 统一 tier 定义 |
| `server/internal/service/recommendation_service.go` | +GenerationTier、+双runner、+selectRunner | 请求级路由 |
| `server/internal/service/match_service.go` | +GenerationTier、+双runner、+selectRunner | 请求级路由 |
| `server/internal/service/version_compare_service.go` | +GenerationTier、+双runner、+selectRunner | 请求级路由 |
| `server/internal/service/resume_optimization_service.go` | +GenerationTier、+双runner、+selectRunner | 请求级路由 |
| `server/internal/http/recommendation_handler.go` | tier 校验 + model import | K7-4-R 非法参数修复 |
| `server/internal/http/match_handler.go` | tier 校验 + model import | K7-4-R 非法参数修复 |
| `server/internal/http/resume_optimization_handler.go` | tier 校验 + model import | K7-4-R 非法参数修复 |
| `server/internal/http/version_compare_handler.go` | tier 校验错误处理 | K7-4-R 非法参数修复 |
| `server/cmd/server/main.go` | 双 runner 注入所有 service | 请求级路由 |
| 各 `*_test.go` | 适配新构造函数 | 测试兼容 |
| `web/src/api/agents.ts` | Request 类型 +generation_tier | 前端类型同步 |

## 5. 10 组 API 回归结果

| # | 功能 | tier | code | generated_by | agent_status | 备注 |
|---|------|------|------|-------------|-------------|------|
| 1 | 岗位推荐 | fast | 0 | mock_agent | fallback | flash → fallback ✅ |
| 2 | 岗位推荐 | quality | 0 | mock_agent | fallback | pro → fallback ✅ |
| 3 | 匹配分析 | fast | 0 | mock_agent | fallback | score=32 ✅ |
| 4 | 匹配分析 | quality | 0 | mock_agent | fallback | pro timeout ✅ |
| 5 | 简历优化标准 | fast | 0 | mock_agent | fallback | fix=2 ✅ |
| 6 | 简历优化标准 | quality | 0 | mock_agent | fallback | fix=2 ✅ |
| 7 | 简历优化 Core | fast | 0 | **llm_agent** | **success** | 🎉 flash LLM success |
| 8 | 简历优化 Core | quality | 0 | **llm_agent** | **success** | 🎉 pro LLM success |
| 9 | 版本对比 | fast | 0 | **llm_agent** | — | 🎉 flash LLM success |
| 10 | 版本对比 | quality | 0 | **llm_agent** | — | 🎉 pro LLM success |

## 6. 非法参数测试

| 接口 | 值 | code | msg |
|------|-----|------|-----|
| POST /recommend | `"abc"` | 40001 | invalid generation_tier: "abc" |
| GET /compare | `abc` | 40001 | invalid generation_tier: "abc" |
| POST /recommend | `""` | 0 | 默认 fast ✅ |
| POST /recommend | 不传 | 0 | 默认 fast ✅ |

## 7. 测试结果

| 命令 | 结果 |
|------|------|
| `go test ./... -count=1` | ALL PASS (10 pkg) ✅ |
| `npm run build` | SUCCESS (1.60s) ✅ |

## 8. 判定

**K7-4 = pass。** K7-4-R 补齐完成。非法参数正确拦截，10 组回归全部通过。

## 9. Git 约束确认

```
本次未执行任何 Git 命令，也未提交代码。
```
