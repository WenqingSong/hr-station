// ─── API → ViewModel 统一 Loader（K8-5-48-C） ────────────────────────────────

import {
  getJobs,
  getJobDetail,
  getSelectedJob,
  getCurrentResume,
  getResumeById,
  getResumes,
  compareResumes,
} from './index';
import {
  runRecommendation,
  runMatchAnalysis,
  runResumeOptimization,
  getLatestRecommendation,
  getLatestResumeOptimization,
  generateResumeOptimization,
} from './index';
import type { LatestRecommendationResponse, GenerateOptimizationRequest } from './agents';
import type {
  OptimizationResultV2,
  OptimizationLatestResponse,
  OptimizationFixItem,
  OptimizationRewriteExample,
  OptimizationLongTermItem,
} from '../types';
import { isApiClientError } from './errors';
import type { Resume } from '../types';

function isAuthError(err: unknown): boolean {
  return isApiClientError(err) && err.status === 401;
}

import type {
  RecommendationRequest,
  MatchAnalysisRequest,
  ResumeOptimizationRequest,
} from './agents';

import {
  toJobCardViewModelsFromApi,
  toJobDetailViewModelFromApi,
  toResumeSummaryViewModel,
  toResumeOptionViewModelsFromResumes,
  toResumeManagementViewModel,
  toResumeComparisonViewModel,
  toRecommendationPageViewModel,
  toMatchAnalysisPageViewModel,
  toResumeOptimizationPageViewModel,
  type JobCardViewModel,
  type JobDetailViewModel,
  type ResumeSummaryViewModel,
  type ResumeOptionViewModel,
  type ResumeManagementViewModel,
  type ResumeComparisonViewModel,
  type RecommendationPageViewModel,
  type MatchAnalysisPageViewModel,
  type ResumeOptimizationPageViewModel,
} from '../adapters';

// ─── Jobs ViewModel Loader ────────────────────────────────────────────────────

export async function loadJobListViewModels(): Promise<JobCardViewModel[]> {
  const res = await getJobs();
  return toJobCardViewModelsFromApi(res.items);
}

export async function loadJobDetailViewModel(
  jobId: string,
): Promise<JobDetailViewModel> {
  const job = await getJobDetail(jobId);
  return toJobDetailViewModelFromApi(job);
}

export async function loadSelectedJobDetailViewModel(jobId: string): Promise<JobDetailViewModel> {
  const job = await getSelectedJob(jobId);
  return toJobDetailViewModelFromApi(job);
}

// ─── Resumes ViewModel Loader ─────────────────────────────────────────────────

export async function loadCurrentResumeSummaryViewModel(): Promise<ResumeSummaryViewModel> {
  const resume = await getCurrentResume();
  return toResumeSummaryViewModel(resume);
}

export async function loadResumeOptionViewModels(): Promise<ResumeOptionViewModel[]> {
  const resumes = await getResumes();
  return toResumeOptionViewModelsFromResumes(resumes);
}

/** 从真实 API 构建简历选项（K8-5-48-C：使用 getResumes） */
export async function loadResumeOptionViewModelsFromApi(): Promise<ResumeOptionViewModel[]> {
  const resumes = await getResumes();
  if (resumes.length > 0) {
    return toResumeOptionViewModelsFromResumes(resumes);
  }
  return [];
}

/** 加载简历管理页完整 ViewModel（K8-5-48-C：使用 current resume） */
export async function loadResumeManagementPageViewModel(): Promise<ResumeManagementViewModel | null> {
  try {
    const resume = await getCurrentResume();
    if (!resume) return null;
    return toResumeManagementViewModel(resume);
  } catch (err: unknown) {
    if (isAuthError(err)) throw err;
    return null;
  }
}

/** 加载简历版本比较 ViewModel */
export async function loadResumeComparisonViewModel(
  fromId: number,
  toId: number,
  baseName: string,
  targetName: string,
): Promise<ResumeComparisonViewModel> {
  const comparison = await compareResumes(fromId, toId);
  return toResumeComparisonViewModel(comparison, baseName, targetName);
}

// ─── K9-4-E: Compare V2 ViewModel Loader ─────────────────────────────────

import { getLatestResumeCompare, generateResumeCompare } from './agents';
import type { LatestCompareResponse } from './agents';

