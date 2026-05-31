import React from "react";
import { FileText, Eye, CheckCircle2, Edit3, Clock, Briefcase } from "lucide-react";
import { Badge } from "./ui/Badge";

export function MyResumesList() {
  const resumes = [
    {
      id: "v2",
      name: "V2 后端开发简历",
      status: "当前使用",
      direction: "后端开发",
      keywords: ["Go", "MySQL", "Redis", "微服务"],
      updateTime: "2 小时前",
      projects: 3,
      analysisCount: 3,
      isActive: true,
    },
    {
      id: "v1",
      name: "V1 通用后端简历",
      status: "历史版本",
      direction: "后端开发",
      keywords: ["Go", "Java", "MySQL"],
      updateTime: "3 天前",
      projects: 2,
      analysisCount: 0,
      isActive: false,
    },
    {
      id: "v3",
      name: "V3 产品方向简历",
      status: "未启用",
      direction: "产品策划",
      keywords: ["需求分析", "用户调研", "原型设计"],
      updateTime: "1 周前",
      projects: 2,
      analysisCount: 0,
      isActive: false,
    }
  ];

  return (
    <aside className="w-[320px] xl:w-[380px] shrink-0 flex flex-col gap-3">
      {resumes.map((resume) => (
        <div 
          key={resume.id}
          className={`bg-white rounded-lg p-4 border transition-all ${
            resume.isActive 
              ? "border-[#0052D9] shadow-[0_0_0_1px_rgba(0,82,217,0.1)]" 
              : "border-gray-200 shadow-sm hover:border-gray-300"
          }`}
        >
          <div className="flex justify-between items-start mb-3">
            <h3 className="font-bold text-gray-900 text-[15px] flex items-center gap-2">
              <FileText className={`w-4 h-4 ${resume.isActive ? "text-[#0052D9]" : "text-gray-400"}`} />
              {resume.name}
            </h3>
            {resume.isActive ? (
              <Badge className="bg-blue-50 text-[#0052D9] border-blue-200 font-medium px-2 py-0.5">
                当前使用
              </Badge>
            ) : (
              <Badge variant="outline" className="text-gray-500 bg-gray-50 border-gray-200 font-normal px-2 py-0.5">
                {resume.status}
              </Badge>
            )}
          </div>
          
          <div className="space-y-2 mb-4 text-[12px] text-gray-600">
            <div className="flex items-center gap-2">
              <Briefcase className="w-3.5 h-3.5 text-gray-400" />
              <span className="text-gray-500 w-[48px]">方向：</span>
              <span className="font-medium text-gray-800">{resume.direction}</span>
            </div>
            <div className="flex items-start gap-2">
              <span className="text-gray-500 shrink-0 w-[48px] ml-5">关键词：</span>
              <span className="text-gray-700 truncate">{resume.keywords.join(" / ")}</span>
            </div>
            <div className="flex items-center gap-2">
              <Clock className="w-3.5 h-3.5 text-gray-400" />
              <span className="text-gray-500 w-[48px]">更新于：</span>
              <span className="text-gray-700">{resume.updateTime}</span>
              <span className="mx-1 text-gray-300">|</span>
              <span className="text-gray-700">{resume.projects} 个项目</span>
            </div>
            {resume.analysisCount > 0 && (
              <div className="flex items-center gap-2 ml-5">
                <span className="text-gray-500 w-[48px]">分析记录：</span>
                <span className="text-gray-700">最近 {resume.analysisCount} 次</span>
              </div>
            )}
          </div>

          <div className="flex items-center gap-2 pt-3 border-t border-gray-100">
            <button className="flex-1 flex justify-center items-center gap-1.5 py-1.5 text-[12px] text-gray-600 hover:text-gray-900 hover:bg-gray-50 rounded transition-colors">
              <Eye className="w-3.5 h-3.5" /> 查看
            </button>
            {!resume.isActive && (
              <button className="flex-1 flex justify-center items-center gap-1.5 py-1.5 text-[12px] text-[#0052D9] hover:bg-blue-50 rounded transition-colors">
                <CheckCircle2 className="w-3.5 h-3.5" /> 设为当前
              </button>
            )}
            <button className="flex-1 flex justify-center items-center gap-1.5 py-1.5 text-[12px] text-gray-600 hover:text-gray-900 hover:bg-gray-50 rounded transition-colors">
              <Edit3 className="w-3.5 h-3.5" /> 编辑信息
            </button>
          </div>
        </div>
      ))}
    </aside>
  );
}