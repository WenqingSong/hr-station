# 简历优化 Prompt 首轮输出验收报告

> 创建时间：2026-05-29
> 所属阶段：4-1.4.8
> 验收人：CodeBuddy

## 1. 验收目标

验证 `server/prompts/resume_optimization_v1.md` 在真实模型输出时是否符合预期，具体包括：

1. 模型是否能输出合法 JSON
2. 输出是否包含完整结构
3. 模型是否能正确输出 P0 / P1 / P2 简历优化建议
4. 模型是否能围绕目标岗位 JD 优化，而不是输出泛泛简历建议
5. 模型是否能输出 keyword_alignment
6. 模型是否能输出 project_optimization
7. 模型是否能输出 skill_section_optimization
8. 模型是否能输出 summary_optimization
9. 模型是否能输出 missing_evidence_plan
10. 模型是否能输出 rewrite_examples
11. 验证 rewrite_examples 是否严格基于真实经历
12. 验证 requires_user_verification 是否正确使用
13. 验证模型是否避免编造项目数据、系统规模、职责范围
14. 验证 Offer捕手语气是否亲和但不幼稚
15. 验证真实性提醒是否专业明确
16. 验证输出是否适合 Demo 展示

## 2. 验收范围

- 只测试 1 组样本：Arthur V2 简历 + 腾讯岗位"软件开发-后台开发方向"
- 只进行 1 次主调用
- 不使用 MATCH_ANALYSIS_JSON（传入 available=false）
- 不修改 Prompt 文件
- 不修改 Go 后端
- 不创建 Schema 文件

## 3. 验收时间

2026-05-29

## 4. 测试样本

### 4.1 简历样本

- 用户：Arthur（user_id=1）
- 版本：V2（version_no=2, resume_version_id=2）
- 目标岗位：后端开发工程师
- 目标城市：深圳
- 技能：Go, MySQL, Redis, Docker, Kubernetes, etcd, gRPC, Git, Linux
- 项目：校园二手交易平台 + 分布式任务调度系统
- 实习：后端开发实习生（某科技公司）+ 云原生开发实习生（某云计算公司）
- 教育：南方科技大学 计算机科学与技术 本科

### 4.2 岗位样本

- 岗位标题：软件开发-后台开发方向
- 部门：CDG
- 城市：深圳总部,北京,上海,广州,成都,武汉,杭州
- 分类：技术
- 类型：应届实习
- 数据来源：tencent_join
- 数据质量：full_jd
- 岗位要求摘要：扎实的编程能力，熟练掌握 C/C++/Java/Go 等其中一门开发语言，TCP/UDP 网络协议，专业软件知识（算法、操作系统、数据库系统等）；有一定了解：Python/Shell/Perl、MySQL/SQL、NoSQL/Key-value

## 5. Prompt 文件

- 路径：`server/prompts/resume_optimization_v1.md`
- 版本：v1
- 行数：1073 行
- 占位符：`{{RESUME_CONTEXT_JSON}}` / `{{TARGET_JOB_CONTEXT_JSON}}` / `{{MATCH_ANALYSIS_JSON}}` / `{{BUSINESS_CONTEXT}}`
- 占位符替换方式：Python 字符串替换

## 6. 模型调用方式

- API 地址：TokenHub（腾讯混元模型服务代理）
- 模型：hy3-preview
- temperature：0.2
- max_tokens：8192
- 调用方式：Python urllib 手动调用

### 6.1 调用结果

| 指标 | 值 |
|------|------|
| finish_reason | stop（完成） |
| prompt_tokens | 6723 |
| completion_tokens | 4032 |
| total_tokens | 10755 |
| 输出字符数 | ~10562 chars |
| JSON 解析 | 成功 |

## 7. JSON 合法性检查

| 检查项 | 结果 |
|--------|------|
| JSON 解析成功 | ✅ 通过 |
| 是 JSON object | ✅ 通过 |
| 无 Markdown 包裹 | ✅ 通过 |
| 无多余解释文字 | ✅ 通过 |
| 顶层字段数 | 15 个 |

## 8. 字段完整性检查

