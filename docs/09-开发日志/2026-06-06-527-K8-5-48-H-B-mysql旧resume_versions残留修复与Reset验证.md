# K8-5-48-H-B：mysql.sh 旧 resume_versions 残留修复与 Reset 验证

## 一、H-A 诊断结论

`make reset-up` 失败因 `scripts/dev/mysql.sh:115` 引用未定义变量 `RESUME_VERSIONS_SQL`。

## 二、修复

| 文件 | 修改 |
| -- | -- |
| `scripts/dev/mysql.sh` | 删除 import action 中 RESUME_VERSIONS_SQL 块 (115-119行) |
| `scripts/dev/mysql.sh` | 删除 reset action 中 RESUME_VERSIONS_SQL 块 (142-146行) |
| `server/sql/init.sql` | 陈旧注释：resume_versions.sql→不再自动导入 |
| `server/sql/init.sql` | TABLE COMMENT：推荐结果表→岗位推荐结果表 |
| `server/sql/init.sql` | TABLE COMMENT：匹配分析结果表→岗位匹配分析结果表 |

## 三、reset-up 验证

| 步骤 | 结果 |
| -- | -- |
| init.sql 导入 | ✅ |
| 腾讯岗位导入 | ✅ 46条 |
| 后端启动 | ✅ PID 104943 |
| 健康检查 | ✅ HTTP 200 |
| 前端启动 | ✅ PID 105328 |
| 最终检查 | ✅ 全部通过 |

## 四、DB 验证

| 项 | 结果 |
| -- | -- |
| resumes + 7 子表 | ✅ 全部存在 |
| resume_versions 表 | ✅ 不存在 |
| TABLE COMMENT | ✅ 全部正确中文 |
| users | 2 ✅ |
| jobs | 46 ✅ |

## 五、构建 / API

| 项 | 结果 |
| -- | -- |
| npm run build | ✅ |
| go build/vet | ✅ |
| check-encoding | ✅ OK |
| healthz | ✅ |
| /api/jobs | ✅ 46 jobs |
| /api/resumes | ✅ 401 (normal) |

## 六、未恢复旧设计

```
未恢复 resume_versions.sql 导入 ✅
未恢复 resume_versions 表 ✅
未恢复 versionNo ✅
未恢复 JSON 大表 ✅
未导入旧示范简历 ✅
```
