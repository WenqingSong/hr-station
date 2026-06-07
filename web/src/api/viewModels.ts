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
} from './index';
import type { LatestRecommendationResponse } from './agents';
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
): Promise<MatchAnalysisPageViewModel> {
  const response = await runMatchAnalysis(jobId, input);
  return toMatchAnalysisPageViewModel(response.result, response.meta);
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
    ? toMatchAnalysisPageViewModel(raw.result, { generated_by: raw.generated_by || "mock_agent" } as any)
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
