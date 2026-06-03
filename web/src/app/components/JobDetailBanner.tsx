import React, { useState, useEffect } from "react";
import { ArrowLeft, MapPin, Building2 } from "lucide-react";
import { useNavigate } from "react-router";
import { Badge } from "./ui/Badge";
import { loadSelectedJobDetailViewModel } from "../../api";
import type { JobDetailViewModel } from "../../adapters";

interface JobDetailBannerProps {
  jobId?: string;
}

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
      <div className="absolute inset-0 bg-gradient-to-br from-[#0052D9] via-[#0066FF] to-[#0040A8]"></div>
      {/* 装饰光斑 */}
      <div className="absolute -top-20 -right-20 w-[500px] h-[500px] rounded-full bg-white/5 blur-3xl pointer-events-none"></div>
      <div className="absolute -bottom-32 -left-32 w-[400px] h-[400px] rounded-full bg-[#00B0F0]/10 blur-3xl pointer-events-none"></div>

      {/* 内容区 - 与主体内容对齐 */}
      <div className="relative z-10 mx-auto max-w-[1200px] px-6 py-8 md:py-10">
        {/* 返回按钮 */}
        <button
          onClick={() => navigate("/jobs")}
          className="inline-flex items-center gap-1.5 text-blue-100 hover:text-white transition-colors mb-6 text-sm font-medium"
        >
          <ArrowLeft className="w-4 h-4" />
          返回岗位列表
        </button>

        {/* 岗位标题 */}
        <h1 className="text-[28px] md:text-[32px] font-bold mb-3 tracking-tight leading-tight">
          {jobVm.title}
        </h1>

        {/* 标签行 */}
        <div className="flex items-center gap-2 mb-4 flex-wrap">
          {jobVm.applicationType && (
            <Badge className="bg-white/15 text-white border-white/10 px-2.5 py-0.5 rounded-sm font-medium text-xs">
              {jobVm.applicationType}
            </Badge>
          )}
          {jobVm.direction && (
            <Badge className="bg-white/10 text-white border-transparent px-2.5 py-0.5 rounded-sm text-xs">
              {jobVm.direction}
            </Badge>
          )}
          {jobVm.category && (
            <Badge className="bg-white/10 text-blue-100 border-transparent px-2.5 py-0.5 rounded-sm text-xs">
              {jobVm.category}
            </Badge>
          )}
        </div>

        {/* 一句话 slogan / 岗位简述 */}
        {jobVm.slogan && (
          <p className="text-blue-100/90 text-[15px] mb-5 max-w-2xl leading-relaxed">
            {jobVm.slogan}
          </p>
        )}

        {/* 工作地点 + 招聘部门 */}
        <div className="flex flex-col sm:flex-row sm:items-center gap-2 sm:gap-6 text-sm text-blue-50/90">
          {jobVm.locations.length > 0 && (
            <div className="flex items-center gap-2">
              <MapPin className="w-4 h-4 opacity-60 shrink-0" />
              <span className="font-medium text-blue-100">工作地点：</span>
              <span>{jobVm.locations.join("、")}</span>
            </div>
          )}
          {jobVm.organizations.length > 0 && (
            <div className="flex items-center gap-2">
              <Building2 className="w-4 h-4 opacity-60 shrink-0" />
              <span className="font-medium text-blue-100">招聘部门：</span>
              <span>{jobVm.organizations.map(o => o.name || o.code).join("、")}</span>
            </div>
          )}
        </div>
      </div>

      {/* 底部微光边缘 */}
      <div className="absolute bottom-0 left-0 right-0 h-px bg-gradient-to-r from-transparent via-white/20 to-transparent" />
    </div>
  );
}
