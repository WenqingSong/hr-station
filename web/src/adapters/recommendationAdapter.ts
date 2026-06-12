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
import { humanizeOneSentenceReason, humanizeMatchHighlights, humanizeRiskPoints, humanizeTagList, isRawTagCode } from './tagLabelMap';

// ─── ViewModel 类型 ────────────────────────────────────────────────────────────

export type RecommendationSource =
  | "mock" | "fast_tag" | "fast_llm" | "deep_llm" | "fallback" | "cache" | "unknown";

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
  // Deep Cascade fields
  displayTags: string[];
  applicationStrategy: string;
  candidateCoreStrength: string;
  fitConfidenceReason: string;
}

export interface RecommendationEvidenceRow {
  module: string;
  jobRequirement: string;
  resumeEvidence: string;
  strength: "strong" | "medium" | "weak";
}

export interface RecommendationPageViewModel {
  mode: "fast" | "deep";
  recommendedJobs: RecommendedJobViewModel[];
  globalSummary: string;
  summary: string;
  applyReadiness: string;
  optimizationStopSignal: string;
  topFixItems: TopFixItemViewModel[];
  evidenceChain: EvidenceItemViewModel[];
  riskItems: string[];
  nextActions: string[];

  sourceLabel: string;
  meta?: AgentMetaViewModel;

  // Deep Cascade / quality fields
  agentStatus: string;
  qualityStatus: string;
  candidatePositioning: string;

  // K9-1-I: profile status fields
  profileStatus: string;
  profileType: string;
  profileSource: string;
  profileDominantProfile: string;
  profileMessage: string;
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

/** 规范化字符串：trim + 统一全角/半角空格 */
function normalizeText(value: unknown): string {
  if (value === null || value === undefined) return "";
  if (typeof value !== "string") return String(value);
  return value.replace(/[\u3000]/g, " ").replace(/\s+/g, " ").trim();
}

/** 判断两个规范化后的文本是否相同（忽略空白差异） */
function isSameText(a?: string, b?: string): boolean {
  return normalizeText(a) === normalizeText(b);
}

/** 对字符串数组去重（规范化后比较） */
function dedupeTextList(values: unknown[]): string[] {
  if (!Array.isArray(values)) return [];
  const seen = new Set<string>();
  const result: string[] = [];
  for (const v of values) {
    const norm = normalizeText(v);
    if (norm.length > 0 && !seen.has(norm)) {
      seen.add(norm);
      result.push(norm);
    }
  }
  return result;
}

/** 判断是否有意义的文本（非空、非模板残骸） */
function isMeaningfulText(value: unknown): boolean {
  const s = normalizeText(value);
  if (s.length === 0) return false;
  if (s === "null" || s === "undefined" || s === "NaN") return false;
  // 过滤模板残留
  if (/^(标签匹配|缺失标签|建议进一步)/.test(s)) return false;
  return true;
}

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
  // Dev-Reboot-Fix-C: mock_quality is mock data, not real LLM
  if (generatedBy === "fast_tag_engine") return "fast_tag";
  if (generatedBy === "mock_agent") return "mock";
  if (generatedBy === "mock_quality") return "mock";
  if (generatedBy === "llm_agent") return "deep_llm";
  if (generatedBy.includes("fallback")) return "fallback";
  return "unknown";
}

