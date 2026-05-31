// ─── 岗位 Mock 数据 ─────────────────────────────────────────────────────────────

import type { Job, JobDetail } from '../types';

// ─── 岗位列表 Mock ──────────────────────────────────────────────────────────────

export const mockJobs: Job[] = [
  {
    id: 'job_backend_001',
    title: '软件开发-后台开发方向',
    category: 'backend',
    type_label: '技术',
    business_group: 'CDG / CSIG / TEG / WXG',
    locations: ['深圳', '北京', '上海', '广州', '成都'],
    tags: ['Go', 'C/C++', 'MySQL', 'Redis', '微服务'],
    short_description:
      '负责业务后台服务器的设计、开发、优化等工作；负责后台系统的高并发、高可用架构实现。',
    status: '开放',
  },
  {
    id: 'job_cloud_001',
    title: '云计算开发实习生',
    category: 'cloud',
    type_label: '实习',
    business_group: 'CSIG',
    locations: ['深圳', '北京'],
    tags: ['Kubernetes', 'Docker', 'Golang', 'Linux'],
    short_description:
      '参与公有云底层 IaaS/PaaS 管控系统开发；参与云原生架构设计与微服务改造。',
    status: '开放',
  },
  {
    id: 'job_algorithm_001',
    title: '算法-多模态方向',
    category: 'algorithm',
    type_label: '技术',
    business_group: 'TEG',
    locations: ['北京'],
    tags: ['NLP', 'CV', 'PyTorch', 'Deep Learning'],
    short_description:
      '负责多模态大模型的基础研发与落地应用；探索图文音等多模态数据的对齐与融合。',
    status: '开放',
  },
  {
    id: 'job_client_001',
    title: '软件开发-PC 客户端开发方向',
    category: 'client',
    type_label: '应届毕业生',
    business_group: 'WXG',
    locations: ['广州'],
    tags: ['C++', 'Qt', 'Windows/Mac API', '多线程'],
    short_description:
      '参与微信 PC 端的功能开发与性能优化；负责跨平台客户端架构演进。',
    status: '开放',
  },
  {
    id: 'job_data_001',
    title: '数据开发实习生',
    category: 'data',
    type_label: '实习',
    business_group: 'CDG',
    locations: ['深圳', '广州'],
    tags: ['Hadoop', 'Spark', 'Hive', 'Python', 'SQL'],
    short_description:
      '参与海量数据处理流程的开发与维护；协助构建数据仓库底座与数据治理体系。',
    status: '开放',
  },
  {
    id: 'job_game_001',
    title: '游戏策划培训生',
    category: 'game',
    type_label: '应届毕业生',
    business_group: 'IEG',
    locations: ['上海'],
    tags: ['系统策划', '关卡策划', '文案策划', '数值策划'],
    short_description:
      '参与重点游戏项目的世界观构建、系统设计、玩法创新或数值体验打磨。',
    status: '开放',
  },
];

// ─── 岗位详情 Mock（完整 JD） ───────────────────────────────────────────────────

