import React from "react";
import { CheckCircle2, ShieldAlert, Zap, ArrowRight } from "lucide-react";
import { Button } from "../ui/Button";
import { Badge } from "../ui/Badge";
import { useNavigate } from "react-router";
import type { RecommendationPageViewModel } from "../../../adapters";

/** K8-5-28-C: map canonical module key to Chinese label */
const MODULE_LABELS: Record<string, string> = {
  resume_basics: "简历基础",
  basic_info: "基础信息",
  education: "教育经历",
  projects: "项目经历",
  skills: "技能信息",
  internships: "实习经历",
  certificates_awards: "证书与获奖",
  portfolio: "作品或个人主页",
  extra_info: "自我评价",
  overall: "综合判断",
};

function moduleLabel(m?: string): string {
  if (!m) return "";
  return MODULE_LABELS[m] ?? m;
}

interface RecommendationDetailProps {
  currentResumeId?: string;
  data?: RecommendationPageViewModel | null;
}

export function RecommendationDetail({ currentResumeId = "", data }: RecommendationDetailProps) {
  const navigate = useNavigate();

  if (!data || !data.recommendedJobs.length) {
    return (
      <div className="flex-1 min-w-[560px] max-w-[680px] bg-white border border-gray-200 rounded-lg shadow-sm overflow-hidden flex items-center justify-center py-20">
        <p className="text-gray-400 text-sm">暂无推荐结果，请先选择简历并发起推荐。</p>
      </div>
    );
  }

  const firstJob = data.recommendedJobs[0];

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
            <div className="grid grid-cols-12 gap-2 bg-gray-50 p-3 text-[13px] font-medium text-gray-600 border-b border-gray-200">
              <div className="col-span-2">模块</div>
              <div className="col-span-4">岗位要求</div>
              <div className="col-span-4">简历证据</div>
              <div className="col-span-2 text-center">证据强度</div>
            </div>
            <div className="divide-y divide-gray-100 text-[13px]">
              {data.evidenceChain.map((ev, i) => (
                <div key={i} className="grid grid-cols-12 gap-2 p-3 items-center hover:bg-gray-50 transition-colors">
                  <div className="col-span-2">
                    {ev.resumeModule ? (
                      <Badge className="bg-blue-50 text-blue-700 border-blue-200 font-normal text-[11px]">{moduleLabel(ev.resumeModule)}</Badge>
                    ) : (
                      <span className="text-gray-300 text-[11px]">—</span>
                    )}
                  </div>
                  <div className="col-span-4 text-gray-800">{ev.jdRequirement}</div>
                  <div className="col-span-4 text-gray-600">{ev.resumeEvidence}</div>
                  <div className="col-span-2 flex justify-center">
                    <Badge className={getStrengthStyle(ev.evidenceStrengthTone) + " font-normal"}>{ev.evidenceStrengthLabel}</Badge>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </section>

        {/* Module 2: Resume fix suggestions (TopFixItems) */}
        {data.topFixItems && data.topFixItems.length > 0 && (
        <section>
          <h3 className="font-bold text-gray-900 mb-4 text-[16px] flex items-center gap-2">
            <span className="w-1 h-4 bg-emerald-400 rounded-full inline-block"></span>
            简历补强建议
          </h3>
          <div className="space-y-3">
            {data.topFixItems.map((item, i) => (
              <div key={i} className="flex items-start gap-3 p-4 bg-gray-50 border border-gray-100 rounded-md hover:border-emerald-200 transition-colors">
                <div className="flex items-center justify-center w-6 h-6 rounded-full bg-white border border-gray-300 text-gray-700 text-[13px] font-bold shrink-0 mt-0.5">
                  {i + 1}
                </div>
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2 mb-1.5">
                    {item.resumeModule ? (
                      <Badge className="bg-emerald-50 text-emerald-700 border-emerald-200 font-normal text-[11px]">{moduleLabel(item.resumeModule)}</Badge>
                    ) : null}
                    <span className={`text-[11px] px-1.5 rounded ${item.priorityTone === 'orange' ? 'text-red-600 bg-red-50' : 'text-blue-600 bg-blue-50'}`}>
                      {item.priorityLabel}
                    </span>
                  </div>
                  <div className="text-[14px] font-medium text-gray-900 mb-1">{item.title}</div>
                  {item.expectedImpact && (
                    <div className="text-[12px] text-gray-500">
                      预期影响：{item.expectedImpact}
                      {item.needUserVerify ? "（需用户核实）" : ""}
                    </div>
                  )}
                  {item.description && (
                    <div className="text-[12px] text-gray-600 mt-1.5 bg-white border border-gray-100 rounded p-2">
                      {item.description}
                    </div>
                  )}
                </div>
              </div>
            ))}
          </div>
        </section>
        )}

        {/* Module 2: Risks */}
        <section>
          <h3 className="font-bold text-gray-900 mb-4 text-[16px] flex items-center gap-2">
            <span className="w-1 h-4 bg-amber-400 rounded-full inline-block"></span>
            主要风险与提醒
          </h3>
          <div className="bg-amber-50/30 border border-amber-100/50 rounded-md p-4">
            <ul className="space-y-2.5">
              {data.riskItems.map((risk, i) => (
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
              {data.nextActions[0] || "建议优先进入岗位详情页，再进行岗位匹配分析。"}<br/>
              {data.nextActions[1] || "如果时间有限，可先修复 Top 3 简历问题后投递。"}
            </div>
          </div>
          <div className="flex gap-4">
            <Button className="flex-1 bg-[#0052D9] hover:bg-[#0040A8] text-white h-11" onClick={() => navigate(`/jobs/${encodeURIComponent(firstJob.jobId)}`)}>
              进入岗位详情<ArrowRight className="w-4 h-4 ml-2" />
            </Button>
            <Button variant="outline" className="flex-1 border-[#0052D9] text-[#0052D9] hover:bg-blue-50 h-11" onClick={() => navigate(`/match?jobId=${encodeURIComponent(firstJob.jobId)}&resumeId=${encodeURIComponent(currentResumeId)}`)}>
              进行岗位匹配分析
            </Button>
          </div>
        </section>
      </div>
    </div>
  );
}
