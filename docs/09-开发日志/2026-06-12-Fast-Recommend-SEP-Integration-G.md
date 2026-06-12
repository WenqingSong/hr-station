# Fast-Recommend-SEP-Integration-G — Fast 推荐正式接入 SEP 语义画像

## 目标
将 Fast 推荐主数据源从旧 raw tag profile 切换到 SEP semantic_profile_tags。

## 新增文件

| 文件 | 内容 |
|---|---|
| `service/fast_recommend_sep_ranker.go` | SEP 标签归一化 + 语义匹配打分 + 候选排序 (220 行) |

## 修改文件

| 文件 | 修改 |
|---|---|
| `service/recommendation_service.go` | +semanticGate/semanticRepo 字段 + fastRecommendWithSEP() + SEP Composer helpers + recommendCore 路由 |
| `cmd/server/main.go` | SEP 初始化 (SemanticProfileRepository→Service→Profiler→Gate) + NewRecommendationService 参数 |
| `service/recommendation_service_test.go` | NewRecommendationService 参数补齐 |

## SEP Ranker 规则

| 模块 | 规则 |
|---|---|
| tag normalize | `normalizeSemanticTagKey()` — 基于 code+label 的同义词映射 |
| role_direction 匹配 | +18 each, max +25 |
| technical_skill 匹配 | +5 each, max +20 |
| engineering_capability 匹配 | +6 each, max +18 |
| domain_context 匹配 | +3 each |
| experience_signal 匹配 | +4 each |
| risk penalty | -4 per unmatched job risk_signal, max -12 |
| base | 50 |
| clamp | 0-100 |
| tie-break | RoleMatch > TechMatch > EngMatch |

## 验证

- go build/vet/test ✅
- npm build ✅
- make check ✅
- Deep 链路未修改 ✅

## 下一步
进入 Deep-Recommend-Cascade-Pipeline-H。
