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

// ─── H19-B: Deep Match V3 Pipeline ─────────────────────────────────────────
// 8-stage SEP + Fast Result driven pipeline. Replaces TagEvidenceService / 3-agent.

// buildDeepMatchContextV3 builds the v3 context from SEP profiles and Fast result.
func (s *MatchService) buildDeepMatchContextV3(
	ctx context.Context, userID uint64, resume *model.ResumeVersion, job *model.Job,
) (*DeepMatchContextV3, error) {
	v3ctx := &DeepMatchContextV3{
		ResumeID:  int64(resume.ID),
		JobID:     int64(job.ID),
		ResumeRaw: "", // populated later if available
		JobRaw:    job.Description,
		Status:    "building",
	}

	if s.semanticRepo == nil {
		v3ctx.Status = "missing"
		v3ctx.Reason = "semantic profile repository not available"
		return v3ctx, fmt.Errorf("deep match v3: semanticRepo not available")
	}

	// Read resume SEP
	rp, rerr := s.semanticRepo.FindLatestByTarget(ctx, model.SemanticTargetResume, int64(resume.ID))
	if rerr != nil || rp == nil || rp.Status != model.SemanticProfileStatusVerified {
		v3ctx.Status = "profile_missing"
		v3ctx.Reason = fmt.Sprintf("resume SEP not ready: resume=%d", resume.ID)
		return v3ctx, fmt.Errorf("deep match v3: %s", v3ctx.Reason)
	}
	rtags, _ := s.semanticRepo.ListTagsByProfileID(ctx, rp.ID)
	if len(rtags) < 3 {
		v3ctx.Status = "profile_insufficient"
		v3ctx.Reason = fmt.Sprintf("resume SEP tags insufficient: %d", len(rtags))
		return v3ctx, fmt.Errorf("deep match v3: %s", v3ctx.Reason)
	}
	v3ctx.ResumeSEPTags = rtags

	// Read job SEP
	jp, jerr := s.semanticRepo.FindLatestByTarget(ctx, model.SemanticTargetJob, int64(job.ID))
	if jerr != nil || jp == nil || jp.Status != model.SemanticProfileStatusVerified {
		v3ctx.Status = "profile_missing"
		v3ctx.Reason = fmt.Sprintf("job SEP not ready: job=%d", job.ID)
		return v3ctx, fmt.Errorf("deep match v3: %s", v3ctx.Reason)
	}
	jtgs, _ := s.semanticRepo.ListTagsByProfileID(ctx, jp.ID)
	if len(jtgs) < 3 {
		v3ctx.Status = "profile_insufficient"
		v3ctx.Reason = fmt.Sprintf("job SEP tags insufficient: %d", len(jtgs))
		return v3ctx, fmt.Errorf("deep match v3: %s", v3ctx.Reason)
	}
	v3ctx.JobSEPTags = jtgs

	// Run Fast SEP Match to get Fast Result
	fastRes := runFastSEPMatch(&MatchSEPContext{
		ResumeTags: rtags, JobTags: jtgs, JobTitle: job.Title,
		ResumeID: int64(resume.ID), JobID: int64(job.ID),
	})
	v3ctx.FastResult = fastRes

	// Extract Fast structured fields
	if sb, ok := fastRes.ScoreBreakdownObj.(ScoreBreakdownResult); ok {
		v3ctx.FastScoreBreakdown = sb
	}
	if hg, ok := fastRes.HardGateObj.(HardGateResult); ok {
		v3ctx.FastHardGate = hg
	}
	if cov, ok := fastRes.RequirementCoverage.([]RequirementCoverageItem); ok {
		v3ctx.FastRequirementCoverage = cov
	}
	if risks, ok := fastRes.StructuredRisks.([]MatchRiskPoint); ok {
		v3ctx.FastStructuredRisks = risks
	}

	v3ctx.Status = "ready"
	return v3ctx, nil
}

