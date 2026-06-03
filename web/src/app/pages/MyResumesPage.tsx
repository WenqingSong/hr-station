import React, { useState, useRef, useEffect, useCallback } from "react";
import { Link, useLocation } from "react-router";
import {
  ChevronLeft,
  ChevronUp,
  ChevronDown,
  Trash2,
  Plus,
  CheckCircle2,
  AlertCircle,
  Clock,
  Sparkles,
  X,
  Loader2,
} from "lucide-react";
import { loadResumeManagementPageViewModel } from "../../api";
import { createResume, updateResume, getCurrentResume, getResumes } from "../../api/resumes";
import type { ResumeManagementViewModel } from "../../adapters";
import type { ResumeEducation, ResumeExperience, ResumeProject } from "../../types";
import { ResumeSkeleton } from "../components/ui/ResumeSkeleton";
import { PageError, PageLoading } from "../components/PageState";
import { ContentPageShell } from "../components/layout/ContentPageShell";

// ─── Types ────────────────────────────────────────────────────────────────────

interface Project {
  id: number;
  name: string;
  stack: string;
  desc: string;
  role: string;
}

interface Education {
  id: number;
  school: string;
  major: string;
  degree: string;
  startDate: string;
  endDate: string;
  courses: string;
}

interface Internship {
  id: number;
  company: string;
  position: string;
  startDate: string;
  endDate: string;
  content: string;
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

// ─── Shared input styles ──────────────────────────────────────────────────────

const inputCls =
  "w-full px-3 py-2 border border-gray-200 rounded text-sm text-gray-900 bg-white focus:outline-none focus:ring-1 focus:ring-[#0052D9] focus:border-[#0052D9] transition-colors placeholder-gray-400";
const textareaCls = inputCls + " resize-none leading-relaxed";

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
        .catch(() => { setResumeError("简历加载失败，请稍后重试。"); setLoading(false); });
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

  // ── Job intention ───────────────────────────────────────────────────────────
  const [direction, setDirection] = useState("");
  const [cities, setCities] = useState("");
  const [jobType, setJobType] = useState("校园招聘 / 实习");

  // ── Education ───────────────────────────────────────────────────────────────
  const [educations, setEducations] = useState<Education[]>([]);

  // ── Internship ──────────────────────────────────────────────────────────────
  const [internships, setInternships] = useState<Internship[]>([]);

  // ── Projects ────────────────────────────────────────────────────────────────
  const [projects, setProjects] = useState<Project[]>([]);

  // ── Skills ──────────────────────────────────────────────────────────────────
  const [skills, setSkills] = useState<string[]>([]);
  const [skillInput, setSkillInput] = useState("");

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
      { id: newId, name: "", stack: "", desc: "", role: "" },
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

  function addEducation() {
    const newId = Math.max(0, ...educations.map((e) => e.id)) + 1;
    setEducations((es) => [
      ...es,
      { id: newId, school: "", major: "", degree: "本科", startDate: "", endDate: "", courses: "" },
    ]);
  }

  function updateEdu(id: number, field: keyof Education, value: string) {
    setEducations((es) => es.map((e) => (e.id === id ? { ...e, [field]: value } : e)));
  }

