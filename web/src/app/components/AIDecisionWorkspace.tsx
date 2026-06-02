import React, { useState, useEffect } from "react";
import { Sparkles, FileText, Target, ArrowRight, LogIn, Plus } from "lucide-react";
import { Button } from "./ui/Button";
import { Badge } from "./ui/Badge";
import { useNavigate } from "react-router";
import { loadCurrentResumeSummaryViewModel, loadSelectedJobDetailViewModel } from "../../api";
import type { ResumeSummaryViewModel } from "../../adapters";
import { SelectResumeModal } from "./SelectResumeModal";
import { getAuthToken } from "../../api/authStorage";

interface AIDecisionWorkspaceProps {
  selectedJobId?: string;
}

export function AIDecisionWorkspace({ selectedJobId }: AIDecisionWorkspaceProps) {
  const navigate = useNavigate();
  const [isResumeModalOpen, setIsResumeModalOpen] = useState(false);
  const [resumeVm, setResumeVm] = useState<ResumeSummaryViewModel | null>(null);
  const [jobTitle, setJobTitle] = useState("");
  const [loading, setLoading] = useState(true);

  const isLoggedIn = !!getAuthToken();

  useEffect(() => {
    if (!isLoggedIn) { setLoading(false); return; }
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
  }, [selectedJobId, isLoggedIn]);

  const handleConfirmRecommend = (resumeId: string) => {
    navigate(`/recommend?resumeId=${encodeURIComponent(resumeId)}`);
  };
  
  return (
    <aside className="w-[340px] shrink-0 sticky top-24 self-start flex flex-col gap-4">
      {/* Header */}
      <div className="mb-2">
        <div className="flex items-center gap-2 mb-1">
          <Sparkles className="h-5 w-5 text-[#0052D9]" />
          <h2 className="text-xl font-bold text-gray-900">AI 求职决策</h2>
        </div>
        <p className="text-sm text-gray-500">
          基于你的简历，帮你筛选并优先关注值得投递的岗位。
        </p>
      </div>

      {/* ── 状态 A：未登录 ── */}
      {!isLoggedIn && (
        <div className="bg-white border border-gray-200 rounded-lg p-5 text-center">
          <LogIn className="w-8 h-8 text-gray-300 mx-auto mb-3" />
          <p className="text-sm text-gray-600 mb-1">登录后使用 AI 求职决策</p>
          <p className="text-xs text-gray-400 mb-4">Offer 捕手会基于你的简历推荐更适合的岗位，并分析匹配度。</p>
          <div className="flex flex-col gap-2">
            <Button className="w-full" onClick={() => navigate("/login")}>
              登录
            </Button>
            <Button variant="outline" className="w-full" onClick={() => navigate("/register")}>
              注册新账号
            </Button>
          </div>
        </div>
      )}

      {/* ── 加载中（已登录） ── */}
      {isLoggedIn && loading && (
        <div className="bg-white border border-gray-200 rounded-lg p-4 text-center text-sm text-gray-400">
          正在加载…
        </div>
      )}

      {/* ── 状态 B：已登录但无简历 ── */}
      {isLoggedIn && !loading && !resumeVm && (
        <div className="bg-white border border-gray-200 rounded-lg p-5 text-center">
          <FileText className="w-8 h-8 text-gray-300 mx-auto mb-3" />
          <p className="text-sm text-gray-600 mb-1">你还没有可用于分析的简历</p>
          <p className="text-xs text-gray-400 mb-4">先创建一份简历，才能进行岗位推荐、匹配分析和优化建议。</p>
          <Button className="w-full" onClick={() => navigate("/resumes/new")}>
            <Plus className="w-4 h-4 mr-1.5" />
            创建简历
          </Button>
        </div>
      )}

      {/* ── 状态 C：已登录且有简历 ── */}
      {isLoggedIn && !loading && resumeVm && (
        <>
          {/* Current Resume Card */}
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

          {/* Main Action Card */}
          <div className="bg-white border border-blue-100 rounded-lg p-5 text-center">
            <h3 className="font-bold text-gray-900 mb-1">基于简历找岗位</h3>
            <p className="text-xs text-gray-500 mb-4 leading-relaxed">
              从当前岗位池中推荐最值得优先关注的 5 个岗位
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
              <p className="text-xs text-gray-500 mb-3">已选中岗位，可进入详情页查看 JD。</p>
              <div className="flex gap-2">
                <Button size="sm" className="flex-1" onClick={() => navigate(`/jobs/${selectedJobId}`)}>
                  查看详情
                </Button>
                <Button size="sm" variant="outline" className="flex-1" onClick={() => navigate(`/match?jobId=${encodeURIComponent(selectedJobId || "")}&resumeId=${encodeURIComponent(resumeVm.id)}`)}>
                  分析匹配度
                </Button>
              </div>
            </div>
          )}
        </>
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
