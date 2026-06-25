# Offer 捕手 —— 面试全景回顾

> 这是一份面试准备文档，帮你回忆起 18 天开发过程中的每一个关键决策、技术细节和故事。不是简历的复述，而是让面试官觉得"你真的做过"的深层记忆。

---

## 一、一句话说清楚这项目是什么

**Offer 捕手是一个面向校招求职者的 AI 投递决策平台。** 学生上传简历后，系统用 LLM 分析「这份简历投这个岗位，初筛能过吗？差在哪里？怎么改？」——不是简单的匹配打分，而是给出可执行的简历修改指令和投递策略。

**核心差异化**：不替用户编造经历，而是帮用户更好地呈现真实经历。

---

## 二、18 天开发全流程：你能讲出来的时间线

面试官如果问「这个项目做了多久？怎么规划的？」，你不需要背数字，只需要讲出**节奏感**：

### 前 3 天（5/28）：从零到有 —— 基础建设

- **第一天**：项目审计 → 创建 850+ 文档体系 → 定义产品范围和不做范围 → Makefile + .env + Docker → Go 后端骨架 + `/healthz` → 统一响应结构 + 中间件 + MySQL 连接。**一天内 67 条开发日志。**
- 关键决策：**ADR-001 用 Go 标准库，不引入 Gin/GORM**；**ADR-002 单主简历，版本迭代**；**ADR-003 先 Mock 后接 LLM**。

### 第 4-5 天（5/29-30）：最密集阶段 —— LLM 全部接入

- 腾讯招聘公开数据源接入（3 个接口确认）→ jobs 表扩展 8 字段 → 5 条真实岗位入库
- **4 个 Prompt + 3 个 JSON Schema 定稿并首轮验收**（均 partial_pass）
- LLM Client 抽象 + HunyuanClient + JSON 工具 + Schema Validate + **JSONRepairAgent**
- **统一 Pipeline 建成**：Prompt → LLM → JSON Parse → Schema Validate → Repair Once → Fallback
- 132+ 单元测试全部通过
- **战略转向**（4-3.5）：从"推荐工具"升级为"投递决策系统"，新增三大杀手锏
- **全容器回退混合式**（4-3.7.B.R）：Docker MySQL+Nginx + 宿主机 Go/Vite

### 第 6 天（5/31）：单日 52 个子任务 —— 前端闭环

- TypeScript 类型体系（7 文件）+ API Client（原生 fetch）+ ViewModel Loader（9 函数）
- 四层数据消费模型建立
- 5 页面骨架统一 + Navbar 统一 + 简历路由补齐
- Build 1.51s，3 条演示路径 11 URL 全部 200

### 第 7-8 天（6/1）：联调 —— 前后端握手

- 6 个组件从 Mock 切到真实 API
- Agent 联调：推荐/匹配/优化 全部 curl 200
- 登录态闭环（JWT + Auth 中间件 + 前端 token 管理）

### 第 9-10 天（6/2）：上下文真实化 + 岗位样本库

- 3 个 Agent 页面的上下文从 Mock 切到真实岗位和简历
- 腾讯招聘 46 条岗位导入（162K JSON）→ 数据库重构 + 前端筛选适配

### 第 11-18 天（6/3-14）：深度打磨与冲刺

- **H18 系列**：Fast Match 全链路重构（830 行新引擎，8 大模块）
- **H19 系列**：Deep Match 后端重构（5 模块 Hy3 驱动）
- **H20 系列**：产品化 + 契约化（页面只回答 4 问；P0 修复 × 2；视觉统一）
- **JM 系列**：匹配核心重构（从规则引擎到真 Hy3 LLM；3 案例全部达标）
- 最终部署前检查 + 5 份提交文档

---

## 三、技术架构：面试官一定会追问的

### 3.1 为什么用 Go 标准库而不是 Gin？

**回答要点**：这不是一个追求"用最流行框架"的项目。我想证明我能从零搭建一个完整的后端，理解 HTTP 的底层机制，而不是依赖框架帮你做路由和中间件。

