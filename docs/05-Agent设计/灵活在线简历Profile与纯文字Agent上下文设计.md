# 灵活在线简历 Profile 与纯文字 Agent 上下文设计

> 创建时间：2026-05-30
> 所属阶段：4-3.5.G
> 依据文档：战略差异化与杀手锏确认、投递边界与优化停止条件设计、JD证据链匹配与START简历优化规则、Agent输出QualityGate设计、学生求职极端情况处理边界、阶段4-3方案、岗位匹配Agent、岗位推荐Agent、Agent总览、Prompt模板、JSON输出规范、失败兜底与Mock回退策略、resume.go、init.sql

## 1. 背景

当前项目已完成三个杀手锏的战略设计：

- 杀手锏一：投递边界与优化停止条件（4-3.5.C）
- 杀手锏二：JD 证据链匹配（4-3.5.D）
- 杀手锏三：Agent 输出 Quality Gate（4-3.5.E）
- 极端情况处理边界（4-3.5.F）

但所有这些设计的输入层——"简历数据如何进入 Agent"——尚未定义。

当前系统使用 `resume_versions` 表存储简历，字段包括 name/phone/email/location/target_city/target_role/education_json/experience_json/project_json/skills_json/self_evaluation/raw_text。这套结构的问题：

1. **字段固定、表单固化**：只有 5 个 JSON 字段（education/experience/project/skills/self_evaluation），无法表达竞赛、开源、证书、论文等校招重要经历
2. **隐私字段混入 Agent 输入**：name/phone/email 直接传给 LLM，存在隐私泄露风险
3. **Agent 上下文格式不确定**：当前 PromptValues 直接 `json.Marshal(resume)`，字段不清晰、格式不统一、Agent 难以稳定消费
4. **前端体验受限**：固定表单无法满足不同学生的差异化展示需求（有人要突出竞赛，有人要突出开源，有人要突出科研）

本文档定义灵活在线简历 Profile 的数据结构、纯文字 Agent 上下文格式、隐私字段处理规则，以及与三个业务 Agent 的关系。

## 2. 核心问题：简历输入层设计影响所有下游 Agent 判断

如果输入层（简历数据 → Agent 上下文）是混乱的，所有下游判断都不稳定：

- apply_readiness 依赖简历信息的完整性，信息不足无法判断
- JD 证据链依赖简历中有对应的经历证据，缺少经历类型就无法形成证据
- Quality Gate 依赖简历证据的可靠性，输入层隐私字段泄露就不安全
- 极端情况处理依赖 candidate_case_type 识别，信息结构不一致就无法识别

**因此必须在 Agent 接入上游先定义好简历输入层。**

## 3. 设计目标

1. 定义灵活的在线简历 Profile 数据结构，不复制招聘网站表单
2. 定义纯文字 Agent 上下文（ResumeContextText）的标准格式，让 Agent 稳定消费
3. 定义隐私字段处理规则，确保敏感信息不进入 LLM
4. 定义 ResumeContextPayload（ResumeContextText + profile_summary + privacy_flags）作为统一 Agent 输入
5. 明确与三个业务 Agent 的关系
6. 明确与已有战略设计（投递边界 / JD 证据链 / Quality Gate / 极端情况）的关系
7. 登记对 Prompt / Schema / 前端的影响

## 4. 基本原则一：不复制招聘网站表单

招聘网站（如 BOSS 直聘、拉勾、牛客）的简历录入是结构化表单，字段固定、类型固定、展示固定。

本项目不做招聘网站，不做：

- 不做"教育经历 × 3、实习经历 × 3、项目经历 × 3"的固定表单
- 不做字段级别的必填校验（如"起止时间必填""公司名必填"）
- 不做格式化的多段式简历预览
- 不做 PDF 简历生成

本项目做的是"投递前决策与简历初筛优化系统"，简历是输入、不是产出。

## 5. 基本原则二：不做 PDF 解析

PDF 解析存在以下问题：

1. PDF 格式不统一，解析准确率不稳定
2. 解析后的结构化质量不可控，可能丢失关键信息
3. 解析需要额外的 OCR / NLP 基础设施，增加工程复杂度
4. 解析后的内容仍需人工校验，不如直接让用户输入

本项目不做 PDF 上传和解析，用户直接在线填写。

## 6. 基本原则三：不做纯自由文本

纯自由文本（一个大文本框，用户随意写）的问题：

1. Agent 无法稳定提取结构化信息（目标岗位、期望城市、教育背景等）
2. 自由文本的格式、顺序、粒度完全不可控
3. 无法判断最低必填信息是否完整
4. 无法为后续前端展示提供结构化支持

本项目不做纯自由文本输入，而是采用"少量必填 + 自由模块"的半结构化方案。

## 7. 基本原则四：半结构化自由简历

核心思路：

