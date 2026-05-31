import React, { useState } from "react";
import { Link } from "react-router";
import { SelectResumeModal } from "../components/SelectResumeModal";
import {
  FileText,
  Plus,
  ArrowRightLeft,
  CheckCircle2,
  Clock,
  Sparkles,
  MapPin,
  Mail,
  Phone,
  GraduationCap,
  Briefcase,
  Code2,
  Award,
  MessageSquare,
  ChevronDown,
} from "lucide-react";

// ─── Data ────────────────────────────────────────────────────────────────────

const resumeVersions = [
  { id: "v2", label: "V2 后端开发简历", active: true },
  { id: "v1", label: "V1 通用后端简历", active: false },
  { id: "v3", label: "V3 产品方向简历", active: false },
];

const skills = ["Go", "MySQL", "Redis", "Docker", "微服务", "RBAC", "接口开发", "缓存优化"];

const projects = [
  {
    name: "Go 微服务订单系统",
    stack: "Go / MySQL / Redis / Docker / 微服务",
    desc: "负责订单服务核心接口设计与实现，使用 MySQL 完成订单数据建模，结合 Redis 缓存优化高频查询路径。",
    role: "接口设计、数据库设计、缓存优化、错误处理、权限校验",
  },
  {
    name: "RBAC 权限管理系统",
    stack: "Go / MySQL / Casbin / JWT",
    desc: "设计并实现用户、角色、菜单与接口权限的关联校验，提升后台系统权限管理的可维护性。",
    role: "权限模型设计、接口开发、角色配置",
  },
  {
    name: "缓存优化项目",
    stack: "Redis / MySQL / 查询优化",
    desc: "针对高频查询场景设计缓存策略，减少重复数据库访问，降低核心查询响应时间。",
    role: "缓存方案设计、SQL 优化、性能测试",
  },
];

// ─── Section wrapper ──────────────────────────────────────────────────────────

function ResumeSection({
  icon,
  title,
  children,
}: {
  icon: React.ReactNode;
  title: string;
  children: React.ReactNode;
}) {
  return (
    <div className="bg-white border border-gray-200">
      {/* Section header */}
      <div className="flex items-center gap-2.5 px-8 py-4 border-b border-gray-100">
        <span className="text-[#0052D9]">{icon}</span>
        <h2 className="font-semibold text-gray-900">{title}</h2>
      </div>
      <div className="px-8 py-5">{children}</div>
    </div>
  );
}

// ─── Main Component ───────────────────────────────────────────────────────────

