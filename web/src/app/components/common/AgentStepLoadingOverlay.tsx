import { useEffect, useState, useRef } from "react";
import { Loader2, CheckCircle2, XCircle, Circle } from "lucide-react";

// ─── Types ────────────────────────────────────────────────────────────

export type AgentLoadingStep = {
  key: string;
  label: string;
  description?: string;
};

type StepStatus = "pending" | "active" | "done" | "error";

interface AgentStepLoadingOverlayProps {
  open: boolean;
  title: string;
  description?: string;
  steps: AgentLoadingStep[];
  /** "loading" auto-advances; "success" completes all; "error" marks current error */
  status: "loading" | "success" | "error";
  errorText?: string;
  /** minimum visible time after success before calling onClose (ms) */
  successDelayMs?: number;
  onClose?: () => void;
}

// ─── Timer config ──────────────────────────────────────────────────────

const STEP_INTERVALS = [800, 1000, 1200]; // ms between steps

// ─── Component ─────────────────────────────────────────────────────────

export function AgentStepLoadingOverlay({
  open,
  title,
  description,
  steps,
  status,
  errorText,
  successDelayMs = 400,
  onClose,
}: AgentStepLoadingOverlayProps) {
  const [activeIdx, setActiveIdx] = useState(0);
  const [finished, setFinished] = useState(false);
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  // Reset on open
  useEffect(() => {
    if (open) {
      setActiveIdx(0);
      setFinished(false);
    }
  }, [open]);

  // Auto-advance during loading
  useEffect(() => {
    if (!open || status !== "loading") return;

    const advance = () => {
      setActiveIdx((prev) => {
        const next = prev + 1;
        if (next >= steps.length) return prev; // stay on last step
        return next;
      });
    };

    if (activeIdx < steps.length - 1) {
      const delay = STEP_INTERVALS[Math.min(activeIdx, STEP_INTERVALS.length - 1)];
      timerRef.current = setTimeout(advance, delay);
    }

    return () => {
      if (timerRef.current) clearTimeout(timerRef.current);
    };
  }, [open, status, activeIdx, steps.length]);

  // On success: mark all done, then close
  useEffect(() => {
    if (!open || status !== "success") return;
    setActiveIdx(steps.length);
    timerRef.current = setTimeout(() => {
      setFinished(true);
      onClose?.();
    }, successDelayMs);
    return () => {
      if (timerRef.current) clearTimeout(timerRef.current);
    };
  }, [open, status, steps.length, successDelayMs, onClose]);

  // Cleanup
  useEffect(() => {
    return () => {
      if (timerRef.current) clearTimeout(timerRef.current);
    };
  }, []);

  if (!open) return null;

  function stepStatus(idx: number): StepStatus {
    if (status === "error" && idx === activeIdx) return "error";
    if (idx < activeIdx) return "done";
    if (idx === activeIdx && status === "loading") return "active";
    if (status === "success" && idx === steps.length) return "done";
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
          <Loader2 className="w-10 h-10 animate-spin text-[#0052D9] mx-auto mb-3" />
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
        {status === "error" && errorText && (
          <div className="mt-4 pt-4 border-t border-slate-100">
            <p className="text-sm text-red-500 text-center">{errorText}</p>
          </div>
        )}

        {/* Footer hint */}
        {status === "loading" && (
          <p className="text-xs text-slate-400 text-center mt-5">
            深度推荐生成可能需要一点时间
          </p>
        )}
      </div>
    </div>
  );
}
