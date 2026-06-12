# 2026-06-12 Recommendation-Fast-Direction-Ranking-Fix-M2B — 修复 Fast 推荐方向排序

## 任务性质
后端 Fast SEP Ranker 方向门控补丁（M2 的收口阶段）。

## 问题
M2 修复了 score 全 100 和 display_tags 空，但排序仍严重错误：Java 后端候选人 Top5 全为 PC 客户端/游戏引擎/算法/市场研究。

## 根因
SEP Ranker 的评分只按标签类别加权（技术+工程），没有方向门控。Java/MySQL/Spring Boot 等通用标签导致非后端岗位也能获得高分。

## Career Direction Gate 设计
- 新增 `careerDirection` 类型（12 个方向常量）
- `inferResumeCareerDirection()`: 从简历 SEP tags + evidence 推断主方向
- `inferJobCareerDirection()`: 从 job title/direction/category + job SEP tags 推断岗位方向
- `directionCompatibility()`: 返回 strong/related/weak/mismatch 级别

### 评分公式
- base: 25
- strong direction: +40
- related direction: +22
- weak direction: -10
- mismatch direction: -40
- tech: ×3 cap 15
- eng: ×3 cap 12
- exp: cap 8, domain: cap 5, risk: ×4 cap 12

### Backend→其他方向映射
- strong: fullstack, test, ai_app, operation
- related: data, frontend, algorithm
- mismatch: client, market, product

## 验证
- Top5: 测试开发 / AI应用开发 / 技术运营 / 全栈开发 / 测试与质量管理
- 全部为开发/技术岗位 ✅
- 零算法/客户端/市场/产品 ✅
- scores: [93,92,91,88,87] ✅
- raw code/mock/fallback: 0 ✅
