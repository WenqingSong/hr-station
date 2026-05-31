import React from "react";
import { ArrowLeft } from "lucide-react";
import { useNavigate } from "react-router";
import { Badge } from "./ui/Badge";

export function JobDetailBanner() {
  const navigate = useNavigate();
  
  return (
    <div className="w-full bg-[#0052D9] relative overflow-hidden text-white">
      {/* Subtle texture/gradient */}
      <div className="absolute inset-0 bg-gradient-to-br from-[#0052D9] via-[#0040A8] to-[#002B73]"></div>
      
      {/* Decorative geometry, subtle */}
      <div className="absolute top-0 right-[10%] w-[600px] h-[600px] rounded-full bg-white opacity-[0.03] blur-3xl mix-blend-overlay pointer-events-none"></div>

      <div className="mx-auto max-w-[1200px] px-6 py-6 relative z-10 flex flex-col items-start">
        <button 
          onClick={() => navigate("/")}
          className="flex items-center gap-1.5 text-blue-100 hover:text-white transition-colors mb-4 text-sm font-medium"
        >
          <ArrowLeft className="w-4 h-4" />
          返回岗位列表
        </button>
        
        <h1 className="text-[28px] font-bold mb-3 tracking-tight">软件开发-后台开发方向</h1>
        
        <div className="flex items-center gap-3 mb-4">
          <Badge className="bg-white/10 text-white border-transparent hover:bg-white/20 font-normal px-2.5 py-0.5 rounded-sm">技术</Badge>
          <Badge className="bg-white/10 text-white border-transparent hover:bg-white/20 font-normal px-2.5 py-0.5 rounded-sm">应届毕业生</Badge>
          <Badge className="bg-white/10 text-white border-transparent hover:bg-white/20 font-normal px-2.5 py-0.5 rounded-sm">2026 校园招聘</Badge>
        </div>

        <div className="flex flex-col gap-2.5 text-sm text-blue-50/90 max-w-3xl">
          <div className="flex items-start gap-3">
            <span className="opacity-60 shrink-0 w-[70px]">事业群：</span>
            <span>CDG / CSIG / TEG / WXG</span>
          </div>
          <div className="flex items-start gap-3">
            <span className="opacity-60 shrink-0 w-[70px]">工作地点：</span>
            <span>深圳 / 北京 / 上海 / 广州 / 成都</span>
          </div>
          <div className="flex items-start gap-3 mt-1.5">
            <span className="opacity-60 shrink-0 w-[70px]">岗位简介：</span>
            <p className="leading-relaxed">参与腾讯产品后台系统的设计、开发与优化，打造稳定、高效、可扩展的后端支撑能力。</p>
          </div>
        </div>
      </div>
    </div>
  );
}