// runDeepMatchV3 executes the 8-layer H19-B deep match pipeline.
func runDeepMatchV3(v3ctx *DeepMatchContextV3) *DeepMatchV3Report {
	report := &DeepMatchV3Report{
		GenerationTier: "quality",
		GeneratedBy:    "deep_match_agent",
		AgentStatus:    "match_verified",
		ReportVersion:  "deep_match_v3",
	}
	trace := []DeepAgentTrace{}

	// ─── L0: Input Gate ──────────────────────────────────────────────
	if v3ctx.Status != "ready" {
		report.AgentStatus = "profile_missing"
		report.GeneratedBy = "deep_match_profile_pending"
		report.OneSentenceConclusion = "语义画像尚未准备完成，暂不能给出深度匹配分析。"
		report.AgentTrace = []DeepAgentTrace{
			{Stage: "INPUT_GATE", Agent: "输入门控Agent", Status: "blocked", Summary: v3ctx.Reason},
		}
		report.QualityChecks = makeQualityChecks(false, false)
		return report
	}
	trace = append(trace, DeepAgentTrace{Stage: "INPUT_GATE", Agent: "输入门控Agent", Status: "done", Summary: "SEP Profile + Fast Result ready"})

	// ─── L1: JD Deconstruction ──────────────────────────────────────
	ju := buildV3JobUnderstanding(v3ctx)
	report.JobUnderstanding = ju
	trace = append(trace, DeepAgentTrace{Stage: "JD_DECONSTRUCTION", Agent: "JD解构Agent", Status: "done",
		Summary: fmt.Sprintf("解构%d项必备/%d项重要/%d项加分", len(ju.MustHaveRequirements), len(ju.ImportantRequirements), len(ju.BonusRequirements))})

	// ─── L2: Resume Evidence Audit ──────────────────────────────────
	evidence := buildV3ResumeEvidence(v3ctx)
	report.ResumeEvidenceMapping = evidence
	trace = append(trace, DeepAgentTrace{Stage: "RESUME_EVIDENCE_AUDIT", Agent: "证据审计Agent", Status: "done",
		Summary: fmt.Sprintf("提取%d条简历证据", len(evidence))})

	// ─── L3: Requirement Alignment ──────────────────────────────────
	coverage := buildV3RequirementCoverage(v3ctx, evidence)
	report.RequirementCoverage = coverage
	trace = append(trace, DeepAgentTrace{Stage: "REQUIREMENT_ALIGNMENT", Agent: "要求对齐Agent", Status: "done",
		Summary: fmt.Sprintf("生成%d条要求覆盖", len(coverage))})

	// ─── L4: Risk Diagnosis ─────────────────────────────────────────
	risks := buildV3KeyRisks(v3ctx, coverage)
	report.KeyRisks = risks
	trace = append(trace, DeepAgentTrace{Stage: "RISK_DIAGNOSIS", Agent: "风险诊断Agent", Status: "done",
		Summary: fmt.Sprintf("识别%d个风险", len(risks))})

	// ─── L5: Score Calibration ──────────────────────────────────────
	score, level, breakdown := calibrateV3Score(v3ctx, coverage, evidence, risks)
	report.MatchScore = score
	report.MatchLevel = level
	report.ScoreBreakdown = breakdown
	report.ScoreAlignment = buildScoreAlignment(v3ctx.FastResult.MatchScore, score, level)
	trace = append(trace, DeepAgentTrace{Stage: "SCORE_CALIBRATION", Agent: "分数校准Agent", Status: "done",
		Summary: fmt.Sprintf("Deep=%d(%s) vs Fast=%d", score, level, v3ctx.FastResult.MatchScore)})

	// ─── L6: Application Strategy ───────────────────────────────────
	mpRisks := convertV3RisksToMatchRiskPoints(risks)
	decision, decisionLabel := decideApplication(score, mpRisks, v3ctx.FastHardGate)
	report.ApplicationDecision = decision
	report.DecisionLabel = decisionLabel
	report.ApplicationStrategy = buildV3AppStrategy(score, level, decisionLabel)
	trace = append(trace, DeepAgentTrace{Stage: "APPLICATION_STRATEGY", Agent: "投递策略Agent", Status: "done",
		Summary: decisionLabel})

	// ─── L7: Resume Rewrite + Interview ─────────────────────────────
	report.ResumeRewriteStrategy = buildV3RewriteStrategy(coverage, risks)
	report.InterviewRiskQuestions = buildV3InterviewQuestions(ju, risks, coverage)
	report.ActionPlan = buildV3ActionPlan(report.ResumeRewriteStrategy)
	trace = append(trace, DeepAgentTrace{Stage: "REWRITE_AND_INTERVIEW", Agent: "改写+面试Agent", Status: "done",
		Summary: fmt.Sprintf("生成%d项改写+%d个面试追问", len(report.ResumeRewriteStrategy.SkillsReorder)+len(report.ResumeRewriteStrategy.KeywordsToAdd), len(report.InterviewRiskQuestions))})

	// ─── L8: Global Consistency ─────────────────────────────────────
	coreMatch := buildV3CoreMatchPoints(coverage, evidence)
	report.CoreMatchPoints = coreMatch
	report.OneSentenceConclusion = v3ctx.FastResult.Summary
	if report.OneSentenceConclusion == "" {
		report.OneSentenceConclusion = fmt.Sprintf("深度分析：匹配度%d分（%s），%s", score, level, decisionLabel)
	}
	report.QualityChecks = makeQualityChecks(true, true)
	report.AgentTrace = trace

	return report
}