> **少量必填 + 自由模块 + 可排序 + 可编辑 + 可生成纯文字 Agent 上下文**

- 少量必填：目标岗位、期望地点、教育背景摘要、至少一个能力证明模块
- 自由模块：用户可以自由添加项目/实习/竞赛/开源/论文/证书/自定义等模块
- 可排序：用户可以调整模块展示顺序
- 可编辑：每个模块的标题和内容都是自由文本
- 可生成纯文字 Agent 上下文：所有模块按顺序拼接为结构化纯文字，供 Agent 消费

## 8. 基本原则五：Profile 只存意图和关键摘要，Section 存具体经历

- ResumeProfile：存整体简历概要和求职意图（目标岗位、期望城市、教育背景摘要、联系方式状态）
- ResumeSection：存具体经历模块（项目、实习、竞赛等，每条一个 Section）
- Profile 不存具体经历内容，Section 不存整体意图
- 两者分离，有利于后续灵活组合、排序、生成不同版本的 Agent 上下文

## 9. 最小必填字段定义

为了让 Agent 能够做出基本判断，简历必须满足最低信息要求。

### 9.1 必填字段

| 字段 | 位置 | 说明 |
|------|------|------|
| target_roles | ResumeProfile | 目标岗位，至少 1 个 |
| expected_locations | ResumeProfile | 期望城市/地点，至少 1 个 |
| education_summary | ResumeProfile | 教育背景摘要（学校+专业+学历+时间段），至少 1 条 |
| 至少一个能力证明 Section | ResumeSection | 类型为 project / internship / competition / paper / open_source / course_project / lab / custom 之一 |

### 9.2 为什么是这些字段

- **target_roles**：没有目标岗位，Agent 无法判断匹配度，无法形成 JD 证据链
- **expected_locations**：没有期望地点，Agent 无法判断地点匹配，无法识别海外/地点风险
- **education_summary**：没有教育背景，Agent 无法判断学历匹配，校招场景教育是硬条件
- **至少一个能力证明 Section**：没有具体经历，Agent 无法提取证据链，apply_readiness 只能是 insufficient_information

### 9.3 不强制必填但建议填写

| 字段 | 位置 | 说明 |
|------|------|------|
| accept_relocation | ResumeProfile | 是否接受调配，影响地点匹配判断 |
| self_evaluation | ResumeSection | 自我评价，帮助 Agent 理解用户自我定位 |
| language | ResumeSection | 语言能力，影响海外/外企岗位匹配 |

### 9.4 必填校验时机

- 前端保存时提示但不阻止
- Agent 调用时校验：缺少必填字段时 `apply_readiness = insufficient_information`，`candidate_case_type = insufficient_information`

## 10. 隐私字段处理规则

### 10.1 核心原则

> **LLM 不需要知道用户是谁，只需要知道用户能做什么。**

### 10.2 隐私字段分类

| 字段 | 页面展示 | 存入 ResumeProfile | 发送至 LLM | Agent 上下文中的表示 |
|------|---------|-------------------|------------|-------------------|
| 姓名（name） | ✅ 可展示 | ✅ 存储 | ❌ 不发送 | 不出现 |
| 手机号（phone） | ✅ 可展示 | ✅ 存储 | ❌ 不发送 | `contact_info_provided: true/false` |
| 邮箱（email） | ✅ 可展示 | ✅ 存储 | ❌ 不发送 | `email_provided: true/false` |
| 身份证号 | ❌ 不收集 | ❌ 不存储 | ❌ 不发送 | 不出现 |
| 精确生日 | ❌ 不收集 | ❌ 不存储 | ❌ 不发送 | 不出现 |
| 目标岗位 | ✅ 可展示 | ✅ 存储 | ✅ 发送 | 原文 |
| 期望城市 | ✅ 可展示 | ✅ 存储 | ✅ 发送 | 原文 |
| 教育背景 | ✅ 可展示 | ✅ 存储 | ✅ 发送 | 脱敏后（学校名可保留"某大学"） |

### 10.3 脱敏规则

- **姓名**：完全移除，不出现在 ResumeContextText 中
- **手机号**：只传布尔值 `contact_info_provided`，不传具体号码
- **邮箱**：只传布尔值 `email_provided`，不传具体地址
- **学校名**：用户可选择脱敏（如"安徽某大学"），默认保留原校名
- **具体地址**：只传城市级别，不传详细地址

### 10.4 为什么不收集身份证号和精确生日

- 身份证号和精确生日不是 Agent 判断匹配度的必要信息
- 收集这类信息增加数据安全风险
- 校招场景下，教育背景和技能经历比年龄更有意义

## 11. ResumeProfile 最小数据结构

ResumeProfile 存储整体简历概要和求职意图。

### 11.1 数据结构

