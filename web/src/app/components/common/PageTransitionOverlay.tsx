/**
 * @deprecated K8-5-38-B 设计冻结后，普通页面跳转不再使用中央 Overlay。
 * 当前页面上方 overlay 已从 routes.tsx RootLayout 中移除。
 * 普通页面跳转统一使用 RouteProgressBar（顶部 2px 进度条）。
 * 本文件保留用于历史追溯，当前不启用。
 */

// ─── PageTransitionOverlay — 页面跳转整体 Loading Overlay ─────────────────
//
// 用途：页面路由切换时，用轻量遮罩覆盖主要内容区域，
//       避免"局部 skeleton / 新旧页面混杂"的割裂感。
//
// 设计（K8-5-37-D 精修）：
//   - 仅覆盖 main 内容区（flex-1 relative 容器内），保留 Navbar 清晰
//   - 遮罩：bg-slate-50/70 + 极轻 backdrop-blur-[1px]，不过度遮挡
//   - 中央：小型白色胶囊（rounded-full）+ 细线 spinner + "正在切换页面"
//   - 无大 Logo、无大 Spinner、无大面积蓝色
//   - SHOW_DELAY=90ms / MIN_VISIBLE=260ms，过渡更快更丝滑
//   - fade 过渡：opacity 0→1 (180ms ease-out)
//   - z-30：低于 Navbar(z-50)、RouteProgressBar(z-9999)、AppBootLoading(z-99999)
//
// K8-5-37-C：初始实现 | K8-5-37-D：视觉与动效专业精修

import { useRef, useState, useEffect } from "react";
import { useLocation } from "react-router";

const SHOW_DELAY = 90;   // 延迟 90ms后才显示，避免极快跳转闪一下
const MIN_VISIBLE = 260;  // 一旦显示至少保持 260ms，让用户感知但不拖沓

export function PageTransitionOverlay() {
  const location = useLocation();
  const [visible, setVisible] = useState(false);
  const isFirstMount = useRef(true);

  useEffect(() => {
    // 跳过首次 mount（首屏由 AppBootLoading / PageLoading 负责）
    if (isFirstMount.current) {
      isFirstMount.current = false;
      return;
    }

    let showTimer: ReturnType<typeof setTimeout> | undefined;
    let hideTimer: ReturnType<typeof setTimeout> | undefined;

    showTimer = setTimeout(() => {
      setVisible(true);
      hideTimer = setTimeout(() => {
        setVisible(false);
      }, MIN_VISIBLE);
    }, SHOW_DELAY);

    return () => {
      if (showTimer) clearTimeout(showTimer);
      if (hideTimer) clearTimeout(hideTimer);
    };
  }, [location.pathname, location.search]);

  return (
    <div
      className="absolute inset-0 z-30 flex items-center justify-center bg-slate-50/70 backdrop-blur-[1px] transition-opacity duration-[180ms] ease-out select-none pointer-events-none"
      style={{ opacity: visible ? 1 : 0 }}
    >
      {/* 小型白色 loading 胶囊 */}
      <div
        className="flex items-center gap-3 px-5 py-3 bg-white/90 rounded-full ring-1 ring-slate-200/80 shadow-[0_8px_30px_rgba(15,23,42,0.08)] transition-all duration-[180ms] ease-out pointer-events-auto"
        style={{
          opacity: visible ? 1 : 0,
          transform: visible ? "translateY(0) scale(1)" : "translateY(4px) scale(0.98)",
        }}
      >
        {/* 方案 A：细线 spinner */}
        <div className="h-4 w-4 animate-spin rounded-full border-2 border-blue-100 border-t-[#0052D9]" />

        {/* 文案 */}
        <span className="text-sm font-medium text-slate-700">
          正在切换页面
        </span>
      </div>
    </div>
  );
}

