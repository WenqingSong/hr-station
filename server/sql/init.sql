-- ============================================================
-- Offer 捕手 - 数据库初始化脚本
-- K8-4-3-R2: 确保导入时使用 utf8mb4 编码
SET NAMES utf8mb4;
-- ============================================================
-- 阶段 3-2：表结构与种子数据
-- 阶段 4-1.3.1：jobs 表扩展
-- 阶段 J1：jobs 表结构重构
-- 阶段 K8-5-22-A：resume_versions 表 V2 字段扩展
-- MySQL 8.0 / InnoDB / utf8mb4
--
-- 约定：
-- 1. 使用 DROP TABLE IF EXISTS + CREATE TABLE，支持可重复执行
-- 2. 删除顺序按外键依赖反向，建表顺序按外键依赖正向
-- 3. 种子数据与表结构统一维护在本文件
-- 4. 清空 volume 后重新导入即可重建数据
-- ============================================================

SET NAMES utf8mb4;
SET CHARACTER SET utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ============================================================
-- 1. 删除表（按外键依赖反向）
-- ============================================================

DROP TABLE IF EXISTS `version_compare_results`;
DROP TABLE IF EXISTS `resume_optimization_results`;
DROP TABLE IF EXISTS `match_analyses`;
DROP TABLE IF EXISTS `recommendation_results`;
DROP TABLE IF EXISTS `job_skills`;
DROP TABLE IF EXISTS `job_locations`;
DROP TABLE IF EXISTS `job_business_groups`;
DROP TABLE IF EXISTS `resume_versions`;
DROP TABLE IF EXISTS `jobs`;
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
-- 2.2 resume_versions - 简历版本管理（单主简历，不做多简历）
-- -----------------------------------------------------------
CREATE TABLE `resume_versions` (
  `id`              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id`         BIGINT UNSIGNED NOT NULL,
  `version_no`      INT             NOT NULL,
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
  -- V2 新增字段 (K8-5-22-A)：schema_version + 摘要结构化列 + 模块 JSON 字段
  -- 旧 V1 字段全部保留，V2 字段与 V1 字段并存
  `schema_version`   INTEGER          NOT NULL DEFAULT 1     COMMENT 'schema版本：1=V1旧版 2=V2新版',
  `target_direction` VARCHAR(100)     NOT NULL DEFAULT ''    COMMENT '目标方向（摘要列）',
  `job_type`         VARCHAR(50)      NOT NULL DEFAULT ''    COMMENT '岗位类型（摘要列）',
  `expected_cities`  VARCHAR(255)     NOT NULL DEFAULT ''    COMMENT '期望城市逗号分隔（摘要列）',
  `school`           VARCHAR(100)     NOT NULL DEFAULT ''    COMMENT '学校（摘要列）',
  `degree`           VARCHAR(20)      NOT NULL DEFAULT ''    COMMENT '学历（摘要列）',
  `major`            VARCHAR(100)     NOT NULL DEFAULT ''    COMMENT '专业（摘要列）',
  `graduation_year`  VARCHAR(10)      NOT NULL DEFAULT ''    COMMENT '毕业时间（摘要列）',
  `profile_json`       JSON           NULL                   COMMENT '基础信息 JSON',
  `job_intention_json` JSON           NULL                   COMMENT '求职意向 JSON',
  `internship_json`    JSON           NULL                   COMMENT '实习经历 JSON（V2新增，与experience_json并存）',
  `certificates_json`  JSON           NULL                   COMMENT '证书与语言 JSON',
  `awards_json`        JSON           NULL                   COMMENT '获奖与竞赛 JSON',
  `portfolio_json`     JSON           NULL                   COMMENT '作品链接 JSON',
  `extra_json`         JSON           NULL                   COMMENT '补充信息 JSON（self_evaluation以列+JSON双存）',
  `self_evaluation` TEXT             NULL,
  `raw_text`        TEXT             NULL,
  `created_at`      DATETIME         NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_resume_user_version` (`user_id`, `version_no`),
  KEY `idx_resume_user_created` (`user_id`, `created_at`),
  CONSTRAINT `fk_resume_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='简历版本表';

-- -----------------------------------------------------------
-- 2.3 jobs - 岗位列表和详情
-- 支持本地 Mock 岗位（external_source=local_mock）和腾讯公开岗位（external_source=tencent_join）
-- 阶段 4-1.3.1 新增 external_source / external_id / raw_json / data_quality 等字段
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
-- 2.3a job_locations - 岗位工作城市关联表（支持筛选聚合）
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
-- 2.3b job_business_groups - 岗位事业群关联表（支持筛选聚合）
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
-- 2.3c job_skills - 岗位技能关联表（支持筛选聚合）
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
-- 2.4 job_filter_groups - 筛选分组元数据 （K8-4-1 设计冻结）
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
-- 2.5 job_filter_options - 筛选选项元数据 （K8-4-1 设计冻结）
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
-- 2.6 job_filter_option_relations - 岗位-筛选项关联 （K8-4-1 设计冻结）
-- -----------------------------------------------------------
CREATE TABLE `job_filter_option_relations` (
  `job_id`    BIGINT UNSIGNED NOT NULL,
  `option_id` INT UNSIGNED   NOT NULL,
  `created_at` DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`job_id`, `option_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='岗位-筛选项关联表';

-- -----------------------------------------------------------
-- 2.7 seed_versions - seed 版本追踪
-- -----------------------------------------------------------
CREATE TABLE `seed_versions` (
  `seed_key`     VARCHAR(64)  NOT NULL,
  `seed_version` VARCHAR(64)  NOT NULL,
  `applied_at`   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `checksum`     VARCHAR(32)  NOT NULL DEFAULT '',
  PRIMARY KEY (`seed_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Seed 版本追踪表';

-- -----------------------------------------------------------
-- 3. 推荐与匹配分析结果表
-- -----------------------------------------------------------

-- -----------------------------------------------------------
-- 2.4 recommendation_results - 岗位推荐 Mock 结果快照
-- -----------------------------------------------------------
CREATE TABLE `recommendation_results` (
  `id`                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id`           BIGINT UNSIGNED NOT NULL,
  `resume_version_id` BIGINT UNSIGNED NOT NULL,
  `result_json`       JSON             NOT NULL,
  `created_at`        DATETIME         NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_rec_user_created` (`user_id`, `created_at`),
  KEY `idx_rec_resume` (`resume_version_id`),
  CONSTRAINT `fk_rec_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`),
  CONSTRAINT `fk_rec_resume` FOREIGN KEY (`resume_version_id`) REFERENCES `resume_versions` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='推荐结果表';

-- -----------------------------------------------------------
-- 2.5 match_analyses - 岗位匹配分析 Mock 结果
-- -----------------------------------------------------------
CREATE TABLE `match_analyses` (
  `id`                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id`           BIGINT UNSIGNED NOT NULL,
  `resume_version_id` BIGINT UNSIGNED NOT NULL,
  `job_id`            BIGINT UNSIGNED NOT NULL,
  `score`             INT              NOT NULL DEFAULT 0,
  `result_json`       JSON             NOT NULL,
  `created_at`        DATETIME         NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_match_user_created` (`user_id`, `created_at`),
  KEY `idx_match_job` (`job_id`),
  KEY `idx_match_resume` (`resume_version_id`),
  CONSTRAINT `fk_match_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`),
  CONSTRAINT `fk_match_resume` FOREIGN KEY (`resume_version_id`) REFERENCES `resume_versions` (`id`),
  CONSTRAINT `fk_match_job` FOREIGN KEY (`job_id`) REFERENCES `jobs` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='匹配分析结果表';

-- -----------------------------------------------------------
-- 2.6 version_compare_results - 简历版本对比 Mock 结果
-- -----------------------------------------------------------
CREATE TABLE `version_compare_results` (
  `id`              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id`         BIGINT UNSIGNED NOT NULL,
  `from_version_id` BIGINT UNSIGNED NOT NULL,
  `to_version_id`   BIGINT UNSIGNED NOT NULL,
  `result_json`     JSON             NOT NULL,
  `created_at`      DATETIME         NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_compare_user_created` (`user_id`, `created_at`),
  KEY `idx_compare_from_to` (`from_version_id`, `to_version_id`),
  CONSTRAINT `fk_compare_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`),
  CONSTRAINT `fk_compare_from` FOREIGN KEY (`from_version_id`) REFERENCES `resume_versions` (`id`),
  CONSTRAINT `fk_compare_to` FOREIGN KEY (`to_version_id`) REFERENCES `resume_versions` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='版本对比结果表';

-- -----------------------------------------------------------
-- 2.7 resume_optimization_results - 简历优化结果快照
-- -----------------------------------------------------------
CREATE TABLE `resume_optimization_results` (
  `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id`     BIGINT UNSIGNED NOT NULL,
  `job_id`      BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '可选，0 表示无岗位引用',
  `result_json` JSON             NOT NULL,
  `created_at`  DATETIME         NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_opt_user_created` (`user_id`, `created_at`),
  KEY `idx_opt_job` (`job_id`),
  CONSTRAINT `fk_opt_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='简历优化结果表';

SET FOREIGN_KEY_CHECKS = 1;

-- ============================================================
-- 3. 种子数据
-- ============================================================

-- -----------------------------------------------------------
-- 3.1 种子用户
-- Demo 密码统一为 123456，password_hash 使用 bcrypt 生成
-- 注意：Demo 密码不应在生产环境使用
-- -----------------------------------------------------------
INSERT INTO `users` (`id`, `email`, `phone`, `password_hash`, `name`) VALUES
(1, 'arthur@example.com',       '13800138000', '$2a$10$hUKwISi3F6uHLRj.cFkUkO23HNLW9nH60me2O9fm7q2Hr/Z1nTG3u', 'Arthur'),
(2, 'demo@offer-catcher.com',   '13900139000', '$2a$10$hUKwISi3F6uHLRj.cFkUkO23HNLW9nH60me2O9fm7q2Hr/Z1nTG3u', 'Demo用户');

-- -----------------------------------------------------------
-- 3.2 种子简历版本
-- Arthur 用户 V1：偏基础的 Go 后端简历
-- Arthur 用户 V2：优化后的版本，方便版本对比 Demo
-- -----------------------------------------------------------
INSERT INTO `resume_versions` (`id`, `user_id`, `version_no`, `name`, `phone`, `email`, `location`, `target_city`, `target_role`, `education_json`, `experience_json`, `project_json`, `skills_json`, `self_evaluation`) VALUES
(1, 1, 1, 'Arthur', '13912345678', 'arthur@example.com', '深圳', '深圳', '后端开发工程师',
  '[{"school":"南方科技大学","major":"计算机科学与技术","degree":"本科","start_date":"2022-09","end_date":"2026-06"}]',
  '[{"company":"某科技公司","position":"后端开发实习生","start_date":"2025-03","end_date":"2025-09","description":"参与公司内部工具的后端开发，使用 Go 语言编写 API 接口，负责数据库设计"}]',
  '[{"name":"校园二手交易平台","role":"后端负责人","start_date":"2024-03","end_date":"2024-09","description":"使用 Go + Gin 框架开发 RESTful API，实现用户认证、商品发布、订单管理等功能"},{"name":"在线评测系统","role":"开发者","start_date":"2023-09","end_date":"2024-01","description":"使用 Python Flask 开发在线代码评测系统"}]',
  '["Go","MySQL","Redis","Docker","Git","Linux"]',
  '计算机专业本科生，对后端开发和云原生技术有浓厚兴趣，正在寻找后端开发工程师岗位。'),

(2, 1, 2, 'Arthur', '13912345678', 'arthur@example.com', '深圳', '深圳', '后端开发工程师',
  '[{"school":"南方科技大学","major":"计算机科学与技术","degree":"本科","start_date":"2022-09","end_date":"2026-06"}]',
  '[{"company":"某科技公司","position":"后端开发实习生","start_date":"2025-03","end_date":"2025-09","description":"参与公司内部工具的后端开发，使用 Go 语言编写 API 接口，负责数据库设计。独立完成 3 个微服务的开发和上线，处理峰值 QPS 达 5000+"},{"company":"某云计算公司","position":"云原生开发实习生","start_date":"2025-10","end_date":"2026-03","description":"参与 Kubernetes Operator 开发，使用 Go 编写自定义控制器，管理集群内的微服务编排与扩缩容"}]',
  '[{"name":"校园二手交易平台","role":"后端负责人","start_date":"2024-03","end_date":"2024-09","description":"使用 Go + Gin 框架开发 RESTful API，实现用户认证、商品发布、订单管理等功能。日均活跃用户 500+，累计订单 2000+ 笔"},{"name":"分布式任务调度系统","role":"核心开发者","start_date":"2025-01","end_date":"2025-06","description":"基于 etcd 实现分布式任务调度，支持 Cron 表达式、任务依赖、失败重试。调度延迟 < 100ms，支持 1000+ 并发任务"}]',
  '["Go","MySQL","Redis","Docker","Kubernetes","etcd","gRPC","Git","Linux"]',
  '计算机专业本科生，具备 Go 后端开发和云原生实践经验。有分布式系统和微服务架构的实战经验，正在寻找后端开发工程师岗位。');

-- -----------------------------------------------------------
-- 3.3 种子岗位
-- 阶段 4-1.3.1 更新：新增 external_source / external_id / raw_json / data_quality / is_featured / synced_at
-- 本地 Mock 岗位：external_source=local_mock, data_quality=manual_enriched, is_featured=1
-- -----------------------------------------------------------
INSERT INTO `jobs` (`id`, `title`, `department`, `city`, `direction`, `category`, `job_type`, `application_type`, `description`, `requirements`, `skill_keywords_json`, `locations_json`, `organizations_json`, `status`, `external_source`, `external_id`, `external_inner_id`, `source_url`, `raw_json`, `data_quality`, `is_featured`, `synced_at`) VALUES
(1, 'Go 后端开发工程师', '技术部', '深圳', '软件开发', '后端开发', '校招', '应届实习',
  '负责公司核心业务系统的后端开发与维护，参与微服务架构设计，保障系统高可用与高性能。与产品经理、前端工程师紧密协作，推动业务快速迭代。',
  '1. 计算机相关专业本科及以上学历\n2. 熟练掌握 Go 语言，了解 Goroutine、Channel 等并发模型\n3. 熟悉 MySQL、Redis 等存储系统\n4. 了解微服务架构，有 Docker/Kubernetes 使用经验优先\n5. 良好的沟通能力和团队协作精神',
  '["Go","MySQL","Redis","Docker","Kubernetes","微服务"]',
  '["深圳","北京"]',
  '[{"code":"CDG","name":"CDG 企业发展事业群"},{"code":"CSIG","name":"CSIG 云与智慧产业事业群"}]',
  'active',
  'local_mock', 'local_mock_1', NULL, NULL, JSON_OBJECT('source', 'local_mock', 'note', 'seed demo job'), 'manual_enriched', 1, NULL),

(2, '云计算研发工程师', '云平台部', '深圳', '软件开发', '云计算', '校招', '应届实习',
  '参与公司云平台核心组件的研发，包括容器编排、资源调度、监控告警等方向。使用 Go 语言开发高性能的云原生基础设施。',
  '1. 计算机相关专业本科及以上学历\n2. 熟练掌握 Go 或 C++ 语言\n3. 熟悉 Linux 操作系统，了解容器技术（Docker/Kubernetes）\n4. 有分布式系统开发经验优先\n5. 对云原生技术有热情，具备快速学习能力',
  '["Go","Kubernetes","Docker","Linux","分布式系统","容器编排"]',
  '["深圳"]',
  '[{"code":"CSIG","name":"CSIG 云与智慧产业事业群"},{"code":"TEG","name":"TEG 技术工程事业群"}]',
  'active',
  'local_mock', 'local_mock_2', NULL, NULL, JSON_OBJECT('source', 'local_mock', 'note', 'seed demo job'), 'manual_enriched', 1, NULL),

(3, 'AI 应用开发工程师', 'AI应用部', '北京', '算法与大数据', 'AI应用', '校招', '应届实习',
  '参与 AI 应用产品的后端开发，构建智能推荐、对话系统、RAG 等应用的 API 服务和工程化落地。',
  '1. 计算机相关专业本科及以上学历\n2. 熟练掌握 Go 或 Python 语言\n3. 熟悉 LLM API 调用（OpenAI/腾讯混元等）\n4. 了解向量数据库和 RAG 架构优先\n5. 有 Prompt 工程经验优先',
  '["Go","Python","LLM","RAG","向量数据库","Prompt工程"]',
  '["北京"]',
  '[{"code":"IEG","name":"IEG 互动娱乐事业群"}]',
  'active',
  'local_mock', 'local_mock_3', NULL, NULL, JSON_OBJECT('source', 'local_mock', 'note', 'seed demo job'), 'manual_enriched', 1, NULL),

(4, '数据开发工程师', '数据平台部', '上海', '算法与大数据', '数据开发', '校招', '应届实习',
  '参与公司大数据平台的开发与维护，建设实时和离线数据处理链路，支持业务数据分析与决策。',
  '1. 计算机或数学相关专业本科及以上学历\n2. 熟悉 SQL，了解 Hive/Spark/Flink 等大数据框架\n3. 掌握至少一门编程语言（Go/Python/Java）\n4. 有数据仓库建设经验优先\n5. 良好的数据敏感度和分析能力',
  '["SQL","Python","Spark","Flink","Hive","数据仓库"]',
  '["上海","深圳"]',
  '[{"code":"TEG","name":"TEG 技术工程事业群"}]',
  'active',
  'local_mock', 'local_mock_4', NULL, NULL, JSON_OBJECT('source', 'local_mock', 'note', 'seed demo job'), 'manual_enriched', 1, NULL),

(5, '测试开发工程师', '质量部', '广州', '软件开发', '测试开发', '校招', '应届实习',
  '负责核心业务的质量保障工作，开发自动化测试框架和工具，推动持续集成和持续交付。',
  '1. 计算机相关专业本科及以上学历\n2. 熟悉至少一门编程语言（Go/Python/Java）\n3. 了解自动化测试框架（如 Selenium、Pytest）\n4. 有 CI/CD 经验优先\n5. 细心、有责任心，善于发现和定位问题',
  '["Go","Python","自动化测试","CI/CD","Selenium","Pytest"]',
  '["广州"]',
  '[{"code":"TEG","name":"TEG 技术工程事业群"},{"code":"WXG","name":"WXG 微信事业群"}]',
  'active',
  'local_mock', 'local_mock_5', NULL, NULL, JSON_OBJECT('source', 'local_mock', 'note', 'seed demo job'), 'manual_enriched', 1, NULL),

(6, '产品经理', '产品部', '杭州', '产品策划与运营', '产品', '校招', '应届实习',
  '负责公司核心产品的规划与设计，深入理解用户需求，制定产品路线图，推动产品落地与迭代。',
  '1. 本科及以上学历，专业不限\n2. 对互联网产品有浓厚兴趣，有产品实习经验优先\n3. 良好的逻辑思维和数据分析能力\n4. 优秀的沟通能力和跨团队协作能力\n5. 有技术背景优先',
  '["产品设计","需求分析","数据分析","用户研究","Figma"]',
  '["杭州","深圳"]',
  '[{"code":"PCG","name":"PCG 平台与内容事业群"}]',
  'active',
  'local_mock', 'local_mock_6', NULL, NULL, JSON_OBJECT('source', 'local_mock', 'note', 'seed demo job'), 'manual_enriched', 1, NULL),

(7, '前端开发工程师', '前端组', '北京', '软件开发', '前端开发', '校招', '应届实习',
  '负责公司 Web 前端的开发与优化，使用 React/Vue 构建高质量的用户界面，推动前端工程化建设。',
  '1. 计算机相关专业本科及以上学历\n2. 熟练掌握 JavaScript/TypeScript，熟悉 React 或 Vue 框架\n3. 了解前端工程化工具（Webpack/Vite）\n4. 有移动端 H5 或小程序开发经验优先\n5. 对用户体验和代码质量有追求',
  '["JavaScript","TypeScript","React","Vue","CSS","前端工程化"]',
  '["北京","上海"]',
  '[{"code":"WXG","name":"WXG 微信事业群"},{"code":"PCG","name":"PCG 平台与内容事业群"}]',
  'active',
  'local_mock', 'local_mock_7', NULL, NULL, JSON_OBJECT('source', 'local_mock', 'note', 'seed demo job'), 'manual_enriched', 1, NULL),

(8, '算法工程师', 'AI算法部', '深圳', '算法与大数据', '算法', '校招', '应届实习',
  '参与推荐系统、NLP、搜索排序等方向的算法研发，将前沿 AI 技术落地到业务场景。',
  '1. 计算机或数学相关专业硕士及以上学历\n2. 熟悉机器学习基础理论，了解深度学习框架（PyTorch/TensorFlow）\n3. 有推荐系统或 NLP 项目经验优先\n4. 熟悉 Python，了解 C++ 优先\n5. 有顶会论文发表优先',
  '["机器学习","深度学习","PyTorch","NLP","推荐系统","Python"]',
  '["深圳","北京"]',
  '[{"code":"IEG","name":"IEG 互动娱乐事业群"},{"code":"TEG","name":"TEG 技术工程事业群"}]',
  'active',
  'local_mock', 'local_mock_8', NULL, NULL, JSON_OBJECT('source', 'local_mock', 'note', 'seed demo job'), 'manual_enriched', 1, NULL);

-- -----------------------------------------------------------
-- 3.4 种子推荐结果
-- 基于 Arthur V2 简历的推荐快照
-- -----------------------------------------------------------
INSERT INTO `recommendation_results` (`id`, `user_id`, `resume_version_id`, `result_json`) VALUES
(1, 1, 2,
  '{"recommended_jobs":[{"job_id":1,"title":"Go 后端开发工程师","city":"深圳","category":"后端开发","recommendation_reason":"简历中的 Go 语言、MySQL、Redis、Docker、Kubernetes 技能与岗位要求高度匹配","match_highlights":["熟练掌握 Go 语言，有并发编程经验","熟悉 MySQL 和 Redis，有实际项目使用经验","有 Docker 和 Kubernetes 使用经验，符合微服务架构要求"],"risk_points":["缺乏大规模系统设计经验","项目经历偏校园级别，缺少企业级项目深度"],"priority":"strong_recommend"},{"job_id":2,"title":"云计算研发工程师","city":"深圳","category":"云计算","recommendation_reason":"简历中的云原生实习经历和 Kubernetes Operator 开发经验与岗位高度相关","match_highlights":["有 Kubernetes Operator 开发经验","熟悉 Docker 和 Go 语言","了解分布式系统，有 etcd 使用经验"],"risk_points":["容器编排经验偏入门级别","分布式系统理论基础需要加强"],"priority":"recommend"},{"job_id":3,"title":"AI 应用开发工程师","city":"北京","category":"AI应用","recommendation_reason":"简历中有后端开发经验，可向 AI 应用方向拓展，但缺乏 LLM 和向量数据库经验","match_highlights":["有 Go 和 Python 编程基础","有 API 开发经验，可快速上手 LLM API 集成","有微服务架构经验"],"risk_points":["缺乏 LLM API 调用经验","不了解向量数据库和 RAG 架构","缺少 Prompt 工程实践经验"],"priority":"try"}],"total":3}');

-- -----------------------------------------------------------
-- 3.5 种子匹配分析
-- Arthur V2 简历 vs Go 后端开发工程师岗位（核心亮点 Demo）
-- -----------------------------------------------------------
INSERT INTO `match_analyses` (`id`, `user_id`, `resume_version_id`, `job_id`, `score`, `result_json`) VALUES
(1, 1, 2, 1, 78,
  '{"match_score":78,"screening_result":"likely_pass","evidence_chain":[{"jd_requirement":"熟练掌握 Go 语言，了解 Goroutine、Channel 等并发模型","resume_evidence":"使用 Go 语言编写 API 接口，开发 Kubernetes Operator 和分布式任务调度系统","match_status":"matched","detail":"简历中多次体现 Go 语言实战经验，包括 API 开发和 Operator 开发"},{"jd_requirement":"熟悉 MySQL、Redis 等存储系统","resume_evidence":"项目使用 MySQL 和 Redis，独立完成微服务开发","match_status":"matched","detail":"简历中体现了 MySQL 和 Redis 的实际使用"},{"jd_requirement":"了解微服务架构，有 Docker/Kubernetes 使用经验","resume_evidence":"参与 Kubernetes Operator 开发，使用 Docker 部署服务","match_status":"matched","detail":"有 Docker 和 Kubernetes 的实战经验"},{"jd_requirement":"良好的沟通能力和团队协作精神","resume_evidence":null,"match_status":"missing","detail":"简历中未体现团队协作和沟通能力的具体实例"}],"risk_points":[{"type":"skill_gap","description":"缺乏大规模系统设计经验，项目峰值 QPS 5000+ 偏小规模","severity":"medium"},{"type":"expression_risk","description":"校园项目描述较为笼统，部分缺少量化指标（如用户量、响应时间等）","severity":"low"}],"optimization_suggestions":[{"priority":"P0","suggestion":"补充团队协作和沟通能力的具体实例，如跨团队协作经验、技术分享等","target_field":"experience","reason":"岗位明确要求沟通能力和团队协作精神，简历中缺失"},{"priority":"P1","suggestion":"在项目描述中增加更多量化指标（如：系统响应时间、错误率、代码覆盖率等）","target_field":"project","reason":"当前部分项目描述缺乏具体技术指标"}],"next_steps":["根据 P0 建议补充团队协作相关内容","根据 P1 建议增加项目量化指标","添加新简历版本后再次匹配"],"authenticity_notice":"本分析基于您简历中的真实内容生成，不会为您虚构不存在的经历。优化建议旨在帮助您更好地呈现真实能力。"}');

-- -----------------------------------------------------------
-- 3.6 种子版本对比结果
-- Arthur V1 → V2 对比
-- -----------------------------------------------------------
INSERT INTO `version_compare_results` (`id`, `user_id`, `from_version_id`, `to_version_id`, `result_json`) VALUES
(1, 1, 1, 2,
  '{"from_version":{"version_no":1,"name":"Arthur"},"to_version":{"version_no":2,"name":"Arthur"},"score_change":{"from_score":65,"to_score":78,"change":"+13"},"field_changes":[{"field":"experience","change_type":"added","description":"新增了云计算公司的云原生开发实习经历，涉及 Kubernetes Operator 和集群管理"},{"field":"project","change_type":"modified","description":"校园交易平台增加了量化指标（日活 500+，订单 2000+ 笔），原有评测系统项目替换为分布式任务调度系统"},{"field":"skills","change_type":"modified","description":"新增 Kubernetes、etcd、gRPC 技能"}],"risk_changes":[{"risk":"缺乏云原生和分布式系统经验","from_severity":"high","to_severity":"low","description":"V2 新增了云原生实习经历和分布式任务调度项目，大幅缓解了此风险"},{"risk":"项目描述缺乏量化指标","from_severity":"medium","to_severity":"low","description":"V2 在校园交易平台项目中补充了用户量和订单量数据"},{"risk":"缺少团队协作相关内容","from_severity":"medium","to_severity":"medium","description":"V2 仍未补充团队协作的具体实例"}],"suggestion_completion":[{"suggestion":"补充云原生或分布式系统相关经验","priority":"P0","status":"completed","description":"V2 新增了云原生开发实习经历和分布式任务调度项目"},{"suggestion":"在项目描述中增加量化指标","priority":"P1","status":"partially_completed","description":"V2 在校园交易平台项目中增加了量化指标，但评测系统项目已被替换"}],"summary":"V2 相比 V1 有显著改进：新增了云原生实习经历和分布式任务调度项目，技术栈更加丰富，项目描述增加了量化指标。匹配度从 65 提升至 78。仍需补充团队协作相关内容。"}');

-- ============================================================
-- K8-5-32-C: Reviewer demo resumes for Offerr鹅
-- Account: 13900139000 / 123456, user_id = 2
-- ============================================================
-- K8-5-32-C: Reviewer demo resumes for Offerr鹅
-- Account: 13900139000 / 123456, user_id = 2
DELETE FROM `resume_versions` WHERE `user_id` = 2;

INSERT INTO `resume_versions` (
  `id`, `user_id`, `version_no`, `name`,
  `phone`, `email`, `location`,
  `target_city`, `target_role`,
  `education_json`, `experience_json`, `project_json`, `skills_json`,
  `schema_version`, `target_direction`, `job_type`, `expected_cities`,
  `school`, `degree`, `major`, `graduation_year`,
  `profile_json`, `job_intention_json`,
  `internship_json`, `certificates_json`, `awards_json`,
  `portfolio_json`, `extra_json`,
  `self_evaluation`
) VALUES
  (201, 2, 1, '01-算法研究生-岗位匹配', '13900010001', '[linzhiyuan.cs@example.com](mailto:linzhiyuan.cs@example.com)', '北京', '""', '""', '[{"school": "清华大学", "degree": "硕士", "major": "计算机科学与技术", "department": "计算机科学与技术系", "start_date": "2023-09", "end_date": "2026-06", "gpa": "3.78", "gpa_scale": "4.0", "rank": "前 15%", "core_courses": ["机器学习", "深度学习", "自然语言处理", "信息检索", "大规模数据挖掘", "高级算法设计"], "research_direction": "推荐系统、用户行为建模、表示学习", "lab": "智能信息处理与推荐系统实验室", "paper_summary": "参与推荐系统方向论文实验与消融分析，主要负责用户行为序列建模、负采样策略对比、离线指标复现实验和误差样例分析。目前论文仍处于实验补充和投稿准备阶段。"}, {"school": "北京邮电大学", "degree": "本科", "major": "软件工程", "department": "计算机学院", "start_date": "2019-09", "end_date": "2023-06", "gpa": "3.62", "gpa_scale": "4.0", "rank": "前 20%", "core_courses": ["数据结构", "操作系统", "数据库系统", "计算机网络", "概率论与数理统计", "软件工程"], "research_direction": "", "lab": "", "paper_summary": ""}]', '[{"company": "某头部内容平台", "role": "搜索推荐算法实习生", "start_date": "2024-07", "end_date": "2024-10", "work_content": "参与内容推荐场景的召回与排序实验，围绕用户短期兴趣变化、内容冷启动和负反馈样本处理进行实验分析。", "tech_stack": ["Python", "PyTorch", "Spark", "Hive", "SQL"], "achievements": ["完成用户行为特征统计与样本质量分析，定位 3 类异常曝光样本", "参与召回侧特征实验，对比不同时间窗口下的点击率变化", "整理实验结论并沉淀为组内周会汇报材料", "协助正式员工完成小流量实验前的数据核验", "对冷启动内容样本进行分层分析，输出低曝光内容的特征分布报告"], "metrics": "参与 8 组离线实验，输出 4 份实验分析文档，其中 2 项特征策略进入小流量验证。"}]', '[{"name": "多目标推荐排序模型实验平台", "role": "核心开发者", "type": "科研项目", "start_date": "2024-03", "end_date": "2024-11", "tech_stack": ["Python", "PyTorch", "Faiss", "Redis", "MySQL", "Docker", "Linux"], "description": "面向实验室推荐系统研究场景，搭建用于多目标排序模型实验、指标对比和实验记录管理的平台，支持召回结果导入、特征拼接、模型训练、离线评估和实验复现。", "responsibilities": "", "achievements": ["负责样本构建、特征处理和模型训练脚本开发", "实现 NDCG、Recall、AUC 等离线评估指标计算", "封装实验配置管理，减少不同模型之间的重复训练代码", "使用 Docker 统一实验环境，降低组内复现实验成本", "整理实验日志和指标对比表，辅助组内同学复盘模型效果变化", "将单次实验配置准备时间从约 40 分钟降低到 10 分钟以内", "支持 5 类排序模型在同一数据集上的离线指标对比", "完成 20 余组消融实验记录，为后续论文实验部分提供数据支持", "将实验结果记录格式统一为配置、指标、结论、问题四部分，减少复盘遗漏"], "link": ""}, {"name": "基于用户行为序列的点击率预估模型复现", "role": "算法实现与实验分析", "type": "课程与科研结合项目", "start_date": "2023-10", "end_date": "2024-01", "tech_stack": ["Python", "PyTorch", "Pandas", "NumPy", "Scikit-learn"], "description": "基于公开电商行为数据集，复现用户兴趣建模相关方法，对比传统 Wide&Deep 模型与序列兴趣建模方法在点击率预估任务中的表现差异。", "responsibilities": "", "achievements": ["完成数据清洗、用户行为序列截断和负样本构造", "实现 Embedding、Attention Pooling 和序列兴趣建模模块", "分析不同序列长度、负采样比例对模型效果的影响", "整理模型训练日志、指标变化和错误样本分析", "在验证集上相比基础 LR 模型 AUC 提升约 6.8%", "整理 12 页实验报告，说明模型结构、指标变化和误差样例", "形成一套可复用的 CTR 任务训练模板", "通过误差样例分析发现短序列用户和低频物品样本更容易出现预测偏差"], "link": "https://github.com/demo-user/ctr-sequence-model"}, {"name": "论文阅读与实验复现管理工具", "role": "个人开发者", "type": "个人工具项目", "start_date": "2024-12", "end_date": "2025-02", "tech_stack": ["Python", "FastAPI", "SQLite", "Vue", "Markdown"], "description": "为了解决论文阅读、实验配置和复现实验记录分散的问题，开发了一个轻量级工具，用于管理论文笔记、实验配置、指标结果和复盘结论。", "responsibilities": "", "achievements": ["设计论文、实验、指标、结论四类核心数据结构", "实现实验记录增删改查和 Markdown 导出", "支持按论文方向、模型类型、数据集进行筛选", "将常见实验问题整理为复盘清单，便于后续复现", "累计整理 30 余篇推荐系统和表示学习相关论文笔记", "管理 40 余条实验配置和指标结果", "帮助自己将论文阅读从零散记录转为结构化沉淀"], "link": "https://github.com/demo-user/paper-exp-tracker"}]', '["Python", "C++", "SQL", "Shell", "Go", "PyTorch", "Scikit-learn", "Pandas", "NumPy", "Faiss", "Redis", "MySQL", "Docker", "Linux", "Git", "Hive", "Spark 基础", "推荐系统离线评估", "实验复现", "数据分析"]', 2, '""', '""', '""', '清华大学', '硕士', '计算机科学与技术', '2026-06', '{"name": "林知远", "current_city": "北京", "school": "清华大学", "degree": "硕士", "major": "计算机科学与技术", "graduation_year": "2026-06"}', NULL, '[{"company": "某头部内容平台", "role": "搜索推荐算法实习生", "start_date": "2024-07", "end_date": "2024-10", "work_content": "参与内容推荐场景的召回与排序实验，围绕用户短期兴趣变化、内容冷启动和负反馈样本处理进行实验分析。", "tech_stack": ["Python", "PyTorch", "Spark", "Hive", "SQL"], "achievements": ["完成用户行为特征统计与样本质量分析，定位 3 类异常曝光样本", "参与召回侧特征实验，对比不同时间窗口下的点击率变化", "整理实验结论并沉淀为组内周会汇报材料", "协助正式员工完成小流量实验前的数据核验", "对冷启动内容样本进行分层分析，输出低曝光内容的特征分布报告"], "metrics": "参与 8 组离线实验，输出 4 份实验分析文档，其中 2 项特征策略进入小流量验证。"}]', '[{"name": "CET-6", "type": "language_exam", "level": "CET-6", "score": "568", "issued_at": ""}, {"name": "阿里云机器学习平台 PAI 基础课程认证", "type": "云计算与机器学习平台认证", "level": "课程认证", "score": "通过", "issued_at": "2024-05"}]', '[{"name": "全国大学生数学建模竞赛省级二等奖", "type": "学科竞赛", "level": "省级二等奖", "awarded_at": "2021-11", "description": "负责数据建模、算法实现和结果可视化部分。"}]', '{"works": [{"name": "CTR 序列建模复现实验项目", "link": "https://github.com/demo-user/ctr-sequence-model", "password": ""}, {"name": "推荐排序实验记录模板", "link": "https://github.com/demo-user/recsys-exp-template", "password": ""}, {"name": "论文阅读与实验复现管理工具", "link": "https://github.com/demo-user/paper-exp-tracker", "password": ""}], "homepage": "https://github.com/demo-user"}', '{"self_evaluation": "", "additional_info": ""}', ''),
  (202, 2, 2, '02-211后台开发-岗位匹配', '13900010002', '[zhoumingyuan.dev@example.com](mailto:zhoumingyuan.dev@example.com)', '武汉', '""', '""', '[{"school": "武汉理工大学", "degree": "本科", "major": "软件工程", "department": "计算机与人工智能学院", "start_date": "2022-09", "end_date": "2026-06", "gpa": "3.42", "gpa_scale": "4.0", "rank": "前 35%", "core_courses": ["数据结构", "操作系统", "计算机网络", "数据库系统", "Java 程序设计", "软件工程", "分布式系统导论"], "research_direction": "后端开发、数据库设计、缓存与接口性能优化", "lab": "", "paper_summary": ""}]', '[{"company": "武汉某软件科技有限公司", "role": "后端开发实习生", "start_date": "2025-03", "end_date": "2025-05", "work_content": "参与企业内部管理系统的后端接口开发，主要负责基础信息维护、列表查询、权限字段校验和部分统计接口。", "tech_stack": ["Java", "Spring Boot", "MySQL", "Redis", "MyBatis"], "achievements": ["参与员工信息、部门信息、审批记录等模块接口开发", "完成 12 个列表查询和详情接口，支持分页、条件筛选和状态过滤", "根据测试反馈修复参数为空、分页异常、状态枚举不一致等问题", "配合前端完成接口联调，补充 Swagger 注释和字段说明", "对部分高频查询增加索引，减少慢查询情况"], "metrics": "参与开发和维护 20 余个接口，修复测试阶段反馈问题 30 余项，负责模块在内部测试环境完成验收。"}]', '[{"name": "校园二手交易平台后端系统", "role": "后端负责人", "type": "课程综合项目 / 校园应用项目", "start_date": "2024-03", "end_date": "2024-07", "tech_stack": ["Go", "Gin", "MySQL", "Redis", "JWT", "Docker", "Nginx"], "description": "面向校内学生的二手物品发布与交易平台，支持用户注册登录、商品发布、商品搜索、收藏、留言、订单状态流转和后台基础管理。", "responsibilities": "", "achievements": ["负责用户、商品、订单、收藏、留言等核心模块的后端接口设计与开发", "使用 JWT 实现登录鉴权，并对商品发布、收藏、订单操作等接口增加登录态校验", "设计 MySQL 表结构和索引，完成商品列表、关键词搜索和用户商品管理接口", "使用 Redis 缓存热门商品列表和商品详情，减少重复查询", "编写 Dockerfile 和 docker-compose 配置，完成本地一键启动部署", "完成 30 余个后端接口，覆盖用户、商品、订单、收藏、留言等核心流程", "商品列表接口在加入 Redis 缓存后，重复访问平均响应时间由约 180ms 降低到 60ms 左右", "通过接口参数校验和权限校验，修复了未登录用户发布商品、越权修改商品状态等问题", "项目最终完成前后端联调，并在班级课程展示中作为小组 Demo 演示"], "link": "https://github.com/demo-user/campus-trade-api"}, {"name": "在线题库与练习系统", "role": "后端开发", "type": "个人练习项目", "start_date": "2024-08", "end_date": "2024-11", "tech_stack": ["Java", "Spring Boot", "MySQL", "Redis", "MyBatis", "Swagger"], "description": "面向学生刷题练习场景，开发在线题库系统，支持题目管理、分类筛选、答题记录、错题本、练习统计和接口文档展示。", "responsibilities": "", "achievements": ["设计题目、分类、答题记录、错题本等数据表", "开发题目分页查询、条件筛选、提交答案、错题记录等接口", "使用 Swagger 维护接口文档，方便前端同学联调", "对题目列表查询增加分页和索引优化，避免一次性返回过多数据", "使用 Redis 记录用户近期练习统计，减少频繁聚合查询", "完成题库、练习记录、错题本 3 个核心模块", "将题目列表查询从全量查询改为分页查询，避免大数据量下接口响应变慢", "整理接口文档和测试用例，减少前后端字段理解不一致的问题", "通过该项目系统学习了 Controller、Service、Mapper 分层和常见后端接口开发流程"], "link": "https://github.com/demo-user/online-question-bank"}, {"name": "简易消息通知服务", "role": "后端开发", "type": "个人技术练习项目", "start_date": "2025-01", "end_date": "2025-02", "tech_stack": ["Go", "RabbitMQ", "MySQL", "Redis"], "description": "为了理解消息队列在后端系统中的作用，开发一个简易消息通知服务，模拟订单状态变更后异步发送站内通知的流程。", "responsibilities": "", "achievements": ["设计通知表和消息投递状态字段", "使用 RabbitMQ 实现订单事件生产和通知消费", "增加消费失败重试和幂等处理，避免重复写入通知", "使用 Redis 简单记录用户未读通知数", "编写接口测试脚本验证消息发送、消费和失败重试流程", "跑通订单事件到站内通知的异步处理流程", "理解了消息队列削峰、解耦和最终一致性的基本使用场景", "通过唯一业务键避免重复消费导致的重复通知问题"], "link": ""}]', '["Go", "Java", "SQL", "Gin", "Spring Boot", "MySQL", "Redis", "MyBatis", "JWT", "RabbitMQ", "Docker", "Nginx", "Linux 基础", "Git", "Swagger", "RESTful API 设计", "基础接口性能优化"]', 2, '""', '""', '""', '武汉理工大学', '本科', '软件工程', '2026-06', '{"name": "周明远", "current_city": "武汉", "school": "武汉理工大学", "degree": "本科", "major": "软件工程", "graduation_year": "2026-06"}', NULL, '[{"company": "武汉某软件科技有限公司", "role": "后端开发实习生", "start_date": "2025-03", "end_date": "2025-05", "work_content": "参与企业内部管理系统的后端接口开发，主要负责基础信息维护、列表查询、权限字段校验和部分统计接口。", "tech_stack": ["Java", "Spring Boot", "MySQL", "Redis", "MyBatis"], "achievements": ["参与员工信息、部门信息、审批记录等模块接口开发", "完成 12 个列表查询和详情接口，支持分页、条件筛选和状态过滤", "根据测试反馈修复参数为空、分页异常、状态枚举不一致等问题", "配合前端完成接口联调，补充 Swagger 注释和字段说明", "对部分高频查询增加索引，减少慢查询情况"], "metrics": "参与开发和维护 20 余个接口，修复测试阶段反馈问题 30 余项，负责模块在内部测试环境完成验收。"}]', '[{"name": "CET-4", "type": "language_exam", "level": "CET-4", "score": "512", "issued_at": ""}, {"name": "全国计算机等级考试二级", "type": "计算机基础认证", "level": "二级", "score": "通过", "issued_at": "2023-09"}]', '[{"name": "优秀学生干部", "type": "综合荣誉", "level": "学院级", "awarded_at": "2024-06", "description": "担任班级学习委员，参与课程资料整理和班级学习打卡组织。"}]', '{"works": [{"name": "校园二手交易平台后端系统", "link": "https://github.com/demo-user/campus-trade-api", "password": ""}, {"name": "在线题库与练习系统", "link": "https://github.com/demo-user/online-question-bank", "password": ""}], "homepage": "https://github.com/demo-user"}', '{"self_evaluation": "", "additional_info": ""}', ''),
  (203, 2, 3, '03-二本产品经理-岗位匹配', '13900010003', '[chensiyu.pm@example.com](mailto:chensiyu.pm@example.com)', '杭州', '""', '""', '[{"school": "浙江万里学院", "degree": "本科", "major": "信息管理与信息系统", "department": "商学院 / 信息管理系", "start_date": "2022-09", "end_date": "2026-06", "gpa": "3.36", "gpa_scale": "4.0", "rank": "前 30%", "core_courses": ["管理信息系统", "数据库原理", "用户体验设计", "电子商务概论", "统计学", "市场调研", "数据分析基础", "项目管理"], "research_direction": "产品需求分析、用户研究、业务流程设计", "lab": "", "paper_summary": ""}]', '[{"company": "杭州某教育科技公司", "role": "产品运营实习生", "start_date": "2025-03", "end_date": "2025-05", "work_content": "参与在线课程后台和学习打卡工具的日常运营支持，协助产品经理整理用户反馈、跟进小需求、测试功能改动和维护运营数据表。", "tech_stack": ["飞书", "Excel", "墨刀", "禅道", "SQL 基础"], "achievements": ["每周整理用户反馈和客服问题，按登录、课程播放、作业提交、打卡提醒等类型进行分类", "协助产品经理补充 6 个小需求说明，包括字段规则、异常提示和后台筛选条件", "参与课程后台筛选功能测试，记录并跟进 17 个问题", "维护学习打卡数据表，统计用户打卡次数、连续打卡天数和中断原因", "根据用户反馈整理“提醒时间不明显”“作业入口不清晰”等问题，协助优化页面文案"], "metrics": "累计整理 230 余条用户反馈，输出 5 份周反馈汇总，协助跟进 6 个小需求上线前验收。"}]', '[{"name": "校园活动报名与签到小程序", "role": "产品负责人", "type": "校园实践项目", "start_date": "2024-03", "end_date": "2024-06", "tech_stack": ["Axure", "墨刀", "Excel", "问卷星", "微信小程序", "飞书文档"], "description": "面向校内社团和学院活动组织场景，设计一个活动报名、名单管理、签到统计和活动反馈收集的小程序原型，解决原本依赖微信群接龙、表格统计和人工签到导致的信息混乱问题。", "responsibilities": "", "achievements": ["访谈 8 名社团负责人和 20 余名学生，梳理活动发布、报名、签到、反馈四个核心流程", "输出用户角色、业务流程图、功能结构图和核心页面原型", "设计活动创建、报名表单、二维码签到、报名名单导出、活动反馈统计等功能", "编写 18 页产品需求文档，说明功能规则、字段说明、异常状态和数据统计口径", "组织 2 轮同学试用，根据反馈调整报名表单字段和签到入口位置", "将活动报名信息从微信群接龙改为结构化表单，减少重复统计和漏报名问题", "原型试用后，活动负责人整理报名名单的时间从约 30 分钟降低到 10 分钟以内", "通过反馈问卷收集 63 份学生意见，定位“报名后不知道是否成功”“签到入口难找”等 5 类问题", "项目在课程答辩中获得小组优秀展示评价"], "link": ""}, {"name": "二手教材交易需求调研与产品原型设计", "role": "产品设计与调研负责人", "type": "课程项目", "start_date": "2024-09", "end_date": "2024-12", "tech_stack": ["Axure", "Figma", "Excel", "XMind", "问卷星"], "description": "围绕大学生二手教材交易场景，完成从需求调研、竞品分析、用户分层到产品原型设计的完整流程，重点解决教材信息不透明、交易沟通成本高和教材版本不一致的问题。", "responsibilities": "", "achievements": ["设计并回收 146 份问卷，分析学生购买二手教材的主要顾虑和交易习惯", "体验闲鱼、孔夫子旧书网、校内微信群等渠道，整理交易流程和信息展示差异", "将用户划分为卖书学生、买书学生、班级教材负责人三类角色", "设计教材发布、教材搜索、版本核对、在线留言、交易状态流转等核心功能", "输出竞品分析表、用户旅程图、功能优先级表和高保真原型", "调研发现 72% 的学生最担心“教材版本不一致”，因此在原型中增加 ISBN / 课程名 / 任课老师三类核对字段", "将需求拆分为 12 个核心功能和 9 个次要功能，明确 MVP 优先实现范围", "形成 25 页产品需求说明文档和 16 个核心页面原型", "课程评分 91 分，获得任课老师关于调研完整性的正向评价"], "link": ""}, {"name": "学习打卡社区产品改版分析", "role": "产品分析与数据整理", "type": "个人分析项目", "start_date": "2025-01", "end_date": "2025-02", "tech_stack": ["Excel", "SQL 基础", "Figma", "飞书文档"], "description": "选择 3 款学习打卡类产品进行体验分析，围绕新用户引导、打卡链路、社群互动和激励机制进行拆解，并提出改版建议。", "responsibilities": "", "achievements": ["拆解 3 款产品的新用户注册、目标创建、每日打卡和复盘流程", "对打卡中断、提醒频率、内容沉淀和好友激励机制进行对比分析", "使用 Excel 整理功能矩阵、体验问题和改版优先级", "设计“目标创建简化”“打卡后复盘提示”“连续打卡反馈”三个改版原型", "输出 11 页产品体验分析报告，包含功能拆解、问题归因和改版建议", "梳理 18 个体验问题，其中 6 个归类为新用户初次使用阻力", "形成 3 个低保真改版原型，用于表达优化思路"], "link": ""}]', '["Axure", "Figma", "墨刀", "XMind", "Excel", "飞书文档", "问卷星", "SQL 基础", "产品需求文档", "用户访谈", "竞品分析", "流程图", "原型设计", "数据整理", "功能优先级拆分", "项目复盘"]', 2, '""', '""', '""', '浙江万里学院', '本科', '信息管理与信息系统', '2026-06', '{"name": "陈思雨", "current_city": "杭州", "school": "浙江万里学院", "degree": "本科", "major": "信息管理与信息系统", "graduation_year": "2026-06"}', NULL, '[{"company": "杭州某教育科技公司", "role": "产品运营实习生", "start_date": "2025-03", "end_date": "2025-05", "work_content": "参与在线课程后台和学习打卡工具的日常运营支持，协助产品经理整理用户反馈、跟进小需求、测试功能改动和维护运营数据表。", "tech_stack": ["飞书", "Excel", "墨刀", "禅道", "SQL 基础"], "achievements": ["每周整理用户反馈和客服问题，按登录、课程播放、作业提交、打卡提醒等类型进行分类", "协助产品经理补充 6 个小需求说明，包括字段规则、异常提示和后台筛选条件", "参与课程后台筛选功能测试，记录并跟进 17 个问题", "维护学习打卡数据表，统计用户打卡次数、连续打卡天数和中断原因", "根据用户反馈整理“提醒时间不明显”“作业入口不清晰”等问题，协助优化页面文案"], "metrics": "累计整理 230 余条用户反馈，输出 5 份周反馈汇总，协助跟进 6 个小需求上线前验收。"}]', '[{"name": "CET-4", "type": "language_exam", "level": "CET-4", "score": "489", "issued_at": ""}, {"name": "Axure 产品原型设计课程证书", "type": "产品工具课程认证", "level": "课程认证", "score": "通过", "issued_at": "2024-08"}]', '[{"name": "商学院优秀课程项目展示", "type": "课程项目荣誉", "level": "学院级", "awarded_at": "2024-12", "description": "二手教材交易项目在课程展示中被评为优秀小组案例。"}]', '{"works": [{"name": "校园活动报名与签到小程序原型", "link": "https://demo.example.com/activity-signup-prototype", "password": ""}, {"name": "二手教材交易产品需求文档", "link": "https://demo.example.com/textbook-trade-prd", "password": ""}, {"name": "学习打卡社区产品体验分析报告", "link": "https://demo.example.com/study-checkin-analysis", "password": ""}], "homepage": ""}', '{"self_evaluation": "", "additional_info": ""}', ''),
  (204, 2, 4, '04-职业探索型学生-岗位推荐', '13900010004', '[xuanran.life@example.com](mailto:xuanran.life@example.com)', '成都', '""', '""', '[{"school": "四川师范大学", "degree": "本科", "major": "教育技术学", "department": "教育科学学院", "start_date": "2022-09", "end_date": "2026-06", "gpa": "3.48", "gpa_scale": "4.0", "rank": "前 30%", "core_courses": ["教育心理学", "学习科学", "教学设计", "教育数据分析", "用户体验基础", "数字媒体技术", "课程与教学论", "统计学基础"], "research_direction": "学习体验设计、数字化学习、学生成长支持", "lab": "数字学习与教育技术创新工作室", "paper_summary": "参与学院数字学习体验调研项目，主要协助完成学生访谈、问卷整理和学习平台使用问题归纳，关注学生在在线学习过程中的动机变化、任务拖延和反馈获得感问题。"}]', '[{"company": "成都某在线教育公司", "role": "用户运营实习生", "start_date": "2025-03", "end_date": "2025-05", "work_content": "参与在线学习社群和训练营运营，协助完成学员提醒、学习数据整理、社群活动策划和用户反馈收集。", "tech_stack": ["企业微信", "飞书", "Excel", "问卷星", "Canva"], "achievements": ["每周整理学员学习进度表，标记未完成任务、连续缺勤和活跃学员", "协助运营 3 个学习社群，负责每日提醒、问题收集和活动通知", "参与设计学习打卡活动和阶段复盘模板", "整理学员对课程难度、作业量和直播时间的反馈", "根据反馈优化社群提醒文案，减少机械催促感"], "metrics": "累计维护约 450 名学员的学习进度数据，整理 6 份周运营复盘，协助完成 2 次学习活动策划。"}]', '[{"name": "大学生自习与学习陪伴小程序调研项目", "role": "调研与产品原型负责人", "type": "校园实践项目", "start_date": "2024-03", "end_date": "2024-06", "tech_stack": ["问卷星", "飞书文档", "Excel", "XMind", "墨刀", "Figma"], "description": "围绕大学生自习效率低、缺少陪伴感和难以坚持学习计划的问题，完成用户调研、需求分析和小程序原型设计，探索“轻量打卡 + 学习搭子 + 阶段复盘”的学习支持方式。", "responsibilities": "", "achievements": ["设计并回收 178 份问卷，分析学生自习频率、拖延原因和学习计划中断原因", "访谈 12 名不同年级学生，整理“缺少反馈”“任务太大”“没人一起坚持”等典型问题", "输出用户画像、使用场景、核心流程和功能优先级表", "设计学习目标创建、学习搭子匹配、每日打卡、阶段复盘和轻量提醒等页面原型", "组织 2 次小范围试用，根据反馈调整打卡入口和复盘提示文案", "调研发现超过 60% 的受访学生会制定学习计划，但连续执行超过 7 天的比例较低", "将原本复杂的学习计划拆成“今日任务 + 本周目标 + 阶段复盘”三层结构", "输出 20 页调研分析文档和 14 个核心页面原型", "试用反馈中，学生对“学习搭子提醒”和“阶段复盘记录”接受度较高"], "link": "https://demo.example.com/study-companion-prototype"}, {"name": "校园生活方式社群运营计划", "role": "社群运营负责人", "type": "校园社团项目", "start_date": "2023-10", "end_date": "2024-05", "tech_stack": ["微信公众号", "飞书", "Excel", "Canva", "剪映", "问卷星"], "description": "参与运营一个面向学生的校园生活方式社群，内容包括阅读、运动、摄影、城市探索和学习复盘，目标是帮助同学形成更稳定的生活节奏和兴趣交流氛围。", "responsibilities": "", "achievements": ["规划每月主题活动，如“21 天阅读记录”“周末城市散步”“运动打卡”“期末复盘夜谈”", "负责活动文案、报名表、社群规则和活动复盘整理", "维护社群日常互动，收集同学对活动时间、形式和内容主题的反馈", "使用 Excel 统计报名人数、参与率、打卡完成率和反馈关键词", "参与公众号推文选题、图片排版和活动总结发布", "累计组织 9 场线上 / 线下活动，覆盖 300 余名学生", "“21 天阅读记录”活动报名 86 人，完整完成打卡 37 人", "通过活动复盘发现，明确的阶段目标和轻量反馈比强制打卡更容易被学生接受", "公众号活动总结平均阅读量约 600，最高单篇阅读量超过 1200"], "link": ""}, {"name": "校园新生适应问题访谈与内容手册设计", "role": "访谈整理与内容策划", "type": "学院学生工作项目", "start_date": "2024-09", "end_date": "2024-11", "tech_stack": ["飞书文档", "Excel", "Canva", "问卷星"], "description": "协助学院老师和学生会完成新生适应问题调研，围绕课程学习、宿舍关系、社团选择、时间管理和情绪压力等主题，整理一份面向新生的适应手册。", "responsibilities": "", "achievements": ["协助设计新生适应问卷和访谈提纲", "整理 20 份访谈记录，归纳高频问题和典型困惑", "将问题分为学习适应、生活适应、人际适应、信息获取四类", "参与手册内容策划，撰写“如何安排第一学期”“如何选择社团”“如何寻求帮助”等章节", "协助完成手册排版和线上发布", "整理出 32 个新生高频问题，并形成对应建议", "完成约 1.5 万字的新生适应手册初稿", "手册在线发布后被学院新生群置顶使用", "通过该项目提升了访谈整理、内容组织和学生需求理解能力"], "link": ""}]', '["问卷设计", "用户访谈", "活动策划", "社群运营", "内容策划", "产品原型基础", "学习体验分析", "Excel 数据整理", "飞书文档", "XMind", "墨刀", "Figma 基础", "Canva", "剪映", "微信公众号排版", "活动复盘"]', 2, '""', '""', '""', '四川师范大学', '本科', '教育技术学', '2026-06', '{"name": "许安然", "current_city": "成都", "school": "四川师范大学", "degree": "本科", "major": "教育技术学", "graduation_year": "2026-06"}', NULL, '[{"company": "成都某在线教育公司", "role": "用户运营实习生", "start_date": "2025-03", "end_date": "2025-05", "work_content": "参与在线学习社群和训练营运营，协助完成学员提醒、学习数据整理、社群活动策划和用户反馈收集。", "tech_stack": ["企业微信", "飞书", "Excel", "问卷星", "Canva"], "achievements": ["每周整理学员学习进度表，标记未完成任务、连续缺勤和活跃学员", "协助运营 3 个学习社群，负责每日提醒、问题收集和活动通知", "参与设计学习打卡活动和阶段复盘模板", "整理学员对课程难度、作业量和直播时间的反馈", "根据反馈优化社群提醒文案，减少机械催促感"], "metrics": "累计维护约 450 名学员的学习进度数据，整理 6 份周运营复盘，协助完成 2 次学习活动策划。"}]', '[{"name": "CET-4", "type": "language_exam", "level": "CET-4", "score": "476", "issued_at": ""}, {"name": "心理健康教育基础课程证书", "type": "课程学习证书", "level": "课程认证", "score": "通过", "issued_at": "2024-11"}]', '[{"name": "校园生活方式社群优秀策划", "type": "活动策划荣誉", "level": "社团内部评选", "awarded_at": "2024-05", "description": "参与策划阅读打卡和城市探索活动，活动参与度和反馈较好。"}]', '{"works": [{"name": "大学生自习与学习陪伴小程序原型", "link": "https://demo.example.com/study-companion-prototype", "password": ""}, {"name": "新生适应问题访谈整理与内容手册", "link": "https://demo.example.com/freshman-guide", "password": ""}, {"name": "校园生活方式社群活动复盘合集", "link": "https://demo.example.com/campus-life-review", "password": ""}], "homepage": ""}', '{"self_evaluation": "", "additional_info": ""}', ''),
  (205, 2, 5, '05-优秀但不适合互联网-岗位推荐', '13900010005', '[guqinghe.research@example.com](mailto:guqinghe.research@example.com)', '南京', '""', '""', '[{"school": "南京大学", "degree": "本科", "major": "社会学", "department": "社会学院", "start_date": "2022-09", "end_date": "2026-06", "gpa": "3.86", "gpa_scale": "4.0", "rank": "前 10%", "core_courses": ["社会学概论", "社会研究方法", "社会统计学", "组织社会学", "教育社会学", "城市社会学", "公共政策分析", "质性研究方法"], "research_direction": "青年发展、教育公平、基层治理、公共服务", "lab": "社会调查与公共政策研究工作坊", "paper_summary": "参与一项关于县域高中学生升学选择与家庭支持的调研项目，主要负责访谈提纲设计、访谈资料整理、编码分析和阶段性报告撰写。个人更关注长期社会问题、公共服务和组织运行机制。"}]', '[{"company": "某公益基金会教育项目组", "role": "项目助理实习生", "start_date": "2025-02", "end_date": "2025-05", "work_content": "参与县域高中生成长支持项目，协助完成学生需求调研、活动执行、材料整理和项目复盘。", "tech_stack": ["飞书", "Excel", "问卷星", "腾讯会议", "微信公众号"], "achievements": ["协助整理学生报名信息、课程参与记录和活动反馈表", "参与 3 场线上成长主题活动执行，负责会前通知、会中记录和会后反馈整理", "整理学生关于升学信息、时间管理、情绪压力和家庭沟通的反馈", "协助撰写项目月度简报，归纳活动参与情况和典型问题", "与志愿者团队协作，跟进学生个案反馈和活动改进建议"], "metrics": "参与维护约 180 名学生的项目参与记录，整理 400 余条反馈信息，协助完成 3 份月度项目简报和 1 份阶段复盘材料。"}]', '[{"name": "县域高中学生升学选择与家庭支持调研", "role": "调研执行与报告撰写", "type": "科研训练项目", "start_date": "2024-04", "end_date": "2024-11", "tech_stack": ["问卷星", "Excel", "SPSS", "NVivo", "飞书文档"], "description": "围绕县域高中学生在升学选择、家庭支持、信息获取和心理压力方面的问题，完成问卷调研、深度访谈和分析报告撰写，为后续教育支持项目提供参考。", "responsibilities": "", "achievements": ["参与问卷设计和访谈提纲设计，关注家庭支持、学校资源、升学信息差等变量", "整理 312 份有效问卷，完成基础描述统计和交叉分析", "参与 26 名学生、家长和教师的半结构化访谈", "使用 NVivo 对访谈文本进行初步编码，归纳学生在信息获取、情绪支持和目标选择上的典型问题", "参与撰写阶段性调研报告，负责“学生信息获取路径”和“家庭支持差异”两个章节", "整理出 5 类升学信息差来源，包括学校资源差异、家庭经验不足、同伴信息偏差、线上信息过载和咨询渠道缺失", "输出约 2.4 万字阶段性调研报告，为公益组织后续设计高中生成长支持活动提供参考", "将访谈材料整理为 18 个典型案例，帮助团队更直观理解学生困境", "通过该项目提升了长期调研、访谈整理和公共议题分析能力"], "link": ""}, {"name": "社区老年人数字服务使用障碍研究", "role": "项目协调与访谈整理", "type": "社会实践项目", "start_date": "2023-10", "end_date": "2024-02", "tech_stack": ["Excel", "飞书文档", "录音转写工具", "Canva"], "description": "围绕社区老年人在使用挂号、缴费、出行、政务服务等数字化工具时遇到的障碍，开展访谈和观察记录，形成面向社区服务站的建议材料。", "responsibilities": "", "achievements": ["协助联系社区工作人员，安排访谈时间和调研流程", "参与 18 名老年居民和 5 名社区工作人员访谈", "整理老年人在健康码、线上挂号、公交乘车码、政务小程序使用过程中的困难", "将问题归类为操作步骤复杂、信息提示不清、子女协助依赖、线下兜底不足四类", "参与制作社区数字服务协助卡片和简易操作说明", "输出 12 页调研简报，提交给社区服务站作为数字服务志愿活动参考", "协助设计 4 张常用服务操作说明卡片，用于社区志愿者现场指导", "发现“不会操作”背后往往是记忆负担、担心出错和缺少即时帮助，而不仅是技术能力不足", "项目被学院社会实践总结会选为展示案例"], "link": ""}, {"name": "高校学生阅读支持计划", "role": "内容策划与活动组织", "type": "校园公益项目", "start_date": "2024-03", "end_date": "2024-06", "tech_stack": ["微信公众号", "飞书", "Excel", "问卷星", "Canva"], "description": "参与组织一个面向本科生的阅读支持计划，通过主题书单、共读小组、读书会和阶段复盘，帮助同学建立更稳定的阅读习惯。", "responsibilities": "", "achievements": ["参与设计 4 期主题书单，包括社会科学入门、城市与生活、教育与成长、女性与社会", "负责报名表、活动说明、共读规则和复盘问题设计", "维护共读小组日常交流，整理成员阅读反馈和讨论问题", "撰写读书会总结推文和阶段复盘材料", "通过问卷收集参与者对活动节奏、书目难度和讨论形式的反馈", "4 期活动累计报名 126 人，稳定参与 48 人", "公众号活动总结平均阅读量约 800，最高单篇超过 1500", "通过反馈发现，相比强制打卡，开放讨论和阶段复盘更能提高参与者持续性", "形成一套活动流程模板，用于后续学生社团继续组织类似活动"], "link": ""}]', '["社会调研", "用户访谈", "问卷设计", "质性资料编码", "公共政策分析", "报告写作", "活动策划", "内容策划", "Excel", "SPSS 基础", "NVivo 基础", "飞书文档", "问卷星", "微信公众号排版", "Canva"]', 2, '""', '""', '""', '南京大学', '本科', '社会学', '2026-06', '{"name": "顾清和", "current_city": "南京", "school": "南京大学", "degree": "本科", "major": "社会学", "graduation_year": "2026-06"}', NULL, '[{"company": "某公益基金会教育项目组", "role": "项目助理实习生", "start_date": "2025-02", "end_date": "2025-05", "work_content": "参与县域高中生成长支持项目，协助完成学生需求调研、活动执行、材料整理和项目复盘。", "tech_stack": ["飞书", "Excel", "问卷星", "腾讯会议", "微信公众号"], "achievements": ["协助整理学生报名信息、课程参与记录和活动反馈表", "参与 3 场线上成长主题活动执行，负责会前通知、会中记录和会后反馈整理", "整理学生关于升学信息、时间管理、情绪压力和家庭沟通的反馈", "协助撰写项目月度简报，归纳活动参与情况和典型问题", "与志愿者团队协作，跟进学生个案反馈和活动改进建议"], "metrics": "参与维护约 180 名学生的项目参与记录，整理 400 余条反馈信息，协助完成 3 份月度项目简报和 1 份阶段复盘材料。"}]', '[{"name": "CET-6", "type": "language_exam", "level": "CET-6", "score": "592", "issued_at": ""}, {"name": "社会调查方法训练营结业证书", "type": "研究方法课程认证", "level": "结业", "score": "优秀学员", "issued_at": "2024-08"}]', '[{"name": "学院优秀学生", "type": "综合荣誉", "level": "学院级", "awarded_at": "2023-12", "description": "学业成绩稳定，参与学院调研项目和学生公益活动较多。"}]', '{"works": [{"name": "县域高中学生升学选择与家庭支持调研报告", "link": "https://demo.example.com/county-high-school-report", "password": ""}, {"name": "社区老年人数字服务使用障碍调研简报", "link": "https://demo.example.com/elder-digital-service-report", "password": ""}, {"name": "高校学生阅读支持计划活动复盘", "link": "https://demo.example.com/reading-support-review", "password": ""}], "homepage": ""}', '{"self_evaluation": "", "additional_info": ""}', ''),
  (206, 2, 6, '06-211后台开发-版本对比A', '13900010006', '[liuziheng.backend@example.com](mailto:liuziheng.backend@example.com)', '合肥', '""', '""', '[{"school": "合肥工业大学", "degree": "本科", "major": "计算机科学与技术", "department": "计算机与信息学院", "start_date": "2022-09", "end_date": "2026-06", "gpa": "3.31", "gpa_scale": "4.0", "rank": "前 40%", "core_courses": ["数据结构", "操作系统", "计算机网络", "数据库系统", "Java 程序设计", "软件工程", "Web 应用开发"], "research_direction": "后端开发、数据库应用、Web 系统开发", "lab": "", "paper_summary": ""}]', '[{"company": "安徽某信息技术有限公司", "role": "后端开发实习生", "start_date": "2025-03", "end_date": "2025-05", "work_content": "参与公司内部管理系统的部分接口开发和维护，主要涉及基础数据管理、列表查询和简单统计功能。", "tech_stack": ["Java", "Spring Boot", "MySQL", "MyBatis"], "achievements": ["参与基础信息管理模块接口开发", "完成部分列表查询、详情查询和新增修改接口", "根据测试反馈修复字段为空、状态显示不正确等问题", "配合前端完成部分页面联调", "学习了企业项目中的分层结构和接口规范"], "metrics": "参与开发和维护 10 余个接口，修复测试反馈问题 20 余项。"}]', '[{"name": "校园课程资料管理系统", "role": "后端开发", "type": "课程项目", "start_date": "2024-03", "end_date": "2024-06", "tech_stack": ["Java", "Spring Boot", "MySQL", "MyBatis", "Redis"], "description": "该项目主要用于管理课程资料，支持用户登录、资料上传记录、资料分类、资料列表查询和后台管理等功能。", "responsibilities": "", "achievements": ["负责用户登录和课程资料管理相关接口开发", "设计课程、资料、用户等基础数据表", "完成资料列表查询、详情查看、分类筛选等功能", "使用 Redis 做部分缓存，提高查询效率", "配合前端完成接口联调", "完成了课程资料管理相关接口开发", "项目可以正常完成登录、查询、分类筛选等流程", "通过该项目熟悉了 Spring Boot 后端开发流程", "对 MySQL 表设计和接口开发有了更完整的理解"], "link": "https://github.com/demo-user/course-resource-system"}, {"name": "校园跑腿服务小程序后端", "role": "后端开发", "type": "小组项目", "start_date": "2024-09", "end_date": "2024-12", "tech_stack": ["Go", "Gin", "MySQL", "Redis", "JWT"], "description": "该项目面向校内学生跑腿服务场景，支持用户发布跑腿需求、接单、订单状态更新和个人订单查询。", "responsibilities": "", "achievements": ["参与订单模块和用户模块接口开发", "使用 JWT 实现简单登录鉴权", "设计订单表、用户表和状态字段", "完成订单创建、订单查询、订单状态修改等接口", "参与接口测试和前后端联调", "完成订单相关接口开发", "支持用户发布订单和查看订单状态", "项目基本跑通了从发布需求到接单处理的流程", "对 Go Web 框架和接口开发有了初步实践"], "link": "https://github.com/demo-user/campus-errand-api"}, {"name": "个人博客系统", "role": "个人开发", "type": "个人练习项目", "start_date": "2023-11", "end_date": "2024-01", "tech_stack": ["Java", "Spring Boot", "MySQL", "Thymeleaf", "Bootstrap"], "description": "基于 Spring Boot 实现一个简单个人博客系统，支持文章发布、文章列表、分类管理、评论和后台登录。", "responsibilities": "", "achievements": ["完成文章、分类、评论等基础模块开发", "设计博客文章和评论相关数据表", "实现文章列表、详情页和后台文章管理", "完成简单的登录校验和后台管理页面", "部署到本地服务器进行演示", "完成一个基础博客系统", "熟悉了 MVC 开发流程", "了解了前后端页面渲染和后端接口之间的关系"], "link": ""}]', '["Java", "Go", "SQL", "Spring Boot", "Gin", "MySQL", "Redis", "MyBatis", "JWT", "Linux 基础", "Git", "Docker 基础", "RESTful API", "接口联调", "Swagger 基础"]', 2, '""', '""', '""', '合肥工业大学', '本科', '计算机科学与技术', '2026-06', '{"name": "刘子恒", "current_city": "合肥", "school": "合肥工业大学", "degree": "本科", "major": "计算机科学与技术", "graduation_year": "2026-06"}', NULL, '[{"company": "安徽某信息技术有限公司", "role": "后端开发实习生", "start_date": "2025-03", "end_date": "2025-05", "work_content": "参与公司内部管理系统的部分接口开发和维护，主要涉及基础数据管理、列表查询和简单统计功能。", "tech_stack": ["Java", "Spring Boot", "MySQL", "MyBatis"], "achievements": ["参与基础信息管理模块接口开发", "完成部分列表查询、详情查询和新增修改接口", "根据测试反馈修复字段为空、状态显示不正确等问题", "配合前端完成部分页面联调", "学习了企业项目中的分层结构和接口规范"], "metrics": "参与开发和维护 10 余个接口，修复测试反馈问题 20 余项。"}]', '[{"name": "CET-4", "type": "language_exam", "level": "CET-4", "score": "493", "issued_at": ""}, {"name": "阿里云云计算基础认证 ACA", "type": "云计算认证", "level": "ACA", "score": "通过", "issued_at": "2025-04"}]', '[{"name": "学院学习进步奖", "type": "综合荣誉", "level": "学院级", "awarded_at": "2024-06", "description": "大二学年专业课程成绩较大一有所提升。"}]', '{"works": [{"name": "校园课程资料管理系统", "link": "https://github.com/demo-user/course-resource-system", "password": ""}, {"name": "校园跑腿服务小程序后端", "link": "https://github.com/demo-user/campus-errand-api", "password": ""}], "homepage": "https://github.com/demo-user"}', '{"self_evaluation": "", "additional_info": ""}', ''),
  (207, 2, 7, '07-211后台开发-版本对比B', '13900010006', '[liuziheng.backend@example.com](mailto:liuziheng.backend@example.com)', '合肥', '""', '""', '[{"school": "合肥工业大学", "degree": "本科", "major": "计算机科学与技术", "department": "计算机与信息学院", "start_date": "2022-09", "end_date": "2026-06", "gpa": "3.31", "gpa_scale": "4.0", "rank": "前 40%", "core_courses": ["数据结构", "操作系统", "计算机网络", "数据库系统", "Java 程序设计", "软件工程", "Web 应用开发", "分布式系统导论"], "research_direction": "后端开发、接口安全、缓存优化、服务部署", "lab": "", "paper_summary": ""}]', '[{"company": "安徽某信息技术有限公司", "role": "后端开发实习生", "start_date": "2025-03", "end_date": "2025-05", "work_content": "参与公司内部管理系统的后端接口开发和维护，主要负责基础数据管理、列表查询、权限字段校验和部分统计接口。", "tech_stack": ["Java", "Spring Boot", "MySQL", "MyBatis", "Redis"], "achievements": ["参与员工信息、部门信息、审批记录等模块接口开发", "完成 14 个列表查询、详情查询、新增修改和状态更新接口", "为列表接口增加分页、条件筛选和状态枚举校验，减少前端异常处理成本", "根据测试反馈修复字段为空、分页异常、状态显示不一致等问题", "配合前端完成接口联调，补充 Swagger 注释和字段说明", "对部分高频查询增加索引，并协助排查 2 个慢查询问题"], "metrics": "参与开发和维护 20 余个接口，修复测试反馈问题 30 余项，负责模块在内部测试环境完成验收。"}]', '[{"name": "校园课程资料管理系统", "role": "后端负责人", "type": "课程项目 / 校园工具项目", "start_date": "2024-03", "end_date": "2024-06", "tech_stack": ["Java", "Spring Boot", "MySQL", "MyBatis", "Redis", "JWT", "Docker", "Nginx", "Swagger"], "description": "面向校内课程资料共享和学习资料管理场景，开发课程资料管理系统，支持用户登录、资料分类、资料检索、资料收藏、下载记录、后台审核和基础统计功能。", "responsibilities": "", "achievements": ["负责用户、课程、资料、收藏、下载记录和审核模块的后端接口设计与开发", "设计课程表、资料表、用户表、收藏表和下载记录表，并为高频查询字段增加索引", "使用 JWT 实现登录鉴权，对资料上传、收藏、下载记录和审核接口增加权限校验", "使用 Redis 缓存热门课程资料列表和资料详情，降低重复查询压力", "使用 Swagger 维护接口文档，配合前端完成字段对齐和接口联调", "使用 Docker 和 Nginx 完成本地部署演示，整理启动说明文档", "完成 28 个后端接口，覆盖资料上传、审核、检索、收藏、下载记录等核心流程", "热门资料列表接口加入 Redis 缓存后，重复访问平均响应时间由约 210ms 降低到 70ms 左右", "为课程 ID、资料状态、创建时间等字段增加索引，优化后台资料筛选查询效率", "补充参数校验和权限判断，修复未登录下载、越权审核、无效资料状态更新等问题", "完成前后端联调和本地部署，项目在课程展示中完成完整流程演示"], "link": "https://github.com/demo-user/course-resource-system"}, {"name": "校园跑腿服务小程序后端", "role": "后端核心开发", "type": "小组项目", "start_date": "2024-09", "end_date": "2024-12", "tech_stack": ["Go", "Gin", "MySQL", "Redis", "JWT", "Docker", "Nginx"], "description": "面向校内学生跑腿服务场景，开发跑腿需求发布与订单流转后端服务，支持用户登录、需求发布、接单、订单状态更新、个人订单查询和未读通知统计。", "responsibilities": "", "achievements": ["负责订单模块、用户模块和通知模块的后端接口开发", "设计订单状态流转规则，区分待接单、进行中、已完成、已取消等状态", "使用 JWT 实现用户身份校验，限制非订单发布者和非接单者修改订单状态", "使用 Redis 记录用户未读通知数，减少每次进入页面时的聚合查询", "对订单创建、接单、取消等接口补充幂等判断，避免重复点击造成状态异常", "编写接口测试脚本，验证订单发布、接单、取消、完成等核心流程", "完成订单相关 16 个接口，跑通从发布需求到接单、完成、通知提醒的主流程", "增加状态流转校验后，避免了已取消订单被继续接单、非本人修改订单状态等问题", "未读通知数使用 Redis 缓存后，减少了重复统计查询", "通过接口测试发现并修复 12 个边界问题，包括重复提交、空参数、非法状态更新等"], "link": "https://github.com/demo-user/campus-errand-api"}, {"name": "简易消息通知服务", "role": "个人开发", "type": "个人技术练习项目", "start_date": "2025-01", "end_date": "2025-02", "tech_stack": ["Go", "RabbitMQ", "MySQL", "Redis", "Docker"], "description": "为了理解消息队列在后端系统中的使用场景，开发一个简易消息通知服务，模拟订单状态变更后异步生成站内通知的流程。", "responsibilities": "", "achievements": ["设计通知表、消息投递记录表和业务唯一键字段", "使用 RabbitMQ 实现订单事件生产和通知消费", "为消息消费增加失败重试和幂等判断，避免重复消费导致重复通知", "使用 Redis 维护用户未读通知数，并在读取通知后更新缓存", "编写本地测试脚本，模拟消息发送、消费失败、重复消费和重试流程", "跑通订单事件到站内通知的异步处理流程", "通过业务唯一键避免重复消费导致的重复通知问题", "记录消息投递状态，便于排查消费失败和重试情况", "加深了对消息队列解耦、异步处理、最终一致性和幂等设计的理解"], "link": "https://github.com/demo-user/simple-notification-service"}]', '["Java", "Go", "SQL", "Spring Boot", "Gin", "MySQL", "Redis", "MyBatis", "JWT", "RabbitMQ", "Docker", "Nginx", "Linux 基础", "Git", "Swagger", "RESTful API 设计", "接口鉴权", "参数校验", "分页查询", "索引优化", "缓存设计", "消息队列基础", "前后端联调"]', 2, '""', '""', '""', '合肥工业大学', '本科', '计算机科学与技术', '2026-06', '{"name": "刘子恒", "current_city": "合肥", "school": "合肥工业大学", "degree": "本科", "major": "计算机科学与技术", "graduation_year": "2026-06"}', NULL, '[{"company": "安徽某信息技术有限公司", "role": "后端开发实习生", "start_date": "2025-03", "end_date": "2025-05", "work_content": "参与公司内部管理系统的后端接口开发和维护，主要负责基础数据管理、列表查询、权限字段校验和部分统计接口。", "tech_stack": ["Java", "Spring Boot", "MySQL", "MyBatis", "Redis"], "achievements": ["参与员工信息、部门信息、审批记录等模块接口开发", "完成 14 个列表查询、详情查询、新增修改和状态更新接口", "为列表接口增加分页、条件筛选和状态枚举校验，减少前端异常处理成本", "根据测试反馈修复字段为空、分页异常、状态显示不一致等问题", "配合前端完成接口联调，补充 Swagger 注释和字段说明", "对部分高频查询增加索引，并协助排查 2 个慢查询问题"], "metrics": "参与开发和维护 20 余个接口，修复测试反馈问题 30 余项，负责模块在内部测试环境完成验收。"}]', '[{"name": "CET-4", "type": "language_exam", "level": "CET-4", "score": "493", "issued_at": ""}, {"name": "腾讯云开发者基础课程结业证书", "type": "云计算课程认证", "level": "课程认证", "score": "通过", "issued_at": "2025-05"}]', '[{"name": "学院学习进步奖", "type": "综合荣誉", "level": "学院级", "awarded_at": "2024-06", "description": "大二学年专业课程成绩较大一有所提升，数据库系统和计算机网络课程成绩表现较好。"}]', '{"works": [{"name": "校园课程资料管理系统", "link": "https://github.com/demo-user/course-resource-system", "password": ""}, {"name": "校园跑腿服务小程序后端", "link": "https://github.com/demo-user/campus-errand-api", "password": ""}, {"name": "简易消息通知服务", "link": "https://github.com/demo-user/simple-notification-service", "password": ""}], "homepage": "https://github.com/demo-user"}', '{"self_evaluation": "", "additional_info": ""}', ''),
  (208, 2, 8, '08-211后台开发-接近投递版-简历优化', '13900010008', '[shenjiayan.backend@example.com](mailto:shenjiayan.backend@example.com)', '长沙', '""', '""', '[{"school": "湖南师范大学", "degree": "本科", "major": "软件工程", "department": "信息科学与工程学院", "start_date": "2022-09", "end_date": "2026-06", "gpa": "3.47", "gpa_scale": "4.0", "rank": "前 30%", "core_courses": ["数据结构", "操作系统", "计算机网络", "数据库系统", "Java 程序设计", "软件工程", "Web 应用开发", "云计算基础"], "research_direction": "后端开发、接口设计、数据库与缓存应用", "lab": "", "paper_summary": ""}]', '[{"company": "长沙某互联网软件公司", "role": "后端开发实习生", "start_date": "2025-03", "end_date": "2025-05", "work_content": "参与公司内部课程管理系统的接口开发和维护，主要负责课程资料、学习记录和后台配置相关功能。", "tech_stack": ["Java", "Spring Boot", "MySQL", "Redis", "MyBatis"], "achievements": ["参与课程资料管理、学习记录查询和后台配置模块开发", "完成 15 个接口的开发和维护，包括列表查询、详情查询、新增修改和状态更新", "根据测试反馈修复参数为空、分页异常、状态枚举不一致等问题", "配合前端完成接口联调，补充 Swagger 字段说明", "对部分课程资料查询接口增加索引，减少后台筛选等待时间"], "metrics": "参与开发和维护 15 个接口，修复测试反馈问题 20 余项，负责模块完成测试环境验收。"}]', '[{"name": "校园课程学习平台后端系统", "role": "后端核心开发", "type": "课程项目 / 个人强化项目", "start_date": "2024-04", "end_date": "2024-09", "tech_stack": ["Go", "GoFrame", "MySQL", "Redis", "JWT", "Docker", "Nginx"], "description": "面向校内课程学习场景，开发课程学习平台后端系统，支持用户登录、课程列表、课程详情、学习进度记录、资料下载和后台课程管理。", "responsibilities": "", "achievements": ["负责用户登录、课程管理、学习进度和资料下载模块的后端接口开发", "使用 JWT 实现登录鉴权，并对课程资料下载、学习进度提交等接口增加登录态校验", "设计用户表、课程表、章节表、学习进度表和资料表", "使用 Redis 缓存课程列表和课程详情，减少重复查询", "使用 Docker 和 Nginx 完成本地部署，整理启动说明文档", "配合前端完成接口联调，修复字段不一致和状态码处理问题", "完成 25 个后端接口，覆盖课程学习平台主要业务流程", "课程详情接口加入缓存后，重复访问响应时间从约 160ms 降低到 70ms 左右", "修复未登录用户下载课程资料、重复提交学习进度等问题", "项目完成前后端联调，并能够通过 docker-compose 在本地一键启动"], "link": "https://github.com/demo-user/course-learning-platform"}, {"name": "在线简历管理与版本对比工具", "role": "后端开发", "type": "个人项目", "start_date": "2024-10", "end_date": "2025-01", "tech_stack": ["Java", "Spring Boot", "MySQL", "Redis", "MyBatis", "Swagger"], "description": "开发一个用于管理个人简历版本的小工具，支持简历版本创建、列表查看、版本详情、版本备注和简单字段对比，帮助用户记录不同投递方向下的简历调整。", "responsibilities": "", "achievements": ["设计简历版本表、用户表和版本备注表", "开发简历新增、编辑、列表查询、详情查询和版本对比接口", "使用 Swagger 维护接口文档，方便测试和后续扩展", "对简历列表接口增加分页查询和创建时间排序", "使用 Redis 缓存当前用户最新简历版本，减少重复查询", "完成简历版本管理的基础功能", "支持同一用户维护多份简历版本，方便后续比较修改内容", "通过该项目理解了版本号、历史记录和用户数据隔离的基本设计方式", "整理接口文档和测试样例，方便后续继续扩展优化建议功能"], "link": "https://github.com/demo-user/resume-version-tool"}, {"name": "简易订单管理系统", "role": "后端开发", "type": "课程综合项目", "start_date": "2023-11", "end_date": "2024-01", "tech_stack": ["Java", "Spring Boot", "MySQL", "MyBatis", "Bootstrap"], "description": "实现一个基础订单管理系统，支持商品管理、订单创建、订单列表、订单状态修改和后台查询。", "responsibilities": "", "achievements": ["负责订单表、商品表和用户表设计", "开发订单创建、订单查询、订单状态修改等接口", "完成后台列表页面和基础查询功能", "根据测试结果修复部分参数为空和状态异常问题", "完成订单管理基础流程", "熟悉了 Spring Boot、MyBatis 和 MySQL 的基础开发方式", "初步理解了订单状态字段和后台管理系统的实现逻辑"], "link": ""}]', '["Go", "Java", "SQL", "GoFrame", "Spring Boot", "MySQL", "Redis", "MyBatis", "JWT", "Docker", "Nginx", "Linux 基础", "Git", "Swagger", "RESTful API", "接口鉴权", "分页查询", "缓存使用", "前后端联调"]', 2, '""', '""', '""', '湖南师范大学', '本科', '软件工程', '2026-06', '{"name": "沈嘉言", "current_city": "长沙", "school": "湖南师范大学", "degree": "本科", "major": "软件工程", "graduation_year": "2026-06"}', NULL, '[{"company": "长沙某互联网软件公司", "role": "后端开发实习生", "start_date": "2025-03", "end_date": "2025-05", "work_content": "参与公司内部课程管理系统的接口开发和维护，主要负责课程资料、学习记录和后台配置相关功能。", "tech_stack": ["Java", "Spring Boot", "MySQL", "Redis", "MyBatis"], "achievements": ["参与课程资料管理、学习记录查询和后台配置模块开发", "完成 15 个接口的开发和维护，包括列表查询、详情查询、新增修改和状态更新", "根据测试反馈修复参数为空、分页异常、状态枚举不一致等问题", "配合前端完成接口联调，补充 Swagger 字段说明", "对部分课程资料查询接口增加索引，减少后台筛选等待时间"], "metrics": "参与开发和维护 15 个接口，修复测试反馈问题 20 余项，负责模块完成测试环境验收。"}]', '[{"name": "CET-4", "type": "language_exam", "level": "CET-4", "score": "526", "issued_at": ""}, {"name": "全国计算机等级考试二级", "type": "计算机基础认证", "level": "二级", "score": "通过", "issued_at": "2023-09"}]', '[{"name": "学院优秀共青团员", "type": "综合荣誉", "level": "学院级", "awarded_at": "2024-05", "description": "参与班级事务和学院志愿活动，综合表现较好。"}]', '{"works": [{"name": "校园课程学习平台后端系统", "link": "https://github.com/demo-user/course-learning-platform", "password": ""}, {"name": "在线简历管理与版本对比工具", "link": "https://github.com/demo-user/resume-version-tool", "password": ""}], "homepage": "https://github.com/demo-user"}', '{"self_evaluation": "", "additional_info": ""}', ''),
  (209, 2, 9, '09-211后台开发-基础薄弱版-简历优化', '13900010008', '[shenjiayan.backend@example.com](mailto:shenjiayan.backend@example.com)', '长沙', '""', '""', '[{"school": "湖南师范大学", "degree": "本科", "major": "软件工程", "department": "信息科学与工程学院", "start_date": "2022-09", "end_date": "2026-06", "gpa": "3.21", "gpa_scale": "4.0", "rank": "前 45%", "core_courses": ["数据结构", "操作系统", "计算机网络", "数据库系统", "Java 程序设计", "Web 应用开发"], "research_direction": "后端开发", "lab": "", "paper_summary": ""}]', '[{"company": "长沙某软件公司", "role": "后端开发实习生", "start_date": "2025-03", "end_date": "2025-04", "work_content": "参与公司内部系统的部分接口维护，主要协助完成简单接口修改、字段调整和测试问题修复。", "tech_stack": ["Java", "Spring Boot", "MySQL", "MyBatis"], "achievements": ["根据需求修改部分接口字段", "协助修复测试反馈的问题", "参与接口联调和简单 SQL 查询", "学习了公司项目的基本分层结构"], "metrics": "参与维护 5 个左右接口，修复若干字段显示和参数校验问题。"}]', '[{"name": "校园课程学习平台", "role": "后端开发", "type": "课程项目", "start_date": "2024-04", "end_date": "2024-07", "tech_stack": ["Java", "Spring Boot", "MySQL", "Redis"], "description": "开发一个校园课程学习平台，主要包括用户登录、课程列表、课程详情、学习记录和资料管理等功能。", "responsibilities": "", "achievements": ["负责部分后端接口开发", "完成课程列表、课程详情和学习记录相关功能", "使用 MySQL 存储课程和用户数据", "使用 Redis 做缓存", "配合前端完成接口联调", "完成课程学习平台的基础功能", "项目能够正常运行", "熟悉了 Spring Boot 和 MySQL 的基本使用", "对后端接口开发流程有了初步理解"], "link": "https://github.com/demo-user/course-learning-platform-basic"}, {"name": "个人博客系统", "role": "个人开发", "type": "个人练习项目", "start_date": "2023-11", "end_date": "2024-01", "tech_stack": ["Java", "Spring Boot", "MySQL", "Bootstrap"], "description": "基于 Spring Boot 开发个人博客系统，支持文章发布、文章列表、文章详情、分类管理和后台登录。", "responsibilities": "", "achievements": ["完成文章模块和分类模块开发", "设计文章表和分类表", "实现后台登录和文章管理功能", "完成基础页面展示", "完成一个基础博客系统", "熟悉了 Java Web 项目的开发流程", "了解了前后端页面和数据库之间的关系"], "link": ""}, {"name": "学生信息管理系统", "role": "后端开发", "type": "课程练习项目", "start_date": "2023-05", "end_date": "2023-06", "tech_stack": ["Java", "MySQL", "JDBC"], "description": "实现一个简单的学生信息管理系统，支持学生信息新增、删除、修改和查询。", "responsibilities": "", "achievements": ["设计学生信息表", "完成增删改查功能", "编写简单的控制台交互逻辑", "完成基础 CRUD 功能", "通过该项目熟悉了 JDBC 和 MySQL 基础操作"], "link": ""}]', '["Java", "SQL", "Spring Boot", "MySQL", "Redis 基础", "HTML", "CSS", "JavaScript 基础", "Git 基础", "Linux 基础", "接口开发", "数据库增删改查"]', 2, '""', '""', '""', '湖南师范大学', '本科', '软件工程', '2026-06', '{"name": "沈嘉言", "current_city": "长沙", "school": "湖南师范大学", "degree": "本科", "major": "软件工程", "graduation_year": "2026-06"}', NULL, '[{"company": "长沙某软件公司", "role": "后端开发实习生", "start_date": "2025-03", "end_date": "2025-04", "work_content": "参与公司内部系统的部分接口维护，主要协助完成简单接口修改、字段调整和测试问题修复。", "tech_stack": ["Java", "Spring Boot", "MySQL", "MyBatis"], "achievements": ["根据需求修改部分接口字段", "协助修复测试反馈的问题", "参与接口联调和简单 SQL 查询", "学习了公司项目的基本分层结构"], "metrics": "参与维护 5 个左右接口，修复若干字段显示和参数校验问题。"}]', '[{"name": "CET-4", "type": "language_exam", "level": "CET-4", "score": "462", "issued_at": ""}, {"name": "全国计算机等级考试二级", "type": "计算机基础认证", "level": "二级", "score": "通过", "issued_at": "2023-09"}]', '[{"name": "学院学习进步奖", "type": "综合荣誉", "level": "学院级", "awarded_at": "2024-06", "description": "大二学年专业课程成绩相比大一有所提升。"}]', '{"works": [{"name": "校园课程学习平台基础版", "link": "https://github.com/demo-user/course-learning-platform-basic", "password": ""}, {"name": "个人博客系统", "link": "", "password": ""}], "homepage": "https://github.com/demo-user"}', '{"self_evaluation": "", "additional_info": ""}', '');