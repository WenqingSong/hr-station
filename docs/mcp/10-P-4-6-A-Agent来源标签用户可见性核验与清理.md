# 10-P-4-6-A Agent 来源标签用户可见性核验与清理

## 1. 阶段信息

* 阶段：10-P-4-6-A
* 日期：2026-06-02
* 范围：Agent 来源标签用户可见性核验 + `formatAgentSourceLabel` 产品化
* 是否修改代码：是
* 是否修改前端：是（1 个文件）
* 是否修改后端：否
* 是否修改数据库：否
* 是否接真实 LLM：否

## 2. 本次结论摘要

* `"Mock 演示结果"` 当前**不可见**于任何活跃页面 — `formatAgentSourceLabel()` 的输出存储在 ViewModel 的 `meta` 字段中，但所有 4 个页面均未渲染该字段
* `AIBadge` 组件包含 "Mock 演示" / "LLM 生成" / "兜底结果" 文案，但**仅被 `StyleGuide.tsx`（P2 死页面）引用**，活跃页面零引用
* 推荐页 `sourceLabel` 来自简历状态（`statusLabel`），非 Agent meta
* 匹配分析页 Header 显示 "分析完成" 硬编码，非 Agent 来源
* 简历优化页 Header 显示 "优化完成" 硬编码，非 Agent 来源
* 版本对比页无 Agent 来源标签
* 已修复 `formatAgentSourceLabel()`：移除所有技术性标签（Mock/演示/兜底/LLM），改为产品化表达（"分析完成"/"分析异常"）
* `generated_by` / `agent_status` 保留为内部元数据，不影响用户体验
* `npm build` / `npm typecheck` / `go build` / `go test` 全部通过
* 未接真实 LLM、未修改后端

## 3. 执行前清查

| 检查项 | 位置 | 是否用户可见 | 结论 |
| --- | -- | ------ | -- |
| `"Mock 演示结果"` | `web/src/adapters/common.ts:34` | 否 | 仅 adapter 映射函数，输出存储在 ViewModel.meta，页面不渲染 |
| `"LLM 生成"` / `"已自动修复"` / `"兜底结果"` / `"异常但可用"` | `web/src/adapters/common.ts:25-37` | 否 | 同上 |
| AIBadge `"Mock 演示"` / `"LLM 生成"` / `"兜底结果"` | `web/src/app/components/ui/AIBadge.tsx:16,26,31` | 否 | 仅被 `StyleGuide.tsx`（P2 死页面）引用 |
| `generated_by` / `agent_status` | adapter VM `meta` 字段 | 否 | 页面代码零引用 `meta` |
| `sourceLabel` (RecommendationHeader) | `RecommendationHeader.tsx:36` | 是 | 但来源是简历 `statusLabel`，非 Agent meta |
| `sourceLabel` (RecommendationContext) | `RecommendationContext.tsx:39` | 是 | 同上 |

## 4. 页面级核验矩阵

| 页面 / 模块 | 是否展示来源标签 | 原文案 | 修改后 | 是否用户可见 |
| ------- | -------- | --- | --- | ------ |
| 推荐岗位页 Header | 是（"推荐来源"） | `selectedOption.statusLabel ± directionLabel`（简历数据） | 不修改 | 是（来自简历，非 Agent） |
| 推荐岗位页 Context | 是（"推荐来源"） | 同上 | 不修改 | 是（来自简历，非 Agent） |
| 匹配分析页 Header | 否（硬编码） | "分析完成" | 不修改 | 是（产品化文案） |
| 简历优化页 Header | 否（硬编码） | "优化完成" | 不修改 | 是（产品化文案） |
| 版本对比页 | 无来源标签 | — | — | 否 |
| 推荐页 Sidebar/Detail | 无来源标签 | — | — | 否 |
| 匹配分析 Detail | 无来源标签 | — | — | 否 |
| 优化 Detail | 无来源标签 | — | — | 否 |
| AIBadge 组件 | 仅 StyleGuide 使用 | "Mock 演示" 等 | 不修改（死代码） | 否（不可达） |

## 5. 修改内容清单

| 文件 | 修改内容 | 是否用户可见 | 说明 |
| -- | ---- | ------ | -- |
| `web/src/adapters/common.ts` | `formatAgentSourceLabel()` 去技术化：`mock_success`/`fallback`/`success` → "分析完成"（blue）；`error` → "分析异常"（red）；`repaired` → "已自动修复"（green） | 否（当前不渲染，预防未来泄漏） | 保留函数和 adapter 合约 |

（仅 1 文件，+8 -9 行）

## 6. 来源标签设计冻结

