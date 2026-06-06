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
} from './index';
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

export async function loadRecommendationViewModel(
  input?: Partial<RecommendationRequest>,
): Promise<RecommendationPageViewModel> {
  const response = await runRecommendation(input);
  return toRecommendationPageViewModel(response.result, response.meta);
}

export async function loadMatchAnalysisViewModel(
  jobId: string,
  input?: Partial<MatchAnalysisRequest>,
): Promise<MatchAnalysisPageViewModel> {
  const response = await runMatchAnalysis(jobId, input);
  return toMatchAnalysisPageViewModel(response.result, response.meta);
}

export async function loadResumeOptimizationViewModel(
  input?: Partial<ResumeOptimizationRequest>,
): Promise<ResumeOptimizationPageViewModel> {
  const response = await runResumeOptimization(input);
  return toResumeOptimizationPageViewModel(response.result, response.meta);
}
