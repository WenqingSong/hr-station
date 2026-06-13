# Recommendation-Page-Lifecycle-Request-Audit-H8：推荐页生命周期与请求链路审计

## 审计目标

用户首次进入 `/recommend?resumeId=4` 的完整请求链路，确认 30s loading 来源。

---

## 1. 环境与数据状态

| 项目 | 结果 |
|------|:---:|
| MySQL/Redis | ✅ healthy |
| backend healthz | ✅ 200 |
| resume_id=4 | 存在，user_id=2，is_current=1 |
| resume_id=4 Fast 历史 | ❌ 无（仅 mock_agent，被过滤） |
| resume_id=4 Quality 历史 | ❌ 无 |

---

## 2. 页面 Mount → 完成请求链路

**URL: `/recommend?resumeId=4`**

| 步骤 | 函数 | 请求 | 耗时 | 结果 |
|------|------|------|:---:|------|
| 1 | L2 redirect | 无（URL 已有 resumeId） | 0 | parsedUrlResumeId=4 |
| 2 | loadResumeOptions | GET /resumes | ~50ms | selectedOption=resume4 |
| 3 | quality-first init | **GET /latest?resume_id=4&tier=quality** | ~50ms | has_result=false |
| 4 | fallback fast | **GET /latest?resume_id=4&tier=fast** | ~50ms | has_result=false |
| 5 | generateFast | **POST /recommend?tier=fast** | 43ms | 5 items |
| **总计** | | | **~200ms** | 展示 Fast Top5 |

**首次进入无历史的耗时约 200ms，不是 30s。**

---

## 3. API 直测

| resumeId | latest fast | latest quality | generate fast |
|:---:|:---:|:---:|:---:|
| 4 | has=false | has=false | 43ms, 5 items |
| 1 | has=false | has=true, 5 items | 45ms, 5 items |
| 8 | has=false | has=true, 5 items | 47ms, 5 items |

---

## 4. 30s 来源判断

**不是 Auto-Init 生命周期问题。是 Deep CTA 按钮混淆。**

| 场景 | 实际耗时 | 用户感知 |
|------|:---:|------|
| 首次进入（auto-init） | ~200ms | 瞬间 |
| 页面 Fast 模式下点击蓝色按钮 | Deep Cascade 20-90s | "30s 转圈" |

页面在 fast_done 状态下的唯一蓝底 CTA 是"生成深度推荐"（调用 `handleDeepRecommendation`）。用户看到页面上写着"快速推荐"，但唯一可点的蓝按钮是"生成深度推荐"——点击触发 Deep Cascade 全链路（LLM + 5 阶段动画 + 轮询）。

**这不是 Fast 慢，也不是生命周期错误——是按钮语义与页面状态标签不匹配。**

---

## 5. 当前生命周期 vs 正确生命周期

| 场景 | 当前实现 | 是否符合 Fast-first |
|------|------|:---:|
| 无历史 | quality not found → fast not found → generateFast | ✅ Fast 最终展示 |
| 有 Fast 历史 | fast found → 展示 Fast | ✅ |
| 有 Deep 历史（完整5条） | quality found → 展示 Deep | ⚠️ 优先展示 Deep（与 Fast-first 理念矛盾） |
| Fast+Deep 都有 | quality found → 展示 Deep | ⚠️ Deep 默认优先 |
| 点击 Deep | handleDeepRecommendation → Deep Cascade | ✅ |
| Deep 成功 | 展示 Deep | ✅ |
| Deep 失败 | 展示 Fast | ✅ |

**核心问题**：当前是 **quality-first**（先查 quality，有就用），而不是 **Fast-first**（默认展示 Fast，Deep 需手动触发）。

---

## 6. 设计判断

| 问题 | 回答 |
|------|------|
| 30s 是 Fast 慢吗？ | ❌ Fast API 43ms，页面初始化 ~200ms |
| 生命周期是否正确？ | ⚠️ quality-first 有效但语义错位 |
| 应该加 Fast 按钮吗？ | ✅ 是——避免 Deep CTA 成为唯一入口 |
| 应该改 quality-first 吗？ | ⚠️ 可选——改为 Fast-first（不优先查 quality）或保持现状但加按钮 |

---

## 7. 构建

| 项目 | 结果 |
|------|:---:|
| npm build | ✅ PASS |
| go build/vet | ✅ PASS |

---

## 8. H9 修复建议

| 优先级 | 修复项 | 说明 |
|:----:|------|------|
| P1 | Fast 模式增加"重新生成快速推荐"按钮 | 解决 Fast 模式下唯一 CTA 是 Deep 的问题 |
| P1 | 首次进入 Fast-first（不优先查 quality） | 可选：符合用户预期 |
| P2 | loading 文案与真实请求一致 | "正在生成快速推荐"已正确 |

---

## 9. 给用户的结论

Fast 推荐不是 30s，后端 43ms，页面初始化约 200ms。用户感知的 30s 来自误点"生成深度推荐"蓝按钮（触发 Deep Cascade 20-90s）。当前生命周期是 quality-first（有 Deep 历史时优先展示 Deep），而非用户期望的 Fast-first（默认 Fast）。但这不是耗时原因——耗时原因是 Deep CTA 在 Fast 模式下成为唯一可点按钮。**H9：增加 Fast 重新生成按钮 + 考虑 Fast-first 策略回调。**
