# K9-UI-ResumeModal-E-BadgeClean — 删除简历选择弹窗"历史版本/当前使用"视觉标签

日期：2026-06-08
性质：前端 UI 清理。1 文件修改。

## 背景

`ResumePickerModal` 简历卡片中仍显示两个视觉标签：
- "当前使用"（蓝色 badge，`isCurrent===true` 时渲染）
- "历史版本"（灰色 badge，`statusLabel` 不为空且 `isCurrent===false` 时渲染）

用户认为这两个标签在弹窗中造成信息干扰，决定删除。

## 变更内容

### 删除 badge 渲染代码

文件：`web/src/app/components/ResumePickerModal.tsx`

1. **删除"当前使用" badge**（原 L358-L362）：
   ```tsx
   // 已删除
   {resume.isCurrent && (
     <span><CheckCircle2/> 当前使用</span>
   )}
   ```

2. **删除"历史版本" badge**（原 L363-L367）：
   ```tsx
   // 已删除
   {resume.statusLabel && !resume.isCurrent && (
     <span>{resume.statusLabel}</span>
   )}
   ```

3. **移除未使用的 `CheckCircle2` 导入**（仅 badge 使用该图标）。

### 保留内容

- `ResumePickerOption` 接口中 `isCurrent` 和 `statusLabel` 字段保持不变
- `SelectResumeModal.toPickerOption()` 映射逻辑不变 — 仍传递 `isCurrent` 和 `statusLabel`
- 默认选中逻辑不变 — `opts.find(o => o.isCurrent) || opts[0]`
- `resumeAdapter.ts` 中 `statusLabel: isCurrent ? '' : '历史版本'` 保持不变

## 验收

### 源码验证

```bash
grep -Rni "当前使用\|历史版本" web/src/app/components/ResumePickerModal.tsx web/src/app/components/SelectResumeModal.tsx
# Output: CLEAN: No badge text found ✅
```

### Badge 删除结果

| 文案 | 结果 |
|------|------|
| "当前使用" badge | ❌ 已删除 ✅ |
| "历史版本" badge | ❌ 已删除 ✅ |

### 默认选中逻辑验收

| 项 | 结果 |
|------|------|
| `isCurrent` 字段保留 | ✅ |
| `statusLabel` 字段保留 | ✅ |
| 默认选中逻辑（`find o.isCurrent`） | ✅ 不变 |
| radio 选中态 | ✅ 不变 |
| 蓝色边框 `border-[#0052D9]` | ✅ 不变 |
| 浅蓝背景 `bg-blue-50/50` | ✅ 不变 |

### 搜索 + 选择能力

| 项 | 结果 |
|------|------|
| 搜索功能 | ✅ 不变 |
| single 选择 | ✅ 不变 |
| compare 选择（最大 2 个）| ✅ 不变 |

### Portal / 交互能力

| 项 | 结果 |
|------|------|
| Portal (`createPortal`) | ✅ 不变 |
| 毛玻璃 (`backdrop-blur-sm`) | ✅ 不变 |
| body scroll lock | ✅ 不变 |
| ESC 关闭 | ✅ 不变 |
| 遮罩关闭 (`onMouseDown`) | ✅ 不变 |

## 构建 / 测试

| 验收项 | 结果 |
|--------|------|
| npm build | ✅ Passed (600.96 KB, -0.37 KB) |
| go build ./... | ✅ Passed |
| go vet ./... | ✅ Passed |
| go test ./... | 6 FAIL — **pre-existing**（agent/service K9-2 测试，与本次修改无关） |
| check-encoding | FAIL — **pre-existing**（init.generated.current.sql mojibake） |

### go test 失败归因

| 失败测试 | 归因 |
|----------|------|
| `TestMatchAgent_FinalV2Snapshot_EvidenceChain` | pre-existing, K9-2 agent 测试 |
| `TestMatchAgent_FinalV2Snapshot_TargetRoleEmptySmoke` | pre-existing, K9-2 agent 测试 |
| `TestIntegration_MatchAnalysis_Mock_Direct` | pre-existing, K9-2 集成测试 |
| `TestIntegration_MatchAnalysis_Mock_Fallback_Metadata` | pre-existing, K9-2 集成测试 |
| `TestMatchAnalysisAgentReturnsDecisionFieldsSmoke` | pre-existing, K9-2 agent 测试 |
| `TestMatchService_MockAgent_StableResult` | pre-existing, K9-2 service 测试 |

**全部 6 个 FAIL 均为 pre-existing，与本次修改无关，不阻断。**

### check-encoding 失败归因

`server/sql/init.generated.current.sql` 中包含历史 dump 中文乱码数据 — **pre-existing，与本次修改无关，不阻断。**

## 未修改范围确认

- 未修改 K9-1 岗位推荐业务逻辑
- 未修改 K9-2 岗位匹配分析后端
- 未修改 K9-3 简历优化建议
- 未修改 K9-4 简历版本对比
- 未修改 server/sql/init.sql
- 未修改数据库结构
- 未删除真实简历
- 未执行 reset-up
- 未执行 Git 命令
