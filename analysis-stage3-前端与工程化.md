# 第 3 阶段分析：前端产品化与工程部署

> 分析时间：2026-06-23
> 源文档：`01-前端设计/`、`05-前端设计/`、`07-前端设计/`(48文件)、`06-工程部署/`(21文件)、`15-前端专业化/`

---

## 一、前端改造全景

### 1.1 起点：Figma Make 生成的原型

项目前端的起点是 Figma Make 自动生成的 React 代码。

| 维度 | 起点状态 | 判断 |
|------|---------|:--:|
| 页面组件 | 9 个核心页面齐全 | ✅ 覆盖全规划 |
| 视觉风格 | 腾讯校招岗位内页风格（浅灰 #F5F7FA、蓝色 #1D6FFF、白底卡片） | ✅ 风格到位 |
| 路由 | `useState` 切换页面，无真实 URL | ❌ 无法刷新/分享/前进后退 |
| 数据层 | 全部硬编码 Mock 数据在组件文件中 | ❌ 无 API Client |
| 类型定义 | 无独立 TypeScript 类型 | ❌ 数据结构与组件耦合 |
| 状态处理 | 无 Loading / Error / Empty 统一处理 | ❌ |
| 工程配置 | 无 tsconfig.json / lock 文件 / ESLint / Prettier | ❌ |

**核心判断：不重写视觉和页面，补路由、API Client、类型定义和状态管理。**

### 1.2 六阶段改造计划

```
第一阶段：路由与页面挂载
    ↓
第二阶段：统一 API Client
    ↓
第三阶段：类型定义与数据结构
    ↓
第四阶段：页面接入真实数据
    ↓
第五阶段：状态处理（Loading / Error / Empty）
    ↓
第六阶段：视觉与交互收敛
```

### 1.3 技术栈

| 技术 | 选择 | 说明 |
|------|------|------|
| 框架 | React 18.3 | 沿用 Figma Make 生成的代码 |
| 构建 | Vite 6.3.5 | 极速 HMR + 可靠构建（1.51s ~ 3.05s） |
| 语言 | TypeScript | .tsx 文件 |
| 样式 | TailwindCSS 4.1 + tokens.ts | 逐步从 inline style 收敛 |
| 路由 | react-router v7.13.0 | 创建 BrowserRouter + AppLayout |
| UI 组件 | shadcn/ui (Radix) | 逐步移除 MUI 冗余依赖 |
| 图标 | Lucide React | |
| HTTP | 原生 fetch（不引入 axios） | `client.ts` 封装 219 行 |

---

## 二、前端数据架构：四层消费模型

### 2.1 核心设计

```
后端 JSON 响应
    ↓
┌──────────────────────────────────────┐
│ API Client 层 (client.ts)           │ ← 底层 HTTP 请求，code!==0 统一错误处理
│   ↓                                  │
│ API 模块层 (jobs.ts/resumes.ts/      │ ← 业务 API 调用 + withFallback 策略
│             agents.ts)               │
│   ↓                                  │
│ ViewModel Loader 层 (viewModels.ts)  │ ← 串联 API + Adapter，返回页面 ViewModel
│   ↓                                  │
│ Adapter 层 (adapters/)               │ ← 纯数据转换，含 tone/label 预计算
│   ↓                                  │
│ 页面组件层 (pages/components)        │ ← 只消费 ViewModel，不直接调 API
└──────────────────────────────────────┘
```

**关键原则：页面组件不直接调用 API，不直接理解后端原始结构，只消费 ViewModel。**

### 2.2 API Client 基础层（5-2.1）

| 文件 | 行数 | 职责 |
|------|:--:|------|
| `api/types.ts` | 25 | HttpMethod / ApiRequestOptions / ApiClientConfig |
| `api/errors.ts` | 28 | ApiClientError + isApiClientError 守卫 |
| `api/client.ts` | 163 | 核心请求层（requestRaw / request / get / post / put / patch） |
| `api/index.ts` | 3 | 统一导出 |

