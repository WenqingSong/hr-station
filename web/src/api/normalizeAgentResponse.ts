/**
 * Frontend-Contract-Fix-B: Unified normalize for Agent API responses.
 *
 * Backend APIs may return fields at either:
 *   - data.result.xxx   (nested)
 *   - data.xxx           (top-level, result=null)
 *
 * This function normalizes both into a single flat object.
 */
import { ContractError } from "../app/utils/agentError";

export interface NormalizedAgentData {
  [key: string]: unknown;
  generated_by?: string;
  agent_status?: string;
  fallback?: boolean;
  meta?: Record<string, unknown>;
}

/**
 * Normalize an Agent API response `data` field.
 * - If data.result is a non-null object, merge it into data.
 * - If data.result is null/absent, use data top-level fields.
 * - Preserve generated_by / fallback / agent_status / meta.
 * - Throw ContractError if both are empty.
 */
export function normalizeAgentData(
  data: Record<string, unknown> | null | undefined,
  feature: string,
): Record<string, unknown> {
  if (!data || typeof data !== "object") {
    throw new ContractError(feature, "响应数据为空或不是对象");
  }

  const result = data.result;

  // Case 1: result is a non-null object — merge into data (result fields win)
  if (result && typeof result === "object" && !Array.isArray(result)) {
    return {
      ...data,
      ...(result as Record<string, unknown>),
      // Preserve top-level meta / generated_by / fallback / agent_status
      meta: data.meta,
      generated_by: data.generated_by ?? (data.meta as any)?.generated_by,
      agent_status: data.agent_status ?? (data.meta as any)?.agent_status,
      fallback: data.fallback ?? (data.meta as any)?.fallback,
    };
  }

  // Case 2: data has recognized agent fields at top level — use as-is
  return data;
}

/**
 * Convenience: normalize and assert key array is present.
 */
export function normalizeAgentArrayField(
  data: Record<string, unknown> | null | undefined,
  feature: string,
  arrayKey: string,
): unknown[] {
  const norm = normalizeAgentData(data, feature);
  const arr = norm[arrayKey];
  if (!Array.isArray(arr)) {
    throw new ContractError(feature, `${arrayKey} 不是数组`);
  }
  return arr;
}
