// ─── 岗位推荐 Adapter ───────────────────────────────────────────────────────────

import type { RecommendationResult, AgentResponseMeta } from '../types';
import {
  toAgentMetaViewModel,
  toTopFixItemViewModels,
  toEvidenceItemViewModels,
  type TopFixItemViewModel,
  type EvidenceItemViewModel,
  type AgentMetaViewModel,
} from './agentAdapter';

// ─── ViewModel 类型 ────────────────────────────────────────────────────────────

export interface RecommendedJobViewModel {
  jobId: string;
  jobTitle: string;
  recommendationLevel: string;
  matchKeywords: string[];
  location: string;
  reason: string;
  riskTags: string[];
  nextAction: string;
}

export interface RecommendationPageViewModel {
  summary: string;
  recommendedJobs: RecommendedJobViewModel[];
  applyReadiness: string;
  optimizationStopSignal: string;
  topFixItems: TopFixItemViewModel[];
  evidenceChain: EvidenceItemViewModel[];
  riskItems: string[];
  nextActions: string[];
  meta?: AgentMetaViewModel;
}

// ─── 转换函数 ──────────────────────────────────────────────────────────────────

function toRecommendedJobViewModel(
  job: RecommendationResult['recommended_jobs'][number],
): RecommendedJobViewModel {
  return {
    jobId: job.job_id,
    jobTitle: job.job_title,
    recommendationLevel: job.recommendation_level,
    matchKeywords: job.match_keywords,
    location: job.location,
    reason: job.reason,
    riskTags: job.risk_tags,
    nextAction: job.next_action,
  };
}

export function toRecommendationPageViewModel(
  result: RecommendationResult,
  meta?: AgentResponseMeta,
): RecommendationPageViewModel {
  const recommendedJobs = Array.isArray(result.recommended_jobs) ? result.recommended_jobs : [];
  return {
    summary: result.recommendation_summary ?? '',
    recommendedJobs: recommendedJobs.map(toRecommendedJobViewModel),
    applyReadiness: result.apply_readiness ?? '',
    optimizationStopSignal: result.optimization_stop_signal ?? '',
    topFixItems: toTopFixItemViewModels(result.top_fix_items),
    evidenceChain: toEvidenceItemViewModels(result.evidence_chain),
    riskItems: Array.isArray(result.risk_items) ? result.risk_items : [],
    nextActions: Array.isArray(result.next_actions) ? result.next_actions : [],
    meta: meta ? toAgentMetaViewModel(meta) : undefined,
  };
}
