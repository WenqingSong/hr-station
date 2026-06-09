# K9-UI-AgentProgress-A：四大深度生成进度弹窗时序只读诊断

> 日期：2026-06-09 | 性质：只读诊断，未修改任何源码

---

## 一、当前状态确认

| 项 | 结果 |
| -- | -- |
| 进度组件 | `AgentStepLoadingOverlay.tsx`（175 行） |
| timer 间隔 | `STEP_INTERVALS = [800, 1000, 1200]` ms |
| 成功停留 | `successDelayMs = 400` ms |
| 用到的页面 | 推荐 / 匹配 / 优化 / 对比（4 个） |
| 成功过渡 | **推荐✅ 对比✅ / 匹配❌ 优化❌**（分裂） |

---

## 二、文件定位

| 功能 | 文件 | 使用组件 | steps |
| -- | -- | -- | -- |
| 岗位推荐深度生成 | `JobRecommendationPage.tsx` | AgentStepLoadingOverlay（含 success） | 读取简历画像→匹配岗位标签→推荐理由→审查一致性 |
| 岗位匹配深度分析 | `JobMatchAnalysisPage.tsx` | AgentStepLoadingOverlay（仅 loading） | 读取画像→拆解岗位要求→匹配证据→差距建议 |
| 简历优化深度报告 | `ResumeOptimizationPage.tsx` | AgentStepLoadingOverlay（仅 loading） | 读JD关键要求→对照简历→生成建议→汇总策略 |
| 简历版本深度对比 | `ResumeComparisonPage.tsx` | AgentStepLoadingOverlay（含 success） | 读两份简历→分析差异→生成合并→汇总建议 |

---

## 三、AgentStepLoadingOverlay 实现诊断

| 检查项 | 当前实现 | 问题 | 建议 |
| -- | -- | -- | -- |
| 步骤推进 | `setTimeout` 自动推进，间隔 800/1000/1200ms | ✅ | — |
| 成功时行为 | `setActiveIdx(steps.length)` 所有步骤跳过→done | **跳过中间步骤**，API 快返回时看不到第 2-3 步 | 成功时不直接跳到 done，加速推进剩余步骤 |
| minDuration | ❌ 无 | 快接口时 overlay 可见时间 < 1s | 新增 `minVisibleMs = 1800` |
| onComplete | ✅ `onClose` 回调 | ✅ | — |
| 状态机 | loading → success → close | ✅ 结构正确 | — |
| 父组件关闭 | 匹配/优化：`open = phase === "deep_loading"` → setPhase("deep_done") 后 **立即 false** | **根本原因**：无 success 过渡态 | 增加 `deepOverlayStatus: "success"` |

---

## 四、四个深度生成关闭时机诊断

| 功能 | loading phase | API resolve 后动作 | overlay 关闭条件 | 问题 |
| -- | -- | -- | -- | -- |
| 岗位推荐 | `phase === "deep_loading" \|\| deepOverlayStatus === "success"` | `setDeepOverlayStatus("success")` → 400ms → `onClose` → `setPhase("deep_done")` | `onClose` 回调触发 | ✅ 正常（但快返回时跳过步骤） |
| 岗位匹配 | `phase === "deep_loading"` | `setPhase("deep_done")` **立即** | `open` 立即变 false → **unmount** | ❌ **API 返回后秒关，步骤中断** |
| 简历优化 | `phase === "deep_loading"` | `setPhase("deep_done")` **立即** | `open` 立即变 false → **unmount** | ❌ **同上** |
| 简历版本对比 | `phase === "deep_loading" \|\| deepOverlayStatus === "success"` | `setDeepOverlayStatus("success")` → 400ms → `onClose` → `setPhase("deep_done")` | `onClose` 回调触发 | ✅ 正常（同推荐） |

---

## 五、根因分析

| 根因 | 证据 | 影响 |
| -- | -- | -- |
| **匹配/优化无 success 过渡** | `open={phase === "deep_loading"}`，无 `deepOverlayStatus` | API 返回 → `setPhase("deep_done")` → overlay `open=false` → 立即 unmount → 步骤中断 |
| **成功时直接跳到 done** | `setActiveIdx(steps.length)` 跳过中间步骤 | 快接口时用户只看到第 1-2 步 |
| 四个页面时序不统一 | 推荐/对比有 success 过渡，匹配/优化没有 | 体验不一致 |

### 完整因果链（匹配/优化—"第二步就关闭"）

```
点击"生成深度分析"
→ phase = "deep_loading"
→ overlay 显示（status: "loading"）
→ timer 推进 step 0→1→2→3
→ API ~900ms 返回
→ setPhase("deep_done")
→ overlay open 变为 false
→ React unmount overlay
→ 用户仅看到 step 0~1（约 800+1000=1800ms 内已执行到 step 2 的 setTimeout，
   但 API 返回时 step 2 尚未渲染到 DOM → 看起来"只走了两步"）
```

