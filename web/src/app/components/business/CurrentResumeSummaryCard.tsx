import { FileText, Clock, Hash, Briefcase } from "lucide-react";

export function CurrentResumeSummaryCard() {
  return (
    <div className="bg-white border border-slate-200 rounded-lg p-4">
      <div className="flex items-center gap-3 mb-4">
        <div className="w-10 h-10 bg-blue-50 rounded flex items-center justify-center text-blue-600">
          <FileText className="w-5 h-5" />
        </div>
        <div>
          <div className="text-sm font-medium text-slate-800">V2 后端开发简历</div>
          <div className="text-xs text-slate-500">2026届校园招聘</div>
        </div>
      </div>

      <div className="space-y-3">
        <div className="flex items-start gap-2">
          <Briefcase className="w-4 h-4 text-slate-400 mt-0.5" />
          <div>
            <div className="text-xs text-slate-500 mb-0.5">求职方向</div>
            <div className="text-sm text-slate-700">软件开发-后台开发方向</div>
          </div>
        </div>
        <div className="flex items-start gap-2">
          <Hash className="w-4 h-4 text-slate-400 mt-0.5" />
          <div>
            <div className="text-xs text-slate-500 mb-0.5">核心关键词</div>
            <div className="text-sm text-slate-700">Go, 微服务, 高并发, MySQL, Redis</div>
          </div>
        </div>
        <div className="flex items-center justify-between text-xs text-slate-500 pt-3 border-t border-slate-100">
          <span className="flex items-center gap-1">
            <Clock className="w-3.5 h-3.5" />
            最近分析: 2小时前
          </span>
          <span>包含 3 个项目经历</span>
        </div>
      </div>
    </div>
  );
}
