# Resumes API 联调：当前简历 / 简历版本

> 阶段：6-2 | 状态：已完成 | 日期：2026-06-01

## 1. 本阶段目标

打通 Resumes 普通业务读接口联调，确认前后端通路可用。

## 2. 环境上下文恢复结果

| 项目 | 值 |
|------|-----|
| 分支 | `main` |
| HEAD | `4a98f83` |
| 工作区 | clean（联调前） |
| Go | 1.24.1 |
| Node | v22.22.2 |
| npm | 10.9.7 |
| server | 存在（完整分层） |
| docs/08-后端联调 | 存在 |
| 6-0 文档 | ✅ 存在 |
| 6-1 文档 | ✅ 存在 |

## 3. 后端启动方式

```bash
make up
```

启动后：MySQL + backend:8080 + frontend:5173。

## 4. 登录 / token 获取

```
POST /api/auth/register
Body: { "email": "stage6-resume@example.com", "password": "123456", "name": "Stage6 Resume" }
→ { code: 0, data: { token: "eyJ...", user: {...} } }

POST /api/auth/login
Body: { "email": "arthur@example.com", "password": "123456" }
→ { code: 0, data: { token: "eyJ...", user: { id: 1, ... } } }
```

JWT `Authorization: Bearer <token>` 用于所有业务请求。seed 用户 `arthur@example.com` 有 2 条简历版本（V1, V2）。

## 5. Resumes API curl 联调结果

### 5.1 结果矩阵

| 接口 | HTTP | code | message | data 结构 | 结论 |
|------|:--:|:--:|------|------|:--:|
| `GET /api/resumes/current` | **200** | 0 | ok | `ResumeVersionResponse` (V2, 9技能, 2项目) | ✅ pass |
| `GET /api/resumes/detail` | **200** | 0 | ok | 同 current | ✅ pass |
| `GET /api/resumes/versions` | **200** | 0 | ok | `{ versions: [...], total: 2 }` | ✅ pass |
| `GET /api/resumes/versions/1` | **200** | 0 | ok | `ResumeVersionResponse` (V1) | ✅ pass |
| `GET /api/resumes/versions/2` | **200** | 0 | ok | `ResumeVersionResponse` (V2) | ✅ pass |

### 5.2 compare 接口边界确认

| 接口 | HTTP | code | message | 结论 |
|------|:--:|:--:|------|:--:|
| `GET /api/resumes/versions/compare?from=1&to=2` | **200** | 0 | ok | ✅ 边界确认 |

compare 接口当前由 `VersionCompareHandler` 处理，返回 `mock_agent` 生成的版本对比数据（字段变化、风险变化、建议完成度等）。此为普通后端逻辑（非真实 LLM 调用），属于阶段 6 可接入范围。如后续改为 Agent/LLM 生成分析，则为阶段 7。

## 6. 前端 Resumes API 路径对齐

### 6.1 路径检查

| 前端函数 | 请求路径 | 后端路径 | 对齐 |
|------|------|------|:--:|
| `getCurrentResume()` | `GET /resumes/current` | `GET /api/resumes/current` | ✅ |
| `getResumeDetail(id)` | `GET /resumes/detail?id={id}` | `GET /api/resumes/detail` | ⚠️ |
| `getResumes()` | `GET /resumes/versions` | `GET /api/resumes/versions` | ✅ |
| `getResumeOptions()` | `GET /resumes/versions` | `GET /api/resumes/versions` | ✅ |
| `compareResumes()` | `GET /resumes/versions/compare?from={}&to={}` | `GET /api/resumes/versions/compare?from={}&to={}` | ✅ 已修复 |

### 6.2 已修复

- **`compareResumes` 参数名**：从 `base_id`/`target_id` 改为 `from`/`to`，对齐后端契约。

### 6.3 已知不匹配（文档记录，不修改后端）

