# K8-5-18-C — Input 改造 + PasswordInput 新增 + Auth 接入

> 日期：2026-06-03 | 性质：implementation

---

## 1. 本阶段目标

- 改造 `ui/input.tsx` — Offerr鹅 统一风格
- 新建 `ui/password-input.tsx`
- `theme.css` 隐藏浏览器默认 password reveal
- AuthDialog / LoginPage / RegisterPage 接入统一控件
- 修复密码双眼睛 P0

---

## 2. 修改方案

| 区域 | 修改内容 | 原因 |
|------|---------|------|
| `ui/input.tsx` | 替换默认 className 为 h-11 rounded-xl border-slate-200 + focus:ring-4；新增 error prop | 统一 Input 规范 |
| `ui/password-input.tsx` | 新建：Eye/EyeOff 切换 + `pr-12` + 隐藏 reveal CSS 配合 | 统一密码输入框 |
| `theme.css` | 新增 `::-ms-reveal` / `::-ms-clear` / `::-webkit-credentials-auto-fill-button` | 隐藏浏览器密码揭示按钮 |
| `AuthDialog.tsx` | 5 个原生 input → Input + PasswordInput ×3 | 接入统一控件，消除双眼睛 |
| `LoginPage.tsx` | 1+1 原生 input → Input + PasswordInput | 接入统一控件 |
| `RegisterPage.tsx` | 1+2 原生 input → Input + PasswordInput ×2 | 接入统一控件 |

---

## 3. 修改统计

- Input：2 个（AuthDialog phone ×1 + reg phone ×1 + LoginPage ×1 + RegisterPage ×1 = 4）
- PasswordInput：6 个（AuthDialog ×3 + LoginPage ×1 + RegisterPage ×2）
- 删除 6 个 showPassword/showConfirm/regShowPassword/regShowConfirm state
- 删除 3 处 Eye/EyeOff import

---

## 4. 密码双眼睛修复

| 修复项 | 实现 |
|--------|------|
| 浏览器默认 reveal | `::-ms-reveal` / `::-ms-clear` `display:none` in theme.css |
| 项目自定义 Eye | 统一在 PasswordInput 内部，只保留一个按钮 |
| AuthDialog | 0 处 Eye 手写 ✅ |
| LoginPage | 0 处 Eye 手写 ✅ |
| RegisterPage | 0 处 Eye 手写 ✅ |

---

## 5. 业务逻辑保护

| 能力 | 结果 |
|------|:---:|
| `login()` / `register()` | ✅ |
| `setAuthToken()` | ✅ |
| `onLoginSuccess` / `onOpenChange` / `defaultMode` | ✅ |
| `getSafeRedirect` / `navigate` | ✅ |
| mode 切换 / mobile 保留 | ✅ |

---

## 6. 构建 / API 结果

| 验收项 | 结果 |
|--------|------|
| `npm run build` | ✅ 通过 (2.47s) |
| 4 API 检查 | ✅ 3×200 + 1×401 |

---

## 7. 未修改范围

```text
未修改后端/数据库/接口/.env/LLM
未修改 routes/RequireAuth/Navbar/client.ts
未修改 MyResumesPage/ResumeComparisonPage/Select/Textarea/AI 页面
未安装依赖/未迁移目录
```

---

## 8. Git 约束确认

```text
本次未执行任何 Git 命令，也未提交代码。
```
