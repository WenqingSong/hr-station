import React from "react";
import { MapPin, Target, LayoutList, Tags } from "lucide-react";
import { Badge } from "./ui/Badge";

export function ResumeOptimizationSidebarLeft() {
  return (
    <aside className="w-[320px] shrink-0 flex flex-col gap-4">
      <div className="bg-white border border-gray-200 rounded-lg p-5 shadow-sm">
        <div className="flex items-center gap-2 text-gray-500 text-[13px] mb-3">
          <Target className="h-4 w-4" />
          目标岗位 JD 摘要
        </div>
        
        <h2 className="font-bold text-gray-900 text-[18px] mb-3">软件开发-后台开发方向</h2>
        
        <div className="flex flex-wrap gap-1.5 mb-4">
          <Badge variant="secondary" className="text-[11px] px-1.5 font-normal">技术</Badge>
          <Badge variant="secondary" className="text-[11px] px-1.5 font-normal">应届毕业生</Badge>
          <Badge variant="secondary" className="text-[11px] px-1.5 font-normal">后台开发</Badge>
          <Badge variant="secondary" className="text-[11px] px-1.5 font-normal">2026 校园招聘</Badge>
        </div>

        <div className="flex items-center text-[13px] text-gray-600 mb-5">
          <MapPin className="w-3.5 h-3.5 mr-1.5 text-gray-400" />
          深圳 / 北京 / 上海 / 广州 / 成都
        </div>

        <div className="pt-4 border-t border-gray-100 mb-5">
          <h3 className="text-[13px] font-bold text-gray-900 mb-3 flex items-center gap-1.5">
            <LayoutList className="w-3.5 h-3.5 text-gray-400" />
            核心要求摘要
          </h3>
          <ul className="space-y-2.5 text-[12px] text-gray-600 list-disc pl-4 marker:text-gray-300 leading-relaxed">
            <li>熟悉 Go / Java / C++ 至少一门开发语言</li>
            <li>理解数据库、缓存、网络、操作系统基础</li>
            <li>具备后台服务设计、接口开发和问题定位能力</li>
            <li>有项目实践、微服务或分布式系统经验优先</li>
            <li>关注工程质量、系统稳定性和可维护性</li>
          </ul>
        </div>

        <div className="pt-4 border-t border-gray-100">
          <h3 className="text-[13px] font-bold text-gray-900 mb-3 flex items-center gap-1.5">
            <Tags className="w-3.5 h-3.5 text-gray-400" />
            JD 关键词
          </h3>
          <div className="flex flex-wrap gap-1.5">
            {["Go", "MySQL", "Redis", "微服务", "高并发", "Linux", "数据库", "接口开发"].map((kw) => (
              <span key={kw} className="text-[11px] px-2 py-0.5 bg-gray-50 border border-gray-100 text-gray-600 rounded-sm">
                {kw}
              </span>
            ))}
          </div>
        </div>
      </div>
    </aside>
  );
}