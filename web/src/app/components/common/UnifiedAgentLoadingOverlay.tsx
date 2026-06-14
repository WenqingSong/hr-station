import { useState, useEffect, useRef } from "react";
import { Clock } from "lucide-react";

// ─── Keyframes ─────────────────────────────────────────────────────────
const KEYFRAMES_STYLE = `
@keyframes ual-breathe {
  0%, 100% { transform: scale(1); opacity: 0.55; }
  50% { transform: scale(1.07); opacity: 1; }
}
@keyframes ual-orbit {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}
@keyframes ual-pulse-dot {
  0%, 100% { opacity: 0.4; transform: scale(0.8); }
  50% { opacity: 1; transform: scale(1.2); }
}
`;

// ─── Types ──────────────────────────────────────────────────────────────

export interface UnifiedAgentLoadingOverlayProps {
  open: boolean;
  title: string;
  description?: string;
  /** Elapsed seconds from external timer */
  elapsedSeconds?: number;
  /** Label shown next to elapsed time */
  phaseLabel?: string;
  /** When true, backend work is done — finish gracefully */
  backendDone?: boolean;
  /** Called after all steps complete + success hold */
  onFinish?: () => void;
  /** Called when user clicks cancel/back */
  onCancel?: () => void;
  cancelLabel?: string;
  /** Footer hint text below steps */
  footerHint?: string;
  /** Success hold time in ms before calling onFinish */
  successHoldMs?: number;
}

// ─── Helpers ────────────────────────────────────────────────────────────

function formatTime(sec: number): string {
  const m = Math.floor(sec / 60);
  const s = sec % 60;
  if (m > 0) return `${m}:${s.toString().padStart(2, "0")}`;
  return `0:${s.toString().padStart(2, "0")}`;
}

// ─── Component ───────────────────────────────────────────────────────────

export function UnifiedAgentLoadingOverlay({
  open,
  title,
  description,
  elapsedSeconds = 0,
  phaseLabel,
  backendDone: backendDoneProp,
  onFinish,
  onCancel,
  cancelLabel,
  footerHint,
  successHoldMs = 800,
}: UnifiedAgentLoadingOverlayProps) {
  const [internalElapsed, setInternalElapsed] = useState(0);
  const [finished, setFinished] = useState(false);
  const startedAtRef = useRef(Date.now());
  const finishTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const backendDone = backendDoneProp ?? false;
  const displayElapsed = elapsedSeconds || internalElapsed;

  // Reset on open
  useEffect(() => {
    if (!open) return;
    startedAtRef.current = Date.now();
    setInternalElapsed(0);
    setFinished(false);
    if (finishTimerRef.current) clearTimeout(finishTimerRef.current);
  }, [open]);

  // Elapsed ticker
  useEffect(() => {
    if (!open || finished) return;
    const ticker = setInterval(() => {
      setInternalElapsed(Math.floor((Date.now() - startedAtRef.current) / 1000));
    }, 1000);
    return () => clearInterval(ticker);
  }, [open, finished]);

  // Handle backendDone → finish after hold
  useEffect(() => {
    if (!open || !backendDone || finished) return;
    finishTimerRef.current = setTimeout(() => {
      setFinished(true);
      if (onFinish) onFinish();
    }, successHoldMs);
    return () => { if (finishTimerRef.current) clearTimeout(finishTimerRef.current); };
  }, [open, backendDone, finished, successHoldMs, onFinish]);

  if (!open) return null;

  const phaseText = phaseLabel || (backendDone ? "分析完成" : "分析中");

  return (
    <>
      <style>{KEYFRAMES_STYLE}</style>
      <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/20 backdrop-blur-md">
        {/* Card — blue-tech style matching DeepAgentLoadingOverlay */}
        <div className="w-[520px] max-w-[calc(100vw-32px)] rounded-3xl border border-blue-100/80 bg-gradient-to-br from-white via-white to-blue-50/70 shadow-2xl shadow-blue-950/10 p-8">
          {/* ── Center Orb ────────────────────────────────────────── */}
          <div className="flex justify-center mb-6">
            <div className="relative w-[96px] h-[96px]">
              {/* Outer breathing ring */}
              <div className="absolute inset-0 rounded-full bg-blue-400/10" style={{ animation: "ual-breathe 2.4s ease-in-out infinite" }} />
              {/* Middle rotating ring */}
              <div className="absolute inset-2 rounded-full border-2 border-transparent border-t-blue-400 border-r-blue-300" style={{ animation: "ual-orbit 3s linear infinite" }} />
              {/* Inner rotating ring (reverse) */}
              <div className="absolute inset-4 rounded-full border border-transparent border-b-cyan-300 border-l-cyan-200" style={{ animation: "ual-orbit 2s linear infinite reverse" }} />
              {/* Core glow dot */}
              <div className="absolute inset-6 rounded-full bg-gradient-to-br from-blue-500 to-cyan-400 shadow-lg shadow-blue-400/40 flex items-center justify-center">
                <div className="w-3 h-3 rounded-full bg-white/80" style={{ animation: "ual-pulse-dot 2s ease-in-out infinite" }} />
              </div>
            </div>
          </div>

          {/* ── Title ─────────────────────────────────────────────── */}
          <h2 className="text-[18px] font-bold text-slate-800 text-center mb-1.5">{title}</h2>
          {description && (
            <p className="text-[13px] text-slate-500 text-center leading-relaxed mb-4">{description}</p>
          )}

          {/* ── Elapsed ───────────────────────────────────────────── */}
          <div className="flex items-center justify-center gap-2 mb-5 text-[13px] text-slate-500">
            <Clock className="w-3.5 h-3.5" />
            <span className="font-mono tabular-nums">已用时 {formatTime(displayElapsed)}</span>
            <span className="text-slate-300">·</span>
            <span className="flex items-center gap-1.5">
              <span className="w-1.5 h-1.5 rounded-full bg-blue-400 inline-block" style={{ animation: "ual-pulse-dot 1.5s ease-in-out infinite" }} />
              {phaseText}
            </span>
          </div>

          {/* ── Footer ─────────────────────────────────────────────── */}
          <div className="mt-5 text-center space-y-2">
            {footerHint && <p className="text-[11px] text-slate-400">{footerHint}</p>}
            {!footerHint && !backendDone && (
              <p className="text-[11px] text-slate-400">完成后将自动展示结果，请耐心等待。</p>
            )}
            {onCancel && (
              <button onClick={onCancel} className="text-sm text-slate-400 hover:text-slate-600 underline underline-offset-2 transition-colors">
                {cancelLabel || "返回"}
              </button>
            )}
          </div>
        </div>
      </div>
    </>
  );
}
