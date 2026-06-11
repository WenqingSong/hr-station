import { useEffect, useState, useRef } from "react";
import { Loader2, CheckCircle2, XCircle, Circle } from "lucide-react";

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
  /** K13-Stability-A: max total duration before auto-finish (ms). Default 90s to stay under backend's 120s LLM timeout. */
  maxDurationMs?: number;
  /** Error state — shows error icon and text */
  error?: boolean;
  errorText?: string;

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
const DEFAULT_MAX_DURATION_MS = 90000; // 90s — under backend LLM_TIMEOUT_SECONDS=120

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
  error,
  errorText,

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

  if (!open) return null;

  const effectiveError = isLegacy ? legacyStatus === "error" : (error ?? false);

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
      <div className="bg-white/95 rounded-2xl shadow-xl border border-white/60 p-8 max-w-md w-[380px] mx-4">
        {/* Header */}
        <div className="text-center mb-6">
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

        {/* Steps */}
        <div className="space-y-3">
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

        {/* Footer hint */}
        {!finished && !effectiveError && (
          <p className="text-xs text-slate-400 text-center mt-5">
            深度生成正在进行中…
          </p>
        )}
      </div>
    </div>
  );
}
