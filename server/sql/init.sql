-- ============================================================
-- Offerr鹅 - 数据库初始化脚本（唯一入口）
-- K8-5-51-B：统一 clean 初始化，15 张目标表
-- MySQL 8.0 / InnoDB / utf8mb4
--
-- 约定：
-- 1. DROP TABLE IF EXISTS + CREATE TABLE，支持可重复执行
-- 2. 删除顺序按外键依赖反向，建表顺序按外键依赖正向
-- 3. 种子数据与表结构统一维护在本文件
-- 4. 清空 volume 后重新导入即可重建数据
-- 5. resumes 替代 resume_versions（移除 version_no）
-- 6. resume_compare_results 替代 version_compare_results
-- ============================================================

SET NAMES utf8mb4;
SET CHARACTER SET utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ============================================================
-- 1. 删除表（按外键依赖反向）
-- ============================================================

DROP TABLE IF EXISTS `resume_optimization_results`;
DROP TABLE IF EXISTS `resume_compare_results`;
DROP TABLE IF EXISTS `version_compare_results`;
DROP TABLE IF EXISTS `match_analyses`;
DROP TABLE IF EXISTS `recommendation_results`;

DROP TABLE IF EXISTS `job_filter_option_relations`;
DROP TABLE IF EXISTS `job_filter_options`;
DROP TABLE IF EXISTS `job_filter_groups`;
DROP TABLE IF EXISTS `job_skills`;
DROP TABLE IF EXISTS `job_locations`;
DROP TABLE IF EXISTS `job_business_groups`;
DROP TABLE IF EXISTS `jobs`;

DROP TABLE IF EXISTS `resumes`;
DROP TABLE IF EXISTS `resume_versions`;
DROP TABLE IF EXISTS `seed_versions`;
DROP TABLE IF EXISTS `users`;

-- ============================================================
-- 2. 创建表（按外键依赖正向）
-- ============================================================