| 项目 | 前端 | 后端 | 影响 |
|------|------|------|------|
| `getResumeDetail(id)` | 查询参数 `?id=` | 后端忽略，始终返回 current | mock fallback 覆盖 |
| `getResumes()` 返回类型 | `Resume[]` | `{ versions: [...], total: N }` | 结构不匹配，mock fallback 覆盖 |
| `getResumeOptions()` 返回类型 | `ResumeOption[]` | `{ versions: [...], total: N }` | 同上 |
| compare 参数值 | resumeId（字符串） | versionNo（整数） | mock fallback 覆盖 |
| POST `/resumes` | 前端有 | 后端有 `POST /api/resumes` | 写操作，留到 6-3 |
| PUT `/resumes/detail` | 前端有 `updateResume` | **后端无此接口** | P1，留到 6-3 确认 |
| PATCH `/resumes/current` | 前端有 `setCurrentResume` | **后端无此接口** | P1，留到 6-3 确认 |

## 7. Authorization header 支持情况

- `client.ts` 已支持 `options.headers` ✅
- `requestRaw()` 可传入 `Authorization: Bearer <token>` ✅
- Resumes API 函数未自动注入 token，页面级 token 注入留到用户态阶段 ✅
- 不写入 localStorage/sessionStorage ✅

## 8. Resume 字段映射结果

### 8.1 后端返回字段（curl 确认）

```json
{
  "id": 2,
  "version_no": 2,
  "name": "Arthur",
  "phone": "13912345678",
  "email": "arthur@example.com",
  "location": "深圳",
  "target_city": "深圳",
  "target_role": "后端开发工程师",
  "education": [{"school":"...","major":"...","degree":"...","start_date":"...","end_date":"..."}],
  "experience": [{"company":"...","position":"...","start_date":"...","end_date":"...","description":"..."}],
  "project": [{"name":"...","role":"...","start_date":"...","end_date":"...","description":"..."}],
  "skills": ["Go","MySQL",...],
  "self_evaluation": "...",
  "raw_text": "",
  "created_at": "2026-06-01T08:02:29+08:00"
}
```

**关键发现**：JSON 字段（education/experience/project/skills）后端使用 `json.RawMessage`，在 HTTP 响应中已作为**内联 JSON 数组**序列化，前端直接解析为 `any[]`，**不需要二次 JSON.parse**。

### 8.2 字段对比

| 后端字段 | 前端 Resume 类型 | 匹配 |
|------|------|:--:|
| id (uint64) | id (string) | ⚠️ 类型 |
| version_no (int) | version (string) | ⚠️ 名+类型 |
| name | name | ✅ |
| phone | profile.phone | ❌ 结构 |
| email | profile.email | ❌ 结构 |
| location | — | ⚠️ 前端无 |
| target_city | expected_cities[] | ❌ 结构 |
| target_role | direction/direction_label | ❌ 映射 |
| education (JSON[]) | education[] (ResumeEducation) | ⚠️ 字段名 |
| experience (JSON[]) | experiences[] (ResumeExperience) | ⚠️ 字段名 |
| project (JSON[]) | projects[] (ResumeProject) | ⚠️ 字段名 |
| skills (JSON[]) | skills (string[]) | ✅ |
| self_evaluation | self_evaluation | ✅ |
| created_at | updated_at | ⚠️ 名 |
| — | status/completeness/gaps/profile | ❌ 前端独有 |
| raw_text | — | ⚠️ 前端无 |

### 8.3 映射结论

需要 adapter 层做以下转换：
1. `version_no` → `version` + 格式化为字符串
2. `target_role` → `direction` / `direction_label`
3. `target_city` → `expected_cities[]`
4. `phone`, `email`, `education[0].school` → `profile`
5. `created_at` → `updated_at`
6. `experience` → `experiences` (数组字段名映射)
7. `project` → `projects` (数组字段名映射，含 tech_stack→string[])
8. `/api/resumes/versions` 的 `{ versions: [...] }`需解包为 `Resume[]`

