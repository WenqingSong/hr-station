import React from "react";
import { Sparkles, FileText, Target, ArrowRight, ShieldAlert } from "lucide-react";
import { Button } from "./ui/Button";
import { Badge } from "./ui/Badge";
import { useNavigate } from "react-router";
import { mockCurrentResume, mockSelectedJob } from "../mocks";
import { toResumeSummaryViewModel, toJobDetailViewModel } from "../adapters";

const resumeVm = toResumeSummaryViewModel(mockCurrentResume);
const jobVm = toJobDetailViewModel(mockSelectedJob);

export function JobDetailAIDecision() {
  const navigate = useNavigate();

  return (
    <aside className="w-[340px] shrink-0 sticky top-24 self-start flex flex-col gap-4">
      {/* Header */}
      <div className="mb-2">
        <div className="flex items-center gap-2 mb-1">
          <Sparkles className="h-5 w-5 text-[#0052D9]" />
          <h2 className="text-xl font-bold text-gray-900">AI 投递决策</h2>
        </div>
        <p className="text-sm text-gray-500">
          分析当前岗位与简历的匹配度和投递风险。
        </p>
      </div>

      {/* Currently Selected Job */}
      <div className="bg-white border border-gray-200 rounded-lg p-4">
        <div className="flex items-center gap-2 text-gray-500 text-xs mb-3">
          <Target className="h-3.5 w-3.5" />
          当前岗位
        </div>
        <h3 className="font-bold text-gray-900 text-sm mb-2">{jobVm.title}</h3>
        <div className="flex flex-wrap gap-1.5 mb-2.5">
          <Badge variant="secondary" className="text-[10px] px-1.5 font-normal">{jobVm.typeLabel}</Badge>
          <Badge variant="secondary" className="text-[10px] px-1.5 font-normal">{jobVm.categoryLabel}</Badge>
          <Badge variant="secondary" className="text-[10px] px-1.5 font-normal">后台开发</Badge>
        </div>
        <p className="text-[11px] text-gray-400">
          工作地点：{jobVm.locationsText}
        </p>
      </div>

      {/* Current Resume Card */}
      <div className="bg-white border border-gray-200 rounded-lg p-4">
        <div className="flex justify-between items-start mb-3">
          <div className="flex items-center gap-2 text-gray-500 text-xs">
            <FileText className="h-3.5 w-3.5" />
            当前简历
          </div>
          <button className="text-[11px] text-[#0052D9] hover:underline">
            切换简历
          </button>
        </div>
        <div className="mb-3">
          <h3 className="font-bold text-gray-900 text-sm mb-2">{resumeVm.name}</h3>
          <div className="flex flex-wrap gap-1.5">
            <Badge variant="outline" className="text-[10px] px-1.5 text-gray-600 bg-gray-50 font-normal">{resumeVm.directionLabel}</Badge>
            {resumeVm.skillTags.slice(0, 4).map((skill) => (
              <Badge key={skill} variant="outline" className="text-[10px] px-1.5 text-gray-600 bg-gray-50 font-normal">{skill}</Badge>
            ))}
          </div>
        </div>
        <p className="text-[11px] text-gray-400">
          包含 {resumeVm.projectCount} 个项目经历，最近更新：{resumeVm.updatedAt}
        </p>
      </div>

      {/* Main Action Card */}
      <div className="bg-gradient-to-b from-[#F4F8FF] to-white border border-[#0052D9] rounded-lg p-5 text-center mt-2 relative overflow-hidden shadow-sm">
        <div className="absolute top-0 left-0 w-full h-1 bg-[#0052D9]"></div>
        <h3 className="font-bold text-gray-900 mb-2 mt-1 text-[15px]">想知道这个岗位适不适合你？</h3>
        <p className="text-xs text-gray-600 mb-5 leading-relaxed">
          系统将基于当前岗位 JD 和当前简历，分析匹配度、证据链、风险短板和投递前优化优先级。
        </p>
        <Button 
          className="w-full shadow-sm group bg-[#0052D9] hover:bg-[#0040A8] text-white"
          onClick={() => navigate("/match")}
        >
          开始岗位匹配分析
          <ArrowRight className="ml-2 h-4 w-4 group-hover:translate-x-1 transition-transform" />
        </Button>
      </div>

      {/* Pre-delivery Tip Card */}
      <div className="bg-orange-50/50 border border-orange-100 rounded-lg p-4 mt-2">
        <div className="flex items-center gap-2 text-orange-800 font-medium mb-3 text-xs">
          <ShieldAlert className="h-3.5 w-3.5" />
          分析前建议确认
        </div>
        <ul className="text-[11px] text-orange-700/80 space-y-2 list-disc pl-4 marker:text-orange-300">
          <li>当前简历是否为目标方向版本</li>
          <li>项目经历是否完整</li>
          <li>技能关键词是否覆盖后端方向</li>
        </ul>
      </div>

      {/* Status Badge */}
      <div className="mt-2 flex justify-center">
        <div className="inline-flex items-center gap-1.5 px-2 py-1 rounded-full text-gray-400 text-[10px] opacity-80">
          <Sparkles className="h-3 w-3" />
          当前结果来源：Mock 演示
        </div>
      </div>
    </aside>
  );
}
