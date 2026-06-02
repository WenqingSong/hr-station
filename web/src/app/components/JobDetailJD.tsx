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
    <div className="flex-1 bg-white border border-gray-200 rounded-lg p-10 min-w-[760px] max-w-[820px] shadow-sm">
      <div className="space-y-10 text-gray-700 text-[15px] leading-relaxed">
        {/* 岗位职责 */}
        {jobVm.responsibilities.length > 0 && (
          <section>
            <h3 className="font-bold text-gray-900 mb-4 text-[17px] flex items-center gap-2">
              <span className="w-1 h-4 bg-[#0052D9] rounded-full inline-block"></span>
              岗位职责
            </h3>
            <ul className="list-disc pl-5 space-y-2.5 marker:text-gray-300">
              {jobVm.responsibilities.map((item, i) => (<li key={i}>{item}</li>))}
            </ul>
          </section>
        )}

        {/* 必须具备 */}
        {jobVm.mustHave.length > 0 && (
          <section>
            <h3 className="font-bold text-gray-900 mb-4 text-[17px] flex items-center gap-2">
              <span className="w-1 h-4 bg-[#0052D9] rounded-full inline-block"></span>
              必须具备
            </h3>
            <ul className="list-disc pl-5 space-y-2.5 marker:text-gray-300">
              {jobVm.mustHave.map((item, i) => (<li key={i}>{item}</li>))}
            </ul>
          </section>
        )}

        {/* 有一定了解 */}
        {jobVm.niceToHave.length > 0 && (
          <section>
            <h3 className="font-bold text-gray-900 mb-4 text-[17px] flex items-center gap-2">
              <span className="w-1 h-4 bg-[#0052D9] rounded-full inline-block"></span>
              有一定了解
            </h3>
            <ul className="list-disc pl-5 space-y-2.5 marker:text-gray-300">
              {jobVm.niceToHave.map((item, i) => (<li key={i}>{item}</li>))}
            </ul>
          </section>
        )}

        {/* 加分项 */}
        {jobVm.bonusPoints.length > 0 && (
          <section>
            <h3 className="font-bold text-gray-900 mb-4 text-[17px] flex items-center gap-2">
              <span className="w-1 h-4 bg-[#0052D9] rounded-full inline-block"></span>
              加分项或注意事项
            </h3>
            <ul className="list-disc pl-5 space-y-2.5 marker:text-gray-300">
              {jobVm.bonusPoints.map((item, i) => (<li key={i}>{item}</li>))}
            </ul>
          </section>
        )}

        {/* 备注 */}
        {jobVm.notes.length > 0 && (
          <section>
            <h3 className="font-bold text-gray-900 mb-4 text-[17px] flex items-center gap-2">
              <span className="w-1 h-4 bg-[#0052D9] rounded-full inline-block"></span>
              备注
            </h3>
            <ul className="list-disc pl-5 space-y-2.5 marker:text-gray-300">
              {jobVm.notes.map((item, i) => (<li key={i}>{item}</li>))}
            </ul>
          </section>
        )}

        {/* 技能关键词 */}
        {jobVm.skills.length > 0 && (
          <section>
            <h3 className="font-bold text-gray-900 mb-4 text-[17px] flex items-center gap-2">
              <span className="w-1 h-4 bg-[#0052D9] rounded-full inline-block"></span>
              技能关键词
            </h3>
            <div className="flex flex-wrap gap-2">
              {jobVm.skills.map((s) => (
                <Badge key={s} variant="outline" className="text-xs px-2.5 py-1 bg-blue-50/50 text-blue-700 border-blue-100">{s}</Badge>
              ))}
            </div>
          </section>
        )}

        {/* 面试城市 */}
        {jobVm.interviewLocations.length > 0 && (
          <section>
            <h3 className="font-bold text-gray-900 mb-4 text-[17px] flex items-center gap-2">
              <span className="w-1 h-4 bg-[#0052D9] rounded-full inline-block"></span>
              参加面试的城市
            </h3>
            <p className="pl-3">{jobVm.interviewLocations.join("、")}</p>
          </section>
        )}

        {/* 招聘部门 */}
        {jobVm.organizations.length > 0 && (
          <section>
            <h3 className="font-bold text-gray-900 mb-4 text-[17px] flex items-center gap-2">
              <span className="w-1 h-4 bg-[#0052D9] rounded-full inline-block"></span>
              招聘部门
            </h3>
            <ul className="list-none pl-3 space-y-1.5">
              {jobVm.organizations.map((o, i) => (
                <li key={i} className="text-sm">
                  <span className="font-mono text-gray-500 text-xs mr-2">{o.code}</span>
                  <span>{o.name}</span>
                </li>
              ))}
            </ul>
          </section>
        )}

        {/* 投递说明 */}
        <section className="bg-[#F4F8FF] p-5 rounded-lg border border-blue-100">
          <h3 className="font-bold text-[#0052D9] mb-2 text-sm flex items-center gap-1.5">投递说明</h3>
          <p className="text-sm text-gray-600 leading-relaxed">
            {jobVm.deliveryNotes || "该岗位支持使用当前简历进行 AI 匹配分析，分析结果仅作为投递前决策参考，不替代用户最终选择。"}
          </p>
        </section>
      </div>
    </div>
  );
}