```
ResumeProfile {
    id                  uint64     // 主键
    user_id             uint64     // 用户 ID（外键）
    title               string     // 简历标题（如"Go后端开发方向"）
    target_roles        []string   // 目标岗位列表（如 ["Go 后端开发","后端开发工程师"]）
    expected_locations  []string   // 期望城市/地点列表（如 ["深圳","北京","上海"]）
    accept_relocation   bool       // 是否接受调配
    education_summary   string     // 教育背景摘要（如"安徽某大学，软件工程，本科，2023.09 - 2027.06"）
    contact_info_status string     // 联系方式状态："complete" / "partial" / "missing"
    created_at          time.Time  // 创建时间
    updated_at          time.Time  // 更新时间
}
```

### 11.2 字段说明

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | uint64 | 自增 | 主键 |
| user_id | uint64 | 是 | 用户 ID，外键关联 users 表 |
| title | string | 否 | 简历标题，方便用户区分不同方向的简历 |
| target_roles | []string | 是 | 目标岗位，至少 1 个 |
| expected_locations | []string | 是 | 期望城市/地点，至少 1 个 |
| accept_relocation | bool | 否 | 是否接受调配，默认 false |
| education_summary | string | 是 | 教育背景摘要，至少包含学校、专业、学历 |
| contact_info_status | string | 否 | 联系方式状态："complete" / "partial" / "missing"，默认 "missing" |
| created_at | time.Time | 是 | 创建时间 |
| updated_at | time.Time | 是 | 更新时间 |

### 11.3 不存入 ResumeProfile 的内容

- 具体经历内容（项目、实习、竞赛等）→ 存入 ResumeSection
- 姓名、手机号、邮箱 → 存入 users 表或 ResumeProfile 的隐私字段（不发给 LLM）
- 简历版本号 → Profile 本身就是可编辑的活文档，不需要版本号（版本管理由独立的版本快照功能负责）

### 11.4 contact_info_status 枚举

| 值 | 含义 |
|---|------|
| `complete` | 手机号和邮箱都已填写 |
| `partial` | 手机号或邮箱只填了一个 |
| `missing` | 手机号和邮箱都未填写 |

Agent 上下文中转换为：
- `contact_info_provided: true`（complete 或 partial）
- `email_provided: true`（complete 或 partial 且邮箱已填）
- `phone_provided: true`（complete 或 partial 且手机号已填）

## 12. ResumeSection 最小数据结构

ResumeSection 存储具体经历模块，每个模块一条记录。

### 12.1 数据结构

```
ResumeSection {
    id            uint64     // 主键
    profile_id    uint64     // 所属 Profile ID（外键）
    section_type  string     // 模块类型枚举
    section_title string     // 模块标题（如"AI 求职智能体"或"ACM 程序设计竞赛"）
    content       string     // 模块内容（自由文本）
    sort_order    int        // 排序序号
    created_at    time.Time  // 创建时间
    updated_at    time.Time  // 更新时间
}
```

### 12.2 字段说明

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | uint64 | 自增 | 主键 |
| profile_id | uint64 | 是 | 所属 Profile ID，外键关联 resume_profiles 表 |
| section_type | string | 是 | 模块类型，见 12.3 枚举 |
| section_title | string | 是 | 模块标题 |
| content | string | 是 | 模块内容，自由文本 |
| sort_order | int | 是 | 排序序号，决定展示和 Agent 上下文中的顺序 |
| created_at | time.Time | 是 | 创建时间 |
| updated_at | time.Time | 是 | 更新时间 |

### 12.3 section_type 枚举

| 枚举值 | 中文 | 说明 |
|--------|------|------|
| `education` | 教育背景 | 学校、专业、学历、时间段 |
| `project` | 项目经历 | 课程项目、个人项目、团队项目 |
| `internship` | 实习经历 | 企业实习 |
| `work` | 工作经历 | 正式工作（非校招常见，但保留） |
| `competition` | 竞赛经历 | ACM、数学建模、创新创业等 |
| `research` | 科研经历 | 实验室、课题、研究项目 |
| `paper` | 论文发表 | 学术论文 |
| `open_source` | 开源贡献 | GitHub 项目、PR、Issue |
| `certificate` | 证书资质 | 阿里云 ACP、AWS、软考等 |
| `portfolio` | 作品集 | 设计、写作、视频等 |
| `self_evaluation` | 自我评价 | 个人优势和定位 |
| `language` | 语言能力 | 英语六级、托福、雅思等 |
| `overseas_experience` | 海外经历 | 交换、访学、国际会议 |
| `course_project` | 课程项目 | 课程设计、大作业 |
| `lab` | 实验室经历 | 实验室工作、助研 |
| `custom` | 自定义模块 | 用户自定义标题和内容 |

### 12.4 能力证明 Section

以下 section_type 被视为"能力证明"：