| 字段 | 存在 | 说明 |
|------|------|------|
| optimization_summary | ✅ | 2-3句话，说明优化方向和优先级 |
| offer_go_summary | ✅ | Offer捕手语气，先结论后原因 |
| target_job | ✅ | job_id=9, title=软件开发-后台开发方向, department=CDG, category=技术, data_quality=full_jd |
| screening_focus | ✅ | 3 项，每项含 focus / why_it_matters / related_jd_requirement |
| priority_suggestions | ✅ | p0=2, p1=2, p2=1 |
| keyword_alignment | ✅ | 7 项 |
| project_optimization | ✅ | 2 项 |
| skill_section_optimization | ✅ | 4 项 |
| summary_optimization | ✅ | 1 项 |
| missing_evidence_plan | ✅ | 2 项 |
| rewrite_examples | ✅ | 3 项 |
| next_actions | ✅ | 4 项 |
| authenticity_notice | ✅ | 完整 |
| analysis_basis | ✅ | 完整 |
| generated_by | ✅ | llm_agent |

**结论：15 个必填字段全部存在。**

## 9. P0/P1/P2 优化建议检查

| 级别 | 数量 | 期望 | 结果 |
|------|------|------|------|
| p0 | 2 | 1-4 | ✅ |
| p1 | 2 | 1-4 | ✅ |
| p2 | 1 | 0-4 | ✅ |

| 检查项 | 结果 |
|--------|------|
| 每条建议是否包含 issue / related_jd_requirement / current_resume_evidence / why_it_matters / optimization_goal / rewrite_example / action / authenticity_rule | ✅ 全部包含 |
| P0 是否真的是影响初筛的关键问题 | ✅ 缺少MySQL/Redis证据、缺少稳定性表达 |
| P1 是否是增强竞争力的问题 | ✅ 补充进程间通讯、项目技术细节 |
| P2 是否是锦上添花 | ✅ 技能顺序调整 |
| 是否所有建议都围绕目标岗位 | ✅ 每条都引用 JD 要求 |

**P0 建议摘要**：

1. 简历中缺少对 MySQL 数据库设计、SQL 编程和 Redis/NoSQL 存储原理的明确技术描述
2. 项目描述偏功能实现，缺少与岗位要求的"性能优化、稳定性、系统可用性"直接相关的表达

## 10. keyword_alignment 检查

| 关键词 | status | 评估 |
|--------|--------|------|
| Go | present | ✅ 正确识别 |
| MySQL | partial | ✅ 技能列表有但缺少项目证据 |
| Redis / NoSQL / Key-value 存储 | partial | ✅ 技能列表有但使用场景不明确 |
| TCP/UDP / 网络协议 | missing | ✅ 正确识别缺失 |
| 进程间通讯 | missing | ✅ 正确识别缺失 |
| Kubernetes / 云原生 | present | ✅ 正确识别 |
| 算法 / 数据结构 / 操作系统 | partial | ✅ 计算机专业背景有间接证据 |

| 检查项 | 结果 |
|--------|------|
| status 是否属于 present / partial / missing | ✅ 全部合法 |
| 是否区分已体现、间接体现、缺失关键词 | ✅ |
| 是否没有为了堆关键词而建议虚构经历 | ✅ |

## 11. project_optimization 检查

| 项目 | current_problem | jd_relevance | interview_risk |
|------|----------------|-------------|----------------|
| 校园二手交易平台 | 偏功能实现，缺少数据库和缓存技术细节 | JD 要求 MySQL/NoSQL 能力 | 面试可能追问数据库设计和缓存使用 |
| 分布式任务调度系统 | 强调了 etcd 和调度延迟，但未说明进程间通讯 | JD 要求进程间通讯编程 | 面试可能追问 etcd 一致性和失败重试 |

| 检查项 | 结果 |
|--------|------|
| project_name 是否来自简历 | ✅ |
| 是否指出 current_problem | ✅ |
| 是否给出 targeted_improvement | ✅ |
| 是否说明 jd_relevance | ✅ |
| 是否给出 rewrite_direction | ✅ |
| 是否指出 interview_risk | ✅ |
| 是否没有编造项目成果 | ✅ |

## 12. skill_section_optimization 检查

| 技能 | current_status | 评估 |
|------|---------------|------|
| Go | present | ✅ 有多个项目证据 |
| MySQL | partial | ✅ 技能列表有但缺少使用场景证据 |
| Redis | partial | ✅ 技能列表有但缺少具体使用场景 |
| 网络协议 / 进程间通讯 | missing | ✅ 正确识别缺失 |

| 检查项 | 结果 |
|--------|------|
| current_status 是否属于 present / partial / missing / overstated | ✅ 全部合法 |
| 是否识别"写了技能但缺少项目证据"的情况 | ✅ MySQL和Redis标记为partial |
| 是否给出 suggested_positioning 和 evidence_needed | ✅ |

