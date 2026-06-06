# K8-5-48-A：Resumes 关系型拆表结构设计冻结

> 状态：✅ 已完成 | 日期：2026-06-06 | 性质：设计冻结，未修改源码

---

## 一、为什么推翻 K8-5-47 JSON 大表设计

```
K8-5-47 方案：resumes 表承载 9 个 JSON 模块字段
问题：
1. 核心业务数据（教育、项目、实习、证书、获奖）长期藏在 JSON 里，不便于 CRUD
2. 全 JSON 设计过度照顾 AI 输入和 Demo 快速实现
3. 字段校验、扩展、查询、管理、统计、维护困难
4. 真实招聘网站的简历系统应采用关系型拆表结构

新方向：
→ 主表存基础信息 + 技能 + AI 文本快照
→ 教育、项目、实习、外语考试、证书、获奖、作品链接各拆子表
→ raw_text 保留供 AI 使用
```

---

## 二、新表结构总览（8 张表）

| 表名 | 用途 | 关键外键 |
|------|------|---------|
| `resumes` | 简历主表（基础信息、技能、补充信息、AI 快照） | `user_id` → `users.id` |
| `resume_educations` | 教育经历 | `resume_id` → `resumes.id` CASCADE |
| `resume_projects` | 项目经历 | `resume_id` → `resumes.id` CASCADE |
| `resume_internships` | 实习经历 | `resume_id` → `resumes.id` CASCADE |
| `resume_language_exams` | 外语考试 | `resume_id` → `resumes.id` CASCADE |
| `resume_certificates` | 证书 | `resume_id` → `resumes.id` CASCADE |
| `resume_awards` | 获奖 | `resume_id` → `resumes.id` CASCADE |
| `resume_works` | 作品链接 | `resume_id` → `resumes.id` CASCADE |

---

## 三、主表 DDL：resumes

```sql
CREATE TABLE `resumes` (
  `id`              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id`         BIGINT UNSIGNED NOT NULL,

  `name`            VARCHAR(100) NOT NULL DEFAULT ''   COMMENT '简历名称',
  `is_current`      TINYINT(1)   NOT NULL DEFAULT 0    COMMENT '是否当前默认简历',

  `real_name`       VARCHAR(50)  NOT NULL DEFAULT ''   COMMENT '姓名',
  `phone`           VARCHAR(50)  NOT NULL DEFAULT ''   COMMENT '手机号',
  `email`           VARCHAR(255) NOT NULL DEFAULT ''   COMMENT '邮箱',
  `city`            VARCHAR(100) NOT NULL DEFAULT ''   COMMENT '当前城市',

  `professional_skills` TEXT NULL COMMENT '专业技能 / 工具能力',
  `ai_application`      TEXT NULL COMMENT 'AI 应用技能',
  `homepage`            VARCHAR(500) NOT NULL DEFAULT '' COMMENT '个人主页',
  `additional_info`     TEXT NULL COMMENT '补充信息',

  `raw_text`        MEDIUMTEXT NULL COMMENT 'AI 文本快照',

  `created_at`      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

  PRIMARY KEY (`id`),
  KEY `idx_resumes_user_current` (`user_id`, `is_current`, `updated_at`),
  KEY `idx_resumes_user_updated`  (`user_id`, `updated_at`),
  CONSTRAINT `fk_resumes_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='简历主表';
```

- `real_name`/`phone`/`email`/`city` 不放 JSON，作为主表正式字段
- `professional_skills`/`ai_application` 不放 JSON，直接主表字段
- `homepage`/`additional_info` 直接主表字段
- `raw_text` 由后端根据主表+子表生成
- `is_current` 替代 `version_no`

---

## 四、子表 DDL

### resume_educations（教育经历）

