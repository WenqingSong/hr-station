// ─── 简历优化 Mock 数据 ────────────────────────────────────────────────────────

import type {
  ResumeOptimizationResult,
  AgentResponseEnvelope,
} from '../types';
import { createMockAgentMeta } from './common';

// ─── 简历优化结果 ──────────────────────────────────────────────────────────────

export const mockResumeOptimizationResult: ResumeOptimizationResult = {
  summary:
    '当前简历与后台开发岗位方向匹配，但 Linux 使用经验、高并发压测结果和线上问题排查案例证据不足。建议先完成 3 个关键修复，再进入投递。',
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
  keyword_alignment: [
    {
      keyword: 'Go',
      current_coverage: '已覆盖',
      suggested_action: '保留并突出项目中的接口开发职责',
    },
    {
      keyword: 'MySQL / Redis',
      current_coverage: '已覆盖',
      suggested_action: '补充缓存命中率或查询优化效果',
    },
    {
      keyword: 'Linux',
      current_coverage: '弱覆盖',
      suggested_action: '补充开发、部署或问题排查场景',
    },
    {
      keyword: '高并发',
      current_coverage: '未明确覆盖',
      suggested_action: '补充压测指标或性能优化过程',
    },
    {
      keyword: '微服务',
      current_coverage: '已覆盖',
      suggested_action: '说明服务拆分、接口设计和治理方式',
    },
  ],
  project_optimization: [
    {
      project_name: 'Go 微服务订单系统',
      current_problem: '描述偏功能实现，缺少性能指标和稳定性证据。',
      suggested_rewrite:
        '突出接口设计、数据库设计、Redis 缓存、错误处理、权限校验和服务拆分。',
      example_text:
        '将"完成订单接口开发"优化为"负责订单服务核心接口设计与实现，使用 MySQL 完成订单数据建模，结合 Redis 缓存降低高频查询压力，并通过统一错误处理和权限校验提升接口稳定性"。',
      need_user_verify: true,
    },
    {
      project_name: 'RBAC 权限管理系统',
      current_problem:
        '与后台开发相关，但没有突出权限校验、接口安全和系统稳定性。',
      suggested_rewrite:
        '强化角色权限模型、接口鉴权、操作日志和异常处理。',
      example_text:
        '将"实现权限管理功能"优化为"设计并实现基于角色的权限控制流程，完成用户、角色、菜单与接口权限的关联校验，并通过操作日志和异常处理提升后台系统可维护性"。',
      need_user_verify: true,
    },
  ],
  missing_evidence_plan: [
    'Linux 使用经验：考虑补充项目部署、服务器操作、日志排查场景',
    '高并发压测数据：如有 QPS/响应时间/缓存命中率等真实指标可补充',
    '线上问题排查案例：补充一次具体的 Bug 定位或性能优化过程',
  ],
  rewrite_examples: [
    {
      before: '完成订单接口开发，使用 MySQL 和 Redis。',
      suggestion:
        '补充职责、技术选择、优化目标和结果指标',
      after:
        '负责订单服务核心接口设计与实现，使用 MySQL 完成订单数据建模，并结合 Redis 缓存优化高频查询路径；如有真实压测数据，可补充 QPS、响应时间或缓存命中率等指标。',
      verification_notice: '需用户核实真实经历和数据后使用。',
    },
  ],
  verification_notice:
    'AI 不替用户编造经历，涉及 Linux 使用、压测数据、线上问题排查等内容必须由用户核实后再写入简历。',
  final_checklist: [
    'Top 3 修复任务是否完成',
    '需要核实的信息是否确认',
    '是否保留真实经历边界',
    '是否准备进入岗位投递',
  ],
};

// ─── 完整 Agent 响应 ───────────────────────────────────────────────────────────

export const mockResumeOptimizationResponse: AgentResponseEnvelope<ResumeOptimizationResult> =
  {
    result: mockResumeOptimizationResult,
    meta: createMockAgentMeta(
      'ResumeOptimizationAgent',
      'mock-resume-optimization-001',
    ),
  };
