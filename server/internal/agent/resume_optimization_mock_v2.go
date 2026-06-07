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

	// target_job: try to reuse match_analyses
	if mode == "target_job" && a.matchLatestFn != nil {
		ma := a.matchLatestFn(ctx, userID, resumeID, jobID)
		if ma != nil && len(ma.ResultJSON) > 0 {
			var r map[string]interface{}
			if json.Unmarshal(ma.ResultJSON, &r) == nil {
				// merge gap_points as risks
				if gaps, ok := r["gap_points"].([]interface{}); ok {
					for _, g := range gaps {
						if s, ok := g.(string); ok && len(risks) < 5 {
							risks = append(risks, "[JD差距] "+s)
						}
					}
				}
				// merge improvement_suggestions as content fixes
				if sugs, ok := r["improvement_suggestions"].([]interface{}); ok {
					for _, s := range sugs {
						if str, ok := s.(string); ok && len(content) < 4 {
							content = append(content, model.OptimizationFixItem{
								Section: "内容建议", Issue: str, Action: "建议在简历中体现",
							})
						}
					}
				}
			}
		}
	}

	// Fast trim
	if tier == "fast" {
		quickFixes = quickFixes[:min(4, len(quickFixes))]
		risks = risks[:min(2, len(risks))]
		structure = nil; content = nil; keyword = nil; evidence = nil
		if len(examples) > 1 { examples = examples[:1] }
		if len(longTerm) > 1 { longTerm = longTerm[:1] }
		if len(focus) > 80 { focus = focus[:80] + "..." }
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