```sql
CREATE TABLE `resume_educations` (
  `id`            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `resume_id`     BIGINT UNSIGNED NOT NULL,
  `school`        VARCHAR(100) NOT NULL DEFAULT '',
  `degree`        VARCHAR(50)  NOT NULL DEFAULT '',
  `major`         VARCHAR(100) NOT NULL DEFAULT '',
  `department`    VARCHAR(100) NOT NULL DEFAULT '',
  `start_date`    VARCHAR(20)  NOT NULL DEFAULT '',
  `end_date`      VARCHAR(20)  NOT NULL DEFAULT '',
  `gpa`           VARCHAR(20)  NOT NULL DEFAULT '',
  `gpa_scale`     VARCHAR(20)  NOT NULL DEFAULT '',
  `rank_text`     VARCHAR(50)  NOT NULL DEFAULT '',
  `core_courses`  TEXT NULL,
  `research_direction` VARCHAR(255) NOT NULL DEFAULT '',
  `lab`           VARCHAR(255) NOT NULL DEFAULT '',
  `paper_summary` TEXT NULL,
  `sort_order`    INT NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_re_resume` (`resume_id`, `sort_order`),
  CONSTRAINT `fk_re_resume` FOREIGN KEY (`resume_id`) REFERENCES `resumes` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

### resume_projects（项目经历）

```sql
CREATE TABLE `resume_projects` (
  `id`            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `resume_id`     BIGINT UNSIGNED NOT NULL,
  `name`          VARCHAR(150) NOT NULL DEFAULT '',
  `role`          VARCHAR(100) NOT NULL DEFAULT '',
  `start_date`    VARCHAR(20)  NOT NULL DEFAULT '',
  `end_date`      VARCHAR(20)  NOT NULL DEFAULT '',
  `description`   TEXT NULL,
  `sort_order`    INT NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_rp_resume` (`resume_id`, `sort_order`),
  CONSTRAINT `fk_rp_resume` FOREIGN KEY (`resume_id`) REFERENCES `resumes` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

### resume_internships（实习经历）

```sql
CREATE TABLE `resume_internships` (
  `id`            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `resume_id`     BIGINT UNSIGNED NOT NULL,
  `company`       VARCHAR(150) NOT NULL DEFAULT '',
  `position`      VARCHAR(100) NOT NULL DEFAULT '',
  `start_date`    VARCHAR(20)  NOT NULL DEFAULT '',
  `end_date`      VARCHAR(20)  NOT NULL DEFAULT '',
  `description`   TEXT NULL,
  `sort_order`    INT NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_ri_resume` (`resume_id`, `sort_order`),
  CONSTRAINT `fk_ri_resume` FOREIGN KEY (`resume_id`) REFERENCES `resumes` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

### resume_language_exams（外语考试）

```sql
CREATE TABLE `resume_language_exams` (
  `id`            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `resume_id`     BIGINT UNSIGNED NOT NULL,
  `exam_type`     VARCHAR(100) NOT NULL DEFAULT '',
  `score`         VARCHAR(100) NOT NULL DEFAULT '',
  `sort_order`    INT NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_rle_resume` (`resume_id`, `sort_order`),
  CONSTRAINT `fk_rle_resume` FOREIGN KEY (`resume_id`) REFERENCES `resumes` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

### resume_certificates（证书）

```sql
CREATE TABLE `resume_certificates` (
  `id`            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `resume_id`     BIGINT UNSIGNED NOT NULL,
  `name`          VARCHAR(150) NOT NULL DEFAULT '',
  `description`   TEXT NULL,
  `sort_order`    INT NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_rc_resume` (`resume_id`, `sort_order`),
  CONSTRAINT `fk_rc_resume` FOREIGN KEY (`resume_id`) REFERENCES `resumes` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

### resume_awards（获奖）

```sql
CREATE TABLE `resume_awards` (
  `id`            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `resume_id`     BIGINT UNSIGNED NOT NULL,
  `name`          VARCHAR(150) NOT NULL DEFAULT '',
  `description`   TEXT NULL,
  `sort_order`    INT NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_ra_resume` (`resume_id`, `sort_order`),
  CONSTRAINT `fk_ra_resume` FOREIGN KEY (`resume_id`) REFERENCES `resumes` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

