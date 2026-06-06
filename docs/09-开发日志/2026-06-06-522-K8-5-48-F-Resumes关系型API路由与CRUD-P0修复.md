# K8-5-48-F：Resumes 关系型 API 路由与 CRUD P0 修复

## 一、K8-5-48-E 发现的 P0 问题

| P0 | 原因 |
| -- | -- |
| POST /api/resumes 500 | CreateVersion 旧 V1/V2 字段映射 + FindByID shim 参数错误 + repository Create SQL 占位符数量不匹配 |
| GET /api/resumes 405 | 路由只注册 POST，未注册 GET |
| PUT /api/resumes/{id} 404 | 路由未实现 |
| POST .../set-current 404 | 路由未实现 |
| 响应为旧 V1/V2 格式 | handler 使用 ResumeVersionResponse |

## 二、修改清单

| 文件 | 修改 |
| -- | -- |
| `server/internal/model/resume.go` | 子结构体添加 JSON tags（educations/projects 等字段可正确序列化/反序列化） |
| `server/internal/service/resume_service.go` | 新增 ResumeRelationalRequest/ResumeResponse + 子响应类型；新增 List/GetCurrentRelational/GetByIDRelational/CreateRelational/UpdateRelational/SetCurrentRelational/Delete 方法 |
| `server/internal/http/resume_handler.go` | 新增 ResumesDispatcher（GET/POST）、ResumeByIDDispatcher（GET/PUT/DELETE/{id} + POST {id}/set-current）；旧 handler 保留为 @deprecated |
| `server/internal/http/router.go` | 注册 GET/POST /api/resumes；注册 /api/resumes/ 子路径调度；新增 deleted 死代码路由 |
| `server/internal/repository/resume_repository.go` | 修复 Create SQL 占位符 13→12 |

## 三、P0 修复结果

| P0 | 结果 |
| -- | -- |
| GET /api/resumes 405 | ✅ 200，返回关系型 ResumeResponse 列表 |
| POST /api/resumes 500 | ✅ 200，创建成功，含 raw_text |
| PUT /api/resumes/{id} 404 | ✅ 200，更新成功，子表重建 + raw_text |
| POST .../set-current 404 | ✅ 200，is_current 切换 |
| 响应为旧格式 | ✅ 新 API 返回关系型嵌套 DTO |

## 四、API 验收结果

| API | 结果 |
| --- | -- |
| GET /api/resumes | ✅ 200，返回关系型列表 |
| GET /api/resumes/current | ✅ 200（兼容旧格式） |
| GET /api/resumes/4 | ✅ 200，含 educations/projects |
| POST /api/resumes | ✅ 200，id=4，raw_text 生成 |
| PUT /api/resumes/4 | ✅ 200，更新 name+additional_info |
| POST /api/resumes/3/set-current | ✅ 200，is_current=True |

## 五、DB 验证

| 表 | 结果 |
| - | -- |
| resumes | 3 行（id=2,3,4），id=3 为 current |
| resume_educations | 2 行（id=2 和 id=4 各 1 行） |
| resume_projects | 2 行 |
| raw_text | 均非空 |

## 六、AI 四核心回归

| 功能 | 结果 |
| -- | -- |
| 推荐 resume_id=4 | ✅ 200, mock_agent |
| 匹配 resume_id=4 | ✅ 200, mock_agent |
| 优化 resume_id=4 | ✅ 200, mock_agent |
| 对比 from_id=2&to_id=4 | ✅ 200, mock_agent |

## 七、构建 / 测试

| 验收项 | 结果 |
| --- | -- |
| go build | ✅ |
| go vet | ✅ |
| go test | ✅ 全部通过 |
| npm run build | ✅ |

## 八、旧路由保留状态

| 路由 | 状态 |
| -- | -- |
| GET /api/resumes/versions | @deprecated，保留 |
| GET /api/resumes/versions/{id} | @deprecated，保留 |
| POST /api/resumes (旧 handler) | 已替换为新 dispatcher |
| GET /api/resumes/detail | @deprecated，保留 |

## 九、未修改范围

```
未修改前端源码
未修改 SQL / seed
未修改数据库结构
未修改 .env
未修改 Docker / Nginx / Compose
未安装依赖
未迁移目录
未执行 make reset-up
未导入旧示范简历
未执行 Git 命令
```