---

## 六、产品化时序设计建议

### 推荐状态机

```
idle → running → backend_done → finishing_steps → success_hold → closed
                                 ↘ error (保留原结果)
```

| 状态 | 行为 |
| -- | -- |
| running | steps 按 800/1000/1200ms 推进，最多到 steps.length-2 |
| backend_done | 不跳到最后一步，改为**加速推进**剩余步骤（每步 250ms） |
| finishing_steps | 所有步骤标记 done |
| success_hold | 停留 400ms（显示全部完成状态） |
| closed | onClose → setPhase("deep_done") |
| error | 保留 deep_error 状态，显示原结果 |

### 推荐时间参数

```
minVisibleMs: 1800  （最少可见时长，防止快接口秒关）
stepIntervalMs: [800, 1000, 1200]（当前，保持不变）
finishStepIntervalMs: 250  （后端完成后加速）
successHoldMs: 400  （当前，保持不变）
maxBeforeBackendDoneStep: steps.length - 2  （最多走到倒数第二步）
```

---

## 七、技术方案判断

| 方案 | 是否推荐 | 理由 |
| -- | -- | -- |
| A：每个页面加 setTimeout delay | ❌ | 四个页面重复改，治标不治本 |
| **B：增强 AgentStepLoadingOverlay** | ✅ **推荐** | 新增 `isBackendDone` / `minVisibleMs` props，内部管理完成时序 |
| C：新增 Hook + 新 Overlay | ⚠️ | 范围大，先验证方案 B |

---

## 八、四个功能 steps 文案整理

| 功能 | 当前文案 | 风格 |
| -- | -- | -- |
| 岗位推荐 | 读取简历画像标签 / 匹配岗位标签矩阵 / 生成推荐理由和风险点 / 审查推荐结果一致性 | ✅ 动词+对象 |
| 岗位匹配 | 读取简历画像标签 / 拆解岗位要求 / 计算匹配证据 / 生成差距与补强建议 | ✅ |
| 简历优化 | 读取目标岗位 JD 与关键要求 / 对照当前简历中的项目、技能和经历证据 / 生成可执行的修改建议与改写示例 / 汇总投递重点和后续补强策略 | ✅ |
| 简历版本对比 | 读取两份简历版本信息 / 分析优势、风险和关键差异 / 生成可合并内容与改写示例 / 汇总下一版简历迭代建议 | ✅ |

文案风格一致，无需修改。

---

## 九、下一阶段建议

### K9-UI-AgentProgress-B：四大深度生成进度补全动画重构

**执行范围**：

1. **AgentStepLoadingOverlay 改造**：
   - 新增 `isBackendDone: boolean` prop
   - 新增 `minVisibleMs: number` prop（默认 1800）
   - backendDone 后不跳到最后，改为加速推进剩余步骤（250ms/步）
   - 保持 timer cleanup 不变

2. **匹配页修复**（P0）：
   - 新增 `matchDeepOverlayStatus` state
   - `handleDeepAnalysis` 中 `.then()` 后不立即 `setPhase("deep_done")`，改为 `setMatchDeepOverlayStatus("success")`
   - overlay `open` 条件改为 `phase === "deep_loading" || matchDeepOverlayStatus === "success"`
   - overlay `onClose` 中 `setPhase("deep_done")`

3. **优化页修复**（P0）：
   - 同匹配页，新增 `optDeepOverlayStatus`

4. **推荐/对比页**：仅需改 overlay prop（`status` → `isBackendDone`）

5. **不改后端 / API / 数据库**

---

## 十、浏览器验收用例

| 用例 | 验收标准 |
| -- | -- |
| 匹配深度分析 | 四步逐步走完 + 成功停留 → 关闭 |
| 优化深度报告 | 同上 |
| 推荐深度生成 | 四步逐步走完 + 不再跳步 |
| 对比深度生成 | 同上 |
| API 快速返回 (<1s) | 不秒关，至少可见 1800ms |
| API 超时/失败 | deep_error 正常，保留原结果 |

---

## 十一、未修改范围确认

- 未修改任何源码 / 数据库 / init.sql
- 未 reset-up
- 未执行 Git

---

## 十二、docs 更新清单

| 文件 | 操作 |
| -- | -- |
| `docs/09-开发日志/2026-06-09-678-K9-UI-AgentProgress-A-...` | 新增 |
| `docs/下一步任务.md` | 更新 |
| `docs/变更记录.md` | 更新 |

---

## 十三、Git 约束确认

- [x] 未执行任何 Git 命令
