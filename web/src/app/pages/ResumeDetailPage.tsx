import { useState, useEffect } from "react";
import { Link, useParams, useNavigate } from "react-router";
import {
  ArrowLeft, Pencil, Clock, Sparkles, GraduationCap,
  Briefcase, Award, Link as LinkIcon, FileText, EyeOff,
} from "lucide-react";
import { getResumeById } from "../../api/resumes";
import { isApiClientError } from "../../api/errors";
import type { Resume, ResumeEducation, ResumeProject, ResumeInternship, ResumeLanguageExam, ResumeCertificate, ResumeAward, ResumeWork } from "../../types";
import { ContentPageShell } from "../components/layout/ContentPageShell";
import { PageError } from "../components/PageState";
import { Badge } from "../components/ui/Badge";
import { Button } from "../components/ui/Button";
import { OptimizationEntryModal } from "../components/OptimizationEntryModal";

// ─── Completeness ─────────────────────────────────────────────────────────────

function calculateCompleteness(resume: Resume): number {
  const checks = [
    Boolean(resume.name?.trim()),
    Boolean(resume.real_name?.trim()),
    Boolean(resume.city?.trim()),
    (resume.educations ?? []).some((e) => e.school || e.major || e.degree),
    (resume.projects ?? []).some((p) => p.name || p.role || p.description),
    Boolean(resume.professional_skills?.trim() || resume.ai_application?.trim()),
    (resume.internships ?? []).some((i) => i.company || i.position || i.description),
    Boolean((resume.language_exams ?? []).length || (resume.certificates ?? []).length || (resume.awards ?? []).length),
    Boolean((resume.works ?? []).some((w) => w.name || w.url) || resume.homepage?.trim()),
    Boolean(resume.additional_info?.trim()),
  ];
  return Math.round((checks.filter(Boolean).length / checks.length) * 100);
}

// ─── Sub-components ───────────────────────────────────────────────────────────

function InfoItem({ label, value, full }: { label: string; value?: React.ReactNode; full?: boolean }) {
  return (
    <div className={full ? "" : ""}>
      <div className="text-[11px] font-medium uppercase tracking-wide text-slate-400 mb-0.5">{label}</div>
      <div className="text-sm text-slate-800">{value ?? <span className="text-slate-300">未填写</span>}</div>
    </div>
  );
}

function DetailSection({ icon: Icon, title, children, empty }: { icon: React.ElementType; title: string; children: React.ReactNode; empty?: boolean }) {
  if (empty) {
    return (
      <section className="bg-white rounded-2xl border border-slate-100 p-6">
        <div className="flex items-center gap-2 mb-3">
          <Icon className="w-4 h-4 text-slate-300" />
          <h2 className="text-[14px] font-semibold text-slate-300">{title}</h2>
        </div>
        <div className="bg-slate-50/70 rounded-xl py-8 text-center">
          <EyeOff className="w-8 h-8 text-slate-200 mx-auto mb-2" />
          <p className="text-sm text-slate-400">暂无{title}</p>
        </div>
      </section>
    );
  }
  return (
    <section className="bg-white rounded-2xl border border-slate-100 overflow-hidden">
      <div className="flex items-center gap-2 px-6 py-3.5 border-b border-slate-50 bg-slate-50/50">
        <Icon className="w-4 h-4 text-[#0052D9]" />
        <h2 className="text-[14px] font-semibold text-slate-800">{title}</h2>
      </div>
      <div className="px-6 py-5">{children}</div>
    </section>
  );
}

function TagList({ items }: { items: string[] }) {
  if (!items.length) return null;
  return (
    <div className="flex flex-wrap gap-1.5 mt-2">
      {items.map((item, i) => (
        <span key={i} className="text-[11px] px-2 py-0.5 rounded-md bg-blue-50 text-blue-700 font-medium">{item}</span>
      ))}
    </div>
  );
}

function formatDate(d?: string): string {
  if (!d) return "";
  try {
    const t = new Date(d);
    if (isNaN(t.getTime())) return d;
    return t.toLocaleDateString("zh-CN", { year: "numeric", month: "2-digit", day: "2-digit" }) +
      " " + t.toLocaleTimeString("zh-CN", { hour: "2-digit", minute: "2-digit" });
  } catch { return d; }
}