具体来说：
- `net/http` 的 `ServeMux` 完全够用（Go 1.22+ 支持方法路由和路径参数）
- 中间件链是自己写的（Recovery → CORS → Logger → Auth），每个都理解原理
- `database/sql` 裸 SQL —— 我能写 JOIN、能设计索引、能管理连接池
- 零 ORM 依赖，减少了一层抽象，也减少了一层隐式 bug 的可能性

### 3.2 四主 Agent 架构是怎么想的？

**回答要点**：类比一个求职顾问团队的视角。

| Agent | 职责 | 面试时怎么讲 |
|-------|------|------------|
| 岗位推荐 Agent | "你适合投哪些岗位？" | 有点像推荐系统，但不是协同过滤——基于简历语义画像和岗位方向匹配 |
| **岗位匹配分析 Agent** ⭐ | "这个岗位你能投吗？差在哪？" | **核心功能**。分为 Fast（秒级）和 Deep（LLM 深度）两阶段 |
| 简历优化 Agent | "简历怎么改才能匹配这个岗位？" | 3 个子 Agent 拆分：内容诊断 + 关键词对齐 + 表达改写 |
| 版本对比 Agent | "新版简历比旧版好在哪？" | 唯一未 LLM 化的 Agent（P1 待做），目前用规则引擎 |

### 3.3 LLM Pipeline 的完整链路

这是你面试时要能画出来的图：

```
业务 Service 构造 PromptValues
    ↓
Prompt Loader（加载 .md 模板 + {{占位符}} 替换）
    ↓
Schema Loader（加载 JSON Schema draft 2020-12）
    ↓
llm.Client.Chat（HunyuanClient / NoopClient）
    ↓
llmjson.ExtractAndParse（去除 Markdown fence + BOM 清理 + JSON parse）
    ↓
Schema Validator（santhosh-tekuri/jsonschema/v5）
    ↓
    ├── 通过 → agent_status = "success"
    ├── 失败 → JSONRepairAgent.Repair()（仅一次，禁止递归）
    │       ├── 修复成功 → agent_status = "repaired"
    │       └── 修复失败 → agent_status = "fallback_required"
    └── fallback_required → Service 层调用 Mock Agent → generated_by = "mock_agent"
```

**五个关键设计点**：
1. **Repair 只做一次**：`LLM_REPAIR_MAX_ATTEMPTS=1`，不递归。修不好就 fallback，避免成本失控
2. **Repair 不改变 generated_by**：修复成功仍标记为 `llm_agent`，不是 `json_repair_agent`
3. **禁止的 generated_by 值**：`llm_agent_fallback` 这种混淆值不允许出现
4. **Pipeline 只做决策，不调 Mock**：Pipeline 返回 fallback 决策给 Service 层，由 Service 决定调哪个 Mock
5. **20+ 错误码全覆盖**：从 JSON 解析失败到 LLM 超时到 API Key 缺失，每种都有明确处理路径

### 3.4 Fast/Deep 双阶段为什么要这么设计？

**回答要点**：用户等不了 30 秒看结果。

- **Fast（秒级）**：基于 SEP 语义画像做规则匹配，**4-5ms** 返回初筛判断。用户一进页面就能看到结果。
- **Deep（分钟级）**：5 模块 Hy3 LLM 深度分析，生成完整投递决策报告。

**一致性校验**：Fast 和 Deep 共享同一套分数语义（90-100 强匹配 → <60 不建议），差值 ≤ 9 通过一致性校验。

**前端表现**：
- 进入页面 → 展示 Fast 结果（秒级）
- 用户点击"生成深度分析" → 触发 Deep（后台异步，前端展示 Loading）
- Deep 完成后替换展示，但 Fast 结果仍保留可查看
- 结果隔离：`FindLatestByResumeAndJobAndTier` 按 tier 过滤，避免 Deep 结果污染 Fast 视图

---

## 四、三个核心杀手锏（把技术决策讲成产品洞察）

