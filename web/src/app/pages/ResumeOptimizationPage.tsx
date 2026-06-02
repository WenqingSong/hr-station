import React, { useState, useEffect } from "react";
import { useSearchParams } from "react-router";
import { ResumeOptimizationHeader } from "../components/ResumeOptimizationHeader";
import { ResumeOptimizationSidebarLeft } from "../components/ResumeOptimizationSidebarLeft";
import { ResumeOptimizationDetail } from "../components/ResumeOptimizationDetail";
import { ResumeOptimizationSidebarRight } from "../components/ResumeOptimizationSidebarRight";
import { loadResumeOptimizationViewModel, loadResumeOptionViewModelsFromApi, loadResumeVersionDetail } from "../../api";
import { getJobDetail } from "../../api/jobs";
import type { JobApiItem } from "../../api/jobs";
import type { ResumeOptimizationPageViewModel, ResumeOptionViewModel } from "../../adapters";
import type { Resume } from "../../types";

/** 从 URL 查询参数中解析 resume versionNo（兼容 legacy resumeId=resume-vN） */
function parseVersionNoFromQuery(resumeId: string | null, resumeVersionNo: string | null): number | null {
  if (resumeVersionNo && /^\d+$/.test(resumeVersionNo)) return Number(resumeVersionNo);
  if (resumeId) {
    if (/^\d+$/.test(resumeId)) return Number(resumeId);
    const match = resumeId.match(/^resume-v(\d+)$/);
    if (match) return Number(match[1]);
  }
  return null;
}

/** 从 Resume 详情组装 resume_text 供优化 Agent 使用 */
function buildResumeText(resume: Resume): string {
  const parts: string[] = [];
  const p = resume.profile;
  if (p?.name || resume.name) parts.push(`姓名：${p?.name || resume.name}`);
  if (resume.direction_label) parts.push(`求职方向：${resume.direction_label}`);
  if (resume.expected_cities?.length) parts.push(`期望城市：${resume.expected_cities.join("、")}`);

  if (resume.education?.length) {
    const eduLines = resume.education.map(
      (e) => `${e.school} ${e.major} ${e.degree} (${e.start_date}-${e.end_date})`
    );
    parts.push(`教育经历：\n${eduLines.join("\n")}`);
  }

  if (resume.experiences?.length) {
    const expLines = resume.experiences.map(
      (e) => `${e.company} ${e.role} (${e.start_date}-${e.end_date})：${e.description}`
    );
    parts.push(`工作/实习经历：\n${expLines.join("\n")}`);
  }

  if (resume.projects?.length) {
    const projLines = resume.projects.map(
      (p) => `【${p.name}】${p.description || ""}（技术栈：${(p.tech_stack || []).join("、")}）`
    );
    parts.push(`项目经历：\n${projLines.join("\n")}`);
  }

  if (resume.skills?.length) parts.push(`技能：${resume.skills.join("、")}`);
  if (resume.self_evaluation) parts.push(`自我评价：${resume.self_evaluation}`);

  return parts.join("\n\n");
}

/** 从 JobApiItem 组装 target_job 供优化 Agent 使用 */
function buildTargetJob(job: JobApiItem) {
  const jdText = [
    job.description,
    job.requirements ? `岗位要求：\n${job.requirements.split('\n').filter(r => r.trim()).join('\n')}` : '',
  ].filter(Boolean).join('\n\n');

  return { title: job.title, jd_text: jdText };
}

