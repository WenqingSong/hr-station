# SemanticProfileGate-Implement-E-Fix — 修复 verified profile 无 tags 并重跑 Job Gate

## 目标
修复上一阶段遗留的 job_id=10001 verified profile 无 tags 问题，增强 Gate 完整性检查与持久化一致性。

## 根因分析

| 问题 | 根因 | 修复 |
|---|---|---|
| Job verified but 0 tags | weight DECIMAL(3,2) 精度不足导致 UpsertTags 失败，但 `SaveVerifiedProfile` 先写 profile（status=verified）再写 tags，写入失败后未回滚 profile 状态 | 新增 tag count 后校验 + CreateProfile 重复键自动更新 |
| Gate 错误复用坏 profile | Gate 只检查 `status==verified`，未检查 `tag_count >= 5` | Gate 增加 tag_count 完整性检查 |
| 重复键导致重跑失败 | 旧 failed profile 占用唯一键 `(target_type, target_id, version, schema, hash)` | `CreateProfile` 遇到重复键时更新已有非 verified 记录 |

## 修改文件

| 文件 | 修改内容 |
|---|---|
| `repository/semantic_profile_repository.go` | +`CountTagsByProfileID`、+`isDuplicateKeyError`、+`findAnyByTargetAndHash`、+`updateProfileAll`、`CreateProfile` 支持重复键复用 |
| `service/semantic_profile_service.go` | +log import、`SaveVerifiedProfile` 增加 tag count 后校验（<5 → failed）|
| `service/semantic_profile_gate.go` | 复用前增加 `tag_count >= 5` 完整性检查，不足时 marked failed 并重新生成 |

## 验证结果

| 项目 | 结果 |
|---|---|
| Job Gate 第一次 | Hy3 high 19249ms → tags=15 → verified ✅ |
| Job Gate 第二次 | reused=true, 0ms ✅ |
| Resume Gate 回归 | reused=true, 14 tags ✅ |
| 相同 hash 重复 verified | 0 ✅ |
| go build/vet/test | 全部通过 ✅ |
| make check | 全部通过 ✅ |

## 下一步
进入 GQCL-Fast-Composer-F 或 Fast-Recommend-SEP-Integration-G。
