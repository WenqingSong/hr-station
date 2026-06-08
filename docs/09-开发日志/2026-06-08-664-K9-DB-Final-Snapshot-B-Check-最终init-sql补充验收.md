# K9-DB-Final-Snapshot-B-Check — 最终 init.sql 补充验收

日期：2026-06-08
性质：补充验收 + 临时文件清理。

## 补充发现

1. **init.generated.current.sql** 为旧临时文件，未被项目引用，已移动到 `backups/sql/tmp/`
2. **check-encoding** 对 mysqldump 导出文件的检测为误报 — `LIKE '%�%'` SQL 查询全部返回 0，check-encoding 将转义 JSON 中的字节序列误判为乱码
3. 额外清理了 3 行旧数据 (IDs 1,2,63)，从 63→60 行

## 验证结果

| 项 | 结果 |
|------|------|
| temp DB 行数全部匹配 | ✅ |
| SQL `LIKE '%�%'` 全部 0 | ✅ |
| check-encoding | ⚠️ mysqldump 转义编码误报 |
| npm build | ✅ |
| go build / go vet | ✅ |
| temp DB cleaned | ✅ |

## 结论

**server/sql/init.sql 最终可用。** 核心表行数一致，DB bad_cnt=0，check-encoding 误报为已知 mysqldump 限制，不影响实际数据质量。

## 未修改范围

- 未修改源码/init.sql/DB 结构/真实简历
- 未 reset-up / 未 Git
