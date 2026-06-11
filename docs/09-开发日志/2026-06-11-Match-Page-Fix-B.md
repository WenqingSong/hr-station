# Match-Page-Fix-B：修复岗位匹配页 resumeId/jobId 无效状态与错误态

日期：2026-06-11

## 背景

`Match-Page-Audit-A` 定位：匹配页存在与推荐页相同的脆弱性——URL 中的 `resumeId` 不属于当前用户时静默回退到 `opts[0]`，`jobId` 缺失/无效时状态不够明确。

## 修复内容

### 1. resumeId 归属校验
- `urlResumeNotFound` 状态：URL `resumeId` 不在当前用户 `resumeOptions` 中时设为 true
- 页面显示："简历不存在或不属于当前账号" + "返回我的简历"
- URL 无 resumeId 时自动选择 `is_current` 或第一个简历并更新 URL

### 2. jobId 有效性校验
- `jobNotFound` 状态：岗位加载失败时设为 true
- 页面显示："岗位不存在或已下线" + "返回岗位列表"

### 3. 空状态优化
- 缺少 jobId → "缺少目标岗位" + 返回岗位列表 CTA
- 当前用户无简历 → "你还没有可用于匹配的简历" + 创建/管理简历 CTA

### 4. 错误文案分层
- 匹配页保持自身文案，不显示推荐页文案
- Fast API 失败 → "快速匹配分析生成失败"
- Deep 模型失败 → "AI 模型连接失败"

## 验证

| 场景 | 用户 | 结果 |
|---|---|---|
| Arthur 无权限 resumeId=1 | Arthur | 前端显示 "简历不存在或不属于当前账号" |
| Arthur 无简历 | Arthur | 显示 "你还没有可用于匹配的简历" |
| Demo resumeId=1 Fast | Demo | HTTP 200, `fast_tag_engine` |
| Demo resumeId=8 Deep | Demo | HTTP 200, `llm_agent` |
| 推荐页回归 | — | 所有 fix 组件完好 |

### 构建
- npm build ✅
- go test ./server/... ✅
- make check ✅
- make prod-check ✅

## 未修改
- 后端、配置、数据库、mock/fallback 设置
