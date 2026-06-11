package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"offer-catcher/server/internal/model"
)

// ProfileProvider for ResumeOptimizationMockAgentV2
type OptimizeProfileProvider interface {
	EnsureProfile(ctx context.Context, userID, resumeID uint64, profileType string) (*model.ResumeProfile, string, error)
}

// MatchLatestProvider for reading match_analyses in target_job mode
type MatchLatestProvider interface {
	GetLatestMatchAnalysisForJob(ctx context.Context, userID, resumeID, jobID uint64) *model.MatchAnalysis
}

// ResumeOptimizationMockAgentV2 is the K9-3-C mock agent supporting 3 modes.
type ResumeOptimizationMockAgentV2 struct {
	profileProvider OptimizeProfileProvider
	// matchLatestFn optionally provided for target_job mode
	matchLatestFn func(ctx context.Context, userID, resumeID, jobID uint64) *model.MatchAnalysis
}

func NewResumeOptimizationMockAgentV2() *ResumeOptimizationMockAgentV2 {
	return &ResumeOptimizationMockAgentV2{}
}

func (a *ResumeOptimizationMockAgentV2) SetProfileProvider(pp OptimizeProfileProvider) {
	a.profileProvider = pp
}

func (a *ResumeOptimizationMockAgentV2) SetMatchLatestProvider(fn func(ctx context.Context, userID, resumeID, jobID uint64) *model.MatchAnalysis) {
	a.matchLatestFn = fn
}

type optProfile struct {
	dominant string
	tags     tagSet
}

func (a *ResumeOptimizationMockAgentV2) getProfile(ctx context.Context, userID, resumeID uint64, resume *model.ResumeVersion) optProfile {
	cp := optProfile{tags: make(tagSet)}
	if a.profileProvider != nil && resume != nil && resume.ID > 0 {
		p, st, err := a.profileProvider.EnsureProfile(ctx, userID, resumeID, "fast")
		if err == nil && st == model.ProfileStatusReady && p != nil && p.DominantProfile != "" {
			cp.dominant = p.DominantProfile
			if p.PrimaryTagsJSON != "" {
				var tags []model.ProfileTag
				if json.Unmarshal([]byte(p.PrimaryTagsJSON), &tags) == nil {
					for _, t := range tags {
						cp.tags.set(t.Tag, t.Confidence)
					}
				}
			}
			return cp
		}
	}
	// fallback
	if resume != nil {
		ct := extractCandidateTags(resume)
		cp.dominant = ct.dominantProfile
		cp.tags = ct.tags
	}
	return cp
}

