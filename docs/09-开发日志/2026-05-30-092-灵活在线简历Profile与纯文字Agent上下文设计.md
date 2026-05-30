# 2026-05-30-092-灵活在线简历Profile与纯文字Agent上下文设计

## 1. 本次任务

阶段 4-3.5.G：灵活在线简历 Profile 与纯文字 Agent 上下文设计。docs only，不修改任何代码。

## 2. 背景

4-3.5.B ~ 4-3.5.F 已完成三个杀手锏和极端情况处理的战略设计，但所有这些设计的输入层——"简历数据如何进入 Agent"——尚未定义。当前系统使用 resume_versions 表存储简历，字段固定（name/phone/email/location/target_city/target_role/5 个 JSON 字段），无法表达竞赛、开源、证书、论文等校招重要经历。此外，隐私字段（姓名/手机/邮箱）直接传给 LLM，存在隐私泄露风险。Agent 上下文格式不确定，PromptValues 直接 `json.Marshal(resume)`，字段不清晰、格式不统一。

## 3. 核心设计

### 3.1 半结构化自由简历

核心思路：**少量必填 + 自由模块 + 可排序 + 可编辑 + 可生成纯文字 Agent 上下文**

- 不复制招聘网站表单
- 不做 PDF 解析
- 不做纯自由文本
- 采用半结构化方案

### 3.2 ResumeProfile + ResumeSection

- ResumeProfile：存整体简历概要和求职意图（target_roles / expected_locations / accept_relocation / education_summary / contact_info_status）
- ResumeSection：存具体经历模块（section_type / section_title / content / sort_order）
- Profile 不存具体经历，Section 不存整体意图

## 4. 最小必填字段

| 字段 | 位置 | 说明 |
|------|------|------|
| target_roles | ResumeProfile | 目标岗位，至少 1 个 |
| expected_locations | ResumeProfile | 期望城市/地点，至少 1 个 |
| education_summary | ResumeProfile | 教育背景摘要，至少 1 条 |
| 至少一个能力证明 Section | ResumeSection | project / internship / competition / paper / open_source / course_project / lab / custom 之一 |

## 5. ResumeProfile 设计

```
ResumeProfile {
    id                  uint64     // 主键
    user_id             uint64     // 用户 ID
    title               string     // 简历标题
    target_roles        []string   // 目标岗位列表
    expected_locations  []string   // 期望城市/地点列表
    accept_relocation   bool       // 是否接受调配
    education_summary   string     // 教育背景摘要
    contact_info_status string     // 联系方式状态
    created_at          time.Time
    updated_at          time.Time
}
```

## 6. ResumeSection 设计

```
ResumeSection {
    id            uint64     // 主键
    profile_id    uint64     // 所属 Profile ID
    section_type  string     // 模块类型枚举
    section_title string     // 模块标题
    content       string     // 模块内容（自由文本）
    sort_order    int        // 排序序号
    created_at    time.Time
    updated_at    time.Time
}
```

section_type 16 种枚举：education / project / internship / work / competition / research / paper / open_source / certificate / portfolio / self_evaluation / language / overseas_experience / course_project / lab / custom

## 7. ResumeContextText 生成规则

- 按【求职意向】→【教育背景】→ 各 Section（按 sort_order）→【补充信息】固定顺序拼接
- section_type 到中文标题有固定映射，custom 类型使用 `【自定义模块：{section_title}】`
- 不包含隐私字段（姓名/手机/邮箱）
- Section content 原样输出，不改写
- 总长度建议不超过 4000 字符

## 8. 隐私字段处理

| 字段 | 页面展示 | 发送至 LLM | Agent 上下文表示 |
|------|---------|-----------|----------------|
| 姓名 | ✅ | ❌ | 不出现 |
| 手机号 | ✅ | ❌ | `contact_info_provided: true/false` |
| 邮箱 | ✅ | ❌ | `email_provided: true/false` |
| 身份证号 | ❌ 不收集 | ❌ | 不出现 |
| 精确生日 | ❌ 不收集 | ❌ | 不出现 |

核心原则：**LLM 不需要知道用户是谁，只需要知道用户能做什么。**

## 9. ResumeContextPayload

```json
{
  "resume_context_text": "【求职意向】...\n【教育背景】...",
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

## 10. 三个 Agent 影响

| Agent | 新输入 | 变化 |
|-------|--------|------|
| RecommendationAgent | ResumeContextText + profile_summary + JOB_CANDIDATES_JSON | 旧 RESUME_CONTEXT_JSON → 新 RESUME_CONTEXT_TEXT + PROFILE_SUMMARY_JSON + PRIVACY_FLAGS_JSON |
| MatchAnalysisAgent | ResumeContextText + profile_summary + JOB_CONTEXT_JSON | 同上 |
| ResumeOptimizationAgent | ResumeContextText + profile_summary + TARGET_JOB_CONTEXT_JSON + CURRENT_TARGET_SECTION | 同上，额外增加 CURRENT_TARGET_SECTION |

## 11. Prompt / Schema 影响

- Prompt 占位符：`RESUME_CONTEXT_JSON` → `RESUME_CONTEXT_TEXT` + `PROFILE_SUMMARY_JSON` + `PRIVACY_FLAGS_JSON`
- Prompt 内容：增加隐私约束、概况辅助说明、Section 引用指引
- Schema：可能新增 resume_context_text / profile_summary / section_count / section_types / privacy_fields
- 本阶段不修改 Prompt / Schema 文件，只做设计登记

## 12. 文档更新

- 新增：`docs/05-Agent设计/灵活在线简历Profile与纯文字Agent上下文设计.md`
- 新增：本开发日志
- 更新 17 篇现有文档（战略差异化、投递边界、JD证据链、QualityGate、极端情况、阶段4-3方案、岗位匹配Agent、岗位推荐Agent、Agent总览、Prompt模板、JSON输出规范、失败兜底、匹配分析接口、推荐接口、项目状态、下一步任务、变更记录、说明书）

## 13. 未修改内容

- 未修改 Go 代码
- 未修改 Prompt 文件
- 未修改 Schema 文件
- 未修改前端代码
- 未调用 Hunyuan
- 未创建 resume_profiles / resume_sections 表
- 未修改 resume_versions 表
- 未进入 4-3.6

## 14. 当前风险

| 风险 | 严重度 | 说明 |
|------|--------|------|
| ResumeContextText 格式需要实际 Prompt 验证 | 🟡 中 | 纯文字格式是否被 LLM 稳定消费，需要在后续 Prompt 更新中验证 |
| ResumeProfile + ResumeSection 表创建时机 | 🟡 中 | 需要在后续阶段创建新表并迁移现有数据 |
| 前端改造工作量大 | 🟡 中 | 从固定表单改为自由模块列表，需要重新设计简历编辑页面 |
| RESUME_CONTEXT_JSON → RESUME_CONTEXT_TEXT 迁移 | 🟡 中 | 现有 Prompt 文件和 Agent 代码都需要适配新的占位符 |

## 15. 是否可进入 4-3.5.H

**可以进入 4-3.5.H。**

4-3.5.G 已完成灵活在线简历 Profile 与纯文字 Agent 上下文设计，定义了：
- 半结构化自由简历方案
- ResumeProfile + ResumeSection 数据结构
- ResumeContextText 生成规则
- ResumeContextPayload 完整设计
- 隐私字段处理规则
- 与三个 Agent 的关系
- 与已有战略设计的关系

下一步进入 4-3.5.H，评估 ResumeContextPayload 对 Prompt / Schema / Agent 的统一影响，制定迁移计划。

不进入 4-3.6。
