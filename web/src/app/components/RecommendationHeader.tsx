import React from "react";
import { ArrowLeft } from "lucide-react";
import { useNavigate } from "react-router";

interface RecommendationHeaderProps {
  resumeName?: string;
  sourceLabel?: string;
}

export function RecommendationHeader({ resumeName = "V2 后端开发简历", sourceLabel = "当前简历" }: RecommendationHeaderProps) {
  const navigate = useNavigate();

  return (
    <div className="bg-white border border-gray-200 rounded-lg p-6 relative shadow-sm">
      <button 
        onClick={() => navigate("/")}
        className="flex items-center gap-1.5 text-gray-500 hover:text-[#0052D9] transition-colors mb-4 text-[13px] font-medium"
      >
        <ArrowLeft className="w-3.5 h-3.5" />
        返回岗位列表
      </button>

      <div className="flex flex-col gap-2">
        <h1 className="text-2xl font-bold text-gray-900 tracking-tight">AI 岗位推荐</h1>
        <p className="text-[15px] text-gray-600">基于当前简历，从岗位池中推荐最值得优先关注的 5 个岗位。</p>
      </div>

      <div className="mt-5 flex flex-wrap items-center gap-6 pt-4 border-t border-gray-100 text-sm">
        <div className="flex items-center gap-2">
          <span className="text-gray-400">当前简历：</span>
          <span className="font-medium text-gray-800">{resumeName}</span>
        </div>
        <div className="w-px h-3.5 bg-gray-200"></div>
        <div className="flex items-center gap-2">
          <span className="text-gray-400">推荐来源：</span>
          <span className="text-[#0052D9] font-medium bg-blue-50 px-2 py-0.5 rounded-sm">{sourceLabel}</span>
        </div>
        <div className="w-px h-3.5 bg-gray-200"></div>
        <div className="flex items-center gap-2">
          <span className="text-gray-400">推荐结果：</span>
          <span className="text-[#0052D9] font-medium bg-blue-50 px-2 py-0.5 rounded-sm">Top 5</span>
        </div>
      </div>
    </div>
  );
}
