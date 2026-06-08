import React from "react";
import { CheckCircle2, ShieldAlert, Zap, Sparkles, Activity, PlusCircle, AlertTriangle, BookOpen, Lightbulb, Target, ArrowRight } from "lucide-react";
import { Button } from "../ui/Button";
import { Badge } from "../ui/Badge";
import { useNavigate } from "react-router";
import type { MatchAnalysisPageViewModel } from "../../../adapters";

interface MatchAnalysisDetailProps {
  currentJobId?: string;
  currentResumeId?: string;
  data?: MatchAnalysisPageViewModel | null;
}

export function MatchAnalysisDetail({ currentJobId = "", currentResumeId = "", data }: MatchAnalysisDetailProps) {
  const navigate = useNavigate();

  if (!data) {
    return (
      <div className="flex-1 min-w-[640px] max-w-[760px] bg-white border border-gray-200 rounded-lg shadow-sm overflow-hidden flex items-center justify-center py-20">
        <p className="text-gray-400 text-sm">暂无匹配分析结果，请先选择岗位和简历发起分析。</p>
      </div>
    );
  }

  const getStrengthStyle = (tone: string) => {
    switch (tone) {
      case 'green': return "bg-green-50 text-green-700 border border-green-200";
      case 'blue': return "bg-blue-50 text-[#0052D9] border border-blue-200";
      case 'orange': return "bg-amber-50 text-amber-700 border border-amber-200";
      default: return "bg-gray-50 text-gray-700";
    }
  };

  const hasEvChain = data.evidenceChain && data.evidenceChain.length > 0;
  const hasGaps = data.gapPoints && data.gapPoints.length > 0;
  const hasFixes = data.topFixItems && data.topFixItems.length > 0;
  const hasTalks = data.interviewTalkingPoints && data.interviewTalkingPoints.length > 0;
  const hasStrategy = data.deliveryStrategy && data.deliveryStrategy.length > 0;
  const hasResumeEv = data.resumeEvidence && data.resumeEvidence.length > 0;
  const hasJobReqs = data.jobRequirements && data.jobRequirements.length > 0;

  return (
    <div className="flex-1 min-w-[640px] max-w-[760px] bg-white border border-gray-200 rounded-lg shadow-sm overflow-hidden flex flex-col">
      {/* Module 1: Conclusion */}
      <div className="bg-gradient-to-br from-[#F4F8FF] to-white p-6 border-b border-blue-100">
        <h2 className="text-[18px] font-bold text-gray-900 mb-5 flex items-center gap-2">
          <Activity className="w-5 h-5 text-[#0052D9]" />
          匹配结论
        </h2>

        <div className="flex flex-col gap-5">
          <div className="grid grid-cols-3 gap-4">
            <div className="bg-white p-4 rounded-lg border border-blue-100 shadow-sm flex flex-col items-center justify-center text-center">
              <div className="text-[13px] text-gray-500 mb-1">匹配分数</div>
              <div className="text-3xl font-black text-[#0052D9] flex items-baseline">
                {data.matchScore}<span className="text-[12px] text-gray-400 font-medium ml-1">/ 100</span>
              </div>
            </div>
            <div className="bg-white p-4 rounded-lg border border-amber-100 shadow-sm flex flex-col items-center justify-center text-center">
              <div className="text-[13px] text-gray-500 mb-1">投递准备度</div>
              <div className="text-[16px] font-bold text-amber-600">{data.applyReadiness || "先修复后投递"}</div>
            </div>
            <div className="bg-white p-4 rounded-lg border border-gray-200 shadow-sm flex flex-col items-center justify-center text-center">
              <div className="text-[13px] text-gray-500 mb-1">优化停止信号</div>
              <div className="text-[15px] font-bold text-gray-800">{data.optimizationStopSignal || "修复 Top 3 后再投递"}</div>
            </div>
          </div>
          {data.summary && (
            <p className="text-[14px] text-gray-700 leading-relaxed bg-white/60 p-3.5 rounded border border-blue-50">
              <span className="font-semibold text-gray-900">一句话总结：</span>
              {data.summary}
            </p>
          )}
        </div>
      </div>

      <div className="p-6 space-y-8">
        {/* Module 2: Job Requirements */}
        {hasJobReqs && (
          <section>
            <h3 className="font-bold text-gray-900 mb-4 text-[16px] flex items-center gap-2">
              <Target className="w-4 h-4 text-[#0052D9]" />
              岗位关键要求
            </h3>
            <ul className="grid grid-cols-2 gap-2">
              {data.jobRequirements!.map((req, i) => (
                <li key={i} className="flex items-start gap-2 bg-blue-50/30 px-3 py-2 rounded text-[13px] text-gray-700">
                  <span className="text-[#0052D9] mt-1 shrink-0">▸</span> {req}
                </li>
              ))}
            </ul>
          </section>
        )}

        {/* Module 3: Resume Evidence */}
        {hasResumeEv && (
          <section>
            <h3 className="font-bold text-gray-900 mb-4 text-[16px] flex items-center gap-2">
              <BookOpen className="w-4 h-4 text-green-600" />
              简历证据对照
            </h3>
            <ul className="space-y-2">
              {data.resumeEvidence!.map((ev, i) => (
                <li key={i} className="flex items-start gap-2 bg-green-50/30 px-3 py-2 rounded text-[13px] text-gray-700 border border-green-100/30">
                  <CheckCircle2 className="w-3.5 h-3.5 text-green-500 mt-0.5 shrink-0" /> {ev}
                </li>
              ))}
            </ul>
          </section>
        )}

        {/* Module 4: Evidence Chain */}
        {hasEvChain && (
          <section>
            <h3 className="font-bold text-gray-900 mb-4 text-[16px] flex items-center gap-2">
              <span className="w-1 h-4 bg-[#0052D9] rounded-full inline-block"></span>
              JD 证据链分析
            </h3>
            <div className="border border-gray-200 rounded-md overflow-hidden">
              <div className="grid grid-cols-12 gap-4 bg-gray-50 p-3 text-[12px] font-medium text-gray-600 border-b border-gray-200">
                <div className="col-span-3">JD 要求</div>
                <div className="col-span-4">简历证据</div>
                <div className="col-span-2 text-center">证据强度</div>
                <div className="col-span-3">风险说明</div>
              </div>
              <div className="divide-y divide-gray-100 text-[12px] leading-relaxed">
                {data.evidenceChain.map((ev, i) => (
                  <div key={i} className="grid grid-cols-12 gap-4 p-3 items-start hover:bg-gray-50 transition-colors">
                    <div className="col-span-3 text-gray-800">{ev.jdRequirement}</div>
                    <div className="col-span-4 text-gray-600">{ev.resumeEvidence}</div>
                    <div className="col-span-2 flex justify-center">
                      <Badge className={getStrengthStyle(ev.evidenceStrengthTone) + " font-normal"}>{ev.evidenceStrengthLabel}</Badge>
                    </div>
                    <div className="col-span-3 text-gray-500">{ev.riskNote || "-----"}</div>
                  </div>
                ))}
              </div>
            </div>
          </section>
        )}

        {/* Module 5: Gaps + Top Fixes */}
        {(hasGaps || hasFixes) && (
          <section>
            <h3 className="font-bold text-gray-900 mb-4 text-[16px] flex items-center gap-2">
              <AlertTriangle className="w-4 h-4 text-amber-500" />
              缺口分析与修复建议
            </h3>
            {hasGaps && (
              <ul className="space-y-2 mb-4">
                {data.gapPoints!.map((gap, i) => (
                  <li key={i} className="flex items-start gap-2 bg-amber-50/30 px-3 py-2 rounded text-[13px] text-amber-800 border border-amber-100/30">
                    <span className="text-amber-500 mt-1 shrink-0">⚠</span> {gap}
                  </li>
                ))}
              </ul>
            )}
            {hasFixes && (
              <div className="space-y-2">
                {data.topFixItems.map((task, i) => (
                  <div key={i} className="flex items-start gap-3 p-3 bg-gray-50 border border-gray-100 rounded-md hover:border-blue-100 transition-colors">
                    <div className="flex items-center justify-center w-6 h-6 rounded-full bg-white border border-gray-300 text-gray-700 text-[13px] font-bold shrink-0 mt-0.5">{i + 1}</div>
                    <div className="flex-1">
                      <div className="text-[14px] font-bold text-gray-900 mb-1.5">{task.title}</div>
                      <div className="text-[12px] text-gray-600 leading-relaxed">{task.description}</div>
                      <div className="flex items-center gap-3 mt-2 text-[11px]">
                        <span className="text-gray-400">优先级：<span className={task.priorityTone === 'orange' ? "text-red-600 font-medium" : "text-amber-600 font-medium"}>{task.priorityLabel}</span></span>
                        <span className="text-gray-400">预计耗时：<span className="text-gray-700">{task.expectedImpact}</span></span>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </section>
        )}

        <div className="grid grid-cols-2 gap-6">
          {/* Module 6: Strengths */}
          <section>
            <h3 className="font-bold text-gray-900 mb-3 text-[15px] flex items-center gap-2">
              <CheckCircle2 className="w-4 h-4 text-green-600" />
              你的优势
            </h3>
            <ul className="space-y-2 text-[13px] text-gray-700 bg-green-50/30 p-4 rounded-md border border-green-100/50">
              {data.matchHighlights.map((item, i) => (
                <li key={i} className="flex items-start gap-2">
                  <div className="w-1.5 h-1.5 rounded-full bg-green-500 shrink-0 mt-1.5"></div>
                  {item}
                </li>
              ))}
            </ul>
          </section>

          {/* Module 7: Risks */}
          <section>
            <h3 className="font-bold text-gray-900 mb-3 text-[15px] flex items-center gap-2">
              <ShieldAlert className="w-4 h-4 text-amber-500" />
              主要风险
            </h3>
            <ul className="space-y-2 text-[13px] text-amber-800 bg-amber-50/30 p-4 rounded-md border border-amber-100/50">
              {data.riskItems.map((item, i) => (
                <li key={i} className="flex items-start gap-2">
                  <div className="w-1.5 h-1.5 rounded-full bg-amber-400 shrink-0 mt-1.5"></div>
                  {item}
                </li>
              ))}
            </ul>
          </section>
        </div>

        {/* Module 8: Interview Talking Points */}
        {hasTalks && (
          <section>
            <h3 className="font-bold text-gray-900 mb-4 text-[16px] flex items-center gap-2">
              <Lightbulb className="w-4 h-4 text-purple-500" />
              面试表达要点
            </h3>
            <ul className="space-y-2">
              {data.interviewTalkingPoints!.map((pt, i) => (
                <li key={i} className="flex items-start gap-2 bg-purple-50/30 px-3 py-2 rounded text-[13px] text-gray-700 border border-purple-100/30">
                  <span className="text-purple-500 mt-1 shrink-0 font-bold">{i + 1}.</span> {pt}
                </li>
              ))}
            </ul>
          </section>
        )}

        {/* Module 9: Delivery Strategy */}
        {hasStrategy && (
          <section>
            <h3 className="font-bold text-gray-900 mb-4 text-[16px] flex items-center gap-2">
              <Sparkles className="w-4 h-4 text-[#0052D9]" />
              投递策略
            </h3>
            <div className="bg-blue-50/20 p-4 rounded-md border border-blue-100/50 text-[13px] text-gray-700 leading-relaxed whitespace-pre-wrap">
              {data.deliveryStrategy}
            </div>
          </section>
        )}

        {/* Module 10: Next Steps */}
        <section className="bg-gray-50 rounded-md p-5 border border-gray-200">
          <div className="flex items-start gap-3 mb-5">
            <Zap className="w-5 h-5 text-[#0052D9] shrink-0 mt-0.5" />
            <div>
              <p className="font-bold text-gray-900 mb-2">下一步建议</p>
              {data.nextActions && data.nextActions.length > 0 ? (
                <ul className="space-y-1.5">
                  {data.nextActions.map((step, i) => (
                    <li key={i} className="flex items-start gap-2 text-[13px] text-gray-700">
                      <ArrowRight className="w-3 h-3 text-[#0052D9] mt-1 shrink-0" />
                      {step}
                    </li>
                  ))}
                </ul>
              ) : (
                <p className="text-[14px] text-gray-500">建议先修复 Top 3 简历问题，再进入投递。</p>
              )}
            </div>
          </div>

          <div className="flex items-center gap-4">
            <Button
              className="flex-1 bg-[#0052D9] hover:bg-[#0040A8] text-white h-11"
              onClick={() => navigate(`/resume-optimization?jobId=${encodeURIComponent(currentJobId)}&resumeId=${encodeURIComponent(currentResumeId)}`)}
            >
              <PlusCircle className="w-4 h-4 mr-2" />
              针对该岗位优化简历
            </Button>
            <Button
              variant="outline"
              className="px-6 border-gray-300 text-gray-700 hover:bg-gray-50 h-11"
              onClick={() => navigate(-1)}
            >
              返回岗位详情
            </Button>
            <Button
              variant="outline"
              className="px-6 border-transparent text-[#0052D9] hover:bg-blue-50 h-11"
              onClick={() => navigate(`/recommend?resumeId=${encodeURIComponent(currentResumeId)}`)}
            >
              查看推荐岗位
            </Button>
          </div>
        </section>
      </div>
    </div>
  );
}
