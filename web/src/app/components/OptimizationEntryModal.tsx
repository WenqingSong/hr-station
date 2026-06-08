import React, { useEffect, useState, useRef, useCallback } from "react";
import { createPortal } from "react-dom";
import { useNavigate } from "react-router";
import { X, Sparkles, ArrowRight, ChevronRight, FileText, Briefcase } from "lucide-react";
import { loadResumeOptionViewModelsFromApi } from "../../api";
import type { ResumeOptionViewModel } from "../../adapters";
import type { JobApiItem } from "../../api/jobs";
import { getAuthRedirectIf401 } from "../utils/authRedirect";
import { JobPicker } from "./JobPicker";

// ─── Types ────────────────────────────────────────────────────────────────────

export interface OptimizationEntryModalProps {
  open: boolean;
  onClose: () => void;
  initialResumeId?: number | null;
  initialJobId?: number | null;
}

const TITLE_ID = "optimization-entry-title";

// ─── Component ────────────────────────────────────────────────────────────────

export function OptimizationEntryModal({
  open,
  onClose,
  initialResumeId = null,
  initialJobId = null,
}: OptimizationEntryModalProps) {
  const navigate = useNavigate();
  const overlayRef = useRef<HTMLDivElement>(null);
  const [step, setStep] = useState<1 | 2>(1);

  // Resume
  const [resumes, setResumes] = useState<ResumeOptionViewModel[]>([]);
  const [loadingResumes, setLoadingResumes] = useState(false);
  const [selectedResume, setSelectedResume] = useState<ResumeOptionViewModel | null>(null);

  // Job
  const [selectedJob, setSelectedJob] = useState<JobApiItem | null>(null);

  // ── Body scroll lock ──
  useEffect(() => {
    if (!open) return;
    const prevOverflow = document.body.style.overflow;
    const prevPaddingRight = document.body.style.paddingRight;
    const scrollbarWidth = window.innerWidth - document.documentElement.clientWidth;
    document.body.style.overflow = "hidden";
    if (scrollbarWidth > 0) document.body.style.paddingRight = `${scrollbarWidth}px`;
    return () => {
      document.body.style.overflow = prevOverflow;
      document.body.style.paddingRight = prevPaddingRight;
    };
  }, [open]);

  // ── ESC close ──
  useEffect(() => {
    if (!open) return;
    const onKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, [open, onClose]);

  // ── Click overlay to close ──
  const handleOverlayMouseDown = useCallback(
    (e: React.MouseEvent) => {
      if (e.target === overlayRef.current) onClose();
    },
    [onClose],
  );

  // ── Load resumes on open ──
  useEffect(() => {
    if (!open) {
      setSelectedResume(null);
      setSelectedJob(null);
      setStep(1);
      return;
    }
    setLoadingResumes(true);
    loadResumeOptionViewModelsFromApi()
      .then((opts) => {
        setResumes(opts);
        if (initialResumeId != null) {
          const matched = opts.find((o) => o.id === initialResumeId);
          if (matched) {
            setSelectedResume(matched);
            setStep(2); // skip to job selection
          }
        }
      })
      .catch((err) => {
        const authRedirect = getAuthRedirectIf401(err, window.location.pathname, window.location.search);
        if (authRedirect) navigate(authRedirect, { replace: true });
      })
      .finally(() => setLoadingResumes(false));
  }, [open, initialResumeId]);

  // ── Confirm ──
  const canConfirm = selectedResume != null && selectedJob != null;

  const handleConfirm = () => {
    if (!canConfirm) return;
    navigate(`/resume-optimization?resumeId=${selectedResume!.id}&jobId=${selectedJob!.id}`);
    onClose();
  };

  const handleGeneralOptimization = () => {
    if (!selectedResume) return;
    navigate(`/resume-optimization?resumeId=${selectedResume.id}`);
    onClose();
  };

  // ── Select resume ──
  const handleSelectResume = (resume: ResumeOptionViewModel) => {
    setSelectedResume(resume);
    setStep(2);
  };

  if (!open) return null;

  return createPortal(
    <div className="fixed inset-0 z-[10000]" ref={overlayRef}>
      {/* ── Overlay ── */}
      <div
        className="absolute inset-0 bg-slate-950/40 backdrop-blur-sm"
        onMouseDown={handleOverlayMouseDown}
      />

      {/* ── Panel ── */}
      <div className="relative z-[10001] flex min-h-full items-center justify-center px-4 py-8">
        <section
          role="dialog"
          aria-modal="true"
          aria-labelledby={TITLE_ID}
          className="flex w-full max-w-[640px] flex-col overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-2xl"
          style={{ maxHeight: "min(680px, calc(100vh - 64px))" }}
        >
          {/* ── Header ── */}
          <header className="shrink-0 border-b border-slate-100 px-6 pt-5 pb-4">
            <div className="flex items-start justify-between">
              <div>
                <h2 id={TITLE_ID} className="font-semibold text-slate-900 text-base flex items-center gap-2">
                  <Sparkles className="w-4 h-4 text-[#0052D9]" />
                  选择简历与目标岗位
                </h2>
                <p className="text-[13px] text-slate-500 mt-1">
                  系统将基于所选简历与目标岗位 JD 生成定向优化建议
                </p>
              </div>
              <button
                onClick={onClose}
                className="p-1.5 text-slate-300 hover:text-slate-500 transition-colors flex-shrink-0 ml-3 rounded-lg hover:bg-slate-100"
                aria-label="关闭"
              >
                <X className="w-4 h-4" />
              </button>
            </div>

            {/* Step indicator */}
            <div className="flex items-center gap-2 mt-4">
              <div className={`flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-medium transition-colors ${
                step === 1 ? "bg-blue-50 text-[#0052D9]" : selectedResume ? "bg-emerald-50 text-emerald-600" : "bg-slate-100 text-slate-400"
              }`}>
                <FileText className="w-3 h-3" />
                {selectedResume ? selectedResume.name : "选择简历"}
                {selectedResume && step === 1 && (
                  <button onClick={() => setStep(1)} className="ml-1 text-[10px] underline">修改</button>
                )}
              </div>
              <ChevronRight className="w-3 h-3 text-slate-300" />
              <div className={`flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-medium transition-colors ${
                step === 2 ? "bg-blue-50 text-[#0052D9]" : selectedJob ? "bg-emerald-50 text-emerald-600" : "bg-slate-100 text-slate-400"
              }`}>
                <Briefcase className="w-3 h-3" />
                {selectedJob ? selectedJob.title : "选择目标岗位"}
                {selectedJob && step === 2 && (
                  <span className="text-[10px] ml-1">已选</span>
                )}
              </div>
            </div>
          </header>

          {/* ── Body ── */}
          <div className="min-h-0 flex-1 overflow-y-auto px-6 py-4">
            {step === 1 ? (
              /* ── Step 1: Select Resume ── */
              <div className="space-y-1.5">
                {loadingResumes ? (
                  <div className="py-10 text-center">
                    <p className="text-[13px] text-slate-400">正在读取你的简历…</p>
                  </div>
                ) : resumes.length === 0 ? (
                  <div className="py-10 text-center">
                    <p className="text-[13px] text-slate-500 mb-1">你还没有可用于优化的简历</p>
                    <p className="text-[12px] text-slate-400 mb-3">请先创建一份简历。</p>
                    <button
                      onClick={() => { navigate("/resumes/new"); onClose(); }}
                      className="text-[13px] text-[#0052D9] hover:underline font-medium"
                    >
                      去添加简历
                    </button>
                  </div>
                ) : (
                  resumes.map((resume) => (
                    <div
                      key={resume.id}
                      onClick={() => handleSelectResume(resume)}
                      className="flex items-center gap-3 px-4 py-3 border border-slate-200 rounded-xl cursor-pointer hover:border-blue-200 hover:bg-slate-50 transition-all"
                    >
                      <div className="w-4 h-4 rounded-full border-2 border-slate-300" />
                      <div className="flex-1 min-w-0">
                        <span className="text-[14px] font-semibold text-slate-900">{resume.name}</span>
                        {resume.directionLabel && (
                          <span className="text-[12px] text-slate-500 ml-2">{resume.directionLabel}</span>
                        )}
                      </div>
                      <ChevronRight className="w-4 h-4 text-slate-300" />
                    </div>
                  ))
                )}
              </div>
            ) : (
              /* ── Step 2: Select Job ── */
              <JobPicker
                selectedJobId={selectedJob?.id ?? null}
                onSelect={setSelectedJob}
              />
            )}
          </div>

          {/* ── Footer ── */}
          <footer className="shrink-0 border-t border-slate-100 bg-white px-6 py-4">
            <div className="flex items-center justify-between">
              <button
                onClick={handleGeneralOptimization}
                disabled={!selectedResume}
                className={`text-[12px] transition-colors ${
                  selectedResume
                    ? "text-slate-400 hover:text-[#0052D9] cursor-pointer"
                    : "text-slate-300 cursor-not-allowed"
                }`}
              >
                继续通用优化
              </button>

              <div className="flex items-center gap-3">
                <button
                  onClick={onClose}
                  className="px-4 py-2 text-[13px] text-slate-500 hover:text-slate-800 transition-colors rounded-lg hover:bg-slate-100"
                >
                  取消
                </button>
                <button
                  onClick={handleConfirm}
                  disabled={!canConfirm}
                  className={`px-5 py-2 text-[13px] font-medium rounded-lg transition-colors flex items-center gap-1.5 ${
                    canConfirm
                      ? "bg-[#0052D9] text-white hover:bg-[#003DA5] cursor-pointer"
                      : "bg-slate-200 text-slate-400 cursor-not-allowed"
                  }`}
                >
                  开始针对岗位优化
                  <ArrowRight className="w-3.5 h-3.5" />
                </button>
              </div>
            </div>
          </footer>
        </section>
      </div>
    </div>,
    document.body,
  );
}
