# K6-2：简历优化子 Agent 拆分设计

## 1. 性质

docs-only design。**未修改任何代码、Prompt、Schema、前端、后端。**

## 2. 当前 Resume Optimization Agent 问题

| 问题 | 说明 |
|------|------|
| Prompt 过长 | 1219 行，包含 18 个 section + 完整 JSON skeleton |
| Schema 过重 | 18 个 required 顶层字段 + 嵌套 required 子字段（总计 26+ values） |
| 真实 LLM 不稳定 | hy3-preview 下 JSON repair/schema 失败；DeepSeek V4 Pro 推理模型超时 |
| 单次输出过多 | 首屏 + 证据链 + 改写示例全部一次生成 |
| fallback 常用 | K5 阶段主要走 Mock agent fallback |

**核心矛盾：** 想要一次 LLM 调用输出高质量、大篇幅、结构化优化建议，但当前模型能力 + 超时限制下难以稳定实现。

## 3. 拆分目标

```
拆分目的不是增加技术复杂度，而是：
1. 降低单次 Prompt 复杂度 → 提升 LLM 成功率
2. 首屏快速展示核心建议 → 提升用户体验
3. 深度证据后置/折叠 → 保留质量不牺牲速度
4. 改写示例按需触发 → 亮点功能不阻塞首屏
5. 每个子 Agent 独立 Schema → 降低 required 字段
```

**原则：** 四主 Agent 不变，子 Agent 是 Resume Optimization Agent 内部的能力拆分。

## 4. 三个子 Agent 设计

### 4.1 Optimization Core Agent（首屏核心）

**定位：** 首屏 30-60s 内生成最有价值的优化结果。

**输入：**
```
resume_text, target_job, resume_version_no, job_id
```

**输出结构：**
```json
{
  "optimization_summary": "一句话总结当前简历针对该岗位最大问题",
  "apply_readiness": "ready | needs_revision | high_risk",
  "top_fix_items": [
    {
      "priority": "P0 | P1 | P2",
      "issue": "问题",
      "jd_evidence": "对应 JD 要求",
      "resume_gap": "简历当前缺口",
      "action": "怎么改",
      "rewrite_example": "可直接替换的表达",
      "why_it_matters": "为什么影响初筛"
    }
  ],
  "missing_keywords": ["string"],
  "final_checklist": ["string"],
  "authenticity_notice": "不编造经历，只强化真实经历表达",
  "generated_by": "llm_agent"
}
```

**约束：**
- `top_fix_items` ≤ 3 条（强制限制，不要求更多）
- required 字段 ≤ 8 个
- Prompt ≤ 200 行
- Schema `additionalProperties: false`，但 required 克制

**展示位置：** 首屏直接展示

**Fallback：** 现有 Mock Agent（top_fix_items=2, apply_readiness 等）

---

### 4.2 Evidence Chain Agent（深度证据链）

**定位：** 解释为什么系统给出这些优化建议。

**输入：**
```
job_detail, resume_detail, core_fix_items
```

**输出结构：**
```json
{
  "evidence_items": [
    {
      "jd_requirement": "JD 原文要求",
      "resume_evidence": "简历中已有证据",
      "gap": "缺失证据",
      "linked_fix_item": "对应优化建议编号",
      "confidence": "high | medium | low"
    }
  ],
  "missing_evidence": ["string"],
  "risk_explanation": ["string"],
  "generated_by": "llm_agent"
}
```

**展示位置：** 折叠区 / "查看证据链" 按钮

**特性：**
- 不阻塞首屏
- 可以后置生成（用户点击后加载）
- 失败时不影响 Core Agent 展示

**Fallback：** 返回空的 `evidence_items: []` + "暂无深度证据"

---

### 4.3 Rewrite Agent（按需改写）

**定位：** 用户点击某条优化建议后，生成更具体的简历表达。

**输入：**
```
原始简历片段, JD 要求, 优化建议 (issue + action), 用户真实经历约束
```

**输出结构：**
```json
{
  "original_text": "...",
  "rewritten_text": "...",
  "improvement_reason": "...",
  "matched_keywords": ["string"],
  "authenticity_check": "是否存在编造风险",
  "need_user_verify": true,
  "generated_by": "llm_agent"
}
```

**展示位置：** 按需按钮 / "生成更具体改写"

**特性：**
- 按需触发（不作为首屏必需）
- 不编造项目经历
- 可以作为 Demo 亮点功能

**Fallback：** 返回 "基于真实经历优化表达，请逐个确认。" + 通用指引

