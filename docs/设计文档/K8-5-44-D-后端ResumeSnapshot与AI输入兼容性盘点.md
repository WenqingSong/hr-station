# K8-5-44-D：后端 Resume Snapshot / AI 输入兼容性盘点

> 状态：✅ 已完成 | 日期：2026-06-06 | 性质：只读审计

---

## 一、核心结论

**后端零改动可行。** Snapshot 正常器对旧字段缺失是优雅降级的（`if len > 0` 守卫），不会崩溃。AI 质量不会因新契约而退化——因为被删除的细分字段内容已合并进 `description`/`additional_info`，AI 看到的上下文字段不减反增。

---

## 二、四核心功能链路

| 功能 | API | 输入方式 | 是否依赖旧字段 | 结论 |
|------|-----|---------|:---:|------|
| 岗位推荐 | `POST /api/jobs/recommend` | `{resume_id, limit}` → 后端加载 Resume → Snapshot → LLM Prompt | ❌ 不依赖 | 加载的是 JSON 列，Snapshot 只提取不校验 |
| 岗位匹配 | `POST /api/jobs/{id}/match` | `{resume_id}` → 后端加载 → Snapshot → LLM | ❌ 不依赖 | 同上 |
| 简历优化 | `POST /api/resumes/optimize` | `{resume_id, job_id}` → Snapshot → LLM | ❌ 不依赖 | 同上 |
| 版本对比 | `POST /api/resumes/compare` | `{from_id, to_id}` → 两份 Snapshot → LLM | ❌ 不依赖 | 同上 |

---

## 三、Resume Snapshot 兼容性

| 项 | 结果 |
|------|:---:|
| Snapshot 是否解析 JSON 内部字段？ | ✅ 是（agent_snapshot.go + resume_snapshot_normalizer.go） |
| 依赖旧 `tech_stack`？ | ⚠️ 提取但非必需（`if len > 0` 守卫 + JSON omitempty） |
| 依赖旧 `achievements`？ | ⚠️ 提取但非必需（同上） |
| 依赖 `self_evaluation`？ | ⚠️ 提取但非必需（同样守卫，fallback 到 DB 列） |
| 旧字段缺失时是否崩溃？ | ❌ 不崩溃（空数组/空字符串，JSON 序列化 omitempty） |
| 新 `description` 是否被提取？ | ✅ 是（已有 `Description` 字段） |
| 新 `additional_info` 是否被提取？ | ✅ 是（已有 `AdditionalInfo` 字段） |
| `SelfEvaluation` 为空，AI 失去上下文？ | ❌ 不会——`AdditionalInfo` 仍携带内容 |

**Snapshot 提取逻辑（核心）：**
```go
// 旧字段：优雅降级
if ts := toStrSlice(exp["tech_stack"]); len(ts) > 0 { intern.TechStack = ts }
if ach := toStrSlice(exp["achievements"]); len(ach) > 0 { intern.Achievements = ach }
// 新字段已存在：description
intern.WorkContent = s(exp, "description")  // V1
intern.Description = s(exp, "description")   // V2
// SelfEvaluation fallback
SelfEvaluation: toString(extra["self_evaluation"])  // 新 seed 中为空
if SelfEvaluation == "" { SelfEvaluation = v.SelfEvaluation }  // DB 列也为空
// AdditionalInfo 不受影响
AdditionalInfo: toString(extra["additional_info"])  // 新 seed 有值
```

---

## 四、旧字段依赖搜索结果

| 旧字段 | 命中位置 | 风险 | 建议 |
|--------|---------|:---:|------|
| `tech_stack` | snapshot normalizer × 4 | 🟢 无风险（if > 0 守卫） | 无需修改 |
| `achievements` | snapshot normalizer × 4 | 🟢 无风险（if > 0 守卫） | 无需修改 |
| `self_evaluation` | snapshot normalizer + DB 列 | 🟢 无风险（fallback 也空） | 无需修改 |
| `related_courses` | 无 Go 代码引用 | 🟢 | — |
| `project_url` | snapshot normalizer（V1 link） | 🟢 无风险（新 seed 无此字段） | — |
| `quant_result` | 无 Go 代码引用 | 🟢 | — |
| `certificate.type` | agent_snapshot_test.go（测试数据） | 🟢 测试受影响 | K8-5-44-E 验收 |
| `award.type` | 同上 | 🟢 | K8-5-44-E |

---

## 五、Agent / Prompt / Schema 兼容性

| 模块 | 是否需要修改 | 原因 |
|------|:---:|------|
| Snapshot Normalizer | ❌ 不需要 | 优雅降级，if 守卫 + omitempty |
| Mock Agent | ❌ 不需要 | 测试数据中旧字段仅用于测试 fixture，不影响生产 |
| Prompt (`version_compare_v1.md`) | ⚠️ 建议更新 | 文档中引用 `tech_stack`/`achievements` 作为 resume 格式说明，非硬依赖 |
| Prompt (其他) | ❌ 不需要 | references 都是通用描述，不绑定具体字段 |
| Schema | ❌ 不需要 | 只约束输出结构，不约束输入 |
| Repository (self_evaluation 列) | ❌ 不需要 | 列仍然存在，只是不再有新数据写入 |

---

## 六、是否需要后端修改

| 层级 | 结论 |
|------|------|
| **Snapshot Normalizer** | ❌ 不需要——`if len > 0` + `omitempty` 保障兜底 |
| **Model** | ❌ 不需要 |
| **Repository** | ❌ 不需要 |
| **Service** | ❌ 不需要 |
| **Agent** | ❌ 不需要（Mock 不影响生产） |
| **Prompt** | ⚠️ 建议更新（非阻塞，next prompt 可做） |
| **Schema** | ❌ 不需要 |

---

## 七、后续阶段建议

```
K8-5-44-E：make reset-up + 9 份示范简历全字段回归验收
  — 验证新 seed 导入正确
  — 验证 Snapshot 对 9 份简历生成正确
  — 验证四核心 AI 功能可用
  — 可选：更新 version_compare_v1.md prompt 中的 resume 格式描述
```

---

## 八、构建 / 测试结果

| 验收项 | 结果 |
|--------|:---:|
| `npm run build` | ✅ |
| `go build ./...` | ✅ |

---

## 九、未修改范围确认

```
✅ 未修改源码（Go / TS / SQL）
✅ 未修改 seed
✅ 未修改数据库
✅ 未执行 make reset-up
✅ 未执行 Git 命令
```