**设计特点**：
- 使用**原生 fetch**（不引入 axios）
- 默认 base URL：`/api`（不是 `/api/v1`）
- 环境变量：`VITE_API_BASE_URL` 可覆盖
- `requestRaw<T>()` 返回完整 `{ code, message, data }`
- `request<T>()` 仅返回 `data`，`code !== 0` 抛 ApiClientError
- JSON 解析失败截断前 200 字符，避免大日志

### 2.3 业务 API 模块（5-2.2 / 5-2.3）

**Jobs API**（3 函数）：
| 函数 | 后端路径 | fallback |
|------|------|------|
| `getJobs()` | `GET /jobs` | `mockJobs` |
| `getJobDetail(jobId)` | `GET /jobs/{jobId}` | `getMockJobById()` |
| `getSelectedJob()` | `GET /jobs/{id}` | `mockSelectedJob` |

**Resumes API**（8 函数）：
| 函数 | 后端路径 |
|------|------|
| `getResumes()` / `getCurrentResume()` / `getResumeDetail()` | `GET /resumes/versions` 等 |
| `getResumeOptions()` / `compareResumes()` | 读取接口 |
| `createResume()` / `updateResume()` / `setCurrentResume()` | 写入接口 |

**Agents API**（3 函数，138 行）：
| 函数 | 路径 | 返回 |
|------|------|------|
| `runRecommendation(input?)` | `POST /jobs/recommend` | `AgentResponseEnvelope<RecommendationResult>` |
| `runMatchAnalysis(jobId, input?)` | `POST /jobs/{jobId}/match` | `AgentResponseEnvelope<MatchAnalysisResult>` |
| `runResumeOptimization(input?)` | `POST /resumes/optimize` | `AgentResponseEnvelope<ResumeOptimizationResult>` |

**统一 fallback 策略**：
```ts
// 网络错误 / HTTP 非 2xx / code!==0 → 自动 fallback
try { return await requestFn(); } catch { return fallbackFn(); }
```

### 2.4 ViewModel Loader 层（5-2.4）

| Loader 函数 | 返回类型 | 说明 |
|------|------|------|
| `loadJobListViewModels()` | `JobCardViewModel[]` | 岗位列表 |
| `loadJobDetailViewModel(jobId)` | `JobDetailViewModel` | 岗位详情 |
| `loadCurrentResumeSummaryViewModel()` | `ResumeSummaryViewModel` | 当前简历摘要 |
| `loadResumeOptionViewModels()` | `ResumeOptionViewModel[]` | 选择简历弹窗 |
| `loadResumeComparisonViewModel()` | `ResumeComparisonViewModel` | 版本比较 |
| `loadRecommendationViewModel()` | `RecommendationPageViewModel` | 推荐页 |
| `loadMatchAnalysisViewModel()` | `MatchAnalysisPageViewModel` | 匹配分析页 |
| `loadResumeOptimizationViewModel()` | `ResumeOptimizationPageViewModel` | 简历优化页 |

**Agent Loader 关键设计**：Loader 层完成 `response.result + response.meta` 拆包后传入 Adapter，页面只消费 ViewModel，不感知 `AgentResponseEnvelope`。

### 2.5 data.meta 前端消费规则

| agent_status | 前端展示文案 |
|-------------|------------|
| `success` | "AI 分析" |
| `repaired` | "AI 分析（已修复）" |
| `mock_success` | "规则分析"（降级提示） |
| `fallback` | "AI 降级分析" |

---

## 三、TypeScript 类型体系

### 3.1 类型文件清单（5-1.1，7 文件 434 行）

| 文件 | 行数 | 内容 |
|------|:--:|------|
| `tsconfig.json` | 21 | 首次新增 TypeScript 编译配置 |
| `types/api.ts` | 31 | ApiResponse / PaginatedResult / ApiErrorInfo |
| `types/agent.ts` | 188 | **核心**：AgentResponseMeta 8 字段 + 三类 Agent Result + 共享 TopFixItem / EvidenceItem |
| `types/job.ts` | 56 | Job / JobDetail / JobFilterState |
| `types/resume.ts` | 108 | Resume / ResumeOption / ResumeVersionComparison |
| `types/ui.ts` | 25 | BadgeTone / StatusBadge / PageAction / DecisionPanelState |
| `types/index.ts` | 5 | 统一导出 |

