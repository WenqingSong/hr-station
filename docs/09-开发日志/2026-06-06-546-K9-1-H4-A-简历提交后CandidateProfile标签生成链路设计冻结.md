# K9-1-H4-A：简历提交后 CandidateProfile 标签生成链路设计冻结

> 日期：2026-06-06 | 状态：设计冻结

## 1. 结论

本阶段完成 CandidateProfile 标签生成全链路设计冻结，包括 ResumeProfile 表结构、Fast/Quality 分型、Profile 状态机、简历保存触发机制、异步生成与防重、Backfill 补偿、推荐接口适配、前端 loading 行为。

## 2. 核心设计

| 项 | 冻结内容 |
| -- | -- |
| Profile 实体 | resume_profiles 表，独立派生表，不混入 resumes 主表 |
| 分型 | Fast(规则同步) / Quality(Agent异步) |
| 状态机 | missing→pending→ready/failed→stale（5 态） |
| 触发 | 简历 Create/Update 后立即生成 |
| 防重 | UNIQUE(resume_id,profile_type) + status=pending 判断 |
| Backfill | 后端启动时扫描缺失，异步补齐 |
| 推荐接口 | 新增 profile_status 字段，非 ready 不返回推荐结果 |
| 前端 | profile_generating → 全局 loading + 轮询(1.5-2s/15s超时) |

## 3. 后续阶段

| 阶段 | 范围 |
| -- | -- |
| K9-1-H4-B | ResumeProfile 存储与生成后端实现 |
| K9-1-H4-C | 推荐链路改为消费 ResumeProfile |
| K9-1-I | 前端 Profile 状态接入 |

## 4. 约束

- 未修改业务源码
- 未修改 SQL / DB
- encoding ✅
- 未执行 reset-up / Git