### 杀手锏一：投递边界与优化停止条件

**问题**：传统 AI 工具会无限给建议——"再优化一下""再补个项目"。学生不知道什么时候该停。

**方案**：
- `apply_readiness`：四档判断（可直接投 / 先改 Top 3 / 先补经历 / 信息不足无法判断）
- `optimization_stop_signal`：告诉用户"修完这 3 个就可以投了，剩下的投完再改"

**面试时怎么讲**："我做了一个设计决策：系统必须有边界感。不能让人无限优化简历，必须给明确的'可以投了'信号。"

### 杀手锏二：JD 证据链匹配

**问题**：AI 经常说"你适合这个岗位"，但不告诉你为什么适合。

**方案**：
- `jd_evidence_chain`：每条判断绑定 JD 要求 + 简历证据 + 匹配度判断
- `evidence_strength`：区分强证据（实习直接相关）、弱证据（课程涉及）、缺失证据

**面试时怎么讲**："系统不输出'你很匹配'这种话。每条结论都绑定具体的 JD 要求和简历证据。用户可以点击证据链逐条核实。"

### 杀手锏三：Agent 输出 Quality Gate

**问题**：LLM 输出可能编造经历、编造指标、给出不当评价。

**方案**：
- 4 级枚举（pass / revise_required / fallback_required / manual_review_suggested）
- 10 项检查维度（JD 依据检查、简历证据检查、编造检查、投递边界检查……）
- **fabrication_check 一旦失败直接 fallback**——零容忍
- revise 只给一次机会，不通过直接降级

---

## 五、面试中一定会被追问的技术细节

### 5.1 SEP 语义画像系统

**是什么**：统一为推荐/匹配/优化/对比四大功能提供语义输入底座。

**怎么做的**：
- Resume SEP：简历保存后异步生成语义标签（label/code/confidence/evidence），含 `content_hash` 比对新旧度
- Job SEP：岗位方向/技能/要求/责任标签的语义表示
- `SemanticProfileGate`：入口检查 profile 存在性 + 新鲜度

**数据**：9/9 简历 + 46/46 岗位全量就绪，Fast Match 4-5ms

### 5.2 Fast Match Engine（H18-D 重构）

**重构前**：纯标签计数 `min(98, 60 + tagScore*0.8)`，base 60 毫无意义

**重构后**（830 行新代码，8 大模块）：
1. **Normalization**：100+ 同义词归一化（"后端开发" ≡ "后台开发" ≡ "服务端开发"）
2. **Direction Fit**：5 级方向适配分析（精准匹配 / 相邻方向 / 可迁移 / 弱相关 / 方向错位）
3. **Hard Gate**：8 条硬门槛（学历硬门槛、实习硬要求、技能关键缺失……）
4. **Requirement Coverage**：8 项结构化覆盖评估（6 状态：strong/partial/transferable/weak/missing/unknown）
5. **Evidence Mapper**：5 级证据映射
6. **Risk Detector**：结构化风险检测（5 字段完整输出）
7. **Score Calibrator**：5 维度加权评分（方向权重 35%、技能 25%、项目 20%、学历 10%、表达 10%）
8. **Result Assembler**：结果组装 + 投递决策 + raw code 人类化

### 5.3 Deep 5 模块架构（JM 系列重构）

**重构前**：Deep v3 是纯规则引擎（无 LLM），内容浅、模板化

**重构后**：真实 Hy3 5 模块 LLM 生成（+450 行代码）

| 模块 | 回答的问题 | 输出内容 |
|------|-----------|---------|
| M1 Decision | 该不该投？ | 分数 + 5 种建议（可直接投递/修改后投递/备选/不建议/建议换方向） |
| M2 Pass Points | 为什么可能过？ | 简历命中点 + 真实证据绑定 |
| M3 Risks | 为什么可能被筛掉？ | 淘汰风险三分类（硬伤/证据不足/方向错误）+ 影响说明 + 修复方向 |
| M4 Rewrite | 简历怎么改？ | before/after 改写示例，基于学生真实经历 |
| M5 Strategy | 下一步投什么？ | 推荐更适合岗位方向，避免方向错位 |

