import { color } from "../../tokens";

interface ScoreRingProps {
  score: number;
  size?: number;
  strokeWidth?: number;
  label?: string;
}

/**
 * 环形匹配度分数组件
 * 颜色随分数自动变化：>=80 蓝色 / >=60 橙色 / <60 红色
 */
export function ScoreRing({
  score,
  size = 64,
  strokeWidth = 6,
  label = "匹配度",
}: ScoreRingProps) {
  const r = (size - strokeWidth) / 2;
  const circumference = 2 * Math.PI * r;
  const dash = (score / 100) * circumference;
  const scoreColor =
    score >= 80 ? color.blue : score >= 60 ? color.warn : color.danger;

  return (
    <div
      style={{
        position: "relative",
        width: size,
        height: size,
        flexShrink: 0,
      }}
    >
      <svg width={size} height={size} viewBox={`0 0 ${size} ${size}`}>
        <circle
          cx={size / 2}
          cy={size / 2}
          r={r}
          fill="none"
          stroke={color.borderLight}
          strokeWidth={strokeWidth}
        />
        <circle
          cx={size / 2}
          cy={size / 2}
          r={r}
          fill="none"
          stroke={scoreColor}
          strokeWidth={strokeWidth}
          strokeDasharray={`${dash} ${circumference - dash}`}
          strokeDashoffset={circumference * 0.25}
          strokeLinecap="round"
        />
      </svg>
      <div
        style={{
          position: "absolute",
          inset: 0,
          display: "flex",
          flexDirection: "column",
          alignItems: "center",
          justifyContent: "center",
        }}
      >
        <span
          style={{ fontSize: size * 0.25, fontWeight: 700, color: scoreColor, lineHeight: 1 }}
        >
          {score}
        </span>
        {label && (
          <span style={{ fontSize: size * 0.14, color: color.textTertiary, marginTop: 2 }}>
            {label}
          </span>
        )}
      </div>
    </div>
  );
}