export type CompareV2ViewModel = {
  compareId: number;
  fromResumeId: number;
  toResumeId: number;
  comparisonMode: string;
  targetDirection: string;
  jobId: number | null;
  generationTier: string;
  generatedBy: string;
  status: string;
  fromProfileType: string;
  fromProfileDominantProfile: string;
  toProfileType: string;
  toProfileDominantProfile: string;
  compareScore: number;
  compareLevel: string;
  winnerResumeId: number;
  winnerLabel: string;
  oneSentenceSummary: string;
  fromResumeSummary: string;
  toResumeSummary: string;
  keyDifferences: string[];
  fromStrengths: string[];
  toStrengths: string[];
  fromRisks: string[];
  toRisks: string[];
  keepFromItems: string[];
  keepToItems: string[];
  mergeSuggestions: string[];
  rewriteExamples: { section: string; fromVersion: string; toVersion: string; suggestedNext: string; reason: string }[];
  nextVersionSuggestions: string[];
  deliveryAdvice: string;
};

export type LatestCompareV2ViewModel = {
  hasResult: boolean;
  compareId?: number;
  generationTier?: string;
  generatedBy?: string;
  status?: string;
  result?: CompareV2ViewModel;
};

function mapCompareV2(raw: any): CompareV2ViewModel {
  return {
    compareId: raw.compare_id ?? 0,
    fromResumeId: raw.from_resume_id ?? 0,
    toResumeId: raw.to_resume_id ?? 0,
    comparisonMode: raw.comparison_mode ?? 'general',
    targetDirection: raw.target_direction ?? '',
    jobId: raw.job_id ?? null,
    generationTier: raw.generation_tier ?? 'fast',
    generatedBy: raw.generated_by ?? '',
    status: raw.status ?? 'ready',
    fromProfileType: raw.from_profile_type ?? '',
    fromProfileDominantProfile: raw.from_profile_dominant_profile ?? '',
    toProfileType: raw.to_profile_type ?? '',
    toProfileDominantProfile: raw.to_profile_dominant_profile ?? '',
    compareScore: raw.compare_score ?? 0,
    compareLevel: raw.compare_level ?? '',
    winnerResumeId: raw.winner_resume_id ?? 0,
    winnerLabel: raw.winner_label ?? '',
    oneSentenceSummary: raw.one_sentence_summary ?? '',
    fromResumeSummary: raw.from_resume_summary ?? '',
    toResumeSummary: raw.to_resume_summary ?? '',
    keyDifferences: Array.isArray(raw.key_differences) ? raw.key_differences : [],
    fromStrengths: Array.isArray(raw.from_strengths) ? raw.from_strengths : [],
    toStrengths: Array.isArray(raw.to_strengths) ? raw.to_strengths : [],
    fromRisks: Array.isArray(raw.from_risks) ? raw.from_risks : [],
    toRisks: Array.isArray(raw.to_risks) ? raw.to_risks : [],
    keepFromItems: Array.isArray(raw.keep_from_items) ? raw.keep_from_items : [],
    keepToItems: Array.isArray(raw.keep_to_items) ? raw.keep_to_items : [],
    mergeSuggestions: Array.isArray(raw.merge_suggestions) ? raw.merge_suggestions : [],
    rewriteExamples: Array.isArray(raw.rewrite_examples) ? raw.rewrite_examples.map((e: any) => ({
      section: e.section ?? '', fromVersion: e.from_version ?? '',
      toVersion: e.to_version ?? '', suggestedNext: e.suggested_next ?? '', reason: e.reason ?? '',
    })) : [],
    nextVersionSuggestions: Array.isArray(raw.next_version_suggestions) ? raw.next_version_suggestions : [],
    deliveryAdvice: raw.delivery_advice ?? '',
  };
}

export async function loadLatestCompareV2ViewModel(params: {
  fromResumeId: number; toResumeId: number; comparisonMode?: string;
  targetDirection?: string; jobId?: number;
}): Promise<LatestCompareV2ViewModel> {
  const raw: LatestCompareResponse = await getLatestResumeCompare({
    from_resume_id: params.fromResumeId,
    to_resume_id: params.toResumeId,
    comparison_mode: params.comparisonMode || 'general',
    target_direction: params.targetDirection,
    job_id: params.jobId,
  });
  if (!raw.has_result) return { hasResult: false };
  return {
    hasResult: true,
    compareId: raw.compare_id,
    generationTier: raw.generation_tier,
    generatedBy: raw.generated_by,
    status: raw.status,
    result: raw.result ? mapCompareV2(raw.result) : undefined,
  };
}

export async function loadCompareV2ViewModel(input: {
  fromResumeId: number; toResumeId: number; comparisonMode?: string;
  generationTier?: string; targetDirection?: string; jobId?: number;
}, signal?: AbortSignal): Promise<CompareV2ViewModel> {
  const raw = await generateResumeCompare({
    from_resume_id: input.fromResumeId,
    to_resume_id: input.toResumeId,
    comparison_mode: input.comparisonMode || 'general',
    generation_tier: input.generationTier || 'fast',
    target_direction: input.targetDirection,
    job_id: input.jobId,
  }, signal);
  return mapCompareV2(raw);
}

