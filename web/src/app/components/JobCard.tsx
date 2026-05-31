import React from "react";
import { MapPin, Briefcase } from "lucide-react";
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
        <div>
          <h3 className={cn("text-lg font-bold mb-2 group-hover:text-[#0052D9] transition-colors", selected && "text-[#0052D9]")}>
            {title}
          </h3>
          <div className="flex items-center gap-2 flex-wrap mb-3">
            <Badge variant="secondary">{type}</Badge>
            <Badge variant="secondary">{department}</Badge>
            <div className="w-[1px] h-3 bg-gray-300 mx-1"></div>
            <div className="flex items-center text-sm text-gray-500 gap-1">
              <MapPin className="h-3.5 w-3.5" />
              <span>{location}</span>
            </div>
            <div className="flex items-center text-sm text-[#0052D9] gap-1 bg-blue-50 px-2 py-0.5 rounded-sm ml-2">
              <span className="w-1.5 h-1.5 rounded-full bg-[#0052D9]"></span>
              招聘中
            </div>
          </div>
        </div>
      </div>

      <p className="text-sm text-gray-600 line-clamp-2 mb-4">
        {description}
      </p>

      <div className="flex items-center justify-between mt-auto pt-4 border-t border-gray-100">
        <div className="flex gap-2 flex-wrap">
          {skills.map((skill) => (
            <Badge key={skill} variant="outline" className="text-gray-500">
              {skill}
            </Badge>
          ))}
        </div>
        <div className="flex gap-3 shrink-0">
          <Button 
            variant="text" 
            size="sm" 
            className="px-0"
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
