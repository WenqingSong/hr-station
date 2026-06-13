package service

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"offer-catcher/server/internal/model"
)

// ─── H19-D: Deep Match V3 Calibration Acceptance Test ─────────────────

func TestH19D_DeepV3PipelineCalibration(t *testing.T) {
	// Build test data matching H18-F Fast calib samples, with SEP tags
	strongResume := []*model.SemanticProfileTag{
		makeSEPTag("Go", "technical_skill", 0.95, "在实习中使用Go开发了RESTful API服务并上线部署", "Go后端开发"),
		makeSEPTag("MySQL", "technical_skill", 0.90, "设计并优化了企业课程平台的数据库schema", "数据库设计"),
		makeSEPTag("Redis", "technical_skill", 0.85, "使用Redis实现缓存和限流功能", "缓存"),
		makeSEPTag("微服务", "engineering_capability", 0.85, "参与微服务拆分和gRPC调用", "微服务架构"),
		makeSEPTag("Linux", "technical_skill", 0.80, "在Linux服务器上部署和维护应用", "运维"),
	}
	backendJob := []*model.SemanticProfileTag{
		makeSEPTag("Go", "technical_skill", 0.95, "", "Go后端"),
		makeSEPTag("MySQL", "technical_skill", 0.90, "", "数据库"),
		makeSEPTag("Redis", "technical_skill", 0.85, "", "缓存"),
		makeSEPTag("Kubernetes", "engineering_capability", 0.80, "", "容器编排"),
	}
	cloudJob := []*model.SemanticProfileTag{
		makeSEPTag("Kubernetes", "engineering_capability", 0.95, "", "K8s"),
		makeSEPTag("Docker", "engineering_capability", 0.90, "", "容器化"),
		makeSEPTag("Linux", "technical_skill", 0.90, "", "运维"),
		makeSEPTag("Python", "technical_skill", 0.80, "", "脚本"),
		makeSEPTag("Terraform", "tool", 0.75, "", "IaC"),
	}
	productJob := []*model.SemanticProfileTag{
		makeSEPTag("产品策划", "role_direction", 0.95, "", "产品"),
		makeSEPTag("用户研究", "domain_context", 0.90, "", "用研"),
		makeSEPTag("数据分析", "domain_context", 0.85, "", "数据"),
	}
	thinResume := []*model.SemanticProfileTag{
		makeSEPTag("Go", "technical_skill", 0.80, "", ""),
		makeSEPTag("MySQL", "technical_skill", 0.75, "", ""),
	}

	type calibCase struct {
		Name          string
		ExpectedRel   string
		ResumeTags    []*model.SemanticProfileTag
		JobTags       []*model.SemanticProfileTag
		JobTitle      string
		ExpectMatch   bool
		ExpectScoreRange [2]int
	}
	cases := []calibCase{
		{"A-强相关-后端x后台", "strong_related", strongResume, backendJob, "后台开发工程师", true, [2]int{65, 100}},
		{"B-相邻-后端x云平台", "adjacent", strongResume, cloudJob, "云平台开发工程师", true, [2]int{45, 79}},
		{"C-不匹配-后端x产品", "mismatch", strongResume[:2], productJob, "产品策划", false, [2]int{0, 55}},
		{"D-明显不匹配-后端x设计", "mismatch", strongResume[:2], []*model.SemanticProfileTag{
			makeSEPTag("视觉设计", "role_direction", 0.95, "", "设计"),
			makeSEPTag("Figma", "tool", 0.95, "", "工具"),
		}, "UI设计师", false, [2]int{0, 55}},
		{"E-证据不足-薄弱x后台", "low_evidence", thinResume, backendJob, "后台开发工程师", false, [2]int{0, 72}},
	}

	var report strings.Builder
	report.WriteString(fmt.Sprintf("%-40s | %5s/%5s | %-8s | %6s | %6s | %4s | %s | %s\n",
		"CASE", "F_SCORE", "D_SCORE", "LEVEL", "CAP", "DELTA", "ALIGN", "CHECKS", "JUDGE"))
	report.WriteString(strings.Repeat("-", 140) + "\n")

	allPass := true
	for _, c := range cases {
		// ── Fast ──
		fastCtx := &MatchSEPContext{
			ResumeTags: c.ResumeTags, JobTags: c.JobTags, JobTitle: c.JobTitle,
			ResumeID: 1, JobID: 10001,
		}
		fastResult := runFastSEPMatch(fastCtx)

		// ── Deep V3 ──
		v3ctx := &DeepMatchContextV3{
			ResumeID: 1, JobID: 10001,
			ResumeSEPTags: c.ResumeTags,
			JobSEPTags:    c.JobTags,
			FastResult:    fastResult,
			Status:        "ready",
		}
		// Extract Fast structured fields
		if sb, ok := fastResult.ScoreBreakdownObj.(ScoreBreakdownResult); ok {
			v3ctx.FastScoreBreakdown = sb
		}
		if hg, ok := fastResult.HardGateObj.(HardGateResult); ok {
			v3ctx.FastHardGate = hg
		}
		if cov, ok := fastResult.RequirementCoverage.([]RequirementCoverageItem); ok {
			v3ctx.FastRequirementCoverage = cov
		}
		if risks, ok := fastResult.StructuredRisks.([]MatchRiskPoint); ok {
			v3ctx.FastStructuredRisks = risks
		}
		deepReport := runDeepMatchV3(v3ctx)

		delta := deepReport.MatchScore - fastResult.MatchScore
		inRange := deepReport.MatchScore >= c.ExpectScoreRange[0] && deepReport.MatchScore <= c.ExpectScoreRange[1]
		if !inRange { allPass = false }

		judge := ""
		if inRange {
			judge = "✅"
		} else {
			judge = fmt.Sprintf("❌ expected %d-%d", c.ExpectScoreRange[0], c.ExpectScoreRange[1])
		}

		capped := ""
		if v3ctx.FastHardGate.ScoreCap < 100 {
			capped = fmt.Sprintf("%d", v3ctx.FastHardGate.ScoreCap)
		} else {
			capped = "-"
		}

		checksOk := "✅"
		if !deepReport.QualityChecks.HasJDUnderstanding { checksOk = "❌ JD" }
		if !deepReport.QualityChecks.HasResumeEvidence { checksOk = "❌ EV" }
		if !deepReport.QualityChecks.HasScoreAlignment { checksOk = "❌ SA" }

		report.WriteString(fmt.Sprintf("%-40s | %5d/%5d | %-8s | %6s | %6d | %4s | %5s | %s\n",
			c.Name, fastResult.MatchScore, deepReport.MatchScore, deepReport.MatchLevel,
			capped, delta,
			checks(b2s(deepReport.ScoreAlignment.IsConsistent)),
			checksOk, judge))

		// Detailed audit
		log.Printf("[H19D] %s: Fast=%d(%s) Deep=%d(%s) delta=%d gate=%s align=%s",
			c.Name, fastResult.MatchScore, fastResult.MatchLevel,
			deepReport.MatchScore, deepReport.MatchLevel, delta,
			v3ctx.FastHardGate.Status, b2s(deepReport.ScoreAlignment.IsConsistent))

		// Verify Deep-specific quality checks
		if deepReport.MatchScore > 0 && len(deepReport.JobUnderstanding.MustHaveRequirements) == 0 {
			t.Errorf("%s: deep v3 has score but no must_have requirements in job_understanding", c.Name)
		}
		if deepReport.MatchScore > 0 && len(deepReport.JobUnderstanding.ImportantRequirements) == 0 && len(deepReport.JobUnderstanding.BonusRequirements) == 0 {
			t.Errorf("%s: deep v3 job_understanding is too thin", c.Name)
		}

		// Verify score_alignment exists and has reason
		if deepReport.ScoreAlignment.DeltaReason == "" {
			t.Errorf("%s: score_alignment has no delta_reason", c.Name)
		}

		// Verify agent_trace has stages
		if len(deepReport.AgentTrace) < 5 {
			t.Errorf("%s: agent_trace too short: %d stages", c.Name, len(deepReport.AgentTrace))
		}

		// Verify key_risks structure
		for _, r := range deepReport.KeyRisks {
			if r.RiskName == "" { t.Errorf("%s: risk missing risk_name", c.Name) }
			if r.RiskLevel == "" { t.Errorf("%s: risk %s missing risk_level", c.Name, r.RiskName) }
		}
	}

	t.Logf("\n=== H19-D Deep Match V3 Calibration ===\n%s", report.String())

	if !allPass {
		t.Errorf("H19-D: some calibration cases failed score range checks")
	}
}

