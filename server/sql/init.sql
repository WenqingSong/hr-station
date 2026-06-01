-- ============================================================
-- Offer 捕手 - 数据库初始化脚本
-- 阶段 3-2：表结构与种子数据
-- 阶段 4-1.3.1：jobs 表扩展（新增 external_source / external_id / raw_json / data_quality 等字段）
-- 阶段 10-P-4-1：新增 resume_optimization_results 表（修复 P0 表缺失）
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
  `email`         VARCHAR(255)     NOT NULL,
  `password_hash` VARCHAR(255)     NOT NULL,
  `name`          VARCHAR(100)     NOT NULL DEFAULT '',
  `created_at`    DATETIME         NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`    DATETIME         NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_users_email` (`email`)
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
  `description`         TEXT             NULL,
  `requirements`        TEXT             NULL,
  `skill_keywords_json` JSON             NULL,
  `status`              VARCHAR(50)      NOT NULL DEFAULT 'active',
  `external_source`     VARCHAR(50)      NOT NULL DEFAULT 'local_mock' COMMENT '外部数据来源：local_mock / tencent_join',
  `external_id`         VARCHAR(100)     NULL COMMENT '外部岗位ID，腾讯岗位使用postId',
  `external_inner_id`   VARCHAR(100)     NULL COMMENT '外部系统内部ID，腾讯岗位可使用列表id',
  `source_url`          VARCHAR(500)     NULL COMMENT '公开岗位来源页面URL',
  `raw_json`            JSON             NULL COMMENT '原始列表与详情响应JSON',
  `data_quality`        VARCHAR(50)      NOT NULL DEFAULT 'manual_enriched' COMMENT '数据质量：full_jd / summary_only / manual_enriched',
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
  KEY `idx_jobs_is_featured` (`is_featured`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='岗位表';

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
INSERT INTO `users` (`id`, `email`, `password_hash`, `name`) VALUES
(1, 'arthur@example.com',       '$2a$10$hUKwISi3F6uHLRj.cFkUkO23HNLW9nH60me2O9fm7q2Hr/Z1nTG3u', 'Arthur'),
(2, 'demo@offer-catcher.com',   '$2a$10$hUKwISi3F6uHLRj.cFkUkO23HNLW9nH60me2O9fm7q2Hr/Z1nTG3u', 'Demo用户');

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
INSERT INTO `jobs` (`id`, `title`, `department`, `city`, `direction`, `category`, `job_type`, `description`, `requirements`, `skill_keywords_json`, `status`, `external_source`, `external_id`, `external_inner_id`, `source_url`, `raw_json`, `data_quality`, `is_featured`, `synced_at`) VALUES
(1, 'Go 后端开发工程师', '技术部', '深圳', '软件开发', '后端开发', '校招',
  '负责公司核心业务系统的后端开发与维护，参与微服务架构设计，保障系统高可用与高性能。与产品经理、前端工程师紧密协作，推动业务快速迭代。',
  '1. 计算机相关专业本科及以上学历\n2. 熟练掌握 Go 语言，了解 Goroutine、Channel 等并发模型\n3. 熟悉 MySQL、Redis 等存储系统\n4. 了解微服务架构，有 Docker/Kubernetes 使用经验优先\n5. 良好的沟通能力和团队协作精神',
  '["Go","MySQL","Redis","Docker","Kubernetes","微服务"]', 'active',
  'local_mock', 'local_mock_1', NULL, NULL, JSON_OBJECT('source', 'local_mock', 'note', 'seed demo job'), 'manual_enriched', 1, NULL),

(2, '云计算研发工程师', '云平台部', '深圳', '软件开发', '云计算', '校招',
  '参与公司云平台核心组件的研发，包括容器编排、资源调度、监控告警等方向。使用 Go 语言开发高性能的云原生基础设施。',
  '1. 计算机相关专业本科及以上学历\n2. 熟练掌握 Go 或 C++ 语言\n3. 熟悉 Linux 操作系统，了解容器技术（Docker/Kubernetes）\n4. 有分布式系统开发经验优先\n5. 对云原生技术有热情，具备快速学习能力',
  '["Go","Kubernetes","Docker","Linux","分布式系统","容器编排"]', 'active',
  'local_mock', 'local_mock_2', NULL, NULL, JSON_OBJECT('source', 'local_mock', 'note', 'seed demo job'), 'manual_enriched', 1, NULL),

(3, 'AI 应用开发工程师', 'AI应用部', '北京', '算法与大数据', 'AI应用', '校招',
  '参与 AI 应用产品的后端开发，构建智能推荐、对话系统、RAG 等应用的 API 服务和工程化落地。',
  '1. 计算机相关专业本科及以上学历\n2. 熟练掌握 Go 或 Python 语言\n3. 熟悉 LLM API 调用（OpenAI/腾讯混元等）\n4. 了解向量数据库和 RAG 架构优先\n5. 有 Prompt 工程经验优先',
  '["Go","Python","LLM","RAG","向量数据库","Prompt工程"]', 'active',
  'local_mock', 'local_mock_3', NULL, NULL, JSON_OBJECT('source', 'local_mock', 'note', 'seed demo job'), 'manual_enriched', 1, NULL),

(4, '数据开发工程师', '数据平台部', '上海', '算法与大数据', '数据开发', '校招',
  '参与公司大数据平台的开发与维护，建设实时和离线数据处理链路，支持业务数据分析与决策。',
  '1. 计算机或数学相关专业本科及以上学历\n2. 熟悉 SQL，了解 Hive/Spark/Flink 等大数据框架\n3. 掌握至少一门编程语言（Go/Python/Java）\n4. 有数据仓库建设经验优先\n5. 良好的数据敏感度和分析能力',
  '["SQL","Python","Spark","Flink","Hive","数据仓库"]', 'active',
  'local_mock', 'local_mock_4', NULL, NULL, JSON_OBJECT('source', 'local_mock', 'note', 'seed demo job'), 'manual_enriched', 1, NULL),

(5, '测试开发工程师', '质量部', '广州', '软件开发', '测试开发', '校招',
  '负责核心业务的质量保障工作，开发自动化测试框架和工具，推动持续集成和持续交付。',
  '1. 计算机相关专业本科及以上学历\n2. 熟悉至少一门编程语言（Go/Python/Java）\n3. 了解自动化测试框架（如 Selenium、Pytest）\n4. 有 CI/CD 经验优先\n5. 细心、有责任心，善于发现和定位问题',
  '["Go","Python","自动化测试","CI/CD","Selenium","Pytest"]', 'active',
  'local_mock', 'local_mock_5', NULL, NULL, JSON_OBJECT('source', 'local_mock', 'note', 'seed demo job'), 'manual_enriched', 1, NULL),

(6, '产品经理', '产品部', '杭州', '产品策划与运营', '产品', '校招',
  '负责公司核心产品的规划与设计，深入理解用户需求，制定产品路线图，推动产品落地与迭代。',
  '1. 本科及以上学历，专业不限\n2. 对互联网产品有浓厚兴趣，有产品实习经验优先\n3. 良好的逻辑思维和数据分析能力\n4. 优秀的沟通能力和跨团队协作能力\n5. 有技术背景优先',
  '["产品设计","需求分析","数据分析","用户研究","Figma"]', 'active',
  'local_mock', 'local_mock_6', NULL, NULL, JSON_OBJECT('source', 'local_mock', 'note', 'seed demo job'), 'manual_enriched', 1, NULL),

(7, '前端开发工程师', '前端组', '北京', '软件开发', '前端开发', '校招',
  '负责公司 Web 前端的开发与优化，使用 React/Vue 构建高质量的用户界面，推动前端工程化建设。',
  '1. 计算机相关专业本科及以上学历\n2. 熟练掌握 JavaScript/TypeScript，熟悉 React 或 Vue 框架\n3. 了解前端工程化工具（Webpack/Vite）\n4. 有移动端 H5 或小程序开发经验优先\n5. 对用户体验和代码质量有追求',
  '["JavaScript","TypeScript","React","Vue","CSS","前端工程化"]', 'active',
  'local_mock', 'local_mock_7', NULL, NULL, JSON_OBJECT('source', 'local_mock', 'note', 'seed demo job'), 'manual_enriched', 1, NULL),

(8, '算法工程师', 'AI算法部', '深圳', '算法与大数据', '算法', '校招',
  '参与推荐系统、NLP、搜索排序等方向的算法研发，将前沿 AI 技术落地到业务场景。',
  '1. 计算机或数学相关专业硕士及以上学历\n2. 熟悉机器学习基础理论，了解深度学习框架（PyTorch/TensorFlow）\n3. 有推荐系统或 NLP 项目经验优先\n4. 熟悉 Python，了解 C++ 优先\n5. 有顶会论文发表优先',
  '["机器学习","深度学习","PyTorch","NLP","推荐系统","Python"]', 'active',
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
