import React, { useState, useRef, useEffect, useCallback } from "react";
import { Link, useLocation, useNavigate, useSearchParams } from "react-router";
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
import { createResume, updateResume, getCurrentResume, getResumes, getResumeDetail } from "../../api/resumes";
import type { ResumeManagementViewModel } from "../../adapters";
import { toResumeManagementViewModel } from "../../adapters";
import type {
  ResumeEducation, ResumeExperience, ResumeProject,
  ResumeProfileV2, ResumeEducationV2,
  ResumeInternshipV2, ResumeProjectV2, ResumeSkillsV2,
  ResumeCertificateV2, ResumeAwardV2, ResumePortfolioV2, ResumeExtraInfoV2,
} from "../../types";
import { ResumeSkeleton } from "../components/ui/ResumeSkeleton";
import { YearMonthPicker } from "../components/resume-editor/YearMonthPicker";
import { PageError, PageLoading } from "../components/PageState";
import { isValidVersionNo } from "../../utils/redirect";
import { getAuthRedirectIf401 } from "../utils/authRedirect";
import { isApiClientError } from "../../api/errors";
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
  ResumeEditorSaveCard,
  ResumeBasicsSection,
  BasicInfoSection,
  EducationSection,
  ProjectSection,
  SkillsSection,
  InternshipSection,
  CertificatesSection,
  PortfolioExtraSection,
} from "../components/resume-editor";

// ─── Types ────────────────────────────────────────────────────────────────────

interface Project {
  id: number;
  name: string;
  desc: string;
  role: string;
  startDate: string;
  endDate: string;
}

/** V1 + V2 draft fields for education entries (K8-5-22-F) */
interface Education {
  id: number;
  school: string;
  major: string;
  degree: string;
  startDate: string;
  endDate: string;
  // V2 draft fields (K8-5-22-F)
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
}

/** V2 skills draft (K8-5-25-D: Tencent-style 2 fields; K8-5-29-B: languageExam moved to independent list) */
interface SkillsDraft {
  developmentLanguages: string;
  aiApplication: string;
}

/** K8-5-29-B: language exam dynamic list */
interface LanguageExamDraft {
  name: string;
  score: string;
}

/** V2 certificate entry draft (K8-5-44-C: simplified) */
interface CertificateDraft {
  id: number;
  name: string;
  description: string;
}

/** V2 award entry draft (K8-5-44-C: simplified) */
interface AwardDraft {
  id: number;
  name: string;
  description: string;
}

/** V2 portfolio links draft (K8-5-25-E: Tencent-style works+homepage) */
interface WorkDraft {
  id: number;
  name: string;
  link: string;
  password: string;
}
interface PortfolioDraft {
  works: WorkDraft[];
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
  { id: "education", label: "教育经历" },
  { id: "internship", label: "实习经历" },
  { id: "projects", label: "项目经历" },
  { id: "skills", label: "技能关键词" },
  { id: "extra-info", label: "补充信息" },
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
  const [searchParams] = useSearchParams();
  const versionNoParam = searchParams.get("versionNo");
  // K8-5-46-B：非法 versionNo（≤0 或非数字）自动跳转 /resumes
  const hasInvalidVersionNo = versionNoParam !== null && !isValidVersionNo(versionNoParam);
  const editingVersionNo = versionNoParam && /^\d+$/.test(versionNoParam) ? Number(versionNoParam) : null;
  const mode: "edit" | "add" = location.pathname.includes("/new") ? "add" : "edit";

  // 非法 versionNo 直接跳转，不发起 API 请求
  useEffect(() => {
    if (hasInvalidVersionNo) {
      navigate("/resumes", { replace: true });
    }
  }, [hasInvalidVersionNo, navigate]);
  const [saving, setSaving] = useState(false);
  const [saveSuccessMsg, setSaveSuccessMsg] = useState("");
  const [saveErrorMsg, setSaveErrorMsg] = useState("");
  const [formVm, setFormVm] = useState<ResumeManagementViewModel | null>(null);
  const [loading, setLoading] = useState(mode === "edit");
  const [resumeError, setResumeError] = useState<string | null>(null);
  const [retryCount, setRetryCount] = useState(0);

  // 编辑模式：根据 query 参数决定加载指定版本或 current
  useEffect(() => {
    if (mode === "edit") {
      setLoading(true);
      setResumeError(null);

      // K8-5-40-B：如果 URL 传了 versionNo，加载指定版本详情
      if (editingVersionNo !== null && editingVersionNo > 0) {
        getResumeDetail(editingVersionNo)
          .then((resume) => {
            // K8-5-40-F：adapter 转换与 API 请求分层
            let vm: ReturnType<typeof toResumeManagementViewModel>;
            try {
              vm = toResumeManagementViewModel(resume);
            } catch (adapterError) {
              setLoading(false);
              console.error("[resume-edit] adapter failed for resume version", {
                versionNo: editingVersionNo,
                error: adapterError,
                resumeKeys: resume ? Object.keys(resume) : null,
              });
              setResumeError("简历数据解析失败，请稍后重试或检查该版本数据结构是否完整。");
              return;
            }
            setFormVm(vm);
            setLoading(false);
          })
          .catch((err: unknown) => {
            // K8-5-40-D：API 请求失败分类处理
            setLoading(false);

            // 401 → token 过期/未登录 → 跳转统一登录，保留当前 URL 作为 redirect
            if (isApiClientError(err) && err.status === 401) {
              console.error("[resume-edit] unauthorized while loading version", { versionNo: editingVersionNo, error: err });
              const loginRedirect = `/?auth=login&redirect=${encodeURIComponent(`/resumes/edit?versionNo=${editingVersionNo}`)}`;
              navigate(loginRedirect, { replace: true });
              return;
            }

            // 404 → 版本不存在
            if (isApiClientError(err) && err.status === 404) {
              console.error("[resume-edit] resume version not found", { versionNo: editingVersionNo, error: err });
              setResumeError("该简历版本不存在或已被删除。");
              return;
            }

            // 其他 API 错误 → 通用提示
            console.error("[resume-edit] failed to load resume version", { versionNo: editingVersionNo, error: err });
            setResumeError("指定简历版本加载失败，请稍后重试。");
          });
        return;
      }

      // 无 versionNo → 保持原逻辑：加载 current
      loadResumeManagementPageViewModel()
        .then((vm) => { setFormVm(vm); setLoading(false); })
        .catch((err) => {
          setLoading(false);
          const authRedirect = getAuthRedirectIf401(err, location.pathname, location.search);
          if (authRedirect) { navigate(authRedirect, { replace: true }); return; }
          console.error("[resume-edit] failed to load current resume", { error: err });
          setResumeError("简历加载失败，请稍后重试。");
        });
    }
  }, [mode, retryCount, editingVersionNo]);

