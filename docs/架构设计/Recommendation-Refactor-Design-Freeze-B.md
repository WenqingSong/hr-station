# 岗位推荐 Agent 全链路重构 — 设计冻结

> 版本：Design-Freeze-B
> 日期：2026-06-12
> 状态：冻结（后续实现 prompt 以本文为准）

---

## 一、背景

Offer捕手当前岗位推荐功能存在以下结构性问题：

1. **Fast 推荐**：只有 Ranker 无 Explainer，raw tag code 直出给用户（`research`、`redis`、`algorithm`、`degree_bachelor`），已通过 `fast_recommend_explainer.go` + `tagLabelMap.ts` 部分修复，但匹配/优化同链仍未修复。
2. **Deep 推荐**：沿用旧 3-Agent 架构（Agent1 rule → Agent2 LLM → Agent3 Verifier），**非岗位级深度分析**，输出偏泛化，无法支撑"学生判断投递价值"的核心目标。
3. **Profile 层**：`resume_profiles`/`job_profiles` 仅规则提取 raw tag，无 `content_hash`/`model_name`，质量不可验证。
4. **数据库**：23 张表，profile 表缺关键字段，结果表存 JSON 快照，缺乏 Semantic Profile 独立表。

本次重构目标：**将岗位推荐从"规则排序+泛化 LLM"升级为"语义证据驱动+级联深度分析"**。

---

## 二、DB-State-Snapshot-A 基线摘要

| 项目 | 值 |
|---|---|
| 表数 | 23 |
| resumes | 8 |
| jobs | 46 |
| resume_profiles | 16（8 fast ready + 2 quality ready + 6 quality stale） |
| job_profiles | 46（全 rule 来源） |
| recommendation_results | 43 |
| content_hash | ❌ 均缺失 |
| model_name | ❌ 均缺失 |
| reasoning_effort | ❌ 均缺失 |
| semantic_profiles | ❌ 不存在 |
| semantic_profile_tags | ❌ 不存在 |
| sql-add 目录 | `/workspace/resources/sql-add`（空，待写入） |

> **所有后续增量 SQL 必须落盘 `/workspace/resources/sql-add`**，不允许仅改 `init.sql`，不允许代码自动建表。

---

## 三、当前推荐链路现状

| 项目 | 当前状态 | 问题 | 处理策略 |
|---|---|---|---|
| Fast 推荐 | `fastRecommendWithTags()` + `fast_recommend_explainer` | 推荐已 humanize ✅，但匹配/优化仍有 raw code | 泛化 explainer → GQCL Fast Composer |
| Deep 推荐 | `recommendWith3Agents()` 旧 3-Agent | 非岗位级分析，输出泛化 | 废弃 → 级联式 Deep Pipeline |
| 前端推荐页 | `recommendationAdapter.ts` + humanize 防御 | 推荐 OK，匹配/优化仍裸 | 前端 Safety Adapter 统一 |
| recommendation_results | JSON 快照存储 | 无 run/item 结构 | 第一阶段兼容，第二阶段按需升级 |

---

## 四、SEP 职责冻结

### 4.1 定位

**SEP（Semantic Evidence Profiling Layer）** 是推荐/匹配/优化/对比的**统一语义画像输入底座**，不直接参与最终展示/排序。

### 4.2 职责

| 类型 | 内容 |
|---|---|
| Resume SEP | 简历语义摘要、核心能力标签（label/code/confidence/evidence）、风险缺口、适配方向、profile_status、content_hash |
| Job SEP | 岗位方向/技能/要求/责任标签、风险信号、JD 摘要 |
| SemanticProfileGate | 核心功能入口 check profile 存在性 + 新鲜度（content_hash 比对） |

### 4.3 与旧 profile 兼容

- 旧 `resume_profiles` / `job_profiles` **保留但冻结写入**
- 新 SEP 数据写入 `semantic_profiles` + `semantic_profile_tags`
- Fast 路径优先读 SEP，fallback 读旧 profile
- Deep 路径强制读 SEP

---

## 五、GQCL 职责冻结

### 5.1 定位

**GQCL（Generation Quality Control Layer）** 是推荐输出的统一质量门禁。

### 5.2 职责