// OptimizeV2 generates optimization result for one of 3 modes.
func (a *ResumeOptimizationMockAgentV2) OptimizeV2(
	ctx context.Context,
	userID, resumeID uint64,
	resume *model.ResumeVersion,
	mode string,
	targetDir string,
	jobID uint64,
	job *model.Job,
	tier string,
) (*model.OptimizationResultV2, error) {
	cp := a.getProfile(ctx, userID, resumeID, resume)

	genBy := "mock_agent"
	if tier == "quality" {
		genBy = "mock_quality"
	}

	score := calcOptScore(cp, mode, targetDir, job)
	level := optLevel(score)
	summary := buildOptSummary(cp, mode, targetDir, job)
	quickFixes := buildQuickFixes(cp, resume, mode, targetDir, job, 5)
	risks := buildOptRisks(cp, mode, targetDir, job)

	// Deep fields
	structure := buildStructureFixes(cp, resume, mode)
	content := buildContentFixes(cp, resume, mode, targetDir, job)
	keyword := buildKeywordFixes(cp, resume, mode, targetDir, job)
	evidence := buildEvidenceFixes(cp, resume, mode)
	examples := buildRewriteExamples(cp, resume, mode, targetDir, job)
	focus := buildDeliveryFocus(cp, mode, targetDir, job)
	longTerm := buildLongTerm(cp, mode)

	// target_job: try to reuse match_analyses with K9-3-C-Fix2 compatibility layer
	if mode == "target_job" && a.matchLatestFn != nil {
		ma := a.matchLatestFn(ctx, userID, resumeID, jobID)
		if ma != nil && len(ma.ResultJSON) > 0 {
			mc := extractMatchContext(ma.ResultJSON)
			if mc.HasMatch {
				// Merge gap_points into risks
				for _, g := range mc.GapPoints {
					if len(risks) < 6 { risks = append(risks, "[JD差距] "+g) }
				}
				// Merge improvement_suggestions into content_fix_items
				for _, s := range mc.ImprovementSuggestions {
					if len(content) < 5 {
						content = append(content, model.OptimizationFixItem{Section: "来自匹配分析", Issue: s, Action: "建议在简历中体现"})
					}
				}
				// Merge resume_evidence into evidence_fix_items
				for _, e := range mc.ResumeEvidence {
					if len(evidence) < 5 {
						evidence = append(evidence, model.OptimizationFixItem{Section: "匹配证据", Issue: "简历中的证据：" + e, Action: "在简历中强化此证据的表述"})
					}
				}
				// Merge match risk_points
				for _, r := range mc.RiskPoints {
					if len(risks) < 6 { risks = append(risks, "[匹配分析] "+r) }
				}
				// Blend delivery_strategy
				if mc.DeliveryStrategy != "" && len(focus) < 200 {
					focus = focus + " | 匹配分析建议：" + mc.DeliveryStrategy
				}
				// Add match-specific rewrite example
				if len(mc.GapPoints) > 0 && len(examples) < 4 {
					examples = append(examples, model.OptimizationRewriteExample{
						Section: "来自匹配分析", Before: "匹配分析发现差距：" + mc.GapPoints[0],
						After: "建议按照匹配分析的改进建议调整对应表述", Reason: "匹配分析明确了该岗位最关键的能力缺口",
					})
				}
			}
		}
	}

	// Fast trim — keep it light
	if tier == "fast" {
		quickFixes = quickFixes[:min(4, len(quickFixes))]
		risks = risks[:min(2, len(risks))]
		structure = nil; content = nil; keyword = nil; evidence = nil
		if len(examples) > 1 { examples = examples[:1] }
		if len(longTerm) > 1 { longTerm = longTerm[:1] }
		// K13-Stability-A: use rune count to safely trim UTF-8 text
		if runes := []rune(focus); len(runes) > 80 { focus = string(runes[:80]) + "..." }
	} else {
		// K9-3-C-Fix: quality must have >= 2 rewrite examples
		if len(examples) < 2 {
			examples = append(examples, buildExtraRewriteExamples(cp, resume, job)...)
		}
		// also expand other sections
		keyword = append(keyword, buildExtraKeywordFixes(cp)...)
		evidence = append(evidence, buildExtraEvidenceFixes(cp, resume)...)
		longTerm = append(longTerm, buildExtraLongTerm(cp)...)
	}

	result := &model.OptimizationResultV2{
		OptimizationID:         0,
		ResumeID:               resumeID,
		JobID:                  jobID,
		OptimizationMode:       mode,
		TargetDirection:        targetDir,
		GenerationTier:         tier,
		GeneratedBy:            genBy,
		Status:                 "ready",
		ProfileType:            tier,
		ProfileDominantProfile: cp.dominant,
		OptimizationScore:      score,
		OptimizationLevel:      level,
		OneSentenceSummary:     summary,
		QuickFixItems:          quickFixes,
		StructureFixItems:      structure,
		ContentFixItems:        content,
		KeywordFixItems:        keyword,
		EvidenceFixItems:       evidence,
		RiskPoints:             risks,
		RewriteExamples:        examples,
		DeliveryFocus:          focus,
		LongTermImprovements:   longTerm,
	}
	return result, nil
}

// ─── Scoring ──────────────────────────────────────────────────

