// ─── AppBootLoading — 应用启动 / Auth Bootstrap 加载态 ──────────────────
//
// 用途：应用启动、首屏登录态校验期间展示。
// 视觉：全屏居中、浅灰蓝背景、Offerr鹅品牌标识 + 轻量 spinner。
// K8-5-37-B：Loading 设计冻结 + 统一 Loading 组件体系实现

import { Loader2 } from "lucide-react";

export function AppBootLoading() {
  return (
    <div className="fixed inset-0 z-[99999] flex flex-col items-center justify-center bg-[#F5F7FB]">
      {/* Offerr鹅 Logo 文本 */}
      <div className="mb-6">
        <span className="text-[28px] font-bold tracking-tight text-slate-800">
          Offerr<span className="text-[#0052D9]">鹅</span>
        </span>
      </div>

      {/* 轻量 spinner */}
      <Loader2 className="w-6 h-6 text-[#0052D9] animate-spin mb-4" />

      {/* 文案 */}
      <p className="text-sm text-gray-400">正在准备 Offerr鹅…</p>
    </div>
  );
}
