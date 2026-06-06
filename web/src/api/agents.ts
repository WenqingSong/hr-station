// ─── 三类 Agent API（K8-5-48-D：resume_version_no → resume_id） ─────────────

import { post } from './client';
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
