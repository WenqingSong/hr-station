// ─── Agent 通用元数据类型 ────────────────────────────────────────────────────────

/** 生成来源 */
export type AgentGeneratedBy = 'llm_agent' | 'mock_agent' | string;

/** Agent 运行状态 */
export type AgentStatus =
  | 'success'
  | 'repaired'
  | 'fallback'
  | 'mock_success'
  | 'error'
  | string;

/** Fallback 触发原因 */
export type AgentFallbackReason =
  | 'none'
  | 'llm_disabled'
  | 'llm_error'
  | 'llm_timeout'
  | 'llm_api_key_missing'
  | 'json_parse_failed'
  | 'json_repair_failed'
  | 'schema_invalid'
  | 'pipeline_error'
  | 'prompt_load_failed'
  | 'unknown'
  | string;

/** Agent 错误码 */
export type AgentErrorCode =
  | 'none'
  | 'LLM_DISABLED'
  | 'LLM_ERROR'
  | 'LLM_TIMEOUT'
  | 'JSON_PARSE_FAILED'
  | 'JSON_REPAIR_FAILED'
  | 'SCHEMA_INVALID'
  | 'PIPELINE_ERROR'
  | 'MOCK_FALLBACK_USED'
  | 'VALIDATION_ERROR'
  | 'INTERNAL_ERROR'
  | 'UNKNOWN_ERROR'
  | string;

/** Agent 响应统一 meta 字段（8 字段） */
export interface AgentResponseMeta {
  agent_name: string;
  generated_by: AgentGeneratedBy;
  agent_status: AgentStatus;
  fallback_reason: AgentFallbackReason;
  error_code: AgentErrorCode;
  repair_applied: boolean;
  schema_valid: boolean;
  request_id: string;
}

/**
 * Agent 返回的 data 内层结构。
 * 后端统一契约：data.result + data.meta
 */
export interface AgentResponseEnvelope<TResult> {
  result: TResult;
  meta: AgentResponseMeta;
}

// ─── 通用业务枚举 ──────────────────────────────────────────────────────────────

/** 投递准备度 */
export type ApplyReadiness =
  | 'ready_to_apply'
  | 'fix_before_apply'
  | 'quick_fix_before_apply'
  | 'not_recommended'
  | string;

/** 优化停止信号 */
export type OptimizationStopSignal =
  | 'stop_and_apply'
  | 'fix_top_items_first'
  | 'fix_top_3_then_apply'
  | 'collect_more_evidence'
  | string;

/** 证据强度 */
export type EvidenceStrength = 'strong' | 'medium' | 'weak' | 'none' | string;

/** 优先级 */
export type PriorityLevel = 'high' | 'medium' | 'low' | string;

// ─── K9-3 新版优化类型 ───────────────────────────────────────────────────────────

export type OptimizationMode = "general" | "target_direction" | "target_job";
export type GenerationTier = "fast" | "quality";

export interface OptimizationFixItem {
  section: string;
  issue: string;
  action: string;
}

export interface OptimizationRewriteExample {
  section: string;
  before: string;
  after: string;
  reason: string;
}

export interface OptimizationLongTermItem {
  direction: string;
  note: string;
}

/** K9-3 V2: New optimization result format */
export interface OptimizationResultV2 {
  optimization_id?: number;
  resume_id?: number;
  job_id?: number;
  optimization_mode?: OptimizationMode;
  target_direction?: string;
  generation_tier?: GenerationTier;
  generated_by?: string;
  status?: string;
  profile_type?: string;
  profile_dominant_profile?: string;

  optimization_score: number;
  optimization_level: string;
  one_sentence_summary: string;

  quick_fix_items: OptimizationFixItem[];
  structure_fix_items: OptimizationFixItem[];
  content_fix_items: OptimizationFixItem[];
  keyword_fix_items: OptimizationFixItem[];
  evidence_fix_items: OptimizationFixItem[];
  risk_points: string[];
  rewrite_examples: OptimizationRewriteExample[];
  delivery_focus: string;
  long_term_improvements: OptimizationLongTermItem[];
}