func calcOptScore(cp optProfile, mode, targetDir string, job *model.Job) int {
	score := 60
	if cp.dominant != "" {
		score += 10
	}
	if len(cp.tags) >= 4 { score += 8 }
	if cp.tags.has("project_evidence", 2) { score += 6 }
	if cp.tags.has("internship_evidence", 2) { score += 4 }
	switch mode {
	case "target_direction":
		if targetDir == cp.dominant { score += 10 }
	case "target_job":
		if cp.dominant == "backend_engineering" && job != nil && strings.Contains(strings.ToLower(job.Title), "后台") { score += 10 }
		if cp.dominant == "algorithm_research" && job != nil && strings.Contains(strings.ToLower(job.Title), "算法") { score += 10 }
	}
	if score > 95 { score = 95 }
	return score
}

func optLevel(score int) string {
	if score >= 90 { return "建议重点优化后投递" }
	if score >= 80 { return "建议针对性优化后投递" }
	if score >= 70 { return "建议补强后投递" }
	return "建议先明确方向再优化"
}

// ─── Builders ─────────────────────────────────────────────────

func buildOptSummary(cp optProfile, mode, targetDir string, job *model.Job) string {
	switch mode {
	case "general":
		return fmt.Sprintf("当前简历在项目表达和技能呈现上有优化空间，建议重点补充量化成果和具体技术方案。")
	case "target_direction":
		return fmt.Sprintf("针对%s方向，建议在项目经历中突出相关技术点和业务理解。", targetDir)
	case "target_job":
		if job != nil {
			return fmt.Sprintf("针对%s岗位，建议围绕JD要求调整项目描述和目标表述。", job.Title)
		}
		return "建议围绕目标岗位JD要求调整简历描述。"
	}
	return ""
}

func buildQuickFixes(cp optProfile, resume *model.ResumeVersion, mode, targetDir string, job *model.Job, n int) []model.OptimizationFixItem {
	var items []model.OptimizationFixItem
	switch cp.dominant {
	case "algorithm_research":
		items = append(items, model.OptimizationFixItem{Section: "项目经历", Issue: "缺少量化实验指标", Action: "补充准确率/召回率/AUC/F1等实验评估指标"})
		items = append(items, model.OptimizationFixItem{Section: "项目经历", Issue: "描述偏笼统", Action: "写明模型方法、数据集规模和消融分析结果"})
		items = append(items, model.OptimizationFixItem{Section: "技能", Issue: "算法关键词不够突出", Action: "在技能栏前置PyTorch/TensorFlow/NLP/推荐系统等核心关键词"})
	case "backend_engineering":
		items = append(items, model.OptimizationFixItem{Section: "项目经历", Issue: "缺少架构和技术决策描述", Action: "写明技术选型原因、系统QPS、缓存策略、部署架构"})
		items = append(items, model.OptimizationFixItem{Section: "项目经历", Issue: "安全/权限描述不足", Action: "补充权限模型(RBAC/JWT)、越权修复、安全设计等"})
		items = append(items, model.OptimizationFixItem{Section: "技能", Issue: "技术栈关键词分散", Action: "在技能栏集中Go/MySQL/Redis/Docker/Linux等核心栈"})
	case "product_operation":
		items = append(items, model.OptimizationFixItem{Section: "项目经历", Issue: "缺少数据驱动成果", Action: "补充DAU/转化率/留存率等业务指标"})
		items = append(items, model.OptimizationFixItem{Section: "实习经历", Issue: "描述偏职责化", Action: "用STAR法则描述需求分析、PRD、用户研究等具体工作"})
	case "business_solution":
		items = append(items, model.OptimizationFixItem{Section: "实习经历", Issue: "客户价值体现不足", Action: "写明对接客户类型、方案价值、落地成果"})
		items = append(items, model.OptimizationFixItem{Section: "技能", Issue: "行业理解未体现", Action: "补充行业分析、客户需求沟通、商务推进等能力描述"})
	}
	if len(items) == 0 {
		items = append(items, model.OptimizationFixItem{Section: "整体", Issue: "简历结构可优化", Action: "建议按「教育背景→实习经历→项目经历→技能→附加信息」排序"})
	}
	if n > 0 && len(items) > n { items = items[:n] }
	return items
}

