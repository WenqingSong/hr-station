# Recommend-Page-Fix-B：修复岗位推荐页 resumeId 与登录用户不匹配

日期：2026-06-11

## 背景

`Recommend-Page-Audit-A` 定位：浏览器访问 `/recommend?resumeId=1` 时，若登录用户（Arthur, id=1）无简历或 `resumeId` 不属于当前用户，前端静默失败或进入 `PageLoading` 空转状态。

## 修复内容

### 1. URL resumeId 归属校验
- 当 URL `resumeId` 不在当前用户的 `resumeOptions` 中时 → `urlResumeNotFound=true`
- 页面显示明确错误："简历不存在或不属于当前账号" + "返回我的简历" 按钮
- 不再静默回退到 `opts[0]` 或进入空 `PageLoading`

### 2. 当前用户无简历空状态
- `resumeOptions.length === 0` → 独立空状态页
- 文案："你还没有可用于推荐的简历，请先创建或导入一份简历"
- CTA："去创建简历"（/resumes/new）+ "管理我的简历"（/resumes）

### 3. URL 无 resumeId 自动选择
- 无 resumeId 时自动选择 `is_current` 或第一个简历
- 通过 `navigate(replace)` 更新 URL 为 `/recommend?resumeId=X`

### 4. 错误文案分层
- URL resumeId 不匹配 → "简历不存在或不属于当前账号"
- API 失败 → "推荐结果加载失败，请稍后重试"（含额外指引）
- Deep 模型失败 → "AI 模型连接失败"（已有 AgentErrorState）

## 验证

### 场景
- Arthur (user 1, 0 resumes) + `/recommend?resumeId=1` → 前端显示 "简历不存在或不属于当前账号"（不再请求推荐 API）
- Demo (user 2, 8 resumes) + `/recommend?resumeId=1` → Fast 推荐正常返回 `fast_tag_engine`
- Demo + `/recommend?resumeId=8` → Fast + Deep 正常
- Demo + `/recommend` → 自动选择当前简历

### 构建
- npm build ✅
- go test ./server/... ✅
- make check ✅
- make prod-check ✅

## 未修改
- 后端源码
- 数据库
- 配置
- mock/fallback 设置
