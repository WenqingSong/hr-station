package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"offer-catcher/server/internal/model"
)

// RecommendationMockAgent generates mock job recommendations using simple rule-based scoring.
// It does NOT call any real LLM or read LLM_API_KEY.
type RecommendationMockAgent struct{}

// NewRecommendationMockAgent creates a new RecommendationMockAgent.
func NewRecommendationMockAgent() *RecommendationMockAgent {
	return &RecommendationMockAgent{}
}

// Recommend generates mock recommendations based on the resume and available jobs.
// It uses simple keyword matching rules to score and rank jobs.
func (a *RecommendationMockAgent) Recommend(_ context.Context, resume *model.ResumeVersion, jobs []*model.Job) ([]model.RecommendedJob, error) {
	type scoredJob struct {
		job   *model.Job
		score int
	}

	var scored []scoredJob

	// Parse resume skills
	resumeSkills := parseJSONStringArray(resume.SkillsJSON)
	resumeSkillSet := toLowerSet(resumeSkills)

	// Parse resume target info
	targetCity := strings.ToLower(strings.TrimSpace(resume.TargetCity))
	targetRole := strings.ToLower(strings.TrimSpace(resume.TargetRole))

	for _, job := range jobs {
		score := 60 // base score

		// City match (+5)
		if targetCity != "" && strings.Contains(strings.ToLower(job.City), targetCity) {
			score += 5
		}

		// Category/role match (+5~+10)
		jobCategoryLower := strings.ToLower(job.Category)
		jobTitleLower := strings.ToLower(job.Title)
		if targetRole != "" {
			if containsAny(targetRole, jobCategoryLower, jobTitleLower) {
				score += 10
			}
			// Check if target role keywords appear in category or title
			roleKeywords := strings.Fields(targetRole)
			for _, kw := range roleKeywords {
				if len(kw) >= 2 && (strings.Contains(jobCategoryLower, kw) || strings.Contains(jobTitleLower, kw)) {
					score += 5
					break
				}
			}
		}

		// Skill overlap from skill_keywords_json (+5 per matched skill, max 20)
		jobSkills := parseJSONStringArray(job.SkillKeywordsJSON)
		jobSkillSet := toLowerSet(jobSkills)
		overlap := 0
		for skill := range resumeSkillSet {
			if jobSkillSet[skill] {
				overlap++
			}
		}
		score += overlap * 5
		if overlap > 4 {
			score += 5 // bonus for high overlap
		}

		// JD keyword matching from description/requirements (+5 per match, max 20)
		if job.Description != "" || job.Requirements != "" {
			jobText := strings.ToLower(job.Description + " " + job.Requirements)
			descMatches := 0
			for skill := range resumeSkillSet {
				if !jobSkillSet[skill] && strings.Contains(jobText, skill) {
					descMatches++
				}
			}
			score += min(20, descMatches*5)
		}

		// data_quality=full_jd bonus (+5)
		if job.DataQuality == "full_jd" {
			score += 5
		}

		// is_featured bonus (+5)
		if job.IsFeatured {
			score += 5
		}

		// summary_only penalty (-10)
		if job.DataQuality == "summary_only" {
			score -= 10
		}

		// Cap at 95
		if score > 95 {
			score = 95
		}
		if score < 0 {
			score = 0
		}

		scored = append(scored, scoredJob{job: job, score: score})
	}

	// Sort by score descending
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	// Take top 5 (or fewer if not enough jobs)
	limit := 5
	if len(scored) < limit {
		limit = len(scored)
	}

	var results []model.RecommendedJob
	for i := 0; i < limit; i++ {
		sj := scored[i]
		priority := getPriority(sj.score)
		reason := generateReason(resume, sj.job, sj.score)
		highlights := generateHighlights(resume, sj.job, resumeSkills, parseJSONStringArray(sj.job.SkillKeywordsJSON))
		risks := generateRisks(resume, sj.job, resumeSkills, parseJSONStringArray(sj.job.SkillKeywordsJSON))
		action := generateAction(sj.job, priority)

		results = append(results, model.RecommendedJob{
			JobID:                sj.job.ID,
			Title:                sj.job.Title,
			Department:           sj.job.Department,
			City:                 sj.job.City,
			Category:             sj.job.Category,
			JobType:              sj.job.JobType,
			Priority:             priority,
			RecommendationScore:  sj.score,
			RecommendationReason: reason,
			MatchHighlights:      highlights,
			RiskPoints:           risks,
			SuggestedAction:      action,
			ExternalSource:       sj.job.ExternalSource,
			ExternalID:           sj.job.ExternalID,
			DataQuality:          sj.job.DataQuality,
			SourceURL:            sj.job.SourceURL,
			IsFeatured:           sj.job.IsFeatured,
		})
	}

	// Ensure at least 3 results if possible
	if len(results) < 3 && len(scored) > len(results) {
		// Already handled by limit above
	}

	return results, nil
}

func getPriority(score int) string {
	if score >= 80 {
		return "strong_recommend"
	}
	if score >= 65 {
		return "recommend"
	}
	return "try"
}

