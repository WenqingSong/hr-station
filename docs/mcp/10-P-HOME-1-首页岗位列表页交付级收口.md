# 10-P-HOME-1 首页岗位列表页交付级收口

## 1. 阶段信息

* 阶段：10-P-HOME-1
* 日期：2026-06-02
* 范围：首页岗位列表页产品化（JobList/JobCard/JobFilterSidebar/AIDecisionWorkspace）
* 是否修改业务代码：是
* 是否修改前端代码：是
* 是否修改后端代码：否
* 是否修改数据库：否
* 是否接真实 LLM：否

## 2. 本次结论摘要

* 首页产品感提升：标题从"岗位列表"改为"找到更适合你的实习岗位"，副标题强调 AI 筛选能力
* AIDecisionWorkspace **三种状态**区分：未登录引导登录/注册、无简历引导创建简历、有简历显示简历摘要+推荐+选中岗位操作
* 选中岗位后可直达"查看详情"和"分析匹配度"，缩短主路径
* 岗位卡片信息层级更清晰：title → subtitle(department/type/location/status) → description → skills → actions
* 筛选器保持真实 facets API，视觉未改（已足够产品化）
* 零 Mock/演示/StyleGuide 反弹，零 href internal link
* npm build 1.50s + typecheck 零错误 + go build/test 全部通过
* 建议继续 10-P-6-3

## 3. 执行前问题

| 问题 | 证据 | 影响 |
| -- | -- | -- |
| AIDecisionWorkspace 无 auth 检测 | `AIDecisionWorkspace.tsx:85` `!resumeVm` 混淆"未登录"和"无简历" | 未登录用户看到"暂无可用的简历"而非登录引导 |
| 首页标题产品感不足 | `JobList.tsx:45` `岗位列表` | 像后台系统而非 AI 产品 |
| 选中岗位后缺少快捷操作 | 仅显示岗位名称，无匹配分析入口 | 从首页到匹配路径偏长 |

## 4. 首页结构调整

| 组件 | 修改内容 | 用户影响 |
| -- | ---- | ---- |
| JobList header | 标题→"找到更适合你的实习岗位"，副标题→"基于你的简历、技能和岗位需求，筛选值得优先关注的机会" | 首页更有产品感 |
| JobCard subtitle | department/type/location 移至 subtitle 行，紧凑内联，用分隔符 | 信息层级更清楚 |
| JobCard skills | 最多显示 4 个技能标签，超出显示"+N" | 避免技能标签溢出 |

## 5. 左侧筛选区调整

保持原有 `JobFilterSidebar` 不变。已有真实 facets API、sticky 定位、loading/error/empty 态，符合产品化要求。

## 6. 岗位列表 / 岗位卡片调整

| 项 | 修改前 | 修改后 |
| - | --- | --- |
| 标题 | `岗位列表` | `找到更适合你的实习岗位` |
| 副标题 | `浏览当前开放岗位，结合当前简历...` | `基于你的简历、技能和岗位需求，筛选值得优先关注的机会` |
| 卡片信息行 | Badge 堆叠 | 内联：`部门 | 类型 | 📍城市 · 状态` |
| 技能标签 | 全量展示 | 最多 4 个，超出 +N |
| 查看详情按钮 | 灰色 text variant | 蓝色 text variant，更突出 |

## 7. 右侧 AI 决策区状态设计

### 状态 A：未登录

```
显示：登录后使用 AI 求职决策
     Offer 捕手会基于你的简历推荐更适合的岗位，并分析匹配度。
按钮：登录 / 注册新账号
```

实现：`AIDecisionWorkspace` 通过 `getAuthToken()` 检测登录态。

### 状态 B：已登录但无简历

```
显示：你还没有可用于分析的简历
     先创建一份简历，才能进行岗位推荐、匹配分析和优化建议。
按钮：创建简历 → /resumes/new
```

### 状态 C：已登录且有简历

```
简历区：显示当前简历摘要（name/direction/skills/projectCount）+ 管理简历链接
推荐区：基于简历找岗位 → 推荐最适合的 5 个岗位
已选岗位区（当 selectedJobId 存在时）：岗位名称 + 查看详情 + 分析匹配度
```

## 8. 首页主路径核验

| 路径 | 入口 | 状态 |
| -- | -- | -- |
| 岗位列表 → 岗位详情 | JobCard "查看详情" → `/jobs/:id` | ✅ |
| 岗位列表 → 选择岗位 → 匹配分析 | AIDecisionWorkspace "分析匹配度" → `/match?jobId=X&resumeId=X` | ✅ 本次新增 |
| 岗位列表 → 推荐岗位 | AIDecisionWorkspace "推荐最适合的 5 个岗位" → SelectResumeModal → `/recommend` | ✅ |
| 首页右侧 → 创建简历 | 状态 B "创建简历" → `/resumes/new` | ✅ |
| 首页右侧 → 登录 | 状态 A "登录" → `/login` | ✅ 本次新增 |
| 首页右侧 → 注册 | 状态 A "注册新账号" → `/register` | ✅ 本次新增 |

