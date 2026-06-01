import React, { useState, useEffect, useCallback } from "react";
import { useSearchParams } from "react-router";
import { RecommendationHeader } from "../components/RecommendationHeader";
import { RecommendationSidebar } from "../components/RecommendationSidebar";
import { RecommendationDetail } from "../components/RecommendationDetail";
import { RecommendationContext } from "../components/RecommendationContext";
import { loadRecommendationViewModel, loadResumeOptionViewModelsFromApi } from "../../api";
import type { RecommendationPageViewModel, ResumeOptionViewModel } from "../../adapters";

/** 从 URL 查询参数中解析 resume_version_no（兼容 legacy resumeId=resume-vN） */
function parseVersionNoFromQuery(resumeId: string | null, resumeVersionNo: string | null): number | null {
  // 优先使用 resumeVersionNo
  if (resumeVersionNo && /^\d+$/.test(resumeVersionNo)) return Number(resumeVersionNo);
  // 兼容 legacy resumeId
  if (resumeId) {
    if (/^\d+$/.test(resumeId)) return Number(resumeId);
    const match = resumeId.match(/^resume-v(\d+)$/);
    if (match) return Number(match[1]);
  }
  return null;
}

export function JobRecommendationPage() {
  const [searchParams] = useSearchParams();
  const resumeId = searchParams.get("resumeId");
  const resumeVersionNo = searchParams.get("resumeVersionNo");

  // 简历选项 & 选中状态
  const [resumeOptions, setResumeOptions] = useState<ResumeOptionViewModel[]>([]);
  const [selectedOption, setSelectedOption] = useState<ResumeOptionViewModel | null>(null);
  const [loadingOptions, setLoadingOptions] = useState(true);

  // 推荐结果
  const [result, setResult] = useState<RecommendationPageViewModel | null>(null);
  const [loadingRecommendation, setLoadingRecommendation] = useState(false);
  const [recommendationError, setRecommendationError] = useState<string | null>(null);

  // 加载真实简历选项列表
  useEffect(() => {
    loadResumeOptionViewModelsFromApi()
      .then((opts) => {
        setResumeOptions(opts);
        // 无简历版本 → 显示空态
        if (opts.length === 0) {
          setSelectedOption(null);
          setLoadingOptions(false);
          return;
        }
        // 有 URL 参数时尝试匹配，否则默认选最新（第一项）
        const targetVersionNo = parseVersionNoFromQuery(resumeId, resumeVersionNo);
        const matched = targetVersionNo
          ? opts.find((o) => Number(o.id) === targetVersionNo)
          : undefined;
        setSelectedOption(matched ?? opts[0]);
        setLoadingOptions(false);
      })
      .catch(() => setLoadingOptions(false));
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  // 选中版本后在 URL 已由父级提供，此处仅触发推荐请求
  useEffect(() => {
    if (!selectedOption) return;
    const versionNo = Number(selectedOption.id);
    if (!versionNo || isNaN(versionNo)) return;

    setLoadingRecommendation(true);
    setRecommendationError(null);
    loadRecommendationViewModel({ resume_version_no: versionNo, limit: 5 })
      .then(setResult)
      .catch(() => setRecommendationError("推荐结果加载失败，请稍后重试。"))
      .finally(() => setLoadingRecommendation(false));
  }, [selectedOption]);

  const handleConfirmResume = useCallback((resumeVersionId: string) => {
    const opt = resumeOptions.find((o) => o.id === resumeVersionId);
    if (opt) setSelectedOption(opt);
  }, [resumeOptions]);

  // ── 空态：正在加载选项 ──
  if (loadingOptions) {
    return (
      <div className="max-w-[1440px] mx-auto px-6 py-12 text-center text-gray-400 text-sm">
        正在加载简历信息…
      </div>
    );
  }

  // ── 空态：无简历版本 ──
  if (resumeOptions.length === 0 || !selectedOption) {
    return (
      <div className="max-w-[1440px] mx-auto px-6 py-12 text-center">
        <p className="text-gray-500 text-sm mb-3">你还没有可用于推荐的简历</p>
        <p className="text-gray-400 text-xs mb-4">请先创建一份简历，系统会基于你的简历版本生成岗位推荐。</p>
        <a href="/resumes/new" className="text-[#0052D9] text-sm underline">去添加简历</a>
      </div>
    );
  }

  // ── 空态：推荐加载中 ──
  if (loadingRecommendation) {
    return (
      <div className="max-w-[1440px] mx-auto px-6 py-12 text-center text-gray-400 text-sm">
        正在生成岗位推荐…
      </div>
    );
  }

  // ── 失败态 ──
  if (!result) {
    return (
      <div className="max-w-[1440px] mx-auto px-6 py-12 text-center">
        <p className="text-gray-500 text-sm mb-2">{recommendationError || "推荐接口暂时不可用，请稍后重试。"}</p>
      </div>
    );
  }

  const displayName = `${selectedOption.name}（V${selectedOption.id}）`;
  const statusLabel = selectedOption.directionLabel
    ? `${selectedOption.statusLabel} · ${selectedOption.directionLabel}`
    : selectedOption.statusLabel;

  return (
    <div className="mx-auto max-w-[1440px] px-6 md:px-8 py-6 pb-12">
      <RecommendationHeader resumeName={displayName} sourceLabel={statusLabel} />
      <div className="flex gap-6 mt-6 items-start">
        <RecommendationSidebar />
        <RecommendationDetail currentResumeId={selectedOption.id} />
        <RecommendationContext resumeName={selectedOption.name} sourceLabel={statusLabel} />
      </div>
    </div>
  );
}
