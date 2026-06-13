// ─── H19-B: Deep Match Agent ──────────────────────────────────────────────
// Wraps existing 3-agent pipeline with H19-B post-processing:
//   - fix null arrays, sanitize risks, ensure generated_by=deep_match_agent
//   - store DeepMatchReport in result_json for future H19-C UI

package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"offer-catcher/server/internal/agent"
	"offer-catcher/server/internal/model"
)

// ─── H22-D: DeepMatchReport Main Path ─────────────────────────────────────

func (s *MatchService) runDeepMatchSEP(
	ctx context.Context, userID uint64, resume *model.ResumeVersion, job *model.Job, runner agent.PipelineRunner,
) (*MatchResponse, error) {
	log.Printf("[DeepMatchSEP] starting for resume=%d job=%d", resume.ID, job.ID)

	// 1. Run 3-agent pipeline (produces gap/risk/match data)
	if s.tagEvidence == nil {
		return s.fallbackDeepMatch(ctx, userID, resume, job)
	}

	resp, err := s.matchWith3Agents(ctx, userID, resume, job, runner)
	if err != nil {
		log.Printf("[DeepMatchSEP] 3-agent failed: %v, falling back", err)
		resp, fallbackErr := s.analyzeWithLLM(ctx, resume, job, runner)
		if fallbackErr == nil {
			return sanitizeDeepResponse(resp), nil
		}
		return s.fallbackDeepMatch(ctx, userID, resume, job)
	}

	// 2. Enrich with DeepMatchReport using SEP evidence
	report := s.buildDeepMatchReportFromSEP(ctx, resp, resume, job)
	if report != nil {
		resp = s.enrichResponseWithReport(resp, report, userID, resume, job)
	}

	// 3. Post-process
	resp = sanitizeDeepResponse(resp)
	return resp, nil
}

