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
 * GET /api/jobs/recommend/latest?resume_id=X
 */
export async function getLatestRecommendation(resumeId: number): Promise<LatestRecommendationResponse> {
  return get<LatestRecommendationResponse>(`/jobs/recommend/latest?resume_id=${resumeId}`);
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
): Promise<MatchAnalysisAgentResponse> {
  const payload: MatchAnalysisRequest = {
    ...input,
  };

  return post<MatchAnalysisAgentResponse, MatchAnalysisRequest>(
    `/jobs/${encodeURIComponent(jobId)}/match`,
    payload,
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

// ─── 简历优化 ──────────────────────────────────────────────────────────────────

/**
 * 运行简历优化。
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
