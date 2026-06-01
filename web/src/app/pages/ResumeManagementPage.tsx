import { useState, useEffect } from "react";
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
} from "lucide-react";
import { loadResumeManagementPageViewModel } from "../../api";
import type { ResumeManagementViewModel } from "../../adapters";

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
      <div className="flex items-center gap-2.5 px-8 py-4 border-b border-gray-100">
        <span className="text-[#0052D9]">{icon}</span>
        <h2 className="font-semibold text-gray-900">{title}</h2>
      </div>
      <div className="px-8 py-5">{children}</div>
    </div>
  );
}

// ─── Empty state ──────────────────────────────────────────────────────────────

function EmptyState() {
  return (
    <div className="max-w-[960px] mx-auto px-4 py-8 pb-12">
      <div className="bg-white border border-gray-200 rounded-sm p-12 text-center">
        <FileText className="w-12 h-12 text-gray-300 mx-auto mb-4" />
        <h2 className="text-lg font-semibold text-gray-700 mb-2">你还没有创建简历</h2>
        <p className="text-sm text-gray-500 mb-6">
          添加一份简历后，这里会展示你的教育经历、项目经历、技能与求职方向。
        </p>
        <Link
          to="/resumes/edit"
          className="inline-flex items-center gap-1.5 px-5 py-2 bg-[#0052D9] text-white text-sm font-medium rounded hover:bg-[#003DA5] transition-colors"
        >
          <Plus className="w-4 h-4" />
          添加简历
        </Link>
      </div>
    </div>
  );
}

// ─── Main Component ───────────────────────────────────────────────────────────

