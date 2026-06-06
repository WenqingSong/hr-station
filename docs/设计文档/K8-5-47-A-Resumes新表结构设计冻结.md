# K8-5-47-A：Resumes 新表结构设计冻结

> 状态：✅ 已完成 | 日期：2026-06-06 | 性质：设计冻结，未修改源码

---

## 一、当前问题

当前 `resumes` 表（K8-5-51-B 从 `resume_versions` 改名）继承了 V1/V2 混合设计的全部历史包袱：

- 24 个列，其中 10 个为旧摘要列/兼容列
- 同时存在 `experience_json` + `internship_json` 两套字段
- `phone`/`email`/`location` 同时存在于根级列和 `profile_json` 中
- `job_intention_json` 与 `target_direction`/`job_type`/`expected_cities` 摘要列冗余
- `self_evaluation` 独立列设计已被 K8-5-44 废弃

**设计原则：不再兼容旧字段，不迁移旧 seed。**

---

## 二、新 resumes 表 DDL

```sql
CREATE TABLE `resumes` (
  `id`              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id`         BIGINT UNSIGNED NOT NULL,

  `schema_version`  SMALLINT         NOT NULL DEFAULT 3   COMMENT '简历结构版本：3=精简模块化',
  `name`            VARCHAR(100)     NOT NULL DEFAULT ''   COMMENT '简历名称',
  `is_current`      TINYINT(1)       NOT NULL DEFAULT 0    COMMENT '是否当前默认简历',

  `profile_json`       JSON NOT NULL COMMENT '基础信息',
  `education_json`     JSON NOT NULL COMMENT '教育经历',
  `project_json`       JSON NOT NULL COMMENT '项目经历',
  `skills_json`        JSON NOT NULL COMMENT '技能信息',
  `internship_json`    JSON NOT NULL COMMENT '实习经历',
  `certificates_json`  JSON NOT NULL COMMENT '外语考试与证书',
  `awards_json`        JSON NOT NULL COMMENT '获奖经历',
  `portfolio_json`     JSON NOT NULL COMMENT '作品链接',
  `extra_json`         JSON NOT NULL COMMENT '补充信息',

  `raw_text`        MEDIUMTEXT       NULL       COMMENT 'AI 文本快照',

  `created_at`      DATETIME         NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`      DATETIME         NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

  PRIMARY KEY (`id`),
  KEY `idx_resumes_user_current` (`user_id`, `is_current`, `updated_at`),
  KEY `idx_resumes_user_updated`  (`user_id`, `updated_at`),
  CONSTRAINT `fk_resumes_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='简历表';
