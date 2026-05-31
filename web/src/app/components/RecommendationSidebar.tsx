import React from "react";
import { MapPin, AlertCircle, ChevronRight } from "lucide-react";
import { Badge } from "./ui/Badge";
import { cn } from "../utils";
import { mockRecommendationResponse } from "../mocks";
import { toRecommendationPageViewModel } from "../adapters";

const recVm = toRecommendationPageViewModel(
  mockRecommendationResponse.result,
  mockRecommendationResponse.meta
);

export function RecommendationSidebar() {
  const selectedId = "job_backend_001";

  const getLevelStyle = (level: string) => {
    switch (level) {
      case "强推荐":
        return "bg-green-50 text-green-700 border-green-200";
      case "推荐":
        return "bg-blue-50 text-blue-700 border-blue-200";
      case "谨慎推荐":
        return "bg-amber-50 text-amber-700 border-amber-200";
      default:
        return "bg-gray-50 text-gray-700";
    }
  };

  return (
    <div className="w-[380px] shrink-0 flex flex-col gap-4">
      {recVm.recommendedJobs.map((job, index) => {
        const isSelected = job.jobId === selectedId;
        
        return (
          <div
            key={job.jobId}
            className={cn(
              "group relative p-4 bg-white border rounded-lg cursor-pointer transition-all duration-200 hover:shadow-md",
              isSelected
                ? "border-[#0052D9] bg-[#F4F8FF] shadow-sm"
                : "border-gray-200 hover:border-[#0052D9]"
            )}
          >
            {isSelected && (
              <div className="absolute left-0 top-0 bottom-0 w-1 bg-[#0052D9] rounded-l-lg"></div>
            )}
            
            <div className="flex items-start justify-between mb-2">
              <div className="flex items-center gap-2">
                <span className={cn(
                  "text-[14px] font-black italic",
                  isSelected ? "text-[#0052D9]" : "text-gray-400"
                )}>#{index + 1}</span>
                <h3 className={cn(
                  "font-bold text-[16px]",
                  isSelected ? "text-[#0052D9]" : "text-gray-900"
                )}>{job.jobTitle}</h3>
              </div>
            </div>

            <div className="flex flex-wrap gap-2 items-center mb-3">
              <span className={cn("text-[11px] px-2 py-0.5 rounded-sm border", getLevelStyle(job.recommendationLevel))}>
                {job.recommendationLevel}
              </span>
              <div className="flex items-center text-[12px] text-gray-500">
                <MapPin className="w-3 h-3 mr-1" />
                {job.location}
              </div>
            </div>

            <div className="flex flex-wrap gap-1.5 mb-3">
              {job.matchKeywords.map((kw, i) => (
                <Badge key={i} variant="secondary" className="text-[10px] font-normal px-1.5 py-0 bg-white border border-gray-100 text-gray-600">
                  {kw}
                </Badge>
              ))}
            </div>

            <div className="text-[12px] text-gray-600 bg-gray-50 p-2 rounded mb-2 leading-relaxed">
              {job.reason}
            </div>

            <div className="flex items-center justify-between mt-3 pt-3 border-t border-gray-100/50">
              <div className="flex items-center text-[11px] text-amber-600 bg-amber-50/50 px-1.5 py-0.5 rounded">
                <AlertCircle className="w-3 h-3 mr-1 text-amber-500" />
                {job.riskTags[0] || job.reason}
              </div>
              <span className={cn(
                "text-[12px] flex items-center font-medium transition-colors",
                isSelected ? "text-[#0052D9]" : "text-[#0052D9] opacity-80 group-hover:opacity-100"
              )}>
                查看推荐分析 <ChevronRight className="w-3.5 h-3.5 ml-0.5" />
              </span>
            </div>
          </div>
        );
      })}
    </div>
  );
}
