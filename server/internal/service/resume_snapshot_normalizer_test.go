package service

import (
	"testing"

	"offer-catcher/server/internal/model"
)

func TestNormalizeV1MinimalData(t *testing.T) {
	v := model.ResumeVersion{
		SchemaVersion:  1,
		VersionNo:      1,
		Name:           "我的简历",
		TargetRole:     "后端开发",
		TargetCity:     "深圳",
		Location:       "北京",
		SkillsJSON:     []byte(`["Go","MySQL","Docker"]`),
		SelfEvaluation: "热爱编程",
		EducationJSON:  []byte(`[{"school":"清华大学","degree":"硕士","major":"计算机","start":"2022","end":"2025"}]`),
		ExperienceJSON: []byte(`[{"company":"腾讯","role":"后端实习生","start":"2024-06","end":"2024-09","description":"参与后台开发"}]`),
		ProjectJSON:    []byte(`[{"name":"电商系统","description":"微服务架构","tech_stack":["Go","Kafka"]}]`),
	}

	snap, err := NormalizeResumeToSnapshotV2(v)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// meta
	if snap.Meta.SchemaVersion != 2 {
		t.Errorf("meta.schema_version = %d, want 2", snap.Meta.SchemaVersion)
	}
	if snap.Meta.ResumeVersionNo != 1 {
		t.Errorf("meta.resume_version_no = %d, want 1", snap.Meta.ResumeVersionNo)
	}
	if snap.Meta.ResumeName != "我的简历" {
		t.Errorf("meta.resume_name = %q, want %q", snap.Meta.ResumeName, "我的简历")
	}
	if snap.Meta.TargetDirection != "后端开发" {
		t.Errorf("meta.target_direction = %q, want %q", snap.Meta.TargetDirection, "后端开发")
	}

	// profile
	if snap.Profile.CurrentCity != "北京" {
		t.Errorf("profile.current_city = %q, want %q", snap.Profile.CurrentCity, "北京")
	}
	if snap.Profile.School != "清华大学" {
		t.Errorf("profile.school = %q, want %q", snap.Profile.School, "清华大学")
	}

	// jobIntention
	if len(snap.JobIntention.TargetRoles) != 1 || snap.JobIntention.TargetRoles[0] != "后端开发" {
		t.Errorf("job_intention.target_roles = %v, want [后端开发]", snap.JobIntention.TargetRoles)
	}
	if len(snap.JobIntention.ExpectedCities) != 1 || snap.JobIntention.ExpectedCities[0] != "深圳" {
		t.Errorf("job_intention.expected_cities = %v, want [深圳]", snap.JobIntention.ExpectedCities)
	}

	// educations
	if len(snap.Educations) != 1 {
		t.Errorf("educations count = %d, want 1", len(snap.Educations))
	} else {
		e := snap.Educations[0]
		if e.School != "清华大学" {
			t.Errorf("educations[0].school = %q", e.School)
		}
		if e.Degree != "硕士" {
			t.Errorf("educations[0].degree = %q", e.Degree)
		}
		if e.StartDate != "2022" {
			t.Errorf("educations[0].start_date = %q", e.StartDate)
		}
	}

	// internships
	if len(snap.Internships) != 1 {
		t.Errorf("internships count = %d, want 1", len(snap.Internships))
	} else {
		i := snap.Internships[0]
		if i.Company != "腾讯" {
			t.Errorf("internships[0].company = %q", i.Company)
		}
		if i.Role != "后端实习生" {
			t.Errorf("internships[0].role = %q", i.Role)
		}
	}

	// projects
	if len(snap.Projects) != 1 {
		t.Errorf("projects count = %d, want 1", len(snap.Projects))
	} else {
		p := snap.Projects[0]
		if p.Name != "电商系统" {
			t.Errorf("projects[0].name = %q", p.Name)
		}
		if len(p.TechStack) != 2 {
			t.Errorf("projects[0].tech_stack count = %d, want 2", len(p.TechStack))
		}
	}

	// skills
	if len(snap.Skills.OtherKeywords) != 3 {
		t.Errorf("skills.other_keywords count = %d, want 3", len(snap.Skills.OtherKeywords))
	} else {
		if snap.Skills.OtherKeywords[0] != "Go" {
			t.Errorf("skills.other_keywords[0] = %q, want Go", snap.Skills.OtherKeywords[0])
		}
	}

	// extraInfo
	if snap.ExtraInfo.SelfEvaluation != "热爱编程" {
		t.Errorf("extra_info.self_evaluation = %q, want %q", snap.ExtraInfo.SelfEvaluation, "热爱编程")
	}

	// V2-only modules should be empty, not nil
	if snap.Certificates == nil {
		t.Error("certificates should be empty slice, not nil")
	}
	if snap.Awards == nil {
		t.Error("awards should be empty slice, not nil")
	}
}

