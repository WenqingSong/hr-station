# K8-5-26-A：Agent / Prompt / Schema 基于最终简历结构的再设计盘点

## 本阶段性质

只读盘点，不修改源码。评估最终 9 模块简历结构对 Agent / Prompt / Schema / Mock 的全链路影响。

---

## 最终简历结构对 Agent 的影响总表

| 变化 | 当前 Agent 是否适配 | 风险等级 |
|------|:---:|:---:|
| `target_direction=""` / `target_role=""` | ❌ 不适配 | P0 — mock 依赖 target_role 做推荐评分 |
| `job_intention` 不发送 | ⚠️ 部分不适配 | P1 — mock 引用了 `job_intention` module |
| skills `development_languages` 新字段 | ❌ 未识别 | P0 — Snapshot 只读 `languages`/`frameworks` 旧 key |
| skills `ai_application` 新字段 | ❌ 未识别 | P1 — Snapshot 只读 `ai_tools`/`ai_scenarios` 旧 key |
| skills 无 `language_exam` | ⚠️ 降级 | P2 — 旧 normalizer 可能尝试读取 |
| `certificates type=language_exam` | ❌ 未识别 | P1 — Snapshot 不区分 type |
| achievements `string[]` | ✅ 已适配 | 无 — agent_snapshot 用 toStrSlice |
| portfolio `works + homepage` | ⚠️ 部分不适配 | P1 — Snapshot 存旧 github/gitee 结构 |
| 通用岗位适配 | ❌ 不适配 | P2 — Mock 仅技术岗场景 |

---

## target_direction / job_intention 影响分析

| 项 | 现状 | 结论 |
|----|------|------|
| agent_snapshot.go | L54 仍读 `target_direction`；L60 仍序列化 `job_intention` | 为空不崩溃，但生成空 JSON 块 |
| recommendation_mock.go | L250 匹配 `target_role`；L333 输出 `target_role_mismatch` | **P0** — 空值导致无法评分 |
| match_mock.go | L267/274 使用 `ResumeModule: "job_intention"` | **P1** — 移除后需更新 module 引用 |
| version_compare_mock.go | L157 `ModuleName: "job_intention"` | **P1** — 同上 |
| match_analysis_v1 prompt | "target_role_mismatch" 作为 optimization_stop_signal | **P1** — 空 target_role 时此信号不适用 |
| 推荐 Agent | L125 读 `input.TargetRoles` | 空值不阻塞，推荐基于 Snapshot 全量数据 |

---

## skills 新结构影响分析

| 项 | 现状 | 结论 |
|----|------|------|
| agent_snapshot.go L266 | 读 `sk["languages"]` → ProgrammingLanguages | **P0** — 不读 `development_languages` |
| agent_snapshot.go L270 | 读 `sk["ai_scenarios"]` → AIApplicationScenarios | **P1** — 不读 `ai_application` |
| resume_snapshot_normalizer.go | 同 agent_snapshot 逻辑 | **P0** — 同上 |
| resume_snapshot.go model | 10 分类字段（无 development_languages/ai_application） | **P0** — 需扩展 |
| Agent 实际调用 | 使用 Snapshot JSON 输入给 LLM | 新字段缺失 → LLM 看不到 |
| Mock skills | 覆盖 programming_languages/other_keywords | **P1** — 需覆盖新字段 |

---

## language_exam 迁移影响分析

| 项 | 现状 | 结论 |
|----|------|------|
| agent_snapshot.go L274-281 | 读 `CertificatesJSON` → 全量解析，不区分 type | **P1** — 不区分 language_exam |
| Normalizer | 同 agent_snapshot | 同上 |
| Mock certificates | 覆盖 name/type/level/score | **P2** — 需添加 language_exam 场景 |
| skills 不再有 language_exam | ✅ 减少噪音 | 无需回退兼容 |

---

## achievements / portfolio 影响分析

| 项 | 现状 | 结论 |
|----|------|------|
| achievements string[] | agent_snapshot L145/158 用 toStrSlice | ✅ 已适配 |
| portfolio works | Snapshot L297 只读 github/gitee/homepage 等旧字段 | **P1** — 不读 works |
| portfolio homepage | ✅ Snapshot 已有 homepage 字段 | 适中 |

---

## 四核心功能评估

