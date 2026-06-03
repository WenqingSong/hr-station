import React from "react";
import { Sparkles, ArrowRight, FileText } from "lucide-react";
import { Link } from "react-router";

/**
 * K8-5-11-1：首页代码绘制 Hero Banner（替换图片版）
 *
 * 设计说明：
 * - 蓝色科技感渐变背景 + 光斑 + 代码绘制装饰
 * - 左侧：标签 + 主标题 + 副标题 + 描述 + CTA 按钮
 * - 右侧：CSS 绘制抽象插画（简历卡片 + 匹配度标签 + 光晕）
 * - 全宽无圆角，紧贴左右和顶部
 * - 无图片引用
 */
export function JobHeroBanner() {
  return (
    <section className="relative w-full overflow-hidden bg-[#001a4d]">
      {/* 背景渐变 + 径向光 */}
      <div className="absolute inset-0 bg-[radial-gradient(circle_at_30%_20%,#38bdf8_0%,transparent_28%),linear-gradient(120deg,#0052d9_0%,#0866ff_45%,#003a9f_100%)]" />

      {/* 光斑装饰 */}
      <div className="absolute -top-32 -right-24 h-[500px] w-[500px] rounded-full bg-[#0066FF]/25 blur-3xl pointer-events-none" />
      <div className="absolute -bottom-40 -left-32 h-[400px] w-[400px] rounded-full bg-[#00B0F0]/15 blur-3xl pointer-events-none" />
      <div className="absolute bottom-10 right-1/4 h-[200px] w-[200px] rounded-full bg-white/5 blur-2xl pointer-events-none" />

      {/* 内容区 */}
      <div className="relative z-10 mx-auto flex max-w-[1440px] flex-col gap-8 px-5 py-10 sm:px-8 lg:min-h-[340px] lg:flex-row lg:items-center lg:px-12">
        
        {/* ── 左侧文案 ── */}
        <div className="flex-1 max-w-2xl">
          {/* 标签胶囊 */}
          <div className="mb-4 inline-flex items-center gap-1.5 rounded-full border border-white/15 bg-white/10 px-3 py-1.5 backdrop-blur-sm">
            <Sparkles className="h-3.5 w-3.5 text-blue-200" />
            <span className="text-[12px] font-medium tracking-wide text-blue-100">
              AI 智能匹配 · 校园招聘
            </span>
          </div>

          {/* 主标题 */}
          <h1
            className="text-[34px] font-extrabold leading-[1.15] tracking-[-0.02em] text-white sm:text-[44px] lg:text-[52px]"
            style={{
              fontFamily:
                '"Alimama ShuHeiTi", "HarmonyOS Sans SC", "MiSans", "Source Han Sans SC", "Noto Sans SC", "Inter", system-ui, sans-serif',
            }}
          >
            Offerr鹅 2026 实习生招聘
          </h1>

          {/* 副标题 */}
          <p className="mt-3 text-[17px] font-semibold text-blue-100 sm:text-[20px]">
            找到更适合你的实习岗位，开启你的职场未来
          </p>

          {/* 辅助描述 */}
          <p className="mt-2 max-w-lg text-[14px] leading-relaxed text-blue-200/80">
            基于岗位信息与简历智能匹配，帮助你更高效完成岗位筛选与投递决策
          </p>

          {/* CTA 按钮组 */}
          <div className="mt-6 flex flex-wrap gap-3">
            <Link
              to="/jobs"
              className="inline-flex items-center gap-2 rounded-full bg-white px-6 py-2.5 text-sm font-semibold text-[#0052D9] shadow-sm transition-colors hover:bg-blue-50"
            >
              查看岗位
              <ArrowRight className="h-4 w-4" />
            </Link>
            <Link
              to="/resumes"
              className="inline-flex items-center gap-2 rounded-full border border-white/30 bg-white/8 px-6 py-2.5 text-sm font-semibold text-white backdrop-blur-sm transition-colors hover:bg-white/14"
            >
              <FileText className="h-4 w-4" />
              我的简历
            </Link>
          </div>
        </div>

        {/* ── 右侧 CSS 绘制插画 ── */}
        <div className="hidden flex-1 lg:flex items-center justify-center">
          <div className="relative h-[280px] w-full max-w-[360px]">
            {/* 简历卡片 */}
            <div className="absolute right-0 top-2 h-52 w-64 rounded-2xl border border-white/15 bg-white/8 backdrop-blur-xl p-5 shadow-xl">
              {/* 卡片内容线条 */}
              <div className="mb-3 h-4 w-3/4 rounded-full bg-white/25" />
              <div className="mb-6 h-3 w-1/2 rounded-full bg-white/15" />
              <div className="mb-2.5 h-3 w-full rounded-full bg-white/10" />
              <div className="mb-2.5 h-3 w-5/6 rounded-full bg-white/10" />
              <div className="h-3 w-2/3 rounded-full bg-white/10" />
            </div>

            {/* 匹配度徽章 */}
            <div className="absolute bottom-4 right-6 flex items-center gap-2 rounded-xl border border-white/20 bg-white/12 px-4 py-2.5 backdrop-blur-md shadow-lg">
              <div className="flex h-8 w-8 items-center justify-center rounded-full bg-[#0052D9]/50 text-[13px] font-bold text-white ring-2 ring-white/20">
                92
              </div>
              <div>
                <p className="text-[11px] font-semibold leading-tight text-white">匹配度</p>
                <p className="text-[10px] text-blue-200">优秀匹配</p>
              </div>
            </div>

            {/* 浮动装饰圆 */}
            <div className="absolute left-6 top-10 h-12 w-12 rounded-full bg-[#00B0F0]/25 blur-sm" />
            <div className="absolute left-0 bottom-16 h-6 w-6 rounded-full bg-white/15" />
            <div className="absolute right-20 top-0 h-3 w-3 rounded-full bg-white/25" />
          </div>
        </div>
      </div>
    </section>
  );
}
