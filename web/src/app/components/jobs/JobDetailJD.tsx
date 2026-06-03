import React, { useState, useEffect } from "react";
import { useParams } from "react-router";
import { Badge } from "../ui/Badge";
import { loadSelectedJobDetailViewModel } from "../../../api";
import type { JobDetailViewModel } from "../../../adapters";

/**
 * K8-5-9-C：重构岗位详情页主内容模块，对齐腾讯招聘详情页结构
 *
 * 设计说明：
 * - 第一张主白卡：岗位描述 / 岗位要求 / 加分项或注意事项 / 参加面试的城市 / 技能关键词 / 投递说明
 * - 第二张独立白卡：招聘部门和工作地 / 工作地点
 * - 底部：投递简历 CTA 按钮（UI only）
 * - 去掉 K8-5-1 的蓝色竖条标题，改为纯文字层级
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

  useEffect(() => {
    if (!jobId) { setLoading(false); return; }
    loadSelectedJobDetailViewModel(jobId).then(setJobVm).catch(() => {}).finally(() => setLoading(false));
  }, [jobId]);

  if (loading) return <div className="flex-1 flex items-center justify-center text-gray-400 text-sm py-20">正在加载岗位详情…</div>;
  if (!jobVm) return <div className="flex-1 flex items-center justify-center text-gray-500 text-sm py-20">暂无岗位详情</div>;

  // 准备数据
  const locationsText = safeJoin(jobVm.locations);
  const interviewLocationsText = safeJoin(jobVm.interviewLocations);
  const hasJobDescription = jobVm.responsibilities.length > 0 || jobVm.bannerSummary;
  const hasMustHave = jobVm.mustHave.length > 0;
  const hasNiceToHave = jobVm.niceToHave.length > 0;
  const hasBonusPoints = jobVm.bonusPoints.length > 0;
  const hasInterviewLocations = jobVm.interviewLocations.length > 0;
  const hasOrganizations = jobVm.organizations.length > 0;
  const hasSkills = jobVm.skills && jobVm.skills.length > 0;
  const skillsToShow = hasSkills ? jobVm.skills.slice(0, 12) : [];
  const skillsRemaining = hasSkills && jobVm.skills.length > 12 ? jobVm.skills.length - 12 : 0;

  // 标题样式：纯文字无装饰
  const sectionTitle = "text-[18px] font-semibold text-gray-900 mb-4";
  const subTitle = "text-[14px] font-semibold text-gray-800 mb-3";

  return (
    <div className="flex-1 space-y-6">
      {/* ================================================================ */}
      {/* 第一张主白卡：岗位 JD */}
      {/* ================================================================ */}
      <div className="bg-white border border-gray-200 rounded-xl p-8 md:p-10 shadow-sm">
        <div className="space-y-10 text-gray-700 text-[15px] leading-relaxed">

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

          {/* ── 技能关键词 ── */}
          {hasSkills && (
            <section>
              <h2 className={sectionTitle}>技能关键词</h2>
              <div className="flex flex-wrap gap-2">
                {skillsToShow.map((skill) => (
                  <span
                    key={skill}
                    className="inline-flex items-center px-2.5 py-1 bg-gray-50 text-gray-600 text-[13px] rounded border border-gray-100"
                  >
                    {skill}
                  </span>
                ))}
                {skillsRemaining > 0 && (
                  <span className="inline-flex items-center px-2.5 py-1 text-gray-400 text-[13px]">
                    +{skillsRemaining}
                  </span>
                )}
              </div>
            </section>
          )}

          {/* ── 投递说明 ── */}
          <section className="bg-blue-50/50 border border-blue-100/60 rounded-lg p-5">
            <h2 className="text-[14px] font-semibold text-gray-800 mb-2">投递说明</h2>
            <p className="text-[13px] text-gray-600 leading-relaxed">
              {jobVm.deliveryNotes || "该岗位支持使用当前简历进行 AI 匹配分析，分析结果仅作为投递前决策参考，不替代用户最终选择。"}
            </p>
          </section>
        </div>
      </div>

      {/* ================================================================ */}
      {/* 第二张独立白卡：招聘部门和工作地 */}
      {/* ================================================================ */}
      {hasOrganizations && (
        <div className="bg-white border border-gray-200 rounded-xl p-8 md:p-10 shadow-sm">
          <h2 className={sectionTitle}>招聘部门和工作地</h2>

          {/* 招聘部门 */}
          <div className="mb-6">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
              {jobVm.organizations.map((o, i) => (
                <div key={i} className="flex items-center gap-3 p-3 bg-gray-50 rounded-lg">
                  <span className="font-mono text-[#0052D9] text-[13px] font-semibold bg-blue-50 px-2 py-0.5 rounded">
                    {o.code}
                  </span>
                  <span className="text-gray-700 text-sm">{o.name}</span>
                </div>
              ))}
            </div>
          </div>

          {/* 工作地点 */}
          {locationsText && (
            <div>
              <h3 className={subTitle}>工作地点</h3>
              <div className="flex flex-wrap gap-2">
                {jobVm.locations.map((location, i) => (
                  <span key={i} className="inline-flex items-center px-3 py-1.5 bg-gray-50 text-gray-700 text-sm rounded-lg border border-gray-100">
                    {location}
                  </span>
                ))}
              </div>
            </div>
          )}
        </div>
      )}

      {/* ================================================================ */}
      {/* 投递简历 CTA（UI only） */}
      {/* ================================================================ */}
      <div className="flex justify-center mt-4">
        <button
          type="button"
          className="inline-flex items-center gap-2 rounded-full bg-[#0052D9] px-8 py-3 text-sm font-medium text-white shadow-sm transition-colors hover:bg-[#0040A8]"
        >
          投递简历
        </button>
      </div>
    </div>
  );
}
