import React from "react";
import { MapPin, ArrowRight } from "lucide-react";
import { useNavigate } from "react-router";
import { Badge } from "./ui/Badge";

export interface JobCardProps {
  id: string;
  title: string;
  direction: string;
  type: string;
  department: string;
  businessGroups: string[];
  skills: string[];
  location: string;
  locations: string[];
  description: string;
  statusLabel?: string;
}

export function JobCard({
  id,
  title,
  direction,
  type,
  department,
  businessGroups,
  skills,
  location,
  locations,
  description,
  statusLabel,
}: JobCardProps) {
  const navigate = useNavigate();

  const safeSkills = Array.isArray(skills) ? skills : [];
  const safeBusinessGroups = Array.isArray(businessGroups) ? businessGroups : [];
  const safeLocations = Array.isArray(locations) && locations.length > 0 ? locations : (location ? [location] : []);

  return (
    <div
      className="group relative p-6 bg-white border border-gray-100 rounded-xl transition-all duration-200 hover:shadow-md hover:border-[#0052D9]/30 cursor-pointer"
      onClick={() => navigate(`/jobs/${id}`)}
    >
      {/* ── 上方：标题 + 状态 ── */}
      <div className="flex items-start justify-between gap-3 mb-3">
        <h3 className="text-[17px] font-bold text-gray-900 leading-snug group-hover:text-[#0052D9] transition-colors">
          {title}
        </h3>
        {statusLabel && (
          <span className="shrink-0 inline-flex items-center gap-1.5 text-[13px] text-[#0052D9] bg-blue-50/70 px-2.5 py-1 rounded-md font-medium">
            <span className="w-1.5 h-1.5 rounded-full bg-[#0052D9]" />
            {statusLabel}
          </span>
        )}
      </div>

      {/* ── 信息副标题行 ── */}
      <div className="flex items-center gap-2 flex-wrap text-[13px] text-gray-500 mb-2.5">
        {direction && (
          <>
            <span>{direction}</span>
            <span className="w-px h-3 bg-gray-300" />
          </>
        )}
        <span>{type}</span>
        <span className="w-px h-3 bg-gray-300" />
        <span>{department}</span>
      </div>

      {/* ── 事业群标签 ── */}
      {safeBusinessGroups.length > 0 && (
        <div className="flex flex-wrap gap-1.5 mb-2.5">
          {safeBusinessGroups.map((bg) => (
            <span key={bg} className="text-[12px] text-gray-400 bg-gray-50 px-2 py-0.5 rounded">
              {bg}
            </span>
          ))}
        </div>
      )}

      {/* ── 工作地点 ── */}
      <div className="flex items-center gap-1 text-[13px] text-gray-400 mb-3">
        <MapPin className="h-3.5 w-3.5 shrink-0" />
        <span>{safeLocations.join(" · ")}</span>
      </div>

      {/* ── 岗位简述 ── */}
      {description && (
        <p className="text-[14px] text-gray-500 leading-relaxed line-clamp-2 mb-4">
          {description}
        </p>
      )}

      {/* ── 底部：技能标签 + 操作 ── */}
      <div className="flex items-center justify-between pt-3.5 border-t border-gray-100">
        <div className="flex gap-1.5 flex-wrap items-center">
          {safeSkills.slice(0, 5).map((skill) => (
            <span
              key={skill}
              className="text-[12px] text-gray-500 bg-gray-50 border border-gray-100 px-2 py-0.5 rounded-md"
            >
              {skill}
            </span>
          ))}
          {safeSkills.length > 5 && (
            <span className="text-[12px] text-gray-400">+{safeSkills.length - 5}</span>
          )}
          {safeSkills.length === 0 && (
            <span className="text-[12px] text-gray-400">暂无技能标签</span>
          )}
        </div>
        <button
          className="inline-flex items-center gap-1 text-[13px] text-[#0052D9] font-medium hover:text-[#0045b8] transition-colors shrink-0 ml-3"
          onClick={(e) => {
            e.stopPropagation();
            navigate(`/jobs/${id}`);
          }}
        >
          查看详情
          <ArrowRight className="h-3.5 w-3.5" />
        </button>
      </div>
    </div>
  );
}
