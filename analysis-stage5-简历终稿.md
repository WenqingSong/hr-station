# 第 5 阶段：简历「项目经历」终稿

> 综合前 4 阶段（850+ 文档、550+ 开发日志）提炼而成

---

## 版本一：简洁版（适合一页简历，约 200 字）

> **Offer 捕手 — 校招 AI 岗位匹配与简历优化平台** | Go + React + TypeScript + LLM
> 
> 独立设计并实现了一个面向校招求职者的 AI 辅助决策平台，核心功能是基于 LLM 的岗位匹配分析和简历优化。后端采用 Go 标准库（零框架依赖），前端采用 React + Vite + TailwindCSS，接入腾讯混元 Hy3 大模型。设计并落地了四主 Agent 架构（推荐/匹配分析/简历优化/版本对比），实现了 Prompt→LLM→JSON Parse→Schema Validate→JSON Repair→Mock Fallback 的完整 LLM 工程化链路。采用 Fast（秒级语义匹配，4-5ms）+ Deep（5 模块 LLM 深度分析）双阶段策略。建立了文档驱动的工程管理体系（850+ 文档，550+ 开发日志，每个任务可追溯），18 天完成从零搭建到可上线 Demo。

---

## 版本二：STAR 展开版（适合详细简历，约 400 字）

### 项目背景（Situation）

校招求职存在严重信息不对称：学生海投无反馈、无法量化匹配度、不知如何优化简历。需要构建一个 AI 驱动的投递前决策工具，帮助学生在投递前判断"该不该投、为什么能过/不能过、简历怎么改"。

### 我的职责与行动（Task & Action）

**独立负责全栈开发与架构设计**，从零搭建项目：

**1. 后端架构与 LLM 工程化**
- 使用 Go 标准库（`net/http` + `database/sql`）构建后端，零 Web 框架/ORM 依赖，14 个 REST API
- 设计并实现四主 Agent 架构（岗位推荐/匹配分析/简历优化/版本对比）
- 构建完整的 LLM Pipeline 安全链路：Prompt 管理 → HunyuanClient 调用 → JSON 提取解析 → JSON Schema 校验（draft 2020-12）→ JSON Repair（单次非递归）→ Mock Fallback
- 设计 Quality Gate 质量门禁体系：4 级枚举 × 10 维检查 × 10 种极端情况处理策略
- 132+ 单元测试覆盖全部核心模块

**2. 前端工程化与产品体验**
- 将 Figma Make 生成的 9 个原型页面改造为完整工程化前端（React + TypeScript + Vite + TailwindCSS）
- 建立四层数据消费模型（API Client → API 模块 → ViewModel Loader → 页面组件），9 个 Loader 函数
- 实现 Fast（秒级）+ Deep（LLM 深度）双阶段展示，6 状态前端状态机管理

**3. 工程管理与文档体系**
- 建立文档驱动的工程管理方法：850+ Markdown 文档、550+ 逐日开发日志、每个任务编号可追溯
- 采用"只读审计→设计冻结→工程实现→验收"四步闭环，每个任务有明确的 P0/P1/P2 分级和验收标准
- 搭建混合式开发环境（Docker MySQL+Nginx + 宿主机 Go/Vite），40+ Makefile 命令统一入口

### 项目成果（Result）

- **18 天**完成 13 个阶段 550+ 子任务，从零搭建到可上线 Demo
- 构建 **46 条**岗位样本库 + **9 份**示范简历，支撑完整演示链路
- Fast Match：**4-5ms** 响应，Fast Recommend：**50-83ms**
- Prompt 资产：**4 个**定稿模板 × **3 个** JSON Schema，均通过首轮输出验收
- 3 个经典演示案例全部达标（score 78/33/42，覆盖 revise/change_direction/not_recommended 三场景）
- **P0=0**，所有阶段验收通过

---

## 版本三：技术亮点版（适合技术岗简历，约 300 字）

### Offer 捕手 — 校招 AI 投递决策平台

**技术栈**：Go（标准库零框架） + React 18 + TypeScript + MySQL + 腾讯混元 Hy3 LLM

**核心工作**：

- **LLM 工程化 Pipeline**：自研 Prompt→LLM Client→JSON Parse→Schema Validate→JSON Repair→Mock Fallback 六层安全链路，支持 Hunyuan/DeepSeek 双 Provider，20+ 错误码全覆盖，132+ 单元测试
- **四主 Agent 架构**：设计岗位推荐/匹配分析/简历优化/版本对比四大 Agent，Fast（秒级语义画像 4-5ms）+ Deep（5 模块 Hy3 LLM 驱动）双阶段策略
- **Quality Gate 质量门禁**：设计 4 级枚举 × 10 维检查 × 10 种极端情况处理的内容质量保障体系，fabrication_check 一旦失败直接 fallback
- **语义画像系统（SEP）**：构建统一的 Semantic Evidence Profiling 底座，46 岗位 + 9 简历全量就绪，content_hash 比对新旧度
- **前端四层数据架构**：API Client → 业务 API 模块 → ViewModel Loader → 页面组件，原生 fetch 零 axios 依赖，5 种 agent_status 前端展示规则
- **文档驱动工程管理**：850+ 文档 + 550+ 逐日日志 + P0/P1/P2 三级问题分级 + 四步闭环（审计→设计→实现→验收），18 天完成全流程

