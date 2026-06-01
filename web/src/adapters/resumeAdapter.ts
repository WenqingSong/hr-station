// ─── 简历数据 Adapter ───────────────────────────────────────────────────────────

import type {
  Resume,
  ResumeEducation,
  ResumeExperience,
  ResumeProject,
  ResumeOption,
  ResumeVersionComparison,
  ResumeVersionListItem,
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

/** 简历管理页完整 ViewModel — 用于 ResumeManagementPage 真实数据展示 */
export interface ResumeManagementViewModel {
  // Summary
  summary: ResumeSummaryViewModel;
  // Personal info
  displayName: string;
  displaySchool: string;
  displayMajor: string;
  displayPhone: string;
  displayEmail: string;
  displayLocation: string;
  directionLabel: string;
  // Arrays
  education: ResumeEducation[];
  experiences: ResumeExperience[];
  projects: ResumeProject[];
  skills: string[];
  selfEvaluation: string;
  // No-profile indicator
  hasResume: boolean;
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

// ─── API versions → Options（10-1） ──────────────────────────────────────────

/** 从真实版本列表构建简历选项 VM（10-1-R：使用 ResumeVersionListItem） */
export function toResumeOptionViewModelsFromVersions(
  versions: ResumeVersionListItem[],
): ResumeOptionViewModel[] {
  return versions.map((v, i) => ({
    id: String(v.version_no),
    name: v.name || `简历版本 V${v.version_no}`,
    statusLabel: i === 0 ? "当前使用" : "历史版本",
    directionLabel: v.target_role || "",
    tags: v.target_city ? [v.target_city] : [],
    projectCount: 0,
    updatedAt: v.created_at?.substring(0, 10) || "",
    isDefault: i === 0,
    isRecommended: i === 0,
  }));
}

// ─── Management Page VM ─────────────────────────────────────────────────────────

/** 从完整 Resume 构建管理页 ViewModel，含空值兜底 */
export function toResumeManagementViewModel(
  resume: Resume,
): ResumeManagementViewModel {
  const summary = toResumeSummaryViewModel(resume);
  const p = resume.profile ?? ({} as Resume['profile']);

  return {
    summary,
    displayName: p.name || resume.name || '暂未填写',
    displaySchool: p.school || '暂未填写',
    displayMajor: p.major || '',
    displayPhone: p.phone || '',
    displayEmail: p.email || '',
    displayLocation: resume.expected_cities?.length
      ? resume.expected_cities.join(' / ')
      : '暂未填写',
    directionLabel: resume.direction_label || '暂未填写',
    education: Array.isArray(resume.education) ? resume.education : [],
    experiences: Array.isArray(resume.experiences) ? resume.experiences : [],
    projects: Array.isArray(resume.projects) ? resume.projects : [],
    skills: Array.isArray(resume.skills) ? resume.skills : [],
    selfEvaluation: resume.self_evaluation || '',
    hasResume: !!(p.name || resume.name || p.school),
  };
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
