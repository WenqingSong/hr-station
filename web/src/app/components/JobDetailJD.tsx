import React, { useState, useEffect } from "react";
import { useParams } from "react-router";
import { Badge } from "./ui/Badge";
import { loadSelectedJobDetailViewModel } from "../../api";
import type { JobDetailViewModel } from "../../adapters";

/**
 * K8-5-1：重构岗位详情页主 JD 卡片，对齐腾讯招聘详情页
 * 
 * 设计说明：
 * - 第一张主卡片：岗位 JD，包含岗位描述、岗位要求、加分项或注意事项、参加面试的城市
 * - 第二张独立模块：招聘部门和工作地
 * - 使用正式招聘页面文案，不使用后台字段名
 * - 页面对齐修复，卡片宽度、间距、对齐统一
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

  return (
    <div className="flex-1 space-y-6">
      {/* 第一张主卡片：岗位 JD */}
      <div className="bg-white border border-gray-200 rounded-xl p-8 md:p-10 shadow-sm">
        <div className="space-y-8 text-gray-700 text-[15px] leading-relaxed">
          
          {/* 岗位描述 */}
          {hasJobDescription && (
            <section>
              <h3 className="font-bold text-gray-900 mb-4 text-[18px] flex items-center gap-2">
                <span className="w-1 h-5 bg-[#0052D9] rounded-full inline-block"></span>
                岗位描述
              </h3>
              
              {/* 简短描述 */}
              {jobVm.bannerSummary && (
                <p className="mb-4 text-gray-600">{jobVm.bannerSummary}</p>
              )}
              
              {/* 职责列表 */}
              {jobVm.responsibilities.length > 0 && (
                <ol className="list-decimal pl-5 space-y-2.5 marker:text-gray-400">
                  {jobVm.responsibilities.map((item, i) => (
                    <li key={i} className="pl-1">{item}</li>
                  ))}
                </ol>
              )}
            </section>
          )}

          {/* 岗位要求 */}
          {(hasMustHave || hasNiceToHave) && (
            <section>
              <h3 className="font-bold text-gray-900 mb-4 text-[18px] flex items-center gap-2">
                <span className="w-1 h-5 bg-[#0052D9] rounded-full inline-block"></span>
                岗位要求
              </h3>
              
              {/* 必须具备的 */}
              {hasMustHave && (
                <div className="mb-5">
                  <h4 className="font-semibold text-gray-800 mb-2.5 text-[15px]">必须具备的：</h4>
                  <ol className="list-decimal pl-5 space-y-2 marker:text-gray-400">
                    {jobVm.mustHave.map((item, i) => (
                      <li key={i} className="pl-1">{item}</li>
                    ))}
                  </ol>
                </div>
              )}
              
              {/* 有一定了解的 */}
              {hasNiceToHave && (
                <div>
                  <h4 className="font-semibold text-gray-800 mb-2.5 text-[15px]">有一定了解的：</h4>
                  <ol className="list-decimal pl-5 space-y-2 marker:text-gray-400">
                    {jobVm.niceToHave.map((item, i) => (
                      <li key={i} className="pl-1">{item}</li>
                    ))}
                  </ol>
                </div>
              )}
            </section>
          )}

          {/* 加分项或注意事项 */}
          {hasBonusPoints && (
            <section>
              <h3 className="font-bold text-gray-900 mb-4 text-[18px] flex items-center gap-2">
                <span className="w-1 h-5 bg-[#0052D9] rounded-full inline-block"></span>
                加分项或注意事项
              </h3>
              <ul className="list-disc pl-5 space-y-2.5 marker:text-gray-300">
                {jobVm.bonusPoints.map((item, i) => (
                  <li key={i}>{item}</li>
                ))}
              </ul>
            </section>
          )}

          {/* 参加面试的城市 */}
          {hasInterviewLocations && (
            <section>
              <h3 className="font-bold text-gray-900 mb-4 text-[18px] flex items-center gap-2">
                <span className="w-1 h-5 bg-[#0052D9] rounded-full inline-block"></span>
                参加面试的城市
              </h3>
              <p className="pl-3 text-gray-600">{interviewLocationsText}</p>
            </section>
          )}

          {/* 投递说明 */}
          <section className="bg-[#F4F8FF] p-5 rounded-lg border border-blue-100 mt-6">
            <h3 className="font-bold text-[#0052D9] mb-2 text-sm flex items-center gap-1.5">投递说明</h3>
            <p className="text-sm text-gray-600 leading-relaxed">
              {jobVm.deliveryNotes || "该岗位支持使用当前简历进行 AI 匹配分析，分析结果仅作为投递前决策参考，不替代用户最终选择。"}
            </p>
          </section>
        </div>
      </div>

      {/* 第二张独立模块：招聘部门和工作地 */}
      {hasOrganizations && (
        <div className="bg-white border border-gray-200 rounded-xl p-8 md:p-10 shadow-sm">
          <h3 className="font-bold text-gray-900 mb-6 text-[18px] flex items-center gap-2">
            <span className="w-1 h-5 bg-[#0052D9] rounded-full inline-block"></span>
            招聘部门和工作地
          </h3>
          
          {/* 招聘部门 */}
          <div className="mb-6">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
              {jobVm.organizations.map((o, i) => (
                <div key={i} className="flex items-center gap-3 p-3 bg-gray-50 rounded-lg">
                  <span className="font-mono text-[#0052D9] text-sm font-semibold bg-blue-50 px-2 py-1 rounded">
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
              <h4 className="font-semibold text-gray-800 mb-3 text-[15px]">工作地点</h4>
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
    </div>
  );
}
