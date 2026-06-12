import { useState, useEffect } from "react";
import { Sparkles, FileText, LogIn, Plus, ArrowRight } from "lucide-react";
import { Button } from "../ui/Button";
import { Badge } from "../ui/Badge";
import { useNavigate } from "react-router";
import { loadCurrentResumeSummaryViewModel } from "../../../api";
import type { ResumeSummaryViewModel } from "../../../adapters";
import { useAuth } from "../../../stores/authStore";
import { AuthDialog } from "../auth/AuthDialog";

export function AIDecisionWorkspace() {
  const navigate = useNavigate();
  const { isAuthenticated, setAuthenticated } = useAuth();
  const [resumeVm, setResumeVm] = useState<ResumeSummaryViewModel | null>(null);
  const [loading, setLoading] = useState(true);

  const [authDialogOpen, setAuthDialogOpen] = useState(false);
  const [authDialogMode, setAuthDialogMode] = useState<"login" | "register">("login");

  useEffect(() => {
    if (!isAuthenticated) { setLoading(false); return; }
    loadCurrentResumeSummaryViewModel()
      .then(setResumeVm)
      .catch(() => setResumeVm(null))
      .finally(() => setLoading(false));
  }, [isAuthenticated]);
  
  return (
    <div className="flex flex-col gap-3">
      {/* Header */}
      <div className="mb-1">
        <div className="flex items-center gap-1.5 mb-1">
          <Sparkles className="h-4 w-4 text-[#0052D9]" />
          <h3 className="text-base font-semibold text-gray-800">Offer捕手</h3>
        </div>
        <p className="text-[12px] text-gray-400 leading-relaxed">
          基于你的简历，帮你筛选更适合的岗位并分析匹配度。
        </p>
      </div>

      {/* ── 状态 A：未登录 ── */}
      {!isAuthenticated && (
        <div className="bg-white border border-gray-100 rounded-xl p-4 text-center">
          <LogIn className="w-6 h-6 text-gray-300 mx-auto mb-2.5" />
          <p className="text-[13px] text-gray-600 mb-1">登录后使用 Offer捕手</p>
          <p className="text-[12px] text-gray-400 mb-3 leading-relaxed">Offer捕手会基于你的简历，帮你筛选更适合的岗位。</p>
          <div className="flex flex-col gap-2">
            <Button className="w-full text-[13px]" onClick={() => { setAuthDialogMode("login"); setAuthDialogOpen(true); }}>
              登录
            </Button>
            <Button variant="outline" className="w-full text-[13px]" onClick={() => { setAuthDialogMode("register"); setAuthDialogOpen(true); }}>
              注册新账号
            </Button>
          </div>
        </div>
      )}

      {/* ── 加载中（已登录） ── */}
      {isAuthenticated && loading && (
        <div className="bg-white border border-gray-100 rounded-xl p-4 text-center text-[13px] text-gray-400">
          正在加载…
        </div>
      )}

      {/* ── 状态 B：已登录但无简历 ── */}
      {isAuthenticated && !loading && !resumeVm && (
        <div className="bg-white border border-gray-100 rounded-xl p-4 text-center">
          <FileText className="w-6 h-6 text-gray-300 mx-auto mb-2.5" />
          <p className="text-[13px] text-gray-600 mb-1">完善简历后开始智能推荐</p>
          <p className="text-[12px] text-gray-400 mb-3 leading-relaxed">系统会基于你的简历版本，推荐岗位并分析岗位匹配度。</p>
          <Button className="w-full text-[13px]" onClick={() => navigate("/resumes/new")}>
            <Plus className="w-3.5 h-3.5 mr-1.5" />
            去创建简历
          </Button>
        </div>
      )}

      {/* ── 状态 C：已登录且有简历 ── */}
      {isAuthenticated && !loading && resumeVm && (
        <>
          {/* Current Resume Card */}
          <div className="bg-white border border-gray-100 rounded-xl p-4">
            <div className="flex justify-between items-start mb-2.5">
              <div className="flex items-center gap-1.5 text-gray-800 font-medium text-[13px]">
                <FileText className="h-3.5 w-3.5 text-[#0052D9]" />
                当前简历
              </div>
              <button className="text-[12px] text-[#0052D9] hover:underline" onClick={() => navigate("/resumes")}>
                管理简历
              </button>
            </div>
            <div className="mb-2.5">
              <h4 className="font-semibold text-gray-900 text-[13px] mb-1.5">{resumeVm.name}</h4>
              <div className="flex flex-wrap gap-1">
                {resumeVm.directionLabel ? (
                  <Badge variant="secondary" className="text-[10px] px-1.5">{resumeVm.directionLabel}</Badge>
                ) : null}
                {(Array.isArray(resumeVm.skillTags) ? resumeVm.skillTags : []).filter(Boolean).slice(0, 3).map((skill) => (
                  <Badge key={skill} variant="secondary" className="text-[10px] px-1.5">{skill}</Badge>
                ))}
              </div>
            </div>
            <p className="text-[11px] text-gray-400">
              {resumeVm.projectCount > 0
                ? `包含 ${resumeVm.projectCount} 个项目经历`
                : "暂未填写项目经历"}
            </p>
          </div>

          {/* Next Steps */}
          <div className="bg-white border border-gray-100 rounded-xl p-4">
            <h4 className="font-semibold text-gray-800 text-[13px] mb-2.5">下一步</h4>
            <Button
              className="w-full text-[13px]"
              onClick={() => navigate(`/recommend?resumeId=${resumeVm.id}`)}
            >
              推荐适合的岗位
              <ArrowRight className="ml-2 h-3.5 w-3.5" />
            </Button>
          </div>
        </>
      )}
      <AuthDialog
        open={authDialogOpen}
        onOpenChange={setAuthDialogOpen}
        defaultMode={authDialogMode}
      />
    </div>
  );
}
