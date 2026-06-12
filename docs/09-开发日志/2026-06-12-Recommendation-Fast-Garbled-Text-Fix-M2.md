# 2026-06-12 Recommendation-Fast-Garbled-Text-Fix-M2 — 修复 Fast 推荐分数、排序、标签与风险展示

## 任务性质
后端 Fast SEP Ranker 修复（不改 Deep/数据库）。

## 问题背景
Fast API 返回全部 100 分、display_tags 为空、risk_points 模板化、排序不合理。

## 修复内容

### 1. 评分公式重构 (`fast_recommend_sep_ranker.go`)
- base: 50→30
- role_direction: 2+匹配 +35，1匹配 +22，0匹配 -18（原: ×18 cap 25）
- technical_skill: ×4 cap 18（原: ×5 cap 20）
- engineering_capability: ×4 cap 14（原: ×6 cap 18）
- domain_context: cap 6（不变）
- experience_signal: cap 8（不变）
- risk: ×5 cap 15（原: ×4 cap 12）

### 2. DisplayTags 修复
- 传递: `DisplayTags: score.DisplayTags` 行新增到 builder
- 截断: 去除 10 字符暴力截断，使用完整中文标签

### 3. risk_points 去模板
- 原: "岗位关注「XX」，建议补充相关证据"
- 新: "岗位要求关注「XX」，当前简历在该方面的体现较弱，可能会降低面试机会。"
- 方向不匹配时追加岗位提示

## 验证结果
- scores: [98, 93, 92, 91, 89] — 不再全 100，range=9
- display_tags: 每岗位 4 个中文标签
- raw code: 0
- go build/vet/test: 全部通过