function dateRange(start?: string, end?: string): string {
  const s = start ? start.replace(/-\d{2}$/, "") : "";
  const e = end ? end.replace(/-\d{2}$/, "") : "至今";
  if (!s && !e) return "";
  return [s, e].filter(Boolean).join(" — ");
}

// ─── Main Component ───────────────────────────────────────────────────────────

export function ResumeDetailPage() {
  const { id } = useParams();
  const navigate = useNavigate();
  const numId = id && /^\d+$/.test(id) ? Number(id) : null;

  const [resume, setResume] = useState<Resume | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [showOptimizationModal, setShowOptimizationModal] = useState(false);

  useEffect(() => {
    if (!numId) { setLoading(false); setError("简历 ID 不合法。"); return; }
    let cancelled = false;
    setLoading(true); setError("");
    getResumeById(numId)
      .then((data) => { if (!cancelled) setResume(data); })
      .catch((err: unknown) => {
        if (cancelled) return;
        if (isApiClientError(err) && (err.status === 404 || err.status === 403)) {
          setError("简历不存在或无权访问。");
        } else if (isApiClientError(err) && err.status === 401) {
          navigate(`/?auth=login&redirect=${encodeURIComponent(`/resumes/detail/${numId}`)}`, { replace: true });
        } else { setError("简历详情加载失败，请稍后重试。"); }
      })
      .finally(() => { if (!cancelled) setLoading(false); });
    return () => { cancelled = true; };
  }, [numId]);

  // ── Loading ────────────────────────────────────────────────────────────────
  if (loading) {
    return (
      <ContentPageShell maxWidthClassName="max-w-4xl">
        <div className="space-y-6 animate-pulse">
          <div className="h-8 bg-slate-200 rounded w-40" />
          <div className="h-64 bg-slate-100 rounded-3xl" />
          <div className="h-40 bg-slate-100 rounded-2xl" />
          <div className="h-40 bg-slate-100 rounded-2xl" />
        </div>
      </ContentPageShell>
    );
  }

  // ── Error ──────────────────────────────────────────────────────────────────
  if (error) {
    return (
      <ContentPageShell maxWidthClassName="max-w-4xl">
        <PageError title="简历详情" description={error}>
          <Link to="/resumes" className="text-[#0052D9] text-sm hover:underline">返回我的简历</Link>
        </PageError>
      </ContentPageShell>
    );
  }

  if (!resume) {
    return (
      <ContentPageShell maxWidthClassName="max-w-4xl">
        <PageError title="简历详情" description="简历不存在或无权访问。">
          <Link to="/resumes" className="text-[#0052D9] text-sm hover:underline">返回我的简历</Link>
        </PageError>
      </ContentPageShell>
    );
  }

  // ── Derived ────────────────────────────────────────────────────────────────
  const completeness = calculateCompleteness(resume);
  const educations = resume.educations ?? [];
  const projects = resume.projects ?? [];
  const internships = resume.internships ?? [];
  const languageExams = resume.language_exams ?? [];
  const certificates = resume.certificates ?? [];
  const awards = resume.awards ?? [];
  const works = resume.works ?? [];

  const hasEducation = educations.some((e) => e.school || e.major || e.degree);
  const hasProjects = projects.some((p) => p.name || p.role || p.description);
  const hasInternships = internships.some((i) => i.company || i.position || i.description);
  const hasCertAwards = languageExams.length > 0 || certificates.some((c) => c.name) || awards.some((a) => a.name);
  const hasWorks = works.some((w) => w.name || w.url) || Boolean(resume.homepage?.trim());
  const hasSkills = Boolean(resume.professional_skills?.trim() || resume.ai_application?.trim() || resume.additional_info?.trim());

  const skillTags = resume.professional_skills
    ? resume.professional_skills.split(/[,，、\n]/).map((s) => s.trim()).filter(Boolean)
    : [];

  // ── Render ─────────────────────────────────────────────────────────────────
  return (<>
    <ContentPageShell maxWidthClassName="max-w-4xl">
      <div className="space-y-6">

        {/* ── Top Bar ─────────────────────────────────────────────────────── */}
        <div className="flex flex-col sm:flex-row sm:items-start sm:justify-between gap-4">
          <div>
            <h1 className="text-[22px] font-semibold text-gray-900">简历详情</h1>
            <p className="mt-1 text-sm text-gray-500">只读查看简历内容，可返回管理页或继续编辑。</p>
          </div>
          <div className="flex items-center gap-2.5 shrink-0 self-start">
            <Link
              to="/resumes"
              className="inline-flex items-center gap-1.5 px-3.5 py-2 text-sm text-slate-600 bg-white border border-slate-200 rounded-lg hover:bg-slate-50 transition-colors"
            >
              <ArrowLeft className="w-3.5 h-3.5" />
              返回我的简历
            </Link>
            <Link
              to={`/resumes/edit?id=${resume.id}`}
              className="inline-flex items-center gap-1.5 px-3.5 py-2 text-sm text-white bg-[#0052D9] rounded-lg hover:bg-[#0047BD] transition-colors"
            >
              <Pencil className="w-3.5 h-3.5" />
              编辑简历
            </Link>
          </div>
        </div>

        {/* ── Summary Row ─────────────────────────────────────────────────── */}
        <div className="flex flex-wrap items-center gap-4 text-xs text-slate-500">
          <div className="flex items-center gap-1.5">
            <Clock className="w-3.5 h-3.5 text-slate-400" />
            最近更新：{formatDate(resume.updated_at) || "—"}
          </div>
          <div className="flex items-center gap-2">
            <span>简历完善度</span>
            <div className="flex items-center gap-1.5">
              <div className="w-28 h-2 bg-slate-200 rounded-full overflow-hidden">
                <div
                  className="h-full rounded-full bg-[#0052D9] transition-all"
                  style={{ width: `${completeness}%` }}
                />
              </div>
              <span className="text-xs font-medium text-slate-600">{completeness}%</span>
            </div>
          </div>
          {resume.is_current && (
            <Badge variant="blue">当前使用</Badge>
          )}
        </div>

        {/* ── Hero Card ───────────────────────────────────────────────────── */}
        <div className="bg-gradient-to-r from-[#0052D9] to-[#2979FF] rounded-3xl p-6 sm:p-8 text-white shadow-lg shadow-blue-500/10">
          <div className="flex flex-col sm:flex-row sm:items-start sm:justify-between gap-4">
            <div className="min-w-0">
              <h2 className="text-xl font-bold mb-1 truncate">{resume.name || "未命名简历"}</h2>
              <div className="grid grid-cols-2 sm:grid-cols-3 gap-x-6 gap-y-3 mt-5">
                <div>
                  <div className="text-[11px] font-medium text-blue-200 uppercase tracking-wide">姓名</div>
                  <div className="text-sm text-white mt-0.5">{resume.real_name || "未填写"}</div>
                </div>
                <div>
                  <div className="text-[11px] font-medium text-blue-200 uppercase tracking-wide">当前城市</div>
                  <div className="text-sm text-white mt-0.5">{resume.city || "未填写"}</div>
                </div>
                <div>
                  <div className="text-[11px] font-medium text-blue-200 uppercase tracking-wide">手机号</div>
                  <div className="text-sm text-white mt-0.5">{resume.phone || "未填写"}</div>
                </div>
                <div>
                  <div className="text-[11px] font-medium text-blue-200 uppercase tracking-wide">邮箱</div>
                  <div className="text-sm text-white mt-0.5">{resume.email || "未填写"}</div>
                </div>
                <div className="col-span-2 sm:col-span-1">
                  <div className="text-[11px] font-medium text-blue-200 uppercase tracking-wide">个人主页</div>
                  <div className="text-sm text-white mt-0.5 truncate">{resume.homepage || "未填写"}</div>
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* ── Skills Summary ───────────────────────────────────────────────── */}
        {hasSkills ? (
          <DetailSection icon={Sparkles} title="能力摘要">
            <div className="space-y-4">
              {resume.professional_skills?.trim() && (
                <div>
                  <div className="text-[11px] font-medium uppercase tracking-wide text-slate-400 mb-1.5">专业技能 / 工具能力</div>
                  {skillTags.length > 0 ? (
                    <TagList items={skillTags} />
                  ) : (
                    <div className="text-sm text-slate-700 leading-relaxed">{resume.professional_skills}</div>
                  )}
                </div>
              )}
              {resume.ai_application?.trim() && (
                <div>
                  <div className="text-[11px] font-medium uppercase tracking-wide text-slate-400 mb-1.5">AI 应用技能</div>
                  <div className="text-sm text-slate-700 leading-relaxed">{resume.ai_application}</div>
                </div>
              )}
              {resume.additional_info?.trim() && (
                <div>
                  <div className="text-[11px] font-medium uppercase tracking-wide text-slate-400 mb-1.5">补充信息</div>
                  <div className="text-sm text-slate-700 leading-relaxed">{resume.additional_info}</div>
                </div>
              )}
            </div>
          </DetailSection>
        ) : (
          <DetailSection icon={Sparkles} title="能力摘要" empty>
            {null}
          </DetailSection>
        )}

        {/* ── Education ────────────────────────────────────────────────────── */}
        <DetailSection icon={GraduationCap} title="教育经历" empty={!hasEducation}>
          {educations.filter((e) => e.school || e.major || e.degree).map((edu, idx) => (
            <div key={edu.id ?? idx} className={idx > 0 ? "border-t border-slate-50 pt-4 mt-4" : ""}>
              <div className="flex flex-wrap items-center gap-2 mb-2">
                <h3 className="text-sm font-semibold text-slate-800">{edu.school || "未填写学校"}</h3>
                {edu.degree && <Badge variant="outline">{edu.degree}</Badge>}
              </div>
              <p className="text-xs text-slate-500 mb-3">
                {[edu.major, edu.department].filter(Boolean).join(" · ")}
                {(edu.start_date || edu.end_date) && ` · ${dateRange(edu.start_date, edu.end_date)}`}
              </p>
              <div className="grid grid-cols-2 sm:grid-cols-3 gap-3">
                {edu.gpa && <InfoItem label="GPA" value={edu.gpa + (edu.gpa_scale ? ` / ${edu.gpa_scale}` : "")} />}
                {edu.rank_text && <InfoItem label="成绩排名" value={edu.rank_text} />}
                {edu.core_courses && <InfoItem label="主修课程" value={edu.core_courses} full />}
                {edu.research_direction && <InfoItem label="研究方向" value={edu.research_direction} />}
                {edu.advisor && <InfoItem label="导师" value={edu.advisor} />}
                {edu.lab && <InfoItem label="实验室" value={edu.lab} />}
                {edu.paper_summary && <InfoItem label="论文 / 学术成果" value={edu.paper_summary} full />}
              </div>
            </div>
          ))}
        </DetailSection>

        {/* ── Projects ─────────────────────────────────────────────────────── */}
        <DetailSection icon={FileText} title="项目经历" empty={!hasProjects}>
          {projects.filter((p) => p.name || p.role || p.description).map((proj, idx) => (
            <div key={proj.id ?? idx} className={idx > 0 ? "border-t border-slate-50 pt-4 mt-4" : ""}>
              <div className="flex flex-wrap items-center gap-2 mb-1.5">
                <h3 className="text-sm font-semibold text-slate-800">{proj.name || "未命名项目"}</h3>
                {proj.role && <Badge variant="secondary">{proj.role}</Badge>}
              </div>
              {(proj.start_date || proj.end_date) && (
                <p className="text-xs text-slate-400 mb-2">{dateRange(proj.start_date, proj.end_date)}</p>
              )}
              {proj.description && <p className="text-sm text-slate-600 leading-relaxed">{proj.description}</p>}
            </div>
          ))}
        </DetailSection>

        {/* ── Internships ──────────────────────────────────────────────────── */}
        <DetailSection icon={Briefcase} title="实习经历" empty={!hasInternships}>
          {internships.filter((i) => i.company || i.position || i.description).map((intern, idx) => (
            <div key={intern.id ?? idx} className={idx > 0 ? "border-t border-slate-50 pt-4 mt-4" : ""}>
              <div className="flex flex-wrap items-center gap-2 mb-1.5">
                <h3 className="text-sm font-semibold text-slate-800">{intern.company || "未填写公司"}</h3>
                {intern.position && <Badge variant="secondary">{intern.position}</Badge>}
              </div>
              {(intern.start_date || intern.end_date) && (
                <p className="text-xs text-slate-400 mb-2">{dateRange(intern.start_date, intern.end_date)}</p>
              )}
              {intern.description && <p className="text-sm text-slate-600 leading-relaxed">{intern.description}</p>}
            </div>
          ))}
        </DetailSection>

        {/* ── Certificates & Awards ────────────────────────────────────────── */}
        <DetailSection icon={Award} title="证书与获奖" empty={!hasCertAwards}>
          {languageExams.length > 0 && (
            <div className="mb-4">
              <h4 className="text-xs font-semibold text-slate-500 uppercase tracking-wide mb-2">外语考试 / 等级</h4>
              <div className="space-y-1.5">
                {languageExams.filter((le) => le.exam_type).map((le, idx) => (
                  <div key={idx} className="flex items-center gap-2 text-sm">
                    <span className="text-slate-700 font-medium">{le.exam_type}</span>
                    {le.score && <span className="text-slate-500">{le.score}</span>}
                  </div>
                ))}
              </div>
            </div>
          )}
          {certificates.filter((c) => c.name).length > 0 && (
            <div className="mb-4">
              <h4 className="text-xs font-semibold text-slate-500 uppercase tracking-wide mb-2">证书</h4>
              <div className="space-y-2">
                {certificates.filter((c) => c.name).map((cert, idx) => (
                  <div key={cert.id ?? idx}>
                    <div className="text-sm font-medium text-slate-700">{cert.name}</div>
                    {cert.description && <div className="text-xs text-slate-500 mt-0.5">{cert.description}</div>}
                  </div>
                ))}
              </div>
            </div>
          )}
          {awards.filter((a) => a.name).length > 0 && (
            <div>
              <h4 className="text-xs font-semibold text-slate-500 uppercase tracking-wide mb-2">获奖</h4>
              <div className="space-y-2">
                {awards.filter((a) => a.name).map((award, idx) => (
                  <div key={award.id ?? idx}>
                    <div className="text-sm font-medium text-slate-700">{award.name}</div>
                    {award.description && <div className="text-xs text-slate-500 mt-0.5">{award.description}</div>}
                  </div>
                ))}
              </div>
            </div>
          )}
        </DetailSection>

        {/* ── Works / Portfolio ────────────────────────────────────────────── */}
        <DetailSection icon={LinkIcon} title="作品链接" empty={!hasWorks}>
          {works.filter((w) => w.name || w.url).map((work, idx) => (
            <div key={work.id ?? idx} className={idx > 0 ? "border-t border-slate-50 pt-3 mt-3" : ""}>
              <div className="text-sm font-medium text-slate-700 mb-0.5">{work.name || "未命名作品"}</div>
              {work.url && (
                <a href={work.url} target="_blank" rel="noreferrer" className="text-xs text-[#0052D9] hover:underline break-all">
                  {work.url}
                </a>
              )}
              {work.password && <div className="text-xs text-slate-400 mt-0.5">密码：{work.password}</div>}
            </div>
          ))}
          {resume.homepage?.trim() && (
            <div className="border-t border-slate-50 pt-3 mt-3">
              <div className="text-xs font-semibold text-slate-500 mb-1">个人主页</div>
              <a href={resume.homepage} target="_blank" rel="noreferrer" className="text-xs text-[#0052D9] hover:underline break-all">
                {resume.homepage}
              </a>
            </div>
          )}
        </DetailSection>

        {/* ── AI Tip ───────────────────────────────────────────────────────── */}
        <div className="bg-blue-50/60 border border-blue-100 rounded-2xl p-5 text-center">
          <div className="flex items-center justify-center gap-2 mb-2">
            <Sparkles className="w-4 h-4 text-[#0052D9]" />
            <span className="text-sm font-medium text-[#0052D9]">这份简历可用于以下功能</span>
          </div>
          <p className="text-xs text-slate-500 mb-4 leading-relaxed max-w-lg mx-auto">
            建议在投递前根据目标岗位继续完善项目经历、技能信息和补充信息，以获得更精准的岗位推荐和简历优化建议。
          </p>
          <div className="flex flex-wrap items-center justify-center gap-2">
            <Link to={`/recommend?resumeId=${resume.id}`} className="text-xs text-[#0052D9] hover:underline">去岗位推荐</Link>
            <span className="text-slate-300">·</span>
            <Link to="/jobs" className="text-xs text-[#0052D9] hover:underline">去岗位广场</Link>
            <span className="text-slate-300">·</span>
            <button onClick={() => setShowOptimizationModal(true)} className="text-xs text-[#0052D9] hover:underline">去简历优化</button>
          </div>
        </div>

      </div>
    </ContentPageShell>
      <OptimizationEntryModal
        open={showOptimizationModal}
        onClose={() => setShowOptimizationModal(false)}
        initialResumeId={numId}
      />
    </>
  );
}
