import React from "react";
import { MapPin } from "lucide-react";
import { useNavigate } from "react-router";
import { Badge } from "./ui/Badge";
import { Button } from "./ui/Button";
import { cn } from "../utils";

export interface JobCardProps {
  id: string;
  title: string;
  type: string;
  department: string;
  skills: string[];
  location: string;
  description: string;
  statusLabel?: string;
  selected?: boolean;
  onSelect?: (id: string) => void;
}

export function JobCard({
  id,
  title,
  type,
  department,
  skills,
  location,
  description,
  statusLabel,
  selected,
  onSelect,
}: JobCardProps) {
  const navigate = useNavigate();
  
  return (
    <div
      onClick={() => onSelect?.(id)}
      className={cn(
        "group relative p-5 bg-white border rounded-lg cursor-pointer transition-all duration-200 hover:shadow-md",
        selected
          ? "border-[#0052D9] bg-[#F4F8FF]"
          : "border-gray-200 hover:border-[#0052D9]"
      )}
    >
      <div className="flex justify-between items-start mb-3">
        <div className="flex-1 min-w-0">
          <h3 className={cn("text-lg font-bold mb-1 group-hover:text-[#0052D9] transition-colors", selected && "text-[#0052D9]")}>
            {title}
          </h3>
          <div className="flex items-center gap-2 flex-wrap text-xs text-gray-500">
            <span>{department}</span>
            <span className="w-px h-3 bg-gray-300"></span>
            <span>{type}</span>
            <span className="w-px h-3 bg-gray-300"></span>
            <span className="flex items-center gap-1">
              <MapPin className="h-3 w-3" />
              {location}
            </span>
            {statusLabel && (
              <span className="inline-flex items-center gap-1 text-[#0052D9] bg-blue-50 px-1.5 py-0.5 rounded-sm">
                <span className="w-1.5 h-1.5 rounded-full bg-[#0052D9]"></span>
                {statusLabel}
              </span>
            )}
          </div>
        </div>
      </div>

      <p className="text-sm text-gray-600 line-clamp-2 mb-4">
        {description}
      </p>

      <div className="flex items-center justify-between mt-auto pt-3 border-t border-gray-100">
        <div className="flex gap-1.5 flex-wrap">
          {skills.slice(0, 4).map((skill) => (
            <Badge key={skill} variant="outline" className="text-gray-500 text-[10px] px-1.5">
              {skill}
            </Badge>
          ))}
          {skills.length > 4 && (
            <span className="text-[10px] text-gray-400 self-center">+{skills.length - 4}</span>
          )}
        </div>
        <div className="flex gap-2 shrink-0">
          <Button 
            variant="text" 
            size="sm" 
            className="px-0 text-[#0052D9]"
            onClick={(e) => {
              e.stopPropagation();
              navigate(`/jobs/${id}`);
            }}
          >
            查看详情
          </Button>
          <Button 
            variant={selected ? "default" : "outline"} 
            size="sm"
            onClick={(e) => {
              e.stopPropagation();
              onSelect?.(id);
            }}
          >
            {selected ? "已选择" : "选择岗位"}
          </Button>
        </div>
      </div>
    </div>
  );
}