### 5.4 前端四层数据消费模型

面试时被问到前端架构，可以这样讲：

```
API Client 层（client.ts）
    ↓  原生 fetch，code!==0 统一错误
API 模块层（jobs.ts / resumes.ts / agents.ts）
    ↓  业务 API + withFallback 策略
ViewModel Loader 层（viewModels.ts，9 个 Loader 函数）
    ↓  串联 API + Adapter，返回页面 ViewModel
页面组件层
    ↓  只消费 ViewModel，不直接调 API
```

**关键原则**：页面组件不直接调用 API，不直接理解后端原始结构。这是我在前端工程化中最重要的架构决策。

### 5.5 前端状态机

**推荐页状态机（9 状态）**：
```
booting → fast_loading → fast_done → deep_loading → deep_profile_pending → deep_done
                                                                              ↘ profile_failed
                                                                              ↘ deep_error
```

**匹配页状态机（6 状态）**：
```
idle → fast_loading → fast_done → deep_loading → deep_done
                                                     ↘ deep_error
```

每个状态都定义了主区域内容、右侧按钮、警告信息、deepResult 规则。

---

## 六、遇到过的最大技术难题（面试官最爱问）

### 问题 1：「语义画像割裂」—— 最核心的架构问题

**发现**（H18-A 审计）：推荐系统已用 SEP 语义画像，匹配系统仍用纯标签计数。导致匹配分数虚高，用户看到的匹配结果不可信。

**影响**：系统内两个核心功能使用完全不同的评价体系。

**解决**（H18-B→H20-A，8 阶段重构路线）：
1. 冻结 5 级分数语义 + 8 条 score cap
2. 接入 SEP Context 到 Fast Match
3. **重写 830 行匹配引擎**（8 大模块）
4. 前端报告页从 4 段式重构为 6 模块化卡片
5. 24 组测试验收，5 样例 ALL PASS

**面试怎么讲**：不要只说"重构了匹配引擎"。要讲出"为什么发现"和"怎么验证"。——"我用 5 份典型简历验证时发现，推荐和匹配用不同评价体系，这是系统性割裂。我花了两天做了全链路重构，统一到 SEP 底座上。"

### 问题 2：Deep 是假 LLM —— 最尴尬的发现

**发现**（JM-R0 审计）：岗位匹配的 Deep 分析，名义上是 LLM 生成的，实际上是纯规则引擎。没有真实 Hy3 调用。

**影响**：用户看到的"深度 AI 分析"是假的，内容模板化、泛泛而谈。

**解决**（JM-R3-R）：重写 450+ 行 `runHy3ScreeningPipeline`，实现真实 5 模块 Hy3 生成。3 个验收案例全部通过（78/33/42 分）。

**面试怎么讲**：诚实面对——"我做过一次基线审计，发现之前的 Deep 生成路径其实是规则管道。虽然不是我做错的，但我选择不回避，直接重写为真实 Hy3 调用。这是对产品质量的负责。"

### 问题 3：Fast Recommend 同步阻塞 —— P0 性能问题

**发现**（H20-C-1）：`fastRecommendWithSEP()` 中同步调用了 `EnsureResumeProfile()`，对没有 SEP 的简历触发同步 LLM 调用（7/9 简历受影响）。

**影响**：用户点"推荐岗位"后可能等几十秒。

**解决**（H20-C-2）：
- 改为 `FindLatestByTarget` 只读 + `ErrFastRecommendSEPMissing` 快速降级
- 新增 `EnsureResumeProfileAsync` 后台异步补齐
- `sync.Mutex` in-flight 去重

**结果**：Fast Recommend 50-83ms，Fast Match 4-5ms

### 问题 4：Deep 结果污染 Fast 视图 —— P0 展示问题

