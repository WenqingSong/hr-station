// @deprecated — 5-3.1 已废弃。岗位筛选侧栏使用 ../JobFilterSidebar.tsx。
// 保留此文件仅用于历史追溯，后续清理阶段可安全删除。

import { Search, ChevronDown, Check } from "lucide-react";
import { useState } from "react";
import { cn } from "../../utils/cn";

export function Sidebar() {
  const [expanded, setExpanded] = useState<Record<string, boolean>>({
    direction: true,
    location: true,
  });

  const toggle = (key: string) => {
    setExpanded((prev) => ({ ...prev, [key]: !prev[key] }));
  };

  return (
    <aside className="w-[260px] shrink-0 sticky top-20">
      <div className="bg-white rounded-lg border border-slate-200 p-4 space-y-6">
        <div>
          <div className="relative">
            <Search className="w-4 h-4 text-slate-400 absolute left-3 top-1/2 -translate-y-1/2" />
            <input
              type="text"
              placeholder="搜索岗位/技能..."
              className="w-full pl-9 pr-3 py-2 bg-slate-50 border border-slate-200 rounded text-sm focus:outline-none focus:ring-1 focus:ring-blue-600 focus:border-blue-600 transition-shadow"
            />
          </div>
        </div>

        <FilterSection
          title="求职方向"
          isOpen={expanded.direction}
          onToggle={() => toggle("direction")}
        >
          <FilterItem label="软件开发" count={124} checked />
          <FilterItem label="算法与大数据" count={89} />
          <FilterItem label="产品策划/运营" count={45} />
          <FilterItem label="设计" count={12} />
        </FilterSection>

        <FilterSection
          title="工作地点"
          isOpen={expanded.location}
          onToggle={() => toggle("location")}
        >
          <FilterItem label="深圳" count={88} checked />
          <FilterItem label="北京" count={65} />
          <FilterItem label="上海" count={54} />
          <FilterItem label="广州" count={23} />
        </FilterSection>
      </div>
    </aside>
  );
}

function FilterSection({ title, isOpen, onToggle, children }: any) {
  return (
    <div>
      <button
        onClick={onToggle}
        className="flex items-center justify-between w-full text-left mb-3 group"
      >
        <span className="font-medium text-sm text-slate-800">{title}</span>
        <ChevronDown
          className={cn(
            "w-4 h-4 text-slate-400 transition-transform duration-200 group-hover:text-slate-600",
            isOpen && "rotate-180"
          )}
        />
      </button>
      {isOpen && <div className="space-y-2">{children}</div>}
    </div>
  );
}

function FilterItem({ label, count, checked = false }: { label: string; count: number; checked?: boolean }) {
  return (
    <label className="flex items-center justify-between cursor-pointer group">
      <div className="flex items-center gap-2">
        <div
          className={cn(
            "w-4 h-4 rounded border flex items-center justify-center transition-colors",
            checked
              ? "bg-blue-600 border-blue-600"
              : "border-slate-300 bg-white group-hover:border-blue-400"
          )}
        >
          {checked && <Check className="w-3 h-3 text-white" />}
        </div>
        <span className="text-sm text-slate-600 group-hover:text-slate-900 transition-colors">
          {label}
        </span>
      </div>
      <span className="text-xs text-slate-400">{count}</span>
    </label>
  );
}
