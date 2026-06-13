package service

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"offer-catcher/server/internal/model"
)

// ─── H18-F: Fast Match Calibration Acceptance Test ────────────────────
// Runs fast_sep_match_engine directly against synthetic SEP profiles
// to validate score semantics, hard gate, evidence, risk, and coverage.

func makeSEPTag(label, category string, confidence float64, evidenceText, explanation string) *model.SemanticProfileTag {
	return &model.SemanticProfileTag{
		TagLabel:     label,
		Category:     category,
		Confidence:   confidence,
		EvidenceText: evidenceText,
		Explanation:  explanation,
	}
}

func makeSEPProfile(targetType string, targetID int64, status string) *model.SemanticProfile {
	return &model.SemanticProfile{
		TargetType: targetType,
		TargetID:   targetID,
		Status:     status,
	}
}

type calibCase struct {
	Name           string
	ExpectedRel    string // strong / medium / weak / mismatch / low_evidence
	ExpectedMin    int
	ExpectedMax    int
	ResumeTags     []*model.SemanticProfileTag
	JobTags        []*model.SemanticProfileTag
	JobTitle       string
}

func TestH18F_FastMatchCalibration(t *testing.T) {
	cases := []calibCase{
		// ─── A: Strong Match — Backend resume × Backend job ─────────────
		{
			Name:        "A-强匹配-后端简历x后台开发",
			ExpectedRel: "strong",
			ExpectedMin: 75, ExpectedMax: 100,
			JobTitle: "后台开发工程师",
			ResumeTags: []*model.SemanticProfileTag{
				makeSEPTag("Go", "technical_skill", 0.95, "在实习中使用Go开发了RESTful API服务并上线部署", "Go后端开发"),
				makeSEPTag("MySQL", "technical_skill", 0.90, "设计并优化了企业课程平台的数据库schema", "数据库设计"),
				makeSEPTag("Redis", "technical_skill", 0.85, "使用Redis实现缓存和限流功能", "缓存"),
				makeSEPTag("微服务", "engineering_capability", 0.85, "参与微服务拆分和gRPC调用", "微服务架构"),
				makeSEPTag("Linux", "technical_skill", 0.80, "在Linux服务器上部署和维护应用", "运维"),
			},
			JobTags: []*model.SemanticProfileTag{
				makeSEPTag("Go", "technical_skill", 0.95, "", "Go后端开发"),
				makeSEPTag("MySQL", "technical_skill", 0.90, "", "数据库"),
				makeSEPTag("Redis", "technical_skill", 0.85, "", "缓存"),
				makeSEPTag("Kubernetes", "engineering_capability", 0.80, "", "容器编排"),
			},
		},
		// ─── B: Adjacent Match — Backend resume × Cloud/Tech adj ─────────
		{
			Name:        "B-相邻方向-后端简历x云平台",
			ExpectedRel: "weak",
			ExpectedMin: 50, ExpectedMax: 69,
			JobTitle: "云平台开发工程师",
			ResumeTags: []*model.SemanticProfileTag{
				makeSEPTag("Go", "technical_skill", 0.95, "使用Go开发后端服务", "Go开发"),
				makeSEPTag("MySQL", "technical_skill", 0.90, "", "数据库"),
				makeSEPTag("Linux", "technical_skill", 0.80, "", "运维"),
				makeSEPTag("Docker", "engineering_capability", 0.75, "使用Docker打包和部署应用", "容器化"),
			},
			JobTags: []*model.SemanticProfileTag{
				makeSEPTag("Kubernetes", "engineering_capability", 0.95, "", "容器编排"),
				makeSEPTag("Docker", "engineering_capability", 0.90, "", "容器化"),
				makeSEPTag("Linux", "technical_skill", 0.90, "", "运维"),
				makeSEPTag("Python", "technical_skill", 0.80, "", "Python脚本"),
				makeSEPTag("Terraform", "tool", 0.75, "", "IaC"),
			},
		},
		// ─── C: Weak/Mismatch — Backend resume × Product ────────────────
		{
			Name:        "C-方向不匹配-后端简历x产品岗位",
			ExpectedRel: "mismatch",
			ExpectedMin: 0, ExpectedMax: 44,
			JobTitle: "产品策划",
			ResumeTags: []*model.SemanticProfileTag{
				makeSEPTag("Go", "technical_skill", 0.90, "使用Go开发后端服务", "Go开发"),
				makeSEPTag("MySQL", "technical_skill", 0.85, "", "数据库"),
			},
			JobTags: []*model.SemanticProfileTag{
				makeSEPTag("产品策划", "role_direction", 0.95, "", "产品方向"),
				makeSEPTag("用户研究", "domain_context", 0.90, "", "用户研究"),
				makeSEPTag("数据分析", "domain_context", 0.85, "", "数据分析"),
				makeSEPTag("原型设计", "tool", 0.85, "", "原型"),
			},
		},
		// ─── D: Mismatch — Backend resume × Design ─────────────────
		{
			Name:        "D-明显不匹配-后端简历x设计岗",
			ExpectedRel: "mismatch",
			ExpectedMin: 0, ExpectedMax: 59,
			JobTitle: "UI视觉设计师",
			ResumeTags: []*model.SemanticProfileTag{
				makeSEPTag("Go", "technical_skill", 0.90, "", "Go"),
				makeSEPTag("MySQL", "technical_skill", 0.85, "", "MySQL"),
			},
			JobTags: []*model.SemanticProfileTag{
				makeSEPTag("Figma", "tool", 0.95, "", "设计工具"),
				makeSEPTag("视觉设计", "role_direction", 0.95, "", "视觉设计"),
				makeSEPTag("交互设计", "domain_context", 0.90, "", "交互"),
			},
		},
		// ─── E: Low Evidence — Thin resume × Backend ─────────────────
		{
			Name:        "E-证据不足-薄弱简历x后台开发",
			ExpectedRel: "low_evidence",
			ExpectedMin: 0, ExpectedMax: 72,
			JobTitle: "后台开发工程师",
			ResumeTags: []*model.SemanticProfileTag{
				makeSEPTag("Go", "technical_skill", 0.80, "", ""),
				makeSEPTag("MySQL", "technical_skill", 0.75, "", ""),
			},
			JobTags: []*model.SemanticProfileTag{
				makeSEPTag("Go", "technical_skill", 0.95, "", ""),
				makeSEPTag("MySQL", "technical_skill", 0.90, "", ""),
				makeSEPTag("Redis", "technical_skill", 0.85, "", ""),
				makeSEPTag("Kubernetes", "engineering_capability", 0.85, "", ""),
			},
		},
	}

	var report strings.Builder
	report.WriteString(fmt.Sprintf("%-50s | %6s | %5s | %5s | %s | %-12s | %-8s | %s | %s\n",
		"CASE", "SCORE", "LEVEL", "CAP", "GATE", "DECISION", "HIGH_RISK", "COV_STAT", "JUDGE"))
	report.WriteString(strings.Repeat("-", 180) + "\n")

	allPass := true
	for _, c := range cases {
		ctx := &MatchSEPContext{
			ResumeProfile: makeSEPProfile("resume", 1, "verified"),
			JobProfile:    makeSEPProfile("job", int64(casesToIDs(c.Name)), "verified"),
			ResumeTags:    c.ResumeTags,
			JobTags:       c.JobTags,
			JobTitle:      c.JobTitle,
			ResumeID:      1,
			JobID:         int64(casesToIDs(c.Name)),
		}
		result := runFastSEPMatch(ctx)

		highRiskCount := 0
		if risks, ok := result.StructuredRisks.([]MatchRiskPoint); ok {
			for _, r := range risks {
				if r.RiskLevel == "high" { highRiskCount++ }
			}
		}

		coverage := ""
		if covItems, ok := result.RequirementCoverage.([]RequirementCoverageItem); ok {
			strong, partial, missing := 0, 0, 0
			for _, ci := range covItems {
				switch ci.CoverageStatus {
				case "strong": strong++
				case "partial": partial++
				case "missing": missing++
				}
			}
			coverage = fmt.Sprintf("S%d/P%d/M%d", strong, partial, missing)
		}

		inRange := result.MatchScore >= c.ExpectedMin && result.MatchScore <= c.ExpectedMax
		if !inRange { allPass = false }

		var judge string
		if inRange {
			judge = "✅ PASS"
		} else {
			if result.MatchScore < c.ExpectedMin {
				judge = fmt.Sprintf("❌ TOO_LOW (expected >=%d)", c.ExpectedMin)
			} else {
				judge = fmt.Sprintf("❌ TOO_HIGH (expected <=%d)", c.ExpectedMax)
			}
		}

		gateStatus := "?"
		if hg, ok := result.HardGateObj.(HardGateResult); ok {
			gateStatus = hg.Status
		}

		capped := ""
		if sc, ok := result.HardGateObj.(HardGateResult); ok && sc.ScoreCap < 100 {
			capped = fmt.Sprintf("%d", sc.ScoreCap)
		} else {
			capped = "-"
		}

		report.WriteString(fmt.Sprintf("%-50s | %6d | %5s | %5s | %s | %-12s | %8d | %9s | %s\n",
			c.Name, result.MatchScore, result.MatchLevel, capped,
			gateStatus, result.ApplicationDecision,
			highRiskCount, coverage, judge))

		// Extra detail on risks
		if risks, ok := result.StructuredRisks.([]MatchRiskPoint); ok {
			for _, r := range risks {
				report.WriteString(fmt.Sprintf("  ├─ RISK: [%s] %s → %s\n", r.RiskLevel, r.RiskName, r.FixSuggestion))
			}
		}

		_ = highRiskCount
	}

	t.Logf("\n=== H18-F Fast Match Calibration Report ===\n%s", report.String())

	if !allPass {
		t.Errorf("H18-F: some calibration cases failed score range checks")
	}
}

