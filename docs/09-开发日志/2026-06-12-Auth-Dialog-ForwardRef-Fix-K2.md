# Auth-Dialog-ForwardRef-Fix-K2：修复登录弹窗 Dialog forwardRef 警告

## 修复目标

基于 K1 审计结论，消除浏览器控制台 `Function components cannot be given refs` 警告。

## 根因

`web/src/app/components/ui/dialog.tsx` 中 4 个 Radix wrapper 组件（DialogOverlay / DialogContent / DialogTitle / DialogDescription）是普通函数组件，未使用 `React.forwardRef`。Radix UI 内部通过 `SlotClone` / `asChild` 机制传递 ref 时失败。

## 修复

| 组件 | 修复前 | 修复后 |
|------|------|------|
| DialogOverlay (line 33) | `function DialogOverlay(...)` | `React.forwardRef<...>` + `displayName` |
| DialogContent (line 49) | `function DialogContent(...)` | `React.forwardRef<...>` + `displayName` |
| DialogTitle (line 97) | `function DialogTitle(...)` | `React.forwardRef<...>` + `displayName` |
| DialogDescription (line 110) | `function DialogDescription(...)` | `React.forwardRef<...>` + `displayName` |

所有样式、className 合并逻辑、Radix Primitive 语义均保持不变。

## 未修改的文件

- AuthDialog.tsx / Navbar.tsx / authStore.tsx / RequireAuth.tsx / API client — 均未改动
- dialog.tsx 仅改组件签名（function → forwardRef），不改业务逻辑

## 验证

| 项目 | 结果 |
|------|:---:|
| npm build | ✅ PASS (1.93s) |
| forwardRef scan | ✅ 4 组件均含 forwardRef + displayName |
| auth 业务逻辑泄漏扫描 | ✅ dialog.tsx 不含任何 auth/login/register/token 引用 |

## 浏览器验证

环境无浏览器。**需用户本机刷新验证**：
1. 打开 `http://localhost:8080/`
2. 打开浏览器控制台
3. 点击右上角登录按钮 → 弹窗正常打开
4. 确认 console 不再出现 `Function components cannot be given refs` / `DialogOverlay` / `SlotClone`

## 修改文件

| 文件 | 修改内容 |
|------|------|
| `web/src/app/components/ui/dialog.tsx` | DialogOverlay/DialogContent/DialogTitle/DialogDescription → `React.forwardRef` + `displayName` |
