import React from "react";
import { Activity, History, Lightbulb, Sparkles, Target } from "lucide-react";
import { Badge } from "./ui/Badge";

export function MyResumesSidebar() {
  return (
    <aside className="w-[300px] xl:w-[360px] shrink-0 flex flex-col gap-4">
      {/* Module 1: AI Status */}
      <div className="bg-white border border-gray-200 rounded-lg p-5 shadow-sm">
        <h3 className="font-bold text-gray-900 text-[15px] mb-4 flex items-center gap-2">
          <Activity className="w-4 h-4 text-[#0052D9]" />
          AI 分析使用状态
        </h3>
        
        <div className="space-y-4 text-[13px]">
          <div>
            <div className="text-gray-400 mb-1">当前简历：</div>
            <div className="font-medium text-gray-800">V2 后端开发简历</div>
          </div>
          <div>
            <div className="text-gray-400 mb-2">可用于：</div>
            <div className="flex flex-wrap gap-2">
              <Badge variant="outline" className="text-gray-600 bg-gray-50 border-gray-200 font-normal">岗位推荐</Badge>
              <Badge variant="outline" className="text-gray-600 bg-gray-50 border-gray-200 font-normal">岗位匹配分析</Badge>
              <Badge variant="outline" className="text-gray-600 bg-gray-50 border-gray-200 font-normal">简历优化</Badge>
            </div>
          </div>
          <div className="pt-3 border-t border-gray-100 flex items-center gap-1.5 text-gray-400 text-[12px]">
            <Sparkles className="w-3.5 h-3.5" />
            当前结果来源：Mock 演示
          </div>
        </div>
      </div>

      {/* Module 2: Recent Analysis */}
      <div className="bg-white border border-gray-200 rounded-lg p-5 shadow-sm">
        <h3 className="font-bold text-gray-900 text-[15px] mb-4 flex items-center gap-2">
          <History className="w-4 h-4 text-gray-400" />
          最近分析记录
        </h3>
        
        <div className="space-y-4">
          <div className="border-b border-gray-100 pb-3 last:border-0 last:pb-0">
            <div className="font-medium text-gray-800 text-[13px] mb-1">软件开发-后台开发方向</div>
            <div className="flex items-center gap-3 text-[12px] mb-1.5">
              <span className="text-gray-500">匹配分数：<span className="text-green-600 font-medium">82</span></span>
              <span className="text-gray-400">|</span>
              <span className="text-gray-400">2 小时前</span>
            </div>
            <div className="text-[12px] text-amber-600 flex items-center gap-1.5 bg-amber-50 px-2 py-1 rounded inline-flex">
              <Target className="w-3 h-3" />
              先修复后投递
            </div>
          </div>

          <div className="border-b border-gray-100 pb-3 last:border-0 last:pb-0">
            <div className="font-medium text-gray-800 text-[13px] mb-1">云计算开发实习生</div>
            <div className="flex items-center gap-3 text-[12px] mb-1.5">
              <span className="text-gray-500">匹配分数：<span className="text-[#0052D9] font-medium">76</span></span>
              <span className="text-gray-400">|</span>
              <span className="text-gray-400">昨天</span>
            </div>
            <div className="text-[12px] text-gray-600 flex items-center gap-1.5 bg-gray-50 px-2 py-1 rounded inline-flex">
              建议补充 Kubernetes 证据
            </div>
          </div>

          <div className="border-b border-gray-100 pb-3 last:border-0 last:pb-0">
            <div className="font-medium text-gray-800 text-[13px] mb-1">数据开发实习生</div>
            <div className="flex items-center gap-3 text-[12px] mb-1.5">
              <span className="text-gray-500">匹配分数：<span className="text-gray-600 font-medium">61</span></span>
              <span className="text-gray-400">|</span>
              <span className="text-gray-400">3 天前</span>
            </div>
            <div className="text-[12px] text-gray-500 flex items-center gap-1.5 bg-gray-50 px-2 py-1 rounded inline-flex">
              暂不建议优先投递
            </div>
          </div>
        </div>
      </div>

      {/* Module 3: Maintenance Tips */}
      <div className="bg-white border border-gray-200 rounded-lg p-5 shadow-sm">
        <h3 className="font-bold text-gray-900 text-[15px] mb-3 flex items-center gap-2">
          <Lightbulb className="w-4 h-4 text-amber-500" />
          简历维护建议
        </h3>
        <ul className="space-y-2 text-[12px] text-gray-600">
          <li className="flex items-start gap-2">
            <div className="w-1.5 h-1.5 rounded-full bg-gray-300 shrink-0 mt-1.5"></div>
            保持不同方向简历分开管理
          </li>
          <li className="flex items-start gap-2">
            <div className="w-1.5 h-1.5 rounded-full bg-gray-300 shrink-0 mt-1.5"></div>
            投递前优先使用目标岗位方向简历
          </li>
          <li className="flex items-start gap-2">
            <div className="w-1.5 h-1.5 rounded-full bg-gray-300 shrink-0 mt-1.5"></div>
            AI 建议仅作投递前参考，不替代真实经历核实
          </li>
        </ul>
      </div>
    </aside>
  );
}