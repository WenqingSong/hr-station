# 全项目错误案例汇编

> 共 60 个案例，按严重程度排列。标记 ⭐ 的建议重点扩展。

---

## P0 级：严重错误（阻塞上线/演示）

### 1. Prompt 加载失败 → 所有 Agent 走 fallback ⭐
- **现象**：三类 Agent 接口均返回 `fallback_reason=prompt_load_failed`
- **根因**：`backend.sh` cd 到 `server/` 执行，`PromptDir="server/prompts"` 解析为 `/workspace/server/server/prompts`（路径重复）
- **修复**：`prompt.NewLoader` 加 `resolveDir`，路径不存在时自动去掉 `"server/"` 前缀

### 2. LLM Smoke 环境变量加载失败 ⭐
- **现象**：`.env` 配置正确但 `make llm-check` 读不到 LLM 配置
- **根因**：`source .env` 遇到 `DB_DSN` 中的括号 `tcp(127.0.0.1:3306)` 被 bash 当作子 shell 语法，静默失败
- **修复**：改用逐行 `while read` 解析（`common.sh` 的 `load_env()`）

### 3. Fast Recommend 同步调用 LLM 阻塞 ⭐
- **现象**：Fast Recommend 对无 SEP 缓存的简历触发同步 Hy3 LLM（7/9 简历受影响，10-60s）
- **根因**：`fastRecommendWithSEP()` 同步调用 `EnsureResumeProfile()`
- **修复**：改为只读 `FindLatestByTarget` + 异步补齐 + SEP missing 降级到 TagEvidenceService

### 4. 岗位匹配 Fast/Deep 结果不隔离 ⭐
- **现象**：Deep 结果成为 latest 后页面默认展示 Deep，覆盖 Fast 结果
- **根因**：`FindLatestByResumeAndJob` 不带 `generation_tier` 过滤；前端单一 `activeResult` 状态
- **修复**：后端新增 tier 过滤 + API `tier` 参数；前端拆分为 `fastResult`/`deepResult`

### 5. 首页黑屏无限加载
- **现象**：浏览器 `/` 黑屏，API 超时
- **根因**：Go 后端进程在 16+ 次 API 调用后僵死
- **修复**：`make stop && make up` 重启（零代码修改）

### 6. 推荐页运行时崩溃（useMemo + Hook 顺序）
- **现象**：`useMemo is not defined` + `Rendered more hooks` crash
- **根因**：(a) `useMemo` 未 import；(b) `useState/useEffect` 放在 early return 之后
- **修复**：补全 import + 提升 hooks 到 early return 之前

### 7. 推荐页 React StrictMode 崩溃
- **现象**：StrictMode double-mount 触发 Hook crash
- **根因**：`useEffect` 内调 `navigate` 在 double-mount 时 hook 注册顺序不一致
- **修复**：移除 useEffect 中的 navigate，改用 `useMemo` 解析 URL

### 8. 推荐页 Link 跳转白屏
- **现象**："进入岗位详情"按钮点击后 500-3000ms 白屏
- **根因**：使用原生 `<a href>` 而非 React Router `<Link>`，触发完整浏览器重载
- **修复**：`<a href>` → `<Link to>`（2 行改动）

### 9. 前端 JobFilterSidebar 崩溃（null length）
- **现象**：`TypeError: Cannot read properties of null (reading 'length')`
- **根因**：API facets 字段为 `null`，代码无空值守卫
- **修复**：引入 `safeItems<T>()` 工具函数

### 10. 指定版本编辑页加载失败
- **现象**：`/resumes/edit?versionNo=1` 显示"加载失败"
- **根因**：API 返回 `expected_cities` 为字符串 `""`，adapter 对字符串调用 `.join()` 抛 TypeError
- **修复**：`Array.isArray()` 先判断再访问数组方法

### 11. 双模型配置丢失
- **现象**：重启后四核心功能回退到单一 `hy3-preview` 模型
- **根因**：修改了运行时 `.env` 但遗漏同步 `.env.example`
- **修复**：补齐模板 + `ensure_env.sh` 新增幂等自愈函数

### 12. go.mod Go 版本不匹配 + 启动崩溃
- **现象**：go 1.25.0 但环境 1.24.1；`LLM_ENABLE=true` 无 Key 时 `log.Fatalf` 崩溃
- **修复**：go.mod 降级到 1.24.1；LLM_ENABLE 默认改为 `false`

### 13. 全项目乱码修复
- **现象**：DB 6 表 15 字段 COMMENT 乱码（双重编码）+ 前端源码含 GBK 字符
- **修复**：修复所有 COMMENT + 重建受损文件 + 新增 `check-encoding.sh` 防线

### 14. 简历优化学历建议 P0 ⭐
- **现象**：普通后台开发岗位输出学历焦虑建议（如建议专升本）
- **根因**：Prompt 约束不完整 + 后端过滤缺失
- **修复**：Prompt 约束（3 文件）+ 后端过滤（8 函数/6 调用点/覆盖 9 字段）

### 15. resumes CRUD 全部 500/405/404
- **现象**：POST 返回 500，GET 返回 405，PUT 返回 404
- **根因**：SQL 占位符数量不匹配（13→12）+ 路由未注册 GET/PUT
- **修复**：新增关系型 DTO + 修复占位符 + 注册完整路由

### 16. 数据库 users 表 schema 不一致 → 注册 500
- **现象**：所有注册返回 500
- **根因**：DB 表与 init.sql 不一致：email 缺 DEFAULT + 缺 uk_users_phone 唯一索引

