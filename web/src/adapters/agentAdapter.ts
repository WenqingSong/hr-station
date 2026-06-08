// ─── Agent 公共字段 Adapter ─────────────────────────────────────────────────────

import type {
  AgentResponseMeta,
  TopFixItem,
  EvidenceItem,
  BadgeTone,
} from '../types';
import {
  formatAgentSourceLabel,
  priorityToTone,
  priorityToLabel,
  evidenceStrengthToTone,
  evidenceStrengthToLabel,
} from './common';

// ─── ViewModel 类型 ────────────────────────────────────────────────────────────

export interface AgentMetaViewModel {
  agentName: string;
  sourceLabel: string;
  sourceTone: BadgeTone;
  statusLabel: string;
  fallbackReason?: string;
  errorCode?: string;
  repairApplied: boolean;
  schemaValid: boolean;
  requestId: string;
}

export interface TopFixItemViewModel {
  title: string;
  priorityLabel: string;
  priorityTone: BadgeTone;
  expectedImpact: string;
  needUserVerify: boolean;
  description?: string;
  /** K8-5-28-B */
  resumeModule?: string;
}

export interface EvidenceItemViewModel {
  jdRequirement: string;
  resumeEvidence: string;
  evidenceStrengthLabel: string;
  evidenceStrengthTone: BadgeTone;
  riskNote?: string;
  /** K8-5-28-B */
  resumeModule?: string;
  /** K9-2-Regression-F-B: fix action from backend */
  fixAction?: string;
}

// ─── 转换函数 ──────────────────────────────────────────────────────────────────

export function toAgentMetaViewModel(meta: AgentResponseMeta): AgentMetaViewModel {
  const source = formatAgentSourceLabel(meta);
  return {
    agentName: meta.agent_name,
    sourceLabel: source.label,
    sourceTone: source.tone,
    statusLabel: meta.agent_status,
    fallbackReason:
      meta.fallback_reason && meta.fallback_reason !== 'none'
        ? meta.fallback_reason
        : undefined,
    errorCode:
      meta.error_code && meta.error_code !== 'none' ? meta.error_code : undefined,
    repairApplied: meta.repair_applied,
    schemaValid: meta.schema_valid,
    requestId: meta.request_id,
  };
}

export function toTopFixItemViewModel(item: TopFixItem): TopFixItemViewModel {
  // K9-2-Regression-C: Map backend JSON fields (issue/why_it_matters/action/etc) to ViewModel
  const title = item.title || item.issue || "";
  const desc = item.description || item.why_it_matters || "";
  const impact = item.expected_impact || item.estimated_effort || (item.action || "").slice(0, 40);
  const priorityVal = item.priority || "P1";
  const needVerify = item.need_user_verify ?? false;
  return {
    title,
    priorityLabel: priorityToLabel(priorityVal),
    priorityTone: priorityToTone(priorityVal),
    expectedImpact: impact,
    needUserVerify: needVerify,
    description: desc,
    resumeModule: item.resume_module,
  };
}

export function toTopFixItemViewModels(items: TopFixItem[]): TopFixItemViewModel[] {
  if (!Array.isArray(items)) return [];
  return items.map(toTopFixItemViewModel);
}

export function toEvidenceItemViewModel(item: EvidenceItem): EvidenceItemViewModel {
  // K9-2-Regression-C: Map backend JSON fields (requirement/resume_point/judgement) to ViewModel
  const jdReq = item.jd_requirement || item.requirement || "";
  const resumeEv = item.resume_evidence || item.resume_point || "";
  const strength = item.evidence_strength || mapJudgementToStrength(item.judgement || "");
  const risk = item.risk_note || item.risk || "";
  const fixAction = item.fix_action || "";
  return {
    jdRequirement: jdReq,
    resumeEvidence: resumeEv,
    evidenceStrengthLabel: evidenceStrengthToLabel(strength),
    evidenceStrengthTone: evidenceStrengthToTone(strength),
    riskNote: risk,
    resumeModule: item.resume_module,
    fixAction: fixAction,
  };
}

// Map backend "匹配"/"部分匹配"/"缺口" to evidence strength
function mapJudgementToStrength(judgement: string): string {
  switch (judgement) {
  case "匹配": return "strong";
  case "部分匹配": return "partial";
  case "缺口": return "missing";
  case "待评估": return "insufficient";
  default: return "insufficient";
  }
}

export function toEvidenceItemViewModels(
  items: EvidenceItem[],
): EvidenceItemViewModel[] {
  if (!Array.isArray(items)) return [];
  return items.map(toEvidenceItemViewModel);
}