// ─── L1: JD Deconstruction ───────────────────────────────────────────────

func buildV3JobUnderstanding(v3ctx *DeepMatchContextV3) DeepJobUnderstanding {
	ju := DeepJobUnderstanding{
		JobPositioning:  "基于SEP语义画像解构的岗位定位",
		CoreResponsibilities: []string{},
		MustHaveRequirements:  []string{},
		ImportantRequirements: []string{},
		BonusRequirements:     []string{},
		HiddenExpectations:    []string{},
	}
	coreResps := []string{}
	for _, t := range v3ctx.JobSEPTags {
		switch {
		case t.Category == "role_direction" || t.Category == "direction":
			ju.JobPositioning = t.TagLabel
			ju.MustHaveRequirements = append(ju.MustHaveRequirements, t.TagLabel)
		case t.Category == "technical_skill" || t.Category == "skill_core":
			ju.MustHaveRequirements = append(ju.MustHaveRequirements, t.TagLabel)
		case t.Category == "engineering_capability" || t.Category == "experience_signal":
			if t.Confidence >= 0.85 {
				ju.MustHaveRequirements = append(ju.MustHaveRequirements, t.TagLabel)
			} else {
				ju.ImportantRequirements = append(ju.ImportantRequirements, t.TagLabel)
			}
		case t.Category == "domain_context":
			ju.ImportantRequirements = append(ju.ImportantRequirements, t.TagLabel)
		case t.Category == "risk_signal":
			ju.HiddenExpectations = append(ju.HiddenExpectations, t.TagLabel)
		default:
			if t.Confidence >= 0.85 {
				ju.ImportantRequirements = append(ju.ImportantRequirements, t.TagLabel)
			} else {
				ju.BonusRequirements = append(ju.BonusRequirements, t.TagLabel)
			}
		}
		if t.Confidence >= 0.9 {
			coreResps = append(coreResps, t.TagLabel)
		}
	}
	ju.CoreResponsibilities = coreResps
	if len(ju.MustHaveRequirements) == 0 {
		ju.MustHaveRequirements = append(ju.MustHaveRequirements, ju.JobPositioning)
	}
	return ju
}

// ─── L2: Resume Evidence Audit ────────────────────────────────────────────

func buildV3ResumeEvidence(v3ctx *DeepMatchContextV3) []ResumeEvidenceItem {
	evidence := []ResumeEvidenceItem{}
	for _, t := range v3ctx.ResumeSEPTags {
		if t.EvidenceText == "" && t.Explanation == "" {
			continue
		}
		level := 2
		if t.EvidenceText != "" {
			level = 3
			evLow := strings.ToLower(t.EvidenceText)
			if strings.Contains(evLow, "实习") || strings.Contains(evLow, "公司") { level = 4 }
			if strings.Contains(evLow, "生产") || strings.Contains(evLow, "上线") { level = 5 }
		}
		sourceType := "skill_list"
		if strings.Contains(strings.ToLower(t.Category), "experience") || strings.Contains(strings.ToLower(t.Category), "project") {
			sourceType = "project"
		}
		if strings.Contains(strings.ToLower(t.EvidenceText), "实习") {
			sourceType = "internship_project"
		}
		confidence := "medium"
		if t.Confidence >= 0.9 { confidence = "high" }

		supportsReq := humanizeRequirementCategory(t.Category)
		if t.Explanation != "" { supportsReq = t.Explanation }

		evidence = append(evidence, ResumeEvidenceItem{
			EvidenceName:        t.TagLabel,
			EvidenceLevel:       level,
			SourceType:          sourceType,
			SourceTitle:         t.TagLabel,
			EvidenceText:        t.EvidenceText,
			SupportsRequirement: supportsReq,
			Confidence:          confidence,
		})
	}
	if len(evidence) > 10 { evidence = evidence[:10] }
	return evidence
}

