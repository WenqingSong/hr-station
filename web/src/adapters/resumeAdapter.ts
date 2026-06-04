// ─── 简历数据 Adapter ───────────────────────────────────────────────────────────

import type {
  Resume,
  ResumeEducation,
  ResumeExperience,
  ResumeProject,
  ResumeOption,
  ResumeVersionComparison,
  ResumeVersionListItem,
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
  /** V2: schema version (1 or 2), undefined for V1 */
  schemaVersion?: number;
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
  /** V2: schema version (1 or 2), undefined for V1 */
  schemaVersion?: number;
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
  /** V2: schema version (1 or 2), undefined for V1 */
  schemaVersion?: number;
}

export interface ResumeComparisonViewModel {
  baseName: string;
  targetName: string;
  compareId: number;
  generatedBy: string;
  fromVersionNo: number;
  toVersionNo: number;
  scoreBefore: number;
  scoreAfter: number;
  scoreChange: number;
  fieldChanges: { field: string; before: string; after: string; changeType: string; impact: string }[];
  riskChanges: { risk: string; before: string; after: string; trend: string }[];
  suggestionCompletion: { suggestion: string; status: string; evidence: string }[];
  keyImprovements: string[];
  remainingRisks: string[];
  summary: string;
  authenticityNotice: string;
  createdAt: string;
}

// ─── V2 helper: extract V2 summary fields with V1 fallback ──────────────────────

function getSchemaVersion(resume: Resume): number | undefined {
  return resume.schema_version;
}

function getV2TargetDirection(resume: Resume): string {
  return resume.target_direction || resume.direction_label || resume.direction || '';
}

function getV2School(resume: Resume): string {
  if (resume.school) return resume.school;
  if (resume.profile?.school) return resume.profile.school;
  if (resume.education?.[0]?.school) return resume.education[0].school;
  return '';
}

function getV2Degree(resume: Resume): string {
  if (resume.degree) return resume.degree;
  if (resume.profile?.education) return resume.profile.education;
  if (resume.education?.[0]?.degree) return resume.education[0].degree;
  return '';
}

function getV2Major(resume: Resume): string {
  if (resume.major) return resume.major;
  if (resume.profile?.major) return resume.profile.major;
  if (resume.education?.[0]?.major) return resume.education[0].major;
  return '';
}

function getV2GraduationYear(resume: Resume): string {
  if (resume.graduation_year) return resume.graduation_year;
  if (resume.profile?.graduation_year) return resume.profile.graduation_year;
  return '';
}

function getV2ExpectedCities(resume: Resume): string[] {
  if (resume.expected_cities?.length) return resume.expected_cities;
  return [];
}

// ─── 转换函数 ──────────────────────────────────────────────────────────────────

function buildProfileSummary(resume: Resume): string {
  const school = getV2School(resume) || resume.profile?.school || '';
  const major = getV2Major(resume) || resume.profile?.major || '';
  const degree = getV2Degree(resume) || resume.profile?.education || '';
  const gradYear = getV2GraduationYear(resume) || resume.profile?.graduation_year || '';
  return [school, major, degree, gradYear].filter(Boolean).join(' / ');
}

const STATUS_LABELS: Record<string, string> = {
  current: '当前使用',
  history: '历史版本',
  inactive: '未激活',
};

export function toResumeSummaryViewModel(resume?: Resume | null): ResumeSummaryViewModel {
  if (!resume) return { id: '', name: '', version: '', statusLabel: '', directionLabel: '', completeness: 0, updatedAt: '', skillTags: [], gapItems: [], projectCount: 0, profileSummary: '', schemaVersion: undefined };
  return {
    id: resume.id,
    name: resume.name,
    version: resume.version,
    statusLabel: STATUS_LABELS[resume.status] || resume.status,
    directionLabel: getV2TargetDirection(resume) || resume.direction_label,
    completeness: resume.completeness,
    updatedAt: resume.updated_at,
    skillTags: resume.skills,
    gapItems: resume.gaps,
    projectCount: resume.projects?.length ?? 0,
    profileSummary: buildProfileSummary(resume),
    schemaVersion: getSchemaVersion(resume),
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
    schemaVersion: undefined,
  };
}

