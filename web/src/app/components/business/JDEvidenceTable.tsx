import { AlertTriangle, CheckCircle, HelpCircle } from "lucide-react";
import { cn } from "../../utils/cn";

interface EvidenceRow {
  jdReq: string;
  evidence: string;
  strength: "strong" | "weak" | "missing";
  risk?: string;
}

export function JDEvidenceTable({ data }: { data: EvidenceRow[] }) {
  const getStrengthConfig = (s: EvidenceRow["strength"]) => {
    switch (s) {
      case "strong": return { icon: CheckCircle, color: "text-green-600", bg: "bg-green-50", label: "强" };
      case "weak": return { icon: HelpCircle, color: "text-orange-600", bg: "bg-orange-50", label: "弱" };
      case "missing": return { icon: AlertTriangle, color: "text-red-600", bg: "bg-red-50", label: "无" };
    }
  };

  return (
    <div className="bg-white border border-slate-200 rounded-lg overflow-hidden">
      <div className="px-4 py-3 border-b border-slate-100 bg-slate-50">
        <div className="font-medium text-sm text-slate-800">JD 证据链分析</div>
      </div>
      <div className="overflow-x-auto">
        <table className="w-full text-left text-sm">
          <thead className="bg-slate-50 text-slate-500 text-xs border-b border-slate-200">
            <tr>
              <th className="px-4 py-2 font-medium w-1/4">JD 要求</th>
              <th className="px-4 py-2 font-medium w-1/3">简历证据</th>
              <th className="px-4 py-2 font-medium w-1/6">强度</th>
              <th className="px-4 py-2 font-medium w-1/4">风险说明</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-100">
            {data.map((row, i) => {
              const conf = getStrengthConfig(row.strength);
              const Icon = conf.icon;
              return (
                <tr key={i} className="hover:bg-slate-50/50 transition-colors">
                  <td className="px-4 py-3 text-slate-700 align-top">{row.jdReq}</td>
                  <td className="px-4 py-3 text-slate-600 align-top">
                    {row.evidence || <span className="text-slate-400 italic">无相关表述</span>}
                  </td>
                  <td className="px-4 py-3 align-top">
                    <div className={cn("inline-flex items-center gap-1 px-2 py-0.5 rounded text-xs", conf.bg, conf.color)}>
                      <Icon className="w-3 h-3" />
                      {conf.label}
                    </div>
                  </td>
                  <td className="px-4 py-3 text-slate-500 text-xs align-top">
                    {row.risk || "-"}
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
    </div>
  );
}