---

## 版本四：一页纸子弹点版（最高密度，适合空间有限）

> **Offer 捕手** — Go + React + TypeScript + LLM | 独立全栈开发 | 2026.05-2026.06
>
> - 独立设计并实现校招 AI 投递决策平台，四主 Agent 架构（推荐/匹配/优化/对比），14 个 REST API
> - 自研 LLM Pipeline：Prompt→LLM→JSON Parse→Schema Validate→JSON Repair→Mock Fallback，20+ 错误码全覆盖，132+ 单元测试
> - 设计 Quality Gate 质量门禁（4 级 × 10 维检查）+ SEP 语义画像底座（46 岗位 + 9 简历）
> - Fast/Deep 双阶段：Fast 秒级语义匹配（4-5ms），Deep 5 模块 Hy3 LLM 深度分析
> - 前端四层数据架构（API Client → 模块 → ViewModel Loader → 页面），原生 fetch 零框架依赖
> - 文档驱动工程管理：850+ 文档 + 550+ 逐日日志，P0/P1/P2 三级分级，18 天完成 13 阶段 550+ 任务

---

## 关键数据速查表

> 以下数据可在面试中被问到时用于支撑回答：

| 数据项 | 数值 |
|--------|------|
| 开发周期 | 18 天 |
| 开发阶段 | 13 个（0→13） |
| 子任务数 | 550+ |
| 后端代码 | 146 个 Go 源文件 |
| 前端代码 | 139 个 TSX + 40 个 TS |
| 文档规模 | 850+ Markdown |
| API 接口 | 14 个 REST |
| Agent 数量 | 4 主 Agent + 3 子 Agent |
| Prompt 模板 | 4 个定稿 |
| JSON Schema | 3 个（draft 2020-12） |
| Go 单元测试 | 132+ |
| LLM 错误码覆盖 | 20+ |
| 岗位样本 | 46 条 |
| 示范简历 | 9 份 |
| Fast Match 延迟 | 4-5ms |
| Fast Recommend 延迟 | 50-83ms |
| 前端 Build | 1.51s |
| Makefile 命令 | 40+ |

---

## 面试可能追问及准备要点

### Q1：LLM Pipeline 如何保证输出稳定性？

**答**：五层防线——① Prompt 约束（强制 JSON + 枚举 + Schema 草案）② JSON 格式校验（draft 2020-12）③ JSON Repair（Hunyuan 修复一次，禁止递归）④ Mock Fallback（规则引擎兜底）⑤ Quality Gate（10 维内容检查，fabrication 一旦失败直接 fallback）。Pipeline 20+ 错误码全覆盖，用户侧始终看到稳定业务结果。

### Q2：Fast/Deep 双阶段策略怎么设计的？

**答**：Fast 用语义画像引擎（SEP）做秒级初筛，输出分数 + 等级 + 2 匹配点 + 1 短板 + 1 建议；Deep 用 5 模块 Hy3 LLM 做深度分析，输出决策/命中点/淘汰风险/改写指令/岗位策略。两者共享同一套分数语义（90-100 强匹配 → <60 不建议），差值 ≤ 9 通过一致性校验。

### Q3：怎么管理 Prompt 的迭代？

**答**：Prompt 作为独立产品资产在 `server/prompts/` 目录管理，人工定稿 → CodeBuddy 登记 → 首轮输出验收（partial_pass）→ Schema 吸收问题 → 最小一致性修复。三个 Prompt 各经历 2-3 轮迭代，每次修改都有开发日志记录。

### Q4：遇到过什么技术难点？

**答**：最大难点是"语义画像割裂"——推荐系统已用 SEP，匹配系统仍用纯标签计数导致分数虚高（base 60）。通过全链路审计发现根因后，重构了 830 行 Fast Match Engine，接入 SEP + Hard Gate（8 条）+ Evidence Mapping + 新分数公式，将 responses 从纯标签改为结构化证据。

### Q5：如何保证工程质量和可维护性？

**答**：文档驱动的四步闭环——每个任务先只读审计（不修代码）→ 设计冻结（写清"做什么/不做什么/验收标准"）→ 工程实现（不改范围外代码）→ 对照验收。550+ 日志每个都有 10 节标准格式、P0/P1/P2 分级，任何问题可追溯。

---

## 使用建议

| 场景 | 推荐版本 |
|------|---------|
| 一页简历 | **版本四**（子弹点版，最高密度） |
| 两页详细简历 | **版本二**（STAR 展开版） |
| 技术岗简历 | **版本三**（技术亮点版） |
| 通用简历 | **版本一**（简洁版） |
| 面试准备 | **关键数据速查表 + 追问要点** |

