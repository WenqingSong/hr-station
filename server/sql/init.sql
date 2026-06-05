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
-- 2.2 resume_versions - 已拆分至 server/sql/resume_versions.sql
-- 请先执行 resume_versions.sql，或在此处 SOURCE server/sql/resume_versions.sql
-- -----------------------------------------------------------

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
-- 3.2 种子简历版本 - 已拆分至 server/sql/resume_versions.sql
-- -----------------------------------------------------------

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
-- K8-5-32-C: Reviewer demo resumes 已拆分至 server/sql/resume_versions.sql
-- ============================================================
