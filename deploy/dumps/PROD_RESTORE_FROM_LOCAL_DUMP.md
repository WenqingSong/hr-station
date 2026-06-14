# Offer Catcher 生产环境数据库全库替换操作手册

> **重要：本文档描述全库替换流程，会清空并重建生产数据库。执行前必须完成备份。**

## 一、前置条件

### 1.1 文件准备

将以下文件上传到生产服务器：

```bash
# 在生产服务器创建目录
mkdir -p /opt/wenqing-hr-demo/deploy/dumps

# 从本地/开发机上传文件（scp 示例）
scp deploy/dumps/offer_catcher_full_dump_20260615_051023.sql.gz \
    user@prod-server:/opt/wenqing-hr-demo/deploy/dumps/

scp deploy/dumps/offer_catcher_full_dump_20260615_051023.check.txt \
    user@prod-server:/opt/wenqing-hr-demo/deploy/dumps/
```

### 1.2 校验文件完整性

```bash
cd /opt/wenqing-hr-demo/deploy/dumps

# 校验 sha256
sha256sum offer_catcher_full_dump_20260615_051023.sql.gz
# 期望: 585c9e468dfd15d7b96cb1a3dde7a63cc0c493d10e84206d4d49e62cb8b6521f

# 校验 gzip 完整性
gunzip -t offer_catcher_full_dump_20260615_051023.sql.gz
# 期望输出: 无输出（表示正常）
```

---

## 二、导入前备份

### 2.1 备份当前生产数据库

```bash
cd /opt/wenqing-hr-demo

# 获取时间戳
TS=$(date +%Y%m%d_%H%M%S)

# 完整备份当前生产数据库
docker compose \
  --project-name wenqing-hr-demo \
  --env-file .env.prod \
  -f docker-compose.prod.yml \
  exec -T mysql \
  mysqldump \
  -uroot \
  -p"${MYSQL_ROOT_PASSWORD}" \
  --single-transaction \
  --quick \
  --routines \
  --triggers \
  --events \
  --hex-blob \
  --default-character-set=utf8mb4 \
  --set-gtid-purged=OFF \
  --skip-column-statistics \
  --no-tablespaces \
  offer_catcher \
  > backups/prod_backup_before_restore_${TS}.sql

# 压缩备份
gzip backups/prod_backup_before_restore_${TS}.sql

echo "BACKUP: backups/prod_backup_before_restore_${TS}.sql.gz"
```

### 2.2 记录当前数据库状态

```bash
docker compose \
  --project-name wenqing-hr-demo \
  --env-file .env.prod \
  -f docker-compose.prod.yml \
  exec -T mysql \
  mysql -uroot -p"${MYSQL_ROOT_PASSWORD}" \
  -e "SELECT TABLE_NAME, TABLE_ROWS FROM information_schema.TABLES WHERE TABLE_SCHEMA='offer_catcher' ORDER BY TABLE_NAME;" \
  > backups/prod_tables_before_restore_${TS}.txt
```

---

## 三、停止应用服务（保留数据库）

```bash
cd /opt/wenqing-hr-demo

# 只停止 server 和 web，不停止 mysql/redis/nginx
docker compose \
  --project-name wenqing-hr-demo \
  --env-file .env.prod \
  -f docker-compose.prod.yml \
  stop server web

# 确认 mysql 和 redis 仍在运行
docker compose \
  --project-name wenqing-hr-demo \
  --env-file .env.prod \
  -f docker-compose.prod.yml \
  ps mysql redis
```

---

## 四、导入前确认

```bash
# 加载环境变量
source .env.prod  # 或手动 export

# 确认数据库可用
docker compose \
  --project-name wenqing-hr-demo \
  --env-file .env.prod \
  -f docker-compose.prod.yml \
  exec -T mysql \
  mysqladmin ping -h localhost -uroot -p"${MYSQL_ROOT_PASSWORD}"

# 确认当前数据库存在
docker compose \
  --project-name wenqing-hr-demo \
  --env-file .env.prod \
  -f docker-compose.prod.yml \
  exec -T mysql \
  mysql -uroot -p"${MYSQL_ROOT_PASSWORD}" \
  -e "SHOW DATABASES LIKE 'offer_catcher';"
```

---

## 五、清空并重建目标数据库