| 组件 | 职责 |
|---|---|
| **Fast Output Composer** | 将 Fast Ranker 的 raw tag 结果 + SEP labels 组合为产品化输出（句子/高亮/风险/建议） |
| **Deep Quality Gate** | 验证 Deep 每层输出满足字段契约、可读性、证据关联性 |
| **raw code 禁止** | 任何用户可见输出不得含 `research`/`redis`/`algorithm`/`degree_bachelor`/`skill_0` 等 code |
| **前端 Safety Adapter** | `tagLabelMap.ts` 泛化 → 所有 adapter 统一 humanize 防御 |

### 5.3 禁止输出清单

```
research | redis | algorithm | degree_bachelor | internship_exp | skill_0 | skill_1
标签匹配： | 缺失标签： | mock_agent | fallback=true | undefined | null
单个英文单词 bullet
```

---

## 六、Fast 推荐目标架构

### 6.1 链路

```
SemanticProfileGate → Resume SEP + Job SEP → Fast Recall/Ranker → Fast Output Composer → GQCL Check → Frontend ViewModel
```

### 6.2 输入

- Resume SEP（tag/risk/fit direction）
- 全量 Job SEP（tag 索引）

### 6.3 输出

```json
{
  "items": [{
    "job_id": 10001,
    "score": 88,
    "level": "优先投递",
    "one_sentence_reason": "你的简历与该岗位在 Redis 工程能力、算法能力上重合度较高，适合优先投递。",
    "match_highlights": ["简历与岗位在 Redis 方向存在匹配", "算法基础与岗位要求重合"],
    "risk_summaries": ["岗位要求本科及以上学历，请确认学历信息清晰"],
    "next_action": "建议进入岗位详情确认要求，再做匹配分析",
    "display_tags": ["后端开发", "Redis", "算法"],
    "debug_tags": { "matched": ["redis"], "missing": ["degree_bachelor"] }
  }],
  "generation_tier": "fast",
  "generated_by": "fast_tag_engine"
}
```

### 6.4 原则

- ❌ 不调用 LLM
- ✅ 保留速度优势（< 1s）
- ✅ 复用现有 `fast_recommend_explainer` 映射表
- ✅ 所有输出经 GQCL 检查

---

## 七、Deep 推荐目标架构

### 7.1 架构名称

**Cascaded Deep Recommendation Pipeline（级联式深度岗位推荐管线）**

### 7.2 四层架构

```
Layer 0：SemanticProfileGate
Layer 1：Candidate Discovery Agent（LLM ×1，high）
Layer 2：Semantic Rerank Agent（LLM ×1，high）
Layer 3：Per-Job Deep Synthesis Agent（LLM ×5，high）
Layer 4：Cross-Job Calibration / GQCL Gate（LLM ×1，low）
```

**总计 LLM 调用：8 次（1+1+5+1）**

### 7.3 各层详情

#### Layer 1：Candidate Discovery Agent

| 项目 | 内容 |
|---|---|
| 输入 | 简历 SEP + 简历摘要 + 岗位池摘要 + 岗位 SEP 摘要 |
| 输出 | 10-15 候选岗位 + 候选理由 + 初步风险 |
| Hy3 | `reasoning_effort=high` |

#### Layer 2：Semantic Rerank Agent

| 项目 | 内容 |
|---|---|
| 输入 | L1 候选 + 简历 SEP + 候选岗位 SEP |
| 输出 | Top 5 + 排序理由 + 排除理由 + 关键证据 + 主要风险 |
| Hy3 | `reasoning_effort=high` |

#### Layer 3：Per-Job Deep Synthesis Agent（核心）

| 项目 | 内容 |
|---|---|
| 输入 | 完整简历 + 简历 SEP + 完整岗位 JD + 岗位 SEP + L2 排序理由 |
| 输出 | 推荐结论 + 为什么推荐 + 简历证据 + 岗位要求对应 + 风险与影响 + 投递策略 + 简历优化提示 + display_tags |
| 调用次数 | 5 次（Top 5 每个岗位 1 次） |
| Hy3 | `reasoning_effort=high` |
| 超时 | 单次 120s，总时长约 10min（可并行） |

#### Layer 4：Cross-Job Calibration / GQCL Gate

| 项目 | 内容 |
|---|---|
| 输入 | 5 个岗位级输出 + L2 排序 + 简历 SEP |
| 输出 | 最终排序 + 全局总结 + 质量问题 + 是否通过展示 |
| Hy3 | `reasoning_effort=low`（可选，GQCL rule 可替代） |

