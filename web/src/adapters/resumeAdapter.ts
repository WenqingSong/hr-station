// ─── 简历数据 Adapter（K8-5-48-C：关系型 DTO） ────────────────────────────────

import type {
  Resume,
  ResumeComparison,
  ResumeCompareResult,
  ResumeEducation,
  ResumeInternship,
  ResumeProject,
  ResumeLanguageExam,
  ResumeCertificate,
  ResumeAward,
  ResumeWork,
  ResumeOption,
} from '../types';

// ─── ViewModel 类型 ────────────────────────────────────────────────────────────

export interface ResumeSummaryViewModel {
  id: number;
  name: string;
  isCurrent: boolean;
  updatedAt: string;
}

/** 简历管理页 ViewModel */
export interface ResumeManagementViewModel {
  id: number;
  name: string;
  isCurrent: boolean;

  realName: string;
  phone: string;
  email: string;
  city: string;

  professionalSkills: string;
  aiApplication: string;
  homepage: string;
  additionalInfo: string;
  rawText?: string;

  educations: ResumeEducation[];
  projects: ResumeProject[];
  internships: ResumeInternship[];
  languageExams: ResumeLanguageExam[];
  certificates: ResumeCertificate[];
  awards: ResumeAward[];
  works: ResumeWork[];

  createdAt?: string;
  updatedAt?: string;
}

export interface ResumeOptionViewModel {
  id: number;
  name: string;
  isCurrent: boolean;
  updatedAt?: string;
}

export interface ResumeComparisonViewModel {
  baseName: string;
  targetName: string;
  compareId: number;
  generatedBy: string;
  fromId: number;
  toId: number;
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

// ─── 转换函数 ──────────────────────────────────────────────────────────────────

export function toResumeSummaryViewModel(resume?: Resume | null): ResumeSummaryViewModel {
  if (!resume) return { id: 0, name: '', isCurrent: false, updatedAt: '' };
  return {
    id: resume.id,
    name: resume.name || '',
    isCurrent: resume.is_current || false,
    updatedAt: resume.updated_at || '',
  };
}

export function toResumeManagementViewModel(
  resume: Resume,
): ResumeManagementViewModel {
  return {
    id: resume.id,
    name: resume.name || '',
    isCurrent: resume.is_current || false,

    realName: resume.real_name || '',
    phone: resume.phone || '',
    email: resume.email || '',
    city: resume.city || '',

    professionalSkills: resume.professional_skills || '',
    aiApplication: resume.ai_application || '',
    homepage: resume.homepage || '',
    additionalInfo: resume.additional_info || '',
    rawText: resume.raw_text || '',

    educations: Array.isArray(resume.educations) ? resume.educations : [],
    projects: Array.isArray(resume.projects) ? resume.projects : [],
    internships: Array.isArray(resume.internships) ? resume.internships : [],
    languageExams: Array.isArray(resume.language_exams) ? resume.language_exams : [],
    certificates: Array.isArray(resume.certificates) ? resume.certificates : [],
    awards: Array.isArray(resume.awards) ? resume.awards : [],
    works: Array.isArray(resume.works) ? resume.works : [],

    createdAt: resume.created_at,
    updatedAt: resume.updated_at,
  };
}

export function toResumeOptionViewModel(
  option: ResumeOption,
): ResumeOptionViewModel {
  return {
    id: option.id,
    name: option.name,
    isCurrent: option.is_current || false,
    updatedAt: option.updated_at,
  };
}

export function toResumeOptionViewModels(
  options: ResumeOption[],
): ResumeOptionViewModel[] {
  return options.map(toResumeOptionViewModel);
}

export function toResumeOptionViewModelsFromResumes(
  resumes: Resume[],
): ResumeOptionViewModel[] {
  return resumes.map((r) => ({
    id: r.id,
    name: r.name || '',
    isCurrent: r.is_current || false,
    updatedAt: r.updated_at,
  }));
}

/** @deprecated use toResumeOptionViewModelsFromResumes */
export const toResumeOptionViewModelsFromVersions = toResumeOptionViewModelsFromResumes;

export function toResumeComparisonViewModel(
  comparison: ResumeComparison,
  baseName: string,
  targetName: string,
): ResumeComparisonViewModel {
  const r = comparison.result;
  return {
    baseName,
    targetName,
    compareId: comparison.compare_id,
    generatedBy: comparison.generated_by,
    fromId: r.from_id,
    toId: r.to_id,
    scoreBefore: r.score_before,
    scoreAfter: r.score_after,
    scoreChange: r.score_change,
    fieldChanges: (r.field_changes || []).map((fc) => ({
      field: fc.field,
      before: fc.before,
      after: fc.after,
      changeType: fc.change_type,
      impact: fc.impact,
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
    summary: r.summary || '',
    authenticityNotice: r.authenticity_notice || '',
    createdAt: comparison.created_at || '',
  };
}