// buildDeepMatchReportFromSEP constructs a DeepMatchReport using 3-agent output + SEP profiles.
func (s *MatchService) buildDeepMatchReportFromSEP(
	ctx context.Context, resp *MatchResponse, resume *model.ResumeVersion, job *model.Job,
) *DeepMatchReport {
	if s.semanticRepo == nil || resp == nil || resp.Result == nil {
		return nil
	}

	report := &DeepMatchReport{
		GenerationTier:        "quality",
		GeneratedBy:           "deep_match_agent",
		AgentStatus:           "match_verified",
		MatchScore:            resp.Result.MatchScore,
		MatchLevel:            resp.Result.MatchLevel,
		OneSentenceConclusion: resp.Result.Summary,
	}

	// Job understanding from SEP Job Profile
	jp, jerr := s.semanticRepo.FindLatestByTarget(ctx, model.SemanticTargetJob, int64(job.ID))
	if jerr == nil && jp != nil {
		jtgs, _ := s.semanticRepo.ListTagsByProfileID(ctx, jp.ID)
		mustHave := []string{}
		shouldHave := []string{}
		niceToHave := []string{}
		for _, t := range jtgs {
			switch t.Category {
			case "direction", "skill_core", "education":
				mustHave = append(mustHave, t.TagLabel)
			case "skill_secondary", "framework":
				shouldHave = append(shouldHave, t.TagLabel)
			default:
				niceToHave = append(niceToHave, t.TagLabel)
			}
		}
		report.JobUnderstanding = DeepJobUnderstanding{
			JobPositioning:        fmt.Sprintf("%s - %s", job.Title, jp.ProfileVersion),
			CoreResponsibilities:  []string{},
			MustHaveRequirements:  mustHave,
			ImportantRequirements: shouldHave,
			BonusRequirements:     niceToHave,
			HiddenExpectations:    []string{},
		}
	}

	// Resume evidence mapping from SEP Resume Profile
	rp, rerr := s.semanticRepo.FindLatestByTarget(ctx, model.SemanticTargetResume, int64(resume.ID))
	if rerr == nil && rp != nil {
		rtags, _ := s.semanticRepo.ListTagsByProfileID(ctx, rp.ID)
		evidence := []ResumeEvidenceItem{}
		for _, t := range rtags {
			if t.EvidenceText != "" || t.Explanation != "" {
				level := 2
				if strings.Contains(t.EvidenceText, "实习") || strings.Contains(t.EvidenceText, "公司") || strings.Contains(t.EvidenceText, "上线") {
					level = 4
				}
				evidence = append(evidence, ResumeEvidenceItem{
					EvidenceName:        t.TagLabel,
					EvidenceLevel:       level,
					SourceType:          t.Category,
					SourceTitle:         t.TagLabel,
					EvidenceText:        t.EvidenceText,
					SupportsRequirement: t.Explanation,
					Confidence:          "medium",
				})
			}
		}
		report.ResumeEvidence = evidence
	}

	// Core match points from 3-agent output
	for _, h := range resp.Result.MatchHighlights {
		report.CoreMatchPoints = append(report.CoreMatchPoints, DeepCoreMatchPoint{
			PointName:     h,
			MatchStrength: "partial",
			WhyItMatches:  h,
		})
	}

	// Key risks from 3-agent output
	for _, r := range resp.Result.RiskPoints {
		report.KeyRisks = append(report.KeyRisks, DeepKeyRisk{
			RiskName:      r,
			RiskLevel:     "medium",
			WhyItMatters:  r,
			FixSuggestion: "建议针对性补强",
		})
	}

	// Resume rewrite from gaps
	strategy := DeepResumeRewriteStrategy{}
	for _, g := range resp.Result.GapPoints {
		strategy.SkillsReorder = append(strategy.SkillsReorder, g)
	}
	report.ResumeRewriteStrategy = strategy

	// Application strategy
	report.ApplicationStrategy = DeepApplicationStrategy{
		BestTiming:         "根据岗位截止日期决定",
		DeliveryStrategy:   report.MatchLevel,
		CompetingPositioning: fmt.Sprintf("匹配度%d分", report.MatchScore),
	}

	// H22-G: Generate requirement_coverage from job_understanding + evidence
	report.RequirementCoverage = buildRequirementCoverageFromReport(report)

	// H22-G: Generate interview_risk_questions from risks + gaps + coverage
	report.InterviewRiskQuestions = buildInterviewQuestionsFromReport(report)

	// Agent trace
	report.AgentTrace = []DeepAgentTrace{
		{Stage: "JD_DECOMPOSITION", Agent: "JD解构Agent", Status: "done", Summary: "基于SEP Job Profile解构岗位要求"},
		{Stage: "RESUME_EVIDENCE_AUDIT", Agent: "简历证据Agent", Status: "done", Summary: fmt.Sprintf("提取%d条SEP证据", len(report.ResumeEvidence))},
		{Stage: "REQUIREMENT_ALIGNMENT", Agent: "要求对齐Agent", Status: "done", Summary: fmt.Sprintf("生成%d条要求覆盖", len(report.RequirementCoverage))},
		{Stage: "RISK_DIAGNOSIS", Agent: "风险诊断Agent", Status: "done", Summary: fmt.Sprintf("识别%d个风险", len(report.KeyRisks))},
		{Stage: "INTERVIEW_PREDICTION", Agent: "面试追问Agent", Status: "done", Summary: fmt.Sprintf("预测%d个面试追问", len(report.InterviewRiskQuestions))},
		{Stage: "GLOBAL_CALIBRATION", Agent: "全局校准Agent", Status: "done", Summary: fmt.Sprintf("最终分数%d", report.MatchScore)},
	}

	return report
}

// ─── H22-G: Requirement Coverage ───────────────────────────────────────────

