# 10-P-4-1 修复数据库 schema 与核心表缺失 P0

## 1. 阶段信息

- 阶段：10-P-4-1
- 日期：2026-06-02
- 范围：补齐 `resume_optimization_results` 表 DDL
- 是否修改代码：是
- 是否修改前端：否
- 是否修改后端：是（仅 init.sql）
- 是否接 LLM：否
- 是否 git commit：是

## 2. 本次结论摘要

- P0 `resume_optimization_results` 表缺失已修复
- init.sql 新增 17 行 DDL（建表 + 索引 + 外键）
- 表字段与代码完全一致（user_id, job_id, result_json, created_at），匹配 `ResumeOptimizationSnapshot` 模型和 repository INSERT 语句
- 索引覆盖 user_id+created_at 联合查询和 job_id 单列查询
- 外键约束 user_id → users(id)，job_id 无 FK（允许 0 作为无岗位引用）
- go build / go vet / go test 全部通过
- 旧环境迁移 SQL 已提供
- 已 commit

## 3. P0 根因

| 问题 | 证据 | 影响 | 修复方式 |
|---|---|---|---|
| `resume_optimization_results` 表在 init.sql 中完全缺失 | `resume_optimization_repository.go:23` 执行 `INSERT INTO resume_optimization_results`；init.sql 全文无此表 DDL | `POST /api/resumes/optimize` 调用 `saveSnapshot` 时数据库写入失败，优化结果无法持久化 | 在 init.sql 中按代码真实字段补齐完整 DDL |

## 4. schema 变更

| 表 | 字段 / 索引 | 目的 | 说明 |
|---|---|---|---|
| `resume_optimization_results` | `id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY` | 主键 | 自增 |
| | `user_id BIGINT UNSIGNED NOT NULL` | 用户隔离 | 外键 → users(id) |
| | `job_id BIGINT UNSIGNED NOT NULL DEFAULT 0` | 关联岗位（可选） | 0 表示无岗位引用 |
| | `result_json JSON NOT NULL` | 优化结果快照 | 序列化 `ResumeOptimizationResult` |
| | `created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP` | 创建时间 | 自动填充 |
| | `KEY idx_opt_user_created (user_id, created_at)` | 用户维度查询 | 覆盖 GetLatestByUserID 查询 |
| | `KEY idx_opt_job (job_id)` | 岗位维度查询 | 按岗位查询优化历史 |
| | `CONSTRAINT fk_opt_user FOREIGN KEY (user_id) REFERENCES users(id)` | 用户数据完整性 | 级联保护 |

## 5. 代码变更清单

| 文件 | 修改内容 | 是否业务代码 | 说明 |
|---|---|---|---|
| `server/sql/init.sql` | 新增 `resume_optimization_results` 表的 DROP + CREATE；更新文件头注释 | 否（schema） | 17 行新增，DROP 插入正确删除顺序位置 |

## 6. 旧环境迁移说明

- **新环境**：init.sql 已包含完整 DDL，`docker compose up -d` 或重新导入即可
- **旧容器**：已有数据库需手动执行以下 SQL：

```sql
CREATE TABLE `resume_optimization_results` (
  `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id`     BIGINT UNSIGNED NOT NULL,
  `job_id`      BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '可选，0 表示无岗位引用',
  `result_json` JSON             NOT NULL,
  `created_at`  DATETIME         NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_opt_user_created` (`user_id`, `created_at`),
  KEY `idx_opt_job` (`job_id`),
  CONSTRAINT `fk_opt_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='简历优化结果表';
```

- **当前环境**：本地无 MySQL 容器运行，SQL 未实际执行，但 init.sql 已更新

## 7. 验证结果

| 命令 | 结果 | 说明 |
|---|---|---|
| `git status --short` 执行前 | 干净 | 零修改 |
| `go build ./cmd/server` | ✅ PASS | 编译通过 |
| `go vet ./...` | ✅ PASS | 零告警 |
| `go test ./...` | ✅ PASS | 9 packages PASS |
| `git diff --stat` | 1 file, +17 lines | 仅 init.sql |
| `git status --short` 执行后 | 仅 server/sql/init.sql | 修改范围符合边界 |

## 8. 未做事项

- 未修改前端
- 未修改 Jobs
- 未修改 Auth
- 未修改 Agent / LLM
- 未统一错误码
- 未添加 GetLatest 方法（模型仅有 Create，GetLatest 需后续补）
- 未接真实 LLM
- 未做上线部署

## 9. 当前风险与后续建议

- **migration 体系缺失**：项目仅有全量 init.sql，无增量 migration。旧环境需手动执行 SQL。建议后续引入简单 migration 机制
- **Repository 仅有 Create 方法**：`ResumeOptimizationRepository` 当前仅 `Create`，无 `GetLatestByUserID` 方法。如果后续需要查询优化历史，需补充（P1-DB3 仍待处理）
- **错误码统一**：P1-ERR1/ERR2 仍待 10-P-4 后续子阶段处理
- **Agent 边界整理**：P1-AGT1（VersionCompare 无 LLM）仍待处理
- **本阶段仅修 P0-DB1**，其余 10-P-4-0 发现的问题未触碰

## 10. docs 更新清单

| 文件 | 新增 / 更新 | 目的 |
|---|---|---|
| `docs/mcp/项目上下文.md` | 更新 | 更新 10-P-4-1 阶段状态 |
| `docs/mcp/10-P-4-1-修复数据库schema与核心表缺失P0.md` | 新增 | 本阶段 MCP 记录 |
| `docs/09-开发日志/2026-06-02-240-10-P-4-1-修复数据库schema.md` | 新增 | 开发日志 |

## 11. commit 信息

- 是否 commit：是
- commit message：`fix(server): add resume optimization results schema`
- commit hash：见下方

## 12. CodeBuddy 重启恢复摘要

10-P-4-1 修复了 10-P-4-0 发现的 P0：`resume_optimization_results` 表在 init.sql 中缺失。表 DDL 与代码模型/Repository 字段完全一致（user_id/job_id/result_json/created_at），含覆盖索引和用户外键约束。旧环境迁移 SQL 已提供。仅修改 1 个文件 17 行。go build/vet/test 全部通过。其余 10-P-4-0 发现的 P1/P2 未处理，留待后续子阶段。

下次继续前读取：`docs/mcp/项目上下文.md`、本文件、`docs/mcp/10-P-4-0-后端现状清查与专业化问题分级.md`。

## 13. 最终判断

1. **`resume_optimization_results` 表缺失 P0 是否已修复？** **是。** init.sql 已补齐完整 DDL。

2. **新环境 init.sql 是否已补齐？** **是。** DROP + CREATE 均按正确顺序插入。

3. **旧环境迁移风险是否已记录？** **是。** 手动迁移 SQL 已提供。

4. **go build / go vet / go test 是否通过？** **是。** 全部通过。

5. **是否修改前端？** **否。**

6. **是否接入 LLM？** **否。**

7. **是否已 commit？** **是。**

8. **本阶段是否可以通过？** **是。**
