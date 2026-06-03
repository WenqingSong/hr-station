# K8-5-13-C：新增 AuthDialog，不接 Navbar，不改路由

## 1. 本阶段目标
K8-5-13-B 已冻结设计。本次新增 `auth/AuthDialog.tsx` 组件，仅 build 验证通过，不接入任何页面。

## 2. 新增组件

| 组件 | 路径 | 用途 | 接入 |
|------|------|------|------|
| AuthDialog | `web/src/app/components/auth/AuthDialog.tsx` | 登录/注册弹窗 | ❌ 未接 |

## 3. Props 设计

```ts
open: boolean          // 是否打开
onOpenChange: (open)   // 关闭回调
defaultMode?: "login"  // 默认 login
onLoginSuccess?: ()    // 成功回调
```

## 4. 登录 / 注册逻辑

| 模式 | 字段 | pending | success | error |
|------|------|---------|---------|-------|
| Login | 手机号+密码 | Loader2+"登录中…" disabled | setAuthToken+关闭 | bg-red-50 |
| Register | 手机号+密码+确认密码 | Loader2+"注册中…" disabled | setAuthToken+关闭 | bg-red-50 |

## 5. 登录/注册切换
`switchMode(to)`：保留手机号，清空密码/confirmPassword/error。

## 6. 玻璃模糊视觉
- Overlay：通过 Dialog 内置 overlay + `DialogContent` 定制
- Content：`rounded-3xl bg-white/85 backdrop-blur-xl ring-1 ring-white/60`
- 不修改 `ui/dialog.tsx`

## 7. 为什么不接 Navbar
等待 K8-5-13-D。当前 AuthDialog 仅组件存在，不影响任何页面。

## 8. build 结果
✅ **pass**（2.05s）

## 9. API 结果
✅ `GET /api/jobs?page=1` — pass
✅ `GET /api/jobs/10001` — pass
✅ `GET /api/jobs/facets` — pass

## 10. 验收
| 验收项 | 结果 |
|--------|------|
| build | pass |
| AuthDialog.tsx 存在 | pass |
| 28 处关键内容匹配 | pass |
| Navbar/routes/App/Login/Register 无引用 | pass |
| 无 alert/navigate(调用) | pass |

## 11. 未修改范围
未修改 Navbar/routes/App/LoginPage/RegisterPage/后端/数据库/接口/ui-dialog/SelectResumeModal/.env/LLM，未新增 AuthContext，未安装依赖。

## 12. Git 约束确认
本次未执行任何 Git 命令，也未提交代码。
