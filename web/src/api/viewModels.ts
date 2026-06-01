// ─── API → ViewModel 统一 Loader ──────────────────────────────────────────────

import {
  getJobs,
  getJobDetail,
  getSelectedJob,
  getCurrentResume,
  getResumeDetail,
  getResumeOptions,
  compareResumes,
} from './index';
import {
  runRecommendation,
  runMatchAnalysis,
  runResumeOptimization,
} from './index';
import type {
  RecommendationRequest,
  MatchAnalysisRequest,
  ResumeOptimizationRequest,
} from './agents';

import {
  toJobCardViewModels,
  toJobDetailViewModel,
  toResumeSummaryViewModel,
  toResumeOptionViewModels,
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

/** 加载岗位列表 ViewModel */
export async function loadJobListViewModels(): Promise<JobCardViewModel[]> {
  const jobs = await getJobs();
  return toJobCardViewModels(jobs);
}

/** 加载岗位详情 ViewModel */
export async function loadJobDetailViewModel(
  jobId: string,
): Promise<JobDetailViewModel> {
  const job = await getJobDetail(jobId);
  return toJobDetailViewModel(job);
}

/** 加载 Demo 默认岗位 ViewModel */
export async function loadSelectedJobDetailViewModel(): Promise<JobDetailViewModel> {
  const job = await getSelectedJob();
  return toJobDetailViewModel(job);
}

// ─── Resumes ViewModel Loader ─────────────────────────────────────────────────

/** 加载当前简历摘要 ViewModel */
export async function loadCurrentResumeSummaryViewModel(): Promise<ResumeSummaryViewModel> {
  const resume = await getCurrentResume();
  return toResumeSummaryViewModel(resume);
}

/** 加载简历选项列表 ViewModel（选择简历弹窗） */
export async function loadResumeOptionViewModels(): Promise<ResumeOptionViewModel[]> {
  const options = await getResumeOptions();
  return toResumeOptionViewModels(options);
}

/** 从真实 API versions 构建简历选项（10-1 新增） */
export async function loadResumeOptionViewModelsFromApi(): Promise<ResumeOptionViewModel[]> {
  try {
    const versions = await getResumes(); // GET /resumes/versions → Resume[]
    return toResumeOptionViewModelsFromApi(versions);
  } catch {
    // fallback 到 mocks（异常兜底）
    const options = await getResumeOptions(); // fallback to mockResumeOptions
    return toResumeOptionViewModels(options);
  }
}

/** 加载简历管理页完整 ViewModel（含教育/项目/经历/技能/自评） */
export async function loadResumeManagementPageViewModel(): Promise<ResumeManagementViewModel | null> {
  try {
    const resume = await getCurrentResume();
    return toResumeManagementViewModel(resume);
  } catch {
    return null; // 无简历或加载失败 → 返回 null，页面展示空态
  }
}

/** 加载简历版本比较 ViewModel */
export async function loadResumeComparisonViewModel(
  baseId: string,
  targetId: string,
): Promise<ResumeComparisonViewModel> {
  const comparison = await compareResumes(baseId, targetId);
  return toResumeComparisonViewModel(comparison);
}

/** 加载指定简历摘要 ViewModel */
export async function loadResumeSummaryViewModel(
  resumeId: string,
): Promise<ResumeSummaryViewModel> {
  const resume = await getResumeDetail(resumeId);
  return toResumeSummaryViewModel(resume);
}

// ─── Agent ViewModel Loader ───────────────────────────────────────────────────

/** 加载岗位推荐页 ViewModel */
export async function loadRecommendationViewModel(
  input?: Partial<RecommendationRequest>,
): Promise<RecommendationPageViewModel> {
  const response = await runRecommendation(input);
  return toRecommendationPageViewModel(response.result, response.meta);
}

/** 加载岗位匹配分析页 ViewModel */
export async function loadMatchAnalysisViewModel(
  jobId: string,
  input?: Partial<MatchAnalysisRequest>,
): Promise<MatchAnalysisPageViewModel> {
  const response = await runMatchAnalysis(jobId, input);
  return toMatchAnalysisPageViewModel(response.result, response.meta);
}

/** 加载简历优化页 ViewModel */
export async function loadResumeOptimizationViewModel(
  input?: Partial<ResumeOptimizationRequest>,
): Promise<ResumeOptimizationPageViewModel> {
  const response = await runResumeOptimization(input);
  return toResumeOptimizationPageViewModel(response.result, response.meta);
}