此适配留到阶段 6-4（页面数据源切换）或通过 viewModel loader 处理。

## 9. JSON 字段解析风险

| 风险 | 评级 | 说明 |
|------|:--:|------|
| JSON 字段二次解析 | **无风险** | 后端 `json.RawMessage` 响应中已作为 JSON 内联数组，前端 `fetch → JSON.parse` 后直接是数组 ✅ |
| JSON 字段类型安全 | P2 | TypeScript 类型 `json.RawMessage` = `any[]`，无编译时类型保护 |
| 空 JSON 字段 | 无风险 | 后端 `normalizeResponseJSON` 对空值返回 `[]` |

## 10. mock fallback 保留情况

- `withFallback` 机制完整保留 ✅
- 后端成功（200 + code:0）且数据兼容时可用真实 API ✅
- 后端失败 / 结构不匹配 / 网络错误时降级 mock ✅
- 不导致页面白屏 ✅

## 11. build 结果

```
✓ built in 1.54s
```

**通过 ✅**

## 12. compare 接口边界说明

- `GET /api/resumes/versions/compare?from=1&to=2` 当前返回 `mock_agent` 生成的对比数据
- 包含：`field_changes`、`risk_changes`、`suggestion_completion`、`summary` 等
- 这是普通后端业务逻辑（非真实 LLM 调用）
- 可归类为阶段 6 可接入范围
- 如后续改为 Agent/LLM 生成分析，归入阶段 7

## 13. P0 / P1 / P2 问题

| 等级 | 数量 | 说明 |
|:--:|:--:|------|
| P0 | 0 | — |
| P1 | 3 | (1) 前端 `getResumeDetail` 查询参数被后端忽略 (2) PUT/PATCH 写接口后端不存在（6-3确认） (3) 版本列表响应结构 `{versions:[],total}` 与前端 `Resume[]` 不匹配 |
| P2 | 2 | (1) 字段名/类型映射需 adapter 处理 (2) compare 参数值从 ID 到版本号需转换 |

## 14. 越界检查

| 检查项 | 结果 |
|------|:--:|
| 是否修改 server | ❌ 零修改 |
| 是否修改 Makefile/docker-compose/nginx/.env | ❌ 零修改 |
| 是否新增 axios | ❌ 无 |
| 是否新增 /api/v1 | ❌ 无 |
| 是否新增 selectors | ❌ 无 |
| 是否新增 localStorage/sessionStorage | ❌ 无 |
| 是否执行 npm install | ❌ 未执行 |
| 是否调用真实 LLM | ❌ 未调用 |
| 是否清空 volume | ❌ 未清空 |

## 15. 本阶段未做事项

| # | 未做项 | 说明 |
|:--:|------|------|
| 1 | 简历新增/修改/设置当前简历 | 留到 6-3 |
| 2 | 页面整体数据源切换 | 留到 6-4 |
| 3 | 修改后端接口 | 禁止 |
| 4 | 修改数据库结构 | 禁止 |
| 5 | 新增登录页面 | 禁止 |
| 6 | 写入 localStorage | 禁止 |
| 7 | 调用真实 LLM | 禁止 |

## 16. 是否可以进入 6-3

**可以。** Resumes 读 API 联调成功（4 个核心读接口全部 200 OK），前后端通路验证通过，路径对齐，token 获取方式明确。P1/P2 问题已记录，不阻断 6-3。

## 17. 最终结论

**pass ✅**

- 4 个 Resumes 读接口全部 200 + code:0 ✅
- token 获取方式明确（register/login） ✅
- 前端路径已对齐（compare 参数名已修复） ✅
- Authorization header 支持明确 ✅
- mock fallback 保留 ✅
- npm run build 1.54s 通过 ✅
- 零越界修改 ✅
- 可进入 6-3 ✅
