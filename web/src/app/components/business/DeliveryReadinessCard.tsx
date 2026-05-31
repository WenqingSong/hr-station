import { CheckCircle2, AlertTriangle, XCircle } from "lucide-react";
import { cn } from "../../utils/cn";

type ReadinessStatus = "ready" | "fix_first" | "not_recommended";

interface DeliveryReadinessCardProps {
  status: ReadinessStatus;
}

export function DeliveryReadinessCard({ status }: DeliveryReadinessCardProps) {
  const config = {
    ready: {
      icon: CheckCircle2,
      title: "建议投递",
      desc: "当前简历与岗位匹配度较高，各项核心要求均已满足。",
      styles: "bg-green-50 border-green-200",
      iconStyles: "text-green-600",
      titleStyles: "text-green-800",
    },
    fix_first: {
      icon: AlertTriangle,
      title: "先修复后投递",
      desc: "存在关键信息缺失或表述不准，建议优先修复 Top 3 问题。",
      styles: "bg-orange-50 border-orange-200",
      iconStyles: "text-orange-600",
      titleStyles: "text-orange-800",
    },
    not_recommended: {
      icon: XCircle,
      title: "暂不建议投递",
      desc: "核心技能或经验严重不匹配，投递风险较高。",
      styles: "bg-red-50 border-red-200",
      iconStyles: "text-red-600",
      titleStyles: "text-red-800",
    },
  };

  const { icon: Icon, title, desc, styles, iconStyles, titleStyles } = config[status];

  return (
    <div className={cn("p-4 rounded-lg border", styles)}>
      <div className="flex items-start gap-3">
        <Icon className={cn("w-5 h-5 mt-0.5", iconStyles)} />
        <div>
          <div className={cn("font-medium text-sm mb-1", titleStyles)}>{title}</div>
          <div className="text-xs text-slate-600 leading-relaxed">{desc}</div>
        </div>
      </div>
    </div>
  );
}
