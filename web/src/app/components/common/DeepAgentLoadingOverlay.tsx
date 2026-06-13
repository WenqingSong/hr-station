import { useState, useEffect, useRef } from "react";
import { Clock, Loader2 } from "lucide-react";

// ─── Types ──────────────────────────────────────────────────────────────

export interface DeepAgentLoadingOverlayProps {
  open: boolean;
  /** Elapsed seconds (external timer) */
  elapsedSeconds?: number;
  /** Called when user clicks "返回快速推荐结果" */
  onBackToFast?: () => void;
}

// ─── CSS keyframes injected via <style> ─────────────────────────────────

const KEYFRAMES_STYLE = `
@keyframes h17c-breathe {
  0%, 100% { transform: scale(1); opacity: 0.55; }
  50% { transform: scale(1.07); opacity: 1; }
}
@keyframes h17c-orbit {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}
@keyframes h17c-shimmer {
  0% { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}
@keyframes h17c-pulse-dot {
  0%, 100% { opacity: 0.4; transform: scale(0.8); }
  50% { opacity: 1; transform: scale(1.2); }
}
`;

// ─── Helpers ────────────────────────────────────────────────────────────

function formatTime(sec: number): string {
  const m = Math.floor(sec / 60);
  const s = sec % 60;
  if (m > 0) return `${m}:${s.toString().padStart(2, "0")}`;
  return `0:${s.toString().padStart(2, "0")}`;
}

const AGENTS = [
  { key: "discovery", label: "候选发现", tip: "从岗位池筛选高潜岗位" },
  { key: "rerank", label: "语义重排", tip: "结合简历画像重新排序" },
  { key: "synthesis", label: "逐岗分析", tip: "生成推荐理由、证据和风险" },
  { key: "calibration", label: "全局校准", tip: "校准最终顺序与展示口径" },
];

const TOP5_LABELS = ["岗位 1", "岗位 2", "岗位 3", "岗位 4", "岗位 5"];

// ─── Component ───────────────────────────────────────────────────────────