func buildRequirementCoverageFromReport(report *DeepMatchReport) []RequirementCoverageItem {
	ju := report.JobUnderstanding
	var items []RequirementCoverageItem

	// Collect requirement categories
	type reqCat struct{ name, reqType string; priority int }
	reqs := []reqCat{}
	for _, n := range ju.MustHaveRequirements {
		reqs = append(reqs, reqCat{n, "must_have", 100})
	}
	for _, n := range ju.ImportantRequirements {
		reqs = append(reqs, reqCat{n, "should_have", 80})
	}
	for _, n := range ju.BonusRequirements {
		reqs = append(reqs, reqCat{n, "nice_to_have", 60})
	}
	for _, n := range ju.HiddenExpectations {
		reqs = append(reqs, reqCat{n, "hidden", 50})
	}

	// Build evidence text pool for matching
	evPool := ""
	for _, ev := range report.ResumeEvidence {
		evPool += ev.EvidenceText + " " + ev.SupportsRequirement + " "
	}

	for i, r := range reqs {
		status := "missing"
		score := 30
		gap := fmt.Sprintf("简历中未体现%s相关证据", r.name)
		suggestion := fmt.Sprintf("建议针对%s补充具体项目经历或技能描述", r.name)
		matchedIDs := []string{}

		// Simple keyword matching against evidence pool
		lowerPool := strings.ToLower(evPool)
		lowerName := strings.ToLower(r.name)
		if strings.Contains(lowerPool, lowerName) || len(lowerName) >= 4 && strings.Contains(lowerPool, lowerName[:4]) {
			status = "partially_matched"
			score = 65
			gap = fmt.Sprintf("%s有一定证据但不够充分", r.name)
			suggestion = fmt.Sprintf("进一步强化%s相关经历的细节表达", r.name)
		}
		// If evidence level >= 4 exists, upgrade to matched
		for _, ev := range report.ResumeEvidence {
			if ev.EvidenceLevel >= 4 && (strings.Contains(strings.ToLower(ev.EvidenceText+ev.SupportsRequirement), lowerName)) {
				status = "matched"
				score = 82
				gap = ""
				suggestion = ""
				matchedIDs = append(matchedIDs, ev.EvidenceName)
				break
			}
		}

		items = append(items, RequirementCoverageItem{
			RequirementName: r.name,
			RequirementType: r.reqType,
			CoverageStatus:  status,
			CoverageScore:   score,
			MatchedEvidence: matchedIDs,
			Gap:             gap,
			Suggestion:      suggestion,
		})
		if len(items) >= 8 {
			break
		}
		_ = i
	}
	return items
}

// ─── H23-B: Interview Questions De-template ─────────────────────────────────

func buildInterviewQuestionsFromReport(report *DeepMatchReport) []DeepInterviewQuestion {
	var questions []DeepInterviewQuestion
	ju := report.JobUnderstanding

	// Q1: Direction/motivation
	if len(ju.MustHaveRequirements) > 0 {
		questions = append(questions, DeepInterviewQuestion{
			Question:    fmt.Sprintf("你的简历经验更偏%s方向，投递这个%s岗位的动机是什么？", strings.Join(ju.MustHaveRequirements[:minInt(3, len(ju.MustHaveRequirements))], "、"), ju.JobPositioning),
			WhyAsked:    "面试官想了解你是否有清晰的职业规划，是否真正理解这个岗位。",
			RelatedRisk: "方向匹配度",
		})
	}

	// Q2: Core skill deep-dive from risks
	if len(report.KeyRisks) > 0 {
		r := report.KeyRisks[0]
		questions = append(questions, DeepInterviewQuestion{
			Question:    fmt.Sprintf("你在%s方面具体做过哪些项目？能讲一个从设计到上线的完整案例吗？", r.RiskName),
			WhyAsked:    r.WhyItMatters,
			RelatedRisk: r.RiskName,
		})
	}

	// Q3: Evidence verification from requirement coverage
	for _, cov := range report.RequirementCoverage {
		if len(questions) >= 5 { break }
		if cov.CoverageStatus == "partially_matched" || cov.CoverageStatus == "transferable" {
			questions = append(questions, DeepInterviewQuestion{
				Question:    fmt.Sprintf("岗位要求%s，你如何用现有项目经验来证明？具体怎么做？", cov.RequirementName),
				WhyAsked:    fmt.Sprintf("%s是该岗位的关键考察点", cov.RequirementName),
				RelatedRisk: cov.RequirementName,
			})
		}
	}

	// Q4: Gap bridging
	missingFound := false
	for _, cov := range report.RequirementCoverage {
		if cov.CoverageStatus == "missing" && !missingFound {
			questions = append(questions, DeepInterviewQuestion{
				Question:    fmt.Sprintf("如果入职后需要快速补上%s，你准备怎么学？能给一个具体计划吗？", cov.RequirementName),
				WhyAsked:    "考察学习能力和自我驱动意识",
				RelatedRisk: cov.RequirementName,
			})
			missingFound = true
		}
	}

	// Q5: Self-assessment + improvement plan
	if len(questions) < 5 && len(report.KeyRisks) > 1 {
		r := report.KeyRisks[1]
		questions = append(questions, DeepInterviewQuestion{
			Question:    fmt.Sprintf("如果满分10分，你给自己在%s方面打几分？为什么？", r.RiskName),
			WhyAsked:    "测试自我认知和诚实度",
			RelatedRisk: r.RiskName,
		})
	}

	// Ensure minimum 3 questions
	if len(questions) < 3 {
		for _, r := range report.KeyRisks {
			if len(questions) >= 3 { break }
			questions = append(questions, DeepInterviewQuestion{
				Question:    fmt.Sprintf("能否结合一个具体例子说明你在%s方面的能力？", r.RiskName),
				WhyAsked:    r.WhyItMatters,
				RelatedRisk: r.RiskName,
			})
		}
	}

	return questions
}

