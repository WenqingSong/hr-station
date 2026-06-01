import React, { useState, useEffect } from "react";
import { Sparkles, FileText, Info, AlertCircle, Target, ArrowRight } from "lucide-react";
import { Button } from "./ui/Button";
import { Badge } from "./ui/Badge";
import { useNavigate } from "react-router";
import { loadCurrentResumeSummaryViewModel, loadSelectedJobDetailViewModel } from "../../api";
import type { ResumeSummaryViewModel, JobDetailViewModel } from "../../adapters";
import { SelectResumeModal } from "./SelectResumeModal";

export function AIDecisionWorkspace() {
  const navigate = useNavigate();
  const [isResumeModalOpen, setIsResumeModalOpen] = useState(false);
  const [resumeVm, setResumeVm] = useState<ResumeSummaryViewModel | null>(null);
  const [jobTitle, setJobTitle] = useState("");
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    Promise.all([
      loadCurrentResumeSummaryViewModel().then(setResumeVm).catch(() => {}),
      loadSelectedJobDetailViewModel().then((j) => setJobTitle(j.title)).catch(() => setJobTitle("演示岗位")),
    ]).finally(() => setLoading(false));
  }, []);

  const handleConfirmRecommend = (resumeId: string) => {
    navigate(`/recommend?resumeId=${encodeURIComponent(resumeId)}`);
  };
  
  return (
    <aside className="w-[340px] shrink-0 sticky top-24 self-start flex flex-col gap-4">
      {/* Header */}
      <div className="mb-2">
        <div className="flex items-center gap-2 mb-1">
          <Sparkles className="h-5 w-5 text-[#0052D9]" />
          <h2 className="text-xl font-bold text-gray-900">AI 投递决策</h2>
        </div>
        <p className="text-sm text-gray-500">
          从当前岗位池中，帮你优先找到更适合的岗位。
        </p>
      </div>

      {loading && (
        <div className="bg-white border border-gray-200 rounded-lg p-4 text-center text-sm text-gray-400">
          正在加载岗位数据…
        </div>
      )}
      {!loading && !resumeVm && (
        <div className="bg-amber-50 border border-amber-200 rounded-lg p-3 text-xs text-amber-700">
          简历接口暂时不可用，已展示演示数据。
        </div>
      )}

      {/* Current Resume Card */}
      <div className="bg-white border border-gray-200 rounded-lg p-4">
        <div className="flex justify-between items-start mb-3">
          <div className="flex items-center gap-2 text-gray-900 font-medium">
            <FileText className="h-4 w-4 text-[#0052D9]" />
            当前简历
          </div>
          <button className="text-xs text-[#0052D9] hover:underline">
            切换简历
          </button>
        </div>
        <div className="mb-3">
          <h3 className="font-bold text-gray-900 text-sm mb-2">{resumeVm?.name || ""}</h3>
          <div className="flex flex-wrap gap-1.5">
            <Badge variant="secondary" className="text-[10px] px-1.5">{resumeVm?.directionLabel || ""}</Badge>
            {resumeVm?.skillTags.slice(0, 4).map((skill) => (
              <Badge key={skill} variant="secondary" className="text-[10px] px-1.5">{skill}</Badge>
            ))}
          </div>
        </div>
        <p className="text-xs text-gray-400">
          包含 {resumeVm?.projectCount ?? 0} 个项目经历，最近更新：{resumeVm?.updatedAt || ""}
        </p>
      </div>

      {/* Main Action Card */}
      <div className="bg-gradient-to-b from-blue-50 to-white border border-blue-100 rounded-lg p-5 text-center">
        <h3 className="font-bold text-gray-900 mb-2">还不知道优先看哪些岗位？</h3>
        <p className="text-sm text-gray-600 mb-5 leading-relaxed">
          基于当前简历，从当前岗位池中推荐最值得优先关注的 5 个岗位。
        </p>
        <Button 
          className="w-full shadow-sm group"
          onClick={() => setIsResumeModalOpen(true)}
        >
          推荐最适合的 5 个岗位
          <ArrowRight className="ml-2 h-4 w-4 group-hover:translate-x-1 transition-transform" />
        </Button>
      </div>

      {/* Currently Selected Job */}
      <div className="bg-white border border-[#0052D9] rounded-lg p-4 relative overflow-hidden">
        <div className="absolute top-0 left-0 w-1 h-full bg-[#0052D9]"></div>
        <div className="flex items-center gap-2 text-gray-900 font-medium mb-2">
          <Target className="h-4 w-4 text-[#0052D9]" />
          当前选择岗位
        </div>
        <h3 className="font-bold text-[#0052D9] text-sm mb-1">{jobTitle}</h3>
        <p className="text-xs text-gray-500">
          已选中岗位，可进入详情页查看 JD。
        </p>
      </div>

      {/* Top Fix Preview */}
      <div className="bg-orange-50 border border-orange-100 rounded-lg p-4">
        <div className="flex items-center gap-2 text-orange-800 font-medium mb-3 text-sm">
          <AlertCircle className="h-4 w-4" />
          当前简历可能影响推荐的 3 个因素
        </div>
        <ul className="text-xs text-orange-700 space-y-2 list-disc pl-4 marker:text-orange-400">
          <li>项目经历中云原生证据偏少</li>
          <li>Go 后端经验较强</li>
          <li>缺少高并发压测结果描述</li>
        </ul>
      </div>

      {/* AI Decision Tip Card */}
      <div className="flex gap-2 items-start bg-gray-50 p-3 rounded-md text-xs text-gray-500">
        <Info className="h-4 w-4 shrink-0 text-gray-400 mt-0.5" />
        <p>
          投递前，建议先完成岗位推荐，缩小目标范围，再进入单岗位匹配分析。
        </p>
      </div>

      {/* Status Badge */}
      <div className="flex justify-center">
        <div className="inline-flex items-center gap-1.5 px-2 py-1 rounded-full text-gray-300 text-[10px] opacity-70">
          <Sparkles className="h-3 w-3" />
          Mock 演示状态
        </div>
      </div>

      <SelectResumeModal
        isOpen={isResumeModalOpen}
        onClose={() => setIsResumeModalOpen(false)}
        onConfirm={handleConfirmRecommend}
        scene="recommendation"
      />
    </aside>
  );
}
