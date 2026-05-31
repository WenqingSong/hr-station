// ─── 简历数据 Adapter ───────────────────────────────────────────────────────────

import type {
  Resume,
  ResumeOption,
  ResumeVersionComparison,
  ResumeKeywordChange,
  ResumeProjectChange,
} from '../types';

// ─── ViewModel 类型 ────────────────────────────────────────────────────────────

export interface ResumeSummaryViewModel {
  id: string;
  name: string;
  version: string;
  statusLabel: string;
  directionLabel: string;
  completeness: number;
  updatedAt: string;
  skillTags: string[];
  gapItems: string[];
  projectCount: number;
  profileSummary: string;
}

export interface ResumeOptionViewModel {
  id: string;
  name: string;
  statusLabel: string;
  directionLabel: string;
  tags: string[];
  projectCount: number;
  updatedAt: string;
  isDefault: boolean;
  isRecommended: boolean;
  recommendationLabel?: string;
  mismatchReason?: string;
}

export interface ResumeComparisonViewModel {
  baseName: string;
  targetName: string;
  summary: string;
  keywordChanges: ResumeKeywordChange[];
  projectChanges: ResumeProjectChange[];
  remainingRisks: string[];
  recommendation: string;
}

// ─── 转换函数 ──────────────────────────────────────────────────────────────────

function buildProfileSummary(resume: Resume): string {
  const { profile } = resume;
  return [profile.school, profile.major, profile.education, profile.graduation_year]
    .filter(Boolean)
    .join(' / ');
}

const STATUS_LABELS: Record<string, string> = {
  current: '当前使用',
  history: '历史版本',
  inactive: '未激活',
};

export function toResumeSummaryViewModel(resume: Resume): ResumeSummaryViewModel {
  return {
    id: resume.id,
    name: resume.name,
    version: resume.version,
    statusLabel: STATUS_LABELS[resume.status] || resume.status,
    directionLabel: resume.direction_label,
    completeness: resume.completeness,
    updatedAt: resume.updated_at,
    skillTags: resume.skills,
    gapItems: resume.gaps,
    projectCount: resume.projects.length,
    profileSummary: buildProfileSummary(resume),
  };
}

export function toResumeOptionViewModel(
  option: ResumeOption,
): ResumeOptionViewModel {
  return {
    id: option.id,
    name: option.name,
    statusLabel: STATUS_LABELS[option.status] || option.status,
    directionLabel: option.direction_label,
    tags: option.tags,
    projectCount: option.project_count,
    updatedAt: option.updated_at,
    isDefault: option.is_default,
    isRecommended: option.is_recommended,
    recommendationLabel: option.recommendation_label,
    mismatchReason: option.mismatch_reason,
  };
}

export function toResumeOptionViewModels(
  options: ResumeOption[],
): ResumeOptionViewModel[] {
  return options.map(toResumeOptionViewModel);
}

export function toResumeComparisonViewModel(
  comparison: ResumeVersionComparison,
): ResumeComparisonViewModel {
  return {
    baseName: comparison.base_resume.name,
    targetName: comparison.target_resume.name,
    summary: comparison.summary,
    keywordChanges: comparison.keyword_changes,
    projectChanges: comparison.project_changes,
    remainingRisks: comparison.remaining_risks,
    recommendation: comparison.recommendation,
  };
}