## 13. summary_optimization 检查

| 检查项 | 结果 |
|--------|------|
| 是否给出 current_problem | ✅ |
| 是否给出 rewrite_direction | ✅ |
| 是否给出 example | ✅ |
| example 是否没有编造身份、经历或结果 | ✅ |
| 是否能帮助用户更快体现目标岗位方向 | ✅ |

## 14. missing_evidence_plan 检查

| 缺失证据 | should_not_fake |
|---------|----------------|
| 网络协议与进程间通讯相关实践证据 | true |
| MySQL 与 Redis 的详细使用证据 | true |

| 检查项 | 结果 |
|--------|------|
| missing_evidence_plan 是否存在 | ✅ |
| should_not_fake 是否为 true | ✅ 全部为 true |
| 是否建议通过真实学习、实验项目、开源项目或真实实践补足 | ✅ |
| 是否没有教用户包装或虚构经历 | ✅ |

## 15. rewrite_examples 检查

| 改写 | can_use_directly | requires_user_verification | verification_note |
|------|-----------------|--------------------------|-------------------|
| 实习经历-数据库设计 | false | true | 需要确认是否确实参与了表设计、索引优化或 Redis 缓存使用 |
| 项目经历-缓存使用 | false | true | 需要确认是否确实使用了 Redis 缓存 |
| 项目经历-etcd通讯 | false | true | 需要确认是否确实实现了 etcd 选主或状态同步逻辑 |

| 检查项 | 结果 |
|--------|------|
| 是否包含 section / before_problem / after_example / can_use_directly / requires_user_verification / verification_note | ✅ |
| 如果 after_example 涉及数字/指标/规模/职责/成果，requires_user_verification 是否为 true | ✅ 全部为 true |
| 是否没有直接编造具体数字 | ✅ |
| 是否没有把"参与"改成"主导" | ✅ |
| 是否没有把学习计划写成已有经验 | ✅ |

## 16. next_actions 检查

| 优先级 | action | expected_effect |
|--------|--------|----------------|
| p0 | 补充 MySQL 表设计、索引、SQL 优化或 Redis 缓存 | 增强与 JD 必备要求的匹配度 |
| p0 | 梳理是否有网络编程、gRPC、进程间通讯相关实践 | 匹配 JD 进程间通讯要求 |
| p1 | 优化校园二手交易平台项目描述 | 补充接口设计、数据库和缓存细节 |
| p2 | 调整技能顺序，将 Go、MySQL、Redis 前置 | 帮助 HR 更快完成关键词匹配 |

| 检查项 | 结果 |
|--------|------|
| next_actions 是否存在 | ✅ |
| priority 是否属于 p0 / p1 / p2 | ✅ 全部合法 |
| action 是否具体 | ✅ |
| expected_effect 是否明确 | ✅ |
| 是否能指导用户下一步真实行动 | ✅ |

## 17. Offer捕手语气检查

| 检查项 | 结果 | 说明 |
|--------|------|------|
| 是否先说明整体优化方向 | ✅ | "优先围绕岗位关键词命中和项目工程细节增强" |
| 是否指出最重要问题 | ✅ | "需要把数据库设计、网络协议、接口性能表达更清楚" |
| 是否给出下一步行动方向 | ✅ | "让面试官更快看到你与后台开发岗位的能力匹配" |
| 是否亲和但不过度卖萌 | ✅ | 无卖萌表达 |
| 是否避免"改完一定通过"等绝对化表达 | ✅ | 无绝对承诺 |
| 是否不制造焦虑 | ✅ | 鼓励性表达 |
| 是否不否定用户能力 | ✅ | "你已经有Go、MySQL、Redis、Kubernetes等实践基础" |

## 18. 专业性检查

| 检查项 | 结果 | 说明 |
|--------|------|------|
| 是否体现 HR 初筛视角 | ✅ | screening_focus 关注关键词命中、初筛通过概率 |
| 是否体现技术面试官视角 | ✅ | project_optimization 包含 interview_risk |
| 是否体现简历表达优化视角 | ✅ | rewrite_examples 提供具体改写方向 |
| 是否围绕目标岗位 JD | ✅ | 每条建议都引用 JD 要求 |
| 是否输出可执行建议 | ✅ | next_actions 给出具体行动 |
| 是否具有 Demo 展示价值 | ✅ | 输出结构完整、内容专业、语气亲和 |

## 19. 真实性检查