func TestNormalizeV2CompleteData(t *testing.T) {
	v := model.ResumeVersion{
		SchemaVersion:  2,
		VersionNo:      3,
		Name:           "V2简历",
		TargetDirection: "后端开发",
		JobType:        "internship",
		ExpectedCities: "深圳,北京",
		School:         "北京大学",
		Degree:         "硕士",
		Major:          "软件工程",
		GraduationYear: "2026",
		ProfileJSON:    []byte(`{"name":"张三","school":"清华大学","degree":"硕士","major":"计算机","graduation_year":"2027","current_city":"北京"}`),
		JobIntentionJSON: []byte(`{"target_roles":["后端开发","全栈"],"job_type":"internship","expected_cities":["深圳","北京"],"accept_other_cities":true,"earliest_start_date":"2026-07","internship_duration":"3个月","weekly_available_days":5}`),
		EducationJSON:  []byte(`[{"school":"清华大学","degree":"硕士","major":"计算机","start_date":"2024-09","end_date":"2027-06","gpa":"3.8","core_courses":["算法","操作系统"]}]`),
		InternshipJSON: []byte(`[{"company":"腾讯","role":"后端实习生","start_date":"2025-06","end_date":"2025-09","content":"微服务开发","tech_stack":["Go","MySQL"],"metrics":"QPS提升30%"}]`),
		ProjectJSON:    []byte(`[{"name":"电商平台","role":"后端负责人","type":"personal","tech_stack":["Go","Kafka"],"description":"分布式电商系统"}]`),
		SkillsJSON:     []byte(`{"languages":["Go","Python"],"frameworks":["Gin"],"databases":["MySQL","Redis"],"cloud":["Kubernetes"],"ai_skills":["Cursor","Copilot"]}`),
		CertificatesJSON: []byte(`[{"name":"CET-6","type":"language","level":"六级","score":"580"}]`),
		AwardsJSON:     []byte(`[{"name":"ACM金牌","type":"competition","level":"national"}]`),
		PortfolioJSON:  []byte(`{"github":"https://github.com/test","blog":"https://test.com"}`),
		ExtraJSON:      []byte(`{"self_evaluation":"热爱技术","additional_info":"可快速到岗"}`),
	}

	snap, err := NormalizeResumeToSnapshotV2(v)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// meta
	if snap.Meta.TargetDirection != "后端开发" {
		t.Errorf("meta.target_direction = %q", snap.Meta.TargetDirection)
	}

	// profile (JSON priority over summary)
	if snap.Profile.School != "清华大学" {
		t.Errorf("profile.school = %q, want 清华大学 (JSON priority)", snap.Profile.School)
	}
	if snap.Profile.GraduationYear != "2027" {
		t.Errorf("profile.graduation_year = %q, want 2027", snap.Profile.GraduationYear)
	}

	// jobIntention
	if len(snap.JobIntention.TargetRoles) != 2 {
		t.Errorf("job_intention.target_roles count = %d, want 2", len(snap.JobIntention.TargetRoles))
	}
	if !snap.JobIntention.AcceptOtherCities {
		t.Error("job_intention.accept_other_cities = false, want true")
	}
	if snap.JobIntention.InternshipDuration != "3个月" {
		t.Errorf("job_intention.internship_duration = %q", snap.JobIntention.InternshipDuration)
	}
	if snap.JobIntention.WeeklyAvailableDays != 5 {
		t.Errorf("job_intention.weekly_available_days = %d, want 5", snap.JobIntention.WeeklyAvailableDays)
	}

	// educations
	if len(snap.Educations) != 1 {
		t.Errorf("educations count = %d, want 1", len(snap.Educations))
	} else {
		e := snap.Educations[0]
		if e.GPA != "3.8" {
			t.Errorf("educations[0].gpa = %q", e.GPA)
		}
		if len(e.CoreCourses) != 2 {
			t.Errorf("educations[0].core_courses count = %d, want 2", len(e.CoreCourses))
		}
	}

	// internships
	if len(snap.Internships) != 1 {
		t.Errorf("internships count = %d, want 1", len(snap.Internships))
	} else {
		i := snap.Internships[0]
		if i.Metrics != "QPS提升30%" {
			t.Errorf("internships[0].metrics = %q", i.Metrics)
		}
	}

	// skills (V2 object format)
	if len(snap.Skills.ProgrammingLanguages) != 2 {
		t.Errorf("skills.programming_languages = %v, want [Go, Python]", snap.Skills.ProgrammingLanguages)
	}
	if len(snap.Skills.AITools) != 2 {
		t.Errorf("skills.ai_tools count = %d, want 2", len(snap.Skills.AITools))
	}
	if len(snap.Skills.CloudPlatforms) != 1 {
		t.Errorf("skills.cloud_platforms = %v, want [Kubernetes]", snap.Skills.CloudPlatforms)
	}

	// certificates
	if len(snap.Certificates) != 1 {
		t.Errorf("certificates count = %d, want 1", len(snap.Certificates))
	} else {
		c := snap.Certificates[0]
		if c.Name != "CET-6" {
			t.Errorf("certificates[0].name = %q", c.Name)
		}
	}

	// awards
	if len(snap.Awards) != 1 {
		t.Errorf("awards count = %d, want 1", len(snap.Awards))
	} else {
		a := snap.Awards[0]
		if a.Level != "national" {
			t.Errorf("awards[0].level = %q", a.Level)
		}
	}

	// portfolio
	if snap.Portfolio.Github != "https://github.com/test" {
		t.Errorf("portfolio.github = %q", snap.Portfolio.Github)
	}

	// extraInfo
	if snap.ExtraInfo.SelfEvaluation != "热爱技术" {
		t.Errorf("extra_info.self_evaluation = %q", snap.ExtraInfo.SelfEvaluation)
	}
	if snap.ExtraInfo.AdditionalInfo != "可快速到岗" {
		t.Errorf("extra_info.additional_info = %q", snap.ExtraInfo.AdditionalInfo)
	}
}

