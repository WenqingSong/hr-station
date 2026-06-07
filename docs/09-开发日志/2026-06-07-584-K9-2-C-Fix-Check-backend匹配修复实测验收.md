# K9-2-C-Fix-Check: backend 匹配修复实测验收

日期：2026-06-07 | 编写：AI Agent | 阶段：K9-2-C-Fix-Check 验收

## 1. K9-2-C-Fix 修复回顾

| 问题 | 修复 |
|------|------|
| `getProfile` 加载成功后被 fallback 覆盖 | `return cp` 立即返回 |
| `calcScore` backend 无方向 boost | `backend_engineering + strong dirF → +10` |
| `buildFitPointsV2` backend tech stack 识别不足 | 增强 Go/MySQL/Redis/Docker/部署 检测 |
| **K9-2-C-Fix-Check 新发现**: `extractJobTags` 将"软件开发-后台开发方向"误标为 `algorithm_core` | `calcScore` 中新增 `dirF` 覆盖逻辑 |
| **K9-2-C-Fix-Check 新发现**: backend 被 algorithm cap 误限 74 | conflict cap 新增 `isActuallyBackendJob` 例外 |

## 2. 实测结果

### 02 后端 + 后台

| 阶段 | Score | Level | gen_by | tier |
|------|:---:|------|------|------|
| Fast | **89** | 高匹配，建议重点投递 | mock_agent | fast |
| Quality | **89** | 高匹配，建议重点投递 | mock_quality | quality |

### 正向样例

| 样例 | Score | Level | 结论 |
|------|:---:|------|:--:|
| 01 算法 + NLP | 95 | 强匹配 | ✅ 未回退 |
| 02 后端 + 后台 | **89** | 高匹配 | ✅ 修复成功 |
| 03 产品 + 策划 | 75 | 可尝试 | ✅ 未回退 |
| 05 商务 + 方案 | 79 | 可尝试 | ✅ 未回退 |

### 错配样例

| 样例 | Score | Level | 结论 |
|------|:---:|------|:--:|
| 03 产品 + NLP | 69 | 备选参考 | ✅ 降级 |
| 05 商务 + NLP | 69 | 备选参考 | ✅ 降级 |
| 01 算法 + 秘书 | 55 | 不建议投递 | ✅ 降级 |

## 3. match_analyses 保存

ID=88: resume=2 job=10001 score=89 tier=**fast** by=mock_agent status=ready ✅  
ID=89: resume=2 job=10001 score=89 tier=**quality** by=mock_quality status=ready ✅

## 4. 构建检查

| 验收项 | 结果 |
|--------|:--:|
| go build | ✅ |
| go vet | ✅ |
| npm build | ✅ |
| check-encoding | ✅ OK |

## 5. 未修改范围

- ✅ 未修改前端源码
- ✅ 未修改 init.sql
- ✅ 未 reset-up / Git

## 6. 结论

✅ **K9-2-C-Fix-Check 通过，进入 K9-2-D：match latest API 与持久化读取实现**