### 3.2 Agent 核心类型契约

`AgentResponseMeta`（与后端 8 字段一致）：

| 字段 | 类型 | 说明 |
|------|------|------|
| `generated_by` | `"llm_agent" \| "mock_agent"` | 生成来源 |
| `agent_status` | `"success" \| "repaired" \| "fallback" \| "mock_success" \| "error"` | 5 种状态 |
| `fallback_reason` | 11 种枚举 | 降级原因 |
| `error_code` | 12 种枚举 | 错误码 |
| `repair_applied` | `boolean` | JSON repair 标记 |
| `schema_valid` | `boolean` | Schema 校验标记 |
| `request_id` | `string` | 请求追踪 |

---

## 四、产品形态冻结

### 4.1 最终产品定位

> Offer 捕手不是普通 AI 聊天助手，也不是后台管理系统，而是在真实校园招聘岗位浏览和投递流程上增加一层 AI 投递前决策能力。

核心产品叙事：
1. 我适合哪些岗位？→ 岗位推荐
2. 这个岗位我现在能不能投？→ 岗位匹配分析
3. 投之前我的简历最应该改哪里？→ 简历优化

### 4.2 页面清单（P0 8 页 + P1 1 页）

| # | 页面 | 路由 | 优先级 |
|---|------|------|:--:|
| 1 | 岗位列表页 | `/jobs` | P0 |
| 2 | 岗位详情页 | `/jobs/:id` | P0 |
| 3 | 岗位推荐页 | `/recommend` | P0 |
| 4 | 岗位匹配分析页 | `/match` | P0 |
| 5 | 简历优化页 | `/resume-optimization` | P0 |
| 6 | 我的简历管理首页 | `/my-resumes` | P0 |
| 7 | 简历编辑/添加页 | `/my-resumes/edit` | P0 |
| 8 | 选择简历弹窗 | 弹窗组件 | P0 |
| 9 | 历史简历比较页 | `/resume-comparison` | P1 |

### 4.3 页面跳转关系

```
岗位列表页 → 岗位推荐页
岗位列表页 → 岗位详情页 → 岗位匹配分析页 → 简历优化页

Navbar「我的简历」
  → 我的简历管理首页
  → 修改简历 / 添加简历
  → 历史简历比较

岗位推荐 / 岗位匹配分析前
  → 选择简历弹窗 → 确认并开始分析
```

### 4.4 全局 Navbar 规则冻结

- **左侧**：固定「Offer 捕手」学生求职匹配智能体
- **右侧**：只显示「我的简历」
- **禁止**：岗位浏览、岗位推荐、首页、头像、用户名、下拉菜单、两层 Navbar
- **Navbar 不承担主流程导航**，主流程通过页面内按钮完成

### 4.5 统一视觉规则

```
白色主背景 / 浅灰页面底色 / 腾讯蓝强调色
白色卡片 + 浅灰边框 / 克制正式 / 真实招聘官网感 / 大量留白
```

**禁止**：黑色主题、过度 AI 科技风、花哨渐变、聊天框、右下角 AI 助手、复杂 Dashboard

### 4.6 简历表单字段边界

**保留**：求职方向、教育经历、实习经历、项目经历（项目名称/技术栈/项目描述/个人职责）、技能关键词、自我评价

**禁止**：头像上传、附件上传、身份证号、政治面貌、民族、户籍、家庭地址、紧急联系人、完整腾讯招聘字段体系

---

## 五、工程部署体系

### 5.1 混合式开发环境

```
Docker Compose：
  ├── MySQL 8.0（端口 3306，数据持久化）
  └── Nginx（alpine，端口 80，统一入口）

宿主机后台进程：
  ├── Backend：localhost:8080（go run）
  └── Frontend：localhost:5173（npm run dev）

Nginx 代理规则：
  http://localhost/         → Frontend :5173
  http://localhost/api/     → Backend :8080（不 strip /api）
  http://localhost/healthz  → Backend :8080/healthz
```