- `project`：项目经历
- `internship`：实习经历
- `competition`：竞赛经历
- `paper`：论文发表
- `open_source`：开源贡献
- `course_project`：课程项目
- `lab`：实验室经历
- `custom`：自定义模块（用户可能用来自定义能力证明）

简历必须至少包含一个能力证明 Section，否则 `apply_readiness = insufficient_information`。

### 12.5 Section 内容格式

Section 的 content 是自由文本，不做结构化约束。用户可以按自己的习惯写，可以包含：

- 时间段
- 技术栈
- 职责描述
- 成果指标
- 任何用户认为重要的内容

Agent 上下文生成时，content 原样拼接，不做二次解析。

## 13. 自由模块设计

### 13.1 设计理念

> 模块是用户的，不是系统的。系统只提供类型标签和排序，不约束内容。

### 13.2 添加模块

- 用户可以从预设 section_type 中选择
- 用户可以选择 `custom` 类型，自定义模块标题
- 每个模块的标题和内容都是自由文本

### 13.3 排序

- 用户可以调整模块的 sort_order
- sort_order 决定前端展示顺序和 Agent 上下文中的出现顺序
- 推荐默认顺序：education → project → internship → competition → ... → self_evaluation

### 13.4 编辑

- 每个模块可以独立编辑标题和内容
- 不做字段级别的校验（不强制"必须写时间段""必须写技术栈"）
- 但 Agent 可以在分析中指出缺失信息（如"项目经历缺少技术栈"）

### 13.5 删除

- 用户可以删除任何非必填模块
- 至少保留一个能力证明 Section

## 14. ResumeContextText 生成规则

ResumeContextText 是从 ResumeProfile + ResumeSection 生成的结构化纯文字，供 Agent 消费。

### 14.1 生成原则

1. **不是随机文本**：按固定顺序、固定标题、固定格式拼接
2. **Agent 可稳定消费**：每次生成的格式一致，Agent 不需要猜测字段位置
3. **不包含隐私字段**：姓名、手机号、邮箱不出现
4. **保留原始表达**：Section content 原样输出，不做改写
5. **可读性强**：人类也能读懂，方便调试

### 14.2 生成顺序

```
1. 【求职意向】      ← 来自 ResumeProfile.target_roles / expected_locations / accept_relocation
2. 【教育背景】      ← 来自 ResumeProfile.education_summary
3. [按 sort_order 排列的各 Section]
   - 【项目经历】    ← section_type=project
   - 【实习经历】    ← section_type=internship
   - 【竞赛经历】    ← section_type=competition
   - ...
4. 【补充信息】      ← 来自 contact_info_status + 其他辅助信息
```

### 14.3 标题映射

section_type 到中文标题的映射：

| section_type | Agent 上下文标题 | 自定义模块标题 |
|-------------|----------------|--------------|
| `education` | 【教育背景】 | 不适用 |
| `project` | 【项目经历】 | 不适用 |
| `internship` | 【实习经历】 | 不适用 |
| `work` | 【工作经历】 | 不适用 |
| `competition` | 【竞赛经历】 | 不适用 |
| `research` | 【科研经历】 | 不适用 |
| `paper` | 【论文发表】 | 不适用 |
| `open_source` | 【开源贡献】 | 不适用 |
| `certificate` | 【证书资质】 | 不适用 |
| `portfolio` | 【作品集】 | 不适用 |
| `self_evaluation` | 【自我评价】 | 不适用 |
| `language` | 【语言能力】 | 不适用 |
| `overseas_experience` | 【海外经历】 | 不适用 |
| `course_project` | 【课程项目】 | 不适用 |
| `lab` | 【实验室经历】 | 不适用 |
| `custom` | 【自定义模块：{section_title}】 | 使用 section_title |

### 14.4 生成规则细节

1. **【求职意向】** 固定在最前面，格式：
   ```
   【求职意向】
   目标岗位：Go 后端开发
   期望地点：深圳、北京、上海
   是否接受调配：是
   ```

2. **【教育背景】** 固定在第二位，格式：
   ```
   【教育背景】
   安徽某大学，软件工程，本科，2023.09 - 2027.06
   ```

3. **各 Section** 按 sort_order 排列，格式：
   ```
   【项目经历】
   项目名称：AI 求职智能体
   技术栈：Go、GoFrame、MySQL、Redis、Hunyuan
   内容：负责岗位推荐、匹配分析、LLM Pipeline、JSON Repair、Mock fallback 等模块设计与实现。
   ```
   或自定义模块：
   ```
   【自定义模块：开源经历】
   参与 xxx 项目，主要负责 xxx。
   ```

4. **【补充信息】** 固定在最后，格式：
   ```
   【补充信息】
   阿里云 ACP，熟悉云计算与后端工程化。
   ```

5. **同类型多 Section**：如果用户添加了多个 project Section，标题相同但内容不同，按 sort_order 排列，每个 Section 独立成段

