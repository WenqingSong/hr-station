import { useState, type KeyboardEvent } from "react";
import { Plus, X, Info } from "lucide-react";
import { color } from "../tokens";
import { PageShell, PageHeader } from "../components/layout/PageShell";
import { Card, CardHeader, CardDivider } from "../components/ui/Card";
import { Button } from "../components/ui/Button";

/* ─────────────── Primitives ─────────────── */

function Field({
  label,
  children,
}: {
  label: string;
  children: React.ReactNode;
}) {
  return (
    <div>
      <label
        style={{
          display: "block",
          fontSize: 13,
          color: color.textSecondary,
          marginBottom: 6,
        }}
      >
        {label}
      </label>
      {children}
    </div>
  );
}

const inputStyle: React.CSSProperties = {
  width: "100%",
  height: 36,
  padding: "0 12px",
  fontSize: 14,
  color: color.textPrimary,
  background: "#fff",
  border: `1px solid ${color.border}`,
  borderRadius: 6,
  outline: "none",
  boxSizing: "border-box",
  transition: "border-color 0.15s",
};

function TextInput({
  placeholder,
  value,
  onChange,
}: {
  placeholder?: string;
  value: string;
  onChange: (v: string) => void;
}) {
  return (
    <input
      style={inputStyle}
      placeholder={placeholder}
      value={value}
      onChange={(e) => onChange(e.target.value)}
      onFocus={(e) => (e.currentTarget.style.borderColor = color.blue)}
      onBlur={(e) => (e.currentTarget.style.borderColor = color.border)}
    />
  );
}

function SelectInput({
  options,
  value,
  onChange,
  placeholder,
}: {
  options: string[];
  value: string;
  onChange: (v: string) => void;
  placeholder?: string;
}) {
  return (
    <select
      style={{ ...inputStyle, cursor: "pointer", appearance: "auto" }}
      value={value}
      onChange={(e) => onChange(e.target.value)}
      onFocus={(e) => (e.currentTarget.style.borderColor = color.blue)}
      onBlur={(e) => (e.currentTarget.style.borderColor = color.border)}
    >
      {placeholder && <option value="">{placeholder}</option>}
      {options.map((o) => (
        <option key={o} value={o}>
          {o}
        </option>
      ))}
    </select>
  );
}

function TextArea({
  placeholder,
  value,
  onChange,
  rows = 4,
}: {
  placeholder?: string;
  value: string;
  onChange: (v: string) => void;
  rows?: number;
}) {
  return (
    <textarea
      style={{
        width: "100%",
        padding: "8px 12px",
        fontSize: 14,
        color: color.textPrimary,
        background: "#fff",
        border: `1px solid ${color.border}`,
        borderRadius: 6,
        outline: "none",
        resize: "vertical",
        boxSizing: "border-box",
        lineHeight: 1.6,
        fontFamily: "inherit",
        transition: "border-color 0.15s",
        minHeight: rows * 24,
      }}
      rows={rows}
      placeholder={placeholder}
      value={value}
      onChange={(e) => onChange(e.target.value)}
      onFocus={(e) => (e.currentTarget.style.borderColor = color.blue)}
      onBlur={(e) => (e.currentTarget.style.borderColor = color.border)}
    />
  );
}

const YEARS = Array.from({ length: 10 }, (_, i) => String(2020 + i));
const MONTHS = ["01", "02", "03", "04", "05", "06", "07", "08", "09", "10", "11", "12"];

function DatePicker({
  value,
  onChange,
}: {
  value: { year: string; month: string };
  onChange: (v: { year: string; month: string }) => void;
}) {
  return (
    <div style={{ display: "flex", gap: 8 }}>
      <select
        style={{ ...inputStyle, flex: 1, cursor: "pointer", appearance: "auto" }}
        value={value.year}
        onChange={(e) => onChange({ ...value, year: e.target.value })}
      >
        <option value="">年份</option>
        {YEARS.map((y) => <option key={y} value={y}>{y}</option>)}
      </select>
      <select
        style={{ ...inputStyle, flex: 1, cursor: "pointer", appearance: "auto" }}
        value={value.month}
        onChange={(e) => onChange({ ...value, month: e.target.value })}
      >
        <option value="">月份</option>
        {MONTHS.map((m) => <option key={m} value={m}>{m}</option>)}
      </select>
    </div>
  );
}

