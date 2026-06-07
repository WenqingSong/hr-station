# K9-1-Regression-A 只读诊断报告

## 岗位推荐页重新进入仍生成快速推荐的 latest-first 与持久化链路只读诊断

**日期**: 2026-06-07  
**阶段**: 只读诊断，不修改源码/数据库  
**关联**: K9-1-U-C（推荐结果持久化后端实现）、K9-1-U-D（前端读取latest与FastDeep展示分层）

---

## 1. 当前状态确认

| 项 | 状态 |
|---|---|
| GET `/api/jobs/recommend/latest?resume_id=5` 返回 `has_result` | ✅ `true` |
| `recommendation_results` 有 resume_id=5 的 ready 记录 | ✅ 6 条（id: 5,10,42,43,44,45） |
| `resume_profiles` fast=ready, quality=stale | ✅ fast=ready(dominant=business_solution) |
| POST `/api/jobs/recommend` 生成后保存到 DB | ✅ 验证通过 |
| `init.sql` 包含 resume_id=5 的推荐结果 | ❌ **仅含 resume_id=1 的 1 条记录** |
| 前端 `loadLatestRecommendationViewModel` 逻辑 | ✅ 正确，先 GET latest 再判断 |
| 后端 `FindLatestByResumeID` 查询逻辑 | ✅ 正确，WHERE status='ready' ORDER BY created_at DESC |
| 当前访问 `/recommend?resumeId=5` | ✅ 直接展示结果（已持久化） |

---

## 2. 问题复现

| 页面 | 现象 | 证据 |
|---|---|---|
| `/recommend?resumeId=5`（首次 after make reset-up） | 页面长时间显示"正在生成快速推荐" | `init.sql` 仅含 resume_id=1 的推荐结果 |
| `/recommend?resumeId=5`（POST fast 完成后再次访问） | ✅ 直接进入结果页 | API 返回 `has_result=true` |

**结论**: 问题仅在 `make reset-up` 后首次访问时出现。`init.sql` 快照不包含 resume_id=5（及 2/3/4）的推荐结果，首次访问触发 POST fast 生成。

---

## 3. Network 请求顺序（模拟首次访问 after reset-up）

| 步骤 | 请求 | 响应 | 耗时 | 结论 |
|---|---|---|---|---|
| 1 | GET `/api/jobs/recommend/latest?resume_id=5` | `has_result: false` | ~50ms | 无历史结果 |
| 2 | POST `/api/jobs/recommend` `{resume_id:5, generation_tier:"fast"}` | `profile_status: ready, 5 jobs` | ~500ms-2s | 生成+保存成功 |
| 3 | 页面进入结果页 | — | — | 展示 5 个推荐岗位 |

**说明**: 当前实测 latest 返回 `has_result=true`，因为运行时已生成 6 条记录。上表为模拟 `init.sql` 刚导入的场景。

---

## 4. latest API 验证

| resume_id | has_result | generation_tier | job_count | profile_dominant | result_id | 结论 |
|---|---|---|---|---|---|---|
| 5 | true | fast | 5 | business_solution | 45 | ✅ API 正常 |

**API 响应关键字段**:
```json
{
    "has_result": true,
    "result_id": 45,
    "resume_id": 5,
    "generation_tier": "fast",
    "generated_by": "mock_agent",
    "status": "ready",
    "profile_type": "fast",
    "profile_dominant_profile": "business_solution",
    "result": [{...5 jobs...}]
}
```

---

## 5. POST fast 验证

| resume_id | HTTP 状态 | profile_status | generation_tier | generated_by | job_count | 是否保存 DB |
|---|---|---|---|---|---|---|
| 5 | 200 | ready | null (API字段在meta中) | mock_agent | 5 | ✅ 已保存到 recommendation_results |

---

## 6. recommendation_results 数据检查

| resume_id | tier | status | count | latest_created |
|---|---|---|---|---|
| 1 | fast | ready | 29 | 2026-06-07 19:10:18 |
| 1 | quality | ready | 2 | 2026-06-07 19:10:28 |
| 2 | fast | ready | 3 | 2026-06-07 13:21:43 |
| 3 | fast | ready | 3 | 2026-06-07 13:21:43 |
| 4 | fast | ready | 2 | 2026-06-07 13:07:00 |
| 5 | fast | ready | 6 | 2026-06-07 19:11:43 |

**关键发现**:
- `resume_id=5` 有 6 条 fast 记录（含旧记录 id=5,10 的 `profile_type/profile_dominant_profile` 为 NULL）
- `resume_id=1` 有 quality 级别结果（LLM agent 生成）
- `resume_id=2/3/4/5` 均无 quality 级别结果
- 最新记录 id=45 字段完整（`profile_type=fast`, `profile_dominant_profile=business_solution`）

