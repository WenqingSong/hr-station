/**
 * Offer 捕手 — 全局设计 Token
 * 风格基底：腾讯校招岗位内页
 * 所有页面必须基于此文件中的值构建，不允许随意使用 hardcode 颜色。
 */

export const color = {
  // 背景
  pageBg: "#F5F7FA",
  cardBg: "#FFFFFF",
  hoverBg: "#F7F8FA",

  // 主色
  blue: "#1D6FFF",
  blueDark: "#1558CC",
  blueLight: "#EBF2FF",
  blueText: "#1D6FFF",

  // 文字层级
  textPrimary: "#1F2329",
  textSecondary: "#646A73",
  textTertiary: "#86909C",
  textDisabled: "#C9CDD4",

  // 边框
  border: "#E5E6EB",
  borderLight: "#F2F3F5",

  // 语义色
  success: "#00B42A",
  successBg: "#E8FFEA",
  successText: "#009A29",

  warn: "#FF7D00",
  warnBg: "#FFF7E8",
  warnText: "#CC6400",

  danger: "#F53F3F",
  dangerBg: "#FFECE8",
  dangerText: "#CB2634",

  // 分割线
  divider: "#F0F1F5",
} as const;

export const radius = {
  card: "12px",
  cardLg: "16px",
  btn: "6px",
  tag: "4px",
  tagRound: "20px",
} as const;

export const shadow = {
  card: "0 1px 4px rgba(0,0,0,0.06)",
  cardHover: "0 4px 16px rgba(0,0,60,0.10)",
  nav: "0 1px 0 #E5E6EB",
} as const;

export const layout = {
  navHeight: "56px",
  contentMaxWidth: "1180px",
  contentPadding: "0 24px",
  sectionGap: "20px",
  cardPadding: "20px 24px",
  cardPaddingLg: "24px 28px",
} as const;

export const fontSize = {
  pageTitle: "20px",
  sectionTitle: "16px",
  cardTitle: "15px",
  body: "14px",
  label: "13px",
  small: "12px",
} as const;
