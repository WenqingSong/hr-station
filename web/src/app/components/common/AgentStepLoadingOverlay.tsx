import { useEffect, useState, useRef } from "react";
import { Loader2, CheckCircle2, XCircle, Circle, Clock } from "lucide-react";

// ─── Types ────────────────────────────────────────────────────────────

export type AgentLoadingStep = {
  key: string;
  label: string;
  description?: string;
};

type StepStatus = "pending" | "active" | "done" | "error";

export interface AgentStepLoadingOverlayProps {
  open: boolean;
  title: string;
  description?: string;
  steps: AgentLoadingStep[];
  /** When true, backend is done — finish remaining steps gracefully */
  backendDone?: boolean;
  /** Called after all steps complete + success hold */
  onFinish?: () => void;
  /** Minimum visible time (ms) to avoid instant-close on fast API */
  minVisibleMs?: number;
  /** Interval between normal step advances (ms) */
  stepIntervalMs?: number;
  /** Interval between finishing steps after backendDone (ms) */
  finishStepIntervalMs?: number;
  /** Hold time after all steps done (ms) */
  successHoldMs?: number;
  /** K13-Stability-A: max total duration before auto-finish (ms). Default 15min for Deep Cascade. */
  maxDurationMs?: number;
  /** Estimated time text displayed to user */
  estimatedTimeText?: string;
  /** Error state — shows error icon and text */
  error?: boolean;
  errorText?: string;
  /** Cancel callback — shows cancel link below steps */
  onCancel?: () => void;
  cancelLabel?: string;
  /** L3: externally-supplied elapsed seconds */
  elapsedSeconds?: number;

  // ── @deprecated: kept for backward compat ──
  status?: "loading" | "success" | "error";
  successDelayMs?: number;
  onClose?: () => void;
}

// ─── Timer config ──────────────────────────────────────────────────────

const DEFAULT_STEP_INTERVAL = 650;
const DEFAULT_FINISH_STEP_INTERVAL = 260;
const DEFAULT_MIN_VISIBLE = 1800;
const DEFAULT_SUCCESS_HOLD = 420;
const DEFAULT_MAX_DURATION_MS = 900000; // 15 min — 覆盖 Deep Cascade 全链路（Layer 0-4）

// ─── Component ─────────────────────────────────────────────────────────

