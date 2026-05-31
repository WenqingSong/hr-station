import React from "react";
import { CheckCircle2, ShieldAlert, Zap, ArrowRight } from "lucide-react";
import { Button } from "./ui/Button";
import { Badge } from "./ui/Badge";
import { useNavigate } from "react-router";
import { mockRecommendationResponse } from "../mocks";
import { toRecommendationPageViewModel } from "../adapters";

const recVm = toRecommendationPageViewModel(
  mockRecommendationResponse.result,
  mockRecommendationResponse.meta
);
const firstJob = recVm.recommendedJobs[0];

export function RecommendationDetail() {
  const navigate = useNavigate();
  const getStrengthStyle = (tone: string) => {
    switch (tone) {
      case 'green': return "bg-green-50 text-green-700 border border-green-200";
      case 'blue': return "bg-blue-50 text-[#0052D9] border border-blue-200";
      case 'orange': return "bg-amber-50 text-amber-700 border border-amber-200";
      default: return "bg-gray-50 text-gray-700";
    }
  };

  return (
    <div className="flex-1 min-w-[560px] max-w-[680px] bg-white border border-gray-200 rounded-lg shadow-sm overflow-hidden">
      {/* Top Info Card */}
      <div className="bg-[#F4F8FF] p-6 border-b border-blue-100">
        <div className="flex items-center gap-2 text-gray-500 text-[13px] mb-2">
          当前选中岗位
        </div>
        <h2 className="text-xl font-bold text-gray-900 mb-4">{firstJob.jobTitle}</h2>
        
        <div className="bg-white rounded-md p-4 border border-blue-50 shadow-sm flex items-start gap-4">
          <div className="shrink-0 flex flex-col items-center justify-center bg-green-50 text-green-700 border border-green-100 rounded p-2 min-w-[80px]">
            <CheckCircle2 className="w-5 h-5 mb-1" />
            <span className="text-[13px] font-bold">{firstJob.recommendationLevel}</span>
          </div>
          <p className="text-[14px] text-gray-700 leading-relaxed pt-0.5">
            <span className="font-semibold text-gray-900">一句话总结：</span>
            {firstJob.reason}
          </p>
        </div>
      </div>

      <div className="p-6 space-y-8">
        {/* Module 1: Evidence */}
        <section>
          <h3 className="font-bold text-gray-900 mb-4 text-[16px] flex items-center gap-2">
            <span className="w-1 h-4 bg-[#0052D9] rounded-full inline-block"></span>
            匹配证据
          </h3>
          <div className="border border-gray-200 rounded-md overflow-hidden">
            <div className="grid grid-cols-12 gap-4 bg-gray-50 p-3 text-[13px] font-medium text-gray-600 border-b border-gray-200">
              <div className="col-span-5">岗位要求</div>
              <div className="col-span-5">简历证据</div>
              <div className="col-span-2 text-center">证据强度</div>
            </div>
            <div className="divide-y divide-gray-100 text-[13px]">
              {recVm.evidenceChain.map((ev, i) => (
                <div key={i} className="grid grid-cols-12 gap-4 p-3 items-center hover:bg-gray-50 transition-colors">
                  <div className="col-span-5 text-gray-800">{ev.jdRequirement}</div>
                  <div className="col-span-5 text-gray-600">{ev.resumeEvidence}</div>
                  <div className="col-span-2 flex justify-center">
                    <Badge className={getStrengthStyle(ev.evidenceStrengthTone) + " font-normal"}>{ev.evidenceStrengthLabel}</Badge>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </section>

        {/* Module 2: Risks */}
        <section>
          <h3 className="font-bold text-gray-900 mb-4 text-[16px] flex items-center gap-2">
            <span className="w-1 h-4 bg-amber-400 rounded-full inline-block"></span>
            主要风险与提醒
          </h3>
          <div className="bg-amber-50/30 border border-amber-100/50 rounded-md p-4">
            <ul className="space-y-2.5">
              {recVm.riskItems.map((risk, i) => (
                <li key={i} className="flex items-center gap-2.5 text-[13px] text-amber-700">
                  <div className="w-1.5 h-1.5 rounded-full bg-amber-400/80 shrink-0"></div>
                  {risk}
                </li>
              ))}
            </ul>
          </div>
        </section>

        {/* Module 3: Suggestions & Actions */}
        <section className="bg-gray-50 rounded-md p-5 border border-gray-100">
          <div className="flex items-start gap-3 mb-5">
            <Zap className="w-5 h-5 text-[#0052D9] shrink-0 mt-0.5" />
            <div className="text-[14px] text-gray-700 leading-relaxed">
              <p className="font-medium text-gray-900 mb-1">投递建议</p>
              {recVm.nextActions[0] || "建议优先进入岗位详情页，再进行岗位匹配分析。"}<br/>
              {recVm.nextActions[1] || "如果时间有限，可先修复 Top 3 简历问题后投递。"}
            </div>
          </div>
          <div className="flex gap-4">
            <Button className="flex-1 bg-[#0052D9] hover:bg-[#0040A8] text-white h-11" onClick={() => navigate("/jobs/1")}>
              进入岗位详情<ArrowRight className="w-4 h-4 ml-2" />
            </Button>
            <Button variant="outline" className="flex-1 border-[#0052D9] text-[#0052D9] hover:bg-blue-50 h-11" onClick={() => navigate("/match")}>
              进行岗位匹配分析
            </Button>
          </div>
        </section>
      </div>
    </div>
  );
}
