import type { ReactNode } from "react";

/**
 * K8-5-2-B：统一两栏页面布局 Shell（不接入页面，仅创建组件）
 *
 * 职责：
 * - 左侧主内容 + 右侧 aside 的响应式两栏布局
 * - 不关心业务类型，只接收 children 和 aside
 *
 * 响应式规则（K8-5-3 调整）：
 * - xl+（≥1280px）：两列 grid，主内容 minmax(0,1fr)，右侧栏 340px
 * - <xl：单列，aside 放在 children 下方
 *
 * 注意：
 * - min-w-0 防止 grid 溢出
 * - 不在此组件内处理 sticky（由 AISidebarShell 负责）
 *
 * K8-5-9-B：新增 maxWidthClassName / asideWidthClassName 可选参数，允许详情页使用更宽布局。
 */
export interface TwoColumnPageShellProps {
  /** 左侧主内容 */
  children: ReactNode;
  /** 右侧辅助栏（可选） */
  aside?: ReactNode;
  /** 外层 wrapper 附加 className */
  className?: string;
  /** 主内容区附加 className */
  contentClassName?: string;
  /** 右侧栏附加 className */
  asideClassName?: string;
  /** 外层 max-width（默认 max-w-6xl，详情页建议 max-w-[1280px]） */
  maxWidthClassName?: string;
  /** grid columns 模板（默认 xl:grid-cols-[minmax(0,1fr)_340px]） */
  gridColumnsClassName?: string;
}

export function TwoColumnPageShell({
  children,
  aside,
  className = "",
  contentClassName = "",
  asideClassName = "",
  maxWidthClassName = "max-w-6xl",
  gridColumnsClassName = "xl:grid-cols-[minmax(0,1fr)_340px]",
}: TwoColumnPageShellProps) {
  return (
    <div className={`mx-auto w-full ${maxWidthClassName} px-4 sm:px-6 lg:px-8 overflow-x-hidden ${className}`}>
      <div className={`grid grid-cols-1 gap-6 ${gridColumnsClassName} xl:gap-8 items-start`}>
        {/* 左侧主内容 */}
        <div className={`min-w-0 ${contentClassName}`}>{children}</div>

        {/* 右侧辅助栏 */}
        {aside ? (
          <div className={`min-w-0 ${asideClassName}`}>{aside}</div>
        ) : null}
      </div>
    </div>
  );
}