/** K9-3 Latest response envelope */
export interface OptimizationLatestResponse {
  has_result: boolean;
  optimization_id?: number;
  resume_id?: number;
  job_id?: number;
  optimization_mode?: OptimizationMode;
  target_direction?: string;
  generation_tier?: GenerationTier;
  generated_by?: string;
  status?: string;
  profile_type?: string;
  profile_dominant_profile?: string;
  created_at?: string;
  result?: OptimizationResultV2;
}

// ─── 共享业务类型 ───────────────────────────────────────────────────────────────

/** Top 快速修复任务 */
export interface TopFixItem {
  title: string;
  priority: PriorityLevel;
  expected_impact: string;
  need_user_verify: boolean;
  description?: string;
  /** K8-5-28-B: resume module (skills/projects/internships/...) */
  resume_module?: string;
}

/** JD 证据链条目 */
export interface EvidenceItem {
  jd_requirement: string;
  resume_evidence: string;
  evidence_strength: EvidenceStrength;
  risk_note?: string;
  /** K8-5-28-B: resume module (skills/projects/certificates/...) */
  resume_module?: string;
}

// ─── 三类 Agent Result ─────────────────────────────────────────────────────────

/** 推荐岗位项 — aligned with backend /api/jobs/recommend response */
export interface RecommendedJobItem {
  job_id: number;
  title: string;
  department?: string;
  city?: string;
  category?: string;
  job_type?: string;

  recommendation_score?: number;
  recommendation_reason?: string;
  match_highlights?: string[];
  risk_points?: string[];
  suggested_action?: string;

  // K9-1-T-D: per-job explanation fields
  why_ranked_here?: string;
  resume_evidence?: string[];
  job_specific_reason?: string;
  job_specific_risks?: string[];
  suggested_improvement?: string;

  external_source?: string;
  generated_by?: string;
  agent_status?: string;
  priority?: string;
  is_featured?: boolean;
  recommendation_label?: string;
  recommendation_priority?: string;
}

/** 岗位推荐 Agent Result */
export interface RecommendationResult {
  recommended_jobs: RecommendedJobItem[];
  recommendation_summary: string;
  apply_readiness: ApplyReadiness;
  optimization_stop_signal: OptimizationStopSignal;
  top_fix_items: TopFixItem[];
  evidence_chain: EvidenceItem[];
  risk_items: string[];
  next_actions: string[];
}

/** 岗位匹配分析 Agent Result */
export interface MatchAnalysisResult {
  match_score: number;
  match_level: string;
  summary: string;
  apply_readiness: ApplyReadiness;
  optimization_stop_signal: OptimizationStopSignal;
  top_fix_items: TopFixItem[];
  evidence_chain: EvidenceItem[];
  match_highlights: string[];
  risk_items: string[];
  next_actions: string[];
}

/** JD 关键词对齐建议 */
export interface KeywordAlignmentItem {
  keyword: string;
  current_coverage: string;
  suggested_action: string;
}

/** 项目经历优化建议 */
export interface ProjectOptimizationItem {
  project_name: string;
  current_problem: string;
  suggested_rewrite: string;
  example_text: string;
  need_user_verify: boolean;
}

/** 简历修改前后对比示例 */
export interface RewriteExample {
  before: string;
  suggestion: string;
  after: string;
  verification_notice: string;
}

/** K8-5-28-B: JD 证据链条目（优化 Agent） — 对齐后端 ResumeOptimizationJDEvidenceItem */
export interface JDEvidenceChainItem {
  jd_requirement: string;
  resume_evidence: string;
  evidence_strength: EvidenceStrength;
  match_judgement: string;
  gap_analysis?: string;
  screening_risk?: string;
}

/** 简历优化 Agent Result */
export interface ResumeOptimizationResult {
  summary: string;
  apply_readiness: ApplyReadiness;
  optimization_stop_signal: OptimizationStopSignal;
  top_fix_items: TopFixItem[];
  keyword_alignment: KeywordAlignmentItem[];
  project_optimization: ProjectOptimizationItem[];
  missing_evidence_plan: string[];
  rewrite_examples: RewriteExample[];
  verification_notice: string;
  final_checklist: string[];
  /** K8-5-28-B: JD evidence chain (优化 Agent) */
  jd_evidence_chain: JDEvidenceChainItem[];
}
