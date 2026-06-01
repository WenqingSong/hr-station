import React, { useState } from "react";
import { Activity, Wrench, Sparkles, AlertCircle, RefreshCw, FileEdit, CheckCircle2, ChevronRight, ChevronDown, ChevronUp } from "lucide-react";
import { Button } from "./ui/Button";
import { Badge } from "./ui/Badge";
import { useNavigate } from "react-router";
import type { ResumeOptimizationPageViewModel } from "../../adapters";

interface ResumeOptimizationDetailProps {
  currentJobId?: string;
  currentResumeId?: string;
  data?: ResumeOptimizationPageViewModel | null;
}

const CollapsibleSection = ({ title, icon: Icon, children, defaultOpen = false }: { title: string, icon: any, children: React.ReactNode, defaultOpen?: boolean }) => {
  const [isOpen, setIsOpen] = useState(defaultOpen);
  return (
    <section className="border border-gray-200 rounded-md overflow-hidden bg-white shadow-sm">
      <button 
        onClick={() => setIsOpen(!isOpen)}
        className="w-full flex items-center justify-between p-4 bg-gray-50 hover:bg-gray-100 transition-colors"
      >
        <div className="font-bold text-gray-700 text-[14px] flex items-center gap-2">
          <Icon className="w-4 h-4 text-gray-500" />
          {title}
        </div>
        {isOpen ? <ChevronUp className="w-4 h-4 text-gray-400" /> : <ChevronDown className="w-4 h-4 text-gray-400" />}
      </button>
      {isOpen && <div className="p-4 border-t border-gray-200">{children}</div>}
    </section>
  );
};

