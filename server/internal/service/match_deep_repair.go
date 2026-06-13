// ─── H19-B: Deep Match JSON Repair & Sanitize ────────────────────────────

package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"offer-catcher/server/internal/model"
)

func extractDeepJSON(raw string) string {
	raw = strings.TrimSpace(raw)
	// Strip markdown code fences
	if idx := strings.Index(raw, "```json"); idx >= 0 {
		raw = raw[idx+7:]
		if j := strings.LastIndex(raw, "```"); j >= 0 {
			raw = raw[:j]
		}
	} else if idx := strings.Index(raw, "```"); idx >= 0 {
		raw = raw[idx+3:]
		if j := strings.LastIndex(raw, "```"); j >= 0 {
			raw = raw[:j]
		}
	}
	raw = strings.TrimSpace(raw)
	// Extract first JSON object
	if !strings.HasPrefix(raw, "{") {
		if idx := strings.Index(raw, "{"); idx >= 0 {
			raw = raw[idx:]
		}
	}
	if !strings.HasSuffix(raw, "}") {
		if idx := strings.LastIndex(raw, "}"); idx >= 0 {
			raw = raw[:idx+1]
		}
	}
	return raw
}

func repairDeepMatchJSON(raw string) (string, error) {
	content := extractDeepJSON(raw)
	if content == "" || content == "{}" {
		return "", fmt.Errorf("empty JSON content")
	}
	// Basic repair: unescaped newlines in strings
	content = strings.ReplaceAll(content, "\n", "\\n")
	// Verify it's valid JSON
	var dummy interface{}
	if err := json.Unmarshal([]byte(content), &dummy); err != nil {
		return content, err
	}
	return content, nil
}

func sanitizeDeepMatchReport(report *DeepMatchReport, fastResult *model.MatchAnalysisResult) *DeepMatchReport {
	if report == nil {
		report = &DeepMatchReport{}
	}
	report.GenerationTier = "quality"
	if report.GeneratedBy == "" {
		report.GeneratedBy = "deep_match_agent"
	}
	if report.AgentStatus == "" || report.AgentStatus == "success" {
		report.AgentStatus = "match_verified"
	}
	// Clamp score
	if report.MatchScore < 0 { report.MatchScore = 0 }
	if report.MatchScore > 100 { report.MatchScore = 100 }
	if report.MatchLevel == "" {
		report.MatchLevel = matchLevel(report.MatchScore)
	}
	if report.ApplicationDecision == "" {
		report.ApplicationDecision, report.DecisionLabel = decideApplicationSimple(report.MatchScore)
	}
	if report.DecisionLabel == "" {
		_, report.DecisionLabel = decideApplicationSimple(report.MatchScore)
	}
	if report.OneSentenceConclusion == "" {
		report.OneSentenceConclusion = fmt.Sprintf("深度分析匹配度%d分（%s）", report.MatchScore, report.MatchLevel)
	}
	// Ensure arrays are not nil
	if report.RequirementCoverage == nil { report.RequirementCoverage = []RequirementCoverageItem{} }
	if report.ResumeEvidence == nil { report.ResumeEvidence = []ResumeEvidenceItem{} }
	if report.CoreMatchPoints == nil { report.CoreMatchPoints = []DeepCoreMatchPoint{} }
	if report.KeyRisks == nil { report.KeyRisks = []DeepKeyRisk{} }
	if report.InterviewRiskQuestions == nil { report.InterviewRiskQuestions = []DeepInterviewQuestion{} }
	if report.AgentTrace == nil { report.AgentTrace = []DeepAgentTrace{} }
	if report.JobUnderstanding.CoreResponsibilities == nil { report.JobUnderstanding.CoreResponsibilities = []string{} }
	if report.JobUnderstanding.MustHaveRequirements == nil { report.JobUnderstanding.MustHaveRequirements = []string{} }
	if report.JobUnderstanding.ImportantRequirements == nil { report.JobUnderstanding.ImportantRequirements = []string{} }
	if report.JobUnderstanding.BonusRequirements == nil { report.JobUnderstanding.BonusRequirements = []string{} }
	if report.JobUnderstanding.HiddenExpectations == nil { report.JobUnderstanding.HiddenExpectations = []string{} }
	if report.ResumeRewriteStrategy.SkillsReorder == nil { report.ResumeRewriteStrategy.SkillsReorder = []string{} }
	if report.ResumeRewriteStrategy.ProjectRewritePoints == nil { report.ResumeRewriteStrategy.ProjectRewritePoints = []string{} }
	if report.ResumeRewriteStrategy.InternshipRewritePoints == nil { report.ResumeRewriteStrategy.InternshipRewritePoints = []string{} }
	if report.ResumeRewriteStrategy.KeywordsToAdd == nil { report.ResumeRewriteStrategy.KeywordsToAdd = []string{} }
	if report.ResumeRewriteStrategy.ContentToWeaken == nil { report.ResumeRewriteStrategy.ContentToWeaken = []string{} }
	// Agent trace stages
	if len(report.AgentTrace) == 0 {
		stages := []string{"INPUT_GATE","JD_DECOMPOSITION","RESUME_EVIDENCE_AUDIT","REQUIREMENT_ALIGNMENT","RISK_DIAGNOSIS","APPLICATION_STRATEGY","RESUME_REWRITE_STRATEGY","INTERVIEW_QUESTION_PREDICTION","GLOBAL_CALIBRATION"}
		agents := []string{"输入网关","JD深度解构","简历证据审计","要求对齐","风险诊断","投递策略","简历改写策略","面试追问预测","全局校准"}
		for i, s := range stages {
			report.AgentTrace = append(report.AgentTrace, DeepAgentTrace{Stage: s, Agent: agents[i], Status: "done", Summary: ""})
		}
	}
	// Score delta from Fast
	if fastResult != nil && report.MatchScore > 0 {
		delta := report.MatchScore - fastResult.MatchScore
		if delta > 15 || delta < -15 {
			if report.ScoreDeltaExplanation == nil {
				report.ScoreDeltaExplanation = &DeepScoreDeltaExplanation{
					FastScore: fastResult.MatchScore,
					DeepScore: report.MatchScore,
					Delta:     delta,
					Reason:    fmt.Sprintf("Fast=%d Deep=%d 差异%d分，Deep 基于SEP语义证据重新评估", fastResult.MatchScore, report.MatchScore, delta),
				}
			}
		}
	}
	return report
}

