# K6-4-F：Optimization Core 前端入口补齐

- **日期**：2026-06-02
- **阶段**：K6-4-F（feat, pass）
- **前置**：K6-4 (commit `32a7dff`)

## 1. 目标

在 `/resume-optimization` 页面补齐 Optimization Core 模式入口，让用户可以从前端触发 `mode=core` 链路。

## 2. 修改文件

| 文件 | 修改内容 | 修改行数 |
|------|----------|----------|
| `web/src/api/agents.ts` | `ResumeOptimizationRequest` 新增 `mode?: "core"` 字段 | +2 |
| `web/src/app/pages/ResumeOptimizationPage.tsx` | 新增模式切换 UI + mode 状态 + 优化触发传参 | +53 / -5 |

## 3. 前端入口设计

- **页面路径**：`/resume-optimization`
- **模式选择**：页面 Header 下方新增 `标准优化` / `Core 优化` 双按钮切换
- **默认模式**：标准优化（保持旧行为）
- **Core 触发**：点击「Core 优化」→ 请求体携带 `mode: "core"`
- **API 调用**：`POST /api/resumes/optimize`，via `runResumeOptimization()` → `loadResumeOptimizationViewModel()`
- **模式切换**：切换后自动重新触发优化请求

### UI 说明

```
[标准优化] [Core 优化]
标准优化：使用完整简历优化链路，含关键词对齐、项目优化等详细建议
Core 优化：使用 Optimization Core Agent，输出最关键的 2-3 条优化建议
```

## 4. 请求体差异

| 模式 | 请求体 |
|------|--------|
| 标准优化 | `{"resume_text":"...","target_job":{...},"target_section":"project"}` |
| Core 优化 | `{"resume_text":"...","target_job":{...},"target_section":"project","mode":"core"}` |

## 5. 验证结果

### curl 验收

| 验收项 | 接口 | HTTP | code | generated_by | agent_status | fix_items |
|--------|------|------|------|-------------|-------------|-----------|
| 标准模式 | `POST /api/resumes/optimize` (no mode) | 200 | 0 | mock_agent | fallback | 2 |
| Core 模式 | `POST /api/resumes/optimize {"mode":"core"}` | 200 | 0 | llm_agent | success | 2 |

### 构建与测试

| 命令 | 结果 |
|------|------|
| `cd web && npm run build` | SUCCESS (1.59s, 1663 modules) |
| `go test ./internal/... -count=1` | ALL PASS (10 packages) |
| `LLM_ENABLE` | `true` ✅ |

## 6. 旧行为兼容

- 默认「标准优化」不传 `mode` → 后端走原始完整 Agent → 与 K6-3-R 兼容 ✅
- 旧接口行为未改变 ✅

## 7. 遗留边界

- Core 模式返回的 `long_term_improvement` 和 `jd_evidence_chain` 不在当前前端 ViewModel 展示范围内（adapter 已妥善处理空数组兜底）
- `top_fix_items` 字段映射存在后端 `ResumeOptimizationFixItem`（issue/why_it_matters/action）与前端 `TopFixItem`（title/priority/expected_impact）的类型差异，属于预存问题，非本次引入，页面不崩溃

## 8. 判定

**K6-4-F = pass。** Core 前端入口已补齐，标准模式兼容，Core 模式可通过页面触发。

## 9. 下一步建议

- K6-5：hy3-preview vs DeepSeek V4 Pro 最终模型对比（按 K6-0 计划）
- **不在此验收中擅自开启**
