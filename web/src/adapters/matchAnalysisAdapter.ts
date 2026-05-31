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
}

// ─── 转换函数 ──────────────────────────────────────────────────────────────────

export function toMatchAnalysisPageViewModel(
  result: MatchAnalysisResult,
  meta?: AgentResponseMeta,
): MatchAnalysisPageViewModel {
  return {
    matchScore: result.match_score,
    matchLevel: result.match_level,
    summary: result.summary,
    applyReadiness: result.apply_readiness,
    optimizationStopSignal: result.optimization_stop_signal,
    topFixItems: toTopFixItemViewModels(result.top_fix_items),
    evidenceChain: toEvidenceItemViewModels(result.evidence_chain),
    matchHighlights: result.match_highlights,
    riskItems: result.risk_items,
    nextActions: result.next_actions,
    meta: meta ? toAgentMetaViewModel(meta) : undefined,
  };
}
