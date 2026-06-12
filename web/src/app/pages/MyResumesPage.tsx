import React, { useState, useCallback, useEffect, useMemo, useRef } from "react";
import { Link, useLocation, useNavigate, useSearchParams } from "react-router";
import { ChevronLeft, ChevronUp, ChevronDown, Trash2, Plus, Sparkles, CheckCircle2, XCircle, AlertCircle, HelpCircle, Loader2, FileText } from "lucide-react";
import { createResume, updateResume, getCurrentResume, getResumes, getResumeById, checkResumeName } from "../../api/resumes";
import type { ResumeManagementViewModel } from "../../adapters";
import { toResumeManagementViewModel } from "../../adapters";
import type { ResumeEducation, ResumeInternship, ResumeProject, ResumeLanguageExam, ResumeCertificate, ResumeAward, ResumeWork, SaveResumePayload } from "../../types";
import { ResumeSkeleton } from "../components/ui/ResumeSkeleton";
import { YearMonthPicker } from "../components/resume-editor/YearMonthPicker";
import { PageError, PageLoading } from "../components/PageState";
import { isValidResumeId } from "../../utils/redirect";
import { isApiClientError } from "../../api/errors";
import { Input } from "../components/ui/input";
import { AutoGrowTextarea } from "../components/ui/auto-grow-textarea";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "../components/ui/select";
import { ResumeEditorShell, ResumeEditorSaveCard, ResumeBasicsSection, BasicInfoSection, EducationSection, ProjectSection, SkillsSection, InternshipSection, CertificatesSection, PortfolioExtraSection } from "../components/resume-editor";
import { ResumeEditorEmptyState } from "../components/resume-editor/ResumeEditorEmptyState";
import { ResumeEditorStatusSidebar, type ReadinessCategory } from "../components/resume-editor/ResumeEditorStatusSidebar";
import { AppConfirmDialog } from "../components/common/AppConfirmDialog";
import { cn } from "../components/ui/utils";
import { AnimatedListItem } from "../components/ui/animated-list-item";

interface Project { id: number; name: string; desc: string; role: string; startDate: string; endDate: string; }
interface Education { id: number; school: string; major: string; degree: string; startDate: string; endDate: string; department: string; gpa: string; gpaScale: string; rank: string; coreCourses: string; researchDirection: string; advisor: string; lab: string; paperSummary: string; }
interface Internship { id: number; company: string; position: string; startDate: string; endDate: string; content: string; }
interface SkillsDraft { developmentLanguages: string; aiApplication: string; }
interface LanguageExamDraft { id: number; name: string; score: string; }
interface CertificateDraft { id: number; name: string; description: string; }
interface AwardDraft { id: number; name: string; description: string; }
interface WorkDraft { id: number; name: string; link: string; password: string; }
interface PortfolioDraft { works: WorkDraft[]; homepage: string; }
interface ExtraInfoDraft { additionalInfo: string; }

const NAV_SECTIONS = [
  { id: "resume-basic", label: "简历基础" },
  { id: "personal-info", label: "基础信息" },
  { id: "education", label: "教育经历" },
  { id: "internship", label: "实习经历" },
  { id: "projects", label: "项目经历" },
  { id: "skills", label: "技能信息" },
  { id: "extra-info", label: "补充信息" },
];

