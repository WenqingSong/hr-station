-- K13-Fix-A: job_profiles table for persistent job tags + evidence
-- Non-destructive: CREATE IF NOT EXISTS, no DROP/ALTER on existing tables.

CREATE TABLE IF NOT EXISTS job_profiles (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    job_id BIGINT UNSIGNED NOT NULL,
    direction VARCHAR(100) DEFAULT '',
    category VARCHAR(100) DEFAULT '',
    primary_tags_json JSON NULL,
    skill_tags_json JSON NULL,
    requirement_tags_json JSON NULL,
    responsibility_tags_json JSON NULL,
    risk_tags_json JSON NULL,
    evidence_json JSON NULL,
    source VARCHAR(50) DEFAULT 'rule',
    version VARCHAR(50) DEFAULT 'tag_agent_v1',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_job_id (job_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