* 用户可见页面不展示 Mock / fallback / generated_by / agent_status 技术标签
* `generated_by` / `agent_status` 仅作为内部元数据（保留在 `AgentResponseMeta` 和 `AgentMetaViewModel.meta` 中）
* 默认 Mock Agent 结果在用户侧统一表达为 "分析完成" / "优化完成" / "推荐完成"
* LLM 真实接入后，用户侧仍不展示技术 provider 名称
* 错误态显示产品化错误文案
* AIBadge 组件为 P2 死代码，不修改（仅被 StyleGuide 引用）

## 7. 反弹检查结果

| 检查项 | 结果 | 说明 |
| --- | -- | -- |
| `"Mock 演示结果"` / `"Mock 演示"` 活跃代码 | 零匹配 | `AIBadge.tsx` 中含 "Mock 演示" 但仅被 StyleGuide（P2 死页面）引用 |
| `"演示数据"` / `"已展示演示"` / `"当前为演示"` | 零匹配 | — |
| `"兜底结果"` / `"LLM 生成"` 活跃代码 | 零匹配 | 同上，仅 AIBadge 死代码 |
| `as any` / `@ts-ignore`（活跃代码） | 零匹配 | — |
| `withFallback` / `withMockFallback` / `fallbackToMock` | 零匹配 | — |
| mock import（活跃代码） | 零匹配 | — |
| `generated_by` / `agent_status` 页面渲染 | 零匹配 | 仅内部存储 |
| `formatAgentSourceLabel` 输出标签 | 已产品化 | "分析完成" / "分析异常" / "已自动修复" |

## 8. 验证结果

| 命令 | 结果 | 说明 |
| -- | -- | -- |
| `git status --short` 执行前 | `web/src/adapters/common.ts` modified | 仅 1 文件 |
| `npm run build` | ✅ PASS | 1.79s，1676 modules |
| `npm run typecheck` | ✅ PASS | 零错误 |
| `go build ./cmd/server` | ✅ PASS | 编译通过 |
| `go test ./...` | ✅ PASS | 9 packages PASS |
| 反弹检查 | ✅ | 零活跃代码技术标签泄漏 |
| `git status --short` 执行后 | 1 文件 modified | 仅 `common.ts` |

## 9. 未做事项

* 未修改后端 Agent 逻辑
* 未修改数据库 schema
* 未新增接口
* 未接真实 LLM
* 未改推荐 / 匹配 / 优化算法
* 未做页面布局重构
* 未修改 AIBadge.tsx（P2 死代码，非本阶段范围）
* 未执行任何版本控制写入操作

## 10. 当前风险与后续建议

* **已解决**：`formatAgentSourceLabel()` 去技术化，所有标签改为产品化表达
* **已解决**：核验确认 4 个 Agent 页面均无技术标签泄漏
* **不需要处理，原因是**：`AIBadge.tsx` 含 "Mock 演示" 但仅被 StyleGuide（P2 死页面）引用，不影响用户。后续死代码清理阶段处理。
* **需要下一阶段处理，阶段名是 10-P-4-6-C（建议）**：10-P-4-6 系列封存验收

## 11. docs 更新清单

| 文件 | 新增 / 更新 | 目的 |
| -- | ------- | -- |
| `docs/mcp/项目上下文.md` | 更新 | 更新当前阶段为 10-P-4-6-A |
| `docs/mcp/10-P-4-6-A-Agent来源标签用户可见性核验与清理.md` | 新增 | 本阶段 MCP 记录 |
| `docs/09-开发日志/2026-06-02-248-10-P-4-6-A-Agent来源标签清理.md` | 新增 | 开发日志 |

## 12. CodeBuddy 重启恢复摘要

10-P-4-6-A 完成 Agent 来源标签用户可见性核验与清理。核验确认 4 个页面的来源标签均不来自 Agent meta（推荐页来自简历数据，匹配/优化页为硬编码产品化文案，版本对比页无标签）。`AIBadge` 组件含技术标签但仅被 StyleGuide（P2 死页面）引用。已修复 `formatAgentSourceLabel()` 去除所有 Mock/演示/兜底/LLM 标签，改为产品化表达。`generated_by`/`agent_status` 保留为内部元数据。npm build/typecheck + go build/test 全部通过。10-P-4-6 系列可以封存。

下次继续前读取：`docs/mcp/项目上下文.md`、`docs/mcp/10-P-4-6-Agent-Mock-LLM边界整理.md`、本文件。

## 13. 最终判断

1. `"Mock 演示结果"` 是否用户可见？否。（输出存储在 ViewModel.meta，页面不渲染）
2. 用户可见页面是否仍有 Mock / 演示 / fallback 技术标签？否。（AIBadge 仅死代码）
3. generated_by / agent_status 是否保留为内部元数据？是。
4. 是否修改后端 Agent 逻辑？否。
5. 是否接入真实 LLM？否。
6. 验证命令是否通过？是。
7. 本阶段是否可以通过？是。
8. 10-P-4-6 是否可以封存？是。
9. 是否建议进入下一阶段？是。
