// ─── 简历关系型 DTO（K8-5-48-C） ──────────────────────────────────────────────

// ─── 子表类型 ─────────────────────────────────────────────────────────────────

export interface ResumeEducation {
  id?: number;
  resume_id?: number;
  school: string;
  degree: string;
  major: string;
  department: string;
  start_date: string;
  end_date: string;
  gpa: string;
  gpa_scale: string;
  rank_text: string;
  core_courses: string;
  research_direction: string;
  lab: string;
  paper_summary: string;
  sort_order?: number;
}

export interface ResumeProject {
  id?: number;
  resume_id?: number;
  name: string;
  role: string;
  start_date: string;
  end_date: string;
  description: string;
  sort_order?: number;
}

export interface ResumeInternship {
  id?: number;
  resume_id?: number;
  company: string;
  position: string;
  start_date: string;
  end_date: string;
  description: string;
  sort_order?: number;
}

export interface ResumeLanguageExam {
  id?: number;
  resume_id?: number;
  exam_type: string;
  score: string;
  sort_order?: number;
}

export interface ResumeCertificate {
  id?: number;
  resume_id?: number;
  name: string;
  description: string;
  sort_order?: number;
}

export interface ResumeAward {
  id?: number;
  resume_id?: number;
  name: string;
  description: string;
  sort_order?: number;
}

export interface ResumeWork {
  id?: number;
  resume_id?: number;
  name: string;
  url: string;
  password: string;
  sort_order?: number;
}

// ─── 主表类型 ─────────────────────────────────────────────────────────────────

export interface Resume {
  id: number;
  user_id?: number;
  name: string;
  is_current: boolean;

  real_name: string;
  phone: string;
  email: string;
  city: string;

  professional_skills: string;
  ai_application: string;
  homepage: string;
  additional_info: string;
  raw_text?: string;

  educations: ResumeEducation[];
  projects: ResumeProject[];
  internships: ResumeInternship[];
  language_exams: ResumeLanguageExam[];
  certificates: ResumeCertificate[];
  awards: ResumeAward[];
  works: ResumeWork[];

  created_at?: string;
  updated_at?: string;
}

// ─── Request 类型 ─────────────────────────────────────────────────────────────

export type SaveResumePayload = Omit<
  Resume,
  "id" | "user_id" | "raw_text" | "created_at" | "updated_at"
>;

// ─── 简历选项（选择弹窗用） ──────────────────────────────────────────────────

export interface ResumeOption {
  id: number;
  name: string;
  is_current: boolean;
  updated_at?: string;
}

// ─── 版本对比（K8-5-48-C：改为简历 ID 对比） ─────────────────────────────────

export interface FieldChange {
  field: string;
  before: string;
  after: string;
  change_type: string;
  impact: string;
  module_name?: string;
}

export interface RiskChange {
  risk: string;
  before: string;
  after: string;
  trend: string;
}

export interface SuggestionCompletion {
  suggestion: string;
  status: string;
  evidence: string;
}

export interface ResumeCompareResult {
  from_id: number;
  to_id: number;
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

export interface ResumeComparison {
  compare_id: number;
  generated_by: string;
  result: ResumeCompareResult;
  created_at: string;
}