### 17. init.sql 被 mysqldump warning 污染
- **现象**：`make up` 导入 init.sql 失败，SQL syntax error
- **根因**：`mysqldump -p > file.sql` 将 stderr warning 写入 SQL 第 1 行
- **修复**：Python 清理污染 + 用 `MYSQL_PWD` 代替 `-p`

### 18. Resume Adapter project/projects 字段不兼容
- **现象**：项目经历数据全部丢失
- **根因**：API 返回 `project`(singular)，adapter 读 `projects`(plural) → null
- **修复**：新增 `getProjectsSafe()` fallback 函数

### 19. 彻底移除旧版 LoginPage 用户可见链路
- **现象**：`/login?redirect=%2Fresumes` 进入旧独立登录页
- **根因**：保留了两套登录 UI：弹窗式 + 全页面式，RequireAuth 重定向到 `/login`
- **修复**：统一为 Navbar + AuthDialog 弹窗式

### 20. 岗位推荐 Deep 生成失败
- **现象**：`POST quality` 返回 pending + 空 jobs，前端显示"推荐失败"
- **根因**：quality profile stale → 异步重建 → 返回 pending → early return
- **修复**：pending 时 fallback 到 fast profile

---

## P1 级：重要问题

### 21. 逻辑漏洞：修改简历=新建简历 ⭐
- **现象**：用户在简历编辑页保存修改，实际创建了新版本而非修改当前版本
- **根因**：前端 `handleSave` 始终调用 `createResume()`，未区分 create/update
- **修复**：edit 模式走 `updateResume(id)`，add 模式走 `createResume()`

### 22. AuthDialog forwardRef 警告
- **现象**：控制台 `Function components cannot be given refs`
- **根因**：4 个 Radix wrapper 组件未用 `React.forwardRef`
- **修复**：4 组件改为 `React.forwardRef` + `displayName`

### 23. Adapter 10 处 .map() 缺 null 守卫
- **现象**：三个 adapter 的 `.map()` 调用缺少 `Array.isArray` 守卫
- **修复**：新增 10 处 `Array.isArray` 检查

### 24. 真实 LLM 调用超时
- **现象**：`error_code=LLM_TIMEOUT`
- **根因**：API Key 无效或 Model 不可用，TokenHub 挂起 30s 超时

### 25. 开发环境重启后回到 mock
- **现象**：重启后所有 Agent 全是 mock（`generated_by=mock_agent`）
- **根因**：三层叠加：`ENABLE_MOCK_AGENT=true` + `LLM_FALLBACK_ENABLE=true` + `job_profiles` 表缺失

### 26. 首页岗位列表 401
- **现象**：首页岗位列表接口返回 401
- **修复**：修复 handler 鉴权逻辑

### 27. `resume_optimization_results` 表缺失
- **修复**：init.sql 补齐 DDL

### 28. `matchHighlights is not defined` 编译错误
- **修复**：RightSidebar 补齐 3 行 prop 传递

---

## P2 级：体验/边界问题（精选）

### 29. 乱码文本截断 ⭐
- **现象**：简历优化输出中出现 U+FFFD 乱码
- **根因**：`focus[:80]` byte-slice 截断中文多字节字符
- **修复**：`[]rune(focus)[:80]` rune-safe 截断

### 30. SEP Profile 单个坏 tag 导致整份 profile 失败
- **根因**：空 `evidence_text` 被 validator 拒绝
- **修复**：新增 `repairSemanticProfileTags()` 自动修复或丢弃坏 tag

### 31. 语言考试 key={idx} 缺陷
- **现象**：删除中间项时 React 错误复用 DOM
- **修复**：`key={idx}` → `key={le.id}`

### 32. 子模块全删空状态闪烁
- **现象**：删除最后一项时空状态被阻塞
- **修复**：三状态机（hasItems/removingLastItem/empty）

### 33. 简历编辑页折叠动画 exit 缺失
- **现象**：7 类列表 `filter` 直接 unmount，无 exit 动画

### 34. Navbar 背景穿透（白色半透导致蓝色穿透）
- **现象**：`bg-white/58` 半透 + Hero 蓝色 → 42% 蓝色透过形成异常浅蓝带
- **修复**：Hero 顶部改为 `bg-slate-950/55` 深色透明

### 35. 岗位匹配 Fast View 数据感知渲染问题
- **现象**：Fast SEP 不生产 `screeningPassPoints` 等 Deep 专属字段
- **修复**：数据感知渲染 + fastCardCount 自适应 + 空数据降级

### 36. 语义画像权重精度不足
- **现象**：`weight DECIMAL(3,2)` 精度不足导致 UpsertTags 失败但未回滚
- **修复**：增加 tag count 后校验 + CreateProfile 重复键自动更新

---

## 统计

| 等级 | 数量 | 主要类别 |
|:--:|:--:|------|
| P0 | 20 | 崩溃/白屏/数据丢失/全链路阻断 |
| P1 | 8 | 功能异常/环境配置/错误处理 |
| P2 | 8 | UI 异常/边界条件/体验问题 |

### 类型分布
- 前端运行时崩溃：5 个
- 环境/路径/配置：5 个
- 数据兼容性/空值：4 个
- API 路由/鉴权：4 个
- Agent/LLM 链路：4 个
- 数据库/乱码：4 个
- UI 视觉/体验：4 个
- 其他：6 个
