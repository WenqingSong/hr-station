# 匹配分析 Prompt 首轮输出验收报告

> 创建时间：2026-05-29
> 所属阶段：4-1.4.2
> 验收人：CodeBuddy

## 1. 验收目标

验证 `server/prompts/match_analysis_v1.md` 在真实模型输出时是否符合预期，具体包括：

1. 模型是否能输出合法 JSON
2. 输出是否包含完整结构
3. 输出是否体现 HR 初筛 + 技术面试官双视角
4. 输出是否包含 JD 关键要求拆解
5. 输出是否包含简历证据链
6. 输出是否包含 P0 / P1 / P2 简历优化建议
7. 输出是否包含面试追问预测
8. 输出是否包含真实性提醒
9. Offer鹅语气是否亲和但不幼稚
10. 分数是否克制，不是否动不动 90+
11. 是否出现幻觉、编造经历、编造指标
12. 输出是否适合 Demo 展示

## 2. 验收范围

- 只测试 1 组样本：Arthur V2 简历 + 腾讯岗位"软件开发-后台开发方向"
- 只进行 1 次主调用（实际执行 2 次，第 1 次因 max_tokens 不足截断）
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
- 实习：某科技公司（后端开发实习）+ 某云计算公司（云原生开发实习）
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

- 路径：`server/prompts/match_analysis_v1.md`
- 版本：v1
- 行数：1040 行
- 占位符：`{{RESUME_CONTEXT_JSON}}` / `{{JOB_CONTEXT_JSON}}` / `{{BUSINESS_CONTEXT}}`
- 占位符替换方式：Python 字符串替换

## 6. 模型调用方式

- API 地址：TokenHub（腾讯混元模型服务代理）
- 模型：hy3-preview
- temperature：0.2
- max_tokens：首次 4096（截断），二次 8192（成功）
- 调用方式：Python urllib 手动调用

### 6.1 调用结果

| 指标 | 首次调用 | 二次调用 |
|------|---------|---------|
| max_tokens | 4096 | 8192 |
| finish_reason | length（截断） | stop（完成） |
| completion_tokens | 4096 | 3618 |
| prompt_tokens | 6604 | 6604 |
| 总 token | 10700 | 10222 |
| 输出字符数 | 9988（截断） | 9281（完整） |
| JSON 解析 | 失败（截断） | 成功 |

## 7. JSON 合法性检查

| 检查项 | 结果 |
|--------|------|
| JSON 解析成功 | ✅ 通过 |
| 是 JSON object | ✅ 通过 |
| 无 Markdown 包裹 | ✅ 通过 |
| 无多余解释文字 | ✅ 通过 |
| 顶层字段数 | 17 个 |

## 8. 字段完整性检查

| 字段 | 存在 | 说明 |
|------|------|------|
| match_score | ✅ | 72 |
| match_level | ✅ | potential_match |
| match_label | ✅ | 建议尝试 |
| screening_result | ✅ | uncertain |
| offer_go_summary | ✅ | 亲和专业，先结论后原因 |
| jd_key_requirements | ✅ | 5 项 |
| resume_evidence_summary | ✅ | 4 项 |
| evidence_chain | ✅ | 5 项 |
| matched_points | ✅ | 3 项 |
| gap_points | ✅ | 3 项 |
| risk_points | ✅ | 3 项 |
| resume_optimization | ✅ | p0=2, p1=2, p2=2 |
| interview_questions | ✅ | 5 项 |
| authenticity_notice | ✅ | 完整 |
| next_actions | ✅ | 4 项 |
| analysis_basis | ✅ | 完整 |
| generated_by | ✅ | llm_agent |

**结论：17 个必填字段全部存在。**

## 9. 枚举合法性检查

| 枚举字段 | 值 | 合法 | 说明 |
|---------|---|------|------|
| match_level | potential_match | ✅ | 属于 strong_match / potential_match / explore_first / not_recommended |
| match_label | 建议尝试 | ✅ | 属于 优先投递 / 建议尝试 / 先了解再决定 / 暂不建议 |
| screening_result | uncertain | ✅ | 属于 likely / uncertain / unlikely |
| generated_by | llm_agent | ✅ | 等于 llm_agent |
| match_level 与 match_label 对应 | potential_match ↔ 建议尝试 | ✅ | 一致 |