export function ResumeManagementPage() {
  const [vm, setVm] = useState<ResumeManagementViewModel | null>(null);
  const [loading, setLoading] = useState(true);
  const [showSelectModal, setShowSelectModal] = useState(false);

  useEffect(() => {
    loadResumeManagementPageViewModel()
      .then(setVm)
      .finally(() => setLoading(false));
  }, []);

  // Loading state
  if (loading) {
    return (
      <div className="max-w-[960px] mx-auto px-4 py-8 pb-12">
        <div className="bg-white border border-gray-200 rounded-sm p-12 text-center">
          <p className="text-sm text-gray-400">加载中…</p>
        </div>
      </div>
    );
  }

  // No resume / error state
  if (!vm || !vm.hasResume) {
    return <EmptyState />;
  }

  const versionLabel = `V${vm.summary.version} ${vm.summary.name}`;

  return (
    <>
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
            <span className="font-medium text-gray-900">{versionLabel}</span>
          </div>
          <div className="w-px h-4 bg-gray-200" />
          <div className="flex items-center gap-2 text-sm text-gray-600">
            <span className="text-gray-400">简历完整度：</span>
            <div className="flex items-center gap-2">
              <div className="w-20 h-1.5 bg-gray-100 rounded-full overflow-hidden">
                <div
                  className="h-full bg-[#0052D9] rounded-full"
                  style={{ width: `${Math.max(vm.summary.completeness, 10)}%` }}
                />
              </div>
              <span className="font-semibold text-[#0052D9]">{vm.summary.completeness}%</span>
            </div>
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
            <div className="flex-1 min-w-0">
              <div className="flex items-center gap-3 mb-2">
                <h2 className="text-white font-semibold text-xl">{vm.displayName}</h2>
                <span className="px-2 py-0.5 bg-white/15 border border-white/25 rounded text-white text-xs">{vm.directionLabel}</span>
              </div>

              <div className="flex items-center gap-4 mb-3">
                {vm.displaySchool !== '暂未填写' && (
                  <>
                    <span className="text-white/70 text-sm">{vm.displaySchool}{vm.displayMajor ? ` · ${vm.displayMajor}` : ''}</span>
                    <span className="w-px h-3 bg-white/30" />
                  </>
                )}
                <span className="text-white/70 text-sm">{vm.directionLabel}</span>
              </div>

              <div className="flex items-center gap-5 mb-3">
                {vm.displayPhone && (
                  <span className="flex items-center gap-1.5 text-white/80 text-sm">
                    <Phone className="w-3.5 h-3.5" />
                    {vm.displayPhone}
                  </span>
                )}
                {vm.displayEmail && (
                  <span className="flex items-center gap-1.5 text-white/80 text-sm">
                    <Mail className="w-3.5 h-3.5" />
                    {vm.displayEmail}
                  </span>
                )}
                <span className="flex items-center gap-1.5 text-white/80 text-sm">
                  <MapPin className="w-3.5 h-3.5" />
                  {vm.displayLocation}
                </span>
              </div>

              <div className="flex items-center gap-2">
                <span className="text-white/50 text-xs">求职方向：</span>
                <span className="text-white text-sm font-medium">{vm.directionLabel}</span>
              </div>
            </div>

            <div className="flex-shrink-0 text-right min-w-[150px]">
              <div className="text-white/50 text-xs mb-1.5">当前简历版本</div>
              <div className="text-white font-medium text-sm mb-2">{versionLabel}</div>
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

        {/* ── AI Status ────────────────────────────────────────────────────── */}
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

          {/* 1. 教育经历 */}
          <ResumeSection icon={<GraduationCap className="w-4 h-4" />} title="教育经历">
            {vm.education.length === 0 ? (
              <p className="text-sm text-gray-400">暂未填写教育经历</p>
            ) : (
              vm.education.map((edu, i) => (
                <div key={i}>
                  {i > 0 && <div className="h-px bg-gray-100 my-4" />}
                  <div className="flex items-start justify-between">
                    <div>
                      <div className="flex items-center gap-3 mb-1.5">
                        <span className="font-medium text-gray-900">{edu.school || '暂未填写'}</span>
                        {edu.degree && <span className="text-gray-400 text-xs">{edu.degree}</span>}
                      </div>
                      {edu.major && <div className="text-sm text-gray-600 mb-2.5">{edu.major}</div>}
                      {edu.courses?.length > 0 && (
                        <div className="text-xs text-gray-400 leading-relaxed">
                          相关课程：{edu.courses.join('、')}
                        </div>
                      )}
                    </div>
                    <div className="text-sm text-gray-400 text-right flex-shrink-0 ml-8">
                      {edu.start_date || ''} {edu.end_date ? `- ${edu.end_date}` : ''}
                    </div>
                  </div>
                </div>
              ))
            )}
          </ResumeSection>

          {/* 2. 实习/工作经历 */}
          <ResumeSection icon={<Briefcase className="w-4 h-4" />} title="实习 / 工作经历">
            {vm.experiences.length === 0 ? (
              <p className="text-sm text-gray-400">暂未填写实习 / 工作经历</p>
            ) : (
              vm.experiences.map((exp, i) => (
                <div key={i}>
                  {i > 0 && <div className="h-px bg-gray-100 my-4" />}
                  <div className="flex items-start justify-between">
                    <div className="flex-1">
                      <div className="flex items-center gap-3 mb-1.5">
                        <span className="font-medium text-gray-900">{exp.company || '暂未填写'}</span>
                        {exp.role && <span className="text-sm text-gray-600">{exp.role}</span>}
                      </div>
                      {exp.description && (
                        <div className="text-sm text-gray-600 leading-relaxed">{exp.description}</div>
                      )}
                    </div>
                    <div className="text-sm text-gray-400 text-right flex-shrink-0 ml-8">
                      {exp.start_date || ''} {exp.end_date ? `- ${exp.end_date}` : ''}
                    </div>
                  </div>
                </div>
              ))
            )}
          </ResumeSection>

          {/* 3. 项目经历 */}
          <ResumeSection icon={<Code2 className="w-4 h-4" />} title="项目经历">
            {vm.projects.length === 0 ? (
              <p className="text-sm text-gray-400">暂未填写项目经历</p>
            ) : (
              <div className="space-y-6">
                {vm.projects.map((proj, i) => (
                  <div key={proj.id || i}>
                    {i > 0 && <div className="h-px bg-gray-100 mb-6 -mt-6" />}
                    <div className="flex items-start justify-between mb-2">
                      <span className="font-medium text-gray-900">{proj.name}</span>
                      <span className="text-xs text-gray-400 flex-shrink-0 ml-8">
                        {proj.tech_stack?.join(' / ') || ''}
                      </span>
                    </div>
                    <p className="text-sm text-gray-600 leading-relaxed mb-2">{proj.description}</p>
                    {proj.responsibilities?.length > 0 && (
                      <div className="flex items-start gap-1.5">
                        <span className="text-xs text-gray-400 flex-shrink-0 mt-0.5">个人职责：</span>
                        <span className="text-xs text-gray-600">{proj.responsibilities.join('、')}</span>
                      </div>
                    )}
                  </div>
                ))}
              </div>
            )}
          </ResumeSection>

          {/* 4. 技能信息 */}
          <ResumeSection icon={<Code2 className="w-4 h-4" />} title="技能信息">
            {vm.skills.length === 0 ? (
              <p className="text-sm text-gray-400">暂未填写技能</p>
            ) : (
              <div className="flex flex-wrap gap-2">
                {vm.skills.map((skill) => (
                  <span
                    key={skill}
                    className="px-3 py-1 bg-gray-50 border border-gray-200 rounded text-sm text-gray-700"
                  >
                    {skill}
                  </span>
                ))}
              </div>
            )}
          </ResumeSection>

          {/* 5. 获奖 / 证书 — kept as placeholder until backend supports */}
          <ResumeSection icon={<Award className="w-4 h-4" />} title="获奖 / 证书 / 竞赛">
            <p className="text-sm text-gray-400">暂未填写获奖 / 证书 / 竞赛信息</p>
          </ResumeSection>

          {/* 6. 自我评价 */}
          <ResumeSection icon={<MessageSquare className="w-4 h-4" />} title="自我评价">
            {vm.selfEvaluation ? (
              <p className="text-sm text-gray-700 leading-relaxed">{vm.selfEvaluation}</p>
            ) : (
              <p className="text-sm text-gray-400">暂未填写自我评价</p>
            )}
          </ResumeSection>

        </div>

        {/* ── Bottom Action Area ───────────────────────────────────────────── */}
        <div className="mt-8 pt-5 border-t border-gray-200 flex items-center justify-between">
          <div className="flex items-center gap-3 text-sm text-gray-500">
            <span>当前简历：</span>
            <span className="text-gray-800 font-medium">{versionLabel}</span>
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