export function ResumeOptimizationPage() {
  const [searchParams] = useSearchParams();
  const jobId = searchParams.get("jobId");
  const resumeId = searchParams.get("resumeId");
  const resumeVersionNo = searchParams.get("resumeVersionNo");

  // 目标岗位
  const [job, setJob] = useState<JobApiItem | null>(null);
  const [loadingJob, setLoadingJob] = useState(!!jobId);

  // 简历选项 & 选中状态
  const [resumeOptions, setResumeOptions] = useState<ResumeOptionViewModel[]>([]);
  const [selectedOption, setSelectedOption] = useState<ResumeOptionViewModel | null>(null);
  const [loadingOptions, setLoadingOptions] = useState(true);

  // 选中版本的简历详情
  const [selectedResumeDetail, setSelectedResumeDetail] = useState<Resume | null>(null);
  const [loadingResumeDetail, setLoadingResumeDetail] = useState(false);
  const [resumeDetailError, setResumeDetailError] = useState<string | null>(null);

  // 优化结果
  const [result, setResult] = useState<ResumeOptimizationPageViewModel | null>(null);
  const [loadingOptimize, setLoadingOptimize] = useState(false);
  const [optimizeError, setOptimizeError] = useState<string | null>(null);

  // 加载目标岗位
  useEffect(() => {
    if (!jobId) { setLoadingJob(false); return; }
    getJobDetail(jobId)
      .then(setJob)
      .catch(() => {})
      .finally(() => setLoadingJob(false));
  }, [jobId]);

  // 加载简历选项列表
  useEffect(() => {
    loadResumeOptionViewModelsFromApi()
      .then((opts) => {
        setResumeOptions(opts);
        if (opts.length === 0) return;
        const targetVersionNo = parseVersionNoFromQuery(resumeId, resumeVersionNo);
        const matched = targetVersionNo
          ? opts.find((o) => Number(o.id) === targetVersionNo)
          : undefined;
        setSelectedOption(matched ?? opts[0]);
      })
      .catch(() => {})
      .finally(() => setLoadingOptions(false));
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  // 根据选中版本号加载简历版本详情
  useEffect(() => {
    if (!selectedOption) return;
    const versionNo = Number(selectedOption.id);
    if (!versionNo || isNaN(versionNo)) return;

    setLoadingResumeDetail(true);
    setResumeDetailError(null);
    setSelectedResumeDetail(null);
    loadResumeVersionDetail(versionNo)
      .then((detail) => {
        if (!detail) {
          setResumeDetailError("简历版本详情加载失败，请重新选择简历版本后再进行优化。");
        } else {
          setSelectedResumeDetail(detail);
        }
      })
      .catch(() =>
        setResumeDetailError("简历版本详情加载失败，请重新选择简历版本后再进行优化。")
      )
      .finally(() => setLoadingResumeDetail(false));
  }, [selectedOption]);

  // 触发优化 Agent（job + resume detail 均加载成功后）
  useEffect(() => {
    if (!job || !selectedResumeDetail) return;

    const resumeText = buildResumeText(selectedResumeDetail);
    const targetJob = buildTargetJob(job);

    setLoadingOptimize(true);
    setOptimizeError(null);
    loadResumeOptimizationViewModel({
      resume_text: resumeText,
      target_job: targetJob,
      target_section: "project",
    })
      .then(setResult)
      .catch(() => setOptimizeError("简历优化失败，请稍后重试。"))
      .finally(() => setLoadingOptimize(false));
  }, [job, selectedResumeDetail]);

  // ── 空态：缺 jobId ──
  if (!jobId) {
    return (
      <div className="max-w-[1440px] mx-auto px-6 py-12 text-center">
        <p className="text-gray-500 text-sm mb-3">缺少目标岗位</p>
        <p className="text-gray-400 text-xs mb-4">请先选择一个岗位，再进行简历优化。</p>
        <a href="/jobs" className="text-[#0052D9] text-sm underline">返回岗位列表</a>
      </div>
    );
  }

  // ── 空态：加载中 ──
  if (loadingJob || loadingOptions) {
    return (
      <div className="max-w-[1440px] mx-auto px-6 py-12 text-center text-gray-400 text-sm">
        正在加载岗位和简历信息…
      </div>
    );
  }

  // ── 空态：岗位加载失败 ──
  if (!job) {
    return (
      <div className="max-w-[1440px] mx-auto px-6 py-12 text-center">
        <p className="text-gray-500 text-sm mb-3">目标岗位加载失败</p>
        <p className="text-gray-400 text-xs mb-4">请返回岗位列表重新选择岗位。</p>
        <a href="/jobs" className="text-[#0052D9] text-sm underline">返回岗位列表</a>
      </div>
    );
  }

  // ── 空态：无简历版本 ──
  if (resumeOptions.length === 0 || !selectedOption) {
    return (
      <div className="max-w-[1440px] mx-auto px-6 py-12 text-center">
        <p className="text-gray-500 text-sm mb-3">你还没有可用于优化的简历</p>
        <p className="text-gray-400 text-xs mb-4">请先创建一份简历，系统会基于你的简历内容给出优化建议。</p>
        <a href="/resumes/new" className="text-[#0052D9] text-sm underline">去添加简历</a>
      </div>
    );
  }

  // ── 空态：版本详情加载中 ──
  if (loadingResumeDetail) {
    return (
      <div className="max-w-[1440px] mx-auto px-6 py-12 text-center text-gray-400 text-sm">
        正在加载简历版本详情…
      </div>
    );
  }

  // ── 空态：版本详情加载失败 ──
  if (resumeDetailError) {
    return (
      <div className="max-w-[1440px] mx-auto px-6 py-12 text-center">
        <p className="text-gray-500 text-sm mb-3">{resumeDetailError}</p>
        <p className="text-gray-400 text-xs mb-4">请重新选择简历版本后再进行优化。</p>
        <a href="/resumes" className="text-[#0052D9] text-sm underline">管理我的简历</a>
      </div>
    );
  }

  // ── 加载优化结果 ──
  if (loadingOptimize) {
    return (
      <div className="max-w-[1440px] mx-auto px-6 py-12 text-center text-gray-400 text-sm">
        正在生成简历优化建议…
      </div>
    );
  }

  // ── 优化失败 ──
  if (!result) {
    return (
      <div className="max-w-[1440px] mx-auto px-6 py-12 text-center">
        <p className="text-gray-500 text-sm mb-2">{optimizeError || "优化接口暂时不可用，请稍后重试。"}</p>
      </div>
    );
  }

  const jobSource = "目标岗位";
  const resumeSource = `${selectedOption.statusLabel} · ${selectedOption.directionLabel || "求职方向未填写"}`;

  // selectedResumeDetail 可能缺少子组件期望的某字段（如 id），用兜底保证传递
  const resumeForDisplay: Resume = selectedResumeDetail
    ? selectedResumeDetail
    : undefined as unknown as Resume;

  return (
    <div className="mx-auto max-w-[1440px] px-6 md:px-8 py-6 pb-12">
      <ResumeOptimizationHeader job={job} resume={resumeForDisplay} jobSource={jobSource} resumeSource={resumeSource} />
      <div className="flex gap-6 mt-6 items-start">
        <ResumeOptimizationSidebarLeft job={job} jobSource={jobSource} />
        <ResumeOptimizationDetail currentJobId={jobId} currentResumeId={String(selectedOption.id)} data={result} />
        <ResumeOptimizationSidebarRight resume={resumeForDisplay} resumeSource={resumeSource} />
      </div>
    </div>
  );
}