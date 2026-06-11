# Dev-Deploy-Fix-A：修复线上部署前配置、迁移与检查缺口

日期：2026-06-11

## 背景

`Dev-Deploy-Prepare-A` 识别 3 个 P0/P1 阻断项：
1. `.env.prod` 缺失
2. 旧 MySQL volume 不会自动执行 init.sql，需 migration
3. `LLM_FALLBACK_ENABLE` 代码默认 true，生产需显式 false

## 修复内容

### 1. .env.prod 创建
- `APP_ENV=production`
- `LLM_PROVIDER=hy3`、`LLM_MODEL=hy3-preview`
- `ENABLE_MOCK_AGENT=false`、`LLM_FALLBACK_ENABLE=false`
- `LLM_API_KEY`、`MYSQL_ROOT_PASSWORD`、`JWT_SECRET` 从 `.env` 复制
- 权限 600

### 2. 生产 job_profiles migration
- `server/sql/migrations/20260611_prod_job_profiles.sql`
- `CREATE TABLE IF NOT EXISTS` — 幂等，可重复执行
- 含 `UNIQUE KEY uk_job_id (job_id)`
- 适用于已有 MySQL volume 的生产库

### 3. make prod-config / prod-check
- `make prod-config`：验证 compose config（不启动）
- `make prod-check`：10 项安全校验
  - .env.prod 存在
  - ENABLE_MOCK_AGENT=false
  - LLM_FALLBACK_ENABLE=false
  - LLM_PROVIDER=hy3
  - 密钥存在
  - compose config 通过
  - 失败时非零退出

## 验证

### prod-check ✅
```
✅ .env.prod
✅ docker-compose.prod.yml
✅ LLM_PROVIDER=hy3
✅ LLM_MODEL=hy3-preview
✅ ENABLE_MOCK_AGENT=false
✅ LLM_FALLBACK_ENABLE=false
✅ MYSQL_ROOT_PASSWORD=EXISTS
✅ JWT_SECRET=EXISTS
✅ LLM_API_KEY or LLM_HY3_API_KEY exists
✅ prod compose config passes
✅ prod-check PASSED
```

### 开发回归 ✅
- npm build ✅
- go build ✅
- go vet ✅
- go test ./server/... 10/10 OK ✅
- make check 8 表通过 ✅

## 未修改
- 后端源码
- 前端 UI
- 数据库数据
- Git 操作
- 未执行实际部署

## 结论
部署前准备完成。可进入 `Dev-Deploy-Plan-B`。
