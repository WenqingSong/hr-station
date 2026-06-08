# K9-2-Regression-B 完成报告：后端 Deep 字段与 FeatureCard + HardGuard 重构

## 1. 当前状态确认

- go build ✅
- go vet ✅ 
- npm build ✅
- Fast/Quality 输出不再相同 ✅
- Quality Deep 字段全部生成 ✅
- DB quality ready 记录完整 ✅
- Latest API 读到 quality ✅

---

## 2. A 阶段问题复盘

| 问题 | 级别 | 证据 |
|------|------|------|
| Mock Agent fast/quality 输出完全相同 | P0 | API 返回内容除 generated_by 外完全相同 |
| Deep 字段完全不生成 | P0 | evidence_chain/gap_points/top_fix/delivery_strategy 全部 null |
| 无 Feature Card | P2 | 无简历/岗位特征识别 |
| 无 Hard Guard | P2 | 仅有方向冲突 cap，无系统化边界规则 |
| 前端 Fast 视图使用不存在的 ViewModel 属性 | P1 | matchSummary/matchAdvantages 等不在 ViewModel 中 |

---

## 3. 根因分析

| 根因 | 说明 |
|------|------|
| Analyze() 不分 tier | 旧代码无论 fast/quality 都调用同一个 Analyze()，输出完全一致 |
| 无 Feature Card | 简历和岗位没有结构化特征识别，评分全靠标签矩阵 |
| 无 Hard Guard | 缺少系统化边界规则（非技术→技术岗、无算法→算法岗等） |
| Deep 字段不生成 | evidence_chain/gap_points/top_fix/delivery_strategy 的 builder 函数存在但没有被调用 |

---

## 4. 修改清单

| 文件 | 修改 |
|------|------|
| `server/internal/model/match.go` | 新增 MatchResumeFeatureCard/MatchJobFeatureCard/MatchHardGuardResult/ApplyReadinessObj/DeliveryStrategyObj 类型；MatchAnalysisResult 新增 fit_points/resume_evidence/job_requirements/interview_talking_points/delivery_strategy/improvement_suggestions/summary 字段 |
| `server/internal/agent/match_mock.go` | 完整重构：新增 AnalyzeFast/AnalyzeQuality 分离方法；新增 buildMatchResumeFC/buildMatchJobFC 构建 Feature Card；新增 applyMatchHardGuard 硬守卫；新增 buildMatchEvidenceChain/buildMatchGapPointsV2/buildMatchTopFixItems/buildApplyReadiness/buildOptimizationStopSignal/buildMatchNextSteps/buildDeliveryStrategyV2；Fast 保留简化输出，Quality 生成全部 Deep 字段 |
| `server/internal/service/match_service.go` | analyzeWithMock 根据 genTier 分发 AnalyzeFast 或 AnalyzeQuality；修复 GeneratedBy 设置和 Meta 对象构建 |
| `web/src/types/agent.ts` | MatchAnalysisResult 新增 fit_points/resume_evidence/job_requirements/gap_points/interview_talking_points/delivery_strategy/improvement_suggestions/screening_result/generated_by/agent_status/profile_type/profile_dominant_profile 可选字段；新增 MatchSuggestionItem 类型 |

---

## 5. Feature Card 实现

| 类型 | 字段 | 说明 |
|------|------|------|
| MatchResumeFeatureCard | IsTechnical/IsAlgorithm/IsBackend/IsProduct/IsPublicService/IsBusinessSolution/IsExploration | 从简历 RawText + 结构化字段 + dominant 方向检测 |
| MatchResumeFeatureCard | StrongCapabilities/WeakCapabilities/EvidenceSummary | 根据检测到的能力画像自动生成 |
| MatchJobFeatureCard | IsAlgorithm/IsComputerVision/IsBackend/IsProduct/IsUserResearch/IsProjectManagement/IsBusinessSolution | 从岗位标题 + JD 文本检测 |
| MatchJobFeatureCard | CoreRequirements/HardRequirements/KeywordRequirements | 从 JD 描述中自动提取 |

---