  const handleRetryLoad = useCallback(() => setRetryCount((c) => c + 1), []);

  // ── Resume basic ────────────────────────────────────────────────────────────
  const [resumeName, setResumeName] = useState(mode === "add" ? "" : "");
  // K8-5-24-B: jobDirection / resumeDesc removed — target direction belongs in JobIntention, optional.

  // 编辑模式：从 VM 回填基础字段
  useEffect(() => {
    if (formVm && mode === "edit") {
      setResumeName(formVm.summary.name || "");
    }
  }, [formVm, mode]);

  // ── Personal info ───────────────────────────────────────────────────────────
  const [personalName, setPersonalName] = useState("");
  const [age, setAge] = useState("");
  const [phone, setPhone] = useState("");
  const [email, setEmail] = useState("");
  // current city (K8-5-22-F)
  const [currentCity, setCurrentCity] = useState("");

  // ── Job intention (K8-5-25-C: removed from editor) ──────────────────────────

  // ── Education ───────────────────────────────────────────────────────────────
  const [educations, setEducations] = useState<Education[]>([]);

  // ── Internship ──────────────────────────────────────────────────────────────
  const [internships, setInternships] = useState<Internship[]>([]);

  // ── Projects ────────────────────────────────────────────────────────────────
  const [projects, setProjects] = useState<Project[]>([]);

  // ── Skills ──────────────────────────────────────────────────────────────────
  const [skills, setSkills] = useState<string[]>([]);
  const [skillInput, setSkillInput] = useState("");
  // V2 draft (K8-5-25-D: Tencent-style 2 fields; K8-5-29-B: languageExam removed)
  const [skillsDraft, setSkillsDraft] = useState<SkillsDraft>({
    developmentLanguages: "",
    aiApplication: "",
  });

  // K8-5-29-D-R: track recently added item to animate only the new one, not backfilled items
  type RecentKey = "education" | "project" | "internship" | "certificate" | "award" | "languageExam" | "work";
  const [recentlyAdded, setRecentlyAdded] = useState<{ type: RecentKey; key: string } | null>(null);

  // K8-5-29-B: language exam dynamic list
  const [languageExamDrafts, setLanguageExamDrafts] = useState<LanguageExamDraft[]>([]);

  // V2 drafts (K8-5-22-H)
  const [certificateDrafts, setCertificateDrafts] = useState<CertificateDraft[]>([]);
  const [awardDrafts, setAwardDrafts] = useState<AwardDraft[]>([]);
  const [portfolioDraft, setPortfolioDraft] = useState<PortfolioDraft>({
    works: [{ id: 1, name: "", link: "", password: "" }],
    homepage: "",
  });
  const [extraInfoDraft, setExtraInfoDraft] = useState<ExtraInfoDraft>({ additionalInfo: "" });

  // ── Self summary ─────────────────────────────────────────────────────────────
  const [summary, setSummary] = useState("");

  // K8-5-23-B: Reverse-conversion helpers from arrays to draft strings.
  function joinList(value?: string[] | string | null): string {
    if (!value) return "";
    if (Array.isArray(value)) return value.filter((s) => s.length > 0).join("、");
    return value;
  }
  function joinLines(value?: string[] | string | null): string {
    if (!value) return "";
    if (Array.isArray(value)) return value.filter((s) => s.length > 0).join("\n");
    return value;
  }