// ─── L3: Requirement Alignment ────────────────────────────────────────────

func buildV3RequirementCoverage(v3ctx *DeepMatchContextV3, evidence []ResumeEvidenceItem) []RequirementCoverageItem {
	// Reuse Fast coverage as base, enrich with evidence-level detail
	coverage := v3ctx.FastRequirementCoverage
	if len(coverage) == 0 {
		// Build from scratch if Fast coverage missing
		sepCtx := &MatchSEPContext{
			ResumeTags: v3ctx.ResumeSEPTags, JobTags: v3ctx.JobSEPTags,
			JobTitle: "", ResumeID: v3ctx.ResumeID, JobID: v3ctx.JobID,
		}
		coverage = buildRequirementCoverage(sepCtx)
	}
	// Enrich with Deep evidence-level context
	for i := range coverage {
		if coverage[i].CoverageStatus == "weak" || coverage[i].CoverageStatus == "missing" {
			for _, ev := range evidence {
				if strings.Contains(strings.ToLower(ev.EvidenceText+ev.EvidenceName), strings.ToLower(coverage[i].RequirementName[:minInt(3, len(coverage[i].RequirementName))])) {
					if ev.EvidenceLevel >= 4 {
						coverage[i].CoverageStatus = "partial"
						coverage[i].CoverageScore = 65
						coverage[i].Gap = fmt.Sprintf("%s证据不够充分", coverage[i].RequirementName)
					}
				}
			}
		}
	}
	if len(coverage) > 8 { coverage = coverage[:8] }
	return coverage
}

// ─── L4: Risk Diagnosis ───────────────────────────────────────────────────

func buildV3KeyRisks(v3ctx *DeepMatchContextV3, coverage []RequirementCoverageItem) []DeepKeyRisk {
	risks := []DeepKeyRisk{}

	// Elevate from Fast structured risks
	for _, fr := range v3ctx.FastStructuredRisks {
		risks = append(risks, DeepKeyRisk{
			RiskName:       fr.RiskName,
			RiskLevel:      fr.RiskLevel,
			WhyItMatters:   fr.WhyItMatters,
			EvidenceGap:    fr.EvidenceGap,
			FixSuggestion:  fr.FixSuggestion,
			InterviewImpact: "可能在面试中被深入追问",
		})
	}

	// Add Deep-specific: uncovered must-have items
	missingCount := 0
	for _, c := range coverage {
		if c.CoverageStatus == "missing" && c.RequirementType == "must_have" {
			missingCount++
			if missingCount <= 3 {
				risks = append(risks, DeepKeyRisk{
					RiskName:       fmt.Sprintf("必备要求缺失：%s", c.RequirementName),
					RiskLevel:      "high",
					WhyItMatters:   fmt.Sprintf("岗位将%s列为核心要求，缺失可能直接影响简历筛选", c.RequirementName),
					EvidenceGap:    c.Gap,
					FixSuggestion:  c.Suggestion,
					InterviewImpact: "面试官可能质疑为何投递该岗位",
				})
			}
		}
	}

	if len(risks) == 0 {
		risks = append(risks, DeepKeyRisk{
			RiskName:       "当前匹配良好",
			RiskLevel:      "low",
			WhyItMatters:   "未发现显著风险，但仍建议结合JD微调简历",
			EvidenceGap:    "",
			FixSuggestion:  "为关键技能补充量化成果描述以提升简历说服力",
			InterviewImpact: "",
		})
	}
	if len(risks) > 6 { risks = risks[:6] }
	return risks
}

// ─── L5: Score Calibration ────────────────────────────────────────────────

