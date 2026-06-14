/**
 * Dev-Reboot-Fix-C: Unified real-agent result assertion.
 *
 * Rules:
 * - Fast: generated_by="fast_tag_engine" is valid.
 * - Deep: generated_by="llm_agent" is the only valid real AI result.
 * - mock_agent / mock_quality / fallback=true must be treated as errors.
 */

/**
 * RealAgentError is thrown when the API returned mock/fallback data
 * instead of real AI-generated results.
 */
export class RealAgentError extends Error {
  title = "AI 模型连接失败";
  status: number | undefined;
  code: string | number | undefined;
  feature: string;
  generatedBy: string | undefined;
  agentStatus: string | undefined;
  fallback: boolean | undefined;

  constructor(
    feature: string,
    reason: string,
    extra?: {
      status?: number;
      code?: string | number;
      generatedBy?: string;
      agentStatus?: string;
      fallback?: boolean;
    },
  ) {
    super(`${feature}: ${reason}`);
    this.name = "RealAgentError";
    this.feature = feature;
    this.status = extra?.status;
    this.code = extra?.code;
    this.generatedBy = extra?.generatedBy;
    this.agentStatus = extra?.agentStatus;
    this.fallback = extra?.fallback;
  }

  /** User-facing reason summary. */
  get displayReason(): string {
    if (this.generatedBy) return `模型返回标记为 ${this.generatedBy}`;
    if (this.fallback) return "后端自动降级到兜底结果";
    if (this.agentStatus) return `Agent 状态: ${this.agentStatus}`;
    return this.message;
  }
}

/** Check whether a response field means mock data. */
export function isMockGeneratedBy(
  generatedBy: string | undefined | null,
): boolean {
  if (!generatedBy) return false;
  // K15-P0: deep_fallback_agent is allowed as a valid Deep fallback, not an error
  return generatedBy === "mock_agent" || generatedBy === "mock_quality";
}

/** Check whether a generated_by value represents a Deep-tier result (LLM or fallback). */
export function isDeepTierGeneratedBy(
  generatedBy: string | undefined | null,
): boolean {
  if (!generatedBy) return false;
  return generatedBy === "llm_agent" || generatedBy === "mock_quality" || generatedBy === "deep_fallback_agent";
}

/**
 * Assert that a response (already parsed JSON body) contains real LLM results.
 *
 * Checks:
 * 1. generated_by must not be mock_agent or mock_quality
 * 2. fallback must not be true
 * 3. For deep tier, generated_by must be llm_agent (fast_tag_engine is for fast tier only)
 */
export function assertRealAgentResult<
  T extends { generated_by?: string; fallback?: boolean; agent_status?: string },
>(
  response: T | undefined | null,
  feature: string,
  opts?: { isDeep?: boolean; status?: number; code?: string | number },
): void {
  if (!response) {
    throw new RealAgentError(feature, "响应数据为空", {
      status: opts?.status,
      code: opts?.code,
    });
  }

  if (isMockGeneratedBy(response.generated_by)) {
    throw new RealAgentError(
      feature,
      `响应标记为 ${response.generated_by}`,
      {
        generatedBy: response.generated_by,
        agentStatus: response.agent_status,
        fallback: response.fallback,
        status: opts?.status,
        code: opts?.code,
      },
    );
  }

  if (response.fallback === true) {
    throw new RealAgentError(feature, "后端自动降级到兜底结果", {
      generatedBy: response.generated_by,
      agentStatus: response.agent_status,
      fallback: response.fallback,
      status: opts?.status,
      code: opts?.code,
    });
  }

  if (opts?.isDeep && response.generated_by !== "llm_agent" && response.generated_by !== "deep_fallback_agent" && response.generated_by !== "mock_quality") {
    throw new RealAgentError(
      feature,
      `深度分析期望 llm_agent / deep_fallback_agent，实际为 ${response.generated_by || "unknown"}`,
      {
        generatedBy: response.generated_by,
        agentStatus: response.agent_status,
        fallback: response.fallback,
        status: opts?.status,
        code: opts?.code,
      },
    );
  }
}

/**
 * Check whether a tier label signals an error state
 * (used in page header rendering).
 */
export function isErrorTier(
  status: string | undefined | null,
): boolean {
  if (!status) return false;
  return status === "fast_error" || status === "deep_error" || status === "latest_error";
}

/**
 * ContractError — backend response structure doesn't match frontend expectation.
 * MUST NOT be shown as "推荐结果加载失败" or network error.
 */
export class ContractError extends Error {
  feature: string;
  constructor(feature: string, detail: string) {
    super(`[${feature}] ${detail}`);
    this.name = "ContractError";
    this.feature = feature;
  }
}
