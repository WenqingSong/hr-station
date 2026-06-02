import React, { useState, useEffect } from "react";
import { Sparkles, FileText, LogIn, Plus, ArrowRight } from "lucide-react";
import { Button } from "./ui/Button";
import { Badge } from "./ui/Badge";
import { useNavigate } from "react-router";
import { loadCurrentResumeSummaryViewModel } from "../../api";
import type { ResumeSummaryViewModel } from "../../adapters";
import { getAuthToken } from "../../api/authStorage";

export function AIDecisionWorkspace() {
  const navigate = useNavigate();
  const [resumeVm, setResumeVm] = useState<ResumeSummaryViewModel | null>(null);
  const [loading, setLoading] = useState(true);

  const isLoggedIn = !!getAuthToken();

  useEffect(() => {
    if (!isLoggedIn) { setLoading(false); return; }
    loadCurrentResumeSummaryViewModel()
      .then(setResumeVm)
      .catch(() => setResumeVm(null))
      .finally(() => setLoading(false));
  }, [isLoggedIn]);
  
  return (
    <aside className="w-[340px] shrink-0 flex flex-col gap-4">
      {/* Header */}
      <div className="mb-2">
        <div className="flex items-center gap-2 mb-1">
          <Sparkles className="h-5 w-5 text-[#0052D9]" />
          <h2 className="text-xl font-bold text-gray-900">AI 求职决策</h2>
        </div>
        <p className="text-sm text-gray-500">
          基于你的简历，帮你筛选更适合的岗位，并分析匹配度。
        </p>
      </div>

      {/* ── 状态 A：未登录 ── */}
      {!isLoggedIn && (
        <div className="bg-white border border-gray-200 rounded-lg p-5 text-center">
          <LogIn className="w-8 h-8 text-gray-300 mx-auto mb-3" />
          <p className="text-sm text-gray-600 mb-1">登录后使用 AI 求职决策</p>
          <p className="text-xs text-gray-400 mb-4">Offer 捕手会基于你的简历，帮你筛选更适合的岗位，并分析匹配度。</p>
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
          <p className="text-sm text-gray-600 mb-1">完善简历后开始智能推荐</p>
          <p className="text-xs text-gray-400 mb-4">系统会基于你的简历版本，推荐岗位并分析岗位匹配度。</p>
          <Button className="w-full" onClick={() => navigate("/resumes/new")}>
            <Plus className="w-4 h-4 mr-1.5" />
            去创建简历
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
              {resumeVm.projectCount > 0
                ? `包含 ${resumeVm.projectCount} 个项目经历`
                : "暂未填写项目经历"}
            </p>
          </div>

          {/* Next Steps */}
          <div className="bg-white border border-blue-100 rounded-lg p-5">
            <h3 className="font-bold text-gray-900 mb-3">下一步</h3>
            <div className="flex flex-col gap-2">
              <Button
                className="w-full shadow-sm group"
                onClick={() => navigate("/recommend")}
              >
                推荐适合的岗位
                <ArrowRight className="ml-2 h-4 w-4 group-hover:translate-x-1 transition-transform" />
              </Button>
            </div>
          </div>
        </>
      )}
    </aside>
  );
}