export function ResumeOptimizationDetail({ currentJobId = "", currentResumeId = "", data }: ResumeOptimizationDetailProps) {
  const navigate = useNavigate();

  if (!data) {
    return (
      <div className="flex-1 min-w-[660px] max-w-[780px] bg-white border border-gray-200 rounded-lg shadow-sm overflow-hidden flex items-center justify-center py-20">
        <p className="text-gray-400 text-sm">暂无简历优化建议，请先选择目标岗位和简历发起优化。</p>
      </div>
    );
  }

  return (
    <div className="flex-1 min-w-[660px] max-w-[780px] bg-white border border-gray-200 rounded-lg shadow-sm overflow-hidden flex flex-col">
      {/* Module 1: Optimization Conclusion */}
      <div className="bg-gradient-to-br from-[#F4F8FF] to-white p-6 border-b border-blue-100">
        <h2 className="text-[16px] font-bold text-gray-900 mb-4 flex items-center gap-2">
          <Activity className="w-4 h-4 text-[#0052D9]" />
          优化结论
        </h2>
        
        <div className="flex flex-col gap-4">
          <div className="grid grid-cols-2 gap-4">
            <div className="bg-white px-5 py-4 rounded-lg border-l-4 border-l-amber-500 border-y border-r border-gray-200 shadow-sm flex items-center justify-between">
              <div className="text-[13px] text-gray-500">投递准备度</div>
              <div className="text-[15px] font-bold text-amber-600">{data.applyReadiness || "先修复后投递"}</div>
            </div>
            
            <div className="bg-white px-5 py-4 rounded-lg border-l-4 border-l-[#0052D9] border-y border-r border-gray-200 shadow-sm flex items-center justify-between">
              <div className="text-[13px] text-gray-500">优化停止信号</div>
              <div className="text-[15px] font-bold text-[#0052D9]">{data.optimizationStopSignal || "修复 Top 3 后再投递"}</div>
            </div>
          </div>
          
          <p className="text-[13px] text-gray-700 leading-relaxed bg-blue-50/50 p-3 rounded">
            <span className="font-bold text-gray-900 mr-2">诊断摘要：</span>
            {data.summary || "请基于岗位要求优化简历表达。"}
          </p>
        </div>
      </div>

      <div className="p-6 space-y-8">
        {/* Module 2: Top 3 Quick Fixes */}
        <section>
          <h3 className="font-bold text-gray-900 mb-4 text-[16px] flex items-center gap-2">
            <Wrench className="w-4 h-4 text-[#0052D9]" />
            投递前最应该先修的 3 件事
          </h3>
          <div className="space-y-4">
            {data.topFixItems.map((task, i) => (
              <div key={i} className="flex items-start gap-3 p-4 bg-gray-50 border border-gray-200 rounded-md">
                <div className="flex items-center justify-center w-6 h-6 rounded-full bg-white border border-gray-300 text-gray-700 text-[13px] font-bold shrink-0">
                  {i + 1}
                </div>
                <div className="flex-1">
                  <div className="text-[15px] font-bold text-gray-900 mb-2">{task.title}</div>
                  <div className="grid grid-cols-1 sm:grid-cols-3 gap-y-2 gap-x-4 text-[13px] mb-3">
                    <div className="flex items-center gap-1.5">
                      <span className="text-gray-500">优先级：</span>
                      <span className={task.priorityTone === 'orange' ? "text-red-600 font-medium" : "text-blue-600 font-medium"}>
                        {task.priorityLabel}
                      </span>
                    </div>
                    <div className="flex items-center gap-1.5 sm:col-span-2">
                      <span className="text-gray-500">预计影响：</span>
                      <span className="text-gray-800">{task.expectedImpact}</span>
                    </div>
                  </div>
                  <div className="bg-amber-50/50 border border-amber-100/50 p-2.5 rounded text-[12px] text-gray-700 flex items-start gap-2">
                    <AlertCircle className="w-3.5 h-3.5 text-amber-500 shrink-0 mt-0.5" />
                    <div>
                      <span className="font-semibold text-amber-800 mr-1">需要用户核实：</span>
                      {task.description || (task.needUserVerify ? "是" : "否")}
                    </div>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </section>

        {/* Module: Verification Checklist */}
        <section>
          <h3 className="font-bold text-gray-900 mb-4 text-[16px] flex items-center gap-2">
            <CheckCircle2 className="w-4 h-4 text-[#0052D9]" />
            需要你核实后再写入简历的信息
          </h3>
          <div className="bg-orange-50/40 border border-orange-100 rounded-md p-4">
            <ul className="space-y-3 text-[13px]">
              {data.missingEvidencePlan.map((item, i) => (
                <li key={i} className="flex items-center justify-between gap-4 border-b border-orange-100/50 pb-2 last:border-0 last:pb-0">
                  <div className="flex items-center gap-2 text-gray-800">
                    <div className="w-1.5 h-1.5 rounded-full bg-orange-400 shrink-0"></div>
                    {item}
                  </div>
                  <Badge className="bg-orange-100 text-orange-700 border-orange-200 font-normal shrink-0">待核实</Badge>
                </li>
              ))}
            </ul>
          </div>
        </section>

        {/* Module: Before & After */}
        {data.rewriteExamples[0] && (
        <section>
          <h3 className="font-bold text-gray-900 mb-4 text-[16px] flex items-center gap-2">
            <Sparkles className="w-4 h-4 text-[#0052D9]" />
            修改前后对比示例
          </h3>
          <div className="flex items-stretch border border-gray-200 rounded-md overflow-hidden bg-gray-50">
            <div className="flex-1 p-4 border-r border-gray-200">
              <div className="text-[12px] font-bold text-gray-500 mb-2">原简历表达</div>
              <div className="text-[13px] text-gray-700">{data.rewriteExamples[0].before}</div>
            </div>
            <div className="flex flex-col items-center justify-center px-4 bg-white">
              <div className="text-[11px] text-[#0052D9] font-medium mb-1 text-center">优化建议</div>
              <ChevronRight className="w-4 h-4 text-gray-300" />
              <div className="text-[10px] text-gray-500 mt-1 text-center w-24">{data.rewriteExamples[0].suggestion}</div>
            </div>
            <div className="flex-[1.5] p-4 bg-blue-50/30 border-l border-blue-100">
              <div className="text-[12px] font-bold text-[#0052D9] mb-2 flex justify-between">
                优化后示例
              </div>
              <div className="text-[13px] text-gray-800 leading-relaxed">
                {data.rewriteExamples[0].after}
              </div>
              <div className="text-[11px] text-amber-600 mt-3 pt-3 border-t border-blue-100/50 flex items-center gap-1">
                <AlertCircle className="w-3.5 h-3.5" />
                {data.rewriteExamples[0].verificationNotice}
              </div>
            </div>
          </div>
        </section>
        )}

        <div className="h-px bg-gray-200 w-full my-6"></div>

        {/* Collapsible Modules */}
        <div className="space-y-4">
          <CollapsibleSection title="展开查看：JD 关键词对齐建议" icon={RefreshCw}>
            <div className="border border-gray-200 rounded-md overflow-hidden">
              <div className="grid grid-cols-12 gap-4 bg-gray-50 p-3 text-[13px] font-medium text-gray-600 border-b border-gray-200">
                <div className="col-span-3">JD 关键词</div>
                <div className="col-span-3">简历当前覆盖</div>
                <div className="col-span-6">建议动作</div>
              </div>
              <div className="divide-y divide-gray-100 text-[13px] leading-relaxed">
                {data.keywordAlignment.map((kw, i) => (
                  <div key={i} className="grid grid-cols-12 gap-4 p-3 items-center hover:bg-gray-50 transition-colors">
                    <div className="col-span-3 font-medium text-gray-800">{kw.keyword}</div>
                    <div className="col-span-3">
                      <Badge className={(kw.currentCoverage === '已覆盖' ? "bg-green-50 text-green-700 border-green-200" : kw.currentCoverage === '弱覆盖' ? "bg-amber-50 text-amber-700 border-amber-200" : "bg-red-50 text-red-600 border-red-200") + " font-normal"}>
                        {kw.currentCoverage}
                      </Badge>
                    </div>
                    <div className="col-span-6 text-gray-600">{kw.suggestedAction}</div>
                  </div>
                ))}
              </div>
            </div>
          </CollapsibleSection>

          <CollapsibleSection title={`展开查看：项目经历详细优化建议 (${data.projectOptimization.length}条)`} icon={FileEdit}>
            <div className="space-y-4">
              {data.projectOptimization.map((proj, i) => (
                <div key={i} className="border border-gray-200 rounded-md p-4">
                  <div className="font-bold text-gray-900 text-[15px] mb-3">{proj.projectName}</div>
                  <div className="space-y-3 text-[13px]">
                    <div className="flex gap-3">
                      <span className="text-gray-500 shrink-0 w-[68px]">当前问题：</span>
                      <span className="text-gray-800">{proj.currentProblem}</span>
                    </div>
                    <div className="flex gap-3">
                      <span className="text-gray-500 shrink-0 w-[68px]">建议改法：</span>
                      <span className="text-[#0052D9] font-medium">{proj.suggestedRewrite}</span>
                    </div>
                    <div className="flex gap-3 bg-gray-50 p-3 rounded border border-gray-100">
                      <span className="text-gray-500 shrink-0 w-[68px]">示例表达：</span>
                      <div className="text-gray-700 leading-relaxed">
                        {proj.exampleText}
                        <div className="text-[11px] text-amber-600 mt-2 flex items-center gap-1">
                          <AlertCircle className="w-3 h-3" />
                          需用户核实后使用
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </CollapsibleSection>
        </div>

        {/* Module: Final Checklist & Actions */}
        <section className="bg-white border border-gray-200 shadow-sm rounded-md p-5 mt-8">
          <div className="mb-5">
            <h3 className="font-bold text-gray-900 mb-3 text-[15px] flex items-center gap-2">
              投递前确认
            </h3>
            <div className="grid grid-cols-2 gap-3 text-[13px] text-gray-700">
              {data.finalChecklist.map((item, i) => (
                <div key={i} className="flex items-center gap-2">
                  <CheckCircle2 className="w-4 h-4 text-gray-300" /> {item}
                </div>
              ))}
            </div>
          </div>
          
          <div className="flex items-center gap-4 pt-5 border-t border-gray-100">
            <Button 
              className="flex-[2] bg-[#0052D9] hover:bg-[#0040A8] text-white h-11 shadow-sm"
            >
              <Sparkles className="w-4 h-4 mr-2" />
              生成优化后简历草稿
            </Button>
            <Button 
              variant="outline" 
              className="flex-1 border-gray-300 text-gray-700 hover:bg-gray-50 h-11"
              onClick={() => navigate(-1)}
            >
              返回岗位匹配分析
            </Button>
            <Button 
              variant="outline" 
              className="flex-1 border-transparent text-[#0052D9] hover:bg-blue-50 h-11"
              onClick={() => navigate(`/jobs/${currentJobId}`)}
            >
              查看岗位详情
            </Button>
          </div>
        </section>
      </div>
    </div>
  );
}
