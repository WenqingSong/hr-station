// ─── 岗位推荐 Adapter（K9-1-C：字段契约修复）─────────────────────────────────

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

export type RecommendationSource =
  | "mock" | "fast_llm" | "deep_llm" | "fallback" | "cache" | "unknown";

export type RecommendationLevel = "high" | "medium" | "low";

export interface RecommendedJobViewModel {
  rank: number;
  jobId: number;
  title: string;
  department: string;
  city: string;
  category: string;
  jobType: string;
  score: number;
  scoreLabel: string;
  level: RecommendationLevel;
  oneSentenceReason: string;
  matchHighlights: string[];
  riskPoints: string[];
  suggestedAction: string;
  source: RecommendationSource;
  sourceLabel: string;
}

export interface RecommendationEvidenceRow {
  module: string;
  jobRequirement: string;
  resumeEvidence: string;
  strength: "strong" | "medium" | "weak";
}

export interface RecommendationPageViewModel {
  recommendedJobs: RecommendedJobViewModel[];
  summary: string;
  applyReadiness: string;
  optimizationStopSignal: string;
  topFixItems: TopFixItemViewModel[];
  evidenceChain: EvidenceItemViewModel[];
  riskItems: string[];
  nextActions: string[];

  sourceLabel: string;
  meta?: AgentMetaViewModel;
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

function getLevel(score: number): RecommendationLevel {
  if (score >= 80) return "high";
  if (score >= 60) return "medium";
  return "low";
}

function getScoreLabel(score: number): string {
  if (score >= 80) return "高度匹配";
  if (score >= 60) return "良好匹配";
  return "基础匹配";
}

function getSource(generatedBy?: string): RecommendationSource {
  if (!generatedBy) return "unknown";
  if (generatedBy === "mock_agent") return "mock";
  if (generatedBy === "llm_agent") return "deep_llm";
  if (generatedBy.includes("fallback") || generatedBy.includes("mock")) return "fallback";
  return "unknown";
}

function getSourceLabel(source: RecommendationSource): string {
  switch (source) {
    case "mock": return "快速推荐 · 演示数据";
    case "fast_llm": return "快速推荐 · AI 生成";
    case "deep_llm": return "深度推荐 · AI 生成";
    case "fallback": return "快速推荐 · 兜底结果";
    case "cache": return "历史推荐 · 缓存结果";
    default: return "推荐结果";
  }
}

// ─── 转换函数 ──────────────────────────────────────────────────────────────────

function toRecommendedJobViewModel(
  job: RecommendationResult['recommended_jobs'][number],
  rank: number,
): RecommendedJobViewModel {
  const score = typeof job.recommendation_score === "number" ? job.recommendation_score : 0;
  const source = getSource(job.generated_by);
  return {
    rank,
    jobId: job.job_id,
    title: job.title || "未命名岗位",
    department: job.department || "未标注部门",
    city: job.city || "工作地点待确认",
    category: job.category || "岗位类别待确认",
    jobType: job.job_type || "岗位类型待确认",
    score,
    scoreLabel: getScoreLabel(score),
    level: getLevel(score),
    oneSentenceReason: job.recommendation_reason || "该岗位与当前简历存在一定匹配度，可进一步查看岗位要求。",
    matchHighlights: (job.match_highlights && job.match_highlights.length > 0) ? job.match_highlights : ["岗位方向与简历背景存在基础匹配"],
    riskPoints: (job.risk_points && job.risk_points.length > 0) ? job.risk_points : ["建议进一步补充与该岗位相关的项目细节"],
    suggestedAction: job.suggested_action || "建议先查看岗位详情，再进行岗位匹配分析。",
    source,
    sourceLabel: getSourceLabel(source),
  };
}

export function toRecommendationPageViewModel(
  result: RecommendationResult,
  meta?: AgentResponseMeta,
): RecommendationPageViewModel {
  const jobs = Array.isArray(result.recommended_jobs) ? result.recommended_jobs : [];
  const recommendedJobs = jobs.map((job, i) => toRecommendedJobViewModel(job, i + 1));
  const firstSource = recommendedJobs[0]?.source ?? "unknown";

  return {
    recommendedJobs,
    summary: result.recommendation_summary ?? "",
    applyReadiness: result.apply_readiness ?? "",
    optimizationStopSignal: result.optimization_stop_signal ?? "",
    topFixItems: toTopFixItemViewModels(result.top_fix_items),
    evidenceChain: toEvidenceItemViewModels(result.evidence_chain),
    riskItems: Array.isArray(result.risk_items) ? result.risk_items : [],
    nextActions: Array.isArray(result.next_actions) ? result.next_actions : [],
    sourceLabel: getSourceLabel(firstSource),
    meta: meta ? toAgentMetaViewModel(meta) : undefined,
  };
}
