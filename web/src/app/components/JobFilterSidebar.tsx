import React from "react";
import { Search, ChevronDown, ChevronUp } from "lucide-react";
import { Checkbox } from "./ui/Checkbox";

const FILTER_GROUPS = [
  {
    title: "求职方向",
    options: [
      { label: "软件开发", count: 342, checked: true },
      { label: "算法与大数据", count: 128 },
      { label: "产品策划 / 运营", count: 89 },
      { label: "设计", count: 45 },
    ],
  },
  {
    title: "工作地点",
    options: [
      { label: "深圳", count: 215, checked: true },
      { label: "北京", count: 180 },
      { label: "上海", count: 156 },
      { label: "广州", count: 98 },
      { label: "成都", count: 64 },
    ],
  },
  {
    title: "招聘类型",
    options: [
      { label: "校园招聘", count: 420, checked: true },
      { label: "实习生", count: 185 },
      { label: "提前批", count: 0 },
    ],
  },
  {
    title: "岗位类别",
    options: [
      { label: "后台开发", count: 156 },
      { label: "Go 语言开发", count: 42 },
      { label: "云计算开发", count: 89 },
      { label: "数据开发", count: 64 },
      { label: "游戏开发", count: 112 },
    ],
  },
];

export function JobFilterSidebar() {
  return (
    <aside className="w-[260px] shrink-0 sticky top-24 self-start bg-white border border-gray-200 p-4 rounded-md">
      <div className="relative mb-6">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-400" />
        <input
          type="text"
          placeholder="搜索岗位 / 技能关键词"
          className="w-full pl-9 pr-3 py-2 text-sm border border-gray-300 rounded-md focus:outline-none focus:ring-1 focus:ring-[#0052D9] focus:border-[#0052D9] placeholder:text-gray-400"
        />
      </div>

      <div className="space-y-6">
        {FILTER_GROUPS.map((group, index) => (
          <div key={index}>
            <div className="flex items-center justify-between mb-3 cursor-pointer group">
              <h3 className="text-sm font-bold text-gray-900">{group.title}</h3>
              <ChevronUp className="h-4 w-4 text-gray-400 group-hover:text-gray-600" />
            </div>
            <div className="space-y-3">
              {group.options.map((option, idx) => (
                <div key={idx} className="flex items-center justify-between group cursor-pointer">
                  <div className="flex items-center gap-2">
                    <Checkbox id={`${group.title}-${idx}`} checked={option.checked} />
                    <label
                      htmlFor={`${group.title}-${idx}`}
                      className="text-sm text-gray-600 cursor-pointer group-hover:text-[#0052D9]"
                    >
                      {option.label}
                    </label>
                  </div>
                  <span className="text-xs text-gray-400">{option.count > 0 ? option.count : ''}</span>
                </div>
              ))}
            </div>
          </div>
        ))}
      </div>
    </aside>
  );
}
