// ─── Tag Label Mapping for Fast Recommendation ──────────────────────────────
//
// 前端防御层：将后端可能返回的 raw tag code 转换为用户可读中文。
// 即使后端已修复，前端仍保留此映射作为安全兜底。

/** 原始 code → 中文 display label */
const TAG_LABEL_MAP: Record<string, string> = {
  // ── 学历 / 背景 ──
  degree_bachelor: "本科背景",
  degree_master: "硕士背景",
  degree_phd: "博士背景",
  major_cs: "计算机相关专业",
  major_software: "软件工程背景",
  major_ee: "电子/通信工程背景",
  major_math: "数学/统计背景",
  major_design: "设计类背景",
  major_business: "商科背景",

  // ── 经历 ──
  internship_exp: "实习经历",
  project_exp: "项目经验",
  backend_project: "后端项目经验",
  engineering_project: "工程项目经验",
  campus_project: "校园项目经验",
  open_source: "开源贡献",
  competition: "竞赛获奖",

  // ── 方向 ──
  research: "研究方向相关",
  algorithm: "算法方向",
  backend: "后端开发",
  frontend: "前端开发",
  fullstack: "全栈开发",
  test_dev: "测试开发",
  data: "数据分析",
  data_engineering: "数据工程",
  product: "产品方向",
  ai_infra: "AI 基础设施",
  cloud_native: "云原生方向",
  mobile: "移动端开发",
  security: "安全方向",
  game_dev: "游戏开发",
  public_service: "公共服务",

  // ── 技术栈 ──
  redis: "Redis 缓存",
  golang: "Go 后端开发",
  go: "Go 后端开发",
  python: "Python 开发",
  java: "Java 后端开发",
  cpp: "C/C++ 开发",
  rust: "Rust 开发",
  nodejs: "Node.js 开发",
  typescript: "TypeScript 开发",
  mysql: "MySQL 数据库",
  postgresql: "PostgreSQL 数据库",
  mongodb: "MongoDB 数据库",
  linux: "Linux 环境",
  docker: "Docker 部署",
  docker_k8s: "容器化部署",
  kubernetes: "Kubernetes",
  k8s: "Kubernetes",
  nginx: "Nginx 网关",
  git: "Git 协作",
  jwt: "JWT 认证",
  rbac: "RBAC 权限",
  api: "接口开发",
  restful: "RESTful API",
  sql: "SQL 能力",
  kafka: "消息队列",
  rabbitmq: "消息队列",
  grpc: "gRPC 开发",
  spring: "Spring 生态",
  spring_boot: "Spring Boot",
  pytorch_tf: "深度学习框架",
  pytorch: "PyTorch 框架",
  tensorflow: "TensorFlow 框架",
  hadoop: "Hadoop 生态",
  spark: "Spark 计算",

  // ── 工程能力 ──
  auth: "认证授权",
  permission: "权限校验",
  idempotency: "幂等设计",
  rate_limit: "限流设计",
  cache: "缓存设计",
  deployment: "部署经验",
  debugging: "问题排查",
  documentation: "文档沉淀",
  collaboration: "协作能力",
  ci_cd: "CI/CD 流程",
  testing: "测试能力",
  code_review: "代码审查",
  system_design: "系统设计",
  performance: "性能优化",
  monitoring: "监控告警",

  // ── 论文 / 学术 ──
  paper: "论文或学术成果",
  patent: "专利成果",
  publication: "学术发表",

  // ── 无意义占位符 → 过滤 ──
  skill_0: "",
  skill_1: "",
  skill_2: "",
  skill_3: "",
  skill_4: "",
  skill_5: "",
};

/** 过滤无效值：null、undefined、"null"、"undefined"、空串、NaN */
function isInvalidValue(value: unknown): boolean {
  if (value === null || value === undefined) return true;
  if (typeof value !== "string") return true;
  const trimmed = value.trim();
  if (trimmed.length === 0) return true;
  if (trimmed === "null" || trimmed === "undefined" || trimmed === "NaN") return true;
  return false;
}

/** 判断是否 look like a raw tag code（含下划线且全英文小写，或在映射表中有映射） */
export function isRawTagCode(value: string): boolean {
  if (isInvalidValue(value)) return true;
  // 已在映射表中 → 不算裸露 raw code（有中文 label 兜底）
  if (value in TAG_LABEL_MAP) return false;
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

/** 将字符串数组批量 humanize：raw code → label，过滤无效值 */
export function humanizeTagList(items: string[]): string[] {
  if (!Array.isArray(items)) return [];
  return items
    .map((item) => {
      // 过滤 null/undefined/"null"/"undefined"/NaN/空串
      if (isInvalidValue(item)) return "";
      // Try label lookup first
      const label = TAG_LABEL_MAP[item];
      if (label !== undefined) return label; // allow empty string for known no-value codes
      // If it's a raw code without mapping, filter it out
      if (isRawTagCode(item)) return "";
      // If it's already a readable phrase, keep it
      return item;
    })
    .filter((s) => s.trim().length > 0); // remove empty strings
}

/** 确保 matchHighlights 显示为可读文案 — 不强制兜底，允许无匹配时不显示 */
export function humanizeMatchHighlights(items: string[]): string[] {
  if (!Array.isArray(items)) return [];
  const result = humanizeTagList(items);
  // 不再强加泛化兜底文案 — 让页面自行判断是否显示该模块
  return result;
}

/** 确保 riskPoints 显示为可读文案 — 不强制兜底 */
export function humanizeRiskPoints(items: string[]): string[] {
  if (!Array.isArray(items)) return [];
  const result = humanizeTagList(items);
  // 不再强加泛化兜底文案 — 让页面自行判断是否显示该模块
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
