// ─── 简历相关类型 ───────────────────────────────────────────────────────────────

/** 求职方向 */
export type ResumeDirection = 'backend' | 'cloud' | 'product' | 'algorithm' | string;

/** 简历状态 */
export type ResumeStatus = 'current' | 'history' | 'inactive' | string;

/** 简历 Schema 版本 */
export type ResumeSchemaVersion = 1 | 2;

/** 个人信息（V1） */
export interface ResumeProfile {
  name: string;
  age?: number;
  education: string;
  school: string;
  major: string;
  graduation_year: string;
  phone: string;
  email: string;
}

/** 项目经历（V1） */
export interface ResumeProject {
  id: string;
  name: string;
  tech_stack: string[];
  description: string;
  responsibilities: string[];
}

/** 实习/工作经历（V1） */
export interface ResumeExperience {
  company: string;
  role: string;
  start_date: string;
  end_date: string;
  description: string;
}

/** 教育经历（V1） */
export interface ResumeEducation {
  school: string;
  major: string;
  degree: string;
  start_date: string;
  end_date: string;
  courses: string[];
}

// ─── V2 模块类型（K8-5-22-D） ─────────────────────────────────────────────────

/** V2 基础信息 — 模块 1 (K8-5-44-C: 精简，移除与教育经历重复的字段) */
export interface ResumeProfileV2 {
  name?: string;
  current_city?: string;
  phone?: string;
  email?: string;
}

/** V2 求职意向 — 模块 2 */
export interface ResumeJobIntentionV2 {
  target_roles?: string[];
  job_type?: string;
  expected_cities?: string[];
  accept_other_cities?: boolean;
  earliest_start_date?: string;
  internship_duration?: string;
  weekly_available_days?: number;
  target_industries?: string[];
  target_companies?: string[];
}

/** V2 教育经历 — 模块 3 */
export interface ResumeEducationV2 {
  school?: string;
  degree?: string;
  major?: string;
  department?: string;
  start_date?: string;
  end_date?: string;
  gpa?: string;
  gpa_scale?: string;
  rank?: string;
  core_courses?: string[];
  research_direction?: string;
  lab?: string;
  paper_summary?: string;
}

/** V2 实习经历 — 模块 4 (K8-5-44-C: 精简) */
export interface ResumeInternshipV2 {
  company?: string;
  position?: string;
  start_date?: string;
  end_date?: string;
  description?: string;
}

/** V2 项目经历 — 模块 5 (K8-5-44-C: 精简) */
export interface ResumeProjectV2 {
  name?: string;
  role?: string;
  start_date?: string;
  end_date?: string;
  description?: string;
}

/** V2 技能 — 模块 6 (K8-5-44-C: 精简为 2 字段) */
export interface ResumeSkillsV2 {
  professional_skills?: string[];
  ai_application?: string;
}

/** V2 证书与语言 — 模块 7 (K8-5-44-C: 精简) */
export interface ResumeCertificateV2 {
  name?: string;
  description?: string;
}

/** V2 获奖与竞赛 — 模块 8 (K8-5-44-C: 精简) */
export interface ResumeAwardV2 {
  name?: string;
  description?: string;
}

/** V2 作品链接 — 模块 9 (K8-5-44-C: 精简为 works+homepage) */
export interface ResumeWorkV2 {
  name?: string;
  url?: string;
  password?: string;
}
export interface ResumePortfolioV2 {
  works?: ResumeWorkV2[];
  homepage?: string;
}

/** V2 补充信息 — 模块 10 (K8-5-44-C: 删除 self_evaluation) */
export interface ResumeExtraInfoV2 {
  additional_info?: string;
}
}

/** 简历完整记录（兼容 V1 + V2） */
export interface Resume {
  // ── V1 fields ──
  id: string;
  name: string;
  version: string;
  status: ResumeStatus;
  direction: ResumeDirection;
  direction_label: string;
  completeness: number;
  updated_at: string;
  profile: ResumeProfile;
  expected_cities: string[];
  job_type: string;
  skills: string[];
  gaps: string[];
  education: ResumeEducation[];
  experiences: ResumeExperience[];
  projects: ResumeProject[];
  self_evaluation: string;