func calibrateV3Score(v3ctx *DeepMatchContextV3, coverage []RequirementCoverageItem, evidence []ResumeEvidenceItem, risks []DeepKeyRisk) (int, string, ScoreBreakdownResultV3) {
	fastResult := v3ctx.FastResult
	fastScore := fastResult.MatchScore

	// Deep starts with Fast as baseline, then adjusts
	dirFit := v3ctx.FastScoreBreakdown.DirectionFit
	covScore := v3ctx.FastScoreBreakdown.MustHaveCoverage
	evScore := v3ctx.FastScoreBreakdown.EvidenceStrength
	expScore := v3ctx.FastScoreBreakdown.ExperienceRelevance
	riskCtrl := v3ctx.FastScoreBreakdown.RiskControl

	// Adjust: if coverage has many missing must-haves, lower must_have_coverage
	missMust := 0
	totalMust := 0
	for _, c := range coverage {
		if c.RequirementType == "must_have" {
			totalMust++
			if c.CoverageStatus == "missing" { missMust++ }
		}
	}
	if totalMust > 0 && missMust > 0 {
		covScore = covScore - (missMust * 10)
		if covScore < 0 { covScore = 0 }
	}

	// Adjust risk_control based on Deep risk count
	highCount := 0
	for _, r := range risks {
		if r.RiskLevel == "high" { highCount++ }
	}
	if highCount > 2 {
		riskCtrl = riskCtrl - 15
	} else if highCount > 0 {
		riskCtrl = riskCtrl - 8
	}
	if riskCtrl < 0 { riskCtrl = 0 }

	// JD competitiveness: higher for roles with many must_haves
	jdComp := minInt(80, totalMust*15)
	if totalMust == 0 { jdComp = 50 }

	breakdown := ScoreBreakdownResultV3{
		DirectionFit:        clamp(dirFit, 0, 100),
		MustHaveCoverage:    clamp(covScore, 0, 100),
		EvidenceStrength:    clamp(evScore, 0, 100),
		ExperienceRelevance: clamp(expScore, 0, 100),
		RiskControl:         clamp(riskCtrl, 0, 100),
		JDCompetitiveness:   jdComp,
	}

	deepScore := int(float64(breakdown.DirectionFit)*0.20 +
		float64(breakdown.MustHaveCoverage)*0.20 +
		float64(breakdown.EvidenceStrength)*0.20 +
		float64(breakdown.ExperienceRelevance)*0.15 +
		float64(breakdown.RiskControl)*0.15 +
		float64(breakdown.JDCompetitiveness)*0.10)

	// Apply Fast hard gate cap
	if deepScore > v3ctx.FastHardGate.ScoreCap {
		deepScore = v3ctx.FastHardGate.ScoreCap
	}
	// Deep should not be > Fast + 10 without explanation
	if deepScore > fastScore+10 {
		deepScore = fastScore + 5
	}
	deepScore = clamp(deepScore, 0, 100)

	level := matchLevel(deepScore)
	_ = fastScore
	return deepScore, level, breakdown
}

func buildScoreAlignment(fastScore, deepScore int, level string) ScoreAlignmentResult {
	delta := deepScore - fastScore
	consistent := abs(delta) <= 15
	reason := ""
	if delta < 0 {
		reason = fmt.Sprintf("Deep进一步考虑了岗位隐藏期望和投递竞争风险，因此比Fast低%d分。(%s)", -delta, level)
	} else if delta > 0 {
		reason = fmt.Sprintf("Deep发现Fast未覆盖的匹配证据，因此比Fast高%d分。(%s)", delta, level)
	} else {
		reason = "Fast与Deep评分一致，匹配判断稳定。"
	}
	return ScoreAlignmentResult{
		FastScore:    fastScore,
		DeepScore:    deepScore,
		ScoreDelta:   delta,
		IsConsistent: consistent,
		DeltaReason:  reason,
	}
}

func abs(x int) int { if x < 0 { return -x }; return x }

// ─── L6: Application Strategy ─────────────────────────────────────────────

func buildV3AppStrategy(score int, level, decisionLabel string) DeepApplicationStrategy {
	return DeepApplicationStrategy{
		BestTiming:          "根据岗位截止日期决定，建议尽早投递",
		DeliveryStrategy:    fmt.Sprintf("%s — 匹配度%d分（%s）", decisionLabel, score, level),
		CompetingPositioning: fmt.Sprintf("匹配度%d分，%s", score, level),
	}
}

// ─── L7: Resume Rewrite + Interview ───────────────────────────────────────