```bash
# WARNING: 此操作不可逆！执行前请确认备份已完成。

docker compose \
  --project-name wenqing-hr-demo \
  --env-file .env.prod \
  -f docker-compose.prod.yml \
  exec -T mysql \
  mysql -uroot -p"${MYSQL_ROOT_PASSWORD}" <<'EOSQL'
DROP DATABASE IF EXISTS offer_catcher;
CREATE DATABASE offer_catcher CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
EOSQL

echo "Database dropped and recreated."
```

---

## 六、导入 SQL dump

```bash
cd /opt/wenqing-hr-demo

# 解压并导入
gunzip -c deploy/dumps/offer_catcher_full_dump_20260615_051023.sql.gz | \
  docker compose \
    --project-name wenqing-hr-demo \
    --env-file .env.prod \
    -f docker-compose.prod.yml \
    exec -T mysql \
    mysql -uroot -p"${MYSQL_ROOT_PASSWORD}" \
    --default-character-set=utf8mb4 \
    offer_catcher

echo "Import completed. Check result below."
```

> 预计耗时：< 5 秒（2.7MB SQL 解压后）

---

## 七、导入后验证

### 7.1 基本健康检查

```bash
# 表数量（期望 25）
docker compose --project-name wenqing-hr-demo --env-file .env.prod -f docker-compose.prod.yml \
  exec -T mysql mysql -uroot -p"${MYSQL_ROOT_PASSWORD}" \
  -s -N -e "SELECT COUNT(*) FROM information_schema.TABLES WHERE TABLE_SCHEMA='offer_catcher';"

# 数据库字符集
docker compose --project-name wenqing-hr-demo --env-file .env.prod -f docker-compose.prod.yml \
  exec -T mysql mysql -uroot -p"${MYSQL_ROOT_PASSWORD}" \
  -s -N -e "SELECT DEFAULT_CHARACTER_SET_NAME FROM information_schema.SCHEMATA WHERE SCHEMA_NAME='offer_catcher';"
```

### 7.2 关键表行数检查

```bash
docker compose --project-name wenqing-hr-demo --env-file .env.prod -f docker-compose.prod.yml \
  exec -T mysql mysql -uroot -p"${MYSQL_ROOT_PASSWORD}" offer_catcher \
  -s -N -e "
SELECT 'users', COUNT(*) FROM users
UNION ALL SELECT 'resumes', COUNT(*) FROM resumes
UNION ALL SELECT 'jobs', COUNT(*) FROM jobs
UNION ALL SELECT 'job_profiles', COUNT(*) FROM job_profiles
UNION ALL SELECT 'resume_profiles', COUNT(*) FROM resume_profiles
UNION ALL SELECT 'semantic_profiles', COUNT(*) FROM semantic_profiles
UNION ALL SELECT 'semantic_profile_tags', COUNT(*) FROM semantic_profile_tags
UNION ALL SELECT 'recommendation_results', COUNT(*) FROM recommendation_results
UNION ALL SELECT 'match_analyses', COUNT(*) FROM match_analyses
UNION ALL SELECT 'resume_optimization_results', COUNT(*) FROM resume_optimization_results
UNION ALL SELECT 'resume_compare_results', COUNT(*) FROM resume_compare_results
UNION ALL SELECT 'seed_versions', COUNT(*) FROM seed_versions;
"
```

期望结果对照（参考）：

| 表 | 期望行数 |
|---|---------|
| users | 3 |
| resumes | 8 |
| jobs | 46 |
| job_profiles | 46 |
| resume_profiles | 16 |
| semantic_profiles | 49 |
| semantic_profile_tags | 737 |
| recommendation_results | 50 |
| match_analyses | 113 |
| resume_optimization_results | 63 |
| resume_compare_results | 0 |
| seed_versions | 1 |

### 7.3 AUTO_INCREMENT 检查

```bash
docker compose --project-name wenqing-hr-demo --env-file .env.prod -f docker-compose.prod.yml \
  exec -T mysql mysql -uroot -p"${MYSQL_ROOT_PASSWORD}" \
  -s -N -e "
SELECT TABLE_NAME, AUTO_INCREMENT FROM information_schema.TABLES
WHERE TABLE_SCHEMA='offer_catcher' AND AUTO_INCREMENT IS NOT NULL
ORDER BY TABLE_NAME;
"
```

