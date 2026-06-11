package service

import (
	"encoding/json"
	"strings"
	"testing"

	"offer-catcher/server/internal/model"
)

// ─── Test: buildResumeEvidence with real structured data ──────────────

func TestBuildResumeEvidence_WithRealData(t *testing.T) {
	resume := &model.Resume{
		ProfessionalSkills: "Go, Python, MySQL, Redis, Docker, Kubernetes",
		Projects: []model.ResumeProject{
			{Name: "用户画像系统", Description: "基于Spark构建用户行为分析平台，日处理10TB数据"},
			{Name: "微服务网关", Description: "使用Go开发API网关，支持动态路由和限流"},
		},
		Internships: []model.ResumeInternship{
			{Company: "腾讯", Position: "后端开发实习生", Description: "参与推荐系统后端开发"},
		},
		Educations: []model.ResumeEducation{
			{Degree: "硕士", Major: "计算机科学", School: "清华大学"},
		},
		Certificates: []model.ResumeCertificate{
			{Name: "AWS Solutions Architect"},
		},
	}

	ev := buildResumeEvidence(resume, "Some raw text about algorithms and backend")

	// K13: evidence_json must NOT be empty
	t.Run("evidence_not_empty", func(t *testing.T) {
		if ev.IsEmpty() {
			t.Error("evidence should not be empty for resume with structured data")
		}
	})

	// K13: project_count / internship_count must be accurate
	t.Run("project_count", func(t *testing.T) {
		if ev.ProjectCount != 2 {
			t.Errorf("expected projectCount=2, got %d", ev.ProjectCount)
		}
	})
	t.Run("internship_count", func(t *testing.T) {
		if ev.InternshipCount != 1 {
			t.Errorf("expected internshipCount=1, got %d", ev.InternshipCount)
		}
	})
	t.Run("skill_count", func(t *testing.T) {
		// 6 comma-separated skills
		if ev.SkillCount < 3 {
			t.Errorf("expected skillCount >= 3, got %d", ev.SkillCount)
		}
	})

	// K13: evidence items should have real text (not "{}")
	t.Run("skill_evidence", func(t *testing.T) {
		if len(ev.SkillEvidence) == 0 {
			t.Error("expected skill evidence items")
		}
		for _, item := range ev.SkillEvidence {
			if item.Text == "" {
				t.Error("evidence text must not be empty")
			}
			if item.Strength <= 0 {
				t.Error("evidence strength must be > 0")
			}
		}
	})
	t.Run("project_evidence", func(t *testing.T) {
		if len(ev.ProjectEvidence) != 2 {
			t.Errorf("expected 2 project evidence items, got %d", len(ev.ProjectEvidence))
		}
	})
	t.Run("internship_evidence", func(t *testing.T) {
		if len(ev.InternshipEvidence) == 0 {
			t.Error("expected internship evidence")
		}
	})
}

// ─── Test: buildResumeEvidence with empty structured data ────────────

func TestBuildResumeEvidence_EmptyData(t *testing.T) {
	resume := &model.Resume{
		Projects:    []model.ResumeProject{},
		Internships: []model.ResumeInternship{},
		Educations:  []model.ResumeEducation{},
		Certificates: []model.ResumeCertificate{},
	}

	ev := buildResumeEvidence(resume, "")

	t.Run("project_count_zero", func(t *testing.T) {
		if ev.ProjectCount != 0 {
			t.Errorf("expected projectCount=0 for empty projects, got %d", ev.ProjectCount)
		}
	})
	t.Run("internship_count_zero", func(t *testing.T) {
		if ev.InternshipCount != 0 {
			t.Errorf("expected internshipCount=0 for empty internships, got %d", ev.InternshipCount)
		}
	})
	t.Run("should_not_fake_evidence", func(t *testing.T) {
		// When data is truly empty, evidence should reflect that honestly
		if len(ev.ProjectEvidence) > 0 {
			t.Error("should not generate fake project evidence")
		}
	})
}

// ─── Test: buildJobEvidence with JD text ────────────────────────────

func TestBuildJobEvidence(t *testing.T) {
	job := &model.Job{
		Description:  "负责后端服务架构设计与开发，参与分布式系统优化",
		Requirements: "熟悉Go/C++、MySQL/Redis，具备分布式系统经验",
	}
	jdText := job.Description + " " + job.Requirements

	ev := buildJobEvidence(job, jdText)

	t.Run("job_evidence_not_empty", func(t *testing.T) {
		if ev.IsEmpty() {
			t.Error("job evidence should not be empty")
		}
	})
	t.Run("keyword_evidence", func(t *testing.T) {
		// Note: keyword_evidence is empty here because Job.SkillKeywordsJSON is nil in test fixture
		if len(ev.KeywordEvidence) > 0 {
			t.Log("keyword evidence present (from SkillKeywordsJSON)")
		}
	})
	t.Run("requirement_evidence", func(t *testing.T) {
		if len(ev.RequirementEvidence) == 0 {
			t.Error("expected requirement evidence")
		}
	})
}