**发现**（H20-D-1）：`FindLatestByResumeAndJob` 不带 tier 过滤，Deep 结果成为 latest 后，Fast 页面默认展示 Deep 内容。

**解决**（H20-D-2）：
- 后端：新增 `FindLatestByResumeAndJobAndTier` + `normalizeMatchGenerationTier` + API tier 参数
- 前端：拆分 `fastResult` / `deepResult` / `deepExists`，mounted 时默认 fast

---

## 七、工程方法论（如果你被问「你是怎么管理项目的」）

**「审计→设计→实现→验收」四步闭环**：

```
阶段 A：只读审计 → 查现状，不修代码 → 输出：P0/P1/P2 问题清单
阶段 B：设计冻结 → 文档先行，写清"做什么/不做什么/验收标准"
阶段 C：工程实现 → 严格按冻结文档，不改范围外代码
阶段 D：验收 → 对照验收标准逐项检查
```

**P0/P1/P2 三级问题分级**：
- P0：必须立即修复（阻塞演示/上线）
- P1：应在当前阶段修复（影响产品完整性）
- P2：锦上添花（不阻塞核心功能）

**「先止血后治病」务实策略**：Fast 先修（用户最常见交互），Deep 后修（深度体验），不追求完美一步到位。

**550+ 开发日志**：每条 10 节标准格式，每个任务有编号可追溯。

---

## 八、关键数据速查（面试时脱口而出）

| 数据项 | 数值 | 面试时怎么用 |
|--------|------|------------|
| 开发周期 | **18 天** | "从零到可上线 Demo，18 天" |
| 后端代码 | **146 个 Go 源文件** | "零框架依赖，全靠标准库" |
| 前端代码 | **139 个 TSX + 40 个 TS** | "从 Figma Make 原型改造而来" |
| 文档规模 | **850+ Markdown** | "文档驱动的工程管理" |
| 开发日志 | **550+** | "每个任务可追溯" |
| API 接口 | **14 个 REST** | "覆盖认证/简历/岗位/Agent 四大域" |
| Agent | **4 主 + 3 子** | "四主 Agent 架构" |
| Prompt 模板 | **4 个定稿** | "每个都经过 2-3 轮迭代" |
| JSON Schema | **3 个（draft 2020-12）** | "给 AI 画格子" |
| 单元测试 | **132+ Go 测试** | "核心模块全覆盖" |
| LLM 错误码 | **20+** | "Pipeline 全覆盖" |
| 岗位样本 | **46 条真实腾讯校招岗位** | "不是造数据，是腾讯招聘的真岗位" |
| 示范简历 | **9 份** | "覆盖算法/后端/产品/初学等多种背景" |
| Fast Match | **4-5ms** | "用户等不了 30 秒" |
| Fast Recommend | **50-83ms** | "秒级初筛" |
| 前端 Build | **1.51s** | "Vite + React 工程化" |
| 变更记录 | **14,460 行 / 500+ 条目** | "不是简单 changelog，含根因分析和决策逻辑" |

---

## 九、3 个经典演示案例

| 案例 | 简历 | 岗位 | 分数 | 结论 | 面试时怎么讲 |
|------|------|------|:--:|------|------------|
| **Case 1** ⭐ | 周明远（211 后台开发，2 段实习 + 4 个项目） | 软件开发-后台开发 | **78** | 修改后投递 | "系统能识别他的 JWT/Redis/Docker 等经验，但也告诉他简历表达还需要强化" |
| **Case 2** | 同上 | 算法-自然语言处理 | **33** | 建议换方向 | "同一个简历投错方向，系统敢说不。帮学生节省投递成本" |
| **Case 3** | 纯小白简历（无实习、无深度项目） | 后台开发 | **42** | 不建议投递 | "不是方向错，是证据不足。系统告诉他先补什么项目" |

**三案例覆盖全部输出场景**：revise_before_apply / change_direction / not_recommended

---

## 十、面试高频追问及准备答案

### Q：LLM Pipeline 怎么保证输出稳定性？

