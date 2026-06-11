/**
 * K13-Stability-A: 安全数据访问工具
 *
 * 提供统一的 null/undefined 兜底，防止前端因后端返回字段不兼容而崩溃。
 * 四个核心功能的 normalizer 层应优先处理字段映射，
 * 这些工具作为页面渲染时的最终防线。
 */

/** 将可能为 null/undefined 的值转为安全数组 */
export function safeArray<T>(v: T[] | null | undefined): T[] {
  return Array.isArray(v) ? v : [];
}

/** 将可能为非字符串的值转为安全字符串 */
export function safeString(v: unknown, fallback = ""): string {
  return typeof v === "string" ? v : fallback;
}

/** 将可能为非数字的值转为安全数字 */
export function safeNumber(v: unknown, fallback = 0): number {
  if (typeof v === "number" && !Number.isNaN(v)) return v;
  return fallback;
}

/** 安全解析 JSON 字符串，失败返回 null，不抛出异常 */
export function safeJsonParse<T = unknown>(v: unknown, fallback: T | null = null): T | null {
  if (typeof v !== "string") return fallback;
  try {
    return JSON.parse(v) as T;
  } catch {
    return fallback;
  }
}

/** 检查对象是否是 truthy 值（null/undefined 保护） */
export function isDefined<T>(v: T | null | undefined): v is T {
  return v !== null && v !== undefined;
}
