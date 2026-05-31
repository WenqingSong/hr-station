import React from "react";
import { Plus } from "lucide-react";
import { Button } from "./ui/Button";

interface MyResumesHeaderProps {
  onAddResume: () => void;
}

export function MyResumesHeader({ onAddResume }: MyResumesHeaderProps) {
  return (
    <div className="bg-white rounded-lg p-5 border border-gray-200 shadow-sm flex items-center justify-between mb-4">
      <div className="flex flex-col gap-1">
        <h1 className="text-[20px] font-bold text-gray-900 tracking-tight">我的简历</h1>
        <p className="text-[13px] text-gray-500">
          管理用于岗位推荐、岗位匹配分析和简历优化的简历版本。
        </p>
      </div>
      <div className="flex items-center gap-4">
        <div className="text-[12px] text-gray-400">
          当前系统会优先使用“当前使用”简历参与 AI 分析。
        </div>
        <Button 
          onClick={onAddResume}
          className="bg-[#0052D9] hover:bg-[#0040A8] text-white h-9 px-4 text-[14px] flex items-center gap-1.5 shadow-sm"
        >
          <Plus className="w-4 h-4" />
          添加简历
        </Button>
      </div>
    </div>
  );
}