func buildOptRisks(cp optProfile, mode, targetDir string, job *model.Job) []string {
	var risks []string
	switch mode {
	case "target_direction":
		if targetDir != cp.dominant {
			risks = append(risks, fmt.Sprintf("当前画像(%s)与目标方向(%s)不完全一致，建议在简历中补强目标方向相关证据", cp.dominant, targetDir))
		}
	case "target_job":
		if cp.dominant == "backend_engineering" && job != nil && strings.Contains(strings.ToLower(job.Title), "算法") {
			risks = append(risks, "目标岗位偏算法方向，当前简历侧重工程，建议补充算法项目或调整投递方向")
		}
	}
	return risks
}

func buildStructureFixes(cp optProfile, resume *model.ResumeVersion, mode string) []model.OptimizationFixItem {
	return []model.OptimizationFixItem{
		{Section: "结构", Issue: "简历顺序", Action: "将最相关的经历放在每个模块最前面"},
		{Section: "结构", Issue: "篇幅控制", Action: "控制在一页以内，删除与目标方向无关的内容"},
	}
}

func buildContentFixes(cp optProfile, resume *model.ResumeVersion, mode, targetDir string, job *model.Job) []model.OptimizationFixItem {
	var items []model.OptimizationFixItem
	items = append(items, model.OptimizationFixItem{Section: "内容", Issue: "StAR原则", Action: "每条经历按「做了什么×怎么做的×结果如何」结构写"})
	items = append(items, model.OptimizationFixItem{Section: "内容", Issue: "动词开头", Action: "用「主导/负责/设计/优化/实现」开头替代「参与/协助」"})
	return items
}

func buildKeywordFixes(cp optProfile, resume *model.ResumeVersion, mode, targetDir string, job *model.Job) []model.OptimizationFixItem {
	switch cp.dominant {
	case "algorithm_research":
		return []model.OptimizationFixItem{{Section: "关键词", Issue: "算法方向", Action: "突出推荐系统/排序模型/消融分析/AUC/NLP/多模态/PyTorch"}}
	case "backend_engineering":
		return []model.OptimizationFixItem{{Section: "关键词", Issue: "工程方向", Action: "突出Go/MySQL/Redis/分布式/高并发/权限/RBAC/部署/Docker"}}
	default:
		return nil
	}
}

func buildEvidenceFixes(cp optProfile, resume *model.ResumeVersion, mode string) []model.OptimizationFixItem {
	return []model.OptimizationFixItem{
		{Section: "证据", Issue: "量化成果", Action: "每条项目补充至少一个量化指标(QPS/准确率/用户数/效率提升百分比)"},
		{Section: "证据", Issue: "技术深度", Action: "复杂技术方案补充关键设计决策和踩坑经验"},
	}
}

func buildRewriteExamples(cp optProfile, resume *model.ResumeVersion, mode, targetDir string, job *model.Job) []model.OptimizationRewriteExample {
	switch cp.dominant {
	case "backend_engineering":
		return []model.OptimizationRewriteExample{
			{Section: "项目经历", Before: "参与课程学习平台开发，负责后端接口开发。",
				After: "主导课程学习平台播放鉴权与权限模块开发，基于courseId/sectionId/videoId三元组校验修复越权播放风险，覆盖课程详情、播放签名、资源下载等核心接口。",
				Reason: "体现负责范围、技术动作、安全价值和业务闭环。"},
		}
	case "algorithm_research":
		return []model.OptimizationRewriteExample{
			{Section: "项目经历", Before: "用推荐算法对用户进行商品推荐。",
				After: "设计并实现多目标排序推荐模型，基于用户行为序列建模与消融分析，离线AUC提升3.2%，在线CTR提升1.8%。",
				Reason: "体现模型方法、实验评估和业务效果。"},
		}
	default:
		return []model.OptimizationRewriteExample{
			{Section: "项目经历", Before: "参与了XX项目开发。",
				After: "主导了XX项目的需求分析与方案设计，推动跨部门协作，按期交付并上线，日活提升15%。",
				Reason: "从参与改为主导，补充量化结果。"},
		}
	}
}