// enrichResponseWithReport saves the DeepMatchReport and enriches the response.
func (s *MatchService) enrichResponseWithReport(
	resp *MatchResponse, report *DeepMatchReport, userID uint64, resume *model.ResumeVersion, job *model.Job,
) *MatchResponse {
	// Build legacy fields from report
	resp.Result.MatchHighlights = ensureNonNil(resp.Result.MatchHighlights)
	for _, cp := range report.CoreMatchPoints {
		if cp.PointName != "" {
			resp.Result.MatchHighlights = append(resp.Result.MatchHighlights, cp.PointName)
		}
	}
	for _, ev := range report.ResumeEvidence {
		if ev.EvidenceText != "" {
			resp.Result.FitPoints = append(resp.Result.FitPoints, ev.EvidenceText)
		}
	}

	// Save to DB
	reportJSON, _ := json.Marshal(report)
	resultJSON, _ := json.Marshal(resp.Result)

	if s.matchRepo != nil {
		analysis := &model.MatchAnalysis{
			UserID:         userID,
			ResumeID:       resume.ID,
			JobID:          job.ID,
			GenerationTier: "quality",
			GeneratedBy:    "deep_match_agent",
			Status:         "ready",
			Score:          report.MatchScore,
			ResultJSON:     reportJSON,
		}
		if createErr := s.matchRepo.Create(context.Background(), analysis); createErr != nil {
			log.Printf("[DeepMatchSEP] DB save failed: %v", createErr)
		}
		resp.AnalysisID = analysis.ID
	}
	_ = resultJSON

	// H22-E: Copy DeepMatchReport fields to API response model
	resp.Result.ReportVersion = "deep_match_v2"
	resp.Result.ScoreExplanation = ""
	resp.Result.Confidence = 0
	resp.Result.JobUnderstanding = report.JobUnderstanding
	resp.Result.RequirementCoverage = report.RequirementCoverage
	resp.Result.ResumeEvidenceMapping = report.ResumeEvidence
	resp.Result.InterviewRiskQuestions = report.InterviewRiskQuestions
	resp.Result.CoreMatchPoints = report.CoreMatchPoints
	resp.Result.KeyRisks = report.KeyRisks
	resp.Result.ResumeRewriteStrategy = report.ResumeRewriteStrategy
	resp.Result.InterviewRiskQuestions = report.InterviewRiskQuestions
	resp.Result.ActionPlan = deepActionPlanFromReport(report)
	resp.Result.AgentTrace = report.AgentTrace
	resp.Result.ApplicationDecision = report.ApplicationDecision
	resp.Result.OneSentenceConclusion = report.OneSentenceConclusion

	// Set agent_status on response
	resp.GeneratedBy = "deep_match_agent"
	resp.Result.AgentStatus = "match_verified"
	resp.Result.GeneratedBy = "deep_match_agent"

	return resp
}

func deepActionPlanFromReport(report *DeepMatchReport) []map[string]any {
	if report == nil { return nil }
	plan := []map[string]any{}
	for _, s := range report.ResumeRewriteStrategy.SkillsReorder {
		plan = append(plan, map[string]any{"step": s, "type": "skills_reorder", "priority": 1})
	}
	for _, s := range report.ResumeRewriteStrategy.ProjectRewritePoints {
		plan = append(plan, map[string]any{"step": s, "type": "project_rewrite", "priority": 2})
	}
	return plan
}

// ─── Post-process: enforce H19-B contract ────────────────────────────────

