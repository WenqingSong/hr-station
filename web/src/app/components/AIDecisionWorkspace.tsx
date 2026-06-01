import React, { useState, useEffect } from "react";
import { Sparkles, FileText, Target, ArrowRight } from "lucide-react";
import { Button } from "./ui/Button";
import { Badge } from "./ui/Badge";
import { useNavigate } from "react-router";
import { loadCurrentResumeSummaryViewModel, loadSelectedJobDetailViewModel } from "../../api";
import type { ResumeSummaryViewModel } from "../../adapters";
import { SelectResumeModal } from "./SelectResumeModal";

interface AIDecisionWorkspaceProps {
  selectedJobId?: string;
}

export function AIDecisionWorkspace({ selectedJobId }: AIDecisionWorkspaceProps) {
  const navigate = useNavigate();
  const [isResumeModalOpen, setIsResumeModalOpen] = useState(false);
  const [resumeVm, setResumeVm] = useState<ResumeSummaryViewModel | null>(null);
  const [jobTitle, setJobTitle] = useState("");
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const promises: Promise<unknown>[] = [
      loadCurrentResumeSummaryViewModel().then(setResumeVm).catch(() => {}),
    ];
    if (selectedJobId) {
      promises.push(
        loadSelectedJobDetailViewModel(selectedJobId).then((j) => setJobTitle(j.title)).catch(() => setJobTitle(""))
      );
    } else {
      promises.push(Promise.resolve(setJobTitle("")));
    }
    Promise.all(promises).finally(() => setLoading(false));
  }, [selectedJobId]);

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
          正在加载…
        </div>
      )}

      {/* Current Resume Card */}
      {resumeVm && (
        <div className="bg-white border border-gray-200 rounded-lg p-4">
          <div className="flex justify-between items-start mb-3">
            <div className="flex items-center gap-2 text-gray-900 font-medium">
              <FileText className="h-4 w-4 text-[#0052D9]" />
              当前简历
            </div>
            <button className="text-xs text-[#0052D9] hover:underline" onClick={() => navigate("/resumes")}>
              管理简历
            </button>
          </div>
          <div className="mb-3">
            <h3 className="font-bold text-gray-900 text-sm mb-2">{resumeVm.name}</h3>
            <div className="flex flex-wrap gap-1.5">
              <Badge variant="secondary" className="text-[10px] px-1.5">{resumeVm.directionLabel}</Badge>
              {resumeVm.skillTags.slice(0, 4).map((skill) => (
                <Badge key={skill} variant="secondary" className="text-[10px] px-1.5">{skill}</Badge>
              ))}
            </div>
          </div>
          <p className="text-xs text-gray-400">
            包含 {resumeVm.projectCount} 个项目经历
          </p>
        </div>
      )}

      {!loading && !resumeVm && (
        <div className="bg-white border border-gray-200 rounded-lg p-4 text-center">
          <p className="text-sm text-gray-400 mb-3">暂无可用的简历</p>
          <Button variant="outline" size="sm" onClick={() => navigate("/resumes/new")}>去创建简历</Button>
        </div>
      )}

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
      {jobTitle && (
        <div className="bg-white border border-[#0052D9] rounded-lg p-4 relative overflow-hidden">
          <div className="absolute top-0 left-0 w-1 h-full bg-[#0052D9]"></div>
          <div className="flex items-center gap-2 text-gray-900 font-medium mb-2">
            <Target className="h-4 w-4 text-[#0052D9]" />
            当前选择岗位
          </div>
          <h3 className="font-bold text-[#0052D9] text-sm mb-1">{jobTitle}</h3>
          <p className="text-xs text-gray-500">已选中岗位，可进入详情页查看 JD。</p>
        </div>
      )}

      <SelectResumeModal
        isOpen={isResumeModalOpen}
        onClose={() => setIsResumeModalOpen(false)}
        onConfirm={handleConfirmRecommend}
        scene="recommendation"
      />
    </aside>
  );
}
