import React from "react";
import { RecommendationHeader } from "../components/RecommendationHeader";
import { RecommendationSidebar } from "../components/RecommendationSidebar";
import { RecommendationDetail } from "../components/RecommendationDetail";
import { RecommendationContext } from "../components/RecommendationContext";

export function JobRecommendationPage() {
  return (
    <div className="min-h-screen bg-[#F3F4F6] font-sans text-gray-900 selection:bg-blue-100 selection:text-blue-900 pb-20">
      <div className="mx-auto max-w-[1440px] px-6 md:px-8 py-6">
        <RecommendationHeader />
        
        <div className="flex gap-6 mt-6 items-start justify-center">
          <RecommendationSidebar />
          <RecommendationDetail />
          <RecommendationContext />
        </div>
      </div>
    </div>
  );
}
