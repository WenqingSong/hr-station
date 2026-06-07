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

export type RecommendationPriorityTone = "strong" | "high" | "try" | "backup" | "not_recommended";

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
  // K9-1-I: priority-based labels
  label: string;
  priority: string;
  priorityText: string;
  priorityTone: RecommendationPriorityTone;
  oneSentenceReason: string;
  matchHighlights: string[];
  riskPoints: string[];
  suggestedAction: string;
  source: RecommendationSource;
  sourceLabel: string;

  // K9-1-T-D: per-job explanation fields
  whyRankedHere: string;
  resumeEvidence: string[];
  jobSpecificReason: string;
  jobSpecificRisks: string[];
  suggestedImprovement: string;
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

  // K9-1-I: profile status fields
  profileStatus: string;
  profileType: string;
  profileSource: string;
  profileDominantProfile: string;
  profileMessage: string;
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
  if (generatedBy === "mock_quality") return "deep_llm";
  if (generatedBy === "llm_agent") return "deep_llm";
  if (generatedBy.includes("fallback")) return "fallback";
  return "unknown";
}

function getSourceLabel(source: RecommendationSource, generatedBy?: string, isQuality?: boolean): string {
  // K9-1-T-D: product-friendly labels — never show "演示数据"
  if (generatedBy === "mock_agent") return "快速推荐 · 规则推荐";
  if (generatedBy === "mock_quality") return "深度推荐 · 规则兜底";
  if (generatedBy === "llm_agent" || generatedBy?.startsWith("llm")) return "深度推荐 · AI生成";
  if (generatedBy?.includes("fallback")) return "推荐失败 · 已保留快速结果";
  switch (source) {
    case "mock": return isQuality ? "深度推荐 · 规则兜底" : "快速推荐 · 规则推荐";
    case "fast_llm": return "快速推荐 · AI生成";
    case "deep_llm": return "深度推荐 · AI生成";
    case "fallback": return "推荐失败 · 已保留快速结果";
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
  const isQuality = job.generated_by === "mock_quality" || job.generated_by === "llm_agent";
  const generatedBy = job.generated_by ?? "";
  // K9-1-I: priority-based labels
  const priority = job.recommendation_priority || defaultPriority(score);
  const label = job.recommendation_label || defaultLabel(priority);

  // K9-1-T-D: new explanation fields with safe fallbacks
  const whyRankedHere = job.why_ranked_here || job.recommendation_reason || "";
  const resumeEvidence = (Array.isArray(job.resume_evidence) && job.resume_evidence.length > 0)
    ? job.resume_evidence
    : [];
  const jobSpecificReason = job.job_specific_reason || job.recommendation_reason || "";
  const jobSpecificRisks = (Array.isArray(job.job_specific_risks) && job.job_specific_risks.length > 0)
    ? job.job_specific_risks
    : (Array.isArray(job.risk_points) ? job.risk_points : []);
  const suggestedImprovement = job.suggested_improvement || job.suggested_action || "";

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
    label,
    priority,
    priorityText: priorityText(priority),
    priorityTone: priority as RecommendationPriorityTone,
    oneSentenceReason: job.recommendation_reason || "该岗位与当前简历存在一定匹配度，可进一步查看岗位要求。",
    matchHighlights: (job.match_highlights && job.match_highlights.length > 0) ? job.match_highlights : ["岗位方向与简历背景存在基础匹配"],
    riskPoints: (job.risk_points && job.risk_points.length > 0) ? job.risk_points : ["建议进一步补充与该岗位相关的项目细节"],
    suggestedAction: job.suggested_action || "建议先查看岗位详情，再进行岗位匹配分析。",
    source,
    sourceLabel: getSourceLabel(source, generatedBy, isQuality),
    // K9-1-T-D: new fields
    whyRankedHere,
    resumeEvidence,
    jobSpecificReason,
    jobSpecificRisks,
    suggestedImprovement,
  };
}

function defaultPriority(score: number): string {
  if (score >= 90) return "strong";
  if (score >= 80) return "high";
  if (score >= 70) return "try";
  if (score >= 60) return "backup";
  return "not_recommended";
}

function defaultLabel(priority: string): string {
  return priorityText(priority);
}

function priorityText(priority: string): string {
  switch (priority) {
    case "strong": return "强匹配，优先投递";
    case "high": return "高匹配，重点关注";
    case "try": return "可尝试，需要补强";
    case "backup": return "备选参考，不建议优先";
    case "not_recommended": return "不推荐";
    default: return priority;
  }
}

export function toRecommendationPageViewModel(
  result: RecommendationResult,
  meta?: AgentResponseMeta,
  profileExtras?: { profileStatus?: string; profileType?: string; profileSource?: string; profileDominantProfile?: string; profileMessage?: string },
): RecommendationPageViewModel {
  const jobs = Array.isArray(result.recommended_jobs) ? result.recommended_jobs : [];
  const recommendedJobs = jobs.map((job, i) => toRecommendedJobViewModel(job, i + 1));
  const firstSource = recommendedJobs[0]?.source ?? "unknown";
  const firstJob = jobs[0];
  const isQuality = firstJob?.generated_by === "mock_quality" || firstJob?.generated_by === "llm_agent";

  return {
    recommendedJobs,
    summary: result.recommendation_summary ?? "",
    applyReadiness: result.apply_readiness ?? "",
    optimizationStopSignal: result.optimization_stop_signal ?? "",
    topFixItems: toTopFixItemViewModels(result.top_fix_items),
    evidenceChain: toEvidenceItemViewModels(result.evidence_chain),
    riskItems: Array.isArray(result.risk_items) ? result.risk_items : [],
    nextActions: Array.isArray(result.next_actions) ? result.next_actions : [],
    sourceLabel: getSourceLabel(firstSource, firstJob?.generated_by ?? "", isQuality),
    meta: meta ? toAgentMetaViewModel(meta) : undefined,
    profileStatus: profileExtras?.profileStatus ?? "ready",
    profileType: profileExtras?.profileType ?? "fast",
    profileSource: profileExtras?.profileSource ?? "rule",
    profileDominantProfile: profileExtras?.profileDominantProfile ?? "",
    profileMessage: profileExtras?.profileMessage ?? "",
  };
}