**关键说明**：
- 非全容器模式，Docker Compose 只管理 MySQL + Nginx
- Backend/Frontend 由 Makefile 以本机后台进程管理
- Nginx 通过 `host.docker.internal` 代理宿主机服务
- 通过 Nginx 访问时前后端同源，**无需 CORS**

### 5.2 一键启动命令

| 命令 | 功能 |
|------|------|
| `make up` | 启动完整环境（ensure-env → doctor → init → infra-up → backend-up → web-up → health） |
| `make down` | 关闭环境，**不删 volume** |
| `make restart` | 软重启 |
| `make status` | 完整环境状态 |
| `make health` | 聚合健康检查 |
| `make logs` | 日志入口 |

**进程管理**：

| 服务 | PID 文件 | 日志文件 | 端口 |
|------|---------|---------|------|
| Backend | `.run/server.pid` | `.logs/server.log` | 8080 |
| Frontend | `.run/web.pid` | `.logs/web.log` | 5173 |

**关键修复**（4-3.7.B.R.2.1）：`make down` 和 `make restart-soft` 递归清理进程树，确保 5173 端口不残留。

### 5.3 .env 环境变量管理

| 文件 | 说明 | 提交 Git |
|------|------|:--:|
| `.env.example` | 模板，只放占位符（14 项 LLM + 15 项腾讯招聘 + 基础配置） | ✅ |
| `.env.local` | 本地真实配置（含 LLM_API_KEY） | ❌ |
| `.env` | 运行时配置，由 `make ensure-env` 自动生成 | ❌ |

**自动复制规则**：
1. `.env` 不存在且 `.env.example` 存在 → 自动复制
2. `.env` 已存在 → 跳过，不覆盖
3. 优先从 `.env.local` 复制，其次从 `.env.example` 复制

### 5.4 CORS 策略（4-3.7.B.R.3 环境变量化）

| 配置项 | 默认值 |
|------|------|
| `CORS_ALLOWED_ORIGINS` | `http://localhost,http://localhost:5173,http://127.0.0.1,http://127.0.0.1:5173` |
| `CORS_ALLOWED_METHODS` | `GET, POST, PUT, PATCH, DELETE, OPTIONS` |
| `CORS_ALLOWED_HEADERS` | `Content-Type, Authorization, X-Requested-With` |
| `CORS_ALLOW_CREDENTIALS` | `false` |

**安全约束**：不允许的 Origin 不返回 CORS 头，生产环境必须只允许正式域名。

### 5.5 Makefile 体系（40+ 命令）

| 类别 | 命令 |
|------|------|
| 环境管理 | `up / down / restart / restart-soft / status / ps / health` |
| 日志 | `logs / backend-logs / frontend-logs / nginx-logs / db-logs` |
| 数据库 | `db-check / db-reset` |
| LLM | `llm-check / llm-smoke / agent-check / agent-env` |
| 岗位数据 | `jobs-fetch / jobs-import / jobs-status / jobs-clean` |
| 构建 | `backend-build / frontend-build` |

### 5.6 生产部署边界

| 维度 | 开发环境 | 生产部署 |
|------|---------|---------|
| Backend/Frontend | 本机后台进程 | 需构建镜像或托管 |
| HTTPS | ❌ 无 | ✅ 必须配置 |
| CORS | 允许 localhost | 只允许正式域名 |
| JWT_SECRET | 示例值 | ✅ 强随机值 ≥32 字符 |
| DB_PASSWORD | 默认 password | ✅ 强密码 |
| reset 命令 | 可用 | ❌ 禁用 |
| init.sql DROP TABLE | 允许 | 🔴 绝对禁止执行 |

---

## 六、工程执行规范体系

### 6.1 任务颗粒度规则

每次 CodeBuddy 任务必须满足：
1. **一个 Prompt 只做一个明确工程任务**
2. 不允许跨阶段做事
3. 不允许顺手修复无关问题
4. 不允许自行新增产品功能
5. 发现问题只记录到文档，不自动修复

### 6.2 执行前/中/后要求

**执行前**：阅读说明书 → 项目状态 → 下一步任务 → 变更记录 → 最近开发日志
**执行中**：只修改授权文件，不确定先记录后提问
**执行后**：更新项目状态 + 下一步任务 + 变更记录 + 创建开发日志（10 节标准格式）

