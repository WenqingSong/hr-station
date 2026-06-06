import React, { useState, useEffect, useCallback } from "react";
import { Link, useSearchParams, useNavigate, useLocation } from "react-router";
import { RecommendationHeader } from "../components/recommendation/RecommendationHeader";
import { RecommendationSidebar } from "../components/recommendation/RecommendationSidebar";
import { RecommendationDetail } from "../components/recommendation/RecommendationDetail";
import { RecommendationContext } from "../components/recommendation/RecommendationContext";
import { loadRecommendationViewModel, loadResumeOptionViewModelsFromApi } from "../../api";
import type { RecommendationPageViewModel, ResumeOptionViewModel } from "../../adapters";
import { PageLoading, PageEmpty } from "../components/PageState";
import { ContentPageShell } from "../components/layout/ContentPageShell";
import { getAuthRedirectIf401 } from "../utils/authRedirect";

/** 从 URL 查询参数中解析 resumeId（K8-5-48-C） */
function parseResumeIdFromQuery(resumeId: string | null): number | null {
  if (resumeId && /^\d+$/.test(resumeId)) return Number(resumeId);
  return null;
}

export function JobRecommendationPage() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const location = useLocation();
  const resumeId = searchParams.get("resumeId");

  const [resumeOptions, setResumeOptions] = useState<ResumeOptionViewModel[]>([]);
  const [selectedOption, setSelectedOption] = useState<ResumeOptionViewModel | null>(null);
  const [loadingOptions, setLoadingOptions] = useState(true);

  const [result, setResult] = useState<RecommendationPageViewModel | null>(null);
  const [loadingRecommendation, setLoadingRecommendation] = useState(false);
  const [recommendationError, setRecommendationError] = useState<string | null>(null);

  useEffect(() => {
    loadResumeOptionViewModelsFromApi()
      .then((opts) => {
        setResumeOptions(opts);
        if (opts.length === 0) {
          setSelectedOption(null);
          setLoadingOptions(false);
          return;
        }
        const targetResumeId = parseResumeIdFromQuery(resumeId);
        const matched = targetResumeId
          ? opts.find((o) => o.id === targetResumeId)
          : undefined;
        setSelectedOption(matched ?? opts[0]);
        setLoadingOptions(false);
      })
      .catch((err) => {
        const authRedirect = getAuthRedirectIf401(err, location.pathname, location.search);
        if (authRedirect) { navigate(authRedirect, { replace: true }); return; }
        setLoadingOptions(false);
      });
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  useEffect(() => {
    if (!selectedOption) return;
    const id = selectedOption.id;
    if (!id || isNaN(id)) return;

    setLoadingRecommendation(true);
    setRecommendationError(null);
    loadRecommendationViewModel({ /* resume_id sent if backend expects it */ })
      .then(setResult)
      .catch(() => setRecommendationError("推荐结果加载失败，请稍后重试。"))
      .finally(() => setLoadingRecommendation(false));
  }, [selectedOption]);

  const handleConfirmResume = useCallback((resumeVersionId: string) => {
    const opt = resumeOptions.find((o) => String(o.id) === resumeVersionId);
    if (opt) setSelectedOption(opt);
  }, [resumeOptions]);

  if (loadingOptions) {
    return <PageLoading message="正在加载简历信息…" />;
  }

  if (resumeOptions.length === 0 || !selectedOption) {
    return (
      <PageEmpty
        title="你还没有可用于推荐的简历"
        description="请先创建一份简历，系统会基于你的简历版本生成岗位推荐。"
      >
        <Link to="/resumes/new" className="text-[#0052D9] text-sm underline">去添加简历</Link>
      </PageEmpty>
    );
  }

  if (loadingRecommendation) {
    return <PageLoading message="正在生成岗位推荐…" />;
  }

  if (!result) {
    return (
      <div className="max-w-[1440px] mx-auto px-6 py-12 text-center">
        <p className="text-gray-500 text-sm mb-2">{recommendationError || "推荐接口暂时不可用，请稍后重试。"}</p>
        <Link to="/resumes" className="text-[#0052D9] text-sm underline mt-3 inline-block">返回我的简历</Link>
      </div>
    );
  }

  const displayName = `${selectedOption.name}（ID: ${selectedOption.id}）`;

  return (
    <ContentPageShell maxWidthClassName="max-w-[1440px]" contentClassName="pb-12">
      <RecommendationHeader resumeName={displayName} sourceLabel={""} />
      <div className="flex gap-6 mt-6 items-start">
        <RecommendationSidebar data={result} />
        <RecommendationDetail currentResumeId={String(selectedOption.id)} data={result} />
        <RecommendationContext resumeName={selectedOption.name} sourceLabel={""} />
      </div>
    </ContentPageShell>
  );
}
