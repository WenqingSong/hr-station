package service

import (
	"encoding/json"
	"fmt"
	"log"

	"offer-catcher/server/internal/model"
)

// SnapshotToPromptMap converts a ResumeSnapshotV2 into a map suitable for injection
// into an LLM Prompt as a JSON context block (RESUME_SNAPSHOT_JSON).
// It replaces the old per-agent manual field assembly that used raw DB fields like
// target_role / target_city / skills_json / education_json etc.
//
// All downstream agents (Recommendation, Match, Optimization, Compare) use this shared helper,
// ensuring a consistent structured input format that matches the V2 module layout.
//
// Privacy note: phone / email / user_id / token are not present in the snapshot.
func SnapshotToPromptMap(snapshot model.ResumeSnapshotV2) map[string]any {
	// helper to avoid null
	emptyStr := func(s string) string {
		if s == "" {
			return "未填写"
		}
		return s
	}
	emptySlice := func(s []string) []string {
		if len(s) == 0 {
			return []string{"未填写"}
		}
		return s
	}

	return map[string]any{
		// ── 简历元信息 ──
		"resume_name":      emptyStr(snapshot.Meta.ResumeName),
		"resume_version_no": snapshot.Meta.ResumeVersionNo,
		"target_direction": emptyStr(snapshot.Meta.TargetDirection),

		// ── 基础信息 ──
		"profile": map[string]any{
			"name":            emptyStr(snapshot.Profile.Name),
			"school":          emptyStr(snapshot.Profile.School),
			"degree":          emptyStr(snapshot.Profile.Degree),
			"major":           emptyStr(snapshot.Profile.Major),
			"graduation_year": emptyStr(snapshot.Profile.GraduationYear),
			"current_city":    emptyStr(snapshot.Profile.CurrentCity),
		},

		// ── 求职意向 ──
		"job_intention": map[string]any{
			"target_roles":         emptySlice(snapshot.JobIntention.TargetRoles),
			"job_type":             emptyStr(snapshot.JobIntention.JobType),
			"expected_cities":      emptySlice(snapshot.JobIntention.ExpectedCities),
			"accept_other_cities":  snapshot.JobIntention.AcceptOtherCities,
			"earliest_start_date":  emptyStr(snapshot.JobIntention.EarliestStartDate),
			"internship_duration":  emptyStr(snapshot.JobIntention.InternshipDuration),
			"weekly_available_days": snapshot.JobIntention.WeeklyAvailableDays,
			"target_industries":    emptySlice(snapshot.JobIntention.TargetIndustries),
			"target_companies":     emptySlice(snapshot.JobIntention.TargetCompanies),
		},

		// ── 教育背景 ──
		"educations": snapshot.Educations,

		// ── 实习经历 ──
		"internships": snapshot.Internships,

		// ── 项目经历 ──
		"projects": snapshot.Projects,

		// ── 技能信息 ──
		"skills": snapshot.Skills,

		// ── 证书与语言能力 ──
		"certificates": snapshot.Certificates,

		// ── 获奖与竞赛 ──
		"awards": snapshot.Awards,

		// ── 作品链接 ──
		"portfolio": snapshot.Portfolio,

		// ── 自我评价与补充信息 ──
		"extra_info": snapshot.ExtraInfo,
	}
}

// BuildResumeSnapshot builds the RESUME_SNAPSHOT_JSON prompt value from a resume version.
// It normalizes the raw DB model into a unified ResumeSnapshotV2 and serializes it.
func BuildResumeSnapshotJSON(resume *model.ResumeVersion) (string, error) {
	snapshot, err := NormalizeResumeToSnapshotV2(*resume)
	if err != nil {
		log.Printf("[BuildResumeSnapshotJSON] WARNING: normalize failed for resume id=%d: %v", resume.ID, err)
		// Fallback: return minimal valid JSON
		return "{}", fmt.Errorf("snapshot normalize failed: %w", err)
	}

	snapshotMap := SnapshotToPromptMap(snapshot)
	data, err := json.Marshal(snapshotMap)
	if err != nil {
		return "{}", fmt.Errorf("failed to marshal snapshot: %w", err)
	}
	return string(data), nil
}
