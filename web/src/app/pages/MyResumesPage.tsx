import React, { useState, useRef, useEffect, useCallback } from "react";
import { Link, useLocation, useNavigate } from "react-router";
import {
  ChevronLeft,
  ChevronUp,
  ChevronDown,
  Trash2,
  Plus,
  Sparkles,
  X,
} from "lucide-react";
import { loadResumeManagementPageViewModel } from "../../api";
import { createResume, updateResume, getCurrentResume, getResumes } from "../../api/resumes";
import type { ResumeManagementViewModel } from "../../adapters";
import type { ResumeEducation, ResumeExperience, ResumeProject } from "../../types";
import { ResumeSkeleton } from "../components/ui/ResumeSkeleton";
import { PageError, PageLoading } from "../components/PageState";
import { getAuthRedirectIf401 } from "../utils/authRedirect";
import { Input } from "../components/ui/input";
import { Textarea } from "../components/ui/textarea";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../components/ui/select";

// ─── K8-5-22-E: V2 Shell components ──────────────────────────────────────────
import {
  ResumeEditorShell,
  ResumeEditorStatusSidebar,
  ResumeEditorSaveCard,
  ResumeBasicsSection,
  BasicInfoSection,
  JobIntentionSection,
  EducationSection,
  ProjectSection,
  SkillsSection,
  InternshipSection,
  CertificatesSection,
  AwardsSection,
  PortfolioExtraSection,
} from "../components/resume-editor";

// ─── Types ────────────────────────────────────────────────────────────────────

interface Project {
  id: number;
  name: string;
  stack: string;
  desc: string;
  role: string;
  // V2 draft fields (K8-5-22-G)
  projectType: string;
  startDate: string;
  endDate: string;
  responsibilities: string;
  achievements: string;
  link: string;
}

/** V1 + V2 draft fields for education entries (K8-5-22-F) */
interface Education {
  id: number;
  school: string;
  major: string;
  degree: string;
  startDate: string;
  endDate: string;
  courses: string;
  // V2 draft fields (K8-5-22-F) — UI only, not in save payload yet
  department: string;
  gpa: string;
  gpaScale: string;
  rank: string;
  coreCourses: string;
  researchDirection: string;
  lab: string;
  paperSummary: string;
}

interface Internship {
  id: number;
  company: string;
  position: string;
  startDate: string;
  endDate: string;
  content: string;
  // V2 draft fields (K8-5-22-G)
  techStack: string;
  achievements: string;
  metrics: string;
}

/** V2 skill categories draft (K8-5-22-G) */
interface SkillsDraft {
  programmingLanguages: string;
  backendFrameworks: string;
  frontendTechnologies: string;
  databases: string;
  middleware: string;
  cloudPlatforms: string;
  devTools: string;
  aiTools: string;
  aiApplicationScenarios: string;
  otherKeywords: string;
}

/** V2 certificate entry draft (K8-5-22-H) */
interface CertificateDraft {
  id: number;
  name: string;
  type: string;
  level: string;
  score: string;
  issuedAt: string;
}

/** V2 award entry draft (K8-5-22-H) */
interface AwardDraft {
  id: number;
  name: string;
  type: string;
  level: string;
  awardedAt: string;
  description: string;
}

/** V2 portfolio links draft (K8-5-22-H) */
interface PortfolioDraft {
  github: string;
  gitee: string;
  cnb: string;
  blog: string;
  demoLinks: string;
  portfolioLinks: string;
  homepage: string;
}

/** V2 extra info draft (K8-5-22-H) */
interface ExtraInfoDraft {
  additionalInfo: string;
}

// ─── Anchor nav sections ───────────────────────────────────────────────────────

const NAV_SECTIONS = [
  { id: "resume-basic", label: "简历基础" },
  { id: "personal-info", label: "基础信息" },
  { id: "job-intention", label: "求职方向" },
  { id: "education", label: "教育经历" },
  { id: "internship", label: "实习经历" },
  { id: "projects", label: "项目经历" },
  { id: "skills", label: "技能关键词" },
  { id: "self-summary", label: "自我评价" },
];

// ─── FormSection wrapper ──────────────────────────────────────────────────────

function FormSection({
  id,
  title,
  children,
  onAdd,
  addLabel,
}: {
  id: string;
  title: string;
  children: React.ReactNode;
  onAdd?: () => void;
  addLabel?: string;
}) {
  return (
    <section id={id} className="bg-white border border-gray-200 scroll-mt-20">
      <div className="flex items-center justify-between px-8 py-4 border-b border-gray-100">
        <h2 className="font-semibold text-gray-900 text-sm">{title}</h2>
        {onAdd && (
          <button
            onClick={onAdd}
            className="flex items-center gap-1 text-xs text-[#0052D9] hover:text-blue-700 transition-colors"
          >
            <Plus className="w-3.5 h-3.5" />
            {addLabel ?? "添加"}
          </button>
        )}
      </div>
      <div className="px-8 py-6">{children}</div>
    </section>
  );
}

// ─── FieldRow ─────────────────────────────────────────────────────────────────

function FieldRow({ label, required, children, hint }: {
  label: string;
  required?: boolean;
  children: React.ReactNode;
  hint?: string;
}) {
  return (
    <div className="flex items-start gap-0">
      <label className="w-[120px] flex-shrink-0 text-sm text-gray-600 pt-2 leading-none">
        {label}
        {required && <span className="text-red-500 ml-0.5">*</span>}
      </label>
      <div className="flex-1">
        {children}
        {hint && <p className="mt-1.5 text-xs text-gray-400">{hint}</p>}
      </div>
    </div>
  );
}

// ─── Main Component ───────────────────────────────────────────────────────────

