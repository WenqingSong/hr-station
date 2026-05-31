import React, { useState, useRef, useEffect } from "react";
import { Link, useLocation } from "react-router";
import {
  ChevronLeft,
  ChevronUp,
  ChevronDown,
  Trash2,
  Plus,
  CheckCircle2,
  Clock,
  Sparkles,
  X,
} from "lucide-react";

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
  const [savedHint, setSavedHint] = useState(false);

  // ── Resume basic ────────────────────────────────────────────────────────────
  const [resumeName, setResumeName] = useState(mode === "add" ? "V3 新简历" : "V2 后端开发简历");
  const [jobDirection, setJobDirection] = useState("后端开发");
  const [resumeDesc, setResumeDesc] = useState(
    "用于后端开发、云计算开发等技术岗位的 AI 推荐与匹配分析。"
  );
  const [setAsCurrent, setSetAsCurrent] = useState(true);

  // ── Personal info ───────────────────────────────────────────────────────────
  const [name] = useState("测试候选人 A");
  const [age] = useState("23");
  const [degree] = useState("本科");
  const [school] = useState("某大学");
  const [major] = useState("软件工程");
  const [gradYear] = useState("2026");
  const [phone] = useState("182****9164");
  const [email] = useState("candidate@example.com");

  // ── Job intention ───────────────────────────────────────────────────────────
  const [direction, setDirection] = useState("后端开发");
  const [cities, setCities] = useState("深圳 / 北京 / 上海");
  const [jobType, setJobType] = useState("校园招聘 / 实习");

  // ── Education ───────────────────────────────────────────────────────────────
  const [educations, setEducations] = useState<Education[]>([
    {
      id: 1,
      school: "某大学",
      major: "软件工程",
      degree: "本科",
      startDate: "2022.09",
      endDate: "2026.06",
      courses: "数据结构、操作系统、计算机网络、数据库、软件工程",
    },
  ]);

  // ── Internship ──────────────────────────────────────────────────────────────
  const [internships, setInternships] = useState<Internship[]>([
    {
      id: 1,
      company: "ExampleTech",
      position: "后端开发实习生",
      startDate: "2025.06",
      endDate: "2025.09",
      content:
        "参与后台接口开发和数据库表设计，协助完成用户权限、订单查询和缓存优化等功能。",
    },
  ]);

  // ── Projects ────────────────────────────────────────────────────────────────
  const [projects, setProjects] = useState<Project[]>([
    {
      id: 1,
      name: "Go 微服务订单系统",
      stack: "Go / MySQL / Redis / Docker / 微服务",
      desc: "负责订单服务核心接口设计与实现，使用 MySQL 完成订单数据建模，结合 Redis 缓存优化高频查询路径。",
      role: "接口设计、数据库设计、缓存优化、错误处理、权限校验",
    },
    {
      id: 2,
      name: "RBAC 权限管理系统",
      stack: "Go / MySQL / Casbin / JWT",
      desc: "设计并实现用户、角色、菜单与接口权限的关联校验，提升后台系统权限管理的可维护性。",
      role: "权限模型设计、接口鉴权、后台管理接口开发",
    },
    {
      id: 3,
      name: "缓存优化项目",
      stack: "Redis / MySQL / 查询优化",
      desc: "针对高频查询场景设计缓存策略，减少重复数据库访问。",
      role: "缓存策略设计、查询优化、接口性能优化",
    },
  ]);

  // ── Skills ──────────────────────────────────────────────────────────────────
  const [skills, setSkills] = useState<string[]>([
    "Go", "MySQL", "Redis", "Docker", "微服务", "RBAC", "接口开发", "缓存优化",
  ]);
  const [skillInput, setSkillInput] = useState("");

  // ── Self summary ─────────────────────────────────────────────────────────────
  const [summary, setSummary] = useState(
    "具备 Go 后端开发基础，熟悉 MySQL、Redis 和常见 Web 后端开发流程，了解微服务、权限控制和接口设计。希望从事后端开发、云计算开发等方向。"
  );

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

  return (
    <>
      {/* ── Page layout ───────────────────────────────────────────────────── */}
      <div className="max-w-[1160px] mx-auto px-4 py-8 pb-12 flex gap-6 items-start">

        {/* ── Left anchor nav ───────────────────────────────────────────── */}
        <aside className="w-[168px] flex-shrink-0 sticky top-[72px]">
          <div className="bg-white border border-gray-200 py-2">
            {NAV_SECTIONS.map((s) => (
              <button
                key={s.id}
                onClick={() => scrollTo(s.id)}
                className={`w-full text-left px-4 py-2.5 text-sm transition-colors ${
                  activeSection === s.id
                    ? "text-[#0052D9] bg-blue-50 border-l-2 border-[#0052D9]"
                    : "text-gray-600 hover:text-gray-900 hover:bg-gray-50 border-l-2 border-transparent"
                }`}
              >
                {s.label}
              </button>
            ))}
          </div>
        </aside>

        {/* ── Main form column ──────────────────────────────────────────── */}
        <div className="flex-1 min-w-0 space-y-0.5">

          {/* Page title area */}
          <div className="mb-5">
            <Link
              to="/resumes"
              className="inline-flex items-center gap-1 text-xs text-gray-400 hover:text-gray-600 transition-colors mb-3"
            >
              <ChevronLeft className="w-3 h-3" />
              返回我的简历
            </Link>
            <div className="flex items-start justify-between">
              <div>
                <div className="flex items-center gap-3">
                  <h1 className="text-xl font-semibold text-gray-900">{mode === "add" ? "添加简历" : "修改简历"}</h1>
                  <span className="px-2 py-0.5 bg-blue-50 border border-blue-100 rounded text-xs text-[#0052D9]">
                    {mode === "add" ? "新建模式" : "修改模式"}
                  </span>
                </div>
                <p className="mt-1 text-sm text-gray-500">
                  维护用于岗位推荐、岗位匹配分析和简历优化的简历信息。
                </p>
              </div>
              <button
                className="px-4 py-1.5 border border-gray-300 text-gray-600 text-sm rounded hover:bg-gray-50 transition-colors"
                onClick={() => setSavedHint(true)}
              >
                保存草稿
              </button>
            </div>
          </div>

          {/* Status bar */}
          <div className="bg-white border border-gray-200 px-5 py-3 mb-4 flex items-center gap-5 flex-wrap">
            <div className="flex items-center gap-2 text-sm">
              <span className="text-gray-400">当前简历：</span>
              <span className="font-medium text-gray-900">V2 后端开发简历</span>
            </div>
            <span className="w-px h-4 bg-gray-200" />
            <div className="flex items-center gap-2 text-sm">
              <span className="text-gray-400">模式：</span>
              <span className="text-gray-700">修改简历</span>
            </div>
            <span className="w-px h-4 bg-gray-200" />
            <div className="flex items-center gap-2 text-sm">
              <span className="text-gray-400">完整度：</span>
              <div className="w-16 h-1.5 bg-gray-100 rounded-full overflow-hidden">
                <div className="h-full bg-[#0052D9] rounded-full" style={{ width: "78%" }} />
              </div>
              <span className="font-semibold text-[#0052D9]">78%</span>
            </div>
            <span className="w-px h-4 bg-gray-200" />
            <div className="flex items-center gap-1.5 text-sm text-green-600">
              <CheckCircle2 className="w-3.5 h-3.5" />
              可用于 AI 分析
            </div>
          </div>

          {/* AI hint strip */}
          <div className="flex items-center gap-2 px-4 py-2.5 bg-blue-50 border border-blue-100 mb-4">
            <Sparkles className="w-3.5 h-3.5 text-[#0052D9] flex-shrink-0" />
            <span className="text-xs text-[#0052D9]">
              当前简历可用于：<strong>岗位推荐</strong> / <strong>岗位匹配分析</strong> / <strong>简历优化</strong>
            </span>
            <span className="w-px h-3.5 bg-blue-200 mx-1" />
            <span className="text-xs text-gray-400">AI 建议仅作投递前参考，不替代真实经历核实。</span>
          </div>

          {/* ── Module 1: 简历基础 ───────────────────────────────────────── */}
          <FormSection id="resume-basic" title="简历基础">
            <div className="space-y-5">
              <FieldRow label="简历名称" required>
                <input
                  className={inputCls}
                  value={resumeName}
                  onChange={(e) => setResumeName(e.target.value)}
                  placeholder="为这份简历起一个名字"
                />
              </FieldRow>
              <FieldRow label="求职方向" required>
                <input
                  className={inputCls}
                  value={jobDirection}
                  onChange={(e) => setJobDirection(e.target.value)}
                  placeholder="如：后端开发"
                />
              </FieldRow>
              <FieldRow label="简历说明">
                <textarea
                  className={textareaCls}
                  rows={2}
                  value={resumeDesc}
                  onChange={(e) => setResumeDesc(e.target.value)}
                  placeholder="简要描述这份简历适用的方向或用途"
                />
              </FieldRow>
              <FieldRow label="设为当前简历">
                <div className="flex items-center gap-3 pt-1">
                  <button
                    onClick={() => setSetAsCurrent(!setAsCurrent)}
                    className={`relative w-10 h-5 rounded-full transition-colors ${
                      setAsCurrent ? "bg-[#0052D9]" : "bg-gray-200"
                    }`}
                  >
                    <span
                      className={`absolute top-0.5 w-4 h-4 bg-white rounded-full shadow transition-transform ${
                        setAsCurrent ? "translate-x-5" : "translate-x-0.5"
                      }`}
                    />
                  </button>
                  <span className="text-xs text-gray-500">
                    当前简历会优先用于岗位推荐、岗位匹配分析和简历优化。
                  </span>
                </div>
              </FieldRow>
            </div>
          </FormSection>

          {/* ── Module 2: 基础信息 ───────────────────────────────────────── */}
          <FormSection id="personal-info" title="基础信息">
            <div className="space-y-5">
              <div className="grid grid-cols-2 gap-x-8 gap-y-5">
                <FieldRow label="姓名" required>
                  <input className={inputCls} defaultValue={name} placeholder="姓名" />
                </FieldRow>
                <FieldRow label="年龄">
                  <input className={inputCls} defaultValue={age} placeholder="年龄" />
                </FieldRow>
                <FieldRow label="学历" required>
                  <select className={inputCls} defaultValue={degree}>
                    <option>本科</option>
                    <option>硕士</option>
                    <option>博士</option>
                    <option>专科</option>
                  </select>
                </FieldRow>
                <FieldRow label="毕业时间">
                  <input className={inputCls} defaultValue={gradYear} placeholder="如：2026" />
                </FieldRow>
                <FieldRow label="学校" required>
                  <input className={inputCls} defaultValue={school} placeholder="学校名称" />
                </FieldRow>
                <FieldRow label="专业" required>
                  <input className={inputCls} defaultValue={major} placeholder="专业名称" />
                </FieldRow>
              </div>
              <div className="border-t border-gray-100 pt-5 grid grid-cols-2 gap-x-8 gap-y-5">
                <FieldRow label="手机号">
                  <input className={inputCls} defaultValue={phone} placeholder="手机号（脱敏展示）" />
                </FieldRow>
                <FieldRow label="邮箱">
                  <input className={inputCls} defaultValue={email} placeholder="邮箱地址" />
                </FieldRow>
              </div>
            </div>
          </FormSection>

          {/* ── Module 3: 求职方向 ───────────────────────────────────────── */}
          <FormSection id="job-intention" title="求职方向">
            <div className="space-y-5">
              <FieldRow label="求职方向" required>
                <input
                  className={inputCls}
                  value={direction}
                  onChange={(e) => setDirection(e.target.value)}
                  placeholder="如：后端开发"
                />
              </FieldRow>
              <FieldRow label="期望城市">
                <input
                  className={inputCls}
                  value={cities}
                  onChange={(e) => setCities(e.target.value)}
                  placeholder="如：深圳 / 北京 / 上海"
                />
              </FieldRow>
              <FieldRow label="岗位类型">
                <select
                  className={inputCls}
                  value={jobType}
                  onChange={(e) => setJobType(e.target.value)}
                >
                  <option>校园招聘 / 实习</option>
                  <option>校园招聘</option>
                  <option>实习</option>
                </select>
              </FieldRow>
            </div>
          </FormSection>

          {/* ── Module 4: 教育经历 ───────────────────────────────────────── */}
          <FormSection
            id="education"
            title="教育经历"
            onAdd={addEducation}
            addLabel="添加教育经历"
          >
            <div className="space-y-6">
              {educations.map((edu, idx) => (
                <div key={edu.id}>
                  {idx > 0 && <div className="h-px bg-gray-100 mb-6" />}
                  <div className="space-y-5">
                    <div className="grid grid-cols-2 gap-x-8 gap-y-5">
                      <FieldRow label="学校" required>
                        <input
                          className={inputCls}
                          value={edu.school}
                          onChange={(e) => updateEdu(edu.id, "school", e.target.value)}
                          placeholder="学校名称"
                        />
                      </FieldRow>
                      <FieldRow label="专业" required>
                        <input
                          className={inputCls}
                          value={edu.major}
                          onChange={(e) => updateEdu(edu.id, "major", e.target.value)}
                          placeholder="专业名称"
                        />
                      </FieldRow>
                      <FieldRow label="学历">
                        <select
                          className={inputCls}
                          value={edu.degree}
                          onChange={(e) => updateEdu(edu.id, "degree", e.target.value)}
                        >
                          <option>本科</option>
                          <option>硕士</option>
                          <option>博士</option>
                          <option>专科</option>
                        </select>
                      </FieldRow>
                      <FieldRow label="时间">
                        <div className="flex items-center gap-2">
                          <input
                            className={inputCls}
                            value={edu.startDate}
                            onChange={(e) => updateEdu(edu.id, "startDate", e.target.value)}
                            placeholder="2022.09"
                          />
                          <span className="text-gray-400 flex-shrink-0">—</span>
                          <input
                            className={inputCls}
                            value={edu.endDate}
                            onChange={(e) => updateEdu(edu.id, "endDate", e.target.value)}
                            placeholder="2026.06"
                          />
                        </div>
                      </FieldRow>
                    </div>
                    <FieldRow label="相关课程">
                      <input
                        className={inputCls}
                        value={edu.courses}
                        onChange={(e) => updateEdu(edu.id, "courses", e.target.value)}
                        placeholder="如：数据结构、操作系统、计算机网络"
                      />
                    </FieldRow>
                    {educations.length > 1 && (
                      <div className="flex justify-end">
                        <button
                          onClick={() => setEducations((es) => es.filter((e) => e.id !== edu.id))}
                          className="flex items-center gap-1 text-xs text-gray-400 hover:text-red-500 transition-colors"
                        >
                          <Trash2 className="w-3.5 h-3.5" />
                          删除
                        </button>
                      </div>
                    )}
                  </div>
                </div>
              ))}
            </div>
          </FormSection>

          {/* ── Module 5: 实习经历 ───────────────────────────────────────── */}
          <FormSection
            id="internship"
            title="实习经历"
            onAdd={addInternship}
            addLabel="添加实习经历"
          >
            <div className="space-y-6">
              {internships.map((intern, idx) => (
                <div key={intern.id}>
                  {idx > 0 && <div className="h-px bg-gray-100 mb-6" />}
                  <div className="space-y-5">
                    <div className="grid grid-cols-2 gap-x-8 gap-y-5">
                      <FieldRow label="公司名称" required>
                        <input
                          className={inputCls}
                          value={intern.company}
                          onChange={(e) => updateIntern(intern.id, "company", e.target.value)}
                          placeholder="公司名称"
                        />
                      </FieldRow>
                      <FieldRow label="岗位名称" required>
                        <input
                          className={inputCls}
                          value={intern.position}
                          onChange={(e) => updateIntern(intern.id, "position", e.target.value)}
                          placeholder="岗位名称"
                        />
                      </FieldRow>
                      <FieldRow label="时间">
                        <div className="flex items-center gap-2">
                          <input
                            className={inputCls}
                            value={intern.startDate}
                            onChange={(e) => updateIntern(intern.id, "startDate", e.target.value)}
                            placeholder="2025.06"
                          />
                          <span className="text-gray-400 flex-shrink-0">—</span>
                          <input
                            className={inputCls}
                            value={intern.endDate}
                            onChange={(e) => updateIntern(intern.id, "endDate", e.target.value)}
                            placeholder="2025.09"
                          />
                        </div>
                      </FieldRow>
                    </div>
                    <FieldRow label="工作内容">
                      <textarea
                        className={textareaCls}
                        rows={3}
                        value={intern.content}
                        onChange={(e) => updateIntern(intern.id, "content", e.target.value)}
                        placeholder="简要描述工作内容和成果"
                      />
                    </FieldRow>
                    {internships.length > 1 && (
                      <div className="flex justify-end">
                        <button
                          onClick={() => setInternships((is) => is.filter((i) => i.id !== intern.id))}
                          className="flex items-center gap-1 text-xs text-gray-400 hover:text-red-500 transition-colors"
                        >
                          <Trash2 className="w-3.5 h-3.5" />
                          删除
                        </button>
                      </div>
                    )}
                  </div>
                </div>
              ))}
            </div>
          </FormSection>

          {/* ── Module 6: 项目经历 ───────────────────────────────────────── */}
          <FormSection
            id="projects"
            title="项目经历"
            onAdd={addProject}
            addLabel="添加项目经历"
          >
            <div className="space-y-0">
              {projects.map((proj, idx) => (
                <div key={proj.id}>
                  {idx > 0 && <div className="h-px bg-gray-100 my-6" />}
                  <div className="space-y-4">
                    {/* Project header: name + controls */}
                    <div className="flex items-center justify-between">
                      <span className="text-xs text-gray-400">项目 {idx + 1}</span>
                      <div className="flex items-center gap-1">
                        <button
                          onClick={() => moveProject(idx, -1)}
                          disabled={idx === 0}
                          className="p-1 text-gray-300 hover:text-gray-600 disabled:opacity-30 transition-colors"
                          title="上移"
                        >
                          <ChevronUp className="w-3.5 h-3.5" />
                        </button>
                        <button
                          onClick={() => moveProject(idx, 1)}
                          disabled={idx === projects.length - 1}
                          className="p-1 text-gray-300 hover:text-gray-600 disabled:opacity-30 transition-colors"
                          title="下移"
                        >
                          <ChevronDown className="w-3.5 h-3.5" />
                        </button>
                        <button
                          onClick={() => removeProject(proj.id)}
                          className="p-1 text-gray-300 hover:text-red-400 transition-colors"
                          title="删除"
                        >
                          <Trash2 className="w-3.5 h-3.5" />
                        </button>
                      </div>
                    </div>
                    <FieldRow label="项目名称" required>
                      <input
                        className={inputCls}
                        value={proj.name}
                        onChange={(e) => updateProject(proj.id, "name", e.target.value)}
                        placeholder="项目名称"
                      />
                    </FieldRow>
                    <FieldRow label="技术栈">
                      <input
                        className={inputCls}
                        value={proj.stack}
                        onChange={(e) => updateProject(proj.id, "stack", e.target.value)}
                        placeholder="如：Go / MySQL / Redis"
                      />
                    </FieldRow>
                    <FieldRow label="项目描述">
                      <textarea
                        className={textareaCls}
                        rows={3}
                        value={proj.desc}
                        onChange={(e) => updateProject(proj.id, "desc", e.target.value)}
                        placeholder="描述项目背景、你负责的部分和主要成果"
                      />
                    </FieldRow>
                    <FieldRow label="个人职责">
                      <input
                        className={inputCls}
                        value={proj.role}
                        onChange={(e) => updateProject(proj.id, "role", e.target.value)}
                        placeholder="如：接口设计、数据库设计、缓存优化"
                      />
                    </FieldRow>
                  </div>
                </div>
              ))}
            </div>
          </FormSection>

          {/* ── Module 7: 技能关键词 ─────────────────────────────────────── */}
          <FormSection id="skills" title="技能关键词">
            <div className="space-y-4">
              {/* Tags */}
              <div className="flex flex-wrap gap-2">
                {skills.map((s) => (
                  <span
                    key={s}
                    className="flex items-center gap-1.5 pl-3 pr-2 py-1 bg-blue-50 border border-blue-100 rounded text-sm text-[#0052D9]"
                  >
                    {s}
                    <button
                      onClick={() => removeSkill(s)}
                      className="text-blue-300 hover:text-blue-600 transition-colors"
                    >
                      <X className="w-3 h-3" />
                    </button>
                  </span>
                ))}
              </div>
              {/* Input */}
              <div className="flex gap-2">
                <input
                  className={inputCls}
                  value={skillInput}
                  onChange={(e) => setSkillInput(e.target.value)}
                  onKeyDown={(e) => e.key === "Enter" && addSkill()}
                  placeholder="输入技能关键词后按回车添加"
                />
                <button
                  onClick={addSkill}
                  className="px-4 py-2 border border-gray-200 text-gray-600 text-sm rounded hover:bg-gray-50 transition-colors flex-shrink-0"
                >
                  添加
                </button>
              </div>
              <p className="text-xs text-gray-400">
                技能关键词会影响岗位推荐和 JD 匹配分析，请填写真实掌握的技能。
              </p>
            </div>
          </FormSection>

          {/* ── Module 8: 自我评价 ───────────────────────────────────────── */}
          <FormSection id="self-summary" title="自我评价">
            <textarea
              className={textareaCls}
              rows={4}
              value={summary}
              onChange={(e) => setSummary(e.target.value)}
              placeholder="简要介绍自己的技术背景、求职意向和个人优势"
            />
          </FormSection>

          {/* ── Bottom action area — static, in page flow ─────────────── */}
          <div className="mt-8 pt-5 border-t border-gray-200 flex items-center justify-between">
            <div className="text-sm text-gray-500 flex items-center gap-3">
              <span>当前模式：<span className="text-gray-700 font-medium">{mode === "add" ? "添加简历" : "修改简历"}</span></span>
              <span className="w-px h-4 bg-gray-200" />
              <span className="flex items-center gap-1.5">
                <Clock className="w-3.5 h-3.5" />
                最近保存：2 小时前
              </span>
            </div>
            <div className="flex items-center gap-3">
              <Link
                to="/resumes"
                className="px-4 py-2 text-gray-500 text-sm hover:text-gray-800 transition-colors"
              >
                取消
              </Link>
              <button
                className="px-4 py-2 border border-gray-300 text-gray-600 text-sm rounded hover:bg-gray-50 transition-colors"
                onClick={() => setSavedHint(true)}
              >
                保存草稿
              </button>
              <button
                className="px-5 py-2 bg-[#0052D9] text-white text-sm font-medium rounded hover:bg-[#003DA5] transition-colors"
                onClick={() => setSavedHint(true)}
              >
                {mode === "add" ? "保存新简历" : "保存并设为当前简历"}
              </button>
            </div>
          </div>

          {savedHint && (
            <div className="mt-4 p-3 bg-green-50 border border-green-200 rounded text-sm text-green-700 flex items-center gap-2">
              <CheckCircle2 className="w-4 h-4 text-green-600" />
              已保存本次演示修改
            </div>
          )}

        </div>
      </div>
    </>
  );
}
