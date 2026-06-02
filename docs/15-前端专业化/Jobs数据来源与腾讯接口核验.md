# Jobs 数据来源与腾讯招聘接口核验

## 1. 核验结论

**当前所有 8 个岗位数据 100% 来自 `server/sql/init.sql` 手工编写的 seed 数据，未调用腾讯招聘接口。** 数据标记为 `external_source=local_mock, data_quality=manual_enriched`。

## 2. 是否调用腾讯招聘接口

**否。** 后端的 `JobRepository` 仅读取本地 MySQL 数据库，无任何外部 HTTP 调用。

腾讯招聘拉取/入库脚本已编写完成但从未执行：
- `scripts/dev/tencent_jobs.sh` — 拉取原始 JSON 到本地
- `scripts/dev/tencent_jobs_import.sh` — 清洗并生成 INSERT SQL

两脚本均存在且有 Makefile 快捷命令，但 `data/tencent-jobs/` 目录不存在，证明从未运行。

## 3. 数据来源明细

| 岗位ID | 标题 | 来源 | 可信度 |
|--------|------|------|--------|
| 1 | Go 后端开发工程师 | init.sql | **C** — 手工编写 |
| 2 | 云计算研发工程师 | init.sql | **C** — 手工编写 |
| 3 | AI 应用开发工程师 | init.sql | **C** — 手工编写 |
| 4 | 数据开发工程师 | init.sql | **C** — 手工编写 |
| 5 | 测试开发工程师 | init.sql | **C** — 手工编写 |
| 6 | 产品经理 | init.sql | **C** — 手工编写 |
| 7 | 前端开发工程师 | init.sql | **C** — 手工编写 |
| 8 | 算法工程师 | init.sql | **C** — 手工编写 |

**全部 `external_source=local_mock, data_quality=manual_enriched`。**

## 4. 数据库状态

```sql
-- MySQL 直接查询
SELECT id,title,external_source,data_quality FROM jobs;
-- 8 rows: all local_mock / manual_enriched
```

无 `tencent_join` 记录。

## 5. facets 核验

```
directions:  软件开发(4), 算法与大数据(3), 产品策划与运营(1)  → 合计 8
cities:      深圳(3), 北京(2), 上海(1), 广州(1), 杭州(1)     → 合计 8
recruit_types: 校招(8)                                         → 合计 8
categories:  8 个类别各 1 条                                     → 合计 8
```

facets count 完全基于 init.sql seed 数据聚合，无外部来源。

## 6. 腾讯接口配置概览

| 配置项 | 值 |
|--------|-----|
| 列表接口 | `https://join.qq.com/api/v1/position/searchPosition` |
| 详情接口 | `https://join.qq.com/api/v1/jobDetails/getJobDetailsByPostId` |
| 字典接口 | `https://join.qq.com/api/v1/dictionary/` |
| 是否需要 Token/Cookie | ❓ 未验证 |
| 是否公开接口 | ❓ 配置中使用 `join.qq.com` 公开页面 URL |

## 7. 前端核验

- `web/src/` 中无 `mock` 目录或文件
- 前端 `JobListPage` / `JobDetailPage` 调用真实 API
- `external_source` 字段已在前端类型定义中但无需展示
- 无 mock fallback 逻辑

## 8. 可信度评级

**C — 不可信假数据。** 原因：
- 岗位内容为手工编写，非腾讯官方数据
- 岗位 JD/要求为通用模板风格
- `local_mock` 标签明确标识为演示数据
- facets 基于 seed 数据聚合，非真实招聘统计

## 9. 建议

### 方案 A：手动维护演示岗位（推荐）

放弃实时调用腾讯招聘接口。人工维护一份演示岗位样本库：
- 岗位明确标注为"项目演示样本"
- 字段完整可验证
- 可通过管理界面或 SQL 编辑

### 方案 B：运行已有腾讯同步脚本

执行 `make tencent-fetch-all` → `make tencent-import` 拉取真实岗位。风险：
- `join.qq.com` 接口稳定性和访问限制未知
- 需验证字段映射
- 正式上线可能有合规风险

## 10. P0/P1/P2

| 级别 | 问题 |
|------|------|
| P2 | 岗位数据为 seed/mock，非真实腾讯岗位 |
| P2 | 腾讯同步脚本存在但从未运行 |
| P2 | 无岗位管理后台 |

## 11. 修改清单

**本阶段未修改任何业务代码，仅新增本文档。**

## 12. 下一步建议

- 如需演示效果：当前 8 条 seed 岗位可继续使用，建议在页面标注"演示岗位样本"
- 如需真实数据：运行腾讯同步脚本或改为手动录入