---

## 5. 页面交互草案

```
┌─────────────────────────────────────────┐
│ 目标岗位: 软件开发-后台开发方向           │
│ 当前简历: Arthur · V3 (最新)             │
├─────────────────────────────────────────┤
│ ▲ Optimization Core Agent (首屏)        │
│   投递准备度: quick_fix_before_apply     │
│   ┌─ P0: 补充 C/C++ 经验证据 ─────────┐ │
│   │ JD要求: C/C++ 相关经验            │ │
│   │ 简历缺口: 仅列出技能无项目支撑      │ │
│   │ 怎么改: 在项目描述中补充使用场景    │ │
│   │ 改写示例: "基于C++开发数据持久化层" │ │
│   │ [生成更具体改写 ↦ Rewrite Agent]   │ │  ← 按需触发
│   └──────────────────────────────────┘ │
│   ┌─ P1: 项目职责描述不够清晰 ────────┐ │
│   │ ...                               │ │
│   └──────────────────────────────────┘ │
│   [✓] 确认所有真实经历                  │
│   [✓] 未编造项目指标                    │
│   [✓] 修改后再次匹配                    │
├─────────────────────────────────────────┤
│ ▼ 深度分析（折叠区）                     │
│   [查看 JD 证据链 ↦ Evidence Agent]     │  ← 后置生成
│   [查看缺失证据分析]                     │
│   [展开全部改写示例]                     │
└─────────────────────────────────────────┘
```

## 6. 接口方案比较

### 方案 A：三个独立接口
```
POST /api/resumes/optimize/core
POST /api/resumes/optimize/evidence
POST /api/resumes/optimize/rewrite
```
- 优点：边界清晰，方便单独 fallback 和性能控制
- 缺点：新增接口多，前端改动大

### 方案 B：一个接口 + mode
```
POST /api/resumes/optimize?mode=core
POST /api/resumes/optimize?mode=evidence
POST /api/resumes/optimize?mode=rewrite
```
- 优点：改动少，兼容现有 `/api/resumes/optimize` 入口
- 缺点：service 内部分支复杂

### 推荐

**短期 Demo：方案 B（一个接口 + mode）**
- 前端改动最小，复用现有 ResumeOptimizationPage 结构
- 只需在 Service 层根据 mode 分发到不同子 Agent

**长期产品：方案 A（三个独立接口）**
- 独立路由、独立 fallback、独立超时控制
- 更清晰的 API 语义

## 7. Prompt / Schema 策略

### Core Agent
- **Prompt**：≤ 200 行，强约束，只输出 3 条核心建议 + JSON-only
- **Schema**：required 6~8 个，`additionalProperties: false`，无复杂嵌套
- **反例**：不要 18+ required、不要大 skeleton

### Evidence Agent
- **Prompt**：≤ 150 行，专注证据链，不要求首屏速度
- **Schema**：数组结构清晰，required ≤ 5

### Rewrite Agent
- **Prompt**：≤ 100 行，短输入，单条改写
- **Schema**：单对象，required ≤ 5

## 8. 模型策略

```
K6-2 不冻结最终模型。模型冻结后置到 K6-5。

使用建议：
- hy3-preview：适合首屏 Core Agent，速度优先
- DeepSeek V4 Pro：适合后置 Evidence/Rewrite，质量优先（需控制超时）
- 所有子 Agent 都必须有 Mock fallback
```

**不得写死"最终上线一定用某个模型"。**

## 9. K6 后续实施计划

```
K6-3：Resume Optimization Core Agent 最小实现
  - 新增 prompt/core + schema/core
  - 实现 Core Agent (LLM + Mock + Pipeline)
  - Service 层 mode=core 分发
  - curl 验证

K6-4：Evidence Chain + Rewrite Agent 核验或最小实现
  - 新增 prompt/evidence + schema/evidence
  - 新增 prompt/rewrite + schema/rewrite
  - 实现两个子 Agent
  - curl 验证

K6-5：hy3-preview vs DeepSeek V4 Pro 最终模型对比

K6-6：四主功能 + 子 Agent 浏览器演示验收
```

## 10. P0 / P1 / P2

| 级别 | 数量 | 说明 |
|------|------|------|
| P0 | 0 | ✅ |
| P1 | 0 | ✅ |
| P2 | 4 | 子 Agent 尚未实现 + 模型未冻结 + 浏览器未验 + OPT 仍依赖 fallback |

## 11. 判定

**K6-2 = pass。** 设计完成，零代码修改。可进入 K6-3。
