# Dev-Reboot-Acceptance-D：前端 Agent 错误态与四核心流程最终验收

日期：2026-06-11

## 背景

Dev-Reboot 系列的最终验收。前两个 Fix 阶段已完成：
- Fix-B：job_profiles 补齐、mock/fallback 关闭、Hy3 真实调用
- Fix-C：前端禁止 mock_agent/mock_quality/fallback=true 展示

## 修复内容

### Optimization fallback 字段补齐

- `types/agent.ts`：`OptimizationResultV2` 新增 `fallback`、`agent_status` 字段
- `api/viewModels.ts`：`OptimizationV2ViewModel` 新增 `fallback`、`agentStatus` 字段，`mapOptV2` 映射
- `ResumeOptimizationPage.tsx`：Deep 结果检测 `isMockGeneratedBy(genBy) || r.fallback || r.agentStatus === "fallback"`

## 验收结果

### 正常 Hy3 场景（8/8 通过）

| 功能 | Fast | Deep | generated_by | agent_status | fallback |
|---|---|---|---|---|---|
| 岗位推荐 | ✅ `fast_tag_engine` | ✅ | `llm_agent` | `repaired` | `false` |
| 岗位匹配 | ✅ `fast_tag_engine` | ✅ | `llm_agent` | `verified` | `false` |
| 简历优化 | ✅ `fast_tag_engine` | ✅ | `llm_agent` | — | `false` |
| 版本对比 | ✅ `fast_tag_engine` | ✅ | `llm_agent` | — | `false` |

### Deep Agent 日志确认

所有 4 条 Deep 链路均正确：
- Agent2 `reasoning_effort=high` ✅
- Agent3 `reasoning_effort=low` ✅
- `LLM-Hy3 reason_effort=high/low` ✅
- `verified=true` ✅

### 模型失败错误态

- 临时配置无效模型 → 后端返回 **HTTP 500**
- `LLM_FALLBACK_ENABLE=false` → 不回退到 mock
- `ENABLE_MOCK_AGENT=false` → 不走 mock path
- 日志确认：`hy3 http 400: invalid model` → 正确报错

### 构建

- npm build ✅
- go build ✅
- go vet ✅

## 未修改
- 后端源码
- 数据库
- 生产配置
- Git 操作

## 结论

Dev-Reboot 系列全部完成。开发环境：Fast=fast_tag_engine，Deep=llm_agent，模型失败=HTTP 500 错误，不再静默返回 mock。