/** 加载指定简历摘要 ViewModel */
export async function loadResumeSummaryViewModel(
  resumeId: number,
): Promise<ResumeSummaryViewModel> {
  const resume = await getResumeById(resumeId);
  return toResumeSummaryViewModel(resume);
}

// ─── Resume Version Detail Loader（K8-5-48-C：改为 getResumeById） ──────────

/**
 * 根据 resumeId 获取简历详情。
 * 调用 GET /resumes/{id}，返回 Resume | null。
 */
export async function loadResumeVersionDetail(
  resumeId: number,
): Promise<Resume | null> {
  try {
    return await getResumeById(resumeId);
  } catch (err: unknown) {
    if (isAuthError(err)) throw err;
    return null;
  }
}

// ─── Agent ViewModel Loader ───────────────────────────────────────────────────

// ─── K9-1-U-D: Latest Recommendation ViewModel Loader ────────────────────────────

export interface LatestRecommendationViewModel {
  hasResult: boolean;
  resultId?: number;
  resumeId?: number;
  generationTier?: string;
  generatedBy?: string;
  status?: string;
  profileType?: string;
  profileDominantProfile?: string;
  createdAt?: string;
  result?: RecommendationPageViewModel;
}

export async function loadLatestRecommendationViewModel(resumeId: number): Promise<LatestRecommendationViewModel> {
  const raw: LatestRecommendationResponse = await getLatestRecommendation(resumeId);
  if (!raw.has_result) {
    return { hasResult: false };
  }
  // Reuse existing adapter for result
  const result = raw.result
    ? toRecommendationPageViewModel(raw.result, undefined, {
        profileStatus: raw.status ?? "ready",
        profileType: raw.profile_type ?? "fast",
        profileDominantProfile: raw.profile_dominant_profile ?? "",
      })
    : undefined;
  return {
    hasResult: true,
    resultId: raw.result_id,
    resumeId: raw.resume_id,
    generationTier: raw.generation_tier,
    generatedBy: raw.generated_by,
    status: raw.status,
    profileType: raw.profile_type,
    profileDominantProfile: raw.profile_dominant_profile,
    createdAt: raw.created_at,
    result,
  };
}

export async function loadRecommendationViewModel(
  input?: Partial<RecommendationRequest>,
): Promise<RecommendationPageViewModel> {
  const response = await runRecommendation(input);
  // K9-1-I: extract profile fields from response data
  const profileExtras = {
    profileStatus: (response as any).profile_status,
    profileType: (response as any).profile_type,
    profileSource: (response as any).profile_source,
    profileDominantProfile: (response as any).profile_dominant_profile,
    profileMessage: (response as any).profile_message,
  };
  return toRecommendationPageViewModel(response.result, response.meta, profileExtras);
}

export async function loadMatchAnalysisViewModel(
  jobId: string,
  input?: Partial<MatchAnalysisRequest>,
  signal?: AbortSignal,
): Promise<MatchAnalysisPageViewModel> {
  const response = await runMatchAnalysis(jobId, input, signal);
  const vm = toMatchAnalysisPageViewModel(response.result, response.meta);
  // Dev-Reboot-Fix-C: propagate generated_by/fallback from meta
  vm.generatedBy = response.meta?.generated_by || "";
  vm.fallback = response.meta?.fallback;
  return vm;
}

// ─── K9-2-E: Latest Match ViewModel Loader ─────────────────────────────────────

export interface LatestMatchViewModel {
  hasResult: boolean;
  analysisId?: number;
  resumeId?: number;
  jobId?: number;
  generationTier?: string;
  generatedBy?: string;
  status?: string;
  profileType?: string;
  profileDominantProfile?: string;
  createdAt?: string;
  result?: MatchAnalysisPageViewModel;
}

import { getLatestMatchAnalysis } from './agents';
import type { MatchLatestResponse } from './agents';

export async function loadLatestMatchViewModel(jobId: string, resumeId: number): Promise<LatestMatchViewModel> {
  const raw: MatchLatestResponse = await getLatestMatchAnalysis(jobId, resumeId);
  if (!raw.has_result) {
    return { hasResult: false };
  }
  const result = raw.result
    ? toMatchAnalysisPageViewModel(raw.result, { generated_by: raw.generated_by || "fast_tag_engine" } as any)
    : undefined;
  return {
    hasResult: true,
    analysisId: raw.analysis_id,
    resumeId: raw.resume_id,
    jobId: raw.job_id,
    generationTier: raw.generation_tier,
    generatedBy: raw.generated_by,
    status: raw.status,
    profileType: raw.profile_type,
    profileDominantProfile: raw.profile_dominant_profile,
    createdAt: raw.created_at,
    result,
  };
}