export function AgentStepLoadingOverlay({
  open,
  title,
  description,
  steps,
  backendDone: backendDoneProp,
  onFinish,
  minVisibleMs = DEFAULT_MIN_VISIBLE,
  stepIntervalMs: stepIntervalMsProp,
  finishStepIntervalMs: finishStepIntervalMsProp,
  successHoldMs: successHoldMsProp,
  maxDurationMs = DEFAULT_MAX_DURATION_MS,
  estimatedTimeText,
  error,
  errorText,
  onCancel,
  cancelLabel,
  elapsedSeconds: externalElapsed,

  // @deprecated compat
  status: legacyStatus,
  successDelayMs,
  onClose,
}: AgentStepLoadingOverlayProps) {
  const [activeIdx, setActiveIdx] = useState(0);
  const [finished, setFinished] = useState(false);
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const startedAtRef = useRef(0);
  const backendDoneDoneRef = useRef(false);
  // L3: display elapsed — state must be declared before any early return (Hook rules)
  const [displayElapsed, setDisplayElapsed] = useState(0);

  const stepIntervalMs = stepIntervalMsProp ?? DEFAULT_STEP_INTERVAL;
  const finishStepIntervalMs = finishStepIntervalMsProp ?? DEFAULT_FINISH_STEP_INTERVAL;
  const successHoldMs = successHoldMsProp ?? DEFAULT_SUCCESS_HOLD;

  // Legacy compat: derive backendDone from status
  const isLegacy = legacyStatus !== undefined;
  const backendDone = isLegacy
    ? (legacyStatus === "success")
    : (backendDoneProp ?? false);

  // Reset on open
  useEffect(() => {
    if (open) {
      setActiveIdx(0);
      setFinished(false);
      startedAtRef.current = Date.now();
      backendDoneDoneRef.current = false;
    }
  }, [open]);

  // Handle backendDone: finish remaining steps sequentially
  useEffect(() => {
    if (!open || !backendDone || finished || backendDoneDoneRef.current) return;
    backendDoneDoneRef.current = true;

    // Ensure min visible time has elapsed
    const elapsed = Date.now() - startedAtRef.current;
    const remaining = Math.max(0, minVisibleMs - elapsed);

    const startFinish = () => {
      let idx = activeIdx;
      // Don't read stale state — use ref if needed
      const timer = () => {
        setActiveIdx((prev) => {
          const next = prev + 1;
          if (next >= steps.length) {
            // All done — hold then finish
            setTimeout(() => {
              setFinished(true);
              // Legacy compat
              if (isLegacy && onClose) onClose();
              if (!isLegacy && onFinish) onFinish();
            }, successHoldMs);
            return prev; // stay at last
          }
          // Schedule next finishing step
          setTimeout(timer, finishStepIntervalMs);
          return next;
        });
      };
      // Start advancing from current position
      setTimeout(timer, 0);
    };

    if (remaining > 0) {
      timerRef.current = setTimeout(startFinish, remaining);
    } else {
      startFinish();
    }

    return () => {
      if (timerRef.current) clearTimeout(timerRef.current);
    };
  }, [open, backendDone, steps.length, minVisibleMs, finishStepIntervalMs, successHoldMs, finished, isLegacy, onClose, onFinish]);

  // Auto-advance during loading (before backendDone)
  useEffect(() => {
    if (!open || backendDone || finished) return;

    const maxStep = Math.max(0, steps.length - 2); // never auto-complete
    const advance = () => {
      setActiveIdx((prev) => {
        const next = prev + 1;
        if (next > maxStep) return prev;
        return next;
      });
    };

    if (activeIdx < maxStep) {
      timerRef.current = setTimeout(advance, stepIntervalMs);
    }

    return () => {
      if (timerRef.current) clearTimeout(timerRef.current);
    };
  }, [open, backendDone, finished, activeIdx, steps.length, stepIntervalMs]);

  // K13-Stability-A: Safety timeout — if backend never responds, auto-finish after maxDurationMs
  useEffect(() => {
    if (!open || backendDone || finished) return;
    const maxTimer = setTimeout(() => {
      if (!backendDone && !finished) {
        // Force finish: mark all steps done and call onFinish
        setActiveIdx(steps.length - 1);
        setFinished(true);
        setTimeout(() => {
          if (onFinish) onFinish();
        }, successHoldMs);
      }
    }, maxDurationMs);
    return () => clearTimeout(maxTimer);
  }, [open, backendDone, finished, maxDurationMs, steps.length, successHoldMs, onFinish]);

  // Cleanup
  useEffect(() => {
    return () => {
      if (timerRef.current) clearTimeout(timerRef.current);
    };
  }, []);

  // L3: tick elapsed every second (needs to be above early return for Hook rules)
  useEffect(() => {
    const elapsed = externalElapsed ?? (startedAtRef.current ? Math.floor((Date.now() - startedAtRef.current) / 1000) : 0);
    setDisplayElapsed(elapsed);
    if (!open || finished) return;
    const ticker = setInterval(() => {
      setDisplayElapsed(Math.floor((Date.now() - startedAtRef.current) / 1000));
    }, 1000);
    return () => clearInterval(ticker);
  }, [open, finished, externalElapsed]);

  if (!open) return null;

  const effectiveError = isLegacy ? legacyStatus === "error" : (error ?? false);

  // L3: computed elapsed time (external or internal)
  const elapsedSec = externalElapsed ?? (startedAtRef.current ? Math.floor((Date.now() - startedAtRef.current) / 1000) : 0);

  // L3: progress percentage (0-96% auto, 100% on done)
  const progressPct = finished ? 100 : backendDone ? 98 : Math.min(96, Math.floor((displayElapsed / 150) * 96));
  const displayPct = Math.min(progressPct, finished ? 100 : 96);

  // L3: long-wait prompt
  function getWaitHint(): string {
    if (displayElapsed > 180) return "本次分析耗时较长，可能正在进行首次画像构建或模型响应较慢。";
    if (displayElapsed > 60) return "深度推荐需要逐岗生成证据和风险点，请继续等待…";
    return "";
  }
  const waitHint = !finished && !backendDone ? getWaitHint() : "";

  function formatTime(sec: number): string {
    const m = Math.floor(sec / 60);
    const s = sec % 60;
    if (m > 0) return `${m}:${s.toString().padStart(2, "0")}`;
    return `0:${s.toString().padStart(2, "0")}`;
  }

  function stepStatus(idx: number): StepStatus {
    if (effectiveError && idx === activeIdx) return "error";
    if (idx < activeIdx) return "done";
    if (idx === activeIdx && !finished) return "active";
    if (finished && idx === activeIdx) return "done";
    return "pending";
  }

  function StepIcon({ s }: { s: StepStatus }) {
    switch (s) {
      case "done":
        return <CheckCircle2 className="w-4 h-4 text-emerald-500" />;
      case "active":
        return <Loader2 className="w-4 h-4 text-[#0052D9] animate-spin" />;
      case "error":
        return <XCircle className="w-4 h-4 text-red-500" />;
      default:
        return <Circle className="w-4 h-4 text-slate-300" />;
    }
  }

  return (
    <div className="fixed inset-0 z-50 bg-slate-900/20 backdrop-blur-md flex items-center justify-center">
      <div className="bg-white/95 rounded-2xl shadow-xl border border-white/60 p-8 max-w-md w-[420px] mx-4">
        {/* Header */}
        <div className="text-center mb-5">
          {finished ? (
            <CheckCircle2 className="w-10 h-10 text-emerald-500 mx-auto mb-3" />
          ) : (
            <Loader2 className="w-10 h-10 animate-spin text-[#0052D9] mx-auto mb-3" />
          )}
          <h2 className="text-[17px] font-bold text-slate-900 mb-1">{title}</h2>
          {description && (
            <p className="text-sm text-slate-500 leading-relaxed">{description}</p>
          )}
        </div>

        {/* L3: Progress bar */}
        {!finished && !effectiveError && (
          <div className="mb-4">
            <div className="flex items-center justify-between text-[11px] text-slate-400 mb-1">
              <span className="flex items-center gap-1"><Clock className="w-3 h-3" />已用 {formatTime(displayElapsed)}</span>
              <span>{displayPct}%</span>
            </div>
            <div className="w-full h-1.5 bg-slate-100 rounded-full overflow-hidden">
              <div className="h-full bg-[#0052D9] rounded-full transition-all duration-1000 ease-out" style={{ width: `${displayPct}%` }} />
            </div>
            {estimatedTimeText && (
              <p className="text-[10px] text-slate-400 text-center mt-1.5">{estimatedTimeText}</p>
            )}
          </div>
        )}

        {/* Steps */}
        <div className="space-y-2.5">
          {steps.map((step, idx) => {
            const s = stepStatus(idx);
            const isActive = s === "active";
            return (
              <div key={step.key} className="flex items-start gap-3">
                <StepIcon s={s} />
                <div className="flex-1 min-w-0">
                  <p className={`text-sm font-medium transition-colors ${
                    s === "done" ? "text-slate-500" :
                    s === "error" ? "text-red-600" :
                    isActive ? "text-slate-900" :
                    "text-slate-400"
                  }`}>
                    {step.label}
                  </p>
                  {step.description && isActive && (
                    <p className="text-xs text-slate-400 mt-0.5">{step.description}</p>
                  )}
                </div>
              </div>
            );
          })}
        </div>

        {/* Error message */}
        {effectiveError && errorText && (
          <div className="mt-4 pt-4 border-t border-slate-100">
            <p className="text-sm text-red-500 text-center">{errorText}</p>
          </div>
        )}

        {/* Footer: wait hint + progress note + cancel */}
        {!finished && !effectiveError && (
          <div className="mt-5 text-center space-y-2">
            {waitHint && (
              <p className="text-[11px] text-amber-600 bg-amber-50/50 rounded-lg px-3 py-1.5">{waitHint}</p>
            )}
            <p className="text-[10px] text-slate-400">
              当前为预估进度，最终结果以分析完成后为准
            </p>
            {onCancel && (
              <button
                onClick={onCancel}
                className="text-xs text-slate-400 hover:text-slate-600 underline transition-colors mt-1"
              >
                {cancelLabel || "返回快速推荐结果"}
              </button>
            )}
          </div>
        )}
      </div>
    </div>
  );
}
