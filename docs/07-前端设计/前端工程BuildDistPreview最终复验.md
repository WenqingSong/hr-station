# 前端工程 Build / Dist / Preview 最终复验

> 阶段：5-9.1 | 状态：已完成 | 日期：2026-05-31

## 1. 本阶段目标

从工程交付角度复验 build、dist 产物、preview 可运行性。

## 2. 执行环境

| 项目 | 值 |
|------|-----|
| Node | v22.22.2 |
| npm | 10.9.7 |
| Vite | 6.3.5 |
| build 命令 | `npm run build` |
| preview 命令 | `npx vite preview --outDir dist --host 0.0.0.0 --port 4173 --strictPort` |

## 3. package / scripts 核验

| 项目 | 状态 |
|------|:--:|
| package.json | ✅ |
| package-lock.json | ✅ |
| node_modules | ✅ |
| vite binary | ✅ `node_modules/.bin/vite` |
| build script (`vite build`) | ✅ |
| preview script | ❌ 未定义（使用 `npx vite preview`） |

## 4. build 最终复验

```
✓ built in 1.51s
1671 modules transformed
dist/index.html          0.40 kB
dist/assets/index-.css   124.14 kB
dist/assets/index-.js    440.78 kB
```

| 检查项 | 结果 |
|------|:--:|
| build 成功 | ✅ |
| 用时 | 1.51s |
| TS/JSX/import 错误 | ❌ 无 |
| warning | ❌ 无 |

## 5. dist 产物核验

| 检查项 | 结果 |
|------|:--:|
| web/dist 存在 | ✅ |
| dist/index.html | ✅ (0.40 kB) |
| dist/assets/ | ✅ |
| JS asset | ✅ `index-CfCJzsBG.js` (440.78 kB) |
| CSS asset | ✅ `index-Xh7FkG0q.css` (124.14 kB) |
| index.html 引用 JS | ✅ `<script ... src="/assets/index-CfCJzsBG.js">` |
| index.html 引用 CSS | ✅ `<link ... href="/assets/index-Xh7FkG0q.css">` |
| dist 总大小 | ✅ 560K |

## 6. preview 启动结果

| 检查项 | 结果 |
|------|:--:|
| 启动成功 | ✅ |
| 本地地址 | `http://localhost:4173/` |
| 错误 | ❌ 无 |

## 7. 关键 URL 抽样复验

| URL | HTTP | 结论 |
|------|:--:|:--:|
| `/` | **200** | ✅ |
| `/recommend?resumeId=resume-v2` | **200** | ✅ |
| `/match?jobId=job-backend-dev&resumeId=resume-v2` | **200** | ✅ |
| `/resume-optimization?jobId=job-backend-dev&resumeId=resume-v2` | **200** | ✅ |
| `/resumes` | **200** | ✅ |
| `/resumes/edit` | **200** | ✅ |
| `/resumes/new` | **200** | ✅ |
| `/resumes/compare` | **200** | ✅ |

**8/8 主路径全部可达 ✅**

旧路径兼容：

| URL | HTTP |
|------|:--:|
| `/recommendations` | 200 → `/recommend` |
| `/match-analysis` | 200 → `/match` |
| `/my-resumes` | 200 → `/resumes` |

## 8. preview 停止结果

✅ 已停止，无残留进程。

## 9. 越界检查

| 检查 | 结果 |
|------|:--:|
| server/Makefile/docker/nginx/env 修改 | ❌ clean |
| axios | ❌ 无 |
| /api/v1 | ❌ 无 |
| selectors | ❌ 无 |
| localStorage/sessionStorage | ❌ 无 |
| web/src 修改 | ❌ 无 |

## 10. 本阶段未做事项

1. 未修改 web/src/
2. 未重新 npm install
3. 未启动 dev
4. 未接真实 API
5. 未调用 LLM

## 11. 是否可以进入 5-9.2

**可以。** build 通过，dist 完整（560K），preview 可运行，8 主路径 + 3 旧路径全部 HTTP 200。零越界。

**下一步：5-9.2 静态 Demo 三路径最终复核。**
