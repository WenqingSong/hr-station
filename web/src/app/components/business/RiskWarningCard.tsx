import { AlertOctagon } from "lucide-react";

interface RiskWarningCardProps {
  risks: string[];
}

export function RiskWarningCard({ risks }: RiskWarningCardProps) {
  if (!risks.length) return null;

  return (
    <div className="bg-red-50 border border-red-100 rounded-lg p-4">
      <div className="flex items-center gap-2 mb-3">
        <AlertOctagon className="w-4 h-4 text-red-600" />
        <span className="font-medium text-sm text-red-800">高风险提示</span>
      </div>
      <ul className="space-y-2">
        {risks.map((risk, i) => (
          <li key={i} className="text-xs text-red-700 flex items-start gap-2">
            <span className="mt-1 w-1 h-1 bg-red-400 rounded-full shrink-0" />
            <span className="leading-relaxed">{risk}</span>
          </li>
        ))}
      </ul>
    </div>
  );
}
