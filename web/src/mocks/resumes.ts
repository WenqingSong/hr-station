// ─── 简历 Mock 数据 ─────────────────────────────────────────────────────────────

import type {
  Resume,
  ResumeOption,
  ResumeVersionComparison,
  ResumeKeywordChange,
  ResumeProjectChange,
} from '../types';

// ─── V2 后端开发简历（当前使用） ─────────────────────────────────────────────────

const resumeV2: Resume = {
  id: 'resume_v2',
  name: 'V2 后端开发简历',
  version: '2',
  status: 'current',
  direction: 'backend',
  direction_label: '后端开发',
  completeness: 78,
  updated_at: '2 小时前',
  profile: {
    name: '测试候选人 A',
    age: 23,
    education: '本科',
    school: '某大学',
    major: '软件工程',
    graduation_year: '2026',
    phone: '182****9164',
    email: 'candidate@example.com',
  },
  expected_cities: ['深圳', '北京', '上海'],
  job_type: '校园招聘 / 实习',
  skills: [
    'Go', 'MySQL', 'Redis', 'Docker', '微服务', 'RBAC', '接口开发', '缓存优化',
  ],
  gaps: [
    'Linux 使用经验描述不足',
    '高并发压测数据缺失',
    '线上问题排查案例较少',
  ],
  education: [
    {
      school: '某大学',
      major: '软件工程',
      degree: '本科',
      start_date: '2022.09',
      end_date: '2026.06',
      courses: ['数据结构', '操作系统', '计算机网络', '数据库', '软件工程'],
    },
  ],
  experiences: [
    {
      company: 'ExampleTech',
      role: '后端开发实习生',
      start_date: '2025.06',
      end_date: '2025.09',
      description:
        '参与后台接口开发和数据库表设计，协助完成用户权限、订单查询和缓存优化等功能。',
    },
  ],
  projects: [
    {
      id: 'proj_1',
      name: 'Go 微服务订单系统',
      tech_stack: ['Go', 'MySQL', 'Redis', 'Docker'],
      description:
        '负责订单服务核心接口设计与实现，使用 MySQL 完成订单数据建模，结合 Redis 缓存优化高频查询路径。',
      responsibilities: [
        '接口设计',
        '数据库设计',
        '缓存优化',
        '错误处理',
        '权限校验',
      ],
    },
    {
      id: 'proj_2',
      name: 'RBAC 权限管理系统',
      tech_stack: ['Go', 'MySQL', 'Casbin', 'JWT'],
      description:
        '设计并实现用户、角色、菜单与接口权限的关联校验，提升后台系统权限管理的可维护性。',
      responsibilities: ['权限模型设计', '接口鉴权', '后台管理接口开发'],
    },
    {
      id: 'proj_3',
      name: '缓存优化项目',
      tech_stack: ['Redis', 'MySQL'],
      description:
        '针对高频查询场景设计缓存策略，减少重复数据库访问。',
      responsibilities: ['缓存策略设计', '查询优化', '接口性能优化'],
    },
  ],
  self_evaluation:
    '具备 Go 后端开发基础，熟悉 MySQL、Redis 和常见 Web 后端开发流程，了解微服务、权限控制和接口设计。希望从事后端开发、云计算开发等方向。',
};

// ─── V1 通用后端简历（历史版本） ─────────────────────────────────────────────────

const resumeV1: Resume = {
  id: 'resume_v1',
  name: 'V1 通用后端简历',
  version: '1',
  status: 'history',
  direction: 'backend',
  direction_label: '后端开发',
  completeness: 55,
  updated_at: '3 天前',
  profile: {
    name: '测试候选人 A',
    age: 23,
    education: '本科',
    school: '某大学',
    major: '软件工程',
    graduation_year: '2026',
    phone: '182****9164',
    email: 'candidate@example.com',
  },
  expected_cities: ['深圳', '北京'],
  job_type: '校园招聘',
  skills: ['Go', 'Java', 'MySQL'],
  gaps: [
    '项目经历数量较少',
    'Redis / 微服务证据不足',
    '与后台开发岗位 JD 的关键词匹配不够完整',
    '缺少简历优化后的投递重点表达',
  ],
  education: [
    {
      school: '某大学',
      major: '软件工程',
      degree: '本科',
      start_date: '2022.09',
      end_date: '2026.06',
      courses: ['数据结构', '操作系统', '计算机网络', '软件工程'],
    },
  ],
  experiences: [
    {
      company: 'ExampleTech',
      role: '后端开发实习生',
      start_date: '2025.06',
      end_date: '2025.09',
      description:
        '参与后台接口开发，协助完成用户权限和订单查询功能。',
    },
  ],
  projects: [
    {
      id: 'proj_v1_1',
      name: 'Go 微服务订单系统',
      tech_stack: ['Go', 'MySQL'],
      description:
        '完成订单接口开发，使用 MySQL 和 Redis。',
      responsibilities: ['接口开发'],
    },
    {
      id: 'proj_v1_2',
      name: 'RBAC 权限管理系统',
      tech_stack: ['Go', 'MySQL', 'JWT'],
      description:
        '实现权限管理功能。',
      responsibilities: ['权限模型设计', '接口开发'],
    },
  ],
  self_evaluation:
    '具备后端开发基础，希望从事后端开发方向。',
};

// ─── V3 产品方向简历（方向不匹配） ────────────────────────────────────────────────

