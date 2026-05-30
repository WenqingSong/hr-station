# 2026-05-31-113-三类Agent当前LLMFallback状态核验与验收矩阵

## 1. 本次任务

阶段 4-3.8.0：三类 Agent 当前 LLM / fallback 状态核验与验收矩阵冻结。

## 2. 背景

4-3.7-fix 系列已完成 Prompt / Schema / Go 契约一致性修复和最终闸门验收，4-3.7-fix.4 已完成核心 Agent 返回内容 Smoke 与固定测试简历沉淀。在正式进入 4-3.8 集成验收之前，必须先核验三类 Agent 当前的 LLM / fallback 真实状态。

## 3. 核验范围

- 三类 Agent 的 LLM Agent / Mock Agent 代码
- Service 层 LLM Pipeline 接入状态
- Pipeline fallback 行为
- generated_by / agent_status 赋值
- 当前测试覆盖
- 是否需要 4-3.8.1-R

## 4. 三类 Agent 状态

### MatchAnalysisAgent

- **状态**：LLM + fallback 已接入
- **LLM Agent**：`match_agent.go`，通过 PipelineRunner 调用 Pipeline
- **Mock Agent**：`match_mock.go`，规则打分
- **Service**：`match_service.go`，pipelineRunner + llmEnabled，LLM 优先 + Mock 回退
- **Fallback**：MatchFallbackError → Mock 回退
- **generated_by**：LLM 成功 → "llm_agent"，Mock → "mock_agent"
- **测试覆盖**：fake Pipeline success / fallback / error / service 多场景

### RecommendationAgent

- **状态**：LLM + fallback 已接入
- **LLM Agent**：`recommendation_agent.go`，通过 PipelineRunner 调用 Pipeline
- **Mock Agent**：`recommendation_mock.go`，规则打分
- **Service**：`recommendation_service.go`，pipelineRunner + llmEnabled，LLM 优先 + Mock 回退
- **Fallback**：RecFallbackError → Mock 回退
- **generated_by**：LLM 成功 → "llm_agent"，Mock → "mock_agent"
- **测试覆盖**：fake Pipeline success / fallback / error / candidate filter / service 多场景

### ResumeOptimizationAgent

- **状态**：Mock-only
- **LLM Agent**：不存在（无 `resume_optimization_agent.go`）
- **Mock Agent**：`resume_optimization_mock.go`，规则分析
- **Service**：`resume_optimization_service.go`，仅 mockAgent，无 pipelineRunner / llmEnabled
- **Fallback**：无（因为没有 LLM 路径）
- **generated_by**：硬编码 "mock_agent"
- **测试覆盖**：仅 Mock 正常返回

## 5. LLM / fallback 矩阵

| Agent | 当前状态 | LLM success | malformed JSON | LLM error | fallback | generated_by | 需要 8.1-R |
|-------|----------|-------------|----------------|-----------|----------|--------------|-----------|
| MatchAnalysis | LLM + fallback 已接入 | 已有 | 已有 | 已有 | 已有 | 已有 | 否 |
| Recommendation | LLM + fallback 已接入 | 已有 | 已有 | 已有 | 已有 | 已有 | 否 |
| ResumeOptimization | **Mock-only** | **缺失** | **缺失** | **缺失** | **缺失** | 仅 mock_agent | **是** |

## 6. 当前测试覆盖

### MatchAnalysisAgent
- ✅ fake LLM success / error / fallback / repair / service 多场景
- ✅ generated_by 断言
- ✅ apply_readiness / optimization_stop_signal / top_fix_items 断言

### RecommendationAgent
- ✅ fake LLM success / error / fallback / candidate filter / service 多场景
- ✅ generated_by 断言
- ✅ apply_readiness / optimization_stop_signal / top_fix_items 断言

### ResumeOptimizationAgent
- ✅ Mock 正常返回 / generated_by / apply_readiness / optimization_stop_signal / top_fix_items
- ❌ 无 LLM 测试
- ❌ 无 fallback 测试
- ❌ 无 malformed JSON / repair 测试

## 7. 是否需要 8.1-R

**需要插入 4-3.8.1-R：ResumeOptimization Pipeline 最小接入。**

理由：
1. ResumeOptimizationAgent 仍为 Mock-only，没有 LLM Agent
2. Pipeline 已支持 TargetResumeOptimization
3. Prompt / Schema / Go model 已就绪
4. Service 层需注入 pipelineRunner + llmEnabled

8.1-R 目标：
1. 创建 ResumeOptimizationAgent（LLM Agent）
2. 定义 Optimizer 接口
3. 保留 ResumeOptimizationMockAgent 作为 fallback
4. Service 层 LLM 优先 + Mock 回退
5. 不进前端，不改页面，不改启动链路

## 8. 未修改内容

- 未修改 Go 代码
- 未修改 Prompt
- 未修改 Schema
- 未修改前端
- 未修改 Makefile
- 未修改 docker-compose
- 未修改 Nginx
- 未读取 .env
- 未调用真实 LLM
- 未执行 reset / reset-up
- 未清空 volume

## 9. 下一步建议

1. **4-3.8.1-R：ResumeOptimization Pipeline 最小接入**
2. **4-3.8.1：三类 Agent LLM / fallback 本地集成测试补齐**
3. **4-3.8.2：三类 Agent 统一字段断言验收**