```

**设计要点：**
- `is_current` 替代 `version_no`，每个 user 只有一份当前简历
- JSON 列为 `NOT NULL`，后端创建时写入空对象/空数组
- `raw_text` 由后端从模块 JSON 生成，供 AI 使用
- schema_version=3 标记新版结构

---

## 三、JSON 模块字段契约

| 模块 | 结构 | 字段 |
|------|------|------|
| `profile_json` | Object | name, phone, email, city |
| `education_json` | Array | school, degree, major, department, start_date, end_date, gpa, gpa_scale, rank, core_courses, research_direction, lab, paper_summary |
| `project_json` | Array | name, role, start_date, end_date, description |
| `skills_json` | Object | professional_skills (string), ai_application (string) |
| `internship_json` | Array | company, position, start_date, end_date, description |
| `certificates_json` | Object | language_exams[{type, score}], certificates[{name, description}] |
| `awards_json` | Array | name, description |
| `portfolio_json` | Object | works[{name, url, password}], homepage |
| `extra_json` | Object | additional_info |

---

## 四、废弃字段/表清单

| 废弃项 | 类型 | 原因 |
|--------|------|------|
| `phone` | 根级列 | 已迁移到 profile_json.phone |
| `email` | 根级列 | 已迁移到 profile_json.email |
| `location` | 根级列 | 已迁移到 profile_json.city |
| `target_city` | 根级列 | 不再单独维护 |
| `target_role` | 根级列 | 不再单独维护 |
| `target_direction` | 根级列 | 不再单独维护 |
| `job_type` | 根级列 | 不再单独维护 |
| `expected_cities` | 根级列 | 不再单独维护 |
| `school` | 摘要列 | 教育经历 JSON 已有 |
| `degree` | 摘要列 | 教育经历 JSON 已有 |
| `major` | 摘要列 | 教育经历 JSON 已有 |
| `graduation_year` | 摘要列 | 教育经历 JSON 已有 |
| `job_intention_json` | JSON 列 | 不再需要 |
| `experience_json` | JSON 列 | 统一为 internship_json |
| `self_evaluation` | 文本列 | 合并到 extra_json.additional_info |
| `raw_text`（旧） | TEXT 列 | 改为 MEDIUMTEXT |
| `version_no` | 设计概念 | 不再使用 |
| `resume_versions` 表 | 表 | 不再存在 |

---

## 五、API 设计方向

| API | 用途 |
|-----|------|
| `GET /api/resumes` | 列出当前用户所有简历 |
| `GET /api/resumes/current` | 获取当前默认简历 |
| `GET /api/resumes/{id}` | 获取指定简历详情 |
| `POST /api/resumes` | 新增简历 |
| `PUT /api/resumes/{id}` | 更新指定简历 |
| `DELETE /api/resumes/{id}` | 删除指定简历 |
| `POST /api/resumes/{id}/set-current` | 设为当前默认简历 |
| `GET /api/resumes/compare?from_id=X&to_id=Y` | 对比两份简历 |

不再使用 `versionNo` 参数。

---

## 六、后端影响面

| 层级 | 后续需要改 |
|------|-----------|
| `model/resume.go` | 重写 Resume struct，移除旧字段 |
| `model/resume_snapshot.go` | 适配新 JSON 字段名 |
| `repository/resume_repository.go` | SQL 全量重写 |
| `service/resume_service.go` | 用 is_current 替代 version_no，适配新 JSON |
| `agent/agent_snapshot.go` | 适配新模块 JSON 提取逻辑 |
| `agent/*mock*.go` | 测试数据适配 |
| `prompts/*.md` | Prompt 中 resume 格式描述更新 |

---

## 七、前端影响面

| 层级 | 后续需要改 |
|------|-----------|
| `types/resume.ts` | 重写类型定义 |
| `api/resumes.ts` | API 路径和参数对齐 |
| `adapters/resumeAdapter.ts` | 适配新 response 结构 |
| `MyResumesPage.tsx` | 保存 payload 对齐新 contract |
| `ResumeManagementPage.tsx` | 列表读取适配 |

---

## 八、AI 四核心影响面

| 功能 | 新输入方式 |
|------|-----------|
| 岗位智能推荐 | `resume_id` → 后端读模块 JSON / raw_text → LLM |
| 岗位匹配分析 | `job_id` + `resume_id` |
| 简历优化建议 | `job_id` + `resume_id` |
| 简历版本对比 | `from_id` + `to_id` |

---

## 九、后续阶段

```
K8-5-47-B：Resumes 表结构与后端全链路重构
K8-5-47-C：前端简历编辑页按新表结构重构
K8-5-47-D：AI 四核心简历读取链路适配
K8-5-47-E：reset-up 与空库手动 seed 验收
```

---

## 十、关键声明

```
✅ 不做旧数据迁移
✅ 不做旧 seed 兼容
✅ 不保留 V1/V2 混合结构
✅ schema_version = 3 为新起点
```

---

## 十一、构建结果

| 项 | 结果 |
|----|:---:|
| `npm run build` | ✅ |
| `go build ./...` | ✅ |

---

## 十二、未修改范围

```
✅ 未修改源码（Go / TS / SQL）
✅ 未修改数据库
✅ 未执行 make reset-up
✅ 未执行 Git 命令
```
