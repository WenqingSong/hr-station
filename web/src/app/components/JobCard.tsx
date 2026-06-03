import React from "react";
import { useNavigate } from "react-router";

// ─── Props ────────────────────────────────────────────────────────────────────

export interface JobCardProps {
  id: string;
  title: string;
  /** 元信息行第一部分：岗位类型（如 技术/产品/设计） */
  jobTypeLabel: string;
  /** 元信息行第二部分：应聘类型（如 应届实习/日常实习/校园招聘） */
  applicationType: string;
  /** 元信息行第三部分：事业群缩写 codes（如 CDG CSIG） */
  organizationCodes: string[];
  /** 工作地点列表 */
  locations: string[];
}

// ─── JobCard ──────────────────────────────────────────────────────────────────

export function JobCard({
  id,
  title,
  jobTypeLabel,
  applicationType,
  organizationCodes,
  locations,
}: JobCardProps) {
  const navigate = useNavigate();

  const safeLocations = Array.isArray(locations) ? locations : [];
  const safeOrgCodes = Array.isArray(organizationCodes) ? organizationCodes : [];

  // 事业群 code 展示：全部展示
  const visibleOrgCodes = safeOrgCodes.join(' ');

  // 工作地点展示：前 7 个 + 余量
  const visibleLocations = safeLocations.slice(0, 7);
  const moreLocCount = safeLocations.length - 7;

  // 元信息行：岗位类型 | 应聘类型 | 事业群 codes
  const metaParts: string[] = [];
  if (jobTypeLabel) metaParts.push(jobTypeLabel);
  if (applicationType) metaParts.push(applicationType);
  if (visibleOrgCodes) metaParts.push(visibleOrgCodes);

  return (
    <div className="group relative bg-white rounded-[18px] hover:shadow-[0_2px_12px_rgba(0,0,0,0.06)] transition-all duration-200">
      <div
        className="py-4 px-5 sm:px-6 cursor-pointer"
        onClick={() => navigate(`/jobs/${id}`)}
        role="link"
        tabIndex={0}
        onKeyDown={(e) => { if (e.key === 'Enter') navigate(`/jobs/${id}`); }}
      >
        {/* ── 标题行 ── */}
        <h3 className="text-[18px] font-semibold text-gray-900 leading-snug group-hover:text-[#0052D9] transition-colors mb-2">
          {title}
        </h3>

        {/* ── 元信息行：技术 | 应届实习 | CDG CSIG IEG PCG TEG WXG S1 S2 ── */}
        {metaParts.length > 0 && (
          <div className="flex items-center gap-1.5 flex-wrap text-[13px] text-gray-500 mb-2">
            {metaParts.map((part, idx) => (
              <React.Fragment key={part}>
                {idx > 0 && <span className="text-gray-300 select-none">|</span>}
                <span>{part}</span>
              </React.Fragment>
            ))}
          </div>
        )}

        {/* ── 工作地点 ── */}
        <div className="text-[13px] text-gray-400">
          工作地点：
          {visibleLocations.length > 0
            ? <span>{visibleLocations.join(' ')}{moreLocCount > 0 ? <span className="ml-1 text-gray-300">+{moreLocCount}</span> : null}</span>
            : '地点不限'}
        </div>
      </div>
    </div>
  );
}
