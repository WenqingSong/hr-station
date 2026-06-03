# K8-5-13-D：Navbar 接入 AuthDialog，保留 /login /register fallback

## 1. 本阶段目标
K8-5-13-C 已新增 AuthDialog。本次将 Navbar 的"登录"入口从页面跳转改为弹窗，同时保留 fallback 路由。

## 2. 接入方案

| 项目 | 方案 |
|------|------|
| 登录入口 | `Link to="/login"` → `button onClick={setAuthDialogOpen(true)}` |
| 弹窗 state | Navbar 内部 `useState(false)` |
| 弹窗渲染 | `<AuthDialog>` 在 Navbar 返回的 Fragment 中 |
| 登录态刷新 | `handleAuthSuccess` → `setIsLoggedIn(Boolean(getAuthToken()))` |
| /login /register | **保留不变** |

## 3. 修改清单

| 文件 | 修改 | 属于 K8-5-13-D |
|------|------|:--:|
| `Navbar.tsx` | 导入 AuthDialog + state + handleAuthSuccess + 按钮替换 + 弹窗渲染 | 是 |
| `docs/09-开发日志/2026-06-03-369-...md` | 新建 | 是 |

## 4. 验收

| 验收项 | 结果 |
|--------|------|
| Navbar 引入 AuthDialog | **pass**（7 处匹配） |
| 登录 Link 移除 | **pass**（无匹配） |
| fallback 路由保留 | **pass**（routes.tsx 含 login+register；LoginPage/RegisterPage 文件完整） |
| `npm run build` | **pass**（1.62s） |
| API | **pass** |

## 5. 未修改范围
未改 /login /register 页面和路由 / routes / App / 后端 / 数据库 / 接口 / .env / ui/dialog / SelectResumeModal / AuthContext，未安装依赖。

## 6. Git 约束确认
本次未执行任何 Git 命令，也未提交代码。
