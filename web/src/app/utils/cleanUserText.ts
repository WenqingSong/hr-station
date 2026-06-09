/**
 * K9-Final-UX-C: Clean user-visible text by replacing internal engineering
 * terms with natural language.
 *
 * Rules:
 *   画像方向 → 简历定位
 *   画像 → 简历定位
 *   标签 → 关键词 (but NOT inside "关键词" itself)
 *   岗位标签 → 岗位关键词
 *   候选人标签 → 候选人能力点
 *   backend_engineering → 后端开发相关经历
 *   career_exploration → 求职目标待聚焦
 *   algorithm_research → 算法 / 数据分析相关经历
 *   product_operation → 产品 / 运营相关经历
 *   dominant_profile → 简历定位
 *   方向方向 → 方向
 *   mock_agent / mock_quality → (should not appear in UI; replace with empty)
 */

/** Ordered replacement pairs. Order matters — longer / more specific matches first. */
const replacementRules: Array<[RegExp, string]> = [
  // Specific multi-char terms first (before single-char overlaps)
  [/画像方向/g, "简历定位"],
  [/候选人标签/g, "候选人能力点"],
  [/岗位标签/g, "岗位关键词"],
  [/简历画像/g, "简历定位"],
  // Internal enum values
  [/backend_engineering/g, "后端开发相关经历"],
  [/career_exploration/g, "求职目标待聚焦"],
  [/algorithm_research/g, "算法 / 数据分析相关经历"],
  [/product_operation/g, "产品 / 运营相关经历"],
  [/dominant_profile/g, "简历定位"],
  // Generic single-char terms after specific ones
  [/画像/g, "简历定位"],
  [/标签/g, "关键词"],
  // Cleanup duplicates
  [/方向方向/g, "方向"],
  [/方向 方向/g, "方向"],
  // mock_agent / mock_quality should never be user-visible
  [/mock_agent/g, ""],
  [/mock_quality/g, ""],
];

/**
 * Clean user-visible text by replacing internal engineering terms.
 * Does NOT modify code variable names or API field names —
 * this function is only for text that will be rendered in the UI.
 */
export function cleanUserText(text?: string | null): string {
  if (!text) return "";
  let result = text;
  for (const [pattern, replacement] of replacementRules) {
    result = result.replace(pattern, replacement);
  }
  return result;
}