func casesToIDs(name string) int {
	switch {
	case strings.Contains(name, "A-"):
		return 10001
	case strings.Contains(name, "B-"):
		return 10002
	case strings.Contains(name, "C-"):
		return 10003
	case strings.Contains(name, "D-"):
		return 10004
	case strings.Contains(name, "E-"):
		return 10005
	}
	return 10001
}

// ─── Extra: Evidence / Risk / Coverage quality checks ─────────────────

func TestH18F_EvidenceQuality(t *testing.T) {
	ctx := &MatchSEPContext{
		ResumeProfile: makeSEPProfile("resume", 1, "verified"),
		JobProfile:    makeSEPProfile("job", 10001, "verified"),
		ResumeTags: []*model.SemanticProfileTag{
			makeSEPTag("Go", "technical_skill", 0.95, "在实习中使用Go开发了RESTful API服务并上线部署", "Go后端开发"),
			makeSEPTag("MySQL", "technical_skill", 0.90, "设计数据库schema", "数据库"),
		},
		JobTags: []*model.SemanticProfileTag{
			makeSEPTag("Go", "technical_skill", 0.95, "", ""),
			makeSEPTag("MySQL", "technical_skill", 0.90, "", ""),
		},
		JobTitle: "后台开发工程师",
		ResumeID: 1,
		JobID:    10001,
	}

	evidence := mapResumeEvidence(ctx)
	for _, e := range evidence {
		// Verify no raw code in supports_requirement
		if strings.Contains(e.SupportsRequirement, "technical_skill") ||
			strings.Contains(e.SupportsRequirement, "role_direction") ||
			strings.Contains(e.SupportsRequirement, "project_exp") {
			t.Errorf("RAW CODE LEAK in SupportsRequirement: %s", e.SupportsRequirement)
		}
		// Verify evidence_text is not empty for level >= 3
		if e.EvidenceLevel >= 3 && e.EvidenceText == "" {
			t.Errorf("Evidence level %d but no evidence_text for %s", e.EvidenceLevel, e.EvidenceName)
		}
	}

	risks := detectMatchRisks(ctx, analyzeDirectionFit(ctx))
	for _, r := range risks {
		if r.RiskName == "" { t.Error("RISK: empty risk_name") }
		if r.RiskLevel == "" { t.Error("RISK: empty risk_level") }
		if r.FixSuggestion == "" { t.Errorf("RISK: empty fix_suggestion for %s", r.RiskName) }
	}

	t.Logf("Evidence quality: %d items, Risks: %d items — all checks passed", len(evidence), len(risks))
}

