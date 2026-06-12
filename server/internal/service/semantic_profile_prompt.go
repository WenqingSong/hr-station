package service

import (
	"fmt"
	"strings"
	"time"
)

// ─── ResumeSemanticProfilePrompt ────────────────────────────────────

// BuildResumeSemanticProfilePrompt builds the system+user prompt for Hy3 resume profiler.
func BuildResumeSemanticProfilePrompt(resumeID int64, resumeName, rawText string) (systemPrompt, userPrompt string) {
	systemPrompt = `你是校园招聘场景下的简历语义画像专家（Resume Semantic Profiler）。

你的任务不是优化简历，也不是匹配岗位，而是从这份简历中提取可供推荐、匹配、优化、对比使用的语义画像。

你必须严格基于简历事实，不允许虚构经历。

输出要求：
1. 一句话候选人画像 summary_sentence（30-80 字）。
2. 8-16 个高质量标签 tags，覆盖：
   - role_direction：角色/方向标签（如后端开发、算法研究、产品运营）
   - technical_skill：技术能力标签（如 Go/Python/Java 后端开发）
   - engineering_capability：工程能力标签（如分布式系统设计、性能优化）
   - domain_context：领域背景标签（如电商、金融、教育科技）
   - experience_signal：经历信号标签（如实习经历、项目落地、科研产出）
3. 每个标签必须有：
   - tag_code：内部代码（snake_case）
   - tag_label：用户可读中文标签（必须是完整中文短语，不允许只写 Go/Redis/MySQL）
   - category：必须是上述 5 类之一
   - confidence：0-1 之间的置信度
   - evidence_text：来自简历原文的证据（不能胡编）
   - explanation：为什么这个标签重要
   - source_section：证据所在简历段落
   - is_user_visible：始终为 true
   - weight：0-1 权重
4. strengths：2-4 条核心优势（每条 15-40 字中文）。
5. risks：1-3 条风险信号（每条 15-40 字中文），包含 label/severity/evidence_text/explanation/confidence。
   风险必须真实，不能为了凑数制造风险。
6. fit_directions：2-3 个适配岗位方向（如"后端开发"、"算法研究"）。
7. profile_notes：可选的分析备注。

严禁：
- 只输出 Go/Redis/MySQL 这种短标签。
- 把岗位要求写成候选人能力。
- 输出 skill_0、degree_bachelor、internship_exp 这种 raw code。
- 没有证据的能力打标签。
- 虚构经历。

输出 JSON 格式严格遵守 schema。`

	userPrompt = fmt.Sprintf(`请为以下简历生成语义画像：

简历ID：%d
简历名称：%s
简历内容：
%s

请输出 JSON。`, resumeID, resumeName, rawText)
	return
}

// ─── JobSemanticProfilePrompt ────────────────────────────────────────

// BuildJobSemanticProfilePrompt builds the system+user prompt for Hy3 job profiler.
func BuildJobSemanticProfilePrompt(jobID int64, title, description, requirements, direction, category string) (systemPrompt, userPrompt string) {
	systemPrompt = `你是校园招聘岗位 JD 语义画像专家（Job Semantic Profiler）。

你的任务不是推荐候选人，而是判断这个岗位真正需要什么能力、经验和筛选信号。

你必须严格基于岗位 JD，不允许添加 JD 没有表达的硬性要求。

输出要求：
1. 一句话岗位画像 summary_sentence（30-80 字）。
2. 8-16 个高质量岗位标签 tags，覆盖：
   - role_direction：岗位方向标签
   - technical_skill：岗位技术要求
   - engineering_capability：工程能力要求
   - domain_context：业务领域要求
   - experience_signal：期望候选人具备的经历类型（不是硬性要求）
   - risk_signal：潜在的筛选门槛（如"偏好硕士"、"更看重实习经历"）
   - preference_signal：偏好信号（如"有 ACM/竞赛经历者优先"）
3. 每个标签必须有 tag_code/tag_label/category/confidence/evidence_text/explanation/source_section/is_user_visible/weight。
   tag_label 必须是中文可读短语，不允许只输出 backend/redis/algorithm。
   evidence_text 必须来自 JD 原文。
4. strengths：2-4 条岗位核心要求总结。
5. risks：1-3 条风险/门槛信号，表达为"岗位可能更看重/偏好"，不要绝对化。
6. fit_directions：2-3 个适合的候选人类型。
7. profile_notes：可选的分析备注。

严禁：
- 把 JD 没写的内容变成硬性要求。
- 输出 degree_bachelor/degree_master/internship_exp 等 raw code。
- 只输出英文短标签。

输出 JSON 格式严格遵守 schema。`

	userPrompt = fmt.Sprintf(`请为以下岗位生成语义画像：

岗位ID：%d
岗位名称：%s
岗位方向：%s
岗位类别：%s
岗位描述：
%s
岗位要求：
%s

请输出 JSON。`, jobID, title, direction, category, description, requirements)

	// dedent for cleaner input
	userPrompt = strings.TrimSpace(userPrompt)
	return
}

// BuildSourceFromResume extracts text for profiling. Truncates if needed.
func BuildSourceFromResume(name, rawText string, updatedAt time.Time) string {
	// Hy3 context ~128k tokens; 20k chars is safe
	const maxLen = 20000
	text := fmt.Sprintf("简历名称：%s\n更新时间：%s\n\n%s", name, updatedAt.Format("2006-01-02"), rawText)
	if len(text) > maxLen {
		text = text[:maxLen] + "\n\n[内容已截断，保留前20000字符]"
	}
	return text
}

// BuildSourceFromJob builds the text input for job profiling.
func BuildSourceFromJob(title, description, requirements, direction, category, city, department string) string {
	var sb strings.Builder
	if title != "" {
		sb.WriteString(fmt.Sprintf("岗位名称：%s\n", title))
	}
	if direction != "" {
		sb.WriteString(fmt.Sprintf("方向：%s\n", direction))
	}
	if category != "" {
		sb.WriteString(fmt.Sprintf("类别：%s\n", category))
	}
	if city != "" {
		sb.WriteString(fmt.Sprintf("城市：%s\n", city))
	}
	if department != "" {
		sb.WriteString(fmt.Sprintf("部门：%s\n", department))
	}
	sb.WriteString("\n")
	if description != "" {
		sb.WriteString(fmt.Sprintf("岗位描述：\n%s\n\n", description))
	}
	if requirements != "" {
		sb.WriteString(fmt.Sprintf("岗位要求：\n%s\n", requirements))
	}
	return strings.TrimSpace(sb.String())
}