  // ── V2 fields (K8-5-22-D) ──
  /** Schema 版本：1=V1 2=V2 */
  schema_version?: ResumeSchemaVersion;
  /** V2 目标方向（摘要列） */
  target_direction?: string;
  /** V2 学校（摘要列，来自 profile JSON） */
  school?: string;
  /** V2 学历（摘要列，来自 profile JSON） */
  degree?: string;
  /** V2 专业（摘要列，来自 profile JSON） */
  major?: string;
  /** V2 毕业年份（摘要列，来自 profile JSON） */
  graduation_year?: string;
  /** V2 基础信息 JSON — raw response field */
  profile_json?: ResumeProfileV2;
  /** V2 求职意向 JSON — raw response field */
  job_intention?: ResumeJobIntentionV2;
  /** V2 实习经历 JSON — raw response field */
  internships?: ResumeInternshipV2[];
  /** V2 证书与语言 JSON — raw response field */
  certificates?: ResumeCertificateV2[];
  /** V2 获奖与竞赛 JSON — raw response field */
  awards?: ResumeAwardV2[];
  /** V2 作品链接 JSON — raw response field */
  portfolio_links?: ResumePortfolioV2;
  /** V2 补充信息 JSON — raw response field */
  extra_info?: ResumeExtraInfoV2;
}

/** 选择简历弹窗中的简历选项 */
export interface ResumeOption {
  id: string;
  name: string;
  status: ResumeStatus;
  direction_label: string;
  tags: string[];
  project_count: number;
  updated_at: string;
  recommendation_label?: string;
  mismatch_reason?: string;
  is_default: boolean;
  is_recommended: boolean;
}

/** 简历关键词变化项 */
export interface ResumeKeywordChange {
  keyword: string;
  before_status: string;
  after_status: string;
  change_note: string;
}

/** 简历项目变化项 */
export interface ResumeProjectChange {
  project_name: string;
  change_note: string;
}

/** 简历版本列表项（from GET /api/resumes/versions） */
export interface ResumeVersionListItem {
  id: number;
  version_no: number;
  name: string;
  target_city: string;
  target_role: string;
  created_at: string;

  // ── V2 fields (K8-5-22-D) ──
  schema_version?: ResumeSchemaVersion;
  target_direction?: string;
  expected_cities?: string;
  school?: string;
  degree?: string;
}

/** 简历版本列表响应 */
export interface ResumeVersionsResponse {
  versions: ResumeVersionListItem[];
  total: number;
}

/** 版本对比 — 字段级变化 */
export interface FieldChange {
  field: string;
  before: string;
  after: string;
  change_type: string;
  impact: string;
  /** K8-5-28-B: V2 module name (skills/projects/certificates/extra_info/overall) */
  module_name?: string;
}

/** 版本对比 — 风险变化 */
export interface RiskChange {
  risk: string;
  before: string;
  after: string;
  trend: string;
}

/** 版本对比 — 建议完成状态 */
export interface SuggestionCompletion {
  suggestion: string;
  status: string;
  evidence: string;
}

/** 版本对比 — 分析结果（对应后端 VersionCompareAnalysis） */
export interface VersionCompareResult {
  from_version_no: number;
  to_version_no: number;
  score_before: number;
  score_after: number;
  score_change: number;
  field_changes: FieldChange[];
  risk_changes: RiskChange[];
  suggestion_completion: SuggestionCompletion[];
  key_improvements: string[];
  remaining_risks: string[];
  summary: string;
  authenticity_notice: string;
}

/** 简历版本比较响应（对应后端 VersionCompareResponse） */
export interface ResumeVersionComparison {
  compare_id: number;
  generated_by: string;
  result: VersionCompareResult;
  created_at: string;
}

// ─── V2 API request types (K8-5-22-D) ──────────────────────────────────────────

/** V2 request payload for create/update resume */
export interface CreateResumeRequestV2 {
  schema_version: 2;
  resume_name?: string;
  target_direction?: string;
  description?: string;
  profile?: ResumeProfileV2;
  job_intention?: ResumeJobIntentionV2;
  educations?: ResumeEducationV2[];
  internships?: ResumeInternshipV2[];
  projects?: ResumeProjectV2[];
  skills?: ResumeSkillsV2;
  certificates?: ResumeCertificateV2[];
  awards?: ResumeAwardV2[];
  portfolio_links?: ResumePortfolioV2;
  extra_info?: ResumeExtraInfoV2;
}

/** V1 request payload (legacy, V1 field names) */
export interface CreateResumeRequestV1 {
  schema_version?: 1;
  name?: string;
  phone?: string;
  email?: string;
  location?: string;
  target_city?: string;
  target_role?: string;
  education?: unknown;
  experience?: unknown;
  project?: unknown;
  skills?: unknown;
  self_evaluation?: string;
  raw_text?: string;
}

/** Union type: V1 or V2 request */
export type CreateResumeRequest = CreateResumeRequestV1 | CreateResumeRequestV2;