### 14.5 长度约束

- ResumeContextText 总长度建议不超过 4000 字符
- 超长时在生成阶段截断，优先保留求职意向、教育背景和靠前的 Section
- 截断时在末尾添加 `[部分内容已省略]` 标记

## 15. ResumeContextText 示例

### 15.1 完整示例

```
【求职意向】
目标岗位：Go 后端开发
期望地点：深圳、北京、上海
是否接受调配：是

【教育背景】
安徽某大学，软件工程，本科，2023.09 - 2027.06

【项目经历】
项目名称：AI 求职智能体
技术栈：Go、GoFrame、MySQL、Redis、Hunyuan
内容：负责岗位推荐、匹配分析、LLM Pipeline、JSON Repair、Mock fallback 等模块设计与实现。

【开源贡献】
参与 GoFrame 框架贡献，提交 3 个 PR，主要涉及中间件和配置模块。

【证书资质】
阿里云 ACP，熟悉云计算与后端工程化。

【自我评价】
具备 Go 后端开发实战经验，熟悉 Web 框架和数据库设计，对 AI 应用开发有浓厚兴趣。
```

### 15.2 信息不足示例

```
【求职意向】
目标岗位：后端开发
期望地点：深圳

【教育背景】
某大学，计算机科学，本科，2022.09 - 2026.06
```

此简历缺少能力证明 Section，Agent 应判断 `apply_readiness = insufficient_information`，`candidate_case_type = insufficient_information`。

### 15.3 自定义模块示例

```
【求职意向】
目标岗位：算法工程师
期望地点：北京、上海
是否接受调配：否

【教育背景】
某 985 大学，人工智能，硕士，2024.09 - 2027.06

【科研经历】
参与导师自然语言处理课题，负责数据清洗和模型训练，发表 1 篇 ACL 论文（二作）。

【自定义模块：数学建模竞赛】
2025 年全国大学生数学建模竞赛国家一等奖，负责建模和编程，题目是城市交通流量预测。

【语言能力】
英语六级 580 分，可阅读英文论文和技术文档。
```

## 16. ResumeContextPayload 设计

ResumeContextPayload 是传给 Agent 的完整输入包，包含纯文字上下文、辅助信号和隐私标记。

### 16.1 数据结构

```json
{
  "resume_context_text": "【求职意向】\n目标岗位：Go 后端开发\n期望地点：深圳、北京、上海\n是否接受调配：是\n\n【教育背景】\n安徽某大学，软件工程，本科，2023.09 - 2027.06\n\n【项目经历】\n项目名称：AI 求职智能体\n技术栈：Go、GoFrame、MySQL、Redis、Hunyuan\n内容：负责岗位推荐、匹配分析、LLM Pipeline、JSON Repair、Mock fallback 等模块设计与实现。\n\n【开源贡献】\n参与 GoFrame 框架贡献，提交 3 个 PR。\n\n【证书资质】\n阿里云 ACP。",
  "profile_summary": {
    "target_roles": ["Go 后端开发"],
    "expected_locations": ["深圳", "北京", "上海"],
    "accept_relocation": true,
    "section_count": 5,
    "has_project": true,
    "has_internship": false,
    "has_research": false,
    "has_competition": false,
    "has_open_source": true,
    "has_paper": false,
    "has_certificate": true,
    "has_self_evaluation": true,
    "has_language": false,
    "has_overseas_experience": false,
    "data_quality": "sufficient"
  },
  "privacy_flags": {
    "contact_info_provided": true,
    "email_provided": true,
    "phone_provided": true,
    "name_excluded": true
  }
}
```

### 16.2 字段说明

#### 16.2.1 resume_context_text

- 类型：string
- 说明：ResumeContextText 纯文字上下文，Agent 的主要输入
- 生成规则：见第 14 节

#### 16.2.2 profile_summary

- 类型：object
- 说明：辅助信号，帮助 Agent 快速了解简历概况，无需从纯文字中推断

| 字段 | 类型 | 说明 |
|------|------|------|
| target_roles | []string | 目标岗位列表 |
| expected_locations | []string | 期望地点列表 |
| accept_relocation | bool | 是否接受调配 |
| section_count | int | Section 总数 |
| has_project | bool | 是否包含项目经历 |
| has_internship | bool | 是否包含实习经历 |
| has_research | bool | 是否包含科研经历 |
| has_competition | bool | 是否包含竞赛经历 |
| has_open_source | bool | 是否包含开源贡献 |
| has_paper | bool | 是否包含论文发表 |
| has_certificate | bool | 是否包含证书资质 |
| has_self_evaluation | bool | 是否包含自我评价 |
| has_language | bool | 是否包含语言能力 |
| has_overseas_experience | bool | 是否包含海外经历 |
| data_quality | string | 数据质量评估 |

#### 16.2.3 data_quality 枚举

