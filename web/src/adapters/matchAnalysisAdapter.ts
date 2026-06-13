import type { MatchAnalysisResult, AgentResponseMeta } from '../types';

// ─── JM-R1: Screening Match V1 Types ──────────────────────────────────

export type ScreeningDecisionVM = {
  score: number; level: string; decision: string; decisionLabel: string;
  summary: string; confidence: string; decisionReasons: string[];
};

export type ScreeningPassPointVM = {
  title: string; whyItHelps: string; resumeEvidence: string; screeningValue: string;
};

export type ScreeningEliminationRiskVM = {
  riskTitle: string; severity: "high" | "medium" | "low" | string;
  whyItBlocksScreening: string; currentResumeProblem: string; fixDirection: string;
};

export type ResumeRewriteCommandVM = {
  targetSection: string; problem: string; rewriteInstruction: string;
  beforeExample: string; afterExample: string; whyItImprovesScreening: string;
};

export type JobStrategyVM = {
  positioning: string; recommendedAction: string; ifApplyNow: string;
  betterFitDirections: string[]; avoidDirections: string[];
};

// ─── ViewModel ────────────────────────────────────────────────────────

export interface MatchAnalysisPageViewModel {
  matchScore: number; matchLevel: string; summary: string;
  generatedBy?: string; isProfilePending: boolean; profilePendingReason: string;
  decisionLabel?: string;
  // screening_match_v1
  screeningDecision: ScreeningDecisionVM | null;
  screeningPassPoints: ScreeningPassPointVM[];
  screeningEliminationRisks: ScreeningEliminationRiskVM[];
  resumeRewriteCommands: ResumeRewriteCommandVM[];
  jobStrategy: JobStrategyVM | null;
  // legacy
  reportVersion?: string; isDeepLegacyFallback: boolean; isDeepMockFallback: boolean;
  profileDominantProfile?: string;
}

// ─── Helpers ──────────────────────────────────────────────────────────

function safeArr<T>(v: unknown, fallback: T[] = []): T[] { return Array.isArray(v) ? (v as T[]) : fallback; }
function safeStr(v: unknown, fallback = ""): string { return typeof v === "string" ? v : fallback; }

function safeDecision(raw: any): ScreeningDecisionVM | null {
  const sd = raw?.screening_decision || raw;
  if (!sd || sd.score === undefined) return null;
  return {
    score: sd.score ?? 0, level: safeStr(sd.level, "—"), decision: safeStr(sd.decision, "not_recommended"),
    decisionLabel: safeStr(sd.decision_label, "—"), summary: safeStr(sd.summary, ""),
    confidence: safeStr(sd.confidence, "medium"), decisionReasons: safeArr(sd.decision_reasons, []),
  };
}

function safePassPoints(raw: any): ScreeningPassPointVM[] {
  const arr = safeArr(raw?.screening_pass_points || raw?.pass_points, []);
  return arr.map((p: any) => ({ title: safeStr(p.title), whyItHelps: safeStr(p.why_it_helps || p.whyItHelps), resumeEvidence: safeStr(p.resume_evidence || p.resumeEvidence), screeningValue: safeStr(p.screening_value || p.screeningValue) }));
}

function safeRisks(raw: any): ScreeningEliminationRiskVM[] {
  const arr = safeArr(raw?.screening_elimination_risks || raw?.elimination_risks, []);
  return arr.map((r: any) => ({ riskTitle: safeStr(r.risk_title || r.riskTitle), severity: safeStr(r.severity, "medium"), whyItBlocksScreening: safeStr(r.why_it_blocks_screening || r.whyItBlocksScreening), currentResumeProblem: safeStr(r.current_resume_problem || r.currentResumeProblem), fixDirection: safeStr(r.fix_direction || r.fixDirection) }));
}

function safeCommands(raw: any): ResumeRewriteCommandVM[] {
  const arr = safeArr(raw?.resume_rewrite_commands || raw?.rewrite_commands, []);
  return arr.map((c: any) => ({ targetSection: safeStr(c.target_section || c.targetSection), problem: safeStr(c.problem), rewriteInstruction: safeStr(c.rewrite_instruction || c.rewriteInstruction), beforeExample: safeStr(c.before_example || c.beforeExample), afterExample: safeStr(c.after_example || c.afterExample), whyItImprovesScreening: safeStr(c.why_it_improves_screening || c.whyItImprovesScreening) }));
}

function safeStrategy(raw: any): JobStrategyVM | null {
  const js = raw?.job_strategy || raw;
  if (!js || !js.positioning) return null;
  return { positioning: safeStr(js.positioning), recommendedAction: safeStr(js.recommended_action || js.recommendedAction), ifApplyNow: safeStr(js.if_apply_now || js.ifApplyNow), betterFitDirections: safeArr(js.better_fit_directions || js.betterFitDirections, []), avoidDirections: safeArr(js.avoid_directions || js.avoidDirections, []) };
}

// ─── Conversion ───────────────────────────────────────────────────────

export function toMatchAnalysisPageViewModel(result: MatchAnalysisResult, meta?: AgentResponseMeta): MatchAnalysisPageViewModel {
  const raw = result as any;
  const genBy = meta?.generated_by || raw?.generated_by || "";
  const agentStatus = meta?.agent_status || raw?.agent_status || "";
  const isPending = genBy === "fast_sep_profile_pending" || agentStatus === "profile_missing" || agentStatus === "profile_building" || agentStatus === "profile_invalid";

  const sd = safeDecision(raw);
  return {
    matchScore: isPending ? 0 : (sd?.score ?? result.match_score),
    matchLevel: isPending ? "待分析" : (sd?.level ?? result.match_level ?? "—"),
    summary: isPending ? "语义画像尚未准备完成。" : (sd?.summary ?? result.summary ?? ""),
    generatedBy: genBy, isProfilePending: isPending,
    profilePendingReason: isPending ? (agentStatus === "profile_building" ? "语义画像正在构建中" : agentStatus === "profile_invalid" ? "语义画像状态异常" : "语义画像尚未准备") : "",
    decisionLabel: sd?.decisionLabel ?? safeStr(raw?.decision_label, ""),
    screeningDecision: sd,
    screeningPassPoints: safePassPoints(raw),
    screeningEliminationRisks: safeRisks(raw),
    resumeRewriteCommands: safeCommands(raw),
    jobStrategy: safeStrategy(raw),
    reportVersion: safeStr(raw?.report_version, ""),
    isDeepLegacyFallback: genBy === "deep_match_legacy_fallback" || agentStatus === "legacy_fallback",
    isDeepMockFallback: genBy === "mock_quality" || agentStatus === "mock_fallback",
    profileDominantProfile: result.profile_dominant_profile || "",
  };
}