func decideApplicationSimple(score int) (string, string) {
	if score >= 90 { return "priority_apply", "优先投递" }
	if score >= 80 { return "apply_directly", "可以直接投递" }
	if score >= 70 { return "revise_before_apply", "修改后投递" }
	if score >= 60 { return "backup_apply", "作为备选投递" }
	return "not_recommended", "暂不建议投递"
}

// ─── Map Deep Report to legacy model.MatchAnalysisResult ────────────────

func mapDeepReportToLegacy(report *DeepMatchReport) *model.MatchAnalysisResult {
	if report == nil { return &model.MatchAnalysisResult{} }
	score := report.MatchScore
	level := report.MatchLevel
	if level == "" { level = matchLevel(score) }
	fitPoints := []string{}
	for _, mp := range report.CoreMatchPoints { fitPoints = append(fitPoints, mp.PointName) }
	gapPoints := []string{}
	for _, c := range report.RequirementCoverage {
		if c.CoverageStatus == "missing" || c.CoverageStatus == "weak" { gapPoints = append(gapPoints, c.RequirementName) }
	}
	riskPoints := []string{}
	for _, r := range report.KeyRisks { riskPoints = append(riskPoints, r.RiskName) }
	evidenceChain := []model.MatchEvidence{}
	for _, e := range report.ResumeEvidence {
		evidenceChain = append(evidenceChain, model.MatchEvidence{Requirement: e.SupportsRequirement, ResumePoint: e.EvidenceText, Judgement: e.Confidence})
	}
	suggestions := []model.MatchSuggestion{}
	for _, r := range report.KeyRisks {
		suggestions = append(suggestions, model.MatchSuggestion{Priority: "P1", Title: r.RiskName, Reason: r.WhyItMatters, Action: r.FixSuggestion})
	}
	matchHighlights := []string{}
	for _, c := range report.RequirementCoverage {
		if c.CoverageStatus == "strong" || c.CoverageStatus == "partial" { matchHighlights = append(matchHighlights, c.RequirementName) }
	}
	actions := []string{}
	for _, r := range report.KeyRisks { if r.FixSuggestion != "" { actions = append(actions, r.FixSuggestion) } }
	return &model.MatchAnalysisResult{
		MatchScore: score, MatchLevel: level, MatchLabel: level, ScreeningResult: "pass",
		Summary: report.OneSentenceConclusion, EvidenceChain: evidenceChain, MatchHighlights: matchHighlights,
		GapPoints: gapPoints, RiskPoints: riskPoints, OptimizationSuggestions: suggestions, NextSteps: actions,
		GeneratedBy: "deep_match_agent", AgentStatus: report.AgentStatus, RepairApplied: false, DataQuality: "llm",
		FitPoints: fitPoints,
	}
}