export function DeepAgentLoadingOverlay({
  open,
  elapsedSeconds = 0,
  onBackToFast,
}: DeepAgentLoadingOverlayProps) {
  // Internal elapsed fallback
  const [internalElapsed, setInternalElapsed] = useState(0);
  const startedAtRef = useRef(Date.now());

  useEffect(() => {
    if (!open) return;
    startedAtRef.current = Date.now();
    setInternalElapsed(0);
    const ticker = setInterval(() => {
      setInternalElapsed(
        Math.floor((Date.now() - startedAtRef.current) / 1000)
      );
    }, 1000);
    return () => clearInterval(ticker);
  }, [open]);

  const displayElapsed = elapsedSeconds || internalElapsed;

  if (!open) return null;

  return (
    <>
      {/* Inject keyframes once */}
      <style>{KEYFRAMES_STYLE}</style>

      {/* Backdrop */}
      <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/20 backdrop-blur-md">
        {/* Card */}
        <div className="w-[520px] max-w-[calc(100vw-32px)] rounded-3xl border border-blue-100/80 bg-gradient-to-br from-white via-white to-blue-50/70 shadow-2xl shadow-blue-950/10 p-8">
          {/* ── Center Orb ────────────────────────────────────────── */}
          <div className="flex justify-center mb-6">
            <div className="relative w-[96px] h-[96px]">
              {/* Outer breathing ring */}
              <div
                className="absolute inset-0 rounded-full bg-blue-400/10"
                style={{ animation: "h17c-breathe 2.4s ease-in-out infinite" }}
              />
              {/* Middle rotating ring */}
              <div
                className="absolute inset-2 rounded-full border-2 border-transparent border-t-blue-400 border-r-blue-300"
                style={{ animation: "h17c-orbit 3s linear infinite" }}
              />
              {/* Inner rotating ring (reverse) */}
              <div
                className="absolute inset-4 rounded-full border border-transparent border-b-cyan-300 border-l-cyan-200"
                style={{ animation: "h17c-orbit 2s linear infinite reverse" }}
              />
              {/* Core glow dot */}
              <div className="absolute inset-6 rounded-full bg-gradient-to-br from-blue-500 to-cyan-400 shadow-lg shadow-blue-400/40 flex items-center justify-center">
                <div
                  className="w-3 h-3 rounded-full bg-white/80"
                  style={{ animation: "h17c-pulse-dot 2s ease-in-out infinite" }}
                />
              </div>
              {/* Orbiting nodes */}
              {[0, 1, 2].map((i) => (
                <div
                  key={i}
                  className="absolute w-2 h-2 rounded-full bg-blue-300/70"
                  style={{
                    top: "50%",
                    left: "50%",
                    animation: `h17c-orbit ${4 + i * 1.5}s linear infinite`,
                    transformOrigin: `-${40 + i * 6}px 0`,
                    animationDelay: `${i * 0.5}s`,
                  }}
                />
              ))}
            </div>
          </div>

          {/* ── Title ─────────────────────────────────────────────── */}
          <h2 className="text-[18px] font-bold text-slate-800 text-center mb-1.5">
            深度推荐智能体正在分析
          </h2>
          <p className="text-[13px] text-slate-500 text-center leading-relaxed mb-4">
            正在基于简历画像、岗位池和岗位证据进行多 Agent 协同分析
          </p>

          {/* ── Elapsed ───────────────────────────────────────────── */}
          <div className="flex items-center justify-center gap-2 mb-5 text-[13px] text-slate-500">
            <Clock className="w-3.5 h-3.5" />
            <span className="font-mono tabular-nums">
              已用时 {formatTime(displayElapsed)}
            </span>
            <span className="text-slate-300">·</span>
            <span className="flex items-center gap-1.5">
              <span
                className="w-1.5 h-1.5 rounded-full bg-blue-400 inline-block"
                style={{ animation: "h17c-pulse-dot 1.5s ease-in-out infinite" }}
              />
              Top5 岗位分析中
            </span>
          </div>

          {/* ── Agent Chain ────────────────────────────────────────── */}
          <div className="mb-5">
            <p className="text-[11px] font-semibold text-slate-400 uppercase tracking-wide text-center mb-2.5">
              Agent 协同链路
            </p>
            <div className="flex flex-wrap justify-center gap-2">
              {AGENTS.map((agent, idx) => (
                <span
                  key={agent.key}
                  title={agent.tip}
                  className={`inline-flex items-center gap-1.5 px-3 py-1.5 rounded-full text-[11px] font-medium border ${
                    idx === 2
                      ? "bg-blue-100/80 text-blue-700 border-blue-200"
                      : "bg-slate-50 text-slate-500 border-slate-200"
                  }`}
                  style={{
                    backgroundSize: "200% 100%",
                    backgroundImage:
                      idx === 2
                        ? "linear-gradient(90deg, transparent 0%, rgba(59,130,246,0.08) 50%, transparent 100%)"
                        : "linear-gradient(90deg, transparent 0%, rgba(148,163,184,0.06) 50%, transparent 100%)",
                    animation: "h17c-shimmer 2.5s ease-in-out infinite",
                    animationDelay: `${idx * 0.2}s`,
                  }}
                >
                  <span
                    className="w-1 h-1 rounded-full bg-current inline-block"
                    style={{ animation: "h17c-pulse-dot 2s ease-in-out infinite" }}
                  />
                  {agent.label}
                </span>
              ))}
            </div>
          </div>

          {/* ── Top5 Job Matrix ────────────────────────────────────── */}
          <div className="mb-2">
            <p className="text-[11px] font-semibold text-slate-400 uppercase tracking-wide text-center mb-2.5">
              Top5 岗位分析矩阵 · 统一生成中
            </p>
            <div className="grid grid-cols-5 gap-2">
              {TOP5_LABELS.map((label, idx) => (
                <div
                  key={label}
                  className="bg-slate-50 border border-slate-100 rounded-xl px-2 py-2.5 text-center"
                  style={{
                    animation: "h17c-shimmer 3s ease-in-out infinite",
                    animationDelay: `${idx * 0.15}s`,
                    backgroundSize: "200% 100%",
                    backgroundImage:
                      "linear-gradient(90deg, transparent 0%, rgba(148,163,184,0.05) 50%, transparent 100%)",
                  }}
                >
                  <p className="text-[11px] font-medium text-slate-500 mb-1">
                    {label}
                  </p>
                  <span
                    className="inline-block w-1.5 h-1.5 rounded-full bg-blue-400"
                    style={{ animation: "h17c-pulse-dot 1.5s ease-in-out infinite" }}
                  />
                  <span className="text-[10px] text-slate-400 ml-1">
                    分析中
                  </span>
                </div>
              ))}
            </div>
          </div>

          {/* ── Footer ─────────────────────────────────────────────── */}
          <div className="mt-5 text-center space-y-2">
            <p className="text-[11px] text-slate-400">
              深度推荐通常需要几十秒到数分钟，完成后将自动展示结果。
            </p>
            {onBackToFast && (
              <button
                onClick={onBackToFast}
                className="text-sm text-slate-400 hover:text-slate-600 underline underline-offset-2 transition-colors"
              >
                返回快速推荐结果
              </button>
            )}
          </div>
        </div>
      </div>
    </>
  );
}
