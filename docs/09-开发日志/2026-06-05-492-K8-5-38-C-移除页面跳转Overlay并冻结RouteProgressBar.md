# K8-5-38-C：移除页面跳转 Overlay 并冻结 RouteProgressBar 跳转反馈

> 日期：2026-06-05 | 性质：前端修改 | 最小修改（2 文件，删 2 行 + 注释）

---

## 一、本阶段目标

执行 K8-5-38-B 设计冻结方案：移除 PageTransitionOverlay 挂载，保留 RouteProgressBar 作为普通页面跳转唯一全局反馈。

---

## 二、修改文件

| 文件 | 操作 | 说明 |
|------|:---:|------|
| `web/src/app/routes.tsx` | 修改 | 移除 `PageTransitionOverlay` 导入行和渲染行；恢复 `<main>` 替换 `div.flex-1.relative` |
| `web/src/app/components/common/PageTransitionOverlay.tsx` | 修改 | 顶部加 `@deprecated` 注释块；组件代码不变 |

---

## 三、routes.tsx 改动细节

**删除**：`import { PageTransitionOverlay } from "./components/common/PageTransitionOverlay";`

**删除**：`<PageTransitionOverlay />` 渲染

**恢复**：`<div className="flex-1 relative">` → `<main className="flex-1">`（恢复语义 HTML）

**保留**：`<RouteProgressBar />`

---

## 四、PageTransitionOverlay.tsx deprecated 注释

```ts
/**
 * @deprecated K8-5-38-B 设计冻结后，普通页面跳转不再使用中央 Overlay。
 * 当前页面上方 overlay 已从 routes.tsx RootLayout 中移除。
 * 普通页面跳转统一使用 RouteProgressBar（顶部 2px 进度条）。
 * 本文件保留用于历史追溯，当前不启用。
 */
```

---

## 五、验收 grep 结果

| 检查项 | grep | 结果 |
|--------|------|:--:|
| routes 不再挂载 Overlay | `grep PageTransitionOverlay routes.tsx` | ✅ NO_ROUTES_REF |
| 文件保留 + deprecated | `grep @deprecated PageTransitionOverlay.tsx` | ✅ 命中 line 2 |
| RouteProgressBar 仍在 | `grep RouteProgressBar routes.tsx` | ✅ import + render 均保留 |
| 不新增依赖 | `grep nprogress/framer/...` | ✅ 0 命中 |

---

## 六、浏览器验证结果

| 场景 | 预期 | 结果 |
|------|------|:--:|
| 页面切换不再有中央胶囊 | 无"正在切换页面"白色胶囊 | 待用户确认 |
| 页面切换不再有遮罩 | 无半透 overlay | 待用户确认 |
| RouteProgressBar 仍出现 | 顶部 2px 蓝条短暂出现 | 待用户确认 |
| 页面 skeleton 正常 | 页面自身骨架屏保留 | 待用户确认 |
| 不再有三层视觉噪声 | 只有顶部蓝条 + 页面骨架 | 待用户确认 |

**浏览器手工验证待用户确认；代码路径、构建和 API 已通过。**

---

## 七、build / test / API 结果

| 验收项 | 结果 |
|--------|:---:|
| `npm run build` | ✅ 通过（3.45s，JS -1KB） |
| `go test ./...` | ✅ 全部通过 |
| `go build ./...` | ✅ 通过 |
| `go vet ./...` | ✅ 通过 |
| API `/api/jobs` | ✅ 200 |
| API `/api/jobs/10001` | ✅ 200 |
| API `/api/jobs/facets` | ✅ 200 |
| API `/api/resumes` | ✅ 401 |

---

## 八、未修改范围确认

```
未修改业务页面
未修改后端 Go 源码
未修改数据库 SQL
未修改 seed 数据
未修改 Agent / Prompt / Schema / Mock
未修改 API 路径
未修改登录鉴权
未修改 jobs 业务接口
未修改简历业务接口
未修改 .env
未安装依赖
未迁移目录
未新增上传能力
未执行 make reset-up
未写入数据库
```

---

## 九、docs 更新清单

- `docs/09-开发日志/2026-06-05-492-K8-5-38-C-移除页面跳转Overlay并冻结RouteProgressBar.md`（新增）
- `docs/下一步任务.md`（更新，K8-5-38 全部完成）
- `docs/变更记录.md`（更新）

---

## 十、Git 约束确认

```
本次未执行任何 Git 命令，也未提交代码。
```
