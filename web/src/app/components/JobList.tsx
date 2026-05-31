import React, { useState } from "react";
import { Search, MapPin } from "lucide-react";
import { Button } from "./ui/Button";
import { Badge } from "./ui/Badge";
import { JobCard } from "./JobCard";

const JOB_MOCKS = [
  {
    id: "job-1",
    title: "软件开发-后台开发方向",
    type: "技术",
    department: "PCG",
    skills: ["Go", "C/C++", "MySQL", "Redis", "微服务"],
    location: "深圳",
    description: "负责业务后台服务器的设计、开发、优化等工作；负责后台系统的高并发、高可用及海量数据处理的架构设计与实现；参与制定和审核技术方案。",
  },
  {
    id: "job-2",
    title: "云计算开发实习生",
    type: "实习",
    department: "CSIG",
    skills: ["Kubernetes", "Docker", "Golang", "Linux"],
    location: "深圳",
    description: "参与公有云底层IaaS/PaaS管控系统开发；参与云原生架构设计与微服务改造；协助进行系统性能测试与调优。",
  },
  {
    id: "job-3",
    title: "算法-多模态方向",
    type: "技术",
    department: "TEG",
    skills: ["NLP", "CV", "PyTorch", "Deep Learning"],
    location: "北京",
    description: "负责多模态大模型的基础研发与落地应用；探索图文音等多模态数据的对齐与融合；推进前沿算法在具体业务场景中的实际效果提升。",
  },
  {
    id: "job-4",
    title: "软件开发-PC 客户端开发方向",
    type: "应届毕业生",
    department: "WXG",
    skills: ["C++", "Qt", "Windows/Mac API", "多线程"],
    location: "广州",
    description: "参与微信PC端的功能开发与性能优化；负责跨平台客户端架构演进；解决各类复杂的客户端体验与稳定性问题。",
  },
  {
    id: "job-5",
    title: "数据开发实习生",
    type: "实习",
    department: "CDG",
    skills: ["Hadoop", "Spark", "Hive", "Python", "SQL"],
    location: "深圳",
    description: "参与海量数据处理流程的开发与维护；协助构建数据仓库底座与数据治理体系；支持业务部门的数据分析需求与报表开发。",
  },
  {
    id: "job-6",
    title: "游戏策划培训生",
    type: "应届毕业生",
    department: "IEG",
    skills: ["系统策划", "关卡策划", "文案策划", "数值策划"],
    location: "上海",
    description: "参与重点游戏项目的世界观构建、系统设计、玩法创新或数值体验打磨；与程序、美术团队紧密配合，跟进设计落地；体验完整的游戏研发周期。",
  },
];

export function JobList() {
  const [selectedJobId, setSelectedJobId] = useState<string>("job-1");

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
        {JOB_MOCKS.map((job) => (
          <JobCard
            key={job.id}
            {...job}
            selected={selectedJobId === job.id}
            onSelect={setSelectedJobId}
          />
        ))}
      </div>
    </main>
  );
}