function TagInput({
  tags,
  onChange,
  placeholder,
}: {
  tags: string[];
  onChange: (tags: string[]) => void;
  placeholder?: string;
}) {
  const [input, setInput] = useState("");

  const add = () => {
    const trimmed = input.trim();
    if (trimmed && !tags.includes(trimmed)) onChange([...tags, trimmed]);
    setInput("");
  };

  const onKey = (e: KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "Enter" || e.key === ",") { e.preventDefault(); add(); }
    if (e.key === "Backspace" && !input && tags.length) {
      onChange(tags.slice(0, -1));
    }
  };

  return (
    <div
      style={{
        display: "flex",
        flexWrap: "wrap",
        gap: 6,
        padding: "6px 8px",
        border: `1px solid ${color.border}`,
        borderRadius: 6,
        background: "#fff",
        minHeight: 36,
        cursor: "text",
        transition: "border-color 0.15s",
      }}
      onClick={(e) => (e.currentTarget.querySelector("input") as HTMLInputElement)?.focus()}
    >
      {tags.map((tag) => (
        <span
          key={tag}
          style={{
            display: "inline-flex",
            alignItems: "center",
            gap: 4,
            background: color.blueLight,
            color: color.blue,
            borderRadius: 4,
            padding: "2px 8px",
            fontSize: 12,
            fontWeight: 500,
          }}
        >
          {tag}
          <X
            size={10}
            style={{ cursor: "pointer" }}
            onClick={(e) => { e.stopPropagation(); onChange(tags.filter((t) => t !== tag)); }}
          />
        </span>
      ))}
      <input
        style={{
          border: "none",
          outline: "none",
          fontSize: 13,
          color: color.textPrimary,
          minWidth: 80,
          flex: 1,
          background: "transparent",
          padding: "0 2px",
        }}
        placeholder={tags.length === 0 ? placeholder : ""}
        value={input}
        onChange={(e) => setInput(e.target.value)}
        onKeyDown={onKey}
        onBlur={add}
      />
    </div>
  );
}

/* 分组条：多条经历之间的分隔线 */
function EntryDivider() {
  return (
    <div
      style={{
        height: 1,
        background: color.divider,
        margin: "20px 0",
      }}
    />
  );
}

/* 两列 grid */
function Grid2({ children }: { children: React.ReactNode }) {
  return (
    <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: "16px 24px" }}>
      {children}
    </div>
  );
}

/* ─────────────── Version Hint ─────────────── */

function VersionHint({ isFirst }: { isFirst: boolean }) {
  return (
    <div
      style={{
        display: "flex",
        alignItems: "center",
        gap: 10,
        background: color.blueLight,
        border: `1px solid #C0D8FF`,
        borderRadius: 8,
        padding: "10px 16px",
        marginBottom: 16,
      }}
    >
      <Info size={15} color={color.blue} strokeWidth={2} style={{ flexShrink: 0 }} />
      <span style={{ fontSize: 13, color: color.blue }}>
        {isFirst
          ? "当前尚无简历版本，本次提交后将创建 V1。"
          : "当前主简历版本：V2，本次提交后将保存为 V3。"}
        <span style={{ color: color.textTertiary, marginLeft: 8 }}>
          你可以只填写当前已有的信息，系统会基于现有内容进行后续分析。
        </span>
      </span>
    </div>
  );
}

/* ─────────────── Section: Basic Info ─────────────── */

interface BasicInfoData {
  name: string;
  phone: string;
  email: string;
  city: string;
  direction: string;
}

function BasicInfoCard({
  data,
  onChange,
}: {
  data: BasicInfoData;
  onChange: (d: BasicInfoData) => void;
}) {
  const set = (k: keyof BasicInfoData) => (v: string) => onChange({ ...data, [k]: v });
  return (
    <Card>
      <CardHeader title="基本信息" />
      <Grid2>
        <Field label="姓名">
          <TextInput placeholder="请输入姓名" value={data.name} onChange={set("name")} />
        </Field>
        <Field label="手机号">
          <TextInput placeholder="请输入手机号" value={data.phone} onChange={set("phone")} />
        </Field>
        <Field label="邮箱">
          <TextInput placeholder="请输入邮箱" value={data.email} onChange={set("email")} />
        </Field>
        <Field label="当前所在城市">
          <TextInput placeholder="如：北京、上海" value={data.city} onChange={set("city")} />
        </Field>
        <Field label="目标岗位方向">
          <TextInput
            placeholder="如：后端开发、产品经理"
            value={data.direction}
            onChange={set("direction")}
          />
        </Field>
      </Grid2>
    </Card>
  );
}

/* ─────────────── Section: Job Intent ─────────────── */

interface JobIntentData {
  position: string;
  city: string;
  func: string;
  available: string;
  keywords: string[];
}

