-- SEP-Recommend-DataLayer-C
-- Purpose: add semantic profile storage for recommendation refactor
-- Safe for incremental deployment: creates new tables only, does not alter existing tables
-- SQL owner: OfferCatcher SEP refactor
-- Do not apply to production before DB snapshot and migration review

CREATE TABLE IF NOT EXISTS semantic_profiles (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,

  target_type VARCHAR(32) NOT NULL COMMENT 'resume/job',
  target_id BIGINT NOT NULL COMMENT 'resume_id/job_id',

  profile_version VARCHAR(64) NOT NULL COMMENT 'SEP profile version, e.g. sep_v1',
  prompt_version VARCHAR(64) NOT NULL COMMENT 'profiler prompt version',
  schema_version VARCHAR(64) NOT NULL COMMENT 'profile output schema version',

  content_hash VARCHAR(128) NOT NULL COMMENT 'hash of source resume/job content',

  model_provider VARCHAR(64) NOT NULL DEFAULT 'hy3',
  model_name VARCHAR(128) NOT NULL,
  reasoning_effort VARCHAR(32) NOT NULL COMMENT 'high/low/no_think',

  summary_sentence TEXT NULL,
  strengths_json JSON NULL,
  risks_json JSON NULL,
  fit_directions_json JSON NULL,
  profile_json JSON NULL,

  status VARCHAR(32) NOT NULL DEFAULT 'pending' COMMENT 'pending/processing/verified/failed/stale',
  error_message TEXT NULL,

  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

  UNIQUE KEY uk_semantic_profile_target_hash (
    target_type,
    target_id,
    profile_version,
    schema_version,
    content_hash
  ),

  KEY idx_semantic_profile_target (
    target_type,
    target_id
  ),

  KEY idx_semantic_profile_status (
    status
  ),

  KEY idx_semantic_profile_version (
    profile_version,
    schema_version
  ),

  KEY idx_semantic_profile_updated_at (
    updated_at
  )
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='SEP semantic profiles for resumes and jobs';

CREATE TABLE IF NOT EXISTS semantic_profile_tags (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,

  profile_id BIGINT NOT NULL,
  target_type VARCHAR(32) NOT NULL COMMENT 'resume/job',
  target_id BIGINT NOT NULL COMMENT 'resume_id/job_id',

  tag_code VARCHAR(128) NOT NULL,
  tag_label VARCHAR(255) NOT NULL,
  category VARCHAR(64) NOT NULL COMMENT 'role_direction/technical_skill/engineering_capability/domain_context/experience_signal/risk_signal/preference_signal',

  confidence DECIMAL(5,4) NOT NULL DEFAULT 0.0000,
  evidence_text TEXT NULL,
  explanation TEXT NULL,
  source_section VARCHAR(128) NULL,

  is_user_visible TINYINT(1) NOT NULL DEFAULT 1,
  weight DECIMAL(5,4) NOT NULL DEFAULT 0.0000,

  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

  KEY idx_semantic_tag_profile_id (
    profile_id
  ),

  KEY idx_semantic_tag_target (
    target_type,
    target_id
  ),

  KEY idx_semantic_tag_code (
    tag_code
  ),

  KEY idx_semantic_tag_category (
    category
  ),

  KEY idx_semantic_tag_visible (
    is_user_visible
  ),

  CONSTRAINT fk_semantic_tag_profile
    FOREIGN KEY (profile_id)
    REFERENCES semantic_profiles(id)
    ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='SEP semantic profile tags with evidence and confidence';

-- End of SEP-Recommend-DataLayer-C
