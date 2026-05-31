// ─── 三类 Agent API ───────────────────────────────────────────────────────────

import { post } from './client';
import type {
  AgentResponseEnvelope,
  RecommendationResult,
  MatchAnalysisResult,
  ResumeOptimizationResult,
  Resume,
  JobDetail,
} from '../types';
import {
  mockRecommendationResponse,
  mockMatchAnalysisResponse,
  mockResumeOptimizationResponse,
  mockCurrentResume,
  mockSelectedJob,
} from '../mocks';

// ─── 请求类型 ──────────────────────────────────────────────────────────────────

export interface RecommendationRequest {
  resume_id?: string;
  resume_snapshot?: Resume;
  limit?: number;
}

export interface MatchAnalysisRequest {
  job_id: string;
  resume_id?: string;
  resume_snapshot?: Resume;
  job_snapshot?: JobDetail;
}

export interface ResumeOptimizationRequest {
  job_id: string;
  resume_id?: string;
  resume_snapshot?: Resume;
  job_snapshot?: JobDetail;
  match_analysis_snapshot?: MatchAnalysisResult;
}

// ─── 类型别名 ──────────────────────────────────────────────────────────────────

export type RecommendationAgentResponse = AgentResponseEnvelope<RecommendationResult>;
export type MatchAnalysisAgentResponse = AgentResponseEnvelope<MatchAnalysisResult>;
export type ResumeOptimizationAgentResponse = AgentResponseEnvelope<ResumeOptimizationResult>;

// ─── Helper ────────────────────────────────────────────────────────────────────

async function withMockFallback<T>(
  requestFn: () => Promise<T>,
  fallbackFn: () => T,
): Promise<T> {
  try {
    return await requestFn();
  } catch {
    return fallbackFn();
  }
}

// ─── 岗位推荐 ──────────────────────────────────────────────────────────────────

/**
 * 运行岗位推荐 — 优先调用后端 POST /jobs/recommend，失败返回 mock。
 */
export async function runRecommendation(
  input?: Partial<RecommendationRequest>,
): Promise<RecommendationAgentResponse> {
  const payload: RecommendationRequest = {
    resume_id: mockCurrentResume.id,
    resume_snapshot: mockCurrentResume,
    limit: 5,
    ...input,
  };

  return withMockFallback(
    () =>
      post<RecommendationAgentResponse, RecommendationRequest>(
        '/jobs/recommend',
        payload,
      ),
    () => mockRecommendationResponse,
  );
}

// ─── 岗位匹配分析 ──────────────────────────────────────────────────────────────

/**
 * 运行岗位匹配分析 — 优先调用后端 POST /jobs/{jobId}/match，失败返回 mock。
 */
export async function runMatchAnalysis(
  jobId: string,
  input?: Partial<MatchAnalysisRequest>,
): Promise<MatchAnalysisAgentResponse> {
  const payload: MatchAnalysisRequest = {
    job_id: jobId,
    resume_id: mockCurrentResume.id,
    resume_snapshot: mockCurrentResume,
    job_snapshot: mockSelectedJob,
    ...input,
  };

  return withMockFallback(
    () =>
      post<MatchAnalysisAgentResponse, MatchAnalysisRequest>(
        `/jobs/${encodeURIComponent(jobId)}/match`,
        payload,
      ),
    () => mockMatchAnalysisResponse,
  );
}

// ─── 简历优化 ──────────────────────────────────────────────────────────────────

/**
 * 运行简历优化 — 优先调用后端 POST /resumes/optimize，失败返回 mock。
 */
export async function runResumeOptimization(
  input?: Partial<ResumeOptimizationRequest>,
): Promise<ResumeOptimizationAgentResponse> {
  const payload: ResumeOptimizationRequest = {
    job_id: mockSelectedJob.id,
    resume_id: mockCurrentResume.id,
    resume_snapshot: mockCurrentResume,
    job_snapshot: mockSelectedJob,
    ...input,
  };

  return withMockFallback(
    () =>
      post<ResumeOptimizationAgentResponse, ResumeOptimizationRequest>(
        '/resumes/optimize',
        payload,
      ),
    () => mockResumeOptimizationResponse,
  );
}