  // 编辑模式：formVm → 表单 state 完整回填（V1 + V2）
  useEffect(() => {
    if (!formVm || mode !== "edit") return;

    // ── V1 baseline backfill ─────────────────────────────────────────
    // Personal info (K8-5-25-C: school/major removed from basic info)
    setPersonalName(formVm.displayName !== "暂未填写" ? formVm.displayName : "");
    setPhone(formVm.displayPhone || "");
    setEmail(formVm.displayEmail || "");

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

    // ── K8-5-23-B: V2 backfill — only for V2 resumes ─────────────────
    const v2 = formVm.v2;
    if (!v2) return;

    // Resume basics — name only (K8-5-24-B)
    const name = formVm.summary.name || "";
    if (name) setResumeName(name);

    // Profile V2 (K8-5-25-C: school/degree/major/gradYear removed from basic info)
    if (v2.profile) {
      const pr = v2.profile;
      if (pr.current_city) setCurrentCity(pr.current_city);
      if (pr.name) setPersonalName(pr.name);
    }

    // K8-5-25-C: Job intention removed from editor — no backfill needed

    // Education V2
    if (v2.educations && v2.educations.length > 0) {
      const v2EduList: Education[] = v2.educations.map((e: ResumeEducationV2, i: number) => ({
        id: i + 1,
        school: e.school || "",
        major: e.major || "",
        degree: e.degree || "本科",
        startDate: e.start_date || "",
        endDate: e.end_date || "",
        courses: (e.core_courses || []).join("、"),
        department: e.department || "",
        gpa: e.gpa || "",
        gpaScale: e.gpa_scale || "",
        rank: e.rank || "",
        coreCourses: joinList(e.core_courses),
        researchDirection: e.research_direction || "",
        lab: e.lab || "",
        paperSummary: e.paper_summary || "",
      }));
      setEducations(v2EduList);
    }

    // Internship V2
    if (v2.internships && v2.internships.length > 0) {
      const v2IntList: Internship[] = v2.internships.map((it: any, i: number) => ({
        id: i + 1,
        company: it.company || "",
        position: it.position || it.role || "",
        startDate: it.start_date || "",
        endDate: it.end_date || "",
        content: it.description || it.work_content || "",
      }));
      setInternships(v2IntList);
    }

    // Project V2
    if (v2.projects && v2.projects.length > 0) {
      const v2ProjList: Project[] = v2.projects.map((pj: ResumeProjectV2, i: number) => ({
        id: i + 1,
        name: pj.name || "",
        desc: pj.description || "",
        role: pj.role || "",
        startDate: pj.start_date || "",
        endDate: pj.end_date || "",
      }));
      setProjects(v2ProjList);
    }

    // Skills V2 (K8-5-25-D: Tencent-style 3 fields)
    if (v2.skills) {
      const sk = v2.skills;
      // New structure
      const newDev = joinList(sk.development_languages);
      const newAI = sk.ai_application || "";
      // Old structure fallback
      const oldDev = joinList(sk.programming_languages) || joinList(sk.other_keywords);
      const oldAI = sk.ai_application
        || joinList(sk.ai_application_scenarios)
        || joinList(sk.ai_tools);

      // K8-5-29-B: language exam moved to independent list; old skills.language_exam fallback
      const oldLangExam = (sk.language_exam as any);
      if (oldLangExam && (oldLangExam.type || oldLangExam.score)) {
        setLanguageExamDrafts([{ name: oldLangExam.type || "", score: oldLangExam.score || "" }]);
      }
      setSkillsDraft({
        developmentLanguages: newDev || oldDev,
        aiApplication: newAI || oldAI,
      });
    }

    // Certificates V2 (K8-5-44-C: simplified backfill)
    if (v2.certificates && v2.certificates.length > 0) {
      const langExams = v2.certificates.filter((c: any) => c.type === "language_exam");
      if (langExams.length > 0) {
        setLanguageExamDrafts(langExams.map((c: any) => ({
          name: c.name || c.level || "",
          score: c.score || "",
        })));
      }
      const normalCerts = v2.certificates.filter((c: any) => c.type !== "language_exam");
      if (normalCerts.length > 0) {
        const certList: CertificateDraft[] = normalCerts.map((c: any, i: number) => ({
          id: i + 1,
          name: c.name || "",
          description: c.description || [c.type, c.level, c.score, c.issued_at].filter(Boolean).join(" / ") || "",
        }));
        setCertificateDrafts(certList);
      }
    }

    // Awards V2 (K8-5-44-C: simplified)
    if (v2.awards && v2.awards.length > 0) {
      const awdList: AwardDraft[] = v2.awards.map((a: any, i: number) => ({
        id: i + 1,
        name: a.name || "",
        description: a.description || [a.type, a.level, a.awarded_at].filter(Boolean).join(" / ") || "",
      }));
      setAwardDrafts(awdList);
    }

    // Portfolio V2 (K8-5-25-E: works+homepage with old-field fallback)
    if (v2.portfolio) {
      const port = v2.portfolio as any;
      const newWorks = port.works as any[];
      let works: WorkDraft[] = [];
      if (newWorks && newWorks.length > 0) {
        works = newWorks.map((w: any, i: number) => ({
          id: i + 1,
          name: w.name || "",
          link: w.link || "",
          password: w.password || "",
        }));
      } else {
        works = [];
        if ((port as any).works) {
          const aw = (port as any).works as any[];
          aw.forEach((w: any) => works.push({ id: works.length + 1, name: w.name || "", link: w.link || w.url || "", password: w.password || "" }));
        }
      }
      if (works.length === 0) works = [{ id: 1, name: "", link: "", password: "" }];
      setPortfolioDraft({
        works,
        homepage: port.homepage || "",
      });
    }

    // Extra Info V2 (K8-5-44-C: merge self_evaluation into additional_info)
    if (v2.extraInfo) {
      const ei = v2.extraInfo as any;
      const merged = [ei.self_evaluation, ei.additional_info].filter(Boolean).join("\n\n");
      setExtraInfoDraft({ additionalInfo: merged || "" });
    }
  }, [formVm, mode]);

  // ── Save handler ─────────────────────────────────────────────────────────────

  // K8-5-22-J: Data cleaning helpers for V2 payload assembly.
  function splitList(value: string): string[] {
    return value
      .split(/[，,、\n/]/)
      .map((s) => s.trim())
      .filter((s) => s.length > 0);
  }
  function splitLines(value: string): string[] {
    return value
      .split(/\n/)
      .map((s) => s.trim())
      .filter((s) => s.length > 0);
  }
  function firstOf(list: string[]): string {
    return list.length > 0 ? list[0] : "";
  }

