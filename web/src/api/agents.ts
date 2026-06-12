// ─── 三类 Agent API（K8-5-48-D：resume_version_no → resume_id） ─────────────

import { post, get } from './client';
import type {
  AgentResponseEnvelope,
  RecommendationResult,
  MatchAnalysisResult,
  ResumeOptimizationResult,
} from '../types';

// ─── 请求类型 ──────────────────────────────────────────────────────────────────

export interface RecommendationRequest {
  resume_id?: number;
  limit?: number;
  generation_tier?: "fast" | "quality";
}

export interface MatchAnalysisRequest {
  resume_id?: number;
  generation_tier?: "fast" | "quality";
}

export interface ResumeOptimizationRequest {
  resume_id?: number;
  job_id?: number;
  resume_text?: string;
  target_job?: { title?: string; jd_text?: string };
  target_section?: string;
  user_intention?: string;
  mode?: "core";
  generation_tier?: "fast" | "quality";
}

// ─── 类型别名 ──────────────────────────────────────────────────────────────────

export type RecommendationAgentResponse = AgentResponseEnvelope<RecommendationResult>;
export type MatchAnalysisAgentResponse = AgentResponseEnvelope<MatchAnalysisResult>;
export type ResumeOptimizationAgentResponse = AgentResponseEnvelope<ResumeOptimizationResult>;

// ─── 岗位推荐 ──────────────────────────────────────────────────────────────────

/** K9-1-U-D: Latest recommendation response type */
export interface LatestRecommendationResponse {
  has_result: boolean;
  result_id?: number;
  resume_id?: number;
  generation_tier?: string;
  generated_by?: string;
  status?: string;
  profile_type?: string;
  profile_dominant_profile?: string;
  created_at?: string;
  result?: RecommendationResult;
}

/**
 * 获取最新推荐结果。
 * GET /api/jobs/recommend/latest?resume_id=X[&generation_tier=Y]
 * If tier is omitted, returns most recent non-mock result regardless of tier.
 */
export async function getLatestRecommendation(resumeId: number, tier?: string): Promise<LatestRecommendationResponse> {
  let url = `/jobs/recommend/latest?resume_id=${resumeId}`;
  if (tier) url += `&generation_tier=${tier}`;
  return get<LatestRecommendationResponse>(url);
}

/**
 * 运行岗位推荐。
 * 调用 POST /api/jobs/recommend（{ resume_id, limit }）。失败时错误直接向上传播。
 */
export async function runRecommendation(
  input?: Partial<RecommendationRequest>,
): Promise<RecommendationAgentResponse> {
  const payload: RecommendationRequest = {
    limit: 5,
    ...input,
  };

  return post<RecommendationAgentResponse, RecommendationRequest>(
    '/jobs/recommend',
    payload,
  );
}

// ─── 岗位匹配分析 ──────────────────────────────────────────────────────────────

/**
 * 运行岗位匹配分析。
 * 调用 POST /api/jobs/{jobId}/match（{ resume_id }）。失败时错误直接向上传播。
 */
export async function runMatchAnalysis(
  jobId: string,
  input?: Partial<MatchAnalysisRequest>,
  signal?: AbortSignal,
): Promise<MatchAnalysisAgentResponse> {
  const payload: MatchAnalysisRequest = {
    ...input,
  };

  return post<MatchAnalysisAgentResponse, MatchAnalysisRequest>(
    `/jobs/${encodeURIComponent(jobId)}/match`,
    payload,
    signal ? { signal } : undefined,
  );
}

// ─── K9-2-E: Match Latest ─────────────────────────────────────────────────────

export interface MatchLatestResponse {
  has_result: boolean;
  analysis_id?: number;
  resume_id?: number;
  job_id?: number;
  generation_tier?: string;
  generated_by?: string;
  status?: string;
  profile_type?: string;
  profile_dominant_profile?: string;
  created_at?: string;
  result?: MatchAnalysisResult;
}