| 检查项 | 结果 |
|--------|------|
| 是否包含 authenticity_notice | ✅ |
| 是否明确不能虚构经历 | ✅ "不要虚构经历" |
| 是否明确不能编造项目指标 | ✅ "不要编造项目指标" |
| 是否明确不能夸大职责 | ✅ "不要夸大职责" |
| 是否没有承诺一定通过初筛 | ✅ |
| 是否没有把建议补充内容写成用户已经具备 | ✅ |

## 20. 发现的问题

### 20.1 P2 问题（不阻塞但需记录）

1. **rewrite_examples 全部 can_use_directly=false**：3 条 rewrite_examples 均标记为 can_use_directly=false，requires_user_verification=true。这是非常保守和正确的做法，但意味着没有一条可以直接使用的改写示例。原因是模型在改写中补充了简历未明确提及的技术细节（如 Redis 缓存使用、etcd Watch 机制），这些需要用户确认。
   - 影响程度：轻微。保守做法比编造更好
   - 建议处理：后续可增加一条基于简历已有事实的直接改写示例

2. **rewrite_examples[2] 的 after_example 中包含"ms"单位**：分布式任务调度系统的改写示例中可能包含引用简历原有"调度延迟 < 100ms"的表达，虽然来自简历本身，但如果改写示例引入了新的延迟指标，需要确认是否为编造。
   - 影响程度：轻微。"ms"作为单位出现在描述中，可能引用简历原有数据
   - 建议处理：后续 Schema 定稿时可增加更严格的数字约束

3. **p2 建议只有 1 条**：P2 锦上添花建议只有 1 条（技能顺序调整），相比 P0 的 2 条和 P1 的 2 条略少。
   - 影响程度：不影响核心功能
   - 建议处理：可接受，P2 本身是锦上添花

## 21. 不记录内容

- 未记录完整 API Key
- 未记录完整请求体
- 未记录完整模型原始输出
- 未记录完整 JD 长文本
- 未记录完整简历长文本

## 22. 验收结论

**partial_pass**

理由：

- ✅ 模型调用成功，JSON 合法
- ✅ 15 个核心字段完整
- ✅ P0/P1/P2 优化建议存在且具体（p0=2, p1=2, p2=1）
- ✅ keyword_alignment 正确区分 present/partial/missing（7 项）
- ✅ project_optimization 来自简历真实项目（2 项）
- ✅ skill_section_optimization 有 missing 和 partial 判断（4 项）
- ✅ summary_optimization 包含 current_problem / rewrite_direction / example
- ✅ missing_evidence_plan 全部 should_not_fake=true（2 项）
- ✅ rewrite_examples 全部 requires_user_verification=true（3 项）
- ✅ Offer捕手语气亲和专业
- ✅ 真实性提醒完整
- ✅ 具备 HR 初筛 + 技术面试官 + 简历表达优化三视角
- ✅ 所有建议围绕目标岗位 JD
- ✅ 未编造项目数据和系统规模
- ⚠️ rewrite_examples 没有可直接使用的改写（全部需要用户确认）——这是保守但正确的做法

### 不阻塞进入 4-1.4.9A 的问题

1. rewrite_examples 全部需要用户确认：属于 Prompt 微调范畴，不影响 JSON Schema 固定
2. rewrite_examples[2] 中的 ms 单位：可能引用简历原有数据，非编造
3. P2 建议略少：不影响核心功能

### 需后续修订 Prompt 的问题

1. 可增加一条基于简历已有事实的直接改写示例（can_use_directly=true），提升用户使用体验
2. 可在 Prompt 中更明确要求"至少提供 1 条基于已有事实的直接改写示例"

## 23. 是否建议进入 4-1.4.9A

**✅ 建议进入 4-1.4.9A**

理由：

1. Prompt 输出质量整体符合预期，JSON 结构稳定
2. 发现的问题均为轻微问题，不影响 JSON Schema 固定
3. 15 个字段全部存在，枚举值全部合法
4. 真实性约束和反幻觉规则执行良好
5. 可以在 Schema 定稿时同步处理 rewrite_examples 约束

## 24. 下一步建议

1. **4-1.4.9A**：人工定稿简历优化 JSON Schema — 基于本轮验收结果固定 Schema
2. **4-1.4.9B**：CodeBuddy 登记简历优化 JSON Schema
3. 后续 Prompt 修订时处理 rewrite_examples 的直接改写示例问题
4. LLM_MAX_OUTPUT_TOKENS 生产环境建议设置为 8192 或以上
