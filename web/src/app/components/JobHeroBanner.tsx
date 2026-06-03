import React from "react";
import { Search, Sparkles } from "lucide-react";

/**
 * K8-4-4-R12：岗位页顶部 Hero Banner（单图静态版，优雅蓝色系）
 *
 * 设计说明：
 * - 蓝色系渐变背景，三层视觉层次
 * - 左侧文案区：主标题 + 副标题 + 辅助说明
 * - 右侧抽象几何装饰：光斑、圆环、柔和图形
 * - 整体高度约 260px，圆角，克制阴影
 * - 轻微淡入动画，无连续动画
 */
export function JobHeroBanner() {
  return (
    <section
      className="relative w-full overflow-hidden rounded-2xl mb-6 animate-hero-fade-in"
      style={{ height: "260px" }}
    >
      {/* ── 第 1 层：背景渐变 ── */}
      <div className="absolute inset-0 bg-gradient-to-br from-[#0052D9] via-[#0066FF] to-[#0040A8]" />

      {/* ── 第 2 层：装饰光斑与几何图形 ── */}
      {/* 主光斑 - 右侧 */}
      <div className="absolute -top-20 -right-20 w-[500px] h-[500px] rounded-full bg-white/5 blur-3xl pointer-events-none" />
      {/* 次光斑 - 左下 */}
      <div className="absolute -bottom-32 -left-32 w-[400px] h-[400px] rounded-full bg-[#00B0F0]/10 blur-3xl pointer-events-none" />
      {/* 小光斑 - 中右 */}
      <div className="absolute top-10 right-1/3 w-[200px] h-[200px] rounded-full bg-white/5 blur-2xl pointer-events-none" />

      {/* 抽象几何：圆环 */}
      <div className="absolute top-1/2 right-[12%] -translate-y-1/2 w-[180px] h-[180px] border border-white/10 rounded-full pointer-events-none" />
      <div className="absolute top-1/2 right-[12%] -translate-y-1/2 w-[140px] h-[140px] border border-white/5 rounded-full pointer-events-none" />

      {/* 抽象几何：小圆点装饰 */}
      <div className="absolute top-8 right-[8%] w-2 h-2 bg-white/20 rounded-full pointer-events-none" />
      <div className="absolute top-16 right-[15%] w-1.5 h-1.5 bg-white/15 rounded-full pointer-events-none" />
      <div className="absolute bottom-12 right-[20%] w-1 h-1 bg-white/25 rounded-full pointer-events-none" />

      {/* 网格纹理叠加（极淡） */}
      <div
        className="absolute inset-0 pointer-events-none opacity-[0.03]"
        style={{
          backgroundImage: `linear-gradient(rgba(255,255,255,0.1) 1px, transparent 1px),
                           linear-gradient(90deg, rgba(255,255,255,0.1) 1px, transparent 1px)`,
          backgroundSize: "40px 40px",
        }}
      />

      {/* ── 第 3 层：内容区 ── */}
      <div className="relative z-10 h-full mx-auto max-w-[1440px] px-8 md:px-12 flex items-center justify-between">
        {/* 左侧文案 */}
        <div className="max-w-[600px]">
          {/* 轻量标签 */}
          <div className="inline-flex items-center gap-1.5 mb-4 px-3 py-1 rounded-full bg-white/10 border border-white/10">
            <Sparkles className="w-3.5 h-3.5 text-blue-200" />
            <span className="text-[12px] font-medium text-blue-100 tracking-wide">
              AI 智能匹配 · 校园招聘
            </span>
          </div>

          {/* 主标题 */}
          <h1 className="text-[32px] md:text-[36px] font-bold text-white leading-tight mb-3">
            Offer 捕手
          </h1>

          {/* 副标题 */}
          <p className="text-[18px] md:text-[20px] text-blue-100 font-medium mb-2">
            找到更适合你的实习岗位
          </p>

          {/* 辅助说明 */}
          <p className="text-[14px] text-blue-200/80 leading-relaxed max-w-[480px]">
            基于岗位信息与简历智能匹配，帮助你更高效地完成筛选与投递决策
          </p>
        </div>

        {/* 右侧视觉元素：抽象图形组合 */}
        <div className="hidden lg:flex items-center justify-center shrink-0 ml-8">
          <div className="relative w-[220px] h-[180px]">
            {/* 主卡片 - 模拟岗位卡片悬浮效果 */}
            <div className="absolute top-4 right-4 w-[160px] h-[100px] bg-white/10 backdrop-blur-sm rounded-xl border border-white/15 p-4 shadow-lg">
              <div className="w-3/4 h-3 bg-white/20 rounded-full mb-2.5" />
              <div className="w-1/2 h-2.5 bg-white/10 rounded-full mb-3" />
              <div className="flex gap-1.5">
                <div className="w-12 h-5 bg-[#0052D9]/40 border border-white/10 rounded-md" />
                <div className="w-16 h-5 bg-white/5 border border-white/10 rounded-md" />
              </div>
            </div>

            {/* 装饰圆 */}
            <div className="absolute bottom-2 left-2 w-12 h-12 bg-[#00B0F0]/20 rounded-full blur-sm" />
            <div className="absolute top-0 left-8 w-6 h-6 bg-white/10 rounded-full" />
          </div>
        </div>
      </div>

      {/* 底部微光边缘 */}
      <div className="absolute bottom-0 left-0 right-0 h-px bg-gradient-to-r from-transparent via-white/20 to-transparent" />
    </section>
  );
}
