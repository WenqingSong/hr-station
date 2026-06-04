package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"offer-catcher/server/internal/model"
)

// NormalizeResumeToSnapshotV2 converts a V1 or V2 ResumeVersion into a unified
// ResumeSnapshotV2 that all downstream agents consume.
//
// - schema_version <= 1 → normalizes V1 legacy data into V2 snapshot.
// - schema_version == 2 → normalizes V2 module data into V2 snapshot.
// - Empty / null JSON fields are safely handled (default to empty arrays/objects).
// - Invalid JSON returns an error, never panics.
// - phone / email / user_id / id / raw_text are NOT included in the snapshot.
func NormalizeResumeToSnapshotV2(version model.ResumeVersion) (model.ResumeSnapshotV2, error) {
	snapshot := model.ResumeSnapshotV2{
		Meta: model.ResumeSnapshotMeta{
			SchemaVersion:   2, // always output V2
			ResumeVersionNo: version.VersionNo,
			ResumeName:      version.Name,
			TargetDirection: version.TargetRole,
		},
	}

	isV2 := version.SchemaVersion >= 2

	if isV2 {
		if err := normalizeV2ToSnapshot(&snapshot, version); err != nil {
			return snapshot, err
		}
	} else {
		if err := normalizeV1ToSnapshot(&snapshot, version); err != nil {
			return snapshot, err
		}
	}

	return snapshot, nil
}

// normalizeV1ToSnapshot fills ResumeSnapshotV2 from V1 legacy resume fields.
func normalizeV1ToSnapshot(snapshot *model.ResumeSnapshotV2, v model.ResumeVersion) error {
	// meta
	snapshot.Meta.TargetDirection = v.TargetRole
	if v.Name != "" {
		snapshot.Meta.ResumeName = v.Name
	}

	// profile: location → currentCity
	snapshot.Profile = model.ResumeSnapshotProfile{
		CurrentCity: v.Location,
	}

	// jobIntention: targetRole → targetRoles[0], targetCity → expectedCities[0]
	snapshot.JobIntention = model.ResumeSnapshotJobIntention{
		TargetRoles:    make([]string, 0),
		ExpectedCities: make([]string, 0),
	}
	if v.TargetRole != "" {
		snapshot.JobIntention.TargetRoles = append(snapshot.JobIntention.TargetRoles, v.TargetRole)
	}
	if v.TargetCity != "" {
		snapshot.JobIntention.ExpectedCities = append(snapshot.JobIntention.ExpectedCities, v.TargetCity)
	}

	// educations: parse V1 education_json (array of objects)
	educations, err := decodeJSONArray[map[string]any](bytesToString(v.EducationJSON))
	if err != nil {
		return fmt.Errorf("v1 education_json: %w", err)
	}
	snapshot.Educations = make([]model.ResumeSnapshotEducation, 0, len(educations))
	for _, e := range educations {
		edu := model.ResumeSnapshotEducation{
			School:            toString(e["school"]),
			Degree:            toString(e["degree"]),
			Major:             toString(e["major"]),
			Department:        toString(e["department"]),
			StartDate:         toString(e["start"]),
			EndDate:           toString(e["end"]),
			GPA:               toString(e["gpa"]),
			GPAScale:          toString(e["gpa_scale"]),
			Rank:              toString(e["ranking"]),
			ResearchDirection: toString(e["research"]),
			Lab:               toString(e["lab"]),
			PaperSummary:      toString(e["papers"]),
		}

		// courses → coreCourses (string array)
		if courses := toSlice(e["courses"]); len(courses) > 0 {
			edu.CoreCourses = courses
		} else if courses := toSlice(e["core_courses"]); len(courses) > 0 {
			edu.CoreCourses = courses
		}
		snapshot.Educations = append(snapshot.Educations, edu)
	}

	// Also set profile summary from first education
	if len(snapshot.Educations) > 0 {
		first := snapshot.Educations[0]
		if snapshot.Profile.School == "" {
			snapshot.Profile.School = first.School
		}
		if snapshot.Profile.Degree == "" {
			snapshot.Profile.Degree = first.Degree
		}
		if snapshot.Profile.Major == "" {
			snapshot.Profile.Major = first.Major
		}
	}

	// internships: parse V1 experience_json (array of objects)
	experiences, err := decodeJSONArray[map[string]any](bytesToString(v.ExperienceJSON))
	if err != nil {
		return fmt.Errorf("v1 experience_json: %w", err)
	}
	snapshot.Internships = make([]model.ResumeSnapshotInternship, 0, len(experiences))
	for _, exp := range experiences {
		intern := model.ResumeSnapshotInternship{
			Company:     toString(exp["company"]),
			Role:        toString(exp["role"]),
			StartDate:   toString(exp["start"]),
			EndDate:     toString(exp["end"]),
			WorkContent: toString(exp["description"]),
			Metrics:     toString(exp["metrics"]),
		}
		if ts := toSlice(exp["tech_stack"]); len(ts) > 0 {
			intern.TechStack = ts
		}
		if ach := toSlice(exp["achievements"]); len(ach) > 0 {
			intern.Achievements = ach
		}
		snapshot.Internships = append(snapshot.Internships, intern)
	}

	// projects: parse V1 project_json (array of objects)
	projects, err := decodeJSONArray[map[string]any](bytesToString(v.ProjectJSON))
	if err != nil {
		return fmt.Errorf("v1 project_json: %w", err)
	}
	snapshot.Projects = make([]model.ResumeSnapshotProject, 0, len(projects))
	for _, p := range projects {
		proj := model.ResumeSnapshotProject{
			Name:             toString(p["name"]),
			Role:             toString(p["role"]),
			Description:      toString(p["description"]),
			Responsibilities: toString(p["responsibilities"]),
			Link:             toString(p["link"]),
			Type:             toString(p["type"]),
			StartDate:        toString(p["start"]),
			EndDate:          toString(p["end"]),
		}
		if ts := toSlice(p["tech_stack"]); len(ts) > 0 {
			proj.TechStack = ts
		}
		if ach := toSlice(p["achievements"]); len(ach) > 0 {
			proj.Achievements = ach
		}
		snapshot.Projects = append(snapshot.Projects, proj)
	}

	// skills: V1 skills_json is a string array → otherKeywords
	skillsRaw := bytesToString(v.SkillsJSON)
	// Try as string array first
	skillStrs, arrErr := decodeJSONArray[string](skillsRaw)
	if arrErr != nil {
		return fmt.Errorf("v1 skills_json: %w", arrErr)
	}
	snapshot.Skills = model.ResumeSnapshotSkills{
		OtherKeywords: skillStrs,
	}

	// extraInfo: self_evaluation
	snapshot.ExtraInfo = model.ResumeSnapshotExtraInfo{
		SelfEvaluation: v.SelfEvaluation,
	}

	// V2-only modules → empty defaults
	snapshot.Certificates = make([]model.ResumeSnapshotCertificate, 0)
	snapshot.Awards = make([]model.ResumeSnapshotAward, 0)
	snapshot.Portfolio = model.ResumeSnapshotPortfolio{}

	return nil
}