func buildDeliveryFocus(cp optProfile, mode, targetDir string, job *model.Job) string {
	switch cp.dominant {
	case "algorithm_research":
		return "投递前重点检查：模型指标是否量化、实验条件是否清晰、与岗位NLP/推荐/多模态方向的关联是否突出。"
	case "backend_engineering":
		return "投递前重点检查：接口QPS、缓存策略、权限模型、部署架构、数据库设计是否在项目描述中体现。"
	default:
		return "投递前重点检查：实习/项目经历是否用STAR法则描述、是否有量化成果、技能栏是否与目标岗位匹配。"
	}
}

func buildLongTerm(cp optProfile, mode string) []model.OptimizationLongTermItem {
	switch cp.dominant {
	case "algorithm_research":
		return []model.OptimizationLongTermItem{
			{Direction: "研究深度", Note: "建议积累论文/竞赛/开源贡献, 增加简历区分度"},
		}
	case "backend_engineering":
		return []model.OptimizationLongTermItem{
			{Direction: "系统广度", Note: "建议补充分布式、消息队列、微服务治理等方向的项目经历"},
		}
	default:
		return []model.OptimizationLongTermItem{{Direction: "职业方向", Note: "建议先明确目标方向，再针对该方向积累对应项目和实习"}}
	}
}

// ─── K9-3-C-Fix: Extra builders for quality differentiation ──────

func buildExtraRewriteExamples(cp optProfile, resume *model.ResumeVersion, job *model.Job) []model.OptimizationRewriteExample {
	switch cp.dominant {
	case "algorithm_research":
		return []model.OptimizationRewriteExample{
			{Section: "技能栏", Before: "熟悉机器学习、深度学习。",
				After: "熟练使用PyTorch/TensorFlow，掌握Transformer、GNN、强化学习等前沿模型架构；在推荐系统/NLP方向有模型训练、离线评估与消融分析经验。",
				Reason: "突出具体框架、模型和方向，而不是泛泛写熟悉。"},
			{Section: "教育经历", Before: "研究方向：推荐系统",
				After: "研究方向：多目标推荐排序模型 | 核心成果：离线AUC提升3.2%，在线CTR提升1.8%",
				Reason: "量化学术成果，补充技术价值和业务效果。"},
		}
	case "backend_engineering":
		return []model.OptimizationRewriteExample{
			{Section: "项目经历", Before: "使用Redis做缓存，提高查询性能。",
				After: "基于Redis实现多级缓存策略，将课程详情接口响应耗时从120ms降至15ms，支撑日均50万+ QPS查询。",
				Reason: "从功能描述改为性能量化，体现技术深度和业务价值。"},
			{Section: "技能栏", Before: "熟悉MySQL、Redis、Docker。",
				After: "MySQL：索引优化、分库分表、慢查询治理 | Redis：缓存策略、分布式锁、限流 | Docker：多服务编排、CI/CD流水线",
				Reason: "展开技术栈细节，让面试官看到深度而非表面了解。"},
		}
	case "product_operation":
		return []model.OptimizationRewriteExample{
			{Section: "实习经历", Before: "负责XX产品日常运营工作。",
				After: "主导XX产品用户增长活动：通过A/B测试优化注册流程转化率（+12%），策划3场裂变活动（新增2.3万用户），搭建用户分层运营体系（次月留存率提升8%）。",
				Reason: "从职责描述改为成果驱动，体现数据sense和项目推进能力。"},
		}
	case "business_solution":
		return []model.OptimizationRewriteExample{
			{Section: "实习经历", Before: "协助完成客户需求分析和方案撰写。",
				After: "独立完成2家头部客户需求调研，输出行业对比分析与定制化解决方案，推动签约金额超500万。",
				Reason: "从协助改为独立负责，量化签约成果，体现商务闭环能力。"},
		}
	default:
		return nil
	}
}