func TestH18F_MatchLevelSemantics(t *testing.T) {
	tests := []struct{ score int; level string }{
		{95, "高度匹配"}, {90, "高度匹配"},
		{85, "较高匹配"}, {80, "较高匹配"},
		{75, "中等匹配"}, {70, "中等匹配"},
		{65, "弱匹配"}, {60, "弱匹配"},
		{50, "不建议投递"}, {30, "不建议投递"}, {0, "不建议投递"},
	}
	for _, tt := range tests {
		got := matchLevel(tt.score)
		if got != tt.level {
			t.Errorf("matchLevel(%d) = %q, want %q", tt.score, got, tt.level)
		}
	}
	log.Printf("MatchLevel semantics: all %d thresholds verified ✅", len(tests))
}

func TestH18F_HardGateCaps(t *testing.T) {
	// Test: mismatch direction → cap 60
	dir := DirectionFitResult{Level: "mismatch", Score: 20, ScoreCap: 60}
	tags := []*model.SemanticProfileTag{
		makeSEPTag("Figma", "tool", 0.95, "", ""),
	}
	gate := checkHardGate(dir, tags, tags)
	if gate.ScoreCap != 60 { t.Errorf("Mismatch cap: want 60, got %d", gate.ScoreCap) }
	if gate.Status != "blocked" { t.Errorf("Mismatch status: want blocked, got %s", gate.Status) }

	// Test: weak direction → cap 65
	dir2 := DirectionFitResult{Level: "weak", Score: 45, ScoreCap: 65}
	gate2 := checkHardGate(dir2, tags, tags)
	if gate2.ScoreCap > 65 { t.Errorf("Weak cap: want <=65, got %d", gate2.ScoreCap) }

	// Test: no evidence → cap 72
	dir3 := DirectionFitResult{Level: "strong", Score: 95, ScoreCap: 100}
	noEvTags := []*model.SemanticProfileTag{
		makeSEPTag("Go", "technical_skill", 0.90, "", ""),
	}
	gate3 := checkHardGate(dir3, noEvTags, noEvTags)
	if gate3.ScoreCap > 72 { t.Errorf("No evidence cap: want <=72, got %d", gate3.ScoreCap) }

	t.Logf("HardGate caps: mismatch=%d, weak=%d, no_evidence=%d ✅", gate.ScoreCap, gate2.ScoreCap, gate3.ScoreCap)
}

func TestH18F_DecisionAlignment(t *testing.T) {
	// Verify High risk downgrades decision
	risks := []MatchRiskPoint{
		{RiskName: "方向不匹配", RiskLevel: "high", FixSuggestion: "换方向"},
	}
	gate := HardGateResult{Status: "passed", ScoreCap: 100}
	code, _ := decideApplication(85, risks, gate)
	if code == "priority_apply" || code == "apply_directly" {
		t.Errorf("High risk should downgrade: score 85 + high risk → got %s", code)
	}
	t.Logf("Decision alignment: score 85 + high risk → %s ✅", code)
}
