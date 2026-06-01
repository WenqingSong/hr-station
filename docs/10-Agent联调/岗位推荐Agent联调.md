# 岗位推荐 Agent 联调

> 阶段：7-1 | 状态：pass | 日期：2026-06-01

## 1. 联调目标

打通岗位推荐 Agent（POST /api/jobs/recommend）前后端通路。

## 2. 后端推荐 Agent 接口核验

| 项目 | 状态 |
|------|:--:|
| 路径 | `POST /api/jobs/recommend` ✅ |
| 请求体 | `{ resume_version_no: int, limit: int }` ✅ |
| 响应 | `{ code, message, data: { result, meta } }` ✅ |
| 当前模式 | Mock Agent（llm_api_key_missing fallback） ✅ |
| Auth | JWT Bearer token ✅ |

## 3. curl 联调结果

| 接口 | HTTP | code | data.result | data.meta | 结论 |
|------|:--:|:--:|------|------|:--:|
| `POST /api/jobs/recommend` | 200 | 0 | `{ recommended_jobs: [3] }` | `agent_status: fallback` | ✅ pass |

返回 3 个推荐岗位（Go后端/云计算/AI应用），Mock Agent 生成，含 recommendation_score、match_highlights、risk_points、suggested_action 等完整字段。

## 4. 前端请求体对齐

| 字段 | 修改前 | 修改后 |
|------|------|------|
| `resume_id` | ✅ 传给后端 | ❌ 移除 |
| `resume_snapshot` | ✅ 传给后端 | ❌ 移除 |
| `resume_version_no` | ❌ 不存在 | ✅ `number`，默认 2 |

修改文件：`web/src/api/agents.ts`（RecommendationRequest 类型 + runRecommendation 默认值）。

## 5. resumeId → resume_version_no 映射

```ts
function resolveResumeVersionNo(resumeId: string | null): number {
  if (!resumeId) return 2;
  const matched = resumeId.match(/\d+/);
  return matched ? Number(matched[0]) : 2;
}
```
resume-v1→1, resume-v2→2, resume-v3→3, fallback→2。

## 6. 推荐页接入

- `JobRecommendationPage.tsx`：从直接 mock 改为 `loadRecommendationViewModel({ resume_version_no, limit })`
- loading 态：`"正在生成岗位推荐…"`
- fallback 态：`"推荐接口暂时不可用，已展示演示数据。"`

## 7. AgentResponseEnvelope 对齐

后端 `data: { result, meta }` → 前端 `client.post<T>()` 返回 `data` → `AgentResponseEnvelope<TResult>` ✅

## 8. mock fallback 保留

`runRecommendation` 中 `withMockFallback` 保留，失败 → mockRecommendationResponse ✅

## 9. 未处理

match / optimization Agent → 7-2 / 7-3

## 10. build 结果

```
✓ built in 1.92s
```
**通过 ✅**

## 11. URL 验证

`/recommend?resumeId=resume-v2` → **200** ✅

## 12. grep 验证

| 检查项 | 结果 |
|------|:--:|
| 推荐用 resume_version_no | ✅ |
| match/optimize 保留 resume_id/snapshot | ✅ (7-2/7-3) |
| 零 api/client 直连 | ✅ |
| 零 localStorage | ✅ |
| 零越界 | ✅ |

## 13. P0/P1/P2

| 等级 | 数量 | 说明 |
|:--:|:--:|------|
| P0 | 0 | — |
| P1 | 0 | 推荐 P1 已解除 |
| P2 | 0 | — |

## 14. 最终结论

**pass ✅** — 可进入 7-2 岗位匹配分析 Agent 联调。
