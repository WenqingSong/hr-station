import React, { useState, useEffect } from "react";
import { useParams } from "react-router";
import { Badge } from "./ui/Badge";
import { loadSelectedJobDetailViewModel } from "../../api";
import type { JobDetailViewModel } from "../../adapters";

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

  return (
    <div className="flex-1 max-w-[820px] space-y-6">
      {/* ── 第一张主卡片：岗位 JD ── */}
      <div className="bg-white border border-gray-200 rounded-xl p-8 shadow-sm">
        <div className="space-y-8 text-gray-700 text-[15px] leading-relaxed">
          {/* 岗位描述 */}
          {(jobVm.description || jobVm.responsibilities.length > 0) && (
            <section>
              <h2 className="text-[18px] font-bold text-gray-900 mb-4 flex items-center gap-2">
                <span className="w-1 h-5 bg-[#0052D9] rounded-full inline-block"></span>
                岗位描述
              </h2>
              {jobVm.description && (
                <p className="mb-4 text-gray-600">{jobVm.description}</p>
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

          {/* 岗位要求 */}
          {(jobVm.mustHave.length > 0 || jobVm.niceToHave.length > 0) && (
            <section>
              <h2 className="text-[18px] font-bold text-gray-900 mb-4 flex items-center gap-2">
                <span className="w-1 h-5 bg-[#0052D9] rounded-full inline-block"></span>
                岗位要求
              </h2>

              {jobVm.mustHave.length > 0 && (
                <div className="mb-4">
                  <h3 className="text-[15px] font-semibold text-gray-800 mb-2">必须具备的：</h3>
                  <ul className="list-disc pl-5 space-y-2 marker:text-gray-300">
                    {jobVm.mustHave.map((item, i) => (
                      <li key={i}>{item}</li>
                    ))}
                  </ul>
                </div>
              )}

              {jobVm.niceToHave.length > 0 && (
                <div>
                  <h3 className="text-[15px] font-semibold text-gray-800 mb-2">有一定了解的：</h3>
                  <ul className="list-disc pl-5 space-y-2 marker:text-gray-300">
                    {jobVm.niceToHave.map((item, i) => (
                      <li key={i}>{item}</li>
                    ))}
                  </ul>
                </div>
              )}
            </section>
          )}

          {/* 加分项或注意事项 */}
          {jobVm.bonusPoints.length > 0 && (
            <section>
              <h2 className="text-[18px] font-bold text-gray-900 mb-4 flex items-center gap-2">
                <span className="w-1 h-5 bg-[#0052D9] rounded-full inline-block"></span>
                加分项或注意事项
              </h2>
              <ul className="list-disc pl-5 space-y-2.5 marker:text-gray-300">
                {jobVm.bonusPoints.map((item, i) => (
                  <li key={i}>{item}</li>
                ))}
              </ul>
            </section>
          )}

          {/* 参加面试的城市 */}
          {jobVm.interviewLocations.length > 0 && (
            <section>
              <h2 className="text-[18px] font-bold text-gray-900 mb-4 flex items-center gap-2">
                <span className="w-1 h-5 bg-[#0052D9] rounded-full inline-block"></span>
                参加面试的城市
              </h2>
              <p className="pl-3 text-gray-600">{jobVm.interviewLocations.join("、")}</p>
            </section>
          )}
        </div>
      </div>

      {/* ── 第二张独立模块：招聘部门和工作地 ── */}
      {(jobVm.organizations.length > 0 || jobVm.locations.length > 0) && (
        <div className="bg-white border border-gray-200 rounded-xl p-8 shadow-sm">
          <h2 className="text-[18px] font-bold text-gray-900 mb-6 flex items-center gap-2">
            <span className="w-1 h-5 bg-[#0052D9] rounded-full inline-block"></span>
            招聘部门和工作地
          </h2>

          {/* 招聘部门 */}
          {jobVm.organizations.length > 0 && (
            <div className="mb-6">
              <h3 className="text-[15px] font-semibold text-gray-800 mb-3">招聘部门</h3>
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
                {jobVm.organizations.map((o, i) => (
                  <div key={i} className="flex items-center gap-2 text-sm py-1.5 px-3 bg-gray-50 rounded-lg">
                    <span className="font-mono text-xs text-gray-400 bg-gray-100 px-1.5 py-0.5 rounded">{o.code}</span>
                    <span className="text-gray-700">{o.name}</span>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* 工作地点 */}
          {jobVm.locations.length > 0 && (
            <div>
              <h3 className="text-[15px] font-semibold text-gray-800 mb-3">工作地点</h3>
              <div className="flex flex-wrap gap-2">
                {jobVm.locations.map((loc, i) => (
                  <span key={i} className="inline-flex items-center px-3 py-1.5 bg-blue-50/50 text-blue-700 border border-blue-100 rounded-lg text-sm">
                    {loc}
                  </span>
                ))}
              </div>
            </div>
          )}
        </div>
      )}

      {/* ── 投递说明 ── */}
      <div className="bg-[#F4F8FF] border border-blue-100 rounded-xl p-5">
        <h3 className="font-bold text-[#0052D9] mb-2 text-sm">投递说明</h3>
        <p className="text-sm text-gray-600 leading-relaxed">
          {jobVm.deliveryNotes || "该岗位支持使用当前简历进行 AI 匹配分析，分析结果仅作为投递前决策参考，不替代用户最终选择。"}
        </p>
      </div>
    </div>
  );
}
