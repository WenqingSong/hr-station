// ─── RouteProgressBar — 页面跳转顶部进度条 ──────────────────────────────
//
// 用途：页面路由切换时，顶部 2px 蓝色进度条轻量反馈。
// 实现：监听 useLocation() 的 pathname + search 变化，
//       变化时触发 CSS transition，150-250ms 内完成。
// K8-5-37-B：Loading 设计冻结 + 统一 Loading 组件体系实现

import { useState, useEffect, useRef } from "react";
import { useLocation } from "react-router";

export function RouteProgressBar() {
  const location = useLocation();
  const [progress, setProgress] = useState(0);
  const [visible, setVisible] = useState(false);
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    // 位置变化时启动进度条
    setVisible(true);
    setProgress(0);

    // 分两段：快速到 70%，再慢速到 100%
    timerRef.current = setTimeout(() => {
      setProgress(70);
    }, 30);

    timerRef.current = setTimeout(() => {
      setProgress(100);
    }, 180);

    // 完成后隐藏（等 transition 结束）
    timerRef.current = setTimeout(() => {
      setVisible(false);
      setProgress(0);
    }, 380);

    return () => {
      if (timerRef.current) clearTimeout(timerRef.current);
    };
  }, [location.pathname, location.search]);

  return (
    <div
      className="fixed top-0 left-0 right-0 z-[9999] h-[2px] pointer-events-none"
      style={{ opacity: visible ? 0.8 : 0, transition: "opacity 0.15s ease" }}
    >
      <div
        className="h-full bg-[#0052D9]"
        style={{
          width: `${progress}%`,
          transition: progress === 100 ? "width 0.2s ease-out" : "width 0.15s ease-out",
        }}
      />
    </div>
  );
}