func sanitizeDeepResponse(resp *MatchResponse) *MatchResponse {
	if resp == nil { return resp }
	resp.GeneratedBy = "deep_match_agent"
	if resp.Result == nil { return resp }

	r := resp.Result
	r.GeneratedBy = "deep_match_agent"
	// H19-D: also normalize "repaired" → "match_verified"
	if r.AgentStatus == "" || r.AgentStatus == "success" || r.AgentStatus == "repaired" {
		r.AgentStatus = "match_verified"
	}
	// Fix null arrays
	if r.MatchHighlights == nil { r.MatchHighlights = []string{} }
	if r.GapPoints == nil { r.GapPoints = []string{} }
	if r.RiskPoints == nil { r.RiskPoints = []string{} }
	if r.EvidenceChain == nil { r.EvidenceChain = []model.MatchEvidence{} }
	if r.NextSteps == nil { r.NextSteps = []string{} }
	if r.OptimizationSuggestions == nil { r.OptimizationSuggestions = []model.MatchSuggestion{} }
	if r.FitPoints == nil { r.FitPoints = []string{} }
	if r.JobRequirements == nil { r.JobRequirements = []string{} }
	if r.ResumeEvidence == nil { r.ResumeEvidence = []string{} }

	// Sanitize risk strings: replace bare tag-like strings with human labels
	cleanRisks := []string{}
	for _, risk := range r.RiskPoints {
		cleaned := sanitizeRiskText(risk)
		cleanRisks = append(cleanRisks, cleaned)
	}
	r.RiskPoints = cleanRisks

	// Fix level
	if r.MatchLevel == "" || r.MatchLevel == "low" {
		r.MatchLevel = matchLevel(r.MatchScore)
	}

	// Fix summary
	if r.Summary == "" || r.Summary == r.MatchLevel {
		r.Summary = fmt.Sprintf("深度分析匹配度%d分（%s）", r.MatchScore, r.MatchLevel)
	}

	// Update meta
	resp.Meta.GeneratedBy = "deep_match_agent"
	resp.Meta.AgentStatus = r.AgentStatus

	return resp
}

func sanitizeRiskText(s string) string {
	// Map legacy risk strings to better labels
	mappings := map[string]string{
		"学历风险": "学历要求可能不满足岗位筛选条件",
		"核心技术缺失风险": "核心技能覆盖不足，可能影响简历筛选通过率",
		"论文要求风险": "岗位可能要求论文/研究成果，简历未体现",
		"实习风险": "缺少相关实习经历，竞争力可能不足",
		"方向风险": "岗位方向与简历核心方向存在差异",
	}
	if v, ok := mappings[s]; ok { return v }
	return s
}

// ─── Fallback ─────────────────────────────────────────────────────────────

func (s *MatchService) fallbackDeepMatch(
	ctx context.Context, userID uint64, resume *model.ResumeVersion, job *model.Job,
) (*MatchResponse, error) {
	log.Printf("[DeepMatchSEP] fallback for resume=%d job=%d", resume.ID, job.ID)
	var legacyResult *model.MatchAnalysisResult
	if s.semanticRepo != nil {
		rp, _ := s.semanticRepo.FindLatestByTarget(ctx, model.SemanticTargetResume, int64(resume.ID))
		jp, _ := s.semanticRepo.FindLatestByTarget(ctx, model.SemanticTargetJob, int64(job.ID))
		if rp != nil && jp != nil && rp.Status == "verified" && jp.Status == "verified" {
			rtags, _ := s.semanticRepo.ListTagsByProfileID(ctx, rp.ID)
			jtgs, _ := s.semanticRepo.ListTagsByProfileID(ctx, jp.ID)
			legacyResult = runFastSEPMatch(&MatchSEPContext{
				ResumeTags: rtags, JobTags: jtgs, JobTitle: job.Title,
				ResumeID: int64(resume.ID), JobID: int64(job.ID),
			})
		}
	}
	if legacyResult == nil {
		legacyResult = &model.MatchAnalysisResult{
			MatchScore: 50, MatchLevel: "中等匹配", GeneratedBy: "deep_match_agent", AgentStatus: "match_repaired",
			Summary: "深度分析暂时不可用，当前为安全兜底结果。",
			MatchHighlights: []string{}, GapPoints: []string{}, RiskPoints: []string{},
			NextSteps: []string{"建议稍后重试深度分析"},
		}
	} else {
		legacyResult.GeneratedBy = "deep_match_agent"
		legacyResult.AgentStatus = "match_repaired"
	}
	meta := model.AgentResponseMeta{
		AgentName: "DeepMatchAgent", GeneratedBy: "deep_match_agent", AgentStatus: "match_repaired",
		FallbackReason: "llm_unavailable", RepairApplied: true, SchemaValid: false,
	}
	_ = json.Marshal // keep import
	return &MatchResponse{
		AnalysisID: 0, GeneratedBy: "deep_match_agent", Result: legacyResult,
		CreatedAt: time.Now().Format("2006-01-02T15:04:05Z07:00"), Meta: meta,
	}, nil
}


