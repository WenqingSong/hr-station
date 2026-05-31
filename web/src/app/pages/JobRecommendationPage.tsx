import React from "react";
import { useSearchParams } from "react-router";
import { RecommendationHeader } from "../components/RecommendationHeader";
import { RecommendationSidebar } from "../components/RecommendationSidebar";
import { RecommendationDetail } from "../components/RecommendationDetail";
import { RecommendationContext } from "../components/RecommendationContext";
import { mockCurrentResume, mockResumeVersions } from "../../mocks";

export function JobRecommendationPage() {
  const [searchParams] = useSearchParams();
  const resumeId = searchParams.get("resumeId");
  const matchedResume = resumeId
    ? mockResumeVersions.find((r) => r.id === resumeId) ?? mockCurrentResume
    : mockCurrentResume;
  const sourceLabel = resumeId && mockResumeVersions.some((r) => r.id === resumeId)
    ? "已选择简历"
    : "默认简历";

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