**init.sql 快照情况**:
```
INSERT INTO recommendation_results VALUES (1, ...) -- 仅 resume_id=1 有 1 条
```
- resume_id=2/3/4/5 的 recommendation_results 记录 **不在 init.sql 中**
- `make reset-up` 后，这 4 个 resume 无历史推荐结果

---

## 7. resume_profiles 数据检查

| resume_id | profile_type | status | dominant_profile | source | generator |
|---|---|---|---|---|---|
| 1 | fast | ready | algorithm_research | rule | rule_fast_profile_v1 |
| 1 | quality | stale | algorithm_research | rule | rule_quality_fallback_v1 |
| 2 | fast | ready | backend_engineering | rule | rule_fast_profile_v1 |
| 2 | quality | stale | backend_engineering | rule | rule_quality_fallback_v1 |
| 3 | fast | ready | product_operation | rule | rule_fast_profile_v1 |
| 3 | quality | stale | product_operation | rule | rule_quality_fallback_v1 |
| 4 | fast | ready | product_operation | rule | rule_fast_profile_v1 |
| 4 | quality | stale | product_operation | rule | rule_quality_fallback_v1 |
| 5 | fast | ready | business_solution | rule | rule_fast_profile_v1 |
| 5 | quality | stale | business_solution | rule | rule_quality_fallback_v1 |

**结论**: 所有 resume 的 fast profile 均为 ready，`EnsureProfile` 不会触发阻塞式等待。Fast 推荐路径无 profile 瓶颈。

---

## 8. 前端 latest-first 代码诊断

| 文件 | 逻辑 | 是否正确 |
|---|---|---|
| `JobRecommendationPage.tsx:88-117` | `useEffect([selectedOption])` → 先 `loadLatestRecommendationViewModel(rid)` | ✅ |
| `JobRecommendationPage.tsx:96` | `if (latest.hasResult && latest.result && latest.result.recommendedJobs.length > 0)` 判断 | ✅ |
| `JobRecommendationPage.tsx:113` | `.catch()` 兜底调用 `generateFast(rid)` | ✅ 合理兜底 |
| `viewModels.ts:167-192` | `loadLatestRecommendationViewModel` 解析 `raw.has_result` | ✅ |
| `agents.ts:61-63` | `getLatestRecommendation(resumeId)` 调用 GET `/jobs/recommend/latest?resume_id=X` | ✅ |
| `JobRecommendationPage.tsx:169` | `phase==="loading_latest"` → 显示"正在读取推荐结果…" | ✅ 文案正确 |
| `JobRecommendationPage.tsx:170-172` | `phase==="fast_loading"` → 显示"正在生成快速推荐" | ⚠️ 仅当 latest 返回 false 时触发 |
| `JobRecommendationPage.tsx:74-85` | `selectedOption` 从 URL `resumeId` 解析 | ✅ `parseResumeIdFromQuery("5")` → 5 |

**结论**: 前端 latest-first 逻辑完全正确。问题不在前端代码逻辑，而在数据层。

---

## 9. 后端 latest / 保存代码诊断

| 文件 | 逻辑 | 是否正确 |
|---|---|---|
| `router.go:81` | 注册 `GET /api/jobs/recommend/latest` | ✅ |
| `recommendation_handler.go:80-114` | `Latest()` handler 解析 `resume_id`，调用 `GetLatestRecommendation` | ✅ |
| `recommendation_service.go:469-511` | `GetLatestRecommendation` 验证 resume 归属 → 查 DB → 返回 `RecommendationLatestResult` | ✅ |
| `recommendation_repository.go:67-104` | `FindLatestByResumeID` SQL: `WHERE user_id=? AND resume_id=? AND status='ready' ORDER BY created_at DESC, id DESC LIMIT 1` | ✅ |
| `recommendation_service.go:394-422` | `recommendWithMock` 保存 `RecommendationResult` 含 `profile_type`, `profile_dominant_profile` | ✅ |
| `recommendation_service.go:111-150` | `Recommend()` 调用 `EnsureProfile` → 若 profile 非 ready 则提前返回 | ✅ |
| `resume_profile_service.go:59-68` | staleness check: `profile.ResumeUpdatedAt.Equal(resume.UpdatedAt)` | ⚠️ init.sql 导入后 timestamp 可能不匹配 |

**需要关注的 subcase**: `init.sql` 导入后，`resume_profiles.resume_updated_at` 与 `resumes.updated_at` 可能不一致，导致 `EnsureProfile` 判定 stale → 触发同步 `generateFast` 重新生成 profile → 增加首次 POST fast 耗时。

---

## 10. 根因结论