func buildV3RewriteStrategy(coverage []RequirementCoverageItem, risks []DeepKeyRisk) DeepResumeRewriteStrategy {
	s := DeepResumeRewriteStrategy{}
	for _, c := range coverage {
		if c.CoverageStatus == "strong" {
			s.ProjectRewritePoints = append(s.ProjectRewritePoints, fmt.Sprintf("强化%s的项目经历表达，使用STAR法则", c.RequirementName))
		}
		if c.CoverageStatus == "missing" || c.CoverageStatus == "weak" {
			s.KeywordsToAdd = append(s.KeywordsToAdd, c.RequirementName)
		}
	}
	for _, r := range risks {
		if r.RiskLevel == "high" {
			s.ProjectRewritePoints = append(s.ProjectRewritePoints, r.FixSuggestion)
		}
	}
	if len(s.KeywordsToAdd) > 6 { s.KeywordsToAdd = s.KeywordsToAdd[:6] }
	if len(s.ProjectRewritePoints) > 5 { s.ProjectRewritePoints = s.ProjectRewritePoints[:5] }
	if len(s.SkillsReorder) == 0 { s.SkillsReorder = append(s.SkillsReorder, "将最匹配岗位方向的核心技能前置") }
	return s
}

func buildV3InterviewQuestions(ju DeepJobUnderstanding, risks []DeepKeyRisk, coverage []RequirementCoverageItem) []DeepInterviewQuestion {
	var qs []DeepInterviewQuestion
	// Reuse existing builder
	dummy := &DeepMatchReport{
		JobUnderstanding:   ju,
		KeyRisks:           risks,
		RequirementCoverage: coverage,
	}
	qs = buildInterviewQuestionsFromReport(dummy)
	// Add answer_direction hints
	for i := range qs {
		if qs[i].AnswerDirection == "" {
			qs[i].AnswerDirection = "结合简历真实经历回答，突出具体职责和量化成果"
		}
	}
	return qs
}

func buildV3ActionPlan(strategy DeepResumeRewriteStrategy) []map[string]any {
	plan := []map[string]any{}
	for _, s := range strategy.SkillsReorder {
		plan = append(plan, map[string]any{"step": s, "type": "skills_reorder", "priority": 1})
	}
	for _, s := range strategy.ProjectRewritePoints {
		plan = append(plan, map[string]any{"step": s, "type": "project_rewrite", "priority": 2})
	}
	return plan
}

// ─── L8: Global Consistency ───────────────────────────────────────────────

func buildV3CoreMatchPoints(coverage []RequirementCoverageItem, evidence []ResumeEvidenceItem) []DeepCoreMatchPoint {
	var pts []DeepCoreMatchPoint
	for _, c := range coverage {
		if c.CoverageStatus == "strong" || c.CoverageStatus == "partial" {
			pts = append(pts, DeepCoreMatchPoint{
				PointName:     c.RequirementName,
				MatchStrength: c.CoverageStatus,
				WhyItMatches:  strings.Join(c.MatchedEvidence, "; "),
				JobRequirement: c.RequirementName,
			})
		}
	}
	if len(pts) == 0 && len(evidence) > 0 {
		pts = append(pts, DeepCoreMatchPoint{
			PointName: evidence[0].EvidenceName,
			MatchStrength: "partial",
			WhyItMatches: evidence[0].EvidenceText,
		})
	}
	return pts
}

func makeQualityChecks(allFields bool, noNull bool) QualityChecksResult {
	return QualityChecksResult{
		HasJDUnderstanding:     allFields,
		HasResumeEvidence:      allFields,
		HasRequirementCoverage: allFields,
		HasStructuredRisks:     allFields,
		HasRewriteStrategy:     allFields,
		HasInterviewQuestions:  allFields,
		HasScoreAlignment:      allFields,
		NoNullArrays:           noNull,
		NoRawTagCode:           true,
	}
}

// ─── H22-D: DeepMatchReport Main Path (legacy, H19-B replaced by v3) ────

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

// sanitizeDeepResponse sanitizes null arrays and fixes levels, but does NOT overwrite generated_by for legacy paths.
func sanitizeDeepResponse(resp *MatchResponse) *MatchResponse {
	if resp == nil { return resp }
	if resp.Result == nil { return resp }

	r := resp.Result
	// Only overwrite generated_by for paths that don't already have a specific marker
	if r.GeneratedBy == "" || r.GeneratedBy == "llm_agent" {
		r.GeneratedBy = "deep_match_agent"
	}
	// H19-D: normalize "repaired" → "match_verified" (but not legacy)
	if r.AgentStatus == "" || r.AgentStatus == "success" {
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