// normalizeV2ToSnapshot fills ResumeSnapshotV2 from V2 module data.
// V2 JSON fields take priority; summary columns serve as fallback.
func normalizeV2ToSnapshot(snapshot *model.ResumeSnapshotV2, v model.ResumeVersion) error {
	// meta: use summary columns
	snapshot.Meta.TargetDirection = v.TargetDirection
	if v.Name != "" {
		snapshot.Meta.ResumeName = v.Name
	}

	// profile: parse ProfileJSON, fallback to summary columns
	profile, err := decodeJSONObject(bytesToString(v.ProfileJSON), map[string]any{})
	if err != nil {
		return fmt.Errorf("v2 profile_json: %w", err)
	}
	snapshot.Profile = model.ResumeSnapshotProfile{
		Name:           toString(profile["name"]),
		School:         toString(profile["school"]),
		Degree:         toString(profile["degree"]),
		Major:          toString(profile["major"]),
		GraduationYear: toString(profile["graduation_year"]),
		CurrentCity:    toString(profile["current_city"]),
	}
	// fallback from summary columns
	if snapshot.Profile.School == "" {
		snapshot.Profile.School = v.School
	}
	if snapshot.Profile.Degree == "" {
		snapshot.Profile.Degree = v.Degree
	}
	if snapshot.Profile.Major == "" {
		snapshot.Profile.Major = v.Major
	}
	if snapshot.Profile.GraduationYear == "" {
		snapshot.Profile.GraduationYear = v.GraduationYear
	}
	if snapshot.Profile.CurrentCity == "" {
		snapshot.Profile.CurrentCity = v.Location
	}

	// jobIntention: parse JobIntentionJSON, fallback to summary columns
	ji, err := decodeJSONObject(bytesToString(v.JobIntentionJSON), map[string]any{})
	if err != nil {
		return fmt.Errorf("v2 job_intention_json: %w", err)
	}
	snapshot.JobIntention = model.ResumeSnapshotJobIntention{
		TargetRoles:        toSlice(ji["target_roles"]),
		JobType:            toString(ji["job_type"]),
		ExpectedCities:     toSlice(ji["expected_cities"]),
		AcceptOtherCities:  toBool(ji["accept_other_cities"]),
		EarliestStartDate:  toString(ji["earliest_start_date"]),
		InternshipDuration: toString(ji["internship_duration"]),
		TargetIndustries:   toSlice(ji["target_industries"]),
		TargetCompanies:    toSlice(ji["target_companies"]),
	}
	if wad, ok := ji["weekly_available_days"]; ok {
		snapshot.JobIntention.WeeklyAvailableDays = toInt(wad)
	}
	// fallback from summary columns
	if snapshot.JobIntention.JobType == "" {
		snapshot.JobIntention.JobType = v.JobType
	}
	if len(snapshot.JobIntention.ExpectedCities) == 0 && v.ExpectedCities != "" {
		parts := strings.Split(v.ExpectedCities, ",")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				snapshot.JobIntention.ExpectedCities = append(snapshot.JobIntention.ExpectedCities, p)
			}
		}
	}
	if len(snapshot.JobIntention.TargetRoles) == 0 && v.TargetRole != "" {
		snapshot.JobIntention.TargetRoles = append(snapshot.JobIntention.TargetRoles, v.TargetRole)
	}

	// educations: parse EducationJSON (V2 extended format)
	educations, err := decodeJSONArray[map[string]any](bytesToString(v.EducationJSON))
	if err != nil {
		return fmt.Errorf("v2 education_json: %w", err)
	}
	snapshot.Educations = make([]model.ResumeSnapshotEducation, 0, len(educations))
	for _, e := range educations {
		edu := model.ResumeSnapshotEducation{
			School:            toString(e["school"]),
			Degree:            toString(e["degree"]),
			Major:             toString(e["major"]),
			Department:        toString(e["department"]),
			StartDate:         toString(e["start_date"]),
			EndDate:           toString(e["end_date"]),
			GPA:               toString(e["gpa"]),
			GPAScale:          toString(e["gpa_scale"]),
			Rank:              toString(e["ranking"]),
			ResearchDirection: toString(e["research_direction"]),
			Lab:               toString(e["advisor"]),
			PaperSummary:      toString(e["papers"]),
		}
		if courses := toSlice(e["core_courses"]); len(courses) > 0 {
			edu.CoreCourses = courses
		} else if courses := toSlice(e["courses"]); len(courses) > 0 {
			edu.CoreCourses = courses
		}
		snapshot.Educations = append(snapshot.Educations, edu)
	}

	// internships: parse InternshipJSON
	internships, err := decodeJSONArray[map[string]any](bytesToString(v.InternshipJSON))
	if err != nil {
		return fmt.Errorf("v2 internship_json: %w", err)
	}
	// If internship_json is empty, fallback to experience_json
	if len(internships) == 0 {
		internships, err = decodeJSONArray[map[string]any](bytesToString(v.ExperienceJSON))
		if err != nil {
			return fmt.Errorf("v2 experience_json (internship fallback): %w", err)
		}
	}
	snapshot.Internships = make([]model.ResumeSnapshotInternship, 0, len(internships))
	for _, exp := range internships {
		intern := model.ResumeSnapshotInternship{
			Company:     toString(exp["company"]),
			Role:        toString(exp["role"]),
			StartDate:   toString(exp["start_date"]),
			EndDate:     toString(exp["end_date"]),
			WorkContent: toString(exp["content"]),
			Metrics:     toString(exp["metrics"]),
		}
		if intern.WorkContent == "" {
			intern.WorkContent = toString(exp["work_content"])
		}
		if intern.StartDate == "" {
			intern.StartDate = toString(exp["start"])
		}
		if intern.EndDate == "" {
			intern.EndDate = toString(exp["end"])
		}
		if ts := toSlice(exp["tech_stack"]); len(ts) > 0 {
			intern.TechStack = ts
		}
		if ach := toSlice(exp["achievements"]); len(ach) > 0 {
			intern.Achievements = ach
		}
		snapshot.Internships = append(snapshot.Internships, intern)
	}

	// projects: parse ProjectJSON
	projects, err := decodeJSONArray[map[string]any](bytesToString(v.ProjectJSON))
	if err != nil {
		return fmt.Errorf("v2 project_json: %w", err)
	}
	snapshot.Projects = make([]model.ResumeSnapshotProject, 0, len(projects))
	for _, p := range projects {
		proj := model.ResumeSnapshotProject{
			Name:             toString(p["name"]),
			Role:             toString(p["role"]),
			Type:             toString(p["type"]),
			StartDate:        toString(p["start_date"]),
			EndDate:          toString(p["end_date"]),
			Description:      toString(p["description"]),
			Responsibilities: toString(p["responsibility"]),
			Link:             toString(p["link"]),
		}
		if proj.Responsibilities == "" {
			proj.Responsibilities = toString(p["responsibilities"])
		}
		if proj.StartDate == "" {
			proj.StartDate = toString(p["start"])
		}
		if proj.EndDate == "" {
			proj.EndDate = toString(p["end"])
		}
		if ts := toSlice(p["tech_stack"]); len(ts) > 0 {
			proj.TechStack = ts
		}
		if ach := toSlice(p["achievements"]); len(ach) > 0 {
			proj.Achievements = ach
		}
		snapshot.Projects = append(snapshot.Projects, proj)
	}

	// skills: parse SkillsJSON (V2 object format)
	skills, err := decodeJSONObject(bytesToString(v.SkillsJSON), map[string]any{})
	if err != nil {
		return fmt.Errorf("v2 skills_json: %w", err)
	}
	// K8-5-26-B: final 9-module skills — new keys first, old keys fallback
	profSkills := toSlice(skills["development_languages"])
	aiApp := toString(skills["ai_application"])

	snapshot.Skills = model.ResumeSnapshotSkills{
		ProfessionalSkills: profSkills,
		AIApplication:      aiApp,

		// Old 10-category fields for backward compatibility
		ProgrammingLanguages:   orSlice(toSlice(skills["languages"]), profSkills),
		BackendFrameworks:      toSlice(skills["frameworks"]),
		FrontendTechnologies:   toSlice(skills["frontend"]),
		Databases:              toSlice(skills["databases"]),
		Middleware:             toSlice(skills["middleware"]),
		CloudPlatforms:         toSlice(skills["cloud"]),
		DevTools:               toSlice(skills["tools"]),
		AITools:                toSlice(skills["ai_skills"]),
		AIApplicationScenarios: orSlice(toSlice(skills["ai_scenarios"]), splitStr(aiApp)),
		OtherKeywords:          toSlice(skills["other_keywords"]),
	}

	// certificates: parse CertificatesJSON
	certs, err := decodeJSONArray[map[string]any](bytesToString(v.CertificatesJSON))
	if err != nil {
		return fmt.Errorf("v2 certificates_json: %w", err)
	}
	snapshot.Certificates = make([]model.ResumeSnapshotCertificate, 0, len(certs))
	for _, c := range certs {
		cert := model.ResumeSnapshotCertificate{
			Name:     toString(c["name"]),
			Type:     toString(c["type"]),
			Level:    toString(c["level"]),
			Score:    toString(c["score"]),
			IssuedAt: toString(c["date"]),
		}
		if cert.IssuedAt == "" {
			cert.IssuedAt = toString(c["issued_at"])
		}
		snapshot.Certificates = append(snapshot.Certificates, cert)
	}

	// awards: parse AwardsJSON
	awards, err := decodeJSONArray[map[string]any](bytesToString(v.AwardsJSON))
	if err != nil {
		return fmt.Errorf("v2 awards_json: %w", err)
	}
	snapshot.Awards = make([]model.ResumeSnapshotAward, 0, len(awards))
	for _, a := range awards {
		awd := model.ResumeSnapshotAward{
			Name:        toString(a["name"]),
			Type:        toString(a["type"]),
			Level:       toString(a["level"]),
			AwardedAt:   toString(a["date"]),
			Description: toString(a["description"]),
		}
		if awd.AwardedAt == "" {
			awd.AwardedAt = toString(a["awarded_at"])
		}
		snapshot.Awards = append(snapshot.Awards, awd)
	}

	// portfolio: parse PortfolioJSON
	portfolio, err := decodeJSONObject(bytesToString(v.PortfolioJSON), map[string]any{})
	if err != nil {
		return fmt.Errorf("v2 portfolio_json: %w", err)
	}
	// K8-5-26-B: new works + homepage, old fields fallback
	hp := toString(portfolio["homepage"])
	works := parsePortfolioWorks(portfolio)
	snapshot.Portfolio = model.ResumeSnapshotPortfolio{
		Works:    works,
		Homepage: hp,
		Github:   toString(portfolio["github"]), Gitee: toString(portfolio["gitee"]),
		CNB: toString(portfolio["cnb"]), Blog: toString(portfolio["blog"]),
		DemoLinks: toSlice(portfolio["demo_links"]), PortfolioLinks: toSlice(portfolio["other_links"]),
	}
	if snapshot.Portfolio.Homepage == "" {
		snapshot.Portfolio.Homepage = toString(portfolio["blog"])
	}

	// extraInfo: parse ExtraJSON, fallback to self_evaluation column
	extra, err := decodeJSONObject(bytesToString(v.ExtraJSON), map[string]any{})
	if err != nil {
		return fmt.Errorf("v2 extra_json: %w", err)
	}
	snapshot.ExtraInfo = model.ResumeSnapshotExtraInfo{
		SelfEvaluation: toString(extra["self_evaluation"]),
		AdditionalInfo: toString(extra["additional_info"]),
	}
	if snapshot.ExtraInfo.SelfEvaluation == "" {
		snapshot.ExtraInfo.SelfEvaluation = v.SelfEvaluation
	}

	return nil
}