export function ResumeManagementPage() {
  const [currentVersion, setCurrentVersion] = useState("v2");
  const [versionOpen, setVersionOpen] = useState(false);
  const [showSelectModal, setShowSelectModal] = useState(false);

  const activeVersion = resumeVersions.find((v) => v.id === currentVersion);

  return (
    <>
      {/* ── Page Content ────────────────────────────────────────────────────── */}
      <div className="max-w-[960px] mx-auto px-4 py-8 pb-12">

        {/* ── Page Title Area ──────────────────────────────────────────────── */}
        <div className="flex items-start justify-between mb-5">
          <div>
            <h1 className="text-xl font-semibold text-gray-900">我的简历</h1>
            <p className="mt-1 text-sm text-gray-500">
              当前简历将用于岗位推荐、岗位匹配分析和简历优化。
            </p>
          </div>
          <Link
            to="/resumes/edit"
            className="flex items-center gap-1.5 px-4 py-2 bg-[#0052D9] text-white text-sm font-medium rounded hover:bg-[#003DA5] transition-colors"
          >
            <Plus className="w-4 h-4" />
            添加简历
          </Link>
        </div>

        {/* ── Resume Status Bar ────────────────────────────────────────────── */}
        <div className="bg-white border border-gray-200 rounded-sm px-5 py-3 mb-5 flex items-center gap-6">
          <div className="flex items-center gap-2 text-sm text-gray-600">
            <span className="text-gray-400">当前简历：</span>
            <span className="font-medium text-gray-900">{activeVersion?.label}</span>
          </div>
          <div className="w-px h-4 bg-gray-200" />
          <div className="flex items-center gap-2 text-sm text-gray-600">
            <span className="text-gray-400">简历完整度：</span>
            <div className="flex items-center gap-2">
              <div className="w-20 h-1.5 bg-gray-100 rounded-full overflow-hidden">
                <div className="h-full bg-[#0052D9] rounded-full" style={{ width: "78%" }} />
              </div>
              <span className="font-semibold text-[#0052D9]">78%</span>
            </div>
          </div>
          <div className="w-px h-4 bg-gray-200" />
          <div className="flex items-center gap-1.5 text-sm text-gray-500">
            <Clock className="w-3.5 h-3.5" />
            最近更新：2 小时前
          </div>
          <div className="w-px h-4 bg-gray-200" />
          <div className="flex items-center gap-1.5 text-sm text-green-600">
            <CheckCircle2 className="w-3.5 h-3.5" />
            可用于 AI 分析
          </div>
        </div>

        {/* ── Blue Personal Info Card ──────────────────────────────────────── */}
        <div className="bg-[#0052D9] rounded-sm mb-0.5 overflow-hidden">
          <div className="px-8 py-6 flex items-start gap-6">

            {/* Center info */}
            <div className="flex-1 min-w-0">
              {/* Name + tags */}
              <div className="flex items-center gap-3 mb-2">
                <h2 className="text-white font-semibold text-xl">测试候选人 A</h2>
                <span className="px-2 py-0.5 bg-white/15 border border-white/25 rounded text-white text-xs">应届毕业生</span>
                <span className="px-2 py-0.5 bg-white/15 border border-white/25 rounded text-white text-xs">后端开发</span>
                <span className="px-2 py-0.5 bg-white/15 border border-white/25 rounded text-white text-xs">2026 校园招聘</span>
              </div>

              {/* Basic info row */}
              <div className="flex items-center gap-4 mb-3">
                <span className="text-white/70 text-sm">男 · 23岁 · 本科</span>
                <span className="w-px h-3 bg-white/30" />
                <span className="text-white/70 text-sm">某大学 · 软件工程</span>
                <span className="w-px h-3 bg-white/30" />
                <span className="text-white/70 text-sm">毕业时间：2026</span>
              </div>

              {/* Contact row */}
              <div className="flex items-center gap-5 mb-3">
                <span className="flex items-center gap-1.5 text-white/80 text-sm">
                  <Phone className="w-3.5 h-3.5" />
                  182****9164
                </span>
                <span className="flex items-center gap-1.5 text-white/80 text-sm">
                  <Mail className="w-3.5 h-3.5" />
                  candidate@example.com
                </span>
                <span className="flex items-center gap-1.5 text-white/80 text-sm">
                  <MapPin className="w-3.5 h-3.5" />
                  深圳 / 北京 / 上海
                </span>
              </div>

              {/* Job direction */}
              <div className="flex items-start gap-4">
                <div className="flex items-center gap-2">
                  <span className="text-white/50 text-xs">求职方向：</span>
                  <span className="text-white text-sm font-medium">后端开发</span>
                </div>
                <span className="w-px h-3.5 bg-white/30 mt-0.5" />
                <div className="flex items-center gap-2 flex-wrap">
                  <span className="text-white/50 text-xs">目标岗位：</span>
                  <span className="text-white text-sm">软件开发-后台开发方向</span>
                  <span className="text-white/50 text-xs">/</span>
                  <span className="text-white text-sm">云计算开发实习生</span>
                </div>
              </div>
            </div>

            {/* Right: resume version — compact, no button cluster */}
            <div className="flex-shrink-0 text-right min-w-[150px]">
              <div className="text-white/50 text-xs mb-1.5">当前简历版本</div>
              <div className="text-white font-medium text-sm mb-2">{activeVersion?.label}</div>
              <div className="text-white/65 text-xs">当前使用｜可用于 AI 分析</div>
            </div>

          </div>
        </div>

        {/* ── Action Buttons ────────────────────────────────────────────────── */}
        <div className="bg-white border border-gray-200 border-t-0 px-8 py-3 mb-4 flex items-center gap-3">
          <button
            onClick={() => setShowSelectModal(true)}
            className="px-5 py-1.5 bg-[#0052D9] text-white text-sm font-medium rounded hover:bg-[#003DA5] transition-colors flex items-center gap-1.5"
          >
            <Sparkles className="w-3.5 h-3.5" />
            开始 AI 分析
          </button>
          <Link
            to="/resumes/edit"
            className="px-5 py-1.5 border border-[#0052D9] text-[#0052D9] text-sm font-medium rounded hover:bg-blue-50 transition-colors"
          >
            修改简历
          </Link>
          <Link
            to="/resumes/edit"
            className="px-5 py-1.5 border border-[#0052D9] text-[#0052D9] text-sm font-medium rounded hover:bg-blue-50 transition-colors"
          >
            添加简历
          </Link>
          <Link
            to="/resumes/compare"
            className="px-5 py-1.5 border border-gray-300 text-gray-700 text-sm font-medium rounded hover:bg-gray-50 transition-colors flex items-center gap-1.5"
          >
            <ArrowRightLeft className="w-3.5 h-3.5" />
            历史简历比较
          </Link>
          <div className="flex-1" />
        </div>

        {/* ── Version Switcher (lightweight) ───────────────────────────────── */}
        <div className="bg-white border border-gray-200 px-8 py-3 mb-4 flex items-center gap-3">
          <span className="text-sm text-gray-500 flex-shrink-0">简历版本：</span>
          <div className="relative">
            <button
              onClick={() => setVersionOpen(!versionOpen)}
              className="flex items-center gap-2 px-3 py-1 border border-gray-300 rounded text-sm text-gray-800 hover:border-[#0052D9] transition-colors bg-white"
            >
              <span className="font-medium text-[#0052D9]">{activeVersion?.label}</span>
              <span className="text-[10px] px-1.5 py-0.5 bg-blue-50 text-[#0052D9] rounded border border-blue-100 ml-1">当前</span>
              <ChevronDown className="w-3.5 h-3.5 text-gray-400" />
            </button>
            {versionOpen && (
              <div className="absolute top-full left-0 mt-1 bg-white border border-gray-200 rounded shadow-md z-10 min-w-[200px]">
                {resumeVersions.map((v) => (
                  <button
                    key={v.id}
                    onClick={() => { setCurrentVersion(v.id); setVersionOpen(false); }}
                    className={`w-full flex items-center justify-between px-4 py-2.5 text-sm hover:bg-gray-50 transition-colors ${
                      v.id === currentVersion ? "text-[#0052D9]" : "text-gray-700"
                    }`}
                  >
                    <span>{v.label}</span>
                    {v.id === currentVersion && (
                      <CheckCircle2 className="w-3.5 h-3.5 text-[#0052D9]" />
                    )}
                  </button>
                ))}
                <div className="border-t border-gray-100 px-4 py-2">
                  <Link
                    to="/resumes/edit"
                    className="text-xs text-[#0052D9] hover:underline"
                    onClick={() => setVersionOpen(false)}
                  >
                    + 添加新版本
                  </Link>
                </div>
              </div>
            )}
          </div>
          <div className="w-px h-4 bg-gray-200" />
          <div className="flex items-center gap-3">
            {resumeVersions.filter(v => v.id !== currentVersion).map(v => (
              <button
                key={v.id}
                onClick={() => setCurrentVersion(v.id)}
                className="text-sm text-gray-500 hover:text-[#0052D9] transition-colors"
              >
                {v.label}
              </button>
            ))}
          </div>
          <div className="flex-1" />
          <button className="text-xs text-gray-400 hover:text-[#0052D9] transition-colors flex items-center gap-1">
            <ArrowRightLeft className="w-3 h-3" />
            历史比较
          </button>
        </div>

        {/* ── AI Status (lightweight one line) ────────────────────────────── */}
        <div className="flex items-center gap-2 px-4 py-2.5 bg-blue-50 border border-blue-100 mb-4 rounded-sm">
          <Sparkles className="w-3.5 h-3.5 text-[#0052D9] flex-shrink-0" />
          <span className="text-xs text-[#0052D9]">
            当前简历可用于：<strong>岗位推荐</strong> / <strong>岗位匹配分析</strong> / <strong>简历优化</strong>
          </span>
          <span className="w-px h-3.5 bg-blue-200 mx-1" />
          <span className="text-xs text-gray-400">AI 建议仅作投递前参考，不替代真实经历核实。</span>
        </div>

        {/* ══════════ Resume Modules ══════════════════════════════════════════ */}

        <div className="space-y-0.5">

          {/* 1. 求职意向 ──────────────────────────────────────────────────── */}
          <ResumeSection icon={<Briefcase className="w-4 h-4" />} title="求职意向">
            <div className="grid grid-cols-2 gap-y-4 gap-x-12">
              <div>
                <div className="text-xs text-gray-400 mb-1">求职方向</div>
                <div className="text-sm text-gray-800">后端开发</div>
              </div>
              <div>
                <div className="text-xs text-gray-400 mb-1">岗位类型</div>
                <div className="text-sm text-gray-800">校园招聘 / 实习生</div>
              </div>
              <div>
                <div className="text-xs text-gray-400 mb-1">目标岗位</div>
                <div className="text-sm text-gray-800 leading-relaxed">
                  软件开发-后台开发方向<br />
                  云计算开发实习生
                </div>
              </div>
              <div>
                <div className="text-xs text-gray-400 mb-1">可到岗时间</div>
                <div className="text-sm text-gray-800">2026 年毕业后 / 可实习</div>
              </div>
              <div>
                <div className="text-xs text-gray-400 mb-1">期望城市</div>
                <div className="text-sm text-gray-800">深圳 / 北京 / 上海</div>
              </div>
            </div>
          </ResumeSection>

          {/* 2. 教育经历 ──────────────────────────────────────────────────── */}
          <ResumeSection icon={<GraduationCap className="w-4 h-4" />} title="教育经历">
            <div className="flex items-start justify-between">
              <div>
                <div className="flex items-center gap-3 mb-1.5">
                  <span className="font-medium text-gray-900">某大学</span>
                  <span className="text-gray-400 text-xs">本科</span>
                </div>
                <div className="text-sm text-gray-600 mb-2.5">软件工程</div>
                <div className="text-xs text-gray-400 leading-relaxed">
                  相关课程：数据结构、操作系统、计算机网络、数据库、软件工程
                </div>
              </div>
              <div className="text-sm text-gray-400 text-right flex-shrink-0 ml-8">
                2022.09 - 2026.06
              </div>
            </div>
          </ResumeSection>

          {/* 3. 实习经历 ──────────────────────────────────────────────────── */}
          <ResumeSection icon={<Briefcase className="w-4 h-4" />} title="实习经历">
            <div className="flex items-start justify-between">
              <div className="flex-1">
                <div className="flex items-center gap-3 mb-1.5">
                  <span className="font-medium text-gray-900">ExampleTech</span>
                  <span className="text-sm text-gray-600">后端开发实习生</span>
                </div>
                <div className="text-sm text-gray-600 leading-relaxed">
                  参与后台接口开发和数据库表设计，协助完成用户权限、订单查询和缓存优化等功能。
                </div>
              </div>
              <div className="text-sm text-gray-400 text-right flex-shrink-0 ml-8">
                2025.06 - 2025.09
              </div>
            </div>
          </ResumeSection>

          {/* 4. 项目经历 ──────────────────────────────────────────────────── */}
          <ResumeSection icon={<Code2 className="w-4 h-4" />} title="项目经历">
            <div className="space-y-6">
              {projects.map((proj, i) => (
                <div key={i}>
                  {i > 0 && <div className="h-px bg-gray-100 mb-6 -mt-6" />}
                  <div className="flex items-start justify-between mb-2">
                    <span className="font-medium text-gray-900">{proj.name}</span>
                    <span className="text-xs text-gray-400 flex-shrink-0 ml-8">{proj.stack}</span>
                  </div>
                  <p className="text-sm text-gray-600 leading-relaxed mb-2">{proj.desc}</p>
                  <div className="flex items-start gap-1.5">
                    <span className="text-xs text-gray-400 flex-shrink-0 mt-0.5">个人职责：</span>
                    <span className="text-xs text-gray-600">{proj.role}</span>
                  </div>
                </div>
              ))}
            </div>
          </ResumeSection>

          {/* 5. 技能信息 ──────────────────────────────────────────────────── */}
          <ResumeSection icon={<Code2 className="w-4 h-4" />} title="技能信息">
            <div className="space-y-4">
              <div>
                <div className="text-xs text-gray-400 mb-2.5">技术技能</div>
                <div className="flex flex-wrap gap-2">
                  {skills.map((skill) => (
                    <span
                      key={skill}
                      className="px-3 py-1 bg-gray-50 border border-gray-200 rounded text-sm text-gray-700"
                    >
                      {skill}
                    </span>
                  ))}
                </div>
              </div>
              <div className="h-px bg-gray-100" />
              <div>
                <div className="text-xs text-gray-400 mb-2">语言能力</div>
                <div className="text-sm text-gray-700">英语：CET-4</div>
              </div>
            </div>
          </ResumeSection>

          {/* 6. 获奖 / 证书 / 竞赛 ───────────────────────────────────────── */}
          <ResumeSection icon={<Award className="w-4 h-4" />} title="获奖 / 证书 / 竞赛">
            <div className="space-y-3">
              <div className="flex items-start justify-between">
                <div className="text-sm text-gray-800">蓝桥杯省赛奖项</div>
                <div className="text-sm text-gray-400 flex-shrink-0 ml-8">2024</div>
              </div>
              <div className="h-px bg-gray-100" />
              <div className="flex items-start justify-between">
                <div className="text-sm text-gray-800">校级程序设计竞赛奖项</div>
                <div className="text-sm text-gray-400 flex-shrink-0 ml-8">2023</div>
              </div>
            </div>
          </ResumeSection>

          {/* 7. 自我评价 ──────────────────────────────────────────────────── */}
          <ResumeSection icon={<MessageSquare className="w-4 h-4" />} title="自我评价">
            <p className="text-sm text-gray-700 leading-relaxed">
              具备 Go 后端开发基础，熟悉 MySQL、Redis 和常见 Web 后端开发流程，了解微服务、权限控制和接口设计。
              希望从事后端开发、云计算开发等方向，对高并发系统和基础设施有浓厚兴趣。
            </p>
          </ResumeSection>

        </div>

        {/* ── Bottom Action Area — static, in page flow ─────────────────────── */}
        <div className="mt-8 pt-5 border-t border-gray-200 flex items-center justify-between">
          <div className="flex items-center gap-3 text-sm text-gray-500">
            <Clock className="w-3.5 h-3.5" />
            最近更新：2 小时前
            <span className="w-px h-4 bg-gray-200" />
            当前简历：<span className="text-gray-800 font-medium">{activeVersion?.label}</span>
          </div>
          <div className="flex items-center gap-3">
            <Link
              to="/resumes/compare"
              className="px-4 py-1.5 border border-gray-300 text-gray-700 text-sm rounded hover:bg-gray-50 transition-colors flex items-center gap-1.5"
            >
              <ArrowRightLeft className="w-3.5 h-3.5" />
              历史简历比较
            </Link>
            <Link
              to="/resumes/edit"
              className="px-4 py-1.5 border border-[#0052D9] text-[#0052D9] text-sm font-medium rounded hover:bg-blue-50 transition-colors"
            >
              添加简历
            </Link>
            <Link
              to="/resumes/edit"
              className="px-5 py-1.5 bg-[#0052D9] text-white text-sm font-medium rounded hover:bg-[#003DA5] transition-colors"
            >
              修改简历
            </Link>
          </div>
        </div>

      </div>

      <SelectResumeModal
        isOpen={showSelectModal}
        onClose={() => setShowSelectModal(false)}
        scene="match-analysis"
        jobTitle="软件开发-后台开发方向"
      />
    </>
  );
}
