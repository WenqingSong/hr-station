/**
 * K9-Final-UX-C: Profile label humanize utilities.
 * Translates internal engineering terms (backend_engineering, etc.)
 * into user-facing natural language (后端开发相关经历, etc.).
 */

const profileLabelMap: Record<string, string> = {
  backend_engineering: "后端开发相关经历",
  career_exploration: "求职目标待聚焦",
  algorithm_research: "算法 / 数据分析相关经历",
  product_operation: "产品 / 运营相关经历",
  business_solution: "商务 / 解决方案相关经历",
};

/**
 * Convert internal profile enum value to user-visible natural language.
 * Returns "简历信息不足" for unknown / empty values.
 */
export function humanizeProfile(value?: string | null): string {
  if (!value || value.trim() === "") return "简历信息不足";
  const mapped = profileLabelMap[value.trim()];
  return mapped || "简历信息不足";
}

/**
 * Generate a natural-language sentence describing profile change between two resumes.
 *
 * Examples:
 *   humanizeProfileChange("career_exploration", "backend_engineering")
 *   → '简历定位从"求职目标待聚焦"转向"后端开发相关经历"，表达更聚焦。'
 *
 *   humanizeProfileChange("backend_engineering", "backend_engineering")
 *   → '两版简历定位基本一致，差异主要体现在内容深度和表达方式。'
 */
export function humanizeProfileChange(from?: string | null, to?: string | null): string {
  const fromLabel = humanizeProfile(from);
  const toLabel = humanizeProfile(to);

  if (from === to) {
    return "两版简历定位基本一致，差异主要体现在内容深度和表达方式。";
  }
  return `简历定位从"${fromLabel}"转向"${toLabel}"，表达更聚焦。`;
}
