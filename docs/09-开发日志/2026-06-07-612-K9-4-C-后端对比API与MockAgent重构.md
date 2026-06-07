# K9-4-C — 后端对比 API 与 Mock Agent 重构 ✅

**日期**: 2026-06-07  
**阶段**: 实现  
**关联**: K9-4-A（诊断）、K9-4-B（设计冻结）

---

## 1. 完成目标

### DB 变更
- `resume_compare_results` 新增 11 列：comparison_mode/generation_tier/generated_by/status/target_direction/job_id/from_profile_*/to_profile_*/updated_at
- 3 个新索引
- 不修改 init.sql ✅

### API
- `POST /api/resumes/compare` — 新版对比 API
- 三模式：general / target_direction / target_job
- Fast / Quality 双阶段
- 旧 GET `/versions/compare` 保留向后兼容

### 数据持久化
- LLM/Mock 路径均保存到 DB（修复 compare_id=0）
- metadata 完整：generation_tier/generated_by/status + profile 字段
- result_json 与 DB metadata 一致

### ResumeProfile 消费
- from/to 两份 Profile 通过 EnsureProfile 获取
- profile_dominant_profile 写入响应和 DB

### 编译验证
- go build ✅
- go vet ✅
- npm build ✅
- check-encoding ⚠️（init.sql 已知误报）

---

## 2. 修改清单

| 文件 | 修改 |
|---|---|
| `model/version_compare.go` | 新增 CompareRequestV2/ResponseV2/RewriteCompareExample/ResumeCompareResultV2 |
| `repository/version_compare_repository.go` | 新增 CreateV2 + toNullStr/toNullUI64 辅助 |
| `agent/version_compare_mock.go` | 新增 CompareV2 + ProfileProvider + 全套 profile-aware 助手函数 |
| `service/version_compare_service.go` | 新增 CompareV2 + jobRepo/profileSvc 依赖 |
| `http/version_compare_handler.go` | Compare 方法支持 GET（旧）+ POST（V2）分发 |
| `cmd/server/main.go` | 注入 jobRepo/profileSvc 到 versionCompareService |

---

## 3. 测试验证

| 测试 | 结果 |
|---|---|
| POST general fast (1 vs 2) | compare_id=1, tier=fast, profile 正确 ✅ |
| POST target_direction quality (1 vs 2, backend) | compare_id=2, rewrite=2, merge=3, next=3 ✅ |
| DB 保存 | id=1/2 均有完整 metadata ✅ |

---

## 4. 约束确认

- [x] 未修改前端源码
- [x] 未修改 init.sql
- [x] 未修改 K9-1/K9-2/K9-3 逻辑
- [x] 未删除真实简历
- [x] 未执行 reset-up
- [x] 未执行 Git 命令

## 5. 后续阶段

K9-4-D：compare latest API 与持久化读取实现
K9-4-E：前端对比页 latest-first + Fast/Deep 展示分层实现
K9-4-F：5 类样例版本对比总验收
