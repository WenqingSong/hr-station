import React from "react";
import { MapPin, AlertCircle, ChevronRight } from "lucide-react";
import { Badge } from "./ui/Badge";
import { cn } from "../utils";

interface JobRec {
  id: string;
  rank: number;
  title: string;
  level: "强推荐" | "推荐" | "谨慎推荐";
  keywords: string[];
  location: string;
  reason: string;
  risk: string;
}

const RECOMMENDATIONS: JobRec[] = [
  {
    id: "1",
    rank: 1,
    title: "软件开发-后台开发方向",
    level: "强推荐",
    keywords: ["Go", "MySQL", "Redis", "微服务"],
    location: "深圳",
    reason: "技术栈高度匹配，项目经验覆盖岗位核心要求。",
    risk: "缺少高并发压测数据",
  },
  {
    id: "2",
    rank: 2,
    title: "云计算开发实习生",
    level: "推荐",
    keywords: ["Go", "Kubernetes", "微服务"],
    location: "深圳 / 北京",
    reason: "具备 Go 与微服务基础，但需补充云原生知识。",
    risk: "云原生项目证据偏少",
  },
  {
    id: "3",
    rank: 3,
    title: "数据开发实习生",
    level: "推荐",
    keywords: ["MySQL", "Redis", "后端开发"],
    location: "深圳 / 广州",
    reason: "数据库基础扎实，有一定数据处理经验。",
    risk: "缺少大数据计算框架经验",
  },
  {
    id: "4",
    rank: 4,
    title: "软件开发-PC 客户端开发方向",
    level: "谨慎推荐",
    keywords: ["C++", "网络协议"],
    location: "深圳",
    reason: "部分底层协议知识可复用，但方向偏差较大。",
    risk: "完全缺少客户端开发经验",
  },
  {
    id: "5",
    rank: 5,
    title: "算法-搜索提效推荐方向",
    level: "谨慎推荐",
    keywords: ["后端架构"],
    location: "北京 / 深圳",
    reason: "具备架构理解，但核心算法能力不匹配。",
    risk: "无推荐算法相关经历",
  },
];

export function RecommendationSidebar() {
  const selectedId = "1";

  const getLevelStyle = (level: string) => {
    switch (level) {
      case "强推荐":
        return "bg-green-50 text-green-700 border-green-200";
      case "推荐":
        return "bg-blue-50 text-blue-700 border-blue-200";
      case "谨慎推荐":
        return "bg-amber-50 text-amber-700 border-amber-200";
      default:
        return "bg-gray-50 text-gray-700";
    }
  };

  return (
    <div className="w-[380px] shrink-0 flex flex-col gap-4">
      {RECOMMENDATIONS.map((job) => {
        const isSelected = job.id === selectedId;
        
        return (
          <div
            key={job.id}
            className={cn(
              "group relative p-4 bg-white border rounded-lg cursor-pointer transition-all duration-200 hover:shadow-md",
              isSelected
                ? "border-[#0052D9] bg-[#F4F8FF] shadow-sm"
                : "border-gray-200 hover:border-[#0052D9]"
            )}
          >
            {isSelected && (
              <div className="absolute left-0 top-0 bottom-0 w-1 bg-[#0052D9] rounded-l-lg"></div>
            )}
            
            <div className="flex items-start justify-between mb-2">
              <div className="flex items-center gap-2">
                <span className={cn(
                  "text-[14px] font-black italic",
                  isSelected ? "text-[#0052D9]" : "text-gray-400"
                )}>#{job.rank}</span>
                <h3 className={cn(
                  "font-bold text-[16px]",
                  isSelected ? "text-[#0052D9]" : "text-gray-900"
                )}>{job.title}</h3>
              </div>
            </div>

            <div className="flex flex-wrap gap-2 items-center mb-3">
              <span className={cn("text-[11px] px-2 py-0.5 rounded-sm border", getLevelStyle(job.level))}>
                {job.level}
              </span>
              <div className="flex items-center text-[12px] text-gray-500">
                <MapPin className="w-3 h-3 mr-1" />
                {job.location}
              </div>
            </div>

            <div className="flex flex-wrap gap-1.5 mb-3">
              {job.keywords.map((kw, i) => (
                <Badge key={i} variant="secondary" className="text-[10px] font-normal px-1.5 py-0 bg-white border border-gray-100 text-gray-600">
                  {kw}
                </Badge>
              ))}
            </div>

            <div className="text-[12px] text-gray-600 bg-gray-50 p-2 rounded mb-2 leading-relaxed">
              {job.reason}
            </div>

            <div className="flex items-center justify-between mt-3 pt-3 border-t border-gray-100/50">
              <div className="flex items-center text-[11px] text-amber-600 bg-amber-50/50 px-1.5 py-0.5 rounded">
                <AlertCircle className="w-3 h-3 mr-1 text-amber-500" />
                {job.risk}
              </div>
              <span className={cn(
                "text-[12px] flex items-center font-medium transition-colors",
                isSelected ? "text-[#0052D9]" : "text-[#0052D9] opacity-80 group-hover:opacity-100"
              )}>
                查看推荐分析 <ChevronRight className="w-3.5 h-3.5 ml-0.5" />
              </span>
            </div>
          </div>
        );
      })}
    </div>
  );
}

