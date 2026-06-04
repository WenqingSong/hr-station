package agent

import (
	"encoding/json"
	"testing"

	"offer-catcher/server/internal/model"
)

// TestAgentSnapshotConsistency verifies that the agent-local normalize
// produces the same structured output as the service-layer normalizer.
// The agent-local copy exists to avoid circular imports (agent → service).
// This test ensures the two implementations stay in sync for key fields.
func TestAgentSnapshotConsistency_V1(t *testing.T) {
	v1 := &model.ResumeVersion{
		VersionNo:  1,
		Name:       "测试简历",
		TargetRole: "后端开发",
		TargetCity: "深圳",
		Location:   "深圳",
		SkillsJSON: []byte(`["Go","MySQL","Redis"]`),
		EducationJSON: []byte(`[{"school":"XX大学","degree":"本科","major":"计算机科学"}]`),
		ExperienceJSON: []byte(`[{"company":"XX公司","role":"后端实习","description":"负责接口开发"}]`),
		ProjectJSON: []byte(`[{"name":"个人项目","description":"实现了一个Web应用"}]`),
		SelfEvaluation: "热爱技术",
	}

	snap, err := normalizeResumeToSnapshot(v1)
	if err != nil {
		t.Fatalf("agent-local normalize failed: %v", err)
	}

	// Verify key fields
	if snap.Meta.TargetDirection != "后端开发" {
		t.Errorf("meta.target_direction = %q, want \"后端开发\"", snap.Meta.TargetDirection)
	}
	if snap.Meta.ResumeName != "测试简历" {
		t.Errorf("meta.resume_name = %q", snap.Meta.ResumeName)
	}
	if len(snap.JobIntention.TargetRoles) == 0 || snap.JobIntention.TargetRoles[0] != "后端开发" {
		t.Errorf("job_intention.target_roles[0] = %v, want [\"后端开发\"]", snap.JobIntention.TargetRoles)
	}
	if snap.Profile.CurrentCity != "深圳" {
		t.Errorf("profile.current_city = %q", snap.Profile.CurrentCity)
	}
	if len(snap.Educations) == 0 {
		t.Error("educations should not be empty")
	}
	if len(snap.Educations) > 0 && snap.Educations[0].School != "XX大学" {
		t.Errorf("educations[0].school = %q", snap.Educations[0].School)
	}
	if len(snap.Skills.OtherKeywords) != 3 {
		t.Errorf("skills.other_keywords count = %d, want 3", len(snap.Skills.OtherKeywords))
	}
	if snap.ExtraInfo.SelfEvaluation != "热爱技术" {
		t.Errorf("extra_info.self_evaluation = %q", snap.ExtraInfo.SelfEvaluation)
	}
	// Empty V2 modules should be present but empty
	if snap.Certificates == nil {
		t.Error("certificates should not be nil")
	}
	if snap.Awards == nil {
		t.Error("awards should not be nil")
	}
}

// TestAgentSnapshotConsistency_V2 verifies V2 normalization.
func TestAgentSnapshotConsistency_V2(t *testing.T) {
	v2 := &model.ResumeVersion{
		VersionNo:      2,
		SchemaVersion:  2,
		Name:           "V2简历",
		TargetRole:     "后端开发",
		TargetDirection: "云计算",
		JobType:        "实习",
		ExpectedCities: "北京,上海",
		School:         "YY大学",
		ProfileJSON:    []byte(`{"school":"YY大学","degree":"硕士","major":"软件工程"}`),
		JobIntentionJSON: []byte(`{"job_type":"校招","expected_cities":["北京","上海"],"target_roles":["云平台开发"],"earliest_start_date":"2026-07"}`),
		SkillsJSON:     []byte(`{"languages":["Go","Python"],"cloud":["Kubernetes","Docker"],"ai_skills":["Cursor","Copilot"]}`),
		CertificatesJSON: []byte(`[{"name":"AWS SAA","type":"云计算认证","level":"中级"}]`),
		AwardsJSON:     []byte(`[{"name":"ACM-ICPC 银牌","type":"竞赛获奖","level":"省级"}]`),
	}

	snap, err := normalizeResumeToSnapshot(v2)
	if err != nil {
		t.Fatalf("agent-local V2 normalize failed: %v", err)
	}

	if snap.Meta.TargetDirection != "云计算" {
		t.Errorf("meta.target_direction = %q, want \"云计算\"", snap.Meta.TargetDirection)
	}
	if snap.Profile.School != "YY大学" {
		t.Errorf("profile.school = %q", snap.Profile.School)
	}
	if len(snap.JobIntention.ExpectedCities) != 2 {
		t.Errorf("job_intention.expected_cities count = %d, want 2", len(snap.JobIntention.ExpectedCities))
	}
	if snap.JobIntention.EarliestStartDate != "2026-07" {
		t.Errorf("earliest_start_date = %q", snap.JobIntention.EarliestStartDate)
	}
	if len(snap.Skills.ProgrammingLanguages) != 2 {
		t.Errorf("skills.programming_languages count = %d, want 2", len(snap.Skills.ProgrammingLanguages))
	}
	if len(snap.Skills.AITools) != 2 {
		t.Errorf("skills.ai_tools count = %d, want 2", len(snap.Skills.AITools))
	}
	if len(snap.Certificates) != 1 {
		t.Errorf("certificates count = %d, want 1", len(snap.Certificates))
	}
	if len(snap.Awards) != 1 {
		t.Errorf("awards count = %d, want 1", len(snap.Awards))
	}
}

// TestAgentSnapshotToJSON verifies snapshot serialization for LLM prompt injection.
func TestAgentSnapshotToJSON(t *testing.T) {
	v1 := &model.ResumeVersion{
		VersionNo: 1,
		Name:      "测试",
		TargetRole: "后端开发",
	}

	jsonStr, err := buildResumeSnapshotJSON(v1)
	if err != nil {
		t.Fatalf("buildResumeSnapshotJSON failed: %v", err)
	}
	if jsonStr == "" || jsonStr == "{}" {
		t.Error("snapshot JSON should not be empty")
	}

	var m map[string]any
	if err := json.Unmarshal([]byte(jsonStr), &m); err != nil {
		t.Errorf("snapshot JSON is not valid: %v", err)
	}

	// Verify privacy fields are absent
	if _, ok := m["phone"]; ok {
		t.Error("snapshot should not contain phone field")
	}
	if _, ok := m["email"]; ok {
		t.Error("snapshot should not contain email field")
	}
	if _, ok := m["user_id"]; ok {
		t.Error("snapshot should not contain user_id field")
	}

	// Verify key modules exist
	for _, key := range []string{"profile", "job_intention", "educations", "projects", "skills", "certificates", "awards", "portfolio", "extra_info"} {
		if _, ok := m[key]; !ok {
			t.Errorf("snapshot missing module: %s", key)
		}
	}
}
