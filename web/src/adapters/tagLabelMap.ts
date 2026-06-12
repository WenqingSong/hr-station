// ─── Tag Label Mapping for Fast Recommendation ──────────────────────────────
//
// 前端防御层：将后端可能返回的 raw tag code 转换为用户可读中文。
// 即使后端已修复，前端仍保留此映射作为安全兜底。

/** 原始 code → 中文 display label */
const TAG_LABEL_MAP: Record<string, string> = {
  // 方向标签
  research: "研究方向相关",
  algorithm: "算法能力",
  backend: "后端开发",
  frontend: "前端开发",
  data: "数据分析",
  product: "产品方向",
  ai_infra: "AI 基础设施",
  // 技能标签
  redis: "Redis 工程能力",
  golang: "Go 语言开发",
  python: "Python 开发",
  java: "Java 开发",
  cpp: "C/C++ 开发",
  mysql: "MySQL 数据库",
  linux: "Linux 系统",
  docker_k8s: "容器化与编排",
  pytorch_tf: "深度学习框架",
  // 要求标签
  degree_bachelor: "本科学历要求",
  degree_master: "硕士学历要求",
  internship_exp: "实习经历",
  project_exp: "项目经验",
  paper: "论文或学术成果",
  // 无意义占位符 → 过滤
  skill_0: "",
  skill_1: "",
  skill_2: "",
  skill_3: "",
};

/** 判断是否 look like a raw tag code（含下划线且全英文小写） */
export function isRawTagCode(value: string): boolean {
  if (!value || typeof value !== "string") return false;
  // 包含中文 = 已是可读文案
  if (/[\u4e00-\u9fff]/.test(value)) return false;
  // 含下划线且全英文 = 大概率 raw code
  if (/^[a-z_]+$/.test(value)) return true;
  return false;
}

/** 将单个 raw code 转为中文 label */
export function humanizeTagCode(code: string): string {
  return TAG_LABEL_MAP[code] || code;
}

/** 将字符串数组批量 humanize：raw code → label，空字符串过滤 */
export function humanizeTagList(items: string[]): string[] {
  return items
    .map((item) => {
      if (!item) return "";
      // Try label lookup first
      const label = TAG_LABEL_MAP[item];
      if (label !== undefined) return label; // allow empty string for known no-value codes
      // If it's a raw code without mapping, try heuristic
      if (isRawTagCode(item)) return "";
      // If it's already a Chinese sentence, keep it
      return item;
    })
    .filter(Boolean); // remove empty strings
}

/** 确保 matchHighlights 显示为可读文案 */
export function humanizeMatchHighlights(items: string[]): string[] {
  const result = humanizeTagList(items);
  if (result.length === 0) {
    return ["该岗位与当前简历在基础背景上存在一定关联"];
  }
  return result;
}

/** 确保 riskPoints 显示为可读文案 */
export function humanizeRiskPoints(items: string[]): string[] {
  const result = humanizeTagList(items);
  if (result.length === 0) {
    return ["建议进一步补充与该岗位相关的项目细节"];
  }
  return result;
}

/**
 * 如果 oneSentenceReason 仍为后端 raw tag 格式，重写为可读版本。
 *
 * 后端格式示例：
 *   "标签匹配：research、redis、algorithm"
 *   你的简历与该岗位在 Redis 工程能力、算法能力上重合度较高，适合作为优先投递岗位。
 */
export function humanizeOneSentenceReason(raw: string): string {
  if (!raw) return "该岗位与当前简历存在一定匹配度，可进一步查看岗位要求。";
  // 如果已是完整中文句子（不含英文 code 模式），原样返回
  if (!/标签匹配|research|algorithm|redis|golang|backend|docker|pytorch|mysql|skill_\d/.test(raw)) {
    return raw;
  }
  // 尝试提取 code 部分并转换
  const codes = raw.match(/[a-z_]+/g);
  if (codes) {
    const labels = codes
      .map((c) => TAG_LABEL_MAP[c])
      .filter(Boolean)
      .slice(0, 3);
    if (labels.length > 0) {
      return `你的简历与该岗位在 ${labels.join("、")} 等方面存在匹配，建议进一步查看岗位详情。`;
    }
  }
  return "该岗位与当前简历存在一定匹配度，可进一步查看岗位要求。";
}