function JobIntentCard({
  data,
  onChange,
}: {
  data: JobIntentData;
  onChange: (d: JobIntentData) => void;
}) {
  const set = (k: keyof JobIntentData) => (v: string | string[]) =>
    onChange({ ...data, [k]: v });
  return (
    <Card>
      <CardHeader title="求职意向" />
      <Grid2>
        <Field label="期望岗位">
          <TextInput placeholder="如：后端开发实习生" value={data.position} onChange={set("position") as (v: string) => void} />
        </Field>
        <Field label="期望城市">
          <TextInput placeholder="如：北京、深圳" value={data.city} onChange={set("city") as (v: string) => void} />
        </Field>
        <Field label="期望职能方向">
          <SelectInput
            placeholder="请选择"
            options={["技术开发", "产品设计", "数据分析", "算法研究", "运营", "市场", "其他"]}
            value={data.func}
            onChange={set("func") as (v: string) => void}
          />
        </Field>
        <Field label="可到岗时间">
          <SelectInput
            placeholder="请选择"
            options={["随时到岗", "一周内", "一个月内", "三个月内", "半年内", "毕业后"]}
            value={data.available}
            onChange={set("available") as (v: string) => void}
          />
        </Field>
        <Field label="求职关键词">
          <TagInput
            tags={data.keywords}
            onChange={set("keywords") as (v: string[]) => void}
            placeholder="输入后回车添加，如 Go、MySQL"
          />
        </Field>
      </Grid2>
    </Card>
  );
}

/* ─────────────── Section: Education ─────────────── */

interface EduEntry {
  id: number;
  school: string;
  major: string;
  degree: string;
  start: { year: string; month: string };
  end: { year: string; month: string };
}

function EducationCard({
  entries,
  onChange,
}: {
  entries: EduEntry[];
  onChange: (e: EduEntry[]) => void;
}) {
  const update = (id: number, patch: Partial<EduEntry>) =>
    onChange(entries.map((e) => (e.id === id ? { ...e, ...patch } : e)));
  const add = () =>
    onChange([
      ...entries,
      { id: Date.now(), school: "", major: "", degree: "", start: { year: "", month: "" }, end: { year: "", month: "" } },
    ]);

  return (
    <Card>
      <CardHeader title="教育经历" />
      {entries.map((entry, i) => (
        <div key={entry.id}>
          {i > 0 && <EntryDivider />}
          <Grid2>
            <Field label="学校">
              <TextInput
                placeholder="如：北京大学"
                value={entry.school}
                onChange={(v) => update(entry.id, { school: v })}
              />
            </Field>
            <Field label="专业">
              <TextInput
                placeholder="如：计算机科学与技术"
                value={entry.major}
                onChange={(v) => update(entry.id, { major: v })}
              />
            </Field>
            <Field label="学历">
              <SelectInput
                placeholder="请选择"
                options={["专科", "本科", "硕士", "博士", "其他"]}
                value={entry.degree}
                onChange={(v) => update(entry.id, { degree: v })}
              />
            </Field>
            <div />
            <Field label="开始时间">
              <DatePicker value={entry.start} onChange={(v) => update(entry.id, { start: v })} />
            </Field>
            <Field label="结束时间">
              <DatePicker value={entry.end} onChange={(v) => update(entry.id, { end: v })} />
            </Field>
          </Grid2>
        </div>
      ))}
      <div style={{ marginTop: 16 }}>
        <button
          onClick={add}
          style={{
            display: "inline-flex",
            alignItems: "center",
            gap: 5,
            fontSize: 13,
            color: color.blue,
            background: "none",
            border: "none",
            cursor: "pointer",
            padding: 0,
          }}
        >
          <Plus size={14} />
          增加教育经历
        </button>
      </div>
    </Card>
  );
}

/* ─────────────── Section: Work Experience ─────────────── */

interface WorkEntry {
  id: number;
  company: string;
  role: string;
  start: { year: string; month: string };
  end: { year: string; month: string };
  desc: string;
}