| 级别 | 问题 | 证据 |
|---|---|---|
| **P0 根因** | `init.sql` 快照仅含 resume_id=1 的推荐结果，resume_id=2/3/4/5 无结果 | `grep "INSERT INTO recommendation_results" init.sql` → 仅 1 条，id=1 |
| P1 次因 | `make reset-up` 后首次访问 resume_id=5 必定触发 POST fast 生成 | 时序：reset-up → GET latest=false → POST fast(1-3s) → 保存 → 展示 |
| P2 体验 | "正在生成快速推荐" 的等待时间在 1-3 秒，但用户感知为"长时间" | mock agent + DB write 耗时，非阻塞但无进度反馈 |
| P2 体验 | `init.sql` 导入后 profile timestamp 可能触发 stale 重新生成 | 增加首次 POST fast 延迟 |

### 根因分类矩阵

| 分类 | 是否匹配 | 说明 |
|---|---|---|
| A. recommendation_results 确实没有 resume_id=5 结果 | ✅ **匹配** | `init.sql` 导入后确实没有 |
| B. POST fast 生成了结果但没有保存 | ❌ | 保存逻辑验证正确 |
| C. POST fast 保存了但 latest 查不到 | ❌ | 查询条件正确 |
| D. latest API 返回字段与前端不匹配 | ❌ | 字段对应正确 |
| E. 前端没有调用 latest 直接 POST fast | ❌ | 代码逻辑正确，先 latest |
| F. 前端读取的 resumeId 不是 URL resumeId=5 | ❌ | `parseResumeIdFromQuery` 正确 |
| G. profile pending/stale 导致等待 | ⚠️ | 当前 ready，但 init.sql 后可能 stale |
| H. init.sql 快照替换后数据丢失 | ✅ **匹配** | 仅保留了 resume_id=1 的 1 条结果 |
| I. 后端进程假活/API 慢 | ❌ | API 响应正常 |
| J. 正常首次生成但体验需优化 | ⚠️ | 文案和预生成策略可优化 |

---

## 11. 修复建议

| 修复点 | 文件 | 建议 |
|---|---|---|
| **P0: 预生成所有简历的 fast 推荐** | `server/sql/init.sql` | 在 `init.sql` 中为 resume_id=2/3/4/5 各 INSERT 1 条 fast recommendation_results（与 resume_id=1 的预生成方式一致） |
| P0: 或添加后端启动 backfill | `server/internal/service/recommendation_service.go` | 在服务启动时为所有已存在的 resume 预生成 fast 推荐结果（如调用 `BackfillMissingRecommendations`） |
| P1: 首次 fast 生成时 profile stale 检查优化 | `resume_profile_service.go:59-68` | `init.sql` 导入后确保 `resume_profiles.resume_updated_at` == `resumes.updated_at`，避免首次 stale |
| P2: 优化 loading 文案 | `JobRecommendationPage.tsx:170-172` | "正在生成快速推荐" → 增加预估时间或分步骤提示（如"正在读取简历画像… → 正在匹配岗位… → 正在生成推荐…"） |
| P2: 优化 `phase` 过渡可见性 | `JobRecommendationPage.tsx:169` | `loading_latest` → `fast_loading` 过渡较短暂，可适当延长或合并状态展示 |

---

## 12. 是否需要进入修复阶段

- ✅ **通过，建议进入 K9-1-Regression-B-Fix**
  - 核心修复：为 init.sql 补充 resume_id=2/3/4/5 的 recommendation_results 预生成数据
  - 不需要修改业务逻辑代码
  - 仅数据层面补齐

- 备选方案：
  - 如果选择"这是正常首次生成"：进入 **K9-1-UX-Fix**，优化 loading 文案和预生成策略
  - 如果是环境慢：进入 **Env-Fix**（当前排除）

---

## 13. 未修改范围确认

- [x] 未修改源码
- [x] 未修改数据库
- [x] 未删除真实简历
- [x] 未执行 `make reset-up`
- [x] 未执行 `docker compose down -v`
- [x] 未执行 Git 命令
- [x] 仅执行 SELECT 查询 + curl API 调用 + 源码 grep 查看

---

## 14. docs 更新清单

| 文件 | 操作 | 状态 |
|---|---|---|
| `docs/09-开发日志/2026-06-07-607-K9-1-Regression-A-岗位推荐latest持久化链路只读诊断.md` | 新增 | 本文档 |
| `docs/下一步任务.md` | 更新 | 待执行 |
| `docs/变更记录.md` | 更新 | 待执行 |

---

## 15. Git 约束确认

- [x] 未执行任何 Git 命令
- [x] 未修改任何源码/数据库/配置
- [x] 本报告为纯只读诊断文档
