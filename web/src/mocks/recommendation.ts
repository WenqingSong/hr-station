// ─── 岗位推荐 Mock 数据 ────────────────────────────────────────────────────────

import type {
  RecommendationResult,
  AgentResponseEnvelope,
  RecommendedJobItem,
} from '../types';
import { createMockAgentMeta } from './common';

// ─── Top 5 推荐岗位 ────────────────────────────────────────────────────────────

const recommendedJobs: RecommendedJobItem[] = [
  {
    job_id: 'job_backend_001',
    job_title: '软件开发-后台开发方向',
    recommendation_level: '强推荐',
    recommendation_score: 95,
    recommendation_reason:
      '技术栈高度匹配，项目经验覆盖岗位核心要求。',
    match_keywords: ['Go', 'MySQL', 'Redis', '微服务'],
    location: '深圳',
    reason: '技术栈高度匹配，项目经验覆盖岗位核心要求。',
    risk_tags: ['高并发压测数据缺失'],
    next_action: '优先进入岗位详情，进行匹配分析',
  },
  {
    job_id: 'job_cloud_001',
    job_title: '云计算开发实习生',
    recommendation_level: '推荐',
    recommendation_score: 78,
    recommendation_reason:
      '具备 Go 与微服务基础，但需补充云原生知识。',
    match_keywords: ['Go', 'Kubernetes', '微服务'],
    location: '深圳 / 北京',
    reason: '具备 Go 与微服务基础，但需补充云原生知识。',
    risk_tags: ['云原生项目证据偏少'],
    next_action: '进入岗位详情查看 JD，评估匹配度',
  },
  {
    job_id: 'job_data_001',
    job_title: '数据开发实习生',
    recommendation_level: '推荐',
    recommendation_score: 72,
    recommendation_reason:
      '数据库基础扎实，有一定数据处理经验。',
    match_keywords: ['MySQL', 'Redis', '后端开发'],
    location: '深圳 / 广州',
    reason: '数据库基础扎实，有一定数据处理经验。',
    risk_tags: ['缺少大数据计算框架经验'],
    next_action: '查看 JD 了解大数据框架要求',
  },
  {
    job_id: 'job_client_001',
    job_title: '软件开发-PC 客户端开发方向',
    recommendation_level: '谨慎推荐',
    recommendation_score: 55,
    recommendation_reason:
      '部分底层协议知识可复用，但方向偏差较大。',
    match_keywords: ['C++', '网络协议'],
    location: '深圳',
    reason: '部分底层协议知识可复用，但方向偏差较大。',
    risk_tags: ['完全缺少客户端开发经验'],
    next_action: '仅作备选参考',
  },
  {
    job_id: 'job_algorithm_001',
    job_title: '算法-搜索提效推荐方向',
    recommendation_level: '谨慎推荐',
    recommendation_score: 45,
    recommendation_reason:
      '具备架构理解，但核心算法能力不匹配。',
    match_keywords: ['后端架构'],
    location: '北京 / 深圳',
    reason: '具备架构理解，但核心算法能力不匹配。',
    risk_tags: ['无推荐算法相关经历'],
    next_action: '仅作备选参考',
  },
];

// ─── 推荐结果 ──────────────────────────────────────────────────────────────────

export const mockRecommendationResult: RecommendationResult = {
  recommended_jobs: recommendedJobs,
  recommendation_summary:
    '基于当前简历（V2 后端开发简历），从岗位池中为您推荐最值得优先关注的 5 个岗位。其中后台开发方向与您的简历高度匹配，建议优先关注。',
  apply_readiness: 'fix_before_apply',
  optimization_stop_signal: 'fix_top_items_first',
  top_fix_items: [
    {
      title: '补充云原生项目经历',
      priority: 'medium',
      expected_impact: '提升云计算方向推荐优先级',
      need_user_verify: true,
      description:
        '如果你确实参与过 Docker 部署、Kubernetes 相关实践，请在项目经历中补充。',
    },
    {
      title: '量化高并发压测数据',
      priority: 'high',
      expected_impact: '增强后台开发岗位推荐信心度',
      need_user_verify: true,
    },
    {
      title: '补充 Linux 使用经验',
      priority: 'medium',
      expected_impact: '增强后台和云计算方向证据覆盖',
      need_user_verify: true,
    },
  ],
  evidence_chain: [
    {
      jd_requirement: '熟悉 Go / Java / C++ 至少一门语言',
      resume_evidence:
        'Go 微服务项目，包含接口开发和数据库设计',
      evidence_strength: 'strong',
    },
    {
      jd_requirement: '理解数据库和缓存',
      resume_evidence:
        '使用 MySQL、Redis 完成订单缓存与查询优化',
      evidence_strength: 'strong',
    },
    {
      jd_requirement: '高并发或分布式经验',
      resume_evidence:
        '有微服务项目，但缺少 QPS、压测、性能优化数据',
      evidence_strength: 'medium',
      risk_note: '缺少压测数据支撑',
    },
  ],
  risk_items: [
    '高并发压测结果描述不足',
    'Linux 使用经验证据不够明确',
    '缺少线上问题排查案例',
  ],
  next_actions: [
    '优先进入软件开发-后台开发方向岗位详情页',
    '进行岗位匹配分析，评估匹配度',
    '如有时间，先修复 Top 3 简历问题后投递',
  ],
};

// ─── 完整 Agent 响应 ───────────────────────────────────────────────────────────

export const mockRecommendationResponse: AgentResponseEnvelope<RecommendationResult> =
  {
    result: mockRecommendationResult,
    meta: createMockAgentMeta('RecommendationAgent', 'mock-recommendation-001'),
  };