| 值 | 含义 |
|---|------|
| `sufficient` | 信息充足，至少有目标岗位、期望地点、教育背景和至少一个能力证明 |
| `minimal` | 信息最少，只有必填字段，能力证明不足 |
| `insufficient` | 信息不足，缺少必填字段 |

#### 16.2.4 privacy_flags

- 类型：object
- 说明：隐私标记，确认敏感字段未发送给 LLM

| 字段 | 类型 | 说明 |
|------|------|------|
| contact_info_provided | bool | 是否提供了联系方式（手机或邮箱） |
| email_provided | bool | 是否提供了邮箱 |
| phone_provided | bool | 是否提供了手机号 |
| name_excluded | bool | 姓名已从 Agent 上下文中排除（始终为 true） |

### 16.3 为什么需要 profile_summary

1. **Agent 无需从纯文字中推断简历概况**：直接知道有几个 Section、包含哪些类型
2. **快速判断 candidate_case_type**：如 has_internship=false 且 has_project=true，可判断为"有项目但无实习"的学生
3. **快速判断 apply_readiness**：如 section_count=0，直接 insufficient_information
4. **辅助 Quality Gate**：如 profile_summary.data_quality=insufficient 但 Agent 输出了 apply_readiness=ready_to_apply，应标记为质量异常

### 16.4 为什么需要 privacy_flags

1. **确认隐私字段已处理**：让 Agent 和调用方确认敏感信息已排除
2. **不传具体值只传布尔标记**：phone/email 只传"是否已提供"，不传具体号码
3. **Agent 可据此补充提示**：如 contact_info_provided=false，Agent 可在输出中提示"建议补充联系方式"

## 17. 与三个业务 Agent 的关系

### 17.1 RecommendationAgent

| 输入项 | 来源 | 说明 |
|--------|------|------|
| resume_context_text | ResumeContextPayload.resume_context_text | 简历纯文字上下文 |
| profile_summary | ResumeContextPayload.profile_summary | 简历概况信号 |
| JOB_CANDIDATES_JSON | jobs 表 | 候选岗位列表 |
| BUSINESS_CONTEXT | 硬编码 | 推荐业务约束 |

**PromptValues 变化**：

```
旧：
  RESUME_CONTEXT_JSON = json.Marshal(resumeVersion)

新：
  RESUME_CONTEXT_TEXT = ResumeContextPayload.resume_context_text
  PROFILE_SUMMARY_JSON = json.Marshal(ResumeContextPayload.profile_summary)
  PRIVACY_FLAGS_JSON = json.Marshal(ResumeContextPayload.privacy_flags)
  JOB_CANDIDATES_JSON = json.Marshal(jobs)
  BUSINESS_CONTEXT = 硬编码
```

### 17.2 MatchAnalysisAgent

| 输入项 | 来源 | 说明 |
|--------|------|------|
| resume_context_text | ResumeContextPayload.resume_context_text | 简历纯文字上下文 |
| profile_summary | ResumeContextPayload.profile_summary | 简历概况信号 |
| JOB_CONTEXT_JSON | jobs 表 | 目标岗位 JD |
| BUSINESS_CONTEXT | 硬编码 | 匹配分析业务约束 |

**PromptValues 变化**：

```
旧：
  RESUME_CONTEXT_JSON = json.Marshal(resumeVersion)

新：
  RESUME_CONTEXT_TEXT = ResumeContextPayload.resume_context_text
  PROFILE_SUMMARY_JSON = json.Marshal(ResumeContextPayload.profile_summary)
  PRIVACY_FLAGS_JSON = json.Marshal(ResumeContextPayload.privacy_flags)
  JOB_CONTEXT_JSON = json.Marshal(job)
  BUSINESS_CONTEXT = 硬编码
```

### 17.3 ResumeOptimizationAgent

| 输入项 | 来源 | 说明 |
|--------|------|------|
| resume_context_text | ResumeContextPayload.resume_context_text | 简历纯文字上下文 |
| profile_summary | ResumeContextPayload.profile_summary | 简历概况信号 |
| TARGET_JOB_CONTEXT_JSON | jobs 表 | 目标岗位 JD |
| CURRENT_TARGET_SECTION | ResumeSection | 当前优化的目标 Section |
| BUSINESS_CONTEXT | 硬编码 | 简历优化业务约束 |

**注意**：ResumeOptimizationAgent 增加 `CURRENT_TARGET_SECTION` 参数，表示用户当前正在优化的具体 Section。这允许 Agent 针对某个模块给出精准优化建议，而不是对整份简历泛泛而谈。

### 17.4 统一变化总结

