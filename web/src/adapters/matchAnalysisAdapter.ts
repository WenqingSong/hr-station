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

// ─── H18-B Unified Match Types ──────────────────────────────────────────────

export type MatchRiskPoint = {
  riskName: string;
  riskLevel: "high" | "medium" | "low" | string;
  whyItMatters: string;
  evidenceGap: string;
  fixSuggestion: string;
};

export type MatchRequirementCoverage = {
  requirementName: string;
  requirementType: string;
  coverageStatus: string;
  coverageScore: number;
  matchedEvidence: string[];
  gap: string;
  suggestion: string;
};

export type MatchResumeEvidence = {
  evidenceName: string;
  evidenceLevel: number;
  sourceType: string;
  sourceTitle: string;
  evidenceText: string;
  supportsRequirement: string;
  confidence: string;
};

export type MatchScoreBreakdown = {
  directionFit?: number;
  mustHaveCoverage?: number;
  evidenceStrength?: number;
  experienceRelevance?: number;
  riskControl?: number;
};

export type MatchHardGate = {
  status: string;
  scoreCap: number;
  warnings: string[];
};

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
  generatedBy?: string;
  fallback?: boolean;

  // H18-B: New unified Fast match fields
  scoreBreakdown?: MatchScoreBreakdown;
  hardGate?: MatchHardGate;
  coreAdvantages: string[];
  mainGaps: string[];
  structuredRisks: MatchRiskPoint[];
  requirementCoverage: MatchRequirementCoverage[];
  matchedEvidence: MatchResumeEvidence[];
  actionSuggestions: string[];
  applicationDecision?: string;
  decisionLabel?: string;

  // Legacy compat
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

// ─── Helpers ─────────────────────────────────────────────────────────────────

function safeArr<T>(v: unknown, fallback: T[] = []): T[] {
  return Array.isArray(v) ? (v as T[]) : fallback;
}

function safeStr(v: unknown, fallback = ""): string {
  return typeof v === "string" ? v : fallback;
}

// ─── 转换函数 ──────────────────────────────────────────────────────────────────

export function toMatchAnalysisPageViewModel(
  result: MatchAnalysisResult,
  meta?: AgentResponseMeta,
): MatchAnalysisPageViewModel {
  // H18-B: extract structured fields from raw result
  const raw = result as any;

  return {
    matchScore: result.match_score,
    matchLevel: result.match_level || "",
    summary: result.summary || "",
    applyReadiness: typeof result.apply_readiness === 'string' ? result.apply_readiness : "",
    optimizationStopSignal: typeof result.optimization_stop_signal === 'string' ? result.optimization_stop_signal : "",
    topFixItems: toTopFixItemViewModels(result.top_fix_items),
    evidenceChain: toEvidenceItemViewModels(result.evidence_chain),
    matchHighlights: result.match_highlights || [],
    riskItems: result.risk_items || result.risk_points || [],
    nextActions: result.next_actions || [],
    meta: meta ? toAgentMetaViewModel(meta) : undefined,
    generatedBy: meta?.generated_by || raw?.generated_by || "",
    fallback: (meta as any)?.fallback || raw?.fallback,

    // H18-B: New structured fields with fallbacks
    scoreBreakdown: raw?.score_breakdown || undefined,
    hardGate: raw?.hard_gate || undefined,
    coreAdvantages: safeArr(raw?.core_advantages, result.match_highlights || []),
    mainGaps: safeArr(raw?.main_gaps, result.gap_points || []),
    structuredRisks: safeArr(raw?.risk_points_obj || raw?.structured_risks, []),
    requirementCoverage: safeArr(raw?.requirement_coverage, []),
    matchedEvidence: safeArr(raw?.resume_evidence_obj || raw?.matched_evidence, []),
    actionSuggestions: safeArr(raw?.action_suggestions, result.next_actions || []),
    applicationDecision: safeStr(raw?.application_decision || raw?.decision, ""),
    decisionLabel: safeStr(raw?.decision_label || raw?.application_decision, ""),

    // Legacy compat
    fitPoints: result.fit_points || result.match_highlights || [],
    resumeEvidence: result.resume_evidence || [],
    jobRequirements: result.job_requirements || [],
    gapPoints: result.gap_points || result.match_highlights || [],
    riskPoints: safeArr(raw?.risk_points || result.risk_points || result.risk_items, []),
    interviewTalkingPoints: result.interview_talking_points || result.interview_questions || [],
    deliveryStrategy: typeof result.delivery_strategy === 'string' ? result.delivery_strategy : "",
    improvementSuggestions: result.improvement_suggestions || (Array.isArray(result.match_highlights) ? result.match_highlights.slice(0, 3).map((h, i) => ({ priority: i === 0 ? "P0" : "P1", title: h, reason: "", action: h })) : []),
    profileDominantProfile: result.profile_dominant_profile || "",
  };
}
