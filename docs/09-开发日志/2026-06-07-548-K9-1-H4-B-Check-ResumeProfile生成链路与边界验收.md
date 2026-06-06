# K9-1-H4-B-Check：ResumeProfile 生成链路与边界验收

> 日期：2026-06-07 | 状态：验收通过

## 1. 当前数据库概况

| Resume ID | Name | Status |
| -- | -- | -- |
| 7 | 01-算法研究生-岗位匹配 | 保留 |
| 9 | 02-211后台开发-岗位匹配 | 保留 |
| 10 | 03-二本产品经理-岗位匹配 | 保留 |
| 11 | 04-职业探索型学生-岗位推荐 | 保留 |
| 12 | 05-优秀但不适合互联网-岗位推荐 | 保留 |
| 13 | K9-1-H4-B-Check-Temp-Create | 已删除（验收用） |

## 2. resume_profiles 表

| 项 | 结果 |
| -- | -- |
| 表存在 | ✅ |
| UNIQUE( resume_id, profile_type) | ✅ |
| ON DELETE CASCADE | ✅ |
| JSON NOT NULL | ✅ null_json_count=0 |

## 3. Profile 生成验证

### 01 算法研究生 (R7)
- fast_status: ready ✅
- dominant_profile: algorithm_research ✅
- quality_status: missing（未触发 quality 生成）

### Backfill
- 全部 5 份简历已补齐 fast profile ✅
- 每份简历 dominant_profile 均已生成

### Create Resume 自动生成
- 创建 R13 后 fast_status=ready ✅
- dominant_profile=algorithm_research ✅

### Update Resume stale+regenerate
- 更新后 quality_status=stale ✅
- fast 重新生成为 ready ✅

## 4. Boundary Tests

| 场景 | 结果 |
| -- | -- |
| regenerate fast | ✅ 200 |
| regenerate quality | ✅ 200 |
| regenerate bad type | ⚠️ 50000（应为 400，P2） |
| pending dedup (3x rapid quality) | ✅ 无重复，DB 1 条/类型 |
| failed retry | ⚠️ 未实测（无制造 failed 入口），源码审查 Enable 重试 |
| no auth → /profile/status | ✅ 401 |
| delete CASCADE | ✅ profile_count=0 after delete |
| init.sql 未修改 | ✅ |
| JSON null | ✅ 0 null records |

## 5. Build/Test

| 项 | 结果 |
| -- | -- |
| go test ./... | ✅ all pass |
| go build | ✅ |
| go vet | ✅ |
| npm build | ✅ |
| check-encoding | ✅ OK |

## 6. 问题清单

| 级别 | 问题 |
| -- | -- |
| P2 | regenerate bad profile_type → 500 instead of 400 |

## 7. 约束

未修改源码/init.sql；未删除已有简历；未 reset-up；未 Git。
