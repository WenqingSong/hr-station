import React, { useState, useEffect } from "react";
import { useSearchParams } from "react-router";
import { RecommendationHeader } from "../components/RecommendationHeader";
import { RecommendationSidebar } from "../components/RecommendationSidebar";
import { RecommendationDetail } from "../components/RecommendationDetail";
import { RecommendationContext } from "../components/RecommendationContext";
import { loadRecommendationViewModel } from "../../api";
import type { RecommendationPageViewModel } from "../../adapters";
import { mockCurrentResume, mockResumeVersions } from "../../mocks";

/** 将前端 resumeId（如 resume-v2）映射为后端 resume_version_no */
function resolveResumeVersionNo(resumeId: string | null): number {
  if (!resumeId) return 2;
  const matched = resumeId.match(/\d+/);
  return matched ? Number(matched[0]) : 2;
}

export function JobRecommendationPage() {
  const [searchParams] = useSearchParams();
  const resumeId = searchParams.get("resumeId");
  const [result, setResult] = useState<RecommendationPageViewModel | null>(null);
  const [loading, setLoading] = useState(true);

  // 解析 resumeId 用于上下文展示（仍通过 mockResumeVersions 查找显示名）
  const matchedResume = resumeId
    ? mockResumeVersions.find((r) => r.id === resumeId) ?? mockCurrentResume
    : mockCurrentResume;
  const sourceLabel = "当前简历";

  useEffect(() => {
    const versionNo = resolveResumeVersionNo(resumeId);
    loadRecommendationViewModel({ resume_version_no: versionNo, limit: 5 })
      .then(setResult)
      .catch(() => {})
      .finally(() => setLoading(false));
  }, [resumeId]);

  if (loading) {
    return (
      <div className="max-w-[1440px] mx-auto px-6 py-12 text-center text-gray-400 text-sm">
        正在生成岗位推荐…
      </div>
    );
  }

  if (!result) {
    return (
      <div className="max-w-[1440px] mx-auto px-6 py-12 text-center">
        <p className="text-gray-500 text-sm mb-2">推荐接口暂时不可用，已展示演示数据。</p>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-[1440px] px-6 md:px-8 py-6 pb-12">
      <RecommendationHeader resumeName={matchedResume.name} sourceLabel={sourceLabel} />
      <div className="flex gap-6 mt-6 items-start">
        <RecommendationSidebar />
        <RecommendationDetail currentResumeId={matchedResume.id} />
        <RecommendationContext resumeName={matchedResume.name} sourceLabel={sourceLabel} />
      </div>
    </div>
  );
}
