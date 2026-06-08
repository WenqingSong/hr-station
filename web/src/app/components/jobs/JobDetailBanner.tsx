import React, { useState, useEffect, useCallback } from "react";
import { ArrowLeft } from "lucide-react";
import { useNavigate } from "react-router";
import { Badge } from "../ui/Badge";
import { loadSelectedJobDetailViewModel } from "../../../api";
import type { JobDetailViewModel } from "../../../adapters";
import { PageError } from "../PageState";

interface JobDetailBannerProps {
  jobId?: string;
}

/**
 * K8-5-12-D1：Hero loading/error/retry 修复
 *
 * 设计说明：
 * - 蓝色系渐变背景
 * - 标题、标签、slogan 为 Hero 核心信息
 * - loading: PageLoading；error: PageError(onRetry)；not found: PageError
 */
export function JobDetailBanner({ jobId }: JobDetailBannerProps) {
  const navigate = useNavigate();
  const [jobVm, setJobVm] = useState<JobDetailViewModel | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [retryCount, setRetryCount] = useState(0);

  useEffect(() => {
    if (!jobId) { setLoading(false); return; }
    setLoading(true);
    setError(null);
    loadSelectedJobDetailViewModel(jobId)
      .then((vm) => { setJobVm(vm); setError(null); })
      .catch(() => setError("岗位信息加载失败，请稍后重试。"))
      .finally(() => setLoading(false));
  }, [jobId, retryCount]);

  const handleRetry = useCallback(() => setRetryCount((c) => c + 1), []);

  if (loading) {
    return (
      <div className="w-full bg-[#0052D9] relative overflow-hidden text-white">
        {/* 背景渐变 */}
        <div className="absolute inset-0 bg-gradient-to-br from-[#0052D9] via-[#0066FF] to-[#0040A8]" />
        {/* 装饰光斑 */}
        <div className="absolute -top-20 -right-20 w-[500px] h-[500px] rounded-full bg-white/5 blur-3xl pointer-events-none" />
        <div className="absolute -bottom-32 -left-32 w-[400px] h-[400px] rounded-full bg-[#00B0F0]/10 blur-3xl pointer-events-none" />

        <div className="mx-auto max-w-[1280px] px-6 pt-20 pb-8 relative z-10 animate-pulse">
          {/* 返回按钮 */}
          <div className="flex items-center gap-1.5 mb-5">
            <ArrowLeft className="w-4 h-4 text-blue-100/60" />
            <div className="h-4 w-24 bg-white/10 rounded" />
          </div>

          {/* 岗位标题 */}
          <div className="h-8 w-3/4 bg-white/15 rounded mb-3" />

          {/* 标签 */}
          <div className="flex items-center gap-2 mb-3">
            <div className="h-6 w-12 bg-white/10 rounded-sm" />
            <div className="h-6 w-16 bg-white/10 rounded-sm" />
          </div>

          {/* Slogan */}
          <div className="h-4 w-2/3 bg-white/10 rounded" />
        </div>

        {/* 底部微光边缘 */}
        <div className="absolute bottom-0 left-0 right-0 h-px bg-gradient-to-r from-transparent via-white/20 to-transparent" />
      </div>
    );
  }

  if (error) {
    return <PageError title="岗位信息加载失败" description={error} onRetry={handleRetry} />;
  }

  if (!jobVm) {
    return (
      <PageError
        title="岗位不存在或已下线"
        description="该岗位可能已失效，请返回岗位列表查看其他机会。"
      />
    );
  }

  // 核心标签：岗位大类（jobTypeLabel="技术"） + 招聘项目（applicationType="应届实习"），去重
  const heroTags = Array.from(
    new Set([jobVm.jobTypeLabel, jobVm.applicationType].filter(Boolean))
  );

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

        {/* 核心标签：岗位大类 + 招聘项目 */}
        {heroTags.length > 0 && (
          <div className="flex items-center gap-2 mb-3 flex-wrap">
            {heroTags.map((tag, idx) => (
              <Badge
                key={tag}
                className={idx === 0
                  ? "bg-white/15 text-white border-white/10 px-2.5 py-0.5 rounded-sm font-medium"
                  : "bg-white/10 text-white border-transparent px-2.5 py-0.5 rounded-sm"
                }
              >
                {tag}
              </Badge>
            ))}
          </div>
        )}

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
