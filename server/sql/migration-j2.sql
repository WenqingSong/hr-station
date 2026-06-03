-- ============================================================
-- Offer 捕手 - 数据库迁移 J2
-- K8-5-22-A：resume_versions 表 V2 字段扩展
--
-- 迁移策略：
-- 1. 只新增列和索引，不删除任何旧字段
-- 2. schema_version 默认为 1（旧数据）
-- 3. 新列为 NULL 或空字符串默认值，不影响现有业务
-- 4. 本阶段只做数据库字段准备，不改 Go/TS 源码
--
-- 注意：
-- - 本文件应只执行一次
-- - 部分数据库不支持 IF NOT EXISTS，需人工在 clean env 确认
-- ============================================================

-- 1. schema_version：标识简历版本（1=V1, 2=V2）
ALTER TABLE `resume_versions` ADD COLUMN `schema_version` INTEGER NOT NULL DEFAULT 1 COMMENT 'schema版本：1=V1旧版 2=V2新版';

-- 2. 摘要结构化列：列表展示、筛选、排序用
ALTER TABLE `resume_versions` ADD COLUMN `target_direction` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '目标方向（摘要列）';
ALTER TABLE `resume_versions` ADD COLUMN `job_type`         VARCHAR(50)  NOT NULL DEFAULT '' COMMENT '岗位类型（摘要列）';
ALTER TABLE `resume_versions` ADD COLUMN `expected_cities`  VARCHAR(255) NOT NULL DEFAULT '' COMMENT '期望城市逗号分隔（摘要列）';
ALTER TABLE `resume_versions` ADD COLUMN `school`           VARCHAR(100) NOT NULL DEFAULT '' COMMENT '学校（摘要列）';
ALTER TABLE `resume_versions` ADD COLUMN `degree`           VARCHAR(20)  NOT NULL DEFAULT '' COMMENT '学历（摘要列）';
ALTER TABLE `resume_versions` ADD COLUMN `major`            VARCHAR(100) NOT NULL DEFAULT '' COMMENT '专业（摘要列）';
ALTER TABLE `resume_versions` ADD COLUMN `graduation_year`  VARCHAR(10)  NOT NULL DEFAULT '' COMMENT '毕业时间（摘要列）';

-- 3. V2 模块 JSON 字段：结构化简历数据
ALTER TABLE `resume_versions` ADD COLUMN `profile_json`       JSON NULL COMMENT '基础信息 JSON';
ALTER TABLE `resume_versions` ADD COLUMN `job_intention_json` JSON NULL COMMENT '求职意向 JSON';
ALTER TABLE `resume_versions` ADD COLUMN `internship_json`    JSON NULL COMMENT '实习经历 JSON（V2新增，与experience_json并存）';
ALTER TABLE `resume_versions` ADD COLUMN `certificates_json`  JSON NULL COMMENT '证书与语言 JSON';
ALTER TABLE `resume_versions` ADD COLUMN `awards_json`        JSON NULL COMMENT '获奖与竞赛 JSON';
ALTER TABLE `resume_versions` ADD COLUMN `portfolio_json`     JSON NULL COMMENT '作品链接 JSON';
ALTER TABLE `resume_versions` ADD COLUMN `extra_json`         JSON NULL COMMENT '补充信息 JSON（self_evaluation以列+JSON双存）';

-- 4. 保留字段确认（以下字段不改不删）
-- `id`, `user_id`, `version_no`, `name`, `phone`, `email`, `location`
-- `target_city`, `target_role`
-- `education_json`, `experience_json`, `project_json`, `skills_json`
-- `self_evaluation`, `raw_text`, `created_at`