### resume_works（作品链接）

```sql
CREATE TABLE `resume_works` (
  `id`            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `resume_id`     BIGINT UNSIGNED NOT NULL,
  `name`          VARCHAR(150) NOT NULL DEFAULT '',
  `url`           VARCHAR(500) NOT NULL DEFAULT '',
  `password`      VARCHAR(100) NOT NULL DEFAULT '',
  `sort_order`    INT NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_rw_resume` (`resume_id`, `sort_order`),
  CONSTRAINT `fk_rw_resume` FOREIGN KEY (`resume_id`) REFERENCES `resumes` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

---

## 五、废弃清单

| 废弃项 | 类型 | 原因 |
|--------|------|------|
| `profile_json` | JSON 列 | 拆入主表 real_name/phone/email/city |
| `education_json` | JSON 列 | 拆入 resume_educations 子表 |
| `project_json` | JSON 列 | 拆入 resume_projects 子表 |
| `skills_json` | JSON 列 | 拆入主表 professional_skills/ai_application |
| `internship_json` | JSON 列 | 拆入 resume_internships 子表 |
| `certificates_json` | JSON 列 | 拆入 resume_certificates + resume_language_exams |
| `awards_json` | JSON 列 | 拆入 resume_awards 子表 |
| `portfolio_json` | JSON 列 | 拆入 main.homepage + resume_works 子表 |
| `extra_json` | JSON 列 | 拆入主表 additional_info |
| `target_city/target_role/...` | 旧字段 | 不再需要 |
| `job_intention_json` | JSON 列 | 不再需要 |
| `experience_json` | JSON 列 | 统一为 resume_internships |
| `self_evaluation` | 文本列 | 合并到 additional_info |
| `resume_versions` 表 | 表 | 不再存在 |
| `version_no` | 概念 | 不再使用 |

---

## 六、后端影响面

| 层级 | 后续需要改（K8-5-48-B） |
|------|------------------------|
| `init.sql` | 创建 8 张 resumes 相关表 DDL + DROP 顺序 |
| `model/` | Resume + Education + Project + Internship + LanguageExam + Certificate + Award + Work |
| `repository/` | 主表 CRUD + 子表批量创建/删除/查询 |
| `service/` | 事务保存整份简历、raw_text 生成、current 逻辑 |
| `handler/` | API request/response 嵌套结构 |
| `agent/` | 从 raw_text 或组装后的结构读取，不再依赖 JSON 字段 |

---

## 七、前端影响面

| 层级 | 后续需要改（K8-5-48-C） |
|------|------------------------|
| `types/resume.ts` | 关系型嵌套 DTO |
| `api/resumes.ts` | 发送/接收嵌套结构 |
| `MyResumesPage.tsx` | 保存 payload 改为嵌套结构 |
| `ResumeManagementPage.tsx` | 使用 resume.id；不再用 versionNo |

---

## 八、AI 四核心影响面

| 功能 | 新输入方式 |
|------|-----------|
| 岗位推荐 | `resume_id` → 后端从主表+子表生成 raw_text → LLM |
| 岗位匹配 | `job_id` + `resume_id` |
| 简历优化 | `job_id` + `resume_id` |
| 简历对比 | `from_id` + `to_id` |

Agent 不再依赖 JSON 字段或 version_no。

---

## 九、后续阶段

```
K8-5-48-B：Resumes 关系型表结构与后端全链路重构
K8-5-48-C：前端简历模块按关系型 DTO 重构
K8-5-48-D：AI 四核心 resume_id / raw_text 链路适配
K8-5-48-E：reset-up 与空库手动 seed 验收
```

---

## 十、构建结果

| 项 | 结果 |
|----|:---:|
| `npm run build` | ✅ |
| `go build ./...` | ✅ |

## 十一、未修改范围

```
✅ 未修改源码（Go / TS / SQL）
✅ 未修改数据库
✅ 未执行 make reset-up
✅ 未执行 Git 命令
✅ 不做旧数据兼容、不做旧 seed 迁移
```