| 变化项 | 旧方式 | 新方式 |
|--------|--------|--------|
| 简历输入 | `json.Marshal(resumeVersion)` | `ResumeContextPayload.resume_context_text` |
| 简历概况 | 无 | `ResumeContextPayload.profile_summary` |
| 隐私标记 | 无 | `ResumeContextPayload.privacy_flags` |
| Prompt 占位符 | `RESUME_CONTEXT_JSON` | `RESUME_CONTEXT_TEXT` + `PROFILE_SUMMARY_JSON` + `PRIVACY_FLAGS_JSON` |

## 18. 与已有战略设计的关系

### 18.1 与投递边界（apply_readiness / optimization_stop_signal）的关系

ResumeContextText 是 apply_readiness 判断的输入基础。

- `apply_readiness = insufficient_information`：当 ResumeContextText 缺少必填字段或无能力证明 Section
- `apply_readiness = ready_to_apply`：当 ResumeContextText 包含充分的证据支撑目标岗位
- `optimization_stop_signal = fix_top_3_then_apply`：基于 ResumeContextText 中的具体经历，给出 Top 3 快速优化项

如果输入层信息不完整，apply_readiness 判断就不可靠。

### 18.2 与 JD 证据链的关系

ResumeContextText 是 JD 证据链中"简历证据"的来源。

- 每条 jd_evidence_chain 的 `resume_evidence` 来自 ResumeContextText 中的具体 Section 内容
- 如果 ResumeContextText 中缺少某类经历（如无实习），evidence_strength 只能是 `missing`
- 半结构化的 Section 让 Agent 能精确引用"你的项目经历中提到了 GoFrame"而不是"你可能用过一些框架"

### 18.3 与 Quality Gate 的关系

ResumeContextPayload 的 profile_summary 辅助 Quality Gate 检查。

- `profile_summary.data_quality = insufficient` 但 Agent 输出了 `apply_readiness = ready_to_apply` → Quality Gate 应标记为质量异常
- `privacy_flags.name_excluded = true` 但 Agent 输出中出现了姓名 → Quality Gate 应标记为质量异常
- `profile_summary.has_internship = false` 但 Agent 输出中引用了实习经历 → Quality Gate 应标记为编造

### 18.4 与极端情况处理的关系

ResumeContextPayload 的 profile_summary 是 candidate_case_type 识别的关键输入。

- `profile_summary.section_count = 0` 且 `data_quality = insufficient` → `candidate_case_type = insufficient_information`
- `profile_summary.has_internship = false` 且 `has_project = false` 且 `has_competition = false` → `candidate_case_type = weak_resume`
- `target_roles = ["Go 后端开发"]` 但所有 Section 都是算法/研究经历 → `candidate_case_type = target_role_mismatch`
- `expected_locations = ["海外"]` 且 `has_overseas_experience = false` → `candidate_case_type = overseas_or_location_risk`

## 19. 对 Prompt 的影响

### 19.1 占位符变化

| Prompt | 旧占位符 | 新占位符 |
|--------|---------|---------|
| match_analysis_v1.md | `{{RESUME_CONTEXT_JSON}}` | `{{RESUME_CONTEXT_TEXT}}` + `{{PROFILE_SUMMARY_JSON}}` + `{{PRIVACY_FLAGS_JSON}}` |
| recommendation_v1.md | `{{RESUME_CONTEXT_JSON}}` | `{{RESUME_CONTEXT_TEXT}}` + `{{PROFILE_SUMMARY_JSON}}` + `{{PRIVACY_FLAGS_JSON}}` |
| resume_optimization_v1.md | `{{RESUME_CONTEXT_JSON}}` | `{{RESUME_CONTEXT_TEXT}}` + `{{PROFILE_SUMMARY_JSON}}` + `{{PRIVACY_FLAGS_JSON}}` + `{{CURRENT_TARGET_SECTION}}` |

### 19.2 Prompt 内容调整方向

1. **简历输入说明**：从"以下是用户的简历 JSON"改为"以下是用户简历的结构化纯文字上下文"
2. **隐私约束**：增加"简历上下文中不包含姓名、手机号、邮箱等隐私字段，不要在输出中推测或编造这些信息"
3. **概况辅助**：增加"profile_summary 提供了简历概况信号，可直接用于快速判断，无需从纯文字中推断"
4. **Section 引用**：增加"引用简历证据时，尽量引用具体的 Section 标题和内容"

### 19.3 本阶段不修改 Prompt 文件

本阶段只做设计，不修改实际 Prompt 文件。Prompt 修改将在后续阶段统一实施。

## 20. 对 Schema 的影响

### 20.1 可能新增的 Schema 字段

| Agent | 可能新增字段 | 说明 |
|-------|------------|------|
| 匹配分析 | `resume_context_text` | 输入的简历上下文摘要（用于审计和调试） |
| 匹配分析 | `profile_summary` | 输入的简历概况信号 |
| 匹配分析 | `section_count` | 简历 Section 数量 |
| 匹配分析 | `section_types` | 简历包含的 Section 类型列表 |
| 匹配分析 | `privacy_flags` | 隐私标记 |
| 推荐 | 同上 | 同上 |
| 简历优化 | 同上 | 同上 |

