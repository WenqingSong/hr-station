// ─── AIThinkingLoading — AI 生成中加载态 ─────────────────────────────────
//
// 用途：四核心 AI 能力（推荐/匹配/优化/对比）生成中展示。
// 视觉：白底卡片、rounded-2xl、ring-1 ring-blue-100、蓝色 spinner + 步骤。
// K8-5-37-B：Loading 设计冻结 + 统一 Loading 组件体系实现

import { Loader2 } from "lucide-react";

export type AIThinkingLoadingProps = {
  title: string;
  description?: string;
  steps?: string[];
};

export function AIThinkingLoading({
  title,
  description,
  steps,
}: AIThinkingLoadingProps) {
  return (
    <div className="max-w-[1440px] mx-auto px-6 py-12">
      <div className="max-w-lg mx-auto bg-white rounded-2xl ring-1 ring-blue-100 shadow-sm p-8">
        {/* 顶部蓝色 spinner */}
        <div className="flex items-center justify-center mb-5">
          <div className="w-10 h-10 rounded-full bg-blue-50 flex items-center justify-center">
            <Loader2 className="w-5 h-5 text-[#0052D9] animate-spin" />
          </div>
        </div>

        {/* title */}
        <h3 className="text-[16px] font-semibold text-slate-800 text-center mb-2">
          {title}
        </h3>

        {/* description */}
        {description && (
          <p className="text-[13px] text-gray-400 text-center leading-relaxed mb-5">
            {description}
          </p>
        )}

        {/* steps（最多 3 条） */}
        {steps && steps.length > 0 && (
          <div className="space-y-2.5">
            {steps.slice(0, 3).map((step, i) => (
              <div key={i} className="flex items-center gap-3">
                {/* 小圆点 */}
                <span className="relative flex h-2 w-2 shrink-0">
                  <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-blue-400 opacity-75" />
                  <span className="relative inline-flex rounded-full h-2 w-2 bg-[#0052D9]" />
                </span>
                <span className="text-[13px] text-gray-500">{step}</span>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