func TestH19D_DeepV3SchemaCompleteness(t *testing.T) {
	// Verify v3 report always outputs complete schema
	resumeTags := []*model.SemanticProfileTag{
		makeSEPTag("Go", "technical_skill", 0.95, "Go后端实习", "Go开发"),
	}
	jobTags := []*model.SemanticProfileTag{
		makeSEPTag("Go", "technical_skill", 0.95, "", "Go"),
	}
	fastCtx := &MatchSEPContext{ResumeTags: resumeTags, JobTags: jobTags, JobTitle: "后台开发", ResumeID: 1, JobID: 1}
	fastResult := runFastSEPMatch(fastCtx)

	v3ctx := &DeepMatchContextV3{
		ResumeID: 1, JobID: 1, ResumeSEPTags: resumeTags, JobSEPTags: jobTags,
		FastResult: fastResult, Status: "ready",
	}
	if sb, ok := fastResult.ScoreBreakdownObj.(ScoreBreakdownResult); ok { v3ctx.FastScoreBreakdown = sb }
	if hg, ok := fastResult.HardGateObj.(HardGateResult); ok { v3ctx.FastHardGate = hg }

	report := runDeepMatchV3(v3ctx)

	// Check all required fields
	checks := []struct {
		name  string
		empty bool
	}{
		{"report_version", report.ReportVersion == ""},
		{"generated_by", report.GeneratedBy == ""},
		{"score_alignment delta_reason", report.ScoreAlignment.DeltaReason == ""},
		{"job_understanding positioning", report.JobUnderstanding.JobPositioning == ""},
		{"quality_checks has_jd_understanding", !report.QualityChecks.HasJDUnderstanding},
		{"quality_checks has_score_alignment", !report.QualityChecks.HasScoreAlignment},
		{"agent_trace", len(report.AgentTrace) == 0},
		{"decision_label", report.DecisionLabel == ""},
		{"application_decision", report.ApplicationDecision == ""},
		{"one_sentence_conclusion", report.OneSentenceConclusion == ""},
	}
	for _, c := range checks {
		if c.empty {
			t.Errorf("SNHEMA MISSING: %s is empty", c.name)
		}
	}
	t.Logf("Schema completeness: all %d required fields verified", len(checks))
}