-- -----------------------------------------------------------
-- 2.1 users - 用户注册、登录、基础身份识别
-- -----------------------------------------------------------
CREATE TABLE `users` (
  `id`            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `email`         VARCHAR(255)     NOT NULL DEFAULT '',
  `phone`         VARCHAR(20)      NOT NULL DEFAULT '',
  `password_hash` VARCHAR(255)     NOT NULL,
  `name`          VARCHAR(100)     NOT NULL DEFAULT '',
  `created_at`    DATETIME         NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`    DATETIME         NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_users_phone` (`phone`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- -----------------------------------------------------------
-- 2.2 resumes - 简历主表（替代 resume_versions，移除 version_no）
-- -----------------------------------------------------------
CREATE TABLE `resumes` (
  `id`              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id`         BIGINT UNSIGNED NOT NULL,
  `schema_version`  INTEGER          NOT NULL DEFAULT 1     COMMENT 'schema版本：1=V1旧版 2=V2新版',
  `name`            VARCHAR(100)     NOT NULL DEFAULT '',
  `phone`           VARCHAR(50)      NOT NULL DEFAULT '',
  `email`           VARCHAR(255)     NOT NULL DEFAULT '',
  `location`        VARCHAR(100)     NOT NULL DEFAULT '',
  `target_city`     VARCHAR(100)     NOT NULL DEFAULT '',
  `target_role`     VARCHAR(100)     NOT NULL DEFAULT '',
  `education_json`  JSON             NULL,
  `experience_json` JSON             NULL,
  `project_json`    JSON             NULL,
  `skills_json`     JSON             NULL,
  -- V2 摘要结构化列
  `target_direction` VARCHAR(100)     NOT NULL DEFAULT ''    COMMENT '目标方向（摘要列）',
  `job_type`         VARCHAR(50)      NOT NULL DEFAULT ''    COMMENT '岗位类型（摘要列）',
  `expected_cities`  VARCHAR(255)     NOT NULL DEFAULT ''    COMMENT '期望城市逗号分隔（摘要列）',
  `school`           VARCHAR(100)     NOT NULL DEFAULT ''    COMMENT '学校（摘要列）',
  `degree`           VARCHAR(20)      NOT NULL DEFAULT ''    COMMENT '学历（摘要列）',
  `major`            VARCHAR(100)     NOT NULL DEFAULT ''    COMMENT '专业（摘要列）',
  `graduation_year`  VARCHAR(10)      NOT NULL DEFAULT ''    COMMENT '毕业时间（摘要列）',
  -- V2 模块 JSON 字段
  `profile_json`       JSON           NULL                   COMMENT '基础信息 JSON',
  `job_intention_json` JSON           NULL                   COMMENT '求职意向 JSON',
  `internship_json`    JSON           NULL                   COMMENT '实习经历 JSON',
  `certificates_json`  JSON           NULL                   COMMENT '证书与语言 JSON',
  `awards_json`        JSON           NULL                   COMMENT '获奖与竞赛 JSON',
  `portfolio_json`     JSON           NULL                   COMMENT '作品链接 JSON',
  `extra_json`         JSON           NULL                   COMMENT '补充信息 JSON',
  `self_evaluation` TEXT             NULL,
  `raw_text`        TEXT             NULL,
  `created_at`      DATETIME         NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`      DATETIME         NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_resumes_user_created` (`user_id`, `created_at`),
  CONSTRAINT `fk_resumes_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='简历表';



-- -----------------------------------------------------------
-- 2.3 jobs - 岗位列表和详情
-- -----------------------------------------------------------
CREATE TABLE `jobs` (
  `id`                  BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `title`               VARCHAR(200)     NOT NULL,
  `department`          VARCHAR(200)     NOT NULL DEFAULT '',
  `city`                VARCHAR(100)     NOT NULL DEFAULT '',
  `direction`           VARCHAR(100)     NOT NULL DEFAULT '',
  `category`            VARCHAR(100)     NOT NULL DEFAULT '',
  `job_type`            VARCHAR(100)     NOT NULL DEFAULT '',
  `application_type`    VARCHAR(64)      NOT NULL DEFAULT '' COMMENT '申请类型：应届实习/校园招聘等',
  `application_project` VARCHAR(64)      NOT NULL DEFAULT '' COMMENT '申请项目：实习生/应届生等',
  `application_subtype` VARCHAR(64)      NOT NULL DEFAULT '' COMMENT '申请子类型',
  `slogan`              VARCHAR(200)     NOT NULL DEFAULT '' COMMENT '岗位宣传语',
  `summary`             TEXT             NULL COMMENT '岗位概要',
  `description`         TEXT             NULL,
  `requirements`        TEXT             NULL,
  `responsibilities_json` JSON           NULL COMMENT '职责列表 JSON',
  `must_have_json`      JSON             NULL COMMENT '必备要求 JSON',
  `nice_to_have_json`   JSON             NULL COMMENT '了解项 JSON',
  `bonus_points_json`   JSON             NULL COMMENT '加分项 JSON',
  `notes_json`          JSON             NULL COMMENT '备注 JSON',
  `skill_keywords_json` JSON             NULL COMMENT '技能关键词 JSON（旧兼容）',
  `skills_json`         JSON             NULL COMMENT '技能列表 JSON',
  `organizations_json`  JSON             NULL COMMENT '事业群/部门 JSON [{code,name}]',
  `locations_json`      JSON             NULL COMMENT '工作城市 JSON',
  `interview_locations_json` JSON        NULL COMMENT '面试城市 JSON',
  `facets_json`         JSON             NULL COMMENT '筛选聚合原始 JSON',
  `status`              VARCHAR(50)      NOT NULL DEFAULT 'active',
  `external_source`     VARCHAR(50)      NOT NULL DEFAULT 'local_mock' COMMENT '数据来源：local_mock / manual_curated / tencent_join',
  `external_id`         VARCHAR(100)     NULL COMMENT '外部岗位ID',
  `external_inner_id`   VARCHAR(100)     NULL COMMENT '外部系统内部ID',
  `source_url`          VARCHAR(500)     NULL COMMENT '来源页面URL',
  `source_name`         VARCHAR(200)     NOT NULL DEFAULT '' COMMENT '数据来源名称',
  `raw_json`            JSON             NULL COMMENT '完整原始岗位JSON',
  `data_quality`        VARCHAR(50)      NOT NULL DEFAULT 'manual_enriched' COMMENT '数据质量',
  `collected_at`        DATETIME         NULL COMMENT '数据采集时间',
  `is_featured`         TINYINT(1)       NOT NULL DEFAULT 0 COMMENT '是否Demo精选岗位',
  `synced_at`           TIMESTAMP        NULL COMMENT '外部岗位同步时间',
  `created_at`          DATETIME         NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`          DATETIME         NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_jobs_external_source_id` (`external_source`, `external_id`),
  KEY `idx_jobs_city` (`city`),
  KEY `idx_jobs_direction` (`direction`),
  KEY `idx_jobs_category` (`category`),
  KEY `idx_jobs_title` (`title`),
  KEY `idx_jobs_external_source` (`external_source`),
  KEY `idx_jobs_data_quality` (`data_quality`),
  KEY `idx_jobs_is_featured` (`is_featured`),
  KEY `idx_jobs_job_type` (`job_type`),
  KEY `idx_jobs_application_type` (`application_type`),
  KEY `idx_jobs_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='岗位表';

-- -----------------------------------------------------------
-- job_locations - 岗位工作城市关联表
-- -----------------------------------------------------------
CREATE TABLE `job_locations` (
  `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `job_id`     BIGINT UNSIGNED NOT NULL,
  `city`       VARCHAR(64)      NOT NULL,
  `created_at` DATETIME         NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_job_location` (`job_id`, `city`),
  KEY `idx_jl_city` (`city`),
  CONSTRAINT `fk_jl_job` FOREIGN KEY (`job_id`) REFERENCES `jobs` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='岗位工作城市表';

-- -----------------------------------------------------------
-- job_business_groups - 岗位事业群关联表
-- -----------------------------------------------------------
CREATE TABLE `job_business_groups` (
  `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `job_id`     BIGINT UNSIGNED NOT NULL,
  `code`       VARCHAR(32)      NOT NULL,
  `name`       VARCHAR(128)     NOT NULL,
  `created_at` DATETIME         NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_job_bg` (`job_id`, `code`),
  KEY `idx_jbg_code` (`code`),
  CONSTRAINT `fk_jbg_job` FOREIGN KEY (`job_id`) REFERENCES `jobs` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='岗位事业群表';

-- -----------------------------------------------------------
-- job_skills - 岗位技能关联表
-- -----------------------------------------------------------
CREATE TABLE `job_skills` (
  `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `job_id`     BIGINT UNSIGNED NOT NULL,
  `skill`      VARCHAR(128)     NOT NULL,
  `created_at` DATETIME         NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_job_skill` (`job_id`, `skill`),
  KEY `idx_js_skill` (`skill`),
  CONSTRAINT `fk_js_job` FOREIGN KEY (`job_id`) REFERENCES `jobs` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='岗位技能表';

-- -----------------------------------------------------------
-- job_filter_groups - 筛选分组元数据
-- -----------------------------------------------------------
CREATE TABLE `job_filter_groups` (
  `id`                  INT UNSIGNED  NOT NULL AUTO_INCREMENT,
  `code`                VARCHAR(64)   NOT NULL COMMENT '分组代码',
  `name`                VARCHAR(64)   NOT NULL COMMENT '显示名称',
  `display_order`       INT           NOT NULL DEFAULT 0,
  `default_expanded`    TINYINT(1)    NOT NULL DEFAULT 1,
  `max_visible_options` INT           NOT NULL DEFAULT 6,
  `visible`             TINYINT(1)    NOT NULL DEFAULT 1,
  `created_at`          DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_filter_group_code` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='筛选分组表';

-- -----------------------------------------------------------
-- job_filter_options - 筛选选项元数据
-- -----------------------------------------------------------
CREATE TABLE `job_filter_options` (
  `id`               INT UNSIGNED NOT NULL AUTO_INCREMENT,
  `group_id`         INT UNSIGNED NOT NULL,
  `code`             VARCHAR(128) NOT NULL,
  `label`            VARCHAR(128) NOT NULL,
  `value`            VARCHAR(128) NOT NULL,
  `display_order`    INT          NOT NULL DEFAULT 0,
  `visible`          TINYINT(1)   NOT NULL DEFAULT 1,
  `parent_option_id` INT UNSIGNED NULL,
  `created_at`       DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_option_group_code` (`group_id`, `code`),
  KEY `idx_option_parent` (`parent_option_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='筛选选项表';

-- -----------------------------------------------------------
-- job_filter_option_relations - 岗位-筛选项关联
-- -----------------------------------------------------------
CREATE TABLE `job_filter_option_relations` (
  `job_id`    BIGINT UNSIGNED NOT NULL,
  `option_id` INT UNSIGNED   NOT NULL,
  `created_at` DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`job_id`, `option_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='岗位-筛选项关联表';

-- -----------------------------------------------------------
-- seed_versions - seed 版本追踪
-- -----------------------------------------------------------
CREATE TABLE `seed_versions` (
  `seed_key`     VARCHAR(64)  NOT NULL,
  `seed_version` VARCHAR(64)  NOT NULL,
  `applied_at`   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `checksum`     VARCHAR(32)  NOT NULL DEFAULT '',
  PRIMARY KEY (`seed_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Seed 版本追踪表';

-- -----------------------------------------------------------
-- recommendation_results - 岗位推荐结果快照
-- -----------------------------------------------------------
CREATE TABLE `recommendation_results` (
  `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id`     BIGINT UNSIGNED NOT NULL,
  `resume_id`   BIGINT UNSIGNED NOT NULL,
  `result_json` JSON             NOT NULL,
  `created_at`  DATETIME         NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_rec_user_created` (`user_id`, `created_at`),
  KEY `idx_rec_resume` (`resume_id`),
  CONSTRAINT `fk_rec_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`),
  CONSTRAINT `fk_rec_resume` FOREIGN KEY (`resume_id`) REFERENCES `resumes` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='推荐结果表';

-- -----------------------------------------------------------
-- match_analyses - 岗位匹配分析结果
-- -----------------------------------------------------------
CREATE TABLE `match_analyses` (
  `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id`     BIGINT UNSIGNED NOT NULL,
  `resume_id`   BIGINT UNSIGNED NOT NULL,
  `job_id`      BIGINT UNSIGNED NOT NULL,
  `score`       INT              NOT NULL DEFAULT 0,
  `result_json` JSON             NOT NULL,
  `created_at`  DATETIME         NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_match_user_created` (`user_id`, `created_at`),
  KEY `idx_match_job` (`job_id`),
  KEY `idx_match_resume` (`resume_id`),
  CONSTRAINT `fk_match_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`),
  CONSTRAINT `fk_match_resume` FOREIGN KEY (`resume_id`) REFERENCES `resumes` (`id`),
  CONSTRAINT `fk_match_job` FOREIGN KEY (`job_id`) REFERENCES `jobs` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='匹配分析结果表';

-- -----------------------------------------------------------
-- resume_compare_results - 简历对比结果（替代 version_compare_results）
-- -----------------------------------------------------------
CREATE TABLE `resume_compare_results` (
  `id`              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id`         BIGINT UNSIGNED NOT NULL,
  `from_resume_id`  BIGINT UNSIGNED NOT NULL,
  `to_resume_id`    BIGINT UNSIGNED NOT NULL,
  `result_json`     JSON             NOT NULL,
  `created_at`      DATETIME         NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_compare_user_created` (`user_id`, `created_at`),
  KEY `idx_compare_from_to` (`from_resume_id`, `to_resume_id`),
  CONSTRAINT `fk_compare_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`),
  CONSTRAINT `fk_compare_from` FOREIGN KEY (`from_resume_id`) REFERENCES `resumes` (`id`),
  CONSTRAINT `fk_compare_to` FOREIGN KEY (`to_resume_id`) REFERENCES `resumes` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='简历对比结果表';

-- -----------------------------------------------------------
-- resume_optimization_results - 简历优化结果快照
-- -----------------------------------------------------------
CREATE TABLE `resume_optimization_results` (
  `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id`     BIGINT UNSIGNED NOT NULL,
  `resume_id`   BIGINT UNSIGNED NULL COMMENT '关联的简历ID',
  `job_id`      BIGINT UNSIGNED NULL COMMENT '关联的岗位ID',
  `result_json` JSON             NOT NULL,
  `created_at`  DATETIME         NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_opt_user_created` (`user_id`, `created_at`),
  KEY `idx_opt_job` (`job_id`),
  KEY `idx_opt_resume` (`resume_id`),
  CONSTRAINT `fk_opt_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='简历优化结果表';

SET FOREIGN_KEY_CHECKS = 1;

-- ============================================================
-- 3. 种子数据
-- ============================================================

-- -----------------------------------------------------------
-- 3.1 种子用户
-- Demo 密码统一为 123456
-- -----------------------------------------------------------
INSERT INTO `users` (`id`, `email`, `phone`, `password_hash`, `name`) VALUES
(1, 'arthur@example.com',       '13800138000', '$2a$10$hUKwISi3F6uHLRj.cFkUkO23HNLW9nH60me2O9fm7q2Hr/Z1nTG3u', 'Arthur'),
(2, 'demo@offer-catcher.com',   '13900139000', '$2a$10$hUKwISi3F6uHLRj.cFkUkO23HNLW9nH60me2O9fm7q2Hr/Z1nTG3u', 'Demo用户');

-- -----------------------------------------------------------
-- 3.2 种子简历（Arthur 用户 x2）
-- -----------------------------------------------------------
INSERT INTO `resumes` (`id`, `user_id`, `name`, `phone`, `email`, `location`, `target_city`, `target_role`, `education_json`, `experience_json`, `project_json`, `skills_json`, `self_evaluation`) VALUES
(1, 1, 'Arthur', '13912345678', 'arthur@example.com', '深圳', '深圳', '后端开发工程师',
  '[{"school":"南方科技大学","major":"计算机科学与技术","degree":"本科","start_date":"2022-09","end_date":"2026-06"}]',
  '[{"company":"某科技公司","position":"后端开发实习生","start_date":"2025-03","end_date":"2025-09","description":"参与公司内部工具的后端开发，使用 Go 语言编写 API 接口，负责数据库设计"}]',
  '[{"name":"校园二手交易平台","role":"后端负责人","start_date":"2024-03","end_date":"2024-09","description":"使用 Go 框架开发 API，实现认证、商品发布等功能"}]',
  '["Go","MySQL","Redis","Docker","Git","Linux"]',
  '计算机专业本科生，对后端开发和云原生技术有浓厚兴趣。'),
(2, 1, 'Arthur', '13912345678', 'arthur@example.com', '深圳', '深圳', '后端开发工程师',
  '[{"school":"南方科技大学","major":"计算机科学与技术","degree":"本科","start_date":"2022-09","end_date":"2026-06"}]',
  '[{"company":"某科技公司","position":"后端开发实习生","start_date":"2025-03","end_date":"2025-09","description":"使用 Go 语言编写 API，独立完成 3 个微服务开发和上线，处理峰值 QPS 达 5000+"},{"company":"某云计算公司","position":"云原生开发实习生","start_date":"2025-10","end_date":"2026-03","description":"参与 Kubernetes Operator 开发，使用 Go 编写自定义控制器"}]',
  '[{"name":"校园二手交易平台","role":"后端负责人","start_date":"2024-03","end_date":"2024-09","description":"使用 Go 框架开发 API，日均活跃用户 500+，累计订单 2000+ 笔"},{"name":"分布式任务调度系统","role":"核心开发者","start_date":"2025-01","end_date":"2025-06","description":"基于 etcd 实现分布式任务调度，调度延迟小于 100ms"}]',
  '["Go","MySQL","Redis","Docker","Kubernetes","etcd","gRPC","Git","Linux"]',
  '计算机专业本科生，具备 Go 后端开发和云原生实践经验。');

-- ============================================================
-- 注意：
-- 腾讯岗位 seed (tencent_jobs.sql) 由 up.sh step 6b 自动导入（含 46 条岗位）
-- 9 份示范简历 seed (resume_versions.sql) 由 mysql.sh import 自动导入
-- AI 结果 seed 暂不导入（依赖 jobs 表，jobs 由 tencent_jobs 导入）
-- ============================================================
