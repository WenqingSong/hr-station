import { color, radius } from "../../tokens";

type Status = "success" | "warn" | "danger" | "info" | "neutral";

interface StatusTagProps {
  status: Status;
  label: string;
  dot?: boolean;
}

const statusMap: Record<Status, { bg: string; text: string }> = {
  success: { bg: color.successBg, text: color.successText },
  warn:    { bg: color.warnBg,    text: color.warnText },
  danger:  { bg: color.dangerBg,  text: color.dangerText },
  info:    { bg: color.blueLight, text: color.blueText },
  neutral: { bg: "#F2F3F5",       text: color.textSecondary },
};

/**
 * 语义状态标签
 *
 * 用法:
 *   <StatusTag status="success" label="建议投递" />
 *   <StatusTag status="warn"    label="建议优化后投递" dot />
 *   <StatusTag status="danger"  label="暂不建议投递" />
 */
export function StatusTag({ status, label, dot = true }: StatusTagProps) {
  const { bg, text } = statusMap[status];
  return (
    <span
      style={{
        display: "inline-flex",
        alignItems: "center",
        gap: 5,
        background: bg,
        color: text,
        borderRadius: radius.tagRound,
        padding: "3px 10px",
        fontSize: 12,
        fontWeight: 500,
        whiteSpace: "nowrap",
      }}
    >
      {dot && (
        <span
          style={{
            width: 6,
            height: 6,
            borderRadius: "50%",
            background: text,
            flexShrink: 0,
          }}
        />
      )}
      {label}
    </span>
  );
}

/**
 * 轻量文字标签（技能 tag、类别 tag 等）
 */
interface TextTagProps {
  label: string;
  style?: React.CSSProperties;
}

export function TextTag({ label, style }: TextTagProps) {
  return (
    <span
      style={{
        display: "inline-block",
        background: "#F2F3F5",
        color: color.textSecondary,
        borderRadius: radius.tag,
        padding: "2px 8px",
        fontSize: 12,
        ...style,
      }}
    >
      {label}
    </span>
  );
}
