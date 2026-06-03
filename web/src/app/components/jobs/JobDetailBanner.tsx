import React, { useState, useEffect } from "react";
import { ArrowLeft } from "lucide-react";
import { useNavigate } from "react-router";
import { Badge } from "../ui/Badge";
import { loadSelectedJobDetailViewModel } from "../../../api";
import type { JobDetailViewModel } from "../../../adapters";

interface JobDetailBannerProps {
  jobId?: string;
}

/**
 * K8-5-10-1：Hero 简化——只展示标题 + 核心标签 + slogan
 *
 * 设计说明：
 * - 蓝色系渐变背景
 * - 标题、标签、slogan 为 Hero 核心信息
 * - 完整字段（地点、部门等）移至正文白卡
 * - 返回按钮在 Hero 顶部，不被 Navbar 遮挡
 * - 内容宽度与正文主体对齐（max-w-[1280px]）
 */
export function JobDetailBanner({ jobId }: JobDetailBannerProps) {
  const navigate = useNavigate();
  const [jobVm, setJobVm] = useState<JobDetailViewModel | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!jobId) { setLoading(false); return; }
    loadSelectedJobDetailViewModel(jobId).then(setJobVm).catch(() => {}).finally(() => setLoading(false));
  }, [jobId]);

  if (loading) return <div className="bg-[#0052D9] text-white text-center py-12 text-sm">正在加载岗位详情…</div>;
  if (!jobVm) return <div className="bg-blue-50 text-gray-500 text-center py-12 text-sm">岗位不存在或已下线</div>;

  return (
    <div className="w-full bg-[#0052D9] relative overflow-hidden text-white">
      {/* 背景渐变 */}
      <div className="absolute inset-0 bg-gradient-to-br from-[#0052D9] via-[#0066FF] to-[#0040A8]" />
      
      {/* 装饰光斑 */}
      <div className="absolute -top-20 -right-20 w-[500px] h-[500px] rounded-full bg-white/5 blur-3xl pointer-events-none" />
      <div className="absolute -bottom-32 -left-32 w-[400px] h-[400px] rounded-full bg-[#00B0F0]/10 blur-3xl pointer-events-none" />

      {/* 内容区 */}
      <div className="mx-auto max-w-[1280px] px-6 pt-20 pb-8 relative z-10">
        {/* 返回按钮 — pt-20 确保不被 Navbar 遮挡 */}
        <button 
          onClick={() => navigate("/jobs")} 
          className="flex items-center gap-1.5 text-blue-100 hover:text-white transition-colors mb-5 text-sm font-medium"
        >
          <ArrowLeft className="w-4 h-4" /> 返回岗位列表
        </button>

        {/* 岗位标题 */}
        <h1 className="text-[28px] md:text-[32px] font-bold mb-3 tracking-tight">{jobVm.title}</h1>

        {/* 核心标签 */}
        <div className="flex items-center gap-2 mb-3 flex-wrap">
          {jobVm.typeLabel && (
            <Badge className="bg-white/15 text-white border-white/10 px-2.5 py-0.5 rounded-sm font-medium">
              {jobVm.typeLabel}
            </Badge>
          )}
          {jobVm.applicationType && (
            <Badge className="bg-white/10 text-white border-transparent px-2.5 py-0.5 rounded-sm">
              {jobVm.applicationType}
            </Badge>
          )}
          {jobVm.direction && (
            <Badge className="bg-white/10 text-white border-transparent px-2.5 py-0.5 rounded-sm">
              {jobVm.direction}
            </Badge>
          )}
        </div>

        {/* Slogan */}
        {jobVm.slogan && (
          <p className="text-blue-100/90 text-[14px] max-w-2xl leading-relaxed">
            {jobVm.slogan}
          </p>
        )}
      </div>

      {/* 底部微光边缘 */}
      <div className="absolute bottom-0 left-0 right-0 h-px bg-gradient-to-r from-transparent via-white/20 to-transparent" />
    </div>
  );
}