const resumeV3: Resume = {
  id: 'resume_v3',
  name: 'V3 产品方向简历',
  version: '3',
  status: 'history',
  direction: 'product',
  direction_label: '产品策划',
  completeness: 60,
  updated_at: '1 周前',
  profile: {
    name: '测试候选人 A',
    age: 23,
    education: '本科',
    school: '某大学',
    major: '软件工程',
    graduation_year: '2026',
    phone: '182****9164',
    email: 'candidate@example.com',
  },
  expected_cities: ['深圳', '北京'],
  job_type: '校园招聘',
  skills: ['需求分析', '用户调研', '原型设计', 'PRD'],
  gaps: [],
  education: [
    {
      school: '某大学',
      major: '软件工程',
      degree: '本科',
      start_date: '2022.09',
      end_date: '2026.06',
      courses: ['软件工程', '产品设计基础'],
    },
  ],
  experiences: [
    {
      company: 'ExampleTech',
      role: '产品实习生',
      start_date: '2025.03',
      end_date: '2025.06',
      description:
        '参与产品需求分析和用户调研，协助完成 PRD 和原型设计。',
    },
  ],
  projects: [
    {
      id: 'proj_v3_1',
      name: '校园社交产品设计',
      tech_stack: ['Figma', 'Axure'],
      description:
        '参与校园社交产品的需求调研、用户画像构建和低保真原型设计。',
      responsibilities: ['需求调研', '原型设计', 'PRD 编写'],
    },
    {
      id: 'proj_v3_2',
      name: '线上活动运营分析',
      tech_stack: ['Excel', 'Python'],
      description:
        '协助分析线上活动的用户参与数据，输出运营优化建议。',
      responsibilities: ['数据分析', '运营建议'],
    },
  ],
  self_evaluation:
    '具备产品策划基础，了解用户调研和需求分析方法，希望从事产品策划方向。',
};

// ─── 导出 ────────────────────────────────────────────────────────────────────

export const mockResumeVersions: Resume[] = [resumeV2, resumeV1, resumeV3];

export const mockCurrentResume: Resume = resumeV2;

export function getMockResumeById(resumeId: string): Resume | undefined {
  return mockResumeVersions.find((r) => r.id === resumeId);
}

// ─── 选择简历弹窗选项 ─────────────────────────────────────────────────────────

export const mockResumeOptions: ResumeOption[] = [
  {
    id: 'resume_v2',
    name: 'V2 后端开发简历',
    status: 'current',
    direction_label: '后端开发',
    tags: ['Go', 'MySQL', 'Redis', '微服务'],
    project_count: 3,
    updated_at: '2 小时前',
    recommendation_label: '推荐使用',
    is_default: true,
    is_recommended: true,
  },
  {
    id: 'resume_v1',
    name: 'V1 通用后端简历',
    status: 'history',
    direction_label: '后端开发',
    tags: ['Go', 'Java', 'MySQL'],
    project_count: 2,
    updated_at: '3 天前',
    recommendation_label: '可用于分析，但项目证据不如 V2 完整',
    is_default: false,
    is_recommended: false,
  },
  {
    id: 'resume_v3',
    name: 'V3 产品方向简历',
    status: 'history',
    direction_label: '产品策划',
    tags: ['需求分析', '用户调研', '原型设计'],
    project_count: 2,
    updated_at: '1 周前',
    mismatch_reason: '与当前技术岗位方向不一致，不建议用于本次分析。',
    is_default: false,
    is_recommended: false,
  },
];

// ─── 历史简历比较 ──────────────────────────────────────────────────────────────

const mockKeywordChanges: ResumeKeywordChange[] = [
  {
    keyword: 'Go',
    before_status: '已覆盖',
    after_status: '已强化',
    change_note: '项目表达更具体',
  },
  {
    keyword: 'MySQL',
    before_status: '已覆盖',
    after_status: '已强化',
    change_note: '增加数据库设计和查询优化描述',
  },
  {
    keyword: 'Redis',
    before_status: '弱覆盖',
    after_status: '已强化',
    change_note: '新增缓存优化项目',
  },
  {
    keyword: 'Docker',
    before_status: '未覆盖',
    after_status: '新增',
    change_note: '补充部署与工程化关键词',
  },
  {
    keyword: '微服务',
    before_status: '弱覆盖',
    after_status: '已强化',
    change_note: '强化服务拆分与接口设计表达',
  },
  {
    keyword: 'RBAC',
    before_status: '未覆盖',
    after_status: '新增',
    change_note: '新增权限管理系统经历',
  },
];

const mockProjectChanges: ResumeProjectChange[] = [
  {
    project_name: 'Go 微服务订单系统',
    change_note:
      '从普通接口开发描述强化为包含接口设计、数据库设计、Redis 缓存、错误处理和权限校验的完整后端项目表达。',
  },
  {
    project_name: 'RBAC 权限管理系统',
    change_note:
      '从功能描述强化为权限模型、接口鉴权、后台管理系统可维护性表达。',
  },
  {
    project_name: '缓存优化项目',
    change_note: 'V2 新增项目，用于补充 Redis / MySQL 查询优化证据。',
  },
];

export const mockResumeComparison: ResumeVersionComparison = {
  base_resume: resumeV1,
  target_resume: resumeV2,
  summary:
    'V2 版本相对 V1 新增了缓存优化项目、Docker / 微服务 / RBAC 等关键词，项目经历从 2 个增加到 3 个，关键词覆盖从 3 个增加到 8 个，求职方向从通用后端聚焦为后端开发 / 云计算开发。',
  keyword_changes: mockKeywordChanges,
  project_changes: mockProjectChanges,
  remaining_risks: [
    'Linux 使用经验仍然描述不足',
    '高并发压测数据仍然缺失',
    '线上问题排查案例仍然较少',
  ],
  recommendation:
    '建议将 V2 后端开发简历设为当前简历，并优先用于后台开发、云计算开发等技术岗位的岗位推荐和岗位匹配分析。',
};
