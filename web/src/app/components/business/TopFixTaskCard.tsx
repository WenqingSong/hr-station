import { CheckCircle2, ChevronRight, AlertCircle } from "lucide-react";
import { Tag } from "../ui/Tag";

interface Task {
  id: string;
  title: string;
  priority: "high" | "medium";
  impact: string;
  needsVerification: boolean;
}

export function TopFixTaskCard({ tasks }: { tasks: Task[] }) {
  return (
    <div className="bg-white border border-slate-200 rounded-lg overflow-hidden">
      <div className="px-4 py-3 border-b border-slate-100 bg-slate-50 flex items-center justify-between">
        <div className="font-medium text-sm text-slate-800">Top 3 快速修复任务</div>
        <span className="text-xs text-slate-500">{tasks.length} 项待办</span>
      </div>
      <div className="divide-y divide-slate-100">
        {tasks.map((task, idx) => (
          <div key={task.id} className="p-4 hover:bg-slate-50 transition-colors group cursor-pointer">
            <div className="flex items-start gap-3">
              <div className="mt-0.5 text-slate-300 group-hover:text-blue-600 transition-colors">
                <CheckCircle2 className="w-4 h-4" />
              </div>
              <div className="flex-1">
                <div className="flex items-center gap-2 mb-1.5">
                  <span className="text-sm font-medium text-slate-800">{task.title}</span>
                  {task.needsVerification && (
                    <div className="flex items-center gap-1 text-[10px] text-orange-600 bg-orange-50 px-1.5 py-0.5 rounded border border-orange-100">
                      <AlertCircle className="w-3 h-3" />
                      需核实
                    </div>
                  )}
                </div>
                <div className="flex items-center gap-2 text-xs">
                  <Tag variant={task.priority === "high" ? "blue" : "default"}>
                    {task.priority === "high" ? "高优" : "中优"}
                  </Tag>
                  <span className="text-slate-500">预计影响: {task.impact}</span>
                </div>
              </div>
              <ChevronRight className="w-4 h-4 text-slate-300 mt-2 group-hover:text-blue-500 transition-colors" />
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