export const mockJobDetails: JobDetail[] = [
  {
    id: 'job_backend_001',
    title: '软件开发-后台开发方向',
    category: 'backend',
    type_label: '技术',
    business_group: 'CDG / CSIG / TEG / WXG',
    locations: ['深圳', '北京', '上海', '广州', '成都'],
    tags: ['Go', 'C/C++', 'MySQL', 'Redis', '微服务'],
    short_description:
      '负责业务后台服务器的设计、开发、优化等工作；负责后台系统的高并发、高可用架构实现。',
    status: '开放',
    banner_summary:
      '参与腾讯产品后台系统的设计、开发与优化，打造稳定、高效、可扩展的后端支撑能力。',
    responsibilities: [
      '负责后台服务的设计、开发和性能优化，支撑高并发、高可用业务场景；',
      '参与系统架构设计，提升服务稳定性、可扩展性和可维护性；',
      '与产品、算法、客户端、测试等团队协作，完成业务需求落地；',
      '参与线上问题定位和性能优化，保障系统稳定运行；',
      '关注工程质量，推动代码规范、接口设计和服务治理优化。',
    ],
    requirements: [
      '扎实的计算机基础，熟悉数据结构、操作系统、网络、数据库等基础知识；',
      '熟悉 C/C++ / Java / Go 中至少一门开发语言；',
      '理解 TCP / UDP、HTTP、RPC、缓存、数据库等常见后端技术；',
      '具备良好的编码习惯、问题定位能力和学习能力；',
      '有项目实践或后端系统开发经验者优先。',
    ],
    bonus_points: [
      '有高并发、分布式系统、微服务、缓存、消息队列相关经验；',
      '有云原生、容器、Kubernetes、DevOps 相关经验；',
      '有开源项目、竞赛、实习或真实项目经验；',
      '对复杂业务系统和工程质量有持续优化意识。',
    ],
    delivery_notes:
      '该岗位支持使用当前简历进行 AI 匹配分析，分析结果仅作为投递前决策参考，不替代用户最终选择。',
  },
  {
    id: 'job_cloud_001',
    title: '云计算开发实习生',
    category: 'cloud',
    type_label: '实习',
    business_group: 'CSIG',
    locations: ['深圳', '北京'],
    tags: ['Kubernetes', 'Docker', 'Golang', 'Linux'],
    short_description:
      '参与公有云底层 IaaS/PaaS 管控系统开发；参与云原生架构设计与微服务改造。',
    status: '开放',
    banner_summary:
      '参与腾讯云底层基础设施和云原生平台的研发，打造业界领先的云计算产品。',
    responsibilities: [
      '参与公有云底层 IaaS/PaaS 管控系统开发；',
      '参与云原生架构设计与微服务改造；',
      '协助进行系统性能测试与调优；',
      '参与容器编排、服务网格等领域的技术调研和实现。',
    ],
    requirements: [
      '熟悉 Go 语言，了解 Docker 和 Linux 使用；',
      '对云计算、容器技术有浓厚兴趣；',
      '了解 Kubernetes 基本概念；',
      '具备良好的编码能力和学习能力。',
    ],
    bonus_points: [
      '有 Kubernetes 或 Docker 实战经验；',
      '了解 Service Mesh 或 Serverless 架构；',
      '有开源社区参与经历。',
    ],
    delivery_notes:
      '该岗位使用当前简历进行 AI 匹配分析，分析结果仅作投递前参考。',
  },
  {
    id: 'job_data_001',
    title: '数据开发实习生',
    category: 'data',
    type_label: '实习',
    business_group: 'CDG',
    locations: ['深圳', '广州'],
    tags: ['Hadoop', 'Spark', 'Hive', 'Python', 'SQL'],
    short_description:
      '参与海量数据处理流程的开发与维护；协助构建数据仓库底座与数据治理体系。',
    status: '开放',
    banner_summary:
      '参与腾讯数据平台的建设，支撑业务部门的数据分析和决策需求。',
    responsibilities: [
      '参与海量数据处理流程的开发与维护；',
      '协助构建数据仓库底座与数据治理体系；',
      '支持业务部门的数据分析需求与报表开发。',
    ],
    requirements: [
      '熟悉 SQL，了解 Hadoop/Spark 生态；',
      '具备 Python 或 Java 编程能力；',
      '有数据建模或 ETL 相关项目经验优先。',
    ],
    bonus_points: ['有 Hive/HBase 使用经验', '了解数据仓库建模方法论'],
  },
];

// ─── 默认选中岗位 ──────────────────────────────────────────────────────────────

export const mockSelectedJob: JobDetail = mockJobDetails[0];

// ─── 快捷查找 ──────────────────────────────────────────────────────────────────

export function getMockJobById(jobId: string): JobDetail | undefined {
  return mockJobDetails.find((j) => j.id === jobId);
}