func generateReason(resume *model.ResumeVersion, job *model.Job, score int) string {
	var parts []string

	targetCity := strings.TrimSpace(resume.TargetCity)
	targetRole := strings.TrimSpace(resume.TargetRole)

	if targetCity != "" && job.City == targetCity {
		parts = append(parts, fmt.Sprintf("目标城市与岗位城市一致（%s）", job.City))
	}

	if targetRole != "" {
		jobLower := strings.ToLower(job.Category + " " + job.Title)
		roleLower := strings.ToLower(targetRole)
		if containsAny(roleLower, jobLower) {
			parts = append(parts, fmt.Sprintf("目标岗位与该岗位的%s方向高度一致", job.Category))
		}
	}

	resumeSkills := parseJSONStringArray(resume.SkillsJSON)
	jobSkills := parseJSONStringArray(job.SkillKeywordsJSON)
	resumeSkillSet := toLowerSet(resumeSkills)
	jobSkillSet := toLowerSet(jobSkills)
	overlap := 0
	for skill := range resumeSkillSet {
		if jobSkillSet[skill] {
			overlap++
		}
	}

	if overlap > 0 {
		// Find top overlapping skills
		var matched []string
		for _, s := range resumeSkills {
			if jobSkillSet[strings.ToLower(s)] {
				matched = append(matched, s)
			}
			if len(matched) >= 3 {
				break
			}
		}
		parts = append(parts, fmt.Sprintf("技能中包含 %s 等关键词", strings.Join(matched, "、")))
	}

	if len(parts) == 0 {
		if score >= 65 {
			return fmt.Sprintf("你的简历与该岗位的%s方向有一定匹配度。", job.Category)
		}
		return fmt.Sprintf("该岗位属于%s方向，可作为求职参考。", job.Category)
	}

	return strings.Join(parts, "，且") + "。"
}

func generateHighlights(resume *model.ResumeVersion, job *model.Job, resumeSkills, jobSkills []string) []string {
	var highlights []string

	// City match
	if strings.TrimSpace(resume.TargetCity) != "" && job.City == strings.TrimSpace(resume.TargetCity) {
		highlights = append(highlights, "目标城市与简历 target_city 一致")
	}

	// Role match
	targetRole := strings.TrimSpace(resume.TargetRole)
	if targetRole != "" {
		jobText := strings.ToLower(job.Category + " " + job.Title)
		if containsAny(strings.ToLower(targetRole), jobText) {
			highlights = append(highlights, "目标岗位与简历 target_role 接近")
		}
	}

	// Skill overlap
	jobSkillSet := toLowerSet(jobSkills)
	var matchedSkills []string
	for _, s := range resumeSkills {
		if jobSkillSet[strings.ToLower(s)] {
			matchedSkills = append(matchedSkills, s)
		}
	}
	if len(matchedSkills) > 0 {
		highlights = append(highlights, fmt.Sprintf("技能关键词与 JD 要求存在交集（%s）", strings.Join(matchedSkills[:min(3, len(matchedSkills))], "、")))
	}

	// Experience mention
	if len(resume.ExperienceJSON) > 0 {
		var experiences []map[string]interface{}
		if json.Unmarshal(resume.ExperienceJSON, &experiences) == nil && len(experiences) > 0 {
			highlights = append(highlights, "项目经历中包含相关实习或工作经验")
		}
	}

	if len(highlights) == 0 {
		highlights = append(highlights, "岗位方向与求职目标部分相关")
	}

	return highlights
}

func generateRisks(resume *model.ResumeVersion, job *model.Job, resumeSkills, jobSkills []string) []string {
	var risks []string

	resumeSkillSet := toLowerSet(resumeSkills)

	// Find missing skills
	var missing []string
	for _, s := range jobSkills {
		if !resumeSkillSet[strings.ToLower(s)] {
			missing = append(missing, s)
		}
	}
	if len(missing) > 0 {
		top := missing[:min(2, len(missing))]
		risks = append(risks, fmt.Sprintf("%s 经验表达仍可进一步强化", strings.Join(top, " 或 ")))
	}

	// City mismatch
	if strings.TrimSpace(resume.TargetCity) != "" && job.City != strings.TrimSpace(resume.TargetCity) {
		risks = append(risks, fmt.Sprintf("工作地点为%s，与目标城市不一致", job.City))
	}

	if len(risks) == 0 {
		risks = append(risks, "项目经历偏校园级别，企业级项目深度可进一步提升")
	}

	return risks
}

func generateAction(job *model.Job, priority string) string {
	switch priority {
	case "strong_recommend":
		return fmt.Sprintf("建议优先查看该岗位详情，并在后续匹配分析中重点检查项目经历与岗位要求的对应关系。")
	case "recommend":
		return fmt.Sprintf("建议关注该岗位，可进一步了解%s方向的具体要求并对照简历内容。", job.Category)
	default:
		return fmt.Sprintf("可作为备选方向了解，建议先补充%s相关技能后再重点关注。", job.Category)
	}
}

// parseJSONStringArray parses a JSON array of strings from bytes.
func parseJSONStringArray(data []byte) []string {
	if len(data) == 0 {
		return nil
	}
	var arr []string
	if err := json.Unmarshal(data, &arr); err != nil {
		return nil
	}
	return arr
}

// toLowerSet converts a string slice to a lowercase map set.
func toLowerSet(items []string) map[string]bool {
	set := make(map[string]bool, len(items))
	for _, item := range items {
		set[strings.ToLower(strings.TrimSpace(item))] = true
	}
	return set
}

// containsAny checks if target contains any of the search strings.
func containsAny(target string, searches ...string) bool {
	for _, s := range searches {
		if strings.Contains(s, target) || strings.Contains(target, s) {
			return true
		}
	}
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
