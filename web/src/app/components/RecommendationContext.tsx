import React from "react";
import { FileText, Sparkles, AlertCircle } from "lucide-react";
import { Badge } from "./ui/Badge";

export function RecommendationContext() {
  return (
    <aside className="w-[320px] shrink-0 flex flex-col gap-4">
      {/* Resume Card */}
      <div className="bg-white border border-gray-200 rounded-lg p-4 shadow-sm">
        <div className="flex justify-between items-start mb-3">
          <div className="flex items-center gap-2 text-gray-500 text-[13px]">
            <FileText className="h-4 w-4" />
            当前简历
          </div>
          <button className="text-[12px] text-[#0052D9] hover:underline">
            切换简历
          </button>
        </div>
        <div className="mb-3">
          <h3 className="font-bold text-gray-900 text-[15px] mb-2">V2 后端开发简历</h3>
          <div className="flex flex-wrap gap-1.5">
            {["后端开发", "Go", "MySQL", "Redis", "微服务"].map((tag) => (
              <Badge key={tag} variant="outline" className="text-[11px] px-1.5 text-gray-600 bg-gray-50 font-normal">
                {tag}
              </Badge>
            ))}
          </div>
        </div>
        <p className="text-[12px] text-gray-400">
          包含 3 个项目经历，最近更新：2 小时前
        </p>
      </div>

      {/* Basis Card */}
      <div className="bg-white border border-gray-200 rounded-lg p-5 shadow-sm">
        <h3 className="font-bold text-gray-900 mb-4 text-[15px]">本次推荐依据</h3>
        <ul className="space-y-3 text-[13px] text-gray-600">
          <li className="flex items-start">
            <span className="w-[70px] shrink-0 text-gray-400">求职方向：</span>
            <span className="font-medium text-gray-800">后端开发</span>
          </li>
          <li className="flex items-start">
            <span className="w-[70px] shrink-0 text-gray-400">核心技能：</span>
            <span className="font-medium text-gray-800">Go / MySQL / Redis</span>
          </li>
          <li className="flex items-start">
            <span className="w-[70px] shrink-0 text-gray-400">项目经历：</span>
            <span className="font-medium text-gray-800 leading-relaxed">微服务、接口开发、数据库设计</span>
          </li>
          <li className="flex items-start pt-2 border-t border-gray-50">
            <span className="w-[70px] shrink-0 text-gray-400">目标岗位池：</span>
            <span className="font-medium text-gray-800">软件开发方向</span>
          </li>
        </ul>
      </div>

      {/* Top Fix Preview */}
      <div className="bg-amber-50/30 border border-amber-100/60 rounded-lg p-5 shadow-sm">
        <div className="flex items-center gap-2 mb-3">
          <AlertCircle className="h-4 w-4 text-amber-500" />
          <h3 className="font-bold text-amber-900 text-[14px]">影响推荐排序的 3 个因素</h3>
        </div>
        <ul className="space-y-2 text-[12px] text-amber-800/80 list-disc pl-5 marker:text-amber-300">
          <li>云原生项目证据偏少</li>
          <li>高并发压测结果缺失</li>
          <li>Linux 使用经验表达不足</li>
        </ul>
      </div>

      {/* Mock Status */}
      <div className="mt-2 flex justify-center">
        <div className="inline-flex items-center gap-1.5 px-3 py-1.5 text-gray-400 text-[11px] opacity-70">
          当前结果来源：Mock 演示
        </div>
      </div>
    </aside>
  );
}
