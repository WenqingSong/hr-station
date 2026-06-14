# Deploy-Precheck-No-Data-Loss

日期：2026-06-14 | 类型：部署前只读检查 | 状态：完成

---

## 检查范围

- `/workspace/resources/sql-add` — 历史增量 SQL 记录
- `server/sql/init.sql` — 全量初始化 SQL
- `server/sql/migrations/` — 历史迁移 SQL
- `docker-compose.yml` / `docker-compose.prod.yml` — 容器配置
- `Makefile` — 危险目标扫描
- 数据库只读检查 — 表结构 / 数据量
- 构建检查 — go build / go vet / go test / npm build

---

## 核心结论

- **允许部署**：是 ✅（满足条件）
- **存在删库/删表/清空数据风险**：是 ⚠️（init.sql 含 DROP TABLE IF EXISTS，生产环境不应执行）
- **是否需要增量 SQL**：否（当前 DB schema 已完整，代码不依赖迁移中的可选字段）
- **是否必须先备份**：是 ✅ 强烈建议

---

## 历史部署记录审计

| 文件 | 内容 | 状态 |
|------|------|------|
| `resources/sql-add/20260612_2100_sep_recommend_datalayer.sql` | 创建 `semantic_profiles` + `semantic_profile_tags` 表 | ✅ 已部署（DB 中存在） |
| `resources/sql-add/20260612_2130_sep_weight_fix.sql` | ALTER TABLE 扩宽 weight/confidence 精度 | ✅ 已部署 |

---

## SQL 风险审计

| 文件 | 风险 | 等级 |
|------|------|------|
| `server/sql/init.sql` | `DROP TABLE IF EXISTS` × 24 + 全量 `INSERT INTO`（含用户数据） | **P0** |
| `server/sql/seed/tencent_jobs.sql` | `DELETE FROM jobs WHERE external_source IN (...)` + `DELETE FROM job_filter_*` | P1 |
| `server/sql/migrations/20260611_k13_result_tables_add_agent_fields.sql` | `ADD COLUMN IF NOT EXISTS`（幂等安全），但当前 DB 无这些列（数据存 JSON 字段，无需执行） | P2 |

---

## Docker / Volume 审计

| 检查项 | 结果 |
|--------|------|
| MySQL 是否 volume 持久化 | ✅ `mysql_data` named volume |
| Redis 是否 volume 持久化 | ✅ `redis_data` named volume |
| 生产环境 has `init.sql` auto-mount | ⚠️ `docker-compose.prod.yml` 第 31 行挂载 `init.sql` → `/docker-entrypoint-initdb.d/`（**只在 MySQL data dir 为空时触发**） |
| `make reset` 会删 volume | ⚠️ 需要 `CONFIRM=yes`，执行 `docker compose down -v` |
| `make clean-volume` 会删 volume | ⚠️ 需要 `CONFIRM=yes` |
| `make down` 保留 volume | ✅ |
| `make stop` 保留数据 | ✅ |

---

## 数据库只读检查

| 表 | 数据量 | 状态 |
|----|--------|------|
| users | 3 | ✅ |
| resumes | 8 | ✅ |
| jobs | 46 | ✅ |
| match_analyses | 113 | ✅ |
| recommendation_results | 48 | ✅ |
| resume_optimization_results | 64 | ✅ |
| resume_compare_results | 0 | 空（无对比记录） |
| semantic_profiles | 48 | ✅ |
| semantic_profile_tags | 722 | ✅ |
| resume_profiles | 16 | ✅ |

---

## 代码与 DB 字段一致性

| 模块 | 结论 |
|------|------|
| `resume_optimization_results` | ✅ 14 列全匹配 Go model 和实际 DB |
| `match_analyses` | ✅ 13 列全匹配 |
| `resumes` / `jobs` | ✅ 主表全匹配 |
| 迁移文件 `k13_result_tables_add_agent_fields.sql` | ⚠️ 列未实际添加到 DB — 但 Go 代码将这些字段序列化到 `result_json` 中，**无需执行迁移** |

---

## 构建检查

| 项目 | 结果 |
|------|------|
| `go build ./...` | ✅ PASS |
| `go vet ./...` | ✅ PASS |
| `go test ./...` | ✅ ALL PASS |
| `npm run build` (web) | ✅ PASS (built in 2.49s) |

---

## 本次修改范围

| 组件 | 说明 |
|------|------|
| 后端 | `resume_optimization_service.go` +197 行（学 limith 制过滤器 + Prompt 约束） |
| Prompt | 3 个 markdown 文件新增学 limith 制 |
| 前端 | 未修改 ✅ |
| 数据库 schema | 未修改 ✅ |

---

## 部署建议

### 安全部署步骤

1. **备份数据库**
   ```bash
   docker compose exec mysql mysqldump -uroot -p<密码> offer_catcher > backup_$(date +%F_%H%M%S).sql
   ```

2. **构建新镜像**
   ```bash
   docker compose -f docker-compose.prod.yml build
   ```

3. **启动新版本**
   ```bash
   docker compose -f docker-compose.prod.yml up -d
   ```

4. **Smoke test**（见下方清单）

5. **观察日志**
   ```bash
   docker compose -f docker-compose.prod.yml logs -f server
   ```

### 严禁执行

- `mysql < server/sql/init.sql` (DROP TABLE IF EXISTS!)
- `docker compose down -v`
- `make reset CONFIRM=yes`
- `make clean-volume CONFIRM=yes`
- `make db-reset CONFIRM=yes`
