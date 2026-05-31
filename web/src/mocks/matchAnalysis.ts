// ─── 岗位匹配分析 Mock 数据 ────────────────────────────────────────────────────

import type {
  MatchAnalysisResult,
  AgentResponseEnvelope,
} from '../types';
import { createMockAgentMeta } from './common';

// ─── 匹配分析结果 ──────────────────────────────────────────────────────────────

export const mockMatchAnalysisResult: MatchAnalysisResult = {
  match_score: 82,
  match_level: '较高匹配',
  summary:
    '你的 Go 后端、MySQL、Redis 和微服务项目经历与岗位核心要求较匹配，但高并发压测、Linux 使用经验和线上问题排查证据仍需补强。',
  apply_readiness: 'fix_before_apply',
  optimization_stop_signal: 'fix_top_items_first',
  top_fix_items: [
    {
      title: '补充 Linux 使用经验',
      priority: 'high',
      expected_impact: '提升后台开发岗位可信度',
      need_user_verify: true,
      description:
        '如果你确实在项目开发或部署中使用过 Linux，请补充具体命令、环境或排查经历；如果没有，不要编造。',
    },
    {
      title: '量化微服务项目 QPS / 响应时间 / 缓存命中率',
      priority: 'high',
      expected_impact: '增强高并发和性能优化证据',
      need_user_verify: true,
      description:
        '如果项目有压测或性能数据，请补充真实指标；如果没有，可以描述优化思路，但不要虚构数据。',
    },
    {
      title: '补充线上问题排查或接口优化案例',
      priority: 'medium',
      expected_impact: '增强工程实践与稳定性意识',
      need_user_verify: true,
      description:
        '补充一次接口慢查询、缓存失效、权限校验、异常排查或日志定位经历。',
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
      risk_note: '可以补充性能数据',
    },
    {
      jd_requirement: '高并发或分布式系统经验',
      resume_evidence:
        '有微服务项目，但缺少 QPS、压测、性能优化数据',
      evidence_strength: 'medium',
      risk_note: '建议量化项目指标',
    },
    {
      jd_requirement: '熟悉 Linux / 后台服务部署',
      resume_evidence:
        '简历中未明确体现 Linux 使用经验',
      evidence_strength: 'weak',
      risk_note: '需要补充 Linux 环境开发或部署经历',
    },
  ],
  match_highlights: [
    '后端开发方向与岗位方向一致',
    'Go / MySQL / Redis 技能与岗位要求高度相关',
    '微服务项目能支撑后台系统设计与接口开发经历',
  ],
  risk_items: [
    '高并发压测结果描述不足',
    'Linux 使用经验不够明确',
    '线上问题排查案例缺失',
  ],
  next_actions: [
    '建议先修复 Top 3 简历问题，再进入投递',
    '当前岗位值得优先关注',
    '简历中仍有几处关键证据需要补强',
  ],
};

// ─── 完整 Agent 响应 ───────────────────────────────────────────────────────────

export const mockMatchAnalysisResponse: AgentResponseEnvelope<MatchAnalysisResult> =
  {
    result: mockMatchAnalysisResult,
    meta: createMockAgentMeta('MatchAnalysisAgent', 'mock-match-001'),
  };