// ─── Generic JSON helpers ────────────────────────────────────────────────

// decodeJSONObject parses a JSON string into an object of type T.
// Empty / "null" strings return fallback.
// Invalid JSON returns an error.
func decodeJSONObject[T any](raw string, fallback T) (T, error) {
	if raw == "" || raw == "null" {
		return fallback, nil
	}
	var result T
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return fallback, fmt.Errorf("invalid JSON object: %w", err)
	}
	return result, nil
}

// decodeJSONArray parses a JSON string into a slice of T.
// Empty / "null" strings return empty slice.
// Invalid JSON returns an error.
func decodeJSONArray[T any](raw string) ([]T, error) {
	if raw == "" || raw == "null" {
		return []T{}, nil
	}
	var result []T
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		// fallback: try as single object
		var single T
		if err2 := json.Unmarshal([]byte(raw), &single); err2 == nil {
			return []T{single}, nil
		}
		return nil, fmt.Errorf("invalid JSON array: %w", err)
	}
	return result, nil
}

// bytesToString converts []byte to string, handling nil.
func bytesToString(b []byte) string {
	return string(b)
}

// ─── Type coercing helpers for map[string]any ────────────────────────────

func toString(v any) string {
	if v == nil {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return s
}

func toSlice(v any) []string {
	if v == nil {
		return nil
	}
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	result := make([]string, 0, len(arr))
	for _, item := range arr {
		if s, ok := item.(string); ok {
			result = append(result, s)
		}
	}
	return result
}

func toBool(v any) bool {
	if v == nil {
		return false
	}
	b, ok := v.(bool)
	if !ok {
		return false
	}
	return b
}

// orSlice returns a if non-empty, otherwise b (fallback). K8-5-26-B.
func orSlice(a, b []string) []string {
	if len(a) > 0 {
		return a
	}
	return b
}

// splitStr splits a string by newlines and commas for AI application fallback. K8-5-26-B.
func splitStr(s string) []string {
	if s == "" {
		return nil
	}
	// Simple split: just return as single-element slice for now
	return []string{s}
}

// parsePortfolioWorks extracts works[] from the portfolio JSON. K8-5-26-B.
func parsePortfolioWorks(port map[string]any) []model.ResumeSnapshotPortfolioWork {
	raw, ok := port["works"]
	if !ok {
		return nil
	}
	arr, ok := raw.([]any)
	if !ok {
		return nil
	}
	works := make([]model.ResumeSnapshotPortfolioWork, 0, len(arr))
	for _, w := range arr {
		wm, ok := w.(map[string]any)
		if !ok {
			continue
		}
		works = append(works, model.ResumeSnapshotPortfolioWork{
			Name:     toString(wm["name"]),
			Link:     toString(wm["link"]),
			Password: toString(wm["password"]),
		})
	}
	return works
}

func toInt(v any) int {
	if v == nil {
		return 0
	}
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case int64:
		return int(n)
	case json.Number:
		if i, err := n.Int64(); err == nil {
			return int(i)
		}
	}
	return 0
}
