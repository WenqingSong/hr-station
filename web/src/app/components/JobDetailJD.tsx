import React, { useState, useEffect } from "react";
import { loadSelectedJobDetailViewModel } from "../../api";
import type { JobDetailViewModel } from "../../adapters";

interface JobDetailJDProps {
  jobId?: string;
}

export function JobDetailJD({ jobId }: JobDetailJDProps) {
  const [jobVm, setJobVm] = useState<JobDetailViewModel | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!jobId) {
      setLoading(false);
      return;
    }
    loadSelectedJobDetailViewModel(jobId)
      .then(setJobVm)
      .catch(() => {})
      .finally(() => setLoading(false));
  }, [jobId]);

  if (loading) return <div className="flex-1 flex items-center justify-center text-gray-400 text-sm py-20">正在加载岗位详情…</div>;
  if (!jobVm) return <div className="flex-1 flex items-center justify-center text-gray-500 text-sm py-20">暂无岗位详情</div>;
  return (
    <div className="flex-1 bg-white border border-gray-200 rounded-lg p-10 min-w-[760px] max-w-[820px] shadow-sm">
      <div className="space-y-10 text-gray-700 text-[15px] leading-relaxed">
        {/* Section 1 */}
        <section>
          <h3 className="font-bold text-gray-900 mb-4 text-[17px] flex items-center gap-2">
            <span className="w-1 h-4 bg-[#0052D9] rounded-full inline-block"></span>
            岗位描述
          </h3>
          <ul className="list-disc pl-5 space-y-2.5 marker:text-gray-300">
            {jobVm.responsibilities.map((item, i) => (
              <li key={i}>{item}</li>
            ))}
          </ul>
        </section>

        {/* Section 2 */}
        <section>
          <h3 className="font-bold text-gray-900 mb-4 text-[17px] flex items-center gap-2">
            <span className="w-1 h-4 bg-[#0052D9] rounded-full inline-block"></span>
            岗位要求
          </h3>
          <p className="mb-3 font-medium text-gray-800">必须具备：</p>
          <ul className="list-disc pl-5 space-y-2.5 marker:text-gray-300">
            {jobVm.requirements.map((item, i) => (
              <li key={i}>{item}</li>
            ))}
          </ul>
        </section>

        {/* Section 3 */}
        <section>
          <h3 className="font-bold text-gray-900 mb-4 text-[17px] flex items-center gap-2">
            <span className="w-1 h-4 bg-[#0052D9] rounded-full inline-block"></span>
            加分项或注意事项
          </h3>
          <ul className="list-disc pl-5 space-y-2.5 marker:text-gray-300">
            {jobVm.bonusPoints.map((item, i) => (
              <li key={i}>{item}</li>
            ))}
          </ul>
        </section>

        {/* Section 4 */}
        <section>
          <h3 className="font-bold text-gray-900 mb-4 text-[17px] flex items-center gap-2">
            <span className="w-1 h-4 bg-[#0052D9] rounded-full inline-block"></span>
            工作地点
          </h3>
          <p className="pl-3">{jobVm.locationsText}</p>
        </section>

        <div className="pt-6 mt-6 border-t border-gray-100"></div>

        {/* Section 5 */}
        <section className="bg-[#F4F8FF] p-5 rounded-lg border border-blue-100">
          <h3 className="font-bold text-[#0052D9] mb-2 text-sm flex items-center gap-1.5">
            投递说明
          </h3>
          <p className="text-sm text-gray-600 leading-relaxed">
            {jobVm.deliveryNotes || "该岗位支持使用当前简历进行 AI 匹配分析，分析结果仅作为投递前决策参考，不替代用户最终选择。"}
          </p>
        </section>
      </div>
    </div>
  );
}
