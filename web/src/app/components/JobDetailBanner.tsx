import React, { useState, useEffect } from "react";
import { ArrowLeft, MapPin, Building2, Tag } from "lucide-react";
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
      <div className="absolute inset-0 bg-gradient-to-br from-[#0052D9] via-[#0040A8] to-[#002B73]"></div>
      <div className="absolute top-0 right-[10%] w-[600px] h-[600px] rounded-full bg-white opacity-[0.03] blur-3xl mix-blend-overlay pointer-events-none"></div>

      <div className="mx-auto max-w-[1200px] px-6 py-6 relative z-10 flex flex-col items-start">
        <button onClick={() => navigate("/jobs")} className="flex items-center gap-1.5 text-blue-100 hover:text-white transition-colors mb-4 text-sm font-medium">
          <ArrowLeft className="w-4 h-4" /> 返回岗位列表
        </button>
        
        <h1 className="text-[28px] font-bold mb-2 tracking-tight">{jobVm.title}</h1>

        {jobVm.slogan && (
          <p className="text-blue-100/80 italic text-sm mb-3 max-w-2xl">{jobVm.slogan}</p>
        )}
        
        <div className="flex items-center gap-2 mb-3 flex-wrap">
          {jobVm.applicationType && <Badge className="bg-white/15 text-white border-white/10 px-2.5 py-0.5 rounded-sm font-medium">{jobVm.applicationType}</Badge>}
          {jobVm.typeLabel && <Badge className="bg-white/10 text-white border-transparent px-2.5 py-0.5 rounded-sm">{jobVm.typeLabel}</Badge>}
          {jobVm.direction && <Badge className="bg-white/10 text-white border-transparent px-2.5 py-0.5 rounded-sm">{jobVm.direction}</Badge>}
          {jobVm.category && <Badge className="bg-white/10 text-blue-100 border-transparent px-2.5 py-0.5 rounded-sm">{jobVm.category}</Badge>}
          {jobVm.statusLabel && <Badge className="bg-green-500/20 text-green-200 border-green-400/20 px-2.5 py-0.5 rounded-sm">{jobVm.statusLabel}</Badge>}
        </div>

        <div className="flex flex-col gap-2 text-sm text-blue-50/90 max-w-3xl">
          {jobVm.locations.length > 0 && (
            <div className="flex items-start gap-2">
              <MapPin className="w-4 h-4 mt-0.5 opacity-50 shrink-0" />
              <span>{jobVm.locations.join("、")}</span>
            </div>
          )}
          {jobVm.organizations.length > 0 && (
            <div className="flex items-start gap-2">
              <Building2 className="w-4 h-4 mt-0.5 opacity-50 shrink-0" />
              <span>{jobVm.organizations.map(o => o.name || o.code).join("、")}</span>
            </div>
          )}
          {jobVm.summary && (
            <div className="flex items-start gap-2 mt-1">
              <Tag className="w-4 h-4 mt-0.5 opacity-50 shrink-0" />
              <p className="leading-relaxed">{jobVm.summary}</p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
