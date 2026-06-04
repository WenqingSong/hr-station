// ─── 简历数据 Adapter ───────────────────────────────────────────────────────────

import type {
  Resume,
  ResumeEducation,
  ResumeExperience,
  ResumeProject,
  ResumeOption,
  ResumeVersionComparison,
  ResumeVersionListItem,
  ResumeProfileV2,
  ResumeJobIntentionV2,
  ResumeEducationV2,
  ResumeInternshipV2,
  ResumeProjectV2,
  ResumeSkillsV2,
  ResumeCertificateV2,
  ResumeAwardV2,
  ResumePortfolioV2,
  ResumeExtraInfoV2,
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
  /** K8-5-23-B: V2 raw modules for edit backfill */
  v2?: {
    profile?: ResumeProfileV2;
    jobIntention?: ResumeJobIntentionV2;
    educations?: ResumeEducationV2[];
    internships?: ResumeInternshipV2[];
    projects?: ResumeProjectV2[];
    skills?: ResumeSkillsV2;
    certificates?: ResumeCertificateV2[];
    awards?: ResumeAwardV2[];
    portfolio?: ResumePortfolioV2;
    extraInfo?: ResumeExtraInfoV2;
  };
}

/** K8-5-34-B: 简历管理页完整 ViewModel（current + versions 列表） */
export interface ResumeManagementPageViewModel {
  /** 当前简历（latest version） */
  current: ResumeManagementViewModel | null;
  /** 所有简历版本列表（来自 /api/resumes/versions） */
  versions: ResumeOptionViewModel[];
  /** 是否有任何简历 */
  hasAnyResume: boolean;
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
  fieldChanges: { field: string; before: string; after: string; changeType: string; impact: string; moduleName?: string }[];
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
  const raw = resume as Record<string, unknown>;
  // K8-5-34-B: backend returns version_no (int) and created_at; also accept V1 "version" / "updated_at"
  const versionStr = raw.version_no != null ? `V${raw.version_no}` : (resume.version || '');
  const updatedAtStr = (resume.updated_at || String(raw.created_at ?? '')) as string;
  return {
    id: String(resume.id ?? ''),
    name: resume.name,
    version: versionStr,
    statusLabel: STATUS_LABELS[resume.status] || resume.status || (raw.version_no != null ? '历史版本' : ''),
    directionLabel: getV2TargetDirection(resume) || resume.direction_label || '',
    completeness: resume.completeness != null ? resume.completeness : 0,
    updatedAt: updatedAtStr,
    skillTags: resume.skills || [],
    gapItems: resume.gaps || [],
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
 * V2 字段优先，V1 fallback。page UI 字段名不变。
 * K8-5-23-B: passes V2 raw modules through v2 field for edit backfill.
 * K8-5-34-B: handles backend field names (experience/version_no/created_at). */
export function toResumeManagementViewModel(
  resume: Resume,
): ResumeManagementViewModel {
  const raw = resume as Record<string, unknown>;
  // Backend returns "experience" (singular) or "internships" (V2), not "experiences"
  const rawExperiences = Array.isArray(resume.experiences) ? resume.experiences
    : Array.isArray(raw.experience) ? raw.experience as ResumeExperience[]
    : [];

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

  // K8-5-23-B: Build V2 raw modules from response for edit backfill
  const hasV2 = schVer === 2;
  const v2 = hasV2 ? {
    profile: resume.profile_json ?? resume.profile,
    jobIntention: resume.job_intention,
    educations: parseV2Educations(resume),
    internships: resume.internships,
    projects: parseV2Projects(resume),
    skills: parseV2Skills(resume),
    certificates: resume.certificates,
    awards: resume.awards,
    portfolio: resume.portfolio_links,
    extraInfo: resume.extra_info,
  } : undefined;

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
    experiences: rawExperiences,
    projects: Array.isArray(resume.projects) ? resume.projects : [],
    skills: Array.isArray(resume.skills) ? resume.skills : [],
    selfEvaluation: resume.self_evaluation || '',
    hasResume: !!(p.name || resume.name || p.school || resume.school),
    schemaVersion: schVer,
    v2,
  };
}

// ─── K8-5-34-B: Management Page VM (current + versions) ───────────────────────

/** 从 current resume 和 versions 列表构建管理页完整 ViewModel。
 * 只在 current + versions 都为空时才 hasAnyResume=false。 */
export function toResumeManagementPageViewModel(
  currentResume: Resume | null,
  versionList: ResumeVersionListItem[],
): ResumeManagementPageViewModel {
  const current = currentResume ? toResumeManagementViewModel(currentResume) : null;
  const versions = toResumeOptionViewModelsFromVersions(versionList);
  return {
    current,
    versions,
    hasAnyResume: versions.length > 0 || current !== null,
  };
}

// ─── K8-5-23-B: V2 education/projects/skills parsers ──────────────────────────

function parseV2Educations(resume: Resume): ResumeEducationV2[] | undefined {
  const edu = resume.education;
  if (!edu || edu.length === 0) return undefined;
  // V2 educations were stored in education_json; API returns them as "education" array.
  // They'd have V2 extended fields if saved with schema_version=2.
  // Simply cast: extra fields beyond V1 get ignored by V1 type but carried here.
  return edu as unknown as ResumeEducationV2[];
}

function parseV2Projects(resume: Resume): ResumeProjectV2[] | undefined {
  const proj = resume.projects;
  if (!proj || proj.length === 0) return undefined;
  return proj as unknown as ResumeProjectV2[];
}

function parseV2Skills(resume: Resume): ResumeSkillsV2 | undefined {
  // V2 skills are stored as an object; V1 skills are string[].
  const s = resume.skills;
  if (!s || (Array.isArray(s) && s.length === 0)) return undefined;
  if (Array.isArray(s)) {
    // V1 flat list — wrap into other_keywords
    return { other_keywords: s };
  }
  return s as unknown as ResumeSkillsV2;
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
      // K8-5-28-B: pass through module_name
      moduleName: fc.module_name,
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
