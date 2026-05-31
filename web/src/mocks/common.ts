// ─── 共享 Mock 工具 ────────────────────────────────────────────────────────────

import type { AgentResponseMeta } from '../types';

/**
 * 创建统一的 Mock AgentResponseMeta。
 * 所有 Mock Agent 响应共享相同的基础字段。
 */
export function createMockAgentMeta(agentName: string, requestId: string): AgentResponseMeta {
  return {
    agent_name: agentName,
    generated_by: 'mock_agent',
    agent_status: 'mock_success',
    fallback_reason: 'none',
    error_code: 'none',
    repair_applied: false,
    schema_valid: true,
    request_id: requestId,
  };
}