  function addInternship() {
    const newId = Math.max(0, ...internships.map((i) => i.id)) + 1;
    setInternships((is) => [
      ...is,
      { id: newId, company: "", position: "", startDate: "", endDate: "", content: "" },
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
    <ContentPageShell maxWidthClassName="max-w-[1040px]">
      {/* ── PageHeader ────────────────────────────────────────────────── */}
      <div className="pt-8 pb-4">
        <Link
          to="/resumes"
          className="inline-flex items-center gap-1 text-xs text-gray-400 hover:text-gray-600 transition-colors mb-3"
        >
          <ChevronLeft className="w-3 h-3" />
          返回我的简历
        </Link>
        <h1 className="text-2xl font-bold text-slate-950">
          {mode === "add" ? "创建简历" : "编辑简历"}
        </h1>
        <p className="mt-2 text-sm text-slate-500">
          完善你的经历和技能，让 Offerr鹅 更准确地完成岗位推荐、匹配分析和简历优化。
        </p>
      </div>

      {/* ── MainLayout ────────────────────────────────────────────────── */}
      <div>
        <div className="space-y-5">

            {/* AI hint strip */}
            <div className="flex items-center gap-2 px-4 py-2.5 rounded-xl bg-blue-50 border border-blue-100">
              <Sparkles className="w-3.5 h-3.5 text-[#0052D9] flex-shrink-0" />
              <span className="text-xs text-[#0052D9]">
                当前简历可用于：<strong>岗位推荐</strong> / <strong>岗位匹配分析</strong> / <strong>简历优化</strong>
              </span>
              <span className="w-px h-3.5 bg-blue-200 mx-1" />
              <span className="text-xs text-gray-400">AI 建议仅作投递前参考，不替代真实经历核实。</span>
            </div>

          {/* ── Module 1: 简历基础 ───────────────────────────────────────── */}
          <div id="resume-basic" className="rounded-2xl bg-white ring-1 ring-gray-100 p-6">
            <h2 className="text-[16px] font-semibold text-slate-900 mb-1">简历基础</h2>
            <p className="text-[13px] text-slate-500 mb-5">
              先确定这份简历的求职定位，后续推荐、匹配和优化都会基于这里的信息。
            </p>
            <div className="space-y-5">
              {/* 简历名称 */}
              <div>
                <label className="mb-1.5 block text-sm font-medium text-slate-700">
                  简历名称 <span className="text-red-500">*</span>
                </label>
                <input
                  className="h-11 w-full rounded-xl border border-slate-200 bg-white px-3 text-sm text-slate-900 placeholder-slate-400 focus:border-[#0052D9] focus:outline-none focus:ring-2 focus:ring-[#0052D9]/25"
                  value={resumeName}
                  onChange={(e) => setResumeName(e.target.value)}
                  placeholder="为这份简历起一个名字"
                />
              </div>
              {/* 求职方向 */}
              <div>
                <label className="mb-1.5 block text-sm font-medium text-slate-700">
                  求职方向 <span className="text-red-500">*</span>
                </label>
                <input
                  className="h-11 w-full rounded-xl border border-slate-200 bg-white px-3 text-sm text-slate-900 placeholder-slate-400 focus:border-[#0052D9] focus:outline-none focus:ring-2 focus:ring-[#0052D9]/25"
                  value={jobDirection}
                  onChange={(e) => setJobDirection(e.target.value)}
                  placeholder="如：后端开发"
                />
              </div>
              {/* 简历说明 */}
              <div>
                <label className="mb-1.5 block text-sm font-medium text-slate-700">简历说明</label>
                <textarea
                  className="min-h-[80px] w-full rounded-xl border border-slate-200 bg-white px-3 py-2 text-sm text-slate-900 placeholder-slate-400 resize-none leading-relaxed focus:border-[#0052D9] focus:outline-none focus:ring-2 focus:ring-[#0052D9]/25"
                  rows={2}
                  value={resumeDesc}
                  onChange={(e) => setResumeDesc(e.target.value)}
                  placeholder="简要描述这份简历适用的方向或用途"
                />
              </div>
              {/* 保存说明 */}
              <div>
                <div className="flex items-start gap-2 rounded-xl bg-blue-50/60 border border-blue-100/40 px-4 py-3">
                  <Sparkles className="w-4 h-4 text-[#0052D9] mt-0.5" />
                  <p className="text-[13px] text-slate-600 leading-relaxed">
                    保存后，最新版本将自动用于岗位推荐、岗位匹配分析和简历优化建议。
                  </p>
                </div>
              </div>
            </div>
          </div>

          {/* ── Module 2: 基础信息 ───────────────────────────────────────── */}
          <div id="personal-info" className="rounded-2xl bg-white ring-1 ring-gray-100 p-6">
            <h2 className="text-[16px] font-semibold text-slate-900 mb-1">基础信息</h2>
            <p className="text-[13px] text-slate-500 mb-5">
              用于生成简历摘要和岗位匹配时的基础背景判断。
            </p>
            <div className="grid grid-cols-1 gap-x-6 gap-y-5 sm:grid-cols-2">
              <div>
                <label className="mb-1.5 block text-sm font-medium text-slate-700">
                  姓名 <span className="text-red-500">*</span>
                </label>
                <input className="h-11 w-full rounded-xl border border-slate-200 bg-white px-3 text-sm focus:border-[#0052D9] focus:outline-none focus:ring-2 focus:ring-[#0052D9]/25" value={personalName} onChange={(e) => setPersonalName(e.target.value)} placeholder="姓名" />
              </div>
              <div>
                <label className="mb-1.5 block text-sm font-medium text-slate-700">年龄</label>
                <input className="h-11 w-full rounded-xl border border-slate-200 bg-white px-3 text-sm focus:border-[#0052D9] focus:outline-none focus:ring-2 focus:ring-[#0052D9]/25" value={age} onChange={(e) => setAge(e.target.value)} placeholder="年龄" />
              </div>
              <div>
                <label className="mb-1.5 block text-sm font-medium text-slate-700">
                  学历 <span className="text-red-500">*</span>
                </label>
                <select className="h-11 w-full rounded-xl border border-slate-200 bg-white px-3 text-sm focus:border-[#0052D9] focus:outline-none focus:ring-2 focus:ring-[#0052D9]/25" value={degree} onChange={(e) => setDegree(e.target.value)}>
                  <option>本科</option>
                  <option>硕士</option>
                  <option>博士</option>
                  <option>专科</option>
                </select>
              </div>
              <div>
                <label className="mb-1.5 block text-sm font-medium text-slate-700">毕业时间</label>
                <input className="h-11 w-full rounded-xl border border-slate-200 bg-white px-3 text-sm focus:border-[#0052D9] focus:outline-none focus:ring-2 focus:ring-[#0052D9]/25" value={gradYear} onChange={(e) => setGradYear(e.target.value)} placeholder="如：2026" />
              </div>
              <div>
                <label className="mb-1.5 block text-sm font-medium text-slate-700">
                  学校 <span className="text-red-500">*</span>
                </label>
                <input className="h-11 w-full rounded-xl border border-slate-200 bg-white px-3 text-sm focus:border-[#0052D9] focus:outline-none focus:ring-2 focus:ring-[#0052D9]/25" value={school} onChange={(e) => setSchool(e.target.value)} placeholder="学校名称" />
              </div>
              <div>
                <label className="mb-1.5 block text-sm font-medium text-slate-700">
                  专业 <span className="text-red-500">*</span>
                </label>
                <input className="h-11 w-full rounded-xl border border-slate-200 bg-white px-3 text-sm focus:border-[#0052D9] focus:outline-none focus:ring-2 focus:ring-[#0052D9]/25" value={major} onChange={(e) => setMajor(e.target.value)} placeholder="专业名称" />
              </div>
            </div>
            <div className="mt-5 pt-5 border-t border-gray-100 grid grid-cols-1 gap-x-6 gap-y-5 sm:grid-cols-2">
              <div>
                <label className="mb-1.5 block text-sm font-medium text-slate-700">手机号</label>
                <input className="h-11 w-full rounded-xl border border-slate-200 bg-white px-3 text-sm focus:border-[#0052D9] focus:outline-none focus:ring-2 focus:ring-[#0052D9]/25" value={phone} onChange={(e) => setPhone(e.target.value)} placeholder="手机号（脱敏展示）" />
              </div>
              <div>
                <label className="mb-1.5 block text-sm font-medium text-slate-700">邮箱</label>
                <input className="h-11 w-full rounded-xl border border-slate-200 bg-white px-3 text-sm focus:border-[#0052D9] focus:outline-none focus:ring-2 focus:ring-[#0052D9]/25" value={email} onChange={(e) => setEmail(e.target.value)} placeholder="邮箱地址" />
              </div>
            </div>
          </div>

          {/* ── Module 3: 求职方向 ───────────────────────────────────────── */}
          <div id="job-intention" className="rounded-2xl bg-white ring-1 ring-blue-100 p-6">
            <h2 className="text-[16px] font-semibold text-slate-900 mb-1">求职方向</h2>
            <p className="text-[13px] text-slate-500 mb-5">
              Offerr鹅会优先根据求职方向、城市和岗位类型筛选更适合你的岗位。
            </p>
            <div className="grid grid-cols-1 gap-x-6 gap-y-5 sm:grid-cols-2">
              <div>
                <label className="mb-1.5 block text-sm font-medium text-slate-700">
                  求职方向 <span className="text-red-500">*</span>
                </label>
                <input
                  className="h-11 w-full rounded-xl border border-slate-200 bg-white px-3 text-sm focus:border-[#0052D9] focus:outline-none focus:ring-2 focus:ring-[#0052D9]/25"
                  value={direction}
                  onChange={(e) => setDirection(e.target.value)}
                  placeholder="如：后端开发"
                />
              </div>
              <div>
                <label className="mb-1.5 block text-sm font-medium text-slate-700">期望城市</label>
                <input
                  className="h-11 w-full rounded-xl border border-slate-200 bg-white px-3 text-sm focus:border-[#0052D9] focus:outline-none focus:ring-2 focus:ring-[#0052D9]/25"
                  value={cities}
                  onChange={(e) => setCities(e.target.value)}
                  placeholder="如：深圳 / 北京 / 上海"
                />
              </div>
            </div>
            <div className="mt-5">
              <label className="mb-1.5 block text-sm font-medium text-slate-700">岗位类型</label>
              <select
                className="h-11 w-full sm:max-w-xs rounded-xl border border-slate-200 bg-white px-3 text-sm focus:border-[#0052D9] focus:outline-none focus:ring-2 focus:ring-[#0052D9]/25"
                value={jobType}
                onChange={(e) => setJobType(e.target.value)}
              >
                <option>校园招聘 / 实习</option>
                <option>校园招聘</option>
                <option>实习</option>
              </select>
            </div>
            <p className="mt-4 text-[12px] text-slate-400">
              这些信息会影响岗位推荐、岗位匹配分析和简历优化建议的准确度。
            </p>
          </div>

          {/* ── Module 4: 教育经历 ───────────────────────────────────────── */}
          <div id="education" className="rounded-2xl bg-white ring-1 ring-gray-100 p-6">
            <div className="flex items-center justify-between mb-1">
              <h2 className="text-[16px] font-semibold text-slate-900">教育经历</h2>
              <button onClick={addEducation} className="flex items-center gap-1 text-[13px] font-medium text-[#0052D9] hover:text-[#0047BD] transition-colors">
                <Plus className="w-3.5 h-3.5" />添加教育经历
              </button>
            </div>
            <p className="text-[13px] text-slate-500 mb-5">
              补充教育背景，帮助 Offerr鹅 判断你的学历、专业和岗位匹配基础。
            </p>

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
                    <div className="grid grid-cols-1 gap-x-5 gap-y-4 sm:grid-cols-2">
                      <div>
                        <label className="mb-1.5 block text-sm font-medium text-slate-700">学校 <span className="text-red-500">*</span></label>
                        <input className="h-10 w-full rounded-xl border border-slate-200 bg-white px-3 text-sm focus:border-[#0052D9] focus:outline-none focus:ring-2 focus:ring-[#0052D9]/25" value={edu.school} onChange={(e) => updateEdu(edu.id, "school", e.target.value)} placeholder="学校名称" />
                      </div>
                      <div>
                        <label className="mb-1.5 block text-sm font-medium text-slate-700">专业 <span className="text-red-500">*</span></label>
                        <input className="h-10 w-full rounded-xl border border-slate-200 bg-white px-3 text-sm focus:border-[#0052D9] focus:outline-none focus:ring-2 focus:ring-[#0052D9]/25" value={edu.major} onChange={(e) => updateEdu(edu.id, "major", e.target.value)} placeholder="专业名称" />
                      </div>
                      <div>
                        <label className="mb-1.5 block text-sm font-medium text-slate-700">学历</label>
                        <select className="h-10 w-full rounded-xl border border-slate-200 bg-white px-3 text-sm focus:border-[#0052D9] focus:outline-none focus:ring-2 focus:ring-[#0052D9]/25" value={edu.degree} onChange={(e) => updateEdu(edu.id, "degree", e.target.value)}>
                          <option>本科</option><option>硕士</option><option>博士</option><option>专科</option>
                        </select>
                      </div>
                      <div>
                        <label className="mb-1.5 block text-sm font-medium text-slate-700">时间</label>
                        <div className="flex items-center gap-2">
                          <input className="h-10 flex-1 rounded-xl border border-slate-200 bg-white px-3 text-sm focus:border-[#0052D9] focus:outline-none focus:ring-2 focus:ring-[#0052D9]/25" value={edu.startDate} onChange={(e) => updateEdu(edu.id, "startDate", e.target.value)} placeholder="2022.09" />
                          <span className="text-slate-400">—</span>
                          <input className="h-10 flex-1 rounded-xl border border-slate-200 bg-white px-3 text-sm focus:border-[#0052D9] focus:outline-none focus:ring-2 focus:ring-[#0052D9]/25" value={edu.endDate} onChange={(e) => updateEdu(edu.id, "endDate", e.target.value)} placeholder="2026.06" />
                        </div>
                      </div>
                    </div>
                    <div>
                      <label className="mb-1.5 block text-sm font-medium text-slate-700">相关课程</label>
                      <input className="h-10 w-full rounded-xl border border-slate-200 bg-white px-3 text-sm focus:border-[#0052D9] focus:outline-none focus:ring-2 focus:ring-[#0052D9]/25" value={edu.courses} onChange={(e) => updateEdu(edu.id, "courses", e.target.value)} placeholder="如：数据结构、操作系统、计算机网络" />
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>

          {/* ── Module 5: 实习经历 ───────────────────────────────────────── */}
          <div id="internship" className="rounded-2xl bg-white ring-1 ring-gray-100 p-6">
            <div className="flex items-center justify-between mb-1">
              <h2 className="text-[16px] font-semibold text-slate-900">实习经历</h2>
              <button onClick={addInternship} className="flex items-center gap-1 text-[13px] font-medium text-[#0052D9] hover:text-[#0047BD] transition-colors">
                <Plus className="w-3.5 h-3.5" />添加实习经历
              </button>
            </div>
            <p className="text-[13px] text-slate-500 mb-5">
              实习经历会显著影响岗位匹配和简历优化建议，建议写清楚职责、技术和结果。
            </p>

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
                        <input className="h-10 w-full rounded-xl border border-slate-200 bg-white px-3 text-sm focus:border-[#0052D9] focus:outline-none focus:ring-2 focus:ring-[#0052D9]/25" value={intern.company} onChange={(e) => updateIntern(intern.id, "company", e.target.value)} placeholder="公司名称" />
                      </div>
                      <div>
                        <label className="mb-1.5 block text-sm font-medium text-slate-700">岗位名称 <span className="text-red-500">*</span></label>
                        <input className="h-10 w-full rounded-xl border border-slate-200 bg-white px-3 text-sm focus:border-[#0052D9] focus:outline-none focus:ring-2 focus:ring-[#0052D9]/25" value={intern.position} onChange={(e) => updateIntern(intern.id, "position", e.target.value)} placeholder="岗位名称" />
                      </div>
                    </div>
                    <div>
                      <label className="mb-1.5 block text-sm font-medium text-slate-700">时间</label>
                      <div className="flex items-center gap-2">
                        <input className="h-10 flex-1 rounded-xl border border-slate-200 bg-white px-3 text-sm focus:border-[#0052D9] focus:outline-none focus:ring-2 focus:ring-[#0052D9]/25" value={intern.startDate} onChange={(e) => updateIntern(intern.id, "startDate", e.target.value)} placeholder="2025.06" />
                        <span className="text-slate-400">—</span>
                        <input className="h-10 flex-1 rounded-xl border border-slate-200 bg-white px-3 text-sm focus:border-[#0052D9] focus:outline-none focus:ring-2 focus:ring-[#0052D9]/25" value={intern.endDate} onChange={(e) => updateIntern(intern.id, "endDate", e.target.value)} placeholder="2025.09" />
                      </div>
                    </div>
                    <div>
                      <label className="mb-1.5 block text-sm font-medium text-slate-700">工作内容</label>
                      <textarea className="min-h-[96px] w-full rounded-xl border border-slate-200 bg-white px-3 py-2 text-sm resize-none leading-relaxed focus:border-[#0052D9] focus:outline-none focus:ring-2 focus:ring-[#0052D9]/25" rows={3} value={intern.content} onChange={(e) => updateIntern(intern.id, "content", e.target.value)} placeholder="简要描述工作内容和成果" />
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>

          {/* ── Module 6: 项目经历 ───────────────────────────────────────── */}
          <div id="projects" className="rounded-2xl bg-white ring-1 ring-gray-100 p-6">
            <div className="flex items-center justify-between mb-1">
              <h2 className="text-[16px] font-semibold text-slate-900">项目经历</h2>
              <button onClick={addProject} className="flex items-center gap-1 text-[13px] font-medium text-[#0052D9] hover:text-[#0047BD] transition-colors">
                <Plus className="w-3.5 h-3.5" />添加项目经历
              </button>
            </div>
            <p className="text-[13px] text-slate-500 mb-5">
              项目经历是技术岗位匹配的重要依据，建议写清楚技术栈、职责和量化结果。
            </p>

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
                      <input className="h-10 w-full rounded-xl border border-slate-200 bg-white px-3 text-sm focus:border-[#0052D9] focus:outline-none focus:ring-2 focus:ring-[#0052D9]/25" value={proj.name} onChange={(e) => updateProject(proj.id, "name", e.target.value)} placeholder="项目名称" />
                    </div>
                    <div>
                      <label className="mb-1.5 block text-sm font-medium text-slate-700">技术栈</label>
                      <input className="h-10 w-full rounded-xl border border-slate-200 bg-white px-3 text-sm focus:border-[#0052D9] focus:outline-none focus:ring-2 focus:ring-[#0052D9]/25" value={proj.stack} onChange={(e) => updateProject(proj.id, "stack", e.target.value)} placeholder="如：Go / MySQL / Redis" />
                    </div>
                    <div>
                      <label className="mb-1.5 block text-sm font-medium text-slate-700">项目描述</label>
                      <textarea className="min-h-[88px] w-full rounded-xl border border-slate-200 bg-white px-3 py-2 text-sm resize-none leading-relaxed focus:border-[#0052D9] focus:outline-none focus:ring-2 focus:ring-[#0052D9]/25" rows={3} value={proj.desc} onChange={(e) => updateProject(proj.id, "desc", e.target.value)} placeholder="描述项目背景、你负责的部分和主要成果" />
                    </div>
                    <div>
                      <label className="mb-1.5 block text-sm font-medium text-slate-700">个人职责</label>
                      <input className="h-10 w-full rounded-xl border border-slate-200 bg-white px-3 text-sm focus:border-[#0052D9] focus:outline-none focus:ring-2 focus:ring-[#0052D9]/25" value={proj.role} onChange={(e) => updateProject(proj.id, "role", e.target.value)} placeholder="如：接口设计、数据库设计、缓存优化" />
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>

          {/* ── Module 7: 技能关键词 ─────────────────────────────────────── */}
          <div id="skills" className="rounded-2xl bg-white ring-1 ring-gray-100 p-6">
            <h2 className="text-[16px] font-semibold text-slate-900 mb-1">技能关键词</h2>
            <p className="text-[13px] text-slate-500 mb-5">
              补充你掌握的技术、工具和能力标签，Offerr鹅 会基于这些关键词进行岗位推荐、匹配分析和简历优化建议。
            </p>

            {skills.length === 0 ? (
              <div className="rounded-xl bg-gray-50/70 py-6 text-center mb-4">
                <p className="text-sm text-slate-500 mb-1">还没有技能关键词</p>
                <p className="text-xs text-slate-400">例如：Go、MySQL、Redis、Docker、云计算、后端开发</p>
              </div>
            ) : (
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

            <div className="flex gap-2">
              <input
                className="h-10 flex-1 rounded-xl border border-slate-200 bg-white px-3 text-sm focus:border-[#0052D9] focus:outline-none focus:ring-2 focus:ring-[#0052D9]/25"
                value={skillInput}
                onChange={(e) => setSkillInput(e.target.value)}
                onKeyDown={(e) => e.key === "Enter" && addSkill()}
                placeholder="输入技能关键词后按回车添加"
              />
              <button onClick={addSkill} className="rounded-xl bg-[#0052D9] px-5 text-[13px] font-medium text-white hover:bg-[#0047BD] transition-colors flex-shrink-0">
                添加
              </button>
            </div>
            <p className="mt-3 text-[12px] text-slate-400">
              技能关键词会影响岗位推荐和 JD 匹配分析，请填写真实掌握的技能。
            </p>
          </div>

          {/* ── Module 8: 自我评价 ───────────────────────────────────────── */}
          <div id="self-summary" className="rounded-2xl bg-white ring-1 ring-gray-100 p-6">
            <h2 className="text-[16px] font-semibold text-slate-900 mb-1">自我评价</h2>
            <p className="text-[13px] text-slate-500 mb-5">
              用一小段话概括你的优势、学习能力和求职方向。建议控制在 100-300 字。
            </p>
            <textarea
              className="min-h-[140px] w-full rounded-xl border border-slate-200 bg-white px-3 py-2 text-sm leading-6 resize-y focus:border-[#0052D9] focus:outline-none focus:ring-2 focus:ring-[#0052D9]/25"
              value={summary}
              onChange={(e) => setSummary(e.target.value)}
              placeholder="简要介绍自己的技术背景、求职意向和个人优势"
            />
            <div className="mt-2 flex items-center justify-between">
              <span className="text-[12px] text-slate-400">建议包含：技术方向、项目经验、学习能力、目标岗位。</span>
              <span className="text-[12px] text-slate-400">当前约 {summary.length} 字</span>
            </div>
          </div>

          {/* ── Save module: 页面最后一个卡片模块 ──────────────────────── */}
          <section className="rounded-2xl bg-white ring-1 ring-gray-100 p-6">
            <div className="flex flex-col gap-4 md:flex-row md:items-start md:justify-between">
              <div className="flex-1">
                <h2 className="text-[16px] font-semibold text-slate-900 mb-1">
                  {mode === "add" ? "创建简历" : "保存简历"}
                </h2>
                <p className="text-[13px] text-slate-500 leading-relaxed">
                  保存后，Offerr鹅 会基于最新保存版本进行岗位推荐、匹配分析和优化建议。
                </p>
              </div>
              <div className="flex items-center gap-3 flex-shrink-0">
                <Link
                  to="/resumes"
                  className="px-4 py-2.5 text-slate-500 text-sm hover:text-slate-800 transition-colors"
                >
                  取消
                </Link>
                <button
                  className="rounded-xl bg-[#0052D9] px-6 py-2.5 text-sm font-medium text-white hover:bg-[#0047BD] transition-colors disabled:opacity-50 shadow-sm flex items-center gap-1.5"
                  onClick={handleSave}
                  disabled={saving}
                >
                  {saving && <Loader2 className="w-4 h-4 animate-spin" />}
                  {saving ? "保存中…" : mode === "add" ? "创建简历" : "保存简历"}
                </button>
              </div>
            </div>

            {saveSuccessMsg && (
              <div className="mt-4 rounded-xl bg-green-50 ring-1 ring-green-200 p-3 text-sm text-green-700 flex items-center gap-2">
                <CheckCircle2 className="w-4 h-3.5 flex-shrink-0" />
                {saveSuccessMsg}
              </div>
            )}

            {saveErrorMsg && (
              <div className="mt-4 rounded-xl bg-red-50 ring-1 ring-red-200 p-3 text-sm text-red-600 flex items-center gap-2">
                <AlertCircle className="w-4 h-4 flex-shrink-0" />
                {saveErrorMsg}
              </div>
            )}
          </section>

        </div>{/* / main form space-y-5 */}
      </div>{/* / main content container */}

    </ContentPageShell>
  );
}
