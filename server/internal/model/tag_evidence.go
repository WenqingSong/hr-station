// ─── K13-Fix-A: Tag/Evidence 统一数据契约 ────────────────────────────────────

package model

// ─── TagItem ──────────────────────────────────────────────────────────────

// TagItem represents a single tag with metadata.
// Used by both ResumeTagProfile and JobTagProfile.
type TagItem struct {
	Tag         string   `json:"tag"`
	Group       string   `json:"group"`
	Weight      float64  `json:"weight"`
	Confidence  int      `json:"confidence"` // 1-3
	EvidenceRef []string `json:"evidence_ref,omitempty"`
}

// ─── EvidenceItem ─────────────────────────────────────────────────────────

// EvidenceItem is a short evidence snippet.
// Text must NOT contain full raw_text or JD — keep it under 500 chars.
type EvidenceItem struct {
	ID       string  `json:"id"`
	Type     string  `json:"type"` // skill / project / internship / education / cert / jd_req / jd_resp / jd_kw
	Text     string  `json:"text"`
	Source   string  `json:"source"`
	Strength float64 `json:"strength"` // 0.0-1.0
}

// ─── ResumeEvidence ───────────────────────────────────────────────────────

// ResumeEvidence groups evidence items for a resume.
type ResumeEvidence struct {
	SkillEvidence       []EvidenceItem `json:"skill_evidence"`
	ProjectEvidence     []EvidenceItem `json:"project_evidence"`
	InternshipEvidence  []EvidenceItem `json:"internship_evidence"`
	EducationEvidence   []EvidenceItem `json:"education_evidence"`
	CertificateEvidence []EvidenceItem `json:"certificate_evidence"`

	ProjectCount    int `json:"project_count"`
	InternshipCount int `json:"internship_count"`
	SkillCount      int `json:"skill_count"`
}

// IsEmpty returns true when evidence has no meaningful content.
func (e ResumeEvidence) IsEmpty() bool {
	return e.ProjectCount == 0 && e.InternshipCount == 0 && e.SkillCount == 0 &&
		len(e.SkillEvidence) == 0 && len(e.ProjectEvidence) == 0 &&
		len(e.InternshipEvidence) == 0
}

// ─── JobEvidence ──────────────────────────────────────────────────────────

// JobEvidence groups evidence items for a job.
type JobEvidence struct {
	ResponsibilityEvidence []EvidenceItem `json:"responsibility_evidence"`
	RequirementEvidence    []EvidenceItem `json:"requirement_evidence"`
	KeywordEvidence        []EvidenceItem `json:"keyword_evidence"`
}

// IsEmpty returns true when all evidence lists are empty.
func (e JobEvidence) IsEmpty() bool {
	return len(e.ResponsibilityEvidence) == 0 &&
		len(e.RequirementEvidence) == 0 &&
		len(e.KeywordEvidence) == 0
}

// ─── ResumeTagProfile ─────────────────────────────────────────────────────

// ResumeTagProfile is the K13 unified resume tag + evidence result.
// It wraps the existing ResumeProfile with parsed tag/evidence data.
type ResumeTagProfile struct {
	ResumeID          uint64  `json:"resume_id"`
	DominantProfile   string  `json:"dominant_profile"`
	ProfileSummary    string  `json:"profile_summary"`

	PrimaryTags       []TagItem `json:"primary_tags"`
	CandidateTags     []TagItem `json:"candidate_tags"`
	SkillTags         []TagItem `json:"skill_tags"`
	ProjectTags       []TagItem `json:"project_tags"`
	InternshipTags    []TagItem `json:"internship_tags"`
	EducationTags     []TagItem `json:"education_tags"`
	RiskTags          []TagItem `json:"risk_tags"`
	IntentTags        []TagItem `json:"intent_tags"`

	Evidence          ResumeEvidence `json:"evidence"`
	Source            string         `json:"source"`
	Version           string         `json:"version"`
}

// ─── JobTagProfile ────────────────────────────────────────────────────────

// JobTagProfile is the K13 unified job tag + evidence result.
type JobTagProfile struct {
	JobID           uint64  `json:"job_id"`
	Direction       string  `json:"direction"`
	Category        string  `json:"category"`

	PrimaryTags      []TagItem `json:"primary_tags"`
	SkillTags        []TagItem `json:"skill_tags"`
	RequirementTags  []TagItem `json:"requirement_tags"`
	ResponsibilityTags []TagItem `json:"responsibility_tags"`
	RiskTags         []TagItem `json:"risk_tags"`

	Evidence         JobEvidence `json:"evidence"`
	Source           string      `json:"source"`
	Version          string      `json:"version"`
}

// ─── TagEvidenceContext ───────────────────────────────────────────────────

// TagEvidenceContext bundles a resume profile and matching job profiles.
// Used by Fast/Deep Agent 1 as shared context input.
type TagEvidenceContext struct {
	ResumeProfile   ResumeTagProfile  `json:"resume_profile"`
	JobProfiles     []JobTagProfile   `json:"job_profiles"`
}

// ─── JobCandidate (for recall) ────────────────────────────────────────────

// JobCandidate represents a recalled job with tag-based matching metadata.
type JobCandidate struct {
	JobID        uint64   `json:"job_id"`
	Title        string   `json:"title"`
	Department   string   `json:"department"`
	City         string   `json:"city"`
	Category     string   `json:"category"`
	Direction    string   `json:"direction"`
	TagScore     float64  `json:"tag_score"`
	MatchedTags  []string `json:"matched_tags"`
	MissingTags  []string `json:"missing_tags"`
	EvidenceRefs []string `json:"evidence_refs"`
}

// ─── Tag/Hardguard constants ──────────────────────────────────────────────

const TagEvidenceVersionV1 = "tag_agent_v1"

// DominantProfile values
const (
	DominantAlgorithmResearch  = "algorithm_research"
	DominantBackendEngineering = "backend_engineering"
	DominantProductOperation   = "product_operation"
	DominantBusinessSolution   = "business_solution"
	DominantCareerExploration  = "career_exploration"
)

// Tag groups
const (
	TagGroupDirection  = "direction"
	TagGroupCapability = "capability"
	TagGroupDomain     = "domain"
	TagGroupSkill      = "skill"
	TagGroupRisk       = "risk"
)