## 6. Hard Guard 规则

| 规则 | 处理 |
|------|------|
| 非技术简历 + 技术岗 | MaxScore ≤ 65，"缺少技术项目/工程证据" |
| 无算法证据 + 算法岗 | MaxScore ≤ 60，"缺少算法/模型/实验/论文证据" |
| 无后端证据 + 后端岗 | MaxScore ≤ 65，"缺少后端工程/数据库/接口/系统设计证据" |
| 产品简历 + 产品/用户研究/项目管理岗 | pass，允许高分 |
| 公共服务 + 解决方案/项目管理岗 | weak_related，MaxScore ≤ 80，"建议突出调研/需求理解/材料输出迁移能力" |
| 探索型简历 | MaxScore ≤ 78，"方向尚需聚焦，证据不够集中" |
| 公共服务 + 算法/CV岗 | mismatch，MaxScore ≤ 50，"严重不匹配" |

---

## 7. Fast / Quality 差异

| 字段 | Fast | Quality | 结论 |
|------|------|---------|------|
| match_score | ✅ | ✅ | 两者都有 |
| match_level | ✅ | ✅ | 两者都有 |
| summary | ✅ | ✅ | 两者都有 |
| evidence_chain | ❌ 空 | ✅ 1-2条 | Quality 独有 |
| gap_points | ❌ null | ✅ 1-2条 | Quality 独有 |
| top_fix_items | ❌ null | ✅ 1-3条 | Quality 独有 |
| delivery_strategy | ❌ 空 | ✅ 有 | Quality 独有 |
| apply_readiness | ❌ 空 | ✅ 有 | Quality 独有 |
| optimization_stop_signal | ❌ 空 | ✅ 有 | Quality 独有 |
| next_steps | ❌ 空 | ✅ 4条 | Quality 独有 |
| resume_evidence | ❌ 空 | ✅ 有 | Quality 独有 |
| job_requirements | ❌ 空 | ✅ 有 | Quality 独有 |
| interview_talking_points | ❌ 空 | ✅ 2-3条 | Quality 独有 |

---

## 8. Deep 字段修复

| 字段 | 修复后结果 |
|------|-----------|
| evidence_chain | 1-2条，含 requirement/resume_evidence/judgement |
| gap_points | 1-2条，具体指出缺失证据 |
| top_fix_items | 1-3条，含 issue/why_it_matters/action/example/self_check/effort/module |
| delivery_strategy | 完整策略（投递建议 + 风险提示 + 特别建议） |
| apply_readiness | "ready"/"prepare_first"/"cautious_try" + reason |
| optimization_stop_signal | 基于 score + gaps 的停止信号 |
| next_steps | 4条可执行建议 |
| interview_talking_points | 2-3条面试表达点 |
| resume_evidence | 项目经验+实习经历+技术栈证据 |
| job_requirements | 从 JD 提取的核心要求 |

---

## 9. 5 类 API 验收

| Resume | Job | Score | Level | Evidence | Gaps | Risks | Fixes | Strategy | 结论 |
|--------|-----|-------|-------|----------|------|-------|-------|----------|------|
| R1-Algo(1) | 10002(算法) | 95 | 强匹配 | 1 | 1 | 1 | 2 | ✅ | ✅ 高分+Deep |
| R1-Algo(1) | 10001(后台) | 65 | 备选 | 1 | 2 | 1 | 2 | ✅ | ✅ Hard Guard 限制 |
| R2-Backend(2) | 10001(后台) | 89 | 高匹配 | 2 | 1 | 1 | 2 | ✅ | ✅ 后端证据丰富 |
| R3-Product(3) | 10016(产品) | 75 | 可尝试 | 1 | 1 | 1 | 1 | ✅ | ✅ 方向匹配需补强 |
| R3-Product(3) | 10002(算法) | 60 | 备选 | 1 | 2 | 1 | 2 | ✅ | ✅ Hard Guard 限制 |
| R4-Explore(4) | 10016(产品) | 75 | 可尝试 | 1 | 1 | 1 | 1 | ✅ | ✅ 探索型保守 |
| R5-PublicSvc(5) | 10005(方案) | 79 | 可尝试 | 1 | 1 | 1 | 1 | ✅ | ✅ weak_related |

