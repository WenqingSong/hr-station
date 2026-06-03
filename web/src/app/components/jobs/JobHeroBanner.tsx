import React from "react";

/**
 * K8-5-5：岗位页顶部 Hero Banner（用户生成图片版）
 *
 * 设计说明：
 * - 全宽横向大横幅，使用用户提供的生成图片
 * - 移除旧版 CSS 渐变、文字叠加层、抽象图形层
 * - 图片使用 object-cover 横向铺满，中心聚焦主视觉
 * - 响应式高度：桌面 320px~360px，中屏 240px，小屏 180px
 * - 非 clickable，无跳转
 */
export function JobHeroBanner() {
  return (
    <section className="relative w-full overflow-hidden bg-[#0052D9]">
      <img
        src="/images/offer-catcher-banner.png"
        alt="Offer 捕手 2026 实习生招聘 Banner"
        className="block w-full h-[180px] sm:h-[240px] lg:h-[320px] xl:h-[360px] object-cover object-center"
        draggable={false}
      />
    </section>
  );
}
