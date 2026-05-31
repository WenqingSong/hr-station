// ─── 通用 Adapter 工具 ─────────────────────────────────────────────────────────

import type { AgentResponseMeta, BadgeTone, StatusBadge } from '../types';

/** 创建统一 StatusBadge */
export function toStatusBadge(label: string, tone: BadgeTone = 'gray'): StatusBadge {
  return { label, tone };
}

/** 数组转分隔文本 */
export function formatListText(items: string[], emptyText = '暂无'): string {
  return items.length > 0 ? items.join(' / ') : emptyText;
}

/**
 * Agent 来源 → 展示标签映射。
 * 基于 generated_by + agent_status 组合判断。
 */
export function formatAgentSourceLabel(meta?: AgentResponseMeta): StatusBadge {
  if (!meta) return { label: '状态未知', tone: 'gray' };

  const { generated_by, agent_status } = meta;

  if (generated_by === 'llm_agent' && agent_status === 'success') {
    return { label: 'LLM 生成', tone: 'blue' };
  }
  if (generated_by === 'llm_agent' && agent_status === 'repaired') {
    return { label: '已自动修复', tone: 'green' };
  }
  if (agent_status === 'fallback') {
    return { label: '兜底结果', tone: 'orange' };
  }
  if (agent_status === 'mock_success') {
    return { label: 'Mock 演示结果', tone: 'gray' };
  }
  if (agent_status === 'error') {
    return { label: '异常但可用', tone: 'red' };
  }
  return { label: '状态未知', tone: 'gray' };
}

/** 优先级 → tone 映射 */
export function priorityToTone(priority: string): BadgeTone {
  switch (priority) {
    case 'high':
      return 'orange';
    case 'medium':
      return 'blue';
    case 'low':
      return 'gray';
    default:
      return 'gray';
  }
}

/** 证据强度 → tone 映射 */
export function evidenceStrengthToTone(strength: string): BadgeTone {
  switch (strength) {
    case 'strong':
      return 'green';
    case 'medium':
      return 'blue';
    case 'weak':
      return 'orange';
    case 'none':
      return 'red';
    default:
      return 'gray';
  }
}

/** 证据强度 → 中文标签 */
export function evidenceStrengthToLabel(strength: string): string {
  switch (strength) {
    case 'strong':
      return '强';
    case 'medium':
      return '中';
    case 'weak':
      return '弱';
    case 'none':
      return '无';
    default:
      return strength || '未知';
  }
}

/** 优先级 → 中文标签 */
export function priorityToLabel(priority: string): string {
  switch (priority) {
    case 'high':
      return '高';
    case 'medium':
      return '中';
    case 'low':
      return '低';
    default:
      return priority || '未知';
  }
}
