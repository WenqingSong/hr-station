import React, { useEffect, useState } from "react";
import { Link } from "react-router";
import {
  X,
  CheckCircle2,
  Sparkles,
  AlertTriangle,
  Clock,
  Briefcase,
  ChevronRight,
} from "lucide-react";
import { loadResumeOptionViewModelsFromApi } from "../../api";
import type { ResumeOptionViewModel } from "../../adapters";

// ─── Types ────────────────────────────────────────────────────────────────────

type Scene = "recommendation" | "match-analysis";

interface SelectResumeModalProps {
  isOpen: boolean;
  onClose: () => void;
  onConfirm?: (resumeId: string) => void;
  scene?: Scene;
  jobTitle?: string;
}

const SCENE_CONFIG: Record<Scene, { label: string; desc: string }> = {
  "recommendation": {
    label: "岗位推荐",
    desc: "系统将从当前岗位池中推荐最值得优先关注的 5 个岗位。",
  },
  "match-analysis": {
    label: "岗位匹配分析",
    desc: "系统将对比岗位 JD 与所选简历，生成匹配度、证据链和投递建议。",
  },
};

// ─── Component ────────────────────────────────────────────────────────────────

export function SelectResumeModal({
  isOpen,
  onClose,
  onConfirm,
  scene = "match-analysis",
  jobTitle = "软件开发-后台开发方向",
}: SelectResumeModalProps) {
  const [selected, setSelected] = useState("1");
  const [resumeOptions, setResumeOptions] = useState<ResumeOptionViewModel[]>([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (isOpen) {
      setLoading(true);
      loadResumeOptionViewModelsFromApi()
        .then((opts) => {
          setResumeOptions(opts);
          // 默认选中最新版本（第一个 = 最新）
          if (opts.length > 0) setSelected(opts[0].id);
        })
        .catch(() => {})
        .finally(() => setLoading(false));
    }
  }, [isOpen]);

  if (!isOpen) return null;

  const sceneInfo = SCENE_CONFIG[scene];

  function handleConfirm() {
    onConfirm?.(selected);
    onClose();
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      {/* Overlay */}
      <div
        className="absolute inset-0 bg-black/35"
        onClick={onClose}
      />

      {/* Modal — 700px wide */}
      <div className="relative bg-white w-full max-w-[700px] mx-4 z-10 shadow-xl flex flex-col max-h-[90vh]">

        {/* ── Header ──────────────────────────────────────────────────── */}
        <div className="flex items-start justify-between px-7 pt-6 pb-5 border-b border-gray-100 flex-shrink-0">
          <div>
            <h2 className="font-semibold text-gray-900 text-base">选择用于分析的简历</h2>
            <p className="text-sm text-gray-500 mt-1">
              系统将基于所选简历进行{sceneInfo.label}。
            </p>
          </div>
          <button
            onClick={onClose}
            className="p-1 text-gray-300 hover:text-gray-500 transition-colors flex-shrink-0 ml-4 mt-0.5"
          >
            <X className="w-4 h-4" />
          </button>
        </div>

        {/* ── Scene context strip ─────────────────────────────────────── */}
        <div className="mx-7 mt-5 px-4 py-3 bg-gray-50 border border-gray-200 flex-shrink-0">
          <div className="flex items-center gap-4 mb-1.5">
            <span className="text-xs text-gray-500">
              当前操作：<span className="font-medium text-gray-800">{sceneInfo.label}</span>
            </span>
            {scene === "match-analysis" && jobTitle && (
              <>
                <span className="w-px h-3 bg-gray-300" />
                <span className="text-xs text-gray-500">
                  目标岗位：<span className="text-gray-700">{jobTitle}</span>
                </span>
              </>
            )}
          </div>
          <p className="text-xs text-gray-400 leading-relaxed">{sceneInfo.desc}</p>
        </div>

        {/* ── Resume cards ────────────────────────────────────────────── */}
        <div className="px-7 py-5 space-y-3 overflow-y-auto flex-1">
          {loading ? (
            <p className="text-sm text-gray-400 py-8 text-center">正在加载简历列表…</p>
          ) : resumeOptions.length === 0 ? (
            <p className="text-sm text-gray-400 py-8 text-center">暂无可选择的简历版本</p>
          ) : (
          resumeOptions.map((resume) => {
            const isMismatch = !!resume.mismatchReason;
            const isSelected = !isMismatch && selected === resume.id;

            return (
              <div
                key={resume.id}
                onClick={() => {
                  if (!isMismatch) setSelected(resume.id);
                }}
                className={`flex items-start gap-4 px-5 py-4 border cursor-pointer transition-all ${
                  isSelected
                    ? "border-[#0052D9] bg-[#EEF4FF]"
                    : isMismatch
                    ? "border-orange-100 bg-orange-50/30 hover:border-orange-200"
                    : "border-gray-200 bg-white hover:border-gray-300"
                }`}
              >
                {/* Radio dot */}
                <div className="flex-shrink-0 mt-0.5">
                  <div
                    className={`w-4 h-4 rounded-full border-2 flex items-center justify-center transition-colors ${
                      isSelected ? "border-[#0052D9]" : "border-gray-300"
                    }`}
                  >
                    {isSelected && (
                      <div className="w-2 h-2 rounded-full bg-[#0052D9]" />
                    )}
                  </div>
                </div>

                {/* Content */}
                <div className="flex-1 min-w-0">
                  {/* Name + badges */}
                  <div className="flex items-center gap-2 mb-2 flex-wrap">
                    <span
                      className={`text-sm font-medium ${
                        isMismatch ? "text-orange-700" : "text-gray-900"
                      }`}
                    >
                      {resume.name}
                    </span>

                    {resume.isRecommended && (
                      <span className="inline-flex items-center gap-1 px-2 py-0.5 bg-blue-100 text-[#0052D9] text-xs rounded">
                        <Sparkles className="w-3 h-3" />
                        推荐使用
                      </span>
                    )}
                    {resume.isDefault && (
                      <span className="inline-flex items-center gap-1 px-2 py-0.5 bg-blue-50 border border-blue-100 text-[#0052D9] text-xs rounded">
                        <CheckCircle2 className="w-3 h-3" />
                        当前使用
                      </span>
                    )}
                    {resume.isRecommended && (
                      <span className="px-2 py-0.5 bg-green-50 border border-green-100 text-green-700 text-xs rounded">
                        可用于 AI 分析
                      </span>
                    )}
                    {resume.statusLabel === '历史版本' && (
                      <span className="px-2 py-0.5 bg-gray-100 text-gray-500 text-xs rounded">
                        {resume.statusLabel}
                      </span>
                    )}
                    {isMismatch && (
                      <span className="inline-flex items-center gap-1 px-2 py-0.5 bg-orange-50 border border-orange-100 text-orange-600 text-xs rounded">
                        <AlertTriangle className="w-3 h-3" />
                        方向不匹配
                      </span>
                    )}
                  </div>

                  {/* Meta row */}
                  <div className="flex items-center gap-3 text-xs text-gray-500 mb-2">
                    <span className="flex items-center gap-1">
                      <Briefcase className="w-3 h-3" />
                      {resume.directionLabel}
                    </span>
                    <span className="w-px h-3 bg-gray-200" />
                    <span>{resume.tags.join(" / ")}</span>
                  </div>

                  <div className="flex items-center gap-3 text-xs text-gray-400 mb-2.5">
                    <span>{resume.projectCount} 个项目经历</span>
                    <span className="w-px h-3 bg-gray-200" />
                    <span className="flex items-center gap-1">
                      <Clock className="w-3 h-3" />
                      最近更新：{resume.updatedAt}
                    </span>
                  </div>

                  {/* Fit note */}
                  <p
                    className={`text-xs leading-relaxed ${
                      isMismatch ? "text-orange-500" : "text-gray-500"
                    }`}
                  >
                    {resume.recommendationLabel || resume.mismatchReason || resume.directionLabel}
                  </p>
                </div>
              </div>
            );
          })
          )}

          {/* Selection guidance */}
          <div className="pt-1 space-y-1.5">
            <p className="text-xs text-gray-500 leading-relaxed">
              建议选择与目标岗位方向最接近、项目经历最完整的简历版本。
            </p>
            <p className="text-xs text-gray-400 leading-relaxed">
              AI 分析仅基于所选简历内容，不会替你编造经历。
            </p>
          </div>
        </div>

        {/* ── Footer ──────────────────────────────────────────────────── */}
        <div className="flex items-center justify-between px-7 py-4 border-t border-gray-100 flex-shrink-0">
          {/* Weak management link */}
          <Link
            to="/resumes"
            onClick={onClose}
            className="flex items-center gap-1 text-xs text-gray-400 hover:text-[#0052D9] transition-colors"
          >
            管理我的简历
            <ChevronRight className="w-3 h-3" />
          </Link>

          {/* Action buttons */}
          <div className="flex items-center gap-3">
            <button
              onClick={onClose}
              className="px-4 py-2 text-sm text-gray-500 hover:text-gray-800 transition-colors"
            >
              取消
            </button>
            <button
              onClick={handleConfirm}
              className="px-5 py-2 bg-[#0052D9] text-white text-sm font-medium rounded hover:bg-[#003DA5] transition-colors"
            >
              确认并开始分析
            </button>
          </div>
        </div>

      </div>
    </div>
  );
}
