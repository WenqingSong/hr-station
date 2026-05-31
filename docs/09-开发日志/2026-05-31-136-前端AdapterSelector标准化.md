# 5-1.3 Adapter / Selector 标准化

## 任务编号

5-1.3

## 任务名称

Adapter / Selector 标准化

## 修改原因

5-1.1 types + 5-1.2 mocks 已就绪，需要新增统一的 Adapter 层将原始数据转换为页面 ViewModel，为 5-1.4 页面改造提供数据转换基础。

## 修改文件

| 文件 | 操作 | 说明 |
|------|------|------|
| `web/src/adapters/common.ts` | 新增 | 通用工具函数（tone/label 映射、agent 来源标签） |
| `web/src/adapters/agentAdapter.ts` | 新增 | Agent 公共字段转换（meta/TopFix/Evidence） |
| `web/src/adapters/jobAdapter.ts` | 新增 | 岗位数据转换（Job → JobCardViewModel） |
| `web/src/adapters/resumeAdapter.ts` | 新增 | 简历数据转换（Resume/ResumeOption/Comparison） |
| `web/src/adapters/recommendationAdapter.ts` | 新增 | 岗位推荐转换（RecommendationResult → PageViewModel） |
| `web/src/adapters/matchAnalysisAdapter.ts` | 新增 | 匹配分析转换（MatchAnalysisResult → PageViewModel） |
| `web/src/adapters/resumeOptimizationAdapter.ts` | 新增 | 简历优化转换（ResumeOptimizationResult → PageViewModel） |
| `web/src/adapters/index.ts` | 新增 | 统一导出 |

## 未修改范围

- web/src/pages/ — 未修改
- web/src/components/ — 未修改
- web/src/types/ — 未修改
- web/src/mocks/ — 未修改
- web/src/api/ — 未创建
- server/ — 未修改

## 自检结果

- ✅ 仅新增 web/src/adapters/ + docs
- ✅ 未修改页面组件
- ✅ 未新增 api client（无 fetch/axios）
- ✅ 未修改后端
- ✅ 未修改 types / mocks
- ✅ 未启动项目
- ✅ 未调用真实 LLM
- ✅ 未清空 volume
- ✅ Adapter 覆盖 9 页对应关系
- ✅ 所有函数为纯函数，类型安全，无 any

## 下一步建议

进入 **5-1.4 页面硬编码数据替换为统一 mocks**。