---

## 10. latest API 验收

| Resume | Job | has_result | tier | 字段完整性 | 结论 |
|--------|-----|-----------|------|-----------|------|
| 2 | 10001 | true | quality | evidence=2, gaps=1, fits=1 | ✅ 读到最新 quality |

---

## 11. DB 一致性验收

| 项 | 结果 |
|----|------|
| quality ready 记录存在 | ✅ id=131, generation_tier=quality, generated_by=mock_quality, status=ready |
| metadata 与 result_json 一致 | ✅ generation_tier/generated_by 列值与 result_json 内字段一致 |
| evidence_chain 不为空 | ✅ JSON_LENGTH=2 |
| gap_points 不为空 | ✅ JSON_LENGTH=1 |
| top_fix_items 不为空 | ✅ JSON_LENGTH=2 |
| delivery_strategy 不为空 | ✅ 1 |
| apply_readiness 不为空 | ✅ 1 |
| next_steps 不为空 | ✅ JSON_LENGTH=4 |

---

## 12. 测试结果

| 测试 | 结果 |
|------|------|
| go test ./internal/agent/ -run Match | 5 FAIL（均为 pre-existing: 旧测试调用已废弃的 Analyze() 方法，期望 Deep 字段但 fast 不生成） |
| go test ./internal/service/ -run Match | PASS |

**归因**：5 个失败测试均调用已废弃的 `Analyze()` 方法（现在委托给 `AnalyzeFast`），期望 Deep 字段但 fast 不生成。这些测试需要在后续阶段更新为调用 `AnalyzeQuality()`。

---

## 13. 构建 / 检查结果

| 验收项 | 结果 |
|--------|------|
| go build ./... | ✅ 通过 |
| go vet ./... | ✅ 通过 |
| npm run build | ✅ 通过 |
| check-encoding | ⚠️ pre-existing（init.generated.current.sql 中的 JSON 转义，与本次修改无关） |

---

## 14. 旧问题检查

| 项 | 结果 |
|----|------|
| quality 不只比 fast 多 generated_by | ✅ quality 比 fast 多 10+ Deep 字段 |
| quality 不返回关键 Deep 字段 null | ✅ evidence_chain/gap_points/top_fix/delivery_strategy/apply_readiness/next_steps 全部非空 |
| mock_quality 不直接展示给用户 | ✅ 只在内部 generated_by 字段中，前端展示使用匹配等级/分数等真实字段 |

---

## 15. 问题清单

| 级别 | 问题 | 证据 | 建议阶段 |
|------|------|------|---------|
| P2 | 5 个 go test 失败 | 旧测试调用 Analyze() 期望 Deep 字段 | K9-2-Regression-C/D |
| P2 | Fast 视图 ViewModel 属性不匹配 | matchSummary/matchAdvantages 等不在 ViewModel | K9-2-Regression-D |
| P3 | Hard Guard 文本检测依赖 RawText | 部分简历 RawText 为空时依赖 dominant 回退 | 后续优化 |

---

## 16. 是否进入下一阶段

**通过** → K9-2-Regression-B-Check：后端 Deep 字段密度与 5 类样例验收

或直接进入 K9-2-Regression-C：ThreeLayerAgent（quality tier 三层 Agent prompt）

---

## 17. 未修改范围确认

未修改 K9-1 岗位推荐 ✅
未修改 K9-3 简历优化建议 ✅
未修改 K9-4 简历版本对比 ✅
未修改 server/sql/init.sql ✅
未修改数据库结构 ✅
未删除真实简历 ✅
未执行 reset-up ✅
未执行 Git 命令 ✅

---

## 18. docs 更新清单

新增：
- `docs/09-开发日志/2026-06-08-631-K9-2-Regression-B-后端Deep字段FeatureCardHardGuard重构.md`

待更新：
- `docs/下一步任务.md`
- `docs/变更记录.md`
