// ─── 岗位匹配分析 Adapter ──────────────────────────────────────────────────────

import type { MatchAnalysisResult, AgentResponseMeta } from '../types';
import {
  toAgentMetaViewModel,
  toTopFixItemViewModels,
  toEvidenceItemViewModels,
  type TopFixItemViewModel,
  type EvidenceItemViewModel,
  type AgentMetaViewModel,
} from './agentAdapter';

// ─── ViewModel 类型 ────────────────────────────────────────────────────────────

export interface MatchAnalysisPageViewModel {
  matchScore: number;
  matchLevel: string;
  summary: string;
  applyReadiness: string;
  optimizationStopSignal: string;
  topFixItems: TopFixItemViewModel[];
  evidenceChain: EvidenceItemViewModel[];
  matchHighlights: string[];
  riskItems: string[];
  nextActions: string[];
  meta?: AgentMetaViewModel;
  // Dev-Reboot-Fix-C: generated_by + fallback for mock detection
  generatedBy?: string;
  fallback?: boolean;

  // K9-2-Regression-C: Full Deep fields from backend
  fitPoints?: string[];
  resumeEvidence?: string[];
  jobRequirements?: string[];
  gapPoints?: string[];
  riskPoints?: string[];
  interviewTalkingPoints?: string[];
  deliveryStrategy?: string;
  improvementSuggestions?: { priority: string; title: string; reason: string; action: string }[];
  profileDominantProfile?: string;
}

// ─── 转换函数 ──────────────────────────────────────────────────────────────────

export function toMatchAnalysisPageViewModel(
  result: MatchAnalysisResult,
  meta?: AgentResponseMeta,
): MatchAnalysisPageViewModel {
  return {
    matchScore: result.match_score,
    matchLevel: result.match_level,
    summary: result.summary || "",
    applyReadiness: typeof result.apply_readiness === 'string' ? result.apply_readiness : "",
    optimizationStopSignal: typeof result.optimization_stop_signal === 'string' ? result.optimization_stop_signal : "",
    topFixItems: toTopFixItemViewModels(result.top_fix_items),
    evidenceChain: toEvidenceItemViewModels(result.evidence_chain),
    matchHighlights: result.match_highlights || [],
    riskItems: result.risk_items || result.risk_points || [],
    nextActions: result.next_actions || [],
    meta: meta ? toAgentMetaViewModel(meta) : undefined,
    // Dev-Reboot-Fix-C: propagate mock/fallback flags
    generatedBy: meta?.generated_by || "",
    fallback: meta?.fallback,
    // K9-2-Regression-C: Deep fields
    fitPoints: result.fit_points || result.match_highlights || [],
    resumeEvidence: result.resume_evidence || [],
    jobRequirements: result.job_requirements || [],
    gapPoints: result.gap_points || (Array.isArray(result.match_highlights) ? result.match_highlights.slice(0, 2) : []),
    riskPoints: result.risk_points || result.risk_items || [],
    interviewTalkingPoints: result.interview_talking_points || result.interview_questions || [],
    deliveryStrategy: typeof result.delivery_strategy === 'string' ? result.delivery_strategy : "",
    improvementSuggestions: result.improvement_suggestions || (Array.isArray(result.match_highlights) ? result.match_highlights.map((_, i) => ({ priority: i === 0 ? "P0" : "P1", title: _, reason: "", action: _ })) : []),
    profileDominantProfile: result.profile_dominant_profile || "",
  };
}