function getSourceLabel(source: RecommendationSource, generatedBy?: string, isQuality?: boolean): string {
  // K9-1-T-D: product-friendly labels — 用户永远看不到 mock/fallback 内部标记
  if (generatedBy === "fast_tag_engine") return "快速推荐 · 标签匹配";
  if (generatedBy === "mock_agent") return "快速推荐 · 基础模型";
  if (generatedBy === "mock_quality") return "深度推荐 · 基础模型";
  if (generatedBy === "llm_agent" || generatedBy?.startsWith("llm")) return "深度推荐 · AI 生成";
  if (generatedBy?.includes("fallback")) return "推荐生成失败 · 已保留当前结果";
  switch (source) {
    case "fast_tag": return "快速推荐 · 标签匹配";
    case "mock": return isQuality ? "深度推荐 · 基础模型" : "快速推荐 · 基础模型";
    case "fast_llm": return "快速推荐 · AI 生成";
    case "deep_llm": return "深度推荐 · AI 生成";
    case "fallback": return "推荐生成失败 · 已保留当前结果";
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

  // ── Field deduplication ──
  const rawRecommendationReason = job.recommendation_reason || "";
  const rawJobSpecificReason = job.job_specific_reason || "";

  // K9-1-T-D: new explanation fields with safe fallbacks
  const whyRankedHere = job.why_ranked_here || humanizeOneSentenceReason(rawRecommendationReason);
  const resumeEvidence = humanizeTagList(
    (Array.isArray(job.resume_evidence) && job.resume_evidence.length > 0)
      ? job.resume_evidence
      : []
  );

  // jobSpecificReason: if same as recommendation_reason, keep only the humanized version
  let jobSpecificReason = rawJobSpecificReason ? humanizeOneSentenceReason(rawJobSpecificReason) : humanizeOneSentenceReason(rawRecommendationReason);

  const rawRisks = (Array.isArray(job.job_specific_risks) && job.job_specific_risks.length > 0)
    ? job.job_specific_risks
    : (Array.isArray(job.risk_points) ? job.risk_points : []);
  const jobSpecificRisks = humanizeRiskPoints(rawRisks);

  const rawSuggestedAction = job.suggested_action || "";
  const rawSuggestedImprovement = job.suggested_improvement || "";

  // L1-fix: prioritize suggested_improvement over suggested_action for the "补强建议" slot
  const suggestedImprovement = isMeaningfulText(rawSuggestedImprovement) && !isSameText(rawSuggestedImprovement, rawSuggestedAction)
    ? rawSuggestedImprovement
    : "";

  // applicationStrategy: if identical to suggestedAction, show only one
  const applicationStrategy = isMeaningfulText(rawSuggestedAction)
    ? rawSuggestedAction
    : (isMeaningfulText(suggestedImprovement) ? suggestedImprovement : "");

  const displayTags = humanizeTagList(
    (Array.isArray(job.display_tags) && job.display_tags.length > 0) ? job.display_tags : []
  );

  // L1-fix: oneSentenceReason — only from recommendation_reason
  const oneSentenceReason = humanizeOneSentenceReason(rawRecommendationReason);

  // L1-fix: if jobSpecificReason === oneSentenceReason after humanize, suppress duplicate
  if (isSameText(jobSpecificReason, oneSentenceReason)) {
    jobSpecificReason = "";
  }

  // L1-fix: riskPoints — only from risk_points, not duplicating jobSpecificRisks
  const rawRiskPoints = Array.isArray(job.risk_points) ? job.risk_points : [];
  const riskPoints = humanizeRiskPoints(rawRiskPoints);

  // L1-fix: if riskPoints all appear in jobSpecificRisks, clear riskPoints
  const riskSet = new Set(riskPoints.map(normalizeText));
  const allRisksCovered = jobSpecificRisks.length > 0 && riskPoints.length > 0
    && jobSpecificRisks.every((r) => riskSet.has(normalizeText(r)));
  const effectiveRiskPoints = allRisksCovered ? [] : riskPoints;

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
    oneSentenceReason,
    matchHighlights: humanizeMatchHighlights(job.match_highlights || []),
    riskPoints: effectiveRiskPoints,
    suggestedAction: job.suggested_action || "建议先查看岗位详情，再进行岗位匹配分析。",
    source,
    sourceLabel: getSourceLabel(source, generatedBy, isQuality),
    // K9-1-T-D: new fields (L1-dedup)
    whyRankedHere,
    resumeEvidence,
    jobSpecificReason,
    jobSpecificRisks,
    suggestedImprovement,
    // Deep Cascade fields (L1-dedup)
    displayTags,
    applicationStrategy,
    candidateCoreStrength: job.candidate_core_strength || "",
    fitConfidenceReason: job.fit_confidence_reason || "",
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
  const mode: "fast" | "deep" = isQuality ? "deep" : "fast";

  // Deep Cascade: extract global_summary, agent_status, quality_status, candidate_positioning
  const globalSummary = (result as any).result?.global_summary
    || (result as any).result?.summary
    || (result as any).global_summary
    || result.recommendation_summary
    || "";
  const qualityStatus = (result as any).result?.quality_status || "";
  const agentStatus = firstJob?.agent_status ?? (result as any).meta?.agent_status ?? "";
  const candidatePositioning = (result as any).result?.candidate_positioning || "";

  return {
    mode,
    recommendedJobs,
    globalSummary,
    summary: result.recommendation_summary ?? "",
    applyReadiness: result.apply_readiness ?? "",
    optimizationStopSignal: result.optimization_stop_signal ?? "",
    topFixItems: toTopFixItemViewModels(result.top_fix_items),
    evidenceChain: toEvidenceItemViewModels(result.evidence_chain),
    riskItems: Array.isArray(result.risk_items) ? result.risk_items : [],
    nextActions: Array.isArray(result.next_actions) ? result.next_actions : [],
    sourceLabel: getSourceLabel(firstSource, firstJob?.generated_by ?? "", isQuality),
    meta: meta ? toAgentMetaViewModel(meta) : undefined,
    agentStatus,
    qualityStatus,
    candidatePositioning,
    profileStatus: profileExtras?.profileStatus ?? "ready",
    profileType: profileExtras?.profileType ?? "fast",
    profileSource: profileExtras?.profileSource ?? "rule",
    profileDominantProfile: profileExtras?.profileDominantProfile ?? "",
    profileMessage: profileExtras?.profileMessage ?? "",
  };
}