**答**：五层防线——
1. **Prompt 约束**：强制 JSON + 枚举 + Schema 草案嵌入
2. **JSON 格式校验**：draft 2020-12，不允许 additionalProperties
3. **JSON Repair**：Hunyuan 修复一次，禁止递归
4. **Mock Fallback**：规则引擎兜底，用户永远看到可用结果
5. **Quality Gate**：10 维内容检查，fabrication 一旦失败直接 fallback

Pipeline 20+ 错误码全覆盖，用户侧始终看到稳定的业务结果。

### Q：怎么管理 Prompt 的迭代？

**答**：Prompt 作为独立产品资产管理。在 `server/prompts/` 目录下：
- 人工定稿 → CodeBuddy 登记 → 首轮输出验收（partial_pass）
- Schema 吸收首轮问题 → 最小一致性修复
- 三个 Prompt 各经历 2-3 轮迭代
- 每次修改都有开发日志记录

### Q：为什么不自动生成简历？

**答**：这是 ADR-004。两个原因：
1. **真实性保护**：AI 不应替用户编造经历，这在求职场景中是红线
2. **差异化定位**：市面上有很多 AI 代写简历的工具，但缺少帮用户"诊断和引导"的工具。我们给 before/after 改写示例，但改成什么样由用户自己决定。

### Q：项目最遗憾的地方？

**答**：版本对比 Agent 是唯一未 LLM 化的核心功能。时间不够，目前用的是规则引擎。如果继续做，我会用 LLM 做"迭代诊断"——比较 V1 和 V2 的语义变化，判断是真的提升了还是只是换了表述。

### Q：如果重新做，架构上会改什么？

**答**：SEP 应该在项目第一天就定义好。我是在 H18 审计时才统一语义画像底座的，如果一开始就有 SEP，可以少走很多弯路。另外全容器部署也走了弯路，中间回退到混合式，浪费了一天多。

### Q：你觉得自己最大的成长是什么？

**答**：学会了"文档驱动的工程管理"。以前我写代码是边想边写，在这个项目里我第一次体验到"先设计冻结、再实现、再验收"的威力——效率反而更高，因为减少了很多返工。550+ 条开发日志不是负担，是让我始终保持清醒的工具。

---

## 十一、快速回忆地图（面试前 10 分钟过一遍）

1. **项目名**：Offer 捕手
2. **一句话**：校招 AI 投递决策平台，不代写简历，帮用户判断该不该投/差在哪/怎么改
3. **技术栈**：Go 标准库 + React 18 + TypeScript + MySQL + 腾讯混元 Hy3 LLM
4. **核心功能**：岗位推荐 + **岗位匹配分析（主推）** + 简历优化 + 版本对比
5. **架构亮点**：
   - 六层 LLM Pipeline（Prompt→LLM→Parse→Validate→Repair→Fallback）
   - Fast/Deep 双阶段（4-5ms 秒级 + Hy3 深度）
   - Quality Gate 10 维质量门禁
   - SEP 语义画像统一底座
   - 前端四层数据消费模型
6. **工程亮点**：
   - 18 天完成 13 阶段 550+ 任务
   - 132+ 单元测试
   - 850+ 文档 + 550+ 开发日志
   - P0=0，所有阶段验收通过
7. **关键决策**：
   - ADR-001：Go 标准库，零框架
   - ADR-003：先 Mock 后 LLM
   - 4-3.5：战略转向（推荐工具→决策系统）
   - H18：Fast Match 830 行重构
   - JM：从规则引擎到真 Hy3 LLM
8. **Demo 案例**：Case 1（78 分/修改后投递）、Case 2（33 分/换方向）、Case 3（42 分/补项目）
9. **最大收获**：文档驱动的四步闭环（审计→设计→实现→验收）
10. **如果重来**：SEP 第一天就定义；不做全容器部署的弯路

---

> 这份文档旨在帮你快速回忆项目的每一个关键细节。面试官问到的任何技术点，你都能在对应的章节找到准备答案。祝你面试顺利！
