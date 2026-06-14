# 10-P-4-4 Resume Optimization Education Advice P0 Fix Acceptance

## 1. 背景

Offer 捕手项目中，简历优化功能存在不合理的学历建议问题：即使简历已有完整教育经历、目标岗位是普通后台开发岗（无高学历门槛），简历优化结果仍可能输出"学历缺失""未标注学历层次""学历不满足岗位要求"等误判性建议。

这类建议对普通开发岗位不适用，会造成用户困惑和学历焦虑。

## 2. 问题

- 普通后台开发、前端开发、测试开发、产品运营等岗位，如果 JD 没有明确更高学历门槛，不应把学历作为投递前必须修改项
- 简历已有教育经历时，不应输出"学历缺失"等误判
- JD 仅写"本科及以上"且简历已是本科，不应生成学历风险

## 3. Prompt 1 已完成修改

### Prompt 文件

| 文件 | 修改 |
|------|------|
| `server/prompts/resume_optimization_v1.md` | 新增 Section 11 · Education Advice Output Constraint（学 limith 制输出约束） |
| `server/prompts/optimization_decision_v1.md` | 新增「学历风险决策约束」小节 |
| `server/prompts/optimization_verifier_v1.md` | 新增「学历建议校验规则」小节 |

### 后端 Service 文件

| 文件 | 修改 |
|------|------|
| `server/internal/service/resume_optimization_service.go` | 新增 8 个过滤函数 + 6 处调用点，共 +197 行 |

## 4. 核心规则

### 4.1 不允许学历建议的条件

1. JD 无明确高学历门槛（硕士及以上、博士、985/211、双一流、科研经历等）
2. 简历已有教育经历
3. JD 仅写"本科及以上" + 简历已有本科
4. 岗位是普通后台开发、前端开发、测试开发、产品运营等

### 4.2 允许学历建议的条件

1. JD 明确要求硕士及以上、博士、985/211
2. 岗位是算法研究、AI Research、技术研究、实验室研究等高学历导向岗位
3. 简历确实不满足 JD 学历要求

## 5. 后端过滤覆盖范围

`filterEducationOptimizationAdvice` 覆盖以下所有字段：

```
RiskPoints
QuickFixItems
StructureFixItems
ContentFixItems
KeywordFixItems
EvidenceFixItems
RewriteExamples
LongTermImprovements
DeliveryFocus
```

## 6. 过滤调用路径（全部覆盖）

| 路径 | 函数 | 调用位置 |
|------|------|----------|
| Fast Tag Engine | `fastOptimizeWithTags` | 行 605，`saveOptV2Snapshot` 前 |
| Quality 3-Agent | `optimizeWith3Agents` | 行 643，`saveOptV2Snapshot` 前 |
| V1 LLM Fallback | `OptimizeV2` | 行 526，`saveOptV2Snapshot` 前 |
| Mock Fallback | `OptimizeV2` | 行 546，`saveOptV2Snapshot` 前 |
| Fast → V2 | `OptimizeV2` | 行 502，`return` 前 |
| Quality → V2 | `OptimizeV2` | 行 514，`return` 前 |

## 7. 验收结果

### 7.1 普通后台开发岗位（Test 1）

- **Job**：`软件开发-后台开发方向`（ID 10001），JD 无任何学历要求
- **Resume**：已有 "清华大学 计算机本科" 教育经历
- **Fast tier**：PASS — `risk_points` 和 `fix_items` 中无学历建议
- **Quality tier (旧结果)**：FAIL（旧数据 ID 66 仍有学历建议）— 符合预期，因为旧结果在修复前生成
- **Quality tier (修复后)**：PASS — 所有 fix_items 中的学历建议被过滤

### 7.2 本科及以上 + 简历已有本科（Test 3）

- **Job**：`前端开发工程师`，Requirements = "本科及以上学历"
- **Resume**：已有本科
- **Filter**：`shouldAllowEducationAdvice()` 返回 false
- **Result**：PASS — 所有学历风险和建议被过滤

### 7.3 高学历导向岗位（Test 2）

- **Job**：`算法研究员-多模态方向`，Requirements = "硕士及以上学历，博士优先，有顶会论文"
- **Resume**：仅本科
- **Filter**：`shouldAllowEducationAdvice()` 返回 true
- **Result**：PASS — 学历相关风险被保留

### 7.4 Mock / fallback 路径

- **代码确认**：Mock/fast_tag_engine/3-agent 所有路径均调用 `filterEducationOptimizationAdvice`
- **Result**：PASS — 即使 Mock 未来生成学历建议，也会被过滤

### 7.5 关键词匹配精度（Test 4）

- Pass cases：学历/学位/硕士/博士/985/211/双一流/教育经历/学历层次 等正确匹配
- Not-match cases：技术术语（Go/MySQL/Redis/QPS）不误删

### 7.6 简历教育检测（Test 5）

- Educations 数组中有数据 → 检测到
- Shim 字段有 school/degree → 检测到
- RawText 含"本科" → 检测到
- 纯技术描述 → 不误判

## 8. 旧结果处理说明

### 当前数据库状态

```
ID 66 (quality, 2026-06-14 15:40:41)：含学历建议（修复前生成）
  risk_points: ["简历未明确标注学历信息，不符合岗位学历要求"]
  quick_fix_items: [{"issue": "未明确标注学历层次", "section": "教育背景"}]
  content_fix_items: [{"issue": "无学历层次、院校、毕业时间相关内容"}]

ID 64-65 (fast, 2026-06-14)：不含学历建议（fast_tag_engine 原本就不生成学历建议）
```

### 策略

1. **旧 `result_json` 不会自动变化**：`LatestV2` 直接读取数据库的 `result_json`，不经过过滤器
2. **旧结果不会被自动清理**：不会执行 DELETE/UPDATE/TRUNCATE
3. **如何获得干净结果**：重新调用 `POST /api/resumes/{id}/optimization/v2` 生成新优化
4. **如何排查旧结果干扰**：检查 `latest` 接口返回的 `generated_by` 和 `created_at`，如果在修复前生成则可能是旧结果

## 9. 本次未修改范围

- 未修改前端 ✅
- 未修改数据库 ✅
- 未修改岗位匹配分析（`match_deep_agent.go`）✅
- 未修改简历版本对比 ✅
- 未执行 Git ✅
- 未修改 `.env` ✅

## 10. 构建结果

```
go test ./...  → 全部通过（含 5 个新增验收测试）
go vet ./...   → 通过
go build ./... → 通过
```

## 11. 后续建议

### Prompt 2 已完成

验收全部通过。普通开发岗位不再输出学历焦虑建议，高学历研究岗可保留合理学历风险。

### 是否需要 Prompt 3

**建议不需要（暂缓）**。

原因：
1. Prompt 1 的 Prompt 约束 + 后端过滤已覆盖所有 V2 优化生成路径
2. 当前匹配分析（`match_deep_agent.go`）中的 `education → mustHave` 分类虽然偏激进，但已通过优化结果过滤器拦截，不会传递给前端
3. 如果未来需要直接在匹配分析层面收敛学历风险，再进入 Prompt 3

### 生产环境接入建议

1. 重新生成所有简历优化结果（或至少对关键简历执行重新优化）
2. 观察 1-2 周，确认普通岗位不再出现学历建议
3. 如仍有问题，检查是否 `latest` 接口读取了旧缓存的旧结果
