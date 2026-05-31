import React, { useState } from "react";
import { Search, MapPin } from "lucide-react";
import { Button } from "./ui/Button";
import { Badge } from "./ui/Badge";
import { JobCard } from "./JobCard";
import { mockJobs } from "../mocks";
import { toJobCardViewModels } from "../adapters";

const jobCards = toJobCardViewModels(mockJobs);

export function JobList() {
  const [selectedJobId, setSelectedJobId] = useState<string>("job_backend_001");

  return (
    <main className="flex-1 max-w-[820px] pb-24">
      {/* Header */}
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-900 mb-2">岗位列表</h1>
        <p className="text-gray-500">
          浏览当前开放岗位，结合当前简历，让 AI 帮你优先筛选最值得关注的岗位。
        </p>
      </div>

      {/* Search Area */}
      <div className="flex gap-3 mb-6">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-5 w-5 text-gray-400" />
          <input
            type="text"
            placeholder="岗位关键词搜索"
            className="w-full pl-10 pr-4 h-11 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-[#0052D9] focus:border-[#0052D9]"
          />
        </div>
        <div className="relative w-40 shrink-0">
          <MapPin className="absolute left-3 top-1/2 -translate-y-1/2 h-5 w-5 text-gray-400" />
          <select className="w-full pl-10 pr-8 h-11 border border-gray-300 rounded-md appearance-none bg-white focus:outline-none focus:ring-2 focus:ring-[#0052D9] focus:border-[#0052D9]">
            <option>全国</option>
            <option>深圳</option>
            <option>北京</option>
            <option>上海</option>
            <option>广州</option>
          </select>
        </div>
        <Button className="h-11 px-8">搜索</Button>
      </div>

      {/* Current Filters */}
      <div className="flex items-center gap-2 mb-6">
        <span className="text-sm text-gray-500 mr-2">当前筛选：</span>
        <Badge variant="blue" className="px-3 py-1 bg-blue-50/50">2026 校园招聘 ✕</Badge>
        <Badge variant="blue" className="px-3 py-1 bg-blue-50/50">软件开发 ✕</Badge>
        <Badge variant="blue" className="px-3 py-1 bg-blue-50/50">深圳 ✕</Badge>
        <button className="text-sm text-gray-400 hover:text-gray-600 ml-2">清空全部</button>
      </div>

      {/* Job Cards List */}
      <div className="space-y-4">
        {jobCards.map((job) => (
          <JobCard
            key={job.id}
            id={job.id}
            title={job.title}
            type={job.typeLabel}
            department={job.businessGroup}
            skills={job.tags}
            location={job.locationsText}
            description={job.shortDescription}
            selected={selectedJobId === job.id}
            onSelect={setSelectedJobId}
          />
        ))}
      </div>
    </main>
  );
}
