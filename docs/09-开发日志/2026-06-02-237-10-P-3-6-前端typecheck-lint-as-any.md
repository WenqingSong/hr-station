# 开发日志 #237：10-P-3-6 前端 typecheck / lint 与 as any 收敛

> 日期：2026-06-02 | 阶段：10-P-3-6 | 状态：已完成

## 1. 本次任务

补齐前端 typecheck script，收敛 as any 类型风险，修复类型错误。

## 2. 发现

- TypeScript 依赖缺失（Vite 直接编译，无 tsc）
- 8 处 `null as any` 在 Agent 相关组件 adapter 调用中
- 1 处 `icon: any` 在 ResumeOptimizationDetail
- 3 个 UI 组件大小写冲突文件（badge/button/checkbox 各有两个）
- 无 ESLint 配置/依赖
- `import.meta.env` 缺少 Vite 类型声明

## 3. 修改

| 类型 | 修改 |
|---|---|
| 配置 | `package.json` 新增 typecheck + typescript/@types/react-dom |
| 配置 | `tsconfig.json` 新增 ignoreDeprecations，exclude 死代码 |
| 配置 | 新增 `vite-env.d.ts` |
| 适配器 | `toJobDetailViewModel`/`toResumeSummaryViewModel` 改为 nullable |
| 组件 | 移除 8 处 `null as any` |
| 组件 | `icon: any` → `React.ComponentType<...>` |
| 类型 | `HttpMethod` 增加 `'HEAD'` |
| 清理 | 删除 3 个大小写冲突文件，修复 5 处 import |

## 4. 验证

| 项 | 结果 |
|---|---|
| npm run typecheck | ✅ 零错误 |
| npm run build | ✅ 2.27s |
| go build | ✅ PASS |
| as any 残留 | 仅 MyResumesPage 2 处（P2 死代码） |
| mock 反弹 | 零匹配 |