## 9. 数据源核验

| 数据 | 来源 | 是否真实 |
| -- | -- | -- |
| 岗位列表 | GET /api/jobs | ✅ |
| facets | GET /api/jobs/facets | ✅ |
| 当前简历 | GET /api/resumes/current | ✅ |
| 岗位详情 | GET /api/jobs/{id} | ✅ |
| 登录态 | getAuthToken() → localStorage | ✅ |

## 10. Mock / 演示反弹检查

| 检查项 | 结果 |
| --- | -- |
| StyleGuide / AIBadge / mocks | 零匹配 ✅ |
| "Mock 演示"/"演示数据" | 零匹配 ✅ |
| withFallback / withMockFallback | 零匹配 ✅ |
| 内部 `<a href>` | 零匹配 ✅ |

## 11. 验证结果

| 命令 | 结果 |
| -- | -- |
| `git status --short` | 3 files M (AIDecisionWorkspace/JobCard/JobList) |
| `npm run build` | ✅ 1.50s, 1677 modules |
| `npm run typecheck` | ✅ 零错误 |
| `go build ./cmd/server` | ✅ PASS |
| `go test ./...` | ✅ 11 packages PASS |
| Mock 反弹 | ✅ 零匹配 |
| href 反弹 | ✅ 零匹配 |

## 12. 修改文件清单

| 文件 | 修改内容 | 是否业务代码 |
| -- | ---- | ------ |
| `web/src/app/components/AIDecisionWorkspace.tsx` | 新增 auth 检测、三种状态 UI、选中岗位快捷操作 | 是 |
| `web/src/app/components/JobCard.tsx` | subtitle 行紧凑内联、技能标签限制 4 个、"查看详情"蓝色 | 是 |
| `web/src/app/components/JobList.tsx` | 标题/副标题产品化文案 | 是 |

## 13. 当前 P0 / P1 / P2 / P3

### 13.1 P0

当前未发现首页 P0。

### 13.2 P1

| 编号 | 问题 | 证据 | 是否阻塞 | 下一步 |
| -- | -- | -- | -- | -- |
| — | 无 P1 | — | — | — |

### 13.3 P2 / P3

| 编号 | 问题 | 等级 | 证据 | 是否阻塞 | 下一步 |
| -- | -- | -- | -- | -- | -- |
| P2-MOBILE | 首页无移动端适配 | P2 | Navbar 1360px vs 列表 820px 固定宽度 | 否 | 后续 |
| P1-LN | lint 缺失 | P1 | 无 ESLint | 否 | 已记录 |

## 14. 未做事项

* 未修改后端代码
* 未修改数据库 schema
* 未新增接口
* 未接真实 LLM
* 未修改无关页面
* 未改整体 layout
* 未引入 mock/fallback
* 未执行任何版本控制写入操作

## 15. docs 更新清单

| 文件 | 新增 / 更新 | 目的 |
| -- | ------- | -- |
| `docs/mcp/项目上下文.md` | 更新 | 当前阶段 10-P-HOME-1 |
| `docs/mcp/10-P-HOME-1-首页岗位列表页交付级收口.md` | 新增 | 本阶段 MCP |
| `docs/09-开发日志/2026-06-02-264-10-P-HOME-1-首页收口.md` | 新增 | 开发日志 |

## 16. CodeBuddy 重启恢复摘要

首页岗位列表页产品化完成。AIDecisionWorkspace 三种状态区分（未登录/无简历/有简历），岗位卡片信息层级优化，首页标题产品化。零 mock/演示反弹。3 文件修改，+122 -79 行。

下次继续前读取：`docs/mcp/项目上下文.md`、`docs/mcp/10-P-HOME-1-首页岗位列表页交付级收口.md`。

## 17. 最终判断

1. 首页首屏产品感是否提升？是。标题/副标题/AI 决策区三种状态。
2. 左侧筛选是否真实且清晰？是。facets API 驱动，未改。
3. 岗位卡片层级是否更专业？是。内联 subtitle + 限制技能标签。
4. 右侧是否区分未登录/无简历/有简历？是。三种状态 + 快捷操作。
5. 选中岗位能否联动右侧？是。查看详情 + 分析匹配度。
6. 无简历时是否引导创建？是。"创建简历"按钮。
7. 未登录时是否引导登录？是。"登录"/"注册新账号"。
8. 是否引入 mock/演示？否。零匹配。
9. npm build/typecheck 是否通过？是。
10. go build/test 是否通过？是。
11. 是否建议首页专项通过？是。
12. 是否建议继续 10-P-6-3？是。
