# Recommendation-Fast-First-Lifecycle-And-CTA-Fix-H9：Fast-first 生命周期与 CTA 层级修复

## 目标

将推荐页从 quality-first（有 Deep 就展示 Deep）改为 Fast-first（默认展示 Fast，Deep 手动触发）。重组 CTA 层级避免用户混淆。

---

## 修复详情

### 1. Fast-first 初始化

| 项目 | 修复前 | 修复后 |
|------|------|------|
| 首请求 | `loadLatest(rid, "quality")` | `loadLatest(rid, "fast")` |
| quality 加载 | 同步阻塞 | 异步后台（不阻塞 Fast 展示） |
| 有 quality 历史时 | 默认展示 Deep | Fast 展示 + "查看深度推荐"入口 |

### 2. displayMode 状态

新增 `const [displayMode, setDisplayMode] = useState<"fast"|"deep">("fast")`。

| 事件 | displayMode 变化 |
|------|:---:|
| 页面初始化 | "fast" |
| generateFast 开始 | "fast" |
| Deep 成功后 | "deep" |
| "返回快速推荐"按钮 | "fast" |
| "查看深度推荐"按钮 | "deep" |

### 3. activeResult 由 displayMode 驱动

```ts
const activeResult = displayMode === "deep" && deepResult ? deepResult : fastResult;
```

不再使用 data-driven 的 `deepReady ? deepResult : fastResult`。

### 4. CTA 层级

| 模式 | 按钮 | 作用 |
|------|------|------|
| Fast (默认) | **升级为深度推荐**（蓝底主 CTA） | 触发 Deep Cascade |
| Fast | 重新快速筛选（灰字弱入口） | 重新生成 Fast |
| Deep | 重新生成深度推荐 | 触发 Deep Cascade |
| Deep | 返回快速推荐 | 切换回 Fast |

旧文案"生成深度推荐"改为"升级为深度推荐"，语义明确表示从 Fast 升级到 Deep。

### 5. Deep 历史可见性

当 displayMode="fast" 但 deepResult 有完整 5 条时，右侧控制台显示：
> 深度推荐历史可用 · 已有 5 个岗位的深度分析结果 · **查看深度推荐**

---

## 修改文件

| 文件 | 修改内容 |
|------|------|
| `web/src/app/pages/JobRecommendationPage.tsx` | Fast-first init + displayMode + CTA 重组 + Deep 历史入口 |

---

## 构建

| 项目 | 结果 |
|------|:---:|
| npm build | ✅ PASS (2.29s) |
| Fast API R4/1/8 | ✅ 0 mock, fast_tag_engine, 5 items |

---

## 浏览器验证

| 场景 | 预期 |
|------|------|
| /recommend?resumeId=4 首屏 | 自动 Fast，主 CTA "升级为深度推荐"，弱入口"重新快速筛选" |
| 有 Deep 历史 | Fast 展示 + "查看深度推荐"入口 |
| 点击"升级为深度推荐" | Deep loading → Deep 展示 |
| Deep 失败 | 保留 Fast，显示错误 + "返回快速推荐" |
| 点击"返回快速推荐" | 切换回 Fast |
