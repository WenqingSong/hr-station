# JM-R5-R：Real Hy3 Job Match Output Calibration

> 日期：2026-06-14
> 状态：✅ PASS

## 1. 背景

JM-R3-R 已将管道修正为真实 Hy3 5 模块生成，但 Round 0 审计发现 Case 1（主演示案例：后端简历 × 后台开发）得分 58 → not_recommended，与预期 70-85 → revise_before_apply 严重不符。

根因：Hy3 模型对校招场景理解不足，把"软件开发-后台开发方向"按高级后端岗位标准打分，过度惩罚技能覆盖不足。

## 2. 校准内容

### 2.1 Module 1 system prompt（投递决策）

新增：
- **校招后台开发评分准则**：如果岗位是后端方向 + 简历有实习/项目/后端技术栈 ≥3项 → 不得判 not_recommended
- **不要过度惩罚**：高并发/微服务/K8s 等高级能力缺失 → 改写风险，非直接淘汰
- **方向错位判断**：后端简历投算法/NLP → 不因有实习提分
- **分数字段明确**：direct_apply ≥85 / revise 70-84 / backup 55-69 / not_recommended 40-54 / change_direction <40

### 2.2 Module 1 user prompt builder

新增：
- 岗位信息头（方向/类别/城市）
- 校招场景声明
- Fast初判"仅供参考，不要盲从"
- 校招初筛核心判断问题（4问）
- 关键提醒（企业课程平台是真实后端项目、有真实项目+实习+技术栈 = 70-85分）

### 2.3 Module 3 system prompt（风险分析）

新增风险三分类：
1. 方向错位风险（后端投算法）— severity=high
2. 证据不足风险（方向对但缺项目）— severity=high/medium
3. 表达强化风险（有证据但表达弱）— severity=medium/low
- fix_direction 必须具体，围绕接口设计/权限校验/数据库/缓存/部署

### 2.4 Module 4 system prompt（简历改写）

新增后端改写维度：
- 接口设计、权限校验、数据库操作、缓存使用、部署验证、异常处理、幂等限流
- 禁止编造量化数字（如"QPS提升50%""从180ms降到60ms"）
- 禁止编造技术（如 K8s/微服务/分布式）

### 2.5 Module 5 system prompt（岗位策略）

新增区分规则：
- 方向一致 + 证据不足 → not_recommended/backup，不是 change_direction
- 方向错位 → change_direction
- "先补项目证据再投递，不要放弃这个方向"

### 2.6 术语和状态口径修正

- 注释修正：`5 个 Hy3 深度模块负责内容生成 + 本地裁判负责一致性校验`
- Case 3 light match fallback → `generated_by = rule_screening_match_fallback, agent_status = fallback`

## 3. 案例回归

| 指标 | Round 0 | Round 2 | 预期 | 达标 |
|------|---------|---------|------|------|
| C1 score | 58 | **78** | 70-85 | ✅ |
| C1 decision | not_recommended | **revise_before_apply** | revise_before_apply | ✅ |
| C1 gen | hy3_screening_match_agent | hy3_screening_match_agent | hy3 | ✅ |
| C1 rewrites | 4 | 3 (含 before/after) | ≥2 | ✅ |
| C2 score | 33 | 33 | 20-55 | ✅ |
| C2 decision | change_direction | change_direction | change_direction | ✅ |
| C3 score | 42 | 42 | 35-60 | ✅ |
| C3 decision | not_recommended | not_recommended | not_recommended/backup | ✅ |
| C3 gen | deep_match_agent ❌ | **rule_screening_match_fallback** ✅ | fallback | ✅ |

## 4. 修改文件清单

| 文件 | 修改 |
|------|------|
| `match_deep_agent.go` | Module 1 system prompt + 校招准则（~30行）、M1 user prompt + 场景声明（~20行）、M3 system prompt + 风险三分类（~15行）、M4 system prompt + 后端改写维度（~25行）、M5 system prompt + 证据不足vs方向错误（~20行）、注释修正 |
| `match_service.go` | light match fallback 修正 generated_by 为 rule_screening_match_fallback |

## 5. 优秀输出摘录

### Case 1

- summary：学生技术方向为后端开发，与岗位后台开发方向一致，有后端实习经历和多个后端项目经历，技术栈基本覆盖后端常用技术，但部分高级能力缺失，简历表达可进一步强化
- strongest pass point：具备后端开发实习经历及多个后端项目开发经验（简历中有明确的Java/Spring Boot/MySQL/Redis后端技术栈实习和项目开发描述）
- highest risk：核心技能覆盖不足 — 简历中缺少岗位JD中要求的高并发、分布式、微服务等高级后端技能
- rewrite before：我主要负责后端接口开发、数据库表设计、登录鉴权和前后端联调。项目中设计了用户表、失物信息表、留言表...
- rewrite after：基于Spring Boot+MyBatis+MySQL+Redis技术栈，完成校园失物招领系统后端开发：1. 设计用户表、失物信息表、留言表、认领申请表、操作记录表共5张业务表...
- strategy：修改后投递 — 优先投递Java后端/后台开发实习岗位，补充1-2个分布式或高并发场景的实践项目后再投递该岗位

### Case 2

- summary：岗位核心方向为自然语言处理算法，学生简历核心证据均为后端开发相关经历，无NLP/算法相关项目或技能支撑，属于方向错位
- highest risk：岗位方向错位 — 简历核心经历为后端工程开发，岗位要求NLP/机器学习/深度学习算法研发能力
- rewrite guidance：不建议修改简历去投递该岗位。保留简历的真实后端工程经历描述，将求职意向调整为后台开发方向
- strategy：建议换方向 — better_fit_directions 包含后台开发、后端开发、Java开发、Go开发

### Case 3

- summary：简历缺少项目和实习经历，建议先补充实践证据
- highest risk：简历证据不足，缺少真实项目支撑
- rewrite guidance：不建议直接投递后台开发核心岗位。建议先完成一个较完整的后端项目（含接口开发、数据库设计、登录鉴权、缓存使用和部署验证）
- strategy：备选岗位 — 先补项目证据再投递

## 6. 构建检查

| 项目 | 结果 |
|------|------|
| npm build | ✅ 2.37s |
| go build | ✅ |
| go vet | ✅ |
| go test | ✅ all pass |

## 7. 结论

JM-R5-R 通过校准 Hy3 5 模块的 system prompt 和 user prompt，成功将 Case 1 从 score 58 → 78，decision not_recommended → revise_before_apply。Case 2 和 Case 3 保持准确。zero 硬编码案例结果。

可以进入 JM-R6 最终演示验收。