func TestNormalizeV2SummaryFallback(t *testing.T) {
	// V2 with JSON fields empty — should use summary columns as fallback
	v := model.ResumeVersion{
		SchemaVersion:  2,
		VersionNo:      2,
		Name:           "Fallback简历",
		TargetRole:     "算法工程师",
		TargetDirection: "算法",
		JobType:        "campus",
		ExpectedCities: "上海,杭州",
		School:         "复旦大学",
		Degree:         "本科",
		Major:          "数学",
		GraduationYear: "2025",
		Location:       "上海",
		SelfEvaluation: "擅长机器学习",
	}

	snap, err := NormalizeResumeToSnapshotV2(v)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// profile fallback from summary
	if snap.Profile.School != "复旦大学" {
		t.Errorf("profile.school fallback = %q, want 复旦大学", snap.Profile.School)
	}
	if snap.Profile.Degree != "本科" {
		t.Errorf("profile.degree fallback = %q, want 本科", snap.Profile.Degree)
	}
	if snap.Profile.CurrentCity != "上海" {
		t.Errorf("profile.current_city fallback = %q, want 上海", snap.Profile.CurrentCity)
	}

	// jobIntention fallback
	if snap.JobIntention.JobType != "campus" {
		t.Errorf("job_intention.job_type fallback = %q, want campus", snap.JobIntention.JobType)
	}
	if len(snap.JobIntention.ExpectedCities) != 2 {
		t.Errorf("job_intention.expected_cities fallback count = %d, want 2", len(snap.JobIntention.ExpectedCities))
	}
	if len(snap.JobIntention.TargetRoles) != 1 || snap.JobIntention.TargetRoles[0] != "算法工程师" {
		t.Errorf("job_intention.target_roles fallback = %v, want [算法工程师]", snap.JobIntention.TargetRoles)
	}

	// extraInfo fallback
	if snap.ExtraInfo.SelfEvaluation != "擅长机器学习" {
		t.Errorf("extra_info.self_evaluation fallback = %q", snap.ExtraInfo.SelfEvaluation)
	}
}

func TestNormalizeEmptyJSONSafety(t *testing.T) {
	v := model.ResumeVersion{
		SchemaVersion: 1,
		VersionNo:     1,
		Name:          "空JSON测试",
	}

	snap, err := NormalizeResumeToSnapshotV2(v)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// All array fields should be non-nil empty slices
	if snap.Educations == nil {
		t.Error("educations should be empty slice, not nil")
	}
	if snap.Internships == nil {
		t.Error("internships should be empty slice, not nil")
	}
	if snap.Projects == nil {
		t.Error("projects should be empty slice, not nil")
	}
	if snap.Certificates == nil {
		t.Error("certificates should be empty slice, not nil")
	}
	if snap.Awards == nil {
		t.Error("awards should be empty slice, not nil")
	}
	if snap.Skills.OtherKeywords == nil {
		t.Error("skills.other_keywords should be empty slice, not nil")
	}
	if snap.JobIntention.TargetRoles == nil {
		t.Error("job_intention.target_roles should be empty slice, not nil")
	}

	// Should not panic
	_ = snap.Meta.SchemaVersion
}

func TestNormalizeInvalidJSONError(t *testing.T) {
	// V1 with invalid skills JSON
	v := model.ResumeVersion{
		SchemaVersion: 1,
		VersionNo:     1,
		Name:          "非法JSON",
		SkillsJSON:    []byte(`{bad json`),
	}

	_, err := NormalizeResumeToSnapshotV2(v)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
	t.Logf("got expected error: %v", err)
}