### 6.3 15 项绝对禁止事项

1. 新增多简历管理 / 文件上传 / 自动生成简历
2. 新增岗位收藏 / 导出报告 / 投递追踪 / 面试模拟
3. 新增社区功能 / HR 后台 / 复杂权限 / 移动端适配
4. 修改 git config / 执行 git push --force
5. 删除用户已有文件

### 6.4 安全红线

- API Key 不得写入 `.env.example`、docs、log、Git、聊天记录
- `LLM_ENABLE` 默认 `false`
- `.env` 不入 Git
- `make doctor` / `agent-env` / `status` 不打印 API Key
- `LLM_LOG_RAW_OUTPUT` 默认 `false`

---

## 七、前端专业化迭代（H20 + K9 系列）

### 7.1 K9 前端状态机设计

**岗位推荐页状态机**（9 状态）：

```
booting → fast_loading → fast_done → deep_loading → deep_profile_pending → deep_done
                                                                              ↘ profile_failed
                                                                              ↘ deep_error
```

**岗位匹配页状态机**（6 状态）：
```
idle → fast_loading → fast_done → deep_loading → deep_done
                                                             ↘ deep_error
```

### 7.2 H20 前端精修系列

| 编号 | 任务 | 内容 |
|------|------|------|
| H20-E-2 | 统一 UI 组件 | 5 个组件：UnifiedAgentLoadingOverlay / AppButton / SectionCard / StatusBadge / AiNotice |
| H20-E-3 | 视觉语言对齐 | 匹配页接入统一组件：1440px + 蓝调科技感 Loading + 统一按钮/标签 |
| H20-E-3R | 视觉平衡修复 | Fast 页面新增核心三卡 + 优先级摘要 + 匹配依据摘要 |
| H20-E-3R-T | 排版对齐 | 全量文字缩小对齐推荐页标准（Score 48→36px、标题 text-base→14px） |
| H20-E-3R2 | 数据感知 UI | Fast 引擎不生产 Deep 专属字段时，自适应渲染 + 空状态降级为摘要卡 |

### 7.3 JM-FIX2 页面重构

围绕 Resume 2（周明远） × Job 10001 的主案例重建页面：
- 新增 B 核心三卡
- 更新 A-F 区域标题
- 改写加载弹窗为 5 步骤
- 清理不存在经历
- npm build 1.84s

### 7.4 推荐页 UX 优化（Recommendation-UX 系列）

| 优化 | 内容 |
|------|------|
| tagLabelMap | 120+ 映射，零 raw code 泄露 |
| Deep Loading 超时 | 90s → 900s，适配 Cascade 深度分析 |
| 卡片降噪 | tags ≤ 3，evidence ≤ 2 |
| 右侧按钮 redesign | 取消等待选项，增加"生成深度推荐"按钮修复（H21-B） |

---

## 八、前端工程数据

### 8.1 代码规模

| 文件类型 | 数量 | 说明 |
|---------|:--:|------|
| API 层 | 8 文件 656 行 | client + errors + types + jobs + resumes + agents + viewModels + index |
| TypeScript 类型 | 7 文件 434 行 | api + agent + job + resume + ui + index + tsconfig |
| Adapter 层 | 4 文件 | 数据转换 + tone/label 预计算 |
| Mock 数据 | 4 文件 | 标准化 Mock，从组件中剥离 |
| 页面组件 | 139 个 TSX | 含 9 个核心页面 + 大量业务组件 |
| 构建产物 | 560K dist | build 1.51s ~ 3.05s |

### 8.2 验收数据

| 验收项 | 结果 |
|------|------|
| 阶段 5 总验收 | ✅ 前端静态闭环/build/Demo/文案/联调前准入全部完成 |
| 演示路径验证 | 3 条路径 11 URL 全部 HTTP 200 |
| Build 准入 | P0 解除，13 项全部通过 |
| 穿帮点清理 | 10 项问题修复，4 个 legacy 组件删除 |
| P0/P1/P2 | P0=0, P1=0, P2=6 |

