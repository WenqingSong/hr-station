// ─── 简历相关类型 ───────────────────────────────────────────────────────────────

/** 求职方向 */
export type ResumeDirection = 'backend' | 'cloud' | 'product' | 'algorithm' | string;

/** 简历状态 */
export type ResumeStatus = 'current' | 'history' | 'inactive' | string;

/** 个人信息 */
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

/** 项目经历 */
export interface ResumeProject {
  id: string;
  name: string;
  tech_stack: string[];
  description: string;
  responsibilities: string[];
}

/** 实习/工作经历 */
export interface ResumeExperience {
  company: string;
  role: string;
  start_date: string;
  end_date: string;
  description: string;
}

/** 教育经历 */
export interface ResumeEducation {
  school: string;
  major: string;
  degree: string;
  start_date: string;
  end_date: string;
  courses: string[];
}

/** 简历完整记录 */
export interface Resume {
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
}

/** 简历版本列表响应 */
export interface ResumeVersionsResponse {
  versions: ResumeVersionListItem[];
  total: number;
}

/** 简历版本比较（历史比��页） */
export interface ResumeVersionComparison {
  base_resume: Resume;
  target_resume: Resume;
  summary: string;
  keyword_changes: ResumeKeywordChange[];
  project_changes: ResumeProjectChange[];
  remaining_risks: string[];
  recommendation: string;
}
