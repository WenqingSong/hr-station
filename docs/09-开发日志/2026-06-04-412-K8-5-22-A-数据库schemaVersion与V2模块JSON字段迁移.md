# K8-5-22-A — 数据库 schema_version + V2 模块 JSON 字段迁移

> 日期：2026-06-04 | 性质：implementation | 仅修改 SQL 文件

---

## 1. 本阶段目标

执行 K8-5-22-A：在 `resume_versions` 表中新增 V2 所需列，不改 Go/TS 源码。

---

## 2. 当前数据库结构确认

### resume_versions 表（CREATE TABLE 在 init.sql L59-80）

| 列 | 类型 | 说明 |
|-----|------|------|
| id | BIGINT PK | 自增 |
| user_id | BIGINT FK | 用户外键 |
| version_no | INT | 版本号 |
| name | VARCHAR(100) | 简历名称 |
| phone | VARCHAR(50) | 手机号 |
| email | VARCHAR(255) | 邮箱 |
| location | VARCHAR(100) | 位置 |
| target_city | VARCHAR(100) | 期望城市 |
| target_role | VARCHAR(100) | 目标岗位 |
| education_json | JSON | 教育经历 |
| experience_json | JSON | 实习经历 |
| project_json | JSON | 项目经历 |
| skills_json | JSON | 技能关键词 |
| self_evaluation | TEXT | 自我评价 |
| raw_text | TEXT | 原始文本 |
| created_at | DATETIME | 创建时间 |

**V2 字段全部不存在。**

---

## 3. 修改方案

| 区域 | 修改内容 | 原因 |
|------|---------|------|
| **init.sql** | CREATE TABLE 新增 16 列（1 schema_version + 7 摘要列 + 8 JSON 列） | 新部署时自动创建 V2 列 |
| **migration-j2.sql** | 新增 ALTER TABLE 16 条 | 已有数据库在线迁移，不改代码 |
| **种子数据** | 不改 | 显式列名 INSERT，新列取默认值 |
| **旧字段** | 全部保留 | V1 兼容 |

---

## 4. 新增字段清单

### schema_version

| 字段 | 类型 | 默认值 | 用途 |
|------|------|:---:|------|
| schema_version | INTEGER | 1 | V1=1, V2=2 |

### 摘要结构化列（7 个）

| 字段 | 类型 | 默认值 | 用途 |
|------|------|:---:|------|
| target_direction | VARCHAR(100) | '' | 目标方向，替代 target_role |
| job_type | VARCHAR(50) | '' | 岗位类型 |
| expected_cities | VARCHAR(255) | '' | 期望城市逗号分隔 |
| school | VARCHAR(100) | '' | 学校 |
| degree | VARCHAR(20) | '' | 学历 |
| major | VARCHAR(100) | '' | 专业 |
| graduation_year | VARCHAR(10) | '' | 毕业时间 |

### 模块 JSON 字段（8 个）

| 字段 | 类型 | 默认值 | 用途 |
|------|------|:---:|------|
| profile_json | JSON | NULL | 基础信息 |
| job_intention_json | JSON | NULL | 求职意向 |
| internship_json | JSON | NULL | 实习经历（与 experience_json 并存） |
| certificates_json | JSON | NULL | 证书与语言 |
| awards_json | JSON | NULL | 获奖与竞赛 |
| portfolio_json | JSON | NULL | 作品链接 |
| extra_json | JSON | NULL | 补充信息 |

---

## 5. 旧字段保留确认

| 旧字段 | 保留 | 说明 |
|--------|:---:|------|
| target_role | ✅ | V1 继续使用 |
| target_city | ✅ | V1 继续使用 |
| education_json | ✅ | V1 继续使用 |
| experience_json | ✅ | V1 继续使用（V2 新增 internship_json 并存） |
| project_json | ✅ | V1 继续使用 |
| skills_json | ✅ | V1 继续使用 |
| self_evaluation | ✅ | V1 继续使用（V2 新增 extra_json 双存） |
| location | ✅ | V1 继续使用 |
| raw_text | ✅ | V1 继续使用 |

---

## 6. 修改清单

| 文件 | 修改内容 | 属于 K8-5-22-A |
|------|---------|:---:|
| `server/sql/init.sql` | CREATE TABLE 新增 16 列 + header 注释 | ✅ |
| `server/sql/migration-j2.sql` | **新增** 18 行 ALTER TABLE + 注释 | ✅ |
| Go 源码 | 未修改 | ❌ 不属于 |
| TypeScript 源码 | 未修改 | ❌ 不属于 |
| 前端页面 | 未修改 | ❌ 不属于 |

---

## 7. 兼容性说明

```
1. V1 旧数据：schema_version=1（默认值），所有新列为 NULL/空字符串。
2. 本阶段只加字段，不让后端写 V2。
3. 旧系统仍按 V1 字段工作（repository/service/handler 未改）。
4. 种子数据 INSERT 显式列名，新列自动取默认值。
5. 后续 K8-5-22-B 会修改后端代码写入 V2 字段。
```

---

## 8. 构建 / API / 测试结果

| 验收项 | 结果 |
|--------|------|
| `npm run build` | ✅ 通过 (2.29s) |
| `GET /api/jobs?page=1&page_size=5` | ✅ 200 |
| `GET /api/jobs/10001` | ✅ 200 |
| `GET /api/jobs/facets` | ✅ 200 |
| `GET /api/resumes` | ✅ 401 |
| `go test ./...` | ✅ 全部通过 |

---

## 9. SQL 检查结果

| 检查项 | 结果 |
|--------|:---:|
| V2 字段在 init.sql 中存在 | ✅ 16 列 |
| V2 字段在 migration-j2.sql 中存在 | ✅ 16 条 ALTER |
| 旧 V1 字段在 init.sql 中保留 | ✅ 9 列全部 |
| migration-j2.sql 无破坏性 SQL | ✅ 无 DROP/TRUNCATE/DELETE |
| 种子数据 INSERT 兼容新列 | ✅ 显式列名 |

---

## 10. 未修改范围确认

```
未修改 Go 源码
未修改 TypeScript 源码
未修改前端页面
未修改后端 handler/service/repository/model
未修改 Agent
未修改 Prompt / Schema
未修改 Mock
未修改 API 逻辑
未修改 .env
未安装依赖
未迁移目录
未删除旧字段
未执行破坏性 SQL
```

---

## 11. docs 更新清单

| 操作 | 文件 |
|:---:|------|
| ✅ 新增 | `server/sql/migration-j2.sql` |
| ✅ 更新 | `server/sql/init.sql` |
| ✅ 新增 | `docs/09-开发日志/2026-06-04-412-K8-5-22-A-数据库schemaVersion与V2模块JSON字段迁移.md` |
| ⬜ | `docs/下一步任务.md` |
| ⬜ | `docs/变更记录.md` |

---

## 12. 风险与边界

```
本阶段只完成数据库字段准备。16 个新列已就绪，schema_version 默认 1。
后端 V2 写入读取留给 K8-5-22-B。
NormalizeResumeToSnapshotV2 留给 K8-5-22-C。
前端 V2 类型和页面留给后续阶段。
```

---

## 13. Git 约束确认

```
本次未执行任何 Git 命令，也未提交代码。
```