### 7.4 旧 Agent3 Verifier 处理意见

**废弃**。理由：
1. 旧 Agent3 仅验证 JSON schema，与 GQCL 职责重叠
2. 新 GQCL Gate（Layer 4）可取代 JSON 校验 + 语义质量检查
3. 保存旧 Agent3 会阻碍岗位级 5 次深度生成

### 7.5 原理

- 不盲目沿用旧 3-Agent
- 岗位级分析是 Deep 推荐核心价值
- Token 不作为核心限制（质量优先）

---

## 八、数据库与迁移原则

| 项目 | 决策 |
|---|---|
| 基线 | `docs/数据库状态快照/DB-State-Snapshot-A.md` |
| SQL 落盘 | `/workspace/resources/sql-add/YYYYMMDD_HHMM_sep_xxx.sql` |
| 本阶段是否创建 SQL | ❌ 不创建（仅冻结设计） |
| semantic_profiles | 新表（统一简历/岗位语义画像） |
| semantic_profile_tags | 新表（标签详情 + evidence + confidence） |
| deep_recommendation_runs/items | 未来可选（第二阶段的 run/item 结构） |
| 旧表兼容 | `resume_profiles`/`job_profiles` 保留，冻结写入 |
| recommendation_results | 第一阶段继续使用，第二阶段按需升级 |
| 上线前 | 必须补充线上库 snapshot |

---

## 九、API 兼容原则

| 项目 | 决策 |
|---|---|
| 路径 | `POST /api/jobs/recommend` 不变 |
| generation_tier | `fast` / `quality` → `fast` / `deep` 保留 |
| 新增字段 | ✅ 允许：`items[]`、`display_tags`、`global_summary`、`quality_status` |
| 删除字段 | ❌ 不删除旧字段（向后兼容） |
| 前端兼容 | `recommendationAdapter.ts` 兼容新旧字段，tagLabelMap 兜底 |

---

## 十、后续实现阶段拆分

| 阶段 | 名称 | 目标 | 数据库风险 |
|---|---|---|---|
| **C** | SEP-Recommend-DataLayer-C | 设计 `semantic_profiles` + `semantic_profile_tags` 表结构，生成增量 SQL（不执行） | 低（仅设计） |
| **D** | SEP-Profiler-Implement-D | 实现 Hy3 Resume/Job Profiler，写入 SEP 表 | 新增 2 张表 |
| **E** | SemanticProfileGate-Implement-E | 实现核心功能入口的 profile 存在+新鲜度检查 | 低 |
| **F** | GQCL-Fast-Composer-F | 泛化 explainer → Fast Output Composer，统一 raw code 禁止规则 | 无 |
| **G** | Fast-Recommend-SEP-Integration-G | Fast 推荐接入 SEP + Fast Composer | 无 |
| **H** | Deep-Recommend-Cascade-Pipeline-H | 实现级联式 4 层 Deep Pipeline | 无（复用现表） |
| **I** | Frontend-Recommend-UX-I | 前端适配新 response、tagLabelMap 泛化 | 无 |
| **J** | Recommendation-Acceptance-J | 端到端验收 Fast+Deep 推荐质量 | 低 |

---

## 十一、风险清单

| 等级 | 风险 | 说明 | 建议 |
|---|---|---|---|
| **P0** | 旧 3-Agent 废弃影响 Deep 匹配/优化 | 匹配/优化也使用旧 Agent pipeline | 先重构推荐，其他功能后续跟进 |
| **P1** | Deep Pipeline Layer 3 5 次 LLM 耗时 10min | 超用户等待预期 | 并行调用 + loading 态告知进度 |
| **P1** | SEP 新表与旧 profile 并存的数据一致性 | 迁移阶段可能出现数据不一致 | 写时双写，读时优先 SEP |
| **P2** | 线上库当前无 snapshot | 无法验证线上迁移动态 | 上线前补充 snapshot |

---

## 十二、文档更新

| 操作 | 文件 |
|---|---|
| 新增 | `docs/架构设计/Recommendation-Refactor-Design-Freeze-B.md` |
| 将更新 | `docs/变更记录.md` |
| 将更新 | `docs/下一步任务.md` |