export type MatchAnalysisAgentResponse = AgentResponseEnvelope<MatchAnalysisResult>;

export async function getLatestMatchAnalysis(jobId: string, resumeId: number): Promise<MatchLatestResponse> {
  return get<MatchLatestResponse>(`/jobs/${encodeURIComponent(jobId)}/match/latest?resume_id=${resumeId}`);
}

// ─── 简历优化（旧版，保留向后兼容） ─────────────────────────────────────────────

/**
 * 运行简历优化（旧版）。
 * 调用 POST /api/resumes/optimize（{ resume_id, job_id, mode }）。
 * 失败时错误直接向上传播。
 */
export async function runResumeOptimization(
  input?: Partial<ResumeOptimizationRequest>,
): Promise<ResumeOptimizationAgentResponse> {
  return post<ResumeOptimizationAgentResponse, ResumeOptimizationRequest>(
    '/resumes/optimize',
    input || {},
  );
}

// ─── K9-3: 新版简历优化 API ──────────────────────────────────────────────────

import type { OptimizationMode, GenerationTier, OptimizationLatestResponse, OptimizationResultV2 } from '../types';

/** K9-3-E: Get latest optimization */
export async function getLatestResumeOptimization(
  resumeId: number,
  params: { optimization_mode?: OptimizationMode; target_direction?: string; job_id?: number },
): Promise<OptimizationLatestResponse> {
  const q = new URLSearchParams();
  if (params.optimization_mode) q.set('optimization_mode', params.optimization_mode);
  if (params.target_direction) q.set('target_direction', params.target_direction);
  if (params.job_id) q.set('job_id', String(params.job_id));
  return get<OptimizationLatestResponse>(`/resumes/${resumeId}/optimization/latest?${q.toString()}`);
}

/** K9-3-E: Generate new optimization */
export interface GenerateOptimizationRequest {
  optimization_mode: OptimizationMode;
  target_direction?: string;
  job_id?: number;
  generation_tier: GenerationTier;
}

export async function generateResumeOptimization(
  resumeId: number,
  input: GenerateOptimizationRequest,
  signal?: AbortSignal,
): Promise<OptimizationResultV2> {
  return post<OptimizationResultV2, GenerateOptimizationRequest>(
    `/resumes/${resumeId}/optimization`,
    input,
    signal ? { signal } : undefined,
  );
}

// ─── K9-4: Resume Compare ──────────────────────────────────────────────────

export interface LatestCompareResponse {
  has_result: boolean;
  compare_id?: number;
  from_resume_id?: number;
  to_resume_id?: number;
  comparison_mode?: string;
  target_direction?: string;
  job_id?: number;
  generation_tier?: string;
  generated_by?: string;
  status?: string;
  from_profile_type?: string;
  from_profile_dominant_profile?: string;
  to_profile_type?: string;
  to_profile_dominant_profile?: string;
  created_at?: string;
  result?: any;
}

export async function getLatestResumeCompare(params: {
  from_resume_id: number;
  to_resume_id: number;
  comparison_mode?: string;
  target_direction?: string;
  job_id?: number;
}): Promise<LatestCompareResponse> {
  const q = new URLSearchParams();
  q.set('from_resume_id', String(params.from_resume_id));
  q.set('to_resume_id', String(params.to_resume_id));
  q.set('comparison_mode', params.comparison_mode || 'general');
  if (params.target_direction) q.set('target_direction', params.target_direction);
  if (params.job_id) q.set('job_id', String(params.job_id));
  return get<LatestCompareResponse>(`/resumes/compare/latest?${q.toString()}`);
}

export interface GenerateCompareRequest {
  from_resume_id: number;
  to_resume_id: number;
  comparison_mode?: string;
  generation_tier?: string;
  target_direction?: string;
  job_id?: number;
}

export async function generateResumeCompare(input: GenerateCompareRequest, signal?: AbortSignal): Promise<any> {
  return post<any, GenerateCompareRequest>('/resumes/compare', input, signal ? { signal } : undefined);
}