export function MyResumesPage() {
  const location = useLocation();
  const navigate = useNavigate();
  const mode: "edit" | "add" = location.pathname.includes("/new") ? "add" : "edit";
  const [saving, setSaving] = useState(false);
  const [saveSuccessMsg, setSaveSuccessMsg] = useState("");
  const [saveErrorMsg, setSaveErrorMsg] = useState("");
  const [formVm, setFormVm] = useState<ResumeManagementViewModel | null>(null);
  const [loading, setLoading] = useState(mode === "edit");
  const [resumeError, setResumeError] = useState<string | null>(null);
  const [retryCount, setRetryCount] = useState(0);

  // 编辑模式：加载完整 current resume
  useEffect(() => {
    if (mode === "edit") {
      setLoading(true);
      setResumeError(null);
      loadResumeManagementPageViewModel()
        .then((vm) => { setFormVm(vm); setLoading(false); })
        .catch((err) => {
          const authRedirect = getAuthRedirectIf401(err, location.pathname, location.search);
          if (authRedirect) { navigate(authRedirect, { replace: true }); return; }
          setResumeError("简历加载失败，请稍后重试。"); setLoading(false);
        });
    }
  }, [mode, retryCount]);

  const handleRetryLoad = useCallback(() => setRetryCount((c) => c + 1), []);

  // ── Resume basic ────────────────────────────────────────────────────────────
  const [resumeName, setResumeName] = useState(mode === "add" ? "" : "");
  const [jobDirection, setJobDirection] = useState("");
  const [resumeDesc, setResumeDesc] = useState("");

  // 编辑模式：从 VM 回填基础字段
  useEffect(() => {
    if (formVm && mode === "edit") {
      setResumeName(formVm.summary.name || "");
      setJobDirection(formVm.directionLabel || "");
    }
  }, [formVm, mode]);

  // ── Personal info ───────────────────────────────────────────────────────────
  const [personalName, setPersonalName] = useState("");
  const [age, setAge] = useState("");
  const [degree, setDegree] = useState("本科");
  const [school, setSchool] = useState("");
  const [major, setMajor] = useState("");
  const [gradYear, setGradYear] = useState("");
  const [phone, setPhone] = useState("");
  const [email, setEmail] = useState("");
  // V2 draft (K8-5-22-F): current city
  const [currentCity, setCurrentCity] = useState("");

  // ── Job intention ───────────────────────────────────────────────────────────
  const [direction, setDirection] = useState("");
  const [cities, setCities] = useState("");
  const [jobType, setJobType] = useState("校园招聘 / 实习");
  // V2 draft (K8-5-22-F): new fields
  const [acceptOtherCities, setAcceptOtherCities] = useState(false);
  const [earliestStartDate, setEarliestStartDate] = useState("");
  const [internshipDuration, setInternshipDuration] = useState("");
  const [weeklyAvailableDays, setWeeklyAvailableDays] = useState("");
  const [targetIndustries, setTargetIndustries] = useState("");
  const [targetCompanies, setTargetCompanies] = useState("");

  // ── Education ───────────────────────────────────────────────────────────────
  const [educations, setEducations] = useState<Education[]>([]);

  // ── Internship ──────────────────────────────────────────────────────────────
  const [internships, setInternships] = useState<Internship[]>([]);

  // ── Projects ────────────────────────────────────────────────────────────────
  const [projects, setProjects] = useState<Project[]>([]);

  // ── Skills ──────────────────────────────────────────────────────────────────
  const [skills, setSkills] = useState<string[]>([]);
  const [skillInput, setSkillInput] = useState("");
  // V2 draft (K8-5-22-G): skill categories
  const [skillsDraft, setSkillsDraft] = useState<SkillsDraft>({
    programmingLanguages: "",
    backendFrameworks: "",
    frontendTechnologies: "",
    databases: "",
    middleware: "",
    cloudPlatforms: "",
    devTools: "",
    aiTools: "",
    aiApplicationScenarios: "",
    otherKeywords: "",
  });

  // V2 drafts (K8-5-22-H)
  const [certificateDrafts, setCertificateDrafts] = useState<CertificateDraft[]>([]);
  const [awardDrafts, setAwardDrafts] = useState<AwardDraft[]>([]);
  const [portfolioDraft, setPortfolioDraft] = useState<PortfolioDraft>({
    github: "", gitee: "", cnb: "", blog: "", demoLinks: "", portfolioLinks: "", homepage: "",
  });
  const [extraInfoDraft, setExtraInfoDraft] = useState<ExtraInfoDraft>({ additionalInfo: "" });

  // ── Self summary ─────────────────────────────────────────────────────────────
  const [summary, setSummary] = useState("");

  // 编辑模式：formVm → 表单 state 完整回填
  useEffect(() => {
    if (!formVm || mode !== "edit") return;

    // Personal info
    setPersonalName(formVm.displayName !== "暂未填写" ? formVm.displayName : "");
    setPhone(formVm.displayPhone || "");
    setEmail(formVm.displayEmail || "");
    setSchool(formVm.displaySchool !== "暂未填写" ? formVm.displaySchool : "");
    setMajor(formVm.displayMajor || "");

    // Job intention
    setDirection(formVm.directionLabel !== "暂未填写" ? formVm.directionLabel : "");
    setCities(formVm.displayLocation !== "暂未填写" ? formVm.displayLocation : "");

    // Education
    const eduList: Education[] = (formVm.education || []).map((e: ResumeEducation, i: number) => ({
      id: i + 1,
      school: e.school || "",
      major: e.major || "",
      degree: e.degree || "本科",
      startDate: e.start_date || "",
      endDate: e.end_date || "",
      courses: (e.courses || []).join("、"),
    }));
    setEducations(eduList.length > 0 ? eduList : []);

    // Internship / experience
    const expList: Internship[] = (formVm.experiences || []).map((e: ResumeExperience, i: number) => ({
      id: i + 1,
      company: e.company || "",
      position: e.role || "",
      startDate: e.start_date || "",
      endDate: e.end_date || "",
      content: e.description || "",
    }));
    setInternships(expList);

    // Projects
    const projList: Project[] = (formVm.projects || []).map((p: ResumeProject, i: number) => ({
      id: i + 1,
      name: p.name || "",
      stack: (p.tech_stack || []).join(" / ") || "",
      desc: p.description || "",
      role: (p.responsibilities || []).join("、"),
    }));
    setProjects(projList);

    // Skills
    setSkills(Array.isArray(formVm.skills) ? formVm.skills : []);

    // Self evaluation
    setSummary(formVm.selfEvaluation || "");
  }, [formVm, mode]);

  // ── Save handler ─────────────────────────────────────────────────────────────

  async function handleSave() {
    setSaveSuccessMsg("");
    setSaveErrorMsg("");
    setSaving(true);

    const payload: Record<string, unknown> = {
      name: resumeName || personalName || "未命名简历",
      phone: phone || "",
      email: email || "",
      location: cities?.split(" / ")[0] || "",
      target_city: cities || "",
      target_role: direction || jobDirection || "",
      education: educations.map((e) => ({
        school: e.school,
        degree: e.degree,
        major: e.major,
        start: e.startDate,
        end: e.endDate,
      })),
      experience: internships.map((i) => ({
        company: i.company,
        role: i.position,
        start: i.startDate,
        end: i.endDate,
        description: i.content,
      })),
      project: projects.map((p) => ({
        name: p.name,
        role: p.role,
        description: p.desc,
      })),
      skills: skills.filter((s) => s.trim().length > 0),
      self_evaluation: summary || "",
    };

    try {
      if (mode === "add") {
        await createResume(payload as any);
      } else {
        await updateResume("", payload as any);
      }

      // 保存成功后刷新 current 和 versions，确认后端最新状态
      await Promise.allSettled([getCurrentResume(), getResumes()]);

      setSaveSuccessMsg(
        mode === "add"
          ? "简历已创建成功，当前简历已更新。"
          : "修改已保存，已生成新的简历版本。当前简历已更新。",
      );
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : "";
      if (msg.includes("401") || msg.includes("Unauthorized")) {
        setSaveErrorMsg("登录状态已失效，请重新登录后再保存。");
      } else {
        setSaveErrorMsg("保存失败，请稍后重试。");
      }
    } finally {
      setSaving(false);
    }
  }

  // ── Active anchor ────────────────────────────────────────────────────────────
  const [activeSection, setActiveSection] = useState("resume-basic");

  useEffect(() => {
    const handler = () => {
      for (const s of [...NAV_SECTIONS].reverse()) {
        const el = document.getElementById(s.id);
        if (el && el.getBoundingClientRect().top <= 100) {
          setActiveSection(s.id);
          return;
        }
      }
      setActiveSection("resume-basic");
    };
    window.addEventListener("scroll", handler, { passive: true });
    return () => window.removeEventListener("scroll", handler);
  }, []);

  // ── Helpers ──────────────────────────────────────────────────────────────────

  function scrollTo(id: string) {
    document.getElementById(id)?.scrollIntoView({ behavior: "smooth", block: "start" });
  }

  function updateProject(id: number, field: keyof Project, value: string) {
    setProjects((ps) => ps.map((p) => (p.id === id ? { ...p, [field]: value } : p)));
  }

  function moveProject(index: number, dir: -1 | 1) {
    setProjects((ps) => {
      const next = [...ps];
      const target = index + dir;
      if (target < 0 || target >= next.length) return ps;
      [next[index], next[target]] = [next[target], next[index]];
      return next;
    });
  }

  function removeProject(id: number) {
    setProjects((ps) => ps.filter((p) => p.id !== id));
  }

  function addProject() {
    const newId = Math.max(0, ...projects.map((p) => p.id)) + 1;
    setProjects((ps) => [
      ...ps,
      {
        id: newId, name: "", stack: "", desc: "", role: "",
        projectType: "", startDate: "", endDate: "",
        responsibilities: "", achievements: "", link: "",
      },
    ]);
  }

  function addSkill() {
    const v = skillInput.trim();
    if (v && !skills.includes(v)) setSkills((s) => [...s, v]);
    setSkillInput("");
  }

  function removeSkill(s: string) {
    setSkills((sk) => sk.filter((x) => x !== s));
  }

  function updateSkillsDraft(field: keyof SkillsDraft, value: string) {
    setSkillsDraft((prev) => ({ ...prev, [field]: value }));
  }

  function updatePortfolioDraft(field: keyof PortfolioDraft, value: string) {
    setPortfolioDraft((prev) => ({ ...prev, [field]: value }));
  }

  function addCertificate() {
    const newId = Math.max(0, ...certificateDrafts.map((c) => c.id)) + 1;
    setCertificateDrafts((cs) => [...cs, { id: newId, name: "", type: "", level: "", score: "", issuedAt: "" }]);
  }
  function updateCertificate(id: number, field: keyof CertificateDraft, value: string) {
    setCertificateDrafts((cs) => cs.map((c) => (c.id === id ? { ...c, [field]: value } : c)));
  }
  function removeCertificate(id: number) {
    setCertificateDrafts((cs) => cs.filter((c) => c.id !== id));
  }

  function addAward() {
    const newId = Math.max(0, ...awardDrafts.map((a) => a.id)) + 1;
    setAwardDrafts((as) => [...as, { id: newId, name: "", type: "", level: "", awardedAt: "", description: "" }]);
  }
  function updateAward(id: number, field: keyof AwardDraft, value: string) {
    setAwardDrafts((as) => as.map((a) => (a.id === id ? { ...a, [field]: value } : a)));
  }
  function removeAward(id: number) {
    setAwardDrafts((as) => as.filter((a) => a.id !== id));
  }

  function addEducation() {
    const newId = Math.max(0, ...educations.map((e) => e.id)) + 1;
    setEducations((es) => [
      ...es,
      {
        id: newId, school: "", major: "", degree: "本科",
        startDate: "", endDate: "", courses: "",
        department: "", gpa: "", gpaScale: "", rank: "",
        coreCourses: "", researchDirection: "", lab: "", paperSummary: "",
      },
    ]);
  }

  function updateEdu(id: number, field: keyof Education, value: string) {
    setEducations((es) => es.map((e) => (e.id === id ? { ...e, [field]: value } : e)));
  }

  function addInternship() {
    const newId = Math.max(0, ...internships.map((i) => i.id)) + 1;
    setInternships((is) => [
      ...is,
      {
        id: newId, company: "", position: "", startDate: "", endDate: "",
        content: "", techStack: "", achievements: "", metrics: "",
      },
    ]);
  }

  function updateIntern(id: number, field: keyof Internship, value: string) {
    setInternships((is) => is.map((i) => (i.id === id ? { ...i, [field]: value } : i)));
  }

  // ── K8-5-14-H：编辑模式加载中 → 全页 ResumeSkeleton
  if (mode === "edit" && loading) {
    return <ResumeSkeleton />;
  }

  // ── K8-5-14-H：编辑模式加载失败 → PageError + retry
  if (mode === "edit" && resumeError) {
    return (
      <PageError
        title="简历加载失败"
        description={resumeError}
        onRetry={handleRetryLoad}
      />
    );
  }

  return (
    <ResumeEditorShell
      title={mode === "add" ? "创建简历" : "编辑简历"}
      subtitle="完善你的经历和技能，让 Offerr鹅 更准确地完成岗位推荐、匹配分析和简历优化。"
      sidebar={
        <ResumeEditorStatusSidebar
          completeness={formVm?.summary.completeness ?? 0}
          aiReady={!!formVm?.hasResume}
          missingRequired={0}
          recommendedToFill={0}
          aiFunctions={{
            recommendation: !!formVm?.hasResume,
            matchAnalysis: !!formVm?.hasResume,
            resumeOptimization: !!formVm?.hasResume,
            versionComparison: !!formVm?.hasResume,
          }}
          saving={saving}
        />
      }
      saveCard={
        <ResumeEditorSaveCard
          mode={mode}
          saving={saving}
          successMessage={saveSuccessMsg}
          errorMessage={saveErrorMsg}
          onSave={handleSave}
        />
      }
    >
      {/* ── Back link ─────────────────────────────────────────────── */}
      <Link
        to="/resumes"
        className="inline-flex items-center gap-1 text-xs text-gray-400 hover:text-gray-600 transition-colors -mt-2 mb-1"
      >
        <ChevronLeft className="w-3 h-3" />
        返回我的简历
      </Link>

      {/* ── AI hint strip ──────────────────────────────────────────── */}
      <div className="flex items-center gap-2 px-4 py-2.5 rounded-xl bg-blue-50 border border-blue-100">
        <Sparkles className="w-3.5 h-3.5 text-[#0052D9] flex-shrink-0" />
        <span className="text-xs text-[#0052D9]">
          当前简历可用于：<strong>岗位推荐</strong> / <strong>岗位匹配分析</strong> / <strong>简历优化</strong>
        </span>
        <span className="w-px h-3.5 bg-blue-200 mx-1" />
        <span className="text-xs text-gray-400">AI 建议仅作投递前参考，不替代真实经历核实。</span>
      </div>

      {/* ── Module 1: 简历基础 ──────────────────────────────────────── */}
      <ResumeBasicsSection>
        <div className="space-y-5">
          <div>
            <label className="mb-1.5 block text-sm font-medium text-slate-700">
              简历名称 <span className="text-red-500">*</span>
            </label>
            <Input
              value={resumeName}
              onChange={(e) => setResumeName(e.target.value)}
              placeholder="为这份简历起一个名字，例如「腾讯云后端实习版」"
            />
          </div>
          <div>
            <label className="mb-1.5 block text-sm font-medium text-slate-700">
              目标方向 <span className="text-red-500">*</span>
            </label>
            <Input
              value={jobDirection}
              onChange={(e) => setJobDirection(e.target.value)}
              placeholder="如：后端开发、云计算、产品"
            />
            <p className="mt-1.5 text-xs text-slate-400">会影响岗位推荐、匹配分析和优化建议</p>
          </div>
          <div>
            <label className="mb-1.5 block text-sm font-medium text-slate-700">简历说明</label>
            <Textarea
              className="min-h-[80px] resize-none"
              rows={2}
              value={resumeDesc}
              onChange={(e) => setResumeDesc(e.target.value)}
              placeholder="可写本版本的定位，例如「腾讯云后端实习版」「春招通用版」"
            />
          </div>
          <div>
            <div className="flex items-start gap-2 rounded-xl bg-blue-50/60 border border-blue-100/40 px-4 py-3">
              <Sparkles className="w-4 h-4 text-[#0052D9] mt-0.5" />
              <p className="text-[13px] text-slate-600 leading-relaxed">
                保存后，最新版本将自动用于岗位推荐、岗位匹配分析和简历优化建议。
              </p>
            </div>
          </div>
        </div>
      </ResumeBasicsSection>

      {/* ── Module 2: 基础信息 ──────────────────────────────────────── */}
      <BasicInfoSection>
        <div className="grid grid-cols-1 gap-x-6 gap-y-5 sm:grid-cols-2">
          <div>
            <label className="mb-1.5 block text-sm font-medium text-slate-700">
              姓名 <span className="text-red-500">*</span>
            </label>
            <Input value={personalName} onChange={(e) => setPersonalName(e.target.value)} placeholder="姓名" />
          </div>
          <div>
            <label className="mb-1.5 block text-sm font-medium text-slate-700">年龄</label>
            <Input value={age} onChange={(e) => setAge(e.target.value)} placeholder="年龄" />
          </div>
          <div>
            <label className="mb-1.5 block text-sm font-medium text-slate-700">
              学历 <span className="text-red-500">*</span>
            </label>
            <Select value={degree} onValueChange={setDegree}>
              <SelectTrigger className="sm:max-w-[200px]">
                <SelectValue placeholder="选择学历" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="本科">本科</SelectItem>
                <SelectItem value="硕士">硕士</SelectItem>
                <SelectItem value="博士">博士</SelectItem>
                <SelectItem value="专科">专科</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div>
            <label className="mb-1.5 block text-sm font-medium text-slate-700">毕业时间</label>
            <Input value={gradYear} onChange={(e) => setGradYear(e.target.value)} placeholder="如：2026" />
          </div>
          <div>
            <label className="mb-1.5 block text-sm font-medium text-slate-700">
              学校 <span className="text-red-500">*</span>
            </label>
            <Input value={school} onChange={(e) => setSchool(e.target.value)} placeholder="学校名称" />
          </div>
          <div>
            <label className="mb-1.5 block text-sm font-medium text-slate-700">
              专业 <span className="text-red-500">*</span>
            </label>
            <Input value={major} onChange={(e) => setMajor(e.target.value)} placeholder="专业名称" />
          </div>
          {/* V2 draft (K8-5-22-F): current city */}
          <div>
            <label className="mb-1.5 block text-sm font-medium text-slate-700">当前城市</label>
            <Input value={currentCity} onChange={(e) => setCurrentCity(e.target.value)} placeholder="如：深圳" />
          </div>
        </div>
        <div className="mt-5 pt-5 border-t border-gray-100 grid grid-cols-1 gap-x-6 gap-y-5 sm:grid-cols-2">
          <div>
            <label className="mb-1.5 block text-sm font-medium text-slate-700">手机号</label>
            <Input value={phone} onChange={(e) => setPhone(e.target.value)} placeholder="手机号（脱敏展示）" />
          </div>
          <div>
            <label className="mb-1.5 block text-sm font-medium text-slate-700">邮箱</label>
            <Input value={email} onChange={(e) => setEmail(e.target.value)} placeholder="邮箱地址" />
          </div>
        </div>
        <p className="mt-5 text-[12px] text-slate-400 leading-relaxed border-t border-gray-100 pt-4">
          基础信息用于简历展示和基础画像。手机号 / 邮箱不会进入 AI 分析。
        </p>
      </BasicInfoSection>

      {/* ── Module 3: 求职意向 ──────────────────────────────────────── */}
      <JobIntentionSection>
        {/* V1 + shared fields */}
        <div className="grid grid-cols-1 gap-x-6 gap-y-5 sm:grid-cols-2 mb-5">
          <div>
            <label className="mb-1.5 block text-sm font-medium text-slate-700">
              目标岗位方向 <span className="text-red-500">*</span>
            </label>
            <Input value={direction} onChange={(e) => setDirection(e.target.value)} placeholder="如：后端开发" />
          </div>
          <div>
            <label className="mb-1.5 block text-sm font-medium text-slate-700">期望城市</label>
            <Input value={cities} onChange={(e) => setCities(e.target.value)} placeholder="如：深圳 / 北京 / 上海" />
          </div>
        </div>
        <div className="mb-5">
          <label className="mb-1.5 block text-sm font-medium text-slate-700">岗位类型</label>
          <Select value={jobType} onValueChange={setJobType}>
            <SelectTrigger className="sm:max-w-xs">
              <SelectValue placeholder="选择岗位类型" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="校园招聘 / 实习">校园招聘 / 实习</SelectItem>
              <SelectItem value="校园招聘">校园招聘</SelectItem>
              <SelectItem value="实习">实习</SelectItem>
            </SelectContent>
          </Select>
        </div>

        {/* ── V2 draft fields (K8-5-22-F) ──────────────────────────────── */}
        <div className="border-t border-gray-100 pt-5 space-y-5">
          <h4 className="text-sm font-medium text-slate-600 flex items-center gap-1.5">
            <span className="inline-block w-1 h-3.5 rounded-full bg-amber-400"></span>
            推荐填写 <span className="text-xs font-normal text-slate-400">— 帮助 Offerr鹅 更精确地筛选岗位</span>
          </h4>
          <div className="grid grid-cols-1 gap-x-6 gap-y-5 sm:grid-cols-2">
            <div>
              <label className="mb-1.5 block text-sm font-medium text-slate-700">最早可入职时间</label>
              <Input
                type="date"
                value={earliestStartDate}
                onChange={(e) => setEarliestStartDate(e.target.value)}
              />
            </div>
            <div>
              <label className="mb-1.5 block text-sm font-medium text-slate-700">可实习时长</label>
              <Input
                value={internshipDuration}
                onChange={(e) => setInternshipDuration(e.target.value)}
                placeholder="如：3个月 / 6个月及以上"
              />
            </div>
            <div>
              <label className="mb-1.5 block text-sm font-medium text-slate-700">每周可出勤天数</label>
              <Select
                value={weeklyAvailableDays || undefined}
                onValueChange={(v) => setWeeklyAvailableDays(v)}
              >
                <SelectTrigger className="sm:max-w-[200px]">
                  <SelectValue placeholder="选择天数" />
                </SelectTrigger>
                <SelectContent>
                  {[1, 2, 3, 4, 5, 6, 7].map((d) => (
                    <SelectItem key={d} value={String(d)}>{d} 天</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div>
              <label className="mb-1.5 block text-sm font-medium text-slate-700">是否接受其他城市</label>
              <div className="flex items-center gap-3 pt-1.5">
                <label className="flex items-center gap-1.5 cursor-pointer">
                  <input
                    type="checkbox"
                    checked={acceptOtherCities}
                    onChange={(e) => setAcceptOtherCities(e.target.checked)}
                    className="w-4 h-4 rounded border-gray-300 text-[#0052D9] focus:ring-[#0052D9]"
                  />
                  <span className="text-sm text-slate-700">接受调剂到其他城市</span>
                </label>
              </div>
            </div>
          </div>
        </div>

        {/* ── V2 optional fields (K8-5-22-F) ──────────────────────────────── */}
        <div className="border-t border-gray-100 pt-5 mt-5 space-y-4">
          <h4 className="text-sm font-medium text-slate-500 flex items-center gap-1.5">
            <span className="inline-block w-1 h-3.5 rounded-full bg-slate-300"></span>
            可选 <span className="text-xs font-normal text-slate-400">— 填写后提升匹配精度</span>
          </h4>
          <div>
            <label className="mb-1.5 block text-sm font-medium text-slate-700">目标行业 / 业务方向</label>
            <Input
              value={targetIndustries}
              onChange={(e) => setTargetIndustries(e.target.value)}
              placeholder="如：互联网、金融科技、人工智能（逗号分隔）"
            />
          </div>
          <div>
            <label className="mb-1.5 block text-sm font-medium text-slate-700">目标公司偏好</label>
            <Input
              value={targetCompanies}
              onChange={(e) => setTargetCompanies(e.target.value)}
              placeholder="如：腾讯、字节跳动（逗号分隔）"
            />
          </div>
        </div>

        <p className="mt-5 text-[12px] text-slate-400 leading-relaxed border-t border-gray-100 pt-4">
          求职意向会影响岗位推荐和匹配排序，建议尽可能完整填写。
        </p>
      </JobIntentionSection>

      {/* ── Module 4: 教育经历 ──────────────────────────────────────── */}
      <EducationSection onAdd={addEducation}>
        {educations.length === 0 ? (
          <div className="rounded-xl bg-gray-50/70 py-8 text-center">
            <p className="text-sm text-slate-500 mb-1">暂无教育经历</p>
            <p className="text-xs text-slate-400 mb-3">可以先添加你的学校、专业和学历信息。</p>
            <button onClick={addEducation} className="text-[13px] font-medium text-[#0052D9] hover:underline">+ 添加教育经历</button>
          </div>
        ) : (
          <div className="space-y-4">
            {educations.map((edu, idx) => (
              <div key={edu.id} className="rounded-2xl bg-gray-50/70 ring-1 ring-gray-100 p-5 space-y-4">
                <div className="flex items-center justify-between">
                  <span className="text-xs font-medium text-slate-500">教育经历 {idx + 1}</span>
                  <button onClick={() => setEducations((es) => es.filter((e) => e.id !== edu.id))} className="text-slate-400 hover:text-red-500 transition-colors">
                    <Trash2 className="w-4 h-4" />
                  </button>
                </div>

                {/* Core fields */}
                <div className="grid grid-cols-1 gap-x-5 gap-y-4 sm:grid-cols-2">
                  <div>
                    <label className="mb-1.5 block text-sm font-medium text-slate-700">学校 <span className="text-red-500">*</span></label>
                    <Input className="h-10" value={edu.school} onChange={(e) => updateEdu(edu.id, "school", e.target.value)} placeholder="学校名称" />
                  </div>
                  <div>
                    <label className="mb-1.5 block text-sm font-medium text-slate-700">专业 <span className="text-red-500">*</span></label>
                    <Input className="h-10" value={edu.major} onChange={(e) => updateEdu(edu.id, "major", e.target.value)} placeholder="专业名称" />
                  </div>
                  <div>
                    <label className="mb-1.5 block text-sm font-medium text-slate-700">学历</label>
                    <Select value={edu.degree} onValueChange={(v) => updateEdu(edu.id, "degree", v)}>
                      <SelectTrigger className="h-10"><SelectValue /></SelectTrigger>
                      <SelectContent>
                        <SelectItem value="本科">本科</SelectItem>
                        <SelectItem value="硕士">硕士</SelectItem>
                        <SelectItem value="博士">博士</SelectItem>
                        <SelectItem value="专科">专科</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                  <div>
                    <label className="mb-1.5 block text-sm font-medium text-slate-700">院系</label>
                    <Input className="h-10" value={edu.department} onChange={(e) => updateEdu(edu.id, "department", e.target.value)} placeholder="如：计算机科学与技术学院" />
                  </div>
                  <div>
                    <label className="mb-1.5 block text-sm font-medium text-slate-700">时间</label>
                    <div className="flex items-center gap-2">
                      <Input className="h-10 flex-1" value={edu.startDate} onChange={(e) => updateEdu(edu.id, "startDate", e.target.value)} placeholder="2022.09" />
                      <span className="text-slate-400">—</span>
                      <Input className="h-10 flex-1" value={edu.endDate} onChange={(e) => updateEdu(edu.id, "endDate", e.target.value)} placeholder="2026.06" />
                    </div>
                  </div>
                </div>

                {/* Recommended fields */}
                <div className="border-t border-gray-200 pt-4">
                  <span className="text-xs font-medium text-amber-600 bg-amber-50 px-2 py-0.5 rounded inline-block mb-3">推荐填写</span>
                  <div className="grid grid-cols-1 gap-x-5 gap-y-4 sm:grid-cols-3">
                    <div>
                      <label className="mb-1.5 block text-sm font-medium text-slate-700">GPA</label>
                      <Input className="h-10" value={edu.gpa} onChange={(e) => updateEdu(edu.id, "gpa", e.target.value)} placeholder="如：3.8" />
                    </div>
                    <div>
                      <label className="mb-1.5 block text-sm font-medium text-slate-700">满绩点</label>
                      <Select value={edu.gpaScale || undefined} onValueChange={(v) => updateEdu(edu.id, "gpaScale", v)}>
                        <SelectTrigger className="h-10">
                          <SelectValue placeholder="选择" />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="4.0">4.0</SelectItem>
                          <SelectItem value="5.0">5.0</SelectItem>
                          <SelectItem value="100">100</SelectItem>
                        </SelectContent>
                      </Select>
                    </div>
                    <div>
                      <label className="mb-1.5 block text-sm font-medium text-slate-700">成绩排名</label>
                      <Input className="h-10" value={edu.rank} onChange={(e) => updateEdu(edu.id, "rank", e.target.value)} placeholder="如：前 20%" />
                    </div>
                  </div>
                  <div className="mt-4">
                    <label className="mb-1.5 block text-sm font-medium text-slate-700">主修课程</label>
                    <Input className="h-10" value={edu.coreCourses} onChange={(e) => updateEdu(edu.id, "coreCourses", e.target.value)} placeholder="如：数据结构、操作系统、计算机网络（逗号分隔）" />
                  </div>
                </div>

                {/* Optional fields */}
                <div className="border-t border-gray-200 pt-4">
                  <span className="text-xs font-medium text-slate-400 bg-gray-50 px-2 py-0.5 rounded inline-block mb-3">可选</span>
                  <div className="grid grid-cols-1 gap-x-5 gap-y-4 sm:grid-cols-2">
                    <div>
                      <label className="mb-1.5 block text-sm font-medium text-slate-700">研究方向</label>
                      <Input className="h-10" value={edu.researchDirection} onChange={(e) => updateEdu(edu.id, "researchDirection", e.target.value)} placeholder="如：分布式系统" />
                    </div>
                    <div>
                      <label className="mb-1.5 block text-sm font-medium text-slate-700">导师 / 实验室</label>
                      <Input className="h-10" value={edu.lab} onChange={(e) => updateEdu(edu.id, "lab", e.target.value)} placeholder="如：XX实验室" />
                    </div>
                  </div>
                  <div className="mt-4">
                    <label className="mb-1.5 block text-sm font-medium text-slate-700">论文 / 学术成果</label>
                    <Textarea
                      className="min-h-[72px] resize-none"
                      rows={2}
                      value={edu.paperSummary}
                      onChange={(e) => updateEdu(edu.id, "paperSummary", e.target.value)}
                      placeholder="简要描述发表的论文或参与的科研项目"
                    />
                  </div>
                </div>

                {/* Legacy courses (V1 compatible) */}
                <div className="border-t border-gray-200 pt-4">
                  <label className="mb-1.5 block text-sm font-medium text-slate-700">相关课程</label>
                  <Input className="h-10" value={edu.courses} onChange={(e) => updateEdu(edu.id, "courses", e.target.value)} placeholder="如：数据结构、操作系统、计算机网络" />
                </div>
              </div>
            ))}
          </div>
        )}
      </EducationSection>

      {/* ── Module 5: 项目经历 ──────────────────────────────────────── */}
      <ProjectSection>
        {projects.length === 0 ? (
          <div className="rounded-xl bg-gray-50/70 py-8 text-center">
            <p className="text-sm text-slate-500 mb-1">暂无项目经历</p>
            <p className="text-xs text-slate-400 mb-3">可以补充课程项目、比赛项目或个人项目。</p>
            <button onClick={addProject} className="text-[13px] font-medium text-[#0052D9] hover:underline">+ 添加项目经历</button>
          </div>
        ) : (
          <div className="space-y-4">
            {projects.map((proj, idx) => (
              <div key={proj.id} className="rounded-2xl bg-gray-50/70 ring-1 ring-gray-100 p-5 space-y-4">
                <div className="flex items-center justify-between">
                  <span className="text-xs font-medium text-slate-500">项目经历 {idx + 1}</span>
                  <div className="flex items-center gap-1">
                    <button onClick={() => moveProject(idx, -1)} disabled={idx === 0} className="p-1 text-slate-400 hover:text-slate-600 disabled:opacity-30 transition-colors" title="上移"><ChevronUp className="w-3.5 h-3.5" /></button>
                    <button onClick={() => moveProject(idx, 1)} disabled={idx === projects.length - 1} className="p-1 text-slate-400 hover:text-slate-600 disabled:opacity-30 transition-colors" title="下移"><ChevronDown className="w-3.5 h-3.5" /></button>
                    <button onClick={() => removeProject(proj.id)} className="p-1 text-slate-400 hover:text-red-500 transition-colors" title="删除"><Trash2 className="w-3.5 h-3.5" /></button>
                  </div>
                </div>
                <div>
                  <label className="mb-1.5 block text-sm font-medium text-slate-700">项目名称 <span className="text-red-500">*</span></label>
                  <Input className="h-10" value={proj.name} onChange={(e) => updateProject(proj.id, "name", e.target.value)} placeholder="项目名称" />
                </div>
                <div className="grid grid-cols-1 gap-x-5 gap-y-4 sm:grid-cols-2">
                  <div>
                    <label className="mb-1.5 block text-sm font-medium text-slate-700">项目类型</label>
                    <Select value={proj.projectType || undefined} onValueChange={(v) => updateProject(proj.id, "projectType", v)}>
                      <SelectTrigger className="h-10"><SelectValue placeholder="选择类型" /></SelectTrigger>
                      <SelectContent>
                        <SelectItem value="课程设计">课程设计</SelectItem>
                        <SelectItem value="比赛项目">比赛项目</SelectItem>
                        <SelectItem value="个人项目">个人项目</SelectItem>
                        <SelectItem value="企业项目">企业项目</SelectItem>
                        <SelectItem value="开源项目">开源项目</SelectItem>
                        <SelectItem value="其他">其他</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                  <div>
                    <label className="mb-1.5 block text-sm font-medium text-slate-700">项目时间</label>
                    <div className="flex items-center gap-2">
                      <Input className="h-10 flex-1" value={proj.startDate} onChange={(e) => updateProject(proj.id, "startDate", e.target.value)} placeholder="开始" />
                      <span className="text-slate-400">—</span>
                      <Input className="h-10 flex-1" value={proj.endDate} onChange={(e) => updateProject(proj.id, "endDate", e.target.value)} placeholder="结束" />
                    </div>
                  </div>
                </div>
                <div>
                  <label className="mb-1.5 block text-sm font-medium text-slate-700">技术栈</label>
                  <Input className="h-10" value={proj.stack} onChange={(e) => updateProject(proj.id, "stack", e.target.value)} placeholder="如：Go / MySQL / Redis" />
                </div>
                <div>
                  <label className="mb-1.5 block text-sm font-medium text-slate-700">项目描述</label>
                  <Textarea className="min-h-[88px] resize-none" rows={3} value={proj.desc} onChange={(e) => updateProject(proj.id, "desc", e.target.value)} placeholder="描述项目背景、你负责的部分和主要成果" />
                </div>
                <div>
                  <label className="mb-1.5 block text-sm font-medium text-slate-700">个人职责</label>
                  <Input className="h-10" value={proj.role} onChange={(e) => updateProject(proj.id, "role", e.target.value)} placeholder="如：接口设计、数据库设计、缓存优化" />
                </div>

                {/* V2 recommended fields */}
                <div className="border-t border-gray-200 pt-4">
                  <span className="text-xs font-medium text-amber-600 bg-amber-50 px-2 py-0.5 rounded inline-block mb-3">推荐填写</span>
                  <div className="grid grid-cols-1 gap-x-5 gap-y-4 sm:grid-cols-2">
                    <div>
                      <label className="mb-1.5 block text-sm font-medium text-slate-700">技术栈（详细）</label>
                      <Textarea
                        className="min-h-[72px] resize-none" rows={2}
                        value={proj.responsibilities}
                        onChange={(e) => updateProject(proj.id, "responsibilities", e.target.value)}
                        placeholder="详细描述使用的技术和承担的角色"
                      />
                    </div>
                    <div>
                      <label className="mb-1.5 block text-sm font-medium text-slate-700">项目成果</label>
                      <Textarea
                        className="min-h-[72px] resize-none" rows={2}
                        value={proj.achievements}
                        onChange={(e) => updateProject(proj.id, "achievements", e.target.value)}
                        placeholder="一行一个成果，例如：&#10;QPS 提升 30%&#10;用户量增长 2000+"
                      />
                    </div>
                  </div>
                </div>

                {/* V2 optional */}
                <div className="border-t border-gray-200 pt-4">
                  <span className="text-xs font-medium text-slate-400 bg-gray-50 px-2 py-0.5 rounded inline-block mb-3">可选</span>
                  <div>
                    <label className="mb-1.5 block text-sm font-medium text-slate-700">项目链接</label>
                    <Input className="h-10" value={proj.link} onChange={(e) => updateProject(proj.id, "link", e.target.value)} placeholder="如：https://github.com/xxx" />
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </ProjectSection>

      {/* ── Module 6: 技能信息 ──────────────────────────────────────── */}
      <SkillsSection>
        {/* V1 legacy skills as tags */}
        {skills.length > 0 && (
          <div className="flex flex-wrap gap-2 mb-4">
            {skills.map((s) => (
              <span
                key={s}
                className="inline-flex items-center gap-1.5 rounded-full bg-blue-50 px-3 py-1 text-[13px] text-[#0052D9] ring-1 ring-blue-100"
              >
                {s}
                <button onClick={() => removeSkill(s)} className="text-blue-300 hover:text-blue-600">
                  <X className="w-3 h-3" />
                </button>
              </span>
            ))}
          </div>
        )}
        <div className="flex gap-2 mb-6">
          <Input className="h-10 flex-1" value={skillInput} onChange={(e) => setSkillInput(e.target.value)} onKeyDown={(e) => e.key === "Enter" && addSkill()} placeholder="输入技能关键词后按回车添加" />
          <button onClick={addSkill} className="rounded-xl bg-[#0052D9] px-5 text-[13px] font-medium text-white hover:bg-[#0047BD] transition-colors flex-shrink-0">添加</button>
        </div>

        {/* V2 分类技能 (K8-5-22-G) */}
        <div className="border-t border-gray-200 pt-5 space-y-5">
          <h4 className="text-sm font-medium text-slate-600 flex items-center gap-1.5">
            <span className="inline-block w-1 h-3.5 rounded-full bg-blue-500"></span>
            常用技术能力
          </h4>
          <div className="grid grid-cols-1 gap-x-5 gap-y-4 sm:grid-cols-2">
            <div>
              <label className="mb-1.5 block text-sm font-medium text-slate-700">编程语言</label>
              <Input className="h-10" value={skillsDraft.programmingLanguages} onChange={(e) => updateSkillsDraft("programmingLanguages", e.target.value)} placeholder="如：Go、Python、Java、C++（逗号分隔）" />
            </div>
            <div>
              <label className="mb-1.5 block text-sm font-medium text-slate-700">后端框架</label>
              <Input className="h-10" value={skillsDraft.backendFrameworks} onChange={(e) => updateSkillsDraft("backendFrameworks", e.target.value)} placeholder="如：Gin、Spring Boot、Django（逗号分隔）" />
            </div>
            <div>
              <label className="mb-1.5 block text-sm font-medium text-slate-700">数据库</label>
              <Input className="h-10" value={skillsDraft.databases} onChange={(e) => updateSkillsDraft("databases", e.target.value)} placeholder="如：MySQL、Redis、MongoDB（逗号分隔）" />
            </div>
            <div>
              <label className="mb-1.5 block text-sm font-medium text-slate-700">中间件</label>
              <Input className="h-10" value={skillsDraft.middleware} onChange={(e) => updateSkillsDraft("middleware", e.target.value)} placeholder="如：Kafka、RabbitMQ、Nginx（逗号分隔）" />
            </div>
            <div>
              <label className="mb-1.5 block text-sm font-medium text-slate-700">云平台</label>
              <Input className="h-10" value={skillsDraft.cloudPlatforms} onChange={(e) => updateSkillsDraft("cloudPlatforms", e.target.value)} placeholder="如：腾讯云、AWS、Docker、K8s（逗号分隔）" />
            </div>
          </div>
        </div>

        <div className="border-t border-gray-200 pt-5 space-y-4">
          <h4 className="text-sm font-medium text-slate-600 flex items-center gap-1.5">
            <span className="inline-block w-1 h-3.5 rounded-full bg-slate-400"></span>
            工程与工具 <span className="text-xs font-normal text-slate-400">— 可选</span>
          </h4>
          <div className="grid grid-cols-1 gap-x-5 gap-y-4 sm:grid-cols-2">
            <div>
              <label className="mb-1.5 block text-sm font-medium text-slate-700">前端技术</label>
              <Input className="h-10" value={skillsDraft.frontendTechnologies} onChange={(e) => updateSkillsDraft("frontendTechnologies", e.target.value)} placeholder="如：React、Vue、TypeScript（逗号分隔）" />
            </div>
            <div>
              <label className="mb-1.5 block text-sm font-medium text-slate-700">开发工具</label>
              <Input className="h-10" value={skillsDraft.devTools} onChange={(e) => updateSkillsDraft("devTools", e.target.value)} placeholder="如：Git、Linux、Postman（逗号分隔）" />
            </div>
          </div>
          <div>
            <label className="mb-1.5 block text-sm font-medium text-slate-700">其他技能关键词</label>
            <Input className="h-10" value={skillsDraft.otherKeywords} onChange={(e) => updateSkillsDraft("otherKeywords", e.target.value)} placeholder="其他未归类的技能（逗号分隔）" />
          </div>
        </div>

        <div className="border-t border-gray-200 pt-5 space-y-4">
          <h4 className="text-sm font-medium text-amber-600 flex items-center gap-1.5">
            <span className="inline-block w-1 h-3.5 rounded-full bg-amber-400"></span>
            AI 应用能力 <span className="text-xs font-normal text-amber-500">— 推荐填写</span>
          </h4>
          <div>
            <label className="mb-1.5 block text-sm font-medium text-slate-700">AI 工具</label>
            <Input className="h-10" value={skillsDraft.aiTools} onChange={(e) => updateSkillsDraft("aiTools", e.target.value)} placeholder="如：CodeBuddy、Cursor、ChatGPT、Copilot（逗号分隔）" />
          </div>
          <div>
            <label className="mb-1.5 block text-sm font-medium text-slate-700">AI 应用场景</label>
            <Textarea
              className="min-h-[72px] resize-none" rows={2}
              value={skillsDraft.aiApplicationScenarios}
              onChange={(e) => updateSkillsDraft("aiApplicationScenarios", e.target.value)}
              placeholder="例如：使用 CodeBuddy 辅助需求拆解、代码生成、Bug 定位、接口文档、测试用例和项目复盘。"
            />
          </div>
        </div>

        <p className="mt-5 text-[12px] text-slate-400 border-t border-gray-100 pt-4">
          技能分类越详细，AI 匹配精度越高。V1 标签输入仍可用于快速添加，随后可在分类中细化。
        </p>
      </SkillsSection>

      {/* ── Module 7: 实习经历 ──────────────────────────────────────── */}
      <InternshipSection>
        {internships.length === 0 ? (
          <div className="rounded-xl bg-gray-50/70 py-8 text-center">
            <p className="text-sm text-slate-500 mb-1">暂无实习经历</p>
            <p className="text-xs text-slate-400 mb-3">可以补充实习公司、岗位职责和成果。</p>
            <button onClick={addInternship} className="text-[13px] font-medium text-[#0052D9] hover:underline">+ 添加实习经历</button>
          </div>
        ) : (
          <div className="space-y-4">
            {internships.map((intern, idx) => (
              <div key={intern.id} className="rounded-2xl bg-gray-50/70 ring-1 ring-gray-100 p-5 space-y-4">
                <div className="flex items-center justify-between">
                  <span className="text-xs font-medium text-slate-500">实习经历 {idx + 1}</span>
                  <button onClick={() => setInternships((is) => is.filter((i) => i.id !== intern.id))} className="text-slate-400 hover:text-red-500 transition-colors">
                    <Trash2 className="w-4 h-4" />
                  </button>
                </div>
                <div className="grid grid-cols-1 gap-x-5 gap-y-4 sm:grid-cols-2">
                  <div>
                    <label className="mb-1.5 block text-sm font-medium text-slate-700">公司名称 <span className="text-red-500">*</span></label>
                    <Input className="h-10" value={intern.company} onChange={(e) => updateIntern(intern.id, "company", e.target.value)} placeholder="公司名称" />
                  </div>
                  <div>
                    <label className="mb-1.5 block text-sm font-medium text-slate-700">岗位名称 <span className="text-red-500">*</span></label>
                    <Input className="h-10" value={intern.position} onChange={(e) => updateIntern(intern.id, "position", e.target.value)} placeholder="岗位名称" />
                  </div>
                </div>
                <div>
                  <label className="mb-1.5 block text-sm font-medium text-slate-700">时间</label>
                  <div className="flex items-center gap-2">
                    <Input className="h-10 flex-1" value={intern.startDate} onChange={(e) => updateIntern(intern.id, "startDate", e.target.value)} placeholder="2025.06" />
                    <span className="text-slate-400">—</span>
                    <Input className="h-10 flex-1" value={intern.endDate} onChange={(e) => updateIntern(intern.id, "endDate", e.target.value)} placeholder="2025.09" />
                  </div>
                </div>
                <div>
                  <label className="mb-1.5 block text-sm font-medium text-slate-700">工作内容</label>
                  <Textarea className="min-h-[96px] resize-none" rows={3} value={intern.content} onChange={(e) => updateIntern(intern.id, "content", e.target.value)} placeholder="简要描述工作内容和成果" />
                </div>

                {/* V2 recommended fields (K8-5-22-G) */}
                <div className="border-t border-gray-200 pt-4">
                  <span className="text-xs font-medium text-amber-600 bg-amber-50 px-2 py-0.5 rounded inline-block mb-3">推荐填写</span>
                  <div className="grid grid-cols-1 gap-x-5 gap-y-4 sm:grid-cols-2">
                    <div>
                      <label className="mb-1.5 block text-sm font-medium text-slate-700">使用技术</label>
                      <Input className="h-10" value={intern.techStack} onChange={(e) => updateIntern(intern.id, "techStack", e.target.value)} placeholder="如：Go、MySQL、Redis（逗号分隔）" />
                    </div>
                    <div>
                      <label className="mb-1.5 block text-sm font-medium text-slate-700">量化结果</label>
                      <Input className="h-10" value={intern.metrics} onChange={(e) => updateIntern(intern.id, "metrics", e.target.value)} placeholder="如：处理 10 万+ 请求/天" />
                    </div>
                  </div>
                  <div className="mt-4">
                    <label className="mb-1.5 block text-sm font-medium text-slate-700">核心成果</label>
                    <Textarea
                      className="min-h-[72px] resize-none" rows={2}
                      value={intern.achievements}
                      onChange={(e) => updateIntern(intern.id, "achievements", e.target.value)}
                      placeholder="一行一个成果，例如：&#10;成功优化核心接口，P99 延迟降低 40%&#10;独立完成 XX 模块从 0 到 1 的开发"
                    />
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </InternshipSection>

      {/* ── Module 8: 证书与语言能力 (K8-5-22-H) ──────────────────────── */}
      <CertificatesSection onAdd={addCertificate}>
        {certificateDrafts.length === 0 ? (
          <div className="rounded-xl bg-gray-50/70 py-8 text-center">
            <p className="text-sm text-slate-500 mb-1">暂无证书或语言成绩</p>
            <p className="text-xs text-slate-400 mb-3">
              可以添加英语四六级、雅思/托福、云计算认证（如 AWS/阿里云/腾讯云）、计算机等级考试、软考等证书。
            </p>
            <button onClick={addCertificate} className="text-[13px] font-medium text-[#0052D9] hover:underline">+ 添加证书</button>
          </div>
        ) : (
          <div className="space-y-4">
            {certificateDrafts.map((cert, idx) => (
              <div key={cert.id} className="rounded-2xl bg-gray-50/70 ring-1 ring-gray-100 p-5 space-y-4">
                <div className="flex items-center justify-between">
                  <span className="text-xs font-medium text-slate-500">证书 {idx + 1}</span>
                  <button onClick={() => removeCertificate(cert.id)} className="text-slate-400 hover:text-red-500 transition-colors"><Trash2 className="w-4 h-4" /></button>
                </div>
                <div>
                  <label className="mb-1.5 block text-sm font-medium text-slate-700">证书名称 <span className="text-red-500">*</span></label>
                  <Input className="h-10" value={cert.name} onChange={(e) => updateCertificate(cert.id, "name", e.target.value)} placeholder="如：CET-4 503、阿里云 ACP 云计算工程师" />
                </div>
                <div className="grid grid-cols-1 gap-x-5 gap-y-4 sm:grid-cols-2">
                  <div>
                    <label className="mb-1.5 block text-sm font-medium text-slate-700">证书类型</label>
                    <Select value={cert.type || undefined} onValueChange={(v) => updateCertificate(cert.id, "type", v)}>
                      <SelectTrigger className="h-10"><SelectValue placeholder="选择类型" /></SelectTrigger>
                      <SelectContent>
                        <SelectItem value="语言能力">语言能力</SelectItem>
                        <SelectItem value="云计算认证">云计算认证</SelectItem>
                        <SelectItem value="软件/计算机认证">软件/计算机认证</SelectItem>
                        <SelectItem value="职业资格">职业资格</SelectItem>
                        <SelectItem value="其他">其他</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                  <div>
                    <label className="mb-1.5 block text-sm font-medium text-slate-700">证书级别</label>
                    <Input className="h-10" value={cert.level} onChange={(e) => updateCertificate(cert.id, "level", e.target.value)} placeholder="如：初级、中级、高级" />
                  </div>
                </div>
                <div className="grid grid-cols-1 gap-x-5 gap-y-4 sm:grid-cols-2">
                  <div>
                    <label className="mb-1.5 block text-sm font-medium text-slate-700">成绩 / 分数</label>
                    <Input className="h-10" value={cert.score} onChange={(e) => updateCertificate(cert.id, "score", e.target.value)} placeholder="如：503、80" />
                  </div>
                  <div>
                    <label className="mb-1.5 block text-sm font-medium text-slate-700">获得时间</label>
                    <Input className="h-10" type="month" value={cert.issuedAt} onChange={(e) => updateCertificate(cert.id, "issuedAt", e.target.value)} />
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </CertificatesSection>

      {/* ── Module 9: 获奖与竞赛 (K8-5-22-H) ──────────────────────────── */}
      <AwardsSection onAdd={addAward}>
        {awardDrafts.length === 0 ? (
          <div className="rounded-xl bg-gray-50/70 py-8 text-center">
            <p className="text-sm text-slate-500 mb-1">暂无获奖或竞赛经历</p>
            <p className="text-xs text-slate-400 mb-3">
              可以添加奖学金（如国家奖学金）、竞赛获奖（如 ACM/数模/挑战杯）、荣誉称号（如优秀学生干部）和科研奖项。
            </p>
            <button onClick={addAward} className="text-[13px] font-medium text-[#0052D9] hover:underline">+ 添加获奖</button>
          </div>
        ) : (
          <div className="space-y-4">
            {awardDrafts.map((award, idx) => (
              <div key={award.id} className="rounded-2xl bg-gray-50/70 ring-1 ring-gray-100 p-5 space-y-4">
                <div className="flex items-center justify-between">
                  <span className="text-xs font-medium text-slate-500">获奖 {idx + 1}</span>
                  <button onClick={() => removeAward(award.id)} className="text-slate-400 hover:text-red-500 transition-colors"><Trash2 className="w-4 h-4" /></button>
                </div>
                <div>
                  <label className="mb-1.5 block text-sm font-medium text-slate-700">奖项名称 <span className="text-red-500">*</span></label>
                  <Input className="h-10" value={award.name} onChange={(e) => updateAward(award.id, "name", e.target.value)} placeholder="如：国家奖学金、ACM-ICPC 银牌" />
                </div>
                <div className="grid grid-cols-1 gap-x-5 gap-y-4 sm:grid-cols-2">
                  <div>
                    <label className="mb-1.5 block text-sm font-medium text-slate-700">奖项类型</label>
                    <Select value={award.type || undefined} onValueChange={(v) => updateAward(award.id, "type", v)}>
                      <SelectTrigger className="h-10"><SelectValue placeholder="选择类型" /></SelectTrigger>
                      <SelectContent>
                        <SelectItem value="奖学金">奖学金</SelectItem>
                        <SelectItem value="竞赛获奖">竞赛获奖</SelectItem>
                        <SelectItem value="荣誉称号">荣誉称号</SelectItem>
                        <SelectItem value="科研奖项">科研奖项</SelectItem>
                        <SelectItem value="其他">其他</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                  <div>
                    <label className="mb-1.5 block text-sm font-medium text-slate-700">奖项级别</label>
                    <Select value={award.level || undefined} onValueChange={(v) => updateAward(award.id, "level", v)}>
                      <SelectTrigger className="h-10"><SelectValue placeholder="选择级别" /></SelectTrigger>
                      <SelectContent>
                        <SelectItem value="校级">校级</SelectItem>
                        <SelectItem value="市级">市级</SelectItem>
                        <SelectItem value="省级">省级</SelectItem>
                        <SelectItem value="国家级">国家级</SelectItem>
                        <SelectItem value="国际级">国际级</SelectItem>
                        <SelectItem value="其他">其他</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                </div>
                <div className="grid grid-cols-1 gap-x-5 gap-y-4 sm:grid-cols-2">
                  <div>
                    <label className="mb-1.5 block text-sm font-medium text-slate-700">获奖时间</label>
                    <Input className="h-10" type="month" value={award.awardedAt} onChange={(e) => updateAward(award.id, "awardedAt", e.target.value)} />
                  </div>
                </div>
                <div>
                  <label className="mb-1.5 block text-sm font-medium text-slate-700">奖项说明</label>
                  <Textarea className="min-h-[72px] resize-none" rows={2} value={award.description} onChange={(e) => updateAward(award.id, "description", e.target.value)} placeholder="简要描述获奖原因或竞赛内容" />
                </div>
              </div>
            ))}
          </div>
        )}
      </AwardsSection>

      {/* ── Module 10: 作品链接与补充信息 (K8-5-22-H) ──────────────────── */}
      <PortfolioExtraSection>
        {/* V1 self evaluation */}
        <div className="mb-6">
          <label className="mb-1.5 block text-sm font-medium text-slate-700">自我评价</label>
          <p className="text-[13px] text-slate-500 mb-5">用一小段话概括你的优势、学习能力和求职方向。建议控制在 100-300 字。</p>
          <Textarea className="min-h-[140px]" value={summary} onChange={(e) => setSummary(e.target.value)} placeholder="简要介绍自己的技术背景、求职意向和个人优势" />
          <div className="mt-2 flex items-center justify-between">
            <span className="text-[12px] text-slate-400">建议包含：技术方向、项目经验、学习能力、目标岗位。</span>
            <span className="text-[12px] text-slate-400">当前约 {summary.length} 字</span>
          </div>
        </div>

        {/* V2 portfolio links */}
        <div className="border-t border-gray-200 pt-5 space-y-5">
          <h4 className="text-sm font-medium text-amber-600 flex items-center gap-1.5">
            <span className="inline-block w-1 h-3.5 rounded-full bg-amber-400"></span>
            作品与个人链接 <span className="text-xs font-normal text-amber-500">— 推荐填写</span>
          </h4>
          <p className="text-[12px] text-slate-400 -mt-3">只支持链接，不支持附件上传。</p>
          <div className="grid grid-cols-1 gap-x-5 gap-y-4 sm:grid-cols-2">
            <div><label className="mb-1.5 block text-sm font-medium text-slate-700">GitHub</label><Input className="h-10" value={portfolioDraft.github} onChange={(e) => updatePortfolioDraft("github", e.target.value)} placeholder="https://github.com/xxx" /></div>
            <div><label className="mb-1.5 block text-sm font-medium text-slate-700">CNB</label><Input className="h-10" value={portfolioDraft.cnb} onChange={(e) => updatePortfolioDraft("cnb", e.target.value)} placeholder="https://cnb.cool/xxx" /></div>
            <div><label className="mb-1.5 block text-sm font-medium text-slate-700">个人博客</label><Input className="h-10" value={portfolioDraft.blog} onChange={(e) => updatePortfolioDraft("blog", e.target.value)} placeholder="https://xxx" /></div>
            <div><label className="mb-1.5 block text-sm font-medium text-slate-700">个人主页</label><Input className="h-10" value={portfolioDraft.homepage} onChange={(e) => updatePortfolioDraft("homepage", e.target.value)} placeholder="https://xxx" /></div>
          </div>
        </div>

        <div className="border-t border-gray-200 pt-5 mt-5 space-y-4">
          <h4 className="text-sm font-medium text-slate-500 flex items-center gap-1.5">
            <span className="inline-block w-1 h-3.5 rounded-full bg-slate-300"></span>
            更多链接 <span className="text-xs font-normal text-slate-400">— 可选</span>
          </h4>
          <div>
            <label className="mb-1.5 block text-sm font-medium text-slate-700">Gitee</label>
            <Input className="h-10" value={portfolioDraft.gitee} onChange={(e) => updatePortfolioDraft("gitee", e.target.value)} placeholder="https://gitee.com/xxx" />
          </div>
          <div>
            <label className="mb-1.5 block text-sm font-medium text-slate-700">项目演示链接</label>
            <Textarea className="min-h-[72px] resize-none" rows={2} value={portfolioDraft.demoLinks} onChange={(e) => updatePortfolioDraft("demoLinks", e.target.value)} placeholder="一行一个链接，例如：&#10;https://demo.example.com&#10;https://xxx" />
          </div>
          <div>
            <label className="mb-1.5 block text-sm font-medium text-slate-700">作品集链接</label>
            <Textarea className="min-h-[72px] resize-none" rows={2} value={portfolioDraft.portfolioLinks} onChange={(e) => updatePortfolioDraft("portfolioLinks", e.target.value)} placeholder="一行一个链接" />
          </div>
        </div>

        <div className="border-t border-gray-200 pt-5 mt-5 space-y-4">
          <h4 className="text-sm font-medium text-amber-600 flex items-center gap-1.5">
            <span className="inline-block w-1 h-3.5 rounded-full bg-amber-400"></span>
            补充信息 <span className="text-xs font-normal text-amber-500">— 推荐填写</span>
          </h4>
          <div>
            <label className="mb-1.5 block text-sm font-medium text-slate-700">补充说明</label>
            <Textarea className="min-h-[88px] resize-none" rows={3} value={extraInfoDraft.additionalInfo} onChange={(e) => setExtraInfoDraft({ additionalInfo: e.target.value })} placeholder="补充其他未覆盖的信息，如社团经历、志愿服务、兴趣爱好等" />
          </div>
        </div>
      </PortfolioExtraSection>

    </ResumeEditorShell>
  );
}