### 9.1 子字段枚举检查

| 子字段枚举 | 检查项 | 结果 |
|-----------|--------|------|
| jd_key_requirements[].importance | core / important / optional | ✅ 全部合法 |
| jd_key_requirements[].source | description / requirements | ✅ 全部合法 |
| evidence_chain[].judgement | matched / partial / missing | ✅ 全部合法 |
| resume_evidence_summary[].strength_level | strong / medium / weak | ✅ 全部合法 |
| risk_points[].severity | high / medium / low | ✅ 全部合法 |
| next_actions[].priority | p0 / p1 / p2 | ✅ 全部合法 |

## 10. 证据链质量检查

| 检查项 | 结果 |
|--------|------|
| evidence_chain 数量 | 5 项（期望 3-6）✅ |
| 真实引用 JD | ✅ 每条 jd_requirement 来自岗位 JD |
| 真实引用简历 | ✅ 每条 resume_evidence 引用简历真实内容 |
| 有 matched / partial / missing 判断 | ✅ matched=2, partial=3, missing=0 |
| 能解释为什么匹配或不匹配 | ✅ 每条有 explanation |
| 是否编造简历经历 | ✅ 未发现编造 |

**证据链摘要**：

- Go 语言能力：简历有多处 Go 实战证据（项目+实习），判断 matched
- 分布式技术栈：具备 gRPC、etcd、K8s 经验，但与岗位网络编程/进程间通讯要求部分匹配，判断 partial
- 计算机专业基础：计算机专业+MySQL/Redis/Linux，但岗位要求的设计模式、网络安全等未直接体现，判断 partial
- MySQL 能力：技能列表有+实习中负责数据库设计，判断 matched
- Redis 能力：技能列表有但缺少具体使用场景和原理理解证据，判断 partial

## 11. P0/P1/P2 优化建议检查

| 级别 | 数量 | 期望 | 结果 |
|------|------|------|------|
| p0 | 2 | 1-3 | ✅ |
| p1 | 2 | 1-3 | ✅ |
| p2 | 2 | 0-3 | ✅ |

| 检查项 | 结果 |
|--------|------|
| p0 是否存在 | ✅ |
| 建议是否具体 | ✅ |
| rewrite_example 是否基于真实经历 | ⚠️ 大部分基于真实经历，p1[1] 的 rewrite_example 中包含简历未提及的具体性能指标（如"200ms降至80ms"），属于建议性表达而非简历事实，但具体数字非基于真实证据 |
| 是否避免虚构指标 | ⚠️ 大部分通过，p1[1] 的 rewrite_example 中的具体性能优化数字超出了简历原内容 |
| action 是否可执行 | ✅ |

**P0 建议摘要**：

1. 项目描述缺少与岗位核心要求（网络编程、进程间通讯）相关的技术细节
2. Redis 技能仅有名词，缺少具体使用场景和原理理解的证据

## 12. Offer鹅语气检查

| 检查项 | 结果 |
|--------|------|
| 先给结论 | ✅ "Offer鹅建议你认真尝试这个岗位" |
| 说明原因 | ✅ "你的 Go 后端开发、MySQL、Redis 以及云原生相关经历，与岗位后台开发方向有明显交集" |
| 给下一步方向 | ✅ "投递前建议补充项目中的技术细节和岗位关键词" |
| 亲和但不过度卖萌 | ✅ |
| 避免绝对化表达 | ✅ 无"绝对适合""一定通过"等表达 |

## 13. 专业性检查

| 检查项 | 结果 |
|--------|------|
| 体现 HR 初筛视角 | ✅ 关注岗位关键词命中、初筛通过可能性 |
| 体现技术面试官视角 | ✅ 面试追问预测涵盖 etcd 分布式一致性、性能优化、goroutine 并发等 |
| JD 拆解 | ✅ 5 项 JD 关键要求，区分 core / important |
| 指出真实风险 | ✅ 网络编程/进程间通讯经验不足、Redis 深度不够 |
| 面试追问预测 | ✅ 5 个具体技术问题 |
| Demo 展示价值 | ✅ 输出结构完整、内容专业、语气亲和 |

## 14. 真实性检查

