// ─── 三类 Agent API ───────────────────────────────────────────────────────────

import { post } from './client';
import type {
  AgentResponseEnvelope,
  RecommendationResult,
  MatchAnalysisResult,
  ResumeOptimizationResult,
} from '../types';

// ─── 请求类型 ──────────────────────────────────────────────────────────────────

export interface RecommendationRequest {
  resume_version_no?: number;
  limit?: number;
}

export interface MatchAnalysisRequest {
  resume_version_no?: number;
}

export interface ResumeOptimizationRequest {
  resume_text?: string;
  target_job?: { title?: string; jd_text?: string };
  target_section?: string;
  user_intention?: string;
  /** "core" — use ResumeOptimizationCoreAgent (default: full agent) */
  mode?: "core";
}

// ─── 类型别名 ──────────────────────────────────────────────────────────────────

export type RecommendationAgentResponse = AgentResponseEnvelope<RecommendationResult>;
export type MatchAnalysisAgentResponse = AgentResponseEnvelope<MatchAnalysisResult>;
export type ResumeOptimizationAgentResponse = AgentResponseEnvelope<ResumeOptimizationResult>;

// ─── 岗位推荐 ──────────────────────────────────────────────────────────────────

/**
 * 运行岗位推荐。
 * 调用 POST /api/jobs/recommend（{ resume_version_no, limit }）。失败时错误直接向上传播。
 */
export async function runRecommendation(
  input?: Partial<RecommendationRequest>,
): Promise<RecommendationAgentResponse> {
  const payload: RecommendationRequest = {
    resume_version_no: 0,
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
 * 调用 POST /api/jobs/{jobId}/match（{ resume_version_no }）。失败时错误直接向上传播。
 */
export async function runMatchAnalysis(
  jobId: string,
  input?: Partial<MatchAnalysisRequest>,
): Promise<MatchAnalysisAgentResponse> {
  const payload: MatchAnalysisRequest = {
    resume_version_no: 0,
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
 * 调用 POST /api/resumes/optimize（{ resume_text, target_job, target_section?, user_intention? }）。
 * 失败时错误直接向上传播。
 */
export async function runResumeOptimization(
  input?: Partial<ResumeOptimizationRequest>,
): Promise<ResumeOptimizationAgentResponse> {
  const payload: ResumeOptimizationRequest = {
    resume_text: '',
    target_job: { title: '', jd_text: '' },
    ...input,
  };

  return post<ResumeOptimizationAgentResponse, ResumeOptimizationRequest>(
    '/resumes/optimize',
    payload,
  );
}
