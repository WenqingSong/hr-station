/**
 * Dev-Reboot-Fix-C: Unified error state for when the AI model is unavailable.
 *
 * Displayed when:
 * - Deep LLM call returns mock_agent / mock_quality / fallback=true
 * - HTTP 500 or business error from Agent endpoint
 * - Network error while requesting AI analysis
 */
import React from "react";

export interface AgentErrorStateProps {
  /** Brief error title for end users */
  title?: string;
  /** Error detail message */
  message?: string;
  /** Which feature was being requested */
  feature?: string;
  /** The technical generated_by value (not shown, used for debugging) */
  generatedBy?: string;
  /** Callback to retry the request */
  onRetry?: () => void;
  /** Optional additional class name */
  className?: string;
}

const AgentErrorState: React.FC<AgentErrorStateProps> = ({
  title = "AI 模型连接失败",
  message = "当前没有生成真实 AI 结果，请检查模型 Key、服务配置或稍后重试。",
  feature,
  generatedBy,
  onRetry,
  className,
}) => {
  return (
    <div
      className={className}
      style={{
        background: "#f8fafc",
        border: "1px solid #e2e8f0",
        borderRadius: 12,
        padding: "40px 32px",
        textAlign: "center",
        maxWidth: 560,
        margin: "24px auto",
      }}
    >
      {/* Icon */}
      <div
        style={{
          width: 56,
          height: 56,
          borderRadius: 28,
          background: "#fee2e2",
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          margin: "0 auto 16px",
        }}
      >
        <svg
          width="28"
          height="28"
          viewBox="0 0 24 24"
          fill="none"
          stroke="#dc2626"
          strokeWidth="2"
          strokeLinecap="round"
          strokeLinejoin="round"
        >
          <circle cx="12" cy="12" r="10" />
          <line x1="12" y1="8" x2="12" y2="12" />
          <line x1="12" y1="16" x2="12.01" y2="16" />
        </svg>
      </div>

      {/* Title */}
      <h3
        style={{
          fontSize: 18,
          fontWeight: 600,
          color: "#1e293b",
          margin: "0 0 8px",
        }}
      >
        {title}
      </h3>

      {/* Message */}
      <p
        style={{
          fontSize: 14,
          color: "#64748b",
          lineHeight: 1.6,
          margin: "0 0 8px",
        }}
      >
        {message}
      </p>

      {/* Technical detail — show only if available */}
      {feature && (
        <p
          style={{
            fontSize: 12,
            color: "#94a3b8",
            margin: "0 0 4px",
          }}
        >
          功能: {feature}
          {generatedBy ? ` · 来源: ${generatedBy}` : ""}
        </p>
      )}

      {/* CTA */}
      {onRetry && (
        <button
          onClick={onRetry}
          style={{
            marginTop: 20,
            padding: "8px 24px",
            background: "#3b82f6",
            color: "#fff",
            border: "none",
            borderRadius: 8,
            fontSize: 14,
            fontWeight: 500,
            cursor: "pointer",
          }}
          onMouseEnter={(e) =>
            (e.currentTarget.style.background = "#2563eb")
          }
          onMouseLeave={(e) =>
            (e.currentTarget.style.background = "#3b82f6")
          }
        >
          重新生成
        </button>
      )}
    </div>
  );
};

export default AgentErrorState;
