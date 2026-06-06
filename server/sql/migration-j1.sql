SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

ALTER TABLE jobs ADD COLUMN application_type VARCHAR(64) NOT NULL DEFAULT '' COMMENT '申请类型' AFTER job_type;
ALTER TABLE jobs ADD COLUMN application_project VARCHAR(64) NOT NULL DEFAULT '' COMMENT '申请项目' AFTER application_type;
ALTER TABLE jobs ADD COLUMN application_subtype VARCHAR(64) NOT NULL DEFAULT '' COMMENT '申请子类型' AFTER application_project;
ALTER TABLE jobs ADD COLUMN slogan VARCHAR(200) NOT NULL DEFAULT '' COMMENT '岗位宣传语' AFTER application_subtype;
ALTER TABLE jobs ADD COLUMN summary TEXT NULL COMMENT '岗位概要' AFTER slogan;
ALTER TABLE jobs ADD COLUMN responsibilities_json JSON NULL COMMENT '职责列表JSON' AFTER requirements;
ALTER TABLE jobs ADD COLUMN must_have_json JSON NULL COMMENT '必备要求JSON' AFTER responsibilities_json;
ALTER TABLE jobs ADD COLUMN nice_to_have_json JSON NULL COMMENT '了解项JSON' AFTER must_have_json;
ALTER TABLE jobs ADD COLUMN bonus_points_json JSON NULL COMMENT '加分项JSON' AFTER nice_to_have_json;
ALTER TABLE jobs ADD COLUMN notes_json JSON NULL COMMENT '备注JSON' AFTER bonus_points_json;
ALTER TABLE jobs ADD COLUMN skills_json JSON NULL COMMENT '技能列表JSON' AFTER skill_keywords_json;
ALTER TABLE jobs ADD COLUMN organizations_json JSON NULL COMMENT '事业群JSON' AFTER skills_json;
ALTER TABLE jobs ADD COLUMN locations_json JSON NULL COMMENT '工作城市JSON' AFTER organizations_json;
ALTER TABLE jobs ADD COLUMN interview_locations_json JSON NULL COMMENT '面试城市JSON' AFTER locations_json;
ALTER TABLE jobs ADD COLUMN facets_json JSON NULL COMMENT '筛选聚合JSON' AFTER interview_locations_json;
ALTER TABLE jobs ADD COLUMN source_name VARCHAR(200) NOT NULL DEFAULT '' COMMENT '数据来源名称' AFTER source_url;
ALTER TABLE jobs ADD COLUMN collected_at DATETIME NULL COMMENT '数据采集时间' AFTER data_quality;
ALTER TABLE jobs ADD INDEX idx_jobs_application_type (application_type);
ALTER TABLE jobs ADD INDEX idx_jobs_status (status);

CREATE TABLE IF NOT EXISTS job_locations (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  job_id BIGINT UNSIGNED NOT NULL,
  city VARCHAR(64) NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_job_location (job_id, city),
  KEY idx_jl_city (city),
  CONSTRAINT fk_jl_job FOREIGN KEY (job_id) REFERENCES jobs (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='岗位工作城市表';

CREATE TABLE IF NOT EXISTS job_business_groups (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  job_id BIGINT UNSIGNED NOT NULL,
  code VARCHAR(32) NOT NULL,
  name VARCHAR(128) NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_job_bg (job_id, code),
  KEY idx_jbg_code (code),
  CONSTRAINT fk_jbg_job FOREIGN KEY (job_id) REFERENCES jobs (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='岗位事业群表';

CREATE TABLE IF NOT EXISTS job_skills (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  job_id BIGINT UNSIGNED NOT NULL,
  skill VARCHAR(128) NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_job_skill (job_id, skill),
  KEY idx_js_skill (skill),
  CONSTRAINT fk_js_job FOREIGN KEY (job_id) REFERENCES jobs (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='岗位技能表';

SET FOREIGN_KEY_CHECKS = 1;
-- DEPRECATED: 2026-06-06 - K8-5-51-B. This file is no longer executed. All content integrated into init.sql.
