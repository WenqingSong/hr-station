import React from "react";
import { MapPin, ArrowRight } from "lucide-react";
import { useNavigate } from "react-router";

// ─── Props ────────────────────────────────────────────────────────────────────

export interface JobCardProps {
  id: string;
  title: string;
  /** 元信息行第一部分：岗位类型（如 技术/产品/设计） */
  jobTypeLabel: string;
  /** 元信息行第二部分：招聘类型（如 校园招聘/实习生） */
  typeLabel: string;
  /** 元信息行第三部分：方向（如 软件开发） */
  direction: string;
  /** K8-4-4-R5: 事业群缩写 codes（如 CDG CSIG） */
  organizationCodes: string[];
  /** 工作地点列表 */
  locations: string[];
  /** 岗位简介（最多 2 行） */
  description: string;
  /** 状态标签（如 "招聘中"） */
  statusLabel?: string;
}

// ─── JobCard ──────────────────────────────────────────────────────────────────

export function JobCard({
  id,
  title,
  jobTypeLabel,
  typeLabel,
  direction,
  organizationCodes,
  locations,
  description,
  statusLabel,
}: JobCardProps) {
  const navigate = useNavigate();

  const safeLocations = Array.isArray(locations) ? locations : [];
  const safeOrgCodes = Array.isArray(organizationCodes) ? organizationCodes : [];

  // 事业群 code 展示：前 8 个 + 余量
  const visibleOrgCodes = safeOrgCodes.slice(0, 8);
  const moreOrgCount = safeOrgCodes.length - 8;

  // 工作地点展示：前 6 个 + 余量
  const visibleLocations = safeLocations.slice(0, 6);
  const moreLocCount = safeLocations.length - 6;

  // 元信息各部分（有值才显示）
  const metaParts: string[] = [];
  if (jobTypeLabel) metaParts.push(jobTypeLabel);
  if (direction && direction !== jobTypeLabel) metaParts.push(direction);
  if (typeLabel) metaParts.push(typeLabel);

  return (
    <div className="group relative bg-white border border-gray-100 rounded-2xl hover:border-[#0052D9]/25 hover:shadow-sm transition-all duration-200">
      <div
        className="p-6 cursor-pointer"
        onClick={() => navigate(`/jobs/${id}`)}
      >
        {/* ── 标题行 + 状态徽章 ── */}
        <div className="flex items-start justify-between gap-3 mb-3">
          <h3 className="text-[18px] font-semibold text-gray-900 leading-snug group-hover:text-[#0052D9] transition-colors">
            {title}
          </h3>
          {statusLabel && (
            <span className="shrink-0 text-[12px] text-[#0052D9] bg-blue-50 px-2.5 py-1 rounded-md font-medium">
              {statusLabel}
            </span>
          )}
        </div>

        {/* ── 元信息行：岗位类型 | 方向 | 招聘类型 ── */}
        {metaParts.length > 0 && (
          <div className="flex items-center gap-1.5 flex-wrap text-[13px] text-gray-500 mb-2.5">
            {metaParts.map((part, idx) => (
              <React.Fragment key={part}>
                {idx > 0 && <span className="text-gray-300 select-none">|</span>}
                <span>{part}</span>
              </React.Fragment>
            ))}
          </div>
        )}

        {/* ── 事业群缩写 ── */}
        {visibleOrgCodes.length > 0 && (
          <div className="text-[13px] text-gray-400 mb-2.5">
            <span className="text-gray-400 mr-1">事业群：</span>
            {visibleOrgCodes.join(' ')}
            {moreOrgCount > 0 && (
              <span className="ml-1 text-gray-300">+{moreOrgCount}</span>
            )}
          </div>
        )}

        {/* ── 工作地点 ── */}
        <div className="flex items-center gap-1.5 text-[13px] text-gray-400 mb-3">
          <MapPin className="h-3.5 w-3.5 shrink-0" />
          <span>
            工作地点：
            {visibleLocations.length > 0
              ? visibleLocations.join(' ')
              : '地点不限'}
            {moreLocCount > 0 && (
              <span className="ml-1 text-gray-300">+{moreLocCount}</span>
            )}
          </span>
        </div>

        {/* ── 岗位简述 ── */}
        {description && (
          <p className="text-[13px] text-gray-400 leading-relaxed line-clamp-2 mb-4">
            {description}
          </p>
        )}

        {/* ── 查看详情 ── */}
        <div className="pt-3.5 border-t border-gray-50 flex justify-end">
          <span className="inline-flex items-center gap-1 text-[13px] text-[#0052D9] font-medium group-hover:text-[#0045b8] transition-colors">
            查看详情
            <ArrowRight className="h-3.5 w-3.5 group-hover:translate-x-0.5 transition-transform" />
          </span>
        </div>
      </div>
    </div>
  );
}
