import React, { useState, useEffect, useCallback } from "react";
import { useParams } from "react-router";
import { Badge } from "../ui/Badge";
import { loadSelectedJobDetailViewModel } from "../../../api";
import type { JobDetailViewModel } from "../../../adapters";
import { PageError } from "../PageState";
import { DetailSkeleton } from "../ui/DetailSkeleton";

/**
 * K8-5-12-D1：正文 loading/error/retry 修复
 *
 * 设计说明：
 * - 第一张主白卡：岗位描述 / 岗位要求 / 加分项或注意事项 / 参加面试的城市 / 技能关键词(弱化) / 投递说明(弱化)
 * - 第二张独立白卡：招聘部门和工作地 / 工作地点
 * - 底部：投递简历 CTA 按钮（UI only）
 * - loading: DetailSkeleton；error: PageError(onRetry)
 */

/** 安全的数组渲染辅助函数 */
function safeJoin(arr: string[] | undefined, separator: string = "、"): string {
  if (!Array.isArray(arr)) return "";
  return arr.filter(Boolean).join(separator);
}

export function JobDetailJD() {
  const { id: jobId } = useParams<{ id: string }>();
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
      .catch(() => setError("岗位详情加载失败，请稍后重试。"))
      .finally(() => setLoading(false));
  }, [jobId, retryCount]);

  const handleRetry = useCallback(() => setRetryCount((c) => c + 1), []);

  if (loading) return <DetailSkeleton sections={2} />;

  if (error) {
    return (
      <PageError
        title="岗位详情加载失败"
        description={error}
        onRetry={handleRetry}
      />
    );
  }

  if (!jobVm) {
    return (
      <PageError
        title="岗位不存在或已下线"
        description="该岗位可能已失效，请返回岗位列表查看其他机会。"
      />
    );
  }

  // 准备数据
  const locationsText = safeJoin(jobVm.locations);
  const interviewLocationsText = safeJoin(jobVm.interviewLocations);
  const hasJobDescription = jobVm.responsibilities.length > 0 || jobVm.bannerSummary;
  const hasMustHave = jobVm.mustHave.length > 0;
  const hasNiceToHave = jobVm.niceToHave.length > 0;
  const hasBonusPoints = jobVm.bonusPoints.length > 0;
  const hasInterviewLocations = jobVm.interviewLocations.length > 0;
  const hasOrganizations = jobVm.organizations.length > 0;

  // 标题样式：纯文字无装饰
  const sectionTitle = "text-[18px] font-semibold text-gray-900 mb-4";
  const subTitle = "text-[14px] font-semibold text-gray-800 mb-3";

  return (
    <div className="flex-1 space-y-6">
      {/* ================================================================ */}
      {/* 第一张主白卡：岗位 JD */}
      {/* ================================================================ */}
      <div className="bg-white rounded-2xl ring-1 ring-gray-100 p-8 md:p-10">
        <div className="space-y-10 text-gray-600 text-[14px] leading-7">

          {/* ── 岗位描述 ── */}
          {hasJobDescription && (
            <section>
              <h2 className={sectionTitle}>岗位描述</h2>
              {jobVm.bannerSummary && (
                <p className="mb-4 text-gray-600 leading-relaxed">{jobVm.bannerSummary}</p>
              )}
              {jobVm.responsibilities.length > 0 && (
                <ol className="list-decimal pl-5 space-y-2.5 marker:text-gray-400">
                  {jobVm.responsibilities.map((item, i) => (
                    <li key={i} className="pl-1">{item}</li>
                  ))}
                </ol>
              )}
            </section>
          )}

          {/* ── 岗位要求 ── */}
          {(hasMustHave || hasNiceToHave) && (
            <section>
              <h2 className={sectionTitle}>岗位要求</h2>

              {hasMustHave && (
                <div className="mb-5">
                  <h3 className={subTitle}>必须具备的：</h3>
                  <ol className="list-decimal pl-5 space-y-2 marker:text-gray-400">
                    {jobVm.mustHave.map((item, i) => (
                      <li key={i} className="pl-1">{item}</li>
                    ))}
                  </ol>
                </div>
              )}

              {hasNiceToHave && (
                <div>
                  <h3 className={subTitle}>有一定了解的：</h3>
                  <ol className="list-decimal pl-5 space-y-2 marker:text-gray-400">
                    {jobVm.niceToHave.map((item, i) => (
                      <li key={i} className="pl-1">{item}</li>
                    ))}
                  </ol>
                </div>
              )}
            </section>
          )}

          {/* ── 加分项或注意事项 ── */}
          {hasBonusPoints && (
            <section>
              <h2 className={sectionTitle}>加分项或注意事项</h2>
              <ul className="list-disc pl-5 space-y-2 marker:text-gray-300">
                {jobVm.bonusPoints.map((item, i) => (
                  <li key={i} className="pl-1">{item}</li>
                ))}
              </ul>
            </section>
          )}

          {/* ── 参加面试的城市 ── */}
          {hasInterviewLocations && (
            <section>
              <h2 className={sectionTitle}>参加面试的城市</h2>
              <p className="text-gray-600">{interviewLocationsText}</p>
            </section>
          )}

          {/* ── 投递说明（Offerr鹅增强，弱化展示） ── */}
          <section className="bg-blue-50/40 border border-blue-100/40 rounded-lg p-4">
            <h2 className="text-[13px] font-semibold text-gray-700 mb-1.5">投递说明</h2>
            <p className="text-[12px] text-gray-500 leading-relaxed">
              {jobVm.deliveryNotes || "该岗位支持使用当前简历进行 AI 匹配分析，分析结果仅作为投递前决策参考，不替代用户最终选择。"}
            </p>
          </section>
        </div>
      </div>

      {/* ================================================================ */}
      {/* 第二张独立白卡：招聘部门和工作地 */}
      {/* ================================================================ */}
      {hasOrganizations && (
        <div className="bg-white rounded-2xl ring-1 ring-gray-100 p-8 md:p-10">
          <h2 className={sectionTitle}>招聘部门和工作地</h2>

          {/* 招聘部门 */}
          <div className="mb-6">
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-x-6 gap-y-2.5">
              {jobVm.organizations.map((o, i) => (
                <div key={i} className="flex items-center gap-2.5">
                  <span className="font-mono text-[13px] font-semibold text-gray-700 shrink-0 w-10">
                    {o.code}
                  </span>
                  <span className="text-[13px] text-gray-500">{o.name}</span>
                </div>
              ))}
            </div>
          </div>

          {/* 工作地点 */}
          {locationsText && (
            <div>
              <h3 className={subTitle}>工作地点</h3>
              <p className="text-[13px] text-gray-500 leading-relaxed">{locationsText}</p>
            </div>
          )}
        </div>
      )}

    </div>
  );
}
