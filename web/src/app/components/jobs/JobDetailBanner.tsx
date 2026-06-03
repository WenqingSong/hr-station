import React, { useState, useEffect } from "react";
import { ArrowLeft, MapPin, Building2 } from "lucide-react";
import { useNavigate } from "react-router";
import { Badge } from "../ui/Badge";
import { loadSelectedJobDetailViewModel } from "../../../api";
import type { JobDetailViewModel } from "../../../adapters";

interface JobDetailBannerProps {
  jobId?: string;
}

/**
 * K8-5-1：重构岗位详情页 Hero 区域，对齐腾讯招聘详情页
 * 
 * 设计说明：
 * - 蓝色系渐变背景，保留层次感
 * - 信息更克制，只展示关键信息
 * - 标题、标签、slogan、地点/部门简要信息
 * - 与下方主体内容对齐
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

  // 格式化地点和部门信息
  const locationsText = jobVm.locations.length > 0 ? jobVm.locations.join("、") : "";
  const organizationsText = jobVm.organizations.length > 0 
    ? jobVm.organizations.map(o => o.code || o.name).join("、") 
    : "";

  return (
    <div className="w-full bg-[#0052D9] relative overflow-hidden text-white">
      {/* 背景渐变 */}
      <div className="absolute inset-0 bg-gradient-to-br from-[#0052D9] via-[#0066FF] to-[#0040A8]"></div>
      
      {/* 装饰光斑 */}
      <div className="absolute -top-20 -right-20 w-[500px] h-[500px] rounded-full bg-white/5 blur-3xl pointer-events-none"></div>
      <div className="absolute -bottom-32 -left-32 w-[400px] h-[400px] rounded-full bg-[#00B0F0]/10 blur-3xl pointer-events-none"></div>

      {/* 内容区 */}
      <div className="mx-auto max-w-[1280px] px-6 py-6 relative z-10">
        {/* 返回按钮 */}
        <button 
          onClick={() => navigate("/jobs")} 
          className="flex items-center gap-1.5 text-blue-100 hover:text-white transition-colors mb-5 text-sm font-medium"
        >
          <ArrowLeft className="w-4 h-4" /> 返回岗位列表
        </button>

        {/* 岗位标题 */}
        <h1 className="text-[28px] md:text-[32px] font-bold mb-3 tracking-tight">{jobVm.title}</h1>

        {/* 标签 */}
        <div className="flex items-center gap-2 mb-3 flex-wrap">
          {jobVm.applicationType && (
            <Badge className="bg-white/15 text-white border-white/10 px-2.5 py-0.5 rounded-sm font-medium">
              {jobVm.applicationType}
            </Badge>
          )}
          {jobVm.typeLabel && (
            <Badge className="bg-white/10 text-white border-transparent px-2.5 py-0.5 rounded-sm">
              {jobVm.typeLabel}
            </Badge>
          )}
          {jobVm.direction && (
            <Badge className="bg-white/10 text-white border-transparent px-2.5 py-0.5 rounded-sm">
              {jobVm.direction}
            </Badge>
          )}
          {jobVm.category && (
            <Badge className="bg-white/10 text-blue-100 border-transparent px-2.5 py-0.5 rounded-sm">
              {jobVm.category}
            </Badge>
          )}
        </div>

        {/* Slogan / 岗位简述 */}
        {jobVm.slogan && (
          <p className="text-blue-100/90 text-[14px] mb-4 max-w-2xl leading-relaxed">
            {jobVm.slogan}
          </p>
        )}

        {/* 地点和部门简要信息 */}
        <div className="flex flex-col sm:flex-row sm:items-center gap-3 sm:gap-5 text-[13px] text-blue-100/70">
          {locationsText && (
            <div className="flex items-center gap-2">
              <MapPin className="w-4 h-4 opacity-70 shrink-0" />
              <span>工作地点：{locationsText}</span>
            </div>
          )}
          {organizationsText && (
            <div className="flex items-center gap-2">
              <Building2 className="w-4 h-4 opacity-70 shrink-0" />
              <span>招聘部门：{organizationsText}</span>
            </div>
          )}
        </div>
      </div>

      {/* 底部微光边缘 */}
      <div className="absolute bottom-0 left-0 right-0 h-px bg-gradient-to-r from-transparent via-white/20 to-transparent"></div>
    </div>
  );
}
