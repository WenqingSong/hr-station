// ─── 简历优化 Adapter ───────────────────────────────────────────────────────────

import type { ResumeOptimizationResult, AgentResponseMeta } from '../types';
import {
  toAgentMetaViewModel,
  toTopFixItemViewModels,
  type TopFixItemViewModel,
  type AgentMetaViewModel,
} from './agentAdapter';

// ─── ViewModel 类型 ────────────────────────────────────────────────────────────

export interface KeywordAlignmentViewModel {
  keyword: string;
  currentCoverage: string;
  suggestedAction: string;
}

export interface ProjectOptimizationViewModel {
  projectName: string;
  currentProblem: string;
  suggestedRewrite: string;
  exampleText: string;
  needUserVerify: boolean;
}

export interface RewriteExampleViewModel {
  before: string;
  suggestion: string;
  after: string;
  verificationNotice: string;
}

export interface ResumeOptimizationPageViewModel {
  summary: string;
  applyReadiness: string;
  optimizationStopSignal: string;
  topFixItems: TopFixItemViewModel[];
  keywordAlignment: KeywordAlignmentViewModel[];
  projectOptimization: ProjectOptimizationViewModel[];
  missingEvidencePlan: string[];
  rewriteExamples: RewriteExampleViewModel[];
  verificationNotice: string;
  finalChecklist: string[];
  meta?: AgentMetaViewModel;
}

// ─── 转换函数 ──────────────────────────────────────────────────────────────────

function toKeywordAlignmentViewModel(
  item: ResumeOptimizationResult['keyword_alignment'][number],
): KeywordAlignmentViewModel {
  return {
    keyword: item.keyword,
    currentCoverage: item.current_coverage,
    suggestedAction: item.suggested_action,
  };
}

function toProjectOptimizationViewModel(
  item: ResumeOptimizationResult['project_optimization'][number],
): ProjectOptimizationViewModel {
  return {
    projectName: item.project_name,
    currentProblem: item.current_problem,
    suggestedRewrite: item.suggested_rewrite,
    exampleText: item.example_text,
    needUserVerify: item.need_user_verify,
  };
}

function toRewriteExampleViewModel(
  item: ResumeOptimizationResult['rewrite_examples'][number],
): RewriteExampleViewModel {
  return {
    before: item.before,
    suggestion: item.suggestion,
    after: item.after,
    verificationNotice: item.verification_notice,
  };
}

export function toResumeOptimizationPageViewModel(
  result: ResumeOptimizationResult,
  meta?: AgentResponseMeta,
): ResumeOptimizationPageViewModel {
  return {
    summary: result.summary,
    applyReadiness: result.apply_readiness,
    optimizationStopSignal: result.optimization_stop_signal,
    topFixItems: toTopFixItemViewModels(result.top_fix_items),
    keywordAlignment: result.keyword_alignment.map(toKeywordAlignmentViewModel),
    projectOptimization: result.project_optimization.map(toProjectOptimizationViewModel),
    missingEvidencePlan: result.missing_evidence_plan,
    rewriteExamples: result.rewrite_examples.map(toRewriteExampleViewModel),
    verificationNotice: result.verification_notice,
    finalChecklist: result.final_checklist,
    meta: meta ? toAgentMetaViewModel(meta) : undefined,
  };
}