| 检查项 | 结果 |
|--------|------|
| authenticity_notice 存在 | ✅ |
| 明确不能虚构经历 | ✅ "不要虚构经历" |
| 明确不能编造项目指标 | ✅ "不要编造项目指标" |
| 未编造项目数据 | ✅ 证据链未编造 |
| 未承诺一定通过初筛 | ✅ |
| 未建议用户造假 | ✅ |

## 15. 发现的问题

### 15.1 P1 问题（需后续修订 Prompt）

1. **rewrite_example 中存在少量超出简历事实的具体数字建议**：p1[1] 的 rewrite_example 中包含"将接口平均响应时间从 200ms 降至 80ms"，此具体数字并非来自简历原始数据。虽然 rewrite_example 是建议性表达而非简历事实，但 Prompt 第 14.5 节要求"如果输入中没有足够事实，不能虚构具体结果、数据"。
   - 影响程度：轻微，不影响主要结论
   - 建议处理：后续修订 Prompt 时可增加"rewrite_example 中的具体数字应标注为示例，或使用'如果实际数据为 X，可以...'的条件式表达"

### 15.2 P2 问题（不阻塞但需记录）

1. **首次调用因 max_tokens=4096 导致输出截断**：模型输出 completion_tokens 达到 4096 上限后截断，finish_reason=length，导致 JSON 不完整。增加至 8192 后完成（实际使用 3618 completion_tokens）。
   - 影响程度：生产环境配置问题，非 Prompt 质量问题
   - 建议处理：LLM_MAX_OUTPUT_TOKENS 应设置为 ≥4096，建议 8192

2. **证据链中缺少 missing 判断**：5 条证据链中只有 matched 和 partial，没有 missing 判断。虽然 resume_evidence 为空时也会说明，但未使用 missing 枚举值。
   - 影响程度：轻微，证据链仍然有效
   - 建议处理：后续修订 Prompt 时可强调"如果简历中完全没有相关证据，必须使用 missing 判断"

## 16. 不记录内容

- 未记录完整 API Key
- 未记录完整请求体
- 未记录完整模型原始输出
- 未记录完整 JD 长文本
- 未记录完整简历长文本

## 17. 验收结论

**partial_pass**

理由：

- ✅ 模型调用成功，JSON 合法
- ✅ 17 个核心字段完整
- ✅ 枚举值全部合法
- ✅ 证据链真实引用 JD 和简历
- ✅ P0/P1/P2 优化建议存在且具体
- ✅ Offer鹅语气亲和专业
- ✅ 分数克制（72 分）
- ✅ 真实性提醒完整
- ✅ 具备 HR 初筛 + 技术面试官双视角
- ⚠️ rewrite_example 中存在少量超出简历事实的具体数字建议（P1 问题）
- ⚠️ 证据链中缺少 missing 判断（P2 问题）

### 不阻塞进入 4-1.4.3A 的问题

1. rewrite_example 中的具体数字建议：属于 Prompt 微调范畴，不影响 JSON Schema 固定
2. max_tokens 配置问题：属于工程配置，非 Prompt 质量问题
3. 证据链缺少 missing 判断：不影响结构，后续 Prompt 微调时补充

### 需后续修订 Prompt 的问题

1. rewrite_example 的具体数字约束：可增加"如果未从输入中获得具体数字，rewrite_example 应使用条件式表达或标注为'示例数据'"
2. 证据链 missing 判断的强化：可在 Prompt 中更明确地要求"当简历中完全没有相关证据时，必须使用 missing"

## 18. 是否建议进入 4-1.4.3A

**✅ 建议进入 4-1.4.3A**

理由：

1. Prompt 输出质量整体符合预期，JSON 结构稳定
2. 发现的问题均为轻微问题，不影响 JSON Schema 固定
3. JSON Schema 定稿后，后续 Prompt 微调不会影响核心字段结构
4. 可以在 Schema 定稿时同步处理 rewrite_example 约束和 missing 判断要求

## 19. 下一步建议

1. **4-1.4.3A**：人工定稿匹配分析 JSON Schema — 基于本轮验收结果固定 Schema
2. **4-1.4.3B**：CodeBuddy 登记匹配分析 JSON Schema
3. 后续 Prompt 修订时处理 rewrite_example 约束和 missing 判断要求
4. LLM_MAX_OUTPUT_TOKENS 生产环境建议设置为 8192 或以上
