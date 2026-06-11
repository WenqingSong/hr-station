# Dev-Reboot-Fix-C：前端错误态处理，禁止模型失败时显示 mock 内容

日期：2026-06-11

## 背景

`Dev-Reboot-Fix-B` 完成后：
- Fast 恢复 `fast_tag_engine`（非 mock）
- Deep 真实调用 Hy3 3-agent（`generated_by=llm_agent`）
- 但前端在接口失败时可能仍渲染 mock 数据/假文案

## 修复内容

### 1. 新增统一错误解析工具

`web/src/app/utils/agentError.ts`：
- `RealAgentError` — mock/fallback 专用错误类
- `isMockGeneratedBy(generatedBy)` — 检测 `mock_agent` / `mock_quality`
- `assertRealAgentResult()` — 统一断言（Fast 允许 `fast_tag_engine`，Deep 只接受 `llm_agent`）

### 2. 新增错误态组件

`web/src/app/components/common/AgentErrorState.tsx`：
- 标题：AI 模型连接失败
- 说明：当前没有生成真实 AI 结果
- 技术详情：功能名 + 来源标记（不暴露 key）
- 重新生成按钮

### 3. 四核心页面修复

| 页面 | 修复点 |
|---|---|
| JobRecommendationPage | deep 结果检测 `isMockGeneratedBy` |
| JobMatchAnalysisPage | 移除硬编码 `mock_agent`/`mock_quality`；fast 使用真实 `generatedBy`；deep 检测 mock |
| ResumeOptimizationPage | `getHeaderLabel` 区分 mock/llm；deep 检测 mock |
| ResumeComparisonPage | deep 检测 mock；错误态使用 `AgentErrorState` |

### 4. 适配层修复

| 文件 | 修复点 |
|---|---|
| `adapters/recommendationAdapter.ts` | `getSource()` 不再将 `mock_quality` 映射为 `deep_llm`；新增 `fast_tag` source |
| `adapters/common.ts` | `formatAgentSourceLabel()` 区分 AI 生成 vs 规则兜底 vs 异常 |
| `adapters/matchAnalysisAdapter.ts` | `MatchAnalysisPageViewModel` 新增 `generatedBy`/`fallback` 字段 |
| `api/viewModels.ts` | `loadMatchAnalysisViewModel` 传递 meta 字段 |

### 5. 成功/失败判断

| 条件 | 处理 |
|---|---|
| `generated_by=llm_agent` | ✅ 正常展示（Deep 合法来源） |
| `generated_by=fast_tag_engine` | ✅ 正常展示（Fast 合法来源） |
| `generated_by=mock_agent` | ❌ 显示错误态 |
| `generated_by=mock_quality` | ❌ 显示错误态 |
| `fallback=true` | ❌ 显示错误态 |
| HTTP 500 | ❌ 显示错误态 |

## 未修改
- 后端源码
- 数据库
- .env / .env.example
- docker-compose
- Git 操作

## 构建
- `npm run build` ✅ 通过

## 结论

前端在模型不可用/返回 mock 时不再静默渲染假数据，显示明确的"AI 模型连接失败"错误态。
