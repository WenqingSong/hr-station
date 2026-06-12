-- Hotfix: widen weight column for SEP profile tags
ALTER TABLE semantic_profile_tags MODIFY COLUMN weight DECIMAL(9,4) NOT NULL DEFAULT 0.0000;
ALTER TABLE semantic_profile_tags MODIFY COLUMN confidence DECIMAL(9,4) NOT NULL DEFAULT 0.0000;