func buildExtraKeywordFixes(cp optProfile) []model.OptimizationFixItem {
	switch cp.dominant {
	case "algorithm_research":
		return []model.OptimizationFixItem{
			{Section: "关键词", Issue: "模型细节缺失", Action: "补充使用的模型架构名(BERT/ResNet/Transformer)和训练框架"},
			{Section: "关键词", Issue: "业务场景空白", Action: "在描述中体现出模型解决的业务问题而非纯技术任务"},
		}
	case "backend_engineering":
		return []model.OptimizationFixItem{
			{Section: "关键词", Issue: "中间件经验", Action: "补充消息队列(Kafka/RabbitMQ)、搜索引擎(ES)等中间件使用经验"},
			{Section: "关键词", Issue: "监控运维", Action: "补充Prometheus/Grafana监控和日志采集经验"},
		}
	case "product_operation":
		return []model.OptimizationFixItem{
			{Section: "关键词", Issue: "数据工具", Action: "补充SQL/Excel/Tableau等数据分析工具使用经验"},
		}
	default:
		return nil
	}
}

func buildExtraEvidenceFixes(cp optProfile, resume *model.ResumeVersion) []model.OptimizationFixItem {
	return []model.OptimizationFixItem{
		{Section: "证据", Issue: "对比维度", Action: "每条优化改动补充before/after对比，让数据变化可感知"},
		{Section: "证据", Issue: "技术文档", Action: "如果有技术博客、开源项目、技术分享，可作为简历附件链接"},
	}
}

func buildExtraLongTerm(cp optProfile) []model.OptimizationLongTermItem {
	switch cp.dominant {
	case "algorithm_research":
		return []model.OptimizationLongTermItem{
			{Direction: "工程能力", Note: "建议补充模型部署和工程化落地经验，提升完整算法交付能力"},
		}
	case "backend_engineering":
		return []model.OptimizationLongTermItem{
			{Direction: "系统设计", Note: "建议积累高并发系统架构设计经验，可以阅读DDIA等经典书籍"},
		}
	default:
		return []model.OptimizationLongTermItem{
			{Direction: "行业认知", Note: "建议持续关注目标行业动态，参加行业分享和社区活动"},
		}
	}
}

// ─── K9-3-C-Fix2: Match context compatibility layer ──────────────

type OptimizationMatchContext struct {
	HasMatch               bool
	MatchAnalysisID         uint64
	GapPoints               []string
	RiskPoints              []string
	ImprovementSuggestions  []string
	JobRequirements         []string
	ResumeEvidence          []string
	DeliveryStrategy        string
}

// extractMatchContext parses match_analyses result_json with field compatibility.
func extractMatchContext(resultJSON []byte) OptimizationMatchContext {
	var raw map[string]interface{}
	if json.Unmarshal(resultJSON, &raw) != nil {
		return OptimizationMatchContext{}
	}
	mc := OptimizationMatchContext{}
	// gap_points — new + old compat
	mc.GapPoints = extractStringArray(raw, "gap_points", "gaps", "weaknesses")
	// risk_points — new + old compat
	mc.RiskPoints = extractStringArray(raw, "risk_points", "risks", "risk_items")
	// improvement_suggestions — new + old compat
	mc.ImprovementSuggestions = extractStringArray(raw, "improvement_suggestions", "optimization_suggestions", "suggestions", "improvement_items")
	// job_requirements
	mc.JobRequirements = extractStringArray(raw, "job_requirements")
	// resume_evidence — new + old compat
	mc.ResumeEvidence = extractStringArray(raw, "resume_evidence", "match_highlights", "evidence")
	// delivery_strategy — new + old compat
	mc.DeliveryStrategy = extractStringField(raw, "delivery_strategy", "suggested_action", "delivery_focus")

	if len(mc.GapPoints) > 0 || len(mc.RiskPoints) > 0 || len(mc.ImprovementSuggestions) > 0 {
		mc.HasMatch = true
	}
	return mc
}

func extractStringArray(raw map[string]interface{}, keys ...string) []string {
	for _, k := range keys {
		v, ok := raw[k]
		if !ok { continue }
		switch arr := v.(type) {
		case []interface{}:
			var res []string
			for _, item := range arr {
				if s, ok := item.(string); ok { res = append(res, s) }
			}
			if len(res) > 0 { return res }
		case string:
			if arr != "" { return []string{arr} }
		}
	}
	return nil
}

func extractStringField(raw map[string]interface{}, keys ...string) string {
	for _, k := range keys {
		v, ok := raw[k]
		if !ok { continue }
		if s, ok := v.(string); ok && s != "" { return s }
	}
	return ""
}