func TestH19D_DeepV3ScoreAlignment(t *testing.T) {
	// Verify delta_reason is meaningful
	resumeTags := []*model.SemanticProfileTag{
		makeSEPTag("Go", "technical_skill", 0.95, "Go后端实习并上线部署", "Go"),
	}
	jobTags := []*model.SemanticProfileTag{
		makeSEPTag("Go", "technical_skill", 0.95, "", "Go"),
	}
	fastCtx := &MatchSEPContext{ResumeTags: resumeTags, JobTags: jobTags, JobTitle: "后台开发", ResumeID: 1, JobID: 1}
	fastResult := runFastSEPMatch(fastCtx)

	v3ctx := &DeepMatchContextV3{
		ResumeID: 1, JobID: 1, ResumeSEPTags: resumeTags, JobSEPTags: jobTags,
		FastResult: fastResult, Status: "ready",
	}
	if sb, ok := fastResult.ScoreBreakdownObj.(ScoreBreakdownResult); ok { v3ctx.FastScoreBreakdown = sb }
	if hg, ok := fastResult.HardGateObj.(HardGateResult); ok { v3ctx.FastHardGate = hg }

	report := runDeepMatchV3(v3ctx)
	sa := report.ScoreAlignment

	if sa.FastScore <= 0 { t.Error("FastScore is zero") }
	if sa.DeepScore <= 0 { t.Error("DeepScore is zero") }
	if sa.DeltaReason == "" { t.Error("DeltaReason is empty") }
	if !sa.IsConsistent && sa.FastScore == sa.DeepScore {
		t.Error("is_consistent=false but scores are equal")
	}

	log.Printf("Score Alignment: Fast=%d Deep=%d delta=%d consistent=%v reason=%s",
		sa.FastScore, sa.DeepScore, sa.ScoreDelta, sa.IsConsistent, sa.DeltaReason)
}

func TestH19D_DeepV3ProfilePending(t *testing.T) {
	// Test SEP missing scenario
	v3ctx := &DeepMatchContextV3{
		ResumeID: 1, JobID: 1, Status: "profile_missing",
		Reason: "resume SEP not ready",
	}
	report := runDeepMatchV3(v3ctx)

	if report.GeneratedBy != "deep_match_profile_pending" {
		t.Errorf("SEP pending: generated_by=%s, want deep_match_profile_pending", report.GeneratedBy)
	}
	if report.MatchScore != 0 {
		t.Errorf("SEP pending: match_score=%d, want 0", report.MatchScore)
	}
	if report.AgentStatus != "profile_missing" {
		t.Errorf("SEP pending: agent_status=%s, want profile_missing", report.AgentStatus)
	}
	t.Logf("Profile pending: generated_by=%s, status=%s, score=%d ✅", report.GeneratedBy, report.AgentStatus, report.MatchScore)
}

func b2s(b bool) string { if b { return "true" }; return "false" }
func checks(s string) string { return s }
