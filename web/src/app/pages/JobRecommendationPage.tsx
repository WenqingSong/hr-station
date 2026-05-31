import React from "react";
import { RecommendationHeader } from "../components/RecommendationHeader";
import { RecommendationSidebar } from "../components/RecommendationSidebar";
import { RecommendationDetail } from "../components/RecommendationDetail";
import { RecommendationContext } from "../components/RecommendationContext";

export function JobRecommendationPage() {
  return (
    <div className="mx-auto max-w-[1440px] px-6 md:px-8 py-6 pb-12">
      <RecommendationHeader />
      <div className="flex gap-6 mt-6 items-start">
        <RecommendationSidebar />
        <RecommendationDetail />
        <RecommendationContext />
      </div>
    </div>
  );
}
