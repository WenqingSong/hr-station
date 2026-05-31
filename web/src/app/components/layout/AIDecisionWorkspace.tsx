// @deprecated — 5-3.1 已废弃。AI 决策工作台使用 ../AIDecisionWorkspace.tsx。
// 保留此文件仅用于历史追溯，后续清理阶段可安全删除。

import { BrainCircuit, ChevronRight, FileText, Briefcase } from "lucide-react";
import { Button } from "../ui/Button";

interface AIDecisionWorkspaceProps {
  currentResume?: string;
  currentJob?: string;
  children?: React.ReactNode;
}

export function AIDecisionWorkspace({
  currentResume = "V2 后端开发简历",
  currentJob = "软件开发-后台开发方向",
  children,
}: AIDecisionWorkspaceProps) {
  return (
    <aside className="w-[320px] shrink-0 sticky top-20 flex flex-col gap-4">
      <div className="bg-white rounded-lg shadow-sm border border-slate-200 overflow-hidden flex flex-col max-h-[calc(100vh-120px)]">
        {/* Header */}
        <div className="px-4 py-3 bg-gradient-to-r from-blue-50 to-white border-b border-slate-100 flex items-center gap-2">
          <BrainCircuit className="w-5 h-5 text-blue-600" />
          <h2 className="font-semibold text-slate-800 text-sm">AI 投递决策</h2>
        </div>

        {/* Context Info */}
        <div className="px-4 py-3 border-b border-slate-100 bg-slate-50/50 space-y-3 shrink-0">
          <div className="flex items-start gap-2">
            <Briefcase className="w-4 h-4 text-slate-400 mt-0.5 shrink-0" />
            <div>
              <div className="text-xs text-slate-500 mb-0.5">当前岗位</div>
              <div className="text-sm font-medium text-slate-800 line-clamp-1">
                {currentJob || <span className="text-slate-400 font-normal">未选择</span>}
              </div>
            </div>
          </div>
          <div className="flex items-start gap-2">
            <FileText className="w-4 h-4 text-slate-400 mt-0.5 shrink-0" />
            <div>
              <div className="text-xs text-slate-500 mb-0.5">当前简历</div>
              <div className="flex items-center gap-1 cursor-pointer hover:text-blue-600 transition-colors">
                <div className="text-sm font-medium text-slate-800">
                  {currentResume || <span className="text-slate-400 font-normal">未选择</span>}
                </div>
                <ChevronRight className="w-3 h-3 text-slate-400" />
              </div>
            </div>
          </div>
        </div>

        {/* Dynamic Content Area (Scrollable) */}
        <div className="flex-1 overflow-y-auto p-4 space-y-5">
          {children}
        </div>
      </div>
    </aside>
  );
}