export function toResumeOptionViewModels(
  options: ResumeOption[],
): ResumeOptionViewModel[] {
  return options.map(toResumeOptionViewModel);
}

// ─── API versions → Options（V2-aware） ────────────────────────────────────────

/** 从真实版本列表构建简历选项 VM（V2-aware）。
 * 优先使用 V2 摘要字段（target_direction、expected_cities），
 * 缺失时 fallback 到 V1 字段（target_role、target_city）。 */
export function toResumeOptionViewModelsFromVersions(
  versions: ResumeVersionListItem[],
): ResumeOptionViewModel[] {
  return versions.map((v, i) => {
    const directionLabel = v.target_direction || v.target_role || '';
    const tags: string[] = [];
    if (v.expected_cities) {
      tags.push(...v.expected_cities.split(',').filter(Boolean));
    } else if (v.target_city) {
      tags.push(v.target_city);
    }
    return {
      id: String(v.version_no),
      name: v.name || `简历版本 V${v.version_no}`,
      statusLabel: i === 0 ? "当前使用" : "历史版本",
      directionLabel,
      tags,
      projectCount: 0,
      updatedAt: v.created_at?.substring(0, 10) || "",
      isDefault: i === 0,
      isRecommended: i === 0,
      schemaVersion: v.schema_version,
    };
  });
}

// ─── Management Page VM（V2-aware） ────────────────────────────────────────────

/** 从完整 Resume 构建管理页 ViewModel（V2-aware）。
 * V2 字段优先，V1 fallback。page UI 字段名不变。 */
export function toResumeManagementViewModel(
  resume: Resume,
): ResumeManagementViewModel {
  const summary = toResumeSummaryViewModel(resume);
  const p = resume.profile ?? ({} as Resume['profile']);
  const schVer = getSchemaVersion(resume);

  // V2: prefer V2 school/degree/major over V1 profile fields
  const displaySchool = getV2School(resume) || p.school || '暂未填写';
  const displayMajor = getV2Major(resume) || p.major || '';
  const displayLocation = getV2ExpectedCities(resume).length
    ? getV2ExpectedCities(resume).join(' / ')
    : (resume.expected_cities?.length
      ? resume.expected_cities.join(' / ')
      : '暂未填写');

  const directionLabel = getV2TargetDirection(resume) || resume.direction_label || '暂未填写';

  return {
    summary,
    displayName: p.name || resume.name || '暂未填写',
    displaySchool,
    displayMajor,
    displayPhone: p.phone || '',
    displayEmail: p.email || '',
    displayLocation,
    directionLabel,
    education: Array.isArray(resume.education) ? resume.education : [],
    experiences: Array.isArray(resume.experiences) ? resume.experiences : [],
    projects: Array.isArray(resume.projects) ? resume.projects : [],
    skills: Array.isArray(resume.skills) ? resume.skills : [],
    selfEvaluation: resume.self_evaluation || '',
    hasResume: !!(p.name || resume.name || p.school),
    schemaVersion: schVer,
  };
}

export function toResumeComparisonViewModel(
  comparison: ResumeVersionComparison,
  baseName: string,
  targetName: string,
): ResumeComparisonViewModel {
  const r = comparison.result;
  return {
    baseName,
    targetName,
    compareId: comparison.compare_id,
    generatedBy: comparison.generated_by,
    fromVersionNo: r.from_version_no,
    toVersionNo: r.to_version_no,
    scoreBefore: r.score_before,
    scoreAfter: r.score_after,
    scoreChange: r.score_change,
    fieldChanges: (r.field_changes || []).map((fc) => ({
      field: fc.field,
      before: fc.before,
      after: fc.after,
      changeType: fc.change_type,
      impact: fc.impact,
    })),
    riskChanges: (r.risk_changes || []).map((rc) => ({
      risk: rc.risk,
      before: rc.before,
      after: rc.after,
      trend: rc.trend,
    })),
    suggestionCompletion: (r.suggestion_completion || []).map((sc) => ({
      suggestion: sc.suggestion,
      status: sc.status,
      evidence: sc.evidence,
    })),
    keyImprovements: r.key_improvements || [],
    remainingRisks: r.remaining_risks || [],
    summary: r.summary,
    authenticityNotice: r.authenticity_notice || '',
    createdAt: comparison.created_at,
  };
}