function WorkCard({
  entries,
  onChange,
}: {
  entries: WorkEntry[];
  onChange: (e: WorkEntry[]) => void;
}) {
  const update = (id: number, patch: Partial<WorkEntry>) =>
    onChange(entries.map((e) => (e.id === id ? { ...e, ...patch } : e)));
  const add = () =>
    onChange([
      ...entries,
      { id: Date.now(), company: "", role: "", start: { year: "", month: "" }, end: { year: "", month: "" }, desc: "" },
    ]);

  return (
    <Card>
      <CardHeader title="实习 / 工作经历" />
      {entries.map((entry, i) => (
        <div key={entry.id}>
          {i > 0 && <EntryDivider />}
          <Grid2>
            <Field label="公司名称">
              <TextInput
                placeholder="如：字节跳动"
                value={entry.company}
                onChange={(v) => update(entry.id, { company: v })}
              />
            </Field>
            <Field label="职位名称">
              <TextInput
                placeholder="如：后端开发实习生"
                value={entry.role}
                onChange={(v) => update(entry.id, { role: v })}
              />
            </Field>
            <Field label="开始时间">
              <DatePicker value={entry.start} onChange={(v) => update(entry.id, { start: v })} />
            </Field>
            <Field label="结束时间">
              <DatePicker value={entry.end} onChange={(v) => update(entry.id, { end: v })} />
            </Field>
          </Grid2>
          <div style={{ marginTop: 14 }}>
            <Field label="工作描述">
              <TextArea
                placeholder="简要描述主要工作内容和负责项目..."
                value={entry.desc}
                onChange={(v) => update(entry.id, { desc: v })}
                rows={4}
              />
            </Field>
          </div>
        </div>
      ))}
      <div style={{ marginTop: 16 }}>
        <button
          onClick={add}
          style={{
            display: "inline-flex",
            alignItems: "center",
            gap: 5,
            fontSize: 13,
            color: color.blue,
            background: "none",
            border: "none",
            cursor: "pointer",
            padding: 0,
          }}
        >
          <Plus size={14} />
          增加经历
        </button>
      </div>
    </Card>
  );
}

/* ─────────────── Section: Projects ─────────────── */

interface ProjectEntry {
  id: number;
  name: string;
  role: string;
  start: { year: string; month: string };
  end: { year: string; month: string };
  stack: string[];
  desc: string;
  contribution: string;
}

function ProjectCard({
  entries,
  onChange,
}: {
  entries: ProjectEntry[];
  onChange: (e: ProjectEntry[]) => void;
}) {
  const update = (id: number, patch: Partial<ProjectEntry>) =>
    onChange(entries.map((e) => (e.id === id ? { ...e, ...patch } : e)));
  const add = () =>
    onChange([
      ...entries,
      {
        id: Date.now(),
        name: "",
        role: "",
        start: { year: "", month: "" },
        end: { year: "", month: "" },
        stack: [],
        desc: "",
        contribution: "",
      },
    ]);

  return (
    <Card>
      <CardHeader title="项目经历" />
      {entries.map((entry, i) => (
        <div key={entry.id}>
          {i > 0 && <EntryDivider />}
          <Grid2>
            <Field label="项目名称">
              <TextInput
                placeholder="如：秒杀系统设计"
                value={entry.name}
                onChange={(v) => update(entry.id, { name: v })}
              />
            </Field>
            <Field label="项目角色">
              <TextInput
                placeholder="如：后端负责人"
                value={entry.role}
                onChange={(v) => update(entry.id, { role: v })}
              />
            </Field>
            <Field label="开始时间">
              <DatePicker value={entry.start} onChange={(v) => update(entry.id, { start: v })} />
            </Field>
            <Field label="结束时间">
              <DatePicker value={entry.end} onChange={(v) => update(entry.id, { end: v })} />
            </Field>
            <div style={{ gridColumn: "1 / -1" }}>
              <Field label="技术栈">
                <TagInput
                  tags={entry.stack}
                  onChange={(v) => update(entry.id, { stack: v })}
                  placeholder="输入后回车添加，如 Go、Redis"
                />
              </Field>
            </div>
          </Grid2>
          <div style={{ marginTop: 14, display: "flex", flexDirection: "column", gap: 14 }}>
            <Field label="项目描述">
              <TextArea
                placeholder="简要描述项目背景与核心功能..."
                value={entry.desc}
                onChange={(v) => update(entry.id, { desc: v })}
                rows={3}
              />
            </Field>
            <Field label="个人贡献">
              <TextArea
                placeholder="描述你在项目中的具体职责与贡献..."
                value={entry.contribution}
                onChange={(v) => update(entry.id, { contribution: v })}
                rows={3}
              />
            </Field>
          </div>
        </div>
      ))}
      <div style={{ marginTop: 16 }}>
        <button
          onClick={add}
          style={{
            display: "inline-flex",
            alignItems: "center",
            gap: 5,
            fontSize: 13,
            color: color.blue,
            background: "none",
            border: "none",
            cursor: "pointer",
            padding: 0,
          }}
        >
          <Plus size={14} />
          增加项目经历
        </button>
      </div>
    </Card>
  );
}

/* ─────────────── Section: Skills ─────────────── */

interface SkillsData {
  core: string[];
  tools: string[];
  language: string;
  awards: string;
}