export function MyResumesPage() {
  const location = useLocation(); const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const idParam = searchParams.get("id");
  const hasInvalidId = idParam !== null && !isValidResumeId(idParam);
  const editingResumeId = idParam && /^\d+$/.test(idParam) ? Number(idParam) : null;
  const isCreateMode = !editingResumeId;
  const mode: "edit" | "add" = location.pathname.includes("/new") ? "add" : "edit";

  useEffect(() => { if (hasInvalidId) { navigate("/resumes", { replace: true }); } }, [hasInvalidId, navigate]);

  const [saving, setSaving] = useState(false);
  const [saveSuccessMsg, setSaveSuccessMsg] = useState("");
  const [saveErrorMsg, setSaveErrorMsg] = useState("");
  const [formVm, setFormVm] = useState<ResumeManagementViewModel | null>(null);
  const [loading, setLoading] = useState(mode === "edit");
  const [resumeError, setResumeError] = useState<string | null>(null);
  const [retryCount, setRetryCount] = useState(0);

  useEffect(() => {
    if (mode !== "edit") return;
    let cancelled = false;
    setLoading(true); setResumeError(null);
    if (editingResumeId !== null && editingResumeId > 0) {
      getResumeById(editingResumeId).then((resume) => {
        if (cancelled) return;
        try { setFormVm(toResumeManagementViewModel(resume)); setLoading(false); }
        catch (adapterError) { setLoading(false); setResumeError("简历数据解析失败，请稍后重试。"); }
      }).catch((err: unknown) => {
        if (cancelled) return;
        setLoading(false);
        if (isApiClientError(err) && err.status === 401) {
          navigate(`/?auth=login&redirect=${encodeURIComponent(`/resumes/edit?id=${editingResumeId}`)}`, { replace: true });
        } else if (isApiClientError(err) && err.status === 404) {
          setResumeError("该简历不存在或已被删除。");
        } else { setResumeError("指定简历加载失败，请稍后重试。"); }
      });
      return () => { cancelled = true; };
    }
    // Create mode: blank form, no API call needed — resetForCreate handles cleanup
    setFormVm(null);
    setLoading(false);
    return () => { cancelled = true; };
  }, [mode, retryCount, editingResumeId]);

  const handleRetryLoad = useCallback(() => setRetryCount((c) => c + 1), []);

  // ── Form reset for create mode ─────────────────────────────────────────────
  function resetFormForCreate() {
    setFormVm(null);
    setResumeName("");
    setPersonalName("");
    setPhone("");
    setEmail("");
    setCurrentCity("");
    setEducations([]);
    setInternships([]);
    setProjects([]);
    setSkillsDraft({ developmentLanguages: "", aiApplication: "" });
    setLanguageExamDrafts([]);
    setCertificateDrafts([]);
    setAwardDrafts([]);
    setPortfolioDraft({ works: [{ id: 1, name: "", link: "", password: "" }], homepage: "" });
    setExtraInfoDraft({ additionalInfo: "" });
    setNameCheckStatus("idle");
    setNameCheckMessage("");
    setSaveSuccessMsg("");
    setSaveErrorMsg("");
  }

  // Reset form whenever we enter create mode (no editingResumeId)
  useEffect(() => {
    if (isCreateMode) resetFormForCreate();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [editingResumeId]);

  // ── Form alert dialog state ────────────────────────────────────────────────
  const [formAlert, setFormAlert] = useState<{
    open: boolean;
    title: string;
    description: string;
    variant?: "default" | "danger";
  }>({ open: false, title: "", description: "", variant: "default" });
  const [resumeName, setResumeName] = useState("");

  // K8-5-49-B: name check state
  type NameCheckStatus = "idle" | "checking" | "available" | "duplicate" | "invalid" | "error";
  const [nameCheckStatus, setNameCheckStatus] = useState<NameCheckStatus>("idle");
  const [nameCheckMessage, setNameCheckMessage] = useState("");

  async function handleNameBlur() {
    const trimmed = resumeName.trim();
    if (!trimmed) {
      setNameCheckStatus("invalid");
      setNameCheckMessage("简历名称不能为空");
      return;
    }
    setNameCheckStatus("checking");
    setNameCheckMessage("正在检查...");
    try {
      const result = await checkResumeName(trimmed, editingResumeId ?? undefined);
      if (result.available) {
        setNameCheckStatus("available");
        setNameCheckMessage("名称可用，保存后将用于岗位推荐、匹配分析和简历优化建议。");
      } else {
        setNameCheckStatus("duplicate");
        setNameCheckMessage(result.reason || "该简历名称已存在");
      }
    } catch {
      setNameCheckStatus("error");
      setNameCheckMessage("暂时无法校验名称，请稍后重试。");
    }
  }

  // Save button disabled when actively saving or still checking name
  const saveDisabled = saving || nameCheckStatus === "checking";
  const [personalName, setPersonalName] = useState(""); const [phone, setPhone] = useState("");
  const [email, setEmail] = useState(""); const [currentCity, setCurrentCity] = useState("");
  const [educations, setEducations] = useState<Education[]>([]);
  const [internships, setInternships] = useState<Internship[]>([]);
  const [projects, setProjects] = useState<Project[]>([]);
  const [skillsDraft, setSkillsDraft] = useState<SkillsDraft>({ developmentLanguages: "", aiApplication: "" });
  type RecentKey = "education" | "project" | "internship" | "certificate" | "award" | "languageExam" | "work";
  const [recentlyAdded, setRecentlyAdded] = useState<{ type: RecentKey; key: string } | null>(null);
  const [languageExamDrafts, setLanguageExamDrafts] = useState<LanguageExamDraft[]>([]);
  const [certificateDrafts, setCertificateDrafts] = useState<CertificateDraft[]>([]);
  const [awardDrafts, setAwardDrafts] = useState<AwardDraft[]>([]);
  const [portfolioDraft, setPortfolioDraft] = useState<PortfolioDraft>({ works: [{ id: 1, name: "", link: "", password: "" }], homepage: "" });
  const [extraInfoDraft, setExtraInfoDraft] = useState<ExtraInfoDraft>({ additionalInfo: "" });

  // ── Delayed remove animation state ─────────────────────────────────────────
  const [removingItemKeys, setRemovingItemKeys] = useState<Set<string>>(new Set());
  const removeTimeoutsRef = useRef<number[]>([]);

  // K9-UI-ResumeEdit-D-B: cleanup timeouts on unmount
  useEffect(() => {
    return () => {
      removeTimeoutsRef.current.forEach((id) => window.clearTimeout(id));
      removeTimeoutsRef.current = [];
    };
  }, []);

  const markRemoving = useCallback((key: string, removeFn: () => void) => {
    setRemovingItemKeys((prev) => {
      if (prev.has(key)) return prev;
      const next = new Set(prev);
      next.add(key);
      return next;
    });
    const timeoutId = window.setTimeout(() => {
      removeFn();
      setRemovingItemKeys((prev) => {
        const next = new Set(prev);
        next.delete(key);
        return next;
      });
    }, 170);
    removeTimeoutsRef.current.push(timeoutId);
  }, []);

  // ── Demo content fill ─────────────────────────────────────────────────────
  const [fillingDemo, setFillingDemo] = useState(false);
  const [demoError, setDemoError] = useState("");

  function fillDemoFromVm(vm: ResumeManagementViewModel | null) {
    if (vm) {
      // Fill from real API data (seed resume 201)
      setResumeName(vm.name || "示例简历");
      setNameCheckStatus("idle");
      setNameCheckMessage("");
      setPersonalName(vm.realName || "");
      setPhone(vm.phone || "");
      setEmail(vm.email || "");
      setCurrentCity(vm.city || "");
      const eduList: Education[] = (vm.educations || []).map((e, i) => ({
        id: i + 1, school: e.school || "", major: e.major || "", degree: e.degree || "本科",
        startDate: e.start_date || "", endDate: e.end_date || "",
        department: e.department || "", gpa: e.gpa || "", gpaScale: e.gpa_scale || "",
        rank: e.rank_text || "", coreCourses: e.core_courses || "",
        researchDirection: e.research_direction || "", advisor: e.advisor || "",
        lab: e.lab || "", paperSummary: e.paper_summary || "",
      }));
      setEducations(eduList);
      const intList: Internship[] = (vm.internships || []).map((it, i) => ({
        id: i + 1, company: it.company || "", position: it.position || "",
        startDate: it.start_date || "", endDate: it.end_date || "", content: it.description || "",
      }));
      setInternships(intList);
      const projList: Project[] = (vm.projects || []).map((p, i) => ({
        id: i + 1, name: p.name || "", desc: p.description || "", role: p.role || "",
        startDate: p.start_date || "", endDate: p.end_date || "",
      }));
      setProjects(projList);
      setSkillsDraft({ developmentLanguages: vm.professionalSkills || "", aiApplication: vm.aiApplication || "" });
      setLanguageExamDrafts((vm.languageExams || []).map((le, i) => ({ id: i + 1, name: le.exam_type || "", score: le.score || "" })));
      setCertificateDrafts((vm.certificates || []).map((c, i) => ({ id: i + 1, name: c.name || "", description: c.description || "" })));
      setAwardDrafts((vm.awards || []).map((a, i) => ({ id: i + 1, name: a.name || "", description: a.description || "" })));
      const ws: WorkDraft[] = (vm.works || []).map((w, i) => ({ id: i + 1, name: w.name || "", link: w.url || "", password: w.password || "" }));
      setPortfolioDraft({ works: ws.length > 0 ? ws : [{ id: 1, name: "", link: "", password: "" }], homepage: vm.homepage || "" });
      setExtraInfoDraft({ additionalInfo: vm.additionalInfo || "" });
      return;
    }

    // Fallback: hardcoded demo data when API is unreachable
    setResumeName("示例简历");
    setNameCheckStatus("idle");
    setNameCheckMessage("");
    setPersonalName("张同学");
    setPhone("13800138000");
    setEmail("zhang@example.com");
    setCurrentCity("北京");
    setEducations([{
      id: 1, school: "清华大学", major: "计算机科学与技术", degree: "硕士",
      startDate: "2024-09", endDate: "2027-06", department: "计算机科学与技术系",
      gpa: "3.8", gpaScale: "4.0", rank: "", coreCourses: "机器学习、深度学习、推荐系统",
      researchDirection: "推荐系统", advisor: "", lab: "", paperSummary: "",
    }]);
    setInternships([{
      id: 1, company: "腾讯科技", position: "后端开发实习生",
      startDate: "2025-06", endDate: "2025-09",
      content: "参与内部平台后端开发，使用 Go + MySQL + Redis 完成用户管理、权限校验和数据统计接口开发。",
    }]);
    setProjects([{
      id: 1, name: "分布式课程论坛", role: "后端负责人",
      startDate: "2024-10", endDate: "2024-12",
      desc: "基于 Go + MySQL + Redis 实现高并发课程论坛，支持帖子发布、评论、点赞、搜索、权限管理和内容审核。",
    }]);
    setSkillsDraft({
      developmentLanguages: "Go、Python、SQL、MySQL、Redis、Docker、Linux、Git",
      aiApplication: "使用 CodeBuddy 辅助需求拆解、代码生成、Bug 定位、接口文档、测试用例和项目复盘。",
    });
    setLanguageExamDrafts([{ id: 1, name: "CET-6", score: "520" }]);
    setCertificateDrafts([{ id: 1, name: "阿里云云计算基础认证 ACA", description: "了解云服务器、云数据库、对象存储等云计算基础概念。" }]);
    setAwardDrafts([{ id: 1, name: "国家奖学金", description: "研究生国家奖学金，2025 年获得。" }]);
    setPortfolioDraft({ works: [{ id: 1, name: "GitHub", link: "https://github.com/demo", password: "" }], homepage: "https://github.com/demo" });
    setExtraInfoDraft({ additionalInfo: "目标岗位：后端开发实习生。具备 Go 后端开发、接口设计、数据库建模、单元测试和前后端联调经验。" });
  }

  async function handleFillDemo() {
    setFillingDemo(true); setDemoError("");
    try {
      const demoResume = await getResumeById(201);
      fillDemoFromVm(toResumeManagementViewModel(demoResume));
    } catch {
      // API unreachable — use hardcoded demo data
      fillDemoFromVm(null);
    } finally {
      setFillingDemo(false);
    }
  }

  useEffect(() => {
    if (formVm && mode === "edit") { setResumeName(formVm.name || ""); }
  }, [formVm, mode]);

  useEffect(() => {
    if (!formVm || mode !== "edit") return;
    setPersonalName(formVm.realName || ""); setPhone(formVm.phone || ""); setEmail(formVm.email || ""); setCurrentCity(formVm.city || "");
    const eduList: Education[] = (formVm.educations || []).map((e: ResumeEducation, i: number) => ({
      id: i + 1, school: e.school || "", major: e.major || "", degree: e.degree || "本科", startDate: e.start_date || "", endDate: e.end_date || "",
      department: e.department || "", gpa: e.gpa || "", gpaScale: e.gpa_scale || "", rank: e.rank_text || "", coreCourses: e.core_courses || "",
      researchDirection: e.research_direction || "", advisor: e.advisor || "", lab: e.lab || "", paperSummary: e.paper_summary || "",
    }));
    setEducations(eduList.length > 0 ? eduList : []);
    const intList: Internship[] = (formVm.internships || []).map((it: ResumeInternship, i: number) => ({
      id: i + 1, company: it.company || "", position: it.position || "", startDate: it.start_date || "", endDate: it.end_date || "", content: it.description || "",
    }));
    setInternships(intList);
    const projList: Project[] = (formVm.projects || []).map((p: ResumeProject, i: number) => ({
      id: i + 1, name: p.name || "", desc: p.description || "", role: p.role || "", startDate: p.start_date || "", endDate: p.end_date || "",
    }));
    setProjects(projList);
    setSkillsDraft({ developmentLanguages: formVm.professionalSkills || "", aiApplication: formVm.aiApplication || "" });
    setLanguageExamDrafts((formVm.languageExams || []).map((le: ResumeLanguageExam, i: number) => ({ id: i + 1, name: le.exam_type || "", score: le.score || "" })));
    setCertificateDrafts((formVm.certificates || []).map((c: ResumeCertificate, i: number) => ({ id: i + 1, name: c.name || "", description: c.description || "" })));
    setAwardDrafts((formVm.awards || []).map((a: ResumeAward, i: number) => ({ id: i + 1, name: a.name || "", description: a.description || "" })));
    const worksList: WorkDraft[] = (formVm.works || []).map((w: ResumeWork, i: number) => ({ id: i + 1, name: w.name || "", link: w.url || "", password: w.password || "" }));
    setPortfolioDraft({ works: worksList.length > 0 ? worksList : [{ id: 1, name: "", link: "", password: "" }], homepage: formVm.homepage || "" });
    setExtraInfoDraft({ additionalInfo: formVm.additionalInfo || "" });
  }, [formVm, mode]);

  async function handleSave() {
    const trimmedName = resumeName.trim();
    const trimmedRealName = personalName.trim();

    // ── Required field checks ──────────────────────────────────────────────
    if (!trimmedName && !trimmedRealName) {
      setFormAlert({ open: true, title: "请填写必填信息", description: "请先填写简历名称和姓名，填写后才能保存。", variant: "danger" });
      return;
    }
    if (!trimmedName) {
      setFormAlert({ open: true, title: "请填写简历名称", description: "简历名称用于区分不同简历，填写后才能保存。", variant: "danger" });
      return;
    }
    if (!trimmedRealName) {
      setFormAlert({ open: true, title: "请填写姓名", description: "姓名是基础信息中的必填项，填写后才能保存简历。", variant: "danger" });
      return;
    }

    // ── Name check status gates ───────────────────────────────────────────
    if (nameCheckStatus === "checking") {
      setFormAlert({ open: true, title: "正在检查简历名称", description: "请等待名称校验完成后再保存。", variant: "default" });
      return;
    }
    if (nameCheckStatus === "duplicate") {
      setFormAlert({ open: true, title: "简历名称已存在", description: "当前账号下已经有同名简历，请换一个名称后再保存。", variant: "danger" });
      return;
    }
    if (nameCheckStatus === "invalid") {
      setFormAlert({ open: true, title: "简历名称无效", description: "简历名称不能为空，请填写一个有效的名称。", variant: "danger" });
      return;
    }
    if (nameCheckStatus === "error") {
      setFormAlert({ open: true, title: "暂时无法校验名称", description: "请稍后重试，确认简历名称可用后再保存。", variant: "default" });
      return;
    }

    // ── Pre-save name re-check if status is idle ───────────────────────────
    if (nameCheckStatus !== "available") {
      setNameCheckStatus("checking");
      setNameCheckMessage("正在检查...");
      try {
        const result = await checkResumeName(trimmedName, editingResumeId ?? undefined);
        if (!result.available) {
          setNameCheckStatus("duplicate");
          setNameCheckMessage(result.reason || "该简历名称已存在");
          setFormAlert({ open: true, title: "简历名称已存在", description: "当前账号下已经有同名简历，请换一个名称后再保存。", variant: "danger" });
          return;
        }
        setNameCheckStatus("available");
        setNameCheckMessage("名称可用，保存后将用于岗位推荐、匹配分析和简历优化建议。");
      } catch {
        setNameCheckStatus("error");
        setNameCheckMessage("暂时无法校验名称，请稍后重试。");
        setFormAlert({ open: true, title: "暂时无法校验名称", description: "请稍后重试，确认简历名称可用后再保存。", variant: "default" });
        return;
      }
    }

    // ── Proceed with save ──────────────────────────────────────────────────
    setSaveSuccessMsg(""); setSaveErrorMsg(""); setSaving(true);
    // Filter out items currently in exit animation
    const visibleEducations = educations.filter((e) => !removingItemKeys.has(`education:${e.id}`));
    const visibleProjects = projects.filter((p) => !removingItemKeys.has(`project:${p.id}`));
    const visibleInternships = internships.filter((i) => !removingItemKeys.has(`internship:${i.id}`));
    const visibleCerts = certificateDrafts.filter((c) => !removingItemKeys.has(`certificate:${c.id}`));
    const visibleAwards = awardDrafts.filter((a) => !removingItemKeys.has(`award:${a.id}`));
    const visibleWorks = portfolioDraft.works.filter((w) => !removingItemKeys.has(`work:${w.id}`));
    const visibleLanguageExams = languageExamDrafts.filter((le) => !removingItemKeys.has(`language:${le.id}`));
    const payload: SaveResumePayload = {
      name: trimmedName, is_current: true,
      real_name: personalName || "", phone: phone || "", email: email || "", city: currentCity || "",
      professional_skills: skillsDraft.developmentLanguages || "", ai_application: skillsDraft.aiApplication || "",
      homepage: portfolioDraft.homepage || "", additional_info: extraInfoDraft.additionalInfo || "",
      educations: visibleEducations.map((e) => ({ school: e.school || "", degree: e.degree || "", major: e.major || "", department: e.department || "", start_date: e.startDate || "", end_date: e.endDate || "", gpa: e.gpa || "", gpa_scale: e.gpaScale || "", rank_text: e.rank || "", core_courses: e.coreCourses || "", research_direction: e.researchDirection || "", advisor: e.advisor || "", lab: e.lab || "", paper_summary: e.paperSummary || "" })),
      projects: visibleProjects.map((p) => ({ name: p.name || "", role: p.role || "", start_date: p.startDate || "", end_date: p.endDate || "", description: p.desc || "" })),
      internships: visibleInternships.map((i) => ({ company: i.company || "", position: i.position || "", start_date: i.startDate || "", end_date: i.endDate || "", description: i.content || "" })),
      language_exams: visibleLanguageExams.filter((le) => le.name.trim()).map((le) => ({ exam_type: le.name.trim(), score: le.score.trim() })),
      certificates: visibleCerts.filter((c) => c.name.trim()).map((c) => ({ name: c.name.trim(), description: c.description || "" })),
      awards: visibleAwards.filter((a) => a.name.trim()).map((a) => ({ name: a.name.trim(), description: a.description || "" })),
      works: visibleWorks.filter((w) => w.name || w.link).map((w) => ({ name: w.name || "", url: w.link || "", password: w.password || "" })),
    };
    try {
      if (mode === "add" || !editingResumeId) { await createResume(payload); } else { await updateResume(editingResumeId, payload); }
      await Promise.allSettled([getCurrentResume(), getResumes()]);
      setSaveSuccessMsg(mode === "add" ? "简历已创建成功。" : "修改已保存。");
      setTimeout(() => navigate("/resumes", { replace: true }), 800);
    } catch (err: unknown) {
      if (isApiClientError(err)) {
        if (err.status === 401) {
          setSaveErrorMsg("登录状态已失效，请重新登录后再保存。");
        } else if (err.status === 409) {
          setNameCheckStatus("duplicate");
          setNameCheckMessage("该简历名称已存在");
          setFormAlert({ open: true, title: "简历名称已存在", description: "当前账号下已经有同名简历，请换一个名称后再保存。", variant: "danger" });
        } else {
          setSaveErrorMsg("保存失败，请稍后重试。");
        }
      } else {
        setSaveErrorMsg("保存失败，请稍后重试。");
      }
    } finally { setSaving(false); }
  }

  const [activeSection, setActiveSection] = useState("resume-basic");
  useEffect(() => {
    const handler = () => { for (const s of [...NAV_SECTIONS].reverse()) { const el = document.getElementById(s.id); if (el && el.getBoundingClientRect().top <= 100) { setActiveSection(s.id); return; } } setActiveSection("resume-basic"); };
    window.addEventListener("scroll", handler, { passive: true }); return () => window.removeEventListener("scroll", handler);
  }, []);

  function updateProject(id: number, field: keyof Project, value: string) { setProjects((ps) => ps.map((p) => (p.id === id ? { ...p, [field]: value } : p))); }
  function moveProject(index: number, dir: -1 | 1) { setProjects((ps) => { const next = [...ps]; const target = index + dir; if (target < 0 || target >= next.length) return ps; [next[index], next[target]] = [next[target], next[index]]; return next; }); }
  function removeProject(id: number) { setProjects((ps) => ps.filter((p) => p.id !== id)); }
  function addProject() { const newId = Math.max(0, ...projects.map((p) => p.id)) + 1; setRecentlyAdded({ type: "project", key: String(newId) }); setProjects((ps) => [...ps, { id: newId, name: "", desc: "", role: "", startDate: "", endDate: "" }]); }
  function updateSkillsDraft(field: keyof SkillsDraft, value: string) { setSkillsDraft((prev) => ({ ...prev, [field]: value })); }
  function addWork() { const newId = Math.max(0, ...portfolioDraft.works.map((w) => w.id)) + 1; setRecentlyAdded({ type: "work", key: String(newId) }); setPortfolioDraft((prev) => ({ ...prev, works: [...prev.works, { id: newId, name: "", link: "", password: "" }] })); }
  function updateWork(id: number, field: keyof WorkDraft, value: string) { setPortfolioDraft((prev) => ({ ...prev, works: prev.works.map((w) => (w.id === id ? { ...w, [field]: value } : w)) })); }
  function removeWork(id: number) { setPortfolioDraft((prev) => ({ ...prev, works: prev.works.filter((w) => w.id !== id) })); }
  function addCertificate() { const newId = Math.max(0, ...certificateDrafts.map((c) => c.id)) + 1; setRecentlyAdded({ type: "certificate", key: String(newId) }); setCertificateDrafts((cs) => [...cs, { id: newId, name: "", description: "" }]); }
  function updateCertificate(id: number, field: keyof CertificateDraft, value: string) { setCertificateDrafts((cs) => cs.map((c) => (c.id === id ? { ...c, [field]: value } : c))); }
  function removeCertificate(id: number) { setCertificateDrafts((cs) => cs.filter((c) => c.id !== id)); }
  const addLanguageExam = () => { const newId = Math.max(0, ...languageExamDrafts.map((le) => le.id)) + 1; setRecentlyAdded({ type: "languageExam", key: String(newId) }); setLanguageExamDrafts((prev) => [...prev, { id: newId, name: "", score: "" }]); };
  const updateLanguageExam = (id: number, field: keyof LanguageExamDraft, value: string) => { setLanguageExamDrafts((prev) => prev.map((le) => (le.id === id ? { ...le, [field]: value } : le))); };
  const removeLanguageExam = (id: number) => setLanguageExamDrafts((prev) => prev.filter((le) => le.id !== id));
  function addAward() { const newId = Math.max(0, ...awardDrafts.map((a) => a.id)) + 1; setRecentlyAdded({ type: "award", key: String(newId) }); setAwardDrafts((as) => [...as, { id: newId, name: "", description: "" }]); }
  function updateAward(id: number, field: keyof AwardDraft, value: string) { setAwardDrafts((as) => as.map((a) => (a.id === id ? { ...a, [field]: value } : a))); }
  function removeAward(id: number) { setAwardDrafts((as) => as.filter((a) => a.id !== id)); }
  function addEducation() { const newId = Math.max(0, ...educations.map((e) => e.id)) + 1; setRecentlyAdded({ type: "education", key: String(newId) }); setEducations((es) => [...es, { id: newId, school: "", major: "", degree: "本科", startDate: "", endDate: "", department: "", gpa: "", gpaScale: "", rank: "", coreCourses: "", researchDirection: "", advisor: "", lab: "", paperSummary: "" }]); }
  function updateEdu(id: number, field: keyof Education, value: string) { setEducations((es) => es.map((e) => (e.id === id ? { ...e, [field]: value } : e))); }
  function addInternship() { const newId = Math.max(0, ...internships.map((i) => i.id)) + 1; setRecentlyAdded({ type: "internship", key: String(newId) }); setInternships((is) => [...is, { id: newId, company: "", position: "", startDate: "", endDate: "", content: "" }]); }
  function updateIntern(id: number, field: keyof Internship, value: string) { setInternships((is) => is.map((i) => (i.id === id ? { ...i, [field]: value } : i))); }

  // ── Readiness computation (must be before any early returns to keep hook count stable) ──
  const readiness = useMemo(() => {
    const weights = { resumeName: 5, realName: 10, contact: 10, education: 15, skills: 15, practice: 20, direction: 10, aiSkill: 5, certAndLink: 10 };
    let score = 0;
    if (resumeName.trim()) score += weights.resumeName;
    if (personalName.trim()) score += weights.realName;
    if (phone.trim() || email.trim()) score += weights.contact;
    if (educations.some((e) => e.school.trim() && e.major.trim())) score += weights.education;
    if (skillsDraft.developmentLanguages.trim()) score += weights.skills;
    if (projects.some((p) => p.name.trim()) || internships.some((i) => i.company.trim() || i.position.trim())) score += weights.practice;
    if (extraInfoDraft.additionalInfo.trim() || currentCity.trim() || skillsDraft.developmentLanguages.trim()) score += weights.direction;
    if (skillsDraft.aiApplication.trim()) score += weights.aiSkill;
    if (languageExamDrafts.some((le) => le.name.trim()) || certificateDrafts.some((c) => c.name.trim()) || awardDrafts.some((a) => a.name.trim()) || portfolioDraft.works.some((w) => w.name.trim() || w.link.trim()) || !!portfolioDraft.homepage.trim()) score += weights.certAndLink;
    return score;
  }, [resumeName, personalName, phone, email, educations, skillsDraft, projects, internships, extraInfoDraft, currentCity, languageExamDrafts, certificateDrafts, awardDrafts, portfolioDraft]);

  const readinessCategories: ReadinessCategory[] = useMemo(() => {
    const hasEdu = educations.some((e) => e.school.trim() && e.major.trim());
    const hasPractice = projects.some((p) => p.name.trim()) || internships.some((i) => i.company.trim() || i.position.trim());
    const hasCert = languageExamDrafts.some((le) => le.name.trim()) || certificateDrafts.some((c) => c.name.trim()) || awardDrafts.some((a) => a.name.trim()) || portfolioDraft.works.some((w) => w.name.trim() || w.link.trim()) || !!portfolioDraft.homepage.trim();
    return [
      { label: "基础信息", status: (personalName.trim() ? "complete" : "missing") as ReadinessCategory["status"] },
      { label: "教育经历", status: (hasEdu ? "complete" : "missing") as ReadinessCategory["status"] },
      { label: "技能能力", status: (skillsDraft.developmentLanguages.trim() ? "complete" : "missing") as ReadinessCategory["status"] },
      { label: "实践经历", status: (hasPractice ? "complete" : "missing") as ReadinessCategory["status"] },
      { label: "AI 应用", status: (skillsDraft.aiApplication.trim() ? "complete" : "missing") as ReadinessCategory["status"] },
      { label: "证书获奖", status: (hasCert ? "complete" : (certificateDrafts.length > 0 || awardDrafts.length > 0 ? "partial" : "missing")) as ReadinessCategory["status"] },
    ];
  }, [personalName, educations, skillsDraft, projects, internships, languageExamDrafts, certificateDrafts, awardDrafts, portfolioDraft]);

  const readinessNextStep = useMemo(() => {
    if (readiness >= 100) return "";
    if (!personalName.trim()) return "先填写姓名，这是简历的基础身份信息。";
    if (!educations.some((e) => e.school.trim() && e.major.trim())) return "补充教育经历，学校、专业和学历是匹配分析的重要依据。";
    if (!skillsDraft.developmentLanguages.trim()) return "填写专业技能或工具能力，帮助岗位匹配时精准定位。";
    if (!projects.some((p) => p.name.trim()) && !internships.some((i) => i.company.trim())) return "补充一段项目或实习经历，会显著提升岗位推荐和匹配分析质量。";
    return "继续补充更多信息，让投递准备度更完善。";
  }, [readiness, personalName, educations, skillsDraft, projects, internships]);

  if (mode === "edit" && loading) return <ResumeSkeleton />;
  if (mode === "edit" && resumeError) return (<PageError title="简历加载失败" description={resumeError} onRetry={handleRetryLoad} />);

  const isEditingExistingResume = Boolean(editingResumeId);
  const hasInternshipContent = internships.length > 0;
  const hasCertificateAwardContent = languageExamDrafts.length > 0 || certificateDrafts.length > 0 || awardDrafts.length > 0;
  const hasPortfolioExtraContent = portfolioDraft.works.some((w) => w.name.trim() || w.link.trim()) || !!portfolioDraft.homepage.trim() || !!extraInfoDraft.additionalInfo.trim();
  // K8-5-49-E-B: 有内容才默认展开，避免空模块全部展开
  const hasEducationContent = educations.some((edu) =>
    edu.school || edu.major || edu.degree !== "本科" || edu.department || edu.startDate || edu.endDate || edu.gpa || edu.gpaScale || edu.rank || edu.coreCourses || edu.researchDirection || edu.advisor || edu.lab || edu.paperSummary
  );
  const hasProjectContent = projects.some((proj) =>
    proj.name || proj.role || proj.desc || proj.startDate || proj.endDate
  );
  const hasSkillsContent = !!(skillsDraft.developmentLanguages.trim() || skillsDraft.aiApplication.trim());
  const sectionAutoOpenKey = isCreateMode ? "create" : `edit-${editingResumeId}:${resumeName || "_empty"}`;

  return (
    <>
    <ResumeEditorShell title={mode === "add" ? "创建简历" : "编辑简历"} subtitle="完善你的经历和技能，让 Offer捕手 更准确地完成岗位推荐、匹配分析和简历优化。" saveCard={<ResumeEditorSaveCard mode={mode} saving={saving} successMessage={saveSuccessMsg} errorMessage={saveErrorMsg} onSave={handleSave} disabled={saveDisabled} />} sidebar={<ResumeEditorStatusSidebar readiness={readiness} categories={readinessCategories} nextStep={readinessNextStep} saving={saving} />}>
      <Link to="/resumes" className="inline-flex items-center gap-1 text-xs text-gray-400 hover:text-gray-600 transition-colors -mt-2 mb-1"><ChevronLeft className="w-3 h-3" />返回我的简历</Link>
      <div className="flex flex-wrap items-center gap-2 px-4 py-2.5 rounded-xl bg-blue-50 border border-blue-100">
        <Sparkles className="w-3.5 h-3.5 text-[#0052D9] flex-shrink-0" />
        <span className="text-xs text-[#0052D9]">当前简历可用于：<strong>岗位推荐</strong> / <strong>岗位匹配分析</strong> / <strong>简历优化</strong></span>
        <span className="w-px h-3.5 bg-blue-200 mx-1 hidden sm:block" /><span className="text-xs text-gray-400 hidden sm:inline">建议仅作投递前参考，不替代真实经历核实。</span>
        {isCreateMode && (
          <>
            <span className="w-px h-3.5 bg-blue-200 mx-1" />
            <button
              type="button"
              onClick={handleFillDemo}
              disabled={fillingDemo}
              className="inline-flex items-center gap-1.5 rounded-lg bg-[#0052D9] px-3 py-1 text-xs font-medium text-white hover:bg-[#0047BD] transition-colors disabled:opacity-50"
            >
              {fillingDemo ? <Loader2 className="w-3 h-3 animate-spin" /> : <Sparkles className="w-3 h-3" />}
              {fillingDemo ? "填入中…" : "填入体验内容"}
            </button>
          </>
        )}
      </div>
      {demoError && (
        <div className="rounded-xl bg-red-50 border border-red-100 px-4 py-2.5 text-xs text-red-600 flex items-center gap-2">
          <AlertCircle className="w-3.5 h-3.5 flex-shrink-0" />
          {demoError}
        </div>
      )}

      <ResumeBasicsSection>
        <div className="space-y-5">
          <div><label className="mb-1.5 block text-sm font-medium text-slate-700">简历名称 <span className="text-red-500">*</span></label><Input value={resumeName} onChange={(e) => { setResumeName(e.target.value); setNameCheckStatus("idle"); setNameCheckMessage(""); }} onBlur={handleNameBlur} placeholder="用于区分不同简历版本" /></div>
          <div
            className={cn(
              "flex items-center gap-2 rounded-xl px-4 py-3 border text-[13px] leading-relaxed min-h-[48px]",
              nameCheckStatus === "available" && "bg-green-50 border-green-100 text-green-700",
              (nameCheckStatus === "duplicate" || nameCheckStatus === "invalid") && "bg-red-50 border-red-200 text-red-600",
              nameCheckStatus === "error" && "bg-red-50 border-red-100 text-red-500",
              nameCheckStatus === "checking" && "bg-blue-50 border-blue-100 text-blue-600",
              nameCheckStatus === "idle" && "bg-blue-50/60 border-blue-100 text-blue-600",
            )}
          >
            {nameCheckStatus === "idle" && <HelpCircle className="w-4 h-4 flex-shrink-0" />}
            {nameCheckStatus === "checking" && <Loader2 className="w-4 h-4 flex-shrink-0 animate-spin" />}
            {nameCheckStatus === "available" && <CheckCircle2 className="w-4 h-4 flex-shrink-0" />}
            {nameCheckStatus === "duplicate" && <XCircle className="w-4 h-4 flex-shrink-0" />}
            {nameCheckStatus === "invalid" && <AlertCircle className="w-4 h-4 flex-shrink-0" />}
            {nameCheckStatus === "error" && <AlertCircle className="w-4 h-4 flex-shrink-0" />}
            <p>
              {nameCheckStatus === "idle" && "输入简历名称后自动检查是否可用"}
              {nameCheckStatus === "checking" && "正在检查名称是否可用…"}
              {nameCheckStatus === "available" && (isEditingExistingResume ? "当前名称可继续使用" : "名称可用")}
              {nameCheckStatus === "duplicate" && "名称已存在，请换一个名称"}
              {nameCheckStatus === "invalid" && "请输入简历名称"}
              {nameCheckStatus === "error" && "检查失败，请稍后重试"}
            </p>
          </div>
        </div>
      </ResumeBasicsSection>

      <BasicInfoSection>
        <div className="grid grid-cols-1 gap-x-6 gap-y-5 sm:grid-cols-2">
          <div><label className="mb-1.5 block text-sm font-medium text-slate-700">姓名 <span className="text-red-500">*</span></label><Input value={personalName} onChange={(e) => setPersonalName(e.target.value)} placeholder="姓名" /></div>
          <div><label className="mb-1.5 block text-sm font-medium text-slate-700">当前城市</label><Input value={currentCity} onChange={(e) => setCurrentCity(e.target.value)} placeholder="如：深圳" /></div>
        </div>
        <div className="mt-5 pt-5 border-t border-gray-100 grid grid-cols-1 gap-x-6 gap-y-5 sm:grid-cols-2">
          <div><label className="mb-1.5 block text-sm font-medium text-slate-700">手机号</label><Input value={phone} onChange={(e) => setPhone(e.target.value)} placeholder="手机号" /></div>
          <div><label className="mb-1.5 block text-sm font-medium text-slate-700">邮箱</label><Input value={email} onChange={(e) => setEmail(e.target.value)} placeholder="邮箱地址" /></div>
        </div>
        <p className="mt-5 text-[12px] text-slate-400 leading-relaxed border-t border-gray-100 pt-4">基础信息用于简历展示和简历概览。手机号 / 邮箱不会用于分析匹配。</p>
      </BasicInfoSection>

      <EducationSection defaultOpen={isEditingExistingResume && hasEducationContent} onAdd={addEducation}>
        {(() => { const eduActive = educations.some((e) => !removingItemKeys.has(`education:${e.id}`)); const eduRemoving = educations.some((e) => removingItemKeys.has(`education:${e.id}`)); if (educations.length === 0 || (!eduActive && eduRemoving)) return <ResumeEditorEmptyState title="暂无教育经历" description="可以先添加你的学校、专业和学历信息。" actionLabel="+ 添加教育经历" onAction={addEducation} />;
          return (
          <div className="space-y-4">{educations.map((edu, idx) => {
            const eduKey = `education:${edu.id}`;
            const isRemoving = removingItemKeys.has(eduKey);
            return (
            <AnimatedListItem key={edu.id} removing={isRemoving}>
              <div className={`rounded-2xl bg-gray-50/70 ring-1 ring-gray-100 p-5 space-y-4 ${recentlyAdded?.type === "education" && recentlyAdded.key === String(edu.id) ? "resume-item-enter" : ""}`}>
              <div className="flex items-center justify-between"><span className="text-xs font-medium text-slate-500">教育经历 {idx + 1}</span><button onClick={() => markRemoving(eduKey, () => setEducations((es) => es.filter((e) => e.id !== edu.id)))} disabled={isRemoving} className={cn("text-slate-400 hover:text-red-500 transition-colors", isRemoving && "opacity-50 pointer-events-none")}><Trash2 className="w-4 h-4" /></button></div>
              <div className="grid grid-cols-1 gap-x-5 gap-y-4 sm:grid-cols-2">
                <div><label className="mb-1.5 block text-sm font-medium text-slate-700">学校 <span className="text-red-500">*</span></label><Input className="h-10" value={edu.school} onChange={(e) => updateEdu(edu.id, "school", e.target.value)} placeholder="学校名称" /></div>
                <div><label className="mb-1.5 block text-sm font-medium text-slate-700">专业 <span className="text-red-500">*</span></label><Input className="h-10" value={edu.major} onChange={(e) => updateEdu(edu.id, "major", e.target.value)} placeholder="专业名称" /></div>
                <div><label className="mb-1.5 block text-sm font-medium text-slate-700">学历</label><Select value={edu.degree} onValueChange={(v) => updateEdu(edu.id, "degree", v)}><SelectTrigger className="h-10"><SelectValue /></SelectTrigger><SelectContent><SelectItem value="本科">本科</SelectItem><SelectItem value="硕士">硕士</SelectItem><SelectItem value="博士">博士</SelectItem><SelectItem value="专科">专科</SelectItem></SelectContent></Select></div>
                <div><label className="mb-1.5 block text-sm font-medium text-slate-700">院系</label><Input className="h-10" value={edu.department} onChange={(e) => updateEdu(edu.id, "department", e.target.value)} placeholder="如：计算机科学与技术学院" /></div>
                <div><label className="mb-1.5 block text-sm font-medium text-slate-700">时间</label><div className="flex items-center gap-2"><YearMonthPicker value={edu.startDate} onChange={(v) => updateEdu(edu.id, "startDate", v)} placeholder="开始时间" /><span className="text-slate-400">—</span><YearMonthPicker value={edu.endDate} onChange={(v) => updateEdu(edu.id, "endDate", v)} placeholder="结束时间" /></div></div>
              </div>
              <div className="border-t border-gray-200 pt-4">
                <div className="grid grid-cols-1 gap-x-5 gap-y-4 sm:grid-cols-3">
                  <div><label className="mb-1.5 block text-sm font-medium text-slate-700">GPA</label><Input className="h-10" value={edu.gpa} onChange={(e) => updateEdu(edu.id, "gpa", e.target.value)} placeholder="如：3.8" /></div>
                  <div><label className="mb-1.5 block text-sm font-medium text-slate-700">满绩点</label><Select value={edu.gpaScale || undefined} onValueChange={(v) => updateEdu(edu.id, "gpaScale", v)}><SelectTrigger className="h-10"><SelectValue placeholder="选择" /></SelectTrigger><SelectContent><SelectItem value="4.0">4.0</SelectItem><SelectItem value="5.0">5.0</SelectItem><SelectItem value="100">100</SelectItem></SelectContent></Select></div>
                  <div><label className="mb-1.5 block text-sm font-medium text-slate-700">成绩排名</label><Input className="h-10" value={edu.rank} onChange={(e) => updateEdu(edu.id, "rank", e.target.value)} placeholder="如：前 20%" /></div>
                </div>
                <div className="mt-4"><label className="mb-1.5 block text-sm font-medium text-slate-700">主修课程</label><Input className="h-10" value={edu.coreCourses} onChange={(e) => updateEdu(edu.id, "coreCourses", e.target.value)} placeholder="如：数据结构、操作系统、计算机网络" /></div>
              </div>
              <div className="border-t border-gray-200 pt-4">
                <div className="grid grid-cols-1 gap-x-5 gap-y-4 sm:grid-cols-2">
                  <div><label className="mb-1.5 block text-sm font-medium text-slate-700">研究方向</label><Input className="h-10" value={edu.researchDirection} onChange={(e) => updateEdu(edu.id, "researchDirection", e.target.value)} placeholder="如：分布式系统" /></div>
                  <div><label className="mb-1.5 block text-sm font-medium text-slate-700">导师</label><Input className="h-10" value={edu.advisor} onChange={(e) => updateEdu(edu.id, "advisor", e.target.value)} placeholder="如：张教授" /></div>
                  <div><label className="mb-1.5 block text-sm font-medium text-slate-700">实验室</label><Input className="h-10" value={edu.lab} onChange={(e) => updateEdu(edu.id, "lab", e.target.value)} placeholder="如：智能信息处理实验室" /></div>
                </div>
                <div className="mt-4"><label className="mb-1.5 block text-sm font-medium text-slate-700">论文 / 学术成果</label><AutoGrowTextarea minRows={2} maxRows={10} value={edu.paperSummary} onChange={(e) => updateEdu(edu.id, "paperSummary", e.target.value)} placeholder="简要描述发表的论文或参与的科研项目" /></div>
              </div>
            </div>
            </AnimatedListItem>
          );})}
            {educations.some((e) => !removingItemKeys.has(`education:${e.id}`)) && <button type="button" onClick={addEducation} className="mt-3 text-[13px] font-medium text-[#0052D9] hover:underline">+ 添加教育经历</button>}
          </div>); })()}
      </EducationSection>

      <ProjectSection defaultOpen={isEditingExistingResume && hasProjectContent}>
        {(() => { const projActive = projects.some((p) => !removingItemKeys.has(`project:${p.id}`)); const projRemoving = projects.some((p) => removingItemKeys.has(`project:${p.id}`)); if (projects.length === 0 || (!projActive && projRemoving)) return <ResumeEditorEmptyState title="暂无项目经历" description="可以补充课程项目、比赛项目或个人项目。" actionLabel="+ 添加项目经历" onAction={addProject} />;
          return (
          <div className="space-y-4">{projects.map((proj, idx) => {
            const projKey = `project:${proj.id}`;
            const isRemoving = removingItemKeys.has(projKey);
            return (
            <AnimatedListItem key={proj.id} removing={isRemoving}>
              <div className={`rounded-2xl bg-gray-50/70 ring-1 ring-gray-100 p-5 space-y-4 ${recentlyAdded?.type === "project" && recentlyAdded.key === String(proj.id) ? "resume-item-enter" : ""}`}>
              <div className="flex items-center justify-between"><span className="text-xs font-medium text-slate-500">项目经历 {idx + 1}</span><div className="flex items-center gap-1"><button onClick={() => moveProject(idx, -1)} disabled={idx === 0 || isRemoving} className="p-1 text-slate-400 hover:text-slate-600 disabled:opacity-30" title="上移"><ChevronUp className="w-3.5 h-3.5" /></button><button onClick={() => moveProject(idx, 1)} disabled={idx === projects.length - 1 || isRemoving} className="p-1 text-slate-400 hover:text-slate-600 disabled:opacity-30" title="下移"><ChevronDown className="w-3.5 h-3.5" /></button><button onClick={() => markRemoving(projKey, () => removeProject(proj.id))} disabled={isRemoving} className={cn("p-1 text-slate-400 hover:text-red-500", isRemoving && "opacity-50 pointer-events-none")} title="删除"><Trash2 className="w-3.5 h-3.5" /></button></div></div>
              <div><label className="mb-1.5 block text-sm font-medium text-slate-700">项目名称 <span className="text-red-500">*</span></label><Input className="h-10" value={proj.name} onChange={(e) => updateProject(proj.id, "name", e.target.value)} placeholder="项目名称" /></div>
              <div><label className="mb-1.5 block text-sm font-medium text-slate-700">项目时间</label><div className="flex items-center gap-2"><YearMonthPicker value={proj.startDate} onChange={(v) => updateProject(proj.id, "startDate", v)} placeholder="开始时间" /><span className="text-slate-400">—</span><YearMonthPicker value={proj.endDate} onChange={(v) => updateProject(proj.id, "endDate", v)} placeholder="结束时间" /></div></div>
              <div><label className="mb-1.5 block text-sm font-medium text-slate-700">在项目中担任的角色</label><Input className="h-10" value={proj.role} onChange={(e) => updateProject(proj.id, "role", e.target.value)} placeholder="如：接口设计、数据库设计" /></div>
              <div><label className="mb-1.5 block text-sm font-medium text-slate-700">描述</label><AutoGrowTextarea minRows={4} maxRows={12} value={proj.desc} onChange={(e) => updateProject(proj.id, "desc", e.target.value)} placeholder="描述项目背景、你负责的部分、使用的技术栈和主要成果" /></div>
            </div>
            </AnimatedListItem>
          );})}
            {projects.some((p) => !removingItemKeys.has(`project:${p.id}`)) && <button type="button" onClick={addProject} className="mt-3 text-[13px] font-medium text-[#0052D9] hover:underline">+ 添加项目经历</button>}
          </div>); })()}
      </ProjectSection>

      <SkillsSection defaultOpen={isEditingExistingResume && hasSkillsContent}>
        <div className="border-t border-gray-200 pt-5 mt-5 space-y-4">
          <div><label className="mb-1.5 block text-sm font-medium text-slate-700">专业技能 / 工具能力</label><Input className="h-10" value={skillsDraft.developmentLanguages} onChange={(e) => updateSkillsDraft("developmentLanguages", e.target.value)} placeholder="技术岗可填 Go / Java / SQL" /><p className="mt-1.5 text-xs text-slate-400">可填写开发语言、设计工具、数据分析方法、办公软件等。</p></div>
          <div><label className="mb-1.5 block text-sm font-medium text-slate-700">AI 应用技能 <span className="text-xs font-normal text-amber-600">推荐</span></label><AutoGrowTextarea minRows={3} maxRows={8} value={skillsDraft.aiApplication} onChange={(e) => updateSkillsDraft("aiApplication", e.target.value)} placeholder="例如：使用 CodeBuddy 辅助需求拆解、代码生成、Bug 定位、接口文档、测试用例和项目复盘。" /></div>
        </div>
        <p className="mt-5 text-[12px] text-slate-400 border-t border-gray-100 pt-4">技能信息有助于 Offer捕手 提升岗位推荐、匹配分析和简历优化质量。</p>
      </SkillsSection>

      <InternshipSection defaultOpen={isEditingExistingResume && hasInternshipContent} autoOpenKey={sectionAutoOpenKey}>
        {(() => { const intActive = internships.some((i) => !removingItemKeys.has(`internship:${i.id}`)); const intRemoving = internships.some((i) => removingItemKeys.has(`internship:${i.id}`)); if (internships.length === 0 || (!intActive && intRemoving)) return <ResumeEditorEmptyState title="暂无实习经历" description="可以补充实习、兼职或校内实践经历。" actionLabel="+ 添加实习经历" onAction={addInternship} />;
          return (
          <div className="space-y-4">{internships.map((intern, idx) => {
            const intKey = `internship:${intern.id}`;
            const isRemoving = removingItemKeys.has(intKey);
            return (
            <AnimatedListItem key={intern.id} removing={isRemoving}>
              <div className={`rounded-2xl bg-gray-50/70 ring-1 ring-gray-100 p-5 space-y-4 ${recentlyAdded?.type === "internship" && recentlyAdded.key === String(intern.id) ? "resume-item-enter" : ""}`}>
              <div className="flex items-center justify-between"><span className="text-xs font-medium text-slate-500">实习经历 {idx + 1}</span><button onClick={() => markRemoving(intKey, () => setInternships((is) => is.filter((i) => i.id !== intern.id)))} disabled={isRemoving} className={cn("text-slate-400 hover:text-red-500", isRemoving && "opacity-50 pointer-events-none")}><Trash2 className="w-4 h-4" /></button></div>
              <div className="grid grid-cols-1 gap-x-5 gap-y-4 sm:grid-cols-2"><div><label className="mb-1.5 block text-sm font-medium text-slate-700">公司名称 <span className="text-red-500">*</span></label><Input className="h-10" value={intern.company} onChange={(e) => updateIntern(intern.id, "company", e.target.value)} placeholder="公司名称" /></div><div><label className="mb-1.5 block text-sm font-medium text-slate-700">职位 <span className="text-red-500">*</span></label><Input className="h-10" value={intern.position} onChange={(e) => updateIntern(intern.id, "position", e.target.value)} placeholder="职位" /></div></div>
              <div><label className="mb-1.5 block text-sm font-medium text-slate-700">时间</label><div className="flex items-center gap-2"><YearMonthPicker value={intern.startDate} onChange={(v) => updateIntern(intern.id, "startDate", v)} placeholder="开始时间" /><span className="text-slate-400">—</span><YearMonthPicker value={intern.endDate} onChange={(v) => updateIntern(intern.id, "endDate", v)} placeholder="结束时间" /></div></div>
              <div><label className="mb-1.5 block text-sm font-medium text-slate-700">描述</label><AutoGrowTextarea minRows={4} maxRows={12} value={intern.content} onChange={(e) => updateIntern(intern.id, "content", e.target.value)} placeholder="描述工作内容、使用的技术栈和主要成果" /></div>
            </div>
            </AnimatedListItem>
          );})}
            {internships.some((i) => !removingItemKeys.has(`internship:${i.id}`)) && <button type="button" onClick={addInternship} className="mt-3 text-[13px] font-medium text-[#0052D9] hover:underline">+ 添加实习经历</button>}
          </div>); })()}
      </InternshipSection>

      <CertificatesSection onAdd={addCertificate} defaultOpen={isEditingExistingResume && hasCertificateAwardContent} autoOpenKey={sectionAutoOpenKey}>
        <div className="mb-5 pb-5 border-b border-gray-100">
          <h4 className="text-sm font-medium text-slate-600 mb-3">外语考试 / 等级</h4>
          {(() => { const langActive = languageExamDrafts.some((le) => !removingItemKeys.has(`language:${le.id}`)); const langRemoving = languageExamDrafts.some((le) => removingItemKeys.has(`language:${le.id}`)); if (languageExamDrafts.length === 0 || (!langActive && langRemoving)) return <ResumeEditorEmptyState title="暂无外语考试" description="可以补充四六级、雅思、托福等语言成绩。" actionLabel="+ 添加外语考试" onAction={addLanguageExam} />;
            return (<>
            {languageExamDrafts.map((le, idx) => {
              const langKey = `language:${le.id}`;
              const isRemoving = removingItemKeys.has(langKey);
              return (
              <AnimatedListItem key={le.id} removing={isRemoving}>
                <div className={`rounded-2xl bg-gray-50/70 ring-1 ring-gray-100 p-4 space-y-3 mb-3 ${recentlyAdded?.type === "languageExam" && recentlyAdded.key === String(le.id) ? "resume-item-enter" : ""}`}><div className="flex items-center justify-between"><span className="text-xs font-medium text-slate-500">外语考试 {idx + 1}</span><button onClick={() => markRemoving(langKey, () => removeLanguageExam(le.id))} disabled={isRemoving} className={cn("text-slate-400 hover:text-red-500", isRemoving && "opacity-50 pointer-events-none")}><Trash2 className="w-4 h-4" /></button></div><div className="grid grid-cols-1 gap-x-5 gap-y-4 sm:grid-cols-2"><div><label className="mb-1.5 block text-sm font-medium text-slate-700">考试类型</label><Input className="h-10" value={le.name} onChange={(e) => updateLanguageExam(le.id, "name", e.target.value)} placeholder="如：CET-4、CET-6、IELTS" /></div><div><label className="mb-1.5 block text-sm font-medium text-slate-700">分数 / 等级</label><Input className="h-10" value={le.score} onChange={(e) => updateLanguageExam(le.id, "score", e.target.value)} placeholder="例如：503 / 6.5" /></div></div></div>
              </AnimatedListItem>
            );})}
            <button onClick={addLanguageExam} className="text-[13px] font-medium text-[#0052D9] hover:underline">+ 添加外语考试</button>
          </>); })()}
        </div>
        {/* 证书 */}
        <div className="mb-5 pb-5 border-b border-gray-100">
          <h4 className="text-sm font-medium text-slate-600 mb-3">证书</h4>
          {(() => {
            const certActive = certificateDrafts.some((c) => !removingItemKeys.has(`certificate:${c.id}`));
            const certRemoving = certificateDrafts.some((c) => removingItemKeys.has(`certificate:${c.id}`));
            const certsEmpty = certificateDrafts.length === 0 || (!certActive && certRemoving);
            if (certsEmpty) {
              return <ResumeEditorEmptyState title="暂无证书" description="可以补充专业技术、语言、云计算等相关证书。" actionLabel="+ 添加证书" onAction={addCertificate} />;
            }
            return (<>
            {certificateDrafts.map((cert, idx) => {
              const certKey = `certificate:${cert.id}`;
              const isRemoving = removingItemKeys.has(certKey);
              return (
              <AnimatedListItem key={cert.id} removing={isRemoving}>
                <div className={`rounded-2xl bg-gray-50/70 ring-1 ring-gray-100 p-5 space-y-4 mb-4 ${recentlyAdded?.type === "certificate" && recentlyAdded.key === String(cert.id) ? "resume-item-enter" : ""}`}><div className="flex items-center justify-between"><span className="text-xs font-medium text-slate-500">证书 {idx + 1}</span><button onClick={() => markRemoving(certKey, () => removeCertificate(cert.id))} disabled={isRemoving} className={cn("text-slate-400 hover:text-red-500", isRemoving && "opacity-50 pointer-events-none")}><Trash2 className="w-4 h-4" /></button></div><div><label className="mb-1.5 block text-sm font-medium text-slate-700">证书名称 <span className="text-red-500">*</span></label><Input className="h-10" value={cert.name} onChange={(e) => updateCertificate(cert.id, "name", e.target.value)} placeholder="如：CET-4 503" /></div><div><label className="mb-1.5 block text-sm font-medium text-slate-700">证书说明</label><AutoGrowTextarea minRows={2} maxRows={6} value={cert.description} onChange={(e) => updateCertificate(cert.id, "description", e.target.value)} placeholder="补充信息" /></div></div>
              </AnimatedListItem>
            );})}
            {certActive && <button onClick={addCertificate} className="text-[13px] font-medium text-[#0052D9] hover:underline">+ 添加证书</button>}
          </>); })()}
        </div>

        {/* 获奖经历 */}
        <div>
          <h4 className="text-sm font-medium text-slate-600 mb-3">获奖经历</h4>
          {(() => {
            const awardActive = awardDrafts.some((a) => !removingItemKeys.has(`award:${a.id}`));
            const awardRemoving = awardDrafts.some((a) => removingItemKeys.has(`award:${a.id}`));
            const awardsEmpty = awardDrafts.length === 0 || (!awardActive && awardRemoving);
            if (awardsEmpty) {
              return <ResumeEditorEmptyState title="暂无获奖经历" description="可以补充竞赛获奖、学术荣誉、奖学金等。" actionLabel="+ 添加获奖" onAction={addAward} />;
            }
            return (<>
            {awardDrafts.map((award, idx) => {
              const awardKey = `award:${award.id}`;
              const isRemoving = removingItemKeys.has(awardKey);
              return (
              <AnimatedListItem key={award.id} removing={isRemoving}>
                <div className={`rounded-2xl bg-gray-50/70 ring-1 ring-gray-100 p-5 space-y-4 mb-4 ${recentlyAdded?.type === "award" && recentlyAdded.key === String(award.id) ? "resume-item-enter" : ""}`}><div className="flex items-center justify-between"><span className="text-xs font-medium text-slate-500">获奖 {idx + 1}</span><button onClick={() => markRemoving(awardKey, () => removeAward(award.id))} disabled={isRemoving} className={cn("text-slate-400 hover:text-red-500", isRemoving && "opacity-50 pointer-events-none")}><Trash2 className="w-4 h-4" /></button></div><div><label className="mb-1.5 block text-sm font-medium text-slate-700">奖项名称 <span className="text-red-500">*</span></label><Input className="h-10" value={award.name} onChange={(e) => updateAward(award.id, "name", e.target.value)} placeholder="如：国家奖学金" /></div><div><label className="mb-1.5 block text-sm font-medium text-slate-700">奖项说明</label><AutoGrowTextarea minRows={2} maxRows={6} value={award.description} onChange={(e) => updateAward(award.id, "description", e.target.value)} placeholder="获奖信息补充" /></div></div>
              </AnimatedListItem>
            );})}
            {awardActive && <button onClick={addAward} className="text-[13px] font-medium text-[#0052D9] hover:underline">+ 添加获奖</button>}
          </>); })()}
        </div>
      </CertificatesSection>

      <PortfolioExtraSection defaultOpen={isEditingExistingResume && hasPortfolioExtraContent} autoOpenKey={sectionAutoOpenKey}>
        <p className="text-[13px] text-slate-500 mb-5">只支持链接，不支持附件上传。</p>
        {(() => { const workActive = portfolioDraft.works.some((w) => !removingItemKeys.has(`work:${w.id}`)); const workRemoving = portfolioDraft.works.some((w) => removingItemKeys.has(`work:${w.id}`)); if (portfolioDraft.works.length === 0 || (!workActive && workRemoving)) return <ResumeEditorEmptyState title="暂无作品链接" description="可以补充 GitHub、博客、作品集或项目地址。" actionLabel="+ 添加作品链接" onAction={addWork} />;
          return (<>
        {portfolioDraft.works.map((work, idx) => {
          const workKey = `work:${work.id}`;
          const isRemoving = removingItemKeys.has(workKey);
          return (
          <AnimatedListItem key={work.id} removing={isRemoving}>
            <div className={`rounded-2xl bg-gray-50/70 ring-1 ring-gray-100 p-5 space-y-4 mb-4 ${recentlyAdded?.type === "work" && recentlyAdded.key === String(work.id) ? "resume-item-enter" : ""}`}><div className="flex items-center justify-between"><span className="text-xs font-medium text-slate-500">作品链接 {idx + 1}</span><button onClick={() => markRemoving(workKey, () => removeWork(work.id))} disabled={isRemoving} className={cn("text-slate-400 hover:text-red-500", isRemoving && "opacity-50 pointer-events-none")}><Trash2 className="w-4 h-4" /></button></div><div><label className="mb-1.5 block text-sm font-medium text-slate-700">作品名称</label><Input className="h-10" value={work.name} onChange={(e) => updateWork(work.id, "name", e.target.value)} placeholder="如：分布式在线考试系统" /></div><div><label className="mb-1.5 block text-sm font-medium text-slate-700">作品链接</label><Input className="h-10" value={work.link} onChange={(e) => updateWork(work.id, "link", e.target.value)} placeholder="如：https://cnb.cool/your-name/project" /></div><div><label className="mb-1.5 block text-sm font-medium text-slate-700">密码或提取码</label><Input className="h-10" value={work.password} onChange={(e) => updateWork(work.id, "password", e.target.value)} placeholder="如无需密码留空" /></div></div>
          </AnimatedListItem>
        );})}
        {portfolioDraft.works.some((w) => !removingItemKeys.has(`work:${w.id}`)) && <button onClick={addWork} className="text-[13px] font-medium text-[#0052D9] hover:underline mb-4 inline-block">+ 添加作品链接</button>}
        </>); })()}
        <div className="mt-4"><label className="mb-1.5 block text-sm font-medium text-slate-700">个人主页</label><Input className="h-10" value={portfolioDraft.homepage} onChange={(e) => setPortfolioDraft((prev) => ({ ...prev, homepage: e.target.value }))} placeholder="如：https://github.com/your-name/project" /></div>
      </PortfolioExtraSection>

      <PortfolioExtraSection title="补充信息" defaultOpen={isEditingExistingResume && hasPortfolioExtraContent} autoOpenKey={sectionAutoOpenKey}>
        <p className="text-[13px] text-slate-500 mb-5">补充其他未覆盖的信息，如技术背景、求职意向、社团经历、志愿服务等。</p>
        <AutoGrowTextarea minRows={4} maxRows={15} value={extraInfoDraft.additionalInfo} onChange={(e) => setExtraInfoDraft({ additionalInfo: e.target.value })} placeholder="补充介绍你的技术方向、项目经验、学习能力、目标岗位" />
        <div className="mt-2 flex items-center justify-between"><span className="text-[12px] text-slate-400">建议包含：技术方向、项目经验、学习能力、目标岗位。</span><span className="text-[12px] text-slate-400">当前约 {extraInfoDraft.additionalInfo.length} 字</span></div>
      </PortfolioExtraSection>
    </ResumeEditorShell>
    <AppConfirmDialog
      open={formAlert.open}
      title={formAlert.title}
      description={formAlert.description}
      confirmText="我知道了"
      variant={formAlert.variant}
      hideCancel
      onConfirm={() => setFormAlert((prev) => ({ ...prev, open: false }))}
      onCancel={() => setFormAlert((prev) => ({ ...prev, open: false }))}
    />
    </>
  );
}