  async function handleSave() {
    setSaveSuccessMsg("");
    setSaveErrorMsg("");
    setSaving(true);

    // K8-5-22-J/K8-5-25-C: V2 payload — school/degree/major/grad_year from educations[0]
    const primaryEducation = educations[0];
    const v2Profile = {
      name: personalName || undefined,
      school: primaryEducation?.school || undefined,
      degree: primaryEducation?.degree || undefined,
      major: primaryEducation?.major || undefined,
      graduation_year: primaryEducation?.endDate || undefined,
      current_city: currentCity || undefined,
    };

    const v2Educations = educations.length > 0 ? educations.map((e) => ({
      school: e.school || undefined,
      degree: e.degree || undefined,
      major: e.major || undefined,
      department: e.department || undefined,
      start_date: e.startDate || undefined,
      end_date: e.endDate || undefined,
      gpa: e.gpa || undefined,
      gpa_scale: e.gpaScale || undefined,
      rank: e.rank || undefined,
      core_courses: splitList(e.coreCourses || e.courses || ""),
      research_direction: e.researchDirection || undefined,
      lab: e.lab || undefined,
      paper_summary: e.paperSummary || undefined,
    })) : undefined;

    const v2Internships = internships.length > 0 ? internships.map((i) => ({
      company: i.company || undefined,
      position: i.position || undefined,
      start_date: i.startDate || undefined,
      end_date: i.endDate || undefined,
      description: i.content || undefined,
    })) : undefined;

    const v2Projects = projects.length > 0 ? projects.map((p) => ({
      name: p.name || undefined,
      role: p.role || undefined,
      start_date: p.startDate || undefined,
      end_date: p.endDate || undefined,
      description: p.desc || undefined,
    })) : undefined;

    // K8-5-44-C: skills simplified to professional_skills + ai_application
    const devLanguages = splitList(skillsDraft.developmentLanguages || "");
    const v2Skills = {
      professional_skills: devLanguages.length > 0 ? devLanguages : undefined,
      ai_application: skillsDraft.aiApplication || undefined,
    };

    // K8-5-44-C: certificates simplified to name + description
    const languageExamCertificates = languageExamDrafts
      .filter((le) => le.name.trim() || le.score.trim())
      .map((le) => ({
        name: le.name.trim() || undefined,
        description: le.score.trim() ? `${le.score.trim()}` : undefined,
      }));
    const allCertificates = [
      ...languageExamCertificates,
      ...certificateDrafts.filter((c) => c.name.trim()).map((c) => ({
        name: c.name || undefined,
        description: c.description || undefined,
      })),
    ];
    const v2Certificates = allCertificates.length > 0 ? allCertificates : undefined;

    const v2Awards = awardDrafts.length > 0 ? awardDrafts.map((a) => ({
      name: a.name || undefined,
      description: a.description || undefined,
    })) : undefined;

    // K8-5-25-E: Tencent-style works + homepage
    const v2Portfolio = {
      works: portfolioDraft.works
        .filter((w) => w.name || w.link)
        .map((w) => ({
          name: w.name || undefined,
          link: w.link || undefined,
          password: w.password || undefined,
        })),
      homepage: portfolioDraft.homepage || undefined,
    };

    const v2ExtraInfo = {
      additional_info: extraInfoDraft.additionalInfo || undefined,
    };

    // V1 backward-compatible fields
    const v1EducationCompat = educations.map((e) => ({
      school: e.school,
      degree: e.degree,
      major: e.major,
      start: e.startDate,
      end: e.endDate,
    }));
    const v1ExperienceCompat = internships.map((i) => ({
      company: i.company,
      role: i.position,
      start: i.startDate,
      end: i.endDate,
      description: i.content,
    }));
    const v1ProjectCompat = projects.map((p) => ({
      name: p.name,
      role: p.role,
      description: p.desc,
    }));

    const payload: Record<string, unknown> = {
      schema_version: 2,
      resume_name: resumeName || personalName || "未命名简历",
      target_direction: "",
      profile: v2Profile,
      educations: v2Educations?.length ? v2Educations : educations.length > 0 ? v1EducationCompat : undefined,
      internships: v2Internships?.length ? v2Internships : undefined,
      projects: v2Projects?.length ? v2Projects : undefined,
      skills: v2Skills,
      certificates: v2Certificates?.length ? v2Certificates : undefined,
      awards: v2Awards?.length ? v2Awards : undefined,
      portfolio_links: v2Portfolio,
      extra_info: v2ExtraInfo,
      name: resumeName || personalName || "未命名简历",
      phone: phone || "",
      email: email || "",
      location: currentCity || "",
      target_city: "",
      target_role: "",
      education: v1EducationCompat,
      experience: v1ExperienceCompat,
      project: v1ProjectCompat,
      skills_v1: devLanguages.length > 0 ? devLanguages : skills.filter((s) => s.trim().length > 0),
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
    setRecentlyAdded({ type: "project", key: String(newId) });
    setProjects((ps) => [
      ...ps,
      { id: newId, name: "", desc: "", role: "", startDate: "", endDate: "" },
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

  // K8-5-25-E: works helpers
  function addWork() {
    const newId = Math.max(0, ...portfolioDraft.works.map((w) => w.id)) + 1;
    setRecentlyAdded({ type: "work", key: String(newId) });
    setPortfolioDraft((prev) => ({
      ...prev,
      works: [...prev.works, { id: newId, name: "", link: "", password: "" }],
    }));
  }
  function updateWork(id: number, field: keyof WorkDraft, value: string) {
    setPortfolioDraft((prev) => ({
      ...prev,
      works: prev.works.map((w) => (w.id === id ? { ...w, [field]: value } : w)),
    }));
  }
  function removeWork(id: number) {
    setPortfolioDraft((prev) => ({ ...prev, works: prev.works.filter((w) => w.id !== id) }));
  }

  function addCertificate() {
    const newId = Math.max(0, ...certificateDrafts.map((c) => c.id)) + 1;
    setRecentlyAdded({ type: "certificate", key: String(newId) });
    setCertificateDrafts((cs) => [...cs, { id: newId, name: "", description: "" }]);
  }
  function updateCertificate(id: number, field: keyof CertificateDraft, value: string) {
    setCertificateDrafts((cs) => cs.map((c) => (c.id === id ? { ...c, [field]: value } : c)));
  }
  function removeCertificate(id: number) {
    setCertificateDrafts((cs) => cs.filter((c) => c.id !== id));
  }

  // K8-5-29-B: language exam dynamic list helpers
  const addLanguageExam = () => {
    setRecentlyAdded({ type: "languageExam", key: String(languageExamDrafts.length) });
    setLanguageExamDrafts((prev) => [...prev, { name: "", score: "" }]);
  };
  const updateLanguageExam = (idx: number, field: keyof LanguageExamDraft, value: string) => {
    setLanguageExamDrafts((prev) => prev.map((le, i) => (i === idx ? { ...le, [field]: value } : le)));
  };
  const removeLanguageExam = (idx: number) => setLanguageExamDrafts((prev) => prev.filter((_, i) => i !== idx));

  function addAward() {
    const newId = Math.max(0, ...awardDrafts.map((a) => a.id)) + 1;
    setRecentlyAdded({ type: "award", key: String(newId) });
    setAwardDrafts((as) => [...as, { id: newId, name: "", description: "" }]);
  }
  function updateAward(id: number, field: keyof AwardDraft, value: string) {
    setAwardDrafts((as) => as.map((a) => (a.id === id ? { ...a, [field]: value } : a)));
  }
  function removeAward(id: number) {
    setAwardDrafts((as) => as.filter((a) => a.id !== id));
  }

  function addEducation() {
    const newId = Math.max(0, ...educations.map((e) => e.id)) + 1;
    setRecentlyAdded({ type: "education", key: String(newId) });
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
    setRecentlyAdded({ type: "internship", key: String(newId) });
    setInternships((is) => [
      ...is,
      {
        id: newId, company: "", position: "", startDate: "", endDate: "", content: "",
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

  // ── K8-5-41-B：有内容模块默认展开判断 ────────────────────────────────
  const isEditingExistingResume = Boolean(editingVersionNo);
  const hasInternshipContent = internships.length > 0;
  const hasCertificateAwardContent =
    languageExamDrafts.length > 0 ||
    certificateDrafts.length > 0 ||
    awardDrafts.length > 0;
  const hasPortfolioExtraContent =
    portfolioDraft.works.some((w) => w.name.trim() || w.link.trim()) ||
    !!portfolioDraft.homepage.trim() ||
    !!summary.trim() ||
    !!extraInfoDraft.additionalInfo.trim();
  // K8-5-41-C：autoOpenKey 包含 resumeName，数据回填完成后 key 会变化（""→实际名称）
  const sectionAutoOpenKey = `${editingVersionNo ?? "_new"}:${resumeName || "_empty"}`;

  return (
    <ResumeEditorShell
      title={mode === "add" ? "创建简历" : "编辑简历"}
      subtitle="完善你的经历和技能，让 Offerr鹅 更准确地完成岗位推荐、匹配分析和简历优化。"
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

      {/* ── Hint strip ──────────────────────────────────────────── */}
      <div className="flex items-center gap-2 px-4 py-2.5 rounded-xl bg-blue-50 border border-blue-100">
        <Sparkles className="w-3.5 h-3.5 text-[#0052D9] flex-shrink-0" />
        <span className="text-xs text-[#0052D9]">
          当前简历可用于：<strong>岗位推荐</strong> / <strong>岗位匹配分析</strong> / <strong>简历优化</strong>
        </span>
        <span className="w-px h-3.5 bg-blue-200 mx-1" />
        <span className="text-xs text-gray-400">建议仅作投递前参考，不替代真实经历核实。</span>
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
              placeholder="用于区分不同简历版本，例如「后端开发通用版」「云计算实习版」「春招通用版」"
            />
          </div>
          <div>
            <div className="flex items-start gap-2 rounded-xl bg-blue-50/60 border border-blue-100/40 px-4 py-3">
              <Sparkles className="w-4 h-4 text-[#0052D9] mt-0.5" />
              <p className="text-[13px] text-slate-600 leading-relaxed">
                保存后，最新版本将自动用于岗位推荐、岗位匹配分析和简历优化建议。如果你暂时不确定求职方向，可以留空求职意向模块，Offerr鹅 会基于你的教育、技能和项目经历推荐适合的岗位。
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
          基础信息用于简历展示和基础画像。手机号 / 邮箱不会用于分析匹配。
        </p>
      </BasicInfoSection>

      {/* K8-5-25-C: 求职意向已删除 — 岗位推荐不需要用户预填意向 */}

      {/* ── Module 3: 教育经历 ──────────────────────────────────────── */}
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
              <div key={edu.id} className={`rounded-2xl bg-gray-50/70 ring-1 ring-gray-100 p-5 space-y-4 ${recentlyAdded?.type === "education" && recentlyAdded.key === String(edu.id) ? "resume-item-enter" : ""}`}>
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
                      <YearMonthPicker value={edu.startDate} onChange={(v) => updateEdu(edu.id, "startDate", v)} placeholder="开始时间" />
                      <span className="text-slate-400">—</span>
                      <YearMonthPicker value={edu.endDate} onChange={(v) => updateEdu(edu.id, "endDate", v)} placeholder="结束时间" />
                    </div>
                  </div>
                </div>

                {/* Recommended fields */}
                <div className="border-t border-gray-200 pt-4">
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
              <div key={proj.id} className={`rounded-2xl bg-gray-50/70 ring-1 ring-gray-100 p-5 space-y-4 ${recentlyAdded?.type === "project" && recentlyAdded.key === String(proj.id) ? "resume-item-enter" : ""}`}>
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
                <div>
                  <label className="mb-1.5 block text-sm font-medium text-slate-700">项目时间</label>
                  <div className="flex items-center gap-2">
                    <YearMonthPicker value={proj.startDate} onChange={(v) => updateProject(proj.id, "startDate", v)} placeholder="开始时间" />
                    <span className="text-slate-400">—</span>
                    <YearMonthPicker value={proj.endDate} onChange={(v) => updateProject(proj.id, "endDate", v)} placeholder="结束时间" />
                  </div>
                </div>
                <div>
                  <label className="mb-1.5 block text-sm font-medium text-slate-700">在项目中担任的角色</label>
                  <Input className="h-10" value={proj.role} onChange={(e) => updateProject(proj.id, "role", e.target.value)} placeholder="如：接口设计、数据库设计、缓存优化" />
                </div>
                <div>
                  <label className="mb-1.5 block text-sm font-medium text-slate-700">描述</label>
                  <Textarea className="min-h-[120px] resize-none" rows={4} value={proj.desc} onChange={(e) => updateProject(proj.id, "desc", e.target.value)} placeholder="描述项目背景、你负责的部分、使用的技术栈和主要成果" />
                </div>
              </div>
            ))}
          </div>
        )}
      </ProjectSection>

      {/* ── Module 6: 技能信息 ──────────────────────────────────────── */}
      <SkillsSection>
        <div className="border-t border-gray-200 pt-5 mt-5 space-y-4">

          {/* 2-专业技能 / 工具能力 */}
          <div>
            <label className="mb-1.5 block text-sm font-medium text-slate-700">专业技能 / 工具能力</label>
            <Input className="h-10" value={skillsDraft.developmentLanguages} onChange={(e) => updateSkillsDraft("developmentLanguages", e.target.value)} placeholder="技术岗可填 Go / Java / SQL；产品运营可填 Axure / 数据分析 / 用户调研；职能商务可填 Excel / PPT / 商务沟通（逗号、顿号或换行分隔）" />
            <p className="mt-1.5 text-xs text-slate-400">可填写开发语言、设计工具、数据分析方法、办公软件等，多个用逗号、顿号或换行分隔。</p>
          </div>

          {/* 3-AI 应用技能 */}
          <div>
            <label className="mb-1.5 block text-sm font-medium text-slate-700">AI 应用技能 <span className="text-xs font-normal text-amber-600">推荐</span></label>
            <Textarea
              className="min-h-[88px] resize-none" rows={3}
              value={skillsDraft.aiApplication}
              onChange={(e) => updateSkillsDraft("aiApplication", e.target.value)}
              placeholder="例如：使用 CodeBuddy 辅助需求拆解、代码生成、Bug 定位、接口文档、测试用例和项目复盘。"
            />
          </div>
        </div>

        <p className="mt-5 text-[12px] text-slate-400 border-t border-gray-100 pt-4">
          技能信息有助于 Offerr鹅 提升岗位推荐、匹配分析和简历优化质量。
        </p>
      </SkillsSection>

      {/* ── Module 7: 实习经历 ──────────────────────────────────────── */}
      <InternshipSection
        defaultOpen={isEditingExistingResume && hasInternshipContent}
        autoOpenKey={sectionAutoOpenKey}
      >
        {internships.length === 0 ? (
          <div className="rounded-xl bg-gray-50/70 py-8 text-center">
            <p className="text-sm text-slate-500 mb-1">暂无实习经历</p>
            <p className="text-xs text-slate-400 mb-3">可以补充实习公司、岗位职责和成果。</p>
            <button onClick={addInternship} className="text-[13px] font-medium text-[#0052D9] hover:underline">+ 添加实习经历</button>
          </div>
        ) : (
          <div className="space-y-4">
            {internships.map((intern, idx) => (
              <div key={intern.id} className={`rounded-2xl bg-gray-50/70 ring-1 ring-gray-100 p-5 space-y-4 ${recentlyAdded?.type === "internship" && recentlyAdded.key === String(intern.id) ? "resume-item-enter" : ""}`}>
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
                    <label className="mb-1.5 block text-sm font-medium text-slate-700">职位 <span className="text-red-500">*</span></label>
                    <Input className="h-10" value={intern.position} onChange={(e) => updateIntern(intern.id, "position", e.target.value)} placeholder="职位" />
                  </div>
                </div>
                <div>
                  <label className="mb-1.5 block text-sm font-medium text-slate-700">时间</label>
                  <div className="flex items-center gap-2">
                    <YearMonthPicker value={intern.startDate} onChange={(v) => updateIntern(intern.id, "startDate", v)} placeholder="开始时间" />
                    <span className="text-slate-400">—</span>
                    <YearMonthPicker value={intern.endDate} onChange={(v) => updateIntern(intern.id, "endDate", v)} placeholder="结束时间" />
                  </div>
                </div>
                <div>
                  <label className="mb-1.5 block text-sm font-medium text-slate-700">描述</label>
                  <Textarea className="min-h-[120px] resize-none" rows={4} value={intern.content} onChange={(e) => updateIntern(intern.id, "content", e.target.value)} placeholder="描述工作内容、使用的技术栈和主要成果" />
                </div>
              </div>
            ))}
          </div>
        )}
      </InternshipSection>

      {/* ── Module 7: 证书与获奖 (K8-5-25-E: merged) ──────────────────────── */}
      <CertificatesSection
        onAdd={addCertificate}
        defaultOpen={isEditingExistingResume && hasCertificateAwardContent}
        autoOpenKey={sectionAutoOpenKey}
      >
        {/* K8-5-29-B: 外语考试 / 等级 — dynamic list */}
        <div className="mb-5 pb-5 border-b border-gray-100">
          <h4 className="text-sm font-medium text-slate-600 mb-3">外语考试 / 等级</h4>
          {languageExamDrafts.length === 0 ? (
            <div className="rounded-xl bg-gray-50/70 py-6 text-center">
              <p className="text-sm text-slate-500 mb-1">暂无外语考试</p>
              <p className="text-xs text-slate-400 mb-3">可以添加 CET-4、CET-6、雅思、托福、日语等级等。</p>
              <button onClick={addLanguageExam} className="text-[13px] font-medium text-[#0052D9] hover:underline">
                + 添加外语考试
              </button>
            </div>
          ) : (
            <>
              {languageExamDrafts.map((le, idx) => (
                <div key={idx} className={`rounded-2xl bg-gray-50/70 ring-1 ring-gray-100 p-4 space-y-3 mb-3 ${recentlyAdded?.type === "languageExam" && recentlyAdded.key === String(idx) ? "resume-item-enter" : ""}`}>
                  <div className="flex items-center justify-between">
                    <span className="text-xs font-medium text-slate-500">外语考试 {idx + 1}</span>
                    <button onClick={() => removeLanguageExam(idx)} className="text-slate-400 hover:text-red-500 transition-colors">
                      <Trash2 className="w-4 h-4" />
                    </button>
                  </div>
                  <div className="grid grid-cols-1 gap-x-5 gap-y-4 sm:grid-cols-2">
                    <div>
                      <label className="mb-1.5 block text-sm font-medium text-slate-700">考试类型</label>
                      <Input className="h-10" value={le.name} onChange={(e) => updateLanguageExam(idx, "name", e.target.value)} placeholder="如：CET-4、CET-6、IELTS、TOEFL、日语 N2" />
                    </div>
                    <div>
                      <label className="mb-1.5 block text-sm font-medium text-slate-700">分数 / 等级</label>
                      <Input className="h-10" value={le.score} onChange={(e) => updateLanguageExam(idx, "score", e.target.value)} placeholder="例如：503 / 6.5 / N2 / 95" />
                    </div>
                  </div>
                </div>
              ))}
              <button onClick={addLanguageExam} className="text-[13px] font-medium text-[#0052D9] hover:underline">
                + 添加外语考试
              </button>
            </>
          )}
        </div>

        {certificateDrafts.length === 0 && awardDrafts.length === 0 ? (
          <div className="rounded-xl bg-gray-50/70 py-8 text-center">
            <p className="text-sm text-slate-500 mb-1">暂无证书或获奖</p>
            <p className="text-xs text-slate-400 mb-3">可以添加语言成绩、云计算认证、竞赛获奖等。</p>
            <button onClick={addCertificate} className="text-[13px] font-medium text-[#0052D9] hover:underline mr-3">+ 添加证书</button>
            <button onClick={addAward} className="text-[13px] font-medium text-[#0052D9] hover:underline">+ 添加获奖</button>
          </div>
        ) : (
          <>
            {/* Certificate entries */}
            {certificateDrafts.length > 0 && <h4 className="text-sm font-medium text-slate-600 mb-3">证书</h4>}
            {certificateDrafts.map((cert, idx) => (
              <div key={cert.id} className={`rounded-2xl bg-gray-50/70 ring-1 ring-gray-100 p-5 space-y-4 mb-4 ${recentlyAdded?.type === "certificate" && recentlyAdded.key === String(cert.id) ? "resume-item-enter" : ""}`}>
                <div className="flex items-center justify-between">
                  <span className="text-xs font-medium text-slate-500">证书 {idx + 1}</span>
                  <button onClick={() => removeCertificate(cert.id)} className="text-slate-400 hover:text-red-500 transition-colors"><Trash2 className="w-4 h-4" /></button>
                </div>
                <div><label className="mb-1.5 block text-sm font-medium text-slate-700">证书名称 <span className="text-red-500">*</span></label><Input className="h-10" value={cert.name} onChange={(e) => updateCertificate(cert.id, "name", e.target.value)} placeholder="如：CET-4 503、阿里云 ACP" /></div>
                <div><label className="mb-1.5 block text-sm font-medium text-slate-700">证书说明</label><Textarea className="min-h-[72px] resize-none" rows={2} value={cert.description} onChange={(e) => updateCertificate(cert.id, "description", e.target.value)} placeholder="如：类型、级别、成绩/分数、获得时间等补充信息" /></div>
              </div>
            ))}
            {/* Award entries */}
            {awardDrafts.length > 0 && <h4 className="text-sm font-medium text-slate-600 mb-3 mt-2">获奖</h4>}
            {awardDrafts.map((award, idx) => (
              <div key={award.id} className={`rounded-2xl bg-gray-50/70 ring-1 ring-gray-100 p-5 space-y-4 mb-4 ${recentlyAdded?.type === "award" && recentlyAdded.key === String(award.id) ? "resume-item-enter" : ""}`}>
                <div className="flex items-center justify-between">
                  <span className="text-xs font-medium text-slate-500">获奖 {idx + 1}</span>
                  <button onClick={() => removeAward(award.id)} className="text-slate-400 hover:text-red-500 transition-colors"><Trash2 className="w-4 h-4" /></button>
                </div>
                <div><label className="mb-1.5 block text-sm font-medium text-slate-700">奖项名称 <span className="text-red-500">*</span></label><Input className="h-10" value={award.name} onChange={(e) => updateAward(award.id, "name", e.target.value)} placeholder="如：国家奖学金、ACM-ICPC 银牌" /></div>
                <div><label className="mb-1.5 block text-sm font-medium text-slate-700">奖项说明</label><Textarea className="min-h-[72px] resize-none" rows={2} value={award.description} onChange={(e) => updateAward(award.id, "description", e.target.value)} placeholder="如：获奖类型、级别、时间、原因等补充信息" /></div>
              </div>
            ))}
            {certificateDrafts.length === 0 && awardDrafts.length === 0 && null}
            {(certificateDrafts.length > 0 || awardDrafts.length > 0) && (
              <div className="flex gap-2 mt-2">
                <button onClick={addCertificate} className="text-[13px] font-medium text-[#0052D9] hover:underline">+ 添加证书</button>
                <button onClick={addAward} className="text-[13px] font-medium text-[#0052D9] hover:underline">+ 添加获奖</button>
              </div>
            )}
          </>
        )}
      </CertificatesSection>

      {/* ── Module 8: 作品或个人主页 (K8-5-25-E: works+homepage) ── */}
      <PortfolioExtraSection
        defaultOpen={isEditingExistingResume && hasPortfolioExtraContent}
        autoOpenKey={sectionAutoOpenKey}
      >
        <p className="text-[13px] text-slate-500 mb-5">只支持链接，不支持附件上传。</p>
        {portfolioDraft.works.length === 0 && (
          <button onClick={addWork} className="text-[13px] font-medium text-[#0052D9] hover:underline mb-4 inline-block">+ 添加作品链接</button>
        )}
        {portfolioDraft.works.map((work, idx) => (
          <div key={work.id} className={`rounded-2xl bg-gray-50/70 ring-1 ring-gray-100 p-5 space-y-4 mb-4 ${recentlyAdded?.type === "work" && recentlyAdded.key === String(work.id) ? "resume-item-enter" : ""}`}>
            <div className="flex items-center justify-between">
              <span className="text-xs font-medium text-slate-500">作品链接 {idx + 1}</span>
              <button onClick={() => removeWork(work.id)} className="text-slate-400 hover:text-red-500 transition-colors"><Trash2 className="w-4 h-4" /></button>
            </div>
            <div><label className="mb-1.5 block text-sm font-medium text-slate-700">作品名称</label><Input className="h-10" value={work.name} onChange={(e) => updateWork(work.id, "name", e.target.value)} placeholder="如：分布式在线考试系统" /></div>
            <div><label className="mb-1.5 block text-sm font-medium text-slate-700">作品链接</label><Input className="h-10" value={work.link} onChange={(e) => updateWork(work.id, "link", e.target.value)} placeholder="如：https://cnb.cool/xxx" /></div>
            <div><label className="mb-1.5 block text-sm font-medium text-slate-700">密码或提取码</label><Input className="h-10" value={work.password} onChange={(e) => updateWork(work.id, "password", e.target.value)} placeholder="如无需密码留空" /></div>
          </div>
        ))}
        <button onClick={addWork} className="text-[13px] font-medium text-[#0052D9] hover:underline mb-4 inline-block">+ 添加作品链接</button>
        <div className="mt-4"><label className="mb-1.5 block text-sm font-medium text-slate-700">个人主页</label><Input className="h-10" value={portfolioDraft.homepage} onChange={(e) => setPortfolioDraft((prev) => ({ ...prev, homepage: e.target.value }))} placeholder="如：https://github.com/xxx" /></div>
      </PortfolioExtraSection>

      {/* ── Module 9: 补充信息 (K8-5-44-C: 替代自我评价) ───────────────── */}
      <PortfolioExtraSection
        title="补充信息"
        defaultOpen={isEditingExistingResume && hasPortfolioExtraContent}
        autoOpenKey={sectionAutoOpenKey}
      >
        <p className="text-[13px] text-slate-500 mb-5">补充其他未覆盖的信息，如技术背景、求职意向、社团经历、志愿服务、兴趣爱好等。</p>
        <Textarea className="min-h-[140px]" value={extraInfoDraft.additionalInfo} onChange={(e) => setExtraInfoDraft({ additionalInfo: e.target.value })} placeholder="补充介绍你的技术方向、项目经验、学习能力、目标岗位以及其他信息" />
        <div className="mt-2 flex items-center justify-between">
          <span className="text-[12px] text-slate-400">建议包含：技术方向、项目经验、学习能力、目标岗位。</span>
          <span className="text-[12px] text-slate-400">当前约 {extraInfoDraft.additionalInfo.length} 字</span>
        </div>
      </PortfolioExtraSection>

    </ResumeEditorShell>
  );
}
