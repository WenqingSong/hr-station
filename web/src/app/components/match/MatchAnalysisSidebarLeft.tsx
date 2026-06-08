import React from "react";
import { MapPin, Target, LayoutList, Tags } from "lucide-react";
import { Badge } from "../ui/Badge";
import { toJobDetailViewModelFromApi } from "../../../adapters";
import type { JobApiItem } from "../../../api/jobs";

interface MatchAnalysisSidebarLeftProps {
  job?: JobApiItem;
  jobSource?: string;
  /** K9-2-Regression-F-B: job requirements for keyword derivation */
  jobRequirements?: string[];
}

function deriveKeywords(reqs: string[], title: string): string[] {
  const all = [...reqs.map(r => r.replace(/与.*相关.*$/, "").replace(/能力$/, "").trim()), title];
  // Deduplicate and limit to 6
  const seen = new Set<string>();
  const kw: string[] = [];
  for (const w of all) {
    if (!seen.has(w) && w.length > 0 && w.length <= 12) { seen.add(w); kw.push(w); }
    if (kw.length >= 6) break;
  }
  return kw;
}

export function MatchAnalysisSidebarLeft({ job, jobSource = "默认岗位", jobRequirements }: MatchAnalysisSidebarLeftProps) {
  const jobVm = toJobDetailViewModelFromApi(job);
  const keywords = deriveKeywords(jobRequirements || jobVm.requirements, jobVm.title);
  return (
    <aside className="w-[320px] shrink-0 flex flex-col gap-4">
      <div className="bg-white border border-gray-200 rounded-lg p-5 shadow-sm">
        <div className="flex items-center gap-2 text-gray-500 text-[13px] mb-3">
          <Target className="h-4 w-4" />
          岗位 JD 摘要
        </div>
        
        <h2 className="font-bold text-gray-900 text-[18px] mb-3">{jobVm.title}</h2>
        
        <div className="flex flex-wrap gap-1.5 mb-4">
          {jobVm.typeLabel && <Badge variant="secondary" className="text-[11px] px-1.5 font-normal">{jobVm.typeLabel}</Badge>}
          {jobVm.categoryLabel && <Badge variant="secondary" className="text-[11px] px-1.5 font-normal">{jobVm.categoryLabel}</Badge>}
        </div>

        <div className="flex items-center text-[13px] text-gray-600 mb-5">
          <MapPin className="w-3.5 h-3.5 mr-1.5 text-gray-400" />
          {jobVm.locationsText}
        </div>

        <div className="pt-4 border-t border-gray-100 mb-5">
          <h3 className="text-[13px] font-bold text-gray-900 mb-3 flex items-center gap-1.5">
            <LayoutList className="w-3.5 h-3.5 text-gray-400" />
            核心要求摘要
          </h3>
          <ul className="space-y-2.5 text-[12px] text-gray-600 list-disc pl-4 marker:text-gray-300 leading-relaxed">
            {jobVm.requirements.slice(0, 5).map((item, i) => (
              <li key={i}>{item}</li>
            ))}
          </ul>
        </div>

        <div className="pt-4 border-t border-gray-100">
          <h3 className="text-[13px] font-bold text-gray-900 mb-3 flex items-center gap-1.5">
            <Tags className="w-3.5 h-3.5 text-gray-400" />
            JD 关键词
          </h3>
          {keywords.length > 0 ? (
            <div className="flex flex-wrap gap-1.5">
              {keywords.map((kw, i) => (
                <span key={i} className="px-2 py-0.5 bg-slate-100 text-slate-600 text-[11px] rounded font-medium">{kw}</span>
              ))}
            </div>
          ) : (
            <p className="text-[12px] text-gray-400">关键词将在深度分析后自动提取</p>
          )}
        </div>
      </div>
    </aside>
  );
}