function SkillsCard({
  data,
  onChange,
}: {
  data: SkillsData;
  onChange: (d: SkillsData) => void;
}) {
  return (
    <Card>
      <CardHeader title="技能与补充能力" />
      <div style={{ display: "flex", flexDirection: "column", gap: 16 }}>
        <Grid2>
          <Field label="核心技能">
            <TagInput
              tags={data.core}
              onChange={(v) => onChange({ ...data, core: v })}
              placeholder="如 Go、MySQL、Redis"
            />
          </Field>
          <Field label="工具 / 框架">
            <TagInput
              tags={data.tools}
              onChange={(v) => onChange({ ...data, tools: v })}
              placeholder="如 Docker、Git、Linux"
            />
          </Field>
        </Grid2>
        <Grid2>
          <Field label="语言能力">
            <TextInput
              placeholder="如：英语 CET-6 / 日语 N2"
              value={data.language}
              onChange={(v) => onChange({ ...data, language: v })}
            />
          </Field>
          <Field label="获奖 / 证书">
            <TextInput
              placeholder="如：蓝桥杯省级一等奖"
              value={data.awards}
              onChange={(v) => onChange({ ...data, awards: v })}
            />
          </Field>
        </Grid2>
      </div>
    </Card>
  );
}

/* ─────────────── Section: Self Assessment ─────────────── */

function SelfAssessCard({
  value,
  onChange,
}: {
  value: string;
  onChange: (v: string) => void;
}) {
  return (
    <Card>
      <CardHeader title="自我评价" />
      <TextArea
        placeholder="可补充求职方向、项目亮点、能力总结等，帮助系统更准确地理解你..."
        value={value}
        onChange={onChange}
        rows={5}
      />
    </Card>
  );
}

/* ─────────────── Bottom Action Bar ─────────────── */

function ActionBar() {
  return (
    <div
      style={{
        display: "flex",
        alignItems: "center",
        justifyContent: "space-between",
        padding: "16px 24px",
        background: "#fff",
        borderRadius: 10,
        border: `1px solid ${color.border}`,
        marginTop: 4,
      }}
    >
      <span style={{ fontSize: 13, color: color.textTertiary }}>
        所有字段均为可选项，系统会基于已填写内容进行后续分析。
      </span>
      <div style={{ display: "flex", gap: 10 }}>
        <Button variant="ghost" size="md">
          返回我的简历
        </Button>
        <Button variant="outline" size="md">
          预览填写内容
        </Button>
        <Button variant="primary" size="md">
          保存为当前版本
        </Button>
      </div>
    </div>
  );
}

/* ─────────────── Page ─────────────── */

export function AddResumePage({ isFirstVersion = false }: { isFirstVersion?: boolean }) {
  const [basic, setBasic] = useState<BasicInfoData>({
    name: "",
    phone: "",
    email: "",
    city: "",
    direction: "",
  });
  const [intent, setIntent] = useState<JobIntentData>({
    position: "",
    city: "",
    func: "",
    available: "",
    keywords: [],
  });
  const [edu, setEdu] = useState<EduEntry[]>([
    { id: 1, school: "", major: "", degree: "", start: { year: "", month: "" }, end: { year: "", month: "" } },
  ]);
  const [work, setWork] = useState<WorkEntry[]>([
    { id: 1, company: "", role: "", start: { year: "", month: "" }, end: { year: "", month: "" }, desc: "" },
  ]);
  const [projects, setProjects] = useState<ProjectEntry[]>([
    { id: 1, name: "", role: "", start: { year: "", month: "" }, end: { year: "", month: "" }, stack: [], desc: "", contribution: "" },
  ]);
  const [skills, setSkills] = useState<SkillsData>({ core: [], tools: [], language: "", awards: "" });
  const [selfAssess, setSelfAssess] = useState("");

  return (
    <PageShell activeNav="resume">
      <PageHeader
        title="添加简历"
        subtitle="通过结构化表单录入简历内容，并保存为新的简历版本。"
      />

      <VersionHint isFirst={isFirstVersion} />

      <div style={{ display: "flex", flexDirection: "column", gap: 14 }}>
        <BasicInfoCard data={basic} onChange={setBasic} />
        <JobIntentCard data={intent} onChange={setIntent} />
        <EducationCard entries={edu} onChange={setEdu} />
        <WorkCard entries={work} onChange={setWork} />
        <ProjectCard entries={projects} onChange={setProjects} />
        <SkillsCard data={skills} onChange={setSkills} />
        <SelfAssessCard value={selfAssess} onChange={setSelfAssess} />
        <ActionBar />
      </div>
    </PageShell>
  );
}