| 功能 | 当前实现 | 问题 | 建议 |
|------|---------|------|------|
| 岗位推荐 | recommendation_agent → buildResumeSnapshotJSON → LLM | target_role 为空时 Mock 评分逻辑失效；LLM 实际调用依赖 Snapshot 全量数据 | **移除** target_role 评分权重；完全基于 SkillSnapshot |
| 岗位匹配 | match_agent → LLM 逐条 analysis | `target_role_mismatch` 信号已过时 | 移除该信号；JD 匹配不依赖简历 target_role |
| 简历优化 | resume_optimization_agent → dual-path | 新 skills/achievements 未进入 Prompt | 更新 Prompt 引用新字段 |
| 版本对比 | version_compare_agent → Snapshot diff | job_intention 字段需移除对比 | 移除 module_name="job_intention" |

---

## Prompt / Schema / Mock 缺口

| 层级 | 缺口 | 建议阶段 |
|------|------|:---:|
| Snapshot model | 缺少 development_languages/ai_application | K8-5-26-B |
| Snapshot model | portfolio 缺少 works[] | K8-5-26-B |
| Normalizer (service) | 不识别新 skills key | K8-5-26-B |
| Normalizer (agent) | 同上 | K8-5-26-B |
| Prompt (match) | target_role_mismatch 残留 | K8-5-26-C |
| Prompt (recommendation) | 未引用专业技能/AI 应用 | K8-5-26-C |
| Prompt (optimization) | 未逐条分析 achievements | K8-5-26-C |
| Schema JSON | evidence 字段已有，无需改 | ⬜ |
| Mock (recommendation) | target_role 评分过时 | K8-5-26-E |
| Mock (match) | job_intention module 引用 | K8-5-26-E |
| Mock (optimization) | 旧 10 分类 skills 残留 | K8-5-26-E |
| Mock (compare) | job_intention module 引用 | K8-5-26-E |

---

## 是否需要修改后端 / SQL / Agent

| 层级 | 是否需要改 | 原因 |
|------|:---:|------|
| SQL | ❌ | 大字段 JSON 列 |
| 后端 Resume API | ❌ | RawMessage 兼容 |
| ResumeSnapshot model | ✅ | 新增 development_languages/ai_application/works |
| Snapshot Normalizer (service) | ✅ | 适配新 skills key 映射 |
| agent_snapshot.go | ✅ | 同上 + portfolio works |
| Agent Prompt | ✅ | 更新字段引用 |
| Response Schema JSON | ❌ | 已有 evidence 字段 |
| Mock Agent | ✅ | 覆盖新字段 |
| 前端 AI 页面 | ⬜ | evidence 展示待后续增强 |

---

## 后续阶段拆分建议

| 阶段 | 目标 |
|:---:|------|
| K8-5-26-B | ResumeSnapshot Normalize 适配（model + service + agent_snapshot） |
| K8-5-26-C | 推荐/匹配/优化/对比 Prompt 重写 |
| K8-5-26-D | Response Schema evidence 模块标注修正（如需） |
| K8-5-26-E | Mock Agent 最终 9 模块场景补齐 |
| K8-5-26-F | 四核心功能 Agent 单元测试回归 |
| K8-5-26-Z | Agent 最终验收 |

---

## 构建 / API / 测试结果

| 验收项 | 结果 |
|--------|:---:|
| `go build ./...` | ✅ |
| `go vet ./...` | ✅ |
| `go test ./...` | ✅ 全部通过 |
| API 200/200/200/401 | ✅ |

---

## 未修改范围确认

```text
未修改源码 ✅
未修改前端页面 ✅
未修改后端 ✅
未修改数据库 SQL ✅
未修改 Agent / Prompt / Schema / Mock ✅
未修改 API 路径 ✅
未新增 /api/resumes/v2 ✅
未修改登录鉴权 ✅
未修改 jobs 页面 ✅
未修改 .env ✅
未安装依赖 ✅
未迁移目录 ✅
未新增上传能力 ✅
未执行真实写入 ✅
```

---

## 最终结论

```text
K8-5-26-A 可以收口 ✅

核心发现：
  P0（2项）：Snapshot 不识别 development_languages / ai_application；推荐 Mock target_role 空值评分失效
  P1（6项）：Mock 残留 job_intention 引用；Prompt target_role_mismatch 过时；
           certificates 不区分 language_exam；portfolio 不识别 works；
           四功能 Prompt 未引用新字段
  P2（3项）：Mock 偏技术岗；通用岗位适配未进 Agent

建议执行顺序：26-B (Snapshot) → 26-C (Prompt) → 26-E (Mock) → 26-F (回归) → 26-Z (验收)
```

## 未执行 Git 命令

本次未执行任何 Git 命令，也未提交代码。