### 20.2 本阶段不修改 Schema 文件

本阶段只做设计登记，不修改实际 Schema 文件。Schema 修改将在后续阶段统一实施。

## 21. 对前端的影响

### 21.1 简历录入页面改造方向

当前前端使用硬编码 Mock 数据，阶段 5 接入后端 API 时需要改造：

1. **从固定表单改为自由模块列表**：不再只有 5 个 JSON 字段，而是支持自由添加/删除/排序 Section
2. **新增 Section 类型选择器**：用户选择 section_type 后自动生成对应的中文标题
3. **新增自定义模块**：用户可以添加 custom 类型的模块
4. **新增排序功能**：用户可以拖拽调整模块顺序
5. **新增必填提示**：缺少必填字段时前端提示（但不阻止保存）

### 21.2 Agent 输出展示改造方向

1. **引用 Section 标题**：Agent 输出中引用简历证据时，可以引用具体 Section 标题
2. **隐私提示**：展示 `contact_info_provided = false` 时提示"建议补充联系方式"
3. **模块级优化建议**：简历优化建议可以针对具体 Section 给出

### 21.3 本阶段不修改前端代码

本阶段只做设计登记，前端改造属于阶段 5/6 范围。

## 22. 当前不做事项

| 不做 | 原因 |
|------|------|
| 不创建 resume_profiles 表 | 本阶段只做设计，数据库变更在后续阶段 |
| 不创建 resume_sections 表 | 本阶段只做设计，数据库变更在后续阶段 |
| 不修改 resume_versions 表 | 本阶段不改变现有表结构 |
| 不修改 Go 后端代码 | 本阶段只做文档 |
| 不修改 Prompt 文件 | Prompt 修改在后续阶段统一实施 |
| 不修改 Schema 文件 | Schema 修改在后续阶段统一实施 |
| 不修改前端代码 | 前端改造属于阶段 5/6 |
| 不调用 Hunyuan | 本阶段不涉及 LLM 调用 |
| 不进入 4-3.6 | 本阶段完成后下一步仍为 4-3.5.H |

## 23. 是否立即实施

**本阶段暂不立即实施工程代码。**

先完成设计文档和影响登记，后续根据优先级决定实施时机：

- resume_profiles / resume_sections 表的创建和迁移
- ResumeService 的 ResumeContextPayload 生成逻辑
- Prompt 文件的占位符更新
- Schema 文件的新增字段
- 前端页面的改造

这些工程实施不在 4-3.5.G 范围内。

## 24. 下一步建议

**下一步进入 4-3.5.H：ResumeContextPayload 对 Prompt / Schema / Agent 的影响统一评估。**

4-3.5.H 的核心内容：

1. 评估 ResumeContextPayload 对现有三个业务 Prompt 的影响
2. 评估 ResumeContextPayload 对现有三个业务 Schema 的影响
3. 评估 ResumeContextPayload 对 MatchAnalysisAgent / RecommendationAgent / ResumeOptimizationAgent 的影响
4. 制定 Prompt 占位符迁移计划
5. 制定 Schema 字段新增计划
6. 仍然以 docs only 为主，不急于改代码

## 25. 总结

本设计定义了灵活在线简历 Profile 和纯文字 Agent 上下文的完整方案：

1. **半结构化自由简历**：少量必填（target_roles / expected_locations / education_summary / 至少一个能力证明 Section）+ 自由模块（16 种 section_type + custom）
2. **ResumeProfile + ResumeSection**：Profile 存意图和摘要，Section 存具体经历
3. **ResumeContextText**：结构化纯文字格式，按固定顺序拼接，Agent 可稳定消费
4. **ResumeContextPayload**：resume_context_text + profile_summary + privacy_flags，Agent 的完整输入包
5. **隐私字段处理**：姓名/手机/邮箱不进入 LLM，只传布尔标记
6. **与三个 Agent 的关系**：统一使用 ResumeContextPayload 替代旧的 `json.Marshal(resumeVersion)`
7. **与已有战略设计的关系**：ResumeContextText 是 apply_readiness / JD 证据链 / Quality Gate / 极端情况处理的输入基础

核心一句话：

> **ResumeContextText 是所有 Agent 判断的输入基础。如果输入层混乱，所有下游判断都不稳定。**

---

4-3.5.H 已完成阶段汇总。
ResumeProfile / ResumeSection / ResumeContextText 暂不立即做数据库和前端工程改造，建议放到阶段 5 或后续产品体验阶段评估。
评审结论：不立即大改工程，先进入 4-3.6。
详见 `docs/05-Agent设计/阶段4-3.5战略转型汇总与工程评审.md`
