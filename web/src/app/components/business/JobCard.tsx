import { MapPin, Clock } from "lucide-react";
import { Tag } from "../ui/Tag";

interface JobCardProps {
  title: string;
  tags: string[];
  location: string;
  status: string;
  deadline?: string;
  isHot?: boolean;
}

export function JobCard({ title, tags, location, status, deadline, isHot }: JobCardProps) {
  return (
    <div className="group bg-white border border-slate-200 rounded-lg p-5 hover:shadow-md transition-shadow cursor-pointer relative overflow-hidden">
      {isHot && (
        <div className="absolute top-0 right-0 bg-red-500 text-white text-[10px] font-bold px-2 py-1 rounded-bl-lg z-10">
          HOT
        </div>
      )}
      
      <div className="flex justify-between items-start mb-3">
        <h3 className="text-lg font-semibold text-slate-800 group-hover:text-blue-600 transition-colors">
          {title}
        </h3>
        <span className="text-xs font-medium text-green-600 bg-green-50 px-2 py-1 rounded border border-green-100 shrink-0">
          {status}
        </span>
      </div>

      <div className="flex flex-wrap gap-2 mb-4">
        {tags.map((tag, i) => (
          <Tag key={i}>{tag}</Tag>
        ))}
      </div>

      <div className="flex items-center gap-4 text-xs text-slate-500">
        <div className="flex items-center gap-1">
          <MapPin className="w-3.5 h-3.5" />
          {location}
        </div>
        {deadline && (
          <div className="flex items-center gap-1">
            <Clock className="w-3.5 h-3.5" />
            {deadline}
          </div>
        )}
      </div>
    </div>
  );
}