### 8.3 阶段 5 交付物清单

```
✅ 前端 TypeScript 数据契约（7 文件）
✅ Mock 数据标准化（从组件中剥离）
✅ Adapter/Selector 统一
✅ 页面硬编码替换为统一 Mock
✅ API Client 基础层（4 文件 219 行）
✅ Jobs/Resumes API 模块（2 文件 175 行）
✅ Agents API 模块（1 文件 138 行）
✅ ViewModel Loader 层（1 文件 120 行，9 函数）
✅ 全局布局/Navbar 统一（4 冗余组件废弃）
✅ 简历模块路由补齐
✅ 5 页骨架统一（1440px max-w）
✅ SelectResumeModal 全局接入
✅ 三页 URL 参数真实传递
✅ 简历管理操作闭环
✅ 前端工程完整性核验（4 层 29 文件）
✅ Build 准入验收（13 项）
✅ 静态 Demo 演示验收（3 路径）
✅ 视觉与文案穿帮点清理（10 项）
```

---

## 九、部署与安全合规

### 9.1 腾讯招聘数据源接入

| 维度 | 数据 |
|------|------|
| 接口确认 | 3 个（列表 POST + 详情 GET + 字典 GET） |
| 配置项 | 15 项 TENCENT_JOBS_* 环境变量 |
| 拉取脚本 | `tencent_jobs.sh`（9 子命令） |
| Makefile | `make jobs-*`（11 个 target） |
| 入库脚本 | `tencent_jobs_import.sh`（7 子命令） |
| 岗位数据 | 46 条手动精选岗位（162K JSON） |
| 安全约束 | 不使用 Cookie/Token/登录态，低频 500ms 间隔 |

### 9.2 环境恢复流程

CNB 云原生环境重启后：
```bash
make doctor    # 检查环境
make init      # 补齐 .env + 创建目录
make up        # 一键启动
make health    # 健康检查
```

### 9.3 上线前检查清单

| # | 检查项 |
|---|--------|
| 1 | JWT_SECRET 更换为强随机值（≥32 字符） |
| 2 | DB_PASSWORD 更换为强密码 |
| 3 | CORS_ALLOWED_ORIGINS 只允许正式域名 |
| 4 | .env 不入 Git |
| 5 | 真实 LLM_API_KEY 只保存在服务器环境变量 |
| 6 | 配置 HTTPS + SSL 证书 |
| 7 | 配置正式域名 + DNS 解析 |
| 8 | 确认数据库备份策略 |
| 9 | 禁止生产执行 reset-up |
| 10 | 🔴 禁止在生产执行 init.sql（含 24 行 DROP TABLE IF EXISTS） |

---

## 十、总结：前端与工程核心亮点

1. **从原型到工程化**：从 Figma Make 生成的 9 个页面原型起步，通过 6 阶段改造（52 个子任务）建立完整前端工程体系
2. **四层数据消费模型**：API Client → API 模块 → ViewModel Loader → 页面组件，层次分明，页面不直接理解后端原始结构
3. **原生 fetch 零依赖**：不引入 axios、Redux/Zustand、Gin/Echo，极致精简
4. **统一 Agent 消费**：`AgentResponseEnvelope<TResult>`（result + meta）封装，5 种 agent_status 前端展示规则
5. **产品形态冻结**：9 个页面完整定义，Navbar 规则固定，视觉规范冻结，Demo 数据统一
6. **混合式开发环境**：Docker MySQL + Nginx + 宿主机 Go/Vite，一键启动/关闭/重启
7. **安全工程体系**：15 项绝对禁止事项 + API Key 保护规则 + 环境变量分级管理 + 上线前 10 项检查
8. **40+ Makefile 命令**：覆盖开发全流程（环境/日志/数据库/LLM/岗位数据/构建）
9. **前端的 120+ tagLabelMap**：零 raw code 泄露，所有用户可见标签完全 humanize

---

## 十一、进入第 4 阶段分析

第 4 阶段将分析：开发迭代与关键决策（`09-开发日志/`、`变更记录.md`、`08-验收报告/`、`10-Agent联调/`等），并提炼项目管理与工程方法论亮点。

