import { Flag, Play } from "lucide-react";
import { Button } from "../ui/Button";

type SignalType = "ready" | "fix_needed" | "insufficient";

interface OptimizationStopSignalCardProps {
  type: SignalType;
}

export function OptimizationStopSignalCard({ type }: OptimizationStopSignalCardProps) {
  const config = {
    ready: {
      title: "可以停止优化，进入投递",
      desc: "已达到该岗位的最优呈现状态",
      buttonText: "一键投递",
      variant: "primary" as const,
    },
    fix_needed: {
      title: "修复 Top 3 后再投递",
      desc: "解决高优问题可显著提升通过率",
      buttonText: "去修复",
      variant: "secondary" as const,
    },
    insufficient: {
      title: "当前证据不足，先补充经历",
      desc: "缺乏核心项目支撑，AI 无法继续优化",
      buttonText: "补充经历",
      variant: "outline" as const,
    },
  };

  const { title, desc, buttonText, variant } = config[type];

  return (
    <div className="bg-white border border-slate-200 rounded-lg p-4">
      <div className="flex items-center gap-2 mb-2">
        <Flag className="w-4 h-4 text-blue-600" />
        <span className="font-medium text-sm text-slate-800">阶段建议</span>
      </div>
      <div className="text-sm font-semibold text-slate-900 mb-1">{title}</div>
      <div className="text-xs text-slate-500 mb-4">{desc}</div>
      <Button variant={variant} className="w-full gap-1.5" size="sm">
        {type === "ready" && <Play className="w-3.5 h-3.5" />}
        {buttonText}
      </Button>
    </div>
  );
}
