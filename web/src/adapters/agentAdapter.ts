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
}

export interface EvidenceItemViewModel {
  jdRequirement: string;
  resumeEvidence: string;
  evidenceStrengthLabel: string;
  evidenceStrengthTone: BadgeTone;
  riskNote?: string;
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
  return {
    title: item.title,
    priorityLabel: priorityToLabel(item.priority),
    priorityTone: priorityToTone(item.priority),
    expectedImpact: item.expected_impact,
    needUserVerify: item.need_user_verify,
    description: item.description,
  };
}

export function toTopFixItemViewModels(items: TopFixItem[]): TopFixItemViewModel[] {
  if (!Array.isArray(items)) return [];
  return items.map(toTopFixItemViewModel);
}

export function toEvidenceItemViewModel(item: EvidenceItem): EvidenceItemViewModel {
  return {
    jdRequirement: item.jd_requirement,
    resumeEvidence: item.resume_evidence,
    evidenceStrengthLabel: evidenceStrengthToLabel(item.evidence_strength),
    evidenceStrengthTone: evidenceStrengthToTone(item.evidence_strength),
    riskNote: item.risk_note,
  };
}

export function toEvidenceItemViewModels(
  items: EvidenceItem[],
): EvidenceItemViewModel[] {
  if (!Array.isArray(items)) return [];
  return items.map(toEvidenceItemViewModel);
}