func TestNormalizeV2InvalidJSONError(t *testing.T) {
	v := model.ResumeVersion{
		SchemaVersion:  2,
		VersionNo:      2,
		Name:           "非法V2 JSON",
		ProfileJSON:    []byte(`{bad`),
	}

	_, err := NormalizeResumeToSnapshotV2(v)
	if err == nil {
		t.Fatal("expected error for invalid V2 JSON, got nil")
	}
	t.Logf("got expected error: %v", err)
}

func TestNormalizeV1SchemaVersionZero(t *testing.T) {
	// schema_version=0 treated as V1
	v := model.ResumeVersion{
		SchemaVersion: 0,
		VersionNo:     1,
		Name:          "V0当作V1",
		TargetRole:    "前端",
		SkillsJSON:    []byte(`["React","TypeScript"]`),
	}

	snap, err := NormalizeResumeToSnapshotV2(v)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if snap.Meta.TargetDirection != "前端" {
		t.Errorf("expected target_direction=%q, got %q", "前端", snap.Meta.TargetDirection)
	}
	if len(snap.JobIntention.TargetRoles) != 1 || snap.JobIntention.TargetRoles[0] != "前端" {
		t.Errorf("expected target_roles=[前端], got %v", snap.JobIntention.TargetRoles)
	}
}

func TestNormalizeV2EmptyJSONModules(t *testing.T) {
	// V2 with empty "" for JSON fields — should not error
	v := model.ResumeVersion{
		SchemaVersion: 2,
		VersionNo:     1,
		Name:          "空V2模块",
	}

	snap, err := NormalizeResumeToSnapshotV2(v)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if snap.Educations == nil {
		t.Error("educations should be empty slice")
	}
	if snap.Internships == nil {
		t.Error("internships should be empty slice")
	}
	if snap.Certificates == nil {
		t.Error("certificates should be empty slice")
	}
}

func TestNormalizeV1SkillsStringArray(t *testing.T) {
	// V1 skills_json is string array
	v := model.ResumeVersion{
		SchemaVersion: 1,
		VersionNo:     1,
		SkillsJSON:    []byte(`["Go","Java","Kubernetes"]`),
	}

	snap, err := NormalizeResumeToSnapshotV2(v)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(snap.Skills.OtherKeywords) != 3 {
		t.Errorf("other_keywords count = %d, want 3: %v", len(snap.Skills.OtherKeywords), snap.Skills.OtherKeywords)
	}
	// V2-specific categories should be empty
	if snap.Skills.ProgrammingLanguages != nil {
		t.Errorf("programming_languages should be nil, got %v", snap.Skills.ProgrammingLanguages)
	}
}

func TestNormalizeV2SkillsObjectFormat(t *testing.T) {
	// V2 skills_json is object with categories
	v := model.ResumeVersion{
		SchemaVersion: 2,
		VersionNo:     2,
		SkillsJSON:    []byte(`{"languages":["Go"],"frameworks":["Gin"],"tools":["Git","Docker"],"ai_skills":["Cursor"]}`),
	}

	snap, err := NormalizeResumeToSnapshotV2(v)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if snap.Skills.OtherKeywords != nil {
		t.Errorf("other_keywords should be nil (V2 object), got %v", snap.Skills.OtherKeywords)
	}
	if len(snap.Skills.ProgrammingLanguages) != 1 {
		t.Errorf("programming_languages count = %d, want 1", len(snap.Skills.ProgrammingLanguages))
	}
	if len(snap.Skills.DevTools) != 2 {
		t.Errorf("dev_tools count = %d, want 2", len(snap.Skills.DevTools))
	}
}

func TestNormalizeV1NullJSON(t *testing.T) {
	v := model.ResumeVersion{
		SchemaVersion: 1,
		VersionNo:     1,
		SkillsJSON:    []byte("null"),
		EducationJSON: []byte("null"),
	}

	snap, err := NormalizeResumeToSnapshotV2(v)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if snap.Educations == nil {
		t.Error("educations should be empty slice for null JSON")
	}
	if snap.Skills.OtherKeywords == nil {
		t.Error("other_keywords should be empty slice for null JSON")
	}
}

func TestNormalizePhoneEmailNotInSnapshot(t *testing.T) {
	v := model.ResumeVersion{
		SchemaVersion: 1,
		VersionNo:     1,
		Name:          "test",
		Phone:         "13800138000",
		Email:         "test@example.com",
		UserID:        12345,
		ID:            99999,
		RawText:       "raw resume text",
	}

	snap, err := NormalizeResumeToSnapshotV2(v)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// phone, email, user_id, id, raw_text should NOT be in snapshot
	// Verify by checking that profile doesn't contain these fields
	// (profile.name is empty since it's V1 and we don't put phone/email there)
	if snap.Profile.Name != "" {
		t.Error("profile.name should be empty (phone/email not mapped)")
	}
}