关键 AUTO_INCREMENT 期望：
- `jobs`: **10047**（下一个岗位 ID 将从 10047 开始）
- `users`: 4
- `resumes`: 9

---

## 八、重启应用服务

```bash
cd /opt/wenqing-hr-demo

docker compose \
  --project-name wenqing-hr-demo \
  --env-file .env.prod \
  -f docker-compose.prod.yml \
  up -d server web

# 等待服务就绪（约 10-15 秒）
sleep 15

# 查看服务状态
docker compose \
  --project-name wenqing-hr-demo \
  --env-file .env.prod \
  -f docker-compose.prod.yml \
  ps
```

---

## 九、验证接口

```bash
# 设置 BASE_URL（根据实际生产地址调整）
BASE_URL="https://your-production-domain.com"

# 1. 健康检查
curl -s -o /dev/null -w "healthz: HTTP %{http_code}\n" ${BASE_URL}/healthz
# 期望: HTTP 200

# 2. 岗位列表
curl -s -o /dev/null -w "jobs: HTTP %{http_code}\n" ${BASE_URL}/api/jobs?limit=5

# 3. 当前简历（需要认证 Token）
TOKEN="your-valid-jwt-token"
curl -s -o /dev/null -w "resumes: HTTP %{http_code}\n" \
  -H "Authorization: Bearer ${TOKEN}" \
  ${BASE_URL}/api/resumes/current

# 4. 岗位推荐（fast）
curl -s -o /dev/null -w "recommend: HTTP %{http_code}\n" \
  -X POST \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"resume_id":9,"generation_tier":"fast","limit":5}' \
  ${BASE_URL}/api/jobs/recommend
# 期望: HTTP 200

# 5. 岗位推荐（deep）—— 会触发 fallback 到 deep_fallback_agent
curl -s -o /dev/null -w "recommend_deep: HTTP %{http_code}\n" \
  -X POST \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"resume_id":9,"generation_tier":"deep","limit":5}' \
  ${BASE_URL}/api/jobs/recommend
# 期望: HTTP 200（非 500/503）

# 6. 岗位匹配分析
curl -s -o /dev/null -w "match: HTTP %{http_code}\n" \
  -X POST \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"resume_id":9,"generation_tier":"fast"}' \
  ${BASE_URL}/api/jobs/10001/match
# 期望: HTTP 200
```

---

## 十、回滚方式

如果导入后发现问题，使用备份文件恢复：

```bash
cd /opt/wenqing-hr-demo

# 1. 停止 server/web
docker compose --project-name wenqing-hr-demo --env-file .env.prod \
  -f docker-compose.prod.yml stop server web

# 2. 清空当前数据库
docker compose --project-name wenqing-hr-demo --env-file .env.prod \
  -f docker-compose.prod.yml exec -T mysql \
  mysql -uroot -p"${MYSQL_ROOT_PASSWORD}" \
  -e "DROP DATABASE IF EXISTS offer_catcher; CREATE DATABASE offer_catcher CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

# 3. 从备份恢复
BACKUP_FILE="backups/prod_backup_before_restore_YYYYMMDD_HHMMSS.sql.gz"
gunzip -c ${BACKUP_FILE} | \
  docker compose --project-name wenqing-hr-demo --env-file .env.prod \
  -f docker-compose.prod.yml exec -T mysql \
  mysql -uroot -p"${MYSQL_ROOT_PASSWORD}" --default-character-set=utf8mb4 offer_catcher

# 4. 重启 server/web
docker compose --project-name wenqing-hr-demo --env-file .env.prod \
  -f docker-compose.prod.yml up -d server web
```

> ⚠️ 不使用 docker volume 物理回滚。通过 SQL 备份/恢复完成回滚。

---

## 十一、注意事项

1. **执行前必须备份**：第二步的备份命令必须执行且确认成功。
2. **不要在业务高峰期执行**：全库替换会导致短暂服务不可用（< 1 分钟）。
3. **确认 Auth Token**：测试接口时需要有效的 JWT Token。
4. **Deep 推荐会 fallback**：导入后首次 deep 推荐可能返回 `deep_fallback_agent`（如 SEP 画像未生成），这是正常行为。
5. **密钥安全**：所有命令中的 `${MYSQL_ROOT_PASSWORD}` 从 `.env.prod` 读取，不要硬编码。
