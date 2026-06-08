import React from "react";
import { CheckCircle2, AlertTriangle, Zap, Sparkles, PlusCircle, Lightbulb, Target, ArrowRight, TrendingUp, Shield } from "lucide-react";
import { Button } from "../ui/Button";
import { Badge } from "../ui/Badge";
import { useNavigate } from "react-router";
import type { MatchAnalysisPageViewModel } from "../../../adapters";

interface MatchAnalysisDetailProps {
  currentJobId?: string;
  currentResumeId?: string;
  data?: MatchAnalysisPageViewModel | null;
}

function readinessToUserText(ar: string): string {
  if (!ar) return "建议先补强关键证据后再投递";
  if (ar.includes("ready")) return "可以投递，建议微调简历突出关键优势";
  if (ar.includes("prepare_first")) return "建议先补强关键证据后再投递";
  if (ar.includes("cautious_try")) return "谨慎尝试，建议作为备选岗位";
  if (ar.includes("not_recommended")) return "不建议当前投递此岗位";
  return ar;
}

function readinessToLabel(ar: string): string {
  if (!ar) return "先补强再投递";
  if (ar.includes("ready")) return "可以投递";
  if (ar.includes("prepare_first")) return "先补强再投递";
  if (ar.includes("cautious_try")) return "谨慎尝试";
  return "不建议";
}

function judgementToText(j: string): string {
  switch (j) {
  case "匹配": return "已匹配";
  case "部分匹配": return "部分匹配";
  case "缺口": return "待补强";
  default: return j || "部分匹配";
  }
}

function judgementTone(j: string): "green" | "amber" | "gray" {
  if (j === "匹配") return "green";
  if (j === "部分匹配") return "amber";
  return "gray";
}