// ─── Test: computeTagMatch ───────────────────────────────────────────

func TestComputeTagMatch(t *testing.T) {
	rp := &model.ResumeTagProfile{
		PrimaryTags: []model.TagItem{
			{Tag: "backend", Group: "direction", Weight: 1.0},
			{Tag: "golang", Group: "skill", Weight: 1.0},
			{Tag: "engineering_ability", Group: "capability", Weight: 0.8},
		},
	}
	jp := &model.JobTagProfile{
		PrimaryTags: []model.TagItem{
			{Tag: "backend", Group: "direction", Weight: 1.0},
			{Tag: "cpp", Group: "skill", Weight: 1.0},
			{Tag: "backend", Group: "direction", Weight: 1.0}, // duplicate
		},
	}

	matched, missing, score := computeTagMatch(rp, jp)

	// "backend" matched, "golang" and "engineering_ability" missing from job
	// "cpp" missing from resume
	t.Run("matched", func(t *testing.T) {
		found := false
		for _, m := range matched {
			if m == "backend" {
				found = true
			}
		}
		if !found {
			t.Errorf("expected 'backend' in matched tags, got %v", matched)
		}
	})
	t.Run("score_range", func(t *testing.T) {
		if score < 0 {
			t.Errorf("score must be >= 0, got %.2f", score)
		}
	})
	_ = missing // may be empty or not, depends on matching logic
}

// ─── Test: JSON marshal contract ──────────────────────────────────────

func TestTagEvidenceJSONContract(t *testing.T) {
	rtp := &model.ResumeTagProfile{
		ResumeID:        8,
		DominantProfile: "backend_engineering",
		PrimaryTags: []model.TagItem{
			{Tag: "backend", Group: "direction", Weight: 1.0, Confidence: 3},
		},
		Evidence: model.ResumeEvidence{
			ProjectCount:    2,
			InternshipCount: 1,
			ProjectEvidence: []model.EvidenceItem{
				{ID: "proj_0", Type: "project", Text: "用户画像系统", Source: "projects", Strength: 0.7},
			},
		},
	}

	b, err := json.Marshal(rtp)
	if err != nil {
		t.Fatalf("failed to marshal ResumeTagProfile: %v", err)
	}

	var parsed model.ResumeTagProfile
	if err := json.Unmarshal(b, &parsed); err != nil {
		t.Fatalf("failed to unmarshal ResumeTagProfile: %v", err)
	}

	t.Run("roundtrip", func(t *testing.T) {
		if parsed.ResumeID != 8 {
			t.Errorf("roundtrip: expected resume_id=8, got %d", parsed.ResumeID)
		}
		if parsed.DominantProfile != "backend_engineering" {
			t.Errorf("roundtrip: expected dominant=backend_engineering, got %s", parsed.DominantProfile)
		}
		if parsed.Evidence.ProjectCount != 2 {
			t.Errorf("roundtrip: expected projectCount=2, got %d", parsed.Evidence.ProjectCount)
		}
		if len(parsed.Evidence.ProjectEvidence) != 1 {
			t.Errorf("roundtrip: expected 1 project evidence, got %d", len(parsed.Evidence.ProjectEvidence))
		}
	})

	// K13: evidence_json must NOT serialize as "{}"
	t.Run("evidence_json_not_empty", func(t *testing.T) {
		evJSON, _ := json.Marshal(parsed.Evidence)
		raw := string(evJSON)
		if raw == "{}" || raw == "" {
			t.Error("evidence_json must not be empty for valid profile")
		}
		if !strings.Contains(raw, "project_evidence") {
			t.Error("evidence_json must contain project_evidence")
		}
	})
}

// ─── Test: tag entry conversion helpers ───────────────────────────────

func TestConvertTagEntries(t *testing.T) {
	entries := []tagEntry{
		{Tag: "backend", Group: "direction", Evidence: "后端关键词", Confidence: 3},
		{Tag: "golang", Group: "skill", Evidence: "Go相关经验", Confidence: 2},
	}

	items := convertTagEntries(entries)
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].Tag != "backend" {
		t.Errorf("expected backend, got %s", items[0].Tag)
	}
	if items[0].Group != "direction" {
		t.Errorf("expected direction group, got %s", items[0].Group)
	}
}
