# K8-5-48-D：AI 四核心 resume_id / raw_text 链路适配

## 一、本阶段目标

彻底适配四核心 AI 功能（推荐/匹配/优化/对比），统一从 `resume_version_no` 改为 `resume_id`/`from_id`/`to_id`，统一使用 `RawText`。

## 二、K8-5-48-B/C 当前状态

K8-5-48-B：后端 resumes 主表 + 7 子表已完成，repository 支持关系型 CRUD + RawText。
K8-5-48-C：前端 types/api/adapter/pages 已切换为关系型嵌套 DTO，versionNo→resumeId。

## 三、修改文件

### 后端

| 文件 | 修改 |
| -- | -- |
| `server/internal/service/resume_optimization_service.go` | Request 新增 resume_id/job_id；Service 新增 resumeRepo/jobRepo；Optimize 根据 resume_id 读取 RawText + 根据 job_id 读取 Job |
| `server/internal/service/recommendation_service.go` | generateSummary 改为使用 resume.Name 替代 TargetRole |
| `server/internal/http/version_compare_handler.go` | Query 参数 from→from_id, to→to_id |
| `server/internal/http/router.go` | 新增 /api/resumes/compare 路由 |
| `server/cmd/server/main.go` | NewResumeOptimizationService 传入 resumeRepo, jobRepo |
| `server/internal/service/resume_optimization_service_test.go` | 适配新构造函数签名（17 处） |
| `server/internal/agent/match_agent_response_smoke_test.go` | 补充 ExperienceJSON/SelfEvaluation/TargetRole/TargetCity/VersionNo 到 struct literal |
| `server/internal/agent/match_agent_test.go` | 补充 ExperienceJSON |
| `server/internal/agent/recommendation_agent_test.go` | 补充 ExperienceJSON |
| `server/internal/agent/agent_pipeline_integration_test.go` | 补充 ExperienceJSON |

### 前端

| 文件 | 修改 |
| -- | -- |
| `web/src/api/agents.ts` | resume_version_no→resume_id；优化 resume_text→resume_id+job_id |
| `web/src/app/pages/ResumeOptimizationPage.tsx` | 移除 buildResumeText/buildTargetJob；优化请求传递 resume_id+job_id |

## 四、四核心 API 适配结果

| 功能 | 旧参数 | 新参数 | AI 输入 |
| -- | --- | --- | ----- |
| 岗位推荐 | resume_version_no | resume_id | 后端根据 resume_id 读取 Resume→Agent |
| 岗位匹配 | resume_version_no | resume_id | 后端根据 resume_id 读取 Resume→Agent |
| 简历优化 | resume_text + target_job | resume_id + job_id | 后端根据 resume_id 读取 RawText，根据 job_id 读取 Job |
| 简历对比 | from + to (version_no) | from_id + to_id (resume ID) | 后端根据 from_id/to_id 读取两份 Resume→Agent |

## 五、后端链路调整结果

| 层级 | 修改 |
| -- | -- |
| Handler (recommend) | 已支持 resume_id（无变更） |
| Handler (match) | 已支持 resume_id（无变更） |
| Handler (version_compare) | from→from_id, to→to_id |
| Handler (resume_optimize) | 通过 service 层解析 resume_id/job_id |
| Service (recommendation) | generateSummary 不再使用 TargetRole |
| Service (optimization) | 新增 resumeRepo/jobRepo，Optimize 按 resume_id 取 RawText |
| Router | 新增 /api/resumes/compare 路由 |

## 六、前端链路调整结果

| 页面 / 文件 | 修改 |
| ------- | -- |
| agents.ts | resume_version_no→resume_id；OptimizationRequest 新增 resume_id/job_id |
| ResumeOptimizationPage.tsx | 传递 resume_id+job_id，移除 buildResumeText/buildTargetJob |

## 七、raw_text / Snapshot 策略

| 场景 | 处理 |
| -- | -- |
| 推荐 Agent | agent_snapshot.go 归一化 ResumeVersion→ResumeSnapshotV2（使用 shim 字段暂时） |
| 匹配 Agent | agent_snapshot.go 归一化（同上） |
| 优化 Agent | service 层从 Resume.RawText 获取，空时由 repository buildRawText 生成 |
| 对比 Agent | agent_snapshot / mock 使用 Resume.RawText + shim 字段 |

## 八、compat shim 清理结果

| shim | 处理 |
| ---- | -- |
| resume_version_no（推荐/匹配） | handler 已移除，service 已使用 resume_id |
| resume_text（优化，主输入） | 已转为 resume_id 主输入，resume_text 保留为 optional fallback |
| from/to（对比） | 已改为 from_id/to_id |
| TargetRole（generateSummary） | 已改为 resume.Name |
| model.Resume.TargetRole/City/SelfEvaluation | 保留为 shim（agent snapshot normalizer 仍使用） |
| model.Resume.EducationJSON/ExperienceJSON 等 | 保留为 shim（test/resume_service 仍使用） |

## 九、旧字段静态检查结果

### 后端

| 检查项 | 结果 |
| --- | -- |
| AI 四核心 service 中 resume_version_no | ✅ 已清除 |
| AI 四核心 handler 中 resume_version_no | ✅ 已清除 |
| version_compare handler from/to | ✅ 已改为 from_id/to_id |
| resume_optimization validate 依赖 resume_text | ✅ 已支持 resume_id fallback |
| agent_snapshot.go SelfEvaluation/VersionNo | ⚠️ 仍使用 model shim 字段（暂时保留） |
| resume_service.go V2 JSON 字段 | ⚠️ 非 AI 核心，保留 |

### 前端

| 检查项 | 结果 |
| --- | -- |
| agents.ts resume_version_no | ✅ 已改为 resume_id |
| agents.ts resume_text（OptimizationRequest） | ⚠️ 保留为 optional deprecated |
| 所有 AI 页面 resumeVersionNo | ✅ 已清除 |
| redirect.ts isValidVersionNo | ⚠️ deprecated wrapper |

## 十、构建 / 测试 / API 结果

| 验收项 | 结果 |
| --- | -- |
| `go build ./...` | ✅ 通过 |
| `go vet ./...` | ✅ 通过 |
| `go test ./...` | ✅ 全部通过 |
| `npm run build` | ✅ 通过 |

## 十一、未修改范围确认

```
未修改 SQL / seed
未修改数据库数据
未修改 .env
未修改 Docker / Nginx / Compose
未安装依赖
未迁移目录
未新增上传能力
未执行 make reset-up
未写入数据库
未手工 INSERT 数据
```

## 十二、未执行

```
未执行 make reset-up
未执行任何 Git 命令
```