function scoreColor(score: number): string {
  if (score >= 85) return "text-[#0052D9]";
  if (score >= 70) return "text-amber-600";
  return "text-gray-500";
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

  const hasEvChain = data.evidenceChain && data.evidenceChain.length > 0;
  const hasGaps = data.gapPoints && data.gapPoints.length > 0;
  const hasFixes = data.topFixItems && data.topFixItems.length > 0;
  const hasTalks = data.interviewTalkingPoints && data.interviewTalkingPoints.length > 0;
  const hasStrategy = data.deliveryStrategy && data.deliveryStrategy.length > 0;
  const isHighScore = data.matchScore >= 85;
  const isMedium = data.matchScore >= 70 && data.matchScore < 85;
  const sColor = scoreColor(data.matchScore);

  return (
    <div className="flex-1 min-w-[640px] max-w-[760px] bg-white border border-gray-200 rounded-lg shadow-sm overflow-hidden flex flex-col">

      {/* ═══ Section 1: AI 投递决策卡 ═══ */}
      <div className="bg-gradient-to-br from-[#F4F8FF] to-white p-6 border-b border-blue-100">
        <h2 className="text-[18px] font-bold text-gray-900 mb-5 flex items-center gap-2">
          <TrendingUp className="w-5 h-5 text-[#0052D9]" />
          AI 投递决策
        </h2>

        <div className="grid grid-cols-2 gap-4 mb-4">
          {/* Score */}
          <div className="bg-white p-4 rounded-xl border border-blue-100 shadow-sm">
            <div className="text-[12px] text-gray-400 mb-1">岗位匹配度</div>
            <div className={`text-4xl font-black ${sColor} flex items-baseline`}>
              {data.matchScore}<span className="text-[12px] text-gray-400 font-medium ml-1">/ 100</span>
            </div>
          </div>
          {/* Delivery recommendation */}
          <div className={`p-4 rounded-xl border shadow-sm ${
            isHighScore ? "bg-green-50/50 border-green-100" : isMedium ? "bg-amber-50/50 border-amber-100" : "bg-gray-50 border-gray-200"
          }`}>
            <div className="text-[12px] text-gray-400 mb-1">投递建议</div>
            <div className={`text-[15px] font-bold ${isHighScore ? "text-green-700" : isMedium ? "text-amber-700" : "text-gray-700"}`}>
              {readinessToLabel(data.applyReadiness)}
            </div>
            <div className="text-[12px] text-gray-500 mt-1 leading-relaxed">{readinessToUserText(data.applyReadiness)}</div>
          </div>
        </div>

        {/* Key gaps */}
        {hasGaps && (
          <div className="bg-white/70 p-4 rounded-xl border border-blue-50">
            <div className="text-[12px] text-gray-400 mb-2">投递前提醒</div>
            <ul className="space-y-1">
              {data.gapPoints!.slice(0, 3).map((g, i) => (
                <li key={i} className="text-[13px] text-amber-800 flex items-start gap-2">
                  <span className="text-amber-400 mt-1 shrink-0">⚠</span> {g}
                </li>
              ))}
            </ul>
          </div>
        )}
      </div>

      <div className="p-6 space-y-8">

        {/* ═══ Section 2: 岗位要求 × 简历证据矩阵 ═══ */}
        {hasEvChain && (
          <section>
            <h3 className="font-bold text-gray-900 mb-4 text-[16px] flex items-center gap-2">
              <Target className="w-4 h-4 text-[#0052D9]" />
              岗位要求 × 简历证据矩阵
            </h3>
            <div className="border border-gray-200 rounded-lg overflow-hidden">
              <div className="grid grid-cols-12 gap-3 bg-gray-50 p-3 text-[12px] font-medium text-gray-500 border-b border-gray-200">
                <div className="col-span-3">岗位要求</div>
                <div className="col-span-4">简历证据</div>
                <div className="col-span-2 text-center">匹配判断</div>
                <div className="col-span-3">怎么补强</div>
              </div>
              <div className="divide-y divide-gray-100 text-[12px] leading-relaxed">
                {data.evidenceChain.map((ev, i) => {
                  const tone = judgementTone(ev.evidenceStrengthLabel === "匹配" ? "匹配" : ev.evidenceStrengthLabel?.includes("部分") ? "部分匹配" : "缺口");
                  const bg = tone === "green" ? "bg-green-50 text-green-700" : tone === "amber" ? "bg-amber-50 text-amber-700" : "bg-gray-50 text-gray-600";
                  return (
                    <div key={i} className="grid grid-cols-12 gap-3 p-3 items-start hover:bg-gray-50 transition-colors">
                      <div className="col-span-3 text-gray-800">{ev.jdRequirement}</div>
                      <div className="col-span-4 text-gray-600">{ev.resumeEvidence}</div>
                      <div className="col-span-2 flex justify-center">
                        <span className={`px-2 py-0.5 rounded text-[11px] font-medium ${bg}`}>{judgementToText(ev.evidenceStrengthLabel === "匹配" ? "匹配" : ev.evidenceStrengthLabel?.includes("部分") ? "部分匹配" : "缺口")}</span>
                      </div>
                      <div className="col-span-3 text-gray-500">{ev.fixAction || ev.riskNote || "补充与该要求直接相关的项目证据、结果指标或技能使用场景"}</div>
                    </div>
                  );
                })}
              </div>
            </div>
          </section>
        )}

        {/* ═══ Section 3: 投递前优先补强 3 件事 ═══ */}
        {(hasGaps || hasFixes) && (
          <section>
            <h3 className="font-bold text-gray-900 mb-4 text-[16px] flex items-center gap-2">
              <Shield className="w-4 h-4 text-amber-500" />
              投递前优先补强{data.topFixItems?.length || 3}件事
            </h3>
            <div className="space-y-3">
              {data.topFixItems && data.topFixItems.length > 0 ? (
                data.topFixItems.slice(0, 3).map((task, i) => (
                  <div key={i} className="flex gap-3 p-4 bg-amber-50/30 border border-amber-100/50 rounded-lg">
                    <div className="flex items-center justify-center w-7 h-7 rounded-full bg-amber-100 text-amber-700 text-[13px] font-bold shrink-0 mt-0.5">{i + 1}</div>
                    <div className="flex-1 min-w-0">
                      <div className="text-[14px] font-bold text-gray-900 mb-1">{task.title}</div>
                      {task.description && <div className="text-[12px] text-gray-500 mb-2">{task.description}</div>}
                      <div className="flex items-center gap-3 text-[11px] text-gray-400">
                        <span>优先级：<span className="text-amber-600 font-medium">{task.priorityLabel}</span></span>
                        {task.expectedImpact && <span>预计投入：<span className="text-gray-600">{task.expectedImpact}</span></span>}
                      </div>
                    </div>
                  </div>
                ))
              ) : hasGaps ? (
                data.gapPoints!.slice(0, 3).map((g, i) => (
                  <div key={i} className="flex gap-3 p-4 bg-amber-50/30 border border-amber-100/50 rounded-lg">
                    <div className="flex items-center justify-center w-7 h-7 rounded-full bg-amber-100 text-amber-700 text-[13px] font-bold shrink-0 mt-0.5">{i + 1}</div>
                    <div className="flex-1 text-[13px] text-gray-700">{g}</div>
                  </div>
                ))
              ) : null}
            </div>
          </section>
        )}

        {/* ═══ Section 4: 面试可以这样讲 ═══ */}
        {hasTalks && (
          <section>
            <h3 className="font-bold text-gray-900 mb-4 text-[16px] flex items-center gap-2">
              <Lightbulb className="w-4 h-4 text-purple-500" />
              面试可以这样讲
            </h3>
            <div className="space-y-2">
              {data.interviewTalkingPoints!.map((pt, i) => (
                <div key={i} className="flex items-start gap-3 bg-purple-50/20 px-4 py-3 rounded-lg border border-purple-100/30 group hover:bg-purple-50/40 transition-colors">
                  <div className="flex items-center justify-center w-6 h-6 rounded-full bg-purple-100 text-purple-700 text-[12px] font-bold shrink-0 mt-0.5">{i + 1}</div>
                  <p className="text-[13px] text-gray-700 leading-relaxed">{pt}</p>
                </div>
              ))}
            </div>
          </section>
        )}

        {/* ═══ Section 5: 投递策略与下一步 ═══ */}
        <section className="bg-gray-50 rounded-xl p-5 border border-gray-200">
          <h3 className="font-bold text-gray-900 mb-4 text-[16px] flex items-center gap-2">
            <Sparkles className="w-4 h-4 text-[#0052D9]" />
            投递策略与下一步
          </h3>

          {hasStrategy && (
            <div className="bg-white p-4 rounded-lg border border-gray-100 mb-4 text-[13px] text-gray-700 leading-relaxed whitespace-pre-wrap">
              {data.deliveryStrategy}
            </div>
          )}

          <div className="mb-4">
            <div className="text-[12px] text-gray-400 mb-2">建议按以下步骤推进</div>
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
              <p className="text-[13px] text-gray-500">建议先修复 Top 3 简历问题，再进入投递。</p>
            )}
          </div>

          <div className="flex items-center gap-3">
            <Button
              className="flex-1 bg-[#0052D9] hover:bg-[#0040A8] text-white h-10 text-[13px]"
              onClick={() => navigate(`/resume-optimization?jobId=${encodeURIComponent(currentJobId)}&resumeId=${encodeURIComponent(currentResumeId)}`)}
            >
              <PlusCircle className="w-3.5 h-3.5 mr-1.5" />
              针对该岗位优化简历
            </Button>
            <Button variant="outline" className="px-5 border-gray-300 text-gray-700 hover:bg-gray-50 h-10 text-[13px]" onClick={() => navigate(-1)}>
              返回岗位详情
            </Button>
            <Button variant="outline" className="px-5 border-transparent text-[#0052D9] hover:bg-blue-50 h-10 text-[13px]" onClick={() => navigate(`/recommend?resumeId=${encodeURIComponent(currentResumeId)}`)}>
              查看推荐岗位
            </Button>
          </div>
        </section>

      </div>
    </div>
  );
}