export async function loadResumeOptimizationViewModel(
  input?: Partial<ResumeOptimizationRequest>,
): Promise<ResumeOptimizationPageViewModel> {
  const response = await runResumeOptimization(input);
  return toResumeOptimizationPageViewModel(response.result, response.meta);
}

// ─── K9-3-E: V2 Optimization ViewModel Loaders ────────────────────────────────

export interface OptimizationV2ViewModel {
  optimizationId?: number;
  resumeId?: number;
  jobId?: number;
  optimizationMode: string;
  targetDirection?: string;
  generationTier: string;
  generatedBy: string;
  fallback?: boolean;
  agentStatus?: string;
  status: string;
  profileType: string;
  profileDominantProfile: string;
  optimizationScore: number;
  optimizationLevel: string;
  oneSentenceSummary: string;
  quickFixItems: { section: string; issue: string; action: string }[];
  structureFixItems: { section: string; issue: string; action: string }[];
  contentFixItems: { section: string; issue: string; action: string }[];
  keywordFixItems: { section: string; issue: string; action: string }[];
  evidenceFixItems: { section: string; issue: string; action: string }[];
  riskPoints: string[];
  rewriteExamples: { section: string; before: string; after: string; reason: string }[];
  deliveryFocus: string;
  longTermImprovements: { direction: string; note: string }[];
}

export interface LatestOptimizationV2ViewModel {
  hasResult: boolean;
  optimizationId?: number;
  resumeId?: number;
  jobId?: number;
  optimizationMode?: string;
  targetDirection?: string;
  generationTier?: string;
  generatedBy?: string;
  status?: string;
  profileType?: string;
  profileDominantProfile?: string;
  createdAt?: string;
  result?: OptimizationV2ViewModel;
}

function mapOptV2(raw: OptimizationResultV2): OptimizationV2ViewModel {
  return {
    optimizationId: raw.optimization_id,
    resumeId: raw.resume_id,
    jobId: raw.job_id,
    optimizationMode: raw.optimization_mode ?? "general",
    targetDirection: raw.target_direction ?? "",
    generationTier: raw.generation_tier ?? "fast",
    generatedBy: raw.generated_by ?? "",
    fallback: raw.fallback,
    agentStatus: raw.agent_status,
    status: raw.status ?? "ready",
    profileType: raw.profile_type ?? "",
    profileDominantProfile: raw.profile_dominant_profile ?? "",
    optimizationScore: raw.optimization_score ?? 0,
    optimizationLevel: raw.optimization_level ?? "",
    oneSentenceSummary: raw.one_sentence_summary ?? "",
    quickFixItems: Array.isArray(raw.quick_fix_items) ? raw.quick_fix_items : [],
    structureFixItems: Array.isArray(raw.structure_fix_items) ? raw.structure_fix_items : [],
    contentFixItems: Array.isArray(raw.content_fix_items) ? raw.content_fix_items : [],
    keywordFixItems: Array.isArray(raw.keyword_fix_items) ? raw.keyword_fix_items : [],
    evidenceFixItems: Array.isArray(raw.evidence_fix_items) ? raw.evidence_fix_items : [],
    riskPoints: Array.isArray(raw.risk_points) ? raw.risk_points : [],
    rewriteExamples: Array.isArray(raw.rewrite_examples) ? raw.rewrite_examples : [],
    deliveryFocus: raw.delivery_focus ?? "",
    longTermImprovements: Array.isArray(raw.long_term_improvements) ? raw.long_term_improvements : [],
  };
}

export async function loadLatestOptimizationV2ViewModel(
  resumeId: number,
  params: { optimization_mode?: string; target_direction?: string; job_id?: number },
): Promise<LatestOptimizationV2ViewModel> {
  const raw: OptimizationLatestResponse = await getLatestResumeOptimization(resumeId, params);
  if (!raw.has_result) return { hasResult: false };
  const result = raw.result ? mapOptV2(raw.result) : undefined;
  return {
    hasResult: true,
    optimizationId: raw.optimization_id,
    resumeId: raw.resume_id,
    jobId: raw.job_id,
    optimizationMode: raw.optimization_mode,
    targetDirection: raw.target_direction,
    generationTier: raw.generation_tier,
    generatedBy: raw.generated_by,
    status: raw.status,
    profileType: raw.profile_type,
    profileDominantProfile: raw.profile_dominant_profile,
    createdAt: raw.created_at,
    result,
  };
}

export async function loadGenerateOptimizationV2ViewModel(
  resumeId: number,
  input: GenerateOptimizationRequest,
  signal?: AbortSignal,
): Promise<OptimizationV2ViewModel> {
  const raw = await generateResumeOptimization(resumeId, input, signal);
  return mapOptV2(raw);
}
